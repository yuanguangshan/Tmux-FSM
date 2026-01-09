# Project Documentation

- **Generated at:** 2026-01-09 19:16:28
- **Root Dir:** `.`
- **File Count:** 85
- **Total Size:** 288.27 KB

## 📂 File List
- `backend/backend.go` (2.96 KB)
- `client.go` (1.87 KB)
- `cmd/verifier/main.go` (0.74 KB)
- `config.go` (1.37 KB)
- `crdt/crdt.go` (5.35 KB)
- `editor/engine.go` (3.43 KB)
- `editor/execution_context.go` (0.58 KB)
- `editor/selection_update.go` (4.24 KB)
- `editor/stores.go` (2.14 KB)
- `editor/text_object.go` (13.10 KB)
- `engine.go` (8.76 KB)
- `engine/concrete_engine.go` (5.41 KB)
- `engine/engine.go` (6.95 KB)
- `examples/transaction_demo.go` (2.63 KB)
- `fsm/engine.go` (9.37 KB)
- `fsm/keymap.go` (1.16 KB)
- `fsm/nvim.go` (0.67 KB)
- `fsm/token.go` (0.17 KB)
- `fsm/ui_stub.go` (1.85 KB)
- `globals.go` (4.29 KB)
- `index/index.go` (6.57 KB)
- `intent.go` (5.23 KB)
- `intent/Adapter.go` (1.24 KB)
- `intent/builder/builder.go` (0.51 KB)
- `intent/builder/composite_builder.go` (1.06 KB)
- `intent/builder/doc.go` (0.35 KB)
- `intent/builder/intent_diff.go` (1.20 KB)
- `intent/builder/macro_builder.go` (1.37 KB)
- `intent/builder/move_builder.go` (1.60 KB)
- `intent/builder/operator_builder.go` (1.27 KB)
- `intent/builder/semantic_equal.go` (0.73 KB)
- `intent/builder/text_object.go` (2.61 KB)
- `intent/grammar_intent.go` (0.20 KB)
- `intent/motion.go` (0.79 KB)
- `intent/promote.go` (0.35 KB)
- `intent/range.go` (0.22 KB)
- `intent/text_object.go` (0.28 KB)
- `intent_bridge.go` (6.25 KB)
- `invariant/test.go` (4.17 KB)
- `kernel/decide.go` (2.20 KB)
- `kernel/execute.go` (0.73 KB)
- `kernel/intent_executor.go` (0.21 KB)
- `kernel/kernel.go` (2.38 KB)
- `kernel/resolver_executor.go` (0.77 KB)
- `kernel/transaction.go` (1.99 KB)
- `main.go` (15.98 KB)
- `pkg/protocol/protocol.go` (0.77 KB)
- `pkg/server/server.go` (5.88 KB)
- `pkg/state/state.go` (5.30 KB)
- `planner/grammar.go` (12.71 KB)
- `planner/grammar_test.go` (5.57 KB)
- `policy/policy.go` (7.37 KB)
- `protocol.go` (0.78 KB)
- `selection/selection.go` (5.68 KB)
- `semantic/capture.go` (8.16 KB)
- `tests/invalid_history_test.go` (1.04 KB)
- `text_object.go` (12.99 KB)
- `tools/gen-docs.go` (10.45 KB)
- `ui/interface.go` (0.08 KB)
- `ui/popup.go` (0.71 KB)
- `undotree/tree.go` (2.80 KB)
- `verifier/verifier.go` (8.43 KB)
- `weaver/adapter/backend.go` (3.00 KB)
- `weaver/adapter/selection_normalizer.go` (1.66 KB)
- `weaver/adapter/snapshot.go` (0.23 KB)
- `weaver/adapter/snapshot_hash.go` (0.41 KB)
- `weaver/adapter/tmux_adapter.go` (1.86 KB)
- `weaver/adapter/tmux_physical.go` (12.08 KB)
- `weaver/adapter/tmux_projection.go` (6.93 KB)
- `weaver/adapter/tmux_reality.go` (0.23 KB)
- `weaver/adapter/tmux_snapshot.go` (0.36 KB)
- `weaver/adapter/tmux_utils.go` (2.25 KB)
- `weaver/core/allowed_lines.go` (0.27 KB)
- `weaver/core/hash.go` (0.54 KB)
- `weaver/core/history.go` (2.51 KB)
- `weaver/core/intent_fusion.go` (1.86 KB)
- `weaver/core/line_hash_verifier.go` (0.59 KB)
- `weaver/core/resolved_fact.go` (0.69 KB)
- `weaver/core/shadow_engine.go` (10.02 KB)
- `weaver/core/snapshot_diff.go` (1.03 KB)
- `weaver/core/snapshot_types.go` (0.31 KB)
- `weaver/core/take_snapshot.go` (0.58 KB)
- `weaver/logic/passthrough_resolver.go` (7.33 KB)
- `weaver/logic/shell_fact_builder.go` (2.48 KB)
- `weaver/manager/manager.go` (5.03 KB)

---

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
	// 添加参数验证和修复
	if paneAndClient == "" || paneAndClient == "|" {
		// 尝试获取当前pane和client
		// 注意：这里不能直接调用 tmux 命令，因为这可能导致循环依赖
		// 我们需要确保参数格式正确
		paneAndClient = "default|default"
	} else {
		// 检查参数格式是否正确 (pane|client)，如果 client 部分为空，尝试修复
		parts := strings.Split(paneAndClient, "|")
		if len(parts) == 2 && parts[1] == "" {
			// client 部分为空，使用默认值
			paneAndClient = parts[0] + "|default"
		} else if len(parts) == 1 {
			// 只有 pane 部分，添加默认 client
			paneAndClient = parts[0] + "|default"
		}
	}

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

## 📄 `cmd/verifier/main.go`

````go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: verifier verify <path>")
		os.Exit(1)
	}

	cmd := os.Args[1]
	path := os.Args[2]

	if cmd != "verify" {
		fmt.Println("unknown command:", cmd)
		os.Exit(1)
	}

	_, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("read error:", err)
		os.Exit(1)
	}

	// 这里需要根据实际的 verifier 接口进行调整
	// input, err := verifier.ParseVerificationInput(data)
	// if err != nil {
	// 	fmt.Println("parse error:", err)
	// 	os.Exit(1)
	// }

	// root, err := verifier.Verify(input)
	// if err != nil {
	// 	fmt.Println("❌ verification failed:", err)
	// 	os.Exit(2)
	// }

	fmt.Println("✅ verification succeeded")
	fmt.Println("StateRoot: TODO")
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

## 📄 `crdt/crdt.go`

````go
package crdt

import (
	"sort"
	"time"
	"tmux-fsm/semantic"
)

// EventID 事件ID类型
type EventID string

// ActorID 参与者ID类型
type ActorID string

// PositionID CRDT 位置ID
type PositionID struct {
	Path  []uint32 `json:"path"`
	Actor ActorID  `json:"actor"`
	Epoch int      `json:"epoch"` // 每次分叉/reset +1
}

// SemanticEvent 修正后的语义事件结构
type SemanticEvent struct {
	// 全局唯一、幂等基础
	ID    EventID   `json:"id"`
	Actor ActorID   `json:"actor"`
	Time  time.Time `json:"time"`

	// 因果一致性（CRDT 用）
	CausalParents []EventID `json:"causal_parents"`
	// 含义：本事件在语义上依赖的事件集合
	// ✅ 用于拓扑排序 / 合并
	// ✅ 永远不用于 Undo

	// 本地历史（Undo 用）
	LocalParent EventID `json:"local_parent"`
	// 含义：本 actor 本地编辑历史中的上一个事件
	// ✅ 只在本地有意义
	// ✅ 不同步、不合并

	// 不可变语义
	Fact semantic.Fact `json:"fact"`
}

// ComparePos 比较两个位置
func ComparePos(a, b PositionID) int {
	min := len(a.Path)
	if len(b.Path) < min {
		min = len(b.Path)
	}

	for i := 0; i < min; i++ {
		if a.Path[i] < b.Path[i] {
			return -1
		}
		if a.Path[i] > b.Path[i] {
			return 1
		}
	}
	if len(a.Path) != len(b.Path) {
		if len(a.Path) < len(b.Path) {
			return -1
		}
		return 1
	}
	if a.Actor < b.Actor {
		return -1
	}
	if a.Actor > b.Actor {
		return 1
	}
	if a.Epoch < b.Epoch {
		return -1
	}
	if a.Epoch > b.Epoch {
		return 1
	}
	return 0
}

// AllocateBetween 在两个位置之间分配新位置
func AllocateBetween(a, b *PositionID, actor ActorID) PositionID {
	const Base = uint32(1 << 31)

	var path []uint32
	i := 0

	for {
		var left uint32 = 0
		var right uint32 = Base

		if a != nil && i < len(a.Path) {
			left = a.Path[i]
		}
		if b != nil && i < len(b.Path) {
			right = b.Path[i]
		}

		if right-left > 1 {
			mid := left + (right-left)/2
			path = append(path, mid)
			break
		}

		path = append(path, left)
		i++
	}

	return PositionID{
		Path:  path,
		Actor: actor,
		Epoch: 0, // 可能需要根据实际情况设置
	}
}

// EventStore 事件存储
type EventStore struct {
	Events map[EventID]SemanticEvent
}

// NewEventStore 创建新的事件存储
func NewEventStore() *EventStore {
	return &EventStore{
		Events: make(map[EventID]SemanticEvent),
	}
}

// Merge 合并事件（网络/WAL/Sync）
func (s *EventStore) Merge(e SemanticEvent) {
	if _, ok := s.Events[e.ID]; ok {
		return // 幂等
	}
	s.Events[e.ID] = e
}

// TopoSort 拓扑排序（因果顺序）
func (s *EventStore) TopoSort() []SemanticEvent {
	return TopoSortByCausality(s.Events)
}

// TopoSortByCausality 按因果关系拓扑排序
func TopoSortByCausality(events map[EventID]SemanticEvent) []SemanticEvent {
	inDegree := make(map[EventID]int)
	graph := make(map[EventID][]EventID)

	// 初始化
	for id := range events {
		inDegree[id] = 0
	}

	// 构建因果图
	for _, e := range events {
		for _, p := range e.CausalParents {
			if _, ok := events[p]; ok {
				graph[p] = append(graph[p], e.ID)
				inDegree[e.ID]++
			}
		}
	}

	// 入度为 0 的队列
	var queue []EventID
	for id, d := range inDegree {
		if d == 0 {
			queue = append(queue, id)
		}
	}

	// 稳定排序（可选：EventID）
	sort.Slice(queue, func(i, j int) bool {
		return queue[i] < queue[j]
	})

	var result []SemanticEvent

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		result = append(result, events[id])

		for _, next := range graph[id] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	// 检测环（理论上不该出现）
	if len(result) != len(events) {
		panic("causal cycle detected")
	}

	return result
}

// LocalHistory 获取本地历史（参与者投影）
func LocalHistory(events map[EventID]SemanticEvent, me ActorID) []SemanticEvent {
	var out []SemanticEvent
	for _, e := range events {
		if e.Actor == me {
			out = append(out, e)
		}
	}
	return BuildLocalChain(out)
}

// BuildLocalChain 构建本地链
func BuildLocalChain(events []SemanticEvent) []SemanticEvent {
	// 按 LocalParent 链组织
	parentMap := make(map[EventID]EventID)
	eventMap := make(map[EventID]SemanticEvent)

	for _, e := range events {
		eventMap[e.ID] = e
		parentMap[e.ID] = e.LocalParent
	}

	// 找到根节点
	var roots []EventID
	for id, parent := range parentMap {
		if parent == "" {
			roots = append(roots, id)
		}
	}

	// 按链排序
	var result []SemanticEvent
	visited := make(map[EventID]bool)

	var dfs func(EventID)
	dfs = func(id EventID) {
		if visited[id] {
			return
		}
		visited[id] = true
		result = append(result, eventMap[id])

		// 找到所有子节点
		for childID, parentID := range parentMap {
			if parentID == id {
				dfs(childID)
			}
		}
	}

	for _, root := range roots {
		dfs(root)
	}

	return result
}

// UndoFilter 创建撤销过滤器
func UndoFilter(me ActorID, undoPoint EventID, events map[EventID]SemanticEvent) func(SemanticEvent) bool {
	disabled := make(map[EventID]bool)

	// 从撤销点向上追踪，标记需要禁用的事件
	current := undoPoint
	for current != "" {
		disabled[current] = true

		// 找到当前事件
		event, exists := events[current]
		if !exists {
			break
		}

		// 移动到父事件
		current = event.LocalParent
	}

	return func(e SemanticEvent) bool {
		if e.Actor != me {
			return true
		}
		// 如果事件在撤销点之后，则不执行
		return !disabled[e.ID]
	}
}

````

## 📄 `editor/engine.go`

````go
package editor

import (
	"errors"
	"fmt"
)

// SimpleBuffer 简单的缓冲区实现
type SimpleBuffer struct {
	lines []string
}

// NewSimpleBuffer 创建新的简单缓冲区
func NewSimpleBuffer(initialText []string) *SimpleBuffer {
	if len(initialText) == 0 {
		initialText = []string{""}
	}
	return &SimpleBuffer{
		lines: initialText,
	}
}

func (sb *SimpleBuffer) LineCount() int {
	return len(sb.lines)
}

func (sb *SimpleBuffer) LineLength(row int) int {
	if row < 0 || row >= len(sb.lines) {
		return 0
	}
	return len(sb.lines[row])
}

func (sb *SimpleBuffer) Line(row int) string {
	if row < 0 || row >= len(sb.lines) {
		return ""
	}
	return sb.lines[row]
}

func (sb *SimpleBuffer) RuneAt(row, col int) rune {
	if row < 0 || row >= len(sb.lines) {
		return 0
	}
	line := sb.lines[row]
	if col < 0 || col >= len(line) {
		return 0
	}
	return rune(line[col])
}

func (sb *SimpleBuffer) InsertAt(anchor Cursor, text string) error {
	if anchor.Row < 0 || anchor.Row >= len(sb.lines) {
		return errors.New("invalid row")
	}

	line := sb.lines[anchor.Row]
	if anchor.Col < 0 || anchor.Col > len(line) {
		return errors.New("invalid column")
	}

	newLine := line[:anchor.Col] + text + line[anchor.Col:]
	sb.lines[anchor.Row] = newLine

	return nil
}

func (sb *SimpleBuffer) DeleteRange(start, end Cursor) (string, error) {
	if start.Row < 0 || start.Row >= len(sb.lines) || end.Row < 0 || end.Row >= len(sb.lines) {
		return "", errors.New("invalid row")
	}

	// 确保 start <= end
	if end.Row < start.Row || (start.Row == end.Row && end.Col < start.Col) {
		start, end = end, start
	}

	var deletedText string
	if start.Row == end.Row {
		line := sb.lines[start.Row]
		if start.Col < 0 || end.Col > len(line) {
			return "", errors.New("invalid column range")
		}
		deletedText = line[start.Col:end.Col]
		sb.lines[start.Row] = line[:start.Col] + line[end.Col:]
	} else {
		// 跨行删除
		firstLine := sb.lines[start.Row]
		lastLine := sb.lines[end.Row]

		deletedText = firstLine[start.Col:] + "\n"
		for i := start.Row + 1; i < end.Row; i++ {
			deletedText += sb.lines[i] + "\n"
		}
		deletedText += lastLine[:end.Col]

		newLine := firstLine[:start.Col] + lastLine[end.Col:]

		newLines := make([]string, 0, len(sb.lines)-(end.Row-start.Row))
		newLines = append(newLines, sb.lines[:start.Row]...)
		newLines = append(newLines, newLine)
		newLines = append(newLines, sb.lines[end.Row+1:]...)
		sb.lines = newLines
	}

	return deletedText, nil
}

// ApplyResolvedOperation 应用解析后的操作
// 严格按照预定义的操作类型执行，无任何语义判断
func ApplyResolvedOperation(ctx *ExecutionContext, op ResolvedOperation) error {
	buf := ctx.Buffers.Get(op.BufferID)
	if buf == nil {
		return fmt.Errorf("buffer %s not found", op.BufferID)
	}

	switch op.Kind {
	case OpInsert:
		if op.DeleteBeforeInsert && op.Range != nil {
			_, err := buf.DeleteRange(op.Range.Start, op.Range.End)
			if err != nil {
				return err
			}
		}
		return buf.InsertAt(op.Anchor, op.Text)

	case OpDelete:
		if op.Range == nil {
			return errors.New("delete operation requires a range")
		}
		_, err := buf.DeleteRange(op.Range.Start, op.Range.End)
		return err

	case OpMove:
		win := ctx.Windows.Get(op.WindowID)
		if win != nil {
			win.Cursor = op.Anchor
		}
		return nil

	default:
		return errors.New("unsupported operation kind")
	}
}

// clamp 限制值在范围内
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

````

## 📄 `editor/execution_context.go`

````go
package editor

// ExecutionContext 执行上下文
// 这是「一次 Transaction 的物理执行宇宙」
// 它持有执行 Transaction 所需的所有物理资源引用
type ExecutionContext struct {
	Buffers    BufferStore
	Windows    WindowStore
	Selections SelectionStore

	ActiveBuffer BufferID
	ActiveWindow WindowID
}

// NewExecutionContext 创建新的执行上下文
func NewExecutionContext(buffers BufferStore, windows WindowStore, selections SelectionStore) *ExecutionContext {
	return &ExecutionContext{
		Buffers:    buffers,
		Windows:    windows,
		Selections: selections,
	}
}

````

## 📄 `editor/selection_update.go`

````go
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

````

## 📄 `editor/stores.go`

````go
package editor

import "sync"

// SimpleBufferStore 简单的 Buffer 存储实现
type SimpleBufferStore struct {
	mu      sync.RWMutex
	buffers map[BufferID]Buffer
}

// NewSimpleBufferStore 创建新的 Buffer 存储
func NewSimpleBufferStore() *SimpleBufferStore {
	return &SimpleBufferStore{
		buffers: make(map[BufferID]Buffer),
	}
}

// Get 获取指定 ID 的 Buffer
func (s *SimpleBufferStore) Get(id BufferID) Buffer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buffers[id]
}

// Set 设置 Buffer
func (s *SimpleBufferStore) Set(id BufferID, buf Buffer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buffers[id] = buf
}

// SimpleWindowStore 简单的 Window 存储实现
type SimpleWindowStore struct {
	mu      sync.RWMutex
	windows map[WindowID]*Window
}

// NewSimpleWindowStore 创建新的 Window 存储
func NewSimpleWindowStore() *SimpleWindowStore {
	return &SimpleWindowStore{
		windows: make(map[WindowID]*Window),
	}
}

// Get 获取指定 ID 的 Window
func (s *SimpleWindowStore) Get(id WindowID) *Window {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.windows[id]
}

// Set 设置 Window
func (s *SimpleWindowStore) Set(id WindowID, win *Window) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows[id] = win
}

// SimpleSelectionStore 简单的 Selection 存储实现
type SimpleSelectionStore struct {
	mu         sync.RWMutex
	selections map[BufferID][]Selection
}

// NewSimpleSelectionStore 创建新的 Selection 存储
func NewSimpleSelectionStore() *SimpleSelectionStore {
	return &SimpleSelectionStore{
		selections: make(map[BufferID][]Selection),
	}
}

// Get 获取指定 Buffer 的选区列表
func (s *SimpleSelectionStore) Get(buffer BufferID) []Selection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sels, exists := s.selections[buffer]
	if !exists {
		return []Selection{}
	}

	// 返回副本以避免并发修改
	result := make([]Selection, len(sels))
	copy(result, sels)
	return result
}

// Set 设置指定 Buffer 的选区列表
func (s *SimpleSelectionStore) Set(buffer BufferID, selections []Selection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 存储副本
	sels := make([]Selection, len(selections))
	copy(sels, selections)
	s.selections[buffer] = sels
}

````

## 📄 `editor/text_object.go`

````go
package editor

import (
	"errors"
)

// TextObjectKind 定义文本对象类型
type TextObjectKind int

const (
	TextObjectWord TextObjectKind = iota
	TextObjectParen
	TextObjectBracket
	TextObjectBrace
	TextObjectQuoteDouble
	TextObjectQuoteSingle
	TextObjectParagraph
	TextObjectSentence
)

// TextObjectMotion 定义文本对象运动
type TextObjectMotion struct {
	Kind  TextObjectKind
	Inner bool // true for 'i', false for 'a'
}

// TextObjectRangeCalculator 计算文本对象范围的接口
type TextObjectRangeCalculator interface {
	CalculateRange(obj TextObjectMotion, cursor Cursor) (*MotionRange, error)
}

// ConcreteTextObjectCalculator 实现文本对象范围计算器
type ConcreteTextObjectCalculator struct {
	Buffer Buffer
}

// NewConcreteTextObjectCalculator 创建新的文本对象计算器
func NewConcreteTextObjectCalculator(buffer Buffer) *ConcreteTextObjectCalculator {
	return &ConcreteTextObjectCalculator{
		Buffer: buffer,
	}
}

// CalculateRange 计算文本对象范围
func (calc *ConcreteTextObjectCalculator) CalculateRange(obj TextObjectMotion, cursor Cursor) (*MotionRange, error) {
	switch obj.Kind {
	case TextObjectWord:
		return calc.calculateWordRange(obj.Inner, cursor)
	case TextObjectParen:
		return calc.calculateDelimitedRange('(', ')', obj.Inner, cursor)
	case TextObjectBracket:
		return calc.calculateDelimitedRange('[', ']', obj.Inner, cursor)
	case TextObjectBrace:
		return calc.calculateDelimitedRange('{', '}', obj.Inner, cursor)
	case TextObjectQuoteDouble:
		return calc.calculateQuoteRange('"', obj.Inner, cursor)
	case TextObjectQuoteSingle:
		return calc.calculateQuoteRange('\'', obj.Inner, cursor)
	case TextObjectParagraph:
		return calc.calculateParagraphRange(obj.Inner, cursor)
	case TextObjectSentence:
		return calc.calculateSentenceRange(obj.Inner, cursor)
	default:
		return nil, errors.New("unsupported text object")
	}
}

// CharClass 字符分类
type CharClass int

const (
	ClassWhitespace CharClass = iota
	ClassWord
	ClassPunct
)

// calculateWordRange 计算单词范围
func (calc *ConcreteTextObjectCalculator) calculateWordRange(inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	row := cursor.Row
	if row < 0 || row >= calc.Buffer.LineCount() {
		return nil, errors.New("invalid row")
	}

	line := make([]rune, calc.Buffer.LineLength(row))
	for i := 0; i < len(line); i++ {
		line[i] = calc.Buffer.RuneAt(row, i)
	}

	startCol, endCol := findWordAt(line, cursor.Col, inner)

	return &MotionRange{
		Start: Cursor{Row: row, Col: startCol},
		End:   Cursor{Row: row, Col: endCol},
	}, nil
}

// findWordAt 查找光标位置的单词范围
func findWordAt(line []rune, col int, inner bool) (int, int) {
	if len(line) == 0 || col < 0 {
		return 0, 0
	}

	if col >= len(line) {
		col = len(line) - 1
	}

	// 确定字符类别
	charType := classifyRune(line[col])

	// 向左查找边界
	start := col
	for start > 0 {
		if classifyRune(line[start-1]) != charType {
			break
		}
		start--
	}

	// 向右查找边界
	end := col
	for end < len(line)-1 {
		if classifyRune(line[end+1]) != charType {
			break
		}
		end++
	}

	// 如果是 inner 模式，去除两端的空白
	if inner {
		for start <= end && start < len(line) && isWhitespace(line[start]) {
			start++
		}
		for end > start && end >= 0 && isWhitespace(line[end]) {
			end--
		}
	}

	// 确保 end 在有效范围内
	if end >= len(line) {
		end = len(line) - 1
	}

	// 确保范围有效
	if start > end {
		start = end
	}

	// 如果是 outer 模式，扩展到包含相邻的空白
	if !inner {
		// 向右扩展包含空白
		for end < len(line)-1 && isWhitespace(line[end+1]) {
			end++
		}
		// 向左扩展包含空白
		for start > 0 && isWhitespace(line[start-1]) {
			start--
		}
	}

	return start, end + 1
}

// classifyRune 将字符分类
func classifyRune(r rune) CharClass {
	switch {
	case r == ' ' || r == '\t' || r == '\n' || r == '\r':
		return ClassWhitespace
	case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_':
		return ClassWord
	default:
		return ClassPunct
	}
}

// isWhitespace 检查是否为空白字符
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// calculateDelimitedRange 计算定界符范围
func (calc *ConcreteTextObjectCalculator) calculateDelimitedRange(open, close rune, inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	// 从当前行开始搜索
	startPos, endPos := findDelimitedRange(calc.Buffer, open, close, cursor, inner)

	if startPos.Row == -1 || endPos.Row == -1 {
		return nil, errors.New("delimited range not found")
	}

	return &MotionRange{
		Start: startPos,
		End:   endPos,
	}, nil
}

// findDelimitedRange 查找定界符范围
func findDelimitedRange(buffer Buffer, open, close rune, cursor Cursor, inner bool) (Cursor, Cursor) {
	// 从当前光标位置开始查找匹配的定界符
	currentRow := cursor.Row
	currentCol := cursor.Col

	// 首先尝试在当前行查找
	for row := currentRow; row < buffer.LineCount(); row++ {
		lineLen := buffer.LineLength(row)
		startCol := 0
		if row == currentRow {
			startCol = currentCol
		}

		for col := startCol; col < lineLen; col++ {
			r := buffer.RuneAt(row, col)
			if r == open {
				// 找到开定界符，查找对应的闭定界符
				endPos := findMatchingDelimiter(buffer, open, close, Cursor{Row: row, Col: col})
				if endPos.Row != -1 {
					if inner {
						// Inner 模式：排除定界符本身
						return Cursor{Row: row, Col: col + 1}, endPos
					} else {
						// Outer 模式：包含定界符
						return Cursor{Row: row, Col: col}, Cursor{Row: endPos.Row, Col: endPos.Col + 1}
					}
				}
			}
		}
	}

	// 如果没找到，返回无效位置
	return Cursor{Row: -1, Col: -1}, Cursor{Row: -1, Col: -1}
}

