# Design Worldview  
**A Structured Interaction Kernel**

---

## 1. 设计对象（What This System Is）

本系统是一个**通用结构化交互内核（Structured Interaction Kernel）**。

它的职责是：

> 将**离散的用户输入序列**解析为**明确的语义意图（Intent）**，  
> 并将这些意图应用为对某个**结构化世界（Document / Model）**的**确定性变换**。

系统本身**不假设任何具体世界形态**，例如：
- 文本缓冲区
- 抽象语法树
- 图层树
- 时间轴
- 任何线性或非线性数据结构

这些都属于**外部语义注入**的范畴。

---

## 2. 核心抽象（Fundamental Abstractions）

系统只承认以下三个一等抽象：

1. **Grammar FSM**
2. **Intent**
3. **Resolver / Execution**

### 2.1 Grammar FSM

Grammar FSM 的职责是：

- 接收原始输入符号流（如按键）
- 解析其**语法结构**
- 在语法完备时生成一个 Intent

FSM **不拥有、也不访问**以下任何概念：

- 文档结构
- 光标位置
- 屏幕状态
- 编辑副作用

FSM 是一个**纯语法解析器**。

---

### 2.2 Intent

Intent 是系统内部唯一的语义载体。

一个 Intent 表示：

> “用户希望在当前上下文中，对某类目标执行某种操作。”

Intent 的特性：

- 不包含执行步骤
- 不包含空间坐标
- 不包含中间状态
- 不描述“如何到达目标”

Intent 是**声明式的**，而非指令式的。

---

### 2.3 Resolver 与 Execution

Resolver 的职责是：

- 在给定上下文中
- 将 Intent 中的 Target
- 映射为**具体的结构对象或范围**

Execution 的职责是：

- 接收 Resolver 的结果
- 应用一次**原子性的结构变换**

二者共同构成 Intent 的语义实现层。

---

## 3. 关注点正交分离（Orthogonal Separation of Concerns）

系统在设计上强制以下正交分离：

| 层级 | 关注点 |
|----|----|
| Grammar | 输入语法 |
| Intent | 语义意图 |
| Resolver | 语义映射 |
| Execution | 结构变换 |

任何一层：

- 不得反向依赖上一层的实现细节
- 不得通过隐式状态影响其他层

这是一条**不可破坏的设计约束**。

---

## 4. 结构优先原则（Structure-First Principle）

本系统采用**结构优先**的世界观：

- 所有可操作对象，必须是**可被命名和识别的结构**
- 空间位置、索引、字节偏移等，仅作为 Resolver 的输入条件之一
- 系统核心从不将“位置”视为事实来源

因此：

> “操作结构”是基本行为，  
> “操作字符或坐标”只是结构的一种实现方式。

---

## 5. 声明式执行模型（Declarative Execution Model）

系统执行模型遵循以下原则：

1. **Intent 是事实来源**
2. Execution 不重放输入历史
3. Execution 不模拟用户行为
4. 相同 Intent + 相同上下文 → 相同结果

系统不保证任何“感觉像某个工具”的行为一致性，只保证**语义一致性**。

---

## 6. 世界无关性（World-Agnostic Design）

系统不对以下概念做任何假设：

- 文本线性性
- 光标唯一性
- 单一视口
- 同步输入流

因此：

- 多光标是自然成立的
- 非线性结构是可操作的
- 并行 Intent 是允许的

这些能力不是扩展特性，而是设计结果。

---

## 7. 扩展方式（Extensibility Model）

系统的扩展边界定义为：

- **新增 Intent 类型**
- **新增 Target 定义**
- **实现新的 Resolver / Execution**

系统**不提供**：

- 行为级脚本注入
- 隐式钩子
- 状态篡改接口

扩展的是**模型本身**，而不是执行路径。

---

## 8. 非目标（Non-Goals）

以下内容**明确不属于**本系统的设计目标：

- 兼容任何既有编辑器行为
- 复刻某种用户肌肉记忆
- 提供完整的 UI 或交互体验
- 成为终端或 GUI 应用

这些能力可以构建在本系统之上，但不反向约束其设计。

---

## 9. 可测试性与形式化倾向

系统被设计为：

- Grammar 可被纯函数测试
- Resolver 可被确定性验证
- Execution 可被原子性断言

系统结构天然支持：
- 属性测试（Property-based Testing）
- 形式化推理
- Headless 执行环境

---

## 10. 设计立场总结

本系统的立场可以总结为：

> 我们不实现“编辑器行为”，  
> 我们实现“结构化意图的执行语义”。

这是一个**模型级系统**，而非工具级系统。



---

# Five Non‑Negotiable Design Axioms

---

## 公理一：Intent 是唯一的语义事实来源  
**(Intent as the Single Source of Truth)**

