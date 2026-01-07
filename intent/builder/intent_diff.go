package builder

import "tmux-fsm/intent"

type IntentDiff struct {
	Field  string
	Legacy interface{}
	Native interface{}
}

func DiffIntent(legacy, native *intent.Intent) []IntentDiff {
	var diffs []IntentDiff

	if legacy == nil || native == nil {
		return diffs
	}

	if legacy.Kind != native.Kind {
		diffs = append(diffs, IntentDiff{"Kind", legacy.Kind, native.Kind})
	}

	if legacy.Count != native.Count {
		diffs = append(diffs, IntentDiff{"Count", legacy.Count, native.Count})
	}

	if legacy.Target.Kind != native.Target.Kind {
		diffs = append(diffs, IntentDiff{"Target.Kind", legacy.Target.Kind, native.Target.Kind})
	}

	if legacy.Target.Direction != native.Target.Direction {
		diffs = append(diffs, IntentDiff{"Target.Direction", legacy.Target.Direction, native.Target.Direction})
	}

	if legacy.Target.Scope != native.Target.Scope {
		diffs = append(diffs, IntentDiff{"Target.Scope", legacy.Target.Scope, native.Target.Scope})
	}

	if legacy.Target.Value != native.Target.Value {
		diffs = append(diffs, IntentDiff{"Target.Value", legacy.Target.Value, native.Target.Value})
	}

	if legacy.PaneID != native.PaneID {
		diffs = append(diffs, IntentDiff{"PaneID", legacy.PaneID, native.PaneID})
	}

	return diffs
}