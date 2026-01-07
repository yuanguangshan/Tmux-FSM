package kernel

import (
	"tmux-fsm/fsm"
	"tmux-fsm/intent"
	"tmux-fsm/planner"
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

// GrammarEmitter 用于将 Grammar 的结果传递给 Kernel
type GrammarEmitter struct {
	grammar *planner.Grammar
	callback func(*intent.Intent)
}

func (g *GrammarEmitter) Emit(token fsm.RawToken) {
	intent := g.grammar.Consume(token)
	if intent != nil && g.callback != nil {
		g.callback(intent)
	}
}

func (k *Kernel) Decide(key string) *Decision {
	// ✅ 1. FSM 永远先拿 key
	if k.FSM != nil {
		var lastIntent *intent.Intent

		// 创建一个 GrammarEmitter 来处理 token
		grammarEmitter := &GrammarEmitter{
			grammar: k.Grammar,
			callback: func(intent *intent.Intent) {
				lastIntent = intent
			},
		}

		// 添加 GrammarEmitter 到 FSM
		k.FSM.AddEmitter(grammarEmitter)

		// 让 FSM 处理按键
		dispatched := k.FSM.Dispatch(key)

		// 移除 GrammarEmitter
		k.FSM.RemoveEmitter(grammarEmitter)

		if dispatched && lastIntent != nil {
			// 直接执行意图，而不是返回决策
			if k.FSM != nil {
				_ = k.FSM.DispatchIntent(lastIntent)
			}
			return nil // 意图已直接执行
		}

		if dispatched {
			return nil // FSM处理了按键，但没有产生意图
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
