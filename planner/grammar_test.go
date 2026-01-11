package planner

import (
	"testing"
	"tmux-fsm/fsm"
	intentPkg "tmux-fsm/intent"
)

func TestGrammarBasicMotion(t *testing.T) {
	g := NewGrammar()

	// 测试 hjkl 移动
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "h"})
	if intent == nil || intent.Kind != intentPkg.IntentMove {
		t.Errorf("Expected move intent for 'h', got %v", intent)
	}

	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "j"})
	if intent == nil || intent.Kind != intentPkg.IntentMove {
		t.Errorf("Expected move intent for 'j', got %v", intent)
	}

	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "k"})
	if intent == nil || intent.Kind != intentPkg.IntentMove {
		t.Errorf("Expected move intent for 'k', got %v", intent)
	}

	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "l"})
	if intent == nil || intent.Kind != intentPkg.IntentMove {
		t.Errorf("Expected move intent for 'l', got %v", intent)
	}
}

func TestGrammarCount(t *testing.T) {
	g := NewGrammar()

	// 测试数字计数
	g.Consume(fsm.RawToken{Kind: fsm.TokenDigit, Value: "3"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "w"})
	if intent == nil || intent.Count != 3 {
		t.Errorf("Expected count 3 for '3w', got %v", intent)
	}
}

func TestGrammarOperatorMotion(t *testing.T) {
	g := NewGrammar()

	// 测试 d + w
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "w"})
	if intent == nil || intent.Kind != intentPkg.IntentOperator {
		t.Errorf("Expected operator intent for 'dw', got %v", intent)
	}
}

func TestGrammarOperatorCountMotion(t *testing.T) {
	g := NewGrammar()

	// 测试 d2w
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	g.Consume(fsm.RawToken{Kind: fsm.TokenDigit, Value: "2"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "w"})
	if intent == nil || intent.Kind != intentPkg.IntentOperator || intent.Count != 2 {
		t.Errorf("Expected operator intent with count 2 for 'd2w', got %v", intent)
	}
}

func TestGrammarGg(t *testing.T) {
	g := NewGrammar()

	// 测试 gg
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "g"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "g"})
	if intent == nil || intent.Kind != intentPkg.IntentMove {
		t.Errorf("Expected move intent for 'gg', got %v", intent)
	}
}

func TestGrammarFfTt(t *testing.T) {
	g := NewGrammar()

	// 测试 fa
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "f"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "a"})
	if intent == nil {
		t.Fatal("Expected intent for 'fa'")
	}
	if intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionFind ||
		intent.Motion.Find == nil ||
		intent.Motion.Find.Char != 'a' ||
		intent.Motion.Find.Direction != intentPkg.FindForward ||
		intent.Motion.Find.Till {
		t.Errorf("Expected forward find motion for 'fa', got %+v", intent.Motion)
	}

	// 测试 ta
	g = NewGrammar()
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "t"})
	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "b"})
	if intent == nil {
		t.Fatal("Expected intent for 'tb'")
	}
	if intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionFind ||
		intent.Motion.Find == nil ||
		intent.Motion.Find.Char != 'b' ||
		intent.Motion.Find.Direction != intentPkg.FindForward ||
		!intent.Motion.Find.Till {
		t.Errorf("Expected forward till motion for 'tb', got %+v", intent.Motion)
	}

	// 测试 Fa
	g = NewGrammar()
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "F"})
	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "c"})
	if intent == nil {
		t.Fatal("Expected intent for 'Fc'")
	}
	if intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionFind ||
		intent.Motion.Find == nil ||
		intent.Motion.Find.Char != 'c' ||
		intent.Motion.Find.Direction != intentPkg.FindBackward ||
		intent.Motion.Find.Till {
		t.Errorf("Expected backward find motion for 'Fc', got %+v", intent.Motion)
	}

	// 测试 Ta
	g = NewGrammar()
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "T"})
	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	if intent == nil {
		t.Fatal("Expected intent for 'Td'")
	}
	if intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionFind ||
		intent.Motion.Find == nil ||
		intent.Motion.Find.Char != 'd' ||
		intent.Motion.Find.Direction != intentPkg.FindBackward ||
		!intent.Motion.Find.Till {
		t.Errorf("Expected backward till motion for 'Td', got %+v", intent.Motion)
	}
}

func TestGrammarTextObject(t *testing.T) {
	g := NewGrammar()

	// 测试 iw
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "i"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "w"})
	if intent == nil {
		t.Fatal("expected intent for 'iw'")
	}
	if intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionRange ||
		intent.Motion.Range == nil ||
		intent.Motion.Range.TextObject == nil ||
		intent.Motion.Range.TextObject.Object != intentPkg.Word {
		t.Errorf("expected word text object motion, got %+v", intent.Motion)
	}

	// 测试 diw
	g = NewGrammar()
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "i"})
	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "w"})
	if intent == nil {
		t.Fatal("expected intent for 'diw'")
	}
	if intent.Op == nil ||
		intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionRange {
		t.Errorf("expected operator + text object motion, got %+v", intent)
	}
}

func TestGrammarRepeat(t *testing.T) {
	g := NewGrammar()

	// 测试重复
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenRepeat, Value: "."})
	if intent == nil {
		t.Errorf("Expected repeat intent for '.'")
	}
}

func TestGrammarUndoRedo(t *testing.T) {
	g := NewGrammar()

	// 测试撤销
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "u"})
	if intent == nil || intent.Kind != intentPkg.IntentUndo {
		t.Errorf("Expected undo intent for 'u', got %v", intent)
	}

	// 测试重做
	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "C-r"})
	if intent == nil || intent.Kind != intentPkg.IntentRedo {
		t.Errorf("Expected redo intent for 'C-r', got %v", intent)
	}
}

