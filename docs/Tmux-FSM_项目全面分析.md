# Tmux-FSM 项目全面分析

## 📌 项目定位

**Tmux-FSM** 远不止是一个 tmux 插件——它是一个**工业级的、与 UI 解耦的编辑语义内核**（称为 **Weaver Core** 或 **FOEK - Fact-Oriented Editing Kernel**）。

- **代码规模**: 约 23,148 行 Go 代码
- **测试覆盖**: 16 个测试文件
- **文档体系**: 40+ 个架构文档和技术白皮书

---

## 🏗️ 核心架构

项目遵循严格的分层架构，由**十条架构戒律**约束：

```
物理按键 → FSM状态机 → Grammar语义解析 → Kernel仲裁层 → Intent意图 → Builder构建 → Transaction事务 → Executor执行 → UI渲染
```

### 架构十大原则

1. **按键不执行行为** — 按键只表达意图，不直接产生效果
2. **FSM只是输入设备** — 只产生 token，不理解语义
3. **Grammar拥有语义** — Vim 语义只存在于 Grammar 中
4. **Kernel是唯一权威** — 所有决策只能发生在 Kernel
5. **Intent是契约** — 与后端无关，可记录、可重放
6. **Builder只做语义翻译** — 冻结映射关系，不读状态
7. **所有编辑必须是Transaction** — 绕过 Transaction 的编辑一律视为 bug
8. **UI永远不是权威** — UI 是派生结果，不能驱动语义
9. **可审计性优先** — 任何编辑裁决必须可解释、可回溯

---

## 🎯 核心功能模块

### 1. FSM 引擎 (`fsm/engine.go`)

纯粹的输入设备，将按键序列转换为 RawToken：
- 支持多层状态（NAV、GOTO 等）
- 层超时自动重置（如 GOTO 层 800ms 超时）
- 数字计数器支持（如 `3j` 中的 `3`）

### 2. Kernel 中央仲裁层 (`kernel/kernel.go`)

系统的唯一权威决策者：
- 处理 Intent 的上下文化（绑定 PaneID）
- 支持 RequestID/ActorID 追踪（审计基础）
- 支持 ShadowIntent 模式（新旧实现对比）

### 3. Planner 语法解析器 (`planner/grammar.go`)

拥有 Vim 语义的唯一模块：
- 解析 Vim 操作符
- 处理文本对象（Text Objects）和 Motion 前缀
- 支持字符查找（f、F、t、T）

### 4. Editor 内核 (`editor/`)

**可组合、可回放、可检测冲突的编辑执行内核**：
- **ResolvedOperation 抽象** — 所有操作都实现了 `Apply()`、`Inverse()`、`Footprint()` 接口
- **Operation DAG** — 非线性的历史图（支持分支、合并）
- **冲突检测** — 基于空间/语义的 Footprint 系统
- **操作代数** — Insert/Delete/Move/Rename 等操作的数学抽象

### 5. CRDT 模块 (`crdt/`)

无冲突复制数据类型：
- 因果有序（Causal Ordering）
- 确定性收敛
- 支持多用户协作编辑

### 6. Weaver 装配层 (`weaver/`)

系统的唯一装配入口：
- 依赖注入和模块管理
- 事实解析与执行
- 历史管理与一致性验证

---

## ✨ 核心特色与创新

### 1. Intent-First 架构

编辑不再是 UI 行为，而是**语义事件**：
- Intent 描述"用户想做什么"，而非"如何实现"
- 可序列化、可验证、跨环境一致
- 支持语义级宏和重放

### 2. Operation DAG 历史模型

- **非线性 undo/redo** — 可以在任意分支间切换
- **语义 diff** — 而非文本 diff
- **自动合并** — 基于空间/语义冲突检测
- **Git 友好** — 历史模型类似 Git

### 3. Snapshot 系统

- 基于 LineID 的稳定引用
- 快照哈希校验
- 环境变化检测
- 投影兼容性

### 4. 审计优先的设计

```go
type HandleContext struct {
    Ctx       context.Context
    RequestID string  // 请求唯一标识
    ActorID   string  // 用户/窗格/客户端标识
}
```

每个操作都有完整的审计链：可解释、可回溯、可被质疑

### 5. 跨环境一致性

在 Shell、Vim、REPL 等不同环境中提供**统一的编辑语义**

---

## 🔐 设计哲学

项目的核心理念是 **FOEK（事实导向的编辑内核）**：

> **"编辑不是 UI 行为，而是语义事件"**

### 关键原则

- **安全高于还原** — 不确定的 Undo 拒绝执行
- **正确高于便利** — 模糊操作必须标注
- **主权必须集中** — Daemon 是唯一语义真值持有者
- **可审计性是信任机制** — 任何编辑裁决都可回溯

---

## 📊 与传统编辑器的对比

