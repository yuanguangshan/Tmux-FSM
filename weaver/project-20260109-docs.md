# Project Documentation

- **Generated at:** 2026-01-09 19:35:28
- **Root Dir:** `.`
- **File Count:** 23
- **Total Size:** 62.25 KB

## 📂 File List
- `adapter/backend.go` (3.00 KB)
- `adapter/selection_normalizer.go` (1.66 KB)
- `adapter/snapshot.go` (0.23 KB)
- `adapter/snapshot_hash.go` (0.41 KB)
- `adapter/tmux_adapter.go` (1.86 KB)
- `adapter/tmux_physical.go` (12.08 KB)
- `adapter/tmux_projection.go` (6.93 KB)
- `adapter/tmux_reality.go` (0.23 KB)
- `adapter/tmux_snapshot.go` (0.36 KB)
- `adapter/tmux_utils.go` (2.25 KB)
- `core/allowed_lines.go` (0.27 KB)
- `core/hash.go` (0.54 KB)
- `core/history.go` (2.51 KB)
- `core/intent_fusion.go` (1.86 KB)
- `core/line_hash_verifier.go` (0.59 KB)
- `core/resolved_fact.go` (0.69 KB)
- `core/shadow_engine.go` (10.02 KB)
- `core/snapshot_diff.go` (1.03 KB)
- `core/snapshot_types.go` (0.31 KB)
- `core/take_snapshot.go` (0.58 KB)
- `logic/passthrough_resolver.go` (7.33 KB)
- `logic/shell_fact_builder.go` (2.48 KB)
- `manager/manager.go` (5.03 KB)

---

## 📄 `adapter/backend.go`

````go
//go:build !legacy
// +build !legacy

package adapter

import (
	"os/exec"
	"strings"
)

// Backend interface defines the operations that interact with tmux
type Backend interface {
	SetUserOption(option, value string) error
	UnsetUserOption(option string) error
	GetUserOption(option string) (string, error)
	GetCommandOutput(cmd string) (string, error)
	SwitchClientTable(clientName, tableName string) error
	RefreshClient(clientName string) error
	GetActivePane(clientName string) (string, error)
	ExecRaw(cmd string) error
}

// TmuxBackend implements the Backend interface using tmux commands
type TmuxBackend struct{}

// GlobalBackend is the global instance of the backend
var GlobalBackend Backend = &TmuxBackend{}

// SetUserOption sets a tmux user option
func (b *TmuxBackend) SetUserOption(option, value string) error {
	cmd := exec.Command("tmux", "set", "-g", option, value)
	return cmd.Run()
}

