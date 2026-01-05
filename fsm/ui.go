package fsm

import "tmux-fsm/fsm/ui"

// UIManager UI 管理器
type UIManager struct {
	active ui.UI
}

// NewUIManager 创建新的 UI 管理器
func NewUIManager() *UIManager {
	return &UIManager{}
}

// 全局 UI 实例
var CurrentUI ui.UI

// OnUpdateUI 是当 FSM 状态变化时需要执行的回调（通常由主程序注入以更新状态栏）
var OnUpdateUI func()

// UI 更新函数
func ShowUI() {
	if CurrentUI != nil {
		CurrentUI.Show()
	}
	if OnUpdateUI != nil {
		OnUpdateUI()
	}
}

func UpdateUI() {
	if CurrentUI != nil {
		CurrentUI.Update()
	}
	if OnUpdateUI != nil {
		OnUpdateUI()
	}
}

func HideUI() {
	if CurrentUI != nil {
		CurrentUI.Hide()
	}
	// 隐藏后通常也需要刷新一次状态栏以移除文字
	if OnUpdateUI != nil {
		OnUpdateUI()
	}
}