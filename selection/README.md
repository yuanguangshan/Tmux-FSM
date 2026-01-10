# selection 模块

## 模块职责概述

`selection/` 是 **Tmux-FSM 的选择区域管理系统**，负责管理文本编辑中的选择区域、光标位置和选择状态。该模块实现了复杂的选择区域计算和管理功能，支持多种选择模式和跨行选择操作。

主要职责包括：
- 管理文本选择区域的状态
- 计算和更新选择区域范围
- 处理选择区域的合并和分割
- 支持多光标和多选择区域操作

## 核心设计思想

- **精确选择**: 提供精确的选择区域计算
- **状态管理**: 维护选择区域的完整状态
- **多选择支持**: 支持多个选择区域
- **高效更新**: 优化选择区域的更新性能

## 文件结构说明

### `selection.go`
- 核心选择区域定义
- 主要结构体：
  - `Selection`: 选择区域
  - `CursorID`: 光标标识符
  - `SelectionType`: 选择类型
  - `SelectionState`: 选择状态
- 主要函数：
  - `NewSelection(start, end Position) Selection`: 创建选择区域
  - `Contains(pos Position) bool`: 检查位置是否在选择区域内
  - `Intersects(other Selection) bool`: 检查选择区域是否相交
  - `Merge(other Selection) Selection`: 合并选择区域
- 定义选择区域的基本操作

### `store.go`
- 选择存储管理器
- 主要结构体：
  - `SelectionStore`: 选择存储接口
  - `SimpleSelectionStore`: 简单选择存储实现
- 主要函数：
  - `NewSimpleSelectionStore() *SimpleSelectionStore`: 创建简单选择存储
  - `GetSelection(actorID string) (Selection, bool)`: 获取选择区域
  - `SetSelection(actorID string, selection Selection)`: 设置选择区域
  - `DeleteSelection(actorID string)`: 删除选择区域
- 管理系统中的所有选择区域

### `calculator.go`
- 选择区域计算器
- 主要函数：
  - `CalculateRange(startPos Position, motion Motion) (Selection, error)`: 计算选择范围
  - `ExpandSelection(sel Selection, direction Direction) Selection`: 扩展选择区域
  - `ContractSelection(sel Selection) Selection`: 收缩选择区域
  - `NormalizeSelection(sel Selection) Selection`: 标准化选择区域
- 计算选择区域的具体范围

### `cursor.go`
- 光标管理
- 主要结构体：
  - `Cursor`: 光标位置
  - `CursorPosition`: 光标位置信息
- 主要函数：
  - `MoveCursor(cursor *Cursor, motion Motion)`: 移动光标
  - `GetPosition(cursor *Cursor) Position`: 获取光标位置
  - `SetPosition(cursor *Cursor, pos Position)`: 设置光标位置
  - `CloneCursor(cursor *Cursor) *Cursor`: 克隆光标
- 管理光标位置和移动

### `transform.go`
- 选择区域变换
- 主要函数：
  - `TransformSelection(sel Selection, transform TransformOp) Selection`: 变换选择区域
  - `ApplyEdit(selection Selection, edit EditOperation) Selection`: 应用编辑操作到选择区域
  - `OffsetSelection(sel Selection, offset Position) Selection`: 偏移选择区域
  - `ScaleSelection(sel Selection, factor float64) Selection`: 缩放选择区域
- 处理选择区域的变换操作

## 选择特性

### 多选择支持
- 支持多个独立的选择区域
- 支持多光标操作
- 选择区域间的协调管理

### 智能计算
- 基于运动意图的智能选择计算
- 支持复杂的选择扩展
- 自动选择区域标准化

### 高效更新
- 增量选择区域更新
- 优化的选择区域合并
- 快速的选择区域查询

## 在整体架构中的角色

Selection 模块是系统的文本选择管理层，它精确管理用户的选择操作和光标位置。Selection 提供了：
- 精确的选择区域计算
- 高效的选择状态管理
- 复杂选择操作的支持
- 与编辑操作的无缝集成