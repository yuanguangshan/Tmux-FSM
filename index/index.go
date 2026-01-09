package index

import (
	"fmt"
	"sort"
	"time"
	"tmux-fsm/crdt"
	"tmux-fsm/semantic"
)

// FactType 事实类型
type FactType string

const (
	FactTypeInsert  FactType = "insert"
	FactTypeDelete  FactType = "delete"
	FactTypeMove    FactType = "move"
	FactTypeReplace FactType = "replace"
)

// EventIndex 事件索引
type EventIndex struct {
	ByActor    map[crdt.ActorID][]crdt.EventID
	ByType     map[FactType][]crdt.EventID
	ByPosition PositionIntervalTree
	ByTime     TimeBTree
	ByContent  map[string][]crdt.EventID // 按内容索引
}

// PositionIntervalTree 位置区间树（简化实现）
type PositionIntervalTree struct {
	// 这里使用一个简单的映射作为示例
	// 实际实现可能需要更复杂的数据结构
	intervals map[string][]crdt.EventID
}

// TimeBTree 时间B树（简化实现）
type TimeBTree struct {
	// 简化为时间戳到事件ID的映射
	timeline map[int64][]crdt.EventID
}

// NewEventIndex 创建新的事件索引
func NewEventIndex() *EventIndex {
	return &EventIndex{
		ByActor:    make(map[crdt.ActorID][]crdt.EventID),
		ByType:     make(map[FactType][]crdt.EventID),
		ByPosition: PositionIntervalTree{intervals: make(map[string][]crdt.EventID)},
		ByTime:     TimeBTree{timeline: make(map[int64][]crdt.EventID)},
		ByContent:  make(map[string][]crdt.EventID),
	}
}

// BuildIndex 构建索引
func BuildIndex(events []crdt.SemanticEvent) *EventIndex {
	index := NewEventIndex()

	for _, event := range events {
		// 按参与者索引
		index.ByActor[event.Actor] = append(index.ByActor[event.Actor], event.ID)

		// 按类型索引
		factType := getFactType(event.Fact)
		index.ByType[factType] = append(index.ByType[factType], event.ID)

		// 按时间索引
		index.ByTime.timeline[event.Time.Unix()] = append(index.ByTime.timeline[event.Time.Unix()], event.ID)

		// 按位置索引
		positionKey := getPositionKey(event.Fact)
		index.ByPosition.intervals[positionKey] = append(index.ByPosition.intervals[positionKey], event.ID)

		// 按内容索引
		contentKey := getContentKey(event.Fact)
		if contentKey != "" {
			index.ByContent[contentKey] = append(index.ByContent[contentKey], event.ID)
		}
	}

	return index
}

// getFactType 获取事实类型
func getFactType(fact semantic.Fact) FactType {
	switch fact.Kind() {
	case semantic.FactInsert:
		return FactTypeInsert
	case semantic.FactDelete:
		return FactTypeDelete
	case semantic.FactMove:
		return FactTypeMove
	case semantic.FactReplace:
		return FactTypeReplace
	default:
		return FactType("unknown")
	}
}

// getPositionKey 获取位置键
func getPositionKey(fact semantic.Fact) string {
	anchor := fact.Anchor()
	return string(anchor.PaneID) + ":" + fmt.Sprintf("%d", anchor.Line) + ":" + fmt.Sprintf("%d", anchor.Col)
}

// getContentKey 获取内容键
func getContentKey(fact semantic.Fact) string {
	return fact.Text()
}

// QueryByActor 按参与者查询
func (idx *EventIndex) QueryByActor(actor crdt.ActorID) []crdt.EventID {
	events, exists := idx.ByActor[actor]
	if !exists {
		return []crdt.EventID{}
	}
	return events
}

// QueryByType 按类型查询
func (idx *EventIndex) QueryByType(ft FactType) []crdt.EventID {
	events, exists := idx.ByType[ft]
	if !exists {
		return []crdt.EventID{}
	}
	return events
}

