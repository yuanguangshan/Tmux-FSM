package kernel

import "tmux-fsm/intent"

// IntentExecutor is the ONLY way Kernel can execute an Intent.
// Kernel does not know who implements it.
type IntentExecutor interface {
	Process(*intent.Intent) error
}
