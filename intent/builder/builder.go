package builder

import (
	"tmux-fsm/intent"
)

// BuildContext 构建上下文
type BuildContext struct {
	Action   string  // legacy action string
	Command  string  // normalized command (future)
	Count    int
	PaneID   string
	SnapshotHash string
}

// Builder Intent构建器接口
type Builder interface {
	// Priority determines evaluation order.
	// Higher value = higher priority.
	Priority() int
	Build(ctx BuildContext) (*intent.Intent, bool)
}

// SemanticEqual 比较两个Intent的语义是否相等
func SemanticEqual(a, b *intent.Intent) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// 比较核心语义字段
	// NOTE: PaneID is included for migration safety.
	// Semantic equality SHOULD eventually ignore routing.
	return a.Kind == b.Kind &&
		   a.Target.Kind == b.Target.Kind &&
		   a.Target.Direction == b.Target.Direction &&
		   a.Target.Scope == b.Target.Scope &&
		   a.Target.Value == b.Target.Value &&
		   a.Count == b.Count &&
		   a.PaneID == b.PaneID
}