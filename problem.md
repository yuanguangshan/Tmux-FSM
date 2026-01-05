很好，这一轮日志**已经把真正的“不可恢复”根因彻底暴露出来了**。  
结论先行：

> ❌ **现在的问题已经不是 tmux buffer 的“问号”问题了**  
> ✅ **真正的根因是：Weaver Undo 恢复的文本内容本身就是错的（只有 1 个字符）**

下面我会非常明确地告诉你：

1. **为什么现在会出现 `Text length: 1 chars`**
2. **为什么第一次 `u` 看似成功、第二次一定 `UNDO_FAIL`**
3. **这不是物理层问题，而是语义→事实构建层的致命漏洞**
4. **你应该怎么修，且只需要修一个地方**

---

## 一、关键证据：日志已经“自白”

你日志里最重要的一行是这一句（你可能已经注意到了，但我们现在给它定罪）：

```
[WEAVER] Undo: Applying 1 inverse facts. Text length: 1 chars.
```

你测试的是：

```
hello universe
dw   （删除 universe，8 个字符）
u
```

**但 Weaver 认为需要恢复的 OldText 只有 1 个字符。**

这意味着什么？

> ✅ **Undo 的失败不是“没贴回去”**  
> ❌ **而是“从一开始就只记录了 1 个字符”**

---

## 二、为什么第一次 `u` 还能“看起来成功”

你这里的行为顺序是：

1. `dw`
2. `u` ✅（看似成功）
3. `u` ❌（world drift）

### 原因拆解

### ✅ 第一次 `u`

- Weaver 用 **错误的 inverse fact**
- 把 **1 个字符** 插回去了
- 物理世界 **已经被污染**
- 但 Weaver **自己不知道它错了**

于是：

- `PostSnapshotHash` ✅（基于错误世界生成）
- History 状态 ✅

### ❌ 第二次 `u`

- Weaver 试图 Undo 上一个 Undo
- 对比 SnapshotHash
- 发现现实世界 ≠ 预期
- **触发世界漂移保护**

于是你看到：

```
Engine Error: world drift: snapshot mismatch
!UNDO_FAIL
```

✅ **这一步是对的，甚至是系统健康的表现**

---

## 三、真正的根因（非常重要）

### ❌ 根因不是：

- tmux `set-buffer`
- tmux `paste-buffer`
- 编码
- jump
- 物理坐标

### ✅ 根因是：

> **ShellFactBuilder / PassthroughResolver 在构建 Delete Fact 时，根本不知道自己删了多少文本**

---

## 四、罪魁祸首在哪里（精确定位）

### 你现在的事实构建路径是：

```
Legacy execute.go
   ↓（Reverse Bridge）
InjectLegacyTransaction
   ↓
convertLegacyFactToCore
   ↓
core.FactDelete { Payload.OldText = lf.Target.Text }
```

但你 **现在这次测试用的不是 Legacy delete**，而是：

```
Weaver 模式下的：
dw → IntentDelete → ShellFactBuilder
```

而我们来看：

### 📄 `weaver/logic/shell_fact_builder.go`

```go
case core.IntentDelete:
    facts = append(facts, core.Fact{
        Kind:   core.FactDelete,
        Anchor: anchor,
        Meta:   meta,
    })
```

❌ **这里根本没有 OldText**

---

## 五、那 OldText 是在哪里“补”的？

你在 `PassthroughResolver` 里做了这件事（Phase 5.3）：

```go
if f.Kind == core.FactDelete && payload.OldText == "" {
    if ra.End >= ra.Start {
        lineText := ...
        payload.OldText = lineText[ra.Start : ra.End+1]
    }
}
```

### ❌ 问题就在这里

`ra.Start` / `ra.End` 是怎么来的？

👉 来自 **AnchorWord / AnchorAtCursor 的语义解析**

而 **`dw` 的语义不是“一个 word anchor”**  
而是：

> **“从 cursor 到下一个 word boundary 的 motion”**

你现在的语义系统：

| Vim 动作 | 真实含义 | Weaver 当前理解 |
|--------|--------|----------------|
| dw | motion-based delete | ❌ AnchorWord |
| cw | motion + insert | ❌ AnchorWord |
| d$ | range delete | ❌ AnchorLine |

结果就是：

> **Resolver 只能解析出 1 个字符宽的范围**

---

## 六、为什么一定是 1 个字符？

因为：

- `AnchorAtCursor`
- `Start == End == col`
- `End+1 - Start == 1`

所以：

```
Text length: 1 chars
```

✅ 这完全解释了你的日志  
✅ 也解释了为什么 pipe 注入“成功但没意义”

---

## 七、结论（非常关键）

> ❗ **你现在的 Weaver Undo 在“dw / cw / d$”这类动作上是“逻辑上不可能正确”的**

不是 bug  
是 **语义模型不完整**

---

## 八、唯一正确的修复路线（不多不少）

### ✅ 正确解法（推荐）

**在 Weaver Phase 5+ 中：**

> ❌ 不允许 ShellFactBuilder 生成 Delete Fact  
> ✅ 必须通过 **Legacy execute.go 的 captureShellDelete() 反向桥接**

也就是说：

```text
Shell editing
   ↓
必须走 Legacy physical delete
   ↓
捕获真实 deletedText
   ↓
InjectLegacyTransaction
   ↓
Weaver History
```

✅ **这正是你当初 Phase 4 设计 Reverse Bridge 的理由**

---

## 九、你可以立刻做的最小修复（10 分钟级）

### ✅ 临时硬防线（强烈建议）

在 Weaver 模式下：

```go
if intent.Kind == IntentDelete && isShellPane {
    // 禁止 Weaver Planner 生成 delete
    return
}
```

或者：

```go
ShellFactBuilder.Build:
case IntentDelete:
    return nil, nil, ErrUnsupportedSemanticDelete
```

👉 强制回退到 Legacy delete + 注入历史

---

## 十、你这次其实赢得很漂亮

你现在已经做到：

- ✅ 防乱码 ✅
- ✅ 物理层防御 ✅
- ✅ Undo 原子性 ✅
- ✅ World Drift 保护 ✅

**现在暴露出来的这个问题，正是 Weaver 真正进入 Phase 5/6 必然要面对的核心难题：**

> 👉 **“motion-based edit 无法通过纯语义 anchor 表达”**

这不是你写错了  
这是你走到了**编辑器内核设计的硬边界**

---

如果你愿意，下一步我可以直接帮你：

- ✅ 设计 **MotionFactBuilder**
- ✅ 或者设计 **Hybrid Delete（Legacy execute + Weaver history）**
- ✅ 或者写一段 **“为什么 Vim 必须有 op+motion” 的理论解释**

你现在这一步，非常值。