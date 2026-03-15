package core

// ScoringRule evaluates a card's special scoring.
type ScoringRule interface {
	Name() string
	Evaluate(ctx ScoringContext) ScoreResult
}

// ScoringContext provides information for scoring evaluation.
type ScoringContext struct {
	CurrentCard  Card
	CardIndex    int    // index in hand (0 = leftmost)
	VisibleCards []Card // current card + all cards to its left
	LeftNeighbor *Card  // immediate left neighbor, nil if none
	LeftCount    int    // number of cards to the left
	AllHandCards []Card // full hand (for global/character rules)
}

// ScoreResult is the output of a scoring rule.
type ScoreResult struct {
	Score         int
	SkipRemaining bool  // if true, all cards to the left are skipped
	OverrideCards []int // card indices whose scores are overridden (absorbed by J/Q/K)
	Description   string
}

// NoopScoringRule always returns 0, no skip.
type NoopScoringRule struct {
	RuleName string
}

func (n *NoopScoringRule) Name() string { return n.RuleName }
func (n *NoopScoringRule) Evaluate(ctx ScoringContext) ScoreResult {
	return ScoreResult{Score: 0, Description: "noop"}
}

// cardScoringRules maps Rank -> ScoringRule
var cardScoringRules = map[Rank]ScoringRule{}

func RegisterScoringRule(rank Rank, rule ScoringRule) {
	cardScoringRules[rank] = rule
}

func GetScoringRule(rank Rank) ScoringRule {
	if rule, ok := cardScoringRules[rank]; ok {
		return rule
	}
	return nil
}

// defaultScoringRules returns a fresh map of all built-in scoring rules.
func defaultScoringRules() map[Rank]ScoringRule {
	return map[Rank]ScoringRule{
		RankSmallJoker: &SmallJokerScoring{},
		RankBigJoker:   &BigJokerScoring{},
		RankA:          &AceScoring{},
		Rank2:          &TwoScoring{},
		Rank3:          &ThreeScoring{},
		Rank4:          &FourScoring{},
		Rank5:          &FiveScoring{},
		Rank6:          &SixScoring{},
		Rank7:          &SevenScoring{},
		Rank8:          &EightScoring{},
		Rank9:          &NineScoring{},
		Rank10:         &TenScoring{},
		RankJ:          &JackScoring{},
		RankQ:          &QueenScoring{},
		RankK:          &KingScoring{},
	}
}

func init() {
	for rank, rule := range defaultScoringRules() {
		RegisterScoringRule(rank, rule)
	}
}

// --- Small Joker: 0 points, left neighbor loses special scoring ---

type SmallJokerScoring struct{}

func (s *SmallJokerScoring) Name() string { return "迷雾" }
func (s *SmallJokerScoring) Evaluate(ctx ScoringContext) ScoreResult {
	return ScoreResult{Score: 0, Description: "小Joker计0分，左邻失去特殊计分"}
}

// --- Big Joker: score = -(left count) ---

type BigJokerScoring struct{}

func (s *BigJokerScoring) Name() string { return "混沌馈赠" }
func (s *BigJokerScoring) Evaluate(ctx ScoringContext) ScoreResult {
	score := -ctx.LeftCount
	return ScoreResult{Score: score, Description: "混沌馈赠：-左侧卡牌数"}
}

// --- Ace: -3 if leftmost, else 1 ---

type AceScoring struct{}

func (s *AceScoring) Name() string { return "先驱" }
func (s *AceScoring) Evaluate(ctx ScoringContext) ScoreResult {
	if ctx.LeftCount == 0 {
		return ScoreResult{Score: -3, Description: "先驱：最左端，计-3分"}
	}
	return ScoreResult{Score: 1, Description: "先驱：非最左端，计1分"}
}

// --- 2: 0 if left neighbor is even, else 2 ---

type TwoScoring struct{}

