// LEGACY — DO NOT EXTEND
// This path exists ONLY for backward compatibility.
// Any new behavior MUST be implemented via native Intent builders.
package main

import (
	"fmt"
	"strings"
)

// actionStringToIntent 将 legacy action string 转换为 Intent
// 这是阶段 1 的临时桥接函数，用于保持向后兼容
// 最终会被移除，直接从 handleXXX 函数返回 Intent
// actionStringToIntent 将 legacy action string 转换为 Intent
// 这是阶段 1 的临时桥接函数，用于保持向后兼容
// 最终会被移除，直接从 handleXXX 函数返回 Intent
func actionStringToIntent(action string, count int, paneID string) Intent {
	return actionStringToIntentWithLineInfo(action, count, paneID, "", 0, 0)
}

// actionStringToIntentWithLineInfo 将 legacy action string 转换为 Intent，包含行信息
// 这是为了解决 projection conflict check failed: missing LineID 的问题
func actionStringToIntentWithLineInfo(action string, count int, paneID string, lineID string, row int, col int) Intent {
	base := Intent{PaneID: paneID}

	if action == "" {
		base.Kind = IntentNone
		return base
	}

	// 特殊的单一动作
	switch action {
	case "undo":
		return createIntentWithAnchor(Intent{Kind: IntentUndo, Count: count, PaneID: paneID}, paneID, lineID, row, col)
	case "redo":
		return createIntentWithAnchor(Intent{Kind: IntentRedo, Count: count, PaneID: paneID}, paneID, lineID, row, col)
	case "repeat_last":
		return createIntentWithAnchor(Intent{Kind: IntentRepeat, Count: count, PaneID: paneID}, paneID, lineID, row, col)
	case "exit":
		return createIntentWithAnchor(Intent{Kind: IntentExit, PaneID: paneID}, paneID, lineID, row, col)
	case "toggle_case":
		return createIntentWithAnchor(Intent{Kind: IntentToggleCase, Count: count, PaneID: paneID}, paneID, lineID, row, col)
	case "search_next":
		return createIntentWithAnchor(Intent{
			Kind:   IntentSearch,
			Target: SemanticTarget{Kind: TargetSearch, Direction: "next"},
			Count:  count,
			PaneID: paneID,
		}, paneID, lineID, row, col)
	case "search_prev":
		return createIntentWithAnchor(Intent{
			Kind:   IntentSearch,
			Target: SemanticTarget{Kind: TargetSearch, Direction: "prev"},
			Count:  count,
			PaneID: paneID,
		}, paneID, lineID, row, col)
	case "start_visual_char":
		return createIntentWithAnchor(Intent{
			Kind:   IntentVisual,
			Target: SemanticTarget{Scope: "char"},
			PaneID: paneID,
		}, paneID, lineID, row, col)
	case "start_visual_line":
		return createIntentWithAnchor(Intent{
			Kind:   IntentVisual,
			Target: SemanticTarget{Scope: "line"},
			PaneID: paneID,
		}, paneID, lineID, row, col)
	case "cancel_selection":
		return createIntentWithAnchor(Intent{
			Kind:   IntentVisual,
			Target: SemanticTarget{Scope: "cancel"},
			PaneID: paneID,
		}, paneID, lineID, row, col)
	}

	// 处理前缀匹配的动作
	if strings.HasPrefix(action, "search_forward_") {
		query := strings.TrimPrefix(action, "search_forward_")
		return createIntentWithAnchor(Intent{
			Kind:   IntentSearch,
			Target: SemanticTarget{Kind: TargetSearch, Value: query},
			Count:  count,
			PaneID: paneID,
		}, paneID, lineID, row, col)
	}

	if strings.HasPrefix(action, "replace_char_") {
		char := strings.TrimPrefix(action, "replace_char_")
		return createIntentWithAnchor(Intent{
			Kind:   IntentReplace,
			Target: SemanticTarget{Value: char},
			Count:  count,
			PaneID: paneID,
		}, paneID, lineID, row, col)
	}

	if strings.HasPrefix(action, "find_") {
		parts := strings.SplitN(action, "_", 3)
		if len(parts) == 3 {
			return createIntentWithAnchor(Intent{
				Kind:  IntentFind,
				Count: count,
				Meta: map[string]interface{}{
					"find_type": parts[1],
					"char":      parts[2],
				},
				PaneID: paneID,
			}, paneID, lineID, row, col)
		}
	}

	if strings.HasPrefix(action, "visual_") {
		op := strings.TrimPrefix(action, "visual_")
		return createIntentWithAnchor(Intent{
			Kind:   IntentVisual,
			Count:  count,
			Meta:   map[string]interface{}{"operation": op},
			PaneID: paneID,
		}, paneID, lineID, row, col)
	}

	// 解析 operation_motion 格式
	parts := strings.SplitN(action, "_", 2)
	if len(parts) < 2 {
		// 单一动作，无法解析
		base.Kind = IntentNone
		return createIntentWithAnchor(base, paneID, lineID, row, col)
	}

	operation := parts[0]
	motion := parts[1]

	var kind IntentKind
	switch operation {
	case "move":
		kind = IntentMove
	case "delete":
		kind = IntentDelete
	case "change":
		kind = IntentChange
	case "yank":
		kind = IntentYank
	case "insert":
		kind = IntentInsert
	case "paste":
		kind = IntentPaste
	default:
		base.Kind = IntentNone
		return base
	}

	// 解析 motion 为 SemanticTarget
	target := parseMotionToTarget(motion)

	// 将原本的 motion 和 operation 存入 Meta 以供 Weaver Projection 使用
	meta := make(map[string]interface{})
	meta["motion"] = motion
	meta["operation"] = operation

	// 注入 LineID 信息以解决 projection conflict check failed: missing LineID 的问题
	// 如果没有提供 lineID，基于 paneID 和光标位置生成一个
	finalLineID := lineID
	if finalLineID == "" && paneID != "" {
		// Generate a line ID based on pane ID and cursor position
		finalLineID = fmt.Sprintf("%s_line_%d", paneID, row)
	}

	if finalLineID != "" {
		meta["line_id"] = finalLineID
		meta["row"] = row
		meta["col"] = col
	}

	// 创建锚点信息，使其与新架构兼容
	// 这样 ShellFactBuilder 就不需要从快照创建锚点
	anchor := Anchor{
		PaneID: paneID,
		LineID: finalLineID,
		Start:  col,
		End:    col,
		Kind:   int(TargetPosition), // 默认使用位置类型锚点
	}

	// 根据目标类型调整锚点类型和范围，使其与 ShellFactBuilder 期望的值匹配
	switch target.Kind {
	case TargetLine:
		anchor.Kind = int(TargetLine) // = 3
		// For line operations, we might want to set the range to cover the whole line
		// But for now, we'll keep the cursor position and let resolver handle the semantic expansion
	case TargetWord:
		anchor.Kind = int(TargetWord) // = 2
	case TargetChar:
		anchor.Kind = int(TargetChar) // = 1
	case TargetTextObject:
		anchor.Kind = int(TargetTextObject) // = 5
	}

	return Intent{
		Kind:    kind,
		Target:  target,
		Count:   count,
		PaneID:  paneID,
		Meta:    meta,
		Anchors: []Anchor{anchor}, // 添加锚点信息
	}
}