### 公理表述

> 系统中，用户意图只能以 **Intent** 的形式存在。  
> 任何语义判断、执行决策、结构变换，必须以 Intent 为输入。

不存在“隐含意图”“当前模式”“最近一次操作”这样的事实来源。

---

### ❌ 违反示例（Anti‑patterns）

- Execution 根据「当前光标状态」决定行为  
- Resolver 读取「最近一次命令」来补全语义  
- Grammar 根据文档内容改变解析结果  
- 通过 mode / flag / global state 推断用户想做什么  

**典型危险代码气味：**

```ts
if (editor.mode === "visual") {
  // 改变执行语义
}
```

---

### ✅ API 约束（Contracts）

```ts
interface Intent {
  type: IntentType
  target: Target
  operator: Operator
  qualifiers?: Qualifier[]
}
```

**强制规则：**

- Execution API 只接受 Intent + Context
- Context 不包含「意图推断信息」
- Intent 必须是完整、自描述的

---

## 公理二：Grammar 只能解析语法，不能触及语义  
**(Grammar Is Purely Syntactic)**

---

### 公理表述

> Grammar FSM 只负责将输入符号解析为 Intent。  
> 它不允许访问任何世界状态或结构信息。

Grammar 是一个**纯语法系统**。

---

### ❌ 违反示例（Anti‑patterns）

- Grammar 根据文档类型切换解析规则  
- Grammar 查询光标位置决定 motion 是否有效  
- Grammar 阻止某个命令，因为“当前不适合”  

```ts
// ❌ Grammar 不应知道 target 是否存在
if (!document.hasParagraphAfter(cursor)) reject()
```

---

### ✅ API 约束（Contracts）

```ts
interface GrammarFSM {
  feed(input: Token): FSMState
  maybeEmitIntent(): Intent | null
}
```

**强制规则：**

- Grammar 不接收 Context
- Grammar 不抛出语义错误
- 所有“是否可执行”的判断延迟到 Resolver

---

## 公理三：Resolver 决定“作用对象”，Execution 决定“如何改变”  
**(Resolution and Execution Are Orthogonal)**

---

### 公理表述

> Resolver 负责将 Intent 的 Target 映射为结构对象。  
> Execution 只负责对已解析对象执行变换。

二者职责严格分离。

---

### ❌ 违反示例（Anti‑patterns）

- Execution 内部重新查找目标  
- Resolver 执行实际修改  
- Execution 根据结构内容改变解析策略  

```ts
// ❌ Execution 不应再定位对象
const node = findNodeByCursor(...)
```

---

### ✅ API 约束（Contracts）

```ts
interface Resolver {
  resolve(intent: Intent, ctx: Context): ResolvedTarget
}

interface Execution {
  apply(resolved: ResolvedTarget, intent: Intent): ChangeSet
}
```

**强制规则：**

- Execution 不访问 Context
- Resolver 不产生副作用
- ChangeSet 必须是原子性的

---

## 公理四：系统以“结构”为本体，而非位置或字符  
**(Structure Is Primary, Position Is Incidental)**

---

### 公理表述

> 所有可操作对象必须是**可命名的结构实体**。  
> 坐标、偏移、行列号仅是 Resolver 的输入条件。

系统核心永远不以“位置”为事实。

---

### ❌ 违反示例（Anti‑patterns）

- Intent 中包含 byte offset  
- Execution 接受 `(start, end)` 作为核心参数  
- API 暴露「当前行」「当前列」  

```ts
deleteRange(startOffset, endOffset) // ❌
```

---

### ✅ API 约束（Contracts）

```ts
interface Target {
  kind: TargetKind
  selector?: Selector
}
```

```ts
interface Selector {
  resolve(context: Context): StructuralObject[]
}
```

**强制规则：**

- Intent 不包含坐标
- 坐标只能存在于 Resolver 内部
- Execution 只处理结构对象

---

## 公理五：系统是世界无关的，不为任何具体应用让步  
**(World‑Agnostic by Construction)**

---

### 公理表述

> 系统不对文本、编辑器、光标、视口等概念做任何假设。  
> 任何具体世界模型只能通过 Resolver / Execution 注入。

世界模型永远在系统之外。

---

### ❌ 违反示例（Anti‑patterns）

- Kernel 默认假设线性文本  
- API 命名直接使用 `line`, `column`, `buffer`  
- 为“文本编辑效率”破坏抽象  

```ts
class EditorKernel { ... } // ❌ 命名即世界假设
```

---

### ✅ API 约束（Contracts）

```ts
interface WorldModel {
  resolveTarget(target: Target): StructuralObject[]
  applyChange(change: ChangeSet): void
}
```

**强制规则：**

- Kernel 不依赖具体 WorldModel
- WorldModel 不影响 Grammar / Intent 结构
- 所有假设必须在适配层显式声明

