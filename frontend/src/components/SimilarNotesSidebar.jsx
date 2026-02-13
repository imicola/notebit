import { useState, useEffect, useRef } from 'react';
import { X, Sparkles, AlertCircle, FileText, Loader2 } from 'lucide-react';
import { similarityService } from '../services/similarityService';
import clsx from 'clsx';
import { SEMANTIC_SEARCH } from '../constants';

const SimilarNotesSidebar = ({
  query,          // Content to search for similar notes
  searchRequest,  // Snapshot generated on save to trigger search
  basePath,
  currentPath,
  isOpen,
  onClose,
  onNoteClick,
  width
}) => {
  const [status, setStatus] = useState(null);
  const [notes, setNotes] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const debounceRef = useRef(null);
  const requestIdRef = useRef(0);
  const mountedRef = useRef(false);

  // Check availability on mount
  useEffect(() => {
    mountedRef.current = true;
    checkStatus();
    return () => {
      mountedRef.current = false;
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, []);

  const checkStatus = async () => {
    try {
      const result = await similarityService.getStatus();
      setStatus(result);
    } catch (err) {
      console.error('Failed to check status:', err);
      setStatus({ available: false, db_initialized: false });
    }
  };

  useEffect(() => {
    if (isOpen) {
      checkStatus();
    }
  }, [isOpen, basePath]);

  // Search only when a new save request arrives
  useEffect(() => {
    const trimmedQuery = (searchRequest?.content || '').trim();
    if (!trimmedQuery || !status?.available || !searchRequest?.id) {
      return;
    }

    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    debounceRef.current = setTimeout(() => {
      const currentRequestId = requestIdRef.current + 1;
      requestIdRef.current = currentRequestId;
      setLoading(true);
      setError(null);

      similarityService.findSimilar(trimmedQuery, SEMANTIC_SEARCH.DEFAULT_LIMIT)
        .then((results) => {
          if (!mountedRef.current || requestIdRef.current !== currentRequestId) {
            return;
          }
          const filtered = results.filter((note) => (
            note.similarity >= SEMANTIC_SEARCH.MIN_SIMILARITY
            && (!currentPath || note.path !== currentPath)
          ));
          setNotes(filtered);
        })
        .catch((err) => {
          if (!mountedRef.current || requestIdRef.current !== currentRequestId) {
            return;
          }
          setError(err.message || 'Search failed');
        })
        .finally(() => {
          if (!mountedRef.current || requestIdRef.current !== currentRequestId) {
            return;
          }
          setLoading(false);
        });
    }, SEMANTIC_SEARCH.DEBOUNCE_MS);
  }, [searchRequest, status?.available]);

  // Not available state
  if (status && !status.available) {
    return (
      <aside
        className={clsx(
          "flex flex-col bg-secondary border-l border-modifier-border shrink-0 transition-all duration-300",
          !isOpen && "hidden"
        )}
        style={{ width: isOpen ? width : 0 }}
      >
        <div className="px-4 py-3 bg-secondary border-b border-modifier-border">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Sparkles className="text-muted" size={14} />
              <h2 className="text-xs font-bold uppercase tracking-wider text-muted">Related Notes</h2>
            </div>
            <button onClick={onClose} className="text-muted hover:text-normal">
              <X size={16} />
            </button>
          </div>
        </div>
        <div className="flex-1 flex flex-col items-center justify-center p-6 text-center">
          <AlertCircle className="text-faint mb-3" size={32} />
          <p className="text-sm text-muted mb-2">Semantic search unavailable</p>
          <p className="text-xs text-faint">
            {!status.db_initialized && "Open a folder to enable"}
            {status.db_initialized && !status.ai_healthy && "Configure AI in Settings"}
          </p>
        </div>
      </aside>
    );
  }

  return (
    <aside
      className={clsx(
        "flex flex-col bg-secondary border-l border-modifier-border shrink-0 transition-all duration-300",
        !isOpen && "hidden"
      )}
      style={{ width: isOpen ? width : 0 }}
    >
      <div className="px-4 py-3 bg-secondary border-b border-modifier-border">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Sparkles className="text-obsidian-purple" size={16} />
            <h2 className="text-xs font-bold uppercase tracking-wider text-normal">Related Notes</h2>
          </div>
          <button onClick={onClose} className="text-muted hover:text-normal">
            <X size={16} />
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {loading && (
          <div className="flex items-center justify-center p-8">
            <Loader2 className="animate-spin text-muted" size={24} />
          </div>
        )}

        {error && (
          <div className="p-4">
            <div className="flex items-center gap-2 text-obsidian-orange text-sm">
              <AlertCircle size={16} />
              <span>{error}</span>
            </div>
          </div>
        )}

        {!loading && !error && notes.length === 0 && query && (
          <div className="p-6 text-center text-faint text-sm">
            No similar notes found
          </div>
        )}

        {!loading && !error && notes.map((note, index) => (
          <div
            key={`${note.path}-${note.chunk_id}-${index}`}
            onClick={() => onNoteClick(note)}
            className="px-4 py-3 hover:bg-modifier-hover cursor-pointer border-b border-modifier-border/50"
          >
            <div className="flex items-start gap-2 mb-1">
              <FileText size={14} className="text-muted shrink-0 mt-0.5" />
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium text-normal truncate">{note.title}</div>
                {note.heading && (
                  <div className="text-xs text-muted truncate">{note.heading}</div>
                )}
              </div>
              <div className="text-xs text-faint ml-2">
                {Math.round(note.similarity * 100)}%
              </div>
            </div>
            <div className="text-xs text-faint line-clamp-2 ml-6">
              {note.content}
            </div>
          </div>
        ))}
      </div>
    </aside>
  );
};

export default SimilarNotesSidebar;
