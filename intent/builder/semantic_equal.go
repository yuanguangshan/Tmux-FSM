package builder

import "tmux-fsm/intent"

type SemanticCompareMode int

const (
	CompareMigration SemanticCompareMode = iota
	CompareStrict
)

// SemanticEqual compares two intents for semantic equality.
// Nil intents are only semantically equal if both are nil.
func SemanticEqual(a, b *intent.Intent, mode SemanticCompareMode) bool {
	if a == nil || b == nil {
		return a == b
	}

	if a.Kind != b.Kind ||
		a.Target.Kind != b.Target.Kind ||
		a.Target.Direction != b.Target.Direction ||
		a.Target.Scope != b.Target.Scope ||
		a.Target.Value != b.Target.Value ||
		a.Count != b.Count {
		return false
	}

	if mode == CompareStrict && a.PaneID != b.PaneID {
		return false
	}

	// Migration mode intentionally ignores routing
	return true
}