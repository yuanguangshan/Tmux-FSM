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

// Direction for character-wise and line-wise motions
type Direction int

const (
	DirectionNone Direction = iota
	DirectionLeft
	DirectionRight
	DirectionUp
	DirectionDown
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
	Kind      MotionKind
	Count     int
	Direction Direction      // For up, down, left, right
	Find      *FindMotion      // 只有 Kind == MotionFind 时非空
	Range     *RangeMotion     // 只有 Kind == MotionRange 时非空
}
