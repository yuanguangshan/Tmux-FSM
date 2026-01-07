package resolver

import (
	"tmux-fsm/intent"
)

// EnterVisual 进入视觉模式
func (r *Resolver) EnterVisual(mode SelectionMode) error {
	if r.selection != nil {
		// 已经在视觉模式中，先退出
		_ = r.ExitVisual()
	}

	// 保存当前光标位置作为锚点
	currentCursor := r.engine.GetCurrentCursor()

	r.selection = &Selection{
		Mode:   mode,
		Anchor: currentCursor,
		Focus:  currentCursor,
	}

	// 通知引擎适配器进入选择模式
	r.engine.EnterSelection(mode)
	return nil
}

// ExitVisual 退出视觉模式
func (r *Resolver) ExitVisual() error {
	if r.selection == nil {
		return nil // 已经不在视觉模式
	}

	// 通知引擎适配器退出选择模式
	r.engine.ExitSelection()

	r.selection = nil
	return nil
}

// UpdateSelection 更新选择区域
func (r *Resolver) UpdateSelection(newFocus Cursor) error {
	if r.selection == nil {
		return nil // 不在视觉模式，无需更新
	}

	r.selection.Focus = newFocus

	// 通知引擎适配器更新选择
	r.engine.UpdateSelection(r.selection.Anchor, r.selection.Focus)
	return nil
}

// resolveVisual 解析视觉模式意图
func (r *Resolver) resolveVisual(i *intent.Intent) error {
	operation, ok := i.Meta["operation"].(string)
	if !ok {
		return nil
	}

	switch operation {
	case "start_char":
		return r.EnterVisual(SelectionChar)
	case "start_line":
		return r.EnterVisual(SelectionLine)
	case "start_block":
		return r.EnterVisual(SelectionBlock)
	case "cancel":
		return r.ExitVisual()
	}

	return nil
}