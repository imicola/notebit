# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Notebit** is a Local-First Markdown note-taking application for PKM enthusiasts and researchers. It combines a distraction-free writing environment ("The Sanctuary") with background AI-powered knowledge management ("The Silent Curator").

**Philosophy**: "Write for Humans, Manage by Silicon" - pure editor during writing, AI processes notes after the fact.

## Tech Stack

- **Frontend**: React 18.2 + Vite 3.0, CodeMirror 6 (editor), Tailwind CSS, vis-network (graph viz)
- **Backend**: Go 1.24 + Wails v2.11 (desktop app framework that binds Go to React)
- **Database**: SQLite with pure Go driver (no CGO) + GORM ORM
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

#### `pkg/database/` - SQLite + Vector Storage

Uses `database.Manager` with GORM ORM. Key models:
- `File`: Path, title, content hash, last modified, file size
- `Chunk`: Content, heading, embedding ([]float32), embedding model, timestamps
- `Tag`: Tags for file categorization with many-to-many relationship

**Repository Pattern**: `Repository` struct provides data access:
- `IndexFileWithChunks(path, content, chunks)`: Atomic file+chunks insertion with embeddings
- `FileNeedsIndexing(path, content)`: SHA256 hash comparison for incremental updates
- `SearchSimilar(embedding, limit)`: Vector similarity search
- `GetStats()`: Database statistics (file/chunk/tag counts)
- `DeleteFile(path)`: Cascade deletes chunks

Database stored at `./data/db.sqlite` relative to app directory.

#### `pkg/config/` - Configuration Management

Singleton pattern with `config.Get()`. JSON-based config stored at `%USERPROFILE%\AppData\Roaming\notebit\config.json`:
- `AIConfig`: Provider selection, OpenAI/Ollama settings, batch size
- `ChunkingConfig`: Strategy, sizes, overlap, heading preservation
- `WatcherConfig`: Enable/disable, debounce delay, worker count
- `LLMConfig`: LLM provider (OpenAI), model, temperature, max tokens
- `RAGConfig`: Max context chunks, temperature, system prompt
- `GraphConfig`: Similarity threshold, max nodes, implicit links toggle

**Defaults**: Ollama preferred (local-first), heading-based chunking, 500ms debounce.

#### `pkg/watcher/` - File System Watcher

The `watcher.Service` provides automatic file indexing:
- **Debouncing**: Configurable delay (default 500ms) to avoid duplicate processing
- **Worker Pool**: Concurrent indexing with configurable worker count (default 3)
- **Event Types**: Create, Write, Remove, Rename (via fsnotify)
- **Ignored**: `.git`, `node_modules`, `.idea`, temp files (`*.swp`, `*~`, `*.tmp`)
- **Full Index**: `IndexAll(ctx)` for background re-indexing with progress tracking

#### `pkg/knowledge/` - Semantic Search Service

The `knowledge.Service` provides semantic search capabilities:
- `IndexFileWithEmbedding(path)`: Index a single file with embeddings
- `ReindexAllWithEmbeddings()`: Full re-indexing with progress tracking
- `FindSimilar(content, limit)`: Find semantically similar notes (max 8000 chars)
- `GetSimilarityStatus()`: Check availability of semantic search

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
