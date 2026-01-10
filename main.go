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
		// Try parsing the new format: "requestID|paneID|clientName|key"
		parts := strings.SplitN(payloadStr, "|", 4)
		var requestID, actorID, paneID, clientName, key string

		if len(parts) == 4 {
			// New format: requestID|paneID|clientName|key
			requestID = parts[0]
			paneID = parts[1]
			clientName = parts[2]
			key = parts[3]

			// Construct actorID from paneID and clientName
			actorID = fmt.Sprintf("%s|%s", paneID, clientName)
		} else if len(parts) == 3 {
			// Legacy format: actorID|pane|key (based on log examples)
			actorID = parts[0]
			paneID = parts[1]
			key = parts[2]

			// Extract clientName from actorID if possible
			actorParts := strings.SplitN(actorID, "|", 2)
			if len(actorParts) == 2 {
				paneID = actorParts[0]
				clientName = actorParts[1]
			} else {
				clientName = "unknown"
			}

			// Generate default requestID for backward compatibility
			requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
		} else if len(parts) == 2 {
			// Fallback for old protocol: PANE|KEY (Client unknown)
			paneID = parts[0]
			key = parts[1]

			// Generate default requestID and actorID for backward compatibility
			requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
			clientName = "unknown"
			actorID = fmt.Sprintf("%s|%s", paneID, clientName)
		} else {
			key = payloadStr
			// Generate default requestID and actorID for backward compatibility
			requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
			paneID = "default"
			clientName = "default"
			actorID = fmt.Sprintf("%s|%s", paneID, clientName)
		}

		log.Printf("[server] string protocol received: requestID='%s', actorID='%s', pane='%s', client='%s', key='%s'", requestID, actorID, paneID, clientName, key)

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

		// 使用 kernel 处理按键 with context containing identity anchors
		if kernelInstance != nil {
			hctx := kernel.HandleContext{
				Ctx:       context.Background(),
				RequestID: requestID,
				ActorID:   actorID,
			}
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
