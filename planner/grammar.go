package planner

import (
	"tmux-fsm/fsm"
	intentPkg "tmux-fsm/intent"
)

// Grammar 是 Stage‑4 Vim Grammar
//
// ⚠️ Architecture rule:
// Grammar MUST NOT construct intent.Intent.
// Grammar ONLY produces intent.GrammarIntent.
// Promotion happens exclusively in Kernel via intent.Promote.
type Grammar struct {
	count             int
	pendingOp         *intentPkg.OperatorKind
	// 新增状态用于处理复杂 motion
	pendingMotion *MotionPendingInfo
	textObj       TextObjPending
}

// MotionPendingInfo 用于处理需要两个按键的 motion
type MotionPendingInfo struct {
	Kind        intentPkg.MotionKind
	FindDir     intentPkg.FindDirection
	FindTill    bool
}

const (
	MPNone = iota
	MPG      // g_
	MPF      // f{c}
	MPT      // t{c}
	MPBigF   // F{c}
	MPBigT   // T{c}
)

// TextObjPending 用于处理文本对象
type TextObjPending int

const (
	TOPNone TextObjPending = iota
	TOPInner
	TOPAround
)

// NewGrammar 创建 Grammar 实例
func NewGrammar() *Grammar {
	return &Grammar{}
}

