package database

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

// VectorOperation represents a vector operation result
type VectorOperation struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// SaveEmbedding saves a vector embedding for a chunk
// Note: This implementation stores vectors as JSON for now
// In future versions, we'll use sqlite-vec extension for efficient similarity search
func (r *Repository) SaveEmbedding(chunkID uint, vector []float32, modelName string) error {
	// For now, store the vector directly as JSONB
	// GORM will handle the serialization of []float32 to JSON
	result := r.db.Model(&Chunk{}).
		Where("id = ?", chunkID).
		Updates(map[string]interface{}{
			"embedding":            vector,
			"embedding_model":      modelName,
			"embedding_created_at": "NOW()", // SQLite will handle this
		})

	return result.Error
}

// GetChunkEmbedding retrieves the embedding for a chunk
func (r *Repository) GetChunkEmbedding(chunkID uint) ([]float32, error) {
	var chunk Chunk
	err := r.db.Select("embedding").First(&chunk, chunkID).Error
	if err != nil {
		return nil, err
	}
	return chunk.Embedding, nil
}

// SearchSimilar performs a similarity search using cosine similarity
// For now, this is a naive implementation that loads all vectors and computes similarity
// In production with sqlite-vec, this will use the vector distance function
func (r *Repository) SearchSimilar(queryVector []float32, limit int) ([]SimilarChunk, error) {
	// Get all chunks that have embeddings
	var chunks []Chunk
	err := r.db.Where("embedding IS NOT NULL").Find(&chunks).Error
	if err != nil {
		return nil, err
	}

	// Calculate cosine similarity for each chunk
	results := make([]SimilarChunk, 0, len(chunks))
	for _, chunk := range chunks {
		if len(chunk.Embedding) != len(queryVector) {
			continue // Skip if dimensions don't match
		}

		similarity := cosineSimilarity(queryVector, chunk.Embedding)
		results = append(results, SimilarChunk{
			ChunkID:    chunk.ID,
			Content:    chunk.Content,
			Heading:    chunk.Heading,
			Similarity: similarity,
		})
	}

	// Sort by similarity (descending) using standard library
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// SimilarChunk represents a chunk with its similarity score
type SimilarChunk struct {
	ChunkID    uint    `json:"chunk_id"`
	Content    string  `json:"content"`
	Heading    string  `json:"heading"`
	Similarity float32 `json:"similarity"`
	File       *File   `json:"file,omitempty"`
}

// cosineSimilarity computes the cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// SaveEmbeddingBatch saves embeddings for multiple chunks in a single transaction
func (r *Repository) SaveEmbeddingBatch(embeddings []ChunkEmbedding) error {
	// Start transaction
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Defer rollback in case of error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, emb := range embeddings {
		err := tx.Model(&Chunk{}).
			Where("id = ?", emb.ChunkID).
			Updates(map[string]interface{}{
				"embedding":            emb.Vector,
				"embedding_model":      emb.ModelName,
				"embedding_created_at": "NOW()", // SQLite will handle this
			}).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save embedding for chunk %d: %w", emb.ChunkID, err)
		}
	}

	return tx.Commit().Error
}

// ChunkEmbedding represents a chunk with its embedding vector
type ChunkEmbedding struct {
	ChunkID   uint
	Vector    []float32
	ModelName string
}

// GetEmbeddingStats returns statistics about embeddings in the database
func (r *Repository) GetEmbeddingStats() (*EmbeddingStats, error) {
	var stats EmbeddingStats

	// Count total chunks
	r.db.Model(&Chunk{}).Count(&stats.TotalChunks)

	// Count embedded chunks
	r.db.Model(&Chunk{}).Where("embedding IS NOT NULL").Count(&stats.EmbeddedChunks)

	// Get unique models
	var models []string
	r.db.Model(&Chunk{}).Where("embedding_model IS NOT NULL AND embedding_model != ''").
		Distinct("embedding_model").
		Pluck("embedding_model", &models)
	stats.Models = models

	return &stats, nil
}

// EmbeddingStats represents statistics about embeddings
type EmbeddingStats struct {
	TotalChunks    int64    `json:"total_chunks"`
	EmbeddedChunks int64    `json:"embedded_chunks"`
	Models         []string `json:"models"`
}

// ToJSON converts the vector to JSON string
func (c *Chunk) ToJSON() (string, error) {
	data, err := json.Marshal(c.Embedding)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetEmbeddingDimension returns the dimension of the embedding vector
func (c *Chunk) GetEmbeddingDimension() int {
	if c.Embedding == nil {
		return 0
	}
	return len(c.Embedding)
}
