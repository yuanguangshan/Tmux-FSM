// Package resolver - DEPRECATED: 冻结状态，不再开发
//
// 此包已被标记为冻结状态，不再接受任何新功能开发。
// 所有新的Vim语义解析逻辑应使用 main 包中的新Resolver实现。
//
// 此包仅用于过渡期兼容，最终将被完全替换。
package resolver

import (
	"errors"
	"tmux-fsm/intent"
)

// RepeatableAction 可重复操作
type RepeatableAction struct {
	Operator *intent.OperatorKind
	Motion   *intent.Motion
	Count    int
}

// Resolver 解析器
type Resolver struct {
	engine EngineAdapter

	lastRepeat *RepeatableAction
	lastFind   *intent.FindMotion

	undoTree     *UndoTree
	macroManager *MacroManager
}

// New 创建新的解析器
func New(_ EngineAdapter) *Resolver {
	return &Resolver{
		engine: &NoopEngine{},
	}
}

// Resolve 解析意图
func (r *Resolver) Resolve(i *intent.Intent) error {
	return r.ResolveWithContext(i, ExecContext{})
}

// ResolveWithContext 解析意图（带上下文）
func (r *Resolver) ResolveWithContext(i *intent.Intent, ctx ExecContext) error {
	if i == nil {
		return errors.New("nil intent")
	}

	// 如果不是来自宏且正在录制宏，则记录意图
	if !ctx.FromMacro && r.macroManager != nil && r.macroManager.recording != nil {
		r.recordIntentForMacro(i)
	}

	var err error

	switch i.Kind {
	case intent.IntentMove:
		err = r.resolveMove(i)

	case intent.IntentOperator:
		err = r.resolveOperatorWithContext(i, ctx)

	case intent.IntentRepeat:
		err = r.resolveRepeatWithContext(i, ctx)

	case intent.IntentUndo:
		err = r.resolveUndo(i)

	case intent.IntentMacro:
		err = r.resolveMacro(i)

	case intent.IntentEnterVisual:
		// 暂时忽略视觉模式相关意图
		return nil

	case intent.IntentExitVisual:
		// 暂时忽略视觉模式相关意图
		return nil

	case intent.IntentRepeatFind:
		err = r.repeatFind(false)

	case intent.IntentRepeatFindReverse:
		err = r.repeatFind(true)

	default:
		// 忽略其他类型
	}

	// 如果不是撤销或重复操作，且不是来自重复操作，则记录操作
	if err == nil && i.Kind != intent.IntentUndo && i.Kind != intent.IntentRepeat && !ctx.FromRepeat {
		r.recordAction(i)
	}

	return err
}