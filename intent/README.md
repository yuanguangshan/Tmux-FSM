# intent 模块

## 模块职责概述

`intent/` 是 **Tmux-FSM 的语义契约层**，负责描述"用户想做什么"这一层的抽象表示。Intent 关注的问题是："用户的真实目的是什么？"而不是按键、命令或API调用。Intent 是系统中最稳定、最长期存在的语义层。

主要职责包括：
- 定义用户"想做什么"的语义契约
- 提供意图的序列化和反序列化
- 支持复合意图的构建
- 作为 Kernel 决策的唯一输入语义

## 核心设计思想

- **语义契约**: Intent 是契约，不是实现（架构戒律 5）
- **意图抽象**: 将具体操作抽象为高层意图
- **类型安全**: 使用强类型定义各种意图
- **可组合性**: 支持简单意图组合成复杂操作
- **可序列化**: 意图可以被序列化传输和存储
- **与后端无关**: Intent 与后端实现无关，可记录、可重放

## 文件结构说明

### `intent.go`
- 意图的基础定义和核心类型
- 主要结构体：
  - `Intent`: 根意图类型
  - `Motion`: 运动意图
  - `Action`: 动作意图
  - `TextObject`: 文本对象意图
  - `FindMotion`: 查找运动意图
- 主要函数：
  - 各种意图类型的构造函数
  - 意图的验证和规范化函数
- 定义了意图的基本结构和语义

### `grammar_intent.go`
- Grammar 专用意图类型
- 主要结构体：
  - `GrammarIntent`: 语法意图，只包含 Grammar 可以设置的字段
- Grammar 不允许直接构造 Intent，只能构造 GrammarIntent

### `promote.go`
- GrammarIntent → Intent 的提升函数
- 主要函数：
  - `Promote(g *GrammarIntent) *Intent`: 将语法意图提升为标准意图
- 这是 GrammarIntent → Intent 的唯一合法通道
- 确保 Grammar 与 Intent 之间的类型安全转换

### `intent_bridge.go`
- 意图与底层执行系统之间的桥接层
- 主要函数：
  - 意图到具体操作的转换函数
  - 意图解析和验证函数
  - 意图执行结果的反馈处理
- 负责将高层意图转换为可执行的操作

## 意图类型详解

### Motion（运动意图）
- `MotionRange`: 范围运动
- `MotionWord`: 单词运动  
- `MotionLine`: 行运动
- `MotionChar`: 字符运动
- `MotionGoto`: 跳转运动
- `MotionFind`: 查找运动

### Action（动作意图）
- 编辑动作（删除、复制、粘贴、改变等）
- 导航动作
- 模式切换动作

### TextObject（文本对象）
- `Word`: 单词
- `Paren`: 括号
- `QuoteDouble`: 双引号
- 支持 `Inner` 和 `Outer` 两种范围

## 在整体架构中的角色

Intent 模块是整个系统的语义层，它将用户的原始输入（如键盘按键）转换为具有明确语义的意图。这些意图随后被传递给 Engine 和 Kernel 进行处理。Intent 的抽象使得系统可以支持不同的输入方式和编辑模式，同时保持核心逻辑的一致性。