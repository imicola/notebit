package database

import "testing"

func TestSearchSimilarBatch_ReturnsPerQueryResults(t *testing.T) {
	repo, cleanup := setupVectorEngineTestDB(t)
	defer cleanup()

	file := File{Path: "batch.md", Title: "batch"}
	if err := repo.db.Create(&file).Error; err != nil {
		t.Fatalf("create file failed: %v", err)
	}

	vectors := [][]float32{
		{1, 0},
		{0.8, 0.2},
		{0, 1},
	}

	for i, vec := range vectors {
		chunk := Chunk{
			FileID:        file.ID,
			Content:       "chunk",
			Heading:       "h",
			Embedding:     vec,
			EmbeddingBlob: floatsToBytes(vec),
		}
		if err := repo.db.Create(&chunk).Error; err != nil {
			t.Fatalf("create chunk %d failed: %v", i, err)
		}
	}

	queries := [][]float32{{1, 0}, {0, 1}}
	results, err := repo.SearchSimilarBatch(queries, 2)
	if err != nil {
		t.Fatalf("SearchSimilarBatch failed: %v", err)
	}

	if len(results) != len(queries) {
		t.Fatalf("expected %d result groups, got %d", len(queries), len(results))
	}

	for i := range results {
		if len(results[i]) == 0 {
			t.Fatalf("expected non-empty results for query %d", i)
		}
		if results[i][0].File == nil {
			t.Fatalf("expected file preload for query %d", i)
		}
	}
}

func TestSearchSimilarBatch_ConsistentWithSingleSearch(t *testing.T) {
	repo, cleanup := setupVectorEngineTestDB(t)
	defer cleanup()

	file := File{Path: "consistency.md", Title: "consistency"}
	if err := repo.db.Create(&file).Error; err != nil {
		t.Fatalf("create file failed: %v", err)
	}

	vectors := [][]float32{
		{1, 0, 0},
		{0.8, 0.2, 0},
		{0, 1, 0},
		{0, 0, 1},
	}

	for i, vec := range vectors {
		chunk := Chunk{
			FileID:        file.ID,
			Content:       "chunk",
			Heading:       "h",
			Embedding:     vec,
			EmbeddingBlob: floatsToBytes(vec),
		}
		if err := repo.db.Create(&chunk).Error; err != nil {
			t.Fatalf("create chunk %d failed: %v", i, err)
		}
	}

	queries := [][]float32{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
	batchResults, err := repo.SearchSimilarBatch(queries, 2)
	if err != nil {
		t.Fatalf("SearchSimilarBatch failed: %v", err)
	}

	for i, query := range queries {
		singleResults, err := repo.SearchSimilar(query, 2)
		if err != nil {
			t.Fatalf("SearchSimilar failed for query %d: %v", i, err)
		}
		if len(singleResults) == 0 || len(batchResults[i]) == 0 {
			t.Fatalf("expected non-empty results for query %d", i)
		}
		if singleResults[0].ChunkID != batchResults[i][0].ChunkID {
			t.Fatalf("query %d top-1 mismatch: single=%d batch=%d", i, singleResults[0].ChunkID, batchResults[i][0].ChunkID)
		}
	}
}
