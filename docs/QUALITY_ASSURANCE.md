# Tmux-FSM 质量保证与测试规范 (v1.0)

## 1. 文档概述

本规范定义了 `tmux-fsm` 项目的全生命周期测试策略。由于项目涉及复杂的状态机、语法解析以及具有因果关系的历史审计系统，测试不仅要验证功能的正确性，还要验证其对“系统宪法”的遵循程度。

---

## 2. 测试架构：分层验证模型

我们采用 **L1-L4 四级验证模型**，确保从原子逻辑到系统架构的全面覆盖。

### L1 - 逻辑单元测试 (Logic Unit Tests)
*   **目标**: 验证独立模块的算法正确性。
*   **重点**: FSM 状态转移表、Vim 语法状态机、CRDT 拓扑排序、WAL 序列化。
*   **工具**: `go test`

### L2 - 组件集成测试 (Component Integration Tests)
*   **目标**: 验证不同模块间的“契约”是否履行。
*   **重点**: `Kernel` 与 `FSM` 的令牌传递、`Weaver` 与 `Backend` 的指令翻译。
*   **工具**: `go test -v ./kernel ./weaver/...`

### L3 - 场景端到端测试 (E2E Scenario Tests)
*   **目标**: 模拟真实用户在 Tmux 中的复杂编辑路径。
*   **重点**: 跨窗格操作、多行文本对象、带计数的撤销/重做。
*   **工具**: `tests/baseline_tests.sh` (基于 `tmux send-keys` 和 `capture-pane`)

### L4 - 架构符合性测试 (Architectural Compliance Tests)
*   **目标**: 确保开发过程中没有违反系统核心原则（如模块越权）。
*   **重点**: 静态代码依赖检查、动态不变量验证 (Invariants)。
*   **工具**: `custom lint scripts` & `invariant/test.go`

---

## 3. 核心测试用例集

### 3.1 基础状态机 (TC-FSM)
| ID | 测试名称 | 触发路径 | 验证点 (Assertion) | 优先级 |
| :--- | :--- | :--- | :--- | :--- |
| FSM-01 | 模式切换自洽 | `<prefix> f` | 命令 `tmux show-option -gv @fsm_active` 返回 `1` | P0 |
| FSM-02 | 快速超时回归 | 进入 `GOTO` 不按键 | 800ms 后 `@fsm_state` 自动从 `GOTO` 回归 `NAV` | P0 |
| FSM-03 | 递归防御机制 | 快速连按相同前缀 | FSM 不应产生死循环，能够正确重置层定时器 | P1 |
| FSM-04 | 计数器溢出/重置 | 输入 `999` 后按 `Esc` | 内部 `count` 变量必须清零，且 UI 清空 | P1 |

### 3.2 语法解析与意图推广 (TC-GRAMMAR)
| ID | 测试名称 | 触发路径 | 验证点 (Assertion) | 优先级 |
| :--- | :--- | :--- | :--- | :--- |
| GRM-01 | 基础 Operator 聚合 | 按键 `d` -> `w` | Kernel 日志显示生成 `IntentKind: Delete`, `Motion: Word` | P0 |
| GRM-02 | 计数倍乘效应 | 按键 `2` -> `d` -> `3` -> `w` | 生成 `Intent` 的总 Count 应为 `6` (2*3) | P0 |
| GRM-03 | 查找动作 (f/t) | 按键 `f` -> `x` | `Grammar` 状态机必须进入 `MotionFind` 并正确捕获字符 `x` | P1 |
| GRM-04 | 文本对象 (Inner/Around) | 按键 `d` -> `i` -> `"` | 生成 `Range: RangeTextObject`, `TextObject: QuoteDouble`, `Scope: Inner` | P1 |

### 3.3 因果审计与执行 (TC-WEAVER)
| ID | 测试名称 | 触发路径 | 验证点 (Assertion) | 优先级 |
| :--- | :--- | :--- | :--- | :--- |
| WEA-01 | 世界漂移检测 (Drift) | 后台修改文本后执行 `dw` | `ShadowEngine` 必须在验证阶段返回 `VerdictBlocked` 或记录警告 | P0 |
| WEA-02 | 原子事务回滚 | 执行 `c3w` 并在中途强制崩溃 | 重启后 `WAL` 能够定位未完成事务，`evidence.log` 无效碎片被标识 | P1 |
| WEA-03 | 撤销的坐标无关性 | 在不同行按 `u` | 文本必须恢复在**原始删除坐标**，而非当前光标位置 | P0 |
| WEA-04 | 证据完整性校验 | 执行 100 次混合操作 | 运行 `verifier verify` 必须通过，SHA256 因果链无断点 | P1 |

### 3.4 架构红线测试 (TC-COMPLIANCE/L4)
| ID | 检查规则 | 方法 | 违规判定 |
| :--- | :--- | :--- | :--- |
| ARC-01 | FSM 禁令 | `grep -r "os/exec" ./fsm` | 若存在任何物理执行调用，则测试失败 |
| ARC-02 | Kernel 解耦 | 检查 `kernel/` 是否 import 了 `ui` 包 | Kernel 只能发起 UI 刷新请求，不得直接渲染 |
| ARC-03 | AI 权限限制 | 模拟 `ActorID: "gemini-ai"` 提交 Intent | 必须被 `policy` 拦截，除非带有 `approval` 签名 |

---

## 4. 测试环境与自动化脚本

### 4.1 单元测试运行
```bash
# 运行所有单元测试并显示覆盖率
go test -v -cover ./...
```

### 4.2 E2E 测试运行 (Baseline)
使用 `tests/baseline_tests.sh` 驱动 Tmux 虚拟终端：
```bash
# 执行端到端基准测试
./tests/baseline_tests.sh --verbose
```

### 4.3 静态架构检查
```bash
# 检查 FSM 层是否存在非法物理越权
! grep -rE "os/exec|backend\.Global" fsm/
```

---

## 5. 验收标准 (Definition of Done)

*   **P0 (Critical)**: 所有 P0 测试用例必须 100% 通过。
*   **L4 (Architecture)**: 架构红线代码扫描 0 冲突。
*   **Regression**: 任何新功能提交必须包含至少一个对应的 L1 或 L2 测试。
*   **Audit**: `verifier` 工具对生成的 1000 条以上随机操作日志验证通过。

---

## 6. 附录：常见故障诊断路径 (Debug Flow)

若测试失败，请遵循以下审计链：
1.  检查 `~/tmux-fsm.log` 查看 FSM 令牌轨迹。
2.  检查 `.weaver/evidence.log` 查看 Intent 解析结果。
3.  使用 `tmux show-messages` 查看服务器 Panic 堆栈。
