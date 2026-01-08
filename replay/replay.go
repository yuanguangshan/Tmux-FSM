package replay

import (
	"tmux-fsm/crdt"
	"tmux-fsm/semantic"
	"tmux-fsm/decide"
)

// TextState 文本状态
type TextState struct {
	Text   string
	Cursor int
}

// Clone 克隆状态
func (s TextState) Clone() TextState {
	return TextState{
		Text:   s.Text,
		Cursor: s.Cursor,
	}
}

// TextNode 表示文本中的一个节点
type TextNode struct {
	Pos  interface{} // 这里应该使用 PositionID，但由于循环依赖问题，暂时使用 interface{}
	Rune rune
}

// ApplyFact 应用语义事实
func ApplyFact(state *TextState, fact semantic.BaseFact) {
	// 这里需要根据实际的 Fact 类型进行处理
	// 由于 BaseFact 的字段是私有的，我们需要通过方法访问
	switch fact.Kind() {
	case "insert":
		anchor := fact.GetAnchor()
		text := fact.GetText()
		if anchor.Col >= 0 && anchor.Col <= len(state.Text) {
			state.Text = state.Text[:anchor.Col] + text + state.Text[anchor.Col:]
			state.Cursor = anchor.Col + len(text)
		}
	case "delete":
		rng := fact.GetRange()
		if rng.Start.Col >= 0 && rng.End.Col <= len(state.Text) && rng.Start.Col < rng.End.Col {
			state.Text = state.Text[:rng.Start.Col] + state.Text[rng.End.Col:]
			state.Cursor = rng.Start.Col
		}
	case "move":
		// 更新光标位置
		anchor := fact.GetAnchor()
		state.Cursor = anchor.Col
	}
}

// Replay 重放事件
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

// UndoCheckout 撤销检出
func UndoCheckout(
	target crdt.EventID,
	global map[crdt.EventID]crdt.SemanticEvent,
	me crdt.ActorID,
	initial TextState,
) TextState {
	// 1. 全局 CRDT 决议
	sorted := crdt.TopoSortByCausality(global)

	// 2. 创建撤销过滤器
	filter := crdt.UndoFilter(me, target, global)

	// 3. 重放
	return Replay(initial, sorted, filter)
}