package resolver

import (
	"tmux-fsm/intent"
)

// SemanticAction 语义动作，代表意图的语义表示
type SemanticAction struct {
	Operator  intent.OperatorKind
	Selection *Selection
	Motion    intent.MotionKind
	Target    intent.TargetKind
	Count     int
}

// ActionType 语义动作类型
type ActionType int

const (
	ActionMove ActionType = iota
	ActionDelete
	ActionYank
	ActionChange
	ActionVisual
	ActionUndo
	ActionRepeat
	ActionMacro
)

// Action 代表一个具体的语义动作
type Action struct {
	Type        ActionType
	Semantic    *SemanticAction
	RawIntent   *intent.Intent
	Description string
}