package main

import (
	"os"
	"testing"
)

// TestExecutionModeConstants 测试执行模式常量
func TestExecutionModeConstants(t *testing.T) {
	if ModeLegacy != 0 {
		t.Errorf("Expected ModeLegacy to be 0, got %d", ModeLegacy)
	}

	if ModeShadow != 1 {
		t.Errorf("Expected ModeShadow to be 1, got %d", ModeShadow)
	}

	if ModeWeaver != 2 {
		t.Errorf("Expected ModeWeaver to be 2, got %d", ModeWeaver)
	}
}

// TestLoadConfigDefault 测试默认配置加载
func TestLoadConfigDefault(t *testing.T) {
	// 确保环境变量未设置
	os.Unsetenv("TMUX_FSM_MODE")
	os.Unsetenv("TMUX_FSM_LOG_FACTS")
	os.Unsetenv("TMUX_FSM_FAIL_FAST")

	// 重置全局配置为默认值
	globalConfig = Config{
		Mode:     ModeLegacy,
		LogFacts: false,
		FailFast: false,
	}

	// 加载配置
	LoadConfig()

	// 验证默认值
	if GetMode() != ModeLegacy {
		t.Errorf("Expected default mode to be ModeLegacy, got %d", GetMode())
	}

	if ShouldLogFacts() {
		t.Errorf("Expected LogFacts to be false by default")
	}

	if ShouldFailFast() {
		t.Errorf("Expected FailFast to be false by default")
	}
}

// TestLoadConfigWithEnvVars 测试从环境变量加载配置
func TestLoadConfigWithEnvVars(t *testing.T) {
	// 设置环境变量
	os.Setenv("TMUX_FSM_MODE", "weaver")
	os.Setenv("TMUX_FSM_LOG_FACTS", "1")
	os.Setenv("TMUX_FSM_FAIL_FAST", "1")

	// 重置全局配置
	globalConfig = Config{
		Mode:     ModeLegacy,
		LogFacts: false,
		FailFast: false,
	}

	// 加载配置
	LoadConfig()

	// 验证配置值
	if GetMode() != ModeWeaver {
		t.Errorf("Expected mode to be ModeWeaver when TMUX_FSM_MODE=weaver, got %d", GetMode())
	}

	if !ShouldLogFacts() {
		t.Errorf("Expected LogFacts to be true when TMUX_FSM_LOG_FACTS=1")
	}

	if !ShouldFailFast() {
		t.Errorf("Expected FailFast to be true when TMUX_FSM_FAIL_FAST=1")
	}

	// 清理环境变量
	os.Unsetenv("TMUX_FSM_MODE")
	os.Unsetenv("TMUX_FSM_LOG_FACTS")
	os.Unsetenv("TMUX_FSM_FAIL_FAST")
}

// TestLoadConfigWithShadowMode 测试影子模式配置
func TestLoadConfigWithShadowMode(t *testing.T) {
	// 设置环境变量为shadow模式
	os.Setenv("TMUX_FSM_MODE", "shadow")

	// 重置全局配置
	globalConfig = Config{
		Mode:     ModeLegacy,
		LogFacts: false,
		FailFast: false,
	}

	// 加载配置
	LoadConfig()

	// 验证配置值
	if GetMode() != ModeShadow {
		t.Errorf("Expected mode to be ModeShadow when TMUX_FSM_MODE=shadow, got %d", GetMode())
	}

	// 清理环境变量
	os.Unsetenv("TMUX_FSM_MODE")
}

// TestLoadConfigWithInvalidMode 测试无效模式配置
func TestLoadConfigWithInvalidMode(t *testing.T) {
	// 设置无效的环境变量
	os.Setenv("TMUX_FSM_MODE", "invalid")

	// 重置全局配置
	globalConfig = Config{
		Mode:     ModeLegacy,
		LogFacts: false,
		FailFast: false,
	}

	// 加载配置
	LoadConfig()

	// 验证默认值（无效模式应使用默认值）
	if GetMode() != ModeLegacy {
		t.Errorf("Expected mode to be ModeLegacy when TMUX_FSM_MODE=invalid, got %d", GetMode())
	}

	// 清理环境变量
	os.Unsetenv("TMUX_FSM_MODE")
}

// TestConfigGetters 测试配置获取器
func TestConfigGetters(t *testing.T) {
	// 测试默认配置
	if GetMode() != ModeLegacy {
		t.Errorf("Expected GetMode() to return ModeLegacy by default, got %d", GetMode())
	}

	if ShouldLogFacts() {
		t.Errorf("Expected ShouldLogFacts() to return false by default")
	}

	if ShouldFailFast() {
		t.Errorf("Expected ShouldFailFast() to return false by default")
	}

	// 修改全局配置进行测试
	globalConfig.Mode = ModeWeaver
	globalConfig.LogFacts = true
	globalConfig.FailFast = true

	if GetMode() != ModeWeaver {
		t.Errorf("Expected GetMode() to return ModeWeaver, got %d", GetMode())
	}

	if !ShouldLogFacts() {
		t.Errorf("Expected ShouldLogFacts() to return true")
	}

	if !ShouldFailFast() {
		t.Errorf("Expected ShouldFailFast() to return true")
	}

	// 恢复默认值
	globalConfig.Mode = ModeLegacy
	globalConfig.LogFacts = false
	globalConfig.FailFast = false
}
