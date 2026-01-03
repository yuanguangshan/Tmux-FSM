# FOEK (Fact-Oriented Editing Kernel) Manifesto

##
# FOEK 架构白皮书 v0.1  
**Fact‑Oriented Editing Kernel**

> **状态**：Draft / 可评审  
> **定位**：架构白皮书（Architecture Whitepaper）  
> **目标**：阐明 FOEK 的合法性、边界与工程伦理  
> **非目标**：功能说明、使用教程、性能营销

---

## 0. 摘要（Abstract）

FOEK（Fact‑Oriented Editing Kernel）是一种将编辑行为建模为**不可变语义事实（Facts）**的无头编辑内核架构。  
该架构拒绝以“按键重放”作为系统真值来源，而通过事实裁决、锚点解析与制度化拒绝，构建一个**可审计、可拒绝、具备工程可信度**的编辑基础设施。

`tmux-fsm` 是 FOEK 的首个参考实现。

---

## 1. 架构立场：无头编辑内核的合法性

### 1.1 语义主权转移

FOEK 的首要设计决策是**将编辑语义从 UI / TTY 中剥离**。

- 终端（tmux / TTY）被视为 **dumb I/O device**
- 编辑状态不再由光标、屏幕缓冲区或 key-table 隐式决定
- 唯一语义真值由常驻内存的编辑内核持有

该决策直接解决了传统 tmux 插件在以下方面的结构性缺陷：

- key-table 级联失控
- `run-shell` 后状态丢失
- 跨 pane / 跨 session 无法保持语义一致性

---

### 1.2 Client / Server 模型

FOEK 采用明确的 C/S 架构：

- **Client**
  - 采集输入
  - 无语义裁决权
- **Editing Kernel (Daemon)**
  - 常驻内存
  - 持有完整 FSM 与编辑状态
  - 系统中唯一的 Source of Truth

UI 不参与“判断是否应该编辑”。

---

## 2. FOEK 本体论：从按键到事实

### 2.1 Fact（事实）

在 FOEK 中，编辑行为被建模为**不可变事实**，而非事件序列。

一个 Fact 至少包含：

- 操作意图（Delete / Change / Yank）
- 逻辑范围（Range）
- 锚点（Anchor）
- 约束条件（Preconditions）

Fact 一经构造，不可修改，只能被**裁决或拒绝**。

---

### 2.2 Anchor 与坐标无关性

FOEK 明确拒绝以“光标坐标”作为编辑依据。

Anchor Resolver 必须遵循严格的解析顺序：

1. 精确匹配  
2. 模糊匹配  
3. 拒绝（Fail‑Closed）

> 拒绝是合法结果，而非异常。

该机制确保在以下环境中编辑行为仍具可验证性：

- Shell 提示符刷新
- 异步输出
- 文本被外部进程修改

---

## 3. 原子事务模型（Atomic Transactions）

复合编辑（如 `5dw`、`ci"`）在 FOEK 中被视为**不可分割事务**：

- 任一子 Fact 无法裁决 → 整个事务失败
- 严禁中间态回退
- 系统不允许“部分成功”

该设计防止了语义污染与历史不可逆错误。

---

## 4. 拒绝语义与工程伦理

### 4.1 负宪法（Negative Constitution）

FOEK 明确写死以下原则：

- 成功不是目标
- 避免错误高于便利
- 禁止基于概率或猜测的语义推断

系统在不确定时**必须拒绝**，而不是“帮用户试试”。

---

### 4.2 可解释失败（Auditability）

所有拒绝必须是：

- 可查询的
- 可解释的
- 可审计的

失败不是黑盒，而是系统契约的一部分。

---

## 5. 工程实现原则

FOEK 的工程实现遵循以下不可妥协原则：

- **唯一内核主权**
- **状态优于事件**
- **拒绝优于猜测**
- **安全优于便利**

