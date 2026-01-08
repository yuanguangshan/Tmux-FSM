package kernel

func (k *Kernel) Execute(decision *Decision) {
	if decision == nil {
		panic("kernel.Execute called with nil decision")
	}

	if k.Exec == nil {
		return
	}

	switch decision.Kind {
	case DecisionNone:
		return // 刻意不作为

	case DecisionFSM:
		if decision.Intent == nil {
			panic("FSM decision without intent")
		}
		_ = k.Exec.Process(decision.Intent)

	case DecisionLegacy:
		if decision.Intent == nil {
			panic("Legacy decision without intent")
		}
		_ = k.Exec.Process(decision.Intent)

	default:
		panic("unknown decision kind")
	}
}
