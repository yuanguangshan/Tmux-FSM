package main

import (
	"os"
	"strings"
)

// ExecutionMode 执行模式
type ExecutionMode int

const (
	ModeLegacy ExecutionMode = iota // 完全使用旧系统
	ModeShadow                      // Weaver 影子模式（记录但不执行）
	ModeWeaver                      // 完全使用 Weaver（阶段 3+）
)

// Config 全局配置
type Config struct {
	Mode     ExecutionMode
	LogFacts bool
	FailFast bool
}

// globalConfig 全局配置实例
var globalConfig = Config{
	Mode:     ModeLegacy, // 默认使用 Legacy 模式
	LogFacts: false,
	FailFast: false,
}

// LoadConfig 从环境变量加载配置
func LoadConfig() {
	// TMUX_FSM_MODE: legacy | shadow | weaver
	mode := strings.ToLower(os.Getenv("TMUX_FSM_MODE"))
	switch mode {
	case "shadow":
		globalConfig.Mode = ModeShadow
	case "weaver":
		globalConfig.Mode = ModeWeaver
	default:
		globalConfig.Mode = ModeLegacy
	}

	// TMUX_FSM_LOG_FACTS: 1 | 0
	if os.Getenv("TMUX_FSM_LOG_FACTS") == "1" {
		globalConfig.LogFacts = true
	}

	// TMUX_FSM_FAIL_FAST: 1 | 0
	if os.Getenv("TMUX_FSM_FAIL_FAST") == "1" {
		globalConfig.FailFast = true
	}
}

// GetMode 获取当前执行模式
func GetMode() ExecutionMode {
	return globalConfig.Mode
}

// ShouldLogFacts 是否记录 Facts
func ShouldLogFacts() bool {
	return globalConfig.LogFacts
}

// ShouldFailFast 是否快速失败
func ShouldFailFast() bool {
	return globalConfig.FailFast
}