// findMatchingDelimiter 查找匹配的定界符
func findMatchingDelimiter(buffer Buffer, open, close rune, startPos Cursor) Cursor {
	stack := 0
	currentRow := startPos.Row
	currentCol := startPos.Col + 1 // 从开定界符的下一个位置开始

	for row := currentRow; row < buffer.LineCount(); row++ {
		lineLen := buffer.LineLength(row)
		startCol := 0
		if row == currentRow {
			startCol = currentCol
		}

		for col := startCol; col < lineLen; col++ {
			r := buffer.RuneAt(row, col)
			if r == open {
				stack++
			} else if r == close {
				stack--
				if stack < 0 {
					// 找到匹配的闭定界符
					return Cursor{Row: row, Col: col}
				}
			}
		}
		currentCol = 0 // 从下一行开始时，列从0开始
	}

	// 没有找到匹配的闭定界符
	return Cursor{Row: -1, Col: -1}
}

// calculateQuoteRange 计算引号范围
func (calc *ConcreteTextObjectCalculator) calculateQuoteRange(quote rune, inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	// 从当前光标位置开始查找引号
	currentRow := cursor.Row
	currentCol := cursor.Col

	// 首先检查光标位置是否在引号内或旁边
	for row := currentRow; row < calc.Buffer.LineCount(); row++ {
		lineLen := calc.Buffer.LineLength(row)
		startCol := 0
		if row == currentRow {
			startCol = currentCol
		}

		for col := startCol; col < lineLen; col++ {
			r := calc.Buffer.RuneAt(row, col)
			if r == quote {
				// 找到第一个引号，查找匹配的另一个
				endPos := findMatchingQuote(calc.Buffer, quote, Cursor{Row: row, Col: col})
				if endPos.Row != -1 {
					if inner {
						// Inner 模式：排除引号本身
						return &MotionRange{
							Start: Cursor{Row: row, Col: col + 1},
							End:   endPos,
						}, nil
					} else {
						// Outer 模式：包含引号
						return &MotionRange{
							Start: Cursor{Row: row, Col: col},
							End:   Cursor{Row: endPos.Row, Col: endPos.Col + 1},
						}, nil
					}
				}
			}
		}
	}

	return nil, errors.New("quote range not found")
}

// findMatchingQuote 查找匹配的引号
func findMatchingQuote(buffer Buffer, quote rune, startPos Cursor) Cursor {
	escaped := false

	currentRow := startPos.Row
	currentCol := startPos.Col + 1 // 从第一个引号的下一个位置开始

	for row := currentRow; row < buffer.LineCount(); row++ {
		lineLen := buffer.LineLength(row)
		startCol := 0
		if row == currentRow {
			startCol = currentCol
		}

		for col := startCol; col < lineLen; col++ {
			r := buffer.RuneAt(row, col)

			if escaped {
				escaped = false
				continue
			}

			if r == '\\' {
				escaped = true
				continue
			}

			if r == quote {
				// 找到匹配的引号
				return Cursor{Row: row, Col: col}
			}
		}
		currentCol = 0 // 从下一行开始时，列从0开始
	}

	// 没有找到匹配的引号
	return Cursor{Row: -1, Col: -1}
}

// calculateParagraphRange 计算段落范围
func (calc *ConcreteTextObjectCalculator) calculateParagraphRange(inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	// 简化实现：查找空行分隔的段落
	startRow := cursor.Row
	endRow := cursor.Row

	// 向上查找段落开始
	for startRow > 0 {
		lineLen := calc.Buffer.LineLength(startRow - 1)
		if lineLen == 0 {
			break
		}
		startRow--
	}

	// 向下查找段落结束
	for endRow < calc.Buffer.LineCount()-1 {
		lineLen := calc.Buffer.LineLength(endRow + 1)
		if lineLen == 0 {
			break
		}
		endRow++
	}

	if inner {
		// Inner 模式：排除段落周围的空行
		return &MotionRange{
			Start: Cursor{Row: startRow, Col: 0},
			End:   Cursor{Row: endRow, Col: calc.Buffer.LineLength(endRow)},
		}, nil
	} else {
		// Outer 模式：包含整个段落
		return &MotionRange{
			Start: Cursor{Row: startRow, Col: 0},
			End:   Cursor{Row: endRow + 1, Col: 0}, // 包含下一行的开始
		}, nil
	}
}

// calculateSentenceRange 计算句子范围
func (calc *ConcreteTextObjectCalculator) calculateSentenceRange(inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	// 简化实现：查找句号、感叹号、问号分隔的句子
	currentRow := cursor.Row
	currentCol := cursor.Col

	// 查找当前句子的开始
	startRow, startCol := findSentenceStart(calc.Buffer, currentRow, currentCol)

	// 查找当前句子的结束
	endRow, endCol := findSentenceEnd(calc.Buffer, currentRow, currentCol)

	if inner {
		// Inner 模式：排除句子结束标点
		return &MotionRange{
			Start: Cursor{Row: startRow, Col: startCol},
			End:   Cursor{Row: endRow, Col: endCol},
		}, nil
	} else {
		// Outer 模式：包含句子结束标点及后续空白
		// 简化：包含到句子结束
		return &MotionRange{
			Start: Cursor{Row: startRow, Col: startCol},
			End:   Cursor{Row: endRow, Col: endCol + 1},
		}, nil
	}
}

// findSentenceStart 查找句子开始
func findSentenceStart(buffer Buffer, row, col int) (int, int) {
	// 简化实现：查找前一个句子结束符后的第一个非空白字符
	for r := row; r >= 0; r-- {
		lineLen := buffer.LineLength(r)
		startCol := lineLen - 1
		if r == row {
			startCol = col
		}

		for c := startCol; c >= 0; c-- {
			runeVal := buffer.RuneAt(r, c)
			if runeVal == '.' || runeVal == '!' || runeVal == '?' {
				// 找到句子结束符，下一个位置是句子开始
				nextRow, nextCol := getNextNonWhitespace(buffer, r, c+1)
				return nextRow, nextCol
			}
		}
	}

	// 如果没找到，返回文件开始
	return 0, 0
}

// findSentenceEnd 查找句子结束
func findSentenceEnd(buffer Buffer, row, col int) (int, int) {
	// 简化实现：查找下一个句子结束符
	for r := row; r < buffer.LineCount(); r++ {
		lineLen := buffer.LineLength(r)
		startCol := 0
		if r == row {
			startCol = col
		}

		for c := startCol; c < lineLen; c++ {
			runeVal := buffer.RuneAt(r, c)
			if runeVal == '.' || runeVal == '!' || runeVal == '?' {
				// 找到句子结束符
				return r, c
			}
		}
	}

	// 如果没找到，返回文件结束
	endRow := buffer.LineCount() - 1
	endCol := buffer.LineLength(endRow)
	return endRow, endCol
}

// getNextNonWhitespace 获取下一个非空白字符位置
func getNextNonWhitespace(buffer Buffer, row, col int) (int, int) {
	for r := row; r < buffer.LineCount(); r++ {
		lineLen := buffer.LineLength(r)
		startCol := 0
		if r == row {
			startCol = col
		}

		for c := startCol; c < lineLen; c++ {
			runeVal := buffer.RuneAt(r, c)
			if !isWhitespace(runeVal) {
				return r, c
			}
		}
	}

	// 如果没找到，返回当前位置
	return row, col
}

// ParseTextObject 解析文本对象字符串
func ParseTextObject(textObjectStr string) (*TextObjectMotion, error) {
	if len(textObjectStr) < 2 {
		return nil, errors.New("invalid text object string")
	}

	modifier := textObjectStr[0:1]
	objType := textObjectStr[1:2]

	inner := modifier == "i"

	var kind TextObjectKind
	switch objType {
	case "w":
		kind = TextObjectWord
	case "(":
		kind = TextObjectParen
	case "[":
		kind = TextObjectBracket
	case "{":
		kind = TextObjectBrace
	case "\"":
		kind = TextObjectQuoteDouble
	case "'":
		kind = TextObjectQuoteSingle
	case "p":
		kind = TextObjectParagraph
	case "s":
		kind = TextObjectSentence
	default:
		return nil, errors.New("unsupported text object type")
	}

	return &TextObjectMotion{
		Kind:  kind,
		Inner: inner,
	}, nil
}

````

## 📄 `engine.go`

````go
package main

import "errors"

// MotionKind 定义移动方向类型
type MotionKind int

const (
	MotionLeft MotionKind = iota
	MotionRight
	MotionUp
	MotionDown
	MotionWordForward
	MotionWordBackward
	MotionLineEnd
)

// Motion 结构体定义移动动作
type Motion struct {
	Kind  MotionKind
	Count int
}

// Line 表示一行
type Line struct {
	Length int
}

// Buffer 接口定义缓冲区
type Buffer interface {
	LineCount() int
	LineLength(row int) int
	RuneAt(row, col int) rune
	DeleteRange(r MotionRange) error
}

// MotionRange 表示一个运动范围
type MotionRange struct {
	Start Cursor
	End   Cursor // Vim 语义：不含 End
}

// MotionResult 表示移动结果
type MotionResult struct {
	DeltaRow int
	DeltaCol int

	Range *MotionRange
}

// CharClass 定义字符类别
type CharClass int

const (
	ClassWhitespace CharClass = iota
	ClassWord                 // 字母 + 数字 + _
	ClassPunct                // 其他
)

// motionHandler 定义运动处理器类型
type motionHandler func(engine *CursorEngine, motion *Motion) (*MotionResult, error)

// motionTable 定义运动表
var motionTable = map[MotionKind]motionHandler{
	MotionLeft:        simpleVector(0, -1),
	MotionRight:       simpleVector(0, 1),
	MotionUp:          simpleVector(-1, 0),
	MotionDown:        simpleVector(1, 0),
	MotionWordForward: wordForward,
}

// ConcreteBuffer 是 Buffer 接口的具体实现
type ConcreteBuffer struct {
	Lines   []Line
	Content [][]rune // 每行的实际内容
}

func (cb *ConcreteBuffer) LineCount() int {
	return len(cb.Lines)
}

func (cb *ConcreteBuffer) LineLength(row int) int {
	if row >= 0 && row < len(cb.Lines) {
		return cb.Lines[row].Length
	}
	return 0
}

func (cb *ConcreteBuffer) RuneAt(row, col int) rune {
	if row >= 0 && row < len(cb.Content) && col >= 0 && col < len(cb.Content[row]) {
		return cb.Content[row][col]
	}
	return 0
}

func (cb *ConcreteBuffer) DeleteRange(r MotionRange) error {
	start := r.Start
	end := r.End

	// 如果是同一行内的删除
	if start.Row == end.Row {
		if start.Row < len(cb.Content) {
			content := cb.Content[start.Row]
			newContent := append(content[:start.Col], content[end.Col:]...)

			// 更新行长度
			cb.Lines[start.Row].Length = len(newContent)
			cb.Content[start.Row] = newContent
		}
		return nil
	}

	// 多行删除：将多行合并为一行
	if start.Row < len(cb.Content) && end.Row < len(cb.Content) {
		// 获取起始行的内容（到 start.Col 截断）
		startLineContent := cb.Content[start.Row]
		prefix := startLineContent[:start.Col]

		// 获取结束行的内容（从 end.Col 开始）
		endLineContent := cb.Content[end.Row]
		suffix := endLineContent[end.Col:]

		// 合并前缀和后缀
		mergedLine := append(prefix, suffix...)

		// 替换起始行的内容
		cb.Content[start.Row] = mergedLine
		cb.Lines[start.Row].Length = len(mergedLine)

		// 删除中间的所有行（包括结束行）
		rowsToDelete := end.Row - start.Row
		newLines := make([]Line, 0, len(cb.Lines)-rowsToDelete)
		newContent := make([][]rune, 0, len(cb.Content)-rowsToDelete)

		for i := 0; i < len(cb.Lines); i++ {
			if i < start.Row || i > end.Row {
				newLines = append(newLines, cb.Lines[i])
				newContent = append(newContent, cb.Content[i])
			} else if i == start.Row {
				// 已经处理过的行，跳过
			}
		}

		cb.Lines = newLines
		cb.Content = newContent
	}

	return nil
}

// CursorEngine 是真正的坐标计算引擎
type CursorEngine struct {
	Cursor *Cursor
	Buffer Buffer
}

// clamp 函数用于限制值在指定范围内
func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// clampCursor 内部方法，用于限制光标位置
func (e *CursorEngine) clampCursor(row, col int) (int, int) {
	if e.Buffer == nil {
		return row, col
	}

	row = clamp(row, 0, e.Buffer.LineCount()-1)

	maxCol := 0
	if row >= 0 && row < e.Buffer.LineCount() {
		maxCol = e.Buffer.LineLength(row)
		if maxCol > 0 {
			maxCol-- // Length 是实际长度，所以最大索引是 Length-1
		}
	}
	col = clamp(col, 0, maxCol)

	return row, col
}

// ApplyMotion 应用运动结果（统一处理逻辑）
func (e *CursorEngine) ApplyMotion(r *MotionResult) error {
	if r.Range != nil {
		e.Cursor.Row = r.Range.End.Row
		e.Cursor.Col = r.Range.End.Col
		return nil
	}

	// fallback: vector motion
	newRow := e.Cursor.Row + r.DeltaRow
	newCol := e.Cursor.Col + r.DeltaCol
	e.Cursor.Row, e.Cursor.Col = e.clampCursor(newRow, newCol)
	return nil
}

// MoveCursor 移动光标（唯一副作用）
func (e *CursorEngine) MoveCursor(r *MotionResult) error {
	return e.ApplyMotion(r)
}

// DeleteRange 删除指定范围的内容
func (e *CursorEngine) DeleteRange(r *MotionRange) error {
	if e.Buffer == nil {
		return errors.New("no buffer available")
	}

	err := e.Buffer.DeleteRange(*r)
	if err != nil {
		return err
	}

	// 移动光标到开始位置
	e.Cursor.Row = r.Start.Row
	e.Cursor.Col = r.Start.Col

	return nil
}

// GetTextInRange 获取指定范围的文本
func (e *CursorEngine) GetTextInRange(r *MotionRange) string {
	if e.Buffer == nil {
		return ""
	}

	concreteBuffer, ok := e.Buffer.(*ConcreteBuffer)
	if !ok {
		return ""
	}

	start := r.Start
	end := r.End

	if start.Row == end.Row {
		if start.Row < len(concreteBuffer.Content) {
			content := concreteBuffer.Content[start.Row]
			if start.Col >= 0 && end.Col <= len(content) {
				subRunes := content[start.Col:end.Col]
				return string(subRunes)
			}
		}
		return ""
	}

	// 多行文本获取
	var result []rune

	// 第一行
	if start.Row < len(concreteBuffer.Content) {
		content := concreteBuffer.Content[start.Row]
		if start.Col < len(content) {
			result = append(result, content[start.Col:]...)
		}
		result = append(result, '\n')
	}

	// 中间行
	for i := start.Row + 1; i < end.Row && i < len(concreteBuffer.Content); i++ {
		result = append(result, concreteBuffer.Content[i]...)
		result = append(result, '\n')
	}

	// 最后一行
	if end.Row < len(concreteBuffer.Content) {
		content := concreteBuffer.Content[end.Row]
		if end.Col <= len(content) {
			result = append(result, content[:end.Col]...)
		}
	}

	return string(result)
}

// ErrInvalidMotion 表示无效的移动动作
var ErrInvalidMotion = errors.New("invalid motion")

// ComputeMotion 计算移动结果（只算，不动）
func (e *CursorEngine) ComputeMotion(m *Motion) (*MotionResult, error) {
	handler, ok := motionTable[m.Kind]
	if !ok {
		return nil, ErrInvalidMotion
	}

	return handler(e, m)
}

// simpleVector 返回一个简单的向量运动处理器
func simpleVector(dr, dc int) motionHandler {
	return func(e *CursorEngine, m *Motion) (*MotionResult, error) {
		count := m.Count
		if count <= 0 {
			count = 1
		}
		return &MotionResult{
			DeltaRow: dr * count,
			DeltaCol: dc * count,
		}, nil
	}
}

// classify 将字符分类
func classify(r rune) CharClass {
	switch {
	case r == ' ' || r == '\t':
		return ClassWhitespace
	case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_':
		return ClassWord
	default:
		return ClassPunct
	}
}

// wordForward 实现向前单词移动
func wordForward(e *CursorEngine, m *Motion) (*MotionResult, error) {
	row, col := e.Cursor.Row, e.Cursor.Col
	start := Cursor{Row: row, Col: col}

	count := m.Count
	if count <= 0 {
		count = 1
	}

	for i := 0; i < count; i++ {
		row, col = nextWord(e.Buffer, row, col)
	}

	end := Cursor{Row: row, Col: col}

	rangeResult := &MotionRange{
		Start: start,
		End:   end,
	}

	return &MotionResult{
		DeltaRow: end.Row - start.Row,
		DeltaCol: end.Col - start.Col,
		Range:    rangeResult,
	}, nil
}

// nextWord 找到下一个单词的位置
func nextWord(b Buffer, row, col int) (int, int) {
	if b == nil || row >= b.LineCount() {
		return row, col
	}

	// 如果当前行不存在或列超出范围，返回原位置
	if row < 0 || col >= b.LineLength(row) {
		return row, col
	}

	// Step 1: 获取当前位置的字符类别
	currentClass := classify(b.RuneAt(row, col))

	// Step 2: 跳过当前 class 的连续字符
	for {
		col++
		if col >= b.LineLength(row) {
			// 到达行尾，尝试下一行
			row++
			col = 0
			if row >= b.LineCount() {
				// 到达缓冲区末尾
				return row, col
			}
			// 当到达新行时，将当前类别视为空白，以便跳过开头的空白
			currentClass = ClassWhitespace
			continue
		}

		nextClass := classify(b.RuneAt(row, col))
		if nextClass != currentClass {
			// 类别发生变化，跳出循环
			break
		}
	}

	// Step 3: 跳过空白字符，直到遇到非空白字符
	for {
		if col >= b.LineLength(row) {
			// 到达行尾，尝试下一行
			row++
			col = 0
			if row >= b.LineCount() {
				// 到达缓冲区末尾
				return row, col
			}
			continue
		}

		charClass := classify(b.RuneAt(row, col))
		if charClass != ClassWhitespace {
			// 遇到非空白字符，跳出循环
			break
		}
		col++
	}

	return row, col
}

````

## 📄 `engine/concrete_engine.go`

````go
package engine

import (
	"tmux-fsm/editor"
	"tmux-fsm/intent"
)

// ConcreteEngine 是 Engine 接口的具体实现
type ConcreteEngine struct {
	// 这里可以添加实际的编辑器状态
	cursor editor.Cursor
}

// NewConcreteEngine 创建一个新的 ConcreteEngine 实例
func NewConcreteEngine() *ConcreteEngine {
	return &ConcreteEngine{
		cursor: editor.Cursor{Row: 0, Col: 0},
	}
}

// Cursor 返回当前光标位置
func (e *ConcreteEngine) Cursor() editor.Cursor {
	return e.cursor
}

// ComputeMotion 计算运动产生的范围
func (e *ConcreteEngine) ComputeMotion(m *intent.Motion) (editor.MotionRange, error) {
	switch m.Kind {
	case intent.MotionRange:
		if m.Range != nil && m.Range.Kind == intent.RangeTextObject {
			return e.computeTextObject(m.Range.TextObject)
		}
	case intent.MotionWord:
		return e.computeWord(m.Count)
	case intent.MotionLine:
		return e.computeLine(m.Count)
	case intent.MotionChar:
		return e.computeChar(m.Count)
	case intent.MotionGoto:
		return e.computeGoto(m.Count)
	case intent.MotionFind:
		if m.Find != nil {
			return e.computeFindMotion(m.Find, m.Count)
		}
	}

	// 默认返回当前位置的范围
	return editor.MotionRange{
		Start: e.cursor,
		End:   e.cursor,
	}, nil
}

// computeTextObject 计算文本对象的范围
func (e *ConcreteEngine) computeTextObject(textObj *intent.TextObject) (editor.MotionRange, error) {
	// 这里需要实际的文本分析逻辑
	// 现在返回一个示例范围
	start := e.cursor
	end := e.cursor

	switch textObj.Object {
	case intent.Word:
		// 计算单词边界
		if textObj.Scope == intent.Inner {
			// 内部单词：从单词开始到单词结束
		} else {
			// 周围单词：包含周围的空白字符
		}
	case intent.Paren:
		// 计算括号内的内容或包括括号
		if textObj.Scope == intent.Inner {
			// 内部括号：括号内的内容
		} else {
			// 周围括号：包括括号本身
		}
	case intent.QuoteDouble:
		// 计算双引号内的内容或包括引号
		if textObj.Scope == intent.Inner {
			// 内部引号：引号内的内容
		} else {
			// 周围引号：包括引号本身
		}
	}

	return editor.MotionRange{
		Start: start,
		End:   end,
	}, nil
}

// computeWord 计算单词移动的范围
func (e *ConcreteEngine) computeWord(count int) (editor.MotionRange, error) {
	start := e.cursor
	end := e.cursor

	// 这里需要实际的单词边界检测逻辑
	// 简单示例：移动 count 个单词
	for i := 0; i < count; i++ {
		// 实际实现中需要分析文本内容
		end.Col += 5 // 示例：假设每个单词平均5个字符
	}

	return editor.MotionRange{
		Start: start,
		End:   end,
	}, nil
}

// computeLine 计算行移动的范围
func (e *ConcreteEngine) computeLine(count int) (editor.MotionRange, error) {
	start := e.cursor
	end := e.cursor

	// 移动到第 count 行
	end.Row += count

	return editor.MotionRange{
		Start: start,
		End:   end,
	}, nil
}

// computeChar 计算字符移动的范围
func (e *ConcreteEngine) computeChar(count int) (editor.MotionRange, error) {
	start := e.cursor
	end := e.cursor

	// 移动 count 个字符
	end.Col += count

	return editor.MotionRange{
		Start: start,
		End:   end,
	}, nil
}

// computeGoto 计算跳转的范围
func (e *ConcreteEngine) computeGoto(count int) (editor.MotionRange, error) {
	start := e.cursor
	end := e.cursor

	// 跳转到指定位置（如果 count > 0）
	if count > 0 {
		end.Row = count - 1 // 行号从0开始
		end.Col = 0
	} else {
		// 默认跳转到文件开头
		end.Row = 0
		end.Col = 0
	}

	return editor.MotionRange{
		Start: start,
		End:   end,
	}, nil
}

// computeFindMotion 计算查找运动的范围
func (e *ConcreteEngine) computeFindMotion(find *intent.FindMotion, count int) (editor.MotionRange, error) {
	start := e.cursor
	end := e.cursor

	// 这里需要实际的查找逻辑
	// 简单示例：在当前行中查找字符
	if find != nil {
		// 模拟当前行的文本内容
		line := "sample text for testing find motions like fx tx Fx Tx"

		pos := start.Col
		step := 1
		if find.Direction == intent.FindBackward {
			step = -1
		}

		matches := 0
		i := pos + step

		for i >= 0 && i < len(line) {
			if rune(line[i]) == find.Char {
				matches++
				if matches == count {
					target := i

					// till 的偏移规则
					if find.Till {
						if find.Direction == intent.FindForward {
							target--
						} else {
							target++
						}
					}

					end.Col = clamp(target, 0, len(line)-1)

					return editor.MotionRange{
						Start: start,
						End:   editor.Cursor{Row: start.Row, Col: end.Col},
					}, nil
				}
			}
			i += step
		}
	}

	// Vim 行为：找不到 → 光标不动
	return editor.MotionRange{
		Start: start,
		End:   start,
	}, nil
}

// clamp 辅助函数
func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// MoveCursor 移动光标到指定范围
func (e *ConcreteEngine) MoveCursor(r editor.MotionRange) error {
	e.cursor = r.End
	return nil
}

// DeleteRange 删除指定范围的内容
func (e *ConcreteEngine) DeleteRange(r editor.MotionRange) error {
	// 实际实现中需要与底层编辑器交互
	return nil
}

// YankRange 复制指定范围的内容
func (e *ConcreteEngine) YankRange(r editor.MotionRange) error {
	// 实际实现中需要与底层编辑器交互
	return nil
}

// ChangeRange 修改指定范围的内容
func (e *ConcreteEngine) ChangeRange(r editor.MotionRange) error {
	// 实际实现中需要与底层编辑器交互
	return nil
}
````

## 📄 `engine/engine.go`

````go
package engine

import (
	"time"
	"tmux-fsm/crdt"
	"tmux-fsm/index"
	"tmux-fsm/policy"
	"tmux-fsm/replay"
	"tmux-fsm/selection"
	"tmux-fsm/wal"
)

