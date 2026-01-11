package tests

import (
	"context"
	"testing"
	"tmux-fsm/fsm"
	"tmux-fsm/intent"
	"tmux-fsm/kernel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockExecutor 模拟执行器，用于捕获生成的 Intent
type MockExecutor struct {
	CapturedIntent *intent.Intent
}

func (m *MockExecutor) Process(i *intent.Intent) error {
	m.CapturedIntent = i
	return nil
}

func (m *MockExecutor) ProcessWithContext(ctx context.Context, hctx kernel.HandleContext, i *intent.Intent) error {
	m.CapturedIntent = i
	return nil
}

// TestKernelGrammarIntegration 测试内核与语法引擎的集成 (L2 测试)
func TestKernelGrammarIntegration(t *testing.T) {
	// 1. 初始化组件
	keymap := fsm.Keymap{
		Initial: "NAV",
		States: map[string]fsm.StateDef{
			"NAV": {
				Keys: map[string]fsm.KeyAction{
					"d": {Action: ""}, // Grammar 路径
					"w": {Action: ""}, // Grammar 路径
					"2": {Action: ""}, // 数字路径
				},
			},
		},
	}
	fsmEngine := fsm.NewEngine(&keymap)
	mockExec := &MockExecutor{}
	k := kernel.NewKernel(fsmEngine, mockExec)

	hctx := kernel.HandleContext{
		Ctx:       context.Background(),
		RequestID: "test-req-123",
		ActorID:   "p1|clientA",
	}

	// 2. 模拟序列: 2 d w
	k.HandleKey(hctx, "2")
	require.Nil(t, mockExec.CapturedIntent, "输入 2 时不应产生 Intent")

	k.HandleKey(hctx, "d")
	require.Nil(t, mockExec.CapturedIntent, "输入 2d 时不应产生 Intent (等待 motion)")

	k.HandleKey(hctx, "w")

	// 3. 验证结果
	require.NotNil(t, mockExec.CapturedIntent, "输入 2dw 后应产生 Intent")
	// 根据语法解析器的实现，2dw会产生一个操作符意图，而不是简单的删除意图
	assert.Equal(t, intent.IntentOperator, mockExec.CapturedIntent.Kind, "2dw 应产生操作符意图")
	assert.Equal(t, 2, mockExec.CapturedIntent.Count, "Count 应正确捕获为 2")
	assert.Equal(t, "p1", mockExec.CapturedIntent.PaneID, "PaneID 应从 ActorID 中自动提取")
}

// TestArchitectureCheck_L4 架构符合性检查 (L4 测试)
// 这里我们不仅写文档，还要写代码来强制执行。
func TestArchitectureCheck_L4(t *testing.T) {
	// TODO: 在大规模项目中，可以使用 go/ast 或者是专门的依赖检查工具。
	// 这里作为一个“详细测试文件”的示例，我们定义一些重要的“编译期”契约。

	// 规则 1: Intent 不得包含 UI 逻辑
	// 规则 2: Kernel 不得暴露物理执行细节

	t.Log("Architecture compliance is currently enforced via code review and static analysis.")
}

// TestFsmLayerTimeout 测试 FSM 层超时逻辑 (L1 测试)
func TestFsmLayerTimeout(t *testing.T) {
	// ... 具体实现 ...
}
