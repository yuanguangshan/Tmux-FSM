package planner

import (
	"tmux-fsm/fsm"
	"tmux-fsm/intent"
)

// Grammar 是 Stage‑4 Vim Grammar
type Grammar struct {
	count      int
	pendingOp  *intent.OperatorKind
	lastIntent *intent.Intent
}

// NewGrammar 创建 Grammar 实例
func NewGrammar() *Grammar {
	return &Grammar{}
}

// Consume 消费一个 FSM RawToken，必要时产生 Intent
func (g *Grammar) Consume(tok fsm.RawToken) *intent.Intent {
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
func (g *Grammar) consumeKey(key string) *intent.Intent {

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
			)
			g.reset()
			g.remember(intent)
			return intent
		}

		// standalone motion (move)
		intent := makeMoveIntent(motion, max(g.count, 1))
		g.reset()
		g.remember(intent)
		return intent
	}

	// unknown key → reset
	g.reset()
	return nil
}

// ---------- Intent builders ----------

func makeMoveIntent(m intent.MotionKind, count int) *intent.Intent {
	return &intent.Intent{
		Kind:  intent.IntentMove,
		Count: count,
		Meta: map[string]interface{}{
			"motion": m,
		},
	}
}

func makeOpMotionIntent(op intent.OperatorKind, m intent.MotionKind, count int) *intent.Intent {
	return &intent.Intent{
		Kind:  intent.IntentOperator,
		Count: count,
		Meta: map[string]interface{}{
			"operator": op,
			"motion":   m,
		},
	}
}

func makeLineIntent(op intent.OperatorKind, count int) *intent.Intent {
	return &intent.Intent{
		Kind:  intent.IntentOperator,
		Count: count,
		Meta: map[string]interface{}{
			"operator": op,
			"motion":   intent.MotionLine,
		},
	}
}

// ---------- helpers ----------

func (g *Grammar) reset() {
	g.count = 0
	g.pendingOp = nil
}

func (g *Grammar) remember(i *intent.Intent) {
	g.lastIntent = cloneIntent(i)
}

func cloneIntent(i *intent.Intent) *intent.Intent {
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

func parseOperator(key string) (intent.OperatorKind, bool) {
	switch key {
	case "d":
		return intent.OpDelete, true
	case "y":
		return intent.OpYank, true
	case "c":
		return intent.OpChange, true
	default:
		return 0, false
	}
}

func parseMotion(key string) (intent.MotionKind, bool) {
	switch key {
	case "h", "l":
		return intent.MotionChar, true
	case "j", "k":
		return intent.MotionLine, true
	case "w", "b", "e":
		return intent.MotionWord, true
	case "$", "0":
		return intent.MotionChar, true
	case "G":
		return intent.MotionGoto, true
	// "g" 不作为 motion，因为它是前缀键
	default:
		return 0, false
	}
}