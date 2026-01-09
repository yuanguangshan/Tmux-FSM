package manager

import (
	"fmt"
	"tmux-fsm/intent"
	"tmux-fsm/weaver/adapter"
	"tmux-fsm/weaver/core"
	"tmux-fsm/weaver/logic"
)

// ExecutionMode 执行模式
type ExecutionMode int

const (
	ModeLegacy ExecutionMode = iota // 传统模式
	ModeWeaver                      // Weaver模式
	ModeShadow                      // 仅观察模式
)

// WeaverManager 全局协调器
// RFC-WC-000: Kernel Sovereignty - 所有编辑决策必须通过Kernel
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
// RFC-WC-005: Audit Escape Prevention - 初始化必须可审计
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

	// Phase 6.3: Reality Reader for consistency adjudication
	reality := &adapter.TmuxRealityReader{Provider: snapProvider}
	resolver.Reality = reality

	var proj core.Projection
	if mode == ModeWeaver {
		proj = &adapter.TmuxProjection{}
	} else {
		proj = &adapter.NoopProjection{}
	}

	engine := core.NewShadowEngine(planner, resolver, proj, reality)

	weaverMgr = &WeaverManager{
		mode:             mode,
		engine:           engine,
		resolver:         resolver,
		projection:       proj,
		snapshotProvider: snapProvider,
	}
}

// ProcessIntentGlobal 全局意图处理入口
// RFC-WC-002: Intent ABI - 统一入口，统一审计
func (m *WeaverManager) ProcessIntentGlobal(intent core.Intent) error {
	if m == nil || m.mode == ModeLegacy {
		return nil // Fallback to legacy
	}

	// Phase 6.2: 获取当前快照作为时间冻结点
	snapshot, err := m.snapshotProvider.TakeSnapshot(intent.GetPaneID())
	if err != nil {
		return fmt.Errorf("failed to take snapshot: %v", err)
	}

	// Phase 6.3: ApplyIntent with frozen world state
	verdict, err := m.engine.ApplyIntent(intent, snapshot)
	if err != nil {
		return fmt.Errorf("engine failed: %v", err)
	}

	// RFC-WC-003: Audit Trail
	if verdict != nil {
		logWeaver("Intent processed: %v, Safety: %v", intent.GetKind(), verdict.Safety)
	}

	return nil
}

// Process 实现 IntentExecutor 接口
func (m *WeaverManager) Process(intent *intent.Intent) error {
	if m == nil || m.mode == ModeLegacy {
		return nil // Fallback to legacy
	}

	// 将统一的intent.Intent转换为core.Intent
	coreIntent := convertToCoreIntent(intent)

	// Phase 6.2: 获取当前快照作为时间冻结点
	snapshot, err := m.snapshotProvider.TakeSnapshot(coreIntent.GetPaneID())
	if err != nil {
		return fmt.Errorf("failed to take snapshot: %v", err)
	}

	// Phase 6.3: ApplyIntent with frozen world state
	verdict, err := m.engine.ApplyIntent(coreIntent, snapshot)
	if err != nil {
		return fmt.Errorf("engine failed: %v", err)
	}

	// RFC-WC-003: Audit Trail
	if verdict != nil {
		logWeaver("Intent processed: %v, Safety: %v", coreIntent.GetKind(), verdict.Safety)
	}

	return nil
}

// convertToCoreIntent 将统一的intent.Intent转换为core.Intent
func convertToCoreIntent(intent *intent.Intent) core.Intent {
	// 由于不能直接访问main.Intent，我们需要创建一个适配器
	return &intentAdapter{intent: intent}
}

// intentAdapter 适配器
type intentAdapter struct {
	intent *intent.Intent
}

func (a *intentAdapter) GetKind() core.IntentKind {
	return core.IntentKind(a.intent.Kind)
}

func (a *intentAdapter) GetTarget() core.SemanticTarget {
	return core.SemanticTarget{
		Kind:      int(a.intent.Target.Kind), // 使用intent中的Kind值
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

func (a *intentAdapter) GetSnapshotHash() string {
	return a.intent.SnapshotHash
}

func (a *intentAdapter) IsPartialAllowed() bool {
	return a.intent.AllowPartial
}

func (a *intentAdapter) GetAnchors() []core.Anchor {
	// 简化处理，返回空切片
	return []core.Anchor{}
}

// GetWeaverManager 获取全局 Weaver 管理器实例
func GetWeaverManager() *WeaverManager {
	return weaverMgr
}

// InjectLegacyTransaction 将传统事务注入 Weaver 系统
// RFC-WC-004: Legacy Bridge - 保持向后兼容但通过统一审计
// TODO: 实现传统事务到Weaver系统的桥接
func (m *WeaverManager) InjectLegacyTransaction(tx interface{}) {
	if m.mode == ModeLegacy {
		return
	}

	// Convert legacy transaction to Weaver-compatible format for audit
	logWeaver("Legacy transaction injected for audit")
}

// logWeaver ...
func logWeaver(format string, args ...interface{}) {
	// 实现日志记录
}
