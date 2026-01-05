package core

type AnchorKind int

const (
	AnchorUnknown AnchorKind = iota

	// Cursor-relative
	AnchorAtCursor

	// Semantic
	AnchorWord
	AnchorLine
	AnchorParagraph

	// Structural
	AnchorSelection

	// Legacy Support
	AnchorLegacyRange
)
