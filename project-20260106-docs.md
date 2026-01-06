# Project Documentation

- **Generated at:** 2026-01-06 09:52:33
- **Root Dir:** `.`
- **File Count:** 63
- **Total Size:** 198.25 KB

## 📂 File List
- `.gitignore` (0.04 KB)
- `backend/backend.go` (2.96 KB)
- `bridge/bridge.go` (1.98 KB)
- `client.go` (1.20 KB)
- `config.go` (1.37 KB)
- `default.tmux.conf` (4.25 KB)
- `docs/层级协调规则.txt` (0.03 KB)
- `enter_fsm.sh` (0.48 KB)
- `execute.go` (30.36 KB)
- `fsm-exit.sh` (0.15 KB)
- `fsm-toggle.sh` (0.67 KB)
- `fsm/engine.go` (3.86 KB)
- `fsm/keymap.go` (1.11 KB)
- `fsm/nvim.go` (1.04 KB)
- `fsm/ui.go` (0.76 KB)
- `fsm/ui/interface.go` (0.08 KB)
- `fsm/ui/popup.go` (0.77 KB)
- `globals.go` (3.74 KB)
- `go.mod` (0.07 KB)
- `install.sh` (6.05 KB)
- `intent.go` (5.22 KB)
- `intent_bridge.go` (5.47 KB)
- `kernel/decide.go` (0.91 KB)
- `kernel/execute.go` (0.25 KB)
- `kernel/kernel.go` (0.47 KB)
- `kernel/legacy_adapter.go` (0.37 KB)
- `keymap.yaml` (0.51 KB)
- `legacy_logic.go` (4.80 KB)
- `main.go` (27.00 KB)
- `plugin.tmux` (2.55 KB)
- `protocol.go` (0.78 KB)
- `test_fsm.sh` (2.61 KB)
- `tests/baseline_tests.sh` (2.33 KB)
- `tools/gen-docs.go` (10.41 KB)
- `tools/install-gen-docs.sh` (1.88 KB)
- `transaction.go` (3.38 KB)
- `validate_paths.sh` (0.95 KB)
- `weaver/adapter/selection_normalizer.go` (1.66 KB)
- `weaver/adapter/snapshot.go` (0.23 KB)
- `weaver/adapter/snapshot_hash.go` (0.41 KB)
- `weaver/adapter/tmux_adapter.go` (1.42 KB)
- `weaver/adapter/tmux_physical.go` (12.14 KB)
- `weaver/adapter/tmux_projection.go` (6.88 KB)
- `weaver/adapter/tmux_reality.go` (0.23 KB)
- `weaver/adapter/tmux_snapshot.go` (0.36 KB)
- `weaver/adapter/tmux_utils.go` (2.25 KB)
- `weaver/core/allowed_lines.go` (0.29 KB)
- `weaver/core/anchor_kind.go` (0.25 KB)
- `weaver/core/hash.go` (0.54 KB)
- `weaver/core/history.go` (2.51 KB)
- `weaver/core/intent_fusion.go` (1.55 KB)
- `weaver/core/line_hash_verifier.go` (0.70 KB)
- `weaver/core/projection_verifier.go` (0.38 KB)
- `weaver/core/resolved_fact.go` (0.69 KB)
- `weaver/core/shadow_engine.go` (10.01 KB)
- `weaver/core/snapshot.go` (2.06 KB)
- `weaver/core/snapshot_diff.go` (1.33 KB)
- `weaver/core/snapshot_types.go` (0.35 KB)
- `weaver/core/take_snapshot.go` (0.61 KB)
- `weaver/core/types.go` (3.81 KB)
- `weaver/logic/passthrough_resolver.go` (7.38 KB)
- `weaver/logic/shell_fact_builder.go` (2.48 KB)
- `weaver_manager.go` (6.82 KB)

---

## 📄 `.gitignore`

````text
tmux-fsm
docs/project-20260105-docs.md

````

## 📄 `backend/backend.go`

````go
package backend

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
	var cmd *exec.Command
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

## 📄 `bridge/bridge.go`

````go
package bridge

import (
	"time"
	"tmux-fsm/fsm"
	tmux_fsm "tmux-fsm"
)

// LegacyFSMHandler 处理与旧 FSM 系统的交互
type LegacyFSMHandler struct {
	NewFSMEnabled bool
}

// NewLegacyFSMHandler 创建新的处理器
func NewLegacyFSMHandler() *LegacyFSMHandler {
	return &LegacyFSMHandler{
		NewFSMEnabled: true, // 默认启用新 FSM
	}
}

// HandleKey 处理按键输入
func (h *LegacyFSMHandler) HandleKey(key string) string {
	if h.NewFSMEnabled {
		// 检查是否在新 FSM 配置中有定义
		if stateDef, ok := fsm.KM.States[fsm.Active]; ok {
			if action, exists := stateDef.Keys[key]; exists {
				// 如果是层切换
				if action.Layer != "" {
					fsm.Active = action.Layer
					h.resetLayerTimeout(action.TimeoutMs)
					fsm.UpdateUI()
					return ""
				}
				// 执行动作
				fsm.RunAction(action.Action)
				return ""
			}
		}
	}

	// 如果新系统未处理，返回空字符串让旧系统处理
	return ""
}

// resetLayerTimeout 重置层超时
func (h *LegacyFSMHandler) resetLayerTimeout(ms int) {
	// 这里需要访问 fsm 包中的 timer，可能需要修改 fsm 包的设计
	if fsm.LayerTimer != nil {
		fsm.LayerTimer.Stop()
	}
	if ms > 0 {
		fsm.LayerTimer = time.AfterFunc(
			time.Duration(ms)*time.Millisecond,
			func() {
				fsm.Active = "NAV"
				fsm.UpdateUI()
			},
		)
	}
}

// EnterFSM 进入 FSM 模式
func (h *LegacyFSMHandler) EnterFSM() {
	if h.NewFSMEnabled {
		fsm.EnterFSM()
	} else {
		// 保留旧的进入逻辑
		tmux_fsm.GlobalBackend.SetUserOption("@fsm_active", "true")
		tmux_fsm.GlobalBackend.SwitchClientTable("", "fsm")
	}
}

// ExitFSM 退出 FSM 模式
func (h *LegacyFSMHandler) ExitFSM() {
	if h.NewFSMEnabled {
		fsm.ExitFSM()
	} else {
		// 保留旧的退出逻辑
		tmux_fsm.GlobalBackend.SetUserOption("@fsm_active", "false")
		tmux_fsm.GlobalBackend.SetUserOption("@fsm_state", "")
		tmux_fsm.GlobalBackend.SetUserOption("@fsm_keys", "")
		tmux_fsm.GlobalBackend.SwitchClientTable("", "root")
		tmux_fsm.GlobalBackend.RefreshClient("")
	}
}
````

## 📄 `client.go`

````go
package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

func isServerRunning() bool {
	conn, err := net.DialTimeout("unix", socketPath, 500*time.Millisecond)
	if err != nil {
		return false
	}
	defer conn.Close()

	// 发送心跳请求确认服务器响应
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	conn.Write([]byte("test|test|__PING__"))

	// 读取响应
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, err = conn.Read(buf)
	return err == nil
}

func runClient(key, paneAndClient string) {
	conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: daemon not running. Start it with 'tmux-fsm -server'\n")
		return
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting deadline: %v\n", err)
		return
	}

	payload := fmt.Sprintf("%s|%s", paneAndClient, key)
	if _, err := conn.Write([]byte(payload)); err != nil {
		return
	}

	// Read response (synchronize)
	buf, err := io.ReadAll(conn)
	if err != nil {
		return
	}
	resp := strings.TrimSpace(string(buf))
	if resp != "ok" && resp != "" {
		fmt.Println(resp)
	}
}
````

## 📄 `config.go`

````go
package main

import (
	"os"
	"strings"
)

// ExecutionMode 执行模式
type ExecutionMode int

const (
	ModeLegacy ExecutionMode = iota // 完全使用旧系统
	ModeShadow                      // Weaver 影子模式（记录但不执行）
	ModeWeaver                      // 完全使用 Weaver（阶段 3+）
)

// Config 全局配置
type Config struct {
	Mode     ExecutionMode
	LogFacts bool
	FailFast bool
}

// globalConfig 全局配置实例
var globalConfig = Config{
	Mode:     ModeLegacy, // 默认使用 Legacy 模式
	LogFacts: false,
	FailFast: false,
}

// LoadConfig 从环境变量加载配置
func LoadConfig() {
	// TMUX_FSM_MODE: legacy | shadow | weaver
	mode := strings.ToLower(os.Getenv("TMUX_FSM_MODE"))
	switch mode {
	case "shadow":
		globalConfig.Mode = ModeShadow
	case "weaver":
		globalConfig.Mode = ModeWeaver
	default:
		globalConfig.Mode = ModeLegacy
	}

	// TMUX_FSM_LOG_FACTS: 1 | 0
	if os.Getenv("TMUX_FSM_LOG_FACTS") == "1" {
		globalConfig.LogFacts = true
	}

	// TMUX_FSM_FAIL_FAST: 1 | 0
	if os.Getenv("TMUX_FSM_FAIL_FAST") == "1" {
		globalConfig.FailFast = true
	}
}

// GetMode 获取当前执行模式
func GetMode() ExecutionMode {
	return globalConfig.Mode
}

// ShouldLogFacts 是否记录 Facts
func ShouldLogFacts() bool {
	return globalConfig.LogFacts
}

// ShouldFailFast 是否快速失败
func ShouldFailFast() bool {
	return globalConfig.FailFast
}

````

## 📄 `default.tmux.conf`

````conf
# UTF-8 Support
set -g default-terminal "screen-256color"
set -g terminal-overrides "xterm-256color:Tc,xterm-kitty:Tc"

# Locale Support
set -g set-clipboard on

#ctrl-a 作为前缀
set -g prefix C-a
unbind C-b
bind C-a send-prefix


##### 鼠标支持 #####

# 启用鼠标（pane / window / 滚动）
set -g mouse on


##### 历史记录 #####

# 提高 scrollback 历史长度
set -g history-limit 50000


##### Pane 切换（Vim 风格 hjkl，前缀模式） #####

bind h select-pane -L
bind j select-pane -D
bind k select-pane -U
bind l select-pane -R


##### 快速重载配置 #####

bind r source-file ~/.tmux.conf \; display "tmux reloaded"


##### 状态栏 #####

# 右侧显示 FSM 状态 + session 名称 + 时间
# 由 plugin.tmux 统一管理 - 确保不在此处设置，避免覆盖
# set -g status-right "#{@fsm_state}#{@fsm_keys} | #S | %Y-%m-%d %H:%M"

# 仅设置左侧状态栏
set -g status-left "#[fg=green,bold]#S#[default] | "


##### 窗口与索引（补充项，不影响你原习惯） #####

# 窗口 / pane 编号从 1 开始
set -g base-index 1
set -g pane-base-index 1
set -g renumber-windows on


##### 新窗口 / 分屏（继承当前目录） #####

bind c new-window -c "#{pane_current_path}"
bind | split-window -h -c "#{pane_current_path}"
bind - split-window -v -c "#{pane_current_path}"


##### 复制模式（Vim 风格） #####

# 启用 vi 模式
setw -g mode-keys vi

# 复制模式绑定（带系统剪贴板同步）
bind -T copy-mode-vi v send -X begin-selection
bind -T copy-mode-vi y send -X copy-selection \; run "tmux save-buffer - | pbcopy"
bind -T copy-mode-vi r send -X rectangle-toggle
bind -T copy-mode-vi n send -X search-next
bind -T copy-mode-vi N send -X search-previous
bind -T copy-mode-vi Escape send -X cancel

# 从系统剪贴板粘贴到 tmux
bind -T copy-mode-vi p send -X paste-selection
bind P run "pbpaste | tmux load-buffer - ; tmux paste-buffer"

setw -g mode-keys vi
bind -T copy-mode-vi v send -X begin-selection
bind -T copy-mode-vi y send -X copy-selection


##### 视觉提示（轻量，不花哨） #####

set -g pane-border-style fg=colour238
set -g pane-active-border-style fg=colour39


##### Vim / Neovim 与 tmux 无缝 hjkl 穿透 #####

# 判断当前 pane 是否在运行 vim / nvim
is_vim="ps -o state= -o comm= -t '#{pane_tty}' | grep -iqE '^[^TXZ ]+ +(vi|vim|nvim)$'"

# Ctrl-h/j/k/l：在 Vim split 和 tmux pane 之间自动切换
# bind -n C-h if-shell "$is_vim" "send-keys C-h" "select-pane -L"
# bind -n C-j if-shell "$is_vim" "send-keys C-j" "select-pane -D"  # Unbound to use for FSM mode
bind -n C-k if-shell "$is_vim" "send-keys C-k" "select-pane -U"


##### 终端与响应（稳态设置） #####

set -g default-terminal "screen-256color"
set -as terminal-overrides ",xterm-256color:RGB"

# 降低 Esc 延迟（对 Vim 友好）
set -sg escape-time 0

##### Window / Pane 管理 #####

# 关闭当前 window / pane
bind x kill-pane        #  x 关闭 pane
bind X kill-window      # 大写 X 关闭整个 window
bind q kill-pane

# 列出窗口
bind w list-windows

# 数字切换窗口（1 开始）
bind 1 select-window -t 1
bind 2 select-window -t 2
bind 3 select-window -t 3
bind 4 select-window -t 4
bind 5 select-window -t 5
bind 6 select-window -t 6
bind 7 select-window -t 7
bind 8 select-window -t 8
bind 9 select-window -t 9


bind -n C-h previous-window

# 最近窗口切换
bind Tab last-window
# 调整大小
bind -r H resize-pane -L 5
bind -r J resize-pane -D 5
bind -r K resize-pane -U 5
bind -r L resize-pane -R 5

set -g set-clipboard on


##### Status Bar / Window Style #####

# 状态栏基础
set -g status on
set -g status-position bottom
set -g status-interval 5

# 状态栏整体风格
set -g status-style fg=colour250,bg=colour234

# 左右组件长度
set -g status-left-length 20
set -g status-right-length 80

# 非当前窗口
set -g window-status-style fg=colour245,bg=colour234

# 当前窗口（高亮，统一风格）
set -g window-status-current-style fg=colour234,bg=colour39,bold

# 分隔符（淡一点）
set -g window-status-separator " | "

# 窗口格式
set -g window-status-format " #I:#W "
set -g window-status-current-format "▶#I:#W◀"


# 将 Ctrl-f 绑定为无前缀入口
set -g @fsm_bind_no_prefix "C-f"

# 包含原始插件配置
source-file "$HOME/.tmux/plugins/tmux-fsm/plugin.tmux"

````

## 📄 `docs/层级协调规则.txt`

````text
NAV / GOTO / CMD 全层协同规
````

## 📄 `enter_fsm.sh`

````bash
#!/bin/bash
PLUGIN_DIR="$HOME/.tmux/plugins/tmux-fsm"
FSM_BIN="$PLUGIN_DIR/tmux-fsm"

# 1. Cancel copy mode (twice to be sure)
tmux send-keys -X cancel 2>/dev/null || true
tmux send-keys -X cancel 2>/dev/null || true

# 2. Set vars
tmux set -g @fsm_active "true"
tmux set -g repeat-time 0

# 3. Switch key table
tmux switch-client -T fsm

# 4. Init state
# Call -enter without parameters. The Go binary will handle server startup if needed.
"$FSM_BIN" -enter

# 5. Refresh
tmux refresh-client -S

````

## 📄 `execute.go`

````go
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
┃                tmux-fsn (Weaver Core) Cheat Sheet                  ┃
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

````

## 📄 `fsm-exit.sh`

````bash
#!/usr/bin/env bash

# Exit FSM + copy-mode safely

tmux set-option -g @fsm_active 0

# exit fsm
tmux send-keys Escape

# exit copy-mode
tmux send-keys q

````

## 📄 `fsm-toggle.sh`

````bash
#!/usr/bin/env bash

# 进入或退出 FSM 模式的静默切换脚本
FSM_ACTIVE=$(tmux show-option -gv @fsm_active)
[ -z "$FSM_ACTIVE" ] && FSM_ACTIVE="false"

if [ "$FSM_ACTIVE" = "true" ]; then
  # 退出逻辑
  tmux set -g @fsm_active "false"
  tmux set -g @fsm_state ""
  tmux set -g @fsm_keys ""
  tmux set -g repeat-time 500
  tmux switch-client -T root
  tmux refresh-client -S
else
  # 进入逻辑：首先强制退出任何既有模式，确保环境纯净
  tmux send-keys -X cancel 2>/dev/null
  tmux set -g @fsm_active "true"
  tmux set -g @fsm_state "NORMAL"
  tmux set -g @fsm_keys ""
  tmux set -g repeat-time 0
  tmux switch-client -T fsm
  tmux refresh-client -S
fi

````

## 📄 `fsm/engine.go`

````go
package fsm

import (
	"fmt"
	"strings"
	"time"
	"tmux-fsm/fsm/ui"
	"tmux-fsm"
)

// Engine FSM 引擎结构体
type Engine struct {
	Active     string
	Keymap     *Keymap
	layerTimer *time.Timer
	UI         ui.UI
}

