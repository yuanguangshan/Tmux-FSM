package ui

import (
	"os/exec"
	"fmt"
)

// StateProvider 接口用于获取状态信息
type StateProvider interface {
	GetActiveState() string
	GetStateHint(state string) string
}

// PopupUI 实现 UI 接口
type PopupUI struct {
	StateProvider StateProvider
}

func (p *PopupUI) Show() {
	if p.StateProvider == nil {
		return
	}

	active := p.StateProvider.GetActiveState()
	hint := p.StateProvider.GetStateHint(active)

	// 如果状态为空，不显示弹窗
	if active == "" {
		return
	}

	cmd := exec.Command("tmux", "display-popup",
		"-E",
		"-w", "50%",
		"-h", "5",
		fmt.Sprintf("echo '%s'; echo '%s'", active, hint),
	)
	cmd.Run()
}

func (p *PopupUI) Update() {
	// 重新显示内容
	p.Show()
}

func (p *PopupUI) Hide() {
	exec.Command("tmux", "display-popup", "-C").Run()
}