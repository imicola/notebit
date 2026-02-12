package database

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(b *testing.B) (*Repository, func()) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "notebit-bench-*")
	if err != nil {
		b.Fatal(err)
	}

	dbPath := filepath.Join(tmpDir, "bench.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		b.Fatal(err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&File{}, &Chunk{}); err != nil {
		os.RemoveAll(tmpDir)
		b.Fatal(err)
	}

	repo := &Repository{db: db}

	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		os.RemoveAll(tmpDir)
	}

	return repo, cleanup
}

func generateRandomEmbedding(dim int) []float32 {
	vec := make([]float32, dim)
	for i := 0; i < dim; i++ {
		vec[i] = rand.Float32()
	}
	return vec
}

func BenchmarkSearchSimilar_1000Chunks(b *testing.B) {
	repo, cleanup := setupTestDB(b)
	defer cleanup()

	// Insert 1000 chunks
	dim := 1536 // OpenAI embedding dimension
	batchSize := 100
	totalChunks := 1000

	file := File{Path: "bench.md", Title: "Bench"}
	repo.db.Create(&file)

	chunks := make([]Chunk, batchSize)
	for i := 0; i < totalChunks; i += batchSize {
		for j := 0; j < batchSize; j++ {
			chunks[j] = Chunk{
				FileID:        file.ID,
				Content:       fmt.Sprintf("Chunk %d", i+j),
				Embedding:     generateRandomEmbedding(dim),
				EmbeddingBlob: floatsToBytes(generateRandomEmbedding(dim)), // Populate blob
			}
		}
		repo.db.Create(&chunks)
	}

	queryVec := generateRandomEmbedding(dim)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.SearchSimilar(queryVec, 5)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSearchSimilar_100Chunks(b *testing.B) {
	repo, cleanup := setupTestDB(b)
	defer cleanup()

	// Insert 100 chunks
	dim := 1536
	batchSize := 100
	totalChunks := 100

	file := File{Path: "bench.md", Title: "Bench"}
	repo.db.Create(&file)

	chunks := make([]Chunk, batchSize)
	for i := 0; i < totalChunks; i += batchSize {
		for j := 0; j < batchSize; j++ {
			chunks[j] = Chunk{
				FileID:        file.ID,
				Content:       fmt.Sprintf("Chunk %d", i+j),
				Embedding:     generateRandomEmbedding(dim),
				EmbeddingBlob: floatsToBytes(generateRandomEmbedding(dim)), // Populate blob
			}
		}
		repo.db.Create(&chunks)
	}

	queryVec := generateRandomEmbedding(dim)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.SearchSimilar(queryVec, 5)
		if err != nil {
			b.Fatal(err)
		}
	}
}
