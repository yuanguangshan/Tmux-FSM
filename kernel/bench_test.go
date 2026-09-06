package kernel

// M3.2 延迟基准（GOAL M3.2）：量化 HandleKey 各按键类别的内存态延迟。
//
// 注意边界：这里测的是 **daemon 内存态** 的按键处理开销（Decide+Execute），
// 不含 tmux-fsm -key 客户端的进程 fork 与 tmux 选项持久化 I/O——
// 端到端数字见 docs/PERFORMANCE.md 的真机实测。
//
// 运行：go test ./kernel/ -bench BenchmarkHandleKey -benchmem -run '^$'

import (
	"context"
	"testing"

	"tmux-fsm/fsm"
	"tmux-fsm/intent"
)

type benchExecutor struct{}

func (benchExecutor) Process(*intent.Intent) error { return nil }

func (benchExecutor) ProcessWithContext(_ context.Context, _ HandleContext, _ *intent.Intent) error {
	return nil
}

func benchKernel() *Kernel {
	keymap := fsm.Keymap{
		Initial: "NAV",
		States: map[string]fsm.StateDef{
			"NAV": {
				Keys: map[string]fsm.KeyAction{
					"h": {}, "j": {}, "k": {}, "l": {},
					"d": {}, "y": {}, "c": {},
					"f": {}, "w": {}, "b": {}, "e": {},
				},
			},
		},
	}
	return NewKernel(fsm.NewEngine(&keymap), benchExecutor{})
}

func benchHctx() HandleContext {
	return HandleContext{Ctx: context.Background(), RequestID: "bench", ActorID: "p1|bench"}
}

// 委托类按键（h/j/k/l）：Decide → FSM 吃键 → Grammar 产 token → 无意图
func BenchmarkHandleKeyDelegated(b *testing.B) {
	k := benchKernel()
	hctx := benchHctx()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k.HandleKey(hctx, "j")
	}
}

// 操作符起始键（d）：进入 pending 状态
func BenchmarkHandleKeyOperator(b *testing.B) {
	k := benchKernel()
	hctx := benchHctx()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k.HandleKey(hctx, "d")
		k.HandleKey(hctx, "Escape") // 复位，避免 pending 状态累积
	}
}

// 完整操作（dw）：每次都走完 Grammar → Promote → Execute 全链路
func BenchmarkHandleKeyFullOp(b *testing.B) {
	k := benchKernel()
	hctx := benchHctx()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k.HandleKey(hctx, "d")
		k.HandleKey(hctx, "w")
	}
}

// 计数 + 移动（3j 类语义的最热路径）
func BenchmarkHandleKeyCounted(b *testing.B) {
	k := benchKernel()
	hctx := benchHctx()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k.HandleKey(hctx, "3")
		k.HandleKey(hctx, "j")
	}
}

// 文本对象（ciw）：操作符 + 文本对象 + 对象词
func BenchmarkHandleKeyTextObject(b *testing.B) {
	k := benchKernel()
	hctx := benchHctx()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k.HandleKey(hctx, "c")
		k.HandleKey(hctx, "i")
		k.HandleKey(hctx, "w")
	}
}
