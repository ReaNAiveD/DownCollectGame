package core

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Engine drives the game state machine.
type Engine struct {
	State   *GameState
	Hooks   *HookRegistry
	Effects map[Rank]CardEffect
	Scoring map[Rank]ScoringRule
}

func NewEngine(config GameConfig) *Engine {
	e := &Engine{
		State:   NewGameState(config),
		Hooks:   NewHookRegistry(),
		Effects: make(map[Rank]CardEffect),
		Scoring: make(map[Rank]ScoringRule),
	}
	e.registerDefaultEffects()
	e.registerDefaultScoringRules()
	return e
}

func (e *Engine) registerDefaultEffects() {
	for _, eff := range defaultEffects() {
		e.Effects[eff.Rank()] = eff
	}
}

func (e *Engine) registerDefaultScoringRules() {
	for rank, rule := range defaultScoringRules() {
		e.Scoring[rank] = rule
	}
}

// GetEffect returns the effect for a rank from this engine's registry.
func (e *Engine) GetEffect(r Rank) CardEffect {
	return e.Effects[r]
}

// GetScoringRule returns the scoring rule for a rank from this engine's registry.
func (e *Engine) GetScoringRule(r Rank) ScoringRule {
	return e.Scoring[r]
}

// AddPlayer adds a player to the game. Returns error if game already started or full.
func (e *Engine) AddPlayer(playerID string) error {
	gs := e.State
	if gs.Phase != PhaseWaitingPlayers {
		return fmt.Errorf("cannot join: game already started")
	}
	if gs.PlayerCount() >= gs.Config.MaxPlayers {
		return fmt.Errorf("cannot join: room is full")
	}
	if gs.GetPlayer(playerID) != nil {
		return fmt.Errorf("player already in room")
	}
	seat := gs.PlayerCount()
	gs.Players = append(gs.Players, NewPlayerState(playerID, seat))
	return nil
}

// StartGame transitions from WaitingPlayers to CharacterSelect.
func (e *Engine) StartGame() error {
	gs := e.State
	if gs.Phase != PhaseWaitingPlayers {
		return fmt.Errorf("game not in waiting state")
	}
	if gs.PlayerCount() < gs.Config.MinPlayers {
		return fmt.Errorf("not enough players: need %d, have %d", gs.Config.MinPlayers, gs.PlayerCount())
	}

	// For now, skip character selection (all Noop) and go straight to setup
	gs.Phase = PhaseCharacterSelect
	return e.skipCharacterSelect()
}

// skipCharacterSelect assigns Noop characters, registers hooks, and moves to setup.
func (e *Engine) skipCharacterSelect() error {
	gs := e.State
	for _, p := range gs.Players {
		p.Character = &NoopCharacter{}
		p.Character.RegisterHooks(e.Hooks, p.PlayerID)
	}
	return e.setupGame()
}

// ChooseCharacter handles a player's character choice. Currently noop.
func (e *Engine) ChooseCharacter(playerID string, characterIndex int) error {
	// Noop: character already assigned
	return nil
}

// setupGame initializes the deck, removes cards, deals initial hands, moves to reveal phase.
func (e *Engine) setupGame() error {
	gs := e.State
	gs.Phase = PhaseSetup

	// Create and shuffle deck
	allCards := NewStandardDeck()
	gs.Deck = NewDeck(allCards)
	gs.Deck.Shuffle()

	// Remove random cards
	gs.RemovedPile = gs.Deck.RemoveRandom(gs.Config.RemovedCards)

	// Shuffle again after removal
	gs.Deck.Shuffle()

	// Deal initial hands
	for i := 0; i < gs.Config.InitialHandSize; i++ {
		for _, p := range gs.Players {
			if card, ok := gs.Deck.DealOne(); ok {
				p.Hand.AddRight(card)
			}
		}
	}

	// Initialize board
	gs.Board = NewBoard()

	// Pick random starting player
	startBig, _ := rand.Int(rand.Reader, big.NewInt(int64(gs.PlayerCount())))
	gs.StartingSeat = int(startBig.Int64())

	// Enter reveal phase
	return e.enterRevealPhase()
}

// enterRevealPhase transitions to the reveal phase.
func (e *Engine) enterRevealPhase() error {
	gs := e.State
	gs.Phase = PhaseReveal
	gs.Round = 0
	gs.TurnInRound = 0
	gs.ActiveSeat = gs.StartingSeat
	gs.TurnStep = StepRevealCard

	e.Hooks.Trigger(OnRevealPhaseStart, gs, HookContext{})
	e.Hooks.Trigger(OnRevealRoundStart, gs, HookContext{})

	// Set pending action for active player to reveal
	e.setPendingReveal()
	return nil
}

