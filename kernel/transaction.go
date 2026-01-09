package kernel

import (
	"fmt"
	"tmux-fsm/editor"
	"tmux-fsm/types"
)

// TransactionRunner 事务执行器
// 负责在 ExecutionContext 中执行 Transaction
type TransactionRunner struct {
	ctx *editor.ExecutionContext
}

// NewTransactionRunner 创建新的事务执行器
func NewTransactionRunner(ctx *editor.ExecutionContext) *TransactionRunner {
	return &TransactionRunner{
		ctx: ctx,
	}
}

// Apply 应用事务（正向执行）
func (tr *TransactionRunner) Apply(tx *types.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	// 收集所有操作（用于更新 selections）
	ops := make([]editor.ResolvedOperation, 0, len(tx.Records))

	// 执行所有操作
	for _, record := range tx.Records {
		if err := editor.ApplyResolvedOperation(tr.ctx, record.ResolvedOp); err != nil {
			return fmt.Errorf("failed to apply operation: %w", err)
		}
		ops = append(ops, record.ResolvedOp)
	}

	// 更新 selections（在所有操作执行完成后）
	tr.updateSelectionsAfterOps(ops)

	return nil
}

// Undo 撤销事务（反向执行）
func (tr *TransactionRunner) Undo(tx *types.Transaction) error {
	return fmt.Errorf("undo not supported: inverse execution not implemented")
}

// Repeat 重复执行事务（用于 . repeat）
func (tr *TransactionRunner) Repeat(tx *types.Transaction) error {
	// Repeat 与 Apply 逻辑相同
	return tr.Apply(tx)
}

// updateSelectionsAfterOps 在操作执行后更新选区
func (tr *TransactionRunner) updateSelectionsAfterOps(ops []editor.ResolvedOperation) {
	if len(ops) == 0 {
		return
	}

	// 按 BufferID 分组操作
	opsByBuffer := make(map[editor.BufferID][]editor.ResolvedOperation)
	for _, op := range ops {
		opsByBuffer[op.BufferID] = append(opsByBuffer[op.BufferID], op)
	}

	// 对每个受影响的 buffer 更新其 selections
	for bufferID, bufferOps := range opsByBuffer {
		currentSels := tr.ctx.Selections.Get(bufferID)
		updatedSels := editor.UpdateSelections(currentSels, bufferOps)
		tr.ctx.Selections.Set(bufferID, updatedSels)
	}
}
