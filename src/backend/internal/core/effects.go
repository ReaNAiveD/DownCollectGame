package core

import "fmt"

// CardEffect defines how a card behaves when revealed.
type CardEffect interface {
	Rank() Rank
	Execute(gs *GameState, card Card, choice *PlayerChoice) error
}

var effectRegistry = map[Rank]CardEffect{}

func RegisterEffect(e CardEffect) { effectRegistry[e.Rank()] = e }
func GetEffect(r Rank) CardEffect { return effectRegistry[r] }

// defaultEffects returns a fresh slice of all built-in card effects.
func defaultEffects() []CardEffect {
	return []CardEffect{
		&SmallJokerEffect{},
		&BigJokerEffect{},
		&AceEffect{},
		&TwoEffect{},
		&ThreeEffect{},
		&FourEffect{},
		&FiveEffect{},
		&SixEffect{},
		&SevenEffect{},
		&EightEffect{},
		&NineEffect{},
		&TenEffect{},
		&JackEffect{},
		&QueenEffect{},
		&KingEffect{},
	}
}

func init() {
	for _, e := range defaultEffects() {
		RegisterEffect(e)
	}
}

// --- Small Joker ---

type SmallJokerEffect struct{}

func (e *SmallJokerEffect) Rank() Rank { return RankSmallJoker }
func (e *SmallJokerEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	affected := gs.Board.GetFaceUpSlots()
	gs.Board.SetAllFaceDown()
	gs.Deck.PushBottom(card)
	gs.RevealedCard = nil
	gs.AddEvent(GameEvent{
		Type: "effectSmallJoker", SeatIndex: gs.ActiveSeat,
		Slots:       affected,
		Description: "小Joker：场上所有明置卡牌已翻为暗置，小Joker放回牌堆底部",
	})
	return nil
}

// --- Big Joker ---

type BigJokerEffect struct{}

func (e *BigJokerEffect) Rank() Rank { return RankBigJoker }
func (e *BigJokerEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	affected := gs.Board.GetFaceDownSlots()
	gs.Board.ShuffleAllFaceDown()
	gs.InvalidateAllPeekInfo()
	gs.Deck.PushBottom(card)
	gs.RevealedCard = nil
	gs.AddEvent(GameEvent{
		Type: "effectBigJoker", SeatIndex: gs.ActiveSeat,
		Slots:       affected,
		Description: "大Joker：场上所有暗置卡牌的位置已被随机洗混，大Joker放回牌堆底部",
	})
	return nil
}

// --- Ace ---

type AceEffect struct{}

func (e *AceEffect) Rank() Rank { return RankA }
func (e *AceEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	gs.Deck.PushBottom(card)
	gs.RevealedCard = nil
	if topCard, ok := gs.Deck.DrawTop(); ok {
		coord := gs.Board.PlaceDefault(gs.CurrentRow(), &topCard, false)
		gs.AddEvent(GameEvent{
			Type: "effectAce", SeatIndex: gs.ActiveSeat,
			Slots:       []SlotCoord{coord},
			Description: "A：放回牌堆底部，牌堆顶一张新卡牌已暗置入场",
		})
	} else {
		gs.AddEvent(GameEvent{
			Type: "effectAce", SeatIndex: gs.ActiveSeat,
			Description: "A：放回牌堆底部（牌堆已空，无法补牌）",
		})
	}
	return nil
}

// --- 2 ---

type TwoEffect struct{}

func (e *TwoEffect) Rank() Rank { return Rank2 }
func (e *TwoEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	gs.Board.PlaceDefault(gs.CurrentRow(), &card, false)
	gs.RevealedCard = nil
	affected := gs.Board.GetFaceDownSlotsInRow(gs.CurrentRow())
	gs.Board.ShuffleFaceDownInRow(gs.CurrentRow())
	gs.InvalidatePeekInfoAt(affected)
	gs.AddEvent(GameEvent{
		Type: "effectShuffle", SeatIndex: gs.ActiveSeat,
		Row: IntPtr(gs.CurrentRow()), Slots: affected,
		Description: "2：置入场后，当前行的暗置卡牌位置已被随机洗混",
	})
	return nil
}

// --- 3: select 2 face-down to swap ---

type ThreeEffect struct{}

