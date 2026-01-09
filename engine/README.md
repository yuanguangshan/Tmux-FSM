# engine 模块

## 模块职责概述

`engine/` 是 **Tmux-FSM 的核心调度与系统粘合层**，负责将高层的 **Intent** 组织、调度并转化为一次 **可执行、可验证、可回放的事务化执行过程**。

如果说：
- `intent/` 定义了「想做什么」
- `kernel/` 决定了「应该怎么做」  
- `backend/` 负责「真正去做」

那么：
> **Engine = 把"想法"安全、可靠地变成一次现实执行的中枢系统**

## Engine 的核心职责

- 接收来自 `cmd/`、`fsm/`、`ui/` 的高层请求
- 将多个 Intent 组织为 **Transaction**
- 协调以下子系统：
  - Kernel（决策 / 推导）
  - Backend（副作用执行）
  - Replay / Verifier（历史与一致性）
- 维护运行时上下文（Context / Session）

## 核心设计思想

- **Intent-first**: Engine 不直接操作状态，一切变化都源自 Intent
- **唯一权威仲裁**: 所有决策、提升与裁决，只能发生在 Engine（架构戒律 4）
- **事务化（Transaction）**: 每一次执行都有明确边界，要么成功、要么可回滚 / 重放
- **可回放（Replayable）**: 所有执行路径都可以被完整重建
- **可验证（Verifiable）**: 执行结果可以被独立系统检查

## 文件结构说明

### `engine.go`
- Engine 核心结构体与接口定义
- 主要函数：
  - `Apply(event crdt.SemanticEvent) error`: 应用事件到状态
  - `Replay(upTo crdt.EventID) replay.TextState`: 重放至指定事件
  - `Snapshot() *Snapshot`: 创建状态快照
  - `Append(event crdt.SemanticEvent) crdt.EventID`: 添加事件到日志
  - `AllocatePosition(after, before *crdt.PositionID) crdt.PositionID`: 分配新位置
  - `ComparePosition(a, b crdt.PositionID) int`: 比较位置
  - `ApplySelection(actor crdt.ActorID, fact selection.SetSelectionFact)`: 应用选择区域变更
  - `CheckPolicy(event crdt.SemanticEvent) error`: 检查策略
- 是其他模块使用 Engine 的唯一入口点
- 负责协调 Kernel、Backend、Replay、Verifier 等子系统

### `concrete_engine.go`
- ConcreteEngine 是 Engine 接口的具体实现
- 主要函数：
  - `NewConcreteEngine() *ConcreteEngine`: 创建新的引擎实例
  - `ComputeMotion(m *intent.Motion) (editor.MotionRange, error)`: 计算运动产生的范围
  - `computeTextObject(textObj *intent.TextObject) (editor.MotionRange, error)`: 计算文本对象的范围
  - `computeWord(count int) (editor.MotionRange, error)`: 计算单词移动的范围
  - `computeLine(count int) (editor.MotionRange, error)`: 计算行移动的范围
  - `computeChar(count int) (editor.MotionRange, error)`: 计算字符移动的范围
  - `computeGoto(count int) (editor.MotionRange, error)`: 计算跳转的范围
  - `computeFindMotion(find *intent.FindMotion, count int) (editor.MotionRange, error)`: 计算查找运动的范围
  - `MoveCursor(r editor.MotionRange) error`: 移动光标到指定范围
  - `DeleteRange(r editor.MotionRange) error`: 删除指定范围的内容
  - `YankRange(r editor.MotionRange) error`: 复制指定范围的内容
  - `ChangeRange(r editor.MotionRange) error`: 修改指定范围的内容

## 在整体架构中的角色

Engine 位于整个系统的中心位置，连接各个子系统：
```
Intent → Engine → Kernel → Backend
           ↓
       Replay/Verifier
```

Engine 确保所有操作都是事务化的、可追溯的，并且能够与其他模块协同工作。