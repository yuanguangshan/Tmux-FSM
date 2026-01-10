# RHM-Go Stability & Semantic Manifesto (v1.0)

本文档定义了 RHM (Reversible History Model) Go 引擎的语义边界与 API 稳定性承诺。

---

## 1. Public Core API (Stable)
**承诺**：保证 SemVer 兼容性。这些包是 RHM 语言的公理系统，不会轻易修改。

- **`core/history`**: 历史图 (DAG) 的基本结构与读取接口。
- **`core/change`**: `ReversibleChange` 接口与 `Mutation` 原子定义。
- **`core/solver`**: `Solve` 函数及其输入输出 Contract。
- **`core/cost`**: 成本模型接口与 SLU (Semantic Logical Unit) 基础常量。

---

## 2. Semi-Stable Extension Points
**承诺**：允许扩展。为了功能进化支持修改内部字段，但在次要版本更新时会提供迁移指引。

- **`core/change` (Implementation)**: 具体业务的操作实现（如 `store/ops.go` 中的 FileSystemOp）。
- **`core/narrative`**: 叙事数据结构。支持添加更丰富的决策元数据。
- **`internal/formatter`**: 报告渲染层。支持自定义 Markdown/HTML 展现形式。

---

## 3. Experimental / Internal (Unstable)
**承诺**：不保证稳定性。这些模块属于“黑盒”实现，可能随时重构以进行性能优化或算法升级。

- **`core/analysis`**: 冲突检测的具体启发式算法。
- **`core/search`**: A* 搜索的内部状态管理与指纹计算。
- **`internal/loader`**: 测试场景加载逻辑。

---

## 4. Determinism & Integrity Guarantees
**公理声明**：

- **确定性 (Determinism)**：在给定相同的 `HistoryDAG` 和 `CostModel` 的情况下，RHM 引擎保证产生**比特级别一致**的解决方案。
- **因果一致性 (Causal Consistency)**：所有被选中的解决方案必须在因果上自洽，且所有变更均为可逆（Reversible）。
- **叙事真实性 (Narrative Truth)**：Narrative 报告不仅是 UI 展示，它是搜索路径的真实转录，必须反映 Solver 的真实决策过程（包含被拒绝的备选方案）。

---

## 5. Solver Contract
> **"We don't tell the system the answer; we define the value space and let the system derive the truth."**

RHM Solver 必须始终严格遵守 `core/solver/solver.go` 中定义的四项基本契约。
