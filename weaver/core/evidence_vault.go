package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// FileAppenderEvidenceLibrary 实现物理不可变的证据室 (RFC-WC-003 Option 1)
type FileAppenderEvidenceLibrary struct {
	mu    sync.RWMutex
	file  *os.File
	path  string
	index map[string]EvidenceMeta // 内存索引，用于快速检索
}

// NewFileAppenderEvidenceLibrary 创建并初始化一个物理证据室
func NewFileAppenderEvidenceLibrary(path string) (*FileAppenderEvidenceLibrary, error) {
	// os.O_APPEND 保证了“物理加注”不可撤回
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open evidence vault: %w", err)
	}

	vault := &FileAppenderEvidenceLibrary{
		file:  f,
		path:  path,
		index: make(map[string]EvidenceMeta),
	}

	// 启动时自动扫描物理文件，重建内存索引
	if err := vault.rebuildIndex(); err != nil {
		return nil, fmt.Errorf("failed to rebuild evidence index: %w", err)
	}

	return vault, nil
}

// Commit 提交案卷笔录。遵循“落盘即裁决”原则。
func (l *FileAppenderEvidenceLibrary) Commit(record *AuditRecord) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(record)
	if err != nil {
		return "", err
	}

	// 计算哈希
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])

	// 获取物理加注起点 (Offset)
	offset, _ := l.file.Seek(0, io.SeekEnd)

	// 物理写入 (JSON Lines 格式)
	line := append(data, '\n')
	if _, err := l.file.Write(line); err != nil {
		return "", fmt.Errorf("failed to write evidence to disk: %w", err)
	}

	// ✅ Atomic Sync: 裁决前证据必须落地物理扇区
	if err := l.file.Sync(); err != nil {
		return "", fmt.Errorf("failed to sync evidence vault: %w", err)
	}

	// 更新内存索引
	meta := EvidenceMeta{
		Hash:      hash,
		Offset:    offset,
		Timestamp: record.TimestampUTC,
		Size:      int64(len(line)),
	}
	l.index[hash] = meta

	return hash, nil
}

// Retrieve 根据案号检索原始案卷
func (l *FileAppenderEvidenceLibrary) Retrieve(hash string) (*AuditRecord, error) {
	l.mu.RLock()
	meta, ok := l.index[hash]
	l.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("evidence not found in vault: %s", hash)
	}

	// 物理跳转读取
	data := make([]byte, meta.Size)
	f, err := os.Open(l.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := f.ReadAt(data, meta.Offset); err != nil {
		return nil, err
	}

	var record AuditRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

// Traverse 巡回复核能力
func (l *FileAppenderEvidenceLibrary) Traverse(fn func(meta EvidenceMeta) error) error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// 建议实际使用时支持有序遍历，目前简单遍历索引
	for _, meta := range l.index {
		if err := fn(meta); err != nil {
			return err
		}
	}
	return nil
}

// rebuildIndex 扫描物理文件，重建司法索引
func (l *FileAppenderEvidenceLibrary) rebuildIndex() error {
	f, err := os.Open(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	// 使用 Scanner 逐行读取，因为我们使用的是 JSON Lines 格式
	// 这比 json.Decoder + Seek 更可靠
	var offset int64
	info, err := f.Stat()
	if err != nil {
		return err
	}
	fileSize := info.Size()

	// 我们需要手动读取以确保护准 offset
	data, err := os.ReadFile(l.path)
	if err != nil {
		return err
	}

	for offset < fileSize {
		// 寻找换行符
		end := offset
		for end < fileSize && data[end] != '\n' {
			end++
		}

		line := data[offset:end]
		if len(line) > 0 {
			var record AuditRecord
			if err := json.Unmarshal(line, &record); err == nil {
				// 计算哈希 (不包含换行符)
				sum := sha256.Sum256(line)
				hash := hex.EncodeToString(sum[:])

				l.index[hash] = EvidenceMeta{
					Hash:      hash,
					Offset:    offset,
					Timestamp: record.TimestampUTC,
					Size:      int64(len(line) + 1), // 包括可能存在的换行符
				}
			}
		}

		offset = end + 1 // 跳过换行符
	}

	return nil
}
