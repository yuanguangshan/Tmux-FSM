package logic

import (
	"tmux-fsm/weaver/core"
)

// ShellFactBuilder 纯语义构建器 (Phase 5.3)
// 不再读取 tmux buffer，不再计算 offset
type ShellFactBuilder struct{}

func (b *ShellFactBuilder) Build(intent core.Intent, snapshot core.Snapshot) ([]core.Fact, []core.Fact, error) {
	facts := make([]core.Fact, 0)
	meta := intent.GetMeta()
	target := intent.GetTarget()

	// 基础语义 Anchor
	// Phase 6.2: 从 Snapshot 获取 Expectation (Line Hash)
	row := snapshot.Cursor.Row
	// col := snapshot.Cursor.Col // If needed for semantic logic refinement

	var lineHash string
	// Find line in snapshot
	// Snapshot Lines order matches Rows? Usually yes, row=index.
	// Check bounds
	if row >= 0 && row < len(snapshot.Lines) {
		lineHash = string(snapshot.Lines[row].Hash)
	}

	anchor := core.Anchor{
		PaneID: snapshot.PaneID,
		Kind:   core.AnchorAtCursor, // 默认为光标处
		Hash:   lineHash,
	}

	// 根据 Target 将语义映射到 AnchorKind
	// 假设 TargetKind: 0=Char, 1=Word, 2=Line, 3=Selection
	switch target.Kind {
	case 1: // Word
		anchor.Kind = core.AnchorWord
	case 2: // Line
		anchor.Kind = core.AnchorLine
	case 3: // Selection
		anchor.Kind = core.AnchorSelection
	}

	switch intent.GetKind() {
	case core.IntentInsert:
		text := target.Value
		facts = append(facts, core.Fact{
			Kind:    core.FactInsert,
			Anchor:  anchor,
			Payload: core.FactPayload{Text: text},
			Meta:    meta,
		})

	case core.IntentDelete:
		facts = append(facts, core.Fact{
			Kind:   core.FactDelete,
			Anchor: anchor,
			Meta:   meta,
		})

	case core.IntentReplace:
		// Replace assumes AtCursor (Char) usually?
		// meta["new_text"]
		newText := ""
		if v, ok := meta["new_text"].(string); ok {
			newText = v
		}
		facts = append(facts, core.Fact{
			Kind:    core.FactReplace,
			Anchor:  anchor,
			Payload: core.FactPayload{NewText: newText},
			Meta:    meta,
		})

	case core.IntentMove:
		// Move is FactMove.
		// Target Value might be motion string?
		facts = append(facts, core.Fact{
			Kind:   core.FactMove,
			Anchor: anchor,
			Meta:   meta,
		})
	}

	// Inverse Facts:
	// Phase 5.3: Planner 无法生成反向事实，因为不仅要读取状态，甚至不知道 Resolve 后的位置。
	// Undo 逻辑必须依赖 Resolver 在 Execution 阶段的捕获，或者 History 存储 ResolvedFact。
	// 这里返回空。
	return facts, []core.Fact{}, nil
}
