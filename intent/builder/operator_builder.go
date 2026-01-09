package builder

import (
	"tmux-fsm/intent"
)

// OperatorBuilder 操作符构建器
type OperatorBuilder struct{}

// Priority 操作符优先级较低，因为需要等待motion
func (b *OperatorBuilder) Priority() int {
	return 5
}

// Build 构建操作符Intent
func (b *OperatorBuilder) Build(ctx BuildContext) (*intent.Intent, bool) {
	switch ctx.Action {
	case "delete":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetChar},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpDelete},
			PaneID: ctx.PaneID,
		}, true
	case "yank":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetChar},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpYank},
			PaneID: ctx.PaneID,
		}, true
	case "change":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetChar},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpChange},
			PaneID: ctx.PaneID,
		}, true
	default:
		return nil, false
	}
}

// TODO: Operator intents currently encode legacy operator semantics in Meta.
// This MUST be replaced by first-class intent kinds before Cut 3.
