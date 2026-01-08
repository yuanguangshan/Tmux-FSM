
---

# 📜 Weaver Core Constitution  
## *RFC‑WC‑000: Fact‑Oriented Headless Editing Kernel*

**Status:** Normative  
**Category:** Core Architecture / Kernel Governance  
**Applies To:** tmux‑fsm, Weaver Core, all compatible implementations  
**Version:** 1.0  
**Date:** 2026‑01‑02

---

## 0. 本文档的地位（Status of This Document）

本文档定义 **Weaver Core 编辑内核的宪法级约束**。

- 本文档中使用的 **MUST / MUST NOT / SHOULD / MAY**  
  其含义遵循 RFC 2119。
- 任何声称“兼容 Weaver Core”的实现，**必须完全遵守本文件中的 Normative 条款**。
- 本文档不定义 UI、快捷键或具体编辑器行为。

---

## 1. 术语定义（Terminology）

### 1.1 Kernel（内核）

长期运行、持有编辑语义真值的进程实体。

### 1.2 Client

无状态输入代理，仅负责将用户输入转发给 Kernel。

### 1.3 Executor

在 Kernel 裁决后，对具体文本表面执行修改的组件。

### 1.4 Fact

对一次已发生编辑语义的不可变描述。

### 1.5 Intent

用户或前端表达的编辑意图，不保证可执行。

### 1.6 Verdict

Kernel 对 Intent 的裁决结果。

### 1.7 Audit

对 Verdict 的解释性记录，用于审计与追责。

---

## 2. 基本公理（Axioms）【Normative】

### Axiom 1 — Kernel Sovereignty

- Kernel **MUST** 是系统中唯一的语义真值来源。
- Client 与 Executor **MUST NOT** 持有或推断语义主权。

---

### Axiom 2 — Fact Primacy

- 所有编辑行为 **MUST** 被建模为 Fact。
- 系统 **MUST NOT** 依赖按键序列重放来恢复编辑语义。

---

### Axiom 3 — Safety First

- 当编辑或撤销存在不确定性时，Kernel **MUST** 拒绝执行。
- 正确性 **MUST** 优先于用户便利性。

---

### Axiom 4 — Auditability

- 每一个 Verdict **MUST** 具备可查询的 Audit。
- “无法解释的编辑裁决”属于 **Undefined Behavior**。

---

## 3. 架构主权模型（Governance Model）

### 3.1 权限划分【Normative】

| 组件 | 权限 |
|----|----|
| Kernel | 意图解析、事实生成、裁决 |
| Client | 输入转发 |
| Executor | 执行裁决 |

- Client **MUST NOT** 修改文本。
- Executor **MUST NOT** 推翻或修改 Verdict。
- Kernel **MUST NOT** 依赖 UI 状态。

---

## 4. Intent / Verdict / Audit ABI

### 4.1 Intent【Normative】

Intent：

- **MUST** 表达“想要做什么”
- **MUST NOT** 包含“如何执行”
- **MAY** 被拒绝

---

### 4.2 Verdict【Normative】

Verdict 类型：

- `ACCEPT`
- `REJECT`
- `DEFER`

Verdict：

- **MUST** 单向、不可回滚
- **MUST** 关联 Audit
- **MUST NOT** 被 Client 或 Executor 修改

---

### 4.3 Audit【Normative】

Audit：

- **MUST** 不可变
- **MUST** 可查询
- **MUST NOT** 与 UI 生命周期绑定

---

## 5. Fact 规范（Fact Specification）

### 5.1 不可变性

- Fact **MUST** 是不可变的。
- Fact **MUST NOT** 被就地修改。

---

### 5.2 坐标独立性

- Fact **MUST NOT** 直接引用光标坐标。
- Fact **MUST** 绑定 Anchor。

---

### 5.3 时间属性

- Fact **MUST** 按提交顺序线性排列。
- Fact **MUST NOT** 被重排序。

---

## 6. Undo / Redo 法则

### 6.1 Undo 语义【Normative】

- Undo **MUST** 生成新的 Fact。
- Undo **MUST NOT** 删除历史 Fact。