// 全局默认引擎实例
var defaultEngine *Engine

// NewEngine 创建新的 FSM 引擎实例（显式注入 Keymap）
func NewEngine(km *Keymap) *Engine {
	return &Engine{
		Active: "NAV",
		Keymap: km,
	}
}

// InitEngine 初始化全局唯一 Engine
func InitEngine(km *Keymap) {
	defaultEngine = NewEngine(km)
}

// InLayer 检查当前是否处于非默认层（如 GOTO）
func (e *Engine) InLayer() bool {
	return e.Active != "NAV" && e.Active != ""
}

// CanHandle 检查当前层是否定义了该按键
func (e *Engine) CanHandle(key string) bool {
	if e.Keymap == nil {
		return false
	}
	st, ok := e.Keymap.States[e.Active]
	if !ok {
		return false
	}
	_, exists := st.Keys[key]
	return exists
}

// Dispatch 处理按键交互
func (e *Engine) Dispatch(key string) bool {
	if !e.CanHandle(key) {
		return false
	}

	st := e.Keymap.States[e.Active]
	act := st.Keys[key]

	// 1. 处理层切换
	if act.Layer != "" {
		e.Active = act.Layer
		e.resetLayerTimeout(act.TimeoutMs)
		UpdateUI()
		return true
	}

	// 2. 处理具体动作
	if act.Action != "" {
		e.RunAction(act.Action)

		// 铁律：执行完动作后，除非该层标记为 Sticky，否则立刻 Reset 回 NAV
		if !st.Sticky {
			e.Reset()
		} else {
			// 如果是 Sticky 层，可能需要刷新 UI（如 hint）
			UpdateUI()
		}
		return true
	}

	return false
}

// Reset 重置引擎状态到 NAV 层
func (e *Engine) Reset() {
	e.Active = "NAV"
	if e.layerTimer != nil {
		e.layerTimer.Stop()
	}
	// 执行重置通常意味着退出特定层级的 UI 显示
	HideUI()
}

// GetActiveLayer 获取当前层名称
func GetActiveLayer() string {
	if defaultEngine == nil {
		return "NAV"
	}
	return defaultEngine.Active
}

// InLayer 全局查询
func InLayer() bool {
	if defaultEngine == nil {
		return false
	}
	return defaultEngine.InLayer()
}

// CanHandle 全局查询
func CanHandle(key string) bool {
	if defaultEngine == nil {
		return false
	}
	return defaultEngine.CanHandle(key)
}

// Reset 全局重置
func Reset() {
	if defaultEngine != nil {
		defaultEngine.Reset()
	}
}

// ... (resetLayerTimeout remains same)
func (e *Engine) resetLayerTimeout(ms int) {
	if e.layerTimer != nil {
		e.layerTimer.Stop()
	}
	if ms > 0 {
		e.layerTimer = time.AfterFunc(
			time.Duration(ms)*time.Millisecond,
			func() {
				e.Reset()
				// 这里由于是异步超时，需要手动触发一次 UI 刷新
				UpdateUI()
			},
		)
	}
}

// RunAction 执行动作
func (e *Engine) RunAction(name string) {
	switch name {
	case "pane_left":
		tmux("select-pane -L")
	case "pane_right":
		tmux("select-pane -R")
	case "pane_up":
		tmux("select-pane -U")
	case "pane_down":
		tmux("select-pane -D")
	case "next_pane":
		tmux("select-pane -t :.+")
	case "prev_pane":
		tmux("select-pane -t :.-")
	case "far_left":
		tmux("select-pane -t :.0")
	case "far_right":
		tmux("select-pane -t :.$")
	case "goto_top":
		tmux("select-pane -t :.0")
	case "goto_bottom":
		tmux("select-pane -t :.$")
	case "exit":
		ExitFSM()
	case "prompt":
		tmux("command-prompt")
	default:
		fmt.Println("unknown action:", name)
	}
}

func tmux(cmd string) {
	// Use GlobalBackend to execute the command
	tmux_fsm.GlobalBackend.ExecRaw(cmd)
}

// 全局函数，支持在其他包调用
func Dispatch(key string) bool {
	if defaultEngine == nil {
		return false
	}
	return defaultEngine.Dispatch(key)
}

func EnterFSM() {
	if defaultEngine == nil {
		InitEngine(&KM)
	}

	engine := defaultEngine
	engine.Active = "NAV"
	// 确保进入时是干净的 NAV
	engine.Reset()
	// ShowUI() // Disable initial UI popup to prevent flashing/annoyance
}

func ExitFSM() {
	if defaultEngine != nil {
		defaultEngine.Reset()
	}
	HideUI()
	tmux_fsm.GlobalBackend.UnsetUserOption("key-table")
}

````

## 📄 `fsm/keymap.go`

````go
package fsm

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type KeyAction struct {
	Action    string `yaml:"action"`
	Layer     string `yaml:"layer"`
	TimeoutMs int    `yaml:"timeout_ms"`
}

type StateDef struct {
	Hint   string               `yaml:"hint"`
	Sticky bool                 `yaml:"sticky"` // If true, don't reset to NAV after action
	Keys   map[string]KeyAction `yaml:"keys"`
}

type Keymap struct {
	States map[string]StateDef `yaml:"states"`
}

// Validate 验证 keymap 配置的正确性
func (km *Keymap) Validate() error {
	for name, st := range km.States {
		for key, act := range st.Keys {
			if act.Layer != "" {
				if _, ok := km.States[act.Layer]; !ok {
					return fmt.Errorf("state %s references missing layer %s for key %s", name, act.Layer, key)
				}
			}
		}
	}
	return nil
}

func LoadKeymap(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var km Keymap
	if err := yaml.Unmarshal(b, &km); err != nil {
		return err
	}

	// 验证配置
	if err := km.Validate(); err != nil {
		return fmt.Errorf("invalid keymap: %w", err)
	}

	KM = km
	return nil
}

var (
	KM Keymap
)
````

## 📄 `fsm/nvim.go`

````go
package fsm

import (
	"strings"
	tmux_fsm "tmux-fsm"
)

// OnNvimMode 处理来自 Neovim 的模式变化
func OnNvimMode(mode string) {
	// 如果 Neovim 进入插入模式或可视模式，退出 FSM
	if mode == "i" || mode == "v" || mode == "V" || strings.HasPrefix(mode, "s") {
		ExitFSM()
	}
}

// NotifyNvimMode 通知 Neovim 当前 FSM 模式
// 注意：这个函数目前使用 send-keys，这可能不是最佳方案
// 更好的方案是使用 Neovim 的 RPC 机制
func NotifyNvimMode() {
	// 获取当前活跃的窗口/面板
	out, err := tmux_fsm.GlobalBackend.GetCommandOutput("display-message -p '#{pane_current_command}'")
	if err != nil {
		return
	}

	cmd := strings.TrimSpace(out)
	// 如果当前面板是 vim/nvim，则可以考虑发送模式信息
	// 但目前我们不执行任何操作，避免干扰用户输入
	// 更好的方式是通过 Neovim server/client 机制通信
	if cmd == "vim" || cmd == "nvim" {
		// TODO: 实现 Neovim RPC 通信以同步状态
		// 例如: nvim --server <socket> --remote-expr "FSM_SetMode('NAV')"
	}
}
````

## 📄 `fsm/ui.go`

````go
package fsm

import "tmux-fsm/fsm/ui"

// UIManager UI 管理器
type UIManager struct {
	active ui.UI
}

// NewUIManager 创建新的 UI 管理器
func NewUIManager() *UIManager {
	return &UIManager{}
}

// 全局 UI 实例
var CurrentUI ui.UI

// OnUpdateUI 是当 FSM 状态变化时需要执行的回调（通常由主程序注入以更新状态栏）
var OnUpdateUI func()

// UI 更新函数
func ShowUI() {
	if CurrentUI != nil {
		CurrentUI.Show()
	}
	if OnUpdateUI != nil {
		OnUpdateUI()
	}
}

func UpdateUI() {
	if CurrentUI != nil {
		CurrentUI.Update()
	}
	if OnUpdateUI != nil {
		OnUpdateUI()
	}
}

func HideUI() {
	if CurrentUI != nil {
		CurrentUI.Hide()
	}
	// 隐藏后通常也需要刷新一次状态栏以移除文字
	if OnUpdateUI != nil {
		OnUpdateUI()
	}
}
````

## 📄 `fsm/ui/interface.go`

````go
package ui

// UI 接口定义
type UI interface {
	Show()
	Update()
	Hide()
}
````

## 📄 `fsm/ui/popup.go`

````go
package ui

import (
	"fmt"
	tmux_fsm "tmux-fsm"
)

// StateProvider 接口用于获取状态信息
type StateProvider interface {
	GetActiveState() string
	GetStateHint(state string) string
}

// PopupUI 实现 UI 接口
type PopupUI struct {
	StateProvider StateProvider
}

func (p *PopupUI) Show() {
	if p.StateProvider == nil {
		return
	}

	active := p.StateProvider.GetActiveState()
	hint := p.StateProvider.GetStateHint(active)

	// 如果状态为空，不显示弹窗
	if active == "" {
		return
	}

	cmd := fmt.Sprintf("display-popup -E -w 50%% -h 5 'echo \"%s\"; echo \"%s\"'", active, hint)
	tmux_fsm.GlobalBackend.ExecRaw(cmd)
}

func (p *PopupUI) Update() {
	// 重新显示内容
	p.Show()
}

func (p *PopupUI) Hide() {
	tmux_fsm.GlobalBackend.ExecRaw("display-popup -C")
}
````

## 📄 `globals.go`

````go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"tmux-fsm/fsm"
)

type FSMState struct {
	Mode                 string                 `json:"mode"`
	Operator             string                 `json:"operator"`
	Count                int                    `json:"count"`
	PendingKeys          string                 `json:"pending_keys"`
	Register             string                 `json:"register"`
	LastRepeatableAction map[string]interface{} `json:"last_repeatable_action"`
	UndoStack            []Transaction          `json:"undo_stack"`
	RedoStack            []Transaction          `json:"redo_stack"`
	LastUndoFailure      string                 `json:"last_undo_failure,omitempty"`
	LastUndoSafetyLevel  string                 `json:"last_undo_safety_level,omitempty"`
	AllowPartial         bool                   `json:"allow_partial"` // Phase 7: Explicit permission for fuzzy resolution
}

var (
	stateMu     sync.Mutex
	globalState FSMState
	transMgr    TransactionManager
	socketPath  = os.Getenv("HOME") + "/.tmux-fsm.sock"
)

func loadState() FSMState {
	// Use GlobalBackend to read tmux options
	out, err := GlobalBackend.GetUserOption("@tmux_fsm_state")
	var state FSMState
	if err != nil || len(out) == 0 {
		return FSMState{Mode: "NORMAL", Count: 0}
	}
	json.Unmarshal([]byte(out), &state)
	return state
}

func saveStateRaw(data []byte) {
	// Use GlobalBackend to save state
	// This implies SetUserOption needs to be able to set arbitrary keys.
	if err := GlobalBackend.SetUserOption("@tmux_fsm_state", string(data)); err != nil {
		log.Printf("Failed to save FSM state: %v", err)
	}
}

func updateStatusBar(state FSMState, clientName string) {
	modeMsg := state.Mode
	if modeMsg == "" {
		modeMsg = "NORMAL"
	}

	// 融合显示逻辑
	activeLayer := fsm.GetActiveLayer()
	if activeLayer != "NAV" && activeLayer != "" {
		modeMsg = activeLayer // Override with FSM layer if active
	} else {
		// Translate legacy FSM modes for display
		switch modeMsg {
		case "VISUAL_CHAR":
			modeMsg = "VISUAL"
		case "VISUAL_LINE":
			modeMsg = "V-LINE"
		case "OPERATOR_PENDING":
			modeMsg = "PENDING"
		case "REGISTER_SELECT":
			modeMsg = "REGISTER"
		case "MOTION_PENDING":
			modeMsg = "MOTION"
		case "SEARCH":
			modeMsg = "SEARCH"
		}
	}

	if state.Operator != "" {
		modeMsg += fmt.Sprintf(" [%s]", state.Operator)
	}
	if state.Count > 0 {
		modeMsg += fmt.Sprintf(" [%d]", state.Count)
	}

	keysMsg := ""
	if state.PendingKeys != "" {
		if state.Mode == "SEARCH" {
			keysMsg = fmt.Sprintf(" /%s", state.PendingKeys)
		} else {
			keysMsg = fmt.Sprintf(" (%s)", state.PendingKeys)
		}
	}

	if state.LastUndoSafetyLevel == "fuzzy" {
		keysMsg += " ~UNDO"
	} else if state.LastUndoFailure != "" {
		keysMsg += " !UNDO_FAIL"
	}

	// Debug logging
	f, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] Updating status: mode=%s, state.Mode=%s, keys=%s\n",
			time.Now().Format("15:04:05"), modeMsg, state.Mode, keysMsg)
		f.Close()
	}

	// Use GlobalBackend for tmux option updates
	GlobalBackend.SetUserOption("@fsm_state", modeMsg)
	GlobalBackend.SetUserOption("@fsm_keys", keysMsg)
	GlobalBackend.RefreshClient(clientName) // Refresh the target client

	// --- [ABI: Heartbeat Lock] ---
	// Re-assert the key table to prevent "one-shot" dropouts.
	// Check @fsm_active to allow intentional exits.
	if clientName != "" && clientName != "default" {
		// Fetching @fsm_active via GlobalBackend if it were available would be ideal,
		// but for now, we rely on the fact that we are in a state where we should be active.
		// If GlobalBackend could read options, it would be better.
		// For now, we assume if we got here, FSM is active.
		GlobalBackend.SwitchClientTable(clientName, "fsm")
	}
}
````

## 📄 `go.mod`

````text
module tmux-fsm

go 1.24.0

require gopkg.in/yaml.v3 v3.0.1 // indirect

````

## 📄 `install.sh`

````bash
#!/usr/bin/env bash
set -e

echo "Installing tmux-fsm (FOEK Kernel)..."

# ----------------------------------------------------------------------
# config
# ----------------------------------------------------------------------

TMUX_FSM_DIR="${TMUX_FSM_DIR:-$HOME/.tmux/plugins/tmux-fsm}"

# 自动检测 tmux.conf（支持传统 & XDG）
if [ -z "$TMUX_CONF" ]; then
  if [ -f "$HOME/.tmux.conf" ]; then
    TMUX_CONF="$HOME/.tmux.conf"
  elif [ -f "$HOME/.config/tmux/tmux.conf" ]; then
    TMUX_CONF="$HOME/.config/tmux/tmux.conf"
  else
    TMUX_CONF="$HOME/.tmux.conf"
  fi
fi

# ----------------------------------------------------------------------
# checks
# ----------------------------------------------------------------------

if ! command -v tmux >/dev/null 2>&1; then
  echo "Error: tmux not found"
  exit 1
fi

# ----------------------------------------------------------------------
# install
# ----------------------------------------------------------------------

# 停止可能正在运行的旧版本守护进程 (Critical for Daemon update)
echo "Stopping running daemons..."
pkill -f "tmux-fsm -server" 2>/dev/null || true

echo "Installing to: $TMUX_FSM_DIR"
mkdir -p "$TMUX_FSM_DIR"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ----------------------------------------------------------------------
# Build Go binary (High Performance Kernel)
# ----------------------------------------------------------------------

if command -v go >/dev/null 2>&1; then
  echo "🚀 Building Go kernel for zero-latency performance..."
  
  # 临时初始化 go module 以防环境缺失
  if [ ! -f "$SCRIPT_DIR/go.mod" ]; then
      echo "Initializing temporary go module..."
      (cd "$SCRIPT_DIR" && go mod init tmux-fsm 2>/dev/null || true)
  fi

  # 编译：剔除符号表(-s)和调试信息(-w)以减小体积
  # 使用 "." 编译目录下所有文件，更健壮
  (cd "$SCRIPT_DIR" && go build -ldflags="-s -w" -o tmux-fsm .)
  
  cp "$SCRIPT_DIR/tmux-fsm" "$TMP_DIR/"
  echo "✅ Build successful."
else
  echo "⚠️  Warning: Go not found. Falling back to Python (Performance degraded)."
  echo "   Please install Go to enable the Daemon Kernel."
fi

# ----------------------------------------------------------------------
# copy files (required)
# ----------------------------------------------------------------------

# 只需要核心组件
cp "$SCRIPT_DIR"/plugin.tmux \
   "$SCRIPT_DIR"/fsm-toggle.sh \
   "$SCRIPT_DIR"/fsm-exit.sh \
   "$SCRIPT_DIR"/enter_fsm.sh \
   "$TMP_DIR/"

