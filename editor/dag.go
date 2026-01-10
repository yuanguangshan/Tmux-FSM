package editor

import (
	"encoding/json"
	"fmt"
	"time"
)

// DAGNodeID Unique identifier for a node in the DAG
type DAGNodeID string

// DAGNode represents a single atomic operation in the edit graph
type DAGNode struct {
	ID        DAGNodeID         `json:"id"`
	Operation ResolvedOperation `json:"operation"`
	Parents   []DAGNodeID       `json:"parents"` // Dependencies
	Timestamp int64             `json:"timestamp"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// OperationDAG represents a Directed Acyclic Graph of operations
// This is the core IR for collaborative editing and advanced history
type OperationDAG struct {
	Nodes map[DAGNodeID]*DAGNode `json:"nodes"`
	Roots []DAGNodeID            `json:"roots"`
	Tips  []DAGNodeID            `json:"tips"` // Operations with no children (latest state)
}

// NewOperationDAG creates a new empty DAG
func NewOperationDAG() *OperationDAG {
	return &OperationDAG{
		Nodes: make(map[DAGNodeID]*DAGNode),
		Roots: []DAGNodeID{},
		Tips:  []DAGNodeID{},
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
