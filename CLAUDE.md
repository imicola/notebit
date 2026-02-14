# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Notebit** is a Local-First Markdown note-taking application for PKM enthusiasts and researchers. It combines a distraction-free writing environment ("The Sanctuary") with background AI-powered knowledge management ("The Silent Curator").

**Philosophy**: "Write for Humans, Manage by Silicon" - pure editor during writing, AI processes notes after the fact.

## Tech Stack

- **Frontend**: React 18.2 + Vite 3.0, CodeMirror 6 (editor), Tailwind CSS, vis-network (graph viz)
- **Backend**: Go 1.24 + Wails v2.11 (desktop app framework that binds Go to React)
- **Database**: SQLite with **CGO driver** (`mattn/go-sqlite3`) + **sqlite-vec extension** for native vector search
- **AI**: OpenAI API or Ollama (local) for embeddings and LLM chat
- **Styling**: Custom Obsidian-inspired theme system with CSS variables
- **File Watching**: fsnotify for automatic file indexing
- **Logging**: Enterprise-grade async logger with file rotation and Kafka support

## Development Commands

```bash
# Live development (hot reload for both Go and React)
wails dev

# Production build
wails build

# Frontend-only development (if needed)
cd frontend && npm run dev
cd frontend && npm run build
```

The Wails dev server runs on `http://localhost:34115` for browser-based debugging with Go method access.

## Architecture

### Go Backend

#### `pkg/logger/` - Enterprise Logging System

Enterprise-grade logging with async buffering and file rotation:
- **Features**: Async batched writes, file rotation (100MB max, 15 backups), Kafka output support
- **Context-aware**: Trace IDs, structured fields, performance timing
- **Smart dropping**: Drops DEBUG logs first when buffer is full
- **API**: `Info/Warn/Error/Fatal` + `*Ctx` variants + `*WithFields` + `*WithDuration` variants
- **Metrics**: Queue length, flush latency, batch sizes, dropped logs

```go
logger.Info("Message")
logger.InfoWithDuration(ctx, duration, "Operation completed")
logger.InfoWithFields(ctx, map[string]interface{}{"key": "value"}, "Message")
```

#### `pkg/files/` - File System Manager

The `files.Manager` struct handles all file I/O operations:
- **SetBasePath**: Sets the working directory (persisted to localStorage)
- **ListFiles**: Returns recursive `FileNode` tree structure (sorted: dirs first, then alphabetically)
- **ReadFile/SaveFile/CreateFile/DeleteFile/RenameFile**: CRUD operations on markdown files
- Thread-safe with `sync.RWMutex` for concurrent access

All file paths are relative to `basePath`. The manager converts to absolute paths internally and uses `FileSystemError` for error wrapping.

#### `pkg/ai/` - AI Service Layer

The `ai.Service` manages embedding generation and text chunking:

**Embedding Providers** (`EmbeddingProvider` interface):
- `OpenAIProvider`: For OpenAI-compatible APIs (supports custom base URLs)
- `OllamaProvider`: For local Ollama instances

**LLM Providers** (`LLMProvider` interface) - for chat completion:
- `OpenAILLMProvider`: OpenAI chat models (gpt-4o, gpt-4o-mini, gpt-3.5-turbo)

**Chunking Strategies** (`ChunkingStrategy` implementations):
- `fixed`: Fixed-size chunks with overlap
- `heading`: Header-based chunking (preserves document structure)
- `sliding`: Sliding window with configurable step
- `sentence`: Sentence boundary-aware chunking

**Key Methods**:
- `ProcessDocument(text)`: Chunks text and generates embeddings for all chunks
- `GenerateEmbedding(text)`: Single embedding generation
- `GenerateEmbeddingsBatch(texts)`: Batch embedding with configurable batch size
- `GetStatus()`: Returns current provider, model, chunking strategy, and health

**Type Safety Improvement** ⚡:
- `TextChunk` now has direct `Embedding []float32` and `ModelName string` fields
- Eliminates type assertions from old `Metadata map[string]interface{}` approach

#### `pkg/database/` - SQLite + Vector Storage ⚡ **REFACTORED**

