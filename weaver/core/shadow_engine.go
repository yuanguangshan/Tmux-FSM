package core

import "time"

// ShadowEngine 影子模式引擎
// 目前也承担了 Execution Engine for Weaver Mode 的职责
type ShadowEngine struct {
	planner Planner // 依赖注入：负责生成 Facts
}

func NewShadowEngine(planner Planner) *ShadowEngine {
	return &ShadowEngine{
		planner: planner,
	}
}

func (e *ShadowEngine) ApplyIntent(intent Intent, resolver AnchorResolver, projection Projection) (*Verdict, error) {
	// 1. Plan: 将 Intent 转换为 Fact 序列
	// 这里的 PaneID 必须从 Intent 中获取
	facts, inverseFacts, err := e.planner.Build(intent, intent.GetPaneID())
	if err != nil {
		// Log error but maybe continue if we want to default to something?
		// For now, return error
		return nil, err
	}

	// 2. Transact: 创建事务
	// TODO: Replace with real UUID generator
	tx := Transaction{
		ID:           TransactionID("tx-" + time.Now().Format("150405")),
		Intent:       intent,
		Facts:        facts,
		InverseFacts: inverseFacts,
		Safety:       SafetyExact, // 暂时假设为 Exact
		Timestamp:    time.Now().Unix(),
		Skipped:      false,
	}

	// 3. Project: 执行 (Apply Facts)
	// 无论是什么 Mode，Engine 都负责调用 Projection。
	// 如果是 Shadow Mode，调用者应该传入 NoopProjection。
	// 如果是 Weaver Mode，调用者应该传入 TmuxProjection。
	if err := projection.Apply(nil, facts); err != nil {
		return nil, err
	}

	tx.Applied = true

	// 4. Return Verdict
	return &Verdict{
		Kind:        VerdictApplied,
		Message:     "Applied via Smart Projection",
		Transaction: &tx,
	}, nil
}

func (e *ShadowEngine) Undo() (*Verdict, error) {
	return nil, nil // Not implemented yet
}

func (e *ShadowEngine) Redo() (*Verdict, error) {
	return nil, nil // Not implemented yet
}
