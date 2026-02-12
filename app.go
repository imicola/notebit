package main

import (
	"context"
	"fmt"
	"notebit/pkg/ai"
	"notebit/pkg/config"
	"notebit/pkg/database"
	"notebit/pkg/files"
	"notebit/pkg/watcher"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	fm      *files.Manager
	dbm     *database.Manager
	ai      *ai.Service
	cfg     *config.Config
	watcher *watcher.Service
}

type watcherLogger struct {
	ctx context.Context
}

func (l watcherLogger) Errorf(format string, args ...interface{}) {
	runtime.LogErrorf(l.ctx, format, args...)
}

// NewApp creates a new App application struct
func NewApp() *App {
	return NewAppWithConfig(config.Get())
}

func NewAppWithConfig(cfg *config.Config) *App {
	if cfg == nil {
		cfg = config.New()
	}
	return &App{
		fm:  files.NewManager(),
		dbm: database.GetInstance(),
		cfg: cfg,
		ai:  ai.NewService(cfg),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.initializeAI()

	// Start file watcher if database is initialized and base path is set
	if a.dbm.IsInitialized() && a.fm.GetBasePath() != "" {
		if err := a.startWatcher(); err != nil {
			runtime.LogErrorf(a.ctx, "Failed to start watcher: %v", err)
		}
	}
}

// initializeAI initializes the AI service
func (a *App) initializeAI() {
	if err := a.ai.Initialize(); err != nil {
		runtime.LogWarningf(a.ctx, "AI service initialization failed: %v", err)
	}
}

// startWatcher starts the file watcher service
func (a *App) startWatcher() error {
	watcherCfg := a.cfg.GetWatcherConfig()
	if !watcherCfg.Enabled {
		return nil
	}

	baseDir := a.fm.GetBasePath()
	if baseDir == "" {
		return fmt.Errorf("no base path set")
	}

	var err error
	a.watcher, err = watcher.NewService(baseDir, a.fm, a.dbm, a.ai)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	// Configure watcher
	a.watcher.SetDebounceDelay(time.Duration(watcherCfg.DebounceMS) * time.Millisecond)
	a.watcher.SetWorkerCount(watcherCfg.Workers)
	a.watcher.SetLogger(watcherLogger{ctx: a.ctx})

	if err := a.watcher.Start(); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	// Run full index in background if enabled
	if watcherCfg.FullIndexOnStart {
		go a.runFullIndex()
	}

	return nil
}

// stopWatcher stops the file watcher service
func (a *App) stopWatcher() {
	if a.watcher != nil {
		if err := a.watcher.Stop(); err != nil {
			runtime.LogErrorf(a.ctx, "Failed to stop watcher: %v", err)
		}
		a.watcher = nil
	}
}

// runFullIndex runs a full background index of all markdown files
func (a *App) runFullIndex() {
	if a.watcher == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	progress, err := a.watcher.IndexAll(ctx)
	if err != nil {
		runtime.LogErrorf(a.ctx, "Full index failed: %v", err)
		return
	}

	// Wait for completion
	<-progress.Done

	total, processed, failed, _ := progress.GetProgress()
	runtime.LogInfof(a.ctx, "Full index complete: %d total, %d processed, %d failed", total, processed, failed)
}

// shutdown is called when the app is shutting down
func (a *App) shutdown(context.Context) {
	a.stopWatcher()
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

	// Stop existing watcher before changing folder
	a.stopWatcher()

	if err := a.fm.SetBasePath(dir); err != nil {
		return "", err
	}

	// Initialize database
	if err := a.dbm.Init(dir); err != nil {
		// Don't fail if database initialization fails - log but continue
		fmt.Printf("Warning: database initialization failed: %v\n", err)
	}

	// Start watcher for new folder
	if err := a.startWatcher(); err != nil {
		runtime.LogErrorf(a.ctx, "Failed to start watcher: %v", err)
	}

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
		fmt.Printf("Warning: database initialization failed: %v\n", err)
	}

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

// ============ AI SERVICE API METHODS ============

// GetOpenAIConfig returns the OpenAI configuration
func (a *App) GetOpenAIConfig() (config.OpenAIConfig, error) {
	return a.cfg.GetOpenAIConfig(), nil
}

// GetOllamaConfig returns the Ollama configuration
func (a *App) GetOllamaConfig() (config.OllamaConfig, error) {
	return a.cfg.GetOllamaConfig(), nil
}

// GetChunkingConfig returns the chunking configuration
func (a *App) GetChunkingConfig() (config.ChunkingConfig, error) {
	return a.cfg.GetChunkingConfig(), nil
}

// SetChunkingConfig sets the chunking configuration
func (a *App) SetChunkingConfig(strategy string, chunkSize, chunkOverlap, minChunkSize, maxChunkSize int, preserveHeading bool, headingSeparator string) error {
	cfg := config.ChunkingConfig{
		Strategy:         strategy,
		ChunkSize:        chunkSize,
		ChunkOverlap:     chunkOverlap,
		MinChunkSize:     minChunkSize,
		MaxChunkSize:     maxChunkSize,
		PreserveHeading:  preserveHeading,
		HeadingSeparator: headingSeparator,
	}
	a.cfg.SetChunkingConfig(cfg)

	// Reconfigure AI service if needed (though chunking is usually stateless or used per request)
	// But if the AI service caches chunking strategy, we might need to update it.
	// Looking at service.go (not read yet, but assumed), it likely uses config.Get() or passed config.
	// The app.ai service was initialized with cfg.

	return a.ai.Reconfigure()
}

// GetAIStatus returns the current status of the AI service
func (a *App) GetAIStatus() (map[string]interface{}, error) {
	status, err := a.ai.GetStatus()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"current_provider":     status.CurrentProvider,
		"available_providers":  status.AvailableProviders,
		"current_model":        status.CurrentModel,
		"chunking_strategy":    status.ChunkingStrategy,
		"available_strategies": status.AvailableStrategies,
		"model_dimension":      status.ModelDimension,
		"provider_healthy":     status.ProviderHealthy,
	}, nil
}

