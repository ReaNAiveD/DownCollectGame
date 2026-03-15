# DownCollect 实体关系设计

本文档描述游戏系统中各实体及其关系。

---

## 实体关系图 (文本描述)

```
GameConfig ──────────────────────────────────────────────────────┐
                                                                 │
Match (1) ─────────── has ──────────── GameState (1)             │
  │                                      │                       │
  │                                      ├── phase: GamePhase    │
  │                                      ├── round: int          │
  │                                      ├── turn: int           │
  │                                      ├── config ─────────────┘
  │                                      │
  │                                      ├── has ── Deck (1)
  │                                      │           └── cards: []Card
  │                                      │
  │                                      ├── has ── Board (1)
  │                                      │           └── rows: []Row
  │                                      │                 └── slots: []CardSlot
  │                                      │                       ├── card: *Card
  │                                      │                       └── faceUp: bool
  │                                      │
  │                                      ├── has ── RemovedPile (1)
  │                                      │           └── cards: []Card
  │                                      │
  │                                      ├── has ── []PlayerState (2-4)
  │                                      │           ├── playerID: string
  │                                      │           ├── seatIndex: int
  │                                      │           ├── hand: Hand
  │                                      │           │     └── cards: []Card
  │                                      │           ├── character: Character
  │                                      │           ├── score: int
  │                                      │           └── peekInfo: map[SlotCoord]Card
  │                                      │
  │                                      ├── has ── PendingAction (*0..1)
  │                                      │           ├── type: ActionType
  │                                      │           ├── playerID: string
  │                                      │           ├── choices: []ChoiceOption
  │                                      │           └── timeout: Duration
  │                                      │
  │                                      ├── has ── ScoringRuleSet
  │                                      │           ├── globalRules: []ScoringRule
  │                                      │           └── playerRules: map[playerID][]ScoringRule
  │                                      │
  │                                      └── has ── HookRegistry
  │                                                  └── hooks: map[HookPoint][]HookHandler
  │
  └── has ── []Player (2-4)
              ├── id: string
              ├── nickname: string
              ├── isHost: bool
              └── connected: bool
```

---

## 核心实体详述

### Card（卡牌）

卡牌是游戏中不可变的基本单元。一旦创建，其 Suit 和 Rank 不会改变。

```
Card
├── ID: int              // 唯一标识 (0-53)
├── Suit: Suit           // 花色
├── Rank: Rank           // 点数
├── BaseScore() int      // 基础分值（由 Rank 决定）
└── ColorGroup() Color   // 颜色组（Red/Black/None）
```

**关键设计点**：Card 本身不持有状态（如明暗、位置）。状态由持有 Card 的容器（CardSlot、Hand、Deck）管理。

### CardSlot（卡槽）

展示区中的一个位置，持有卡牌及其可见性状态。

```
CardSlot
├── Card: *Card          // 可为 nil（空位）
├── FaceUp: bool         // 明置/暗置
└── Coord: SlotCoord     // 坐标 (Row, Position)
```

### Board（展示区）

由多行组成的公共展示区域。

```
Board
├── Rows: []Row
│
├── GetSlot(row, pos) *CardSlot
├── InsertAt(row, pos, card, faceUp)      // 7/8 的插入操作
├── GetFaceDownSlots() []SlotCoord        // 获取所有暗置卡槽
├── GetFaceDownSlotsInRow(row) []SlotCoord
├── FlipSlot(row, pos)                    // 翻转单个卡槽
├── RemoveCard(row, pos) *Card            // 取走卡牌
└── PlaceCard(row, pos, card, faceUp)     // 放置卡牌
```

### Deck（牌堆）

```
Deck
├── cards: []Card        // 内部卡牌序列
│
├── DrawTop() *Card      // 抽取顶部
├── PushBottom(card)     // 放入底部
├── Remaining() int      // 剩余数量
├── IsEmpty() bool
└── Shuffle()            // 洗混
```

### Hand（手牌）

有序卡牌序列，位置语义重要（最右侧是最新加入的）。

