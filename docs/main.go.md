# main.go: 守护进程框架

`main.go` 是整个 tmux-fsm 的骨架，实现了高性能的常驻内存服务。

## 核心职责

1. **C/S 架构**：实现 Unix Socket 服务端与极简客户端。
2. **并发控制**：利用 `sync.Mutex` 解决 Goroutine 间的状态竞争。
3. **生命周期管理**：处理信号 (SIGINT/SIGTERM)、自动保存与 Socket 清理。
4. **持久化**：通过"读写分离锁"策略安全地将内存状态映射到 tmux options。
5. **状态管理**：维护 FSM 状态、撤销/重做栈和事务管理。
6. **协议处理**：支持新的协议格式 "PANE_ID|CLIENT_NAME|KEY"。

## 运行模式

- **Server 模式 (`-server`)**: 持久运行。
- **Client 模式**: 发送单个按键并退出。
- **停止服务模式 (`-stop`)**: 停止正在运行的守护进程。

## 核心数据结构

- **Anchor**: 定位文本位置的结构，包含 pane ID、行提示、行哈希和光标位置。
- **Range**: 表示文本范围，用于记录操作的目标。
- **Fact**: 表示操作事实，包含类型、目标范围、元数据和副作用。
- **ActionRecord**: 操作记录，包含事实和其逆操作。
- **Transaction**: 事务结构，包含 ID、记录列表、创建时间等。
- **TransactionManager**: 事务管理器，负责事务的开始、追加和提交。
- **FSMState**: FSM 状态，包含模式、操作符、计数器等。

## ABI (应用二进制接口) 层

- **Intent Submission Layer**: 前端通过 Unix Socket 发送原始信号或内部命令到内核。
- **Verdict Trigger**: 内核开始对给定意图进行审议。
- **Audit Closure**: 内核完成审议并提交到时间线。
- **Heartbeat Lock**: 更新状态栏并重新断言键表以防止"一次性"断开。
- **Side Effect Projection**: 将审议结果最终确定为"已应用"，并将事实投射到物理TTY。
- **Verdict Deliberation**: 内核评估意图与当前世界状态。
- **Inverse Verdict Deliberation**: 处理反向操作（撤销）。

## 性能表现

- **响应延迟**: < 1ms。
- **内存占用**: 极低，二进制体积优化后约 2.9MB。
