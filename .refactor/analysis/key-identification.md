# Key Identification & Core Features
**Last Updated**: 2026-02-13

## Core Feature List

### Feature 1: Workspace & File Operations
- **User Story**: 用户打开文件夹、浏览文件树、读写 Markdown 文件。
- **Frontend Entry**: `frontend/src/App.jsx` (`handleOpenFolder`, `handleFileSelect`, `handleSave`)
- **Backend Entry**: `app.go` (`OpenFolder`, `ListFiles`, `ReadFile`, `SaveFile`)
- **Involved Modules**: `pkg/files`, `pkg/database`, `frontend/src/services/fileService.js`, `frontend/src/hooks/useFileOperations.js`
- **Complexity**: High

### Feature 2: AI Retrieval & Similarity
- **User Story**: 用户发起问答、获取相似笔记、查看图谱。
- **Frontend Entry**: `ChatPanel.jsx`, `SimilarNotesSidebar.jsx`, `GraphPanel.jsx`
- **Backend Entry**: `app.go` (`RAGQuery`, `FindSimilar`, `GetGraphData`)
- **Involved Modules**: `pkg/ai`, `pkg/rag`, `pkg/graph`, `pkg/knowledge`, `pkg/database`
- **Complexity**: High

### Feature 3: Settings & UI Preferences
- **User Story**: 用户调整字体与界面行为，配置可持久化。
- **Entry Point**: `frontend/src/App.jsx`, `frontend/src/components/SettingsModal.jsx`
- **Involved Modules**: `frontend/src/hooks/useSettings.js`, `frontend/src/constants/index.js`
- **Complexity**: Medium

### Feature 4: File Watching & Incremental Indexing
- **User Story**: 目录变化自动触发索引更新。
- **Entry Point**: `app.go` (`startWatcher`, `runFullIndex`)
- **Involved Modules**: `pkg/watcher`, `pkg/knowledge`, `pkg/database`, `pkg/ai`
- **Complexity**: High

### Feature 5: Logging & Runtime Observability
- **User Story**: 关键操作可观测、可排障、可追踪。
- **Entry Point**: `main.go` logger init + pervasive logger calls
- **Involved Modules**: `pkg/logger/*`, `pkg/database/manager.go`, `app.go`
- **Complexity**: Medium

## High-Frequency / High-Risk Hotspots
1. `app.go` (786 lines): 绑定层 + 编排 + 部分业务逻辑混合。
2. `frontend/src/App.jsx` (464 lines): UI 容器与业务逻辑耦合。
3. `frontend/src/components/AISettings.jsx` (647 lines): 大体量配置组件。
4. `pkg/logger/logger.go` (517 lines): 基础设施核心，变更需谨慎。

## Key Path Tracing (Current Reality)

### Path A: 文件打开与保存
`frontend/src/App.jsx` (direct Wails call)
→ `app.go` Wails binding
→ `pkg/files/manager.go`
→ `pkg/database` (save 后触发索引)

### Path B: 聊天问答
`frontend/src/components/ChatPanel.jsx` (direct Wails call)
→ `app.go:RAGQuery`
→ `pkg/rag/service.go`
→ `pkg/ai` + `pkg/database`

### Path C: 自动索引
`app.go:startWatcher`
→ `pkg/watcher/service.go`
→ `pkg/knowledge/service.go`
→ `pkg/database/repository.go`

## Identification Conclusion
- 前端已有 hooks/service 资源，但主流程未充分接入（存在“孤儿化资产”）。
- 后端分层总体存在，但 Wails `App` 仍过重。
- 重构应以“边界一致性 + 热点文件拆分 + 质量安全线（context/error）”为主线推进。
