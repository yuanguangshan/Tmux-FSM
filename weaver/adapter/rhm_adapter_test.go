package adapter

import (
	"rhm-go/core/change"
	"rhm-go/core/history"
	"testing"
	"tmux-fsm/editor"
)

type mockOp struct {
	id   editor.OperationID
	kind editor.OpKind
}

func (m *mockOp) OpID() editor.OperationID                   { return m.id }
func (m *mockOp) Kind() editor.OpKind                        { return m.kind }
func (m *mockOp) Apply(buf editor.Buffer) error              { return nil }
func (m *mockOp) Inverse() (editor.ResolvedOperation, error) { return nil, nil }
func (m *mockOp) Footprint() editor.Footprint                { return editor.Footprint{} }

func TestRHMAdapter_MapToDAG(t *testing.T) {
	adapter := NewRHMAdapter()

	ops := []editor.ResolvedOperation{
		&mockOp{id: "root", kind: editor.OpInsert},
		&mockOp{id: "nodeA", kind: editor.OpInsert},
		&mockOp{id: "nodeB", kind: editor.OpDelete},
	}

	dependencies := map[editor.OperationID][]editor.OperationID{
		"nodeA": {"root"},
		"nodeB": {"root"},
	}

	dag := adapter.MapToDAG(ops, dependencies)

	if len(dag.Nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(dag.Nodes))
	}

	nodeA := dag.GetNode("nodeA")
	if nodeA == nil || len(nodeA.Parents) != 1 || nodeA.Parents[0] != "root" {
		t.Errorf("NodeA mapping failed")
	}
}

func TestRHMAdapter_Solve(t *testing.T) {
	adapter := NewRHMAdapter()

	dag := history.NewHistoryDAG()

	// Root
	dag.AddOp("root", &mockOpWrapper{desc: "Root"}, []history.NodeID{})

	// 为了触发演示场景中的冲突（Edit vs Delete）
	// analysis 逻辑是字符串包含 "Edit" 和 "Delete"
	dag.AddOp("nodeA", &mockOpWrapper{desc: "Edit:README.md"}, []history.NodeID{"root"})
	dag.AddOp("nodeB", &mockOpWrapper{desc: "Delete:README.md"}, []history.NodeID{"root"})

	plan := adapter.Solve(dag, "nodeA", "nodeB")

	if !plan.Resolved {
		t.Errorf("Expected conflict to be resolved")
	}

	if plan.Narrative.TotalCost != 50 {
		t.Errorf("Expected optimal cost 50, got %d", plan.Narrative.TotalCost)
	}
}

type mockOpWrapper struct {
	desc string
}

func (m *mockOpWrapper) Describe() string { return m.desc }
func (m *mockOpWrapper) Hash() string     { return m.desc }
func (m *mockOpWrapper) ToNoOp() change.ReversibleChange {
	return &mockOpWrapper{desc: "NoOp(Neutralized)"}
}
func (m *mockOpWrapper) Downgrade() change.ReversibleChange {
	if m.desc == "Delete:README.md" {
		return &mockOpWrapper{desc: "Move(Trash/README.md)"}
	}
	return nil
}
