package core

import (
	"testing"
)

// TestIntentKindString 测试IntentKind的String方法
func TestIntentKindString(t *testing.T) {
	testCases := []struct {
		kind     IntentKind
		expected string
	}{
		{IntentMove, "MOVE"},
		{IntentDelete, "DELETE"},
		{IntentChange, "CHANGE"},
		{IntentYank, "YANK"},
		{IntentInsert, "INSERT"},
		{IntentPaste, "PASTE"},
		{IntentUndo, "UNDO"},
		{IntentRedo, "REDO"},
		{IntentSearch, "SEARCH"},
		{IntentVisual, "VISUAL"},
		{IntentToggleCase, "TOGGLE_CASE"},
		{IntentReplace, "REPLACE"},
		{IntentRepeat, "REPEAT"},
		{IntentFind, "FIND"},
		{IntentExit, "EXIT"},
		{IntentCount, "COUNT"},
		{IntentOperator, "OPERATOR"},
		{IntentMotion, "MOTION"},
		{IntentMacro, "MACRO"},
		{IntentEnterVisual, "ENTER_VISUAL"},
		{IntentExitVisual, "EXIT_VISUAL"},
		{IntentExtendSelection, "EXTEND_SELECTION"},
		{IntentOperatorSelection, "OPERATOR_SELECTION"},
		{IntentRepeatFind, "REPEAT_FIND"},
		{IntentRepeatFindReverse, "REPEAT_FIND_REVERSE"},
		{IntentKind(-1), "NONE"}, // 测试默认情况
	}

	for _, tc := range testCases {
		result := tc.kind.String()
		if result != tc.expected {
			t.Errorf("Expected IntentKind(%d).String() to return '%s', got '%s'", tc.kind, tc.expected, result)
		}
	}
}

// TestTargetKindString 测试TargetKind的String方法
func TestTargetKindString(t *testing.T) {
	testCases := []struct {
		kind     TargetKind
		expected string
	}{
		{TargetChar, "CHAR"},
		{TargetWord, "WORD"},
		{TargetLine, "LINE"},
		{TargetFile, "FILE"},
		{TargetTextObject, "TEXT_OBJECT"},
		{TargetPosition, "POSITION"},
		{TargetSearch, "SEARCH"},
		{TargetKind(-1), "UNKNOWN"}, // 测试默认情况
	}

	for _, tc := range testCases {
		result := tc.kind.String()
		if result != tc.expected {
			t.Errorf("Expected TargetKind(%d).String() to return '%s', got '%s'", tc.kind, tc.expected, result)
		}
	}
}

// TestSemanticTarget 测试语义目标结构
func TestSemanticTarget(t *testing.T) {
	st := SemanticTarget{
		Kind:      TargetWord,
		Direction: "forward",
		Scope:     "inner",
		Value:     "test",
	}

	if st.Kind != TargetWord {
		t.Errorf("Expected Kind to be TargetWord, got %v", st.Kind)
	}

	if st.Direction != "forward" {
		t.Errorf("Expected Direction to be 'forward', got '%s'", st.Direction)
	}

	if st.Scope != "inner" {
		t.Errorf("Expected Scope to be 'inner', got '%s'", st.Scope)
	}

	if st.Value != "test" {
		t.Errorf("Expected Value to be 'test', got '%s'", st.Value)
	}
}

// TestEvidenceMeta 测试证据元数据结构
func TestEvidenceMeta(t *testing.T) {
	meta := EvidenceMeta{
		Hash:      "abc123",
		Offset:    100,
		Timestamp: 1234567890,
		Size:      512,
	}

	if meta.Hash != "abc123" {
		t.Errorf("Expected Hash to be 'abc123', got '%s'", meta.Hash)
	}

	if meta.Offset != 100 {
		t.Errorf("Expected Offset to be 100, got %d", meta.Offset)
	}

	if meta.Timestamp != 1234567890 {
		t.Errorf("Expected Timestamp to be 1234567890, got %d", meta.Timestamp)
	}

	if meta.Size != 512 {
		t.Errorf("Expected Size to be 512, got %d", meta.Size)
	}
}
