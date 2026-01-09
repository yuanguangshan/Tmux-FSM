package kernel

import (
	"log"
)

// Execute a decision made by the kernel.
func (k *Kernel) Execute(decision *Decision) {
	if decision == nil {
		log.Println("kernel.Execute called with nil decision")
		return
	}

	if k.Exec == nil {
		log.Println("kernel.Execute called with nil executor")
		return
	}

	switch decision.Kind {
	case DecisionNone, DecisionLegacy:
		return // Do nothing intentionally.

	case DecisionIntent:
		// This is a full-fledged intent from the grammar.
		// Process it via the standard execution path.
		if decision.Intent == nil {
			log.Println("DecisionIntent without an intent")
			return
		}
		_ = k.Exec.Process(decision.Intent)

	default:
		log.Printf("Unknown or unhandled decision kind: %v", decision.Kind)
	}
}
