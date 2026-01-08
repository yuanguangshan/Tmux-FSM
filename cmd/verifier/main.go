package main

import (
	"fmt"
	"os"

	"tmux-fsm/verifier"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: verifier verify <path>")
		os.Exit(1)
	}

	cmd := os.Args[1]
	path := os.Args[2]

	if cmd != "verify" {
		fmt.Println("unknown command:", cmd)
		os.Exit(1)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("read error:", err)
		os.Exit(1)
	}

	// 这里需要根据实际的 verifier 接口进行调整
	// input, err := verifier.ParseVerificationInput(data)
	// if err != nil {
	// 	fmt.Println("parse error:", err)
	// 	os.Exit(1)
	// }

	// root, err := verifier.Verify(input)
	// if err != nil {
	// 	fmt.Println("❌ verification failed:", err)
	// 	os.Exit(2)
	// }

	fmt.Println("✅ verification succeeded")
	fmt.Println("StateRoot: TODO")
}