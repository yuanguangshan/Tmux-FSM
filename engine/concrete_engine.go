package engine

import (
	"tmux-fsm/editor"
	"tmux-fsm/intent"
)

// ConcreteEngine 是 Engine 接口的具体实现
type ConcreteEngine struct {
	// 这里可以添加实际的编辑器状态
	cursor editor.Cursor
}

// NewConcreteEngine 创建一个新的 ConcreteEngine 实例
func NewConcreteEngine() *ConcreteEngine {
	return &ConcreteEngine{
		cursor: editor.Cursor{Row: 0, Col: 0},
	}
}

// Cursor 返回当前光标位置
func (e *ConcreteEngine) Cursor() editor.Cursor {
	return e.cursor
}

// ComputeMotion 计算运动产生的范围
func (e *ConcreteEngine) ComputeMotion(m *intent.Motion) (editor.MotionRange, error) {
	switch m.Kind {
	case intent.MotionRange:
		if m.Range != nil && m.Range.Kind == intent.RangeTextObject {
			return e.computeTextObject(m.Range.TextObject)
		}
	case intent.MotionWord:
		return e.computeWord(m.Count)
	case intent.MotionLine:
		return e.computeLine(m.Count)
	case intent.MotionChar:
		return e.computeChar(m.Count)
	case intent.MotionGoto:
		return e.computeGoto(m.Count)
	case intent.MotionFind:
		if m.Find != nil {
			return e.computeFindMotion(m.Find, m.Count)
		}
	}

	// 默认返回当前位置的范围
	return editor.MotionRange{
		Start: e.cursor,
		End:   e.cursor,
	}, nil
}

// computeTextObject 计算文本对象的范围
func (e *ConcreteEngine) computeTextObject(textObj *intent.TextObject) (editor.MotionRange, error) {
	// 这里需要实际的文本分析逻辑
	// 现在返回一个示例范围
	start := e.cursor
	end := e.cursor

	switch textObj.Object {
	case intent.Word:
		// 计算单词边界
		if textObj.Scope == intent.Inner {
			// 内部单词：从单词开始到单词结束
		} else {
			// 周围单词：包含周围的空白字符
		}
	case intent.Paren:
		// 计算括号内的内容或包括括号
		if textObj.Scope == intent.Inner {
			// 内部括号：括号内的内容
		} else {
			// 周围括号：包括括号本身
		}
	case intent.QuoteDouble:
		// 计算双引号内的内容或包括引号
		if textObj.Scope == intent.Inner {
			// 内部引号：引号内的内容
		} else {
			// 周围引号：包括引号本身
		}
	}

	return editor.MotionRange{
		Start: start,
		End:   end,
	}, nil
}

// computeWord 计算单词移动的范围
func (e *ConcreteEngine) computeWord(count int) (editor.MotionRange, error) {
	start := e.cursor
	end := e.cursor

	// 这里需要实际的单词边界检测逻辑
	// 简单示例：移动 count 个单词
	for i := 0; i < count; i++ {
		// 实际实现中需要分析文本内容
		end.Col += 5 // 示例：假设每个单词平均5个字符
	}

	return editor.MotionRange{
		Start: start,
		End:   end,
	}, nil
}

// computeLine 计算行移动的范围
func (e *ConcreteEngine) computeLine(count int) (editor.MotionRange, error) {
	start := e.cursor
	end := e.cursor

	// 移动到第 count 行
	end.Row += count

	return editor.MotionRange{
		Start: start,
		End:   end,
	}, nil
}

// computeChar 计算字符移动的范围
func (e *ConcreteEngine) computeChar(count int) (editor.MotionRange, error) {
	start := e.cursor
	end := e.cursor

	// 移动 count 个字符
	end.Col += count

	return editor.MotionRange{
		Start: start,
		End:   end,
	}, nil
}

// computeGoto 计算跳转的范围
func (e *ConcreteEngine) computeGoto(count int) (editor.MotionRange, error) {
	start := e.cursor
	end := e.cursor

	// 跳转到指定位置（如果 count > 0）
	if count > 0 {
		end.Row = count - 1 // 行号从0开始
		end.Col = 0
	} else {
		// 默认跳转到文件开头
		end.Row = 0
		end.Col = 0
	}

	return editor.MotionRange{
		Start: start,
		End:   end,
	}, nil
}

// computeFindMotion 计算查找运动的范围
func (e *ConcreteEngine) computeFindMotion(find *intent.FindMotion, count int) (editor.MotionRange, error) {
	start := e.cursor
	end := e.cursor

	// 这里需要实际的查找逻辑
	// 简单示例：在当前行中查找字符
	if find != nil {
		// 模拟当前行的文本内容
		line := "sample text for testing find motions like fx tx Fx Tx"

		pos := start.Col
		step := 1
		if find.Direction == intent.FindBackward {
			step = -1
		}

		matches := 0
		i := pos + step

		for i >= 0 && i < len(line) {
			if rune(line[i]) == find.Char {
				matches++
				if matches == count {
					target := i

					// till 的偏移规则
					if find.Till {
						if find.Direction == intent.FindForward {
							target--
						} else {
							target++
						}
					}

					end.Col = clamp(target, 0, len(line)-1)

					return editor.MotionRange{
						Start: start,
						End:   editor.Cursor{Row: start.Row, Col: end.Col},
					}, nil
				}
			}
			i += step
		}
	}

	// Vim 行为：找不到 → 光标不动
	return editor.MotionRange{
		Start: start,
		End:   start,
	}, nil
}

// clamp 辅助函数
func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// MoveCursor 移动光标到指定范围
func (e *ConcreteEngine) MoveCursor(r editor.MotionRange) error {
	e.cursor = r.End
	return nil
}

// DeleteRange 删除指定范围的内容
func (e *ConcreteEngine) DeleteRange(r editor.MotionRange) error {
	// 实际实现中需要与底层编辑器交互
	return nil
}

// YankRange 复制指定范围的内容
func (e *ConcreteEngine) YankRange(r editor.MotionRange) error {
	// 实际实现中需要与底层编辑器交互
	return nil
}

// ChangeRange 修改指定范围的内容
func (e *ConcreteEngine) ChangeRange(r editor.MotionRange) error {
	// 实际实现中需要与底层编辑器交互
	return nil
}
