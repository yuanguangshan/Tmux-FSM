package core

import (
	"errors"
)

// ErrWorldDrift 世界漂移错误（快照不匹配）
// 表示 Intent 基于的历史与当前现实不一致
var ErrWorldDrift = errors.New("world drift: snapshot mismatch")

// Fact 表示一个已发生的编辑事实（不可变）
// 这是 Weaver Core 的核心数据结构
// Phase 5.3: 不再包含物理 Range
type Fact struct {
	Kind        FactKind               `json:"kind"`
	Anchor      Anchor                 `json:"anchor"`
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

// AnchorKind 锚点类型
type AnchorKind int

const (
	AnchorNone AnchorKind = iota
	AnchorAtCursor
	AnchorWord
	AnchorLine
	AnchorAbsolute
	AnchorLegacyRange
)

// Anchor 描述“我们想要操作的目标”，而不是“它在哪里”
// Phase 5.3: 纯语义 Anchor
type Anchor struct {
	PaneID string     `json:"pane_id"`
	Kind   AnchorKind `json:"kind"`
	Ref    any        `json:"ref,omitempty"`
	Hash   string     `json:"hash,omitempty"` // Phase 5.4: Reconciliation Expectation
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
	ID               TransactionID `json:"id"`
	Intent           Intent        `json:"intent"`        // 原始意图
	Facts            []Fact        `json:"facts"`         // 正向事实序列
	InverseFacts     []Fact        `json:"inverse_facts"` // 反向事实序列（用于 Undo）
	Safety           SafetyLevel   `json:"safety"`
	Timestamp        int64         `json:"timestamp"`
	Applied          bool          `json:"applied"`
	Skipped          bool          `json:"skipped"`
	PostSnapshotHash string        `json:"post_snapshot_hash,omitempty"` // Phase 7: State after application
	AllowPartial     bool          `json:"allow_partial,omitempty"`      // Phase 7: Explicit flag for fuzzy match
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
	Audit       []AuditEntry       `json:"audit,omitempty"` // Renamed from Details
}

// VerdictKind 裁决类型
type VerdictKind int

const (
	VerdictApplied VerdictKind = iota
	VerdictRejected
	VerdictSkipped
	VerdictBlocked // Phase 5.4: Blocked by Reconciliation
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
