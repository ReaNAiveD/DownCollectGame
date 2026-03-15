package main

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/naived/downcollect/internal/core"
	"github.com/naived/downcollect/internal/view"

	"github.com/coder/websocket"
)

// Room represents a game room.
type Room struct {
	mu        sync.Mutex
	Code      string
	Engine    *core.Engine
	HostID    string
	Nicknames map[string]string
	Connected map[string]bool
	Clients   map[string]*Client
}

// Client represents a connected WebSocket client.
type Client struct {
	PlayerID string
	Nickname string
	Conn     *websocket.Conn
}

// Server manages rooms and WebSocket connections.
type Server struct {
	mu    sync.Mutex
	rooms map[string]*Room // keyed by room code
}

const maxRooms = 100

// WsMessage is the envelope for all WebSocket messages.
type WsMessage struct {
	Op   int             `json:"op"`
	Data json.RawMessage `json:"data"`
}

// Client -> Server opcodes
const (
	COpSetNickname  = 1
	COpCreateRoom   = 2
	COpJoinRoom     = 3
	COpStartGame    = 4
	COpPlayerAction = 5
)

// Server -> Client opcodes
const (
	SOpGameState   = 101
	SOpError       = 107
	SOpRoomUpdate  = 108
	SOpRoomCreated = 109
)

const roomCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
const roomCodeLength = 4

func generateRoomCode() string {
	b := make([]byte, roomCodeLength)
	for i := range b {
		idx, _ := crand.Int(crand.Reader, big.NewInt(int64(len(roomCodeChars))))
		b[i] = roomCodeChars[idx.Int64()]
	}
	return string(b)
}

// Version is set at build time via -ldflags.
var Version = "dev"

func main() {
	log.Printf("DownCollect server version %s", Version)
	server := &Server{rooms: make(map[string]*Room)}

	// Determine frontend path: check env, then ./public next to binary, then fallback
	frontendDir := ""
	if envDir := os.Getenv("FRONTEND_DIR"); envDir != "" {
		frontendDir = envDir
	} else {
		// Look for "public" directory next to the executable
		exe, _ := os.Executable()
		exeDir := filepath.Dir(exe)
		candidate := filepath.Join(exeDir, "public")
		if _, err := os.Stat(candidate); err == nil {
			frontendDir = candidate
		}
	}

	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", server.handleWebSocket)

	// API endpoints
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"status":"ok","version":"%s"}`, Version)))
	})

	// Serve frontend static files
	if frontendDir != "" {
		absPath, _ := filepath.Abs(frontendDir)
		if _, err := os.Stat(absPath); err == nil {
			fs := http.FileServer(http.Dir(absPath))
			mux.Handle("/", fs)
			log.Printf("Serving frontend from %s", absPath)
		} else {
			log.Printf("Frontend directory not found at %s, serving API only", absPath)
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"message":"DownCollect API server. Frontend not built yet."}`))
			})
		}
	} else {
		log.Printf("No frontend directory configured, serving API only")
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message":"DownCollect API server. Set FRONTEND_DIR or place files in public/."}`))
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("DownCollect server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("WebSocket accept error: %v", err)
		return
	}

	// Generate a player ID from query param or random
	playerID := r.URL.Query().Get("deviceId")
	if playerID == "" {
		playerID = generateRoomCode() + generateRoomCode() // reuse the code generator for UUIDs
	}

	client := &Client{
		PlayerID: playerID,
		Nickname: "Player",
		Conn:     conn,
	}

	log.Printf("Client connected: %s", playerID)

	ctx := r.Context()
	var currentRoom *Room

	defer func() {
		if currentRoom != nil {
			currentRoom.mu.Lock()
			currentRoom.Connected[playerID] = false
			delete(currentRoom.Clients, playerID)
			currentRoom.mu.Unlock()
			broadcastRoomState(currentRoom)
		}
		conn.Close(websocket.StatusNormalClosure, "")
		log.Printf("Client disconnected: %s", playerID)
	}()

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			break
		}

		var msg WsMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			sendErrorToClient(conn, "invalid message format")
			continue
		}

		switch msg.Op {
		case COpSetNickname:
			var req struct {
				Nickname string `json:"nickname"`
			}
			json.Unmarshal(msg.Data, &req)
			if req.Nickname != "" {
				client.Nickname = req.Nickname
			}

		case COpCreateRoom:
			var req struct {
				Config *core.GameConfig `json:"config,omitempty"`
			}
			json.Unmarshal(msg.Data, &req)
			room, err := s.createRoom(client, req.Config)
			if err != nil {
				sendErrorToClient(conn, err.Error())
				continue
			}
			currentRoom = room
			sendToClient(conn, SOpRoomCreated, map[string]string{
				"roomCode": room.Code,
				"playerId": playerID,
			})
			broadcastRoomState(room)

		case COpJoinRoom:
			var req struct {
				RoomCode string `json:"roomCode"`
			}
			json.Unmarshal(msg.Data, &req)
			room, err := s.joinRoom(req.RoomCode, client)
			if err != nil {
				sendErrorToClient(conn, err.Error())
				continue
			}
			currentRoom = room
			sendToClient(conn, SOpRoomCreated, map[string]string{
				"roomCode": room.Code,
				"playerId": playerID,
			})
			broadcastRoomState(room)

		case COpStartGame:
			if currentRoom == nil {
				sendErrorToClient(conn, "not in a room")
				continue
			}
			currentRoom.mu.Lock()
			if playerID != currentRoom.HostID {
				currentRoom.mu.Unlock()
				sendErrorToClient(conn, "only host can start the game")
				continue
			}
			if err := currentRoom.Engine.StartGame(); err != nil {
				currentRoom.mu.Unlock()
				sendErrorToClient(conn, err.Error())
				continue
			}
			broadcastGameStateLocked(currentRoom)
			currentRoom.mu.Unlock()

		case COpPlayerAction:
			if currentRoom == nil {
				sendErrorToClient(conn, "not in a room")
				continue
			}
			var choice core.PlayerChoice
			json.Unmarshal(msg.Data, &choice)
			currentRoom.mu.Lock()
			err := currentRoom.Engine.HandleAction(playerID, choice)
			if err != nil {
				currentRoom.mu.Unlock()
				sendErrorToClient(conn, err.Error())
				continue
			}
			broadcastGameStateLocked(currentRoom)
			currentRoom.mu.Unlock()
		}
	}
}

