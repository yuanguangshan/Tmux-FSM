package main

import (
	"fmt"
	"strconv"
	"strings"
)

// PendingOp 表示待处理的操作
type PendingOp int

const (
	OpNone PendingOp = iota
	OpDelete
	OpChange
	OpYank
)

// FSM 结构体用于管理状态机
type FSM struct {
	pending PendingOp
}

// processKeyToIntent 将按键转换为 Intent（阶段 1：新增的语义层）
// 这是从 string-based action 到 Intent-based 的过渡函数
func processKeyToIntent(state *FSMState, key string) Intent {
	// 尝试使用 Native Intent Builder 处理特定命令 (dw, cw, dd)
	intent := processKeyWithNativeBuilder(state, key)
	if intent.Kind != IntentNone {
		return intent
	}

	// 对于不支持的命令，仍然使用 legacy bridge
	action := processKeyLegacy(state, key)

	// 如果没有 action，返回空 Intent
	if action == "" {
		return Intent{Kind: IntentNone}
	}

	// 获取当前光标位置
	// 注意：这是临时解决方案，直接从 tmux 获取光标位置
	// 真正的快照感知实现应该在 Resolver 阶段从快照中获取位置信息
	var cursorPos [2]int // [col, row]
	if state.PaneID != "" {
		cursorPos = GetTmuxCursorPos(state.PaneID)
	} else {
		// 如果没有 pane ID，使用状态中的光标位置
		cursorPos[0] = state.Cursor.Col
		cursorPos[1] = state.Cursor.Row
	}

	// 将 action string 转换为 Intent
	// 注意：这是一个临时的反向转换，最终会被移除
	// 使用新的函数，传入行信息以解决 projection conflict check failed: missing LineID 的问题
	// 重要：当前实现使用临时 LineID 策略，真正的快照感知实现需要在 Resolver 阶段完成
	return actionStringToIntentWithLineInfo(action, state.Count, state.PaneID, "", cursorPos[1], cursorPos[0])
}

// processKeyWithNativeBuilder 使用 Native Intent Builder 处理特定命令
func processKeyWithNativeBuilder(state *FSMState, key string) Intent {
	// 创建 IntentBuilder 实例
	builder := NewIntentBuilder(state.PaneID)

	// 处理数字计数
	if val, err := strconv.Atoi(key); err == nil && (val > 0 || state.Count > 0) {
		// 在 Native Intent 模式下，我们直接更新状态中的计数
		state.Count = state.Count*10 + val
		return Intent{Kind: IntentNone}
	}

	// 根据当前模式处理按键
	switch state.Mode {
	case "NORMAL":
		return handleNormalWithNativeBuilder(state, key, builder)
	case "OPERATOR_PENDING":
		return handleOperatorPendingWithNativeBuilder(state, key, builder)
	}

	// 对于其他模式，返回 IntentNone 以使用 legacy bridge
	return Intent{Kind: IntentNone}
}

