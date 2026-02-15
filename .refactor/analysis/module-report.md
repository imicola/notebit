# Module Layer Analysis Report (Round 2)

## Date: 2026-02-15
## Status: Analysis Complete ✅

---

## 分析覆盖

| 模块 | 分析状态 | 问题数 | 报告位置 |
|------|----------|--------|----------|
| pkg/ai/ | ✅ | 9 | modules/ai-service.md |
| pkg/database/ | ✅ | 7 | modules/database.md |
| pkg/config/ | ✅ | 3 | modules/other-packages.md |
| pkg/files/ | ✅ | 2 | modules/other-packages.md |
| pkg/indexing/ | ✅ | 4 | modules/other-packages.md |
| pkg/rag/ | ✅ | 4 | modules/other-packages.md |
| pkg/graph/ | ✅ | 3 | modules/other-packages.md |
| pkg/watcher/ | ✅ | 3 | modules/other-packages.md |
| pkg/chat/ | ✅ | 4 | modules/other-packages.md |
| pkg/logger/ | ✅ | 3 | modules/other-packages.md |
| Frontend | ✅ | 18 | modules/frontend.md |

## 问题严重度汇总

### P0 Critical (必须修复)

| # | 问题 | 包 | 影响 |
|---|------|----|------|
| 1 | 路径遍历漏洞 | `files` | **安全** — 可访问基目录外文件 |
| 2 | `GetStatus()` RLock 重入死锁 | `ai` | 运行时死锁 |
| 3 | `MigrateToVec` 死循环 | `database` | 空 blob chunk 导致无限循环 |
| 4 | `IndexAll` 数据竞争 | `indexing` | 并发写无同步 |
| 5 | Editor 重建丢失状态 | 前端 | 用户丢失光标和撤销历史 |

### P1 High (应尽快修复)

| # | 问题 | 包 |
|---|------|----|
| 6 | Boolean merge 默认值覆盖 | `config` |
| 7 | `rag.buildContext` nil panic | `rag` |
| 8 | `context` 变量遮蔽包名 | `rag` |
| 9 | `VectorDimension` 不被合并 | `config` |
| 10 | Stop/IndexFile 死锁风险 | `indexing` |
| 11 | BackupTicker panic 风险 | `chat` |
| 12 | `ChunkText()` 线程安全缺陷 | `ai` |
| 13 | 静默吞掉 vec_chunks 删除错误 | `database` |
| 14 | ChatPanel 职责爆炸 (523行单组件) | 前端 |
| 15 | Markdown render 未缓存 | 前端 |

### P2 Medium (改善代码质量)

| # | 问题 | 包 |
|---|------|----|
| 16 | `errors.go` 死代码 | `ai` |
| 17 | `stat.go` 死代码 | `watcher` |
| 18 | Metadata 冗余写入 | `ai` |
| 19 | 三处重复维度映射 | `ai`, `config` |
| 20 | Regex 重编译 | `database` |
| 21 | N+1 查询 | `chat` |
| 22 | 递归 watch 缺失 | `watcher` |
| 23 | HTTP 请求逻辑重复 | `ai` |
| 24 | 无配置值域验证 | `config` |
| 25 | context 未传播 | `indexing` |
| 26 | BuildGraph 持锁做 DB 查询 | `graph` |
| 27 | 非唯一 MessageID | `rag` |
| 28 | Singleton Reset 竞态 | `database` |
| 29 | 过期注释 | `database` |
| 30 | Smart Dropping 逻辑失效 | `logger` |
| 31 | GraphPanel setTimeout 泄漏 | 前端 |
| 32 | Editor 滚动同步 timeout 泄漏 | 前端 |
| 33 | FileTreeRow 缺 React.memo | 前端 |
| 34 | useKeyboardShortcuts 重注册 | 前端 |
| 35 | 前端死代码清理 | 前端 |
| 36 | Graph 主题变化触发数据重获取 | 前端 |
| 37 | Profile 管理代码重复 | 前端 |
| 38 | Goroutine 泄漏风险 (Ollama batch) | `ai` |
| 39 | ChatPanel error 状态语义滥用 | 前端 |

### P3 Low (可选改进)

| # | 问题 |
|---|------|
| 40-50 | 硬编码值、注释代码、i18n、可访问性、返回类型等 |

---

## Module Layer 任务清单

| ID | 任务 | 涉及问题# |
|----|------|-----------|
| M-201 | 修复 AI GetStatus 死锁 + ChunkText 线程安全 | 2, 12 |
| M-202 | 修复 MigrateToVec 死循环 | 3 |
| M-203 | 修复 IndexAll 数据竞争 + Stop 死锁 | 4, 10 |
| M-204 | 添加路径遍历防护 | 1 |
| M-205 | 修复 Config boolean merge + VectorDimension | 6, 9 |
| M-206 | 修复 RAG context 遮蔽 + nil panic | 7, 8 |
| M-207 | 修复 Chat BackupTicker panic | 11 |
| M-208 | 修复 Database 静默错误吞掉 | 13 |
| M-209 | 清理死代码 (Go) | 16, 17, 18, 29 |
| M-210 | 修复前端内存泄漏 + 性能 | 5, 15, 31, 32, 33, 34, 36 |
| M-211 | 前端死代码清理 | 35 |
| M-212 | 提取重复逻辑 (维度映射、Regex) | 19, 20, 23 |
| M-213 | ChatPanel 拆分 + error 语义修复 | 14, 39 |
