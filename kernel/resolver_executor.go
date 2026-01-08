package kernel

import (
	"tmux-fsm/intent"
)

// ResolverExecutor 基于新Resolver的意图执行器
// 实现 IntentExecutor 接口
// NOTE: 这是一个过渡性实现，将来会被 TransactionRunner 完全替代
type ResolverExecutor struct {
	// 暂时保留为空结构
}

// NewResolverExecutor 创建新的基于Resolver的执行器
func NewResolverExecutor() *ResolverExecutor {
	return &ResolverExecutor{}
}

// Process 实现 IntentExecutor 接口
// NOTE: 当前实现为空，等待集成新的 Transaction 系统
func (re *ResolverExecutor) Process(i *intent.Intent) error {
	// TODO: 集成 TransactionRunner
	// 1. 将 intent.Intent 转换为 ResolvedOperation
	// 2. 创建 Transaction
	// 3. 使用 TransactionRunner.Apply
	_ = i
	return nil
}
