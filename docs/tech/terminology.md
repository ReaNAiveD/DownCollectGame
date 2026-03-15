# DownCollect 术语文档

本文档定义 DownCollect 游戏系统中所有核心术语，作为开发过程中的统一语言。

---

## 卡牌相关

| 术语 | 英文 | 说明 |
|------|------|------|
| 卡牌 | Card | 游戏中的基本单元。每张卡牌有花色（Suit）、点数（Rank）、基础分值（BaseScore）。 |
| 花色 | Suit | 卡牌的花色类型：Hearts(♥)、Spades(♠)、Diamonds(♦)、Clubs(♣)、None（Joker无花色）。 |
| 点数 | Rank | 卡牌的点数：A, 2-10, J, Q, K, SmallJoker, BigJoker。 |
| 基础分值 | BaseScore | 卡牌的固有分值。A=1, 2-10=点数, J=11, Q=12, K=13, Joker=0。 |
| 颜色组 | ColorGroup | 红色系（Hearts, Diamonds）和黑色系（Spades, Clubs）。 |
| 明置 | FaceUp | 卡牌正面朝上，所有玩家可见。 |
| 暗置 | FaceDown | 卡牌正面朝下，所有玩家不可见（偷看信息除外）。 |

---

## 区域相关

| 术语 | 英文 | 说明 |
|------|------|------|
| 牌堆 | Deck | 公共牌堆，暗置叠放，仅能从顶部取牌或放到底部。 |
| 卡牌展示区 | Board | 公共区域，由多行（Row）组成，每行对应一个揭示轮次。卡牌以明置或暗置状态放置。 |
| 行 | Row | 展示区中的一行，对应揭示阶段的一个轮次。行号从 1 开始。行内卡牌按位置（Position）排列。 |
| 位置 | Position | 行内卡牌的索引，从 0 开始，从左到右递增。 |
| 卡槽 | CardSlot | 展示区中一个具体位置，包含卡牌引用和明暗状态。坐标为 (Row, Position)。 |
| 手牌 | Hand | 玩家私有的卡牌序列，有序排列，仅本人可见。最右侧为最新加入的卡牌。 |
| 移除牌堆 | RemovedPile | 游戏开始时随机移除的卡牌，不参与本局游戏，所有人不可见。 |

---

## 玩家相关

| 术语 | 英文 | 说明 |
|------|------|------|
| 玩家 | Player | 游戏参与者，拥有唯一 ID、昵称、手牌、角色。 |
| 房主 | Host | 创建房间的玩家，拥有开始游戏的权限。 |
| 当前行动玩家 | ActivePlayer | 当前轮到行动的玩家。 |
| 起始玩家 | StartingPlayer | 每轮第一个行动的玩家。 |
| 玩家座位 | SeatIndex | 玩家在房间中的固定序号（0-based），决定行动顺序。 |

---

## 角色相关

| 术语 | 英文 | 说明 |
|------|------|------|
| 角色 | Character | 玩家选择的游戏角色，提供特殊能力。当前为 Noop 实现。 |
| 角色能力 | CharacterAbility | 角色在特定时间点可使用的能力。通过 Hook 机制触发。 |
| 角色池 | CharacterPool | 所有可用角色的集合。 |

---

## 游戏流程相关

| 术语 | 英文 | 说明 |
|------|------|------|
| 游戏阶段 | GamePhase | 游戏的主要状态：Setup, Reveal, Pick, Scoring, Finished。 |
| 轮次 | Round | 阶段内的一个完整循环，所有玩家各行动一次为一个轮次。 |
| 回合 | Turn | 单个玩家的一次完整行动机会。 |
| 揭示 | Reveal | 翻开牌堆顶牌并触发其效果的行为。 |
| 选牌 | Pick | 从展示区选择一张卡牌加入手牌的行为。 |
| 交换 | Swap | 在揭示阶段，玩家用手牌与展示区当前行最右侧卡牌互换。 |
| 偷看 | Peek | 仅当前玩家可看到某张暗置卡牌的内容，不改变明暗状态。 |

---

## 卡牌效果相关

| 术语 | 英文 | 说明 |
|------|------|------|
| 卡牌效果 | CardEffect | 卡牌在揭示阶段被揭示时触发的效果。 |
| 放置策略 | PlacementType | 卡牌效果执行后卡牌自身的去向：EnterBoard（置入展示区）、EnterHand（进入手牌）、ReturnToDeck（回到牌堆底）。 |
| 效果动作 | EffectAction | 卡牌效果的具体操作，如洗混、翻转、替换、偷看等。 |
| 玩家选择 | PlayerChoice | 效果执行过程中需要玩家做出的决策。 |
| 选择类型 | ChoiceType | 玩家选择的分类：SelectSlot（选卡槽）、SelectRow（选行）、SelectPosition（选插入位置）、SelectHandCard（选手牌）、Confirm（确认）。 |

