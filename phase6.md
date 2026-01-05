好，这里我给你一个**非常明确的技术裁决**，而不是“二选一式建议”。

---

# ✅ 决策结论（强烈）

👉 **直接进入 Phase 6 — Snapshot & Time Travel**  
⏸️ **Phase 5.5 暂缓，不先做**

原因不是“可以以后再优化”，而是：

> **Phase 5.5 的所有工作，在没有 Snapshot 之前都会返工。**

你现在的系统已经**结构性地要求 Snapshot**，否则继续打磨只会加深技术债。

---

## 为什么 **不能先做 Phase 5.5**

我们逐条拆解你列的 5.5 内容：

### ❌ 冲突合并策略
- 冲突的本质是：  
  **Intent 期望的世界 ≠ 执行时的世界**
- 没有 Snapshot：
  - Planner 用的是 *瞬时现实*
  - Resolver 又重新读现实
  - “冲突”根本没有统一参照系

✅ **Snapshot 才是冲突判定的坐标系**

---

### ❌ 性能优化
你现在的性能瓶颈是：

- 多次 tmux IO
- Planner / Resolver 重复 capture
- Hash 反复计算

这些在 Phase 6 会自然消失：

```text
tmux → Snapshot → Planner / Resolver / Projection
```

现在做性能优化 = 优化即将被删除的代码。

---

### ❌ Projection 去 motion 化
Projection 的 “motion” 问题本质是：

> Projection 仍然在感知“现实正在变化”

Snapshot 出现后：

- Projection = **纯函数**
- 输入：ResolvedFacts + Snapshot
- 输出：Commands

✅ motion 会自然消失，而不是“被重构掉”

---

## ✅ 为什么 **现在必须进入 Phase 6**

你在 Phase 5.4 已经做了这三件“不可逆”的事：

1. ✅ Anchor 携带 Hash（Expectation）
2. ✅ Resolver 执行前验证 Reality
3. ✅ ShadowEngine 拒绝不一致世界

这意味着：

> **系统已经不再信任“当前现实”**

这正是 Snapshot 存在的理由。

---

# 🚀 Phase 6 的真实目标（不是表面理解）

> ❌ 不是“缓存 tmux 状态”  
> ✅ 而是 **引入时间维度**

Phase 6 要解决的是：

```text
What did the world look like
when this Intent was formed?
```

---

## 🧠 Phase 6 的核心设计（预告）

### 1️⃣ Snapshot（不可变世界）

```go
type Snapshot struct {
    PaneID   string
    Cursor   CursorPos
    Lines    []Line
    HashTree SnapshotHash
    TakenAt  time.Time
}
```

- ✅ 一次 tmux IO
- ✅ 所有组件共享
- ✅ 可复现、可审计

---

### 2️⃣ Planner 不再读 tmux

```text
Before: Planner → tmux
After:  Planner → Snapshot
```

Planner 成为 **纯推理组件**

---

### 3️⃣ Resolver = Snapshot → Reality 校验器

Resolver 的角色会升级为：

> “Snapshot 是否仍然可以被安全投射到当前 Reality”

Hash 校验会移动到 **Snapshot 层级**

---

### 4️⃣ Projection 变成纯函数

```go
func Project(snapshot Snapshot, facts []ResolvedFact) []Command
```

- ✅ 无 tmux
- ✅ 无 side-effect
- ✅ 可 time-travel / dry-run / rewind

---

### 5️⃣ Time Travel（不是噱头）

一旦 Snapshot 存在，你自动获得：

- ⏪ Replay
- 🧪 Deterministic testing
- 🔁 Undo / Redo 不再脆弱
- 🧾 完整审计链

---

## ✅ 推荐执行顺序（强烈）

```text
Phase 6.1  Snapshot 数据结构 & Adapter
Phase 6.2  Planner → Snapshot 化
Phase 6.3  Resolver → Snapshot 校验
Phase 6.4  Projection → 纯函数
Phase 6.5  Time Travel / Replay
```

👉 **Phase 5.5 会自然“融化”在 Phase 6 里**

---

## ✅ 我的建议（直说）

你现在的系统：

- 已经是 **执行引擎**
- 正在进化成 **时间一致性引擎**

