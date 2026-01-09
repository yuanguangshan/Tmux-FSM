# cmd 模块

## 模块职责概述

`cmd/` 是 **Tmux-FSM 的命令行工具入口**，作为 Tmux-FSM 与外部世界交互的命令行边界。该模块定义了不同运行模式下的启动入口，解析命令行参数，并将 CLI 行为转换为内部 Engine / Kernel 调用，方便开发、调试和脚本化集成。

主要职责包括：
- 提供可执行命令的统一入口点
- 解析和处理命令行参数
- 封装系统功能的命令行接口
- 支持开发调试和运维操作

## 核心设计思想

- **单一职责**: 每个命令文件对应一个明确用途的命令
- **薄层设计**: 尽量保持薄层，不包含业务逻辑
- **参数解析**: 专注于参数解析、初始化、调用转发
- **统一入口**: 作为所有非交互式入口的统一位置

## 文件结构说明

### `root.go`
- CLI 根命令定义
- 主要函数：
  - `NewRootCmd() *cobra.Command`: 创建根命令
  - `Execute() error`: 执行命令
  - `InitConfig()`: 初始化配置
- 负责：
  - 初始化配置
  - 注册子命令
  - 统一处理 flags

### `fsm.go`
- FSM 相关命令定义
- 主要函数：
  - `NewFsmCmd() *cobra.Command`: 创建 FSM 命令
  - `StartFsmMode() error`: 启动 FSM 模式
  - `StopFsmMode() error`: 停止 FSM 模式
- 对应：
  - 启动 tmux-fsm 服务
  - 进入 FSM 监听/事件循环

### `debug.go`
- 调试相关命令
- 主要函数：
  - `NewDebugCmd() *cobra.Command`: 创建调试命令
  - `PrintFsmState()`: 打印当前 FSM 状态
  - `DumpIntents()`: 导出 Intent / Transaction
  - `TriggerReplay()`: 手动触发回放
- 用于：
  - 打印当前 FSM 状态
  - Dump Intent / Transaction
  - 手动触发回放

### `replay.go`
- 历史回放相关命令
- 主要函数：
  - `NewReplayCmd() *cobra.Command`: 创建回放命令
  - `ReplayHistory(from, to string) error`: 执行历史回放
  - `ValidateReplay() error`: 验证回放结果
- 通常会调用 replay/ 和 verifier/ 模块
- 用于调试一致性问题

### `server.go`
- 服务器相关命令
- 主要函数：
  - `NewServerCmd() *cobra.Command`: 创建服务器命令
  - `StartServer() error`: 启动服务器
  - `StopServer() error`: 停止服务器
- 用于启动 HTTP/gRPC 服务

### `test.go`
- 测试相关命令
- 主要函数：
  - `NewTestCmd() *cobra.Command`: 创建测试命令
  - `RunIntegrationTests()`: 运行集成测试
  - `GenerateTestData()`: 生成测试数据
- 用于测试和实验性功能

## 使用场景

- **本地调试**: 调试 FSM 行为和状态
- **手动回放**: 手动 replay 一段历史
- **CI/CD**: 在持续集成中运行一致性校验
- **开发验证**: 开发阶段快速验证 Intent → 执行路径
- **运维操作**: 系统监控、状态查询等运维任务

## 在整体架构中的角色

cmd 模块是系统的命令行接口层，它将命令行参数转换为对内部系统的调用。作为所有"非 tmux 热键"入口的统一位置，cmd 模块提供了系统功能的标准化命令行访问方式，支持自动化脚本和运维操作。