package builder

import (
	"tmux-fsm/intent"
)

// BuildContext 构建上下文
type BuildContext struct {
	Action   string                 // legacy action string
	Command  string                 // normalized command (future)
	Count    int
	PaneID   string
	SnapshotHash string
	Meta     map[string]interface{} // 额外元数据
}

// Builder Intent构建器接口
type Builder interface {
	// Priority determines evaluation order.
	// Higher value = higher priority.
	Priority() int
	Build(ctx BuildContext) (*intent.Intent, bool)
}