func (e *ThreeEffect) Rank() Rank { return Rank3 }
func (e *ThreeEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	if choice == nil {
		placedCoord := gs.Board.PlaceDefault(gs.CurrentRow(), &card, false)
		gs.RevealedCard = nil
		faceDown := ExcludeSlot(gs.Board.GetFaceDownSlots(), placedCoord)
		if len(faceDown) < 2 {
			gs.AddEvent(GameEvent{
				Type: "effectSkipped", SeatIndex: gs.ActiveSeat,
				Description: "3：暗置卡牌不足2张，无法执行交换效果",
			})
			return nil
		}
		gs.PendingAction = &PendingAction{
			Type: ActionSelectSlots, PlayerID: gs.ActivePlayer().PlayerID,
			Effect: EffectContext{Rank: Rank3, EffectName: "three_swap", SelectCount: 2, ValidSlots: faceDown},
		}
		return nil
	}
	if len(choice.SelectedSlots) != 2 {
		return fmt.Errorf("must select exactly 2 slots")
	}
	for _, s := range choice.SelectedSlots {
		if !ValidateSlotInSet(s, gs.PendingAction.Effect.ValidSlots) {
			return fmt.Errorf("invalid slot selection")
		}
	}
	gs.Board.SwapCards(choice.SelectedSlots[0], choice.SelectedSlots[1])
	gs.InvalidatePeekInfoAt(choice.SelectedSlots)
	gs.AddEvent(GameEvent{
		Type: "effectSwap", SeatIndex: gs.ActiveSeat,
		Slots: choice.SelectedSlots, Description: "3：展示区两张暗置卡牌已互换位置",
	})
	gs.PendingAction = nil
	return nil
}

// --- 4: place on board face-down, then recursive reveal ---

type FourEffect struct{}

func (e *FourEffect) Rank() Rank { return Rank4 }
func (e *FourEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	coord := gs.Board.PlaceDefault(gs.CurrentRow(), &card, false)
	gs.RevealedCard = nil
	gs.RecursiveReveal = true
	gs.AddEvent(GameEvent{
		Type: "effectFour", SeatIndex: gs.ActiveSeat,
		Slots:       []SlotCoord{coord},
		Description: "4：暗置入场，继续从牌堆顶揭示下一张卡牌",
	})
	return nil
}

// --- 5: replace a face-down slot ---

type FiveEffect struct{}

func (e *FiveEffect) Rank() Rank { return Rank5 }
func (e *FiveEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	return executeReplace(gs, card, choice, Rank5, "five_replace")
}

// --- 6: same as 5 ---

type SixEffect struct{}

func (e *SixEffect) Rank() Rank { return Rank6 }
func (e *SixEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	return executeReplace(gs, card, choice, Rank6, "six_replace")
}

// executeReplace is shared logic for effects 5 and 6.
func executeReplace(gs *GameState, card Card, choice *PlayerChoice, rank Rank, name string) error {
	if choice == nil {
		placedCoord := gs.Board.PlaceDefault(gs.CurrentRow(), &card, false)
		gs.RevealedCard = nil
		faceDown := ExcludeSlot(gs.Board.GetFaceDownSlots(), placedCoord)
		if len(faceDown) == 0 {
			gs.AddEvent(GameEvent{
				Type: "effectSkipped", SeatIndex: gs.ActiveSeat,
				Description: name + "：展示区无可替换的暗置卡牌，效果跳过",
			})
			return nil
		}
		gs.PendingAction = &PendingAction{
			Type: ActionSelectSlots, PlayerID: gs.ActivePlayer().PlayerID,
			Effect: EffectContext{Rank: rank, EffectName: name, SelectCount: 1, ValidSlots: faceDown},
		}
		return nil
	}
	if len(choice.SelectedSlots) != 1 {
		return fmt.Errorf("must select exactly 1 slot")
	}
	if !ValidateSlotInSet(choice.SelectedSlots[0], gs.PendingAction.Effect.ValidSlots) {
		return fmt.Errorf("invalid slot selection")
	}
	coord := choice.SelectedSlots[0]
	gs.InvalidatePeekInfoAt([]SlotCoord{coord})
	removed := gs.Board.RemoveCard(coord)
	if removed != nil {
		gs.Deck.PushBottom(*removed)
	}
	if topCard, ok := gs.Deck.DrawTop(); ok {
		slot := gs.Board.GetSlot(coord)
		if slot != nil {
			slot.Card = &topCard
			slot.FaceUp = false
		}
	}
	gs.AddEvent(GameEvent{
		Type: "effectReplace", SeatIndex: gs.ActiveSeat,
		Slots: []SlotCoord{coord}, Description: "已将一张暗置卡牌移至牌堆底，并用牌堆顶新牌替补",
	})
	gs.PendingAction = nil
	return nil
}

