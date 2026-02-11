# Phase 1: Key Identification

**Generated**: 2026-02-11
**Status**: Complete

---

## 1. Core Feature List

### Feature 1: Folder Management
- **User Story**: User can open a folder to browse and manage markdown files
- **Entry Point**: `App.jsx:65` - `handleOpenFolder()`
- **Backend**: `app.go:30` - `OpenFolder()`, `pkg/files/manager.go:24` - `SetBasePath()`
- **Involved Modules**: App.jsx, wailsjs, app.go, files/manager.go
- **Complexity**: Low

### Feature 2: File Tree Navigation
- **User Story**: User can browse folder structure and select files to edit
- **Entry Point**: `FileTree.jsx:5` - `FileTreeNode`
- **Involved Modules**: FileTree.jsx, App.jsx (state)
- **Complexity**: Low

### Feature 3: Markdown Editor
- **User Story**: User can write and edit markdown with syntax highlighting
- **Entry Point**: `Editor.jsx:92` - `Editor` component
- **Involved Modules**: Editor.jsx, CodeMirror extensions
- **Complexity**: High (CodeMirror configuration, plugins)

### Feature 4: Live Preview
- **User Story**: User can see rendered markdown alongside the editor
- **Entry Point**: `Editor.jsx:230` - Preview pane
- **Involved Modules**: Editor.jsx, markdown-it
- **Complexity**: Medium

### Feature 5: File Save
- **User Story**: User can save edited content to disk
- **Entry Point**: `App.jsx:109` - `handleSave()`
- **Backend**: `app.go:60` - `SaveFile()`, `pkg/files/manager.go:169`
- **Involved Modules**: App.jsx, Editor.jsx, wailsjs, app.go, files/manager.go
- **Complexity**: Low

### Feature 6: Command Palette
- **User Story**: User can quickly search files and execute commands
- **Entry Point**: `CommandPalette.jsx:7` - `CommandPalette`
- **Involved Modules**: CommandPalette.jsx, App.jsx, fuse.js
- **Complexity**: Medium

### Feature 7: Zen Mode
- **User Story**: User can focus on writing by hiding UI chrome
- **Entry Point**: `App.jsx:159` - `setIsZenMode`
- **Involved Modules**: App.jsx, Editor.jsx
- **Complexity**: Low

### Feature 8: Settings Management
- **User Story**: User can customize fonts and appearance
- **Entry Point**: `SettingsModal.jsx:24`, `App.jsx:48` - `handleUpdateSettings`
- **Involved Modules**: SettingsModal.jsx, App.jsx, localStorage
- **Complexity**: Low

---

## 2. High-Frequency Code Identification

### Most Referenced Modules

| Module | Import Count | Referenced By |
|--------|--------------|---------------|
| `lucide-react` | 6 | All components |
| `clsx` | 5 | App.jsx, Editor.jsx, FileTree.jsx, CommandPalette.jsx, Toast.jsx |
| `../wailsjs/go/main/App` | 1 | App.jsx (multiple functions) |
| `@headlessui/react` | 2 | CommandPalette.jsx, SettingsModal.jsx, Toast.jsx |
| `markdown-it` | 1 | Editor.jsx |

### Critical Paths

#### Path 1: Open Folder Flow
```
App.jsx:handleOpenFolder()
  → wailsjs:OpenFolder()
    → app.go:OpenFolder()
      → runtime.OpenDirectoryDialog()
      → files.Manager.SetBasePath()
  → refreshFileTree()
    → wailsjs:ListFiles()
      → app.go:ListFiles()
        → files.Manager.ListFiles()
          → buildTree() [recursive]
```

#### Path 2: File Edit Flow
```
FileTree.jsx:handleClick()
  → App.jsx:handleFileSelect(node)
    → wailsjs:ReadFile(path)
      → app.go:ReadFile()
        → files.Manager.ReadFile()
    → setCurrentFile(), setCurrentContent()
  → Editor.jsx receives content via props
    → CodeMirror updates document
```

#### Path 3: File Save Flow
```
Editor.jsx:handleSave() [via Cmd+S or button]
  → App.jsx:handleSave(content)
    → wailsjs:SaveFile(path, content)
      → app.go:SaveFile()
        → files.Manager.SaveFile()
    → showToast('File saved successfully')
```

---

## 3. Component Dependency Graph

```
App.jsx (Root)
├── Toast
├── SettingsModal
├── CommandPalette
├── Header
└── Main Layout
    ├── FileTree
    │   └── FileTreeNode (recursive)
    └── Editor
        └── CodeMirror View
```

---

## 4. State Flow Analysis

### App.jsx State (Central)

| State | Type | Used By | Purpose |
|-------|------|---------|---------|
| `fileTree` | FileNode \| null | FileTree, CommandPalette | Folder structure |
| `currentFile` | FileNode \| null | Editor | Active file metadata |
| `currentContent` | string | Editor, handleSave | Active file content |
| `basePath` | string | Header | Selected folder path |
| `loading` | boolean | Header button | Loading indicator |
| `error` | string \| null | Error Banner | Error display |
| `sidebarWidth` | number | Sidebar, Resizer | Layout |
| `isResizing` | boolean | Resizer | Resize state |
| `isZenMode` | boolean | Header, Sidebar, Editor | Zen mode toggle |
| `isCommandPaletteOpen` | boolean | CommandPalette | Modal state |
| `isSettingsOpen` | boolean | SettingsModal | Modal state |
| `toast` | {show, message} | Toast | Notification state |
| `appSettings` | {fontInterface, fontText} | SettingsModal, CSS | User preferences |

### Editor.jsx State (Local)

| State | Type | Purpose |
|-------|------|---------|
| `viewMode` | 'edit' \| 'preview' \| 'split' | View mode toggle |
| `unsaved` | boolean | Unsaved changes indicator |

---

## 5. Key Code Patterns

### Pattern 1: Async Operation Handler
```javascript
// Found in App.jsx - repeated 4 times with slight variations
const handleSomeOperation = async () => {
  setLoading(true);
  setError(null);
  try {
    await SomeWailsMethod();
    // success handling
  } catch (err) {
    setError(err.message || 'Default error');
    console.error('Error:', err);
  } finally {
    setLoading(false);
  }
};
```

### Pattern 2: Props Drilling
```javascript
// App.jsx
<Editor
  content={currentContent}
  onChange={handleContentChange}
  onSave={handleSave}
  filename={currentFile?.name}
  isZenMode={isZenMode}
/>
```

### Pattern 3: LocalStorage Persistence
```javascript
// App.jsx - Settings
useEffect(() => {
  const saved = localStorage.getItem('notebit-settings');
  if (saved) {
    setAppSettings(prev => ({ ...prev, ...JSON.parse(saved) }));
  }
}, []);

const handleUpdateSettings = (key, value) => {
  const newSettings = { ...appSettings, [key]: value };
  localStorage.setItem('notebit-settings', JSON.stringify(newSettings));
};
```

---

## 6. Priority Analysis

### High Priority for Analysis
1. **Editor.jsx** - Most complex, 243 lines, multiple CodeMirror plugins
2. **App.jsx** - Central state management, 329 lines
3. **CommandPalette.jsx** - Search logic with Fuse.js

### Medium Priority
4. **SettingsModal.jsx** - Settings persistence
5. **FileTree.jsx** - Recursive component pattern

### Low Priority
6. **Toast.jsx** - Simple, isolated component

---

## Next Steps

Proceed to **Phase 2: Architecture Layer Analysis** to:
1. Check layering violations
2. Detect circular dependencies
3. Evaluate directory structure
