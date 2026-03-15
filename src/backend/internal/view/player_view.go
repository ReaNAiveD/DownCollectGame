package view

import (
	"github.com/naived/downcollect/internal/core"
)

// CardInfo represents a card visible to a player.
type CardInfo struct {
	ID        int    `json:"id"`
	Suit      string `json:"suit"`
	Rank      string `json:"rank"`
	BaseScore int    `json:"baseScore"`
}

// SlotView represents a board slot as seen by a specific player.
type SlotView struct {
	HasCard    bool      `json:"hasCard"`
	FaceUp     bool      `json:"faceUp"`
	Card       *CardInfo `json:"card,omitempty"`       // non-nil if face-up
	PeekedCard *CardInfo `json:"peekedCard,omitempty"` // non-nil if this player peeked
}

// PlayerInfo shows public info about a player.
type PlayerInfo struct {
	PlayerID  string `json:"playerId"`
	Nickname  string `json:"nickname"`
	SeatIndex int    `json:"seatIndex"`
	HandSize  int    `json:"handSize"`
	IsHost    bool   `json:"isHost"`
	Connected bool   `json:"connected"`
	Character string `json:"character"`
	Score     *int   `json:"score,omitempty"` // only visible in scoring/finished phase
}

// ActionView describes what action the player needs to take.
type ActionView struct {
	Type         string                 `json:"type"`
	ValidSlots   []SlotCoordV           `json:"validSlots,omitempty"`
	ValidRows    []int                  `json:"validRows,omitempty"`
	MaxPositions int                    `json:"maxPositions,omitempty"` // for insert position
	SelectCount  int                    `json:"selectCount,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

type SlotCoordV struct {
	Row      int `json:"row"`
	Position int `json:"position"`
}

// ScoreDetailView shows scoring details for a card.
type ScoreDetailView struct {
	CardIndex   int      `json:"cardIndex"`
	Card        CardInfo `json:"card"`
	RuleName    string   `json:"ruleName"`
	Score       int      `json:"score"`
	Description string   `json:"description"`
	Skipped     bool     `json:"skipped"`
}

// PlayerScoreView shows a player's scoring result.
type PlayerScoreView struct {
	PlayerID   string            `json:"playerId"`
	SeatIndex  int               `json:"seatIndex"`
	TotalScore int               `json:"totalScore"`
	Hand       []CardInfo        `json:"hand"`
	Details    []ScoreDetailView `json:"details"`
}

// RevealedCardView shows the currently revealed card.
type RevealedCardView struct {
	Card *CardInfo `json:"card,omitempty"`
}

// GameEventView is a game event visible to all players.
type GameEventView struct {
	Type        string       `json:"type"`
	SeatIndex   int          `json:"seatIndex"`
	Card        *CardInfo    `json:"card,omitempty"`
	Slots       []SlotCoordV `json:"slots,omitempty"`
	Cards       []CardInfo   `json:"cards,omitempty"`
	Row         *int         `json:"row,omitempty"`
	Position    *int         `json:"position,omitempty"`
	Description string       `json:"description"`
}

// PlayerView is the complete game state from one player's perspective.
type PlayerView struct {
	Phase         string            `json:"phase"`
	Round         int               `json:"round"`
	Turn          int               `json:"turn"`
	ActiveSeat    int               `json:"activeSeat"`
	MySeat        int               `json:"mySeat"`
	ScoringMode   int               `json:"scoringMode"`
	MyHand        []CardInfo        `json:"myHand"`
	Players       []PlayerInfo      `json:"players"`
	Board         [][]SlotView      `json:"board"`
	DeckRemaining int               `json:"deckRemaining"`
	PendingAction *ActionView       `json:"pendingAction,omitempty"`
	RevealedCard  *CardInfo         `json:"revealedCard,omitempty"`
	Events        []GameEventView   `json:"events,omitempty"`
	Scores        []PlayerScoreView `json:"scores,omitempty"`
	Winners       []string          `json:"winners,omitempty"`
}

// GeneratePlayerView creates a PlayerView for a specific player.
func GeneratePlayerView(engine *core.Engine, playerID string, nicknames map[string]string, hostID string, connected map[string]bool) PlayerView {
	gs := engine.State
	player := gs.GetPlayer(playerID)
	if player == nil {
		return PlayerView{}
	}

	pv := PlayerView{
		Phase:         gs.Phase.String(),
		Round:         gs.Round,
		Turn:          gs.TurnInRound,
		ActiveSeat:    gs.ActiveSeat,
		MySeat:        player.SeatIndex,
		ScoringMode:   int(gs.Config.ScoringMode),
		DeckRemaining: 0,
	}

	if gs.Deck != nil {
		pv.DeckRemaining = gs.Deck.Remaining()
	}

	// My hand
	pv.MyHand = make([]CardInfo, player.Hand.Size())
	for i, c := range player.Hand.Cards {
		pv.MyHand[i] = cardToInfo(c)
	}

	// Players
	pv.Players = make([]PlayerInfo, len(gs.Players))
	for i, p := range gs.Players {
		pi := PlayerInfo{
			PlayerID:  p.PlayerID,
			Nickname:  nicknames[p.PlayerID],
			SeatIndex: p.SeatIndex,
			HandSize:  p.Hand.Size(),
			IsHost:    p.PlayerID == hostID,
			Connected: connected[p.PlayerID],
			Character: p.Character.Name(),
		}
		if gs.Phase == core.PhaseFinished {
			score := p.Score
			pi.Score = &score
		}
		pv.Players[i] = pi
	}

	// Board
	if gs.Board != nil {
		pv.Board = make([][]SlotView, len(gs.Board.Rows))
		for r, row := range gs.Board.Rows {
			pv.Board[r] = make([]SlotView, len(row.Slots))
			for p, slot := range row.Slots {
				sv := SlotView{
					HasCard: slot.Card != nil,
					FaceUp:  slot.FaceUp,
				}
				if slot.Card != nil && slot.FaceUp {
					info := cardToInfo(*slot.Card)
					sv.Card = &info
				}
				// Check if this player has peeked at this slot
				coord := core.SlotCoord{Row: r, Position: p}
				if peeked, ok := player.PeekInfo[coord]; ok {
					info := cardToInfo(peeked)
					sv.PeekedCard = &info
				}
				pv.Board[r][p] = sv
			}
		}
	}

	// Revealed card visibility:
	// During StepSwapDecision, only the active player sees the drawn card.
	// During StepRevealConfirm and after, everyone sees it.
	if gs.RevealedCard != nil {
		if gs.TurnStep != core.StepSwapDecision || gs.ActivePlayer().PlayerID == playerID {
			info := cardToInfo(*gs.RevealedCard)
			pv.RevealedCard = &info
		}
	}

	// Pending action (only if it's this player's action)
	if gs.PendingAction != nil && gs.PendingAction.PlayerID == playerID {
		pv.PendingAction = buildActionView(gs.PendingAction)
	}

	// Events visible to all players
	if len(gs.EventLog) > 0 {
		pv.Events = make([]GameEventView, len(gs.EventLog))
		for i, evt := range gs.EventLog {
			ev := GameEventView{
				Type:        evt.Type,
				SeatIndex:   evt.SeatIndex,
				Row:         evt.Row,
				Position:    evt.Position,
				Description: evt.Description,
			}
			if evt.Card != nil {
				info := cardToInfo(*evt.Card)
				ev.Card = &info
			}
			if len(evt.Slots) > 0 {
				ev.Slots = coordsToView(evt.Slots)
			}
			if len(evt.Cards) > 0 {
				ev.Cards = make([]CardInfo, len(evt.Cards))
				for j, c := range evt.Cards {
					ev.Cards[j] = cardToInfo(c)
				}
			}
			pv.Events[i] = ev
		}
	}

	// Scores (only in finished phase)
	if gs.Phase == core.PhaseFinished {
		pv.Scores = make([]PlayerScoreView, len(gs.Players))
		for i, p := range gs.Players {
			sv := PlayerScoreView{
				PlayerID:   p.PlayerID,
				SeatIndex:  p.SeatIndex,
				TotalScore: p.Score,
				Hand:       make([]CardInfo, p.Hand.Size()),
				Details:    make([]ScoreDetailView, len(p.ScoreDetails)),
			}
			for j, c := range p.Hand.Cards {
				sv.Hand[j] = cardToInfo(c)
			}
			for j, d := range p.ScoreDetails {
				sv.Details[j] = ScoreDetailView{
					CardIndex:   d.CardIndex,
					Card:        cardToInfo(d.Card),
					RuleName:    d.RuleName,
					Score:       d.Score,
					Description: d.Description,
					Skipped:     d.Skipped,
				}
			}
			pv.Scores[i] = sv
		}
		pv.Winners = engine.GetWinners()
	}

	return pv
}

func cardToInfo(c core.Card) CardInfo {
	return CardInfo{
		ID:        c.ID,
		Suit:      c.Suit.String(),
		Rank:      c.Rank.String(),
		BaseScore: c.BaseScore(),
	}
}

func buildActionView(pa *core.PendingAction) *ActionView {
	av := &ActionView{
		Context: make(map[string]interface{}),
	}

	ec := pa.Effect

	switch pa.Type {
	case core.ActionRevealCard:
		av.Type = "revealCard"
	case core.ActionConfirmReveal:
		av.Type = "confirmReveal"
	case core.ActionSelectSlots:
		av.Type = "selectSlots"
		av.ValidSlots = coordsToView(ec.ValidSlots)
		av.SelectCount = ec.SelectCount
	case core.ActionSelectPosition:
		av.Type = "selectPosition"
		av.MaxPositions = ec.MaxInsertPos
	case core.ActionSelectRow:
		av.Type = "selectRow"
		av.ValidRows = ec.ValidRows
	case core.ActionSelectSlotsPerRow:
		av.Type = "selectSlotsPerRow"
		av.ValidRows = ec.ValidRows
	case core.ActionSwapDecision:
		av.Type = "swapDecision"
	case core.ActionPickCard:
		av.Type = "pickCard"
		av.ValidSlots = coordsToView(ec.ValidSlots)
	case core.ActionChooseCharacter:
		av.Type = "chooseCharacter"
	}

	if ec.EffectName != "" {
		av.Context["effect"] = ec.EffectName
	}

	return av
}

func coordsToView(coords []core.SlotCoord) []SlotCoordV {
	if coords == nil {
		return nil
	}
	result := make([]SlotCoordV, len(coords))
	for i, s := range coords {
		result[i] = SlotCoordV{Row: s.Row, Position: s.Position}
	}
	return result
}
