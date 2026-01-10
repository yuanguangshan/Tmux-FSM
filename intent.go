package main

// Intent 表示用户的编辑意图（语义层）
// 这是从 FSM 到执行器的中间层，将"按键序列"转换为"编辑语义"
type Intent struct {
	Kind         IntentKind             `json:"kind"`
	Target       SemanticTarget         `json:"target"`
	Count        int                    `json:"count"`
	Meta         map[string]interface{} `json:"meta,omitempty"`
	PaneID       string                 `json:"pane_id"`
	SnapshotHash string                 `json:"snapshot_hash"`     // Phase 6.2
	AllowPartial bool                   `json:"allow_partial"`     // Phase 7: Explicit permission for fuzzy resolution
	Anchors      []Anchor               `json:"anchors,omitempty"` // Phase 11.0: Support for multi-cursor / multi-selection
}

// GetPaneID 获取 PaneID
func (i Intent) GetPaneID() string {
	return i.PaneID
}

func (i Intent) GetKind() int {
	return int(i.Kind)
}

func (i Intent) GetSnapshotHash() string {
	return i.SnapshotHash
}

func (i Intent) IsPartialAllowed() bool {
	return i.AllowPartial
}

// GetAnchors returns the anchors for this intent
func (i Intent) GetAnchors() []Anchor {
	return i.Anchors
}

// IntentKind 意图类型
type IntentKind int

const (
	IntentNone IntentKind = iota
	IntentMove
	IntentDelete
	IntentChange
	IntentYank
	IntentInsert
	IntentPaste
	IntentUndo
	IntentRedo
	IntentSearch
	IntentVisual
	IntentToggleCase
	IntentReplace
	IntentRepeat
	IntentFind
	IntentExit
	// 快照相关意图
	IntentSnapshotUpdate
)

// SemanticTarget 语义目标（而非物理位置）
type SemanticTarget struct {
	Kind      TargetKind `json:"kind"`
	Direction string     `json:"direction,omitempty"` // forward, backward
	Scope     string     `json:"scope,omitempty"`     // char, line, word, etc.
	Value     string     `json:"value,omitempty"`     // 用于搜索、替换等
}

// TargetKind 目标类型
type TargetKind int

const (
	TargetNone TargetKind = iota
	TargetChar
	TargetWord
	TargetLine
	TargetFile
	TargetTextObject
	TargetPosition
	TargetSearch
)

// Anchor 锚点结构 (Phase 11.0)
type Anchor struct {
	PaneID string      `json:"pane_id"`
	Kind   int         `json:"kind"`
	Ref    interface{} `json:"ref,omitempty"`
	Hash   string      `json:"hash,omitempty"`    // Phase 5.4: Reconciliation Expectation
	LineID string      `json:"line_id,omitempty"` // Phase 9: Stable line identifier
	Start  int         `json:"start,omitempty"`   // Phase 11: Start position in line
	End    int         `json:"end,omitempty"`     // Phase 11: End position in line
}

// ToActionString 将 Intent 转换为 legacy action string
// 这是过渡期的桥接函数，最终会被移除
func (i Intent) ToActionString() string {
	if i.Kind == IntentNone {
		return ""
	}

	// 特殊处理：直接返回的动作
	switch i.Kind {
	case IntentUndo:
		return "undo"
	case IntentRedo:
		return "redo"
	case IntentRepeat:
		return "repeat_last"
	case IntentExit:
		return "exit"
	}

	// 组合型动作
	var action string

	// 操作类型
	switch i.Kind {
	case IntentMove:
		action = "move"
	case IntentDelete:
		action = "delete"
	case IntentChange:
		action = "change"
	case IntentYank:
		action = "yank"
	case IntentInsert:
		action = "insert"
	case IntentPaste:
		action = "paste"
	case IntentSearch:
		if i.Target.Value != "" {
			return "search_forward_" + i.Target.Value
		}
		if i.Target.Direction == "next" {
			return "search_next"
		}
		if i.Target.Direction == "prev" {
			return "search_prev"
		}
		return ""
	case IntentVisual:
		if i.Target.Scope == "char" {
			return "start_visual_char"
		}
		if i.Target.Scope == "line" {
			return "start_visual_line"
		}
		if i.Meta != nil {
			if op, ok := i.Meta["operation"].(string); ok {
				return "visual_" + op
			}
		}
		return "cancel_selection"
	case IntentToggleCase:
		return "toggle_case"
	case IntentReplace:
		if i.Target.Value != "" {
			return "replace_char_" + i.Target.Value
		}
		return ""
	case IntentFind:
		// Native Target support
		if i.Target.Value != "" {
			char := i.Target.Value
			var fType string
			if i.Target.Direction == "next" {
				if i.Target.Scope == "inclusive" {
					fType = "f"
				} else {
					fType = "t"
				}
			} else {
				if i.Target.Scope == "inclusive" {
					fType = "F"
				} else {
					fType = "T"
				}
			}
			if fType != "" {
				return "find_" + fType + "_" + char
			}
		}

		if i.Meta != nil {
			if fType, ok := i.Meta["find_type"].(string); ok {
				if char, ok := i.Meta["char"].(string); ok {
					return "find_" + fType + "_" + char
				}
			}
		}
		return ""
	}

	// 目标/运动
	var motion string
	switch i.Target.Kind {
	case TargetChar:
		if i.Target.Direction == "left" {
			motion = "left"
		} else if i.Target.Direction == "right" {
			motion = "right"
		}
	case TargetWord:
		if i.Target.Direction == "forward" {
			motion = "word_forward"
		} else if i.Target.Direction == "backward" {
			motion = "word_backward"
		} else if i.Target.Scope == "end" {
			motion = "end_of_word"
		}
	case TargetLine:
		if i.Target.Scope == "start" {
			motion = "start_of_line"
		} else if i.Target.Scope == "end" {
			motion = "end_of_line"
		} else if i.Target.Scope == "whole" {
			motion = "line"
		}
	case TargetFile:
		if i.Target.Scope == "start" {
			motion = "start_of_file"
		} else if i.Target.Scope == "end" {
			motion = "end_of_file"
		}
	case TargetPosition:
		if i.Target.Direction == "up" {
			motion = "up"
		} else if i.Target.Direction == "down" {
			motion = "down"
		}
	case TargetTextObject:
		// 文本对象：inside_word, around_quote, etc.
		motion = i.Target.Value
	}

	// Insert 的特殊位置
	if i.Kind == IntentInsert {
		if i.Target.Scope == "before" {
			return "insert_before"
		} else if i.Target.Scope == "after" {
			return "insert_after"
		} else if i.Target.Scope == "start_of_line" {
			return "insert_start_of_line"
		} else if i.Target.Scope == "end_of_line" {
			return "insert_end_of_line"
		} else if i.Target.Scope == "open_below" {
			return "insert_open_below"
		} else if i.Target.Scope == "open_above" {
			return "insert_open_above"
		}
	}

	// Paste 的特殊位置
	if i.Kind == IntentPaste {
		if i.Target.Scope == "after" {
			return "paste_after"
		} else if i.Target.Scope == "before" {
			return "paste_before"
		}
	}

	if motion == "" {
		return ""
	}

	return action + "_" + motion
}
