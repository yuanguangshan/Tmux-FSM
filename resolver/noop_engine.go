package resolver

import "tmux-fsm/intent"

// NoopEngine 空操作引擎实现
type NoopEngine struct{}

func (n *NoopEngine) SendKeys(keys ...string) {}

func (n *NoopEngine) GetVisualMode() intent.VisualMode {
	return intent.VisualModeNormal
}

func (n *NoopEngine) EnterVisualMode(mode intent.VisualMode) {}

func (n *NoopEngine) ExitVisualMode() {}

func (n *NoopEngine) GetCurrentCursor() ResolverCursor {
	return ResolverCursor{}
}

func (n *NoopEngine) ComputeMotion(m *intent.Motion) (ResolverRange, error) {
	return ResolverRange{}, nil
}

func (n *NoopEngine) MoveCursor(r ResolverRange) error {
	return nil
}

func (n *NoopEngine) DeleteRange(r ResolverRange) error {
	return nil
}

func (n *NoopEngine) YankRange(r ResolverRange) error {
	return nil
}

func (n *NoopEngine) ChangeRange(r ResolverRange) error {
	return nil
}