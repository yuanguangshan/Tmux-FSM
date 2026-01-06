package bridge

import (
	"time"
	"tmux-fsm/backend"
	"tmux-fsm/fsm"
)

// LegacyFSMHandler 处理与旧 FSM 系统的交互
type LegacyFSMHandler struct {
	NewFSMEnabled bool
}

// NewLegacyFSMHandler 创建新的处理器
func NewLegacyFSMHandler() *LegacyFSMHandler {
	return &LegacyFSMHandler{
		NewFSMEnabled: true, // 默认启用新 FSM
	}
}

// HandleKey 处理按键输入
func (h *LegacyFSMHandler) HandleKey(key string) string {
	if h.NewFSMEnabled {
		// 检查是否在新 FSM 配置中有定义
		if stateDef, ok := fsm.KM.States[fsm.Active]; ok {
			if action, exists := stateDef.Keys[key]; exists {
				// 如果是层切换
				if action.Layer != "" {
					fsm.Active = action.Layer
					h.resetLayerTimeout(action.TimeoutMs)
					fsm.UpdateUI()
					return ""
				}
				// 执行动作
				fsm.RunAction(action.Action)
				return ""
			}
		}
	}

	// 如果新系统未处理，返回空字符串让旧系统处理
	return ""
}

// resetLayerTimeout 重置层超时
func (h *LegacyFSMHandler) resetLayerTimeout(ms int) {
	// 这里需要访问 fsm 包中的 timer，可能需要修改 fsm 包的设计
	if fsm.LayerTimer != nil {
		fsm.LayerTimer.Stop()
	}
	if ms > 0 {
		fsm.LayerTimer = time.AfterFunc(
			time.Duration(ms)*time.Millisecond,
			func() {
				fsm.Active = "NAV"
				fsm.UpdateUI()
			},
		)
	}
}

// EnterFSM 进入 FSM 模式
func (h *LegacyFSMHandler) EnterFSM() {
	if h.NewFSMEnabled {
		fsm.EnterFSM()
	} else {
		// 保留旧的进入逻辑
		backend.GlobalBackend.SetUserOption("@fsm_active", "true")
		backend.GlobalBackend.SwitchClientTable("", "fsm")
	}
}

// ExitFSM 退出 FSM 模式
func (h *LegacyFSMHandler) ExitFSM() {
	if h.NewFSMEnabled {
		fsm.ExitFSM()
	} else {
		// 保留旧的退出逻辑
		backend.GlobalBackend.SetUserOption("@fsm_active", "false")
		backend.GlobalBackend.SetUserOption("@fsm_state", "")
		backend.GlobalBackend.SetUserOption("@fsm_keys", "")
		backend.GlobalBackend.SwitchClientTable("", "root")
		backend.GlobalBackend.RefreshClient("")
	}
}