func (e *Engine) setPendingReveal() {
	gs := e.State
	player := gs.ActivePlayer()
	e.Hooks.Trigger(OnRevealTurnStart, gs, HookContext{PlayerID: player.PlayerID})

	gs.TurnStep = StepRevealCard
	gs.PendingAction = &PendingAction{
		Type:     ActionRevealCard,
		PlayerID: player.PlayerID,
	}
}

// HandleAction processes a player action. This is the main input entry point.
func (e *Engine) HandleAction(playerID string, choice PlayerChoice) error {
	gs := e.State

	if gs.PendingAction == nil {
		return fmt.Errorf("no action pending")
	}
	if gs.PendingAction.PlayerID != playerID {
		return fmt.Errorf("not this player's turn")
	}

	switch gs.Phase {
	case PhaseReveal:
		return e.handleRevealAction(playerID, choice)
	case PhasePick:
		return e.handlePickAction(playerID, choice)
	default:
		return fmt.Errorf("no action expected in phase %s", gs.Phase)
	}
}

func (e *Engine) handleRevealAction(playerID string, choice PlayerChoice) error {
	gs := e.State

	switch gs.TurnStep {
	case StepRevealCard:
		return e.doReveal()
	case StepSwapDecision:
		return e.handleSwapWithDrawn(choice)
	case StepRevealConfirm:
		return e.confirmReveal()
	case StepEffectAction:
		return e.resolveEffectChoice(choice)
	default:
		return fmt.Errorf("unexpected turn step: %d", gs.TurnStep)
	}
}

// doReveal draws the top card. Only the active player sees it (private).
// If the player has hand cards, they can swap; otherwise skip to reveal.
func (e *Engine) doReveal() error {
	gs := e.State

	gs.PendingAction = nil
	gs.ClearEvents()

	e.Hooks.Trigger(BeforeRevealCard, gs, HookContext{PlayerID: gs.ActivePlayer().PlayerID})

	card, ok := gs.Deck.DrawTop()
	if !ok {
		return e.endRevealPhase()
	}

	revealCopy := card
	gs.RevealedCard = &revealCopy

	e.Hooks.Trigger(AfterRevealCard, gs, HookContext{
		PlayerID: gs.ActivePlayer().PlayerID,
		Card:     &card,
	})

	player := gs.ActivePlayer()
	if player.Hand.Size() > 0 {
		// Player can swap drawn card with a hand card
		gs.TurnStep = StepSwapDecision
		gs.PendingAction = &PendingAction{
			Type:     ActionSwapDecision,
			PlayerID: player.PlayerID,
		}
		return nil
	}

	// No hand cards — go straight to public reveal
	return e.showAndConfirmReveal()
}

// handleSwapWithDrawn processes the player's decision to swap drawn card with a hand card.
func (e *Engine) handleSwapWithDrawn(choice PlayerChoice) error {
	gs := e.State
	player := gs.ActivePlayer()

	if choice.DoSwap {
		if choice.HandCardIndex < 0 || choice.HandCardIndex >= player.Hand.Size() {
			return fmt.Errorf("invalid hand card index")
		}

		// Swap: hand card goes to RevealedCard (will be executed), drawn card goes to hand
		drawnCard := *gs.RevealedCard
		handCard, _ := player.Hand.RemoveAt(choice.HandCardIndex)

		revealCopy := handCard
		gs.RevealedCard = &revealCopy
		player.Hand.AddRight(drawnCard)

		gs.AddEvent(GameEvent{
			Type:        "swapWithDrawn",
			SeatIndex:   player.SeatIndex,
			Description: "将抽到的卡牌收入手牌，用一张手牌作为本回合需要处理的卡牌",
		})
	} else {
		gs.AddEvent(GameEvent{
			Type:        "swapSkipped",
			SeatIndex:   player.SeatIndex,
			Description: "选择不交换，直接处理抽到的卡牌",
		})
	}

	e.Hooks.Trigger(AfterSwap, gs, HookContext{PlayerID: player.PlayerID})
	return e.showAndConfirmReveal()
}

// showAndConfirmReveal shows the card to ALL players and waits for active player to confirm.
func (e *Engine) showAndConfirmReveal() error {
	gs := e.State

	eventCopy := *gs.RevealedCard
	gs.AddEvent(GameEvent{
		Type:        "cardRevealed",
		SeatIndex:   gs.ActiveSeat,
		Card:        &eventCopy,
		Description: "本回合需要处理的卡牌已展示给所有人",
	})

	gs.TurnStep = StepRevealConfirm
	gs.PendingAction = &PendingAction{
		Type:     ActionConfirmReveal,
		PlayerID: gs.ActivePlayer().PlayerID,
	}
	return nil
}

