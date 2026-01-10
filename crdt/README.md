

# CRDT 模块文档 ⇄ 代码映射表

---

## 1. 模块职责概述

> “将多个事件历史合并为一个因果一致、确定性的事件序列”

### ✅ 对应代码

- `EventStore.Merge(e SemanticEvent)`
- `TopoSort() []SemanticEvent`
- `TopoSortByCausality(events map[EventID]SemanticEvent)`

```go
func (s *EventStore) Merge(e SemanticEvent)
func (s *EventStore) TopoSort() []SemanticEvent
func TopoSortByCausality(events map[EventID]SemanticEvent) []SemanticEvent
```

### 🧠 设计体现

- 不依赖事件到达顺序
- 只依赖事件集合 + 因果边
- 输出顺序完全由数据决定

---

## 2. 无冲突合并（Conflict‑Free）

> “合并操作满足幂等性”

### ✅ 对应代码

```go
func (s *EventStore) Merge(e SemanticEvent) {
    if existing, ok := s.Events[e.ID]; ok {
        if e.Version > existing.Version {
            s.Events[e.ID] = e
        }
        return
    }
    s.Events[e.ID] = e
}
```

### 🧠 设计体现

- `EventID` 作为全局主键
- 重复事件不会改变集合
- Merge 顺序不影响最终结果

---

## 3. 确定性收敛（Deterministic Convergence）

> “相同事件集合 → 相同事件顺序”

### ✅ 对应代码

```go
sort.Slice(queue, func(i, j int) bool {
    return queue[i] < queue[j]
})
```

（位于 `TopoSortByCausality`）

### 🧠 设计体现

- 对入度为 0 的并发事件进行稳定排序
- 排序键是 `EventID`（字符串全序）
- 保证跨副本顺序一致

⚠️ **这是一个系统性保证，不是单一函数**

---

## 4. 因果有序（Causal Ordering）

> “严格遵循 CausalParents”

### ✅ 对应代码

```go
for _, e := range events {
    for _, p := range e.CausalParents {
        if _, ok := events[p]; ok {
            graph[p] = append(graph[p], e.ID)
            inDegree[e.ID]++
        }
    }
}
```

```go
if len(result) != len(events) {
    panic("causal cycle detected")
}
```

### 🧠 设计体现

- 明确建模因果 DAG
- 禁止因果环
- 所有排序必须满足因果约束

---

## 5. 本地历史与全局历史分离

> “LocalParent 不参与 CRDT 合并”

### ✅ 对应代码

- **未出现在任何合并 / 排序逻辑中**

```go
// SemanticEvent
CausalParents []EventID
LocalParent   EventID
```

```go
// TopoSortByCausality 中完全未使用 LocalParent
```

### 🧠 设计体现

- LocalParent 是“结构性忽略”
- 这是设计约束，而不是实现疏漏

---

## 6. crdt.go：核心类型定义

### ✅ PositionID 顺序定义

```go
func ComparePos(a, b PositionID) int
```

比较顺序：

1. Path
2. Actor
3. Epoch

### ✅ 位置分配

```go
func AllocateBetween(a, b *PositionID, actor ActorID) PositionID
```

### 🧠 设计体现

- 无需全局协调
- 支持并发插入
- 始终可分配新位置

---

## 7. event_store.go：事件集合与排序

> “不是 WAL，不是 append‑only log”

### ✅ 对应代码

```go
type EventStore struct {
    Events map[EventID]SemanticEvent
}
```

- 使用 map 而非 slice
- 不记录插入顺序
- 不暴露 offset / index

### ✅ TopoSort

```go
func (s *EventStore) TopoSort() []SemanticEvent
```

只是 `TopoSortByCausality` 的薄封装。

---

## 8. position.go：逻辑位置管理

### ✅ 对应代码

```go
type PositionID struct {
    Path  []uint32
    Actor ActorID
    Epoch int
}
```

```go
func AllocateBetween(a, b *PositionID, actor ActorID) PositionID
```

### 🧠 设计体现

- 路径型位置（Prefix Ordering）
- Actor + Epoch 作为最终裁决维度

---

## 9. vector_clock.go：向量时钟（存在但未强制）

> “结构存在，但未参与合并裁决”

### ✅ 对应代码

```go
type VectorClock struct { ... }
func (vc *VectorClock) Compare(...)
func (vc *VectorClock) Merge(...)
```

（未在 `Merge` / `TopoSort` 中调用）

### 🧠 设计体现

- 为未来的 delta sync / frontier 对齐预留
- 当前系统不依赖它保证正确性

---

## 10. Undo / Redo 行为

> “Undo 是执行时过滤，而非历史修改”

### ✅ 对应代码

```go
func UndoFilter(
    me ActorID,
    undoPoint EventID,
    events map[EventID]SemanticEvent,
) func(SemanticEvent) bool
```

```go
func LocalHistory(events map[EventID]SemanticEvent, me ActorID) []SemanticEvent
func BuildLocalChain(events []SemanticEvent) []SemanticEvent
```

### 🧠 设计体现

- Undo 只影响本 Actor
- 不删除事件
- 不破坏 CRDT 收敛性

---

## 11. 在整体架构中的角色（结构性结论）

> “为可回放、可验证系统提供基础”

### ✅ 对应代码事实

- 所有事件不可变
- 因果关系显式
- 排序可重放
- Merge 无副作用

⚠️ **这是跨文件的整体性质，而非某一函数**

---

# 总结性一句话（非常关键）

> **这份文档不是“抽象描述代码”，  
> 而是“把代码里隐含的系统约束写成了人类可读的形式”。**

文档与实现现在是：

✅ 同一抽象层级  
✅ 同一能力边界  
✅ 同一责任划分  
