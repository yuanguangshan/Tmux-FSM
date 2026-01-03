- ✅ `README.md`（普通用户，3–5 分钟上手）
- ✅ `DESIGN.md`（FSM / 架构 / 设计哲学）
- ✅ `CONTRIBUTING.md`（开发者指南）

我已经**刻意去掉“论文味 / 作者自嗨”内容**，把信息密度重新分配给真正的目标读者。

---

# ✅ 新版 README.md（面向普通用户）

```markdown
# tmux-fsm

**tmux-fsm** 是一个为 tmux 提供 **Vim 风格操作模式** 的插件。

它让你在 tmux 的 copy-mode / pane 中，使用熟悉的 Vim 按键：
`h j k l`、`dw`、`y2w`、`gg`、寄存器、系统剪贴板等 ——  
**无需复杂的 tmux key table 配置**。

---

## ✨ 功能亮点

- ✅ Vim 风格移动：`h j k l w b e 0 $ gg G`
- ✅ 操作符模式：`d y c` + motion（如 `dw`、`d$`、`y2w`）
- ✅ 数字前缀：`3dw`、`5j`
- ✅ Vim 风格寄存器：`"a yw`、`"b p`
- ✅ 系统剪贴板同步：`s` / `"+`
- ✅ 自动识别 Vim / 非 Vim pane
- ✅ 状态栏实时显示当前模式和按键路径

---

## 📦 安装

### 方法一（推荐）

```bash
git clone https://github.com/yourname/tmux-fsm.git
cd tmux-fsm
./install.sh
```

安装完成后，脚本会提示你是否自动修改 tmux 配置。

---

### 手动安装

1. 复制以下文件到：

```
~/.tmux/plugins/tmux-fsm/
```

- `main.go`
- `logic.go`
- `execute.go`
- `plugin.tmux`

2. 在 tmux 配置文件中添加：

```tmux
source-file "$HOME/.tmux/plugins/tmux-fsm/plugin.tmux"
```

3. 重新加载 tmux：

```bash
tmux source-file ~/.tmux.conf
```

---

## 🚀 使用方法

### 进入 / 退出 FSM 模式

- **进入**：`<prefix> f`
- **退出**：`Esc` 或 `C-c`

> 默认 tmux prefix 是 `C-b`

---

### 常用命令速查

#### 移动

| 按键 | 功能 |
|----|----|
| `h j k l` | 方向移动 |
| `w b e` | 单词移动 |
| `0 $` | 行首 / 行尾 |
| `gg G` | 文件首 / 尾 |

#### 删除 / 复制 / 修改

| 命令 | 说明 |
|----|----|
| `dw` | 删除单词 |
| `d$` | 删除到行尾 |
| `y2w` | 复制 2 个单词 |
| `cw` | 修改单词 |

#### 寄存器

```text
"a yw    # 复制到寄存器 a
"a p     # 从寄存器 a 粘贴
```

#### 系统剪贴板

```text
s        # 默认寄存器 → 系统剪贴板
"a s     # 系统剪贴板 → 寄存器 a
"+yw     # Vim 风格写法
```

---

## 🧠 工作方式（简单理解）

- tmux-fsm 会进入一个 **独立的模式**
- 所有按键都会被一个 **Vim 风格状态机**解析
- 在 Vim pane 中：发送 Vim 命令
- 在普通 pane 中：使用 tmux buffer / paste

你可以把它理解为：  
**“tmux 里的 Vim 操作层”**

---

## 🧯 遇到问题？

```bash
tmux show-messages
```

或查看：

```bash
docs/status_bar_troubleshooting.py
```

---

## 📄 License

MIT
```

---

# ✅ DESIGN.md（FSM / 架构 / 给“看源码的人”）

```markdown
# tmux-fsm 设计说明

本文档描述 tmux-fsm 的 **架构、FSM 设计以及核心取舍**。

如果你只想用插件，请直接看 README。

---

## 1. 设计目标

tmux-fsm 的核心目标只有一个：

> **在 tmux 中复现 Vim 的「操作符 + 动作」语义，而不让 tmux 配置失控**

