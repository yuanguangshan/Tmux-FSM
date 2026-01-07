package main

import (
	"errors"
)

// TextObjectKind 定义文本对象类型
type TextObjectKind int

const (
	TextObjectWord TextObjectKind = iota
	TextObjectParen
	TextObjectBracket
	TextObjectBrace
	TextObjectQuoteDouble
	TextObjectQuoteSingle
	TextObjectParagraph
	TextObjectSentence
)

// TextObjectMotion 定义文本对象运动
type TextObjectMotion struct {
	Kind     TextObjectKind
	Inner    bool // true for 'i', false for 'a'
}

// TextObjectRangeCalculator 计算文本对象范围的接口
type TextObjectRangeCalculator interface {
	CalculateRange(obj TextObjectMotion, cursor Cursor) (*MotionRange, error)
}

// ConcreteTextObjectCalculator 实现文本对象范围计算器
type ConcreteTextObjectCalculator struct {
	Buffer Buffer
}

// NewConcreteTextObjectCalculator 创建新的文本对象计算器
func NewConcreteTextObjectCalculator(buffer Buffer) *ConcreteTextObjectCalculator {
	return &ConcreteTextObjectCalculator{
		Buffer: buffer,
	}
}

// CalculateRange 计算文本对象范围
func (calc *ConcreteTextObjectCalculator) CalculateRange(obj TextObjectMotion, cursor Cursor) (*MotionRange, error) {
	switch obj.Kind {
	case TextObjectWord:
		return calc.calculateWordRange(obj.Inner, cursor)
	case TextObjectParen:
		return calc.calculateDelimitedRange('(', ')', obj.Inner, cursor)
	case TextObjectBracket:
		return calc.calculateDelimitedRange('[', ']', obj.Inner, cursor)
	case TextObjectBrace:
		return calc.calculateDelimitedRange('{', '}', obj.Inner, cursor)
	case TextObjectQuoteDouble:
		return calc.calculateQuoteRange('"', obj.Inner, cursor)
	case TextObjectQuoteSingle:
		return calc.calculateQuoteRange('\'', obj.Inner, cursor)
	case TextObjectParagraph:
		return calc.calculateParagraphRange(obj.Inner, cursor)
	case TextObjectSentence:
		return calc.calculateSentenceRange(obj.Inner, cursor)
	default:
		return nil, errors.New("unsupported text object")
	}
}

// calculateWordRange 计算单词范围
func (calc *ConcreteTextObjectCalculator) calculateWordRange(inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	row := cursor.Row
	if row < 0 || row >= calc.Buffer.LineCount() {
		return nil, errors.New("invalid row")
	}

	line := make([]rune, calc.Buffer.LineLength(row))
	for i := 0; i < len(line); i++ {
		line[i] = calc.Buffer.RuneAt(row, i)
	}

	startCol, endCol := findWordAt(line, cursor.Col, inner)

	return &MotionRange{
		Start: Cursor{Row: row, Col: startCol},
		End:   Cursor{Row: row, Col: endCol},
	}, nil
}

// findWordAt 查找光标位置的单词范围
func findWordAt(line []rune, col int, inner bool) (int, int) {
	if len(line) == 0 || col < 0 {
		return 0, 0
	}

	if col >= len(line) {
		col = len(line) - 1
	}

	// 确定字符类别
	charType := classifyRune(line[col])

	// 向左查找边界
	start := col
	for start > 0 {
		if classifyRune(line[start-1]) != charType {
			break
		}
		start--
	}

	// 向右查找边界
	end := col
	for end < len(line)-1 {
		if classifyRune(line[end+1]) != charType {
			break
		}
		end++
	}

	// 如果是 inner 模式，去除两端的空白
	if inner {
		for start <= end && start < len(line) && isWhitespace(line[start]) {
			start++
		}
		for end > start && end >= 0 && isWhitespace(line[end]) {
			end--
		}
	}

	// 确保 end 在有效范围内
	if end >= len(line) {
		end = len(line) - 1
	}

	// 确保范围有效
	if start > end {
		start = end
	}

	// 如果是 outer 模式，扩展到包含相邻的空白
	if !inner {
		// 向右扩展包含空白
		for end < len(line)-1 && isWhitespace(line[end+1]) {
			end++
		}
		// 向左扩展包含空白
		for start > 0 && isWhitespace(line[start-1]) {
			start--
		}
	}

	return start, end + 1
}

// classifyRune 将字符分类
func classifyRune(r rune) CharClass {
	switch {
	case r == ' ' || r == '\t' || r == '\n' || r == '\r':
		return ClassWhitespace
	case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_':
		return ClassWord
	default:
		return ClassPunct
	}
}

// isWhitespace 检查是否为空白字符
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// calculateDelimitedRange 计算定界符范围
func (calc *ConcreteTextObjectCalculator) calculateDelimitedRange(open, close rune, inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	// 从当前行开始搜索
	startPos, endPos := findDelimitedRange(calc.Buffer, open, close, cursor, inner)
	
	if startPos.Row == -1 || endPos.Row == -1 {
		return nil, errors.New("delimited range not found")
	}

	return &MotionRange{
		Start: startPos,
		End:   endPos,
	}, nil
}

