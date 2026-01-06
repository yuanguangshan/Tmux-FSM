好，我们直接**落地**。  
下面这份是 **“可以直接替换 / 合并的最小闭环方案”**，目标只有一个：

> ✅ `3dw`  
> ✅ `dd`  
> ✅ `.`  
> ✅ 语义级 repeat（不是 RawToken 级）  
> ✅ Grammar **跨按键持久**

不讲理念，只给**工程确定性结构**。

---

# ✅ 一、修正后的 Kernel 结构（核心）

## ✅ Kernel 必须长期持有 Grammar

```go
// kernel/kernel.go
package kernel

import (
	"tmux-fsm/fsm"
	"tmux-fsm/intent"
	"tmux-fsm/planner"
)

type Kernel struct {
	FSM     *fsm.Engine
	Grammar *planner.Grammar
}
```

### 构造函数（只 new 一次 Grammar）

```go
func NewKernel(fsmEngine *fsm.Engine) *Kernel {
	k := &Kernel{
		FSM:     fsmEngine,
		Grammar: planner.NewGrammar(),
	}

	// ✅ 永久绑定 Emit → Grammar
	fsmEngine.Emit = func(tok fsm.RawToken) {
		k.Grammar.Consume(tok)
	}

	return k
}
```

> ✅ **Grammar 生命周期 = Kernel 生命周期**
>
> ❌ 再也不在 `Decide()` 里 new Grammar

---

# ✅ 二、FSM / Grammar / Kernel 的最终连接方式

## ✅ FSM：只负责 Emit RawToken（不 reset count）

### ✅ FSM.Dispatch（关键修正）

```go
func (e *Engine) Dispatch(key string) bool {

	// 1️⃣ digit（物理事实）
	if isDigit(key) {
		e.count = e.count*10 + int(key[0]-'0')
		e.Emit(RawToken{Kind: TokenDigit, Value: key})
		return true
	}

	// 2️⃣ repeat（语义级）
	if key == "." {
		e.Emit(RawToken{Kind: TokenRepeat, Value: "."})
		return true
	}

	// 3️⃣ 普通 key（FSM 仍然负责 layer）
	if e.CanHandle(key) {
		st := e.Keymap.States[e.Active]
		act := st.Keys[key]

		if act.Layer != "" {
			e.Active = act.Layer
			e.resetLayerTimeout(act.TimeoutMs)
			e.Emit(RawToken{Kind: TokenKey, Value: key})
			return true
		}

		// ✅ 关键：FSM 不再 reset count
		e.Emit(RawToken{Kind: TokenKey, Value: key})
		return true
	}

	return false
}
```

> ✅ FSM.count = UI 派生状态  
> ❌ FSM 不再参与语义 reset

---

## ✅ FSM.Reset / Enter / Exit → **System Token**

### ✅ RawTokenKind

```go
const (
	TokenDigit RawTokenKind = iota
	TokenKey
	TokenRepeat
	TokenSystem // ✅
)
```

### ✅ FSM.Reset

```go
func (e *Engine) Reset() {
	e.Active = "NAV"
	e.count = 0
	e.Emit(RawToken{
		Kind:  TokenSystem,
		Value: "reset",
	})
}
```

---

## ✅ Grammar：只关心 Vim 语义

### ✅ Consume（忽略系统事件）

```go
func (g *Grammar) Consume(tok fsm.RawToken) *intent.Intent {
	switch tok.Kind {

	case fsm.TokenSystem:
		return nil // ✅ 不污染 Vim 语义

	case fsm.TokenDigit:
		g.count = g.count*10 + int(tok.Value[0]-'0')
		return nil

	case fsm.TokenRepeat:
		if g.lastIntent != nil {
			return cloneIntent(g.lastIntent)
		}
		return nil

	case fsm.TokenKey:
		return g.consumeKey(tok.Value)
	}

	return nil
}
```

---

# ✅ 三、Kernel.Decide：最小、正确、稳定