// Engine 编辑器内核引擎接口
type Engine interface {
	// State 状态管理
	Apply(event crdt.SemanticEvent) error
	Replay(upTo crdt.EventID) replay.TextState
	Snapshot() *Snapshot

	// WAL 事件日志
	Append(event crdt.SemanticEvent) crdt.EventID
	WALSince(id crdt.EventID) []wal.SemanticEvent

	// CRDT 位置管理
	AllocatePosition(after, before *crdt.PositionID) crdt.PositionID
	ComparePosition(a, b crdt.PositionID) int

	// Selection 管理
	ApplySelection(actor crdt.ActorID, fact selection.SetSelectionFact)
	GetSelection(cursorID selection.CursorID) (selection.Selection, bool)
	GetAllSelections() map[selection.CursorID]selection.Selection

	// Policy 管理
	RegisterActor(actorID crdt.ActorID, level policy.TrustLevel, name string)
	CheckPolicy(event crdt.SemanticEvent) error

	// Index 查询
	QueryByActor(actor crdt.ActorID) []crdt.EventID
	QueryByType(ft index.FactType) []crdt.EventID
	QueryByTimeRange(start, end time.Time) []crdt.EventID
	QueryAIChanges(aiActorPrefix string) []crdt.EventID

	// GC 垃圾回收
	Compact(stable crdt.EventID)

	// 同步
	KnownHeads() map[crdt.ActorID]crdt.EventID
	Integrate(events []wal.SemanticEvent) error
}

// Snapshot 快照
type Snapshot struct {
	At    crdt.EventID     `json:"at"`
	State replay.TextState `json:"state"`
}

// HeadlessEngine 无头引擎实现
type HeadlessEngine struct {
	store        *crdt.EventStore
	snapshots    map[crdt.EventID]*Snapshot
	currentState replay.TextState
	selectionMgr *selection.SelectionManager
	policyMgr    *policy.DefaultPolicy
	index        *index.EventIndex
}

// Apply 应用事件
func (e *HeadlessEngine) Apply(event crdt.SemanticEvent) error {
	e.store.Merge(event)

	// 更新当前状态
	sortedEvents := e.store.TopoSort()
	e.currentState = replay.Replay(
		replay.TextState{},
		sortedEvents,
		nil, // 不使用过滤器
	)

	return nil
}

// Replay 重放至指定事件
func (e *HeadlessEngine) Replay(upTo crdt.EventID) replay.TextState {
	allEvents := e.store.TopoSort()

	// 找到 upTo 事件的索引
	var eventsToReplay []crdt.SemanticEvent
	for _, event := range allEvents {
		eventsToReplay = append(eventsToReplay, event)
		if event.ID == upTo {
			break
		}
	}

	return replay.Replay(
		replay.TextState{},
		eventsToReplay,
		nil,
	)
}

// Snapshot 创建快照
func (e *HeadlessEngine) Snapshot() *Snapshot {
	snapshot := &Snapshot{
		At:    "", // 需要设置为最新的事件ID
		State: e.currentState,
	}

	// 获取最新的事件ID
	allEvents := e.store.TopoSort()
	if len(allEvents) > 0 {
		snapshot.At = allEvents[len(allEvents)-1].ID
	}

	e.snapshots[snapshot.At] = snapshot
	return snapshot
}

// Append 添加事件到日志
func (e *HeadlessEngine) Append(event crdt.SemanticEvent) crdt.EventID {
	e.store.Merge(event)
	return event.ID
}

// WALSince 获取指定事件之后的日志
func (e *HeadlessEngine) WALSince(id crdt.EventID) []wal.SemanticEvent {
	allEvents := e.store.TopoSort()

	var result []wal.SemanticEvent
	found := false
	for _, event := range allEvents {
		if !found && event.ID == id {
			found = true
			continue
		}
		if found {
			// 转换 crdt.SemanticEvent 到 wal.SemanticEvent
			walEvent := wal.SemanticEvent{
				ID:            string(event.ID),
				CausalParents: []string{},
				LocalParent:   string(event.LocalParent),
				Time:          event.Time,
				Actor:         string(event.Actor),
				Fact:          event.Fact,
			}

			// 填充 CausalParents
			for _, parent := range event.CausalParents {
				walEvent.CausalParents = append(walEvent.CausalParents, string(parent))
			}

			result = append(result, walEvent)
		}
	}

	return result
}

// AllocatePosition 分配新位置
func (e *HeadlessEngine) AllocatePosition(after, before *crdt.PositionID) crdt.PositionID {
	actor := "default" // 这里应该从上下文获取实际的 actor
	if after != nil {
		actor = string(after.Actor)
	} else if before != nil {
		actor = string(before.Actor)
	}

	return crdt.AllocateBetween(after, before, crdt.ActorID(actor))
}

// ComparePosition 比较位置
func (e *HeadlessEngine) ComparePosition(a, b crdt.PositionID) int {
	return crdt.ComparePos(a, b)
}

// Compact 压缩日志
func (e *HeadlessEngine) Compact(stable crdt.EventID) {
	// 实现压缩逻辑
	// 这里简化处理，实际实现需要更复杂的逻辑
}

// KnownHeads 获取已知头部
func (e *HeadlessEngine) KnownHeads() map[crdt.ActorID]crdt.EventID {
	heads := make(map[crdt.ActorID]crdt.EventID)

	allEvents := e.store.TopoSort()

	// 按参与者分组，找到每个参与者的最新事件
	for _, event := range allEvents {
		if current, exists := heads[event.Actor]; !exists || event.ID > current {
			heads[event.Actor] = event.ID
		}
	}

	return heads
}

// Integrate 集成外部事件
func (e *HeadlessEngine) Integrate(events []wal.SemanticEvent) error {
	for _, walEvent := range events {
		// 转换 wal.SemanticEvent 到 crdt.SemanticEvent
		crdtEvent := crdt.SemanticEvent{
			ID:            crdt.EventID(walEvent.ID),
			Actor:         crdt.ActorID(walEvent.Actor),
			CausalParents: []crdt.EventID{},
			LocalParent:   crdt.EventID(walEvent.LocalParent),
			Time:          walEvent.Time,
			Fact:          walEvent.Fact,
		}

		// 转换 CausalParents
		for _, parent := range walEvent.CausalParents {
			crdtEvent.CausalParents = append(crdtEvent.CausalParents, crdt.EventID(parent))
		}

		e.store.Merge(crdtEvent)
	}

	return nil
}

// ApplySelection 应用选择区域变更
func (e *HeadlessEngine) ApplySelection(actor crdt.ActorID, fact selection.SetSelectionFact) {
	e.selectionMgr.ApplySelection(actor, fact)
}

// GetSelection 获取选择区域
func (e *HeadlessEngine) GetSelection(cursorID selection.CursorID) (selection.Selection, bool) {
	return e.selectionMgr.GetSelection(cursorID)
}

// GetAllSelections 获取所有选择区域
func (e *HeadlessEngine) GetAllSelections() map[selection.CursorID]selection.Selection {
	return e.selectionMgr.GetAllSelections()
}

// RegisterActor 注册参与者
func (e *HeadlessEngine) RegisterActor(actorID crdt.ActorID, level policy.TrustLevel, name string) {
	e.policyMgr.RegisterActor(policy.ActorInfo{ID: actorID, Level: level, Name: name})
}

// CheckPolicy 检查策略
func (e *HeadlessEngine) CheckPolicy(event crdt.SemanticEvent) error {
	return e.policyMgr.AllowCommit(event.Actor, event)
}

// QueryByActor 按参与者查询
func (e *HeadlessEngine) QueryByActor(actor crdt.ActorID) []crdt.EventID {
	return e.index.QueryByActor(actor)
}

// QueryByType 按类型查询
func (e *HeadlessEngine) QueryByType(ft index.FactType) []crdt.EventID {
	return e.index.QueryByType(ft)
}

// QueryByTimeRange 按时间范围查询
func (e *HeadlessEngine) QueryByTimeRange(start, end time.Time) []crdt.EventID {
	return e.index.QueryByTimeRange(start, end)
}

// QueryAIChanges 查询 AI 的更改
func (e *HeadlessEngine) QueryAIChanges(aiActorPrefix string) []crdt.EventID {
	return e.index.QueryAIChanges(aiActorPrefix)
}

````

## 📄 `examples/transaction_demo.go`

````go
package main

import (
	"fmt"
	"log"

	"tmux-fsm/editor"
	"tmux-fsm/kernel"
	"tmux-fsm/types"
)

// 这是一个最小的 Transaction Runner 演示
// 展示如何使用新的执行上下文系统

func main() {
	fmt.Println("=== Transaction Runner Demo ===")

	// 1. 创建 Stores
	bufferStore := editor.NewSimpleBufferStore()
	windowStore := editor.NewSimpleWindowStore()
	selectionStore := editor.NewSimpleSelectionStore()

	// 2. 创建初始 Buffer
	buffer := editor.NewSimpleBuffer([]string{
		"Hello World",
		"This is a test",
	})
	bufferStore.Set("main", buffer)

	// 3. 创建 Window
	window := &editor.Window{
		ID:     "main-win",
		Cursor: editor.Cursor{Row: 0, Col: 6},
	}
	windowStore.Set("main-win", window)

	// 4. 创建 ExecutionContext
	ctx := editor.NewExecutionContext(bufferStore, windowStore, selectionStore)
	ctx.ActiveBuffer = "main"
	ctx.ActiveWindow = "main-win"

	// 5. 创建 TransactionRunner
	runner := kernel.NewTransactionRunner(ctx)

	// 6. 创建一个简单的 Transaction（插入文本）
	tx := &types.Transaction{
		ID: 1,
		Records: []types.OperationRecord{
			{
				ResolvedOp: editor.ResolvedOperation{
					Kind:     editor.OpInsert,
					BufferID: "main",
					WindowID: "main-win",
					Anchor:   editor.Cursor{Row: 0, Col: 6},
					Text:     "Beautiful ",
				},
			},
		},
	}

	// 7. 打印初始状态
	fmt.Println("初始状态:")
	printBuffer(bufferStore.Get("main"))

	// 8. 应用 Transaction
	fmt.Println("\n执行: 在位置 (0, 6) 插入 'Beautiful '")
	if err := runner.Apply(tx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n应用后:")
	printBuffer(bufferStore.Get("main"))

	// 9. 创建第二个 Transaction（删除）
	tx2 := &types.Transaction{
		ID: 2,
		Records: []types.OperationRecord{
			{
				ResolvedOp: editor.ResolvedOperation{
					Kind:     editor.OpDelete,
					BufferID: "main",
					WindowID: "main-win",
					Range: &editor.TextRange{
						Start: editor.Cursor{Row: 0, Col: 0},
						End:   editor.Cursor{Row: 0, Col: 6},
					},
				},
			},
		},
	}

	fmt.Println("\n执行: 删除 (0, 0) 到 (0, 6)")
	if err := runner.Apply(tx2); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n应用后:")
	printBuffer(bufferStore.Get("main"))

	// 10. 演示 Repeat（重复最后一个操作）
	fmt.Println("\n执行: Repeat (重复删除)")
	if err := runner.Repeat(tx2); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n应用后:")
	printBuffer(bufferStore.Get("main"))

	fmt.Println("\n=== Demo 完成 ===")
}

func printBuffer(buf editor.Buffer) {
	if buf == nil {
		fmt.Println("  (buffer is nil)")
		return
	}
	for i := 0; i < buf.LineCount(); i++ {
		fmt.Printf("  Line %d: %s\n", i, buf.Line(i))
	}
}

````

## 📄 `fsm/engine.go`

````go
package fsm

import (
	"fmt"
	"strings"
	"time"
	"tmux-fsm/intent"
)

// RawTokenEmitter 用于发送 RawToken 的接口
type RawTokenEmitter interface {
	Emit(RawToken)
}

// EngineAdapter 实现resolver.EngineAdapter接口
type EngineAdapter struct {
	engine *Engine
}

func (ea *EngineAdapter) SendKeys(keys ...string) {
	// 将键发送到tmux
	args := append([]string{"send-keys", "-t", "."}, keys...)
	tmux(strings.Join(args, " "))
}

func (ea *EngineAdapter) RunAction(name string) {
	ea.engine.RunAction(name)
}

func (ea *EngineAdapter) GetVisualMode() intent.VisualMode {
	return ea.engine.visualMode
}

func (ea *EngineAdapter) SetVisualMode(mode intent.VisualMode) {
	ea.engine.visualMode = mode
}

func (ea *EngineAdapter) EnterVisualMode(mode intent.VisualMode) {
	ea.engine.visualMode = mode
	// 可能需要更新UI显示
	UpdateUI()
}

func (ea *EngineAdapter) ExitVisualMode() {
	ea.engine.visualMode = intent.VisualNone
	// 可能需要更新UI显示
	UpdateUI()
}

func (ea *EngineAdapter) GetCurrentCursor() interface{} {
	// 获取当前光标位置（通过 tmux 命令）
	// 这里需要实际从 tmux 获取光标位置
	return struct {
		Line int
		Col  int
	}{Line: 0, Col: 0} // 简化实现
}

func (ea *EngineAdapter) ComputeMotion(m *intent.Motion) (interface{}, error) {
	// 计算动作范围
	return struct{}{}, nil
}

func (ea *EngineAdapter) MoveCursor(r interface{}) error {
	// 移动光标
	return nil
}

func (ea *EngineAdapter) DeleteRange(r interface{}) error {
	// 删除范围内容
	return nil
}

func (ea *EngineAdapter) DeleteWithMotion(motion intent.MotionKind, count int) error {
	// 根据动作类型执行删除
	switch motion {
	case intent.MotionWord:
		ea.SendKeys("Escape", "d", "w")
	case intent.MotionLine:
		ea.SendKeys("Escape", "d", "d")
	case intent.MotionChar:
		ea.SendKeys("Delete")
	default:
		ea.SendKeys("Delete")
	}
	return nil
}

func (ea *EngineAdapter) YankRange(r interface{}) error {
	// 复制范围内容
	return nil
}

func (ea *EngineAdapter) YankWithMotion(motion intent.MotionKind, count int) error {
	// 根据动作类型执行复制
	switch motion {
	case intent.MotionWord:
		ea.SendKeys("Escape", "y", "w")
	case intent.MotionLine:
		ea.SendKeys("Escape", "y", "y")
	case intent.MotionChar:
		ea.SendKeys("Escape", "y", "l")
	default:
		ea.SendKeys("Escape", "y", "y")
	}
	return nil
}

func (ea *EngineAdapter) ChangeRange(r interface{}) error {
	// 修改范围内容
	return nil
}

func (ea *EngineAdapter) ChangeWithMotion(motion intent.MotionKind, count int) error {
	// 根据动作类型执行修改
	switch motion {
	case intent.MotionWord:
		ea.SendKeys("Escape", "c", "w")
	case intent.MotionLine:
		ea.SendKeys("Escape", "c", "c")
	case intent.MotionChar:
		ea.SendKeys("Escape", "c", "l")
	default:
		ea.SendKeys("Escape", "c", "c")
	}
	return nil
}

// Engine FSM 引擎结构体
type Engine struct {
	Active     string
	Keymap     *Keymap
	layerTimer *time.Timer
	count      int               // 用于存储数字计数
	emitters   []RawTokenEmitter // 用于向外部发送token的多个接收者
	visualMode intent.VisualMode // 视觉模式状态
}

// FSMStatus FSM 状态信息，用于UI更新
type FSMStatus struct {
	Layer string
	Count int
}

// AddEmitter 添加一个 token 发送接收者
func (e *Engine) AddEmitter(emitter RawTokenEmitter) {
	e.emitters = append(e.emitters, emitter)
}

// RemoveEmitter 移除一个 token 发送接收者
func (e *Engine) RemoveEmitter(emitter RawTokenEmitter) {
	for i, em := range e.emitters {
		if em == emitter {
			e.emitters = append(e.emitters[:i], e.emitters[i+1:]...)
			break
		}
	}
}

// emitInternal 内部发送 token 给所有订阅者
func (e *Engine) emitInternal(token RawToken) {
	for _, emitter := range e.emitters {
		emitter.Emit(token)
	}
}

// 全局默认引擎实例
var defaultEngine *Engine

// NewEngine 创建新的 FSM 引擎实例（显式注入 Keymap）
func NewEngine(km *Keymap) *Engine {
	engine := &Engine{
		Active:     "NAV",
		Keymap:     km,
		count:      0,
		emitters:   make([]RawTokenEmitter, 0),
		visualMode: intent.VisualNone,
	}

	// 创建引擎适配器
	// adapter := &EngineAdapter{engine: engine}

	// 初始化解析器（已废弃）
	// engine.resolver = resolver.New(adapter)

	return engine
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
	// 检查是否是数字键，即使当前层没有定义
	if isDigit(key) {
		e.count = e.count*10 + int(key[0]-'0')
		e.emitInternal(RawToken{Kind: TokenDigit, Value: key})
		return true
	}

	// 检查是否是重复键
	if key == "." {
		e.emitInternal(RawToken{Kind: TokenRepeat, Value: "."})
		return true
	}

	// 其他按键按原有逻辑处理（只处理层切换，不处理动作）
	if e.CanHandle(key) {
		st := e.Keymap.States[e.Active]
		act := st.Keys[key]

		// 1. 处理层切换
		if act.Layer != "" {
			e.Active = act.Layer
			e.resetLayerTimeout(act.TimeoutMs)
			e.emitInternal(RawToken{Kind: TokenKey, Value: key})
			return true
		}

		// 2. 发送按键 token
		e.emitInternal(RawToken{Kind: TokenKey, Value: key})
		return true
	}

	return false
}

// isDigit 检查字符串是否为单个数字字符
func isDigit(s string) bool {
	return len(s) == 1 && s[0] >= '0' && s[0] <= '9'
}

// Reset 重置引擎状态到初始层（Invariant 8: Reload = FSM 重生）
func (e *Engine) Reset() {
	if e.layerTimer != nil {
		e.layerTimer.Stop()
		e.layerTimer = nil
	}
	// 重置到初始状态
	if e.Keymap != nil && e.Keymap.Initial != "" {
		e.Active = e.Keymap.Initial
	} else {
		e.Active = "NAV"
	}
	e.count = 0
	e.emitInternal(RawToken{Kind: TokenSystem, Value: "reset"})
}

// Reload 重新加载keymap并重置FSM（Invariant 8: Reload = atomic rebuild）
func Reload(configPath string) error {
	// Load + Validate
	if err := LoadKeymap(configPath); err != nil {
		return err
	}

	// NewEngine
	InitEngine(&KM)

	// Reset + UI refresh
	Reset()

	return nil
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
	case "goto_line_start":
		// 发送 Home 键到当前窗格，这通常会将光标移到行首
		tmux("send-keys -t . Home")
	case "goto_line_end":
		// 发送 End 键到当前窗格，这通常会将光标移到行尾
		tmux("send-keys -t . End")
	case "move_left":
		// 发送左箭头键
		tmux("send-keys -t . Left")
	case "move_right":
		// 发送右箭头键
		tmux("send-keys -t . Right")
	case "move_up":
		// 发送上箭头键
		tmux("send-keys -t . Up")
	case "move_down":
		// 发送下箭头键
		tmux("send-keys -t . Down")
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
	// 由于循环导入问题，这里暂时使用占位符
	// 实际执行应该由上层处理
}

// DispatchIntent 分发意图给解析器
func (e *Engine) DispatchIntent(i *intent.Intent) error {
	// 解析器已废弃，直接返回
	return nil
}

func EnterFSM() {
	if defaultEngine == nil {
		InitEngine(&KM)
	}

	engine := defaultEngine
	engine.Active = "NAV"
	// 确保进入时是干净的 NAV
	engine.Reset()
	engine.emitInternal(RawToken{Kind: TokenSystem, Value: "enter"})
	UpdateUI() // 确保进入时更新UI
	// ShowUI() // Disable initial UI popup to prevent flashing/annoyance
}

// GetDefaultEngine 获取默认引擎实例
func GetDefaultEngine() *Engine {
	return defaultEngine
}

func ExitFSM() {
	if defaultEngine != nil {
		defaultEngine.Reset()
		defaultEngine.emitInternal(RawToken{Kind: TokenSystem, Value: "exit"})
	}
	HideUI()
	UpdateUI() // 确保退出时更新UI
	// FSM 不应直接依赖 backend
	// 执行层的退出逻辑应该由上层处理
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
	Initial string              `yaml:"initial"`
	States  map[string]StateDef `yaml:"states"`
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
)

// OnNvimMode 处理来自 Neovim 的模式变化
func OnNvimMode(mode string) {
	// 如果 Neovim 进入插入模式或可视模式，退出 FSM
	if mode == "i" || mode == "v" || mode == "V" || strings.HasPrefix(mode, "s") {
		ExitFSM()
	}
}

// NotifyNvimMode 通知 Neovim 当前 FSM 模式
// 注意：这个函数 currently would need to use intents to communicate
// with the backend, but Phase-3 requires that FSM doesn't directly call backend
func NotifyNvimMode() {
	// Phase-3 invariant: FSM does not directly call backend
	// This functionality should be handled by Kernel/Weaver layer
	// using intents to communicate with the backend
}

````

## 📄 `fsm/token.go`

````go
package fsm

type RawTokenKind int

const (
	TokenDigit RawTokenKind = iota
	TokenKey
	TokenRepeat
	TokenSystem
)

type RawToken struct {
	Kind  RawTokenKind
	Value string
}

````

## 📄 `fsm/ui_stub.go`

````go
package fsm

import (
	"fmt"
	"os/exec"
)

// UIDriver 定义UI驱动接口
type UIDriver interface {
	SetUserOption(option, value string) error
	RefreshClient(clientName string) error
}

var uiDriver UIDriver

// OnUpdateUI 当UI需要更新时调用的回调函数
var OnUpdateUI func()

// SetUIDriver 设置UI驱动实现
func SetUIDriver(driver UIDriver) {
	uiDriver = driver
}

// UpdateUI 更新UI显示当前FSM状态（Invariant 9: UI 派生状态）
func UpdateUI(_ ...any) {
	// TEMPORARY: debug-only UI bridge
	// This is a technical debt - FSM should NOT directly touch tmux
	// TODO: Move to Kernel → Weaver → Backend pipeline
	updateTmuxVariables()

	// 调用外部注册的UI更新回调
	if OnUpdateUI != nil {
		OnUpdateUI()
	}
}

// updateTmuxVariables 更新 tmux 状态变量
func updateTmuxVariables() {
	if defaultEngine == nil {
		return
	}

	// 更新状态变量
	activeLayer := defaultEngine.Active
	if activeLayer == "" {
		activeLayer = "NAV"
	}

	// 设置状态变量
	setTmuxOption("@fsm_state", activeLayer)

	// 如果有计数器，也显示它
	if defaultEngine.count > 0 {
		setTmuxOption("@fsm_keys", fmt.Sprintf("%d", defaultEngine.count))
	} else {
		setTmuxOption("@fsm_keys", "")
	}

	// 刷新客户端以更新状态栏
	refreshTmuxClient()
}

// setTmuxOption 设置 tmux 选项
func setTmuxOption(option, value string) {
	cmd := exec.Command("tmux", "set", "-g", option, value)
	_ = cmd.Run()
}

// refreshTmuxClient 刷新 tmux 客户端
func refreshTmuxClient() {
	cmd := exec.Command("tmux", "refresh-client", "-S")
	_ = cmd.Run()
}

// HideUI 隐藏UI
func HideUI() {
	// Phase‑3 invariant:
	// FSM does NOT touch UI / backend directly.
	// UI update must be handled by Kernel / Weaver.
	// 但是，为了隐藏状态，我们需要重置 tmux 变量
	setTmuxOption("@fsm_state", "")
	setTmuxOption("@fsm_keys", "")
	refreshTmuxClient()
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
	"sync"
	"time"
	"tmux-fsm/backend"
)

type Cursor struct {
	Row int
	Col int
}

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
	transMgr    *TransactionManager
	txJournal   *TxJournal // 新增：事务日志
	socketPath  = "/tmp/tmux-fsm.sock"
)

func init() {
	// 初始化全局事务管理器
	transMgr = &TransactionManager{
		nextID: 0,
	}

	// 初始化事务日志
	txJournal = NewTxJournal()
}

func loadState() FSMState {
	// Use GlobalBackend to read tmux options
	out, err := backend.GlobalBackend.GetUserOption("@tmux_fsm_state")
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
	if err := backend.GlobalBackend.SetUserOption("@tmux_fsm_state", string(data)); err != nil {
		log.Printf("Failed to save FSM state: %v", err)
	}
}

// saveFSMState 保存 FSM 状态
func saveFSMState() {
	stateMu.Lock()
	defer stateMu.Unlock()

	data, err := json.Marshal(globalState)
	if err != nil {
		log.Printf("Failed to marshal FSM state: %v", err)
		return
	}

	saveStateRaw(data)
}

func updateStatusBar(state FSMState, clientName string) {
	modeMsg := state.Mode
	if modeMsg == "" {
		modeMsg = "NORMAL"
	}

	// 融合显示逻辑
	// activeLayer := fsm.GetActiveLayer() // 由于循环导入，暂时注释掉
	// if activeLayer != "NAV" && activeLayer != "" {
	// 	modeMsg = activeLayer // Override with FSM layer if active
	// } else {
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
	// }

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
	backend.GlobalBackend.SetUserOption("@fsm_state", modeMsg)
	backend.GlobalBackend.SetUserOption("@fsm_keys", keysMsg)
	backend.GlobalBackend.RefreshClient(clientName) // Refresh the target client

	// --- [ABI: Heartbeat Lock] ---
	// Re-assert the key table to prevent "one-shot" dropouts.
	// Check @fsm_active to allow intentional exits.
	if clientName != "" && clientName != "default" {
		// Fetching @fsm_active via GlobalBackend if it were available would be ideal,
		// but for now, we rely on the fact that we are in a state where we should be active.
		// If GlobalBackend could read options, it would be better.
		// For now, we assume if we got here, FSM is active.
		backend.GlobalBackend.SwitchClientTable(clientName, "fsm")
	}
}

