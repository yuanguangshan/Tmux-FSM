package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
	"tmux-fsm/backend"
)

type Cursor struct {
	Row int
	Col int
}

type FSMState struct {
	Mode                 string                 `json:"mode"`
	Operator             string                 `json:"operator"`
	Count                int                    `json:"count"`
	PendingKeys          string                 `json:"pending_keys"`
	Register             string                 `json:"register"`
	LastRepeatableAction map[string]interface{} `json:"last_repeatable_action"`
	UndoStack            []Transaction          `json:"undo_stack"`
	RedoStack            []Transaction          `json:"redo_stack"`
	LastUndoFailure      string                 `json:"last_undo_failure,omitempty"`
	LastUndoSafetyLevel  string                 `json:"last_undo_safety_level,omitempty"`
	AllowPartial         bool                   `json:"allow_partial"` // Phase 7: Explicit permission for fuzzy resolution
	PaneID               string                 `json:"pane_id"`       // Current pane ID for intent processing
	Cursor               Cursor                 `json:"cursor"`        // Current cursor position
}

var (
	stateMu     sync.Mutex
	globalState FSMState
	transMgr    *TransactionManager
	txJournal   *TxJournal // 新增：事务日志
	socketPath  = "/tmp/tmux-fsm.sock"
)

func init() {
	// 初始化全局事务管理器
	transMgr = &TransactionManager{
		nextID: 0,
	}

	// 初始化事务日志
	txJournal = NewTxJournal()
}

func loadState() FSMState {
	// Use GlobalBackend to read tmux options
	out, err := backend.GlobalBackend.GetUserOption("@tmux_fsm_state")
	var state FSMState
	if err != nil || len(out) == 0 {
		return FSMState{Mode: "NORMAL", Count: 0, Cursor: Cursor{Row: 0, Col: 0}}
	}
	json.Unmarshal([]byte(out), &state)
	return state
}

// GetTmuxCursorPos 获取 tmux 光标位置 [col, row]
func GetTmuxCursorPos(paneID string) [2]int {
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", paneID, "#{pane_cursor_x},#{pane_cursor_y}").Output()
	var x, y int
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d,%d", &x, &y)
	return [2]int{x, y}
}

func saveStateRaw(data []byte) {
	// Use GlobalBackend to save state
	// This implies SetUserOption needs to be able to set arbitrary keys.
	if err := backend.GlobalBackend.SetUserOption("@tmux_fsm_state", string(data)); err != nil {
		log.Printf("Failed to save FSM state: %v", err)
	}
}

// saveFSMState 保存 FSM 状态
func saveFSMState() {
	stateMu.Lock()
	defer stateMu.Unlock()

	data, err := json.Marshal(globalState)
	if err != nil {
		log.Printf("Failed to marshal FSM state: %v", err)
		return
	}

	saveStateRaw(data)
}

func updateStatusBar(state FSMState, clientName string) {
	modeMsg := state.Mode
	if modeMsg == "" {
		modeMsg = "NORMAL"
	}

	// 融合显示逻辑
	// activeLayer := fsm.GetActiveLayer() // 由于循环导入，暂时注释掉
	// if activeLayer != "NAV" && activeLayer != "" {
	// 	modeMsg = activeLayer // Override with FSM layer if active
	// } else {
	// Translate legacy FSM modes for display
	switch modeMsg {
	case "VISUAL_CHAR":
		modeMsg = "VISUAL"
	case "VISUAL_LINE":
		modeMsg = "V-LINE"
	case "OPERATOR_PENDING":
		modeMsg = "PENDING"
	case "REGISTER_SELECT":
		modeMsg = "REGISTER"
	case "MOTION_PENDING":
		modeMsg = "MOTION"
	case "SEARCH":
		modeMsg = "SEARCH"
	}
	// }

	if state.Operator != "" {
		modeMsg += fmt.Sprintf(" [%s]", state.Operator)
	}
	if state.Count > 0 {
		modeMsg += fmt.Sprintf(" [%d]", state.Count)
	}

	keysMsg := ""
	if state.PendingKeys != "" {
		if state.Mode == "SEARCH" {
			keysMsg = fmt.Sprintf(" /%s", state.PendingKeys)
		} else {
			keysMsg = fmt.Sprintf(" (%s)", state.PendingKeys)
		}
	}

	if state.LastUndoSafetyLevel == "fuzzy" {
		keysMsg += " ~UNDO"
	} else if state.LastUndoFailure != "" {
		keysMsg += " !UNDO_FAIL"
	}

	// Debug logging
	f, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] Updating status: mode=%s, state.Mode=%s, keys=%s\n",
			time.Now().Format("15:04:05"), modeMsg, state.Mode, keysMsg)
		f.Close()
	}

	// Use GlobalBackend for tmux option updates
	backend.GlobalBackend.SetUserOption("@fsm_state", modeMsg)
	backend.GlobalBackend.SetUserOption("@fsm_keys", keysMsg)
	backend.GlobalBackend.RefreshClient(clientName) // Refresh the target client

	// --- [ABI: Heartbeat Lock] ---
	// Re-assert the key table to prevent "one-shot" dropouts.
	// Check @fsm_active to allow intentional exits.
	if clientName != "" && clientName != "default" {
		// Fetching @fsm_active via GlobalBackend if it were available would be ideal,
		// but for now, we rely on the fact that we are in a state where we should be active.
		// If GlobalBackend could read options, it would be better.
		// For now, we assume if we got here, FSM is active.
		backend.GlobalBackend.SwitchClientTable(clientName, "fsm")
	}
}
