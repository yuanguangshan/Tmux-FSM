package ui

import "fmt"

type Backend interface {
	ExecRaw(cmd string)
}

type StateProvider interface {
	GetActiveState() string
	GetStateHint(state string) string
}

type PopupUI struct {
	StateProvider StateProvider
	Backend       Backend
}

func (p *PopupUI) Show() {
	if p.StateProvider == nil || p.Backend == nil {
		return
	}

	active := p.StateProvider.GetActiveState()
	if active == "" {
		return
	}

	hint := p.StateProvider.GetStateHint(active)

	cmd := fmt.Sprintf(
		"display-popup -E -w 50%% -h 5 'echo \"%s\"; echo \"%s\"'",
		active,
		hint,
	)

	p.Backend.ExecRaw(cmd)
}

func (p *PopupUI) Update() {
	p.Show()
}

func (p *PopupUI) Hide() {
	if p.Backend != nil {
		p.Backend.ExecRaw("display-popup -C")
	}
}
