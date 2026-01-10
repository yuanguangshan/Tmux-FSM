package analysis

import (
	"rhm-go/core/history"
	"strings"
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
	// Simple analysis: if tipA and tipB refer to the same target in their operations, they conflict.
	// In the real system, this would use Footprint/Operation Algebra.

	nodeA := view.GetNode(tipA)
	nodeB := view.GetNode(tipB)

	if nodeA == nil || nodeB == nil {
		return MergeResult{}
	}

	descA := nodeA.Op.Describe()
	descB := nodeB.Op.Describe()

	// Conflict detection logic for demo:
	// If one is Delete and other is Edit/Write on same entity.
	// Here we simulate by checking if descriptions "Delete" and "Edit" appear.
	if (strings.Contains(descA, "Delete") && strings.Contains(descB, "Edit")) ||
		(strings.Contains(descB, "Delete") && strings.Contains(descA, "Edit")) {
		return MergeResult{
			Conflicts: []Conflict{
				{NodeA: tipA, NodeB: tipB, Reason: "Edit vs Delete Conflict"},
			},
		}
	}

	return MergeResult{}
}
