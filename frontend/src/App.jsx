import { useState } from 'react';
import { FolderOpen } from 'lucide-react';
import { OpenFolder, ListFiles, ReadFile, SaveFile } from '../wailsjs/go/main/App';
import FileTree from './components/FileTree';
import Editor from './components/Editor';
import './App.css';

function App() {
  const [fileTree, setFileTree] = useState(null);
  const [currentFile, setCurrentFile] = useState(null);
  const [currentContent, setCurrentContent] = useState('');
  const [basePath, setBasePath] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleOpenFolder = async () => {
    setLoading(true);
    setError(null);
    try {
      const path = await OpenFolder();
      if (path) {
        setBasePath(path);
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

  return (
    <div className="app">
      <header className="app-header">
        <div className="app-title">
          <h1>Notebit</h1>
          <span className="app-subtitle">The Sanctuary</span>
        </div>
        <div className="header-actions">
          <button
            className="open-folder-button"
            onClick={handleOpenFolder}
            disabled={loading}
          >
            <FolderOpen size={18} />
            <span>Open Folder</span>
          </button>
        </div>
      </header>

      {error && (
        <div className="error-banner">
          <span>{error}</span>
          <button onClick={() => setError(null)}>✕</button>
        </div>
      )}

      <main className="app-main">
        <aside className="sidebar">
          <div className="sidebar-header">
            <h2>Files</h2>
            {basePath && <span className="path-hint">{basePath}</span>}
          </div>
          <FileTree
            tree={fileTree}
            onSelect={handleFileSelect}
            selectedPath={currentFile?.path}
          />
        </aside>

        <div className="editor-area">
          {currentFile ? (
            <Editor
              content={currentContent}
              onChange={handleContentChange}
              onSave={handleSave}
              filename={currentFile.name}
            />
          ) : (
            <div className="empty-state">
              <div className="empty-icon">
                <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                  <polyline points="14 2 14 8 20 8"></polyline>
                  <line x1="16" y1="13" x2="8" y2="13"></line>
                  <line x1="16" y1="17" x2="8" y2="17"></line>
                  <polyline points="10 9 9 9 8 9"></polyline>
                </svg>
              </div>
              <h2>Welcome to Notebit</h2>
              <p>Select a file from the sidebar or open a folder to get started</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

export default App;
