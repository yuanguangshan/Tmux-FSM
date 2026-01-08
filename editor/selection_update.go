package editor

import "sort"

// UpdateSelections 根据已执行的操作更新选区
// 这是确定性的、可预测的选区更新算法
// 输入：当前选区列表 + 已执行的操作记录
// 输出：更新后的选区列表
func UpdateSelections(selections []Selection, ops []ResolvedOperation) []Selection {
	if len(selections) == 0 {
		return selections
	}

	// 逐条应用物理修改
	for _, op := range ops {
		switch op.Kind {
		case OpDelete:
			if op.Range != nil {
				selections = applyDelete(selections, op.Range.Start, op.Range.End)
			}

		case OpInsert:
			// 计算插入文本的长度（简化版，假设单行）
			textLen := len(op.Text)
			selections = applyInsert(selections, op.Anchor, textLen)

		// OpMove 不影响 selections
		case OpMove:
			// 移动光标不改变选区
			continue
		}
	}

	return normalizeSelections(selections)
}

// applyDelete 应用删除操作到选区
func applyDelete(sels []Selection, dStart, dEnd Cursor) []Selection {
	if len(sels) == 0 {
		return sels
	}

	result := make([]Selection, 0, len(sels))

	for _, sel := range sels {
		// 完全在删除范围之前
		if sel.End.LessThan(dStart) || sel.End.Equal(dStart) {
			result = append(result, sel)
			continue
		}

		// 完全在删除范围之后
		if (sel.Start.Row > dEnd.Row) || (sel.Start.Row == dEnd.Row && sel.Start.Col >= dEnd.Col) {
			// 向前平移
			newSel := shiftSelection(sel, dStart, dEnd)
			result = append(result, newSel)
			continue
		}

		// 与删除范围相交 - collapse 到删除起点
		result = append(result, Selection{
			Start: dStart,
			End:   dStart,
		})
	}

	return result
}

// applyInsert 应用插入操作到选区
func applyInsert(sels []Selection, insertPos Cursor, textLen int) []Selection {
	if len(sels) == 0 {
		return sels
	}

	result := make([]Selection, 0, len(sels))

	for _, sel := range sels {
		// 如果选区在插入点之前或刚好在插入点，不受影响
		if sel.End.LessThan(insertPos) {
			result = append(result, sel)
			continue
		}

		// 如果选区在插入点之后，需要向后平移
		if sel.Start.Row > insertPos.Row || (sel.Start.Row == insertPos.Row && sel.Start.Col >= insertPos.Col) {
			// 简化版：假设插入在同一行
			newSel := Selection{
				Start: Cursor{Row: sel.Start.Row, Col: sel.Start.Col + textLen},
				End:   Cursor{Row: sel.End.Row, Col: sel.End.Col + textLen},
			}
			result = append(result, newSel)
			continue
		}

		// 插入点在选区内部 - 扩展选区
		result = append(result, Selection{
			Start: sel.Start,
			End:   Cursor{Row: sel.End.Row, Col: sel.End.Col + textLen},
		})
	}

	return result
}

// shiftSelection 平移选区（用于删除后的调整）
func shiftSelection(sel Selection, dStart, dEnd Cursor) Selection {
	// 简化版：假设单行删除
	if dStart.Row == dEnd.Row {
		delta := dEnd.Col - dStart.Col
		return Selection{
			Start: Cursor{Row: sel.Start.Row, Col: sel.Start.Col - delta},
			End:   Cursor{Row: sel.End.Row, Col: sel.End.Col - delta},
		}
	}

	// 多行删除的情况（更复杂，暂时简化处理）
	return sel
}

// normalizeSelections 规范化选区列表
// 1. 确保 Start <= End
// 2. 按 Start 排序
// 3. 合并重叠的选区
func normalizeSelections(sels []Selection) []Selection {
	if len(sels) == 0 {
		return sels
	}

	// 1. 确保每个选区的 Start <= End
	for i := range sels {
		if sels[i].End.LessThan(sels[i].Start) {
			sels[i].Start, sels[i].End = sels[i].End, sels[i].Start
		}
	}

	// 2. 按 Start 排序
	sort.Slice(sels, func(i, j int) bool {
		return sels[i].Start.LessThan(sels[j].Start)
	})

	// 3. 合并重叠的选区
	result := make([]Selection, 0, len(sels))
	current := sels[0]

	for i := 1; i < len(sels); i++ {
		next := sels[i]

		// 如果当前选区与下一个选区重叠或相邻
		if !current.End.LessThan(next.Start) {
			// 合并
			if next.End.LessThan(current.End) {
				// next 完全包含在 current 中
				continue
			}
			current.End = next.End
		} else {
			// 不重叠，保存当前选区，开始新的选区
			result = append(result, current)
			current = next
		}
	}

	// 添加最后一个选区
	result = append(result, current)

	return result
}

// Equal 判断两个 Cursor 是否相等
func (c Cursor) Equal(other Cursor) bool {
	return c.Row == other.Row && c.Col == other.Col
}
