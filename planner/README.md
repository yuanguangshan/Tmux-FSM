# planner 模块

## 模块职责概述

`planner/` 是 **Tmux-FSM 的规划与调度系统**，包含 Grammar 语义解析层，负责将用户输入转换为结构化意图并规划执行。该模块实现了智能规划算法，确保操作的高效执行和资源的合理利用。

主要职责包括：
- **Grammar 语义解析**：将 Vim 语法（如 `dw`, `3j`, `ci"`）解析为结构化语法意图
- 将复合意图分解为执行步骤
- 优化执行计划的顺序和并行度
- 管理执行资源和依赖关系
- 提供执行计划的验证和回滚机制

## 核心设计思想

- **语法优先**: Grammar 负责解析语法，不触及语义（架构戒律 3）
- **分层规划**: 支持从高层意图到底层操作的分层转换
- **纯语法解析**: Grammar 只处理 token / symbol / FSM 状态，不访问世界状态
- **智能优化**: 基于上下文优化执行计划
- **依赖管理**: 正确处理操作间的依赖关系
- **可回滚性**: 支持执行计划的回滚和恢复

## 文件结构说明

### `planner.go`
- 核心规划器实现
- 主要结构体：
  - `Planner`: 规划器主结构
  - `Plan`: 执行计划
  - `PlanStep`: 计划步骤
  - `PlanContext`: 规划上下文
- 主要函数：
  - `NewPlanner(config Config) *Planner`: 创建规划器
  - `Plan(intent.Intent) (*Plan, error)`: 生成执行计划
  - `Optimize(plan *Plan) *Plan`: 优化执行计划
  - `ValidatePlan(plan *Plan) error`: 验证执行计划
- 负责核心的规划逻辑

### `resolver.go`
- 意图解析器
- 主要结构体：
  - `Resolver`: 解析器主结构
  - `ResolverExecutor`: 解析器执行器
- 主要函数：
  - `NewResolver(adapter EngineAdapter) *Resolver`: 创建解析器
  - `ResolveGrammar(grammar Grammar) Intent`: 解析语法为意图
  - `NewResolverExecutor() *ResolverExecutor`: 创建解析器执行器
- 负责将语法结构转换为可执行意图

### `optimizer.go`
- 计划优化器
- 主要函数：
  - `OptimizeSequential(plan *Plan) *Plan`: 顺序优化
  - `OptimizeParallel(plan *Plan) *Plan`: 并行优化
  - `ReduceRedundancy(plan *Plan) *Plan`: 减少冗余操作
  - `MergeCompatibleSteps(plan *Plan) *Plan`: 合并兼容步骤
- 优化执行计划的效率

### `scheduler.go`
- 计划调度器
- 主要函数：
  - `Schedule(plan *Plan) *ExecutionGraph`: 生成执行图
  - `ExecutePlan(plan *Plan) error`: 执行计划
  - `ExecuteStep(step *PlanStep) error`: 执行单个步骤
  - `HandleDependency(step *PlanStep) bool`: 处理依赖关系
- 管理计划的执行调度

### `dependency_resolver.go`
- 依赖解析器
- 主要函数：
  - `BuildDependencyGraph(steps []*PlanStep) *DependencyGraph`: 构建依赖图
  - `ResolveDependencies(graph *DependencyGraph) [][]*PlanStep`: 解析依赖
  - `DetectCycles(graph *DependencyGraph) bool`: 检测循环依赖
  - `TopologicalSort(graph *DependencyGraph) []*PlanStep`: 拓扑排序
- 管理操作间的依赖关系

### `rollback_planner.go`
- 回滚规划器
- 主要函数：
  - `PlanRollback(plan *Plan) *Plan`: 生成回滚计划
  - `GenerateUndoSteps(executedSteps []*PlanStep) []*PlanStep`: 生成撤销步骤
  - `ValidateRollbackPlan(rollbackPlan *Plan) error`: 验证回滚计划
- 提供执行失败时的回滚机制

## 规划特性

### 智能分解
- 将复合操作分解为原子步骤
- 保持操作的语义完整性
- 优化步骤的执行顺序

### 并行优化
- 识别可并行执行的步骤
- 管理并行执行的资源竞争
- 确保并行执行的正确性

### 依赖管理
- 自动检测操作间的依赖关系
- 正确处理数据依赖和控制依赖
- 防止死锁和竞态条件

## 在整体架构中的角色

Planner 模块是系统的智能调度层，它将高层意图转换为可执行的详细计划。通过智能规划和优化，Planner 确保了：
- 操作的高效执行
- 资源的合理利用
- 依赖关系的正确处理
- 执行计划的可靠性和可回滚性