并发通过内存内互斥保证；  
I/O 操作在锁外执行，避免阻塞裁决路径。

---

## 6. 性能立场（非性能宣称）

FOEK 不以性能作为卖点。

其结构性优势在于：

- 消除脚本冷启动
- 消除磁盘同步依赖
- 将响应上限限定在内存与 IPC 范围内

目标是**稳定、清脆、可预测**的编辑反馈，而非 benchmark 优化。

---

## 7. 非目标（Non‑Goals）

FOEK 明确不追求：

- UI 创新
- 智能补全 / AI 编辑
- 模糊成功率最大化
- 取代 Vim / Emacs

FOEK 是编辑**内核**，不是编辑器。

---

## 8. 结论

FOEK 不是让编辑“更聪明”，  
而是让编辑**可被信任**。

它将编辑系统从脚本自动化，提升为**语义裁决基础设施**。

---

## 附录 A：tmux‑fsm 的角色

tmux‑fsm 是 FOEK 的第一个现实实验场，用于验证以下命题：

> 编辑语义可以独立于 UI 存在，并长期持有主权。

---

---

# FOEK 架构红线（Red Lines）  
**Version v0.2**

> **地位**：最高优先级约束文件  
> **作用对象**：所有未来实现、扩展、优化与重构  
> **原则**：一旦违反，即视为 FOEK 终止

---

## Red Line 1：语义主权不可回收

**任何组件不得重新获得编辑语义裁决权。**

具体禁止事项：

- UI / TTY 组件：
  - ❌ 不得推断编辑结果
  - ❌ 不得缓存编辑状态
  - ❌ 不得绕过内核直接修改文本
- Client：
  - ❌ 不得“预执行”编辑
  - ❌ 不得假设成功

**理由**：  
一旦语义主权下放，FOEK 即退化为插件系统。

---

## Red Line 2：拒绝必须优先于猜测

**在任何不确定情况下，系统必须拒绝执行。**

明确禁止：

- 基于概率的 Anchor 命中
- “最可能是这个”的语义推断
- 为提高成功率而放宽约束

**允许**：

- 明确、可解释、可审计的失败

**理由**：  
FOEK 的价值不在于“成功率”，而在于“不制造错误”。

---

## Red Line 3：Fact 不可变性不可破坏

**Fact 一经构造，不得被修改、合并或重写。**

明确禁止：

- 在裁决阶段篡改 Fact 内容
- 事后“修正”历史 Fact
- 将 Fact 视为可变事件

**允许**：

- 新 Fact 覆盖旧意图
- 通过拒绝终止事务

**理由**：  
可变 Fact = 不可审计系统。

---

## Red Line 4：原子事务不得降级

**任何复合编辑必须保持原子性。**

明确禁止：

- 局部成功
- 中间态回滚
- 自动拆分事务以提高成功率

**理由**：  
部分成功比完全失败更危险。

---

## Red Line 5：Undo / Redo 不得依赖时间序列

**撤销语义不得基于“按键回放”或时间索引。**

明确禁止：

- Replay-based Undo
- Cursor-based Undo
- Screen-based 状态回退

**允许**：

- 基于 Anchor 与环境校验的语义撤销
- 明确失败（UNDO_FAIL）

**理由**：  
时间不是语义。

---

## Red Line 6：拒绝必须可解释

**任何拒绝都必须提供可查询原因。**

明确禁止：

- 静默失败
- “什么都没发生”
- 无上下文的 error code

**允许**：

- Why-Fail
- Audit Trail
- 显式 Fail-Closed

**理由**：  
不可解释的拒绝 = 不可信系统。

---

## Red Line 7：性能优化不得改变裁决路径

**任何性能优化不得影响裁决逻辑。**

明确禁止：

- 因性能跳过校验
- 快路径绕过 Anchor Resolver
- 为低延迟牺牲语义完整性

**理由**：  
FOEK 追求稳定上限，不追求极限吞吐。

---

