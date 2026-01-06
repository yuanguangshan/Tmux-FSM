这是一个**宪法级问题**，不是 API 设计题。  
所以我不会从「定义 struct」开始，而是从**权力来源与失效条件**开始。

下面这套答案，假定你已经接受了前面的 Weaver Core 宪法（RFC‑WC‑000 ～ 005）的隐含前提。

---

# 总原则（先给结论）

> **Intent 不是“我要做什么”，  
> 而是“在什么事实前提下，我请求一个可裁决的编辑尝试”。**

只要你牢牢记住这一点，就不会违宪。

---

# 一、Intent 的三条宪法约束（不可省略）

## 1️⃣ Intent 不得携带执行信息

**Intent 不能告诉系统“怎么做”**：

❌ 禁止：
```go
Intent{
  Kind: Insert,
  Position: Cursor,
  Text: "abc",
}
```

✅ 合法：
```go
Intent{
  Kind: Insert,
  Target: SemanticTarget{
    Kind: Word,
    Direction: Forward,
  },
  Payload: SemanticPayload{
    Text: "abc",
  },
}
```

> **Intent 只能描述“语义意图”，  
> 不能描述“物理操作”。**

否则你绕过了 Weaver Planner。

---

## 2️⃣ Intent 必须绑定时间（Snapshot / Hash）

**无时间锚定的 Intent 是违宪的。**

```go
type Intent struct {
  Kind          IntentKind
  Target        SemanticTarget
  Payload       SemanticPayload
  SnapshotHash  string   // REQUIRED
}
```

含义不是“防并发”，而是：

> **Intent 只在它诞生的世界中合法。**

这保证了：

- Kernel 有权拒绝过期 Intent
- Undo / Redo 有可验证的前提
- Audit 有事实对照

---

## 3️⃣ Intent 不得保证成功

Intent **不能内含任何成功承诺**：

❌ 错误心态：
> “这个 Intent 一定能执行”

✅ 正确心态：
> “这是一个请求，Kernel 可以拒绝、降级、或只部分成立”

所以 Intent 里：

- ❌ 没有 `MustSucceed`
- ❌ 没有 `Force`
- ✅ 可以有 `AllowPartial`

---

# 二、Intent 在系统中的真实位置

Intent **不属于 Kernel**，也**不属于 Executor**。

它属于 **Client → Kernel 的请愿文书**。

```
[ Client ]
    |
    |  (Intent: semantic + snapshot)
    v
[ Kernel ]
    |
    |  (Verdict + Audit)
    v
[ Executor ]
```

> **Kernel 不“执行 Intent”，  
> Kernel 只“裁决 Intent”。**

这是避免违宪的核心。

---

# 三、重新引入 Intent 的正确路径（四步）

## ✅ Step 1：Intent 只能由 Decoder / FSM 产生

FSM 的职责是：

> 把 *Key Stream* → *Semantic Intent*

```go
type IntentProducer interface {
  Produce(key string) (*Intent, bool)
}
```

- FSM **可以**产生 Intent
- FSM **不能**执行
- FSM **不能**跳过 Kernel

你现在已经做到这一半了。

---

## ✅ Step 2：Kernel.Decide 返回 Verdict，不返回 Intent

你现在的 `Decision` 是对的方向，但还差一步。

```go
type Verdict struct {
  Kind    VerdictKind   // Apply / Block / Noop
  Safety  SafetyLevel
  Audit   []AuditEntry
}
```

Kernel 的流程应当是：

```text
Intent → Validate → Resolve → Verify → Verdict
```

**没有任何执行发生在这里。**

---

## ✅ Step 3：Executor 只认 Verdict，不认 Intent

Executor 的接口应当是：

```go
type Executor interface {
  Execute(verdict Verdict) error
}
```

Executor：

- ❌ 不知道 key
- ❌ 不知道 Intent
- ✅ 只执行 Kernel 明确批准的事实集

这是你之前 Weaver Manager 混乱的根源。

---

