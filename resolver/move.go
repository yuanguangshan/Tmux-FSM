package resolver

import "tmux-fsm/intent"

// resolveMove 解析移动意图
func (r *Resolver) resolveMove(i *intent.Intent) error {
	if i.Motion == nil {
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

	// 移动光标到指定范围
	return r.engine.MoveCursor(rng)
}