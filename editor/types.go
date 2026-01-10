package editor

import (
	"fmt"
)

// BufferID 代表缓冲区ID
type BufferID string

// WindowID 代表窗口ID
type WindowID string

// OperationID 代表操作唯一ID
type OperationID string

// SymbolID 代表语义符号唯一ID
type SymbolID string

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

// Advance 在当前位置基础上推进（简单按列推进，不考虑换行，用于 Footprint 计算）
func (c Cursor) Advance(cols int) Cursor {
	return Cursor{Row: c.Row, Col: c.Col + cols}
}

// TextRange 定义文本范围（半开区间 [Start, End)）
type TextRange struct {
	Start Cursor `json:"start"`
	End   Cursor `json:"end"`
}

// MotionRange 定义 motion 操作的范围
// 用于 text object 和 motion 计算
type MotionRange struct {
	Start Cursor
	End   Cursor
}

// ResolvedOperationKind 定义解析后操作的类型
type OpKind int

const (
	OpInsert OpKind = iota
	OpDelete
	OpMove
	OpMoveCursor
	OpComposite
	OpRename
)

// MoveCursorOperation 光标移动操作
type MoveCursorOperation struct {
	ID       OperationID `json:"id"`
	WindowID WindowID    `json:"window_id"`
	To       Cursor      `json:"to"`
}

func (op *MoveCursorOperation) OpID() OperationID { return op.ID }
func (op *MoveCursorOperation) Kind() OpKind      { return OpMoveCursor }
func (op *MoveCursorOperation) Apply(buf Buffer) error {
	// Buffer context is not enough for MoveCursor, handled in engine.go
	return nil
}
func (op *MoveCursorOperation) Inverse() (ResolvedOperation, error) {
	// Note: True inverse requires knowing previous cursor position.
	// For now, this is a placeholder.
	return nil, fmt.Errorf("MoveCursor inverse requires context")
}
func (op *MoveCursorOperation) Footprint() Footprint {
	return Footprint{
		Effects: []EffectKind{EffectRead}, // Touching window state
	}
}

// EffectKind 定义操作对 Footprint 的影响类型
type EffectKind int

const (
	EffectRead EffectKind = iota
	EffectWrite
	EffectDelete
	EffectRename
	EffectCreate
)

// SymbolRef 代表对语义符号的引用
type SymbolRef struct {
	ID   SymbolID   `json:"id"`
	Kind SymbolKind `json:"kind"`
}

// SymbolKind 代表语义符号类型
type SymbolKind int

const (
	SymbolFunction SymbolKind = iota
	SymbolVariable
	SymbolType
)

// Footprint 代表操作触碰的事实集合
type Footprint struct {
	Buffers []BufferID   `json:"buffers"`
	Ranges  []TextRange  `json:"ranges"`
	Symbols []SymbolRef  `json:"symbols"`
	Effects []EffectKind `json:"effects"`
}

// ResolvedOperation 表示解析后的物理操作接口
// 它是可逆、可组合、可判冲突的代数对象
type ResolvedOperation interface {
	OpID() OperationID
	Kind() OpKind

	Apply(buf Buffer) error
	Inverse() (ResolvedOperation, error)
	Footprint() Footprint
}

// Concrete Operations

// InsertOperation 插入操作
type InsertOperation struct {
	ID     OperationID `json:"id"`
	Buffer BufferID    `json:"buffer_id"`
	At     Cursor      `json:"at"`
	Text   string      `json:"text"`
}

func (op *InsertOperation) OpID() OperationID { return op.ID }
func (op *InsertOperation) Kind() OpKind      { return OpInsert }
func (op *InsertOperation) Apply(buf Buffer) error {
	return buf.InsertAt(op.At, op.Text)
}
func (op *InsertOperation) Inverse() (ResolvedOperation, error) {
	return &DeleteOperation{
		ID:     OperationID(fmt.Sprintf("inv_%s", op.ID)),
		Buffer: op.Buffer,
		Range: TextRange{
			Start: op.At,
			End:   op.At.Advance(len(op.Text)),
		},
		DeletedText: op.Text,
	}, nil
}
func (op *InsertOperation) Footprint() Footprint {
	return Footprint{
		Buffers: []BufferID{op.Buffer},
		Ranges:  []TextRange{{Start: op.At, End: op.At}},
		Effects: []EffectKind{EffectWrite},
	}
}

// DeleteOperation 删除操作
type DeleteOperation struct {
	ID          OperationID `json:"id"`
	Buffer      BufferID    `json:"buffer_id"`
	Range       TextRange   `json:"range"`
	DeletedText string      `json:"deleted_text"`
}

