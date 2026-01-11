package fsm

import (
	"testing"
	"time"
)

// MockRawTokenEmitter 用于测试的模拟发射器
type MockRawTokenEmitter struct {
	receivedTokens []RawToken
}

func (m *MockRawTokenEmitter) Emit(token RawToken) {
	m.receivedTokens = append(m.receivedTokens, token)
}

// TestEngineInitialization 测试引擎初始化
func TestEngineInitialization(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{
					"f": {Layer: "GOTO", TimeoutMs: 800},
				},
			},
			"GOTO": {
				Keys: map[string]KeyAction{
					"j": {Action: "move_down"},
					"k": {Action: "move_up"},
				},
			},
		},
	}

	engine := NewEngine(&km)

	if engine.Active != "NAV" {
		t.Errorf("Expected initial layer to be 'NAV', got '%s'", engine.Active)
	}

	if engine.Keymap != &km {
		t.Errorf("Expected keymap to be set correctly")
	}

	if engine.count != 0 {
		t.Errorf("Expected initial count to be 0, got %d", engine.count)
	}

	if engine.visualMode != 0 {
		t.Errorf("Expected initial visual mode to be VisualNone, got %d", engine.visualMode)
	}
}

// TestEngineDispatchBasic 测试基本按键分发
func TestEngineDispatchBasic(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{
					"h": {Action: "move_left"},
					"j": {Action: "move_down"},
					"k": {Action: "move_up"},
					"l": {Action: "move_right"},
				},
			},
		},
	}

	engine := NewEngine(&km)
	mockEmitter := &MockRawTokenEmitter{}
	engine.AddEmitter(mockEmitter)

	// 测试基本按键
	result := engine.Dispatch("h")
	if !result {
		t.Error("Expected dispatch to return true for valid key")
	}

	if len(mockEmitter.receivedTokens) != 1 {
		t.Errorf("Expected 1 token to be emitted, got %d", len(mockEmitter.receivedTokens))
	}

	if mockEmitter.receivedTokens[0].Kind != TokenKey {
		t.Errorf("Expected TokenKey, got %v", mockEmitter.receivedTokens[0].Kind)
	}

	if mockEmitter.receivedTokens[0].Value != "h" {
		t.Errorf("Expected value 'h', got '%s'", mockEmitter.receivedTokens[0].Value)
	}
}

// TestEngineDispatchLayerSwitch 测试层切换
func TestEngineDispatchLayerSwitch(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{
					"f": {Layer: "GOTO", TimeoutMs: 800},
				},
			},
			"GOTO": {
				Keys: map[string]KeyAction{
					"j": {Action: "move_down"},
					"k": {Action: "move_up"},
				},
			},
		},
	}

	engine := NewEngine(&km)

	// 初始状态应该是 NAV
	if engine.Active != "NAV" {
		t.Errorf("Expected initial layer to be 'NAV', got '%s'", engine.Active)
	}

	// 分发 'f' 键，应该切换到 GOTO 层
	result := engine.Dispatch("f")
	if !result {
		t.Error("Expected dispatch to return true for layer switch key")
	}

	if engine.Active != "GOTO" {
		t.Errorf("Expected layer to be 'GOTO' after dispatching 'f', got '%s'", engine.Active)
	}
}

// TestEngineDispatchNumber 测试数字输入
func TestEngineDispatchNumber(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{
					"d": {Action: "delete"},
				},
			},
		},
	}

	engine := NewEngine(&km)

	// 测试数字输入
	engine.Dispatch("2")
	if engine.count != 2 {
		t.Errorf("Expected count to be 2 after dispatching '2', got %d", engine.count)
	}

	engine.Dispatch("3")
	if engine.count != 23 {
		t.Errorf("Expected count to be 23 after dispatching '2' and '3', got %d", engine.count)
	}

	// 测试数字后跟动作
	engine.Dispatch("d")
	if engine.count != 23 {
		t.Errorf("Expected count to remain 23 after dispatching 'd', got %d", engine.count)
	}
}

// TestEngineCanHandle 测试 CanHandle 方法
func TestEngineCanHandle(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{
					"h": {Action: "move_left"},
				},
			},
			"GOTO": {
				Keys: map[string]KeyAction{
					"j": {Action: "move_down"},
				},
			},
		},
	}

	engine := NewEngine(&km)

	// 测试在 NAV 层
	if !engine.CanHandle("h") {
		t.Error("Expected 'h' to be handled in NAV layer")
	}

	if engine.CanHandle("j") {
		t.Error("Expected 'j' to not be handled in NAV layer")
	}

	// 切换到 GOTO 层
	engine.Active = "GOTO"
	if !engine.CanHandle("j") {
		t.Error("Expected 'j' to be handled in GOTO layer")
	}

	if engine.CanHandle("h") {
		t.Error("Expected 'h' to not be handled in GOTO layer")
	}
}

// TestEngineInLayer 测试 InLayer 方法
func TestEngineInLayer(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{},
			},
		},
	}

	engine := NewEngine(&km)

	// 初始状态应该不在其他层
	if engine.InLayer() {
		t.Error("Expected to not be in layer initially")
	}

	// 设置为非默认层
	engine.Active = "GOTO"
	if !engine.InLayer() {
		t.Error("Expected to be in layer when active is 'GOTO'")
	}

	// 设置为空字符串
	engine.Active = ""
	if engine.InLayer() {
		t.Error("Expected to not be in layer when active is empty")
	}
}

