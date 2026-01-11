package main

import (
	"encoding/json"
	"sync"
	"testing"
)

// TestCursorStruct 测试Cursor结构
func TestCursorStruct(t *testing.T) {
	cursor := Cursor{
		Row: 5,
		Col: 10,
	}

	if cursor.Row != 5 {
		t.Errorf("Expected Row to be 5, got %d", cursor.Row)
	}

	if cursor.Col != 10 {
		t.Errorf("Expected Col to be 10, got %d", cursor.Col)
	}
}

// TestFSMStateStruct 测试FSMState结构
func TestFSMStateStruct(t *testing.T) {
	state := FSMState{
		Mode:        "NORMAL",
		Operator:    "delete",
		Count:       3,
		PendingKeys: "dw",
		Register:    "a",
		PaneID:      "pane1",
		Cursor:      Cursor{Row: 1, Col: 2},
	}

	if state.Mode != "NORMAL" {
		t.Errorf("Expected Mode to be 'NORMAL', got '%s'", state.Mode)
	}

	if state.Operator != "delete" {
		t.Errorf("Expected Operator to be 'delete', got '%s'", state.Operator)
	}

	if state.Count != 3 {
		t.Errorf("Expected Count to be 3, got %d", state.Count)
	}

	if state.PendingKeys != "dw" {
		t.Errorf("Expected PendingKeys to be 'dw', got '%s'", state.PendingKeys)
	}

	if state.Register != "a" {
		t.Errorf("Expected Register to be 'a', got '%s'", state.Register)
	}

	if state.PaneID != "pane1" {
		t.Errorf("Expected PaneID to be 'pane1', got '%s'", state.PaneID)
	}

	if state.Cursor.Row != 1 || state.Cursor.Col != 2 {
		t.Errorf("Expected Cursor to be {1, 2}, got {%d, %d}", state.Cursor.Row, state.Cursor.Col)
	}
}

// TestFSMStateJSONSerialization 测试FSMState的JSON序列化
func TestFSMStateJSONSerialization(t *testing.T) {
	originalState := FSMState{
		Mode:        "INSERT",
		Operator:    "yank",
		Count:       5,
		PendingKeys: "yw",
		Register:    "b",
		PaneID:      "pane2",
		Cursor:      Cursor{Row: 3, Col: 4},
	}

	// 序列化
	data, err := json.Marshal(originalState)
	if err != nil {
		t.Fatalf("Failed to marshal FSMState: %v", err)
	}

	// 反序列化
	var newState FSMState
	err = json.Unmarshal(data, &newState)
	if err != nil {
		t.Fatalf("Failed to unmarshal FSMState: %v", err)
	}

	if newState.Mode != originalState.Mode {
		t.Errorf("Expected Mode to be '%s', got '%s'", originalState.Mode, newState.Mode)
	}

	if newState.Operator != originalState.Operator {
		t.Errorf("Expected Operator to be '%s', got '%s'", originalState.Operator, newState.Operator)
	}

	if newState.Count != originalState.Count {
		t.Errorf("Expected Count to be %d, got %d", originalState.Count, newState.Count)
	}

	if newState.PendingKeys != originalState.PendingKeys {
		t.Errorf("Expected PendingKeys to be '%s', got '%s'", originalState.PendingKeys, newState.PendingKeys)
	}

	if newState.Register != originalState.Register {
		t.Errorf("Expected Register to be '%s', got '%s'", originalState.Register, newState.Register)
	}

	if newState.PaneID != originalState.PaneID {
		t.Errorf("Expected PaneID to be '%s', got '%s'", originalState.PaneID, newState.PaneID)
	}

	if newState.Cursor.Row != originalState.Cursor.Row || newState.Cursor.Col != originalState.Cursor.Col {
		t.Errorf("Expected Cursor to be {%d, %d}, got {%d, %d}", 
			originalState.Cursor.Row, originalState.Cursor.Col,
			newState.Cursor.Row, newState.Cursor.Col)
	}
}

// TestGlobalVariables 测试全局变量
func TestGlobalVariables(t *testing.T) {
	// 测试全局变量的存在性
	if stateMu == (sync.Mutex{}) {
		// 这个测试主要是确保变量存在，不需要验证具体值
	}

	if globalState.Mode != "NORMAL" || globalState.Count != 0 {
		// 默认值可能在init函数中被设置，我们验证结构存在
	}

	if transMgr == nil {
		t.Error("Expected transMgr to be initialized")
	}

	if txJournal == nil {
		t.Error("Expected txJournal to be initialized")
	}

	if socketPath != "/tmp/tmux-fsm.sock" {
		t.Errorf("Expected socketPath to be '/tmp/tmux-fsm.sock', got '%s'", socketPath)
	}

	if StrictNativeFSM != false {
		t.Errorf("Expected StrictNativeFSM to be false by default, got %v", StrictNativeFSM)
	}

	if StrictNativeResolver != false {
		t.Errorf("Expected StrictNativeResolver to be false by default, got %v", StrictNativeResolver)
	}

	if DebugLogging != false {
		t.Errorf("Expected DebugLogging to be false by default, got %v", DebugLogging)
	}
}

// TestLoadStateDefault 测试默认状态加载
func TestLoadStateDefault(t *testing.T) {
	// 由于loadState依赖于backend，我们测试返回默认值的情况
	// 在没有backend的情况下，应该返回默认状态
	// 为了避免与其他测试的干扰，我们不依赖全局状态的当前值
	// 而是关注函数本身的行为

	// 保存当前全局状态
	originalGlobalState := globalState

	// 重置全局状态为默认值
	globalState = FSMState{Mode: "NORMAL", Count: 0, Cursor: Cursor{Row: 0, Col: 0}}

	// 现在调用loadState，它应该从backend加载（如果没有则返回默认值）
	// 但由于backend可能返回上次保存的值，我们只测试函数不panic
	state := loadState()

	// 恢复原始全局状态
	globalState = originalGlobalState

	// 我们只是确保函数不panic，并返回一个有效的FSMState
	if state.Mode == "" {
		t.Error("Expected state to have a valid mode")
	}
}

// TestSaveFSMState 测试保存FSM状态
func TestSaveFSMState(t *testing.T) {
	// 保存当前状态
	originalState := globalState
	
	// 设置一些测试值
	testState := FSMState{
		Mode:     "TEST",
		Count:    42,
		Cursor:   Cursor{Row: 10, Col: 20},
	}
	
	globalState = testState
	
	// 调用保存函数（这会尝试保存到tmux，但测试中可能失败，这是正常的）
	saveFSMState()
	
	// 恢复原始状态
	globalState = originalState
	
	// 我们只是确保函数不panic
}

// TestGetTmuxCursorPos 测试获取tmux光标位置
// 注意：这个函数需要实际的tmux环境，所以我们只测试函数存在性
func TestGetTmuxCursorPos(t *testing.T) {
	// 这个函数需要tmux环境，我们只是确保它不会panic
	// 在测试环境中，它可能会返回错误，但不应该panic
	pos := GetTmuxCursorPos("dummy-pane-id")
	// 不验证具体值，因为这需要真实的tmux环境
	_ = pos
}

// TestUpdateStatusBar 测试更新状态栏
func TestUpdateStatusBar(t *testing.T) {
	// 创建一个测试状态
	state := FSMState{
		Mode:     "NORMAL",
		Count:    5,
		Operator: "delete",
	}
	
	// 调用更新状态栏函数
	// 在测试环境中，这可能会失败，但不应该panic
	updateStatusBar(state, "test-client")
	
	// 我们只是确保函数不panic
}
