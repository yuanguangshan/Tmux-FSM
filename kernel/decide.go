package kernel

import (
	"tmux-fsm/intent"
)

type DecisionKind int

const (
	DecisionNone DecisionKind = iota
	DecisionFSM
	DecisionLegacy
)

type Decision struct {
	Kind   DecisionKind
	Intent *intent.Intent
}

func (k *Kernel) Decide(key string) *Decision {
	// ✅ 1. FSM 永远先拿 key
	if k.FSM != nil {
		intent, ok := k.FSM.Produce(key)
		if ok && intent != nil {
			return &Decision{
				Kind:   DecisionFSM,
				Intent: intent,
			}
		}
		// 如果FSM明确处理了但不产生意图，说明是层切换等操作
		if k.FSM.InLayer() && k.FSM.CanHandle(key) {
			return nil // FSM吞掉按键，不产生决策
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