**Vector Search Engine Architecture**:
- **sqlite-vec Engine**: Native KNN search using virtual table `vec_chunks` with `MATCH` operator
- **BruteForce Engine**: Optimized streaming fallback with O(K·log(K)) top-K heap (was O(N) full-cache)
- **Auto-fallback**: Gracefully downgrades to BruteForce if sqlite-vec unavailable

Uses `database.Manager` with GORM ORM + CGO SQLite driver. Key models:
- `File`: Path, title, content hash, last modified, file size
- `Chunk`: Content, heading, **`EmbeddingBlob []byte`** (binary format), embedding model, timestamps, **`VecIndexed bool`** flag
- `Tag`: Tags for file categorization with many-to-many relationship

**Repository Pattern**: `Repository` struct provides data access:
- `IndexFileWithChunks(path, content, chunks)`: Atomic file+chunks insertion + **synchronous `vec_chunks` write**
- `FileNeedsIndexing(path, content)`: SHA256 hash comparison for incremental updates
- `SearchSimilar(embedding, limit)`: **Pluggable vector engine** with auto-fallback
- `GetStats()`: Database statistics (file/chunk/tag counts)
- `DeleteFile(path)`: Cascade deletes chunks + vec_chunks entries
- `SetVectorEngine()`/`GetVectorEngine()`: Runtime engine switching

**SQLite Optimizations**:
- WAL mode (`journal_mode=WAL`)
- 64MB page cache (`cache_size=-64000`)
- 256MB memory-mapped I/O (`mmap_size=268435456`)
- 5s busy timeout for concurrent access
- Foreign keys enabled

**Migration System** (`pkg/database/migrations.go`, `migration_vec.go`):
- Versioned schema migrations with `schema_version` table
- **Automatic `embedding_blob` → `vec_chunks` migration** on startup (batch size: 500)
- Idempotent via `vec_indexed` flag (allows resume after crash)
- Background goroutine execution with progress logging

Database stored at `./data/db.sqlite` relative to app directory.

#### `pkg/config/` - Configuration Management

Singleton pattern with `config.Get()`. JSON-based config stored at `%USERPROFILE%\AppData\Roaming\notebit\config.json`:
- `AIConfig`: Provider selection, OpenAI/Ollama settings, batch size, **`VectorDimension`** (default 1536), **`VectorSearchEngine`**
- `ChunkingConfig`: Strategy, sizes, overlap, heading preservation
- `WatcherConfig`: Enable/disable, debounce delay, worker count
- `LLMConfig`: LLM provider (OpenAI), model, temperature, max tokens
- `RAGConfig`: Max context chunks, temperature, system prompt
- `GraphConfig`: Similarity threshold, max nodes, implicit links toggle
- **`IndexingConfig`**: Worker count (default 4), queue size (default 100), migration batch size (default 500)

**Defaults**: Ollama preferred (local-first), heading-based chunking, 500ms debounce.

#### `pkg/indexing/` - Unified Indexing Pipeline ⚡ **NEW**

The `IndexingPipeline` provides centralized, thread-safe file indexing:
- **4-worker concurrent pool** (configurable via `IndexingConfig.WorkerCount`)
- **100-item buffered queue** (configurable via `IndexingConfig.QueueSize`)
- **Path deduplication** map to prevent duplicate concurrent indexing
- **3-level fallback**: Full embeddings → Chunks-only → Metadata-only (graceful degradation on AI failure)

**Key Methods**:
- `Enqueue(path, content, opts)`: Async indexing (non-blocking)
- `IndexFile(ctx, path, opts)`: Sync indexing (waits for completion)
- `IndexContent(ctx, path, content, opts)`: Direct content indexing
- `IndexAll(ctx, paths, opts)`: Batch indexing with progress tracking

**IndexOptions**:
- `SkipIfUnchanged bool`: Check content hash before indexing
- `FallbackToMetadataOnly bool`: Enable graceful degradation
- `ForceReindex bool`: Ignore hash comparison

**Architecture Impact**:
- Replaced 3 duplicate indexing implementations in `app_files.go`, `watcher/service.go`, `knowledge/service.go`
- Eliminated old `indexQueue chan indexJob` pattern
- All file operations (SaveFile, CreateFile, file watcher events) now use pipeline

#### `pkg/watcher/` - File System Watcher

