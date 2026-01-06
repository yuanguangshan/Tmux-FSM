package kernel

import (
	"tmux-fsm/intent"
)

func (k *Kernel) Execute(decision *Decision) {
	if decision == nil || decision.Intent == nil {
		return
	}

	if k.Exec == nil {
		return
	}

	switch decision.Kind {
	case DecisionFSM:
		_ = k.Exec.Process(decision.Intent)
	case DecisionLegacy:
		_ = k.Exec.Process(decision.Intent)
	}
}