---

## 计分相关

| 术语 | 英文 | 说明 |
|------|------|------|
| 特殊计分规则 | ScoringRule | 每张卡牌附带的特殊计分逻辑，在计分阶段从右到左依次执行。 |
| 公共计分规则 | GlobalScoringRule | 适用于所有玩家的额外计分规则。当前为 Noop。 |
| 私有计分规则 | PrivateScoringRule | 仅适用于特定玩家的计分规则。当前为 Noop。 |
| 角色计分规则 | CharacterScoringRule | 由角色提供的计分规则。当前为 Noop。 |
| 可见卡牌 | VisibleCards | 计分时，当前卡牌自身 + 其左侧所有卡牌。 |
| 跳过 | Skip | J/Q/K 触发牌型后，左侧所有卡牌不再单独计分。 |

---

## 网络与房间相关

| 术语 | 英文 | 说明 |
|------|------|------|
| 房间 | Room / Match | 一局游戏的容器，通过房间号标识。 |
| 房间号 | RoomCode | 唯一标识房间的短字符串（如 "A3X9"）。 |
| 透明认证 | DeviceAuth | 基于设备 ID 的自动匿名认证，用户无感知。 |
| 操作码 | OpCode | WebSocket 消息的类型标识，用于区分不同的消息用途。 |
| 游戏视图 | PlayerView | 从权威游戏状态中为特定玩家生成的可见信息子集。处理信息不对称。 |

---

## 游戏配置

| 术语 | 英文 | 说明 |
|------|------|------|
| 游戏配置 | GameConfig | 可调参数集合，包括：移除牌数、初始手牌数、揭示轮数、选牌轮数、玩家人数范围、超时时间等。 |

---

## Hook 时间点

角色能力通过 Hook 系统在特定时间点触发。完整的 Hook 点列表：

### 揭示阶段 Hook

| Hook | 英文标识 | 时间点 |
|------|---------|--------|
| 揭示阶段开始 | OnRevealPhaseStart | 揭示阶段开始时 |
| 揭示轮次开始 | OnRevealRoundStart | 每个揭示轮次开始时 |
| 玩家行动开始 | OnRevealTurnStart | 玩家的回合开始时 |
| 揭示卡牌前 | BeforeRevealCard | 玩家翻牌前 |
| 揭示卡牌后 | AfterRevealCard | 翻牌后、效果处理前 |
| 效果处理后 | AfterEffectResolved | 卡牌效果处理并放置后 |
| 交换前 | BeforeSwap | 玩家选择交换前 |
| 交换后 | AfterSwap | 玩家选择交换后 |
| 玩家行动结束 | OnRevealTurnEnd | 玩家的回合结束前 |
| 揭示轮次结束 | OnRevealRoundEnd | 每个揭示轮次结束时 |
| 揭示阶段结束 | OnRevealPhaseEnd | 揭示阶段结束时 |

### 选牌阶段 Hook

| Hook | 英文标识 | 时间点 |
|------|---------|--------|
| 选牌阶段开始 | OnPickPhaseStart | 选牌阶段开始时 |
| 选牌轮次开始 | OnPickRoundStart | 每个选牌轮次开始时 |
| 玩家行动开始 | OnPickTurnStart | 玩家的回合开始时 |
| 选择卡牌前 | BeforePickCard | 玩家选牌前 |
| 选择卡牌后 | AfterPickCard | 玩家选牌后 |
| 玩家行动结束 | OnPickTurnEnd | 玩家的回合结束前 |
| 选牌轮次结束 | OnPickRoundEnd | 每个选牌轮次结束时 |
| 选牌阶段结束 | OnPickPhaseEnd | 选牌阶段结束时 |

### 计分阶段 Hook

| Hook | 英文标识 | 时间点 |
|------|---------|--------|
| 计分阶段开始 | OnScoringPhaseStart | 计分阶段开始时 |
| 玩家计分前 | BeforePlayerScoring | 单个玩家开始计分前 |
| 私有规则计分前 | BeforePrivateScoring | 私有计分规则执行前 |
| 私有规则计分后 | AfterPrivateScoring | 私有计分规则执行后 |
| 公共规则计分前 | BeforeGlobalScoring | 公共计分规则执行前 |
| 公共规则计分后 | AfterGlobalScoring | 公共计分规则执行后 |
| 卡牌计分前 | BeforeCardScoring | 单张卡牌计分前 |
| 卡牌计分后 | AfterCardScoring | 单张卡牌计分后 |
| 玩家计分后 | AfterPlayerScoring | 单个玩家计分完成后 |
| 计分阶段结束 | OnScoringPhaseEnd | 计分阶段结束时 |
