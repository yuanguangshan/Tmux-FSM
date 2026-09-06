package tests

// M4.5 探测/回归：扩展文本对象（i( i{ i[ i' i` aw）端到端 intent 产出。
// grammar 对象键表（grammar.go:439）声称支持 w ( ) b [ ] { } B " ' `，
// 本测试验证 Promote 后的 Meta 是否携带正确的 text_object 信息。

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"tmux-fsm/fsm"
	"tmux-fsm/kernel"
)

func m4kernel(t *testing.T) (*kernel.Kernel, *MockExecutor) {
	t.Helper()
	keymap := fsm.Keymap{
		Initial: "NAV",
		States: map[string]fsm.StateDef{
			"NAV": {
				Keys: map[string]fsm.KeyAction{
					"d": {Action: ""},
					"y": {Action: ""},
					"c": {Action: ""},
				},
			},
		},
	}
	eng := fsm.NewEngine(&keymap)
	exec := &MockExecutor{}
	return kernel.NewKernel(eng, exec), exec
}

func pressAll(k *kernel.Kernel, keys []string) {
	for _, key := range keys {
		k.HandleKey(kernel.HandleContext{
			Ctx:       context.Background(),
			RequestID: "m4",
			ActorID:   "p1|m4",
		}, key)
	}
}

func TestM45_ExtendedTextObjects(t *testing.T) {
	cases := []struct {
		name string
		keys []string
	}{
		{"ci( 括号内", []string{"c", "i", "("}},
		{"ci{ 花括号内", []string{"c", "i", "{"}},
		{"ci[ 方括号内", []string{"c", "i", "["}},
		{"ci' 单引号内", []string{"c", "i", "'"}},
		{"ci` 反引号内", []string{"c", "i", "`"}},
		{"daw 删周围词", []string{"d", "a", "w"}},
		{"ciw 改内词", []string{"c", "i", "w"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			k, exec := m4kernel(t)
			pressAll(k, tc.keys)
			if exec.CapturedIntent == nil {
				t.Errorf("序列 %v 未产生 Intent", tc.keys)
			}
		})
	}
}

// M4.5 验收：Meta 必须携带 text_object 信息供物理层匹配
func TestM45_TextObjectMeta(t *testing.T) {
	cases := []struct {
		name      string
		keys      []string
		wantParts []string
	}{
		{"ci(", []string{"c", "i", "("}, []string{"paren"}},
		{"ci{", []string{"c", "i", "{"}, []string{"brace"}},
		{"ci[", []string{"c", "i", "["}, []string{"bracket"}},
		{`ci"`, []string{"c", "i", "\""}, []string{"quote"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			k, exec := m4kernel(t)
			pressAll(k, tc.keys)
			if exec.CapturedIntent == nil {
				t.Fatalf("未产生 Intent")
			}
			metaJSON, _ := json.Marshal(exec.CapturedIntent.Meta)
			for _, part := range tc.wantParts {
				if !strings.Contains(string(metaJSON), part) {
					t.Errorf("Meta %s 缺少 %q", metaJSON, part)
				}
			}
		})
	}
}
