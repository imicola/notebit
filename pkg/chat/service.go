package chat

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	SyncModeLocal = "local"
	SyncModeCloud = "cloud"
)

type StorageOptions struct {
	EncryptAtRest       bool   `json:"encrypt_at_rest"`
	SyncMode            string `json:"sync_mode"`
	CloudEndpoint       string `json:"cloud_endpoint"`
	AutoBackupEnabled   bool   `json:"auto_backup_enabled"`
	BackupIntervalMins  int    `json:"backup_interval_mins"`
	PreferredExportType string `json:"preferred_export_type"`
}

type SessionFilter struct {
	Keyword       string
	StartTS       int64
	EndTS         int64
	Category      string
	ArchivedOnly  bool
	FavoritesOnly bool
	Tag           string
	Page          int
	PageSize      int
}

type SessionListItem struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Category      string   `json:"category"`
	Archived      bool     `json:"archived"`
	Favorite      bool     `json:"favorite"`
	Tags          []string `json:"tags"`
	CreatedAt     int64    `json:"created_at"`
	UpdatedAt     int64    `json:"updated_at"`
	LastMessageAt int64    `json:"last_message_at"`
	MessageCount  int64    `json:"message_count"`
	Preview       string   `json:"preview"`
}

type SessionListResult struct {
	Items []SessionListItem `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}

type MessageDTO struct {
	ID         string           `json:"id"`
	SessionID  string           `json:"session_id"`
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	Sources    []map[string]any `json:"sources,omitempty"`
	TokensUsed *int             `json:"tokens_used,omitempty"`
	Status     string           `json:"status"`
	Timestamp  int64            `json:"timestamp"`
}

type MessageListResult struct {
	Items []MessageDTO `json:"items"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
}

type Service struct {
	db       *gorm.DB
	basePath string
	mu       sync.RWMutex
	options  StorageOptions
	key      []byte
	stopCh   chan struct{}
	doneCh   chan struct{}
}

func NewService(db *gorm.DB, basePath string) (*Service, error) {
	if db == nil {
		return nil, fmt.Errorf("database is nil")
	}

	s := &Service{db: db, basePath: basePath}
	s.options = StorageOptions{
		EncryptAtRest:       true,
		SyncMode:            SyncModeLocal,
		AutoBackupEnabled:   true,
		BackupIntervalMins:  30,
		PreferredExportType: "json",
	}

	if err := s.autoMigrate(); err != nil {
		return nil, err
	}
	if err := s.loadOptions(); err != nil {
		return nil, err
	}
	s.key = s.deriveKey()
	s.startBackupTicker()
	return s, nil
}

func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopCh != nil {
		close(s.stopCh)
		<-s.doneCh
		s.stopCh = nil
		s.doneCh = nil
	}
}

func (s *Service) autoMigrate() error {
	if err := s.db.AutoMigrate(&Session{}, &Message{}, &SessionTag{}, &Setting{}); err != nil {
		return err
	}
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_chat_sessions_last_message_at ON chat_sessions(last_message_at)",
		"CREATE INDEX IF NOT EXISTS idx_chat_messages_session_time ON chat_messages(session_id, timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_chat_messages_timestamp ON chat_messages(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_chat_session_tags_tag ON chat_session_tags(tag)",
	}
	for _, stmt := range indexes {
		if err := s.db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) loadOptions() error {
	var settings []Setting
	if err := s.db.Where("scope = ?", "chat.storage").Find(&settings).Error; err != nil {
		return err
	}
	for _, item := range settings {
		switch item.Key {
		case "encrypt_at_rest":
			s.options.EncryptAtRest = item.Value == "true"
		case "sync_mode":
			s.options.SyncMode = item.Value
		case "cloud_endpoint":
			s.options.CloudEndpoint = item.Value
		case "auto_backup_enabled":
			s.options.AutoBackupEnabled = item.Value == "true"
		case "backup_interval_mins":
			var minutes int
			_, _ = fmt.Sscanf(item.Value, "%d", &minutes)
			if minutes > 0 {
				s.options.BackupIntervalMins = minutes
			}
		case "preferred_export_type":
			if item.Value != "" {
				s.options.PreferredExportType = item.Value
			}
		}
	}
	if s.options.SyncMode == "" {
		s.options.SyncMode = SyncModeLocal
	}
	if s.options.BackupIntervalMins <= 0 {
		s.options.BackupIntervalMins = 30
	}
	return nil
}

