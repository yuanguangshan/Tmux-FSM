# RHM: Reversible History Model (Go Engine v1.0)

RHM 是一个**因果感知**的版本控制与合并引擎。它不比较文本行，而是推理历史意图。

## 核心特性

*   **Causal Solver**: 使用 A* 算法在平行宇宙中寻找语义代价最小的合并方案。
*   **Responsibility Narrative**: 自动生成包含“已拒绝备选方案”的决策审计报告。
*   **Ephemeral Sandbox**: 在内存中低成本 Fork/Rewrite 历史，不污染真实数据。

## 快速开始

### CLI 模式
```bash
go run cmd/rhm/main.go solve
```

### 服务模式
```bash
docker build -t rhm-engine .
docker run -p 8080:8080 rhm-engine
# 测试
curl "http://localhost:8080/solve?format=markdown"
```

## 演示场景
预置场景为 **Edit vs Delete** 冲突：
1. Branch A: `Edit(File)` (需要文件存在)
2. Branch B: `Delete(File)` (销毁文件)

**RHM 裁决结果**:
将 `Delete` 降级为 `Move(Trash/File)`，保留数据，消除冲突，代价 50 SLU。
