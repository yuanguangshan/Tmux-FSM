package fsm

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
	// Phase‑3 invariant:
	// FSM does NOT touch UI / backend directly.
	// UI update must be handled by Kernel / Weaver.
}

// HideUI 隐藏UI
func HideUI() {
	// Phase‑3 invariant:
	// FSM does NOT touch UI / backend directly.
	// UI update must be handled by Kernel / Weaver.
}