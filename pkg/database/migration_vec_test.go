package database

import (
	"context"
	"testing"
)

func TestMigrateToVec_MigratesChunksAndMarksIndexed(t *testing.T) {
	repo, cleanup := setupRepositoryTestDB(t)
	defer cleanup()

	if err := repo.db.Exec("DROP TABLE IF EXISTS vec_chunks").Error; err != nil {
		t.Fatalf("drop vec_chunks failed: %v", err)
	}
	if err := repo.db.Exec("CREATE TABLE vec_chunks (chunk_id INTEGER PRIMARY KEY, embedding BLOB NOT NULL)").Error; err != nil {
		t.Fatalf("create vec_chunks failed: %v", err)
	}

	file := File{Path: "migrate.md", Title: "migrate"}
	if err := repo.db.Create(&file).Error; err != nil {
		t.Fatalf("create file failed: %v", err)
	}

	chunks := []Chunk{
		{FileID: file.ID, Content: "c1", Heading: "h", EmbeddingBlob: floatsToBytes([]float32{1, 0, 0})},
		{FileID: file.ID, Content: "c2", Heading: "h", EmbeddingBlob: floatsToBytes([]float32{0, 1, 0})},
	}
	for i := range chunks {
		if err := repo.db.Create(&chunks[i]).Error; err != nil {
			t.Fatalf("create chunk %d failed: %v", i, err)
		}
	}

	m := &Manager{db: repo.db}
	if err := m.MigrateToVec(context.Background()); err != nil {
		t.Fatalf("MigrateToVec failed: %v", err)
	}

	var indexedCount int64
	if err := repo.db.Model(&Chunk{}).Where("vec_indexed = ?", true).Count(&indexedCount).Error; err != nil {
		t.Fatalf("count vec_indexed failed: %v", err)
	}
	if indexedCount != 2 {
		t.Fatalf("expected 2 indexed chunks, got %d", indexedCount)
	}

	var vecCount int64
	if err := repo.db.Raw("SELECT COUNT(*) FROM vec_chunks").Scan(&vecCount).Error; err != nil {
		t.Fatalf("count vec_chunks failed: %v", err)
	}
	if vecCount != 2 {
		t.Fatalf("expected 2 vec_chunks rows, got %d", vecCount)
	}
}

func TestMigrateToVec_IsIdempotent(t *testing.T) {
	repo, cleanup := setupRepositoryTestDB(t)
	defer cleanup()

	if err := repo.db.Exec("DROP TABLE IF EXISTS vec_chunks").Error; err != nil {
		t.Fatalf("drop vec_chunks failed: %v", err)
	}
	if err := repo.db.Exec("CREATE TABLE vec_chunks (chunk_id INTEGER PRIMARY KEY, embedding BLOB NOT NULL)").Error; err != nil {
		t.Fatalf("create vec_chunks failed: %v", err)
	}

	file := File{Path: "idem.md", Title: "idem"}
	if err := repo.db.Create(&file).Error; err != nil {
		t.Fatalf("create file failed: %v", err)
	}

	chunk := Chunk{
		FileID:        file.ID,
		Content:       "chunk",
		Heading:       "h",
		EmbeddingBlob: floatsToBytes([]float32{0.2, 0.4}),
	}
	if err := repo.db.Create(&chunk).Error; err != nil {
		t.Fatalf("create chunk failed: %v", err)
	}

	m := &Manager{db: repo.db}
	if err := m.MigrateToVec(context.Background()); err != nil {
		t.Fatalf("first MigrateToVec failed: %v", err)
	}
	if err := m.MigrateToVec(context.Background()); err != nil {
		t.Fatalf("second MigrateToVec failed: %v", err)
	}

	var vecCount int64
	if err := repo.db.Raw("SELECT COUNT(*) FROM vec_chunks").Scan(&vecCount).Error; err != nil {
		t.Fatalf("count vec_chunks failed: %v", err)
	}
	if vecCount != 1 {
		t.Fatalf("expected 1 vec_chunks row after rerun, got %d", vecCount)
	}
}
