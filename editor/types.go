package editor

import (
	"errors"
)

// IntentKind 定义意图类型
type IntentKind int

const (
	IntentMove IntentKind = iota
	IntentDelete
	IntentChange
	IntentYank
)

// TargetKind 定义目标类型
type TargetKind int

const (
	TargetChar TargetKind = iota
	TargetWord
	TargetLine
	TargetTextObject
)

// Target 定义目标结构
type Target struct {
	Kind      TargetKind
	Direction string
	Scope     string
	Value     string
}

// MotionKind 定义动作类型
type MotionKind int

const (
	MotionCharForward MotionKind = iota
	MotionCharBackward
	MotionWordForward
	MotionWordBackward
	MotionLineStart
	MotionLineEnd
)

// OperatorKind 定义操作符类型
type OperatorKind int

const (
	OpNone OperatorKind = iota
	OpDelete
	OpYank
	OpChange
)

// Cursor 定义光标位置
type Cursor struct {
	Row int
	Col int
}

// MotionRange 定义动作范围
type MotionRange struct {
	Start Cursor
	End   Cursor
}

// Intent 定义意图结构
type Intent struct {
	Kind  IntentKind
	Count int
	Target Target
}

// ResolvedOperationKind 定义解析后操作的类型
type ResolvedOperationKind int

const (
	OpInsert ResolvedOperationKind = iota
	OpDeleteResolvedOp = OpInsert + 1
	OpMoveResolvedOp = OpInsert + 2
)

// BufferID 代表缓冲区ID
type BufferID string

// WindowID 代表窗口ID
type WindowID string

// TextRange 定义文本范围
type TextRange struct {
	Start Cursor
	End   Cursor
}

// ResolvedOperation 表示解析后的操作
// 这是用于 . repeat 和 undo 的核心数据结构
// 所有语义解析应在 resolve 阶段完成，replay 阶段只执行预定义的操作
type ResolvedOperation struct {
	Kind ResolvedOperationKind

	BufferID BufferID
	WindowID WindowID

	// 执行位置（执行前就已确定）
	Anchor Cursor

	// Insert 专用
	Text string

	// Delete/Move 专用（半开区间）
	Range *TextRange

	// Delete 时记录被删除的文本，用于 undo
	DeletedText string

	// 是否需要先删除范围内容（用于替换操作）
	DeleteBeforeInsert bool
}

// Resolver 负责解析意图到具体操作
type Resolver struct {
	engine           *CursorEngine
	textObjectCalc   *ConcreteTextObjectCalculator
}

// NewResolver 创建新的解析器
func NewResolver(engine *CursorEngine) *Resolver {
	return &Resolver{
		engine:         engine,
		textObjectCalc: NewConcreteTextObjectCalculator(engine.Buffer),
	}
}

// Resolve 解析意图
func (r *Resolver) Resolve(intent Intent) (*ResolvedOperation, error) {
	start := *r.engine.Cursor

	switch intent.Kind {
	case IntentMove:
		return r.resolveMove(&intent, start)
	case IntentDelete, IntentChange, IntentYank:
		return r.resolveOperator(&intent, start)
	}
	return nil, errors.New("unknown intent type")
}

// resolveMove 解析移动意图
func (r *Resolver) resolveMove(intent *Intent, start Cursor) (*ResolvedOperation, error) {
	if intent.Target.Kind == TargetTextObject {
		// 处理文本对象移动
		obj, err := ParseTextObject(intent.Target.Value)
		if err != nil {
			return nil, err
		}

		textRange, err := r.textObjectCalc.CalculateRange(*obj, start)
		if err != nil {
			return nil, err
		}

		return &ResolvedOperation{
			Operator: OpNone,
			Motion:   MotionKind(intent.Target.Kind),
			Count:    intent.Count,
			From:     start,
			To:       textRange.End, // 移动到文本对象的结束位置
			Range:    textRange,
		}, nil
	} else {
		// 处理普通移动
		motion := &Motion{
			Kind:  MotionKind(intent.Target.Kind),
			Count: intent.Count,
		}

		mr, err := r.engine.ComputeMotion(motion)
		if err != nil {
			return nil, err
		}

		// 虚拟计算终点（不改 cursor）
		end := start
		end.Row += mr.DeltaRow
		end.Col += mr.DeltaCol
		end.Row, end.Col = r.clampCursor(end.Row, end.Col)

		return &ResolvedOperation{
			Operator: OpNone,
			Motion:   motion.Kind,
			Count:    motion.Count,
			From:     start,
			To:       end,
			Range:    nil, // 移动通常不产生范围
		}, nil
	}
}

