# Task: Add Preview Sanitization

## Status
- **ID**: task-M013
- **Status**: Completed
- **Created**: 2026-02-12
- **Completed**: 2026-02-12
- **Priority**: Medium

## Goal
Sanitize HTML content in the Markdown preview to prevent Cross-Site Scripting (XSS) attacks. The application uses CodeMirror for editing, but we need to check how the preview is rendered.

## Implementation Steps
- [x] Check if `dompurify` is installed (It was missing)
- [x] Install `dompurify` if missing: `npm install dompurify`
- [x] Create `frontend/src/utils/sanitizer.js` (optional, or just import in component) -> Imported directly in `Editor.jsx`
- [x] Update the component rendering the preview (likely `Editor.jsx` or a separate Preview component) to use `DOMPurify.sanitize()` before `dangerouslySetInnerHTML`.
- [x] Verify functionality with a test string (e.g. `<img src=x onerror=alert(1)>`)

## References
- [DOMPurify](https://github.com/cure53/DOMPurify)

## Changes
- Installed `dompurify`
- Updated `frontend/src/components/Editor.jsx` to use `DOMPurify.sanitize` with `ADD_ATTR: ['target', 'class']`
