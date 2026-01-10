package main

import (
	"testing"
	"github.com/stretchr/testify/require"
)

// testSnapshot 创建测试用的快照
func testSnapshot() Snapshot {
	return Snapshot{
		ID: "test-snapshot-1",
		Lines: []LineSnapshot{
			{ID: "L1", Text: "hello world"},
			{ID: "L2", Text: "second line"},
			{ID: "L3", Text: "third line here"},
		},
	}
}

// TestResolve_LegacyDeleteWord 测试解析遗留的删除单词意图
func TestResolve_LegacyDeleteWord(t *testing.T) {
	snap := testSnapshot()
	
	intent := Intent{
		Kind: IntentDelete,
		Target: SemanticTarget{
			Kind: TargetWord,
		},
		Anchors: []Anchor{
			{
				PaneID: "p1",
				LineID: "legacy::pane::p1::row::0::time::123456789",
				Start:  6,
				End:    11,
				Kind:   int(TargetWord),
			},
		},
		PaneID: "p1",
	}

	ctx := ResolveContext{
		Snapshot: snap,
		Cursor:   CursorState{LineID: "L1", Offset: 6},
	}

	resolved, err := ResolveIntent(ctx, intent)

	require.NoError(t, err)
	require.Equal(t, 1, len(resolved.Anchors))
	require.Equal(t, "L1", resolved.Anchors[0].LineID)
	require.Equal(t, 6, resolved.Anchors[0].Range.Start)
	require.Equal(t, 11, resolved.Anchors[0].Range.End)
	require.Equal(t, AnchorOriginLegacy, resolved.Anchors[0].Origin)
}

// TestResolve_NativeDeleteWord 测试解析原生的删除单词意图
func TestResolve_NativeDeleteWord(t *testing.T) {
	snap := testSnapshot()

	intent := Intent{
		Kind: IntentDelete,
		Target: SemanticTarget{
			Kind: TargetWord,
		},
		Count: 1,
		Anchors: []Anchor{
			CursorAnchor(CursorRef{Kind: CursorPrimary}),
		},
		PaneID: "p1",
	}

	ctx := ResolveContext{
		Snapshot: snap,
		Cursor:   CursorState{LineID: "L1", Offset: 6},
	}

	resolved, err := ResolveIntent(ctx, intent)

	require.NoError(t, err)
	require.Equal(t, 1, len(resolved.Anchors))
	// 确保没有遗留锚点泄漏
	require.NotEqual(t, AnchorOriginLegacy, resolved.Anchors[0].Origin)
}

// TestResolve_NativeMove 测试解析原生的移动意图
func TestResolve_NativeMove(t *testing.T) {
	snap := testSnapshot()

	intent := Intent{
		Kind: IntentMove,
		Target: SemanticTarget{
			Kind:      TargetWord,
			Direction: "forward",
		},
		Count: 1,
		Anchors: []Anchor{
			CursorAnchor(CursorRef{Kind: CursorPrimary}),
		},
		PaneID: "p1",
	}

	ctx := ResolveContext{
		Snapshot: snap,
		Cursor:   CursorState{LineID: "L1", Offset: 0}, // 从 "hello" 开始
	}

	resolved, err := ResolveIntent(ctx, intent)

	require.NoError(t, err)
	require.Equal(t, IntentMove, resolved.Kind)
	require.Equal(t, 1, len(resolved.Anchors))
	// 确保没有遗留锚点泄漏
	require.NotEqual(t, AnchorOriginLegacy, resolved.Anchors[0].Origin)
}