// SetAIProvider sets the current AI provider
func (a *App) SetAIProvider(provider string) error {
	return a.ai.SetProvider(provider)
}

// SetAIModel sets the default embedding model
func (a *App) SetAIModel(model string) error {
	a.cfg.SetEmbeddingModel(model)
	return a.ai.Reconfigure()
}

// SetOpenAIConfig sets the OpenAI configuration
func (a *App) SetOpenAIConfig(apiKey, baseURL, organization string) error {
	return a.ai.SetOpenAIConfig(apiKey, baseURL, organization)
}

// SetOllamaConfig sets the Ollama configuration
func (a *App) SetOllamaConfig(baseURL, model string, timeout int) error {
	return a.ai.SetOllamaConfig(baseURL, model, timeout)
}

// GenerateEmbedding generates an embedding for a single text
func (a *App) GenerateEmbedding(text string) ([]float32, error) {
	resp, err := a.ai.GenerateEmbedding(text)
	if err != nil {
		return nil, err
	}
	return resp.Embedding, nil
}

// GenerateEmbeddingsBatch generates embeddings for multiple texts
func (a *App) GenerateEmbeddingsBatch(texts []string) ([][]float32, error) {
	responses, err := a.ai.GenerateEmbeddingsBatch(texts)
	if err != nil {
		return nil, err
	}

	result := make([][]float32, len(responses))
	for i, resp := range responses {
		if resp != nil {
			result[i] = resp.Embedding
		}
	}
	return result, nil
}

// ChunkText splits text using the configured chunking strategy
func (a *App) ChunkText(text string) ([]map[string]interface{}, error) {
	chunks, err := a.ai.ChunkText(text)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(chunks))
	for i, chunk := range chunks {
		result[i] = map[string]interface{}{
			"content":  chunk.Content,
			"heading":  chunk.Heading,
			"index":    chunk.Index,
			"metadata": chunk.Metadata,
		}
	}
	return result, nil
}

// ProcessDocument chunks text and generates embeddings for all chunks
func (a *App) ProcessDocument(text string) ([]map[string]interface{}, error) {
	chunks, err := a.ai.ProcessDocument(text)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(chunks))
	for i, chunk := range chunks {
		result[i] = map[string]interface{}{
			"content":  chunk.Content,
			"heading":  chunk.Heading,
			"index":    chunk.Index,
			"metadata": chunk.Metadata,
		}
	}
	return result, nil
}

// IndexFileWithEmbedding indexes a file and generates embeddings for its chunks
func (a *App) IndexFileWithEmbedding(path string) error {
	if !a.dbm.IsInitialized() {
		return fmt.Errorf("database not initialized")
	}

	// Read file content
	content, err := a.fm.ReadFile(path)
	if err != nil {
		return err
	}

	// Get file stats
	fullPath := filepath.Join(a.fm.GetBasePath(), path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}

	// Process document to get chunks with embeddings
	chunks, err := a.ai.ProcessDocument(content.Content)
	if err != nil {
		return fmt.Errorf("failed to process document: %w", err)
	}

	// Convert chunks to database format
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

	// Index with embeddings
	repo := a.dbm.Repository()
	return repo.IndexFileWithChunks(path, content.Content, info.ModTime().Unix(), info.Size(), dbChunks)
}

// ReindexAllWithEmbeddings reindexes all files with embeddings
func (a *App) ReindexAllWithEmbeddings() (map[string]interface{}, error) {
	if !a.dbm.IsInitialized() {
		return nil, fmt.Errorf("database not initialized")
	}

	files, err := a.fm.ListFiles()
	if err != nil {
		return nil, err
	}

	// Collect all markdown files
	var mdFiles []string
	collectFiles(files, &mdFiles)

	results := map[string]interface{}{
		"total":     len(mdFiles),
		"processed": 0,
		"failed":    0,
		"errors":    []string{},
	}

	for _, path := range mdFiles {
		if err := a.IndexFileWithEmbedding(path); err != nil {
			results["failed"] = results["failed"].(int) + 1
			errs := results["errors"].([]string)
			results["errors"] = append(errs, fmt.Sprintf("%s: %v", path, err))
		} else {
			results["processed"] = results["processed"].(int) + 1
		}
	}

	return results, nil
}

// collectFiles recursively collects all markdown file paths
func collectFiles(node *files.FileNode, paths *[]string) {
	if !node.IsDir {
		*paths = append(*paths, node.Path)
	} else {
		for _, child := range node.Children {
			collectFiles(child, paths)
		}
	}
}