func (s *TwoScoring) Name() string { return "共鸣" }
func (s *TwoScoring) Evaluate(ctx ScoringContext) ScoreResult {
	if ctx.LeftNeighbor != nil && ctx.LeftNeighbor.Rank.IsEven() {
		return ScoreResult{Score: 0, Description: "共鸣：左邻为偶数牌，计0分"}
	}
	return ScoreResult{Score: 2, Description: "共鸣：左邻非偶数牌，计2分"}
}

// --- 3: 3 - (number of distinct suits in visible cards), min 0 ---

type ThreeScoring struct{}

func (s *ThreeScoring) Name() string { return "多彩" }
func (s *ThreeScoring) Evaluate(ctx ScoringContext) ScoreResult {
	suits := make(map[Suit]bool)
	for _, c := range ctx.VisibleCards {
		if c.Suit != SuitNone {
			suits[c.Suit] = true
		}
	}
	score := 3 - len(suits)
	if score < 0 {
		score = 0
	}
	return ScoreResult{Score: score, Description: "多彩：花色越多分越低"}
}

// --- 4: max(0, 4 - 2*leftCount) ---

type FourScoring struct{}

func (s *FourScoring) Name() string { return "纵深" }
func (s *FourScoring) Evaluate(ctx ScoringContext) ScoreResult {
	score := 4 - 2*ctx.LeftCount
	if score < 0 {
		score = 0
	}
	return ScoreResult{Score: score, Description: "纵深：左侧越多分越低"}
}

// --- 5: 2 if visible >= 3, else 5 ---

type FiveScoring struct{}

func (s *FiveScoring) Name() string { return "聚众" }
func (s *FiveScoring) Evaluate(ctx ScoringContext) ScoreResult {
	if len(ctx.VisibleCards) >= 3 {
		return ScoreResult{Score: 2, Description: "聚众：可见≥3张，计2分"}
	}
	return ScoreResult{Score: 5, Description: "聚众：可见<3张，计5分"}
}

// --- 6: 3 if visible count is even, else 6 ---

type SixScoring struct{}

func (s *SixScoring) Name() string { return "均衡" }
func (s *SixScoring) Evaluate(ctx ScoringContext) ScoreResult {
	if len(ctx.VisibleCards)%2 == 0 {
		return ScoreResult{Score: 3, Description: "均衡：可见数为偶数，计3分"}
	}
	return ScoreResult{Score: 6, Description: "均衡：可见数为奇数，计6分"}
}

// --- 7: 3 if left neighbor base score within [5,9], else 7 ---

type SevenScoring struct{}

func (s *SevenScoring) Name() string { return "亲和" }
func (s *SevenScoring) Evaluate(ctx ScoringContext) ScoreResult {
	if ctx.LeftNeighbor != nil {
		bs := ctx.LeftNeighbor.BaseScore()
		if bs >= 5 && bs <= 9 {
			return ScoreResult{Score: 3, Description: "亲和：左邻基础分值在5-9，计3分"}
		}
	}
	return ScoreResult{Score: 7, Description: "亲和：不满足条件，计7分"}
}

// --- 8: 8 - 2*(same color group cards visible, excluding self), min 2 ---

type EightScoring struct{}

func (s *EightScoring) Name() string { return "褪色" }
func (s *EightScoring) Evaluate(ctx ScoringContext) ScoreResult {
	myColor := ctx.CurrentCard.ColorGroup()
	sameCount := 0
	for _, c := range ctx.VisibleCards {
		if c.ID != ctx.CurrentCard.ID && SuitColorGroup(c.Suit) == myColor && myColor != ColorNone {
			sameCount++
		}
	}
	score := 8 - 2*sameCount
	if score < 2 {
		score = 2
	}
	return ScoreResult{Score: score, Description: "褪色：同色越多分越低"}
}

// --- 9: 4 if left neighbor is 10/J/Q/K, else 9 ---

type NineScoring struct{}

