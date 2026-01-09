# crdt 模块

## 模块职责概述

`crdt/` 是 **Tmux-FSM 的事实合并与冲突解决系统**，负责处理多重合法历史的合并问题。该模块关注的问题是："当出现冲突的历史时，如何将它们'编织'成一个全新的、统一的、合法的历史？"，是系统解决"主权内战"（多重合法历史）的根本大法。

主要职责包括：
- 实现 CRDT 数据结构（如有序树、向量时钟等）
- 处理并发编辑的合并逻辑
- 维护因果关系和版本向量
- 提供高效的位置分配和比较算法
- 将冲突的历史"编织"成统一的合法历史

## 核心设计思想

- **无冲突**: 确保并发操作能够自动合并而不会产生冲突
- **强一致性**: 保证所有副本最终收敛到相同状态
- **因果有序**: 维护事件间的因果关系
- **高效同步**: 支持增量同步和压缩

## 文件结构说明

### `crdt.go`
- CRDT 核心类型定义
- 主要结构体：
  - `PositionID`: 位置标识符
  - `EventID`: 事件标识符  
  - `ActorID`: 参与者标识符
  - `SemanticEvent`: 语义事件
- 主要函数：
  - `ComparePos(a, b PositionID) int`: 比较两个位置
  - `AllocateBetween(after, before *PositionID, actor ActorID) PositionID`: 在两个位置间分配新位置
  - `MergeEvents(events []SemanticEvent) []SemanticEvent`: 合并事件
- 定义了 CRDT 的基础数据类型和操作

### `event_store.go`
- 事件存储实现
- 主要结构体：
  - `EventStore`: 事件存储器
  - `EventLog`: 事件日志
- 主要函数：
  - `NewEventStore() *EventStore`: 创建事件存储
  - `Merge(event SemanticEvent)`: 合并事件
  - `TopoSort() []SemanticEvent`: 拓扑排序事件
  - `Query(filter QueryFilter) []SemanticEvent`: 查询事件
- 负责存储和检索 CRDT 事件

### `position.go`
- 位置管理实现
- 主要函数：
  - `NewPosition(actor ActorID, seq uint64) PositionID`: 创建新位置
  - `ParsePosition(str string) (PositionID, error)`: 解析位置字符串
  - `String() string`: 位置转字符串
- 管理文档中的逻辑位置

### `vector_clock.go`
- 向量时钟实现
- 主要结构体：
  - `VectorClock`: 向量时钟
- 主要函数：
  - `Increment(actor ActorID)`: 递增参与者时钟
  - `Compare(other VectorClock) ClockRelation`: 比较时钟关系
  - `Merge(other VectorClock)`: 合并向量时钟
- 维护因果关系和版本信息

## CRDT 算法特性

### 位置分配算法
- 支持在两个位置之间分配新位置
- 保证位置的全序关系
- 支持高效的插入操作

### 事件合并规则
- 基于因果关系的事件排序
- 支持并发操作的自动合并
- 保证操作的交换律和结合律

## 在整体架构中的角色

CRDT 模块为整个系统提供了强一致性的数据基础，特别是在支持多用户并发编辑的场景下。它确保了即使在网络分区或并发操作的情况下，系统仍能保持数据的一致性和可预测的行为，是实现可回放和可验证特性的关键技术基础。