package database

import (
	"fmt"
	"sort"
)

type SQLiteVecEngine struct{}

func NewSQLiteVecEngine() *SQLiteVecEngine {
	return &SQLiteVecEngine{}
}

func (e *SQLiteVecEngine) Name() string {
	return VectorEngineSQLiteVec
}

func (e *SQLiteVecEngine) Search(repo *Repository, queryVector []float32, limit int) ([]SimilarChunk, error) {
	if limit <= 0 {
		limit = 10
	}

	type vecResult struct {
		ChunkID  uint
		Distance float32
	}

	queryBlob := floatsToBytes(queryVector)
	rows, err := repo.db.Raw(
		"SELECT chunk_id, distance FROM vec_chunks WHERE embedding MATCH ? ORDER BY distance ASC LIMIT ?",
		queryBlob,
		limit,
	).Rows()
	if err != nil {
		return nil, fmt.Errorf("sqlite-vec query failed: %w", err)
	}
	defer rows.Close()

	ordered := make([]vecResult, 0, limit)
	idOrder := make(map[uint]int, limit)
	for rows.Next() {
		var item vecResult
		if err := rows.Scan(&item.ChunkID, &item.Distance); err != nil {
			return nil, err
		}
		idOrder[item.ChunkID] = len(ordered)
		ordered = append(ordered, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ordered) == 0 {
		return []SimilarChunk{}, nil
	}

	ids := make([]uint, 0, len(ordered))
	for _, item := range ordered {
		ids = append(ids, item.ChunkID)
	}

	var chunks []Chunk
	if err := repo.db.Preload("File").Where("id IN ?", ids).Find(&chunks).Error; err != nil {
		return nil, err
	}

	distanceMap := make(map[uint]float32, len(ordered))
	for _, item := range ordered {
		distanceMap[item.ChunkID] = item.Distance
	}

	results := make([]SimilarChunk, 0, len(chunks))
	for _, chunk := range chunks {
		distance := distanceMap[chunk.ID]
		similarity := float32(1.0 / (1.0 + distance))
		results = append(results, SimilarChunk{
			ChunkID:    chunk.ID,
			Content:    chunk.Content,
			Heading:    chunk.Heading,
			Similarity: similarity,
			File:       chunk.File,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return idOrder[results[i].ChunkID] < idOrder[results[j].ChunkID]
	})

	return results, nil
}
