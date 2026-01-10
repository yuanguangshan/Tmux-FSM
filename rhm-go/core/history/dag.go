package history

import "rhm-go/core/change"

type NodeID string

type Node struct {
	ID      NodeID
	Op      change.ReversibleChange
	Parents []NodeID
}

// DagView 允许对真实历史和沙盒历史进行统一读取
type DagView interface {
	GetNode(id NodeID) *Node
	GetParents(id NodeID) []NodeID
}

type HistoryDAG struct {
	Nodes map[NodeID]*Node
	Roots []NodeID
}

func NewHistoryDAG() *HistoryDAG {
	return &HistoryDAG{Nodes: make(map[NodeID]*Node)}
}

func (d *HistoryDAG) AddOp(id NodeID, op change.ReversibleChange, parents []NodeID) {
	d.Nodes[id] = &Node{ID: id, Op: op, Parents: parents}
	if len(parents) == 0 {
		d.Roots = append(d.Roots, id)
	}
}

func (d *HistoryDAG) GetNode(id NodeID) *Node { return d.Nodes[id] }
func (d *HistoryDAG) GetParents(id NodeID) []NodeID {
	if n, ok := d.Nodes[id]; ok {
		return n.Parents
	}
	return nil
}
