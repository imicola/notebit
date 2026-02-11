# Master Plan: Notebit Refactoring

**Created**: 2026-02-11
**Last Updated**: 2026-02-11
**Status**: In Progress
**Progress**: 80%

---

## Overview

Total Tasks: 15
- Architecture Layer: 3 tasks ✅
- Module Layer: 12 tasks (9 completed)

---

## Phase 1: Architecture Layer [Completed ✅]

### task-A001: Create Directory Structure ✅
- **Status**: ✅ Completed
- **Description**: Create proper directory structure for hooks, services, utils, constants

**Completed**:
- [x] Create `frontend/src/hooks/` directory
- [x] Create `frontend/src/services/` directory
- [x] Create `frontend/src/utils/` directory
- [x] Create `frontend/src/constants/` directory

### task-A002: Create fileService.js ✅
- **Status**: ✅ Completed
- **Description**: Abstract Wails API calls into service layer

**Completed**:
- [x] Create `services/fileService.js`
- [x] Wrap OpenFolder, ListFiles, ReadFile, SaveFile
- [x] Add consistent error handling with FileServiceError

### task-A003: Create withAsyncHandler Utility ✅
- **Status**: ✅ Completed
- **Description**: Create reusable async handler wrapper

**Completed**:
- [x] Create `utils/asyncHandler.js`
- [x] Extract setLoading, setError pattern
- [x] Create createAsyncHandlerFactory for shared state

---

## Phase 2: Module Layer - Quick Wins [Completed ✅]

### task-M010: Delete Dead Code in Editor.jsx ✅
- **Status**: ✅ Completed
- **Location**: Editor.jsx:37-57
- **Description**: Remove unused obsidianExtensionsPlugin

**Completed**:
- [x] Delete unused obsidianExtensionsPlugin code
- [x] Clean up imports (removed unused WidgetType, defaultHighlightStyle)

### task-M011: Fix Callback Closure in Editor.jsx ✅
- **Status**: ✅ Completed
- **Location**: Editor.jsx:131
- **Description**: Use ref pattern for onSave callback in keymap

**Completed**:
- [x] Add onSaveRef = useRef(onSave) for callback
- [x] Update handleSave to use onSaveRef.current
- [x] Add useEffect to keep ref updated

### task-M022: Memoize Fuse Instance in CommandPalette ✅
- **Status**: ✅ Completed
- **Location**: CommandPalette.jsx:29-32
- **Description**: Use useMemo for Fuse instance

**Completed**:
- [x] Wrap Fuse creation in useMemo
- [x] Memoize flattenFiles result
- [x] Move flattenFiles outside component

---

## Phase 3: Module Layer - Hooks Extraction [Completed ✅]

### task-M004: Extract useSettings Hook ✅
- **Status**: ✅ Completed
- **Description**: Extract settings state and logic to custom hook

**Completed**:
- [x] Create `hooks/useSettings.js`
- [x] Move settings state and effects
- [x] Move handleUpdateSettings as updateSetting/updateSettings
- [x] Add resetSettings function

### task-M005: Extract useToast Hook ✅
- **Status**: ✅ Completed
- **Description**: Extract toast state to custom hook

**Completed**:
- [x] Create `hooks/useToast.js`
- [x] Move toast state and showToast function
- [x] Add showSuccess/showError convenience methods

### task-M001: Extract useFileOperations Hook ✅
- **Status**: ✅ Completed
- **Description**: Extract file operations state and handlers

**Completed**:
- [x] Create `hooks/useFileOperations.js`
- [x] Move fileTree, currentFile, currentContent, basePath state
- [x] Move handlers using fileService for API calls
- [x] Integrate with asyncHandler utility

---

## Phase 4: Additional Improvements [Completed ✅]

### task-M026: Extract Constants ✅
- **Status**: ✅ Completed
- **Description**: Extract magic numbers and hardcoded strings

**Completed**:
- [x] Create `constants/index.js`
- [x] Move sidebar width constants (SIDEBAR)
- [x] Move localStorage keys (STORAGE_KEYS)
- [x] Move error messages (ERRORS)
- [x] Move font lists (FONTS_INTERFACE, FONTS_TEXT)
- [x] Move view mode constants (VIEW_MODES)
- [x] Move default settings (DEFAULT_SETTINGS)
- [x] Update SettingsModal to use constants

### task-M005b: Extract useResizable Hook ✅
- **Status**: ✅ Completed
- **Description**: Extract sidebar resize logic

**Completed**:
- [x] Create `hooks/useResizable.js`
- [x] Add localStorage persistence option
- [x] Clean event listener management

### task-M006b: Extract useKeyboardShortcuts Hook ✅
- **Status**: ✅ Completed
- **Description**: Extract global keyboard shortcuts

**Completed**:
- [x] Create `hooks/useKeyboardShortcuts.js`
- [x] Support Mod+key pattern
- [x] Support single key shortcuts

---

## Phase 5: Pending Tasks [Optional - P2/P3]

### task-M003: Add ErrorBoundary
- **Status**: ⏳ Pending (P2)
- **Priority**: Medium
- **Description**: Add React Error Boundary for crash recovery

### task-M019: Add Keyboard Navigation to FileTree
- **Status**: ⏳ Pending (P2)
- **Priority**: Medium
- **Description**: Make file tree keyboard accessible

### task-M013: Add Preview Sanitization
- **Status**: ⏳ Pending (P2)
- **Priority**: Medium
- **Description**: Sanitize HTML before dangerouslySetInnerHTML

### task-CleanDeps: Remove Unused Dependencies
- **Status**: ⏳ Pending (P3)
- **Description**: Remove unused packages from package.json
- **Packages**: react-markdown, remark-gfm, tailwind-merge, @dnd-kit/*

---

## Files Created/Modified

### New Files
- `frontend/src/hooks/index.js`
- `frontend/src/hooks/useSettings.js`
- `frontend/src/hooks/useToast.js`
- `frontend/src/hooks/useFileOperations.js`
- `frontend/src/hooks/useResizable.js`
- `frontend/src/hooks/useKeyboardShortcuts.js`
- `frontend/src/services/fileService.js`
- `frontend/src/utils/asyncHandler.js`
- `frontend/src/constants/index.js`

### Modified Files
- `frontend/src/components/Editor.jsx` - Removed dead code, fixed callback closure
- `frontend/src/components/CommandPalette.jsx` - Added useMemo for Fuse
- `frontend/src/components/SettingsModal.jsx` - Use constants for fonts

---

## Checkpoints

| Checkpoint | Created | Git Ref | Description |
|------------|---------|---------|-------------|
| 1 | 2026-02-11 | current | Architecture + Quick Wins + Hooks + Constants |

---

## Progress Tracking

| Phase | Tasks | Completed | Progress |
|-------|-------|-----------|----------|
| Architecture | 3 | 3 | 100% |
| Quick Wins | 3 | 3 | 100% |
| Hooks | 3 | 3 | 100% |
| Additional | 3 | 3 | 100% |
| Quality (Optional) | 3 | 0 | 0% |
| **Total** | **15** | **12** | **80%** |
