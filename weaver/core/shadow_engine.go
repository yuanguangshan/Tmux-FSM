package core

import (
	"fmt"
	"time"
)

// ShadowEngine 影子引擎实现
// 阶段 2：只产生 Fact，不执行 Projection
type ShadowEngine struct {
	history []Transaction
	cursor  int
	nextID  int
}

// NewShadowEngine 创建新的影子引擎
func NewShadowEngine() *ShadowEngine {
	return &ShadowEngine{
		history: make([]Transaction, 0),
		cursor:  -1,
		nextID:  1,
	}
}

// ApplyIntent 应用意图（Shadow 模式：只记录，不执行）
func (e *ShadowEngine) ApplyIntent(intent Intent, resolver AnchorResolver, projection Projection) (*Verdict, error) {
	// 阶段 2：我们只是记录 Intent，不真正执行
	// 这是"影子模式"的核心：观察但不干预

	txID := TransactionID(fmt.Sprintf("tx-%d", e.nextID))
	e.nextID++

	// 创建一个空的事务（阶段 2 暂不生成真实 Fact）
	tx := Transaction{
		ID:        txID,
		Facts:     []Fact{}, // 阶段 3 才会填充
		Safety:    SafetyExact,
		CreatedAt: time.Now(),
		Applied:   false, // Shadow 模式不执行
	}

	// 记录到历史（但不执行）
	e.history = append(e.history[:e.cursor+1], tx)
	e.cursor++

	return &Verdict{
		Kind:    VerdictSkipped, // Shadow 模式总是 Skipped
		Safety:  SafetyExact,
		Message: fmt.Sprintf("Shadow mode: Intent recorded but not applied (tx: %s)", txID),
	}, nil
}

// Undo 撤销（Shadow 模式：不实现）
func (e *ShadowEngine) Undo() (*Verdict, error) {
	return &Verdict{
		Kind:    VerdictSkipped,
		Message: "Shadow mode: Undo not implemented",
	}, nil
}

// Redo 重做（Shadow 模式：不实现）
func (e *ShadowEngine) Redo() (*Verdict, error) {
	return &Verdict{
		Kind:    VerdictSkipped,
		Message: "Shadow mode: Redo not implemented",
	}, nil
}
