package rewrite

import (
	"rhm-go/core/change"
	"rhm-go/core/history"
)

// EphemeralDAG 是内存中的平行宇宙
type EphemeralDAG struct {
	Base    history.DagView
	Overlay map[history.NodeID]*history.Node
	Head    history.NodeID
}

func NewEphemeralDAG(base history.DagView, startPoint history.NodeID) *EphemeralDAG {
	return &EphemeralDAG{
		Base:    base,
		Overlay: make(map[history.NodeID]*history.Node),
		Head:    startPoint,
	}
}

func (e *EphemeralDAG) GetNode(id history.NodeID) *history.Node {
	if n, ok := e.Overlay[id]; ok {
		return n
	}
	return e.Base.GetNode(id)
}

func (e *EphemeralDAG) GetParents(id history.NodeID) []history.NodeID {
	if n := e.GetNode(id); n != nil {
		return n.Parents
	}
	return nil
}

// RewriteBatch 在沙盒中批量执行手术
func RewriteBatch(base history.DagView, startPoint history.NodeID, mutations []change.Mutation) *EphemeralDAG {
	sandbox := NewEphemeralDAG(base, startPoint)
	for _, m := range mutations {
		if m.Type == change.ReplaceOp {
			orig := sandbox.GetNode(history.NodeID(m.Target))
			if orig != nil {
				newNode := *orig
				newNode.Op = m.NewOp
				sandbox.Overlay[history.NodeID(m.Target)] = &newNode
			}
		}
	}
	// 在完整版中，此处需执行 Causal Replay
	return sandbox
}
