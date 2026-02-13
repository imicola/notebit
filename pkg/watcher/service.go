package watcher

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"notebit/pkg/ai"
	"notebit/pkg/database"
	"notebit/pkg/files"
	"notebit/pkg/logger"

	"github.com/fsnotify/fsnotify"
)

// Service handles file system watching and automatic indexing
type Service struct {
	baseDir    string
	fm         *files.Manager
	dbm        *database.Manager
	ai         *ai.Service
	logger     Logger
	watcher    *fsnotify.Watcher
	eventQueue chan FileEvent
	done       chan struct{}
	mu         sync.RWMutex

	// Debouncing
	pendingEvents map[string]*time.Timer
	pendingMu     sync.Mutex
	debounceDelay time.Duration

	// Worker pool
	workerSem chan struct{}
}

// FileEvent represents a file system event
type FileEvent struct {
	Path      string
	Op        fsnotify.Op
	Timestamp time.Time
}

// IndexProgress tracks the progress of a full index operation
type IndexProgress struct {
	Total     int
	Processed int
	Failed    int
	Current   string
	mu        sync.Mutex
	Done      chan struct{}
}

type Logger interface {
	Errorf(format string, args ...interface{})
}

// NewService creates a new watcher service
func NewService(baseDir string, fm *files.Manager, dbm *database.Manager, aiService *ai.Service) (*Service, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	// Default debounce delay: 500ms
	debounceDelay := 500 * time.Millisecond

	return &Service{
		baseDir:       baseDir,
		fm:            fm,
		dbm:           dbm,
		ai:            aiService,
		watcher:       watcher,
		eventQueue:    make(chan FileEvent, 100),
		done:          make(chan struct{}),
		pendingEvents: make(map[string]*time.Timer),
		debounceDelay: debounceDelay,
		workerSem:     make(chan struct{}, 3), // Default 3 workers
	}, nil
}

// SetDebounceDelay sets the debounce delay for file events
func (s *Service) SetDebounceDelay(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.debounceDelay = d
}

// SetWorkerCount sets the number of concurrent indexing workers
func (s *Service) SetWorkerCount(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workerSem = make(chan struct{}, n)
}

func (s *Service) SetLogger(logger Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = logger
}

// Start begins watching the base directory
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add the base directory to the watcher
	if err := s.watcher.Add(s.baseDir); err != nil {
		return fmt.Errorf("failed to watch directory %s: %w", s.baseDir, err)
	}

	// Start event processing goroutines
	go s.eventLoop()
	go s.workerLoop()

	return nil
}

// Stop stops the watcher service gracefully
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Signal all goroutines to stop
	close(s.done)

	// Stop the watcher
	if s.watcher != nil {
		if err := s.watcher.Close(); err != nil {
			return fmt.Errorf("failed to close watcher: %w", err)
		}
	}

	return nil
}

// eventLoop processes fsnotify events
func (s *Service) eventLoop() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			s.handleEvent(event)

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			if s.logger != nil {
				s.logger.Errorf("Watcher error: %v", err)
			} else {
				logger.Error("Watcher error: %v", err)
			}

		case <-s.done:
			return
		}
	}
}

// workerLoop processes file events from the queue
func (s *Service) workerLoop() {
	for {
		select {
		case event := <-s.eventQueue:
			s.processFile(event.Path, event.Op)

		case <-s.done:
			// Drain remaining events
			for len(s.eventQueue) > 0 {
				event := <-s.eventQueue
				s.processFile(event.Path, event.Op)
			}
			return
		}
	}
}