---

### 6.2 Anchor Resolver

Resolver **MUST** 按以下顺序尝试：

1. 精确匹配（Exact）
2. 模糊匹配（Fuzzy）
3. 拒绝（Fail）

---

### 6.3 Fuzzy 行为约束

- Fuzzy Undo **MUST** 显式告知用户。
- Fuzzy Undo **MUST** 降级信任等级。

---

## 7. 事务模型（Transactions）

- 复合编辑 **MUST** 被视为原子事务。
- 任一 Fact 定位失败：
  - 整个事务 **MUST** 被拒绝
  - Redo **MUST NOT** 可用

---

## 8. Executor 契约（Executor Contract）

Executor：

- **MUST** 严格执行 Verdict
- **MUST NOT** 修改 Fact
- **MUST NOT** 执行未裁决编辑

---

## 9. Undefined Behavior（违宪行为）

以下行为属于 **违宪实现**：

- Client 直接修改文本
- Executor 私自回滚
- 未生成 Audit 即执行编辑
- UI 状态被视为真值来源

---

## 10. Informative Appendix（非规范性）

### A. 设计哲学摘要（非规范）

> 编辑不是 UI 行为，而是语义事件。  
> Kernel 的职责不是“尽量满足用户”，而是在不确定环境中维护语义尊严。

---

## 11. Compatibility Statement

任何实现若：

- 完全遵守第 2–9 章  
- 不引入新的语义主权源  

则可声明为：

> **Weaver Core Compatible**

---


---

# 📜 RFC‑WC‑001  
## *Anchor Model & Resolver Specification*

**Status:** Normative  
**Category:** Core Semantic Model  
**Version:** 1.0  
**Date:** 2026‑01‑02

---

## 1. Scope

本文档定义 **Weaver Core 中 Anchor 的语义模型与解析规则**。

Anchor 是 **Fact 得以脱离光标、UI 与 Pane 存在的根本机制**。

---

## 2. Anchor 定义

### 2.1 Anchor（锚点）

Anchor 是一种 **逻辑定位描述**，用于在动态文本表面中定位编辑语义目标。

Anchor **不是**：

- 光标坐标
- 行号
- 偏移量快照

---

### 2.2 Anchor 的必要性【Normative】

- 所有 Fact **MUST** 绑定 Anchor
- 无 Anchor 的 Fact 属于 **Invalid Fact**

---

## 3. Anchor 组成（Anchor Components）

一个 Anchor **MUST** 由以下部分组成：

### 3.1 Semantic Token

- 描述目标文本的 **语义特征**
- **MUST** 独立于具体位置

示例（非规范）：
- 命令名
- 函数签名片段
- Shell Prompt 结构

---

### 3.2 Context Window

- Anchor 周围的上下文摘要
- **MUST** 有限长度
- **MUST NOT** 包含全量文本

---

### 3.3 Temporal Hint

- Anchor 创建时的时间顺序信息
- **MAY** 用于消歧
- **MUST NOT** 单独作为定位依据

---

## 4. Anchor Resolver

### 4.1 Resolver 职责【Normative】

Resolver：

- **MUST** 将 Anchor 映射为具体编辑范围
- **MUST NOT** 产生副作用
- **MUST** 返回 SafetyLevel

---

### 4.2 解析等级（SafetyLevel）

Resolver **MUST** 返回以下之一：

| Level | 含义 |
|----|----|
| EXACT | 唯一、无歧义匹配 |
| FUZZY | 存在不确定性 |
| FAIL | 无法安全定位 |

---

### 4.3 解析顺序【Normative】

Resolver **MUST** 严格按以下顺序执行：

1. EXACT 匹配
2. FUZZY 搜索
3. FAIL

**MUST NOT** 跳过阶段。

---

## 5. 拒绝条件

Resolver **MUST** 返回 FAIL 当：

- 匹配结果多于 1 且不可判别
- 上下文漂移超出阈值
- 文本表面发生不可逆破坏

---

## 6. 安全原则

- Anchor Resolver **MUST** 偏向拒绝
- 错误匹配属于 **违宪行为**

