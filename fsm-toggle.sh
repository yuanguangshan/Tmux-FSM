#!/usr/bin/env bash
#
# Silent toggle FSM mode (client-level, safe)

FSM_BIN="$HOME/.tmux/plugins/tmux-fsm/tmux-fsm"

FSM_ACTIVE="$(tmux show-option -gv @fsm_active)"
[ -z "$FSM_ACTIVE" ] && FSM_ACTIVE="0"

if [ "$FSM_ACTIVE" = "1" ]; then
  #### EXIT FSM ####

  # 1. Clear FSM flags
  tmux set-option -g @fsm_active 0
  tmux set-option -g @fsm_state ""
  tmux set-option -g @fsm_keys ""

  # 2. Restore repeat-time (optional legacy behavior)
  tmux set-option -g repeat-time 500

  # 3. Exit copy-mode safely (if any)
  tmux copy-mode -q

  # 4. Restore client key-table (核心)
  tmux switch-client -T root

  # 5. Refresh UI
  tmux refresh-client -S

  # 6. Notify FSM runtime (best-effort)
  "$FSM_BIN" -exit 2>/dev/null || true

else
  #### ENTER FSM ####

  # 1. Ensure clean state (client-level only)
  tmux copy-mode -q

  # 2. Set FSM flags
  tmux set-option -g @fsm_active 1
  tmux set-option -g @fsm_state "FSM"
  tmux set-option -g @fsm_keys ""

  # 3. Disable repeat for chord-style FSM
  tmux set-option -g repeat-time 0

  # 4. Switch client key-table (核心)
  tmux switch-client -T fsm

  # 5. Refresh UI
  tmux refresh-client -S

  # 6. Notify FSM runtime (best-effort)
  "$FSM_BIN" -enter 2>/dev/null || true
fi