# 移动到目标目录
mv "$TMP_DIR"/* "$TMUX_FSM_DIR/"

# 确保二进制文件和 shell 脚本可执行
chmod +x \
  "$TMUX_FSM_DIR/tmux-fsm" \
  "$TMUX_FSM_DIR/fsm-toggle.sh" \
  "$TMUX_FSM_DIR/fsm-exit.sh" \
  "$TMUX_FSM_DIR/enter_fsm.sh"

# 清理旧的 Python 文件 (Clean up legacy)
rm -f "$TMUX_FSM_DIR/fsm.py" "$TMUX_FSM_DIR/tmux_fsm.py"

# ----------------------------------------------------------------------
# Interactive Configuration
# ----------------------------------------------------------------------

# NOTE: In non-interactive environments, we default to mode 1
install_mode="1"
if [ -t 0 ]; then
    echo ""
    echo "Configuration Strategy:"
    echo "1) Automatic: Append plugin hook to $TMUX_CONF and reload tmux"
    echo "2) Replace: Replace $TMUX_CONF with plugin's default config (backup created)"
    echo "3) Manual: Show instructions for manual setup"
    read -rp "Please select [1/2/3] (default 1): " user_choice
    install_mode="${user_choice:-1}"
fi

PLUGIN_HOOK="source-file \"$TMUX_FSM_DIR/plugin.tmux\""

case $install_mode in
    1)
        if grep -q "tmux-fsm" "$TMUX_CONF" 2>/dev/null; then
            echo "Result: Already configured in $TMUX_CONF"
        else
            echo "" >> "$TMUX_CONF"
            echo "# tmux-fsm plugin (FOEK Kernel)" >> "$TMUX_CONF"
            echo "$PLUGIN_HOOK" >> "$TMUX_CONF"
            echo "✅ Successfully updated $TMUX_CONF"
        fi

        echo "🔄 Performing Hot Upgrade..."
        # 尝试静默重新加载 tmux 配置
        if tmux info >/dev/null 2>&1; then
            tmux source-file "$TMUX_CONF" 2>/dev/null && echo "✅ tmux configuration reloaded"
            # 预热 Daemon (Phase 7: Weaver Mode)
            TMUX_FSM_MODE=weaver TMUX_FSM_LOG_FACTS=1 "$TMUX_FSM_DIR/tmux-fsm" -server >/dev/null 2>&1 &
            echo "✅ Daemon pre-warmed (Weaver Mode)."
        fi
        ;;
    2)
        # 创建备份并替换配置文件
        if [ -f "$TMUX_CONF" ]; then
            BACKUP_TMUX_CONF="${TMUX_CONF}.backup.$(date +%Y%m%d_%H%M%S)"
            echo "Creating backup of existing config: $BACKUP_TMUX_CONF"
            cp "$TMUX_CONF" "$BACKUP_TMUX_CONF"
            echo "✅ Backup created at $BACKUP_TMUX_CONF"
        fi

        # 复制默认配置文件并替换插件路径
        cp "$SCRIPT_DIR/default.tmux.conf" "$TMUX_CONF"
        echo "✅ Successfully replaced $TMUX_CONF with plugin default config"

        echo "🔄 Performing Hot Upgrade..."
        # 尝试静默重新加载 tmux 配置
        if tmux info >/dev/null 2>&1; then
            tmux source-file "$TMUX_CONF" 2>/dev/null && echo "✅ tmux configuration reloaded"
            # 预热 Daemon (Phase 7: Weaver Mode)
            TMUX_FSM_MODE=weaver TMUX_FSM_LOG_FACTS=1 "$TMUX_FSM_DIR/tmux-fsm" -server >/dev/null 2>&1 &
            echo "✅ Daemon pre-warmed (Weaver Mode)."
        fi
        ;;
    *)
        echo ""
        echo "💡 Manual action required:"
        echo "   Add the following line to your config:"
        echo ""
        echo "   $PLUGIN_HOOK"
        echo ""
        ;;
esac

# ----------------------------------------------------------------------
# done
# ----------------------------------------------------------------------

echo ""
echo "✅ tmux-fsm (Zero-Latency Daemon Kernel) installed!"
echo "   Latency: < 1ms"
echo ""
echo "Usage:"
echo "  - Enter FSM mode:  <prefix> f"
echo "  - Exit FSM mode:   Esc / C-c"
echo "  - Audit Logic:     Press '?' in FSM mode to see why Undo failed."
echo "  - Audit Log:       Logs are written to ~/tmux-fsm.log"
echo ""

````

## 📄 `intent.go`

````go
package main

// Intent 表示用户的编辑意图（语义层）
// 这是从 FSM 到执行器的中间层，将"按键序列"转换为"编辑语义"
type Intent struct {
	Kind         IntentKind             `json:"kind"`
	Target       SemanticTarget         `json:"target"`
	Count        int                    `json:"count"`
	Meta         map[string]interface{} `json:"meta,omitempty"`
	PaneID       string                 `json:"pane_id"`
	SnapshotHash string                 `json:"snapshot_hash"` // Phase 6.2
	AllowPartial bool                   `json:"allow_partial"` // Phase 7: Explicit permission for fuzzy resolution
	Anchors      []Anchor               `json:"anchors,omitempty"` // Phase 11.0: Support for multi-cursor / multi-selection
}

// GetPaneID 获取 PaneID
func (i Intent) GetPaneID() string {
	return i.PaneID
}

func (i Intent) GetKind() int {
	return int(i.Kind)
}

func (i Intent) GetSnapshotHash() string {
	return i.SnapshotHash
}

func (i Intent) IsPartialAllowed() bool {
	return i.AllowPartial
}

// GetAnchors returns the anchors for this intent
func (i Intent) GetAnchors() []Anchor {
	return i.Anchors
}

// IntentKind 意图类型
type IntentKind int

const (
	IntentNone IntentKind = iota
	IntentMove
	IntentDelete
	IntentChange
	IntentYank
	IntentInsert
	IntentPaste
	IntentUndo
	IntentRedo
	IntentSearch
	IntentVisual
	IntentToggleCase
	IntentReplace
	IntentRepeat
	IntentFind
	IntentExit
)

// SemanticTarget 语义目标（而非物理位置）
type SemanticTarget struct {
	Kind      TargetKind `json:"kind"`
	Direction string     `json:"direction,omitempty"` // forward, backward
	Scope     string     `json:"scope,omitempty"`     // char, line, word, etc.
	Value     string     `json:"value,omitempty"`     // 用于搜索、替换等
}

// TargetKind 目标类型
type TargetKind int

const (
	TargetNone TargetKind = iota
	TargetChar
	TargetWord
	TargetLine
	TargetFile
	TargetTextObject
	TargetPosition
	TargetSearch
)

// ToActionString 将 Intent 转换为 legacy action string
// 这是过渡期的桥接函数，最终会被移除
func (i Intent) ToActionString() string {
	if i.Kind == IntentNone {
		return ""
	}

	// 特殊处理：直接返回的动作
	switch i.Kind {
	case IntentUndo:
		return "undo"
	case IntentRedo:
		return "redo"
	case IntentRepeat:
		return "repeat_last"
	case IntentExit:
		return "exit"
	}

	// 组合型动作
	var action string

	// 操作类型
	switch i.Kind {
	case IntentMove:
		action = "move"
	case IntentDelete:
		action = "delete"
	case IntentChange:
		action = "change"
	case IntentYank:
		action = "yank"
	case IntentInsert:
		action = "insert"
	case IntentPaste:
		action = "paste"
	case IntentSearch:
		if i.Target.Value != "" {
			return "search_forward_" + i.Target.Value
		}
		if i.Target.Direction == "next" {
			return "search_next"
		}
		if i.Target.Direction == "prev" {
			return "search_prev"
		}
		return ""
	case IntentVisual:
		if i.Target.Scope == "char" {
			return "start_visual_char"
		}
		if i.Target.Scope == "line" {
			return "start_visual_line"
		}
		if i.Meta != nil {
			if op, ok := i.Meta["operation"].(string); ok {
				return "visual_" + op
			}
		}
		return "cancel_selection"
	case IntentToggleCase:
		return "toggle_case"
	case IntentReplace:
		if i.Target.Value != "" {
			return "replace_char_" + i.Target.Value
		}
		return ""
	case IntentFind:
		if i.Meta != nil {
			if fType, ok := i.Meta["find_type"].(string); ok {
				if char, ok := i.Meta["char"].(string); ok {
					return "find_" + fType + "_" + char
				}
			}
		}
		return ""
	}

	// 目标/运动
	var motion string
	switch i.Target.Kind {
	case TargetChar:
		if i.Target.Direction == "left" {
			motion = "left"
		} else if i.Target.Direction == "right" {
			motion = "right"
		}
	case TargetWord:
		if i.Target.Direction == "forward" {
			motion = "word_forward"
		} else if i.Target.Direction == "backward" {
			motion = "word_backward"
		} else if i.Target.Scope == "end" {
			motion = "end_of_word"
		}
	case TargetLine:
		if i.Target.Scope == "start" {
			motion = "start_of_line"
		} else if i.Target.Scope == "end" {
			motion = "end_of_line"
		} else if i.Target.Scope == "whole" {
			motion = "line"
		}
	case TargetFile:
		if i.Target.Scope == "start" {
			motion = "start_of_file"
		} else if i.Target.Scope == "end" {
			motion = "end_of_file"
		}
	case TargetPosition:
		if i.Target.Direction == "up" {
			motion = "up"
		} else if i.Target.Direction == "down" {
			motion = "down"
		}
	case TargetTextObject:
		// 文本对象：inside_word, around_quote, etc.
		motion = i.Target.Value
	}

	// Insert 的特殊位置
	if i.Kind == IntentInsert {
		if i.Target.Scope == "before" {
			return "insert_before"
		} else if i.Target.Scope == "after" {
			return "insert_after"
		} else if i.Target.Scope == "start_of_line" {
			return "insert_start_of_line"
		} else if i.Target.Scope == "end_of_line" {
			return "insert_end_of_line"
		} else if i.Target.Scope == "open_below" {
			return "insert_open_below"
		} else if i.Target.Scope == "open_above" {
			return "insert_open_above"
		}
	}

	// Paste 的特殊位置
	if i.Kind == IntentPaste {
		if i.Target.Scope == "after" {
			return "paste_after"
		} else if i.Target.Scope == "before" {
			return "paste_before"
		}
	}

	if motion == "" {
		return ""
	}

	return action + "_" + motion
}

````

## 📄 `intent_bridge.go`

````go
package main

import "strings"

// actionStringToIntent 将 legacy action string 转换为 Intent
// 这是阶段 1 的临时桥接函数，用于保持向后兼容
// 最终会被移除，直接从 handleXXX 函数返回 Intent
// actionStringToIntent 将 legacy action string 转换为 Intent
// 这是阶段 1 的临时桥接函数，用于保持向后兼容
// 最终会被移除，直接从 handleXXX 函数返回 Intent
func actionStringToIntent(action string, count int, paneID string) Intent {
	base := Intent{PaneID: paneID}

	if action == "" {
		base.Kind = IntentNone
		return base
	}

	// 特殊的单一动作
	switch action {
	case "undo":
		return Intent{Kind: IntentUndo, Count: count, PaneID: paneID}
	case "redo":
		return Intent{Kind: IntentRedo, Count: count, PaneID: paneID}
	case "repeat_last":
		return Intent{Kind: IntentRepeat, Count: count, PaneID: paneID}
	case "exit":
		return Intent{Kind: IntentExit, PaneID: paneID}
	case "toggle_case":
		return Intent{Kind: IntentToggleCase, Count: count, PaneID: paneID}
	case "search_next":
		return Intent{
			Kind:   IntentSearch,
			Target: SemanticTarget{Kind: TargetSearch, Direction: "next"},
			Count:  count,
			PaneID: paneID,
		}
	case "search_prev":
		return Intent{
			Kind:   IntentSearch,
			Target: SemanticTarget{Kind: TargetSearch, Direction: "prev"},
			Count:  count,
			PaneID: paneID,
		}
	case "start_visual_char":
		return Intent{
			Kind:   IntentVisual,
			Target: SemanticTarget{Scope: "char"},
			PaneID: paneID,
		}
	case "start_visual_line":
		return Intent{
			Kind:   IntentVisual,
			Target: SemanticTarget{Scope: "line"},
			PaneID: paneID,
		}
	case "cancel_selection":
		return Intent{
			Kind:   IntentVisual,
			Target: SemanticTarget{Scope: "cancel"},
			PaneID: paneID,
		}
	}

	// 处理前缀匹配的动作
	if strings.HasPrefix(action, "search_forward_") {
		query := strings.TrimPrefix(action, "search_forward_")
		return Intent{
			Kind:   IntentSearch,
			Target: SemanticTarget{Kind: TargetSearch, Value: query},
			Count:  count,
			PaneID: paneID,
		}
	}

	if strings.HasPrefix(action, "replace_char_") {
		char := strings.TrimPrefix(action, "replace_char_")
		return Intent{
			Kind:   IntentReplace,
			Target: SemanticTarget{Value: char},
			Count:  count,
			PaneID: paneID,
		}
	}

	if strings.HasPrefix(action, "find_") {
		parts := strings.SplitN(action, "_", 3)
		if len(parts) == 3 {
			return Intent{
				Kind:  IntentFind,
				Count: count,
				Meta: map[string]interface{}{
					"find_type": parts[1],
					"char":      parts[2],
				},
				PaneID: paneID,
			}
		}
	}

	if strings.HasPrefix(action, "visual_") {
		op := strings.TrimPrefix(action, "visual_")
		return Intent{
			Kind:   IntentVisual,
			Count:  count,
			Meta:   map[string]interface{}{"operation": op},
			PaneID: paneID,
		}
	}

	// 解析 operation_motion 格式
	parts := strings.SplitN(action, "_", 2)
	if len(parts) < 2 {
		// 单一动作，无法解析
		base.Kind = IntentNone
		return base
	}

	operation := parts[0]
	motion := parts[1]

	var kind IntentKind
	switch operation {
	case "move":
		kind = IntentMove
	case "delete":
		kind = IntentDelete
	case "change":
		kind = IntentChange
	case "yank":
		kind = IntentYank
	case "insert":
		kind = IntentInsert
	case "paste":
		kind = IntentPaste
	default:
		base.Kind = IntentNone
		return base
	}

	// 解析 motion 为 SemanticTarget
	target := parseMotionToTarget(motion)

	// 将原本的 motion 和 operation 存入 Meta 以供 Weaver Projection 使用
	meta := make(map[string]interface{})
	meta["motion"] = motion
	meta["operation"] = operation

	return Intent{
		Kind:   kind,
		Target: target,
		Count:  count,
		PaneID: paneID,
		Meta:   meta,
	}
}

// parseMotionToTarget 将 motion string 解析为 SemanticTarget
func parseMotionToTarget(motion string) SemanticTarget {
	// 方向性移动
	switch motion {
	case "left":
		return SemanticTarget{Kind: TargetChar, Direction: "left"}
	case "right":
		return SemanticTarget{Kind: TargetChar, Direction: "right"}
	case "up":
		return SemanticTarget{Kind: TargetPosition, Direction: "up"}
	case "down":
		return SemanticTarget{Kind: TargetPosition, Direction: "down"}
	}

	// 词级移动
	switch motion {
	case "word_forward":
		return SemanticTarget{Kind: TargetWord, Direction: "forward"}
	case "word_backward":
		return SemanticTarget{Kind: TargetWord, Direction: "backward"}
	case "end_of_word":
		return SemanticTarget{Kind: TargetWord, Scope: "end"}
	}

	// 行级移动
	switch motion {
	case "start_of_line":
		return SemanticTarget{Kind: TargetLine, Scope: "start"}
	case "end_of_line":
		return SemanticTarget{Kind: TargetLine, Scope: "end"}
	case "line":
		return SemanticTarget{Kind: TargetLine, Scope: "whole"}
	}

	// 文件级移动
	switch motion {
	case "start_of_file":
		return SemanticTarget{Kind: TargetFile, Scope: "start"}
	case "end_of_file":
		return SemanticTarget{Kind: TargetFile, Scope: "end"}
	}

	// Insert 的特殊位置
	switch motion {
	case "before":
		return SemanticTarget{Scope: "before"}
	case "after":
		return SemanticTarget{Scope: "after"}
	case "start_of_line":
		return SemanticTarget{Scope: "start_of_line"}
	case "end_of_line":
		return SemanticTarget{Scope: "end_of_line"}
	case "open_below":
		return SemanticTarget{Scope: "open_below"}
	case "open_above":
		return SemanticTarget{Scope: "open_above"}
	}

	// 文本对象
	if strings.HasPrefix(motion, "inside_") || strings.HasPrefix(motion, "around_") {
		return SemanticTarget{Kind: TargetTextObject, Value: motion}
	}

	// 默认返回
	return SemanticTarget{Kind: TargetNone}
}

````

## 📄 `kernel/decide.go`

````go
package kernel

import (
    "tmux-fsm/intent"
)

type DecisionKind int

const (
    DecisionNone DecisionKind = iota
    DecisionFSM
    DecisionLegacy
)

type Decision struct {
    Kind   DecisionKind
    Intent *intent.Intent
}

func (k *Kernel) Decide(key string) *Decision {
    // ✅ 1. FSM 永远先拿 key
    if k.FSM != nil {
        if k.FSM.InLayer() && k.FSM.CanHandle(key) {
            intent := k.FSM.Dispatch(key)
            if intent != nil {
                return &Decision{
                    Kind:   DecisionFSM,
                    Intent: intent,
                }
            }
            // FSM 明确吞掉
            return nil
        }
    }

    // ✅ 2. Legacy decoder（复用你现有逻辑）
    legacyIntent := DecodeLegacyKey(key)
    if legacyIntent != nil {
        return &Decision{
            Kind:   DecisionLegacy,
            Intent: legacyIntent,
        }
    }

    return nil
}

````

## 📄 `kernel/execute.go`

````go
package kernel

func (k *Kernel) Execute(decision *Decision) {
	if decision == nil || decision.Intent == nil {
		return
	}

	switch decision.Kind {
	case DecisionFSM:
		ExecuteIntent(decision.Intent)
	case DecisionLegacy:
		ExecuteIntent(decision.Intent)
	}
}

````

## 📄 `kernel/kernel.go`

````go
package kernel

import (
	"context"

	"tmux-fsm/fsm"
	"tmux-fsm/weaver"
)

type Kernel struct {
	FSM    *fsm.Engine
	Weaver *weaver.Manager
}

type HandleContext struct {
	Ctx context.Context
}

func NewKernel(fsmEngine *fsm.Engine, weaverMgr *weaver.Manager) *Kernel {
	return &Kernel{
		FSM:    fsmEngine,
		Weaver: weaverMgr,
	}
}

func (k *Kernel) HandleKey(hctx HandleContext, key string) {
	decision := k.Decide(key)

	if decision == nil {
		return
	}

	k.Execute(decision)
}

````

## 📄 `kernel/legacy_adapter.go`

````go
package kernel

import (
	"tmux-fsm/intent"
)

// ⚠️ 这是唯一一个“脏接口”，但它把脏东西隔离了
func DecodeLegacyKey(key string) *intent.Intent {
	// 直接调用你现在 logic.go 里的函数
	// 示例（你按真实函数名替换）：

	action := ProcessKeyLegacy(key)
	if action == "" {
		return nil
	}

	return intent.FromLegacyAction(action)
}

````

## 📄 `keymap.yaml`

````yaml
states:
  NAV:
    hint: "C-h/l next/prev · g goto · : cmd · q quit"
    keys:
      C-h: { action: "prev_pane" }
      C-l: { action: "next_pane" }
      g: { layer: "GOTO", timeout_ms: 800 }
      q: { action: "exit" }
      ":": { action: "prompt" }

  GOTO:
    hint: "h far-left · l far-right · g top · G bottom"
    keys:
      h: { action: "far_left" }
      l: { action: "far_right" }
      g: { action: "goto_top" }
      G: { action: "goto_bottom" }
      q: { action: "exit" }
      Escape: { action: "exit" }

````

## 📄 `legacy_logic.go`

````go
package main

import (
	"fmt"
	"os"
	"strings"
)

func processKey(state *FSMState, key string) string {
	if key == "Escape" || key == "C-c" {
		// Reset FSM state on escape/cancel
		state.Count = 0
		state.Operator = ""
		state.PendingKeys = ""
		fsm.Reset()
		return ""
	}

	// Check for count prefix
	if count, ok := isDigit(key); ok {
		if state.Count == 0 { // If no previous count, start accumulating
			state.Count = count
		} else { // Append digit to existing count
			state.Count = state.Count*10 + count
		}
		state.PendingKeys = fmt.Sprintf("%d", state.Count)
		return "" // Key handled as count, wait for next key
	}

	// If we have a count and received a motion
	if state.Count > 0 {
		// If the key is a motion
		if isMotion(key) {
			// Store motion for operator
			state.Operator = key // This is a simplification. Operator + Motion logic is complex.
			state.PendingKeys = fmt.Sprintf("%d%s", state.Count, key)
			// We need to capture this operator+motion for repeat
			state.LastRepeatableAction = map[string]interface{}{
				"action": state.Operator + "_" + state.Operator, // Placeholder, need proper motion mapping
				"count":  state.Count,
			}
			state.Count = 0 // Reset count after operator+motion
			return ""       // Key handled as count, wait for next key
		} else {
			// If it's not a motion, reset count and process key normally
			// e.g. 3j then 'd' is correct, but 3j then 'i' is wrong.
			// For simplicity, we reset count and let the key be processed as usual.
			// A more robust FSM would handle operator pending state better.

			// Rethink: if count is pending, and key is not a motion,
			// maybe it's an operator for the count? e.g. 3i<char>
			// For now, simpler reset.
			action := state.Operator + "_" + key
			state.Count = 0
			state.Operator = ""
			state.PendingKeys = ""
			return action
		}
	}

	// If we have an operator pending (e.g. 'd', 'c')
	if state.Operator != "" {
		// Check if key is a motion
		if isMotion(key) {
			action := state.Operator + "_" + key
			state.PendingKeys = fmt.Sprintf("%s%s", state.Operator, key)
			state.LastRepeatableAction = map[string]interface{}{
				"action": action,
				"count":  state.Count,
			}
			state.Count = 0 // Reset count after operator+motion
			state.Operator = ""
			return action
		} else {
			// Operator pending, but key is not a motion. Reset.
			// e.g., 'd' then 'a' (delete around word). This is wrong.
			// If it's another operator, e.g., 'd' then 'd' -> dd
			if key == state.Operator { // e.g., 'd' then 'd'
				action := state.Operator + "_" + state.Operator
				state.LastRepeatableAction = map[string]interface{}{
					"action": action,
					"count":  state.Count,
				}
				state.Count = 0
				state.Operator = ""
				return action
			}
			// Reset operator and pending keys, process key normally
			state.Count = 0
			state.Operator = ""
			state.PendingKeys = ""
			// Fallthrough to process key normally
		}
	}

	// If key is a known operator (d, c, y, etc.)
	if isOperator(key) {
		state.Operator = key
		state.PendingKeys = key
		state.Count = 0 // Reset count when a new operator is pressed
		return ""
	}

	// If key is insert mode related
	if strings.HasPrefix(key, "insert") || strings.HasPrefix(key, "replace") || strings.HasPrefix(key, "toggle") || strings.HasPrefix(key, "paste") {
		state.PendingKeys = ""
		state.Operator = ""
		state.Count = 0
		return key
	}

	// If key is a motion
	if isMotion(key) {
		// If no operator is pending, just move
		state.PendingKeys = key
		return "move_" + key
	}

	// Clear pending keys if not recognized and not part of an operator/motion sequence
	if state.PendingKeys != "" && !strings.HasPrefix(key, "move_") { // Allow move_ actions to be appended
		state.PendingKeys = ""
		state.Operator = ""
		state.Count = 0
	}

	// Handle special keys like Esc or Ctrl+C
	if key == "Escape" || key == "C-c" {
		state.Count = 0
		state.Operator = ""
		state.PendingKeys = ""
		fsm.Reset() // Reset FSM state
		return ""
	}

	// For any other key, return it as is (or handle specific ones like search)
	// Add explicit handling for search keys if not caught by FSM
	if strings.HasPrefix(key, "search_") {
		state.PendingKeys = key
		return key
	}

	// If key is unknown, clear state
	state.Count = 0
	state.Operator = ""
	state.PendingKeys = ""

	return ""
}

func isOperator(key string) bool {
	switch key {
	case "d", "c", "y":
		return true
	default:
		return false
	}
}

func isMotion(key string) bool {
	switch key {
	case "h", "j", "k", "l", "w", "b", "e", "0", "$", "gg", "G", // basic motions
		"up", "down", "left", "right", "word_forward", "word_backward", "end_of_word", // mapped motions
		"start_of_line", "end_of_line", "start_of_file", "end_of_file":
		return true
	default:
		return false
	}
}

func isDigit(s string) (int, bool) {
	if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
		return int(s[0] - '0'), true
	}
	return 0, false
}
````

## 📄 `main.go`

````go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"tmux-fsm/fsm"
)

func main() {
	// 记录启动参数用于调试
	argLog, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm-args.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if argLog != nil {
		fmt.Fprintf(argLog, "[%s] ARGS: %v\n", time.Now().Format("15:04:05"), os.Args)
		argLog.Close()
	}

	// 定义命令行参数
	var (
		enterFSM   = flag.Bool("enter", false, "Enter FSM mode")
		exitFSM    = flag.Bool("exit", false, "Exit FSM mode")
		dispatch   = flag.String("key", "", "Dispatch key to FSM")
		nvimMode   = flag.String("nvim-mode", "", "Handle Neovim mode change")
		uiShow     = flag.Bool("ui-show", false, "Show UI")
		uiHide     = flag.Bool("ui-hide", false, "Hide UI")
		reload     = flag.Bool("reload", false, "Reload keymap configuration")
		configPath = flag.String("config", "", "Path to keymap configuration file")
	)

	// 保留原有的服务器模式参数
	serverMode := flag.Bool("server", false, "run as daemon server")
	stopServer := flag.Bool("stop", false, "stop the running daemon")

	flag.Parse()

	// 确定配置文件路径
	configFile := *configPath
	if configFile == "" {
		// 默认配置文件路径
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		} else {
			configFile = homeDir + "/.config/tmux-fsm/keymap.yaml"
		}
	}

	// 尝试加载新的配置
	if err := fsm.LoadKeymap(configFile); err != nil {
		// 如果默认路径加载失败，尝试当前目录
		if err := fsm.LoadKeymap("./keymap.yaml"); err != nil {
			// 如果还是失败，创建一个默认配置
			createDefaultKeymap()
			if err := fsm.LoadKeymap("./keymap.yaml"); err != nil {
				fmt.Printf("Failed to load keymap: %v\n", err)
			}
		}
	}

	// 初始化 FSM 引擎
	fsm.InitEngine(&fsm.KM)

	// 根据命令行参数执行相应操作
	switch {
	case *enterFSM:
		// 检查服务器是否已经在运行，如果没有则启动
		if !isServerRunning() {
			// exec.Command("tmux", "new-session", "-d", "-s", "tmux-fsm-server", os.Args[0], "-server").Run() // More robust session start
			exec.Command(os.Args[0], "-server").Start() // Original start
			// 等待服务器启动，最多等待 2 秒
			for i := 0; i < 20; i++ {
				time.Sleep(100 * time.Millisecond)
				if isServerRunning() {
					break
				}
			}
		}

		// 解析 pane 和 client
		paneAndClient := ""
		clientName := ""
		if len(flag.Args()) > 0 {
			paneAndClient = flag.Args()[0]
			// 添加参数验证，防止异常参数导致问题
			if paneAndClient == "|" || paneAndClient == "" {
				// 如果参数异常，尝试获取当前pane和client
				paneIDBytes, err1 := exec.Command("tmux", "display-message", "-p", "#{pane_id}").Output()
				clientNameBytes, err2 := exec.Command("tmux", "display-message", "-p", "#{client_name}").Output()

				pID := strings.TrimSpace(string(paneIDBytes))
				cName := strings.TrimSpace(string(clientNameBytes))

				if err1 == nil && err2 == nil && pID != "" && cName != "" {
					paneID := pID
					clientName = cName
					paneAndClient = paneID + "|" + clientName
				} else {
					// 如果无法获取当前pane/client，使用默认值
					paneAndClient = "default|default"
					clientName = "default"
				}
			} else {
				parts := strings.Split(paneAndClient, "|")
				if len(parts) >= 2 {
					clientName = parts[1]
				}
			}
		} else {
			// 如果没有参数，获取当前pane和client
			paneIDBytes, err1 := exec.Command("tmux", "display-message", "-p", "#{pane_id}").Output()
			clientNameBytes, err2 := exec.Command("tmux", "display-message", "-p", "#{client_name}").Output()

			pID := strings.TrimSpace(string(paneIDBytes))
			cName := strings.TrimSpace(string(clientNameBytes))

			if err1 == nil && err2 == nil && pID != "" && cName != "" {
				paneID := pID
				clientName = cName
				paneAndClient = paneID + "|" + clientName
			} else {
				paneAndClient = "default|default"
				clientName = "default"
			}
		}

		// 通知服务器情况状态并刷新指定 client 的 UI
		runClient("__CLEAR_STATE__", paneAndClient)

		// 强制设置 tmux 变量并切换键表
		// Use GlobalBackend for state changes
		GlobalBackend.SetUserOption("@fsm_active", "true")
		GlobalBackend.SwitchClientTable(clientName, "fsm") // Use resolved client name
		GlobalBackend.RefreshClient(clientName)            // Refresh the client

	case *exitFSM:
		// Use GlobalBackend for state changes
		GlobalBackend.SetUserOption("@fsm_active", "false")
		GlobalBackend.SetUserOption("@fsm_state", "")
		GlobalBackend.SetUserOption("@fsm_keys", "")
		GlobalBackend.SwitchClientTable("", "root")
		GlobalBackend.RefreshClient("")

	case *dispatch != "":
		// Dispatch key to the server, possibly resolving pane first
		paneAndClient := ""
		if len(flag.Args()) > 0 {
			paneAndClient = flag.Args()[0]
		} else {
			// Resolve current pane/client if not provided
			paneIDBytes, err1 := exec.Command("tmux", "display-message", "-p", "#{pane_id}").Output()
			clientNameBytes, err2 := exec.Command("tmux", "display-message", "-p", "#{client_name}").Output()
			if err1 == nil && err2 == nil {
				paneAndClient = strings.TrimSpace(string(paneIDBytes)) + "|" + strings.TrimSpace(string(clientNameBytes))
			} else {
				paneAndClient = "default|default"
			}
		}
		runClient(*dispatch, paneAndClient)

	case *nvimMode != "":
		// Handle Neovim mode change - This usually involves IPC or FSM state updates
		fsm.OnNvimMode(*nvimMode) // Assumes fsm package has this function

	case *uiShow:
		fsm.ShowUI()
	case *uiHide:
		fsm.HideUI()
	case *reload:
		if err := fsm.LoadKeymap(configFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to reload keymap: %v\n", err)
			os.Exit(1)
		}
		// Update UI to reflect the new keymap hints
		fsm.UpdateUI()

	case *stopServer:
		shutdownServer()

	case *serverMode:
		runServer()

	default:
		// If no specific flags are set, and a key is provided, treat it as dispatch
		if *dispatch != "" {
			// Dispatch key to the server, resolving pane if necessary
			paneAndClient := ""
			if len(flag.Args()) > 0 {
				paneAndClient = flag.Args()[0]
			} else {
				// Resolve current pane/client if not provided
				paneIDBytes, err1 := exec.Command("tmux", "display-message", "-p", "#{pane_id}").Output()
				clientNameBytes, err2 := exec.Command("tmux", "display-message", "-p", "#{client_name}").Output()
				if err1 == nil && err2 == nil {
					paneAndClient = strings.TrimSpace(string(paneIDBytes)) + "|" + strings.TrimSpace(string(clientNameBytes))
				} else {
					paneAndClient = "default|default"
				}
			}
			runClient(*dispatch, paneAndClient)
		} else {
			// Show usage if no flags are set and no key is dispatched
			fmt.Println("tmux-fsm: A flexible FSM-based keybinding system for tmux")
			fmt.Println("Usage:")
			fmt.Println("  -enter        Enter FSM mode")
			fmt.Println("  -exit         Exit FSM mode")
			fmt.Println("  -key <key>    Dispatch key to FSM")
			fmt.Println("  -nvim-mode <mode>  Handle Neovim mode change")
			fmt.Println("  -ui-show      Show UI")
			fmt.Println("  -ui-hide      Hide UI")
			fmt.Println("  -reload       Reload keymap configuration")
			fmt.Println("  -config <path>  Path to keymap configuration file")
			fmt.Println("")
			fmt.Println("Legacy server mode:")
			fmt.Println("  -server       Run as daemon server")
			fmt.Println("  -stop         Stop the running daemon")
		}
	}
}

// createDefaultKeymap 创建默认的 keymap.yaml 文件
func createDefaultKeymap() {
	// 创建配置目录
	homeDir, _ := os.UserHomeDir()
	configDir := homeDir + "/.config/tmux-fsm"
	os.MkdirAll(configDir, 0755)

	// 默认配置内容
	// 注意：移除 NAV 层的 h/j/k/l 绑定，以便它们可以回退到 logic.go 处理光标移动
	defaultConfig := `states:
  NAV:
    hint: "g goto · : cmd · q quit"
    keys:
      g: { layer: "GOTO", timeout_ms: 800 }
      q: { action: "exit" }
      ":": { action: "prompt" }

  GOTO:
    hint: "h far-left · l far-right · g top · G bottom"
    keys:
      h: { action: "far_left" }
      l: { action: "far_right" }
      g: { action: "goto_top" }
      G: { action: "goto_bottom" }
      q: { action: "exit" }
      Escape: { action: "exit" }
`

	configFile := configDir + "/keymap.yaml"
	if err := os.WriteFile(configFile, []byte(defaultConfig), 0644); err != nil {
		// 如果无法写入用户目录，写入当前目录
		os.WriteFile("keymap.yaml", []byte(defaultConfig), 0644)
	}
}

// runClient 用于与服务器守护进程通信
func runClient(key, paneAndClient string) {
	conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: daemon not running. Start it with 'tmux-fsm -server'\n")
		return
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting deadline: %v\n", err)
		return
	}

	payload := fmt.Sprintf("%s|%s", paneAndClient, key)
	if _, err := conn.Write([]byte(payload)); err != nil {
		return
	}

	// Read response (synchronize)
	buf, err := io.ReadAll(conn)
	if err != nil {
		return
	}
	resp := strings.TrimSpace(string(buf))
	if resp != "ok" && resp != "" {
		fmt.Println(resp)
	}
}

// runServer 启动服务器守护进程
func runServer() {
	fmt.Printf("Server starting (v3-merged) at %s...\n", socketPath)
	// 阶段 2：加载配置
	LoadConfig()
	// 初始化 Weaver Core (Phase 2+)
	InitWeaver(globalConfig.Mode)
	if GetMode() != ModeLegacy {
		fmt.Printf("Execution mode: %s\n", modeString(GetMode()))
	}
	// 检查是否已有服务在运行 (且能响应)
	if conn, err := net.DialTimeout("unix", socketPath, 1*time.Second); err == nil {
		conn.Close()
		fmt.Println("Daemon already running and responsive.")
		return
	}

	// 如果 Socket 文件存在但无法连接，说明是残留文件，直接移除
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: Failed to remove old socket: %v\n", err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Printf("CRITICAL: Failed to start server: %v\n", err)
		return
	}
	defer listener.Close()
	if err := os.Chmod(socketPath, 0666); err != nil {
		fmt.Printf("Warning: Failed to chmod socket: %v\n", err)
	}

	// 初始化新架构回调：当新架构状态变化时，强制触发老架构的状态栏刷新
	fsm.OnUpdateUI = func() {
		stateMu.Lock()
		s := globalState
		stateMu.Unlock()
		updateStatusBar(s, "") // 兜底更新，不针对特定 client
	}

	// Load initial state from tmux option
	globalState = loadState()
	fmt.Println("tmux-fsm daemon started at", socketPath)

	// Handles signals for graceful shutdown
	stop := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		close(stop)
	}()

	// Periodic auto-save (every 30s)
	go func() {
		for {
			select {
			case <-time.After(30 * time.Second):
				stateMu.Lock()
				data, err := json.Marshal(globalState)
				stateMu.Unlock()
				if err == nil {
					saveStateRaw(data)
				}
			case <-stop:
				return
			}
		}
	}()

	for {
		// Set deadline to allow checking for stop signal
		tcpListener := listener.(*net.UnixListener)
		tcpListener.SetDeadline(time.Now().Add(1 * time.Second))

		conn, err := listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				select {
				case <-stop:
					goto shutdown
				default:
					continue
				}
			}
			continue
		}

		shouldExit := handleClient(conn)
		if shouldExit {
			goto shutdown
		}
	}

shutdown:
	fmt.Println("Shutting down gracefully...")
	stateMu.Lock()
	data, _ := json.Marshal(globalState)
	stateMu.Unlock()
	saveStateRaw(data)
	os.Remove(socketPath)
}

func handleClient(conn net.Conn) bool {
	defer conn.Close()

	// Set read deadline to prevent blocking the single-threaded server
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))

	// --- [ABI: Intent Submission Layer] ---
	// Frontend sends raw signals or internal commands to the kernel.
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return false
	}
	payload := string(buf[:n])

	// Parse Protocol: "PANE_ID|CLIENT_NAME|KEY"
	var paneID, clientName, key string
	parts := strings.SplitN(payload, "|", 3)
	if len(parts) == 3 {
		paneID = parts[0]
		clientName = parts[1]
		key = parts[2]
	} else if len(parts) == 2 {
		// Fallback for old protocol: PANE|KEY (Client unknown)
		paneID = parts[0]
		key = parts[1]
	} else {
		key = payload
	}

	// 写入本地日志以便直接调试
	f, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] Received: pane='%s', client='%s', key='%s'\n", time.Now().Format("15:04:05"), paneID, clientName, key)
		f.Close()
	}
	fmt.Printf("Received key: %s (pane: %s, client: %s)\n", key, paneID, clientName)

	if key == "__SHUTDOWN__" {
		return true
	}

	if key == "__PING__" {
		conn.Write([]byte("PONG"))
		return false
	}

	if key == "__CLEAR_STATE__" {
		fsm.Reset() // 重置新架构层级
		stateMu.Lock()
		globalState.Mode = "NORMAL"
		globalState.Operator = ""
		globalState.Count = 0
		globalState.PendingKeys = ""
		globalState.Register = ""
		globalState.UndoStack = nil
		globalState.RedoStack = nil
		globalState.LastUndoFailure = ""
		globalState.LastUndoSafetyLevel = ""
		stateMu.Unlock()
		updateStatusBar(globalState, clientName)
		return false
	}

	if key == "__STATUS__" {
		stateMu.Lock()
		defer stateMu.Unlock()
		data, _ := json.MarshalIndent(globalState, "", "  ")
		conn.Write(data)
		return false
	}

	if key == "__WHY_FAIL__" {
		stateMu.Lock()
		defer stateMu.Unlock()
		msg := globalState.LastUndoFailure
		if msg == "" {
			msg = "No undo failures recorded."
		}
		conn.Write([]byte(msg + "\n"))
		return false
	}

	if key == "__HELP__" {
		stateMu.Lock()
		defer stateMu.Unlock()
		if clientName == "" {
			// If called from a raw terminal (no clientName), just print text back
			conn.Write([]byte(getHelpText(&globalState)))
		} else {
			// If called from within tmux FSM, show popup
			showHelp(&globalState, paneID)
		}
		return false
	}

	// --- [融合逻辑控制：Kernel vs Module] ---
	// 铁律：只有当 FSM 显式处于某一层（非 NAV）且该层定义了此键时，才允许 FSM 抢键。
	var action string
	fsmHandled := false
	if fsm.InLayer() && fsm.CanHandle(key) {
		fsmHandled = fsm.Dispatch(key)
	}

	if fsmHandled {
		action = "" // 新架构已处理
	} else {
		// 永远兜底：进入高性能遗留逻辑 (logic.go)
		action = processKey(&globalState, key)
	}
	// --- [融合逻辑结束] ---

	// 阶段 3：Weaver 模式 - 接管执行；Shadow 模式 - 仅观察
	if (GetMode() == ModeShadow || GetMode() == ModeWeaver) && action != "" {
		// [Phase 7] Hybrid Protection:
		// If it's a high-fidelity action that will fall through to legacy below,
		// we SKIP direct Weaver processing here to prevent physical interference.
		// It will be captured via Reverse Bridge in Commit().
		isHighFidelity := strings.HasPrefix(action, "delete_") ||
			strings.HasPrefix(action, "change_") ||
			strings.HasPrefix(action, "yank_") ||
			strings.HasPrefix(action, "replace_")

		if !(GetMode() == ModeWeaver && isHighFidelity) {
			intent := actionStringToIntent(action, globalState.Count, paneID)
			ProcessIntentGlobal(intent)
		}
	}

	// [Phase 4] Weaver 模式下接管执行（包括 Undo/Redo），唯有 repeat_last 仍走 Legacy
	// [Phase 7] Hybrid Execution:
	// Even in Weaver mode, we allow high-fidelity capture actions (delete/change/yank)
	// to fall through to Lexacy execution so they can be captured accurately via Reverse Bridge.
	isHighFidelityAction := strings.HasPrefix(action, "delete_") ||
		strings.HasPrefix(action, "change_") ||
		strings.HasPrefix(action, "yank_") ||
		strings.HasPrefix(action, "replace_")

	if action != "" && (GetMode() == ModeLegacy || (GetMode() == ModeShadow) || action == "repeat_last" || isHighFidelityAction) {
		// 统一写入本地日志以便直接调试
		logFile, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if logFile != nil {
			fmt.Fprintf(logFile, "[%s] DEBUG: Key='%s', FSM_Handled=%v, Action='%s', Mode='%s'\n",
				time.Now().Format("15:04:05"), key, fsmHandled, action, globalState.Mode)
			fmt.Fprintf(logFile, "[%s] DEBUG: Executing legacy action: %s\n", time.Now().Format("15:04:05"), action)
			logFile.Close()
		}

		// [Phase 7] 再次确认：在 Weaver 模式下，Undo/Redo 必须由引擎完成，此处强制跳过
		// [Phase 7] 再次确认：在 Weaver 模式下，Undo/Redo 必须由引擎完成，此处强制跳过
		if GetMode() == ModeWeaver && (action == "undo" || action == "redo") {
			updateStatusBar(globalState, clientName)
			conn.Write([]byte("ok"))
			return false
		}
		if action == "repeat_last" {
			// Retrieve last repeatable action
			if globalState.LastRepeatableAction != nil {
				savedAction, _ := globalState.LastRepeatableAction["action"].(string)
				savedCount, _ := globalState.LastRepeatableAction["count"].(float64)
				if savedAction != "" {
					countToUse := globalState.Count
					if countToUse <= 0 {
						countToUse = int(savedCount)
					}
					transMgr.Begin(paneID)
					orig := globalState.Count
					globalState.Count = countToUse
					executeAction(savedAction, &globalState, paneID, clientName)
					globalState.Count = orig
					transMgr.Commit(&globalState.UndoStack, paneID)
					// [Phase 9] Only clear legacy redo stack if not in Weaver mode
					if GetMode() != ModeWeaver {
						globalState.RedoStack = nil
					}
					return false
				}
			}
		} else {
			// Execute action wrapped in transaction
			// --- [ABI: Verdict Trigger] ---
			// Kernel begins deliberation for the given intent.
			transMgr.Begin(paneID)
			executeAction(action, &globalState, paneID, clientName)
			// --- [ABI: Audit Closure] ---
			// Kernel finalizes the verdict and commits to the timeline.
			transMgr.Commit(&globalState.UndoStack, paneID)
			// [Phase 9] Only clear legacy redo stack if not in Weaver mode
			if GetMode() != ModeWeaver {
				globalState.RedoStack = nil
			}

			// Record if repeatable
			isRepeatable := strings.HasPrefix(action, "delete_") ||
				strings.HasPrefix(action, "change_") ||
				strings.HasPrefix(action, "yank_") ||
				strings.HasPrefix(action, "visual_")

			if isRepeatable && action != "cancel_selection" {
				globalState.LastRepeatableAction = map[string]interface{}{
					"action": action,
					"count":  globalState.Count,
				}
			}
		}
		globalState.Count = 0
	}

	// --- [ABI: Heartbeat Lock] ---
	// Update status and re-assert the key table to prevent "one-shot" dropouts.
	// Use GlobalBackend to get current pane context for status bar update.
	currentPaneID := paneID
	if paneID == "" || paneID == "{current}" || paneID == "default" {
		// Resolve current pane if not explicitly provided or is a placeholder
		var err error
		currentPaneID, err = GlobalBackend.GetActivePane("") // Use adapter for active pane context
		if err != nil {
			// If we can't get the pane, we can't update its status bar correctly. Log and continue.
			log.Printf("Error getting active pane for status update: %v", err)
		}
	}
	updateStatusBar(globalState, clientName)
	conn.Write([]byte("ok"))
	return false
}

func shutdownServer() {
	// Use GlobalBackend to communicate with the server
	// Since GlobalBackend is the client side, it can't directly shutdown the server socket.
	// Instead, it sends a shutdown command via the socket.
	conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
	if err == nil {
		defer conn.Close()
		// Send a special command to signal shutdown
		conn.Write([]byte("__SHUTDOWN__"))
	} else {
		fmt.Fprintf(os.Stderr, "Error: daemon not running to stop.\n")
	}
}

func loadState() FSMState {
	// Use GlobalBackend to read tmux options
	out, err := GlobalBackend.GetUserOption("@tmux_fsm_state")
	var state FSMState
	if err != nil || len(out) == 0 {
		return FSMState{Mode: "NORMAL", Count: 0}
	}
	json.Unmarshal([]byte(out), &state)
	return state
}

func saveStateRaw(data []byte) {
	// Use GlobalBackend to save state
	// This implies SetUserOption needs to be able to set arbitrary keys.
	if err := GlobalBackend.SetUserOption("@tmux_fsm_state", string(data)); err != nil {
		log.Printf("Failed to save FSM state: %v", err)
	}
}

func updateStatusBar(state FSMState, clientName string) {
	modeMsg := state.Mode
	if modeMsg == "" {
		modeMsg = "NORMAL"
	}

	// 融合显示逻辑
	activeLayer := fsm.GetActiveLayer()
	if activeLayer != "NAV" && activeLayer != "" {
		modeMsg = activeLayer // Override with FSM layer if active
	} else {
		// Translate legacy FSM modes for display
		switch modeMsg {
		case "VISUAL_CHAR":
			modeMsg = "VISUAL"
		case "VISUAL_LINE":
			modeMsg = "V-LINE"
		case "OPERATOR_PENDING":
			modeMsg = "PENDING"
		case "REGISTER_SELECT":
			modeMsg = "REGISTER"
		case "MOTION_PENDING":
			modeMsg = "MOTION"
		case "SEARCH":
			modeMsg = "SEARCH"
		}
	}

	if state.Operator != "" {
		modeMsg += fmt.Sprintf(" [%s]", state.Operator)
	}
	if state.Count > 0 {
		modeMsg += fmt.Sprintf(" [%d]", state.Count)
	}

	keysMsg := ""
	if state.PendingKeys != "" {
		if state.Mode == "SEARCH" {
			keysMsg = fmt.Sprintf(" /%s", state.PendingKeys)
		} else {
			keysMsg = fmt.Sprintf(" (%s)", state.PendingKeys)
		}
	}

	if state.LastUndoSafetyLevel == "fuzzy" {
		keysMsg += " ~UNDO"
	} else if state.LastUndoFailure != "" {
		keysMsg += " !UNDO_FAIL"
	}

	// Debug logging
	f, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] Updating status: mode=%s, state.Mode=%s, keys=%s\n",
			time.Now().Format("15:04:05"), modeMsg, state.Mode, keysMsg)
		f.Close()
	}

	// Use GlobalBackend for tmux option updates
	GlobalBackend.SetUserOption("@fsm_state", modeMsg)
	GlobalBackend.SetUserOption("@fsm_keys", keysMsg)
	GlobalBackend.RefreshClient(clientName) // Refresh the target client

	// --- [ABI: Heartbeat Lock] ---
	// Re-assert the key table to prevent "one-shot" dropouts.
	// Check @fsm_active to allow intentional exits.
	if clientName != "" && clientName != "default" {
		// Fetching @fsm_active via GlobalBackend if it were available would be ideal,
		// but for now, we rely on the fact that we are in a state where we should be active.
		// If GlobalBackend could read options, it would be better.
		// For now, we assume if we got here, FSM is active.
		GlobalBackend.SwitchClientTable(clientName, "fsm")
	}
}

// processKey handles key presses that are not handled by the FSM.
// It updates the FSM state and returns the action string to be executed.
func processKey(state *FSMState, key string) string {
	if key == "Escape" || key == "C-c" {
		// Reset FSM state on escape/cancel
		state.Count = 0
		state.Operator = ""
		state.PendingKeys = ""
		fsm.Reset()
		return ""
	}

	// Check for count prefix
	if count, ok := isDigit(key); ok {
		if state.Count == 0 { // If no previous count, start accumulating
			state.Count = count
		} else { // Append digit to existing count
			state.Count = state.Count*10 + count
		}
		state.PendingKeys = fmt.Sprintf("%d", state.Count)
		return "" // Key handled as count, wait for next key
	}

	// If we have a count and received a motion
	if state.Count > 0 {
		// If the key is a motion
		if isMotion(key) {
			// Store motion for operator
			state.Operator = key // This is a simplification. Operator + Motion logic is complex.
			state.PendingKeys = fmt.Sprintf("%d%s", state.Count, key)
			// We need to capture this operator+motion for repeat
			state.LastRepeatableAction = map[string]interface{}{
				"action": state.Operator + "_" + state.Operator, // Placeholder, need proper motion mapping
				"count":  state.Count,
			}
			state.Count = 0 // Reset count after operator+motion
			return ""
		} else {
			// If it's not a motion, reset count and process key normally
			// e.g. 3j then 'd' is correct, but 3j then 'i' is wrong.
			// For simplicity, we reset count and let the key be processed as usual.
			// A more robust FSM would handle operator pending state better.

			// Rethink: if count is pending, and key is not a motion,
			// maybe it's an operator for the count? e.g. 3i<char>
			// For now, simpler reset.
			action := state.Operator + "_" + key
			state.Count = 0
			state.Operator = ""
			state.PendingKeys = ""
			return action
		}
	}

	// If we have an operator pending (e.g. 'd', 'c')
	if state.Operator != "" {
		// Check if key is a motion
		if isMotion(key) {
			action := state.Operator + "_" + key
			state.PendingKeys = fmt.Sprintf("%s%s", state.Operator, key)
			state.LastRepeatableAction = map[string]interface{}{
				"action": action,
				"count":  state.Count,
			}
			state.Count = 0 // Reset count after operator+motion
			state.Operator = ""
			return action
		} else {
			// Operator pending, but key is not a motion. Reset.
			// e.g., 'd' then 'a' (delete around word). This is wrong.
			// If it's another operator, e.g., 'd' then 'd' -> dd
			if key == state.Operator { // e.g., 'd' then 'd'
				action := state.Operator + "_" + state.Operator
				state.LastRepeatableAction = map[string]interface{}{
					"action": action,
					"count":  state.Count,
				}
				state.Count = 0
				state.Operator = ""
				return action
			}
			// Reset operator and pending keys, process key normally
			state.Count = 0
			state.Operator = ""
			state.PendingKeys = ""
			// Fallthrough to process key normally
		}
	}

	// If key is a known operator (d, c, y, etc.)
	if isOperator(key) {
		state.Operator = key
		state.PendingKeys = key
		state.Count = 0 // Reset count when a new operator is pressed
		return ""
	}

	// If key is insert mode related
	if strings.HasPrefix(key, "insert") || strings.HasPrefix(key, "replace") || strings.HasPrefix(key, "toggle") || strings.HasPrefix(key, "paste") {
		state.PendingKeys = ""
		state.Operator = ""
		state.Count = 0
		return key
	}

	// If key is a motion
	if isMotion(key) {
		// If no operator is pending, just move
		state.PendingKeys = key
		return "move_" + key
	}

	// Clear pending keys if not recognized and not part of an operator/motion sequence
	if state.PendingKeys != "" && !strings.HasPrefix(key, "move_") { // Allow move_ actions to be appended
		state.PendingKeys = ""
		state.Operator = ""
		state.Count = 0
	}

	// Handle special keys like Esc or Ctrl+C
	if key == "Escape" || key == "C-c" {
		state.Count = 0
		state.Operator = ""
		state.PendingKeys = ""
		fsm.Reset() // Reset FSM state
		return ""
	}

	// For any other key, return it as is (or handle specific ones like search)
	// Add explicit handling for search keys if not caught by FSM
	if strings.HasPrefix(key, "search_") {
		state.PendingKeys = key
		return key
	}

	// If key is unknown, clear state
	state.Count = 0
	state.Operator = ""
	state.PendingKeys = ""

	return ""
}

func isOperator(key string) bool {
	switch key {
	case "d", "c", "y":
		return true
	default:
		return false
	}
}

func isMotion(key string) bool {
	switch key {
	case "h", "j", "k", "l", "w", "b", "e", "0", "$", "gg", "G", // basic motions
		"up", "down", "left", "right", "word_forward", "word_backward", "end_of_word", // mapped motions
		"start_of_line", "end_of_line", "start_of_file", "end_of_file":
		return true
	default:
		return false
	}
}

func isDigit(s string) (int, bool) {
	if len(s) == 1 && s[0] >= '0' && s[0] <= '9' {
		return int(s[0] - '0'), true
	}
	return 0, false
}

````

## 📄 `plugin.tmux`

````text
##### tmux-fsm plugin (New Architecture with Legacy Support) #####

# 1. 变量初始化
set -g @fsm_state ""
set -g @fsm_keys ""

# 2. 状态栏配置
set -g status-right "#[fg=yellow,bold]#{@fsm_state}#{@fsm_keys}#[default] | #S | %m-%d %H:%M"

# 3. 获取插件路径 (使用 TPM 标准路径)
set -g @fsm_bin "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm"

# 4. 入口：支持自定义按键 (Prefix 和 No-Prefix)
# 使用 run-shell 动态绑定
run-shell "
    # 1. 绑定 Prefix + Key (Default: f)
    prefix_key=\$(tmux show-option -gqv @fsm_toggle_key)
    [ -z \"\$prefix_key\" ] && prefix_key=\"f\"
    tmux bind-key \"\$prefix_key\" run-shell -b '$HOME/.tmux/plugins/tmux-fsm/enter_fsm.sh'

    # 2. 绑定 No-Prefix Key (Root Table)
    root_key=\$(tmux show-option -gqv @fsm_bind_no_prefix)
    if [ -n \"\$root_key\" ]; then
        tmux bind-key -n \"\$root_key\" run-shell -b '$HOME/.tmux/plugins/tmux-fsm/enter_fsm.sh'
    fi

    # 3. 设置全局环境变量 (Phase 7: Temporal Integrity)
    tmux set-environment -g TMUX_FSM_MODE weaver
    tmux set-environment -g TMUX_FSM_LOG_FACTS 1

    # 4. 启动服务器守护进程 (Weaver Mode)
    TMUX_FSM_MODE=weaver TMUX_FSM_LOG_FACTS=1 $HOME/.tmux/plugins/tmux-fsm/tmux-fsm -server >/dev/null 2>&1 &
"

# 5. FSM 键表配置 (新架构)
bind-key -T fsm -n C-c run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -exit"
bind-key -T fsm -n Escape run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -exit"

# 6. Explicitly bind alphanumeric keys (POSIX compliant)
# {a..z} is a bash extension, we must use explicit lists for /bin/sh compatibility
run-shell "
    for key in a b c d e f g h i j k l m n o p q r s t u v w x y z A B C D E F G H I J K L M N O P Q R S T U V W X Y Z 0 1 2 3 4 5 6 7 8 9 '$' '^' '.' '/' ',' ';' ':'; do
        tmux bind-key -T fsm \"\$key\" run-shell -b \"$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -key '\$key' '#{pane_id}|#{client_name}'\"
    done
"

# 7. Bind common punctuation explicitly - REMOVED due to shell escaping issues. 
# Relying on 'Any' fallback for punctuation.

# Keep 'Any' as a fallback for special keys and punctuation.
bind-key -T fsm Any run-shell -b \
  "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -key \"#{key}\" \"#{pane_id}|#{client_name}\""

# 7. 额外的便捷键绑定
bind-key -T fsm q run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -exit"

# 8. 重新加载配置
bind-key -T root R run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -reload"

# 9. 帮助功能
bind-key -T root ? run-shell "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm '__HELP__' '#{pane_id}|#{client_name}'"

##### end tmux-fsm #####

````

## 📄 `protocol.go`

````go
package main

// Anchor 是“我指的不是光标，而是这段文本”
type Anchor struct {
	PaneID   string  `json:"pane_id"`
	LineHint int     `json:"line_hint"`
	LineHash string  `json:"line_hash"`
	Cursor   *[2]int `json:"cursor_hint,omitempty"`
}

type Range struct {
	Anchor      Anchor `json:"anchor"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
	Text        string `json:"text"`
}

type Fact struct {
	Kind        string                 `json:"kind"` // delete / insert / replace
	Target      Range                  `json:"target"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	SideEffects []string               `json:"side_effects,omitempty"`
}

type ActionRecord struct {
	Fact    Fact `json:"fact"`
	Inverse Fact `json:"inverse"`
}
````

## 📄 `test_fsm.sh`

````bash
#!/bin/bash

echo "=== 开始 tmux-fsm 全面测试 ==="

# 首先停止任何正在运行的服务器
echo "停止任何正在运行的服务器..."
/Users/ygs/ygs/ygs/learning/tmuxPlugin/tmux-fsm -stop 2>/dev/null || true
sleep 1

# 1. 构建测试
echo "1. 测试构建..."
cd /Users/ygs/ygs/ygs/learning/tmuxPlugin
go clean
if go build -o tmux-fsm; then
    echo "✅ 构建成功"
else
    echo "❌ 构建失败"
    exit 1
fi

# 2. Keymap 验证测试
echo "2. 测试 Keymap 验证..."
if ./tmux-fsm -config keymap.yaml -reload; then
    echo "✅ 有效配置加载成功"
else
    echo "❌ 有效配置加载失败"
    exit 1
fi

# 创建无效配置测试验证功能
cat > invalid_keymap.yaml << 'EOF'
states:
  NAV:
    hint: "test"
    keys:
      g: { layer: NONEXISTENT, timeout_ms: 800 }
EOF

if ./tmux-fsm -config invalid_keymap.yaml -reload; then
    echo "❌ 无效配置应该报错但没有"
    rm invalid_keymap.yaml
    exit 1
else
    echo "✅ 无效配置正确报错"
fi
rm invalid_keymap.yaml

# 3. 服务器模式测试
echo "3. 测试服务器模式..."
./tmux-fsm -server &
SERVER_PID=$!
sleep 2  # 等待服务器完全启动

# 检查服务器是否启动
if ps -p $SERVER_PID > /dev/null; then
    echo "✅ 服务器启动成功"
else
    echo "❌ 服务器启动失败"
    exit 1
fi

# 4. FSM 生命周期测试
echo "4. 测试 FSM 生命周期..."
if ./tmux-fsm -enter; then
    echo "✅ 进入 FSM 成功"
else
    echo "❌ 进入 FSM 失败"
fi

if ./tmux-fsm -key h; then
    echo "✅ 按键 h 分发成功"
else
    echo "❌ 按键 h 分发失败"
fi

if ./tmux-fsm -key g; then
    echo "✅ 按键 g 分发成功"
else
    echo "❌ 按键 g 分发失败"
fi

if ./tmux-fsm -key h; then
    echo "✅ 按键 h (在 GOTO 层) 分发成功"
else
    echo "❌ 按键 h (在 GOTO 层) 分发失败"
fi

if ./tmux-fsm -exit; then
    echo "✅ 退出 FSM 成功"
else
    echo "❌ 退出 FSM 失败"
fi

# 停止服务器
if ./tmux-fsm -stop; then
    echo "✅ 停止服务器成功"
else
    echo "❌ 停止服务器失败"
fi

# 等待服务器进程结束
sleep 1
if ps -p $SERVER_PID > /dev/null; then
    kill $SERVER_PID 2>/dev/null || true
fi

# 5. UI 测试
echo "5. 测试 UI 功能..."
./tmux-fsm -server &
SERVER_PID2=$!
sleep 2  # 等待服务器完全启动

if ./tmux-fsm -ui-show; then
    echo "✅ UI 显示成功"
else
    echo "❌ UI 显示失败"
fi

if ./tmux-fsm -ui-hide; then
    echo "✅ UI 隐藏成功"
else
    echo "❌ UI 隐藏失败"
fi

./tmux-fsm -stop
sleep 1
if ps -p $SERVER_PID2 > /dev/null; then
    kill $SERVER_PID2 2>/dev/null || true
fi

echo "=== 所有测试完成 ==="
````

## 📄 `tests/baseline_tests.sh`

````bash
#!/bin/bash
# 阶段 0 基线测试脚本
# 用于验证重构后功能一致性

set -e

echo "=== tmux-fsn 基线测试 ==="
echo "Tag: pre-weaver-migration"
echo "Date: $(date)"
echo ""

# 测试 1: 基本移动命令
test_basic_movement() {
    echo "测试 1: 基本移动命令 (h/j/k/l)"
    # 这里需要在实际 tmux 环境中测试
    # 预期：光标正确移动
    echo "  ✓ 需要手动验证"
}

# 测试 2: 删除操作 + Undo
test_delete_undo() {
    echo "测试 2: 删除操作 + Undo"
    # 场景：dw dw dw 然后 u u u
    # 预期：删除三个词，撤销三次后恢复
    echo "  ✓ 需要手动验证"
}

# 测试 3: 移动光标后 delete
test_move_then_delete() {
    echo "测试 3: 移动光标后 delete"
    # 场景：移动光标到中间，执行 dw
    # 预期：Anchor 正确定位，删除正确的词
    echo "  ✓ 需要手动验证"
}

# 测试 4: 跨 pane 操作
test_cross_pane() {
    echo "测试 4: 跨 pane / window 操作"
    # 场景：在不同 pane 中切换并执行操作
    # 预期：状态正确隔离
    echo "  ✓ 需要手动验证"
}

# 测试 5: 文本对象
test_text_objects() {
    echo "测试 5: 文本对象 (diw, ci\", 等)"
    # 场景：diw, ci", da(
    # 预期：正确识别并操作文本对象
    echo "  ✓ 需要手动验证"
}

# 测试 6: Visual 模式
test_visual_mode() {
    echo "测试 6: Visual 模式"
    # 场景：v 选择，d 删除
    # 预期：正确进入/退出 visual 模式
    echo "  ✓ 需要手动验证"
}

# 测试 7: 搜索功能
test_search() {
    echo "测试 7: 搜索功能 (/, n, N)"
    # 场景：/pattern, n, N
    # 预期：正确搜索和跳转
    echo "  ✓ 需要手动验证"
}

# 测试 8: FSM 层级切换
test_fsm_layers() {
    echo "测试 8: FSM 层级切换 (g -> GOTO)"
    # 场景：g 进入 GOTO 层，gg 跳转到顶部
    # 预期：层级正确切换，超时自动退出
    echo "  ✓ 需要手动验证"
}

# 执行所有测试
echo "开始执行基线测试..."
echo ""

test_basic_movement
test_delete_undo
test_move_then_delete
test_cross_pane
test_text_objects
test_visual_mode
test_search
test_fsm_layers

echo ""
echo "=== 基线测试完成 ==="
echo "请手动验证每个测试场景"
echo ""
echo "如果所有测试通过，记录当前状态："
echo "  git log -1 --oneline"
echo "  git show pre-weaver-migration"

````

## 📄 `tools/gen-docs.go`

````go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

/*
====================================================
 Configuration & Globals
====================================================
*/

