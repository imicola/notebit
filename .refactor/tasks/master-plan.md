# Master Plan: Notebit Refactoring
**Last Updated**: 2026-02-12
**Status**: In Progress
**Progress**: 85%

---

## Overview
Previous phases focused on **Creation** of the new architecture.
The current phase focuses on **Integration** of these artifacts into the main application.

---

## Phase 1-5: Creation & Preparation [Completed ✅]
(See previous logs for details. All hooks, services, and utils have been created.)

---

## Phase 6: Integration (The Missing Link) [Priority: High]

### task-M030: Integrate useToast into App.jsx
- **Status**: Pending
- **Description**: Replace local toast state in App.jsx with `useToast` hook.
- **Steps**:
  - [ ] Import `useToast` from `./hooks`
  - [ ] Replace `const [toast, setToast]` with `const { toast, showToast, hideToast } = useToast()`
  - [ ] Update `Toast` component props
  - [ ] Verify toast notifications still work

### task-M031: Integrate useSettings into App.jsx
- **Status**: Pending
- **Description**: Replace local settings state and effects with `useSettings` hook.
- **Steps**:
  - [ ] Import `useSettings`
  - [ ] Replace `const [appSettings, setAppSettings]` and `useEffect` with `const { settings, updateSetting } = useSettings()`
  - [ ] Update `SettingsModal` props
  - [ ] Verify font application and persistence

### task-M032: Integrate useFileOperations into App.jsx
- **Status**: Pending
- **Description**: Replace file handling logic with `useFileOperations`.
- **Steps**:
  - [ ] Import `useFileOperations`
  - [ ] Remove `fileTree`, `currentFile`, `currentContent`, `basePath` state
  - [ ] Remove `handleOpenFolder`, `handleFileSelect`, `handleSave`, `refreshFileTree`
  - [ ] Use hook return values: `{ fileTree, currentFile, openFolder, selectFile, saveFile }`
  - [ ] Wire up `FileTree`, `Editor`, and `CommandPalette` to hook methods
  - [ ] Verify file opening, saving, and tree navigation

### task-M033: Integrate useResizable into App.jsx
- **Status**: Pending
- **Description**: Replace sidebar resize logic.
- **Steps**:
  - [ ] Import `useResizable`
  - [ ] Replace `startResizing`, `stopResizing`, `resize` and `useEffect` with `const { width, startResizing } = useResizable()`
  - [ ] Update sidebar style and resizer handle

### task-M034: Integrate useKeyboardShortcuts into App.jsx
- **Status**: Pending
- **Description**: Centralize keyboard shortcuts.
- **Steps**:
  - [ ] Import `useKeyboardShortcuts`
  - [ ] Move `F11` (Zen Mode) and `Cmd+K` (Palette) to the hook config or verify they are already there
  - [ ] Remove `useEffect` listener in App.jsx

### task-M035: Final Cleanup of App.jsx
- **Status**: Pending
- **Description**: Remove unused imports and dead code.
- **Steps**:
  - [ ] Remove `Wails` imports (OpenFolder, etc.)
  - [ ] Remove unused `useState`, `useEffect`, `useCallback`
  - [ ] Remove unused icons or constants
  - [ ] Ensure `App.jsx` is under 150 lines (currently ~350)

---

## Future Improvements
- [ ] Add unit tests for hooks
- [ ] Add drag-and-drop support for files
