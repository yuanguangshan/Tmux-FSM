# GOAL — Tmux-FSM 复活计划（自主执行直到 2026-09-07 09:00）

> 本文件是全自动优化循环的**唯一事实来源**。每个新会话：先读本文档与 PROGRESS.md，
> 从「任务队列」取第一个未完成任务执行，完成后更新 PROGRESS.md 并 git commit。
> **不要重新规划**——计划已定，执行即可。遇到与本文档冲突的现实，更新文档并继续。

## 使命

tmux-fsm 是把 Vim 模态编辑（动词-对象模型、count、文本对象、undo）搬进 tmux
shell 的守护进程插件。当前为半成品：核心 advertised 功能大面积损坏、测试全红、
每键延迟 40-60ms。目标：**2026-09-07 09:00 前，让它成为可用、可靠、可维护的
完整体**。完成的定义见下方「完成判定」。

## 铁律（每轮必守）

1. 每轮结束必须：`go build ./... && go vet ./... && go test ./...` 全绿 + `git commit`。
2. 一次只做一个任务（除非任务明确包含多个子步）；做完在 PROGRESS.md 打勾。
3. 不删除 docs/、优化建议、历史 RFC——只加不停。
4. 遇到与本计划冲突的现实（如接口对不上），按"最小改动让测试全绿"原则处理，
   并在 PROGRESS.md 的「决策记录」登记。
5. 重要里程碑（M0/M1/M2/M3/M4 完成）发微信：`python3 ~/.pi/agent/skills/wechat-send/scripts/send.py "..."`。
6. 2026-09-07 09:00 之后：不再开始新任务，只收尾已开始的；然后发送最终报告。

## 任务队列（按序执行，禁止跳号）

### M0 — 复活：构建与测试全绿（地基，最高优先级）

- [ ] M0.1 `tools` 包编译失败：gen-docs.go 与 codoc.go 重复声明
      （versionStr/Config 等）。最小修复：给其中一个文件加 `//go:build ignore`
      或合并重复声明。验收：`go build ./tools/` 通过。
- [ ] M0.2 `fsm/engine_test.go:76` Dispatch 签名不匹配（两返回值改动未同步测试）。
      适配测试。验收：`go test ./fsm/` 编译并运行。
- [ ] M0.3 `tests/integration_test.go:68` PaneID 提取为空 → 按 ActorID 提取链路修正。
- [ ] M0.4 `weaver/adapter` 测试期望代价 50 得 0 → 判断是实现缺陷还是期望过时，
      修对应一侧并写明理由。
- [ ] M0.5 `gofmt -w .` 全仓格式化 + 移除 git 跟踪的大文件
      （chat.txt 253KB / codoc.md 1.6MB / 本地二进制）。
- [ ] M0.6 **M0 验收**：`go build ./... && go vet ./... && go test ./...` 全绿。
      → 发微信「M0 完成」。

### M1 — 核心可用：广告过的功能必须真的能用

- [ ] M1.1 exit 死路修复：kernel/execute.go:96 假 exit 分支删除，
      统一走 kernel.go:211 的 executeAction（fsm.ExitFSM()）；
      plugin.tmux 第 6 节循环剔除 q/Escape/C-c（避免覆盖第 5 节表级退出）。
- [ ] M1.2 计数器生命周期：fsm/engine.go Dispatch 产 token 后清零 e.count；
      FSM count 与 Grammar count 单一来源化（保留 Grammar 侧）。
- [ ] M1.3 未知键透传 Grammar：kernel/decide.go 对 keymap 未命中键也发
      TokenKey 给 Grammar（报告方案①），keymap 只管 action/layer；
      删除 planner/grammar.go:101-112 的死分支。
- [ ] M1.4 Direction 元数据补全：planner/grammar.go make*Intent 补 w/b/e/ge；
      intent/promote.go 对 MotionWord+DirectionNone fail-fast 报错。
- [ ] M1.5 f/t/F/T 目标字符进入 Grammar（依赖 M1.3 的透传），
      pendingMotion 挂起后等待下一字符。
