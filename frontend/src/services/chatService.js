import {
  EnsureDefaultChatSession,
  CreateChatSession,
  ListChatSessions,
  ListChatMessages,
  RenameChatSession,
  DeleteChatSession,
  SetChatSessionArchived,
  SetChatSessionFavorite,
  SetChatSessionCategory,
  SetChatSessionTags,
  ExportChatSession,
  BackupChatNow,
  GetChatStorageOptions,
  SetChatStorageOptions,
} from '../../wailsjs/go/main/App';

const wrap = async (op, fn) => {
  try {
    return await fn();
  } catch (error) {
    throw new Error(`${op} failed: ${error?.message || error}`);
  }
};

export const chatService = {
  ensureDefaultSession() {
    return wrap('ensureDefaultSession', EnsureDefaultChatSession);
  },

  createSession(title, category = '', tags = []) {
    return wrap('createSession', () => CreateChatSession(title, category, tags));
  },

  listSessions(filters = {}, page = 1, pageSize = 20) {
    return wrap('listSessions', () => ListChatSessions(
      filters.keyword || '',
      filters.startTS || 0,
      filters.endTS || 0,
      filters.category || '',
      Boolean(filters.archivedOnly),
      Boolean(filters.favoritesOnly),
      filters.tag || '',
      page,
      pageSize,
    ));
  },

  listMessages(sessionId, page = 1, pageSize = 100) {
    return wrap('listMessages', () => ListChatMessages(sessionId, page, pageSize));
  },

  renameSession(sessionId, title) {
    return wrap('renameSession', () => RenameChatSession(sessionId, title));
  },

  deleteSession(sessionId) {
    return wrap('deleteSession', () => DeleteChatSession(sessionId));
  },

  setArchived(sessionId, archived) {
    return wrap('setArchived', () => SetChatSessionArchived(sessionId, archived));
  },

  setFavorite(sessionId, favorite) {
    return wrap('setFavorite', () => SetChatSessionFavorite(sessionId, favorite));
  },

  setCategory(sessionId, category) {
    return wrap('setCategory', () => SetChatSessionCategory(sessionId, category));
  },

  setTags(sessionId, tags = []) {
    return wrap('setTags', () => SetChatSessionTags(sessionId, tags));
  },

  exportSession(sessionId, format = 'json') {
    return wrap('exportSession', () => ExportChatSession(sessionId, format));
  },

  backupNow() {
    return wrap('backupNow', BackupChatNow);
  },

  getStorageOptions() {
    return wrap('getStorageOptions', GetChatStorageOptions);
  },

  setStorageOptions(options) {
    return wrap('setStorageOptions', () => SetChatStorageOptions(
      Boolean(options.encrypt_at_rest),
      options.sync_mode || 'local',
      options.cloud_endpoint || '',
      Boolean(options.auto_backup_enabled),
      Number(options.backup_interval_mins || 30),
      options.preferred_export_type || 'json',
    ));
  },
};

export default chatService;
