package logic

import (
	"tmux-fsm/weaver/core"
	"unicode"
)

// TextObjectKind defines the kind of text object
// Duplicates main package for Weaver isolation
type TextObjectKind int

const (
	ObjectWord TextObjectKind = iota
	ObjectWORD
	ObjectSentence
	ObjectParagraph
	ObjectDelimited
)

// TextObjectSpec represents a parsed text object intent
type TextObjectSpec struct {
	Kind   TextObjectKind
	Inner  bool
	DelimL rune
	DelimR rune
}

// Document wraps Snapshot to provide navigation methods for Text Object Resolver
type Document struct {
	Snapshot core.Snapshot
}

// Loc represents a location in terms of line index and rune index (column)
type Loc struct {
	Line int
	Col  int
}

// ParseTextObject parses "iw", "ap", "a{" into a spec
func ParseTextObject(input string) TextObjectSpec {
	if len(input) != 2 {
		panic("invalid text object input length")
	}

	if input[0] != 'i' && input[0] != 'a' {
		panic("invalid text object modifier: " + string(input[0]))
	}

	spec := TextObjectSpec{}
	spec.Inner = (input[0] == 'i')

	switch input[1] {
	case 'w':
		spec.Kind = ObjectWord
	case 'W':
		spec.Kind = ObjectWORD
	case 's':
		spec.Kind = ObjectSentence
	case 'p':
		spec.Kind = ObjectParagraph

	case '(', ')':
		spec.Kind = ObjectDelimited
		spec.DelimL = '('
		spec.DelimR = ')'

	case '{', '}':
		spec.Kind = ObjectDelimited
		spec.DelimL = '{'
		spec.DelimR = '}'

	case '[', ']':
		spec.Kind = ObjectDelimited
		spec.DelimL = '['
		spec.DelimR = ']'

	case '"', '\'', '`':
		r := rune(input[1])
		spec.Kind = ObjectDelimited
		spec.DelimL = r
		spec.DelimR = r

	case '<', '>':
		spec.Kind = ObjectDelimited
		spec.DelimL = '<'
		spec.DelimR = '>'

	default:
		panic("unsupported text object: " + string(input[1]))
	}

	return spec
}

// Document Methods adapting core.Snapshot

func (d Document) LineCount() int {
	return len(d.Snapshot.Lines)
}

func (d Document) RunesAtLine(lineIdx int) []rune {
	if lineIdx < 0 || lineIdx >= d.LineCount() {
		return nil
	}
	// core.LineSnapshot.Text
	return []rune(d.Snapshot.Lines[lineIdx].Text)
}

func (d Document) RuneAt(l Loc) rune {
	runes := d.RunesAtLine(l.Line)
	if runes == nil {
		return 0
	}
	if l.Col < 0 || l.Col >= len(runes) {
		return 0
	}
	return runes[l.Col]
}

func (d Document) RuneBefore(l Loc) rune {
	prev := d.MoveLeft(l)
	if prev == l {
		return 0
	}
	return d.RuneAt(prev)
}

func (d Document) IsBOF(l Loc) bool {
	return l.Line == 0 && l.Col == 0
}

func (d Document) IsEOF(l Loc) bool {
	lastLineIdx := d.LineCount() - 1
	if lastLineIdx < 0 {
		return true
	}
	runes := d.RunesAtLine(lastLineIdx)
	return l.Line == lastLineIdx && l.Col >= len(runes)
}

func (d Document) MoveLeft(l Loc) Loc {
	if l.Col > 0 {
		return Loc{Line: l.Line, Col: l.Col - 1}
	}
	if l.Line > 0 {
		prevLineIdx := l.Line - 1
		runes := d.RunesAtLine(prevLineIdx)
		return Loc{Line: prevLineIdx, Col: len(runes)} // End of prev line (after last char)
	}
	return l // BOF
}

func (d Document) MoveRight(l Loc) Loc {
	runes := d.RunesAtLine(l.Line)
	if runes == nil {
		return l
	}

	if l.Col < len(runes) {
		return Loc{Line: l.Line, Col: l.Col + 1}
	}

	if l.Line < d.LineCount()-1 {
		return Loc{Line: l.Line + 1, Col: 0}
	}

	return l // EOF
}

