package intent

type RangeKind int

const (
	RangeTextObject RangeKind = iota
	RangeVisual
	RangeLineStart // For '0'
	RangeLineEnd   // For '$'
)

type RangeMotion struct {
	Kind       RangeKind
	TextObject *TextObject
}
