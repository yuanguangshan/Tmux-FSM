package crdt

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"time"
	"tmux-fsm/semantic"
)

// EventID 事件ID类型
type EventID string

// ActorID 参与者ID类型
type ActorID string

// PositionID CRDT 位置ID
type PositionID struct {
	Path  []uint32 `json:"path"`
	Actor ActorID  `json:"actor"`
	Epoch int      `json:"epoch"` // 每次分叉/reset +1
}

// SemanticEvent 修正后的语义事件结构
type SemanticEvent struct {
	// 全局唯一、幂等基础
	ID    EventID   `json:"id"`
	Actor ActorID   `json:"actor"`
	Time  time.Time `json:"time"`

	// Version control for event integrity
	Version int `json:"version"` // Event version for tracking changes

	// 因果一致性（CRDT 用）
	CausalParents []EventID `json:"causal_parents"`
	// 含义：本事件在语义上依赖的事件集合
	// ✅ 用于拓扑排序 / 合并
	// ✅ 永远不用于 Undo

	// 本地历史（Undo 用）
	LocalParent EventID `json:"local_parent"`
	// 含义：本 actor 本地编辑历史中的上一个事件
	// ✅ 只在本地有意义
	// ✅ 不同步、不合并

	// 不可变语义
	Fact semantic.Fact `json:"fact"`
}

// ComparePos 比较两个位置
func ComparePos(a, b PositionID) int {
	min := len(a.Path)
	if len(b.Path) < min {
		min = len(b.Path)
	}

	for i := 0; i < min; i++ {
		if a.Path[i] < b.Path[i] {
			return -1
		}
		if a.Path[i] > b.Path[i] {
			return 1
		}
	}
	if len(a.Path) != len(b.Path) {
		if len(a.Path) < len(b.Path) {
			return -1
		}
		return 1
	}
	if a.Actor < b.Actor {
		return -1
	}
	if a.Actor > b.Actor {
		return 1
	}
	if a.Epoch < b.Epoch {
		return -1
	}
	if a.Epoch > b.Epoch {
		return 1
	}
	return 0
}

// AllocateBetween 在两个位置之间分配新位置
func AllocateBetween(a, b *PositionID, actor ActorID) PositionID {
	const Base = uint32(1 << 31)

	var path []uint32
	i := 0

	for {
		var left uint32 = 0
		var right uint32 = Base

		if a != nil && i < len(a.Path) {
			left = a.Path[i]
		}
		if b != nil && i < len(b.Path) {
			right = b.Path[i]
		}

		if right-left > 1 {
			mid := left + (right-left)/2
			path = append(path, mid)
			break
		}

		path = append(path, left)
		i++
	}

	return PositionID{
		Path:  path,
		Actor: actor,
		Epoch: 0, // 可能需要根据实际情况设置
	}
}

// EventStore 事件存储
type EventStore struct {
	Events map[EventID]SemanticEvent
}

// NewEventStore 创建新的事件存储
func NewEventStore() *EventStore {
	return &EventStore{
		Events: make(map[EventID]SemanticEvent),
	}
}

// GenerateStableEventID generates a stable, unique event ID based on content
func GenerateStableEventID(actor ActorID, timestamp time.Time, fact semantic.Fact) EventID {
	// Create a stable ID based on actor, timestamp, and fact content
	// This ensures that identical events get the same ID, maintaining consistency
	content := fmt.Sprintf("%s_%d_%s_%d", actor, timestamp.UnixNano(), fact.Text(), fact.Kind())
	hash := sha256.Sum256([]byte(content))
	return EventID(fmt.Sprintf("%x", hash[:16])) // Use first 16 bytes for shorter ID
}

