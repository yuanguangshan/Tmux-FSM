package decide

import (
	"fmt"
	"os/exec"
	"tmux-fsm/semantic"
)

// Transaction 接口定义
type Transaction interface {
	Apply() error
	Inverse() Transaction
	Kind() string
	Tags() []string
	CanMerge(next Transaction) bool
	Merge(next Transaction) Transaction
}

// Decide 将语义事实转换为事务
func Decide(f semantic.Fact) []Transaction {
	switch f.Kind() {
	case "delete":
		return decideDelete(f)
	case "insert":
		return decideInsert(f)
	case "replace":
		return decideReplace(f)
	case "move":
		return decideMove(f)
	default:
		return nil
	}
}

// decideDelete 处理删除操作
func decideDelete(f semantic.Fact) []Transaction {
	r := f.GetRange()
	return []Transaction{
		&TmuxDeleteTx{
			PaneID: r.Start.PaneID,
			Range:  r,
			Text:   r.Text,
		},
	}
}

// decideInsert 处理插入操作
func decideInsert(f semantic.Fact) []Transaction {
	return []Transaction{
		&TmuxInsertTx{
			PaneID: f.GetAnchor().PaneID,
			Pos:    f.GetAnchor(),
			Text:   f.GetText(),
		},
	}
}

// decideReplace 处理替换操作
func decideReplace(f semantic.Fact) []Transaction {
	r := f.GetRange()

	return []Transaction{
		&TmuxDeleteTx{
			PaneID: r.Start.PaneID,
			Range:  r,
			Text:   r.Text,
		},
		&TmuxInsertTx{
			PaneID: r.Start.PaneID,
			Pos:    r.Start,
			Text:   f.GetText(),
		},
	}
}

// decideMove 处理移动操作
func decideMove(f semantic.Fact) []Transaction {
	from := f.GetRange().Start
	to := f.GetAnchor()
	return []Transaction{
		&TmuxMoveCursorTx{
			From:   from,
			To:     to,
			PaneID: from.PaneID,
		},
	}
}

// TmuxDeleteTx tmux 删除事务
type TmuxDeleteTx struct {
	PaneID string
	Range  semantic.Range
	Text   string
}

func (t *TmuxDeleteTx) Apply() error {
	// 执行删除操作
	args := []string{"send-keys", "-t", t.PaneID, "-N", fmt.Sprint(len(t.Text)), "Delete"}
	return exec.Command("tmux", args...).Run()
}

func (t *TmuxDeleteTx) Inverse() Transaction {
	return &TmuxInsertTx{
		PaneID: t.PaneID,
		Pos:    t.Range.Start,
		Text:   t.Text,
	}
}

func (t *TmuxDeleteTx) Kind() string {
	return "tmux_delete"
}

func (t *TmuxDeleteTx) Tags() []string {
	return []string{"tmux", "delete", "atomic"}
}

func (t *TmuxDeleteTx) CanMerge(next Transaction) bool {
	nextTx, ok := next.(*TmuxDeleteTx)
	return ok && nextTx.PaneID == t.PaneID
}

func (t *TmuxDeleteTx) Merge(next Transaction) Transaction {
	nextTx := next.(*TmuxDeleteTx)
	return &TmuxDeleteTx{
		PaneID: t.PaneID,
		Range: semantic.Range{
			Start: t.Range.Start,
			End:   nextTx.Range.End,
			Text:  t.Text + nextTx.Text,
		},
		Text: t.Text + nextTx.Text,
	}
}

// TmuxInsertTx tmux 插入事务
type TmuxInsertTx struct {
	PaneID string
	Pos    semantic.Anchor
	Text   string
}

func (t *TmuxInsertTx) Apply() error {
	// 执行插入操作
	args := []string{"send-keys", "-t", t.PaneID, "i", t.Text, "Escape"}
	return exec.Command("tmux", args...).Run()
}

func (t *TmuxInsertTx) Inverse() Transaction {
	return &TmuxDeleteTx{
		PaneID: t.PaneID,
		Range: semantic.Range{
			Start: t.Pos,
			End:   t.Pos,
			Text:  t.Text,
		},
		Text: t.Text,
	}
}

func (t *TmuxInsertTx) Kind() string {
	return "tmux_insert"
}

func (t *TmuxInsertTx) Tags() []string {
	return []string{"tmux", "insert", "atomic"}
}

func (t *TmuxInsertTx) CanMerge(next Transaction) bool {
	nextTx, ok := next.(*TmuxInsertTx)
	return ok && nextTx.PaneID == t.PaneID
}

func (t *TmuxInsertTx) Merge(next Transaction) Transaction {
	nextTx := next.(*TmuxInsertTx)
	return &TmuxInsertTx{
		PaneID: t.PaneID,
		Pos:    t.Pos,
		Text:   t.Text + nextTx.Text,
	}
}

// TmuxMoveCursorTx tmux 光标移动事务
type TmuxMoveCursorTx struct {
	From   semantic.Anchor
	To     semantic.Anchor
	PaneID string
}

func (t *TmuxMoveCursorTx) Apply() error {
	// 计算移动距离并执行移动操作
	dx := t.To.Col - t.From.Col
	dy := t.To.Line - t.From.Line

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

	if len(keys) > 0 {
		args := append([]string{"send-keys", "-t", t.PaneID}, keys...)
		return exec.Command("tmux", args...).Run()
	}
	return nil
}

func (t *TmuxMoveCursorTx) Inverse() Transaction {
	return &TmuxMoveCursorTx{
		From:   t.To,
		To:     t.From,
		PaneID: t.PaneID,
	}
}

func (t *TmuxMoveCursorTx) Kind() string {
	return "tmux_move"
}

func (t *TmuxMoveCursorTx) Tags() []string {
	return []string{"tmux", "move"}
}

func (t *TmuxMoveCursorTx) CanMerge(next Transaction) bool {
	return false // 移动操作一般不合并
}

func (t *TmuxMoveCursorTx) Merge(next Transaction) Transaction {
	return next
}