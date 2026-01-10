package scheduler

import (
	"container/heap"
	"rhm-go/core/analysis"
)

// ConflictItem 包装冲突并添加优先级
type ConflictItem struct {
	conflict analysis.Conflict
	priority int
}

// PriorityQueue 实现堆接口
type PriorityQueue struct {
	heap []*ConflictItem
}

func (pq PriorityQueue) Len() int { return len(pq.heap) }
func (pq PriorityQueue) Less(i, j int) bool {
	// 优先级越高越先处理
	return pq.heap[i].priority > pq.heap[j].priority
}
func (pq PriorityQueue) Swap(i, j int) {
	pq.heap[i], pq.heap[j] = pq.heap[j], pq.heap[i]
}
func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*ConflictItem)
	pq.heap = append(pq.heap, item)
}
func (pq *PriorityQueue) Pop() interface{} {
	old := pq.heap
	n := len(old)
	item := old[n-1]
	pq.heap = old[0 : n-1]
	return item
}

// ConflictScheduler 管理冲突处理顺序
type ConflictScheduler struct {
	queue *PriorityQueue
}

func NewScheduler() *ConflictScheduler {
	return &ConflictScheduler{
		queue: &PriorityQueue{heap: make([]*ConflictItem, 0)},
	}
}

func (s *ConflictScheduler) AddConflict(c analysis.Conflict) {
	priority := analysis.ConflictSeverity(c)
	heap.Push(s.queue, &ConflictItem{conflict: c, priority: priority})
}

func (s *ConflictScheduler) HasNext() bool {
	return s.queue.Len() > 0
}

func (s *ConflictScheduler) Next() analysis.Conflict {
	item := heap.Pop(s.queue).(*ConflictItem)
	return item.conflict
}