```
Hand
├── cards: []Card
│
├── AddRight(card)       // 加入最右侧
├── RemoveAt(index) Card // 移除指定位置
├── Get(index) Card
├── Size() int
├── LeftOf(index) []Card // 获取左侧所有卡牌（计分用）
└── RightmostIndex() int
```

### PlayerState（玩家游戏状态）

```
PlayerState
├── PlayerID: string
├── SeatIndex: int
├── Hand: Hand
├── Character: Character
├── Score: int
├── ScoreDetails: []ScoreEntry   // 每张卡牌的得分明细
└── PeekInfo: map[SlotCoord]Card // 该玩家通过偷看获得的私有信息
```

### PendingAction（待处理动作）

当游戏流程需要等待玩家输入时产生。

```
PendingAction
├── Type: ActionType              // 动作类型
├── PlayerID: string              // 需要操作的玩家
├── ValidChoices: []ChoiceOption  // 合法选项列表
├── Context: map[string]any       // 上下文信息（如揭示的卡牌）
└── Deadline: time.Time           // 超时时间
```

ActionType 枚举：
- `RevealCard` — 揭示卡牌（点击揭示）
- `SelectSlots` — 选择卡槽（3选两张交换, 5/6选一张替换, 9翻明, 10偷看）
- `SelectPosition` — 选择插入位置（7, 8）
- `SelectRow` — 选择行（J, Q）
- `SelectSlotsPerRow` — 每行各选一张（K）
- `SwapDecision` — 是否交换手牌
- `PickCard` — 选牌阶段选牌
- `ChooseCharacter` — 选择角色

---

## 卡牌效果体系

卡牌效果与卡牌本身分离。每种 Rank 对应一个 CardEffect 实现。

```
CardEffect (interface)
├── GetPlacement() PlacementType          // 卡牌自身去向
├── NeedsPlayerChoice() bool              // 是否需要玩家选择
├── GetRequiredAction(gs, card) *PendingAction  // 生成待处理动作
├── Execute(gs, card, choice) error       // 执行效果
└── Rank() Rank                           // 对应的点数
```

PlacementType 枚举：
- `PlaceOnBoard` — 置入展示区默认位置
- `PlaceOnBoardCustom` — 置入展示区玩家选择的位置（7, 8）
- `ReturnToDeckBottom` — 回到牌堆底（A, 小Joker, 大Joker）
- `EnterHand` — 进入手牌（4）

### 效果分类表

| Rank | Placement | 需要玩家选择 | 选择类型 | 额外效果 |
|------|-----------|-------------|---------|---------|
| 小Joker | ReturnToDeckBottom | 否 | - | 所有明置→暗置 |
| 大Joker | ReturnToDeckBottom | 否 | - | 洗混所有暗置位置 |
| A | ReturnToDeckBottom | 否 | - | 牌堆顶→置入场 |
| 2 | PlaceOnBoard | 否 | - | 洗混当前行暗置牌 |
| 3 | PlaceOnBoard | 是 | SelectSlots(2张暗置) | 交换两张暗置牌位置 |
| 4 | EnterHand | 否(但触发递归揭示) | - | 递归揭示下一张 |
| 5 | PlaceOnBoard | 是 | SelectSlots(1张暗置) | 选中牌→牌堆底，牌堆顶→场 |
| 6 | PlaceOnBoard | 是 | SelectSlots(1张暗置) | 同5 |
| 7 | PlaceOnBoardCustom | 是 | SelectPosition | 插入当前行任意位置 |
| 8 | PlaceOnBoardCustom | 是 | SelectPosition | 同7 |
| 9 | PlaceOnBoard | 是 | SelectSlots(1张暗置) | 选中牌翻明 |
| 10 | PlaceOnBoard | 是 | SelectSlots(1张暗置) | 偷看选中牌 |
| J | PlaceOnBoard(明置) | 是 | SelectRow | 选定行明暗互换 |
| Q | PlaceOnBoard(明置) | 是 | SelectRow | 选定行展示后低分牌洗混 |
| K | (混入场中) | 是 | SelectSlotsPerRow | 每行选一张，连同K洗混放回 |

---

## 计分体系

