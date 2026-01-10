package logic

import (
	"tmux-fsm/weaver/core"
)

// ShellFactBuilder 纯语义构建器 (Phase 5.3)
// 不再读取 tmux buffer，不再计算 offset
type ShellFactBuilder struct{}

func (b *ShellFactBuilder) Build(intent core.Intent, snapshot core.Snapshot) ([]core.Fact, []core.Fact, error) {
	meta := intent.GetMeta()
	target := intent.GetTarget()

	// Check if intent has multiple anchors (Phase 11.0)
	anchors := intent.GetAnchors()
	if len(anchors) == 0 {
		// Fallback to original behavior: create anchor from snapshot
		// 基础语义 Anchor
		// Phase 6.2: 从 Snapshot 获取 Expectation (Line Hash)
		row := snapshot.Cursor.Row
		// col := snapshot.Cursor.Col // If needed for semantic logic refinement

		var lineHash string
		var lineID core.LineID
		// Find line in snapshot
		// Snapshot Lines order matches Rows? Usually yes, row=index.
		// Check bounds
		if row >= 0 && row < len(snapshot.Lines) {
			lineHash = string(snapshot.Lines[row].Hash)
			lineID = snapshot.Lines[row].ID
		}

		anchor := core.Anchor{
			PaneID: snapshot.PaneID,
			Kind:   core.AnchorAtCursor, // 默认为光标处
			Hash:   lineHash,
			LineID: lineID, // Phase 9: Include stable LineID
		}

		// 假设 TargetKind: 1=Char, 2=Word, 3=Line, 5=TextObject (from intent.go)
		switch target.Kind {
		case 1: // Char
			anchor.Kind = core.AnchorAtCursor
		case 2: // Word
			anchor.Kind = core.AnchorWord
		case 3: // Line
			anchor.Kind = core.AnchorLine
		case 6: // TextObject
			anchor.Kind = core.AnchorTextObject
			// We need to attach the text object spec to the anchor.
			// Anchor has 'Ref'. usage: Ref = "iw"
			anchor.Ref = target.Value
		}

		anchors = []core.Anchor{anchor}
	}

	// Build facts for each anchor
	facts := make([]core.Fact, 0)
	for _, anchor := range anchors {
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
			// Phase 5.5: Support Text Object Delete in shell builder
			// If target is Text Object, we must generate a FactDelete with AnchorTextObject
			if target.Kind == 6 { // TextObject (TargetTextObject=6)
				// Extract "iw", "ap" etc from value
				// The semantic target value for TextObject is the spec string (e.g. "iw")
				meta["text_object"] = target.Value
				facts = append(facts, core.Fact{
					Kind:   core.FactDelete,
					Anchor: anchor, // This anchor needs to be Kind=AnchorTextObject
					Meta:   meta,
				})
			} else {
				// Handle other delete types (Character, Word, Line, etc.)
				facts = append(facts, core.Fact{
					Kind:   core.FactDelete,
					Anchor: anchor,
					Meta:   meta,
				})
			}

		case core.IntentMove:
			// Move is FactMove.
			// Bridge semantic Motion to legacy meta for TmuxProjection
			// We need to convert the strong-typed Motion from the intent to legacy meta
			// First, we need to check if this is a core.Intent that has access to the original intent.Intent
			// Since we can't directly access the original intent.Intent, we'll need to work with what's available
			// The meta map might contain the motion information if it was populated during promotion
			// If not, we need to create a bridge to extract motion from the semantic intent
			// For now, we'll add a helper to populate motion from semantic intent if not present in meta
			updatedMeta := populateMotionMeta(meta, intent)

			facts = append(facts, core.Fact{
				Kind:   core.FactMove,
				Anchor: anchor,
				Meta:   updatedMeta,
			})

		case core.IntentOperator:
			// Phase 17+ Architecture: High Level Operators (dd, dw, cw, yy)
			updatedMeta := populateMotionMeta(meta, intent)
			opPtr := intent.GetOperator()
			if opPtr != nil {
				op := *opPtr
				// Corresponding Op kinds in intent/intent.go:
				// OpMove = 0, OpDelete = 1, OpYank = 2, OpChange = 3
				if op == 1 { // OpDelete
					facts = append(facts, core.Fact{
						Kind:   core.FactDelete,
						Anchor: anchor,
						Meta:   updatedMeta,
					})
				} else if op == 3 { // OpChange
					// Change is delete + insert mode side effect
					updatedMeta["operation"] = "change"
					facts = append(facts, core.Fact{
						Kind:   core.FactInsert, // Projection knows to enter insert mode
						Anchor: anchor,
						Meta:   updatedMeta,
					})
				}
			}

		case core.IntentEnterVisual, core.IntentVisual:
			// Enter visual mode side effect
			facts = append(facts, core.Fact{
				Kind:   core.FactNone,
				Anchor: anchor,
				Meta: map[string]interface{}{
					"operation": "visual_enter",
				},
			})

		case core.IntentExitVisual:
			// Exit visual mode side effect
			facts = append(facts, core.Fact{
				Kind:   core.FactNone,
				Anchor: anchor,
				Meta: map[string]interface{}{
					"operation": "exit",
				},
			})
		}
	}

	// Inverse Facts:
	// Phase 5.3: Planner 无法生成反向事实，因为不仅要读取状态，甚至不知道 Resolve 后的位置。
	// Undo 逻辑必须依赖 Resolver 在 Execution 阶段的捕获，或者 History 存储 ResolvedFact。
	// 这里返回空。
	return facts, []core.Fact{}, nil
}

// populateMotionMeta 将语义化的运动信息转换为遗留的 Meta 字段
// 这是桥接新架构和现有实现的必要步骤
func populateMotionMeta(meta map[string]interface{}, intent core.Intent) map[string]interface{} {
	// 如果 meta 为 nil，创建一个新的 map
	if meta == nil {
		meta = make(map[string]interface{})
	}

	// 检查 meta 中是否已存在 motion 信息
	if _, exists := meta["motion"]; !exists {
		// 对于 Move 类型的 Intent，如果 Meta 中没有 motion 信息，
		// 我们已经通过 intent.Promote 在 intent.Meta 中填充了相关信息
		// 所以这里不需要额外处理，只需返回现有的 meta
		// 但如果需要进一步处理，可以在这里添加逻辑
	}

	return meta
}
