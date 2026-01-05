package fsm

import (
	"strings"
	tmux_fsm "tmux-fsm"
)

// OnNvimMode 处理来自 Neovim 的模式变化
func OnNvimMode(mode string) {
	// 如果 Neovim 进入插入模式或可视模式，退出 FSM
	if mode == "i" || mode == "v" || mode == "V" || strings.HasPrefix(mode, "s") {
		ExitFSM()
	}
}

// NotifyNvimMode 通知 Neovim 当前 FSM 模式
// 注意：这个函数目前使用 send-keys，这可能不是最佳方案
// 更好的方案是使用 Neovim 的 RPC 机制
func NotifyNvimMode() {
	// 获取当前活跃的窗口/面板
	out, err := tmux_fsm.GlobalBackend.GetCommandOutput("display-message -p '#{pane_current_command}'")
	if err != nil {
		return
	}

	cmd := strings.TrimSpace(out)
	// 如果当前面板是 vim/nvim，则可以考虑发送模式信息
	// 但目前我们不执行任何操作，避免干扰用户输入
	// 更好的方式是通过 Neovim server/client 机制通信
	if cmd == "vim" || cmd == "nvim" {
		// TODO: 实现 Neovim RPC 通信以同步状态
		// 例如: nvim --server <socket> --remote-expr "FSM_SetMode('NAV')"
	}
}