// confirmReveal processes the confirmation and executes the card effect.
func (e *Engine) confirmReveal() error {
	gs := e.State
	if gs.RevealedCard == nil {
		return fmt.Errorf("no card to confirm")
	}
	card := *gs.RevealedCard
	gs.PendingAction = nil
	gs.ClearEvents()
	return e.executeEffect(card, nil)
}

// executeEffect runs a card's effect, potentially setting up a PendingAction.
func (e *Engine) executeEffect(card Card, choice *PlayerChoice) error {
	gs := e.State

	effect := e.GetEffect(card.Rank)
	if effect == nil {
		// No effect, just place on board
		gs.Board.PlaceDefault(gs.CurrentRow(), &card, false)
		gs.RevealedCard = nil
		return e.afterEffectResolved()
	}

	err := effect.Execute(gs, card, choice)
	if err != nil {
		return err
	}

	// Check if effect needs player input
	if gs.PendingAction != nil && choice == nil {
		gs.TurnStep = StepEffectAction
		return nil
	}

	// Check for recursive reveal (card 4)
	if gs.RecursiveReveal {
		gs.RecursiveReveal = false
		return e.doReveal()
	}

	return e.afterEffectResolved()
}

// resolveEffectChoice processes the player's choice for an effect.
func (e *Engine) resolveEffectChoice(choice PlayerChoice) error {
	gs := e.State

	if gs.PendingAction == nil {
		return e.afterEffectResolved()
	}

	rank := gs.PendingAction.Effect.Rank
	card := Card{Rank: rank}
	if gs.PendingAction.Effect.Card != nil {
		card = *gs.PendingAction.Effect.Card
	}

	effect := e.GetEffect(rank)
	if effect == nil {
		gs.PendingAction = nil
		return e.afterEffectResolved()
	}

	err := effect.Execute(gs, card, &choice)
	if err != nil {
		return err
	}

	// Check if effect still needs more input
	if gs.PendingAction != nil {
		return nil
	}

	// Check for recursive reveal (card 4)
	if gs.RecursiveReveal {
		gs.RecursiveReveal = false
		return e.doReveal()
	}

	return e.afterEffectResolved()
}

// afterEffectResolved runs after a card effect is fully processed.
func (e *Engine) afterEffectResolved() error {
	gs := e.State
	e.Hooks.Trigger(AfterEffectResolved, gs, HookContext{PlayerID: gs.ActivePlayer().PlayerID})
	return e.endRevealTurn()
}

// endRevealTurn finishes the current turn and advances.
func (e *Engine) endRevealTurn() error {
	gs := e.State

	e.Hooks.Trigger(OnRevealTurnEnd, gs, HookContext{PlayerID: gs.ActivePlayer().PlayerID})

	gs.TurnInRound++

	if gs.TurnInRound >= gs.PlayerCount() {
		// Round finished
		e.Hooks.Trigger(OnRevealRoundEnd, gs, HookContext{})

		gs.Round++
		if gs.Round >= gs.Config.RevealRounds {
			// All reveal rounds done
			return e.endRevealPhase()
		}

		// Start new round
		gs.TurnInRound = 0
		gs.StartingSeat = gs.NextSeat(gs.StartingSeat)
		gs.ActiveSeat = gs.StartingSeat
		e.Hooks.Trigger(OnRevealRoundStart, gs, HookContext{})
		e.setPendingReveal()
		return nil
	}

	// Next player in this round
	gs.ActiveSeat = gs.NextSeat(gs.ActiveSeat)
	e.setPendingReveal()
	return nil
}

// endRevealPhase transitions from reveal to pick phase.
func (e *Engine) endRevealPhase() error {
	gs := e.State
	e.Hooks.Trigger(OnRevealPhaseEnd, gs, HookContext{})
	return e.enterPickPhase()
}

// enterPickPhase transitions to the pick phase.
func (e *Engine) enterPickPhase() error {
	gs := e.State
	gs.Phase = PhasePick
	gs.Round = 0
	gs.TurnInRound = 0
	gs.StartingSeat = gs.NextSeat(gs.StartingSeat)
	gs.ActiveSeat = gs.StartingSeat

	e.Hooks.Trigger(OnPickPhaseStart, gs, HookContext{})
	e.Hooks.Trigger(OnPickRoundStart, gs, HookContext{})

	e.setPendingPick()
	return nil
}

