package kernel

import (
	"tmux-fsm/editor"
	"tmux-fsm/intent"
)

// ResolverExecutor 基于新Resolver的意图执行器
// 实现 IntentExecutor 接口
type ResolverExecutor struct {
	resolver *editor.Resolver // 使用新的Resolver实现
}

// NewResolverExecutor 创建新的基于Resolver的执行器
func NewResolverExecutor(resolver *editor.Resolver) *ResolverExecutor {
	return &ResolverExecutor{
		resolver: resolver,
	}
}

// Process 实现 IntentExecutor 接口
// 将 intent.Intent 转换为新的 Intent 类型，然后使用新 Resolver 处理
func (re *ResolverExecutor) Process(i *intent.Intent) error {
	// 将旧的 intent.Intent 转换为新的 Intent 类型
	newIntent := convertOldIntentToNew(i)

	// 使用新的 Resolver 解析意图
	resolvedOp, err := re.resolver.Resolve(newIntent)
	if err != nil {
		return err
	}

	// TODO: 执行 ResolvedOperation
	// 这里需要将 ResolvedOperation 转换为实际的编辑操作
	_ = resolvedOp

	return nil
}

// convertOldIntentToNew 将旧的 intent.Intent 转换为新的 Intent 类型
func convertOldIntentToNew(old *intent.Intent) editor.Intent {
	// 这里需要映射旧的 Intent 类型到新的 Intent 类型
	newIntent := editor.Intent{
		Kind:  editor.IntentKind(old.Kind),
		Count: old.Count,
		Target: editor.Target{
			Kind:      editor.TargetKind(old.Target.Kind),
			Direction: old.Target.Direction,
			Scope:     old.Target.Scope,
			Value:     old.Target.Value,
		},
	}

	return newIntent
}