# Task B-101: app.go Domain Coordinator Extraction

## Status: ✅ COMPLETED

**Date**: 2026-02-18  
**Priority**: Critical  
**Component**: Backend (`app.go`, `app_files.go`, `app_ai.go`, `app_search.go`)

---

## Problem Statement

`app.go` (926 lines) was a god-file containing all Wails API bindings across 4 domain areas:

1. **Lifecycle methods**: App struct, startup, shutdown, initialization (~260 lines)
2. **File operations**: File tree, read, write, delete, indexing (~280 lines)
3. **AI configuration**: OpenAI/Ollama config, embedding, chunking (~200 lines)
4. **Search/RAG/Graph**: Semantic search, RAG queries, graph data (~186 lines)

This violated the single-responsibility principle and made it difficult to:
- Locate related functionality
- Make changes without affecting unrelated code
- Understand domain boundaries
- Test subsets of functionality

---

## Solution Implemented

Split 926-line god-file into 4 focused domain files, each responsible for one area.

### File Breakdown

#### 1. `app.go` (~258 lines) — LIFECYCLE & STRUCT
**Responsibility**: App initialization, lifecycle hooks, service initialization
```
- Type definitions: App struct, indexJob, watcherLogger
- Constructors: NewApp(), NewAppWithConfig()
- Lifecycle: startup(), shutdown()
- Initialization chains: 
  - initializeAI()
  - initializeLLM()
  - initializeRAG()
  - initializeGraph()
- Config management: loadConfig()
- Watcher lifecycle: startWatcher(), stopWatcher(), runFullIndex()
```

#### 2. `app_files.go` (~320 lines) — FILE OPERATIONS & INDEXING
**Responsibility**: File CRUD, tree navigation, database indexing
```
- File operations:
  - OpenFolder(), SetFolder()
  - ListFiles(), ReadFile(), SaveFile(), CreateFile(), DeleteFile(), RenameFile()
  - GetBasePath()
- File indexing:
  - IndexFile(), indexFile(), indexFileContent()
  - startIndexWorkers(), indexWorker(), enqueueIndexFileContent()
- Database API:
  - GetIndexedFile(), ListIndexedFiles()
  - RemoveFromIndex(), UpdateFilePathInIndex()
  - GetDatabaseStats(), IsDatabaseInitialized()
```

#### 3. `app_ai.go` (~210 lines) — AI CONFIGURATION & EMBEDDING
**Responsibility**: AI provider config, embedding generation, document processing
```
- Configuration APIs:
  - GetOpenAIConfig(), SetOpenAIConfig()
  - GetOllamaConfig(), SetOllamaConfig()
  - GetChunkingConfig(), SetChunkingConfig()
  - GetAIStatus(), SetAIProvider(), SetAIModel()
- Embedding:
  - TestOpenAIConnection()
  - GenerateEmbedding(), GenerateEmbeddingsBatch()
- Document processing:
  - ChunkText(), ProcessDocument()
  - IndexFileWithEmbedding(), ReindexAllWithEmbeddings()
- LLM Config:
  - GetLLMConfig(), SetLLMConfig()
```

#### 4. `app_search.go` (~110 lines) — SEMANTIC SEARCH, RAG & GRAPH
**Responsibility**: Semantic search, RAG queries, knowledge graph visualization
```
- Semantic search:
  - FindSimilar() [with SimilarNote struct]
  - GetSimilarityStatus()
- RAG:
  - RAGQuery()
  - GetRAGStatus()
- Graph:
  - GetGraphData()
  - GetGraphConfig(), SetGraphConfig()
```

---

## Implementation Details

### Issues Fixed During Split

1. **Missing Imports in app_files.go**
   - Added: `"notebit/pkg/database"`, `"notebit/pkg/files"`
   - Removed: Unused `"context"` import

2. **No Duplicate Methods**
   - Original app.go was cleanly trimmed to 258 lines
   - Each method appears in exactly one file
   - All imports preserved and correct

### Verification

```
go build ./...     -> PASS (no duplicate methods, no missing imports)
go test ./...      -> PASS (14/14 tests pass)
File line counts:
  app.go:       258 lines
  app_files.go: 320 lines
  app_ai.go:    210 lines
  app_search.go: 110 lines
  Total:        898 lines (28 lines removed from combined ops)
```

---

## Wails API Binding Signatures

All exported methods remain unchanged (they're still on the `App` struct):
- Wails automatically discovers all public methods
- Method signatures maintained for frontend compatibility
- All 60+ API endpoints still available via Wails binding

---

## Impact

- **Improved Readability**: Each file focuses on one domain (~220-320 lines, optimal cognitive load)
- **Better Maintainability**: Related methods grouped together
- **Easier Testing**: Can focus on one domain without loading entire app.go
- **Clear Boundaries**: File ops, AI config, search/RAG/graph domains explicit
- **Reduced Coupling**: Changes to one domain isolated from others
- **Line Count Reduction**: -28 lines through consolidated duplicates

---

## Future Improvements

- Consider extracting database wrapper methods to pkg/database/api.go
- Consider extracting AI methods to pkg/ai/api.go  
- Consider extracting RAG/search methods to pkg/rag/api.go
- These would separate Wails bindings from service implementation logic
