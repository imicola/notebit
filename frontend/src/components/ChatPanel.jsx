import { useState, useEffect, useRef, useCallback } from 'react';
import { Send, Loader2, Sparkles, Plus, Star, Archive, Trash2, Download, Shield, Cloud, HardDrive, Search } from 'lucide-react';
import clsx from 'clsx';
import { ragService } from '../services/ragService';
import { chatService } from '../services/chatService';

const PAGE_SIZE = 20;

export default function ChatPanel() {
  const [messages, setMessages] = useState([]);
  const [sessions, setSessions] = useState([]);
  const [activeSessionId, setActiveSessionId] = useState('');
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [loadingSessions, setLoadingSessions] = useState(false);
  const [ragStatus, setRagStatus] = useState(null);
  const [error, setError] = useState('');

  const [filters, setFilters] = useState({
    keyword: '',
    category: '',
    tag: '',
    favoritesOnly: false,
    archivedOnly: false,
    startTS: 0,
    endTS: 0,
  });

  const [storageOptions, setStorageOptions] = useState({
    encrypt_at_rest: true,
    sync_mode: 'local',
    cloud_endpoint: '',
    auto_backup_enabled: true,
    backup_interval_mins: 30,
    preferred_export_type: 'json',
  });

  const [messagePage, setMessagePage] = useState(1);
  const [messageTotal, setMessageTotal] = useState(0);
  const messagesEndRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, loading]);

  const loadSessions = useCallback(async () => {
    setLoadingSessions(true);
    try {
      const result = await chatService.listSessions(filters, 1, PAGE_SIZE);
      const items = result.items || [];
      setSessions(items);

      if (!activeSessionId && items.length > 0) {
        setActiveSessionId(items[0].id);
      }

      if (items.length === 0) {
        const created = await chatService.ensureDefaultSession();
        const session = created.session;
        if (session) {
          setSessions([session]);
          setActiveSessionId(session.id);
        }
      }
    } catch (e) {
      setError(e.message || '加载会话失败');
    } finally {
      setLoadingSessions(false);
    }
  }, [activeSessionId, filters]);

  const loadMessages = useCallback(async (sessionId, page = 1) => {
    if (!sessionId) return;
    try {
      const result = await chatService.listMessages(sessionId, page, 100);
      setMessages(result.items || []);
      setMessagePage(page);
      setMessageTotal(result.total || 0);
    } catch (e) {
      setError(e.message || '加载消息失败');
    }
  }, []);

  useEffect(() => {
    ragService.getStatus().then(setRagStatus).catch(() => {
      setRagStatus({ available: false, database_ready: false });
    });

    chatService.getStorageOptions().then((opts) => {
      setStorageOptions((prev) => ({ ...prev, ...opts }));
    }).catch(() => {});
  }, []);

  useEffect(() => {
    loadSessions();
  }, [loadSessions]);

  useEffect(() => {
    if (activeSessionId) {
      loadMessages(activeSessionId, 1);
    }
  }, [activeSessionId, loadMessages]);

  const handleCreateSession = async () => {
    try {
      const title = window.prompt('输入会话名称', '新会话');
      if (title === null) return;
      const category = window.prompt('会话分类（可选）', '') || '';
      const tagInput = window.prompt('标签（逗号分隔，可选）', '') || '';
      const tags = tagInput.split(',').map((tag) => tag.trim()).filter(Boolean);
      const res = await chatService.createSession(title, category, tags);
      const session = res.session;
      if (session) {
        setActiveSessionId(session.id);
        await loadSessions();
      }
    } catch (e) {
      setError(e.message || '创建会话失败');
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const trimmed = input.trim();
    if (!trimmed || loading || !activeSessionId) return;

    if (ragStatus && !ragStatus.available) {
      setError('RAG 服务不可用，请先检查 LLM 与数据库状态。');
      return;
    }

    setLoading(true);
    setError('');
    setInput('');

    const optimisticUser = {
      id: `local-${Date.now()}`,
      role: 'user',
      content: trimmed,
      timestamp: Date.now(),
      status: 'sent',
    };
    setMessages((prev) => [...prev, optimisticUser]);

    try {
      const response = await ragService.query(trimmed, activeSessionId);
      const assistant = {
        id: response.message_id || `${Date.now()}-assistant`,
        role: 'assistant',
        content: response.content,
        sources: response.sources,
        tokens_used: response.tokens_used,
        timestamp: Date.now(),
        status: 'done',
      };
      setMessages((prev) => [...prev, assistant]);
      await loadSessions();
      await loadMessages(activeSessionId, messagePage);
    } catch (err) {
      setError(err.message || '查询失败');
      setMessages((prev) => [...prev, {
        id: `${Date.now()}-error`,
        role: 'system',
        content: `Error: ${err.message || 'Query failed'}`,
        timestamp: Date.now(),
        status: 'error',
      }]);
    } finally {
      setLoading(false);
    }
  };

  const handleSourceClick = (source) => {
    window.dispatchEvent(new CustomEvent('open-file', {
      detail: { path: source.path },
    }));
  };

  const updateSessionMeta = async (sessionId, action) => {
    try {
      await action();
      await loadSessions();
      if (activeSessionId === sessionId) {
        await loadMessages(sessionId, messagePage);
      }
    } catch (e) {
      setError(e.message || '会话更新失败');
    }
  };

  const activeSession = sessions.find((s) => s.id === activeSessionId);

  return (
    <div className="flex h-full bg-secondary">
      <aside className="w-80 border-r border-modifier-border flex flex-col">
        <div className="p-3 border-b border-modifier-border space-y-2">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-normal flex items-center gap-2">
              <Sparkles size={16} className="text-obsidian-purple" />
              Chat Sessions
            </h2>
            <button onClick={handleCreateSession} className="p-1 rounded hover:bg-modifier-hover text-muted hover:text-normal" title="新建会话">
              <Plus size={16} />
            </button>
          </div>
          <div className="grid grid-cols-2 gap-2">
            <input
              value={filters.keyword}
              onChange={(e) => setFilters((prev) => ({ ...prev, keyword: e.target.value }))}
              placeholder="关键词"
              className="px-2 py-1 text-xs rounded bg-primary-alt border border-modifier-border"
            />
            <input
              value={filters.category}
              onChange={(e) => setFilters((prev) => ({ ...prev, category: e.target.value }))}
              placeholder="分类"
              className="px-2 py-1 text-xs rounded bg-primary-alt border border-modifier-border"
            />
            <input
              value={filters.tag}
              onChange={(e) => setFilters((prev) => ({ ...prev, tag: e.target.value }))}
              placeholder="标签"
              className="px-2 py-1 text-xs rounded bg-primary-alt border border-modifier-border"
            />
            <button
              onClick={() => loadSessions()}
              className="px-2 py-1 text-xs rounded bg-modifier-hover hover:bg-modifier-border"
            >
              <span className="inline-flex items-center gap-1"><Search size={12} />查询</span>
            </button>
            <input
              type="date"
              onChange={(e) => setFilters((prev) => ({
                ...prev,
                startTS: e.target.value ? new Date(`${e.target.value}T00:00:00`).getTime() : 0,
              }))}
              className="px-2 py-1 text-xs rounded bg-primary-alt border border-modifier-border"
            />
            <input
              type="date"
              onChange={(e) => setFilters((prev) => ({
                ...prev,
                endTS: e.target.value ? new Date(`${e.target.value}T23:59:59`).getTime() : 0,
              }))}
              className="px-2 py-1 text-xs rounded bg-primary-alt border border-modifier-border"
            />
          </div>
          <div className="flex gap-2 text-xs">
            <label className="flex items-center gap-1"><input type="checkbox" checked={filters.favoritesOnly} onChange={(e) => setFilters((p) => ({ ...p, favoritesOnly: e.target.checked }))} />收藏</label>
            <label className="flex items-center gap-1"><input type="checkbox" checked={filters.archivedOnly} onChange={(e) => setFilters((p) => ({ ...p, archivedOnly: e.target.checked }))} />归档</label>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-2 space-y-2">
          {loadingSessions ? (
            <div className="text-xs text-muted p-2">加载会话中...</div>
          ) : sessions.length === 0 ? (
            <div className="text-xs text-muted p-2">暂无会话</div>
          ) : sessions.map((session) => (
            <button
              key={session.id}
              onClick={() => setActiveSessionId(session.id)}
              className={clsx(
                'w-full text-left p-2 rounded border transition-colors',
                activeSessionId === session.id ? 'border-obsidian-purple bg-modifier-hover' : 'border-modifier-border hover:bg-modifier-hover',
              )}
            >
              <div className="flex items-center justify-between gap-1">
                <div className="truncate text-sm font-medium text-normal">{session.title}</div>
                {session.favorite && <Star size={12} className="text-obsidian-yellow" />}
              </div>
              <div className="text-[11px] text-muted truncate">{session.preview || '暂无消息'}</div>
              <div className="text-[10px] text-faint mt-1">
                {session.category || '未分类'} · {session.message_count || 0} 条
              </div>
            </button>
          ))}
        </div>

        <div className="border-t border-modifier-border p-3 space-y-2 text-xs">
          <div className="font-medium text-muted">存储与备份</div>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={storageOptions.encrypt_at_rest}
              onChange={(e) => setStorageOptions((p) => ({ ...p, encrypt_at_rest: e.target.checked }))}
            />
            <Shield size={12} /> 加密存储
          </label>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={storageOptions.auto_backup_enabled}
              onChange={(e) => setStorageOptions((p) => ({ ...p, auto_backup_enabled: e.target.checked }))}
            />
            自动备份
          </label>
          <div className="flex gap-2">
            <button
              onClick={() => setStorageOptions((p) => ({ ...p, sync_mode: 'local' }))}
              className={clsx('px-2 py-1 rounded border text-[11px] flex items-center gap-1', storageOptions.sync_mode === 'local' ? 'border-obsidian-purple' : 'border-modifier-border')}
            >
              <HardDrive size={11} />本地
            </button>
            <button
              onClick={() => setStorageOptions((p) => ({ ...p, sync_mode: 'cloud' }))}
              className={clsx('px-2 py-1 rounded border text-[11px] flex items-center gap-1', storageOptions.sync_mode === 'cloud' ? 'border-obsidian-purple' : 'border-modifier-border')}
            >
              <Cloud size={11} />云同步
            </button>
          </div>
          {storageOptions.sync_mode === 'cloud' && (
            <input
              value={storageOptions.cloud_endpoint}
              onChange={(e) => setStorageOptions((p) => ({ ...p, cloud_endpoint: e.target.value }))}
              placeholder="Cloud endpoint"
              className="w-full px-2 py-1 rounded bg-primary-alt border border-modifier-border"
            />
          )}
          <div className="flex gap-2">
            <input
              type="number"
              min="5"
              value={storageOptions.backup_interval_mins}
              onChange={(e) => setStorageOptions((p) => ({ ...p, backup_interval_mins: Number(e.target.value || 30) }))}
              className="w-20 px-2 py-1 rounded bg-primary-alt border border-modifier-border"
            />
            <span className="text-faint self-center">分钟</span>
          </div>
          <button
            onClick={async () => {
              try {
                await chatService.setStorageOptions(storageOptions);
                await loadSessions();
              } catch (e) {
                setError(e.message || '保存存储选项失败');
              }
            }}
            className="w-full px-2 py-1 rounded bg-modifier-hover hover:bg-modifier-border"
          >
            保存存储配置
          </button>
        </div>
      </aside>

      <section className="flex-1 flex flex-col">
        <div className="px-4 py-3 bg-secondary border-b border-modifier-border flex justify-between items-center">
          <div className="flex items-center gap-2">
            <Sparkles className="text-obsidian-purple" size={18} />
            <h2 className="text-sm font-semibold text-normal">{activeSession?.title || 'Knowledge Chat'}</h2>
          </div>
          <div className="flex items-center gap-2 text-xs">
            {activeSession && (
              <>
                <button className="p-1 rounded hover:bg-modifier-hover" title="重命名" onClick={() => {
                  const title = window.prompt('新会话名称', activeSession.title);
                  if (title !== null) {
                    updateSessionMeta(activeSession.id, () => chatService.renameSession(activeSession.id, title));
                  }
                }}>重命名</button>
                <button className="p-1 rounded hover:bg-modifier-hover" title="设置标签" onClick={() => {
                  const current = (activeSession.tags || []).join(',');
                  const value = window.prompt('标签（逗号分隔）', current);
                  if (value !== null) {
                    const tags = value.split(',').map((item) => item.trim()).filter(Boolean);
                    updateSessionMeta(activeSession.id, () => chatService.setTags(activeSession.id, tags));
                  }
                }}>标签</button>
                <button className="p-1 rounded hover:bg-modifier-hover" title="收藏" onClick={() => updateSessionMeta(activeSession.id, () => chatService.setFavorite(activeSession.id, !activeSession.favorite))}>
                  <Star size={14} className={activeSession.favorite ? 'text-obsidian-yellow' : 'text-muted'} />
                </button>
                <button className="p-1 rounded hover:bg-modifier-hover" title="归档" onClick={() => updateSessionMeta(activeSession.id, () => chatService.setArchived(activeSession.id, !activeSession.archived))}>
                  <Archive size={14} className={activeSession.archived ? 'text-obsidian-cyan' : 'text-muted'} />
                </button>
                <button className="p-1 rounded hover:bg-modifier-hover" title="删除" onClick={async () => {
                  if (!window.confirm('确认删除该会话？')) return;
                  await updateSessionMeta(activeSession.id, () => chatService.deleteSession(activeSession.id));
                  const next = sessions.find((s) => s.id !== activeSession.id);
                  setActiveSessionId(next?.id || '');
                }}>
                  <Trash2 size={14} className="text-muted" />
                </button>
                <button className="p-1 rounded hover:bg-modifier-hover" title="导出 JSON" onClick={async () => {
                  try {
                    const path = await chatService.exportSession(activeSession.id, 'json');
                    setError(`导出成功: ${path}`);
                  } catch (e) {
                    setError(e.message || '导出失败');
                  }
                }}>
                  <Download size={14} className="text-muted" />
                </button>
                <button className="px-2 py-1 rounded hover:bg-modifier-hover" title="导出 TXT" onClick={async () => {
                  try {
                    const path = await chatService.exportSession(activeSession.id, 'txt');
                    setError(`导出成功: ${path}`);
                  } catch (e) {
                    setError(e.message || '导出失败');
                  }
                }}>TXT</button>
                <button className="px-2 py-1 rounded hover:bg-modifier-hover" onClick={async () => {
                  try {
                    const path = await chatService.backupNow();
                    setError(`备份成功: ${path}`);
                  } catch (e) {
                    setError(e.message || '备份失败');
                  }
                }}>备份</button>
              </>
            )}
            {ragStatus && !ragStatus.available && <div className="text-orange-500">RAG unavailable</div>}
          </div>
        </div>

        {error && <div className="px-4 py-2 text-xs text-obsidian-orange border-b border-modifier-border">{error}</div>}

        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {messages.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-center">
              <Sparkles className="text-faint mb-4" size={48} />
              <h3 className="text-lg font-medium text-muted mb-2">Ask Your Notes</h3>
              <p className="text-sm text-muted">选择或新建会话后开始提问，消息会实时保存到本地数据库。</p>
            </div>
          ) : (
            messages.map((msg) => (
              <div key={msg.id} className={clsx('flex gap-3', msg.role === 'user' ? 'flex-row-reverse' : '')}>
                <div className={clsx(
                  'w-8 h-8 rounded-full flex items-center justify-center shrink-0',
                  msg.role === 'user' ? 'bg-obsidian-purple text-white' : 'bg-obsidian-purple/10 text-obsidian-purple',
                )}>
                  {msg.role === 'user' ? 'U' : <Sparkles size={16} />}
                </div>

                <div className={clsx(
                  'max-w-[80%] rounded-lg px-4 py-2',
                  msg.role === 'user' ? 'bg-obsidian-purple text-white' : 'bg-modifier-hover text-normal',
                )}>
                  <p className="text-sm whitespace-pre-wrap break-words">{msg.content}</p>

                  {msg.sources && msg.sources.length > 0 && (
                    <div className="mt-3 pt-3 border-t border-modifier-border">
                      <p className="text-xs font-medium text-muted mb-2">Sources:</p>
                      <div className="space-y-1">
                        {msg.sources.map((source, idx) => (
                          <button
                            key={`${source.chunk_id || idx}`}
                            onClick={() => handleSourceClick(source)}
                            className="flex items-start gap-2 p-2 rounded hover:bg-modifier-hover group transition-colors text-left"
                          >
                            <Sparkles size={14} className="text-faint group-hover:text-obsidian-purple shrink-0 mt-0.5" />
                            <div className="flex-1 min-w-0">
                              <div className="text-sm font-medium text-normal truncate">{source.title}</div>
                              {source.heading && <div className="text-xs text-muted truncate">{source.heading}</div>}
                              <div className="text-xs text-faint">{Math.round((source.similarity || 0) * 100)}% relevance</div>
                            </div>
                          </button>
                        ))}
                      </div>
                    </div>
                  )}

                  {msg.tokens_used && <div className="mt-1 text-xs text-faint">{msg.tokens_used} tokens</div>}
                </div>
              </div>
            ))
          )}
          <div ref={messagesEndRef} />
        </div>

        {messageTotal > messages.length && (
          <div className="px-4 pb-2">
            <button
              onClick={() => loadMessages(activeSessionId, messagePage + 1)}
              className="text-xs px-3 py-1 rounded bg-modifier-hover hover:bg-modifier-border"
            >
              加载更多历史
            </button>
          </div>
        )}

        {loading && (
          <div className="flex gap-3 px-4 py-3 bg-secondary border-b border-modifier-border">
            <div className="w-8 h-8 rounded-full flex items-center justify-center bg-secondary">
              <Loader2 size={16} className="animate-spin text-muted" />
            </div>
            <div className="text-sm text-muted">Thinking...</div>
          </div>
        )}

        <form onSubmit={handleSubmit} className="p-4 border-t border-modifier-border">
          <div className="flex gap-2">
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault();
                  handleSubmit(e);
                }
              }}
              placeholder="Ask a question about your notes..."
              disabled={loading || !activeSessionId}
              className="flex-1 px-4 py-2 rounded-lg bg-primary-alt border border-modifier-border text-normal focus:border-obsidian-purple focus:outline-none"
            />
            <button
              type="submit"
              disabled={loading || !input.trim() || !activeSessionId}
              className="px-4 py-2 bg-obsidian-purple hover:bg-obsidian-purple-hover disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
            >
              <Send size={18} />
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}
