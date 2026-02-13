package database

import "sort"

const (
	VectorEngineBruteForce = "brute-force"
	VectorEngineSQLiteVec  = "sqlite-vec"
)

// VectorSearchEngine defines a pluggable vector retrieval backend.
// Current default implementation is brute-force cosine search over cached embeddings.
type VectorSearchEngine interface {
	Search(repo *Repository, queryVector []float32, limit int) ([]SimilarChunk, error)
	Name() string
}

// BruteForceVectorEngine is the default in-process search implementation.
type BruteForceVectorEngine struct{}

func NewBruteForceVectorEngine() *BruteForceVectorEngine {
	return &BruteForceVectorEngine{}
}

func (e *BruteForceVectorEngine) Name() string {
	return VectorEngineBruteForce
}

// SetVectorEngine selects a vector search engine by name.
// Returns the effective engine name (falls back to brute-force when unsupported).
func (r *Repository) SetVectorEngine(name string) string {
	switch name {
	case VectorEngineSQLiteVec:
		r.vectorEngine = NewSQLiteVecEngine()
	default:
		r.vectorEngine = NewBruteForceVectorEngine()
	}
	return r.vectorEngine.Name()
}

// GetVectorEngine returns the current vector search engine name.
func (r *Repository) GetVectorEngine() string {
	if r.vectorEngine == nil {
		r.vectorEngine = NewBruteForceVectorEngine()
	}
	return r.vectorEngine.Name()
}

type scoredChunk struct {
	ID         uint
	Similarity float32
}

func (e *BruteForceVectorEngine) Search(repo *Repository, queryVector []float32, limit int) ([]SimilarChunk, error) {
	cache, err := repo.getVectorCache()
	if err != nil {
		return nil, err
	}

	scores := make([]scoredChunk, 0, len(cache))
	for id, vec := range cache {
		if len(vec) != len(queryVector) {
			continue
		}
		scores = append(scores, scoredChunk{
			ID:         id,
			Similarity: cosineSimilarity(queryVector, vec),
		})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Similarity > scores[j].Similarity
	})

	if limit > 0 && len(scores) > limit {
		scores = scores[:limit]
	}
	if len(scores) == 0 {
		return []SimilarChunk{}, nil
	}

	topIDs := make([]uint, len(scores))
	scoreMap := make(map[uint]float32, len(scores))
	for i, s := range scores {
		topIDs[i] = s.ID
		scoreMap[s.ID] = s.Similarity
	}

	var fullChunks []Chunk
	if err := repo.db.Preload("File").
		Where("id IN ?", topIDs).
		Find(&fullChunks).Error; err != nil {
		return nil, err
	}

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
