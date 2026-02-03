#!/usr/bin/env bash
# 项目文档生成工具安装脚本（优化版）
# 解决多路径版本冲突，优先安装至用户 PATH 中的高优先级目录

set -e

echo "🚀 开始安装 codoc..."

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

# 找出当前 codoc 的位置（如果已安装）
EXISTING_PATH=$(which codoc 2>/dev/null || true)

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
# 尝试获取 git hash，如果失败则使用 current
GIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "current")
BUILD_DATE=$(date +%Y%m%d)
# 注入版本信息 (main.versionStr 必须是 var 且原本在 main 包中)
GO_LDFLAGS="-X main.versionStr=v2.1.0-${BUILD_DATE}-${GIT_HASH}"

go build -ldflags "${GO_LDFLAGS}" -o codoc_new codoc.go

# -------- 安装主程序 --------
echo "📥 正在安装到 $INSTALL_DIR/codoc ..."

# 检查权限
USE_SUDO=""
if [ ! -w "$INSTALL_DIR" ]; then
    echo "🔑 需要 sudo 权限来写入 $INSTALL_DIR"
    USE_SUDO="sudo"
fi

$USE_SUDO mv codoc_new "$INSTALL_DIR/codoc"
$USE_SUDO chmod +x "$INSTALL_DIR/codoc"

# -------- 处理软链接 (gd) --------
echo "🔗 创建/更新 gd 快捷命令"
$USE_SUDO ln -sf "$INSTALL_DIR/codoc" "$INSTALL_DIR/gd"

# -------- 清理旧版本冲突 --------
echo "🧹 清理旧版本..."
# 删除已知路径的 gen-docs
if [ -f "$USER_LOCAL_BIN/gen-docs" ]; then
    rm "$USER_LOCAL_BIN/gen-docs"
    echo "  - 已删除 $USER_LOCAL_BIN/gen-docs"
fi
if [ -f "$SYSTEM_BIN/gen-docs" ]; then
    $USE_SUDO rm "$SYSTEM_BIN/gen-docs"
    echo "  - 已删除 $SYSTEM_BIN/gen-docs"
fi

# -------- PATH 检查 --------
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "‼️  重要: $INSTALL_DIR 不在你的 PATH 环境变量中！"
    echo "请将以下内容添加到你的 ~/.zshrc 或 ~/.bashrc:"
    echo "    export PATH=\"$INSTALL_DIR:\$PATH\""
else
    echo "✓ 验证结果: $(codoc -version) 安装成功"
fi

echo ""
echo "✅ 安装完成！"
echo ""
echo "👉 注意事项："
echo "1. 你可以使用 'codoc' 来运行程序。"
echo "2. 我们创建了 'gd' 作为 codoc 的快捷方式。"
echo "   ⚠️  如果你之前设置了 'alias gd=gen-docs'，请在 .zshrc/.bashrc 中删除它"
echo "      然后执行: unalias gd"
echo "3. 输入 'codoc -s' (或 gd -s) 体验统计功能！"
echo ""