// --- 7: insert at chosen position ---

type SevenEffect struct{}

func (e *SevenEffect) Rank() Rank { return Rank7 }
func (e *SevenEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	return executeInsert(gs, card, choice, Rank7, "seven_insert")
}

// --- 8: same as 7 ---

type EightEffect struct{}

func (e *EightEffect) Rank() Rank { return Rank8 }
func (e *EightEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	return executeInsert(gs, card, choice, Rank8, "eight_insert")
}

// executeInsert is shared logic for effects 7 and 8.
func executeInsert(gs *GameState, card Card, choice *PlayerChoice, rank Rank, name string) error {
	if choice == nil {
		gs.RevealedCard = nil
		maxPos := gs.Board.RowSlotCount(gs.CurrentRow()) + 1
		gs.PendingAction = &PendingAction{
			Type: ActionSelectPosition, PlayerID: gs.ActivePlayer().PlayerID,
			Effect: EffectContext{Rank: rank, Card: &card, EffectName: name, MaxInsertPos: maxPos},
		}
		return nil
	}
	pa := gs.PendingAction
	if choice.SelectedPos < 0 || choice.SelectedPos >= pa.Effect.MaxInsertPos {
		return fmt.Errorf("invalid insert position: %d (max %d)", choice.SelectedPos, pa.Effect.MaxInsertPos-1)
	}
	effectCard := pa.Effect.Card
	if effectCard == nil {
		return fmt.Errorf("missing card in effect context")
	}
	gs.Board.InsertAt(gs.CurrentRow(), choice.SelectedPos, effectCard, false)
	// Invalidate peek info for shifted slots from insertion point onward
	row := gs.CurrentRow()
	rowLen := gs.Board.RowSlotCount(row)
	var shifted []SlotCoord
	for i := choice.SelectedPos; i < rowLen; i++ {
		shifted = append(shifted, SlotCoord{Row: row, Position: i})
	}
	gs.InvalidatePeekInfoAt(shifted)
	gs.AddEvent(GameEvent{
		Type: "effectInsert", SeatIndex: gs.ActiveSeat,
		Row: IntPtr(row), Position: IntPtr(choice.SelectedPos),
		Description: "已将卡牌插入当前行的指定位置，后方卡牌已右移",
	})
	gs.PendingAction = nil
	return nil
}

// --- 9: flip a face-down card face-up ---

type NineEffect struct{}

func (e *NineEffect) Rank() Rank { return Rank9 }
func (e *NineEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	if choice == nil {
		placedCoord := gs.Board.PlaceDefault(gs.CurrentRow(), &card, false)
		gs.RevealedCard = nil
		faceDown := ExcludeSlot(gs.Board.GetFaceDownSlots(), placedCoord)
		if len(faceDown) == 0 {
			gs.AddEvent(GameEvent{
				Type: "effectSkipped", SeatIndex: gs.ActiveSeat,
				Description: "9：展示区无可翻明的暗置卡牌，效果跳过",
			})
			return nil
		}
		gs.PendingAction = &PendingAction{
			Type: ActionSelectSlots, PlayerID: gs.ActivePlayer().PlayerID,
			Effect: EffectContext{Rank: Rank9, EffectName: "nine_flip", SelectCount: 1, ValidSlots: faceDown},
		}
		return nil
	}
	if len(choice.SelectedSlots) != 1 {
		return fmt.Errorf("must select exactly 1 slot")
	}
	if !ValidateSlotInSet(choice.SelectedSlots[0], gs.PendingAction.Effect.ValidSlots) {
		return fmt.Errorf("invalid slot selection")
	}
	gs.Board.SetFaceUp(choice.SelectedSlots[0], true)
	// Capture the flipped card value for the event
	var flippedCard *Card
	flipSlot := gs.Board.GetSlot(choice.SelectedSlots[0])
	if flipSlot != nil && flipSlot.Card != nil {
		c := *flipSlot.Card
		flippedCard = &c
	}
	gs.AddEvent(GameEvent{
		Type: "effectFlip", SeatIndex: gs.ActiveSeat,
		Slots: choice.SelectedSlots, Card: flippedCard,
		Description: "9：已将一张暗置卡牌翻为明置",
	})
	gs.PendingAction = nil
	return nil
}

// --- 10: peek at a face-down card ---

type TenEffect struct{}

