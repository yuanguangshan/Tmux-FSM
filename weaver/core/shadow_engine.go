package core

import (
	"fmt"
	"time"
)

// ShadowEngine 核心执行引擎
// 负责处理 Intent，生成并应用 Transaction，维护 History
type ShadowEngine struct {
	planner Planner
	history History
}

func NewShadowEngine(planner Planner) *ShadowEngine {
	return &ShadowEngine{
		planner: planner,
		history: NewInMemoryHistory(100),
	}
}

func (e *ShadowEngine) ApplyIntent(intent Intent, resolver AnchorResolver, projection Projection) (*Verdict, error) {
	// 1. Handle Undo/Redo explicitly
	kind := intent.GetKind()
	if kind == IntentUndo {
		return e.performUndo(projection)
	}
	if kind == IntentRedo {
		return e.performRedo(projection)
	}

	// 2. Plan: Generate Facts
	facts, inverseFacts, err := e.planner.Build(intent, intent.GetPaneID())
	if err != nil {
		return nil, err
	}

	// 3. Create Transaction
	txID := TransactionID(fmt.Sprintf("tx-%d", time.Now().UnixNano()))
	tx := &Transaction{
		ID:           txID,
		Intent:       intent,
		Facts:        facts,
		InverseFacts: inverseFacts,
		Safety:       SafetyExact,
		Timestamp:    time.Now().Unix(),
	}

	// 4. Project: Execute
	if err := projection.Apply(nil, facts); err != nil {
		return nil, err
	}
	tx.Applied = true

	// 5. Update History
	// 只有产生副作用的操作才记录历史
	if len(facts) > 0 {
		e.history.Push(tx)
	}

	return &Verdict{
		Kind:        VerdictApplied,
		Message:     "Applied via Smart Projection",
		Transaction: tx,
		Safety:      SafetyExact,
	}, nil
}

func (e *ShadowEngine) performUndo(projection Projection) (*Verdict, error) {
	tx := e.history.PopUndo()
	if tx == nil {
		return &Verdict{Kind: VerdictSkipped, Message: "Nothing to undo"}, nil
	}

	// Apply Inverse Facts
	if err := projection.Apply(nil, tx.InverseFacts); err != nil {
		// Undo failure is critical.
		// For now, return error.
		return nil, err
	}

	// Move to Redo Stack
	e.history.AddRedo(tx)

	return &Verdict{
		Kind:        VerdictApplied,
		Message:     fmt.Sprintf("Undone tx: %s", tx.ID),
		Transaction: tx,
	}, nil
}

func (e *ShadowEngine) performRedo(projection Projection) (*Verdict, error) {
	tx := e.history.PopRedo()
	if tx == nil {
		return &Verdict{Kind: VerdictSkipped, Message: "Nothing to redo"}, nil
	}

	// Apply Facts (Original Facts)
	if err := projection.Apply(nil, tx.Facts); err != nil {
		return nil, err
	}

	// Restore to Undo Stack
	e.history.PushBack(tx)

	return &Verdict{
		Kind:        VerdictApplied,
		Message:     fmt.Sprintf("Redone tx: %s", tx.ID),
		Transaction: tx,
	}, nil
}

// GetHistory 获取历史管理器 (用于 Reverse Bridge)
func (e *ShadowEngine) GetHistory() History {
	return e.history
}
