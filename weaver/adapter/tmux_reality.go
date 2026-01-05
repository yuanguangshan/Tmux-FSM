package adapter

import "tmux-fsm/weaver/core"

type TmuxRealityReader struct {
	Provider *TmuxSnapshotProvider
}

func (r *TmuxRealityReader) ReadCurrent(paneID string) (core.Snapshot, error) {
	return r.Provider.TakeSnapshot(paneID)
}
