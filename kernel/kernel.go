package kernel

import (
	"context"
	"log"
	"tmux-fsm/fsm"
	"tmux-fsm/intent/builder"
	"tmux-fsm/planner"
)

type Kernel struct {
	FSM           *fsm.Engine
	Grammar       *planner.Grammar
	Exec          IntentExecutor
	NativeBuilder *builder.CompositeBuilder
	ShadowIntent  bool
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

	// 通过legacy路径生成intent（权威执行路径）
	legacyDecision := k.Decide(key)

	// 如果启用了shadow intent，同时生成native intent进行对比
	if k.ShadowIntent && k.NativeBuilder != nil {
		nativeIntent, ok := k.NativeBuilder.Build(builder.BuildContext{
			Action: key,
			Count:  1, // 默认计数
		})

		if ok && nativeIntent != nil && legacyDecision != nil && legacyDecision.Intent != nil {
			// 比较native和legacy intent的语义
			if !builder.SemanticEqual(nativeIntent, legacyDecision.Intent) {
				log.Printf(
					"[INTENT MISMATCH]\nlegacy=%+v\nnative=%+v\n",
					legacyDecision.Intent,
					nativeIntent,
				)
			}
		} else if ok && nativeIntent != nil && legacyDecision == nil {
			// native intent生成成功，但legacy没有intent
			log.Printf(
				"[INTENT MISSING] legacy did not generate intent for action=%s, native=%+v",
				key,
				nativeIntent,
			)
		} else if !ok {
			// native intent生成失败
			log.Printf(
				"[INTENT MISSING] native builder did not handle action=%s",
				key,
			)
		}
	}

	// 只执行legacy intent（当前阶段）
	if legacyDecision != nil {
		k.Execute(legacyDecision)
	}
}