func (s *Service) persistOption(key, value string) error {
	setting := Setting{Scope: "chat.storage", Key: key, Value: value}
	return s.db.Where("scope = ? AND key = ?", setting.Scope, setting.Key).Assign(setting).FirstOrCreate(&setting).Error
}

func (s *Service) deriveKey() []byte {
	host, _ := os.Hostname()
	material := fmt.Sprintf("notebit-chat:%s:%s", s.basePath, host)
	sum := sha256.Sum256([]byte(material))
	key := make([]byte, 32)
	copy(key, sum[:])
	return key
}

func (s *Service) encryptText(plain string) (string, bool, error) {
	if !s.options.EncryptAtRest {
		return plain, false, nil
	}
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", false, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", false, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", false, err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(plain), nil)
	payload := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(payload), true, nil
}

func (s *Service) decryptText(content string, encrypted bool) (string, error) {
	if !encrypted {
		return content, nil
	}
	payload, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < gcm.NonceSize() {
		return "", fmt.Errorf("invalid encrypted payload")
	}
	nonce := payload[:gcm.NonceSize()]
	ciphertext := payload[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (s *Service) GetStorageOptions() StorageOptions {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.options
}

func (s *Service) SetStorageOptions(opts StorageOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if opts.SyncMode == "" {
		opts.SyncMode = SyncModeLocal
	}
	if opts.BackupIntervalMins <= 0 {
		opts.BackupIntervalMins = 30
	}
	s.options = opts
	if err := s.persistOption("encrypt_at_rest", fmt.Sprintf("%t", opts.EncryptAtRest)); err != nil {
		return err
	}
	if err := s.persistOption("sync_mode", opts.SyncMode); err != nil {
		return err
	}
	if err := s.persistOption("cloud_endpoint", opts.CloudEndpoint); err != nil {
		return err
	}
	if err := s.persistOption("auto_backup_enabled", fmt.Sprintf("%t", opts.AutoBackupEnabled)); err != nil {
		return err
	}
	if err := s.persistOption("backup_interval_mins", fmt.Sprintf("%d", opts.BackupIntervalMins)); err != nil {
		return err
	}
	if err := s.persistOption("preferred_export_type", opts.PreferredExportType); err != nil {
		return err
	}
	s.startBackupTicker()
	return nil
}

func (s *Service) CreateSession(title, category string, tags []string) (*SessionListItem, error) {
	now := time.Now().UnixMilli()
	if strings.TrimSpace(title) == "" {
		title = "新会话"
	}
	session := Session{
		ID:            uuid.NewString(),
		Title:         strings.TrimSpace(title),
		Category:      strings.TrimSpace(category),
		CreatedAtUnix: now,
		UpdatedAtUnix: now,
		LastMessageAt: now,
	}
	if err := s.db.Create(&session).Error; err != nil {
		return nil, err
	}
	if err := s.ReplaceTags(session.ID, tags); err != nil {
		return nil, err
	}
	item, err := s.GetSession(session.ID)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) EnsureDefaultSession() (*SessionListItem, error) {
	var session Session
	err := s.db.Order("updated_at_unix DESC").First(&session).Error
	if err == nil {
		return s.GetSession(session.ID)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return s.CreateSession("默认会话", "", nil)
}

func (s *Service) GetSession(sessionID string) (*SessionListItem, error) {
	var session Session
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		return nil, err
	}
	tags, _ := s.getTags(session.ID)
	var count int64
	_ = s.db.Model(&Message{}).Where("session_id = ?", session.ID).Count(&count).Error
	preview, _ := s.getSessionPreview(session.ID)
	return &SessionListItem{
		ID:            session.ID,
		Title:         session.Title,
		Category:      session.Category,
		Archived:      session.Archived,
		Favorite:      session.Favorite,
		Tags:          tags,
		CreatedAt:     session.CreatedAtUnix,
		UpdatedAt:     session.UpdatedAtUnix,
		LastMessageAt: session.LastMessageAt,
		MessageCount:  count,
		Preview:       preview,
	}, nil
}

func (s *Service) ListSessions(filter SessionFilter) (*SessionListResult, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	q := s.db.Model(&Session{})
	if filter.Category != "" {
		q = q.Where("category = ?", filter.Category)
	}
	if filter.ArchivedOnly {
		q = q.Where("archived = ?", true)
	}
	if filter.FavoritesOnly {
		q = q.Where("favorite = ?", true)
	}
	if filter.StartTS > 0 {
		q = q.Where("last_message_at >= ?", filter.StartTS)
	}
	if filter.EndTS > 0 {
		q = q.Where("last_message_at <= ?", filter.EndTS)
	}
	if filter.Tag != "" {
		q = q.Joins("JOIN chat_session_tags ON chat_session_tags.session_id = chat_sessions.id").Where("chat_session_tags.tag = ?", filter.Tag)
	}

	countQuery := q
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	var sessions []Session
	if err := q.Order("last_message_at DESC").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&sessions).Error; err != nil {
		return nil, err
	}

	items := make([]SessionListItem, 0, len(sessions))
	for _, session := range sessions {
		tags, _ := s.getTags(session.ID)
		var count int64
		_ = s.db.Model(&Message{}).Where("session_id = ?", session.ID).Count(&count).Error
		preview, _ := s.getSessionPreview(session.ID)
		item := SessionListItem{
			ID:            session.ID,
			Title:         session.Title,
			Category:      session.Category,
			Archived:      session.Archived,
			Favorite:      session.Favorite,
			Tags:          tags,
			CreatedAt:     session.CreatedAtUnix,
			UpdatedAt:     session.UpdatedAtUnix,
			LastMessageAt: session.LastMessageAt,
			MessageCount:  count,
			Preview:       preview,
		}
		items = append(items, item)
	}

	if strings.TrimSpace(filter.Keyword) != "" {
		kw := strings.ToLower(strings.TrimSpace(filter.Keyword))
		filtered := make([]SessionListItem, 0, len(items))
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Title), kw) || strings.Contains(strings.ToLower(item.Preview), kw) || s.sessionContainsKeyword(item.ID, kw) {
				filtered = append(filtered, item)
			}
		}
		total = int64(len(filtered))
		items = filtered
	}

	return &SessionListResult{Items: items, Total: total, Page: filter.Page, Size: filter.PageSize}, nil
}

