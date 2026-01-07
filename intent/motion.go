package intent

type MotionKind int

const (
	MotionChar MotionKind = iota
	MotionWord
	MotionLine
	MotionGoto
	MotionRange // ✅ 新增
	MotionFind
)

type FindDirection int

const (
	FindForward FindDirection = iota
	FindBackward
)

type FindMotion struct {
	Char      rune          // 要查找的字符
	Direction FindDirection // Forward / Backward
	Till      bool          // t / T
}

type Motion struct {
	Kind  MotionKind
	Count int
	Find  *FindMotion      // 只有 Kind == MotionFind 时非空
	Range *RangeMotion     // 只有 Kind == MotionRange 时非空
}