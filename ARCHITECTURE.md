# Tmux-FSM 架构规范（权威版）

## 一、架构目标

> **将「按键输入」与「行为执行」彻底解耦，**
> **通过 FSM → Grammar → Kernel → Intent → Transaction 的单向管道，**
> **构建一个可推理、可重放、不可腐化的编辑系统。**

---

## 二、最终三层架构（权威版）

### 🧠 Layer 1：Semantic（不可变世界）

**保留目录**
```
semantic/
crdt/
replay/
wal/
index/
policy/
selection/    (Ephemeral)
verifier/
```

**职责**
- 世界发生了什么
- 因果顺序
- 可验证重放
- CRDT 合并
- Policy 校验

**禁止**
- tmux
- cursor
- editor buffer
- motion 细节

---

### 🧩 Layer 2：Decide（解释语义）

**保留 / 重构**
```
planner/        ✅（Grammar → Intent）
intent/         ✅
kernel/decide   ✅
```

**职责**

```
Intent
↓
Facts
↓
ResolvedFacts（抽象）
```

- ✅ 可以有 Meta
- ✅ 可以组合多个 Fact
- ❌ 不执行

---

### ⚙️ Layer 3：Project（现实世界）

**唯一执行入口**
```
weaver/
  adapter/
  core/
```

**职责**

```
ResolvedFact
↓
tmux / vim / shell
↓
Reality
```

- ✅ Snapshot 校验
- ✅ Projection Verify
- ✅ Undo / Redo

**唯一允许 exec.Command 的地方**

```
weaver/adapter/tmux_*
```

---

### 🚨 绝对规则（写在 README 顶部）

> **任何物理副作用只能发生在 Project 层**

---

## 三、语义层冻结清单

### 1. Semantic 的定义必须满足这 5 条

**semantic.BaseFact 只能表达：**

| 项目 | 允许 |
|----|----|
| 世界发生了什么 | ✅ |
| 抽象目标（anchor / range） | ✅ |
| 文本内容 | ✅ |
| 可逆关系（Inverse） | ✅ |
| 执行方式 | ❌ |
| tmux / vim / shell | ❌ |
| key / motion | ❌ |
| cursor mode | ❌ |

---

### 2. Fact.Kind 冻结

允许的 **最终 Fact 集合**：

```go
insert
delete
replace
move
set_selection   // Ephemeral
```

- ❌ 不再新增 Fact.Kind
- ❌ 禁止在 Projection 中 switch 新 Kind
- ✅ 新行为只能通过 **Meta + Decide**

---

### 3. Semantic 不得 import：

```text
os/exec
tmux
editor
resolver
kernel
weaver
```

---

### 4. Replay 必须是纯函数

允许：

```go
Replay(initial, events, filter) TextState
```

禁止：
- tmux
- time.Now()
- global state

---

### 5. Undo = Replay，而不是 Exec Inverse

允许：

```go
UndoCheckout(
  targetEvent,
  events,
  actor,
  initial,
)
```

禁止：
- exec inverse send-keys
- shell undo
- vim undo delegation

---

## 四、禁止 import 列表

以下导入在指定层是严格禁止的：

### Semantic 层禁止：
- `os/exec`
- `tmux-fsm/weaver`
- `tmux-fsm/editor`
- `tmux-fsm/resolver`

### Decide 层禁止：
- `os/exec`
- `tmux-fsm/weaver/adapter`

### Project 层（weaver）：
- 仅允许在此层执行物理操作
- 仅允许在此层调用 `exec.Command`

---

## 五、架构宪法（最高准则）

1. **按键不执行行为**：按键只表达意图，不直接产生效果
2. **FSM 只是输入设备**：FSM 只产生 token，永远不理解语义
3. **Grammar 拥有语义**：Vim 语义只存在于 Grammar 中，不执行、不裁决
4. **Kernel 是唯一权威**：所有决策、提升与裁决，只能发生在 Kernel
5. **Intent 是契约，不是实现**：Intent 与后端无关，可记录、可重放
6. **Builder 只做语义翻译**：Builder 冻结映射关系，不读状态、不执行
7. **Resolver 是技术债**：Resolver 只用于兼容，严禁新增功能
8. **所有编辑必须是 Transaction**：绕过 Transaction 的编辑一律视为 bug
9. **UI 永远不是权威**：UI 是派生结果，不能驱动语义或逻辑
10. **怀疑不确定性**：如果逻辑不知道该放哪一层，说明设计已经出问题