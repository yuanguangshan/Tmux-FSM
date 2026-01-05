package logic

import (
	"os/exec"
	"strings"
	"time"
	"tmux-fsm/weaver/adapter"
	"tmux-fsm/weaver/core"
)

// ShellFactBuilder 生成 Fact 的逻辑组件
type ShellFactBuilder struct{}

func (b *ShellFactBuilder) Build(intent core.Intent, paneID string) ([]core.Fact, []core.Fact, error) {
	kind := intent.GetKind()
	target := intent.GetTarget()
	count := intent.GetCount()
	if count <= 0 {
		count = 1
	}

	var facts []core.Fact
	var inverses []core.Fact

	// 1. 转换 SemanticTarget 为 legacy motion string
	// 这是为了 Phase 3 Smart Projection 能工作
	motion := targetToMotion(target)

	// 根据 Intent 类型生成 Fact
	switch kind {
	case core.IntentKind(2): // Delete
		// 模拟 execute.go 的 logic:
		// 先 capture，再 delete
		if motion == "word_forward" { // 只有这个 motion 支持 capture
			startPos := adapter.TmuxGetCursorPos(paneID)
			content := captureText(motion, paneID) // 使用 adapter 里的 helper

			if content != "" {
				// 生成精确的 Undo Fact
				row, col := adapter.TmuxCurrentCursor(paneID)
				line := adapter.TmuxCaptureLine(paneID, row)

				anchor := core.Anchor{
					ResourceID: paneID,
					Hint:       core.AnchorHint{Line: row, Column: col},
					Hash:       []byte(adapter.TmuxHashLine(line)),
					Offset:     col,
				}

				r := core.Range{
					StartOffset: startPos[0],
					EndOffset:   startPos[0] + len(content),
				}

				// 主 Fact (Carry motion for execution, Carry text/range for Legacy ActionRecord)
				f := core.Fact{
					Kind:    core.FactDelete,
					Anchor:  anchor,
					Range:   r,
					Payload: core.FactPayload{OldText: content},
					Meta: map[string]interface{}{
						"motion": motion,
						"count":  count,
					},
				}
				facts = append(facts, f)

				// Inverse Fact (Insert)
				inv := core.Fact{
					Kind:    core.FactInsert,
					Anchor:  anchor,
					Range:   r,
					Payload: core.FactPayload{Text: content},
				}
				inverses = append(inverses, inv)
			}
		}

		if len(facts) == 0 {
			// 如果没抓到文本（或者不支持抓取），只生成执行 Fact
			// Projection 会盲执行
			f := core.Fact{
				Kind:   core.FactDelete,
				Anchor: core.Anchor{ResourceID: paneID},
				Meta: map[string]interface{}{
					"motion": motion,
					"count":  count,
				},
			}
			facts = append(facts, f)
			// No inverse for dumb delete
		}

	case core.IntentKind(3): // Change
		// Change = Delete + Insert Mode
		// 这里简化处理，类似 Delete，但在 Meta 中标记 change
		// 注意：Legacy Logic 中，Change word 也支持 capture
		captured := false
		if motion == "word_forward" {
			startPos := adapter.TmuxGetCursorPos(paneID)
			content := captureText(motion, paneID)
			if content != "" {
				// Undo Fact (Replace)
				// Legacy changes are recorded as "replace" facts for Undo
				row, col := adapter.TmuxCurrentCursor(paneID)
				line := adapter.TmuxCaptureLine(paneID, row)
				anchor := core.Anchor{
					ResourceID: paneID,
					Hint:       core.AnchorHint{Line: row, Column: col},
					Hash:       []byte(adapter.TmuxHashLine(line)),
					Offset:     col,
				}
				r := core.Range{
					StartOffset: startPos[0],
					EndOffset:   startPos[0] + len(content),
				}

				f := core.Fact{
					Kind:    core.FactReplace, // Legacy uses replace for change undo
					Anchor:  anchor,
					Range:   r,
					Payload: core.FactPayload{OldText: content},
					Meta: map[string]interface{}{
						"motion":    motion,
						"count":     count,
						"operation": "change", // Trigger exitFSM in Projection
					},
				}
				facts = append(facts, f)

				// Inverse: Replace back to old text
				inv := core.Fact{
					Kind:    core.FactReplace,
					Anchor:  anchor,
					Range:   r,
					Payload: core.FactPayload{NewText: content},
					Meta:    map[string]interface{}{"new_text": content},
				}
				inverses = append(inverses, inv)
				captured = true
			}
		}

		if !captured {
			f := core.Fact{
				Kind:   core.FactInsert, // Change ends up in Insert mode
				Anchor: core.Anchor{ResourceID: paneID},
				Meta: map[string]interface{}{
					"motion":    motion,
					"count":     count,
					"operation": "change",
				},
			}
			facts = append(facts, f)
		}

	case core.IntentKind(1): // Move
		f := core.Fact{
			Kind:   core.FactMove,
			Anchor: core.Anchor{ResourceID: paneID},
			Meta: map[string]interface{}{
				"motion": motion,
				"count":  count,
			},
		}
		facts = append(facts, f)

	case core.IntentKind(5): // Insert
		f := core.Fact{
			Kind:   core.FactInsert,
			Anchor: core.Anchor{ResourceID: paneID},
			Meta: map[string]interface{}{
				"motion": motion,
				"count":  count,
			},
		}
		facts = append(facts, f)

	case core.IntentKind(6): // Paste
		f := core.Fact{
			Kind:   core.FactInsert,
			Anchor: core.Anchor{ResourceID: paneID},
			Meta: map[string]interface{}{
				"motion":     "paste", // Hack
				"sub_motion": motion,
				"count":      count,
			},
			Payload: core.FactPayload{Text: "paste_placeholder"}, // 标记这是 paste
		}
		facts = append(facts, f)

	case core.IntentKind(14): // Find
		// Find char
		// Need to extract char from intent?
		// Intent struct doesn't have easy access to Meta here since interface
		// We assume Intent is the struct from main... but here it is core.Intent
		// We need to extend core.Intent interface or use type assertion
		// Let's assume motion carries the info or we rely on Meta passing if we implement it
		// For now, let's assume we can get it.
		// Actually, in Phase 1, we put find_type and char in Meta.
		// We need core.Intent to access Meta.
		// Let's update core.Intent interface to include GetMeta()
	case core.IntentKind(15): // Exit
		f := core.Fact{
			Kind:   core.FactNone,
			Anchor: core.Anchor{ResourceID: paneID},
			Meta:   map[string]interface{}{"operation": "exit"},
		}
		facts = append(facts, f)

	}

	return facts, inverses, nil
}

