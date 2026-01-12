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
	DecisionIntent
)

func (k DecisionKind) String() string {
	switch k {
	case DecisionNone:
		return "None"
	case DecisionFSM:
		return "FSM"
	case DecisionLegacy:
		return "Legacy"
	case DecisionIntent:
		return "Intent"
	default:
		return "Unknown"
	}
}

type Decision struct {
	Kind   DecisionKind
	Intent *intent.Intent
	Action string // For simple FSM actions
}

// GrammarEmitter 用于将 Grammar 的结果传递给 Kernel
type GrammarEmitter struct {
	grammar  *planner.Grammar
	callback func(*intent.GrammarIntent)
}

func (g *GrammarEmitter) Emit(token fsm.RawToken) {
	grammarIntent := g.grammar.Consume(token)
	if grammarIntent != nil && g.callback != nil {
		g.callback(grammarIntent)
	}
}

func (k *Kernel) Decide(key string) *Decision {
	// ✅ 1. 优先检查是否有简单的 FSM 动作（最高优先级）
	if k.FSM != nil {
		if k.FSM.CanHandle(key) {
			if state, ok := k.FSM.Keymap.States[k.FSM.Active]; ok {
				if keyAction, ok := state.Keys[key]; ok && keyAction.Action != "" {
					// 这是一个简单的 FSM 动作，优先执行
					return &Decision{
						Kind:   DecisionFSM,
						Action: keyAction.Action,
					}
				}
			}
		}

		// ✅ 2. 如果没有简单的 FSM 动作，再让 Grammar 处理
		var lastGrammarIntent *intent.GrammarIntent

		// 创建一个 GrammarEmitter 来处理 token
		grammarEmitter := &GrammarEmitter{
			grammar: k.Grammar,
			callback: func(grammarIntent *intent.GrammarIntent) {
				lastGrammarIntent = grammarIntent
			},
		}

		// 添加 GrammarEmitter 到 FSM
		k.FSM.AddEmitter(grammarEmitter)

		// 让 FSM 处理按键，这会生成 token
		_, dispatched := k.FSM.Dispatch(key)

		// 同步 Grammar 的 PendingOperator 到 FSM (用于 UI 显示)
		if k.Grammar != nil {
			k.FSM.PendingOperator = k.Grammar.GetPendingOp()
		}

		// 刷新 UI
		fsm.UpdateUI()

		// 移除 GrammarEmitter
		k.FSM.RemoveEmitter(grammarEmitter)

		if dispatched && lastGrammarIntent != nil {
			// 将 GrammarIntent 提升为 Intent
			finalIntent := intent.Promote(lastGrammarIntent)

			// 返回意图供执行
			return &Decision{
				Kind:   DecisionIntent, // This is a full-fledged intent
				Intent: finalIntent,
			}
		}

		if dispatched {
			// ✅ 合法状态：key 被 FSM 吃了，但 Grammar 没有生成意图
			// 这是正常情况，例如在等待更多按键时
			return &Decision{
				Kind: DecisionNone, // FSM 吃了，但还没决定
			}
		}
	}

	// 没有 FSM 处理，明确返回 Legacy 决策
	return &Decision{
		Kind: DecisionLegacy,
	}
}

// GetPendingOp 获取当前处于 pending 状态的操作符名称
func (k *Kernel) GetPendingOp() string {
	if k.Grammar != nil {
		return k.Grammar.GetPendingOp()
	}
	return ""
}

// GetCount 获取当前 FSM 计数
func (k *Kernel) GetCount() int {
	if k.FSM != nil {
		return k.FSM.GetCount()
	}
	return 0
}
