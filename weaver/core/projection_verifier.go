package core

type SafetyLevel int

const (
    SafetyExact SafetyLevel = iota
    SafetyFuzzy
    SafetyUnsafe
)

type VerificationResult struct {
    OK      bool
    Safety  SafetyLevel
    Diffs   []SnapshotDiff
    Message string
}

type ProjectionVerifier interface {
    Verify(
        pre Snapshot,
        facts []ResolvedFact,
        post Snapshot,
    ) VerificationResult
}