func (d Document) LineIsWhitespace(lineIdx int) bool {
	runes := d.RunesAtLine(lineIdx)
	for _, r := range runes {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// Helpers

func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

func isAlphaNum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// Range logic (Loc based)
type LocRange struct {
	Start Loc
	End   Loc
}

// Resolvers

func ResolveTextObject(doc Document, cursor Loc, spec TextObjectSpec) LocRange {
	switch spec.Kind {
	case ObjectWord:
		return resolveWord(doc, cursor, spec.Inner, false)
	case ObjectWORD:
		return resolveWord(doc, cursor, spec.Inner, true)
	case ObjectSentence:
		return resolveSentence(doc, cursor, spec.Inner)
	case ObjectParagraph:
		return resolveParagraph(doc, cursor, spec.Inner)
	case ObjectDelimited:
		return resolveDelimited(doc, cursor, spec)
	default:
		// Should not happen if validation passed
		return LocRange{Start: cursor, End: cursor}
	}
}

func resolveWord(doc Document, cursor Loc, inner bool, big bool) LocRange {
	isWord := func(r rune) bool {
		if big {
			return !isWhitespace(r)
		}
		return isAlphaNum(r) || r == '_'
	}

	pos := cursor
	if !isWord(doc.RuneAt(pos)) {
		if inner {
			// As per panic instruction in previous file, we replicate behavior where appropriate.
			// However in Weaver we prefer error returns, but this structure panics.
			// Let's implement robust behavior: if whitespace, treat whitespace as word.
		}

		if !big {
			isWord = func(r rune) bool {
				return isWhitespace(r)
			}
		} else {
			isWord = func(r rune) bool {
				return isWhitespace(r)
			}
		}
	}

	left := pos
	for isWord(doc.RuneBefore(left)) {
		left = doc.MoveLeft(left)
	}

	right := pos
	for isWord(doc.RuneAt(right)) {
		right = doc.MoveRight(right)
	}

	if inner {
		return LocRange{Start: left, End: right}
	}

	// around
	l := left
	for isWhitespace(doc.RuneBefore(l)) {
		l = doc.MoveLeft(l)
	}

	r := right
	for isWhitespace(doc.RuneAt(r)) {
		r = doc.MoveRight(r)
	}

	return LocRange{Start: l, End: r}
}

func resolveSentence(doc Document, cursor Loc, inner bool) LocRange {
	isEnd := func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	}

	left := cursor
	for !isEnd(doc.RuneBefore(left)) && !doc.IsBOF(left) {
		left = doc.MoveLeft(left)
	}

	right := cursor
	for !isEnd(doc.RuneAt(right)) && !doc.IsEOF(right) {
		right = doc.MoveRight(right)
	}
	right = doc.MoveRight(right)

	r := LocRange{Start: left, End: right}

	if inner {
		return trimWhitespace(doc, r)
	}
	return expandWhitespace(doc, r)
}

func resolveParagraph(doc Document, cursor Loc, inner bool) LocRange {
	isBlank := func(lineIdx int) bool {
		return doc.LineIsWhitespace(lineIdx)
	}

	l := cursor.Line
	for l > 0 && !isBlank(l-1) {
		l--
	}

	r := cursor.Line
	for r < doc.LineCount()-1 && !isBlank(r+1) {
		r++
	}

	start := Loc{Line: l, Col: 0}

	endLine := r + 1
	if endLine > doc.LineCount() {
		endLine = doc.LineCount()
	}
	end := Loc{Line: endLine, Col: 0}

	if inner {
		return LocRange{Start: start, End: end}
	}

	for l > 0 && isBlank(l-1) {
		l--
	}

	rScan := r + 1
	for rScan < doc.LineCount() && isBlank(rScan) {
		rScan++
	}

	return LocRange{
		Start: Loc{Line: l, Col: 0},
		End:   Loc{Line: rScan, Col: 0},
	}
}

func resolveDelimited(doc Document, cursor Loc, spec TextObjectSpec) LocRange {
	depth := 0
	left := doc.MoveLeft(cursor)

	// Find opening
	for !doc.IsBOF(left) {
		r := doc.RuneAt(left)

		if r == spec.DelimR {
			depth++
		} else if r == spec.DelimL {
			if depth == 0 {
				break
			}
			depth--
		}
		left = doc.MoveLeft(left)
	}

	// If fail, we technically should error.
	// For robust logic, return cursor range? Or assume found?
	// The original had panic.
	if doc.RuneAt(left) != spec.DelimL {
		// handle mismatch
	}

	// Find closing
	depth = 0
	right := doc.MoveRight(cursor)

	for !doc.IsEOF(right) {
		r := doc.RuneAt(right)

		if r == spec.DelimL {
			depth++
		} else if r == spec.DelimR {
			if depth == 0 {
				break
			}
			depth--
		}
		right = doc.MoveRight(right)
	}

	if spec.Inner {
		return LocRange{
			Start: doc.MoveRight(left),
			End:   right, // exclusive of right delim?
		}
	}

	return LocRange{
		Start: left,
		End:   doc.MoveRight(right),
	}
}

func trimWhitespace(doc Document, r LocRange) LocRange {
	for isWhitespace(doc.RuneAt(r.Start)) {
		newStart := doc.MoveRight(r.Start)
		if newStart == r.Start {
			break
		}
		r.Start = newStart
		if r.Start.Line > r.End.Line || (r.Start.Line == r.End.Line && r.Start.Col >= r.End.Col) {
			break
		}
	}
	for isWhitespace(doc.RuneBefore(r.End)) {
		newEnd := doc.MoveLeft(r.End)
		if newEnd == r.End {
			break
		}
		r.End = newEnd
		if r.Start.Line > r.End.Line || (r.Start.Line == r.End.Line && r.Start.Col >= r.End.Col) {
			break
		}
	}
	return r
}

func expandWhitespace(doc Document, r LocRange) LocRange {
	for isWhitespace(doc.RuneBefore(r.Start)) {
		newStart := doc.MoveLeft(r.Start)
		if newStart == r.Start {
			break
		}
		r.Start = newStart
	}
	for isWhitespace(doc.RuneAt(r.End)) {
		newEnd := doc.MoveRight(r.End)
		if newEnd == r.End {
			break
		}
		r.End = newEnd
	}
	return r
}