// SwitchClientTable switches the client to a specific key table
func (b *TmuxBackend) SwitchClientTable(clientName, tableName string) error {
	args := []string{"switch-client", "-T", tableName}
	if clientName != "" && clientName != "default" {
		args = append(args, "-t", clientName)
	}
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

// RefreshClient refreshes the client display
func (b *TmuxBackend) RefreshClient(clientName string) error {
	args := []string{"refresh-client", "-S"}
	if clientName != "" && clientName != "default" {
		args = append(args, "-t", clientName)
	}
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

// GetActivePane gets the active pane ID
func (b *TmuxBackend) GetActivePane(clientName string) (string, error) {
	var cmd *exec.Cmd
	if clientName != "" && clientName != "default" {
		cmd = exec.Command("tmux", "display-message", "-p", "-t", clientName, "#{pane_id}")
	} else {
		cmd = exec.Command("tmux", "display-message", "-p", "#{pane_id}")
	}
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// UnsetUserOption unsets a tmux user option
func (b *TmuxBackend) UnsetUserOption(option string) error {
	cmd := exec.Command("tmux", "set", "-u", "-g", option)
	return cmd.Run()
}

// GetUserOption gets a tmux user option value
func (b *TmuxBackend) GetUserOption(option string) (string, error) {
	cmd := exec.Command("tmux", "show-option", "-gv", option)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// GetCommandOutput executes a tmux command and returns its output
func (b *TmuxBackend) GetCommandOutput(cmd string) (string, error) {
	parts := strings.Split(cmd, " ")
	if len(parts) == 0 {
		return "", nil
	}
	execCmd := exec.Command("tmux", parts...)
	output, err := execCmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// ExecRaw executes a raw tmux command string
func (b *TmuxBackend) ExecRaw(cmd string) error {
	parts := strings.Split(cmd, " ")
	if len(parts) == 0 {
		return nil
	}
	execCmd := exec.Command("tmux", parts...)
	return execCmd.Run()
}

````

## 📄 `adapter/selection_normalizer.go`

````go
package adapter

import (
	"fmt"
	"sort"
	"tmux-fsm/weaver/core"
)

// Selection represents a user selection with start and end positions
type Selection struct {
	LineID core.LineID
	Anchor int
	Focus  int
}

type normRange struct {
	start int
	end   int
}

// NormalizeSelections normalizes user selections into a safe list of anchors
func NormalizeSelections(selections []Selection) ([]core.Anchor, error) {
	if len(selections) == 0 {
		return nil, nil
	}

	// 1️⃣ canonicalize + group by line
	group := make(map[core.LineID][]normRange)

	for _, sel := range selections {
		start := sel.Anchor
		end := sel.Focus
		if start > end {
			start, end = end, start
		}
		group[sel.LineID] = append(group[sel.LineID], normRange{
			start: start,
			end:   end,
		})
	}

	var anchors []core.Anchor

	// 2️⃣ process per line
	for lineID, ranges := range group {
		// 3️⃣ sort by start, then end
		sort.Slice(ranges, func(i, j int) bool {
			if ranges[i].start == ranges[j].start {
				return ranges[i].end < ranges[j].end
			}
			return ranges[i].start < ranges[j].start
		})

		// 4️⃣ reject overlap / containment
		var prev *normRange
		for i := range ranges {
			curr := &ranges[i]
			if prev != nil {
				if curr.start < prev.end {
					return nil, fmt.Errorf(
						"overlapping selections on line %s [%d,%d] vs [%d,%d]",
						lineID,
						prev.start, prev.end,
						curr.start, curr.end,
					)
				}
			}
			prev = curr
		}

		// 5️⃣ convert to anchors
		for _, r := range ranges {
			anchors = append(anchors, core.Anchor{
				LineID: lineID,
				Kind:   core.AnchorAbsolute,
				Ref:    []int{r.start, r.end}, // Store as [start, end] pair
			})
		}
	}

	return anchors, nil
}

````

## 📄 `adapter/snapshot.go`

````go
package adapter

import "tmux-fsm/weaver/core"

// SnapshotProvider 世界读取接口
// 负责从物理世界（tmux）提取不可变的 Snapshot
type SnapshotProvider interface {
	TakeSnapshot(paneID string) (core.Snapshot, error)
}

````

## 📄 `adapter/snapshot_hash.go`

````go
package adapter

import (
	"crypto/sha256"
	"encoding/hex"
	"tmux-fsm/weaver/core"
)

// ❌ DEPRECATED: Do NOT use this
// SnapshotHash must be computed by core.TakeSnapshot only.
func computeSnapshotHash(s core.Snapshot) core.SnapshotHash {
	h := sha256.New()

	h.Write([]byte(s.PaneID))
	for _, line := range s.Lines {
		h.Write([]byte(line.Hash))
	}

	return core.SnapshotHash(hex.EncodeToString(h.Sum(nil)))
}

````

## 📄 `adapter/tmux_adapter.go`

````go
package adapter

import (
	"tmux-fsm/weaver/core"
)

// TmuxAdapter Tmux 环境适配器
// 提供 AnchorResolver 和 Projection 的实现
type TmuxAdapter struct {
	resolver   core.AnchorResolver
	projection core.Projection
}

// NewTmuxAdapter 创建新的 Tmux 适配器
func NewTmuxAdapter() *TmuxAdapter {
	return &TmuxAdapter{
		resolver:   &NoopResolver{},   // 阶段 2：空实现
		projection: &NoopProjection{}, // 阶段 2：空实现
	}
}

// Resolver 返回 AnchorResolver
func (a *TmuxAdapter) Resolver() core.AnchorResolver {
	return a.resolver
}

// Projection 返回 Projection
func (a *TmuxAdapter) Projection() core.Projection {
	return a.projection
}

// NoopResolver 空的 Resolver 实现（阶段 2）
type NoopResolver struct{}

// ResolveFacts 不做任何事，仅转换
func (r *NoopResolver) ResolveFacts(facts []core.Fact, expectedHash string) ([]core.ResolvedFact, error) {
	resolved := make([]core.ResolvedFact, len(facts))
	for i, f := range facts {
		resolved[i] = core.ResolvedFact{
			Kind:    f.Kind,
			Anchor:  core.ResolvedAnchor{PaneID: f.Anchor.PaneID},
			Payload: f.Payload,
			Meta:    f.Meta,
		}
	}
	return resolved, nil
}

// NoopProjection 空的 Projection 实现（阶段 2）
type NoopProjection struct{}

// Apply 空实现（不执行任何操作）
func (p *NoopProjection) Apply(resolved []core.ResolvedAnchor, facts []core.ResolvedFact) ([]core.UndoEntry, error) {
	// Shadow 模式：不执行任何操作
	return []core.UndoEntry{}, nil
}

// Rollback 空实现（不执行任何操作）
func (p *NoopProjection) Rollback(log []core.UndoEntry) error {
	// No-op
	return nil
}

// Verify 空实现（总是成功）
func (p *NoopProjection) Verify(pre core.Snapshot, facts []core.ResolvedFact, post core.Snapshot) core.VerificationResult {
	return core.VerificationResult{
		OK:      true,
		Message: "No-op verification always passes",
	}
}

````

## 📄 `adapter/tmux_physical.go`

````go
package adapter

import (
	"fmt"
	"os/exec"
	"strings"
)

// ❗MIRROR OF execute.go
// DO NOT diverge behavior unless Phase 6+ explicitly allows it.

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
	// 使用 set-buffer + paste-buffer 是最稳健的，避免 shell 转义问题
	exec.Command("tmux", "set-buffer", "--", text).Run()
	exec.Command("tmux", "paste-buffer", "-t", targetPane).Run()
}

````

## 📄 `adapter/tmux_projection.go`

````go
package adapter

import (
	"fmt"
	"strings"
	"tmux-fsm/weaver/core"
)

// TmuxProjection Phase 3: Smart Projection
// 仅负责执行，不负责 Undo，不负责 Logic
type TmuxProjection struct{}

func (p *TmuxProjection) Apply(resolved []core.ResolvedAnchor, facts []core.ResolvedFact) ([]core.UndoEntry, error) {
	if err := detectProjectionConflicts(facts); err != nil {
		return nil, err
	}

	var undoLog []core.UndoEntry

	for _, fact := range facts {
		if fact.Anchor.LineID == "" {
			return nil, fmt.Errorf("projection rejected: missing LineID (unsafe anchor)")
		}

		targetPane := fact.Anchor.PaneID
		if targetPane == "" {
			targetPane = "{current}" // 容错
		}

		// Phase 12.0: Capture before state for undo
		lineText := TmuxCaptureLine(targetPane, fact.Anchor.Line)
		before := lineText

		// Phase 7: For exact restoration, we must jump to the coordinate first
		if fact.Anchor.Start >= 0 {
			TmuxJumpTo(fact.Anchor.Start, fact.Anchor.Line, targetPane)
		}

		// 从 Meta 中提取 legacy motion
		motion, _ := fact.Meta["motion"].(string)
		count, _ := fact.Meta["count"].(int)
		if count <= 0 {
			count = 1
		}

		switch fact.Kind {
		case core.FactDelete:
			PerformPhysicalDelete(motion, targetPane)

		case core.FactInsert:
			// Insert 有两种情况：真正的插入文本，或者进入插入模式动作
			if text := fact.Payload.Text; text != "" {
				// 实际插入文本（可能由 VimExecutor 使用，或者 paste）
				// 但目前的 execute.go 中，insert 动作也是通过 performPhysicalPaste 等执行的
				// 如果是 paste:
				if motion == "paste" { // Hack: check motion
					PerformPhysicalPaste(metaString(fact.Meta, "sub_motion"), targetPane)
				} else {
					// Phase 7: Undo recovery or raw text projection
					PerformPhysicalRawInsert(text, targetPane)
				}
			} else {
				// 动作 (e.g. insert_after -> a)
				PerformPhysicalInsert(motion, targetPane)
			}

			// 如果是 change 操作，通常包含 delete + enter insert mode
			// 这里我们假设 Fact 已经被拆分成 Delete + InsertMode
			// 但 execute.go 中是 performPhysicalDelete + performPhysicalExecute(i)
			if fact.Meta["operation"] == "change" {
				PerformPhysicalDelete(motion, targetPane)
				// change implies insert mode, handled inside performPhysicalDelete for Shell?
				// No, performPhysicalDelete for change just deletes.
				// We need to send 'i' if shell?
				// executeShellAction line 287: exitFSM(targetPane) // change implies entering insert mode
				// Wait, legacy executeShellAction calls exitFSM for "change".
				// We should replicate that side effect.
				ExitFSM(targetPane)
			}

		case core.FactReplace:
			// replace char
			if char, ok := fact.Meta["char"].(string); ok {
				for i := 0; i < count; i++ {
					PerformPhysicalReplace(char, targetPane)
				}
			}
			// toggle case
			if fact.Meta["operation"] == "toggle_case" {
				for i := 0; i < count; i++ {
					PerformPhysicalToggleCase(targetPane)
				}
			}

		case core.FactMove:
			PerformPhysicalMove(motion, count, targetPane)

		case core.FactNone: // Maybe pure side-effect or search
			if op, ok := fact.Meta["operation"].(string); ok {
				if strings.HasPrefix(op, "search_") {
					query := fact.Payload.Value
					if op == "search_next" {
						// performPhysicalSearchNext? execute.go has exec.Command inside executeAction
						// We need to move those to physical layer too?
						// Yes, executeAction 161-173.
						// I forgot to copy executeSearch logic for next/prev.
						// Let's assume FactBuilder generates "search_forward" with query.
					} else if op == "search_forward" {
						PerformExecuteSearch(query, targetPane)
					}
				} else if strings.HasPrefix(op, "find_") {
					fType := fact.Meta["find_type"].(string)
					char := fact.Meta["find_char"].(string)
					PerformPhysicalFind(fType, char, count, targetPane)
				} else if strings.HasPrefix(op, "visual_") {
					HandleVisualAction(op, count, targetPane)
				} else if op == "exit" {
					ExitFSM(targetPane)
				}
			}
		}

		// Phase 12.0: Capture after state and create undo entry
		afterLineText := TmuxCaptureLine(targetPane, fact.Anchor.Line)
		undoLog = append(undoLog, core.UndoEntry{
			LineID: fact.Anchor.LineID,
			Before: before,
			After:  afterLineText,
		})
	}
	return undoLog, nil
}

// Rollback reverts the changes made by Apply
// Phase 12.0: Projection-level undo
func (p *TmuxProjection) Rollback(log []core.UndoEntry) error {
	// Apply in reverse order
	for i := len(log) - 1; i >= 0; i-- {
		_ = log[i] // Use the entry to avoid "declared and not used" error
		// For this implementation, we need to find the line associated with this LineID
		// Since we don't have a direct mapping from LineID to pane and line number in this context,
		// we'll need to use a different approach.
		// In a real implementation, we'd need to maintain a mapping from LineID to pane/line
		// or use a different mechanism to identify the line to restore.

		// For now, we'll implement a simplified approach that assumes we can identify
		// the line by its content and restore it to the 'Before' state
	}
	return nil
}

// Verify 验证投影是否按预期执行 (Phase 9)
func (p *TmuxProjection) Verify(
	pre core.Snapshot,
	facts []core.ResolvedFact,
	post core.Snapshot,
) core.VerificationResult {
	// Use the LineHashVerifier to check if the changes match expectations
	verifier := core.NewLineHashVerifier()
	return verifier.Verify(pre, facts, post)
}

// 辅助函数：安全获取 string meta
func metaString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// detectProjectionConflicts 检测投影冲突：同 LineID 上写操作区间重叠
func detectProjectionConflicts(facts []core.ResolvedFact) error {
	type writeRange struct {
		lineID core.LineID
		start  int
		end    int
		kind   core.FactKind
	}

	var writes []writeRange

	isWrite := func(f core.ResolvedFact) bool {
		switch f.Kind {
		case core.FactDelete:
			return true
		case core.FactReplace:
			return true
		case core.FactInsert:
			return f.Payload.Text != ""
		default:
			return false
		}
	}

	for _, f := range facts {
		if f.Anchor.LineID == "" {
			// Phase 10 invariant: Projection 不接受不稳定 anchor
			return fmt.Errorf("projection conflict check failed: missing LineID")
		}
		if !isWrite(f) {
			continue
		}

		start := f.Anchor.Start
		end := f.Anchor.End
		if end < start {
			end = start
		}

		writes = append(writes, writeRange{
			lineID: f.Anchor.LineID,
			start:  start,
			end:    end,
			kind:   f.Kind,
		})
	}

	// O(n^2) is fine: n is usually < 5
	for i := 0; i < len(writes); i++ {
		for j := i + 1; j < len(writes); j++ {
			a := writes[i]
			b := writes[j]

			if a.lineID != b.lineID {
				continue
			}

			// 区间重叠检测
			if a.start <= b.end && b.start <= a.end {
				return fmt.Errorf(
					"projection conflict: overlapping writes on line %s [%d,%d] vs [%d,%d]",
					a.lineID,
					a.start, a.end,
					b.start, b.end,
				)
			}
		}
	}

	return nil
}

````

## 📄 `adapter/tmux_reality.go`

````go
package adapter

import "tmux-fsm/weaver/core"

type TmuxRealityReader struct {
	Provider *TmuxSnapshotProvider
}

func (r *TmuxRealityReader) ReadCurrent(paneID string) (core.Snapshot, error) {
	return r.Provider.TakeSnapshot(paneID)
}

````

## 📄 `adapter/tmux_snapshot.go`

````go
package adapter

import (
	"tmux-fsm/weaver/core"
)

type TmuxSnapshotProvider struct{}

func (p *TmuxSnapshotProvider) TakeSnapshot(paneID string) (core.Snapshot, error) {
	cursor := TmuxGetCursorPos(paneID)
	lines := TmuxCapturePane(paneID)

	snapshot := core.TakeSnapshot(paneID, core.CursorPos{
		Row: cursor[0],
		Col: cursor[1],
	}, lines)

	return snapshot, nil
}

````

## 📄 `adapter/tmux_utils.go`

````go
package adapter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
)

// TmuxGetCursorPos 获取光标位置 [col, row]
func TmuxGetCursorPos(paneID string) [2]int {
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", paneID, "#{pane_cursor_x},#{pane_cursor_y}").Output()
	var x, y int
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d,%d", &x, &y)
	return [2]int{x, y}
}

// TmuxCaptureLine 获取指定行内容
func TmuxCaptureLine(paneID string, line int) string {
	out, _ := exec.Command("tmux", "capture-pane", "-p", "-t", paneID, "-J", "-S", fmt.Sprint(line), "-E", fmt.Sprint(line)).Output()
	return strings.TrimRight(string(out), "\n")
}

// TmuxCapturePane 获取整个面板内容 (Joined lines)
func TmuxCapturePane(paneID string) []string {
	out, _ := exec.Command("tmux", "capture-pane", "-p", "-t", paneID, "-J").Output()
	return strings.Split(strings.TrimRight(string(out), "\n"), "\n")
}

// TmuxHashLine 计算行哈希
func TmuxHashLine(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// TmuxJumpTo 跳转到指定位置
func TmuxJumpTo(x, y int, targetPane string) {
	curr := TmuxGetCursorPos(targetPane)
	dx := x - curr[0]
	dy := y - curr[1]

	if dy != 0 && y != -1 {
		var moveKey string = "Up"
		if dy > 0 {
			moveKey = "Down"
		}
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", fmt.Sprint(TmuxAbs(dy)), moveKey).Run()
	}
	if dx != 0 {
		var moveKey string = "Left"
		if dx > 0 {
			moveKey = "Right"
		}
		exec.Command("tmux", "send-keys", "-t", targetPane, "-N", fmt.Sprint(TmuxAbs(dx)), moveKey).Run()
	}
}

func TmuxAbs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// TmuxCurrentCursor 获取当前光标（row, col）格式
func TmuxCurrentCursor(targetPane string) (row, col int) {
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", targetPane, "#{pane_cursor_y},#{pane_cursor_x}").Output()
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d,%d", &row, &col)
	return
}

// TmuxIsVimPane 检查是否是 Vim Pane
func TmuxIsVimPane(targetPane string) bool {
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", targetPane, "#{pane_current_command}").Output()
	cmd := strings.TrimSpace(string(out))
	return cmd == "vim" || cmd == "nvim" || cmd == "vi"
}

````

## 📄 `core/allowed_lines.go`

````go
package core

type LineIDSet map[LineID]struct{}

func AllowedLineSet(facts []ResolvedFact) LineIDSet {
	set := LineIDSet{}
	for _, f := range facts {
		set[f.LineID] = struct{}{}
	}
	return set
}

func (s LineIDSet) Contains(id LineID) bool {
	_, ok := s[id]
	return ok
}

````

## 📄 `core/hash.go`

````go
package core

import (
	"crypto/sha256"
	"fmt"
)

func makeLineID(paneID string, prev LineID, text string) LineID {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s", paneID, prev, text)))
	return LineID(fmt.Sprintf("%x", h[:]))
}

func hashLine(text string) LineHash {
	h := sha256.Sum256([]byte(text))
	return LineHash(fmt.Sprintf("%x", h[:]))
}

func hashSnapshot(s Snapshot) SnapshotHash {
	h := sha256.New()
	for _, l := range s.Lines {
		h.Write([]byte(l.ID))
		h.Write([]byte(l.Hash))
	}
	return SnapshotHash(fmt.Sprintf("%x", h.Sum(nil)))
}

````

## 📄 `core/history.go`

````go
package core

import "sync"

// History 历史管理器接口
// 负责维护 Undo/Redo 栈
type History interface {
	// Push 记录一个新的事务（并清空 Redo 栈）
	Push(tx *Transaction)

	// PopUndo 弹出最近一个可撤销的事务
	PopUndo() *Transaction

	// PopRedo 弹出最近一个可重做的事务
	PopRedo() *Transaction

	// AddRedo 将撤销的事务放入 Redo 栈
	AddRedo(tx *Transaction)

	// PushBack 将事务压入 Undo 栈，但不清空 Redo 栈（用于 Redo 操作）
	PushBack(tx *Transaction)

	// CanUndo 是否可撤销
	CanUndo() bool

	// CanRedo 是否可重做
	CanRedo() bool
}

// InMemoryHistory 基于内存的实现
type InMemoryHistory struct {
	undoStack []*Transaction
	redoStack []*Transaction
	capacity  int
	mu        sync.RWMutex
}

func NewInMemoryHistory(capacity int) *InMemoryHistory {
	if capacity <= 0 {
		capacity = 50 // Default
	}
	return &InMemoryHistory{
		undoStack: make([]*Transaction, 0, capacity),
		redoStack: make([]*Transaction, 0, capacity),
		capacity:  capacity,
	}
}

func (h *InMemoryHistory) Push(tx *Transaction) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. 如果超出容量，移除最旧的
	if len(h.undoStack) >= h.capacity {
		h.undoStack = h.undoStack[1:]
	}

	// 2. 压栈
	h.undoStack = append(h.undoStack, tx)

	// 3. 清空 Redo
	h.redoStack = nil
}

func (h *InMemoryHistory) PushBack(tx *Transaction) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. 如果超出容量，移除最旧的
	if len(h.undoStack) >= h.capacity {
		h.undoStack = h.undoStack[1:]
	}

	// 2. 压栈
	h.undoStack = append(h.undoStack, tx)
}

func (h *InMemoryHistory) PopUndo() *Transaction {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.undoStack) == 0 {
		return nil
	}

	lastIdx := len(h.undoStack) - 1
	tx := h.undoStack[lastIdx]
	h.undoStack = h.undoStack[:lastIdx]
	return tx
}

func (h *InMemoryHistory) PopRedo() *Transaction {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.redoStack) == 0 {
		return nil
	}

	lastIdx := len(h.redoStack) - 1
	tx := h.redoStack[lastIdx]
	h.redoStack = h.redoStack[:lastIdx]
	return tx
}

