# undotree 模块

## 模块职责概述

`undotree/` 是 **Tmux-FSM 的撤销树管理系统**，负责实现复杂的撤销/重做功能，支持分支式的撤销历史和时间旅行编辑。该模块实现了基于树结构的撤销历史管理，允许用户在编辑历史的不同分支间切换。

主要职责包括：
- 管理编辑操作的撤销历史树
- 支持分支式的撤销/重做操作
- 实现时间旅行和历史分支切换
- 提供撤销历史的持久化和恢复

## 核心设计思想

- **树形结构**: 使用树结构管理撤销历史
- **分支支持**: 支持撤销历史的分支和合并
- **时间旅行**: 支持跳转到历史任意节点
- **持久化**: 撤销历史的持久化存储

## 文件结构说明

### `undotree.go`
- 核心撤销树实现
- 主要结构体：
  - `UndoTree`: 撤销树
  - `TreeNode`: 树节点
  - `UndoState`: 撤销状态
  - `Branch`: 分支信息
- 主要函数：
  - `NewUndoTree() *UndoTree`: 创建撤销树
  - `AddChange(change Change) *TreeNode`: 添加变更
  - `Undo() *UndoState`: 执行撤销
  - `Redo() *UndoState`: 执行重做
  - `SwitchBranch(branchID string) *UndoState`: 切换分支
  - `GetCurrentState() *UndoState`: 获取当前状态
- 负责核心的撤销树管理

### `node.go`
- 树节点管理
- 主要结构体：
  - `NodeID`: 节点标识符
  - `Change`: 变更信息
  - `NodeMetadata`: 节点元数据
- 主要函数：
  - `NewNode(change Change, parent *TreeNode) *TreeNode`: 创建节点
  - `SetBranch(branchID string)`: 设置分支
  - `GetChildren() []*TreeNode`: 获取子节点
  - `GetParent() *TreeNode`: 获取父节点
  - `IsAncestorOf(node *TreeNode) bool`: 检查祖先关系
- 管理撤销树的节点结构

### `branch_manager.go`
- 分支管理器
- 主要函数：
  - `CreateBranch(fromNode *TreeNode) string`: 创建新分支
  - `GetActiveBranch() string`: 获取活跃分支
  - `SetBranch(branchID string)`: 切换分支
  - `MergeBranch(source, target string) error`: 合并分支
  - `GetBranchHistory(branchID string) []*TreeNode`: 获取分支历史
- 管理撤销历史的分支操作

### `history.go`
- 历史管理
- 主要函数：
  - `SaveHistory(tree *UndoTree, path string) error`: 保存历史
  - `LoadHistory(path string) (*UndoTree, error): 加载历史
  - `PruneHistory(maxNodes int)`: 修剪历史
  - `GetHistoryStats() HistoryStats`: 获取历史统计
- 管理撤销历史的持久化

### `snapshot.go`
- 快照管理
- 主要函数：
  - `CreateSnapshot(state UndoState) Snapshot`: 创建快照
  - `RestoreSnapshot(snapshot Snapshot) UndoState`: 恢复快照
  - `DiffStates(state1, state2 UndoState) []Change`: 比较状态差异
  - `CompressSnapshot(snapshot Snapshot) Snapshot`: 压缩快照
- 管理状态快照

## 撤销特性

### 树形历史
- 支持分支式的撤销历史
- 可以在不同历史分支间切换
- 支持历史的合并和分叉

### 时间旅行
- 可以跳转到历史任意节点
- 支持基于时间点的恢复
- 提供历史浏览功能

### 高效管理
- 智能的历史修剪
- 增量的变更记录
- 内存优化的存储

## 在整体架构中的角色

Undotree 模块是系统的高级撤销管理层，它提供了比传统线性撤销更强大的功能。Undotree 提供了：
- 分支式的撤销历史管理
- 时间旅行编辑能力
- 高级的撤销/重做功能
- 历史状态的持久化支持