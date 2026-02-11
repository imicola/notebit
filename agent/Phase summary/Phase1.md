# Progress Summary: Notebit Phase 1 - The Sanctuary

We've been building the foundation of **Notebit**, a Local-First Markdown note-taking app for PKM enthusiasts. The goal was to create a distraction-free writing environment ("The Sanctuary") where users can open folders and edit markdown files without any AI intrusion—pure editor, pure focus.

## What We Built

### Backend: File System Manager

Created `pkg/files/` package to handle all file system operations. The key insight here was separating file system logic from the Wails app layer—this keeps the code testable and clean.

- **OpenFolder**: Directory picker dialog via `runtime.OpenDirectoryDialog`
- **ListFiles**: Recursive tree builder returning `FileNode` structures
- **ReadFile/SaveFile**: Simple file I/O for markdown content
- **Create/Delete/Rename**: CRUD operations for notes

One tricky bit was handling `time.Time` in JSON serialization. Wails doesn't know how to marshal Go's time type, so we wrapped it in a custom `JSONTime` struct with proper `MarshalJSON/UnmarshalJSON` methods.

```
┌───────────────────────────────────────────────────────────┐
│  Frontend (React)                                  │
│    ↓ OpenFolder()                                    │
├───────────────────────────────────────────────────────────┤
│  Wails Binding (app.go)                              │
│    ↓ fm.SetBasePath(dir)                              │
├───────────────────────────────────────────────────────────┤
│  files.Manager (pkg/files/)                            │
│    ↓ buildTree() recursive traversal                     │
├───────────────────────────────────────────────────────────┤
│  File System (OS)                                    │
└───────────────────────────────────────────────────────────┘
```

### Frontend: Editor + File Tree

The UI is intentionally minimal—VSCode-inspired dark theme, no clutter.

**FileTree Component**: Recursive tree with expand/collapse for folders, lucide-react icons (Folder/FolderOpen/File), and selection highlighting.

**Editor Component**: Built with `react-markdown` and `remark-gfm` for GitHub-flavored markdown support. The editor has three modes:
- **Edit**: Pure markdown source
- **Split**: Side-by-side editing with live preview
- **Preview**: Rendered markdown only

We hit a snag with `@uiw/react-md-editor`—it had dependency resolution issues with Vite 3 (#minpath imports from vfile package). Switched to a lighter approach: raw textarea + `react-markdown` for preview. This actually aligns better with our "no AI autocomplete" constraint since we control the textarea entirely.

```
┌────────────────────────────────────────────────────────────┐
│  App Header                                          │
│  ┌─────────────┐  ┌─────────────────┐               │
│  │ Notebit     │  │ [Open Folder]   │               │
│  │ The Sanctuary│  └─────────────────┘               │
│  └─────────────┘                                     │
├────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────────────────────┐   │
│  │ File Tree    │  │ Editor                      │   │
│  │              │  │ ┌──────────┬───────────┐   │   │
│  │ 📁 notes/    │  │ │ Markdown │ Preview   │   │   │
│  │   📄 todo.md │  │ │ Edit     │           │   │   │
│  │   📄 welcome │  │ └──────────┴───────────┘   │   │
│  │              │  │ [Edit] [Split] [Preview]    │   │
│  └──────────────┘  │ [Save] (Ctrl+S)            │   │
│                   └───────────────────────────────┘   │
└────────────────────────────────────────────────────────────┘
```

### Vite Configuration Fix

The `react-markdown` dependency chain includes packages using subpath imports (`#minpath`, `#minurl`, etc.) that Vite's Rollup doesn't resolve by default. Fixed with regex external pattern in `vite.config.js`:

```js
external: [/^(?:#[a-z]+)$/]
```

## Current State

The app is **running and functional** at `http://localhost:34115`. You can:

1. Click "Open Folder" to select a directory
2. Browse the file tree (folders expand/collapse)
3. Click any `.md` file to open it in the editor
4. Edit in split view with live preview
5. Save with Ctrl+S or the Save button

Created `test-notes/` folder with sample markdown files to verify functionality.

## What's Next

Phase 1 is complete. The MVP for "The Sanctuary" is done—pure markdown editing works, local file I/O works, UI is clean.

**Phase 2** will add "The Silent Curator":
- SQLite database for metadata and vector embeddings
- `fsnotify` watcher to trigger background embedding on file save
- Basic semantic search (console output first)

Ready to start when you are.