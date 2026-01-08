package editor

import (
	"errors"
)

// Buffer 接口定义
type Buffer interface {
	LineCount() int
	LineLength(row int) int
	RuneAt(row, col int) rune
}

// SimpleBuffer 简单的缓冲区实现
type SimpleBuffer struct {
	lines []string
}

// NewSimpleBuffer 创建新的简单缓冲区
func NewSimpleBuffer() *SimpleBuffer {
	return &SimpleBuffer{
		lines: []string{""}, // 默认一行空字符串
	}
}

// LineCount 返回行数
func (sb *SimpleBuffer) LineCount() int {
	return len(sb.lines)
}

// LineLength 返回指定行的长度
func (sb *SimpleBuffer) LineLength(row int) int {
	if row < 0 || row >= len(sb.lines) {
		return 0
	}
	return len(sb.lines[row])
}

// RuneAt 返回指定位置的字符
func (sb *SimpleBuffer) RuneAt(row, col int) rune {
	if row < 0 || row >= len(sb.lines) {
		return 0
	}
	line := sb.lines[row]
	if col < 0 || col >= len(line) {
		return 0
	}
	return rune(line[col])
}

// InsertAt 在指定位置插入文本
func (sb *SimpleBuffer) InsertAt(anchor Cursor, text string) error {
	if anchor.Row < 0 || anchor.Row >= len(sb.lines) {
		return errors.New("invalid row")
	}

	line := sb.lines[anchor.Row]
	if anchor.Col < 0 || anchor.Col > len(line) {
		return errors.New("invalid column")
	}

	// 在指定位置插入文本
	newLine := line[:anchor.Col] + text + line[anchor.Col:]
	sb.lines[anchor.Row] = newLine

	return nil
}

// DeleteRange 删除指定范围的文本
func (sb *SimpleBuffer) DeleteRange(start, end Cursor) error {
	if start.Row < 0 || start.Row >= len(sb.lines) || end.Row < 0 || end.Row >= len(sb.lines) {
		return errors.New("invalid row")
	}

	if start.Row == end.Row {
		// 同一行删除
		if start.Col < 0 || end.Col > len(sb.lines[start.Row]) || start.Col >= end.Col {
			return errors.New("invalid column range")
		}

		line := sb.lines[start.Row]
		newLine := line[:start.Col] + line[end.Col:]
		sb.lines[start.Row] = newLine
	} else if start.Row < end.Row {
		// 跨行删除
		// 1. 删除结束行的开始部分
		endLine := sb.lines[end.Row]
		remainder := endLine[end.Col:]

		// 2. 删除起始行的结束部分
		startLine := sb.lines[start.Row]
		prefix := startLine[:start.Col]

		// 3. 合并前缀和后缀
		newLine := prefix + remainder

		// 4. 删除中间的整行
		newLines := make([]string, 0, len(sb.lines))
		newLines = append(newLines, sb.lines[:start.Row]...)
		newLines = append(newLines, newLine)
		newLines = append(newLines, sb.lines[end.Row+1:]...)

		sb.lines = newLines
	} else {
		return errors.New("end position is before start position")
	}

	return nil
}

// GetTextInRange 获取指定范围的文本
func (sb *SimpleBuffer) GetTextInRange(start, end Cursor) (string, error) {
	if start.Row < 0 || start.Row >= len(sb.lines) || end.Row < 0 || end.Row >= len(sb.lines) {
		return "", errors.New("invalid row")
	}

	if start.Row == end.Row {
		// 同一行
		if start.Col < 0 || end.Col > len(sb.lines[start.Row]) || start.Col >= end.Col {
			return "", errors.New("invalid column range")
		}

		return sb.lines[start.Row][start.Col:end.Col], nil
	} else if start.Row < end.Row {
		// 跨行
		result := sb.lines[start.Row][start.Col:] // 从起始位置到行尾

		// 添加中间的整行
		for i := start.Row + 1; i < end.Row; i++ {
			result += sb.lines[i] + "\n" // 假设换行符
		}

		// 添加结束行的开始部分
		result += sb.lines[end.Row][:end.Col]

		return result, nil
	} else {
		return "", errors.New("end position is before start position")
	}
}

// Motion 定义动作
type Motion struct {
	Kind  MotionKind
	Count int
}

// MotionResult 动作结果
type MotionResult struct {
	DeltaRow int
	DeltaCol int
}

// CursorEngine 光标引擎
type CursorEngine struct {
	Cursor *Cursor
	Buffer Buffer
}

// NewCursorEngine 创建新的光标引擎
func NewCursorEngine(buffer Buffer) *CursorEngine {
	initialCursor := &Cursor{Row: 0, Col: 0}
	return &CursorEngine{
		Cursor: initialCursor,
		Buffer: buffer,
	}
}

