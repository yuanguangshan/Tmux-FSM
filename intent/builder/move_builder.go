package builder

import (
	"tmux-fsm/intent"
)

// MoveBuilder 移动操作构建器
type MoveBuilder struct{}

// Priority 移动操作优先级较高，因为是立即执行的motion
func (b *MoveBuilder) Priority() int {
	return 10
}

// Build 构建移动Intent
func (b *MoveBuilder) Build(ctx BuildContext) (*intent.Intent, bool) {
	switch ctx.Action {
	case "move_left":
		return &intent.Intent{
			Kind:   intent.IntentMove,
			Target: intent.SemanticTarget{Kind: intent.TargetChar, Direction: "left"},
			Count:  ctx.Count,
			PaneID: ctx.PaneID,
		}, true
	case "move_right":
		return &intent.Intent{
			Kind:   intent.IntentMove,
			Target: intent.SemanticTarget{Kind: intent.TargetChar, Direction: "right"},
			Count:  ctx.Count,
			PaneID: ctx.PaneID,
		}, true
	case "move_up":
		return &intent.Intent{
			Kind:   intent.IntentMove,
			Target: intent.SemanticTarget{Kind: intent.TargetChar, Direction: "up"},
			Count:  ctx.Count,
			PaneID: ctx.PaneID,
		}, true
	case "move_down":
		return &intent.Intent{
			Kind:   intent.IntentMove,
			Target: intent.SemanticTarget{Kind: intent.TargetChar, Direction: "down"},
			Count:  ctx.Count,
			PaneID: ctx.PaneID,
		}, true
	case "move_line_start":
		return &intent.Intent{
			Kind:   intent.IntentMove,
			Target: intent.SemanticTarget{Kind: intent.TargetLine, Scope: "start"},
			Count:  ctx.Count,
			PaneID: ctx.PaneID,
		}, true
	case "move_line_end":
		return &intent.Intent{
			Kind:   intent.IntentMove,
			Target: intent.SemanticTarget{Kind: intent.TargetLine, Scope: "end"},
			Count:  ctx.Count,
			PaneID: ctx.PaneID,
		}, true
	default:
		return nil, false
	}
}