// TestResolve_LegacyMove 测试解析遗留的移动意图
func TestResolve_LegacyMove(t *testing.T) {
	snap := testSnapshot()

	intent := Intent{
		Kind: IntentMove,
		Target: SemanticTarget{
			Kind:      TargetWord,
			Direction: "forward",
		},
		Anchors: []Anchor{
			{
				PaneID: "p1",
				LineID: "legacy::pane::p1::row::0::time::123456789",
				Start:  0,
				End:    5, // "hello"
				Kind:   int(TargetWord),
			},
		},
		PaneID: "p1",
	}

	ctx := ResolveContext{
		Snapshot: snap,
		Cursor:   CursorState{LineID: "L1", Offset: 0},
	}

	resolved, err := ResolveIntent(ctx, intent)

	require.NoError(t, err)
	require.Equal(t, IntentMove, resolved.Kind)
	require.Equal(t, 1, len(resolved.Anchors))
	require.Equal(t, AnchorOriginLegacy, resolved.Anchors[0].Origin)
}

// TestResolvedIntent_NoLegacyLeak 测试防止遗留锚点泄漏
func TestResolvedIntent_NoLegacyLeak(t *testing.T) {
	// 创建一个包含遗留锚点的解析后意图
	resolved := ResolvedIntent{
		Intent: Intent{
			Kind: IntentDelete,
		},
		Anchors: []ResolvedAnchor{
			{
				LineID: "L1",
				Origin: AnchorOriginLegacy, // 故意设置为遗留类型
			},
		},
	}

	// 这里我们测试断言函数
	// 在实际使用中，这个函数会在解析完成后被调用
	defer func() {
		if r := recover(); r != nil {
			// 预期会有 panic，因为我们故意设置了遗留锚点
			require.Equal(t, "legacy anchor leaked past resolver", r)
		}
	}()
	
	// 这会触发 panic，因为我们有遗留锚点
	resolved.AssertNoLegacy()
	
	// 如果没有 panic，测试失败
	t.Error("Expected panic from AssertNoLegacy due to legacy anchor")
}

// TestResolve_UndoIntent 测试解析撤销意图
func TestResolve_UndoIntent(t *testing.T) {
	snap := testSnapshot()

	intent := Intent{
		Kind:   IntentUndo,
		PaneID: "p1",
		Anchors: []Anchor{
			CursorAnchor(CursorRef{Kind: CursorPrimary}),
		},
	}

	ctx := ResolveContext{
		Snapshot: snap,
		Cursor:   CursorState{LineID: "L1", Offset: 5},
	}

	resolved, err := ResolveIntent(ctx, intent)

	require.NoError(t, err)
	require.Equal(t, IntentUndo, resolved.Kind)
	// Undo 意图应该有锚点用于投影兼容性
	require.Equal(t, 1, len(resolved.Anchors))
}

// TestResolve_RedoIntent 测试解析重做意图
func TestResolve_RedoIntent(t *testing.T) {
	snap := testSnapshot()

	intent := Intent{
		Kind:   IntentRedo,
		PaneID: "p1",
		Anchors: []Anchor{
			CursorAnchor(CursorRef{Kind: CursorPrimary}),
		},
	}

	ctx := ResolveContext{
		Snapshot: snap,
		Cursor:   CursorState{LineID: "L1", Offset: 5},
	}

	resolved, err := ResolveIntent(ctx, intent)

	require.NoError(t, err)
	require.Equal(t, IntentRedo, resolved.Kind)
	// Redo 意图应该有锚点用于投影兼容性
	require.Equal(t, 1, len(resolved.Anchors))
}

// TestIsLegacyAnchor_Detection 测试遗留锚点检测
func TestIsLegacyAnchor_Detection(t *testing.T) {
	// 测试遗留锚点
	legacyAnchor := Anchor{
		LineID: "legacy::pane::p1::row::0::time::123456789",
	}
	require.True(t, isLegacyAnchor(legacyAnchor))

	// 测试原生锚点
	nativeAnchor := Anchor{
		LineID: "L123456789",
	}
	require.False(t, isLegacyAnchor(nativeAnchor))

	// 测试空锚点
	emptyAnchor := Anchor{}
	require.False(t, isLegacyAnchor(emptyAnchor))
}