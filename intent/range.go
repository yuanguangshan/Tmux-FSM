package intent

type RangeKind int

const (
	RangeTextObject RangeKind = iota
	RangeVisual
)

type RangeMotion struct {
	Kind       RangeKind
	TextObject *TextObject
}