// clampCursor 限制光标位置
func (r *Resolver) clampCursor(row, col int) (int, int) {
	if r.engine.Buffer == nil {
		return row, col
	}

	row = clamp(row, 0, r.engine.Buffer.LineCount()-1)

	maxCol := 0
	if row >= 0 && row < r.engine.Buffer.LineCount() {
		maxCol = r.engine.Buffer.LineLength(row)
		if maxCol > 0 {
			maxCol-- // Length 是实际长度，所以最大索引是 Length-1
		}
	}
	col = clamp(col, 0, maxCol)

	return row, col
}

// resolveOperator 解析操作意图
func (r *Resolver) resolveOperator(intent *Intent, start Cursor) (*ResolvedOperation, error) {
	var opKind OperatorKind = OpNone
	switch intent.Kind {
	case IntentDelete:
		opKind = OpDeleteResolved
	case IntentChange:
		opKind = OpChange
	case IntentYank:
		opKind = OpYank
	}

	var rng *MotionRange

	if intent.Target.Kind == TargetTextObject {
		// 处理文本对象操作
		obj, err := ParseTextObject(intent.Target.Value)
		if err != nil {
			return nil, err
		}

		textRange, err := r.textObjectCalc.CalculateRange(*obj, start)
		if err != nil {
			return nil, err
		}

		rng = textRange
	} else {
		// 处理普通运动操作
		motion := &Motion{
			Kind:  MotionKind(intent.Target.Kind),
			Count: intent.Count,
		}

		mr, err := r.engine.ComputeMotion(motion)
		if err != nil {
			return nil, err
		}

		// 虚拟计算终点（不改 cursor）
		end := start
		end.Row += mr.DeltaRow
		end.Col += mr.DeltaCol
		end.Row, end.Col = r.clampCursor(end.Row, end.Col)

		rng = resolveRange(opKind, start, end, motion.Kind)
	}

	return &ResolvedOperation{
		Operator: opKind,
		Motion:   MotionKind(intent.Target.Kind),
		Count:    intent.Count,
		From:     start,
		To:       start, // 操作后光标位置可能不同，这里先设置为起始位置
		Range:    rng,
	}, nil
}

// resolveRange 计算操作范围
func resolveRange(op OperatorKind, from Cursor, to Cursor, motion MotionKind) *MotionRange {
	switch motion {
	case MotionWordForward:
		switch op {
		case OpDeleteResolved, OpYank:
			return &MotionRange{Start: from, End: to}
		case OpChange:
			// Vim: cw 不包含 word 后的空白
			adjusted := to
			adjusted.Col-- // 简化版
			return &MotionRange{Start: from, End: adjusted}
		}
	}

	// fallback
	return &MotionRange{Start: from, End: to}
}

