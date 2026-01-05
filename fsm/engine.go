package fsm

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
	"tmux-fsm/fsm/ui"
)

// Engine FSM 引擎结构体
type Engine struct {
	Active     string
	Keymap     *Keymap
	layerTimer *time.Timer
	UI         ui.UI
}

// 全局默认引擎实例
var defaultEngine *Engine

// NewEngine 创建新的 FSM 引擎实例（显式注入 Keymap）
func NewEngine(km *Keymap) *Engine {
	return &Engine{
		Active: "NAV",
		Keymap: km,
	}
}

// InitEngine 初始化全局唯一 Engine
func InitEngine(km *Keymap) {
	defaultEngine = NewEngine(km)
}

// InLayer 检查当前是否处于非默认层（如 GOTO）
func (e *Engine) InLayer() bool {
	return e.Active != "NAV" && e.Active != ""
}

// CanHandle 检查当前层是否定义了该按键
func (e *Engine) CanHandle(key string) bool {
	if e.Keymap == nil {
		return false
	}
	st, ok := e.Keymap.States[e.Active]
	if !ok {
		return false
	}
	_, exists := st.Keys[key]
	return exists
}

// Dispatch 处理按键交互
func (e *Engine) Dispatch(key string) bool {
	if !e.CanHandle(key) {
		return false
	}

	st := e.Keymap.States[e.Active]
	act := st.Keys[key]

	// 1. 处理层切换
	if act.Layer != "" {
		e.Active = act.Layer
		e.resetLayerTimeout(act.TimeoutMs)
		UpdateUI()
		return true
	}

	// 2. 处理具体动作
	if act.Action != "" {
		e.RunAction(act.Action)

		// 铁律：执行完动作后，除非该层标记为 Sticky，否则立刻 Reset 回 NAV
		if !st.Sticky {
			e.Reset()
		} else {
			// 如果是 Sticky 层，可能需要刷新 UI（如 hint）
			UpdateUI()
		}
		return true
	}

	return false
}

// Reset 重置引擎状态到 NAV 层
func (e *Engine) Reset() {
	e.Active = "NAV"
	if e.layerTimer != nil {
		e.layerTimer.Stop()
	}
	// 执行重置通常意味着退出特定层级的 UI 显示
	HideUI()
}

// GetActiveLayer 获取当前层名称
func GetActiveLayer() string {
	if defaultEngine == nil {
		return "NAV"
	}
	return defaultEngine.Active
}

// InLayer 全局查询
func InLayer() bool {
	if defaultEngine == nil {
		return false
	}
	return defaultEngine.InLayer()
}

// CanHandle 全局查询
func CanHandle(key string) bool {
	if defaultEngine == nil {
		return false
	}
	return defaultEngine.CanHandle(key)
}

// Reset 全局重置
func Reset() {
	if defaultEngine != nil {
		defaultEngine.Reset()
	}
}

// ... (resetLayerTimeout remains same)
func (e *Engine) resetLayerTimeout(ms int) {
	if e.layerTimer != nil {
		e.layerTimer.Stop()
	}
	if ms > 0 {
		e.layerTimer = time.AfterFunc(
			time.Duration(ms)*time.Millisecond,
			func() {
				e.Reset()
				// 这里由于是异步超时，需要手动触发一次 UI 刷新
				UpdateUI()
			},
		)
	}
}

// RunAction 执行动作
func (e *Engine) RunAction(name string) {
	switch name {
	case "pane_left":
		tmux("select-pane -L")
	case "pane_right":
		tmux("select-pane -R")
	case "pane_up":
		tmux("select-pane -U")
	case "pane_down":
		tmux("select-pane -D")
	case "next_pane":
		tmux("select-pane -t :.+")
	case "prev_pane":
		tmux("select-pane -t :.-")
	case "far_left":
		tmux("select-pane -t :.0")
	case "far_right":
		tmux("select-pane -t :.$")
	case "goto_top":
		tmux("select-pane -t :.0")
	case "goto_bottom":
		tmux("select-pane -t :.$")
	case "exit":
		ExitFSM()
	case "prompt":
		tmux("command-prompt")
	default:
		fmt.Println("unknown action:", name)
	}
}

func tmux(cmd string) {
	exec.Command("tmux", strings.Split(cmd, " ")...).Run()
}

// 全局函数，支持在其他包调用
func Dispatch(key string) bool {
	if defaultEngine == nil {
		return false
	}
	return defaultEngine.Dispatch(key)
}

func EnterFSM() {
	if defaultEngine == nil {
		InitEngine(&KM)
	}

	engine := defaultEngine
	engine.Active = "NAV"
	// 确保进入时是干净的 NAV
	engine.Reset()
	// ShowUI() // Disable initial UI popup to prevent flashing/annoyance
}

