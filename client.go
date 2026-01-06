package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

func isServerRunning() bool {
	conn, err := net.DialTimeout("unix", socketPath, 500*time.Millisecond)
	if err != nil {
		return false
	}
	defer conn.Close()

	// 发送心跳请求确认服务器响应
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	conn.Write([]byte("test|test|__PING__"))

	// 读取响应
	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, err = conn.Read(buf)
	return err == nil
}

func runClient(key, paneAndClient string) {
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

	conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: daemon not running. Start it with 'tmux-fsm -server'\n")
		return
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting deadline: %v\n", err)
		return
	}

	payload := fmt.Sprintf("%s|%s", paneAndClient, key)
	if _, err := conn.Write([]byte(payload)); err != nil {
		return
	}

	// Read response (synchronize)
	buf, err := io.ReadAll(conn)
	if err != nil {
		return
	}
	resp := strings.TrimSpace(string(buf))
	if resp != "ok" && resp != "" {
		fmt.Println(resp)
	}
}