---

# 📜 RFC‑WC‑002  
## *FSM & Intent Grammar*

**Status:** Normative  
**Category:** Kernel Control Logic  
**Version:** 1.0

---

## 1. Scope

定义：

- Weaver Core 的 **有限状态机（FSM）**
- Intent 的 **语法与生命周期**

---

## 2. FSM 总览

Kernel **MUST** 实现以下最小状态集：

```
IDLE
│
├─► EVALUATING
│     ├─► ACCEPTED
│     ├─► REJECTED
│     └─► DEFERRED
│
└─► ERROR
```

---

## 3. 状态约束【Normative】

- 任一 Intent **MUST** 经由 `EVALUATING`
- Verdict **MUST** 在终态产生
- FSM **MUST NOT** 回退到历史状态

---

## 4. Intent Grammar

### 4.1 Intent 基本结构

Intent **MUST** 包含：

- IntentType
- Target Descriptor
- Optional Constraints

---

### 4.2 Intent 的不确定性

- Intent **MAY** 表达模糊目标
- Kernel **MUST NOT** 假设 Intent 可执行

---

## 5. FSM 与安全交互

- Anchor 解析失败 → FSM **MUST** 转入 REJECTED
- Audit 生成失败 → FSM **MUST** 转入 ERROR

---

# 📜 RFC‑WC‑003  
## *Audit & Failure Taxonomy*

**Status:** Normative  
**Category:** Audit / Trust Model  
**Version:** 1.0

---

## 1. Scope

定义：

- Audit 的结构
- Failure 的分类体系

---

## 2. Audit 结构【Normative】

Audit **MUST** 包含：

- Verdict
- SafetyLevel
- Resolver Outcome
- Failure Code（若有）

---

## 3. Failure 分类

### 3.1 Failure Class

| Class | 描述 |
|----|----|
| INTENT | 意图不成立 |
| ANCHOR | 定位失败 |
| ENV | 环境破坏 |
| INTERNAL | 内核错误 |

---

### 3.2 Failure 的不可抹除性

- Failure **MUST** 被记录
- **MUST NOT** 被静默吞掉

---

## 4. 用户可见性

- 所有 REJECT **MUST** 可解释
- 无解释拒绝属于 **违宪行为**

---

# 📜 RFC‑WC‑004  
## *Spatial Echo Semantics*

**Status:** Normative  
**Category:** Cross‑Surface Semantics  
**Version:** 1.0

---

## 1. Scope

定义 **编辑语义在不同空间表面中的回声行为**。

---

## 2. Spatial Echo 定义

Spatial Echo 指：

> 同一 Fact 在不同 Pane / Surface 中的语义一致性表现。

---

## 3. Echo 原则【Normative】

- Echo **MUST** 基于 Fact，而非 UI
- Echo **MUST NOT** 假设空间连续性

---

## 4. Echo 失败处理

- 任一空间解析失败：
  - 整个 Echo **MUST** 降级或拒绝
- 不允许部分成功但不审计

---

## 5. 非目标（Non‑Goals）

- 不保证视觉同步
- 不保证实时性
- 不保证用户感知一致

---

 
**RFC‑WC‑005 是“负宪法”**——它不告诉实现者 *该做什么*，而是明确写死 **绝对不能做什么**。  
这正是内核级规范真正成熟的标志。


---

# 📜 RFC‑WC‑005  
## *Non‑Goals & Explicit Rejections*

**Status:** Normative  
**Category:** Constitutional Constraints  
**Version:** 1.0  
**Date:** 2026‑01‑02

---

## 0. Purpose

本文档定义 **Weaver Core 明确不追求的目标（Non‑Goals）**，  
以及 **任何声称兼容的实现必须拒绝的行为（Explicit Rejections）**。

> **未在本文件中明确拒绝的行为，不自动视为允许。**

---

## 1. 非目标原则（Non‑Goal Principle）

- Weaver Core **不是** 一个 UI 框架  
- Weaver Core **不是** 一个“尽量帮用户完成事情”的系统  
- Weaver Core **不是** 一个宽容失败的编辑器内核  

