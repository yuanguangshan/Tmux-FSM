// ❗LEGACY PHYSICAL REFERENCE
// This file defines the canonical physical behavior.
// Any change here MUST be mirrored in weaver/adapter/tmux_physical.go.

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
)

type Executor interface {
	CanExecute(f Fact) bool
	Execute(f Fact) error
}

type ResolveResult int

const (
	ResolveExact ResolveResult = iota
	ResolveFuzzy
	ResolveFail
)

type ResolvedAnchor struct {
	Row    int
	Result ResolveResult
}

func ResolveAnchor(a Anchor) (ResolvedAnchor, error) {
	// Axiom 3: Exactness Preference - Always try Exact first
	line := captureLine(a.PaneID, a.LineHint)
	if hashLine(line) == a.LineHash {
		return ResolvedAnchor{Row: a.LineHint, Result: ResolveExact}, nil
	}

	// Axiom 6: Permitted Fuzzy Conditions - Only try Fuzzy in narrow window
	window := 5
	for i := 1; i <= window; i++ {
		// Check below
		rowBelow := a.LineHint + i
		if hashLine(captureLine(a.PaneID, rowBelow)) == a.LineHash {
			return ResolvedAnchor{Row: rowBelow, Result: ResolveFuzzy}, nil
		}
		// Check above
		rowAbove := a.LineHint - i
		if rowAbove >= 0 && hashLine(captureLine(a.PaneID, rowAbove)) == a.LineHash {
			return ResolvedAnchor{Row: rowAbove, Result: ResolveFuzzy}, nil
		}
	}

	// Axiom 4: Mandatory Failure Conditions - Anchor not found in window
	return ResolvedAnchor{Result: ResolveFail}, fmt.Errorf("anchor invalid")
}

type ShellExecutor struct{}

func (s *ShellExecutor) CanExecute(f Fact) bool {
	return true // Shell is the fallback
}

func (s *ShellExecutor) Execute(f Fact) error {
	targetPane := f.Target.Anchor.PaneID
	if targetPane == "" {
		targetPane = "{current}"
	}

	switch f.Kind {
	case "insert":
		// Resolve anchor and jump
		jumpTo(f.Target.StartOffset, f.Target.Anchor.LineHint, targetPane)
		exec.Command("tmux", "send-keys", "-t", targetPane, f.Target.Text).Run()
	case "delete":
		jumpTo(f.Target.EndOffset-1, f.Target.Anchor.LineHint, targetPane)
		dist := f.Target.EndOffset - f.Target.StartOffset
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", fmt.Sprint(dist), "BSpace").Run()
	case "replace":
		newText, _ := f.Meta["new_text"].(string)
		// Delete old, insert new
		jumpTo(f.Target.EndOffset-1, f.Target.Anchor.LineHint, targetPane)
		dist := f.Target.EndOffset - f.Target.StartOffset
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", fmt.Sprint(dist), "BSpace").Run()
		exec.Command("tmux", "send-keys", "-t", targetPane, newText).Run()
	}
	return nil
}

type VimExecutor struct{}

func (v *VimExecutor) CanExecute(f Fact) bool {
	return isVimPane(f.Target.Anchor.PaneID)
}

func (v *VimExecutor) Execute(f Fact) error {
	targetPane := f.Target.Anchor.PaneID
	if targetPane == "" {
		targetPane = "{current}"
	}

	// Resolve target location if possible
	// For Vim, we might want to jump to the location first
	jumpTo(f.Target.StartOffset, f.Target.Anchor.LineHint, targetPane)

	switch f.Kind {
	case "insert":
		// Enter insert mode, type text, return to normal
		exec.Command("tmux", "send-keys", "-t", targetPane, "i", f.Target.Text, "Escape").Run()
	case "delete":
		dist := f.Target.EndOffset - f.Target.StartOffset
		exec.Command("tmux", "send-keys", "-t", targetPane, fmt.Sprintf("%dl", dist), "Escape").Run() // Simple delete logic for Vim
	case "replace":
		newText, _ := f.Meta["new_text"].(string)
		dist := f.Target.EndOffset - f.Target.StartOffset
		exec.Command("tmux", "send-keys", "-t", targetPane, fmt.Sprintf("%dc", dist), newText, "Escape").Run()
	case "undo":
		exec.Command("tmux", "send-keys", "-t", targetPane, "u").Run()
	case "redo":
		exec.Command("tmux", "send-keys", "-t", targetPane, "C-r").Run()
	}
	return nil
}

