##### tmux-fsm plugin (New Architecture with Legacy Support) #####

# 1. 变量初始化
set -g @fsm_state ""
set -g @fsm_keys ""

# 2. 状态栏配置
set -g status-right "#[fg=yellow,bold]#{@fsm_state}#{@fsm_keys}#[default] | #S | %m-%d %H:%M"

# 3. 获取插件路径 (使用 TPM 标准路径)
set -g @fsm_bin "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm"

# 4. 入口：支持自定义按键 (Prefix 和 No-Prefix)
# 使用 run-shell 动态绑定
run-shell "
    # 1. 获取当前的 prefix key
    current_prefix=\$(tmux show-option -gqv prefix)

    # 2. 绑定 Prefix + Key (Default: f) - works for both C-b and C-a
    prefix_key=\$(tmux show-option -gqv @fsm_toggle_key)
    [ -z \"\$prefix_key\" ] && prefix_key=\"f\"
    tmux bind-key \"\$prefix_key\" run-shell -b '$HOME/.tmux/plugins/tmux-fsm/enter_fsm.sh'

    # 3. 绑定 No-Prefix Key (Root Table)
    root_key=\$(tmux show-option -gqv @fsm_bind_no_prefix)
    if [ -n \"\$root_key\" ]; then
        tmux bind-key -n \"\$root_key\" run-shell -b '$HOME/.tmux/plugins/tmux-fsm/enter_fsm.sh'
    fi

    # 4. 添加 Ctrl+F 绑定作为额外选项（无论当前prefix是什么）
    tmux bind-key -n C-f run-shell -b '$HOME/.tmux/plugins/tmux-fsm/enter_fsm.sh'

    # 5. 设置全局环境变量 (Phase 7: Temporal Integrity)
    tmux set-environment -g TMUX_FSM_MODE weaver
    tmux set-environment -g TMUX_FSM_LOG_FACTS 1

    # 6. 启动服务器守护进程 (Weaver Mode)
    TMUX_FSM_MODE=weaver TMUX_FSM_LOG_FACTS=1 $HOME/.tmux/plugins/tmux-fsm/tmux-fsm -server >/dev/null 2>&1 &
"

# 5. FSM 键表配置 (新架构)
bind-key -T fsm -n C-c run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -exit"
bind-key -T fsm -n Escape run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -exit"

# 6. Explicitly bind alphanumeric keys (POSIX compliant)
# {a..z} is a bash extension, we must use explicit lists for /bin/sh compatibility
run-shell "
    for key in a b c d e f g h i j k l m n o p q r s t u v w x y z A B C D E F G H I J K L M N O P Q R S T U V W X Y Z 0 1 2 3 4 5 6 7 8 9 '$' '^' '.' '/' ',' ';' ':'; do
        tmux bind-key -T fsm \"\$key\" run-shell -b \"$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -key '\$key' '#{pane_id}|#{client_name}'\"
    done
"

# 7. Bind common punctuation explicitly - REMOVED due to shell escaping issues. 
# Relying on 'Any' fallback for punctuation.

# Keep 'Any' as a fallback for special keys and punctuation.
bind-key -T fsm Any run-shell -b \
  "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -key \"#{key}\" \"#{pane_id}|#{client_name}\""

# 7. 额外的便捷键绑定
bind-key -T fsm q run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -exit"

# 8. 重新加载配置
bind-key -T root R run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -reload"

# 9. 帮助功能
bind-key -T root ? run-shell "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm '__HELP__' '#{pane_id}|#{client_name}'"

##### end tmux-fsm #####
