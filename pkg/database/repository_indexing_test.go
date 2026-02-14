package database

import (
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupRepositoryTestDB(t *testing.T) (*Repository, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "notebit-repo-test-*")
	if err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(tmpDir, "repo.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&File{}, &Chunk{}); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	repo := &Repository{
		db:           db,
		vectorEngine: NewBruteForceVectorEngine(),
	}

	cleanup := func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		_ = os.RemoveAll(tmpDir)
	}
	return repo, cleanup
}

func TestFileNeedsIndexing_WhenOnlyMetadataIndexed(t *testing.T) {
	repo, cleanup := setupRepositoryTestDB(t)
	defer cleanup()

	content := "# Title\n\ncontent"
	if err := repo.IndexFile("a.md", content, 1, int64(len(content))); err != nil {
		t.Fatalf("index metadata failed: %v", err)
	}

	needs, err := repo.FileNeedsIndexing("a.md", content)
	if err != nil {
		t.Fatalf("FileNeedsIndexing failed: %v", err)
	}
	if !needs {
		t.Fatalf("expected reindex needed when chunks/embeddings are missing")
	}
}

func TestFileNeedsIndexing_WhenContentChanged(t *testing.T) {
	repo, cleanup := setupRepositoryTestDB(t)
	defer cleanup()

	content := "# Title\n\nold"
	if err := repo.IndexFile("b.md", content, 1, int64(len(content))); err != nil {
		t.Fatalf("index metadata failed: %v", err)
	}

	needs, err := repo.FileNeedsIndexing("b.md", "# Title\n\nnew")
	if err != nil {
		t.Fatalf("FileNeedsIndexing failed: %v", err)
	}
	if !needs {
		t.Fatalf("expected reindex needed when content hash changes")
	}
}

func TestFileNeedsIndexing_WhenEmbeddingsComplete(t *testing.T) {
	repo, cleanup := setupRepositoryTestDB(t)
	defer cleanup()

	content := "# Title\n\ncontent"
	chunks := []ChunkInput{
		{Content: "chunk-1", Heading: "Title", Embedding: []float32{0.1, 0.2}, EmbeddingModel: "m1"},
		{Content: "chunk-2", Heading: "Title", Embedding: []float32{0.3, 0.4}, EmbeddingModel: "m1"},
	}
	if err := repo.IndexFileWithChunks("c.md", content, 1, int64(len(content)), chunks); err != nil {
		t.Fatalf("index with chunks failed: %v", err)
	}

	needs, err := repo.FileNeedsIndexing("c.md", content)
	if err != nil {
		t.Fatalf("FileNeedsIndexing failed: %v", err)
	}
	if needs {
		t.Fatalf("expected no reindex when content unchanged and embeddings complete")
	}
}