此时如果停下来“打磨 Phase 5.5”，  
相当于在**即将起飞前给飞机打蜡**。

---

### ✅ 下一步只需你一句话

> **「开始 Phase 6.1：Snapshot 设计」**

我会直接给你：
- Snapshot 的最小可行结构
- Adapter 接口定义
- 第一次代码落点（不破坏现有 Phase 5）

我们继续。


很好。  
**Phase 6.1 正式开始：Snapshot 设计（世界冻结点）**

这一阶段我们只做一件事：

> **引入一个“不可变的世界视图”，并且不破坏 Phase 5 的任何行为。**

不做 Planner 重写、不做 Resolver 改造、不动 Projection。  
**只加一层：Snapshot。**

---

# Phase 6.1 目标（严格约束）

✅ 一次 tmux IO  
✅ 捕获执行所需的 *最小充分现实*  
✅ 不可变、可 hash、可复用  
✅ 可以被 Planner / Resolver / Projection 共享  
✅ **暂时不参与执行流程（Phase 6.2 才接入）**

---

# 一、Snapshot 的真实职责（不要误解）

Snapshot **不是缓存**，也不是状态同步。

它是：

> **Intent 形成时，对“世界”的数学定义**

换句话说：

```text
Intent + Snapshot = 可被验证、可被重放的历史事实
```

---

# 二、最小可行 Snapshot（MVP）

我们从 **Shell / tmux** 的最小世界开始。

## ✅ Snapshot 核心结构

📄 `weaver/core/snapshot.go`

```go
package core

import "time"

// Snapshot 世界快照（不可变）
type Snapshot struct {
    PaneID string

    Cursor CursorPos
    Lines  []LineSnapshot

    Hash   SnapshotHash
    TakenAt time.Time
}

// CursorPos 光标位置
type CursorPos struct {
    Row int
    Col int
}

// LineSnapshot 单行快照
type LineSnapshot struct {
    Row  int
    Text string
    Hash LineHash
}

// 强类型 Hash（避免 string 污染）
type LineHash string
type SnapshotHash string
```

### 设计说明（重要）

- ✅ **Row 显式存储**（避免依赖 slice index）
- ✅ LineHash ≠ SnapshotHash（层级清晰）
- ✅ Snapshot 本身携带 Hash（未来用于快速一致性验证）
- ✅ TakenAt 用于 Time Travel / Audit

---

# 三、Adapter 层：SnapshotBuilder

Snapshot **必须来自 Adapter**，不能从 Planner 偷读 tmux。

📄 `weaver/adapter/snapshot.go`

```go
package adapter

import "tmux-fsm/weaver/core"

// SnapshotProvider 世界读取接口
type SnapshotProvider interface {
    TakeSnapshot(paneID string) (core.Snapshot, error)
}
```

### tmux 实现（Phase 6.1 只做同步版）

📄 `weaver/adapter/tmux_snapshot.go`

```go
package adapter

import (
    "time"
    "tmux-fsm/weaver/core"
)

type TmuxSnapshotProvider struct{}

func (p *TmuxSnapshotProvider) TakeSnapshot(paneID string) (core.Snapshot, error) {
    cursor := TmuxGetCursorPos(paneID)
    lines := TmuxCapturePane(paneID)

    var snapLines []core.LineSnapshot
    for i, line := range lines {
        snapLines = append(snapLines, core.LineSnapshot{
            Row:  i,
            Text: line,
            Hash: core.LineHash(TmuxHashLine(line)),
        })
    }

    snapshot := core.Snapshot{
        PaneID: paneID,
        Cursor: core.CursorPos{
            Row: cursor[0],
            Col: cursor[1],
        },
        Lines:   snapLines,
        TakenAt: time.Now(),
    }

    snapshot.Hash = computeSnapshotHash(snapshot)
    return snapshot, nil
}
```

> ⚠️ `TmuxCapturePane` 是新函数（整 pane capture），  
> Phase 6.1 **允许粗暴实现**

---

## Snapshot Hash（必须）

📄 `weaver/adapter/snapshot_hash.go`

