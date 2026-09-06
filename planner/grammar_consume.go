package planner

// M4-core 可用性修复的语法层门控（2026-09-07）：
//
// FSM NAV 模式会捕获所有按键（key-table + Any 兜底）。此前未绑定键
// 被 DecisionLegacy 静默吞掉——用户在导航模式下**无法输入任何文本**
// （敲 git、敲路径、敲引号全部蒸发），这是"不可用"的核心症状。
//
// 修复后的按键三分法（kernel.Decide 顺序）：
//   1. keymap 声明了 action 的键 → FSM 动作（exit/prompt/goto_*）
//   2. Grammar 能消费的键   → 语法层（motion/操作符/文本对象）
//   3. 其余可打印单字符     → 字面透传给目标 pane（用户自由输入）
//
// WillConsume 回答"Grammar 是否需要这个键"，供 kernel 判定走 2 还是 3。

// WillConsume 报告该按键是否会被语法层消费。
// 序列进行中（pendingOp/pendingMotion/textObj 任一挂起）时，
// 任何按键都属于序列的一部分（等待目标字符、文本对象、motion）。
func (g *Grammar) WillConsume(key string) bool {
	// 序列进行中：一切按键都是候选目标/延续
	if g.pendingOp != nil || g.pendingMotion != nil || g.textObj != TOPNone {
		return true
	}

	// 已知语法键（与 keymap 委托清单一致）
	switch key {
	case "h", "j", "k", "l", "w", "b", "e",
		"0", "$", "^", "G", "gg",
		"d", "y", "c", "u",
		"f", "F", "t", "T":
		return true
	}

	// 数字：count 累加
	if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
		return true
	}

	// C-r（undo 重做）等控制键由 keymap 声明，不在语法判定内
	return false
}
