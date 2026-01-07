package resolver

import (
	"tmux-fsm/intent"
)

// resolveMove 解析移动意图
func (r *Resolver) resolveMove(i *intent.Intent) error {
	count := max(1, i.Count)

	// 检查是否在视觉模式下
	inVisualMode := r.selection != nil

	switch i.Target.Kind {
	case intent.TargetChar:
		return r.resolveCharMove(i, count, inVisualMode)

	case intent.TargetLine:
		return r.resolveLineMove(i, inVisualMode)

	case intent.TargetWord:
		return r.resolveWordMove(i, count, inVisualMode)

	default:
		return nil
	}
}

// resolveCharMove 解析字符级移动
func (r *Resolver) resolveCharMove(i *intent.Intent, count int, inVisualMode bool) error {
	// 创建移动动作
	action := &Action{
		Type: ActionMove,
		Semantic: &SemanticAction{
			Motion: intent.MotionChar,
			Target: i.Target.Kind,
			Count:  count,
		},
		RawIntent:   i,
		Description: "char move",
	}

	// 执行移动动作
	err := r.executeAction(action)
	if err != nil {
		return err
	}

	// 如果在视觉模式下，更新选择区域
	if inVisualMode {
		newFocus := r.engine.GetCurrentCursor()
		_ = r.UpdateSelection(newFocus)
	}

	return nil
}

// executeMove 执行移动动作
func (r *Resolver) executeMove(action *Action) error {
	var key string

	// 根据方向确定按键
	switch action.RawIntent.Target.Direction {
	case "left":
		key = "Left"
	case "right":
		key = "Right"
	case "up":
		key = "Up"
	case "down":
		key = "Down"
	default:
		// 如果没有明确方向，尝试从Value中获取
		if action.RawIntent.Target.Value == "h" {
			key = "Left"
		} else if action.RawIntent.Target.Value == "j" {
			key = "Down"
		} else if action.RawIntent.Target.Value == "k" {
			key = "Up"
		} else if action.RawIntent.Target.Value == "l" {
			key = "Right"
		}
	}

	// 发送按键
	for n := 0; n < action.Semantic.Count; n++ {
		r.engine.SendKeys(key)
	}

	return nil
}

// resolveLineMove 解析行级移动
func (r *Resolver) resolveLineMove(i *intent.Intent, inVisualMode bool) error {
	// 创建移动动作
	action := &Action{
		Type: ActionMove,
		Semantic: &SemanticAction{
			Motion: intent.MotionLine,
			Target: i.Target.Kind,
			Count:  1,
		},
		RawIntent:   i,
		Description: "line move",
	}

	// 执行移动动作
	err := r.executeAction(action)
	if err != nil {
		return err
	}

	// 如果在视觉模式下，更新选择区域
	if inVisualMode {
		newFocus := r.engine.GetCurrentCursor()
		_ = r.UpdateSelection(newFocus)
	}

	return nil
}

// resolveWordMove 解析单词级移动
func (r *Resolver) resolveWordMove(i *intent.Intent, count int, inVisualMode bool) error {
	// 创建移动动作
	action := &Action{
		Type: ActionMove,
		Semantic: &SemanticAction{
			Motion: intent.MotionWord,
			Target: i.Target.Kind,
			Count:  count,
		},
		RawIntent:   i,
		Description: "word move",
	}

	// 执行移动动作
	err := r.executeAction(action)
	if err != nil {
		return err
	}

	// 如果在视觉模式下，更新选择区域
	if inVisualMode {
		newFocus := r.engine.GetCurrentCursor()
		_ = r.UpdateSelection(newFocus)
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}