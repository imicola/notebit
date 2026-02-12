# Task: Add Error Boundary

## Status
- **ID**: task-M003
- **Status**: Completed
- **Created**: 2026-02-12
- **Completed**: 2026-02-12
- **Priority**: Medium

## Goal
Implement a React Error Boundary component to catch JavaScript errors anywhere in the child component tree, log those errors, and display a fallback UI instead of the component tree that crashed.

## Implementation Steps
- [x] Create `frontend/src/components/ErrorBoundary.jsx` class component
- [x] Add error logging logic (console or service)
- [x] Create a user-friendly fallback UI
- [x] Wrap the application root (or main routes) in `App.jsx` or `main.jsx` with the ErrorBoundary
- [x] Verify functionality

## References
- [React Error Boundaries Documentation](https://react.dev/reference/react/Component#catching-errors-in-component-tree)

## Changes
- Created `frontend/src/components/ErrorBoundary.jsx`
- Updated `frontend/src/main.jsx` to wrap App
- Updated `frontend/src/constants/index.js` with error message
