# FOEK (Fact-Oriented Editing Kernel) Manifesto

## 1. 意图与事实的二元论 (Intent vs. Facts)

在 FOEK 中，编辑不是“按键的模拟”，而是“意图对空间事实的投影”。

- **Action (Intent)**: 用户想做什么（例如：删除一个单词）。
- **Range (Facts)**: 在特定物理时间点，空间中受影响的真实范围。

## 2. 空间事实的一等公民地位

`Range` 不仅仅是选区，它是“历史的存证”。

- 它记录了 `start_hint` (物理锚点) 和 `content` (生存记录)。
- 所有的撤销 (Undo) 都是对空间事实的 **Reconstruction (重建)**，而不是对命令的 **Inversion (反转)**。

## 3. 跨环境的“尽力而为”一致性 (Best-effort Consistency)

内核是纯粹的语义层。Vim, Shell, Tmux 只是 Adapter。

- 虽然 Shell 不支持随机访问，但内核通过 `jump_to_pos` 建立了“空间线索层”。
- 语义在内核中是 100% 精确的，物理回放随 Adapter 的能力而“优雅降级”。

## 4. 时间线即因果链 (Timeline as Causality)

- **Undo**: 执行空间事实的逆向重建。
- **Redo**: 重新激活原始动作意图 (Intent Replay)，而不是反向执行撤销动作。

## 5. 多光标的降维打击 (Multi-cursor as Batch Facts)

在 FOEK 中，多光标不是一种“模式”，而是 `List[Range]` 的并发迭代。

- 当一个 Action 作用于多个 Range 时，多光标自然浮现。
- 不需要复杂的 UI，只需要对“事实批处理”的掌控。