The `watcher.Service` provides automatic file indexing:
- **Debouncing**: Configurable delay (default 500ms) to avoid duplicate processing
- **Event Types**: Create, Write, Remove, Rename (via fsnotify)
- **Ignored**: `.git`, `node_modules`, `.idea`, temp files (`*.swp`, `*~`, `*.tmp`)
- **Integration**: Uses `IndexingPipeline` for all indexing operations (no duplicate logic)

#### `pkg/knowledge/` - Semantic Search Service

The `knowledge.Service` provides semantic search capabilities:
- `IndexFileWithEmbedding(path)`: Index a single file with embeddings
- `ReindexAllWithEmbeddings()`: Full re-single file (delegates to `IndexingPipeline`)
- `ReindexAllWithEmbeddings()`: Full re-indexing using pipeline's `IndexAll()` batch API
- `FindSimilar(content, limit)`: Find semantically similar notes (max 8000 chars)
- `GetSimilarityStatus()`: Check availability of semantic search (includes vector engine info)
#### `pkg/rag/` - RAG Chat Service

The `rag.Service` implements Retrieval-Augmented Generation:
- **Query Flow**: Generate query embedding → Search similar chunks → Build context → Generate completion
- `Query(ctx, query)`: Returns response with source citations
- `ChatMessage`/`ChunkRef`: Message and source data structures
- `GetStatus()`: LLM provider, model, database status
- Streaming responses via Wails events (`rag_chunk`)

#### `pkg/graph/` - Knowledge Graph Visualization

The `graph.Service` builds knowledge graphs:
- **Node Types**: File nodes with size based on connections
- **Link Types**:
  - `explicit`: Wiki-style `[[links]]` parsed from markdown
  - `implicit`: Semantic similarity (configurable threshold)
- `BuildGraph()`: Returns nodes and links for vis-network rendering
- Configurable max nodes (default 100) for performance

### React Frontend Structure

- **`App.jsx`**: Main container, manages:
  - File tree state and current file selection
  - Sidebar width (resizable) and collapse state
  - Zen mode (F11), command palette (Cmd+K), settings modal
  - Folder path persistence (`localStorage`)
  - Panel management (chat, graph, similar notes)

- **`components/Editor.jsx`**: CodeMirror 6-based editor with:
  - Three view modes: edit, split, preview
  - Synchronized scrolling between editor and preview
  - Custom syntax highlighting (Obsidian-style: `==highlight==`, `[[wiki]]`)
  - Save indicator (unsaved changes)
  - Markdown rendering via `markdown-it` with GitHub alerts

- **`components/FileTree.jsx`**: Recursive tree component with expand/collapse

- **`components/CommandPalette.jsx`**: Fuzzy search over files (Fuse.js) + command execution

- **`components/SettingsModal.jsx`**: Font customization (interface + text)

- **`components/AIStatusIndicator.jsx`**: Real-time AI service health indicator

- **`components/SimilarNotesSidebar.jsx`**: Semantic search results panel (auto-opens on save, 500ms debounce)

- **`components/ChatPanel.jsx`**: RAG chat interface with streaming responses

- **`components/GraphPanel.jsx`**: Interactive knowledge graph viewer (vis-network)

- **`components/AISettings.jsx`**: Comprehensive AI settings with tabs (Embeddings, LLM Chat, RAG, Graph)

### Theme System

Uses CSS variables defined in `style.css` mapped through Tailwind config:
- Base colors: `primary`, `secondary`, `normal`, `muted`, `faint`
- Modifiers: `modifier-hover`, `modifier-border`, `modifier-border-focus`
- Accent colors for syntax highlighting (red, orange, yellow, green, cyan, blue, purple, pink)

Theme variables use HSL with `--accent-h/l/s` for easy theming. Apply by setting CSS variables on `:root` or through settings.

### Wails Go-React Binding Pattern

Go methods are automatically bound to React via Wails codegen:
1. Define method on `App` struct in `app.go`
2. Import from `../wailsjs/go/main/App` in React
3. Call as Promise: `Greet(name).then(result)`

**Streaming Events**: For real-time updates (e.g., RAG chat):
```go
EventsEmit(ctx, "rag_chunk", RagChunkData{Content: "...", Done: false})
```

Generated bindings are in `frontend/wailsjs/` - do not edit these directly.

### Key Constraints

