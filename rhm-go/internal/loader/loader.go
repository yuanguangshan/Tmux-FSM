package loader

import (
	"rhm-go/core/history"
	"rhm-go/store"
)

func LoadDemoScenario() (*history.HistoryDAG, history.NodeID, history.NodeID) {
	dag := history.NewHistoryDAG()

	// Root
	dag.AddOp("root", store.FileSystemOp{Kind: "Create", Arg: "README.md"}, []history.NodeID{})

	// Branch A: Edit(README.md)
	dag.AddOp("nodeA", store.FileSystemOp{Kind: "Edit", Arg: "README.md"}, []history.NodeID{"root"})

	// Branch B: Delete(README.md)
	dag.AddOp("nodeB", store.FileSystemOp{Kind: "Delete", Arg: "README.md"}, []history.NodeID{"root"})

	return dag, "nodeA", "nodeB"
}
