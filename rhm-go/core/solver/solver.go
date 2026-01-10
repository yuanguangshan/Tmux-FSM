package solver

import (
	"container/heap"
	"fmt"
	"rhm-go/core/analysis"
	"rhm-go/core/change"
	"rhm-go/core/cost"
	"rhm-go/core/history"
	"rhm-go/core/narrative"
	"rhm-go/core/rewrite"
	"rhm-go/core/search"
)

type ResolutionPlan struct {
	Mutations []change.Mutation
	Resolved  bool
	Narrative narrative.Narrative
}

// Solve finds the optimal resolution path between two conflict tips in a DAG.
//
// [RHM Solver Contract]:
// 1. Search space is generated ONLY by reversible operations defined in core/change.
// 2. Solver NEVER mutates history irreversibly; it explores parallel ephemeral universes.
// 3. Any optimal solution corresponds to a coherent, auditable responsibility narrative.
// 4. Deterministic: Given identical HistoryDAG and CostModel, RHM produces bit-identical solutions.
func Solve(dag *history.HistoryDAG, tipA, tipB history.NodeID) ResolutionPlan {
	costModel := cost.DefaultModel{}
	ctx := cost.Context{}

	pq := &search.PriorityQueue{}
	heap.Init(pq)
	closedSet := make(map[uint64]bool)

	// 1. Initial Analysis
	initialRes := analysis.AnalyzeMerge(dag, tipA, tipB)
	if len(initialRes.Conflicts) == 0 {
		return ResolutionPlan{Resolved: true, Narrative: narrative.Narrative{Summary: "No conflict detected."}}
	}

	h0 := cost.Cost(len(initialRes.Conflicts)) * cost.Tweak
	heap.Push(pq, &search.State{Mutations: []change.Mutation{}, Cost: 0, Heuristic: h0})

	// 2. A* Loop
	for pq.Len() > 0 {
		current := heap.Pop(pq).(*search.State)

		if closedSet[current.Fingerprint] {
			continue
		}
		closedSet[current.Fingerprint] = true

		// Fork Sandbox
		forkPoint := tipA // Simplified LCA
		if parents := dag.GetParents(tipA); len(parents) > 0 {
			forkPoint = parents[0]
		}
		sandbox := rewrite.RewriteBatch(dag, forkPoint, current.Mutations)

		// Check Goal
		res := analysis.AnalyzeMerge(sandbox, tipA, tipB)
		if len(res.Conflicts) == 0 {
			return ResolutionPlan{
				Mutations: current.Mutations,
				Resolved:  true,
				Narrative: narrative.Narrative{
					Summary:   "Optimal Timeline Found",
					Steps:     current.Narrative,
					TotalCost: int(current.Cost),
				},
			}
		}

		// Expand: Explore all involved nodes in the first conflict to find the lowest cost resolution
		if len(res.Conflicts) > 0 {
			conflict := res.Conflicts[0]
			// Try both sides of the conflict
			involved := []history.NodeID{conflict.NodeA, conflict.NodeB}

			for _, offenderID := range involved {
				candidates := generateCandidates(sandbox, offenderID)

				// Score Candidates for Narrative Comparison
				type Scored struct {
					Mut  change.Mutation
					Cost cost.Cost
				}
				scoredCands := []Scored{}
				for _, m := range candidates {
					scoredCands = append(scoredCands, Scored{m, costModel.Calculate(m, ctx)})
				}

				for _, sc := range scoredCands {
					newCost := current.Cost + sc.Cost
					if newCost >= cost.Infinite {
						continue
					}

					// Build Narrative Step
					rejected := []narrative.RejectedAlternative{}
					// In a real expanded search, we'd compare across all branches,
					// here we record alternatives at the local decision point.
					for _, other := range scoredCands {
						if other.Mut.String() != sc.Mut.String() {
							reason := "Alternative path"
							if other.Cost > sc.Cost {
								reason = "Higher semantic cost"
							}
							rejected = append(rejected, narrative.RejectedAlternative{
								Description: other.Mut.String(),
								Cost:        int(other.Cost),
								Reason:      reason,
							})
						}
					}

					step := narrative.DecisionStep{
						ProblemContext: fmt.Sprintf("Conflict between %s and %s (handling %s)",
							conflict.NodeA, conflict.NodeB, offenderID),
						Decision:     sc.Mut.String(),
						DecisionCost: int(sc.Cost),
						Rejected:     rejected,
					}

					newState := &search.State{
						Mutations: append(append([]change.Mutation{}, current.Mutations...), sc.Mut),
						Cost:      newCost,
						Heuristic: 0,
						Narrative: append(append([]narrative.DecisionStep{}, current.Narrative...), step),
					}
					newState.Fingerprint = search.ComputeFingerprint(newState.Mutations)
					heap.Push(pq, newState)
				}
			}
		}
	}
	return ResolutionPlan{Resolved: false}
}

func generateCandidates(view history.DagView, id history.NodeID) []change.Mutation {
	node := view.GetNode(id)
	if node == nil {
		return nil
	}
	muts := []change.Mutation{}

	if noop := node.Op.ToNoOp(); noop != nil {
		muts = append(muts, change.Mutation{Type: change.ReplaceOp, Target: string(id), NewOp: noop})
	}
	if down := node.Op.Downgrade(); down != nil {
		muts = append(muts, change.Mutation{Type: change.ReplaceOp, Target: string(id), NewOp: down})
	}
	return muts
}
