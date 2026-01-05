package adapter

import (
	"tmux-fsm/weaver/core"
)

type TmuxSnapshotProvider struct{}

func (p *TmuxSnapshotProvider) TakeSnapshot(paneID string) (core.Snapshot, error) {
	cursor := TmuxGetCursorPos(paneID)
	lines := TmuxCapturePane(paneID)

	snapshot := core.TakeSnapshot(paneID, core.CursorPos{
		Row: cursor[0],
		Col: cursor[1],
	}, lines)

	return snapshot, nil
}
