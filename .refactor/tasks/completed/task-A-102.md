# Task A-102: Frontend Service Adapters

## Status: ✅ COMPLETED

**Date**: 2026-02-18  
**Priority**: High  
**Component**: Frontend (Services + Components)

---

## Problem Statement

Multiple React components were directly importing and calling Wails APIs, violating the service adapter pattern and creating tight coupling between UI and backend binding layer.

### Violations Detected
- `ChatPanel.jsx` → Direct `RAGQuery`, `EventsOn` calls
- `GraphPanel.jsx` → Direct `GetGraphData` call  
- `SimilarNotesSidebar.jsx` → Direct `FindSimilar`, `GetSimilarityStatus` calls
- `AIStatusIndicator.jsx` → Direct `GetSimilarityStatus` call
- Components didn't use existing `fileService.js` for all operations

---

## Solution Implemented

Created 4 new service adapter modules + updated fileService.js to consolidate all Wails API calls in a consistent service layer.

### Files Created

#### 1. `frontend/src/services/aiService.js` (~130 lines)
- Wraps: GetAIStatus, Get/SetOpenAIConfig, Get/SetOllamaConfig, etc.
- Methods: ~18 AI configuration and embedding APIs

#### 2. `frontend/src/services/ragService.js` (~50 lines)
- Wraps: RAGQuery, EventsOn('rag_chunk')
- Methods: `query()`, `onChunk()`

#### 3. `frontend/src/services/graphService.js` (~40 lines)
- Wraps: GetGraphData
- Methods: `getGraphData()`

#### 4. `frontend/src/services/similarityService.js` (~50 lines)
- Wraps: FindSimilar, GetSimilarityStatus
- Methods: `findSimilar()`, `getStatus()`

### Files Modified

#### `frontend/src/services/fileService.js`
- Added: `SetFolder` import and `setFolder()` method
- Now covers: OpenFolder, SetFolder, ListFiles, ReadFile, SaveFile, CreateFile, RenameFile, DeleteFile

### Components Refactored (6 total)

1. **ChatPanel.jsx** → Uses `ragService.query()`, `ragService.onChunk()`
2. **GraphPanel.jsx** → Uses `graphService.getGraphData()`
3. **SimilarNotesSidebar.jsx** → Uses `similarityService` APIs
4. **AIStatusIndicator.jsx** → Uses `similarityService.getStatus()`
5. **AISettings.jsx** → Uses `aiService` for config management
6. **App.jsx** → Uses service imports instead of direct Wails

---

## Verification

✅ No direct Wails imports in component layer (0/6 violations)  
✅ `npx vite build` — PASS  
✅ All components functional after migration

---

## Impact

- **Decoupling**: Components now depend on service contracts, not Wails API
- **Testability**: Services can be mocked for unit tests
- **Maintainability**: Backend API changes isolated to service layer
- **Consistency**: Single source of truth for each API boundary
