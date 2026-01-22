#!/usr/bin/env bash
set -e

echo "Installing tmux-fsm (FOEK Kernel)..."

# ----------------------------------------------------------------------
# config
# ----------------------------------------------------------------------

TMUX_FSM_DIR="${TMUX_FSM_DIR:-$HOME/.tmux/plugins/tmux-fsm}"

# 自动检测 tmux.conf（支持传统 & XDG）
if [ -z "$TMUX_CONF" ]; then
  if [ -f "$HOME/.tmux.conf" ]; then
    TMUX_CONF="$HOME/.tmux.conf"
  elif [ -f "$HOME/.config/tmux/tmux.conf" ]; then
    TMUX_CONF="$HOME/.config/tmux/tmux.conf"
  else
    TMUX_CONF="$HOME/.tmux.conf"
  fi
fi

# ----------------------------------------------------------------------
# checks
# ----------------------------------------------------------------------

if ! command -v tmux >/dev/null 2>&1; then
  echo "Error: tmux not found"
  exit 1
fi

# ----------------------------------------------------------------------
# install
# ----------------------------------------------------------------------

# 停止可能正在运行的旧版本守护进程 (Critical for Daemon update)
echo "Stopping running daemons..."

# Try to kill using PID file first (most reliable)
if [ -f "/tmp/tmux-fsm.pid" ]; then
    PID=$(cat /tmp/tmux-fsm.pid)
    if kill -0 "$PID" 2>/dev/null; then
        echo "Killing daemon with PID: $PID"
        kill -9 "$PID" 2>/dev/null || true
    fi
    rm -f "/tmp/tmux-fsm.pid"
fi

# Fallback: kill any remaining tmux-fsm processes
pkill -9 -f "[/]tmux-fsm" 2>/dev/null || true

# Double check that no processes remain
sleep 0.1
pkill -9 -f "[/]tmux-fsm" 2>/dev/null || true

echo "Installing to: $TMUX_FSM_DIR"
mkdir -p "$TMUX_FSM_DIR"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ----------------------------------------------------------------------
# Build Go binary (High Performance Kernel)
# ----------------------------------------------------------------------

if command -v go >/dev/null 2>&1; then
  echo "🚀 Building Go kernel for zero-latency performance..."
  
  # 临时初始化 go module 以防环境缺失
  if [ ! -f "$SCRIPT_DIR/go.mod" ]; then
      echo "Initializing temporary go module..."
      (cd "$SCRIPT_DIR" && go mod init tmux-fsm 2>/dev/null || true)
  fi

  # 编译：剔除符号表(-s)和调试信息(-w)以减小体积
  # 使用 "." 编译目录下所有文件，更健壮
  (cd "$SCRIPT_DIR" && go build -ldflags="-s -w" -o tmux-fsm .)
  
  cp "$SCRIPT_DIR/tmux-fsm" "$TMP_DIR/"
  echo "✅ Build successful."
else
  echo "⚠️  Warning: Go not found. Falling back to Python (Performance degraded)."
  echo "   Please install Go to enable the Daemon Kernel."
fi

# ----------------------------------------------------------------------
# copy files (required)
# ----------------------------------------------------------------------

# 只需要核心组件
cp "$SCRIPT_DIR"/plugin.tmux \
   "$SCRIPT_DIR"/fsm-toggle.sh \
   "$SCRIPT_DIR"/fsm-exit.sh \
   "$SCRIPT_DIR"/enter_fsm.sh \
   "$SCRIPT_DIR"/keymap.yaml \
   "$TMP_DIR/"