```go
package adapter

import (
    "crypto/sha256"
    "encoding/hex"
    "tmux-fsm/weaver/core"
)

func computeSnapshotHash(s core.Snapshot) core.SnapshotHash {
    h := sha256.New()

    h.Write([]byte(s.PaneID))
    for _, line := range s.Lines {
        h.Write([]byte(line.Hash))
    }

    return core.SnapshotHash(hex.EncodeToString(h.Sum(nil)))
}
```

✅ Snapshot Hash = 世界整体指纹  
✅ Line Hash = 局部验证工具

---

# 四、Snapshot 的“不变量”（Phase 6 的根基）

你现在必须**强约束**以下规则：

### ✅ 不变量 1：Snapshot 不可变

- 无 setter
- 无 UpdateLine
- 无 Append

> Snapshot 是 **值对象（Value Object）**

---

### ✅ 不变量 2：Snapshot ≠ Reality

Snapshot **不会自动同步** tmux。

Reality 变化 → Snapshot 失效  
这是 **正确行为**

---

### ✅ 不变量 3：任何执行必须绑定 Snapshot（Phase 6.2）

现在还没做，但这是接下来强制执行的规则。

---

# 五、Phase 6.1 不做的事（非常重要）

❌ 不改 ShellFactBuilder  
❌ 不改 Planner  
❌ 不改 Resolver  
❌ 不改 Projection  
❌ 不改 ShadowEngine

我们只是在系统中 **“引入时间”**，但还没使用它。

---

# 六、你现在应该做的提交（Checklist）

✅ 新增 `core/snapshot.go`  
✅ 新增 `adapter/SnapshotProvider`  
✅ 新增 `TmuxSnapshotProvider`  
✅ 新增 Snapshot Hash 计算  
✅ 不影响现有 Phase 5 流程

---

# 七、完成后的状态（你将站在这里）

```text
tmux ──► Snapshot ──► (尚未接入)
```

下一步（Phase 6.2）我们会做：

```text
Snapshot ──► Planner
```

Planner **彻底断开 tmux**。

---

## ✅ 下一步指令

完成 Phase 6.1 后，直接告诉我：

> **「Phase 6.1 完成，进入 Phase 6.2」**

我将带你做最关键的一步：

> **Planner 去现实化（Planner becomes pure reasoning）**

这是 Weaver 从“自动化工具”进化为  
**时间一致性推理引擎**的拐点。

好。  
**Phase 6.2 正式开始：Planner 去现实化（Planner → 纯推理）**

这是整个 Phase 6 的**拐点阶段**。  
从这一刻起，Planner **不再拥有读取现实的权力**。

---

# Phase 6.2 目标（不可妥协）

✅ Planner **只接受 Snapshot**  
✅ Planner **不再 import tmux / adapter**  
✅ Planner 输出的 Intent **隐式绑定 Snapshot**  
✅ 现有 Phase 5 行为不破坏（功能等价）

一句话：

> **Planner = Snapshot → Intent（纯函数）**

---

# 一、核心裁决：Planner 的输入签名必须改变

## ❌ 旧模型（必须废弃）

```go
func Plan(paneID string) Intent {
    facts := BuildFactsFromTmux(paneID)
    ...
}
```

Planner 直接感知现实 —— **这是不允许的**。

---

## ✅ 新模型（Phase 6.2）

```go
func Plan(snapshot core.Snapshot) Intent {
    facts := BuildFactsFromSnapshot(snapshot)
    ...
}
```

Planner 不知道 tmux 存不存在。  
它只“观察一个被冻结的世界”。

---

# 二、Snapshot → Facts（桥接层）

我们**不改 Fact 的语义**，只改来源。

## 新接口：SnapshotFactBuilder

📄 `weaver/facts/from_snapshot.go`

```go
package facts

import "tmux-fsm/weaver/core"

// BuildFactsFromSnapshot 将世界快照转换为 Planner Facts
func BuildFactsFromSnapshot(s core.Snapshot) []Fact {
    var facts []Fact

    facts = append(facts, CursorAt{
        Row: s.Cursor.Row,
        Col: s.Cursor.Col,
    })

    for _, line := range s.Lines {
        if IsPrompt(line.Text) {
            facts = append(facts, PromptAt{
                Row: line.Row,
            })
        }

        if IsErrorLine(line.Text) {
            facts = append(facts, ErrorLine{
                Row:  line.Row,
                Text: line.Text,
            })
        }
    }

    return facts
}
```

