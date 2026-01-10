package main

import (
	"fmt"
	"log"

	"tmux-fsm/editor"
	"tmux-fsm/kernel"
	"tmux-fsm/types"
)

// 这是一个最小的 Transaction Runner 演示
// 展示如何使用新的执行上下文系统

func main() {
	fmt.Println("=== Transaction Runner Demo ===")

	// 1. 创建 Stores
	bufferStore := editor.NewSimpleBufferStore()
	windowStore := editor.NewSimpleWindowStore()
	selectionStore := editor.NewSimpleSelectionStore()

	// 2. 创建初始 Buffer
	buffer := editor.NewSimpleBuffer([]string{
		"Hello World",
		"This is a test",
	})
	bufferStore.Set("main", buffer)

	// 3. 创建 Window
	window := &editor.Window{
		ID:     "main-win",
		Cursor: editor.Cursor{Row: 0, Col: 6},
	}
	windowStore.Set("main-win", window)

	// 4. 创建 ExecutionContext
	ctx := editor.NewExecutionContext(bufferStore, windowStore, selectionStore)
	ctx.ActiveBuffer = "main"
	ctx.ActiveWindow = "main-win"

	// 5. 创建 TransactionRunner
	runner := kernel.NewTransactionRunner(ctx)

	// 6. 创建一个简单的 Transaction（插入文本）
	tx := &types.Transaction{
		ID: 1,
		Records: []types.OperationRecord{
			{
				ResolvedOp: &editor.InsertOperation{
					ID:     "demo_insert_1",
					Buffer: "main",
					At:     editor.Cursor{Row: 0, Col: 6},
					Text:   "Beautiful ",
				},
			},
		},
	}

	// 7. 打印初始状态
	fmt.Println("初始状态:")
	printBuffer(bufferStore.Get("main"))

	// 8. 应用 Transaction
	fmt.Println("\n执行: 在位置 (0, 6) 插入 'Beautiful '")
	if err := runner.Apply(tx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n应用后:")
	printBuffer(bufferStore.Get("main"))

	// 9. 创建第二个 Transaction（删除）
	tx2 := &types.Transaction{
		ID: 2,
		Records: []types.OperationRecord{
			{
				ResolvedOp: &editor.DeleteOperation{
					ID:     "demo_delete_1",
					Buffer: "main",
					Range: editor.TextRange{
						Start: editor.Cursor{Row: 0, Col: 0},
						End:   editor.Cursor{Row: 0, Col: 6},
					},
				},
			},
		},
	}

	fmt.Println("\n执行: 删除 (0, 0) 到 (0, 6)")
	if err := runner.Apply(tx2); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n应用后:")
	printBuffer(bufferStore.Get("main"))

	// 10. 演示 Repeat（重复最后一个操作）
	fmt.Println("\n执行: Repeat (重复删除)")
	if err := runner.Repeat(tx2); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n应用后:")
	printBuffer(bufferStore.Get("main"))

	fmt.Println("\n=== Demo 完成 ===")
}

func printBuffer(buf editor.Buffer) {
	if buf == nil {
		fmt.Println("  (buffer is nil)")
		return
	}
	for i := 0; i < buf.LineCount(); i++ {
		fmt.Printf("  Line %d: %s\n", i, buf.Line(i))
	}
}
