package main

import (
	"errors"
	"fmt"
)

// SnapshotManager 管理快照和历史记录
type SnapshotManager struct {
	history *HistoryForResolver
}

// NewSnapshotManager 创建新的快照管理器
func NewSnapshotManager(initialSnapshot Snapshot) *SnapshotManager {
	return &SnapshotManager{
		history: NewHistoryForResolver(initialSnapshot),
	}
}

// PushSnapshot 将新快照推送到历史记录
func (sm *SnapshotManager) PushSnapshot(snapshot Snapshot) {
	sm.history.Push(snapshot)
}

// PerformUndo 执行撤销操作
func (sm *SnapshotManager) PerformUndo() (Snapshot, error) {
	if !sm.history.HasUndo() {
		return sm.history.present, errors.New("nothing to undo")
	}

	snapshot, success := sm.history.Undo()
	if !success {
		return sm.history.present, errors.New("failed to undo")
	}

	return snapshot, nil
}

// PerformRedo 执行重做操作
func (sm *SnapshotManager) PerformRedo() (Snapshot, error) {
	if !sm.history.HasRedo() {
		return sm.history.present, errors.New("nothing to redo")
	}

	snapshot, success := sm.history.Redo()
	if !success {
		return sm.history.present, errors.New("failed to redo")
	}

	return snapshot, nil
}

// GetCurrentSnapshot 获取当前快照
func (sm *SnapshotManager) GetCurrentSnapshot() Snapshot {
	return sm.history.present
}

// HasUndo 检查是否可以撤销
func (sm *SnapshotManager) HasUndo() bool {
	return sm.history.HasUndo()
}

// HasRedo 检查是否可以重做
func (sm *SnapshotManager) HasRedo() bool {
	return sm.history.HasRedo()
}

// TransactionalEditor 提供事务性编辑操作
type TransactionalEditor struct {
	manager *SnapshotManager
}

// NewTransactionalEditor 创建新的事务性编辑器
func NewTransactionalEditor(initialSnapshot Snapshot) *TransactionalEditor {
	return &TransactionalEditor{
		manager: NewSnapshotManager(initialSnapshot),
	}
}

// ApplyIntent 应用意图并更新快照
func (te *TransactionalEditor) ApplyIntent(intent Intent, currentSnapshot Snapshot) (Snapshot, error) {
	// 这里应该根据意图类型应用相应的编辑操作
	// 为了简化，我们只是将当前快照推送到历史记录
	newSnapshot := te.simulateEdit(currentSnapshot, intent)
	te.manager.PushSnapshot(newSnapshot)

	return newSnapshot, nil
}

// simulateEdit 模拟编辑操作（在实际实现中，这里会根据意图执行具体的编辑）
func (te *TransactionalEditor) simulateEdit(snapshot Snapshot, intent Intent) Snapshot {
	// 在实际实现中，这里会根据 Intent 的类型执行相应的编辑操作
	// 例如：删除文本、插入文本、移动光标等
	// 并返回一个新的快照

	// 为了演示目的，我们简单地克隆快照并添加一些变化
	newLines := make([]LineSnapshot, len(snapshot.Lines))
	copy(newLines, snapshot.Lines)

	// 根据意图类型模拟不同的编辑操作
	switch intent.Kind {
	case IntentDelete:
		// 模拟删除操作
		if len(newLines) > 0 {
			// 简单地截断第一行的一部分
			if len(newLines[0].Text) > 5 {
				newLines[0] = LineSnapshot{
					ID:   newLines[0].ID,
					Text: newLines[0].Text[:len(newLines[0].Text)-5],
				}
			}
		}
	case IntentInsert:
		// 模拟插入操作
		if len(newLines) > 0 {
			newLines[0] = LineSnapshot{
				ID:   newLines[0].ID,
				Text: newLines[0].Text + "_inserted",
			}
		}
		// 其他意图类型的处理...
	}

	return Snapshot{
		ID:    generateSnapshotID(),
		Lines: newLines,
	}
}

// generateSnapshotID 生成快照ID
func generateSnapshotID() string {
	// 在实际实现中，这可能是基于内容的哈希或其他唯一标识符
	// 这里我们返回一个简单的字符串，因为无法访问外部的 snapshot 变量
	return fmt.Sprintf("snapshot_%d", len("dummy"))
}
