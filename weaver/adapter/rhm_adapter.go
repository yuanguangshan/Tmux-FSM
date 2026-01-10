package adapter

import (
	"fmt"
	"rhm-go/core/change"
	"rhm-go/core/history"
	"rhm-go/core/solver"
	"tmux-fsm/editor"
)

// RHMAdapter 是 Tmux-FSM 主项目与 RHM-Go 引擎之间的桥梁。
// 它负责将主项目的 ResolvedOperation 映射到 RHM 的因果模型中。
type RHMAdapter struct {
	// 目前保持无状态，未来可注入自定义 CostModel
}

// NewRHMAdapter 创建一个新的适配器
func NewRHMAdapter() *RHMAdapter {
	return &RHMAdapter{}
}

// OpWrapper 将 editor.ResolvedOperation 包装为 rhm-go 的 change.ReversibleChange
type OpWrapper struct {
	op editor.ResolvedOperation
}

func (w *OpWrapper) Describe() string {
	// 简单的描述生成
	return fmt.Sprintf("%d:%s", w.op.Kind(), w.op.OpID())
}

func (w *OpWrapper) ToNoOp() change.ReversibleChange {
	// 在 RHM 中，NoOp 是中和冲突的选择
	return &NoOpWrapper{id: w.op.OpID()}
}

func (w *OpWrapper) Downgrade() change.ReversibleChange {
	// 如果是删除操作，可以降级为某种形式的“保留式删除”
	if w.op.Kind() == editor.OpDelete {
		return &DowngradeWrapper{op: w.op}
	}
	return nil
}

func (w *OpWrapper) Hash() string {
	return string(w.op.OpID())
}

func (w *OpWrapper) GetFootprints() []change.Footprint {
	fp := w.op.Footprint()
	res := make([]change.Footprint, 0, len(fp.Buffers))
	mode := change.Shared
	for _, e := range fp.Effects {
		if e == editor.EffectWrite || e == editor.EffectDelete {
			mode = change.Exclusive
		}
		if e == editor.EffectCreate {
			mode = change.Create
		}
	}
	for _, b := range fp.Buffers {
		res = append(res, change.Footprint{ResourceID: string(b), Mode: mode})
	}
	return res
}

// NoOpWrapper 代表被中和的操作
type NoOpWrapper struct {
	id editor.OperationID
}

func (w *NoOpWrapper) Describe() string                   { return "NoOp(Neutralized)" }
func (w *NoOpWrapper) ToNoOp() change.ReversibleChange    { return w }
func (w *NoOpWrapper) Downgrade() change.ReversibleChange { return nil }
func (w *NoOpWrapper) Hash() string                       { return "noop:" + string(w.id) }
func (w *NoOpWrapper) GetFootprints() []change.Footprint  { return nil }

// DowngradeWrapper 代表降级后的操作
type DowngradeWrapper struct {
	op editor.ResolvedOperation
}

func (w *DowngradeWrapper) Describe() string {
	return "Downgraded(" + string(w.op.OpID()) + ")"
}
func (w *DowngradeWrapper) ToNoOp() change.ReversibleChange    { return &NoOpWrapper{id: w.op.OpID()} }
func (w *DowngradeWrapper) Downgrade() change.ReversibleChange { return nil }
func (w *DowngradeWrapper) Hash() string                       { return "down:" + string(w.op.OpID()) }
func (w *DowngradeWrapper) GetFootprints() []change.Footprint {
	// 降级通常意味着将 Exclusive 变为 Shared 或更弱的形式
	return []change.Footprint{{ResourceID: "trash", Mode: change.Shared}}
}

// MapToDAG 将主项目的一组操作及其因果关系映射为 RHM 的 HistoryDAG
func (a *RHMAdapter) MapToDAG(ops []editor.ResolvedOperation, dependencies map[editor.OperationID][]editor.OperationID) *history.HistoryDAG {
	dag := history.NewHistoryDAG()
	for _, op := range ops {
		parents := []history.NodeID{}
		if deps, ok := dependencies[op.OpID()]; ok {
			for _, d := range deps {
				parents = append(parents, history.NodeID(d))
			}
		}
		dag.AddOp(history.NodeID(op.OpID()), &OpWrapper{op: op}, parents)
	}
	return dag
}

// Solve 利用 RHM 引擎求解冲突
func (a *RHMAdapter) Solve(dag *history.HistoryDAG, tipA, tipB editor.OperationID) solver.ResolutionPlan {
	return solver.Solve(dag, history.NodeID(tipA), history.NodeID(tipB))
}

// ResolutionAction 代表适配器转换回来的最终行动
type ResolutionAction struct {
	TargetID editor.OperationID
	NewOp    editor.ResolvedOperation // 如果为 nil 且是 ReplaceOp，可能代表 Neutralize (NoOp)
	IsNoOp   bool
}

// ExtractActions 从 RHM 的求解计划中提取主项目可识别的动作序列
func (a *RHMAdapter) ExtractActions(plan solver.ResolutionPlan) []ResolutionAction {
	actions := make([]ResolutionAction, 0, len(plan.Mutations))
	for _, m := range plan.Mutations {
		action := ResolutionAction{
			TargetID: editor.OperationID(m.Target),
		}

		switch op := m.NewOp.(type) {
		case *OpWrapper:
			action.NewOp = op.op
		case *NoOpWrapper:
			action.IsNoOp = true
		case *DowngradeWrapper:
			// 这里假设 DowngradeWrapper 内部包装了一个降级后的真实 Op
			action.NewOp = op.op // 在实际集成中，此处应为真正的降级实现
		}
		actions = append(actions, action)
	}
	return actions
}