安全性、可解释性与语义尊严 **优先于成功率与体验流畅度**。

---

## 2. UI 与交互层非目标

以下能力 **明确不属于 Weaver Core 的目标**：

### 2.1 视觉一致性

- 不保证多 Pane 之间的像素同步
- 不保证光标位置一致
- 不保证屏幕刷新顺序

---

### 2.2 即时反馈

- 不保证毫秒级响应
- 不保证输入与编辑之间的实时耦合

---

### 2.3 用户感知连续性

- 不保证 Undo 行为“看起来像传统编辑器”
- 不保证 Redo 可用

---

## 3. 编辑成功率非目标

### 3.1 成功不是目标【Normative】

- Kernel **MUST NOT** 以“尽量成功”为目标
- Kernel **MUST** 以“避免错误”为目标

---

### 3.2 静默失败的拒绝

- Kernel **MUST NOT**：
  - 猜测用户真实意图
  - 自动选择多个可能目标之一
  - 在不确定时“帮用户试试”

---

## 4. 状态便利性非目标

以下行为 **被明确拒绝**：

### 4.1 UI 状态真值化

- 将光标位置视为语义依据
- 将 Pane 可见性作为编辑合法性条件
- 将焦点状态作为 Anchor 辅助判断

---

### 4.2 快捷键驱动语义

- 依据按键序列恢复编辑语义
- 允许 Executor 推断语义意图

---

## 5. 宽松 Undo / Redo 的拒绝

### 5.1 非确定性 Undo【Explicit Rejection】

- 不允许“可能撤销到正确位置”的 Undo
- 不允许模糊撤销而不告知用户

---

### 5.2 历史篡改的拒绝

- 不允许删除或合并历史 Fact
- 不允许为“看起来更合理”而重写历史

---

## 6. 隐式语义扩展的拒绝

### 6.1 魔法行为【Explicit Rejection】

- 不允许：
  - 自动补全编辑语义
  - 隐式扩大编辑范围
  - 基于统计或频率的语义推断

---

### 6.2 学习型内核的拒绝

- Kernel **MUST NOT**：
  - 学习用户行为
  - 调整裁决策略
  - 引入概率模型

> Weaver Core 是 **裁判**，不是 **助手**。

---

## 7. 审计逃逸的拒绝

以下行为构成 **违宪实现**：

- 允许无 Audit 的 ACCEPT
- 将错误折叠为成功
- 使用日志代替 Audit

---

## 8. 性能优先级的拒绝

### 8.1 性能不是第一目标

- 不允许为性能牺牲 SafetyLevel
- 不允许为减少延迟跳过 Resolver 阶段

---

### 8.2 并发投机的拒绝

- 不允许 speculative execution
- 不允许回滚型优化

---

## 9. 可扩展性边界

### 9.1 插件能力限制【Normative】

- 插件 **MUST NOT**：
  - 创建 Fact
  - 修改 Anchor
  - 干预 Resolver

---

### 9.2 API 主权保护

- Kernel API **MUST** 是最小集
- 不提供“高级捷径接口”

---

## 10. 明确拒绝清单（Summary）

以下行为 **一律视为不兼容**：

- UI 驱动语义
- 模糊但不告知的编辑
- 静默修复失败
- 成功率导向裁决
- 无审计编辑
- 历史可变

---

## 11. Compatibility Impact

任何实现若：

- 违反本文件任一条款  
- 或以“用户体验”为由绕过拒绝条款  

则 **不得** 声称：

> Weaver Core Compatible

---

## 12. Informative Appendix（非规范）

> *一个拒绝说“不”的系统，  
> 最终只能对错误负责。*

Weaver Core 的价值在于：  
**它在不确定时，敢于什么都不做。**

---

## ✅ 到这里为止，你已经完成了什么

你现在拥有的是：

- 一套 **正宪法（000）**
- 四份 **操作性技术法典（001–004）**
- 一份 **负宪法（005）**

这已经不是“编辑器设计”，而是：

> **一种可审计、可拒绝、可追责的编辑制度。**

---

