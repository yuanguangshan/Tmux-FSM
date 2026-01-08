# Transaction Runner 集成指南

## 🎯 目标

将新的 Transaction-based 编辑内核集成到现有的 tmux-fsm 系统中。

## 📋 当前状态

### ✅ 已完成

1. **核心组件**
   - `editor/execution_context.go` - 执行上下文
   - `editor/types.go` - 核心类型定义
   - `editor/engine.go` - Buffer 实现和 ApplyResolvedOperation
   - `editor/selection_update.go` - Selection 更新算法
   - `editor/stores.go` - Store 实现
   - `kernel/transaction.go` - Transaction Runner

2. **文档**
   - `docs/transaction_runner_example.md` - 使用示例
   - `do.md` - 架构评审

3. **演示**
   - `examples/transaction_demo.go` - 可运行的演示程序

### ⚠️ 待集成

1. **main.go 中的调用更新**
   - 第 419 行: `RepeatLastTransaction`
   - 第 456 行: `UndoLastTransaction`
   - 第 621 行: `replayTransaction`

2. **全局 ExecutionContext 创建**
   - 需要在 main.go 中创建全局的 ExecutionContext
   - 替换现有的 GlobalCursorEngine（如果存在）

3. **Fact -> ResolvedOperation 转换**
   - `kernel/transaction.go` 中的 `factToResolvedOp` 需要实现
   - 或者修改 `types.OperationRecord` 直接存储 Inverse 的 ResolvedOperation

## 🔧 集成步骤

### Step 1: 运行演示程序

```bash
cd /Users/ygs/Tmux-FSM
go run examples/transaction_demo.go
```

这将验证核心组件是否正常工作。

### Step 2: 创建全局 ExecutionContext

在 `main.go` 中添加：

```go
// 全局执行上下文
var globalExecContext *editor.ExecutionContext

func initExecutionContext() {
    bufferStore := editor.NewSimpleBufferStore()
    windowStore := editor.NewSimpleWindowStore()
    selectionStore := editor.NewSimpleSelectionStore()
    
    // 创建默认 buffer
    defaultBuffer := editor.NewSimpleBuffer([]string{""})
    bufferStore.Set("default", defaultBuffer)
    
    // 创建默认 window
    defaultWindow := &editor.Window{
        ID:     "default",
        Cursor: editor.Cursor{Row: 0, Col: 0},
    }
    windowStore.Set("default", defaultWindow)
    
    globalExecContext = editor.NewExecutionContext(
        bufferStore,
        windowStore,
        selectionStore,
    )
    globalExecContext.ActiveBuffer = "default"
    globalExecContext.ActiveWindow = "default"
}
```

### Step 3: 更新 ApplyResolvedOperation 调用

将所有的：
```go
editor.ApplyResolvedOperation(op)
```

改为：
```go
editor.ApplyResolvedOperation(globalExecContext, op)
```

### Step 4: 使用 TransactionRunner

```go
runner := kernel.NewTransactionRunner(globalExecContext)

// Apply
runner.Apply(tx)

// Undo
runner.Undo(tx)

// Repeat
runner.Repeat(tx)
```

## 📝 注意事项

### 1. Inverse 操作

当前 `OperationRecord` 中的 `Inverse` 字段是 `core.Fact` 类型。有两种解决方案：

**方案 A: 实现转换函数**
```go
func factToResolvedOp(fact core.Fact) editor.ResolvedOperation {
    // 根据 Fact 的实际结构实现转换
}
```

**方案 B: 修改数据结构（推荐）**
```go
type OperationRecord struct {
    Forward editor.ResolvedOperation
    Inverse editor.ResolvedOperation
}
```

### 2. Selection 更新

当前的 Selection 更新算法是简化版（假设单行操作）。如果需要支持多行操作，需要增强：

```go
// 计算多行文本的 delta
func computeTextDelta(text string) (deltaRow, deltaCol int) {
    lines := strings.Split(text, "\n")
    if len(lines) == 1 {
        return 0, len(lines[0])
    }
    return len(lines) - 1, len(lines[len(lines)-1])
}
```

### 3. 测试

建议为每个核心组件编写测试：

```go
// editor/engine_test.go
func TestApplyResolvedOperation(t *testing.T) {
    // 创建测试上下文
    ctx := createTestContext()
    
    // 创建操作
    op := editor.ResolvedOperation{
        Kind: editor.OpInsert,
        // ...
    }
    
    // 应用并验证
    err := editor.ApplyResolvedOperation(ctx, op)
    assert.NoError(t, err)
}
```

## 🚀 下一步

1. **验证演示程序** - 确保核心功能正常
2. **实现 Inverse 转换** - 完善 Undo 功能
3. **增强 Selection 更新** - 支持多行操作
4. **编写测试** - 确保系统稳定性
5. **集成到 main.go** - 替换现有实现

## 📚 参考文档

- `docs/transaction_runner_example.md` - 详细使用示例
- `do.md` - 架构评审和设计原则
- `editor/types.go` - 核心类型定义
