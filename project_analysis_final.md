# Tmux-FSM 项目分析

## **项目概述：Tmux-FSM**

这是一个基于有限状态机(FSM)的 tmux 插件系统，旨在为终端导航和窗格管理提供类似 Vim 的模态编辑体验。该项目采用模块化架构，支持配置驱动的按键绑定和意图驱动的编辑操作。

### **核心功能**
- **模态导航**：类似 Vim 的状态机驱动界面
- **配置驱动**：YAML 格式的按键映射配置
- **意图系统**：高级编辑意图到低级操作的转换
- **事务管理**：支持撤销/重做和宏录制
- **协作编辑**：CRDT 支持的多用户编辑
- **状态同步**：与 Neovim 的模式同步

### **主要文件分析**

#### **核心入口文件**
- **`main.go`** (711行)：程序主入口，包含服务器/客户端模式、事务管理、宏系统、信号处理等核心逻辑
- **`config.go`** (69行)：配置管理，支持 Legacy/Shadow/Weaver 三种执行模式
- **`engine.go`** (407行)：光标移动引擎，定义了 Motion 系统、Buffer 接口和 CursorEngine

#### **架构核心**
- **`architecture_scaffolding.go`**：架构脚手架代码
- **`weaver_scaffolding.go`**：Weaver 系统脚手架
- **`builder.go`**：构建器模式实现
- **`logic.go`**：核心逻辑处理

#### **核心模块**

**FSM 状态机模块 (`fsm/`)**：
- `fsm/engine.go`：状态机引擎核心
- `fsm/keymap.go`：按键映射处理
- `fsm/nvim.go`：Neovim 集成
- `fsm/ui_stub.go`：UI 接口实现

**内核模块 (`kernel/`)**：
- `kernel/kernel.go`：系统内核，协调各组件
- `kernel/decide.go`：决策逻辑
- `kernel/execute.go`：执行引擎
- `kernel/resolver_executor.go`：解析器执行器

**编辑器模块 (`editor/`)**：
- `editor/engine.go`：编辑器引擎
- `editor/execution_context.go`：执行上下文
- `editor/operation.go`：操作定义
- `editor/stores.go`：存储抽象
- `editor/text_object.go`：文本对象处理

**意图系统 (`intent/`)**：
- `intent/intent.go`：意图定义和处理
- `intent/motion.go`：移动意图
- `intent/text_object.go`：文本对象意图
- `intent/range.go`：范围操作意图

**Weaver 系统 (`weaver/`)**：
- `weaver/core/`：Weaver 核心逻辑
- `weaver/manager/`：Weaver 管理器
- `weaver/adapter/`：适配器层

#### **高级功能模块**

**协作编辑 (`crdt/`)**：
- `crdt/crdt.go`：CRDT 数据结构
- `crdt/engine.go`：CRDT 引擎

**撤销/重做系统 (`undotree/`)**：
- `undotree/tree.go`：撤销树结构

**事务管理**：
- `transaction.go`：事务定义和管理
- `undo_redo.go`：撤销/重做实现

**快照系统**：
- `snapshot.go`：状态快照功能

#### **后端和协议**
- **`backend/`**：tmux 命令执行后端
- **`protocol.go`**：通信协议定义
- **`client.go`**：客户端实现

#### **配置和UI**
- **`keymap.yaml`**：按键映射配置文件
- **`ui/`**：用户界面组件
- **`default.tmux.conf`**：默认 tmux 配置

#### **测试和文档**
- **`tests/`**：集成测试和单元测试
- **`docs/`**：详细文档和架构说明
- **`cmd/`**：命令行工具

#### **工具和脚本**
- **`install.sh`**：安装脚本
- **`enter_fsm.sh`** / **`fsm-exit.sh`** / **`fsm-toggle.sh`**：状态切换脚本
- **`validate_paths.sh`**：路径验证脚本

### **技术特点**

1. **多模式执行**：支持 Legacy、Shadow、Weaver 三种模式
2. **事务安全**：完整的事务管理和回滚机制  
3. **意图驱动**：高级用户意图到具体操作的映射
4. **模块化设计**：清晰的组件分离和接口定义
5. **协作友好**：CRDT 支持的多用户编辑
6. **可扩展性**：插件化的架构设计

这是一个功能丰富、架构完善的终端编辑增强工具，旨在提供类似现代编辑器的编辑体验，同时保持终端环境的轻量级特性。

---

## 项目结构图

```
Tmux-FSM/
├── 📁 核心入口
│   ├── main.go (711行) - 程序主入口
│   ├── config.go (69行) - 配置管理
│   └── engine.go (407行) - 光标移动引擎
├── 📁 架构核心
│   ├── architecture_scaffolding.go - 架构脚手架
│   ├── weaver_scaffolding.go - Weaver系统脚手架
│   ├── builder.go - 构建器模式
│   └── logic.go - 核心逻辑
├── 📁 FSM状态机 (fsm/)
│   ├── engine.go - 状态机引擎
│   ├── keymap.go - 按键映射
│   ├── nvim.go - Neovim集成
│   └── ui_stub.go - UI接口
├── 📁 内核 (kernel/)
│   ├── kernel.go - 系统内核
│   ├── decide.go - 决策逻辑
│   ├── execute.go - 执行引擎
│   └── resolver_executor.go - 解析器执行器
├── 📁 编辑器 (editor/)
│   ├── engine.go - 编辑器引擎
│   ├── execution_context.go - 执行上下文
│   ├── operation.go - 操作定义
│   ├── stores.go - 存储抽象
│   └── text_object.go - 文本对象处理
├── 📁 意图系统 (intent/)
│   ├── intent.go - 意图定义
│   ├── motion.go - 移动意图
│   ├── text_object.go - 文本对象意图
│   └── range.go - 范围操作意图
├── 📁 Weaver系统 (weaver/)
│   ├── core/ - 核心逻辑
│   ├── manager/ - 管理器
│   └── adapter/ - 适配器层
├── 📁 高级功能
│   ├── crdt/ - 协作编辑
│   ├── undotree/ - 撤销树
│   ├── transaction.go - 事务管理
│   ├── undo_redo.go - 撤销/重做
│   └── snapshot.go - 状态快照
├── 📁 后端和协议
│   ├── backend/ - tmux命令后端
│   ├── protocol.go - 通信协议
│   └── client.go - 客户端实现
├── 📁 配置和UI
│   ├── keymap.yaml - 按键映射配置
│   ├── ui/ - UI组件
│   └── default.tmux.conf - 默认配置
├── 📁 测试和文档
│   ├── tests/ - 测试套件
│   ├── docs/ - 项目文档
│   └── cmd/ - 命令工具
└── 📁 工具和脚本
    ├── install.sh - 安装脚本
    ├── enter_fsm.sh - 进入FSM
    ├── fsm-exit.sh - 退出FSM
    ├── fsm-toggle.sh - 切换FSM
    └── validate_paths.sh - 路径验证
```

## 核心数据流

```
用户输入 → FSM状态机 → 意图解析 → 内核决策 → 编辑器执行 → 状态更新 → UI刷新
    ↓           ↓         ↓         ↓         ↓         ↓         ↓
  按键事件    状态转换   Intent    Kernel   Editor   State     Tmux
  keymap.yaml  Engine   Builder   Decide   Engine   Sync     Status
```

---

*本文档基于 Tmux-FSM 项目的代码结构和架构分析生成，记录了项目的主要功能和文件作用。*