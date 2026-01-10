package search

import (
	"crypto/sha256"
	"encoding/binary"
	"rhm-go/core/change"
	"rhm-go/core/cost"
	"rhm-go/core/narrative"
)

// State 代表搜索树中的一个节点
type State struct {
	Mutations   []change.Mutation      // 已经选定的手术路径
	Cost        cost.Cost              // 当前累积的语义成本
	Narrative   []narrative.DecisionStep // 决策记录
	Fingerprint uint64                 // 状态指纹（去重用）
}

// ComputeFingerprint 核心算法：确保状态唯一性，防止环路
func ComputeFingerprint(mutations []change.Mutation) uint64 {
	h := sha256.New()
	for _, m := range mutations {
		h.Write([]byte(m.Target))
		// 获取操作的哈希值
		h.Write([]byte(m.NewOp.Hash()))
	}
	sum := h.Sum(nil)
	return binary.BigEndian.Uint64(sum[:8])
}

// PriorityQueue 为 A* 搜索提供支持
type PriorityQueue []*State

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].Cost < pq[j].Cost }
func (pq PriorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i] }

func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*State))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
