#!/bin/bash
PLUGIN_DIR="$HOME/.tmux/plugins/tmux-fsm"
FSM_BIN="$PLUGIN_DIR/tmux-fsm"

# 1. Start Server (silent)
"$FSM_BIN" -server >/dev/null 2>&1 &

# 2. Cancel copy mode (twice to be sure)
tmux send-keys -X cancel 2>/dev/null || true
tmux send-keys -X cancel 2>/dev/null || true

# 3. Set vars
tmux set -g @fsm_active "true"
tmux set -g repeat-time 0

# 4. Switch key table
tmux switch-client -T fsm

# 5. Init state (Ensure clean buffer)
"$FSM_BIN" "__CLEAR_STATE__" "#{pane_id}|#{client_name}"

# 6. Refresh
tmux refresh-client -S
