太好了，你这段 **TxJournal** 已经是「能当内核用」的水平了。  
我不会泛泛夸，也不会推倒重来，只做三件事：

1. **精确确认：你现在这套设计“已经成立”的地方**
2. **指出 5 个“现在不炸、但迟早会炸”的结构性问题**
3. **给出最小、可落地的修正方案（不引入新概念）**

---

## 一、先给结论（很重要）

> ✅ **这是一个“可回滚、可合并、可重放”的真实事务日志**  
> ❌ 但现在它 **在 3 个关键边界上还不够“铁”**：
>
> - 合并（Merge）的历史语义不完整  
> - Undo / Redo 的时间轴在“合并后”会产生歧义  
> - `FuncTx` 允许你偷偷破坏内核不变量

换句话说：

> **现在它是一个“强内核 + 松 API”**

---

## 二、逐条精确点评（按严重性排序）

### 🔴 问题 1：`appendTx` **不会清空 undone（Redo 栈）**

```go
func (j *TxJournal) appendTx(tx Transaction) {
    ...
}
```

而在 `ApplyTxs` 里你有这句：

```go
// 新历史出现 → Redo 失效
j.undone = nil
```

✅ **这是正确的**

❌ 但 `appendTx` 是一个“隐式 Apply”，却**没有做同样的事情**

#### 后果（必现）：
- Undo 一次
- appendTx 一个新 tx
- 再 Redo  
👉 **时间线分叉但你允许重放**

✅ **最小修复：**

```go
func (j *TxJournal) appendTx(tx Transaction) {
    j.undone = nil // ✅ 必须清空 redo
    ...
}
```

---

### 🔴 问题 2：Merge 后，Inverse 的语义是**不明确的**

现在逻辑是：

```go
last := j.applied[n-1].Tx
if last.CanMerge(tx) {
    merged := last.Merge(tx)
    j.applied[n-1] = TxRecord{Tx: merged, ...}
}
```

但你**没有任何约束保证**：

> `merged.Inverse() == tx.Inverse() + last.Inverse()`

这在 **字符输入 / 移动 / Vim send-keys 合并** 中会悄悄出错。

#### 举例（真实）：

- Tx1: 插入 `"a"`
- Tx2: 插入 `"b"`
- 合并成：插入 `"ab"`

👉 但 `Inverse()` 如果还是 `"BSpace"` 一次  
👉 Undo 只删一个字符

✅ **最小设计约束（不用改接口）**

你只需要在 **注释 + code review 级别**确立这一条：

```go
// 规则：
// 如果 Transaction.CanMerge == true
// 那么 Merge 后的 Transaction.Inverse()
// 必须等价于：Inverse(next) 再 Inverse(self)
```

✅ **推荐（但可延后）的实现方式**

在 `FuncTx` 里给一个 helper：

```go
func ChainInverse(a, b Transaction) Transaction {
    return FuncTx{
        apply: func() error {
            _ = b.Inverse().Apply()
            return a.Inverse().Apply()
        },
        inverse: func() Transaction { return NoopTx{} },
        kind: "chained_inverse",
    }
}
```

---

### 🟠 问题 3：`Undo()` / `Redo()` 只能处理 **单事务粒度**

```go
func (j *TxJournal) Undo() error {
    rec := j.applied[len(j.applied)-1]
    ...
}
```

但你系统里已经存在：

- 视觉模式
- change = delete + insert
- search + move
- batch ApplyTxs

这些**在语义上是一个操作**，  
但在日志中会拆成多个 Tx。

#### 后果（用户感知）：
> “我按了一次 undo，怎么只撤了一半？”

✅ **你不需要引入新类型，只需要一个 Tag 约定**

你已经有：

```go
Tags() []string
```

✅ **最小增强方案**

- 约定：`"atomic"` tag
- Undo 时：

```go
func (j *TxJournal) Undo() error {
    if len(j.applied) == 0 {
        return nil
    }

    rec := j.applied[len(j.applied)-1]
    atomic := hasTag(rec.Tx, "atomic")

    for {
        _ = rec.Tx.Inverse().Apply()
        j.applied = j.applied[:len(j.applied)-1]
        j.undone = append(j.undone, rec)

        if !atomic || len(j.applied) == 0 {
            break
        }

        rec = j.applied[len(j.applied)-1]
    }
    return nil
}
```