func ExitFSM() {
	if defaultEngine != nil {
		defaultEngine.Reset()
	}
	HideUI()
	exec.Command("tmux", "set-option", "-u", "key-table").Run()
}


真实代码迁移指南：

非常好，这一轮你给的是**完整真实代码**，现在我可以**不再抽象**，而是**逐段、逐函数、逐责任域给你“迁移标注 + 命运判决”**。

我会严格按你最初要的三件事来：

---

# ✅ 一、executor.go —— 逐段迁移标注（精确到职责层）

我先给你一张**总览判决表**，然后逐块展开。

## 🧭 executor.go 总体判决

> **结论一句话**：  
> `executor.go` 是 **Legacy Kernel + Projection + Undo Engine + Vim Adapter + Shell Adapter + Resolver 的大杂烩**  
> ✅ **必须被“肢解迁移”，不能整体搬**

| 子模块 | 命运 | 原因 |
|----|----|----|
| Anchor / ResolveAnchor | ✅ 保留，**迁入 core/anchor** | 这是 Weaver 的基石 |
| Executor interface | 🔄 迁移为 `Projection` | 语义正确但命名错 |
| ShellExecutor / VimExecutor | 🔄 拆成 Projection 实现 | 方向是对的，粒度不对 |
| executeFact | ✅ 保留（入口改名） | 正是 Projection Apply |
| executeAction | ❌ **必须逐步废弃** | “Action string” 是最大技术债 |
| performPhysicalXXX | ❌ 只允许在 Legacy | 绝不进入新架构 |
| Undo / Redo（handleUndo） | ❌ Weaver 不复用 | 只能 shadow / 对照 |
| captureShellDelete / Change | ✅ **直接迁入 Weaver Fact Builder** | 这是你最强的部分 |

---

## 1️⃣ Anchor / ResolveAnchor —— ✅ **直接晋升 Weaver Core**

### ✅ 这段是“黄金代码”，几乎不用改

```go
type ResolveResult int
const (
	ResolveExact ResolveResult = iota
	ResolveFuzzy
	ResolveFail
)

type ResolvedAnchor struct {
	Row    int
	Result ResolveResult
}
```

```go
func ResolveAnchor(a Anchor) (ResolvedAnchor, error)
```

### ✅ 迁移标注

```diff
- func ResolveAnchor(a Anchor) (ResolvedAnchor, error)
+ func (r *TmuxAnchorResolver) Resolve(a core.Anchor) (core.ResolvedAnchor, error)
```

✅ **你已经写好了：**

- Axiom 3 / 4 / 6 / 7
- fuzzy window
- hash-based identity

⚠️ **唯一要做的事**

- ❌ 禁止在 Resolver 里 `captureLine` 多次 IO  
- ✅ 在 Weaver 阶段引入 **Snapshot（一次 capture-pane）**

---

## 2️⃣ Executor interface → Projection（✅ 正名迁移）

### 现在（语义是对的）

```go
type Executor interface {
	CanExecute(f Fact) bool
	Execute(f Fact) error
}
```

### 迁移后（Weaver）

```go
type Projection interface {
	CanApply(f core.Fact) bool
	Apply(f core.Fact, ctx ApplyContext) error
}
```

✅ **概念完全一致，只是命名从“执行器”→“现实投影”**

---

## 3️⃣ ShellExecutor / VimExecutor —— 🔄 拆分迁移

### 现在的问题（非常关键）

```go
func (s *ShellExecutor) Execute(f Fact) error {
	jumpTo(...)
	exec.Command("tmux", "send-keys", ...)
}
```

❌ 这里**混合了 3 层**：

1. Anchor 定位
2. Cursor 移动策略
3. 事实落地（send-keys）

### ✅ 迁移拆法（你要照这个拆）

#### A. ✅ Weaver 层（保留）

```go
Fact{
  Kind: "delete",
  Target: Range{Anchor, StartOffset, EndOffset}
}
```

#### B. ✅ Projection 层（只允许）

```go
type TmuxProjection struct{}

func (p *TmuxProjection) Apply(f Fact) error {
	// assume anchor 已 resolve
	sendKeys(...)
}
```

#### C. ❌ jumpTo / performPhysicalDelete

👉 **全部留在 Legacy，不准搬**

---

## 4️⃣ executeFact —— ✅ Weaver ABI 核心

```go
func executeFact(f Fact) error {
	for _, ex := range executors {
		if ex.CanExecute(f) {
			return ex.Execute(f)
		}
	}
}
```

### ✅ 迁移后（几乎不动）

