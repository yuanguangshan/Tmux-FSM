# 🚨 迁移失败止损清单

**重要性**: ⭐⭐⭐⭐⭐  
**请打印此文档并放在手边**

---

## 一、立即停止迁移的信号

遇到以下任何情况，**立即停止**，不要继续：

### 🔴 严重问题（立即回滚）

1. **Undo 不可预测删除**
   - 症状：执行 `u` 后删除了错误的内容
   - 症状：Undo 删除的范围超出预期
   - 症状：Undo 导致 pane 内容混乱

2. **Anchor 定位错位 > 1 行**
   - 症状：日志显示 `ResolveFuzzy` 且偏移 > 1 行
   - 症状：操作作用在错误的行上
   - 症状：`ResolveAnchor` 返回 `ResolveFail`

3. **Panic / Tmux 卡死**
   - 症状：tmux-fsm 进程崩溃
   - 症状：tmux 无响应
   - 症状：键盘输入无反应

4. **无法切回 Legacy 模式**
   - 症状：设置 `TMUX_FSM_MODE=legacy` 后仍然使用新代码
   - 症状：Feature flag 不生效

### 🟡 警告问题（暂停并调查）

1. **Fuzzy Undo 频率过高**
   - 症状：超过 30% 的 Undo 是 fuzzy
   - 症状：状态栏频繁显示 `~UNDO`

2. **性能明显下降**
   - 症状：按键响应延迟 > 100ms
   - 症状：CPU 占用异常

3. **日志中出现大量错误**
   - 症状：`tmux-fsm.log` 中 error 数量激增
   - 症状：出现未预期的错误消息

---

## 二、30 秒止损操作流程

### 步骤 1: 立即切回 Legacy 模式（10 秒）

```bash
# 设置环境变量
export TMUX_FSM_MODE=legacy

# 刷新 tmux 客户端
tmux refresh-client -S
```

**验证**: 按几个键，确认功能正常

---

### 步骤 2: 如果仍异常，重启服务器（10 秒）

```bash
# 停止当前服务器
tmux-fsm -stop

# 等待 1 秒
sleep 1

# 重新启动服务器
tmux-fsm -server &

# 等待服务器启动
sleep 1

# 重新进入 FSM 模式
tmux-fsm -enter
```

**验证**: 测试基本操作（如 `dw` + `u`）

---

### 步骤 3: 如果还是异常，回滚代码（10 秒）

```bash
# 回到安全的 tag
git checkout pre-weaver-migration

# 重新编译
go build -o tmux-fsm

# 重启服务器
tmux-fsm -stop
tmux-fsm -server &
```

**验证**: 运行 `tests/baseline_tests.sh`

---

## 三、止损后的调查步骤

### 1. 保存现场

```bash
# 保存日志
cp ~/tmux-fsm.log ~/tmux-fsm-failure-$(date +%Y%m%d-%H%M%S).log

# 保存当前状态
tmux-fsm -key __STATUS__ > ~/tmux-fsm-state-$(date +%Y%m%d-%H%M%S).json

# 保存 git 状态
git diff > ~/tmux-fsm-diff-$(date +%Y%m%d-%H%M%S).patch
```

### 2. 分析日志

查找关键错误信息：

```bash
# 查找 panic
grep -i "panic" ~/tmux-fsm.log

# 查找 Anchor 失败
grep "ResolveFail" ~/tmux-fsm.log

# 查找 Undo 错误
grep "UNDO-SKIP" ~/tmux-fsm.log

# 查找最近的错误
tail -100 ~/tmux-fsm.log | grep -i "error"
```

### 3. 确定失败阶段

- 阶段 0: 不应该有任何变化
- 阶段 1: 只有 Intent 抽取，执行路径未变
- 阶段 2: Shadow 模式，不应影响实际行为
- 阶段 3+: 执行路径已变化

---

## 四、绝对禁止的操作

在迁移过程中，**绝对不要做**以下事情：

### ❌ 1. 同时修改多个层次

**错误示例**:
```
同时修改 FSM + Executor + Undo
```

**正确做法**:
```
一次只改一层，验证通过后再改下一层
```

---

### ❌ 2. 删除 Legacy 代码后再写新代码

