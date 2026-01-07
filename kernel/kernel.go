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

	// 通过legacy路径生成intent（权威执行路径）
	legacyDecision := k.Decide(key)

	// 如果启用了shadow intent，同时生成native intent进行对比
	if k.ShadowIntent && k.NativeBuilder != nil {
		// 从legacy decision中提取上下文信息
		var legacyIntent *intent.Intent
		if legacyDecision != nil {
			legacyIntent = legacyDecision.Intent
		}

		k.ShadowStats.Total++

		if legacyIntent != nil {
			ctx := builder.BuildContext{
				Action:       key,
				Count:        legacyIntent.Count,
				PaneID:       legacyIntent.PaneID,
				SnapshotHash: legacyIntent.SnapshotHash,
			}

			nativeIntent, ok := k.NativeBuilder.Build(ctx)

			if ok {
				k.ShadowStats.Built++
			}

			if ok && nativeIntent != nil {
				// 比较native和legacy intent的语义
				if !builder.SemanticEqual(nativeIntent, legacyIntent, builder.CompareMigration) {
					diffs := builder.DiffIntent(legacyIntent, nativeIntent)
					log.Printf("[INTENT MISMATCH] action=%s diffs=%+v", key, diffs)
					k.ShadowStats.Mismatched++
				} else {
					k.ShadowStats.Matched++
				}
			} else if ok {
				// native intent生成失败
				log.Printf(
					"[INTENT MISSING] native builder did not handle action=%s",
					key,
				)
				k.ShadowStats.Mismatched++
			}
		} else {
			// legacy intent为空，尝试构建native intent
			ctx := builder.BuildContext{
				Action: key,
				Count:  1, // 默认计数
			}

			nativeIntent, ok := k.NativeBuilder.Build(ctx)
			if ok && nativeIntent != nil {
				// native intent生成成功，但legacy没有intent
				log.Printf(
					"[INTENT MISSING] legacy did not generate intent for action=%s, native=%+v",
					key,
					nativeIntent,
				)
				k.ShadowStats.Mismatched++
			} else if !ok {
				// native intent生成失败
				log.Printf(
					"[INTENT MISSING] native builder did not handle action=%s",
					key,
				)
				k.ShadowStats.Mismatched++
			}
		}
	}

	// 只执行legacy intent（当前阶段）
	if legacyDecision != nil {
		k.Execute(legacyDecision)
	}
}

