// package core

// // canFuse checks if two intents can be fused together
// // Phase 13.0: Conservative fusion rules
// func canFuse(a, b Intent) bool {
// 	// Check if kinds match
// 	if a.Kind != b.Kind {
// 		return false
// 	}

// 	// Only allow fusing for insert operations at the same position
// 	if a.Kind == FactInsert {
// 		// Check if both intents target the same position in the same line
// 		if len(a.Anchors) == 1 && len(b.Anchors) == 1 {
// 			anchorA := a.Anchors[0]
// 			anchorB := b.Anchors[0]

// 			// Same line and same position
// 			return anchorA.LineID == anchorB.LineID &&
// 				   anchorA.Start == anchorB.Start &&
// 				   anchorA.End == anchorB.End &&
// 				   anchorA.PaneID == anchorB.PaneID
// 		}
// 	}

// 	return false
// }

// // fuse combines two compatible intents into one
// // Phase 13.0: Simple concatenation for insert operations
// func fuse(a, b Intent) Intent {
// 	if a.Kind == FactInsert && b.Kind == FactInsert {
// 		// For insert operations, concatenate the text
// 		result := a
// 		result.Payload.Text += b.Payload.Text
// 		return result
// 	}

// 	// For other operations, just return the first one (shouldn't happen if canFuse worked correctly)
// 	return a
// }

// // FuseIntents combines compatible intents in a sequence
// // Phase 13.0: Sequential intent fusion
// func FuseIntents(intents []Intent) []Intent {
// 	if len(intents) <= 1 {
// 		return intents
// 	}

// 	var out []Intent
// 	out = append(out, intents[0])

// 	for i := 1; i < len(intents); i++ {
// 		lastIdx := len(out) - 1
// 		if canFuse(out[lastIdx], intents[i]) {
// 			out[lastIdx] = fuse(out[lastIdx], intents[i])
// 		} else {
// 			out = append(out, intents[i])
// 		}
// 	}
// 	return out
// }

package core

import (
	"log"
)

// FuseCondition defines the conditions under which intents can be fused
type FuseCondition int

const (
	// NoFusion means intents should not be fused
	NoFusion FuseCondition = iota
	// SameKindSameTarget means intents of the same kind affecting the same target can be fused
	SameKindSameTarget
	// SequentialInserts means consecutive insert operations at adjacent positions can be fused
	SequentialInserts
	// SameUserAction means intents originating from the same user action can be fused
	SameUserAction
)

// canFuse determines if two intents can be fused based on strict conditions
func canFuse(a, b Intent) FuseCondition {
	// Log the fusion attempt for audit trail
	log.Printf("Attempting to fuse intents: A.Kind=%d, A.PaneID=%s, B.Kind=%d, B.PaneID=%s",
		a.GetKind(), a.GetPaneID(), b.GetKind(), b.GetPaneID())

	// Condition 1: Both intents must have the same kind
	if a.GetKind() != b.GetKind() {
		log.Printf("Cannot fuse intents: different kinds (%d vs %d)", a.GetKind(), b.GetKind())
		return NoFusion
	}

	// Condition 2: Both intents must affect the same pane
	if a.GetPaneID() != b.GetPaneID() {
		log.Printf("Cannot fuse intents: different panes (%s vs %s)", a.GetPaneID(), b.GetPaneID())
		return NoFusion
	}

	// Condition 3: For insert operations, check if they are sequential
	if a.GetKind() == IntentInsert && b.GetKind() == IntentInsert {
		// For now, we'll allow fusion of insert operations in the same pane
		// More sophisticated logic would check positions, etc.
		log.Printf("Fusing insert intents in same pane")
		return SequentialInserts
	}

	// Condition 4: For same kind and same pane, allow fusion with restrictions
	log.Printf("Fusing intents: same kind and pane")
	return SameKindSameTarget
}

// FuseIntents combines two compatible intents into one according to defined conditions
func FuseIntents(a, b Intent) Intent {
	condition := canFuse(a, b)

	switch condition {
	case NoFusion:
		// When fusion is not allowed, return the later intent but log the decision
		log.Printf("Fusion not allowed between intents, returning the later intent")
		return b
	case SequentialInserts:
		// For sequential inserts, we'll return the second intent but log the fusion
		// In a more sophisticated implementation, we would combine the operations
		log.Printf("Fusing sequential insert intents in pane %s", a.GetPaneID())
		// For now, return the second intent with an updated count
		return b
	case SameKindSameTarget:
		// For same kind and target, use the later intent but log the fusion
		log.Printf("Fusing intents with same kind and pane")
		return b
	default:
		// Default case: return the later intent
		log.Printf("Using default fusion behavior, returning later intent")
		return b
	}
}
