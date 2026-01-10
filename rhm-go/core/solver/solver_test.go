package solver

import (
	"testing"
	"rhm-go/core/history"
	"rhm-go/store"
)

func TestSolveWithFootprintAnalysis(t *testing.T) {
	// Create a real HistoryDAG with conflicting operations
	dag := history.NewHistoryDAG()

	// Add two conflicting operations: Delete and Edit on the same resource
	deleteOp := store.FileSystemOp{Kind: "Delete", Arg: "file.txt"}
	editOp := store.FileSystemOp{Kind: "Edit", Arg: "file.txt"}

	tipA := history.NodeID("tipA")
	tipB := history.NodeID("tipB")

	dag.AddOp(tipA, deleteOp, []history.NodeID{})
	dag.AddOp(tipB, editOp, []history.NodeID{})

	// Call the solver to resolve the conflict
	result := Solve(dag, tipA, tipB)

	// The solver should find a resolution (either by downgrading or neutralizing)
	if !result.Resolved {
		t.Errorf("Expected solver to find a resolution, but it didn't")
	}

	// The result should contain mutations
	if len(result.Mutations) == 0 {
		t.Log("No mutations were needed to resolve the conflict")
	} else {
		t.Logf("Found %d mutations to resolve the conflict", len(result.Mutations))
		for i, mut := range result.Mutations {
			t.Logf("Mutation %d: %s", i, mut.String())
		}
	}
}

func TestSolveWithNoConflict(t *testing.T) {
	// Create a real HistoryDAG with non-conflicting operations
	dag := history.NewHistoryDAG()

	// Add two non-conflicting operations: operations on different resources
	editOp1 := store.FileSystemOp{Kind: "Edit", Arg: "file1.txt"}
	editOp2 := store.FileSystemOp{Kind: "Edit", Arg: "file2.txt"}

	tipA := history.NodeID("tipA")
	tipB := history.NodeID("tipB")

	dag.AddOp(tipA, editOp1, []history.NodeID{})
	dag.AddOp(tipB, editOp2, []history.NodeID{})

	// Call the solver - there should be no conflict
	result := Solve(dag, tipA, tipB)

	// Since there's no conflict, the result should be resolved with no mutations
	if !result.Resolved {
		t.Errorf("Expected solver to recognize no conflict exists, but it didn't")
	}

	// No mutations should be needed
	if len(result.Mutations) != 0 {
		t.Errorf("Expected 0 mutations for non-conflicting operations, got %d", len(result.Mutations))
	}
}

func TestSolveWithCreateVsCreateConflict(t *testing.T) {
	// Create a real HistoryDAG with Create vs Create conflict on the same resource
	dag := history.NewHistoryDAG()

	// Add two Create operations on the same resource - this should conflict
	createOp1 := store.FileSystemOp{Kind: "Create", Arg: "newfile.txt"}
	createOp2 := store.FileSystemOp{Kind: "Create", Arg: "newfile.txt"}

	tipA := history.NodeID("tipA")
	tipB := history.NodeID("tipB")

	dag.AddOp(tipA, createOp1, []history.NodeID{})
	dag.AddOp(tipB, createOp2, []history.NodeID{})

	// Call the solver to resolve the conflict
	result := Solve(dag, tipA, tipB)

	// The solver should find a resolution
	if !result.Resolved {
		t.Errorf("Expected solver to find a resolution for Create vs Create conflict, but it didn't")
	}

	t.Logf("Found resolution for Create vs Create conflict with %d mutations", len(result.Mutations))
}