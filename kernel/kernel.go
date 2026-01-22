package kernel

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"tmux-fsm/backend"
	"tmux-fsm/fsm"
	"tmux-fsm/intent"
	"tmux-fsm/intent/builder"
	"tmux-fsm/planner"
)

// ShadowStats records statistics for shadow intent comparison.
// NOTE: ShadowStats is not concurrency-safe.
// Kernel.HandleKey must be serialized.
//
// Phase-5 Lifecycle Strategy:
// - Current: Stats grow indefinitely (daemon lifetime)
// - Future reset points (choose one):
//   - fsm.Reload() - reset on config reload
//   - fsm.EnterFSM()/ExitFSM() - reset on mode entry/exit
//   - __SHADOW_RESET__ command - explicit reset via server protocol
//
// Semantics:
// - Total: All keys processed
// - Built: Grammar produced an Intent (DecisionIntent)
// - Mismatched: Grammar didn't cover key (DecisionLegacy)
// - Matched: Reserved for future shadow comparison logic
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

	log.Printf("Handling key: RequestID=%s, ActorID=%s, Key=%s", requestID, actorID, key)

	decision := k.Decide(key)
	k.Execute(decision)

	// --- Shadow Intent Coverage Stats ---
	if k.ShadowIntent {
		k.ShadowStats.Total++

		if decision != nil && decision.Kind == DecisionIntent {
			k.ShadowStats.Built++
		}

		if decision != nil && decision.Kind == DecisionLegacy {
			k.ShadowStats.Mismatched++

			log.Printf(
				"[SHADOW] Legacy key not covered by Grammar: key=%q, actor=%s, total=%d, legacy=%d",
				key,
				actorID,
				k.ShadowStats.Total,
				k.ShadowStats.Mismatched,
			)
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

	// Inject PaneID if not already set (Grammar never produces PaneID)
	if intent.PaneID == "" && hctx.ActorID != "" {
		// ActorID format is "paneID|clientName", extract paneID
		parts := strings.SplitN(hctx.ActorID, "|", 2)
		intent.PaneID = parts[0]
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

func (k *Kernel) executeAction(action string) {
	log.Printf("Executing action: %s", action)
	switch action {
	case "pane_left":
		tmux("select-pane -L")
	case "pane_right":
		tmux("select-pane -R")
	case "pane_up":
		tmux("select-pane -U")
	case "pane_down":
		tmux("select-pane -D")
	case "next_pane":
		tmux("select-pane -t :.+")
	case "prev_pane":
		tmux("select-pane -t :.-")
	case "far_left":
		tmux("select-pane -t :.0")
	case "far_right":
		tmux("select-pane -t :.$")
	case "goto_top":
		tmux("select-pane -t :.0")
	case "goto_bottom":
		tmux("select-pane -t :.$")
	case "goto_line_start":
		tmux("send-keys -t . Home")
	case "goto_line_end":
		tmux("send-keys -t . End")
	case "move_left":
		tmux("send-keys -t . Left")
	case "move_right":
		tmux("send-keys -t . Right")
	case "move_up":
		tmux("send-keys -t . Up")
	case "move_down":
		tmux("send-keys -t . Down")
	case "exit":
		fsm.ExitFSM()
	case "prompt":
		tmux("command-prompt")
	case "repeat":
		// This will be handled by a proper implementation of repeat later
		log.Println("Repeat action is not yet implemented.")
	default:
		log.Printf("Unknown action: %s", action)
	}
}

func tmux(cmd string) {
	err := backend.GlobalBackend.ExecRaw(cmd)
	if err != nil {
		log.Printf("Error executing tmux command '%s': %v", cmd, err)
	}
}
