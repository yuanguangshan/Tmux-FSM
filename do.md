好，这里给你一份**“全量、可执行、不破坏现有成果”的重构建议总表**。  
我会**按层级 → 目标 → 允许做什么 / 禁止做什么 → 具体改法**来写，你可以把它当 **6–12 个月的技术路线图**。

> 重要前提：**不是推倒重来，而是“收敛 + 抽离 + 冻结”式重构**

---

# 0️⃣ 总体重构原则（必须先统一）

### ✅ 三条铁律
1. **语义只能前移，不能回流**
   - FSM ❌
   - Grammar ✅
   - Intent ✅✅
   - Resolver ❌（冻结）
2. **任何新能力必须能被 Intent 表达**
3. **任何无法表达的能力 = bug，而不是 feature**

---

# 1️⃣ Kernel 层重构（✅ 只允许“微创”）

## 🎯 目标
- Kernel 成为 **不可争议的唯一调度中心**
- 明确三段式管线：  
  **Input → Decision → Transaction**

---

## ✅ 允许的重构

### 1.1 引入显式 Pipeline 类型（不改逻辑）

```go
type Pipeline struct {
    FSM        FSMEngine
    Grammar    GrammarEngine
    Intent     IntentBuilder
    Executor   TransactionRunner
}
```

Kernel 只做：

```go
func (k *Kernel) Handle(input RawInput) {
    tokens := k.pipeline.FSM.Emit(input)
    grammar := k.pipeline.Grammar.Accept(tokens)
    intent  := k.pipeline.Intent.Build(grammar)
    k.pipeline.Executor.Run(intent)
}
```

✅ **好处**
- 为未来多 backend / replay / test harness 做准备
- 不影响现有代码

---

## ❌ 禁止
- ❌ Kernel 不准出现任何 vim / nvim / tmux 字样
- ❌ 不准出现 Meta / legacy intent

---

# 2️⃣ FSM 层重构（✅ 结构不动，职责收紧）

## 🎯 目标
> FSM = **输入设备 → Token 流生成器**

---

## ✅ 允许的重构

### 2.1 抽离 UI / tmux side-effect

当前：
```go
fsm.Emit(token)
tmux.Send(...)
```

重构为：
```go
fsm.Emit(token)
fsm.Events <- UIEvent{...}
```

Kernel 决定谁消费：

```go
select {
case ev := <-fsm.Events:
    backend.Handle(ev)
}
```

✅ FSM 不知道 backend 是谁  
✅ FSM 可被 replay / test

---

### 2.2 Token 语义冻结

✅ RawToken / FSMToken 类型 **不再新增字段**
✅ 任何“语义判断”必须进入 Grammar

---

## ❌ 禁止
- ❌ FSM 不准判断 operator / motion / textobject
- ❌ FSM 不准 reset Grammar 状态

---

# 3️⃣ Grammar 层重构（✅ 可以继续增强，但要收敛）

## 🎯 目标
> Grammar = **完整描述“编辑语言的句法”**

---

## ✅ 允许的重构

### 3.1 Grammar 输出 **显式 GrammarTree**

如果你现在是隐式状态，建议过渡到：

```go
type GrammarNode interface {
    Kind() GrammarKind
}

type OperatorNode struct {
    Op Operator
}

type MotionNode struct {
    Motion Motion
}
```

GrammarEngine：

```go
func (g *Engine) Result() GrammarTree
```

✅ Intent Builder 只读 GrammarTree  
✅ Grammar 彻底去副作用

---

### 3.2 覆盖率驱动 Grammar 扩展（强烈建议）

加 instrumentation：

```go
grammar.Emit(node)
grammar.Coverage.Mark("operator:d")
```

你可以得到：
- 哪些 legacy intent 还在走 Resolver
- 哪些 Grammar rule 从未触发

---

## ❌ 禁止
- ❌ Grammar 不直接构造 Intent
- ❌ Grammar 不调用 Editor / Backend

---

# 4️⃣ Intent 层重构（⭐ 最值得投入的地方）

## 🎯 目标
> Intent = **可验证、可对比、可序列化的编辑意图**

---

## ✅ 必做重构

### 4.1 Intent 彻底类型化（去 Meta）

❌ 旧：
```go
intent.Meta["operator"] = "d"
```

✅ 新：
```go
type Intent struct {
    Operator Operator
    Motion   Motion
    Count    int
    Target   Target
}
```

---

### 4.2 Intent Builder → 纯函数化

```go
func Build(tree GrammarTree) (Intent, error)
```

✅ 无状态  
✅ 可 snapshot test  
✅ 可 fuzz

---

