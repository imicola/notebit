package knowledge

import (
	"fmt"
	"notebit/pkg/ai"
	"notebit/pkg/database"
	"notebit/pkg/files"
	"os"
	"path/filepath"
)

const maxFindSimilarContentLength = 8000

// Service handles knowledge base operations (indexing, search)
type Service struct {
	fm  *files.Manager
	dbm *database.Manager
	ai  *ai.Service
}

// NewService creates a new knowledge service
func NewService(fm *files.Manager, dbm *database.Manager, ai *ai.Service) *Service {
	return &Service{
		fm:  fm,
		dbm: dbm,
		ai:  ai,
	}
}

// IndexFileWithEmbedding indexes a file and generates embeddings for its chunks
func (s *Service) IndexFileWithEmbedding(path string) error {
	if !s.dbm.IsInitialized() {
		return fmt.Errorf("database not initialized")
	}

	// Read file content
	content, err := s.fm.ReadFile(path)
	if err != nil {
		return err
	}

	// Get file stats
	fullPath := filepath.Join(s.fm.GetBasePath(), path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}

	// Process document to get chunks with embeddings
	chunks, err := s.ai.ProcessDocument(content.Content)
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
	repo := s.dbm.Repository()
	return repo.IndexFileWithChunks(path, content.Content, info.ModTime().Unix(), info.Size(), dbChunks)
}

// ReindexAllWithEmbeddings reindexes all files with embeddings
func (s *Service) ReindexAllWithEmbeddings() (map[string]interface{}, error) {
	if !s.dbm.IsInitialized() {
		return nil, fmt.Errorf("database not initialized")
	}

	filesList, err := s.fm.ListFiles()
	if err != nil {
		return nil, err
	}

	// Collect all markdown files
	var mdFiles []string
	collectFiles(filesList, &mdFiles)

	results := map[string]interface{}{
		"total":     len(mdFiles),
		"processed": 0,
		"failed":    0,
		"errors":    []string{},
	}

	for _, path := range mdFiles {
		if err := s.IndexFileWithEmbedding(path); err != nil {
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
	vectorEngine := ""
	if stats != nil {
		embeddedChunks = stats.EmbeddedChunks
		totalChunks = stats.TotalChunks
	}
	if dbInitialized {
		vectorEngine = s.dbm.Repository().GetVectorEngine()
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
