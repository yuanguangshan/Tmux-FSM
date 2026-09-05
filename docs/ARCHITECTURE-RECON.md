# Tmux-FSM 架构侦察报告（ARCHITECTURE-RECON）

> 生成日期：2026-09-06。只读侦察，未修改任何源码。
> 范围：`/Users/ygs/ygs/Tmux-FSM`（不含 `rhm-go/` vendored 子仓库），共 22,548 行 Go。
> 结论速览：这是一个 **6 代架构叠层**的半成品。当前唯一活跃的运行链路是
> `plugin.tmux → tmux-fsm -key → unix socket → kernel → planner(Grammar) → weaver → tmux send-keys`。
> 约 40% 的代码（root 包 gen-1/gen-2、engine+crdt 集群、pkg/、tools/）不在任何活跃路径上。
> Vim 语义在活跃链路上**大面积断裂**：w/b/e/gg/G/f{c}/文本对象/yank/undo/`.`/搜索全部无效果，
> 且 keymap.yaml 未绑定 i/a/o/p/x 等键 → 这些按键被守护进程静默吞掉。

---

## 1. 包依赖图

### 1.1 逐包清单（行数含 _test.go）

| 包 | 行数 | 一句话职责 | 被谁依赖 | 依赖谁 | 状态 |
|---|---|---|---|---|---|
| **root (package main)** | ~4,970 | 二进制入口：daemon socket server、FSM 状态持久化、以及 3 代遗留实现 | （二进制） | fsm, kernel, intent, editor, types, weaver/core, weaver/manager, backend | 核心（但内含大量死代码） |
| **fsm** | 1,057 | key-table 层引擎：keymap.yaml 加载/Dispatch/层超时、Enter/ExitFSM 原子操作、UI stub | main, kernel, planner, pkg/*, tests | backend, intent | **核心（活跃）** |
| **kernel** | 833 | 路由中枢：按键三分类决策（FSM 动作/Grammar 意图/Legacy）+ 执行器接口 | main, pkg/server, examples, tests | fsm, intent, intent/builder, planner, backend, editor, types, weaver/core, weaver/manager | **核心（活跃）** |
| **planner** | 1,009 | Vim 语法器：RawToken → GrammarIntent（count/op/motion/text-object） | kernel（+tests） | fsm, intent | **核心（活跃）** |
| **intent** | 495 | 统一 Intent 类型；全部常量类型别名到 weaver/core；`Promote` 是 GrammarIntent→Intent 唯一通道 | main, fsm, kernel, planner, weaver/manager, engine | weaver/core | **核心（活跃）** |
| **intent/builder** | 410 | CompositeBuilder 等语义等价/diff 构建器 | kernel（kernel.go:60） | intent | 冻结（创建了但 kernel.NativeBuilder 从未被使用） |
| **weaver/core** | 2,731 | ShadowEngine 执行管线：Adjudicate→Plan→Resolve→Apply→Verify→Evidence 审计 | intent, weaver/manager, weaver/logic, weaver/adapter, kernel, main, types | editor | **核心（活跃）** |
| **weaver/manager** | 403 | Weaver 编排器：InitWeaver(mode)、snapshot 获取、调用 engine.ApplyIntent | kernel, main | intent, weaver/{adapter,core,logic} | **核心（活跃）** |
| **weaver/adapter** | 1,371 | tmux 物理层：capture-pane 快照、send-keys 投影（PerformPhysical*）、RHM 桥 | weaver/manager, weaver/logic | weaver/core, editor, **rhm-go** | **核心（活跃）** |
| **weaver/logic** | 924 | Planner 实现 ShellFactBuilder + PassthroughResolver + 纯函数文本对象解析 | weaver/manager | weaver/core, weaver/adapter | **核心（活跃，text_object.go 未接线）** |
| **backend** | 273 | tmux 命令抽象（TmuxBackend / ControlModeBackend） | main, fsm, kernel | （无内部依赖） | 核心（活跃） |
| **editor** | 1,875 | 内存编辑模型：Buffer/Window/Selection store、ResolvedOperation、OperationDAG、文本对象 | types, weaver/core, weaver/adapter, engine, kernel/transaction, main, examples | （无内部依赖） | 半接线（仅 DAG/类型被 weaver 用） |
| **types** | 34 | Transaction/OperationRecord 定义 | main, kernel/transaction, examples | editor, weaver/core | 核心（活跃） |
| **engine** | 510 | 「三代」内存引擎（CursorEngine+执行器），import crdt/index/policy/replay/selection/wal | **无人导入** | editor, intent, crdt, index, policy, replay, selection, wal | **死代码（集群 A 入口）** |
| **crdt** | 316 | 语义事件 CRDT（PositionID/CausalParents） | engine, index, policy, replay, selection, verifier | semantic | 死代码（集群 A） |
| **semantic** | 294 | 语义 Fact/Anchor/Motion（集群 A 的类型层） | crdt, wal, index, policy, replay, verifier | （无） | 死代码（集群 A） |
| **wal** | 176 | 带校验和的语义事件追加日志 | engine, undotree | semantic | 死代码（集群 A） |
| **index** | 263 | 事件索引 + 区间树 | engine | crdt, semantic | 死代码（集群 A） |
| **policy** | 235 | 信任模型（TrustLevel: System/User/Device/AI/External） | engine | crdt, semantic | 死代码（集群 A） |
| **replay** | 151 | TextState 重放 | engine, verifier | crdt, semantic | 死代码（集群 A） |
| **selection** | 194 | 多光标选区模型 | engine | crdt | 死代码（集群 A） |
| **undotree** | 107 | 基于 WAL 的 undo 树 | **无人导入** | wal | 死代码（集群 A） |
| **verifier** | 292 | 状态根哈希校验 | tests/invalid_history_test.go、cmd/verifier（但 cmd/verifier 实际未 import） | crdt, replay, semantic | 死代码（集群 A） |
| **invariant** | 167 | 属性测试辅助（TextState 模型） | **无人导入** | （仅 testing） | 死代码（集群 A） |
| **ui** | 56 | Popup UI 抽象 | **无人导入** | （无） | 死代码 |
| **pkg/server** | 255 | 旧代 daemon（socket `~/.tmux-fsm.sock`，与 root main.go 的 `/tmp/tmux-fsm.sock` 重复） | **无人导入** | fsm, kernel | 死代码（集群 B） |
| **pkg/state** | 180 | 旧代状态管理 | **无人导入** | fsm | 死代码（集群 B） |
| **pkg/protocol** | 28 | 旧代协议 | **无人导入** | （无） | 死代码（集群 B） |
| **tools** | 2,165 | 文档生成器 gen-docs v3.1.0 + codoc v2.1.0（非产品代码） | （各自 main） | （无） | 独立工具（codoc 已加 build tag） |
| **cmd/verifier** | 43 | verifier CLI stub（`os.ReadFile` 后直接打印成功，核心逻辑全被注释） | （二进制） | （连 verifier 包都没 import） | 死代码 |
| **examples** | 118 | transaction_demo 示例 | （示例） | editor, kernel, types | 示例 |
| **tests** | 127 | 集成测试 + verifier 历史测试 | （测试） | fsm, intent, kernel, verifier | 测试 |

### 1.2 依赖方向（活跃链路）

```
main ──► kernel ──► fsm ──► backend
 │         │  └───► planner ──► intent ──► weaver/core ──► editor
 │         └──────► weaver/manager ──► weaver/{logic,adapter,core}
 └──────► types ──► editor, weaver/core
```

- **循环依赖嫌疑**：无真实循环（编译器已保证）。设计上刻意用「类型别名」绕开了两处险情：
  - `intent/intent.go:8,50,52-62,99,102` 把 IntentKind/TargetKind/SemanticTarget/Anchor 全部别名到 `weaver/core`，否则 `intent ↔ weaver/core` 会成环；
  - `types/types.go:6-7` 同时 import editor 和 weaver/core，自身保持叶子。
- **多代重复实现**（同一功能的第 N 代）：
  - **按键→语义**：gen1 `logic.go:460 processKeyLegacy`（action 字符串）→ gen2 `logic.go:26 processKeyToIntent`（builder.go IntentBuilder + intent_bridge.go 字符串↔Intent 互转）→ gen3 `planner/grammar.go Consume`（活跃）；
  - **Intent 类型**：`intent.go:5`（root）vs `intent/intent.go:84`（活跃）vs `weaver/core/interfaces.go:58`（接口）vs `semantic/capture.go`（死）；
  - **daemon**：`main.go:262 Server.Run`（/tmp/tmux-fsm.sock，活跃）vs `pkg/server/server.go:34`（~/.tmux-fsm.sock，死）；
  - **tmux 动作映射**：`kernel/kernel.go:176 executeAction` 与 `kernel/execute.go:62 getTmuxCommandForAction` 完全重复；`weaver/adapter/tmux_physical.go:9` 自述 "MIRROR OF execute.go"；
  - **count 累计**：`fsm/engine.go:211-218` 与 `planner/grammar.go:56-58` 各自一份（前者给 UI，后者给语义）；
  - **undo/redo**：`main.go:521 TransactionManager.Undo`、`kernel/transaction.go:47 TransactionRunner.Undo`、`weaver/core/shadow_engine.go:588 performUndoWithRequestID`、`undotree/tree.go`、`undo_redo.go:26 SnapshotManager` —— 五代并存，无一可用。
  - **fsm.EngineAdapter**（fsm/engine.go:17-124，给旧 resolver 的适配器，注释自述 "解析器已废弃"）—— 死代码，含大量 `return nil` 空实现。

---

## 2. 入口与运行时

### 2.1 main.go 起什么进程

`main.go:110-241`：
- `tmux-fsm -server`：写 PID 文件 `/tmp/tmux-fsm.pid`（main.go:227），在 `/tmp/tmux-fsm.sock` 起 unix socket daemon（main.go:262-285），每连接一个 goroutine `handleClient`（main.go:288-478）。协议两种：字符串 `requestID|paneID|clientName|key`（main.go:305-417）与 JSON Intent（main.go:420-477，仅 pkg 时代遗留，含 `__PING__`/`__SHUTDOWN__`/`__CLEAR_STATE__`/`reload`/`nvim-mode` 元命令）。
- `tmux-fsm -enter / -exit`：调 `fsm.EnterFSM()/ExitFSM()`（main.go:170-194 → fsm/enter_exit.go:10/35）。
- `tmux-fsm -key <key> [pane|client]`：`runClient`（client.go:42-137）连 socket 发 4 段协议，等待 "ok"。
- `tmux-fsm -reload`：`fsm.Reload`（fsm/engine.go:274，LoadKeymap+InitEngine+Reset 原子重建）。
- 初始化顺序（main.go:123-150）：`fsm.LoadKeymap` → `fsm.InitEngine` → `kernel.NewResolverExecutor` → `editor.NewExecutionContext`（内存 store，实际不接 tmux）→ `kernel.NewKernel(fsm.GetDefaultEngine(), resolverExecutor)` → `manager.InitWeaver(ModeWeaver)`（weaver/manager/manager.go:37-84：ShellFactBuilder + PassthroughResolver + TmuxSnapshotProvider + TmuxProjection + FileAppenderEvidenceLibrary `.weaver/evidence.log`，证据文件已 477KB）。

### 2.2 与 tmux 的交互方式

- **key-table 接管**（不是 display-popup）：`plugin.tmux:15-18` 绑定 `prefix+f` / `C-f` 进 FSM；`plugin.tmux:31-40` 循环给 `fsm` 键表绑定全部字母/数字/常用标点 → `tmux-fsm -key '<key>' '#{pane_id}|#{client_name}'`（每个按键 fork 一个 Go 进程走 socket）；`plugin.tmux:43-44` `Any` 兜底；`plugin.tmux:21-28` Escape/C-c/q 先 `switch-client -T root` 再 `tmux-fsm -exit`。
- **状态显示**：status-right 渲染 `#{@fsm_state}#{@fsm_keys}`（plugin.tmux:8），由 `updateStatusBar`（globals.go:114-201）经 backend 写 user option + `refresh-client -S`。
- **一致性 reconcile**：每次按键后 `reconcileFSMState`（main.go:741-809）按 `@fsm_active × client_key_table` 2×2 矩阵强制纠偏 key-table。
- **编辑执行**：`capture-pane`（快照/行内容，weaver/adapter/tmux_utils.go）+ `send-keys`（箭头/Home/End/M-f/M-d/C-k/C-w/BSpace 等，weaver/adapter/tmux_physical.go）+ `copy-mode`/`begin-selection`/`copy-pipe-and-cancel`（可视与 yank，tmux_physical.go:127-135,313-352）+ `set-buffer`/`paste-buffer`（粘贴，tmux_physical.go:473-481）。

### 2.3 安装流程

`install.sh`：
1. 杀旧 daemon（PID 文件 + pkill，install.sh:40-54）；
2. `go build -ldflags="-s -w" -o tmux-fsm .`（**只编译根包**，install.sh:79 —— 所以 tools/ 的编译失败不影响安装）；
3. 拷贝 plugin.tmux、三个 sh、keymap.yaml、二进制到 `~/.tmux/plugins/tmux-fsm/`（install.sh:93-108）；
4. 交互选 1/2/3：追加 `source-file plugin.tmux` 到 tmux.conf 或用 default.tmux.conf 整体替换（带备份），然后 `tmux source-file` 热加载并以 `TMUX_FSM_MODE=weaver` 预热 daemon（install.sh:147）。
`default.tmux.conf` 是一份"用户 tmux.conf 样例"（C-a 前缀、vi copy-mode 等），仅安装选项 2 使用。
注意 `TMUX_FSM_MODE` 环境变量实际无人读取（InitWeaver 由 main.go:150 硬编码 ModeWeaver）。

### 2.4 三个 sh 的角色

- `enter_fsm.sh`：copy-mode -q → 置 `@fsm_state=FSM`、`@fsm_active=1` → `switch-client -T fsm` → `tmux-fsm -enter`（双保险：shell 与 daemon 各做一遍状态+键表切换）。
- `fsm-exit.sh`：`@fsm_active=0` → `switch-client -T root` → `tmux-fsm -exit`。**plugin.tmux 并未引用它**（plugin.tmux:21-28 直接内联 Escape/C-c/q 退出），是 fsm-toggle.sh 的库。
- `fsm-toggle.sh`：按 `@fsm_active` 翻转；进入时 `repeat-time 0`（配合 chord 式按键），退出时恢复 500。install.sh 也没拷它……实际上 install.sh:94 拷了 fsm-toggle.sh，但 plugin.tmux 只用 enter_fsm.sh —— toggle 入口未挂接任何 tmux 绑定（原 prefix+f 的切换语义退化成了单向进入）。

---

## 3. 按键处理全链路（活跃路径，带证据）

```
用户按键
└─ tmux fsm key-table（plugin.tmux:31-44）
   └─ run-shell: tmux-fsm -key '<key>' '<pane>|<client>'
      └─ client.go:42 runClient —— 组 4 段协议 requestID|paneID|client|key（client.go:118）
         └─ unix socket /tmp/tmux-fsm.sock
            └─ main.go:288 handleClient —— 解析协议（main.go:305-354）；__PING__/__CLEAR_STATE__ 特判（main.go:359-371）
               ├─ main.go:383-392 读 @tmux_fsm_state → 更新 Mode/Count/Operator → saveFSMState
               ├─ main.go:407 updateStatusBar（globals.go:114）
               └─ main.go:413 reconcileFSMState（main.go:741）—— @fsm_active×key_table 纠偏
               └─ kernel.HandleKey（kernel/kernel.go:65-105）—— 要求 RequestID/ActorID 非空
                  └─ kernel.Decide（kernel/decide.go:52-119）
                     ├─ ① CanHandle 且有 action → DecisionFSM（pane_left/exit/prompt 等，execute.go:62 映射 tmux 命令）
                     ├─ ② fsm.Dispatch（fsm/engine.go:210-248：digit 累计 / "." → TokenRepeat / 层切换 / action）
                     │    └─ RawToken → GrammarEmitter.Emit（decide.go:45-50）
                     │       └─ planner.Grammar.Consume（planner/grammar.go:53-77）→ consumeKey（grammar.go:80-246）
                     │          （pendingMotion > textObj > operator(dd) > i/a > g/f/F/t/T 前缀 > motion > 模式切换 > ;/, > u/C-r）
                     │       └─ intent.Promote（intent/promote.go:5-34）→ DecisionIntent
                     └─ ③ 未被 FSM 吃掉 → DecisionLegacy（decide.go:116-118）
                  └─ kernel.Execute（kernel/execute.go:10-45）
                     ├─ DecisionFSM → executeFSMAction → backend.ExecRaw（tmux send-keys/select-pane）
                     ├─ DecisionNone/DecisionLegacy → 什么都不做 ← ★ 未绑定按键在此被吞
                     └─ DecisionIntent → ResolverExecutor.Process（kernel/resolver_executor.go:20-41）
                        └─ weaver/manager.ProcessIntentGlobalWithContext（manager.go:228-251）
                           ├─ TmuxSnapshotProvider.TakeSnapshot（adapter/tmux_snapshot.go:9-19：cursor+capture-pane）
                           └─ ShadowEngine.ApplyIntent（core/shadow_engine.go:39-544）
                              ├─ Adjudicate：SnapshotHash 世界漂移检查（乐观放行，shadow_engine.go:60-103）
                              ├─ Plan：ShellFactBuilder.Build（logic/shell_fact_builder.go:11-162）
                              │    仅支持 IntentInsert/IntentDelete/IntentMove/IntentOperator(op=Delete|Change)/EnterVisual/ExitVisual
                              │    ← ★ IntentYank/IntentRepeat/IntentRepeatFind/搜索/粘贴 无 case → 0 facts
                              ├─ Resolve：PassthroughResolver.ResolveFacts（logic/passthrough_resolver.go:15-97）
                              │    （行哈希不匹配 → 降级 Fuzzy 放行，passthrough_resolver.go:114-119）
                              ├─ Apply：TmuxProjection.Apply（adapter/tmux_projection.go:13-137）
                              │    FactDelete→PerformPhysicalDelete / FactInsert→Insert|Paste|RawInsert / FactMove→PerformPhysicalMove
                              │    FactNone→search_*/find_*/visual_*/exit / 写冲突检测 detectProjectionConflicts（tmux_projection.go:179-248）
                              │    前后 capture 该行生成 UndoEntry{Before,After}（tmux_projection.go:128-134）
                              ├─ Verify：LineHashVerifier（tmux_projection.go:158-166）
                              └─ History.Push + ProofBuilder + OperationDAG + EvidenceLibrary.Commit（shadow_engine.go:448-541）
                                 证据落盘 .weaver/evidence.log（manager.go:64-73）
