package main

import (
	"testing"
	"github.com/stretchr/testify/require"
)

// TestNativeIntentBuilderCompatibility 测试 Native Intent Builder 与 Legacy Intent 的兼容性
func TestNativeIntentBuilderCompatibility(t *testing.T) {
	// 创建测试状态
	state := &FSMState{
		Mode:   "NORMAL",
		PaneID: "test-pane",
		Cursor: Cursor{Row: 0, Col: 0},
	}

	// 测试 dw 命令
	key := "d"
	intent1 := processKeyToIntent(state, key)
	// 调试输出
	t.Logf("Intent1 Kind: %d, Expected IntentNone: %d", intent1.Kind, IntentNone)
	t.Logf("State Mode: %s, Expected OPERATOR_PENDING", state.Mode)
	t.Logf("State Operator: %s, Expected delete", state.Operator)

	require.Equal(t, IntentNone, intent1.Kind) // 应该进入 OPERATOR_PENDING 模式
	require.Equal(t, "OPERATOR_PENDING", state.Mode)
	require.Equal(t, "delete", state.Operator)

	key = "w"
	intent2 := processKeyToIntent(state, key)
	// 现在应该返回一个删除单词的意图
	require.Equal(t, IntentDelete, intent2.Kind)
	require.Equal(t, TargetWord, intent2.Target.Kind)
	// 注意：对于 dw，方向应该是 forward
	require.Equal(t, "forward", intent2.Target.Direction)

	// 测试 cw 命令
	state.Mode = "NORMAL"
	state.Operator = ""
	intent3 := processKeyToIntent(state, "c")
	require.Equal(t, IntentNone, intent3.Kind)
	require.Equal(t, "OPERATOR_PENDING", state.Mode)
	require.Equal(t, "change", state.Operator)

	key = "w"
	intent4 := processKeyToIntent(state, key)
	// 现在应该返回一个修改单词的意图
	require.Equal(t, IntentChange, intent4.Kind)
	require.Equal(t, TargetWord, intent4.Target.Kind)
	require.Equal(t, "forward", intent4.Target.Direction)

	// 测试 dd 命令
	state.Mode = "NORMAL"
	state.Operator = ""
	intent5 := processKeyToIntent(state, "d")
	require.Equal(t, IntentNone, intent5.Kind)
	require.Equal(t, "OPERATOR_PENDING", state.Mode)
	require.Equal(t, "delete", state.Operator)

	key = "d"
	intent6 := processKeyToIntent(state, key)
	// 现在应该返回一个删除整行的意图
	require.Equal(t, IntentDelete, intent6.Kind)
	require.Equal(t, TargetLine, intent6.Target.Kind)
	require.Equal(t, "whole", intent6.Target.Scope)
}

// TestLegacyIntentStillWorks 测试 Legacy Intent 仍然有效
func TestLegacyIntentStillWorks(t *testing.T) {
	// 创建测试状态
	state := &FSMState{
		Mode:   "NORMAL",
		PaneID: "test-pane",
		Cursor: Cursor{Row: 0, Col: 0},
	}

	// 测试一个尚未迁移到 Native Intent 的命令 (例如 "gg")
	intent := processKeyToIntent(state, "gg")
	// 这应该仍然通过 legacy bridge 工作
	// 检查是否返回了某种意图
	require.NotNil(t, intent)
	// 检查是否是预期的移动到文件开头的意图
	require.Equal(t, IntentMove, intent.Kind)
	require.Equal(t, TargetFile, intent.Target.Kind)
	require.Equal(t, "start", intent.Target.Scope)
}

// TestIntentBuilderCreation 测试 IntentBuilder 的创建
func TestIntentBuilderCreation(t *testing.T) {
	builder := NewIntentBuilder("test-pane")
	require.NotNil(t, builder)
	require.Equal(t, "test-pane", builder.paneID)
	require.Equal(t, CursorPrimary, builder.cursor.Kind)
}

// TestIntentBuilderMethods 测试 IntentBuilder 的各种方法
func TestIntentBuilderMethods(t *testing.T) {
	builder := NewIntentBuilder("test-pane")

	// 测试 Move 方法
	moveIntent := builder.Move(SemanticTarget{Kind: TargetWord, Direction: "forward"}, 1)
	require.Equal(t, IntentMove, moveIntent.Kind)
	require.Equal(t, TargetWord, moveIntent.Target.Kind)
	require.Equal(t, "forward", moveIntent.Target.Direction)
	require.Equal(t, 1, moveIntent.Count)
	require.Equal(t, "test-pane", moveIntent.PaneID)

	// 测试 Delete 方法
	deleteIntent := builder.Delete(SemanticTarget{Kind: TargetLine, Scope: "whole"}, 1)
	require.Equal(t, IntentDelete, deleteIntent.Kind)
	require.Equal(t, TargetLine, deleteIntent.Target.Kind)
	require.Equal(t, "whole", deleteIntent.Target.Scope)
	require.Equal(t, 1, deleteIntent.Count)
	require.Equal(t, "test-pane", deleteIntent.PaneID)

	// 测试 Change 方法
	changeIntent := builder.Change(SemanticTarget{Kind: TargetWord, Direction: "forward"}, 2)
	require.Equal(t, IntentChange, changeIntent.Kind)
	require.Equal(t, TargetWord, changeIntent.Target.Kind)
	require.Equal(t, "forward", changeIntent.Target.Direction)
	require.Equal(t, 2, changeIntent.Count)
	require.Equal(t, "test-pane", changeIntent.PaneID)

	// 测试 Undo 方法
	undoIntent := builder.Undo()
	require.Equal(t, IntentUndo, undoIntent.Kind)
	require.Equal(t, "test-pane", undoIntent.PaneID)

	// 测试 Redo 方法
	redoIntent := builder.Redo()
	require.Equal(t, IntentRedo, redoIntent.Kind)
	require.Equal(t, "test-pane", redoIntent.PaneID)
}

