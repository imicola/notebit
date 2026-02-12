# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Notebit** is a Local-First Markdown note-taking application for PKM enthusiasts and researchers. It combines a distraction-free writing environment ("The Sanctuary") with background AI-powered knowledge management ("The Silent Curator").

**Philosophy**: "Write for Humans, Manage by Silicon" - pure editor during writing, AI processes notes after the fact.

## Tech Stack

- **Frontend**: React 18.2 + Vite 3.0, CodeMirror 6 (editor), Tailwind CSS
- **Backend**: Go 1.23 + Wails v2.11 (desktop app framework that binds Go to React)
- **Styling**: Custom Obsidian-inspired theme system with CSS variables
- **Planned**: SQLite for metadata/vector embeddings, Ollama (local) or OpenAI for AI features

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

### Go Backend (`pkg/files/`)

The `files.Manager` struct handles all file I/O operations:
- **SetBasePath**: Sets the working directory (persisted to localStorage)
- **ListFiles**: Returns recursive `FileNode` tree structure (sorted: dirs first, then alphabetically)
- **ReadFile/SaveFile/CreateFile/DeleteFile/RenameFile**: CRUD operations on markdown files
- Thread-safe with `sync.RWMutex` for concurrent access

All file paths are relative to `basePath`. The manager converts to absolute paths internally and uses `FileSystemError` for error wrapping.

### React Frontend Structure

- **`App.jsx`**: Main container, manages:
  - File tree state and current file selection
  - Sidebar width (resizable) and collapse state
  - Zen mode (F11), command palette (Cmd+K), settings modal
  - Folder path persistence (`localStorage`)

- **`components/Editor.jsx`**: CodeMirror 6-based editor with:
  - Three view modes: edit, split, preview
  - Synchronized scrolling between editor and preview
  - Custom syntax highlighting (Obsidian-style: `==highlight==`, `[[wiki]]`)
  - Save indicator (unsaved changes)
  - Markdown rendering via `markdown-it` with GitHub alerts

- **`components/FileTree.jsx`**: Recursive tree component with expand/collapse

- **`components/CommandPalette.jsx`**: Fuzzy search over files (Fuse.js) + command execution

- **`components/SettingsModal.jsx`**: Font customization (interface + text fonts)

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

Generated bindings are in `frontend/wailsjs/` - do not edit these directly.

### Key Constraints

- **Local-First**: Notes stored as plain `.md` files; embeddings planned for `./data/db.sqlite`
- **Performance**: Cold start < 2s, vector search < 200ms, non-blocking UI (use goroutines)
- **Privacy**: No external calls by default; Ollama preferred over cloud APIs
- **Data Integrity**: App must function 100% even if AI services crash
- **NO AI autocomplete**: The editor should never show ghost text/inline suggestions during typing

## Module Status

- **Module A - The Sanctuary**: Distraction-free Markdown editor (mostly complete)
- **Module B - The Silent Curator**: Background embedding and semantic search (not implemented)
- **Module C - Knowledge Review**: RAG-based chat interface (not implemented)
- **Module D - Knowledge Graph**: Graph visualization (not implemented)

## Current Implementation Status

Completed features:
- File tree with nested folders and markdown file filtering
- Markdown editor with syntax highlighting (CodeMirror 6)
- Split view with synchronized scroll
- Obsidian-style syntax (`==highlight==`, `[[wiki]]`)
- Zen mode (F11), command palette (Cmd+K)
- Resizable sidebar, collapsible
- Folder persistence (last opened folder restored on startup)
- Custom fonts (interface + text)
- Toast notifications for save confirmation
