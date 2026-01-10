package history

import (
	"errors"
)

// FindLCA 寻找两个节点的最近公共祖先 (Lowest Common Ancestor)
// 在合并场景中，这通常被称为 Merge Base。
// 这里实现一个适用于多父节点 DAG 的 BFS/祖先遍历版本。
func (d *HistoryDAG) FindLCA(a, b NodeID) (NodeID, error) {
	if a == b {
		return a, nil
	}

	ancestorsA := d.getAllAncestors(a)

	// 从 b 开始反向搜索，第一个出现在 ancestorsA 中的节点即为 LCA
	queue := []NodeID{b}
	visited := make(map[NodeID]bool)

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if visited[curr] {
			continue
		}
		visited[curr] = true

		if ancestorsA[curr] {
			return curr, nil
		}

		for _, p := range d.GetParents(curr) {
			queue = append(queue, p)
		}
	}

	return "", errors.New("no common ancestor found")
}

func (d *HistoryDAG) getAllAncestors(id NodeID) map[NodeID]bool {
	ancestors := make(map[NodeID]bool)
	queue := []NodeID{id}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if ancestors[curr] {
			continue
		}
		ancestors[curr] = true

		for _, p := range d.GetParents(curr) {
			queue = append(queue, p)
		}
	}
	return ancestors
}