func (s *Service) sessionContainsKeyword(sessionID, keyword string) bool {
	var messages []Message
	if err := s.db.Where("session_id = ?", sessionID).Order("timestamp DESC").Limit(30).Find(&messages).Error; err != nil {
		return false
	}
	for _, msg := range messages {
		text, err := s.decryptText(msg.Content, msg.Encrypted)
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(text), keyword) {
			return true
		}
	}
	return false
}

func (s *Service) ListMessages(sessionID string, page, pageSize int) (*MessageListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	var rows []Message
	q := s.db.Model(&Message{}).Where("session_id = ?", sessionID)
	if err := q.Order("timestamp ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	items := make([]MessageDTO, 0, len(rows))
	for _, row := range rows {
		text, err := s.decryptText(row.Content, row.Encrypted)
		if err != nil {
			continue
		}
		var sources []map[string]any
		if row.Sources != "" {
			srcText, decErr := s.decryptText(row.Sources, row.SourcesEncrypted)
			if decErr == nil {
				_ = json.Unmarshal([]byte(srcText), &sources)
			}
		}
		items = append(items, MessageDTO{
			ID:         row.ID,
			SessionID:  row.SessionID,
			Role:       row.Role,
			Content:    text,
			Sources:    sources,
			TokensUsed: row.TokensUsed,
			Status:     row.Status,
			Timestamp:  row.Timestamp,
		})
	}
	return &MessageListResult{Items: items, Total: total, Page: page, Size: pageSize}, nil
}

func (s *Service) AppendMessage(sessionID, role, content string, sources any, tokensUsed *int, status string) (*MessageDTO, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("session id is required")
	}
	now := time.Now().UnixMilli()
	encContent, encrypted, err := s.encryptText(content)
	if err != nil {
		return nil, err
	}
	message := Message{
		ID:         uuid.NewString(),
		SessionID:  sessionID,
		Role:       role,
		Content:    encContent,
		Encrypted:  encrypted,
		Status:     status,
		Timestamp:  now,
		TokensUsed: tokensUsed,
	}
	if sources != nil {
		payload, _ := json.Marshal(sources)
		encSrc, srcEncrypted, srcErr := s.encryptText(string(payload))
		if srcErr == nil {
			message.Sources = encSrc
			message.SourcesEncrypted = srcEncrypted
		}
	}
	if err := s.db.Create(&message).Error; err != nil {
		return nil, err
	}
	_ = s.db.Model(&Session{}).Where("id = ?", sessionID).Updates(map[string]any{
		"last_message_at": now,
		"updated_at_unix": now,
	}).Error
	var srcList []map[string]any
	if sources != nil {
		payload, _ := json.Marshal(sources)
		_ = json.Unmarshal(payload, &srcList)
	}
	return &MessageDTO{
		ID:         message.ID,
		SessionID:  message.SessionID,
		Role:       role,
		Content:    content,
		Sources:    srcList,
		TokensUsed: tokensUsed,
		Status:     status,
		Timestamp:  now,
	}, nil
}

