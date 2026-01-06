package intent

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
)

// SemanticTarget 语义目标（而非物理位置）
type SemanticTarget struct {
	Kind      int    `json:"kind"`
	Direction string `json:"direction,omitempty"` // forward, backward
	Scope     string `json:"scope,omitempty"`     // char, line, word, etc.
	Value     string `json:"value,omitempty"`     // 用于搜索、替换等
}

// Anchor 锚点结构
type Anchor struct {
	PaneID string      `json:"pane_id"`
	Kind   int         `json:"kind"`
	Ref    interface{} `json:"ref,omitempty"`
	Hash   string      `json:"hash,omitempty"` // Phase 5.4: Reconciliation Expectation
	LineID string      `json:"line_id,omitempty"` // Phase 9: Stable line identifier
	Start  int         `json:"start,omitempty"`   // Phase 11: Start position in line
	End    int         `json:"end,omitempty"`     // Phase 11: End position in line
}

// Intent 意图结构
type Intent struct {
	Kind         IntentKind             `json:"kind"`
	Target       SemanticTarget         `json:"target"`
	Count        int                    `json:"count"`
	Meta         map[string]interface{} `json:"meta,omitempty"`
	PaneID       string                 `json:"pane_id"`
	SnapshotHash string                 `json:"snapshot_hash"` // Phase 6.2
	AllowPartial bool                   `json:"allow_partial"` // Phase 7: Explicit permission for fuzzy resolution
	Anchors      []Anchor               `json:"anchors,omitempty"` // Phase 11.0: Support for multi-cursor / multi-selection
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