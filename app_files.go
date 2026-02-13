package main

import (
	"fmt"
	"notebit/pkg/database"
	"notebit/pkg/files"
	"notebit/pkg/logger"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ============ FILE OPERATION API METHODS ============

// OpenFolder opens a directory dialog and sets the base path
func (a *App) OpenFolder() (string, error) {
	timer := logger.StartTimer()
	logger.Info("Opening folder dialog")

	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Notes Folder",
	})
	if err != nil {
		logger.ErrorWithFields(a.ctx, map[string]interface{}{"error": err.Error()}, "Failed to open directory dialog")
		return "", err
	}

	if dir == "" {
		logger.Debug("User cancelled folder selection")
		return "", nil
	}

	// Stop existing watcher before changing folder
	a.stopWatcher()

	if err := a.fm.SetBasePath(dir); err != nil {
		logger.ErrorWithFields(a.ctx, map[string]interface{}{"path": dir, "error": err.Error()}, "Failed to set base path")
		return "", err
	}

	// Initialize database
	if err := a.dbm.Init(dir); err != nil {
		// Don't fail if database initialization fails - log but continue
		logger.WarnWithFields(a.ctx, map[string]interface{}{"path": dir, "error": err.Error()}, "Database initialization failed")
	}
	a.applyVectorEngineConfig()
	a.initializeRAG()
	a.initializeGraph()

	// Start watcher for new folder
	if err := a.startWatcher(); err != nil {
		logger.ErrorWithFields(a.ctx, map[string]interface{}{"path": dir}, "Failed to start watcher: %v", err)
		runtime.LogErrorf(a.ctx, "Failed to start watcher: %v", err)
	}

	logger.InfoWithDuration(a.ctx, timer(), "Folder opened successfully: %s", dir)
	return dir, nil
}

// SetFolder sets the base path without opening a dialog
func (a *App) SetFolder(path string) error {
	// Stop existing watcher before changing folder
	a.stopWatcher()

	if err := a.fm.SetBasePath(path); err != nil {
		return err
	}

	// Initialize database
	if err := a.dbm.Init(path); err != nil {
		// Don't fail if database initialization fails
		logger.Warn("Warning: database initialization failed: %v", err)
	}
	a.applyVectorEngineConfig()
	a.initializeRAG()
	a.initializeGraph()

	// Start watcher for new folder
	if err := a.startWatcher(); err != nil {
		runtime.LogErrorf(a.ctx, "Failed to start watcher: %v", err)
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
	timer := logger.StartTimer()
	logger.DebugWithFields(a.ctx, map[string]interface{}{
		"path":         path,
		"content_size": len(content),
	}, "Saving file")

	err := a.fm.SaveFile(path, content)
	if err != nil {
		logger.ErrorWithFields(a.ctx, map[string]interface{}{
			"path":  path,
			"error": err.Error(),
		}, "Failed to save file")
		return err
	}

	// Index the file in database after saving (pass content to avoid re-reading)
	if a.dbm.IsInitialized() {
		a.enqueueIndexFileContent(path, content)
	}

	logger.InfoWithDuration(a.ctx, timer(), "File saved: %s", path)
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
		a.enqueueIndexFileContent(path, content)
	}

	return nil
}

// DeleteFile deletes a markdown file or directory
func (a *App) DeleteFile(path string) error {
	timer := logger.StartTimer()
	logger.InfoWithFields(a.ctx, map[string]interface{}{"path": path}, "Deleting file")

	err := a.fm.DeleteFile(path)
	if err != nil {
		logger.ErrorWithFields(a.ctx, map[string]interface{}{
			"path":  path,
			"error": err.Error(),
		}, "Failed to delete file")
		return err
	}

	// Remove from database index
	if a.dbm.IsInitialized() {
		repo := a.dbm.Repository()
		if err := repo.DeleteFile(path); err != nil {
			logger.WarnWithFields(a.ctx, map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			}, "Failed to delete file from index")
		}
	}

	logger.InfoWithDuration(a.ctx, timer(), "File deleted: %s", path)
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

// ============ INDEX OPERATIONS ============

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
	timer := logger.StartTimer()

	if !a.dbm.IsInitialized() {
		return fmt.Errorf("database not initialized")
	}

	// Get file stats
	fullPath := filepath.Join(a.fm.GetBasePath(), path)
	info, err := os.Stat(fullPath)
	if err != nil {
		logger.ErrorWithFields(a.ctx, map[string]interface{}{
			"path":  path,
			"error": err.Error(),
		}, "Failed to stat file")
		runtime.LogErrorf(a.ctx, "Failed to stat file %s: %v", path, err)
		return err
	}

	if err := a.indexFileWithEmbeddings(path, content, info.ModTime().Unix(), info.Size()); err != nil {
		logger.ErrorWithFields(a.ctx, map[string]interface{}{
			"path":  path,
			"error": err.Error(),
		}, "Failed to index file")
		runtime.LogErrorf(a.ctx, "Failed to index file %s: %v", path, err)
		return err
	}

	logger.DebugWithDuration(a.ctx, timer(), "File indexed: %s", path)
	return nil
}

// indexFileWithEmbeddings indexes file metadata + chunks/embeddings in one path.
// Falls back to metadata-only indexing when AI processing is unavailable.
func (a *App) indexFileWithEmbeddings(path, content string, modTime, size int64) error {
	repo := a.dbm.Repository()

	chunks, err := a.ai.ProcessDocument(content)
	if err != nil {
		logger.WarnWithFields(a.ctx, map[string]interface{}{
			"path":  path,
			"error": err.Error(),
		}, "Embedding processing failed, fallback to metadata-only index")
		bareChunks, chunkErr := a.ai.ChunkText(content)
		if chunkErr != nil {
			return repo.IndexFile(path, content, modTime, size)
		}
		dbChunks := make([]database.ChunkInput, len(bareChunks))
		for i, chunk := range bareChunks {
			dbChunks[i] = database.ChunkInput{
				Content: chunk.Content,
				Heading: chunk.Heading,
			}
		}
		return repo.IndexFileWithChunks(path, content, modTime, size, dbChunks)
	}

	dbChunks := make([]database.ChunkInput, len(chunks))
	for i, chunk := range chunks {
		dbChunks[i] = database.ChunkInput{
			Content: chunk.Content,
			Heading: chunk.Heading,
		}
		if embedding, ok := chunk.Metadata["embedding"].([]float32); ok {
			dbChunks[i].Embedding = embedding
		}
		if model, ok := chunk.Metadata["embedding_model"].(string); ok {
			dbChunks[i].EmbeddingModel = model
		}
	}

	return repo.IndexFileWithChunks(path, content, modTime, size, dbChunks)
}

func (a *App) startIndexWorkers(count int) {
	if count <= 0 {
		count = 1
	}
	for i := 0; i < count; i++ {
		go a.indexWorker()
	}
}

func (a *App) indexWorker() {
	for job := range a.indexQueue {
		_ = a.indexFileContent(job.path, job.content)
	}
}

func (a *App) enqueueIndexFileContent(path, content string) {
	if a.indexQueue == nil {
		go a.indexFileContent(path, content)
		return
	}
	a.indexQueue <- indexJob{path: path, content: content}
}

// ============ DATABASE API METHODS ============

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
