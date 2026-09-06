package fsm

import (
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
//
// M2.4 修复（双 UI 写入者合并）：本函数曾直接写 tmux 变量
// （updateTmuxVariables：`NAV [delete]` 风格），与 globals.go 的
// updateStatusBar（`PENDING [delete] [3]` 风格）无锁争写 @fsm_state，
// 两者格式互斥且顺序不定 → 状态栏随机漂移。
//
// 现在的唯一写入者是 main.go 每键路径上的 updateStatusBar
// （经 GlobalBackend 统一通道，格式更完整）。本函数仅保留
// OnUpdateUI 回调钩子供外部架构联动；FSM 层不再直接触碰 tmux
// ——兑现本文件自己的 TODO："FSM should NOT directly touch tmux"。
// 顺带消除每键 3 次 exec 的性能负担（M3.1 的一部分）。
func UpdateUI(_ ...any) {
	// 调用外部注册的UI更新回调
	if OnUpdateUI != nil {
		OnUpdateUI()
	}
}

// HideUI 隐藏UI
//
// 退出时清空 tmux 变量——这是唯一允许 FSM 层直接写 tmux 变量的场景
// （退出清理，无漂移风险：此时状态机已停止）。
func HideUI() {
	setTmuxOption("@fsm_state", "")
	setTmuxOption("@fsm_keys", "")
	refreshTmuxClient()
}

// setTmuxOption 设置 tmux 选项（仅供 HideUI 退出清理使用）
func setTmuxOption(option, value string) {
	cmd := exec.Command("tmux", "set", "-g", option, value)
	_ = cmd.Run()
}

// refreshTmuxClient 刷新 tmux 客户端（仅供 HideUI 使用）
func refreshTmuxClient() {
	cmd := exec.Command("tmux", "refresh-client", "-S")
	_ = cmd.Run()
}