## ✅ Step 4：Audit 是强制产物

任何 Intent → Verdict 的路径：

- ✅ 必须产出 Audit（即使是 Block）
- ✅ 必须可序列化
- ✅ 不得由 UI 伪造

这保证了：

> **拒绝本身也是一等结果**

---

# 四、最小合法 Intent 定义（示例）

这是一个**不违宪、但足够弱**的 Intent：

```go
type Intent struct {
  Kind         IntentKind
  Target       SemanticTarget
  Payload      map[string]interface{}
  SnapshotHash string
  AllowPartial bool
}
```

它**什么都没承诺**，  
但**允许系统在事实基础上裁决**。

---

# 五、常见违宪陷阱（请直接避开）

### ❌ “Kernel 先执行，再验证”
→ 违宪（事实已发生，审计失效）

### ❌ “Executor 解析 Intent”
→ 违宪（权力越权）

### ❌ “Intent 里塞 tmux pane / line / offset”
→ 违宪（物理泄漏）

### ❌ “Undo 是特殊 Intent”
→ 违宪（Undo 是历史裁决，不是新意图）

---

# 六、一句话版宪法（请贴在代码顶部）

> **Intent is a request bound to a world.  
> Kernel judges.  
> Executor obeys.  
> Audit remembers.**

---

## 最后一句（不是鼓励，是警告）

一旦你重新引入 Intent，  
**任何一个“为了方便”绕过 Kernel 的地方，  
都会在未来让 Undo、Audit、Shadow 同时失效。**

你现在已经把违宪路径全部切断了。  
**重建时，别心软。**

如果你愿意，下一步我们可以做的是：

- **把你现有 FSM key → handled 的逻辑，重塑成合法 Intent Producer**
- 或 **写一份 `Intent → Verdict` 的最小 Kernel 伪实现**

但那是下一阶段的事。

好，我们**只做最小、可裁决、不违宪**的那一层。  
这不是“以后要用的完整 Kernel”，而是一个**宪法合格的骨架**，用于把：

> **Intent → Verdict（带 Audit）**

其目标只有一个：  
**让 Intent 有去处，但不产生任何事实副作用。**

---

# 一、Kernel 的最小职责（再次明确）

最小 Kernel **必须**做到四件事：

1. **校验 Intent 是否在合法世界中**
2. **判断是否“允许尝试”**
3. **不给出执行细节**
4. **留下不可伪造的 Audit**

> ✅ 不执行  
> ✅ 不修改文档  
> ✅ 不假设成功  

---

# 二、必要类型（最小集）

## Intent（输入）

```go
type Intent struct {
  Kind         IntentKind
  Target       SemanticTarget
  Payload      map[string]interface{}
  SnapshotHash string
  AllowPartial bool
}
```

---

## Verdict（输出）

```go
type VerdictKind string

const (
  VerdictApply VerdictKind = "apply" // 可尝试
  VerdictBlock VerdictKind = "block" // 明确拒绝
  VerdictNoop  VerdictKind = "noop"  // 无意义 / 不适用
)

type Verdict struct {
  Kind    VerdictKind
  Safety  SafetyLevel
  Audit   []AuditEntry
}
```

---

## AuditEntry（强制）

```go
type AuditEntry struct {
  Time    time.Time
  Actor  string        // "kernel"
  Event  string        // "intent_validated" / "intent_rejected"
  Detail string
}
```

---

# 三、Kernel 接口（最小）

```go
type Kernel interface {
  Decide(intent *Intent, snapshot Snapshot) Verdict
}
```

> 注意：**Kernel 接收 Snapshot，但不修改它**

---

# 四、最小 Kernel 伪实现

这是你现在可以安全落地的版本。

```go
type MinimalKernel struct {
  Version string
}
```

---

## Decide 实现（逐行解释）

