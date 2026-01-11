package intent

import (
	"testing"
)

// TestIntentCreation 测试意图创建
func TestIntentCreation(t *testing.T) {
	intent := Intent{
		Kind:   IntentDelete,
		Count:  3,
		PaneID: "pane1",
	}

	if intent.Kind != IntentDelete {
		t.Errorf("Expected Kind to be IntentDelete, got %v", intent.Kind)
	}

	if intent.Count != 3 {
		t.Errorf("Expected Count to be 3, got %d", intent.Count)
	}

	if intent.PaneID != "pane1" {
		t.Errorf("Expected PaneID to be 'pane1', got '%s'", intent.PaneID)
	}
}

// TestIntentGetters 测试意图获取器
func TestIntentGetters(t *testing.T) {
	intent := Intent{
		Kind:         IntentInsert,
		Count:        5,
		PaneID:       "pane2",
		SnapshotHash: "abc123",
		AllowPartial: true,
	}

	if intent.GetKind() != IntentInsert {
		t.Errorf("Expected GetKind() to return IntentInsert, got %v", intent.GetKind())
	}

	if intent.GetCount() != 5 {
		t.Errorf("Expected GetCount() to return 5, got %d", intent.GetCount())
	}

	if intent.GetPaneID() != "pane2" {
		t.Errorf("Expected GetPaneID() to return 'pane2', got '%s'", intent.GetPaneID())
	}

	if intent.GetSnapshotHash() != "abc123" {
		t.Errorf("Expected GetSnapshotHash() to return 'abc123', got '%s'", intent.GetSnapshotHash())
	}

	if !intent.IsPartialAllowed() {
		t.Errorf("Expected IsPartialAllowed() to return true")
	}
}

// TestIntentWithMotion 测试带有Motion的意图
func TestIntentWithMotion(t *testing.T) {
	motion := &Motion{
		Kind:  MotionWord,
		Count: 2,
	}

	intent := Intent{
		Kind:   IntentDelete,
		Motion: motion,
		Count:  1,
	}

	if intent.Motion == nil {
		t.Fatal("Expected Motion to be set")
	}

	if intent.Motion.Kind != MotionWord {
		t.Errorf("Expected Motion.Kind to be MotionWord, got %v", intent.Motion.Kind)
	}

	if intent.Motion.Count != 2 {
		t.Errorf("Expected Motion.Count to be 2, got %d", intent.Motion.Count)
	}
}

// TestIntentWithOperator 测试带有Operator的意图
func TestIntentWithOperator(t *testing.T) {
	op := OpDelete
	intent := Intent{
		Kind:     IntentOperator,
		Operator: &op,
		Count:    1,
	}

	if intent.Operator == nil {
		t.Fatal("Expected Operator to be set")
	}

	if *intent.Operator != OpDelete {
		t.Errorf("Expected Operator to be OpDelete, got %v", *intent.Operator)
	}

	// 测试GetOperator方法
	opPtr := intent.GetOperator()
	if opPtr == nil {
		t.Fatal("Expected GetOperator() to return non-nil")
	}

	if *opPtr != int(OpDelete) {
		t.Errorf("Expected GetOperator() to return %d, got %d", int(OpDelete), *opPtr)
	}
}

// TestIntentWithEmptyOperator 测试空Operator的意图
func TestIntentWithEmptyOperator(t *testing.T) {
	intent := Intent{
		Kind: IntentMove,
		Count: 1,
	}

	// Operator为nil时，GetOperator应该返回nil
	opPtr := intent.GetOperator()
	if opPtr != nil {
		t.Errorf("Expected GetOperator() to return nil when Operator is nil, got %v", *opPtr)
	}
}