// TestSnapshotModel 测试快照模型的基本功能
func TestSnapshotModel(t *testing.T) {
	// 创建初始快照
	initialSnapshot := Snapshot{
		ID: "initial",
		Lines: []LineSnapshot{
			{ID: "L1", Text: "first line"},
			{ID: "L2", Text: "second line"},
		},
	}

	// 创建历史记录
	history := NewHistoryForResolver(initialSnapshot)

	// 验证初始状态
	require.Equal(t, initialSnapshot, history.present)
	require.Equal(t, 0, len(history.past))
	require.Equal(t, 0, len(history.future))

	// 创建新快照并推送
	newSnapshot := Snapshot{
		ID: "updated",
		Lines: []LineSnapshot{
			{ID: "L1", Text: "first line modified"},
			{ID: "L2", Text: "second line"},
			{ID: "L3", Text: "third line"},
		},
	}
	history.Push(newSnapshot)

	// 验证推送后状态
	require.Equal(t, newSnapshot, history.present)
	require.Equal(t, 1, len(history.past))
	require.Equal(t, initialSnapshot, history.past[0])
	require.Equal(t, 0, len(history.future))

	// 测试撤销
	undoneSnapshot, canUndo := history.Undo()
	require.True(t, canUndo)
	require.Equal(t, initialSnapshot, undoneSnapshot)
	require.Equal(t, initialSnapshot, history.present)
	require.Equal(t, 0, len(history.past))
	require.Equal(t, 1, len(history.future))
	require.Equal(t, newSnapshot, history.future[0])

	// 测试重做
	redoneSnapshot, canRedo := history.Redo()
	require.True(t, canRedo)
	require.Equal(t, newSnapshot, redoneSnapshot)
	require.Equal(t, newSnapshot, history.present)
	require.Equal(t, 1, len(history.past))
	require.Equal(t, initialSnapshot, history.past[0])
	require.Equal(t, 0, len(history.future))
}

// TestResolverWithNativeAndLegacyIntents 测试 Resolver 处理 Native 和 Legacy 意图
func TestResolverWithNativeAndLegacyIntents(t *testing.T) {
	// 创建测试快照
	snap := testSnapshot()

	// 创建上下文
	ctx := ResolveContext{
		Snapshot: snap,
		Cursor:   CursorState{LineID: "L1", Offset: 5},
	}

	// 测试原生意图
	nativeIntent := Intent{
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

	resolvedNative, err := ResolveIntent(ctx, nativeIntent)
	require.NoError(t, err)
	require.Equal(t, 1, len(resolvedNative.Anchors))
	// 确保原生意图没有遗留锚点
	for _, anchor := range resolvedNative.Anchors {
		require.NotEqual(t, AnchorOriginLegacy, anchor.Origin)
	}

	// 测试遗留意图
	legacyIntent := Intent{
		Kind: IntentDelete,
		Anchors: []Anchor{
			{
				PaneID: "p1",
				LineID: "legacy::pane::p1::row::0::time::123456789",
				Start:  6,
				End:    11,
			},
		},
		PaneID: "p1",
	}

	resolvedLegacy, err := ResolveIntent(ctx, legacyIntent)
	require.NoError(t, err)
	require.Equal(t, 1, len(resolvedLegacy.Anchors))
	// 确保遗留意图有遗留锚点
	for _, anchor := range resolvedLegacy.Anchors {
		require.Equal(t, AnchorOriginLegacy, anchor.Origin)
	}
}

// TestCursorRefToStateConversion 测试光标引用到状态的转换
func TestCursorRefToStateConversion(t *testing.T) {
	snap := testSnapshot()

	// 测试主光标
	cursorRef := CursorRef{Kind: CursorPrimary}
	cursorState, err := CursorRefToState(cursorRef, snap)
	require.NoError(t, err)
	require.Equal(t, snap.Lines[0].ID, cursorState.LineID)
	require.Equal(t, 0, cursorState.Offset)

	// 测试选择开始光标
	selectionStartRef := CursorRef{Kind: CursorSelectionStart}
	selectionStartState, err := CursorRefToState(selectionStartRef, snap)
	require.NoError(t, err)
	require.Equal(t, snap.Lines[0].ID, selectionStartState.LineID)
	require.Equal(t, 0, selectionStartState.Offset)

	// 测试选择结束光标
	selectionEndRef := CursorRef{Kind: CursorSelectionEnd}
	selectionEndState, err := CursorRefToState(selectionEndRef, snap)
	require.NoError(t, err)
	require.Equal(t, snap.Lines[0].ID, selectionEndState.LineID)
	require.Equal(t, 0, selectionEndState.Offset)
}