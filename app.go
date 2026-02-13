package main

import (
	"context"
	"fmt"
	"notebit/pkg/ai"
	"notebit/pkg/config"
	"notebit/pkg/database"
	"notebit/pkg/files"
	"notebit/pkg/graph"
	"notebit/pkg/knowledge"
	"notebit/pkg/rag"
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
	ks      *knowledge.Service
	cfg     *config.Config
	watcher  *watcher.Service
	rag      *rag.Service
	graph    *graph.Service
	llm      ai.LLMProvider
	indexQueue chan indexJob
}

type indexJob struct {
	path    string
	content string
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
	fm := files.NewManager()
	dbm := database.GetInstance()
	aiService := ai.NewService(cfg)

	app := &App{
		fm:         fm,
		dbm:        dbm,
		cfg:        cfg,
		ai:         aiService,
		ks:         knowledge.NewService(fm, dbm, aiService),
		indexQueue: make(chan indexJob, 128),
	}
	app.startIndexWorkers(4)
	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.loadConfig(); err != nil {
		runtime.LogErrorf(a.ctx, "Failed to load config: %v", err)
	}
	a.initializeAI()
	a.initializeLLM()

	// Start file watcher if database is initialized and base path is set
	if a.dbm.IsInitialized() && a.fm.GetBasePath() != "" {
		if err := a.startWatcher(); err != nil {
			runtime.LogErrorf(a.ctx, "Failed to start watcher: %v", err)
		}
	}

	// Initialize RAG and Graph services after database is ready
	a.initializeRAG()
	a.initializeGraph()
}

func (a *App) loadConfig() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "notebit", "config.json")
	return a.cfg.LoadFromFile(configPath)
}

// initializeAI initializes the AI service
func (a *App) initializeAI() {
	if err := a.ai.Initialize(); err != nil {
		runtime.LogWarningf(a.ctx, "AI service initialization failed: %v", err)
	}
}

// initializeLLM initializes the LLM provider for chat completion
func (a *App) initializeLLM() {
	llmConfig := a.cfg.GetLLMConfig()
	if llmConfig.Provider == "" {
		return // No LLM provider configured
	}

	if llmConfig.Provider == "openai" {
		openAIConfig := a.cfg.GetOpenAIConfig()
		llm, err := ai.NewOpenAILLMProvider(openAIConfig)
		if err == nil {
			a.llm = llm
		} else {
			runtime.LogWarningf(a.ctx, "Failed to initialize OpenAI LLM: %v", err)
		}
	}
}

// initializeRAG initializes the RAG service
func (a *App) initializeRAG() {
	if a.llm != nil && a.dbm.IsInitialized() {
		a.rag = rag.NewService(a.dbm, a.ai, a.llm, a.cfg)
	}
}