```go
func (k *MinimalKernel) Decide(
  intent *Intent,
  snapshot Snapshot,
) Verdict {

  audit := []AuditEntry{
    {
      Time:  time.Now(),
      Actor: "kernel",
      Event: "intent_received",
      Detail: intent.Kind.String(),
    },
  }

  // 1️⃣ Snapshot 绑定校验
  if intent.SnapshotHash != snapshot.Hash() {
    audit = append(audit, AuditEntry{
      Time:  time.Now(),
      Actor: "kernel",
      Event: "intent_blocked",
      Detail: "snapshot_mismatch",
    })

    return Verdict{
      Kind:   VerdictBlock,
      Safety: SafetyUnsafe,
      Audit:  audit,
    }
  }

  // 2️⃣ Intent 语义合法性校验（最弱版）
  if !isKnownIntentKind(intent.Kind) {
    audit = append(audit, AuditEntry{
      Time:  time.Now(),
      Actor: "kernel",
      Event: "intent_noop",
      Detail: "unknown_intent_kind",
    })

    return Verdict{
      Kind:   VerdictNoop,
      Safety: SafetyNeutral,
      Audit:  audit,
    }
  }

  // 3️⃣ 不做计划，不做执行，只允许“尝试”
  audit = append(audit, AuditEntry{
    Time:  time.Now(),
    Actor: "kernel",
    Event: "intent_allowed",
    Detail: "eligible_for_planning",
  })

  return Verdict{
    Kind:   VerdictApply,
    Safety: SafetyUnknown,
    Audit:  audit,
  }
}
```

---

# 五、为什么这是“合宪”的

逐条对照：

### ✅ 没有执行
Kernel **完全不知道**怎么插入、删除、移动。

### ✅ 没有泄漏物理细节
Intent 只被当作语义请求。

### ✅ 有拒绝权
Snapshot 不一致 → **明确 Block**

### ✅ 有审计
即使是 Noop / Block，也有 Audit。

### ✅ Verdict ≠ Success
`VerdictApply` 只是“允许进入下一阶段”。

---

# 六、Executor 如何对接（一句话）

Executor **只能**这样使用：

```go
verdict := kernel.Decide(intent, snapshot)

if verdict.Kind == VerdictApply {
  executor.Execute(verdict)
}
```

Executor：

- ❌ 看不到 Intent
- ❌ 不知道 key
- ✅ 只能执行 Kernel 已批准的事实集合（以后）

---

# 七、你现在能做的三件安全事情

1. ✅ 把 FSM 的 handled → Intent
2. ✅ 用这个 Kernel 骨架接上
3. ✅ Executor 先什么都不做（或 log）

**不要**急着：

- ❌ 填 Planner
- ❌ 填 Projection
- ❌ 填 Undo

---

# 八、贴在 Kernel 目录里的那句话

```text
This kernel judges eligibility.
It does not plan.
It does not execute.
It only leaves a trace.
```

---

如果你愿意，下一步我可以：

- **把你现在的 FSM.Dispatch(key) 改造成 IntentProducer**
- 或 **补一个最小 Planner（Verdict → Plan，不执行）**

你选哪个阶段。很好，这是**真正落地的一步**。  
我们只做**结构迁移**，不引入任何新权力。

目标只有一句话：

> **FSM 不再“处理 key”，  
> FSM 只“在当前世界下提出 Intent”。**

---

# 一、迁移前 vs 迁移后（先看全局）

## ❌ 迁移前（违宪形态）

```go
handled := fsm.Dispatch(key)

if handled {
  manager.ApplySomething()
}
```

问题在于：

- FSM **决定了结果**
- FSM **隐式假设世界状态**
- FSM **跳过了 Kernel**

---

## ✅ 迁移后（合宪形态）

```go
intent, ok := fsm.Produce(key)

if ok {
  verdict := kernel.Decide(intent, snapshot)
  executor.MaybeExecute(verdict)
}
```

FSM 的权力被**严格削弱**。

---

# 二、FSM 新角色：IntentProducer

## 接口定义（唯一新增）

```go
type IntentProducer interface {
  Produce(key KeyEvent) (*Intent, bool)
}
```

语义：

