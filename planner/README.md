

# Planner Grammar — Implementation Documentation

> 本模块实现的是 **Stage‑4 Vim Grammar**  
> 它位于 **FSM → Intent** 之间，是一个**纯状态消费器**。

---

# 一、Grammar 的硬边界（非常重要）

## ✅ Grammar 可以做的

- 消费 `fsm.RawToken`
- 维护 **局部语法状态**
- 生成 **`intent.GrammarIntent`**
- 表达 Vim 风格的：
  - count
  - operator
  - motion
  - text object
  - repeat / undo / redo
  - mode switch（有限）

## ❌ Grammar 绝对不做的

- ❌ **不构造 `intent.Intent`**
- ❌ **不绑定 PaneID / ActorID**
- ❌ **不执行任何行为**
- ❌ **不关心 FSM 状态**
- ❌ **不跨 key 保存副作用**

> 🔒 **Promotion 是 Kernel 的唯一职责**

---

# 二、Grammar 的状态模型（真实字段）

```go
type Grammar struct {
    count         int
    pendingOp     *OperatorKind
    pendingMotion *MotionPendingInfo
    textObj       TextObjPending
}
```

---

## 1️⃣ count（数字前缀）

- 由 `TokenDigit` 累积
- 行为：
  ```
  count = count * 10 + digit
  ```
- 默认使用 `max(count, 1)`
- 在 **任何 GrammarIntent 生成后 reset**

---

## 2️⃣ pendingOp（操作符）

```go
d, y, c
```

状态含义：

| 状态 | 意义 |
|----|----|
| nil | 没有操作符 |
| 非 nil | 等待 motion / text object |

支持：

- `dw`
- `d2w`
- `dd`
- `diw`

---

## 3️⃣ pendingMotion（双键 motion）

```go
g{?}
f{c}
t{c}
F{c}
T{c}
```

```go
type MotionPendingInfo struct {
    Kind
    FindDir
    FindTill
}
```

这是 **唯一允许 Grammar “吃两次 key” 的地方**。

---

## 4️⃣ textObj（文本对象）

```go
i{obj}
a{obj}
```

状态枚举：

```go
TOPNone | TOPInner | TOPAround
```

---

# 三、Grammar.Consume：总入口

```go
func (g *Grammar) Consume(tok fsm.RawToken) *GrammarIntent
```

## Token 分类

| TokenKind | 行为 |
|----|----|
| TokenDigit | 累加 count |
| TokenRepeat | 直接产出 Repeat |
| TokenKey | 进入 Vim Grammar |
| TokenSystem | reset |

---

## System Token 的真实行为

```go
reset / exit / enter → g.reset()
```

📌 **System token 永远不会产生 intent**

---

# 四、Key Grammar 的优先级（极其关键）

在 `consumeKey` 中，顺序是**硬编码语义**：

```
1. pendingMotion
2. text object
3. operator
4. text object prefix
5. motion prefix
6. motion
7. mode switch
8. find repeat (; ,)
9. undo / redo
10. reset
```

> ✅ **顺序就是语义**  
> ✅ 没有回溯  
> ✅ 没有 lookahead  

---

# 五、Operator Grammar（操作符）

## parseOperator

```go
d → delete
y → yank
c → change
```

### 行为分支

#### ✅ 连续操作符（dd / yy）

```go
if pendingOp == op → line operation
```

生成：

```go
IntentOperator + MotionLine
```

---

#### ✅ 操作符 + motion

```go
d w
d 2 w
```

生成：

```go
IntentOperator
```

---

#### ✅ 操作符 + text object

```go
d i w
```

由 `consumeTextObject` 完成。

---

# 六、Motion Grammar（移动）

## parseMotion（单键）

支持：

- char: `h l`
- line: `j k`
- word: `w b e ge`
- range: `0 ^ $`
- goto: `G`
- screen line: `H M L`

---

### standalone motion

```go
w
3j
```

生成：

```go
IntentMove
```

---

### operator + motion

```go
dw
d3j
```

生成：

```go
IntentOperator
```

---

# 七、双键 Motion（Find / Goto）

## motion prefix

```go
g f t F T
```

进入 `pendingMotion`

---

## consumePendingMotion

### `gg`

```go
g + g → MotionGoto
```

---

### `f{c} / t{c} / F{c} / T{c}`

生成：

```go
MotionFind{
    Char
    Direction
    Till
}
```

📌 **Intent 类型取决于是否有 pendingOp**

---

# 八、Text Object Grammar

## 进入方式

```go
i{obj}
a{obj}
d i w
```

---

## parseTextObject 支持

| key | Object |
|----|----|
| w | Word |
| ( ) b | Paren |
| [ ] | Bracket |
| { } B | Brace |
| " | QuoteDouble |
| ' | QuoteSingle |
| ` | Backtick |

---

## 生成结构（事实）

```go
Motion{
    Kind: MotionRange
    Range: RangeTextObject{
        Scope: Inner / Around
        Object
    }
}
```

- 无 op → `IntentMove`
- 有 op → `IntentOperator`

---

# 九、Mode Switch Grammar（有限）

## parseModeSwitch

```go
i → insert
v → visual_char
V → visual_line
Esc / C-c → normal
```

### 当前实现状态

- ✅ 生成 Intent
- ❌ **未区分 visual 子模式**
- ❌ 依赖后续 Intent.Meta 扩展

这是**未完成但明确限制的功能**。

---

# 十、Repeat / Undo / Redo

## Repeat

```go
TokenRepeat (.)
→ IntentRepeat
```

---

## Find Repeat

```go
; → IntentRepeatFind
, → IntentRepeatFindReverse
```

---

## Undo / Redo

```go
u   → IntentUndo
C-r → IntentRedo
```

---

# 十一、reset 的真实语义

```go
func (g *Grammar) reset()
```

清空：

- count
- pendingOp
- pendingMotion
- textObj

📌 reset 发生在：

- intent 生成后
- 非法组合
- system token
- 未知 key

---

# 十二、测试覆盖说明（grammar_test.go）

✅ 覆盖的行为：

- 基础移动（hjkl）
- count（3w）
- operator + motion（dw）
- operator + count + motion（d2w）
- gg
- f / t / F / T
- text object（iw / diw）
- repeat（.）

❌ **未覆盖的事实**

- mode switch
- undo / redo
- ; , repeat
- 非法 key reset

---

# 十三、Grammar 的一句话定性

> **Grammar 是一个“纯粹的 Vim 句法折叠器”：**
>
> - 它只知道「按键如何组成语法」
> - 它不关心「这意味着什么」
> - 它不执行任何行为
>
> **Grammar 的输出不是动作，而是“句法完成信号”。**

---

# 十四、与 Kernel 的契约总结

| Grammar | Kernel |
|----|----|
| GrammarIntent | Intent |
| 无上下文 | 绑定 Pane / Actor |
| 状态内聚 | 全局路由 |
| Vim 语法 | 系统语义 |

---