// handleNormalWithNativeBuilder 处理 NORMAL 模式下的按键，使用 Native Intent Builder
func handleNormalWithNativeBuilder(state *FSMState, key string, builder *IntentBuilder) Intent {
	switch key {
	case "d":
		// 设置待处理操作为删除
		state.Operator = "delete"
		state.Mode = "OPERATOR_PENDING"
		return Intent{Kind: IntentNone}
	case "c":
		// 设置待处理操作为修改
		state.Operator = "change"
		state.Mode = "OPERATOR_PENDING"
		return Intent{Kind: IntentNone}
	case "y":
		// 设置待处理操作为复制
		state.Operator = "yank"
		state.Mode = "OPERATOR_PENDING"
		return Intent{Kind: IntentNone}
	case "x":
		// 直接删除右侧字符
		return builder.Delete(SemanticTarget{Kind: TargetChar, Direction: "right"}, 1)
	case "X":
		// 直接删除左侧字符
		return builder.Delete(SemanticTarget{Kind: TargetChar, Direction: "left"}, 1)
	case "D":
		// 删除到行尾
		return builder.Delete(SemanticTarget{Kind: TargetLine, Scope: "end"}, 1)
	case "C":
		// 修改到行尾
		return builder.Change(SemanticTarget{Kind: TargetLine, Scope: "end"}, 1)
	case "S":
		// 修改整行
		return builder.Change(SemanticTarget{Kind: TargetLine, Scope: "whole"}, 1)
	case "u":
		// 撤销
		return builder.Undo()
	case "C-r":
		// 重做
		return builder.Redo()
	case "n":
		// 搜索下一个
		return builder.Search(SemanticTarget{Kind: TargetSearch, Direction: "next"})
	case "N":
		// 搜索上一个
		return builder.Search(SemanticTarget{Kind: TargetSearch, Direction: "prev"})
	}

	// 基础移动命令
	motions := map[string]SemanticTarget{
		"h": {Kind: TargetChar, Direction: "left"},
		"j": {Kind: TargetPosition, Direction: "down"},
		"k": {Kind: TargetPosition, Direction: "up"},
		"l": {Kind: TargetChar, Direction: "right"},
		"w": {Kind: TargetWord, Direction: "forward"},
		"b": {Kind: TargetWord, Direction: "backward"},
		"e": {Kind: TargetWord, Scope: "end"},
		"0": {Kind: TargetLine, Scope: "start"},
		"$": {Kind: TargetLine, Scope: "end"},
		"G": {Kind: TargetFile, Scope: "end"},
		"^": {Kind: TargetLine, Scope: "start"},
		"C-b": {Kind: TargetWord, Direction: "backward"},
		"C-f": {Kind: TargetWord, Direction: "forward"},
		"Home": {Kind: TargetLine, Scope: "start"},
		"End":  {Kind: TargetLine, Scope: "end"},
	}

	if motion, ok := motions[key]; ok {
		return builder.Move(motion, state.Count)
	}

	return Intent{Kind: IntentNone}
}

// handleOperatorPendingWithNativeBuilder 处理 OPERATOR_PENDING 模式下的按键，使用 Native Intent Builder
func handleOperatorPendingWithNativeBuilder(state *FSMState, key string, builder *IntentBuilder) Intent {
	// 处理数字计数 (允许 d2w 这种形式)
	if val, err := strconv.Atoi(key); err == nil && (val > 0 || state.Count > 0) {
		state.Count = state.Count*10 + val
		return Intent{Kind: IntentNone}
	}

	// 定义运动映射
	motions := map[string]SemanticTarget{
		"h": {Kind: TargetChar, Direction: "left"},
		"j": {Kind: TargetPosition, Direction: "down"},
		"k": {Kind: TargetPosition, Direction: "up"},
		"l": {Kind: TargetChar, Direction: "right"},
		"w": {Kind: TargetWord, Direction: "forward"},
		"b": {Kind: TargetWord, Direction: "backward"},
		"e": {Kind: TargetWord, Scope: "end"},
		"$": {Kind: TargetLine, Scope: "end"},
		"0": {Kind: TargetLine, Scope: "start"},
		"^": {Kind: TargetLine, Scope: "start"},
		"G": {Kind: TargetFile, Scope: "end"},
	}

	// 检查是否是运动命令
	if motion, ok := motions[key]; ok {
		// 根据操作符创建相应的 Intent
		var intent Intent
		count := state.Count
		if count == 0 {
			count = 1
		}

		switch state.Operator {
		case "delete":
			intent = builder.Delete(motion, count)
		case "change":
			intent = builder.Change(motion, count)
		case "yank":
			intent = builder.Yank(motion, count)
		}

		// 重置状态
		state.Mode = "NORMAL"
		state.Operator = ""
		state.Count = 0

		return intent
	}

	// 检查是否是重复操作符 (例如在 d 后再按 d)
	if key == "d" && state.Operator == "delete" {
		// 重复删除操作符意味着对整行进行操作
		intent := builder.Delete(SemanticTarget{Kind: TargetLine, Scope: "whole"}, 1)
		state.Mode = "NORMAL"
		state.Operator = ""
		return intent
	} else if key == "c" && state.Operator == "change" {
		// 重复修改操作符意味着对整行进行操作
		intent := builder.Change(SemanticTarget{Kind: TargetLine, Scope: "whole"}, 1)
		state.Mode = "NORMAL"
		state.Operator = ""
		return intent
	} else if key == "y" && state.Operator == "yank" {
		// 重复复制操作符意味着对整行进行操作
		intent := builder.Yank(SemanticTarget{Kind: TargetLine, Scope: "whole"}, 1)
		state.Mode = "NORMAL"
		state.Operator = ""
		return intent
	}

	// 检查是否进入文本对象模式 (i 或 a)
	if key == "i" || key == "a" {
		// 设置文本对象待处理状态
		state.Mode = "TEXT_OBJECT_PENDING"
		state.PendingKeys = key // 记录是 inside 还是 around
		return Intent{Kind: IntentNone}
	}

	// 如果没有匹配，取消操作符待处理状态
	state.Mode = "NORMAL"
	state.Operator = ""
	return Intent{Kind: IntentNone}
}

