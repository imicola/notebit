# Phase 2: Architecture Layer Analysis Report

**Generated**: 2026-02-11
**Status**: Complete

---

## 1. Layering Structure Check

### Expected Layer Structure (Desktop App with Wails)

```
┌─────────────────────────────────────────────────────────┐
│                    UI Layer                              │
│  React Components (App, Editor, FileTree, etc.)         │
├─────────────────────────────────────────────────────────┤
│                  Application Layer                       │
│  State Management, Business Logic, Event Handling       │
├─────────────────────────────────────────────────────────┤
│                  Infrastructure Layer                    │
│  API Bindings (wailsjs), Storage, Utilities             │
├─────────────────────────────────────────────────────────┤
│                    Core Layer                            │
│  Types, Constants, Domain Models                        │
└─────────────────────────────────────────────────────────┘
          │
          ▼ Wails Bridge
┌─────────────────────────────────────────────────────────┐
│                    Backend (Go)                          │
│  app.go → pkg/files/manager.go → File System            │
└─────────────────────────────────────────────────────────┘
```

### Current Structure Analysis

| Layer | Expected Location | Actual Location | Status |
|-------|-------------------|-----------------|--------|
| UI | components/ | components/ | ✅ Correct |
| Application | App.jsx or separate | Mixed in App.jsx | ⚠️ Mixed |
| Infrastructure | services/, hooks/ | Direct wailsjs imports | ❌ Missing |
| Core | types/, constants/ | wailsjs/go/models.ts only | ⚠️ Generated only |

### Layering Violations

#### Violation A-001: Missing Service Layer
- **Severity**: Medium
- **Location**: `App.jsx:3` - Direct wailsjs import
- **Problem**: Components directly import and call backend API
- **Impact**: No abstraction, hard to test, hard to add error handling/logging

```javascript
// Current (violation)
import { OpenFolder, ListFiles, ReadFile, SaveFile } from '../wailsjs/go/main/App';

// Expected
import { fileService } from '../services/fileService';
```

#### Violation A-002: Business Logic in UI Component
- **Severity**: Medium
- **Location**: `App.jsx` - All state and logic in one component
- **Problem**: App.jsx is 329 lines mixing UI, state, and business logic
- **Impact**: Hard to maintain, no separation of concerns

---

## 2. Circular Dependency Detection

### Frontend Module Graph

```
App.jsx
  ├── imports → FileTree.jsx
  ├── imports → Editor.jsx
  ├── imports → CommandPalette.jsx
  ├── imports → Toast.jsx
  ├── imports → SettingsModal.jsx
  └── imports → wailsjs/go/main/App
```

**Result**: ✅ No circular dependencies detected

All dependencies flow downward from App.jsx to child components. No child imports parent.

### Backend Module Graph

```
main.go
  └── imports → app.go
        └── imports → pkg/files
              └── standard library only
```

**Result**: ✅ No circular dependencies detected

---

## 3. Directory Structure Evaluation

### Current Structure
```
frontend/src/
├── App.jsx              # Main component (329 lines)
├── main.jsx             # Entry point
├── style.css            # Global styles
└── components/
    ├── Editor.jsx
    ├── FileTree.jsx
    ├── CommandPalette.jsx
    ├── Toast.jsx
    └── SettingsModal.jsx
```

### Issues Found

#### Issue A-003: Missing Directory Structure
| Missing Directory | Purpose |
|-------------------|---------|
| `hooks/` | Custom React hooks |
| `services/` | API service abstraction |
| `utils/` | Utility functions |
| `types/` | TypeScript type definitions |
| `constants/` | Application constants |
| `context/` | React context providers |

#### Issue A-004: No TypeScript
- **Current**: All files are `.jsx` (JavaScript)
- **PRD Requirement**: TypeScript mentioned in stack
- **Impact**: No compile-time type safety

