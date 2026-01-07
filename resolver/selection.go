package resolver

type SelectionMode int

const (
	SelectionNone SelectionMode = iota
	SelectionChar
	SelectionLine
	SelectionBlock
)

type Cursor struct {
	Line int
	Col  int
}

type Selection struct {
	Mode   SelectionMode
	Anchor Cursor
	Focus  Cursor
}