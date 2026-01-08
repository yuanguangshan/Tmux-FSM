package resolver

import "tmux-fsm/intent"

// resolveRepeatWithContext 解析重复意图（带上下文）
func (r *Resolver) resolveRepeatWithContext(i *intent.Intent, ctx ExecContext) error {
	if r.lastRepeat == nil {
		return nil
	}

	// 创建重复操作的意图
	repeatIntent := &intent.Intent{
		Kind:     intent.IntentOperator,
		Operator: r.lastRepeat.Operator,
		Motion:   r.lastRepeat.Motion,
		Count:    r.lastRepeat.Count,
	}

	// 使用新的上下文（标记为来自重复）
	newCtx := ExecContext{
		FromRepeat: true,
		FromMacro:  ctx.FromMacro,
		FromUndo:   ctx.FromUndo,
	}

	// 重新执行最后一次可重复操作
	return r.ResolveWithContext(repeatIntent, newCtx)
}

// repeatFind 处理 ; 和 , 重复查找操作
func (r *Resolver) repeatFind(reverse bool) error {
	if r.lastFind == nil {
		return nil
	}

	// 复制 lastFind 并根据 reverse 参数调整方向
	find := *r.lastFind
	if reverse {
		if find.Direction == intent.FindForward {
			find.Direction = intent.FindBackward
		} else {
			find.Direction = intent.FindForward
		}
	}

	// 创建查找运动
	motion := &intent.Motion{
		Kind:  intent.MotionFind,
		Count: 1,
		Find:  &find,
	}

	// 计算范围并移动光标
	rng, err := r.engine.ComputeMotion(motion)
	if err != nil {
		return err
	}

	return r.engine.MoveCursor(rng)
}