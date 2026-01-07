package main

import "errors"

// MotionKind 定义移动方向类型
type MotionKind int

const (
	MotionLeft MotionKind = iota
	MotionRight
	MotionUp
	MotionDown
	MotionWordForward
	MotionWordBackward
	MotionLineEnd
)

// Motion 结构体定义移动动作
type Motion struct {
	Kind  MotionKind
	Count int
}

// Line 表示一行
type Line struct {
	Length int
}

// Buffer 接口定义缓冲区
type Buffer interface {
	LineCount() int
	LineLength(row int) int
	RuneAt(row, col int) rune
	DeleteRange(r MotionRange) error
}

// MotionRange 表示一个运动范围
type MotionRange struct {
	Start Cursor
	End   Cursor // Vim 语义：不含 End
}

// MotionResult 表示移动结果
type MotionResult struct {
	DeltaRow int
	DeltaCol int

	Range *MotionRange
}

// CharClass 定义字符类别
type CharClass int

const (
	ClassWhitespace CharClass = iota
	ClassWord       // 字母 + 数字 + _
	ClassPunct      // 其他
)

// motionHandler 定义运动处理器类型
type motionHandler func(engine *CursorEngine, motion *Motion) (*MotionResult, error)

// motionTable 定义运动表
var motionTable = map[MotionKind]motionHandler{
	MotionLeft:        simpleVector(0, -1),
	MotionRight:       simpleVector(0, 1),
	MotionUp:          simpleVector(-1, 0),
	MotionDown:        simpleVector(1, 0),
	MotionWordForward: wordForward,
}

// ConcreteBuffer 是 Buffer 接口的具体实现
type ConcreteBuffer struct {
	Lines []Line
	Content [][]rune  // 每行的实际内容
}

func (cb *ConcreteBuffer) LineCount() int {
	return len(cb.Lines)
}

func (cb *ConcreteBuffer) LineLength(row int) int {
	if row >= 0 && row < len(cb.Lines) {
		return cb.Lines[row].Length
	}
	return 0
}

func (cb *ConcreteBuffer) RuneAt(row, col int) rune {
	if row >= 0 && row < len(cb.Content) && col >= 0 && col < len(cb.Content[row]) {
		return cb.Content[row][col]
	}
	return 0
}

func (cb *ConcreteBuffer) DeleteRange(r MotionRange) error {
	start := r.Start
	end := r.End

	// 如果是同一行内的删除
	if start.Row == end.Row {
		if start.Row < len(cb.Content) {
			content := cb.Content[start.Row]
			newContent := append(content[:start.Col], content[end.Col:]...)

			// 更新行长度
			cb.Lines[start.Row].Length = len(newContent)
			cb.Content[start.Row] = newContent
		}
		return nil
	}

	// 多行删除：将多行合并为一行
	if start.Row < len(cb.Content) && end.Row < len(cb.Content) {
		// 获取起始行的内容（到 start.Col 截断）
		startLineContent := cb.Content[start.Row]
		prefix := startLineContent[:start.Col]

		// 获取结束行的内容（从 end.Col 开始）
		endLineContent := cb.Content[end.Row]
		suffix := endLineContent[end.Col:]

		// 合并前缀和后缀
		mergedLine := append(prefix, suffix...)

		// 替换起始行的内容
		cb.Content[start.Row] = mergedLine
		cb.Lines[start.Row].Length = len(mergedLine)

		// 删除中间的所有行（包括结束行）
		rowsToDelete := end.Row - start.Row
		newLines := make([]Line, 0, len(cb.Lines)-rowsToDelete)
		newContent := make([][]rune, 0, len(cb.Content)-rowsToDelete)

		for i := 0; i < len(cb.Lines); i++ {
			if i < start.Row || i > end.Row {
				newLines = append(newLines, cb.Lines[i])
				newContent = append(newContent, cb.Content[i])
			} else if i == start.Row {
				// 已经处理过的行，跳过
			}
		}

		cb.Lines = newLines
		cb.Content = newContent
	}

	return nil
}

// CursorEngine 是真正的坐标计算引擎
type CursorEngine struct {
	Cursor *Cursor
	Buffer Buffer
}

// clamp 函数用于限制值在指定范围内
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// clampCursor 内部方法，用于限制光标位置
func (e *CursorEngine) clampCursor(row, col int) (int, int) {
	if e.Buffer == nil {
		return row, col
	}

	row = clamp(row, 0, e.Buffer.LineCount()-1)

	maxCol := 0
	if row >= 0 && row < e.Buffer.LineCount() {
		maxCol = e.Buffer.LineLength(row)
		if maxCol > 0 {
			maxCol-- // Length 是实际长度，所以最大索引是 Length-1
		}
	}
	col = clamp(col, 0, maxCol)

	return row, col
}