func (op *DeleteOperation) OpID() OperationID { return op.ID }
func (op *DeleteOperation) Kind() OpKind      { return OpDelete }
func (op *DeleteOperation) Apply(buf Buffer) error {
	deleted, err := buf.DeleteRange(op.Range.Start, op.Range.End)
	if err != nil {
		return err
	}
	// 校验被删除的文本是否匹配（可选，增加鲁棒性）
	if op.DeletedText != "" && deleted != op.DeletedText {
		// 这里可以返回警告或错误，但目前为了兼容性先不严格限制
	}
	return nil
}
func (op *DeleteOperation) Inverse() (ResolvedOperation, error) {
	return &InsertOperation{
		ID:     OperationID(fmt.Sprintf("inv_%s", op.ID)),
		Buffer: op.Buffer,
		At:     op.Range.Start,
		Text:   op.DeletedText,
	}, nil
}
func (op *DeleteOperation) Footprint() Footprint {
	return Footprint{
		Buffers: []BufferID{op.Buffer},
		Ranges:  []TextRange{op.Range},
		Effects: []EffectKind{EffectDelete},
	}
}

// MoveOperation 移动操作（语义上是删除+插入的复合体）
type MoveOperation struct {
	ID     OperationID `json:"id"`
	Buffer BufferID    `json:"buffer_id"`
	From   TextRange   `json:"from"`
	To     Cursor      `json:"to"`
	Text   string      `json:"text"`
}

func (op *MoveOperation) OpID() OperationID { return op.ID }
func (op *MoveOperation) Kind() OpKind      { return OpMove }
func (op *MoveOperation) Apply(buf Buffer) error {
	_, err := buf.DeleteRange(op.From.Start, op.From.End)
	if err != nil {
		return err
	}
	return buf.InsertAt(op.To, op.Text)
}
func (op *MoveOperation) Inverse() (ResolvedOperation, error) {
	return &MoveOperation{
		ID:     OperationID(fmt.Sprintf("inv_%s", op.ID)),
		Buffer: op.Buffer,
		From: TextRange{
			Start: op.To,
			End:   op.To.Advance(len(op.Text)),
		},
		To:   op.From.Start,
		Text: op.Text,
	}, nil
}
func (op *MoveOperation) Footprint() Footprint {
	return Footprint{
		Buffers: []BufferID{op.Buffer},
		Ranges:  []TextRange{op.From},
		Effects: []EffectKind{EffectDelete, EffectWrite},
	}
}

// RenameOperation 重命名操作
type RenameOperation struct {
	ID      OperationID `json:"id"`
	Buffer  BufferID    `json:"buffer_id"`
	Symbol  SymbolRef   `json:"symbol"`
	OldName string      `json:"old_name"`
	NewName string      `json:"new_name"`
}

func (op *RenameOperation) OpID() OperationID { return op.ID }
func (op *RenameOperation) Kind() OpKind      { return OpRename }
func (op *RenameOperation) Apply(buf Buffer) error {
	// Rename is a semantic operation, usually handled by projection/LSP
	return nil
}
func (op *RenameOperation) Inverse() (ResolvedOperation, error) {
	return &RenameOperation{
		ID:      OperationID(fmt.Sprintf("inv_%s", op.ID)),
		Buffer:  op.Buffer,
		Symbol:  op.Symbol,
		OldName: op.NewName,
		NewName: op.OldName,
	}, nil
}
func (op *RenameOperation) Footprint() Footprint {
	return Footprint{
		Buffers: []BufferID{op.Buffer},
		Symbols: []SymbolRef{op.Symbol},
		Effects: []EffectKind{EffectRename},
	}
}

// CompositeOperation 复合操作
type CompositeOperation struct {
	ID       OperationID         `json:"id"`
	Children []ResolvedOperation `json:"children"`
}

func (op *CompositeOperation) OpID() OperationID { return op.ID }
func (op *CompositeOperation) Kind() OpKind      { return OpComposite }
func (op *CompositeOperation) Apply(buf Buffer) error {
	for _, child := range op.Children {
		if err := child.Apply(buf); err != nil {
			return err
		}
	}
	return nil
}
func (op *CompositeOperation) Inverse() (ResolvedOperation, error) {
	inv := make([]ResolvedOperation, 0, len(op.Children))
	for i := len(op.Children) - 1; i >= 0; i-- {
		childInv, err := op.Children[i].Inverse()
		if err != nil {
			return nil, err
		}
		inv = append(inv, childInv)
	}
	return &CompositeOperation{
		ID:       OperationID(fmt.Sprintf("inv_%s", op.ID)),
		Children: inv,
	}, nil
}
func (op *CompositeOperation) Footprint() Footprint {
	fp := Footprint{
		Buffers: []BufferID{},
		Ranges:  []TextRange{},
		Symbols: []SymbolRef{},
		Effects: []EffectKind{},
	}
	for _, child := range op.Children {
		childFP := child.Footprint()
		fp.Buffers = append(fp.Buffers, childFP.Buffers...)
		fp.Ranges = append(fp.Ranges, childFP.Ranges...)
		fp.Symbols = append(fp.Symbols, childFP.Symbols...)
		fp.Effects = append(fp.Effects, childFP.Effects...)
	}
	return fp
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
