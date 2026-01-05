// NAV-010: Semantic keys must never be intercepted here

---

# NAV 键权定律 · Design Rules（v0.1）

> **Document Status**: Draft  
> **Scope**: tmux-fsm / NAV Layer  
> **Intent**: Prevent long-term semantic erosion of the NAV layer  
> **Audience**: Core maintainers, reviewers, future contributors  

---

## 0. 定义（Normative）

### 0.1 NAV 层（Navigation Layer）

NAV 是 tmux-fsm 的**交互基态（Interaction Baseline）**，具有以下不可变属性：

- 系统启动后的默认状态
- 所有瞬态层（GOTO / CMD / …）结束、超时或异常后的唯一回归点
- PanicGuard 的强制复位目标

> **NAV 不是一个“模式（Mode）”，而是系统的稳定基态（Baseline State）。**

---

## 1. 设计目标（Design Intent）

NAV 层的存在目的**不是提供功能密度**，而是：

1. 提供**最低侵入性**的空间级导航能力
2. 作为**模式跃迁的唯一入口**
3. 作为**系统失序时的稳定吸引子（Global Attractor State）**

任何设计决策，若削弱上述三点之一，均视为违反本规则集的精神。

---

## 2. NAV 键权基本法（Key Sovereignty Law）

### NAV-001：白名单原则（Whitelist Only）

NAV 层**只允许处理显式列入白名单的按键**。

- 白名单之外的所有按键：
  - **必须无条件 PassThrough**
  - **不得被拦截、重映射或解释**

> ✅ NAV 是“允许什么”  
> ❌ NAV 不是“禁止什么”

---

### NAV-002：主权放行（PassThrough is Mandatory）

PassThrough 不是实现细节，而是**主权声明**。

- Shell / Vim / Application Kernel 对以下内容拥有永久主权：
  - 文本编辑
  - 语义跳转
  - 历史、撤销、选择、删除

NAV 层不得以任何理由“临时借用”这些按键。

---

## 3. 语义禁区（Semantic Exclusion Zone）

### NAV-010：禁止编辑语义键（Hard Prohibition）

NAV 层**严禁**定义或拦截以下具有强编辑语义的单键（包括但不限于）：

```
w e b
d c y
u r
0 ^ $
x s
```

**理由**：  
这些按键在 Vim / Shell 生态中具有高度稳定、深度肌肉记忆绑定的语义。

> **任何在 NAV 中拦截上述键位的行为，均构成架构级破坏。**

---

### NAV-011：数字键永久放行

```
0–9
```

- 不得用于计数
- 不得用于前缀
- 不得用于隐藏功能

**原因**：  
数字键在 Vim 中承载组合语义，其被拦截会造成不可预测的认知断裂。

---

## 4. 允许的 NAV 行为类别（Exhaustive）

### NAV-020：允许的行为类型（Only These）

NAV 层**仅允许**以下三类行为：

#### 4.1 空间级导航（Spatial Navigation）

- Pane / Window / Session 的**直接切换**
- 不涉及内容、历史或语义状态

示例（非穷举）：

```
h j k l
```

---

#### 4.2 模式跃迁（Mode Transition）

- 进入瞬态子层（如 GOTO / CMD）
- 跃迁必须是：
  - 显式的
  - 可逆的
  - 有超时兜底的

示例：

```
g → GOTO
: → CMD
```

---

#### 4.3 主权让渡（Sovereignty Yield）

- 明确将控制权交还给 Kernel
- FSM 在此期间不再参与输入处理

示例：

```
i → Suspend FSM
```

---

### NAV-021：禁止隐式复合行为

NAV 层不得存在：

- 多步状态积累
- 隐式上下文
- “下一次按键才生效”的设计

> NAV 中的每一次按键，都必须是**瞬时、原子、无记忆的**。

---

## 5. 时间与回归规则（Temporal Safety）

### NAV-030：NAV 是唯一回归目标

以下事件**必须无条件回到 NAV**：

- 瞬态层操作完成
- 超时（如 ≥800ms 无输入）
- `Esc`
- PanicGuard 触发

不得回到：
- 上一个子层
- 上一个子状态
- 上一个上下文

---

## 6. 代码结构约束（Enforcement by Structure）

### NAV-040：集中定义原则

- 所有 NAV 键绑定：
  - **必须集中定义**
  - **不得分散在功能模块中**

推荐结构（示意）：

```
nav/
 ├── allowed_keys.ts
 ├── passthrough.ts
 ├── forbidden_semantics.ts
```

> **“顺手在某个文件里加一个 NAV 键”被视为流程违规。**

---

## 7. Code Review 检查清单（Mandatory）

任何涉及 NAV 的 PR，Reviewer 必须回答以下问题：

1. 该按键是否已在 NAV 白名单中？
2. 它是否可能拦截 Vim / Shell 的既有肌肉记忆？
3. 如果用户此刻正在 Vim Normal Mode：
   - 这个拦截是否会让用户困惑或愤怒？
4. 这个行为是否可以放在瞬态层而不是 NAV？

**若任一问题无法明确回答 “否 / 安全” → PR 不得合并。**

---

## 8. 版本声明

> **NAV 键权定律是破坏性变更保护规则。**  
> 任何对本规则的修改，必须：
>
> - 明确标注版本
> - 给出破坏理由
> - 说明为什么“未来不会后悔”

---

# 结束语（非规范性）

> NAV 的价值不在于它能做什么，  
> 而在于 **它坚持不做什么**。

---
