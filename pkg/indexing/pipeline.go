package indexing

import (
	"context"
	"errors"
	"fmt"
	"notebit/pkg/ai"
	"notebit/pkg/config"
	"notebit/pkg/database"
	"notebit/pkg/files"
	"notebit/pkg/logger"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// IndexingPipeline provides a unified, thread-safe interface for file indexing
// with automatic deduplication and configurable fallback strategies
type IndexingPipeline struct {
	ai   *ai.Service
	repo *database.Repository
	fm   *files.Manager

	// Worker pool for concurrent indexing
	workQueue chan *IndexJob
	workers   int
	isStarted bool
	mu        sync.Mutex

	// Deduplication map to prevent concurrent indexing of the same file
	inProgress sync.Map // map[string]bool
}

var errPipelineStopped = errors.New("indexing pipeline not started")

// IndexJob represents a single file indexing job
type IndexJob struct {
	Path    string
	Content string // Optional: if empty, will be read from disk
	Opts    IndexOptions
	ErrChan chan error // Optional: channel to receive error result
}

// IndexOptions controls indexing behavior
type IndexOptions struct {
	// SkipIfUnchanged checks FileNeedsIndexing before processing
	SkipIfUnchanged bool

	// FallbackToMetadataOnly enables graceful degradation on AI failure
	FallbackToMetadataOnly bool

	// ForceReindex ignores hash comparison and always reindexes
	ForceReindex bool
}

// IndexProgress tracks multi-file indexing progress
type IndexProgress struct {
	Total     int
	Processed atomic.Int64
	Errors    atomic.Int64
	Done      chan struct{} // Closed when indexing completes
}

// NewPipeline creates a new indexing pipeline with worker pool
func NewPipeline(ai *ai.Service, repo *database.Repository, fm *files.Manager) *IndexingPipeline {
	cfg := config.Get()
	workers := cfg.Indexing.WorkerCount
	if workers <= 0 {
		workers = 4
	}
	queueSize := cfg.Indexing.QueueSize
	if queueSize <= 0 {
		queueSize = 256
	}

	p := &IndexingPipeline{
		ai:        ai,
		repo:      repo,
		fm:        fm,
		workQueue: make(chan *IndexJob, queueSize),
		workers:   workers,
	}

	return p
}

// Start initializes the worker pool
func (p *IndexingPipeline) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isStarted {
		return
	}

	for i := 0; i < p.workers; i++ {
		go p.worker(i)
	}

	p.isStarted = true
	logger.InfoWithFields(context.Background(), map[string]interface{}{
		"workers":    p.workers,
		"queue_size": cap(p.workQueue),
	}, "Indexing pipeline started")
}

// worker processes indexing jobs from the queue
func (p *IndexingPipeline) worker(id int) {
	for job := range p.workQueue {
		err := p.processJob(job)
		if job.ErrChan != nil {
			job.ErrChan <- err
			close(job.ErrChan)
		}
	}
}

