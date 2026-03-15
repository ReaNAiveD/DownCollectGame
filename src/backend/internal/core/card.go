package core

// Suit represents a card's suit.
type Suit int

const (
	SuitNone     Suit = iota // Joker
	SuitHearts               // ♥
	SuitDiamonds             // ♦
	SuitSpades               // ♠
	SuitClubs                // ♣
)

func (s Suit) String() string {
	switch s {
	case SuitHearts:
		return "Hearts"
	case SuitDiamonds:
		return "Diamonds"
	case SuitSpades:
		return "Spades"
	case SuitClubs:
		return "Clubs"
	default:
		return "None"
	}
}

// ColorGroup represents the color grouping of suits.
type ColorGroup int

const (
	ColorNone  ColorGroup = iota
	ColorRed              // Hearts, Diamonds
	ColorBlack            // Spades, Clubs
)

func SuitColorGroup(s Suit) ColorGroup {
	switch s {
	case SuitHearts, SuitDiamonds:
		return ColorRed
	case SuitSpades, SuitClubs:
		return ColorBlack
	default:
		return ColorNone
	}
}

// Rank represents a card's rank.
type Rank int

const (
	RankSmallJoker Rank = iota
	RankBigJoker
	RankA
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
	Rank9
	Rank10
	RankJ
	RankQ
	RankK
)

func (r Rank) String() string {
	names := map[Rank]string{
		RankSmallJoker: "SmallJoker",
		RankBigJoker:   "BigJoker",
		RankA:          "A",
		Rank2:          "2",
		Rank3:          "3",
		Rank4:          "4",
		Rank5:          "5",
		Rank6:          "6",
		Rank7:          "7",
		Rank8:          "8",
		Rank9:          "9",
		Rank10:         "10",
		RankJ:          "J",
		RankQ:          "Q",
		RankK:          "K",
	}
	if name, ok := names[r]; ok {
		return name
	}
	return "Unknown"
}

// BaseScore returns the base score for a rank.
func (r Rank) BaseScore() int {
	switch r {
	case RankSmallJoker, RankBigJoker:
		return 0
	case RankA:
		return 1
	case Rank2:
		return 2
	case Rank3:
		return 3
	case Rank4:
		return 4
	case Rank5:
		return 5
	case Rank6:
		return 6
	case Rank7:
		return 7
	case Rank8:
		return 8
	case Rank9:
		return 9
	case Rank10:
		return 10
	case RankJ:
		return 11
	case RankQ:
		return 12
	case RankK:
		return 13
	default:
		return 0
	}
}

// IsEven returns true if the base score is even.
func (r Rank) IsEven() bool {
	bs := r.BaseScore()
	return bs > 0 && bs%2 == 0
}

// IsHighCard returns true for 10, J, Q, K.
func (r Rank) IsHighCard() bool {
	return r == Rank10 || r == RankJ || r == RankQ || r == RankK
}

// Card represents a single playing card. Cards are immutable value types.
type Card struct {
	ID   int
	Suit Suit
	Rank Rank
}

func (c Card) BaseScore() int {
	return c.Rank.BaseScore()
}

func (c Card) ColorGroup() ColorGroup {
	return SuitColorGroup(c.Suit)
}

// NewStandardDeck creates all 54 cards.
func NewStandardDeck() []Card {
	cards := make([]Card, 0, 54)
	id := 0

	// Small Joker
	cards = append(cards, Card{ID: id, Suit: SuitNone, Rank: RankSmallJoker})
	id++
	// Big Joker
	cards = append(cards, Card{ID: id, Suit: SuitNone, Rank: RankBigJoker})
	id++

	suits := []Suit{SuitHearts, SuitDiamonds, SuitSpades, SuitClubs}
	ranks := []Rank{RankA, Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9, Rank10, RankJ, RankQ, RankK}

	for _, suit := range suits {
		for _, rank := range ranks {
			cards = append(cards, Card{ID: id, Suit: suit, Rank: rank})
			id++
		}
	}
	return cards
}
