package semantic

import (
	"fmt"
	"os/exec"
)

// Decide 将语义事实转换为事务
func Decide(f Fact) []Transaction {
	switch f.Kind() {
	case "delete":
		df := f.(*DeleteFact)
		return []Transaction{
			NewTmuxSendKeysTx(df.anchor.PaneID, []string{"-N", fmt.Sprint(len(df.text)), "Delete"}),
		}
	case "insert":
		inf := f.(*InsertFact)
		keys := append([]string{inf.text}, "Escape")
		return []Transaction{
			NewTmuxSendKeysTx(inf.anchor.PaneID, append([]string{"i"}, keys...)),
		}
	case "replace":
		rf := f.(*ReplaceFact)
		// 先删除旧文本，再插入新文本
		return []Transaction{
			NewTmuxSendKeysTx(rf.anchor.PaneID, []string{"-N", fmt.Sprint(len(rf.oldText)), "Delete"}),
			NewTmuxSendKeysTx(rf.anchor.PaneID, []string{rf.text, "Escape"}),
		}
	case "move":
		mf := f.(*MoveFact)
		// 简单的移动实现
		dx := mf.to.Col - mf.from.Col
		dy := mf.to.Line - mf.from.Line
		var keys []string
		if dx > 0 {
			for i := 0; i < dx; i++ {
				keys = append(keys, "Right")
			}
		} else if dx < 0 {
			for i := 0; i < -dx; i++ {
				keys = append(keys, "Left")
			}
		}
		if dy > 0 {
			for i := 0; i < dy; i++ {
				keys = append(keys, "Down")
			}
		} else if dy < 0 {
			for i := 0; i < -dy; i++ {
				keys = append(keys, "Up")
			}
		}
		return []Transaction{
			NewTmuxSendKeysTx(mf.anchor.PaneID, keys),
		}
	default:
		return nil
	}
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

// TmuxSendKeysTx tmux send-keys 事务
type TmuxSendKeysTx struct {
	Pane string
	Keys []string
}

// NewTmuxSendKeysTx 创建新的 TmuxSendKeysTx
func NewTmuxSendKeysTx(pane string, keys []string) *TmuxSendKeysTx {
	return &TmuxSendKeysTx{
		Pane: pane,
		Keys: keys,
	}
}

func (t *TmuxSendKeysTx) Apply() error {
	args := append([]string{"send-keys", "-t", t.Pane}, t.Keys...)
	return exec.Command("tmux", args...).Run()
}

func (t *TmuxSendKeysTx) Inverse() Transaction {
	// 对于 send-keys 操作，逆操作通常是撤销操作
	// 这里返回一个空操作作为占位符
	return &NoopTx{}
}

func (t *TmuxSendKeysTx) Kind() string {
	return "tmux_send_keys"
}

func (t *TmuxSendKeysTx) Tags() []string {
	return []string{"tmux", "atomic"}
}

func (t *TmuxSendKeysTx) CanMerge(next Transaction) bool {
	// 检查是否可以合并到下一个事务
	nextTx, ok := next.(*TmuxSendKeysTx)
	return ok && nextTx.Pane == t.Pane
}

func (t *TmuxSendKeysTx) Merge(next Transaction) Transaction {
	// 合并两个 TmuxSendKeysTx 事务
	nextTx := next.(*TmuxSendKeysTx)
	// 简单地将键序列连接
	mergedKeys := append(t.Keys, nextTx.Keys...)
	return &TmuxSendKeysTx{
		Pane: t.Pane,
		Keys: mergedKeys,
	}
}

// NoopTx 空操作事务
type NoopTx struct{}

func (n *NoopTx) Apply() error {
	return nil
}

func (n *NoopTx) Inverse() Transaction {
	return n
}

func (n *NoopTx) Kind() string {
	return "noop"
}

func (n *NoopTx) Tags() []string {
	return []string{"noop"}
}

func (n *NoopTx) CanMerge(next Transaction) bool {
	return false
}

func (n *NoopTx) Merge(next Transaction) Transaction {
	return next
}