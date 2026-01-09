package wal

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	"tmux-fsm/semantic"
)

//
// ─────────────────────────────────────────────────────────────
//  Semantic Event
// ─────────────────────────────────────────────────────────────
//

type SemanticEvent struct {
	ID            string            `json:"id"`
	CausalParents []string          `json:"causal_parents"`
	LocalParent   string            `json:"local_parent"`
	Time          time.Time         `json:"time"`
	Actor         string            `json:"actor"`
	Fact          semantic.BaseFact `json:"fact"`
}

//
// ─────────────────────────────────────────────────────────────
//  WAL Record (Versioned)
// ─────────────────────────────────────────────────────────────
//

type walRecord struct {
	Version  int            `json:"v"`
	Checksum string         `json:"checksum"`
	Event    SemanticEvent `json:"event"`
}

const walVersion = 1

//
// ─────────────────────────────────────────────────────────────
//  WAL
// ─────────────────────────────────────────────────────────────
//

type WAL struct {
	file   *os.File
	writer *bufio.Writer
}

func NewWAL(filename string) (*WAL, error) {
	f, err := os.OpenFile(
		filename,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}

	return &WAL{
		file:   f,
		writer: bufio.NewWriterSize(f, 64*1024),
	}, nil
}

//
// ─────────────────────────────────────────────────────────────
//  Append (Crash-Safe)
// ─────────────────────────────────────────────────────────────
//

func (w *WAL) Append(event SemanticEvent) error {

	evBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	sum := sha256.Sum256(evBytes)

	rec := walRecord{
		Version:  walVersion,
		Checksum: hex.EncodeToString(sum[:]),
		Event:    event,
	}

	line, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	// 1️⃣ write line + newline
	if _, err := w.writer.Write(append(line, '\n')); err != nil {
		return err
	}

	// 2️⃣ flush userspace buffer
	if err := w.writer.Flush(); err != nil {
		return err
	}

	// 3️⃣ fsync —— 这是 WAL 的灵魂
	return w.file.Sync()
}

//
// ─────────────────────────────────────────────────────────────
//  Close
// ─────────────────────────────────────────────────────────────
//

func (w *WAL) Close() error {
	if w.writer != nil {
		_ = w.writer.Flush()
	}
	return w.file.Close()
}

//
// ─────────────────────────────────────────────────────────────
//  Load (Resilient)
// ─────────────────────────────────────────────────────────────
//

func LoadFromWAL(filename string) ([]SemanticEvent, error) {

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := bufio.NewReaderSize(f, 64*1024)

	var events []SemanticEvent
	lineNo := 0

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return events, err
		}
		lineNo++

		var rec walRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			// ✅ 遇到坏行，直接停止（最后一条可能未写完）
			break
		}

		if rec.Version != walVersion {
			return events, errors.New("unsupported wal version")
		}

		evBytes, _ := json.Marshal(rec.Event)
		sum := sha256.Sum256(evBytes)

		if hex.EncodeToString(sum[:]) != rec.Checksum {
			// ✅ 校验失败，停止回放
			break
		}

		events = append(events, rec.Event)
	}

	return events, nil
}
