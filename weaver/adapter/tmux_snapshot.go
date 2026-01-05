package adapter

import (
	"time"
	"tmux-fsm/weaver/core"
)

type TmuxSnapshotProvider struct{}

func (p *TmuxSnapshotProvider) TakeSnapshot(paneID string) (core.Snapshot, error) {
	cursor := TmuxGetCursorPos(paneID)
	lines := TmuxCapturePane(paneID)

	var snapLines []core.LineSnapshot
	for i, line := range lines {
		snapLines = append(snapLines, core.LineSnapshot{
			Row:  i,
			Text: line,
			Hash: core.LineHash(TmuxHashLine(line)),
		})
	}

	snapshot := core.Snapshot{
		PaneID: paneID,
		Cursor: core.CursorPos{
			Row: cursor[0],
			Col: cursor[1],
		},
		Lines:   snapLines,
		TakenAt: time.Now(),
	}

	snapshot.Hash = computeSnapshotHash(snapshot)
	return snapshot, nil
}
