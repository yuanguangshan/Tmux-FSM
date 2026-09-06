package main

// M2.2 单实例锁测试：Flock 互斥语义（第二个持有者必须失败，释放后可重取）。

import (
	"path/filepath"
	"testing"
)

func TestInstanceLockExclusive(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "instance.lock")

	first, err := acquireInstanceLock(lockPath)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	if _, err := acquireInstanceLock(lockPath); err == nil {
		t.Fatal("second acquire 应失败（锁被占用）")
	}

	// 释放后可重新获取
	first.Close()
	second, err := acquireInstanceLock(lockPath)
	if err != nil {
		t.Fatalf("re-acquire after release: %v", err)
	}
	second.Close()
}
