package fsm

// M2.1 并发回归：多 goroutine 并发 Dispatch/Reset 必须通过 -race 检测。
// 修复前 Engine 无锁——并发按键（每连接 goroutine）与 layer 超时回调
// 会竞态改写 Active/count。

import (
	"sync"
	"testing"
)

func TestEngineConcurrentDispatchRace(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{
					"h": {}, "j": {}, "d": {}, "3": {}, "0": {},
				},
			},
		},
	}
	e := NewEngine(&km)

	var wg sync.WaitGroup
	keys := []string{"h", "j", "d", "3", "0", "x"}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for round := 0; round < 50; round++ {
				_, _ = e.Dispatch(keys[(n+round)%len(keys)])
				if round%10 == 0 {
					e.Reset()
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestEngineResetCountConcurrent(t *testing.T) {
	km := Keymap{Initial: "NAV", States: map[string]StateDef{"NAV": {Keys: map[string]KeyAction{}}}}
	e := NewEngine(&km)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			e.Dispatch("3")
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			e.ResetCount()
		}
	}()
	wg.Wait()
}

func TestEngineDotRepeat(t *testing.T) {
	km := Keymap{
		Initial: "NAV",
		States: map[string]StateDef{
			"NAV": {
				Keys: map[string]KeyAction{
					"d": {}, "w": {}, "x": {},
				},
			},
		},
	}
	_ = km
	e := NewEngine(&km)

	// 完成 dw 序列
	e.Dispatch("d")
	e.Dispatch("w")
	// 生产环境中 kernel 在意图派发后调用 ResetCount 固化序列
	e.ResetCount()
	if len(e.lastOpKeys) != 2 {
		t.Fatalf("lastOpKeys = %v, want [d w]", e.lastOpKeys)
	}

	// "." 重放：重放期间 lastOpKeys 保持稳定（不会自我吞并）
	e.Dispatch(".")
	if len(e.lastOpKeys) != 2 {
		t.Errorf("重放后 lastOpKeys = %v, want 仍为 2 项", e.lastOpKeys)
	}

	// 无历史时 "." 为无害 no-op
	e2 := NewEngine(&km)
	e2.Dispatch(".")
	if len(e2.lastOpKeys) != 0 {
		t.Error("无历史时 lastOpKeys 应为空")
	}
}

func TestEngineAbortClearsPending(t *testing.T) {
	km := Keymap{Initial: "NAV", States: map[string]StateDef{"NAV": {Keys: map[string]KeyAction{}}}}
	e := NewEngine(&km)
	e.Dispatch("d") // pending
	e.Reset()       // 用户 Escape 之类的复位
	if len(e.curOpKeys) != 0 {
		t.Error("Reset 应清空中断的序列")
	}
}