// handleEvent handles a single fsnotify event
func (s *Service) handleEvent(event fsnotify.Event) {
	// Skip if not a markdown file
	if !isMarkdownFile(event.Name) {
		return
	}

	// Convert to relative path
	relPath, err := filepath.Rel(s.baseDir, event.Name)
	if err != nil {
		return
	}

	// Skip temporary/editor files
	if isTemporaryFile(relPath) {
		return
	}

	// Skip ignored directories
	if isInIgnoredDir(relPath) {
		return
	}

	// Handle Create event on directories - add to watcher
	if event.Op&fsnotify.Create == fsnotify.Create && isDir(event.Name) {
		s.mu.RLock()
		if s.watcher != nil {
			_ = s.watcher.Add(event.Name)
		}
		s.mu.RUnlock()
		return
	}

	// Debounce file events
	s.pendingMu.Lock()
	if timer, exists := s.pendingEvents[relPath]; exists {
		timer.Stop()
	}

	// Create new timer for debouncing
	s.pendingEvents[relPath] = time.AfterFunc(s.debounceDelay, func() {
		s.eventQueue <- FileEvent{
			Path:      relPath,
			Op:        event.Op,
			Timestamp: time.Now(),
		}

		s.pendingMu.Lock()
		delete(s.pendingEvents, relPath)
		s.pendingMu.Unlock()
	})
	s.pendingMu.Unlock()
}

// processFile processes a file event (Create, Write, Remove, Rename)
func (s *Service) processFile(path string, op fsnotify.Op) {
	// Acquire worker semaphore
	select {
	case s.workerSem <- struct{}{}:
		defer func() { <-s.workerSem }()
	case <-s.done:
		return
	}

	// Handle different operation types
	switch {
	case op&fsnotify.Remove == fsnotify.Remove:
		s.handleRemove(path)

	case op&fsnotify.Rename == fsnotify.Rename:
		s.handleRename(path)

	case op&fsnotify.Create == fsnotify.Create, op&fsnotify.Write == fsnotify.Write:
		s.handleWrite(path)
	}
}

// handleWrite handles file creation/modification
func (s *Service) handleWrite(path string) {
	// Check if database is initialized
	if !s.dbm.IsInitialized() {
		return
	}

	// Read file content
	content, err := s.fm.ReadFile(path)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorf("Failed to read file %s: %v", path, err)
		} else {
			logger.Error("Failed to read file %s: %v", path, err)
		}
		return
	}

	// Check if re-indexing is needed
	repo := s.dbm.Repository()
	needsIndexing, err := repo.FileNeedsIndexing(path, content.Content)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorf("Failed to check indexing status for %s: %v", path, err)
		} else {
			logger.Error("Failed to check indexing status for %s: %v", path, err)
		}
		return
	}

	if !needsIndexing {
		return
	}

	// Index with embeddings
	s.indexFileWithEmbeddings(path, content.Content)
}

// handleRemove handles file deletion
func (s *Service) handleRemove(path string) {
	if !s.dbm.IsInitialized() {
		return
	}

	repo := s.dbm.Repository()
	if err := repo.DeleteFile(path); err != nil {
		if s.logger != nil {
			s.logger.Errorf("Failed to delete file from index: %s: %v", path, err)
		} else {
			logger.Error("Failed to delete file from index: %s: %v", path, err)
		}
	}
}

// handleRename handles file rename
func (s *Service) handleRename(oldPath string) {
	// After rename, fsnotify sends a Create event for the new path
	// But we must remove the old path from the index to avoid ghost files
	s.handleRemove(oldPath)
}

// indexFileWithEmbeddings indexes a file with its embeddings
func (s *Service) indexFileWithEmbeddings(path, content string) {
	if !s.dbm.IsInitialized() {
		return
	}

	// Get file stats
	fullPath := filepath.Join(s.fm.GetBasePath(), path)
	info, err := getFileStat(fullPath)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorf("Failed to stat file %s: %v", path, err)
		} else {
			logger.Error("Failed to stat file %s: %v", path, err)
		}
		return
	}

	// Process document to get chunks with embeddings
	chunks, err := s.ai.ProcessDocument(content)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorf("Failed to process document %s: %v", path, err)
		} else {
			logger.Error("Failed to process document %s: %v", path, err)
		}
		// Still try to index without embeddings
		s.indexFileMetadataOnly(path, content, info.ModTime, info.Size)
		return
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
	repo := s.dbm.Repository()
	if err := repo.IndexFileWithChunks(path, content, info.ModTime, info.Size, dbChunks); err != nil {
		if s.logger != nil {
			s.logger.Errorf("Failed to index file %s: %v", path, err)
		} else {
			logger.Error("Failed to index file %s: %v", path, err)
		}
	}
}

