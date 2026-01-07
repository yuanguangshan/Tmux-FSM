package planner

import (
	"tmux-fsm/fsm"
	intentPkg "tmux-fsm/intent"
)

// Grammar 是 Stage‑4 Vim Grammar
type Grammar struct {
	count      int
	pendingOp  *intentPkg.OperatorKind
	lastIntent *intentPkg.Intent
}

// NewGrammar 创建 Grammar 实例
func NewGrammar() *Grammar {
	return &Grammar{}
}

// Consume 消费一个 FSM RawToken，必要时产生 Intent
func (g *Grammar) Consume(tok fsm.RawToken) *intentPkg.Intent {
	switch tok.Kind {

	case fsm.TokenDigit:
		g.count = g.count*10 + int(tok.Value[0]-'0')
		return nil

	case fsm.TokenRepeat:
		if g.lastIntent != nil {
			return cloneIntent(g.lastIntent)
		}
		return nil

	case fsm.TokenKey:
		return g.consumeKey(tok.Value)

	case fsm.TokenSystem:
		// 系统事件，重置状态
		if tok.Value == "reset" || tok.Value == "exit" || tok.Value == "enter" {
			g.reset()
		}
		return nil
	}

	return nil
}

// consumeKey 处理普通按键
func (g *Grammar) consumeKey(key string) *intentPkg.Intent {

	// 1️⃣ operator
	if op, ok := parseOperator(key); ok {
		// dd / yy
		if g.pendingOp != nil && *g.pendingOp == op {
			intent := makeLineIntent(op, max(g.count, 1))
			g.reset()
			g.remember(intent)
			return intent
		}

		g.pendingOp = &op
		return nil
	}

	// 2️⃣ motion
	if motion, ok := parseMotion(key); ok {

		// op + motion
		if g.pendingOp != nil {
			intent := makeOpMotionIntent(
				*g.pendingOp,
				motion,
				max(g.count, 1),
				key,
			)
			g.reset()
			g.remember(intent)
			return intent
		}

		// standalone motion (move)
		intent := makeMoveIntent(motion, max(g.count, 1), key)
		g.reset()
		g.remember(intent)
		return intent
	}

	// unknown key → reset
	g.reset()
	return nil
}

// ---------- Intent builders ----------

func makeMoveIntent(m intentPkg.MotionKind, count int, key string) *intentPkg.Intent {
	intentObj := &intentPkg.Intent{
		Kind:  intentPkg.IntentMove,
		Count: count,
		Meta: map[string]interface{}{
			"motion": m,
		},
	}

	// 设置Target
	intentObj.Target = intentPkg.SemanticTarget{
		Kind: intentPkg.TargetChar,
	}

	// 特殊处理 $ 和 0
	if key == "$" {
		intentObj.Meta["motion_special"] = "line_end"
		intentObj.Target.Kind = intentPkg.TargetLine
		intentObj.Target.Scope = "end"
	} else if key == "0" {
		intentObj.Meta["motion_special"] = "line_start"
		intentObj.Target.Kind = intentPkg.TargetLine
		intentObj.Target.Scope = "start"
	} else {
		// 根据MotionKind设置Target
		switch m {
		case intentPkg.MotionChar:
			intentObj.Target.Kind = intentPkg.TargetChar
			// 根据具体按键设置方向
			if key == "h" {
				intentObj.Target.Direction = "left"
			} else if key == "l" {
				intentObj.Target.Direction = "right"
			} else if key == "j" {
				intentObj.Target.Direction = "down"
			} else if key == "k" {
				intentObj.Target.Direction = "up"
			}
		case intentPkg.MotionLine:
			intentObj.Target.Kind = intentPkg.TargetLine
		case intentPkg.MotionWord:
			intentObj.Target.Kind = intentPkg.TargetWord
		}
	}

	return intentObj
}

func makeOpMotionIntent(op intentPkg.OperatorKind, m intentPkg.MotionKind, count int, key string) *intentPkg.Intent {
	intentObj := &intentPkg.Intent{
		Kind:  intentPkg.IntentOperator,
		Count: count,
		Meta: map[string]interface{}{
			"operator": op,
			"motion":   m,
		},
	}

	// 设置Target
	intentObj.Target = intentPkg.SemanticTarget{
		Kind: intentPkg.TargetChar, // 默认为字符级移动
	}

	// 特殊处理 $ 和 0
	if key == "$" {
		intentObj.Meta["motion_special"] = "line_end"
		intentObj.Target.Kind = intentPkg.TargetLine
		intentObj.Target.Scope = "end"
	} else if key == "0" {
		intentObj.Meta["motion_special"] = "line_start"
		intentObj.Target.Kind = intentPkg.TargetLine
		intentObj.Target.Scope = "start"
	}

	return intentObj
}

func makeLineIntent(op intentPkg.OperatorKind, count int) *intentPkg.Intent {
	return &intentPkg.Intent{
		Kind:  intentPkg.IntentOperator,
		Count: count,
		Meta: map[string]interface{}{
			"operator": op,
			"motion":   intentPkg.MotionLine,
		},
	}
}

// ---------- helpers ----------

func (g *Grammar) reset() {
	g.count = 0
	g.pendingOp = nil
}

func (g *Grammar) remember(i *intentPkg.Intent) {
	g.lastIntent = cloneIntent(i)
}

func cloneIntent(i *intentPkg.Intent) *intentPkg.Intent {
	c := *i
	if i.Meta != nil {
		c.Meta = make(map[string]interface{})
		for k, v := range i.Meta {
			c.Meta[k] = v
		}
	}
	return &c
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ---------- key parsing (Grammar owns Vim) ----------

func parseOperator(key string) (intentPkg.OperatorKind, bool) {
	switch key {
	case "d":
		return intentPkg.OpDelete, true
	case "y":
		return intentPkg.OpYank, true
	case "c":
		return intentPkg.OpChange, true
	default:
		return 0, false
	}
}

func parseMotion(key string) (intentPkg.MotionKind, bool) {
	switch key {
	case "h", "l":
		return intentPkg.MotionChar, true
	case "j", "k":
		return intentPkg.MotionLine, true
	case "w", "b", "e":
		return intentPkg.MotionWord, true
	case "$":
		return intentPkg.MotionChar, true
	case "0":
		return intentPkg.MotionChar, true
	case "G":
		return intentPkg.MotionGoto, true
	// "g" 不作为 motion，因为它是前缀键
	default:
		return 0, false
	}
}