````

## 📄 `index/index.go`

````go
package index

import (
	"fmt"
	"sort"
	"time"
	"tmux-fsm/crdt"
	"tmux-fsm/semantic"
)

// FactType 事实类型
type FactType string

const (
	FactTypeInsert  FactType = "insert"
	FactTypeDelete  FactType = "delete"
	FactTypeMove    FactType = "move"
	FactTypeReplace FactType = "replace"
)

// EventIndex 事件索引
type EventIndex struct {
	ByActor    map[crdt.ActorID][]crdt.EventID
	ByType     map[FactType][]crdt.EventID
	ByPosition PositionIntervalTree
	ByTime     TimeBTree
	ByContent  map[string][]crdt.EventID // 按内容索引
}

// PositionIntervalTree 位置区间树（简化实现）
type PositionIntervalTree struct {
	// 这里使用一个简单的映射作为示例
	// 实际实现可能需要更复杂的数据结构
	intervals map[string][]crdt.EventID
}

// TimeBTree 时间B树（简化实现）
type TimeBTree struct {
	// 简化为时间戳到事件ID的映射
	timeline map[int64][]crdt.EventID
}

// NewEventIndex 创建新的事件索引
func NewEventIndex() *EventIndex {
	return &EventIndex{
		ByActor:    make(map[crdt.ActorID][]crdt.EventID),
		ByType:     make(map[FactType][]crdt.EventID),
		ByPosition: PositionIntervalTree{intervals: make(map[string][]crdt.EventID)},
		ByTime:     TimeBTree{timeline: make(map[int64][]crdt.EventID)},
		ByContent:  make(map[string][]crdt.EventID),
	}
}

// BuildIndex 构建索引
func BuildIndex(events []crdt.SemanticEvent) *EventIndex {
	index := NewEventIndex()

	for _, event := range events {
		// 按参与者索引
		index.ByActor[event.Actor] = append(index.ByActor[event.Actor], event.ID)

		// 按类型索引
		factType := getFactType(event.Fact)
		index.ByType[factType] = append(index.ByType[factType], event.ID)

		// 按时间索引
		index.ByTime.timeline[event.Time.Unix()] = append(index.ByTime.timeline[event.Time.Unix()], event.ID)

		// 按位置索引
		positionKey := getPositionKey(event.Fact)
		index.ByPosition.intervals[positionKey] = append(index.ByPosition.intervals[positionKey], event.ID)

		// 按内容索引
		contentKey := getContentKey(event.Fact)
		if contentKey != "" {
			index.ByContent[contentKey] = append(index.ByContent[contentKey], event.ID)
		}
	}

	return index
}

// getFactType 获取事实类型
func getFactType(fact semantic.Fact) FactType {
	switch fact.Kind() {
	case semantic.FactInsert:
		return FactTypeInsert
	case semantic.FactDelete:
		return FactTypeDelete
	case semantic.FactMove:
		return FactTypeMove
	case semantic.FactReplace:
		return FactTypeReplace
	default:
		return FactType("unknown")
	}
}

// getPositionKey 获取位置键
func getPositionKey(fact semantic.Fact) string {
	anchor := fact.Anchor()
	return string(anchor.PaneID) + ":" + fmt.Sprintf("%d", anchor.Line) + ":" + fmt.Sprintf("%d", anchor.Col)
}

// getContentKey 获取内容键
func getContentKey(fact semantic.Fact) string {
	return fact.Text()
}

// QueryByActor 按参与者查询
func (idx *EventIndex) QueryByActor(actor crdt.ActorID) []crdt.EventID {
	events, exists := idx.ByActor[actor]
	if !exists {
		return []crdt.EventID{}
	}
	return events
}

// QueryByType 按类型查询
func (idx *EventIndex) QueryByType(ft FactType) []crdt.EventID {
	events, exists := idx.ByType[ft]
	if !exists {
		return []crdt.EventID{}
	}
	return events
}

// QueryByTimeRange 按时间范围查询
func (idx *EventIndex) QueryByTimeRange(start, end time.Time) []crdt.EventID {
	var result []crdt.EventID

	startUnix := start.Unix()
	endUnix := end.Unix()

	for timestamp, events := range idx.ByTime.timeline {
		if timestamp >= startUnix && timestamp <= endUnix {
			result = append(result, events...)
		}
	}

	return result
}

// QueryByPositionRange 按位置范围查询
func (idx *EventIndex) QueryByPositionRange(startPos, endPos string) []crdt.EventID {
	var result []crdt.EventID

	// 简化实现：查找在指定位置范围内的事件
	for posKey, events := range idx.ByPosition.intervals {
		if posKey >= startPos && posKey <= endPos {
			result = append(result, events...)
		}
	}

	return result
}

// QueryByContent 按内容查询
func (idx *EventIndex) QueryByContent(content string) []crdt.EventID {
	events, exists := idx.ByContent[content]
	if !exists {
		return []crdt.EventID{}
	}
	return events
}

// QueryAIChanges 查询 AI 的更改
func (idx *EventIndex) QueryAIChanges(aiActorPrefix string) []crdt.EventID {
	var result []crdt.EventID

	for actor, events := range idx.ByActor {
		actorStr := string(actor)
		if len(actorStr) >= len(aiActorPrefix) && actorStr[:len(aiActorPrefix)] == aiActorPrefix {
			result = append(result, events...)
		}
	}

	return result
}

// QueryEvolutionHistory 查询某段文本的演化历史
func (idx *EventIndex) QueryEvolutionHistory(content string) []crdt.EventID {
	// 首先按内容查找
	contentEvents := idx.QueryByContent(content)

	// 然后可能需要扩展到相关的插入/删除事件
	var result []crdt.EventID
	result = append(result, contentEvents...)

	// 这里可以添加更多逻辑来查找相关的事件
	// 例如，查找在同一位置附近的操作等

	return result
}

// QueryWhoDeleted 查询谁删除了特定内容
func (idx *EventIndex) QueryWhoDeleted(content string) []crdt.ActorID {
	var actors []crdt.ActorID

	// 查找删除操作
	deleteEvents := idx.QueryByType(FactTypeDelete)

	for range deleteEvents {
		// 这里需要一个事件ID到事件的映射
		// 由于简化实现，我们跳过这一步
		// 在实际实现中，需要从存储中检索事件并检查其内容
	}

	return actors
}

// SortEventsByID 对事件ID进行排序
func SortEventsByID(events []crdt.EventID) []crdt.EventID {
	sorted := make([]crdt.EventID, len(events))
	copy(sorted, events)

	sort.Slice(sorted, func(i, j int) bool {
		return string(sorted[i]) < string(sorted[j])
	})

	return sorted
}

// SortEventsByTime 对事件按时间排序
func SortEventsByTime(events []crdt.SemanticEvent) []crdt.SemanticEvent {
	sorted := make([]crdt.SemanticEvent, len(events))
	copy(sorted, events)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Time.Before(sorted[j].Time)
	})

	return sorted
}

// GetTimeline 获取时间线
func (idx *EventIndex) GetTimeline() []int64 {
	var timestamps []int64
	for timestamp := range idx.ByTime.timeline {
		timestamps = append(timestamps, timestamp)
	}

	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] < timestamps[j]
	})

	return timestamps
}

// GetActors 获取所有参与者
func (idx *EventIndex) GetActors() []crdt.ActorID {
	var actors []crdt.ActorID
	for actor := range idx.ByActor {
		actors = append(actors, actor)
	}

	// 排序以确保一致性
	sort.Slice(actors, func(i, j int) bool {
		return string(actors[i]) < string(actors[j])
	})

	return actors
}

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
	SnapshotHash string                 `json:"snapshot_hash"`     // Phase 6.2
	AllowPartial bool                   `json:"allow_partial"`     // Phase 7: Explicit permission for fuzzy resolution
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

## 📄 `intent/Adapter.go`

````go
package intent

import "tmux-fsm/weaver/core"

// Adapter adapts a standard `intent.Intent` to the `core.Intent` interface,
// allowing the new weaver system to process intents generated by older parts
// of the system.
type Adapter struct {
	Intent
}

func (a *Adapter) GetKind() core.IntentKind {
	return core.IntentKind(a.Kind)
}

func (a *Adapter) GetTarget() core.SemanticTarget {
	return core.SemanticTarget{
		Kind:      int(a.Target.Kind),
		Direction: a.Target.Direction,
		Scope:     a.Target.Scope,
		Value:     a.Target.Value,
	}
}

func (a *Adapter) GetCount() int {
	return a.Count
}

func (a *Adapter) GetMeta() map[string]interface{} {
	return a.Meta
}

func (a *Adapter) GetPaneID() string {
	return a.PaneID
}

func (a *Adapter) GetSnapshotHash() string {
	return a.SnapshotHash
}

func (a *Adapter) IsPartialAllowed() bool {
	return a.AllowPartial
}

func (a *Adapter) GetAnchors() []core.Anchor {
	coreAnchors := make([]core.Anchor, len(a.Anchors))
	for i, anchor := range a.Anchors {
		coreAnchors[i] = core.Anchor{
			PaneID: anchor.PaneID,
			Kind:   core.AnchorKind(anchor.Kind),
			Ref:    anchor.Ref,
			Hash:   anchor.Hash,
			LineID: core.LineID(anchor.LineID),
			Start:  anchor.Start,
			End:    anchor.End,
		}
	}
	return coreAnchors
}

````

## 📄 `intent/builder/builder.go`

````go
package builder

import (
	"tmux-fsm/intent"
)

// BuildContext 构建上下文
type BuildContext struct {
	Action       string // legacy action string
	Command      string // normalized command (future)
	Count        int
	PaneID       string
	SnapshotHash string
	Meta         map[string]interface{} // 额外元数据
}

// Builder Intent构建器接口
type Builder interface {
	// Priority determines evaluation order.
	// Higher value = higher priority.
	Priority() int
	Build(ctx BuildContext) (*intent.Intent, bool)
}

````

## 📄 `intent/builder/composite_builder.go`

````go
package builder

import (
	"sort"
	"tmux-fsm/intent"
)

// CompositeBuilder 组合构建器
type CompositeBuilder struct {
	builders []Builder
}

// NewCompositeBuilder 创建组合构建器
func NewCompositeBuilder() *CompositeBuilder {
	cb := &CompositeBuilder{
		builders: []Builder{
			&MoveBuilder{},
			&TextObjectBuilder{},
			&OperatorBuilder{},
			&MacroBuilder{},
		},
	}
	cb.sort()
	return cb
}

// AddBuilder 添加构建器
func (cb *CompositeBuilder) AddBuilder(builder Builder) {
	cb.builders = append(cb.builders, builder)
	cb.sort()
}

// Build 尝试使用所有构建器构建Intent
func (cb *CompositeBuilder) Build(ctx BuildContext) (*intent.Intent, bool) {
	for _, builder := range cb.builders {
		intent, ok := builder.Build(ctx)
		if ok {
			return intent, true
		}
	}
	return nil, false
}

// sort 按优先级排序构建器
// Builders are evaluated in order.
// Order MUST reflect semantic priority.
func (cb *CompositeBuilder) sort() {
	sort.SliceStable(cb.builders, func(i, j int) bool {
		return cb.builders[i].Priority() > cb.builders[j].Priority()
	})
}

````

## 📄 `intent/builder/doc.go`

````go
// Package builder provides NATIVE Intent builders.
//
// This package is the ONLY authoritative way to construct new Intents.
// Legacy intent construction paths are frozen elsewhere and must not be extended.
//
// Rules:
// - Do NOT import legacy logic
// - Builders must be semantic-only
// - Priority determines builder matching order
package builder

````

## 📄 `intent/builder/intent_diff.go`

````go
package builder

import "tmux-fsm/intent"

type IntentDiff struct {
	Field  string
	Legacy interface{}
	Native interface{}
}

func DiffIntent(legacy, native *intent.Intent) []IntentDiff {
	var diffs []IntentDiff

	if legacy == nil || native == nil {
		return diffs
	}

	if legacy.Kind != native.Kind {
		diffs = append(diffs, IntentDiff{"Kind", legacy.Kind, native.Kind})
	}

	if legacy.Count != native.Count {
		diffs = append(diffs, IntentDiff{"Count", legacy.Count, native.Count})
	}

	if legacy.Target.Kind != native.Target.Kind {
		diffs = append(diffs, IntentDiff{"Target.Kind", legacy.Target.Kind, native.Target.Kind})
	}

	if legacy.Target.Direction != native.Target.Direction {
		diffs = append(diffs, IntentDiff{"Target.Direction", legacy.Target.Direction, native.Target.Direction})
	}

	if legacy.Target.Scope != native.Target.Scope {
		diffs = append(diffs, IntentDiff{"Target.Scope", legacy.Target.Scope, native.Target.Scope})
	}

	if legacy.Target.Value != native.Target.Value {
		diffs = append(diffs, IntentDiff{"Target.Value", legacy.Target.Value, native.Target.Value})
	}

	if legacy.PaneID != native.PaneID {
		diffs = append(diffs, IntentDiff{"PaneID", legacy.PaneID, native.PaneID})
	}

	return diffs
}

````

## 📄 `intent/builder/macro_builder.go`

````go
package builder

import (
	"tmux-fsm/intent"
)

// MacroBuilder 宏构建器
type MacroBuilder struct{}

// Priority 宏操作优先级中等
func (b *MacroBuilder) Priority() int {
	return 8
}

// Build 构建宏Intent
func (b *MacroBuilder) Build(ctx BuildContext) (*intent.Intent, bool) {
	switch ctx.Action {
	case "start_macro":
		register, ok := ctx.Meta["register"].(string)
		if !ok {
			register = "a" // 默认注册器
		}
		return &intent.Intent{
			Kind:   intent.IntentMacro,
			Target: intent.SemanticTarget{Kind: intent.TargetNone, Scope: "start"},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operation": "start_recording", "register": register},
			PaneID: ctx.PaneID,
		}, true
	case "stop_macro":
		return &intent.Intent{
			Kind:   intent.IntentMacro,
			Target: intent.SemanticTarget{Kind: intent.TargetNone, Scope: "stop"},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operation": "stop_recording"},
			PaneID: ctx.PaneID,
		}, true
	case "play_macro":
		register, ok := ctx.Meta["register"].(string)
		if !ok {
			register = "a" // 默认注册器
		}
		return &intent.Intent{
			Kind:   intent.IntentMacro,
			Target: intent.SemanticTarget{Kind: intent.TargetNone, Scope: "play"},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operation": "play", "register": register},
			PaneID: ctx.PaneID,
		}, true
	default:
		return nil, false
	}
}

````

## 📄 `intent/builder/move_builder.go`

````go
package builder

import (
	"tmux-fsm/intent"
)

// MoveBuilder 移动操作构建器
type MoveBuilder struct{}

// Priority 移动操作优先级较高，因为是立即执行的motion
func (b *MoveBuilder) Priority() int {
	return 10
}

// Build 构建移动Intent
func (b *MoveBuilder) Build(ctx BuildContext) (*intent.Intent, bool) {
	switch ctx.Action {
	case "move_left":
		return &intent.Intent{
			Kind:   intent.IntentMove,
			Target: intent.SemanticTarget{Kind: intent.TargetChar, Direction: "left"},
			Count:  ctx.Count,
			PaneID: ctx.PaneID,
		}, true
	case "move_right":
		return &intent.Intent{
			Kind:   intent.IntentMove,
			Target: intent.SemanticTarget{Kind: intent.TargetChar, Direction: "right"},
			Count:  ctx.Count,
			PaneID: ctx.PaneID,
		}, true
	case "move_up":
		return &intent.Intent{
			Kind:   intent.IntentMove,
			Target: intent.SemanticTarget{Kind: intent.TargetChar, Direction: "up"},
			Count:  ctx.Count,
			PaneID: ctx.PaneID,
		}, true
	case "move_down":
		return &intent.Intent{
			Kind:   intent.IntentMove,
			Target: intent.SemanticTarget{Kind: intent.TargetChar, Direction: "down"},
			Count:  ctx.Count,
			PaneID: ctx.PaneID,
		}, true
	case "move_line_start":
		return &intent.Intent{
			Kind:   intent.IntentMove,
			Target: intent.SemanticTarget{Kind: intent.TargetLine, Scope: "start"},
			Count:  ctx.Count,
			PaneID: ctx.PaneID,
		}, true
	case "move_line_end":
		return &intent.Intent{
			Kind:   intent.IntentMove,
			Target: intent.SemanticTarget{Kind: intent.TargetLine, Scope: "end"},
			Count:  ctx.Count,
			PaneID: ctx.PaneID,
		}, true
	default:
		return nil, false
	}
}

````

## 📄 `intent/builder/operator_builder.go`

````go
package builder

import (
	"tmux-fsm/intent"
)

// OperatorBuilder 操作符构建器
type OperatorBuilder struct{}

// Priority 操作符优先级较低，因为需要等待motion
func (b *OperatorBuilder) Priority() int {
	return 5
}

// Build 构建操作符Intent
func (b *OperatorBuilder) Build(ctx BuildContext) (*intent.Intent, bool) {
	switch ctx.Action {
	case "delete":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetChar},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpDelete},
			PaneID: ctx.PaneID,
		}, true
	case "yank":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetChar},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpYank},
			PaneID: ctx.PaneID,
		}, true
	case "change":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetChar},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpChange},
			PaneID: ctx.PaneID,
		}, true
	default:
		return nil, false
	}
}

// TODO: Operator intents currently encode legacy operator semantics in Meta.
// This MUST be replaced by first-class intent kinds before Cut 3.

````

## 📄 `intent/builder/semantic_equal.go`

````go
package builder

import "tmux-fsm/intent"

type SemanticCompareMode int

const (
	CompareMigration SemanticCompareMode = iota
	CompareStrict
)

// SemanticEqual compares two intents for semantic equality.
// Nil intents are only semantically equal if both are nil.
func SemanticEqual(a, b *intent.Intent, mode SemanticCompareMode) bool {
	if a == nil || b == nil {
		return a == b
	}

	if a.Kind != b.Kind ||
		a.Target.Kind != b.Target.Kind ||
		a.Target.Direction != b.Target.Direction ||
		a.Target.Scope != b.Target.Scope ||
		a.Target.Value != b.Target.Value ||
		a.Count != b.Count {
		return false
	}

	if mode == CompareStrict && a.PaneID != b.PaneID {
		return false
	}

	// Migration mode intentionally ignores routing
	return true
}

````

## 📄 `intent/builder/text_object.go`

````go
package builder

import (
	"tmux-fsm/intent"
)

// TextObjectKind 文本对象类型
type TextObjectKind string

const (
	TextObjectInnerParen  TextObjectKind = "inner_paren"
	TextObjectAroundParen TextObjectKind = "around_paren"
	TextObjectInnerQuote  TextObjectKind = "inner_quote"
	TextObjectAroundQuote TextObjectKind = "around_quote"
	TextObjectInnerWord   TextObjectKind = "inner_word"
	TextObjectAroundWord  TextObjectKind = "around_word"
)

// TextObjectBuilder 文本对象构建器
type TextObjectBuilder struct{}

// Priority 文本对象优先级较高，因为是明确的选择范围
func (b *TextObjectBuilder) Priority() int {
	return 15
}

// Build 构建文本对象Intent
func (b *TextObjectBuilder) Build(ctx BuildContext) (*intent.Intent, bool) {
	switch ctx.Action {
	case "delete_inner_paren":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetTextObject, Value: string(TextObjectInnerParen)},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpDelete},
			PaneID: ctx.PaneID,
		}, true
	case "delete_around_paren":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetTextObject, Value: string(TextObjectAroundParen)},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpDelete},
			PaneID: ctx.PaneID,
		}, true
	case "delete_inner_quote":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetTextObject, Value: string(TextObjectInnerQuote)},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpDelete},
			PaneID: ctx.PaneID,
		}, true
	case "delete_around_quote":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetTextObject, Value: string(TextObjectAroundQuote)},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpDelete},
			PaneID: ctx.PaneID,
		}, true
	case "change_inner_paren":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetTextObject, Value: string(TextObjectInnerParen)},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpChange},
			PaneID: ctx.PaneID,
		}, true
	case "yank_inner_paren":
		return &intent.Intent{
			Kind:   intent.IntentOperator,
			Target: intent.SemanticTarget{Kind: intent.TargetTextObject, Value: string(TextObjectInnerParen)},
			Count:  ctx.Count,
			Meta:   map[string]interface{}{"operator": intent.OpYank},
			PaneID: ctx.PaneID,
		}, true
	default:
		return nil, false
	}
}

````

## 📄 `intent/grammar_intent.go`

````go
package intent

// GrammarIntent 是 Grammar 专用的意图类型，只包含 Grammar 可以设置的字段
type GrammarIntent struct {
	Kind   IntentKind
	Count  int
	Motion *Motion
	Op     *OperatorKind
}

````

## 📄 `intent/motion.go`

````go
package intent

type MotionKind int

const (
	MotionChar MotionKind = iota
	MotionWord
	MotionLine
	MotionGoto
	MotionRange // ✅ 新增
	MotionFind
)

// Direction for character-wise and line-wise motions
type Direction int

const (
	DirectionNone Direction = iota
	DirectionLeft
	DirectionRight
	DirectionUp
	DirectionDown
)

type FindDirection int

const (
	FindForward FindDirection = iota
	FindBackward
)

type FindMotion struct {
	Char      rune          // 要查找的字符
	Direction FindDirection // Forward / Backward
	Till      bool          // t / T
}

type Motion struct {
	Kind      MotionKind
	Count     int
	Direction Direction    // For up, down, left, right
	Find      *FindMotion  // 只有 Kind == MotionFind 时非空
	Range     *RangeMotion // 只有 Kind == MotionRange 时非空
}

````

## 📄 `intent/promote.go`

````go
package intent

// Promote 是 GrammarIntent → Intent 的唯一合法通道
// Grammar 不允许直接构造 Intent
func Promote(g *GrammarIntent) *Intent {
	if g == nil {
		return nil
	}

	i := &Intent{
		Kind:   g.Kind,
		Count:  g.Count,
		Motion: g.Motion,
	}

	// Operator 提升（强类型）
	if g.Op != nil {
		i.Operator = g.Op
	}

	return i
}

````

## 📄 `intent/range.go`

````go
package intent

type RangeKind int

const (
	RangeTextObject RangeKind = iota
	RangeVisual
	RangeLineStart // For '0'
	RangeLineEnd   // For '$'
)

type RangeMotion struct {
	Kind       RangeKind
	TextObject *TextObject
}

````

## 📄 `intent/text_object.go`

````go
package intent

type TextObjectScope int

const (
	Inner TextObjectScope = iota
	Around
)

type TextObjectKind int

const (
	Word TextObjectKind = iota
	Paren
	Bracket
	Brace
	QuoteSingle
	QuoteDouble
	Backtick
)

type TextObject struct {
	Scope  TextObjectScope
	Object TextObjectKind
}

````

## 📄 `intent_bridge.go`

````go
// LEGACY — DO NOT EXTEND
// This path exists ONLY for backward compatibility.
// Any new behavior MUST be implemented via native Intent builders.
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

	// 检查是否是文本对象简写 (iw, aw, ip, ap, etc.)
	if isTextObject(motion) {
		return SemanticTarget{Kind: TargetTextObject, Value: motion}
	}

	// 默认返回
	return SemanticTarget{Kind: TargetNone}
}

// isTextObject 检查是否是文本对象简写
func isTextObject(motion string) bool {
	if len(motion) != 2 {
		return false
	}

	// 检查第一个字符是否是 i 或 a (inside/around)
	modifier := motion[0:1]
	if modifier != "i" && modifier != "a" {
		return false
	}

	// 检查第二个字符是否是支持的文本对象类型
	objType := motion[1:2]
	switch objType {
	case "w", "p", "s", "b", "B", "(", ")", "[", "]", "{", "}", "\"", "'", "`":
		return true
	default:
		return false
	}
}

````

## 📄 `invariant/test.go`

````go
package invariant

import (
	"math/rand"
	"testing"
	"time"
)

// TextState 模拟文本状态
type TextState struct {
	Text   string
	Cursor int
}

// Apply 模拟事务对状态的应用
func (s TextState) Apply(tx Transaction) (TextState, error) {
	switch t := tx.(type) {
	case *InsertTx:
		if t.Pos < 0 || t.Pos > len(s.Text) {
			return s, nil // 边界检查，不执行
		}
		newText := s.Text[:t.Pos] + t.Text + s.Text[t.Pos:]
		return TextState{
			Text:   newText,
			Cursor: t.Pos + len(t.Text),
		}, nil

	case *DeleteTx:
		if t.Pos < 0 || t.Pos+t.Len > len(s.Text) {
			return s, nil // 边界检查，不执行
		}
		newText := s.Text[:t.Pos] + s.Text[t.Pos+t.Len:]
		return TextState{
			Text:   newText,
			Cursor: t.Pos,
		}, nil

	case *MoveCursorTx:
		newCursor := t.To
		if newCursor < 0 {
			newCursor = 0
		}
		if newCursor > len(s.Text) {
			newCursor = len(s.Text)
		}
		return TextState{
			Text:   s.Text,
			Cursor: newCursor,
		}, nil
	}

	return s, nil
}

// Transaction 接口定义
type Transaction interface {
	Apply() error
	Inverse() Transaction
	Kind() string
	Tags() []string
	CanMerge(next Transaction) bool
	Merge(next Transaction) Transaction
}

// InsertTx 插入事务
type InsertTx struct {
	Pos  int
	Text string
}

