package editor

// ExecutionContext 执行上下文
// 这是「一次 Transaction 的物理执行宇宙」
// 它持有执行 Transaction 所需的所有物理资源引用
type ExecutionContext struct {
	Buffers    BufferStore
	Windows    WindowStore
	Selections SelectionStore

	ActiveBuffer BufferID
	ActiveWindow WindowID
}

// NewExecutionContext 创建新的执行上下文
func NewExecutionContext(buffers BufferStore, windows WindowStore, selections SelectionStore) *ExecutionContext {
	return &ExecutionContext{
		Buffers:    buffers,
		Windows:    windows,
		Selections: selections,
	}
}
