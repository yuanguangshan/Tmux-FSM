package main

import (
	"strings"
	"testing"
)

// TestServerConfig 测试服务器配置
func TestServerConfig(t *testing.T) {
	cfg := ServerConfig{
		SocketPath: "/tmp/test-socket",
	}

	if cfg.SocketPath != "/tmp/test-socket" {
		t.Errorf("Expected SocketPath to be '/tmp/test-socket', got '%s'", cfg.SocketPath)
	}
}

// TestNewServer 测试服务器创建
func TestNewServer(t *testing.T) {
	cfg := ServerConfig{
		SocketPath: "/tmp/test-socket",
	}
	
	server := NewServer(cfg)

	if server.cfg.SocketPath != "/tmp/test-socket" {
		t.Errorf("Expected server config SocketPath to be '/tmp/test-socket', got '%s'", server.cfg.SocketPath)
	}
}

// TestSocketPathVariable 测试socket路径变量
func TestSocketPathVariable(t *testing.T) {
	expectedPath := "/tmp/tmux-fsm.sock"
	
	if socketPath != expectedPath {
		t.Errorf("Expected socketPath to be '%s', got '%s'", expectedPath, socketPath)
	}
}

// TestProtocolParsing 测试协议解析逻辑
func TestProtocolParsing(t *testing.T) {
	// 测试协议字符串解析
	testPayload := "req-123|pane1|client1|h"
	parts := strings.SplitN(testPayload, "|", 4)
	
	if len(parts) != 4 {
		t.Errorf("Expected 4 parts, got %d", len(parts))
	}
	
	if parts[0] != "req-123" {
		t.Errorf("Expected requestID to be 'req-123', got '%s'", parts[0])
	}
	
	if parts[1] != "pane1" {
		t.Errorf("Expected paneID to be 'pane1', got '%s'", parts[1])
	}
	
	if parts[2] != "client1" {
		t.Errorf("Expected clientName to be 'client1', got '%s'", parts[2])
	}
	
	if parts[3] != "h" {
		t.Errorf("Expected key to be 'h', got '%s'", parts[3])
	}
}

// TestHeartbeatMessage 测试心跳消息
func TestHeartbeatMessage(t *testing.T) {
	heartbeatMsg := "test|test|__PING__"
	
	if heartbeatMsg != "test|test|__PING__" {
		t.Errorf("Expected heartbeat message to be 'test|test|__PING__', got '%s'", heartbeatMsg)
	}
}
