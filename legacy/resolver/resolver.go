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

// Macro 宏结构
type Macro struct {
	Name           string
	IntentSequence []*intent.Intent
	Active         bool
}

// MacroManager 宏管理器
type MacroManager struct {
	macros    map[string]*Macro
	recording *Macro
}

// Resolver 解析器
type Resolver struct {
	engine EngineAdapter

	lastRepeat *RepeatableAction
	lastFind   *intent.FindMotion

	undoTree     *UndoTree
	macroManager *MacroManager
}

// NewMacroManager 创建新的宏管理器
func NewMacroManager() *MacroManager {
	return &MacroManager{
		macros: make(map[string]*Macro),
	}
}

// StartRecording 开始录制宏
func (mm *MacroManager) StartRecording(name string) {
	macro := &Macro{
		Name:           name,
		IntentSequence: make([]*intent.Intent, 0),
		Active:         true,
	}
	mm.recording = macro
}

// StopRecording 停止录制宏
func (mm *MacroManager) StopRecording() {
	if mm.recording != nil {
		mm.macros[mm.recording.Name] = mm.recording
		mm.recording = nil
	}
}

// AddIntentToRecording 向正在录制的宏添加意图
func (mm *MacroManager) AddIntentToRecording(i *intent.Intent) {
	if mm.recording != nil {
		// 只记录某些类型的意图
		if i.Kind == intent.IntentMove || i.Kind == intent.IntentOperator {
			// 深拷贝意图以避免后续修改影响录制内容
			mm.recording.IntentSequence = append(mm.recording.IntentSequence, cloneIntent(i))
		}
	}
}

// GetMacro 获取宏
func (mm *MacroManager) GetMacro(name string) *Macro {
	return mm.macros[name]
}

// PlayMacro 撪放宏
func (mm *MacroManager) PlayMacro(name string) []*intent.Intent {
	macro := mm.macros[name]
	if macro == nil {
		return nil
	}
	return macro.IntentSequence
}

// New 创建新的解析器
// NOTE: Resolver currently runs in semantic-only mode.
// EngineAdapter will be injected in Phase-2.
func New(adapter EngineAdapter) *Resolver {
	return &Resolver{
		engine:       adapter,
		macroManager: NewMacroManager(),
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

	// 如果不是来自宏，且正在录制宏，则记录意图
	if !ctx.FromMacro && r.macroManager != nil && r.macroManager.recording != nil {
		r.recordIntentForMacro(i)
	}

	// 如果不是撤销或重复操作，且不是来自重复操作，则记录操作
	if err == nil && i.Kind != intent.IntentUndo && i.Kind != intent.IntentRepeat && !ctx.FromRepeat {
		r.recordAction(i)
	}

	return err
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

// resolveMacro 解析宏意图
func (r *Resolver) resolveMacro(i *intent.Intent) error {
	operation, ok := i.Meta["operation"].(string)
	if !ok {
		return nil
	}

	switch operation {
	case "start_recording":
		name, ok := i.Meta["register"].(string)
		if ok {
			r.macroManager.StartRecording(name)
		}
	case "stop_recording":
		r.macroManager.StopRecording()
	case "play":
		name, ok := i.Meta["register"].(string)
		if ok {
			sequence := r.macroManager.PlayMacro(name)

			// 创建新的上下文，标记为来自宏
			newCtx := ExecContext{
				FromMacro:  true,
				FromRepeat: false, // 宏播放时不应记录重复
				FromUndo:   false, // 宏播放时不应记录撤销
			}

			// 递归执行宏中的每个意图
			for _, intent := range sequence {
				// 根据计数重复执行
				count := i.Count
				if count <= 0 {
					count = 1
				}

				for j := 0; j < count; j++ {
					_ = r.ResolveWithContext(intent, newCtx)
				}
			}
		}
	}

	return nil
}

// recordIntentForMacro 在执行意图时，如果正在录制宏，则添加到宏中
func (r *Resolver) recordIntentForMacro(i *intent.Intent) {
	if r.macroManager != nil && r.macroManager.recording != nil {
		r.macroManager.AddIntentToRecording(i)
	}
}
