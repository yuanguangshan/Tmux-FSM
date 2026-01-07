package resolver

// ExecContext 执行上下文，用于隔离不同类型的执行
type ExecContext struct {
	FromMacro  bool // 是否来自宏播放
	FromRepeat bool // 是否来自重复操作
	FromUndo   bool // 是否来自撤销操作
}