package wal

import (
	"bufio"
	"encoding/json"
	"os"
	"time"
	"tmux-fsm/semantic"
)

// SemanticEvent 语义事件
type SemanticEvent struct {
	ID       string             `json:"id"`
	ParentID string             `json:"parent_id"`
	Time     time.Time          `json:"time"`
	Actor    string             `json:"actor"`
	Fact     semantic.BaseFact  `json:"fact"`
}

// WAL Write-ahead Log
type WAL struct {
	file *os.File
	writer *bufio.Writer
}

// NewWAL 创建新的 WAL
func NewWAL(filename string) (*WAL, error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	w := &WAL{
		file: f,
		writer: bufio.NewWriter(f),
	}

	return w, nil
}

// Append 向 WAL 追加事件
func (w *WAL) Append(event SemanticEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// 添加换行符分隔
	data = append(data, '\n')

	_, err = w.writer.Write(data)
	if err != nil {
		return err
	}

	// 确保写入磁盘
	return w.writer.Flush()
}

// Close 关闭 WAL
func (w *WAL) Close() error {
	if w.writer != nil {
		w.writer.Flush()
	}
	return w.file.Close()
}

// LoadFromWAL 从 WAL 加载事件
func LoadFromWAL(filename string) ([]SemanticEvent, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []SemanticEvent
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		var event SemanticEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return events, nil
}