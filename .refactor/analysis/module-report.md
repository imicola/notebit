# Module Layer Analysis Report
**Last Updated**: 2026-02-13

## Analysis Coverage
| Feature/Module | Analysis Status | Problem Count | Report Scope |
|----------------|-----------------|---------------|--------------|
| Frontend App Shell (`App.jsx`) | ✅ Complete | 8 | orchestration, state, API boundary |
| Frontend AI Panels (`ChatPanel`, `GraphPanel`, `SimilarNotesSidebar`) | ✅ Complete | 6 | Wails coupling, event flow, error pattern |
| Frontend Settings (`AISettings`, `SettingsModal`, hooks) | ✅ Complete | 6 | decomposition, consistency |
| Backend Binding (`app.go`) | ✅ Complete | 5 | oversized orchestration |
| Backend Infrastructure (`pkg/database`, `pkg/logger`) | ✅ Complete | 4 | context safety, contract hygiene |

## Problem Summary

### Resources Not Reused (P1)
| Module | Problem | Location | Should Use |
|--------|---------|----------|------------|
| App shell | Existing hooks not consumed | `frontend/src/App.jsx` | `useFileOperations`, `useSettings`, `useToast`, `useResizable`, `useKeyboardShortcuts` |
| AI panels | Direct Wails API in UI | `ChatPanel.jsx`, `GraphPanel.jsx`, `SimilarNotesSidebar.jsx` | `services/*` adapters |

### Duplicate Implementations (P1)
| Functionality | Duplicate Locations | Suggestion |
|---------------|---------------------|------------|
| File open/save flow | `App.jsx` + `useFileOperations.js` | Keep hook as single path; remove App-local copies |
| Settings persistence + CSS var apply | `App.jsx` + `useSettings.js` | Keep `useSettings` only |
| Resize event binding | `App.jsx` + `useResizable.js` | Keep hook-only listener management |

### Inconsistent Patterns (P2)
| Module | Problem | Current | Should Unify To |
|--------|---------|---------|-----------------|
| Frontend API calls | Invocation style inconsistent | direct Wails import in components | service adapter pattern |
| Error reporting | Mixed `console.error` usage | ad-hoc per component | centralized error/toast + logger bridge |
| Cross-component open-file event | custom window event | `window.dispatchEvent` + global listener | typed event utility or shared state/action |

### Code Quality Findings (P2/P3)
| Finding | Location | Severity | Note |
|--------|----------|----------|------|
| Logger called with `nil` context | `pkg/database/manager.go` | High | flagged by diagnostics |
| Oversized file | `app.go` (786 LoC) | Medium | split by domain use-cases |
| Oversized component | `AISettings.jsx` (647 LoC) | Medium | split tabs/panels and hooks |
| Empty structural directories | `frontend/src/components/*` subdirs | Low | cleanup or materialize structure |

## Module Layer Refactoring Tasks
1. **[M-101] App Hook Integration**: migrate App shell to existing hooks and remove duplicate logic.
2. **[M-102] Frontend AI Service Adapters**: create service modules for RAG/similarity/graph and migrate components.
3. **[M-103] App Event Flow Refactor**: replace global `open-file` custom event with explicit shared action path.
4. **[M-104] AISettings Decomposition**: split large settings component into presentational sections + hook.
5. **[M-105] Backend App Decomposition**: split `app.go` methods by domain coordinator.
6. **[M-106] Logger Context Compliance**: enforce non-nil context in database and nearby modules.
7. **[M-107] Cleanup Pass**: remove empty/temporary directories and dead imports after refactor.

## Suggested Execution Priority
- **P0 (Safety)**: `M-106`
- **P1 (Boundary & Duplication)**: `M-101`, `M-102`
- **P2 (Scalability)**: `M-105`, `M-104`, `M-103`
- **P3 (Hygiene)**: `M-107`
