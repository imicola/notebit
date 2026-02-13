package database

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupVectorEngineTestDB(t *testing.T) (*Repository, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "notebit-vector-engine-*")
	if err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(tmpDir, "vector.sqlite")
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

	repo := &Repository{db: db, vectorEngine: NewBruteForceVectorEngine()}
	cleanup := func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		_ = os.RemoveAll(tmpDir)
	}
	return repo, cleanup
}

func TestSearchSimilar_FallbackFromSQLiteVecToBruteForce(t *testing.T) {
	repo, cleanup := setupVectorEngineTestDB(t)
	defer cleanup()

	file := File{Path: "note.md", Title: "note"}
	if err := repo.db.Create(&file).Error; err != nil {
		t.Fatalf("create file failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		vec := []float32{float32(i) + 1, 0.5}
		chunk := Chunk{
			FileID:        file.ID,
			Content:       fmt.Sprintf("chunk-%d", i),
			Heading:       "h",
			Embedding:     vec,
			EmbeddingBlob: floatsToBytes(vec),
		}
		if err := repo.db.Create(&chunk).Error; err != nil {
			t.Fatalf("create chunk failed: %v", err)
		}
	}

	repo.SetVectorEngine(VectorEngineSQLiteVec)

	results, err := repo.SearchSimilar([]float32{1, 0.5}, 2)
	if err != nil {
		t.Fatalf("search should fallback to brute-force, got err: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected fallback search results")
	}
	if repo.GetVectorEngine() != VectorEngineBruteForce {
		t.Fatalf("expected engine switched to brute-force after fallback")
	}
}

func TestSetVectorEngine_UnknownFallsBack(t *testing.T) {
	repo, cleanup := setupVectorEngineTestDB(t)
	defer cleanup()

	effective := repo.SetVectorEngine("unknown-engine")
	if effective != VectorEngineBruteForce {
		t.Fatalf("expected fallback to %s, got %s", VectorEngineBruteForce, effective)
	}
}
