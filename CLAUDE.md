# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Notebit** is a Local-First Markdown note-taking application for PKM enthusiasts and researchers. It combines a distraction-free writing environment ("The Sanctuary") with background AI-powered knowledge management ("The Silent Curator").

**Philosophy**: "Write for Humans, Manage by Silicon" - pure editor during writing, AI processes notes after the fact.

## Tech Stack

- **Frontend**: React 18.2 + Vite (no framework UI library yet - plain CSS)
- **Backend**: Go 1.23 + Wails v2.11 (desktop app framework that binds Go to React)
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

### Dual-Module Design

1. **Module A - The Sanctuary (P0)**: Distraction-free Markdown editor. Key constraint: **NO AI autocomplete/ghost text during typing**.

2. **Module B - The Silent Curator (P0)**: Background embedding and semantic search. Uses `fsnotify` to trigger vector generation on file save. Must have graceful degradation - if AI services are offline, sidebar shows "Service Offline" but editor works 100%.

3. **Module C - Knowledge Review (P1)**: RAG-based chat interface with citations linking back to source files.

4. **Module D - Knowledge Graph (P2)**: Force-directed graph visualization of note connections.

### Project Structure

- `main.go` - Wails app entry point, binds App struct for frontend RPC calls
- `app.go` - Core application struct and Go methods exported to frontend
- `frontend/src/` - React frontend
- `wails.json` - Wails project configuration
- `docs/` - Product requirements (notebit-prd.md is authoritative)

### Wails Go-React Binding Pattern

Go methods are automatically bound to React via Wails codegen:
1. Define method on App struct in `app.go`
2. Import from `../wailsjs/go/main/App` in React
3. Call as Promise: `Greet(name).then(result)`

Generated bindings are in `frontend/wailsjs/` - do not edit these directly.

### Key Constraints

- **Local-First**: Notes stored as plain `.md` files; embeddings in `./data/db.sqlite`
- **Performance**: Cold start < 2s, vector search < 200ms, non-blocking UI (use goroutines)
- **Privacy**: No external calls by default; Ollama preferred over cloud APIs
- **Data Integrity**: App must function 100% even if AI services crash

## Current Status

Sprint 1 completed - basic Wails + React skeleton. Next: File I/O and Markdown editor implementation. See `docs/notebit-prd.md` for full roadmap and requirements.
