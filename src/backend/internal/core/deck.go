package core

import (
	"crypto/rand"
	"math/big"
)

// Deck represents the draw pile.
type Deck struct {
	cards []Card
}

func NewDeck(cards []Card) *Deck {
	c := make([]Card, len(cards))
	copy(c, cards)
	return &Deck{cards: c}
}

func (d *Deck) Shuffle() {
	for i := len(d.cards) - 1; i > 0; i-- {
		jBig, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(jBig.Int64())
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	}
}

func (d *Deck) DrawTop() (Card, bool) {
	if len(d.cards) == 0 {
		return Card{}, false
	}
	card := d.cards[0]
	d.cards = d.cards[1:]
	return card, true
}

func (d *Deck) PushBottom(card Card) {
	d.cards = append(d.cards, card)
}

func (d *Deck) Remaining() int {
	return len(d.cards)
}

func (d *Deck) IsEmpty() bool {
	return len(d.cards) == 0
}

// RemoveRandom removes n random cards and returns them.
func (d *Deck) RemoveRandom(n int) []Card {
	if n >= len(d.cards) {
		removed := d.cards
		d.cards = nil
		return removed
	}
	d.Shuffle()
	removed := make([]Card, n)
	copy(removed, d.cards[:n])
	d.cards = d.cards[n:]
	return removed
}

// DealOne removes the top card from the deck.
func (d *Deck) DealOne() (Card, bool) {
	return d.DrawTop()
}
