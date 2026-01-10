package change

type MutationType int

const ReplaceOp MutationType = iota

type AccessMode int

const (
	Shared    AccessMode = iota // 共享访问（读）
	Exclusive                 // 独占访问（写/删）
	Create                    // 命名空间占用（新建）
)

// Footprint 描述操作在资源空间留下的痕迹
type Footprint struct {
	ResourceID string
	Mode       AccessMode
}

// ReversibleChange 定义了时间旅行的物理定律
type ReversibleChange interface {
	Describe() string
	ToNoOp() ReversibleChange    // 返回 nil 表示不支持
	Downgrade() ReversibleChange // 返回 nil 表示不支持
	Hash() string                // 用于指纹计算
}

// SemanticChange 扩展接口：支持足迹获取
type SemanticChange interface {
	ReversibleChange
	GetFootprints() []Footprint
}

type Mutation struct {
	Type   MutationType
	Target string
	NewOp  ReversibleChange
}

func (m Mutation) String() string {
	return "Mutate " + m.Target + " -> " + m.NewOp.Describe()
}