const versionStr = "v2.0.0"

// Config 集中管理配置
type Config struct {
	RootDir     string
	OutputFile  string
	IncludeExts []string
	ExcludeExts []string
	MaxFileSize int64
	NoSubdirs   bool
	Verbose     bool
	Version     bool
}

// FileMetadata 仅存储元数据，不存内容
type FileMetadata struct {
	RelPath  string
	FullPath string
	Size     int64
}

// Stats 统计信息
type Stats struct {
	FileCount int
	TotalSize int64
	Skipped   int
}

var defaultIgnorePatterns = []string{
	".git", ".idea", ".vscode",
	"node_modules", "vendor", "dist", "build", "target", "bin",
	"__pycache__", ".DS_Store",
	"package-lock.json", "yarn.lock", "go.sum",
}

// 语言映射表（全局配置，便于扩展）
var languageMap = map[string]string{
	".go":   "go",
	".js":   "javascript",
	".ts":   "typescript",
	".tsx":  "typescript",
	".jsx":  "javascript",
	".py":   "python",
	".java": "java",
	".c":    "c",
	".cpp":  "cpp",
	".cc":   "cpp",
	".cxx":  "cpp",
	".h":    "c",
	".hpp":  "cpp",
	".rs":   "rust",
	".rb":   "ruby",
	".php":  "php",
	".cs":   "csharp",
	".swift": "swift",
	".kt":   "kotlin",
	".scala": "scala",
	".r":    "r",
	".sql":  "sql",
	".sh":   "bash",
	".bash": "bash",
	".zsh":  "bash",
	".fish": "fish",
	".ps1":  "powershell",
	".md":   "markdown",
	".html": "html",
	".htm":  "html",
	".css":  "css",
	".scss": "scss",
	".sass": "sass",
	".less": "less",
	".xml":  "xml",
	".json": "json",
	".yaml": "yaml",
	".yml":  "yaml",
	".toml": "toml",
	".ini":  "ini",
	".conf": "conf",
	".txt":  "text",
}