✅ 所有 `IsXxx(...)` 逻辑保持不变  
✅ 只是输入从 string → Snapshot

---

# 三、Planner 的最小改造路径（不炸系统）

## 1️⃣ 新 Planner 函数（不要立刻删旧的）

📄 `weaver/planner/plan_with_snapshot.go`

```go
package planner

import (
    "tmux-fsm/weaver/core"
    "tmux-fsm/weaver/facts"
)

func PlanWithSnapshot(s core.Snapshot) Intent {
    fs := facts.BuildFactsFromSnapshot(s)
    return PlanFromFacts(fs)
}
```

✅ `PlanFromFacts` 是你 Phase 5 已有的逻辑  
✅ Planner 的核心算法 **零改动**

---

## 2️⃣ 执行入口（临时桥接）

在你当前的执行入口（比如 `ExecuteIntent`）：

```go
snapshot, _ := snapshotProvider.TakeSnapshot(paneID)
intent := planner.PlanWithSnapshot(snapshot)
```

⚠️ Resolver / Projection 仍然会读 tmux  
这是 **Phase 6.3 才解决的**

---

# 四、Intent 必须开始“携带时间”

**Phase 6.2 的关键副作用：**

Planner 输出的 Intent **语义上绑定 Snapshot**。

## ✅ 最小做法（现在就要）

📄 `weaver/core/intent.go`

```go
type Intent struct {
    Goal string
    Steps []Step

    SnapshotHash SnapshotHash // ← 新增（必须）
}
```

在 Planner 中：

```go
intent.SnapshotHash = s.Hash
```

> ⚠️ 注意：**不是 Snapshot 本体**
> 只携带 Hash，避免意外耦合

---

# 五、你现在拥有的能力（即使还没用）

即便 Resolver 还没改，你已经能做到：

- ✅ 知道 Intent 是“在哪个世界”生成的
- ✅ 能拒绝在错误世界执行（Phase 6.3）
- ✅ 可重放 Planner 决策（给同一个 Snapshot）

---

# 六、Phase 6.2 的危险红线（千万别踩）

❌ Planner import adapter  
❌ Planner 调 tmux  
❌ Planner 调 SnapshotProvider  
❌ Planner 在内部生成 Snapshot  

> Planner **永远不知道 Snapshot 是怎么来的**

---

# 七、Phase 6.2 完成检查表

✅ Planner 函数签名接受 Snapshot  
✅ FactBuilder 支持 Snapshot  
✅ Intent 携带 SnapshotHash  
✅ 执行入口注入 Snapshot  
✅ Phase 5 行为仍然正确

---

# 八、系统此刻的真实形态

```text
tmux ──► Snapshot ──► Planner ──► Intent(hash)
                    ↑
              时间在这里冻结
```

Resolver / Projection 仍在“现实时间”  
但 Planner 已经 **时间安全**。

---

## ✅ 下一步指令

完成后，直接告诉我：

> **「Phase 6.2 完成，进入 Phase 6.3」**

下一步是**整个系统一致性的关键**：

> **Resolver：Reality vs Snapshot 的最终裁决者**

这里我们会：
- 阻止过期 Intent 执行
- 引入“安全拒绝”而不是“错误执行”

你已经站在时间引擎的门口了。


很好。  
**Phase 6.3 正式开始：Resolver 时间一致性裁决（Reality vs Snapshot）**

这是 Phase 6 的**终点关**。  
从这里开始，Weaver **不再盲目执行任何 Intent**。

> **Resolver 成为“现实与历史之间的法官”。**

---

# Phase 6.3 目标（硬约束）

✅ Resolver **必须验证 SnapshotHash**  
✅ Resolver **在不一致时拒绝执行（安全失败）**  
✅ Resolver **永不修复、不重算、不补救**  
✅ Projection 仍然可以直接操作现实（下一阶段再改）

一句话：

> **Intent ≠ 命令  
> Intent = 在某个世界成立的计划**

---

# 一、核心新概念：Reality Snapshot