| 维度 | Vim | Tmux-FSM |
|------|-----|----------|
| 架构 | 紧耦合 | 分层解耦 |
| Undo | 线性栈 | DAG + Snapshot |
| 语义 | 隐式 | 显式 Intent |
| 审计 | 无 | 完整审计链 |
| 跨环境 | 否 | 是（Shell/Vim/REPL） |
| 协作 | 否 | CRDT DAG |
| 冲突检测 | 无 | Footprint 系统 |

---

## 📁 项目结构

```
Tmux-FSM/
├── 核心组件
│   ├── main.go              # 入口、事务管理
│   ├── fsm/                 # 有限状态机引擎
│   ├── kernel/              # 中央仲裁层
│   ├── intent/              # 意图表示系统
│   └── planner/             # Vim 语法解析器
│
├── 执行层
│   ├── editor/              # 编辑执行内核（DAG、操作代数）
│   ├── backend/             # tmux 命令执行层
│   └── weaver/              # 系统装配与事实解析
│
├── 数据模型
│   ├── types/               # 核心类型定义
│   ├── crdt/                # 无冲突复制数据类型
│   ├── undotree/            # 撤销树管理
│   └── snapshot.go          # 快照系统
│
├── 配置
│   ├── keymap.yaml          # FSM 键盘映射
│   └── default.tmux.conf    # tmux 配置模板
│
└── 文档
    └── docs/                # 40+ 个架构文档
```

---

## 💡 技术亮点

1. **代数化操作** — Insert/Delete 等操作有严格的数学定义
2. **类型安全** — 强类型系统（IntentKind、OperatorKind、MotionKind 等）
3. **事务模型** — 原子性、Vim 语义规则（`.` repeat）、事务日志
4. **冲突检测** — 基于 Footprint 的空间/语义冲突分析
5. **模块化设计** — 职责明确，边界清晰

---

## 📈 项目成熟度

### 已完成 ✅
- FSM 引擎与状态管理
- Kernel 仲裁层
- Grammar 解析器
- Intent 系统
- Editor 内核（DAG、操作代数、冲突检测）
- CRDT 基础
- 快照系统
- 事务管理
- 审计日志
- 16 个测试文件

### 进行中 🔄
- Weaver 系统集成
- Resolver 重构（标记为 legacy）

### 待实现 📋
- 合并逻辑
- Rebase/Squash
- 完整的协作编辑
- LSP/AST 投影

---

## 🎯 适用场景

### 适合 ✅
- 需要严格审计的编辑环境
- 复杂的协作编辑场景
- 对 Undo 安全性要求极高的系统
- 研究编辑器架构的学术/工程探索

### 不太适合 ❌
- 只需要简单快捷键的用户
- 追求轻量级的场景
- 不关心架构哲学的实用主义者

---

## 🔑 关键文件索引

### 核心代码
- `main.go` — 入口、事务管理
- `fsm/engine.go` — FSM 引擎
- `kernel/kernel.go` — 中央仲裁层
- `planner/grammar.go` — Vim 语法解析
- `intent/intent.go` — Intent 定义
- `editor/dag.go` — 操作 DAG
- `editor/footprint.go` — 冲突检测

### 架构文档
- `docs/ARCHITECTURE.md` — 架构宪法（十条戒律）
- `docs/reference/DESIGN_PHILOSOPHY.md` — 设计哲学
- `docs/EDITOR_IR_SPEC.md` — 编辑器 IR 规范
- `docs/COLLABORATIVE_EDITING_MODEL.md` — 协作编辑模型

### 配置文件
- `keymap.yaml` — 键盘映射配置
- `default.tmux.conf` — tmux 配置模板

---

## 🏆 总结评价

这是一个**野心勃勃、设计严谨、哲学深刻**的工程项目。它不满足于做一个好用的工具，而是试图**重新定义编辑系统的本质**。

### 核心价值

- **理论贡献**：提出了 FOEK（事实导向的编辑内核）范式
- **工程实践**：实现了严格的分层架构和意图系统
- **安全优先**：将 Undo 从"便利功能"提升为"安全裁决"
- **跨环境一致性**：在不同环境间提供统一的编辑语义

### 代码质量

- **架构清晰**：分层明确，职责单一
- **文档完善**：40+ 个架构文档
- **类型安全**：强类型系统
- **可测试性**：16 个测试文件，阶段完成标记

### 学习曲线

**复杂度较高，学习曲线陡峭**，但其核心理念——**将编辑能力从 UI 中剥离，建立可审计的语义内核**——具有前瞻性和启发性。

### 推荐人群

如果你对以下主题感兴趣，这是一个值得深入研究的项目：
- 编辑器架构设计
- 协作编辑和 CRDT
- 分布式系统的语义一致性
- 意图驱动的用户界面
- 可审计的系统设计

---

*生成时间: 2026-01-14*
*分析工具: Claude Code*