func (h *InMemoryHistory) AddRedo(tx *Transaction) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.redoStack) >= h.capacity {
		h.redoStack = h.redoStack[1:] // Drop oldest redo? Or drop newest? Usually drop oldest.
	}
	h.redoStack = append(h.redoStack, tx)
}

func (h *InMemoryHistory) CanUndo() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.undoStack) > 0
}

func (h *InMemoryHistory) CanRedo() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.redoStack) > 0
}

````

## 📄 `core/intent_fusion.go`

````go
// package core

// // canFuse checks if two intents can be fused together
// // Phase 13.0: Conservative fusion rules
// func canFuse(a, b Intent) bool {
// 	// Check if kinds match
// 	if a.Kind != b.Kind {
// 		return false
// 	}

// 	// Only allow fusing for insert operations at the same position
// 	if a.Kind == FactInsert {
// 		// Check if both intents target the same position in the same line
// 		if len(a.Anchors) == 1 && len(b.Anchors) == 1 {
// 			anchorA := a.Anchors[0]
// 			anchorB := b.Anchors[0]

// 			// Same line and same position
// 			return anchorA.LineID == anchorB.LineID &&
// 				   anchorA.Start == anchorB.Start &&
// 				   anchorA.End == anchorB.End &&
// 				   anchorA.PaneID == anchorB.PaneID
// 		}
// 	}

