package types

import (
	"time"
	"tmux-fsm/weaver/core"
)

// TransactionID 事务ID类型
type TransactionID uint64

// ActionRecord 动作记录
type ActionRecord struct {
	Fact    core.Fact `json:"fact"`
	Inverse core.Fact `json:"inverse"`
}

// Transaction 事务
// RFC-WC-003: Audit Trail - 所有编辑操作必须可追溯
type Transaction struct {
	ID               TransactionID  `json:"id"`
	Records          []ActionRecord `json:"records"`
	CreatedAt        time.Time      `json:"created_at"`
	Applied          bool           `json:"applied"`
	Skipped          bool           `json:"skipped"`
	SafetyLevel      string         `json:"safety_level,omitempty"`       // exact, fuzzy
	PreSnapshotHash  string         `json:"pre_snapshot_hash,omitempty"`  // Phase 8: World state before transaction
	PostSnapshotHash string         `json:"post_snapshot_hash,omitempty"` // Phase 8: World state after transaction
}