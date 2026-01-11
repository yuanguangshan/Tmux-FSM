# Tmux-FSM 项目架构全景深度分析报告 (超详细版)

## 序言

本文档是对 `tmux-fsm` 项目的一次全维度、原子级的深度技术解剖。该项目不仅仅是一个 Tmux 插件，它是一套严密的、基于因果逻辑的“结构化编辑内核”，旨在通过数学验证和语义追踪，彻底解决编辑过程中的不确定性。

---

## 1. 核心架构：命令的“单向权力流动”

项目严格遵循 **编排 (Orchestration)** 模式，权力流动是单向且互不越权的：

**物理按键 (Physical Key) → FSM (感官) → 令牌 (Tokens) → Grammar (理解) → Kernel (决策) → 意图 (Intent) → Weaver (核心处理) → Backend (执行)**

### 架构准则 (Architecture Rules)
*   **不越权性**: FSM 引擎绝对不能理解动作，Grammar 绝对不能直接生成 Intent，只能生成中间态的 `GrammarIntent`。
*   **集中决策**: Kernel 是唯一的权力仲裁中心，负责将所有信息“提升 (Promote)”为可执行的意图。

---

## 2. 感知层：FSM 引擎 (fsm/engine.go)

FSM 引擎是系统的感知前端，其设计采用了 **“Emitter (发射器)”** 模式。

### 核心机制
*   **状态维护**: 记录 `Active` 状态（如 `NAV`, `GOTO`）、处理 `Layer` 切换及 `Timer` 自动超时返回。
*   **令牌生成**: 将按键序列转换为 `TokenKey` (字母)、`TokenDigit` (数字)、`TokenRepeat` (`.`) 等。
*   **`RunAction` 便利工具箱**: 包含一个映射表。对于简单的命令（如状态栏切换、简单的窗格移动），Kernel 会直接调用此方法，绕过复杂的解析逻辑以保证响应性能。

---

## 3. 解析层：Vim 语法解析器 (planner/grammar.go)

`Grammar` 结构体实现了一个高度复杂的复合解析器，通过一个微型状态机来处理 Vim 语法的“谓语-宾语”结构。

### 解析状态机
*   **Pending 状态**: 包含 `MPNone`, `MPG` (g-前缀), `MPF` (f-查找), `MPT` (t-跳转) 等。
*   **Count 累加**: 自动处理 `2dw`, `10j` 中的数字叠加。
*   **Text Object 映射**: 区分 `Inner` (i) 和 `Around` (a)，并映射到 `Word`, `Paren`, `Quote` 等对象。
*   **Operator 匹配**: 处理 `d`, `y`, `c` 等操作符，支持 `dd` 这样的双击行操作缩写。

---

## 4. 决策层：内核逻辑 (kernel/decide.go)

内核是系统的大脑，它利用 `GrammarEmitter` 这个巧妙的“中间人”将 FSM 与 Grammar 连接。

### 决策路径 (`DecisionKind`)
1.  **`DecisionIntent`**: Grammar 成功识别了一个完整句子（如 `dw`），Kernel 注入 PaneID 和 RequestID 形成完整意图。
2.  **`DecisionFSM`**: 这是一个简单的快捷键绑定。
3.  **`DecisionNone`**: FSM 消耗了按键但句子未完成（如按了 `d`），系统进入合法的等待状态。

---

## 5. 执行核心：Weaver (weaver/core/shadow_engine.go)

Weaver 是项目的灵魂，采用了 **“影子引擎 (Shadow Engine)”** 架构，负责处理具有因果关系的 Intent。

### 执行六步走 (The Six Steps)
1.  **Adjudicate (审判)**: 检查 **World Drift (世界漂移)**。对比 Intent 预期的 SnapshotHash 与当前系统的实际 Hash。若不匹配，根据策略拒绝或进行乐观执行。
2.  **Plan (计划)**: 生成 `Facts` (事实) 和 `InverseFacts` (逆向事实)。
3.  **Resolve (解析)**: 将抽象事实绑定到物理坐标（行号/列号）。这是解决“位置漂移”的关键步骤。
4.  **Project (投影)**: 调用 `Backend` 执行物理修改。
5.  **Verify (验证)**: 执行后立即对系统重新拍照，验证物理变更是否完全符合 `Facts` 预期，若违背则记录异常。
6.  **Audit (审计)**: 将完整的执行轨迹（Version v2, ActorID, RequestID）写入证据室。

### 撤销公理 (Undo Axioms)
*   **Axiom 7.5 (Verified Replay)**: 撤销不仅是简单的“反向操作”，而是经过验证的“状态回溯”。
*   **Axiom 7.6 (Return-to-Origin)**: 撤销操作被视为回归原点，而非新的分支分叉。

---

## 6. 证据持久化：证据室 (weaver/core/evidence_vault.go)

证据室实现了物理不可变的日志系统 (RFC-WC-003)。

*   **格式**: JSON Lines (JSONL)，便于流式读取和追加。
*   **落盘即裁决 (Atomic Sync)**: 调用 `f.Sync()` 强制物理写盘。只有写入成功，系统才认为操作已“决策”。
*   **司法索引 (Rebuild Index)**: 启动时扫描全文件，重建内存索引，支持通过 SHA256 哈希值实现 O(1) 级别的证据调阅。

---

## 7. 协作基因：CRDT 与 语义事实

### `SemanticEvent` (crdt/crdt.go)
每一个操作都包含 `CausalParents`。系统推崇的是“因果序”而非“时间序”，这解决了分布式环境下按键到达顺序错乱的问题。

### `TextObject` 与 `Anchor`
系统定义了多种锚点 (`AnchorAbsolute`, `AnchorSemantic`)。这种设计允许系统在文本内容被非实时修改时（如他人同时编辑），依然能准确寻找到用户的编辑目标。

---

## 8. 安全保障：Policy 与 Invariants

### 安全策略 (`Policy`)
将系统权限分为 `TrustSystem`, `TrustUser`, `TrustAI`。**AI 操作受到严格限制**，它不具备提交修订历史的直接权力，必须通过“提议-批准”模型进行。

### 逻辑物理定律 (`Invariant`)
系统内置了基于属性的随机测试 (Property-Based Testing)，持续验证“(操作 + 撤销) == 原状态”这一等式。这是系统敢于称其为“高可靠编辑内核”的底气。

---

## 总结：项目的本质

`tmux-fsm` 不是一个简单的按键绑定脚本集合。

它是一套 **“可证明的编辑系统原型”**。它将用户的每一个按键赋予了法律级别的严谨性：通过 FSM 捕获，Kernel 裁决，Weaver 编织，证据室公证。

这种高度结构化的设计，使其未来能够无缝扩展为：
1.  **AI 手术刀**: 让 AI 精准地修改代码范围，而不是生成充满错误的 diff。
2.  **跨端协作云**: 多个用户通过不同的宿主（Tmux/Neovim）在同一个因果链条上联合编辑。
3.  **语义版本控制**: 取代基于文本行的 Git，通过意图合并彻底告别 Merge Conflict。