// ResolveDelete 将 Motion 转换为 ResolvedOperation
// 这是将高级意图转换为可重复操作的关键步骤
func ResolveDelete(cursor Cursor, motion Motion, buffer Buffer) (ResolvedOperation, ResolvedOperation, error) {
	// 计算运动结束位置
	start := cursor
	end := start // 简化实现，实际需要根据 motion 计算 end 位置

	// 根据不同的运动类型计算结束位置
	switch motion.Kind {
	case MotionCharForward:
		end.Col += motion.Count
	case MotionCharBackward:
		end.Col -= motion.Count
	case MotionWordForward:
		// 简化实现：向前移动几个单词
		end.Col += motion.Count * 5 // 假设每个单词平均5个字符
	case MotionWordBackward:
		end.Col -= motion.Count * 5
	case MotionLineStart:
		end.Col = 0
	case MotionLineEnd:
		// 需要获取当前行的长度
		if buffer != nil {
			end.Col = buffer.LineLength(start.Row)
		}
	}

	// 确保 end 位置有效
	if buffer != nil {
		end.Row, end.Col = clamp(end.Row, end.Col, buffer.LineCount(), buffer.LineLength(end.Row))
	}

	// 标准化区间（确保 start 在前，end 在后）
	if end.Row < start.Row || (end.Row == start.Row && end.Col < start.Col) {
		start, end = end, start
	}

	// 获取被删除的文本
	deletedText := ""
	if buffer != nil {
		// 尝试获取范围内的文本
		if sb, ok := buffer.(*SimpleBuffer); ok {
			text, err := sb.GetTextInRange(start, end)
			if err == nil {
				deletedText = text
			}
		}
	}

	// 创建删除操作
	deleteOp := ResolvedOperation{
		Kind:        OpDeleteResolved,
		BufferID:    "", // 实际应用中应设置适当的 BufferID
		WindowID:    "", // 实际应用中应设置适当的 WindowID
		Anchor:      start,
		Range:       &TextRange{Start: start, End: end},
		DeletedText: deletedText,
	}

	// 创建对应的插入操作（作为反向操作，用于 undo）
	insertOp := ResolvedOperation{
		Kind:     OpInsert,
		BufferID: deleteOp.BufferID,
		WindowID: deleteOp.WindowID,
		Anchor:   start,
		Text:     deletedText,
	}

	return deleteOp, insertOp, nil
}

// ResolveInsert 将插入意图转换为 ResolvedOperation
// 返回插入操作和其反向操作（删除操作）
func ResolveInsert(cursor Cursor, text string) (ResolvedOperation, ResolvedOperation) {
	insertOp := ResolvedOperation{
		Kind:     OpInsert,
		BufferID: "", // 实际应用中应设置适当的 BufferID
		WindowID: "", // 实际应用中应设置适当的 WindowID
		Anchor:   cursor,
		Text:     text,
	}

	// 创建对应的删除操作（作为反向操作，用于 undo）
	deleteOp := ResolvedOperation{
		Kind:        OpDeleteResolved,
		BufferID:    insertOp.BufferID,
		WindowID:    insertOp.WindowID,
		Anchor:      cursor,
		Range:       &TextRange{Start: cursor, End: Cursor{Row: cursor.Row, Col: cursor.Col + len(text)}},
		DeletedText: text,
	}

	return insertOp, deleteOp
}

// ResolveChange 将变更意图转换为 ResolvedOperation
// 变更是先删除指定范围的内容，然后在原位置插入新内容
func ResolveChange(cursor Cursor, rangeToDelete TextRange, newText string, buffer Buffer) (ResolvedOperation, ResolvedOperation, error) {
	// 获取被删除的文本
	deletedText := ""
	if buffer != nil {
		if sb, ok := buffer.(*SimpleBuffer); ok {
			text, err := sb.GetTextInRange(rangeToDelete.Start, rangeToDelete.End)
			if err == nil {
				deletedText = text
			}
		}
	}

	// 创建变更操作（删除后插入）
	changeOp := ResolvedOperation{
		Kind:               OpInsert, // Change 本质上是替换，我们用带删除标记的插入表示
		BufferID:          "",       // 实际应用中应设置适当的 BufferID
		WindowID:          "",       // 实际应用中应设置适当的 WindowID
		Anchor:            rangeToDelete.Start, // 插入位置是删除范围的起点
		Text:              newText,
		DeleteBeforeInsert: true,     // 标记需要先删除范围内容
		Range:             &rangeToDelete, // 要删除的范围
		DeletedText:       deletedText, // 被删除的文本内容
	}

	// 创建对应的反向操作（撤销变更：删除新插入的文本，恢复原来的内容）
	undoOp := ResolvedOperation{
		Kind:               OpInsert,
		BufferID:          changeOp.BufferID,
		WindowID:          changeOp.WindowID,
		Anchor:            rangeToDelete.Start,
		Text:              deletedText, // 恢复原来的文本
		DeleteBeforeInsert: true,       // 标记需要先删除新插入的内容
		Range:             &TextRange{
			Start: rangeToDelete.Start,
			End:   Cursor{Row: rangeToDelete.Start.Row, Col: rangeToDelete.Start.Col + len(newText)},
		},
		DeletedText: newText, // 新插入的文本（现在要被删除）
	}

	return changeOp, undoOp, nil
}

