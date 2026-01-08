package selection

import (
	"tmux-fsm/crdt"
)

// CursorID 光标ID类型
type CursorID string

// Affinity 亲和性类型
type Affinity int

const (
	AffinityForward Affinity = iota
	AffinityBackward
	AffinityNeutral
)

// Selection 选择区域
type Selection struct {
	Cursor   CursorID
	Actor    crdt.ActorID
	Anchor   crdt.PositionID
	Focus    crdt.PositionID
	Affinity Affinity
}

// SetSelectionFact 设置选择区域的事实
type SetSelectionFact struct {
	Cursor CursorID      `json:"cursor"`
	Anchor crdt.PositionID `json:"anchor"`
	Focus  crdt.PositionID `json:"focus"`
}

// EphemeralFact 临时事实接口（不进入快照）
type EphemeralFact interface {
	IsEphemeral() bool
}

// IsEphemeral 表示这是一个临时事实
func (f SetSelectionFact) IsEphemeral() bool {
	return true
}

// SelectionManager 选择区域管理器
type SelectionManager struct {
	selections map[CursorID]Selection
}

// NewSelectionManager 创建新的选择区域管理器
func NewSelectionManager() *SelectionManager {
	return &SelectionManager{
		selections: make(map[CursorID]Selection),
	}
}

// ApplySelection 应用选择区域变更
func (sm *SelectionManager) ApplySelection(actor crdt.ActorID, fact SetSelectionFact) {
	selection := Selection{
		Cursor:   fact.Cursor,
		Actor:    actor,
		Anchor:   fact.Anchor,
		Focus:    fact.Focus,
		Affinity: AffinityNeutral,
	}
	
	sm.selections[fact.Cursor] = selection
}

// GetSelection 获取选择区域
func (sm *SelectionManager) GetSelection(cursorID CursorID) (Selection, bool) {
	selection, exists := sm.selections[cursorID]
	return selection, exists
}

// GetAllSelections 获取所有选择区域
func (sm *SelectionManager) GetAllSelections() map[CursorID]Selection {
	result := make(map[CursorID]Selection)
	for id, sel := range sm.selections {
		result[id] = sel
	}
	return result
}

// UpdateForInsert 处理插入操作对选择区域的影响
func (sm *SelectionManager) UpdateForInsert(pos crdt.PositionID) {
	for cursorID, selection := range sm.selections {
		// 如果插入位置在选择区域内，扩展选择区域
		anchorComp := crdt.ComparePos(selection.Anchor, pos)
		focusComp := crdt.ComparePos(selection.Focus, pos)
		
		// 如果插入在选择区域内
		if (anchorComp <= 0 && focusComp >= 0) || (anchorComp >= 0 && focusComp <= 0) {
			// 根据亲和性决定如何调整
			if selection.Affinity == AffinityForward {
				// 向前扩展
				newFocus := pos
				sm.selections[cursorID] = Selection{
					Cursor:   selection.Cursor,
					Actor:    selection.Actor,
					Anchor:   selection.Anchor,
					Focus:    newFocus,
					Affinity: selection.Affinity,
				}
			}
		} else if anchorComp > 0 {
			// 如果插入在锚点之前，平移整个选择区域
			newAnchor := pos
			sm.selections[cursorID] = Selection{
				Cursor:   selection.Cursor,
				Actor:    selection.Actor,
				Anchor:   newAnchor,
				Focus:    selection.Focus,
				Affinity: selection.Affinity,
			}
		} else if focusComp > 0 {
			// 如果插入在焦点之前，平移焦点
			newFocus := pos
			sm.selections[cursorID] = Selection{
				Cursor:   selection.Cursor,
				Actor:    selection.Actor,
				Anchor:   selection.Anchor,
				Focus:    newFocus,
				Affinity: selection.Affinity,
			}
		}
	}
}

// UpdateForDelete 处理删除操作对选择区域的影响
func (sm *SelectionManager) UpdateForDelete(startPos, endPos crdt.PositionID) {
	for cursorID, selection := range sm.selections {
		anchorCompStart := crdt.ComparePos(selection.Anchor, startPos)
		anchorCompEnd := crdt.ComparePos(selection.Anchor, endPos)
		focusCompStart := crdt.ComparePos(selection.Focus, startPos)
		focusCompEnd := crdt.ComparePos(selection.Focus, endPos)
		
		// 如果锚点在删除范围内，将其吸附到最近的存活位置
		if anchorCompStart >= 0 && anchorCompEnd <= 0 {
			// 锚点在删除范围内，吸附到删除范围的开始
			newAnchor := startPos
			sm.selections[cursorID] = Selection{
				Cursor:   selection.Cursor,
				Actor:    selection.Actor,
				Anchor:   newAnchor,
				Focus:    selection.Focus,
				Affinity: selection.Affinity,
			}
		}
		
		// 如果焦点在删除范围内，将其吸附到最近的存活位置
		if focusCompStart >= 0 && focusCompEnd <= 0 {
			// 焦点在删除范围内，吸附到删除范围的开始
			newFocus := startPos
			currentSel := sm.selections[cursorID]
			sm.selections[cursorID] = Selection{
				Cursor:   currentSel.Cursor,
				Actor:    currentSel.Actor,
				Anchor:   currentSel.Anchor,
				Focus:    newFocus,
				Affinity: currentSel.Affinity,
			}
		}
	}
}