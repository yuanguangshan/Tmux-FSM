package intent

import (
	"tmux-fsm/weaver/core"
)

// IntentKind 意图类型
type IntentKind = core.IntentKind

const (
	IntentNone              = core.IntentNone
	IntentMove              = core.IntentMove
	IntentDelete            = core.IntentDelete
	IntentChange            = core.IntentChange
	IntentYank              = core.IntentYank
	IntentInsert            = core.IntentInsert
	IntentPaste             = core.IntentPaste
	IntentUndo              = core.IntentUndo
	IntentRedo              = core.IntentRedo
	IntentSearch            = core.IntentSearch
	IntentVisual            = core.IntentVisual
	IntentToggleCase        = core.IntentToggleCase
	IntentReplace           = core.IntentReplace
	IntentRepeat            = core.IntentRepeat
	IntentFind              = core.IntentFind
	IntentExit              = core.IntentExit
	IntentCount             = core.IntentCount
	IntentOperator          = core.IntentOperator
	IntentMotion            = core.IntentMotion
	IntentMacro             = core.IntentMacro
	IntentEnterVisual       = core.IntentEnterVisual
	IntentExitVisual        = core.IntentExitVisual
	IntentExtendSelection   = core.IntentExtendSelection
	IntentOperatorSelection = core.IntentOperatorSelection
	IntentRepeatFind        = core.IntentRepeatFind
	IntentRepeatFindReverse = core.IntentRepeatFindReverse
)

// OperatorKind 操作符类型
type OperatorKind int

const (
	OpMove OperatorKind = iota
	OpDelete
	OpYank
	OpChange
)

// TargetKind 目标类型
type TargetKind = core.TargetKind

const (
	TargetNone       = core.TargetNone
	TargetUnknown    = core.TargetUnknown
	TargetChar       = core.TargetChar
	TargetWord       = core.TargetWord
	TargetLine       = core.TargetLine
	TargetFile       = core.TargetFile
	TargetTextObject = core.TargetTextObject
	TargetPosition   = core.TargetPosition
	TargetSearch     = core.TargetSearch
)

// RangeType 范围类型
type RangeType int

const (
	Exclusive RangeType = iota
	Inclusive
	LineWise
)

// VisualMode 视觉模式类型
type VisualMode int

const (
	VisualNone VisualMode = iota
	VisualChar
	VisualLine
	VisualBlock
)

// Intent 意图结构（用于执行层）
type Intent struct {
	Kind         IntentKind             `json:"kind"`
	Target       SemanticTarget         `json:"target,omitempty"` // ⚠️ DEPRECATED — migration only
	Count        int                    `json:"count"`
	Meta         map[string]interface{} `json:"meta,omitempty"` // ⚠️ DEPRECATED — migration only
	PaneID       string                 `json:"pane_id"`
	SnapshotHash string                 `json:"snapshot_hash"`      // Phase 6.2
	AllowPartial bool                   `json:"allow_partial"`      // Phase 7: Explicit permission for fuzzy resolution
	Anchors      []Anchor               `json:"anchors,omitempty"`  // Phase 11.0: Support for multi-cursor / multi-selection
	UseRange     bool                   `json:"use_range"`          // Phase 12: Use range-based operations
	Motion       *Motion                `json:"motion,omitempty"`   // ✅ 新增：强类型 Motion 结构
	Operator     *OperatorKind          `json:"operator,omitempty"` // ✅ 新增：强类型 Operator 结构
}

// SemanticTarget 语义目标（而非物理位置）
type SemanticTarget = core.SemanticTarget

// Anchor 锚点结构
type Anchor = core.Anchor

// GetKind 获取意图类型
func (i Intent) GetKind() core.IntentKind {
	return i.Kind
}

// GetTarget 获取语义目标
func (i Intent) GetTarget() core.SemanticTarget {
	return i.Target
}

// GetCount 获取计数
func (i Intent) GetCount() int {
	return i.Count
}

// GetMeta 获取元数据
func (i Intent) GetMeta() map[string]interface{} {
	return i.Meta
}

// GetPaneID 获取面板ID
func (i Intent) GetPaneID() string {
	return i.PaneID
}

// GetSnapshotHash 获取快照哈希
func (i Intent) GetSnapshotHash() string {
	return i.SnapshotHash
}

// IsPartialAllowed 是否允许部分匹配
func (i Intent) IsPartialAllowed() bool {
	return i.AllowPartial
}

// GetAnchors 获取锚点
func (i Intent) GetAnchors() []core.Anchor {
	return i.Anchors
}

// GetOperator 获取操作符
func (i Intent) GetOperator() *int {
	if i.Operator == nil {
		return nil
	}
	val := int(*i.Operator)
	return &val
}
