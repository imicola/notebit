# Project Partition & Resource Inventory
**Last Updated**: 2026-02-12

## 1. Directory Structure

```
frontend/src/
├── components/          # UI Components
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
```

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