// createIntentWithAnchor creates an intent with proper anchor information
func createIntentWithAnchor(base Intent, paneID string, lineID string, row int, col int) Intent {
	// Generate LineID if not provided
	finalLineID := lineID
	if finalLineID == "" && paneID != "" {
		finalLineID = fmt.Sprintf("%s_line_%d", paneID, row)
	}

	// Create anchor with LineID
	anchor := Anchor{
		PaneID: paneID,
		LineID: finalLineID,
		Start:  col,
		End:    col,
		Kind:   int(TargetPosition), // Default position anchor
	}

	// Add LineID to meta if available
	if finalLineID != "" && base.Meta == nil {
		base.Meta = make(map[string]interface{})
		base.Meta["line_id"] = finalLineID
		base.Meta["row"] = row
		base.Meta["col"] = col
	} else if finalLineID != "" && base.Meta != nil {
		base.Meta["line_id"] = finalLineID
		base.Meta["row"] = row
		base.Meta["col"] = col
	}

	base.Anchors = []Anchor{anchor}
	return base
}

// parseMotionToTarget 将 motion string 解析为 SemanticTarget
func parseMotionToTarget(motion string) SemanticTarget {
	// 方向性移动
	switch motion {
	case "left":
		return SemanticTarget{Kind: TargetChar, Direction: "left"}
	case "right":
		return SemanticTarget{Kind: TargetChar, Direction: "right"}
	case "up":
		return SemanticTarget{Kind: TargetPosition, Direction: "up"}
	case "down":
		return SemanticTarget{Kind: TargetPosition, Direction: "down"}
	}

	// 词级移动
	switch motion {
	case "word_forward":
		return SemanticTarget{Kind: TargetWord, Direction: "forward"}
	case "word_backward":
		return SemanticTarget{Kind: TargetWord, Direction: "backward"}
	case "end_of_word":
		return SemanticTarget{Kind: TargetWord, Scope: "end"}
	}

	// 行级移动
	switch motion {
	case "start_of_line":
		return SemanticTarget{Kind: TargetLine, Scope: "start"}
	case "end_of_line":
		return SemanticTarget{Kind: TargetLine, Scope: "end"}
	case "line":
		return SemanticTarget{Kind: TargetLine, Scope: "whole"}
	}

	// 文件级移动
	switch motion {
	case "start_of_file":
		return SemanticTarget{Kind: TargetFile, Scope: "start"}
	case "end_of_file":
		return SemanticTarget{Kind: TargetFile, Scope: "end"}
	}

	// Insert 的特殊位置
	switch motion {
	case "before":
		return SemanticTarget{Scope: "before"}
	case "after":
		return SemanticTarget{Scope: "after"}
	case "start_of_line":
		return SemanticTarget{Scope: "start_of_line"}
	case "end_of_line":
		return SemanticTarget{Scope: "end_of_line"}
	case "open_below":
		return SemanticTarget{Scope: "open_below"}
	case "open_above":
		return SemanticTarget{Scope: "open_above"}
	}

	// 文本对象
	if strings.HasPrefix(motion, "inside_") || strings.HasPrefix(motion, "around_") {
		return SemanticTarget{Kind: TargetTextObject, Value: motion}
	}

	// 检查是否是文本对象简写 (iw, aw, ip, ap, etc.)
	if isTextObject(motion) {
		return SemanticTarget{Kind: TargetTextObject, Value: motion}
	}

	// 默认返回
	return SemanticTarget{Kind: TargetNone}
}

// isTextObject 检查是否是文本对象简写
func isTextObject(motion string) bool {
	if len(motion) != 2 {
		return false
	}

	// 检查第一个字符是否是 i 或 a (inside/around)
	modifier := motion[0:1]
	if modifier != "i" && modifier != "a" {
		return false
	}

	// 检查第二个字符是否是支持的文本对象类型
	objType := motion[1:2]
	switch objType {
	case "w", "p", "s", "b", "B", "(", ")", "[", "]", "{", "}", "\"", "'", "`":
		return true
	default:
		return false
	}
}
