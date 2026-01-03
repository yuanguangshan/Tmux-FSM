##### tmux-fsm plugin (FOEK Kernel v1.0 - Silent Stealth) #####

# 1. 变量初始化
set -g @fsm_state ""
set -g @fsm_keys ""

# 2. 状态栏配置
set -g status-right "#{@fsm_state}#{@fsm_keys} | #S | %Y-%m-%d %H:%M"

# 3. 入口：支持自定义按键 (Prefix 和 No-Prefix)
# 使用 run-shell 动态绑定
run-shell "
    # 1. 绑定 Prefix + Key (Default: f)
    prefix_key=\$(tmux show-option -gqv @fsm_toggle_key)
    [ -z \"\$prefix_key\" ] && prefix_key=\"f\"
    tmux bind-key \"\$prefix_key\" run-shell -b \"\$HOME/.tmux/plugins/tmux-fsm/enter_fsm.sh\"

    # 2. 绑定 No-Prefix Key (Root Table)
    root_key=\$(tmux show-option -gqv @fsm_bind_no_prefix)
    if [ -n \"\$root_key\" ]; then
        tmux bind-key -n \"\$root_key\" run-shell -b \"\$HOME/.tmux/plugins/tmux-fsm/enter_fsm.sh\"
    fi

    # 3. 直接帮助入口 (方便手机操作)
    # Prefix + h
    tmux bind-key h run-shell \"\$HOME/.tmux/plugins/tmux-fsm/tmux-fsm '__HELP__' '#{pane_id}|#{client_name}'\"
    # Alt + ? (无需进入 FSM)
    tmux bind-key -n M-? run-shell \"\$HOME/.tmux/plugins/tmux-fsm/tmux-fsm '__HELP__' '#{pane_id}|#{client_name}'\"
"

# 4. 退出逻辑 (保持不变, 但为了对称性，可以考虑 extract later, but simple here)

# 4. 退出逻辑
bind-key -T fsm Escape \
    set -g @fsm_active "false" \; \
    set -g repeat-time 500 \; \
    switch-client -T root \; \
    run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm '__CLEAR_STATE__'" \; \
    refresh-client -S

bind-key -T fsm C-c \
    set -g @fsm_active "false" \; \
    set -g repeat-time 500 \; \
    switch-client -T root \; \
    run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm '__CLEAR_STATE__' '#{pane_id}|#{client_name}'" \; \
    refresh-client -S

# 5. 单入口键：继续使用后台模式，确保绝对静默
bind-key -T fsm 0 run-shell -b "TMUX_FSM_KEY='0' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '0' '#{pane_id}|#{client_name}'"
bind-key -T fsm 1 run-shell -b "TMUX_FSM_KEY='1' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '1' '#{pane_id}|#{client_name}'"
bind-key -T fsm 2 run-shell -b "TMUX_FSM_KEY='2' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '2' '#{pane_id}|#{client_name}'"
bind-key -T fsm 3 run-shell -b "TMUX_FSM_KEY='3' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '3' '#{pane_id}|#{client_name}'"
bind-key -T fsm 4 run-shell -b "TMUX_FSM_KEY='4' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '4' '#{pane_id}|#{client_name}'"
bind-key -T fsm 5 run-shell -b "TMUX_FSM_KEY='5' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '5' '#{pane_id}|#{client_name}'"
bind-key -T fsm 6 run-shell -b "TMUX_FSM_KEY='6' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '6' '#{pane_id}|#{client_name}'"
bind-key -T fsm 7 run-shell -b "TMUX_FSM_KEY='7' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '7' '#{pane_id}|#{client_name}'"
bind-key -T fsm 8 run-shell -b "TMUX_FSM_KEY='8' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '8' '#{pane_id}|#{client_name}'"
bind-key -T fsm 9 run-shell -b "TMUX_FSM_KEY='9' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '9' '#{pane_id}|#{client_name}'"

# 添加字母键绑定
bind-key -T fsm a run-shell -b "TMUX_FSM_KEY='a' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'a' '#{pane_id}|#{client_name}'"
bind-key -T fsm b run-shell -b "TMUX_FSM_KEY='b' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'b' '#{pane_id}|#{client_name}'"
bind-key -T fsm c run-shell -b "TMUX_FSM_KEY='c' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'c' '#{pane_id}|#{client_name}'"
bind-key -T fsm d run-shell -b "TMUX_FSM_KEY='d' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'd' '#{pane_id}|#{client_name}'"
bind-key -T fsm e run-shell -b "TMUX_FSM_KEY='e' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'e' '#{pane_id}|#{client_name}'"
bind-key -T fsm f run-shell -b "TMUX_FSM_KEY='f' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'f' '#{pane_id}|#{client_name}'"
bind-key -T fsm g run-shell -b "TMUX_FSM_KEY='g' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'g' '#{pane_id}|#{client_name}'"
bind-key -T fsm h run-shell -b "TMUX_FSM_KEY='h' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'h' '#{pane_id}|#{client_name}'"
bind-key -T fsm i run-shell -b "TMUX_FSM_KEY='i' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'i' '#{pane_id}|#{client_name}'"
bind-key -T fsm j run-shell -b "TMUX_FSM_KEY='j' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'j' '#{pane_id}|#{client_name}'"
bind-key -T fsm k run-shell -b "TMUX_FSM_KEY='k' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'k' '#{pane_id}|#{client_name}'"
bind-key -T fsm l run-shell -b "TMUX_FSM_KEY='l' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'l' '#{pane_id}|#{client_name}'"
bind-key -T fsm m run-shell -b "TMUX_FSM_KEY='m' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'm' '#{pane_id}|#{client_name}'"
bind-key -T fsm n run-shell -b "TMUX_FSM_KEY='n' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'n' '#{pane_id}|#{client_name}'"
bind-key -T fsm o run-shell -b "TMUX_FSM_KEY='o' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'o' '#{pane_id}|#{client_name}'"
bind-key -T fsm p run-shell -b "TMUX_FSM_KEY='p' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'p' '#{pane_id}|#{client_name}'"
bind-key -T fsm q run-shell -b "TMUX_FSM_KEY='q' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'q' '#{pane_id}|#{client_name}'"
bind-key -T fsm r run-shell -b "TMUX_FSM_KEY='r' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'r' '#{pane_id}|#{client_name}'"
bind-key -T fsm s run-shell -b "TMUX_FSM_KEY='s' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 's' '#{pane_id}|#{client_name}'"
bind-key -T fsm t run-shell -b "TMUX_FSM_KEY='t' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 't' '#{pane_id}|#{client_name}'"
bind-key -T fsm u run-shell -b "TMUX_FSM_KEY='u' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'u' '#{pane_id}|#{client_name}'"
bind-key -T fsm v run-shell -b "TMUX_FSM_KEY='v' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'v' '#{pane_id}|#{client_name}'"
bind-key -T fsm w run-shell -b "TMUX_FSM_KEY='w' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'w' '#{pane_id}|#{client_name}'"
bind-key -T fsm x run-shell -b "TMUX_FSM_KEY='x' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'x' '#{pane_id}|#{client_name}'"
bind-key -T fsm y run-shell -b "TMUX_FSM_KEY='y' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'y' '#{pane_id}|#{client_name}'"
bind-key -T fsm z run-shell -b "TMUX_FSM_KEY='z' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'z' '#{pane_id}|#{client_name}'"

bind-key -T fsm A run-shell -b "TMUX_FSM_KEY='A' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'A' '#{pane_id}|#{client_name}'"
bind-key -T fsm B run-shell -b "TMUX_FSM_KEY='B' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'B' '#{pane_id}|#{client_name}'"
bind-key -T fsm C run-shell -b "TMUX_FSM_KEY='C' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'C' '#{pane_id}|#{client_name}'"
bind-key -T fsm D run-shell -b "TMUX_FSM_KEY='D' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'D' '#{pane_id}|#{client_name}'"
bind-key -T fsm E run-shell -b "TMUX_FSM_KEY='E' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'E' '#{pane_id}|#{client_name}'"
bind-key -T fsm F run-shell -b "TMUX_FSM_KEY='F' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'F' '#{pane_id}|#{client_name}'"
bind-key -T fsm G run-shell -b "TMUX_FSM_KEY='G' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'G' '#{pane_id}|#{client_name}'"
bind-key -T fsm H run-shell -b "TMUX_FSM_KEY='H' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'H' '#{pane_id}|#{client_name}'"
bind-key -T fsm I run-shell -b "TMUX_FSM_KEY='I' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'I' '#{pane_id}|#{client_name}'"
bind-key -T fsm J run-shell -b "TMUX_FSM_KEY='J' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'J' '#{pane_id}|#{client_name}'"
bind-key -T fsm K run-shell -b "TMUX_FSM_KEY='K' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'K' '#{pane_id}|#{client_name}'"
bind-key -T fsm L run-shell -b "TMUX_FSM_KEY='L' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'L' '#{pane_id}|#{client_name}'"
bind-key -T fsm M run-shell -b "TMUX_FSM_KEY='M' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'M' '#{pane_id}|#{client_name}'"
bind-key -T fsm N run-shell -b "TMUX_FSM_KEY='N' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'N' '#{pane_id}|#{client_name}'"
bind-key -T fsm O run-shell -b "TMUX_FSM_KEY='O' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'O' '#{pane_id}|#{client_name}'"
bind-key -T fsm P run-shell -b "TMUX_FSM_KEY='P' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'P' '#{pane_id}|#{client_name}'"
bind-key -T fsm Q run-shell -b "TMUX_FSM_KEY='Q' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'Q' '#{pane_id}|#{client_name}'"
bind-key -T fsm R run-shell -b "TMUX_FSM_KEY='R' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'R' '#{pane_id}|#{client_name}'"
bind-key -T fsm S run-shell -b "TMUX_FSM_KEY='S' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'S' '#{pane_id}|#{client_name}'"
bind-key -T fsm T run-shell -b "TMUX_FSM_KEY='T' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'T' '#{pane_id}|#{client_name}'"
bind-key -T fsm U run-shell -b "TMUX_FSM_KEY='U' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'U' '#{pane_id}|#{client_name}'"
bind-key -T fsm V run-shell -b "TMUX_FSM_KEY='V' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'V' '#{pane_id}|#{client_name}'"
bind-key -T fsm W run-shell -b "TMUX_FSM_KEY='W' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'W' '#{pane_id}|#{client_name}'"
bind-key -T fsm X run-shell -b "TMUX_FSM_KEY='X' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'X' '#{pane_id}|#{client_name}'"
bind-key -T fsm Y run-shell -b "TMUX_FSM_KEY='Y' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'Y' '#{pane_id}|#{client_name}'"
bind-key -T fsm Z run-shell -b "TMUX_FSM_KEY='Z' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'Z' '#{pane_id}|#{client_name}'"

# 添加其他常用键绑定
bind-key -T fsm - run-shell -b "TMUX_FSM_KEY='-' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '-' '#{pane_id}|#{client_name}'"
bind-key -T fsm = run-shell -b "TMUX_FSM_KEY='=' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '=' '#{pane_id}|#{client_name}'"
bind-key -T fsm $ run-shell -b "TMUX_FSM_KEY='$' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '$' '#{pane_id}|#{client_name}'"
bind-key -T fsm ^ run-shell -b "TMUX_FSM_KEY='^' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '^' '#{pane_id}|#{client_name}'"
bind-key -T fsm G run-shell -b "TMUX_FSM_KEY='G' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'G' '#{pane_id}|#{client_name}'"

# 控制键绑定 (PageUp/Down/Left/Right 映射)
bind-key -T fsm C-b run-shell -b "TMUX_FSM_KEY='C-b' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'C-b' '#{pane_id}|#{client_name}'"
bind-key -T fsm C-f run-shell -b "TMUX_FSM_KEY='C-f' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm 'C-f' '#{pane_id}|#{client_name}'"

# 补充符号键
bind-key -T fsm . run-shell -b "TMUX_FSM_KEY='.' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '.' '#{pane_id}|#{client_name}'"
bind-key -T fsm ? run-shell "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm '__HELP__' '#{pane_id}|#{client_name}'"
bind-key -T fsm / run-shell -b "TMUX_FSM_KEY='/' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '/' '#{pane_id}|#{client_name}'"

# 通配符 (Catch-all) - 防止未绑定的键穿透到 Shell
# Payload Format: "PANE_ID|CLIENT_NAME" passed as arg2
bind-key -T fsm Any run-shell -b \
  "TMUX_FSM_KEY='#{key}' $HOME/.tmux/plugins/tmux-fsm/tmux-fsm '#{key}' '#{pane_id}|#{client_name}'"

##### end tmux-fsm #####

# 6. 预拉起服务端 (Pre-start) - 保证 Socket 随时就绪，消除冷启动延迟
run-shell -b "$HOME/.tmux/plugins/tmux-fsm/tmux-fsm -server >/dev/null 2>&1 || true"
