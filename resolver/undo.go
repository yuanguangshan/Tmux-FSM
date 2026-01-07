package resolver

import (
	"tmux-fsm/intent"
)

// UndoEntry 撤销条目
type UndoEntry struct {
	Intent *intent.Intent
	Action string // 执行的动作
	State  map[string]interface{} // 执行前的状态
}

// UndoTree 撤销树结构
type UndoTree struct {
	entries []*UndoEntry
	current int
	maxSize int
}

// NewUndoTree 创建新的撤销树
func NewUndoTree(maxSize int) *UndoTree {
	return &UndoTree{
		entries: make([]*UndoEntry, 0, maxSize),
		current: -1,
		maxSize: maxSize,
	}
}

// Add 添加撤销条目
func (ut *UndoTree) Add(entry *UndoEntry) {
	// 如果当前不在末尾，截断后续历史
	if ut.current < len(ut.entries)-1 {
		ut.entries = ut.entries[:ut.current+1]
	}

	// 添加新条目
	ut.entries = append(ut.entries, entry)
	ut.current = len(ut.entries) - 1

	// 如果超出最大大小，移除最早的条目
	if len(ut.entries) > ut.maxSize {
		ut.entries = ut.entries[1:]
		ut.current = len(ut.entries) - 1
	}
}

// Undo 执行撤销
func (ut *UndoTree) Undo() *UndoEntry {
	if ut.current < 0 {
		return nil
	}

	entry := ut.entries[ut.current]
	ut.current--
	return entry
}

// Redo 执行重做
func (ut *UndoTree) Redo() *UndoEntry {
	if ut.current >= len(ut.entries)-1 {
		return nil
	}

	ut.current++
	entry := ut.entries[ut.current]
	return entry
}

// 在resolver中添加undo tree
func (r *Resolver) initUndo() {
	if r.undoTree == nil {
		r.undoTree = NewUndoTree(100) // 最多保存100个操作
	}
}

// resolveUndo 解析撤销意图
func (r *Resolver) resolveUndo(i *intent.Intent) error {
	r.initUndo()

	operation, ok := i.Meta["operation"].(string)
	if !ok {
		operation = "undo" // 默认是撤销
	}

	switch operation {
	case "undo":
		return r.performUndo()
	case "redo":
		return r.performRedo()
	default:
		return r.performUndo() // 默认撤销
	}
}

// performUndo 执行撤销
func (r *Resolver) performUndo() error {
	entry := r.undoTree.Undo()
	if entry == nil {
		// 没有可撤销的操作
		return nil
	}

	// 执行逆向操作
	// 这里需要根据之前的操作来执行逆向操作
	// 例如，如果是删除操作，可能需要粘贴之前删除的内容
	// 如果是插入操作，可能需要删除插入的内容
	switch entry.Action {
	case "delete":
		// 如果有之前删除的内容，可以尝试恢复
		// 这里需要更复杂的逻辑来处理具体撤销
		r.engine.SendKeys("C-z") // 尝试使用系统撤销
	case "insert":
		// 撤销插入可能需要删除插入的内容
		// 这需要更复杂的逻辑
		r.engine.SendKeys("C-z") // 尝试使用系统撤销
	default:
		r.engine.SendKeys("C-z") // 通用撤销
	}

	return nil
}

// performRedo 执行重做
func (r *Resolver) performRedo() error {
	entry := r.undoTree.Redo()
	if entry == nil {
		// 没有可重做的操作
		return nil
	}

	// 重新执行之前的操作
	// 这里需要根据之前的意图重新执行操作
	// 由于Intent是语义化的，我们可以重新解析并执行
	_ = r.Resolve(entry.Intent)

	return nil
}

// recordAction 记录操作以便撤销
func (r *Resolver) recordAction(i *intent.Intent, action string) {
	r.initUndo()

	entry := &UndoEntry{
		Intent: i,
		Action: action,
		State:  make(map[string]interface{}), // 可以保存执行前的状态
	}

	r.undoTree.Add(entry)
}

// 在resolver结构体中添加undo tree字段
// 注意：我们需要在resolver.go中添加这个字段