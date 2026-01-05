# Tmux-FSM 基线行为文档

**版本**: pre-weaver-migration  
**日期**: 2026-01-05  
**目的**: 记录重构前的正确行为，作为后续验证标准

---

## 1. 基本移动命令

### 场景
```
初始文本: "hello world test"
光标位置: h (第一个字符)
```

### 操作序列
- `l l l` → 光标移动到 'l'
- `w` → 光标跳到 'world'
- `$` → 光标跳到行尾
- `0` → 光标回到行首

### 预期结果
✅ 光标位置正确  
✅ 状态栏显示 "NORMAL"  
✅ 无错误日志

---

## 2. 删除操作 + Undo

### 场景
```
初始文本: "one two three four"
光标位置: 'o' (one)
```

### 操作序列
1. `dw` → 删除 "one "
2. `dw` → 删除 "two "
3. `dw` → 删除 "three "
4. `u` → 撤销最后一次删除
5. `u` → 再撤销一次
6. `u` → 再撤销一次

### 预期结果
✅ 最终文本恢复为 "one two three four"  
✅ Anchor 定位精确（exact）  
✅ 状态栏显示 Undo 安全级别（如果是 fuzzy 会显示 ~UNDO）  
✅ 日志中记录 Fact 和 Transaction

---

## 3. 移动光标后 Delete

### 场景
```
初始文本: "apple banana cherry"
光标位置: 'a' (apple)
```

### 操作序列
1. `w` → 移动到 'banana'
2. `dw` → 删除 "banana "
3. `u` → 撤销

### 预期结果
✅ 删除正确的词（banana）  
✅ Undo 后光标和文本都恢复  
✅ Anchor 能正确解析（即使光标移动了）

### 关键验证点
- Anchor.LineHash 应该匹配当前行
- ResolveAnchor 应该返回 ResolveExact
- 如果行内容变化，应该在 5 行窗口内 fuzzy 匹配

---

## 4. 跨 Pane 操作

### 场景
- Pane 1: 编辑文本 "test1"
- Pane 2: 编辑文本 "test2"

### 操作序列
1. 在 Pane 1 执行 `dw`
2. 切换到 Pane 2
3. 执行 `dw`
4. 切换回 Pane 1
5. 执行 `u`

### 预期结果
✅ 每个 pane 的状态独立  
✅ Undo stack 按 pane 隔离  
✅ 状态栏正确显示当前 pane 的状态

---

## 5. 文本对象操作

### 场景
```
初始文本: 'hello "world" test'
光标位置: 'w' (world 内部)
```

### 操作序列
- `diw` → 删除 "world"（不含引号）
- `u` → 撤销
- `da"` → 删除 "world" 含引号
- `u` → 撤销

### 预期结果
✅ `diw` 删除 "world"，保留引号  
✅ `da"` 删除整个 "world" 包括引号  
✅ Undo 正确恢复

---

## 6. Visual 模式

### 场景
```
初始文本: "select this text"
光标位置: 's' (select)
```

### 操作序列
1. `v` → 进入 VISUAL_CHAR 模式
2. `l l l` → 扩展选择
3. `d` → 删除选中内容
4. `u` → 撤销

### 预期结果
✅ 状态栏显示 "VISUAL"  
✅ 删除选中的字符  
✅ Undo 恢复

---

## 7. 搜索功能

### 场景
```
文本内容:
line1: apple
line2: banana
line3: apple
```

### 操作序列
1. `/apple` + Enter → 搜索
2. `n` → 下一个匹配
3. `N` → 上一个匹配

### 预期结果
✅ 正确跳转到匹配位置  
✅ `n` 向前，`N` 向后  
✅ 状态栏显示搜索模式

---

## 8. FSM 层级切换

### 场景
初始状态: NAV 层

### 操作序列
1. `g` → 进入 GOTO 层
2. 状态栏应显示 "GOTO" 和提示
3. `h` → 执行 far_left 动作
4. 或等待 800ms 超时自动退出

### 预期结果
✅ 状态栏显示当前层级  
✅ 层级按键正确响应  
✅ 超时自动回到 NAV  
✅ 执行动作后立即回到 NAV（非 sticky 层）

---

## 9. Undo 安全级别

### Exact Undo
- Anchor.LineHash 完全匹配
- 状态栏无特殊标记

### Fuzzy Undo
- Anchor 在 ±5 行窗口内找到
- 状态栏显示 "~UNDO"
- 日志记录 fuzzy 级别

### Failed Undo
- Anchor 无法解析
- 状态栏显示 "!UNDO_FAIL"
- 整个 Transaction 被跳过
- 可通过 `?` 查看失败原因

---

## 10. 帮助系统

### 操作
- `?` → 显示帮助弹窗

### 预期结果
✅ 显示完整的 cheat sheet  
✅ 包含作者信息  
✅ 如果有 Undo 失败，显示失败原因  
✅ 弹窗尺寸适配内容

---

## 关键不变量（Invariants）

这些是**任何重构都必须保持**的核心约束：

1. **Anchor Primacy**: 所有可撤销操作必须基于 Anchor，不能基于光标位置
2. **Transaction Atomicity**: 一个操作的所有 Fact 要么全部成功，要么全部失败
3. **Exactness Preference**: 优先尝试 exact 匹配，只在必要时 fuzzy
4. **Fuzzy Transparency**: Fuzzy undo 必须在 UI 明确标识
5. **No Partial Trust**: 如果 Transaction 中任何 Fact 失败，整个 TX 标记为 skipped

---

## 测试通过标准

✅ 所有 8 个测试场景手动验证通过  
✅ 日志中无 panic 或 critical error  
✅ Undo/Redo 行为符合预期  
✅ 状态栏显示正确  
✅ 跨 pane 状态隔离正确

---

**签名**: _______________  
**日期**: _______________  
**备注**: _______________