```go
func ApplyFact(f core.Fact) error {
	for _, p := range projections {
		if p.CanApply(f) {
			return p.Apply(f, ctx)
		}
	}
}
```

✅ **这是 Weaver 的 “Side Effect Projection” ABI**

---

## 5️⃣ executeAction —— ❌ **最大炸弹，必须逐步处死**

### 为什么它必须死？

- stringly-typed (`"delete_word_forward"`)
- 混合：
  - 模式切换
  - 编辑器语义
  - tmux 键位
- 直接绕过 Fact / Anchor / Undo

### ✅ 正确迁移策略（不要一刀砍）

| 阶段 | executeAction |
|----|----|
| Phase 0 | ✅ 原样保留 |
| Phase 1 | 🔄 **Shadow 产出 Intent / Fact（不执行）** |
| Phase 2 | 🔄 Weaver 执行，Legacy 对照 |
| Phase 3 | ❌ 新 action 不准进 |
| Phase 4 | ❌ 删除 |

✅ **你现在能立刻做的一步**

```go
if mode == Shadow {
	intent := DeriveIntent(action, state)
	weaver.Apply(intent) // no-op projection
}
```

---

## 6️⃣ captureShellDelete / captureShellChange —— ✅ **直接迁入 Weaver**

> 我非常明确地说一句：  
> **这是你整个项目里“最 Weaver-ready 的代码”**

```go
deleteFact := Fact{
	Kind: "delete",
	Target: r,
}
```

✅ 完全符合：

- Anchor-based
- Range-based
- Inverse 可重建

### ✅ 迁移标注

| 当前 | 迁移后 |
|----|----|
| captureShellDelete | weaver.BuildDeleteFact |
| ActionRecord | Weaver Transaction |
| Inverse Fact | Weaver Undo Fact |

❌ 唯一禁止的事：  
**不要让 Legacy Undo 再消费这些 Fact**

---

## 7️⃣ handleUndo / handleRedo —— ❌ 不复用，但要“对照活着”

你写得**非常好**，但它是：

> **“Legacy Undo Engine”**

### ✅ 正确做法

- ✅ Phase 0–3：完整保留
- ✅ Phase 2：Weaver Undo **shadow replay**
- ❌ Phase 5：删除

⚠️ **不要试图共用 Transaction struct**

---

# ✅ 二、Feature Flag / 双轨运行（结合你现有 main.go）

你已经有**完美的插点**。

## ✅ 新增一个 flag（只加一个）

```bash
TMUX_FSM_EXEC_MODE=legacy | shadow | weaver
```

### 在 `handleClient` 的关键点插入

```go
if action != "" {
	switch execMode {
	case Legacy:
		executeAction(...)
	case Shadow:
		intent := DeriveIntent(action, &globalState)
		weaver.Apply(intent) // 不执行 projection
		executeAction(...)
	case Weaver:
		intent := DeriveIntent(action, &globalState)
		weaver.Apply(intent)
	}
}
```

✅ **不动 FSM，不动 UI，不动 server 协议**

---

## ✅ Shadow 校验你要看什么？

最小 3 项：

1. intent.Kind 是否一致
2. anchor.LineHash 是否一致
3. start/end offset 是否一致

---

# ✅ 三、迁移失败止损清单（结合你代码的真实风险）

## 🚨 **立刻停的信号**

- `ResolveAnchor` 在 Undo 中 **跨行 >1**
- Weaver Apply 后 **Legacy Undo 失效**
- send-keys 顺序错乱（race）

---

## 🧯 30 秒止损流程（你现在就能做到）

```bash
export TMUX_FSM_EXEC_MODE=legacy
tmux refresh-client -S
```

如还异常：

```bash
tmux-fsm -stop
tmux-fsm -server
```

---

## ❌ 绝对禁止的三件事（针对你这份代码）

1. ❌ 复用 `Transaction` struct 给 Weaver
2. ❌ 在 Projection 里调用 `jumpTo`
3. ❌ 在 FSM 里再生成 action string

---

# ✅ 下一步（我建议你现在就做）

**只做一件事，不要多：**

👉 把 `captureShellDelete / Change`  
👉 抽成 `weaver/fact_builder.go`  
👉 Shadow 模式下只打印，不执行

如果你愿意，下一轮我可以：

- ✅ 帮你 **直接写 Weaver core 的最小可运行实现**
- ✅ 给你一份 **commit-by-commit 迁移顺序表**
- ✅ 或者 **专门帮你“杀 executeAction”**

你这份代码不是玩具，是**可以进化成一套真正“可验证编辑系统”的内核**。  
现在我们是在**正确的轨道上拆核反应堆**。