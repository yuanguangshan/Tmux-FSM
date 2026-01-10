# “删 resolver legacy 清洗器那一天”的最后检查表

> 这是 **Phase 3 的 D-Day Checklist**  
> **全部 ✅ 才允许删代码**

---

## ✅ A. 编译期保证（必须 100%）

- [ ] `AnchorOriginLegacy` enum 未被任何非 test 文件引用
- [ ] `resolveLegacyAnchor()` **只在 test 中存在**
- [ ] `processKeyLegacy()` 无任何调用点
- [ ] `actionStringToIntentWithLineInfo()` 被删除

✅ 推荐工具：

```bash
rg "Legacy" src/ | grep -v _test.go
```

---

## ✅ B. 运行期保证（严格模式全开）

```go
StrictNativeResolver = true
StrictNativeFSM = true
```

- [ ] 全量单测通过
- [ ] Undo / Redo fuzz 测试 1万次无 panic
- [ ] 随机 key replay 无 legacy 泄漏

---

## ✅ C. 行为一致性保证（对照测试）

### 必须通过的三类黄金测试：

#### 1️⃣ Native vs Legacy 行为一致（快照 diff）

```text
Initial Snapshot
→ Apply native intent sequence
→ Apply legacy intent sequence
→ Snapshot hash must match
```

---

#### 2️⃣ 长链 Undo / Redo

- [ ] ≥ 500 步操作
- [ ] 任意中断点恢复一致

---

#### 3️⃣ 边界语义

- EOF motion
- 空行 delete
- Unicode combining chars
- Wrapped lines

---

## ✅ D. 仓库级信号（这是“仪式感”的一刻）

- [ ] 删除 `legacy/` 目录
- [ ] 删除 `legacy_*.go`
- [ ] 删除 `AnchorOriginLegacy`
- [ ] 删除 checklist 中所有 legacy TODO

---

## ✅ E. 提交规范（强烈建议）

```text
commit: Remove legacy intent resolver

- Phase 3 completed
- StrictNativeResolver always enabled
- All legacy intent paths removed
```
