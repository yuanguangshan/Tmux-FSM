#!/bin/bash
# 阶段 0 基线测试脚本
# 用于验证重构后功能一致性

set -e

echo "=== tmux-fsn 基线测试 ==="
echo "Tag: pre-weaver-migration"
echo "Date: $(date)"
echo ""

# 测试 1: 基本移动命令
test_basic_movement() {
    echo "测试 1: 基本移动命令 (h/j/k/l)"
    # 这里需要在实际 tmux 环境中测试
    # 预期：光标正确移动
    echo "  ✓ 需要手动验证"
}

# 测试 2: 删除操作 + Undo
test_delete_undo() {
    echo "测试 2: 删除操作 + Undo"
    # 场景：dw dw dw 然后 u u u
    # 预期：删除三个词，撤销三次后恢复
    echo "  ✓ 需要手动验证"
}

# 测试 3: 移动光标后 delete
test_move_then_delete() {
    echo "测试 3: 移动光标后 delete"
    # 场景：移动光标到中间，执行 dw
    # 预期：Anchor 正确定位，删除正确的词
    echo "  ✓ 需要手动验证"
}

# 测试 4: 跨 pane 操作
test_cross_pane() {
    echo "测试 4: 跨 pane / window 操作"
    # 场景：在不同 pane 中切换并执行操作
    # 预期：状态正确隔离
    echo "  ✓ 需要手动验证"
}

# 测试 5: 文本对象
test_text_objects() {
    echo "测试 5: 文本对象 (diw, ci\", 等)"
    # 场景：diw, ci", da(
    # 预期：正确识别并操作文本对象
    echo "  ✓ 需要手动验证"
}

# 测试 6: Visual 模式
test_visual_mode() {
    echo "测试 6: Visual 模式"
    # 场景：v 选择，d 删除
    # 预期：正确进入/退出 visual 模式
    echo "  ✓ 需要手动验证"
}

# 测试 7: 搜索功能
test_search() {
    echo "测试 7: 搜索功能 (/, n, N)"
    # 场景：/pattern, n, N
    # 预期：正确搜索和跳转
    echo "  ✓ 需要手动验证"
}

# 测试 8: FSM 层级切换
test_fsm_layers() {
    echo "测试 8: FSM 层级切换 (g -> GOTO)"
    # 场景：g 进入 GOTO 层，gg 跳转到顶部
    # 预期：层级正确切换，超时自动退出
    echo "  ✓ 需要手动验证"
}

# 执行所有测试
echo "开始执行基线测试..."
echo ""

test_basic_movement
test_delete_undo
test_move_then_delete
test_cross_pane
test_text_objects
test_visual_mode
test_search
test_fsm_layers

echo ""
echo "=== 基线测试完成 ==="
echo "请手动验证每个测试场景"
echo ""
echo "如果所有测试通过，记录当前状态："
echo "  git log -1 --oneline"
echo "  git show pre-weaver-migration"
