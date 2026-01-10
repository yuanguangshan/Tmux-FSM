# Legacy Intent 删除 Checklist

这是用于指导从 Legacy Intent 迁移到 Native Intent 的检查清单。每个阶段都必须完成所有检查项才能进入下一阶段。

## Phase 0：准备期（已完成）

- [x] 所有 legacy intent 都有 Anchor
- [x] Anchor 明确标记 legacy（legacy:: 前缀）
- [x] Resolver 能清洗 legacy anchor
- [x] Projection / Executor 不接触 legacy

## Phase 1：Builder 接管（进行中）

- [x] FSM 中新增 native intent path（与 legacy 并存）
- [x] `dw / cw / dd` 等命令已通过 Native Intent Builder 实现
- [ ] 所有新功能禁止使用 legacy intent
- [ ] IntentBuilder 成为唯一 new Intent 入口
- [ ] Resolver 断言：native intent 不得包含 legacy anchor

```go
if intent.IsNative() && intent.HasLegacyAnchor() {
    panic("native intent must not contain legacy anchor")
}
```

## Phase 2：FSM 去 legacy 化

- [ ] 每个 legacy key binding 都有 native builder 对应
- [ ] FSM 不再产生 action string
- [ ] `processKeyLegacy()` 标记为 deprecated
- [ ] legacy intent bridge 不再新增代码

## Phase 3：硬删除（不可回滚）

- [ ] 删除 legacy LineID 生成逻辑
- [ ] 删除 row / col 依赖
- [ ] 删除 tmux-aware 逻辑
- [ ] 删除 legacy intent bridge 文件
- [ ] Resolver 中删除 legacy 清洗器

## 代码审查标准

### PR 必须检查项

1. **新代码不能使用 legacy intent bridge**
   - 不得调用 `actionStringToIntent` 或 `actionStringToIntentWithLineInfo`
   - 不得依赖 `Meta["line_id"]`、`Meta["row"]`、`Meta["col"]`
   - 必须使用 `IntentBuilder` 创建新 Intent

2. **Anchor 使用规范**
   - Native intent 必须使用语义 Anchor，而非坐标 Anchor
   - 不得在 Native intent 中使用 `legacy::` 前缀的 LineID
   - Undo/Redo intent 的 Anchor 仅用于 projection 兼容性

3. **Resolver 兼容性**
   - 新 Intent 必须能被 Resolver 正确解析
   - 不得绕过 Resolver 直接执行操作

### 迁移优先级

1. **高优先级命令**：
   - 移动命令 (h, j, k, l, w, b, e, 0, $, G, gg)
   - 删除命令 (x, X, dd, dw, d + motion)
   - 修改命令 (c + motion, cc, cw)
   - 复制命令 (y + motion, yy, yw)

2. **中优先级命令**：
   - 视觉模式 (v, V, 字符/行选择)
   - 搜索命令 (/, n, N)
   - 撤销/重做 (u, C-r)

3. **低优先级命令**：
   - 特殊命令 (f, F, t, T, r, ~, .)
   - 文本对象 (iw, aw, etc.)

## 测试要求

### 单元测试覆盖

- [ ] Legacy Intent 路径的回归测试
- [ ] Native Intent 路径的正确性测试
- [ ] Resolver 对两种 Intent 的处理测试
- [ ] 从 Legacy 到 Native 的过渡兼容性测试

### 集成测试覆盖

- [ ] 所有迁移的命令在实际 tmux 环境中正常工作
- [ ] Undo/Redo 在 Native Intent 下正常工作
- [ ] 快照感知功能正常工作

## 迁移后验证

### 功能验证

1. **行为一致性**：Native Intent 实现与 Legacy 实现行为完全一致
2. **性能一致性**：Native Intent 不引入额外性能开销
3. **错误处理一致性**：Native Intent 错误处理与 Legacy 一致

### 架构验证

1. **解耦验证**：Intent 创建不再依赖 tmux 坐标
2. **语义验证**：Intent 只表达意图，不包含执行细节
3. **扩展验证**：新功能可以轻松通过 Native Intent 实现

## 安全网

### 回滚计划

- [ ] Legacy 代码保持可运行状态直到完全迁移
- [ ] 提供开关控制使用 Legacy 或 Native 路径
- [ ] 完整的回归测试套件验证 Legacy 路径

### 监控指标

- [ ] Legacy 路径使用率逐渐降低
- [ ] Native 路径错误率保持低位
- [ ] 性能指标无明显下降