var executors = []Executor{
	&VimExecutor{},
	&ShellExecutor{},
}

func executeFact(f Fact) error {
	// --- [ABI: Side Effect Projection] ---
	// The verdict is finalized as 'Applied'. The kernel projects the fact onto the physical TTY.
	for _, ex := range executors {
		if ex.CanExecute(f) {
			return ex.Execute(f)
		}
	}
	return fmt.Errorf("no executor for fact")
}

func executeAction(action string, state *FSMState, targetPane string, clientName string) {
	// --- [ABI: Verdict Deliberation Starts] ---
	// The kernel evaluates the intent against the current world state.
	if action == "" {
		return
	}
	// Default to current if empty (though should be provided)
	if targetPane == "" {
		targetPane = "{current}"
	}

	// 1. 处理特殊内核动作：Undo / Redo
	// [Phase 9] Dispatch to Weaver as single source of truth
	if action == "undo" {
		// Create undo intent and dispatch to Weaver
		undoIntent := Intent{
			Kind:   IntentUndo,
			PaneID: targetPane,
		}
		ProcessIntentGlobal(undoIntent)
		return
	}
	if action == "redo" {
		// Create redo intent and dispatch to Weaver
		redoIntent := Intent{
			Kind:   IntentRedo,
			PaneID: targetPane,
		}
		ProcessIntentGlobal(redoIntent)
		return
	}

	if action == "search_next" {
		exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "search-again").Run()
		return
	}
	if action == "search_prev" {
		exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "search-reverse").Run()
		return
	}
	if strings.HasPrefix(action, "search_forward_") {
		query := strings.TrimPrefix(action, "search_forward_")
		executeSearch(query, targetPane)
		return
	}

	// 2. 处理VISUAL模式相关动作
	if action == "start_visual_char" {
		if isVimPane(targetPane) {
			exec.Command("tmux", "send-keys", "-t", targetPane, "v").Run()
		} else {
			exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "begin-selection").Run()
		}
		return
	}
	if action == "start_visual_line" {
		if isVimPane(targetPane) {
			exec.Command("tmux", "send-keys", "-t", targetPane, "V").Run()
		} else {
			exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "select-line").Run()
		}
		return
	}
	if action == "cancel_selection" {
		if isVimPane(targetPane) {
			exec.Command("tmux", "send-keys", "-t", targetPane, "Escape").Run()
		} else {
			exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "clear-selection").Run()
		}
		return
	}
	if strings.HasPrefix(action, "visual_") {
		// 处理视觉模式下的操作 (如 visual_delete, visual_yank, visual_change)
		handleVisualAction(action, state, targetPane)
		return
	}

	// 3. 环境探测：Vim vs Shell
	if isVimPane(targetPane) {
		executeVimAction(action, state, targetPane)
	} else {
		executeShellAction(action, state, targetPane)
	}
}

func isVimPane(targetPane string) bool {
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", targetPane, "#{pane_current_command}").Output()
	cmd := strings.TrimSpace(string(out))
	return cmd == "vim" || cmd == "nvim" || cmd == "vi"
}

