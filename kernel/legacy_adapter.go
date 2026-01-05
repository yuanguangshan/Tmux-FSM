package kernel

import (
	"yourmodule/intent"
)

// ⚠️ 这是唯一一个“脏接口”，但它把脏东西隔离了
func DecodeLegacyKey(key string) *intent.Intent {
	// 直接调用你现在 logic.go 里的函数
	// 示例（你按真实函数名替换）：

	action := ProcessKeyLegacy(key)
	if action == "" {
		return nil
	}

	return intent.FromLegacyAction(action)
}