// QueryByTimeRange 按时间范围查询
func (idx *EventIndex) QueryByTimeRange(start, end time.Time) []crdt.EventID {
	var result []crdt.EventID

	startUnix := start.Unix()
	endUnix := end.Unix()

	for timestamp, events := range idx.ByTime.timeline {
		if timestamp >= startUnix && timestamp <= endUnix {
			result = append(result, events...)
		}
	}

	return result
}

// QueryByPositionRange 按位置范围查询
func (idx *EventIndex) QueryByPositionRange(startPos, endPos string) []crdt.EventID {
	var result []crdt.EventID

	// 简化实现：查找在指定位置范围内的事件
	for posKey, events := range idx.ByPosition.intervals {
		if posKey >= startPos && posKey <= endPos {
			result = append(result, events...)
		}
	}

	return result
}

// QueryByContent 按内容查询
func (idx *EventIndex) QueryByContent(content string) []crdt.EventID {
	events, exists := idx.ByContent[content]
	if !exists {
		return []crdt.EventID{}
	}
	return events
}

// QueryAIChanges 查询 AI 的更改
func (idx *EventIndex) QueryAIChanges(aiActorPrefix string) []crdt.EventID {
	var result []crdt.EventID

	for actor, events := range idx.ByActor {
		actorStr := string(actor)
		if len(actorStr) >= len(aiActorPrefix) && actorStr[:len(aiActorPrefix)] == aiActorPrefix {
			result = append(result, events...)
		}
	}

	return result
}

// QueryEvolutionHistory 查询某段文本的演化历史
func (idx *EventIndex) QueryEvolutionHistory(content string) []crdt.EventID {
	// 首先按内容查找
	contentEvents := idx.QueryByContent(content)

	// 然后可能需要扩展到相关的插入/删除事件
	var result []crdt.EventID
	result = append(result, contentEvents...)

	// 这里可以添加更多逻辑来查找相关的事件
	// 例如，查找在同一位置附近的操作等

	return result
}

// QueryWhoDeleted 查询谁删除了特定内容
func (idx *EventIndex) QueryWhoDeleted(content string) []crdt.ActorID {
	var actors []crdt.ActorID

	// 查找删除操作
	deleteEvents := idx.QueryByType(FactTypeDelete)

	for _, eventID := range deleteEvents {
		// 这里需要一个事件ID到事件的映射
		// 由于简化实现，我们跳过这一步
		// 在实际实现中，需要从存储中检索事件并检查其内容
	}

	return actors
}

// SortEventsByID 对事件ID进行排序
func SortEventsByID(events []crdt.EventID) []crdt.EventID {
	sorted := make([]crdt.EventID, len(events))
	copy(sorted, events)

	sort.Slice(sorted, func(i, j int) bool {
		return string(sorted[i]) < string(sorted[j])
	})

	return sorted
}

// SortEventsByTime 对事件按时间排序
func SortEventsByTime(events []crdt.SemanticEvent) []crdt.SemanticEvent {
	sorted := make([]crdt.SemanticEvent, len(events))
	copy(sorted, events)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Time.Before(sorted[j].Time)
	})

	return sorted
}

// GetTimeline 获取时间线
func (idx *EventIndex) GetTimeline() []int64 {
	var timestamps []int64
	for timestamp := range idx.ByTime.timeline {
		timestamps = append(timestamps, timestamp)
	}

	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] < timestamps[j]
	})

	return timestamps
}

// GetActors 获取所有参与者
func (idx *EventIndex) GetActors() []crdt.ActorID {
	var actors []crdt.ActorID
	for actor := range idx.ByActor {
		actors = append(actors, actor)
	}

	// 排序以确保一致性
	sort.Slice(actors, func(i, j int) bool {
		return string(actors[i]) < string(actors[j])
	})

	return actors
}
