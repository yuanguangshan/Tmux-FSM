# Project Documentation

- **Generated at:** 2026-01-07 21:06:30
- **Root Dir:** `.`
- **File Count:** 10
- **Total Size:** 6.09 KB

## 📂 File List
- `core.conf` (0.95 KB)
- `fsm.conf` (0.22 KB)
- `fsm/core.conf` (1.44 KB)
- `fsm/motion.conf` (0.48 KB)
- `fsm/operator.conf` (0.84 KB)
- `fsm/textobj.conf` (0.37 KB)
- `fsm/visual.conf` (0.61 KB)
- `motion.conf` (0.69 KB)
- `operator.conf` (0.26 KB)
- `visual.conf` (0.20 KB)

---

## 📄 `core.conf`

````conf



# ======================
# CORE STATE
# ======================

# 回到 normal（fsm）态
bind-key -T root Escape switch-client -T fsm
bind-key -T fsm Escape switch-client -T fsm

# 数字前缀（count）
bind-key -T fsm 1 set-option -g @fsm_count 1
bind-key -T fsm 2 set-option -g @fsm_count 2
bind-key -T fsm 3 set-option -g @fsm_count 3
bind-key -T fsm 4 set-option -g @fsm_count 4
bind-key -T fsm 5 set-option -g @fsm_count 5
bind-key -T fsm 6 set-option -g @fsm_count 6
bind-key -T fsm 7 set-option -g @fsm_count 7
bind-key -T fsm 8 set-option -g @fsm_count 8
bind-key -T fsm 9 set-option -g @fsm_count 9

# ======================
# DOT REPEAT
# ======================

bind-key -T fsm . \
  run-shell '
    action=$(tmux display -p "#{@fsm_last_action}")
    count=$(tmux display -p "#{@fsm_count}")
    [ -z "$count" ] && count=1
    [ -z "$action" ] && exit 0

    for i in $(seq 1 "$count"); do
      tmux $action
    done
  ' \;\
  set-option -g @fsm_count ""


````

## 📄 `fsm.conf`

````conf


##### FSM ENTRY #####

# core
source-file ~/.tmux/fsm/core.conf

# grammar
source-file ~/.tmux/fsm/operator.conf
source-file ~/.tmux/fsm/motion.conf
source-file ~/.tmux/fsm/textobj.conf
source-file ~/.tmux/fsm/visual.conf


````

## 📄 `fsm/core.conf`

````conf


##### FSM CORE #####

# state
set-option -g @fsm_state ""
set-option -g @fsm_count ""
set-option -g @fsm_last_action ""
set-option -g @fsm_target_pane ""
set-option -g @fsm_log ""

# panic
bind-key -n C-g \
  set-option -g @fsm_state "" \;\
  set-option -g @fsm_count "" \;\
  set-option -g @fsm_last_action "" \;\
  set-option -g @fsm_target_pane "" \;\
  set-option -g @fsm_log "" \;\
  switch-client -T root

# enter FSM
bind-key -T root Space \
  set-option -g @fsm_state "NORMAL" \;\
  set-option -g @fsm_log "ENTER" \;\
  switch-client -T fsm

# exit FSM
bind-key -T fsm Escape \
  set-option -g @fsm_state "" \;\
  switch-client -T root

# count prefix
bind-key -T fsm 0 set-option -g @fsm_count "#{@fsm_count}0"
bind-key -T fsm 1 set-option -g @fsm_count "#{@fsm_count}1"
bind-key -T fsm 2 set-option -g @fsm_count "#{@fsm_count}2"
bind-key -T fsm 3 set-option -g @fsm_count "#{@fsm_count}3"
bind-key -T fsm 4 set-option -g @fsm_count "#{@fsm_count}4"
bind-key -T fsm 5 set-option -g @fsm_count "#{@fsm_count}5"
bind-key -T fsm 6 set-option -g @fsm_count "#{@fsm_count}6"
bind-key -T fsm 7 set-option -g @fsm_count "#{@fsm_count}7"
bind-key -T fsm 8 set-option -g @fsm_count "#{@fsm_count}8"
bind-key -T fsm 9 set-option -g @fsm_count "#{@fsm_count}9"

# dot repeat
bind-key -T fsm . \
  run-shell "for i in \$(seq 1 ${@fsm_count:-1}); do tmux #{@fsm_last_action}; done" \;\
  set-option -g @fsm_count ""

# debug
bind-key -T fsm ? display-message "FSM: #{@fsm_log}"


````

## 📄 `fsm/motion.conf`

````conf


##### MOTIONS #####

