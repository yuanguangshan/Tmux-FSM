#!/bin/bash
PLUGIN_DIR="$HOME/.tmux/plugins/tmux-fsm"
FSM_BIN="$PLUGIN_DIR/tmux-fsm"

# 1. 强制退出 copy-mode（安全）
tmux copy-mode -q 2>/dev/null || true

# 2. Init FSM vars
tmux set -g @fsm_active "true"
tmux set -g @fsm_state "FSM"
tmux set -g @fsm_keys ""
tmux set -g repeat-time 0

# 3. Switch key table（关键）
tmux switch-client -T fsm

# 4. Enter FSM logic
"$FSM_BIN" -enter

# 5. Refresh
tmux refresh-client -S
