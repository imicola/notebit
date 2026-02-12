# Backend & AI Module Analysis

## Basic Information
- **Scope**: `pkg/ai`, `pkg/database` (Vector), `pkg/knowledge`, `pkg/watcher`
- **Entry Points**: `App.go` (Wails bindings), `watcher/service.go` (Auto-indexing)
- **Tech Stack**: Go, SQLite, GORM, Ollama/OpenAI

## Analysis Summary

### 1. Vector Database Performance (CRITICAL)
- **Initial State**: Naive O(N) search in `pkg/database/vector.go`. Loaded full content + JSON parsing for every chunk.
- **Problem**: 
  - `SearchSimilar` latency > 230ms for 1000 chunks.
  - JSON unmarshalling of `[]float32` was the bottleneck.
  - Excessive memory usage loading file content during search.
- **Refactoring Status**: **COMPLETED**
  - **Change**: Introduced `EmbeddingBlob` (binary storage) in `Chunk` model.
  - **Optimization**: `SearchSimilar` now selects only `ID` and `Blob`, avoiding JSON parsing and content loading.
  - **Result**: Benchmark shows **~30ms** for 1000 chunks (7.7x speedup).

### 2. Architecture Layering
- **Initial State**: `App` struct in `app.go` contained all indexing logic (`IndexFileWithEmbedding`, `FindSimilar`).
- **Problem**: `App` was becoming a God Object, mixing UI binding with business logic.
- **Refactoring Status**: **COMPLETED**
  - **Change**: Extracted `pkg/knowledge/service.go`.
  - **Role**: `KnowledgeService` now handles the coordination between `Files`, `Database`, and `AI`.
  - **Benefit**: `App.go` is now a thin wrapper for Wails.

### 3. AI Service (`pkg/ai`)
- **Status**: Well structured.
- **Components**:
  - `Service`: Central manager.
  - `Provider` interface: Clean abstraction for OpenAI/Ollama.
  - `Chunker` interface: Strategy pattern for text splitting.
- **Configuration**: Unified in `pkg/config`.

## Recommendations for Next Steps
1.  **SQLite-vec Integration**: For production scaling > 100k chunks, replace the pure Go cosine similarity with `sqlite-vec` extension.
2.  **Re-indexing**: Users need to re-index existing files to populate the new `EmbeddingBlob` field. A UI prompt should be added.
3.  **Testing**: Add more unit tests for `KnowledgeService` to ensure edge cases (e.g., file deletion, rename) are handled correctly.