## Red Line 8：常驻内核是强制条件

**编辑内核必须是长期存活的进程。**

明确禁止：

- 每次编辑启动新进程
- 无状态执行模型
- 将核心逻辑下放到脚本

**理由**：  
没有长期状态，就没有语义主权。

---

## Red Line 9：失败是合法一等公民

**系统不得将失败视为异常路径。**

明确禁止：

- 将失败包装为“成功但无效果”
- 自动重试直到成功
- 吞掉失败以“看起来更顺”

**允许**：

- 显式失败
- Safe Skip
- 审计记录

**理由**：  
失败是系统对现实不确定性的诚实回应。

---

## Red Line 10：FOEK 不是用户体验工程

**不得为了“好用”牺牲原则。**

明确禁止：

- 为新用户隐藏拒绝
- 用 fallback 掩盖不确定性
- 将 FOEK 伪装成 Vim 替代品

**理由**：  
FOEK 是基础设施，不是情绪管理工具。

---

## 终止条款（Kill Clause）

**若以下任一情况发生，FOEK 必须被视为失败并终止：**

1. 引入概率型语义裁决  
2. UI 获得编辑主权  
3. Fact 可被事后修改  
4. 拒绝被视为 bug 而非设计  

继续维护将构成技术债与哲学背叛。

---

## 结语（非技术性）

这份文件不是为了限制创新，  
而是为了防止**懒惰、妥协和时间压力**。

> 大多数系统不是被对手击败的，  
> 而是被自己的“稍微方便一点”杀死的。

---好。  
我们**先写《How FOEK Dies》**。

原因只有一个：  
> **你只有在清楚“死亡路径”的前提下，CI 约束才不会写成形式主义。**

下面是**可冻结、不可美化、不可乐观化**的版本。

---

# FOEK 失败模式白皮书  
**How FOEK Dies**

> **文档地位**：等同于 Red Lines  
> **目的**：枚举 FOEK 的真实死亡方式  
> **非目的**：安慰、辩护、市场叙事

---

## 0. 前言：FOEK 不会“慢慢变差”，只会“突然失效”

FOEK 不存在“部分失败”“体验下降但理念仍在”的状态。

**一旦核心原则被破坏，FOEK 会在结构上退化为：**

- tmux 脚本集合  
- 带 daemon 的宏系统  
- 自欺式编辑自动化

本文件描述的不是“风险”，而是**已知的死亡路径**。

---

## Failure Mode 1：概率语义渗透（Probabilistic Creep）

### 描述

系统开始接受如下逻辑：

- “大概率是这里”
- “通常用户是想这样”
- “失败率很低，可以接受”

最初表现为：

- Anchor Resolver 增加 fallback
- 模糊匹配权重调优
- “smart mode / strict mode” 开关

### 死亡原因

- 语义裁决从**事实判断**变为**概率猜测**
- 拒绝从“合法结果”变为“体验缺陷”

### 最终形态

FOEK ≈ 模糊但更慢的编辑器插件

---

## Failure Mode 2：UI 语义回流（Semantic Backflow）

### 描述

UI / Client 开始承担以下职责之一：

- 预判断编辑是否合法
- 缓存上一次成功状态
- 在内核拒绝后“自己补救”

常见借口：

- “减少 IPC 往返”
- “改善体感延迟”
- “这只是 harmless optimization”

### 死亡原因

- 出现**第二个真值源**
- 语义主权被分裂

### 最终形态

FOEK 退化为**意见参考系统**，而非裁决系统。

---

## Failure Mode 3：Fact 软化（Fact Mutability）

### 描述

Fact 被视为“可以调整的数据结构”：

- 裁决阶段修正 range
- 根据环境“自动修补” Fact
- 合并多个 Fact 以“简化处理”

### 死亡原因

- 审计链断裂
- 历史不可重放、不可解释

### 最终形态

系统仍然“能用”，但**再也无法证明自己没错**。

