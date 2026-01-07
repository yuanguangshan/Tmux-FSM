#!/usr/bin/env bash

FSM_BIN="$HOME/.tmux/plugins/tmux-fsm/tmux-fsm"

# 1. Mark FSM inactive
tmux set-option -g @fsm_active 0
tmux set-option -g @fsm_state ""
tmux set-option -g @fsm_keys ""

# 2. Exit copy-mode safely (if active)
tmux copy-mode -q

# 3. Restore client key-table (核心)
tmux switch-client -T root

# 4. Notify FSM runtime (best-effort)
"$FSM_BIN" -exit 2>/dev/null || true
