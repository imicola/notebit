package knowledge

import (
	"context"
	"fmt"
	"notebit/pkg/ai"
	"notebit/pkg/database"
	"notebit/pkg/files"
	"notebit/pkg/indexing"
)

const maxFindSimilarContentLength = 8000

// Service handles knowledge base operations (indexing, search)
type Service struct {
	fm       *files.Manager
	dbm      *database.Manager
	ai       *ai.Service
	pipeline *indexing.IndexingPipeline
}

// NewService creates a new knowledge service
func NewService(fm *files.Manager, dbm *database.Manager, ai *ai.Service, pipeline *indexing.IndexingPipeline) *Service {
	return &Service{
		fm:       fm,
		dbm:      dbm,
		ai:       ai,
		pipeline: pipeline,
	}
}

// IndexFileWithEmbedding indexes a file and generates embeddings for its chunks
func (s *Service) IndexFileWithEmbedding(path string) error {
	if s.pipeline == nil {
		return fmt.Errorf("indexing pipeline not initialized")
	}

	return s.pipeline.IndexFile(context.Background(), path, indexing.IndexOptions{
		ForceReindex:           true,
		FallbackToMetadataOnly: true,
	})
}

// ReindexAllWithEmbeddings reindexes all files with embeddings
func (s *Service) ReindexAllWithEmbeddings() (map[string]interface{}, error) {
	if s.pipeline == nil {
		return nil, fmt.Errorf("indexing pipeline not initialized")
	}

	filesList, err := s.fm.ListFiles()
	if err != nil {
		return nil, err
	}

	// Collect all markdown files
	var mdFiles []string
	collectFiles(filesList, &mdFiles)

	// Use pipeline's IndexAll for batch processing
	progress, err := s.pipeline.IndexAll(context.Background(), mdFiles, indexing.IndexOptions{
		ForceReindex:           true,
		FallbackToMetadataOnly: true,
	})
	if err != nil {
		return nil, err
	}

	// Wait for completion
	<-progress.Done

	return map[string]interface{}{
		"total":     progress.Total,
		"processed": progress.Processed,
		"failed":    progress.Errors,
	}, nil
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
func (s *Service) FindSimilar(content string, limit int) ([]SimilarNote, error) {
	// 1. Check if database is initialized
	if !s.dbm.IsInitialized() {
		return nil, fmt.Errorf("database not initialized")
	}

	// 2. Check if AI service is healthy
	status, err := s.ai.GetStatus()
	if err != nil || !status.ProviderHealthy {
		return nil, fmt.Errorf("AI service not available")
	}

	// 3. Generate embedding for query content
	if len(content) > maxFindSimilarContentLength {
		runes := []rune(content)
		if len(runes) > maxFindSimilarContentLength {
			content = string(runes[:maxFindSimilarContentLength])
		}
	}
	resp, err := s.ai.GenerateEmbedding(content)
	if err != nil {
		return nil, err
	}

	// 4. Search similar chunks
	chunks, err := s.dbm.Repository().SearchSimilar(resp.Embedding, limit)
	if err != nil {
		return nil, err
	}

	// 5. Enrich with file information
	results := make([]SimilarNote, 0, len(chunks))
	for _, chunk := range chunks {
		if chunk.File == nil {
			continue
		}
		results = append(results, SimilarNote{
			Path:       chunk.File.Path,
			Title:      chunk.File.Title,
			Content:    chunk.Content,
			Heading:    chunk.Heading,
			Similarity: chunk.Similarity,
			ChunkID:    chunk.ChunkID,
		})
	}

	return results, nil
}

// GetSimilarityStatus returns the availability status of semantic search
func (s *Service) GetSimilarityStatus() (map[string]interface{}, error) {
	dbInitialized := s.dbm.IsInitialized()

	var aiStatus *ai.ServiceStatus
	if dbInitialized {
		aiStatus, _ = s.ai.GetStatus()
	}

	// Get embedding stats
	var stats *database.EmbeddingStats
	if dbInitialized {
		stats, _ = s.dbm.Repository().GetEmbeddingStats()
	}

	available := dbInitialized && aiStatus != nil && aiStatus.ProviderHealthy

	var embeddedChunks int64
	var totalChunks int64
	vectorEngine := "unknown"
	if stats != nil {
		embeddedChunks = stats.EmbeddedChunks
		totalChunks = stats.TotalChunks
	}
	if dbInitialized && s.dbm.Repository() != nil {
		engine := s.dbm.Repository().GetVectorEngine()
		if engine != "" {
			vectorEngine = engine
		}
	}

	return map[string]interface{}{
		"available":      available,
		"db_initialized": dbInitialized,
		"ai_healthy":     aiStatus != nil && aiStatus.ProviderHealthy,
		"indexed_chunks": embeddedChunks,
		"total_chunks":   totalChunks,
		"vector_engine":  vectorEngine,
	}, nil
}
