## 下面这份文本不是说明、不是愿景、不是设计笔记。  
它是**约束未来一切实现的法律文本**。

任何代码、PR、重构、实验，**只要违反本宪章，即视为架构缺陷**。

---

# tmux-fsm 架构宪章（v1）

**状态**：生效  
**适用范围**：`kernel / weaver / fsm / resolver / projection / history / ui`  
**非适用范围**：临时调试工具、一次性迁移脚本（需显式标注）

---

## 序言（Preamble）

`tmux-fsm` 不是一个快捷键增强器，也不是一个脚本集合。

它是一个**语义执行内核**，其职责是在**不可信、异步、可漂移的终端现实**中，  
对用户或系统提出的**编辑意图（Intent）**进行：

> **解析 → 校验 → 重基 → 投影 → 审计**

本宪章定义了该系统**赖以存在的不可变约束**。

---

## 第一章：核心不变量（Invariants）

### §1.1 语义先于物理（Semantic > Physical）

- 任何用户行为，必须首先被建模为 **Intent**
- 按键序列、命令字符串、tmux API 均不构成系统真理

**禁止**：
- 任何模块直接基于物理输入推导逻辑语义

---

### §1.2 现实必须被校验（No Blind Projection）

- 任何 Intent 在执行前，必须与 **Snapshot** 对齐
- Snapshot 不一致不是异常，而是**条件失败**

**允许**：
- 拒绝（Reject）
- 语义重基（Semantic Rebase）

**禁止**：
- 在未验证现实前直接投影

---

### §1.3 单向因果流（Unidirectional Causality）

系统因果顺序 **必须且只能是**：

```
Intent
  → Resolve
    → Verdict
      → Projection
        → Backend
```

**禁止**：
- FSM / Resolver / Kernel 直接调用 Backend
- UI 直接产生物理副作用

---

### §1.4 失败是结果，不是异常（Failure as Result）

- 拒绝、漂移、策略否决均为合法结果
- panic / fatal 仅限不可恢复的系统错误

**要求**：
- 所有失败必须可分类、可解释、可审计

---

### §1.5 历史是事实（History as Fact）

- 每一个成功的 Intent 都必须进入事实日志
- Undo / Replay 基于历史事实，而非物理逆操作

**禁止**：
- 基于 send-keys 的“反向猜测”

---

### §1.6 物理层永远可替换（Backend Replaceability）

- tmux 是当前 Backend，不是系统前提
- 任何逻辑不得依赖 tmux 的偶然行为

---

## 第二章：模块主权与边界（Sovereignty & Boundaries）

### §2.1 Kernel

- 定义 Intent、Transaction、Verdict 的语义结构
- 不接触物理现实

**唯一职责**：语义裁决

---

### §2.2 Weaver

- 系统的**执行仲裁者**
- 唯一允许接触 Projection 的模块

**必须具备**：
- World Drift 处理
- Semantic Rebase 能力
- Transaction 生命周期管理

---

### §2.3 Resolver

- 负责将 Intent 解析为语义操作
- **系统中只能存在一个权威 Resolver 路径**

**禁止**：
- 并存多套 Resolver
- Legacy Resolver 进入主路径

---

### §2.4 Projection

- 负责将 Verdict 转换为物理 SideEffect
- 所有物理调用必须在此发生

---

### §2.5 UI

- UI 是 Verdict 的被动观察者
- UI 更新必须来源于 Transaction SideEffect

**禁止**：
- UI 直接调用 tmux / shell
- UI 改变系统状态

---

## 第三章：事务与时间（Transactions & Time）

### §3.1 事务是显式的

- 所有执行必须属于某个 Transaction
- Transaction 必须有明确的：
  - Begin
  - Commit / Abort

---

### §3.2 事务可失败、可回滚

- Abort 是正常结束路径
- 不允许“半投影”的成功

---

### §3.3 推测执行是合法的（Speculative Projection）

- 允许在内存模拟器中提前渲染结果
- 最终物理投影仍需校验与裁决

---

## 第四章：历史与撤销（History & Undo）

### §4.1 全保真事实捕获

- 历史必须包含：
  - 文本
  - Cell 属性（SGR 等）
  - 空间位置信息

---

### §4.2 Undo 是事实回放

- Undo = Replay 反向事实
- 不允许基于当前现实推导“应该怎么撤销”

---

## 第五章：伪存在与降级原则（Anti-Phantom Rules）

### §5.1 未进入主因果链的模块视为伪存在

- 未被 Kernel / Weaver 调用
- 不影响 Resolve / Verdict / Projection

**处理原则**：
- 降级 或 删除
- 不允许长期“占位”

---

### §5.2 CRDT 特别条款

- 在单作者线性执行模型下：
  - CRDT 仅可作为 History substrate
  - 不得引入并发 Epoch 语义

---

## 第六章：禁令（Hard Prohibitions）

以下行为视为**架构违规**：

1. FSM 直接调用 tmux / shell
2. UI 产生物理副作用
3. 同一 Intent 存在多条权威解释路径
4. 未经 Snapshot 校验的投影
5. Legacy 模块无限期共存

---

## 第七章：Kill-Switch 条款

任何模块必须声明：

- ✅ **存在条件**
- ❌ **删除条件**

未声明删除条件的模块，**不得进入核心路径**。

---

## 终章：宪章的地位

- 本宪章高于任何实现细节
- 本宪章可修订，但必须：
  - 明确版本号
  - 明确废止条款

> **当代码与宪章冲突时，  
> 代码是错的。**

---

### ✅ 生效声明

**tmux-fsm 架构宪章（v1）**  
自此刻起，视为系统最高约束。

