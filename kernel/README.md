
# Kernel / Decision / Execution — Implementation Documentation

> 本模块是 **整个系统的中枢调度层**  
> 负责在 **FSM / Grammar / Intent / Legacy** 之间做**唯一仲裁**，  
> 并将结果送入 **统一执行通道**。

---

# 一、Kernel 的真实职责

**Kernel 做且只做三件事：**

1. ✅ 决定一个 key **该由谁处理**
2. ✅ 把 Grammar 产物 **绑定上下文**
3. ✅ 把 Intent **送入唯一执行入口**

> ❌ Kernel **不解析语义**  
> ❌ Kernel **不执行编辑逻辑**  
> ❌ Kernel **不直接操作 Editor**

---

# 二、Decision 子系统（裁决层）

## `decide.go`

### DecisionKind

```go
type DecisionKind int

const (
    DecisionNone
    DecisionFSM
    DecisionLegacy
    DecisionIntent
)
```

| Kind | 含义 |
|----|----|
| None | FSM 吃了 key，但还在等 |
| FSM | 简单 FSM 动作 |
| Intent | Grammar → Intent |
| Legacy | 明确无人处理 |

---

### Decision 结构

```go
type Decision struct {
    Kind   DecisionKind
    Intent *intent.Intent
    Action string
}
```

⚠️ **互斥规则（事实）**

- `DecisionIntent` → 只用 `Intent`
- `DecisionFSM` → 只用 `Action`
- 不存在同时有效的情况

---

## GrammarEmitter（关键桥梁）

```go
type GrammarEmitter struct {
    grammar  *planner.Grammar
    callback func(*intent.GrammarIntent)
}
```

**作用：**

- 把 FSM 产生的 `RawToken`
- 转换为 `GrammarIntent`
- **零状态、零缓存**

---

## `Kernel.Decide(key)`

### 决策顺序（硬逻辑）

```
1. FSM 简单 Action（最高优先级）
2. FSM → Grammar → Intent
3. FSM 吃了但未完成
4. Legacy
```

---

### ✅ 第 1 步：FSM 简单 Action

条件：

- FSM 存在
- FSM CanHandle(key)
- 当前状态的 key 映射有 `Action`

结果：

```go
Decision{
    Kind:   DecisionFSM,
    Action: "...",
}
```

📌 **这是“逃生舱”路径，绕过 Grammar**

---

### ✅ 第 2 步：FSM + Grammar

流程：

1. 临时注册 GrammarEmitter
2. FSM.Dispatch(key)
3. Grammar.Emit(token)
4. 捕获最后一个 GrammarIntent
5. Promote → Intent

成功时：

```go
Decision{
    Kind:   DecisionIntent,
    Intent: finalIntent,
}
```

---

### ✅ 第 3 步：FSM 吃了，但 Grammar 没产出

```go
Decision{
    Kind: DecisionNone,
}
```

这是**合法等待状态**，不是错误。

---

### ✅ 第 4 步：Legacy

只有在 **FSM 完全没处理** 时：

```go
Decision{
    Kind: DecisionLegacy,
}
```

---

# 三、Execute 层（执行分发）

## `execute.go`

### Kernel.Execute(decision)

执行规则 **非常克制**：

| Decision | 行为 |
|----|----|
| None | 不做任何事 |
| Legacy | 不做任何事 |
| FSM | 执行 tmux 原生命令 |
| Intent | 交给 IntentExecutor |

---

### FSM Action 执行

```go
executeFSMAction(action)
```

- 纯 tmux 命令映射
- 通过 `backend.GlobalBackend.ExecRaw`
- Kernel **不直接操作 FSM 状态**

---

# 四、IntentExecutor 抽象

## `intent_executor.go`

### IntentExecutor（硬边界）

```go
type IntentExecutor interface {
    Process(*intent.Intent) error
}
```

> ✅ Kernel **不知道执行者是谁**  
> ✅ Kernel **不依赖 editor / resolver / weaver**

