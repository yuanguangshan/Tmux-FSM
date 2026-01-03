# tmux-fsm `plugin.tmux` 代码逻辑详解

该文件是 `tmux-fsm` 插件的入口配置文件，负责定义 tmux 中的状态变量、状态栏显示以及进入/退出 FSM 模式的关键按键绑定。

## 1. 状态变量初始化 (Lines 3-6)

```tmux
set -g @fsm_state ""
set -g @fsm_buffer ""
set -g @fsm_keys ""
```

- **`set -g @variable`**: 使用 tmux 的用户定义选项（以 `@` 开头）作为全局变量。
- **`@fsm_state`**: 存储当前的 FSM 状态名称（如 "NORMAL", "OPERATOR_PENDING"），用于状态栏显示。
- **`@fsm_buffer`**: 用于内部存储命令缓冲区。
- **`@fsm_keys`**: 存储当前已按下的按键序列（如 `d2w`），实时反馈给用户。

## 2. 状态栏集成 (Line 9)

```tmux
set -g status-right "#{@fsm_state}#{@fsm_keys} #S | %Y-%m-%d %H:%M"
```

- 通过在 `status-right` 中引用上述变量，使得 FSM 的当前状态和按键序列能实时显示在 tmux 状态栏的右侧。

## 3. 进入 FSM 模式 (Lines 11-17)

```tmux
bind-key f \
    set -g @fsm_active "true" \; \
    set -g @fsm_state "FSM MODE" \; \
    set -g @fsm_buffer "" \; \
    switch-client -T fsm \; \
    display-message "Enter FSM Mode"
```

- **`bind-key f`**: 在 prefix（默认 `C-b`）后按下 `f` 触发。
- **`switch-client -T fsm`**: **核心逻辑**。将当前的按键表（Key Table）切换为 `fsm`。切换后，所有后续按键都将在 `fsm` 表中查找，而不是默认的 `root` 表。
- 同时初始化相关变量并弹出提示。

## 4. 退出机制 (Lines 19-33)

```tmux
bind-key -T fsm Escape ...
bind-key -T fsm C-c ...
```

- 在 `fsm` 按键表中专门绑定了 `Escape` 和 `Ctrl-c`。
- **`switch-client -T root`**: 将按键表切换回 `root`（默认表），从而实现“退出模式”的效果。
- 退出时会清空所有 FSM 相关的状态变量。

## 5. 极速按键分发 (Lines 35-42)

```tmux
bind-key -T fsm -r Any run-shell -b \
  "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm '#{key}'"
```

- **`-T fsm`**: 指定在 `fsm` 按键表中生效。
- **`Any`**: 捕获所有未显式绑定的按键。
- **`#{key}`**: tmux 占位符，替换为实际按键。
- **Go 客户端**: 所有的按键（除 Esc/C-c 外）都会被传给 `tmux-fsm` 二进制文件。
- **C/S 架构优势**:
    - **逻辑流**: 客户端通过 Unix Socket 将按键秒发给常驻内存的 **Daemon**。
    - **零开销**: 无需初始化 Python 解释器，无需加载重型库，响应时间由原来的 ~50ms 缩短至 **< 1ms**。
    - **状态一致性**: Daemon 内部使用互斥锁 (`sync.Mutex`)，确保高频按键下状态机的原子性。

## 总结

`plugin.tmux` 的核心依然是利用 tmux 的 **Key Table** 机制实现模态切换，但其后端引擎已从 Python 脚本进化为 **高性能 Go 守护进程**。这种架构不仅消除了输入延迟，也为未来实现如宏 (Macro) 等复杂的跨按键功能提供了稳定的状态底座。
