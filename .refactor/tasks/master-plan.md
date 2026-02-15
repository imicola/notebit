# Master Plan: Notebit Refactoring Round 2
**Created**: 2026-02-15
**Status**: Ready for Execution
**Progress**: 0%

---

## 🎯 Project Objectives
1. 修复所有 P0 Critical 安全/稳定性问题
2. 修复 P1 High 级别的功能 bug 和崩溃风险
3. 清理死代码、消除重复逻辑
4. 改善前端性能和内存管理
5. 确保所有修改通过编译测试

---

## ⚠️ 编译说明
项目使用 CGO (sqlite-vec)，编译速度较慢。所有任务完成后**统一编译测试**。

---

## Phase 1: P0 Critical 修复 [Architecture + Module Layer]

### A-201: 统一模型维度映射到单一来源
- **严重度**: P2 (架构层)
- **文件**: `pkg/ai/openai.go`, `pkg/ai/ollama.go`, `pkg/config/config.go`
- **操作**:
  1. 在 `pkg/ai/` 中创建 `dimensions.go`，统一维度映射
  2. `openai.go` 和 `ollama.go` 引用统一映射
  3. `config.go` 删除重复映射，引用 `pkg/ai/`
- **状态**: ⏳ 待执行

### A-202: 提取 initializeServices 消除初始化重复
- **严重度**: P2 (架构层)
- **文件**: `app.go`, `app_files.go`
- **操作**:
  1. 将 `OpenFolder` / `SetFolder` / `startup` 中的重复逻辑提取为 `initializeServices(basePath)`
- **状态**: ⏳ 待执行

---

### M-201: 修复 AI GetStatus 死锁 + ChunkText 线程安全
- **严重度**: P0
- **文件**: `pkg/ai/service.go`
- **操作**:
  1. `GetStatus()`: 提取内部无锁方法 `getAvailableProvidersLocked()` / `getAvailableStrategiesLocked()`
  2. `ChunkText()`: 添加 `s.mu.RLock()` / `defer s.mu.RUnlock()`
- **状态**: ⏳ 待执行

### M-202: 修复 MigrateToVec 死循环
- **严重度**: P0
- **文件**: `pkg/database/migration_vec.go`
- **操作**:
  1. 对空 `EmbeddingBlob` 的 chunk 也标记 `vec_indexed = true`
  2. 添加安全计数器防止无限循环
- **状态**: ⏳ 待执行

### M-203: 修复 IndexAll 数据竞争 + Stop 死锁防护
- **严重度**: P0
- **文件**: `pkg/indexing/pipeline.go`
- **操作**:
  1. `IndexAll()`: 使用 `atomic.Int64` 替代裸并发写
  2. `Stop()`: 向等待的 `errChan` 发送停止信号
  3. 添加 `started` 守卫标志
- **状态**: ⏳ 待执行

### M-204: 添加路径遍历防护
- **严重度**: P0 安全
- **文件**: `pkg/files/manager.go`
- **操作**:
  1. 添加 `validatePath()` 方法: `filepath.Abs()` + `strings.HasPrefix(abs, basePath)`
  2. 在所有文件操作方法中调用
  3. 添加 `data/` 到文件树跳过列表
- **状态**: ⏳ 待执行

---

## Phase 2: P1 High 修复

### M-205: 修复 Config boolean merge + VectorDimension
- **严重度**: P1
- **文件**: `pkg/config/config.go`
- **操作**:
  1. 使用 `json.RawMessage` 或手动检查已设置字段，防止零值覆盖布尔默认值
  2. 在 `mergeWithDefaults` 中添加 `VectorDimension` 合并
- **状态**: ⏳ 待执行

### M-206: 修复 RAG context 遮蔽 + nil panic
- **严重度**: P1
- **文件**: `pkg/rag/service.go`
- **操作**:
  1. `context := s.buildContext(...)` → `ragContext := ...`
  2. `buildContext` 中添加 `chunk.File` nil 检查
  3. 删除 `// "strconv"` 注释
  4. `generateMessageID()` 改用 UUID 或 nanoid
- **状态**: ⏳ 待执行

### M-207: 修复 Chat BackupTicker panic
- **严重度**: P1
- **文件**: `pkg/chat/service.go`
- **操作**:
  1. `close(s.stopCh)` 前用 `sync.Once` 或 select 检查
  2. 修复硬编码中文字符串为常量
- **状态**: ⏳ 待执行

### M-208: 修复 Database 静默错误吞掉
- **严重度**: P1
- **文件**: `pkg/database/repository.go`
- **操作**:
  1. `DeleteFile` / `DeleteChunksForFile` 中 vec_chunks 删除错误改为 logger.Warn
  2. 不再使用 `_ =` 忽略错误
