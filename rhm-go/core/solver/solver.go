package solver

import (
	"container/heap"
	"rhm-go/core/analysis"
	"rhm-go/core/change"
	"rhm-go/core/cost"
	"rhm-go/core/history"
	"rhm-go/core/narrative"
	"rhm-go/core/rewrite"
	"rhm-go/core/search"
)

type ResolutionPlan struct {
	Mutations []change.Mutation
	Resolved  bool
	Narrative narrative.Narrative
}

// Solve 核心入口：寻找最优时间线
func Solve(dag *history.HistoryDAG, tipA, tipB history.NodeID) ResolutionPlan {
	costModel := cost.DefaultModel{}
	pq := &search.PriorityQueue{}
	heap.Init(pq)

	// closedSet 用于存储已探索过的状态指纹，避免指数爆炸
	closedSet := make(map[uint64]bool)

	// 1. 初始化空状态 (没有突变的状态)
	heap.Push(pq, &search.State{
		Mutations: []change.Mutation{},
		Cost:      0,
		Fingerprint: 0,
	})

	for pq.Len() > 0 {
		// 取出当前 Cost 最低的状态进行扩展
		current := heap.Pop(pq).(*search.State)

		// 指纹检查
		if closedSet[current.Fingerprint] {
			continue
		}
		closedSet[current.Fingerprint] = true

		// 2. 环境重构：在沙盒中应用当前的突变计划
		// 这里的 "root" 应该通过 LCA 算法计算，为了演示简化为 "root"
		sandbox := rewrite.RewriteBatch(dag, "root", current.Mutations)

		// 3. 冲突分析：利用 Footprint 代数检查新环境是否还有冲突
		res := analysis.AnalyzeMerge(sandbox, tipA, tipB)

		// 目标达成：没有冲突，当前 current 即为最优解
		if len(res.Conflicts) == 0 {
			return ResolutionPlan{
				Mutations: current.Mutations,
				Resolved:  true,
				Narrative: narrative.Narrative{
					Summary:   "Conflict resolved via optimized causal path",
					TotalCost: int(current.Cost),
					Steps:     current.Narrative,
				},
			}
		}

		// 4. 定向扩展：只处理第一个被检测到的冲突
		conflict := res.Conflicts[0]
		involved := []history.NodeID{conflict.NodeA, conflict.NodeB}

		for _, offenderID := range involved {
			// 定向获取该节点的候选变体 (Downgrade/NoOp)
			candidates := generateTargetedCandidates(sandbox, offenderID)

			for _, mut := range candidates {
				c := costModel.Calculate(mut, cost.Context{})

				// 记录决策轨迹
				step := narrative.DecisionStep{
					ProblemContext: conflict.Reason,
					Decision:       mut.String(),
					DecisionCost:   int(c),
				}

				// 创建新状态并入队
				nextMutations := make([]change.Mutation, len(current.Mutations))
				copy(nextMutations, current.Mutations)
				nextMutations = append(nextMutations, mut)

				nextState := &search.State{
					Mutations:   nextMutations,
					Cost:        current.Cost + c,
					Narrative:   append(append([]narrative.DecisionStep{}, current.Narrative...), step),
					Fingerprint: search.ComputeFingerprint(nextMutations),
				}

				heap.Push(pq, nextState)
			}
		}
	}

	return ResolutionPlan{Resolved: false}
}

// generateTargetedCandidates 基于冲突节点生成局部候选方案
func generateTargetedCandidates(view history.DagView, id history.NodeID) []change.Mutation {
	node := view.GetNode(id)
	if node == nil {
		return nil
	}

	muts := []change.Mutation{}

	// 尝试一：降级语义 (如 Delete -> Move，保留大部分意图)
	if down := node.Op.Downgrade(); down != nil {
		muts = append(muts, change.Mutation{
			Type:   change.ReplaceOp,
			Target: string(id),
			NewOp:  down,
		})
	}

	// 尝试二：彻底中和 (NoOp，牺牲意图以换取一致性)
	if noop := node.Op.ToNoOp(); noop != nil {
		muts = append(muts, change.Mutation{
			Type:   change.ReplaceOp,
			Target: string(id),
			NewOp:  noop,
		})
	}

	return muts
}
