package kernel

import (
	"context"
	"testing"
	"tmux-fsm/fsm"
	"tmux-fsm/intent"
)

// MockIntentExecutor 用于测试的模拟执行器
type MockIntentExecutor struct {
	processedIntent *intent.Intent
	processError    error
}

func (m *MockIntentExecutor) Process(intent *intent.Intent) error {
	m.processedIntent = intent
	return m.processError
}

// MockContextualIntentExecutor 用于测试的模拟上下文执行器
type MockContextualIntentExecutor struct {
	processedIntent *intent.Intent
	processError    error
}

func (m *MockContextualIntentExecutor) ProcessWithContext(ctx context.Context, hctx HandleContext, intent *intent.Intent) error {
	m.processedIntent = intent
	return m.processError
}

func (m *MockContextualIntentExecutor) Process(intent *intent.Intent) error {
	m.processedIntent = intent
	return m.processError
}

// TestNewKernel 测试Kernel创建
func TestNewKernel(t *testing.T) {
	fsmEngine := fsm.NewEngine(nil)
	executor := &MockIntentExecutor{}

	kernel := NewKernel(fsmEngine, executor)

	if kernel.FSM != fsmEngine {
		t.Errorf("Expected FSM to be set correctly")
	}

	if kernel.Exec != executor {
		t.Errorf("Expected executor to be set correctly")
	}

	if kernel.Grammar == nil {
		t.Errorf("Expected Grammar to be initialized")
	}

	if kernel.NativeBuilder == nil {
		t.Errorf("Expected NativeBuilder to be initialized")
	}

	if !kernel.ShadowIntent {
		t.Errorf("Expected ShadowIntent to be true by default")
	}
}

// TestKernelHandleContext 测试HandleContext结构
func TestKernelHandleContext(t *testing.T) {
	ctx := HandleContext{
		Ctx:       context.Background(),
		RequestID: "req-test",
		ActorID:   "actor-test",
	}

	if ctx.RequestID != "req-test" {
		t.Errorf("Expected RequestID to be 'req-test', got '%s'", ctx.RequestID)
	}

	if ctx.ActorID != "actor-test" {
		t.Errorf("Expected ActorID to be 'actor-test', got '%s'", ctx.ActorID)
	}
}

// TestKernelGetPendingOp 测试获取待处理操作符
func TestKernelGetPendingOp(t *testing.T) {
	fsmEngine := fsm.NewEngine(nil)
	executor := &MockIntentExecutor{}
	kernel := NewKernel(fsmEngine, executor)

	// 初始状态下，待处理操作符应为空
	op := kernel.GetPendingOp()
	if op != "" {
		t.Errorf("Expected pending op to be empty initially, got '%s'", op)
	}
}

// TestKernelGetCount 测试获取计数
func TestKernelGetCount(t *testing.T) {
	// 创建一个带keymap的FSM引擎
	km := &fsm.Keymap{
		Initial: "NAV",
		States: map[string]fsm.StateDef{
			"NAV": {
				Keys: map[string]fsm.KeyAction{},
			},
		},
	}
	fsmEngine := fsm.NewEngine(km)
	executor := &MockIntentExecutor{}
	kernel := NewKernel(fsmEngine, executor)

	// 初始状态下，计数应为0
	count := kernel.GetCount()
	if count != 0 {
		t.Errorf("Expected count to be 0 initially, got %d", count)
	}

	// 设置FSM计数
	fsmEngine.Dispatch("2")
	count = kernel.GetCount()
	if count != 2 {
		t.Errorf("Expected count to be 2 after dispatching '2', got %d", count)
	}
}

// TestKernelProcessIntent 测试处理意图
func TestKernelProcessIntent(t *testing.T) {
	fsmEngine := fsm.NewEngine(nil)
	executor := &MockIntentExecutor{}
	kernel := NewKernel(fsmEngine, executor)

	testIntent := &intent.Intent{
		Kind:   intent.IntentInsert,
		Count:  1,
		PaneID: "test-pane",
	}

	err := kernel.ProcessIntent(testIntent)
	if err != nil {
		t.Errorf("Expected ProcessIntent to succeed, got error: %v", err)
	}

	if executor.processedIntent == nil {
		t.Errorf("Expected executor to receive intent")
	}

	if executor.processedIntent.Kind != intent.IntentInsert {
		t.Errorf("Expected processed intent to be INSERT, got %v", executor.processedIntent.Kind)
	}
}

// TestKernelProcessIntentWithContext 测试处理意图with上下文
func TestKernelProcessIntentWithContext(t *testing.T) {
	fsmEngine := fsm.NewEngine(nil)
	executor := &MockContextualIntentExecutor{}
	kernel := NewKernel(fsmEngine, executor)

	testIntent := &intent.Intent{
		Kind:   intent.IntentDelete,
		Count:  3,
		PaneID: "test-pane",
	}

	hctx := HandleContext{
		Ctx:       context.Background(),
		RequestID: "req-test",
		ActorID:   "actor-test",
	}

	err := kernel.ProcessIntentWithContext(hctx, testIntent)
	if err != nil {
		t.Errorf("Expected ProcessIntentWithContext to succeed, got error: %v", err)
	}

	if executor.processedIntent == nil {
		t.Errorf("Expected executor to receive intent")
	}

	if executor.processedIntent.Kind != intent.IntentDelete {
		t.Errorf("Expected processed intent to be DELETE, got %v", executor.processedIntent.Kind)
	}
}

// TestDecisionKindString 测试DecisionKind的String方法
func TestDecisionKindString(t *testing.T) {
	testCases := []struct {
		kind     DecisionKind
		expected string
	}{
		{DecisionNone, "None"},
		{DecisionFSM, "FSM"},
		{DecisionLegacy, "Legacy"},
		{DecisionIntent, "Intent"},
		{DecisionKind(-1), "Unknown"}, // 测试默认情况
	}

	for _, tc := range testCases {
		result := tc.kind.String()
		if result != tc.expected {
			t.Errorf("Expected DecisionKind(%d).String() to return '%s', got '%s'", tc.kind, tc.expected, result)
		}
	}
}

// TestDecisionStruct 测试Decision结构
func TestDecisionStruct(t *testing.T) {
	intentObj := &intent.Intent{
		Kind: intent.IntentMove,
	}

	decision := &Decision{
		Kind:   DecisionIntent,
		Intent: intentObj,
		Action: "move_left",
	}

	if decision.Kind != DecisionIntent {
		t.Errorf("Expected Kind to be DecisionIntent, got %v", decision.Kind)
	}

	if decision.Intent == nil {
		t.Errorf("Expected Intent to be set")
	}

	if decision.Action != "move_left" {
		t.Errorf("Expected Action to be 'move_left', got '%s'", decision.Action)
	}
}
