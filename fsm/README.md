
# FSM (Finite State Machine) Module Documentation

> 本文档**严格对应当前 `fsm` 包的代码实现**（共 5 个文件），  
> 描述的是一个 **键驱动的、层级化的 FSM 引擎**，用于在 tmux / nvim 环境中
> 捕获按键、维护状态、并向外部系统发射 token。

---

## 模块职责边界（非常重要）

### FSM **负责**

- ✅ 键输入 → 状态转移
- ✅ 数字计数（Vim 风格）
- ✅ Layer（模式层）管理
- ✅ RawToken 发射
- ✅ UI 状态派生（当前为临时实现）

### FSM **不负责**

- ❌ 文本编辑
- ❌ 动作语义解析（Intent Resolver 已废弃）
- ❌ Buffer / Cursor 计算
- ❌ 真正的 tmux / nvim 执行（当前存在技术债桥接）

---

## 核心对象关系

```
Key Input
   ↓
Engine.Dispatch(key)
   ↓
FSM State Transition
   ↓
RawToken Emission
   ↓
(UI / Recorder / Debugger)
```

---

## `token.go` —— **最小输入语义单元**

### RawTokenKind

```go
type RawTokenKind int
```

已定义的 token 类型：

| Token | 含义 |
|----|----|
| TokenDigit | 数字计数输入（如 3、42） |
| TokenKey | 普通按键 |
| TokenRepeat | 重复命令（`.`） |
| TokenSystem | 系统事件（enter / exit / reset） |

---

### RawToken

```go
type RawToken struct {
    Kind  RawTokenKind
    Value string
}
```

- FSM 对外的**唯一事件输出格式**
- 不携带语义，只携带**事实**

---

## `engine.go` —— **FSM 引擎核心**

---

### RawTokenEmitter（输出接口）

```go
type RawTokenEmitter interface {
    Emit(RawToken)
}
```

- FSM **不关心 token 去哪**
- 可以有多个 emitter（UI、Recorder、Debugger）

---

### Engine 结构

```go
type Engine struct {
    Active     string
    Keymap     *Keymap
    layerTimer *time.Timer
    count      int
    emitters   []RawTokenEmitter
    visualMode intent.VisualMode
}
```

字段语义：

| 字段 | 说明 |
|----|----|
| Active | 当前 FSM 层（state） |
| Keymap | 状态机定义 |
| layerTimer | 层超时自动 reset |
| count | 数字前缀（Vim 风格） |
| emitters | RawToken 订阅者 |
| visualMode | 当前可视模式（仅记录，不驱动行为） |

---

### Engine 生命周期

#### 创建

```go
func NewEngine(km *Keymap) *Engine
```

- 初始层为 `"NAV"`
- 不自动启动
- 不注册 UI

#### 全局实例

```go
var defaultEngine *Engine
```

通过：

```go
InitEngine(km)
GetDefaultEngine()
```

管理

---

### Dispatch —— **FSM 的核心入口**

```go
func (e *Engine) Dispatch(key string) bool
```

处理顺序（**严格按代码顺序**）：

#### 1️⃣ 数字计数

- 任意层都接受数字
- `0` 在 count == 0 时视为普通键
- 其他数字累积到 `count`
- 发射 `TokenDigit`

#### 2️⃣ 重复键

```go
key == "."
```

- 发射 `TokenRepeat`
- 不改变 FSM 状态

#### 3️⃣ Keymap 匹配

- 只在当前 `Active` 层查找
- 如果匹配：

##### a. Layer 切换

```go
KeyAction.Layer != ""
```

- 切换 `Active`
- 启动超时（如配置）
- 发射 `TokenKey`

##### b. 普通按键

- 不执行 action
- 只发射 `TokenKey`

#### 4️⃣ 未处理

返回 `false`

---

### 数字计数规则（实现事实）

- FSM **只记录数字**
- FSM **不消费数字**
- `count` 只影响 UI & token 流
- 动作层如何使用 count 不属于 FSM

---

### Reset / Reload

```go
func (e *Engine) Reset()
```

行为：

- 停止 layerTimer
- 回到 initial 或 NAV
- 清空 count
- 发射 `TokenSystem("reset")`

---

```go
func Reload(configPath string) error
```

- 重新加载 Keymap
- 重建 Engine
- Reset FSM
- 更新 UI

---

### RunAction（tmux 动作桥接）

```go
func (e *Engine) RunAction(name string)
```

- **硬编码动作名**
- 直接映射到 tmux 命令
- 这是一个**过渡期实现**
- FSM 本身并不理解这些动作

---

### EnterFSM / ExitFSM

#### EnterFSM

- 初始化引擎
- Reset 到 NAV
- 发射 `TokenSystem("enter")`
- 更新 UI

#### ExitFSM

- Reset
- 发射 `TokenSystem("exit")`
- 隐藏 UI

---

## `keymap.go` —— **FSM 定义数据结构**

---

### Keymap

```go
type Keymap struct {
    Initial string
    States  map[string]StateDef
}
```

- `Initial`：初始层名
- `States`：FSM 的所有状态

---

### StateDef

```go
type StateDef struct {
    Hint   string
    Sticky bool
    Keys   map[string]KeyAction
}
```

当前 FSM **只使用 Keys**

- `Hint` / `Sticky` 尚未被 Engine 使用

---

### KeyAction

```go
type KeyAction struct {
    Action    string
    Layer     string
    TimeoutMs int
}
```

FSM **只关心**：

- `Layer`
- `TimeoutMs`

`Action` 不在 FSM 中执行，只用于上层。

---

### Validate

```go
func (km *Keymap) Validate() error
```

唯一校验规则：

- 所有 `Layer` 引用必须存在

---

## `ui_stub.go` —— **UI 派生状态（临时桥接）**

> ⚠️ 本文件明确标注为 **技术债实现**

---

### UI 不变量（写在代码里的）

> **Invariant 9: UI 是 FSM 派生状态**

---

### UpdateUI

```go
func UpdateUI(_ ...any)
```

当前行为：

1. **直接操作 tmux**
2. 设置：
   - `@fsm_state`
   - `@fsm_keys`
3. 刷新 tmux client
4. 调用 `OnUpdateUI` 回调

---

### HideUI

- 清空 tmux 变量
- 刷新 client

---

### UIDriver（未使用）

```go
type UIDriver interface {
    SetUserOption(...)
    RefreshClient(...)
}
```

当前代码 **未使用此接口**

---

## `nvim.go` —— **Neovim 模式联动**

---

### OnNvimMode

```go
func OnNvimMode(mode string)
```

规则：

- 当 nvim 进入：
  - insert
  - visual
  - select
- FSM **立即 Exit**

FSM **不尝试同步 nvim 状态**

---

### NotifyNvimMode

- 空实现
- 明确声明应由 Kernel / Weaver 处理

---

## 当前 FSM 的真实能力总结

✅ **已实现**

- 层级 FSM
- 数字计数
- RawToken 流
- 超时自动 reset
- tmux UI 状态展示（临时）

❌ **未实现**

- Intent 解析
- 动作语义
- Buffer / Motion
- 可逆性
- 历史记录

---

## 一句话结论

> **这是一个“键 → 状态 → token”的纯 FSM 核心，**
> 它刻意不理解编辑语义，只保证：
>
> - 输入是确定的  
> - 状态是可预测的  
> - 输出是可订阅的  

---