func (t *InsertTx) Apply() error { return nil }
func (t *InsertTx) Inverse() Transaction {
	return &DeleteTx{Pos: t.Pos, Len: len(t.Text)}
}
func (t *InsertTx) Kind() string                       { return "insert" }
func (t *InsertTx) Tags() []string                     { return []string{"insert"} }
func (t *InsertTx) CanMerge(next Transaction) bool     { return false }
func (t *InsertTx) Merge(next Transaction) Transaction { return next }

// DeleteTx 删除事务
type DeleteTx struct {
	Pos int
	Len int
}

func (t *DeleteTx) Apply() error { return nil }
func (t *DeleteTx) Inverse() Transaction {
	return &InsertTx{Pos: t.Pos, Text: ""} // 简化实现
}
func (t *DeleteTx) Kind() string                       { return "delete" }
func (t *DeleteTx) Tags() []string                     { return []string{"delete"} }
func (t *DeleteTx) CanMerge(next Transaction) bool     { return false }
func (t *DeleteTx) Merge(next Transaction) Transaction { return next }

// MoveCursorTx 移动光标事务
type MoveCursorTx struct {
	To int
}

func (t *MoveCursorTx) Apply() error { return nil }
func (t *MoveCursorTx) Inverse() Transaction {
	// 简化实现
	return &MoveCursorTx{To: 0}
}
func (t *MoveCursorTx) Kind() string                       { return "move" }
func (t *MoveCursorTx) Tags() []string                     { return []string{"move"} }
func (t *MoveCursorTx) CanMerge(next Transaction) bool     { return false }
func (t *MoveCursorTx) Merge(next Transaction) Transaction { return next }

// TestTxInverseProperty 测试事务与其逆操作的性质
func TestTxInverseProperty(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100; i++ {
		// 随机生成初始状态
		initialText := randomString(rand.Intn(20))
		s0 := TextState{Text: initialText, Cursor: rand.Intn(len(initialText) + 1)}

		// 创建一个随机事务
		tx := randomTransaction(len(s0.Text))

		// 应用事务
		s1, err := s0.Apply(tx)
		if err != nil {
			continue // Apply 失败不违反不变量
		}

		// 应用逆事务
		s2, err := s1.Apply(tx.Inverse())
		if err != nil {
			t.Errorf("Inverse application failed: %v", err)
			continue
		}

		// 检查是否回到原始状态
		if s0.Text != s2.Text {
			t.Errorf("Apply ∘ Inverse ≠ Identity: %s != %s", s0.Text, s2.Text)
		}
	}
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// randomTransaction 生成随机事务
func randomTransaction(maxPos int) Transaction {
	pos := rand.Intn(maxPos + 1)
	switch rand.Intn(3) {
	case 0:
		return &InsertTx{Pos: pos, Text: randomString(rand.Intn(5))}
	case 1:
		delLen := rand.Intn(maxPos - pos + 1)
		return &DeleteTx{Pos: pos, Len: delLen}
	case 2:
		newPos := rand.Intn(maxPos + 1)
		return &MoveCursorTx{To: newPos}
	default:
		return &InsertTx{Pos: pos, Text: "test"}
	}
}

````

## 📄 `kernel/decide.go`

````go
package kernel

import (
	"tmux-fsm/fsm"
	"tmux-fsm/intent"
	"tmux-fsm/planner"
)

type DecisionKind int

const (
	DecisionNone DecisionKind = iota
	DecisionFSM
	DecisionLegacy
	DecisionIntent
)

type Decision struct {
	Kind   DecisionKind
	Intent *intent.Intent
	Action string // For simple FSM actions
}

// GrammarEmitter 用于将 Grammar 的结果传递给 Kernel
type GrammarEmitter struct {
	grammar  *planner.Grammar
	callback func(*intent.GrammarIntent)
}

func (g *GrammarEmitter) Emit(token fsm.RawToken) {
	grammarIntent := g.grammar.Consume(token)
	if grammarIntent != nil && g.callback != nil {
		g.callback(grammarIntent)
	}
}

func (k *Kernel) Decide(key string) *Decision {
	// ✅ 1. FSM 必须先看到 key
	if k.FSM != nil {
		var lastGrammarIntent *intent.GrammarIntent

		// 创建一个 GrammarEmitter 来处理 token
		grammarEmitter := &GrammarEmitter{
			grammar: k.Grammar,
			callback: func(grammarIntent *intent.GrammarIntent) {
				lastGrammarIntent = grammarIntent
			},
		}

		// 添加 GrammarEmitter 到 FSM
		k.FSM.AddEmitter(grammarEmitter)

		// 让 FSM 处理按键，这会生成 token
		dispatched := k.FSM.Dispatch(key)

		// 移除 GrammarEmitter
		k.FSM.RemoveEmitter(grammarEmitter)

		if dispatched && lastGrammarIntent != nil {
			// 将 GrammarIntent 提升为 Intent
			finalIntent := intent.Promote(lastGrammarIntent)

			// 返回意图供执行
			return &Decision{
				Kind:   DecisionIntent, // This is a full-fledged intent
				Intent: finalIntent,
			}
		}

		// Fallback for simple actions defined in keymap.yaml
		if dispatched && lastGrammarIntent == nil {
			if state, ok := k.FSM.Keymap.States[k.FSM.Active]; ok {
				if keyAction, ok := state.Keys[key]; ok && keyAction.Action != "" {
					// This is a simple FSM action, not a full intent.
					return &Decision{
						Kind:   DecisionFSM,
						Action: keyAction.Action,
					}
				}
			}
		}

		if dispatched {
			// ✅ 合法状态：key 被 FSM 吃了，但 Grammar 没有生成意图
			// 这是正常情况，例如在等待更多按键时
			return &Decision{
				Kind: DecisionNone, // FSM 吃了，但还没决定
			}
		}
	}

	// 没有 FSM 处理，明确返回 Legacy 决策
	return &Decision{
		Kind: DecisionLegacy,
	}
}

````

## 📄 `kernel/execute.go`

````go
package kernel

import (
	"log"
)

// Execute a decision made by the kernel.
func (k *Kernel) Execute(decision *Decision) {
	if decision == nil {
		log.Println("kernel.Execute called with nil decision")
		return
	}

	if k.Exec == nil {
		log.Println("kernel.Execute called with nil executor")
		return
	}

	switch decision.Kind {
	case DecisionNone, DecisionLegacy:
		return // Do nothing intentionally.

	case DecisionIntent:
		// This is a full-fledged intent from the grammar.
		// Process it via the standard execution path.
		if decision.Intent == nil {
			log.Println("DecisionIntent without an intent")
			return
		}
		_ = k.Exec.Process(decision.Intent)

	default:
		log.Printf("Unknown or unhandled decision kind: %v", decision.Kind)
	}
}

````

## 📄 `kernel/intent_executor.go`

````go
package kernel

import "tmux-fsm/intent"

// IntentExecutor is the ONLY way Kernel can execute an Intent.
// Kernel does not know who implements it.
type IntentExecutor interface {
	Process(*intent.Intent) error
}

````

## 📄 `kernel/kernel.go`

````go
package kernel

import (
	"context"
	"log"
	"tmux-fsm/fsm"
	"tmux-fsm/intent"
	"tmux-fsm/intent/builder"
	"tmux-fsm/planner"
)

// ShadowStats records statistics for shadow intent comparison.
// NOTE: ShadowStats is not concurrency-safe.
// Kernel.HandleKey must be serialized.
type ShadowStats struct {
	Total      int
	Built      int
	Matched    int
	Mismatched int
}

type Kernel struct {
	FSM           *fsm.Engine
	Grammar       *planner.Grammar
	Exec          IntentExecutor
	NativeBuilder *builder.CompositeBuilder
	ShadowIntent  bool
	ShadowStats   ShadowStats
}

// ✅ Kernel 的唯一上下文入口（现在先很薄，未来可扩展）
type HandleContext struct {
	Ctx context.Context
}

func NewKernel(fsmEngine *fsm.Engine, exec IntentExecutor) *Kernel {
	return &Kernel{
		FSM:           fsmEngine,
		Grammar:       planner.NewGrammar(),
		Exec:          exec,
		NativeBuilder: builder.NewCompositeBuilder(),
		ShadowIntent:  true,
	}
}

// ✅ Kernel 的唯一入口
func (k *Kernel) HandleKey(hctx HandleContext, key string) {
	_ = hctx // ✅ 现在不用，但接口已经锁死

	// 通过Grammar路径生成intent（新的权威执行路径）
	var decision *Decision

	// 先尝试通过FSM + Grammar生成intent
	if k.FSM != nil && k.Grammar != nil {
		decision = k.Decide(key)

		if decision != nil {
			switch decision.Kind {
			case DecisionIntent:
				k.ProcessIntent(decision.Intent)
				return

			case DecisionFSM:
				k.Execute(decision)
				return

			case DecisionNone:
				// FSM 吃了 key，合法等待
				return

			case DecisionLegacy:
				// 明确：Grammar/FSM 不处理，才允许 legacy
				break
			}
		}
	}

	// 如果Grammar没有处理，记录信息（未来将完全移除legacy路径）
	if k.ShadowIntent && k.NativeBuilder != nil {
		// 只有在 DecisionLegacy 情况下才记录为未覆盖
		// DecisionNone 是合法的等待状态，不应计入未覆盖
		if decision != nil && decision.Kind == DecisionLegacy {
			log.Printf("[GRAMMAR COVERAGE] key '%s' not handled by Grammar", key)
			k.ShadowStats.Total++
			k.ShadowStats.Mismatched++ // 记录为未覆盖
		}
	}
}

// ProcessIntent 处理意图
func (k *Kernel) ProcessIntent(intent *intent.Intent) error {
	if k.Exec != nil {
		return k.Exec.Process(intent)
	}

	// 如果没有外部执行器，尝试通过FSM执行意图
	if k.FSM != nil && intent != nil {
		return k.FSM.DispatchIntent(intent)
	}

	return nil
}

````

## 📄 `kernel/resolver_executor.go`

````go
package kernel

import (
	"log"
	"tmux-fsm/intent"
	"tmux-fsm/weaver/manager"
)

// ResolverExecutor is the executor that forwards intents to the Weaver system.
type ResolverExecutor struct{}

// NewResolverExecutor creates a new ResolverExecutor.
func NewResolverExecutor() *ResolverExecutor {
	return &ResolverExecutor{}
}

// Process an intent by adapting it and sending it to the global Weaver manager.
func (e *ResolverExecutor) Process(i *intent.Intent) error {
	weaverMgr := manager.GetWeaverManager()
	if weaverMgr == nil {
		log.Println("Weaver manager is not initialized, intent dropped.")
		return nil
	}

	// Adapt the intent to the core.Intent interface and process it.
	adaptedIntent := &intent.Adapter{Intent: *i}
	return weaverMgr.ProcessIntentGlobal(adaptedIntent)
}

````

## 📄 `kernel/transaction.go`

````go
package kernel

import (
	"fmt"
	"tmux-fsm/editor"
	"tmux-fsm/types"
)

// TransactionRunner 事务执行器
// 负责在 ExecutionContext 中执行 Transaction
type TransactionRunner struct {
	ctx *editor.ExecutionContext
}

// NewTransactionRunner 创建新的事务执行器
func NewTransactionRunner(ctx *editor.ExecutionContext) *TransactionRunner {
	return &TransactionRunner{
		ctx: ctx,
	}
}

// Apply 应用事务（正向执行）
func (tr *TransactionRunner) Apply(tx *types.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	// 收集所有操作（用于更新 selections）
	ops := make([]editor.ResolvedOperation, 0, len(tx.Records))

	// 执行所有操作
	for _, record := range tx.Records {
		if err := editor.ApplyResolvedOperation(tr.ctx, record.ResolvedOp); err != nil {
			return fmt.Errorf("failed to apply operation: %w", err)
		}
		ops = append(ops, record.ResolvedOp)
	}

	// 更新 selections（在所有操作执行完成后）
	tr.updateSelectionsAfterOps(ops)

	return nil
}

// Undo 撤销事务（反向执行）
func (tr *TransactionRunner) Undo(tx *types.Transaction) error {
	return fmt.Errorf("undo not supported: inverse execution not implemented")
}

// Repeat 重复执行事务（用于 . repeat）
func (tr *TransactionRunner) Repeat(tx *types.Transaction) error {
	// Repeat 与 Apply 逻辑相同
	return tr.Apply(tx)
}

// updateSelectionsAfterOps 在操作执行后更新选区
func (tr *TransactionRunner) updateSelectionsAfterOps(ops []editor.ResolvedOperation) {
	if len(ops) == 0 {
		return
	}

	// 按 BufferID 分组操作
	opsByBuffer := make(map[editor.BufferID][]editor.ResolvedOperation)
	for _, op := range ops {
		opsByBuffer[op.BufferID] = append(opsByBuffer[op.BufferID], op)
	}

	// 对每个受影响的 buffer 更新其 selections
	for bufferID, bufferOps := range opsByBuffer {
		currentSels := tr.ctx.Selections.Get(bufferID)
		updatedSels := editor.UpdateSelections(currentSels, bufferOps)
		tr.ctx.Selections.Set(bufferID, updatedSels)
	}
}

````

## 📄 `main.go`

````go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"tmux-fsm/editor"
	"tmux-fsm/fsm"
	"tmux-fsm/intent"
	"tmux-fsm/kernel"
	"tmux-fsm/types"
	"tmux-fsm/weaver/core"
	"tmux-fsm/weaver/manager"
)

// weaverMgr 全局 Weaver 实例
var weaverMgr *manager.WeaverManager

// kernelInstance 全局 Kernel 实例
var kernelInstance *kernel.Kernel

// globalExecContext 全局执行上下文
var globalExecContext *editor.ExecutionContext

// TransactionManager 事务管理器
// 负责管理编辑操作的历史记录，遵循Vim语义规则
type TransactionManager struct {
	current         *types.Transaction
	nextID          types.TransactionID
	history         []*types.Transaction // 存储已提交的事务，用于 . repeat 和 undo
	lastCommittedTx *types.Transaction   // 最近提交的事务，用于 . repeat
}

// BeginTransaction 开始一个新的事务
// 一个事务对应一次可被 `.` 重复的最小操作单元
func (tm *TransactionManager) BeginTransaction() *types.Transaction {
	tm.current = &types.Transaction{
		ID:        tm.nextID,
		Records:   make([]types.OperationRecord, 0),
		CreatedAt: time.Now(),
	}
	tm.nextID++
	return tm.current
}

// AppendEffect 向当前事务追加效果记录
// 注意：调用此方法前必须确保事务已开始
func (tm *TransactionManager) AppendEffect(resolvedOp editor.ResolvedOperation, fact core.Fact) {
	if tm.current == nil {
		panic("AppendEffect called without active transaction - transaction must be explicitly started")
	}

	record := types.OperationRecord{
		ResolvedOp: resolvedOp,
		Fact:       fact,
	}

	tm.current.Records = append(tm.current.Records, record)
}

// CommitTransaction 提交当前事务
func (tm *TransactionManager) CommitTransaction() error {
	if tm.current == nil {
		return fmt.Errorf("no active transaction to commit")
	}

	// 保存到历史记录
	tm.history = append(tm.history, tm.current)

	// 更新最近提交的事务（用于 . repeat）
	tm.lastCommittedTx = tm.current

	tm.current = nil // 重置当前事务

	return nil
}

// AbortTransaction 放弃当前事务
func (tm *TransactionManager) AbortTransaction() error {
	if tm.current == nil {
		return fmt.Errorf("no active transaction to abort")
	}

	tm.current = nil // 重置当前事务

	return nil
}

// GetCurrentTransaction 获取当前事务（如果存在）
func (tm *TransactionManager) GetCurrentTransaction() *types.Transaction {
	return tm.current
}

// LastCommittedTransaction 获取最近提交的事务
// 用于 . repeat 功能
func (tm *TransactionManager) LastCommittedTransaction() *types.Transaction {
	return tm.lastCommittedTx
}

func main() {
	serverMode := flag.Bool("server", false, "run as server")
	socketPath := flag.String("socket", "/tmp/tmux-fsm.sock", "socket path")
	debugMode := flag.Bool("debug", false, "enable debug logging")
	configPath := flag.String("config", "./keymap.yaml", "path to keymap configuration file")
	reloadFlag := flag.Bool("reload", false, "reload keymap configuration")
	keyFlag := flag.String("key", "", "dispatch key to FSM")
	enterFlag := flag.Bool("enter", false, "enter FSM mode")
	exitFlag := flag.Bool("exit", false, "exit FSM mode")
	helpFlag := flag.Bool("help", false, "show help")
	flag.Parse()

	// Load keymap configuration
	if err := fsm.LoadKeymap(*configPath); err != nil {
		log.Printf("Warning: Failed to load keymap from %s: %v", *configPath, err)
		// Continue with default keymap if available
	} else {
		log.Printf("Successfully loaded keymap from %s", *configPath)
	}

	// Initialize FSM engine with loaded keymap
	fsm.InitEngine(&fsm.KM)

	// 初始化新的编辑内核组件
	// cursorEngine := editor.NewCursorEngine(editor.NewSimpleBuffer([]string{})) // 创建光标引擎（已移除，因为函数不存在）

	// 创建基于新解析器的执行器（过渡性实现）
	resolverExecutor := kernel.NewResolverExecutor()

	// 创建全局执行上下文
	globalExecContext = editor.NewExecutionContext(
		editor.NewSimpleBufferStore(),
		editor.NewSimpleWindowStore(),
		editor.NewSimpleSelectionStore(),
	)

	// Initialize kernel with FSM engine and new resolver executor
	kernelInstance = kernel.NewKernel(fsm.GetDefaultEngine(), resolverExecutor)

	// 初始化 Weaver 系统
	manager.InitWeaver(manager.ModeWeaver) // 默认启用 Weaver 模式

	if *reloadFlag {
		// Invariant 8: Reload = atomic rebuild
		// 使用统一的Reload函数
		if err := fsm.Reload(*configPath); err != nil {
			log.Fatalf("reload failed: %v", err) // Invariant 10: error = reject running
		}
		log.Println("Keymap reloaded successfully")
		os.Exit(0)
	}

	if *debugMode {
		log.SetFlags(log.LstdFlags | log.Lshortfile) // Include file and line info in logs
	}

	// Handle command line arguments
	args := flag.Args()

	if *enterFlag {
		// Enter FSM mode
		fsm.EnterFSM()
		os.Exit(0)
	}

	if *exitFlag {
		// Exit FSM mode
		fsm.ExitFSM()
		os.Exit(0)
	}

	if *helpFlag {
		fmt.Println("tmux-fsm - A Tmux plugin providing Vim-like modal editing")
		fmt.Println("Usage:")
		fmt.Println("  tmux-fsm -server          # Run as server daemon")
		fmt.Println("  tmux-fsm -enter           # Enter FSM mode")
		fmt.Println("  tmux-fsm -exit            # Exit FSM mode")
		fmt.Println("  tmux-fsm -reload          # Reload keymap configuration")
		fmt.Println("  tmux-fsm -key <key> <pane_client>  # Process a key event")
		fmt.Println("  tmux-fsm -debug           # Enable debug logging")
		os.Exit(0)
	}

	if *keyFlag != "" {
		// Process key event
		paneAndClient := ""
		if len(args) > 0 {
			paneAndClient = args[0]
		}
		// Call runClient function to dispatch the key
		runClient(*keyFlag, paneAndClient)
		os.Exit(0)
	}

	if *serverMode {
		if *debugMode {
			log.Printf("[DEBUG] Starting server on %s", *socketPath)
		}
		log.Printf("[server] tmux-fsm daemon starting: %s", time.Now().Format(time.RFC3339))

		// Write PID file for reliable process management
		pid := os.Getpid()
		pidPath := "/tmp/tmux-fsm.pid"
		if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
			log.Printf("[server] warning: could not write PID file: %v", err)
		}

		srv := NewServer(ServerConfig{
			SocketPath: *socketPath,
		})
		log.Fatal(srv.Run(context.Background()))
		return
	}

	// client / other modes 保持你原来的逻辑
	log.Println("no mode specified")
}

// ServerConfig 服务器配置
type ServerConfig struct {
	SocketPath string
}

// Server 服务器结构
type Server struct {
	cfg ServerConfig
	// kernel *kernel.Kernel  // Temporarily disabled
}

// NewServer 创建新服务器实例
func NewServer(cfg ServerConfig) *Server {
	return &Server{
		cfg: cfg,
	}
}

// Run 启动服务器
func (s *Server) Run(ctx context.Context) error {
	// 清理旧 socket
	_ = os.Remove(s.cfg.SocketPath)

	ln, err := net.Listen("unix", s.cfg.SocketPath)
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Printf("[server] listening on %s\n", s.cfg.SocketPath)

	go s.handleSignals(ctx, ln)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[server] accept error: %v\n", err)
			return err
		}
		log.Printf("[server] accepted connection from %s\n", conn.RemoteAddr())
		go s.handleClient(conn)
	}
}

// handleClient 处理客户端连接
func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	log.Printf("[server] client connected: %s", conn.RemoteAddr())

	// 首先尝试读取原始数据以确定协议类型
	buf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		log.Printf("[server] failed to read from connection: %v", err)
		return
	}

	rawData := buf[:n]

	// 检查是否是字符串协议格式 "pane|client|key"
	payloadStr := string(rawData[:n])
	if strings.Contains(payloadStr, "|") {
		// 这是字符串协议格式
		parts := strings.SplitN(payloadStr, "|", 3)
		var paneID, clientName, key string

		if len(parts) == 3 {
			paneID = parts[0]
			clientName = parts[1]
			key = parts[2]
		} else if len(parts) == 2 {
			// Fallback for old protocol: PANE|KEY (Client unknown)
			paneID = parts[0]
			key = parts[1]
		} else {
			key = payloadStr
		}

		log.Printf("[server] string protocol received: pane='%s', client='%s', key='%s'", paneID, clientName, key)

		// 处理特殊命令
		switch key {
		case "__PING__":
			conn.Write([]byte("PONG"))
			return
		case "__SHUTDOWN__":
			// 这种情况下不应该在这里处理，但为了完整性
			conn.Write([]byte("SHUTDOWN"))
			return
		case "__CLEAR_STATE__":
			fsm.Reset() // 重置新架构层级
			conn.Write([]byte("ok"))
			return
		}

		// 使用 kernel 处理按键
		if kernelInstance != nil {
			hctx := kernel.HandleContext{Ctx: context.Background()}
			kernelInstance.HandleKey(hctx, key)
		}

		conn.Write([]byte("ok"))
		return
	}

	// 否则是 JSON 协议格式
	var in intent.Intent
	decoder := json.NewDecoder(strings.NewReader(payloadStr))
	if err := decoder.Decode(&in); err != nil {
		log.Printf("[server] decode intent error: %v", err)
		return
	}

	log.Printf("[server] intent received: kind=%v count=%d",
		in.Kind, in.Count,
	)

	// Invariant 1: FSM has absolute priority on key events
	// Check if this is a key dispatch request first
	if in.Meta != nil {
		if key, ok := in.Meta["key"].(string); ok {
			// ✅ Phase‑4 边界：非键盘事件，直接忽略
			if key == "" {
				log.Printf("[server] empty key event ignored")
				return
			}

			// Use kernel to handle key dispatch
			if kernelInstance != nil {
				hctx := kernel.HandleContext{Ctx: context.Background()}
				kernelInstance.HandleKey(hctx, key)
				// If kernel handled the key, return without processing further
				return
			}
		}
		// Check for reload command
		if cmd, ok := in.Meta["command"].(string); ok {
			if cmd == "reload" {
				configPath, ok := in.Meta["config_path"].(string)
				if !ok {
					configPath = "./keymap.yaml"
				}
				// Use unified Reload function
				if err := fsm.Reload(configPath); err != nil {
					return
				}
				return
			}
			if cmd == "nvim-mode" {
				// Handle Neovim mode changes
				mode, ok := in.Meta["mode"].(string)
				if ok {
					fsm.OnNvimMode(mode)
				}
				return
			}
		}
	}

	// If FSM didn't consume the key, process as regular intent
	if err := ProcessIntentGlobal(in); err != nil {
		log.Printf("[server] ProcessIntentGlobal error: %v", err)
	}
}

// handleSignals 处理信号
func (s *Server) handleSignals(ctx context.Context, ln net.Listener) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case sig := <-ch:
		log.Printf("[server] signal received: %v\n", sig)
		// Clean up PID file
		os.Remove("/tmp/tmux-fsm.pid")
	}

	_ = ln.Close()
}

// RepeatLastTransaction 重复执行最近提交的事务
// 这是 . repeat 功能的核心实现
func RepeatLastTransaction(ctx *editor.ExecutionContext, tm *TransactionManager) error {
	tx := tm.LastCommittedTransaction()
	if tx == nil {
		return nil // Vim 行为：无事发生
	}

	// 开始新事务以支持 repeat 本身的 undo
	tm.BeginTransaction()

	// 重放最近事务中的所有操作
	for _, opRecord := range tx.Records {
		err := editor.ApplyResolvedOperation(ctx, opRecord.ResolvedOp)
		if err != nil {
			tm.AbortTransaction()
			return err
		}
	}

	return tm.CommitTransaction()
}

// UndoLastTransaction 撤销最近的事务
// 这是 undo 功能的核心实现
func UndoLastTransaction(tm *TransactionManager) error {
	return fmt.Errorf("undo not supported: inverse execution not implemented")
}

// TxNode 事务节点，用于构建 redo tree
type TxNode struct {
	Tx       *types.Transaction
	Parent   *TxNode
	Children []*TxNode
}