func executeShellAction(action string, state *FSMState, targetPane string) {
	parts := strings.Split(action, "_")
	if len(parts) < 1 {
		return
	}

	op := parts[0]
	count := state.Count
	if count <= 0 {
		count = 1
	}

	// 1. 处理特殊单一动词
	if op == "insert" {
		motion := strings.Join(parts[1:], "_")
		performPhysicalInsert(motion, targetPane)
		exitFSM(targetPane)
		return
	}
	if op == "paste" {
		motion := strings.Join(parts[1:], "_")
		for i := 0; i < count; i++ {
			performPhysicalPaste(motion, targetPane)
		}
		return
	}
	if op == "toggle" { // toggle_case
		for i := 0; i < count; i++ {
			performPhysicalToggleCase(targetPane)
		}
		return
	}
	if op == "replace" && len(parts) >= 3 && parts[1] == "char" {
		char := strings.Join(parts[2:], "_")
		for i := 0; i < count; i++ {
			performPhysicalReplace(char, targetPane)
		}
		return
	}

	// 2. 处理传统 Op+Motion 组合
	if len(parts) < 2 {
		return
	}
	motion := strings.Join(parts[1:], "_")

	if op == "delete" || op == "change" {
		// FOEK Multi-Range 模拟
		for i := 0; i < count; i++ {
			// Check if it's a text object action (e.g., delete_inside_word)
			if strings.Contains(motion, "inside_") || strings.Contains(motion, "around_") {
				performPhysicalTextObject(op, motion, targetPane)
				continue
			}

			// Capture deleted text before it's gone
			startPos := getCursorPos(targetPane) // [col, row]
			content := captureText(motion, targetPane)

			if content != "" {
				// Record semantic Fact in active transaction
				record := captureShellDelete(targetPane, startPos[0], content)
				transMgr.Append(record)

				// [Phase 7] Robust Deletion:
				// Since we know EXACTLY what we captured, we delete by character count.
				// This is much safer than relying on shell M-d bindings.
				exec.Command("tmux", "send-keys", "-t", targetPane, "-N", fmt.Sprint(len(content)), "Delete").Run()
			} else {
				// Fallback if capture failed
				performPhysicalDelete(motion, targetPane)
			}
		}
		if op == "change" {
			exitFSM(targetPane) // change implies entering insert mode
		}
		state.RedoStack = nil
	} else if op == "yank" {
		if strings.Contains(motion, "inside_") || strings.Contains(motion, "around_") {
			performPhysicalTextObject(op, motion, targetPane)
		} else {
			// standard yank logic
		}
	} else if strings.HasPrefix(action, "find_") {
		parts := strings.SplitN(action, "_", 3)
		if len(parts) == 3 {
			performPhysicalFind(parts[1], parts[2], count, targetPane)
		}
	} else if op == "move" {
		performPhysicalMove(motion, count, targetPane)
	}
}

func currentCursor(targetPane string) (row, col int) {
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", targetPane, "#{pane_cursor_y},#{pane_cursor_x}").Output()
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d,%d", &row, &col)
	return
}

func captureLine(paneID string, line int) string {
	// Capture only the specific line
	out, _ := exec.Command("tmux", "capture-pane", "-p", "-t", paneID, "-J", "-S", fmt.Sprint(line), "-E", fmt.Sprint(line)).Output()
	return strings.TrimRight(string(out), "\n")
}

