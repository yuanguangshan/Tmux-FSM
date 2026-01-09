package kernel

import (
	"context"
	"log"
	"tmux-fsm/fsm"
	"tmux-fsm/intent"
	"tmux-fsm/intent/builder"
	"tmux-fsm/planner"
)

// ShadowStats records statistics for shadow intent comparison.
// NOTE: ShadowStats is not concurrency-safe.
// Kernel.HandleKey must be serialized.
type ShadowStats struct {
	Total      int
	Built      int
	Matched    int
	Mismatched int
}

type Kernel struct {
	FSM           *fsm.Engine
	Grammar       *planner.Grammar
	Exec          IntentExecutor
	NativeBuilder *builder.CompositeBuilder
	ShadowIntent  bool
	ShadowStats   ShadowStats
}

// ✅ Kernel 的唯一上下文入口（现在先很薄，未来可扩展）
type HandleContext struct {
	Ctx context.Context
}


func NewKernel(fsmEngine *fsm.Engine, exec IntentExecutor) *Kernel {
	return &Kernel{
		FSM:           fsmEngine,
		Grammar:       planner.NewGrammar(),
		Exec:          exec,
		NativeBuilder: builder.NewCompositeBuilder(),
		ShadowIntent:  true,
	}
}


// ✅ Kernel 的唯一入口
func (k *Kernel) HandleKey(hctx HandleContext, key string) {
	_ = hctx // ✅ 现在不用，但接口已经锁死

	// 通过Grammar路径生成intent（新的权威执行路径）
	var decision *Decision

	// 先尝试通过FSM + Grammar生成intent
	if k.FSM != nil && k.Grammar != nil {
		decision = k.Decide(key)

		if decision != nil {
			switch decision.Kind {
			case DecisionFSM:
				k.Execute(decision)
				return

			case DecisionNone:
				// FSM 吃了 key，合法等待
				return

			case DecisionLegacy:
				// 明确：Grammar/FSM 不处理，才允许 legacy
				break
			}
		}
	}

	// 如果Grammar没有处理，记录信息（未来将完全移除legacy路径）
	if k.ShadowIntent && k.NativeBuilder != nil {
		// 只有在 DecisionLegacy 情况下才记录为未覆盖
		// DecisionNone 是合法的等待状态，不应计入未覆盖
		if decision != nil && decision.Kind == DecisionLegacy {
			log.Printf("[GRAMMAR COVERAGE] key '%s' not handled by Grammar", key)
			k.ShadowStats.Total++
			k.ShadowStats.Mismatched++ // 记录为未覆盖
		}
	}
}

// ProcessIntent 处理意图
func (k *Kernel) ProcessIntent(intent *intent.Intent) error {
	if k.Exec != nil {
		return k.Exec.Process(intent)
	}

	// 如果没有外部执行器，尝试通过FSM执行意图
	if k.FSM != nil && intent != nil {
		return k.FSM.DispatchIntent(intent)
	}

	return nil
}

