package selection

import (
	"tmux-fsm/crdt"
)

//
// ─────────────────────────────────────────────────────────────
//  Types
// ─────────────────────────────────────────────────────────────
//

// CursorID 光标 ID
type CursorID string

// Affinity 亲和性
type Affinity int

const (
	AffinityForward Affinity = iota
	AffinityBackward
	AffinityNeutral
)

// Selection 表示一个选择区域（Anchor → Focus）
type Selection struct {
	Cursor   CursorID
	Actor    crdt.ActorID
	Anchor   crdt.PositionID
	Focus    crdt.PositionID
	Affinity Affinity
}

//
// ─────────────────────────────────────────────────────────────
//  Facts
// ─────────────────────────────────────────────────────────────
//

// SetSelectionFact 设置选择区域（Ephemeral）
type SetSelectionFact struct {
	Cursor CursorID          `json:"cursor"`
	Anchor crdt.PositionID  `json:"anchor"`
	Focus  crdt.PositionID  `json:"focus"`
}

// EphemeralFact 标记接口（不进入 snapshot）
type EphemeralFact interface {
	IsEphemeral() bool
}

// IsEphemeral implements EphemeralFact
func (SetSelectionFact) IsEphemeral() bool {
	return true
}

//
// ─────────────────────────────────────────────────────────────
//  Edit Operations (for transform)
// ─────────────────────────────────────────────────────────────
//

type EditKind int

const (
	EditInsert EditKind = iota
	EditDelete
)

// EditOp 描述一次文本编辑对 selection 的影响
type EditOp struct {
	Kind   EditKind
	Pos    crdt.PositionID // insert position / delete start
	EndPos crdt.PositionID // only for delete
}

//
// ─────────────────────────────────────────────────────────────
//  Selection Transform (Pure Functions)
// ─────────────────────────────────────────────────────────────
//

// TransformSelection 根据编辑操作变换 selection（幂等）
func TransformSelection(sel Selection, op EditOp) Selection {
	switch op.Kind {
	case EditInsert:
		return transformForInsert(sel, op.Pos)
	case EditDelete:
		return transformForDelete(sel, op.Pos, op.EndPos)
	default:
		return sel
	}
}

// 插入操作对 selection 的影响
func transformForInsert(sel Selection, pos crdt.PositionID) Selection {
	a := crdt.ComparePos(pos, sel.Anchor)
	f := crdt.ComparePos(pos, sel.Focus)

	// 插入在 selection 之前或之后 → 不变
	if (a < 0 && f < 0) || (a > 0 && f > 0) {
		return sel
	}

	// 插入正好在 Anchor / Focus，需看 Affinity
	if a == 0 && sel.Affinity == AffinityBackward {
		return sel
	}
	if f == 0 && sel.Affinity == AffinityForward {
		return sel
	}

	// 插入在 selection 内部或中性边界 → 扩展 Focus
	sel.Focus = pos
	return sel
}

// 删除操作对 selection 的影响
func transformForDelete(sel Selection, start, end crdt.PositionID) Selection {
	newAnchor := sel.Anchor
	newFocus := sel.Focus

	// Anchor 被删除 → 吸附到 start
	if crdt.ComparePos(sel.Anchor, start) >= 0 &&
		crdt.ComparePos(sel.Anchor, end) <= 0 {
		newAnchor = start
	}

	// Focus 被删除 → 吸附到 start
	if crdt.ComparePos(sel.Focus, start) >= 0 &&
		crdt.ComparePos(sel.Focus, end) <= 0 {
		newFocus = start
	}

	sel.Anchor = newAnchor
	sel.Focus = newFocus
	return sel
}

//
// ─────────────────────────────────────────────────────────────
//  Selection Manager
// ─────────────────────────────────────────────────────────────
//

// SelectionManager 管理当前所有 selection（可重建）
type SelectionManager struct {
	selections map[CursorID]Selection
}

// NewSelectionManager 创建新的管理器
func NewSelectionManager() *SelectionManager {
	return &SelectionManager{
		selections: make(map[CursorID]Selection),
	}
}

// ApplySelection 应用 SetSelectionFact（覆盖式）
func (sm *SelectionManager) ApplySelection(
	actor crdt.ActorID,
	fact SetSelectionFact,
) {
	sm.selections[fact.Cursor] = Selection{
		Cursor:   fact.Cursor,
		Actor:    actor,
		Anchor:   fact.Anchor,
		Focus:    fact.Focus,
		Affinity: AffinityNeutral,
	}
}

// ApplyEdit 将一次编辑作用到所有 selection
func (sm *SelectionManager) ApplyEdit(op EditOp) {
	for id, sel := range sm.selections {
		sm.selections[id] = TransformSelection(sel, op)
	}
}

// GetSelection 获取指定 cursor 的 selection
func (sm *SelectionManager) GetSelection(
	cursorID CursorID,
) (Selection, bool) {
	sel, ok := sm.selections[cursorID]
	return sel, ok
}

// GetAllSelections 返回 selection 的快照（防止外部 mutate）
func (sm *SelectionManager) GetAllSelections() map[CursorID]Selection {
	out := make(map[CursorID]Selection, len(sm.selections))
	for k, v := range sm.selections {
		out[k] = v
	}
	return out
}
