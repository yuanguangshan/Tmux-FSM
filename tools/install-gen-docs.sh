#!/usr/bin/env bash
# 项目文档生成工具安装脚本（优化版）
# 解决多路径版本冲突，优先安装至用户 PATH 中的高优先级目录

set -e

echo "🚀 开始安装 gen-docs..."

# -------- 基础检查 --------
if ! command -v go &> /dev/null; then
    echo "❌ 未检测到 Go 编译器"
    echo "请先安装 Go: https://go.dev/dl/"
    exit 1
fi

echo "✓ Go 版本: $(go version)"

# -------- 确定安装路径 --------
# 优先检查用户专有目录，减少对 sudo 的依赖，且通常这些目录在 PATH 优先级更高
USER_LOCAL_BIN="$HOME/.local/bin"
SYSTEM_BIN="/usr/local/bin"
INSTALL_DIR=""

# 找出当前 gen-docs 的位置（如果已安装）
EXISTING_PATH=$(which gen-docs 2>/dev/null || true)

if [[ ":$PATH:" == *":$USER_LOCAL_BIN:"* ]]; then
    INSTALL_DIR="$USER_LOCAL_BIN"
    echo "💡 检测到 $USER_LOCAL_BIN 已在 PATH 中，将优先安装至此。"
elif [[ ":$PATH:" == *":$SYSTEM_BIN:"* ]]; then
    INSTALL_DIR="$SYSTEM_BIN"
    echo "💡 将安装至系统目录 $SYSTEM_BIN。"
else
    INSTALL_DIR="$USER_LOCAL_BIN"
    echo "⚠️  未在 PATH 中发现常用 bin 目录，默认安装至 $INSTALL_DIR"
fi

mkdir -p "$INSTALL_DIR"

# -------- 编译 --------
echo "📦 正在本地编译..."
go build -o gen-docs_new gen-docs.go

# -------- 安装主程序 --------
echo "📥 正在安装到 $INSTALL_DIR/gen-docs ..."

# 检查权限
USE_SUDO=""
if [ ! -w "$INSTALL_DIR" ]; then
    echo "🔑 需要 sudo 权限来写入 $INSTALL_DIR"
    USE_SUDO="sudo"
fi

$USE_SUDO mv gen-docs_new "$INSTALL_DIR/gen-docs"
$USE_SUDO chmod +x "$INSTALL_DIR/gen-docs"

# -------- 处理软链接 (gd) --------
echo "🔗 创建/更新 gd 快捷命令"
$USE_SUDO ln -sf "$INSTALL_DIR/gen-docs" "$INSTALL_DIR/gd"

# -------- 清理旧版本冲突 --------
# 如果之前在不同目录下安装过，提醒用户或尝试清理
if [ -n "$EXISTING_PATH" ] && [ "$EXISTING_PATH" != "$INSTALL_DIR/gen-docs" ]; then
    echo "⚠️  发现旧版本存在于: $EXISTING_PATH"
    echo "建议手动删除旧版本以避免冲突: rm $EXISTING_PATH"
fi

# -------- PATH 检查 --------
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "‼️  重要: $INSTALL_DIR 不在你的 PATH 环境变量中！"
    echo "请将以下内容添加到你的 ~/.zshrc 或 ~/.bashrc:"
    echo "    export PATH=\"$INSTALL_DIR:\$PATH\""
else
    echo "✓ 验证结果: $(gen-docs -version) 安装成功"
fi

echo ""
echo "✅ 安装完成！"
echo "你可以使用 'gen-docs' 或简写 'gd' 来运行程序。"
echo "输入 'gd -s' 体验你新增的统计功能！"
echo ""
