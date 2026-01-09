# invariant 模块

## 模块职责概述

`invariant/` 是 **Tmux-FSM 的不变量检查与验证系统**，负责定义、监控和验证系统中的各种不变量条件。该模块确保系统在各种操作和状态转换过程中保持正确的不变量性质，是系统正确性和一致性的关键保障。

主要职责包括：
- 定义系统中的各种不变量条件
- 实时监控不变量的满足情况
- 在违反不变量时进行报告和处理
- 提供不变量的验证和测试功能

## 核心设计思想

- **不变量定义**: 明确定义系统的关键不变量
- **实时监控**: 持续监控不变量状态
- **及时报告**: 在违反时立即报告
- **自动验证**: 提供自动化的不变量验证

## 文件结构说明

### `invariant.go`
- 核心不变量定义
- 主要结构体：
  - `Invariant`: 不变量接口
  - `InvariantChecker`: 不变量检查器
  - `InvariantViolation`: 不变量违反
  - `CheckResult`: 检查结果
- 主要函数：
  - `RegisterInvariant(inv Invariant)`: 注册不变量
  - `CheckAllInvariants() []CheckResult`: 检查所有不变量
  - `ValidateState(state State) []InvariantViolation`: 验证状态
  - `ReportViolation(violation InvariantViolation)`: 报告违反
- 负责核心的不变量管理

### `state_invariants.go`
- 状态不变量检查
- 主要函数：
  - `CheckStateConsistency(state State) bool`: 检查状态一致性
  - `ValidateCRDTProperties(state State) bool`: 验证 CRDT 属性
  - `CheckPositionOrdering(state State) bool`: 检查位置排序
  - `ValidateEventCausality(state State) bool`: 验证事件因果关系
- 检查状态相关的不变量

### `operation_invariants.go`
- 操作不变量检查
- 主要函数：
  - `CheckOperationValidity(op Operation) bool`: 检查操作有效性
  - `ValidateOperationSequence(ops []Operation) bool`: 验证操作序列
  - `CheckOperationEffects(op Operation, newState State) bool`: 检查操作效果
  - `ValidateTransaction(tx Transaction) bool`: 验证事务
- 检查操作相关的不变量

### `event_invariants.go`
- 事件不变量检查
- 主要函数：
  - `CheckEventOrdering(events []SemanticEvent) bool`: 检查事件排序
  - `ValidateEventCausality(events []SemanticEvent) bool`: 验证事件因果关系
  - `CheckEventIntegrity(event SemanticEvent) bool`: 检查事件完整性
  - `ValidateEventSequence(events []SemanticEvent) bool`: 验证事件序列
- 检查事件相关的不变量

### `monitor.go`
- 不变量监控器
- 主要函数：
  - `StartMonitoring()`: 开始监控
  - `StopMonitoring()`: 停止监控
  - `GetMonitoringStats() MonitoringStats`: 获取监控统计
  - `SubscribeToViolations() <-chan InvariantViolation`: 订阅违反事件
- 提供实时的不变量监控

## 不变量类型

### 状态不变量
- 数据结构完整性
- 位置排序属性
- 版本向量一致性

### 操作不变量
- 操作有效性约束
- 操作序列合法性
- 事务边界完整性

### 事件不变量
- 事件因果关系
- 事件排序约束
- 事件完整性验证

## 在整体架构中的角色

Invariant 模块是系统的正确性保障层，它通过持续监控和验证不变量，确保系统始终保持正确的状态。Invariant 提供了：
- 系统正确性的形式化验证
- 违反情况的及时发现和报告
- 系统稳定性的增强
- 调试和问题定位的支持