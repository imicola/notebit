# Phase 0: Project Partition & Resource Inventory

**Generated**: 2026-02-11
**Status**: Complete

---

## 1. Directory Structure

```
notebit/
├── main.go                    # Wails entry point (37 lines)
├── app.go                     # App struct + Go methods (83 lines)
├── pkg/
│   └── files/
│       ├── manager.go         # File operations (300 lines)
│       └── types.go           # Type definitions (65 lines)
├── frontend/
│   ├── src/
│   │   ├── App.jsx            # Main component (329 lines)
│   │   ├── main.jsx           # Entry point (15 lines)
│   │   ├── style.css          # Global styles (124 lines)
│   │   └── components/
│   │       ├── Editor.jsx     # CodeMirror editor (243 lines)
│   │       ├── FileTree.jsx   # File navigation (82 lines)
│   │       ├── CommandPalette.jsx  # Search UI (132 lines)
│   │       ├── Toast.jsx      # Notification (51 lines)
│   │       └── SettingsModal.jsx   # Settings (165 lines)
│   ├── wailsjs/               # Auto-generated bindings
│   ├── package.json
│   └── tailwind.config.js
└── docs/
    └── notebit-prd.md         # Product requirements
```

---

## 2. Domain Division

### Frontend Domain
| Domain | Location | Purpose |
|--------|----------|---------|
| **Components** | `frontend/src/components/` | UI components |
| **Main App** | `frontend/src/App.jsx` | Application state & orchestration |
| **Styles** | `frontend/src/style.css` | Theme & global CSS |

### Backend Domain
| Domain | Location | Purpose |
|--------|----------|---------|
| **Entry Point** | `main.go` | Wails app initialization |
| **App Binding** | `app.go` | Go methods exposed to frontend |
| **File Service** | `pkg/files/` | File system operations |

### Shared Domain
| Resource | Location | Description |
|----------|----------|-------------|
| **Type Definitions** | `pkg/files/types.go` | Go structs (FileNode, NoteContent) |
| **Generated Types** | `frontend/wailsjs/go/models.ts` | TypeScript types auto-generated |

---

## 3. Resource Inventory

### 3.1 UI Layer Resources

#### Component Inventory
| Component | Location | Lines | Dependencies |
|-----------|----------|-------|--------------|
| App | `App.jsx` | 329 | lucide-react, clsx, wailsjs |
| Editor | `components/Editor.jsx` | 243 | codemirror, markdown-it, lucide-react |
| FileTree | `components/FileTree.jsx` | 82 | lucide-react, clsx |
| CommandPalette | `components/CommandPalette.jsx` | 132 | headlessui/react, fuse.js, lucide-react |
| Toast | `components/Toast.jsx` | 51 | headlessui/react, lucide-react |
| SettingsModal | `components/SettingsModal.jsx` | 165 | headlessui/react, lucide-react |

#### Icon Library
- **Source**: lucide-react
- **Usage**: FolderOpen, X, Monitor, Save, Settings, ChevronRight, ChevronDown, File, Folder, FolderOpen, Search, Command, Edit3, Eye, Split, CheckCircle, Type

#### Style System
- **Framework**: Tailwind CSS 3.4
- **Theme Variables**: CSS custom properties in `style.css`
  - `--font-interface`, `--font-text`
  - `--text-normal`, `--text-muted`, `--text-faint`, `--text-accent`
  - `--background-primary`, `--background-secondary`, etc.
  - `--color-red`, `--color-orange`, etc. (Obsidian colors)
- **Tailwind Config**: Custom colors mapped to CSS variables

### 3.2 Utility Layer Resources

#### Utility Functions
| Function | Location | Purpose |
|----------|----------|---------|
| `clsx` | External | Conditional class names |

**Note**: No custom utility functions found. All utilities come from external packages.

#### Custom Hooks
**None found** - All state management uses React's built-in hooks.

#### Type Definitions (Frontend)
| Type | Location | Description |
|------|----------|-------------|
| `files.FileNode` | `wailsjs/go/models.ts` | File tree node |
| `files.NoteContent` | `wailsjs/go/models.ts` | File content |
| `files.JSONTime` | `wailsjs/go/models.ts` | Custom time type |

### 3.3 Service Layer Resources

