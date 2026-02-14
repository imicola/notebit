package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"notebit/pkg/indexing"
	"notebit/pkg/logger"

	"github.com/fsnotify/fsnotify"
)

// Service handles file system watching and automatic indexing
type Service struct {
	baseDir    string
	pipeline   *indexing.IndexingPipeline
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
func NewService(baseDir string, pipeline *indexing.IndexingPipeline) (*Service, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	// Default debounce delay: 500ms
	debounceDelay := 500 * time.Millisecond

	return &Service{
		baseDir:       baseDir,
		pipeline:      pipeline,
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
	if s.pipeline == nil {
		return
	}

	// Queue for async indexing
	s.pipeline.Enqueue(path, "", indexing.IndexOptions{
		SkipIfUnchanged:        true,
		FallbackToMetadataOnly: true,
	})
}

// handleRemove handles file deletion
func (s *Service) handleRemove(path string) {
	if s.pipeline == nil {
		return
	}

	repo := s.pipeline.Repository()
	if repo == nil {
		return
	}

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

// Helper functions

func isMarkdownFile(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".md"
}

func isTemporaryFile(path string) bool {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") && strings.HasSuffix(base, ".swp") {
		return true
	}
	if strings.HasSuffix(base, "~") || strings.HasSuffix(base, ".tmp") {
		return true
	}
	return false
}

func isInIgnoredDir(path string) bool {
	path = filepath.ToSlash(path)
	parts := strings.Split(path, "/")
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
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