func (s *NineScoring) Name() string { return "庇护" }
func (s *NineScoring) Evaluate(ctx ScoringContext) ScoreResult {
	if ctx.LeftNeighbor != nil && ctx.LeftNeighbor.Rank.IsHighCard() {
		return ScoreResult{Score: 4, Description: "庇护：左邻为大牌，计4分"}
	}
	return ScoreResult{Score: 9, Description: "庇护：左邻非大牌，计9分"}
}

// --- 10: 5 if any pair in visible, else 10 ---

type TenScoring struct{}

func (s *TenScoring) Name() string { return "对子" }
func (s *TenScoring) Evaluate(ctx ScoringContext) ScoreResult {
	seen := make(map[Rank]bool)
	for _, c := range ctx.VisibleCards {
		if seen[c.Rank] {
			return ScoreResult{Score: 5, Description: "对子：存在相同点数，计5分"}
		}
		seen[c.Rank] = true
	}
	return ScoreResult{Score: 10, Description: "对子：无相同点数，计10分"}
}

// --- J: Flush - if all left cards same suit as J, absorb left ---

type JackScoring struct{}

func (s *JackScoring) Name() string { return "同花" }
func (s *JackScoring) Evaluate(ctx ScoringContext) ScoreResult {
	if ctx.LeftCount == 0 {
		return ScoreResult{Score: 11, Description: "同花：左侧无牌，计11分"}
	}

	leftCards := ctx.VisibleCards[:ctx.LeftCount]
	mySuit := ctx.CurrentCard.Suit
	allSame := true
	sumBase := 0
	for _, c := range leftCards {
		if c.Suit != mySuit {
			allSame = false
			break
		}
		sumBase += c.BaseScore()
	}

	if !allSame {
		return ScoreResult{Score: 11, Description: "同花：花色不一致，计11分"}
	}

	score := 11 + sumBase/4
	overrides := make([]int, ctx.LeftCount)
	for i := range overrides {
		overrides[i] = i
	}
	return ScoreResult{
		Score:         score,
		SkipRemaining: true,
		OverrideCards: overrides,
		Description:   "同花：统率同花色，吸收左侧卡牌",
	}
}

// --- Q: Straight - if left cards form consecutive sequence ---

type QueenScoring struct{}

func (s *QueenScoring) Name() string { return "顺子" }
func (s *QueenScoring) Evaluate(ctx ScoringContext) ScoreResult {
	if ctx.LeftCount < 2 {
		return ScoreResult{Score: 12, Description: "顺子：左侧不足2张，计12分"}
	}

	leftCards := ctx.VisibleCards[:ctx.LeftCount]
	scores := make([]int, len(leftCards))
	sumBase := 0
	for i, c := range leftCards {
		scores[i] = c.BaseScore()
		sumBase += scores[i]
	}

	// Check if scores form a consecutive sequence (order doesn't matter)
	if !isConsecutive(scores) {
		return ScoreResult{Score: 12, Description: "顺子：未形成顺子，计12分"}
	}

	score := 12 + sumBase/3
	overrides := make([]int, ctx.LeftCount)
	for i := range overrides {
		overrides[i] = i
	}
	return ScoreResult{
		Score:         score,
		SkipRemaining: true,
		OverrideCards: overrides,
		Description:   "顺子：串联连续序列，吸收左侧卡牌",
	}
}

func isConsecutive(nums []int) bool {
	if len(nums) == 0 {
		return false
	}
	minVal := nums[0]
	maxVal := nums[0]
	seen := make(map[int]bool)
	for _, n := range nums {
		if seen[n] {
			return false // duplicates
		}
		seen[n] = true
		if n < minVal {
			minVal = n
		}
		if n > maxVal {
			maxVal = n
		}
	}
	return maxVal-minVal == len(nums)-1
}

// --- K: Rainbow - if left has ≥3 cards with ≥3 suits, absorb left ---

type KingScoring struct{}

