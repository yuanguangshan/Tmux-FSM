package builder

import (
	"tmux-fsm/intent"
)

// TextObjectKind 文本对象类型
type TextObjectKind string

const (
	TextObjectInnerParen   TextObjectKind = "inner_paren"
	TextObjectAroundParen  TextObjectKind = "around_paren"
	TextObjectInnerQuote   TextObjectKind = "inner_quote"
	TextObjectAroundQuote  TextObjectKind = "around_quote"
	TextObjectInnerWord    TextObjectKind = "inner_word"
	TextObjectAroundWord   TextObjectKind = "around_word"
)

// TextObjectBuilder 文本对象构建器
type TextObjectBuilder struct{}

// Priority 文本对象优先级较高，因为是明确的选择范围
func (b *TextObjectBuilder) Priority() int {
	return 15
}

// Build 构建文本对象Intent
func (b *TextObjectBuilder) Build(ctx BuildContext) (*intent.Intent, bool) {
	switch ctx.Action {
	case "delete_inner_paren":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetTextObject, Value: string(TextObjectInnerParen)},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpDelete},
			PaneID: ctx.PaneID,
		}, true
	case "delete_around_paren":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetTextObject, Value: string(TextObjectAroundParen)},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpDelete},
			PaneID: ctx.PaneID,
		}, true
	case "delete_inner_quote":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetTextObject, Value: string(TextObjectInnerQuote)},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpDelete},
			PaneID: ctx.PaneID,
		}, true
	case "delete_around_quote":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetTextObject, Value: string(TextObjectAroundQuote)},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpDelete},
			PaneID: ctx.PaneID,
		}, true
	case "change_inner_paren":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetTextObject, Value: string(TextObjectInnerParen)},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpChange},
			PaneID: ctx.PaneID,
		}, true
	case "yank_inner_paren":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetTextObject, Value: string(TextObjectInnerParen)},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpYank},
			PaneID: ctx.PaneID,
		}, true
	default:
		return nil, false
	}
}