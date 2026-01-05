package core

// Fact 表示一个已发生的编辑事实（不可变）
// 这是 Weaver Core 的核心数据结构
type Fact struct {
	Kind        FactKind               `json:"kind"`
	Anchor      Anchor                 `json:"anchor"`
	Range       Range                  `json:"range"`
	Payload     FactPayload            `json:"payload"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	Timestamp   int64                  `json:"timestamp"`
	SideEffects []string               `json:"side_effects,omitempty"`
}

// FactKind 事实类型
type FactKind int

const (
	FactNone FactKind = iota
	FactInsert
	FactDelete
	FactReplace
	FactMove
)

// Anchor 语义锚点（"我指的不是光标，而是这段文本"）
type Anchor struct {
	ResourceID string     `json:"resource_id"` // pane_id, buffer_id, etc.
	Hint       AnchorHint `json:"hint"`
	Hash       []byte     `json:"hash"`
	Offset     int        `json:"offset"`
}

// AnchorHint 锚点提示（用于快速定位）
type AnchorHint struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Range 范围（基于 Anchor）
type Range struct {
	StartOffset int `json:"start_offset"`
	EndOffset   int `json:"end_offset"`
}

// FactPayload 事实的具体内容
type FactPayload struct {
	Text     string `json:"text,omitempty"`
	OldText  string `json:"old_text,omitempty"`
	NewText  string `json:"new_text,omitempty"`
	Value    string `json:"value,omitempty"`
	Position int    `json:"position,omitempty"`
}

// Transaction 事务
// 包含一组 Facts，具有原子性
type Transaction struct {
	ID           TransactionID `json:"id"`
	Intent       Intent        `json:"intent"`        // 原始意图
	Facts        []Fact        `json:"facts"`         // 正向事实序列
	InverseFacts []Fact        `json:"inverse_facts"` // 反向事实序列（用于 Undo）
	Safety       SafetyLevel   `json:"safety"`
	Timestamp    int64         `json:"timestamp"`
	Applied      bool          `json:"applied"`
	Skipped      bool          `json:"skipped"`
}

// TransactionID 事务 ID
type TransactionID string

// SafetyLevel 安全级别
type SafetyLevel int

const (
	SafetyExact SafetyLevel = iota
	SafetyFuzzy
	SafetyUnsafe
)

// Verdict 裁决结果（可审计输出）
type Verdict struct {
	Kind        VerdictKind        `json:"kind"`
	Safety      SafetyLevel        `json:"safety"`
	Message     string             `json:"message"`
	Transaction *Transaction       `json:"transaction,omitempty"`
	Resolutions []AnchorResolution `json:"resolutions,omitempty"`
	Details     []AuditEntry       `json:"details,omitempty"`
}

// VerdictKind 裁决类型
type VerdictKind int

const (
	VerdictApplied VerdictKind = iota
	VerdictRejected
	VerdictSkipped
)

// AuditEntry 审计条目
type AuditEntry struct {
	Step   string `json:"step"`
	Result string `json:"result"`
}

// AnchorResolution Anchor 解析结果
type AnchorResolution int

const (
	AnchorExact AnchorResolution = iota
	AnchorFuzzy
	AnchorFailed
)

// ResolvedAnchor 已解析的 Anchor
type ResolvedAnchor struct {
	ResourceID string
	Offset     int
	Resolution AnchorResolution
}