func (s *KingScoring) Name() string { return "彩虹" }
func (s *KingScoring) Evaluate(ctx ScoringContext) ScoreResult {
	if ctx.LeftCount < 3 {
		return ScoreResult{Score: 13, Description: "彩虹：左侧不足3张，计13分"}
	}

	leftCards := ctx.VisibleCards[:ctx.LeftCount]
	suits := make(map[Suit]bool)
	sumBase := 0
	for _, c := range leftCards {
		if c.Suit != SuitNone {
			suits[c.Suit] = true
		}
		sumBase += c.BaseScore()
	}

	if len(suits) < 3 {
		return ScoreResult{Score: 13, Description: "彩虹：花色不足3种，计13分"}
	}

	score := 13 + sumBase/2
	overrides := make([]int, ctx.LeftCount)
	for i := range overrides {
		overrides[i] = i
	}
	return ScoreResult{
		Score:         score,
		SkipRemaining: true,
		OverrideCards: overrides,
		Description:   "彩虹：召唤彩虹，吸收左侧卡牌",
	}
}

// CalculatePlayerScore runs the full scoring pipeline for a player.
func CalculatePlayerScore(player *PlayerState, mode ScoringMode, rules map[Rank]ScoringRule) {
	hand := player.Hand
	if hand.Size() == 0 {
		player.Score = 0
		return
	}

	// Base-only mode: just sum base scores
	if mode == ScoringModeBaseOnly {
		totalScore := 0
		details := make([]ScoreEntry, hand.Size())
		for i, card := range hand.Cards {
			score := card.BaseScore()
			totalScore += score
			details[i] = ScoreEntry{
				CardIndex:   i,
				Card:        card,
				RuleName:    "基础分值",
				Score:       score,
				Description: "仅计基础分值",
			}
		}
		player.Score = totalScore
		player.ScoreDetails = details
		return
	}

	// Special scoring mode
	totalScore := 0
	details := make([]ScoreEntry, hand.Size())
	skippedSet := make(map[int]bool)

	// Identify all Small Joker fog targets (each fogs its left neighbor)
	foggedSet := make(map[int]bool)
	for i := 0; i < hand.Size(); i++ {
		if hand.Cards[i].Rank == RankSmallJoker && i > 0 {
			foggedSet[i-1] = true
		}
	}

	// Score from right to left
	for i := hand.Size() - 1; i >= 0; i-- {
		card := hand.Cards[i]
		entry := ScoreEntry{
			CardIndex: i,
			Card:      card,
		}

		if skippedSet[i] {
			entry.Skipped = true
			entry.Score = 0
			entry.Description = "被跳过"
			details[i] = entry
			continue
		}

		// Check if this card is fogged by Small Joker
		if foggedSet[i] {
			entry.Score = card.BaseScore()
			entry.RuleName = "被迷雾笼罩"
			entry.Description = "特殊计分被封锁，仅计基础分值"
			totalScore += entry.Score
			details[i] = entry
			continue
		}

		rule := rules[card.Rank]
		if rule == nil {
			entry.Score = card.BaseScore()
			entry.RuleName = "基础"
			entry.Description = "无特殊规则"
			totalScore += entry.Score
			details[i] = entry
			continue
		}

		visible := hand.VisibleFrom(i)
		var leftNeighbor *Card
		if i > 0 {
			c := hand.Cards[i-1]
			leftNeighbor = &c
		}

		ctx := ScoringContext{
			CurrentCard:  card,
			CardIndex:    i,
			VisibleCards: visible,
			LeftNeighbor: leftNeighbor,
			LeftCount:    i,
			AllHandCards: hand.Cards,
		}

		result := rule.Evaluate(ctx)
		entry.Score = result.Score
		entry.RuleName = rule.Name()
		entry.Description = result.Description
		totalScore += entry.Score
		details[i] = entry

		if result.SkipRemaining {
			for _, idx := range result.OverrideCards {
				skippedSet[idx] = true
			}
		}
	}

	player.Score = totalScore
	player.ScoreDetails = details
}
