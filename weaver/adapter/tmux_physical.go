package adapter

import (
	"fmt"
	"os/exec"
	"strings"
)

// NOTE:
// This file is a verbatim copy of physical execution logic from execute.go.
// Phase 3 rule:
//   - NO behavior change
//   - NO refactor
//   - NO abstraction
//   - exec.Command is used directly
//
// This file exists to allow Weaver Projection to execute shell actions
// while keeping legacy execute.go untouched as a control group.
//
// Allowed changes:
//   - package name
//   - imports adjustment
//   - renamed private helpers (if collision)
//   - exported functions for Layout (TmuxProjection to use)
//
// This file MUST NOT be modified until Phase 6.

// PerformPhysicalInsert 插入操作
func PerformPhysicalInsert(motion, targetPane string) {
	switch motion {
	case "after":
		exec.Command("tmux", "send-keys", "-t", targetPane, "Right").Run()
	case "start_of_line":
		exec.Command("tmux", "send-keys", "-t", targetPane, "Home").Run()
	case "end_of_line":
		exec.Command("tmux", "send-keys", "-t", targetPane, "End").Run()
	case "open_below":
		exec.Command("tmux", "send-keys", "-t", targetPane, "End", "Enter").Run()
	case "open_above":
		exec.Command("tmux", "send-keys", "-t", targetPane, "Home", "Enter", "Up").Run()
	}
}

// PerformPhysicalPaste 粘贴操作
func PerformPhysicalPaste(motion, targetPane string) {
	if motion == "after" {
		exec.Command("tmux", "send-keys", "-t", targetPane, "Right").Run()
	}
	exec.Command("tmux", "paste-buffer", "-t", targetPane).Run()
}

// PerformPhysicalReplace 替换字符
func PerformPhysicalReplace(char, targetPane string) {
	exec.Command("tmux", "send-keys", "-t", targetPane, "Delete", char).Run()
}

// PerformPhysicalToggleCase 切换大小写
func PerformPhysicalToggleCase(targetPane string) {
	// Captures the char under cursor, toggles it, and replaces it.
	pos := TmuxGetCursorPos(targetPane) // Use helper from tmux_utils.go
	out, _ := exec.Command("tmux", "capture-pane", "-p", "-t", targetPane, "-S", fmt.Sprint(pos[1]), "-E", fmt.Sprint(pos[1])).Output()
	line := string(out)
	if pos[0] < len(line) {
		char := line[pos[0]]
		newChar := char
		if char >= 'a' && char <= 'z' {
			newChar = char - 'a' + 'A'
		} else if char >= 'A' && char <= 'Z' {
			newChar = char - 'A' + 'a'
		}
		if newChar != char {
			exec.Command("tmux", "send-keys", "-t", targetPane, "Delete", string(newChar)).Run()
		}
	}
}

// PerformPhysicalMove 移动操作
func PerformPhysicalMove(motion string, count int, targetPane string) {
	cStr := fmt.Sprint(count)
	switch motion {
	case "up":
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", cStr, "Up").Run()
	case "down":
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", cStr, "Down").Run()
	case "left":
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", cStr, "Left").Run()
	case "right":
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", cStr, "Right").Run()
	case "start_of_line": // 0
		exec.Command("tmux", "send-keys", "-t", targetPane, "C-a").Run()
	case "end_of_line": // $
		exec.Command("tmux", "send-keys", "-t", targetPane, "C-e").Run()
	case "word_forward": // w
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", cStr, "M-f").Run()
	case "word_backward": // b
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", cStr, "M-b").Run()
	case "end_of_word": // e
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", cStr, "M-f").Run()
	case "start_of_file": // gg
		exec.Command("tmux", "send-keys", "-t", targetPane, "Home").Run()
	case "end_of_file": // G
		exec.Command("tmux", "send-keys", "-t", targetPane, "End").Run()
	}
}

// PerformExecuteSearch 执行搜索
func PerformExecuteSearch(query string, targetPane string) {
	// 1. Enter copy mode if not in it
	// 2. Start search-forward
	exec.Command("tmux", "copy-mode", "-t", targetPane).Run()
	exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "search-forward", query).Run()
}