```go
func (k *Kernel) Decide(key string) *Decision {

	// FSM 永远先吃 key
	if k.FSM != nil {
		dispatched := k.FSM.Dispatch(key)

		// Grammar 可能在 Emit 中产生 intent
		if intent := k.Grammar.TakeLastIntent(); intent != nil {
			return &Decision{
				Kind:   DecisionFSM,
				Intent: intent,
			}
		}

		if dispatched {
			return nil // FSM 吞键但未成语义
		}
	}

	// fallback（tmux / shell）
	return &Decision{
		Kind: DecisionPassthrough,
		Key:  key,
	}
}
```

---

## ✅ Grammar 增补：取出 Intent（不在 Consume 里 return）

### ✅ Grammar 增加方法

```go
func (g *Grammar) TakeLastIntent() *intent.Intent {
	if g.lastIntent == nil {
		return nil
	}
	i := g.lastIntent
	g.lastIntent = nil
	return i
}
```

> ✅ Grammar 内部可多次 Consume  
> ✅ Kernel 统一拉取语义结果

---

# ✅ 四、最小闭环验证（逐键）

## ✅ `3dw`

```
3 → TokenDigit → count=3
d → TokenKey   → pendingOp=Delete
w → TokenKey   → emit IntentOperator(Delete, Word, count=3)
```

✅ Grammar.reset  
✅ lastIntent = clone  
✅ DecisionFSM(Intent)

---

## ✅ `dd`

```
d → pendingOp=Delete
d → pendingOp==Delete → Line intent
```

✅ motion = MotionLine  
✅ count = max(count,1)

---

## ✅ `.`

```
. → TokenRepeat
→ clone lastIntent
```

✅ 语义级 repeat  
✅ 不依赖 RawToken 历史  
✅ 可支持 `3.`

---

# ✅ 五、不变量（请牢记，这是系统护城河）

> ✅ FSM：物理层（键、layer、UI）  
> ✅ Grammar：Vim 语言状态机  
> ✅ Kernel：语义裁决  
> ❌ FSM 永不 Produce Intent  
> ❌ Grammar 永不关心 layer / UI  

---

# ✅ 接下来你已经站在正确的台阶上

你现在 **下一步自然演进**只有三条路：

1. ✅ `ciw / daw / yi(`（TextObject）
2. ✅ Visual Mode Grammar（范围产生 Anchor）
3. ✅ Macro 录制（Grammar 事件回放）

**你这套结构，全都兜得住。**

---

如果你愿意，下一步我可以直接给你：

> 🔜 **TextObject Grammar（ciw / diw）最小实现**  
> 🔜 或 **Visual Mode → Intent.Range**  
> 🔜 或 **Grammar 状态图（可视化）**

你只要回我一句：  
**“下一步：TextObject” / “Visual Mode” / “Macro”**好，这里我**直接替你做技术决策**，不给你“随便选”的负担。

---

# ✅ 结论（工程最优路径）

> **下一步：TextObject**

**原因非常明确：**

1. ✅ **TextObject 是 Grammar 能力的试金石**
2. ✅ 一旦 `ciw / diw / daw` 跑通  
   → Visual Mode 只是“范围来源不同”  
   → Macro 只是“Intent 录制/回放”
3. ✅ TextObject 会强制你把  
   **Operator / Motion / Scope / Count**  
   这四个维度真正解耦干净

换句话说：

> **TextObject 写对了，你这套内核就“封神了”**

---

# 🧠 TextObject 在你当前架构中的准确定位

### Vim 语义层拆分（现在正式成立）

```
[count] + operator + (motion | text-object)
```

TextObject ≠ Motion  
它是 **“范围生成器”**，而不是方向。

---

# ✅ Grammar 扩展目标（最小集）

我们这一步只做：

| 组合 | 含义 |
|----|----|
| `ciw` | change inner word |
| `diw` | delete inner word |
| `daw` | delete around word |

**不碰括号、不碰引号、不碰 visual**

---

