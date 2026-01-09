package fsm

import (
	"fmt"
	"os/exec"
)

// UIDriver 定义UI驱动接口
type UIDriver interface {
	SetUserOption(option, value string) error
	RefreshClient(clientName string) error
}

var uiDriver UIDriver

// OnUpdateUI 当UI需要更新时调用的回调函数
var OnUpdateUI func()

// SetUIDriver 设置UI驱动实现
func SetUIDriver(driver UIDriver) {
	uiDriver = driver
}

// UpdateUI 更新UI显示当前FSM状态（Invariant 9: UI 派生状态）
func UpdateUI(_ ...any) {
	// TEMPORARY: debug-only UI bridge
	// This is a technical debt - FSM should NOT directly touch tmux
	// TODO: Move to Kernel → Weaver → Backend pipeline
	updateTmuxVariables()

	// 调用外部注册的UI更新回调
	if OnUpdateUI != nil {
		OnUpdateUI()
	}
}

// updateTmuxVariables 更新 tmux 状态变量
func updateTmuxVariables() {
	if defaultEngine == nil {
		return
	}

	// 更新状态变量
	activeLayer := defaultEngine.Active
	if activeLayer == "" {
		activeLayer = "NAV"
	}

	// 设置状态变量
	setTmuxOption("@fsm_state", activeLayer)

	// 如果有计数器，也显示它
	if defaultEngine.count > 0 {
		setTmuxOption("@fsm_keys", fmt.Sprintf("%d", defaultEngine.count))
	} else {
		setTmuxOption("@fsm_keys", "")
	}

	// 刷新客户端以更新状态栏
	refreshTmuxClient()
}

// setTmuxOption 设置 tmux 选项
func setTmuxOption(option, value string) {
	cmd := exec.Command("tmux", "set", "-g", option, value)
	_ = cmd.Run()
}

// refreshTmuxClient 刷新 tmux 客户端
func refreshTmuxClient() {
	cmd := exec.Command("tmux", "refresh-client", "-S")
	_ = cmd.Run()
}

// HideUI 隐藏UI
func HideUI() {
	// Phase‑3 invariant:
	// FSM does NOT touch UI / backend directly.
	// UI update must be handled by Kernel / Weaver.
	// 但是，为了隐藏状态，我们需要重置 tmux 变量
	setTmuxOption("@fsm_state", "")
	setTmuxOption("@fsm_keys", "")
	refreshTmuxClient()
}
