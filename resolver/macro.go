package resolver

import "tmux-fsm/intent"

// resolveMacro 解析宏意图
func (r *Resolver) resolveMacro(i *intent.Intent) error {
	// 暂时留空，实际实现需要宏管理器
	return nil
}

// recordIntentForMacro 记录意图到宏
func (r *Resolver) recordIntentForMacro(i *intent.Intent) {
	// 暂时留空，实际实现需要宏管理器
}