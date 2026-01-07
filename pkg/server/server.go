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