```

死链路（保留在仓库但不运行）：JSON Intent 入口（main.go:420-477 → ProcessIntentGlobal main.go:693 → weaverMgr.Process）、gen-1/gen-2 root 包整条（见 §1.2）。

---

## 4. 构建与测试现状

### 4.1 go build ./...

**当前状态：通过（exit 0）。**
- 已知问题「tools/gen-docs.go 与 tools/codoc.go 同为 package main、重复声明 versionStr/Config/FileMetadata/Stats/DirStats/ExtStats/languageMap/main/parseFlags/splitAndTrim」——根因：新旧两代文档生成器同居一个包。
- **最小修复（已于本次侦察期间被并发会话应用）**：tools/codoc.go:1-2 已加 `//go:build ignore`（文件 mtime 2026-09-06 05:45，注释明说 "两者曾同居 package main 造成重复声明"）。侦察早期 go build 仍报 22 行错误、随后复测通过。
- 若要更彻底：把 codoc.go 移入 `tools/codoc/` 子目录或加独立 module；根包 `go build .`（install.sh:79 所用）从未受影响。

### 4.2 go vet ./...

- `tmux-fsm/fsm`：`fsm/engine_test.go:76: assignment mismatch: 1 variable but engine.Dispatch returns 2 values`（其余包 vet 干净）。

