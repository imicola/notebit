package main

import (
	"context"
	"fmt"
	"notebit/pkg/database"
	"notebit/pkg/files"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
	fm  *files.Manager
	dbm *database.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		fm:  files.NewManager(),
		dbm: database.GetInstance(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// OpenFolder opens a directory dialog and sets the base path
func (a *App) OpenFolder() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Notes Folder",
	})
	if err != nil {
		return "", err
	}

	if dir == "" {
		return "", nil
	}

	if err := a.fm.SetBasePath(dir); err != nil {
		return "", err
	}

	// Initialize database
	if err := a.dbm.Init(dir); err != nil {
		// Don't fail if database initialization fails - log but continue
		fmt.Printf("Warning: database initialization failed: %v\n", err)
	}

	return dir, nil
}

// SetFolder sets the base path without opening a dialog
func (a *App) SetFolder(path string) error {
	if err := a.fm.SetBasePath(path); err != nil {
		return err
	}

	// Initialize database
	if err := a.dbm.Init(path); err != nil {
		// Don't fail if database initialization fails
		fmt.Printf("Warning: database initialization failed: %v\n", err)
	}

	return nil
}

// ListFiles returns the file tree structure
func (a *App) ListFiles() (*files.FileNode, error) {
	return a.fm.ListFiles()
}

// ReadFile reads the content of a markdown file
func (a *App) ReadFile(path string) (*files.NoteContent, error) {
	return a.fm.ReadFile(path)
}

// SaveFile saves content to a markdown file
func (a *App) SaveFile(path, content string) error {
	err := a.fm.SaveFile(path, content)
	if err != nil {
		return err
	}

	// Index the file in database after saving (pass content to avoid re-reading)
	if a.dbm.IsInitialized() {
		go a.indexFileContent(path, content)
	}

	return nil
}

// CreateFile creates a new markdown file
func (a *App) CreateFile(path, content string) error {
	err := a.fm.CreateFile(path, content)
	if err != nil {
		return err
	}

	// Index the file in database after creating (pass content to avoid re-reading)
	if a.dbm.IsInitialized() {
		go a.indexFileContent(path, content)
	}

	return nil
}

// DeleteFile deletes a markdown file or directory
func (a *App) DeleteFile(path string) error {
	err := a.fm.DeleteFile(path)
	if err != nil {
		return err
	}

	// Remove from database index
	if a.dbm.IsInitialized() {
		repo := a.dbm.Repository()
		_ = repo.DeleteFile(path)
	}

	return nil
}

// RenameFile renames a file or directory
func (a *App) RenameFile(oldPath, newPath string) error {
	err := a.fm.RenameFile(oldPath, newPath)
	if err != nil {
		return err
	}

	// Update path in database index
	if a.dbm.IsInitialized() {
		repo := a.dbm.Repository()
		_ = repo.RenameFile(oldPath, newPath)
	}

	return nil
}

// GetBasePath returns the current base path
func (a *App) GetBasePath() string {
	return a.fm.GetBasePath()
}

// ============ DATABASE API METHODS ============

// IndexFile indexes a file in the database
func (a *App) IndexFile(path string) error {
	return a.indexFile(path)
}

// indexFile is the internal implementation (can be called as goroutine)
func (a *App) indexFile(path string) error {
	if !a.dbm.IsInitialized() {
		return fmt.Errorf("database not initialized")
	}

	// Read file content
	content, err := a.fm.ReadFile(path)
	if err != nil {
		return err
	}

	return a.indexFileContent(path, content.Content)
}

// indexFileContent indexes a file with given content (avoids re-reading file)
func (a *App) indexFileContent(path, content string) error {
	if !a.dbm.IsInitialized() {
		return fmt.Errorf("database not initialized")
	}

	// Get file stats
	fullPath := filepath.Join(a.fm.GetBasePath(), path)
	info, err := os.Stat(fullPath)
	if err != nil {
		runtime.LogErrorf(a.ctx, "Failed to stat file %s: %v", path, err)
		return err
	}

	// Index in database
	repo := a.dbm.Repository()
	if err := repo.IndexFile(path, content, info.ModTime().Unix(), info.Size()); err != nil {
		runtime.LogErrorf(a.ctx, "Failed to index file %s: %v", path, err)
		return err
	}

	return nil
}

// GetIndexedFile retrieves metadata from database
func (a *App) GetIndexedFile(path string) (*database.File, error) {
	if !a.dbm.IsInitialized() {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.dbm.Repository().GetFileByPath(path)
}

// ListIndexedFiles returns all indexed files
func (a *App) ListIndexedFiles() ([]database.File, error) {
	if !a.dbm.IsInitialized() {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.dbm.Repository().ListFiles()
}

// RemoveFromIndex removes file from database index
func (a *App) RemoveFromIndex(path string) error {
	if !a.dbm.IsInitialized() {
		return fmt.Errorf("database not initialized")
	}
	return a.dbm.Repository().DeleteFile(path)
}

// UpdateFilePathInIndex updates file path after rename
func (a *App) UpdateFilePathInIndex(oldPath, newPath string) error {
	if !a.dbm.IsInitialized() {
		return fmt.Errorf("database not initialized")
	}
	return a.dbm.Repository().RenameFile(oldPath, newPath)
}

// GetDatabaseStats returns database statistics
func (a *App) GetDatabaseStats() (map[string]interface{}, error) {
	if !a.dbm.IsInitialized() {
		return nil, fmt.Errorf("database not initialized")
	}

	stats, err := a.dbm.Repository().GetStats()
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"files":  stats["files"],
		"chunks": stats["chunks"],
		"tags":   stats["tags"],
		"path":   a.dbm.GetDBPath(),
	}

	return result, nil
}

// IsDatabaseInitialized returns true if database is initialized
func (a *App) IsDatabaseInitialized() bool {
	return a.dbm.IsInitialized()
}
