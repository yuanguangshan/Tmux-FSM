package core

import (
	"crypto/sha256"
	"fmt"
)

func makeLineID(paneID string, prev LineID, text string) LineID {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s", paneID, prev, text)))
	return LineID(fmt.Sprintf("%x", h[:]))
}

func hashLine(text string) LineHash {
	h := sha256.Sum256([]byte(text))
	return LineHash(fmt.Sprintf("%x", h[:]))
}

func hashSnapshot(s Snapshot) SnapshotHash {
	h := sha256.New()
	for _, l := range s.Lines {
		h.Write([]byte(l.ID))
		h.Write([]byte(l.Hash))
	}
	return SnapshotHash(fmt.Sprintf("%x", h.Sum(nil)))
}
