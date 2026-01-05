package main

import (
	"fmt"
	"os"
	"time"
	"tmux-fsm/weaver/adapter"
	"tmux-fsm/weaver/core"
	"tmux-fsm/weaver/logic"
)

// WeaverManager 全局协调器
type WeaverManager struct {
	mode       ExecutionMode
	engine     core.Engine
	resolver   core.AnchorResolver
	projection core.Projection
}

// weaverMgr 全局 Weaver 实例
var weaverMgr *WeaverManager

// InitWeaver 初始化 Weaver 系统
func InitWeaver(mode ExecutionMode) {
	if mode == ModeLegacy {
		return
	}

	// 初始化组件
	// 1. Planner (Logic)
	planner := &logic.ShellFactBuilder{}

	// 2. Engine (Core)
	engine := core.NewShadowEngine(planner)

	// 3. Adapter (Environment)
	// 根据模式选择 Projection
	var proj core.Projection
	if mode == ModeWeaver {
		proj = &adapter.TmuxProjection{}
	} else {
		proj = &adapter.NoopProjection{}
	}

	// Resolver 目前还是 Noop
	resolver := &adapter.NoopResolver{}

	weaverMgr = &WeaverManager{
		mode:       mode,
		engine:     engine,
		resolver:   resolver,
		projection: proj,
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

// ProcessIntent 处理 Intent
func (m *WeaverManager) ProcessIntent(intent Intent) {
	// 1. 适配 Intent
	coreIntent := &intentAdapter{intent: intent}

	// 2. 调用 Engine (Apply)
	verdict, err := m.engine.ApplyIntent(coreIntent, m.resolver, m.projection)

	if err != nil {
		logWeaver("Error applying intent: %v", err)
		return
	}

	// 3. 日志记录
	if config.LogFacts {
		txID := "nil"
		safety := core.SafetyExact
		if verdict.Transaction != nil {
			txID = string(verdict.Transaction.ID)
			safety = verdict.Transaction.Safety
		}
		logWeaver("Verdict: %s (tx: %s) (Safety: %v)", verdict.Message, txID, safety)
	}

	// 4. Undo 注入 (Phase 3 关键：Legacy Undo Bridge)
	if m.mode == ModeWeaver && verdict.Transaction != nil && len(verdict.Transaction.Facts) > 0 {
		record := convertToLegacyRecord(verdict.Transaction)
		if record != nil {
			stateMu.Lock()
			// globalState 是 main 包的全局变量
			globalState.transMgr.Append(*record)
			stateMu.Unlock()

			logWeaver("Injected Legacy ActionRecord for tx: %s", verdict.Transaction.ID)
		}
	}
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

// logWeaverInfo 记录 Weaver 信息日志
func logWeaverInfo(msg string) {
	f, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] [WEAVER] %s\n", time.Now().Format("15:04:05"), msg)
		f.Close()
	}
}

// logWeaverError 记录 Weaver 错误日志
func logWeaverError(msg string) {
	f, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] [WEAVER-ERROR] %s\n", time.Now().Format("15:04:05"), msg)
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
