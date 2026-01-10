package core

// Engine Weaver Core 引擎接口
// 这是整个系统的唯一入口
type Engine interface {
	// ApplyIntent 处理一个意图
	// Phase 6.2: 接收 Time-Frozen Snapshot
	// Phase X: 接收 HandleContext for RequestID/ActorID propagation
	ApplyIntent(hctx HandleContext, intent Intent, snapshot Snapshot) (*Verdict, error)
	GetHistory() History
}

// RealityReader 读取当前世界状态（用于一致性验证）
// Phase 6.3: 移至 core 以支持 Engine 级裁决
type RealityReader interface {
	ReadCurrent(paneID string) (Snapshot, error)
}

// EvidenceLibrary 证据库接口 (RFC-WC-003)
// 负责持久化存储审计笔录 (AuditRecord)，并提供基于 Hash 的检索
type EvidenceLibrary interface {
	Commit(record *AuditRecord) (string, error)
	Retrieve(hash string) (*AuditRecord, error)

	// Traverse 巡回复核能力: 允许第三方审计按照物理顺序遍历所有证据
	Traverse(fn func(meta EvidenceMeta) error) error
}

// EvidenceMeta 证据元数据
type EvidenceMeta struct {
	Hash      string `json:"hash"`
	Offset    int64  `json:"offset"`
	Timestamp int64  `json:"timestamp"`
	Size      int64  `json:"size"`
}

// AnchorResolver Anchor 解析器接口
// 由环境层实现（tmux, vim, etc.）
type AnchorResolver interface {
	// ResolveFacts 解析一组事实的 Anchor
	// Phase 5.2: 返回 ResolvedFact
	// Phase 6.3: 增加 expectedHash 用于一致性验证
	ResolveFacts(facts []Fact, expectedHash string) ([]ResolvedFact, error)
}

// Projection 投影接口
// 将 Fact 投影到实际环境（tmux send-keys, vim commands, etc.）
type Projection interface {
	// Apply 应用一组 ResolvedFacts (Phase 5.2)
	Apply(resolved []ResolvedAnchor, facts []ResolvedFact) ([]UndoEntry, error)
	// Rollback 回滚已应用的更改 (Phase 12.0)
	Rollback(log []UndoEntry) error
	// Verify 验证投影是否按预期执行 (Phase 9)
	Verify(pre Snapshot, facts []ResolvedFact, post Snapshot) VerificationResult
}

// Intent 意图接口（从主包导入）
type Intent interface {
	GetKind() IntentKind
	GetTarget() SemanticTarget
	GetCount() int
	GetMeta() map[string]interface{}
	GetPaneID() string
	GetSnapshotHash() string // Phase 6.2
	IsPartialAllowed() bool  // Phase 7: Explicit permission for fuzzy resolution
	GetAnchors() []Anchor    // Phase 11.0: Support for multi-cursor / multi-selection
	GetOperator() *int       // Added: Support for high-level operators
} // 新增：Phase 3 需要

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
	IntentCount
	IntentOperator
	IntentMotion
	IntentMacro
	IntentEnterVisual
	IntentExitVisual
	IntentExtendSelection
	IntentOperatorSelection
	IntentRepeatFind
	IntentRepeatFindReverse
)

func (k IntentKind) String() string {
	switch k {
	case IntentMove:
		return "MOVE"
	case IntentDelete:
		return "DELETE"
	case IntentChange:
		return "CHANGE"
	case IntentYank:
		return "YANK"
	case IntentInsert:
		return "INSERT"
	case IntentPaste:
		return "PASTE"
	case IntentUndo:
		return "UNDO"
	case IntentRedo:
		return "REDO"
	case IntentSearch:
		return "SEARCH"
	case IntentVisual:
		return "VISUAL"
	case IntentToggleCase:
		return "TOGGLE_CASE"
	case IntentReplace:
		return "REPLACE"
	case IntentRepeat:
		return "REPEAT"
	case IntentFind:
		return "FIND"
	case IntentExit:
		return "EXIT"
	case IntentCount:
		return "COUNT"
	case IntentOperator:
		return "OPERATOR"
	case IntentMotion:
		return "MOTION"
	case IntentMacro:
		return "MACRO"
	case IntentEnterVisual:
		return "ENTER_VISUAL"
	case IntentExitVisual:
		return "EXIT_VISUAL"
	case IntentExtendSelection:
		return "EXTEND_SELECTION"
	case IntentOperatorSelection:
		return "OPERATOR_SELECTION"
	case IntentRepeatFind:
		return "REPEAT_FIND"
	case IntentRepeatFindReverse:
		return "REPEAT_FIND_REVERSE"
	default:
		return "NONE"
	}
}

// TargetKind 目标类型
type TargetKind int

const (
	TargetNone TargetKind = iota
	TargetUnknown
	TargetChar
	TargetWord
	TargetLine
	TargetFile
	TargetTextObject
	TargetPosition
	TargetSearch
)

func (k TargetKind) String() string {
	switch k {
	case TargetChar:
		return "CHAR"
	case TargetWord:
		return "WORD"
	case TargetLine:
		return "LINE"
	case TargetFile:
		return "FILE"
	case TargetTextObject:
		return "TEXT_OBJECT"
	case TargetPosition:
		return "POSITION"
	case TargetSearch:
		return "SEARCH"
	default:
		return "UNKNOWN"
	}
}

// SemanticTarget 语义目标
type SemanticTarget struct {
	Kind      TargetKind
	Direction string
	Scope     string
	Value     string
}

// Planner 规划器接口
// 负责将 Intent 转换为 Facts
type Planner interface {
	// Build 根据意图和世界快照生成事实序列
	// Phase 6.2: Planner 变为纯函数，不读 IO
	Build(intent Intent, snapshot Snapshot) ([]Fact, []Fact, error)
}
