package kernel

import (
	"context"
	"log"
	"strings"
	"tmux-fsm/intent"
	"tmux-fsm/weaver/core"
	"tmux-fsm/weaver/manager"
)

// ResolverExecutor is the executor that forwards intents to the Weaver system.
type ResolverExecutor struct{}

// NewResolverExecutor creates a new ResolverExecutor.
func NewResolverExecutor() *ResolverExecutor {
	return &ResolverExecutor{}
}

// Process an intent by adapting it and sending it to the global Weaver manager.
func (e *ResolverExecutor) Process(i *intent.Intent) error {
	// For backward compatibility, call ProcessWithContext with default context
	return e.ProcessWithContext(context.Background(), HandleContext{}, i)
}

// ProcessWithContext processes an intent with context information.
func (e *ResolverExecutor) ProcessWithContext(ctx context.Context, hctx HandleContext, i *intent.Intent) error {
	// PaneID 边界注入（与 kernel.go legacy 路径对齐）：
	// Grammar 永不产生 PaneID，ActorID 格式为 "paneID|clientName"。
	// 此前 weaver 路径漏了这步注入，下游拿到的 PaneID 恒为空。
	if i.PaneID == "" && hctx.ActorID != "" {
		if parts := strings.SplitN(hctx.ActorID, "|", 2); len(parts) > 0 {
			i.PaneID = parts[0]
		}
	}

	weaverMgr := manager.GetWeaverManager()
	if weaverMgr == nil {
		log.Println("Weaver manager is not initialized, intent dropped.")
		return nil
	}

	// Convert kernel HandleContext to core HandleContext
	coreHctx := core.HandleContext{
		RequestID: hctx.RequestID,
		ActorID:   hctx.ActorID,
	}

	// intent.Intent now implements core.Intent interface directly.
	return weaverMgr.ProcessIntentGlobalWithContext(coreHctx, i)
}
