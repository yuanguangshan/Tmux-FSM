package analysis

import (
	"rhm-go/core/change"
	"rhm-go/core/history"
)

type Conflict struct {
	NodeA    history.NodeID
	NodeB    history.NodeID
	Reason   string
	Resource string
	ModeA    change.AccessMode
	ModeB    change.AccessMode
}

type MergeResult struct {
	Conflicts []Conflict
}

func AnalyzeMerge(view history.DagView, tipA, tipB history.NodeID) MergeResult {
	nodeA := view.GetNode(tipA)
	nodeB := view.GetNode(tipB)
	if nodeA == nil || nodeB == nil {
		return MergeResult{}
	}

	semA, okA := nodeA.Op.(change.SemanticChange)
	semB, okB := nodeB.Op.(change.SemanticChange)

	// 如果无法进行语义分析，保守认为无冲突或由外层处理
	if !okA || !okB {
		return MergeResult{}
	}

	for _, fA := range semA.GetFootprints() {
		for _, fB := range semB.GetFootprints() {
			if fA.ResourceID == fB.ResourceID {
				if isMutuallyExclusive(fA.Mode, fB.Mode) {
					return MergeResult{
						Conflicts: []Conflict{{
							NodeA:    tipA,
							NodeB:    tipB,
							Reason:   "Resource Contention: " + fA.ResourceID,
							Resource: fA.ResourceID,
							ModeA:    fA.Mode,
							ModeB:    fB.Mode,
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
	if m1 == change.Exclusive || m2 == change.Exclusive {
		return true
	}
	if m1 == change.Create && m2 == change.Create {
		return true
	}
	return false
}

// ConflictSeverity 返回冲突严重性评级 (50, 80, 100)
func ConflictSeverity(c Conflict) int {
	if c.ModeA == change.Exclusive || c.ModeB == change.Exclusive {
		return 100
	}
	if c.ModeA == change.Create && c.ModeB == change.Create {
		return 80
	}
	return 50
}