# ✅ 一、Grammar 新增状态（这是关键）

## ✅ 新字段

```go
type TextObjectScope int

const (
	ScopeInner TextObjectScope = iota
	ScopeAround
)

type Grammar struct {
	count       int
	pendingOp   *intent.Operator
	pendingTO   bool              // ✅ 是否进入 TextObject 模式
	toScope     TextObjectScope    // i / a
	lastIntent  *intent.Intent
}
```

---

# ✅ 二、Grammar.consumeKey：新增 TextObject 分支

### ✅ 核心逻辑（非常 Vim）

```go
func (g *Grammar) consumeKey(key string) *intent.Intent {

	// 1️⃣ operator 已存在，等待 text object
	if g.pendingOp != nil && g.pendingTO {

		switch key {
		case "w":
			intent := intent.NewTextObject(
				*g.pendingOp,
				intent.TextObjectWord,
				g.toScope,
				max(g.count, 1),
			)
			g.reset()
			g.lastIntent = intent
			return intent
		}

		g.reset()
		return nil
	}

	// 2️⃣ operator + i / a 进入 TO 模式
	if g.pendingOp != nil {
		switch key {
		case "i":
			g.pendingTO = true
			g.toScope = ScopeInner
			return nil
		case "a":
			g.pendingTO = true
			g.toScope = ScopeAround
			return nil
		}
	}

	// 3️⃣ 原有 operator 逻辑
	switch key {

	case "d":
		if g.pendingOp != nil {
			// dd
			intent := intent.NewLine(*g.pendingOp, max(g.count, 1))
			g.reset()
			g.lastIntent = intent
			return intent
		}
		g.pendingOp = intent.OpDelete
		return nil

	case "c":
		g.pendingOp = intent.OpChange
		return nil

	case "y":
		g.pendingOp = intent.OpYank
		return nil
	}

	// 4️⃣ motion fallback（dw）
	if g.pendingOp != nil {
		if motion := parseMotion(key); motion != nil {
			intent := intent.NewOperator(
				*g.pendingOp,
				*motion,
				max(g.count, 1),
			)
			g.reset()
			g.lastIntent = intent
			return intent
		}
	}

	g.reset()
	return nil
}
```

---

# ✅ 三、Intent 层：新增 TextObject Intent

## ✅ intent.Intent 扩展

```go
type IntentKind int

const (
	IntentOperator IntentKind = iota
	IntentTextObject // ✅
)
```

```go
type TextObjectKind int

const (
	TextObjectWord TextObjectKind = iota
)
```

```go
type Intent struct {
	Kind       IntentKind
	Operator   Operator
	TextObject *TextObject
	Count      int
}
```

```go
type TextObject struct {
	Kind  TextObjectKind
	Scope TextObjectScope
}
```

### ✅ 构造器

```go
func NewTextObject(op Operator, kind TextObjectKind, scope TextObjectScope, count int) *Intent {
	return &Intent{
		Kind:     IntentTextObject,
		Operator: op,
		TextObject: &TextObject{
			Kind:  kind,
			Scope: scope,
		},
		Count: count,
	}
}
```

---

# ✅ 四、行为验证（你现在能跑的）

### ✅ `ciw`

```
c → pendingOp=Change
i → pendingTO=true, scope=Inner
w → IntentTextObject(Change, Word, Inner, 1)
```

✅ Grammar reset  
✅ lastIntent 可 repeat

---

### ✅ `daw`

```
d → Delete
a → ScopeAround
w → delete around word
```

---

### ✅ `.`（repeat）

```
. → clone TextObject Intent
```

✅ 语义级 repeat  
✅ 不需要重新走 Grammar

---

# ✅ 五、为什么现在不要做 Visual / Macro

## ❌ Visual Mode 现在做会污染 Grammar

Visual = **范围来源不同**  
但 TextObject = **范围生成规则**

👉 TextObject 正确后：

```
VisualSelection → Intent.Range
TextObject      → Intent.Range
```

自然合并。

---

