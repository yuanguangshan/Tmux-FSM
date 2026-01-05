package core

// Engine Weaver Core 引擎接口
// 这是整个系统的唯一入口
type Engine interface {
	// ApplyIntent 应用一个意图
	// 返回裁决结果，包含安全级别和审计信息
	ApplyIntent(intent Intent, resolver AnchorResolver, projection Projection) (*Verdict, error)
}

// AnchorResolver Anchor 解析器接口
// 由环境层实现（tmux, vim, etc.）
type AnchorResolver interface {
	// Resolve 解析一个 Anchor 到具体位置
	Resolve(anchor Anchor) (ResolvedAnchor, AnchorResolution, error)
}

// Projection 投影接口
// 将 Fact 投影到实际环境（tmux send-keys, vim commands, etc.）
type Projection interface {
	// Apply 应用一组 Facts
	Apply(resolved []ResolvedAnchor, facts []Fact) error
}

// Intent 意图接口（从主包导入）
type Intent interface {
	GetKind() IntentKind
	GetTarget() SemanticTarget
	GetCount() int
	GetMeta() map[string]interface{}
	GetPaneID() string // 新增：Phase 3 需要
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
)

// SemanticTarget 语义目标
type SemanticTarget struct {
	Kind      int
	Direction string
	Scope     string
	Value     string
}

// Planner 规划器接口
// 负责将 Intent 转换为 Facts
type Planner interface {
	Build(intent Intent, paneID string) ([]Fact, []Fact, error)
}
