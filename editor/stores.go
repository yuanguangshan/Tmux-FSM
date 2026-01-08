package editor

import "sync"

// SimpleBufferStore 简单的 Buffer 存储实现
type SimpleBufferStore struct {
	mu      sync.RWMutex
	buffers map[BufferID]Buffer
}

// NewSimpleBufferStore 创建新的 Buffer 存储
func NewSimpleBufferStore() *SimpleBufferStore {
	return &SimpleBufferStore{
		buffers: make(map[BufferID]Buffer),
	}
}

// Get 获取指定 ID 的 Buffer
func (s *SimpleBufferStore) Get(id BufferID) Buffer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buffers[id]
}

// Set 设置 Buffer
func (s *SimpleBufferStore) Set(id BufferID, buf Buffer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buffers[id] = buf
}

// SimpleWindowStore 简单的 Window 存储实现
type SimpleWindowStore struct {
	mu      sync.RWMutex
	windows map[WindowID]*Window
}

// NewSimpleWindowStore 创建新的 Window 存储
func NewSimpleWindowStore() *SimpleWindowStore {
	return &SimpleWindowStore{
		windows: make(map[WindowID]*Window),
	}
}

// Get 获取指定 ID 的 Window
func (s *SimpleWindowStore) Get(id WindowID) *Window {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.windows[id]
}

// Set 设置 Window
func (s *SimpleWindowStore) Set(id WindowID, win *Window) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows[id] = win
}

// SimpleSelectionStore 简单的 Selection 存储实现
type SimpleSelectionStore struct {
	mu         sync.RWMutex
	selections map[BufferID][]Selection
}

// NewSimpleSelectionStore 创建新的 Selection 存储
func NewSimpleSelectionStore() *SimpleSelectionStore {
	return &SimpleSelectionStore{
		selections: make(map[BufferID][]Selection),
	}
}

// Get 获取指定 Buffer 的选区列表
func (s *SimpleSelectionStore) Get(buffer BufferID) []Selection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sels, exists := s.selections[buffer]
	if !exists {
		return []Selection{}
	}

	// 返回副本以避免并发修改
	result := make([]Selection, len(sels))
	copy(result, sels)
	return result
}

// Set 设置指定 Buffer 的选区列表
func (s *SimpleSelectionStore) Set(buffer BufferID, selections []Selection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 存储副本
	sels := make([]Selection, len(selections))
	copy(sels, selections)
	s.selections[buffer] = sels
}
