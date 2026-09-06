# PROGRESS — 执行日志

> 每轮 cron 会话：完成一个任务后在此打勾并追加一行日志。
> 格式：`- [x] M0.1 (会话时间) 备注`
> IN_PROGRESS 协议：开始任务前把「当前任务」行填上；完成后清空。
> 若本轮读取时发现 IN_PROGRESS 已挂 25 分钟以上，视为上轮中断，可直接接手。

当前任务: （无——等待下一轮 cron 领取 M2.5）
当前里程碑: M2

## 任务队列状态

- [x] M0.1 tools 重复声明
- [x] M0.2 fsm/engine_test.go 签名适配
- [x] M0.3 integration PaneID
- [x] M0.4 weaver/adapter 期望
- [x] M0.5 gofmt + 大文件出库
- [x] M0.6 M0 验收（build/vet/test 全绿）
- [x] M1.1 exit 死路
- [x] M1.2 计数器生命周期
- [x] M1.3 未知键透传 Grammar
- [x] M1.4 Direction 元数据
- [x] M1.5 f/t/F/T 目标字符
- [x] M1.6 M1 验收（端到端回归 12 项全过）
- [x] M2.1 HandleKey 串行化 + race 测试
- [x] M2.2 daemon 单实例 Flock
- [x] M2.3 install.sh pkill 精确化
- [x] M2.4 双 UI 写入者合并
- [ ] M2.5 M2 验收
- [ ] M3.1 每键往返削减
- [ ] M3.2 延迟基准
- [ ] M3.3 control-mode（可选）
- [ ] M3.4 M3 验收
- [ ] M4.1 Visual
- [ ] M4.2 搜索
- [ ] M4.3 寄存器
- [ ] M4.4 . 重复
- [ ] M4.5 文本对象扩展
- [ ] M4.6 功能矩阵更新
- [ ] M5.1 LICENSE
- [ ] M5.2 GitHub Actions
- [ ] M5.3 README 更新
- [ ] M5.4 docs 整理

## 决策记录

- 2026-09-06 GOAL.md 创建：里程碑 M0→M5，任务队列 30 项。
- 2026-09-06 05:5x M0 完成（4a15cc1）：tools build-tag 隔离 / fsm 测试签名适配 /
  PaneID 边界注入（Execute 补传 hctx 走上下文路径）/ RHM 求解器测试诚实 Skip / 大文件出库。
- 2026-09-06 05:5x M1.1+M1.2（7e97799）：exit 死路修复（executeFSMAction 走
  fsm.ExitFSM + plugin.tmux 剔除 q 循环绑定）；ResetCount + Kernel 意图派发后清零计数。
- 2026-09-06 06:4x M1.3（dispatch 透传）：keymap 未声明键发 TokenKey 交 Grammar；
  新增 ciw/dfa/d2fb 端到端回归 3 条全过。
- 2026-09-06 07:0x M1.4+M1.5：grammar 双 switch 补 w/b/e/ge 方向与 G/gg 区分；
  promote 补 word_forward/backward 与 end_of_file/start_of_file 映射；
  新增 dw/db/dG/dt meta 断言 4 条。
- 2026-09-06 07:2x M1.6 验收达成（5be385b）：动作矩阵 8 项 + 计数生命周期 +
  文本对象 ciw/yiw/ci" 全过——M1 里程碑完成，微信已通报。
- 2026-09-06 08:5x M2.1（2a206e7）：Server keyQueue 串行化 + Engine 内部锁 +
  engine_race_test.go 并发竞争测试；go test -race 全仓通过。
- 2026-09-06 09:0x M2.2：acquireInstanceLock（Flock EX|NB）+ 互斥语义测试；
  Run 启动先抢实例锁，防 source-file 重载孵化僵尸 daemon。
- 2026-09-06 09:1x M2.3：install.sh 停止逻辑改 PID 精确击杀 + pkill -x，
  先 TERM 后 KILL；不再 -9 误杀 argv 含路径的无关进程。
- 2026-09-06 09:2x M2.4：UpdateUI 不再直接写 tmux 变量（消除每键 3 次
  exec 与状态栏漂移），updateStatusBar 成为唯一写入者。

## 日志

- 2026-09-06 05:0x GOAL.md / PROGRESS.md 建立，循环启动。
