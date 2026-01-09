# verifier 模块

## 模块职责概述

`verifier/` 是 **Tmux-FSM 的正确性守门人**，负责验证系统状态的一致性、操作的正确性和历史记录的完整性。该模块关注的问题是："系统的决策与执行，是否符合我们定义的规则与不变量？"，是系统信任模型的重要组成部分。Verifier 不生成行为，而是对 Intent → Kernel → Operation → Execution 这一完整链路进行事后或准实时校验。

主要职责包括：
- 验证操作的语义正确性
- 检查系统状态的一致性
- 验证历史记录的完整性
- 执行策略合规性检查
- 提供验证报告和审计功能

## 核心设计思想

- **全面验证**: 从多个维度验证系统状态和操作
- **策略驱动**: 基于策略进行合规性检查
- **可证明性**: 提供验证证据和证明
- **实时监控**: 支持实时验证和告警

## 文件结构说明

### `verifier.go`
- 核心验证器实现
- 主要结构体：
  - `Verifier`: 验证器主结构
  - `VerificationResult`: 验证结果
  - `ValidationError`: 验证错误
- 主要函数：
  - `NewVerifier(config Config) *Verifier`: 创建验证器
  - `VerifyOperation(op Operation) *VerificationResult`: 验证操作
  - `VerifyState(state State) *VerificationResult`: 验证状态
  - `VerifyHistory(events []SemanticEvent) *VerificationResult`: 验证历史
  - `RunConsistencyCheck() []ValidationError`: 运行一致性检查
- 负责核心的验证逻辑

### `consistency_checker.go`
- 一致性检查器
- 主要函数：
  - `CheckStateConsistency(current, expected State) bool`: 检查状态一致性
  - `CheckEventOrdering(events []SemanticEvent) bool`: 检查事件顺序
  - `ValidateCausalRelationships(events []SemanticEvent) bool`: 验证因果关系
  - `CheckInvariantViolations(events []SemanticEvent) []InvariantViolation`: 检查不变量违反
- 确保系统状态的一致性

### `policy_verifier.go`
- 策略验证器
- 主要函数：
  - `VerifyPolicyCompliance(intent Intent, actor ActorID) bool`: 验证策略合规性
  - `CheckTrustLevel(actor ActorID, operation Operation) bool`: 检查信任级别
  - `ValidatePermissions(intent Intent, actor ActorID) bool`: 验证权限
  - `AuditOperation(operation Operation, actor ActorID) AuditRecord`: 审计操作
- 基于策略的安全验证

### `proof_generator.go`
- 证明生成器
- 主要函数：
  - `GenerateProof(operation Operation) Proof`: 生成操作证明
  - `VerifyProof(proof Proof) bool`: 验证证明有效性
  - `CreateEvidence(operation Operation) Evidence`: 创建证据
  - `ValidateEvidence(evidence Evidence) bool`: 验证证据
- 提供可验证的证明机制

### `audit_logger.go`
- 审计日志记录器
- 主要函数：
  - `LogVerification(result VerificationResult)`: 记录验证结果
  - `GenerateAuditReport(from, to time.Time) AuditReport`: 生成审计报告
  - `CheckForAnomalies() []Anomaly`: 检测异常
  - `ExportAuditTrail() []AuditRecord`: 导出审计轨迹
- 提供审计和监控功能

## 验证特性

### 多层验证
- 语法验证：检查操作格式的正确性
- 语义验证：检查操作语义的合理性
- 策略验证：检查操作的策略合规性
- 一致性验证：检查系统状态的一致性

### 实时验证
- 支持操作前的预验证
- 支持操作后的后验证
- 支持周期性的状态验证

### 证明机制
- 为每个验证结果提供证明
- 支持验证结果的独立验证
- 提供可追溯的验证链

## 在整体架构中的角色

Verifier 模块是系统的质量保障层，它确保所有操作都符合预期的行为和策略要求。通过多层次的验证机制，Verifier 为系统提供了：
- 操作正确性的保证
- 系统状态一致性的维护
- 策略合规性的强制执行
- 可审计的操作轨迹