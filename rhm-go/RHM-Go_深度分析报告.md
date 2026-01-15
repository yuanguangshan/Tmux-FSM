# RHM-Go 子模块深度分析报告

## 1. 项目概述

### 1.1 项目定位

**RHM (Reversible History Model)** 是一个创新的**因果感知版本控制与合并引擎**，采用语义推理而非传统的文本行比较来解决冲突。这是对传统版本控制系统（如 Git）的范式转移。

**核心哲学**：
- 不比较文本行，而是推理历史意图
- 通过"平行宇宙"计算寻找最优合并方案
- 所有决策都可审计、可追溯、可逆

### 1.2 项目元信息

- **语言**：Go 1.23+ (toolchain go1.24.0)
- **模块名**：rhm-go
- **代码规模**：
  - 核心代码：773 行
  - 总代码量：1,127 行
  - 40 个函数定义（核心层）
- **依赖项**：
  - CLI 框架：cobra v1.8.0
  - 监控：prometheus/client_golang v1.23.2
  - 哈希算法：xxhash v2.3.0

---

## 2. 目录结构与职责分析

```
rhm-go/
├── api/http/          # HTTP API 服务层
├── cmd/               # 命令行工具入口
├── core/              # 核心引擎（9个子模块）
│   ├── analysis/      # 冲突分析引擎
│   ├── change/        # 代数底座（操作定义）
│   ├── cost/          # 语义代价模型
│   ├── history/       # 历史 DAG 管理
│   ├── narrative/     # 决策审计报告
│   ├── rewrite/       # 临时沙箱（平行宇宙）
│   ├── scheduler/     # 冲突调度优先级
│   ├── search/        # A* 搜索算法
│   └── solver/        # 主求解器
├── internal/          # 内部工具
│   ├── formatter/     # 报告渲染（Markdown/HTML）
│   └── loader/        # 测试场景加载器
├── store/             # 业务操作实现（文件系统示例）
└── telemetry/         # Prometheus 监控指标
```

---

## 3. 核心功能深度剖析

### 3.1 RHM 的核心概念

**时间旅行物理定律** (`core/change/change.go`)：

```go
// 定义操作在资源空间留下的痕迹
type Footprint struct {
    ResourceID string
    Mode       AccessMode  // Shared | Exclusive | Create
}

// 可逆变更接口 - 所有操作必须实现
type ReversibleChange interface {
    Describe() string           // 人类可读描述
    ToNoOp() ReversibleChange   // 中和为空操作
    Downgrade() ReversibleChange // 降级为温和操作
    Hash() string               // 指纹计算
}
```

**访问模式代数**：
- `Shared`：共享访问（读操作，如 Edit）
- `Exclusive`：独占访问（写/删，如 Delete）
- `Create`：命名空间占用（新建资源）

**互斥矩阵**：
```go
func isMutuallyExclusive(m1, m2 AccessMode) bool {
    if m1 == Exclusive || m2 == Exclusive { return true }
    if m1 == Create && m2 == Create { return true }
    return false
}
```

### 3.2 Causal Solver（因果求解器）工作原理

**核心算法**：A* 搜索在平行宇宙中寻找最优路径

**求解流程** (`core/solver/solver.go`)：

```go
func Solve(dag *HistoryDAG, tipA, tipB NodeID) ResolutionPlan {
    // 1. 初始化搜索空间
    pq := &search.PriorityQueue{}  // 最小堆
    heap.Push(pq, &search.State{
        Mutations: []change.Mutation{},  // 初始无突变
        Cost:      0,
    })

    // 2. 迭代扩展状态
    for pq.Len() > 0 {
        current := heap.Pop(pq)

        // 3. 环境重构：在沙盒中应用突变
        sandbox := rewrite.RewriteBatch(dag, lca, current.Mutations)

        // 4. 冲突分析
        res := analysis.AnalyzeMerge(sandbox, tipA, tipB)
        if len(res.Conflicts) == 0 {
            return ResolutionPlan{Resolved: true}  // 找到解！
        }

        // 5. 定向扩展（只处理冲突节点）
        conflict := res.Conflicts[0]
        candidates := generateTargetedCandidates(sandbox, offenderID)
        for _, mut := range candidates {
            heap.Push(pq, nextState)
        }
    }
}
```

**关键优化**：
- **定向候选生成**：只针对冲突节点生成变体，避免指数爆炸
- **状态指纹**：通过 SHA-256 哈希防止环路
- **LCA 计算**：从最近公共祖先开始回放，减少搜索空间
- **超时保护**：5 秒后返回当前最优解

