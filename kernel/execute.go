package kernel

func (k *Kernel) Execute(decision *Decision) {
	if decision == nil {
		return
	}

	// switch decision.Kind {
	// case DecisionFSM:
	// 	ExecuteIntent(decision.Intent)
	// case DecisionLegacy:
	// 	ExecuteIntent(decision.Intent)
	// }

	// Temporarily disabled until interfaces are properly defined
}
