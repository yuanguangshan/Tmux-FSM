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
	mode             ExecutionMode
	engine           core.Engine // Interface? No, ShadowEngine struct usually.
	resolver         core.AnchorResolver
	projection       core.Projection
	snapshotProvider adapter.SnapshotProvider // Phase 6.2
}

// weaverMgr 全局 Weaver 实例
var weaverMgr *WeaverManager

// InitWeaver 初始化 Weaver 系统
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

	var proj core.Projection
	if mode == ModeWeaver {
		proj = &adapter.TmuxProjection{}
	} else {
		proj = &adapter.NoopProjection{}
	}

	engine := core.NewShadowEngine(planner, resolver, proj)

	weaverMgr = &WeaverManager{
		mode:             mode,
		engine:           engine,
		resolver:         resolver,
		projection:       proj,
		snapshotProvider: snapProvider,
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

// ProcessIntent 处理意图 (Gateway)
func (m *WeaverManager) ProcessIntent(intent Intent) {
	logWeaver("ProcessIntent: Kind=%v Target=%v", intent.Kind, intent.Target)

	// Phase 6.2: Capture Snapshot (Time Freeze)
	paneID := intent.GetPaneID()
	if paneID == "" {
		// Try to deduce or fail
		logWeaver("No PaneID in intent, skipping snapshot")
		return // Or handle non-pane intents
	}

	snapshot, err := m.snapshotProvider.TakeSnapshot(paneID)
	if err != nil {
		logWeaver("Snapshot failed: %v", err)
		return
	}

	// Inject Hash into Intent (mutable struct in main)
	intent.SnapshotHash = string(snapshot.Hash)

	coreIntent := &intentAdapter{intent: intent}

	// 此时如果是 Undo/Redo，它们不需要 Snapshot?
	// Phase 6.2 定义：Any ApplyIntent needs Snapshot.
	// Undo/Redo often imply "Previous State", but current implementation calls Planner even for Undo/Redo?
	// No, `ApplyIntent` handles Undo/Redo specially.
	// It calls `performUndo`.

	if m.mode == ModeShadow || m.mode == ModeWeaver {
		verdict, err := m.engine.ApplyIntent(coreIntent, snapshot)
		if err != nil {
			logWeaver("Engine Error: %v", err)
		} else {
			logWeaver("Verdict: %v (Safe=%v)", verdict.Kind, verdict.Safety)
			if len(verdict.Audit) > 0 {
				logWeaver("Audit: %v", verdict.Audit)
			}
		}
	}

	// [Phase 4] Phase 3 的 Weaver -> Legacy 桥接已禁用
	// 现在 Weaver History 是 Source of Truth，Legacy 操作将通过反向桥接注入 Weaver
}

// InjectLegacyTransaction 将 Legacy 事务注入到 Weaver History (Reverse Bridge)
func (m *WeaverManager) InjectLegacyTransaction(legacyTx *Transaction) {
	if m.engine == nil {
		return
	}
	// 获取 ShadowEngine 的 History
	se, ok := m.engine.(*core.ShadowEngine)
	if !ok {
		return
	}

	coreTx := &core.Transaction{
		ID:           core.TransactionID(fmt.Sprintf("legacy-%d", legacyTx.ID)),
		Timestamp:    legacyTx.CreatedAt.Unix(),
		Facts:        make([]core.Fact, 0),
		InverseFacts: make([]core.Fact, 0),
		Applied:      true,
		Safety:       core.SafetyExact,
	}

	// 转换正向事实
	for _, rec := range legacyTx.Records {
		f := convertLegacyFactToCore(rec.Fact)
		coreTx.Facts = append(coreTx.Facts, f)
	}

	// 转换反向事实 (通常 Inverse 用于 Undo。Legacy Undo 执行 Inverse。
	// Weaver Undo 执行 InverseFacts。顺序：Record1, Record2. Undo: Inv2, Inv1。
	// 所以我们需要倒序遍历 Records)
	for i := len(legacyTx.Records) - 1; i >= 0; i-- {
		rec := legacyTx.Records[i]
		inv := convertLegacyFactToCore(rec.Inverse)
		coreTx.InverseFacts = append(coreTx.InverseFacts, inv)
	}

	if len(coreTx.Facts) > 0 {
		se.GetHistory().Push(coreTx)
		logWeaver("Injected Legacy Transaction %d -> %s", legacyTx.ID, coreTx.ID)
	}
}

func convertLegacyFactToCore(lf Fact) core.Fact {
	// Construct Semantic Anchor with Legacy Physical Info for Resolver to unpack
	ref := map[string]int{
		"line":  lf.Target.Anchor.LineHint,
		"start": lf.Target.StartOffset,
		"end":   lf.Target.EndOffset,
	}

	cf := core.Fact{
		Anchor: core.Anchor{
			PaneID: lf.Target.Anchor.PaneID,
			Kind:   core.AnchorLegacyRange,
			Ref:    ref,
		},
		SideEffects: lf.SideEffects,
	}
	// Note: Hash is currently ignored in legacy conversion

	switch lf.Kind {
	case "delete":
		cf.Kind = core.FactDelete
		cf.Payload.OldText = lf.Target.Text
	case "insert":
		cf.Kind = core.FactInsert
		cf.Payload.Text = lf.Target.Text
	case "replace":
		cf.Kind = core.FactReplace
		cf.Payload.OldText = lf.Target.Text
		if s, ok := lf.Meta["new_text"].(string); ok {
			cf.Payload.NewText = s
		}
	default:
		cf.Kind = core.FactNone
	}
	return cf
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

// logWeaver ...
func logWeaver(format string, args ...interface{}) {
	if !globalConfig.LogFacts {
		return
	}
	f, _ := os.OpenFile(os.Getenv("HOME")+"/tmux-fsm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%s] [WEAVER] %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
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
