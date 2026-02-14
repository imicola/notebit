package database

import (
	"container/heap"
	"sort"
)

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
	if r == nil {
		return ""
	}
	if r.vectorEngine == nil {
		r.vectorEngine = NewBruteForceVectorEngine()
	}
	return r.vectorEngine.Name()
}

type scoredChunk struct {
	ID         uint
	Similarity float32
}

type scoredChunkHeap []scoredChunk

func (h scoredChunkHeap) Len() int            { return len(h) }
func (h scoredChunkHeap) Less(i, j int) bool  { return h[i].Similarity < h[j].Similarity }
func (h scoredChunkHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *scoredChunkHeap) Push(x interface{}) { *h = append(*h, x.(scoredChunk)) }
func (h *scoredChunkHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

func (e *BruteForceVectorEngine) Search(repo *Repository, queryVector []float32, limit int) ([]SimilarChunk, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := repo.db.Model(&Chunk{}).
		Select("id, embedding_blob").
		Where("embedding_blob IS NOT NULL AND length(embedding_blob) > 0").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	topK := &scoredChunkHeap{}
	heap.Init(topK)

	for rows.Next() {
		var id uint
		var blob []byte
		if err := rows.Scan(&id, &blob); err != nil {
			return nil, err
		}

		vec := bytesToFloats(blob)
		if len(vec) != len(queryVector) {
			continue
		}

		score := scoredChunk{ID: id, Similarity: cosineSimilarity(queryVector, vec)}
		if topK.Len() < limit {
			heap.Push(topK, score)
			continue
		}

		if topK.Len() > 0 && score.Similarity > (*topK)[0].Similarity {
			heap.Pop(topK)
			heap.Push(topK, score)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if topK.Len() == 0 {
		return []SimilarChunk{}, nil
	}

	scores := make([]scoredChunk, topK.Len())
	for i := len(scores) - 1; i >= 0; i-- {
		scores[i] = heap.Pop(topK).(scoredChunk)
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Similarity > scores[j].Similarity
	})

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