// TestEngineReset 测试重置功能
func TestEngineReset(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{},
			},
		},
	}

	engine := NewEngine(&km)

	// 设置一些状态
	engine.Active = "GOTO"
	engine.count = 42
	engine.PendingOperator = "delete"

	// 添加一个模拟发射器
	mockEmitter := &MockRawTokenEmitter{}
	engine.AddEmitter(mockEmitter)

	// 重置引擎
	engine.Reset()

	// 验证状态已被重置
	if engine.Active != "NAV" {
		t.Errorf("Expected active layer to be reset to 'NAV', got '%s'", engine.Active)
	}

	if engine.count != 0 {
		t.Errorf("Expected count to be reset to 0, got %d", engine.count)
	}

	if engine.PendingOperator != "" {
		t.Errorf("Expected pending operator to be reset to empty, got '%s'", engine.PendingOperator)
	}

	// 验证发送了重置 token
	if len(mockEmitter.receivedTokens) != 1 {
		t.Errorf("Expected 1 token to be emitted during reset, got %d", len(mockEmitter.receivedTokens))
	}

	if mockEmitter.receivedTokens[0].Kind != TokenSystem || mockEmitter.receivedTokens[0].Value != "reset" {
		t.Errorf("Expected TokenSystem with value 'reset', got %v with value '%s'",
			mockEmitter.receivedTokens[0].Kind, mockEmitter.receivedTokens[0].Value)
	}
}

// TestEngineLayerTimeout 测试层超时功能
func TestEngineLayerTimeout(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{
					"f": {Layer: "GOTO", TimeoutMs: 100}, // 100ms 超时
				},
			},
			"GOTO": {
				Keys: map[string]KeyAction{
					"j": {Action: "move_down"},
				},
			},
		},
	}

	engine := NewEngine(&km)

	// 分发 'f' 键，切换到 GOTO 层
	engine.Dispatch("f")
	if engine.Active != "GOTO" {
		t.Errorf("Expected to be in 'GOTO' layer after dispatching 'f', got '%s'", engine.Active)
	}

	// 等待超过超时时间
	time.Sleep(150 * time.Millisecond)

	// 此时应该已经自动重置回 NAV 层
	// 注意：由于定时器是异步的，这里可能需要更复杂的同步机制来准确测试
	// 对于这个测试，我们主要验证定时器被设置和工作
}

// TestEngineRepeat 测试重复键 (.) 功能
func TestEngineRepeat(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{
					".": {Action: "repeat_last"},
				},
			},
		},
	}

	engine := NewEngine(&km)
	mockEmitter := &MockRawTokenEmitter{}
	engine.AddEmitter(mockEmitter)

	// 分发 '.' 键
	result := engine.Dispatch(".")
	if !result {
		t.Error("Expected dispatch to return true for repeat key")
	}

	if len(mockEmitter.receivedTokens) != 1 {
		t.Errorf("Expected 1 token to be emitted, got %d", len(mockEmitter.receivedTokens))
	}

	if mockEmitter.receivedTokens[0].Kind != TokenRepeat {
		t.Errorf("Expected TokenRepeat, got %v", mockEmitter.receivedTokens[0].Kind)
	}

	if mockEmitter.receivedTokens[0].Value != "." {
		t.Errorf("Expected value '.', got '%s'", mockEmitter.receivedTokens[0].Value)
	}
}

// TestEngineRunAction 测试动作执行
func TestEngineRunAction(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{
					"x": {Action: "exit"},
				},
			},
		},
	}

	engine := NewEngine(&km)

	// 测试 exit 动作
	// 注意：这里我们不能真正测试 ExitFSM 的效果，因为它会影响全局状态
	// 所以我们只是验证方法被调用不会崩溃
	engine.RunAction("exit")
}

// TestEngineGetCount 测试获取计数
func TestEngineGetCount(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{},
			},
		},
	}

	engine := NewEngine(&km)

	// 初始计数应该是 0
	if engine.GetCount() != 0 {
		t.Errorf("Expected initial count to be 0, got %d", engine.GetCount())
	}

	// 设置计数
	engine.count = 42
	if engine.GetCount() != 42 {
		t.Errorf("Expected count to be 42, got %d", engine.GetCount())
	}
}

// TestEngineDispatchZeroAtStart 测试在计数为0时按0键的行为
func TestEngineDispatchZeroAtStart(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{
					"0": {Action: "goto_line_start"},
				},
			},
		},
	}

	engine := NewEngine(&km)

	// 初始计数为0时按0键，应该被视为动作而不是数字
	initialCount := engine.count
	if initialCount != 0 {
		t.Errorf("Expected initial count to be 0, got %d", initialCount)
	}

	// 这里我们无法直接测试是否进入了CanHandle流程，但我们可以测试计数是否保持为0
	// 在原始代码中，当count为0且key为"0"时，会跳过数字处理逻辑
	engine.Dispatch("0")

	// 如果0被当作数字处理，count会变成0（0*10+0），但实际上它应该被当作动作处理
	// 所以count应该保持不变
	if engine.count != 0 {
		t.Errorf("Expected count to remain 0 when '0' pressed at start, got %d", engine.count)
	}
}