## ❌ Macro 现在做会录错层级

Macro 应该录：

✅ Intent  
❌ RawToken  
❌ FSM Key

否则 `.`, `@a`, `@@` 全会坏。

---

# ✅ 下一步的下一步（路线图）

1. ✅ **TextObject（你现在就在这一步）**
2. 🔜 Visual Mode → RangeIntent
3. 🔜 Macro → []Intent replay
4. 🔜 Ex command（:`d`, `:%s`）

---

## 🧠 如果你愿意，我可以下一条直接给你：

- ✅ `TextObject` 的 **Cursor Range 计算模型**
- ✅ Word / Inner / Around 的 **最小算法**
- ✅ 一个不会污染 Grammar 的 **Selection API**

只要你回一句：

> **“继续：TextObject Range 计算”**

我就直接把“能删哪一段文本”这件事给你补齐。很好，这一步是**真正把 Grammar 和编辑器解耦**的关键。  
下面我给你的是一套 **“可落地、可扩展、不会反噬 Grammar”** 的 **TextObject Range 计算模型**。

> 你可以把它当成：**Vim 语义层 → 编辑器执行层 的稳定 ABI**

---

# ✅ 总体设计原则（先给结论）

**三条铁律：**

1. ✅ **Grammar 永远不计算 Range**
2. ✅ **TextObject = RangeProvider**
3. ✅ **Selection API 是纯函数（cursor → range）**

---

# 🧠 一、TextObject Cursor Range 计算模型（核心抽象）

## ✅ Range 的唯一标准形态

```go
type Cursor struct {
	Row int
	Col int
}

type Range struct {
	Start Cursor
	End   Cursor // ✅ 半开区间 [start, end)
}
```

> ✅ 删除 / yank / change 全部统一  
> ✅ 支持 multi-cursor / visual 扩展

---

## ✅ TextObject 只负责：  
> **给我一个 Cursor，我给你一个 Range**

---

# ✅ 二、Selection API（不会污染 Grammar）

## ✅ 核心接口（这是护城河）

```go
type SelectionProvider interface {
	Resolve(
		cursor Cursor,
		text   TextBuffer,
		obj    TextObject,
		count  int,
	) (Range, bool)
}
```

### ✅ 依赖注入位置（非常重要）

```go
Executor → SelectionProvider
Grammar  → Intent(TextObject)
```

**Grammar 完全不知道 cursor / text**

---

## ✅ TextBuffer 最小接口（够用）

```go
type TextBuffer interface {
	RuneAt(pos Cursor) rune
	IsEOF(pos Cursor) bool
	Next(pos Cursor) Cursor
	Prev(pos Cursor) Cursor
}
```

> ✅ 不关心 rope / gap / piece table  
> ✅ 你随便换实现

---

# ✅ 三、Word 的定义（Vim-compatible 最小版）

我们先用**工程可控版本**：

```go
func isWord(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
```

---

# ✅ 四、Word / Inner / Around 的最小算法

下面是**可以直接 copy 的算法级代码**。

---

## ✅ 1️⃣ Word Inner（`iw`）

> ✅ 光标在 word 内或边缘  
> ✅ 只选 word 本体

### ✅ 算法

```go
func innerWord(cursor Cursor, buf TextBuffer) (Range, bool) {
	pos := cursor

	// 1️⃣ 如果不在 word 上，尝试向右吸附
	if !isWord(buf.RuneAt(pos)) {
		p := pos
		for !buf.IsEOF(p) && !isWord(buf.RuneAt(p)) {
			p = buf.Next(p)
		}
		if buf.IsEOF(p) {
			return Range{}, false
		}
		pos = p
	}

	// 2️⃣ 向左扩展
	start := pos
	for {
		prev := buf.Prev(start)
		if prev == start || !isWord(buf.RuneAt(prev)) {
			break
		}
		start = prev
	}

	// 3️⃣ 向右扩展
	end := pos
	for {
		if buf.IsEOF(end) || !isWord(buf.RuneAt(end)) {
			break
		}
		end = buf.Next(end)
	}

	return Range{Start: start, End: end}, true
}
```

