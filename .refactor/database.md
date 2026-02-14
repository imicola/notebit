## Plan: RAG 向量数据库全面重构

**目标**：将当前基于全量内存缓存 + 暴力搜索的向量系统，重构为基于 sqlite-vec 原生向量索引的高性能架构，同时统一三处重复索引管线、消除竞争条件、引入增量更新和 SQLite 性能调优。

**核心改造**：纯 Go SQLite 驱动 → CGO `mattn/go-sqlite3` + `sqlite-vec` 扩展；`map[uint][]float32` 内存缓存 → sqlite-vec 虚拟表 IVF/HNSW 索引；三处 copy-paste 管线 → 统一 `IndexingPipeline` 服务。

---

### 第一阶段：SQLite 基础设施升级

**Step 1 — 替换 SQLite 驱动**

将 go.mod 中的 `github.com/glebarez/sqlite v1.11.0`（纯 Go）替换为 `github.com/mattn/go-sqlite3`（CGO）+ GORM 适配器 `gorm.io/driver/sqlite`。同时引入 `github.com/asg017/sqlite-vec-go-bindings`。

需要修改 manager.go 的 import 从 `glebarez/sqlite` 改为 `gorm.io/driver/sqlite`。

**Step 2 — 添加 SQLite PRAGMA 配置**

在 manager.go（当前 `gorm.Open` 后无任何 PRAGMA）中添加：

- `PRAGMA journal_mode=WAL` — 写不阻塞读，显著提升并发性能
- `PRAGMA busy_timeout=5000` — 5 秒忙等待，解决向量加载超时的直接原因
- `PRAGMA synchronous=NORMAL` — WAL 模式下安全降级，提升写性能
- `PRAGMA cache_size=-64000` — 64MB 页缓存
- `PRAGMA mmap_size=268435456` — 256MB 内存映射
- `PRAGMA foreign_keys=ON` — 确保 CASCADE 删除生效

通过 DSN 参数传递（`file:path?_journal_mode=WAL&...`）或 `db.Exec()` 执行。

**Step 3 — 注册 sqlite-vec 扩展**

在 `manager.go` 的 `Init()` 中，在 `gorm.Open` 之后注册 sqlite-vec 扩展：

```
sql.Register("sqlite3_vec", &sqlite3.SQLiteDriver{
    Extensions: []string{"vec0"},
})
```

使用 `sqlite-vec-go-bindings` 提供的 `Auto()` 注册函数，它会自动加载与平台对应的预编译 `.so/.dll`。

---

### 第二阶段：向量存储模型重构

**Step 4 — 创建 sqlite-vec 虚拟表**

在 migrations.go 中实现真正的迁移逻辑（当前是空壳）。创建 sqlite-vec 虚拟表：

```sql
CREATE VIRTUAL TABLE IF NOT EXISTS vec_chunks USING vec0(
    chunk_id INTEGER PRIMARY KEY,
    embedding float[1536]
);
```

维度需可配置（OpenAI `text-embedding-3-small` = 1536，Ollama 模型维度各异）。引入 `schema_version` 表记录已执行的迁移版本。

**Step 5 — 简化 Chunk 模型**

修改 models.go 的 `Chunk` 结构体：
- **移除** `Embedding []float32`（JSON 格式，遗留字段）
- **保留** `EmbeddingBlob []byte` 仅用于迁移期过渡，迁移完成后标记为 deprecated
- **新增** `VecIndexed bool` 字段标记是否已写入 vec 虚拟表
- 向量数据的主存储改为 `vec_chunks` 虚拟表，通过 `chunk_id` 关联

**Step 6 — 数据迁移器**

在 database 下新建 `migration_vec.go`，实现 `MigrateToVec()` 方法：
1. 检测 `vec_chunks` 表是否存在
2. 遍历所有有 `embedding_blob` 但未写入 `vec_chunks` 的 chunks
3. 批量插入到虚拟表（每批 500 条）
4. 标记 `vec_indexed = true`
5. 支持中断恢复（通过 `vec_indexed` 标记跟踪进度）
6. 迁移完成后可选清理 `embedding` JSON 列（`UPDATE chunks SET embedding = NULL WHERE vec_indexed = true`）