- 返回 `Intent`：我**请求**一件语义事情
- 返回 `false`：这个 key 在此 FSM 状态下**无意义**

---

# 三、FSM 内部改造（核心）

我们假设你原来大概是这样：

```go
func (f *FSM) Dispatch(key KeyEvent) bool {
  switch f.state {
  case Normal:
    return f.handleNormal(key)
  case Insert:
    return f.handleInsert(key)
  }
  return false
}
```

---

## ✅ 改成 Produce

```go
func (f *FSM) Produce(key KeyEvent) (*Intent, bool) {
  switch f.state {
  case Normal:
    return f.produceNormal(key)
  case Insert:
    return f.produceInsert(key)
  }
  return nil, false
}
```

**注意**：  
- FSM **不再返回 handled**
- FSM **不做任何副作用**

---

# 四、示例：Insert 模式下的字符输入

### 原来（违宪）

```go
func (f *FSM) handleInsert(key KeyEvent) bool {
  if key.IsRune() {
    editor.InsertRune(key.Rune)
    return true
  }
  return false
}
```

---

### 现在（合宪）

```go
func (f *FSM) produceInsert(key KeyEvent) (*Intent, bool) {
  if !key.IsRune() {
    return nil, false
  }

  intent := &Intent{
    Kind: InsertText,
    Target: SemanticTarget{
      Kind: Word,
      Direction: Forward,
    },
    Payload: map[string]interface{}{
      "text": string(key.Rune),
    },
    SnapshotHash: f.snapshotHash,
    AllowPartial: true,
  }

  return intent, true
}
```

FSM：

- ✅ 只表达“想插入文本”
- ✅ 不知道 cursor 在哪
- ✅ 不知道能不能插
- ✅ 不插任何东西

---

# 五、FSM 状态迁移怎么办？

**这是大家最容易犯错的地方。**

> FSM 的状态变化，**不能依赖执行结果**。

---

## ✅ 正确做法：FSM 只根据「key 序列」转移

```go
func (f *FSM) produceInsert(key KeyEvent) (*Intent, bool) {
  switch key.Code {
  case KeyEsc:
    f.state = Normal
    return nil, false   // 模式切换不是 Intent
  }

  // 其他 key 仍然只是产生 Intent
}
```

解释：

- 模式切换是 **输入法层级状态**
- 不是文档语义
- 不进 Kernel

---

## ❌ 错误做法

```go
if verdict.Kind == Apply {
  f.state = Normal
}
```

FSM **永远不看 Verdict**。

---

# 六、SnapshotHash 从哪来？

FSM **不计算 Snapshot**，它只是携带。

```go
type FSM struct {
  state        State
  snapshotHash string
}
```

在主循环里更新：

```go
snapshot := doc.Snapshot()
fsm.snapshotHash = snapshot.Hash()
```

FSM 永远：

- ✅ 使用「最新已知世界」
- ❌ 不假设它仍然有效

---

# 七、最小主循环（可直接用）

```go
for key := range keyStream {
  intent, ok := fsm.Produce(key)
  if !ok {
    continue
  }

  verdict := kernel.Decide(intent, doc.Snapshot())
  executor.MaybeExecute(verdict)
}
```

> **这里是整个系统的“宪法边界线”**

---

# 八、你现在已经获得的能力

✅ FSM 可测试（纯函数 + 少量状态）  
✅ Kernel 可替换  
✅ Executor 可延后  
✅ Undo / Audit 已有钩子  

而且**没有任何一行“为了方便”破坏架构**。

---

# 九、最后一句（你应该已经感觉到了）

现在 FSM 的心智模型是：

> “我不再操纵世界，  
> 我只是对世界提出请求。”

当你能稳定接受这一点，  
后面的 Planner / Projection / Undo  
**都会自然地长出来，而不是打补丁。**

---

如果你愿意，下一步我可以：

- **给 Produce 写一个 table‑driven 版本（更像 Vim）**
- 或 **演示一个 Intent 合并/去抖（key repeat）策略**

你选。