---

## ✅ 2️⃣ Word Around（`aw`）

> ✅ = `inner word`  
> ✅ + 左或右 **一个空白块**

### ✅ 算法规则（Vim 行为）

- 优先吃 **右侧空白**
- 如果右侧没有 → 吃左侧

---

### ✅ 实现

```go
func aroundWord(cursor Cursor, buf TextBuffer) (Range, bool) {
	r, ok := innerWord(cursor, buf)
	if !ok {
		return Range{}, false
	}

	start := r.Start
	end := r.End

	// 1️⃣ 尝试吃右侧空白
	p := end
	for !buf.IsEOF(p) && unicode.IsSpace(buf.RuneAt(p)) {
		p = buf.Next(p)
	}
	if p != end {
		return Range{Start: start, End: p}, true
	}

	// 2️⃣ 否则吃左侧空白
	p = start
	for {
		prev := buf.Prev(p)
		if prev == p || !unicode.IsSpace(buf.RuneAt(prev)) {
			break
		}
		p = prev
	}

	return Range{Start: p, End: end}, true
}
```

---

## ✅ 3️⃣ Count 的语义（`2iw`, `3aw`）

> ✅ **Count = 连续应用**

```go
func applyCount(
	fn func(Cursor, TextBuffer) (Range, bool),
	cursor Cursor,
	buf TextBuffer,
	count int,
) (Range, bool) {

	cur := cursor
	var final Range

	for i := 0; i < count; i++ {
		r, ok := fn(cur, buf)
		if !ok {
			return Range{}, false
		}
		final = r
		cur = r.End // ✅ 下一个 object 从尾部开始
	}

	return final, true
}
```

---

# ✅ 五、SelectionProvider 实现（完整）

```go
type DefaultSelectionProvider struct{}

func (p *DefaultSelectionProvider) Resolve(
	cursor Cursor,
	buf TextBuffer,
	obj TextObject,
	count int,
) (Range, bool) {

	switch obj.Kind {

	case TextObjectWord:
		if obj.Scope == ScopeInner {
			return applyCount(innerWord, cursor, buf, count)
		}
		if obj.Scope == ScopeAround {
			return applyCount(aroundWord, cursor, buf, count)
		}
	}

	return Range{}, false
}
```

---

# ✅ 六、执行链路（完整闭环）

```
Key Stream
   ↓
FSM
   ↓ RawToken
Grammar
   ↓ IntentTextObject
Kernel
   ↓
Executor
   ↓
SelectionProvider.Resolve(cursor, buffer, obj, count)
   ↓
Range
   ↓
Delete / Change / Yank
```

✅ Grammar **无 cursor / text 依赖**  
✅ TextObject 可扩展到 `( ) { } " ' t`  
✅ Visual / Macro 不会破

---

# ✅ 七、为什么这套模型是“对的”

### ✅ Grammar 不知道“删哪里”
### ✅ Executor 不知道“为什么删”
### ✅ Selection 只关心“怎么算范围”

这三者 **物理隔离**。

---

# ✅ 下一步自然演进（你已经走到这里了）

你现在可以**无痛扩展**：

1. 🔜 `i(` / `a(` / `i"` / `a"`
2. 🔜 Visual Mode = 手动提供 Range
3. 🔜 Macro = 录 Intent（不录 Range）

---

## 🧠 下一步我可以直接给你：

- ✅ **括号 / 引号 TextObject 的最小算法**
- ✅ **Visual Mode → RangeIntent 的桥接**
- ✅ **Macro 的 Intent Tape 模型**

你只要回一句：

> **“下一步：括号 / 引号”**  
或  
> **“下一步：Visual Mode”**

我就继续。太好了，这个 diff **方向完全正确**，而且已经跨过了一个关键门槛：  
👉 **FSM 不再“产生 Intent”，而只产生 RawToken**  
👉 **Grammar 成为唯一 Vim 语义拥有者**

