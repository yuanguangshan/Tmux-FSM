package resolver

import (
	"errors"
	"tmux-fsm/intent"
)

// EngineAdapter 定义引擎适配器接口
type EngineAdapter interface {
	SendKeys(keys ...string)
	RunAction(name string)
	GetVisualMode() intent.VisualMode
	SetVisualMode(mode intent.VisualMode)
	EnterVisualMode(mode intent.VisualMode)
	ExitVisualMode()

	// Selection 相关方法
	EnterSelection(mode SelectionMode)
	UpdateSelection(anchor, focus Cursor)
	ExitSelection()
	GetCurrentCursor() Cursor

	// 语义操作方法
	DeleteSelection(selection *Selection) error
	DeleteWithMotion(motion intent.MotionKind, count int) error
	YankSelection(selection *Selection) error
	YankWithMotion(motion intent.MotionKind, count int) error
	ChangeSelection(selection *Selection) error
	ChangeWithMotion(motion intent.MotionKind, count int) error
}

// Resolver 解析器结构体
type Resolver struct {
	engine          EngineAdapter
	undoTree        *UndoTree
	macroManager    *MacroManager
	lastRepeatAction *RepeatableAction
	selection       *Selection
}

// RepeatableAction 可重复操作
type RepeatableAction struct {
	Operator *intent.Intent  // 操作符（如 delete）
	Motion   *intent.Intent  // 动作（如 word）
	Count    int             // 重复次数
	// 可选：执行前的状态快照
	PreState map[string]interface{} // 执行前状态（用于复杂操作）
}

// New 创建新的解析器实例
func New(engine EngineAdapter) *Resolver {
	return &Resolver{engine: engine}
}

// Resolve 解析意图并执行相应操作（默认上下文）
func (r *Resolver) Resolve(i *intent.Intent) error {
	return r.ResolveWithContext(i, ExecContext{})
}

// ResolveWithContext 解析意图并执行相应操作（带上下文）
func (r *Resolver) ResolveWithContext(i *intent.Intent, ctx ExecContext) error {
	if i == nil {
		return errors.New("nil intent")
	}

	// 如果不是来自宏，且正在录制宏，则记录意图
	if !ctx.FromMacro && r.macroManager != nil && r.macroManager.recording != nil {
		r.recordIntentForMacro(i)
	}

	// 处理视觉模式切换
	if err := r.handleVisualMode(i); err != nil {
		return err
	}

	var err error
	switch i.Kind {
	case intent.IntentMove:
		err = r.resolveMove(i)

	case intent.IntentOperator:
		err = r.resolveOperatorWithContext(i, ctx)

	case intent.IntentVisual:
		err = r.resolveVisual(i)

	case intent.IntentUndo:
		err = r.resolveUndo(i)

	case intent.IntentRepeat:
		err = r.resolveRepeatWithContext(i, ctx)

	case intent.IntentMacro:
		err = r.resolveMacro(i)

	default:
		return nil
	}

	// 如果不是Undo或Repeat操作，且不是来自重复操作，则记录到撤销树
	if i.Kind != intent.IntentUndo && i.Kind != intent.IntentRepeat && !ctx.FromRepeat {
		r.recordAction(i, intentKindToString(i.Kind))
	}

	// 如果不是来自重复操作，则更新lastRepeatAction（仅对可重复操作）
	if !ctx.FromRepeat {
		r.updateLastRepeatAction(i)
	}

	return err
}

// intentKindToString 将IntentKind转换为字符串
func intentKindToString(kind intent.IntentKind) string {
	switch kind {
	case intent.IntentMove:
		return "move"
	case intent.IntentOperator:
		return "operator"
	case intent.IntentVisual:
		return "visual"
	case intent.IntentInsert:
		return "insert"
	case intent.IntentDelete:
		return "delete"
	case intent.IntentMacro:
		return "macro"
	case intent.IntentRepeat:
		return "repeat"
	default:
		return "other"
	}
}

// cloneIntent 深拷贝意图
func cloneIntent(i *intent.Intent) *intent.Intent {
	if i == nil {
		return nil
	}

	meta := make(map[string]interface{})
	for k, v := range i.Meta {
		meta[k] = v
	}

	anchors := make([]intent.Anchor, len(i.Anchors))
	copy(anchors, i.Anchors)

	return &intent.Intent{
		Kind:         i.Kind,
		Target:       i.Target,
		Count:        i.Count,
		Meta:         meta,
		PaneID:       i.PaneID,
		SnapshotHash: i.SnapshotHash,
		AllowPartial: i.AllowPartial,
		Anchors:      anchors,
		UseRange:     i.UseRange,
	}
}

// isRepeatableIntent 判断意图是否可重复
func (r *Resolver) isRepeatableIntent(i *intent.Intent) bool {
	switch i.Kind {
	case intent.IntentOperator:
		// 操作符意图通常是可重复的
		return true
	case intent.IntentMove:
		// 某些移动意图可能可重复，但通常不是
		return false
	case intent.IntentVisual, intent.IntentUndo, intent.IntentRepeat, intent.IntentMacro:
		// 这些意图通常不可重复
		return false
	default:
		return false
	}
}

// updateLastRepeatAction 更新最后可重复操作
func (r *Resolver) updateLastRepeatAction(i *intent.Intent) {
	// 只有特定类型的意图才可重复
	if r.isRepeatableIntent(i) {
		r.lastRepeatAction = &RepeatableAction{
			Operator: cloneIntent(i), // 使用深拷贝
			Count:    i.Count,
		}
	}
}

// resolveRepeatWithContext 解析重复意图（带上下文）
func (r *Resolver) resolveRepeatWithContext(i *intent.Intent, ctx ExecContext) error {
	if r.lastRepeatAction == nil || r.lastRepeatAction.Operator == nil {
		return nil
	}

	// 创建新的上下文，标记为来自重复
	newCtx := ExecContext{
		FromRepeat: true,
		FromMacro:  ctx.FromMacro, // 保持宏上下文
		FromUndo:   ctx.FromUndo,  // 保持撤销上下文
	}

	// 重新执行最后一次可重复操作
	return r.ResolveWithContext(r.lastRepeatAction.Operator, newCtx)
}

// handleVisualMode 处理视觉模式切换
func (r *Resolver) handleVisualMode(i *intent.Intent) error {
	if i.Kind == intent.IntentVisual {
		switch i.Meta["operation"] {
		case "start_char":
			return r.EnterVisual(SelectionChar)
		case "start_line":
			return r.EnterVisual(SelectionLine)
		case "start_block":
			return r.EnterVisual(SelectionBlock)
		case "cancel":
			return r.ExitVisual()
		}
	}
	return nil
}