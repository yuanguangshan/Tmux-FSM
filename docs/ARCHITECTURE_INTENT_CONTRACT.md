# Intent / Resolver 架构公约

> **这是“宪法”，不是建议**

## ✅ Intent 公约（Intent Contract）

### 1️⃣ Intent 是**纯意图（Pure Intent）**

Intent **必须**：

- ✅ 表达 **“用户想做什么”**
- ✅ 使用 **语义目标（Semantic Target）**
- ✅ 与输入设备（键盘/鼠标）无关

Intent **禁止**：

- ❌ 包含 row / col / offset
- ❌ 读取 snapshot / buffer 内容
- ❌ 携带可执行逻辑

---

### 2️⃣ Intent 必须是**可序列化、可回放的**

Intent 必须满足：

- ✅ JSON 序列化后语义不变
- ✅ 在不同机器上 Resolve 结果一致（给定相同 snapshot）
- ✅ 不依赖全局状态

> **理由**：Undo / Redo / Replay / 协同编辑

---

### 3️⃣ Intent 只允许由 IntentBuilder 构造

```go
// Forbidden:
Intent{ Kind: IntentDelete }

// Allowed:
builder.Delete(target, count)
```

✅ Code Review 规则：  
**任何直接 new Intent 的代码直接拒绝**

---

### 4️⃣ Intent 不允许携带 Legacy 标记

```go
// Forbidden:
Intent{ Kind: IntentMove, LegacyLineID: "..." }
```

✅ Legacy **只存在于 Resolver 输入兼容层**

---

## ✅ Resolver 公约（Resolver Contract）

---

### 5️⃣ Resolver 是**唯一允许读取 Snapshot 的层**

Resolver **必须**：

- ✅ 从 Intent → Anchor / Range
- ✅ 读取 snapshot / rope / buffer
- ✅ 处理 Unicode / grapheme / wrap

FSM / Builder **禁止**触碰以上内容。

---

### 6️⃣ Resolver 输出必须是**执行级结构**

Resolver 输出：

```go
type ResolvedIntent struct {
    Anchors []Anchor
    Ranges  []Range
}
```

✅ 输出 **不允许再携带语义歧义**

---

### 7️⃣ Resolver 对 Legacy 的态度是“**清洗，不传播**”

Resolver **允许**：

- ✅ 接受 legacy intent
- ✅ 解析 legacy anchor

Resolver **禁止**：

- ❌ 生成新的 legacy anchor
- ❌ 把 legacy 标记传播到执行层

---

### 8️⃣ Resolver 必须支持“严格模式”

```go
StrictNativeResolver = true
```

在严格模式下：

- ✅ 任意 legacy 泄漏 → panic
- ✅ 任意未解析 semantic target → panic

---

### 9️⃣ Resolver 不允许“部分成功”

Resolver 要么：

- ✅ 完整 Resolve
- ❌ 返回 error / panic

❌ 禁止 silent fallback

---

## ✅ FSM 公约（FSM Contract）

（这是 Phase 2 生效的）

---

### 🔟 FSM 永远不关心“如何执行”

FSM **只做三件事**：

1. 解析按键序列
2. 管理 PendingOp / Count / Motion
3. 生成 Intent

---

### 1️⃣1️⃣ FSM 是**纯状态机**

FSM **必须**：

- ✅ 可快照
- ✅ 可回放
- ✅ 无副作用

---

### 1️⃣2️⃣ FSM 永远不知道 Legacy 是否存在

FSM：

- ❌ 不调用 legacy bridge
- ❌ 不产生 legacy 字符串
- ❌ 不知道 AnchorOriginLegacy

---

> ✅ **这 12 条，就是你整个编辑器架构的“物理定律”**
