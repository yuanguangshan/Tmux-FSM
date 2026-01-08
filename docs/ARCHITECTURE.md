很好，这一步就是**把“理念”变成“制度”**。  
下面是把 **① 精简版十条戒律 + ② 一页架构图 + ③ 戒律对照表** **整合成一份「单一权威文档」** 的结果。

你可以把它直接作为：

- ✅ `ARCHITECTURE.md`（**唯一权威**）
- 或 `docs/architecture.md`
- 或仓库首页 README（如果这是核心项目）

—

# 🧠 Tmux‑FSM 架构规范（单一权威）

> **本文件是该项目的架构宪法。**  
> 任何代码、设计、PR、Review，均必须以此为最高标准。  
>  
> **违反本规范的问题属于架构缺陷，而非实现缺陷。**

—

## 一、架构目标（一句话）

> **将「按键输入」与「行为执行」彻底解耦，  
> 通过 FSM → Grammar → Kernel → Intent → Transaction 的单向管道，  
> 构建一个可推理、可重放、不可腐化的编辑系统。**

—

## 二、整体架构（一页图）

```mermaid
flowchart TD
    K[Keys<br/>按键] —> FSM[FSM<br/>输入状态机]
    FSM —> RT[RawToken]

    RT —> G[Grammar<br/>Vim 语义]
    G —> GI[GrammarIntent<br/>未裁决语义]

    GI —> KERNEL[Kernel<br/>唯一权威]

    KERNEL —> INTENT[Intent<br/>语义契约]

    INTENT —> B[Builder<br/>语义翻译]
    B —> TX[Transaction<br/>不可变操作]

    TX —> EXEC[Executor / Editor]
    EXEC —> BACKEND[Backend<br/>tmux / editor]

    %% Legacy
    INTENT -. legacy .-> RES[Resolver<br/>⚠️ 冻结]

    %% UI
    UI[UI / View] -. 派生 .- EXEC
```

—

## 三、分层心智模型

| 层级 | 名称 | 职责边界 |
|-—|-—|-—|
| Keys | 输入 | 物理按键事件 |
| FSM | 状态机 | 输入组合 → token |
| Grammar | 语义层 | Vim 语言规则 |
| Kernel | 仲裁层 | 唯一决策者 |
| Intent | 契约层 | 描述“想做什么” |
| Builder | 翻译层 | Intent → Transaction |
| Transaction | 操作层 | 可重放、不可变 |
| Executor | 执行层 | 应用 Transaction |
| UI | 视图层 | 派生状态 |

—

## 四、十条架构戒律（精简版）

> **任何破坏以下戒律的改动，都是架构级缺陷。**

### 1️⃣ 按键不执行行为  
按键只表达意图，不直接产生效果。

### 2️⃣ FSM 只是输入设备  
FSM 只产生 token，永远不理解语义。

### 3️⃣ Grammar 拥有语义  
Vim 语义只存在于 Grammar 中，不执行、不裁决。

### 4️⃣ Kernel 是唯一权威  
所有决策、提升与裁决，只能发生在 Kernel。

### 5️⃣ Intent 是契约，不是实现  
Intent 与后端无关，可记录、可重放。

### 6️⃣ Builder 只做语义翻译  
Builder 冻结映射关系，不读状态、不执行。

### 7️⃣ Resolver 是技术债  
Resolver 只用于兼容，严禁新增功能。

### 8️⃣ 所有编辑必须是 Transaction  
绕过 Transaction 的编辑一律视为 bug。

### 9️⃣ UI 永远不是权威  
UI 是派生结果，不能驱动语义或逻辑。

### 🔟 怀疑不确定性  
如果逻辑不知道该放哪一层，说明设计已经出问题。

—

## 五、戒律 × 层级对照表（强制对号入座）

| # | 戒律 | ✅ 允许层 | ❌ 禁止层 |
|—|—|—|—|
| 1 | 按键不执行行为 | Executor | FSM / Grammar |
| 2 | FSM 只是输入设备 | FSM | Grammar / Kernel |
| 3 | Grammar 拥有语义 | Grammar | FSM / Kernel |
| 4 | Kernel 是唯一权威 | Kernel | 任何其他层 |
| 5 | Intent 是契约 | Intent | UI / Resolver |
| 6 | Builder 只做翻译 | Builder | Kernel / Executor |
| 7 | Resolver 是技术债 | Resolver（冻结） | 新逻辑 |
| 8 | 必须是 Transaction | Transaction | Executor 直改 |
| 9 | UI 不是权威 | UI（派生） | Grammar / Kernel |
|10 | 怀疑不确定性 | 设计阶段 | 实现层 |

—

## 六、代码审查快速判定

### 如果你看到：

- FSM 中出现 `if (vim_mode == ...)`  
  → ❌ 戒律 2 / 3

- Grammar 直接调用执行函数  
  → ❌ 戒律 3 / 8

- Kernel 里写 tmux 细节  
  → ❌ 戒律 4 / 5

- Intent 含 UI / 后端字段  
  → ❌ 戒律 5 / 9

- Executor 修改状态但无 Transaction  
  → ❌ 戒律 8（高危）

- Resolver 新增分支  
  → ❌ 戒律 7（红线）

—

## 七、贡献与治理规则

- ✅ 所有 PR 必须声明**涉及的架构层**
- ✅ PR 必须自检「是否违反任一戒律」
- ✅ 架构争议以本文件为最终裁决
- ✅ 新人 **先读本文件，再读代码**

—

## 八、最终原则（写给未来）

> **这不是一个“灵活”的系统，  
> 而是一个“抗腐化”的系统。**

> **不方便写的代码，  
> 往往正是未来最安全的代码。**

—

如果你愿意，我可以下一步直接给你：

- 📄 **PR 模板（强制戒律勾选）**
- ✅ **新功能 Checklist（一步步走对层）**
- 🧩 **FSM / Grammar / Kernel / Intent 的最小代码骨架**
- 🔒 **“什么时候允许破戒”的唯一例外条款**

你现在已经不是“写代码”，  
而是在**立法**了。