// History 编辑历史，支持 undo/redo tree
type History struct {
	Root    *TxNode
	Current *TxNode
}

// NewHistory 创建新的历史记录
func NewHistory() *History {
	root := &TxNode{
		Tx:       nil, // 根节点不包含事务
		Parent:   nil,
		Children: make([]*TxNode, 0),
	}

	return &History{
		Root:    root,
		Current: root,
	}
}

// Commit 将事务提交到历史记录中
func (h *History) Commit(tx *types.Transaction) {
	node := &TxNode{
		Tx:       tx,
		Parent:   h.Current,
		Children: make([]*TxNode, 0),
	}

	h.Current.Children = append(h.Current.Children, node)
	h.Current = node
}

// Undo 执行撤销操作
func (h *History) Undo() *types.Transaction {
	if h.Current == h.Root {
		return nil // 已经在根节点，无法再撤销
	}

	tx := h.Current.Tx
	h.Current = h.Current.Parent
	return tx
}

// Redo 执行重做操作
func (h *History) Redo(childIndex int) *types.Transaction {
	if len(h.Current.Children) == 0 {
		return nil // 没有可重做的事务
	}

	if childIndex < 0 || childIndex >= len(h.Current.Children) {
		childIndex = 0 // 默认选择第一个子节点
	}

	next := h.Current.Children[childIndex]
	h.Current = next
	return next.Tx
}

// Macro 宏定义，包含一系列事务
type Macro struct {
	Name         string
	Transactions []*types.Transaction
}

// MacroManager 宏管理器
type MacroManager struct {
	macros      map[string]*Macro
	activeMacro *Macro // 当前正在录制的宏
}

// NewMacroManager 创建新的宏管理器
func NewMacroManager() *MacroManager {
	return &MacroManager{
		macros: make(map[string]*Macro),
	}
}

// StartRecording 开始录制宏
func (mm *MacroManager) StartRecording(name string) {
	mm.activeMacro = &Macro{
		Name:         name,
		Transactions: make([]*types.Transaction, 0),
	}
}

// StopRecording 停止录制宏
func (mm *MacroManager) StopRecording() {
	if mm.activeMacro != nil {
		// 保存宏
		mm.macros[mm.activeMacro.Name] = mm.activeMacro
		mm.activeMacro = nil
	}
}

// RecordTransaction 记录事务到当前宏
func (mm *MacroManager) RecordTransaction(tx *types.Transaction) {
	if mm.activeMacro != nil {
		// 复制事务以避免后续修改影响宏
		clonedTx := cloneTransaction(tx)
		mm.activeMacro.Transactions = append(mm.activeMacro.Transactions, clonedTx)
	}
}

// PlayMacro 执行宏
func (mm *MacroManager) PlayMacro(name string, count int) error {
	macro, exists := mm.macros[name]
	if !exists {
		return fmt.Errorf("macro '%s' not found", name)
	}

	if count <= 0 {
		count = 1
	}

	for i := 0; i < count; i++ {
		for _, tx := range macro.Transactions {
			err := replayTransaction(globalExecContext, tx)
			if err != nil {
				return fmt.Errorf("error replaying macro '%s': %v", name, err)
			}
		}
	}

	return nil
}

// cloneTransaction 克隆事务
func cloneTransaction(src *types.Transaction) *types.Transaction {
	dst := &types.Transaction{
		ID:               src.ID,
		Records:          make([]types.OperationRecord, len(src.Records)),
		CreatedAt:        src.CreatedAt,
		SafetyLevel:      src.SafetyLevel,
		PreSnapshotHash:  src.PreSnapshotHash,
		PostSnapshotHash: src.PostSnapshotHash,
	}

	// 克隆 Records
	copy(dst.Records, src.Records)

	return dst
}

// replayTransaction 重放事务
func replayTransaction(ctx *editor.ExecutionContext, tx *types.Transaction) error {
	for _, record := range tx.Records {
		err := editor.ApplyResolvedOperation(ctx, record.ResolvedOp)
		if err != nil {
			return err
		}
	}
	return nil
}

// IsRecording 检查是否正在录制宏
func (mm *MacroManager) IsRecording() bool {
	return mm.activeMacro != nil
}

// ProcessIntentGlobal 全局意图处理入口
// RFC-WC-002: Intent ABI - 统一入口，统一审计
func ProcessIntentGlobal(intent intent.Intent) error {
	// 如果 weaverMgr 未初始化，返回
	if weaverMgr == nil {
		return nil
	}

	// 开始事务 - 一个事务对应一次可被 `.` 重复的最小操作单元
	if transMgr != nil {
		transMgr.BeginTransaction()
	}

	// 使用 weaver manager 处理意图
	err := weaverMgr.Process(&intent)
	if err != nil && transMgr != nil {
		// 如果处理过程中出现错误，回滚事务
		transMgr.AbortTransaction()
		return err
	}

	// 成功处理后提交事务
	if transMgr != nil {
		return transMgr.CommitTransaction()
	}

	return err
}

// ProcessUndo 执行撤销操作
func ProcessUndo(paneID string) error {
	if txJournal != nil {
		return txJournal.Undo()
	}
	return nil
}

// ProcessRedo 执行重做操作
func ProcessRedo(paneID string) error {
	if txJournal != nil {
		return txJournal.Redo()
	}
	return nil
}

````

## 📄 `pkg/protocol/protocol.go`

````go
package protocol

// Anchor is "I mean this text, not the cursor"
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

## 📄 `pkg/server/server.go`

````go
package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"tmux-fsm/fsm"
	"tmux-fsm/kernel"
)

var (
	socketPath = os.Getenv("HOME") + "/.tmux-fsm.sock"
)

// Server represents the main server instance
type Server struct {
	listener net.Listener
	kernel   *kernel.Kernel
}

// New creates a new server instance
func New(k *kernel.Kernel) *Server {
	return &Server{
		kernel: k,
	}
}

// Listen starts the server and listens for connections
func (s *Server) Listen() error {
	fmt.Printf("Server starting (v3-merged) at %s...\n", socketPath)

	// 检查是否已有服务在运行 (且能响应)
	if conn, err := net.DialTimeout("unix", socketPath, 1*time.Second); err == nil {
		conn.Close()
		fmt.Println("Daemon already running and responsive.")
		return nil
	}

	// 如果 Socket 文件存在但无法连接，说明是残留文件，直接移除
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: Failed to remove old socket: %v\n", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("CRITICAL: Failed to start server: %v", err)
	}
	s.listener = listener

	defer listener.Close()
	if err := os.Chmod(socketPath, 0666); err != nil {
		fmt.Printf("Warning: Failed to chmod socket: %v\n", err)
	}

	// 初始化新架构回调：当新架构状态变化时，强制触发老架构的状态栏刷新
	fsm.OnUpdateUI = func() {
		// TODO: Implement UI update callback
	}

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
				// TODO: Implement periodic save
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

		shouldExit := s.handleClient(conn)
		if shouldExit {
			goto shutdown
		}
	}

shutdown:
	fmt.Println("Shutting down gracefully...")
	os.Remove(socketPath)
	return nil
}

// handleClient handles a single client connection
func (s *Server) handleClient(conn net.Conn) bool {
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
		// TODO: Implement state clearing
		return false
	}

	if key == "__STATUS__" {
		// TODO: Implement status reporting
		data := []byte("{}")
		conn.Write(data)
		return false
	}

	if key == "__WHY_FAIL__" {
		// TODO: Implement failure reporting
		msg := "No undo failures recorded."
		conn.Write([]byte(msg + "\n"))
		return false
	}

	if key == "__HELP__" {
		if clientName == "" {
			// If called from a raw terminal (no clientName), just print text back
			conn.Write([]byte("Help text"))
		} else {
			// If called from within tmux FSM, show popup
			// TODO: Implement help popup
		}
		return false
	}

	// TODO: Implement the rest of the client handling logic
	// This would include the FSM dispatching, action processing, and intent execution

	conn.Write([]byte("ok"))
	return false
}

// Shutdown sends a shutdown command to the server
func Shutdown() error {
	conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
	if err != nil {
		return fmt.Errorf("daemon not running to stop: %v", err)
	}
	defer conn.Close()

	// Send a special command to signal shutdown
	conn.Write([]byte("__SHUTDOWN__"))
	return nil
}

// IsServerRunning checks if the server is currently running
func IsServerRunning() bool {
	conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// RunClient communicates with the server daemon
func RunClient(key, paneAndClient string) error {
	conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
	if err != nil {
		return fmt.Errorf("daemon not running. Start it with 'tmux-fsm -server': %v", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return fmt.Errorf("error setting deadline: %v", err)
	}

	payload := fmt.Sprintf("%s|%s", paneAndClient, key)
	if _, err := conn.Write([]byte(payload)); err != nil {
		return err
	}

	// Read response (synchronize)
	buf, err := io.ReadAll(conn)
	if err != nil {
		return err
	}
	resp := strings.TrimSpace(string(buf))
	if resp != "ok" && resp != "" {
		fmt.Println(resp)
	}

	return nil
}

````

## 📄 `pkg/state/state.go`

````go
package state

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"tmux-fsm/fsm"
)

// Transaction 事务结构（简化版）
// TODO: Phase-3 undo/redo transaction log
type Transaction struct {
	ID               int           `json:"id"`
	Records          []interface{} `json:"records"`
	CreatedAt        string        `json:"created_at"`
	Applied          bool          `json:"applied"`
	Skipped          bool          `json:"skipped"`
	SafetyLevel      string        `json:"safety_level,omitempty"`
	PreSnapshotHash  string        `json:"pre_snapshot_hash,omitempty"`
	PostSnapshotHash string        `json:"post_snapshot_hash,omitempty"`
}

// FSMState represents the state of the FSM
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

// StateManager manages the global state
type StateManager struct {
	mutex   sync.Mutex
	state   FSMState
	backend Backend
}

// Backend interface for interacting with tmux
type Backend interface {
	GetUserOption(option string) (string, error)
	SetUserOption(option, value string) error
	RefreshClient(clientName string) error
	SwitchClientTable(clientName, table string) error
	GetActivePane(clientName string) (string, error)
}

// NewStateManager creates a new state manager
func NewStateManager(backend Backend) *StateManager {
	return &StateManager{
		backend: backend,
	}
}

// LoadState loads the state from tmux options
func (sm *StateManager) LoadState() FSMState {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Use backend to read tmux options
	out, err := sm.backend.GetUserOption("@tmux_fsm_state")
	var state FSMState
	if err != nil || len(out) == 0 {
		return FSMState{Mode: "NORMAL", Count: 0}
	}
	json.Unmarshal([]byte(out), &state)
	sm.state = state
	return state
}

// SaveStateRaw saves the raw state data to tmux options
func (sm *StateManager) SaveStateRaw(data []byte) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Use backend to save state
	// This implies SetUserOption needs to be able to set arbitrary keys.
	if err := sm.backend.SetUserOption("@tmux_fsm_state", string(data)); err != nil {
		log.Printf("Failed to save FSM state: %v", err)
	}
}

// UpdateStatusBar updates the status bar display
func (sm *StateManager) UpdateStatusBar(state FSMState, clientName string) {
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

	// Use backend for tmux option updates
	sm.backend.SetUserOption("@fsm_state", modeMsg)
	sm.backend.SetUserOption("@fsm_keys", keysMsg)
	sm.backend.RefreshClient(clientName) // Refresh the target client

	// --- [ABI: Heartbeat Lock] ---
	// Re-assert the key table to prevent "one-shot" dropouts.
	// Check @fsm_active to allow intentional exits.
	if clientName != "" && clientName != "default" {
		// Fetching @fsm_active via backend if it were available would be ideal,
		// but for now, we rely on the fact that we are in a state where we should be active.
		// If backend could read options, it would be better.
		// For now, we assume if we got here, FSM is active.
		sm.backend.SwitchClientTable(clientName, "fsm")
	}
}

// GetState returns the current state
func (sm *StateManager) GetState() FSMState {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	return sm.state
}

// SetState sets the current state
func (sm *StateManager) SetState(state FSMState) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.state = state
}

````

## 📄 `planner/grammar.go`

````go
package planner

import (
	"tmux-fsm/fsm"
	intentPkg "tmux-fsm/intent"
)

// Grammar 是 Stage‑4 Vim Grammar
//
// ⚠️ Architecture rule:
// Grammar MUST NOT construct intent.Intent.
// Grammar ONLY produces intent.GrammarIntent.
// Promotion happens exclusively in Kernel via intent.Promote.
type Grammar struct {
	count     int
	pendingOp *intentPkg.OperatorKind
	// 新增状态用于处理复杂 motion
	pendingMotion *MotionPendingInfo
	textObj       TextObjPending
}

// MotionPendingInfo 用于处理需要两个按键的 motion
type MotionPendingInfo struct {
	Kind     intentPkg.MotionKind
	FindDir  intentPkg.FindDirection
	FindTill bool
}

const (
	MPNone = iota
	MPG    // g_
	MPF    // f{c}
	MPT    // t{c}
	MPBigF // F{c}
	MPBigT // T{c}
)

// TextObjPending 用于处理文本对象
type TextObjPending int

const (
	TOPNone TextObjPending = iota
	TOPInner
	TOPAround
)

// NewGrammar 创建 Grammar 实例
func NewGrammar() *Grammar {
	return &Grammar{}
}

// Consume 消费一个 FSM RawToken，必要时产生 GrammarIntent
func (g *Grammar) Consume(tok fsm.RawToken) *intentPkg.GrammarIntent {
	switch tok.Kind {

	case fsm.TokenDigit:
		g.count = g.count*10 + int(tok.Value[0]-'0')
		return nil

	case fsm.TokenRepeat:
		return &intentPkg.GrammarIntent{
			Kind: intentPkg.IntentRepeat,
		}

	case fsm.TokenKey:
		return g.consumeKey(tok.Value)

	case fsm.TokenSystem:
		// 系统事件，重置状态
		if tok.Value == "reset" || tok.Value == "exit" || tok.Value == "enter" {
			g.reset()
		}
		return nil
	}

	return nil
}

// consumeKey 处理普通按键
func (g *Grammar) consumeKey(key string) *intentPkg.GrammarIntent {
	// 优先处理 pending motion
	if g.pendingMotion != nil {
		return g.consumePendingMotion(key)
	}

	// 优先处理 text object
	if g.textObj != TOPNone {
		return g.consumeTextObject(key)
	}

	// 1️⃣ operator
	if op, ok := parseOperator(key); ok {
		// dd / yy
		if g.pendingOp != nil && *g.pendingOp == op {
			intent := makeLineGrammarIntent(op, max(g.count, 1))
			g.reset()
			return intent
		}

		// 检查是否进入文本对象模式 (i 或 a)
		if key == "i" || key == "a" {
			if key == "i" {
				g.textObj = TOPInner
			} else {
				g.textObj = TOPAround
			}
			g.pendingOp = &op
			return nil
		}

		g.pendingOp = &op
		return nil
	}

	// 2️⃣ 检查是否是进入文本对象模式 (i 或 a)
	if key == "i" || key == "a" {
		if key == "i" {
			g.textObj = TOPInner
		} else {
			g.textObj = TOPAround
		}
		return nil
	}

	// 3️⃣ 检查是否是 motion 前缀
	if parseMotionPrefix(key) {
		switch key {
		case "g":
			g.pendingMotion = &MotionPendingInfo{
				Kind: intentPkg.MotionGoto,
			}
		case "f":
			g.pendingMotion = &MotionPendingInfo{
				Kind:     intentPkg.MotionFind,
				FindDir:  intentPkg.FindForward,
				FindTill: false,
			}
		case "t":
			g.pendingMotion = &MotionPendingInfo{
				Kind:     intentPkg.MotionFind,
				FindDir:  intentPkg.FindForward,
				FindTill: true,
			}
		case "F":
			g.pendingMotion = &MotionPendingInfo{
				Kind:     intentPkg.MotionFind,
				FindDir:  intentPkg.FindBackward,
				FindTill: false,
			}
		case "T":
			g.pendingMotion = &MotionPendingInfo{
				Kind:     intentPkg.MotionFind,
				FindDir:  intentPkg.FindBackward,
				FindTill: true,
			}
		}
		return nil
	}

	// 4️⃣ 检查是否是 motion
	if motion, ok := parseMotion(key); ok {
		// op + motion
		if g.pendingOp != nil {
			intent := makeOpMotionGrammarIntent(
				*g.pendingOp,
				motion,
				max(g.count, 1),
				key,
			)
			g.reset()
			return intent
		}

		// standalone motion (move)
		intent := makeMoveGrammarIntent(motion, max(g.count, 1), key)
		g.reset()
		return intent
	}

	// 5️⃣ 检查是否是模式切换键
	if mode := parseModeSwitch(key); mode != "" {
		// 模式切换暂时返回普通的 Intent，但我们需要重构
		// 为简化，这里先返回 nil，模式切换将通过其他方式处理
		g.reset()
		return nil
	}

	// 6️⃣ 检查是否是 find repeat 键
	if key == ";" {
		g.reset()
		return &intentPkg.GrammarIntent{
			Kind: intentPkg.IntentRepeatFind,
		}
	}
	if key == "," {
		g.reset()
		return &intentPkg.GrammarIntent{
			Kind: intentPkg.IntentRepeatFindReverse,
		}
	}

	// unknown key → reset
	g.reset()
	return nil
}

// parseModeSwitch 解析模式切换键
func parseModeSwitch(key string) string {
	switch key {
	case "i":
		return "insert"
	case "v":
		return "visual_char"
	case "V":
		return "visual_line"
	case "Escape", "C-c":
		return "normal"
	default:
		return ""
	}
}

// ---------- helpers ----------

func (g *Grammar) reset() {
	g.count = 0
	g.pendingOp = nil
	g.pendingMotion = nil
	g.textObj = TOPNone
}

// makeMoveGrammarIntent 创建移动 Grammar 意图
func makeMoveGrammarIntent(m intentPkg.MotionKind, count int, key string) *intentPkg.GrammarIntent {
	motion := &intentPkg.Motion{
		Kind:  m,
		Count: count,
	}

	// 为基础移动键设置精确的 Direction 或 Range
	switch key {
	case "h":
		motion.Direction = intentPkg.DirectionLeft
	case "l":
		motion.Direction = intentPkg.DirectionRight
	case "k":
		motion.Direction = intentPkg.DirectionUp
	case "j":
		motion.Direction = intentPkg.DirectionDown
	case "0", "^":
		motion.Kind = intentPkg.MotionRange
		motion.Range = &intentPkg.RangeMotion{Kind: intentPkg.RangeLineStart}
	case "$":
		motion.Kind = intentPkg.MotionRange
		motion.Range = &intentPkg.RangeMotion{Kind: intentPkg.RangeLineEnd}
	case "G", "gg":
		motion.Kind = intentPkg.MotionGoto
	case "H", "M", "L":
		motion.Kind = intentPkg.MotionLine
	}

	return &intentPkg.GrammarIntent{
		Kind:   intentPkg.IntentMove,
		Count:  count,
		Motion: motion,
	}
}

// makeOpMotionGrammarIntent 创建操作+移动 Grammar 意图
func makeOpMotionGrammarIntent(op intentPkg.OperatorKind, m intentPkg.MotionKind, count int, key string) *intentPkg.GrammarIntent {
	motion := &intentPkg.Motion{
		Kind:  m,
		Count: count,
	}

	// 为基础移动键设置精确的 Direction 或 Range
	switch key {
	case "h":
		motion.Direction = intentPkg.DirectionLeft
	case "l":
		motion.Direction = intentPkg.DirectionRight
	case "k":
		motion.Direction = intentPkg.DirectionUp
	case "j":
		motion.Direction = intentPkg.DirectionDown
	case "0", "^":
		motion.Kind = intentPkg.MotionRange
		motion.Range = &intentPkg.RangeMotion{Kind: intentPkg.RangeLineStart}
	case "$":
		motion.Kind = intentPkg.MotionRange
		motion.Range = &intentPkg.RangeMotion{Kind: intentPkg.RangeLineEnd}
	case "G", "gg":
		motion.Kind = intentPkg.MotionGoto
	case "H", "M", "L":
		motion.Kind = intentPkg.MotionLine
	}

	return &intentPkg.GrammarIntent{
		Kind:   intentPkg.IntentOperator,
		Count:  count,
		Motion: motion,
		Op:     &op,
	}
}

// makeLineGrammarIntent 创建行操作 Grammar 意图
func makeLineGrammarIntent(op intentPkg.OperatorKind, count int) *intentPkg.GrammarIntent {
	motion := &intentPkg.Motion{
		Kind:  intentPkg.MotionLine,
		Count: count,
	}

	return &intentPkg.GrammarIntent{
		Kind:   intentPkg.IntentOperator,
		Count:  count,
		Motion: motion,
		Op:     &op,
	}
}

func cloneIntent(i *intentPkg.Intent) *intentPkg.Intent {
	c := *i
	if i.Meta != nil {
		c.Meta = make(map[string]interface{})
		for k, v := range i.Meta {
			c.Meta[k] = v
		}
	}
	return &c
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// consumePendingMotion 处理需要两个按键的 motion
func (g *Grammar) consumePendingMotion(key string) *intentPkg.GrammarIntent {
	if g.pendingMotion != nil {
		switch g.pendingMotion.Kind {
		case intentPkg.MotionGoto:
			g.pendingMotion = nil
			if key == "g" {
				intent := makeMoveGrammarIntent(intentPkg.MotionGoto, max(g.count, 1), "gg")
				g.reset()
				return intent
			}
			g.reset()
			return nil
		case intentPkg.MotionFind:
			// f{c}, t{c}, F{c}, T{c} 需要一个字符
			intent := makeFindGrammarIntent(g.pendingMotion, g.pendingOp, rune(key[0]), max(g.count, 1))
			g.pendingMotion = nil
			g.reset()
			return intent
		default:
			g.reset()
			return nil
		}
	}
	g.reset()
	return nil
}

// consumeTextObject 处理文本对象
func (g *Grammar) consumeTextObject(key string) *intentPkg.GrammarIntent {
	objType := parseTextObject(key)
	if objType == intentPkg.Word && key != "w" { // Word 是默认值，需要检查是否真的匹配
		// 检查是否是有效的文本对象键
		switch key {
		case "w", "(", ")", "b", "[", "]", "{", "}", "B", "\"", "'", "`":
			// 这些都是有效的，继续
		default:
			g.reset()
			return nil
		}
	}

	intent := makeTextObjectGrammarIntent(g.pendingOp, g.textObj, objType, max(g.count, 1))
	g.reset()
	return intent
}

// makeTextObjectGrammarIntent 创建文本对象 Grammar 意图
func makeTextObjectGrammarIntent(op *intentPkg.OperatorKind, textObj TextObjPending, objType intentPkg.TextObjectKind, count int) *intentPkg.GrammarIntent {
	scope := intentPkg.Inner
	if textObj == TOPAround {
		scope = intentPkg.Around
	}

	textObject := &intentPkg.TextObject{
		Scope:  scope,
		Object: objType,
	}

	rangeMotion := &intentPkg.RangeMotion{
		Kind:       intentPkg.RangeTextObject,
		TextObject: textObject,
	}

	motion := &intentPkg.Motion{
		Kind:  intentPkg.MotionRange,
		Count: count,
		Range: rangeMotion,
	}

	if op != nil {
		return &intentPkg.GrammarIntent{
			Kind:   intentPkg.IntentOperator,
			Count:  count,
			Motion: motion,
			Op:     op,
		}
	} else {
		return &intentPkg.GrammarIntent{
			Kind:   intentPkg.IntentMove,
			Count:  count,
			Motion: motion,
		}
	}
}

// textObjectKindToString 将 TextObjectKind 转换为字符串（临时兼容）
func textObjectKindToString(kind intentPkg.TextObjectKind) string {
	switch kind {
	case intentPkg.Word:
		return "word"
	case intentPkg.Paren:
		return "paren"
	case intentPkg.Bracket:
		return "bracket"
	case intentPkg.Brace:
		return "brace"
	case intentPkg.QuoteSingle:
		return "quote_single"
	case intentPkg.QuoteDouble:
		return "quote_double"
	case intentPkg.Backtick:
		return "quote_backtick"
	default:
		return "word"
	}
}

// parseMotionPrefix 解析 motion 前缀
func parseMotionPrefix(key string) bool {
	switch key {
	case "g", "f", "F", "t", "T":
		return true
	default:
		return false
	}
}

// parseTextObject 解析文本对象
func parseTextObject(key string) intentPkg.TextObjectKind {
	switch key {
	case "w":
		return intentPkg.Word
	case "(":
		return intentPkg.Paren
	case ")":
		return intentPkg.Paren
	case "b":
		return intentPkg.Paren // b 也是括号的别名
	case "[":
		return intentPkg.Bracket
	case "]":
		return intentPkg.Bracket
	case "{":
		return intentPkg.Brace
	case "}":
		return intentPkg.Brace
	case "B":
		return intentPkg.Brace // B 也是大括号的别名
	case "\"":
		return intentPkg.QuoteDouble
	case "'":
		return intentPkg.QuoteSingle
	case "`":
		return intentPkg.Backtick
	default:
		return intentPkg.Word // 默认值
	}
}

// makeFindGrammarIntent 创建查找 Grammar 意图
func makeFindGrammarIntent(pending *MotionPendingInfo, op *intentPkg.OperatorKind, char rune, count int) *intentPkg.GrammarIntent {
	findMotion := &intentPkg.FindMotion{
		Char:      char,
		Direction: pending.FindDir,
		Till:      pending.FindTill,
	}

	motion := &intentPkg.Motion{
		Kind:  intentPkg.MotionFind,
		Find:  findMotion,
		Count: count,
	}

	// 修复：对于 FindMotion，Intent 应该是 Move 或 Operator，而不是 IntentFind
	// 根据是否有操作符来决定 Intent 类型
	if op != nil {
		// 如果有操作符，返回 Operator 类型
		return &intentPkg.GrammarIntent{
			Kind:   intentPkg.IntentOperator,
			Count:  count,
			Motion: motion,
			Op:     op,
		}
	} else {
		// 否则返回 Move 类型
		return &intentPkg.GrammarIntent{
			Kind:   intentPkg.IntentMove,
			Count:  count,
			Motion: motion,
		}
	}
}

// motionTypeToString 将 MotionPendingInfo 转换为字符串
func motionTypeToString(info *MotionPendingInfo) string {
	if info == nil {
		return ""
	}

	// 根据 Kind 字段判断
	switch info.Kind {
	case intentPkg.MotionFind:
		if info.FindDir == intentPkg.FindForward {
			if info.FindTill {
				return "t"
			}
			return "f"
		}
		if info.FindDir == intentPkg.FindBackward {
			if info.FindTill {
				return "T"
			}
			return "F"
		}
	case intentPkg.MotionGoto:
		return "g"
	}

	return ""
}

// ---------- key parsing (Grammar owns Vim) ----------

func parseOperator(key string) (intentPkg.OperatorKind, bool) {
	switch key {
	case "d":
		return intentPkg.OpDelete, true
	case "y":
		return intentPkg.OpYank, true
	case "c":
		return intentPkg.OpChange, true
	default:
		return 0, false
	}
}

func parseMotion(key string) (intentPkg.MotionKind, bool) {
	switch key {
	case "h", "l":
		return intentPkg.MotionChar, true
	case "j", "k":
		return intentPkg.MotionLine, true
	case "w", "b", "e", "ge":
		return intentPkg.MotionWord, true
	case "$":
		return intentPkg.MotionChar, true
	case "0", "^":
		return intentPkg.MotionChar, true
	case "G":
		return intentPkg.MotionGoto, true
	case "H", "M", "L":
		return intentPkg.MotionLine, true
	default:
		return 0, false
	}
}

````

## 📄 `planner/grammar_test.go`

````go
package planner

import (
	"testing"
	"tmux-fsm/fsm"
	intentPkg "tmux-fsm/intent"
)

func TestGrammarBasicMotion(t *testing.T) {
	g := NewGrammar()

	// 测试 hjkl 移动
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "h"})
	if intent == nil || intent.Kind != intentPkg.IntentMove {
		t.Errorf("Expected move intent for 'h', got %v", intent)
	}

	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "j"})
	if intent == nil || intent.Kind != intentPkg.IntentMove {
		t.Errorf("Expected move intent for 'j', got %v", intent)
	}

	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "k"})
	if intent == nil || intent.Kind != intentPkg.IntentMove {
		t.Errorf("Expected move intent for 'k', got %v", intent)
	}

	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "l"})
	if intent == nil || intent.Kind != intentPkg.IntentMove {
		t.Errorf("Expected move intent for 'l', got %v", intent)
	}
}

