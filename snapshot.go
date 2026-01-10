package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// LineSnapshot 表示一行内容（具有稳定 ID）
// 行号不可信，ID 是唯一锚点
type LineSnapshot struct {
	ID   string // 稳定 ID，跨编辑保持不变
	Text string // 行内容
}

// Snapshot 表示代码快照（不可变）
// 这是 Resolver / Projection 只读的数据结构
type Snapshot struct {
	ID    string // 快照唯一标识
	Lines []LineSnapshot
}

// NewLine 创建一个带稳定 ID 的新行
func NewLine(text string) LineSnapshot {
	return LineSnapshot{
		ID:   generateStableID(text),
		Text: text,
	}
}

// generateStableID 生成一个稳定 ID
// 在实际实现中，这可能基于内容哈希或其他稳定标识符
func generateStableID(text string) string {
	// 生成随机 ID，实际实现可能使用内容哈希或其他机制
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000000))
	return fmt.Sprintf("line_%d_%s", n.Int64(), text[:min(len(text), 5)])
}

// min 是一个辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// LineByID 根据 ID 查找行
func (s Snapshot) LineByID(id string) *LineSnapshot {
	for i := range s.Lines {
		if s.Lines[i].ID == id {
			return &s.Lines[i]
		}
	}
	return nil
}

// LineAtCursor 根据光标状态查找行
func (s Snapshot) LineAtCursor(cursor CursorState) *LineSnapshot {
	return s.LineByID(cursor.LineID)
}

// CursorState 表示运行时光标状态（不序列化，不进 Intent）
type CursorState struct {
	LineID string // 当前行的稳定 ID
	Offset int    // 在行中的偏移量
}

// CursorRefToState 将语义光标引用解析为运行时光标状态
// 这是 Resolver 的职责
func CursorRefToState(ref CursorRef, snapshot Snapshot) (CursorState, error) {
	switch ref.Kind {
	case CursorPrimary:
		// 在实际实现中，这里会从快照中获取主光标位置
		// 现在我们简化处理，返回第一行的开始位置
		if len(snapshot.Lines) > 0 {
			return CursorState{
				LineID: snapshot.Lines[0].ID,
				Offset: 0,
			}, nil
		}
		return CursorState{}, fmt.Errorf("no lines in snapshot")
	case CursorSelectionStart, CursorSelectionEnd:
		// 在实际实现中，这里会从快照中获取选择区域的开始/结束位置
		// 现在我们简化处理
		if len(snapshot.Lines) > 0 {
			return CursorState{
				LineID: snapshot.Lines[0].ID,
				Offset: 0,
			}, nil
		}
		return CursorState{}, fmt.Errorf("no lines in snapshot")
	default:
		return CursorState{}, fmt.Errorf("unknown cursor kind: %d", ref.Kind)
	}
}

// HistoryForResolver 用于实现快照模型下的 Undo/Redo
type HistoryForResolver struct {
	past    []Snapshot
	present Snapshot
	future  []Snapshot
}

// NewHistoryForResolver 创建新的历史记录
func NewHistoryForResolver(initial Snapshot) *HistoryForResolver {
	return &HistoryForResolver{
		past:    []Snapshot{},
		present: initial,
		future:  []Snapshot{},
	}
}

// Push 将新快照添加到历史记录
func (h *HistoryForResolver) Push(snap Snapshot) {
	h.past = append(h.past, h.present)
	h.present = snap
	// 丢弃 future，因为我们在新的分支上
	h.future = []Snapshot{}
}

// Undo 执行撤销操作
func (h *HistoryForResolver) Undo() (Snapshot, bool) {
	if len(h.past) == 0 {
		return h.present, false // 无法撤销
	}

	lastIdx := len(h.past) - 1
	previous := h.past[lastIdx]

	h.future = append([]Snapshot{h.present}, h.future...) // 将当前快照移到 future
	h.present = previous
	h.past = h.past[:lastIdx] // 移除最后一个 past 快照

	return h.present, true
}

// Redo 执行重做操作
func (h *HistoryForResolver) Redo() (Snapshot, bool) {
	if len(h.future) == 0 {
		return h.present, false // 无法重做
	}

	nextIdx := 0
	next := h.future[nextIdx]

	h.past = append(h.past, h.present) // 将当前快照移到 past
	h.present = next
	h.future = h.future[1:] // 移除第一个 future 快照

	return h.present, true
}

// HasUndo 检查是否有可撤销的快照
func (h *HistoryForResolver) HasUndo() bool {
	return len(h.past) > 0
}

// HasRedo 检查是否有可重做的快照
func (h *HistoryForResolver) HasRedo() bool {
	return len(h.future) > 0
}
