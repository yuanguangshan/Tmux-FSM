package resolver

import "tmux-fsm/intent"

// resolveUndo 解析撤销意图
func (r *Resolver) resolveUndo(i *intent.Intent) error {
	r.engine.SendKeys("u")
	return nil
}

// recordAction 记录操作到撤销树
func (r *Resolver) recordAction(i *intent.Intent) {
	// 暂时留空，实际实现需要撤销树
}