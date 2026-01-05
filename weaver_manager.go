package main

import (
	"fmt"
	"os"
	"time"
	"tmux-fsm/weaver/adapter"
	"tmux-fsm/weaver/core"
)

// WeaverManager Weaver 管理器
// 负责协调 Weaver Core 和 Legacy 系统
type WeaverManager struct {
	engine  core.Engine
	adapter *adapter.TmuxAdapter
	enabled bool
}

// globalWeaver 全局 Weaver 实例
var globalWeaver *WeaverManager

// InitWeaver 初始化 Weaver
func InitWeaver() {
	mode := GetMode()
	
	// 只在 Shadow 或 Weaver 模式下初始化
	if mode == ModeLegacy {
		globalWeaver = nil
		return
	}

	globalWeaver = &WeaverManager{
		engine:  core.NewShadowEngine(),
		adapter: adapter.NewTmuxAdapter(),
		enabled: true,
	}

	logWeaverInfo(fmt.Sprintf("Weaver initialized in %s mode", modeString(mode)))
}

// ProcessIntent 处理 Intent（Shadow 模式）
func (w *WeaverManager) ProcessIntent(intent Intent) {
	if !w.enabled {
		return
	}

	// 阶段 2：只记录 Intent，不执行
	verdict, err := w.engine.ApplyIntent(
		&intentAdapter{intent}, // 适配器模式
		w.adapter.Resolver(),
		w.adapter.Projection(),
	)

	if err != nil {
		logWeaverError(fmt.Sprintf("Weaver error: %v", err))
		return
	}

	// 记录 Verdict
	if ShouldLogFacts() {
		logWeaverInfo(fmt.Sprintf("Verdict: %s (Safety: %d)", verdict.Message, verdict.Safety))
	}
}

// intentAdapter 适配器，将主包的 Intent 转换为 core.Intent
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

// ProcessIntentGlobal 全局函数，处理 Intent
func ProcessIntentGlobal(intent Intent) {
	if globalWeaver != nil {
		globalWeaver.ProcessIntent(intent)
	}
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