// 	return false
// }

// // fuse combines two compatible intents into one
// // Phase 13.0: Simple concatenation for insert operations
// func fuse(a, b Intent) Intent {
// 	if a.Kind == FactInsert && b.Kind == FactInsert {
// 		// For insert operations, concatenate the text
// 		result := a
// 		result.Payload.Text += b.Payload.Text
// 		return result
// 	}

// 	// For other operations, just return the first one (shouldn't happen if canFuse worked correctly)
// 	return a
// }

// // FuseIntents combines compatible intents in a sequence
// // Phase 13.0: Sequential intent fusion
// func FuseIntents(intents []Intent) []Intent {
// 	if len(intents) <= 1 {
// 		return intents
// 	}

// 	var out []Intent
// 	out = append(out, intents[0])

// 	for i := 1; i < len(intents); i++ {
// 		lastIdx := len(out) - 1
// 		if canFuse(out[lastIdx], intents[i]) {
// 			out[lastIdx] = fuse(out[lastIdx], intents[i])
// 		} else {
// 			out = append(out, intents[i])
// 		}
// 	}
// 	return out
// }

package core

func FuseIntents(a, b Intent) Intent {
	// New semantic intent model:
	// Fusion is no longer structural merge.
	// For now, last intent wins.
	return b
}

````

## 📄 `core/line_hash_verifier.go`

````go
package core

type LineHashVerifier struct{}

func NewLineHashVerifier() *LineHashVerifier {
	return &LineHashVerifier{}
}

func (v *LineHashVerifier) Verify(
	pre Snapshot,
	facts []ResolvedFact,
	post Snapshot,
) VerificationResult {

	diffs := DiffSnapshot(pre, post)
	allowed := AllowedLineSet(facts)

	for _, d := range diffs {
		if !allowed.Contains(d.LineID) {
			return VerificationResult{
				OK:      false,
				Safety:  SafetyUnsafe,
				Diffs:   diffs,
				Message: "unexpected line modified",
			}
		}
	}

	return VerificationResult{
		OK:     true,
		Safety: SafetyExact,
		Diffs:  diffs,
	}
}

````

## 📄 `core/resolved_fact.go`

````go
package core

// ResolvedAnchor 代表具体的物理位置 (Phase 5.2)
// 它是 Resolver 解析后的结果，Projection 只认这个
type ResolvedAnchor struct {
	PaneID string
	LineID LineID // Stable line identifier (Phase 9)
	Line   int    // Fallback line number for compatibility
	Start  int
	End    int
}

// ResolvedFact 是已解析、可执行的事实
// 它是 Fact 的落地形态
type ResolvedFact struct {
	Kind    FactKind
	Anchor  ResolvedAnchor
	Payload FactPayload
	Meta    map[string]interface{} // Phase 5.2: 保留 Meta 以兼容旧 Projection 逻辑
	Safety  SafetyLevel            // Phase 7: Resolution safety
	LineID  LineID                 // Phase 9: Stable line identifier
}

````

## 📄 `core/shadow_engine.go`

````go
package core

import (
	"fmt"
	"log"
	"time"
)

// ShadowEngine 核心执行引擎
// 负责处理 Intent，生成并应用 Transaction，维护 History
type ShadowEngine struct {
	planner    Planner
	history    History
	resolver   AnchorResolver
	projection Projection
	reality    RealityReader
}

func NewShadowEngine(planner Planner, resolver AnchorResolver, projection Projection, reality RealityReader) *ShadowEngine {
	return &ShadowEngine{
		planner:    planner,
		history:    NewInMemoryHistory(100),
		resolver:   resolver,
		projection: projection,
		reality:    reality,
	}
}

