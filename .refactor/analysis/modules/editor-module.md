# Module Analysis: Editor.jsx (Markdown Editor)

**File**: `frontend/src/components/Editor.jsx`
**Lines**: 243
**Purpose**: CodeMirror-based Markdown editor with live preview

---

## 1. Basic Information

| Attribute | Value |
|-----------|-------|
| File Size | 243 lines |
| Component | Editor (function component) |
| Dependencies | @codemirror/*, @lezer/*, markdown-it, lucide-react, clsx |
| External Libraries | CodeMirror 6, markdown-it |
| Props | 5 (content, onChange, onSave, filename, isZenMode) |

---

## 2. Component Structure

### Props Interface
```javascript
const Editor = ({ content, onChange, onSave, filename, isZenMode }) => {
```

| Prop | Type | Purpose |
|------|------|---------|
| content | string | Document content |
| onChange | (content: string) => void | Content change callback |
| onSave | (content: string) => void | Save callback |
| filename | string | Current file name |
| isZenMode | boolean | Hide toolbar when true |

### Local State
```javascript
const [viewMode, setViewMode] = useState('split'); // 'edit', 'preview', 'split'
const [unsaved, setUnsaved] = useState(false);
```

---

## 3. CodeMirror Configuration Analysis

### Extensions Configured (Lines 116-143)
| Extension | Source | Purpose |
|-----------|--------|---------|
| lineNumbers | @codemirror/view | Line numbers gutter |
| highlightActiveLineGutter | @codemirror/view | Highlight active line number |
| history | @codemirror/commands | Undo/redo |
| drawSelection | @codemirror/view | Selection rendering |
| EditorState.allowMultipleSelections | @codemirror/state | Multi-cursor |
| markdown | @codemirror/lang-markdown | Markdown syntax |
| syntaxHighlighting | @codemirror/language | Syntax colors |
| editorTheme | Custom | Obsidian-like theme |
| highlightActiveLine | @codemirror/view | Highlight active line |
| keymap | @codemirror/commands | Keyboard shortcuts |
| highlightPlugin | Custom | ==highlight== syntax |
| wikiPlugin | Custom | [[wiki-link]] syntax |
| EditorView.updateListener | @codemirror/view | Change detection |

### Custom Syntax Highlighting (Lines 16-24)
```javascript
const obsidianHighlightStyle = HighlightStyle.define([
  { tag: tags.heading1, class: 'cm-heading-1 text-2xl font-bold text-accent' },
  { tag: tags.heading2, class: 'cm-heading-2 text-xl font-bold text-accent' },
  // ...
]);
```

**Problem**: Tailwind classes in syntax highlighting may not apply correctly to CodeMirror.

### Custom Decorators (Lines 27-74)

#### Dead Code Detected
```javascript
// Lines 37-57: obsidianExtensionsPlugin - NOT USED
const obsidianExtensionsPlugin = ViewPlugin.fromClass(...);

// This plugin is defined but never included in extensions array!
```

**Issue**: `obsidianExtensionsPlugin` is defined but unused. The separate `highlightPlugin` and `wikiPlugin` are used instead.

---

## 4. Vibe Coding Problem Detection

### Problem M-010: Dead Code
**Severity**: Low
**Location**: Lines 37-57
**Problem**: `obsidianExtensionsPlugin` is defined but never used

```javascript
// Dead code - defined but never added to extensions
const obsidianExtensionsPlugin = ViewPlugin.fromClass(
  class {
    constructor(view) {
      this.highlightDeco = highlightDecorator.createDeco(view);
      this.wikiDeco = wikiLinkDecorator.createDeco(view);
    }
    // ...
  },
  { decorations: (v) => { ... } }
);
```

**Fix**: Remove dead code.

### Problem M-011: Save Handler in useEffect Dependency
**Severity**: Medium
**Location**: Line 131
**Problem**: `handleSave` referenced in keymap but closure may be stale

```javascript
keymap.of([
  ...defaultKeymap,
  ...historyKeymap,
  { key: "Mod-s", run: () => { handleSave(); return true; } }  // handleSave closure
]),
```

**Issue**: `handleSave` is defined inside component but referenced in extension config that's only created once. If `handleSave` changes, the keymap won't update.

**Fix**: Use `useRef` for callback or recreate keymap on save callback change.

### Problem M-012: useMemo for markdown-it
**Severity**: Low (Not a problem, actually good pattern)
**Location**: Lines 99-110
**Analysis**: Correct use of useMemo to prevent recreation

```javascript
const md = useMemo(() => {
  const m = new MarkdownIt({...});
  m.use(markdownItGithubAlerts);
  return m;
}, []);  // Empty deps = created once
```

### Problem M-013: Direct DOM Manipulation Risk
**Severity**: Medium
**Location**: Line 234
**Problem**: `dangerouslySetInnerHTML` with user content

```javascript
<div dangerouslySetInnerHTML={{ __html: md.render(content || '') }} />
```

**Risk**: If markdown-it has XSS vulnerability, user notes could execute scripts.

**Fix**: Consider sanitization or use React component rendering.

### Problem M-014: Editor Instance Not Updated on Props Change
**Severity**: Low
**Location**: Lines 113-155
**Problem**: Editor only initialized once on mount; content updates handled separately

This is actually correct pattern, but the `useEffect` for content update (lines 158-166) could have edge cases.

---

## 5. Resource Reuse Check

### Using Existing Infrastructure
| Resource | Status |
|----------|--------|
| Theme Variables | ✅ Using `var(--font-text)`, `var(--text-normal)`, etc. |
| Tailwind Classes | ✅ Using clsx + Tailwind |
| lucide-react Icons | ✅ Using Split, Eye, Edit3, Save |

### Missing Reuse Opportunities
| Resource | Issue |
|----------|-------|
| Toolbar Buttons | Could be extracted to shared component |
| View Mode Toggle | Could be reusable component |

---

## 6. Code Quality Issues

### Issue M-015: Magic Strings for View Mode
```javascript
const [viewMode, setViewMode] = useState('split'); // 'edit', 'preview', 'split'
```
**Fix**: Use constants or enum.

### Issue M-016: Inconsistent Class Application
```javascript
// Line 17-24: Tailwind classes for syntax
{ tag: tags.heading1, class: 'cm-heading-1 text-2xl font-bold text-accent' },

// But custom CSS classes (cm-heading-1) not defined in style.css
```
**Issue**: `cm-heading-1`, `cm-heading-2` classes are referenced but not defined.

### Issue M-017: Complex Conditional JSX
Lines 222-236 have complex conditional class logic:
```javascript
<div className={clsx("h-full overflow-hidden transition-all duration-300",
    viewMode === 'preview' ? "hidden" : (viewMode === 'split' ? "w-1/2 border-r border-modifier-border" : "w-full"),
    isZenMode && "bg-primary"
)}>
```

---

## 7. Refactoring Recommendations

### High Priority
1. **Remove dead code** - Delete unused `obsidianExtensionsPlugin`
2. **Fix save callback closure** - Use ref pattern for callback in keymap

### Medium Priority
3. **Extract CodeMirror config** - Create `useCodeMirrorConfig` hook
4. **Extract Preview component** - Separate preview pane logic
5. **Extract Toolbar component** - Separate toolbar from editor

### Low Priority
6. **Add sanitization** - Protect against XSS in preview
7. **Define missing CSS classes** - cm-heading-1, cm-heading-2, etc.
8. **Extract view mode constants** - VIEW_MODES enum

---

## 8. Suggested Refactored Structure

```javascript
// Editor.jsx (simplified)
const Editor = ({ content, onChange, onSave, filename, isZenMode }) => {
  const editorRef = useRef(null);
  const [viewMode, setViewMode] = useState(VIEW_MODES.SPLIT);
  const [unsaved, setUnsaved] = useState(false);
  const onSaveRef = useLatest(onSave);  // Always current callback

  const { view, extensions } = useCodeMirrorSetup({
    content,
    onChange,
    onSave: () => onSaveRef.current?.(view.current.state.doc.toString()),
    onDirty: () => setUnsaved(true),
  });

  return (
    <div className="flex flex-col h-full w-full">
      {!isZenMode && <EditorToolbar {...toolbarProps} />}
      <EditorContent viewMode={viewMode}>
        <CodeMirrorPane ref={editorRef} view={view} />
        <PreviewPane content={content} />
      </EditorContent>
    </div>
  );
};
```

---

## 9. Metrics

| Metric | Before | Target |
|--------|--------|--------|
| Lines | 243 | ~100 (main) + hooks |
| Dead Code | 20 lines | 0 |
| Props | 5 | 5 (unchanged) |
| Local State | 2 | 2 (unchanged) |
