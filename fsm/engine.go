package fsm

import (
	"fmt"
	"time"
)

// RawTokenEmitter 用于发送 RawToken 的接口
type RawTokenEmitter interface {
	Emit(RawToken)
}

// Engine FSM 引擎结构体
type Engine struct {
	Active     string
	Keymap     *Keymap
	layerTimer *time.Timer
	count      int              // 用于存储数字计数
	emitters   []RawTokenEmitter // 用于向外部发送token的多个接收者
}

// FSMStatus FSM 状态信息，用于UI更新
type FSMStatus struct {
	Layer string
	Count int
}

// AddEmitter 添加一个 token 发送接收者
func (e *Engine) AddEmitter(emitter RawTokenEmitter) {
	e.emitters = append(e.emitters, emitter)
}

// RemoveEmitter 移除一个 token 发送接收者
func (e *Engine) RemoveEmitter(emitter RawTokenEmitter) {
	for i, em := range e.emitters {
		if em == emitter {
			e.emitters = append(e.emitters[:i], e.emitters[i+1:]...)
			break
		}
	}
}

// emitInternal 内部发送 token 给所有订阅者
func (e *Engine) emitInternal(token RawToken) {
	for _, emitter := range e.emitters {
		emitter.Emit(token)
	}
}

// 全局默认引擎实例
var defaultEngine *Engine


// NewEngine 创建新的 FSM 引擎实例（显式注入 Keymap）
func NewEngine(km *Keymap) *Engine {
	return &Engine{
		Active:   "NAV",
		Keymap:   km,
		count:    0,
		emitters: make([]RawTokenEmitter, 0),
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
	// 检查是否是数字键，即使当前层没有定义
	if isDigit(key) {
		e.count = e.count*10 + int(key[0]-'0')
		e.emitInternal(RawToken{Kind: TokenDigit, Value: key})
		return true
	}

	// 检查是否是重复键
	if key == "." {
		e.emitInternal(RawToken{Kind: TokenRepeat, Value: "."})
		return true
	}

	// 其他按键按原有逻辑处理（只处理层切换，不处理动作）
	if e.CanHandle(key) {
		st := e.Keymap.States[e.Active]
		act := st.Keys[key]

		// 1. 处理层切换
		if act.Layer != "" {
			e.Active = act.Layer
			e.resetLayerTimeout(act.TimeoutMs)
			e.emitInternal(RawToken{Kind: TokenKey, Value: key})
			return true
		}

		// 2. 发送按键 token
		e.emitInternal(RawToken{Kind: TokenKey, Value: key})
		return true
	}

	return false
}

// isDigit 检查字符串是否为单个数字字符
func isDigit(s string) bool {
	return len(s) == 1 && s[0] >= '0' && s[0] <= '9'
}

// Reset 重置引擎状态到初始层（Invariant 8: Reload = FSM 重生）
func (e *Engine) Reset() {
	if e.layerTimer != nil {
		e.layerTimer.Stop()
		e.layerTimer = nil
	}
	// 重置到初始状态
	if e.Keymap != nil && e.Keymap.Initial != "" {
		e.Active = e.Keymap.Initial
	} else {
		e.Active = "NAV"
	}
	e.count = 0
	e.emitInternal(RawToken{Kind: TokenSystem, Value: "reset"})
}


// Reload 重新加载keymap并重置FSM（Invariant 8: Reload = atomic rebuild）
func Reload(configPath string) error {
	// Load + Validate
	if err := LoadKeymap(configPath); err != nil {
		return err
	}

	// NewEngine
	InitEngine(&KM)

	// Reset + UI refresh
	Reset()

	return nil
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
	case "goto_line_start":
		// 发送 Home 键到当前窗格，这通常会将光标移到行首
		tmux("send-keys -t . Home")
	case "goto_line_end":
		// 发送 End 键到当前窗格，这通常会将光标移到行尾
		tmux("send-keys -t . End")
	case "move_left":
		// 发送左箭头键
		tmux("send-keys -t . Left")
	case "move_right":
		// 发送右箭头键
		tmux("send-keys -t . Right")
	case "move_up":
		// 发送上箭头键
		tmux("send-keys -t . Up")
	case "move_down":
		// 发送下箭头键
		tmux("send-keys -t . Down")
	case "exit":
		ExitFSM()
	case "prompt":
		tmux("command-prompt")
	default:
		fmt.Println("unknown action:", name)
	}
}


func tmux(cmd string) {
	// Use GlobalBackend to execute the command
	// 由于循环导入问题，这里暂时使用占位符
	// 实际执行应该由上层处理
}


func EnterFSM() {
	if defaultEngine == nil {
		InitEngine(&KM)
	}

	engine := defaultEngine
	engine.Active = "NAV"
	// 确保进入时是干净的 NAV
	engine.Reset()
	engine.emitInternal(RawToken{Kind: TokenSystem, Value: "enter"})
	UpdateUI() // 确保进入时更新UI
	// ShowUI() // Disable initial UI popup to prevent flashing/annoyance
}

// GetDefaultEngine 获取默认引擎实例
func GetDefaultEngine() *Engine {
	return defaultEngine
}

func ExitFSM() {
	if defaultEngine != nil {
		defaultEngine.Reset()
		defaultEngine.emitInternal(RawToken{Kind: TokenSystem, Value: "exit"})
	}
	HideUI()
	UpdateUI() // 确保退出时更新UI
	// FSM 不应直接依赖 backend
	// 执行层的退出逻辑应该由上层处理
}
