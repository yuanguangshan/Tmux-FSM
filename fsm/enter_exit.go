package fsm

import "tmux-fsm/backend"

// FSMActive indicates whether FSM is in active state (ABI-level invariant)
var FSMActive bool

// EnterFSM enters FSM mode as an atomic operation
// ABI: State + Input Routing + UI
func EnterFSM() {
	// 1️⃣ Logic state
	if defaultEngine == nil {
		InitEngine(&KM)
	}

	FSMActive = true
	SetMode("NAV")

	engine := defaultEngine
	engine.Active = "NAV"
	engine.Reset()
	engine.emitInternal(RawToken{Kind: TokenSystem, Value: "enter"})

	// 2️⃣ Input routing (critical)
	// Switch current client to fsm key table
	backend.GlobalBackend.SwitchClientTable("", "fsm")
	// Fix: Sync authoritative state to tmux so external scripts/hooks respect it
	backend.GlobalBackend.SetUserOption(" @fsm_active", "1")

	// 3️⃣ UI (status bar refresh)
	UpdateUI()
}

// ExitFSM exits FSM mode as an atomic operation
func ExitFSM() {
	// 1️⃣ Logic state
	if defaultEngine != nil {
		FSMActive = false
		// Don't call SetMode("NORMAL") - let UpdateUI() read actual engine state
		// SetMode() only affects currentMode, not defaultEngine.Active

		defaultEngine.Reset()
		defaultEngine.emitInternal(RawToken{Kind: TokenSystem, Value: "exit"})
	}

	// 2️⃣ Input routing restore
	backend.GlobalBackend.SwitchClientTable("", "root")
	// Fix: Clear authoritative state in tmux
	backend.GlobalBackend.SetUserOption(" @fsm_active", "0")

	// 3️⃣ UI
	// Note: Don't call UpdateUI() as it conflicts with HideUI()
	HideUI()
}
