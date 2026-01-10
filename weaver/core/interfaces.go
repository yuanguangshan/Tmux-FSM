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
