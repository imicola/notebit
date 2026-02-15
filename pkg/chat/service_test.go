package chat

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupChatTestService(t *testing.T) (*Service, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "notebit-chat-test-*")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmpDir, "chat.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatal(err)
	}
	svc, err := NewService(db, tmpDir)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatal(err)
	}
	cleanup := func() {
		svc.Close()
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		_ = os.RemoveAll(tmpDir)
	}
	return svc, cleanup
}

func TestSessionMessagePersistence(t *testing.T) {
	svc, cleanup := setupChatTestService(t)
	defer cleanup()

	session, err := svc.CreateSession("测试会话", "qa", []string{"alpha", "beta"})
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	_, err = svc.AppendMessage(session.ID, "user", "你好，世界", nil, nil, "sent")
	if err != nil {
		t.Fatalf("append user message failed: %v", err)
	}

	tokens := 123
	_, err = svc.AppendMessage(session.ID, "assistant", "已收到", []map[string]any{{"path": "a.md"}}, &tokens, "done")
	if err != nil {
		t.Fatalf("append assistant message failed: %v", err)
	}

	msgs, err := svc.ListMessages(session.ID, 1, 50)
	if err != nil {
		t.Fatalf("list messages failed: %v", err)
	}
	if len(msgs.Items) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs.Items))
	}
	if msgs.Items[0].Content != "你好，世界" {
		t.Fatalf("unexpected first message content: %s", msgs.Items[0].Content)
	}
	if msgs.Items[1].TokensUsed == nil || *msgs.Items[1].TokensUsed != 123 {
		t.Fatalf("unexpected tokens used")
	}
}

func TestSessionFiltersAndManagement(t *testing.T) {
	svc, cleanup := setupChatTestService(t)
	defer cleanup()

	s1, _ := svc.CreateSession("工作计划", "work", []string{"project"})
	s2, _ := svc.CreateSession("学习记录", "study", []string{"research"})

	_, _ = svc.AppendMessage(s1.ID, "user", "项目进度", nil, nil, "sent")
	_, _ = svc.AppendMessage(s2.ID, "user", "学习 golang", nil, nil, "sent")

	if err := svc.SetFavorite(s1.ID, true); err != nil {
		t.Fatalf("set favorite failed: %v", err)
	}
	if err := svc.SetArchive(s2.ID, true); err != nil {
		t.Fatalf("set archive failed: %v", err)
	}

	favoriteOnly, err := svc.ListSessions(SessionFilter{FavoritesOnly: true, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list favorites failed: %v", err)
	}
	if len(favoriteOnly.Items) != 1 || favoriteOnly.Items[0].ID != s1.ID {
		t.Fatalf("favorite filter mismatch")
	}

	archivedOnly, err := svc.ListSessions(SessionFilter{ArchivedOnly: true, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list archived failed: %v", err)
	}
	if len(archivedOnly.Items) != 1 || archivedOnly.Items[0].ID != s2.ID {
		t.Fatalf("archived filter mismatch")
	}

	if err := svc.RenameSession(s1.ID, "工作计划-v2"); err != nil {
		t.Fatalf("rename failed: %v", err)
	}
	if err := svc.SetCategory(s1.ID, "ops"); err != nil {
		t.Fatalf("set category failed: %v", err)
	}
	if err := svc.ReplaceTags(s1.ID, []string{"ops", "critical"}); err != nil {
		t.Fatalf("replace tags failed: %v", err)
	}

	updated, err := svc.GetSession(s1.ID)
	if err != nil {
		t.Fatalf("get session failed: %v", err)
	}
	if updated.Title != "工作计划-v2" || updated.Category != "ops" {
		t.Fatalf("session update mismatch")
	}
}

func TestExportAndBackup(t *testing.T) {
	svc, cleanup := setupChatTestService(t)
	defer cleanup()

	session, _ := svc.CreateSession("导出测试", "", nil)
	_, _ = svc.AppendMessage(session.ID, "user", "hello", nil, nil, "sent")
	_, _ = svc.AppendMessage(session.ID, "assistant", "world", nil, nil, "done")

	jsonPath, err := svc.ExportSession(session.ID, "json")
	if err != nil {
		t.Fatalf("export json failed: %v", err)
	}
	if _, statErr := os.Stat(jsonPath); statErr != nil {
		t.Fatalf("json export file missing: %v", statErr)
	}

	txtPath, err := svc.ExportSession(session.ID, "txt")
	if err != nil {
		t.Fatalf("export txt failed: %v", err)
	}
	if _, statErr := os.Stat(txtPath); statErr != nil {
		t.Fatalf("txt export file missing: %v", statErr)
	}

	backupPath, err := svc.BackupNow(nil)
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}
	if _, statErr := os.Stat(backupPath); statErr != nil {
		t.Fatalf("backup file missing: %v", statErr)
	}
}

func TestStorageOptionsReload(t *testing.T) {
	svc, cleanup := setupChatTestService(t)
	defer cleanup()

	err := svc.SetStorageOptions(StorageOptions{
		EncryptAtRest:       false,
		SyncMode:            SyncModeCloud,
		CloudEndpoint:       "https://example.com/sync",
		AutoBackupEnabled:   false,
		BackupIntervalMins:  15,
		PreferredExportType: "txt",
	})
	if err != nil {
		t.Fatalf("set storage options failed: %v", err)
	}

	opts := svc.GetStorageOptions()
	if opts.SyncMode != SyncModeCloud || opts.PreferredExportType != "txt" {
		t.Fatalf("unexpected options: %+v", opts)
	}

	start := time.Now()
	_, err = svc.BackupNow(nil)
	if err != nil {
		t.Fatalf("backup now failed: %v", err)
	}
	if time.Since(start) <= 0 {
		t.Fatalf("backup timing assertion failed")
	}
}
