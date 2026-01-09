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
				OK:      false,
				Safety:  SafetyUnsafe,
				Diffs:   diffs,
				Message: "unexpected line modified",
			}
		}
	}

	return VerificationResult{
		OK:     true,
		Safety: SafetyExact,
		Diffs:  diffs,
	}
}
