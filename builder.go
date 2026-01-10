package main

// IntentBuilder 是用于创建 Intent 的构建器
// 这是 Native Intent 的唯一入口，取代了 legacy intent bridge
type IntentBuilder struct {
	paneID string
	cursor CursorRef
}

// CursorRef 表示光标引用（语义位置，而非物理坐标）
type CursorRef struct {
	Kind CursorKind
}

// CursorKind 定义光标类型
type CursorKind int

const (
	CursorPrimary CursorKind = iota
	CursorSelectionStart
	CursorSelectionEnd
)

// NewIntentBuilder 创建新的 IntentBuilder 实例
func NewIntentBuilder(paneID string) *IntentBuilder {
	return &IntentBuilder{
		paneID: paneID,
		cursor: CursorRef{Kind: CursorPrimary},
	}
}

// IntentBuilder MUST NOT:
// - read snapshot
// - know row / col
// - depend on tmux / screen
//
// IntentBuilder 只表达"我想做什么"，而不是"我在屏幕的哪一格"

// Move 创建移动意图
func (b *IntentBuilder) Move(target SemanticTarget, count int) Intent {
	return Intent{
		Kind:   IntentMove,
		Target: target,
		Count:  count,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Delete 创建删除意图
func (b *IntentBuilder) Delete(target SemanticTarget, count int) Intent {
	return Intent{
		Kind:   IntentDelete,
		Target: target,
		Count:  count,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Change 创建修改意图
func (b *IntentBuilder) Change(target SemanticTarget, count int) Intent {
	return Intent{
		Kind:   IntentChange,
		Target: target,
		Count:  count,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Yank 创建复制意图
func (b *IntentBuilder) Yank(target SemanticTarget, count int) Intent {
	return Intent{
		Kind:   IntentYank,
		Target: target,
		Count:  count,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Insert 创建插入意图
func (b *IntentBuilder) Insert(target SemanticTarget, count int) Intent {
	return Intent{
		Kind:   IntentInsert,
		Target: target,
		Count:  count,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Paste 创建粘贴意图
func (b *IntentBuilder) Paste(target SemanticTarget, count int) Intent {
	return Intent{
		Kind:   IntentPaste,
		Target: target,
		Count:  count,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Undo 创建撤销意图
func (b *IntentBuilder) Undo() Intent {
	return Intent{
		Kind:   IntentUndo,
		PaneID: b.paneID,
		// Undo/Redo anchors are for projection compatibility only.
		// Resolver MUST ignore anchor for history-based intents.
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Redo 创建重做意图
func (b *IntentBuilder) Redo() Intent {
	return Intent{
		Kind:   IntentRedo,
		PaneID: b.paneID,
		// Undo/Redo anchors are for projection compatibility only.
		// Resolver MUST ignore anchor for history-based intents.
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Search 创建搜索意图
func (b *IntentBuilder) Search(target SemanticTarget) Intent {
	return Intent{
		Kind:   IntentSearch,
		Target: target,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Visual 创建视觉模式意图
func (b *IntentBuilder) Visual(target SemanticTarget) Intent {
	return Intent{
		Kind:   IntentVisual,
		Target: target,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// ToggleCase 创建切换大小写意图
func (b *IntentBuilder) ToggleCase() Intent {
	return Intent{
		Kind:   IntentToggleCase,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Replace 创建替换意图
func (b *IntentBuilder) Replace(target SemanticTarget) Intent {
	return Intent{
		Kind:   IntentReplace,
		Target: target,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Repeat 创建重复意图
func (b *IntentBuilder) Repeat() Intent {
	return Intent{
		Kind:   IntentRepeat,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Find 创建查找意图
func (b *IntentBuilder) Find(target SemanticTarget) Intent {
	return Intent{
		Kind:   IntentFind,
		Target: target,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// Exit 创建退出意图
func (b *IntentBuilder) Exit() Intent {
	return Intent{
		Kind:   IntentExit,
		PaneID: b.paneID,
		Anchors: []Anchor{
			CursorAnchor(b.cursor),
		},
	}
}

// CursorAnchor 创建光标锚点
func CursorAnchor(ref CursorRef) Anchor {
	return Anchor{
		Kind: int(TargetPosition), // 使用位置类型的锚点
		Ref:  ref,                 // 使用 CursorRef 作为引用
	}
}

// DEPRECATED: Meta["line_id"] is legacy-only. Do not use in new code.
// All new code should rely on Anchor structures for positional information.
