package fsm

import (
	"tmux-fsm/backend"
)

// UIDriver 定义UI驱动接口
type UIDriver interface {
	SetUserOption(option, value string) error
	RefreshClient(clientName string) error
}

var uiDriver UIDriver

// SetUIDriver 设置UI驱动实现
func SetUIDriver(driver UIDriver) {
	uiDriver = driver
}

// UpdateUI 更新UI显示当前FSM状态（Invariant 9: UI 派生状态）
func UpdateUI(_ ...any) {
	if defaultEngine == nil {
		return
	}

	// 获取当前活跃层
	activeLayer := defaultEngine.Active
	if activeLayer == "" {
		activeLayer = "NAV"
	}

	// 获取当前层的提示信息
	hint := ""
	if stateDef, ok := KM.States[activeLayer]; ok {
		hint = stateDef.Hint
	}

	// 优先使用设置的UI驱动，否则使用全局backend
	if uiDriver != nil {
		uiDriver.SetUserOption("@fsm_state", activeLayer)
		uiDriver.SetUserOption("@fsm_keys", hint)
		uiDriver.RefreshClient("")
	} else if backend.GlobalBackend != nil {
		backend.GlobalBackend.SetUserOption("@fsm_state", activeLayer)
		backend.GlobalBackend.SetUserOption("@fsm_keys", hint)
		backend.GlobalBackend.RefreshClient("")
	}
}

// HideUI 隐藏UI
func HideUI() {
	if uiDriver != nil {
		uiDriver.SetUserOption("@fsm_state", "")
		uiDriver.SetUserOption("@fsm_keys", "")
		uiDriver.RefreshClient("")
	} else if backend.GlobalBackend != nil {
		backend.GlobalBackend.SetUserOption("@fsm_state", "")
		backend.GlobalBackend.SetUserOption("@fsm_keys", "")
		backend.GlobalBackend.RefreshClient("")
	}
}