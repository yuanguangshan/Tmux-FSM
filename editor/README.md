# editor 模块

## 模块职责概述

`editor/` 是 **Tmux-FSM 的编辑器抽象与核心编辑功能层**，负责提供统一的编辑器接口和实现基础的编辑操作功能。该模块实现了编辑器的核心概念抽象，包括光标、位置、范围等，并提供基础的编辑操作支持。

主要职责包括：
- 定义编辑器的核心数据结构和接口
- 实现基础的编辑操作（移动、删除、插入等）
- 管理光标位置和编辑状态
- 提供编辑操作的统一抽象

## 核心设计思想

- **统一抽象**: 提供统一的编辑器接口抽象
- **状态管理**: 管理编辑器的完整状态
- **操作封装**: 封装基础编辑操作
- **位置感知**: 精确的位置和范围管理

## 文件结构说明

### `editor.go`
- 编辑器核心接口定义
- 主要结构体：
  - `Editor`: 编辑器接口
  - `EditorState`: 编辑器状态
  - `EditOperation`: 编辑操作
- 主要函数：
  - `Insert(text string) error`: 插入文本
  - `Delete(range Range) error`: 删除文本
  - `Replace(range Range, text string) error`: 替换文本
  - `GetCurrentLine() string`: 获取当前行
  - `GetText() string`: 获取全文
- 定义编辑器的核心接口

### `cursor.go`
- 光标管理
- 主要结构体：
  - `Cursor`: 光标位置
  - `CursorState`: 光标状态
- 主要函数：
  - `MoveTo(row, col int)`: 移动光标到指定位置
  - `MoveUp(count int)`: 向上移动
  - `MoveDown(count int)`: 向下移动
  - `MoveLeft(count int)`: 向左移动
  - `MoveRight(count int)`: 向右移动
  - `GetPosition() Position`: 获取当前位置
- 管理光标的精确定位

### `motion.go`
- 运动范围定义
- 主要结构体：
  - `MotionRange`: 运动范围
  - `Position`: 位置坐标
  - `Range`: 文本范围
- 主要函数：
  - `NewRange(start, end Position) Range`: 创建范围
  - `Contains(pos Position) bool`: 检查位置是否在范围内
  - `Intersects(other Range) bool`: 检查范围是否相交
  - `Expand(direction Direction, count int) Range`: 扩展范围
- 定义位置和范围的基本操作

### `operation.go`
- 编辑操作实现
- 主要函数：
  - `ApplyInsert(pos Position, text string) EditorState`: 应用插入操作
  - `ApplyDelete(range Range) EditorState`: 应用删除操作
  - `ApplyReplace(range Range, text string) EditorState`: 应用替换操作
  - `ApplyMotion(motion Motion) EditorState`: 应用运动操作
- 实现具体的编辑操作逻辑

### `buffer.go`
- 缓冲区管理
- 主要结构体：
  - `Buffer`: 文本缓冲区
  - `Line`: 文本行
- 主要函数：
  - `NewBuffer(content string) *Buffer`: 创建缓冲区
  - `GetLine(lineNum int) string`: 获取指定行
  - `GetLines(start, end int) []string`: 获取多行
  - `InsertLine(lineNum int, content string)`: 插入行
  - `DeleteLine(lineNum int)`: 删除行
  - `ReplaceLine(lineNum int, content string)`: 替换行
- 管理文本缓冲区

## 编辑特性

### 精确定位
- 行列坐标系统
- 精确的位置计算
- 高效的范围操作

### 基础操作
- 插入、删除、替换操作
- 移动和选择操作
- 撤销和重做支持

### 状态管理
- 编辑器状态跟踪
- 光标位置管理
- 选择区域维护

## 在整体架构中的角色

Editor 模块是系统的编辑核心层，它提供了编辑操作的基础抽象和实现。Editor 为上层模块提供了：
- 统一的编辑器接口
- 基础的编辑操作支持
- 精确的位置和范围管理
- 编辑状态的完整维护