- **状态**: ⏳ 待执行

---

## Phase 3: 死代码清理 + 重复消除

### M-209: Go 死代码清理
- **严重度**: P2
- **文件**: `pkg/ai/errors.go`, `pkg/ai/service.go`, `pkg/watcher/stat.go`, `pkg/database/models.go`
- **操作**:
  1. 删除 `pkg/ai/errors.go` (未使用的自定义错误类型)
  2. 删除 `service.go` 中 Metadata 冗余写入
  3. 删除 `pkg/watcher/stat.go` (死代码)
  4. 清理 `models.go` 过期 TODO 和注释代码
  5. 删除 `pkg/rag/service.go` 中注释的 import
- **状态**: ⏳ 待执行

### M-211: 前端死代码清理
- **严重度**: P2
- **文件**: 
  - `frontend/src/services/graphService.js` — 删除未使用的 `detectNodeType()`, `getNodeColorScheme()`, `enhanceNode()`
  - `frontend/src/utils/asyncHandler.js` — 删除未使用的 `createAsyncHandlerFactory`
  - `frontend/src/components/AISettings/index.jsx` — 验证是否被消费，若未使用则删除
- **状态**: ⏳ 待执行

### M-212: 提取重复逻辑
- **严重度**: P2
- **文件**: `pkg/database/repository.go`, `pkg/ai/openai.go`, `pkg/ai/openai_llm.go`
- **操作**:
  1. `repository.go:extractTitle` 中 `regexp.MustCompile` 提升为包级变量
  2. 维度映射统一到 `dimensions.go` (配合 A-201)
- **状态**: ⏳ 待执行

---

## Phase 4: 前端性能 + 内存管理

### M-210: 前端内存泄漏 + 性能修复
- **严重度**: P1-P2
- **文件**: Editor.jsx, GraphPanel.jsx, FileTree.jsx, App.jsx
- **操作**:
  1. **Editor.jsx**: `md.render(content)` 包裹 `useMemo`
  2. **Editor.jsx**: 滚动同步 `timeoutRef` 添加 cleanup
  3. **GraphPanel.jsx**: `clickTimeout` 和初始 `setTimeout` 添加 cleanup
  4. **GraphPanel.jsx**: 分离数据获取和主题应用，避免主题变化重获取数据
  5. **FileTree.jsx**: `FileTreeRow` 添加 `React.memo`
  6. **App.jsx**: `shortcuts` 对象 `useMemo` 缓存
  7. **ChatPanel.jsx**: `setError('导出成功')` → 使用 toast 或 success 状态
- **状态**: ⏳ 待执行

---

## Phase 5: 编译验证 (由用户执行)

### V-201: 统一编译测试
- **操作**:
  ```bash
  go build ./...
  go test ./...
  cd frontend && npx vite build
  ```
- **状态**: ⏳ 等待所有 Phase 完成后执行

---

## 执行优先级排序

| 顺序 | 任务ID | 描述 | 预估改动 |
|------|--------|------|----------|
| 1 | M-201 | AI 死锁 + 线程安全 | ~30 行 |
| 2 | M-202 | 迁移死循环 | ~15 行 |
| 3 | M-203 | IndexAll 数据竞争 | ~40 行 |
| 4 | M-204 | 路径遍历防护 | ~30 行 |
| 5 | M-205 | Config boolean merge | ~40 行 |
| 6 | M-206 | RAG context + nil | ~20 行 |
| 7 | M-207 | Chat ticker panic | ~15 行 |
| 8 | M-208 | DB 静默错误 | ~10 行 |
| 9 | M-209 | Go 死代码清理 | -100 行 |
| 10 | M-212 | 重复逻辑提取 | ~20 行 |
| 11 | A-201 | 维度映射统一 | ~50 行 |
| 12 | A-202 | 服务初始化统一 | ~60 行 |
| 13 | M-210 | 前端性能修复 | ~50 行 |
| 14 | M-211 | 前端死代码清理 | -70 行 |
| 15 | V-201 | 编译验证 | 用户执行 |

---

## 不在本轮范围内的改进 (Future)

- ChatPanel 拆分为子组件 (M-213) — 改动大，风险高
- useAISettings 巨型 Hook 拆分 — 功能性改动不高
- ARIA 可访问性 — 独立工作流
- i18n 国际化 — 独立工作流  
- Toast 多类型支持 — 功能增强
- 弱加密密钥派生 (P0 安全) — 需要单独的安全审计
- N+1 查询优化 — 性能优化，需要基准测试
