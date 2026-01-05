package kernel

import (
    "yourmodule/intent"
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
        if k.FSM.InLayer() && k.FSM.CanHandle(key) {
            intent := k.FSM.Dispatch(key)
            if intent != nil {
                return &Decision{
                    Kind:   DecisionFSM,
                    Intent: intent,
                }
            }
            // FSM 明确吞掉
            return nil
        }
    }

    // ✅ 2. Legacy decoder（复用你现有逻辑）
    legacyIntent := DecodeLegacyKey(key)
    if legacyIntent != nil {
        return &Decision{
            Kind:   DecisionLegacy,
            Intent: legacyIntent,
        }
    }

    return nil
}
