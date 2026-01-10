package editor

import (
	"container/list"
	"fmt"
)

// GetAncestors returns a set of all ancestor IDs for the given node
func (dag *OperationDAG) GetAncestors(nodeID DAGNodeID) map[DAGNodeID]bool {
	ancestors := make(map[DAGNodeID]bool)
	queue := list.New()
	queue.PushBack(nodeID)

	visited := make(map[DAGNodeID]bool)
	visited[nodeID] = true

	for queue.Len() > 0 {
		element := queue.Front()
		queue.Remove(element)
		currentID := element.Value.(DAGNodeID)

		node, exists := dag.Nodes[currentID]
		if !exists {
			continue
		}

		for _, parentID := range node.Parents {
			if !visited[parentID] {
				ancestors[parentID] = true
				visited[parentID] = true
				queue.PushBack(parentID)
			}
		}
	}
	return ancestors
}

// FindLCA finds the Lowest Common Ancestor(s) between two nodes
// Note: In a DAG, there can be multiple LCAs. This returns one of them, usually the most recent.
func (dag *OperationDAG) FindLCA(a, b DAGNodeID) DAGNodeID {
	ancestorsA := dag.GetAncestors(a)
	ancestorsA[a] = true // Include self

	// BFS from b upwards to find the first node that is in ancestorsA
	queue := list.New()
	queue.PushBack(b)
	visited := make(map[DAGNodeID]bool)
	visited[b] = true

	if ancestorsA[b] {
		return b
	}

	for queue.Len() > 0 {
		element := queue.Front()
		queue.Remove(element)
		currentID := element.Value.(DAGNodeID)

		// If current is in A's ancestry, it's a common ancestor.
		// Since we traverse BFS (reverse time), the first one we see is an "LCA".
		// (Approximate definition for "Recent" common ancestor)
		if ancestorsA[currentID] {
			return currentID
		}

		node, exists := dag.Nodes[currentID]
		if !exists {
			continue
		}

		for _, parentID := range node.Parents {
			if !visited[parentID] {
				visited[parentID] = true
				queue.PushBack(parentID)
			}
		}
	}

	return "" // No common ancestor found (disjoint graphs)
}

// Diff returns the list of operations required to move from 'base' to 'target'.
// It returns the nodes that are in Target's history but NOT in Base's history.
// This is effectively "git log base..target".
// The operations are returned in topological order (dependency order).
func (dag *OperationDAG) Diff(base, target DAGNodeID) ([]*DAGNode, error) {
	if _, ok := dag.Nodes[base]; !ok {
		return nil, fmt.Errorf("base node %s not found", base)
	}
	if _, ok := dag.Nodes[target]; !ok {
		return nil, fmt.Errorf("target node %s not found", target)
	}

	baseAncestors := dag.GetAncestors(base)
	baseAncestors[base] = true

	// Collect all nodes in Target's ancestry that are NOT in Base's ancestry

	// We need topological sort.
	// Simple approach: Collect all candidates, then sort.

	candidates := make(map[DAGNodeID]*DAGNode)
	queue := list.New()
	queue.PushBack(target)
	visited := make(map[DAGNodeID]bool)
	visited[target] = true

	for queue.Len() > 0 {
		element := queue.Front()
		queue.Remove(element)
		currentID := element.Value.(DAGNodeID)

		if baseAncestors[currentID] {
			continue // Stop traversing down this branch, it's already known to base
		}

		node, _ := dag.Nodes[currentID]
		candidates[currentID] = node

		for _, parentID := range node.Parents {
			if !visited[parentID] {
				visited[parentID] = true
				queue.PushBack(parentID)
			}
		}
	}

	// Now sort candidates topologically
	// Kahn's algorithm or simpler: just reverse the BFS?
	// BFS reverse gives roughly topological but not strict.
	// Since we have the full map, we can just sort by dependency.

	result := make([]*DAGNode, 0, len(candidates))

	// Copy map to work with
	remaining := make(map[DAGNodeID]bool)
	for id := range candidates {
		remaining[id] = true
	}

	for len(remaining) > 0 {
		var nextBatch []DAGNodeID

		// Find nodes whose parents are ALL either not in 'remaining' (i.e. processed or base)
		for id := range remaining {
			node := candidates[id]
			ready := true
			for _, p := range node.Parents {
				if remaining[p] {
					ready = false
					break
				}
			}
			if ready {
				nextBatch = append(nextBatch, id)
			}
		}

		if len(nextBatch) == 0 {
			// Cycle detected or logic error, break to avoid infinite loop
			return nil, fmt.Errorf("cycle detected or topo sort error")
		}

		// Sort batch by timestamp for determinism?
		// For now just append
		for _, id := range nextBatch {
			result = append(result, candidates[id])
			delete(remaining, id)
		}
	}

	return result, nil
}
