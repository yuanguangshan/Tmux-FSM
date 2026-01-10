package store

import (
	"fmt"
	"rhm-go/core/change"
)

type FileSystemOp struct {
	Kind   string
	Arg    string
	IsNoOp bool
}

func (op FileSystemOp) Describe() string {
	if op.IsNoOp {
		return fmt.Sprintf("NoOp(%s)", op.Arg)
	}
	return fmt.Sprintf("%s(%s)", op.Kind, op.Arg)
}

func (op FileSystemOp) ToNoOp() change.ReversibleChange {
	return FileSystemOp{Kind: "NoOp", Arg: "Neutralized", IsNoOp: true}
}

func (op FileSystemOp) Downgrade() change.ReversibleChange {
	if op.Kind == "Delete" {
		return FileSystemOp{Kind: "Move", Arg: "Trash/" + op.Arg}
	}
	return nil
}

func (op FileSystemOp) Hash() string { return op.Kind + ":" + op.Arg }
