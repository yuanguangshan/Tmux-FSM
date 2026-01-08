package semantic

// Anchor 描述文本位置的锚点
type Anchor struct {
	PaneID string
	Line   int
	Col    int
	Hash   string
}

// Motion 动作类型
type Motion struct {
	Kind  string
	Count int
}

// Range 表示文本范围
type Range struct {
	Start Anchor
	End   Anchor
	Text  string
}

// Fact 表示一个语义事实
type Fact interface {
	Kind() string
	Inverse() Fact
	GetAnchor() Anchor
	GetRange() Range
	GetText() string
}

// BaseFact 基础事实结构
type BaseFact struct {
	kind   string
	anchor Anchor
	rng    Range
	text   string
}

func (f *BaseFact) Kind() string {
	return f.kind
}

func (f *BaseFact) GetAnchor() Anchor {
	return f.anchor
}

func (f *BaseFact) GetRange() Range {
	return f.rng
}

func (f *BaseFact) GetText() string {
	return f.text
}

// DeleteFact 删除事实
type DeleteFact struct {
	BaseFact
}

func (f *DeleteFact) Inverse() Fact {
	return &InsertFact{
		BaseFact: BaseFact{
			kind:   "insert",
			anchor: f.anchor,
			rng:    f.rng,
			text:   f.text,
		},
	}
}

// InsertFact 插入事实
type InsertFact struct {
	BaseFact
}

func (f *InsertFact) Inverse() Fact {
	return &DeleteFact{
		BaseFact: BaseFact{
			kind:   "delete",
			anchor: f.anchor,
			rng:    f.rng,
			text:   f.text,
		},
	}
}

// ReplaceFact 替换事实
type ReplaceFact struct {
	BaseFact
	oldText string
}

func (f *ReplaceFact) Inverse() Fact {
	return &ReplaceFact{
		BaseFact: BaseFact{
			kind:   "replace",
			anchor: f.anchor,
			rng:    f.rng,
			text:   f.oldText,
		},
		oldText: f.text,
	}
}

// MoveFact 移动事实
type MoveFact struct {
	BaseFact
	from Anchor
	to   Anchor
}

func (f *MoveFact) Inverse() Fact {
	return &MoveFact{
		BaseFact: BaseFact{
			kind:   "move",
			anchor: f.anchor,
			rng:    f.rng,
			text:   f.text,
		},
		from: f.to,
		to:   f.from,
	}
}

// CaptureAnchor 纯语义锚点捕获（不产生副作用）
func CaptureAnchor(a Anchor) Anchor {
	return a
}

// CaptureRange 捕获一个范围（纯语义，不访问外部状态）
func CaptureRange(anchor Anchor, motion Motion, text string) Range {
	start := anchor
	end := anchor

	switch motion.Kind {
	case "word_forward":
		// 模拟单词前进的范围计算
		end.Col += 5 // 模拟前进到下一个单词
	case "line":
		// 整行范围
		end.Col = 1 << 30 // 语义行尾
	}

	return Range{
		Start: start,
		End:   end,
		Text:  text, // 由上层提供的已知文本
	}
}

// CaptureDelete 捕获删除操作
func CaptureDelete(rng Range) Fact {
	return &DeleteFact{
		BaseFact: BaseFact{
			kind:   "delete",
			anchor: rng.Start,
			rng:    rng,
			text:   rng.Text,
		},
	}
}

// CaptureInsert 捕获插入操作
func CaptureInsert(anchor Anchor, text string) Fact {
	return &InsertFact{
		BaseFact: BaseFact{
			kind:   "insert",
			anchor: anchor,
			text:   text,
		},
	}
}

// CaptureReplace 捕获替换操作
func CaptureReplace(rng Range, text string) Fact {
	return &ReplaceFact{
		BaseFact: BaseFact{
			kind:   "replace",
			anchor: rng.Start,
			rng:    rng,
			text:   text,
		},
		oldText: rng.Text,
	}
}

// CaptureMove 捕获移动操作
func CaptureMove(from, to Anchor) Fact {
	return &MoveFact{
		BaseFact: BaseFact{
			kind:   "move",
			anchor: from,
		},
		from: from,
		to:   to,
	}
}