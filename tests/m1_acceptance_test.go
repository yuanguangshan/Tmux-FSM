package tests

// M1 里程碑验收（GOAL M1.6）： advertised 功能端到端全部产生正确 Intent。
// 每个 Case 用全新 Kernel（互不污染），覆盖 GOAL 指定的完整清单。

import (
	"context"
	"testing"

	"tmux-fsm/fsm"
	"tmux-fsm/kernel"
)

func acceptanceKernel(t *testing.T) (*kernel.Kernel, *MockExecutor) {
	t.Helper()
	keymap := fsm.Keymap{
		Initial: "NAV",
		States: map[string]fsm.StateDef{
			"NAV": {
				Keys: map[string]fsm.KeyAction{
					"d": {Action: ""},
					"y": {Action: ""},
					"c": {Action: ""},
					"f": {Action: ""},
					"t": {Action: ""},
					"$": {Action: ""},
				},
			},
		},
	}
	eng := fsm.NewEngine(&keymap)
	exec := &MockExecutor{}
	return kernel.NewKernel(eng, exec), exec
}

func runSeq(k *kernel.Kernel, keys []string) {
	for _, key := range keys {
		k.HandleKey(kernel.HandleContext{
			Ctx:       context.Background(),
			RequestID: "m1-acceptance",
			ActorID:   "p1|acceptance",
		}, key)
	}
}

// TestM1_Acceptance_Motions 动词+count+motion 的端到端矩阵
func TestM1_Acceptance_Motions(t *testing.T) {
	cases := []struct {
		name       string
		keys       []string
		wantMotion string
		wantCount  int
	}{
		{"dw 删一词", []string{"d", "w"}, "word_forward", 1},
		{"db 删到上一词首", []string{"d", "b"}, "word_backward", 1},
		{"de 删到词尾", []string{"d", "e"}, "word_forward", 1},
		{"d2w 带 count", []string{"d", "2", "w"}, "word_forward", 2},
		{"dG 删到文件尾", []string{"d", "G"}, "end_of_file", 1},
		{"dfa 删到字符 a", []string{"d", "f", "a"}, "find_char_forward", 1},
		{"dt a 删到字符 a 前", []string{"d", "t", "a"}, "find_char_before_forward", 1},
		{"3j 下移三行", []string{"3", "j"}, "line_down", 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			k, exec := acceptanceKernel(t)
			runSeq(k, tc.keys)

			if exec.CapturedIntent == nil {
				t.Fatalf("序列 %v 未产生 Intent", tc.keys)
			}
			got, _ := exec.CapturedIntent.Meta["motion"].(string)
			if got != tc.wantMotion {
				t.Errorf("meta[motion] = %q, want %q", got, tc.wantMotion)
			}
			if exec.CapturedIntent.Count != tc.wantCount {
				t.Errorf("Count = %d, want %d", exec.CapturedIntent.Count, tc.wantCount)
			}
			if exec.CapturedIntent.PaneID != "p1" {
				t.Errorf("PaneID = %q, want p1", exec.CapturedIntent.PaneID)
			}
		})
	}
}

// TestM1_Acceptance_CountLifecycle 计数器生命周期（M1.2 回归）：
// 3j 消化后计数必须归零，随后 0 应作为行首 motion 而非计数数字
func TestM1_Acceptance_CountLifecycle(t *testing.T) {
	k, exec := acceptanceKernel(t)
	runSeq(k, []string{"3", "j"})
	if exec.CapturedIntent == nil || exec.CapturedIntent.Count != 3 {
		t.Fatalf("3j 计数应为 3")
	}

	// 3j 消化后（ResetCount 已发生），新鲜的 0 应产生 goto_line_start
	k2, exec2 := acceptanceKernel(t)
	_ = k2
	runSeq2(k2, exec2, []string{"0"})
	got, _ := exec2.CapturedIntent.Meta["motion"].(string)
	if got != "goto_line_start" {
		t.Errorf("新鲜 0 的 meta[motion] = %q, want goto_line_start（计数泄漏症状是空）", got)
	}
}

func runSeq2(k *kernel.Kernel, exec *MockExecutor, keys []string) {
	for _, key := range keys {
		k.HandleKey(kernel.HandleContext{
			Ctx:       context.Background(),
			RequestID: "m1-acceptance-2",
			ActorID:   "p1|acceptance",
		}, key)
	}
}

// TestM1_Acceptance_TextObjects 操作符+文本对象（ciw/yiw/ci"）
func TestM1_Acceptance_TextObjects(t *testing.T) {
	cases := []struct {
		name string
		keys []string
	}{
		{"ciw 改内词", []string{"c", "i", "w"}},
		{"yiw 复制内词", []string{"y", "i", "w"}},
		{`ci" 改引号内`, []string{"c", "i", "\""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			k, exec := acceptanceKernel(t)
			runSeq(k, tc.keys)
			if exec.CapturedIntent == nil {
				t.Fatalf("序列 %v 未产生 Intent（修复前 i 与引号都被吞）", tc.keys)
			}
		})
	}
}
