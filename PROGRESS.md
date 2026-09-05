# PROGRESS — 执行日志

> 每轮 cron 会话：完成一个任务后在此打勾并追加一行日志。
> 格式：`- [x] M0.1 (会话时间) 备注`
> IN_PROGRESS 协议：开始任务前把「当前任务」行填上；完成后清空。
> 若本轮读取时发现 IN_PROGRESS 已挂 25 分钟以上，视为上轮中断，可直接接手。

当前任务: （无）
当前里程碑: M0

## 任务队列状态

- [ ] M0.1 tools 重复声明
- [ ] M0.2 fsm/engine_test.go 签名适配
- [ ] M0.3 integration PaneID
- [ ] M0.4 weaver/adapter 期望
- [ ] M0.5 gofmt + 大文件出库
- [ ] M0.6 M0 验收（build/vet/test 全绿）
- [ ] M1.1 exit 死路
- [ ] M1.2 计数器生命周期
- [ ] M1.3 未知键透传 Grammar
- [ ] M1.4 Direction 元数据
- [ ] M1.5 f/t/F/T 目标字符
- [ ] M1.6 M1 验收（端到端回归）
- [ ] M2.1 HandleKey 串行化 + race 测试
- [ ] M2.2 daemon 单实例 Flock
- [ ] M2.3 install.sh pkill 精确化
- [ ] M2.4 双 UI 写入者合并
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

## 日志

- 2026-09-06 05:0x GOAL.md / PROGRESS.md 建立，循环启动。