### 3.3 平行宇宙计算（Ephemeral Sandbox）

**设计思想**：在内存中低成本 Fork/Rewrite 历史，不污染真实数据

**实现** (`core/rewrite/ephemeral.go`)：

```go
type EphemeralDAG struct {
    Base    DagView                    // 基础历史（只读）
    Overlay map[NodeID]*Node           // 覆盖层（突变）
    Head    NodeID                     // 当前时间线头部
}

func RewriteBatch(base DagView, startPoint NodeID, mutations []Mutation) *EphemeralDAG {
    sandbox := NewEphemeralDAG(base, startPoint)
    for _, m := range mutations {
        if m.Type == ReplaceOp {
            orig := sandbox.GetNode(m.Target)
            newNode := *orig
            newNode.Op = m.NewOp  // 替换操作
            sandbox.Overlay[m.Target] = &newNode
        }
    }
    return sandbox  // 返回平行宇宙视图
}
```

**特性**：
- Copy-on-Write 语义
- 零性能开销（修改只在 Overlay 中）
- 支持批量突变操作
- 自动回滚（丢弃沙盒即可）

### 3.4 语义代价评估（Semantic Cost）

**代价单位**：SLU (Semantic Logical Unit)

```go
const (
    Zero        Cost = 0      // 无代价
    Tweak       Cost = 20     // 微调
    Downgrade   Cost = 50     // 降级（Delete → Move）
    Neutralize  Cost = 100    // 中和（Op → NoOp）
    Destructive Cost = 500    // 破坏性操作
    Infinite    Cost = 10000  // 不可接受
)
```

**代价模型** (`core/cost/registry.go`)：

```go
type Model interface {
    Calculate(m Mutation, ctx Context) Cost
}

type DefaultModel struct{}
func (d DefaultModel) Calculate(m Mutation, ctx Context) Cost {
    if desc := m.NewOp.Describe(); desc == "NoOp(Neutralized)" {
        return Neutralize  // 100 SLU
    }
    return Downgrade  // 50 SLU (默认)
}
```

**可扩展性**：支持注册自定义模型
```go
func RegisterModel(name string, model Model)
func GetModel(name string) Model
```

### 3.5 Responsibility Narrative（责任叙事）

**设计理念**：决策报告不是 UI 装饰，而是搜索路径的真实转录

**数据结构** (`core/narrative/model.go`)：

```go
type Narrative struct {
    Summary   string         // "Conflict resolved via optimized causal path"
    TotalCost int            // 总代价（如 50 SLU）
    Steps     []DecisionStep // 完整决策轨迹
}

type DecisionStep struct {
    ProblemContext string                // "Resource Contention: README.md"
    Decision       string                // "Mutate nodeB -> Move(Trash/README.md)"
    DecisionCost   int                   // 50
    Rejected       []RejectedAlternative // 被拒绝的备选方案
}

type RejectedAlternative struct {
    Description string // "Mutate nodeB -> NoOp(Neutralized)"
    Cost        int    // 100
    Reason      string // "Higher semantic cost"
}
```

**渲染输出**：
- Markdown 格式（CLI 默认）
- HTML 格式（Web 界面）
- JSON 格式（API 集成）

---

## 4. 技术架构深度分析

### 4.1 核心数据结构

**历史 DAG** (`core/history/dag.go`)：
```go
type Node struct {
    ID      NodeID
    Op      ReversibleChange
    Parents []NodeID  // 支持多父节点（merge commits）
}

type HistoryDAG struct {
    Nodes map[NodeID]*Node
    Roots []NodeID
}
```

**搜索状态** (`core/search/search.go`)：
```go
type State struct {
    Mutations   []Mutation        // 已选定的手术路径
    Cost        Cost              // g(n)：实际代价
    Heuristic   Cost              // h(n)：启发式预估
    Narrative   []DecisionStep   // 决策记录
    Fingerprint uint64           // 防重复（SHA-256）
}
```

### 4.2 关键算法实现

**LCA 算法** (`core/history/lca.go`)：
- BFS 双向搜索
- 支持多父节点 DAG
- 时间复杂度：O(V + E)

**冲突优先级调度** (`core/scheduler/priority.go`)：
```go
func ConflictSeverity(c Conflict) int {
    if c.ModeA == Exclusive || c.ModeB == Exclusive { return 100 }  // 高
    if c.ModeA == Create && c.ModeB == Create { return 80 }        // 中
    return 50  // 低
}
```

