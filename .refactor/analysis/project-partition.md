# Project Partition & Resource Inventory (Full Repository)
**Last Updated**: 2026-02-12
**Last Updated**: 2026-02-13
## 1. Directory Structure

## Domain Division

### Frontend Domain (`frontend/src`)
│   ├── AISettings.jsx   # AI Configuration UI
│   ├── CommandPalette.jsx # Command palette & search
│   ├── Editor.jsx       # Main Markdown editor (CodeMirror)
│   ├── ErrorBoundary.jsx # React Error Boundary
│   ├── FileTree.jsx     # Sidebar file navigation
│   ├── SettingsModal.jsx # General settings
│   └── Toast.jsx        # Notification display
├── constants/           # Global constants
│   └── index.js         # Unified constants export
├── hooks/               # Custom React Hooks
│   ├── index.js         # Unified hooks export
│   ├── useFileOperations.js # File system logic
│   ├── useKeyboardShortcuts.js # Keyboard handling
│   ├── useResizable.js  # Sidebar resizing
│   ├── useSettings.js   # Settings management
│   └── useToast.js      # Notification state
├── services/            # API Services
│   └── fileService.js   # Wails API wrapper
├── utils/               # Utility functions
│   └── asyncHandler.js  # Async error handling
├── App.jsx              # Layout & routing
└── main.jsx             # Entry point
- **Wails Binding Layer**: `app.go` (frontend-facing methods + orchestration)
- **Core Services**: `pkg/files`, `pkg/database`, `pkg/ai`, `pkg/knowledge`, `pkg/watcher`, `pkg/rag`, `pkg/graph`
## 2. Resource Inventory

### UI Layer (Components)
| Component | Location | Responsibility |
|-----------|----------|----------------|
| Editor | components/Editor.jsx | Markdown editing, live preview, sanitization |
| FileTree | components/FileTree.jsx | File navigation, recursive display |
| CommandPalette | components/CommandPalette.jsx | Global search, command execution |
| SettingsModal | components/SettingsModal.jsx | App configuration (Fonts, Theme) |
| AISettings | components/AISettings.jsx | AI provider config |
| Toast | components/Toast.jsx | Notification UI |
| ErrorBoundary | components/ErrorBoundary.jsx | Crash recovery UI |
| Core Shell | `frontend/src/App.jsx` | `App` |
### Logic Layer (Hooks)
| Hook | Location | Responsibility |
|------|----------|----------------|
| useFileOperations | hooks/useFileOperations.js | Open/Read/Save files |
| useSettings | hooks/useSettings.js | Load/Save user preferences |
| useToast | hooks/useToast.js | Show/Hide notifications |
| useResizable | hooks/useResizable.js | Sidebar resize logic |
| useKeyboardShortcuts | hooks/useKeyboardShortcuts.js | Global shortcut management |

### Service Layer
| Service | Location | Responsibility |
|---------|----------|----------------|
| fileService | services/fileService.js | Interface with Wails backend (Go) |
| Settings hook | `frontend/src/hooks/useSettings.js` | `useSettings` |
### Utilities
| Utility | Location | Responsibility |
|---------|----------|----------------|
| asyncHandler | utils/asyncHandler.js | Standardized async/await error handling |

### Constants
| Constant | Location | Contents |
|----------|----------|----------|
| constants | constants/index.js | SIDEBAR, STORAGE_KEYS, ERRORS, FONTS, DEFAULT_SETTINGS |

## 3. Infrastructure
- **Build System**: Vite
- **Styling**: Tailwind CSS + CSS Modules (Editor.css, FileTree.css)
- **State Management**: React Context / Local State + Custom Hooks
- **Backend Bridge**: Wails runtime (window.go.main.App)
- **Sanitization**: DOMPurify (in Editor.jsx)
| File manager | `pkg/files/manager.go` | File listing, read/write, path safety |
| DB manager/repository | `pkg/database/*` | Persistence, vectors, migrations |
| AI service | `pkg/ai/*` | Embedding and LLM provider abstraction |
| Knowledge service | `pkg/knowledge/service.go` | High-level note indexing orchestration |
| Watcher service | `pkg/watcher/service.go` | FS watch and incremental indexing |
| Logger infra | `pkg/logger/*` | Async logging + metrics + sinks |

### Infrastructure & Hygiene Findings
| Type | Location | Finding |
|------|----------|---------|
| Empty directories | `frontend/src/components/Editor|FileTree|Layout|Preview` | Leftover structural artifacts |
| Generated + manual split | `frontend/wailsjs` + direct frontend imports | API boundary partially bypassed |
| Temp folders | root `tmpclaude-*` | Candidate cleanup after validation |

## Complexity Hotspots (LoC)
- `app.go`: 786
- `frontend/src/components/AISettings.jsx`: 647
- `pkg/logger/logger.go`: 517
- `pkg/config/config.go`: 508
- `frontend/src/App.jsx`: 464

## Partition Conclusion
Current architecture intent is good (hooks/service/backend service split exists), but integration is incomplete and boundaries are inconsistent. Refactoring should prioritize boundary enforcement and decomposition of top hotspot files.