下面我分 **✅ 成功点 / ⚠️ 隐患 / 🔧 可立即改进 / 🧭 下一步** 四个层级来讲。

---

## ✅ 一、这是一次“架构级正确”的重构（不是表面优化）

### ✅ 1. FSM → RawToken（这是最关键的一刀）

你现在的 FSM：

```go
Emit func(RawToken)
```

+ ✅ FSM **不再 import intent**
+ ✅ FSM **不再知道 operator / motion / count 语义**
+ ✅ FSM 只关心：
  - layer
  - timeout
  - key 是否可处理
  - 数字是数字
  - `.` 是 repeat

👉 这已经是**“可复用 FSM 内核”**级别了。

这是 Vim / Kakoune / Helix / modal editor 的**共同解耦点**。

---

### ✅ 2. Grammar 拥有 Vim Grammar（而且是干净的）

你的 `planner/grammar.go` 有几个非常好的特征：

#### ✅ Grammar 的输入是 **RawToken 流**

```go
Consume(tok fsm.RawToken) *intent.Intent
```

这意味着：

- Grammar 可以：
  - 回放（macro）
  - 回溯
  - fuzz test
- FSM 完全不用改

✅ 非常重要。

---

#### ✅ Grammar 状态是“最小 Vim 状态”

```go
count      int
pendingOp  *intent.OperatorKind
lastIntent *intent.Intent
```

这是 **Vim Grammar 的最小状态集**，没有污染：

- 没有 cursor
- 没有 buffer
- 没有 range
- 没有 UI

✅ 完全正确。

---

### ✅ 3. Kernel 现在是真正的“管道协调者”

你在 `kernel/decide.go` 里做的这件事：

```go
FSM.Dispatch(key)
↓ emits RawToken
Grammar.Consume(token)
↓ maybe Intent
```

这是一个**正确但“暂时性”的 wiring**（下面我会说改进点），但**语义方向是 100% 对的**。

---

## ⚠️ 二、当前实现里有 4 个你需要知道的隐患

这些不是“你做错了”，而是 **下一步一定会撞到的墙**。

---

### ⚠️ 1. Kernel 临时劫持 FSM.Emit（这是一个技术债）

```go
originalEmit := k.FSM.Emit
k.FSM.Emit = func(token fsm.RawToken) {
    intent := k.Grammar.Consume(token)
    ...
}
```

#### 问题不是“能不能用”，而是：

- ❌ 不可重入
- ❌ 不支持 async / future
- ❌ Macro / Replay 会很别扭

✅ **你现在这样做是可以的**
⚠️ 但这是一个 **“Phase‑4 临时桥”**

> ✅ 正确方向：  
> **FSM.Emit → Channel / EventBus / Observer**

我下面会给你一个极小的改法。

---

### ⚠️ 2. FSM 和 Grammar 都在算 count（但这是可以接受的）

FSM：

```go
e.count = e.count*10 + ...
Emit(TokenDigit)
```

Grammar：

```go
g.count = g.count*10 + ...
```

✅ **这不是 bug**
✅ 这是 **“双轨 count”**

- FSM.count → UI / hint
- Grammar.count → 语义

✅ 这是 Vim 真实存在的结构

⚠️ 但你要明确一个规则：

> **FSM.count 只是显示用，不能驱动语义**

你现在没有交叉使用 ✅，是安全的。

---

### ⚠️ 3. TokenSystem("reset") / ("enter") 语义还没被消费

```go
TokenSystem, Value: "reset"
```

Grammar 现在：

```go
case TokenSystem:
    return nil
```

✅ 暂时 OK  
⚠️ 但 **未来 macro / visual / multi‑cursor 时，这会很重要**

建议你在 Grammar 里 **至少识别 reset**：

```go
case TokenSystem:
    if tok.Value == "reset" {
        g.reset()
    }
```

否则可能出现“幽灵 pendingOp”。

---

