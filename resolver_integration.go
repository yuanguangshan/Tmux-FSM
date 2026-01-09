package main

import (
	"errors"
)

// IntentOperate 表示操作意图
type IntentOperate struct {
	Operator *Operator
	Motion   *Motion
}

// OperatorKind 定义操作符类型
type OperatorKind int

const (
	OpNone OperatorKind = iota
	OpDelete
	OpYank
	OpChange
)

// Operator 表示操作符
type Operator struct {
	Kind OperatorKind
}

// ResolvedOperation 表示解析后的操作
type ResolvedOperation struct {
	Operator OperatorKind // OpDelete / OpChange / OpYank / OpNone
	Motion   MotionKind   // 原始 motion（用于 repeat / undo 语义）
	Count    int

	From Cursor
	To   Cursor

	Range *MotionRange // nil 表示纯移动
}

// Resolver 负责解析意图到具体操作
type Resolver struct {
	engine         *CursorEngine
	textObjectCalc *ConcreteTextObjectCalculator
}

// NewResolver 创建新的解析器
func NewResolver(engine *CursorEngine) *Resolver {
	return &Resolver{
		engine:         engine,
		textObjectCalc: NewConcreteTextObjectCalculator(engine.Buffer),
	}
}

// Resolve 解析意图
func (r *Resolver) Resolve(intent Intent) (*ResolvedOperation, error) {
	start := *r.engine.Cursor

	switch intent.Kind {
	case IntentMove:
		return r.resolveMove(&intent, start)
	case IntentDelete, IntentChange, IntentYank:
		return r.resolveOperator(&intent, start)
	}
	return nil, errors.New("unknown intent type")
}

// resolveMove 解析移动意图
func (r *Resolver) resolveMove(intent *Intent, start Cursor) (*ResolvedOperation, error) {
	if intent.Target.Kind == TargetTextObject {
		// 处理文本对象移动
		obj, err := ParseTextObject(intent.Target.Value)
		if err != nil {
			return nil, err
		}

		textRange, err := r.textObjectCalc.CalculateRange(*obj, start)
		if err != nil {
			return nil, err
		}

		return &ResolvedOperation{
			Operator: OpNone,
			Motion:   MotionKind(intent.Target.Kind),
			Count:    intent.Count,
			From:     start,
			To:       textRange.End, // 移动到文本对象的结束位置
			Range:    textRange,
		}, nil
	} else {
		// 处理普通移动
		motion := &Motion{
			Kind:  MotionKind(intent.Target.Kind),
			Count: intent.Count,
		}

		mr, err := r.engine.ComputeMotion(motion)
		if err != nil {
			return nil, err
		}

		// 虚拟计算终点（不改 cursor）
		end := start
		end.Row += mr.DeltaRow
		end.Col += mr.DeltaCol
		end.Row, end.Col = r.clampCursor(end.Row, end.Col)

		return &ResolvedOperation{
			Operator: OpNone,
			Motion:   motion.Kind,
			Count:    motion.Count,
			From:     start,
			To:       end,
			Range:    nil, // 移动通常不产生范围
		}, nil
	}
}

// clampCursor 限制光标位置
func (r *Resolver) clampCursor(row, col int) (int, int) {
	if r.engine.Buffer == nil {
		return row, col
	}

	row = clamp(row, 0, r.engine.Buffer.LineCount()-1)

	maxCol := 0
	if row >= 0 && row < r.engine.Buffer.LineCount() {
		maxCol = r.engine.Buffer.LineLength(row)
		if maxCol > 0 {
			maxCol-- // Length 是实际长度，所以最大索引是 Length-1
		}
	}
	col = clamp(col, 0, maxCol)

	return row, col
}

// resolveOperator 解析操作意图
func (r *Resolver) resolveOperator(intent *Intent, start Cursor) (*ResolvedOperation, error) {
	var opKind OperatorKind = OpNone
	switch intent.Kind {
	case IntentDelete:
		opKind = OpDelete
	case IntentChange:
		opKind = OpChange
	case IntentYank:
		opKind = OpYank
	}

	var rng *MotionRange

	if intent.Target.Kind == TargetTextObject {
		// 处理文本对象操作
		obj, err := ParseTextObject(intent.Target.Value)
		if err != nil {
			return nil, err
		}

		textRange, err := r.textObjectCalc.CalculateRange(*obj, start)
		if err != nil {
			return nil, err
		}

		rng = textRange
	} else {
		// 处理普通运动操作
		motion := &Motion{
			Kind:  MotionKind(intent.Target.Kind),
			Count: intent.Count,
		}

		mr, err := r.engine.ComputeMotion(motion)
		if err != nil {
			return nil, err
		}

		// 虚拟计算终点（不改 cursor）
		end := start
		end.Row += mr.DeltaRow
		end.Col += mr.DeltaCol
		end.Row, end.Col = r.clampCursor(end.Row, end.Col)

		rng = resolveRange(opKind, start, end, motion.Kind)
	}

	return &ResolvedOperation{
		Operator: opKind,
		Motion:   MotionKind(intent.Target.Kind),
		Count:    intent.Count,
		From:     start,
		To:       start, // 操作后光标位置可能不同，这里先设置为起始位置
		Range:    rng,
	}, nil
}

// resolveRange 计算操作范围
func resolveRange(op OperatorKind, from Cursor, to Cursor, motion MotionKind) *MotionRange {
	switch motion {
	case MotionWordForward:
		switch op {
		case OpDelete, OpYank:
			return &MotionRange{Start: from, End: to}
		case OpChange:
			// Vim: cw 不包含 word 后的空白
			adjusted := to
			adjusted.Col-- // 简化版
			return &MotionRange{Start: from, End: adjusted}
		}
	}

	// fallback
	return &MotionRange{Start: from, End: to}
}
