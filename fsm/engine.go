package fsm

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
	"tmux-fsm/fsm/ui"
)

// Engine FSM 引擎结构体
type Engine struct {
	Active     string
	Keymap     *Keymap
	layerTimer *time.Timer
	UI         ui.UI
}

// 全局默认引擎实例
var defaultEngine *Engine

// NewEngine 创建新的 FSM 引擎实例（显式注入 Keymap）
func NewEngine(km *Keymap) *Engine {
	return &Engine{
		Active: "NAV",
		Keymap: km,
	}
}

// InitEngine 初始化全局唯一 Engine
func InitEngine(km *Keymap) {
	defaultEngine = NewEngine(km)
}

// InLayer 检查当前是否处于非默认层（如 GOTO）
func (e *Engine) InLayer() bool {
	return e.Active != "NAV" && e.Active != ""
}

// CanHandle 检查当前层是否定义了该按键
func (e *Engine) CanHandle(key string) bool {
	if e.Keymap == nil {
		return false
	}
	st, ok := e.Keymap.States[e.Active]
	if !ok {
		return false
	}
	_, exists := st.Keys[key]
	return exists
}

// Dispatch 处理按键交互
func (e *Engine) Dispatch(key string) bool {
	if !e.CanHandle(key) {
		return false
	}

	st := e.Keymap.States[e.Active]
	act := st.Keys[key]

	// 1. 处理层切换
	if act.Layer != "" {
		e.Active = act.Layer
		e.resetLayerTimeout(act.TimeoutMs)
		UpdateUI()
		return true
	}

	// 2. 处理具体动作
	if act.Action != "" {
		e.RunAction(act.Action)

		// 铁律：执行完动作后，除非该层标记为 Sticky，否则立刻 Reset 回 NAV
		if !st.Sticky {
			e.Reset()
		} else {
			// 如果是 Sticky 层，可能需要刷新 UI（如 hint）
			UpdateUI()
		}
		return true
	}

	return false
}

// Reset 重置引擎状态到 NAV 层
func (e *Engine) Reset() {
	e.Active = "NAV"
	if e.layerTimer != nil {
		e.layerTimer.Stop()
	}
	// 执行重置通常意味着退出特定层级的 UI 显示
	HideUI()
}

// GetActiveLayer 获取当前层名称
func GetActiveLayer() string {
	if defaultEngine == nil {
		return "NAV"
	}
	return defaultEngine.Active
}

// InLayer 全局查询
func InLayer() bool {
	if defaultEngine == nil {
		return false
	}
	return defaultEngine.InLayer()
}

// CanHandle 全局查询
func CanHandle(key string) bool {
	if defaultEngine == nil {
		return false
	}
	return defaultEngine.CanHandle(key)
}

// Reset 全局重置
func Reset() {
	if defaultEngine != nil {
		defaultEngine.Reset()
	}
}

// ... (resetLayerTimeout remains same)
func (e *Engine) resetLayerTimeout(ms int) {
	if e.layerTimer != nil {
		e.layerTimer.Stop()
	}
	if ms > 0 {
		e.layerTimer = time.AfterFunc(
			time.Duration(ms)*time.Millisecond,
			func() {
				e.Reset()
				// 这里由于是异步超时，需要手动触发一次 UI 刷新
				UpdateUI()
			},
		)
	}
}

// RunAction 执行动作
func (e *Engine) RunAction(name string) {
	switch name {
	case "pane_left":
		tmux("select-pane -L")
	case "pane_right":
		tmux("select-pane -R")
	case "pane_up":
		tmux("select-pane -U")
	case "pane_down":
		tmux("select-pane -D")
	case "next_pane":
		tmux("select-pane -t :.+")
	case "prev_pane":
		tmux("select-pane -t :.-")
	case "far_left":
		tmux("select-pane -t :.0")
	case "far_right":
		tmux("select-pane -t :.$")
	case "goto_top":
		tmux("select-pane -t :.0")
	case "goto_bottom":
		tmux("select-pane -t :.$")
	case "exit":
		ExitFSM()
	case "prompt":
		tmux("command-prompt")
	default:
		fmt.Println("unknown action:", name)
	}
}

func tmux(cmd string) {
	exec.Command("tmux", strings.Split(cmd, " ")...).Run()
}

// 全局函数，支持在其他包调用
func Dispatch(key string) bool {
	if defaultEngine == nil {
		return false
	}
	return defaultEngine.Dispatch(key)
}

func EnterFSM() {
	if defaultEngine == nil {
		InitEngine(&KM)
	}

	engine := defaultEngine
	engine.Active = "NAV"
	// 确保进入时是干净的 NAV
	engine.Reset()
	// ShowUI() // Disable initial UI popup to prevent flashing/annoyance
}

func ExitFSM() {
	if defaultEngine != nil {
		defaultEngine.Reset()
	}
	HideUI()
	exec.Command("tmux", "set-option", "-u", "key-table").Run()
}