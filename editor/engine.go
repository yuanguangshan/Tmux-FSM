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