package main

import (
	"context"
	"fmt"
	"strings"

	"notebit/pkg/chat"
)

func (a *App) ensureChatService() error {
	if a.chatSvc == nil {
		if a.dbm.IsInitialized() {
			a.initializeChat()
		}
	}
	if a.chatSvc == nil {
		return fmt.Errorf("chat service not initialized")
	}
	return nil
}

func (a *App) EnsureDefaultChatSession() (map[string]interface{}, error) {
	if err := a.ensureChatService(); err != nil {
		return nil, err
	}
	session, err := a.chatSvc.EnsureDefaultSession()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"session": session}, nil
}

func (a *App) CreateChatSession(title, category string, tags []string) (map[string]interface{}, error) {
	if err := a.ensureChatService(); err != nil {
		return nil, err
	}
	session, err := a.chatSvc.CreateSession(title, category, tags)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"session": session}, nil
}

func (a *App) ListChatSessions(keyword string, startTS, endTS int64, category string, archivedOnly, favoritesOnly bool, tag string, page, pageSize int) (map[string]interface{}, error) {
	if err := a.ensureChatService(); err != nil {
		return nil, err
	}
	result, err := a.chatSvc.ListSessions(chat.SessionFilter{
		Keyword:       strings.TrimSpace(keyword),
		StartTS:       startTS,
		EndTS:         endTS,
		Category:      strings.TrimSpace(category),
		ArchivedOnly:  archivedOnly,
		FavoritesOnly: favoritesOnly,
		Tag:           strings.TrimSpace(tag),
		Page:          page,
		PageSize:      pageSize,
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"items": result.Items, "total": result.Total, "page": result.Page, "size": result.Size}, nil
}

func (a *App) ListChatMessages(sessionID string, page, pageSize int) (map[string]interface{}, error) {
	if err := a.ensureChatService(); err != nil {
		return nil, err
	}
	result, err := a.chatSvc.ListMessages(strings.TrimSpace(sessionID), page, pageSize)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"items": result.Items, "total": result.Total, "page": result.Page, "size": result.Size}, nil
}

func (a *App) RenameChatSession(sessionID, title string) error {
	if err := a.ensureChatService(); err != nil {
		return err
	}
	return a.chatSvc.RenameSession(strings.TrimSpace(sessionID), strings.TrimSpace(title))
}

func (a *App) DeleteChatSession(sessionID string) error {
	if err := a.ensureChatService(); err != nil {
		return err
	}
	return a.chatSvc.DeleteSession(strings.TrimSpace(sessionID))
}

func (a *App) SetChatSessionArchived(sessionID string, archived bool) error {
	if err := a.ensureChatService(); err != nil {
		return err
	}
	return a.chatSvc.SetArchive(strings.TrimSpace(sessionID), archived)
}

func (a *App) SetChatSessionFavorite(sessionID string, favorite bool) error {
	if err := a.ensureChatService(); err != nil {
		return err
	}
	return a.chatSvc.SetFavorite(strings.TrimSpace(sessionID), favorite)
}

func (a *App) SetChatSessionCategory(sessionID, category string) error {
	if err := a.ensureChatService(); err != nil {
		return err
	}
	return a.chatSvc.SetCategory(strings.TrimSpace(sessionID), strings.TrimSpace(category))
}

func (a *App) SetChatSessionTags(sessionID string, tags []string) error {
	if err := a.ensureChatService(); err != nil {
		return err
	}
	return a.chatSvc.ReplaceTags(strings.TrimSpace(sessionID), tags)
}

func (a *App) GetChatStorageOptions() (map[string]interface{}, error) {
	if err := a.ensureChatService(); err != nil {
		return nil, err
	}
	opts := a.chatSvc.GetStorageOptions()
	return map[string]interface{}{
		"encrypt_at_rest":       opts.EncryptAtRest,
		"sync_mode":             opts.SyncMode,
		"cloud_endpoint":        opts.CloudEndpoint,
		"auto_backup_enabled":   opts.AutoBackupEnabled,
		"backup_interval_mins":  opts.BackupIntervalMins,
		"preferred_export_type": opts.PreferredExportType,
	}, nil
}

func (a *App) SetChatStorageOptions(encryptAtRest bool, syncMode, cloudEndpoint string, autoBackup bool, backupIntervalMins int, preferredExportType string) error {
	if err := a.ensureChatService(); err != nil {
		return err
	}
	return a.chatSvc.SetStorageOptions(chat.StorageOptions{
		EncryptAtRest:       encryptAtRest,
		SyncMode:            strings.TrimSpace(syncMode),
		CloudEndpoint:       strings.TrimSpace(cloudEndpoint),
		AutoBackupEnabled:   autoBackup,
		BackupIntervalMins:  backupIntervalMins,
		PreferredExportType: strings.TrimSpace(preferredExportType),
	})
}

func (a *App) ExportChatSession(sessionID, format string) (string, error) {
	if err := a.ensureChatService(); err != nil {
		return "", err
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format != "txt" {
		format = "json"
	}
	return a.chatSvc.ExportSession(strings.TrimSpace(sessionID), format)
}

func (a *App) BackupChatNow() (string, error) {
	if err := a.ensureChatService(); err != nil {
		return "", err
	}
	return a.chatSvc.BackupNow(context.Background())
}