- **Local-First**: Notes stored as plain `.md` files; embeddings in `./data/db.sqlite`
- **Performance**: Cold start < 2s, vector search < 200ms, non-blocking UI (use goroutines)
- **Privacy**: No external calls by default; Ollama preferred over cloud APIs
- **Data Integrity**: App must function 100% even if AI services crash
- **NO AI autocomplete**: The editor should never show ghost text/inline suggestions during typing
- **Thread Safety**: All Go services use appropriate mutex types (`sync.RWMutex`, `sync.Mutex`)

## Module Status

- **Module A - The Sanctuary**: Distraction-free Markdown editor ✅
- **Module B - The Silent Curator**: Background embedding and semantic search ✅
- **Module C - Knowledge Review**: RAG-based chat interface ✅
- **Module D - Knowledge Graph**: Graph visualization ✅

## Current Implementation Status

**Completed Features:**
- File tree with nested folders and markdown file filtering
- Markdown editor with syntax highlighting (CodeMirror 6)
- Split view with synchronized scroll
- Obsidian-style syntax (`==highlight==`, `[[wiki]]`)
- Zen mode (F11), command palette (Cmd+K)
- Resizable sidebar, collapsible
- Folder persistence (last opened folder restored on startup)
- Custom fonts (interface + text)
- Toast notifications for save confirmation
- Enterprise logging with file rotation and async batching
- Semantic search with similar notes sidebar
- RAG chat with streaming responses
- Knowledge graph visualization (explicit + implicit links)
- Comprehensive AI settings (embeddings, LLM, RAG, graph tabs)
- Real-time AI service status indicator
## Recent Refactoring (Feb 2026) 🚀

### Motivation
Previous implementation suffered from:
- **Memory inefficiency**: Full vector cache loaded into memory (O(N×1536×4 bytes))
- **Slow vector search**: Brute-force O(N) without indexing
- **No incremental updates**: Cache invalidation required full reload
- **Duplicate indexing code**: 3 separate implementations across app/watcher/knowledge
- **Type safety issues**: Heavy reliance on `map[string]interface{}` with type assertions

### Key Changes

#### 1. Vector Storage Architecture (2026-02-14)
**Before**:
- Pure Go SQLite driver (no CGO)
- Embeddings stored as JSON `[]float32` in `chunks.embedding`
- Global in-memory cache `map[uint][]float32`
- Full cache reload on any update

**After**:
- **CGO SQLite driver** (`mattn/go-sqlite3`) + `sqlite-vec` extension
- **`vec_chunks` virtual table** for native KNN search via `MATCH` operator
- Binary blob storage (`EmbeddingBlob []byte`) for backward compatibility
- **Zero global cache** - all searches query database directly
- **Pluggable engine system**: `VectorEngineSQLiteVec` (native) ↔ `VectorEngineBruteForce` (fallback)

**Performance Impact**:
- Memory: O(N) → O(K) where K = search limit (typically 5-10)
- Search: O(N) brute-force → O(log N) index lookup (sqlite-vec) or O(N·log K) streaming heap (brute-force)
- Updates: O(N) cache rebuild → O(1) incremental insert

#### 2. Automatic Migration System
**`pkg/database/migration_vec.go`**:
- Automatic migration from `embedding_blob` → `vec_chunks` on startup
- Batch processing (500 chunks/batch) for large datasets
- **Idempotent via `vec_indexed` flag** - can resume after crash/restart
- Background goroutine execution (non-blocking startup)
- Graceful handling of missing/corrupted data

#### 3. Unified Indexing Pipeline
**`pkg/indexing/pipeline.go`** - NEW service to eliminate code duplication:

**Replaces**:
- `app_files.go`: `indexFileWithEmbeddings()`, `startIndexWorkers()`, `indexQueue chan`
- `watcher/service.go`: `indexFileWithEmbeddings()`, `indexFileMetadataOnly()`, worker pool
- `knowledge/service.go`: Direct `ProcessDocument()` + `IndexFileWithChunks()` calls

**Features**:
- Configurable worker pool (default 4 workers)
- Buffered task queue (default 100 items)
- Path deduplication map (prevents concurrent duplicate indexing)
- 3-level fallback: embeddings → chunks-only → metadata-only
- Both async (`Enqueue`) and sync (`IndexFile`) APIs
- Batch processing (`IndexAll`) with progress tracking

