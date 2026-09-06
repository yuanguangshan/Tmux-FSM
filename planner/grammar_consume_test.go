package planner

// WillConsume 门控测试（M4-core 可用性修复）：
// 语法键消费 / 未知可打印键透传 / 序列进行中全消费。

import (
	"testing"

	"tmux-fsm/intent"
)

func TestWillConsumeGrammarKeys(t *testing.T) {
	g := &Grammar{}
	for _, k := range []string{
		"h", "j", "k", "l", "w", "b", "e",
		"0", "$", "^", "G", "gg",
		"d", "y", "c", "u",
		"f", "F", "t", "T",
		"3", "7",
	} {
		if !g.WillConsume(k) {
			t.Errorf("语法键 %q 应被消费", k)
		}
	}
}

func TestWillConsumeRejectsUnknown(t *testing.T) {
	g := &Grammar{}
	// 未知可打印键：不消费 → kernel 层会字面透传给 pane（用户可自由输入）
	for _, k := range []string{"i", "a", "x", "z", "\"", "'", "(", ")", "{", "}", "s", "m"} {
		if g.WillConsume(k) {
			t.Errorf("未知键 %q 不应被语法消费", k)
		}
	}
}

func TestWillConsumeSequencePending(t *testing.T) {
	g := &Grammar{}
	op := intent.OpDelete
	g.pendingOp = &op

	// 操作符挂起：等待 motion/文本对象——任何后续键都属于序列
	if !g.WillConsume("x") {
		t.Error("pendingOp 挂起时应消费任意后续键")
	}
}