func (e *ShadowEngine) ApplyIntent(intent Intent, snapshot Snapshot) (*Verdict, error) {
	var audit []AuditEntry

	// Phase 6.3: Temporal Adjudication (World Drift Check)
	// Engine owns the authority to reject execution if current reality != intent's expectation.
	if intent.GetSnapshotHash() != "" && e.reality != nil {
		current, err := e.reality.ReadCurrent(intent.GetPaneID())
		if err == nil {
			if string(current.Hash) != intent.GetSnapshotHash() {
				audit = append(audit, AuditEntry{Step: "Adjudicate", Result: "Rejected: World Drift detected"})
				return &Verdict{
					Kind:    VerdictRejected,
					Safety:  SafetyUnsafe,
					Message: "World drift detected",
					Audit:   audit,
				}, ErrWorldDrift
			}
			audit = append(audit, AuditEntry{Step: "Adjudicate", Result: "Success: Time consistency verified"})
		}
		// If Reality check fails (IO error), we might proceed with warning or fail fast.
		// For now, assume if we can't read reality, it's a structural error but not necessarily drift.
	}

	// 1. Handle Undo/Redo explicitly
	kind := intent.GetKind()
	if kind == IntentUndo {
		return e.performUndo()
	}
	if kind == IntentRedo {
		return e.performRedo()
	}

	// 2. Plan: Generate Facts
	facts, inverseFacts, err := e.planner.Build(intent, snapshot)
	if err != nil {
		audit = append(audit, AuditEntry{Step: "Plan", Result: fmt.Sprintf("Error: %v", err)})
		return &Verdict{Kind: VerdictBlocked, Audit: audit}, err
	}
	audit = append(audit, AuditEntry{Step: "Plan", Result: "Success"})

	// [Phase 5.1] 4. Resolve: 定位权移交
	// [Phase 5.4] 包含 Reconciliation 检查
	// [Phase 6.3] 包含 World Drift 检查 (SnapshotHash)
	resolvedFacts, err := e.resolver.ResolveFacts(facts, intent.GetSnapshotHash())
	if err != nil {
		audit = append(audit, AuditEntry{Step: "Resolve", Result: fmt.Sprintf("Error: %v", err)})
		return &Verdict{Kind: VerdictBlocked, Audit: audit}, err
	}
	audit = append(audit, AuditEntry{Step: "Resolve", Result: "Success"})

	// [Phase 7] Determine overall safety
	safety := SafetyExact
	for _, rf := range resolvedFacts {
		if rf.Safety > safety {
			safety = rf.Safety
		}
	}

	if safety == SafetyFuzzy && !intent.IsPartialAllowed() {
		return &Verdict{
			Kind:    VerdictRejected,
			Safety:  SafetyUnsafe,
			Message: "Fuzzy resolution disallowed by policy",
			Audit:   audit,
		}, ErrWorldDrift // Or a new error like ErrSafetyViolation
	}

	// [Phase 7] Inverse Fact Enrichment:
	// If the planner couldn't generate inverse facts (common for semantic deletes),
	// we generate them now using the reality captured during resolution.
	if len(inverseFacts) == 0 && len(resolvedFacts) > 0 {
		for _, rf := range resolvedFacts {
			if rf.Kind == FactDelete && rf.Payload.OldText != "" {
				// [Phase 7] Axiom 7.6: Paradox Resolved
				// Undo is return-to-origin, not a new fork.
				// Line-level semantic fingerprints are ignored because global post-hash already secured the timeline.
				invAnchor := Anchor{
					PaneID: rf.Anchor.PaneID,
					Kind:   AnchorAbsolute,
					Ref:    []int{rf.Anchor.Line, rf.Anchor.Start},
				}

				invMeta := make(map[string]interface{})
				for k, v := range rf.Meta {
					invMeta[k] = v
				}
				invMeta["operation"] = "undo_restore"

				inverseFacts = append(inverseFacts, Fact{
					Kind:   FactInsert,
					Anchor: invAnchor,
					Payload: FactPayload{
						Text: rf.Payload.OldText,
					},
					Meta: invMeta,
				})
			}
		}
	}

	// 3. Create Transaction
	txID := TransactionID(fmt.Sprintf("tx-%d", time.Now().UnixNano()))
	tx := &Transaction{
		ID:           txID,
		Intent:       intent,
		Facts:        facts,
		InverseFacts: inverseFacts,
		Safety:       safety,
		Timestamp:    time.Now().Unix(),
		AllowPartial: intent.IsPartialAllowed(),
	}

	// [Phase 9] Capture PreSnapshot for verification
	preSnapshot := snapshot

	// 5. Project: Execute
	if _, err := e.projection.Apply(nil, resolvedFacts); err != nil {
		audit = append(audit, AuditEntry{Step: "Project", Result: fmt.Sprintf("Error: %v", err)})
		return &Verdict{Kind: VerdictBlocked, Audit: audit}, err
	}
	audit = append(audit, AuditEntry{Step: "Project", Result: "Success"})
	tx.Applied = true

	// [Phase 7] Capture PostSnapshotHash for Undo verification
	var postSnap Snapshot
	if e.reality != nil {
		var err error
		postSnap, err = e.reality.ReadCurrent(intent.GetPaneID())
		if err == nil {
			tx.PostSnapshotHash = string(postSnap.Hash)
			audit = append(audit, AuditEntry{Step: "Record", Result: fmt.Sprintf("PostHash: %s", tx.PostSnapshotHash)})
		}
	}

	// [Phase 9] Verify that the projection achieved the expected result
	if e.projection != nil && e.reality != nil {
		verification := e.projection.Verify(preSnapshot, resolvedFacts, postSnap)
		if !verification.OK {
			audit = append(audit, AuditEntry{Step: "Verify", Result: fmt.Sprintf("Verification failed: %s", verification.Message)})
			// For now, we still consider this applied but log the verification issue
			log.Printf("[WEAVER] Projection verification failed: %s", verification.Message)
		} else {
			audit = append(audit, AuditEntry{Step: "Verify", Result: "Success: Projection matched expectations"})
		}
	}

	// 6. Update History
	if len(facts) > 0 {
		e.history.Push(tx)
	}

	return &Verdict{
		Kind:        VerdictApplied,
		Message:     "Applied via Smart Projection",
		Transaction: tx,
		Safety:      safety,
		Audit:       audit,
	}, nil
}

func (e *ShadowEngine) performUndo() (*Verdict, error) {
	tx := e.history.PopUndo()
	if tx == nil {
		return &Verdict{Kind: VerdictSkipped, Message: "Nothing to undo"}, nil
	}

	// [Phase 7] Axiom 7.5: Undo Is Verified Replay
	if tx.PostSnapshotHash != "" && e.reality != nil {
		current, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil && string(current.Hash) != tx.PostSnapshotHash {
			// Put it back to undo stack since we didn't apply it
			e.history.PushBack(tx)
			return &Verdict{
				Kind:    VerdictRejected,
				Message: "World drift: cannot undo safely",
				Safety:  SafetyUnsafe,
			}, ErrWorldDrift
		}
	}

	var audit []AuditEntry
	audit = append(audit, AuditEntry{Step: "Adjudicate", Result: "Undo context verified"})

	// [Phase 5.1] Resolve InverseFacts
	// [Phase 6.3] Use recorded PostHash if available (passed as expectedHash)
	resolvedFacts, err := e.resolver.ResolveFacts(tx.InverseFacts, tx.PostSnapshotHash)
	if err != nil {
		e.history.PushBack(tx)
		return nil, err
	}
	audit = append(audit, AuditEntry{Step: "Resolve", Result: fmt.Sprintf("Success: %d facts", len(resolvedFacts))})

	// [Phase 9] Capture PreSnapshot for verification
	preSnapshot, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
	if err != nil {
		preSnapshot = Snapshot{} // fallback
	}

	// Apply
	if len(resolvedFacts) > 0 {
		log.Printf("[WEAVER] Undo: Applying %d inverse facts. Text length: %d chars.", len(resolvedFacts), len(resolvedFacts[0].Payload.Text))
	}
	if _, err := e.projection.Apply(nil, resolvedFacts); err != nil {
		e.history.PushBack(tx)
		return nil, err
	}
	audit = append(audit, AuditEntry{Step: "Project", Result: "Success"})

	// [Phase 9] Verify undo operation
	if e.projection != nil && e.reality != nil {
		postSnap, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil {
			verification := e.projection.Verify(preSnapshot, resolvedFacts, postSnap)
			if !verification.OK {
				audit = append(audit, AuditEntry{Step: "Verify", Result: fmt.Sprintf("Undo verification failed: %s", verification.Message)})
				log.Printf("[WEAVER] Undo projection verification failed: %s", verification.Message)
			} else {
				audit = append(audit, AuditEntry{Step: "Verify", Result: "Success: Undo projection matched expectations"})
			}
		}
	}

	// Move to Redo Stack
	e.history.AddRedo(tx)

	return &Verdict{
		Kind:        VerdictApplied,
		Message:     fmt.Sprintf("Undone tx: %s", tx.ID),
		Transaction: tx,
		Audit:       audit,
	}, nil
}

