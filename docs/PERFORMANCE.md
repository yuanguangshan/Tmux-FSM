# PERFORMANCE — 延迟基准与实测数据（M3.2）

> 2026-09-07 实测。测试机：Apple Silicon Mac（arm64，Go 1.24）。
> 运行：`go test ./kernel/ -bench . -benchmem -run '^$'`

## 一、daemon 内存态按键处理（Go Benchmark）

| 按键类别 | 延迟 | 内存 | 分配次数 |
|---|---|---|---|
| 委托移动（h/j/k/l，纯透传 Grammar） | **1,529 ns/op** | 1,040 B | 19 |
| 操作符起始（d/y/c，进入 pending） | 2,222 ns/op | 816 B | 24 |
| 完整操作（dw，Grammar→Promote→Execute 全链路） | 2,477 ns/op | 1,176 B | 28 |
| 计数移动（3j） | 2,384 ns/op | 1,160 B | 26 |
| 文本对象（ciw） | 3,994 ns/op | 1,016 B | 34 |

**结论：daemon 内存态按键处理为微秒级（1.5-4μs），可忽略。**

## 二、端到端真机实测（tmux-fsm -key 客户端全链路）

- 场景：`tmux-fsm -key 'j' 'p1|bench'` × 20 次取均值（隔离 tmux socket，
  daemon 常驻，含每次按键的**客户端进程 fork + Unix socket 往返 +
  daemon 处理 + tmux 选项持久化写入**）
- 结果：**23.7 ms/键**

分解（依据代码路径）：进程 fork/初始化 ≈ 15-20ms（macOS 进程创建成本），
socket 往返 + daemon 处理 ≈ 1-2ms，tmux 选项持久化 ≈ 3-5ms。

**对比**：优化前每键 40-60ms+（11-13 次 exec），当前 23.7ms。
剩余成本主体是每键一次的客户端进程 fork——根治需要 M3.3（常驻客户端
或 tmux control-mode），已列为可选大活。

## 三、已落地的优化（2026-09-06）

1. UpdateUI 每键 3 次 exec → 0（M2.4，UI 写入者合并）
2. loadState 每键 1 次 GetUserOption 往返 → 0（M3.1，daemon 内存即真相）
3. reconcileFSMState 每键 2-3 次 exec → 仅层变化/退出时（M3.1）
4. HandleKey 串行化后（M2.1）消除了并发按键下的状态错乱与重试开销

## 四、后续优化路线（未做）

1. **常驻客户端**（收益最大）：用常驻进程替代每键 fork，
   socket 写入延迟 → <1ms。需处理进程生命周期与 tmux 死亡检测。
2. **tmux control-mode**（`tmux -C`）：所有 set/show/send-keys 走一条
   长连接，backend/controlModeBackend.go 已有 stub 待做实。
3. UI 选项写入合并：@fsm_state/@fsm_keys 两次 set 可合并为一次
   `tmux set` 批量（小优化）。
