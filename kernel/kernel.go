package kernel

import (
	"context"
	"tmux-fsm/fsm"
)

type Kernel struct {
	FSM    *fsm.Engine
	// Weaver *WeaverManager  // Temporarily disabled to avoid import issues
}

// ✅ Kernel 的唯一上下文入口（现在先很薄，未来可扩展）
type HandleContext struct {
	Ctx context.Context
}

func NewKernel(fsmEngine *fsm.Engine, weaverMgr interface{}) *Kernel {
	return &Kernel{
		FSM: fsmEngine,
		// Weaver: weaverMgr,  // Temporarily disabled
	}
}

// ✅ Kernel 的唯一入口
func (k *Kernel) HandleKey(hctx HandleContext, key string) {
	_ = hctx // ✅ 现在不用，但接口已经锁死

	decision := k.Decide(key)
	if decision == nil {
		return
	}

	k.Execute(decision)
}