// processKey 保持原有签名，内部调用 processKeyToIntent
// 这确保了向后兼容性（阶段 1 的关键：行为 100% 不变）
func processKey(state *FSMState, key string) string {
	intent := processKeyToIntent(state, key)
	return intent.ToActionString()
}

// processKeyLegacy 是原来的 processKey 实现
// 重命名以保留原有逻辑
func processKeyLegacy(state *FSMState, key string) string {
	var res string
	switch state.Mode {
	case "NORMAL":
		res = handleNormal(state, key)
	case "OPERATOR_PENDING":
		res = handleOperatorPending(state, key)
	case "REGISTER_SELECT":
		res = handleRegisterSelect(state, key)
	case "VISUAL_CHAR":
		res = handleVisualChar(state, key)
	case "VISUAL_LINE":
		res = handleVisualLine(state, key)
	case "MOTION_PENDING":
		res = handleMotionPending(state, key)
	case "FIND_CHAR":
		res = handleFindChar(state, key)
	case "TEXT_OBJECT_PENDING":
		res = handleTextObjectPending(state, key)
	case "SEARCH":
		res = handleSearch(state, key)
	case "REPLACE_CHAR":
		res = handleReplaceChar(state, key)
	default:
		// 处理空字符串或未知模式，默认为 NORMAL
		state.Mode = "NORMAL"
		res = handleNormal(state, key)
	}

	// 只在需要显示等待状态时记录按键序列
	if res == "" && state.Mode != "NORMAL" && state.Mode != "SEARCH" {
		// 在非NORMAL/SEARCH模式下（如FIND_CHAR, OPERATOR_PENDING），如果还没有产生动作，则记录按键
		if key == "Escape" || key == "C-c" {
			state.PendingKeys = ""
		} else if len(key) == 1 {
			// 只记录单字符，避免记录 Escape, Enter 等特殊词
			state.PendingKeys += key
		}
	} else if res != "" {
		// 如果产生了动作，清空按键序列
		state.PendingKeys = ""
	} else if state.Mode == "NORMAL" {
		// 在NORMAL模式下，强制清空，不积累
		state.PendingKeys = ""
	}
	// 注意：SEARCH 模式下的 PendingKeys 由 handleSearch 自己管理，不再这里累加

	return res
}

func handleNormal(state *FSMState, key string) string {
	// 处理数字计数
	if val, err := strconv.Atoi(key); err == nil && (val > 0 || state.Count > 0) {
		state.Count = state.Count*10 + val
		return ""
	}

	switch key {
	case "d":
		state.Mode = "OPERATOR_PENDING"
		state.Operator = "delete"
		return ""
	case "y":
		state.Mode = "OPERATOR_PENDING"
		state.Operator = "yank"
		return ""
	case "c":
		state.Mode = "OPERATOR_PENDING"
		state.Operator = "change"
		return ""
	case "\"":
		state.Mode = "REGISTER_SELECT"
		return ""
	case "u":
		return "undo"
	case "C-r":
		return "redo"
	case "g":
		if state.Operator == "" {
			state.Mode = "MOTION_PENDING"
		}
		return ""
	case "gg": // special case, logic.go needs Buffer for this usually, but we implement simple check
		if state.Operator == "" {
			return "move_start_of_file"
		}
		return ""
	case "v":
		state.Mode = "VISUAL_CHAR"
		return "start_visual_char"
	case "V":
		state.Mode = "VISUAL_LINE"
		return "start_visual_line"
	case "f", "F", "t", "T":
		state.Mode = "FIND_CHAR"
		state.PendingKeys = key // Store which find type
		return ""
	case "x":
		return "delete_right"
	case "X":
		return "delete_left"
	case "D":
		return "delete_end_of_line"
	case "C":
		return "change_end_of_line"
	case "S":
		return "change_line"
	case "I":
		return "insert_start_of_line"
	case "A":
		return "insert_end_of_line"
	case "i":
		return "insert_before"
	case "a":
		return "insert_after"
	case "o":
		return "insert_open_below"
	case "O":
		return "insert_open_above"
	case "r":
		state.Mode = "REPLACE_CHAR"
		return ""
	case "p":
		return "paste_after"
	case "P":
		return "paste_before"
	case "~":
		return "toggle_case"
	case ".":
		return "repeat_last"
	case "/":
		state.Mode = "SEARCH"
		return ""
	case "n":
		return "search_next"
	case "N":
		return "search_prev"
	}

	// 基础移动命令
	motions := map[string]string{
		"h": "left", "j": "down", "k": "up", "l": "right",
		"w": "word_forward", "b": "word_backward", "e": "end_of_word",
		"0": "start_of_line", "$": "end_of_line",
		"G":    "end_of_file",
		"^":    "start_of_line",
		"C-b":  "word_backward", // Adding C-b as word_backward alias
		"C-f":  "word_forward",  // Adding C-f as word_forward alias
		"Home": "start_of_line",
		"End":  "end_of_line",
	}
	if m, ok := motions[key]; ok {
		res := fmt.Sprintf("move_%s", m)
		// 不在这里重置 Count，交给执行器
		return res
	}

	return ""
}

