package types

import (
	"time"
	"tmux-fsm/weaver/core"
	"tmux-fsm/editor"
)

// TransactionID 事务ID类型
type TransactionID uint64

// OperationRecord 操作记录，基于ResolvedOperation
type OperationRecord struct {
	ResolvedOp editor.ResolvedOperation `json:"resolved_operation"`
	Inverse    editor.ResolvedOperation `json:"inverse"`  // 用于 undo 的反向操作
	Fact       core.Fact               `json:"fact"`
}

// Transaction 事务
// RFC-WC-003: Audit Trail - 所有编辑操作必须可追溯
// 
// 更新：现在使用ResolvedOperation作为核心操作表示
// 这样可以更好地支持Vim语义的repeat/undo操作
type Transaction struct {
	ID               TransactionID      `json:"id"`
	Records          []OperationRecord  `json:"records"`
	CreatedAt        time.Time          `json:"created_at"`
	Applied          bool               `json:"applied"`
	Skipped          bool               `json:"skipped"`
	SafetyLevel      string             `json:"safety_level,omitempty"`       // exact, fuzzy
	PreSnapshotHash  string             `json:"pre_snapshot_hash,omitempty"`  // Phase 8: World state before transaction
	PostSnapshotHash string             `json:"post_snapshot_hash,omitempty"` // Phase 8: World state after transaction
}