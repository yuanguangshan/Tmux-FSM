package core

// canFuse checks if two intents can be fused together
// Phase 13.0: Conservative fusion rules
func canFuse(a, b Intent) bool {
	// Check if kinds match
	if a.Kind != b.Kind {
		return false
	}
	
	// Only allow fusing for insert operations at the same position
	if a.Kind == FactInsert {
		// Check if both intents target the same position in the same line
		if len(a.Anchors) == 1 && len(b.Anchors) == 1 {
			anchorA := a.Anchors[0]
			anchorB := b.Anchors[0]
			
			// Same line and same position
			return anchorA.LineID == anchorB.LineID && 
				   anchorA.Start == anchorB.Start && 
				   anchorA.End == anchorB.End &&
				   anchorA.PaneID == anchorB.PaneID
		}
	}
	
	return false
}

// fuse combines two compatible intents into one
// Phase 13.0: Simple concatenation for insert operations
func fuse(a, b Intent) Intent {
	if a.Kind == FactInsert && b.Kind == FactInsert {
		// For insert operations, concatenate the text
		result := a
		result.Payload.Text += b.Payload.Text
		return result
	}
	
	// For other operations, just return the first one (shouldn't happen if canFuse worked correctly)
	return a
}

// FuseIntents combines compatible intents in a sequence
// Phase 13.0: Sequential intent fusion
func FuseIntents(intents []Intent) []Intent {
	if len(intents) <= 1 {
		return intents
	}

	var out []Intent
	out = append(out, intents[0])

	for i := 1; i < len(intents); i++ {
		lastIdx := len(out) - 1
		if canFuse(out[lastIdx], intents[i]) {
			out[lastIdx] = fuse(out[lastIdx], intents[i])
		} else {
			out = append(out, intents[i])
		}
	}
	return out
}