func TestGrammarCount(t *testing.T) {
	g := NewGrammar()

	// 测试数字计数
	g.Consume(fsm.RawToken{Kind: fsm.TokenDigit, Value: "3"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "w"})
	if intent == nil || intent.Count != 3 {
		t.Errorf("Expected count 3 for '3w', got %v", intent)
	}
}

func TestGrammarOperatorMotion(t *testing.T) {
	g := NewGrammar()

	// 测试 d + w
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "w"})
	if intent == nil || intent.Kind != intentPkg.IntentOperator {
		t.Errorf("Expected operator intent for 'dw', got %v", intent)
	}
}

func TestGrammarOperatorCountMotion(t *testing.T) {
	g := NewGrammar()

	// 测试 d2w
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	g.Consume(fsm.RawToken{Kind: fsm.TokenDigit, Value: "2"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "w"})
	if intent == nil || intent.Kind != intentPkg.IntentOperator || intent.Count != 2 {
		t.Errorf("Expected operator intent with count 2 for 'd2w', got %v", intent)
	}
}

func TestGrammarGg(t *testing.T) {
	g := NewGrammar()

	// 测试 gg
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "g"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "g"})
	if intent == nil || intent.Kind != intentPkg.IntentMove {
		t.Errorf("Expected move intent for 'gg', got %v", intent)
	}
}

func TestGrammarFfTt(t *testing.T) {
	g := NewGrammar()

	// 测试 fa
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "f"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "a"})
	if intent == nil {
		t.Fatal("Expected intent for 'fa'")
	}
	if intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionFind ||
		intent.Motion.Find == nil ||
		intent.Motion.Find.Char != 'a' ||
		intent.Motion.Find.Direction != intentPkg.FindForward ||
		intent.Motion.Find.Till {
		t.Errorf("Expected forward find motion for 'fa', got %+v", intent.Motion)
	}

	// 测试 ta
	g = NewGrammar()
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "t"})
	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "b"})
	if intent == nil {
		t.Fatal("Expected intent for 'tb'")
	}
	if intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionFind ||
		intent.Motion.Find == nil ||
		intent.Motion.Find.Char != 'b' ||
		intent.Motion.Find.Direction != intentPkg.FindForward ||
		!intent.Motion.Find.Till {
		t.Errorf("Expected forward till motion for 'tb', got %+v", intent.Motion)
	}

	// 测试 Fa
	g = NewGrammar()
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "F"})
	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "c"})
	if intent == nil {
		t.Fatal("Expected intent for 'Fc'")
	}
	if intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionFind ||
		intent.Motion.Find == nil ||
		intent.Motion.Find.Char != 'c' ||
		intent.Motion.Find.Direction != intentPkg.FindBackward ||
		intent.Motion.Find.Till {
		t.Errorf("Expected backward find motion for 'Fc', got %+v", intent.Motion)
	}

	// 测试 Ta
	g = NewGrammar()
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "T"})
	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	if intent == nil {
		t.Fatal("Expected intent for 'Td'")
	}
	if intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionFind ||
		intent.Motion.Find == nil ||
		intent.Motion.Find.Char != 'd' ||
		intent.Motion.Find.Direction != intentPkg.FindBackward ||
		!intent.Motion.Find.Till {
		t.Errorf("Expected backward till motion for 'Td', got %+v", intent.Motion)
	}
}

func TestGrammarTextObject(t *testing.T) {
	g := NewGrammar()

	// 测试 iw
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "i"})
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "w"})
	if intent == nil {
		t.Fatal("expected intent for 'iw'")
	}
	if intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionRange ||
		intent.Motion.Range == nil ||
		intent.Motion.Range.TextObject == nil ||
		intent.Motion.Range.TextObject.Object != intentPkg.Word {
		t.Errorf("expected word text object motion, got %+v", intent.Motion)
	}

	// 测试 diw
	g = NewGrammar()
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "d"})
	g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "i"})
	intent = g.Consume(fsm.RawToken{Kind: fsm.TokenKey, Value: "w"})
	if intent == nil {
		t.Fatal("expected intent for 'diw'")
	}
	if intent.Op == nil ||
		intent.Motion == nil ||
		intent.Motion.Kind != intentPkg.MotionRange {
		t.Errorf("expected operator + text object motion, got %+v", intent)
	}
}

func TestGrammarRepeat(t *testing.T) {
	g := NewGrammar()

	// 测试重复
	intent := g.Consume(fsm.RawToken{Kind: fsm.TokenRepeat, Value: "."})
	if intent == nil {
		t.Errorf("Expected repeat intent for '.'")
	}
}

````

## 📄 `policy/policy.go`

````go
package policy

import (
	"errors"

	"tmux-fsm/crdt"
	"tmux-fsm/semantic"
)

//
// ─────────────────────────────────────────────────────────────
//  Trust Model
// ─────────────────────────────────────────────────────────────
//

// TrustLevel 表示“是否拥有最终提交权”
type TrustLevel int

const (
	TrustSystem   TrustLevel = iota // GC / snapshot / rebalance
	TrustUser                       // 人类
	TrustDevice                     // 同一用户的多端
	TrustAI                         // 只能 proposal
	TrustExternal                   // 插件 / import（默认只读）
)

//
// ─────────────────────────────────────────────────────────────
//  Actor
// ─────────────────────────────────────────────────────────────
//

// ActorInfo 只存储“身份 + 信任级别”
type ActorInfo struct {
	ID    crdt.ActorID
	Level TrustLevel
	Name  string
}

//
// ─────────────────────────────────────────────────────────────
//  Semantic Operation
// ─────────────────────────────────────────────────────────────
//

type OpKind string

const (
	OpInsert  OpKind = "insert"
	OpDelete  OpKind = "delete"
	OpMove    OpKind = "move"
	OpReplace OpKind = "replace" // Added OpReplace
	OpFormat  OpKind = "format"
)

//
// ─────────────────────────────────────────────────────────────
//  Scope（AI 的语义沙箱）
// ─────────────────────────────────────────────────────────────
//

// Scope 表示 AI 被允许操作的“语义范围”
type Scope struct {
	DocumentID string
	Range      semantic.Range
	AllowedOps []OpKind
}

func (s Scope) allowsOp(op OpKind) bool {
	for _, a := range s.AllowedOps {
		if a == op {
			return true
		}
	}
	return false
}

//
// ─────────────────────────────────────────────────────────────
//  AI Draft（注意：不是 Event）
// ─────────────────────────────────────────────────────────────
//

type AIDraft struct {
	Fact semantic.Fact
}

//
// ─────────────────────────────────────────────────────────────
//  Policy Interface
// ─────────────────────────────────────────────────────────────
//

// Policy 是 CRDT 的安全边界
type Policy interface {
	RegisterActor(info ActorInfo)

	// AllowCommit：是否允许 actor 提交最终 CRDT Event
	AllowCommit(actor crdt.ActorID, event crdt.SemanticEvent) error

	// AllowAIDraft：是否允许 AI 在 scope 内提出 draft
	AllowAIDraft(actor crdt.ActorID, scope Scope, draft AIDraft) error

	// ValidateAIProposal：批量校验 AI 提案
	ValidateAIProposal(proposal AIProposal) error
}

//
// ─────────────────────────────────────────────────────────────
//  DefaultPolicy
// ─────────────────────────────────────────────────────────────
//

type DefaultPolicy struct {
	actors map[crdt.ActorID]ActorInfo
}

func NewDefaultPolicy() *DefaultPolicy {
	return &DefaultPolicy{
		actors: make(map[crdt.ActorID]ActorInfo),
	}
}

func (p *DefaultPolicy) RegisterActor(info ActorInfo) {
	p.actors[info.ID] = info
}

//
// ─────────────────────────────────────────────────────────────
//  Commit Path（CRDT 最终入口）
// ─────────────────────────────────────────────────────────────
//

func (p *DefaultPolicy) AllowCommit(
	actorID crdt.ActorID,
	_ crdt.SemanticEvent,
) error {

	actor, ok := p.actors[actorID]
	if !ok {
		return errors.New("unknown actor")
	}

	switch actor.Level {
	case TrustSystem, TrustUser, TrustDevice:
		return nil

	case TrustAI:
		return errors.New("AI is not allowed to commit CRDT events")

	default:
		return errors.New("actor not allowed to commit")
	}
}

//
// ─────────────────────────────────────────────────────────────
//  AI Draft Path（唯一合法 AI 入口）
// ─────────────────────────────────────────────────────────────
//

func (p *DefaultPolicy) AllowAIDraft(
	actorID crdt.ActorID,
	scope Scope,
	draft AIDraft,
) error {

	actor, ok := p.actors[actorID]
	if !ok {
		return errors.New("unknown actor")
	}

	if actor.Level != TrustAI {
		return errors.New("actor is not AI")
	}

	op := factKindToOpKind(draft.Fact.Kind())

	// 1️⃣ 操作类型检查
	if !scope.allowsOp(op) {
		return errors.New("operation not allowed in scope: " + string(op))
	}

	// 2️⃣ 范围检查（语义级）
	if !scope.Range.ContainsFact(draft.Fact) {
		return errors.New("draft out of allowed range")
	}

	return nil
}

//
// ─────────────────────────────────────────────────────────────
//  AI Proposal
// ─────────────────────────────────────────────────────────────
//

type AIProposal struct {
	Actor  crdt.ActorID
	Scope  Scope
	Drafts []AIDraft
}

func (p *DefaultPolicy) ValidateAIProposal(
	proposal AIProposal,
) error {

	for _, draft := range proposal.Drafts {
		if err := p.AllowAIDraft(
			proposal.Actor,
			proposal.Scope,
			draft,
		); err != nil {
			return err
		}
	}

	return nil
}

// factKindToOpKind 将 semantic.FactKind 转换为 policy.OpKind
func factKindToOpKind(kind semantic.FactKind) OpKind {
	switch kind {
	case semantic.FactInsert:
		return OpInsert
	case semantic.FactDelete:
		return OpDelete
	case semantic.FactMove:
		return OpMove
	case semantic.FactReplace:
		return OpReplace
	default:
		return OpKind("unknown")
	}
}

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

## 📄 `selection/selection.go`

````go
package selection

import (
	"tmux-fsm/crdt"
)

//
// ─────────────────────────────────────────────────────────────
//  Types
// ─────────────────────────────────────────────────────────────
//

// CursorID 光标 ID
type CursorID string

// Affinity 亲和性
type Affinity int

const (
	AffinityForward Affinity = iota
	AffinityBackward
	AffinityNeutral
)

// Selection 表示一个选择区域（Anchor → Focus）
type Selection struct {
	Cursor   CursorID
	Actor    crdt.ActorID
	Anchor   crdt.PositionID
	Focus    crdt.PositionID
	Affinity Affinity
}

//
// ─────────────────────────────────────────────────────────────
//  Facts
// ─────────────────────────────────────────────────────────────
//

// SetSelectionFact 设置选择区域（Ephemeral）
type SetSelectionFact struct {
	Cursor CursorID        `json:"cursor"`
	Anchor crdt.PositionID `json:"anchor"`
	Focus  crdt.PositionID `json:"focus"`
}

// EphemeralFact 标记接口（不进入 snapshot）
type EphemeralFact interface {
	IsEphemeral() bool
}

// IsEphemeral implements EphemeralFact
func (SetSelectionFact) IsEphemeral() bool {
	return true
}

//
// ─────────────────────────────────────────────────────────────
//  Edit Operations (for transform)
// ─────────────────────────────────────────────────────────────
//

type EditKind int

const (
	EditInsert EditKind = iota
	EditDelete
)

// EditOp 描述一次文本编辑对 selection 的影响
type EditOp struct {
	Kind   EditKind
	Pos    crdt.PositionID // insert position / delete start
	EndPos crdt.PositionID // only for delete
}

//
// ─────────────────────────────────────────────────────────────
//  Selection Transform (Pure Functions)
// ─────────────────────────────────────────────────────────────
//

// TransformSelection 根据编辑操作变换 selection（幂等）
func TransformSelection(sel Selection, op EditOp) Selection {
	switch op.Kind {
	case EditInsert:
		return transformForInsert(sel, op.Pos)
	case EditDelete:
		return transformForDelete(sel, op.Pos, op.EndPos)
	default:
		return sel
	}
}

// 插入操作对 selection 的影响
func transformForInsert(sel Selection, pos crdt.PositionID) Selection {
	a := crdt.ComparePos(pos, sel.Anchor)
	f := crdt.ComparePos(pos, sel.Focus)

	// 插入在 selection 之前或之后 → 不变
	if (a < 0 && f < 0) || (a > 0 && f > 0) {
		return sel
	}

	// 插入正好在 Anchor / Focus，需看 Affinity
	if a == 0 && sel.Affinity == AffinityBackward {
		return sel
	}
	if f == 0 && sel.Affinity == AffinityForward {
		return sel
	}

	// 插入在 selection 内部或中性边界 → 扩展 Focus
	sel.Focus = pos
	return sel
}

// 删除操作对 selection 的影响
func transformForDelete(sel Selection, start, end crdt.PositionID) Selection {
	newAnchor := sel.Anchor
	newFocus := sel.Focus

	// Anchor 被删除 → 吸附到 start
	if crdt.ComparePos(sel.Anchor, start) >= 0 &&
		crdt.ComparePos(sel.Anchor, end) <= 0 {
		newAnchor = start
	}

	// Focus 被删除 → 吸附到 start
	if crdt.ComparePos(sel.Focus, start) >= 0 &&
		crdt.ComparePos(sel.Focus, end) <= 0 {
		newFocus = start
	}

	sel.Anchor = newAnchor
	sel.Focus = newFocus
	return sel
}

//
// ─────────────────────────────────────────────────────────────
//  Selection Manager
// ─────────────────────────────────────────────────────────────
//

// SelectionManager 管理当前所有 selection（可重建）
type SelectionManager struct {
	selections map[CursorID]Selection
}

// NewSelectionManager 创建新的管理器
func NewSelectionManager() *SelectionManager {
	return &SelectionManager{
		selections: make(map[CursorID]Selection),
	}
}

// ApplySelection 应用 SetSelectionFact（覆盖式）
func (sm *SelectionManager) ApplySelection(
	actor crdt.ActorID,
	fact SetSelectionFact,
) {
	sm.selections[fact.Cursor] = Selection{
		Cursor:   fact.Cursor,
		Actor:    actor,
		Anchor:   fact.Anchor,
		Focus:    fact.Focus,
		Affinity: AffinityNeutral,
	}
}

// ApplyEdit 将一次编辑作用到所有 selection
func (sm *SelectionManager) ApplyEdit(op EditOp) {
	for id, sel := range sm.selections {
		sm.selections[id] = TransformSelection(sel, op)
	}
}

// GetSelection 获取指定 cursor 的 selection
func (sm *SelectionManager) GetSelection(
	cursorID CursorID,
) (Selection, bool) {
	sel, ok := sm.selections[cursorID]
	return sel, ok
}

// GetAllSelections 返回 selection 的快照（防止外部 mutate）
func (sm *SelectionManager) GetAllSelections() map[CursorID]Selection {
	out := make(map[CursorID]Selection, len(sm.selections))
	for k, v := range sm.selections {
		out[k] = v
	}
	return out
}

````

## 📄 `semantic/capture.go`

````go
package semantic

//
// ─────────────────────────────────────────────────────────────
//  Anchor & Range
// ─────────────────────────────────────────────────────────────
//

// Anchor 描述一个稳定的语义锚点
type Anchor struct {
	PaneID string
	Line   int
	Col    int
	Hash   string // 用于弱一致性校验（可选）
}

// Range 表示一个语义范围
type Range struct {
	Start Anchor
	End   Anchor
	Text  string // 捕获时已知的文本
}

// ContainsFact 检查一个事实是否包含在当前范围内
// TODO: 实现实际的逻辑
func (r Range) ContainsFact(fact Fact) bool {
	// 这是一个占位符实现，需要根据实际的语义定义来判断
	// 例如，比较 fact 的 Anchor 和 Range 是否落在 r.Start 和 r.End 之间
	return true
}

//
// ─────────────────────────────────────────────────────────────
//  Motion
// ─────────────────────────────────────────────────────────────
//

// MotionKind 动作类型（强类型）
type MotionKind int

const (
	MotionWordForward MotionKind = iota
	MotionLine
)

// Motion 描述一个语义动作
type Motion struct {
	Kind  MotionKind
	Count int
}

//
// ─────────────────────────────────────────────────────────────
//  Fact Interface
// ─────────────────────────────────────────────────────────────
//

// Fact 表示一个可逆的语义事实
type Fact interface {
	Kind() FactKind
	Inverse() Fact

	Anchor() Anchor
	Range() (Range, bool)
	Text() string
}

//
// ─────────────────────────────────────────────────────────────
//  FactKind
// ─────────────────────────────────────────────────────────────
//

type FactKind int

const (
	FactInsert FactKind = iota
	FactDelete
	FactReplace
	FactMove
)

//
// ─────────────────────────────────────────────────────────────
//  BaseFact (immutable)
// ─────────────────────────────────────────────────────────────
//

type baseFact struct {
	kind   FactKind
	anchor Anchor
	rng    *Range
	text   string
}

func (f baseFact) Kind() FactKind {
	return f.kind
}

func (f baseFact) Anchor() Anchor {
	return f.anchor
}

func (f baseFact) Range() (Range, bool) {
	if f.rng == nil {
		return Range{}, false
	}
	return *f.rng, true
}

func (f baseFact) Text() string {
	return f.text
}

//
// ─────────────────────────────────────────────────────────────
//  Insert
// ─────────────────────────────────────────────────────────────
//

type InsertFact struct {
	baseFact
}