- [ ] M1.6 **M1 验收**：新增端到端回归测试 dw/db/de/dG/dfa/dat/ciw/ci"/yiw/
      3j0/$ 全部产生正确 Intent；`go test ./...` 全绿。→ 发微信「M1 完成」。

### M2 — 正确性：并发与生命周期

- [ ] M2.1 HandleKey 串行化：Server 层 chan func() + 单 worker
      （kernel.go:17 注释要求的串行语义），附并发按键测试 + `go test -race`。
- [ ] M2.2 daemon 单实例：socket 文件 syscall.Flock 独占锁，
      抢不到锁的新实例打印 "already running" 后退出（替代 os.Remove 硬删）。
- [ ] M2.3 install.sh 的 pkill -9 -f 精确化（匹配可执行文件路径而非 -f 模糊）。
- [ ] M2.4 双 UI 写入者合并：fsm/ui_stub.go updateTmuxVariables 与
      globals.go updateStatusBar 二选一保留，另一处删除。
- [ ] M2.5 **M2 验收**：全绿 + 手动场景（start → 按键 → q 退出 → daemon 无残留）。
      → 发微信「M2 完成」。

### M3 — 性能：把 40-60ms 打下来

- [ ] M3.1 砍每键 tmux 往返：reconcileFSMState 只在 enter/exit 跑；
      loadState 不在每键路径读 tmux 选项（daemon 内存即真相）；
      UpdateUI 与 updateStatusBar 合并。目标：单键 exec 次数 ≤4。
- [ ] M3.2 延迟基准测试（ BenchmarkHandleKey + 真机 time 实测）写入 docs。
- [ ] M3.3 （大活，时间允许才做）control-mode backend 做实：一条持久
      `tmux -C` 连接承载全部 set/show/send-keys。
- [ ] M3.4 **M3 验收**：单键延迟实测 <15ms 并记录；全绿。→ 发微信「M3 完成」。

### M4 — 功能完善（Vim 完整性，按剩余时间做）

- [ ] M4.1 Visual 模式真实现或诚实移除：给 GrammarIntent 加 VisualMode 字段，
      v/V 区分字符/行选区，配合选择与操作符；做不到就按报告建议移除假绑定。
- [ ] M4.2 搜索 / ? n N（基于 shell CWD 与 readline 语义评估可行性；
      不可行则在 README 标注不支持并从 keymap 移除）。
- [ ] M4.3 寄存器（"a-y 操作）最小实现。
- [ ] M4.4 `.` 重复上一操作。
- [ ] M4.5 文本对象扩展：i( i{ i[ i' i` 以及 aw/a"。
- [ ] M4.6 每完成一项：测试 + README 功能矩阵更新。→ 微信通报。

### M5 — 工程化收尾（09:00 前的最后半天做）

- [ ] M5.1 LICENSE 文件补上（README 已声称 MIT）。
- [ ] M5.2 GitHub Actions：build + vet + test + gofmt 检查。
- [ ] M5.3 README 更新：真实延迟数据、功能矩阵、安装说明与实际一致。
- [ ] M5.4 docs/ 整理：把 ARCHITECTURE-RECON.md 与本书章节引用互链。

## 完成判定（Definition of Done）

1. `go build ./... && go vet ./... && go test ./... -race` 全绿；
2. 功能矩阵（M1 验收 + M4 完成项）端到端通过；
3. 单键延迟实测并记录；
4. README/ONBOARDING/AUTH-CHAIN 与实现一致；
5. PROGRESS.md 全部打勾，git 历史干净（每任务一提交）。

## 给每个新会话的操作模板

```
1. cd /Users/ygs/ygs/Tmux-FSM && cat GOAL.md && cat PROGRESS.md
2. 取第一个未完成任务 → 实现 → go build ./... && go vet ./... && go test ./...
3. git add -A && git commit -m "..."（提交信息写清根因与方案）
4. 更新 PROGRESS.md 勾选 + 决策记录；里程碑完成则发微信
5. 重复，直到本轮时间片用完（不留未提交的半成品；宁可少做一任务）
```
