package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
)

// InMemoryEvidenceLibrary 实现 EvidenceLibrary 接口的内存版本
type InMemoryEvidenceLibrary struct {
	mu    sync.RWMutex
	store map[string]*AuditRecord
}

func NewInMemoryEvidenceLibrary() *InMemoryEvidenceLibrary {
	return &InMemoryEvidenceLibrary{
		store: make(map[string]*AuditRecord),
	}
}

func (l *InMemoryEvidenceLibrary) Commit(record *AuditRecord) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 计算 Hash 作为引用 (Ref)
	b, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	hash := hex.EncodeToString(sum[:])

	l.store[hash] = record
	return hash, nil
}

func (l *InMemoryEvidenceLibrary) Retrieve(hash string) (*AuditRecord, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	record, ok := l.store[hash]
	if !ok {
		return nil, fmt.Errorf("evidence not found: %s", hash)
	}
	return record, nil
}

func (l *InMemoryEvidenceLibrary) Traverse(fn func(meta EvidenceMeta) error) error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for h, r := range l.store {
		meta := EvidenceMeta{
			Hash:      h,
			Timestamp: r.TimestampUTC,
		}
		if err := fn(meta); err != nil {
			return err
		}
	}
	return nil
}
