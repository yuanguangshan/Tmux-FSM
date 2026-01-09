package builder

import (
	"sort"
	"tmux-fsm/intent"
)

// CompositeBuilder 组合构建器
type CompositeBuilder struct {
	builders []Builder
}

// NewCompositeBuilder 创建组合构建器
func NewCompositeBuilder() *CompositeBuilder {
	cb := &CompositeBuilder{
		builders: []Builder{
			&MoveBuilder{},
			&TextObjectBuilder{},
			&OperatorBuilder{},
			&MacroBuilder{},
		},
	}
	cb.sort()
	return cb
}

// AddBuilder 添加构建器
func (cb *CompositeBuilder) AddBuilder(builder Builder) {
	cb.builders = append(cb.builders, builder)
	cb.sort()
}

// Build 尝试使用所有构建器构建Intent
func (cb *CompositeBuilder) Build(ctx BuildContext) (*intent.Intent, bool) {
	for _, builder := range cb.builders {
		intent, ok := builder.Build(ctx)
		if ok {
			return intent, true
		}
	}
	return nil, false
}

// sort 按优先级排序构建器
// Builders are evaluated in order.
// Order MUST reflect semantic priority.
func (cb *CompositeBuilder) sort() {
	sort.SliceStable(cb.builders, func(i, j int) bool {
		return cb.builders[i].Priority() > cb.builders[j].Priority()
	})
}
