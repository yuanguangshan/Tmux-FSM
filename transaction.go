package main

import (
	"time"
	"tmux-fsm/weaver/adapter"
	"tmux-fsm/weaver/core"
)

type TransactionID uint64

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

type TransactionManager struct {
	current *Transaction
	nextID  TransactionID
}

// takeSnapshotForPane takes a snapshot of the given pane using the global weaver manager
func takeSnapshotForPane(paneID string) (string, error) {
	if weaverMgr != nil && weaverMgr.snapshotProvider != nil {
		snapshot, err := weaverMgr.snapshotProvider.TakeSnapshot(paneID)
		if err != nil {
			return "", err
		}
		return string(snapshot.Hash), nil
	}

	// Fallback: Use direct tmux capture if weaver is not available
	// This is a simplified approach - we'll capture the current line and hash it
	cursor := adapter.TmuxGetCursorPos(paneID)
	lines := adapter.TmuxCapturePane(paneID)

	// Use the new snapshot structure with LineID
	snapshot := core.TakeSnapshot(paneID, core.CursorPos{
		Row: cursor[0],
		Col: cursor[1],
	}, lines)

	return string(snapshot.Hash), nil
}

// computeSnapshotHash computes the hash of a snapshot
// NOTE: This is currently "Pane-only" scoped (Phase 8)
// For Phase 9+ (Split/Multi-pane), this will need to be upgraded to "World-scoped"
// where the hash represents the state of the affected world subgraph, not just a single pane
// [Phase 9] This function is now redundant since core.TakeSnapshot already computes the hash
func computeSnapshotHash(s core.Snapshot) core.SnapshotHash {
	return s.Hash
}

func (tm *TransactionManager) Begin(paneID string) {
	tm.current = &Transaction{
		ID:        tm.nextID,
		CreatedAt: time.Now(),
		Records:   []ActionRecord{},
	}

	// Take a snapshot before any changes occur
	if hash, err := takeSnapshotForPane(paneID); err == nil {
		tm.current.PreSnapshotHash = hash
	}

	tm.nextID++
}

func (tm *TransactionManager) Append(r ActionRecord) {
	if tm.current != nil {
		tm.current.Records = append(tm.current.Records, r)
	}
}

func (tm *TransactionManager) Commit(
	stack *[]Transaction,
	paneID string,
) {
	// --- Phase 8.0: 空事务直接丢弃 ---
	if tm.current == nil || len(tm.current.Records) == 0 {
		tm.current = nil
		return
	}

	tx := tm.current

	// --- Phase 8.1: 记录 PostSnapshot（事实，不做判断） ---
	if hash, err := takeSnapshotForPane(paneID); err == nil {
		tx.PostSnapshotHash = hash
	}

	// --- Phase 8.2: 标记为 Applied（仅表示"已执行完成"） ---
	tx.Applied = true

	// --- Phase 8.3: 提交到 Legacy 时间线（只有非跳过事务） ---
	// [Phase 9] Only add to legacy stack if not in Weaver mode
	// Weaver becomes the single source of truth for undo/redo
	if !tx.Skipped && GetMode() != ModeWeaver {
		*stack = append(*stack, *tx)
	}

	// --- Phase 8.4: 注入 Weaver（只有"存在的事务"才允许） ---
	if weaverMgr != nil && !tx.Skipped {
		weaverMgr.InjectLegacyTransaction(tx)
	}

	// --- Phase 8.5: 结束事务 ---
	tm.current = nil
}