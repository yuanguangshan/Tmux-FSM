package builder

import (
	"tmux-fsm/intent"
)

// MacroBuilder 宏构建器
type MacroBuilder struct{}

// Priority 宏操作优先级中等
func (b *MacroBuilder) Priority() int {
	return 8
}

// Build 构建宏Intent
func (b *MacroBuilder) Build(ctx BuildContext) (*intent.Intent, bool) {
	switch ctx.Action {
	case "start_macro":
		register, ok := ctx.Meta["register"].(string)
		if !ok {
			register = "a" // 默认注册器
		}
		return &intent.Intent{
			Kind:   intent.IntentMacro,
			Target: intent.SemanticTarget{Kind: intent.TargetNone, Scope: "start"},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operation": "start_recording", "register": register},
			PaneID: ctx.PaneID,
		}, true
	case "stop_macro":
		return &intent.Intent{
			Kind:   intent.IntentMacro,
			Target: intent.SemanticTarget{Kind: intent.TargetNone, Scope: "stop"},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operation": "stop_recording"},
			PaneID: ctx.PaneID,
		}, true
	case "play_macro":
		register, ok := ctx.Meta["register"].(string)
		if !ok {
			register = "a" // 默认注册器
		}
		return &intent.Intent{
			Kind:   intent.IntentMacro,
			Target: intent.SemanticTarget{Kind: intent.TargetNone, Scope: "play"},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operation": "play", "register": register},
			PaneID: ctx.PaneID,
		}, true
	default:
		return nil, false
	}
}