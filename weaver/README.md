# weaver 模块

## 模块职责概述

`weaver/` 是 **Tmux-FSM 的系统装配层（Composition Root）与事实解析系统**，负责将各个模块正确地创建、组合并协同工作，并处理事实的解析与执行。该模块关注的问题是："这些模块应该如何被创建、组合，并协同工作？"以及"如何将抽象事实解析为具体操作？"，是系统的装配工厂和执行枢纽。

主要职责包括：
- 系统模块的装配和依赖注入
- 管理不同环境下的模块实例化（真实/模拟后端）
- 提供系统的统一装配入口
- 控制系统的运行模式配置
- 事实解析与执行（Resolver 负责将抽象事实解析为具体操作）

## 核心设计思想

- **单一装配入口**: 系统中只有一个地方负责模块实例化和依赖注入
- **显式依赖**: 所有依赖通过构造函数参数明确注入
- **可替换性**: 支持不同实现的模块替换（真实/模拟后端）
- **装配工厂**: Weaver 只负责装配，不参与业务逻辑

## 文件结构说明

### `weaver.go`
- 系统装配器实现
- 主要结构体：
  - `Weaver`: 系统编织器
  - `Config`: 配置定义
- 主要函数：
  - `NewWeaver(config Config) *Weaver`: 创建编织器
  - `AssembleSystem()`: 装配系统
  - `ConfigureRuntime()`: 配置运行时
- 负责系统的装配和配置

### `core/resolved_fact.go`
- 事实解析系统
- 主要结构体：
  - `ResolvedAnchor`: 解析后的锚点
  - `ResolvedFact`: 解析后的事实
- 负责将抽象事实解析为具体操作位置
- 实现 Phase 5.2: Anchor Primacy 原则

### `logic/passthrough_resolver.go`
- 传递解析器
- 主要结构体：
  - `PassthroughResolver`: 传递解析器
- 主要函数：
  - `ResolveFacts(facts []core.Fact, expectedHash string) ([]core.ResolvedFact, error)`: 解析事实
  - `resolveAnchorWithSnapshot(a core.Anchor, s core.Snapshot) (core.ResolvedAnchor, error)`: 使用快照解析锚点
  - `resolveAnchor(a core.Anchor) (core.ResolvedAnchor, error)`: 解析锚点
- 负责将抽象事实解析为具体操作位置
- 实现 Phase 5.3: Inverse Integrity 原则

### `core/history.go`
- 历史管理器
- 主要结构体：
  - `History`: 历史记录
  - `HistoryEntry`: 历史条目
- 管理编辑操作的历史记录

### `core/shadow_engine.go`
- 影子引擎
- 主要结构体：
  - `ShadowEngine`: 影子引擎
- 负责在不修改真实状态的情况下验证操作
- 实现 Phase 6: Temporal Freezing 原则

### `core/snapshot_diff.go`
- 快照差异比较器
- 主要函数：
  - `CompareSnapshots(a, b core.Snapshot) core.Diff`: 比较快照差异
- 用于检测理念世界与现实世界的差异

## 核心特性

### 事实解析
- 将抽象事实解析为具体操作位置
- 支持基于快照的精确解析
- 实现 Anchor Primacy 原则

### 历史管理
- 管理编辑操作的历史记录
- 支持可验证的历史追踪
- 提供时间旅行能力

### 一致性验证
- 影子引擎验证操作安全性
- 快照差异检测机制
- 实现 Temporal Freezing 原则

## 在整体架构中的角色

Weaver 模块是系统的执行枢纽，它确保所有操作都能被正确解析和执行。通过事实解析和一致性验证，Weaver 确保了：
- 事实的正确解析和执行
- 操作的安全性验证
- 历史的可追溯性
- 理念世界与现实世界的同步