package core

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// TransactionID represents a unique identifier for a transaction
type TransactionID string

// History 历史管理器接口
// 负责维护 Undo/Redo 栈
type History interface {
	// Push 记录一个新的事务（并清空 Redo 栈）
	Push(tx *Transaction)

	// PopUndo 弹出最近一个可撤销的事务
	PopUndo() *Transaction

	// PopRedo 弹出最近一个可重做的事务
	PopRedo() *Transaction

	// AddRedo 将撤销的事务放入 Redo 栈
	AddRedo(tx *Transaction)

	// PushBack 将事务压入 Undo 栈，但不清空 Redo 栈（用于 Redo 操作）
	PushBack(tx *Transaction)

	// CanUndo 是否可撤销
	CanUndo() bool

	// CanRedo 是否可重做
	CanRedo() bool

	// GetTransactionID 获取当前事务的ID
	GetTransactionID(tx *Transaction) TransactionID

	// GetLastTransactionID 获取最后一个事务的ID
	GetLastTransactionID() TransactionID
}

// InMemoryHistory 基于内存的实现
type InMemoryHistory struct {
	undoStack      []*Transaction
	redoStack      []*Transaction
	transactionIDs map[*Transaction]TransactionID // Track transaction IDs
	lastTxID       TransactionID                 // Track the last transaction ID
	capacity       int
	mu             sync.RWMutex
}

func NewInMemoryHistory(capacity int) *InMemoryHistory {
	if capacity <= 0 {
		capacity = 50 // Default
	}
	return &InMemoryHistory{
		undoStack:      make([]*Transaction, 0, capacity),
		redoStack:      make([]*Transaction, 0, capacity),
		transactionIDs: make(map[*Transaction]TransactionID),
		capacity:       capacity,
	}
}

func (h *InMemoryHistory) Push(tx *Transaction) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Generate and store transaction ID
	txID := h.generateTransactionID(tx)
	h.transactionIDs[tx] = txID
	h.lastTxID = txID

	// 1. 如果超出容量，移除最旧的
	if len(h.undoStack) >= h.capacity {
		oldestTx := h.undoStack[0]
		delete(h.transactionIDs, oldestTx) // Clean up old transaction ID
		h.undoStack = h.undoStack[1:]
	}

	// 2. 压栈
	h.undoStack = append(h.undoStack, tx)

	// 3. 清空 Redo and its associated transaction IDs
	for _, redoTx := range h.redoStack {
		delete(h.transactionIDs, redoTx) // Clean up redo transaction IDs
	}
	h.redoStack = nil
}

func (h *InMemoryHistory) PushBack(tx *Transaction) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Generate and store transaction ID
	txID := h.generateTransactionID(tx)
	h.transactionIDs[tx] = txID
	h.lastTxID = txID

	// 1. 如果超出容量，移除最旧的
	if len(h.undoStack) >= h.capacity {
		oldestTx := h.undoStack[0]
		delete(h.transactionIDs, oldestTx) // Clean up old transaction ID
		h.undoStack = h.undoStack[1:]
	}

	// 2. 压栈
	h.undoStack = append(h.undoStack, tx)
}

func (h *InMemoryHistory) PopUndo() *Transaction {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.undoStack) == 0 {
		return nil
	}

	lastIdx := len(h.undoStack) - 1
	tx := h.undoStack[lastIdx]
	h.undoStack = h.undoStack[:lastIdx]

	// Remove the transaction ID from the map
	delete(h.transactionIDs, tx)

	// Update lastTxID to the new last transaction in undo stack, if any
	if len(h.undoStack) > 0 {
		lastTx := h.undoStack[len(h.undoStack)-1]
		h.lastTxID = h.transactionIDs[lastTx]
	} else {
		h.lastTxID = ""
	}

	return tx
}

func (h *InMemoryHistory) PopRedo() *Transaction {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.redoStack) == 0 {
		return nil
	}

	lastIdx := len(h.redoStack) - 1
	tx := h.redoStack[lastIdx]
	h.redoStack = h.redoStack[:lastIdx]

	// Remove the transaction ID from the map
	delete(h.transactionIDs, tx)

	return tx
}

func (h *InMemoryHistory) AddRedo(tx *Transaction) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Generate and store transaction ID if not already present
	if _, exists := h.transactionIDs[tx]; !exists {
		txID := h.generateTransactionID(tx)
		h.transactionIDs[tx] = txID
		h.lastTxID = txID
	}

	if len(h.redoStack) >= h.capacity {
		oldestTx := h.redoStack[0]
		delete(h.transactionIDs, oldestTx) // Clean up old transaction ID
		h.redoStack = h.redoStack[1:] // Drop oldest redo? Or drop newest? Usually drop oldest.
	}
	h.redoStack = append(h.redoStack, tx)
}

// GetTransactionID generates or retrieves the transaction ID for a given transaction
func (h *InMemoryHistory) GetTransactionID(tx *Transaction) TransactionID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if txID, exists := h.transactionIDs[tx]; exists {
		return txID
	}

	// If not found, generate a new one (shouldn't happen in normal operation)
	return h.generateTransactionID(tx)
}

// GetLastTransactionID returns the ID of the last transaction
func (h *InMemoryHistory) GetLastTransactionID() TransactionID {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastTxID
}

// generateTransactionID creates a unique ID for a transaction based on its content and timestamp
func (h *InMemoryHistory) generateTransactionID(tx *Transaction) TransactionID {
	// Create a unique ID based on transaction content and current time
	content := fmt.Sprintf("%v_%d", tx, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(content))
	return TransactionID(fmt.Sprintf("%x", hash[:16])) // Use first 16 bytes for shorter ID
}

func (h *InMemoryHistory) CanUndo() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.undoStack) > 0
}

func (h *InMemoryHistory) CanRedo() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.redoStack) > 0
}
