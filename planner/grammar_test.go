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