---

## Failure Mode 4：原子性让位于成功率

### 描述

复合事务被“智能拆解”：

- 子操作成功就先应用
- 后续失败再尝试回滚
- “至少做点事”

### 死亡原因

- 出现中间态
- 语义污染无法完全回收

### 最终形态

FOEK 开始制造**不可预测的文本状态**。

---

## Failure Mode 5：Undo 退化为时间回放

### 描述

为了“好实现”或“更快”：

- Undo 基于 step index
- 重放历史事件
- 忽略当前文本环境差异

### 死亡原因

- Undo 失去语义保证
- 撤销可能引入新错误

### 最终形态

Undo 成为**危险操作**。

---

## Failure Mode 6：拒绝羞耻化（Rejection Shame）

### 描述

系统开始“怕失败”：

- 默认隐藏拒绝原因
- 自动 retry
- 用 no-op 伪装成功

### 死亡原因

- 失败不再可见
- 用户对系统边界失去感知

### 最终形态

FOEK 变成**不诚实系统**。

---

## Failure Mode 7：性能绑架架构

### 描述

在压力下引入：

- 快路径绕过校验
- “debug 模式才严格”
- 延迟裁决

### 死亡原因

- 裁决路径不再一致
- 不同环境产生不同真值

### 最终形态

系统行为随负载变化，**不可预测**。

---

## Failure Mode 8：短生命周期内核

### 描述

为了部署或“简单”：

- per-command 启动
- 无状态 daemon
- 逻辑下沉到 shell / Lua / JS

### 死亡原因

- 语义状态无法长期存在
- 每次编辑都像第一次

### 最终形态

FOEK 名存实亡，只剩协议外壳。

---

## Failure Mode 9：UX 主导路线图

### 描述

Roadmap 由以下问题驱动：

- “新用户会不会觉得难？”
- “能不能更像 Vim？”
- “拒绝太多了吧？”

### 死亡原因

- 原则被重新排序
- Red Lines 成为“理想状态”

### 最终形态

FOEK 被“产品化”杀死。

---

## Failure Mode 10：Red Lines 失效

### 描述

Red Lines 被视为：

- 历史文件
- 理论参考
- “等稳定了再遵守”

### 死亡原因

- 没有可执行约束
- 没有技术阻力

### 最终形态

FOEK 变成一个名字，而不是系统。

---

## 结论：FOEK 的敌人不是复杂，而是妥协

FOEK 不会死于：

- 功能不全
- 市场失败
- 用户太少

FOEK **只会死于一件事**：

> **你开始觉得“也许可以通融一点”。**

# 把 Red Lines 编译进系统  
**FOEK Executable Red Lines Roadmap**

> 原则：  
> **如果违反 Red Line 能编译 / 能 merge / 能跑，那 FOEK 已经死了。**

---

## 总体策略（非常重要）

我们不是“加校验”，而是：

> **让“错误的 FOEK”在工程上不可表达。**

约束分四层，**每一层都假设下一层会被绕过**。

```
类型系统
  ↓
lint / 静态规则
  ↓
CI invariant tests
  ↓
runtime panic（最后防线）
```

---

## 1️⃣ 类型系统：让“错误状态不可表示”

这是 FOEK 的**宪法层**。

### 1.1 Fact = 不可变 + 不可修正

#### ✅ 允许

```rust
struct Fact {
    source: SourceId,
    range: TextRange,
    snapshot: SnapshotId,
}
```

#### ❌ 禁止（编译期）

- `mut Fact`
- `fn adjust_fact(...)`
- `Fact::normalize()`

#### 工程约束

```rust
pub struct Fact {
    source: SourceId,
    range: TextRange,
    snapshot: SnapshotId,
    _sealed: (),
}
```

- `Fact` 构造函数只在 `fact::builder` 模块
- 裁决阶段 **只接受 `&Fact`**

