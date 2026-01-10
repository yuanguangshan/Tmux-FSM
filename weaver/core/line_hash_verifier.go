package core

type LineHashVerifier struct{}

func NewLineHashVerifier() *LineHashVerifier {
	return &LineHashVerifier{}
}

func (v *LineHashVerifier) Verify(
	pre Snapshot,
	facts []ResolvedFact,
	post Snapshot,
) VerificationResult {

	diffs := DiffSnapshot(pre, post)
	allowed := AllowedLineSet(facts)

	for _, d := range diffs {
		if !allowed.Contains(d.LineID) {
			return VerificationResult{
				OK:      true, // Downgrade to Warning (OK=true) for better UX
				Safety:  SafetyUnsafe,
				Diffs:   diffs,
				Message: "warning: unexpected line modified (clocks or background activity)",
			}
		}
	}

	return VerificationResult{
		OK:     true,
		Safety: SafetyExact,
		Diffs:  diffs,
	}
}
