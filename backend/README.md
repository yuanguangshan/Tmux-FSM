# backend 模块

## 模块职责概述

`backend/` 是 **Tmux-FSM 的后端执行层**，负责承接来自 Engine / Kernel 的执行请求，管理真实世界的副作用（tmux、IO、系统状态）。该模块作为"纯逻辑世界"与"外部系统"之间的边界层，将抽象的 Operation 映射为可执行动作。Backend 不理解 Intent，只执行 Operation。

主要职责包括：
- 执行具体的系统操作（tmux 命令、文件操作等）
- 管理外部系统的状态
- 提供可替换的执行后端（mock、tmux、test 等）
- 生成状态快照用于回放和验证

## 核心设计思想

- **最小语义**: 不包含策略，只执行命令
- **可替换性**: 不同 backend 可并存（mock / tmux / test）
- **幂等友好**: 配合 replay / verifier 使用
- **副作用隔离**: 将副作用限制在 backend 层

## 文件结构说明

### `backend.go`
- Backend 抽象定义
- 主要接口：
  - `Backend`: 后端执行接口
- 主要函数：
  - `Init() error`: 初始化后端
  - `Execute(operation.Operation) error`: 执行操作
  - `GetState() StateSnapshot`: 获取当前状态
  - `Close() error`: 关闭后端
- 是 Kernel / Engine 面向 backend 的唯一依赖点

### `tmux_backend.go`
- 基于 tmux 的具体 Backend 实现
- 主要函数：
  - `NewTmuxBackend() Backend`: 创建 tmux 后端实例
  - `ExecuteTmuxCommand(cmd string) (string, error)`: 执行 tmux 命令
  - `ManagePane(action string) error`: 管理 pane
  - `ManageWindow(action string) error`: 管理 window
  - `ManageSession(action string) error`: 管理 session
- 是 FSM 与 tmux 世界真正发生交互的地方

### `exec.go`
- 命令执行工具封装
- 主要函数：
  - `ExecuteCommand(cmd string, args ...string) (string, error)`: 执行命令
  - `ExecuteShellCommand(shellCmd string) (string, error)`: 执行 shell 命令
  - `CaptureOutput(cmd string) (string, error)`: 捕获命令输出
  - `CheckCommandExists(cmd string) bool`: 检查命令是否存在
- 统一处理 shell 调用、错误捕获、输出解析

### `state_snapshot.go`
- 后端状态快照
- 主要结构体：
  - `StateSnapshot`: 状态快照结构
  - `TmuxState`: tmux 状态信息
- 主要函数：
  - `CaptureState() StateSnapshot`: 捕获当前状态
  - `CompareState(a, b StateSnapshot) StateDiff`: 比较状态差异
  - `SerializeState(snapshot StateSnapshot) []byte`: 序列化状态
  - `DeserializeState(data []byte) StateSnapshot`: 反序列化状态
- 用于回放（Replay）和校验（Verifier）

### `mock_backend.go`
- 测试 / 仿真用 Backend
- 主要函数：
  - `NewMockBackend() Backend`: 创建模拟后端
  - `GetExecutedCommands() []string`: 获取执行的命令列表
  - `SetMockResponse(cmd string, response string)`: 设置模拟响应
- 用于单元测试和 FSM / Kernel 行为验证，不产生真实副作用

## 执行流程

```
Intent → Engine → Kernel → Backend → tmux/OS/IO
                        ↓
                   State Snapshot
```

## 在整体架构中的角色

Backend 是系统的执行层，它接收来自 Kernel 的决策结果，并将其转换为对实际系统的操作。Backend 确保所有副作用都被正确处理，并提供状态快照功能以支持系统的可回放性和可验证性。