👉 **这一步会让你的 undo 体验直接上一个档次**

---

### 🟠 问题 4：`FuncTx` 是一把「内核级危险品」

你现在允许：

```go
FuncTx{
    apply: func() error {
        exec.Command("tmux", ...)
        state.RedoStack = nil
        globalVar++
        return nil
    },
}
```

也就是说：

> **FuncTx 可以绕过你所有架构边界**

✅ 你现在用它是“过渡期合理”
❌ 但如果不加约束，它会腐蚀整个系统

✅ **最小约束（不删 FuncTx）**

在代码层面明确 2 条规则：

```go
// 规则 1：FuncTx.apply 只能调用 physical 层
// 规则 2：FuncTx.apply 不得修改任何 journal / state / semantic 数据
```

✅ **技术性防护（推荐）**

- 把 `FuncTx` 移到 `physical/tx_func.go`
- 不允许 import semantic / journal

---

### 🟡 问题 5：`Time` 是记录了，但没被“使用”

```go
Time time.Time
```

现在它只是日志，不参与任何决策。

✅ 这是对的  
但你已经站在一个**很近的门口**：

- 时间分组 merge
- 输入 debounce
- Replay / audit

✅ **建议（不改行为）**

至少加一个 helper：

```go
func (r TxRecord) Age() time.Duration {
    return time.Since(r.Time)
}
```

后续 merge/undo 策略会非常自然。

---

## 三、你现在这套设计的「真实定位」

我非常明确地说一句：

> **这不是 toy journal，也不是 demo undo**  
> **这是一个“编辑器级事务内核”，而且已经比 80% 编辑器干净**

你现在缺的不是新功能，而是**三条铁律**：

1. **Transaction 是唯一副作用入口**
2. **Merge 必须保持 Undo 语义**
3. **Redo 永远只存在于一条时间线**

一旦这三条锁死，  
你这个系统可以非常放心地继续往上长。

---

## 如果你愿意，下一步我可以非常具体地帮你做一件事：

- ✅ 给你 **一份“Transaction 不变量清单”**（像数据库那样）
- ✅ 帮你 **定义 5 个官方 Tx 类型，禁止滥用 FuncTx**
- ✅ 帮你 **设计一个 GroupTx / AtomicTx，但不改 Undo API**
- ✅ 或者直接：**拿你现在的 tx，推演 10 个极端 undo/redo 场景**

你选一个，我继续陪你把这颗内核打磨到“不可动摇”。这一步我需要**先把边界说清楚**，然后我会**立刻开始干活**。

—

## 先说一个关键事实（必须直说）

👉 **要做到第 1️⃣ 和第 2️⃣：**

> ✅「画出 *现在这 3000 行真实的调用图（精确到函数）*」  
> ✅「给出 *可以直接删、100% 不影响行为* 的代码清单」

**我必须看到完整代码**（或至少 main → action → executor → physical 这一整条链）。

**没有源码，我不可能“精确到函数”而不胡编。**  
你这套系统太成熟了，任何臆测都会害你。

✅ **所以正确顺序是：**

- **我现在立刻完成：**
  - ✅ **3️⃣ 设计级别上“彻底消掉 Executor / FuncTx（不改行为）”**
  - ✅ **4️⃣ 给你一个 ≤6 个函数的最小 Semantic API（可直接落地）**

- **你下一步只需要做一件事：**
  - 把代码贴出来（或分文件贴）
  - 我就给你 **① 调用图 + ② 删代码清单（精确到函数名）**

下面我先把 **3️⃣ + 4️⃣** 一次性给你，**不是概念，是可以直接用的设计**。

—

# ✅ 4️⃣ 最小 Semantic API（≤ 6 个函数）

这是**你整个系统的“语义心脏”**，也是你之后能彻底解耦 physical / tx / undo 的关键。

### 🎯 设计目标

- ✅ 不知道 tmux / vim / exec
- ✅ 不产生副作用
- ✅ 只回答一个问题：

> **“这次用户意图，在语义上发生了什么？”**

—

## ✅ 最小 Semantic API（6 个函数，封顶）