// findDelimitedRange 查找定界符范围
func findDelimitedRange(buffer Buffer, open, close rune, cursor Cursor, inner bool) (Cursor, Cursor) {
	// 从当前光标位置开始查找匹配的定界符
	currentRow := cursor.Row
	currentCol := cursor.Col
	
	// 首先尝试在当前行查找
	for row := currentRow; row < buffer.LineCount(); row++ {
		lineLen := buffer.LineLength(row)
		startCol := 0
		if row == currentRow {
			startCol = currentCol
		}
		
		for col := startCol; col < lineLen; col++ {
			r := buffer.RuneAt(row, col)
			if r == open {
				// 找到开定界符，查找对应的闭定界符
				endPos := findMatchingDelimiter(buffer, open, close, Cursor{Row: row, Col: col})
				if endPos.Row != -1 {
					if inner {
						// Inner 模式：排除定界符本身
						return Cursor{Row: row, Col: col + 1}, endPos
					} else {
						// Outer 模式：包含定界符
						return Cursor{Row: row, Col: col}, Cursor{Row: endPos.Row, Col: endPos.Col + 1}
					}
				}
			}
		}
	}
	
	// 如果没找到，返回无效位置
	return Cursor{Row: -1, Col: -1}, Cursor{Row: -1, Col: -1}
}

// findMatchingDelimiter 查找匹配的定界符
func findMatchingDelimiter(buffer Buffer, open, close rune, startPos Cursor) Cursor {
	stack := 0
	currentRow := startPos.Row
	currentCol := startPos.Col + 1 // 从开定界符的下一个位置开始
	
	for row := currentRow; row < buffer.LineCount(); row++ {
		lineLen := buffer.LineLength(row)
		startCol := 0
		if row == currentRow {
			startCol = currentCol
		}
		
		for col := startCol; col < lineLen; col++ {
			r := buffer.RuneAt(row, col)
			if r == open {
				stack++
			} else if r == close {
				stack--
				if stack < 0 {
					// 找到匹配的闭定界符
					return Cursor{Row: row, Col: col}
				}
			}
		}
		currentCol = 0 // 从下一行开始时，列从0开始
	}
	
	// 没有找到匹配的闭定界符
	return Cursor{Row: -1, Col: -1}
}

// calculateQuoteRange 计算引号范围
func (calc *ConcreteTextObjectCalculator) calculateQuoteRange(quote rune, inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	// 从当前光标位置开始查找引号
	currentRow := cursor.Row
	currentCol := cursor.Col
	
	// 首先检查光标位置是否在引号内或旁边
	for row := currentRow; row < calc.Buffer.LineCount(); row++ {
		lineLen := calc.Buffer.LineLength(row)
		startCol := 0
		if row == currentRow {
			startCol = currentCol
		}
		
		for col := startCol; col < lineLen; col++ {
			r := calc.Buffer.RuneAt(row, col)
			if r == quote {
				// 找到第一个引号，查找匹配的另一个
				endPos := findMatchingQuote(calc.Buffer, quote, Cursor{Row: row, Col: col})
				if endPos.Row != -1 {
					if inner {
						// Inner 模式：排除引号本身
						return &MotionRange{
							Start: Cursor{Row: row, Col: col + 1},
							End:   endPos,
						}, nil
					} else {
						// Outer 模式：包含引号
						return &MotionRange{
							Start: Cursor{Row: row, Col: col},
							End:   Cursor{Row: endPos.Row, Col: endPos.Col + 1},
						}, nil
					}
				}
			}
		}
	}
	
	return nil, errors.New("quote range not found")
}

// findMatchingQuote 查找匹配的引号
func findMatchingQuote(buffer Buffer, quote rune, startPos Cursor) Cursor {
	escaped := false

	currentRow := startPos.Row
	currentCol := startPos.Col + 1 // 从第一个引号的下一个位置开始

	for row := currentRow; row < buffer.LineCount(); row++ {
		lineLen := buffer.LineLength(row)
		startCol := 0
		if row == currentRow {
			startCol = currentCol
		}

		for col := startCol; col < lineLen; col++ {
			r := buffer.RuneAt(row, col)

			if escaped {
				escaped = false
				continue
			}

			if r == '\\' {
				escaped = true
				continue
			}

			if r == quote {
				// 找到匹配的引号
				return Cursor{Row: row, Col: col}
			}
		}
		currentCol = 0 // 从下一行开始时，列从0开始
	}

	// 没有找到匹配的引号
	return Cursor{Row: -1, Col: -1}
}