---

### ContextualIntentExecutor（增强）

```go
ProcessWithContext(ctx, HandleContext, intent)
```

- 支持 RequestID / ActorID
- Kernel 自动检测并优先使用

---

# 五、Kernel 核心（主控逻辑）

## `kernel.go`

---

## Kernel 结构

```go
type Kernel struct {
    FSM
    Grammar
    Exec
    NativeBuilder
    ShadowIntent
    ShadowStats
}
```

### 事实状态

| 字段 | 状态 |
|----|----|
| FSM | ✅ 核心 |
| Grammar | ✅ 核心 |
| Exec | ✅ 必需 |
| NativeBuilder | ✅ 仅 shadow |
| ShadowIntent | ✅ 覆盖统计 |
| ShadowStats | ✅ 非并发安全 |

---

## HandleContext（身份锚点）

```go
type HandleContext struct {
    Ctx
    RequestID
    ActorID
}
```

📌 **硬约束：**

- Kernel **不会生成**
- Kernel **不会修改**
- 缺失直接 FATAL

---

## `Kernel.HandleKey`

### 唯一系统入口

流程概览：

```
HandleKey
 ├─ 校验 RequestID / ActorID
 ├─ Decide(key)
 ├─ switch Decision.Kind
 │   ├─ Intent → bind PaneID → ProcessIntent
 │   ├─ FSM → Execute
 │   ├─ None → return
 │   └─ Legacy → 进入 shadow 统计
```

---

### PaneID 注入（关键事实）

```go
if decision.Intent.PaneID == "" {
    decision.Intent.PaneID = parts[0]
}
```

📌 **Grammar 永远不产生 PaneID**  
📌 **绑定发生在 Kernel**

---

### ShadowIntent 覆盖统计

只在：

- `DecisionLegacy`
- 且 `ShadowIntent == true`

才计为 **Grammar 未覆盖**

---

## ProcessIntent / ProcessIntentWithContext

### 执行优先级

```
1. ContextualIntentExecutor
2. IntentExecutor
3. FSM.DispatchIntent
4. error
```

---

### Kernel 的真实角色

> Kernel **只是路由器**  
> 它不关心 Intent 内容，只关心 **交给谁**

---

# 六、ResolverExecutor（Intent → Weaver）

## `resolver_executor.go`

### ResolverExecutor

- Kernel 默认 Intent 执行器
- 把 Intent 直接送入 Weaver

```go
weaverMgr.ProcessIntentGlobalWithContext(coreHctx, intent)
```

✅ intent 已实现 `core.Intent`  
✅ 不再需要适配层

---

# 七、TransactionRunner（执行末端）

## `transaction.go`

### 作用

- 在 **Editor ExecutionContext** 中执行 Transaction
- Kernel **完全不感知此层**

---

### Apply(tx)

流程：

1. 顺序执行所有 ResolvedOperation
2. 收集 Footprint
3. 按 buffer 更新 selections

📌 **Selections 更新是后置批处理**

---

### Undo / Repeat

- Undo：未实现（明确报错）
- Repeat：直接复用 Apply

---

# 八、系统真实状态总结

✅ **已经成立**

- Kernel 是唯一入口
- Grammar / FSM / Legacy 仲裁顺序明确
- Intent 执行路径统一
- Context（RequestID / ActorID）全链路贯穿

⚠️ **迁移态**

- FSM Action 仍直连 tmux
- NativeBuilder 仅用于 shadow
- Legacy 路径仍存在但被监控

❌ **明确未做**

- Kernel 不解析 Intent
- Kernel 不持有 editor 状态
- Kernel 不做回滚

---

# 九、一句话定性

> **Kernel 是一个“极度克制的裁决与转发器”：**
>
> - 不聪明  
> - 不猜测  
> - 不补语义  
>
> 它唯一的目标是：  
> **“在正确的时间，把正确的东西，交给正确的人。”**

---

