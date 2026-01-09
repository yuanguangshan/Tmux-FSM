package intent

type TextObjectScope int

const (
	Inner TextObjectScope = iota
	Around
)

type TextObjectKind int

const (
	Word TextObjectKind = iota
	Paren
	Bracket
	Brace
	QuoteSingle
	QuoteDouble
	Backtick
)

type TextObject struct {
	Scope  TextObjectScope
	Object TextObjectKind
}
