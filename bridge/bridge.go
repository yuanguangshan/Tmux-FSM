package bridge

import (
	"tmux-fsm/fsm"
	"tmux-fsm/intent"
)

// Invariant: FSM.Dispatch MUST only be called by bridge.HandleIntent
// This ensures single source of truth for key consumption semantics

// HandleIntent 处理意图，优先处理FSM按键事件
// 返回值表示是否消费了该意图
func HandleIntent(in intent.Intent) bool {
	// Invariant 1: FSM has absolute priority on key events
	if in.Meta != nil {
		if key, ok := in.Meta["key"].(string); ok {
			// FSM.Dispatch 返回 true 表示按键已被消费
			consumed := fsm.Dispatch(key)
			if consumed {
				// Invariant 7: consumed key MUST NOT fall through
				return true
			}
		}
		// 检查是否是reload命令
		if cmd, ok := in.Meta["command"].(string); ok {
			if cmd == "reload" {
				// 处理reload命令
				configPath, ok := in.Meta["config_path"].(string)
				if !ok {
					configPath = "./keymap.yaml"
				}
				// 使用统一的Reload函数
				if err := fsm.Reload(configPath); err != nil {
					// 这里应该记录错误，但根据不变量10，我们不能继续运行
					// 在实际实现中，这可能需要更复杂的错误处理
					return true
				}
				return true
			}
			if cmd == "nvim-mode" {
				// 处理Neovim模式变化
				mode, ok := in.Meta["mode"].(string)
				if ok {
					fsm.OnNvimMode(mode)
				}
				return true
			}
		}
	}
	return false
}