// initializeGraph initializes the Graph service
func (a *App) initializeGraph() {
	if a.dbm.IsInitialized() {
		a.graph = graph.NewService(a.dbm, a.cfg)
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
	if a.indexQueue != nil {
		close(a.indexQueue)
		a.indexQueue = nil
	}
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
		a.enqueueIndexFileContent(path, content)
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
		a.enqueueIndexFileContent(path, content)
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

	if err := a.ai.Reconfigure(); err != nil {
		return err
	}
	return a.cfg.Save()
}

// GetRAGConfig returns the RAG configuration
func (a *App) GetRAGConfig() (config.RAGConfig, error) {
	return a.cfg.GetRAGConfig(), nil
}

// SetRAGConfig sets the RAG configuration
func (a *App) SetRAGConfig(maxContextChunks int, temperature float32, systemPrompt string) error {
	cfg := config.RAGConfig{
		MaxContextChunks: maxContextChunks,
		Temperature:      temperature,
		SystemPrompt:     systemPrompt,
	}
	a.cfg.SetRAGConfig(cfg)
	return a.cfg.Save()
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
	if err := a.ai.SetProvider(provider); err != nil {
		return err
	}
	return a.cfg.Save()
}

// SetAIModel sets the default embedding model
func (a *App) SetAIModel(model string) error {
	a.cfg.SetEmbeddingModel(model)
	if err := a.ai.Reconfigure(); err != nil {
		return err
	}
	return a.cfg.Save()
}

// SetOpenAIConfig sets the OpenAI configuration
func (a *App) SetOpenAIConfig(apiKey, baseURL, organization string) error {
	if err := a.ai.SetOpenAIConfig(apiKey, baseURL, organization); err != nil {
		return err
	}
	return a.cfg.Save()
}

// SetOllamaConfig sets the Ollama configuration
func (a *App) SetOllamaConfig(baseURL, model string, timeout int) error {
	if err := a.ai.SetOllamaConfig(baseURL, model, timeout); err != nil {
		return err
	}
	return a.cfg.Save()
}

func (a *App) TestOpenAIConnection(apiKey, baseURL, organization, model string) (map[string]interface{}, error) {
	provider, err := ai.NewOpenAIProvider(ai.OpenAIConfig{
		APIKey:       apiKey,
		BaseURL:      baseURL,
		Organization: organization,
		Timeout:      15 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	if model == "" {
		model = provider.GetDefaultModel()
	}

	resp, err := provider.GenerateEmbedding(&ai.EmbeddingRequest{
		Text:  "ping",
		Model: model,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"model":     resp.Model,
		"dimension": len(resp.Embedding),
	}, nil
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
	return a.ks.IndexFileWithEmbedding(path)
}

// ReindexAllWithEmbeddings reindexes all files with embeddings
func (a *App) ReindexAllWithEmbeddings() (map[string]interface{}, error) {
	return a.ks.ReindexAllWithEmbeddings()
}

// ============ SEMANTIC SEARCH API METHODS ============

// SimilarNote represents a note with similarity score for semantic search results
type SimilarNote struct {
	Path       string  `json:"path"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Heading    string  `json:"heading"`
	Similarity float32 `json:"similarity"`
	ChunkID    uint    `json:"chunk_id"`
}

// FindSimilar finds semantically similar notes based on content
func (a *App) FindSimilar(content string, limit int) ([]SimilarNote, error) {
	results, err := a.ks.FindSimilar(content, limit)
	if err != nil {
		return nil, err
	}

	// Convert to local struct to maintain API compatibility
	notes := make([]SimilarNote, len(results))
	for i, r := range results {
		notes[i] = SimilarNote{
			Path:       r.Path,
			Title:      r.Title,
			Content:    r.Content,
			Heading:    r.Heading,
			Similarity: r.Similarity,
			ChunkID:    r.ChunkID,
		}
	}
	return notes, nil
}

// GetSimilarityStatus returns the availability status of semantic search
func (a *App) GetSimilarityStatus() (map[string]interface{}, error) {
	return a.ks.GetSimilarityStatus()
}

// ============ RAG CHAT API METHODS ============

// RAGQuery performs a RAG query
func (a *App) RAGQuery(query string) (map[string]interface{}, error) {
	if a.rag == nil {
		return nil, fmt.Errorf("RAG service not initialized")
	}

	response, err := a.rag.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message_id": response.MessageID,
		"content":    response.Content,
		"sources":    response.Sources,
		"tokens_used": response.TokensUsed,
	}, nil
}

// GetRAGStatus returns the status of the RAG service
func (a *App) GetRAGStatus() (map[string]interface{}, error) {
	if a.rag == nil {
		return map[string]interface{}{
			"available":      false,
			"llm_provider":  "",
			"llm_model":     "",
			"database_ready": a.dbm.IsInitialized(),
		}, nil
	}

	status := a.rag.GetStatus()

	return map[string]interface{}{
		"available":      status.Available,
		"llm_provider":   status.LLMProvider,
		"llm_model":     status.LLMModel,
		"database_ready": status.DatabaseReady,
	}, nil
}

// ============ GRAPH API METHODS ============

// GetGraphData returns the knowledge graph data
func (a *App) GetGraphData() (*graph.GraphData, error) {
	if a.graph == nil {
		return &graph.GraphData{Nodes: []graph.Node{}, Links: []graph.Link{}}, nil
	}

	return a.graph.BuildGraph()
}

// GetGraphConfig returns the graph configuration
func (a *App) GetGraphConfig() (config.GraphConfig, error) {
	return a.cfg.GetGraphConfig(), nil
}

// SetGraphConfig sets the graph configuration
func (a *App) SetGraphConfig(minSimilarityThreshold float32, maxNodes int, showImplicitLinks bool) error {
	cfg := config.GraphConfig{
		MinSimilarityThreshold: minSimilarityThreshold,
		MaxNodes:                 maxNodes,
		ShowImplicitLinks:       showImplicitLinks,
	}
	a.cfg.SetGraphConfig(cfg)
	return a.cfg.Save()
}

// ============ LLM CONFIG API METHODS ============

// GetLLMConfig returns the LLM configuration
func (a *App) GetLLMConfig() (config.LLMConfig, error) {
	return a.cfg.GetLLMConfig(), nil
}

// SetLLMConfig sets the LLM configuration
func (a *App) SetLLMConfig(provider string, model string, temperature float32, maxTokens int) error {
	llmConfig := config.LLMConfig{
		Provider:    provider,
		Model:       model,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	a.cfg.SetLLMConfig(llmConfig)

	// Reinitialize LLM provider
	a.initializeLLM()

	// Reinitialize RAG service if needed
	if a.rag != nil {
		a.initializeRAG()
	}

	return a.cfg.Save()
}