// ComputeMotion 计算动作结果（不实际移动光标）
func (ce *CursorEngine) ComputeMotion(motion *Motion) (MotionResult, error) {
	// 这里实现动作计算逻辑
	// 为了简化，我们只处理一些基本动作
	switch motion.Kind {
	case MotionCharForward:
		return MotionResult{DeltaRow: 0, DeltaCol: motion.Count}, nil
	case MotionCharBackward:
		return MotionResult{DeltaRow: 0, DeltaCol: -motion.Count}, nil
	case MotionWordForward:
		// 简化实现：假设每个单词之间有一个空格
		return MotionResult{DeltaRow: 0, DeltaCol: motion.Count * 5}, nil // 每个单词平均5个字符
	case MotionWordBackward:
		return MotionResult{DeltaRow: 0, DeltaCol: -motion.Count * 5}, nil
	case MotionLineStart:
		return MotionResult{DeltaRow: 0, DeltaCol: -ce.Cursor.Col}, nil
	case MotionLineEnd:
		lineLength := ce.Buffer.LineLength(ce.Cursor.Row)
		return MotionResult{DeltaRow: 0, DeltaCol: lineLength - ce.Cursor.Col}, nil
	default:
		return MotionResult{}, errors.New("unsupported motion")
	}
}

// clamp 限制值在范围内
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// GlobalCursorEngine 全局光标引擎，用于操作缓冲区
var GlobalCursorEngine *CursorEngine

// ApplyResolvedOperation 应用解析后的操作
// 这是 . repeat 的核心执行函数
// 注意：此函数只执行预定义的操作，不做任何语义判断
func ApplyResolvedOperation(op ResolvedOperation) error {
	switch op.Kind {
	case OpInsert:
		return applyInsert(op)
	case OpDelete:
		return applyDelete(op)
	case OpMove:
		return applyMove(op)
	default:
		return errors.New("unsupported operation kind")
	}
}

// applyInsert 执行插入操作
func applyInsert(op ResolvedOperation) error {
	if GlobalCursorEngine == nil || GlobalCursorEngine.Buffer == nil {
		return errors.New("buffer not initialized")
	}

	buffer, ok := GlobalCursorEngine.Buffer.(interface{ InsertAt(Cursor, string) error })
	if !ok {
		return errors.New("buffer does not support InsertAt")
	}

	// 如果需要先删除范围内容（例如替换操作）
	if op.DeleteBeforeInsert && op.Range != nil {
		err := deleteRange(op.Range.Start, op.Range.End)
		if err != nil {
			return err
		}
	}

	// 在指定位置插入文本
	return buffer.InsertAt(op.Anchor, op.Text)
}

// applyDelete 执行删除操作
func applyDelete(op ResolvedOperation) error {
	if op.Range == nil {
		return errors.New("delete operation requires a range")
	}

	if GlobalCursorEngine == nil || GlobalCursorEngine.Buffer == nil {
		return errors.New("buffer not initialized")
	}

	buffer, ok := GlobalCursorEngine.Buffer.(interface{ DeleteRange(Cursor, Cursor) error })
	if !ok {
		return errors.New("buffer does not support DeleteRange")
	}

	return buffer.DeleteRange(op.Range.Start, op.Range.End)
}

// applyMove 执行移动操作
func applyMove(op ResolvedOperation) error {
	if GlobalCursorEngine != nil {
		// 更新光标位置到指定位置
		GlobalCursorEngine.Cursor.Row = op.Anchor.Row
		GlobalCursorEngine.Cursor.Col = op.Anchor.Col
	}

	return nil
}

// insertAt 在指定位置插入文本
func insertAt(anchor Cursor, text string) error {
	if GlobalCursorEngine == nil || GlobalCursorEngine.Buffer == nil {
		return errors.New("buffer not initialized")
	}

	buffer, ok := GlobalCursorEngine.Buffer.(interface{ InsertAt(Cursor, string) error })
	if !ok {
		return errors.New("buffer does not support InsertAt")
	}

	return buffer.InsertAt(anchor, text)
}

// deleteRange 删除指定范围的文本
func deleteRange(start, end Cursor) error {
	if GlobalCursorEngine == nil || GlobalCursorEngine.Buffer == nil {
		return errors.New("buffer not initialized")
	}

	buffer, ok := GlobalCursorEngine.Buffer.(interface{ DeleteRange(Cursor, Cursor) error })
	if !ok {
		return errors.New("buffer does not support DeleteRange")
	}

	return buffer.DeleteRange(start, end)
}