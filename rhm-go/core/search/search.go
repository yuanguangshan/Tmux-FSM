package search

import (
	"crypto/sha256"
	"encoding/binary"
	"rhm-go/core/change"
	"rhm-go/core/cost"
	"rhm-go/core/narrative"
)

type State struct {
	Mutations   []change.Mutation
	Cost        cost.Cost
	Heuristic   cost.Cost
	Narrative   []narrative.DecisionStep
	Fingerprint uint64
}

type PriorityQueue []*State

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
	return (pq[i].Cost + pq[i].Heuristic) < (pq[j].Cost + pq[j].Heuristic)
}
func (pq PriorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i] }

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*State)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func ComputeFingerprint(mutations []change.Mutation) uint64 {
	h := sha256.New()
	for _, m := range mutations {
		h.Write([]byte(m.Target))
		h.Write([]byte(m.NewOp.Hash()))
	}
	sum := h.Sum(nil)
	return binary.BigEndian.Uint64(sum[:8])
}
