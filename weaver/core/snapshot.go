package core

import "time"

// SnapshotHash 快照哈希（世界指纹）
type SnapshotHash string

// LineHash 行哈希（局部指纹）
type LineHash string

// Snapshot 世界快照（不可变）
// 代表 Intent 形成时对世界的冻结视图
type Snapshot struct {
	PaneID string

	Cursor CursorPos
	Lines  []LineSnapshot

	Hash    SnapshotHash
	TakenAt time.Time
}

// CursorPos 光标位置
type CursorPos struct {
	Row int
	Col int
}

// LineSnapshot 单行快照
type LineSnapshot struct {
	Row  int
	Text string
	Hash LineHash
}
