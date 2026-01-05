#!/usr/bin/env bash
# 路径验证脚本

echo "=== tmux-fsm 路径验证 ==="

# 检查二进制文件是否存在
BINARY_PATH="$HOME/.tmux/plugins/tmux-fsm/tmux-fsm"

if [ -f "$BINARY_PATH" ]; then
    echo "✅ 二进制文件存在: $BINARY_PATH"
    echo "   文件大小: $(ls -lh "$BINARY_PATH" | awk '{print $5}')"
    echo "   可执行权限: $(if [ -x "$BINARY_PATH" ]; then echo "是"; else echo "否"; fi)"
else
    echo "❌ 二进制文件不存在: $BINARY_PATH"
    echo "   请先运行 install.sh 或手动构建"
    exit 1
fi

# 测试二进制文件是否可以执行
echo ""
echo "=== 测试二进制文件功能 ==="
if "$BINARY_PATH" -h >/dev/null 2>&1; then
    echo "✅ 二进制文件可执行"
else
    echo "❌ 二进制文件执行失败"
    exit 1
fi

# 检查版本信息
echo ""
echo "=== 二进制文件信息 ==="
"$BINARY_PATH" -h

echo ""
echo "=== 路径验证完成 ==="
echo "所有路径配置正确，tmux-fsm 可以正常工作"