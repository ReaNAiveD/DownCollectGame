package core

import (
	"testing"
)

func TestNewStandardDeck(t *testing.T) {
	cards := NewStandardDeck()
	if len(cards) != 54 {
		t.Fatalf("expected 54 cards, got %d", len(cards))
	}

	// Check jokers
	if cards[0].Rank != RankSmallJoker {
		t.Errorf("card 0 should be SmallJoker, got %s", cards[0].Rank)
	}
	if cards[1].Rank != RankBigJoker {
		t.Errorf("card 1 should be BigJoker, got %s", cards[1].Rank)
	}

	// Check base scores
	if RankA.BaseScore() != 1 {
		t.Errorf("Ace base score should be 1")
	}
	if RankK.BaseScore() != 13 {
		t.Errorf("King base score should be 13")
	}
	if RankSmallJoker.BaseScore() != 0 {
		t.Errorf("SmallJoker base score should be 0")
	}
}

func TestDeckShuffle(t *testing.T) {
	cards := NewStandardDeck()
	deck := NewDeck(cards)
	deck.Shuffle()

	if deck.Remaining() != 54 {
		t.Fatalf("deck should have 54 cards, got %d", deck.Remaining())
	}

	card, ok := deck.DrawTop()
	if !ok {
		t.Fatal("should be able to draw from deck")
	}
	if deck.Remaining() != 53 {
		t.Errorf("deck should have 53 after draw, got %d", deck.Remaining())
	}

	deck.PushBottom(card)
	if deck.Remaining() != 54 {
		t.Errorf("deck should have 54 after push, got %d", deck.Remaining())
	}
}

func TestHand(t *testing.T) {
	hand := NewHand()
	c1 := Card{ID: 0, Suit: SuitHearts, Rank: RankA}
	c2 := Card{ID: 1, Suit: SuitSpades, Rank: Rank5}

	hand.AddRight(c1)
	hand.AddRight(c2)

	if hand.Size() != 2 {
		t.Fatalf("hand should have 2 cards, got %d", hand.Size())
	}

	visible := hand.VisibleFrom(1)
	if len(visible) != 2 {
		t.Errorf("visible from index 1 should be 2, got %d", len(visible))
	}

	left := hand.LeftOf(1)
	if len(left) != 1 {
		t.Errorf("left of index 1 should be 1, got %d", len(left))
	}
}

func TestBoard(t *testing.T) {
	board := NewBoard()
	c := Card{ID: 0, Suit: SuitHearts, Rank: RankA}

	coord := board.PlaceDefault(0, &c, false)
	if coord.Row != 0 || coord.Position != 0 {
		t.Errorf("expected (0,0), got (%d,%d)", coord.Row, coord.Position)
	}

	slot := board.GetSlot(coord)
	if slot == nil || slot.Card == nil {
		t.Fatal("slot should have a card")
	}
	if slot.FaceUp {
		t.Error("should be face down")
	}

	board.FlipSlot(coord)
	if !slot.FaceUp {
		t.Error("should be face up after flip")
	}
}

func TestBoardInsert(t *testing.T) {
	board := NewBoard()
	c1 := Card{ID: 0, Suit: SuitHearts, Rank: RankA}
	c2 := Card{ID: 1, Suit: SuitSpades, Rank: Rank2}
	c3 := Card{ID: 2, Suit: SuitClubs, Rank: Rank3}

	board.PlaceDefault(0, &c1, false)
	board.PlaceDefault(0, &c2, false)
	board.InsertAt(0, 1, &c3, false) // insert at position 1

	if board.RowSlotCount(0) != 3 {
		t.Fatalf("row should have 3 slots, got %d", board.RowSlotCount(0))
	}

	// c1 at 0, c3 at 1, c2 at 2
	s0 := board.GetSlot(SlotCoord{Row: 0, Position: 0})
	s1 := board.GetSlot(SlotCoord{Row: 0, Position: 1})
	s2 := board.GetSlot(SlotCoord{Row: 0, Position: 2})

	if s0.Card.ID != 0 {
		t.Errorf("position 0 should be card 0, got %d", s0.Card.ID)
	}
	if s1.Card.ID != 2 {
		t.Errorf("position 1 should be card 2, got %d", s1.Card.ID)
	}
	if s2.Card.ID != 1 {
		t.Errorf("position 2 should be card 1, got %d", s2.Card.ID)
	}
}