#### Issue A-005: Mixed Concerns in App.jsx
- **Lines**: 329 lines in single file
- **Contains**:
  - Component rendering
  - State management (12 useState calls)
  - Event handlers
  - Settings persistence
  - Keyboard shortcuts
  - Resize logic
  - Toast management

---

## 4. Architecture Smells

### Smell A-006: Props Drilling
- **Location**: `App.jsx` → `Editor.jsx`
- **Problem**: Multiple props passed down (content, onChange, onSave, filename, isZenMode)
- **Severity**: Low (only 1 level deep)

### Smell A-007: Duplicate Error Handling Pattern
- **Location**: `App.jsx:65-79`, `App.jsx:92-107`, `App.jsx:109-124`
- **Problem**: Same try-catch-finally pattern repeated 3+ times
- **Severity**: Medium

```javascript
// Repeated pattern
try {
  // operation
} catch (err) {
  setError(err.message || 'Default message');
  console.error('Error:', err);
} finally {
  setLoading(false);
}
```

### Smell A-008: Hardcoded Strings
- **Location**: Throughout codebase
- **Examples**:
  - `"notebit-settings"` (localStorage key)
  - `"Select Notes Folder"` (dialog title)
  - `"Search files or commands..."` (placeholder)
- **Severity**: Low (no i18n requirement yet)

### Smell A-009: Unused Dependencies
- **Location**: `package.json`
- **Packages**:
  - `react-markdown` - Editor uses markdown-it
  - `remark-gfm` - Paired with react-markdown
  - `tailwind-merge` - Using clsx instead
  - `@dnd-kit/*` - No drag-drop implementation
- **Severity**: Low (bloat only)

---

## 5. Architecture Layer Refactoring Tasks

### Priority 1: Foundation
| ID | Task | Severity | Effort |
|----|------|----------|--------|
| A-001 | Create service layer abstraction | Medium | Medium |
| A-002 | Extract hooks from App.jsx | Medium | Medium |
| A-003 | Create proper directory structure | Low | Low |

### Priority 2: Code Quality
| ID | Task | Severity | Effort |
|----|------|----------|--------|
| A-007 | Create error handling utility | Medium | Low |
| A-009 | Remove unused dependencies | Low | Low |

### Priority 3: Future Considerations
| ID | Task | Severity | Effort |
|----|------|----------|--------|
| A-004 | Migrate to TypeScript | Medium | High |
| A-008 | Extract constants/strings | Low | Low |

---

## 6. Recommended Target Architecture

```
frontend/src/
├── main.jsx                    # Entry point
├── App.jsx                     # Root component (simplified)
├── style.css                   # Global styles
│
├── components/                 # UI Components
│   ├── layout/
│   │   ├── Header.jsx
│   │   ├── Sidebar.jsx
│   │   └── MainContent.jsx
│   ├── editor/
│   │   ├── Editor.jsx
│   │   └── Preview.jsx
│   ├── file-tree/
│   │   ├── FileTree.jsx
│   │   └── FileTreeNode.jsx
│   └── ui/                     # Reusable UI components
│       ├── Toast.jsx
│       ├── Modal.jsx
│       └── Button.jsx
│
├── hooks/                      # Custom Hooks
│   ├── useFileOperations.js
│   ├── useSettings.js
│   └── useToast.js
│
├── services/                   # API Abstraction
│   └── fileService.js
│
├── context/                    # React Context
│   ├── AppContext.jsx
│   └── SettingsContext.jsx
│
├── utils/                      # Utility Functions
│   ├── errorHandling.js
│   └── localStorage.js
│
├── constants/                  # Constants
│   └── index.js
│
└── types/                      # TypeScript (future)
    └── index.d.ts
```

---

## Next Steps

Proceed to **Phase 3: Module Layer Analysis** to:
1. Deep dive into each key module
2. Identify Vibe Coding problems
3. Detect duplicate implementations
4. Check resource reuse