```go
package semantic

// 1️⃣ 捕获当前位置（Anchor）
func CaptureCursor() Anchor

// 2️⃣ 捕获一个 motion / range
func CaptureRange(a Anchor, motion Motion) Range

// 3️⃣ 语义删除（不会真的删）
func CaptureDelete(r Range) Fact

// 4️⃣ 语义插入
func CaptureInsert(a Anchor, text string) Fact

// 5️⃣ 语义替换（= delete + insert，但原子）
func CaptureReplace(r Range, text string) Fact

// 6️⃣ 语义移动（cursor / selection）
func CaptureMove(from Anchor, to Anchor) Fact
```

### ✅ 非常重要的约束（请你记住）

- ❌ **Semantic 层不调用任何 exec / tmux / vim**
- ❌ **Semantic 层不产生 Transaction**
- ✅ 只产生 `Fact`

> **Fact = 对“文本世界发生了什么”的陈述**

—

## Fact 的最小形态（示例）

```go
type Fact interface {
	Kind() string
	Inverse() Fact
}
```

你现在已经有类似的东西，只是被埋在 physical 里了。

—

# ✅ 3️⃣ 彻底消掉 Executor / FuncTx（不改行为）

这是关键的一刀，我会非常克制。

—

## ❌ 你现在的问题本质

你现在有三种“副作用入口”：

| 入口 | 状态 |
|-—|-—|
| `Executor.Execute(f Fact)` | ❌ 多余 |
| `FuncTx.apply()` | ❌ 过于自由 |
| `Transaction.Apply()` | ✅ 正确 |

👉 **现实世界只能有一个入口。**

—

## ✅ 铁律（请你直接贴在代码里）

> **任何会 exec.Command / tmux / vim 的代码，只允许存在于：**
>
> ```go
> Transaction.Apply()
> ```

—

## ✅ Executor 怎么“消失”而不改行为？

### 原来（示意）：

```go
Fact → Executor.Execute → physical
```

### 现在（唯一通路）：

```go
Fact → Decide → []Transaction → Journal.ApplyTxs
```

—

## ✅ Action → Transaction 的标准形态

```go
func Decide(f Fact) []Transaction {
	switch f.Kind() {

	case ”delete“:
		return []Transaction{
			VimDeleteTx{Range: f.Range()},
		}

	case ”insert“:
		return []Transaction{
			VimInsertTx{Pos: f.Anchor(), Text: f.Text()},
		}
	}
	return nil
}
```

✅ **Executor 不需要了**  
✅ **每个 Tx 都是物理投影**

—

## ✅ FuncTx 怎么安全地“死掉”？

你现在可能到处是：

```go
FuncTx{
	apply: func() error {
		exec.Command(”tmux“, ...)
		return nil
	},
}
```

### ✅ 替代方案（等价、不改行为）

把它们**全部替换为命名 Tx 类型**：

```go
type TmuxSendKeysTx struct {
	Pane string
	Keys string
}

func (t TmuxSendKeysTx) Apply() error {
	return tmux.SendKeys(t.Pane, t.Keys)
}

func (t TmuxSendKeysTx) Inverse() Transaction {
	return NoopTx{}
}
```

👉 **每一个 FuncTx = 一个真实 Tx 类型**

你不是删功能，你是在**把“匿名副作用”变成“可推理对象”**。

—

## ✅ FuncTx 最终命运（推荐）

- 🚫 不再对外暴露
- ✅ 只在测试里使用
- ✅ 或只允许在 `physical/` 包内部

—

# ✅ 现在你已经拥有的清晰结构

```text
Intent
  ↓
Semantic (Fact)
  ↓
Decide (Fact → []Tx)
  ↓
Journal.ApplyTxs
  ↓
Transaction.Apply
  ↓
Physical (tmux / vim)
```

这条链一旦成立：

- Undo / Redo 会自然正确
- Merge 才有语义基础
- Physical 再乱也污染不到内核

—

# ✅ 接下来我需要你做的一件事（非常具体）

👉 **把代码贴出来（或 repo）**，至少包括：

- main / loop
- action / intent
- executor / physical
- transaction 定义
- semantic / capture 相关

然后我会**严格按你原代码**给你：

### ✅ 1️⃣ 精确到函数的真实调用图  
### ✅ 2️⃣ 一份“可以直接删，不影响行为”的清单（逐条说明）

> **不是建议，是“你现在就可以删”的那种。**

你已经把最难的事做完了，  
剩下的只是**拆箱，而不是重写**。

把代码给我，我们继续。