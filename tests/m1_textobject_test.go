package tests

// M1.3/M1.5 回归测试：keymap 未声明的按键必须透传给 Grammar——
// 修复前 i/a/引号/查找目标字符被 DecisionLegacy 静默吞掉，
// ci"/ciw/dfa 等 advertised 功能完全无响应。

import (
	"context"
	"testing"

	"tmux-fsm/fsm"
	"tmux-fsm/intent"
	"tmux-fsm/kernel"
)

func newM1Kernel(t *testing.T) (*kernel.Kernel, *MockExecutor) {
	t.Helper()
	// 故意不含 i / w / a / 查找目标字符——它们必须经 M1.3 透传路径
	// 进入 Grammar，而不是被 keymap 过滤掉。
	keymap := fsm.Keymap{
		Initial: "NAV",
		States: map[string]fsm.StateDef{
			"NAV": {
				Keys: map[string]fsm.KeyAction{
					"d": {Action: ""},
					"y": {Action: ""},
					"c": {Action: ""},
					"f": {Action: ""},
				},
			},
		},
	}
	eng := fsm.NewEngine(&keymap)
	exec := &MockExecutor{}
	return kernel.NewKernel(eng, exec), exec
}

func pressKey(k *kernel.Kernel, key string) {
	k.HandleKey(kernel.HandleContext{
		Ctx:       context.Background(),
		RequestID: "m1-test",
		ActorID:   "p1|tester",
	}, key)
}

// ciw：操作符 + 文本对象（inner word）——修复前 "i" 被吞、语义丢失
func TestM13_TextObject_CiW(t *testing.T) {
	k, exec := newM1Kernel(t)
	pressKey(k, "c")
	pressKey(k, "i")
	pressKey(k, "w")

	if exec.CapturedIntent == nil {
		t.Fatal("ciw 应产生 Intent（修复前 i 被吞，永远不产生）")
	}
	if exec.CapturedIntent.Kind != intent.IntentOperator {
		t.Errorf("Kind = %v, want IntentOperator", exec.CapturedIntent.Kind)
	}
	if exec.CapturedIntent.PaneID != "p1" {
		t.Errorf("PaneID = %q, want p1（边界注入）", exec.CapturedIntent.PaneID)
	}
}

// dfa：查找 motion（f + 目标字符）——修复前 "a" 被吞、pendingMotion 永久挂起
func TestM13_FindMotion_DF(t *testing.T) {
	k, exec := newM1Kernel(t)
	pressKey(k, "d")
	pressKey(k, "f")
	pressKey(k, "a")

	if exec.CapturedIntent == nil {
		t.Fatal("dfa 应产生 Intent（修复前目标字符被吞，永久挂起）")
	}
}

// d2fa：带 count 的查找 motion
func TestM13_CountedFindMotion(t *testing.T) {
	k, exec := newM1Kernel(t)
	pressKey(k, "d")
	pressKey(k, "2")
	pressKey(k, "f")
	pressKey(k, "b")

	if exec.CapturedIntent == nil {
		t.Fatal("d2fb 应产生 Intent")
	}
	if exec.CapturedIntent.Count != 2 {
		t.Errorf("Count = %d, want 2", exec.CapturedIntent.Count)
	}
}

// M1.4 回归：词级 motion 必须携带 Direction → promote 层映射 meta["motion"]
func TestM14_WordMotionMeta(t *testing.T) {
	k, exec := newM1Kernel(t)
	pressKey(k, "d")
	pressKey(k, "w")

	if exec.CapturedIntent == nil {
		t.Fatal("dw 应产生 Intent")
	}
	meta := exec.CapturedIntent.Meta
	if got, _ := meta["motion"].(string); got != "word_forward" {
		t.Errorf("dw meta[motion] = %v, want word_forward", got)
	}
}

func TestM14_WordBackwardMeta(t *testing.T) {
	k, exec := newM1Kernel(t)
	pressKey(k, "d")
	pressKey(k, "b")

	if exec.CapturedIntent == nil {
		t.Fatal("db 应产生 Intent")
	}
	meta := exec.CapturedIntent.Meta
	if got, _ := meta["motion"].(string); got != "word_backward" {
		t.Errorf("db meta[motion] = %v, want word_backward", got)
	}
}

// M1.4 回归：G/gg 方向区分——dG 应删到文件尾（此前 gg/G 同码、meta 为空）
func TestM14_GotoMeta(t *testing.T) {
	k, exec := newM1Kernel(t)
	pressKey(k, "d")
	pressKey(k, "G")

	if exec.CapturedIntent == nil {
		t.Fatal("dG 应产生 Intent")
	}
	meta := exec.CapturedIntent.Meta
	if got, _ := meta["motion"].(string); got != "end_of_file" {
		t.Errorf("dG meta[motion] = %v, want end_of_file", got)
	}
}

// M1.5 回归：till 变体（dt→find_char_before）——此前目标字符被吞
func TestM13_TillMotion_DT(t *testing.T) {
	k, exec := newM1Kernel(t)
	pressKey(k, "d")
	pressKey(k, "t")
	pressKey(k, "a")

	if exec.CapturedIntent == nil {
		t.Fatal("dt a 应产生 Intent")
	}
	meta := exec.CapturedIntent.Meta
	if got, _ := meta["motion"].(string); got != "find_char_before" {
		t.Errorf("dt meta[motion] = %v, want find_char_before", got)
	}
}
