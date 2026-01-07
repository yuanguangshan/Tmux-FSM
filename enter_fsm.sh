#!/bin/bash
PLUGIN_DIR="$HOME/.tmux/plugins/tmux-fsm"
FSM_BIN="$PLUGIN_DIR/tmux-fsm"

# 1. Cancel copy mode (twice to be sure)
tmux send-keys -X cancel 2>/dev/null || true
tmux send-keys -X cancel 2>/dev/null || true

# 2. Set vars - 初始化状态变量
tmux set -g @fsm_active "true"
tmux set -g @fsm_state "FSM"  # 设置初始状态显示
tmux set -g @fsm_keys ""
tmux set -g repeat-time 0

# 3. Switch key table
tmux switch-client -T fsm

# 4. Init state
# Call -enter without parameters. The Go binary will handle server startup if needed.
"$FSM_BIN" -enter

# 5. Refresh
tmux refresh-client -S
