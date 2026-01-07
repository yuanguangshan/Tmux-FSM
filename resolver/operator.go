package resolver

import (
	"tmux-fsm/intent"
)

// resolveOperatorWithContext 解析操作意图（带上下文）
func (r *Resolver) resolveOperatorWithContext(i *intent.Intent, ctx ExecContext) error {
	op, ok := i.Meta["operator"].(intent.OperatorKind)
	if !ok {
		return nil
	}

	// 创建语义动作
	action := &Action{
		Type: ActionDelete,
		Semantic: &SemanticAction{
			Operator: op,
			Selection: r.selection, // 使用当前选择
			Motion:   intent.MotionKind(0), // 从元数据获取或默认值
			Target:   i.Target.Kind,
			Count:    i.Count,
		},
		RawIntent:   i,
		Description: "operator action",
	}

	// 尝试从元数据获取motion
	if motionVal, ok := i.Meta["motion"]; ok {
		if motionInt, ok := motionVal.(int); ok {
			action.Semantic.Motion = intent.MotionKind(motionInt)
		} else if motionStr, ok := motionVal.(string); ok {
			// 如果是字符串，需要映射
			switch motionStr {
			case "char":
				action.Semantic.Motion = intent.MotionChar
			case "word":
				action.Semantic.Motion = intent.MotionWord
			case "line":
				action.Semantic.Motion = intent.MotionLine
			case "goto":
				action.Semantic.Motion = intent.MotionGoto
			case "find":
				action.Semantic.Motion = intent.MotionFind
			}
		}
	}

	// 执行语义动作
	return r.executeAction(action)
}

// executeAction 执行语义动作
func (r *Resolver) executeAction(action *Action) error {
	switch action.Type {
	case ActionDelete:
		return r.executeDelete(action)
	case ActionYank:
		return r.executeYank(action)
	case ActionChange:
		return r.executeChange(action)
	case ActionMove:
		return r.executeMove(action)
	case ActionVisual:
		return r.executeVisual(action)
	case ActionUndo:
		return r.executeUndo(action)
	case ActionRepeat:
		return r.executeRepeat(action)
	case ActionMacro:
		return r.executeMacro(action)
	default:
		return nil
	}
}

// executeVisual 执行视觉模式动作
func (r *Resolver) executeVisual(action *Action) error {
	// 视觉模式动作由专门的visual.go处理
	return nil
}

// executeUndo 执行撤销动作
func (r *Resolver) executeUndo(action *Action) error {
	return nil
}

// executeRepeat 执行重复动作
func (r *Resolver) executeRepeat(action *Action) error {
	return nil
}

// executeMacro 执行宏动作
func (r *Resolver) executeMacro(action *Action) error {
	return nil
}

// executeDelete 执行删除动作
func (r *Resolver) executeDelete(action *Action) error {
	// 根据选择范围执行删除
	if action.Semantic.Selection != nil {
		// 有选择范围，删除选择的内容
		return r.engine.DeleteSelection(action.Semantic.Selection)
	} else {
		// 没有选择范围，根据动作执行删除
		return r.engine.DeleteWithMotion(action.Semantic.Motion, action.Semantic.Count)
	}
}

// executeYank 执行复制动作
func (r *Resolver) executeYank(action *Action) error {
	if action.Semantic.Selection != nil {
		return r.engine.YankSelection(action.Semantic.Selection)
	} else {
		return r.engine.YankWithMotion(action.Semantic.Motion, action.Semantic.Count)
	}
}

// executeChange 执行修改动作
func (r *Resolver) executeChange(action *Action) error {
	if action.Semantic.Selection != nil {
		return r.engine.ChangeSelection(action.Semantic.Selection)
	} else {
		return r.engine.ChangeWithMotion(action.Semantic.Motion, action.Semantic.Count)
	}
}

// resolveDelete 解析删除操作
func (r *Resolver) resolveDelete(i *intent.Intent) error {
	// 创建语义动作
	action := &Action{
		Type: ActionDelete,
		Semantic: &SemanticAction{
			Operator: intent.OpDelete,
			Selection: r.selection, // 使用当前选择
			Motion:   intent.MotionKind(i.Meta["motion"].(int)),
			Target:   i.Target.Kind,
			Count:    i.Count,
		},
		RawIntent:   i,
		Description: "delete action",
	}

	// 执行语义动作
	return r.executeAction(action)
}

// resolveYank 解析复制操作
func (r *Resolver) resolveYank(i *intent.Intent) error {
	// 创建语义动作
	action := &Action{
		Type: ActionYank,
		Semantic: &SemanticAction{
			Operator: intent.OpYank,
			Selection: r.selection, // 使用当前选择
			Motion:   intent.MotionKind(i.Meta["motion"].(int)),
			Target:   i.Target.Kind,
			Count:    i.Count,
		},
		RawIntent:   i,
		Description: "yank action",
	}

	// 执行语义动作
	return r.executeAction(action)
}

// resolveChange 解析修改操作
func (r *Resolver) resolveChange(i *intent.Intent) error {
	// 创建语义动作
	action := &Action{
		Type: ActionChange,
		Semantic: &SemanticAction{
			Operator: intent.OpChange,
			Selection: r.selection, // 使用当前选择
			Motion:   intent.MotionKind(i.Meta["motion"].(int)),
			Target:   i.Target.Kind,
			Count:    i.Count,
		},
		RawIntent:   i,
		Description: "change action",
	}

	// 执行语义动作
	return r.executeAction(action)
}