**错误示例**:
```go
// 删除旧的 executeAction
// 然后写新的 Weaver.Apply
```

**正确做法**:
```go
// 保留 executeAction
// 新增 Weaver.Apply
// 用 feature flag 切换
// 验证通过后再删除旧代码
```

---

### ❌ 3. "顺手"清理代码

**错误示例**:
```
"反正都在改，顺便重命名一下变量"
"顺便优化一下这个函数"
```

**正确做法**:
```
只做迁移相关的修改
其他优化等迁移完成后再做
```

---

### ❌ 4. 在同一阶段同时动 FSM + Undo

**错误示例**:
```
阶段 3: 修改 FSM 输出 + 修改 Undo 逻辑
```

**正确做法**:
```
阶段 3: 只修改执行路径
阶段 5: 再修改 Undo
```

---

### ❌ 5. 跳过测试直接进入下一阶段

**错误示例**:
```
阶段 1 完成，没测试就开始阶段 2
```

**正确做法**:
```
每个阶段完成后：
1. 运行 baseline_tests.sh
2. 手动验证关键场景
3. 检查日志无异常
4. 打 tag (如 phase-1-complete)
5. 再开始下一阶段
```

---

## 五、每个阶段的回滚点

### 阶段 0
- **Tag**: `pre-weaver-migration`
- **回滚**: `git checkout pre-weaver-migration`

### 阶段 1
- **Tag**: `phase-1-intent-extraction`
- **回滚**: `git checkout phase-1-intent-extraction`
- **Feature Flag**: 不需要（只是抽取，未改行为）

### 阶段 2
- **Tag**: `phase-2-shadow-mode`
- **回滚**: `TMUX_FSM_MODE=legacy` 或 `git checkout phase-1-intent-extraction`
- **Feature Flag**: `TMUX_FSM_MODE=legacy|shadow`

### 阶段 3
- **Tag**: `phase-3-projection`
- **回滚**: `TMUX_FSM_MODE=legacy` 或 `git checkout phase-2-shadow-mode`
- **Feature Flag**: `TMUX_FSM_MODE=legacy|weaver`

### 阶段 4
- **Tag**: `phase-4-anchor-resolver`
- **回滚**: `TMUX_FSM_MODE=legacy` 或 `git checkout phase-3-projection`

### 阶段 5
- **Tag**: `phase-5-undo-migration`
- **回滚**: `TMUX_FSM_MODE=legacy` 或 `git checkout phase-4-anchor-resolver`

---

## 六、联系与支持

### 自查清单

遇到问题时，先问自己：

- [ ] 我是否跳过了某个阶段？
- [ ] 我是否同时修改了多个层次？
- [ ] 我是否运行了测试？
- [ ] 我是否保存了日志？
- [ ] 我是否可以用 feature flag 切回 legacy？

### 调试技巧

1. **对比日志**
   ```bash
   # Legacy 模式日志
   TMUX_FSM_MODE=legacy tmux-fsm -key d
   
   # Weaver 模式日志
   TMUX_FSM_MODE=weaver tmux-fsm -key d
   
   # 对比差异
   diff legacy.log weaver.log
   ```

2. **单步调试**
   ```bash
   # 只测试一个按键
   TMUX_FSM_LOG_FACTS=1 tmux-fsm -key d
   
   # 查看产生的 Intent/Fact
   tail -20 ~/tmux-fsm.log
   ```

3. **状态检查**
   ```bash
   # 查看当前状态
   tmux-fsm -key __STATUS__
   
   # 查看 Undo 失败原因
   tmux-fsm -key __WHY_FAIL__
   ```

---

## 七、成功标准

每个阶段完成后，必须满足：

✅ 所有 baseline 测试通过  
✅ 日志中无 error/panic  
✅ 性能无明显下降  
✅ 可以随时切回 legacy  
✅ 代码已提交并打 tag

---

**最后提醒**:

> 🔥 **迁移不是一次性的，是渐进的**  
> 🔥 **任何时候都可以回滚**  
> 🔥 **不要急于删除旧代码**  
> 🔥 **测试 > 速度**

---

**打印日期**: _______________  
**当前阶段**: _______________  
**最后测试**: _______________
