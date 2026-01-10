
# Editor Core Documentation

> 本文档 **严格对应当前代码实现**（共 9 个文件），  
> 不包含未来规划、未落地抽象或外部系统假设。

---

## 模块总体定位

`editor` 模块实现的是一个 **可组合、可回放、可判冲突的编辑执行内核**，核心能力包括：

- ✅ 文本编辑的 **ResolvedOperation** 抽象
- ✅ 操作的 **Footprint 冲突判定**
- ✅ 操作历史的 **Operation DAG**
- ✅ 确定性的 **选区更新算法**
- ✅ Vim 风格 **Text Object / Motion Range 计算**
- ✅ 最小可执行的 **物理执行引擎**

该模块 **不包含**：
- 网络同步
- CRDT
- 权限 / Policy
- UI / TUI
- LSP / AST 投影

---

## 核心抽象关系图

```
ResolvedOperation
   ├── Footprint()        → 冲突检测
   ├── Apply(Buffer)     → 物理执行
   ├── Inverse()         → 可逆性
   ↓
OperationDAG
   ├── 历史结构
   ├── Diff / LCA
   └── 冲突节点
   ↓
ExecutionContext
   ├── BufferStore
   ├── WindowStore
   └── SelectionStore
```

---

## 文件级说明（逐一对应）

---

## `types.go` —— **核心类型与操作代数**

### 基础 ID 类型

```go
type BufferID string
type WindowID string
type OperationID string
type SymbolID string
```

---

### Cursor

```go
type Cursor struct {
    Row int
    Col int
}
```

- 表示文本中的逻辑位置
- 使用 **(Row, Col)**，不是字节偏移
- 提供：
  - `LessThan`
  - `Advance`
  - `Equal`

---

### TextRange / MotionRange

```go
type TextRange struct {
    Start Cursor
    End   Cursor // 半开区间 [Start, End)
}

type MotionRange struct {
    Start Cursor
    End   Cursor
}
```

- `TextRange` 用于 **物理修改**
- `MotionRange` 用于 **语义 motion / text object**

---

### ResolvedOperation（核心接口）

```go
type ResolvedOperation interface {
    OpID() OperationID
    Kind() OpKind
    Apply(buf Buffer) error
    Inverse() (ResolvedOperation, error)
    Footprint() Footprint
}
```

这是系统中**唯一可以被执行、判冲突、组合的操作单位**。

---

### 已实现的操作类型

| 操作 | 说明 |
|----|----|
| InsertOperation | 文本插入 |
| DeleteOperation | 文本删除 |
| MoveOperation | 删除 + 插入 |
| MoveCursorOperation | 光标移动（不改文本） |
| RenameOperation | 语义重命名（不直接改 buffer） |
| CompositeOperation | 复合操作 |

---

### Footprint & EffectKind

```go
type Footprint struct {
    Buffers []BufferID
    Ranges  []TextRange
    Symbols []SymbolRef
    Effects []EffectKind
}
```

`Footprint` 是 **冲突检测的唯一依据**。

`EffectKind`：

- Read
- Write
- Delete
- Rename
- Create

---

## `footprint.go` —— **冲突检测内核**

### 冲突检测入口

```go
func (a Footprint) ConflictsWith(b Footprint)
```

冲突判定顺序：

1. **Buffer 剪枝**
2. **Symbol 冲突（优先级最高）**
3. **TextRange 空间冲突**
4. **EffectKind 决策矩阵**

---

### 冲突输出

```go
type Conflict struct {
    ID     ConflictID
    Left   OperationID
    Right  OperationID
    Reason ConflictReason
    Overlap FootprintOverlap
}
```

用于 DAG 中的 **ConflictNode**。

---

## `dag.go` —— **操作历史 DAG**

### DAGNode

```go
type DAGNode struct {
    ID        DAGNodeID
    Operation ResolvedOperation
    Parents   []DAGNodeID
    Timestamp int64
}
```

- 每个节点 = **一个原子 ResolvedOperation**
- 支持 JSON 序列化（含 op_type）

---

### ConflictNode

```go
type ConflictNode struct {
    Parents   []DAGNodeID
    Conflicts []Conflict
    Resolved  bool
}
```

- 表示 **自动合并失败的阻塞点**

---

### OperationDAG

```go
type OperationDAG struct {
    Nodes     map[DAGNodeID]*DAGNode
    Conflicts map[DAGNodeID]*ConflictNode
    Roots     []DAGNodeID
    Tips      []DAGNodeID
}
```

支持：

- `AddNode`
- `Serialize / Deserialize`

---

## `dag_traversal.go` —— **DAG 算法**

### 提供能力

- `GetAncestors`
- `FindLCA`
- `Diff(base, target)`

`Diff` 语义等价于：

```
git log base..target
```

并返回 **拓扑排序后的操作序列**。

---

## `engine.go` —— **物理执行引擎**

### SimpleBuffer

最小可执行 Buffer 实现：

```go
type SimpleBuffer struct {
    lines []string
}
```

支持：

- InsertAt
- DeleteRange
- RuneAt
- Line / LineCount / LineLength

---

### ApplyResolvedOperation（执行入口）

```go
func ApplyResolvedOperation(ctx *ExecutionContext, op ResolvedOperation) error
```

规则：

- **不做语义判断**
- **严格按 ResolvedOperation 执行**
- `MoveCursorOperation` 走 WindowStore
- 其他操作通过 `Footprint` 找 Buffer

---

## `execution_context.go` —— **执行宇宙**

```go
type ExecutionContext struct {
    Buffers    BufferStore
    Windows    WindowStore
    Selections SelectionStore
}
```

表示 **一次事务执行所需的全部物理资源引用**。

---

## `stores.go` —— **内存存储实现**

提供最小线程安全实现：

- `SimpleBufferStore`
- `SimpleWindowStore`
- `SimpleSelectionStore`

全部为 **map + RWMutex**，无隐藏逻辑。

---

## `selection_update.go` —— **确定性选区更新**

```go
func UpdateSelections(
    selections []Selection,
    ops []ResolvedOperation,
) []Selection
```

特性：

- ✅ 顺序执行
- ✅ 与操作历史无关
- ✅ 仅依赖 ResolvedOperation
- ✅ 结果可重放、可测试

处理：

- Insert
- Delete
- Move
- Composite（递归）

---

## `text_object.go` —— **Vim Text Object 引擎**

### 支持的 Text Object

- word
- ()
- []
- {}
- ""
- ''
- paragraph
- sentence

---

### 核心接口

```go
type TextObjectRangeCalculator interface {
    CalculateRange(obj TextObjectMotion, cursor Cursor)
}
```

实现：

```go
ConcreteTextObjectCalculator
```

**完全基于 Buffer 接口**，无副作用。

---

## 设计不变量（当前代码真实保证）

- ✅ 所有文本修改都通过 `ResolvedOperation`
- ✅ 冲突检测只依赖 `Footprint`
- ✅ DAG 中每个节点 = 一个原子操作
- ✅ 选区更新是确定性的
- ✅ Text Object 计算是纯函数

---

## 一句话总结

> **这是一个“已完成的编辑执行内核”，不是草稿。**  
> 它已经具备：
>
> - 可逆操作  
> - 冲突检测  
> - 历史 DAG  
> - 确定性执行  
> - Vim 级编辑语义  

---

