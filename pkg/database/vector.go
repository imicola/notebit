package database

import (
	"encoding/binary"
	"math"
)

// GetChunkEmbedding retrieves the embedding for a chunk
func (r *Repository) GetChunkEmbedding(chunkID uint) ([]float32, error) {
	var chunk Chunk
	// Try fetching blob first
	err := r.db.Select("embedding_blob").First(&chunk, chunkID).Error
	if err == nil && len(chunk.EmbeddingBlob) > 0 {
		floats := bytesToFloats(chunk.EmbeddingBlob)
		return floats, nil
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
	if r.vectorEngine == nil {
		r.vectorEngine = NewBruteForceVectorEngine()
	}

	results, err := r.vectorEngine.Search(r, queryVector, limit)
	if err == nil {
		return results, nil
	}

	if r.vectorEngine.Name() == VectorEngineBruteForce {
		return nil, err
	}

	fallback := NewBruteForceVectorEngine()
	r.vectorEngine = fallback
	return fallback.Search(r, queryVector, limit)
}

// SearchSimilarBatch performs similarity search for multiple query vectors.
// Refactored to use the pluggable VectorEngine and avoid O(N) memory usage.
// For large datasets (10k+ notes), this prevents loading all chunks into memory.
func (r *Repository) SearchSimilarBatch(queryVectors [][]float32, limit int) ([][]SimilarChunk, error) {
	if len(queryVectors) == 0 {
		return [][]SimilarChunk{}, nil
	}
	if limit <= 0 {
		limit = 10
	}

	// Ensure vector engine is initialized
	if r.vectorEngine == nil {
		r.vectorEngine = NewBruteForceVectorEngine()
	}

	results := make([][]SimilarChunk, len(queryVectors))

	// TODO: Implement optimized batch search for sqlite-vec engine
	// Current implementation iterates queries but uses the efficient Search() method
	// which avoids loading all chunks into memory (uses streaming or indexed search)
	for i, query := range queryVectors {
		if len(query) == 0 {
			// Skip invalid query vectors
			results[i] = []SimilarChunk{}
			continue
		}

		// Use the optimized single-search method via VectorEngine
		matches, err := r.vectorEngine.Search(r, query, limit)
		if err != nil {
			// Fail fast on error - partial results could be misleading
			return nil, err
		}
		results[i] = matches
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

// GetEmbeddingStats returns statistics about embeddings in the database
func (r *Repository) GetEmbeddingStats() (*EmbeddingStats, error) {
	var stats EmbeddingStats

	// Count total chunks
	r.db.Model(&Chunk{}).Count(&stats.TotalChunks)

	// Count embedded chunks
	r.db.Model(&Chunk{}).Where("embedding_blob IS NOT NULL AND length(embedding_blob) > 0").Count(&stats.EmbeddedChunks)

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

func floatsToBytes(floats []float32) []byte {
	bytes := make([]byte, len(floats)*4)
	for i, f := range floats {
		bits := math.Float32bits(f)
		binary.LittleEndian.PutUint32(bytes[i*4:], bits)
	}
	return bytes
}

func bytesToFloats(bytes []byte) []float32 {
	if len(bytes) == 0 || len(bytes)%4 != 0 {
		return nil
	}
	floats := make([]float32, len(bytes)/4)
	for i := 0; i < len(floats); i++ {
		bits := binary.LittleEndian.Uint32(bytes[i*4:])
		floats[i] = math.Float32frombits(bits)
	}
	return floats
}