#### API Services (Backend)
| Method | Location | Purpose |
|--------|----------|---------|
| `OpenFolder` | `app.go:30` | Open folder dialog |
| `ListFiles` | `app.go:50` | Get file tree |
| `ReadFile` | `app.go:55` | Read file content |
| `SaveFile` | `app.go:60` | Save file content |
| `CreateFile` | `app.go:65` | Create new file |
| `DeleteFile` | `app.go:70` | Delete file |
| `RenameFile` | `app.go:75` | Rename file |
| `GetBasePath` | `app.go:80` | Get current base path |

#### Frontend API Calls
**Pattern**: Direct import from wailsjs
```javascript
import { OpenFolder, ListFiles, ReadFile, SaveFile } from '../wailsjs/go/main/App';
```

#### State Management
**Pattern**: React useState in App.jsx (no global state)
- `fileTree`, `currentFile`, `currentContent`, `basePath`
- `loading`, `error`
- `sidebarWidth`, `isResizing`, `isZenMode`
- `isCommandPaletteOpen`, `isSettingsOpen`, `toast`
- `appSettings`

### 3.4 Infrastructure Inventory

| Infrastructure | Status | Location/Notes |
|----------------|--------|----------------|
| **Logging** | ❌ Not Found | Only `console.log/error` used |
| **Event System** | ❌ Not Found | Props drilling pattern |
| **i18n** | ❌ Not Found | Hardcoded English strings |
| **Theme System** | ✅ Partial | CSS variables + localStorage |
| **Error Handling** | ⚠️ Basic | try-catch with `setError(null)` |
| **Configuration** | ⚠️ Partial | localStorage for settings |
| **Router** | ❌ Not Found | Single-page app, no routing |
| **Permission** | N/A | Desktop app |

---

## 4. Dependency Analysis

### Frontend Dependencies (package.json)

#### Production Dependencies
| Package | Version | Purpose |
|---------|---------|---------|
| `@codemirror/*` | Various | Editor core |
| `@headlessui/react` | 2.2.9 | UI primitives (Dialog, Combobox) |
| `@lezer/*` | Various | Parser infrastructure |
| `clsx` | 2.1.1 | Class name utility |
| `fuse.js` | 7.1.0 | Fuzzy search |
| `lucide-react` | 0.563.0 | Icon library |
| `markdown-it` | 14.1.1 | Markdown parser |
| `markdown-it-github-alerts` | 1.0.1 | GitHub-style alerts |
| `react` | 18.2.0 | UI framework |
| `react-dom` | 18.2.0 | React DOM |
| `react-markdown` | 10.1.0 | React markdown (unused?) |
| `remark-gfm` | 4.0.1 | GitHub Flavored Markdown |
| `tailwind-merge` | 3.4.0 | Tailwind class merge (unused?) |

#### Potentially Unused Dependencies
| Package | Reason |
|---------|--------|
| `react-markdown` | Editor uses `markdown-it` instead |
| `remark-gfm` | Paired with react-markdown |
| `tailwind-merge` | Using `clsx` instead |
| `@dnd-kit/*` | No drag-and-drop implementation found |

### Backend Dependencies
| Package | Version | Purpose |
|---------|---------|---------|
| `wailsapp/wails/v2` | v2.11 | Desktop app framework |

---

## 5. Key Observations

### Positive Patterns
1. **Clean separation**: Go backend handles file I/O, React frontend handles UI
2. **Type safety**: Go types auto-generate TypeScript bindings
3. **Theme system**: CSS variables allow runtime theme changes
4. **Component library**: Using Headless UI for accessible components

### Potential Issues
1. **No utility abstraction**: Direct API calls in App.jsx, no service layer
2. **No error boundary**: React errors will crash the app
3. **Mixed markdown parsers**: Both `markdown-it` and `react-markdown` installed
4. **Unused dependencies**: Several packages installed but not used
5. **Props drilling**: Settings passed through multiple levels
6. **No loading states abstraction**: Each handler sets loading/error manually

### Missing Infrastructure
1. No logging system (only console.log)
2. No event bus for cross-component communication
3. No i18n support (hardcoded strings)
4. No centralized error handling
5. No API service abstraction layer

---

## 6. Code Metrics

| Metric | Value |
|--------|-------|
| Total Go Files | 3 |
| Total Go Lines | ~485 |
| Total JSX Files | 7 |
| Total JSX Lines | ~1,017 |
| Total CSS Lines | 124 |
| Components | 6 |
| Backend Methods | 8 |

---

## Next Steps

Proceed to **Phase 1: Key Identification** to:
1. List core features
2. Identify high-frequency code
3. Trace key paths
