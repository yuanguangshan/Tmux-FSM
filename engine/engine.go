package engine

import (
	"tmux-fsm/intent"
)

type Cursor struct {
	Line int
	Col  int
}

type Range struct {
	Start Cursor
	End   Cursor
}

// Engine 定义了编辑引擎的接口
type Engine interface {
	Cursor() Cursor

	ComputeMotion(m *intent.Motion) (Range, error)

	MoveCursor(r Range) error

	DeleteRange(r Range) error
	YankRange(r Range) error
	ChangeRange(r Range) error
}