func hashLine(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func captureShellDelete(paneID string, startCol int, deletedText string) ActionRecord {
	row, col := currentCursor(paneID)
	line := captureLine(paneID, row)

	anchor := Anchor{
		PaneID:   paneID,
		LineHint: row,
		LineHash: hashLine(line),
		Cursor:   &[2]int{row, col},
	}

	r := Range{
		Anchor:      anchor,
		StartOffset: startCol,
		EndOffset:   startCol + len(deletedText),
		Text:        deletedText,
	}

	deleteFact := Fact{
		Kind:        "delete",
		Target:      r,
		SideEffects: []string{"clipboard_modified"},
	}

	insertInverse := Fact{
		Kind:   "insert",
		Target: r,
	}

	return ActionRecord{
		Fact:    deleteFact,
		Inverse: insertInverse,
	}
}

func captureShellChange(paneID string, startCol int, oldText, newText string) ActionRecord {
	row, col := currentCursor(paneID)
	line := captureLine(paneID, row)

	anchor := Anchor{
		PaneID:   paneID,
		LineHint: row,
		LineHash: hashLine(line),
		Cursor:   &[2]int{row, col},
	}

	r := Range{
		Anchor:      anchor,
		StartOffset: startCol,
		EndOffset:   startCol + len(oldText),
		Text:        oldText,
	}

	changeFact := Fact{
		Kind:        "replace",
		Target:      r,
		Meta:        map[string]interface{}{"new_text": newText},
		SideEffects: []string{"clipboard_modified"},
	}

	inverse := Fact{
		Kind:   "replace",
		Target: r,
		Meta:   map[string]interface{}{"new_text": oldText},
	}

	return ActionRecord{
		Fact:    changeFact,
		Inverse: inverse,
	}
}

func performPhysicalMove(motion string, count int, targetPane string) {
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
		exec.Command("tmux", "send-keys", "-t", targetPane, "Home").Run()
	case "end_of_line": // $
		exec.Command("tmux", "send-keys", "-t", targetPane, "End").Run()
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

func executeSearch(query string, targetPane string) {
	// 1. Enter copy mode if not in it
	// 2. Start search-forward
	exec.Command("tmux", "copy-mode", "-t", targetPane).Run()
	exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "search-forward", query).Run()
}

func performPhysicalTextObject(op, motion, targetPane string) {
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
		// Word detection logic
		start, end = findWordRange(currentLine, cursorX, strings.Contains(motion, "around_"))
	} else if strings.Contains(motion, "quote_") {
		// Quote detection
		quoteChar := "\""
		if strings.Contains(motion, "single") {
			quoteChar = "'"
		}
		start, end = findQuoteRange(currentLine, cursorX, quoteChar, strings.Contains(motion, "around_"))
	} else if strings.Contains(motion, "paren") || strings.Contains(motion, "bracket") || strings.Contains(motion, "brace") {
		// Bracket detection
		start, end = findBracketRange(currentLine, cursorX, motion, strings.Contains(motion, "around_"))
	}

	if start != -1 && end != -1 {
		// Execute
		if op == "delete" || op == "change" {
			// Jump to end, then backspace to start
			jumpTo(end, -1, targetPane)
			dist := end - start + 1
			exec.Command("tmux", "send-keys", "-t", targetPane, "-N", fmt.Sprint(dist), "BSpace").Run()
			if op == "change" {
				exec.Command("tmux", "send-keys", "-t", targetPane, "i").Run()
			}
		} else if op == "yank" {
			// Use tmux selection
			jumpTo(start, -1, targetPane)
			exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "begin-selection").Run()
			jumpTo(end, -1, targetPane)
			exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "copy-pipe-and-cancel", "tmux save-buffer -").Run()
		}
	}
}

func findWordRange(line string, x int, around bool) (int, int) {
	if x >= len(line) {
		return -1, -1
	}

	isWordChar := func(c byte) bool {
		return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
	}

	// Find start
	start := x
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}
	// Find end
	end := x
	for end < len(line)-1 && isWordChar(line[end+1]) {
		end++
	}

	if around {
		// Include one trailing space if exists
		if end < len(line)-1 && line[end+1] == ' ' {
			end++
		} else if start > 0 && line[start-1] == ' ' {
			// Or leading if trailing not found
			start--
		}
	}

	return start, end
}