// TextObjectKind 定义文本对象类型
type TextObjectKind int

const (
	TextObjectInnerParen TextObjectKind = iota
	TextObjectAroundParen
	TextObjectInnerQuote
	TextObjectAroundQuote
	TextObjectInnerBracket
	TextObjectAroundBracket
	TextObjectInnerBrace
	TextObjectAroundBrace
)

// TextObject 定义文本对象
type TextObject struct {
	Kind TextObjectKind
}

// ResolveInnerParen 解析内部括号文本对象
func ResolveInnerParen(cursor Cursor, buffer Buffer) (*TextRange, error) {
	if buffer == nil {
		return nil, errors.New("buffer is nil")
	}

	// 从当前位置向前查找匹配的左括号
	leftParenPos, err := findMatchingBackward(cursor, '(', ')', buffer)
	if err != nil {
		return nil, err
	}

	// 从左括号位置向后查找匹配的右括号
	rightParenPos, err := findMatchingForward(*leftParenPos, '(', ')', buffer)
	if err != nil {
		return nil, err
	}

	// 返回括号内的范围（不包括括号本身）
	result := &TextRange{
		Start: Cursor{Row: leftParenPos.Row, Col: leftParenPos.Col + 1},
		End:   Cursor{Row: rightParenPos.Row, Col: rightParenPos.Col},
	}

	return result, nil
}

// ResolveAroundParen 解析周围括号文本对象
func ResolveAroundParen(cursor Cursor, buffer Buffer) (*TextRange, error) {
	if buffer == nil {
		return nil, errors.New("buffer is nil")
	}

	// 从当前位置向前查找匹配的左括号
	leftParenPos, err := findMatchingBackward(cursor, '(', ')', buffer)
	if err != nil {
		return nil, err
	}

	// 从左括号位置向后查找匹配的右括号
	rightParenPos, err := findMatchingForward(*leftParenPos, '(', ')', buffer)
	if err != nil {
		return nil, err
	}

	// 返回括号及其中内容的范围（包括括号本身）
	result := &TextRange{
		Start: *leftParenPos,
		End:   Cursor{Row: rightParenPos.Row, Col: rightParenPos.Col + 1}, // 包含右括号
	}

	return result, nil
}

// ResolveInnerQuote 解析内部引号文本对象
func ResolveInnerQuote(cursor Cursor, quoteChar rune, buffer Buffer) (*TextRange, error) {
	if buffer == nil {
		return nil, errors.New("buffer is nil")
	}

	// 从当前位置向前查找匹配的左引号
	leftQuotePos, err := findCharBackward(cursor, quoteChar, buffer)
	if err != nil {
		return nil, err
	}

	// 从左引号位置向后查找匹配的右引号
	rightQuotePos, err := findCharForward(*leftQuotePos, quoteChar, buffer)
	if err != nil {
		return nil, err
	}

	// 返回引号内的范围（不包括引号本身）
	result := &TextRange{
		Start: Cursor{Row: leftQuotePos.Row, Col: leftQuotePos.Col + 1},
		End:   *rightQuotePos,
	}

	return result, nil
}

// findMatchingBackward 向后查找匹配的括号
func findMatchingBackward(cursor Cursor, open, close rune, buffer Buffer) (*Cursor, error) {
	// 从当前位置开始向前搜索
	row, col := cursor.Row, cursor.Col

	// 首先尝试当前位置是否是右括号
	if row >= 0 && row < buffer.LineCount() && col >= 0 {
		lineLen := buffer.LineLength(row)
		if col < lineLen {
			char := buffer.RuneAt(row, col)
			if char == close {
				// 如果当前位置是右括号，直接从这里开始匹配
				return findMatchingPair(row, col-1, open, close, true, buffer)
			}
		}
	}

	// 否则从当前位置前面开始搜索
	return findMatchingPair(row, col-1, open, close, true, buffer)
}

