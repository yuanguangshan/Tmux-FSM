package main

import (
	"encoding/json"
	"fmt"
	"os"
	"tmux-fsm/verifier"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: verifier verify --facts <facts.jsonl> --expect-root <root_hash>")
		os.Exit(1)
	}

	command := os.Args[1]
	if command == "verify" {
		verifyCommand(os.Args[2:])
	} else {
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func verifyCommand(args []string) {
	var factsFile, expectedRoot string
	
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--facts" && i+1 < len(args) {
			factsFile = args[i+1]
			i++
		} else if arg == "--expect-root" && i+1 < len(args) {
			expectedRoot = args[i+1]
			i++
		}
	}

	if factsFile == "" || expectedRoot == "" {
		fmt.Println("Error: --facts and --expect-root are required")
		os.Exit(1)
	}

	// 读取事实文件
	factsData, err := os.ReadFile(factsFile)
	if err != nil {
		fmt.Printf("Error reading facts file: %v\n", err)
		os.Exit(1)
	}

	// 创建验证器
	verifierInst := verifier.NewVerifier(nil) // 简化版，不使用策略

	// 执行验证
	result, err := verifierInst.VerifyFromJSON(factsData, verifier.Hash(expectedRoot))
	if err != nil {
		fmt.Printf("Verification error: %v\n", err)
		os.Exit(1)
	}

	if result.OK {
		fmt.Println("✔ VERIFIED")
		fmt.Printf("StateRoot: %s\n", result.StateRoot)
		fmt.Printf("FactsUsed: %d\n", result.FactsUsed)
		fmt.Printf("Policies: %d\n", result.Policies)
	} else {
		fmt.Println("✘ VERIFICATION FAILED")
		fmt.Printf("Reason: %s\n", result.Error)
		os.Exit(1)
	}
}