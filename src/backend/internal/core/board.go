package core

import (
	"crypto/rand"
	"math/big"
)

// SlotCoord identifies a position on the board.
type SlotCoord struct {
	Row      int `json:"row"`
	Position int `json:"position"`
}

// CardSlot holds a card and its face-up/down state on the board.
type CardSlot struct {
	Card   *Card `json:"card,omitempty"`
	FaceUp bool  `json:"faceUp"`
}

// Row represents a single row on the board.
type Row struct {
	Slots []CardSlot
}

// Board represents the card display area with multiple rows.
type Board struct {
	Rows []Row
}

func NewBoard() *Board {
	return &Board{Rows: make([]Row, 0)}
}

// EnsureRow ensures that at least rowIndex+1 rows exist.
func (b *Board) EnsureRow(rowIndex int) {
	for len(b.Rows) <= rowIndex {
		b.Rows = append(b.Rows, Row{Slots: make([]CardSlot, 0)})
	}
}

// PlaceDefault places a card at the leftmost empty position in the given row (face down).
func (b *Board) PlaceDefault(rowIndex int, card *Card, faceUp bool) SlotCoord {
	b.EnsureRow(rowIndex)
	pos := len(b.Rows[rowIndex].Slots)
	b.Rows[rowIndex].Slots = append(b.Rows[rowIndex].Slots, CardSlot{Card: card, FaceUp: faceUp})
	return SlotCoord{Row: rowIndex, Position: pos}
}

// InsertAt inserts a card at a specific position in a row, shifting existing cards right.
func (b *Board) InsertAt(rowIndex int, pos int, card *Card, faceUp bool) SlotCoord {
	b.EnsureRow(rowIndex)
	row := &b.Rows[rowIndex]
	if pos > len(row.Slots) {
		pos = len(row.Slots)
	}
	slot := CardSlot{Card: card, FaceUp: faceUp}
	row.Slots = append(row.Slots, CardSlot{})
	copy(row.Slots[pos+1:], row.Slots[pos:])
	row.Slots[pos] = slot
	return SlotCoord{Row: rowIndex, Position: pos}
}

// GetSlot returns the slot at the given coordinate.
func (b *Board) GetSlot(coord SlotCoord) *CardSlot {
	if coord.Row < 0 || coord.Row >= len(b.Rows) {
		return nil
	}
	row := &b.Rows[coord.Row]
	if coord.Position < 0 || coord.Position >= len(row.Slots) {
		return nil
	}
	return &row.Slots[coord.Position]
}

// RemoveCard removes and returns the card at the given coordinate.
func (b *Board) RemoveCard(coord SlotCoord) *Card {
	slot := b.GetSlot(coord)
	if slot == nil || slot.Card == nil {
		return nil
	}
	card := slot.Card
	slot.Card = nil
	return card
}

// FlipSlot toggles the face-up state of a slot.
func (b *Board) FlipSlot(coord SlotCoord) {
	slot := b.GetSlot(coord)
	if slot != nil && slot.Card != nil {
		slot.FaceUp = !slot.FaceUp
	}
}

// SetFaceUp sets the face state of a slot.
func (b *Board) SetFaceUp(coord SlotCoord, faceUp bool) {
	slot := b.GetSlot(coord)
	if slot != nil && slot.Card != nil {
		slot.FaceUp = faceUp
	}
}

// GetAllSlotCoords returns coordinates of all non-empty slots.
func (b *Board) GetAllSlotCoords() []SlotCoord {
	var coords []SlotCoord
	for r, row := range b.Rows {
		for p, slot := range row.Slots {
			if slot.Card != nil {
				coords = append(coords, SlotCoord{Row: r, Position: p})
			}
		}
	}
	return coords
}

// GetFaceDownSlots returns coordinates of all face-down non-empty slots.
func (b *Board) GetFaceDownSlots() []SlotCoord {
	var coords []SlotCoord
	for r, row := range b.Rows {
		for p, slot := range row.Slots {
			if slot.Card != nil && !slot.FaceUp {
				coords = append(coords, SlotCoord{Row: r, Position: p})
			}
		}
	}
	return coords
}

// GetFaceDownSlotsInRow returns face-down slot coordinates in a specific row.
func (b *Board) GetFaceDownSlotsInRow(rowIndex int) []SlotCoord {
	if rowIndex < 0 || rowIndex >= len(b.Rows) {
		return nil
	}
	var coords []SlotCoord
	for p, slot := range b.Rows[rowIndex].Slots {
		if slot.Card != nil && !slot.FaceUp {
			coords = append(coords, SlotCoord{Row: rowIndex, Position: p})
		}
	}
	return coords
}