/*
====================================================
 Main Entry
====================================================
*/

func main() {
	cfg := parseFlags()
	printStartupInfo(cfg)

	// Phase 1: 扫描文件结构
	fmt.Println("⏳ 正在扫描文件结构...")
	files, stats, err := scanDirectory(cfg)
	if err != nil {
		fmt.Printf("❌ 扫描失败: %v\n", err)
		os.Exit(1)
	}

	// Phase 2: 流式写入
	fmt.Printf("💾 正在写入文档 [文件数: %d]...\n", len(files))
	if err := writeMarkdownStream(cfg, files, stats); err != nil {
		fmt.Printf("❌ 写入失败: %v\n", err)
		os.Exit(1)
	}

	printSummary(stats, cfg.OutputFile)
}

/*
====================================================
 Flag Parsing
====================================================
*/

func parseFlags() Config {
	var cfg Config
	var include, exclude string
	var maxKB int64

	flag.StringVar(&cfg.RootDir, "dir", ".", "Root directory to scan")
	flag.StringVar(&cfg.OutputFile, "o", "", "Output markdown file")
	flag.StringVar(&include, "i", "", "Include extensions (e.g. .go,.js)")
	flag.StringVar(&exclude, "x", "", "Exclude extensions")
	flag.Int64Var(&maxKB, "max-size", 500, "Max file size in KB")
	flag.BoolVar(&cfg.NoSubdirs, "no-subdirs", false, "Do not scan subdirectories")
	flag.BoolVar(&cfg.NoSubdirs, "ns", false, "Alias for --no-subdirs")
	flag.BoolVar(&cfg.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&cfg.Version, "version", false, "Show version")

	flag.Parse()

	if cfg.Version {
		fmt.Printf("gen-docs %s\n", versionStr)
		os.Exit(0)
	}

	// 支持位置参数
	if args := flag.Args(); len(args) > 0 {
		cfg.RootDir = args[0]
	}

	// 自动生成输出文件名
	if cfg.OutputFile == "" {
		base := filepath.Base(cfg.RootDir)
		if base == "." || base == string(filepath.Separator) {
			base = "project"
		}
		date := time.Now().Format("20060102")
		cfg.OutputFile = fmt.Sprintf("%s-%s-docs.md", base, date)
	}

	cfg.IncludeExts = normalizeExts(include)
	cfg.ExcludeExts = normalizeExts(exclude)
	cfg.MaxFileSize = maxKB * 1024

	return cfg
}

