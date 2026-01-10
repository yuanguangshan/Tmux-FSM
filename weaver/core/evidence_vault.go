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
		return err
	}
	defer f.Close()

	var offset int64
	decoder := json.NewDecoder(f)
	for {
		start := offset
		var record AuditRecord
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// 估计当前读取结束的位置，更新并计算哈希
		newOffset, _ := f.Seek(0, io.SeekCurrent)
		size := newOffset - start

		// 重新计算哈希（保证索引与物理内容严格对应）
		f.Seek(start, io.SeekStart)
		data := make([]byte, size)
		f.ReadAt(data, start)
		sum := sha256.Sum256(data[:len(data)-1]) // 去掉末尾换行符
		hash := hex.EncodeToString(sum[:])

		l.index[hash] = EvidenceMeta{
			Hash:      hash,
			Offset:    start,
			Timestamp: record.TimestampUTC,
			Size:      size,
		}
		offset = newOffset
	}
	return nil
}