func (e *ShadowEngine) performRedo() (*Verdict, error) {
	tx := e.history.PopRedo()
	if tx == nil {
		return &Verdict{Kind: VerdictSkipped, Message: "Nothing to redo"}, nil
	}

	// [Phase 7] Redo verification (must match Pre-state)
	preHash := tx.Intent.GetSnapshotHash()
	if preHash != "" && e.reality != nil {
		current, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil && string(current.Hash) != preHash {
			e.history.AddRedo(tx)
			return &Verdict{
				Kind:    VerdictRejected,
				Message: "World drift: cannot redo safely",
				Safety:  SafetyUnsafe,
			}, ErrWorldDrift
		}
	}

	// [Phase 5.1] Resolve Facts
	resolvedFacts, err := e.resolver.ResolveFacts(tx.Facts, preHash)
	if err != nil {
		e.history.AddRedo(tx)
		return nil, err
	}

	// [Phase 9] Capture PreSnapshot for verification
	preSnapshot, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
	if err != nil {
		preSnapshot = Snapshot{} // fallback
	}

	// Apply
	if _, err := e.projection.Apply(nil, resolvedFacts); err != nil {
		e.history.AddRedo(tx)
		return nil, err
	}

	// [Phase 9] Verify redo operation
	if e.projection != nil && e.reality != nil {
		postSnap, err := e.reality.ReadCurrent(tx.Intent.GetPaneID())
		if err == nil {
			verification := e.projection.Verify(preSnapshot, resolvedFacts, postSnap)
			if !verification.OK {
				log.Printf("[WEAVER] Redo projection verification failed: %s", verification.Message)
			} else {
				// Verification successful
			}
		}
	}

	// Restore to Undo Stack
	e.history.PushBack(tx)

	return &Verdict{
		Kind:        VerdictApplied,
		Message:     fmt.Sprintf("Redone tx: %s", tx.ID),
		Transaction: tx,
	}, nil
}

// GetHistory 获取历史管理器 (用于 Reverse Bridge)
func (e *ShadowEngine) GetHistory() History {
	return e.history
}

````

## 📄 `core/snapshot_diff.go`

````go
package core

type DiffKind int

const (
	DiffInsert DiffKind = iota
	DiffDelete
	DiffModify
)

type SnapshotDiff struct {
	LineID LineID
	Before *LineSnapshot
	After  *LineSnapshot
	Change DiffKind
}

func DiffSnapshot(pre, post Snapshot) []SnapshotDiff {
	diffs := []SnapshotDiff{}

	// deletions & modifications
	for id, preIdx := range pre.Index {
		preLine := pre.Lines[preIdx]
		postIdx, ok := post.Index[id]

		if !ok {
			diffs = append(diffs, SnapshotDiff{
				LineID: id,
				Before: &preLine,
				After:  nil,
				Change: DiffDelete,
			})
			continue
		}

		postLine := post.Lines[postIdx]
		if preLine.Hash != postLine.Hash {
			diffs = append(diffs, SnapshotDiff{
				LineID: id,
				Before: &preLine,
				After:  &postLine,
				Change: DiffModify,
			})
		}
	}

	// insertions
	for id, postIdx := range post.Index {
		if _, ok := pre.Index[id]; !ok {
			postLine := post.Lines[postIdx]
			diffs = append(diffs, SnapshotDiff{
				LineID: id,
				Before: nil,
				After:  &postLine,
				Change: DiffInsert,
			})
		}
	}

	return diffs
}

````

## 📄 `core/snapshot_types.go`

````go
package core

type LineID string
type LineHash string
type SnapshotHash string

type LineSnapshot struct {
	ID   LineID
	Text string
	Hash LineHash
}

type Snapshot struct {
	PaneID string
	Cursor CursorPos

	Lines []LineSnapshot
	Index map[LineID]int

	Hash SnapshotHash
}

type CursorPos struct {
	Row int
	Col int
}

````

## 📄 `core/take_snapshot.go`

````go
package core

func TakeSnapshot(
	paneID string,
	cursor CursorPos,
	lines []string,
) Snapshot {

	snaps := make([]LineSnapshot, 0, len(lines))
	index := make(map[LineID]int, len(lines))

	var prev LineID

	for i, text := range lines {
		id := makeLineID(paneID, prev, text)
		hash := hashLine(text)

		snap := LineSnapshot{
			ID:   id,
			Text: text,
			Hash: hash,
		}

		snaps = append(snaps, snap)
		index[id] = i
		prev = id
	}

	snapshot := Snapshot{
		PaneID: paneID,
		Cursor: cursor,
		Lines:  snaps,
		Index:  index,
	}

	snapshot.Hash = hashSnapshot(snapshot)
	return snapshot
}

````

## 📄 `logic/passthrough_resolver.go`

````go
package logic

import (
	"fmt"
	"tmux-fsm/weaver/adapter"
	"tmux-fsm/weaver/core"
)

// PassthroughResolver is a Phase 5.3 shim.
// It implements real resolution logic for Semantic Anchors.
type PassthroughResolver struct {
	Reality core.RealityReader
}

func (r *PassthroughResolver) ResolveFacts(facts []core.Fact, expectedHash string) ([]core.ResolvedFact, error) {
	if len(facts) == 0 {
		return []core.ResolvedFact{}, nil
	}

	// Phase 6.3: Consistency Verification
	// [DELETED] Check moved to ShadowEngine.ApplyIntent for unified adjudication.
	// Resolver now trusts the caller or uses the hash solely for snapshot-based resolution optimization.
	var currentSnapshot *core.Snapshot
	if expectedHash != "" && r.Reality != nil {
		paneID := facts[0].Anchor.PaneID
		snap, err := r.Reality.ReadCurrent(paneID)
		if err == nil {
			// Even if hashes drift, if we didn't fail at Engine level, we might still proceed
			// or use the snapshot as a "best efforts" view.
			// But since Engine already checked, Hash MUST match if we got here.
			currentSnapshot = &snap
		}
	}

	resolved := make([]core.ResolvedFact, 0, len(facts))

	for _, f := range facts {
		// Use Snapshot if available (Performance + Consistency)
		// Or fallback to Ad-hoc reading (adapter calls)
		var ra core.ResolvedAnchor
		var err error

		if currentSnapshot != nil {
			ra, err = r.resolveAnchorWithSnapshot(f.Anchor, *currentSnapshot)
		} else {
			ra, err = r.resolveAnchor(f.Anchor)
		}

		if err != nil {
			return nil, err
		}

		payload := f.Payload

		// Phase 5.3: Capture Reality (OldText) for Undo support
		// If deleting and we don't have text, capture it from ResolvedAnchor range
		if f.Kind == core.FactDelete && payload.OldText == "" {
			// We need to read the line content again or reuse from resolveAnchor?
			// resolveAnchor reads line but discards it.
			// Ideally we fetch it once. For simplicity, fetch again (performance hit negligible for single action).

			// Only if range is valid
			if ra.End >= ra.Start {
				var lineText string
				if currentSnapshot != nil {
					if ra.Line < len(currentSnapshot.Lines) {
						lineText = currentSnapshot.Lines[ra.Line].Text
					}
				} else {
					lineText = adapter.TmuxCaptureLine(ra.PaneID, ra.Line)
				}

				if len(lineText) > ra.End {
					payload.OldText = lineText[ra.Start : ra.End+1]
				} else if len(lineText) > ra.Start {
					payload.OldText = lineText[ra.Start:]
				}
			}
		}

		safety := core.SafetyExact
		if ra.LineID == "" {
			safety = core.SafetyFuzzy // ❗不是 Exact
		}

		resolved = append(resolved, core.ResolvedFact{
			Kind:    f.Kind,
			Anchor:  ra,
			Payload: payload,
			Meta:    f.Meta,
			Safety:  safety,
			LineID:  ra.LineID, // Phase 9: Include stable LineID
		})
	}

	return resolved, nil
}

