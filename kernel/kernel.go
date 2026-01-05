package kernel

import (
	"context"

	"yourmodule/fsm"
	"yourmodule/weaver"
)

type Kernel struct {
	FSM    *fsm.Engine
	Weaver *weaver.Manager
}

type HandleContext struct {
	Ctx context.Context
}

func NewKernel(fsmEngine *fsm.Engine, weaverMgr *weaver.Manager) *Kernel {
	return &Kernel{
		FSM:    fsmEngine,
		Weaver: weaverMgr,
	}
}

func (k *Kernel) HandleKey(hctx HandleContext, key string) {
	decision := k.Decide(key)

	if decision == nil {
		return
	}

	k.Execute(decision)
}
