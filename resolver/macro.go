package resolver

import (
	"tmux-fsm/intent"
)

// Macro 宏结构
type Macro struct {
	Name   string
	IntentSequence []*intent.Intent
	Active bool
}

// MacroManager 宏管理器
type MacroManager struct {
	macros   map[string]*Macro
	recording *Macro
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
			mm.recording.IntentSequence = append(mm.recording.IntentSequence, i)
		}
	}
}

// GetMacro 获取宏
func (mm *MacroManager) GetMacro(name string) *Macro {
	return mm.macros[name]
}

// PlayMacro 播放宏
func (mm *MacroManager) PlayMacro(name string) []*intent.Intent {
	macro := mm.macros[name]
	if macro == nil {
		return nil
	}
	return macro.IntentSequence
}

// 在resolver中添加macro manager
func (r *Resolver) initMacro() {
	if r.macroManager == nil {
		r.macroManager = NewMacroManager()
	}
}

// resolveMacroWithContext 解析宏意图（带上下文）
func (r *Resolver) resolveMacroWithContext(i *intent.Intent, ctx ExecContext) error {
	r.initMacro()

	operation, ok := i.Meta["operation"].(string)
	if !ok {
		return nil
	}

	switch operation {
	case "start_recording":
		name, ok := i.Meta["name"].(string)
		if ok {
			r.macroManager.StartRecording(name)
		}
	case "stop_recording":
		r.macroManager.StopRecording()
	case "play":
		name, ok := i.Meta["name"].(string)
		if ok {
			sequence := r.macroManager.PlayMacro(name)
			for _, intent := range sequence {
				// 创建新的上下文，标记为来自宏
				newCtx := ExecContext{
					FromMacro:  true,
					FromRepeat: ctx.FromRepeat, // 保持重复上下文
					FromUndo:   ctx.FromUndo,   // 保持撤销上下文
				}
				// 递归执行宏中的每个意图
				_ = r.ResolveWithContext(intent, newCtx)
			}
		}
	}

	return nil
}

// resolveMacro 解析宏意图（兼容旧接口）
func (r *Resolver) resolveMacro(i *intent.Intent) error {
	return r.resolveMacroWithContext(i, ExecContext{})
}

// 在执行意图时，如果正在录制宏，则添加到宏中
func (r *Resolver) recordIntentForMacro(i *intent.Intent) {
	if r.macroManager != nil && r.macroManager.recording != nil {
		r.macroManager.AddIntentToRecording(i)
	}
}