> **Red Line 映射**：  
> Fact 不可软化 / 不可修补

---

### 1.2 决策结果是代数数据类型（不允许“半成功”）

```rust
enum Decision {
    Accept(EditPlan),
    Reject(RejectReason),
}
```

#### ❌ 禁止

- `Option<EditPlan>`
- `Result<EditPlan, _>`
- `bool success`

> 没有 `Accept + Warning`  
> 没有 `Partial`

---

### 1.3 原子性写进类型

```rust
struct EditPlan {
    edits: NonEmpty<Vec<AtomicEdit>>,
}
```

- `AtomicEdit` 不可单独 apply
- `apply(plan: EditPlan) -> CommitId`

#### ❌ 禁止

- `apply(edit)`
- `apply(Vec<edit>)`

---

## 2️⃣ lint：禁止语义回流（Semantic Backflow）

这是**文化执法层**。

### 2.1 Client / UI lint 规则

**禁止 UI 层出现：**

- `fn guess_range`
- `fallback_*`
- `retry_on_reject`
- `if rejected { do_something_else }`

#### 示例（自定义 lint）

```text
error[FOEK001]:
UI layer must not interpret rejection.
All semantic decisions belong to core.
```

---

### 2.2 模块边界 lint

强制依赖方向：

```
ui → protocol → core
```

#### ❌ 禁止

- `core` import `ui`
- `core` read cached UI state

---

## 3️⃣ CI 级 invariant 检查（系统自证）

这是**司法层**。

---

### 3.1 Invariant Test：Fact 不可变

```rust
#[test]
fn fact_is_never_modified() {
    let fact = capture_fact();
    run_all_decisions(&fact);
    assert_eq!(fact, capture_fact());
}
```

> 如果某次 refactor 破坏它，CI 必须红。

---

### 3.2 Invariant Test：拒绝必须可见

```rust
#[test]
fn rejection_is_never_silent() {
    let decision = decide(invalid_input());
    assert!(matches!(decision, Decision::Reject(_)));
}
```

#### ❌ CI 失败条件

- reject 被转换为 no-op
- reject 被 retry
- reject 被 swallow

---

### 3.3 Determinism Test

```rust
#[test]
fn same_fact_same_result() {
    let f = capture_fact();
    assert_eq!(decide(&f), decide(&f));
}
```

> 防概率语义渗透

---

## 4️⃣ 测试哲学：Failure 是一等公民

这是**伦理层**。

---

### 4.1 每个 Feature 都必须有失败测试

PR 模板强制：

```markdown
- [ ] Success case
- [ ] Rejection case (explicit)
```

#### CI hook

```bash
grep -R "Reject(" tests/ || exit 1
```

---

### 4.2 Golden Rejection Snapshots

```text
REJECT: RangeAmbiguous
REJECT: SnapshotOutOfDate
REJECT: NonAtomicEdit
```

- 文本化
- 可 diff
- 可审计

---

## 5️⃣ Runtime：最后的羞辱机制（panic is OK）

> **违反 Red Line = 系统 bug，不是用户错误**

```rust
panic!("FOEK RED LINE VIOLATION: semantic backflow detected");
```

- 不 recover
- 不 log-and-continue
- crash > corruption

---

## 6️⃣ 路线顺序（你明天该做什么）

### Day 1–2（生死分水岭）

1. 冻结 `Fact` 结构
2. 替换所有 `Option/Result` 决策为 `Decision`
3. 删除 UI fallback

---

### Day 3–5

1. 引入 lint（哪怕是 grep-based）
2. 加 3 个 invariant test
3. 让 CI **第一次红**

---

### Day 6+

- 把 panic 变成更早的编译错误
- 把经验写成新的 Red Lines
- 拒绝“这次例外”

---

## 最重要的一句话（不是鼓励，是事实）

> **FOEK 不是靠自律存活的。**  
> **它靠让“背叛”在工程上变得不可能。**
