package core

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// SnapshotHash 快照哈希（世界指纹）
type SnapshotHash string

// LineHash 行哈希（局部指纹）
type LineHash string

// LineID 行ID（基于内容的稳定ID）
type LineID string

// Snapshot 世界快照（不可变）
// 代表 Intent 形成时对世界的冻结视图
// Now uses stable LineID for diffing capabilities
type Snapshot struct {
	PaneID string
	Cursor CursorPos
	Lines  []LineSnapshot
	Index  map[LineID]int // Stable mapping from LineID to position
	Hash   SnapshotHash
	TakenAt time.Time
}

// CursorPos 光标位置
type CursorPos struct {
	Row int
	Col int
}

// LineSnapshot 单行快照
type LineSnapshot struct {
	ID   LineID // Stable ID based on content
	Text string
	Hash LineHash
}

// TakeSnapshot creates a new snapshot with stable LineIDs
func TakeSnapshot(paneID string, cursor CursorPos, lines []string) Snapshot {
	snaps := make([]LineSnapshot, 0, len(lines))
	index := make(map[LineID]int, len(lines))

	var prev LineID

	for i, text := range lines {
		id := makeLineID(paneID, prev, text)
		hash := hashLine(text)

		snap := LineSnapshot{
			ID:   id,
			Text: text,
			Hash: hash,
		}

		snaps = append(snaps, snap)
		index[id] = i
		prev = id
	}

	snapshot := Snapshot{
		PaneID: paneID,
		Cursor: cursor,
		Lines:  snaps,
		Index:  index,
		TakenAt: time.Now(),
	}

	snapshot.Hash = computeSnapshotHash(snapshot)
	return snapshot
}

// makeLineID creates a stable LineID based on content
func makeLineID(paneID string, prev LineID, text string) LineID {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s", paneID, prev, text)))
	return LineID(fmt.Sprintf("%x", h[:]))
}

// hashLine computes the hash of a line
func hashLine(text string) LineHash {
	h := sha256.Sum256([]byte(text))
	return LineHash(fmt.Sprintf("%x", h[:]))
}

// computeSnapshotHash computes the hash of a snapshot
func computeSnapshotHash(s Snapshot) SnapshotHash {
	h := sha256.New()
	for _, l := range s.Lines {
		h.Write([]byte(l.ID))
		h.Write([]byte(l.Hash))
	}
	return SnapshotHash(fmt.Sprintf("%x", h.Sum(nil)))
}
