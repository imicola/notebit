# Architecture Layer Analysis Report
**Last Updated**: 2026-02-18
**Status**: RESOLVED ✅

## Layering Violations — ALL RESOLVED

| Violation Location | Problem | Severity | Fix Applied | Status |
|--------------------|---------|----------|-------------|--------|
| `frontend/src/App.jsx` | UI root directly imported Wails APIs | High | Integrated 5 hooks (useFileOperations, useSettings, useToast, useResizable, useKeyboardShortcuts) | ✅ Fixed |
| `frontend/src/components/ChatPanel.jsx` | Direct Wails RAGQuery call | High | Routes through `services/ragService.js` | ✅ Fixed |
| `frontend/src/components/GraphPanel.jsx` | Direct GetGraphData call | High | Routes through `services/graphService.js` | ✅ Fixed |
| `frontend/src/components/SimilarNotesSidebar.jsx` | Direct FindSimilar/GetSimilarityStatus calls | High | Routes through `services/similarityService.js` | ✅ Fixed |
| `frontend/src/components/AISettings.jsx` | Monolithic 675-line component | High | Decomposed into useAISettings hook + 4 Tab sub-components | ✅ Fixed |
| `pkg/database/manager.go` | Logger APIs receive nil context | High | All nil context args → context.TODO() | ✅ Fixed |

## Circular Dependencies
- No explicit circular dependency detected from current scan.
- Risk remains in `app.go` due to broad orchestration and cross-service touchpoints.

## Directory Structure Issues

| Issue | Location | Suggestion |
|-------|----------|------------|
| Empty placeholder directories | `frontend/src/components/Editor`, `FileTree`, `Layout`, `Preview` | Remove or populate with actual split modules |
| Temporary workspace residues | root `tmpclaude-*` directories | Clean after confirming no runtime reliance |
| Hybrid API access style | service layer exists but bypassed in components | Enforce service-only Wails access guideline |

## Architecture Risk Assessment
1. **Boundary Drift**: direct Wails calls scattered in UI components reduce testability and consistency.
2. **God Object Risk**: `app.go` and `App.jsx` both exceed practical orchestration scope.
3. **Operational Safety Risk**: context misuse in logger calls weakens observability contracts.

## Architecture Layer Refactoring Tasks
1. **[A-101] Logger Context Safety**: replace `nil` contexts in `pkg/database/manager.go`.
2. **[A-102] Frontend API Boundary Enforcement**: migrate all direct component-level Wails imports into service adapters.
3. **[A-103] App Shell Slimming**: convert `frontend/src/App.jsx` to composition shell using existing hooks.
4. **[A-104] Backend Binding Slimming**: split `app.go` orchestration into domain-specific coordinators.
5. **[A-105] Structure Hygiene**: clean empty dirs and temp residues; keep folder responsibilities explicit.
