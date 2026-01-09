# ui 模块

## 模块职责概述

`ui/` 是 **Tmux-FSM 的用户界面抽象层**，负责提供统一的用户界面接口和状态展示功能。该模块实现了界面的抽象和解耦，支持多种界面后端和状态可视化，为用户提供直观的状态反馈和交互提示。

主要职责包括：
- 提供统一的界面抽象接口
- 管理状态显示和用户提示
- 处理用户界面的输入输出
- 支持多种界面后端的适配

## 核心设计思想

- **界面抽象**: 提供统一的界面接口，解耦具体实现
- **状态可视化**: 直观展示系统状态和操作反馈
- **多后端支持**: 支持不同的界面渲染后端
- **响应式更新**: 实时响应状态变化

## 文件结构说明

### `ui.go`
- 核心界面接口定义
- 主要结构体：
  - `UI`: 用户界面接口
  - `DisplayState`: 显示状态
  - `UIEvent`: 界面事件
- 主要函数：
  - `Render(state DisplayState) error`: 渲染界面状态
  - `ShowHint(hint string) error`: 显示提示信息
  - `UpdateStatus(status StatusInfo) error`: 更新状态信息
  - `HandleInput() (UIEvent, error)`: 处理用户输入
- 定义界面的核心接口和抽象

### `renderer.go`
- 界面渲染器
- 主要函数：
  - `NewRenderer(backend Backend) Renderer`: 创建渲染器
  - `DrawText(x, y int, text string) error`: 绘制文本
  - `DrawBox(x, y, width, height int) error`: 绘制框
  - `ClearScreen() error`: 清屏
  - `Refresh() error`: 刷新显示
- 负责具体的界面绘制操作

### `status_bar.go`
- 状态栏管理器
- 主要函数：
  - `UpdateStatusBar(info StatusInfo) error`: 更新状态栏
  - `ShowMode(mode string) error`: 显示当前模式
  - `ShowPosition(pos Position) error`: 显示当前位置
  - `ShowMessage(msg string, duration time.Duration) error`: 显示消息
- 管理状态栏的信息显示

### `input_handler.go`
- 输入处理器
- 主要函数：
  - `ProcessInput(input InputEvent) UIEvent`: 处理输入事件
  - `MapKey(key KeyCode) Command`: 键映射
  - `HandleMouse(event MouseEvent) UIEvent`: 处理鼠标事件
  - `GetInput() InputEvent`: 获取用户输入
- 处理用户界面输入

### `theme.go`
- 主题管理器
- 主要结构体：
  - `Theme`: 主题定义
  - `ColorScheme`: 颜色方案
- 主要函数：
  - `LoadTheme(path string) (*Theme, error)`: 加载主题
  - `ApplyTheme(theme *Theme) error`: 应用主题
  - `SetColor(foreground, background Color) error`: 设置颜色
- 管理界面的主题和样式

## 界面特性

### 状态显示
- 实时显示当前编辑模式
- 显示光标位置和文档状态
- 提供操作反馈和提示信息

### 多后端支持
- 支持终端界面后端
- 支持图形界面后端
- 支持 Web 界面后端

### 响应式更新
- 自动响应状态变化
- 高效的增量更新
- 平滑的界面过渡效果

## 在整体架构中的角色

UI 模块是系统的用户交互前端，它为用户提供直观的状态反馈和操作界面。通过界面抽象，UI 模块实现了：
- 界面与核心逻辑的解耦
- 多种界面后端的支持
- 实时的状态可视化
- 用户友好的交互体验