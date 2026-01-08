package crdt

import (
	"sort"
	"time"
	"tmux-fsm/semantic"
)

// EventID 事件ID类型
type EventID string

// ActorID 参与者ID类型
type ActorID string

// SemanticEvent 修正后的语义事件结构
type SemanticEvent struct {
	// 全局唯一、幂等基础
	ID     EventID `json:"id"`
	Actor  ActorID `json:"actor"`
	Time   time.Time `json:"time"`

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
	Fact semantic.BaseFact `json:"fact"`
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

// Merge 合并事件（网络/WAL/Sync）
func (s *EventStore) Merge(e SemanticEvent) {
	if _, ok := s.Events[e.ID]; ok {
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