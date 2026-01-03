package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

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
	Inverse Fact `inverse:"fact"`
}

type TransactionID uint64

type Transaction struct {
	ID          TransactionID  `json:"id"`
	Records     []ActionRecord `json:"records"`
	CreatedAt   time.Time      `json:"created_at"`
	Applied     bool           `json:"applied"`
	Skipped     bool           `json:"skipped"`
	SafetyLevel string         `json:"safety_level,omitempty"` // exact, fuzzy
}

type TransactionManager struct {
	current *Transaction
	nextID  TransactionID
}

func (tm *TransactionManager) Begin() {
	tm.current = &Transaction{
		ID:        tm.nextID,
		CreatedAt: time.Now(),
		Records:   []ActionRecord{},
	}
	tm.nextID++
}

func (tm *TransactionManager) Append(r ActionRecord) {
	if tm.current != nil {
		tm.current.Records = append(tm.current.Records, r)
	}
}

func (tm *TransactionManager) Commit(stack *[]Transaction) {
	if tm.current == nil || len(tm.current.Records) == 0 {
		tm.current = nil
		return
	}
	*stack = append(*stack, *tm.current)
	tm.current = nil
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
}

var (
	stateMu     sync.Mutex
	globalState FSMState
	transMgr    TransactionManager
	socketPath  = os.Getenv("HOME") + "/.tmux-fsm.sock"
)

func main() {
	serverMode := flag.Bool("server", false, "run as daemon server")
	stopServer := flag.Bool("stop", false, "stop the running daemon")
	flag.Parse()

	if *stopServer {
		shutdownServer()
		return
	}

	if *serverMode {
		runServer()
	} else {
		// Client mode
		key := ""
		paneID := ""

		// Parse arguments more robustly
		// Expect: tmux-fsm [key] [pane_id] OR flags
		// For backward compatibility and simplicity, let's look at Args
		args := flag.Args()
		if len(args) >= 1 {
			key = args[0]
		} else {
			key = os.Getenv("TMUX_FSM_KEY")
		}

		if len(args) >= 2 {
			paneID = args[1]
		} else {
			paneID = os.Getenv("TMUX_FSM_PANE")
		}

		if key == "" {
			return
		}
		// If paneID is empty, server might default to something or fail,
		// but we should try to send what we have.
		runClient(key, paneID)
	}
}

func runClient(key, paneID string) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: daemon not running. Start it with 'tmux-fsm -server'\n")
		return
	}
	defer conn.Close()

	payload := fmt.Sprintf("%s|%s", paneID, key)
	conn.Write([]byte(payload))

	// Read response (synchronize)
	buf, err := io.ReadAll(conn)
	if err != nil {
		return
	}
	// Only print if it's not the standard "ok" heartbeat
	if string(buf) != "ok" {
		os.Stdout.Write(buf)
	}
}

func runServer() {
	// 检查是否已有服务在运行 (且能响应)
	if conn, err := net.DialTimeout("unix", socketPath, 1*time.Second); err == nil {
		conn.Close()
		fmt.Println("Daemon already running and responsive.")
		return
	}

	// 如果 Socket 文件存在但无法连接，说明是残留文件，直接移除
	os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	defer listener.Close()
	os.Chmod(socketPath, 0666)

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
		exec.Command("tmux", "display-message", "PONG: Server is responsive").Run()
		return false
	}

	if key == "__CLEAR_STATE__" {
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

	// Process Key through FSM
	action := processKey(&globalState, key)
	// 添加调试日志
	logFile, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logFile != nil {
		fmt.Fprintf(logFile, "[%s] DEBUG: Key='%s', Action='%s', Mode='%s'\n", time.Now().Format("15:04:05"), key, action, globalState.Mode)
		if action != "" {
			fmt.Fprintf(logFile, "[%s] DEBUG: Executing action: %s\n", time.Now().Format("15:04:05"), action)
		}
		logFile.Close()
	}

	if action != "" {
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
					transMgr.Begin()
					orig := globalState.Count
					globalState.Count = countToUse
					executeAction(savedAction, &globalState, paneID, clientName)
					globalState.Count = orig
					transMgr.Commit(&globalState.UndoStack)
					return false
				}
			}
		} else {
			// Execute action wrapped in transaction
			// --- [ABI: Verdict Trigger] ---
			// Kernel begins deliberation for the given intent.
			transMgr.Begin()
			executeAction(action, &globalState, paneID, clientName)
			// --- [ABI: Audit Closure] ---
			// Kernel finalizes the verdict and commits to the timeline.
			transMgr.Commit(&globalState.UndoStack)

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
	updateStatusBar(globalState, clientName)
	conn.Write([]byte("ok"))
	return false
}

func shutdownServer() {
	conn, err := net.Dial("unix", socketPath)
	if err == nil {
		conn.Write([]byte("__SHUTDOWN__"))
		conn.Close()
	} else {
		fmt.Fprintf(os.Stderr, "Error: daemon not running to stop.\n")
	}
}

func loadState() FSMState {
	cmd := exec.Command("tmux", "show-option", "-gv", "@tmux_fsm_state")
	out, err := cmd.Output()
	var state FSMState
	if err != nil || len(out) == 0 {
		return FSMState{Mode: "NORMAL", Count: 0}
	}
	json.Unmarshal(out, &state)
	return state
}

func saveStateRaw(data []byte) {
	exec.Command("tmux", "set-option", "-g", "@tmux_fsm_state", string(data)).Run()
}

func updateStatusBar(state FSMState, clientName string) {
	modeMsg := state.Mode
	if modeMsg == "" {
		modeMsg = "NORMAL"
	} else if modeMsg == "VISUAL_CHAR" {
		modeMsg = "VISUAL"
	} else if modeMsg == "VISUAL_LINE" {
		modeMsg = "V-LINE"
	} else if modeMsg == "OPERATOR_PENDING" {
		modeMsg = "PENDING"
	} else if modeMsg == "REGISTER_SELECT" {
		modeMsg = "REGISTER"
	} else if modeMsg == "MOTION_PENDING" {
		modeMsg = "MOTION"
	} else if modeMsg == "SEARCH" {
		modeMsg = "SEARCH"
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
		// Axiom 8: Fuzzy Transparency - UI must explicitly notify the user
		keysMsg += " ~UNDO"
	} else if state.LastUndoFailure != "" {
		// Axiom 11: Explainability - Failures must be visible and explainable
		keysMsg += " !UNDO_FAIL"
	}

	// 调试日志
	f, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] Updating status: mode=%s, state.Mode=%s, keys=%s\n",
			time.Now().Format("15:04:05"), modeMsg, state.Mode, keysMsg)
		f.Close()
	}

	// 设置 tmux 变量 - 这是与 plugin.tmux 配合的关键
	exec.Command("tmux", "set-option", "-g", "@fsm_state", modeMsg).Run()
	exec.Command("tmux", "set-option", "-g", "@fsm_keys", keysMsg).Run()

	// 强制刷新所有客户端，确保状态栏更新
	exec.Command("tmux", "refresh-client", "-S").Run()

	// --- [ABI: Heartbeat Lock] ---
	// Re-assert the key table to prevent "one-shot" dropouts during run-shell.
	// We MUST check @fsm_active to allow intentional exits (e.g. 'c', 'i', Esc).
	if clientName != "" {
		out, _ := exec.Command("tmux", "show-option", "-gv", "@fsm_active").Output()
		if strings.TrimSpace(string(out)) == "true" {
			exec.Command("tmux", "switch-client", "-t", clientName, "-T", "fsm").Run()
		}
	}
}
