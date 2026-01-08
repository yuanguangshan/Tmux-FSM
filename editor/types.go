package editor

import (
	"fmt"
)

// BufferID 代表缓冲区ID
type BufferID string

// WindowID 代表窗口ID
type WindowID string

// Cursor 定义光标位置
type Cursor struct {
	Row int
	Col int
}

func (c Cursor) String() string {
	return fmt.Sprintf("%d:%d", c.Row, c.Col)
}

// LessThan 比较两个光标位置
func (c Cursor) LessThan(other Cursor) bool {
	if c.Row < other.Row {
		return true
	}
	if c.Row == other.Row {
		return c.Col < other.Col
	}
	return false
}

// TextRange 定义文本范围（半开区间 [Start, End)）
type TextRange struct {
	Start Cursor
	End   Cursor
}

// ResolvedOperationKind 定义解析后操作的类型
type ResolvedOperationKind int

const (
	OpInsert ResolvedOperationKind = iota
	OpDelete
	OpMove
)

// ResolvedOperation 表示解析后的物理操作
// 这是用于 . repeat 和 undo 的核心数据结构
// 所有语义解析应在 resolve 阶段完成，replay 阶段只执行预定义的操作
type ResolvedOperation struct {
	Kind     ResolvedOperationKind `json:"kind"`
	BufferID BufferID              `json:"buffer_id"`
	WindowID WindowID              `json:"window_id"`

	// 执行位置
	Anchor Cursor `json:"anchor"`

	// Insert 专用：插入的文本
	Text string `json:"text,omitempty"`

	// Delete/Move 专用：作用范围
	Range *TextRange `json:"range,omitempty"`

	// Delete 时记录被删除的文本，用于 undo
	DeletedText string `json:"deleted_text,omitempty"`

	// 是否需要先删除范围内容（用于替换操作）
	DeleteBeforeInsert bool `json:"delete_before_insert,omitempty"`
}

// Selection 表示一个选区
type Selection struct {
	Start Cursor `json:"start"`
	End   Cursor `json:"end"`
}

// Buffer 接口定义
type Buffer interface {
	InsertAt(pos Cursor, text string) error
	DeleteRange(start, end Cursor) (deleted string, err error)
	Line(row int) string
	LineCount() int
	LineLength(row int) int
	RuneAt(row, col int) rune
}

// BufferStore 接口定义
type BufferStore interface {
	Get(id BufferID) Buffer
}

// Window 结构定义
type Window struct {
	ID     WindowID
	Cursor Cursor
}

// WindowStore 接口定义
type WindowStore interface {
	Get(id WindowID) *Window
}

// SelectionStore 接口定义
type SelectionStore interface {
	Get(buffer BufferID) []Selection
	Set(buffer BufferID, selections []Selection)
}