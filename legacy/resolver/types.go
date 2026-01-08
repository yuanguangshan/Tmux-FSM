package resolver

import (
	"tmux-fsm/intent"
)

// EngineAdapter 引擎适配器接口
type EngineAdapter interface {
	SendKeys(keys ...string)
	GetVisualMode() intent.VisualMode
	EnterVisualMode(mode intent.VisualMode)
	ExitVisualMode()

	// 光标/范围操作
	GetCurrentCursor() ResolverCursor
	ComputeMotion(m *intent.Motion) (ResolverRange, error)
	MoveCursor(r ResolverRange) error

	// 操作范围
	DeleteRange(r ResolverRange) error
	YankRange(r ResolverRange) error
	ChangeRange(r ResolverRange) error
}

// ResolverCursor 解析器光标位置
type ResolverCursor struct {
	Line int
	Col  int
}

// ResolverRange 解析器范围
type ResolverRange struct {
	Start ResolverCursor
	End   ResolverCursor
}

// UndoTree 撤销树（占位）
type UndoTree struct {
	// 实际实现需要更复杂的撤销机制
}

