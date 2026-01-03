#!/usr/bin/env bash

# Exit FSM + copy-mode safely

tmux set-option -g @fsm_active 0

# exit fsm
tmux send-keys Escape

# exit copy-mode
tmux send-keys q
