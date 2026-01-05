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
	planner := &logic.ShellFactBuilder{}
	engine := core.NewShadowEngine(planner)

	var proj core.Projection
	if mode == ModeWeaver {
		proj = &adapter.TmuxProjection{}
	} else {
		proj = &adapter.NoopProjection{}
	}

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
	coreIntent := &intentAdapter{intent: intent}
	verdict, err := m.engine.ApplyIntent(coreIntent, m.resolver, m.projection)

	if err != nil {
		logWeaver("Error applying intent: %v", err)
		return
	}

	if globalConfig.LogFacts {
		txID := "nil"
		safety := core.SafetyExact
		if verdict.Transaction != nil {
			txID = string(verdict.Transaction.ID)
			safety = verdict.Transaction.Safety
		}
		logWeaver("Verdict: %s (tx: %s) (Safety: %v)", verdict.Message, txID, safety)
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
	// 构造 Fact
	cf := core.Fact{
		Anchor: core.Anchor{
			ResourceID: lf.Target.Anchor.PaneID,
			Hint: core.AnchorHint{
				Line:   lf.Target.Anchor.LineHint,
				Column: 0, // Legacy Anchor doesn't strictly track column in hint, but Fact.Target might? Cursor is [row, col]
			},
			Hash: []byte(lf.Target.Anchor.LineHash),
		},
		Range: core.Range{
			StartOffset: lf.Target.StartOffset,
			EndOffset:   lf.Target.EndOffset,
		},
		SideEffects: lf.SideEffects,
	}

	if lf.Target.Anchor.Cursor != nil {
		cf.Anchor.Hint.Column = lf.Target.Anchor.Cursor[1]
	}

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
	return a.intent.PaneID
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
