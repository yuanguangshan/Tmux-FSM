##### tmux-fsm plugin (New Architecture with Legacy Support) #####

##### 1. 变量初始化 #####
set -g @fsm_state ""
set -g @fsm_keys ""

##### 2. 状态栏配置 #####
set -g status-right "#[fg=yellow,bold]#{@fsm_state}#{@fsm_keys}#[default] | #S | %m-%d %H:%M"

##### 3. 插件路径 #####
set -g @fsm_bin "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm"

##### 4. FSM 入口（静态绑定，声明式） #####
# Prefix + f
bind-key f run-shell -b "$HOME/.tmux/plugins/tmux-fsm/enter_fsm.sh"

# No-prefix Ctrl+f
bind-key -n C-f run-shell -b "$HOME/.tmux/plugins/tmux-fsm/enter_fsm.sh"

##### 5. FSM 键表：安全退出（先退表，再通知 runtime） #####
bind-key -T fsm Escape switch-client -T root \; \
  run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -exit"

bind-key -T fsm C-c switch-client -T root \; \
  run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -exit"

bind-key -T fsm q switch-client -T root \; \
  run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -exit"

##### 6. 显式绑定字母 / 数字（POSIX 兼容） #####
run-shell "
for key in \
  a b c d e f g h i j k l m n o p q r s t u v w x y z \
  A B C D E F G H I J K L M N O P Q R S T U V W X Y Z \
  0 1 2 3 4 5 6 7 8 9; do
    tmux bind-key -T fsm \"\$key\" \
      run-shell -b \"$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -key '\$key' '#{pane_id}|#{client_name}'\"
done
"

##### 7. Any fallback（兜底所有特殊键 / 标点） #####
bind-key -T fsm Any run-shell -b \
  "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -key \"#{key}\" \"#{pane_id}|#{client_name}\""

##### 8. 重新加载 FSM（不影响 client 表） #####
bind-key -T root R run-shell -b \
  "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -reload"

##### 9. Help #####
bind-key -T root ? run-shell \
  "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm '__HELP__' '#{pane_id}|#{client_name}'"

##### 10. 启动 FSM Server（一次性，后台） #####
run-shell -b "
TMUX_FSM_MODE=weaver TMUX_FSM_LOG_FACTS=1 \
$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -server >/dev/null 2>&1 &
"

##### end tmux-fsm #####