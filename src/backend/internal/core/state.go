package core

// GamePhase represents the current phase of the game.
type GamePhase int

const (
	PhaseWaitingPlayers GamePhase = iota
	PhaseCharacterSelect
	PhaseSetup
	PhaseReveal
	PhasePick
	PhaseScoring
	PhaseFinished
)

func (p GamePhase) String() string {
	names := []string{
		"WaitingPlayers",
		"CharacterSelect",
		"Setup",
		"Reveal",
		"Pick",
		"Scoring",
		"Finished",
	}
	if int(p) < len(names) {
		return names[p]
	}
	return "Unknown"
}

// TurnStep represents sub-steps within a reveal phase turn.
type TurnStep int

const (
	StepRevealCard    TurnStep = iota // Waiting for player to draw
	StepSwapDecision                  // Player decides whether to swap drawn card with hand
	StepRevealConfirm                 // All players see the card to be executed; active player confirms
	StepEffectAction                  // Waiting for player choice for effect
	StepEffectResolve                 // Executing effect (including recursive reveal)
	StepTurnEnd                       // Turn ending
)

// ActionType defines the type of player action required.
type ActionType int

const (
	ActionRevealCard        ActionType = iota // Click to reveal
	ActionConfirmReveal                       // Confirm after seeing revealed card
	ActionSelectSlots                         // Select board slots (3:2 slots, 5/6/9/10:1 slot)
	ActionSelectPosition                      // Select insert position in row (7, 8)
	ActionSelectRow                           // Select a row (J, Q)
	ActionSelectSlotsPerRow                   // Select one slot per row (K)
	ActionSwapDecision                        // Swap or not
	ActionPickCard                            // Pick a card from board
	ActionChooseCharacter                     // Choose character
)

// PendingAction represents an action the game is waiting for from a player.
type PendingAction struct {
	Type     ActionType
	PlayerID string
	Effect   EffectContext
}

// EffectContext holds typed context for effect-related pending actions.
type EffectContext struct {
	Rank         Rank
	Card         *Card       // The revealed card (for 7/8/K that haven't placed yet)
	EffectName   string      // Human-readable effect name
	SelectCount  int         // How many slots to select
	ValidSlots   []SlotCoord // Valid slots for selection
	ValidRows    []int       // Valid rows for selection
	MaxInsertPos int         // Max insert position (0..MaxInsertPos-1)
}

// PlayerChoice represents a player's response to a PendingAction.
type PlayerChoice struct {
	Type           ActionType
	SelectedSlots  []SlotCoord `json:"selectedSlots,omitempty"`
	SelectedRow    int         `json:"selectedRow,omitempty"`
	SelectedPos    int         `json:"selectedPos,omitempty"`
	HandCardIndex  int         `json:"handCardIndex,omitempty"`
	DoSwap         bool        `json:"doSwap,omitempty"`
	CharacterIndex int         `json:"characterIndex,omitempty"`
}

// ScoreEntry records a single card's scoring detail.
type ScoreEntry struct {
	CardIndex   int    `json:"cardIndex"`
	Card        Card   `json:"card"`
	RuleName    string `json:"ruleName"`
	Score       int    `json:"score"`
	Description string `json:"description"`
	Skipped     bool   `json:"skipped"`
}

// PlayerState holds the game state for a single player.
type PlayerState struct {
	PlayerID     string
	SeatIndex    int
	Hand         *Hand
	Character    Character
	Score        int
	ScoreDetails []ScoreEntry
	PeekInfo     map[SlotCoord]Card // Cards this player has peeked at
}

func NewPlayerState(playerID string, seatIndex int) *PlayerState {
	return &PlayerState{
		PlayerID:     playerID,
		SeatIndex:    seatIndex,
		Hand:         NewHand(),
		Character:    &NoopCharacter{},
		Score:        0,
		ScoreDetails: make([]ScoreEntry, 0),
		PeekInfo:     make(map[SlotCoord]Card),
	}
}

// ScoringMode determines which scoring rules to use.
type ScoringMode int

const (
	ScoringModeSpecial  ScoringMode = iota // Full special scoring rules per card
	ScoringModeBaseOnly                    // Base score only, no special rules
)

// GameConfig holds configurable game parameters.
type GameConfig struct {
	MinPlayers      int         `json:"minPlayers"`
	MaxPlayers      int         `json:"maxPlayers"`
	RemovedCards    int         `json:"removedCards"`
	InitialHandSize int         `json:"initialHandSize"`
	RevealRounds    int         `json:"revealRounds"`
	PickRounds      int         `json:"pickRounds"`
	TurnTimeoutSec  int         `json:"turnTimeoutSec"`
	ScoringMode     ScoringMode `json:"scoringMode"`
}

