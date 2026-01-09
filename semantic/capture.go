package semantic

//
// ─────────────────────────────────────────────────────────────
//  Anchor & Range
// ─────────────────────────────────────────────────────────────
//

// Anchor 描述一个稳定的语义锚点
type Anchor struct {
	PaneID string
	Line   int
	Col    int
	Hash   string // 用于弱一致性校验（可选）
}

// Range 表示一个语义范围
type Range struct {
	Start Anchor
	End   Anchor
	Text  string // 捕获时已知的文本
}

// ContainsFact 检查一个事实是否包含在当前范围内
// TODO: 实现实际的逻辑
func (r Range) ContainsFact(fact Fact) bool {
	// 这是一个占位符实现，需要根据实际的语义定义来判断
	// 例如，比较 fact 的 Anchor 和 Range 是否落在 r.Start 和 r.End 之间
	return true
}

//
// ─────────────────────────────────────────────────────────────
//  Motion
// ─────────────────────────────────────────────────────────────
//

// MotionKind 动作类型（强类型）
type MotionKind int

const (
	MotionWordForward MotionKind = iota
	MotionLine
)

// Motion 描述一个语义动作
type Motion struct {
	Kind  MotionKind
	Count int
}

//
// ─────────────────────────────────────────────────────────────
//  Fact Interface
// ─────────────────────────────────────────────────────────────
//

// Fact 表示一个可逆的语义事实
type Fact interface {
	Kind() FactKind
	Inverse() Fact

	Anchor() Anchor
	Range() (Range, bool)
	Text() string
}

//
// ─────────────────────────────────────────────────────────────
//  FactKind
// ─────────────────────────────────────────────────────────────
//

type FactKind int

const (
	FactInsert FactKind = iota
	FactDelete
	FactReplace
	FactMove
)

//
// ─────────────────────────────────────────────────────────────
//  BaseFact (immutable)
// ─────────────────────────────────────────────────────────────
//

type baseFact struct {
	kind   FactKind
	anchor Anchor
	rng    *Range
	text   string
}

func (f baseFact) Kind() FactKind {
	return f.kind
}

func (f baseFact) Anchor() Anchor {
	return f.anchor
}

func (f baseFact) Range() (Range, bool) {
	if f.rng == nil {
		return Range{}, false
	}
	return *f.rng, true
}

func (f baseFact) Text() string {
	return f.text
}

//
// ─────────────────────────────────────────────────────────────
//  Insert
// ─────────────────────────────────────────────────────────────
//

type InsertFact struct {
	baseFact
}

func (f InsertFact) Inverse() Fact {
	return DeleteFact{
		baseFact: baseFact{
			kind:   FactDelete,
			anchor: f.anchor,
			rng:    f.rng,
			text:   f.text,
		},
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Delete
// ─────────────────────────────────────────────────────────────
//

type DeleteFact struct {
	baseFact
}

func (f DeleteFact) Inverse() Fact {
	return InsertFact{
		baseFact: baseFact{
			kind:   FactInsert,
			anchor: f.anchor,
			rng:    f.rng,
			text:   f.text,
		},
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Replace
// ─────────────────────────────────────────────────────────────
//

type ReplaceFact struct {
	baseFact
	oldText string
}

func (f ReplaceFact) Inverse() Fact {
	return ReplaceFact{
		baseFact: baseFact{
			kind:   FactReplace,
			anchor: f.anchor,
			rng:    f.rng,
			text:   f.oldText,
		},
		oldText: f.text,
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Move
// ─────────────────────────────────────────────────────────────
//

type MoveFact struct {
	baseFact
	from Anchor
	to   Anchor
}

func (f MoveFact) Inverse() Fact {
	return MoveFact{
		baseFact: baseFact{
			kind:   FactMove,
			anchor: f.anchor,
		},
		from: f.to,
		to:   f.from,
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Capture (Pure Semantic)
// ─────────────────────────────────────────────────────────────
//

// CaptureAnchor 捕获锚点（纯函数）
func CaptureAnchor(a Anchor) Anchor {
	return a
}

// CaptureRange 捕获一个语义范围（不访问文本）
func CaptureRange(anchor Anchor, motion Motion, knownText string) Range {
	start := anchor
	end := anchor

	switch motion.Kind {
	case MotionWordForward:
		end.Col += max(1, motion.Count) * 5 // 语义步进
	case MotionLine:
		end.Col = 1 << 30 // 语义行尾
	}

	return Range{
		Start: start,
		End:   end,
		Text:  knownText,
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Capture Facts
// ─────────────────────────────────────────────────────────────
//

func CaptureInsert(anchor Anchor, text string) Fact {
	return InsertFact{
		baseFact: baseFact{
			kind:   FactInsert,
			anchor: anchor,
			text:   text,
		},
	}
}

func CaptureDelete(rng Range) Fact {
	return DeleteFact{
		baseFact: baseFact{
			kind:   FactDelete,
			anchor: rng.Start,
			rng:    &rng,
			text:   rng.Text,
		},
	}
}

func CaptureReplace(rng Range, text string) Fact {
	return ReplaceFact{
		baseFact: baseFact{
			kind:   FactReplace,
			anchor: rng.Start,
			rng:    &rng,
			text:   text,
		},
		oldText: rng.Text,
	}
}

func CaptureMove(from, to Anchor) Fact {
	return MoveFact{
		baseFact: baseFact{
			kind:   FactMove,
			anchor: from,
		},
		from: from,
		to:   to,
	}
}

//
// ─────────────────────────────────────────────────────────────
//  Helpers
// ─────────────────────────────────────────────────────────────
//

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
