package core

// Character is the interface for character abilities.
// Implementations register hooks during game setup.
type Character interface {
	Name() string
	// RegisterHooks registers this character's abilities with the hook registry.
	RegisterHooks(registry *HookRegistry, playerID string)
}

// NoopCharacter is a placeholder character with no abilities.
type NoopCharacter struct{}

func (c *NoopCharacter) Name() string                                          { return "None" }
func (c *NoopCharacter) RegisterHooks(registry *HookRegistry, playerID string) {}

// HookPoint defines when a hook can trigger.
type HookPoint int

const (
	// Reveal phase hooks
	OnRevealPhaseStart HookPoint = iota
	OnRevealRoundStart
	OnRevealTurnStart
	BeforeRevealCard
	AfterRevealCard
	AfterEffectResolved
	BeforeSwap
	AfterSwap
	OnRevealTurnEnd
	OnRevealRoundEnd
	OnRevealPhaseEnd

	// Pick phase hooks
	OnPickPhaseStart
	OnPickRoundStart
	OnPickTurnStart
	BeforePickCard
	AfterPickCard
	OnPickTurnEnd
	OnPickRoundEnd
	OnPickPhaseEnd

	// Scoring phase hooks
	OnScoringPhaseStart
	BeforePlayerScoring
	BeforePrivateScoring
	AfterPrivateScoring
	BeforeGlobalScoring
	AfterGlobalScoring
	BeforeCardScoring
	AfterCardScoring
	AfterPlayerScoring
	OnScoringPhaseEnd
)

// HookContext carries contextual information for hook triggers.
type HookContext struct {
	PlayerID  string
	Card      *Card
	SlotCoord *SlotCoord
	ExtraData map[string]interface{}
}

// HookResult contains the result of a hook execution.
type HookResult struct {
	Message string
}

// HookHandler is the interface for hook implementations.
type HookHandler interface {
	OnTrigger(gs *GameState, ctx HookContext) HookResult
}

// HookRegistry manages all registered hooks.
type HookRegistry struct {
	hooks map[HookPoint][]HookHandler
}

func NewHookRegistry() *HookRegistry {
	return &HookRegistry{hooks: make(map[HookPoint][]HookHandler)}
}

func (hr *HookRegistry) Register(point HookPoint, handler HookHandler) {
	hr.hooks[point] = append(hr.hooks[point], handler)
}

func (hr *HookRegistry) Trigger(point HookPoint, gs *GameState, ctx HookContext) []HookResult {
	handlers, ok := hr.hooks[point]
	if !ok {
		return nil
	}
	results := make([]HookResult, 0, len(handlers))
	for _, h := range handlers {
		results = append(results, h.OnTrigger(gs, ctx))
	}
	return results
}

func (hr *HookRegistry) Clear() {
	hr.hooks = make(map[HookPoint][]HookHandler)
}
