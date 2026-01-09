# Tmux-FSM 架构规范检查清单

根据 `docs/ARCHITECTURE_MUST_READ.md` 文件中的架构规范，对项目中的文件进行逐一检查的结果。

## 检查结果汇总

### ✅ 符合架构规范的文件

1. **client.go**
   - 符合戒律1（按键不执行行为）：仅作为客户端发送按键信息到服务器，不直接执行行为

2. **config.go**
   - 符合架构规范：仅负责全局配置管理，不涉及业务逻辑

3. **engine.go**
   - 符合架构规范：实现光标引擎和计算逻辑，不直接执行编辑操作

4. **globals.go**
   - 符合架构规范：主要负责全局状态管理，不违反架构戒律

5. **intent.go**
   - 符合戒律5（Intent 是契约）：定义了 Intent 结构，与后端无关，可记录、可重放

6. **keymap.yaml**
   - 符合架构规范：属于 FSM 层面的配置，用于将按键输入映射到动作

7. **logic.go**
   - 基本符合架构规范：处理 FSM 逻辑，将按键转换为动作，不直接执行行为

### ⚠️ 部分符合架构规范的文件

1. **execute.go**
   - 违反戒律8（必须是 Transaction）：直接执行 tmux 命令，而不是应用 Transaction
   - 违反戒律4（Kernel 是唯一权威）：包含语义决策逻辑（判断是否为 Vim pane）
   - 违反戒律9（UI 不是权威）：处理 undo/redo 逻辑，这应该是 Intent 或 Transaction 层的职责

2. **intent_bridge.go**
   - 符合戒律7（Resolver 是技术债）：文件顶部明确标注为遗留桥接代码，仅用于向后兼容

3. **legacy_logic.go**
   - 符合戒律7（Resolver 是技术债）：文件顶部明确标注为遗留逻辑，仅用于向后兼容

### 📝 Shell 脚本文件

1. **enter_fsm.sh**
   - 符合架构规范：仅负责切换到 FSM 模式，不涉及核心逻辑

2. **fsm-exit.sh**
   - 符合架构规范：仅负责退出 FSM 模式，不涉及核心逻辑

3. **fsm-toggle.sh**
   - 符合架构规范：仅负责切换 FSM 模式，不涉及核心逻辑

4. **install.sh**
   - 符合架构规范：安装脚本，不涉及核心架构逻辑

## 子目录检查结果

### ✅ 符合架构规范的子目录

#### backend/
- **backend.go**: 符合架构规范，提供后端抽象接口，不包含业务逻辑

#### editor/
- **engine.go**: 符合架构规范，提供编辑器引擎接口和实现，专注于缓冲区操作
- **execution_context.go**, **selection_update.go**, **stores.go**, **text_object.go**, **types.go**: 符合架构规范

#### engine/
- **engine.go**: 符合架构规范，定义编辑引擎接口
- **concrete_engine.go**: 符合架构规范，提供引擎的具体实现

#### fsm/
- **engine.go**: 符合架构规范，实现 FSM 引擎，处理状态转换
- **keymap.go**, **nvim.go**, **token.go**, **ui_stub.go**: 符合架构规范

#### intent/
- **intent.go**: 符合戒律5（Intent 是契约），定义了 Intent 结构
- **grammar_intent.go**: 符合架构规范，定义 Grammar 专用的意图类型
- **motion.go**, **promote.go**, **range.go**, **text_object.go**: 符合架构规范
- **builder/** 目录：符合架构规范，实现 Intent 构建器

#### kernel/
- **kernel.go**: 符合架构规范，实现 Kernel，作为系统的唯一权威
- **decide.go**, **execute.go**, **intent_executor.go**, **resolver_executor.go**, **transaction.go**: 符合架构规范

#### planner/
- **grammar.go**: 符合架构规范，实现 Vim 语法解析，遵循 Grammar → Kernel → Intent 流程
- **grammar_test.go**: 符合架构规范

#### resolver/
- **resolver.go**: 符合戒律7（Resolver 是技术债），文件顶部明确标注为冻结状态，不再开发
- **context.go**, **motion_resolver.go**, **move.go**, **noop_engine.go**, **operator.go**, **repeat.go**, **types.go**, **undo.go**: 符合架构规范

#### types/
- **types.go**: 符合架构规范，定义类型结构

#### weaver/
- **weaver/core/types.go**: 符合架构规范，定义核心数据结构，包括 Fact、Transaction 等
- **weaver/adapter/tmux_adapter.go**: 符合架构规范，提供 Tmux 环境适配器
- **weaver/core/** 和 **weaver/adapter/** 目录中的其他文件：符合架构规范

## 架构规范违反情况总结

### 主要问题（execute.go）

1. **直接执行行为**：文件中包含大量直接执行 tmux 命令的代码，违反了"必须是 Transaction"的戒律
2. **语义决策逻辑**：包含判断是否为 Vim pane 的逻辑，违反了"Kernel 是唯一权威"的戒律
3. **执行层处理 undo/redo**：undo/redo 应该由 Intent 或 Transaction 层处理

### 建议改进

1. **重构 execute.go**：
   - 将直接执行 tmux 命令的逻辑改为生成 Transaction
   - 将语义决策逻辑移到 Kernel 层
   - 将 undo/redo 逻辑移到 Intent 或 Transaction 层

2. **保持遗留文件状态**：
   - intent_bridge.go 和 legacy_logic.go 按照注释说明保持为技术债状态，不扩展功能
   - resolver/ 目录中的文件已标记为冻结，不再开发新功能

## 架构合规性评分

- **完全合规文件**：大部分文件符合架构规范
- **部分合规文件**：少数文件存在架构问题
- **主要问题文件**：execute.go 存在明显的架构违规

## 总体评估

项目整体架构基本符合规范，特别是核心架构层（Kernel、Grammar、Intent、Transaction）的设计遵循了 FSM → Grammar → Kernel → Intent → Transaction 的单向管道架构。但 execute.go 文件存在明显违反架构规范的问题，需要重构以完全符合架构要求。