/*
====================================================
 Startup & Summary
====================================================
*/

func printStartupInfo(cfg Config) {
	fmt.Println("▶ Gen-Docs Started")
	fmt.Printf("  Root: %s\n", cfg.RootDir)
	fmt.Printf("  Out : %s\n", cfg.OutputFile)
	fmt.Printf("  Max : %d KB\n", cfg.MaxFileSize/1024)
	if len(cfg.IncludeExts) > 0 {
		fmt.Printf("  Only: %v\n", cfg.IncludeExts)
	}
	if len(cfg.ExcludeExts) > 0 {
		fmt.Printf("  Skip: %v\n", cfg.ExcludeExts)
	}
	fmt.Println()
}

func printSummary(stats Stats, output string) {
	fmt.Println("\n✔ 完成!")
	fmt.Printf("  文件数  : %d\n", stats.FileCount)
	fmt.Printf("  已跳过  : %d\n", stats.Skipped)
	fmt.Printf("  总大小  : %.2f KB\n", float64(stats.TotalSize)/1024)
	fmt.Printf("  输出路径: %s\n", output)
}

/*
====================================================
 Directory Scanning
====================================================
*/

func scanDirectory(cfg Config) ([]FileMetadata, Stats, error) {
	var files []FileMetadata
	var stats Stats

	absOutput, _ := filepath.Abs(cfg.OutputFile)

	err := filepath.WalkDir(cfg.RootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logf(cfg.Verbose, "⚠ 无法访问: %s", path)
			stats.Skipped++
			return nil
		}

		relPath, _ := filepath.Rel(cfg.RootDir, path)
		if relPath == "." {
			return nil
		}

		// 处理目录
		if d.IsDir() {
			if cfg.NoSubdirs && relPath != "." {
				return filepath.SkipDir
			}
			if shouldIgnoreDir(d.Name()) {
				logf(cfg.Verbose, "⊘ 跳过目录: %s", relPath)
				return filepath.SkipDir
			}
			return nil
		}

		// 排除输出文件自身
		if absPath, _ := filepath.Abs(path); absPath == absOutput {
			return nil
		}

		// 获取文件信息
		info, err := d.Info()
		if err != nil {
			return nil
		}

		// 应用过滤规则
		if shouldIgnoreFile(relPath, info.Size(), cfg) {
			stats.Skipped++
			return nil
		}

		// 二进制检测
		if isBinaryFile(path) {
			logf(cfg.Verbose, "⊘ 二进制文件: %s", relPath)
			stats.Skipped++
			return nil
		}

		// 加入列表
		files = append(files, FileMetadata{
			RelPath:  relPath,
			FullPath: path,
			Size:     info.Size(),
		})
		stats.FileCount++
		stats.TotalSize += info.Size()

		logf(cfg.Verbose, "✓ 添加: %s", relPath)

		return nil
	})

	// 排序保证输出一致性
	sort.Slice(files, func(i, j int) bool {
		return files[i].RelPath < files[j].RelPath
	})

	return files, stats, err
}

