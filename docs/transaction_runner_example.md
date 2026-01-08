# Transaction Runner 使用示例

## 概述

Transaction Runner 是编辑器内核的核心执行组件，负责在 `ExecutionContext` 中执行事务。

## 核心设计原则

### ✅ 三大不可变原则

1. **ExecutionContext = 执行宇宙**
   - 所有执行都在一个明确的上下文中进行
   - 不依赖任何全局状态
   - 支持多 buffer / 多 window

2. **ResolvedOperation = 冻结的物理操作**
   - 所有语义在 resolve 阶段完成
   - replay 阶段只执行预定义操作
   - 可序列化、可重放

3. **Selection 更新 = Transaction 后的确定性计算**
   - Selection 不是操作的副作用
   - 只在 Transaction commit 后更新
   - 基于物理修改的确定性算法

## 基本使用

### 1. 创建 ExecutionContext

```go
// 创建 stores
bufferStore := editor.NewSimpleBufferStore()
windowStore := editor.NewSimpleWindowStore()
selectionStore := editor.NewSimpleSelectionStore()

// 创建 buffer 和 window
buffer := editor.NewSimpleBuffer([]string{"hello world"})
bufferStore.Set("buf1", buffer)

window := &editor.Window{
    ID:     "win1",
    Cursor: editor.Cursor{Row: 0, Col: 0},
}
windowStore.Set("win1", window)

// 创建执行上下文
ctx := editor.NewExecutionContext(bufferStore, windowStore, selectionStore)
ctx.ActiveBuffer = "buf1"
ctx.ActiveWindow = "win1"
```

### 2. 创建 Transaction Runner

```go
runner := kernel.NewTransactionRunner(ctx)
```

### 3. 执行事务

```go
// 创建一个事务
tx := &types.Transaction{
    ID: 1,
    Records: []types.OperationRecord{
        {
            ResolvedOp: editor.ResolvedOperation{
                Kind:     editor.OpInsert,
                BufferID: "buf1",
                WindowID: "win1",
                Anchor:   editor.Cursor{Row: 0, Col: 5},
                Text:     " beautiful",
            },
        },
    },
}

// 执行事务
if err := runner.Apply(tx); err != nil {
    log.Fatal(err)
}
```

### 4. 撤销事务

```go
if err := runner.Undo(tx); err != nil {
    log.Fatal(err)
}
```

### 5. 重复事务（. repeat）

```go
if err := runner.Repeat(tx); err != nil {
    log.Fatal(err)
}
```

## 完整示例

```go
package main

import (
    "log"
    "tmux-fsm/editor"
    "tmux-fsm/kernel"
    "tmux-fsm/types"
)

func main() {
    // 1. 设置执行环境
    bufferStore := editor.NewSimpleBufferStore()
    windowStore := editor.NewSimpleWindowStore()
    selectionStore := editor.NewSimpleSelectionStore()

    // 2. 创建初始 buffer
    buffer := editor.NewSimpleBuffer([]string{
        "The quick brown fox",
        "jumps over the lazy dog",
    })
    bufferStore.Set("main", buffer)

    // 3. 创建 window
    window := &editor.Window{
        ID:     "main-win",
        Cursor: editor.Cursor{Row: 0, Col: 0},
    }
    windowStore.Set("main-win", window)

    // 4. 创建执行上下文
    ctx := editor.NewExecutionContext(bufferStore, windowStore, selectionStore)
    ctx.ActiveBuffer = "main"
    ctx.ActiveWindow = "main-win"

    // 5. 创建 runner
    runner := kernel.NewTransactionRunner(ctx)

    // 6. 执行一系列操作
    tx := &types.Transaction{
        ID: 1,
        Records: []types.OperationRecord{
            // 移动光标
            {
                ResolvedOp: editor.ResolvedOperation{
                    Kind:     editor.OpMove,
                    BufferID: "main",
                    WindowID: "main-win",
                    Anchor:   editor.Cursor{Row: 0, Col: 4},
                },
            },
            // 插入文本
            {
                ResolvedOp: editor.ResolvedOperation{
                    Kind:     editor.OpInsert,
                    BufferID: "main",
                    WindowID: "main-win",
                    Anchor:   editor.Cursor{Row: 0, Col: 4},
                    Text:     "very ",
                },
            },
        },
    }

    // 7. 应用事务
    if err := runner.Apply(tx); err != nil {
        log.Fatal(err)
    }

    // 8. 验证结果
    buf := bufferStore.Get("main")
    log.Printf("Line 0: %s", buf.Line(0))
    // 输出: Line 0: The very quick brown fox

    // 9. 撤销
    if err := runner.Undo(tx); err != nil {
        log.Fatal(err)
    }

    log.Printf("After undo: %s", buf.Line(0))
    // 输出: After undo: The quick brown fox
}
```

## Selection 更新示例

```go
// 设置初始选区
selectionStore.Set("main", []editor.Selection{
    {
        Start: editor.Cursor{Row: 0, Col: 4},
        End:   editor.Cursor{Row: 0, Col: 9},
    },
})

// 执行插入操作
tx := &types.Transaction{
    Records: []types.OperationRecord{
        {
            ResolvedOp: editor.ResolvedOperation{
                Kind:     editor.OpInsert,
                BufferID: "main",
                Anchor:   editor.Cursor{Row: 0, Col: 4},
                Text:     "XXX",
            },
        },
    },
}

runner.Apply(tx)

// Selection 会自动更新
// 原来: [4, 9)
// 插入 3 个字符在位置 4
// 新的: [4, 12)
updatedSels := selectionStore.Get("main")
log.Printf("Updated selection: %v", updatedSels)
```

## 架构优势

### ✅ 可测试性

```go
func TestInsertOperation(t *testing.T) {
    // 创建隔离的测试环境
    ctx := createTestContext()
    runner := kernel.NewTransactionRunner(ctx)
    
    // 执行操作
    tx := createInsertTransaction("hello")
    runner.Apply(tx)
    
    // 验证结果
    buf := ctx.Buffers.Get("test-buf")
    assert.Equal(t, "hello", buf.Line(0))
}
```

### ✅ 可重放性

```go
// 宏录制
macro := []types.Transaction{tx1, tx2, tx3}

// 宏重放
for _, tx := range macro {
    runner.Repeat(tx)
}
```

### ✅ 跨 Buffer 操作

```go
tx := &types.Transaction{
    Records: []types.OperationRecord{
        // 在 buffer A 中插入
        {
            ResolvedOp: editor.ResolvedOperation{
                BufferID: "bufferA",
                Kind:     editor.OpInsert,
                // ...
            },
        },
        // 在 buffer B 中删除
        {
            ResolvedOp: editor.ResolvedOperation{
                BufferID: "bufferB",
                Kind:     editor.OpDelete,
                // ...
            },
        },
    },
}

// 一次性执行跨 buffer 的原子操作
runner.Apply(tx)
```

## 下一步

1. **实现 Fact -> ResolvedOperation 转换**
   - 当前 `factToResolvedOp` 是占位实现
   - 需要根据实际的 `core.Fact` 结构完善

2. **增强 Selection 更新算法**
   - 当前实现是简化版（假设单行操作）
   - 需要支持多行插入/删除的完整语义

3. **添加 Redo Tree 支持**
   - 当前只有线性 undo
   - 需要实现完整的 redo tree

4. **性能优化**
   - 考虑使用 Rope 或 Piece Table 替换 SimpleBuffer
   - 优化 Selection 更新的批量操作
