# Key Identification & Core Features
**Last Updated**: 2026-02-12

## Core Feature List

### Feature 1: File Management
- **User Story**: User can open folders, browse file tree, and select files.
- **Entry Point**: `App.jsx` (handleOpenFolder, handleFileSelect)
- **Involved Modules**: 
  - `components/FileTree.jsx`
  - `services/fileService.js` (Available but not fully integrated)
  - `hooks/useFileOperations.js` (Created but unused in App)
- **Complexity**: Medium

### Feature 2: Markdown Editing
- **User Story**: User can edit Markdown text with syntax highlighting and live preview.
- **Entry Point**: `components/Editor.jsx`
- **Involved Modules**: 
  - `CodeMirror` (library)
  - `dompurify` (sanitization)
- **Complexity**: High

### Feature 3: Global Settings
- **User Story**: User can configure fonts and themes.
- **Entry Point**: `App.jsx` (appSettings state)
- **Involved Modules**: 
  - `components/SettingsModal.jsx`
  - `hooks/useSettings.js` (Created but unused in App)
- **Complexity**: Low

### Feature 4: Layout & Navigation
- **User Story**: User can resize sidebar, toggle Zen mode, and use command palette.
- **Entry Point**: `App.jsx` (sidebarWidth, isZenMode state)
- **Involved Modules**: 
  - `components/CommandPalette.jsx`
  - `hooks/useResizable.js` (Created but unused in App)
  - `hooks/useKeyboardShortcuts.js` (Partially used?)
- **Complexity**: Medium

## High-Frequency Code
- `App.jsx` is the central hub and currently acts as a "God Object", containing logic that should be in hooks.
- `services/fileService.js` is imported but direct state management still happens in App.

## Key Path Tracing (Current Reality)
- **File Open**: App.jsx `handleOpenFolder` -> `OpenFolder` (Wails) -> `ListFiles` -> `setFileTree`
- **Save**: App.jsx `handleSave` -> `SaveFile` (Wails) -> `showToast`
- **Settings**: App.jsx `handleUpdateSettings` -> `localStorage` -> CSS Variables

**Observation**: The Refactoring artifacts (hooks, services) exist but are **orphaned**. The next phase must focus on **Integration**.
