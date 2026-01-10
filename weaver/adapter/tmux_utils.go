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
	if paneID == "default" || paneID == "{current}" {
		paneID = ""
	}
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", paneID, "#{pane_cursor_x},#{pane_cursor_y}").Output()
	var x, y int
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d,%d", &x, &y)
	return [2]int{x, y}
}

// TmuxCaptureLine 获取指定行内容
func TmuxCaptureLine(paneID string, line int) string {
	if paneID == "default" || paneID == "{current}" {
		paneID = ""
	}
	out, _ := exec.Command("tmux", "capture-pane", "-p", "-t", paneID, "-J", "-S", fmt.Sprint(line), "-E", fmt.Sprint(line)).Output()
	return strings.TrimRight(string(out), "\n")
}

// TmuxCapturePane 获取整个面板内容 (Joined lines)
func TmuxCapturePane(paneID string) []string {
	if paneID == "default" || paneID == "{current}" {
		paneID = ""
	}
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
	if targetPane == "default" || targetPane == "{current}" {
		targetPane = ""
	}
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
	if targetPane == "default" || targetPane == "{current}" {
		targetPane = ""
	}
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", targetPane, "#{pane_cursor_y},#{pane_cursor_x}").Output()
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d,%d", &row, &col)
	return
}

// TmuxIsVimPane 检查是否是 Vim Pane
func TmuxIsVimPane(targetPane string) bool {
	if targetPane == "default" || targetPane == "{current}" {
		targetPane = ""
	}
	out, _ := exec.Command("tmux", "display-message", "-p", "-t", targetPane, "#{pane_current_command}").Output()
	cmd := strings.TrimSpace(string(out))
	return cmd == "vim" || cmd == "nvim" || cmd == "vi"
}
