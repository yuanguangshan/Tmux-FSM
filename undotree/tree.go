package undotree

import (
	"sort"

	"tmux-fsm/wal"
)

//
// ─────────────────────────────────────────────────────────────
//  Undo Node
// ─────────────────────────────────────────────────────────────
//

type UndoNode struct {
	Event    *wal.SemanticEvent
	Parent   *UndoNode
	Children []*UndoNode
}

// IsRoot 判断是否为虚拟根
func (n *UndoNode) IsRoot() bool {
	return n.Event == nil
}

//
// ─────────────────────────────────────────────────────────────
//  Build Undo Tree
// ─────────────────────────────────────────────────────────────
//

func BuildUndoTree(events []wal.SemanticEvent) *UndoNode {

	root := &UndoNode{} // ✅ 虚拟根
	nodes := make(map[string]*UndoNode)

	// 1️⃣ 创建节点
	for i := range events {
		e := &events[i]
		nodes[e.ID] = &UndoNode{
			Event: e,
		}
	}

	// 2️⃣ 建立父子关系（LocalParent）
	for _, n := range nodes {
		lp := n.Event.LocalParent

		if lp == "" {
			n.Parent = root
			root.Children = append(root.Children, n)
			continue
		}

		if p, ok := nodes[lp]; ok {
			n.Parent = p
			p.Children = append(p.Children, n)
		} else {
			// ✅ 父缺失 → 挂到 root（WAL 截断 / 合并时常见）
			n.Parent = root
			root.Children = append(root.Children, n)
		}
	}

	// 3️⃣ 稳定排序（按时间 + ID）
	sortTree(root)

	return root
}

func sortTree(n *UndoNode) {
	sort.Slice(n.Children, func(i, j int) bool {
		ei := n.Children[i].Event
		ej := n.Children[j].Event

		if ei.Time.Equal(ej.Time) {
			return ei.ID < ej.ID
		}
		return ei.Time.Before(ej.Time)
	})

	for _, c := range n.Children {
		sortTree(c)
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Path Utilities
// ─────────────────────────────────────────────────────────────
//

// PathToRoot 返回从 root → node 的事件路径（不含虚拟 root）
func PathToRoot(n *UndoNode) []*wal.SemanticEvent {
	var rev []*wal.SemanticEvent

	for cur := n; cur != nil && !cur.IsRoot(); cur = cur.Parent {
		rev = append(rev, cur.Event)
	}

	// reverse
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}

	return rev
}
