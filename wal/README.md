# wal 模块

## 模块职责概述

`wal/` 是 **Tmux-FSM 的 Write-Ahead Log（预写日志）系统**，负责持久化记录所有的状态变更操作，确保数据的持久性和可恢复性。该模块实现了高性能的日志记录和恢复机制，是系统数据安全和故障恢复的重要保障。

主要职责包括：
- 记录所有状态变更操作到持久化日志
- 提供高效的日志写入和读取能力
- 支持日志的压缩和清理
- 实现基于日志的系统恢复机制

## 核心设计思想

- **持久性保障**: 确保所有操作都能持久化记录
- **高性能写入**: 优化日志写入性能
- **顺序访问**: 利用顺序I/O提高性能
- **自动恢复**: 支持基于日志的自动恢复

## 文件结构说明

### `wal.go`
- 核心 WAL 实现
- 主要结构体：
  - `WAL`: 预写日志管理器
  - `LogEntry`: 日志条目
  - `LogPosition`: 日志位置
  - `WALConfig`: WAL 配置
- 主要函数：
  - `NewWAL(path string, config WALConfig) (*WAL, error)`: 创建 WAL 实例
  - `Write(entry LogEntry) error`: 写入日志条目
  - `Read(position LogPosition) (LogEntry, error)`: 读取日志条目
  - `Sync() error`: 同步日志到磁盘
  - `Close() error`: 关闭 WAL
- 负责核心的日志管理功能

### `entry.go`
- 日志条目定义
- 主要结构体：
  - `SemanticEvent`: 语义事件
  - `EventHeader`: 事件头部
  - `Checksum`: 校验和
- 主要函数：
  - `MarshalBinary() ([]byte, error)`: 序列化事件
  - `UnmarshalBinary(data []byte) error`: 反序列化事件
  - `Validate() error`: 验证事件完整性
  - `CalculateChecksum() Checksum`: 计算校验和
- 定义日志条目的结构和操作

### `recovery.go`
- 恢复管理器
- 主要函数：
  - `RecoverFromLogs(logDir string) (State, error)`: 从日志恢复状态
  - `ValidateLogIntegrity(logFile string) bool`: 验证日志完整性
  - `FindRecoveryPoint() LogPosition`: 查找恢复点
  - `ApplyLogEntries(entries []LogEntry, state State) State`: 应用日志条目
- 实现系统恢复功能

### `compactor.go`
- 日志压缩器
- 主要函数：
  - `CompactLogs(threshold int64) error`: 压缩日志
  - `CleanupOldSegments(retention time.Duration) error`: 清理旧日志段
  - `GetLogStats() LogStats`: 获取日志统计信息
  - `OptimizeStorage() error`: 优化存储空间
- 管理日志的存储优化

### `reader.go`
- 日志读取器
- 主要函数：
  - `NewReader(wal *WAL) *Reader`: 创建读取器
  - `IterateFrom(position LogPosition) <-chan LogEntry`: 从指定位置迭代
  - `Scan(filter FilterFunc) []LogEntry`: 扫描过滤日志
  - `GetPosition() LogPosition`: 获取当前位置
- 提供高效的日志读取能力

## WAL 特性

### 高性能
- 批量写入优化
- 异步写入支持
- 内存映射文件技术

### 数据安全
- 事务性写入保证
- 校验和验证
- 崩溃恢复能力

### 存储管理
- 自动日志轮转
- 智能压缩策略
- 存储空间回收

## 在整体架构中的角色

WAL 模块是系统的持久化保障层，它确保所有重要的状态变更都被安全地记录下来。通过预写日志机制，WAL 提供了：
- 数据的持久性保障
- 系统崩溃后的恢复能力
- 高性能的日志记录
- 可靠的故障恢复机制