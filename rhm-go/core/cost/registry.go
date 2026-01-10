package cost

import "rhm-go/core/change"

type Cost int

const (
	Zero        Cost = 0
	Tweak       Cost = 20
	Downgrade   Cost = 50
	Neutralize  Cost = 100
	Destructive Cost = 500
	Infinite    Cost = 10000
)

type Context struct{}

var modelRegistry = make(map[string]Model)

func RegisterModel(name string, model Model) {
	modelRegistry[name] = model
}

func GetModel(name string) Model {
	if model, ok := modelRegistry[name]; ok {
		return model
	}
	return DefaultModel{}
}

type Model interface {
	Calculate(m change.Mutation, ctx Context) Cost
}

type DefaultModel struct{}

func (d DefaultModel) Calculate(m change.Mutation, ctx Context) Cost {
	if m.Type == change.ReplaceOp {
		desc := m.NewOp.Describe()
		if desc == "NoOp(Neutralized)" {
			return Neutralize
		}
		// 启发式检测 Downgrade
		return Downgrade
	}
	return Destructive
}