// processJob handles a single indexing job with deduplication
func (p *IndexingPipeline) processJob(job *IndexJob) error {
	// Deduplication: skip if already in progress
	if _, loaded := p.inProgress.LoadOrStore(job.Path, true); loaded {
		logger.InfoWithFields(context.Background(), map[string]interface{}{
			"path": job.Path,
		}, "Skipping duplicate indexing job")
		return nil
	}
	defer p.inProgress.Delete(job.Path)

	ctx := context.Background()

	// Read content if not provided
	content := job.Content
	if content == "" {
		noteContent, err := p.fm.ReadFile(job.Path)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		content = noteContent.Content
	}

	// Get file stats
	fullPath := filepath.Join(p.fm.GetBasePath(), job.Path)
	stat, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	// Check if indexing is needed
	if job.Opts.SkipIfUnchanged && !job.Opts.ForceReindex {
		needsIndex, err := p.repo.FileNeedsIndexing(job.Path, content)
		if err != nil {
			logger.WarnWithFields(ctx, map[string]interface{}{
				"path":  job.Path,
				"error": err.Error(),
			}, "FileNeedsIndexing check failed, proceeding anyway")
		} else if !needsIndex {
			logger.InfoWithFields(ctx, map[string]interface{}{
				"path": job.Path,
			}, "File unchanged, skipping indexing")
			return nil
		}
	}

	// Try full indexing with embeddings
	err = p.indexWithEmbeddings(ctx, job.Path, content, stat.ModTime().Unix(), stat.Size())
	if err == nil {
		return nil
	}

	// If fallback disabled, return error
	if !job.Opts.FallbackToMetadataOnly {
		return err
	}

	logger.WarnWithFields(ctx, map[string]interface{}{
		"path":  job.Path,
		"error": err.Error(),
	}, "Embedding generation failed, trying chunking only")

	// Fallback 1: Chunk without embeddings
	if err := p.indexWithChunking(ctx, job.Path, content, stat.ModTime().Unix(), stat.Size()); err != nil {
		logger.WarnWithFields(ctx, map[string]interface{}{
			"path":  job.Path,
			"error": err.Error(),
		}, "Chunking failed, indexing metadata only")

		// Fallback 2: Metadata only
		return p.repo.IndexFile(job.Path, content, stat.ModTime().Unix(), stat.Size())
	}

	return nil
}

// indexWithEmbeddings performs full indexing with AI embeddings
func (p *IndexingPipeline) indexWithEmbeddings(ctx context.Context, path, content string, modTime, size int64) error {
	// Process document: chunking + embeddings
	chunks, err := p.ai.ProcessDocument(content)
	if err != nil {
		return fmt.Errorf("ProcessDocument failed: %w", err)
	}

	if len(chunks) == 0 {
		return fmt.Errorf("no chunks generated")
	}

	// Convert to database ChunkInput
	chunkInputs := make([]database.ChunkInput, len(chunks))
	for i, chunk := range chunks {
		chunkInputs[i] = database.ChunkInput{
			Content:        chunk.Content,
			Heading:        chunk.Heading,
			Embedding:      chunk.Embedding,
			EmbeddingModel: chunk.ModelName,
		}
	}

	// Index file with chunks
	if err := p.repo.IndexFileWithChunks(path, content, modTime, size, chunkInputs); err != nil {
		return fmt.Errorf("IndexFileWithChunks failed: %w", err)
	}

	logger.InfoWithFields(ctx, map[string]interface{}{
		"path":   path,
		"chunks": len(chunks),
		"model":  chunks[0].ModelName,
	}, "File indexed with embeddings")

	return nil
}

// indexWithChunking indexes file with chunks but without embeddings
func (p *IndexingPipeline) indexWithChunking(ctx context.Context, path, content string, modTime, size int64) error {
	// Chunk text without embeddings
	chunks, err := p.ai.ChunkText(content)
	if err != nil {
		return fmt.Errorf("ChunkText failed: %w", err)
	}

	if len(chunks) == 0 {
		return fmt.Errorf("no chunks generated")
	}

	// Convert to database ChunkInput (no embeddings)
	chunkInputs := make([]database.ChunkInput, len(chunks))
	for i, chunk := range chunks {
		chunkInputs[i] = database.ChunkInput{
			Content: chunk.Content,
			Heading: chunk.Heading,
		}
	}

	// Index file with chunks
	if err := p.repo.IndexFileWithChunks(path, content, modTime, size, chunkInputs); err != nil {
		return fmt.Errorf("IndexFileWithChunks failed: %w", err)
	}

	logger.InfoWithFields(ctx, map[string]interface{}{
		"path":   path,
		"chunks": len(chunks),
	}, "File indexed without embeddings")

	return nil
}

// Enqueue submits a file for async indexing (non-blocking)
func (p *IndexingPipeline) Enqueue(path, content string, opts IndexOptions) {
	p.mu.Lock()
	started := p.isStarted
	p.mu.Unlock()
	if !started {
		logger.Warn("Indexing pipeline not started, dropping job for: %s", path)
		return
	}

	job := &IndexJob{
		Path:    path,
		Content: content,
		Opts:    opts,
	}

	if err := p.safeEnqueueNonBlocking(job); err != nil {
		if errors.Is(err, errPipelineStopped) {
			logger.Warn("Indexing pipeline stopped during enqueue, dropping job for: %s", path)
			return
		}
		logger.WarnWithFields(context.Background(), map[string]interface{}{
			"path": path,
		}, "Indexing queue full, dropping job")
	}
}

