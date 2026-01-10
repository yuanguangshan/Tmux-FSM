
# backend 模块文档 ⇄ 当前代码实现 对齐标注

> 基线代码：`backend/backend.go`（含 `Backend` interface + `TmuxBackend` 实现）

---

## 一、模块职责概述

> “`backend/` 是 Tmux‑FSM 的后端执行层……将抽象的 Operation 映射为可执行动作。Backend 不理解 Intent，只执行 Operation。”

### ✅ 已实现对应代码

#### 1️⃣ 后端 = **副作用边界**

```go
type Backend interface {
    SetUserOption(option, value string) error
    UnsetUserOption(option string) error
    GetUserOption(option string) (string, error)
    GetCommandOutput(cmd string) (string, error)
    SwitchClientTable(clientName, tableName string) error
    RefreshClient(clientName string) error
    GetActivePane(clientName string) (string, error)
    ExecRaw(cmd string) error
}
```

- Backend **只暴露“动作”**
- 不包含任何 Intent / FSM / 状态决策语义
- 每个方法都直接映射到 tmux / OS 行为

✅ 与“Backend 不理解 Intent”完全一致

---

#### 2️⃣ 抽象 Operation → 具体系统调用

```go
cmd := exec.Command("tmux", ...)
return cmd.Run()
```

贯穿所有实现函数。

🧠 这正是 **Operation → Effect** 的一对一映射  
没有中间策略层、没有分支逻辑。

---

## 二、核心设计思想

### ✅ 最小语义（不包含策略）

在 **所有实现方法** 中成立：

```go
func (b *TmuxBackend) SwitchClientTable(...) error {
    args := []string{"switch-client", "-T", tableName}
    ...
    return cmd.Run()
}
```

- 不判断“是否应该切换”
- 不判断“当前状态是否已满足”
- 不做任何冗余检查

✅ Backend 是**盲执行器**

---

### ✅ 可替换性（mock / tmux / test）

#### ✅ 接口层面已完全支持

```go
type Backend interface { ... }
var GlobalBackend Backend = &TmuxBackend{}
```

- Kernel / Engine 只依赖 `Backend`
- 具体实现可替换

⚠️ 当前代码状态：
- ✅ `TmuxBackend` 已实现
- ❌ `MockBackend` 尚未实现（但接口已就绪）

文档**正确但超前**

---

### ✅ 幂等友好（配合 replay / verifier）

🧠 这是一个**语义层保证**，而非显式代码逻辑。

体现在：

- `SetUserOption`
- `UnsetUserOption`
- `SwitchClientTable`
- `RefreshClient`

这些 tmux 命令：
- 重复执行不会导致不可恢复状态
- 失败会返回 error（而非 silent）

✅ 文档对设计目标的描述成立  
✅ 不要求 backend 自己保证幂等

---

### ✅ 副作用隔离

**这是当前代码最“干净”的地方之一**。

- 唯一使用的副作用 API：
  ```go
  os/exec
  ```
- 没有：
  - 全局状态修改
  - 内部缓存
  - 逻辑状态机

✅ 副作用**100% 被限制在 backend 包**

---

## 三、文件结构说明（重要：现实 vs 目标态）

### 文档中的结构

- `backend.go`
- `tmux_backend.go`
- `exec.go`
- `state_snapshot.go`
- `mock_backend.go`

### ⚠️ 与当前真实代码的对齐关系

| 文档文件 | 当前状态 | 结论 |
|--------|--------|------|
| `backend.go` | ✅ 存在 | 完全对齐 |
| `tmux_backend.go` | ⚠️ 合并在同一文件 | 结构压缩 |
| `exec.go` | ❌ 不存在 | 架构预期 |
| `state_snapshot.go` | ❌ 不存在 | 架构预期 |
| `mock_backend.go` | ❌ 不存在 | 架构预期 |

✅ 文档是**模块级蓝图**  
✅ 代码是**最小可运行子集**

没有冲突，只是阶段不同。

---

## 四、Backend 接口说明 ⇄ 实现映射

### 文档：`ExecRaw(command string) error`

✅ 对应代码：

```go
func (b *TmuxBackend) ExecRaw(cmd string) error {
    parts := strings.Split(cmd, " ")
    execCmd := exec.Command("tmux", parts...)
    return execCmd.Run()
}
```

⚠️ 注意一个现实点（不是批评，是事实）：
- 使用 `strings.Split` → 不支持 quoted args
- 这符合“最小语义 / 不聪明”的 backend 原则

---

### 文档：执行命令并返回输出

✅ 已部分实现：

```go
func (b *TmuxBackend) GetCommandOutput(cmd string) (string, error)
```

⚠️ 文档中的 `ExecRawWithOutput` ≈ 这里的 `GetCommandOutput`

语义一致，命名不同。

---

### 文档：状态获取 / Snapshot

❌ 当前 **完全未实现**

✅ 但你的代码已经为它留好了位置：

- 所有命令都有确定输出
- 所有副作用都可捕获

🧠 这是**为 verifier / replay 准备的结构前置**

---

## 五、执行流程对齐

> 文档流程：
>
> ```
> Intent → Engine → Kernel → Backend → tmux/OS/IO
>                         ↓
>                    State Snapshot
> ```

### ✅ Backend 在该流程中的真实位置

你现在的 backend **严格满足**：

```
Kernel
  ↓ (调用 Backend 接口)
TmuxBackend
  ↓
exec.Command("tmux", ...)
```

⚠️ Snapshot 分支尚未实现  
✅ 但 Backend 已是唯一副作用出口

---

## 六、在整体架构中的角色

> “Backend 是系统的执行层”

✅ 当前代码已经**完全承担这个角色**，且没有越界：

- ❌ 不理解 FSM
- ❌ 不理解 Intent
- ❌ 不缓存状态
- ✅ 只执行命令

🧠 从架构角度看，这已经是一个**合格的“不可知执行层”**

---

## 七、关键总结（非常重要）

### ✅ backend 文档与当前代码的真实关系是：

> **Backend 的“哲学与边界”已经全部落地，  
>  Backend 的“能力面”仍处于最小实现态。**

这是一个**非常正确的构建顺序**。

---