// PerformPhysicalDelete 删除操作
func PerformPhysicalDelete(motion string, targetPane string) {
	// 首先取消任何现有的选择
	exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "cancel").Run()

	switch motion {
	case "start_of_line": // d0
		// Robust implementation: Get cursor X position and backspace that many times
		pos := TmuxGetCursorPos(targetPane) // Use helper
		cursorX := pos[0]
		if cursorX > 0 {
			exec.Command("tmux", "send-keys", "-t", targetPane, "-N", fmt.Sprint(cursorX), "BSpace").Run()
		}

	case "end_of_line": // d$
		// C-k: Kill to end of line
		exec.Command("tmux", "send-keys", "-t", targetPane, "C-k").Run()

	case "word_forward", "inside_word", "around_word": // dw
		// Simple and robust: most shells bind M-d to delete-word-forward
		exec.Command("tmux", "send-keys", "-t", targetPane, "M-d").Run()

	case "word_backward": // db
		// C-w: Unix word rubout (backward)
		exec.Command("tmux", "send-keys", "-t", targetPane, "C-w").Run()

	case "right": // x / dl
		exec.Command("tmux", "send-keys", "-t", targetPane, "Delete").Run()

	case "left": // dh
		exec.Command("tmux", "send-keys", "-t", targetPane, "BSpace").Run()

	case "line": // dd
		// Delete line: Go to start (C-a) then Kill line (C-k), then Delete (consume newline if possible)
		exec.Command("tmux", "send-keys", "-t", targetPane, "C-a", "C-k", "Delete").Run()

	default:
		// Default fallback
		exec.Command("tmux", "send-keys", "-t", targetPane, "M-d").Run()
	}
}

// PerformPhysicalTextObject 文本对象操作
func PerformPhysicalTextObject(op, motion, targetPane string) {
	// 1. Capture current line
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", targetPane, "#{pane_cursor_x}").Output()
	var cursorX int
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &cursorX)

	out, _ = exec.Command("tmux", "capture-pane", "-p", "-t", targetPane, "-J").Output()
	lines := strings.Split(string(out), "\n")
	var currentLine string
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			currentLine = lines[i]
			break
		}
	}
	if currentLine == "" {
		return
	}

	start, end := -1, -1

	if strings.Contains(motion, "word") {
		start, end = findWordRange(currentLine, cursorX, strings.Contains(motion, "around_"))
	} else if strings.Contains(motion, "quote_") {
		quoteChar := "\""
		if strings.Contains(motion, "single") {
			quoteChar = "'"
		}
		start, end = findQuoteRange(currentLine, cursorX, quoteChar, strings.Contains(motion, "around_"))
	} else if strings.Contains(motion, "paren") || strings.Contains(motion, "bracket") || strings.Contains(motion, "brace") {
		start, end = findBracketRange(currentLine, cursorX, motion, strings.Contains(motion, "around_"))
	}

	if start != -1 && end != -1 {
		if op == "delete" || op == "change" {
			TmuxJumpTo(end, -1, targetPane) // Use helper
			dist := end - start + 1
			exec.Command("tmux", "send-keys", "-t", targetPane, "-N", fmt.Sprint(dist), "BSpace").Run()
			if op == "change" {
				exec.Command("tmux", "send-keys", "-t", targetPane, "i").Run()
			}
		} else if op == "yank" {
			TmuxJumpTo(start, -1, targetPane) // Use helper
			exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "begin-selection").Run()
			TmuxJumpTo(end, -1, targetPane) // Use helper
			exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "copy-pipe-and-cancel", "tmux save-buffer -").Run()
		}
	}
}

// PerformPhysicalFind 字符查找
func PerformPhysicalFind(fType, char string, count int, targetPane string) {
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", targetPane, "#{pane_cursor_x}").Output()
	var cursorX int
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &cursorX)

	out, _ = exec.Command("tmux", "capture-pane", "-p", "-t", targetPane, "-J").Output()
	lines := strings.Split(string(out), "\n")

	var currentLine string
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			currentLine = lines[i]
			break
		}
	}

	if currentLine == "" {
		return
	}

	targetX := -1
	foundCount := 0

	switch fType {
	case "f":
		for x := cursorX + 1; x < len(currentLine); x++ {
			if string(currentLine[x]) == char {
				foundCount++
				if foundCount == count {
					targetX = x
					break
				}
			}
		}
	case "F":
		for x := cursorX - 1; x >= 0; x-- {
			if string(currentLine[x]) == char {
				foundCount++
				if foundCount == count {
					targetX = x
					break
				}
			}
		}
	case "t":
		for x := cursorX + 1; x < len(currentLine); x++ {
			if string(currentLine[x]) == char {
				foundCount++
				if foundCount == count {
					targetX = x - 1
					break
				}
			}
		}
	case "T":
		for x := cursorX - 1; x >= 0; x-- {
			if string(currentLine[x]) == char {
				foundCount++
				if foundCount == count {
					targetX = x + 1
					break
				}
			}
		}
	}

	if targetX != -1 {
		TmuxJumpTo(targetX, -1, targetPane) // Use helper
	}
}

