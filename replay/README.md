# replay 模块

## 模块职责概述

`replay/` 是 **Tmux-FSM 的执行历史记录与重放能力**，负责将历史事件重放以重建特定时间点的系统状态。该模块关注的问题是："系统是如何一步一步走到当前状态的？"和"如果从同样的输入重新开始，是否还能得到同样的结果？"，是系统可验证性、可调试性与可回溯性的基础。

主要职责包括：
- 将历史事件序列重放为系统状态
- 支持任意时间点的状态重建
- 提供状态差异比较功能
- 验证操作的正确性和一致性
- 与 Verifier 配合提供执行历史验证

## 核心设计思想

- **完全可回放**: 任何状态都可以从事件历史中重建
- **精确恢复**: 支持精确到特定事件的状态恢复
- **验证驱动**: 通过回放验证操作的正确性
- **高效重建**: 优化回放性能，支持快速状态重建

## 文件结构说明

### `replay.go`
- 核心回放逻辑实现
- 主要结构体：
  - `TextState`: 文本状态
  - `ReplayResult`: 回放结果
- 主要函数：
  - `Replay(initial TextState, events []SemanticEvent, filter EventFilter) TextState`: 执行回放
  - `ReplayRange(from, to EventID, events []SemanticEvent) TextState`: 范围回放
  - `ValidateReplay(events []SemanticEvent, expected TextState) bool`: 验证回放结果
- 负责核心的事件重放逻辑

### `state_builder.go`
- 状态构建器
- 主要函数：
  - `BuildStateFromEvents(events []SemanticEvent) TextState`: 从事件构建状态
  - `ApplyEvent(state TextState, event SemanticEvent) TextState`: 将单个事件应用到状态
  - `BuildIncrementalState(events []SemanticEvent, checkpoint EventID, base TextState) TextState`: 增量状态构建
- 管理状态的逐步构建过程

### `snapshot_manager.go`
- 快照管理器
- 主要函数：
  - `CreateSnapshot(state TextState, at EventID) Snapshot`: 创建状态快照
  - `LoadSnapshot(id string) (TextState, bool)`: 加载状态快照
  - `GetCheckpointEvents() []EventID`: 获取检查点事件
  - `CleanupOldSnapshots(keepCount int)`: 清理旧快照
- 管理状态快照以优化回放性能

### `validator.go`
- 回放验证器
- 主要函数：
  - `ValidateSequence(events []SemanticEvent) error`: 验证事件序列的有效性
  - `CheckConsistency(events []SemanticEvent) bool`: 检查一致性
  - `DetectAnomalies(events []SemanticEvent) []Anomaly`: 检测异常
- 确保回放过程的正确性

## 回放特性

### 增量回放
- 支持从任意检查点开始的增量回放
- 避免重复处理早期事件
- 提高大历史数据的回放效率

### 选择性回放
- 支持按参与者过滤的回放
- 支持按时间范围的回放
- 支持按事件类型的回放

### 验证回放
- 支持与预期状态的对比验证
- 提供详细的差异报告
- 支持自动修复检测到的问题

## 在整体架构中的角色

Replay 模块是系统可验证性的核心组件，它确保了所有操作的历史都可以被准确重建和验证。通过回放功能，系统能够：
- 调试和分析历史问题
- 验证操作的正确性
- 支持精确的状态恢复
- 提供操作审计功能