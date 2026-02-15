package main

import (
	"context"
	"fmt"
	"notebit/pkg/ai"
	"notebit/pkg/chat"
	"notebit/pkg/config"
	"notebit/pkg/database"
	"notebit/pkg/files"
	"notebit/pkg/graph"
	"notebit/pkg/indexing"
	"notebit/pkg/knowledge"
	"notebit/pkg/logger"
	"notebit/pkg/rag"
	"notebit/pkg/watcher"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	fm       *files.Manager
	dbm      *database.Manager
	ai       *ai.Service
	ks       *knowledge.Service
	cfg      *config.Config
	watcher  *watcher.Service
	rag      *rag.Service
	graph    *graph.Service
	llm      ai.LLMProvider
	pipeline *indexing.IndexingPipeline
	chatSvc  *chat.Service
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
		fm:  fm,
		dbm: dbm,
		cfg: cfg,
		ai:  aiService,
	}
	return app
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	timer := logger.StartTimer()
	a.ctx = ctx
	logger.Info("App startup initiated")

	if err := a.loadConfig(); err != nil {
		logger.ErrorWithFields(ctx, map[string]interface{}{"error": err.Error()}, "Failed to load config")
		runtime.LogErrorf(a.ctx, "Failed to load config: %v", err)
	}

	a.initializeAI()
	a.initializeLLM()

	// Initialize indexing pipeline after database is ready
	if a.dbm.IsInitialized() {
		a.pipeline = indexing.NewPipeline(a.ai, a.dbm.Repository(), a.fm)
		a.pipeline.Start()
		a.ks = knowledge.NewService(a.fm, a.dbm, a.ai, a.pipeline)
		a.initializeChat()
	}

	// Start file watcher if database is initialized and base path is set
	if a.dbm.IsInitialized() && a.fm.GetBasePath() != "" {
		if err := a.startWatcher(); err != nil {
			logger.ErrorWithFields(ctx, map[string]interface{}{"base_path": a.fm.GetBasePath()}, "Failed to start watcher: %v", err)
			runtime.LogErrorf(a.ctx, "Failed to start watcher: %v", err)
		}
	}

	// Initialize RAG and Graph services after database is ready
	a.initializeRAG()
	a.initializeGraph()
	a.applyVectorEngineConfig()

	logger.InfoWithDuration(ctx, timer(), "App startup completed")
}

func (a *App) applyVectorEngineConfig() {
	if !a.dbm.IsInitialized() {
		return
	}

	repo := a.dbm.Repository()
	configured := a.cfg.GetVectorSearchEngine()
	if configured == "" {
		configured = "brute-force"
	}
	effective := repo.SetVectorEngine(configured)
	if effective != configured {
		logger.WarnWithFields(a.ctx, map[string]interface{}{
			"requested": configured,
			"effective": effective,
		}, "Vector engine fallback applied")
	}
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
	timer := logger.StartTimer()
	logger.Debug("Initializing AI service")

	if err := a.ai.Initialize(); err != nil {
		logger.WarnWithFields(a.ctx, map[string]interface{}{"error": err.Error()}, "AI service initialization failed")
		runtime.LogWarningf(a.ctx, "AI service initialization failed: %v", err)
	} else {
		logger.InfoWithDuration(a.ctx, timer(), "AI service initialized successfully")
	}
}

// initializeLLM initializes the LLM provider for chat completion
func (a *App) initializeLLM() {
	llmConfig := a.cfg.GetLLMConfig()
	if llmConfig.Provider == "" {
		return // No LLM provider configured
	}

	if llmConfig.Provider == "openai" {
		// Start with dedicated LLM OpenAI config
		openAIConfig := llmConfig.OpenAI

		// Fallback to global AI config if API Key is missing
		// This maintains backward compatibility and ease of use
		globalOpenAI := a.cfg.GetOpenAIConfig()

		if openAIConfig.APIKey == "" {
			openAIConfig.APIKey = globalOpenAI.APIKey
		}

		// Use global BaseURL if local is empty, or default
		if openAIConfig.BaseURL == "" {
			if globalOpenAI.BaseURL != "" {
				openAIConfig.BaseURL = globalOpenAI.BaseURL
			} else {
				openAIConfig.BaseURL = "https://api.openai.com/v1"
			}
		}

		if openAIConfig.Organization == "" {
			openAIConfig.Organization = globalOpenAI.Organization
		}

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

func (a *App) initializeChat() {
	if !a.dbm.IsInitialized() {
		return
	}
	if a.chatSvc != nil {
		a.chatSvc.Close()
		a.chatSvc = nil
	}
	svc, err := chat.NewService(a.dbm.GetDB(), a.dbm.GetBasePath())
	if err != nil {
		runtime.LogWarningf(a.ctx, "Failed to initialize chat service: %v", err)
		return
	}
	a.chatSvc = svc
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

	if a.pipeline == nil {
		return fmt.Errorf("indexing pipeline not initialized")
	}

	var err error
	a.watcher, err = watcher.NewService(baseDir, a.pipeline)
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

	if a.ks == nil {
		a.ks = knowledge.NewService(a.fm, a.dbm, a.ai, a.pipeline)
	}

	results, err := a.ks.ReindexAllWithEmbeddings()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Full index failed: %v", err)
		return
	}

	runtime.LogInfof(a.ctx, "Full index complete: %v", results)
}

// shutdown is called when the app is shutting down
func (a *App) shutdown(context.Context) {
	a.stopWatcher()
	if a.pipeline != nil {
		a.pipeline.Stop()
	}
	if a.chatSvc != nil {
		a.chatSvc.Close()
	}
}
