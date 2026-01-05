package adapter

import "tmux-fsm/weaver/core"

// RealityReader 读取当前世界状态（用于一致性验证）
type RealityReader interface {
	ReadCurrent(paneID string) (core.Snapshot, error)
}
