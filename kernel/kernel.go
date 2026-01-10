package kernel

import (
	"context"
	"fmt"
	"log"
	"time"
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
	Ctx       context.Context
	RequestID string // Unique identifier for this user request
	ActorID   string // User / pane / client identifier
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
	// ⚠️ Invariant: RequestID / ActorID are authoritative once received.
	// Server MUST NOT generate or modify them.
	requestID := hctx.RequestID
	if requestID == "" {
		log.Printf("[FATAL] missing RequestID at Kernel boundary")
		return
	}

	actorID := hctx.ActorID
	if actorID == "" {
		log.Printf("[FATAL] missing ActorID at Kernel boundary")
		return
	}

	// Log the incoming key for audit trail with identity anchors
	log.Printf("Handling key: RequestID=%s, ActorID=%s, Key=%s", requestID, actorID, key)

	// 通过Grammar路径生成intent（新的权威执行路径）
	var decision *Decision

	// 先尝试通过FSM + Grammar生成intent
	if k.FSM != nil && k.Grammar != nil {
		decision = k.Decide(key)

		if decision != nil {
			// Log decision details for audit trail
			log.Printf("Decision made for key '%s': RequestID=%s, ActorID=%s, Kind=%s, Intent=%v",
				key, requestID, actorID, decision.Kind, decision.Intent)

			switch decision.Kind {
			case DecisionIntent:
				log.Printf("Processing intent for key '%s': RequestID=%s, ActorID=%s", key, requestID, actorID)
				k.ProcessIntentWithContext(hctx, decision.Intent)
				return

			case DecisionFSM:
				log.Printf("Executing FSM decision for key '%s': RequestID=%s, ActorID=%s", key, requestID, actorID)
				k.Execute(decision)
				return

			case DecisionNone:
				// FSM 吃了 key，合法等待
				log.Printf("FSM consumed key '%s', valid wait state: RequestID=%s, ActorID=%s", key, requestID, actorID)
				return

			case DecisionLegacy:
				// 明确：Grammar/FSM 不处理，才允许 legacy
				log.Printf("Key '%s' falls back to legacy handling: RequestID=%s, ActorID=%s", key, requestID, actorID)
				break
			}
		}
	}

	// 如果Grammar没有处理，记录信息（未来将完全移除legacy路径）
	if k.ShadowIntent && k.NativeBuilder != nil {
		// 只有在 DecisionLegacy 情况下才记录为未覆盖
		// DecisionNone 是合法的等待状态，不应计入未覆盖
		if decision != nil && decision.Kind == DecisionLegacy {
			log.Printf("[GRAMMAR COVERAGE] key '%s' not handled by Grammar: RequestID=%s, ActorID=%s", key, requestID, actorID)
			k.ShadowStats.Total++
			k.ShadowStats.Mismatched++ // 记录为未覆盖
		}
	}
}

// ProcessIntent 处理意图
func (k *Kernel) ProcessIntent(intent *intent.Intent) error {
	// Create a default context with generated IDs for backward compatibility
	hctx := HandleContext{
		Ctx:       context.Background(),
		RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()),
		ActorID:   "unknown",
	}
	return k.ProcessIntentWithContext(hctx, intent)
}

// ProcessIntentWithContext 处理意图 with context containing identity anchors
func (k *Kernel) ProcessIntentWithContext(hctx HandleContext, intent *intent.Intent) error {
	if intent == nil {
		log.Printf("ProcessIntent called with nil intent: RequestID=%s, ActorID=%s", hctx.RequestID, hctx.ActorID)
		return fmt.Errorf("intent is nil")
	}

	// Log intent details for audit trail with identity anchors
	log.Printf("Processing intent: RequestID=%s, ActorID=%s, Kind=%d, PaneID=%s",
		hctx.RequestID, hctx.ActorID, intent.Kind, intent.PaneID)

	if k.Exec != nil {
		log.Printf("Processing intent through external executor: RequestID=%s, ActorID=%s", hctx.RequestID, hctx.ActorID)

		// Check if executor supports contextual processing
		if ctxExec, ok := k.Exec.(ContextualIntentExecutor); ok {
			err := ctxExec.ProcessWithContext(hctx.Ctx, hctx, intent)
			if err != nil {
				log.Printf("Contextual intent execution failed: RequestID=%s, ActorID=%s, Error=%v", hctx.RequestID, hctx.ActorID, err)
				return err
			}
			log.Printf("Intent processed successfully by contextual external executor: RequestID=%s, ActorID=%s", hctx.RequestID, hctx.ActorID)
			return nil
		} else {
			// Fallback to non-contextual processing
			err := k.Exec.Process(intent)
			if err != nil {
				log.Printf("Intent execution failed: RequestID=%s, ActorID=%s, Error=%v", hctx.RequestID, hctx.ActorID, err)
				return err
			}
			log.Printf("Intent processed successfully by external executor: RequestID=%s, ActorID=%s", hctx.RequestID, hctx.ActorID)
			return nil
		}
	}

	// 如果没有外部执行器，尝试通过FSM执行意图
	if k.FSM != nil {
		log.Printf("Processing intent through FSM: RequestID=%s, ActorID=%s", hctx.RequestID, hctx.ActorID)
		err := k.FSM.DispatchIntent(intent)
		if err != nil {
			log.Printf("FSM dispatch failed: RequestID=%s, ActorID=%s, Error=%v", hctx.RequestID, hctx.ActorID, err)
			return err
		}
		log.Printf("Intent dispatched successfully through FSM: RequestID=%s, ActorID=%s", hctx.RequestID, hctx.ActorID)
		return nil
	}

	log.Printf("No executor available for intent: RequestID=%s, ActorID=%s, Intent=%v", hctx.RequestID, hctx.ActorID, intent)
	return fmt.Errorf("no executor available for intent")
}
