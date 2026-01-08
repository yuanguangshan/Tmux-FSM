

# 编辑器内核设计文档（v1.0）

## 1. 设计目标（Design Goals）

本编辑器内核的目标是提供一个 **可组合、可回放、可验证** 的文本编辑执行模型，使以下能力在同一抽象下成立：

- 单命令 / 组合命令 / 宏
- 多选区（Kakoune 风格）
- Text Object
- Ex / 批量编辑
- 跨 buffer 操作
- 稳定 undo / redo / repeat

**核心原则**：  
> 所有编辑行为都必须能够被表示为一组可重放的 Transaction。

---

## 2. 核心执行模型（Execution Model）

编辑行为被严格分解为四个阶段：

```
Parse → Resolve → Transaction → Replay
```

### 2.1 Parse（语法解析）

- 将用户输入解析为语义指令（Command）
- 不访问 buffer
- 不产生副作用

### 2.2 Resolve（语义冻结）

- 在给定 EditorState + BufferSnapshot 上
- 将 Command 解析为 **ResolvedOperation**
- Resolve 过程中：
  - ❌ 不修改 buffer
  - ✅ 可枚举多选区 / 多 buffer
  - ✅ 冻结所有坐标与文本内容

### 2.3 Transaction（原子编辑单元）

```go
type Transaction struct {
	Ops []OperationRecord
}
```

- Transaction 是 undo / redo 的最小单位
- Transaction 内 Operation 必须：
  - 按 buffer 坐标 **逆序应用**
  - 不依赖运行时状态

### 2.4 Replay（执行）

- 唯一修改 buffer 的阶段
- 不进行逻辑判断
- 完全 deterministic

---

## 3. ResolvedOperation 与不变量（Invariants）

### 3.1 ResolvedOperation

```go
type ResolvedOperation struct {
	BufferID BufferID
	Range    TextRange
	Payload  OperationPayload
}
```

### 3.2 内核不变量（必须永远成立）

1. **Resolve 不修改 buffer**
2. **Replay 不做语义判断**
3. **Selection 更新不是 Operation 的副作用**
4. **Transaction 必须可完全 replay**
5. **同一输入 + 同一状态 = 同一 Transaction**

违反以上任一条，即为内核 bug。

---

## 4. Selection 模型

### 4.1 Selection

```go
type Selection struct {
	Anchor CursorPos
	Caret  CursorPos
}
```

- Selection 有方向
- Anchor 不随 motion 改变
- Caret 为活动端

### 4.2 SelectionSet

```go
type SelectionSet struct {
	Selections []Selection
	Primary    int
}
```

- 所有 resolve 都对 SelectionSet 工作
- 每个 Selection 独立 resolve
- 多选区编辑结果合并为单个 Transaction

---

## 5. Text Object 语义

- Text Object 是 **范围解析器**
- 不移动 cursor
- 不直接产生 Operation
- 可组合（inner / around）

Text Object 只参与 Resolve 阶段。

---

## 6. Ex 命令模型

- Ex 命令是 **Transaction 生成器**
- Ex 命令：
  - 可以生成一个或多个 Transaction
  - 不直接参与 undo / redo
- `:s`、`:global`、`:windo` 本质为批量 resolve

---

## 7. 宏与重复（Macro / Repeat）

- 宏存储为 `[]Transaction`
- 执行宏 = replay Transaction 列表
- 宏：
  - 可嵌套
  - 可跨 buffer
  - 与 undo tree 正交

---

## 8. Undo / Redo 模型

- Undo / Redo 基于 Transaction
- 支持分叉（tree）
- Redo 在新 Transaction 后失效

---

## 9. 扩展边界（Non-goals）

以下内容不属于内核职责：

- 渲染 / UI
- 键位绑定
- LSP / AST 语义（但内核为其预留接口）
- 性能实现细节（rope / piece table）

---

## 10. 版本与演进策略

- Transaction / Operation 需携带版本号
- 内核 API 修改需保证：
  - 旧 Transaction 可 replay
  - 不破坏不变量

---

## 结语

> **这是一个编辑器 DSL 的执行内核，而不是命令解释器。**

只要不变量成立，  
任何新能力都只能是 **Resolve 层的扩展**。

---
	
		