```
ScoringRule (interface)
├── Name() string
├── Evaluate(context ScoringContext) ScoreResult
└── Priority() int    // 未来扩展用

ScoringContext
├── CurrentCard: Card
├── VisibleCards: []Card          // 当前卡牌 + 左侧所有
├── LeftNeighbor: *Card           // 紧邻左侧
├── LeftCount: int                // 左侧卡牌数
├── AllHandCards: []Card          // 完整手牌（角色/全局规则用）
└── PlayerState: *PlayerState

ScoreResult
├── Score: int
├── SkipRemaining: bool           // 是否跳过左侧所有卡牌
├── OverrideCards: []int          // 被接管计分的卡牌索引（J/Q/K）
└── Description: string          // 计分说明
```

每张卡牌的 ScoringRule 实现：全部按规则文档实现。
GlobalScoringRule / PrivateScoringRule / CharacterScoringRule：当前均为 Noop（返回 0 分，不跳过）。

---

## Hook 体系

```
HookPoint (enum)
// 见术语文档中的完整 Hook 点列表

HookHandler (interface)
├── OnTrigger(gs *GameState, ctx HookContext) HookResult

HookResult
├── ModifiedState: bool          // 是否修改了游戏状态
├── Cancel: bool                 // 是否取消后续操作（预留）
└── Message: string              // 提示信息（预留）

HookRegistry
├── Register(point HookPoint, handler HookHandler)
├── Trigger(point HookPoint, gs *GameState, ctx HookContext) []HookResult
└── Clear()
```

当前所有 HookHandler 为 Noop（不修改状态，不取消）。角色选择后注册角色对应的 Hook。

---

## 游戏状态机

```
GamePhase (enum)
├── PhaseWaitingPlayers    // 等待玩家加入
├── PhaseCharacterSelect   // 角色选择
├── PhaseSetup             // 初始化发牌
├── PhaseReveal            // 揭示阶段
├── PhasePick              // 选牌阶段
├── PhaseScoring           // 计分阶段
├── PhaseFinished          // 游戏结束

TurnStep (enum)  // 揭示阶段回合内的子步骤
├── StepRevealCard         // 等待揭示
├── StepEffectAction       // 等待效果相关的玩家选择
├── StepEffectResolve      // 效果执行中（含递归揭示）
├── StepSwapDecision       // 等待交换决策
├── StepTurnEnd            // 回合结束处理
```

---

## 网络消息设计

### OpCode 定义

```
// 客户端 → 服务器
OpJoinRoom          = 1    // 加入房间
OpStartGame         = 2    // 开始游戏
OpChooseCharacter   = 3    // 选择角色
OpRevealCard        = 4    // 揭示卡牌
OpPlayerChoice      = 5    // 玩家选择（统一入口）
OpSwapDecision      = 6    // 交换决策
OpPickCard          = 7    // 选牌

// 服务器 → 客户端
OpGameState         = 101  // 游戏状态更新（PlayerView）
OpActionRequired    = 102  // 请求玩家操作
OpCardRevealed      = 103  // 卡牌揭示通知
OpEffectResolved    = 104  // 效果结果通知
OpPhaseChanged      = 105  // 阶段变更通知
OpGameResult        = 106  // 游戏结果
OpError             = 107  // 错误消息
OpRoomUpdate        = 108  // 房间状态更新（玩家列表等）
```

### PlayerView 生成

在服务器的 WebSocket 层中，从权威 GameState 为每个玩家生成 PlayerView：

```
PlayerView
├── Phase: GamePhase
├── Round: int
├── Turn: int
├── ActivePlayerSeat: int
├── MyHand: []CardInfo            // 自己的手牌（可见）
├── Players: []PlayerInfo         // 所有玩家公开信息
│     ├── Nickname, SeatIndex, HandSize, Connected
│     └── Character: string       // 角色名（公开）
├── Board: [][]SlotView           // 展示区视图
│     ├── HasCard: bool
│     ├── FaceUp: bool
│     ├── Card: *CardInfo         // 明置时有值，暗置时为 nil
│     └── PeekedCard: *CardInfo   // 该玩家偷看过时有值
├── DeckRemaining: int
├── PendingAction: *ActionView    // 当前需要的操作（如果是自己）
└── Scores: map[seat]ScoreView    // 计分阶段可见
```
