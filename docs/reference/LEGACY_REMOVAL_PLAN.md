# Legacy 删除清单

## 目标
完全移除 legacy 逻辑，使系统完全基于 FSM → Grammar → Intent → Kernel 架构运行。

## 删除前检查清单

### Grammar 覆盖确认
在删除任何 legacy 代码前，确保 Grammar 能处理：

- [x] hjkl 移动
- [x] w b e 移动  
- [x] $ 0 移动
- [x] gg G 移动
- [x] f F t T 移动
- [x] d y c 操作符
- [x] dd yy cc 单行操作
- [x] dw diw da( 等 text-object
- [x] 2dw 3dd 等 count
- [x] i a o 进入插入
- [x] v V 进入 visual
- [x] . 重复
- [x] u Ctrl-r 撤销重做

### 单元测试覆盖
确保所有 Grammar 单元测试通过：

```bash
go test ./planner/... -v
```

## 可删除的文件/函数

### 1. legacy_logic.go
```bash
rm legacy_logic.go
```

### 2. intent_bridge.go
```bash
rm intent_bridge.go
```

### 3. logic.go 中的 legacy 函数
删除以下函数：
- `processKeyToIntent`
- `processKey`
- `processKeyLegacy`
- `handleNormal`
- `handleOperatorPending`
- `handleRegisterSelect`
- `handleVisualChar`
- `handleVisualLine`
- `handleSearch`
- `handleTextObjectPending`
- `handleFindChar`
- `handleMotionPending`
- `handleReplaceChar`

### 4. main.go 中的 legacy 调用
删除相关的 legacy 处理逻辑

## 重构后验证步骤

1. **Grammar 完整性测试**：运行所有 Grammar 单元测试
2. **集成测试**：手动测试 `d2w`, `ci(`, `3gg` 等复杂组合
3. **性能测试**：确保 Grammar 解析性能可接受
4. **删除 legacy**：按文件逐一删除，每次删除后测试

## 完整的 Grammar 覆盖表

### Motion（必须 100% 覆盖）
- 基础字符移动: h j k l
- 词级移动: w b e ge
- 行内移动: 0 ^ $
- 行/屏幕移动: gg G H M L
- 查找型: f{c} F{c} t{c} T{c}
- 文本对象: iw aw i( a( i{ a{ i" a" a' i'

### Operator（Grammar 核心）
- d: delete
- c: change  
- y: yank
- > <: indent
- =: reindent

### Count（Grammar 全权负责）
- 3w: move 3 words
- d2w: delete 2 words
- 2dw: delete 2 words

### Mode 切换（Intent 级）
- i a o O: EnterInsert
- v V Ctrl-v: EnterVisual
- Esc: EnterNormal

### 重复 / 历史
- .: RepeatLast
- u: Undo
- Ctrl-r: Redo

## Kernel.Decide 的最终规范实现

```go
func (k *Kernel) Decide(key string) *Decision {
    // 1. FSM 永远先拿 key
    if k.FSM != nil {
        var lastIntent *intent.Intent

        // 创建一个 GrammarEmitter 来处理 token
        grammarEmitter := &GrammarEmitter{
            grammar: k.Grammar,
            callback: func(intent *intent.Intent) {
                lastIntent = intent
            },
        }

        // 添加 GrammarEmitter 到 FSM
        k.FSM.AddEmitter(grammarEmitter)

        // 让 FSM 处理按键
        dispatched := k.FSM.Dispatch(key)

        // 移除 GrammarEmitter
        k.FSM.RemoveEmitter(grammarEmitter)

        if dispatched && lastIntent != nil {
            // 直接执行意图，而不是返回决策
            if k.FSM != nil {
                _ = k.FSM.DispatchIntent(lastIntent)
            }
            return nil // 意图已直接执行
        }

        if dispatched {
            return nil // FSM处理了按键，但没有产生意图（合法状态）
        }
    }

    // 没有 legacy fallback，所有逻辑都由 Grammar 处理
    return nil
}
```

## Grammar 单元测试策略

使用表驱动测试，覆盖所有关键用例：

```go
func TestGrammarComplete(t *testing.T) {
    cases := []struct {
        keys   []string
        intent Intent
    }{
        {"j", NewMoveIntent(MoveDown, 1)},
        {"3j", NewMoveIntent(MoveDown, 3)},
        {"dw", NewOperatorMotionIntent(OpDelete, MoveWord, 1)},
        {"d2w", NewOperatorMotionIntent(OpDelete, MoveWord, 2)},
        {"gg", NewMoveIntent(MoveFileStart, 1)},
        {"fa", NewMoveIntent(MoveChar{Char: 'a', Sub: MPF}, 1)},
        {"di(", NewOperatorTextObjectIntent(OpDelete, TextParen(TOPInner), 1)},
        // ... 更多测试用例
    }
    
    for _, tc := range cases {
        g := NewGrammar()
        var finalIntent Intent
        for _, key := range tc.keys {
            if intent, ok := g.Consume(RawToken{Value: key}); ok {
                finalIntent = intent
            }
        }
        assert.Equal(t, tc.intent, finalIntent)
    }
}
```