func handleOperatorPending(state *FSMState, key string) string {
	// 处理数字计数 (允许 d2w 这种形式)
	if val, err := strconv.Atoi(key); err == nil && (val > 0 || state.Count > 0) {
		state.Count = state.Count*10 + val
		return ""
	}

	// 将 operator + motion 组合
	motions := map[string]string{
		"h": "left", "j": "down", "k": "up", "l": "right",
		"w": "word_forward", "b": "word_backward", "e": "end_of_word",
		"$": "end_of_line", "0": "start_of_line", "^": "start_of_line",
		"G": "end_of_file", "gg": "start_of_file", // gg needs special handling generally, but key here is single char? No 'gg' passed as 'gg' from client?
		// Main.go client sends key by key. 'gg' logic requires MOTION_PENDING mode.
		// For simplicity, let's assume 'g' puts us in MOTION_PENDING from handleNormal.
	}

	// 清理可能的空白字符
	cleanKey := strings.TrimSpace(key)
	if m, ok := motions[cleanKey]; ok {
		op := state.Operator
		state.Mode = "NORMAL"
		state.Operator = ""
		res := fmt.Sprintf("%s_%s", op, m)
		// 不在这里重置 Count
		return res
	}

	// 检查是否是重复操作符 (例如在 d 后再按 d)
	if cleanKey == state.Operator || (state.Operator == "delete" && cleanKey == "d") || (state.Operator == "yank" && cleanKey == "y") || (state.Operator == "change" && cleanKey == "c") {
		// 重复操作符通常意味着对整行进行操作
		op := state.Operator
		state.Mode = "NORMAL"
		state.Operator = ""
		res := fmt.Sprintf("%s_line", op) // 例如: delete_line
		return res
	}

	// 检查是否进入文本对象模式 (i 或 a)
	if cleanKey == "i" || cleanKey == "a" {
		state.Mode = "TEXT_OBJECT_PENDING"
		state.PendingKeys = cleanKey // 记录是 inside 还是 around
		return ""
	}

	// 取消
	state.Mode = "NORMAL"
	state.Operator = ""
	return ""
}

func handleRegisterSelect(state *FSMState, key string) string {
	state.Mode = "NORMAL"
	state.Register = key
	return ""
}

func handleVisualChar(state *FSMState, key string) string {
	// 在字符选择模式下处理按键
	switch key {
	case "Escape", "C-c":
		state.Mode = "NORMAL"
		return "cancel_selection"
	case "v":
		// 退出字符选择模式
		state.Mode = "NORMAL"
		return "cancel_selection"
	case "V":
		// 转换为行选择模式
		state.Mode = "VISUAL_LINE"
		return "start_visual_line"
	}

	// 处理移动命令
	motions := map[string]string{
		"h": "left", "j": "down", "k": "up", "l": "right",
		"w": "word_forward", "b": "word_backward",
	}
	if m, ok := motions[key]; ok {
		res := fmt.Sprintf("move_%s", m)
		return res
	}

	// 处理操作符
	operators := map[string]string{
		"d": "delete",
		"c": "change",
		"y": "yank",
	}
	if op, ok := operators[key]; ok {
		// 在视觉模式下执行操作后返回NORMAL模式
		state.Mode = "NORMAL"
		return fmt.Sprintf("visual_%s", op)
	}

	return ""
}

