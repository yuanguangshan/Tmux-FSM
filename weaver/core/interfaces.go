package core

// Engine Weaver Core 引擎接口
// 这是整个系统的唯一入口
type Engine interface {
	// ApplyIntent 应用一个意图
	// 返回裁决结果，包含安全级别和审计信息
	ApplyIntent(intent Intent, resolver AnchorResolver, projection Projection) (*Verdict, error)

	// Undo 撤销最后一个事务
	Undo() (*Verdict, error)

	// Redo 重做最后一个撤销的事务
	Redo() (*Verdict, error)
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
// 这里只是声明，实际定义在主包的 intent.go
type Intent interface {
	GetKind() IntentKind
	GetTarget() SemanticTarget
	GetCount() int
}

// IntentKind 意图类型（占位符）
type IntentKind int

// SemanticTarget 语义目标（占位符）
type SemanticTarget struct {
	Kind      int
	Direction string
	Scope     string
	Value     string
}
