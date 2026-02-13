# Architecture Layer Analysis Report
**Last Updated**: 2026-02-13

## Layering Violations

| Violation Location | Wrong Dependency / Pattern | Severity | Fix Suggestion |
|--------------------|----------------------------|----------|----------------|
| `frontend/src/App.jsx` | UI root directly imports Wails APIs (`OpenFolder`, `ListFiles`, `ReadFile`, `SaveFile`, `SetFolder`) | High | Route through `services/fileService.js` + hooks |
| `frontend/src/components/ChatPanel.jsx` | Component directly calls Wails `RAGQuery` | High | Introduce `services/aiService.js` and keep component presentation-focused |
| `frontend/src/components/GraphPanel.jsx` | Component directly calls Wails `GetGraphData` | High | Introduce graph service adapter |
| `frontend/src/components/SimilarNotesSidebar.jsx` | Component directly calls `FindSimilar` / `GetSimilarityStatus` | High | Move backend invocation to service layer |
| `frontend/src/components/AISettings.jsx` | Component directly owns persistence/API details | Medium | Split into settings service + view model hook |
| `pkg/database/manager.go` | Logger APIs receive `nil` context | High | Use `context.TODO()` or propagated context |

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
