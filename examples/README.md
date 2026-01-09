# examples 模块

## 模块职责概述

`examples/` 是 **Tmux-FSM 的示例代码与使用案例集合**，负责提供各种使用场景的完整示例代码和配置案例。该模块展示了如何正确使用系统的各项功能，为开发者提供实际的参考和学习资源。

主要职责包括：
- 提供系统功能的完整使用示例
- 展示各种配置和使用场景
- 包含实际应用的代码案例
- 提供学习和参考的示例资源

## 核心设计思想

- **实用性**: 提供真实的使用场景示例
- **完整性**: 每个示例都是可运行的完整代码
- **教育性**: 帮助理解系统功能的使用方法
- **多样性**: 覆盖不同的使用场景

## 文件结构说明

### `basic_usage/`
- 基础使用示例
- 主要内容：
  - `simple_intent.go`: 简单意图使用示例
  - `basic_fsm.go`: 基础 FSM 使用示例
  - `engine_example.go`: Engine 使用示例
  - `crdt_example.go`: CRDT 使用示例
- 展示系统基础功能的使用方法

### `advanced_features/`
- 高级功能示例
- 主要内容：
  - `multi_user_editing.go`: 多用户编辑示例
  - `complex_intents.go`: 复合意图示例
  - `custom_policy.go`: 自定义策略示例
  - `advanced_replay.go`: 高级回放示例
- 展示系统高级功能的使用方法

### `integration/`
- 集成示例
- 主要内容：
  - `neovim_integration.go`: Neovim 集成示例
  - `tmux_integration.go`: Tmux 集成示例
  - `api_integration.go`: API 集成示例
  - `cli_integration.go`: CLI 集成示例
- 展示与其他系统的集成方法

### `configuration/`
- 配置示例
- 主要内容：
  - `keymap_config.yaml`: 键映射配置示例
  - `policy_config.yaml`: 策略配置示例
  - `engine_config.yaml`: 引擎配置示例
  - `fsm_config.yaml`: FSM 配置示例
- 提供各种配置文件的示例

### `workflows/`
- 工作流示例
- 主要内容：
  - `editing_workflow.go`: 编辑工作流示例
  - `collaboration_workflow.go`: 协作工作流示例
  - `automation_workflow.go`: 自动化工作流示例
  - `debugging_workflow.go`: 调试工作流示例
- 展示完整的使用工作流

## 示例特性

### 可运行性
- 所有示例都是可直接运行的代码
- 包含完整的依赖和配置
- 提供详细的运行说明

### 教学性
- 详细的注释说明
- 渐进式的复杂度
- 清晰的概念展示

### 实用性
- 基于真实使用场景
- 包含最佳实践
- 提供常见问题的解决方案

## 在整体架构中的角色

Examples 模块是项目的实践指导层，它通过具体的示例代码帮助用户理解系统的使用方法。Examples 提供了：
- 功能使用的实际演示
- 配置和集成的参考
- 学习和开发的起点
- 最佳实践的展示