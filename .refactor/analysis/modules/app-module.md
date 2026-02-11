# Module Analysis: App.jsx (Main Application)

**File**: `frontend/src/App.jsx`
**Lines**: 329
**Purpose**: Root component, state management, orchestration

---

## 1. Basic Information

| Attribute | Value |
|-----------|-------|
| File Size | 329 lines |
| Component | App (function component) |
| Dependencies | lucide-react, clsx, wailsjs |
| State Variables | 12 useState hooks |
| Effect Hooks | 4 useEffect hooks |

---

## 2. State Analysis

### State Variables (12 total)

```javascript
// Data State
const [fileTree, setFileTree] = useState(null);
const [currentFile, setCurrentFile] = useState(null);
const [currentContent, setCurrentContent] = useState('');
const [basePath, setBasePath] = useState('');

// UI State
const [loading, setLoading] = useState(false);
const [error, setError] = useState(null);
const [sidebarWidth, setSidebarWidth] = useState(280);
const [isResizing, setIsResizing] = useState(false);
const [isZenMode, setIsZenMode] = useState(false);
const [isCommandPaletteOpen, setIsCommandPaletteOpen] = useState(false);
const [isSettingsOpen, setIsSettingsOpen] = useState(false);
const [toast, setToast] = useState({ show: false, message: '' });

// Settings State
const [appSettings, setAppSettings] = useState({...});
```

### State Grouping Opportunity
| Group | States | Suggested Extraction |
|-------|--------|----------------------|
| File Operations | fileTree, currentFile, currentContent, basePath | `useFileOperations()` |
| UI State | sidebarWidth, isResizing, isZenMode | `useLayout()` |
| Modal State | isCommandPaletteOpen, isSettingsOpen | Combine or context |
| Feedback | loading, error, toast | `useFeedback()` |
| Settings | appSettings | `useSettings()` |

---

## 3. Vibe Coding Problem Detection

### Problem M-001: Repeated Async Handler Pattern
**Severity**: High
**Location**: Lines 65-79, 92-107, 109-124
**Problem**: Same try-catch-finally pattern repeated 3 times

```javascript
// Pattern repeated 3 times:
const handleXXX = async () => {
  setLoading(true);
  setError(null);
  try {
    // operation
  } catch (err) {
    setError(err.message || '...');
    console.error('...', err);
  } finally {
    setLoading(false);
  }
};
```

**Fix**: Create a generic async handler wrapper or custom hook.

### Problem M-002: Direct API Calls
**Severity**: Medium
**Location**: Throughout file
**Problem**: No service layer abstraction

```javascript
// Current
import { OpenFolder, ListFiles, ReadFile, SaveFile } from '../wailsjs/go/main/App';

// Direct calls in handlers
const path = await OpenFolder();
const tree = await ListFiles();
```

**Fix**: Create `fileService.js` to wrap API calls.

### Problem M-003: No Error Boundary
**Severity**: Medium
**Location**: Entire component
**Problem**: React errors will crash the entire app

**Fix**: Add React Error Boundary wrapper.

### Problem M-004: Settings Logic Mixed with Component
**Severity**: Low
**Location**: Lines 33-59
**Problem**: Settings persistence logic embedded in component

```javascript
useEffect(() => {
  const saved = localStorage.getItem('notebit-settings');
  if (saved) {
    // parse and apply settings
  }
}, []);

const handleUpdateSettings = (key, value) => {
  // save to localStorage
  // apply CSS variables
};
```

**Fix**: Extract to `useSettings()` hook.

### Problem M-005: Sidebar Resize Logic Embedded
**Severity**: Low
**Location**: Lines 131-151
**Problem**: Resize logic not reusable

**Fix**: Extract to `useResizable()` hook.

### Problem M-006: Keyboard Shortcuts Mixed with Component
**Severity**: Low
**Location**: Lines 154-170
**Problem**: Global keyboard handler in main component

**Fix**: Extract to `useKeyboardShortcuts()` hook.

---

## 4. Resource Reuse Check

### Infrastructure Usage (vs Phase 0 Inventory)

| Infrastructure | Available | Currently Using | Status |
|----------------|-----------|-----------------|--------|
| Logging | ❌ Not found | console.error | ⚠️ Ad-hoc |
| Error Handling | ❌ Not found | try-catch | ⚠️ Ad-hoc |
| Storage | ❌ Not found | Direct localStorage | ⚠️ Ad-hoc |
| Theme | ✅ CSS variables | Applied correctly | ✅ Good |

### Component Reuse (vs Phase 0 Inventory)

| Component | Available | Currently Using | Status |
|-----------|-----------|-----------------|--------|
| Toast | ✅ Available | Imported and used | ✅ Good |
| Modal Pattern | ⚠️ SettingsModal | Used for settings | ✅ Good |

---

## 5. Code Quality Issues

### Issue M-007: Magic Numbers
- Line 20: `280` (default sidebar width)
- Line 136: `200`, `600` (min/max sidebar width)
- Line 136: `e.clientX` (implicit assumption)

### Issue M-008: Hardcoded Strings
- `"notebit-settings"` (localStorage key)
- `"Failed to open folder"`, `"Failed to load files"`, etc.
- `"Search files or commands..."`
- `"Welcome to Notebit"`

### Issue M-009: Complex JSX in Single File
The return statement spans 125 lines (200-325) with:
- Toast component
- SettingsModal component
- CommandPalette component
- Header section
- Error Banner
- Main layout with Sidebar, Resizer, Editor area
- Empty state

---

## 6. Refactoring Recommendations

### High Priority
1. **Extract `useFileOperations` hook** - Consolidate file-related state and handlers
2. **Create `withAsyncHandler` utility** - Standardize error handling pattern
3. **Create `fileService.js`** - Abstract API calls

### Medium Priority
4. **Extract `useSettings` hook** - Settings persistence
5. **Extract `useToast` hook** - Toast state management
6. **Split JSX into layout components** - Header, Sidebar, MainContent

### Low Priority
7. **Extract constants** - Magic numbers, hardcoded strings
8. **Extract `useResizable` hook** - Sidebar resize logic
9. **Extract `useKeyboardShortcuts` hook** - Global shortcuts

---

## 7. Suggested Refactored Structure

```javascript
// App.jsx (simplified)
function App() {
  const { fileTree, currentFile, handleFileSelect, handleSave } = useFileOperations();
  const { settings, updateSettings } = useSettings();
  const { toast, showToast, hideToast } = useToast();
  const { isZenMode, toggleZenMode } = useZenMode();
  const [isCommandPaletteOpen, setIsCommandPaletteOpen] = useState(false);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);

  useKeyboardShortcuts({
    onToggleZen: toggleZenMode,
    onOpenCommandPalette: () => setIsCommandPaletteOpen(true),
  });

  return (
    <AppProvider>
      <Toast {...toast} onClose={hideToast} />
      <SettingsModal isOpen={isSettingsOpen} onClose={...} />
      <CommandPalette isOpen={isCommandPaletteOpen} />
      <Header onSettings={() => setIsSettingsOpen(true)} />
      <MainLayout>
        <Sidebar fileTree={fileTree} onSelect={handleFileSelect} />
        <Editor content={currentFile?.content} onSave={handleSave} />
      </MainLayout>
    </AppProvider>
  );
}
```

---

## 8. Metrics

| Metric | Before | Target |
|--------|--------|--------|
| Lines | 329 | ~100 |
| useState | 12 | 4-5 |
| useEffect | 4 | 1-2 |
| Handler Functions | 8 | 2-3 |