func findQuoteRange(line string, x int, quote string, around bool) (int, int) {
	// Simple quote range: find surrounding quotes on current line
	first := strings.LastIndex(line[:x+1], quote)
	if first == -1 {
		// Try looking ahead if not found sitting on it
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

	// Find the pair that surrounds x
	// Search backward for opening
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

	// Search forward for closing
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

func performPhysicalFind(fType, char string, count int, targetPane string) {
	// 1. Capture current line content
	// We use tmux capture-pane to get the current row
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", targetPane, "#{pane_cursor_x}").Output()
	var cursorX int
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &cursorX)

	out, _ = exec.Command("tmux", "capture-pane", "-p", "-t", targetPane, "-J").Output()
	lines := strings.Split(string(out), "\n")

	// Get the line the cursor is on. This is tricky because capture-pane -p results
	// might have different wrapping. A safer way is using 'display-message -p' for line.
	// But let's simplified for single line shell context:
	// We'll use the last non-empty line as the "current line" for Shell prompt
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
	case "f": // forward find
		for x := cursorX + 1; x < len(currentLine); x++ {
			if string(currentLine[x]) == char {
				foundCount++
				if foundCount == count {
					targetX = x
					break
				}
			}
		}
	case "F": // backward find
		for x := cursorX - 1; x >= 0; x-- {
			if string(currentLine[x]) == char {
				foundCount++
				if foundCount == count {
					targetX = x
					break
				}
			}
		}
	case "t": // forward until
		for x := cursorX + 1; x < len(currentLine); x++ {
			if string(currentLine[x]) == char {
				foundCount++
				if foundCount == count {
					targetX = x - 1
					break
				}
			}
		}
	case "T": // backward until
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
		jumpTo(targetX, -1, targetPane) // -1 means stay on current Y
	}
}

func handleUndo(state *FSMState, targetPane string) {
	// [Phase 9] Legacy undo now handled by Weaver as single source of truth
	// This function should not be called directly anymore
	// Undo is now dispatched as Intent to Weaver via ProcessIntentGlobal
}

func logLine(msg string) {
	f, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("15:04:05"), msg)
		f.Close()
	}
}

// 辅助函数...
func getCursorPos(targetPane string) [2]int {
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", targetPane, "#{pane_cursor_x},#{pane_cursor_y}").Output()
	var x, y int
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d,%d", &x, &y)
	return [2]int{x, y}
}

