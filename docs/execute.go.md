# execute.go: 动作执行引擎

`execute.go` 负责将 FSM 产生的高层意图 (Actions) 转化为具体的系统操作 (tmux/vim 命令)。

## 核心职责

1. **tmux 交互**：通过 `exec.Command` 调用 tmux CLI。
2. **Vim 感知**：检测当前 pane 是否处于 Vim 环境，从而决定是发送 `tmux-next-word` 还是 `w`。
3. **文本捕获 (Fact Capture)**：实现 `captureText`，从 tmux 缓冲区抓取文本片段作为"空间事实"。
4. **撤销/重做执行**：执行由 `logic.go` 生成的反向操作。
5. **锚点解析**：实现精确和模糊的锚点解析，支持文本定位。
6. **执行器模式**：支持多种执行器（ShellExecutor、VimExecutor）。
7. **事务管理**：记录操作事实和反向操作，支持事务性撤销。

## 核心数据结构

- **Executor**: 执行器接口，定义了执行操作的方法。
- **ResolvedAnchor**: 解析后的锚点，包含行号和解析结果。
- **ResolveResult**: 解析结果类型（精确、模糊、失败）。

## Axioms (公理)

- **Axiom 1: Transaction Atomicity**: 事务原子性 - 整个事务单元处理。
- **Axiom 2: Anchor Primacy**: 锚点优先 - 执行前始终解析锚点。
- **Axiom 3: Exactness Preference**: 精确性优先 - 首先尝试精确匹配。
- **Axiom 4: Mandatory Failure Conditions**: 强制失败条件 - 锚点未找到时失败。
- **Axiom 5: No Partial Trust**: 无部分信任 - 模糊事务不与精确事务混合。
- **Axiom 6: Permitted Fuzzy Conditions**: 允许模糊条件 - 仅在狭窄窗口内尝试模糊匹配。
- **Axiom 7: Transaction-Level Degradation**: 事务级降级 - 一个模糊污染整个事务。
- **Axiom 8: Fuzzy Transparency**: 模糊透明 - UI 必须明确通知用户。
- **Axiom 11: Explainability**: 可解释性 - 失败必须可见且可解释。

## 支持的操作

- **移动操作**: 上下左右、单词前后、行首行尾、文件首尾等。
- **删除操作**: 字符、单词、行等的删除。
- **修改操作**: 改变文本内容。
- **复制操作**: yank 操作。
- **视觉模式操作**: 在视觉模式下执行操作。
- **文本对象操作**: 处理单词、括号、引号等文本对象。
- **查找操作**: f/F/t/T 查找字符。
- **搜索操作**: / 搜索、n/N 导航。

## 关键技术点

- **竞态消除**：在 `copy-pipe` 之后等待 5ms，确保 `show-buffer` 读到的数据是完整的。
- **静默执行**：所有命令通过 `run-shell -b` 或后台执行，避免阻塞 UI。
- **锚点解析**：使用 SHA256 哈希验证行内容，实现精确和模糊定位。
- **多执行器支持**：根据环境选择合适的执行器（Vim 或 Shell）。
- **安全撤销**：基于锚点的撤销机制，确保只在文本未改变时执行撤销。
