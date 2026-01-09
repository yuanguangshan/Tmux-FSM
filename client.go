package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func isServerRunning() bool {
	conn, err := net.DialTimeout("unix", socketPath, 500*time.Millisecond)
	if err != nil {
		log.Printf("Network connection failed: %v", err)
		return false
	}
	defer conn.Close()

	// 发送心跳请求确认服务器响应
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	_, err = conn.Write([]byte("test|test|__PING__"))
	if err != nil {
		log.Printf("Failed to send heartbeat: %v", err)
		return false
	}

	// 读取响应
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, err = conn.Read(buf)
	if err != nil {
		log.Printf("Failed to read heartbeat response: %v", err)
		return false
	}

	return err == nil
}

func runClient(key, paneAndClient string) {
	// Generate a RequestID for this client request
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Get the actor ID from the paneAndClient string
	actorID := paneAndClient
	if actorID == "" || actorID == "|" {
		actorID = "default|default"
	}

	// 添加参数验证和修复
	if paneAndClient == "" || paneAndClient == "|" {
		// 尝试获取当前pane和client
		// 注意：这里不能直接调用 tmux 命令，因为这可能导致循环依赖
		// 我们需要确保参数格式正确
		paneAndClient = "default|default"
	} else {
		// 检查参数格式是否正确 (pane|client)，如果 client 部分为空，尝试修复
		parts := strings.Split(paneAndClient, "|")
		if len(parts) == 2 && parts[1] == "" {
			// client 部分为空，使用默认值
			paneAndClient = parts[0] + "|default"
		} else if len(parts) == 1 {
			// 只有 pane 部分，添加默认 client
			paneAndClient = parts[0] + "|default"
		}
	}

	log.Printf("Client sending request: RequestID=%s, ActorID=%s, Key=%s", requestID, actorID, key)

	// Retry mechanism with logging
	maxRetries := 3
	var conn net.Conn
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = net.DialTimeout("unix", socketPath, 1*time.Second)
		if err == nil {
			break // Success, exit retry loop
		}

		log.Printf("Attempt %d: Failed to connect to daemon: %v", i+1, err)
		time.Sleep(500 * time.Millisecond) // Wait before retry
	}

	if err != nil {
		log.Printf("Error: daemon not running after %d attempts. Start it with 'tmux-fsm -server'", maxRetries)
		fmt.Fprintf(os.Stderr, "Error: daemon not running. Start it with 'tmux-fsm -server'\n")
		return
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		log.Printf("Error setting deadline: %v", err)
		fmt.Fprintf(os.Stderr, "Error setting deadline: %v\n", err)
		return
	}

	// Keep the original payload format for compatibility with server
	payload := fmt.Sprintf("%s|%s", paneAndClient, key)
	if _, err := conn.Write([]byte(payload)); err != nil {
		log.Printf("Failed to send payload '%s': %v", payload, err)
		return
	}

	// Read response (synchronize)
	buf, err := io.ReadAll(conn)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		return
	}
	resp := strings.TrimSpace(string(buf))
	if resp != "ok" && resp != "" {
		fmt.Println(resp)
	}

	log.Printf("Client request completed: RequestID=%s, ActorID=%s", requestID, actorID)
}