func (e *TenEffect) Rank() Rank { return Rank10 }
func (e *TenEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	if choice == nil {
		placedCoord := gs.Board.PlaceDefault(gs.CurrentRow(), &card, false)
		gs.RevealedCard = nil
		faceDown := ExcludeSlot(gs.Board.GetFaceDownSlots(), placedCoord)
		if len(faceDown) == 0 {
			gs.AddEvent(GameEvent{
				Type: "effectSkipped", SeatIndex: gs.ActiveSeat,
				Description: "10：展示区无可偷看的暗置卡牌，效果跳过",
			})
			return nil
		}
		gs.PendingAction = &PendingAction{
			Type: ActionSelectSlots, PlayerID: gs.ActivePlayer().PlayerID,
			Effect: EffectContext{Rank: Rank10, EffectName: "ten_peek", SelectCount: 1, ValidSlots: faceDown},
		}
		return nil
	}
	if len(choice.SelectedSlots) != 1 {
		return fmt.Errorf("must select exactly 1 slot")
	}
	if !ValidateSlotInSet(choice.SelectedSlots[0], gs.PendingAction.Effect.ValidSlots) {
		return fmt.Errorf("invalid slot selection")
	}
	slot := gs.Board.GetSlot(choice.SelectedSlots[0])
	if slot != nil && slot.Card != nil {
		gs.ActivePlayer().PeekInfo[choice.SelectedSlots[0]] = *slot.Card
	}
	gs.AddEvent(GameEvent{
		Type: "effectPeek", SeatIndex: gs.ActiveSeat,
		Slots:       choice.SelectedSlots,
		Description: "10：已偷看一张暗置卡牌（仅操作者可见）",
	})
	gs.PendingAction = nil
	return nil
}

// --- J: flip all in a row ---

type JackEffect struct{}

func (e *JackEffect) Rank() Rank { return RankJ }
func (e *JackEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	if choice == nil {
		gs.Board.PlaceDefault(gs.CurrentRow(), &card, true)
		gs.RevealedCard = nil
		rows := gs.Board.GetNonEmptyRowIndices()
		if len(rows) == 0 {
			gs.AddEvent(GameEvent{
				Type: "effectSkipped", SeatIndex: gs.ActiveSeat,
				Description: "J：展示区无可选行，效果跳过",
			})
			return nil
		}
		gs.PendingAction = &PendingAction{
			Type: ActionSelectRow, PlayerID: gs.ActivePlayer().PlayerID,
			Effect: EffectContext{Rank: RankJ, EffectName: "jack_flip_row", ValidRows: rows},
		}
		return nil
	}
	if !ValidateRowInSet(choice.SelectedRow, gs.PendingAction.Effect.ValidRows) {
		return fmt.Errorf("invalid row selection")
	}
	for p := range gs.Board.Rows[choice.SelectedRow].Slots {
		slot := &gs.Board.Rows[choice.SelectedRow].Slots[p]
		if slot.Card != nil {
			slot.FaceUp = !slot.FaceUp
		}
	}
	gs.AddEvent(GameEvent{
		Type: "effectFlipRow", SeatIndex: gs.ActiveSeat,
		Row: IntPtr(choice.SelectedRow), Description: "J：已将选定行所有卡牌的明暗状态翻转",
	})
	gs.PendingAction = nil
	return nil
}

// --- Q: reveal row, shuffle low cards back ---

type QueenEffect struct{}

func (e *QueenEffect) Rank() Rank { return RankQ }
func (e *QueenEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	if choice == nil {
		gs.Board.PlaceDefault(gs.CurrentRow(), &card, true)
		gs.RevealedCard = nil
		rows := gs.Board.GetNonEmptyRowIndices()
		if len(rows) == 0 {
			gs.AddEvent(GameEvent{
				Type: "effectSkipped", SeatIndex: gs.ActiveSeat,
				Description: "Q：展示区无可选行，效果跳过",
			})
			return nil
		}
		gs.PendingAction = &PendingAction{
			Type: ActionSelectRow, PlayerID: gs.ActivePlayer().PlayerID,
			Effect: EffectContext{Rank: RankQ, EffectName: "queen_reveal_shuffle", ValidRows: rows},
		}
		return nil
	}
	if !ValidateRowInSet(choice.SelectedRow, gs.PendingAction.Effect.ValidRows) {
		return fmt.Errorf("invalid row selection")
	}
	row := &gs.Board.Rows[choice.SelectedRow]
	// Reveal all cards in the row - capture their values for event
	var allRowCards []Card
	var allRowCoords []SlotCoord
	for p := range row.Slots {
		if row.Slots[p].Card != nil {
			row.Slots[p].FaceUp = true
			allRowCards = append(allRowCards, *row.Slots[p].Card)
			allRowCoords = append(allRowCoords, SlotCoord{Row: choice.SelectedRow, Position: p})
		}
	}
	var lowCoords []SlotCoord
	var lowCards []*Card
	for p := range row.Slots {
		if row.Slots[p].Card != nil && row.Slots[p].Card.BaseScore() <= 6 {
			lowCoords = append(lowCoords, SlotCoord{Row: choice.SelectedRow, Position: p})
			lowCards = append(lowCards, row.Slots[p].Card)
		}
	}
	if len(lowCards) > 1 {
		cryptoShuffle(lowCards)
	}
	for i, coord := range lowCoords {
		gs.Board.Rows[coord.Row].Slots[coord.Position].Card = lowCards[i]
		gs.Board.Rows[coord.Row].Slots[coord.Position].FaceUp = false
	}
	gs.InvalidatePeekInfoAt(lowCoords)
	gs.AddEvent(GameEvent{
		Type: "effectQueenShuffle", SeatIndex: gs.ActiveSeat,
		Row: IntPtr(choice.SelectedRow), Slots: allRowCoords, Cards: allRowCards,
		Description: "Q：已展示该行所有卡牌，基础分值≤6的卡牌已洗混暗置放回",
	})
	gs.PendingAction = nil
	return nil
}

