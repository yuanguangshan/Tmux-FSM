package policy

import (
	"errors"
	"tmux-fsm/crdt"
	"tmux-fsm/semantic"
)

// TrustLevel 信任级别
type TrustLevel int

const (
	TrustSystem TrustLevel = iota // GC / rebalance
	TrustUser                     // 人类
	TrustDevice                   // 同用户多端
	TrustAI                       // AI
	TrustExternal                 // 插件 / import
)

// ActorInfo 参与者信息
type ActorInfo struct {
	ID      crdt.ActorID
	Level   TrustLevel
	Name    string
	Allowed []string // 允许的操作类型
}

// PolicyContext 策略上下文
type PolicyContext struct {
	ActorInfo ActorInfo
	AllowedSlice EventSlice
	Timestamp int64
}

// EventSlice 事件切片
type EventSlice struct {
	From crdt.EventID
	To   crdt.EventID
	Events []crdt.SemanticEvent
}

// Policy 策略接口
type Policy interface {
	Allow(event crdt.SemanticEvent, ctx PolicyContext) error
}

// DefaultPolicy 默认策略实现
type DefaultPolicy struct {
	actors map[crdt.ActorID]ActorInfo
}

// NewDefaultPolicy 创建默认策略
func NewDefaultPolicy() *DefaultPolicy {
	return &DefaultPolicy{
		actors: make(map[crdt.ActorID]ActorInfo),
	}
}

// RegisterActor 注册参与者
func (p *DefaultPolicy) RegisterActor(actorID crdt.ActorID, level TrustLevel, name string) {
	p.actors[actorID] = ActorInfo{
		ID:      actorID,
		Level:   level,
		Name:    name,
		Allowed: []string{"insert", "delete", "move"}, // 默认允许的操作
	}
}

// Allow 检查事件是否被允许
func (p *DefaultPolicy) Allow(event crdt.SemanticEvent, ctx PolicyContext) error {
	actorInfo, exists := p.actors[event.Actor]
	if !exists {
		return errors.New("unknown actor")
	}

	// 根据信任级别进行不同的检查
	switch actorInfo.Level {
	case TrustAI:
		// AI 的特殊检查
		return p.checkAIEvent(event, ctx)
	case TrustUser:
		// 用户检查
		return p.checkUserEvent(event, ctx)
	case TrustSystem:
		// 系统操作检查
		return p.checkSystemEvent(event, ctx)
	default:
		// 其他类型的检查
		return p.checkGeneralEvent(event, ctx)
	}
}

// checkAIEvent 检查 AI 事件
func (p *DefaultPolicy) checkAIEvent(event crdt.SemanticEvent, ctx PolicyContext) error {
	// 检查 AI 是否在允许的范围内操作
	if ctx.AllowedSlice.From != "" && ctx.AllowedSlice.To != "" {
		// 检查事件是否在允许的范围内
		// 这里简化处理，实际实现需要更复杂的逻辑
	}

	// 检查操作类型是否被允许
	factKind := event.Fact.Kind()
	allowed := false
	for _, allowedOp := range ctx.ActorInfo.Allowed {
		if allowedOp == factKind {
			allowed = true
			break
		}
	}
	
	if !allowed {
		return errors.New("AI operation not allowed: " + factKind)
	}

	return nil
}

// checkUserEvent 检查用户事件
func (p *DefaultPolicy) checkUserEvent(event crdt.SemanticEvent, ctx PolicyContext) error {
	// 用户通常可以执行所有基本操作
	factKind := event.Fact.Kind()
	
	// 检查是否是允许的操作
	allowed := false
	for _, allowedOp := range ctx.ActorInfo.Allowed {
		if allowedOp == factKind {
			allowed = true
			break
		}
	}
	
	if !allowed {
		return errors.New("user operation not allowed: " + factKind)
	}

	return nil
}

// checkSystemEvent 检查系统事件
func (p *DefaultPolicy) checkSystemEvent(event crdt.SemanticEvent, ctx PolicyContext) error {
	// 系统操作通常只允许特定类型
	factKind := event.Fact.Kind()
	
	// 系统操作可能包括：rebalance, gc, snapshot 等
	systemOps := []string{"rebalance", "gc", "snapshot"}
	
	for _, sysOp := range systemOps {
		if sysOp == factKind {
			return nil
		}
	}
	
	return errors.New("system operation not allowed: " + factKind)
}

// checkGeneralEvent 检查一般事件
func (p *DefaultPolicy) checkGeneralEvent(event crdt.SemanticEvent, ctx PolicyContext) error {
	// 一般检查
	factKind := event.Fact.Kind()
	
	allowed := false
	for _, allowedOp := range ctx.ActorInfo.Allowed {
		if allowedOp == factKind {
			allowed = true
			break
		}
	}
	
	if !allowed {
		return errors.New("operation not allowed: " + factKind)
	}

	return nil
}

// ValidateEventSlice 验证事件切片
func (p *DefaultPolicy) ValidateEventSlice(slice EventSlice) error {
	for _, event := range slice.Events {
		ctx := PolicyContext{
			ActorInfo: p.actors[event.Actor],
			AllowedSlice: slice,
		}
		
		if err := p.Allow(event, ctx); err != nil {
			return err
		}
	}
	
	return nil
}

// GetActorTrustLevel 获取参与者信任级别
func (p *DefaultPolicy) GetActorTrustLevel(actorID crdt.ActorID) (TrustLevel, bool) {
	info, exists := p.actors[actorID]
	if !exists {
		return TrustExternal, false
	}
	return info.Level, true
}

// AIProposal AI 提案
type AIProposal struct {
	SessionID string
	Actor     crdt.ActorID
	Context   EventSlice
	Proposed  []DraftEvent
}

// DraftEvent 草案事件
type DraftEvent struct {
	Fact semantic.BaseFact
}

// ValidateAIProposal 验证 AI 提案
func (p *DefaultPolicy) ValidateAIProposal(proposal AIProposal) error {
	// 检查提案者是否是 AI
	level, exists := p.GetActorTrustLevel(proposal.Actor)
	if !exists || level != TrustAI {
		return errors.New("proposal must come from AI actor")
	}

	// 验证上下文
	if err := p.ValidateEventSlice(proposal.Context); err != nil {
		return err
	}

	// 验证提议的事件
	ctx := PolicyContext{
		ActorInfo:    p.actors[proposal.Actor],
		AllowedSlice: proposal.Context,
	}
	
	for _, draft := range proposal.Proposed {
		// 创建一个临时事件来检查
		tempEvent := crdt.SemanticEvent{
			Fact: draft.Fact,
		}
		
		if err := p.Allow(tempEvent, ctx); err != nil {
			return err
		}
	}

	return nil
}