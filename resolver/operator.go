package resolver

import "tmux-fsm/intent"

// resolveOperatorWithContext 解析操作符意图（带上下文）
func (r *Resolver) resolveOperatorWithContext(i *intent.Intent, ctx ExecContext) error {
	if i.Operator == nil || i.Motion == nil {
		return nil
	}

	// 如果是查找动作，记录为 lastFind
	if i.Motion.Kind == intent.MotionFind && i.Motion.Find != nil {
		r.lastFind = i.Motion.Find
	}

	// 使用引擎计算运动范围
	rng, err := r.engine.ComputeMotion(i.Motion)
	if err != nil {
		return err
	}

	var execErr error

	// 根据操作符执行相应操作
	switch *i.Operator {
	case intent.OpDelete:
		execErr = r.engine.DeleteRange(rng)
	case intent.OpYank:
		execErr = r.engine.YankRange(rng)
	case intent.OpChange:
		execErr = r.engine.ChangeRange(rng)
	}

	// 如果执行成功且不是来自重复操作，则记录为可重复操作
	if execErr == nil && !ctx.FromRepeat {
		r.lastRepeat = &RepeatableAction{
			Operator: i.Operator,
			Motion:   i.Motion,
			Count:    i.Count,
		}
	}

	return execErr
}