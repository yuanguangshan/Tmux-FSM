package kernel

import (
	"context"
	"log"
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

	// Adapt the intent to the core.Intent interface and process it.
	adaptedIntent := &intent.Adapter{Intent: *i}
	return weaverMgr.ProcessIntentGlobalWithContext(coreHctx, adaptedIntent)
}
