package kernel

import (
	"context"
	"tmux-fsm/intent"
)

// IntentExecutor is the ONLY way Kernel can execute an Intent.
// Kernel does not know who implements it.
type IntentExecutor interface {
	Process(*intent.Intent) error
}

// ContextualIntentExecutor extends IntentExecutor to support context passing.
type ContextualIntentExecutor interface {
	IntentExecutor
	ProcessWithContext(ctx context.Context, hctx HandleContext, intent *intent.Intent) error
}
