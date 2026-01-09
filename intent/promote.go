package intent

// Promote 是 GrammarIntent → Intent 的唯一合法通道
// Grammar 不允许直接构造 Intent
func Promote(g *GrammarIntent) *Intent {
	if g == nil {
		return nil
	}

	i := &Intent{
		Kind:   g.Kind,
		Count:  g.Count,
		Motion: g.Motion,
	}

	// Operator 提升（强类型）
	if g.Op != nil {
		i.Operator = g.Op
	}

	return i
}