### 4.3 go test ./... 清单

| 包 | 结果 | 分类 | 根因 |
|---|---|---|---|
| tmux-fsm（root） | ok | — | globals_test/config_test/main_comm_test/resolver_integration_test 测的是 gen-1/gen-2 死代码，自成闭环 |
| tmux-fsm/intent | ok | — | — |
| tmux-fsm/kernel | ok | — | — |
| tmux-fsm/planner | ok | — | grammar_test.go 367 行，语法层唯一像样的测试 |
| tmux-fsm/weaver/core | ok | — | — |
| tmux-fsm/weaver/manager | ok | — | — |
| tmux-fsm/fsm | FAIL [build failed] | **编译失败** | `engine.Dispatch` 签名从 `bool` 改为 `(string,bool)`（fsm/engine.go:210）后测试未更新：engine_test.go:76 `result := engine.Dispatch("h")`；同型调用还有 :121、:147、:152、:158 等。**最小修复**：改为 `_, result := engine.Dispatch(...)`（如需断言 action 再接收第一个返回值） |
| tmux-fsm/tests | FAIL | **断言失败** | `TestKernelGrammarIntegration`（tests/integration_test.go:68）：2dw 产生 IntentOperator/Count=2 正确，但 `PaneID` 期望 "p1" 实为 ""。根因：PaneID 注入只存在于 `ProcessIntentWithContext`（kernel/kernel.go:126-130），而按键路径走 `Decide→intent.Promote→Execute→Exec.Process`（decide.go:97-103, execute.go:32），全程无人填 PaneID。**最小修复**：在 decide.go Promote 之后、构造 Decision 前，从 `k.FSM` 无关的 hctx 拿不到——需把 hctx 传入 Decide 或在 `HandleKey`（kernel.go:82-83）中 `decision.Intent.PaneID = actorID 前 1 段` |
| tmux-fsm/weaver/adapter | FAIL | **断言失败（依赖 rhm-go）** | `TestRHMAdapter_Solve`（rhm_adapter_test.go:66-68）：期望冲突消解代价 50，实得 0。求解器在 vendored `rhm-go/core/solver/solver.go:73`（TotalCost=int(current.Cost)），测试用 mockOpWrapper 以字符串前缀 "Edit:"/"Delete:" 触发冲突分析。代价模型权重或 mutation 匹配与测试契约不符。**最小修复**：先输出 plan.Narrative.Mutations 确认 0 变更是 "未识别冲突" 还是 "代价为 0"；若是前者，修 mock 描述串与 solver 的约定，若后者修期望值 |
| tmux-fsm/tools | [no test files]（曾 build failed，现修复） | — | — |
| 其余 20 包 | no test files | **覆盖缺口** | editor（1,875 行）、backend、weaver/adapter 的物理层、fsm 运行层全部无测试；tests 需要 tmux 环境的集成用例不存在（tests/ 里只有纯内存集成） |

