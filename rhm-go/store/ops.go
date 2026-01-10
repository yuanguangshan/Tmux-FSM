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

func (op FileSystemOp) GetFootprints() []change.Footprint {
	if op.IsNoOp { return nil }
	switch op.Kind {
	case "Edit":
		return []change.Footprint{{ResourceID: op.Arg, Mode: change.Shared}}
	case "Delete":
		return []change.Footprint{{ResourceID: op.Arg, Mode: change.Exclusive}}
	case "Create":
		return []change.Footprint{{ResourceID: op.Arg, Mode: change.Create}}
	}
	return nil
}

func (op FileSystemOp) Describe() string {
	if op.IsNoOp { return "NoOp(Neutralized)" }
	return fmt.Sprintf("%s(%s)", op.Kind, op.Arg)
}

func (op FileSystemOp) ToNoOp() change.ReversibleChange {
	return FileSystemOp{IsNoOp: true}
}

func (op FileSystemOp) Downgrade() change.ReversibleChange {
	if op.Kind == "Delete" {
		return FileSystemOp{Kind: "Move", Arg: "Trash/" + op.Arg}
	}
	return nil
}

func (op FileSystemOp) Hash() string { return op.Kind + ":" + op.Arg }
