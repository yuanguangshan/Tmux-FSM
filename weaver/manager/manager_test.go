package manager

import (
	"testing"
	"tmux-fsm/intent"
	"tmux-fsm/weaver/core"
)

// MockIntent 用于测试的模拟意图
type MockIntent struct {
	kind         core.IntentKind
	count        int
	paneID       string
	snapshotHash string
	allowPartial bool
}

func (m *MockIntent) GetKind() core.IntentKind {
	return m.kind
}

func (m *MockIntent) GetTarget() core.SemanticTarget {
	return core.SemanticTarget{}
}

func (m *MockIntent) GetCount() int {
	return m.count
}

func (m *MockIntent) GetMeta() map[string]interface{} {
	return nil
}

func (m *MockIntent) GetPaneID() string {
	return m.paneID
}

func (m *MockIntent) GetSnapshotHash() string {
	return m.snapshotHash
}

func (m *MockIntent) IsPartialAllowed() bool {
	return m.allowPartial
}

func (m *MockIntent) GetAnchors() []core.Anchor {
	return nil
}

func (m *MockIntent) GetOperator() *int {
	return nil
}

// TestInitWeaver 测试Weaver初始化
func TestInitWeaver(t *testing.T) {
	// 测试不同模式下的初始化
	InitWeaver(ModeLegacy)
	if weaverMgr != nil {
		t.Errorf("Expected weaverMgr to be nil in Legacy mode")
	}

	InitWeaver(ModeWeaver)
	if weaverMgr == nil {
		t.Errorf("Expected weaverMgr to be initialized in Weaver mode")
	}

	InitWeaver(ModeShadow)
	if weaverMgr == nil {
		t.Errorf("Expected weaverMgr to be initialized in Shadow mode")
	}
}

// TestConvertToCoreIntent 测试意图转换
func TestConvertToCoreIntent(t *testing.T) {
	// 创建一个统一的intent.Intent
	originalIntent := &intent.Intent{
		Kind:   intent.IntentDelete,
		Count:  3,
		PaneID: "pane1",
	}

	// 转换为core.Intent
	coreIntent := convertToCoreIntent(originalIntent)

	if coreIntent.GetKind() != core.IntentKind(intent.IntentDelete) {
		t.Errorf("Expected converted intent kind to be %d, got %d", 
			core.IntentKind(intent.IntentDelete), coreIntent.GetKind())
	}

	if coreIntent.GetCount() != 3 {
		t.Errorf("Expected converted intent count to be 3, got %d", coreIntent.GetCount())
	}

	if coreIntent.GetPaneID() != "pane1" {
		t.Errorf("Expected converted intent paneID to be 'pane1', got '%s'", coreIntent.GetPaneID())
	}
}

// TestGetWeaverManager 测试获取Weaver管理器
func TestGetWeaverManager(t *testing.T) {
	// 先初始化
	InitWeaver(ModeWeaver)

	mgr := GetWeaverManager()
	if mgr == nil {
		t.Errorf("Expected GetWeaverManager to return non-nil manager")
	}
}

// TestWeaverManagerProcess 测试Weaver管理器处理意图
func TestWeaverManagerProcess(t *testing.T) {
	// 初始化管理器
	InitWeaver(ModeWeaver)

	mgr := GetWeaverManager()
	if mgr == nil {
		t.Fatal("Failed to initialize weaver manager")
	}

	// 创建一个测试意图
	testIntent := &intent.Intent{
		Kind:   intent.IntentInsert,
		Count:  1,
		PaneID: "test-pane",
	}

	// 尝试处理意图（在测试环境中，这可能会失败，但不应该panic）
	err := mgr.Process(testIntent)
	// 注意：在测试环境中，由于没有实际的Tmux环境，这可能会返回错误
	// 但我们至少要确保它不会panic
	if err != nil {
		// 这是可以接受的，因为测试环境中没有实际的Tmux
		t.Logf("Process returned error (expected in test environment): %v", err)
	}
}
