##### ==================================================
##### tmux-fsm plugin
##### ==================================================

# Plugin root
set -g @fsm_root "$HOME/.tmux/plugins/tmux-fsm"
set -g @fsm_bin  "#{@fsm_root}/tmux-fsm"


##### --------------------------------------------------
##### Config (user overridable)
##### --------------------------------------------------

# No-prefix entry key (default C-f)
if -F '#{?@fsm_bind_no_prefix,1,0}' '' \
  'set -g @fsm_bind_no_prefix "C-f"'


##### --------------------------------------------------
##### Internal helpers
##### --------------------------------------------------

# Server (singleton handled inside binary)
run-shell -b "#{@fsm_bin} -server"

# Enter FSM
enter_fsm="run-shell -b \
  'TMUX_FSM_ENTER=1 #{@fsm_bin} enter \"#{pane_id}|#{client_name}\"'"

# Exit FSM
exit_fsm="run-shell -b \
  'TMUX_FSM_EXIT=1 #{@fsm_bin} exit \"#{pane_id}|#{client_name}\"'"


##### --------------------------------------------------
##### Key Bindings
##### --------------------------------------------------

# Entry (no prefix)
bind -n #{@fsm_bind_no_prefix} $enter_fsm

# FSM key table
bind-key -T fsm Escape $exit_fsm
bind-key -T fsm C-c    $exit_fsm
bind-key -T fsm q      $exit_fsm

# Help
bind-key -T fsm ? run-shell -b "#{@fsm_bin} help"

# Universal dispatcher (🔥 核心精简点)
bind-key -T fsm Any run-shell -b \
  "TMUX_FSM_KEY='#{key}' #{@fsm_bin} key '#{key}' '#{pane_id}|#{client_name}'"


##### --------------------------------------------------
##### Status integration
##### --------------------------------------------------

# These variables are updated by tmux-fsm binary
#   @fsm_state  → e.g. [NAV], [CMD]
#   @fsm_keys   → hint keys
#
# tmux.conf only needs to reference them in status-right
#
# Example:
# set -g status-right '#{@fsm_state}#{@fsm_keys} | #S | %H:%M'
