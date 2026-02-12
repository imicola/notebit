/**
 * Application Constants
 */

// Storage keys
export const STORAGE_KEYS = {
  SETTINGS: 'notebit-settings',
  SIDEBAR_WIDTH: 'notebit-sidebar-width'
};

// Sidebar dimensions
export const SIDEBAR = {
  DEFAULT_WIDTH: 280,
  MIN_WIDTH: 200,
  MAX_WIDTH: 600
};

// View modes for editor
export const VIEW_MODES = {
  EDIT: 'edit',
  PREVIEW: 'preview',
  SPLIT: 'split'
};

// Error messages
export const ERRORS = {
  OPEN_FOLDER: 'Failed to open folder',
  LOAD_FILES: 'Failed to load files',
  READ_FILE: 'Failed to read file',
  SAVE_FILE: 'Failed to save file',
  NO_BASE_PATH: 'No folder selected',
  APP_CRASH: 'Something went wrong. The application has crashed.'
};

// Toast duration (ms)
export const TOAST_DURATION = 2000;

// Font options
export const FONTS_INTERFACE = [
  { name: 'System Default', value: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif' },
  { name: 'Inter', value: '"Inter", sans-serif' },
  { name: 'Roboto', value: '"Roboto", sans-serif' },
  { name: 'Segoe UI', value: '"Segoe UI", sans-serif' },
  { name: 'Helvetica Neue', value: '"Helvetica Neue", Arial, sans-serif' },
];

export const FONTS_TEXT = [
  { name: 'Maple Mono', value: '"Maple Mono NF CN"' },
  { name: 'SF Mono (Default)', value: '"SF Mono", "JetBrains Mono", "Segoe UI", Roboto, Helvetica, Arial, sans-serif' },
  { name: 'JetBrains Mono', value: '"JetBrains Mono", monospace' },
  { name: 'Fira Code', value: '"Fira Code", monospace' },
  { name: 'Source Code Pro', value: '"Source Code Pro", monospace' },
  { name: 'Consolas', value: 'Consolas, monospace' },
  { name: 'System Sans', value: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif' },
];

// Default settings
export const DEFAULT_SETTINGS = {
  fontInterface: FONTS_INTERFACE[0].value,
  fontText: FONTS_TEXT[0].value
};

// Keyboard shortcuts
export const SHORTCUTS = {
  TOGGLE_ZEN_MODE: 'F11',
  COMMAND_PALETTE: 'Cmd+K',
  SAVE_FILE: 'Cmd+S'
};
