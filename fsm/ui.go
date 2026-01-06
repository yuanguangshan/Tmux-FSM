package ui

// UI 接口
type UI interface {
	Show()
	Update()
	Hide()
}

// StateProvider 由 FSM 实现
type StateProvider interface {
	GetActiveState() string
	GetStateHint(state string) string
}

// Backend 由 tmux backend 实现
type Backend interface {
	ExecRaw(cmd string)
}
