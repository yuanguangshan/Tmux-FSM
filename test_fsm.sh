#!/bin/bash

echo "=== 开始 tmux-fsm 全面测试 ==="

# 首先停止任何正在运行的服务器
echo "停止任何正在运行的服务器..."
/Users/ygs/ygs/ygs/learning/tmuxPlugin/tmux-fsm -stop 2>/dev/null || true
sleep 1

# 1. 构建测试
echo "1. 测试构建..."
cd /Users/ygs/ygs/ygs/learning/tmuxPlugin
go clean
if go build -o tmux-fsm; then
    echo "✅ 构建成功"
else
    echo "❌ 构建失败"
    exit 1
fi

# 2. Keymap 验证测试
echo "2. 测试 Keymap 验证..."
if ./tmux-fsm -config keymap.yaml -reload; then
    echo "✅ 有效配置加载成功"
else
    echo "❌ 有效配置加载失败"
    exit 1
fi

# 创建无效配置测试验证功能
cat > invalid_keymap.yaml << 'EOF'
states:
  NAV:
    hint: "test"
    keys:
      g: { layer: NONEXISTENT, timeout_ms: 800 }
EOF

if ./tmux-fsm -config invalid_keymap.yaml -reload; then
    echo "❌ 无效配置应该报错但没有"
    rm invalid_keymap.yaml
    exit 1
else
    echo "✅ 无效配置正确报错"
fi
rm invalid_keymap.yaml

# 3. 服务器模式测试
echo "3. 测试服务器模式..."
./tmux-fsm -server &
SERVER_PID=$!
sleep 2  # 等待服务器完全启动

# 检查服务器是否启动
if ps -p $SERVER_PID > /dev/null; then
    echo "✅ 服务器启动成功"
else
    echo "❌ 服务器启动失败"
    exit 1
fi

# 4. FSM 生命周期测试
echo "4. 测试 FSM 生命周期..."
if ./tmux-fsm -enter; then
    echo "✅ 进入 FSM 成功"
else
    echo "❌ 进入 FSM 失败"
fi

if ./tmux-fsm -key h; then
    echo "✅ 按键 h 分发成功"
else
    echo "❌ 按键 h 分发失败"
fi

if ./tmux-fsm -key g; then
    echo "✅ 按键 g 分发成功"
else
    echo "❌ 按键 g 分发失败"
fi

if ./tmux-fsm -key h; then
    echo "✅ 按键 h (在 GOTO 层) 分发成功"
else
    echo "❌ 按键 h (在 GOTO 层) 分发失败"
fi

if ./tmux-fsm -exit; then
    echo "✅ 退出 FSM 成功"
else
    echo "❌ 退出 FSM 失败"
fi

# 停止服务器
if ./tmux-fsm -stop; then
    echo "✅ 停止服务器成功"
else
    echo "❌ 停止服务器失败"
fi

# 等待服务器进程结束
sleep 1
if ps -p $SERVER_PID > /dev/null; then
    kill $SERVER_PID 2>/dev/null || true
fi

# 5. UI 测试
echo "5. 测试 UI 功能..."
./tmux-fsm -server &
SERVER_PID2=$!
sleep 2  # 等待服务器完全启动

if ./tmux-fsm -ui-show; then
    echo "✅ UI 显示成功"
else
    echo "❌ UI 显示失败"
fi

if ./tmux-fsm -ui-hide; then
    echo "✅ UI 隐藏成功"
else
    echo "❌ UI 隐藏失败"
fi

./tmux-fsm -stop
sleep 1
if ps -p $SERVER_PID2 > /dev/null; then
    kill $SERVER_PID2 2>/dev/null || true
fi

echo "=== 所有测试完成 ==="