func (s *Service) RenameSession(sessionID, title string) error {
	return s.db.Model(&Session{}).Where("id = ?", sessionID).Updates(map[string]any{
		"title":           strings.TrimSpace(title),
		"updated_at_unix": time.Now().UnixMilli(),
	}).Error
}

func (s *Service) DeleteSession(sessionID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", sessionID).Delete(&SessionTag{}).Error; err != nil {
			return err
		}
		if err := tx.Where("session_id = ?", sessionID).Delete(&Message{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", sessionID).Delete(&Session{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *Service) SetArchive(sessionID string, archived bool) error {
	return s.db.Model(&Session{}).Where("id = ?", sessionID).Updates(map[string]any{
		"archived":        archived,
		"updated_at_unix": time.Now().UnixMilli(),
	}).Error
}

func (s *Service) SetFavorite(sessionID string, favorite bool) error {
	return s.db.Model(&Session{}).Where("id = ?", sessionID).Updates(map[string]any{
		"favorite":        favorite,
		"updated_at_unix": time.Now().UnixMilli(),
	}).Error
}

func (s *Service) SetCategory(sessionID, category string) error {
	return s.db.Model(&Session{}).Where("id = ?", sessionID).Updates(map[string]any{
		"category":        strings.TrimSpace(category),
		"updated_at_unix": time.Now().UnixMilli(),
	}).Error
}

func (s *Service) ReplaceTags(sessionID string, tags []string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", sessionID).Delete(&SessionTag{}).Error; err != nil {
			return err
		}
		uniq := make(map[string]struct{})
		for _, tag := range tags {
			tag = strings.TrimSpace(strings.ToLower(tag))
			if tag == "" {
				continue
			}
			uniq[tag] = struct{}{}
		}
		tagList := make([]string, 0, len(uniq))
		for tag := range uniq {
			tagList = append(tagList, tag)
		}
		sort.Strings(tagList)
		for _, tag := range tagList {
			if err := tx.Create(&SessionTag{SessionID: sessionID, Tag: tag}).Error; err != nil {
				return err
			}
		}
		return tx.Model(&Session{}).Where("id = ?", sessionID).Update("updated_at_unix", time.Now().UnixMilli()).Error
	})
}

