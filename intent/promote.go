package intent

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
	}

	return i
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
		switch motion.Direction {
		case DirectionLeft:
			motionStr = "word_backward"
		case DirectionRight:
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
		switch motion.Direction {
		case DirectionLeft:
			motionStr = "goto_line_start"
		case DirectionRight:
			motionStr = "goto_line_end"
		default:
			// gg or G
			if motion.Count > 1 {
				motionStr = "goto_line" // Not fully supported yet?
			} else {
				// Assuming if no count and goto, it is gg/G?
				// Grammar sets MotionGoto but doesn't set direction for gg/G
				// TmuxPhysical expects start_of_file/end_of_file
				// For now let's leave it as is or handle in next step
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