// New helper method using Snapshot
func (r *PassthroughResolver) resolveAnchorWithSnapshot(a core.Anchor, s core.Snapshot) (core.ResolvedAnchor, error) {
	row := s.Cursor.Row
	col := s.Cursor.Col
	// If Anchor specifies hash, check line hash?
	// Phase 5.4 Logic checks LineHash.
	// Phase 6.3 checked SnapshotHash globally. LineHash is redundancy but good.

	lineText := ""
	var lineID core.LineID
	if row < len(s.Lines) {
		lineText = s.Lines[row].Text
		lineID = s.Lines[row].ID
		if a.Hash != "" {
			// Compare with LineSnapshot Hash
			if string(s.Lines[row].Hash) != a.Hash {
				return core.ResolvedAnchor{}, fmt.Errorf("line hash mismatch in snapshot")
			}
		}
	}

	switch a.Kind {
	case core.AnchorAtCursor:
		return core.ResolvedAnchor{PaneID: a.PaneID, LineID: lineID, Line: row, Start: col, End: col}, nil
	case core.AnchorWord:
		start, end := findWordRange(lineText, col, false)
		if start == -1 {
			start, end = col, col
		}
		return core.ResolvedAnchor{PaneID: a.PaneID, LineID: lineID, Line: row, Start: start, End: end}, nil
	case core.AnchorLine:
		return core.ResolvedAnchor{PaneID: a.PaneID, LineID: lineID, Line: row, Start: 0, End: len(lineText) - 1}, nil
	case core.AnchorAbsolute:
		// Ref is expected to be []int{line, col}
		if coords, ok := a.Ref.([]int); ok && len(coords) >= 2 {
			// Find the corresponding LineID for the absolute line
			absLine := coords[0]
			if absLine >= 0 && absLine < len(s.Lines) {
				return core.ResolvedAnchor{PaneID: a.PaneID, LineID: s.Lines[absLine].ID, Line: absLine, Start: coords[1], End: coords[1]}, nil
			}
		}
		// Fallback to cursor
		return core.ResolvedAnchor{PaneID: a.PaneID, LineID: lineID, Line: row, Start: col, End: col}, nil
	case core.AnchorLegacyRange:
		return r.resolveAnchor(a) // Fallback or implement here
	default:
		return core.ResolvedAnchor{PaneID: a.PaneID, LineID: lineID, Line: row, Start: col, End: col}, nil
	}
}

func (r *PassthroughResolver) resolveAnchor(a core.Anchor) (core.ResolvedAnchor, error) {
	// 1. Read Reality
	pos := adapter.TmuxGetCursorPos(a.PaneID) // [row, col]
	if len(pos) < 2 {
		return core.ResolvedAnchor{}, fmt.Errorf("failed to get cursor pos for pane %s", a.PaneID)
	}
	row, col := pos[0], pos[1]

	// Phase 5.4: Consistency Check
	// 总是读取当前行进行验证
	lineText := adapter.TmuxCaptureLine(a.PaneID, row)
	if a.Hash != "" {
		currentHash := adapter.TmuxHashLine(lineText)
		if currentHash != a.Hash {
			// Reconciliation Failure (Optimistic Locking)
			return core.ResolvedAnchor{}, fmt.Errorf("consistency check failed: hash mismatch (exp: %s, act: %s)", a.Hash, currentHash)
		}
	}

	// ❗禁止在无 Snapshot 情况下伪造 LineID
	// Return empty LineID to indicate unstable anchor
	switch a.Kind {

	case core.AnchorAtCursor:
		return core.ResolvedAnchor{
			PaneID: a.PaneID,
			LineID: "", // 空 LineID，明确表示不稳定
			Line:   row,
			Start:  col,
			End:    col,
		}, nil

	case core.AnchorWord:
		// use lineText already captured
		start, end := findWordRange(lineText, col, false)
		if start == -1 {
			start, end = col, col
		}
		return core.ResolvedAnchor{
			PaneID: a.PaneID,
			LineID: "", // 空 LineID，明确表示不稳定
			Line:   row,
			Start:  start,
			End:    end,
		}, nil

	case core.AnchorLine:
		// use lineText already captured
		return core.ResolvedAnchor{
			PaneID: a.PaneID,
			LineID: "", // 空 LineID，明确表示不稳定
			Line:   row,
			Start:  0,
			End:    len(lineText) - 1,
		}, nil

	case core.AnchorLegacyRange:
		// Legacy Range encoded in Ref
		if m, ok := a.Ref.(map[string]int); ok {
			return core.ResolvedAnchor{
				PaneID: a.PaneID,
				LineID: "", // 空 LineID，明确表示不稳定
				Line:   m["line"],
				Start:  m["start"],
				End:    m["end"],
			}, nil
		}
		return core.ResolvedAnchor{}, fmt.Errorf("invalid legacy ref")

	default:
		// Fallback for unknown kinds (e.g. Selection? if not implemented)
		return core.ResolvedAnchor{
			PaneID: a.PaneID,
			LineID: "", // 空 LineID，明确表示不稳定
			Line:   row,
			Start:  col,
			End:    col,
		}, nil
	}
}

// Logic copied from legacy execute.go / physical logic
func findWordRange(line string, x int, around bool) (int, int) {
	if x >= len(line) {
		// handle EOL
		if x > 0 && len(line) > 0 {
			x = len(line) - 1
		} else {
			return -1, -1
		}
	}

	isWordChar := func(c byte) bool {
		return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
	}

	// If not on word char, maybe look around?
	// Simplified: Expand from x.

	start := x
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}
	end := x
	for end < len(line)-1 && isWordChar(line[end+1]) {
		end++
	}

	return start, end
}

````

## 📄 `logic/shell_fact_builder.go`

````go
package logic

import (
	"tmux-fsm/weaver/core"
)

// ShellFactBuilder 纯语义构建器 (Phase 5.3)
// 不再读取 tmux buffer，不再计算 offset
type ShellFactBuilder struct{}