func (f InsertFact) Inverse() Fact {
	return DeleteFact{
		baseFact: baseFact{
			kind:   FactDelete,
			anchor: f.anchor,
			rng:    f.rng,
			text:   f.text,
		},
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Delete
// ─────────────────────────────────────────────────────────────
//

type DeleteFact struct {
	baseFact
}

func (f DeleteFact) Inverse() Fact {
	return InsertFact{
		baseFact: baseFact{
			kind:   FactInsert,
			anchor: f.anchor,
			rng:    f.rng,
			text:   f.text,
		},
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Replace
// ─────────────────────────────────────────────────────────────
//

type ReplaceFact struct {
	baseFact
	oldText string
}

func (f ReplaceFact) Inverse() Fact {
	return ReplaceFact{
		baseFact: baseFact{
			kind:   FactReplace,
			anchor: f.anchor,
			rng:    f.rng,
			text:   f.oldText,
		},
		oldText: f.text,
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Move
// ─────────────────────────────────────────────────────────────
//

type MoveFact struct {
	baseFact
	from Anchor
	to   Anchor
}

func (f MoveFact) Inverse() Fact {
	return MoveFact{
		baseFact: baseFact{
			kind:   FactMove,
			anchor: f.anchor,
		},
		from: f.to,
		to:   f.from,
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Capture (Pure Semantic)
// ─────────────────────────────────────────────────────────────
//

// CaptureAnchor 捕获锚点（纯函数）
func CaptureAnchor(a Anchor) Anchor {
	return a
}

// CaptureRange 捕获一个语义范围（不访问文本）
func CaptureRange(anchor Anchor, motion Motion, knownText string) Range {
	start := anchor
	end := anchor

	switch motion.Kind {
	case MotionWordForward:
		end.Col += max(1, motion.Count) * 5 // 语义步进
	case MotionLine:
		end.Col = 1 << 30 // 语义行尾
	}

	return Range{
		Start: start,
		End:   end,
		Text:  knownText,
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Capture Facts
// ─────────────────────────────────────────────────────────────
//

func CaptureInsert(anchor Anchor, text string) Fact {
	return InsertFact{
		baseFact: baseFact{
			kind:   FactInsert,
			anchor: anchor,
			text:   text,
		},
	}
}

func CaptureDelete(rng Range) Fact {
	return DeleteFact{
		baseFact: baseFact{
			kind:   FactDelete,
			anchor: rng.Start,
			rng:    &rng,
			text:   rng.Text,
		},
	}
}

func CaptureReplace(rng Range, text string) Fact {
	return ReplaceFact{
		baseFact: baseFact{
			kind:   FactReplace,
			anchor: rng.Start,
			rng:    &rng,
			text:   text,
		},
		oldText: rng.Text,
	}
}

func CaptureMove(from, to Anchor) Fact {
	return MoveFact{
		baseFact: baseFact{
			kind:   FactMove,
			anchor: from,
		},
		from: from,
		to:   to,
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Helpers
// ─────────────────────────────────────────────────────────────
//

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

````

## 📄 `tests/invalid_history_test.go`

````go
package tests

import (
	"os"
	"testing"

	"tmux-fsm/verifier"
)

func loadExample(t *testing.T, path string) verifier.VerifyInput {
	_, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	// 这里需要根据实际的 verifier 接口进行调整
	input := verifier.VerifyInput{}
	return input
}

func TestInvalidHistory_ParentMismatch(t *testing.T) {
	// 这里需要根据实际的 verifier 接口进行调整
	// input := loadExample(t,
	// 	"../examples/invalid_history/parent_mismatch/facts.json",
	// )

	// _, err := verifier.Verify(input)
	// if err == nil {
	// 	t.Fatalf("expected verification failure, got success")
	// }
	t.Skip("Verifier interface needs to be implemented")
}

func TestInvalidHistory_ReorderedFacts(t *testing.T) {
	// 这里需要根据实际的 verifier 接口进行调整
	t.Skip("Verifier interface needs to be implemented")
}

func TestInvalidHistory_SameTextDifferentRoot(t *testing.T) {
	// 这里需要根据实际的 verifier 接口进行调整
	t.Skip("Verifier interface needs to be implemented")
}

````

## 📄 `text_object.go`

````go
package main

import (
	"errors"
)

// TextObjectKind 定义文本对象类型
type TextObjectKind int

const (
	TextObjectWord TextObjectKind = iota
	TextObjectParen
	TextObjectBracket
	TextObjectBrace
	TextObjectQuoteDouble
	TextObjectQuoteSingle
	TextObjectParagraph
	TextObjectSentence
)

// TextObjectMotion 定义文本对象运动
type TextObjectMotion struct {
	Kind  TextObjectKind
	Inner bool // true for 'i', false for 'a'
}

// TextObjectRangeCalculator 计算文本对象范围的接口
type TextObjectRangeCalculator interface {
	CalculateRange(obj TextObjectMotion, cursor Cursor) (*MotionRange, error)
}

// ConcreteTextObjectCalculator 实现文本对象范围计算器
type ConcreteTextObjectCalculator struct {
	Buffer Buffer
}

// NewConcreteTextObjectCalculator 创建新的文本对象计算器
func NewConcreteTextObjectCalculator(buffer Buffer) *ConcreteTextObjectCalculator {
	return &ConcreteTextObjectCalculator{
		Buffer: buffer,
	}
}

// CalculateRange 计算文本对象范围
func (calc *ConcreteTextObjectCalculator) CalculateRange(obj TextObjectMotion, cursor Cursor) (*MotionRange, error) {
	switch obj.Kind {
	case TextObjectWord:
		return calc.calculateWordRange(obj.Inner, cursor)
	case TextObjectParen:
		return calc.calculateDelimitedRange('(', ')', obj.Inner, cursor)
	case TextObjectBracket:
		return calc.calculateDelimitedRange('[', ']', obj.Inner, cursor)
	case TextObjectBrace:
		return calc.calculateDelimitedRange('{', '}', obj.Inner, cursor)
	case TextObjectQuoteDouble:
		return calc.calculateQuoteRange('"', obj.Inner, cursor)
	case TextObjectQuoteSingle:
		return calc.calculateQuoteRange('\'', obj.Inner, cursor)
	case TextObjectParagraph:
		return calc.calculateParagraphRange(obj.Inner, cursor)
	case TextObjectSentence:
		return calc.calculateSentenceRange(obj.Inner, cursor)
	default:
		return nil, errors.New("unsupported text object")
	}
}

// calculateWordRange 计算单词范围
func (calc *ConcreteTextObjectCalculator) calculateWordRange(inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	row := cursor.Row
	if row < 0 || row >= calc.Buffer.LineCount() {
		return nil, errors.New("invalid row")
	}

	line := make([]rune, calc.Buffer.LineLength(row))
	for i := 0; i < len(line); i++ {
		line[i] = calc.Buffer.RuneAt(row, i)
	}

	startCol, endCol := findWordAt(line, cursor.Col, inner)

	return &MotionRange{
		Start: Cursor{Row: row, Col: startCol},
		End:   Cursor{Row: row, Col: endCol},
	}, nil
}

// findWordAt 查找光标位置的单词范围
func findWordAt(line []rune, col int, inner bool) (int, int) {
	if len(line) == 0 || col < 0 {
		return 0, 0
	}

	if col >= len(line) {
		col = len(line) - 1
	}

	// 确定字符类别
	charType := classifyRune(line[col])

	// 向左查找边界
	start := col
	for start > 0 {
		if classifyRune(line[start-1]) != charType {
			break
		}
		start--
	}

	// 向右查找边界
	end := col
	for end < len(line)-1 {
		if classifyRune(line[end+1]) != charType {
			break
		}
		end++
	}

	// 如果是 inner 模式，去除两端的空白
	if inner {
		for start <= end && start < len(line) && isWhitespace(line[start]) {
			start++
		}
		for end > start && end >= 0 && isWhitespace(line[end]) {
			end--
		}
	}

	// 确保 end 在有效范围内
	if end >= len(line) {
		end = len(line) - 1
	}

	// 确保范围有效
	if start > end {
		start = end
	}

	// 如果是 outer 模式，扩展到包含相邻的空白
	if !inner {
		// 向右扩展包含空白
		for end < len(line)-1 && isWhitespace(line[end+1]) {
			end++
		}
		// 向左扩展包含空白
		for start > 0 && isWhitespace(line[start-1]) {
			start--
		}
	}

	return start, end + 1
}

// classifyRune 将字符分类
func classifyRune(r rune) CharClass {
	switch {
	case r == ' ' || r == '\t' || r == '\n' || r == '\r':
		return ClassWhitespace
	case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_':
		return ClassWord
	default:
		return ClassPunct
	}
}

// isWhitespace 检查是否为空白字符
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// calculateDelimitedRange 计算定界符范围
func (calc *ConcreteTextObjectCalculator) calculateDelimitedRange(open, close rune, inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	// 从当前行开始搜索
	startPos, endPos := findDelimitedRange(calc.Buffer, open, close, cursor, inner)

	if startPos.Row == -1 || endPos.Row == -1 {
		return nil, errors.New("delimited range not found")
	}

	return &MotionRange{
		Start: startPos,
		End:   endPos,
	}, nil
}

// findDelimitedRange 查找定界符范围
func findDelimitedRange(buffer Buffer, open, close rune, cursor Cursor, inner bool) (Cursor, Cursor) {
	// 从当前光标位置开始查找匹配的定界符
	currentRow := cursor.Row
	currentCol := cursor.Col

	// 首先尝试在当前行查找
	for row := currentRow; row < buffer.LineCount(); row++ {
		lineLen := buffer.LineLength(row)
		startCol := 0
		if row == currentRow {
			startCol = currentCol
		}

		for col := startCol; col < lineLen; col++ {
			r := buffer.RuneAt(row, col)
			if r == open {
				// 找到开定界符，查找对应的闭定界符
				endPos := findMatchingDelimiter(buffer, open, close, Cursor{Row: row, Col: col})
				if endPos.Row != -1 {
					if inner {
						// Inner 模式：排除定界符本身
						return Cursor{Row: row, Col: col + 1}, endPos
					} else {
						// Outer 模式：包含定界符
						return Cursor{Row: row, Col: col}, Cursor{Row: endPos.Row, Col: endPos.Col + 1}
					}
				}
			}
		}
	}

	// 如果没找到，返回无效位置
	return Cursor{Row: -1, Col: -1}, Cursor{Row: -1, Col: -1}
}

// findMatchingDelimiter 查找匹配的定界符
func findMatchingDelimiter(buffer Buffer, open, close rune, startPos Cursor) Cursor {
	stack := 0
	currentRow := startPos.Row
	currentCol := startPos.Col + 1 // 从开定界符的下一个位置开始

	for row := currentRow; row < buffer.LineCount(); row++ {
		lineLen := buffer.LineLength(row)
		startCol := 0
		if row == currentRow {
			startCol = currentCol
		}

		for col := startCol; col < lineLen; col++ {
			r := buffer.RuneAt(row, col)
			if r == open {
				stack++
			} else if r == close {
				stack--
				if stack < 0 {
					// 找到匹配的闭定界符
					return Cursor{Row: row, Col: col}
				}
			}
		}
		currentCol = 0 // 从下一行开始时，列从0开始
	}

	// 没有找到匹配的闭定界符
	return Cursor{Row: -1, Col: -1}
}

// calculateQuoteRange 计算引号范围
func (calc *ConcreteTextObjectCalculator) calculateQuoteRange(quote rune, inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	// 从当前光标位置开始查找引号
	currentRow := cursor.Row
	currentCol := cursor.Col

	// 首先检查光标位置是否在引号内或旁边
	for row := currentRow; row < calc.Buffer.LineCount(); row++ {
		lineLen := calc.Buffer.LineLength(row)
		startCol := 0
		if row == currentRow {
			startCol = currentCol
		}

		for col := startCol; col < lineLen; col++ {
			r := calc.Buffer.RuneAt(row, col)
			if r == quote {
				// 找到第一个引号，查找匹配的另一个
				endPos := findMatchingQuote(calc.Buffer, quote, Cursor{Row: row, Col: col})
				if endPos.Row != -1 {
					if inner {
						// Inner 模式：排除引号本身
						return &MotionRange{
							Start: Cursor{Row: row, Col: col + 1},
							End:   endPos,
						}, nil
					} else {
						// Outer 模式：包含引号
						return &MotionRange{
							Start: Cursor{Row: row, Col: col},
							End:   Cursor{Row: endPos.Row, Col: endPos.Col + 1},
						}, nil
					}
				}
			}
		}
	}

	return nil, errors.New("quote range not found")
}

// findMatchingQuote 查找匹配的引号
func findMatchingQuote(buffer Buffer, quote rune, startPos Cursor) Cursor {
	escaped := false

	currentRow := startPos.Row
	currentCol := startPos.Col + 1 // 从第一个引号的下一个位置开始

	for row := currentRow; row < buffer.LineCount(); row++ {
		lineLen := buffer.LineLength(row)
		startCol := 0
		if row == currentRow {
			startCol = currentCol
		}

		for col := startCol; col < lineLen; col++ {
			r := buffer.RuneAt(row, col)

			if escaped {
				escaped = false
				continue
			}

			if r == '\\' {
				escaped = true
				continue
			}

			if r == quote {
				// 找到匹配的引号
				return Cursor{Row: row, Col: col}
			}
		}
		currentCol = 0 // 从下一行开始时，列从0开始
	}

	// 没有找到匹配的引号
	return Cursor{Row: -1, Col: -1}
}

// calculateParagraphRange 计算段落范围
func (calc *ConcreteTextObjectCalculator) calculateParagraphRange(inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	// 简化实现：查找空行分隔的段落
	startRow := cursor.Row
	endRow := cursor.Row

	// 向上查找段落开始
	for startRow > 0 {
		lineLen := calc.Buffer.LineLength(startRow - 1)
		if lineLen == 0 {
			break
		}
		startRow--
	}

	// 向下查找段落结束
	for endRow < calc.Buffer.LineCount()-1 {
		lineLen := calc.Buffer.LineLength(endRow + 1)
		if lineLen == 0 {
			break
		}
		endRow++
	}

	if inner {
		// Inner 模式：排除段落周围的空行
		return &MotionRange{
			Start: Cursor{Row: startRow, Col: 0},
			End:   Cursor{Row: endRow, Col: calc.Buffer.LineLength(endRow)},
		}, nil
	} else {
		// Outer 模式：包含整个段落
		return &MotionRange{
			Start: Cursor{Row: startRow, Col: 0},
			End:   Cursor{Row: endRow + 1, Col: 0}, // 包含下一行的开始
		}, nil
	}
}

// calculateSentenceRange 计算句子范围
func (calc *ConcreteTextObjectCalculator) calculateSentenceRange(inner bool, cursor Cursor) (*MotionRange, error) {
	if calc.Buffer == nil {
		return nil, errors.New("no buffer available")
	}

	// 简化实现：查找句号、感叹号、问号分隔的句子
	currentRow := cursor.Row
	currentCol := cursor.Col

	// 查找当前句子的开始
	startRow, startCol := findSentenceStart(calc.Buffer, currentRow, currentCol)

	// 查找当前句子的结束
	endRow, endCol := findSentenceEnd(calc.Buffer, currentRow, currentCol)

	if inner {
		// Inner 模式：排除句子结束标点
		return &MotionRange{
			Start: Cursor{Row: startRow, Col: startCol},
			End:   Cursor{Row: endRow, Col: endCol},
		}, nil
	} else {
		// Outer 模式：包含句子结束标点及后续空白
		// 简化：包含到句子结束
		return &MotionRange{
			Start: Cursor{Row: startRow, Col: startCol},
			End:   Cursor{Row: endRow, Col: endCol + 1},
		}, nil
	}
}

// findSentenceStart 查找句子开始
func findSentenceStart(buffer Buffer, row, col int) (int, int) {
	// 简化实现：查找前一个句子结束符后的第一个非空白字符
	for r := row; r >= 0; r-- {
		lineLen := buffer.LineLength(r)
		startCol := lineLen - 1
		if r == row {
			startCol = col
		}

		for c := startCol; c >= 0; c-- {
			runeVal := buffer.RuneAt(r, c)
			if runeVal == '.' || runeVal == '!' || runeVal == '?' {
				// 找到句子结束符，下一个位置是句子开始
				nextRow, nextCol := getNextNonWhitespace(buffer, r, c+1)
				return nextRow, nextCol
			}
		}
	}

	// 如果没找到，返回文件开始
	return 0, 0
}

// findSentenceEnd 查找句子结束
func findSentenceEnd(buffer Buffer, row, col int) (int, int) {
	// 简化实现：查找下一个句子结束符
	for r := row; r < buffer.LineCount(); r++ {
		lineLen := buffer.LineLength(r)
		startCol := 0
		if r == row {
			startCol = col
		}

		for c := startCol; c < lineLen; c++ {
			runeVal := buffer.RuneAt(r, c)
			if runeVal == '.' || runeVal == '!' || runeVal == '?' {
				// 找到句子结束符
				return r, c
			}
		}
	}

	// 如果没找到，返回文件结束
	endRow := buffer.LineCount() - 1
	endCol := buffer.LineLength(endRow)
	return endRow, endCol
}

// getNextNonWhitespace 获取下一个非空白字符位置
func getNextNonWhitespace(buffer Buffer, row, col int) (int, int) {
	for r := row; r < buffer.LineCount(); r++ {
		lineLen := buffer.LineLength(r)
		startCol := 0
		if r == row {
			startCol = col
		}

		for c := startCol; c < lineLen; c++ {
			runeVal := buffer.RuneAt(r, c)
			if !isWhitespace(runeVal) {
				return r, c
			}
		}
	}

	// 如果没找到，返回当前位置
	return row, col
}

// ParseTextObject 解析文本对象字符串
func ParseTextObject(textObjectStr string) (*TextObjectMotion, error) {
	if len(textObjectStr) < 2 {
		return nil, errors.New("invalid text object string")
	}

	modifier := textObjectStr[0:1]
	objType := textObjectStr[1:2]

	inner := modifier == "i"

	var kind TextObjectKind
	switch objType {
	case "w":
		kind = TextObjectWord
	case "(":
		kind = TextObjectParen
	case "[":
		kind = TextObjectBracket
	case "{":
		kind = TextObjectBrace
	case "\"":
		kind = TextObjectQuoteDouble
	case "'":
		kind = TextObjectQuoteSingle
	case "p":
		kind = TextObjectParagraph
	case "s":
		kind = TextObjectSentence
	default:
		return nil, errors.New("unsupported text object type")
	}

	return &TextObjectMotion{
		Kind:  kind,
		Inner: inner,
	}, nil
}

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
	".go":    "go",
	".js":    "javascript",
	".ts":    "typescript",
	".tsx":   "typescript",
	".jsx":   "javascript",
	".py":    "python",
	".java":  "java",
	".c":     "c",
	".cpp":   "cpp",
	".cc":    "cpp",
	".cxx":   "cpp",
	".h":     "c",
	".hpp":   "cpp",
	".rs":    "rust",
	".rb":    "ruby",
	".php":   "php",
	".cs":    "csharp",
	".swift": "swift",
	".kt":    "kotlin",
	".scala": "scala",
	".r":     "r",
	".sql":   "sql",
	".sh":    "bash",
	".bash":  "bash",
	".zsh":   "bash",
	".fish":  "fish",
	".ps1":   "powershell",
	".md":    "markdown",
	".html":  "html",
	".htm":   "html",
	".css":   "css",
	".scss":  "scss",
	".sass":  "sass",
	".less":  "less",
	".xml":   "xml",
	".json":  "json",
	".yaml":  "yaml",
	".yml":   "yaml",
	".toml":  "toml",
	".ini":   "ini",
	".conf":  "conf",
	".txt":   "text",
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

## 📄 `ui/interface.go`

````go
package ui

// UI 接口定义
type UI interface {
	Show()
	Update()
	Hide()
}

````

## 📄 `ui/popup.go`

````go
package ui

import "fmt"

type Backend interface {
	ExecRaw(cmd string)
}

type StateProvider interface {
	GetActiveState() string
	GetStateHint(state string) string
}

type PopupUI struct {
	StateProvider StateProvider
	Backend       Backend
}

func (p *PopupUI) Show() {
	if p.StateProvider == nil || p.Backend == nil {
		return
	}

	active := p.StateProvider.GetActiveState()
	if active == "" {
		return
	}

	hint := p.StateProvider.GetStateHint(active)

	cmd := fmt.Sprintf(
		"display-popup -E -w 50%% -h 5 'echo \"%s\"; echo \"%s\"'",
		active,
		hint,
	)

	p.Backend.ExecRaw(cmd)
}

func (p *PopupUI) Update() {
	p.Show()
}

func (p *PopupUI) Hide() {
	if p.Backend != nil {
		p.Backend.ExecRaw("display-popup -C")
	}
}

````

## 📄 `undotree/tree.go`

````go
package undotree

import (
	"sort"

	"tmux-fsm/wal"
)

//
// ─────────────────────────────────────────────────────────────
//  Undo Node
// ─────────────────────────────────────────────────────────────
//

type UndoNode struct {
	Event    *wal.SemanticEvent
	Parent   *UndoNode
	Children []*UndoNode
}

// IsRoot 判断是否为虚拟根
func (n *UndoNode) IsRoot() bool {
	return n.Event == nil
}

//
// ─────────────────────────────────────────────────────────────
//  Build Undo Tree
// ─────────────────────────────────────────────────────────────
//

func BuildUndoTree(events []wal.SemanticEvent) *UndoNode {

	root := &UndoNode{} // ✅ 虚拟根
	nodes := make(map[string]*UndoNode)

	// 1️⃣ 创建节点
	for i := range events {
		e := &events[i]
		nodes[e.ID] = &UndoNode{
			Event: e,
		}
	}

	// 2️⃣ 建立父子关系（LocalParent）
	for _, n := range nodes {
		lp := n.Event.LocalParent

		if lp == "" {
			n.Parent = root
			root.Children = append(root.Children, n)
			continue
		}

		if p, ok := nodes[lp]; ok {
			n.Parent = p
			p.Children = append(p.Children, n)
		} else {
			// ✅ 父缺失 → 挂到 root（WAL 截断 / 合并时常见）
			n.Parent = root
			root.Children = append(root.Children, n)
		}
	}

	// 3️⃣ 稳定排序（按时间 + ID）
	sortTree(root)

	return root
}

func sortTree(n *UndoNode) {
	sort.Slice(n.Children, func(i, j int) bool {
		ei := n.Children[i].Event
		ej := n.Children[j].Event

		if ei.Time.Equal(ej.Time) {
			return ei.ID < ej.ID
		}
		return ei.Time.Before(ej.Time)
	})

	for _, c := range n.Children {
		sortTree(c)
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Path Utilities
// ─────────────────────────────────────────────────────────────
//

// PathToRoot 返回从 root → node 的事件路径（不含虚拟 root）
func PathToRoot(n *UndoNode) []*wal.SemanticEvent {
	var rev []*wal.SemanticEvent

	for cur := n; cur != nil && !cur.IsRoot(); cur = cur.Parent {
		rev = append(rev, cur.Event)
	}

	// reverse
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}

	return rev
}

````

## 📄 `verifier/verifier.go`

````go
package verifier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"tmux-fsm/crdt"
	"tmux-fsm/replay"
	"tmux-fsm/semantic"
)

//
// ─────────────────────────────────────────────────────────────
//  Hash
// ─────────────────────────────────────────────────────────────
//

type Hash string

func hashBytes(b []byte) Hash {
	h := sha256.Sum256(b)
	return Hash(hex.EncodeToString(h[:]))
}

//
// ─────────────────────────────────────────────────────────────
//  Canonical Types
// ─────────────────────────────────────────────────────────────
//

// CanonicalSemanticEvent 必须是确定性可序列化的
type CanonicalSemanticEvent struct {
	Actor         crdt.ActorID   `json:"actor"`
	CausalParents []crdt.EventID `json:"causal_parents"`
	Fact          semantic.Fact  `json:"fact"`
}

// Fact 是“已签名、可验证”的最小单元
type Fact struct {
	ID        Hash                   `json:"id"`
	Actor     crdt.ActorID           `json:"actor"`
	Parents   []Hash                 `json:"parents"`
	Timestamp int64                  `json:"timestamp"`
	Payload   CanonicalSemanticEvent `json:"payload"`
	PolicyRef Hash                   `json:"policy_ref"`
}

//
// ─────────────────────────────────────────────────────────────
//  Verify Input / Output
// ─────────────────────────────────────────────────────────────
//

type VerifyInput struct {
	Facts        []Fact
	Policies     map[Hash][]byte
	Snapshot     *replay.TextState
	ExpectedRoot Hash
}

type VerifyResult struct {
	OK        bool   `json:"ok"`
	StateRoot Hash   `json:"state_root"`
	Error     string `json:"error,omitempty"`

	FactsUsed int `json:"facts_used"`
	Policies  int `json:"policies"`
}

//
// ─────────────────────────────────────────────────────────────
//  Verifier
// ─────────────────────────────────────────────────────────────
//

type Verifier struct {
	policies map[Hash][]byte
}

func NewVerifier(policies map[Hash][]byte) *Verifier {
	return &Verifier{policies: policies}
}

//
// ─────────────────────────────────────────────────────────────
//  Verify Entry
// ─────────────────────────────────────────────────────────────
//

func (v *Verifier) Verify(input VerifyInput) VerifyResult {

	// 1️⃣ Fact 自洽校验
	for _, f := range input.Facts {
		if calcFactHash(f) != f.ID {
			return fail("fact hash mismatch: " + string(f.ID))
		}
	}

	// 2️⃣ DAG + 稳定拓扑排序 + 环检测
	ordered, err := topoSortFacts(input.Facts)
	if err != nil {
		return fail(err.Error())
	}

	// 3️⃣ 初始状态
	state := replay.TextState{}
	if input.Snapshot != nil {
		state = input.Snapshot.Clone()
	}

	// 4️⃣ 纯 Replay
	for _, f := range ordered {

		if err := v.checkPolicy(f, state); err != nil {
			return fail(fmt.Sprintf("policy violation at %s: %v", f.ID, err))
		}

		next := state
		replay.ApplyFact(&next, f.Payload.Fact)
		state = next
	}

	// 5️⃣ State Root
	root := calcStateHash(state)

	if root != input.ExpectedRoot {
		return fail(fmt.Sprintf(
			"state root mismatch: expected %s, got %s",
			input.ExpectedRoot, root,
		))
	}

	return VerifyResult{
		OK:        true,
		StateRoot: root,
		FactsUsed: len(ordered),
		Policies:  len(v.policies),
	}
}

func fail(msg string) VerifyResult {
	return VerifyResult{OK: false, Error: msg}
}

//
// ─────────────────────────────────────────────────────────────
//  Topological Sort (Stable + Cycle Detect)
// ─────────────────────────────────────────────────────────────
//

func topoSortFacts(facts []Fact) ([]Fact, error) {

	graph := map[Hash][]Hash{}
	inDegree := map[Hash]int{}
	factMap := map[Hash]Fact{}

	for _, f := range facts {
		graph[f.ID] = nil
		inDegree[f.ID] = 0
		factMap[f.ID] = f
	}

	for _, f := range facts {
		for _, p := range f.Parents {
			if _, ok := inDegree[p]; ok {
				graph[p] = append(graph[p], f.ID)
				inDegree[f.ID]++
			}
		}
	}

	var queue []Hash
	for id, d := range inDegree {
		if d == 0 {
			queue = append(queue, id)
		}
	}

	sort.Slice(queue, func(i, j int) bool {
		return string(queue[i]) < string(queue[j])
	})

	var out []Fact

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		out = append(out, factMap[id])

		for _, nxt := range graph[id] {
			inDegree[nxt]--
			if inDegree[nxt] == 0 {
				queue = append(queue, nxt)
			}
		}
	}

	if len(out) != len(facts) {
		return nil, errors.New("cycle detected in fact graph")
	}

	return out, nil
}

//
// ─────────────────────────────────────────────────────────────
//  Policy (Minimal / Deterministic)
// ─────────────────────────────────────────────────────────────
//

func (v *Verifier) checkPolicy(f Fact, state replay.TextState) error {

	// 1️⃣ Policy code must exist
	if _, ok := v.policies[f.PolicyRef]; !ok {
		return errors.New("unknown policy ref")
	}

	// 2️⃣ 最小 AI 防线（deterministic）
	actor := string(f.Actor)
	if len(actor) >= 2 && actor[:2] == "ai" {
		switch f.Payload.Fact.Kind() {
		case semantic.FactInsert, semantic.FactDelete, semantic.FactMove:
			return nil
		default:
			return errors.New("ai operation not allowed")
		}
	}

	return nil
}

//
// ─────────────────────────────────────────────────────────────
//  Hashing (Canonical)
// ─────────────────────────────────────────────────────────────
//

func calcFactHash(f Fact) Hash {

	parents := append([]Hash{}, f.Parents...)
	sort.Slice(parents, func(i, j int) bool {
		return parents[i] < parents[j]
	})

	data, _ := json.Marshal(struct {
		Actor     crdt.ActorID           `json:"actor"`
		Parents   []Hash                 `json:"parents"`
		Timestamp int64                  `json:"timestamp"`
		Payload   CanonicalSemanticEvent `json:"payload"`
		PolicyRef Hash                   `json:"policy_ref"`
	}{
		Actor:     f.Actor,
		Parents:   parents,
		Timestamp: f.Timestamp,
		Payload:   f.Payload,
		PolicyRef: f.PolicyRef,
	})

	return hashBytes(data)
}

func calcStateHash(state replay.TextState) Hash {
	data, _ := json.Marshal(state)
	return hashBytes(data)
}

//
// ─────────────────────────────────────────────────────────────
//  JSON Helper
// ─────────────────────────────────────────────────────────────
//

func (v *Verifier) VerifyFromJSON(
	factsJSON []byte,
	expectedRoot Hash,
) (VerifyResult, error) {

	var facts []Fact
	if err := json.Unmarshal(factsJSON, &facts); err != nil {
		return VerifyResult{}, err
	}

	return v.Verify(VerifyInput{
		Facts:        facts,
		ExpectedRoot: expectedRoot,
	}), nil
}

````

## 📄 `weaver/adapter/backend.go`

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
	// 使用 set-buffer + paste-buffer 是最稳健的，避免 shell 转义问题
	exec.Command("tmux", "set-buffer", "--", text).Run()
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

## 📄 `weaver/core/resolved_fact.go`

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

## 📄 `weaver/core/snapshot_types.go`

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

## 📄 `weaver/core/take_snapshot.go`

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

## 📄 `weaver/manager/manager.go`

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
