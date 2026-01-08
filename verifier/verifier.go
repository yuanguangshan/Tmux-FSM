package verifier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"tmux-fsm/crdt"
	"tmux-fsm/replay"
	"tmux-fsm/semantic"
)

// Hash 用于表示哈希值
type Hash string

// Fact 表示一个经过验证的事件事实
type Fact struct {
	ID        Hash                    `json:"id"`
	Actor     crdt.ActorID           `json:"actor"`
	Parents   []Hash                 `json:"parents"`
	Timestamp int64                  `json:"timestamp"`
	Payload   crdt.SemanticEvent     `json:"payload"`
	PolicyRef Hash                   `json:"policy_ref"`
}

// VerifyInput 验证输入
type VerifyInput struct {
	Facts       []Fact
	Policies    map[Hash][]byte // Policy code as bytes
	Snapshot    *replay.TextState `json:"snapshot"`
	ExpectedRoot Hash            `json:"expected_root"`
}

// VerifyResult 验证结果
type VerifyResult struct {
	OK        bool   `json:"ok"`
	StateRoot Hash   `json:"state_root"`
	Error     string `json:"error,omitempty"`
	FactsUsed int    `json:"facts_used"`
	Policies  int    `json:"policies"`
}

// Verifier 可验证编辑器内核
type Verifier struct {
	policies map[Hash][]byte
}

// NewVerifier 创建新的验证器
func NewVerifier(policies map[Hash][]byte) *Verifier {
	return &Verifier{
		policies: policies,
	}
}

// Verify 执行验证
func (v *Verifier) Verify(input VerifyInput) VerifyResult {
	// 1. 校验 Fact 哈希自洽
	for _, f := range input.Facts {
		expectedID := calculateFactHash(f)
		if expectedID != f.ID {
			return VerifyResult{
				OK:    false,
				Error: fmt.Sprintf("Fact tampered: expected %s, got %s", expectedID, f.ID),
			}
		}
	}

	// 2. 构建 DAG + 拓扑排序（稳定）
	orderedFacts := v.topoSort(input.Facts)

	// 3. Replay（纯函数）
	initialState := replay.TextState{}
	if input.Snapshot != nil {
		initialState = *input.Snapshot
	}
	
	state := initialState
	for _, f := range orderedFacts {
		// 检查策略（简化版）
		if !v.checkPolicy(f, state) {
			return VerifyResult{
				OK:    false,
				Error: fmt.Sprintf("Policy violation at Fact %s", f.ID),
			}
		}
		// 应用事实
		state = v.applyFact(state, f.Payload)
	}

	// 4. 计算 State Root
	root := calculateStateHash(state)

	// 5. 比对
	if root != input.ExpectedRoot {
		return VerifyResult{
			OK:    false,
			Error: fmt.Sprintf("Root mismatch: expected %s, got %s", input.ExpectedRoot, root),
		}
	}

	return VerifyResult{
		OK:        true,
		StateRoot: root,
		FactsUsed: len(orderedFacts),
		Policies:  len(v.policies),
	}
}

// topoSort 拓扑排序
func (v *Verifier) topoSort(facts []Fact) []Fact {
	// 构建依赖图
	graph := make(map[Hash][]Hash)
	inDegree := make(map[Hash]int)
	
	for _, f := range facts {
		inDegree[f.ID] = 0
		graph[f.ID] = []Hash{}
	}
	
	// 建立边
	for _, f := range facts {
		for _, parent := range f.Parents {
			if _, exists := inDegree[parent]; exists {
				graph[parent] = append(graph[parent], f.ID)
				inDegree[f.ID]++
			}
		}
	}

	// Kahn 算法
	var queue []Hash
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	// 稳定排序
	sort.Slice(queue, func(i, j int) bool {
		return string(queue[i]) < string(queue[j])
	})

	var result []Fact
	factMap := make(map[Hash]Fact)
	for _, f := range facts {
		factMap[f.ID] = f
	}

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		result = append(result, factMap[id])

		for _, next := range graph[id] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	return result
}

// checkPolicy 检查策略（简化版）
func (v *Verifier) checkPolicy(f Fact, state replay.TextState) bool {
	// 这里可以实现策略检查逻辑
	// 例如：检查 AI 是否在允许的范围内操作
	actor := string(f.Actor)
	if len(actor) >= 2 && actor[:2] == "ai" {
		// AI 操作的特殊检查
		// 这里可以实现更复杂的策略检查
	}
	return true
}

// applyFact 应用事实
func (v *Verifier) applyFact(state replay.TextState, event crdt.SemanticEvent) replay.TextState {
	// 使用 replay 包来应用事实
	newState := state
	replay.ApplyFact(&newState, event.Fact)
	return newState
}

// calculateFactHash 计算事实的哈希
func calculateFactHash(f Fact) Hash {
	data, _ := json.Marshal(struct {
		Actor     crdt.ActorID       `json:"actor"`
		Parents   []Hash             `json:"parents"`
		Timestamp int64              `json:"timestamp"`
		Payload   crdt.SemanticEvent `json:"payload"`
		PolicyRef Hash               `json:"policy_ref"`
	}{
		Actor:     f.Actor,
		Parents:   f.Parents,
		Timestamp: f.Timestamp,
		Payload:   f.Payload,
		PolicyRef: f.PolicyRef,
	})
	
	hash := sha256.Sum256(data)
	return Hash(hex.EncodeToString(hash[:]))
}

// calculateStateHash 计算状态哈希
func calculateStateHash(state replay.TextState) Hash {
	data, _ := json.Marshal(state)
	hash := sha256.Sum256(data)
	return Hash(hex.EncodeToString(hash[:]))
}

// VerifyFromJSON 从 JSON 数据验证
func (v *Verifier) VerifyFromJSON(factsJSON []byte, expectedRoot Hash) (VerifyResult, error) {
	var facts []Fact
	if err := json.Unmarshal(factsJSON, &facts); err != nil {
		return VerifyResult{}, err
	}

	input := VerifyInput{
		Facts:        facts,
		ExpectedRoot: expectedRoot,
	}

	return v.Verify(input), nil
}