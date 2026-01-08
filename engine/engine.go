package engine

import (
	"tmux-fsm/crdt"
	"tmux-fsm/replay"
	"tmux-fsm/wal"
)

// Engine 编辑器内核引擎接口
type Engine interface {
	// State 状态管理
	Apply(event crdt.SemanticEvent) error
	Replay(upTo crdt.EventID) replay.TextState
	Snapshot() *Snapshot

	// WAL 事件日志
	Append(event crdt.SemanticEvent) crdt.EventID
	WALSince(id crdt.EventID) []wal.SemanticEvent

	// CRDT 位置管理
	AllocatePosition(after, before *crdt.PositionID) crdt.PositionID
	ComparePosition(a, b crdt.PositionID) int

	// GC 垃圾回收
	Compact(stable crdt.EventID)

	// 同步
	KnownHeads() map[crdt.ActorID]crdt.EventID
	Integrate(events []wal.SemanticEvent) error
}

// Snapshot 快照
type Snapshot struct {
	At    crdt.EventID    `json:"at"`
	State replay.TextState `json:"state"`
}

// HeadlessEngine 无头引擎实现
type HeadlessEngine struct {
	store      *crdt.EventStore
	snapshots  map[crdt.EventID]*Snapshot
	currentState replay.TextState
}

// NewEngine 创建新的引擎
func NewEngine() *HeadlessEngine {
	return &HeadlessEngine{
		store:     crdt.NewEventStore(),
		snapshots: make(map[crdt.EventID]*Snapshot),
		currentState: replay.TextState{
			Text:   "",
			Cursor: 0,
		},
	}
}

// Apply 应用事件
func (e *HeadlessEngine) Apply(event crdt.SemanticEvent) error {
	e.store.Merge(event)
	
	// 更新当前状态
	sortedEvents := e.store.TopoSort()
	e.currentState = replay.Replay(
		replay.TextState{}, 
		sortedEvents, 
		nil, // 不使用过滤器
	)
	
	return nil
}

// Replay 重放至指定事件
func (e *HeadlessEngine) Replay(upTo crdt.EventID) replay.TextState {
	allEvents := e.store.TopoSort()
	
	// 找到 upTo 事件的索引
	var eventsToReplay []crdt.SemanticEvent
	for _, event := range allEvents {
		eventsToReplay = append(eventsToReplay, event)
		if event.ID == upTo {
			break
		}
	}
	
	return replay.Replay(
		replay.TextState{}, 
		eventsToReplay, 
		nil,
	)
}

// Snapshot 创建快照
func (e *HeadlessEngine) Snapshot() *Snapshot {
	snapshot := &Snapshot{
		At:    "", // 需要设置为最新的事件ID
		State: e.currentState,
	}
	
	// 获取最新的事件ID
	allEvents := e.store.TopoSort()
	if len(allEvents) > 0 {
		snapshot.At = allEvents[len(allEvents)-1].ID
	}
	
	e.snapshots[snapshot.At] = snapshot
	return snapshot
}

// Append 添加事件到日志
func (e *HeadlessEngine) Append(event crdt.SemanticEvent) crdt.EventID {
	e.store.Merge(event)
	return event.ID
}

// WALSince 获取指定事件之后的日志
func (e *HeadlessEngine) WALSince(id crdt.EventID) []wal.SemanticEvent {
	allEvents := e.store.TopoSort()
	
	var result []wal.SemanticEvent
	found := false
	for _, event := range allEvents {
		if !found && event.ID == id {
			found = true
			continue
		}
		if found {
			// 转换 crdt.SemanticEvent 到 wal.SemanticEvent
			walEvent := wal.SemanticEvent{
				ID:            string(event.ID),
				CausalParents: []string{},
				LocalParent:   string(event.LocalParent),
				Time:          event.Time,
				Actor:         string(event.Actor),
				Fact:          event.Fact,
			}
			
			// 填充 CausalParents
			for _, parent := range event.CausalParents {
				walEvent.CausalParents = append(walEvent.CausalParents, string(parent))
			}
			
			result = append(result, walEvent)
		}
	}
	
	return result
}

// AllocatePosition 分配新位置
func (e *HeadlessEngine) AllocatePosition(after, before *crdt.PositionID) crdt.PositionID {
	actor := "default" // 这里应该从上下文获取实际的 actor
	if after != nil {
		actor = string(after.Actor)
	} else if before != nil {
		actor = string(before.Actor)
	}
	
	return crdt.AllocateBetween(after, before, crdt.ActorID(actor))
}

// ComparePosition 比较位置
func (e *HeadlessEngine) ComparePosition(a, b crdt.PositionID) int {
	return crdt.ComparePos(a, b)
}

// Compact 压缩日志
func (e *HeadlessEngine) Compact(stable crdt.EventID) {
	// 实现压缩逻辑
	// 这里简化处理，实际实现需要更复杂的逻辑
}

// KnownHeads 获取已知头部
func (e *HeadlessEngine) KnownHeads() map[crdt.ActorID]crdt.EventID {
	heads := make(map[crdt.ActorID]crdt.EventID)
	
	allEvents := e.store.TopoSort()
	
	// 按参与者分组，找到每个参与者的最新事件
	for _, event := range allEvents {
		if current, exists := heads[event.Actor]; !exists || event.ID > current {
			heads[event.Actor] = event.ID
		}
	}
	
	return heads
}

// Integrate 集成外部事件
func (e *HeadlessEngine) Integrate(events []wal.SemanticEvent) error {
	for _, walEvent := range events {
		// 转换 wal.SemanticEvent 到 crdt.SemanticEvent
		crdtEvent := crdt.SemanticEvent{
			ID:            crdt.EventID(walEvent.ID),
			Actor:         crdt.ActorID(walEvent.Actor),
			CausalParents: []crdt.EventID{},
			LocalParent:   crdt.EventID(walEvent.LocalParent),
			Time:          walEvent.Time,
			Fact:          walEvent.Fact,
		}
		
		// 转换 CausalParents
		for _, parent := range walEvent.CausalParents {
			crdtEvent.CausalParents = append(crdtEvent.CausalParents, crdt.EventID(parent))
		}
		
		e.store.Merge(crdtEvent)
	}
	
	return nil
}