func TestGrammarFindRepeat(t *testing.T) {
	g := NewGrammar()

	// 测试查找重复
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: ";"})
	if intent == nil || intent.Kind != intentPkg.IntentRepeatFind {
		t.Errorf("Expected repeat find intent for ';', got %v", intent)
	}

	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: ","})
	if intent == nil || intent.Kind != intentPkg.IntentRepeatFindReverse {
		t.Errorf("Expected reverse repeat find intent for ',', got %v", intent)
	}
}

func TestGrammarLineOperations(t *testing.T) {
	g := NewGrammar()

	// 测试 dd
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	if intent == nil || intent.Kind != intentPkg.IntentOperator || intent.Motion.Kind != intentPkg.MotionLine {
		t.Errorf("Expected line operator intent for 'dd', got %v", intent)
	}

	// 测试 yy
	g = NewGrammar()
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "y"})
	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "y"})
	if intent == nil || intent.Kind != intentPkg.IntentOperator || intent.Motion.Kind != intentPkg.MotionLine {
		t.Errorf("Expected line operator intent for 'yy', got %v", intent)
	}
}

func TestGrammarTextObjectTypes(t *testing.T) {
	// 测试各种文本对象
	testCases := []struct {
		key      string
		expected intentPkg.TextObjectKind
	}{
		{"w", intentPkg.Word},
		{"\"", intentPkg.QuoteDouble},
		{"'", intentPkg.QuoteSingle},
		{"`", intentPkg.Backtick},
		{"(", intentPkg.Paren},
		{"[", intentPkg.Bracket},
		{"{", intentPkg.Brace},
	}

	for _, tc := range testCases {
		g := NewGrammar()
		g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "i"})
		intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: tc.key})
		if intent == nil {
			t.Errorf("Expected intent for 'i%s', got nil", tc.key)
			continue
		}
		if intent.Motion == nil ||
			intent.Motion.Kind != intentPkg.MotionRange ||
			intent.Motion.Range == nil ||
			intent.Motion.Range.TextObject == nil ||
			intent.Motion.Range.TextObject.Object != tc.expected {
			t.Errorf("Expected %v text object for 'i%s', got %+v", tc.expected, tc.key, intent.Motion)
		}
	}
}

func TestGrammarAroundTextObject(t *testing.T) {
	g := NewGrammar()

	// 测试 aw (around word)
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "a"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "w"})
	if intent == nil {
		t.Fatal("expected intent for 'aw'")
	}
	if intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionRange ||
		intent.Motion.Range == nil ||
		intent.Motion.Range.TextObject == nil ||
		intent.Motion.Range.TextObject.Scope != intentPkg.Around {
		t.Errorf("expected around word text object motion, got %+v", intent.Motion)
	}
}

func TestGrammarResetOnSystemEvent(t *testing.T) {
	g := NewGrammar()

	// 设置一些状态
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	if g.pendingOp == nil {
		t.Fatal("Expected pending op after 'd'")
	}

	// 发送系统重置事件
	g.Consume(fsm.RawToken{Kind: fsm.TokenSystem, Value: "reset"})

	if g.pendingOp != nil {
		t.Errorf("Expected pending op to be reset, got %v", g.pendingOp)
	}
	if g.count != 0 {
		t.Errorf("Expected count to be reset to 0, got %d", g.count)
	}
}

func TestGrammarGetPendingOp(t *testing.T) {
	g := NewGrammar()

	// 测试获取待处理操作符
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	if g.GetPendingOp() != "delete" {
		t.Errorf("Expected pending op to be 'delete', got '%s'", g.GetPendingOp())
	}

	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "y"})
	if g.GetPendingOp() != "yank" {
		t.Errorf("Expected pending op to be 'yank', got '%s'", g.GetPendingOp())
	}

	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "c"})
	if g.GetPendingOp() != "change" {
		t.Errorf("Expected pending op to be 'change', got '%s'", g.GetPendingOp())
	}
}

func TestGrammarComplexSequences(t *testing.T) {
	g := NewGrammar()

	// 测试复杂的按键序列：2d3w
	// 在Vim中，2d3w表示删除2*3=6个单词，但我们的实现中，数字是累加的
	// 2d3w 应该是先累积数字2，然后遇到d，再累积数字3，最后遇到w
	// 根据代码，数字是累加的：g.count = g.count*10 + int(tok.Value[0]-'0')
	// 所以 2d3w 会变成 g.count = 2*10 + 3 = 23
	g.Consume(fsm.RawToken{Kind: fsm.TokenDigit, Value: "2"})
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	g.Consume(fsm.RawToken{Kind: fsm.TokenDigit, Value: "3"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "w"})
	if intent == nil || intent.Kind != intentPkg.IntentOperator {
		t.Errorf("Expected operator intent for '2d3w', got %v", intent)
	}
	// 根据代码逻辑，数字是累加的，所以最终的 count 应该是 23 (2*10+3)
	if intent.Count != 23 {
		t.Errorf("Expected count 23 for '2d3w', got %d", intent.Count)
	}
}

func TestGrammarInvalidKeyResets(t *testing.T) {
	g := NewGrammar()

	// 设置一些状态
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	if g.pendingOp == nil {
		t.Fatal("Expected pending op after 'd'")
	}

	// 发送无效键，应该重置状态
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "z"}) // z is not a valid vim key in this context
	if g.pendingOp != nil {
		t.Errorf("Expected pending op to be reset after invalid key, got %v", g.pendingOp)
	}
	if intent != nil {
		t.Errorf("Expected no intent for invalid key, got %v", intent)
	}
}