### 为什么不用 tmux key table 直接写？

- key table 会指数级膨胀
- 无法自然表达 `3dw` / `diw`
- 状态不可视、不可测试

---

## 2. 总体架构

```text
tmux (Key Table)
   ↓
tmux-fsm client  ← 极简二进制
   ↓
tmux-fsm daemon  ← 常驻内存 (Unix Socket)
   ↓
logic.go         ← 纯 FSM (Mutex 保护)
   ↓
Action String    ← 语义结果
   ↓
execute.go       ← 外部 IO 副作用
```

**关键原则：**

- FSM 只产出“语义”
- tmux_fsm.py 负责副作用（tmux / vim / clipboard）

---

## 3. FSM 核心状态

| 状态 | 说明 |
|----|----|
| NORMAL | 默认状态 |
| OPERATOR_PENDING | 等待 motion（d / y / c） |
| MOTION_PENDING | `gg` / `f{char}` |
| FIND_PENDING | 等待查找字符 |
| MODIFIER | `a` / `i` |
| REGISTER_SELECT | `"` 后选择寄存器 |

FSM 入口只有一个：

```python
press_key(key) -> Optional[action]
```

---

## 4. Vim 语义模型

FSM 实现了经典 Vim 语法：

```text
[Register][Count][Operator][Modifier][Motion]
```

示例：

- `dw`
- `"a y2w`
- `gg`
- `f x`

---

## 5. 状态持久化策略

由于 tmux `run-shell` **每次都是新进程**：

- FSM 状态序列化为 JSON
- 存入 tmux 全局 option
- 每次运行前 load / 结束后 save

tmux 本身充当“内存”

---

## 6. Vim 集成策略（Source of Truth）

- **tmux buffer 是唯一真实来源**
- Vim 寄存器只是镜像
- 所有跨 pane / clipboard 行为最终落在 tmux buffer

---

## 7. 为什么 FSM 是正确模型？

- 状态显式
- 路径可推理
- 行为可测试
- 新命令 = 新转移，而不是新 key table

---

## 8. 已知限制

- 已实现 visual mode (v / V)
- 内置撤销重做堆栈 (Undo/Redo Stack)
- 线程安全的状态更新

这些都可以在 FSM 层自然扩展。
```

---

# ✅ CONTRIBUTING.md（开发者）

```markdown
# Contributing to tmux-fsm

欢迎贡献 tmux-fsm 🎉  
本项目非常适合 **FSM / 编辑器语义 / tmux 高级玩法** 爱好者。

---

## 1. 代码结构

```text
fsm.py         # 纯 FSM（无副作用）
tmux_fsm.py   # tmux / vim / clipboard glue
plugin.tmux   # tmux 入口
docs/          # 测试 & 文档
```

**黄金法则：**

> ❌ 不要在 fsm.py 中调用 tmux / shell  
> ✅ fsm.py 只能返回 action string

---

## 2. 开发流程

### 本地测试 FSM

```bash
python docs/test_fsm_functionality.py
```

### 端到端测试（需要 tmux）

```bash
python docs/test_end_to_end.py
```

---

## 3. 添加新命令的正确方式

### ✅ 推荐

1. 在 `fsm.py` 中：
   - 添加输入分类
   - 添加状态转移
   - 返回 action string

2. 在 `tmux_fsm.py` 中：
   - 解析 action
   - 实现副作用

### ❌ 不推荐

- 直接在 tmux key table 写逻辑
- 在 FSM 中做 IO / 剪贴板操作

---

## 4. 常见扩展方向

- visual / visual-line mode
- text object（iw / aw / ip）
- repeat (`.`)
- macro recording
- 更智能的 Vim detection

---

## 5. 风格建议

- FSM 逻辑优先清晰而非“聪明”
- 可读性 > 抽象层数
- 新状态必须可测试

---

## 6. 提交规范

- 小 PR
- 一个 PR 只做一件事
- 带测试更佳

---

感谢你让 tmux 更像一个编辑器 ❤️
```

---
