package editor

import (
	"errors"
	"fmt"
	"log"
)

// SimpleBuffer 简单的缓冲区实现
type SimpleBuffer struct {
	lines []string
}

// NewSimpleBuffer 创建新的简单缓冲区
func NewSimpleBuffer(initialText []string) *SimpleBuffer {
	if len(initialText) == 0 {
		initialText = []string{""}
	}
	return &SimpleBuffer{
		lines: initialText,
	}
}

func (sb *SimpleBuffer) LineCount() int {
	return len(sb.lines)
}

func (sb *SimpleBuffer) LineLength(row int) int {
	if row < 0 || row >= len(sb.lines) {
		return 0
	}
	return len(sb.lines[row])
}

func (sb *SimpleBuffer) Line(row int) string {
	if row < 0 || row >= len(sb.lines) {
		return ""
	}
	return sb.lines[row]
}

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

func (sb *SimpleBuffer) InsertAt(anchor Cursor, text string) error {
	if anchor.Row < 0 || anchor.Row >= len(sb.lines) {
		return errors.New("invalid row")
	}

	line := sb.lines[anchor.Row]
	if anchor.Col < 0 || anchor.Col > len(line) {
		return errors.New("invalid column")
	}

	newLine := line[:anchor.Col] + text + line[anchor.Col:]
	sb.lines[anchor.Row] = newLine

	return nil
}

func (sb *SimpleBuffer) DeleteRange(start, end Cursor) (string, error) {
	if start.Row < 0 || start.Row >= len(sb.lines) || end.Row < 0 || end.Row >= len(sb.lines) {
		return "", errors.New("invalid row")
	}

	// 确保 start <= end
	if end.Row < start.Row || (start.Row == end.Row && end.Col < start.Col) {
		start, end = end, start
	}

	var deletedText string
	if start.Row == end.Row {
		line := sb.lines[start.Row]
		if start.Col < 0 || end.Col > len(line) {
			return "", errors.New("invalid column range")
		}
		deletedText = line[start.Col:end.Col]
		sb.lines[start.Row] = line[:start.Col] + line[end.Col:]
	} else {
		// 跨行删除
		firstLine := sb.lines[start.Row]
		lastLine := sb.lines[end.Row]

		deletedText = firstLine[start.Col:] + "\n"
		for i := start.Row + 1; i < end.Row; i++ {
			deletedText += sb.lines[i] + "\n"
		}
		deletedText += lastLine[:end.Col]

		newLine := firstLine[:start.Col] + lastLine[end.Col:]

		newLines := make([]string, 0, len(sb.lines)-(end.Row-start.Row))
		newLines = append(newLines, sb.lines[:start.Row]...)
		newLines = append(newLines, newLine)
		newLines = append(newLines, sb.lines[end.Row+1:]...)
		sb.lines = newLines
	}

	return deletedText, nil
}

// ApplyResolvedOperation 应用解析后的操作
// 严格按照预定义的操作类型执行，无任何语义判断
func ApplyResolvedOperation(ctx *ExecutionContext, op ResolvedOperation) error {
	// Log the operation for audit trail
	log.Printf("Executing operation: Kind=%v, ID=%s", op.Kind(), op.OpID())

	// Handle generic buffer operations
	// Most operations (Insert, Delete, Move) follow the Buffer interface
	// For operations that need special context (like MoveCursor needing WindowStore),
	// we handle them via type switch or extension.

	switch actualOp := op.(type) {
	case *MoveCursorOperation:
		win := ctx.Windows.Get(actualOp.WindowID)
		if win != nil {
			log.Printf("Moving cursor in window %s from %v to %v", actualOp.WindowID, win.Cursor, actualOp.To)
			win.Cursor = actualOp.To
		} else {
			log.Printf("Window %s not found for move cursor operation", actualOp.WindowID)
		}
		return nil

	case *CompositeOperation:
		return applyInterface(ctx, op)

	default:
		return applyInterface(ctx, op)
	}
}

func applyInterface(ctx *ExecutionContext, op ResolvedOperation) error {
	// Determine BufferID from Footprint
	fp := op.Footprint()
	if len(fp.Buffers) == 0 {
		return op.Apply(nil) // Some operations might be context-free
	}

	bufferID := fp.Buffers[0]
	buf := ctx.Buffers.Get(bufferID)
	if buf == nil {
		return fmt.Errorf("buffer %s not found", bufferID)
	}

	return op.Apply(buf)
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