func jumpTo(x, y int, targetPane string) {
	// 简单的跳转模拟 (Arrow keys)
	curr := getCursorPos(targetPane)
	dx := x - curr[0]
	dy := y - curr[1]

	if dy != 0 && y != -1 {
		var moveKey string = "Up"
		if dy > 0 {
			moveKey = "Down"
		}
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", fmt.Sprint(abs(dy)), moveKey).Run()
	}
	if dx != 0 {
		var moveKey string = "Left"
		if dx > 0 {
			moveKey = "Right"
		}
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", fmt.Sprint(abs(dx)), moveKey).Run()
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func captureText(motion string, targetPane string) string {
	if motion == "word_forward" {
		// [Phase 7] Axiom 9: Deterministic Reality
		// Instead of copy-mode UI (which is asynchronous and flaky),
		// we use capture-pane and parse the word boundary in Go.
		row, col := currentCursor(targetPane)
		line := captureLine(targetPane, row)

		if col >= len(line) {
			return ""
		}

		isWordChar := func(c byte) bool {
			return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
		}

		// Find end of current word
		end := col
		// If at start of word, or non-word chars, identify the range to delete
		if isWordChar(line[col]) {
			// Forward to end of word
			for end < len(line) && isWordChar(line[end]) {
				end++
			}
			// Include trailing whitespace (standard 'dw' behavior)
			for end < len(line) && line[end] == ' ' {
				end++
			}
		} else {
			// On whitespace/punctuation: delete the sequence of those
			for end < len(line) && !isWordChar(line[end]) {
				end++
			}
		}

		return line[col:end]
	}
	return ""
}

func performPhysicalDelete(motion string, targetPane string) {
	// 首先取消任何现有的选择
	exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "cancel").Run()

	switch motion {
	case "start_of_line": // d0
		// Robust implementation: Get cursor X position and backspace that many times
		// This avoids Zsh/Bash differences with C-u
		pos := getCursorPos(targetPane)
		cursorX := pos[0]
		if cursorX > 0 {
			exec.Command("tmux", "send-keys", "-t", targetPane, "-N", fmt.Sprint(cursorX), "BSpace").Run()
		}

	case "end_of_line": // d$
		// C-k: Kill to end of line
		exec.Command("tmux", "send-keys", "-t", targetPane, "C-k").Run()

	case "word_forward", "inside_word", "around_word": // dw
		// Robust implementation: M-d (Alt-d) is the shell standard for delete-word-forward.
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

func handleVisualAction(action string, state *FSMState, targetPane string) {
	// 提取操作类型 (delete, yank, change)
	parts := strings.Split(action, "_")
	if len(parts) < 2 {
		return
	}

	op := parts[1] // delete, yank, 或 change

	if isVimPane(targetPane) {
		// 在Vim中执行视觉模式操作
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
		// 在Shell中执行视觉模式操作
		if op == "yank" {
			// 复制选中内容
			exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "copy-pipe-and-cancel", "tmux save-buffer -").Run()
		} else if op == "delete" || op == "change" {
			// 删除选中内容
			exec.Command("tmux", "send-keys", "-t", targetPane, "-X", "copy-pipe-and-cancel", "tmux save-buffer -").Run()
			if op == "change" {
				// change 操作需要额外输入
				exec.Command("tmux", "send-keys", "-t", targetPane, "i").Run()
			}
		}
	}
}

func handleRedo(state *FSMState, targetPane string) {
	// [Phase 9] Legacy redo now handled by Weaver as single source of truth
	// This function should not be called directly anymore
	// Redo is now dispatched as Intent to Weaver via ProcessIntentGlobal
}

func executeVimAction(action string, state *FSMState, targetPane string) {
	// Map FSM actions to Vim native keys
	vimKey := ""
	isEdit := false

	switch action {
	case "move_left":
		vimKey = "h"
	case "move_down":
		vimKey = "j"
	case "move_up":
		vimKey = "k"
	case "move_right":
		vimKey = "l"
	case "move_word_forward":
		vimKey = "w"
	case "move_word_backward":
		vimKey = "b"
	case "move_end_of_word":
		vimKey = "e"
	case "move_start_of_line":
		vimKey = "0"
	case "move_end_of_line":
		vimKey = "$"
	case "move_start_of_file":
		vimKey = "gg"
	case "move_end_of_file":
		vimKey = "G"
	case "delete_line":
		vimKey = "dd"
		isEdit = true
	case "delete_word_forward":
		vimKey = "dw"
		isEdit = true
	case "delete_word_backward":
		vimKey = "db"
		isEdit = true
	case "delete_end_of_word":
		vimKey = "de"
		isEdit = true
	case "delete_right":
		vimKey = "x"
		isEdit = true
	case "delete_left":
		vimKey = "X"
		isEdit = true
	case "delete_end_of_line":
		vimKey = "D"
		isEdit = true
	case "change_end_of_line":
		vimKey = "C"
		isEdit = true
	case "change_line":
		vimKey = "S"
		isEdit = true
	case "insert_start_of_line":
		vimKey = "I"
		isEdit = true
	case "insert_end_of_line":
		vimKey = "A"
		isEdit = true
	case "insert_before":
		vimKey = "i"
		isEdit = true
	case "insert_after":
		vimKey = "a"
		isEdit = true
	case "insert_open_below":
		vimKey = "o"
		isEdit = true
	case "insert_open_above":
		vimKey = "O"
		isEdit = true
	case "paste_after":
		vimKey = "p"
		isEdit = true
	case "paste_before":
		vimKey = "P"
		isEdit = true
	case "toggle_case":
		vimKey = "~"
		isEdit = true
	case "undo":
		vimKey = "u"
	case "redo":
		vimKey = "C-r"
	}

	if strings.HasPrefix(action, "replace_char_") {
		char := strings.TrimPrefix(action, "replace_char_")
		vimKey = "r" + char
		isEdit = true
	}

	if vimKey == "" {
		// Fallback: if not mapped, it might be a direct key or sequence
		return
	}

	if isEdit {
		// Record a Fact that delegates undo to Vim
		anchor := Anchor{PaneID: targetPane}
		record := ActionRecord{
			Fact:    Fact{Kind: "insert", Target: Range{Anchor: anchor, Text: vimKey}, Meta: map[string]interface{}{"is_vim_raw": true}}, // Pseudo-fact
			Inverse: Fact{Kind: "undo", Target: Range{Anchor: anchor}},
		}
		transMgr.Append(record)
	}

	// For Vim, we just send the count + key
	countStr := ""
	if state.Count > 0 {
		countStr = fmt.Sprint(state.Count)
	}
	exec.Command("tmux", "send-keys", "-t", targetPane, countStr+vimKey).Run()
}

func getHelpText(state *FSMState) string {
	helpText := `
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃                Tmux-FSM (Weaver Core) Cheat Sheet                  ┃
┃                   苑广山@yuanguangshan@gmail.com                   ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

  MOTIONS (移动)            OPERATORS (操作)          TEXT OBJECTS (对象)
  ──────────────            ────────────────          ───────────────────
  h/j/k/l : 左/下/上/右     d : Delete (删除)         iw/aw : 单词 (Word)
  w/b/e   : 词首/词退/词尾  c : Change (修改)         i"/a" : 引号 (Quote)
  0 / $   : 行首 / 行尾     y : Yank   (复制)         i(/i[ : 括号 (Bracket)
  gg / G  : 文首 / 文末     u : Undo   (撤销)         i{    : 大括号 (Brace)
  C-b/C-f : 向上/下翻页     C-r : Redo (重做)         
                            . : Repeat (重复上次)     SEARCH & FIND (查找)
  EDITING (编辑)            p / P : Paste (粘贴)      ───────────────────
  ──────────────            r : Replace (单字替换)    / / ? : 向前/后搜索
  x / X   : 删后/前一个字   ~ : Toggle Case(大小写)   n / N : 下个/上个匹配
  i / a   : 前 / 后插入                               f/F/t/T : 字符跳跃
  I / A   : 行首 / 行尾插入  META (元命令)
  o / O   : 下 / 上开新行    ──────────────
                             Esc/C-c : 退出模式(Exit)
                             ?       : 查看此帮助/审计
`
	if state.LastUndoFailure != "" {
		helpText += fmt.Sprintf("  [!] LAST AUDIT FAILURE (上轮撤销失败原因):\n      >> %s\n\n", state.LastUndoFailure)
	} else {
		helpText += "  ( 💡 审计说明: 若撤销由于安全校验被拦截，此处将显示异常原因 )\n\n"
	}
	return helpText
}

func showHelp(state *FSMState, targetPane string) {
	helpText := getHelpText(state)
	// Use fixed dimensions for a clean, centered look on desktop.
	// 80x28 is sufficient for the cheat sheet content.
	exec.Command("tmux", "display-popup", "-t", targetPane, "-E", "-w", "80", "-h", "28", fmt.Sprintf("echo '%s'; read -n 1", helpText)).Run()
}

func exitFSM(targetPane string) {
	exec.Command("tmux", "set", "-g", "@fsm_active", "false").Run()
	exec.Command("tmux", "set", "-g", "@fsm_state", "").Run()
	exec.Command("tmux", "set", "-g", "@fsm_keys", "").Run()
	exec.Command("tmux", "switch-client", "-T", "root").Run()
	exec.Command("tmux", "refresh-client", "-S").Run()
}

func performPhysicalInsert(motion, targetPane string) {
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

func performPhysicalPaste(motion, targetPane string) {
	if motion == "after" {
		exec.Command("tmux", "send-keys", "-t", targetPane, "Right").Run()
	}
	exec.Command("tmux", "paste-buffer", "-t", targetPane).Run()
}

func performPhysicalReplace(char, targetPane string) {
	exec.Command("tmux", "send-keys", "-t", targetPane, "Delete", char).Run()
}

func performPhysicalToggleCase(targetPane string) {
	// Captures the char under cursor, toggles it, and replaces it.
	pos := getCursorPos(targetPane)
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
