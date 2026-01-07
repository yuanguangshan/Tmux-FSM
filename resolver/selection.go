package resolver

// SelectionMode 选择模式
type SelectionMode int

const (
	SelectionChar SelectionMode = iota
	SelectionLine
	SelectionBlock
)