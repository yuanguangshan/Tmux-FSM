package resolver

import (
	"tmux-fsm/intent"
	"unicode"
)

// Range 表示一个范围
type Range struct {
	Start Pos
	End   Pos
}

// Pos 表示位置
type Pos struct {
	Line int
	Col  int
}

// Buffer 接口，用于获取文本内容
type Buffer interface {
	Line(lineNum int) string
}

// MotionResolver 负责解析 motion 到范围
type MotionResolver struct {
	Buffer Buffer
}

// NewMotionResolver 创建新的 MotionResolver
func NewMotionResolver(buffer Buffer) *MotionResolver {
	return &MotionResolver{
		Buffer: buffer,
	}
}

// ResolveOpMotion 解析操作符+motion 到范围
func (r *MotionResolver) ResolveOpMotion(
	intentObj *intent.Intent,
	cursor Pos,
) ([]Range, error) {

	if intentObj.Kind != intent.IntentOperator {
		return nil, nil
	}

	meta, ok := intentObj.Meta["operator"]
	if !ok {
		return nil, nil
	}

	_, ok = meta.(intent.OperatorKind)
	if !ok {
		return nil, nil
	}

	motionMeta, ok := intentObj.Meta["motion"]
	if !ok {
		return nil, nil
	}

	motion, ok := motionMeta.(intent.MotionKind)
	if !ok {
		return nil, nil
	}

	// 特殊处理 $ 和 0 motion
	count := intentObj.Count
	if intentObj.Meta["motion_special"] != nil {
		// 如果有特殊 motion 标记，调整 count
		if special, ok := intentObj.Meta["motion_special"].(string); ok {
			switch special {
			case "line_end": // $
				count = -1
			case "line_start": // 0
				count = -2
			}
		}
	}

	end, err := r.resolveMotion(motion, cursor, count)
	if err != nil {
		return nil, err
	}

	return []Range{r.normalize(cursor, end)}, nil
}

// resolveMotion 解析 motion 到结束位置
func (r *MotionResolver) resolveMotion(
	motion intent.MotionKind,
	cursor Pos,
	count int,
) (Pos, error) {

	if count <= 0 {
		count = 1
	}

	switch motion {
	case intent.MotionChar:
		// 特殊处理行首和行尾
		if count == -1 { // 行尾
			return r.resolveLineEndMotion(cursor)
		} else if count == -2 { // 行首
			return r.resolveLineStartMotion(cursor)
		}
		return r.resolveCharMotion(cursor, count)
	case intent.MotionWord:
		return r.resolveWordMotion(cursor, count)
	case intent.MotionLine:
		return r.resolveLineMotion(cursor, count)
	case intent.MotionGoto:
		return r.resolveGotoMotion(cursor, count)
	default:
		return cursor, nil
	}
}

// resolveCharMotion 解析字符 motion
func (r *MotionResolver) resolveCharMotion(cursor Pos, count int) (Pos, error) {
	line := r.Buffer.Line(cursor.Line)
	newCol := cursor.Col

	// 一般字符移动
	if newCol+count < len(line) {
		newCol += count
	} else {
		newCol = len(line)
	}

	return Pos{Line: cursor.Line, Col: newCol}, nil
}

// resolveLineEndMotion 解析行尾 motion ($)
func (r *MotionResolver) resolveLineEndMotion(cursor Pos) (Pos, error) {
	line := r.Buffer.Line(cursor.Line)
	return Pos{Line: cursor.Line, Col: len(line)}, nil
}

// resolveLineStartMotion 解析行首 motion (0)
func (r *MotionResolver) resolveLineStartMotion(cursor Pos) (Pos, error) {
	return Pos{Line: cursor.Line, Col: 0}, nil
}



// resolveWordMotion 解析单词 motion
func (r *MotionResolver) resolveWordMotion(cursor Pos, count int) (Pos, error) {
	line := r.Buffer.Line(cursor.Line)
	i := cursor.Col

	for c := 0; c < count; c++ {
		// 跳过当前 word 或空白
		if i < len(line) {
			if isWordChar(rune(line[i])) {
				// 跳过当前 word
				for i < len(line) && isWordChar(rune(line[i])) {
					i++
				}
			} else {
				// 跳过空白
				for i < len(line) && !isWordChar(rune(line[i])) {
					i++
				}
				// 如果现在在 word 上，跳过这个 word
				for i < len(line) && isWordChar(rune(line[i])) {
					i++
				}
			}
		}
	}

	return Pos{Line: cursor.Line, Col: i}, nil
}

// resolveLineMotion 解析行 motion
func (r *MotionResolver) resolveLineMotion(cursor Pos, count int) (Pos, error) {
	newLine := cursor.Line + count
	if newLine < 0 {
		newLine = 0
	}
	// 这里不处理超过文件范围的情况，由上层处理

	return Pos{Line: newLine, Col: cursor.Col}, nil
}

// resolveGotoMotion 解析跳转 motion
func (r *MotionResolver) resolveGotoMotion(cursor Pos, count int) (Pos, error) {
	// 对于 G (跳转到底部) 和 gg (跳转到顶部)
	// 这里简化处理，实际实现需要知道总行数
	if count == -1 { // 特殊标记表示跳转到底部
		// 假设跳转到最后一行
		return Pos{Line: 999999, Col: 0}, nil // 实际实现需要获取总行数
	}
	
	return cursor, nil
}

// normalize 规范化范围
func (r *MotionResolver) normalize(a, b Pos) Range {
	if r.before(b, a) {
		return Range{Start: b, End: a}
	}
	return Range{Start: a, End: b}
}

// before 判断 a 是否在 b 之前
func (r *MotionResolver) before(a, b Pos) bool {
	if a.Line != b.Line {
		return a.Line < b.Line
	}
	return a.Col < b.Col
}

// isWordChar 判断是否为单词字符
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}