// GetFaceUpSlots returns coordinates of all face-up non-empty slots.
func (b *Board) GetFaceUpSlots() []SlotCoord {
	var coords []SlotCoord
	for r, row := range b.Rows {
		for p, slot := range row.Slots {
			if slot.Card != nil && slot.FaceUp {
				coords = append(coords, SlotCoord{Row: r, Position: p})
			}
		}
	}
	return coords
}

// GetSlotsInRow returns all non-empty slot coordinates in a row.
func (b *Board) GetSlotsInRow(rowIndex int) []SlotCoord {
	if rowIndex < 0 || rowIndex >= len(b.Rows) {
		return nil
	}
	var coords []SlotCoord
	for p, slot := range b.Rows[rowIndex].Slots {
		if slot.Card != nil {
			coords = append(coords, SlotCoord{Row: rowIndex, Position: p})
		}
	}
	return coords
}

// RightmostInRow returns the rightmost occupied slot coordinate in a row.
func (b *Board) RightmostInRow(rowIndex int) (SlotCoord, bool) {
	if rowIndex < 0 || rowIndex >= len(b.Rows) {
		return SlotCoord{}, false
	}
	row := &b.Rows[rowIndex]
	for p := len(row.Slots) - 1; p >= 0; p-- {
		if row.Slots[p].Card != nil {
			return SlotCoord{Row: rowIndex, Position: p}, true
		}
	}
	return SlotCoord{}, false
}

// RowCount returns the number of rows.
func (b *Board) RowCount() int {
	return len(b.Rows)
}

// RowSlotCount returns the number of slots in a row.
func (b *Board) RowSlotCount(rowIndex int) int {
	if rowIndex < 0 || rowIndex >= len(b.Rows) {
		return 0
	}
	return len(b.Rows[rowIndex].Slots)
}

// ShuffleFaceDownInRow collects all face-down cards in a row, shuffles, and places them back.
func (b *Board) ShuffleFaceDownInRow(rowIndex int) {
	if rowIndex < 0 || rowIndex >= len(b.Rows) {
		return
	}
	coords := b.GetFaceDownSlotsInRow(rowIndex)
	if len(coords) <= 1 {
		return
	}
	cards := make([]*Card, len(coords))
	for i, coord := range coords {
		cards[i] = b.Rows[coord.Row].Slots[coord.Position].Card
	}
	cryptoShuffle(cards)
	for i, coord := range coords {
		b.Rows[coord.Row].Slots[coord.Position].Card = cards[i]
	}
}

// ShuffleAllFaceDown collects all face-down cards on the entire board, shuffles, and places them back.
func (b *Board) ShuffleAllFaceDown() {
	coords := b.GetFaceDownSlots()
	if len(coords) <= 1 {
		return
	}
	cards := make([]*Card, len(coords))
	for i, coord := range coords {
		cards[i] = b.Rows[coord.Row].Slots[coord.Position].Card
	}
	cryptoShuffle(cards)
	for i, coord := range coords {
		b.Rows[coord.Row].Slots[coord.Position].Card = cards[i]
	}
}

// SetAllFaceDown sets all face-up cards to face-down.
func (b *Board) SetAllFaceDown() {
	for r := range b.Rows {
		for p := range b.Rows[r].Slots {
			if b.Rows[r].Slots[p].Card != nil {
				b.Rows[r].Slots[p].FaceUp = false
			}
		}
	}
}

// SwapCards swaps the cards (not face state) at two coordinates.
func (b *Board) SwapCards(a, c2 SlotCoord) {
	slotA := b.GetSlot(a)
	slotB := b.GetSlot(c2)
	if slotA == nil || slotB == nil {
		return
	}
	slotA.Card, slotB.Card = slotB.Card, slotA.Card
}

// GetNonEmptyRowIndices returns indices of rows that have at least one card.
func (b *Board) GetNonEmptyRowIndices() []int {
	var indices []int
	for r, row := range b.Rows {
		for _, slot := range row.Slots {
			if slot.Card != nil {
				indices = append(indices, r)
				break
			}
		}
	}
	return indices
}

func cryptoShuffle[T any](slice []T) {
	for i := len(slice) - 1; i > 0; i-- {
		jBig, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(jBig.Int64())
		slice[i], slice[j] = slice[j], slice[i]
	}
}
