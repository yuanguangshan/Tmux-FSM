package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tmux-fsm/fsm"
	"tmux-fsm/intent"
	"tmux-fsm/kernel"
	"tmux-fsm/weaver/core"
	"tmux-fsm/weaver/manager"
)

// weaverMgr 全局 Weaver 实例
var weaverMgr *manager.WeaverManager

// kernelInstance 全局 Kernel 实例
var kernelInstance *kernel.Kernel

// TransactionManager 事务管理器
type TransactionManager struct {
	current *Transaction
	nextID  TransactionID
}

// Append 向事务管理器追加记录
func (tm *TransactionManager) Append(record ActionRecord) {
	if tm.current == nil {
		tm.current = &Transaction{
			ID:        tm.nextID,
			Records:   []ActionRecord{record},
			CreatedAt: time.Now(),
			Applied:   false,
			Skipped:   false,
		}
		tm.nextID++
	} else {
		tm.current.Records = append(tm.current.Records, record)
	}
}

func main() {
	serverMode := flag.Bool("server", false, "run as server")
	socketPath := flag.String("socket", "/tmp/tmux-fsm.sock", "socket path")
	debugMode := flag.Bool("debug", false, "enable debug logging")
	configPath := flag.String("config", "./keymap.yaml", "path to keymap configuration file")
	reloadFlag := flag.Bool("reload", false, "reload keymap configuration")
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

	// Initialize kernel with FSM engine
	kernelInstance = kernel.NewKernel(fsm.GetDefaultEngine(), nil) // Will set executor later

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

	if *serverMode {
		if *debugMode {
			log.Printf("[DEBUG] Starting server on %s", *socketPath)
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

	var in intent.Intent
	dec := json.NewDecoder(conn)

	if err := dec.Decode(&in); err != nil {
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
	}

	_ = ln.Close()
}

// intentAdapter 适配 intent.Intent 到 core.Intent
type intentAdapter struct {
	intent intent.Intent
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

func (a *intentAdapter) GetAnchors() []core.Anchor {
	// 将 intent.Anchor 转换为 core.Anchor
	anchors := a.intent.Anchors
	coreAnchors := make([]core.Anchor, len(anchors))
	for i, anchor := range anchors {
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

// ProcessIntentGlobal 全局意图处理入口
// RFC-WC-002: Intent ABI - 统一入口，统一审计
func ProcessIntentGlobal(intent intent.Intent) error {
	// 如果 weaverMgr 未初始化，返回
	if weaverMgr == nil {
		return nil
	}

	// 使用 weaver manager 处理意图
	return weaverMgr.ProcessIntentGlobal(&intentAdapter{intent: intent})
}