// --- K: select one per row, mix with K ---

type KingEffect struct{}

func (e *KingEffect) Rank() Rank { return RankK }
func (e *KingEffect) Execute(gs *GameState, card Card, choice *PlayerChoice) error {
	if choice == nil {
		gs.RevealedCard = nil
		rows := gs.Board.GetNonEmptyRowIndices()
		if len(rows) == 0 {
			gs.Board.PlaceDefault(gs.CurrentRow(), &card, false)
			gs.AddEvent(GameEvent{
				Type: "effectSkipped", SeatIndex: gs.ActiveSeat,
				Description: "K：展示区无卡牌，K直接暗置入场",
			})
			return nil
		}
		gs.PendingAction = &PendingAction{
			Type: ActionSelectSlotsPerRow, PlayerID: gs.ActivePlayer().PlayerID,
			Effect: EffectContext{Rank: RankK, Card: &card, EffectName: "king_mix", ValidRows: rows},
		}
		return nil
	}
	pa := gs.PendingAction
	if len(choice.SelectedSlots) != len(pa.Effect.ValidRows) {
		return fmt.Errorf("must select exactly one slot per valid row")
	}
	selectedRows := make(map[int]bool)
	for _, s := range choice.SelectedSlots {
		if !ValidateRowInSet(s.Row, pa.Effect.ValidRows) {
			return fmt.Errorf("slot row %d not in valid rows", s.Row)
		}
		if selectedRows[s.Row] {
			return fmt.Errorf("duplicate selection in row %d", s.Row)
		}
		selectedRows[s.Row] = true
	}
	effectCard := pa.Effect.Card
	if effectCard == nil {
		return fmt.Errorf("missing card in effect context")
	}

	allCoords := make([]SlotCoord, 0, len(choice.SelectedSlots)+1)
	allCards := make([]*Card, 0, len(choice.SelectedSlots)+1)
	for _, coord := range choice.SelectedSlots {
		slot := gs.Board.GetSlot(coord)
		if slot == nil || slot.Card == nil {
			return fmt.Errorf("no card at selected slot")
		}
		allCoords = append(allCoords, coord)
		allCards = append(allCards, slot.Card)
	}
	kCoord := gs.Board.PlaceDefault(gs.CurrentRow(), effectCard, false)
	allCoords = append(allCoords, kCoord)
	allCards = append(allCards, gs.Board.GetSlot(kCoord).Card)

	for _, coord := range allCoords {
		gs.Board.SetFaceUp(coord, true)
	}
	// Collect revealed card identities before shuffle
	revealedCards := make([]Card, len(allCards))
	for i, c := range allCards {
		revealedCards[i] = *c
	}

	cryptoShuffle(allCards)
	for i, coord := range allCoords {
		slot := gs.Board.GetSlot(coord)
		slot.Card = allCards[i]
		slot.FaceUp = false
	}
	gs.InvalidatePeekInfoAt(allCoords)
	gs.AddEvent(GameEvent{
		Type: "effectKingMix", SeatIndex: gs.ActiveSeat,
		Slots: allCoords, Cards: revealedCards,
		Description: "K：已从每行选取卡牌，连同K一起展示后洗混暗置放回原位",
	})
	gs.PendingAction = nil
	return nil
}
