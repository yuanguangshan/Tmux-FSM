package fsm

import (
	"strings"
)

// OnNvimMode 处理来自 Neovim 的模式变化
func OnNvimMode(mode string) {
	// 如果 Neovim 进入插入模式或可视模式，退出 FSM
	if mode == "i" || mode == "v" || mode == "V" || strings.HasPrefix(mode, "s") {
		ExitFSM()
	}
}

// NotifyNvimMode 通知 Neovim 当前 FSM 模式
// 注意：这个函数 currently would need to use intents to communicate
// with the backend, but Phase-3 requires that FSM doesn't directly call backend
func NotifyNvimMode() {
	// Phase-3 invariant: FSM does not directly call backend
	// This functionality should be handled by Kernel/Weaver layer
	// using intents to communicate with the backend
}