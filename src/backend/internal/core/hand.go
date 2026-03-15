package core

// Hand represents a player's ordered hand of cards.
type Hand struct {
	Cards []Card
}

func NewHand() *Hand {
	return &Hand{Cards: make([]Card, 0)}
}

func (h *Hand) AddRight(card Card) {
	h.Cards = append(h.Cards, card)
}

func (h *Hand) RemoveAt(index int) (Card, bool) {
	if index < 0 || index >= len(h.Cards) {
		return Card{}, false
	}
	card := h.Cards[index]
	h.Cards = append(h.Cards[:index], h.Cards[index+1:]...)
	return card, true
}

func (h *Hand) Get(index int) (Card, bool) {
	if index < 0 || index >= len(h.Cards) {
		return Card{}, false
	}
	return h.Cards[index], true
}

func (h *Hand) Size() int {
	return len(h.Cards)
}

// LeftOf returns all cards to the left of (not including) the given index.
func (h *Hand) LeftOf(index int) []Card {
	if index <= 0 || index > len(h.Cards) {
		return nil
	}
	result := make([]Card, index)
	copy(result, h.Cards[:index])
	return result
}

// VisibleFrom returns cards visible from position index (self + all left).
func (h *Hand) VisibleFrom(index int) []Card {
	if index < 0 || index >= len(h.Cards) {
		return nil
	}
	result := make([]Card, index+1)
	copy(result, h.Cards[:index+1])
	return result
}

func (h *Hand) RightmostIndex() int {
	return len(h.Cards) - 1
}

// ReplaceAt replaces the card at index and returns the old card.
func (h *Hand) ReplaceAt(index int, card Card) (Card, bool) {
	if index < 0 || index >= len(h.Cards) {
		return Card{}, false
	}
	old := h.Cards[index]
	h.Cards[index] = card
	return old, true
}
