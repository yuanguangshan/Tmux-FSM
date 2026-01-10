package solver

import (
	"reflect"
	"rhm-go/core/history"
	"rhm-go/internal/loader"
	"rhm-go/store"
	"testing"
)

// TestStability_OrderInvariance 验证：DAG 构造顺序不影响裁决结果
func TestStability_OrderInvariance(t *testing.T) {
	// 构造方式 A
	dagA, tipA, tipB := loader.LoadDemoScenario()
	resA := Solve(dagA, tipA, tipB)

	// 构造方式 B：反转分支插入顺序
	dagB := history.NewHistoryDAG()
	dagB.AddOp("root", store.FileSystemOp{Kind: "Create", Arg: "README.md"}, []history.NodeID{})
	dagB.AddOp("nodeB", store.FileSystemOp{Kind: "Delete", Arg: "README.md"}, []history.NodeID{"root"})
	dagB.AddOp("nodeA", store.FileSystemOp{Kind: "Edit", Arg: "README.md"}, []history.NodeID{"root"})

	resB := Solve(dagB, "nodeA", "nodeB")

	if resA.Narrative.TotalCost != resB.Narrative.TotalCost {
		t.Errorf("Order Invariance Failed: Cost mismatch %d vs %d", resA.Narrative.TotalCost, resB.Narrative.TotalCost)
	}
	if len(resA.Mutations) != len(resB.Mutations) {
		t.Errorf("Order Invariance Failed: Plan length mismatch")
	}
}

// TestStability_CostDominance 验证：Solver 必须选择 Cost 最小的“降级”路径 (50) 而非“中和”路径 (100)
func TestStability_CostDominance(t *testing.T) {
	dag, tipA, tipB := loader.LoadDemoScenario()
	res := Solve(dag, tipA, tipB)

	const expectedOptimalCost = 50 // Downgrade (Delete -> Move) should be 50 SLU
	if res.Narrative.TotalCost != expectedOptimalCost {
		t.Errorf("Cost Dominance Failed: Expected %d, got %d. Solver might be biased or search space incomplete.", expectedOptimalCost, res.Narrative.TotalCost)
	}

	// 确认决策确实是针对 nodeB 的 Move (因为 nodeB 是 Delete)
	foundDowngrade := false
	for _, step := range res.Narrative.Steps {
		if step.DecisionCost == expectedOptimalCost {
			foundDowngrade = true
		}
	}
	if !foundDowngrade {
		t.Errorf("Cost Dominance Failed: Narrative does not reflect the optimal downgrade decision")
	}
}

// TestStability_Determinism 验证：同 DAG 下 100 次运行结果必须比特级一致
func TestStability_Determinism(t *testing.T) {
	dag, tipA, tipB := loader.LoadDemoScenario()

	firstRes := Solve(dag, tipA, tipB)

	for i := 0; i < 100; i++ {
		currentRes := Solve(dag, tipA, tipB)
		if !reflect.DeepEqual(firstRes.Narrative, currentRes.Narrative) {
			t.Fatalf("Determinism Failed at iteration %d: Narrative mismatch", i)
		}
		if !reflect.DeepEqual(firstRes.Mutations, currentRes.Mutations) {
			t.Fatalf("Determinism Failed at iteration %d: Mutations mismatch", i)
		}
	}
}