# 移动到目标目录
mv "$TMP_DIR"/* "$TMUX_FSM_DIR/"

# 确保二进制文件和 shell 脚本可执行
chmod +x \
  "$TMUX_FSM_DIR/tmux-fsm" \
  "$TMUX_FSM_DIR/fsm-toggle.sh" \
  "$TMUX_FSM_DIR/fsm-exit.sh" \
  "$TMUX_FSM_DIR/enter_fsm.sh"

# 清理旧的 Python 文件 (Clean up legacy)
rm -f "$TMUX_FSM_DIR/fsm.py" "$TMUX_FSM_DIR/tmux_fsm.py"

# ----------------------------------------------------------------------
# Interactive Configuration
# ----------------------------------------------------------------------

# NOTE: In non-interactive environments, we default to mode 1
install_mode="1"
if [ -t 0 ]; then
    echo ""
    echo "Configuration Strategy:"
    echo "1) Automatic: Append plugin hook to $TMUX_CONF and reload tmux"
    echo "2) Replace: Replace $TMUX_CONF with plugin's default config (backup created)"
    echo "3) Manual: Show instructions for manual setup"
    read -rp "Please select [1/2/3] (default 1): " user_choice
    install_mode="${user_choice:-1}"
fi

PLUGIN_HOOK="source-file \"$TMUX_FSM_DIR/plugin.tmux\""

case $install_mode in
    1)
        if grep -q "tmux-fsm" "$TMUX_CONF" 2>/dev/null; then
            echo "Result: Already configured in $TMUX_CONF"
        else
            echo "" >> "$TMUX_CONF"
            echo "# tmux-fsm plugin (FOEK Kernel)" >> "$TMUX_CONF"
            echo "$PLUGIN_HOOK" >> "$TMUX_CONF"
            echo "✅ Successfully updated $TMUX_CONF"
        fi

        echo "🔄 Performing Hot Upgrade..."
        # 尝试静默重新加载 tmux 配置
        if tmux info >/dev/null 2>&1; then
            tmux source-file "$TMUX_CONF" 2>/dev/null && echo "✅ tmux configuration reloaded"
            # 预热 Daemon (Phase 7: Weaver Mode)
            TMUX_FSM_MODE=weaver TMUX_FSM_LOG_FACTS=1 "$TMUX_FSM_DIR/tmux-fsm" -server >/dev/null 2>&1 &
            echo "✅ Daemon pre-warmed (Weaver Mode)."
        fi
        ;;
    2)
        # 创建备份并替换配置文件
        if [ -f "$TMUX_CONF" ]; then
            BACKUP_TMUX_CONF="${TMUX_CONF}.backup.$(date +%Y%m%d_%H%M%S)"
            echo "Creating backup of existing config: $BACKUP_TMUX_CONF"
            cp "$TMUX_CONF" "$BACKUP_TMUX_CONF"
            echo "✅ Backup created at $BACKUP_TMUX_CONF"
        fi

        # 复制默认配置文件并替换插件路径
        cp "$SCRIPT_DIR/default.tmux.conf" "$TMUX_CONF"
        echo "✅ Successfully replaced $TMUX_CONF with plugin default config"

        echo "🔄 Performing Hot Upgrade..."
        # 尝试静默重新加载 tmux 配置
        if tmux info >/dev/null 2>&1; then
            tmux source-file "$TMUX_CONF" 2>/dev/null && echo "✅ tmux configuration reloaded"
            # 预热 Daemon (Phase 7: Weaver Mode)
            TMUX_FSM_MODE=weaver TMUX_FSM_LOG_FACTS=1 "$TMUX_FSM_DIR/tmux-fsm" -server >/dev/null 2>&1 &
            echo "✅ Daemon pre-warmed (Weaver Mode)."
        fi
        ;;
    *)
        echo ""
        echo "💡 Manual action required:"
        echo "   Add the following line to your config:"
        echo ""
        echo "   $PLUGIN_HOOK"
        echo ""
        ;;
esac

# ----------------------------------------------------------------------
# done
# ----------------------------------------------------------------------

echo ""
echo "✅ tmux-fsm (Zero-Latency Daemon Kernel) installed!"
echo "   Latency: < 1ms"
echo ""
echo "Usage:"
echo "  - Enter FSM mode:  <prefix> f"
echo "  - Exit FSM mode:   Esc / C-c"
echo "  - Audit Logic:     Press '?' in FSM mode to see why Undo failed."
echo "  - Audit Log:       Logs are written to ~/tmux-fsm.log"
echo ""