// HandleVisualAction 视觉模式操作
func HandleVisualAction(action string, stateCount int, targetPane string) {
	parts := strings.Split(action, "_")
	if len(parts) < 2 {
		return
	}

	op := parts[1]

	if TmuxIsVimPane(targetPane) { // Use helper
		vimOp := ""
		switch op {
		case "delete":
			vimOp = "d"
		case "yank":
			vimOp = "y"
		case "change":
			vimOp = "c"
		}

		if vimOp != "" {
			exec.Command("tmux", "send-keys", "-t", targetPane, vimOp).Run()
		}
	} else {
		if op == "yank" {
			exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "copy-pipe-and-cancel", "tmux save-buffer -").Run()
		} else if op == "delete" || op == "change" {
			exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "copy-pipe-and-cancel", "tmux save-buffer -").Run()
			if op == "change" {
				exec.Command("tmux", "send-keys", "-t", targetPane, "i").Run()
			}
		}
	}
}

// ExitFSM 退出 FSM
func ExitFSM(targetPane string) {
	exec.Command("tmux", "set", "-g", "@fsm_active", "false").Run()
	exec.Command("tmux", "set", "-g", "@fsm_state", "").Run()
	exec.Command("tmux", "set", "-g", "@fsm_keys", "").Run()
	exec.Command("tmux", "switch-client", "-T", "root").Run()
	exec.Command("tmux", "refresh-client", "-S").Run()
}

// Private helper functions for text objects (copied verbatim)

func findWordRange(line string, x int, around bool) (int, int) {
	if x >= len(line) {
		return -1, -1
	}

	isWordChar := func(c byte) bool {
		return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
	}

	start := x
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}
	end := x
	for end < len(line)-1 && isWordChar(line[end+1]) {
		end++
	}

	if around {
		if end < len(line)-1 && line[end+1] == ' ' {
			end++
		} else if start > 0 && line[start-1] == ' ' {
			start--
		}
	}

	return start, end
}

func findQuoteRange(line string, x int, quote string, around bool) (int, int) {
	first := strings.LastIndex(line[:x+1], quote)
	if first == -1 {
		first = strings.Index(line[x:], quote)
		if first != -1 {
			first += x
		}
	}
	if first == -1 {
		return -1, -1
	}

	second := strings.Index(line[first+1:], quote)
	if second == -1 {
		return -1, -1
	}
	second += first + 1

	if around {
		return first, second
	}
	return first + 1, second - 1
}

func findBracketRange(line string, x int, motion string, around bool) (int, int) {
	opening, closing := "", ""
	if strings.Contains(motion, "paren") {
		opening, closing = "(", ")"
	} else if strings.Contains(motion, "bracket") {
		opening, closing = "[", "]"
	} else if strings.Contains(motion, "brace") {
		opening, closing = "{", "}"
	}

	start := -1
	balance := 0
	for i := x; i >= 0; i-- {
		c := string(line[i])
		if c == closing {
			balance--
		} else if c == opening {
			balance++
			if balance == 1 {
				start = i
				break
			}
		}
	}
	if start == -1 {
		return -1, -1
	}

	end := -1
	balance = 1
	for i := start + 1; i < len(line); i++ {
		c := string(line[i])
		if c == opening {
			balance++
		} else if c == closing {
			balance--
			if balance == 0 {
				end = i
				break
			}
		}
	}
	if end == -1 {
		return -1, -1
	}

	if around {
		return start, end
	}
	return start + 1, end - 1
}

// PerformPhysicalRawInsert 物理插入原始文本
func PerformPhysicalRawInsert(text, targetPane string) {
	// 使用管道 load-buffer 是最健壮的，彻底避免 '?' 乱码问题
	cmd := exec.Command("tmux", "load-buffer", "-")
	cmd.Stdin = strings.NewReader(text)
	cmd.Run()

	// 确保粘贴到目标
	exec.Command("tmux", "paste-buffer", "-t", targetPane).Run()
}
