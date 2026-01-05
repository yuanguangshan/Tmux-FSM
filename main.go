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
	"tmux-fsm/fsm"
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
	Inverse Fact `json:"inverse"`
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

	// [Phase 4] Reverse Bridge: Inject into Weaver History
	if weaverMgr != nil {
		weaverMgr.InjectLegacyTransaction(tm.current)
	}

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
	AllowPartial         bool                   `json:"allow_partial"` // Phase 7: Explicit permission for fuzzy resolution
}

var (
	stateMu     sync.Mutex
	globalState FSMState
	transMgr    TransactionManager
	socketPath  = os.Getenv("HOME") + "/.tmux-fsm.sock"
)

// isServerRunning 检查服务器是否已经在运行
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
			exec.Command(os.Args[0], "-server").Start()
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
					// 验证 clientName 是否为有效的 tmux client
					// Tmux 支持使用 TTY 路径作为 client target，所以不需要过滤 /dev/
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
		exec.Command("tmux", "set-option", "-g", "@fsm_active", "true").Run()
		if clientName != "" && clientName != "default" {
			exec.Command("tmux", "switch-client", "-t", clientName, "-T", "fsm").Run()
		} else {
			exec.Command("tmux", "switch-client", "-T", "fsm").Run()
		}
		exec.Command("tmux", "refresh-client", "-S").Run()

	case *exitFSM:
		// 直接通过 tmux 直接设置退出状态，保证响应速度
		exec.Command("tmux", "set-option", "-g", "@fsm_active", "false").Run()
		exec.Command("tmux", "set-option", "-g", "@fsm_state", "").Run()
		exec.Command("tmux", "set-option", "-g", "@fsm_keys", "").Run()
		exec.Command("tmux", "switch-client", "-T", "root").Run()
		exec.Command("tmux", "refresh-client", "-S").Run()
	case *dispatch != "":
		// 使用 Legacy 系统：将按键发送到服务器
		paneAndClient := ""
		if len(flag.Args()) > 0 {
			paneAndClient = flag.Args()[0]
		}
		runClient(*dispatch, paneAndClient)
	case *nvimMode != "":
		// Neovim 模式同步 (如果需要的话，也可以通过服务器同步)
		fsm.OnNvimMode(*nvimMode)
	case *uiShow:
		// 使用新的 FSM 系统
		fsm.ShowUI()
	case *uiHide:
		// 使用新的 FSM 系统
		fsm.HideUI()
	case *reload:
		// 使用新的 FSM 系统
		if err := fsm.LoadKeymap(configFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to reload keymap: %v\n", err)
			os.Exit(1)
		}
		fsm.UpdateUI()
	case *stopServer:
		shutdownServer()
	case *serverMode:
		runServer()
	default:
		// If key is empty but we were called, it might be a ghost trigger or #{key} fail.
		// Don't show usage manual as it interrupts user.
		if *dispatch == "" {
			return
		}

		// 如果没有参数，显示帮助
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

// 以下是原有的服务器模式代码
func runClient(key, paneID string) {
	// 添加参数验证
	if paneID == "" || paneID == "|" {
		// 尝试获取当前pane和client
		paneIDBytes, err1 := exec.Command("tmux", "display-message", "-p", "#{pane_id}").Output()
		clientNameBytes, err2 := exec.Command("tmux", "display-message", "-p", "#{client_name}").Output()
		if err1 == nil && err2 == nil {
			paneID = strings.TrimSpace(string(paneIDBytes)) + "|" + strings.TrimSpace(string(clientNameBytes))
		} else {
			paneID = "default|default"
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

	payload := fmt.Sprintf("%s|%s", paneID, key)
	if _, err := conn.Write([]byte(payload)); err != nil {
		return
	}

	// Read response (synchronize)
	buf, err := io.ReadAll(conn)
	if err != nil {
		return
	}
	// Only print if it's not the standard "ok" heartbeat
	resp := strings.TrimSpace(string(buf))
	if resp != "ok" && resp != "" {
		fmt.Println(resp)
	}
}

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
		// 将 action string 转换为 Intent
		intent := actionStringToIntent(action, globalState.Count, paneID)
		// 让 Weaver 处理（只记录，不执行）
		ProcessIntentGlobal(intent)
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
	}

	// 融合显示逻辑
	activeLayer := fsm.GetActiveLayer()
	if activeLayer != "NAV" && activeLayer != "" {
		// 如果处于新架构的层级（如 GOTO），覆盖显示
		modeMsg = activeLayer
	} else {
		// 转换老架构的模式名称用于显示
		if modeMsg == "VISUAL_CHAR" {
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