// indexFileMetadataOnly indexes a file without embeddings (fallback)
func (s *Service) indexFileMetadataOnly(path, content string, modTime int64, size int64) {
	repo := s.dbm.Repository()
	if err := repo.IndexFile(path, content, modTime, size); err != nil {
		if s.logger != nil {
			s.logger.Errorf("Failed to index file metadata %s: %v", path, err)
		} else {
			logger.Error("Failed to index file metadata %s: %v", path, err)
		}
	}
}

// IndexAll performs a full index of all markdown files in the base directory
func (s *Service) IndexAll(ctx context.Context) (*IndexProgress, error) {
	// Get file tree
	files, err := s.fm.ListFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Collect all markdown files
	var mdFiles []string
	collectMarkdownFiles(files, &mdFiles)

	progress := &IndexProgress{
		Total: len(mdFiles),
		Done:  make(chan struct{}),
	}

	// Start background indexing
	go s.runFullIndex(ctx, mdFiles, progress)

	return progress, nil
}

// runFullIndex performs the actual full indexing
func (s *Service) runFullIndex(ctx context.Context, files []string, progress *IndexProgress) {
	var wg sync.WaitGroup

	for _, path := range files {
		// Check for cancellation or stop signal
		select {
		case <-ctx.Done():
			goto Done
		case <-s.done:
			goto Done
		default:
		}

		// Acquire worker semaphore to limit concurrency
		select {
		case s.workerSem <- struct{}{}:
			// Acquired
		case <-ctx.Done():
			goto Done
		case <-s.done:
			goto Done
		}

		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			defer func() { <-s.workerSem }()

			// Check context again inside goroutine
			select {
			case <-ctx.Done():
				return
			case <-s.done:
				return
			default:
			}

			progress.mu.Lock()
			progress.Current = filePath
			progress.mu.Unlock()

			// Read and check if needs indexing
			content, err := s.fm.ReadFile(filePath)
			if err != nil {
				progress.mu.Lock()
				progress.Failed++
				progress.mu.Unlock()
				return
			}

			repo := s.dbm.Repository()
			needsIndexing, _ := repo.FileNeedsIndexing(filePath, content.Content)
			if !needsIndexing {
				return
			}

			// Index the file
			s.indexFileWithEmbeddings(filePath, content.Content)

			progress.mu.Lock()
			progress.Processed++
			progress.mu.Unlock()
		}(path)
	}

Done:
	// Wait for all workers to finish
	wg.Wait()
	progress.Done <- struct{}{}
}

// GetProgress returns the current progress of an index operation
func (p *IndexProgress) GetProgress() (total, processed, failed int, current string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Total, p.Processed, p.Failed, p.Current
}

// Helper functions

func isMarkdownFile(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".md"
}

func isTemporaryFile(path string) bool {
	base := filepath.Base(path)
	// Skip vim swap files
	if strings.HasPrefix(base, ".") && strings.HasSuffix(base, ".swp") {
		return true
	}
	// Skip backup files
	if strings.HasSuffix(base, "~") || strings.HasSuffix(base, ".tmp") {
		return true
	}
	return false
}

func isInIgnoredDir(path string) bool {
	// Normalize path separators
	path = filepath.ToSlash(path)
	parts := strings.Split(path, "/")

	// Skip .git, node_modules, .idea, etc.
	ignored := []string{".git", "node_modules", ".idea", "target", "dist", "build"}

	for _, part := range parts {
		for _, ignore := range ignored {
			if part == ignore {
				return true
			}
		}
	}
	return false
}

func isDir(path string) bool {
	info, err := getFileStat(path)
	if err != nil {
		return false
	}
	return info.IsDir
}

func collectMarkdownFiles(node *files.FileNode, paths *[]string) {
	if !node.IsDir {
		*paths = append(*paths, node.Path)
	} else {
		for _, child := range node.Children {
			collectMarkdownFiles(child, paths)
		}
	}
}

// fileInfo wraps os.FileInfo for easier mocking/testing
type fileInfo struct {
	Name    string
	Size    int64
	ModTime int64
	IsDir   bool
}

func getFileStat(path string) (*fileInfo, error) {
	info, err := getFileStatRaw(path)
	if err != nil {
		return nil, err
	}
	return &fileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		ModTime: info.ModTime().Unix(),
		IsDir:   info.IsDir(),
	}, nil
}
