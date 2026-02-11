# Module Analysis: Supporting Components

**Files**: FileTree.jsx, CommandPalette.jsx, Toast.jsx, SettingsModal.jsx
**Purpose**: Supporting UI components

---

## 1. FileTree.jsx Analysis

**Lines**: 82

### Structure
- `FileTreeNode` (recursive component, lines 5-58)
- `FileTree` (main component, lines 61-79)

### Vibe Coding Problems

#### Problem M-018: Expanded State Lost on Re-render
**Severity**: Low
**Location**: Line 6
**Problem**: Each node manages its own `isExpanded` state, lost when parent re-renders

```javascript
const FileTreeNode = ({ node, level = 0, onSelect, selectedPath }) => {
  const [isExpanded, setIsExpanded] = useState(false);
  // State is local to each node instance
```

**Fix**: Lift expanded state to parent or use controlled pattern.

#### Problem M-019: No Keyboard Navigation
**Severity**: Medium
**Location**: Entire component
**Problem**: File tree not keyboard accessible

**Fix**: Add keyboard event handlers (Arrow keys, Enter).

#### Problem M-020: Duplicate Icon Logic
**Severity**: Low
**Location**: Lines 31-41
**Problem**: Similar icon rendering patterns for directory and file

```javascript
{node.isDir ? (
  <>
    {isExpanded ? <ChevronDown /> : <ChevronRight />}
    {isExpanded ? <FolderOpen /> : <Folder />}
  </>
) : (
  <>
    <span style={{ width: 16 }} />  {/* Spacer */}
    <File />
  </>
)}
```

**Fix**: Extract to icon components or use consistent pattern.

### Positive Patterns
- ✅ Uses existing icons from lucide-react
- ✅ Uses clsx for conditional classes
- ✅ Clean recursion pattern
- ✅ Empty state handling (lines 62-67)

---

## 2. CommandPalette.jsx Analysis

**Lines**: 132

### Structure
- Combines file search (Fuse.js) with command execution
- Uses Headless UI Combobox and Dialog

### Vibe Coding Problems

#### Problem M-021: Inline Function in Render
**Severity**: Low
**Location**: Lines 11-20
**Problem**: `flattenFiles` function recreated on every render

```javascript
const CommandPalette = ({ isOpen, setIsOpen, files, onFileSelect, commands }) => {
  // ...
  const flattenFiles = (node, acc = []) => {
    if (!node) return acc;
    // ...
  };
  const flatFiles = files ? flattenFiles(files) : [];
```

**Fix**: Use useMemo for flattened files.

#### Problem M-022: Fuse Instance Recreated
**Severity**: Medium
**Location**: Lines 29-32
**Problem**: New Fuse instance created on every render

```javascript
const fuse = new Fuse(allItems, {
  keys: ['label', 'name'],
  threshold: 0.3,
});
```

**Fix**: Memoize Fuse instance with useMemo.

#### Problem M-023: Inline Styles Not Extracted
**Severity**: Low
**Location**: Throughout
**Problem**: Could use Tailwind or CSS variables for consistency

### Positive Patterns
- ✅ Uses existing Headless UI components
- ✅ Uses existing Fuse.js library
- ✅ Clean transition animations
- ✅ Proper accessibility (Combobox)

---

## 3. Toast.jsx Analysis

**Lines**: 51

### Structure
- Simple notification component
- Auto-dismisses after duration

### Vibe Coding Problems

#### Problem M-024: Hardcoded Icon
**Severity**: Low
**Location**: Line 36
**Problem**: Always shows CheckCircle, not configurable

```javascript
<CheckCircle className="h-5 w-5 text-obsidian-green" />
```

**Fix**: Add `type` prop with different icons (success, error, warning, info).

#### Problem M-025: useEffect Missing Cleanup
**Severity**: Low (Actually has cleanup)
**Analysis**: Line 12 correctly returns cleanup function.

### Positive Patterns
- ✅ Uses Headless UI Transition
- ✅ Proper cleanup in useEffect
- ✅ Configurable duration
- ✅ Clean component structure

---

## 4. SettingsModal.jsx Analysis

**Lines**: 165

### Structure
- Modal for appearance settings
- Font selection with custom option

### Vibe Coding Problems

#### Problem M-026: Font Lists in Component
**Severity**: Low
**Location**: Lines 6-22
**Problem**: Font options hardcoded in component

```javascript
const FONTS_INTERFACE = [
  { name: 'System Default', value: '-apple-system, ...' },
  // ...
];
```

**Fix**: Extract to constants file.

#### Problem M-027: forceCustom State
**Severity**: Low
**Location**: Lines 25-27
**Problem**: Complex logic for showing custom input

```javascript
const [forceCustom, setForceCustom] = useState(false);
const isPreset = FONTS_TEXT.some(f => f.value === settings.fontText);
const showCustomInput = forceCustom || !isPreset;
```

**Analysis**: Logic is actually correct but could be clearer.

#### Problem M-028: Commented Out Code
**Severity**: Low
**Location**: Lines 79-82
**Problem**: Placeholder comments for future tabs

```javascript
{/* Placeholder for more tabs */}
{/* <button className="...">General</button> */}
```

**Fix**: Remove or create proper TODO.

### Positive Patterns
- ✅ Uses Headless UI Dialog and Transition
- ✅ Proper modal structure with header
- ✅ Real-time preview of font changes
- ✅ Auto-save behavior

---

## 5. Cross-Component Issues

### Issue M-029: Inconsistent Transition Patterns
Different components use different transition patterns:
- CommandPalette: Headless UI Transition with Fragment
- Toast: Headless UI Transition (without Fragment wrapper)
- SettingsModal: Headless UI Transition with Fragment

**Fix**: Standardize transition pattern.

### Issue M-030: Hardcoded Text Strings
Multiple components have hardcoded English text:
- FileTree: "No folder selected"
- CommandPalette: "Search files or commands...", "No results found."
- SettingsModal: "Settings", "Interface Font", "Text Font", etc.
- Toast: No text (good - passed as prop)

**Fix**: Consider i18n extraction for future.

---

## 6. Summary of Module Layer Problems

| Priority | Problem ID | Description | Location | Effort |
|----------|------------|-------------|----------|--------|
| High | M-022 | Fuse instance recreation | CommandPalette | Low |
| Medium | M-021 | flattenFiles recreation | CommandPalette | Low |
| Medium | M-019 | No keyboard navigation | FileTree | Medium |
| Low | M-018 | Expanded state lost | FileTree | Medium |
| Low | M-020 | Duplicate icon logic | FileTree | Low |
| Low | M-024 | Hardcoded Toast icon | Toast | Low |
| Low | M-026 | Font lists in component | SettingsModal | Low |
| Low | M-029 | Inconsistent transitions | Multiple | Low |
| Low | M-030 | Hardcoded strings | Multiple | Low |

---

## 7. Positive Patterns Found

| Pattern | Components | Notes |
|---------|------------|-------|
| Lucide React Icons | All | Consistent icon library |
| clsx for classes | 5/6 components | Good conditional class handling |
| Headless UI | 4/6 components | Accessible UI primitives |
| CSS Variables | All | Theme system used correctly |