---

# 总结：五条公理，一条红线

| 公理 | 破坏后果 |
|----|----|
| Intent 唯一性 | 语义不可推理 |
| Grammar 纯粹性 | 行为不可复现 |
| Resolver / Execution 分离 | 系统失控 |
| 结构优先 | 抽象崩塌 |
| 世界无关 | 系统退化为工具 |

**任何 PR、RFC、Feature，只要违反其中任意一条，都是设计回归（Design Regression）。**

---


- `docs/review/CHECKLIST.md`
- PR 模板的 *Required Review Items*
- 架构评审（ADR / RFC）的强制附录

---

# Design Code Review Checklist  
**Structured Interaction Kernel**

> 评审目标：确保任何变更 **不违反五条设计公理**。

评审时，逐条回答 **Yes / No**。  
任何 **No** 都必须阻止合并。

---

## ✅ 公理一：Intent 是唯一语义事实来源

### 检查项

- [ ] 是否 **新增或修改** 的行为 **完全由 Intent 驱动**
- [ ] 是否存在 **根据全局状态 / mode / flag** 推断用户意图的逻辑
- [ ] 是否有 Execution / Resolver **绕过 Intent** 直接做决策
- [ ] Intent 是否是 **完整、自描述、不可歧义的**

### ❌ 一票否决信号

- 出现 `currentMode` / `lastCommand` / `implicitState`
- Intent 结构中出现“可选但实际上必需”的字段
- 执行路径中出现“如果用户可能是想要……”

---

## ✅ 公理二：Grammar 只能解析语法，不能触及语义

### 检查项

- [ ] Grammar 是否 **完全不依赖 Context / World**
- [ ] Grammar 是否只处理 token / symbol，而非结构对象
- [ ] Grammar 是否 **始终可以生成 Intent**，而不是提前失败
- [ ] 是否所有“是否合法 / 是否存在”的判断都发生在 Resolver

### ❌ 一票否决信号

- Grammar 中出现 `document`, `cursor`, `buffer`
- Grammar 抛出 “目标不存在”“当前位置无效” 类错误
- Grammar 根据世界状态改变解析路径

---

## ✅ 公理三：Resolver 与 Execution 职责严格分离

### 检查项

- [ ] Resolver 是否 **只做映射，不做修改**
- [ ] Execution 是否 **只对 ResolvedTarget 操作**
- [ ] Execution 是否 **完全不访问 Context**
- [ ] Resolver 是否无副作用、可重复调用

### ❌ 一票否决信号

- Execution 中重新定位目标
- Resolver 中出现 mutate / apply / commit
- Execution 根据结构内容改变策略

---

## ✅ 公理四：系统以结构为本体，而非位置或字符

### 检查项

- [ ] Intent 是否 **不包含任何坐标、索引、偏移**
- [ ] Execution API 是否 **只接受结构对象**
- [ ] 所有 position / offset 是否仅存在于 Resolver 内部
- [ ] 新增 Target 是否是**可命名结构**而非范围描述

### ❌ 一票否决信号

- API 参数出现 `(start, end)`
- 使用 `line`, `column`, `offset` 作为核心概念
- 执行逻辑基于“当前位置”

---

## ✅ 公理五：系统必须保持世界无关性

### 检查项

- [ ] Kernel 是否未引入任何具体世界假设
- [ ] 新代码是否可在 **非文本世界** 下成立
- [ ] 模块命名是否中立（无 editor / buffer / cursor）
- [ ] 世界特有逻辑是否全部封装在适配层

### ❌ 一票否决信号

- Kernel 中出现 “文本特化优化”
- 接口语义暗含线性文本假设
- 为某一应用场景破坏抽象

---

## ✅ 横向一致性检查（必须全部通过）

- [ ] Intent → Resolver → Execution 链路是否完整
- [ ] 任一层是否存在跨层访问
- [ ] 是否可以 **移除 UI / 世界模型** 后仍能运行核心逻辑
- [ ] 是否可以在 Headless 环境下测试

---

## ✅ 测试与验证要求

- [ ] 新增 Grammar 是否有纯函数测试
- [ ] Resolver 是否具备确定性测试
- [ ] Execution 是否有原子性断言
- [ ] 是否新增违反公理的 Anti‑pattern 测试用例

---

# 评审结论规则（必须执行）

- ✅ **全部通过** → 可合并
- ❌ **任意一项失败** → 标记为 *Design Regression*，禁止合并
- ⚠️ **无法判断** → 要求作者补充设计说明

---

## 评审者最后一问（强制）

> **“这段代码是否在任何地方偷偷假设了一个编辑器世界？”**

如果答案不是**明确的“没有”**，  
那么这次评审必须拒绝。

---