无需 tmux 环境：现有全部测试均为纯内存测试，不会因无 tmux 而失败。

---

## 5. 功能完成度矩阵（以活跃链路为准）

图例：✅已实现｜🟡半成品｜🔴损坏（走了但结果错）｜❌缺失（按键被吞/无实现）。"gen-X" 表示实现存在于死代码的第 X 代。

| 功能 | 状态 | 证据与说明 |
|---|---|---|
| 进入 FSM（prefix+f / C-f） | ✅ | plugin.tmux:15-18 → enter_fsm.sh（key-table+@fsm_active+daemon 三同步） |
| 退出 FSM（Esc/C-c/q） | ✅ | plugin.tmux:21-28 + fsm/enter_exit.go:35-55 + reconcileFSMState（main.go:741-809） |
| 状态栏显示（模式/算符/计数） | ✅ | globals.go:114-201、fsm/ui_stub.go:38-65、plugin.tmux:8 |
| count（3j、2dw、d2w） | 🟡 | FSM 与 Grammar 双份累计（fsm/engine.go:211-218、planner/grammar.go:56-58）；hjkl 的 count 生效（tmux_physical.go:100-107 `-N`）；**delete/yank 的 count 被丢弃**：PerformPhysicalDelete 无 count 参数（tmux_physical.go:138）；Promote 只在 motion.Count>1 时写 meta（intent/promote.go:122-125） |
| 动词 d（dd/dw/d0/d$/x） | 🟡 | dd→C-a C-k Delete（tmux_physical.go:173-175）；dw→M-d（:159-161）；d0→BSpace×cursorX（:146-152）；d$→C-k（:154-157）。但全部是 shell 行编辑近似，非真实 Vim 语义；gen1/gen2 有更完整的 action 版（logic.go:510-618）未接线 |
| 动词 c | 🟡 | Grammar op==3 → FactInsert{operation:change} → PerformPhysicalDelete + ExitFSM（tmux_projection.go:72-84）——删除后退出 FSM 让按键直通（近似进入 insert），cc/cw 同 d 的近似 |
| 动词 y（yy/yw） | 🔴 | planner 无 case：shell_fact_builder.go:110-133 只处理 op==1(D)/op==3(C)，op==2(Yank) 落空 → 0 facts。PerformPhysicalTextObject 里的 yank 分支（tmux_physical.go:228-233）**无任何调用方** |
| 移动 h j k l | ✅ | grammar.go:282-289 Direction → Promote left/right/up/down（promote.go:46-56）→ 箭头键（tmux_physical.go:100-107） |
| 移动 0 $ ^ | ✅ | grammar.go:290-295 RangeLineStart/End → goto_line_start/end（promote.go:106-114）→ Home/End（tmux_physical.go:108-111） |
| 移动 w b e | 🔴 | parseMotion 归 MotionWord（grammar.go:612）但 makeMoveGrammarIntent 不设 Direction（grammar.go:280-300 无 w/b/e 分支）→ Promote MotionWord 分支空 Direction → motionStr=""（promote.go:57-63）→ PerformPhysicalMove("") 无 case → 无动作。物理层其实有 M-f/M-b（tmux_physical.go:113-118），只差 meta 没填上 |
| 移动 gg / G | 🔴 | 同上：MotionGoto 无 Direction → Promote 留空（promote.go:73-89 注释自认 "For now let's leave it as is"）→ goto 无效；物理层 start_of_file/end_of_file（tmux_physical.go:119-122）够不着 |
| f F t T（含 ; ,） | 🔴 | f{c} 生成 MotionFind+Char（grammar.go:530-561），但 Promote 只产出 meta["motion"]=find_char_forward（promote.go:90-105），**Char 没进 meta**，且 FactMove→PerformPhysicalMove 无 find 分支；带 find_type/find_char 的 PerformPhysicalFind（tmux_physical.go:238-310）挂在 FactNone 分支（tmux_projection.go:116-119）永远够不着。;/, → IntentRepeatFind 无 planner case → 无动作 |
| 文本对象 iw aw i" a" i( 等 | 🟡 | Grammar 识别 i/a + w()b[]{}B"'`（grammar.go:408-424,498-527）→ IntentOperator{MotionRange{TextObject}}，但 **Promote 对 RangeTextObject 不产任何 meta**（promote.go:106-114 只处理 LineStart/LineEnd）→ FactDelete motion="" → PerformPhysicalDelete 走 default→M-d（tmux_physical.go:177-179），diw 碰巧≈删一个词。真实现三处齐备却互不相连：PerformPhysicalTextObject（tmux_physical.go:184-235，无调用方）、weaver/logic/text_object.go（LocRange 解析器 434 行，无调用方）、editor/text_object.go（537 行，仅 editor 包内） |
| 寄存器 "x / p P | ❌ | `"` 不在 keymap.yaml → 被吞；Register 字段只有 gen1 handleRegisterSelect（logic.go:671-675，死）；p/P 不在 keymap；PerformPhysicalPaste（tmux_physical.go:52-60）仅被 motion=="paste" hack 触发（tmux_projection.go:61-62），无 planner 产该 meta |
| undo u | 🔴（审计空转） | u→IntentUndo→performUndoWithRequestID（shadow_engine.go:588-684）：PopUndo 后 Resolve **InverseFacts**，而 InverseFacts 恒为空（shell_fact_builder.go:161 返回空 slice；shadow_engine.go:322 原样入账）→ 0 事实可应用。UndoEntry{Before,After} 在 Apply 时已捕获（tmux_projection.go:128-134）但 Rollback 是空壳（tmux_projection.go:141-155）。main.go:522 与 kernel/transaction.go:47 直接返回 "not implemented" |
| redo C-r | 🔴 | performRedoWithRequestID（shadow_engine.go:891）同样依赖空 InverseFacts；C-r 在 keymap 有绑定（keymap.yaml:36），全程空转 |
| 可视模式 v / V | 🟡 | v→IntentEnterVisual（grammar.go:181-213）→ FactNone visual_enter → HandleVisualAction → copy-mode+begin-selection（tmux_physical.go:338-343）。**V 与 v 不可区分**（grammar.go:194-202 TODO 注释）；选中后 d/y 无可视态感知（fsm.Engine.visualMode 存在但 EngineAdapter 死，fsm/engine.go:27-45）；vim pane 内发 d/y/c 键的分支（tmux_physical.go:324-337）需要 visual_delete 等 meta，活跃链路产不出 |
| 搜索 / ? n N | ❌ | `/` `?` `n` `N` 均不在 keymap.yaml → FSM 吃掉后 DecisionLegacy 丢弃。搜索实现只存在于 gen1（logic.go:758-792 handleSearch）与 PerformExecuteSearch（tmux_physical.go:127-135，需 meta operation=search_forward，无 planner 产出） |
| 重复 . | ❌ | Dispatch 特判 "."（fsm/engine.go:220-223）→ IntentRepeat → Promote → planner 无 case → 0 facts。gen2 的 RepeatLastTransaction（main.go:498-517，事务重放模型完整）与 kernel TransactionRunner.Repeat（kernel/transaction.go:51-54）均未接线 |
| 插入系 i a I A o O | ❌（最高危） | keymap.yaml NAV（keymap.yaml:4-39）**没绑 i/a/o/I/A/O** → 按键被守护进程吞掉（decide.go:116-118 DecisionLegacy → execute.go:22-23 no-op）。且 grammar.go:116-123 把 NORMAL 下的 "i"/"a" 先吃成 text-object 前缀，连 parseModeSwitch 的 "insert" 分支（grammar.go:251-252）都不可达。gen1/gen2 的 insert builder（logic.go:178-193）与 PerformPhysicalInsert（tmux_physical.go:32-49）都在睡觉。**后果：FSM 模式下无法开始输入文字** |
| x X D C S r ~ s | ❌ | 不在 keymap（除 S？无，均无）→ 被吞；gen2 builder 支持（logic.go:151-197） |
| pane 导航（hjkl 切 pane） | ✅（FSM 动作路径） | kernel/execute.go:63-95 select-pane -L/R/U/D 等；但 keymap.yaml NAV 的 h/j/k/l action 为空（让给 Grammar 移动光标），GOTO 层（keymap.yaml:41-49）的 far_left 等需要先到 GOTO 层，而 NAV 没有 `layer: GOTO` 绑定 → GOTO 层不可达 |
| Neovim 模式联动 | 🟡 | fsm/nvim.go:8-22 OnNvimMode（insert/visual→ExitFSM），入口在 JSON 协议 `nvim-mode`（main.go:463-470）；无任何发送方配置（NotifyNvimMode 是空函数 nvim.go:18-22） |
| 审计/证据库 | ✅ | .weaver/evidence.log 追加式审计（manager.go:64-73, interfaces.go:20-27），shadow_engine 每个 Phase 写 AuditEntryV2——这是仓库里完成度最高的子系统，但对用户功能无直接贡献 |

---

## 6. 修复顺序建议（依赖排序）

**第 0 层（编译与测试基建，半天）**
1. tools 冲突：确认 `//go:build ignore`（tools/codoc.go:1）已入库；或把 codoc.go 移到 `tools/codoc/`。
2. fsm 测试编译：engine_test.go 全部 `engine.Dispatch(...)` 双返回值适配（fsm/engine_test.go:76,121,147,152,158）。
3. PaneID 注入：在 `HandleKey` 对 DecisionIntent 补 `Intent.PaneID`（修 tests/integration_test.go:68）；这是后续所有按 pane 执行的前置。

**第 1 层（按键存活，核心，先于一切功能）**
4. **keymap.yaml 补全或放行策略**：两选一 —— (a) NAV 层补绑 i/a/o/I/A/O/x/p/P///?/n/N 等（action 留空交 Grammar）；(b) 在 kernel/execute.go 对 DecisionLegacy 改为 `send-keys` 原键回 pane（更 Vim：未知键透传）。没有这一步，其余修复用户都感知不到。
5. Grammar "i/a" 吞键 bug：consumeKey 中 text-object 分支应仅在 pendingOp 存在时进入（planner/grammar.go:116-123 现在无条件吃 i/a）。

**第 2 层（语义完整：一次改动解锁一批功能）**
6. **Promote/meta 单一事实源**：把 Direction/Range/Goto/find/text_object/count 统一在 `intent.Promote` 填全（promote.go:36-126），删除三处重复映射（promote.go、logic/shell_fact_builder.go:166-181 populateMotionMeta 空转、tmux 字符串散在 tmux_physical.go）。修完后 w/b/e、gg/G、f/t/F/T、文本对象、count 五组同时复活（物理层已有对应实现）。
7. yank：shell_fact_builder 补 op==2 → FactYank，Projection 用 capture→set-buffer 落 tmux buffer；顺带 p/P。
8. undo：把 TmuxProjection.Apply 已捕获的 UndoEntry{Before,After}（tmux_projection.go:128-134）转成 InverseFacts 存入 Transaction（shadow_engine.go:322），并实现 Rollback 真逻辑；redo 同理。u/C-r 在 keymap 已就绪，只差这一段。
9. `.` repeat：活跃链路 Transaction 已入 weaver History（shadow_engine.go:451），在 planner/projection 层补 IntentRepeat → 重放上一事务的 facts。

**第 3 层（模式与搜索）**
10. 插入直通：i/a/I/A/o/O → ExitFSM（复用 fsm.ExitFSM 原子操作），o/O 需 send-keys End/Enter。
11. 可视模式：GrammarIntent 带 VisualMode（grammar.go:186-202 的 TODO），fsm.Engine.visualMode 与 FSM 动作联动，选中后 d/y/c 走 copy-mode -X。
12. 搜索：/ 进入 SEARCH 态（gen1 logic.go:758-792 可整体移植），n/N 映射 copy-mode search-forward/previous。

**第 4 层（清理包袱，功能稳定后）**
13. 删除/封存死集群（见 §7 清单），预计可减 ~8,000 行；root 包把活跃文件（main/client/globals/updateStatusBar/reconcile）与其余 gen-1/gen-2 文件拆开。
14. 合并 kernel/kernel.go:176 executeAction 与 kernel/execute.go:62（重复映射表）。
15. 补 weaver/adapter 物理层与 fsm 层的 tmux 环境集成测试（现有 0 个）。
16. rhm-go 依赖决策：weaver/adapter 仅 rhm_adapter.go/mapToDAG/Solve 用于冲突消解（rhm_adapter.go:110-112），活跃链路并不调用 Solve → 可冻结或移除该测试期望。

### 保留 / 冻结 / 删除清单

| 处置 | 包/文件 |
|---|---|
| **保留（核心）** | fsm, kernel, planner, intent, weaver/{core,manager,adapter,logic}, backend, editor（仅 types/DAG/stores 被 core 引用）, types, root 的 main/client/globals/logic(updateStatusBar 部分)/config |
| **冻结（先别动，等第 2 层接线）** | weaver/logic/text_object.go、weaver/adapter/tmux_physical.go 的 PerformPhysicalTextObject/Find/ExecuteSearch、intent/builder、kernel/transaction.go、main.go 的 TransactionManager/History/MacroManager（. 重复的实现基础）、fsm/ui_stub.go |
| **删除/归档（死代码）** | engine/, crdt/, semantic/, wal/, index/, policy/, replay/, selection/, undotree/, verifier/, invariant/, ui/, pkg/{server,state,protocol}/, cmd/verifier/, tools/（独立仓库化）, root 的 engine.go, resolver.go, resolver_text_objects.go, logic.go 的 processKeyLegacy 段, intent.go(root), intent_bridge.go, builder.go(root), transaction.go(root), undo_redo.go, snapshot.go, protocol.go(root), fsm/engine.go 的 EngineAdapter 段（:17-124） |

---

## 附：关键交叉验证用文件:行 索引

- 活跃入口：main.go:110,219,262,288；client.go:42,118
- 决策：kernel/decide.go:52,95,116；kernel/execute.go:22,32,62
- 语法：planner/grammar.go:53,80,116,161,249,408,593,606
- 意图提升：intent/promote.go:5,38（MotionWord:57，MotionGoto:73，Find:90，Range:106）
- 执行：weaver/manager/manager.go:37,228；weaver/core/shadow_engine.go:39,118,176,322,448,588,891
- 规划：weaver/logic/shell_fact_builder.go:11,110-133,161；weaver/logic/passthrough_resolver.go:15,114
- 物理：weaver/adapter/tmux_projection.go:13,46,72,100,128,141,179；tmux_physical.go:32,94,127,138,159,173,184,238,313,355,473
- 快照：weaver/adapter/tmux_snapshot.go:9；weaver/adapter/tmux_utils.go
- key-table/UI：fsm/enter_exit.go:10,35；fsm/ui_stub.go:25,80；fsm/keymap.go:41；main.go:741；globals.go:62,92,114
- 安装：install.sh:40,79,93,129,147；plugin.tmux:15,21,31,43,47,55
- 测试失败点：fsm/engine_test.go:76；tests/integration_test.go:68；weaver/adapter/rhm_adapter_test.go:67
