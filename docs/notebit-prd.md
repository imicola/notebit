# Product Requirements Document: Notebit

**Version**: 1.0
**Date**: 2026-02-11
**Author**: Sarah (Product Owner)
**Status**: Approved
**Quality Score**: 95/100

---

## 1. Executive Summary

Notebit is a **Local-First** Markdown note-taking application designed for **PKM Enthusiasts and Researchers**. It bridges the gap between distraction-free writing and AI-driven knowledge management. Unlike tools that intrude with autocomplete, Notebit acts as a "Silent Curator," using local or cloud LLMs and embeddings to automatically connect, tag, and review knowledge *after* the writing process, ensuring user privacy and data sovereignty.

---

## 2. Problem Statement

**Current Situation**: 
- Traditional note apps are static silos; notes get lost in folders.
- Modern AI editors intrude on the creative process with "Ghost Text" and autocomplete, breaking flow.
- Cloud-based solutions compromise privacy for researchers and privacy-conscious users.

**Proposed Solution**: 
- A "Sanctuary" for writing (pure editor) paired with a "Silent Curator" (background AI) that runs locally.
- **Philosophy**: "Write for Humans, Manage by Silicon."

**Business Impact**: 
- High user retention through "Rediscovery" value (notes become more valuable over time).
- strong appeal to privacy-focused niche (Local-First).

---

## 3. Success Metrics & KPIs

**Primary KPIs**:
1.  **Knowledge Retrieval Rate**: Users successfully finding relevant past notes via Sidebar/Chat within 30 seconds.
2.  **System Latency**: App cold start < 2 seconds; Search/Vector query < 200ms.
3.  **Stability**: 100% data integrity even if AI services crash.

**Validation**: 
- User feedback loops on "Related Notes" relevance.
- Automated performance testing pipelines.

---

## 4. User Personas

### Primary: The Deep Diver (Researcher)
- **Role**: Academic or Technical Researcher.
- **Goals**: Connect current writing with citations/past findings without context switching.
- **Pain Points**: Losing track of sources; privacy concerns with cloud AI.
- **Behavior**: Writes long-form; values accuracy and references.

### Secondary: The Gardener (PKM Enthusiast)
- **Role**: Knowledge Worker / Lifelong Learner.
- **Goals**: Cultivate a "Second Brain"; revisit and refine old ideas.
- **Pain Points**: "Write and Forget" syndrome; unorganized folders.
- **Behavior**: Uses "Daily Digest" to review connections; values serendipity.

---

## 5. Functional Requirements

### Module A: The Sanctuary (Pure Editor)
**Priority: P0**
*   **User Story**: As a user, I want a distraction-free typing experience so I can focus on my thoughts.
*   **Acceptance Criteria**:
    *   Standard Markdown syntax highlighting (Gfm compatible).
    *   Real-time local file system saving.
    *   **Constraint**: NO AI autocomplete/ghost text triggers during typing.
    *   WYSIWYG / Source mode toggle (P1).

### Module B: The Silent Curator (Backend Intelligence)
**Priority: P0 (Embedding) / P1 (Sidebar)**
*   **User Story**: As a user, I want my notes to be automatically connected so I can discover related concepts.
*   **Acceptance Criteria**:
    *   **Live Embedding**: On file save (`fsnotify`), background process chunks text and generates vectors (SQLite).
    *   **Semantic Sidebar**: Shows top 5 related notes based on Cosine Similarity.
    *   **Graceful Degradation**: If Ollama/Embedding service is offline, the sidebar displays a "Service Offline" icon but the editor functions 100% normally. No error popups blocking writing.

### Module C: Knowledge Review (RAG & Chat)
**Priority: P1**
*   **User Story**: As a Researcher, I want to chat with my notes to synthesize information.
*   **Acceptance Criteria**:
    *   Chat interface separate from writing area.
    *   RAG pipeline retrieves relevant chunks before sending to LLM.
    *   **Citation**: Responses MUST link back to source files (e.g., `[[Filename]]`).

### Module D: Knowledge Graph
**Priority: P2**
*   **User Story**: As a Gardener, I want to visualize connections to see the "shape" of my knowledge.
*   **Acceptance Criteria**:
    *   Force-directed graph visualization.
    *   Auto-linking based on semantic similarity + explicit wiki-links.
    *   Interactive: Clicking a node opens the note.

---

## 6. Technical Architecture & Constraints

**Stack**:
*   **Frontend**: React + Tailwind CSS (Vite build).
*   **Backend**: Wails (Go) for system binding.
*   **Database**: SQLite (Metadata + Vector Embeddings).
*   **AI**: Ollama (Local) / OpenAI Interface.(P0)

**Non-Functional Requirements**:
*   **Performance**: Vector search must not block the UI thread (use Go goroutines).
*   **Privacy**: All embeddings stored in `./data/db.sqlite`; no external calls by default.
*   **Compatibility**: Notes stored as plain `.md` files. If app is uninstalled, files remain readable.

---

## 7. Roadmap (Phased Delivery)

### Sprint 1: The Skeleton (MVP)
*   Init Wails + React.
*   File I/O (Open folder, Read/Write .md).
*   Basic Markdown Editor (UI).

### Sprint 2: The Brain (Core AI)
*   Integrate SQLite + Vector extension.
*   Implement "Save -> Embed" pipeline.
*   Basic Semantic Search (Console output).

### Sprint 3: The Interface (UX)
*   Implement Semantic Sidebar (UI).
*   Connect Chat Interface to Ollama.
*   Add Graceful Degradation handling.