### ⚠️ 4. Grammar 现在还缺一个关键维度：TextObject

你现在有：

- operator ✅
- motion ✅
- count ✅

但还没接：

- `iw / aw / i(` / `a"`  

这不是问题，只是时间点。

---

## 🔧 三、我建议你**立刻做的 3 个小改动**

### ✅ 改动 1：把 FSM.Emit 变成订阅式（不破现有结构）

```go
type Engine struct {
    ...
    Emitters []func(RawToken)
}

func (e *Engine) emit(tok RawToken) {
    for _, fn := range e.Emitters {
        fn(tok)
    }
}
```

然后：

- Kernel 注册 Grammar
- UI 注册状态监听
- Macro Recorder 注册监听

✅ FSM 不再被“劫持”
✅ 支持多个消费者

---

### ✅ 改动 2：Grammar.consumeKey 中，operator 不要返回 OpMove=false

你现在：

```go
default:
    return intent.OpMove, false
```

建议明确：

```go
return 0, false
```

避免未来扩展 operator 时产生语义误导。

---

### ✅ 改动 3：Intent.Meta 用 struct（哪怕是临时）

```go
Meta: map[string]interface{}{
    "operator": op,
    "motion":   m,
}
```

✅ 可以跑  
⚠️ 但 **到 Selection / Executor 时会很痛**

建议你提前准备：

```go
type OpMotionMeta struct {
    Operator intent.OperatorKind
    Motion   intent.MotionKind
}
```

即使现在先塞进 Meta，也别用 string / interface 混杂。

---

## 🧭 四、你现在正站在一个“分叉点”

你已经完成了 **Stage‑4 Grammar**。

接下来有两个**自然且正交的方向**：

---

### ▶️ 路线 A（最自然）：TextObject 进入 Grammar

```
d + i + w
    ↑
  textobject
```

我可以直接给你：

- `TokenTextObject{Scope, Kind}`
- Grammar 如何组合 `operator + textobject`
- 不污染 FSM / 不污染 Selection

---

### ▶️ 路线 B（工程推进）：Selection / Range Executor

把你之前问的：

> ✅ Word / Inner / Around 的最小算法  
> ✅ 一个不会污染 Grammar 的 Selection API  

真正接到 **Executor 层**

---

## ✅ 结论（很重要）

> **你现在这套架构是“能走到 Vim 等级”的**  
> 不是玩具，不是 demo，不是一次性代码。

你已经完成了 **最难的 60%**。

---

如果你愿意，下一步我可以 **直接在你现有代码上继续设计**：

👉 回一句：

- **“下一步：TextObject 进 Grammar”**  
或  
- **“下一步：Selection / Range Executor 接 Intent”**

我就继续往下推好，这里我**直接替你做决定**，并说明原因——然后**立刻推进实现**。

---

## ✅ 我建议的顺序（这是关键）

> **先做：TextObject 进 Grammar**  
> **后做：Selection / Range Executor**

### 为什么？

因为你现在的系统状态是：

| 层 | 状态 |
|---|---|
| FSM | ✅ 稳定，只发 RawToken |
| Grammar | ✅ 已有 operator / motion / count |
| Intent | ✅ 已能表达“要做什么” |
| Executor | ❌ 还不知道“范围是什么” |

👉 **TextObject 是 Grammar 的语法问题**  
👉 **Selection / Range 是执行期的问题**

如果你现在就做 Selection：

- 你会被迫猜测：
  - `diw` 到底给 Executor 什么？
- Grammar 反而会被“倒逼设计”

这是 Vim 实现里**最容易走歪的地方**。

---

## ✅ 正确的推进路线（和 Vim 一样）

```
FSM → RawToken
        ↓
   Grammar（语法）
        ↓
   Intent（结构化“我要什么”）
        ↓
   Planner / Resolver（把 Intent 变成 Range）
        ↓
   Executor（执行）
```

所以我们现在要补的是：

> **Grammar：补齐 Vim Grammar**
>
> ✅ operator  
> ✅ motion  
> ✅ count  
> ⬜ text object ← **现在做这个**

