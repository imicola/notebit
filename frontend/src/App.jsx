import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { FolderOpen, X, Monitor, Save, Settings, Menu, Sparkles, MessageSquare, Network } from 'lucide-react';
import { fileService } from './services/fileService';
import FileTree from './components/FileTree';
import Editor from './components/Editor';
import ChatPanel from './components/ChatPanel';
import GraphPanel from './components/GraphPanel';
import CommandPalette from './components/CommandPalette';
import Toast from './components/Toast';
import SettingsModal from './components/SettingsModal';
import SimilarNotesSidebar from './components/SimilarNotesSidebar';
import AIStatusIndicator from './components/AIStatusIndicator';
import clsx from 'clsx';
import { SIDEBAR, SEMANTIC_SEARCH, STORAGE_KEYS } from './constants';
import { useSettings, useToast, useResizable, useKeyboardShortcuts, useFileOperations, useTheme } from './hooks';

function App() {
  // --- Hooks integration ---
  const { settings, updateSetting } = useSettings();
  const { toast, showToast, hideToast } = useToast();
  
  // Initialize theme system (must be called early)
  useTheme();
  
  const {
    fileTree, currentFile, currentContent, basePath,
    loading, error,
    openFolder: baseOpenFolder, selectFile, saveFile: baseSaveFile,
    updateContent, refreshFileTree, setFileTree, setBasePath, setError
  } = useFileOperations({ onSuccess: (msg) => showToast(msg) });

  const {
    width: sidebarWidth,
    isResizing,
    startResizing,
  } = useResizable({
    defaultWidth: SIDEBAR.DEFAULT_WIDTH,
    minWidth: SIDEBAR.MIN_WIDTH,
    maxWidth: SIDEBAR.MAX_WIDTH,
    persist: true,
  });

  const {
    width: rightSidebarWidth,
    isResizing: isResizingRight,
    startResizing: startResizingRight,
  } = useResizable({
    defaultWidth: SEMANTIC_SEARCH.DEFAULT_WIDTH,
    minWidth: SEMANTIC_SEARCH.MIN_WIDTH,
    maxWidth: SEMANTIC_SEARCH.MAX_WIDTH,
    persist: false,
  });

  // UI State
  const [isSidebarOpen, setIsSidebarOpen] = useState(true);
  const [isZenMode, setIsZenMode] = useState(false);
  const [isCommandPaletteOpen, setIsCommandPaletteOpen] = useState(false);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);

  // Right sidebar state
  const [isRightSidebarOpen, setIsRightSidebarOpen] = useState(true);
  const [searchRequest, setSearchRequest] = useState(null);

  // View mode state
  const [viewMode, setViewMode] = useState('editor');

  // Restore last opened folder on mount
  useEffect(() => {
    const loadLastFolder = async () => {
      const lastFolder = localStorage.getItem('notebit-last-folder');
      if (lastFolder) {
        try {
          await fileService.setFolder(lastFolder);
          await refreshFileTree();
          setBasePath(lastFolder);
        } catch (e) {
          console.error('Failed to restore folder', e);
        }
      }
    };
    loadLastFolder();
  }, []);

  // Wrap openFolder to persist to localStorage
  const handleOpenFolder = useCallback(async () => {
    const path = await baseOpenFolder();
    if (path) {
      localStorage.setItem('notebit-last-folder', path);
    }
  }, [baseOpenFolder]);

  // Wrap saveFile to trigger similarity search
  const handleSave = useCallback(async (content) => {
    const result = await baseSaveFile(content);
    if (result) {
      setSearchRequest({
        id: Date.now(),
        content,
      });
    }
  }, [baseSaveFile]);

  // Keyboard shortcuts via hook
  useKeyboardShortcuts({
    'F11': () => setIsZenMode(prev => !prev),
    'Mod+Shift+f': () => setIsZenMode(prev => !prev), // Focus Mode (same as Zen Mode)
    'Mod+k': () => setIsCommandPaletteOpen(prev => !prev),
  });

  // Handle open-file event from ChatPanel and GraphPanel
  useEffect(() => {
    const handleOpenFile = (e) => {
      const { path } = e.detail;
      const findAndSelectFile = (nodes) => {
        for (const node of nodes) {
          if (node.path === path && !node.isDir) {
            selectFile(node);
            return true;
          }
          if (node.children && findAndSelectFile(node.children)) {
            return true;
          }
        }
        return false;
      };
      if (fileTree) findAndSelectFile(fileTree);
      setViewMode('editor');
    };
    window.addEventListener('open-file', handleOpenFile);
    return () => window.removeEventListener('open-file', handleOpenFile);
  }, [fileTree, selectFile]);

  // Ref for currentContent to avoid stale closure in commands
  const currentContentRef = useRef(currentContent);
  currentContentRef.current = currentContent;

  // Commands for Palette
  const commands = useMemo(() => [
    {
      id: 'toggle-zen-mode',
      label: isZenMode ? 'Exit Zen Mode' : 'Enter Zen Mode',
      shortcut: 'F11',
      icon: Monitor,
      action: () => setIsZenMode(prev => !prev),
    },
    {
      id: 'open-folder',
      label: 'Open Folder',
      icon: FolderOpen,
      action: handleOpenFolder,
    },
    {
      id: 'save-file',
      label: 'Save File',
      shortcut: 'Cmd+S',
      icon: Save,
      action: () => handleSave(currentContentRef.current),
    },
  ], [isZenMode, handleOpenFolder, handleSave]);

  // Stable callback refs for child components
  const handleCloseSettings = useCallback(() => setIsSettingsOpen(false), []);
  const handleCloseRightSidebar = useCallback(() => setIsRightSidebarOpen(false), []);
  const handleNoteClick = useCallback((note) => selectFile({ path: note.path, name: note.title }), [selectFile]);

  return (
    <div className="flex flex-col h-screen bg-primary text-normal font-sans">
      <Toast 
        show={toast.show} 
        message={toast.message} 
        onClose={hideToast} 
      />

      <SettingsModal
        isOpen={isSettingsOpen}
        onClose={handleCloseSettings}
        settings={settings}
        onUpdateSettings={updateSetting}
      />

      <CommandPalette 
        isOpen={isCommandPaletteOpen}
        setIsOpen={setIsCommandPaletteOpen}
        files={fileTree}
        onFileSelect={selectFile}
        commands={commands}
      />

      {/* Header */}
      <header
        className={clsx(
          'flex justify-between items-center px-5 py-3 bg-secondary border-b border-modifier-border h-[60px] shrink-0 transition-all duration-300',
          isZenMode ? 'hidden' : 'flex'
        )}
      >
        <div className="flex items-center">
            <button
                className="mr-2 text-muted hover:text-normal transition-colors"
                onClick={() => setIsRightSidebarOpen(!isRightSidebarOpen)}
                title={isRightSidebarOpen ? "Close Related Notes" : "Open Related Notes"}
            >
                <Sparkles size={18} />
            </button>
            <button
                className="mr-3 text-muted hover:text-normal transition-colors"
                onClick={() => setIsSidebarOpen(!isSidebarOpen)}
                title={isSidebarOpen ? "Close Sidebar" : "Open Sidebar"}
            >
                <Menu size={20} />
            </button>
            <div className="flex flex-col">
              <div className="flex items-center gap-2">
                <h1 className="text-lg font-semibold text-normal leading-tight">Notebit</h1>
                <AIStatusIndicator />
              </div>
              <span className="text-xs text-muted">The Sanctuary</span>
            </div>
        </div>
        <div className="flex gap-2 items-center">
          {/* View Mode Buttons */}
          <div className="flex gap-1 bg-modifier-hover rounded-lg p-1 mr-3">
            <button
              onClick={() => setViewMode('editor')}
              className={clsx(
                'p-1.5 rounded transition-colors',
                viewMode === 'editor' ? 'bg-primary-alt text-normal' : 'text-muted hover:text-normal'
              )}
              title="Editor View"
            >
              <MessageSquare size={16} />
            </button>
            <button
              onClick={() => setViewMode('chat')}
              className={clsx(
                'p-1.5 rounded transition-colors',
                viewMode === 'chat' ? 'bg-primary-alt text-normal' : 'text-muted hover:text-normal'
              )}
              title="Chat View"
            >
              <Sparkles size={16} />
            </button>
            <button
              onClick={() => setViewMode('graph')}
              className={clsx(
                'p-1.5 rounded transition-colors',
                viewMode === 'graph' ? 'bg-primary-alt text-normal' : 'text-muted hover:text-normal'
              )}
              title="Graph View"
            >
              <Network size={16} />
            </button>
          </div>
          <button
            className="flex items-center justify-center w-9 h-9 bg-modifier-hover text-muted border border-modifier-border rounded hover:bg-modifier-border-focus hover:text-normal transition-colors"
            onClick={() => setIsSettingsOpen(true)}
            title="Settings"
          >
            <Settings size={18} />
          </button>
          <button
            className="flex items-center gap-2 px-4 py-2 bg-modifier-hover text-normal border border-modifier-border rounded hover:bg-modifier-border-focus disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-sm font-medium"
            onClick={handleOpenFolder}
            disabled={loading}
          >
            <FolderOpen size={18} />
            <span>Open Folder</span>
          </button>
        </div>
      </header>

      {/* Error Banner */}
      {error && (
        <div className="flex justify-between items-center px-5 py-3 bg-obsidian-red/20 border-b border-obsidian-red/50 text-obsidian-red text-sm">
          <span>{error}</span>
          <button onClick={() => setError(null)} className="hover:text-normal">
            <X size={18} />
          </button>
        </div>
      )}

      {/* Main Layout */}
      <main className="flex flex-1 overflow-hidden relative">
        {/* Sidebar */}
        <aside
          className={clsx(
            'flex flex-col bg-secondary border-r border-modifier-border shrink-0 transition-all duration-300',
            (isZenMode || !isSidebarOpen) ? 'hidden' : 'flex'
          )}
          style={{ width: (isZenMode || !isSidebarOpen) ? 0 : sidebarWidth }}
        >
          <div className="px-4 py-3 bg-secondary border-b border-modifier-border">
            <h2 className="text-xs font-bold uppercase tracking-wider text-muted">Files</h2>
            {basePath && <span className="block text-[10px] text-faint truncate mt-1" title={basePath}>{basePath}</span>}
          </div>
          <FileTree
            tree={fileTree}
            onSelect={selectFile}
            selectedPath={currentFile?.path}
          />
        </aside>

        {/* Resizer Handle */}
        {!isZenMode && isSidebarOpen && (
          <div
            className="w-1 bg-transparent hover:bg-accent cursor-col-resize absolute top-0 bottom-0 z-10 transition-colors"
            style={{ left: sidebarWidth }}
            onMouseDown={startResizing}
          />
        )}

        {/* Main Content Area */}
        {viewMode === 'chat' ? (
          /* Chat Panel View */
          <div className="flex-1 overflow-hidden bg-primary">
            <ChatPanel />
          </div>
        ) : viewMode === 'graph' ? (
          /* Graph Panel View */
          <div className="flex-1 overflow-hidden bg-primary">
            <GraphPanel />
          </div>
        ) : (
          /* Editor View (default) */
          <div className="flex-1 flex flex-col overflow-hidden bg-primary relative">
            {currentFile ? (
              <Editor
                content={currentContent}
                onChange={updateContent}
                onSave={handleSave}
                filename={currentFile.name}
                isZenMode={isZenMode}
                fileTree={fileTree}
                onSelectFile={selectFile}
              />
            ) : (
              <div className="flex flex-col items-center justify-center h-full text-faint">
                <div className="opacity-30 mb-5">
                  <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                    <polyline points="14 2 14 8 20 8"></polyline>
                    <line x1="16" y1="13" x2="8" y2="13"></line>
                    <line x1="16" y1="17" x2="8" y2="17"></line>
                    <polyline points="10 9 9 9 8 9"></polyline>
                  </svg>
                </div>
                <h2 className="text-2xl font-medium text-muted mb-2">Welcome to Notebit</h2>
                <p className="text-sm">Select a file from the sidebar or open a folder to get started</p>
                <div className="mt-8 text-xs text-muted flex gap-4">
                  <span className="flex items-center gap-1"><span className="px-1.5 py-0.5 bg-primary-alt rounded border border-modifier-border font-mono">Cmd+K</span> to search</span>
                  <span className="flex items-center gap-1"><span className="px-1.5 py-0.5 bg-primary-alt rounded border border-modifier-border font-mono">F11</span> for Zen Mode</span>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Right Sidebar - Similar Notes */}
        {/* Only show in editor mode */}
        {viewMode === 'editor' && (
          <SimilarNotesSidebar
            query={currentContent}
            searchRequest={searchRequest}
            basePath={basePath}
            currentPath={currentFile?.path}
            isOpen={isRightSidebarOpen && !isZenMode}
            onClose={handleCloseRightSidebar}
            onNoteClick={handleNoteClick}
            width={rightSidebarWidth}
          />
        )}

        {/* Right Sidebar Resizer Handle */}
        {/* Only show in editor mode */}
        {viewMode === 'editor' && !isZenMode && isRightSidebarOpen && (
          <div
            className="w-1 bg-transparent hover:bg-accent cursor-col-resize absolute top-0 bottom-0 z-10 transition-colors"
            style={{ right: rightSidebarWidth - 4 }}
            onMouseDown={startResizingRight}
          />
        )}
      </main>
    </div>
  );
}

export default App;
