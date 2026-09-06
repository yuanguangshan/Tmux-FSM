package intent

import "log"

// Promote 是 GrammarIntent → Intent 的唯一合法通道
// Grammar 不允许直接构造 Intent
func Promote(g *GrammarIntent) *Intent {
	if g == nil {
		return nil
	}

	// 初始化 Meta 映射
	meta := make(map[string]interface{})

	// 如果 GrammarIntent 包含 Motion，将其转换为遗留的 Meta 字段
	if g.Motion != nil {
		// 将强类型的 Motion 转换为遗留的 Meta 字段
		populateLegacyMotionMeta(meta, g.Motion)
	}

	i := &Intent{
		Kind:   g.Kind,
		Count:  g.Count,
		Motion: g.Motion,
		Meta:   meta, // 添加初始化的 Meta 映射
		// 对于基本的移动意图，允许部分匹配（模糊解析）
		AllowPartial: g.Kind == IntentMove,
	}

	// Operator 提升（强类型）
	if g.Op != nil {
		i.Operator = g.Op
		// M4.5：操作名进入 meta——weaver 投影的 change 分支据此
		// 在删除后进入插入模式
		switch *g.Op {
		case OpDelete:
			meta["operation"] = "delete"
		case OpYank:
			meta["operation"] = "yank"
		case OpChange:
			meta["operation"] = "change"
		}
	}

	return i
}

// textObjectSpec 生成物理层可匹配的规范名：
// inner_X / around_X（X ∈ word/paren/bracket/brace/quote_double/quote_single/quote_backtick）
func textObjectSpec(scope TextObjectScope, object TextObjectKind) string {
	prefix := "inner_"
	if scope == Around {
		prefix = "around_"
	}
	switch object {
	case Word:
		return prefix + "word"
	case Paren:
		return prefix + "paren"
	case Bracket:
		return prefix + "bracket"
	case Brace:
		return prefix + "brace"
	case QuoteSingle:
		return prefix + "quote_single"
	case QuoteDouble:
		return prefix + "quote_double"
	case Backtick:
		return prefix + "quote_backtick"
	}
	return prefix + "unknown"
}

// populateLegacyMotionMeta 将强类型的 Motion 结构转换为遗留的 Meta 字段
// 这是桥接新架构和现有实现的必要步骤
func populateLegacyMotionMeta(meta map[string]interface{}, motion *Motion) {
	if motion == nil || meta == nil {
		return
	}

	// 根据 Motion.Kind 和 Direction 生成对应的运动字符串
	var motionStr string
	switch motion.Kind {
	case MotionChar:
		switch motion.Direction {
		case DirectionLeft:
			motionStr = "left"
		case DirectionRight:
			motionStr = "right"
		case DirectionUp:
			motionStr = "up"
		case DirectionDown:
			motionStr = "down"
		}
	case MotionWord:
		// M1.4：补 DirectionNone 兜底并显式告警（此前 None → meta 缺失 →
		// 物理层静默退化 default 发 M-d，语义无声漂移）
		switch motion.Direction {
		case DirectionLeft:
			motionStr = "word_backward"
		default:
			if motion.Direction != DirectionRight {
				log.Printf("[PROMOTE][WARN] MotionWord direction=%v 未映射，按向前处理", motion.Direction)
			}
			motionStr = "word_forward"
		}
	case MotionLine:
		switch motion.Direction {
		case DirectionUp:
			motionStr = "line_up"
		case DirectionDown:
			motionStr = "line_down"
		default:
			motionStr = "line"
		}
	case MotionGoto:
		// M1.4：grammar 现在为 G/gg 设置 Down/Up 方向，精准映射文件尾/头
		switch motion.Direction {
		case DirectionDown:
			motionStr = "end_of_file"
		case DirectionUp:
			motionStr = "start_of_file"
		default:
			if motion.Count > 1 {
				motionStr = "goto_line" // Ngg/N G 尚未完全支持
			} else {
				motionStr = "start_of_file" // 无方向时按 gg（文件头）处理
			}
		}
	case MotionFind:
		if motion.Find != nil {
			if motion.Find.Direction == FindForward {
				if motion.Find.Till {
					motionStr = "find_char_before_forward"
				} else {
					motionStr = "find_char_forward"
				}
			} else {
				if motion.Find.Till {
					motionStr = "find_char_before_backward"
				} else {
					motionStr = "find_char_backward"
				}
			}
		}
	case MotionRange:
		if motion.Range != nil {
			switch motion.Range.Kind {
			case RangeLineStart:
				motionStr = "goto_line_start"
			case RangeLineEnd:
				motionStr = "goto_line_end"
			case RangeTextObject:
				// M4.5：文本对象的元数据桥接——物理层
				// PerformPhysicalTextObject 按 Contains 匹配这些 canonical 名
				if to := motion.Range.TextObject; to != nil {
					meta["text_object"] = textObjectSpec(to.Scope, to.Object)
				}
			}
		}
	}

	// 如果生成了运动字符串，将其添加到 Meta 中
	if motionStr != "" {
		meta["motion"] = motionStr
	}

	// 添加计数信息
	if motion.Count > 1 {
		meta["count"] = motion.Count
	}
}
