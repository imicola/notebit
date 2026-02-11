import { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import './Editor.css';

const Editor = ({ content, onChange, onSave, filename }) => {
  const [value, setValue] = useState(content || '');
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [viewMode, setViewMode] = useState('split'); // 'edit', 'preview', 'split'

  useEffect(() => {
    setValue(content || '');
  }, [content]);

  const handleChange = (e) => {
    const newValue = e.target.value;
    setValue(newValue);
    setHasUnsavedChanges(newValue !== content);
    onChange?.(newValue);
  };

  const handleKeyDown = (e) => {
    // Tab key support
    if (e.key === 'Tab') {
      e.preventDefault();
      const start = e.target.selectionStart;
      const end = e.target.selectionEnd;
      const newValue = value.substring(0, start) + '  ' + value.substring(end);
      setValue(newValue);
      onChange?.(newValue);

      // Restore cursor position
      setTimeout(() => {
        e.target.selectionStart = e.target.selectionEnd = start + 2;
      }, 0);
    }

    // Ctrl+S or Cmd+S to save
    if ((e.ctrlKey || e.metaKey) && e.key === 's') {
      e.preventDefault();
      handleSave();
    }
  };

  const handleSave = () => {
    onSave?.(value);
    setHasUnsavedChanges(false);
  };

  return (
    <div className="editor-container">
      <div className="editor-header">
        <div className="editor-title">
          <span className="filename">{filename || 'Untitled'}</span>
          {hasUnsavedChanges && <span className="unsaved-indicator">●</span>}
        </div>
        <div className="editor-actions">
          <div className="view-modes">
            <button
              className={`mode-button ${viewMode === 'edit' ? 'active' : ''}`}
              onClick={() => setViewMode('edit')}
              title="Edit only"
            >
              Edit
            </button>
            <button
              className={`mode-button ${viewMode === 'split' ? 'active' : ''}`}
              onClick={() => setViewMode('split')}
              title="Split view"
            >
              Split
            </button>
            <button
              className={`mode-button ${viewMode === 'preview' ? 'active' : ''}`}
              onClick={() => setViewMode('preview')}
              title="Preview only"
            >
              Preview
            </button>
          </div>
          <button
            className="save-button"
            onClick={handleSave}
            disabled={!hasUnsavedChanges}
            title="Save (Ctrl+S)"
          >
            Save
          </button>
        </div>
      </div>
      <div className={`editor-content view-${viewMode}`}>
        <div className="editor-pane">
          <textarea
            value={value}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            placeholder="Start writing your note..."
            spellCheck={false}
          />
        </div>
        <div className="preview-pane">
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            components={{
              code: ({ node, inline, className, children, ...props }) => {
                const match = /language-(\w+)/.exec(className || '');
                return !inline ? (
                  <pre className={className}>
                    <code className={className} {...props}>
                      {children}
                    </code>
                  </pre>
                ) : (
                  <code className={className} {...props}>
                    {children}
                  </code>
                );
              },
            }}
          >
            {value || '*Start writing to see preview...*'}
          </ReactMarkdown>
        </div>
      </div>
    </div>
  );
};

export default Editor;