Resolver 在执行前 **重新看一眼现实**。

📌 但它不是 Planner 的 Snapshot  
📌 它只是用于 **一致性对比**

---

## 新接口：RealityReader

📄 `weaver/adapter/reality.go`

```go
package adapter

import "tmux-fsm/weaver/core"

// RealityReader 读取当前世界状态（用于一致性验证）
type RealityReader interface {
    ReadCurrent(paneID string) (core.Snapshot, error)
}
```

### tmux 实现（复用 SnapshotProvider）

```go
type TmuxRealityReader struct {
    SnapshotProvider
}

func (r *TmuxRealityReader) ReadCurrent(paneID string) (core.Snapshot, error) {
    return r.TakeSnapshot(paneID)
}
```

✅ RealityReader = SnapshotProvider 的语义别名  
✅ 职责不同，名字必须不同

---

# 二、Resolver 的新入口（关键）

## ❌ 旧 Resolver（不安全）

```go
func Resolve(intent Intent) {
    ExecuteSteps(intent.Steps)
}
```

---

## ✅ 新 Resolver（Phase 6.3）

📄 `weaver/resolver/resolve_with_guard.go`

```go
package resolver

import (
    "errors"
    "tmux-fsm/weaver/adapter"
    "tmux-fsm/weaver/core"
)

var ErrWorldDrift = errors.New("world drift: snapshot mismatch")

type Resolver struct {
    Reality adapter.RealityReader
}

func (r *Resolver) Resolve(paneID string, intent core.Intent) error {
    current, err := r.Reality.ReadCurrent(paneID)
    if err != nil {
        return err
    }

    if current.Hash != intent.SnapshotHash {
        return ErrWorldDrift
    }

    return ExecuteSteps(intent.Steps)
}
```

✅ **唯一裁决条件：Hash 是否相等**  
✅ 不对差异做分析  
✅ 不尝试修复  
✅ 不降级执行

> Resolver 是法官，不是医生。

---

# 三、ErrWorldDrift 的系统语义（非常重要）

**ErrWorldDrift ≠ 错误**

它表示：

> *“这个 Intent 在当前世界不再成立”*

这是一个 **正确、安全、成功的拒绝**。

---

## 上层必须这样对待：

✅ 记录  
✅ 丢弃 Intent  
✅ （可选）重新 Planner + 新 Snapshot  

❌ 重试旧 Intent  
❌ 忽略  
❌ 强制执行  

---

# 四、你现在获得的系统性质（这是质变）

从这一刻起：

### ✅ 时间一致性（Temporal Safety）

- Planner 在过去
- Resolver 在现在
- 不一致 → 停止

---

### ✅ 执行可证明性

你可以回答：

> “为什么这个命令被执行？”

答案是：

> “因为世界仍然和 Planner 看到的一样。”

---

### ✅ Weaver 不会再“乱敲键盘”

这不是“更聪明”，  
这是 **更有纪律**。

---

# 五、Phase 6.3 禁区（任何一条都是架构退化）

❌ Resolver 调 Planner  
❌ Resolver 生成 SnapshotHash  
❌ Resolver 修复 Intent  
❌ Resolver 忽略 Hash 不匹配  
❌ Resolver 访问 Facts

---

# 六、Phase 6.3 完成清单

✅ 新增 RealityReader  
✅ Resolver 在执行前对比 SnapshotHash  
✅ ErrWorldDrift 作为合法结果  
✅ 上层能识别并中止执行  
✅ Projection 未改（允许）

---

# 七、系统最终形态（Phase 6 完成）

```text
tmux ──► Snapshot ──► Planner ──► Intent(hash)
   ▲                                   │
   │                                   ▼
   └──────── Reality ◄── Resolver ◄── Projection
```

这是一个 **时间闭环**。

---

## ✅ Phase 6 完成确认

如果你已经全部接好，请直接告诉我：

> **「Phase 6 完成」**

接下来我会问你一个关键问题：

> **你想让 Weaver 成为“确定性引擎”，还是“可回溯推理系统”？**

这将决定 Phase 7 是 **Caching / Replay**  
还是 **Branching / Simulation**。

你已经完成了一次真正的系统跃迁。