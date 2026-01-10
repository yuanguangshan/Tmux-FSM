package editor

// IntersectRanges 检查两个范围集合是否有交集
func IntersectRanges(a, b []TextRange) []TextRange {
	var results []TextRange
	for _, ra := range a {
		for _, rb := range b {
			if overlap, ok := rangeOverlap(ra, rb); ok {
				results = append(results, overlap)
			}
		}
	}
	return results
}

func rangeOverlap(a, b TextRange) (TextRange, bool) {
	// a.End <= b.Start or b.End <= a.Start
	if !a.Start.LessThan(b.End) || !b.Start.LessThan(a.End) {
		return TextRange{}, false
	}

	start := a.Start
	if b.Start.LessThan(start) {
		start = b.Start
	} else if a.Start.LessThan(b.Start) {
		start = b.Start
	}

	end := a.End
	if b.End.LessThan(end) {
		end = b.End
	}

	// Double check if start < end
	if !start.LessThan(end) {
		return TextRange{}, false
	}

	return TextRange{Start: start, End: end}, true
}

// IntersectSymbols 检查两个符号集合是否有交集
func IntersectSymbols(a, b []SymbolRef) []SymbolRef {
	var results []SymbolRef
	for _, sa := range a {
		for _, sb := range b {
			if sa.ID == sb.ID {
				results = append(results, sa)
			}
		}
	}
	return results
}

// IntersectEffects 检查两个影响集合是否有交集
func IntersectEffects(a, b []EffectKind) []EffectKind {
	var results []EffectKind
	m := make(map[EffectKind]bool)
	for _, e := range a {
		m[e] = true
	}
	for _, e := range b {
		if m[e] {
			results = append(results, e)
		}
	}
	return results
}

// EffectsConflict 判定影响集合是否冲突 (核心判定矩阵)
func EffectsConflict(a, b []EffectKind) bool {
	// 判定矩阵实现：
	// |        | Read | Write | Delete | Rename | Create |
	// |--------|------|-------|--------|--------|--------|
	// | Read   | No   | Yes   | Yes    | Yes    | No     |
	// | Write  | Yes  | Yes   | Yes    | Yes    | No     |
	// | Delete | Yes  | Yes   | Yes    | Yes    | No     |
	// | Rename | Yes  | Yes   | Yes    | Yes    | No     |
	// | Create | No   | No    | No     | No     | Yes*   |
	// *Create vs Create: 如果发生在同一语义槽点则冲突（由 Footprint.ConflictsWith 处理 Symbol/Range 交集）

	hasMutation := func(effects []EffectKind) bool {
		for _, e := range effects {
			if e == EffectWrite || e == EffectDelete || e == EffectRename || e == EffectCreate {
				return true
			}
		}
		return false
	}

	hasRead := func(effects []EffectKind) bool {
		for _, e := range effects {
			if e == EffectRead {
				return true
			}
		}
		return false
	}

	// 1. Read-Read 不冲突
	if !hasMutation(a) && !hasMutation(b) {
		return false
	}

	// 2. Mutation vs Read 冲突
	if (hasMutation(a) && hasRead(b)) || (hasMutation(b) && hasRead(a)) {
		return true
	}

	// 3. Mutation vs Mutation 冲突
	// 特殊处理：Create vs Create 在同一位置/符号下始终冲突
	// 其他 Mutation 对 Mutation 也始终冲突（Lost Update / Causality Break）
	return true
}

// ConflictsWith 判定两个 Footprint 是否冲突
func (a Footprint) ConflictsWith(b Footprint) (bool, ConflictReason, FootprintOverlap) {
	// 1. Buffer 层剪枝
	if !intersectBuffers(a.Buffers, b.Buffers) {
		return false, 0, FootprintOverlap{}
	}

	// 2. Symbol 冲突判定 (优先级更高)
	symbolOverlap := IntersectSymbols(a.Symbols, b.Symbols)
	if len(symbolOverlap) > 0 {
		if EffectsConflict(a.Effects, b.Effects) {
			return true, ConflictSemantic, FootprintOverlap{
				Symbols: symbolOverlap,
				Effects: IntersectEffects(a.Effects, b.Effects),
			}
		}
	}

	// 3. 空间冲突判定
	overlapRanges := IntersectRanges(a.Ranges, b.Ranges)
	if len(overlapRanges) > 0 {
		if EffectsConflict(a.Effects, b.Effects) {
			return true, ConflictSpatial, FootprintOverlap{
				Ranges:  overlapRanges,
				Effects: IntersectEffects(a.Effects, b.Effects),
			}
		}
	}

	return false, 0, FootprintOverlap{}
}

func intersectBuffers(a, b []BufferID) bool {
	m := make(map[BufferID]bool)
	for _, id := range a {
		m[id] = true
	}
	for _, id := range b {
		if m[id] {
			return true
		}
	}
	return false
}

// ConflictReason 定义冲突原因
type ConflictReason int

const (
	ConflictSpatial ConflictReason = iota
	ConflictSemantic
	ConflictDependency
)

// FootprintOverlap 定义冲突的具体证据
type FootprintOverlap struct {
	Ranges  []TextRange  `json:"ranges,omitempty"`
	Symbols []SymbolRef  `json:"symbols,omitempty"`
	Effects []EffectKind `json:"effects,omitempty"`
}

// Conflict 定义具体的冲突
type Conflict struct {
	ID      ConflictID       `json:"id"`
	Left    OperationID      `json:"left"`
	Right   OperationID      `json:"right"`
	Reason  ConflictReason   `json:"reason"`
	Detail  string           `json:"detail"`
	Overlap FootprintOverlap `json:"overlap"`
}

type ConflictID string