// Consume 消费一个 FSM RawToken，必要时产生 GrammarIntent
func (g *Grammar) Consume(tok fsm.RawToken) *intentPkg.GrammarIntent {
	switch tok.Kind {

	case fsm.TokenDigit:
		g.count = g.count*10 + int(tok.Value[0]-'0')
		return nil

	case fsm.TokenRepeat:
		return &intentPkg.GrammarIntent{
			Kind: intentPkg.IntentRepeat,
		}

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
func (g *Grammar) consumeKey(key string) *intentPkg.GrammarIntent {
	// 优先处理 pending motion
	if g.pendingMotion != MPNone {
		return g.consumePendingMotion(key)
	}

	// 优先处理 text object
	if g.textObj != TOPNone {
		return g.consumeTextObject(key)
	}

	// 1️⃣ operator
	if op, ok := parseOperator(key); ok {
		// dd / yy
		if g.pendingOp != nil && *g.pendingOp == op {
			intent := makeLineGrammarIntent(op, max(g.count, 1))
			g.reset()
			g.rememberGrammar(intent)
			return intent
		}

		// 检查是否进入文本对象模式 (i 或 a)
		if key == "i" || key == "a" {
			if key == "i" {
				g.textObj = TOPInner
			} else {
				g.textObj = TOPAround
			}
			g.pendingOp = &op
			return nil
		}

		g.pendingOp = &op
		return nil
	}

	// 2️⃣ 检查是否是进入文本对象模式 (i 或 a)
	if key == "i" || key == "a" {
		if key == "i" {
			g.textObj = TOPInner
		} else {
			g.textObj = TOPAround
		}
		return nil
	}

	// 3️⃣ 检查是否是 motion 前缀
	if parseMotionPrefix(key) {
		switch key {
		case "g":
			g.pendingMotion = &MotionPendingInfo{
				Kind: intentPkg.MotionGoto,
			}
		case "f":
			g.pendingMotion = &MotionPendingInfo{
				Kind:     intentPkg.MotionFind,
				FindDir:  intentPkg.FindForward,
				FindTill: false,
			}
		case "t":
			g.pendingMotion = &MotionPendingInfo{
				Kind:     intentPkg.MotionFind,
				FindDir:  intentPkg.FindForward,
				FindTill: true,
			}
		case "F":
			g.pendingMotion = &MotionPendingInfo{
				Kind:     intentPkg.MotionFind,
				FindDir:  intentPkg.FindBackward,
				FindTill: false,
			}
		case "T":
			g.pendingMotion = &MotionPendingInfo{
				Kind:     intentPkg.MotionFind,
				FindDir:  intentPkg.FindBackward,
				FindTill: true,
			}
		}
		return nil
	}

	// 4️⃣ 检查是否是 motion
	if motion, ok := parseMotion(key); ok {
		// op + motion
		if g.pendingOp != nil {
			intent := makeOpMotionGrammarIntent(
				*g.pendingOp,
				motion,
				max(g.count, 1),
				key,
			)
			g.reset()
			return intent
		}

		// standalone motion (move)
		intent := makeMoveGrammarIntent(motion, max(g.count, 1), key)
		g.reset()
		return intent
	}

	// 5️⃣ 检查是否是模式切换键
	if mode := parseModeSwitch(key); mode != "" {
		// 模式切换暂时返回普通的 Intent，但我们需要重构
		// 为简化，这里先返回 nil，模式切换将通过其他方式处理
		g.reset()
		return nil
	}

	// 6️⃣ 检查是否是 find repeat 键
	if key == ";" {
		g.reset()
		return &intentPkg.GrammarIntent{
			Kind: intentPkg.IntentRepeatFind,
		}
	}
	if key == "," {
		g.reset()
		return &intentPkg.GrammarIntent{
			Kind: intentPkg.IntentRepeatFindReverse,
		}
	}

	// unknown key → reset
	g.reset()
	return nil
}

// parseModeSwitch 解析模式切换键
func parseModeSwitch(key string) string {
	switch key {
	case "i":
		return "insert"
	case "v":
		return "visual_char"
	case "V":
		return "visual_line"
	case "Escape", "C-c":
		return "normal"
	default:
		return ""
	}
}




// ---------- helpers ----------

func (g *Grammar) reset() {
	g.count = 0
	g.pendingOp = nil
	g.pendingMotion = nil
	g.textObj = TOPNone
}




// makeMoveGrammarIntent 创建移动 Grammar 意图
func makeMoveGrammarIntent(m intentPkg.MotionKind, count int, key string) *intentPkg.GrammarIntent {
	motion := &intentPkg.Motion{
		Kind:  m,
		Count: count,
	}

	// 特殊处理某些按键，设置更精确的 Motion 类型
	switch key {
	case "$":
		motion.Kind = intentPkg.MotionLine
	case "0", "^":
		motion.Kind = intentPkg.MotionLine
	case "G", "gg":
		motion.Kind = intentPkg.MotionGoto
	case "H", "M", "L":
		motion.Kind = intentPkg.MotionLine
	}

	return &intentPkg.GrammarIntent{
		Kind:   intentPkg.IntentMove,
		Count:  count,
		Motion: motion,
	}
}

// makeOpMotionGrammarIntent 创建操作+移动 Grammar 意图
func makeOpMotionGrammarIntent(op intentPkg.OperatorKind, m intentPkg.MotionKind, count int, key string) *intentPkg.GrammarIntent {
	motion := &intentPkg.Motion{
		Kind:  m,
		Count: count,
	}

	// 特殊处理某些按键，设置更精确的 Motion 类型
	switch key {
	case "$":
		motion.Kind = intentPkg.MotionLine
	case "0", "^":
		motion.Kind = intentPkg.MotionLine
	case "G", "gg":
		motion.Kind = intentPkg.MotionGoto
	case "H", "M", "L":
		motion.Kind = intentPkg.MotionLine
	}

	return &intentPkg.GrammarIntent{
		Kind:   intentPkg.IntentOperator,
		Count:  count,
		Motion: motion,
		Op:     &op,
	}
}

// makeLineGrammarIntent 创建行操作 Grammar 意图
func makeLineGrammarIntent(op intentPkg.OperatorKind, count int) *intentPkg.GrammarIntent {
	motion := &intentPkg.Motion{
		Kind:  intentPkg.MotionLine,
		Count: count,
	}

	return &intentPkg.GrammarIntent{
		Kind:   intentPkg.IntentOperator,
		Count:  count,
		Motion: motion,
		Op:     &op,
	}
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

// consumePendingMotion 处理需要两个按键的 motion
func (g *Grammar) consumePendingMotion(key string) *intentPkg.GrammarIntent {
	if g.pendingMotion != nil {
		switch g.pendingMotion.Kind {
		case intentPkg.MotionGoto:
			g.pendingMotion = nil
			if key == "g" {
				intent := makeMoveGrammarIntent(intentPkg.MotionGoto, max(g.count, 1), "gg")
				g.reset()
				g.rememberGrammar(intent)
				return intent
			}
			g.reset()
			return nil
		case intentPkg.MotionFind:
			// f{c}, t{c}, F{c}, T{c} 需要一个字符
			intent := makeFindGrammarIntent(g.pendingMotion, g.pendingOp, rune(key[0]), max(g.count, 1))
			g.pendingMotion = nil
			g.reset()
			g.rememberGrammar(intent)
			return intent
		default:
			g.reset()
			return nil
		}
	}
	g.reset()
	return nil
}

// consumeTextObject 处理文本对象
func (g *Grammar) consumeTextObject(key string) *intentPkg.GrammarIntent {
	objType := parseTextObject(key)
	if objType == intentPkg.Word && key != "w" { // Word 是默认值，需要检查是否真的匹配
		// 检查是否是有效的文本对象键
		switch key {
		case "w", "(", ")", "b", "[", "]", "{", "}", "B", "\"", "'", "`":
			// 这些都是有效的，继续
		default:
			g.reset()
			return nil
		}
	}

	intent := makeTextObjectGrammarIntent(g.pendingOp, g.textObj, objType, max(g.count, 1))
	g.reset()
	return intent
}

// makeTextObjectGrammarIntent 创建文本对象 Grammar 意图
func makeTextObjectGrammarIntent(op *intentPkg.OperatorKind, textObj TextObjPending, objType intentPkg.TextObjectKind, count int) *intentPkg.GrammarIntent {
	scope := intentPkg.Inner
	if textObj == TOPAround {
		scope = intentPkg.Around
	}

	textObject := &intentPkg.TextObject{
		Scope:  scope,
		Object: objType,
	}

	rangeMotion := &intentPkg.RangeMotion{
		Kind:       intentPkg.RangeTextObject,
		TextObject: textObject,
	}

	motion := &intentPkg.Motion{
		Kind:  intentPkg.MotionRange,
		Count: count,
		Range: rangeMotion,
	}

	if op != nil {
		return &intentPkg.GrammarIntent{
			Kind:   intentPkg.IntentOperator,
			Count:  count,
			Motion: motion,
			Op:     op,
		}
	} else {
		return &intentPkg.GrammarIntent{
			Kind:   intentPkg.IntentMove,
			Count:  count,
			Motion: motion,
		}
	}
}

// textObjectKindToString 将 TextObjectKind 转换为字符串（临时兼容）
func textObjectKindToString(kind intentPkg.TextObjectKind) string {
	switch kind {
	case intentPkg.Word:
		return "word"
	case intentPkg.Paren:
		return "paren"
	case intentPkg.Bracket:
		return "bracket"
	case intentPkg.Brace:
		return "brace"
	case intentPkg.QuoteSingle:
		return "quote_single"
	case intentPkg.QuoteDouble:
		return "quote_double"
	case intentPkg.Backtick:
		return "quote_backtick"
	default:
		return "word"
	}
}

// parseMotionPrefix 解析 motion 前缀
func parseMotionPrefix(key string) bool {
	switch key {
	case "g", "f", "F", "t", "T":
		return true
	default:
		return false
	}
}

// parseTextObject 解析文本对象
func parseTextObject(key string) intentPkg.TextObjectKind {
	switch key {
	case "w":
		return intentPkg.Word
	case "(":
		return intentPkg.Paren
	case ")":
		return intentPkg.Paren
	case "b":
		return intentPkg.Paren // b 也是括号的别名
	case "[":
		return intentPkg.Bracket
	case "]":
		return intentPkg.Bracket
	case "{":
		return intentPkg.Brace
	case "}":
		return intentPkg.Brace
	case "B":
		return intentPkg.Brace // B 也是大括号的别名
	case "\"":
		return intentPkg.QuoteDouble
	case "'":
		return intentPkg.QuoteSingle
	case "`":
		return intentPkg.Backtick
	default:
		return intentPkg.Word // 默认值
	}
}

// makeFindGrammarIntent 创建查找 Grammar 意图
func makeFindGrammarIntent(pending *MotionPendingInfo, op *intentPkg.OperatorKind, char rune, count int) *intentPkg.GrammarIntent {
	findMotion := &intentPkg.FindMotion{
		Char:      char,
		Direction: pending.FindDir,
		Till:      pending.FindTill,
	}

	motion := &intentPkg.Motion{
		Kind: intentPkg.MotionFind,
		Find: findMotion,
		Count: count,
	}

	// 修复：对于 FindMotion，Intent 应该是 Move 或 Operator，而不是 IntentFind
	// 根据是否有操作符来决定 Intent 类型
	if op != nil {
		// 如果有操作符，返回 Operator 类型
		return &intentPkg.GrammarIntent{
			Kind:   intentPkg.IntentOperator,
			Count:  count,
			Motion: motion,
			Op:     op,
		}
	} else {
		// 否则返回 Move 类型
		return &intentPkg.GrammarIntent{
			Kind:   intentPkg.IntentMove,
			Count:  count,
			Motion: motion,
		}
	}
}

// motionTypeToString 将 MotionPending 转换为字符串
func motionTypeToString(motionType MotionPending) string {
	switch motionType {
	case MPF:
		return "f"
	case MPBigF:
		return "F"
	case MPT:
		return "t"
	case MPBigT:
		return "T"
	default:
		return ""
	}
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
	case "w", "b", "e", "ge":
		return intentPkg.MotionWord, true
	case "$":
		return intentPkg.MotionChar, true
	case "0", "^":
		return intentPkg.MotionChar, true
	case "G":
		return intentPkg.MotionGoto, true
	case "H", "M", "L":
		return intentPkg.MotionLine, true
	default:
		return 0, false
	}
}