func (s *Server) createRoom(client *Client, cfg *core.GameConfig) (*Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cleanup finished/empty rooms before checking limit
	for code, room := range s.rooms {
		room.mu.Lock()
		phase := room.Engine.State.Phase
		connected := 0
		for _, v := range room.Connected {
			if v {
				connected++
			}
		}
		room.mu.Unlock()
		if phase == core.PhaseFinished || connected == 0 {
			delete(s.rooms, code)
		}
	}

	if len(s.rooms) >= maxRooms {
		return nil, fmt.Errorf("server is full, please try again later")
	}

	code := generateRoomCode()
	for s.rooms[code] != nil {
		code = generateRoomCode()
	}

	config := core.DefaultGameConfig()
	if cfg != nil {
		if cfg.InitialHandSize >= 0 && cfg.InitialHandSize <= 5 {
			config.InitialHandSize = cfg.InitialHandSize
		}
		if cfg.RevealRounds >= 1 && cfg.RevealRounds <= 10 {
			config.RevealRounds = cfg.RevealRounds
		}
		if cfg.PickRounds >= 1 && cfg.PickRounds <= 8 {
			config.PickRounds = cfg.PickRounds
		}
		if cfg.ScoringMode == core.ScoringModeSpecial || cfg.ScoringMode == core.ScoringModeBaseOnly {
			config.ScoringMode = cfg.ScoringMode
		}
	}

	engine := core.NewEngine(config)
	engine.AddPlayer(client.PlayerID)

	room := &Room{
		Code:      code,
		Engine:    engine,
		HostID:    client.PlayerID,
		Nicknames: map[string]string{client.PlayerID: client.Nickname},
		Connected: map[string]bool{client.PlayerID: true},
		Clients:   map[string]*Client{client.PlayerID: client},
	}

	s.rooms[code] = room
	log.Printf("Room created: %s by %s (%s) config=%+v", code, client.Nickname, client.PlayerID, config)
	return room, nil
}

func (s *Server) joinRoom(roomCode string, client *Client) (*Room, error) {
	s.mu.Lock()
	room, ok := s.rooms[roomCode]
	s.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomCode)
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	// Check if reconnecting
	if room.Connected[client.PlayerID] != false || room.Nicknames[client.PlayerID] != "" {
		// Reconnecting
		room.Connected[client.PlayerID] = true
		room.Clients[client.PlayerID] = client
		room.Nicknames[client.PlayerID] = client.Nickname
		return room, nil
	}

	if err := room.Engine.AddPlayer(client.PlayerID); err != nil {
		return nil, err
	}

	room.Nicknames[client.PlayerID] = client.Nickname
	room.Connected[client.PlayerID] = true
	room.Clients[client.PlayerID] = client
	log.Printf("Player %s (%s) joined room %s", client.Nickname, client.PlayerID, roomCode)
	return room, nil
}

func broadcastRoomState(room *Room) {
	room.mu.Lock()
	defer room.mu.Unlock()

	gs := room.Engine.State
	if gs.Phase == core.PhaseWaitingPlayers {
		// Send room update
		players := make([]map[string]interface{}, 0, len(gs.Players))
		for _, p := range gs.Players {
			players = append(players, map[string]interface{}{
				"playerId":  p.PlayerID,
				"nickname":  room.Nicknames[p.PlayerID],
				"seatIndex": p.SeatIndex,
				"isHost":    p.PlayerID == room.HostID,
				"connected": room.Connected[p.PlayerID],
			})
		}
		data := map[string]interface{}{
			"phase":    "WaitingPlayers",
			"roomCode": room.Code,
			"players":  players,
			"config":   gs.Config,
		}
		for _, client := range room.Clients {
			sendToClient(client.Conn, SOpRoomUpdate, data)
		}
	} else {
		broadcastGameStateLocked(room)
	}
}

func broadcastGameState(room *Room) {
	room.mu.Lock()
	defer room.mu.Unlock()
	broadcastGameStateLocked(room)
}

func broadcastGameStateLocked(room *Room) {
	for playerID, client := range room.Clients {
		if !room.Connected[playerID] {
			continue
		}
		pv := view.GeneratePlayerView(room.Engine, playerID, room.Nicknames, room.HostID, room.Connected)
		sendToClient(client.Conn, SOpGameState, pv)
	}
}

func sendToClient(conn *websocket.Conn, op int, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}
	msg := WsMessage{Op: op, Data: payload}
	msgBytes, _ := json.Marshal(msg)
	conn.Write(context.Background(), websocket.MessageText, msgBytes)
}

func sendErrorToClient(conn *websocket.Conn, message string) {
	sendToClient(conn, SOpError, map[string]string{"error": message})
}
