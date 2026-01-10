package main

import (
	"errors"
	"strings"
)

// ResolveContext 包含 Resolver 所需的上下文信息
type ResolveContext struct {
	Snapshot Snapshot
	Cursor   CursorState
}

// ResolvedIntent 表示解析后的意图
type ResolvedIntent struct {
	Intent
	Anchors []ResolvedAnchor
}

// ResolvedAnchor 表示解析后的锚点
type ResolvedAnchor struct {
	PaneID string
	LineID string
	Range  TextRange
	Origin AnchorOrigin
}

// TextRange 表示文本范围
type TextRange struct {
	Start int
	End   int
}

// AnchorOrigin 表示锚点来源
type AnchorOrigin int

const (
	AnchorOriginNative AnchorOrigin = iota
	AnchorOriginLegacy
)

// ResolveIntent 解析意图
func ResolveIntent(ctx ResolveContext, intent Intent) (ResolvedIntent, error) {
	// 特殊处理 Undo 和 Redo 意图
	switch intent.Kind {
	case IntentUndo:
		return resolveUndoIntent(ctx, intent)
	case IntentRedo:
		return resolveRedoIntent(ctx, intent)
	}

	// 创建基础解析后的意图
	resolved := ResolvedIntent{
		Intent:  intent,
		Anchors: []ResolvedAnchor{},
	}

	// 解析锚点
	for _, anchor := range intent.Anchors {
		if isLegacyAnchor(anchor) {
			// 解析遗留锚点
			resolvedAnchor, err := resolveLegacyAnchor(ctx, anchor)
			if err != nil {
				return ResolvedIntent{}, err
			}
			resolvedAnchor.Origin = AnchorOriginLegacy
			resolved.Anchors = append(resolved.Anchors, resolvedAnchor)
		} else {
			// 解析原生锚点
			resolvedAnchor, err := resolveNativeAnchor(ctx, anchor)
			if err != nil {
				return ResolvedIntent{}, err
			}
			resolvedAnchor.Origin = AnchorOriginNative
			resolved.Anchors = append(resolved.Anchors, resolvedAnchor)
		}
	}

	return resolved, nil
}

// isLegacyAnchor 检查锚点是否为遗留锚点
func isLegacyAnchor(anchor Anchor) bool {
	return strings.HasPrefix(anchor.LineID, "legacy::")
}

// resolveLegacyAnchor 解析遗留锚点
func resolveLegacyAnchor(ctx ResolveContext, anchor Anchor) (ResolvedAnchor, error) {
	// 从遗留 LineID 中提取行号
	var row int
	// 这里简化处理，实际实现需要解析 "legacy::pane::<paneID>::row::<row>" 格式
	// 使用 engine.go 中的 clamp 函数
	if len(ctx.Snapshot.Lines) > row {
		line := ctx.Snapshot.Lines[row]
		return ResolvedAnchor{
			PaneID: anchor.PaneID,
			LineID: line.ID, // 使用快照中的稳定 ID
			Range: TextRange{
				Start: clamp(anchor.Start, 0, len(line.Text)),
				End:   clamp(anchor.End, 0, len(line.Text)),
			},
		}, nil
	}

	// 如果找不到对应行，返回错误
	return ResolvedAnchor{}, errors.New(ErrLineNotFound)
}

// resolveNativeAnchor 解析原生锚点
func resolveNativeAnchor(ctx ResolveContext, anchor Anchor) (ResolvedAnchor, error) {
	// 根据锚点类型解析
	switch anchor.Kind {
	case int(TargetPosition):
		// 如果锚点引用光标位置
		if ref, ok := anchor.Ref.(CursorRef); ok {
			cursorState, err := CursorRefToState(ref, ctx.Snapshot)
			if err != nil {
				return ResolvedAnchor{}, err
			}

			return ResolvedAnchor{
				PaneID: anchor.PaneID,
				LineID: cursorState.LineID,
				Range: TextRange{
					Start: cursorState.Offset,
					End:   cursorState.Offset,
				},
			}, nil
		}
		// 如果没有引用光标，使用锚点中的信息
		return ResolvedAnchor{
			PaneID: anchor.PaneID,
			LineID: anchor.LineID,
			Range: TextRange{
				Start: anchor.Start,
				End:   anchor.End,
			},
		}, nil
	default:
		// 其他类型的锚点处理
		return ResolvedAnchor{
			PaneID: anchor.PaneID,
			LineID: anchor.LineID,
			Range: TextRange{
				Start: anchor.Start,
				End:   anchor.End,
			},
		}, nil
	}
}

// NOTE: Undo/Redo anchors are for projection compatibility only.
// Resolver MUST ignore anchor for history-based intents.
func resolveUndoIntent(ctx ResolveContext, intent Intent) (ResolvedIntent, error) {
	// Undo 意图的解析主要是为了保持投影兼容性
	// 实际的撤销操作由专门的 UndoManager 处理
	resolved := ResolvedIntent{
		Intent:  intent,
		Anchors: []ResolvedAnchor{},
	}

	// 为 Undo 意图添加当前光标位置的锚点，用于投影兼容性
	cursorAnchor := ResolvedAnchor{
		PaneID: intent.PaneID,
		LineID: ctx.Cursor.LineID,
		Range: TextRange{
			Start: ctx.Cursor.Offset,
			End:   ctx.Cursor.Offset,
		},
		Origin: AnchorOriginNative, // Undo 意图使用原生锚点
	}

	resolved.Anchors = append(resolved.Anchors, cursorAnchor)

	return resolved, nil
}

// resolveRedoIntent 解析重做意图
func resolveRedoIntent(ctx ResolveContext, intent Intent) (ResolvedIntent, error) {
	// Redo 意图的解析主要是为了保持投影兼容性
	// 实际的重做操作由专门的 UndoManager 处理
	resolved := ResolvedIntent{
		Intent:  intent,
		Anchors: []ResolvedAnchor{},
	}

	// 为 Redo 意图添加当前光标位置的锚点，用于投影兼容性
	cursorAnchor := ResolvedAnchor{
		PaneID: intent.PaneID,
		LineID: ctx.Cursor.LineID,
		Range: TextRange{
			Start: ctx.Cursor.Offset,
			End:   ctx.Cursor.Offset,
		},
		Origin: AnchorOriginNative, // Redo 意图使用原生锚点
	}

	resolved.Anchors = append(resolved.Anchors, cursorAnchor)

	return resolved, nil
}

// AssertNoLegacy 确保解析后的意图不包含遗留锚点
func (r ResolvedIntent) AssertNoLegacy() {
	for _, anchor := range r.Anchors {
		if anchor.Origin == AnchorOriginLegacy {
			panic("legacy anchor leaked past resolver")
		}
	}
}

// 错误定义
var ErrLineNotFound = "line not found"