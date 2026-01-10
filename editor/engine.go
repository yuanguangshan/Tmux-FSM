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
	log.Printf("Executing operation: Kind=%s, BufferID=%s, WindowID=%s, Anchor=%v",
		op.Kind, op.BufferID, op.WindowID, op.Anchor)

	// Stronger validation of operation parameters
	if op.BufferID == "" {
		err := errors.New("operation requires a valid buffer ID")
		log.Printf("Validation error: %v", err)
		return err
	}

	buf := ctx.Buffers.Get(op.BufferID)
	if buf == nil {
		err := fmt.Errorf("buffer %s not found", op.BufferID)
		log.Printf("Execution error: %v", err)
		return err
	}

	// Validate cursor position for insert operations
	if op.Kind == OpInsert {
		if op.Anchor.Row < 0 || op.Anchor.Row >= buf.LineCount() {
			err := fmt.Errorf("insert position out of bounds: row %d, total rows %d", op.Anchor.Row, buf.LineCount())
			log.Printf("Validation error: %v", err)
			return err
		}
		if op.Anchor.Col < 0 || op.Anchor.Col > buf.LineLength(op.Anchor.Row) {
			err := fmt.Errorf("insert position out of bounds: col %d, line length %d", op.Anchor.Col, buf.LineLength(op.Anchor.Row))
			log.Printf("Validation error: %v", err)
			return err
		}
	}

	// Validate range for delete operations
	if op.Kind == OpDelete && op.Range != nil {
		if op.Range.Start.Row < 0 || op.Range.Start.Row >= buf.LineCount() ||
			op.Range.End.Row < 0 || op.Range.End.Row >= buf.LineCount() {
			err := fmt.Errorf("delete range out of bounds: start row %d, end row %d, total rows %d",
				op.Range.Start.Row, op.Range.End.Row, buf.LineCount())
			log.Printf("Validation error: %v", err)
			return err
		}
	}

	switch op.Kind {
	case OpInsert:
		log.Printf("Applying insert operation: text='%s' at %v", op.Text, op.Anchor)
		if op.DeleteBeforeInsert && op.Range != nil {
			deletedText, err := buf.DeleteRange(op.Range.Start, op.Range.End)
			if err != nil {
				log.Printf("Failed to delete before insert: %v", err)
				return err
			}
			log.Printf("Deleted text before insert: '%s'", deletedText)
		}
		err := buf.InsertAt(op.Anchor, op.Text)
		if err != nil {
			log.Printf("Failed to insert text: %v", err)
			return err
		}
		log.Printf("Successfully inserted text: '%s' at %v", op.Text, op.Anchor)
		return nil

	case OpDelete:
		if op.Range == nil {
			err := errors.New("delete operation requires a range")
			log.Printf("Validation error: %v", err)
			return err
		}
		deletedText, err := buf.DeleteRange(op.Range.Start, op.Range.End)
		if err != nil {
			log.Printf("Failed to delete range: %v", err)
			return err
		}
		log.Printf("Successfully deleted text: '%s' from range %v to %v", deletedText, op.Range.Start, op.Range.End)
		return nil

	case OpMove:
		win := ctx.Windows.Get(op.WindowID)
		if win != nil {
			log.Printf("Moving cursor from %v to %v", win.Cursor, op.Anchor)
			win.Cursor = op.Anchor
		} else {
			log.Printf("Window %s not found for move operation", op.WindowID)
		}
		return nil

	default:
		err := errors.New("unsupported operation kind")
		log.Printf("Execution error: %v", err)
		return err
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
