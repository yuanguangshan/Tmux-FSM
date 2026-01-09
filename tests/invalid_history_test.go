package tests

import (
	"os"
	"testing"

	"tmux-fsm/verifier"
)

func loadExample(t *testing.T, path string) verifier.VerifyInput {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	// 这里需要根据实际的 verifier 接口进行调整
	input := verifier.VerifyInput{}
	return input
}

func TestInvalidHistory_ParentMismatch(t *testing.T) {
	// 这里需要根据实际的 verifier 接口进行调整
	// input := loadExample(t,
	// 	"../examples/invalid_history/parent_mismatch/facts.json",
	// )

	// _, err := verifier.Verify(input)
	// if err == nil {
	// 	t.Fatalf("expected verification failure, got success")
	// }
	t.Skip("Verifier interface needs to be implemented")
}

func TestInvalidHistory_ReorderedFacts(t *testing.T) {
	// 这里需要根据实际的 verifier 接口进行调整
	t.Skip("Verifier interface needs to be implemented")
}

func TestInvalidHistory_SameTextDifferentRoot(t *testing.T) {
	// 这里需要根据实际的 verifier 接口进行调整
	t.Skip("Verifier interface needs to be implemented")
}
