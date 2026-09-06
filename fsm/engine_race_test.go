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