func DefaultGameConfig() GameConfig {
	return GameConfig{
		MinPlayers:      2,
		MaxPlayers:      4,
		RemovedCards:    5,
		InitialHandSize: 1,
		RevealRounds:    6,
		PickRounds:      4,
		TurnTimeoutSec:  30,
		ScoringMode:     ScoringModeSpecial,
	}
}

// GameEvent records a visible game event for all players.
type GameEvent struct {
	Type        string      `json:"type"`
	SeatIndex   int         `json:"seatIndex"` // which player triggered this
	Card        *Card       `json:"card,omitempty"`
	Slots       []SlotCoord `json:"slots,omitempty"`
	Cards       []Card      `json:"cards,omitempty"` // revealed cards during effect
	Row         *int        `json:"row,omitempty"`
	Position    *int        `json:"position,omitempty"`
	Description string      `json:"description"`
}

// IntPtr is a helper to create a pointer to an int (for JSON omitempty).
func IntPtr(v int) *int { return &v }

// GameState holds the authoritative state of a game.
type GameState struct {
	Config          GameConfig
	Phase           GamePhase
	TurnStep        TurnStep
	Round           int // 0-based current round
	TurnInRound     int // 0-based current turn within round
	StartingSeat    int // seat index of starting player for current round
	ActiveSeat      int // seat index of current active player
	Players         []*PlayerState
	Deck            *Deck
	Board           *Board
	RemovedPile     []Card
	PendingAction   *PendingAction
	RevealedCard    *Card       // Currently revealed card (for display)
	RecursiveReveal bool        // True when processing recursive reveal (card 4)
	EventLog        []GameEvent // Recent events visible to all players
}

func NewGameState(config GameConfig) *GameState {
	return &GameState{
		Config:      config,
		Phase:       PhaseWaitingPlayers,
		Players:     make([]*PlayerState, 0),
		Board:       NewBoard(),
		RemovedPile: make([]Card, 0),
		EventLog:    make([]GameEvent, 0),
	}
}

// AddEvent appends a game event visible to all players.
func (gs *GameState) AddEvent(event GameEvent) {
	gs.EventLog = append(gs.EventLog, event)
}

// ClearEvents clears the event log (call at start of each action cycle).
func (gs *GameState) ClearEvents() {
	gs.EventLog = gs.EventLog[:0]
}

// GetPlayer returns the player state for a given player ID.
func (gs *GameState) GetPlayer(playerID string) *PlayerState {
	for _, p := range gs.Players {
		if p.PlayerID == playerID {
			return p
		}
	}
	return nil
}

// GetPlayerBySeat returns the player state for a given seat index.
func (gs *GameState) GetPlayerBySeat(seat int) *PlayerState {
	for _, p := range gs.Players {
		if p.SeatIndex == seat {
			return p
		}
	}
	return nil
}

// ActivePlayer returns the current active player.
func (gs *GameState) ActivePlayer() *PlayerState {
	return gs.GetPlayerBySeat(gs.ActiveSeat)
}

// PlayerCount returns the number of players.
func (gs *GameState) PlayerCount() int {
	return len(gs.Players)
}

// NextSeat returns the next seat index (wraps around).
func (gs *GameState) NextSeat(seat int) int {
	return (seat + 1) % gs.PlayerCount()
}

// CurrentRow returns the 0-based row index for the current reveal round.
func (gs *GameState) CurrentRow() int {
	return gs.Round
}

// InvalidatePeekInfoAt removes stale peek info for the given coordinates across all players.
func (gs *GameState) InvalidatePeekInfoAt(coords []SlotCoord) {
	for _, p := range gs.Players {
		for _, c := range coords {
			delete(p.PeekInfo, c)
		}
	}
}

// InvalidateAllPeekInfo removes all peek info for all players.
func (gs *GameState) InvalidateAllPeekInfo() {
	for _, p := range gs.Players {
		for k := range p.PeekInfo {
			delete(p.PeekInfo, k)
		}
	}
}

// ValidateSlotInSet checks that a slot coordinate is in the valid set.
func ValidateSlotInSet(slot SlotCoord, valid []SlotCoord) bool {
	for _, v := range valid {
		if v.Row == slot.Row && v.Position == slot.Position {
			return true
		}
	}
	return false
}

// ExcludeSlot returns a new slice with the given coordinate removed.
func ExcludeSlot(slots []SlotCoord, exclude SlotCoord) []SlotCoord {
	result := make([]SlotCoord, 0, len(slots))
	for _, s := range slots {
		if s.Row != exclude.Row || s.Position != exclude.Position {
			result = append(result, s)
		}
	}
	return result
}

// ValidateRowInSet checks that a row index is in the valid set.
func ValidateRowInSet(row int, valid []int) bool {
	for _, v := range valid {
		if v == row {
			return true
		}
	}
	return false
}