// CreateSemanticEvent creates a new semantic event with proper versioning and timestamps
func CreateSemanticEvent(actor ActorID, fact semantic.Fact, causalParents []EventID, localParent EventID) SemanticEvent {
	timestamp := time.Now()
	version := 1 // Start with version 1 for new events

	return SemanticEvent{
		ID:            GenerateStableEventID(actor, timestamp, fact),
		Actor:         actor,
		Time:          timestamp,
		Version:       version,
		CausalParents: causalParents,
		LocalParent:   localParent,
		Fact:          fact,
	}
}

// Merge 合并事件（网络/WAL/Sync）
func (s *EventStore) Merge(e SemanticEvent) {
	if existing, ok := s.Events[e.ID]; ok {
		// Check if this is a newer version of the same event
		if e.Version > existing.Version {
			// Update with newer version
			s.Events[e.ID] = e
		}
		return // 幂等
	}
	s.Events[e.ID] = e
}

// TopoSort 拓扑排序（因果顺序）
func (s *EventStore) TopoSort() []SemanticEvent {
	return TopoSortByCausality(s.Events)
}

// TopoSortByCausality 按因果关系拓扑排序
func TopoSortByCausality(events map[EventID]SemanticEvent) []SemanticEvent {
	inDegree := make(map[EventID]int)
	graph := make(map[EventID][]EventID)

	// 初始化
	for id := range events {
		inDegree[id] = 0
	}

	// 构建因果图
	for _, e := range events {
		for _, p := range e.CausalParents {
			if _, ok := events[p]; ok {
				graph[p] = append(graph[p], e.ID)
				inDegree[e.ID]++
			}
		}
	}

	// 入度为 0 的队列
	var queue []EventID
	for id, d := range inDegree {
		if d == 0 {
			queue = append(queue, id)
		}
	}

	// 稳定排序（可选：EventID）
	sort.Slice(queue, func(i, j int) bool {
		return queue[i] < queue[j]
	})

	var result []SemanticEvent

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		result = append(result, events[id])

		for _, next := range graph[id] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	// 检测环（理论上不该出现）
	if len(result) != len(events) {
		panic("causal cycle detected")
	}

	return result
}

// LocalHistory 获取本地历史（参与者投影）
func LocalHistory(events map[EventID]SemanticEvent, me ActorID) []SemanticEvent {
	var out []SemanticEvent
	for _, e := range events {
		if e.Actor == me {
			out = append(out, e)
		}
	}
	return BuildLocalChain(out)
}

// BuildLocalChain 构建本地链
func BuildLocalChain(events []SemanticEvent) []SemanticEvent {
	// 按 LocalParent 链组织
	parentMap := make(map[EventID]EventID)
	eventMap := make(map[EventID]SemanticEvent)

	for _, e := range events {
		eventMap[e.ID] = e
		parentMap[e.ID] = e.LocalParent
	}

	// 找到根节点
	var roots []EventID
	for id, parent := range parentMap {
		if parent == "" {
			roots = append(roots, id)
		}
	}

	// 按链排序
	var result []SemanticEvent
	visited := make(map[EventID]bool)

	var dfs func(EventID)
	dfs = func(id EventID) {
		if visited[id] {
			return
		}
		visited[id] = true
		result = append(result, eventMap[id])

		// 找到所有子节点
		for childID, parentID := range parentMap {
			if parentID == id {
				dfs(childID)
			}
		}
	}

	for _, root := range roots {
		dfs(root)
	}

	return result
}

// UndoFilter 创建撤销过滤器
func UndoFilter(me ActorID, undoPoint EventID, events map[EventID]SemanticEvent) func(SemanticEvent) bool {
	disabled := make(map[EventID]bool)

	// 从撤销点向上追踪，标记需要禁用的事件
	current := undoPoint
	for current != "" {
		disabled[current] = true

		// 找到当前事件
		event, exists := events[current]
		if !exists {
			break
		}

		// 移动到父事件
		current = event.LocalParent
	}

	return func(e SemanticEvent) bool {
		if e.Actor != me {
			return true
		}
		// 如果事件在撤销点之后，则不执行
		return !disabled[e.ID]
	}
}
