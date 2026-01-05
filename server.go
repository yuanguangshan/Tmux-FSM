package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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