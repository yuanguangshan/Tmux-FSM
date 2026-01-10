package analysis

import (
	"rhm-go/core/change"
	"rhm-go/core/history"
)

type Conflict struct {
	NodeA  history.NodeID
	NodeB  history.NodeID
	Reason string
}

type MergeResult struct {
	Conflicts []Conflict
}

func AnalyzeMerge(view history.DagView, tipA, tipB history.NodeID) MergeResult {
	nodeA := view.GetNode(tipA)
	nodeB := view.GetNode(tipB)
	if nodeA == nil || nodeB == nil { return MergeResult{} }

	semA, okA := nodeA.Op.(change.SemanticChange)
	semB, okB := nodeB.Op.(change.SemanticChange)

	// 如果无法进行语义分析，保守认为无冲突或由外层处理
	if !okA || !okB { return MergeResult{} }

	for _, fA := range semA.GetFootprints() {
		for _, fB := range semB.GetFootprints() {
			if fA.ResourceID == fB.ResourceID {
				if isMutuallyExclusive(fA.Mode, fB.Mode) {
					return MergeResult{
						Conflicts: []Conflict{{
							NodeA: tipA, NodeB: tipB,
							Reason: "Resource Contention: " + fA.ResourceID,
						}},
					}
				}
			}
		}
	}
	return MergeResult{}
}

func isMutuallyExclusive(m1, m2 change.AccessMode) bool {
	// 互斥矩阵实现
	if m1 == change.Exclusive || m2 == change.Exclusive { return true }
	if m1 == change.Create && m2 == change.Create { return true }
	return false
}