func TestEngineBasicFlow(t *testing.T) {
	config := DefaultGameConfig()
	config.RevealRounds = 3
	config.PickRounds = 2

	engine := NewEngine(config)

	// Add players
	if err := engine.AddPlayer("p1"); err != nil {
		t.Fatal(err)
	}
	if err := engine.AddPlayer("p2"); err != nil {
		t.Fatal(err)
	}

	// Start game
	if err := engine.StartGame(); err != nil {
		t.Fatal(err)
	}

	gs := engine.State
	if gs.Phase != PhaseReveal {
		t.Fatalf("expected Reveal phase, got %s", gs.Phase)
	}

	// Process reveal phase: handle all actions until phase changes
	maxOuter := 100
	for i := 0; i < maxOuter && gs.Phase == PhaseReveal; i++ {
		if gs.PendingAction == nil {
			t.Fatal("expected pending action")
		}
		playerID := gs.PendingAction.PlayerID
		actionType := gs.PendingAction.Type

		t.Logf("Outer iter %d: playerID=%s actionType=%d turnStep=%d phase=%s", i, playerID, actionType, gs.TurnStep, gs.Phase)

		switch actionType {
		case ActionRevealCard, ActionConfirmReveal:
			if err := engine.HandleAction(playerID, PlayerChoice{}); err != nil {
				t.Fatalf("action failed: %v", err)
			}
		case ActionSwapDecision:
			if err := engine.HandleAction(playerID, PlayerChoice{DoSwap: false}); err != nil {
				t.Fatalf("swap failed: %v", err)
			}
		default:
			// Effect choice
			choice := makeAutoChoice(gs)
			t.Logf("  Effect choice: actionType=%d choice=%+v", actionType, choice)
			if err := engine.HandleAction(playerID, choice); err != nil {
				t.Fatalf("effect choice failed: %v", err)
			}
		}
	}

	// Should be in Pick phase now
	if gs.Phase != PhasePick {
		t.Fatalf("expected Pick phase, got %s", gs.Phase)
	}

	// Process pick turns
	for i := 0; i < 10 && gs.Phase == PhasePick; i++ {
		if gs.PendingAction == nil {
			t.Fatal("expected pending action for pick")
		}
		playerID := gs.PendingAction.PlayerID

		allSlots := gs.Board.GetAllSlotCoords()
		if len(allSlots) == 0 {
			t.Fatal("no cards to pick")
		}
		err := engine.HandleAction(playerID, PlayerChoice{
			SelectedSlots: []SlotCoord{allSlots[0]},
		})
		if err != nil {
			t.Fatalf("pick failed: %v", err)
		}
	}

	// Should be finished
	if gs.Phase != PhaseFinished {
		t.Fatalf("expected Finished phase, got %s", gs.Phase)
	}

	// Check scores calculated
	for _, p := range gs.Players {
		if len(p.ScoreDetails) == 0 {
			t.Errorf("player %s should have score details", p.PlayerID)
		}
	}

	winners := engine.GetWinners()
	if len(winners) == 0 {
		t.Error("should have at least one winner")
	}
}

// makeAutoChoice generates a valid choice for the current pending action.
func makeAutoChoice(gs *GameState) PlayerChoice {
	pa := gs.PendingAction
	choice := PlayerChoice{}

	switch pa.Type {
	case ActionSelectSlots:
		count := pa.Effect.SelectCount
		for i := 0; i < count && i < len(pa.Effect.ValidSlots); i++ {
			choice.SelectedSlots = append(choice.SelectedSlots, pa.Effect.ValidSlots[i])
		}
	case ActionSelectPosition:
		choice.SelectedPos = 0
	case ActionSelectRow:
		if len(pa.Effect.ValidRows) > 0 {
			choice.SelectedRow = pa.Effect.ValidRows[0]
		}
	case ActionSelectSlotsPerRow:
		for _, row := range pa.Effect.ValidRows {
			slots := gs.Board.GetSlotsInRow(row)
			if len(slots) > 0 {
				choice.SelectedSlots = append(choice.SelectedSlots, slots[0])
			}
		}
	}

	return choice
}

