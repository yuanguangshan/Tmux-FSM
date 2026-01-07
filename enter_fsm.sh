#!/bin/sh

FSM_BIN="$HOME/.tmux/plugins/tmux-fsm/tmux-fsm"

# 1. 确保不在 copy-mode
tmux copy-mode -q 2>/dev/null || true

# 2. 初始化 FSM 状态
tmux set -g @fsm_state "FSM"
tmux set -g @fsm_keys ""
tmux set -g @fsm_active "1"

# 3. 切换 client key-table（核心）
tmux switch-client -T fsm

# 4. 通知 FSM runtime
"$FSM_BIN" -enter 2>/dev/null || true
