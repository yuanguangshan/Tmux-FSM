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

// IntentExecutor 定义意图执行器接口
type IntentExecutor interface {
	Execute(*Decision) error
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

// Execute 执行决策
func (k *Kernel) Execute(decision *Decision) {
	if decision == nil {
		return
	}

	if k.Exec != nil {
		// 使用外部执行器
		k.Exec.Execute(decision)
		return
	}

	// 如果没有外部执行器，尝试通过FSM执行意图
	if k.FSM != nil && decision.Intent != nil {
		// 通过FSM的DispatchIntent方法执行意图
		_ = k.FSM.DispatchIntent(decision.Intent)
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

		// 如果Grammar成功生成了intent，直接执行
		if decision != nil && decision.Intent != nil {
			k.Execute(decision)
			return
		}
	}

	// 如果Grammar没有处理，记录信息（未来将完全移除legacy路径）
	if k.ShadowIntent && k.NativeBuilder != nil {
		// 记录未被Grammar处理的按键
		log.Printf("[GRAMMAR COVERAGE] key '%s' not handled by Grammar", key)
		k.ShadowStats.Total++
		k.ShadowStats.Mismatched++ // 记录为未覆盖
	}
}