func TestScoring(t *testing.T) {
	// Test basic scoring: A at leftmost = -3
	player := NewPlayerState("test", 0)
	player.Hand.AddRight(Card{ID: 0, Suit: SuitSpades, Rank: RankA})

	CalculatePlayerScore(player, ScoringModeSpecial, defaultScoringRules())
	if player.Score != -3 {
		t.Errorf("expected score -3, got %d", player.Score)
	}

	// Test: A (left), 5 (right) -> 5 visible=2 < 3 -> 5pts; A not leftmost -> 1pt = 6
	player2 := NewPlayerState("test2", 0)
	player2.Hand.AddRight(Card{ID: 0, Suit: SuitSpades, Rank: RankA})
	player2.Hand.AddRight(Card{ID: 1, Suit: SuitHearts, Rank: Rank5})

	CalculatePlayerScore(player2, ScoringModeSpecial, defaultScoringRules())
	// 5: visible=2 < 3 -> 5pts; A: leftmost -> -3
	if player2.Score != 2 { // 5 + (-3) = 2
		t.Errorf("expected score 2, got %d", player2.Score)
	}
}

func TestScoringJackFlush(t *testing.T) {
	player := NewPlayerState("test", 0)
	player.Hand.AddRight(Card{ID: 0, Suit: SuitHearts, Rank: Rank2})
	player.Hand.AddRight(Card{ID: 1, Suit: SuitHearts, Rank: Rank5})
	player.Hand.AddRight(Card{ID: 2, Suit: SuitHearts, Rank: RankJ})

	CalculatePlayerScore(player, ScoringModeSpecial, defaultScoringRules())
	// J: all left are Hearts -> flush triggered
	// score = 11 + (2+5)/4 = 11 + 1 = 12; left cards skipped
	if player.Score != 12 {
		t.Errorf("expected score 12, got %d (details: %+v)", player.Score, player.ScoreDetails)
	}
}

func TestScoringSmallJokerFog(t *testing.T) {
	player := NewPlayerState("test", 0)
	player.Hand.AddRight(Card{ID: 0, Suit: SuitSpades, Rank: RankA})
	player.Hand.AddRight(Card{ID: 1, Suit: SuitNone, Rank: RankSmallJoker})
	player.Hand.AddRight(Card{ID: 2, Suit: SuitClubs, Rank: RankK})

	CalculatePlayerScore(player, ScoringModeSpecial, defaultScoringRules())
	// K: left has 2 cards, but not 3 suits (None + Spades) -> 13
	// SmallJoker: 0, fogs A
	// A: fogged -> base score only = 1
	if player.Score != 14 { // 13 + 0 + 1
		t.Errorf("expected score 14, got %d (details: %+v)", player.Score, player.ScoreDetails)
	}
}

func TestScoringBaseOnly(t *testing.T) {
	player := NewPlayerState("test", 0)
	player.Hand.AddRight(Card{ID: 0, Suit: SuitSpades, Rank: RankA})        // base=1
	player.Hand.AddRight(Card{ID: 1, Suit: SuitNone, Rank: RankSmallJoker}) // base=0
	player.Hand.AddRight(Card{ID: 2, Suit: SuitHearts, Rank: RankJ})        // base=11

	CalculatePlayerScore(player, ScoringModeBaseOnly, defaultScoringRules())
	// No special rules, no fog, no skip: 1 + 0 + 11 = 12
	if player.Score != 12 {
		t.Errorf("expected base-only score 12, got %d", player.Score)
	}
	for _, d := range player.ScoreDetails {
		if d.RuleName != "基础分值" {
			t.Errorf("expected all rules to be base score, got %s", d.RuleName)
		}
	}
}

func TestScoringMultiSmallJokerFog(t *testing.T) {
	player := NewPlayerState("test", 0)
	player.Hand.AddRight(Card{ID: 0, Suit: SuitSpades, Rank: RankA})        // index 0, fogged by SJ at 1
	player.Hand.AddRight(Card{ID: 1, Suit: SuitNone, Rank: RankSmallJoker}) // index 1, fogged by SJ at 2
	player.Hand.AddRight(Card{ID: 2, Suit: SuitNone, Rank: RankSmallJoker}) // index 2
	player.Hand.AddRight(Card{ID: 3, Suit: SuitHearts, Rank: Rank5})        // index 3

	CalculatePlayerScore(player, ScoringModeSpecial, defaultScoringRules())
	// Card 3 (5): visible=4 >= 3 -> 聚众 2pts
	// Card 2 (SmallJoker): 0pts, fogs card 1
	// Card 1 (SmallJoker): fogged -> base=0, fogs card 0
	// Card 0 (A): fogged -> base=1
	// Total: 2 + 0 + 0 + 1 = 3
	if player.Score != 3 {
		t.Errorf("expected multi-fog score 3, got %d (details: %+v)", player.Score, player.ScoreDetails)
	}
}
