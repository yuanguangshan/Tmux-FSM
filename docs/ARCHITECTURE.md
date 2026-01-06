# tmux-fsm 架构不变量 (System Invariants)

## 概述

本文档定义了 tmux-fsm 系统的核心架构不变量，这些不变量是系统长期可维护性的基础。

## 1. 输入层不变量（Input Sovereignty）

### Invariant 1：FSM 对按键拥有绝对优先裁决权
- 任意一次按键事件 \`key\`
- **FSM 必须在 Intent / legacy 逻辑之前收到它**
- 若 FSM 命中（consume = true）：
  - **该按键不得再流向任何后续系统**

## 2. Keymap 定义不变量（Configuration Authority）

### Invariant 2：keymap.yaml 是 FSM 行为的唯一权威来源
- FSM **不得**：
  - 硬编码任何快捷键
  - 在 Go 代码中推断快捷键语义
- FSM **只能**：
  - 执行 keymap.yaml 中明确定义的行为

## 3. Layer（层级）不变量（State Semantics）

### Invariant 3：FSM 任意时刻只能处于一个 Layer
- FSM.Active ∈ keymap.yaml.states
- 不存在：
  - 多层并存
  - 临时未定义层
- Layer 切换是 **原子操作**

### Invariant 4：Layer 切换必须立即生效
- 一旦 key 触发 layer 变化：
  - **下一次按键必须在新 layer 下解析**

## 4. Action 执行不变量（Execution Semantics）

### Invariant 5：FSM Action 是确定性的
- 给定：
  - 当前 Layer
  - 按键 key
- 结果只能是三种之一：
  1. 执行 action
  2. 切换 layer
  3. 显式拒绝（no-op / reject）

### Invariant 6：FSM 不得"部分执行"
- Action：
  - 要么完整执行
  - 要么完全不执行

## 5. 未命中行为不变量（Rejection Semantics）

### Invariant 7：FSM 未命中 ≠ 错误
- 若当前 layer 未定义该 key：
  - FSM 必须**明确拒绝**
  - 并允许事件继续流向 legacy / weaver

## 6. Reload 行为不变量（Temporal Consistency）

### Invariant 8：Reload 必须是原子重建
Reload 等价于：
1. 丢弃旧 Keymap
2. 重新 Load + Validate
3. 重建 FSM Engine
4. FSM.Active = 初始 layer（通常 NAV）
5. 清空 timeout / sticky
6. 强制刷新 UI

## 7. UI 不变量（Observability）

### Invariant 9：UI 必须真实反映 FSM 状态
- UI 显示的 layer：
  - 必须等于 FSM.Active
- UI 是 **派生状态**
  - 不得反向影响 FSM

## 8. 错误处理不变量（Safety）

### Invariant 10：Keymap 错误必须在启动或 reload 时失败
- keymap.yaml：
  - 非法 → **拒绝加载**
  - FSM 不得运行在非法配置上

## 9. 架构依赖不变量（Dependency Semantics）

### Invariant 11：FSM.Dispatch 必须只有一个入口
- **FSM.Dispatch 只能被 bridge.HandleIntent 调用**
- 任何直接调用 fsm.Dispatch 的代码都是架构错误
- 这确保了单一裁决点的完整性

## 总结

> **FSM 是按键的第一裁决者，
> keymap.yaml 是唯一法源，
> layer 是唯一语境，
> 未定义即拒绝，
> reload 即重生，
> dispatch 有唯一入口。**

这些不变量是整个系统架构的"宪法"，任何违反这些不变量的修改都可能导致系统退化。