bind-key -T operator-d w \
  run-shell "
    pane=${@fsm_target_pane:-''}
    for i in \$(seq 1 ${@fsm_count:-1}); do
      tmux send-keys \$pane M-d
    done
  " \;\
  set-option -g @fsm_last_action \"send-keys ${@fsm_target_pane:-''} M-d\" \;\
  set-option -g @fsm_target_pane "" \;\
  set-option -g @fsm_count "" \;\
  switch-client -T fsm

bind-key -T operator-d b send-keys M-b M-d \;\
  set-option -g @fsm_last_action "send-keys M-b M-d" \;\
  switch-client -T fsm


````

## 📄 `fsm/operator.conf`

````conf


##### OPERATORS #####

# delete
bind-key -T fsm d \
  set-option -g @fsm_log "#{@fsm_log} d" \;\
  switch-client -T operator-d

# change = delete + insert
bind-key -T fsm c \
  set-option -g @fsm_log "#{@fsm_log} c" \;\
  switch-client -T operator-c

# yank
bind-key -T fsm y \
  set-option -g @fsm_log "#{@fsm_log} y" \;\
  switch-client -T operator-y

# paste
bind-key -T fsm p \
  send-keys C-y \;\
  set-option -g @fsm_last_action "send-keys C-y"

# pane targeting
bind-key -T operator-d h set-option -g @fsm_target_pane "-L"
bind-key -T operator-d j set-option -g @fsm_target_pane "-D"
bind-key -T operator-d k set-option -g @fsm_target_pane "-U"
bind-key -T operator-d l set-option -g @fsm_target_pane "-R"

# yank motion
bind-key -T operator-y w \
  send-keys M-d M-y \;\
  set-option -g @fsm_last_action "send-keys M-d M-y" \;\
  switch-client -T fsm


````

## 📄 `fsm/textobj.conf`

````conf


##### TEXTOBJECTS #####

bind-key -T operator-d i \
  switch-client -T operator-d-i

bind-key -T operator-d-i w \
  send-keys M-b M-f M-d \;\
  set-option -g @fsm_last_action "send-keys M-b M-f M-d" \;\
  switch-client -T fsm

bind-key -T operator-d-i '(' \
  send-keys C-M-b C-M-f M-d \;\
  set-option -g @fsm_last_action "send-keys C-M-b C-M-f M-d" \;\
  switch-client -T fsm


````

## 📄 `fsm/visual.conf`

````conf
SUAL MODE #####

# enter visual
bind-key -T fsm v \
  set-option -g @fsm_state "VISUAL" \;\
  send-keys C-Space \;\
  switch-client -T visual

# visual motions
bind-key -T visual w send-keys M-f
bind-key -T visual b send-keys M-b
bind-key -T visual e send-keys M-f

# visual delete / yank
bind-key -T visual d \
  send-keys M-d \;\
  set-option -g @fsm_last_action "send-keys M-d" \;\
  switch-client -T fsm

bind-key -T visual y \
  send-keys M-w \;\
  set-option -g @fsm_last_action "send-keys M-w" \;\
  switch-client -T fsm

# exit visual
bind-key -T visual Escape \
  set-option -g @fsm_state "" \;\
  switch-client -T fsm


````

## 📄 `motion.conf`

````conf


# ======================
# MOTIONS FOR OPERATOR d
# ======================

# dw
bind-key -T operator-d w \
  run-shell '
    count=$(tmux display -p "#{@fsm_count}")
    [ -z "$count" ] && count=1

    for i in $(seq 1 "$count"); do
      tmux send-keys M-d
    done
  ' \;\
  set-option -g @fsm_last_action "send-keys M-d" \;\
  set-option -g @fsm_count "" \;\
  switch-client -T fsm

# dd
bind-key -T operator-d d \
  run-shell '
    count=$(tmux display -p "#{@fsm_count}")
    [ -z "$count" ] && count=1

    for i in $(seq 1 "$count"); do
      tmux send-keys C-u C-k
    done
  ' \;\
  set-option -g @fsm_last_action "send-keys C-u C-k" \;\
  set-option -g @fsm_count "" \;\
  switch-client -T fsm


````

## 📄 `operator.conf`

````conf


# ======================
# OPERATORS
# ======================

# d → delete operator
bind-key -T fsm d \
  set-option -g @fsm_operator d \;\
  switch-client -T operator-d

# y → yank operator（预留）
bind-key -T fsm y \
  display-message "y operator (TODO)"


````

## 📄 `visual.conf`

````conf


# ======================
# VISUAL MODE
# ======================

bind-key -T fsm v \
  display-message "VISUAL MODE (stub)" \;\
  switch-client -T visual

bind-key -T visual Escape \
  switch-client -T fsm


````
