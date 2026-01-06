package kernel

type DecisionKind int

const (
	DecisionNone DecisionKind = iota
	DecisionFSM
	DecisionLegacy
)

type Decision struct {
	Kind DecisionKind
	// Intent *intent.Intent  // Temporarily disabled
}

func (k *Kernel) Decide(key string) *Decision {
	// ✅ 1. FSM 永远先拿 key
	if k.FSM != nil {
		if k.FSM.InLayer() && k.FSM.CanHandle(key) {
			handled := k.FSM.Dispatch(key)
			if handled {
				return &Decision{
					Kind: DecisionFSM,
					// Intent: intent,  // Temporarily disabled
				}
			}
			// FSM 明确吞掉
			return nil
		}
	}

	// ✅ 2. Legacy decoder（复用你现有逻辑）
	// legacyIntent := DecodeLegacyKey(key)  // Temporarily disabled
	// if legacyIntent != nil {
	// 	return &Decision{
	// 		Kind:   DecisionLegacy,
	// 		Intent: legacyIntent,
	// 	}
	// }

	return nil
}
