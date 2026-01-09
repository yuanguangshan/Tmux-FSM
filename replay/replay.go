package replay

import (
	"tmux-fsm/crdt"
	"tmux-fsm/semantic"
)

//
// ─────────────────────────────────────────────────────────────
//  Text State
// ─────────────────────────────────────────────────────────────
//

// TextState 表示文本 + 光标的最小可重放状态
type TextState struct {
	Text   string
	Cursor int
}

// Clone 返回值拷贝（Replay 入口使用）
func (s TextState) Clone() TextState {
	return TextState{
		Text:   s.Text,
		Cursor: s.Cursor,
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Internal Helpers
// ─────────────────────────────────────────────────────────────
//

func clampCursor(pos, textLen int) int {
	if pos < 0 {
		return 0
	}
	if pos > textLen {
		return textLen
	}
	return pos
}

//
// ─────────────────────────────────────────────────────────────
//  Apply Semantic Fact
// ─────────────────────────────────────────────────────────────
//

// ApplyFact 将一个 semantic.BaseFact 应用到文本状态
//
// ⚠️ replay 层不负责“合法性判断”
//    默认 Fact 已通过 policy / semantic 校验
func ApplyFact(state *TextState, fact semantic.BaseFact) {
	switch fact.Kind() {

	case "insert":
		anchor := fact.GetAnchor()
		text := fact.GetText()

		col := clampCursor(anchor.Col, len(state.Text))
		if text == "" {
			return
		}

		state.Text =
			state.Text[:col] +
				text +
				state.Text[col:]

		state.Cursor = col + len(text)

	case "delete":
		rng := fact.GetRange()

		start := clampCursor(rng.Start.Col, len(state.Text))
		end := clampCursor(rng.End.Col, len(state.Text))

		if start >= end {
			return
		}

		state.Text =
			state.Text[:start] +
				state.Text[end:]

		state.Cursor = start

	case "move":
		anchor := fact.GetAnchor()
		state.Cursor = clampCursor(anchor.Col, len(state.Text))

	default:
		// 未识别的 Fact —— replay 层选择忽略
		return
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Replay
// ─────────────────────────────────────────────────────────────
//

// Replay 对一组语义事件进行重放
//
// filter == nil 表示不过滤
func Replay(
	initial TextState,
	events []crdt.SemanticEvent,
	filter func(crdt.SemanticEvent) bool,
) TextState {

	state := initial.Clone()

	for _, e := range events {
		if filter != nil && !filter(e) {
			continue
		}
		ApplyFact(&state, e.Fact)
	}

	return state
}

//
// ─────────────────────────────────────────────────────────────
//  Undo / Checkout
// ─────────────────────────────────────────────────────────────
//

// UndoCheckout 从全局事件集中，
// 以 target 为撤销目标，
// 重放“我视角下”的文本状态
func UndoCheckout(
	target crdt.EventID,
	global map[crdt.EventID]crdt.SemanticEvent,
	me crdt.ActorID,
	initial TextState,
) TextState {

	// 1️⃣ 全局因果排序（CRDT 层负责）
	ordered := crdt.TopoSortByCausality(global)

	// 2️⃣ 构造撤销过滤器（CRDT 层负责）
	filter := crdt.UndoFilter(me, target, global)

	// 3️⃣ 纯 replay
	return Replay(initial, ordered, filter)
}
