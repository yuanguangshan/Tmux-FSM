# kernel 模块

## 模块职责概述

`kernel/` 是 **Tmux-FSM 的唯一权威仲裁层**，负责将 Intent 转换为具体的执行操作。该模块是系统的**唯一决策者**，根据当前状态和策略决定如何处理传入的意图，确保所有决策的权威性和一致性。

主要职责包括：
- 作为**唯一权威**处理所有 Intent
- 解析和验证传入的 Intent
- 根据当前上下文做出执行决策
- 协调 Engine、Backend 等子系统的协作
- 确保操作的原子性和一致性
- 处理并发和冲突情况

## 核心设计思想

- **唯一权威**: Kernel 是系统中唯一的决策者，所有裁决只能发生在 Kernel
- **Intent 驱动**: Kernel 只处理 Intent，不读取物理状态
- **策略驱动**: 根据策略和上下文做出智能决策
- **状态无关**: Kernel 不直接读取 tmux/编辑器状态，只基于历史事实决策
- **可验证性**: 所有决策过程可追溯和验证

## 文件结构说明

### `kernel.go`
- Kernel 核心结构体和接口定义
- 主要结构体：
  - `Kernel`: 核心决策引擎
  - `ExecutionContext`: 执行上下文
  - `DecisionResult`: 决策结果
  - `HandleContext`: 处理上下文
- 主要函数：
  - `Process(intent.Intent) error`: 处理传入的意图
  - `Execute(operation.Operation) error`: 执行具体操作
  - `Validate(intent.Intent) error`: 验证意图的有效性
  - `ResolveConflicts([]intent.Intent) []intent.Intent`: 解决意图冲突
- 负责核心的决策逻辑和执行调度
- 集成 Grammar 解析和 Intent 提升机制

### `decide.go`
- 决策引擎实现
- 主要结构体：
  - `GrammarEmitter`: 语法解析发射器
  - `Decision`: 决策结果
- 主要函数：
  - `Decide(key string) *Decision`: 决策处理
  - `HandleKey(key string) error`: 处理按键输入
- 负责 FSM → Grammar → Intent 的解析流程
- 实现 Grammar → Intent 的提升机制

### `scheduler.go`
- 任务调度器
- 主要函数：
  - `Schedule(intent.Intent)`: 调度意图执行
  - `QueueSize() int`: 获取队列大小
  - `Flush()`: 清空并执行队列中的所有操作
- 管理意图的执行顺序和并发控制

### `policy_enforcer.go`
- 策略执行器
- 主要函数：
  - `Enforce(intent.Intent) error`: 执行策略检查
  - `CheckPermissions(actorID string, intent.Intent) bool`: 检查权限
  - `ApplyConstraints(intent.Intent) intent.Intent`: 应用约束条件
- 确保所有操作符合系统策略

## 决策流程

1. **接收意图**: 从 Engine 或 FSM 层接收 Intent
2. **验证意图**: 检查意图的有效性和安全性
3. **上下文分析**: 分析当前系统状态和上下文
4. **决策制定**: 根据策略和上下文制定执行决策
5. **操作生成**: 将意图转换为具体的操作序列
6. **执行协调**: 协调 Backend 和其他子系统执行操作
7. **结果反馈**: 返回执行结果和状态更新

## 在整体架构中的角色

Kernel 是整个系统的"大脑"，它接收来自上层的意图，经过分析和决策后，将这些意图转换为具体的执行操作。Kernel 确保所有操作都符合系统策略，并且能够正确处理并发和冲突情况，维护系统的整体一致性和可靠性。