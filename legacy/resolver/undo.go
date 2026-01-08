package resolver

import (
	"fmt"
	"tmux-fsm/intent"
)

// ErrUndoNotSupportedYet 表示撤销功能尚未实现
var ErrUndoNotSupportedYet = fmt.Errorf("undo not supported: undo tree not implemented")

// resolveUndo 解析撤销意图
func (r *Resolver) resolveUndo(i *intent.Intent) error {
	return ErrUndoNotSupportedYet
}

// recordAction 记录操作到撤销树
func (r *Resolver) recordAction(i *intent.Intent) {
	panic("recordAction called but undo tree not implemented")
}