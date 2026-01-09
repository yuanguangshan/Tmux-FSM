# policy 模块

## 模块职责概述

`policy/` 是 **Tmux-FSM 的策略与权限管理系统**，负责定义和执行系统的安全策略、访问控制和信任级别管理。该模块实现了灵活的策略框架，支持细粒度的权限控制和动态策略调整。

主要职责包括：
- 管理参与者信任级别
- 执行访问控制策略
- 定义操作权限规则
- 提供策略的动态配置和更新机制

## 核心设计思想

- **分层策略**: 支持系统级、用户级、操作级的分层策略
- **动态配置**: 支持运行时策略的动态更新
- **信任管理**: 基于信任级别的差异化策略执行
- **可扩展性**: 支持自定义策略规则的扩展

## 文件结构说明

### `policy.go`
- 核心策略定义
- 主要结构体：
  - `PolicyManager`: 策略管理器
  - `TrustLevel`: 信任级别枚举
  - `Permission`: 权限定义
  - `PolicyRule`: 策略规则
- 主要函数：
  - `NewPolicyManager() *PolicyManager`: 创建策略管理器
  - `AllowCommit(actor ActorID, event SemanticEvent) error`: 允许提交检查
  - `CheckPermission(actor ActorID, operation Operation) bool`: 检查权限
  - `UpdatePolicy(rule PolicyRule) error`: 更新策略
- 负责核心的策略管理逻辑

### `trust_manager.go`
- 信任级别管理器
- 主要结构体：
  - `ActorInfo`: 参与者信息
  - `TrustProfile`: 信任档案
- 主要函数：
  - `RegisterActor(info ActorInfo)`: 注册参与者
  - `GetTrustLevel(actor ActorID) TrustLevel`: 获取信任级别
  - `UpdateTrustScore(actor ActorID, score float64)`: 更新信任分数
  - `AdjustTrustLevel(actor ActorID, adjustment TrustAdjustment)`: 调整信任级别
- 管理参与者的信任级别

### `permission_checker.go`
- 权限检查器
- 主要函数：
  - `CheckAccess(actor ActorID, resource Resource, action Action) bool`: 检查访问权限
  - `ValidateOperation(actor ActorID, operation Operation) error`: 验证操作权限
  - `GetEffectivePermissions(actor ActorID) []Permission`: 获取有效权限
  - `CanPerform(actor ActorID, intent Intent) bool`: 检查能否执行意图
- 执行具体的权限检查

### `rule_engine.go`
- 策略规则引擎
- 主要函数：
  - `EvaluateRules(context PolicyContext, rules []PolicyRule) Decision`: 评估策略规则
  - `AddRule(rule PolicyRule)`: 添加策略规则
  - `RemoveRule(ruleID string)`: 移除策略规则
  - `MatchConditions(context PolicyContext, conditions []Condition) bool`: 匹配条件
- 执行策略规则的评估

### `config_loader.go`
- 策略配置加载器
- 主要函数：
  - `LoadPolicyConfig(path string) (*PolicyConfig, error)`: 加载策略配置
  - `ValidateConfig(config *PolicyConfig) error`: 验证配置有效性
  - `ApplyConfig(config *PolicyConfig) error`: 应用配置
  - `ReloadConfig() error`: 重新加载配置
- 管理策略配置的加载和应用

## 策略特性

### 信任级别
- `TrustLevel_Trusted`: 受信任的参与者
- `TrustLevel_Untrusted`: 不受信任的参与者  
- `TrustLevel_System`: 系统级权限
- 基于信任级别实施差异化策略

### 权限控制
- 细粒度的操作权限控制
- 基于角色的访问控制
- 时间和上下文相关的权限检查

### 动态策略
- 支持运行时策略更新
- 策略热加载和生效
- 策略版本管理和回滚

## 在整体架构中的角色

Policy 模块是系统的安全控制层，它确保所有操作都符合预定义的安全策略。通过灵活的策略框架，Policy 模块提供了：
- 基于信任级别的差异化访问控制
- 动态的权限管理和策略调整
- 安全策略的集中管理和执行
- 策略合规性的强制执行