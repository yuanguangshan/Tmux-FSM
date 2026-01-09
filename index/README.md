# index 模块

## 模块职责概述

`index/` 是 **Tmux-FSM 的事件索引与查询系统**，负责为大量的事件数据建立高效的索引结构，支持快速的事件查询和检索。该模块实现了多维度的索引机制，使系统能够快速定位和检索特定的事件。

主要职责包括：
- 为事件数据建立多维度索引
- 提供高效的事件查询能力
- 支持复杂查询条件的快速检索
- 维护索引的一致性和实时性

## 核心设计思想

- **多维索引**: 支持按多种维度建立索引
- **实时更新**: 索引随事件实时更新
- **高效查询**: 提供亚线性时间复杂度的查询
- **内存优化**: 平衡内存使用和查询性能

## 文件结构说明

### `index.go`
- 核心索引管理器
- 主要结构体：
  - `EventIndex`: 事件索引管理器
  - `IndexEntry`: 索引条目
  - `QueryResult`: 查询结果
  - `IndexConfig`: 索引配置
- 主要函数：
  - `NewEventIndex(config IndexConfig) *EventIndex`: 创建事件索引
  - `IndexEvent(event SemanticEvent) error`: 为事件建立索引
  - `QueryByActor(actor ActorID) []EventID`: 按参与者查询
  - `QueryByType(ft FactType) []EventID`: 按类型查询
  - `QueryByTimeRange(start, end time.Time) []EventID`: 按时间范围查询
  - `QueryAIChanges(aiActorPrefix string) []EventID`: 查询 AI 变更
- 负责核心的索引和查询功能

### `actor_index.go`
- 参与者索引
- 主要函数：
  - `BuildActorIndex(events []SemanticEvent) map[ActorID][]EventID`: 构建参与者索引
  - `GetActorEvents(actor ActorID) []EventID`: 获取参与者事件
  - `GetActorStats(actor ActorID) ActorStats`: 获取参与者统计
  - `UpdateActorIndex(actor ActorID, event EventID)`: 更新参与者索引
- 按参与者维度建立和维护索引

### `time_index.go`
- 时间索引
- 主要函数：
  - `BuildTimeIndex(events []SemanticEvent) TimeIndexTree`: 构建时间索引树
  - `QueryByTimestamp(ts time.Time) []EventID`: 按时间戳查询
  - `QueryByTimeRange(start, end time.Time) []EventID`: 按时间范围查询
  - `GetTimeDistribution() map[string]int`: 获取时间分布
- 按时间维度建立索引

### `type_index.go`
- 类型索引
- 主要函数：
  - `BuildTypeIndex(events []SemanticEvent) map[FactType][]EventID`: 构建类型索引
  - `GetTypeEvents(ft FactType) []EventID`: 获取指定类型事件
  - `GetTypeStats() map[FactType]TypeStats`: 获取类型统计
  - `UpdateTypeIndex(ft FactType, event EventID)`: 更新类型索引
- 按事件类型建立索引

### `ai_index.go`
- AI 变更索引
- 主要函数：
  - `BuildAIIndex(events []SemanticEvent) map[string][]EventID`: 构建 AI 索引
  - `QueryAIChanges(prefix string) []EventID`: 查询 AI 变更
  - `GetAIActivityStats() map[string]AIStats`: 获取 AI 活动统计
  - `IsAIEvent(event SemanticEvent) bool`: 判断是否为 AI 事件
- 专门索引 AI 相关的变更

## 索引特性

### 多维度查询
- 支持按参与者查询
- 支持按时间范围查询
- 支持按事件类型查询
- 支持按 AI 参与者查询

### 高效性能
- O(log n) 查询时间复杂度
- 增量索引更新
- 内存友好的数据结构

### 实时性
- 事件到达时即时索引
- 索引与事件数据一致性
- 支持实时查询

## 在整体架构中的角色

Index 模块是系统的查询加速层，它通过建立高效的索引结构，使系统能够快速检索和分析历史事件。Index 提供了：
- 快速的事件查询能力
- 多维度的数据分析支持
- 高效的历史数据分析
- 实时的统计和监控功能