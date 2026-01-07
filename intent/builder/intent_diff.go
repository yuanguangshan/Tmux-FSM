package builder

import "tmux-fsm/intent"

type IntentDiff struct {
	Field  string
	Legacy interface{}
	Native interface{}
}

func DiffIntent(a, b *intent.Intent) []IntentDiff {
	var diffs []IntentDiff

	if a.Kind != b.Kind {
		diffs = append(diffs, IntentDiff{"Kind", a.Kind, b.Kind})
	}

	if a.Count != b.Count {
		diffs = append(diffs, IntentDiff{"Count", a.Count, b.Count})
	}

	if a.Target.Kind != b.Target.Kind {
		diffs = append(diffs, IntentDiff{"Target.Kind", a.Target.Kind, b.Target.Kind})
	}

	if a.Target.Direction != b.Target.Direction {
		diffs = append(diffs, IntentDiff{"Target.Direction", a.Target.Direction, b.Target.Direction})
	}

	if a.Target.Scope != b.Target.Scope {
		diffs = append(diffs, IntentDiff{"Target.Scope", a.Target.Scope, b.Target.Scope})
	}

	if a.Target.Value != b.Target.Value {
		diffs = append(diffs, IntentDiff{"Target.Value", a.Target.Value, b.Target.Value})
	}

	if a.PaneID != b.PaneID {
		diffs = append(diffs, IntentDiff{"PaneID", a.PaneID, b.PaneID})
	}

	return diffs
}