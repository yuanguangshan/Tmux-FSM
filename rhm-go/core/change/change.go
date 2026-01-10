package change

type MutationType int

const (
	ReplaceOp MutationType = iota
	// RemoveNode (Reserved for future)
)

// ReversibleChange 定义了时间旅行的物理定律
type ReversibleChange interface {
	Describe() string
	ToNoOp() ReversibleChange    // 返回 nil 表示不支持
	Downgrade() ReversibleChange // 返回 nil 表示不支持
	Hash() string                // 用于指纹计算
}

type Mutation struct {
	Type   MutationType
	Target string
	NewOp  ReversibleChange
}

func (m Mutation) String() string {
	if m.Type == ReplaceOp {
		return "Mutate " + m.Target + " -> " + m.NewOp.Describe()
	}
	return "Unknown Mutation"
}
