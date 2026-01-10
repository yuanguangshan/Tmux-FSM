package core

import (
	"errors"
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

// SafetyLevel 安全级别
type SafetyLevel int

const (
	SafetyExact SafetyLevel = iota
	SafetyFuzzy
	SafetyUnsafe
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

// Anchor 描述“我们想要操作的目标”，而不是“它在哪里”
// Phase 5.3: 纯语义 Anchor
type Anchor struct {
	PaneID string     `json:"pane_id"`
	Kind   AnchorKind `json:"kind"`
	Ref    any        `json:"ref,omitempty"`
	Hash   string     `json:"hash,omitempty"`    // Phase 5.4: Reconciliation Expectation
	LineID LineID     `json:"line_id,omitempty"` // Phase 9: Stable line identifier
	Start  int        `json:"start,omitempty"`   // Phase 11: Start position in line
	End    int        `json:"end,omitempty"`     // Phase 11: End position in line
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
	ProofHash        string        `json:"proof_hash,omitempty"`         // Anchor for proof verification
}

// VerificationResult for verifier
type VerificationResult struct {
	OK      bool
	Safety  SafetyLevel
	Diffs   []SnapshotDiff
	Message string
}

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

// AuditEntry 审计条目 (v1 - legacy)
type AuditEntry struct {
	Step   string `json:"step"`
	Result string `json:"result"`
}

// AuditRecord v2 - 完整的审计记录
type AuditRecord struct {
	Version       string `json:"version"`
	RequestID     string `json:"request_id"`
	TransactionID string `json:"transaction_id"`
	ActorID       string `json:"actor_id"`
	TimestampUTC  int64  `json:"timestamp_utc"` // Unix timestamp

	IntentKind   string `json:"intent_kind"`
	DecisionPath string `json:"decision_path"`

	Entries []AuditEntryV2 `json:"entries"`
	Result  AuditResult    `json:"result"`
}

// AuditEntryV2 - 结构化的审计条目 (v2)
type AuditEntryV2 struct {
	Phase   string            `json:"phase"`
	Action  string            `json:"action"`
	Outcome string            `json:"outcome"`
	Detail  string            `json:"detail"`
	Meta    map[string]string `json:"meta"`
	At      int64             `json:"at"` // Unix timestamp
}

// AuditResult - 审计结果
type AuditResult struct {
	Status      string `json:"status"` // Committed / Rejected / RolledBack
	WorldDrift  bool   `json:"world_drift"`
	DriftReason string `json:"drift_reason,omitempty"`
	Error       string `json:"error,omitempty"`
}

// DriftReason - 漂移原因类型
type DriftReason string

const (
	DriftSnapshotMismatch DriftReason = "snapshot_mismatch"
	DriftUndoMismatch     DriftReason = "undo_mismatch"
	DriftRedoMismatch     DriftReason = "redo_mismatch"
)

// WorldDriftError - 带原因的世界漂移错误
type WorldDriftError struct {
	Reason   DriftReason
	Expected string
	Actual   string
	Message  string
}

func (e *WorldDriftError) Error() string {
	return e.Message
}

// Proof - 证明对象
type Proof struct {
	TransactionID string `json:"transaction_id"`
	PreStateHash  string `json:"pre_state_hash"`
	PostStateHash string `json:"post_state_hash"`
	FactsHash     string `json:"facts_hash"`
	AuditHash     string `json:"audit_hash"`
}

// AnchorResolution Anchor 解析结果
type AnchorResolution int

const (
	AnchorExact AnchorResolution = iota
	AnchorFuzzy
	AnchorFailed
)

// HandleContext 用于传递请求上下文信息
type HandleContext struct {
	Ctx       interface{} // Using interface{} as context.Context might not be available here
	RequestID string      // Unique identifier for this user request
	ActorID   string      // User / pane / client identifier
}

// UndoEntry represents a single undo operation
// Phase 12.0: Projection-level undo log
type UndoEntry struct {
	LineID LineID `json:"line_id"`
	Before string `json:"before"`
	After  string `json:"after"`
}