**状态指纹计算**：
```go
func ComputeFingerprint(mutations []Mutation) uint64 {
    var h maphash.Hash
    for _, m := range mutations {
        // Zero-allocation 优化
        targetBytes := *(*[]byte)(unsafe.Pointer(&m.Target))
        h.Write(targetBytes)
        h.WriteString(m.NewOp.Hash())
    }
    return h.Sum64()
}
```

### 4.3 API 接口设计

**HTTP API** (`api/http/`)：

```
GET /solve?format=markdown  # Markdown 报告
GET /solve?format=html      # HTML 可视化报告
GET /solve                  # JSON 结构化数据
GET /health                 # 健康检查
```

**处理流程**：
```go
func solveHandler(w http.ResponseWriter, r *http.Request) {
    dag, tipA, tipB := loader.LoadDemoScenario()
    plan := solver.Solve(dag, tipA, tipB)

    format := r.URL.Query().Get("format")
    switch format {
    case "markdown":
        w.Write([]byte(formatter.ToMarkdown(plan.Narrative)))
    case "html":
        w.Write([]byte(formatter.ToHTML(plan.Narrative)))
    default:
        json.NewEncoder(w).Encode(plan)
    }
}
```

### 4.4 与 Tmux-FSM 的集成方式

**观察到的集成点**：
1. **子模块位置**：位于 `/Users/ygs/ygs/Tmux-FSM/rhm-go`
2. **不是独立 Git 仓库**：在父项目内部
3. **潜在集成方式**：
   - 作为库导入：`import "rhm-go/core/solver"`
   - HTTP 服务间通信：通过 REST API
   - 状态管理：使用 RHM 引擎处理 tmux 会话状态冲突

**可能的集成场景**：
- 多人协作编辑 tmux 会话配置
- 并发窗口操作的冲突解决
- 会话状态的版本控制和回溯

---

## 5. 特色功能与创新点

### 5.1 智能合并

**传统方式 vs RHM**：
- **Git 三方合并**：基于行比较，易产生虚假冲突
- **RHM 因果推理**：基于意图分析，智能降级冲突

**示例：Edit vs Delete 冲突**
```
Branch A: Edit(file.txt)      # 需要文件存在
Branch B: Delete(file.txt)    # 销毁文件

传统 Git：CONFLICT (content/markdown/folder)
RHM 裁决：Delete → Move(Trash/file.txt)
        代价：50 SLU（降级）
        结果：文件保留到垃圾箱，Edit 可以继续
```

### 5.2 临时沙箱机制

**优势**：
- **零成本**：不需要克隆整个仓库
- **即时性**：内存操作，毫秒级响应
- **隔离性**：不影响原始历史
- **批处理**：一次重写多个节点

**应用场景**：
- 尝试性合并
- 冲突预览
- 历史分支模拟

### 5.3 A* 搜索算法的应用

**启发式函数**：
```go
nextState.Heuristic = cost.Cost(len(res.Conflicts)-1) * cost.Tweak
```

**搜索剪枝策略**：
- **定向扩展**：只对冲突节点生成候选
- **代价排序**：优先探索低成本路径
- **闭环检测**：指纹去重防止重复探索

### 5.4 冲突处理策略

**两级降级机制**：

1. **Downgrade（降级）**：50 SLU
   - Delete → Move(Trash/)
   - 保留大部分意图
   - 首选策略

2. **Neutralize（中和）**：100 SLU
   - Op → NoOp
   - 牺牲意图换取一致性
   - 备选策略

**智能选择**：Solver 自动选择代价最小的路径

---

## 6. 稳定性保证（STABILITY.md）

### 6.1 公共核心 API（Stable）

保证 SemVer 兼容性：
- `core/history`：历史图结构
- `core/change`：可逆变更接口
- `core/solver`：Solve 函数契约
- `core/cost`：代价模型接口

### 6.2 半稳定扩展点

允许扩展但可能变更：
- `store/ops.go`：具体业务实现
- `internal/formatter`：报告渲染
- `core/narrative`：叙事数据结构

### 6.3 实验性内部模块

不保证稳定性：
- `core/analysis`：冲突检测启发式
- `core/search`：A* 内部状态管理
- `internal/loader`：测试加载器

---

## 7. 质量保证

### 7.1 稳定性测试套件

**三大核心保证**：

1. **顺序不变性**：
```go
func TestStability_OrderInvariance(t *testing.T)
// DAG 构造顺序不影响裁决结果
```

