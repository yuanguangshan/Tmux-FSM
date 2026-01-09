package main

import (
	"time"
)

// Transaction 代表一个可执行和可撤销的操作
type Transaction interface {
	Apply() error
	Inverse() Transaction
	Kind() string
	Tags() []string
	CanMerge(next Transaction) bool
	Merge(next Transaction) Transaction
}

// ChainInverse 将两个事务的逆操作链接起来
func ChainInverse(a, b Transaction) Transaction {
	return FuncTx{
		apply: func() error {
			_ = b.Inverse().Apply()
			return a.Inverse().Apply()
		},
		inverse: func() Transaction {
			// 逆操作的逆操作是原操作
			return FuncTx{
				apply: func() error {
					_ = a.Apply()
					return b.Apply()
				},
				inverse: func() Transaction { return ChainInverse(a, b) },
				kind: "chained_apply",
				tags: []string{"chained"},
			}
		},
		kind: "chained_inverse",
		tags: []string{"chained"},
	}
}

// FuncTx 是 Transaction 的通用实现
type FuncTx struct {
	apply   func() error
	inverse func() Transaction
	kind    string
	tags    []string
	mergeFn func(Transaction) (Transaction, bool)
}

func (t FuncTx) Apply() error {
	return t.apply()
}

func (t FuncTx) Inverse() Transaction {
	return t.inverse()
}

func (t FuncTx) Kind() string {
	return t.kind
}

func (t FuncTx) Tags() []string {
	return t.tags
}

func (t FuncTx) CanMerge(next Transaction) bool {
	if t.mergeFn == nil {
		return false
	}
	_, ok := t.mergeFn(next)
	return ok
}

func (t FuncTx) Merge(next Transaction) Transaction {
	if t.mergeFn == nil {
		return next
	}
	merged, ok := t.mergeFn(next)
	if !ok {
		return next
	}

	// 确保合并后的事务的逆操作是正确的
	// merged.Inverse() 应该等价于 Inverse(next) 再 Inverse(self)
	return FuncTx{
		apply: func() error { return merged.Apply() },
		inverse: func() Transaction { return ChainInverse(t.inverse(), next.Inverse()) },
		kind: merged.Kind(),
		tags: merged.Tags(),
		mergeFn: nil, // 合并后的事务不再支持进一步合并
	}
}

// TxRecord 事务记录
type TxRecord struct {
	Tx      Transaction
	Applied bool
	Failed  bool
	Time    time.Time
}

// Age 返回记录的时间差
func (r TxRecord) Age() time.Duration {
	return time.Since(r.Time)
}

// TxJournal 事务日志
type TxJournal struct {
	applied []TxRecord
	undone  []TxRecord
}

// NewTxJournal 创建新的事务日志
func NewTxJournal() *TxJournal {
	return &TxJournal{
		applied: make([]TxRecord, 0),
		undone:  make([]TxRecord, 0),
	}
}

// ApplyTxs 批量应用事务
func (j *TxJournal) ApplyTxs(txs []Transaction) error {
	var appliedNow []Transaction

	for _, tx := range txs {
		if err := tx.Apply(); err != nil {
			// 失败则立即回滚本批
			for i := len(appliedNow) - 1; i >= 0; i-- {
				_ = appliedNow[i].Inverse().Apply()
			}
			return err
		}
		appliedNow = append(appliedNow, tx)

		j.applied = append(j.applied, TxRecord{
			Tx:      tx,
			Applied: true,
			Time:    time.Now(),
		})
	}

	// 新历史出现 → Redo 失效
	j.undone = nil
	return nil
}

// hasTag 检查事务是否包含指定标签
func hasTag(tx Transaction, tag string) bool {
	tags := tx.Tags()
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// Undo 撤销最后一个事务（支持原子操作）
func (j *TxJournal) Undo() error {
	if len(j.applied) == 0 {
		return nil
	}

	rec := j.applied[len(j.applied)-1]
	atomic := hasTag(rec.Tx, "atomic")

	for {
		if err := rec.Tx.Inverse().Apply(); err != nil {
			return err
		}

		// 从 applied 移除
		j.applied = j.applied[:len(j.applied)-1]
		// 添加到 undone
		j.undone = append(j.undone, rec)

		if !atomic || len(j.applied) == 0 {
			break
		}

		// 检查前一个事务是否也是原子操作的一部分
		rec = j.applied[len(j.applied)-1]
		if !hasTag(rec.Tx, "atomic") {
			break
		}
	}
	return nil
}

// Redo 重做最后一个撤销的事务
func (j *TxJournal) Redo() error {
	if len(j.undone) == 0 {
		return nil
	}

	rec := j.undone[len(j.undone)-1]
	j.undone = j.undone[:len(j.undone)-1]

	if err := rec.Tx.Apply(); err != nil {
		return err
	}

	j.applied = append(j.applied, rec)
	return nil
}

// appendTx 添加事务并尝试合并
func (j *TxJournal) appendTx(tx Transaction) {
	// 新历史出现 → Redo 失效
	j.undone = nil

	n := len(j.applied)
	if n == 0 {
		j.applied = append(j.applied, TxRecord{Tx: tx, Applied: true, Time: time.Now()})
		return
	}

	last := j.applied[n-1].Tx
	if last.CanMerge(tx) {
		merged := last.Merge(tx)
		j.applied[n-1] = TxRecord{Tx: merged, Applied: true, Time: time.Now()}
	} else {
		j.applied = append(j.applied, TxRecord{Tx: tx, Applied: true, Time: time.Now()})
	}
}