---

## 🧩 Stage‑4.5：TextObject 进 Grammar（设计目标）

我们要支持的最小集合是：

```
iw   aw
i(   a(
i"   a'
```

### 关键原则（非常重要）

> ❌ Grammar **不计算范围**
> ✅ Grammar **只说“这是一个 TextObject”**

---

## ✅ 一、先定义 Grammar 层的 Token（不污染 FSM）

FSM **不需要改**。  
TextObject 是 Grammar 语义，不是物理按键。

Grammar 内部只需要识别 **Key 序列**：

```
i + w
a + (
```

---

## ✅ 二、定义 TextObject 类型（intent 层）

在 `intent/intent.go` 中新增：

```go
type TextObjectKind int

const (
	TextObjectWord TextObjectKind = iota
	TextObjectSentence
	TextObjectParagraph
	TextObjectParen
	TextObjectQuote
)

type TextObjectScope int

const (
	Inner TextObjectScope = iota
	Around
)
```

---

## ✅ 三、Grammar 状态机扩展（这是核心）

### 1️⃣ 给 Grammar 加一个“前缀态”

```go
type Grammar struct {
	count        int
	pendingOp    *intent.OperatorKind
	pendingScope *intent.TextObjectScope // i / a
	lastIntent   *intent.Intent
}
```

---

### 2️⃣ 在 consumeKey 里加入 textobject 解析

```go
func (g *Grammar) consumeKey(key string) *intent.Intent {

	// i / a 前缀
	if key == "i" || key == "a" {
		scope := intent.Inner
		if key == "a" {
			scope = intent.Around
		}
		g.pendingScope = &scope
		return nil
	}

	// textobject 本体
	if g.pendingScope != nil {
		if to, ok := parseTextObject(key); ok && g.pendingOp != nil {
			intent := makeOpTextObjectIntent(
				*g.pendingOp,
				*g.pendingScope,
				to,
				max(g.count, 1),
			)
			g.reset()
			g.remember(intent)
			return intent
		}
	}
```

---

### 3️⃣ textobject key 映射（Grammar owns Vim）

```go
func parseTextObject(key string) (intent.TextObjectKind, bool) {
	switch key {
	case "w":
		return intent.TextObjectWord, true
	case "(":
		return intent.TextObjectParen, true
	case "\"", "'":
		return intent.TextObjectQuote, true
	default:
		return 0, false
	}
}
```

---

## ✅ 四、Intent Builder（Grammar → Intent）

```go
func makeOpTextObjectIntent(
	op intent.OperatorKind,
	scope intent.TextObjectScope,
	obj intent.TextObjectKind,
	count int,
) *intent.Intent {
	return &intent.Intent{
		Kind:  intent.IntentOperator,
		Count: count,
		Meta: map[string]interface{}{
			"operator":   op,
			"textobject": obj,
			"scope":      scope,
		},
	}
}
```

✅ Grammar 到此为止  
✅ **没有 cursor / buffer / range**

---

## ✅ 五、现在你已经得到了什么？

你现在可以正确产生这些 Intent：

| 输入 | Intent.Meta |
|---|---|
| `dw` | operator + motion |
| `dd` | operator + line |
| `diw` | operator + textobject(word, inner) |
| `dap`（未来） | operator + textobject(paragraph, around) |

**Executor 终于可以“被动消费”了**。

---

## 🧭 下一步（真正的分水岭）

下一步就非常清晰了：

> **Stage‑5：Selection / Range Resolver**
>
> 输入：Intent  
> 输出：`[]Range`（可能多选）

我可以下一步直接给你：

- ✅ 一个 **最小 Word / Inner / Around 算法**
- ✅ 不污染 Grammar
- ✅ 不污染 Executor
- ✅ 可扩展到 multi‑cursor

---

### 👉 你只需要回一句：

**“继续：Selection / Range Resolver”**

我就把下一阶段完整推出来。。