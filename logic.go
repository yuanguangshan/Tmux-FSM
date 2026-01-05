package main

import (
	"fmt"
	"strconv"
	"strings"
)

func processKey(state *FSMState, key string) string {
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
		"G":   "end_of_file",
		"^":   "start_of_line",
		"C-b": "word_backward", // Adding C-b as word_backward alias
		"C-f": "word_forward",  // Adding C-f as word_forward alias
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

	// 目前支持的对象类型
	objTypes := map[string]string{
		"w": "word",
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

	return ""
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