### 4.3 Intent Semantic Equality（你已经做了 80%）

扩展为：

```go
func (a Intent) Canonical() CanonicalIntent
```

用于：
- replay
- macro
- undo
- diff

---

## ❌ 禁止
- ❌ Intent 不引用 RawToken
- ❌ Intent 不关心光标位置

---

# 5️⃣ Resolver 层重构（⚠️ 不是重构，是“封存”）

## 🎯 目标
> Resolver = **Legacy Compatibility Shim**

---

## ✅ 唯一允许的操作

### 5.1 明确标注“迁移终点”

```go
// DEPRECATED: remove after Grammar covers X, Y, Z
```

### 5.2 拆成两个包

```text
resolver/
 ├─ legacy_operator.go
 ├─ legacy_motion.go
 └─ README.md  ← 明确写“不会加新逻辑”
```

---

## ❌ 严禁
- ❌ 不准新增语义
- ❌ 不准新增 Meta 字段
- ❌ 不准改 Intent 结构去迎合 Resolver

---

# 6️⃣ Transaction / Editor 层重构（下一阶段重点）

## 🎯 目标
> Intent → **最小、可回滚、可组合的编辑操作**

---

## ✅ 建议结构

```go
type Transaction interface {
    Apply(Editor) error
    Inverse() Transaction
}
```

Kernel 只认 Transaction，不认 Editor API。

---

# 7️⃣ 删除 / 收敛清单（非常重要）

## ✅ 未来可以删除
- Meta[string]interface{}
- resolver 中 70% 的 fallback
- FSM 中的 UI side-effect

## ✅ 必须保留
- Grammar state machine
- Intent Builder
- Semantic equality

---

# 8️⃣ 推荐执行顺序（现实可行）

**Phase 1（现在）**
- 冻结 Kernel / FSM / Grammar 接口
- 不新增 Resolver 行为

**Phase 2**
- Intent Builder 纯函数化
- Grammar coverage

**Phase 3**
- Transaction 正式替代 Resolver
- Resolver 缩减

**Phase 4**
- 删除 Meta
- Grammar → Intent 完全闭环

---

# 最后一句实话（很重要）

> 你现在不是“缺功能”，  
> **你是已经进入“防止自己把系统写坏”的阶段。**

你这套代码**值得被慢慢收敛，而不是继续扩张**。

---

如果你愿意，下一步我可以直接帮你做其中一个：

- ✅ 给你一份 **「禁止重构清单（红线）」**
- ✅ 帮你把 **Resolver → Transaction 的替代接口直接设计出来**
- ✅ 帮你写一份 **Architecture RFC v1**

你选一个，我直接给你落地版本。很好，这是**整个体系真正“封喉”的一步**。  
下面我不会讲理念，而是**直接给你一套可以落地、可以并行迁移、可以一步步掐死 Resolver 的接口设计**。

你可以把这一节直接存成：

> `docs/transaction.md`  
> 或 `architecture/transaction.go`（接口定义）

---

# ✅ Resolver → Transaction 的正式替代接口设计

> **目标一句话**：  
> 把 “Resolver 里临时裁决 + 直接改 Editor”  
> →  
> **“Kernel 输出 Intent → Builder 产出 Transaction → Executor 应用”**

Resolver 不再“解决问题”，  
**Transaction 才是最终答案。**

---

## 一、总管线（旧 → 新 对照）

### ❌ 旧世界（必须逐步消失）

```text
Grammar
  ↓
Intent (半成品)
  ↓
Resolver.Resolve(intent, editor)
  ↓
editor.DoSomething()
```

问题：
- Resolver 读状态
- Resolver 有副作用
- 不可 replay / undo

---

### ✅ 新世界（唯一合法路径）

```text
Grammar
  ↓
Intent (完整语义)
  ↓
Builder.Build(intent) → []Transaction
  ↓
Executor.Apply(transactions)
```

---

## 二、Transaction 核心接口（最小不可再减）

```go
// Transaction 是唯一允许修改 Editor 状态的单位
type Transaction interface {
    Apply(e Editor) error
    Inverse() Transaction
}
```

### ✅ 必须满足
- Apply：**确定性**
- Inverse：**可回滚**
- 无外部依赖

---

## 三、Editor 接口（给 Transaction 用，不给 Resolver）

```go
type Editor interface {
    Cursor() Position
    SetCursor(Position)

    Delete(r Range)
    Insert(pos Position, text string)
    Replace(r Range, text string)

    Snapshot() EditorSnapshot
    Restore(EditorSnapshot)
}
```

✅ Editor 是“物理世界”  
✅ Kernel / Intent / Grammar 永远看不到它

