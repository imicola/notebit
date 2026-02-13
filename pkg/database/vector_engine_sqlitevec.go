package database

import "fmt"

// SQLiteVecEngine is a placeholder adapter for sqlite-vec backend.
// It allows gray rollout with safe fallback while native extension integration is pending.
type SQLiteVecEngine struct{}

func NewSQLiteVecEngine() *SQLiteVecEngine {
	return &SQLiteVecEngine{}
}

func (e *SQLiteVecEngine) Name() string {
	return VectorEngineSQLiteVec
}

func (e *SQLiteVecEngine) Search(repo *Repository, queryVector []float32, limit int) ([]SimilarChunk, error) {
	return nil, fmt.Errorf("sqlite-vec engine not available in current build")
}
