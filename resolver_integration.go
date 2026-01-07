package main

import "errors"

// IntentOperate 表示操作意图
type IntentOperate struct {
	Operator *Operator
	Motion   *Motion
}

// OperatorKind 定义操作符类型
type OperatorKind int

const (
	OpDelete OperatorKind = iota
	OpYank
	OpChange
)

// Operator 表示操作符
type Operator struct {
	Kind OperatorKind
}

// Resolver 负责解析意图到具体操作
type Resolver struct {
	engine *CursorEngine
}

// NewResolver 创建新的解析器
func NewResolver(engine *CursorEngine) *Resolver {
	return &Resolver{
		engine: engine,
	}
}

// Resolve 解析意图
func (r *Resolver) Resolve(intent Intent) error {
	switch intent.Kind {
	case IntentMove:
		return r.resolveMove(&intent)
	case IntentDelete, IntentChange, IntentYank:
		// 处理操作符意图
		return r.resolveOperator(&intent)
	}
	return errors.New("unknown intent type")
}

// resolveMove 解析移动意图
func (r *Resolver) resolveMove(intent *Intent) error {
	// 从 Intent 中提取 Motion 信息
	motion := &Motion{
		Kind:  MotionKind(intent.Target.Kind), // 假设 Target.Kind 映射到 MotionKind
		Count: intent.Count,
	}

	result, err := r.engine.ComputeMotion(motion)
	if err != nil {
		return err
	}
	return r.engine.MoveCursor(result)
}

// resolveOperator 解析操作意图
func (r *Resolver) resolveOperator(intent *Intent) error {
	// 从 Intent 中提取 Motion 信息
	motion := &Motion{
		Kind:  MotionKind(intent.Target.Kind),
		Count: intent.Count,
	}

	mr, err := r.engine.ComputeMotion(motion)
	if err != nil {
		return err
	}

	switch intent.Kind {
	case IntentDelete:
		if mr.Range != nil {
			return r.engine.DeleteRange(mr.Range)
		} else {
			// 如果没有范围，执行单个字符删除
			return r.engine.MoveCursor(mr)
		}
	}

	return errors.New("operator not implemented")
}

// resolveOperate 解析操作意图
func (r *Resolver) resolveOperate(op *IntentOperate) error {
	mr, err := r.engine.ComputeMotion(op.Motion)
	if err != nil {
		return err
	}
	if mr.Range == nil {
		return errors.New("operator requires range")
	}

	switch op.Operator.Kind {
	case OpDelete:
		return r.engine.DeleteRange(mr.Range)
	}
	return errors.New("operator not implemented")
}