// IndexFile indexes a single file synchronously
func (p *IndexingPipeline) IndexFile(ctx context.Context, path string, opts IndexOptions) error {
	p.mu.Lock()
	started := p.isStarted
	p.mu.Unlock()
	if !started {
		return errPipelineStopped
	}

	job := &IndexJob{
		Path:    path,
		Opts:    opts,
		ErrChan: make(chan error, 1),
	}

	if err := p.safeEnqueueWithTimeout(ctx, job, 5*time.Second); err != nil {
		return err
	}

	// Wait for result
	select {
	// Wait for result
	case err := <-job.ErrChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IndexContent indexes file content directly (synchronous)
func (p *IndexingPipeline) IndexContent(ctx context.Context, path, content string, opts IndexOptions) error {
	p.mu.Lock()
	started := p.isStarted
	p.mu.Unlock()
	if !started {
		return errPipelineStopped
	}

	job := &IndexJob{
		Path:    path,
		Content: content,
		Opts:    opts,
		ErrChan: make(chan error, 1),
	}

	if err := p.safeEnqueueWithTimeout(ctx, job, 5*time.Second); err != nil {
		return err
	}

	select {
	case err := <-job.ErrChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IndexAll indexes multiple files concurrently
func (p *IndexingPipeline) IndexAll(ctx context.Context, paths []string, opts IndexOptions) (*IndexProgress, error) {
	p.mu.Lock()
	started := p.isStarted
	p.mu.Unlock()
	if !started {
		return nil, errPipelineStopped
	}

	progress := &IndexProgress{
		Total: len(paths),
		Done:  make(chan struct{}),
	}

	go func() {
		defer close(progress.Done)

		var wg sync.WaitGroup

		// Limit concurrent enqueue goroutines to avoid resource pressure
		// when processing 10k+ files. Using a semaphore pattern.
		maxConcurrentEnqueue := 50 // Reasonable limit for concurrent submission
		sem := make(chan struct{}, maxConcurrentEnqueue)

	Loop:
		for _, path := range paths {
			if ctx.Err() != nil {
				break Loop
			}

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				break Loop
			}

			wg.Add(1)
			go func(filePath string) {
				defer wg.Done()
				defer func() { <-sem }() // Release semaphore

				job := &IndexJob{
					Path:    filePath,
					Opts:    opts,
					ErrChan: make(chan error, 1),
				}

				if err := p.safeEnqueueWithTimeout(ctx, job, 5*time.Second); err != nil {
					if !errors.Is(err, context.Canceled) {
						progress.Errors.Add(1)
						progress.Processed.Add(1)
					}
					return
				}

				err := <-job.ErrChan
				if err != nil {
					progress.Errors.Add(1)
				}
				progress.Processed.Add(1)
			}(path)
		}

		wg.Wait()
	}()

	return progress, nil
}

func (p *IndexingPipeline) safeEnqueueNonBlocking(job *IndexJob) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errPipelineStopped
		}
	}()

	select {
	case p.workQueue <- job:
		return nil
	default:
		return fmt.Errorf("indexing queue full")
	}
}

func (p *IndexingPipeline) safeEnqueueWithTimeout(ctx context.Context, job *IndexJob, timeout time.Duration) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errPipelineStopped
		}
	}()

	select {
	case p.workQueue <- job:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("indexing queue full, timeout")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop gracefully shuts down the worker pool
func (p *IndexingPipeline) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isStarted {
		return
	}

	close(p.workQueue)
	p.isStarted = false

	logger.Info("Indexing pipeline stopped")
}

// Repository exposes underlying repository for operations not covered by queue jobs (e.g. delete sync).
func (p *IndexingPipeline) Repository() *database.Repository {
	if p == nil {
		return nil
	}
	return p.repo
}
