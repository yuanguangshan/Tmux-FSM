package main

import (
	"context"
	"flag"
	"log"
	"time"

	"tmux-fsm/server"
	"tmux-fsm/weaver/core"
	"tmux-fsm/weaver/manager"
)

// weaverMgr 全局 Weaver 实例
var weaverMgr *manager.WeaverManager

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
	flag.Parse()

	// 初始化 Weaver 系统
	manager.InitWeaver(manager.ModeWeaver) // 默认启用 Weaver 模式

	if *serverMode {
		srv := server.New(server.Config{
			SocketPath: *socketPath,
		})
		log.Fatal(srv.Run(context.Background()))
		return
	}

	// client / other modes 保持你原来的逻辑
	log.Println("no mode specified")
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

func (a *intentAdapter) GetAnchors() []core.Anchor {
	return a.intent.GetAnchors()
}

// ProcessIntentGlobal 全局意图处理入口
// RFC-WC-002: Intent ABI - 统一入口，统一审计
func ProcessIntentGlobal(intent Intent) error {
	// 如果 weaverMgr 未初始化，返回
	if weaverMgr == nil {
		return nil
	}

	// 使用 weaver manager 处理意图
	return weaverMgr.ProcessIntentGlobal(&intentAdapter{intent: intent})
}
