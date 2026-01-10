
# Intent & Builder Modules — Implementation Documentation

> 本文档描述的是 **Intent 构建与语义表示层**，  
> 它位于 **FSM RawToken 之后、执行层之前**，  
> 负责把“动作语义”稳定地表达为 **可迁移、可比较的 Intent**。

---

# 一、总体结构关系

```
FSM RawToken / Action
        ↓
    BuildContext
        ↓
CompositeBuilder
        ↓
     Intent
        ↓
  (Promote / Migrate)
        ↓
 Execution Layer
```

---

# 二、Builder 子系统

> **Builder 是唯一合法的新 Intent 构造路径**

---

## `builder/doc.go` —— 权威声明

代码中的规则是**强约束**：

- ✅ Builder 是 **唯一** 原生 Intent 构建方式
- ❌ 不得引入 legacy 构造逻辑
- ✅ Builder 只做 **语义判断**
- ✅ Priority 决定匹配顺序

---

## `builder/builder.go`

### BuildContext（输入）

```go
type BuildContext struct {
    Action       string
    Command      string
    Count        int
    PaneID       string
    SnapshotHash string
    Meta         map[string]interface{}
}
```

**事实说明：**

| 字段 | 当前用途 |
|----|----|
| Action | 核心匹配字段（字符串） |
| Count | Vim 风格计数 |
| PaneID | 路由信息 |
| Meta | 传递 register / operator 等 |
| Command | 预留（未使用） |
| SnapshotHash | 预留（未使用） |

---

### Builder 接口

```go
type Builder interface {
    Priority() int
    Build(ctx BuildContext) (*intent.Intent, bool)
}
```

- `Build` **必须是纯函数**
- 返回 `(intent, true)` 即表示匹配成功
- 不允许副作用

---

## `builder/composite_builder.go`

### CompositeBuilder

```go
type CompositeBuilder struct {
    builders []Builder
}
```

默认注册顺序（按优先级排序后）：

| Builder | Priority |
|----|----|
| TextObjectBuilder | 15 |
| MoveBuilder | 10 |
| MacroBuilder | 8 |
| OperatorBuilder | 5 |

---

### Build 行为

```go
func (cb *CompositeBuilder) Build(ctx BuildContext)
```

- **按优先级顺序**
- 第一个成功即返回
- 不做回溯、不合并

---

## `builder/move_builder.go`

### MoveBuilder（立即 Motion）

- **最高即时执行优先级**
- 不依赖 operator

构造的 Intent 特点：

```go
Kind   = IntentMove
Target = SemanticTarget{Kind: TargetChar / TargetLine}
Count  = ctx.Count
```

方向通过 `Target.Direction` 表达。

---

## `builder/operator_builder.go`

### OperatorBuilder（等待 motion）

- 优先级最低
- 仅声明“我要做什么操作”

```go
Kind   = IntentOperator
Target = TargetChar (占位)
Meta["operator"] = OpDelete / OpYank / OpChange
```

⚠️ **重要事实**

> Operator 仍然编码在 `Meta` 中  
> 这是明确标注的迁移态实现

---

## `builder/text_object.go`

### TextObjectBuilder（最高优先级）

- 明确语义范围
- 直接生成 **完整 operator intent**

```go
Target.Kind  = TargetTextObject
Target.Value = "inner_paren" / "around_word" 等
Meta["operator"] = OpDelete / OpChange / OpYank
```

📌 **这是当前系统中语义最完整的一类 Intent**

---

## `builder/macro_builder.go`

### MacroBuilder

生成：

```go
Kind = IntentMacro
Target.Scope = start | stop | play
Meta["operation"]
Meta["register"]
```

- register 缺省为 `"a"`
- 不涉及 motion / operator

---

## `builder/intent_diff.go`

### IntentDiff（迁移对比工具）

```go
type IntentDiff struct {
    Field
    Legacy
    Native
}
```

用于：

- 对比 legacy intent vs builder intent
- **只比较可观测语义字段**
- 不比较 Meta 深层结构

---

## `builder/semantic_equal.go`

### SemanticEqual

支持两种模式：

| 模式 | 行为 |
|----|----|
| CompareMigration | 忽略 PaneID |
| CompareStrict | PaneID 也必须一致 |

比较字段：

- Kind
- Target.*
- Count

---

# 三、Intent 核心模型

---

## `intent/grammar_intent.go`

### GrammarIntent（受限 Intent）

```go
type GrammarIntent struct {
    Kind
    Count
    Motion
    Op
}
```

规则：

- Grammar **只能构造这个**
- Grammar **不能触碰 Intent**

---

## `intent/promote.go`

### Promote（唯一合法提升路径）

```go
func Promote(g *GrammarIntent) *Intent
```

行为：

1. 初始化空 Meta
2. 若存在 Motion：
   - 同时保留强类型 Motion
   - 生成 legacy Meta["motion"]
3. 设置：
   - Kind
   - Count
   - Operator（强类型）
4. AllowPartial = true（仅 IntentMove）

📌 **Promote 是迁移桥的“闸门”**

---

### populateLegacyMotionMeta（桥接层）

- 将强类型 Motion 映射为旧字符串 motion
- 只覆盖当前已支持的 motion
- 未生成字符串 → Meta 不写入

---

## `intent/intent.go`

### Intent 结构（真实执行模型）

```go
type Intent struct {
    Kind
    Target        // ⚠️ deprecated
    Count
    Meta          // ⚠️ deprecated
    PaneID
    SnapshotHash
    AllowPartial
    Anchors
    UseRange
    Motion        // ✅ 强类型
    Operator      // ✅ 强类型
}
```

**事实状态**

| 字段 | 状态 |
|----|----|
| Target | 迁移期 |
| Meta | 迁移期 |
| Motion | ✅ 新主通道 |
| Operator | ✅ 新主通道 |
| Range / Anchors | 未被使用 |

---

## IntentKind / TargetKind

- 直接 re-export `weaver/core`
- Intent 层 **不定义语义，只承载**

---

## Getter 方法

- 全部是薄封装
- 没有副作用
- 主要用于接口适配

---

# 四、Motion / Range / TextObject

---

## `intent/motion.go`

### Motion

```go
type Motion struct {
    Kind
    Count
    Direction
    Find
    Range
}
```

- 强类型 motion 表达
- `Find` / `Range` 互斥
- Count 可独立于 Intent.Count

---

## `intent/range.go`

### RangeMotion

当前仅支持：

- 行首 (`0`)
- 行尾 (`$`)
- 文本对象（预留）

---

## `intent/text_object.go`

### TextObject

```go
type TextObject struct {
    Scope  Inner | Around
    Object Word | Paren | Quote...
}
```

⚠️ 当前 **Builder 未使用此强类型结构**  
TextObjectBuilder 仍通过字符串 Value 表达

---

# 五、当前系统的真实状态总结

✅ **已经成立的事实**

- Builder 是 Intent 构建唯一入口
- Intent 已支持强类型 Motion / Operator
- Promote 是 Grammar → Intent 的硬边界
- 迁移态被清晰标注（Meta / Target）

❌ **尚未完成**

- TextObject 强类型化
- Operator 完全脱离 Meta
- Range-based 执行
- 多 Anchor / 多光标

---

# 六、一句话定性

> **这是一个处于“语义冻结 + 表达升级”阶段的 Intent 系统：**
>
> - Builder 负责“我想干什么”
> - Grammar 负责“我看懂了什么”
> - Intent 负责“我能被执行什么”
>
> 所有迁移路径都被显式标注，没有隐式魔法。

---

