package undotree

import (
	"time"
	"tmux-fsm/semantic"
	"tmux-fsm/wal"
)

// UndoNode 撤销树节点
type UndoNode struct {
	Event    wal.SemanticEvent
	Parent   *UndoNode
	Children []*UndoNode
}

// BuildUndoTree 构建撤销树
func BuildUndoTree(events []wal.SemanticEvent) *UndoNode {
	nodes := make(map[string]*UndoNode)

	// 创建所有节点
	for _, e := range events {
		nodes[e.ID] = &UndoNode{Event: e}
	}

	var root *UndoNode

	// 建立父子关系
	for _, n := range nodes {
		if n.Event.ParentID == "" {
			root = n
			continue
		}
		p := nodes[n.Event.ParentID]
		n.Parent = p
		p.Children = append(p.Children, n)
	}

	return root
}

// Checkout 检出特定节点的状态
func Checkout(node *UndoNode, initial TextState, decideFn func(semantic.Fact) []Transaction) (TextState, error) {
	var path []wal.SemanticEvent
	current := node
	for current != nil {
		path = append([]wal.SemanticEvent{current.Event}, path...) // prepend
		current = current.Parent
	}

	state := initial
	for _, e := range path {
		// 将事件转换为事务并应用
		fact := e.Fact
		// 注意：这里需要根据实际的 Fact 类型进行转换
		// 这里简化处理
		_ = fact
	}

	return state, nil
}

// TextState 文本状态（简化版）
type TextState struct {
	Text   string
	Cursor int
}

// Transaction 事务接口（简化版）
type Transaction interface {
	Apply() error
}