// calculateParagraphRange 计算段落范围
func (calc *ConcreteTextObjectCalculator) calculateParagraphRange(inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	// 简化实现：查找空行分隔的段落
	startRow := cursor.Row
	endRow := cursor.Row

	// 向上查找段落开始
	for startRow > 0 {
		lineLen := calc.Buffer.LineLength(startRow - 1)
		if lineLen == 0 {
			break
		}
		startRow--
	}

	// 向下查找段落结束
	for endRow < calc.Buffer.LineCount()-1 {
		lineLen := calc.Buffer.LineLength(endRow + 1)
		if lineLen == 0 {
			break
		}
		endRow++
	}

	if inner {
		// Inner 模式：排除段落周围的空行
		return &MotionRange{
			Start: Cursor{Row: startRow, Col: 0},
			End:   Cursor{Row: endRow, Col: calc.Buffer.LineLength(endRow)},
		}, nil
	} else {
		// Outer 模式：包含整个段落
		return &MotionRange{
			Start: Cursor{Row: startRow, Col: 0},
			End:   Cursor{Row: endRow + 1, Col: 0}, // 包含下一行的开始
		}, nil
	}
}

// calculateSentenceRange 计算句子范围
func (calc *ConcreteTextObjectCalculator) calculateSentenceRange(inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	// 简化实现：查找句号、感叹号、问号分隔的句子
	currentRow := cursor.Row
	currentCol := cursor.Col

	// 查找当前句子的开始
	startRow, startCol := findSentenceStart(calc.Buffer, currentRow, currentCol)

	// 查找当前句子的结束
	endRow, endCol := findSentenceEnd(calc.Buffer, currentRow, currentCol)

	if inner {
		// Inner 模式：排除句子结束标点
		return &MotionRange{
			Start: Cursor{Row: startRow, Col: startCol},
			End:   Cursor{Row: endRow, Col: endCol},
		}, nil
	} else {
		// Outer 模式：包含句子结束标点及后续空白
		// 简化：包含到句子结束
		return &MotionRange{
			Start: Cursor{Row: startRow, Col: startCol},
			End:   Cursor{Row: endRow, Col: endCol + 1},
		}, nil
	}
}

// findSentenceStart 查找句子开始
func findSentenceStart(buffer Buffer, row, col int) (int, int) {
	// 简化实现：查找前一个句子结束符后的第一个非空白字符
	for r := row; r >= 0; r-- {
		lineLen := buffer.LineLength(r)
		startCol := lineLen - 1
		if r == row {
			startCol = col
		}
		
		for c := startCol; c >= 0; c-- {
			runeVal := buffer.RuneAt(r, c)
			if runeVal == '.' || runeVal == '!' || runeVal == '?' {
				// 找到句子结束符，下一个位置是句子开始
				nextRow, nextCol := getNextNonWhitespace(buffer, r, c+1)
				return nextRow, nextCol
			}
		}
	}
	
	// 如果没找到，返回文件开始
	return 0, 0
}

// findSentenceEnd 查找句子结束
func findSentenceEnd(buffer Buffer, row, col int) (int, int) {
	// 简化实现：查找下一个句子结束符
	for r := row; r < buffer.LineCount(); r++ {
		lineLen := buffer.LineLength(r)
		startCol := 0
		if r == row {
			startCol = col
		}
		
		for c := startCol; c < lineLen; c++ {
			runeVal := buffer.RuneAt(r, c)
			if runeVal == '.' || runeVal == '!' || runeVal == '?' {
				// 找到句子结束符
				return r, c
			}
		}
	}
	
	// 如果没找到，返回文件结束
	endRow := buffer.LineCount() - 1
	endCol := buffer.LineLength(endRow)
	return endRow, endCol
}

// getNextNonWhitespace 获取下一个非空白字符位置
func getNextNonWhitespace(buffer Buffer, row, col int) (int, int) {
	for r := row; r < buffer.LineCount(); r++ {
		lineLen := buffer.LineLength(r)
		startCol := 0
		if r == row {
			startCol = col
		}
		
		for c := startCol; c < lineLen; c++ {
			runeVal := buffer.RuneAt(r, c)
			if !isWhitespace(runeVal) {
				return r, c
			}
		}
	}
	
	// 如果没找到，返回当前位置
	return row, col
}

// ParseTextObject 解析文本对象字符串
func ParseTextObject(textObjectStr string) (*TextObjectMotion, error) {
	if len(textObjectStr) < 2 {
		return nil, errors.New("invalid text object string")
	}

	modifier := textObjectStr[0:1]
	objType := textObjectStr[1:2]

	inner := modifier == "i"
	
	var kind TextObjectKind
	switch objType {
	case "w":
		kind = TextObjectWord
	case "(":
		kind = TextObjectParen
	case "[":
		kind = TextObjectBracket
	case "{":
		kind = TextObjectBrace
	case "\"":
		kind = TextObjectQuoteDouble
	case "'":
		kind = TextObjectQuoteSingle
	case "p":
		kind = TextObjectParagraph
	case "s":
		kind = TextObjectSentence
	default:
		return nil, errors.New("unsupported text object type")
	}

	return &TextObjectMotion{
		Kind:  kind,
		Inner: inner,
	}, nil
}