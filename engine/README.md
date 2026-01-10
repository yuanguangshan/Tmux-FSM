
# engine 模块

## 模块定位（现实版本）

`engine/` 是 **Tmux-FSM 的执行中枢与系统权威层**，负责：

- 接收上层（cmd / fsm / ui / intent）的执行请求
- 将「意图（Intent）」转化为 **可验证、可回放、可同步的语义事件**
- 作为 **唯一权威仲裁者**，统一管理：
  - 状态演化
  - 事务边界
  - CRDT 合并
  - 策略（Policy）
  - Replay / Snapshot / WAL

⚠️ 注意：  
`engine/` 内部 **存在两种不同层级的 Engine 形态**，它们职责完全不同，但共同构成完整执行链。

---

## Engine 的双层结构（非常重要）

### 1️⃣ HeadlessEngine（权威执行引擎）

**HeadlessEngine 是系统中唯一的“真实 Engine”。**

它负责一切**不可绕过、不可旁路的权威职责**：

- ✅ 状态唯一来源（CRDT SemanticEvent）
- ✅ 事务化提交
- ✅ Replay / Snapshot
- ✅ Policy 校验
- ✅ Index 查询
- ✅ WAL / 同步 / 合并

> 任何会“改变世界状态”的行为，最终都必须进入 HeadlessEngine。

---

### 2️⃣ ConcreteEngine（编辑语义计算引擎）

**ConcreteEngine 不是权威引擎，而是一个“Intent → 编辑语义”的计算层。**

它的职责是：

- 解析编辑 Intent（Motion / TextObject / Find / Goto）
- 根据当前光标状态，计算：
  - MotionRange
  - Cursor 变化
- 提供 Vim 风格的编辑语义解释

它 **不负责**：

- ❌ CRDT
- ❌ Replay
- ❌ Policy
- ❌ WAL
- ❌ 多人同步

> ConcreteEngine 的本质是一个 **“编辑语义编译器”**，而不是事务引擎。

---

## 整体架构关系

```
Intent
  ↓
ConcreteEngine        （语义计算 / Motion / Range）
  ↓
SemanticEvent
  ↓
HeadlessEngine        （唯一权威 / 事务 / CRDT / Replay）
  ↓
Kernel
  ↓
Backend
```

---

## 核心设计原则

### ✅ Intent-first

- Engine 本身不直接操作最终状态
- 一切状态变化都来源于 `crdt.SemanticEvent`

---

### ✅ 唯一权威仲裁（Architecture Rule #4）

- Policy 校验
- 状态演化
- 冲突解决

**只能发生在 HeadlessEngine**

---

### ✅ 事务化（Transactional）

- 每一次执行都有清晰边界
- 可回放、可验证、可重建

---

### ✅ Replayable / Verifiable

- 任意状态都可通过事件日志重建
- Snapshot 只是优化，不是权威

---

## 文件结构说明

---

## `engine.go`

### 角色

- 定义 **Engine 权威接口**
- 提供 **HeadlessEngine 的无 UI 实现**

---

### Engine 接口能力

#### 状态与事务

```go
Apply(event crdt.SemanticEvent) error
Replay(upTo crdt.EventID) replay.TextState
Snapshot() *Snapshot
```

---

#### WAL / 同步

```go
Append(event crdt.SemanticEvent) crdt.EventID
WALSince(id crdt.EventID) []wal.SemanticEvent
Integrate(events []wal.SemanticEvent) error
KnownHeads() map[crdt.ActorID]crdt.EventID
```

---

#### CRDT 位置管理

```go
AllocatePosition(after, before *crdt.PositionID) crdt.PositionID
ComparePosition(a, b crdt.PositionID) int
```

---

#### Selection 管理

```go
ApplySelection(actor crdt.ActorID, fact selection.SetSelectionFact)
GetSelection(cursorID selection.CursorID)
GetAllSelections()
```

---

#### Policy / Actor

```go
RegisterActor(actorID crdt.ActorID, level policy.TrustLevel, name string)
CheckPolicy(event crdt.SemanticEvent) error
```

---

#### Index 查询

```go
QueryByActor(actor crdt.ActorID)
QueryByType(ft index.FactType)
QueryByTimeRange(start, end time.Time)
QueryAIChanges(aiActorPrefix string)
```

---

### HeadlessEngine 特点

- 无 UI、无编辑器假设
- 完全基于事件驱动
- 可用于：
  - 本地执行
  - 远程同步
  - AI Actor
  - 回放 / 审计

---

## `concrete_engine.go`

### 角色

**ConcreteEngine = 编辑语义计算引擎**

用于解释「用户在编辑器里 *想做什么*」。

---

### 主要职责

#### Motion 计算

```go
ComputeMotion(m *intent.Motion) (editor.MotionRange, error)
```

支持：

- Word / Line / Char
- Goto
- Find / Till（前向 / 后向）
- TextObject（word / paren / quote）
- Inner / Around 语义

---

#### 编辑操作（语义层）

```go
MoveCursor(r editor.MotionRange)
DeleteRange(r editor.MotionRange)
YankRange(r editor.MotionRange)
ChangeRange(r editor.MotionRange)
```

⚠️ 注意：  
这些操作 **只代表“编辑意图的几何语义”**，  
并不意味着最终状态已经被提交。

---

### 设计约束（非常重要）

ConcreteEngine **必须保持以下特性**：

- ✅ 纯计算（deterministic）
- ✅ 不依赖 CRDT
- ✅ 不接触 WAL / Policy
- ✅ 可被替换 / 测试

---

## Snapshot

```go
type Snapshot struct {
    At    crdt.EventID
    State replay.TextState
}
```

- Snapshot 是优化手段
- **不是权威**
- 任意 Snapshot 都必须可由 Replay 重建

---

## 总结一句话

> **Engine 模块是 Tmux-FSM 的“执行宪法”。**  
>
> - ConcreteEngine 负责「理解人类编辑语义」
> - HeadlessEngine 负责「裁决、记录并重放现实」

两者职责严格分离，  
共同保证系统 **安全、可追溯、可演化**。

---
```