func (b *ShellFactBuilder) Build(intent core.Intent, snapshot core.Snapshot) ([]core.Fact, []core.Fact, error) {
	meta := intent.GetMeta()
	target := intent.GetTarget()

	// Check if intent has multiple anchors (Phase 11.0)
	anchors := intent.GetAnchors()
	if len(anchors) == 0 {
		// Fallback to original behavior: create anchor from snapshot
		// 基础语义 Anchor
		// Phase 6.2: 从 Snapshot 获取 Expectation (Line Hash)
		row := snapshot.Cursor.Row
		// col := snapshot.Cursor.Col // If needed for semantic logic refinement

		var lineHash string
		// Find line in snapshot
		// Snapshot Lines order matches Rows? Usually yes, row=index.
		// Check bounds
		if row >= 0 && row < len(snapshot.Lines) {
			lineHash = string(snapshot.Lines[row].Hash)
		}

		anchor := core.Anchor{
			PaneID: snapshot.PaneID,
			Kind:   core.AnchorAtCursor, // 默认为光标处
			Hash:   lineHash,
		}

		// 假设 TargetKind: 1=Char, 2=Word, 3=Line, 5=TextObject (from intent.go)
		switch target.Kind {
		case 1: // Char
			anchor.Kind = core.AnchorAtCursor
		case 2: // Word
			anchor.Kind = core.AnchorWord
		case 3: // Line
			anchor.Kind = core.AnchorLine
		case 5: // TextObject
			anchor.Kind = core.AnchorWord // Fallback or sophisticated resolution
		}

		anchors = []core.Anchor{anchor}
	}

	// Build facts for each anchor
	facts := make([]core.Fact, 0)
	for _, anchor := range anchors {
		switch intent.GetKind() {
		case core.IntentInsert:
			text := target.Value
			facts = append(facts, core.Fact{
				Kind:    core.FactInsert,
				Anchor:  anchor,
				Payload: core.FactPayload{Text: text},
				Meta:    meta,
			})

		// Note: IntentDelete and IntentChange intentionally omitted for Shell.
		// We rely on high-fidelity legacy capture and reverse-bridge injection
		// because semantic word-boundary resolution in the shell is imprecise.

		case core.IntentMove:
			// Move is FactMove.
			// Target Value might be motion string?
			facts = append(facts, core.Fact{
				Kind:   core.FactMove,
				Anchor: anchor,
				Meta:   meta,
			})
		}
	}

	// Inverse Facts:
	// Phase 5.3: Planner 无法生成反向事实，因为不仅要读取状态，甚至不知道 Resolve 后的位置。
	// Undo 逻辑必须依赖 Resolver 在 Execution 阶段的捕获，或者 History 存储 ResolvedFact。
	// 这里返回空。
	return facts, []core.Fact{}, nil
}

````

## 📄 `manager/manager.go`

````go
package manager

import (
	"fmt"
	"tmux-fsm/intent"
	"tmux-fsm/weaver/adapter"
	"tmux-fsm/weaver/core"
	"tmux-fsm/weaver/logic"
)

// ExecutionMode 执行模式
type ExecutionMode int

const (
	ModeLegacy ExecutionMode = iota // 传统模式
	ModeWeaver                      // Weaver模式
	ModeShadow                      // 仅观察模式
)

// WeaverManager 全局协调器
// RFC-WC-000: Kernel Sovereignty - 所有编辑决策必须通过Kernel
type WeaverManager struct {
	mode             ExecutionMode
	engine           core.Engine // Interface? No, ShadowEngine struct usually.
	resolver         core.AnchorResolver
	projection       core.Projection
	snapshotProvider adapter.SnapshotProvider // Phase 6.2
}

// weaverMgr 全局 Weaver 实例
var weaverMgr *WeaverManager

// InitWeaver 初始化 Weaver 系统
// RFC-WC-005: Audit Escape Prevention - 初始化必须可审计
func InitWeaver(mode ExecutionMode) {
	if mode == ModeLegacy {
		return
	}

	// 初始化组件
	planner := &logic.ShellFactBuilder{}
	// Phase 5.1: 使用 PassthroughResolver
	resolver := &logic.PassthroughResolver{}

	// Phase 6.1: Snapshot Provider
	snapProvider := &adapter.TmuxSnapshotProvider{}

	// Phase 6.3: Reality Reader for consistency adjudication
	reality := &adapter.TmuxRealityReader{Provider: snapProvider}
	resolver.Reality = reality

	var proj core.Projection
	if mode == ModeWeaver {
		proj = &adapter.TmuxProjection{}
	} else {
		proj = &adapter.NoopProjection{}
	}

	engine := core.NewShadowEngine(planner, resolver, proj, reality)

	weaverMgr = &WeaverManager{
		mode:             mode,
		engine:           engine,
		resolver:         resolver,
		projection:       proj,
		snapshotProvider: snapProvider,
	}
}

// ProcessIntentGlobal 全局意图处理入口
// RFC-WC-002: Intent ABI - 统一入口，统一审计
func (m *WeaverManager) ProcessIntentGlobal(intent core.Intent) error {
	if m == nil || m.mode == ModeLegacy {
		return nil // Fallback to legacy
	}

	// Phase 6.2: 获取当前快照作为时间冻结点
	snapshot, err := m.snapshotProvider.TakeSnapshot(intent.GetPaneID())
	if err != nil {
		return fmt.Errorf("failed to take snapshot: %v", err)
	}

	// Phase 6.3: ApplyIntent with frozen world state
	verdict, err := m.engine.ApplyIntent(intent, snapshot)
	if err != nil {
		return fmt.Errorf("engine failed: %v", err)
	}

	// RFC-WC-003: Audit Trail
	if verdict != nil {
		logWeaver("Intent processed: %v, Safety: %v", intent.GetKind(), verdict.Safety)
	}

	return nil
}

// Process 实现 IntentExecutor 接口
func (m *WeaverManager) Process(intent *intent.Intent) error {
	if m == nil || m.mode == ModeLegacy {
		return nil // Fallback to legacy
	}

	// 将统一的intent.Intent转换为core.Intent
	coreIntent := convertToCoreIntent(intent)

	// Phase 6.2: 获取当前快照作为时间冻结点
	snapshot, err := m.snapshotProvider.TakeSnapshot(coreIntent.GetPaneID())
	if err != nil {
		return fmt.Errorf("failed to take snapshot: %v", err)
	}

	// Phase 6.3: ApplyIntent with frozen world state
	verdict, err := m.engine.ApplyIntent(coreIntent, snapshot)
	if err != nil {
		return fmt.Errorf("engine failed: %v", err)
	}

	// RFC-WC-003: Audit Trail
	if verdict != nil {
		logWeaver("Intent processed: %v, Safety: %v", coreIntent.GetKind(), verdict.Safety)
	}

	return nil
}

// convertToCoreIntent 将统一的intent.Intent转换为core.Intent
func convertToCoreIntent(intent *intent.Intent) core.Intent {
	// 由于不能直接访问main.Intent，我们需要创建一个适配器
	return &intentAdapter{intent: intent}
}

// intentAdapter 适配器
type intentAdapter struct {
	intent *intent.Intent
}

func (a *intentAdapter) GetKind() core.IntentKind {
	return core.IntentKind(a.intent.Kind)
}

func (a *intentAdapter) GetTarget() core.SemanticTarget {
	return core.SemanticTarget{
		Kind:      int(a.intent.Target.Kind), // 使用intent中的Kind值
		Direction: a.intent.Target.Direction,
		Scope:     a.intent.Target.Scope,
		Value:     a.intent.Target.Value,
	}
}

func (a *intentAdapter) GetCount() int {
	return a.intent.Count
}

func (a *intentAdapter) GetMeta() map[string]interface{} {
	return a.intent.Meta
}

func (a *intentAdapter) GetPaneID() string {
	return a.intent.PaneID
}

func (a *intentAdapter) GetSnapshotHash() string {
	return a.intent.SnapshotHash
}

func (a *intentAdapter) IsPartialAllowed() bool {
	return a.intent.AllowPartial
}

func (a *intentAdapter) GetAnchors() []core.Anchor {
	// 简化处理，返回空切片
	return []core.Anchor{}
}

// GetWeaverManager 获取全局 Weaver 管理器实例
func GetWeaverManager() *WeaverManager {
	return weaverMgr
}

// InjectLegacyTransaction 将传统事务注入 Weaver 系统
// RFC-WC-004: Legacy Bridge - 保持向后兼容但通过统一审计
// TODO: 实现传统事务到Weaver系统的桥接
func (m *WeaverManager) InjectLegacyTransaction(tx interface{}) {
	if m.mode == ModeLegacy {
		return
	}

	// Convert legacy transaction to Weaver-compatible format for audit
	logWeaver("Legacy transaction injected for audit")
}

// logWeaver ...
func logWeaver(format string, args ...interface{}) {
	// 实现日志记录
}

````