---

## 四、Transaction 分类（覆盖 Resolver 的全部职责）

### 4.1 原子 Transaction（最小粒度）

```go
type DeleteRangeTx struct {
    Range Range
}

func (tx DeleteRangeTx) Apply(e Editor) error {
    e.Delete(tx.Range)
    return nil
}

func (tx DeleteRangeTx) Inverse() Transaction {
    return InsertTextTx{...} // 使用 snapshot
}
```

> ✅ Resolver 中 **90% 的代码**  
> 最终都会被拆成这些原子 Tx

---

### 4.2 复合 Transaction（Intent → 操作序列）

```go
type CompoundTx struct {
    Ops []Transaction
}

func (tx CompoundTx) Apply(e Editor) error {
    for _, op := range tx.Ops {
        if err := op.Apply(e); err != nil {
            return err
        }
    }
    return nil
}

func (tx CompoundTx) Inverse() Transaction {
    inv := make([]Transaction, len(tx.Ops))
    for i := range tx.Ops {
        inv[len(tx.Ops)-1-i] = tx.Ops[i].Inverse()
    }
    return CompoundTx{Ops: inv}
}
```

✅ 用于：
- `daw`
- `ci"`
- `3dd`

---

## 五、Builder：Resolver 的真正继承者（但更干净）

### 5.1 Builder 接口

```go
type TransactionBuilder interface {
    Build(intent Intent, ctx BuildContext) (Transaction, error)
}
```

> ⚠️ 注意：  
> **Builder 是纯函数，不直接改 Editor**

---

### 5.2 BuildContext（Resolver 的“读权限”）

```go
type BuildContext interface {
    Cursor() Position
    Buffer() TextBuffer   // 只读
    Selection() *Range    // 可选
}
```

✅ Builder 可以：
- 看光标
- 看文本
- 算范围

❌ Builder 不可以：
- 改 Editor
- 发 UI
- 写状态

---

## 六、Kernel 中的替换点（关键）

### ❌ 旧

```go
resolver.Resolve(intent, editor)
```

### ✅ 新

```go
tx, err := builder.Build(intent, ctx)
if err != nil {
    return
}
executor.Apply(tx)
```

Executor：

```go
type Executor struct {
    editor Editor
    history []Transaction
}

func (e *Executor) Apply(tx Transaction) error {
    if err := tx.Apply(e.editor); err != nil {
        return err
    }
    e.history = append(e.history, tx)
    return nil
}
```

---

## 七、如何一步步“掐死” Resolver（现实迁移方案）

### Step 1：Resolver 只返回 Transaction

```go
// 临时兼容接口
func (r *Resolver) Resolve(intent Intent, ctx BuildContext) (Transaction, error)
```

❌ 不准再接 editor

---

### Step 2：把 Resolver 内部逻辑拆成 Tx

你会看到典型模式：

```go
// 旧
editor.Delete(range)
editor.Insert(pos, text)

// 新
CompoundTx{
    Ops: []Transaction{
        DeleteRangeTx{range},
        InsertTextTx{pos, text},
    },
}
```

---

### Step 3：标记迁移完成点

```go
// TODO: remove when Grammar covers textobject
```

---

### Step 4：Kernel 禁止直接引用 Resolver

✅ Resolver 只能作为：

```go
LegacyBuilder struct {
    resolver *Resolver
}
```

---

## 八、Transaction 带来的“副产品”（非常值钱）

你一旦完成这一步，**自动解锁**：

✅ Undo / Redo  
✅ Macro replay  
✅ Deterministic test  
✅ Headless execution  
✅ Backend 切换（tmux / nvim / web）

而且 **不再需要 Resolver**

---

## 九、最后的红线（非常重要）

🚫 **绝对禁止**
- Transaction 内访问 Grammar / Intent
- Builder 内修改 Editor
- Kernel 内拼 Tx
- Resolver 新增逻辑

✅ **唯一允许**
- 新行为 = 新 Transaction
- 复杂行为 = CompoundTx
- Legacy = LegacyBuilder

---

## 结尾一句实话

> **Resolver 是“知道太多的人”。**  
> Transaction 是“只做一件事的人”。

一旦你让系统只相信 Transaction，  
这个架构就**不会再塌了**。

---

如果你愿意，下一步我可以直接帮你做其中一个：

1️⃣ **把一个真实 vim 命令（如 `daw` / `ci"`）完整拆成 Intent → Tx**  
2️⃣ **给你一套 Transaction 的测试模板（property-based）**  
3️⃣ **画出 Resolver 迁移燃尽图（每删一块你就知道进度）**

