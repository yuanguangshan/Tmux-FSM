package resolver

import "tmux-fsm/intent"

// handleVisualMode 处理视觉模式切换
func (r *Resolver) handleVisualMode(i *intent.Intent) error {
	if i.Kind == intent.IntentVisual {
		// 使用 Target.Scope 来判断操作
		switch i.Target.Scope {
		case "char":
			return r.EnterVisual(SelectionChar)
		case "line":
			return r.EnterVisual(SelectionLine)
		case "block":
			return r.EnterVisual(SelectionBlock)
		case "cancel":
			return r.ExitVisual()
		}
	}
	return nil
}

// resolveEnterVisual 解析进入视觉模式意图
func (r *Resolver) resolveEnterVisual(i *intent.Intent) error {
	// 通过 Target.Scope 来判断模式
	switch i.Target.Scope {
	case "char":
		return r.EnterVisual(SelectionChar)
	case "line":
		return r.EnterVisual(SelectionLine)
	case "block":
		return r.EnterVisual(SelectionBlock)
	}
	return nil
}

// resolveExitVisual 解析退出视觉模式意图
func (r *Resolver) resolveExitVisual(i *intent.Intent) error {
	return r.ExitVisual()
}

// EnterVisual 进入视觉模式
func (r *Resolver) EnterVisual(mode SelectionMode) error {
	r.engine.EnterVisualMode(intent.VisualMode(mode))
	return nil
}

// ExitVisual 退出视觉模式
func (r *Resolver) ExitVisual() error {
	r.engine.ExitVisualMode()
	return nil
}