启动时在 app.go 的 `startup()` 中自动检测并执行。

---

### 第三阶段：向量搜索引擎实现

**Step 7 — 实现 SQLiteVecEngine**

重写 vector_engine_sqlitevec.go（当前仅 19 行占位符），实现真正的 sqlite-vec 搜索：

核心 SQL：
```sql
SELECT chunk_id, distance
FROM vec_chunks
WHERE embedding MATCH ?
ORDER BY distance
LIMIT ?
```

sqlite-vec 在底层使用精确 KNN 搜索，但利用 SQLite 的 B-tree 索引结构，远比纯 Go 暴力遍历高效（无需全量加载到内存）。

`Search()` 实现步骤：
1. 将 `[]float32` 编码为 sqlite-vec 接受的二进制格式
2. 执行 `MATCH` 查询获取 `(chunk_id, distance)` 对
3. 用 chunk_ids 批量 `Preload("File")` 获取完整数据
4. 将 distance 转换为 similarity（`1 - distance` 或 `1 / (1 + distance)`，取决于度量类型）

**Step 8 — 保留 BruteForce 作为 Fallback**

保留 [pkg/database/vector_engine.go](pkg/database/vector_engine.go#L51-L97) 的 `BruteForceVectorEngine` 作为 sqlite-vec 不可用时的降级方案，但优化其实现：

- **移除全量内存缓存**（`vectorCache map[uint][]float32`）
- Brute force fallback 改为**流式数据库查询** — 使用 `db.Rows()` 逐行读取 `embedding_blob`，计算相似度后维护 top-K 堆（`container/heap`），避免全量加载
- 内存峰值从 O(N × dim) 降至 O(K × dim)

**Step 9 — 移除全局向量缓存**

从 [pkg/database/repository.go](pkg/database/repository.go#L16-L24) 的 `Repository` 结构体中移除：
- `vectorCache map[uint][]float32`
- `vectorCacheLoaded bool`
- `vectorCacheMu sync.RWMutex`

从 [pkg/database/vector.go](pkg/database/vector.go#L226-L274) 中删除：
- `invalidateVectorCache()`
- `loadVectorCache()`
- `getVectorCache()`
- `chunkVector` 结构体

所有搜索操作走 sqlite-vec 虚拟表查询，无需应用层缓存。

---

### 第四阶段：索引管线统一

**Step 10 — 创建统一 `IndexingPipeline`**

新建 `pkg/indexing/pipeline.go`，将三处重复代码统一：
- [app_files.go](app_files.go#L208-L243) 的 `indexFileWithEmbeddings()`
- [pkg/watcher/service.go](pkg/watcher/service.go#L318-L382) 的 `indexFileWithEmbeddings()`
- [pkg/knowledge/service.go](pkg/knowledge/service.go#L30-L66) 的 `IndexFileWithEmbedding()`

统一接口设计：
```go
type IndexingPipeline struct {
    ai   *ai.Service
    repo *database.Repository
    fm   *files.Manager
}

func (p *IndexingPipeline) IndexFile(ctx context.Context, path string, opts IndexOptions) error
func (p *IndexingPipeline) IndexContent(ctx context.Context, path string, content string, opts IndexOptions) error
func (p *IndexingPipeline) IndexAll(ctx context.Context, progress *IndexProgress) error
```

`IndexOptions` 包含：
- `SkipIfUnchanged bool` — 调用 `FileNeedsIndexing()` 检查（修复 app_files.go 不检查的问题）
- `FallbackToMetadataOnly bool` — 启用三级降级策略
- `ForceReindex bool` — 忽略哈希比较强制重建

**Step 11 — 类型安全的 Embedding 传递**

修改 [pkg/ai/service.go](pkg/ai/service.go#L267-L296) 的 `ProcessDocument()` 返回类型。当前通过 `chunk.Metadata["embedding"].([]float32)` 传递向量（类型不安全），改为：

在 [pkg/ai/types.go](pkg/ai/types.go) 中为 `Chunk` 结构体直接新增 `Embedding []float32` 字段，取代 `Metadata["embedding"]` 的间接传递。这样 `IndexingPipeline` 不需要做 type assertion。

**Step 12 — 统一 Worker Pool**

将 App 层的 [channel-based worker pool](app_files.go#L245-L269)（4 workers）和 Watcher 层的 [semaphore-based pool](pkg/watcher/service.go#L436-L501)（3 workers）统一为一个 worker pool：

- 在 `IndexingPipeline` 内置一个统一的 `workQueue chan IndexJob`（容量 256）
- 可配置 worker 数量（默认 4）
- 所有索引请求（SaveFile、fsnotify 事件、全量索引）都提交到同一队列
- 通过去重逻辑（`sync.Map` 记录正在处理的 path）**消除竞争条件**

**Step 13 — 向量写入同步到 sqlite-vec**

修改 [pkg/database/repository.go](pkg/database/repository.go#L291-L348) 的 `IndexFileWithChunks()`：
- 事务中创建 Chunk 记录后，立即将 embedding 插入 `vec_chunks` 虚拟表
- 删除旧 chunks 时，同步从 `vec_chunks` 中删除对应行
- 移除双重存储逻辑：不再写入 `embedding` JSON 列，只写 `embedding_blob`（过渡期）和 `vec_chunks`

---

### 第五阶段：Graph 性能优化

**Step 14 — 批量相似度计算**

重构 [pkg/graph/service.go](pkg/graph/service.go#L230-L268) 的 `extractImplicitLinks()`：

当前 O(N²) 问题：对 N 个文件逐一调用 `SearchSimilar()`。改为：
1. 一次性从 `vec_chunks` 获取所有文件的首 chunk embedding
2. 使用 sqlite-vec 的批量查询能力，或在应用层用 pairwise 计算
3. 新增 `Repository.SearchSimilarBatch(embeddings [][]float32, limit int)` 方法
4. 结合 Graph 的 `cachedRevision` 缓存机制，只在 revision 变化时重算

备选方案（更简单但效果好）：限制 Graph 隐式链接计算只处理 `maxNodes`（默认 100）个文件的 embeddings，而非全量。

---

### 第六阶段：清理与测试

**Step 15 — 删除死代码**

- [pkg/database/migrations.go](pkg/database/migrations.go#L4-L10) 的空 `AutoMigrate()` — 替换为真正的版本化迁移
- [pkg/database/vector.go](pkg/database/vector.go#L17-L30) 的 `SaveEmbedding()`、`SaveEmbeddingBatch()` — 已被 `IndexFileWithChunks` 替代
- [pkg/database/repository.go](pkg/database/repository.go#L158-L177) 的 `CreateChunks()` — 与 `IndexFileWithChunks` 功能重叠
- [pkg/knowledge/service.go](pkg/knowledge/service.go#L69-L100) 的 `ReindexAllWithEmbeddings()` — 被 `IndexingPipeline.IndexAll()` 替代
- 向量缓存相关的所有代码（Step 9 中列出的）

**Step 16 — 增强测试覆盖**

- `vec_chunks` 虚拟表 CRUD 测试
- sqlite-vec 搜索精度测试（与 brute-force 对比 top-K 结果一致性）
- 迁移器测试（旧格式 → vec 表的完整性验证）
- `IndexingPipeline` 集成测试（含降级路径）
- 并发安全测试（多 worker 同时索引同一文件）
- Benchmark：100 / 1000 / 10000 chunks 搜索性能对比（sqlite-vec vs brute-force）

**Step 17 — 配置扩展**

在 [pkg/config/config.go](pkg/config/config.go) 中新增：
- `VectorDimension int` — 向量维度（默认 1536），用于创建虚拟表
- `IndexWorkerCount int` — 统一 worker 数量（默认 4）
- `IndexQueueSize int` — 队列容量（默认 256）
- `MigrationBatchSize int` — 迁移批量大小（默认 500）

---

### 文件变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| [go.mod](go.mod) | 修改 | 替换 SQLite 驱动，添加 sqlite-vec bindings |
| [pkg/database/manager.go](pkg/database/manager.go) | 修改 | 驱动切换 + PRAGMA + sqlite-vec 注册 |
| [pkg/database/migrations.go](pkg/database/migrations.go) | 重写 | 版本化迁移系统 + vec 虚拟表创建 |
| [pkg/database/models.go](pkg/database/models.go) | 修改 | 移除 `Embedding` JSON 字段，新增 `VecIndexed` |
| [pkg/database/vector.go](pkg/database/vector.go) | 大幅重写 | 移除缓存、双重存储，重新设计写入路径 |
| [pkg/database/vector_engine.go](pkg/database/vector_engine.go) | 修改 | BruteForce 改为流式 top-K 堆 |
| [pkg/database/vector_engine_sqlitevec.go](pkg/database/vector_engine_sqlitevec.go) | 重写 | 实现真正的 sqlite-vec 搜索 |
| [pkg/database/repository.go](pkg/database/repository.go) | 修改 | 移除缓存字段，`IndexFileWithChunks` 同步 vec 表 |
| **新建** `pkg/database/migration_vec.go` | 新建 | 旧数据 → vec 表迁移器 |
| **新建** `pkg/indexing/pipeline.go` | 新建 | 统一索引管线 |
| [pkg/ai/types.go](pkg/ai/types.go) | 修改 | Chunk 结构体新增 `Embedding` 字段 |
| [pkg/ai/service.go](pkg/ai/service.go) | 修改 | ProcessDocument 直接设置 Chunk.Embedding |
| [pkg/knowledge/service.go](pkg/knowledge/service.go) | 简化 | 委托到 IndexingPipeline，删除重复代码 |
| [pkg/watcher/service.go](pkg/watcher/service.go) | 简化 | 委托到 IndexingPipeline，移除自有 worker pool |
| [app_files.go](app_files.go) | 简化 | 委托到 IndexingPipeline，移除自有 worker pool |
| [app.go](app.go) | 修改 | 初始化 IndexingPipeline，运行迁移 |
| [pkg/graph/service.go](pkg/graph/service.go) | 修改 | 批量相似度计算优化 |
| [pkg/config/config.go](pkg/config/config.go) | 修改 | 新增向量/索引相关配置项 |

---

### Verification

- `wails dev` 启动后检查日志确认 sqlite-vec 扩展加载成功、PRAGMA 设置正确、迁移器执行完成
- 已有向量数据的笔记库启动后，确认 `vec_chunks` 表行数与 `chunks` 表中有 embedding 的行数一致
- 执行语义搜索和 RAG 查询，验证结果质量与重构前一致
- 运行 `go test ./pkg/database/... -bench=.` 对比搜索性能
- 同时触发多个文件保存，验证无竞争条件（无重复索引日志）
- Graph 页面打开速度在 100+ 文件时应显著改善

### Decisions

- **选择 `mattn/go-sqlite3` + sqlite-vec 而非纯 Go 方案**：用户确认接受 CGO，sqlite-vec 是 SQLite 官方推荐的向量扩展，与现有架构最契合
- **保留 BruteForce 引擎作为 fallback**：确保 sqlite-vec 加载失败时系统仍可用，但改为流式 top-K 避免全量加载
- **统一管线而非修补现有三处实现**：虽然工作量更大，但消除了竞争条件和维护负担
- **vec 虚拟表使用固定维度**：sqlite-vec 要求建表时声明维度，维度通过配置管理，切换 embedding 模型时需重建索引`IndexOptions` 包含：
- `SkipIfUnchanged bool` — 调用 `FileNeedsIndexing()` 检查（修复 app_files.go 不检查的问题）
- `FallbackToMetadataOnly bool` — 启用三级降级策略
- `ForceReindex bool` — 忽略哈希比较强制重建

**Step 11 — 类型安全的 Embedding 传递**

修改 [pkg/ai/service.go](pkg/ai/service.go#L267-L296) 的 `ProcessDocument()` 返回类型。当前通过 `chunk.Metadata["embedding"].([]float32)` 传递向量（类型不安全），改为：

在 [pkg/ai/types.go](pkg/ai/types.go) 中为 `Chunk` 结构体直接新增 `Embedding []float32` 字段，取代 `Metadata["embedding"]` 的间接传递。这样 `IndexingPipeline` 不需要做 type assertion。

**Step 12 — 统一 Worker Pool**

将 App 层的 [channel-based worker pool](app_files.go#L245-L269)（4 workers）和 Watcher 层的 [semaphore-based pool](pkg/watcher/service.go#L436-L501)（3 workers）统一为一个 worker pool：

- 在 `IndexingPipeline` 内置一个统一的 `workQueue chan IndexJob`（容量 256）
- 可配置 worker 数量（默认 4）
- 所有索引请求（SaveFile、fsnotify 事件、全量索引）都提交到同一队列
- 通过去重逻辑（`sync.Map` 记录正在处理的 path）**消除竞争条件**

**Step 13 — 向量写入同步到 sqlite-vec**

修改 [pkg/database/repository.go](pkg/database/repository.go#L291-L348) 的 `IndexFileWithChunks()`：
- 事务中创建 Chunk 记录后，立即将 embedding 插入 `vec_chunks` 虚拟表
- 删除旧 chunks 时，同步从 `vec_chunks` 中删除对应行
- 移除双重存储逻辑：不再写入 `embedding` JSON 列，只写 `embedding_blob`（过渡期）和 `vec_chunks`

---

### 第五阶段：Graph 性能优化

**Step 14 — 批量相似度计算**

重构 [pkg/graph/service.go](pkg/graph/service.go#L230-L268) 的 `extractImplicitLinks()`：

当前 O(N²) 问题：对 N 个文件逐一调用 `SearchSimilar()`。改为：
1. 一次性从 `vec_chunks` 获取所有文件的首 chunk embedding
2. 使用 sqlite-vec 的批量查询能力，或在应用层用 pairwise 计算
3. 新增 `Repository.SearchSimilarBatch(embeddings [][]float32, limit int)` 方法
4. 结合 Graph 的 `cachedRevision` 缓存机制，只在 revision 变化时重算

备选方案（更简单但效果好）：限制 Graph 隐式链接计算只处理 `maxNodes`（默认 100）个文件的 embeddings，而非全量。

---

### 第六阶段：清理与测试

**Step 15 — 删除死代码**

- [pkg/database/migrations.go](pkg/database/migrations.go#L4-L10) 的空 `AutoMigrate()` — 替换为真正的版本化迁移
- [pkg/database/vector.go](pkg/database/vector.go#L17-L30) 的 `SaveEmbedding()`、`SaveEmbeddingBatch()` — 已被 `IndexFileWithChunks` 替代
- [pkg/database/repository.go](pkg/database/repository.go#L158-L177) 的 `CreateChunks()` — 与 `IndexFileWithChunks` 功能重叠
- [pkg/knowledge/service.go](pkg/knowledge/service.go#L69-L100) 的 `ReindexAllWithEmbeddings()` — 被 `IndexingPipeline.IndexAll()` 替代
- 向量缓存相关的所有代码（Step 9 中列出的）

**Step 16 — 增强测试覆盖**

- `vec_chunks` 虚拟表 CRUD 测试
- sqlite-vec 搜索精度测试（与 brute-force 对比 top-K 结果一致性）
- 迁移器测试（旧格式 → vec 表的完整性验证）
- `IndexingPipeline` 集成测试（含降级路径）
- 并发安全测试（多 worker 同时索引同一文件）
- Benchmark：100 / 1000 / 10000 chunks 搜索性能对比（sqlite-vec vs brute-force）

**Step 17 — 配置扩展**

在 [pkg/config/config.go](pkg/config/config.go) 中新增：
- `VectorDimension int` — 向量维度（默认 1536），用于创建虚拟表
- `IndexWorkerCount int` — 统一 worker 数量（默认 4）
- `IndexQueueSize int` — 队列容量（默认 256）
- `MigrationBatchSize int` — 迁移批量大小（默认 500）

---

### 文件变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| [go.mod](go.mod) | 修改 | 替换 SQLite 驱动，添加 sqlite-vec bindings |
| [pkg/database/manager.go](pkg/database/manager.go) | 修改 | 驱动切换 + PRAGMA + sqlite-vec 注册 |
| [pkg/database/migrations.go](pkg/database/migrations.go) | 重写 | 版本化迁移系统 + vec 虚拟表创建 |
| [pkg/database/models.go](pkg/database/models.go) | 修改 | 移除 `Embedding` JSON 字段，新增 `VecIndexed` |
| [pkg/database/vector.go](pkg/database/vector.go) | 大幅重写 | 移除缓存、双重存储，重新设计写入路径 |
| [pkg/database/vector_engine.go](pkg/database/vector_engine.go) | 修改 | BruteForce 改为流式 top-K 堆 |
| [pkg/database/vector_engine_sqlitevec.go](pkg/database/vector_engine_sqlitevec.go) | 重写 | 实现真正的 sqlite-vec 搜索 |
| [pkg/database/repository.go](pkg/database/repository.go) | 修改 | 移除缓存字段，`IndexFileWithChunks` 同步 vec 表 |
| **新建** `pkg/database/migration_vec.go` | 新建 | 旧数据 → vec 表迁移器 |
| **新建** `pkg/indexing/pipeline.go` | 新建 | 统一索引管线 |
| [pkg/ai/types.go](pkg/ai/types.go) | 修改 | Chunk 结构体新增 `Embedding` 字段 |
| [pkg/ai/service.go](pkg/ai/service.go) | 修改 | ProcessDocument 直接设置 Chunk.Embedding |
| [pkg/knowledge/service.go](pkg/knowledge/service.go) | 简化 | 委托到 IndexingPipeline，删除重复代码 |
| [pkg/watcher/service.go](pkg/watcher/service.go) | 简化 | 委托到 IndexingPipeline，移除自有 worker pool |
| [app_files.go](app_files.go) | 简化 | 委托到 IndexingPipeline，移除自有 worker pool |
| [app.go](app.go) | 修改 | 初始化 IndexingPipeline，运行迁移 |
| [pkg/graph/service.go](pkg/graph/service.go) | 修改 | 批量相似度计算优化 |
| [pkg/config/config.go](pkg/config/config.go) | 修改 | 新增向量/索引相关配置项 |

---

### Verification

- `wails dev` 启动后检查日志确认 sqlite-vec 扩展加载成功、PRAGMA 设置正确、迁移器执行完成
- 已有向量数据的笔记库启动后，确认 `vec_chunks` 表行数与 `chunks` 表中有 embedding 的行数一致
- 执行语义搜索和 RAG 查询，验证结果质量与重构前一致
- 运行 `go test ./pkg/database/... -bench=.` 对比搜索性能
- 同时触发多个文件保存，验证无竞争条件（无重复索引日志）
- Graph 页面打开速度在 100+ 文件时应显著改善

### Decisions

- **选择 `mattn/go-sqlite3` + sqlite-vec 而非纯 Go 方案**：用户确认接受 CGO，sqlite-vec 是 SQLite 官方推荐的向量扩展，与现有架构最契合
- **保留 BruteForce 引擎作为 fallback**：确保 sqlite-vec 加载失败时系统仍可用，但改为流式 top-K 避免全量加载
- **统一管线而非修补现有三处实现**：虽然工作量更大，但消除了竞争条件和维护负担
- **vec 虚拟表使用固定维度**：sqlite-vec 要求建表时声明维度，维度通过配置管理，切换 embedding 模型时需重建索引