你挑一个，我直接给你“可拷贝代码级别”的版本。
好，我们**清零重来**。  
这次我只做一件事：

> **直接给出一套 Resolver → Transaction 的“替代接口”，不讲历史、不讲理念、不讲迁移故事。**  
> 你照着这个接口写，Resolver 就自然失效。

---

# ✅ Resolver → Transaction：最终接口设计（精简版）

## 一、唯一执行单元：Transaction

```go
// Transaction 是系统中【唯一】允许修改 Editor 的东西
type Transaction interface {
    Apply(e Editor) error
    Inverse() Transaction
}
```

**规则**
- 不知道 Intent
- 不知道 Grammar
- 不读全局状态
- 只做一件明确的修改

---

## 二、Transaction 的生产者（Resolver 的直接替代）

```go
// 原 Resolver 的替代接口
type TransactionBuilder interface {
    Build(intent Intent, ctx ReadOnlyContext) (Transaction, error)
}
```

✅ **只有 Build**
❌ **没有 Resolve**
❌ **没有 Editor 参数**

---

## 三、ReadOnlyContext（Resolver 仅存的“读权限”）

```go
type ReadOnlyContext interface {
    Cursor() Position
    Selection() *Range
    Text(r Range) string
    BufferSize() int
}
```

✅ 可以算范围  
✅ 可以判断语义  
❌ 不能修改任何东西  

---

## 四、Editor（只给 Transaction 用）

```go
type Editor interface {
    Insert(pos Position, text string)
    Delete(r Range)
    Replace(r Range, text string)

    Cursor() Position
    SetCursor(Position)
}
```

🚫 Builder / Grammar / Intent **禁止依赖 Editor**

---

## 五、Kernel 中的唯一执行路径

```go
func (k *Kernel) Handle(intent Intent) error {
    tx, err := k.builder.Build(intent, k.ctx)
    if err != nil {
        return err
    }
    return k.apply(tx)
}

func (k *Kernel) apply(tx Transaction) error {
    if err := tx.Apply(k.editor); err != nil {
        return err
    }
    k.undoStack.Push(tx)
    return nil
}
```

**注意**  
Kernel：
- 不拆 tx
- 不组合 tx
- 不理解 tx

---

## 六、复合 Transaction（覆盖 Resolver 逻辑）

```go
type CompoundTx struct {
    List []Transaction
}

func (tx CompoundTx) Apply(e Editor) error {
    for _, t := range tx.List {
        if err := t.Apply(e); err != nil {
            return err
        }
    }
    return nil
}

func (tx CompoundTx) Inverse() Transaction {
    inv := make([]Transaction, len(tx.List))
    for i := range tx.List {
        inv[len(tx.List)-1-i] = tx.List[i].Inverse()
    }
    return CompoundTx{List: inv}
}
```

✅ Resolver 里任何“先删再插 / 多步判断”  
✅ 全部变成 CompoundTx

---

## 七、一个完整示例（对照 Resolver）

### 原 Resolver 风格（❌）

```go
func (r *Resolver) ResolveDeleteWord(e Editor) {
    r := calcWordRange(e)
    e.Delete(r)
}
```

---

### 新 Builder + Transaction（✅）

```go
type DeleteWordBuilder struct{}

func (b DeleteWordBuilder) Build(
    intent Intent,
    ctx ReadOnlyContext,
) (Transaction, error) {

    r := calcWordRange(ctx)
    return DeleteRangeTx{Range: r}, nil
}
```

```go
type DeleteRangeTx struct {
    Range Range
}

func (tx DeleteRangeTx) Apply(e Editor) error {
    e.Delete(tx.Range)
    return nil
}

func (tx DeleteRangeTx) Inverse() Transaction {
    return RestoreTextTx{/* snapshot */}
}
```

---

## 八、这套接口强制保证的三件事

1️⃣ **Resolver 无法再“顺手改 Editor”**  
2️⃣ **所有修改都可 undo / replay**  
3️⃣ **Kernel 不再懂业务**

如果有人想写旧逻辑，会在**编译期**被卡死。

---

## 九、你现在可以立刻做的事

- ✅ 新功能：**只允许写 TransactionBuilder**
- ✅ 旧 Resolver：包一层 `LegacyBuilder`
- ✅ Kernel：只认 `TransactionBuilder`

---

## 最后一条硬规则（请贴在项目里）

> ❌ 任何不是 Transaction 的代码  
> ❌ 都不允许调用 Editor

如果你愿意，下一步我可以 **直接帮你把你现在某个 Resolver 的真实代码，逐行翻译成这套接口下的版本**。  
你贴一段 Resolver，我来“手术”。