package database

import (
	"encoding/binary"
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
	// Store as both JSON (legacy) and Blob (optimized)
	blob := floatsToBytes(vector)

	result := r.db.Model(&Chunk{}).
		Where("id = ?", chunkID).
		Updates(map[string]interface{}{
			"embedding":            vector,
			"embedding_blob":       blob,
			"embedding_model":      modelName,
			"embedding_created_at": "NOW()", // SQLite will handle this
		})

	if result.Error == nil {
		r.invalidateVectorCache()
	}
	return result.Error
}

// GetChunkEmbedding retrieves the embedding for a chunk
func (r *Repository) GetChunkEmbedding(chunkID uint) ([]float32, error) {
	var chunk Chunk
	// Try fetching blob first
	err := r.db.Select("embedding_blob").First(&chunk, chunkID).Error
	if err == nil && len(chunk.EmbeddingBlob) > 0 {
		return bytesToFloats(chunk.EmbeddingBlob), nil
	}

	// Fallback to JSON embedding
	err = r.db.Select("embedding").First(&chunk, chunkID).Error
	if err != nil {
		return nil, err
	}
	return chunk.Embedding, nil
}

// SearchSimilar performs a similarity search using cosine similarity
// For now, this is a naive implementation that loads all vectors and computes similarity
// In production with sqlite-vec, this will use the vector distance function
func (r *Repository) SearchSimilar(queryVector []float32, limit int) ([]SimilarChunk, error) {
	cache, err := r.getVectorCache()
	if err != nil {
		return nil, err
	}

	// 2. Calculate cosine similarity for each chunk
	type ScoredChunk struct {
		ID         uint
		Similarity float32
	}

	scores := make([]ScoredChunk, 0, len(cache))

	for id, vec := range cache {
		if len(vec) != len(queryVector) {
			continue
		}

		similarity := cosineSimilarity(queryVector, vec)
		scores = append(scores, ScoredChunk{
			ID:         id,
			Similarity: similarity,
		})
	}

	// 3. Sort by similarity (descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Similarity > scores[j].Similarity
	})

	// 4. Apply limit
	if limit > 0 && len(scores) > limit {
		scores = scores[:limit]
	}

	if len(scores) == 0 {
		return []SimilarChunk{}, nil
	}

	// 5. Fetch full content for top K results
	topIDs := make([]uint, len(scores))
	scoreMap := make(map[uint]float32)
	for i, s := range scores {
		topIDs[i] = s.ID
		scoreMap[s.ID] = s.Similarity
	}

	var fullChunks []Chunk
	err = r.db.Preload("File").
		Where("id IN ?", topIDs).
		Find(&fullChunks).Error
	if err != nil {
		return nil, err
	}

	// 6. Construct final result
	results := make([]SimilarChunk, 0, len(fullChunks))
	for _, chunk := range fullChunks {
		results = append(results, SimilarChunk{
			ChunkID:    chunk.ID,
			Content:    chunk.Content,
			Heading:    chunk.Heading,
			Similarity: scoreMap[chunk.ID],
			File:       chunk.File,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

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
		blob := floatsToBytes(emb.Vector)
		err := tx.Model(&Chunk{}).
			Where("id = ?", emb.ChunkID).
			Updates(map[string]interface{}{
				"embedding":            emb.Vector,
				"embedding_blob":       blob,
				"embedding_model":      emb.ModelName,
				"embedding_created_at": "NOW()", // SQLite will handle this
			}).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save embedding for chunk %d: %w", emb.ChunkID, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	r.invalidateVectorCache()
	return nil
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

func floatsToBytes(floats []float32) []byte {
	bytes := make([]byte, len(floats)*4)
	for i, f := range floats {
		bits := math.Float32bits(f)
		binary.LittleEndian.PutUint32(bytes[i*4:], bits)
	}
	return bytes
}

func bytesToFloats(bytes []byte) []float32 {
	if len(bytes)%4 != 0 {
		return nil
	}
	floats := make([]float32, len(bytes)/4)
	for i := 0; i < len(floats); i++ {
		bits := binary.LittleEndian.Uint32(bytes[i*4:])
		floats[i] = math.Float32frombits(bits)
	}
	return floats
}

type chunkVector struct {
	ID            uint   `gorm:"primarykey"`
	EmbeddingBlob []byte `gorm:"type:blob"`
}

func (r *Repository) invalidateVectorCache() {
	r.vectorCacheMu.Lock()
	r.vectorCache = nil
	r.vectorCacheLoaded = false
	r.vectorCacheMu.Unlock()
}

func (r *Repository) loadVectorCache() error {
	r.vectorCacheMu.Lock()
	defer r.vectorCacheMu.Unlock()
	if r.vectorCacheLoaded {
		return nil
	}

	var chunkVectors []chunkVector
	err := r.db.Model(&Chunk{}).
		Select("id, embedding_blob").
		Where("embedding_blob IS NOT NULL").
		Scan(&chunkVectors).Error
	if err != nil {
		return err
	}

	cache := make(map[uint][]float32, len(chunkVectors))
	for _, cv := range chunkVectors {
		vec := bytesToFloats(cv.EmbeddingBlob)
		if vec == nil {
			continue
		}
		cache[cv.ID] = vec
	}

	r.vectorCache = cache
	r.vectorCacheLoaded = true
	return nil
}

func (r *Repository) getVectorCache() (map[uint][]float32, error) {
	r.vectorCacheMu.RLock()
	if r.vectorCacheLoaded {
		cache := r.vectorCache
		r.vectorCacheMu.RUnlock()
		return cache, nil
	}
	r.vectorCacheMu.RUnlock()

	if err := r.loadVectorCache(); err != nil {
		return nil, err
	}

	r.vectorCacheMu.RLock()
	cache := r.vectorCache
	r.vectorCacheMu.RUnlock()
	return cache, nil
}