2. **代价支配性**：
```go
func TestStability_CostDominance(t *testing.T)
// 必须选择 50 SLU 的降级路径，而非 100 SLU 的中和路径
```

3. **确定性**：
```go
func TestStability_Determinism(t *testing.T)
// 同一 DAG 运行 100 次必须比特级一致
```

### 7.2 功能测试

- `TestSolveWithFootprintAnalysis`：足迹冲突分析
- `TestSolveWithNoConflict`：无冲突场景
- `TestSolveWithCreateVsCreateConflict`：创建冲突

### 7.3 Prometheus 监控

**指标**：
- `rhm_solve_duration_seconds`：求解耗时
- `rhm_conflict_count`：冲突数量（按严重程度）
- `rhm_memory_usage_bytes`：内存消耗

**自动仪表化**：
```go
func InstrumentSolver(originalSolver) func(...) ResolutionPlan
```

---

## 8. 部署与运行

### 8.1 CLI 模式

```bash
go run cmd/rhm/main.go solve
```

输出示例：
```markdown
# Conflict resolved via optimized causal path

**Total Semantic Cost:** `50 SLU`

## Decision Trail

### Step 1: Resource Contention: README.md
> **Selected:** `Mutate nodeB -> Move(Trash/README.md)` (Cost 50)
```

### 8.2 服务模式

```bash
docker build -t rhm-engine .
docker run -p 8080:8080 rhm-engine
curl "http://localhost:8080/solve?format=markdown"
```

### 8.3 Docker 多阶段构建

- **Builder 阶段**：golang:1.21-alpine 编译
- **Run 阶段**：alpine:latest 运行
- **最终镜像**：< 20 MB

---

## 9. 设计思想总结

### 9.1 核心创新

1. **语义优先**：不比较文本，比较意图
2. **可逆性**：所有操作都可以撤销或降级
3. **审计性**：每个决策都有完整记录
4. **因果性**：基于依赖关系而非文本差异

### 9.2 工程哲学

从 `do.md` 可以看出设计进化路径：

**第一代**：骨架
- 基础数据结构定义

**第二代**：肌肉
- 足迹代数引入
- 定向候选生成
- 冲突优先级调度

**第三代**：生产级
- 确定性保证
- 性能优化
- 完整测试覆盖

### 9.3 技术亮点

- **零分配指纹**：`unsafe.Pointer` 优化
- **堆优化**：A* 搜索使用标准库 `container/heap`
- **并发安全**：状态指纹防止竞态
- **可扩展性**：插件化代价模型

---

## 10. 关键文件路径索引

**核心引擎**（773 行）：
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/core/change/change.go` - 代数底座
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/core/solver/solver.go` - 主求解器
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/core/analysis/analysis.go` - 冲突分析
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/core/search/search.go` - A* 搜索
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/core/rewrite/ephemeral.go` - 平行宇宙

**业务层**：
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/store/ops.go` - 文件系统操作实现

**API 层**：
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/api/http/handlers.go` - HTTP 处理器
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/api/http/server.go` - 服务器启动

**工具层**：
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/internal/formatter/markdown.go` - Markdown 渲染
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/internal/formatter/html.go` - HTML 渲染
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/telemetry/metrics.go` - Prometheus 指标

**文档**：
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/README.md` - 快速开始指南
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/STABILITY.md` - 稳定性承诺
- `/Users/ygs/ygs/Tmux-FSM/rhm-go/do.md` - 设计演进文档

---

## 结论

RHM-Go 是一个**高度创新的版本控制引擎**，它通过以下核心特性重新定义了冲突解决：

1. **因果推理**：基于操作意图而非文本内容
2. **平行宇宙**：在内存沙箱中低成本探索所有可能性
3. **语义代价**：量化每个决策的成本，自动选择最优路径
4. **责任叙事**：完整的决策审计，包含被拒绝的备选方案
5. **可逆性**：所有操作都可以撤销或降级

这是一个**生产级内核**，具备：
- 完整的测试覆盖（稳定性 + 功能性）
- 严格的 API 契约
- Prometheus 监控集成
- Docker 容器化部署
- CLI 和 HTTP 双模式支持

在 Tmux-FSM 项目中，RHM-Go 可以作为**智能冲突解决引擎**，处理多人协作时的会话状态冲突，提供比传统版本控制系统更智能、更温和的合并策略。

---

*生成时间: 2026-01-14*
*分析工具: Claude Code*
