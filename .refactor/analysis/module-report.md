# Phase 3: Module Layer Analysis Report

**Generated**: 2026-02-11
**Status**: Complete

---

## 1. Analysis Coverage

| Module | Analysis Status | Problem Count | Report |
|--------|-----------------|---------------|--------|
| App.jsx | ✅ Complete | 9 | [modules/app-module.md](modules/app-module.md) |
| Editor.jsx | ✅ Complete | 8 | [modules/editor-module.md](modules/editor-module.md) |
| FileTree.jsx | ✅ Complete | 3 | [modules/other-modules.md](modules/other-modules.md) |
| CommandPalette.jsx | ✅ Complete | 3 | [modules/other-modules.md](modules/other-modules.md) |
| Toast.jsx | ✅ Complete | 1 | [modules/other-modules.md](modules/other-modules.md) |
| SettingsModal.jsx | ✅ Complete | 3 | [modules/other-modules.md](modules/other-modules.md) |

**Total Problems Found**: 30

---

## 2. Problem Summary

### 2.1 Duplicate Implementations (P1)

| Functionality | Locations | Suggestion |
|---------------|-----------|------------|
| Error handling pattern | App.jsx (3 places) | Create `withAsyncHandler` utility |
| Loading/error state | App.jsx handlers | Consolidate in hook |

### 2.2 Resources Not Reused (P1)

| Module | Problem | Location | Should Use |
|--------|---------|----------|------------|
| App.jsx | No service layer | Direct wailsjs import | Create fileService.js |
| App.jsx | No custom hooks | useState x12 | Extract hooks |
| CommandPalette | Fuse recreation | Line 29 | useMemo |

### 2.3 Dead Code (P1)

| Location | Code | Action |
|----------|------|--------|
| Editor.jsx:37-57 | obsidianExtensionsPlugin | Delete |

### 2.4 Inconsistent Patterns (P2)

| Module | Problem | Current | Should Unify To |
|--------|---------|---------|-----------------|
| Multiple | Transition patterns | Varying | Standardize |
| Multiple | Hardcoded strings | Inline | Constants or i18n |

### 2.5 Code Quality Issues (P2)

| Module | Issue | Location |
|--------|-------|----------|
| Editor.jsx | Stale callback closure | Line 131 |
| Editor.jsx | XSS risk | Line 234 |
| FileTree.jsx | No keyboard nav | Entire component |
| App.jsx | No error boundary | Entire component |

---

## 3. Detailed Problem List

### Priority 1 (High Impact)

| ID | Problem | Module | Fix |
|----|---------|--------|-----|
| M-001 | Repeated async handler pattern | App.jsx | Create withAsyncHandler utility |
| M-002 | No service layer | App.jsx | Create fileService.js |
| M-004 | Settings logic mixed | App.jsx | Extract useSettings hook |
| M-010 | Dead code | Editor.jsx | Delete obsidianExtensionsPlugin |
| M-011 | Stale callback closure | Editor.jsx | Use ref pattern |
| M-022 | Fuse instance recreation | CommandPalette | useMemo |

### Priority 2 (Medium Impact)

| ID | Problem | Module | Fix |
|----|---------|--------|-----|
| M-003 | No error boundary | App.jsx | Add ErrorBoundary |
| M-019 | No keyboard navigation | FileTree.jsx | Add keyboard handlers |
| M-013 | XSS risk in preview | Editor.jsx | Add sanitization |

### Priority 3 (Low Impact)

| ID | Problem | Module | Fix |
|----|---------|--------|-----|
| M-005 | Resize logic embedded | App.jsx | Extract useResizable |
| M-006 | Keyboard shortcuts embedded | App.jsx | Extract useKeyboardShortcuts |
| M-007 | Magic numbers | App.jsx | Extract constants |
| M-008 | Hardcoded strings | Multiple | Extract constants |
| M-015 | Magic strings for view mode | Editor.jsx | Use enum |
| M-016 | Missing CSS classes | Editor.jsx | Define or remove |
| M-018 | Expanded state lost | FileTree.jsx | Lift state or persist |
| M-020 | Duplicate icon logic | FileTree.jsx | Extract component |
| M-021 | flattenFiles recreation | CommandPalette | useMemo |
| M-024 | Hardcoded Toast icon | Toast.jsx | Add type prop |
| M-026 | Font lists in component | SettingsModal | Extract constants |
| M-029 | Inconsistent transitions | Multiple | Standardize |
| M-030 | Hardcoded strings | Multiple | Extract for i18n |

---

## 4. Resource Reuse Summary

### Currently Using (Good)
| Resource | Used By | Status |
|----------|---------|--------|
| lucide-react | All components | ✅ Consistent |
| clsx | 5/6 components | ✅ Consistent |
| Headless UI | 4/6 components | ✅ Good |
| CSS Variables | All components | ✅ Good |
| Tailwind | All components | ✅ Good |

### Not Using (Missing)
| Resource | Should Be | Impact |
|----------|-----------|--------|
| Custom hooks | App.jsx | State management bloated |
| Service layer | App.jsx | API calls scattered |
| Constants | All | Magic values |
| Error boundary | App.jsx | Crash risk |
| Utility functions | App.jsx | Repeated patterns |

---

## 5. Refactoring Tasks

### Architecture Layer Tasks (Priority First)

| ID | Task | Effort |
|----|------|--------|
| A-001 | Create service layer abstraction | Medium |
| A-002 | Extract hooks from App.jsx | Medium |
| A-003 | Create directory structure | Low |

### Module Layer Tasks

| ID | Task | Module | Priority | Effort |
|----|------|--------|----------|--------|
| M-001 | Create withAsyncHandler utility | App.jsx | P1 | Low |
| M-002 | Create fileService.js | App.jsx | P1 | Medium |
| M-004 | Extract useSettings hook | App.jsx | P1 | Low |
| M-010 | Delete dead code | Editor.jsx | P1 | Low |
| M-011 | Fix callback closure | Editor.jsx | P1 | Low |
| M-022 | Memoize Fuse instance | CommandPalette | P1 | Low |
| M-003 | Add ErrorBoundary | App.jsx | P2 | Low |
| M-019 | Add keyboard navigation | FileTree.jsx | P2 | Medium |
| M-013 | Add preview sanitization | Editor.jsx | P2 | Medium |

---

## 6. Recommended Execution Order

```
Phase 4.1: Foundation (Architecture Layer)
├── Create directory structure
├── Create fileService.js
└── Extract useSettings hook

Phase 4.2: Cleanup (Quick Wins)
├── Delete dead code in Editor.jsx
├── Fix callback closure in Editor.jsx
├── Memoize Fuse in CommandPalette
└── Create withAsyncHandler utility

Phase 4.3: Quality (Medium Effort)
├── Extract remaining hooks from App.jsx
├── Add ErrorBoundary
├── Add keyboard navigation to FileTree
└── Add preview sanitization

Phase 4.4: Polish (Low Priority)
├── Extract constants
├── Standardize transitions
└── Prepare for i18n
```

---

## 7. Metrics Summary

| Metric | Current | Target |
|--------|---------|--------|
| App.jsx lines | 329 | ~100 |
| Dead code lines | 20 | 0 |
| useState in App.jsx | 12 | 4-5 |
| Repeated patterns | 4 | 0 |
| Missing abstractions | 5 | 0 |

---

## Next Steps

Proceed to **Phase 4: Execute Refactoring** with:
1. Create `.refactor/tasks/master-plan.md`
2. Begin with architecture layer tasks
3. Execute module layer tasks by priority