// findMatchingForward 向前查找匹配的括号
func findMatchingForward(cursor Cursor, open, close rune, buffer Buffer) (*Cursor, error) {
	row, col := cursor.Row, cursor.Col
	return findMatchingPair(row, col+1, open, close, false, buffer)
}

// findMatchingPair 查找匹配的括号对
func findMatchingPair(startRow, startCol int, open, close rune, backward bool, buffer Buffer) (*Cursor, error) {
	if buffer == nil {
		return nil, errors.New("buffer is nil")
	}

	count := 0
	row, col := startRow, startCol

	for {
		// 检查边界
		if row < 0 || row >= buffer.LineCount() {
			break
		}

		lineLen := buffer.LineLength(row)
		if backward {
			if col < 0 {
				row--
				if row < 0 {
					break
				}
				col = buffer.LineLength(row) - 1
				if col < 0 {
					col = 0
				}
				continue
			}
		} else {
			if col >= lineLen {
				row++
				if row >= buffer.LineCount() {
					break
				}
				col = 0
				continue
			}
		}

		char := buffer.RuneAt(row, col)

		if char == open {
			count++
		} else if char == close {
			count--
			if count == -1 {
				// 找到了匹配的右括号
				pos := Cursor{Row: row, Col: col}
				return &pos, nil
			}
		}

		if backward {
			col--
		} else {
			col++
		}
	}

	// 如果是向后查找且没找到，尝试从当前位置开始向前查找右括号
	if backward {
		row, col = startRow, startCol
		count = 0

		for {
			// 检查边界
			if row < 0 || row >= buffer.LineCount() {
				break
			}

			lineLen := buffer.LineLength(row)
			if col < 0 {
				row--
				if row < 0 {
					break
				}
				col = buffer.LineLength(row) - 1
				if col < 0 {
					col = 0
				}
				continue
			}

			if col >= lineLen {
				col = lineLen - 1
				if col < 0 {
					col = 0
				}
			}

			char := buffer.RuneAt(row, col)

			if char == close {
				// 找到右括号，开始匹配
				return findMatchingPair(row, col-1, open, close, true, buffer)
			}

			col--
		}
	}

	return nil, errors.New("matching bracket not found")
}

// findCharBackward 向后查找字符
func findCharBackward(cursor Cursor, target rune, buffer Buffer) (*Cursor, error) {
	row, col := cursor.Row, cursor.Col

	for {
		if row < 0 || row >= buffer.LineCount() {
			break
		}

		if col < 0 {
			row--
			if row < 0 {
				break
			}
			col = buffer.LineLength(row) - 1
			if col < 0 {
				col = 0
			}
			continue
		}

		char := buffer.RuneAt(row, col)
		if char == target {
			pos := Cursor{Row: row, Col: col}
			return &pos, nil
		}

		col--
	}

	return nil, errors.New("character not found")
}

// findCharForward 向前查找字符
func findCharForward(cursor Cursor, target rune, buffer Buffer) (*Cursor, error) {
	row, col := cursor.Row, cursor.Col

	for {
		if row < 0 || row >= buffer.LineCount() {
			break
		}

		lineLen := buffer.LineLength(row)
		if col >= lineLen {
			row++
			if row >= buffer.LineCount() {
				break
			}
			col = 0
			continue
		}

		char := buffer.RuneAt(row, col)
		if char == target {
			pos := Cursor{Row: row, Col: col}
			return &pos, nil
		}

		col++
	}

	return nil, errors.New("character not found")
}

// clamp 限制光标位置
func clamp(row, col, maxRow, maxCol int) (int, int) {
	if row < 0 {
		row = 0
	}
	if row >= maxRow {
		row = maxRow - 1
		if row < 0 {
			row = 0
		}
	}

	if col < 0 {
		col = 0
	}
	if col >= maxCol {
		col = maxCol - 1
		if col < 0 {
			col = 0
		}
	}

	return row, col
}