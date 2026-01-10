package editor

import (
	"encoding/json"
	"fmt"
	"time"
)

// DAGNodeID Unique identifier for a node in the DAG
type DAGNodeID string

// ConflictNode represents a blocking point in the history where automated merge failed
type ConflictNode struct {
	ID         DAGNodeID   `json:"id"`
	Parents    []DAGNodeID `json:"parents"` // The tips that are in conflict
	Conflicts  []Conflict  `json:"conflicts"`
	Timestamp  int64       `json:"timestamp"`
	Resolved   bool        `json:"resolved"`
	Resolution DAGNodeID   `json:"resolution_node,omitempty"` // The node that resolves this conflict
}

// DAGNode represents a single atomic operation in the edit graph
type DAGNode struct {
	ID        DAGNodeID         `json:"id"`
	Operation ResolvedOperation `json:"operation"`
	Parents   []DAGNodeID       `json:"parents"` // Dependencies
	Timestamp int64             `json:"timestamp"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// Custom JSON marshaling for DAGNode to handle ResolvedOperation interface
func (n *DAGNode) MarshalJSON() ([]byte, error) {
	type Alias DAGNode
	return json.Marshal(&struct {
		*Alias
		OpType OpKind `json:"op_type"`
	}{
		Alias:  (*Alias)(n),
		OpType: n.Operation.Kind(),
	})
}

func (n *DAGNode) UnmarshalJSON(data []byte) error {
	type Alias DAGNode
	aux := &struct {
		*Alias
		OpType OpKind          `json:"op_type"`
		OpRaw  json.RawMessage `json:"operation"`
	}{
		Alias: (*Alias)(n),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var op ResolvedOperation
	switch aux.OpType {
	case OpInsert:
		op = &InsertOperation{}
	case OpDelete:
		op = &DeleteOperation{}
	case OpMove:
		op = &MoveOperation{}
	case OpComposite:
		op = &CompositeOperation{}
	case OpRename:
		op = &RenameOperation{}
	default:
		return fmt.Errorf("unknown operation kind: %v", aux.OpType)
	}

	if err := json.Unmarshal(aux.OpRaw, op); err != nil {
		return err
	}
	n.Operation = op
	return nil
}

// OperationDAG represents a Directed Acyclic Graph of operations
// This is the core IR for collaborative editing and advanced history
type OperationDAG struct {
	Nodes     map[DAGNodeID]*DAGNode      `json:"nodes"`
	Conflicts map[DAGNodeID]*ConflictNode `json:"conflicts"` // Blocking conflict nodes
	Roots     []DAGNodeID                 `json:"roots"`
	Tips      []DAGNodeID                 `json:"tips"` // Operations with no children (latest state)
}

// NewOperationDAG creates a new empty DAG
func NewOperationDAG() *OperationDAG {
	return &OperationDAG{
		Nodes:     make(map[DAGNodeID]*DAGNode),
		Conflicts: make(map[DAGNodeID]*ConflictNode),
		Roots:     []DAGNodeID{},
		Tips:      []DAGNodeID{},
	}
}

// AddNode adds a new operation to the DAG
func (dag *OperationDAG) AddNode(op ResolvedOperation, parents []DAGNodeID) (*DAGNode, error) {
	// Verify parents exist
	for _, pid := range parents {
		if _, ok := dag.Nodes[pid]; !ok {
			return nil, fmt.Errorf("parent node %s not found", pid)
		}
	}

	node := &DAGNode{
		ID:        DAGNodeID(fmt.Sprintf("node_%d_%d", time.Now().UnixNano(), len(dag.Nodes))),
		Operation: op,
		Parents:   parents,
		Timestamp: time.Now().UnixNano(),
	}

	dag.Nodes[node.ID] = node

	// Update Tips
	// 1. Remove parents from Tips (they are no longer tips)
	newTips := []DAGNodeID{}
	parentSet := make(map[DAGNodeID]bool)
	for _, pid := range parents {
		parentSet[pid] = true
	}

	for _, tip := range dag.Tips {
		if !parentSet[tip] {
			newTips = append(newTips, tip)
		}
	}
	// 2. Add new node to Tips
	newTips = append(newTips, node.ID)
	dag.Tips = newTips

	// Update Roots if no parents
	if len(parents) == 0 {
		dag.Roots = append(dag.Roots, node.ID)
	}

	return node, nil
}

// Serialize serializes the DAG to JSON
func (dag *OperationDAG) Serialize() ([]byte, error) {
	return json.Marshal(dag)
}

// DeserializeDAG deserializes a DAG from JSON
func DeserializeDAG(data []byte) (*OperationDAG, error) {
	var dag OperationDAG
	if err := json.Unmarshal(data, &dag); err != nil {
		return nil, err
	}
	return &dag, nil
}