func (s *Service) getTags(sessionID string) ([]string, error) {
	var rows []SessionTag
	if err := s.db.Where("session_id = ?", sessionID).Order("tag ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	tags := make([]string, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, row.Tag)
	}
	return tags, nil
}

func (s *Service) getSessionPreview(sessionID string) (string, error) {
	var row Message
	if err := s.db.Where("session_id = ?", sessionID).Order("timestamp DESC").First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	text, err := s.decryptText(row.Content, row.Encrypted)
	if err != nil {
		return "", err
	}
	text = strings.TrimSpace(text)
	r := []rune(text)
	if len(r) > 120 {
		return string(r[:120]) + "...", nil
	}
	return text, nil
}

func (s *Service) ExportSession(sessionID, format string) (string, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return "", err
	}
	messages, err := s.ListMessages(sessionID, 1, 5000)
	if err != nil {
		return "", err
	}
	exportDir := filepath.Join(s.basePath, "data", "chat_exports")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", err
	}
	ts := time.Now().Format("20060102_150405")
	base := sanitizeFilename(session.Title)
	if base == "" {
		base = sessionID
	}
	if format == "txt" {
		path := filepath.Join(exportDir, fmt.Sprintf("%s_%s.txt", base, ts))
		var sb strings.Builder
		sb.WriteString("Notebit Chat Export\n")
		sb.WriteString(fmt.Sprintf("Session: %s\n", session.Title))
		sb.WriteString(fmt.Sprintf("Category: %s\n\n", session.Category))
		for _, m := range messages.Items {
			sb.WriteString(fmt.Sprintf("[%s] %s\n", strings.ToUpper(m.Role), time.UnixMilli(m.Timestamp).Format(time.RFC3339)))
			sb.WriteString(m.Content)
			sb.WriteString("\n\n")
		}
		if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
			return "", err
		}
		return path, nil
	}

	path := filepath.Join(exportDir, fmt.Sprintf("%s_%s.json", base, ts))
	payload := map[string]any{
		"session":  session,
		"messages": messages.Items,
		"exported": time.Now().UnixMilli(),
	}
	b, _ := json.MarshalIndent(payload, "", "  ")
	if err := os.WriteFile(path, b, 0644); err != nil {
		return "", err
	}
	return path, nil
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	return replacer.Replace(name)
}

func (s *Service) BackupNow(ctx context.Context) (string, error) {
	result, err := s.ListSessions(SessionFilter{Page: 1, PageSize: 500})
	if err != nil {
		return "", err
	}
	type sessionDump struct {
		Session  SessionListItem `json:"session"`
		Messages []MessageDTO    `json:"messages"`
	}
	dump := make([]sessionDump, 0, len(result.Items))
	for _, item := range result.Items {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
			}
		}
		messages, msgErr := s.ListMessages(item.ID, 1, 5000)
		if msgErr != nil {
			continue
		}
		dump = append(dump, sessionDump{Session: item, Messages: messages.Items})
	}

	backupDir := filepath.Join(s.basePath, "data", "chat_backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}
	filePath := filepath.Join(backupDir, fmt.Sprintf("chat_backup_%s.json", time.Now().Format("20060102_150405")))
	b, _ := json.MarshalIndent(map[string]any{
		"created_at": time.Now().UnixMilli(),
		"sync_mode":  s.options.SyncMode,
		"sessions":   dump,
	}, "", "  ")
	if err := os.WriteFile(filePath, b, 0644); err != nil {
		return "", err
	}
	return filePath, nil
}

func (s *Service) startBackupTicker() {
	if s.stopCh != nil {
		close(s.stopCh)
		<-s.doneCh
		s.stopCh = nil
		s.doneCh = nil
	}
	if !s.options.AutoBackupEnabled {
		return
	}
	interval := time.Duration(s.options.BackupIntervalMins) * time.Minute
	if interval <= 0 {
		interval = 30 * time.Minute
	}
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	go func(stop <-chan struct{}, done chan<- struct{}) {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_, _ = s.BackupNow(context.Background())
			case <-stop:
				return
			}
		}
	}(s.stopCh, s.doneCh)
}
