#!/usr/bin/env bash

# 进入或退出 FSM 模式的静默切换脚本
FSM_ACTIVE=$(tmux show-option -gv @fsm_active)
[ -z "$FSM_ACTIVE" ] && FSM_ACTIVE="false"

if [ "$FSM_ACTIVE" = "true" ]; then
  # 退出逻辑
  tmux set -g @fsm_active "false"
  tmux set -g @fsm_state ""
  tmux set -g @fsm_keys ""
  tmux set -g repeat-time 500
  tmux switch-client -T root
  tmux refresh-client -S
else
  # 进入逻辑：首先强制退出任何既有模式，确保环境纯净
  tmux send-keys -X cancel 2>/dev/null
  tmux set -g @fsm_active "true"
  tmux set -g @fsm_state "NORMAL"
  tmux set -g @fsm_keys ""
  tmux set -g repeat-time 0
  tmux switch-client -T fsm
  tmux refresh-client -S
fi
