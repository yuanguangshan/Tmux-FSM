package verifier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"tmux-fsm/crdt"
	"tmux-fsm/replay"
	"tmux-fsm/semantic"
)

//
// ─────────────────────────────────────────────────────────────
//  Hash
// ─────────────────────────────────────────────────────────────
//

type Hash string

func hashBytes(b []byte) Hash {
	h := sha256.Sum256(b)
	return Hash(hex.EncodeToString(h[:]))
}

//
// ─────────────────────────────────────────────────────────────
//  Canonical Types
// ─────────────────────────────────────────────────────────────
//

// CanonicalSemanticEvent 必须是确定性可序列化的
type CanonicalSemanticEvent struct {
	Actor         crdt.ActorID      `json:"actor"`
	CausalParents []crdt.EventID    `json:"causal_parents"`
	Fact          semantic.BaseFact `json:"fact"`
}

// Fact 是“已签名、可验证”的最小单元
type Fact struct {
	ID        Hash                    `json:"id"`
	Actor     crdt.ActorID           `json:"actor"`
	Parents   []Hash                 `json:"parents"`
	Timestamp int64                  `json:"timestamp"`
	Payload   CanonicalSemanticEvent `json:"payload"`
	PolicyRef Hash                   `json:"policy_ref"`
}

//
// ─────────────────────────────────────────────────────────────
//  Verify Input / Output
// ─────────────────────────────────────────────────────────────
//

type VerifyInput struct {
	Facts        []Fact
	Policies     map[Hash][]byte
	Snapshot     *replay.TextState
	ExpectedRoot Hash
}

type VerifyResult struct {
	OK        bool   `json:"ok"`
	StateRoot Hash   `json:"state_root"`
	Error     string `json:"error,omitempty"`

	FactsUsed int `json:"facts_used"`
	Policies  int `json:"policies"`
}

//
// ─────────────────────────────────────────────────────────────
//  Verifier
// ─────────────────────────────────────────────────────────────
//

type Verifier struct {
	policies map[Hash][]byte
}

func NewVerifier(policies map[Hash][]byte) *Verifier {
	return &Verifier{policies: policies}
}

//
// ─────────────────────────────────────────────────────────────
//  Verify Entry
// ─────────────────────────────────────────────────────────────
//

func (v *Verifier) Verify(input VerifyInput) VerifyResult {

	// 1️⃣ Fact 自洽校验
	for _, f := range input.Facts {
		if calcFactHash(f) != f.ID {
			return fail("fact hash mismatch: " + string(f.ID))
		}
	}

	// 2️⃣ DAG + 稳定拓扑排序 + 环检测
	ordered, err := topoSortFacts(input.Facts)
	if err != nil {
		return fail(err.Error())
	}

	// 3️⃣ 初始状态
	state := replay.TextState{}
	if input.Snapshot != nil {
		state = input.Snapshot.Clone()
	}

	// 4️⃣ 纯 Replay
	for _, f := range ordered {

		if err := v.checkPolicy(f, state); err != nil {
			return fail(fmt.Sprintf("policy violation at %s: %v", f.ID, err))
		}

		next := state
		replay.ApplyFact(&next, f.Payload.Fact)
		state = next
	}

	// 5️⃣ State Root
	root := calcStateHash(state)

	if root != input.ExpectedRoot {
		return fail(fmt.Sprintf(
			"state root mismatch: expected %s, got %s",
			input.ExpectedRoot, root,
		))
	}

	return VerifyResult{
		OK:        true,
		StateRoot: root,
		FactsUsed: len(ordered),
		Policies:  len(v.policies),
	}
}

func fail(msg string) VerifyResult {
	return VerifyResult{OK: false, Error: msg}
}

//
// ─────────────────────────────────────────────────────────────
//  Topological Sort (Stable + Cycle Detect)
// ─────────────────────────────────────────────────────────────
//

func topoSortFacts(facts []Fact) ([]Fact, error) {

	graph := map[Hash][]Hash{}
	inDegree := map[Hash]int{}
	factMap := map[Hash]Fact{}

	for _, f := range facts {
		graph[f.ID] = nil
		inDegree[f.ID] = 0
		factMap[f.ID] = f
	}

	for _, f := range facts {
		for _, p := range f.Parents {
			if _, ok := inDegree[p]; ok {
				graph[p] = append(graph[p], f.ID)
				inDegree[f.ID]++
			}
		}
	}

	var queue []Hash
	for id, d := range inDegree {
		if d == 0 {
			queue = append(queue, id)
		}
	}

	sort.Slice(queue, func(i, j int) bool {
		return string(queue[i]) < string(queue[j])
	})

	var out []Fact

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		out = append(out, factMap[id])

		for _, nxt := range graph[id] {
			inDegree[nxt]--
			if inDegree[nxt] == 0 {
				queue = append(queue, nxt)
			}
		}
	}

	if len(out) != len(facts) {
		return nil, errors.New("cycle detected in fact graph")
	}

	return out, nil
}

//
// ─────────────────────────────────────────────────────────────
//  Policy (Minimal / Deterministic)
// ─────────────────────────────────────────────────────────────
//

func (v *Verifier) checkPolicy(f Fact, state replay.TextState) error {

	// 1️⃣ Policy code must exist
	if _, ok := v.policies[f.PolicyRef]; !ok {
		return errors.New("unknown policy ref")
	}

	// 2️⃣ 最小 AI 防线（deterministic）
	actor := string(f.Actor)
	if len(actor) >= 2 && actor[:2] == "ai" {
		switch f.Payload.Fact.Kind() {
		case "insert", "delete", "move":
			return nil
		default:
			return errors.New("ai operation not allowed")
		}
	}

	return nil
}

//
// ─────────────────────────────────────────────────────────────
//  Hashing (Canonical)
// ─────────────────────────────────────────────────────────────
//

func calcFactHash(f Fact) Hash {

	parents := append([]Hash{}, f.Parents...)
	sort.Slice(parents, func(i, j int) bool {
		return parents[i] < parents[j]
	})

	data, _ := json.Marshal(struct {
		Actor     crdt.ActorID           `json:"actor"`
		Parents   []Hash                 `json:"parents"`
		Timestamp int64                  `json:"timestamp"`
		Payload   CanonicalSemanticEvent `json:"payload"`
		PolicyRef Hash                   `json:"policy_ref"`
	}{
		Actor:     f.Actor,
		Parents:   parents,
		Timestamp: f.Timestamp,
		Payload:   f.Payload,
		PolicyRef: f.PolicyRef,
	})

	return hashBytes(data)
}

func calcStateHash(state replay.TextState) Hash {
	data, _ := json.Marshal(state)
	return hashBytes(data)
}

//
// ─────────────────────────────────────────────────────────────
//  JSON Helper
// ─────────────────────────────────────────────────────────────
//

func (v *Verifier) VerifyFromJSON(
	factsJSON []byte,
	expectedRoot Hash,
) (VerifyResult, error) {

	var facts []Fact
	if err := json.Unmarshal(factsJSON, &facts); err != nil {
		return VerifyResult{}, err
	}

	return v.Verify(VerifyInput{
		Facts:        facts,
		ExpectedRoot: expectedRoot,
	}), nil
}
