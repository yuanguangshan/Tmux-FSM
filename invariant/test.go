package invariant

import (
	"math/rand"
	"testing"
	"time"
)

// TextState 模拟文本状态
type TextState struct {
	Text   string
	Cursor int
}

// Apply 模拟事务对状态的应用
func (s TextState) Apply(tx Transaction) (TextState, error) {
	switch t := tx.(type) {
	case *InsertTx:
		if t.Pos < 0 || t.Pos > len(s.Text) {
			return s, nil // 边界检查，不执行
		}
		newText := s.Text[:t.Pos] + t.Text + s.Text[t.Pos:]
		return TextState{
			Text:   newText,
			Cursor: t.Pos + len(t.Text),
		}, nil

	case *DeleteTx:
		if t.Pos < 0 || t.Pos+t.Len > len(s.Text) {
			return s, nil // 边界检查，不执行
		}
		newText := s.Text[:t.Pos] + s.Text[t.Pos+t.Len:]
		return TextState{
			Text:   newText,
			Cursor: t.Pos,
		}, nil

	case *MoveCursorTx:
		newCursor := t.To
		if newCursor < 0 {
			newCursor = 0
		}
		if newCursor > len(s.Text) {
			newCursor = len(s.Text)
		}
		return TextState{
			Text:   s.Text,
			Cursor: newCursor,
		}, nil
	}

	return s, nil
}

// Transaction 接口定义
type Transaction interface {
	Apply() error
	Inverse() Transaction
	Kind() string
	Tags() []string
	CanMerge(next Transaction) bool
	Merge(next Transaction) Transaction
}

// InsertTx 插入事务
type InsertTx struct {
	Pos  int
	Text string
}

func (t *InsertTx) Apply() error { return nil }
func (t *InsertTx) Inverse() Transaction {
	return &DeleteTx{Pos: t.Pos, Len: len(t.Text)}
}
func (t *InsertTx) Kind() string                       { return "insert" }
func (t *InsertTx) Tags() []string                     { return []string{"insert"} }
func (t *InsertTx) CanMerge(next Transaction) bool     { return false }
func (t *InsertTx) Merge(next Transaction) Transaction { return next }

// DeleteTx 删除事务
type DeleteTx struct {
	Pos int
	Len int
}

func (t *DeleteTx) Apply() error { return nil }
func (t *DeleteTx) Inverse() Transaction {
	return &InsertTx{Pos: t.Pos, Text: ""} // 简化实现
}
func (t *DeleteTx) Kind() string                       { return "delete" }
func (t *DeleteTx) Tags() []string                     { return []string{"delete"} }
func (t *DeleteTx) CanMerge(next Transaction) bool     { return false }
func (t *DeleteTx) Merge(next Transaction) Transaction { return next }

// MoveCursorTx 移动光标事务
type MoveCursorTx struct {
	To int
}

func (t *MoveCursorTx) Apply() error { return nil }
func (t *MoveCursorTx) Inverse() Transaction {
	// 简化实现
	return &MoveCursorTx{To: 0}
}
func (t *MoveCursorTx) Kind() string                       { return "move" }
func (t *MoveCursorTx) Tags() []string                     { return []string{"move"} }
func (t *MoveCursorTx) CanMerge(next Transaction) bool     { return false }
func (t *MoveCursorTx) Merge(next Transaction) Transaction { return next }

// TestTxInverseProperty 测试事务与其逆操作的性质
func TestTxInverseProperty(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100; i++ {
		// 随机生成初始状态
		initialText := randomString(rand.Intn(20))
		s0 := TextState{Text: initialText, Cursor: rand.Intn(len(initialText) + 1)}

		// 创建一个随机事务
		tx := randomTransaction(len(s0.Text))

		// 应用事务
		s1, err := s0.Apply(tx)
		if err != nil {
			continue // Apply 失败不违反不变量
		}

		// 应用逆事务
		s2, err := s1.Apply(tx.Inverse())
		if err != nil {
			t.Errorf("Inverse application failed: %v", err)
			continue
		}

		// 检查是否回到原始状态
		if s0.Text != s2.Text {
			t.Errorf("Apply ∘ Inverse ≠ Identity: %s != %s", s0.Text, s2.Text)
		}
	}
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// randomTransaction 生成随机事务
func randomTransaction(maxPos int) Transaction {
	pos := rand.Intn(maxPos + 1)
	switch rand.Intn(3) {
	case 0:
		return &InsertTx{Pos: pos, Text: randomString(rand.Intn(5))}
	case 1:
		delLen := rand.Intn(maxPos - pos + 1)
		return &DeleteTx{Pos: pos, Len: delLen}
	case 2:
		newPos := rand.Intn(maxPos + 1)
		return &MoveCursorTx{To: newPos}
	default:
		return &InsertTx{Pos: pos, Text: "test"}
	}
}
