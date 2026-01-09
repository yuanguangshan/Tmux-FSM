package policy

import (
	"errors"

	"tmux-fsm/crdt"
	"tmux-fsm/semantic"
)

//
// ─────────────────────────────────────────────────────────────
//  Trust Model
// ─────────────────────────────────────────────────────────────
//

// TrustLevel 表示“是否拥有最终提交权”
type TrustLevel int

const (
	TrustSystem   TrustLevel = iota // GC / snapshot / rebalance
	TrustUser                       // 人类
	TrustDevice                     // 同一用户的多端
	TrustAI                         // 只能 proposal
	TrustExternal                   // 插件 / import（默认只读）
)

//
// ─────────────────────────────────────────────────────────────
//  Actor
// ─────────────────────────────────────────────────────────────
//

// ActorInfo 只存储“身份 + 信任级别”
type ActorInfo struct {
	ID    crdt.ActorID
	Level TrustLevel
	Name  string
}

//
// ─────────────────────────────────────────────────────────────
//  Semantic Operation
// ─────────────────────────────────────────────────────────────
//

type OpKind string

const (
	OpInsert  OpKind = "insert"
	OpDelete  OpKind = "delete"
	OpMove    OpKind = "move"
	OpReplace OpKind = "replace" // Added OpReplace
	OpFormat  OpKind = "format"
)

//
// ─────────────────────────────────────────────────────────────
//  Scope（AI 的语义沙箱）
// ─────────────────────────────────────────────────────────────
//

// Scope 表示 AI 被允许操作的“语义范围”
type Scope struct {
	DocumentID string
	Range      semantic.Range
	AllowedOps []OpKind
}

func (s Scope) allowsOp(op OpKind) bool {
	for _, a := range s.AllowedOps {
		if a == op {
			return true
		}
	}
	return false
}

//
// ─────────────────────────────────────────────────────────────
//  AI Draft（注意：不是 Event）
// ─────────────────────────────────────────────────────────────
//

type AIDraft struct {
	Fact semantic.Fact
}

//
// ─────────────────────────────────────────────────────────────
//  Policy Interface
// ─────────────────────────────────────────────────────────────
//

// Policy 是 CRDT 的安全边界
type Policy interface {
	RegisterActor(info ActorInfo)

	// AllowCommit：是否允许 actor 提交最终 CRDT Event
	AllowCommit(actor crdt.ActorID, event crdt.SemanticEvent) error

	// AllowAIDraft：是否允许 AI 在 scope 内提出 draft
	AllowAIDraft(actor crdt.ActorID, scope Scope, draft AIDraft) error

	// ValidateAIProposal：批量校验 AI 提案
	ValidateAIProposal(proposal AIProposal) error
}

//
// ─────────────────────────────────────────────────────────────
//  DefaultPolicy
// ─────────────────────────────────────────────────────────────
//

type DefaultPolicy struct {
	actors map[crdt.ActorID]ActorInfo
}

func NewDefaultPolicy() *DefaultPolicy {
	return &DefaultPolicy{
		actors: make(map[crdt.ActorID]ActorInfo),
	}
}

func (p *DefaultPolicy) RegisterActor(info ActorInfo) {
	p.actors[info.ID] = info
}

//
// ─────────────────────────────────────────────────────────────
//  Commit Path（CRDT 最终入口）
// ─────────────────────────────────────────────────────────────
//

func (p *DefaultPolicy) AllowCommit(
	actorID crdt.ActorID,
	_ crdt.SemanticEvent,
) error {

	actor, ok := p.actors[actorID]
	if !ok {
		return errors.New("unknown actor")
	}

	switch actor.Level {
	case TrustSystem, TrustUser, TrustDevice:
		return nil

	case TrustAI:
		return errors.New("AI is not allowed to commit CRDT events")

	default:
		return errors.New("actor not allowed to commit")
	}
}

//
// ─────────────────────────────────────────────────────────────
//  AI Draft Path（唯一合法 AI 入口）
// ─────────────────────────────────────────────────────────────
//

func (p *DefaultPolicy) AllowAIDraft(
	actorID crdt.ActorID,
	scope Scope,
	draft AIDraft,
) error {

	actor, ok := p.actors[actorID]
	if !ok {
		return errors.New("unknown actor")
	}

	if actor.Level != TrustAI {
		return errors.New("actor is not AI")
	}

	op := factKindToOpKind(draft.Fact.Kind())

	// 1️⃣ 操作类型检查
	if !scope.allowsOp(op) {
		return errors.New("operation not allowed in scope: " + string(op))
	}

	// 2️⃣ 范围检查（语义级）
	if !scope.Range.ContainsFact(draft.Fact) {
		return errors.New("draft out of allowed range")
	}

	return nil
}

//
// ─────────────────────────────────────────────────────────────
//  AI Proposal
// ─────────────────────────────────────────────────────────────
//

type AIProposal struct {
	Actor  crdt.ActorID
	Scope  Scope
	Drafts []AIDraft
}

func (p *DefaultPolicy) ValidateAIProposal(
	proposal AIProposal,
) error {

	for _, draft := range proposal.Drafts {
		if err := p.AllowAIDraft(
			proposal.Actor,
			proposal.Scope,
			draft,
		); err != nil {
			return err
		}
	}

	return nil
}

// factKindToOpKind 将 semantic.FactKind 转换为 policy.OpKind
func factKindToOpKind(kind semantic.FactKind) OpKind {
	switch kind {
	case semantic.FactInsert:
		return OpInsert
	case semantic.FactDelete:
		return OpDelete
	case semantic.FactMove:
		return OpMove
	case semantic.FactReplace:
		return OpReplace
	default:
		return OpKind("unknown")
	}
}
