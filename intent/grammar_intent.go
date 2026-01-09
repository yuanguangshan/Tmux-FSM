package intent

// GrammarIntent 是 Grammar 专用的意图类型，只包含 Grammar 可以设置的字段
type GrammarIntent struct {
	Kind   IntentKind
	Count  int
	Motion *Motion
	Op     *OperatorKind
}