// targetToMotion 反向转换
func targetToMotion(t core.SemanticTarget) string {
	// 简单的映射，覆盖大多数 case
	if t.Scope == "word" && t.Direction == "forward" {
		return "word_forward"
	}
	if t.Scope == "word" && t.Direction == "backward" {
		return "word_backward"
	}
	if t.Scope == "end" && t.Kind == 2 {
		return "end_of_word"
	} // Kind 2 = Word
	if t.Kind == 1 && t.Direction == "left" {
		return "left"
	}
	if t.Kind == 1 && t.Direction == "right" {
		return "right"
	}
	if t.Kind == 7 && t.Direction == "up" {
		return "up"
	}
	if t.Kind == 7 && t.Direction == "down" {
		return "down"
	}
	if t.Scope == "start" && t.Kind == 3 {
		return "start_of_line"
	}
	if t.Scope == "end" && t.Kind == 3 {
		return "end_of_line"
	}
	if t.Scope == "start" && t.Kind == 4 {
		return "start_of_file"
	}
	if t.Scope == "end" && t.Kind == 4 {
		return "end_of_file"
	}

	// Insert scopes
	if t.Scope == "after" {
		return "after"
	}
	if t.Scope == "start_of_line" {
		return "start_of_line"
	}
	if t.Scope == "end_of_line" {
		return "end_of_line"
	}
	if t.Scope == "open_below" {
		return "open_below"
	}
	if t.Scope == "open_above" {
		return "open_above"
	}

	return "" // Default
}

// captureText Helper (with execute.go logic)
func captureText(motion, paneID string) string {
	if motion == "word_forward" {
		exec.Command("tmux", "send-keys", "-t", paneID, "-X", "begin-selection").Run()
		exec.Command("tmux", "send-keys", "-t", paneID, "-X", "next-word-end").Run()
		exec.Command("tmux", "send-keys", "-t", paneID, "-X", "copy-pipe", "tmux save-buffer -").Run()
		time.Sleep(5 * time.Millisecond)
		out, _ := exec.Command("tmux", "show-buffer").Output()
		return strings.TrimSpace(string(out))
	}
	return ""
}