func (e *Engine) setPendingPick() {
	gs := e.State
	player := gs.ActivePlayer()
	e.Hooks.Trigger(OnPickTurnStart, gs, HookContext{PlayerID: player.PlayerID})

	allSlots := gs.Board.GetAllSlotCoords()
	gs.PendingAction = &PendingAction{
		Type:     ActionPickCard,
		PlayerID: player.PlayerID,
		Effect:   EffectContext{ValidSlots: allSlots},
	}
}

func (e *Engine) handlePickAction(playerID string, choice PlayerChoice) error {
	gs := e.State
	player := gs.ActivePlayer()

	e.Hooks.Trigger(BeforePickCard, gs, HookContext{PlayerID: player.PlayerID})

	if len(choice.SelectedSlots) != 1 {
		return fmt.Errorf("must select exactly 1 card")
	}
	coord := choice.SelectedSlots[0]
	if !ValidateSlotInSet(coord, gs.PendingAction.Effect.ValidSlots) {
		return fmt.Errorf("invalid slot selection")
	}
	card := gs.Board.RemoveCard(coord)
	if card == nil {
		return fmt.Errorf("no card at selected position")
	}

	player.Hand.AddRight(*card)
	gs.PendingAction = nil

	e.Hooks.Trigger(AfterPickCard, gs, HookContext{
		PlayerID: player.PlayerID,
		Card:     card,
	})

	return e.endPickTurn()
}

func (e *Engine) endPickTurn() error {
	gs := e.State

	e.Hooks.Trigger(OnPickTurnEnd, gs, HookContext{PlayerID: gs.ActivePlayer().PlayerID})

	gs.TurnInRound++

	if gs.TurnInRound >= gs.PlayerCount() {
		e.Hooks.Trigger(OnPickRoundEnd, gs, HookContext{})

		gs.Round++
		if gs.Round >= gs.Config.PickRounds {
			return e.endPickPhase()
		}

		gs.TurnInRound = 0
		gs.StartingSeat = gs.NextSeat(gs.StartingSeat)
		gs.ActiveSeat = gs.StartingSeat
		e.Hooks.Trigger(OnPickRoundStart, gs, HookContext{})
		e.setPendingPick()
		return nil
	}

	gs.ActiveSeat = gs.NextSeat(gs.ActiveSeat)
	e.setPendingPick()
	return nil
}

func (e *Engine) endPickPhase() error {
	gs := e.State
	e.Hooks.Trigger(OnPickPhaseEnd, gs, HookContext{})
	return e.enterScoringPhase()
}

// enterScoringPhase scores all players and ends the game.
func (e *Engine) enterScoringPhase() error {
	gs := e.State
	gs.Phase = PhaseScoring

	e.Hooks.Trigger(OnScoringPhaseStart, gs, HookContext{})

	for _, p := range gs.Players {
		e.Hooks.Trigger(BeforePlayerScoring, gs, HookContext{PlayerID: p.PlayerID})

		// Character scoring (noop)
		// Private scoring (noop)
		e.Hooks.Trigger(BeforePrivateScoring, gs, HookContext{PlayerID: p.PlayerID})
		e.Hooks.Trigger(AfterPrivateScoring, gs, HookContext{PlayerID: p.PlayerID})

		// Global scoring (noop)
		e.Hooks.Trigger(BeforeGlobalScoring, gs, HookContext{PlayerID: p.PlayerID})
		e.Hooks.Trigger(AfterGlobalScoring, gs, HookContext{PlayerID: p.PlayerID})

		// Card scoring
		CalculatePlayerScore(p, gs.Config.ScoringMode, e.Scoring)

		e.Hooks.Trigger(AfterPlayerScoring, gs, HookContext{PlayerID: p.PlayerID})
	}

	e.Hooks.Trigger(OnScoringPhaseEnd, gs, HookContext{})

	gs.Phase = PhaseFinished
	gs.PendingAction = nil
	return nil
}

// GetWinners returns player IDs of those with the lowest score.
func (e *Engine) GetWinners() []string {
	gs := e.State
	if gs.Phase != PhaseFinished {
		return nil
	}

	minScore := gs.Players[0].Score
	for _, p := range gs.Players[1:] {
		if p.Score < minScore {
			minScore = p.Score
		}
	}

	var winners []string
	for _, p := range gs.Players {
		if p.Score == minScore {
			winners = append(winners, p.PlayerID)
		}
	}
	return winners
}