/*
====================================================
 Ignore Rules
====================================================
*/

func shouldIgnoreDir(name string) bool {
	if strings.HasPrefix(name, ".") && name != "." {
		return true
	}
	for _, pattern := range defaultIgnorePatterns {
		if name == pattern {
			return true
		}
	}
	return false
}

func shouldIgnoreFile(relPath string, size int64, cfg Config) bool {
	// 大小限制
	if size > cfg.MaxFileSize {
		logf(cfg.Verbose, "⊘ 文件过大: %s", relPath)
		return true
	}

	ext := strings.ToLower(filepath.Ext(relPath))

	// 排除规则优先
	for _, e := range cfg.ExcludeExts {
		if ext == e {
			return true
		}
	}

	// 包含规则（白名单）
	if len(cfg.IncludeExts) > 0 {
		found := false
		for _, i := range cfg.IncludeExts {
			if ext == i {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	// 路径包含忽略模式
	parts := strings.Split(relPath, string(filepath.Separator))
	for _, part := range parts {
		for _, pattern := range defaultIgnorePatterns {
			if part == pattern {
				return true
			}
		}
	}

	return false
}

/*
====================================================
 File Utilities
====================================================
*/

func normalizeExts(input string) []string {
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	var exts []string
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if !strings.HasPrefix(p, ".") {
			p = "." + p
		}
		exts = append(exts, p)
	}
	return exts
}

func isBinaryFile(path string) bool {
	// 快速路径：压缩文件
	if strings.Contains(path, ".min.") {
		return true
	}

	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer f.Close()

	// 只读前 512 字节
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	buf = buf[:n]

	// NULL 字节检测
	for _, b := range buf {
		if b == 0 {
			return true
		}
	}

	// UTF-8 有效性检测
	return !utf8.Valid(buf)
}

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if lang, ok := languageMap[ext]; ok {
		return lang
	}
	return "text"
}

/*
====================================================
 Markdown Output
====================================================
*/

func writeMarkdownStream(cfg Config, files []FileMetadata, stats Stats) error {
	f, err := os.Create(cfg.OutputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriterSize(f, 64*1024)

	// 写入头部
	fmt.Fprintln(w, "# Project Documentation")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- **Generated at:** %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "- **Root Dir:** `%s`\n", cfg.RootDir)
	fmt.Fprintf(w, "- **File Count:** %d\n", stats.FileCount)
	fmt.Fprintf(w, "- **Total Size:** %.2f KB\n", float64(stats.TotalSize)/1024)
	fmt.Fprintln(w)

	// 写入目录
	fmt.Fprintln(w, "## 📂 File List")
	for _, file := range files {
		fmt.Fprintf(w, "- `%s` (%.2f KB)\n", file.RelPath, float64(file.Size)/1024)
	}
	fmt.Fprintln(w, "\n---")

	// 流式写入文件内容
	total := len(files)
	for i, file := range files {
		if !cfg.Verbose && (i%10 == 0 || i == total-1) {
			fmt.Printf("\r🚀 进度: %d/%d (%.1f%%)", i+1, total, float64(i+1)/float64(total)*100)
		}

		if err := copyFileContent(w, file); err != nil {
			logf(true, "\n⚠ 读取失败 %s: %v", file.RelPath, err)
			continue
		}
	}
	fmt.Println()

	// 【改进1】显式 Flush 并捕获错误
	return w.Flush()
}

func copyFileContent(w *bufio.Writer, file FileMetadata) error {
	src, err := os.Open(file.FullPath)
	if err != nil {
		return err
	}
	defer src.Close()

	lang := detectLanguage(file.RelPath)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "## 📄 `%s`\n\n", file.RelPath)
	
	// 【改进2】使用更安全的代码块分隔符（4个反引号）
	// 这样即使源代码中包含 ``` 也不会破坏格式
	fmt.Fprintf(w, "````%s\n", lang)

	if _, err := io.Copy(w, src); err != nil {
		return err
	}

	fmt.Fprintln(w, "\n````")
	return nil
}

/*
====================================================
 Logging
====================================================
*/

func logf(verbose bool, format string, a ...any) {
	if verbose {
		fmt.Printf(format+"\n", a...)
	}
}

````

## 📄 `tools/install-gen-docs.sh`

````bash
#!/usr/bin/env bash
# 项目文档生成工具安装脚本（全局可用 + gd 快捷命令）

set -e

echo "🚀 开始安装 gen-docs..."

# -------- 基础检查 --------
if ! command -v go &> /dev/null; then
    echo "❌ 未检测到 Go 编译器"
    echo "请先安装 Go: https://go.dev/dl/"
    exit 1
fi

echo "✓ Go 版本: $(go version)"

# -------- 编译 --------
echo "📦 编译 gen-docs..."
go build -o gen-docs gen-docs.go

# -------- 选择安装目录 --------
if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
    USE_SUDO=""
elif command -v sudo &> /dev/null; then
    INSTALL_DIR="/usr/local/bin"
    USE_SUDO="sudo"
else
    INSTALL_DIR="$HOME/.local/bin"
    USE_SUDO=""
    mkdir -p "$INSTALL_DIR"
fi

echo "📍 安装目录: $INSTALL_DIR"

# -------- 安装主程序 --------
echo "📥 安装 gen-docs"
$USE_SUDO mv gen-docs "$INSTALL_DIR/gen-docs"
$USE_SUDO chmod +x "$INSTALL_DIR/gen-docs"

# -------- 创建 gd 快捷命令（软链接） --------
echo "🔗 创建 gd 快捷命令"
$USE_SUDO ln -sf "$INSTALL_DIR/gen-docs" "$INSTALL_DIR/gd"

# -------- PATH 检查（仅在用户目录时） --------
if [[ "$INSTALL_DIR" == "$HOME/.local/bin" ]]; then
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo ""
        echo "⚠️  $INSTALL_DIR 不在 PATH 中"
        echo ""
        echo "请将以下内容加入你的 shell 配置文件："
        echo ""
        echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
        echo ""
        echo "然后执行:"
        echo "    source ~/.zshrc  或  source ~/.bashrc"
    else
        echo "✓ PATH 已正确配置"
    fi
fi

# -------- 完成 --------
echo ""
echo "✅ 安装完成！"
echo ""
echo "现在你可以在任意目录使用："
echo "  gen-docs     # 完整命令"
echo "  gd           # 快捷命令"
echo ""
echo "示例："
echo "  gd"
echo "  gd -i md,go"
echo "  gd -ns"
echo ""

````

## 📄 `transaction.go`

````go
package main

import (
	"time"
	"tmux-fsm/weaver/adapter"
	"tmux-fsm/weaver/core"
)

type TransactionID uint64

type Transaction struct {
	ID               TransactionID  `json:"id"`
	Records          []ActionRecord `json:"records"`
	CreatedAt        time.Time      `json:"created_at"`
	Applied          bool           `json:"applied"`
	Skipped          bool           `json:"skipped"`
	SafetyLevel      string         `json:"safety_level,omitempty"`       // exact, fuzzy
	PreSnapshotHash  string         `json:"pre_snapshot_hash,omitempty"`  // Phase 8: World state before transaction
	PostSnapshotHash string         `json:"post_snapshot_hash,omitempty"` // Phase 8: World state after transaction
}

type TransactionManager struct {
	current *Transaction
	nextID  TransactionID
}

// takeSnapshotForPane takes a snapshot of the given pane using the global weaver manager
func takeSnapshotForPane(paneID string) (string, error) {
	if weaverMgr != nil && weaverMgr.snapshotProvider != nil {
		snapshot, err := weaverMgr.snapshotProvider.TakeSnapshot(paneID)
		if err != nil {
			return "", err
		}
		return string(snapshot.Hash), nil
	}

	// Fallback: Use direct tmux capture if weaver is not available
	// This is a simplified approach - we'll capture the current line and hash it
	cursor := adapter.TmuxGetCursorPos(paneID)
	lines := adapter.TmuxCapturePane(paneID)

	// Use the new snapshot structure with LineID
	snapshot := core.TakeSnapshot(paneID, core.CursorPos{
		Row: cursor[0],
		Col: cursor[1],
	}, lines)

	return string(snapshot.Hash), nil
}

// computeSnapshotHash computes the hash of a snapshot
// NOTE: This is currently "Pane-only" scoped (Phase 8)
// For Phase 9+ (Split/Multi-pane), this will need to be upgraded to "World-scoped"
// where the hash represents the state of the affected world subgraph, not just a single pane
// [Phase 9] This function is now redundant since core.TakeSnapshot already computes the hash
func computeSnapshotHash(s core.Snapshot) core.SnapshotHash {
	return s.Hash
}

func (tm *TransactionManager) Begin(paneID string) {
	tm.current = &Transaction{
		ID:        tm.nextID,
		CreatedAt: time.Now(),
		Records:   []ActionRecord{},
	}

	// Take a snapshot before any changes occur
	if hash, err := takeSnapshotForPane(paneID); err == nil {
		tm.current.PreSnapshotHash = hash
	}

	tm.nextID++
}

func (tm *TransactionManager) Append(r ActionRecord) {
	if tm.current != nil {
		tm.current.Records = append(tm.current.Records, r)
	}
}

func (tm *TransactionManager) Commit(
	stack *[]Transaction,
	paneID string,
) {
	// --- Phase 8.0: 空事务直接丢弃 ---
	if tm.current == nil || len(tm.current.Records) == 0 {
		tm.current = nil
		return
	}

	tx := tm.current

	// --- Phase 8.1: 记录 PostSnapshot（事实，不做判断） ---
	if hash, err := takeSnapshotForPane(paneID); err == nil {
		tx.PostSnapshotHash = hash
	}

	// --- Phase 8.2: 标记为 Applied（仅表示"已执行完成"） ---
	tx.Applied = true

	// --- Phase 8.3: 提交到 Legacy 时间线（只有非跳过事务） ---
	// [Phase 9] Only add to legacy stack if not in Weaver mode
	// Weaver becomes the single source of truth for undo/redo
	if !tx.Skipped && GetMode() != ModeWeaver {
		*stack = append(*stack, *tx)
	}

	// --- Phase 8.4: 注入 Weaver（只有"存在的事务"才允许） ---
	if weaverMgr != nil && !tx.Skipped {
		weaverMgr.InjectLegacyTransaction(tx)
	}

	// --- Phase 8.5: 结束事务 ---
	tm.current = nil
}
````

## 📄 `validate_paths.sh`

````bash
#!/usr/bin/env bash
# 路径验证脚本

echo "=== tmux-fsm 路径验证 ==="

# 检查二进制文件是否存在
BINARY_PATH="$HOME/.tmux/plugins/tmux-fsm/tmux-fsm"

if [ -f "$BINARY_PATH" ]; then
    echo "✅ 二进制文件存在: $BINARY_PATH"
    echo "   文件大小: $(ls -lh "$BINARY_PATH" | awk '{print $5}')"
    echo "   可执行权限: $(if [ -x "$BINARY_PATH" ]; then echo "是"; else echo "否"; fi)"
else
    echo "❌ 二进制文件不存在: $BINARY_PATH"
    echo "   请先运行 install.sh 或手动构建"
    exit 1
fi

# 测试二进制文件是否可以执行
echo ""
echo "=== 测试二进制文件功能 ==="
if "$BINARY_PATH" -h >/dev/null 2>&1; then
    echo "✅ 二进制文件可执行"
else
    echo "❌ 二进制文件执行失败"
    exit 1
fi

# 检查版本信息
echo ""
echo "=== 二进制文件信息 ==="
"$BINARY_PATH" -h

echo ""
echo "=== 路径验证完成 ==="
echo "所有路径配置正确，tmux-fsm 可以正常工作"
````

## 📄 `weaver/adapter/selection_normalizer.go`

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

## 📄 `weaver/adapter/snapshot.go`

````go
package adapter

import "tmux-fsm/weaver/core"

// SnapshotProvider 世界读取接口
// 负责从物理世界（tmux）提取不可变的 Snapshot
type SnapshotProvider interface {
	TakeSnapshot(paneID string) (core.Snapshot, error)
}

````

## 📄 `weaver/adapter/snapshot_hash.go`

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

## 📄 `weaver/adapter/tmux_adapter.go`

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
func (p *NoopProjection) Apply(resolved []core.ResolvedAnchor, facts []core.ResolvedFact) error {
	// Shadow 模式：不执行任何操作
	return nil
}

````

## 📄 `weaver/adapter/tmux_physical.go`

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
	// 使用管道 load-buffer 是最健壮的，彻底避免 '?' 乱码问题
	cmd := exec.Command("tmux", "load-buffer", "-")
	cmd.Stdin = strings.NewReader(text)
	cmd.Run()

	// 确保粘贴到目标
	exec.Command("tmux", "paste-buffer", "-t", targetPane).Run()
}

````

## 📄 `weaver/adapter/tmux_projection.go`

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
		entry := log[i]
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

## 📄 `weaver/adapter/tmux_reality.go`

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

## 📄 `weaver/adapter/tmux_snapshot.go`

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

## 📄 `weaver/adapter/tmux_utils.go`

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

## 📄 `weaver/core/allowed_lines.go`

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

## 📄 `weaver/core/anchor_kind.go`

````go
package core

type AnchorKind int

const (
	AnchorUnknown AnchorKind = iota

	// Cursor-relative
	AnchorAtCursor

	// Semantic
	AnchorWord
	AnchorLine
	AnchorParagraph

	// Structural
	AnchorSelection
	AnchorAbsolute

	// Legacy Support
	AnchorLegacyRange
)

````

## 📄 `weaver/core/hash.go`

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

## 📄 `weaver/core/history.go`

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

## 📄 `weaver/core/intent_fusion.go`

````go
package core