func handleVisualLine(state *FSMState, key string) string {
	// 在行选择模式下处理按键
	switch key {
	case "Escape", "C-c":
		state.Mode = "NORMAL"
		return "cancel_selection"
	case "V":
		// 退出行选择模式
		state.Mode = "NORMAL"
		return "cancel_selection"
	case "v":
		// 转换为字符选择模式
		state.Mode = "VISUAL_CHAR"
		return "start_visual_char"
	}

	// 处理移动命令
	motions := map[string]string{
		"h": "left", "j": "down", "k": "up", "l": "right",
		"w": "word_forward", "b": "word_backward",
	}
	if m, ok := motions[key]; ok {
		res := fmt.Sprintf("move_%s", m)
		return res
	}

	// 处理操作符
	operators := map[string]string{
		"d": "delete",
		"c": "change",
		"y": "yank",
	}
	if op, ok := operators[key]; ok {
		// 在视觉模式下执行操作后返回NORMAL模式
		state.Mode = "NORMAL"
		return fmt.Sprintf("visual_%s", op)
	}

	return ""
}
func handleSearch(state *FSMState, key string) string {
	if key == "Enter" || key == "C-m" || key == "Return" {
		query := state.PendingKeys
		state.Mode = "NORMAL"
		state.PendingKeys = ""
		// Store query for n/N
		state.Register = query // Reuse Register for search history for now
		return "search_forward_" + query
	}
	if key == "Escape" || key == "C-c" {
		state.Mode = "NORMAL"
		state.PendingKeys = ""
		return ""
	}
	if key == "BSpace" {
		if len(state.PendingKeys) > 0 {
			state.PendingKeys = state.PendingKeys[:len(state.PendingKeys)-1]
		}
		return ""
	}

	// Add character to buffer
	if len(key) == 1 {
		state.PendingKeys += key
		return ""
	}

	// Handle special keys like Space
	if key == "Space" {
		state.PendingKeys += " "
		return ""
	}

	return ""
}

func handleTextObjectPending(state *FSMState, key string) string {
	objModifier := state.PendingKeys // 'i' or 'a'
	state.Mode = "NORMAL"
	state.PendingKeys = ""

	op := state.Operator
	state.Operator = ""

	// 支持的对象类型
	objTypes := map[string]string{
		"w": "word",
		"p": "paragraph",
		"s": "sentence",
		"(": "paren", ")": "paren", "b": "paren",
		"[": "bracket", "]": "bracket",
		"{": "brace", "}": "brace", "B": "brace",
		"\"": "quote_double", "'": "quote_single", "`": "quote_backtick",
	}

	if t, ok := objTypes[key]; ok {
		var suffix string
		if objModifier == "i" {
			suffix = "inside_" + t
		} else {
			suffix = "around_" + t
		}
		return fmt.Sprintf("%s_%s", op, suffix)
	}

	// 如果不是标准文本对象，直接返回组合
	// 例如 "iw", "aw", "ip", "ap" 等
	textObjectStr := objModifier + key
	return fmt.Sprintf("%s_%s", op, textObjectStr)
}

func handleFindChar(state *FSMState, key string) string {
	findType := state.PendingKeys
	state.Mode = "NORMAL"
	state.PendingKeys = ""

	// Action format: find_[f|F|t|T]_[char]
	// Using a special prefix for the executor to handle
	return fmt.Sprintf("find_%s_%s", findType, key)
}

func handleMotionPending(state *FSMState, key string) string {
	switch key {
	case "g":
		state.Mode = "NORMAL"
		return "move_start_of_file"
	default:
		// Reset if not a valid motion continuation
		state.Mode = "NORMAL"
		return ""
	}
}

func handleReplaceChar(state *FSMState, key string) string {
	state.Mode = "NORMAL"
	if key == "Escape" || key == "C-c" {
		return ""
	}
	return "replace_char_" + key
}