**Code Reduction**:
- ~200 lines of duplicate indexing logic eliminated
- Single source of truth for all file indexing
- Consistent error handling and fallback behavior

#### 4. Type Safety Improvements
**Before**:
```go
chunk.Metadata["embedding"].([]float32)  // Type assertion required
chunk.Metadata["embedding_model"].(string)
```

**After**:
```go
chunk.Embedding  // []float32, direct field access
chunk.ModelName  // string, direct field access
chunk.GetEmbedding()  // Helper for database model
```

**Impact**:
- Eliminates ~50+ type assertions across codebase
- Compile-time type checking instead of runtime panics
- Clearer data flow from AI service → database

#### 5. Configuration Extensions
New fields in `config.json`:
```json
{
  "ai": {
    "vector_dimension": 1536,
    "vector_search_engine": "sqlite-vec"
  },
  "indexing": {
    "worker_count": 4,
    "queue_size": 100,
    "migration_batch_size": 500
  }
}
```

#### 6. Database Optimizations
- **WAL mode** for better concurrent access
- **256MB mmap** for faster I/O
- **64MB page cache** for hot data
- **5s busy timeout** for write contention
- Foreign keys strictly enforced

### Backward Compatibility
- ✅ Old `embedding_blob` data automatically migrated
- ✅ Graceful fallback to BruteForce if sqlite-vec unavailable
- ✅ Config file backward compatible (new fields use defaults)
- ✅ API surface unchanged (frontend unaffected)

### Files Modified (23 total)
**Core Infrastructure**:
- `go.mod` - Added CGO dependencies
- `pkg/database/manager.go` - SQLite PRAGMA, vec extension loading
- `pkg/database/models.go` - Added `VecIndexed`, removed JSON `Embedding`, added `GetEmbedding()`
- `pkg/database/migrations.go` - NEW versioned migration system
- `pkg/database/migration_vec.go` - NEW automatic vec_chunks migrator
- `pkg/database/repository.go` - Removed cache, sync vec_chunks on insert/delete
- `pkg/database/vector.go` - Removed cache invalidation
- `pkg/database/vector_engine.go` - Optimized BruteForce to streaming heap
- `pkg/database/vector_engine_sqlitevec.go` - Complete KNN implementation

**Indexing Pipeline**:
- `pkg/indexing/pipeline.go` - NEW unified indexing service

**Type Safety**:
- `pkg/ai/types.go` - Added `Embedding`, `ModelName` fields to `TextChunk`
- `pkg/ai/service.go` - Populate embedding fields directly

**Configuration**:
- `pkg/config/config.go` - Added `IndexingConfig`, `VectorDimension`

**Service Integration**:
- `app.go` - Initialize pipeline, remove indexQueue
- `app_files.go` - Use pipeline for SaveFile/CreateFile
- `app_search.go` - Added nil checks for knowledge service
- `pkg/watcher/service.go` - Removed duplicate indexing, use pipeline
- `pkg/knowledge/service.go` - Delegate to pipeline, batch operations
- `pkg/graph/service.go` - Use `GetEmbedding()` helper

**Tests**:
- `pkg/database/vector_test.go` - Fixed EmbeddingBlob references
- `pkg/database/vector_engine_test.go` - Updated test data

### Performance Benchmarks (Estimated)

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| **Memory (10K notes)** | ~245 MB cache | ~160 KB heap | **1500x less** |
| **Search (sqlite-vec)** | N/A | ~10-50ms | **New capability** |
| **Search (brute-force)** | ~200ms | ~80ms | **2.5x faster** |
| **Incremental update** | ~500ms rebuild | ~5ms insert | **100x faster** |
| **Startup (existing DB)** | Immediate | +2s migration | One-time cost |

### Known Limitations
- sqlite-vec requires CGO (not available on all platforms/builds)
- BruteForce fallback still O(N), but optimized with heap
- Migration is one-way (no downgrade path to old schema)

### Future Enhancements
- [ ] Graph O(N²) optimization: Batch `SearchSimilarBatch()` for implicit links
- [ ] Cleanup old `embedding_blob` column after migration verification
- [ ] HNSW index for brute-force engine (pure Go alternative to sqlite-vec)
- [ ] Vector compression (e.g., product quantization) for memory/storage reduction