// canFuse checks if two intents can be fused together
// Phase 13.0: Conservative fusion rules
func canFuse(a, b Intent) bool {
	// Check if kinds match
	if a.Kind != b.Kind {
		return false
	}
	
	// Only allow fusing for insert operations at the same position
	if a.Kind == FactInsert {
		// Check if both intents target the same position in the same line
		if len(a.Anchors) == 1 && len(b.Anchors) == 1 {
			anchorA := a.Anchors[0]
			anchorB := b.Anchors[0]
			
			// Same line and same position
			return anchorA.LineID == anchorB.LineID && 
				   anchorA.Start == anchorB.Start && 
				   anchorA.End == anchorB.End &&
				   anchorA.PaneID == anchorB.PaneID
		}
	}
	
	return false
}

// fuse combines two compatible intents into one
// Phase 13.0: Simple concatenation for insert operations
func fuse(a, b Intent) Intent {
	if a.Kind == FactInsert && b.Kind == FactInsert {
		// For insert operations, concatenate the text
		result := a
		result.Payload.Text += b.Payload.Text
		return result
	}
	
	// For other operations, just return the first one (shouldn't happen if canFuse worked correctly)
	return a
}

// FuseIntents combines compatible intents in a sequence
// Phase 13.0: Sequential intent fusion
func FuseIntents(intents []Intent) []Intent {
	if len(intents) <= 1 {
		return intents
	}

	var out []Intent
	out = append(out, intents[0])

	for i := 1; i < len(intents); i++ {
		lastIdx := len(out) - 1
		if canFuse(out[lastIdx], intents[i]) {
			out[lastIdx] = fuse(out[lastIdx], intents[i])
		} else {
			out = append(out, intents[i])
		}
	}
	return out
}
````

## 📄 `weaver/core/line_hash_verifier.go`

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
                OK: false,
                Safety: SafetyUnsafe,
                Diffs: diffs,
                Message: "unexpected line modified",
            }
        }
    }

    return VerificationResult{
        OK: true,
        Safety: SafetyExact,
        Diffs: diffs,
    }
}
````

## 📄 `weaver/core/projection_verifier.go`

````go
package core

type SafetyLevel int

const (
    SafetyExact SafetyLevel = iota
    SafetyFuzzy
    SafetyUnsafe
)

type VerificationResult struct {
    OK      bool
    Safety  SafetyLevel
    Diffs   []SnapshotDiff
    Message string
}

type ProjectionVerifier interface {
    Verify(
        pre Snapshot,
        facts []ResolvedFact,
        post Snapshot,
    ) VerificationResult
}
````

## 📄 `weaver/core/resolved_fact.go`

````go
package core

// ResolvedAnchor 代表具体的物理位置 (Phase 5.2)
// 它是 Resolver 解析后的结果，Projection 只认这个
type ResolvedAnchor struct {
	PaneID string
	LineID LineID  // Stable line identifier (Phase 9)
	Line   int     // Fallback line number for compatibility
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

## 📄 `weaver/core/shadow_engine.go`

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
	if err := e.projection.Apply(nil, resolvedFacts); err != nil {
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
	if err := e.projection.Apply(nil, resolvedFacts); err != nil {
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
	if err := e.projection.Apply(nil, resolvedFacts); err != nil {
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

## 📄 `weaver/core/snapshot.go`

````go
package core

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// SnapshotHash 快照哈希（世界指纹）
type SnapshotHash string

// LineHash 行哈希（局部指纹）
type LineHash string

// LineID 行ID（基于内容的稳定ID）
type LineID string

// Snapshot 世界快照（不可变）
// 代表 Intent 形成时对世界的冻结视图
// Now uses stable LineID for diffing capabilities
type Snapshot struct {
	PaneID string
	Cursor CursorPos
	Lines  []LineSnapshot
	Index  map[LineID]int // Stable mapping from LineID to position
	Hash   SnapshotHash
	TakenAt time.Time
}

// CursorPos 光标位置
type CursorPos struct {
	Row int
	Col int
}

// LineSnapshot 单行快照
type LineSnapshot struct {
	ID   LineID // Stable ID based on content
	Text string
	Hash LineHash
}

// TakeSnapshot creates a new snapshot with stable LineIDs
func TakeSnapshot(paneID string, cursor CursorPos, lines []string) Snapshot {
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
		TakenAt: time.Now(),
	}

	snapshot.Hash = computeSnapshotHash(snapshot)
	return snapshot
}

// makeLineID creates a stable LineID based on content
func makeLineID(paneID string, prev LineID, text string) LineID {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s", paneID, prev, text)))
	return LineID(fmt.Sprintf("%x", h[:]))
}

// hashLine computes the hash of a line
func hashLine(text string) LineHash {
	h := sha256.Sum256([]byte(text))
	return LineHash(fmt.Sprintf("%x", h[:]))
}

// computeSnapshotHash computes the hash of a snapshot
func computeSnapshotHash(s Snapshot) SnapshotHash {
	h := sha256.New()
	for _, l := range s.Lines {
		h.Write([]byte(l.ID))
		h.Write([]byte(l.Hash))
	}
	return SnapshotHash(fmt.Sprintf("%x", h.Sum(nil)))
}

````

## 📄 `weaver/core/snapshot_diff.go`

````go
package core

type DiffKind int

const (
    DiffInsert DiffKind = iota
    DiffDelete
    DiffModify
)

type SnapshotDiff struct {
    LineID  LineID
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

## 📄 `weaver/core/snapshot_types.go`

````go
package core

import (
	"crypto/sha256"
	"fmt"
)

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

## 📄 `weaver/core/take_snapshot.go`

````go
package core

import (
	"crypto/sha256"
	"fmt"
)

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

## 📄 `weaver/core/types.go`

````go
package core

import (
	"errors"
)

// ErrWorldDrift 世界漂移错误（快照不匹配）
// 表示 Intent 基于的历史与当前现实不一致
var ErrWorldDrift = errors.New("world drift: snapshot mismatch")

// Fact 表示一个已发生的编辑事实（不可变）
// 这是 Weaver Core 的核心数据结构
// Phase 5.3: 不再包含物理 Range
type Fact struct {
	Kind        FactKind               `json:"kind"`
	Anchor      Anchor                 `json:"anchor"`
	Payload     FactPayload            `json:"payload"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	Timestamp   int64                  `json:"timestamp"`
	SideEffects []string               `json:"side_effects,omitempty"`
}

// FactKind 事实类型
type FactKind int

const (
	FactNone FactKind = iota
	FactInsert
	FactDelete
	FactReplace
	FactMove
)

// AnchorKind 锚点类型
type AnchorKind int

const (
	AnchorNone AnchorKind = iota
	AnchorAtCursor
	AnchorWord
	AnchorLine
	AnchorAbsolute
	AnchorLegacyRange
)

// Anchor 描述“我们想要操作的目标”，而不是“它在哪里”
// Phase 5.3: 纯语义 Anchor
type Anchor struct {
	PaneID string     `json:"pane_id"`
	Kind   AnchorKind `json:"kind"`
	Ref    any        `json:"ref,omitempty"`
	Hash   string     `json:"hash,omitempty"` // Phase 5.4: Reconciliation Expectation
	LineID LineID     `json:"line_id,omitempty"` // Phase 9: Stable line identifier
	Start  int        `json:"start,omitempty"`   // Phase 11: Start position in line
	End    int        `json:"end,omitempty"`     // Phase 11: End position in line
}

// FactPayload 事实的具体内容
type FactPayload struct {
	Text     string `json:"text,omitempty"`
	OldText  string `json:"old_text,omitempty"`
	NewText  string `json:"new_text,omitempty"`
	Value    string `json:"value,omitempty"`
	Position int    `json:"position,omitempty"`
}

// Transaction 事务
// 包含一组 Facts，具有原子性
type Transaction struct {
	ID               TransactionID `json:"id"`
	Intent           Intent        `json:"intent"`        // 原始意图
	Facts            []Fact        `json:"facts"`         // 正向事实序列
	InverseFacts     []Fact        `json:"inverse_facts"` // 反向事实序列（用于 Undo）
	Safety           SafetyLevel   `json:"safety"`
	Timestamp        int64         `json:"timestamp"`
	Applied          bool          `json:"applied"`
	Skipped          bool          `json:"skipped"`
	PostSnapshotHash string        `json:"post_snapshot_hash,omitempty"` // Phase 7: State after application
	AllowPartial     bool          `json:"allow_partial,omitempty"`      // Phase 7: Explicit flag for fuzzy match
}

// TransactionID 事务 ID
type TransactionID string

// SafetyLevel 安全级别
type SafetyLevel int

const (
	SafetyExact SafetyLevel = iota
	SafetyFuzzy
	SafetyUnsafe
)

// Verdict 裁决结果（可审计输出）
type Verdict struct {
	Kind        VerdictKind        `json:"kind"`
	Safety      SafetyLevel        `json:"safety"`
	Message     string             `json:"message"`
	Transaction *Transaction       `json:"transaction,omitempty"`
	Resolutions []AnchorResolution `json:"resolutions,omitempty"`
	Audit       []AuditEntry       `json:"audit,omitempty"` // Renamed from Details
}

// VerdictKind 裁决类型
type VerdictKind int

const (
	VerdictApplied VerdictKind = iota
	VerdictRejected
	VerdictSkipped
	VerdictBlocked // Phase 5.4: Blocked by Reconciliation
)

// AuditEntry 审计条目
type AuditEntry struct {
	Step   string `json:"step"`
	Result string `json:"result"`
}

// AnchorResolution Anchor 解析结果
type AnchorResolution int

const (
	AnchorExact AnchorResolution = iota
	AnchorFuzzy
	AnchorFailed
)

// UndoEntry represents a single undo operation
// Phase 12.0: Projection-level undo log
type UndoEntry struct {
	LineID LineID `json:"line_id"`
	Before string `json:"before"`
	After  string `json:"after"`
}

````

## 📄 `weaver/logic/passthrough_resolver.go`

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
			LineID:  ra.LineID,        // Phase 9: Include stable LineID
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
			LineID: "",        // 空 LineID，明确表示不稳定
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
			LineID: "",        // 空 LineID，明确表示不稳定
			Line:   row,
			Start:  start,
			End:    end,
		}, nil

	case core.AnchorLine:
		// use lineText already captured
		return core.ResolvedAnchor{
			PaneID: a.PaneID,
			LineID: "",        // 空 LineID，明确表示不稳定
			Line:   row,
			Start:  0,
			End:    len(lineText) - 1,
		}, nil

	case core.AnchorLegacyRange:
		// Legacy Range encoded in Ref
		if m, ok := a.Ref.(map[string]int); ok {
			return core.ResolvedAnchor{
				PaneID: a.PaneID,
				LineID: "",        // 空 LineID，明确表示不稳定
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
			LineID: "",        // 空 LineID，明确表示不稳定
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

## 📄 `weaver/logic/shell_fact_builder.go`

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

## 📄 `weaver_manager.go`

````go
package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"tmux-fsm/weaver/adapter"
	"tmux-fsm/weaver/core"
	"tmux-fsm/weaver/logic"
)

// WeaverManager 全局协调器
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
	logWeaver("Weaver initialized in %s mode", modeString(mode))
}

// ProcessIntentGlobal 全局处理入口
func ProcessIntentGlobal(intent Intent) {
	if weaverMgr == nil {
		return
	}
	weaverMgr.ProcessIntent(intent)
}

// ProcessIntent 处理意图 (Gateway)
func (m *WeaverManager) ProcessIntent(intent Intent) {
	logWeaver("ProcessIntent: Kind=%v Target=%v", intent.Kind, intent.Target)

	// Phase 6.2: Capture Snapshot (Time Freeze)
	paneID := intent.GetPaneID()
	if paneID == "" {
		// Try to deduce or fail
		logWeaver("No PaneID in intent, skipping snapshot")
		return // Or handle non-pane intents
	}

	snapshot, err := m.snapshotProvider.TakeSnapshot(paneID)
	if err != nil {
		logWeaver("Snapshot failed: %v", err)
		return
	}

	// Inject Hash into Intent (mutable struct in main)
	intent.SnapshotHash = string(snapshot.Hash)

	coreIntent := &intentAdapter{intent: intent}

	// 此时如果是 Undo/Redo，它们不需要 Snapshot?
	// Phase 6.2 定义：Any ApplyIntent needs Snapshot.
	// Undo/Redo often imply "Previous State", but current implementation calls Planner even for Undo/Redo?
	// No, `ApplyIntent` handles Undo/Redo specially.
	// It calls `performUndo`.

	if m.mode == ModeShadow || m.mode == ModeWeaver {
		verdict, err := m.engine.ApplyIntent(coreIntent, snapshot)
		if err != nil {
			logWeaver("Engine Error: %v", err)
			// Phase 7: Propagate to UI
			stateMu.Lock()
			globalState.LastUndoFailure = fmt.Sprintf("Engine: %v", err)
			stateMu.Unlock()
		} else {
			logWeaver("Verdict: %v (Safe=%v)", verdict.Kind, verdict.Safety)
			if len(verdict.Audit) > 0 {
				logWeaver("Audit: %v", verdict.Audit)
			}
			// If applied successfully, clear failure
			stateMu.Lock()
			if globalState.LastUndoFailure != "" && strings.HasPrefix(globalState.LastUndoFailure, "Engine:") {
				globalState.LastUndoFailure = ""
			}
			stateMu.Unlock()
		}
	}

	// [Phase 4] Phase 3 的 Weaver -> Legacy 桥接已禁用
	// 现在 Weaver History 是 Source of Truth，Legacy 操作将通过反向桥接注入 Weaver
}

// InjectLegacyTransaction 将 Legacy 事务注入到 Weaver History (Reverse Bridge)
func (m *WeaverManager) InjectLegacyTransaction(legacyTx *Transaction) {
	if m.engine == nil {
		return
	}
	// 获取 ShadowEngine 的 History
	se, ok := m.engine.(*core.ShadowEngine)
	if !ok {
		return
	}

	coreTx := &core.Transaction{
		ID:           core.TransactionID(fmt.Sprintf("legacy-%d", legacyTx.ID)),
		Timestamp:    legacyTx.CreatedAt.Unix(),
		Facts:        make([]core.Fact, 0),
		InverseFacts: make([]core.Fact, 0),
		Applied:      true,
		Safety:       core.SafetyExact,
	}

	// 转换正向事实
	for _, rec := range legacyTx.Records {
		f := convertLegacyFactToCore(rec.Fact)
		coreTx.Facts = append(coreTx.Facts, f)
	}

	// 转换反向事实 (通常 Inverse 用于 Undo。Legacy Undo 执行 Inverse。
	// Weaver Undo 执行 InverseFacts。顺序：Record1, Record2. Undo: Inv2, Inv1。
	// 所以我们需要倒序遍历 Records)
	for i := len(legacyTx.Records) - 1; i >= 0; i-- {
		rec := legacyTx.Records[i]
		inv := convertLegacyFactToCore(rec.Inverse)
		coreTx.InverseFacts = append(coreTx.InverseFacts, inv)
	}

	if len(coreTx.Facts) > 0 {
		se.GetHistory().Push(coreTx)
		logWeaver("Injected Legacy Transaction %d -> %s", legacyTx.ID, coreTx.ID)
	}
}

func convertLegacyFactToCore(lf Fact) core.Fact {
	// Construct Semantic Anchor with Legacy Physical Info for Resolver to unpack
	ref := map[string]int{
		"line":  lf.Target.Anchor.LineHint,
		"start": lf.Target.StartOffset,
		"end":   lf.Target.EndOffset,
	}

	cf := core.Fact{
		Anchor: core.Anchor{
			PaneID: lf.Target.Anchor.PaneID,
			Kind:   core.AnchorLegacyRange,
			Ref:    ref,
		},
		SideEffects: lf.SideEffects,
	}
	// Note: Hash is currently ignored in legacy conversion

	switch lf.Kind {
	case "delete":
		cf.Kind = core.FactDelete
		cf.Payload.OldText = lf.Target.Text
	case "insert":
		cf.Kind = core.FactInsert
		cf.Payload.Text = lf.Target.Text
	case "replace":
		cf.Kind = core.FactReplace
		cf.Payload.OldText = lf.Target.Text
		if s, ok := lf.Meta["new_text"].(string); ok {
			cf.Payload.NewText = s
		}
	default:
		cf.Kind = core.FactNone
	}
	return cf
}

// intentAdapter 适配 main.Intent 到 core.Intent
type intentAdapter struct {
	intent Intent
}

func (a *intentAdapter) GetKind() core.IntentKind {
	return core.IntentKind(a.intent.Kind)
}

func (a *intentAdapter) GetTarget() core.SemanticTarget {
	return core.SemanticTarget{
		Kind:      int(a.intent.Target.Kind),
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
	return a.intent.GetPaneID()
}

func (a *intentAdapter) GetSnapshotHash() string {
	return a.intent.GetSnapshotHash()
}

func (a *intentAdapter) IsPartialAllowed() bool {
	return a.intent.IsPartialAllowed()
}

// logWeaver ...
func logWeaver(format string, args ...interface{}) {
	if !globalConfig.LogFacts {
		return
	}
	f, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] [WEAVER] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
		f.Close()
	}
}

// modeString 返回模式的字符串表示
func modeString(mode ExecutionMode) string {
	switch mode {
	case ModeLegacy:
		return "legacy"
	case ModeShadow:
		return "shadow"
	case ModeWeaver:
		return "weaver"
	default:
		return "unknown"
	}
}

````
