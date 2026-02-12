import { useState, useEffect, useCallback } from 'react';
import { FolderOpen, X, Monitor, Save, Settings, Menu, Sparkles } from 'lucide-react';
import { OpenFolder, ListFiles, ReadFile, SaveFile, SetFolder } from '../wailsjs/go/main/App';
import FileTree from './components/FileTree';
import Editor from './components/Editor';
import CommandPalette from './components/CommandPalette';
import Toast from './components/Toast';
import SettingsModal from './components/SettingsModal';
import SimilarNotesSidebar from './components/SimilarNotesSidebar';
import AIStatusIndicator from './components/AIStatusIndicator';
import clsx from 'clsx';
import { SIDEBAR, SEMANTIC_SEARCH } from './constants';

function App() {
  const [fileTree, setFileTree] = useState(null);
  const [currentFile, setCurrentFile] = useState(null);
  const [currentContent, setCurrentContent] = useState('');
  const [basePath, setBasePath] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  // UI State
  const [sidebarWidth, setSidebarWidth] = useState(SIDEBAR.DEFAULT_WIDTH);
  const [isSidebarOpen, setIsSidebarOpen] = useState(true);
  const [isResizing, setIsResizing] = useState(false);
  const [isZenMode, setIsZenMode] = useState(false);
  const [isCommandPaletteOpen, setIsCommandPaletteOpen] = useState(false);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [toast, setToast] = useState({ show: false, message: '' });

  // Right sidebar state
  const [isRightSidebarOpen, setIsRightSidebarOpen] = useState(true);
  const [rightSidebarWidth, setRightSidebarWidth] = useState(SEMANTIC_SEARCH.DEFAULT_WIDTH);
  const [isResizingRight, setIsResizingRight] = useState(false);
  const [searchTrigger, setSearchTrigger] = useState(0);  // Trigger for search on save

  // Settings State
  const [appSettings, setAppSettings] = useState({
    fontInterface: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif',
    fontText: '"Maple Mono NF CN", "SF Mono", "JetBrains Mono", "Segoe UI", Roboto, Helvetica, Arial, sans-serif'
  });

  useEffect(() => {
    // Load last opened folder
    const loadLastFolder = async () => {
      const lastFolder = localStorage.getItem('notebit-last-folder');
      if (lastFolder) {
        try {
          await SetFolder(lastFolder);
          setBasePath(lastFolder);
          const tree = await ListFiles();
          setFileTree(tree);
        } catch (e) {
          console.error('Failed to restore folder', e);
        }
      }
    };
    loadLastFolder();

    // Load settings from localStorage
    const saved = localStorage.getItem('notebit-settings');
    if (saved) {
        try {
            const parsed = JSON.parse(saved);
            setAppSettings(prev => ({ ...prev, ...parsed }));
            
            // Apply fonts
            if (parsed.fontInterface) document.documentElement.style.setProperty('--font-interface', parsed.fontInterface);
            if (parsed.fontText) document.documentElement.style.setProperty('--font-text', parsed.fontText);
        } catch (e) { console.error('Failed to load settings', e); }
    }
  }, []);

  const handleUpdateSettings = (key, value) => {
      const newSettings = { ...appSettings, [key]: value };
      setAppSettings(newSettings);
      localStorage.setItem('notebit-settings', JSON.stringify(newSettings));
      
      // Apply CSS variable
      if (key === 'fontInterface') {
          document.documentElement.style.setProperty('--font-interface', value);
      } else if (key === 'fontText') {
          document.documentElement.style.setProperty('--font-text', value);
      }
  };

  const showToast = (message) => {
    setToast({ show: true, message });
  };

  const handleOpenFolder = async () => {
    setLoading(true);
    setError(null);
    try {
      const path = await OpenFolder();
      if (path) {
        setBasePath(path);
        localStorage.setItem('notebit-last-folder', path);
        await refreshFileTree();
      }
    } catch (err) {
      setError(err.message || 'Failed to open folder');
      console.error('Error opening folder:', err);
    } finally {
      setLoading(false);
    }
  };

  const refreshFileTree = async () => {
    try {
      const tree = await ListFiles();
      setFileTree(tree);
    } catch (err) {
      setError(err.message || 'Failed to load files');
      console.error('Error loading files:', err);
    }
  };

  const handleFileSelect = async (node) => {
    if (node.isDir) return;

    setLoading(true);
    setError(null);
    try {
      const result = await ReadFile(node.path);
      setCurrentFile(node);
      setCurrentContent(result.content);
    } catch (err) {
      setError(err.message || 'Failed to read file');
      console.error('Error reading file:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (content) => {
    if (!currentFile) return;

    setLoading(true);
    setError(null);
    try {
      await SaveFile(currentFile.path, content);
      setCurrentContent(content);
      setSearchTrigger(prev => prev + 1);  // Trigger similarity search
      showToast('File saved successfully');
    } catch (err) {
      setError(err.message || 'Failed to save file');
      console.error('Error saving file:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleContentChange = (content) => {
    setCurrentContent(content);
  };

  // Sidebar Resizing Logic
  const startResizing = useCallback(() => setIsResizing(true), []);
  const stopResizing = useCallback(() => setIsResizing(false), []);
  const resize = useCallback(
    (e) => {
      if (isResizing) {
        setSidebarWidth(Math.max(SIDEBAR.MIN_WIDTH, Math.min(e.clientX, SIDEBAR.MAX_WIDTH)));
      }
    },
    [isResizing]
  );

  useEffect(() => {
    if (isResizing) {
      window.addEventListener('mousemove', resize);
      window.addEventListener('mouseup', stopResizing);
    }
    return () => {
      window.removeEventListener('mousemove', resize);
      window.removeEventListener('mouseup', stopResizing);
    };
  }, [isResizing, resize, stopResizing]);

  // Right Sidebar Resizing Logic
  const startResizingRight = useCallback(() => setIsResizingRight(true), []);
  const stopResizingRight = useCallback(() => setIsResizingRight(false), []);
  const resizeRight = useCallback(
    (e) => {
      if (isResizingRight) {
        const newWidth = window.innerWidth - e.clientX;
        setRightSidebarWidth(Math.max(SEMANTIC_SEARCH.MIN_WIDTH, Math.min(newWidth, SEMANTIC_SEARCH.MAX_WIDTH)));
      }
    },
    [isResizingRight]
  );

  useEffect(() => {
    if (isResizingRight) {
      window.addEventListener('mousemove', resizeRight);
      window.addEventListener('mouseup', stopResizingRight);
    }
    return () => {
      window.removeEventListener('mousemove', resizeRight);
      window.removeEventListener('mouseup', stopResizingRight);
    };
  }, [isResizingRight, resizeRight, stopResizingRight]);

  // Global Keyboard Shortcuts
  useEffect(() => {
    const handleKeyDown = (e) => {
      // Toggle Zen Mode (F11)
      if (e.key === 'F11') {
        e.preventDefault();
        setIsZenMode((prev) => !prev);
      }
      
      // Toggle Command Palette (Cmd+K)
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setIsCommandPaletteOpen((prev) => !prev);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  // Commands for Palette
  const commands = [
    {
        id: 'toggle-zen-mode',
        label: isZenMode ? 'Exit Zen Mode' : 'Enter Zen Mode',
        shortcut: 'F11',
        icon: Monitor,
        action: () => setIsZenMode((prev) => !prev)
    },
    {
        id: 'open-folder',
        label: 'Open Folder',
        icon: FolderOpen,
        action: handleOpenFolder
    },
    // Save command is tricky because it depends on editor state/content.
    // Ideally, we trigger the save on the current file with current content.
    // But currentContent in App state might be stale if Editor hasn't pushed up changes on every keystroke?
    // Actually Editor pushes onChange, so currentContent should be up to date.
    {
        id: 'save-file',
        label: 'Save File',
        shortcut: 'Cmd+S',
        icon: Save,
        action: () => handleSave(currentContent)
    }
  ];

  return (
    <div className="flex flex-col h-screen bg-primary text-normal font-sans">
      <Toast 
        show={toast.show} 
        message={toast.message} 
        onClose={() => setToast({ ...toast, show: false })} 
      />

      <SettingsModal
        isOpen={isSettingsOpen}
        onClose={() => setIsSettingsOpen(false)}
        settings={appSettings}
        onUpdateSettings={handleUpdateSettings}
      />

      <CommandPalette 
        isOpen={isCommandPaletteOpen}
        setIsOpen={setIsCommandPaletteOpen}
        files={fileTree}
        onFileSelect={handleFileSelect}
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
        <div className="flex gap-3">
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
            onSelect={handleFileSelect}
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

        {/* Editor Area */}
        <div className="flex-1 flex flex-col overflow-hidden bg-primary relative">
          {currentFile ? (
            <Editor
              content={currentContent}
              onChange={handleContentChange}
              onSave={handleSave}
              filename={currentFile.name}
              isZenMode={isZenMode}
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

        {/* Right Sidebar - Similar Notes */}
        <SimilarNotesSidebar
          query={currentContent}
          triggerSearch={searchTrigger}
          isOpen={isRightSidebarOpen && !isZenMode}
          onClose={() => setIsRightSidebarOpen(false)}
          onNoteClick={(note) => handleFileSelect({ path: note.path, name: note.title })}
          width={rightSidebarWidth}
        />

        {/* Right Sidebar Resizer Handle */}
        {!isZenMode && isRightSidebarOpen && (
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