// ApplyMotion 应用运动结果（统一处理逻辑）
func (e *CursorEngine) ApplyMotion(r *MotionResult) error {
	if r.Range != nil {
		e.Cursor.Row = r.Range.End.Row
		e.Cursor.Col = r.Range.End.Col
		return nil
	}

	// fallback: vector motion
	newRow := e.Cursor.Row + r.DeltaRow
	newCol := e.Cursor.Col + r.DeltaCol
	e.Cursor.Row, e.Cursor.Col = e.clampCursor(newRow, newCol)
	return nil
}

// MoveCursor 移动光标（唯一副作用）
func (e *CursorEngine) MoveCursor(r *MotionResult) error {
	return e.ApplyMotion(r)
}

// DeleteRange 删除指定范围的内容
func (e *CursorEngine) DeleteRange(r *MotionRange) error {
	if e.Buffer == nil {
		return errors.New("no buffer available")
	}

	err := e.Buffer.DeleteRange(*r)
	if err != nil {
		return err
	}

	// 移动光标到开始位置
	e.Cursor.Row = r.Start.Row
	e.Cursor.Col = r.Start.Col

	return nil
}

// ErrInvalidMotion 表示无效的移动动作
var ErrInvalidMotion = errors.New("invalid motion")

// ComputeMotion 计算移动结果（只算，不动）
func (e *CursorEngine) ComputeMotion(m *Motion) (*MotionResult, error) {
	handler, ok := motionTable[m.Kind]
	if !ok {
		return nil, ErrInvalidMotion
	}

	return handler(e, m)
}

// simpleVector 返回一个简单的向量运动处理器
func simpleVector(dr, dc int) motionHandler {
	return func(e *CursorEngine, m *Motion) (*MotionResult, error) {
		count := m.Count
		if count <= 0 {
			count = 1
		}
		return &MotionResult{
			DeltaRow: dr * count,
			DeltaCol: dc * count,
		}, nil
	}
}

// classify 将字符分类
func classify(r rune) CharClass {
	switch {
	case r == ' ' || r == '\t':
		return ClassWhitespace
	case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_':
		return ClassWord
	default:
		return ClassPunct
	}
}

// wordForward 实现向前单词移动
func wordForward(e *CursorEngine, m *Motion) (*MotionResult, error) {
	row, col := e.Cursor.Row, e.Cursor.Col
	start := Cursor{Row: row, Col: col}

	count := m.Count
	if count <= 0 {
		count = 1
	}

	for i := 0; i < count; i++ {
		row, col = nextWord(e.Buffer, row, col)
	}

	end := Cursor{Row: row, Col: col}

	rangeResult := &MotionRange{
		Start: start,
		End:   end,
	}

	return &MotionResult{
		DeltaRow: end.Row - start.Row,
		DeltaCol: end.Col - start.Col,
		Range:    rangeResult,
	}, nil
}

// nextWord 找到下一个单词的位置
func nextWord(b Buffer, row, col int) (int, int) {
	if b == nil || row >= b.LineCount() {
		return row, col
	}

	// 如果当前行不存在或列超出范围，返回原位置
	if row < 0 || col >= b.LineLength(row) {
		return row, col
	}

	// Step 1: 获取当前位置的字符类别
	currentClass := classify(b.RuneAt(row, col))

	// Step 2: 跳过当前 class 的连续字符
	for {
		col++
		if col >= b.LineLength(row) {
			// 到达行尾，尝试下一行
			row++
			col = 0
			if row >= b.LineCount() {
				// 到达缓冲区末尾
				return row, col
			}
			// 当到达新行时，将当前类别视为空白，以便跳过开头的空白
			currentClass = ClassWhitespace
			continue
		}

		nextClass := classify(b.RuneAt(row, col))
		if nextClass != currentClass {
			// 类别发生变化，跳出循环
			break
		}
	}

	// Step 3: 跳过空白字符，直到遇到非空白字符
	for {
		if col >= b.LineLength(row) {
			// 到达行尾，尝试下一行
			row++
			col = 0
			if row >= b.LineCount() {
				// 到达缓冲区末尾
				return row, col
			}
			continue
		}

		charClass := classify(b.RuneAt(row, col))
		if charClass != ClassWhitespace {
			// 遇到非空白字符，跳出循环
			break
		}
		col++
	}

	return row, col
}