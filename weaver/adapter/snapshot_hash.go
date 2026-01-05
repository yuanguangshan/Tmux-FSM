package adapter

import (
	"crypto/sha256"
	"encoding/hex"
	"tmux-fsm/weaver/core"
)

// ❌ DEPRECATED: Do NOT use this
// SnapshotHash must be computed by core.TakeSnapshot only.
func computeSnapshotHash(s core.Snapshot) core.SnapshotHash {
	h := sha256.New()

	h.Write([]byte(s.PaneID))
	for _, line := range s.Lines {
		h.Write([]byte(line.Hash))
	}

	return core.SnapshotHash(hex.EncodeToString(h.Sum(nil)))
}
