# `WEAVER_CONSTITUTION.md`

> **Status:** Ratified  
> **Scope:** Weaver Core (Kernel, Grammar, Intent, Resolver, Execution)  
> **Lasting Authority:** This document supersedes design discussions, PR descriptions, and implementation convenience.

---

## 0. 宪章目的（Purpose）

Weaver Core 的存在目的只有一个：

> **构建一个世界无关、结构优先、意图驱动的交互内核。**

本宪章用于防止以下风险：

- 架构因“方便”“性能”“主流编辑器需求”而退化  
- 核心语义被隐式状态、模式或 UI 假设污染  
- 系统逐步滑回传统编辑器模型（Vim / Emacs / IDE 内核）

**任何违反本宪章的变更，均视为 Design Regression。**

---

## 1. 不可变设计公理（Non‑Negotiable Axioms）

以下五条公理 **不可削弱、不可绕过、不可妥协**。

---

### 公理一：Intent 是唯一的语义事实来源  
**Intent Is the Single Source of Truth**

- 系统中不存在隐含意图、推断意图或默认意图  
- 所有行为必须由显式 Intent 驱动  
- 不允许根据 mode、历史状态或 UI 状态推断意图

✅ 合法：
```
Intent → Resolver → Execution
```

❌ 非法：
```
State → Guess → Action
```

---

### 公理二：Grammar 只能解析语法，不能触及语义  
**Grammar Is Purely Syntactic**

- Grammar 只能处理 token / symbol / FSM 状态  
- Grammar 不得访问任何世界状态（文档、光标、结构）  
- Grammar 不得因为“目标不存在”“当前不合法”而失败

Grammar 的唯一职责是：**生成 Intent**。

---

### 公理三：Resolver 与 Execution 职责正交  
**Resolution and Execution Are Orthogonal**

- Resolver：Intent → 结构对象  
- Execution：结构对象 → 变更集合（ChangeSet）

严格禁止：

- Execution 重新解析目标  
- Resolver 产生副作用  
- 两者互相调用

---

### 公理四：结构是本体，位置只是条件  
**Structure Is Primary, Position Is Incidental**

- Intent 中不得出现 offset / range / line / column  
- 所有坐标仅允许存在于 Resolver 内部  
- Execution 只能操作结构对象

系统不认识“字符范围”，  
只认识 **可命名、可推理的结构实体**。

---

### 公理五：核心系统必须世界无关  
**World‑Agnostic by Construction**

- Weaver Core 不假设“文本”“编辑器”“光标”  
- 不为任何具体应用场景（VSCode / Vim / IDE）让步  
- 世界模型只能通过适配层注入

Kernel ≠ Editor  
Kernel ≠ Tool  
Kernel = **语义变换引擎**

---

## 2. Weaver Core 永远不会做的事情

以下行为 **永久禁止**，无论收益多大：

- ❌ 引入 mode / state machine 作为行为来源  
- ❌ 在 Kernel 中加入文本特化优化  
- ❌ 允许 Grammar 访问世界模型  
- ❌ 允许 Execution 依赖 UI / Cursor  
- ❌ 因“用户习惯”破坏结构抽象

---

## 3. 合宪性判定规则（Constitutional Review）

### 任何 PR / RFC / Feature，必须回答：

1. 是否完全由 Intent 驱动？
2. Grammar 是否保持纯语法？
3. Resolver / Execution 是否严格分离？
4. 是否以结构而非位置为核心？
5. 是否不引入任何世界假设？

**只要有一个问题回答为「否」 → 拒绝合并。**

---

## 4. 破坏性变更（Breaking Changes）

允许 Breaking Change 的 **唯一理由**：

> **为了更严格地符合本宪章。**

以下理由一律无效：

- 性能优化  
- 用户熟悉度  
- 编辑器兼容性  
- 实现复杂度

---

## 5. 权威性声明（Authority）

- 本宪章高于：
  - README
  - 设计文档
  - Issue / PR 讨论
  - 个人意见（包括作者本人）

- 当实现与宪章冲突时：
  > **实现必须修改，宪章不可修改。**

---

## 6. 附录 A：执行性文档

以下文件 **必须** 与本宪章保持一致：

- `DESIGN_CODE_REVIEW_CHECKLIST.md`
- PR Template
- CI / Lint 规则
- 协议与接口规范

如存在冲突，以 **本宪章为准**。

---

## 7. 最终条款

> **Weaver Core 的价值不在于它能做什么，  
而在于它拒绝做什么。**

本宪章一经采纳，即视为长期有效。

---


这套东西，已经值得被“保护”了。