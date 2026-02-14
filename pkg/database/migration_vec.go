package database

import (
	"context"
	"fmt"
	"notebit/pkg/config"
	"notebit/pkg/logger"

	"gorm.io/gorm"
)

// MigrateToVec migrates existing embeddings from embedding_blob to vec_chunks virtual table
// This migration is idempotent and can be interrupted/resumed via the vec_indexed flag
func (m *Manager) MigrateToVec(ctx context.Context) error {
	timer := logger.StartTimer()
	db := m.GetDB()
	if db == nil {
		return &DatabaseError{Op: "migrate_to_vec", Err: fmt.Errorf("database not initialized")}
	}

	// Check if vec_chunks table exists
	var tableExists bool
	if err := db.Raw("SELECT COUNT(*) > 0 FROM sqlite_master WHERE type='table' AND name='vec_chunks'").Scan(&tableExists).Error; err != nil {
		logger.ErrorWithFields(ctx, map[string]interface{}{
			"error": err.Error(),
		}, "Failed to check vec_chunks table existence")
		return &DatabaseError{Op: "check_vec_table", Err: err}
	}

	if !tableExists {
		logger.Warn("vec_chunks table does not exist, skipping migration")
		return nil
	}

	// Count chunks needing migration
	var totalCount int64
	if err := db.Model(&Chunk{}).
		Where("embedding_blob IS NOT NULL AND vec_indexed = ?", false).
		Count(&totalCount).Error; err != nil {
		return &DatabaseError{Op: "count_chunks", Err: err}
	}

	if totalCount == 0 {
		logger.Info("No chunks need vec migration")
		return nil
	}

	logger.InfoWithFields(ctx, map[string]interface{}{
		"total_chunks": totalCount,
	}, "Starting vec migration")

	cfg := config.Get()
	batchSize := cfg.Indexing.MigrationBatchSize
	if batchSize <= 0 {
		batchSize = 500
	}

	var processed int64
	offset := 0

	for {
		var chunks []Chunk
		if err := db.Where("embedding_blob IS NOT NULL AND vec_indexed = ?", false).
			Limit(batchSize).
			Offset(offset).
			Find(&chunks).Error; err != nil {
			return &DatabaseError{Op: "fetch_chunks", Err: err}
		}

		if len(chunks) == 0 {
			break
		}

		// Process batch in transaction
		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, chunk := range chunks {
				if len(chunk.EmbeddingBlob) == 0 {
					continue
				}

				// Decode blob to float32 slice
				embedding := bytesToFloats(chunk.EmbeddingBlob)
				if len(embedding) == 0 {
					logger.WarnWithFields(ctx, map[string]interface{}{
						"chunk_id": chunk.ID,
					}, "Empty embedding after decoding, skipping")
					continue
				}

				// Insert into vec_chunks
				if err := insertVecChunk(tx, chunk.ID, embedding); err != nil {
					logger.WarnWithFields(ctx, map[string]interface{}{
						"chunk_id": chunk.ID,
						"error":    err.Error(),
					}, "Failed to insert into vec_chunks, skipping")
					continue
				}

				// Mark as indexed
				if err := tx.Model(&Chunk{}).Where("id = ?", chunk.ID).Update("vec_indexed", true).Error; err != nil {
					return fmt.Errorf("update vec_indexed flag: %w", err)
				}

				processed++
			}
			return nil
		}); err != nil {
			logger.ErrorWithFields(ctx, map[string]interface{}{
				"offset": offset,
				"error":  err.Error(),
			}, "Batch migration failed")
			return &DatabaseError{Op: "migrate_batch", Err: err}
		}

		// Log progress
		if processed%1000 == 0 || processed == totalCount {
			logger.InfoWithFields(ctx, map[string]interface{}{
				"processed": processed,
				"total":     totalCount,
				"progress":  fmt.Sprintf("%.1f%%", float64(processed)/float64(totalCount)*100),
			}, "Vec migration progress")
		}

		// Continue to next batch (don't increment offset since processed rows are now vec_indexed=true)
	}

	logger.InfoWithDuration(ctx, timer(), "Vec migration completed: %d chunks migrated", processed)

	return nil
}

// CleanupLegacyEmbeddings removes the deprecated embedding_blob column after successful migration
// This is optional and should only be called after verifying vec_chunks is working correctly
func (m *Manager) CleanupLegacyEmbeddings(ctx context.Context) error {
	db := m.GetDB()
	if db == nil {
		return &DatabaseError{Op: "cleanup_legacy", Err: fmt.Errorf("database not initialized")}
	}

	// Verify all chunks are indexed
	var unindexedCount int64
	if err := db.Model(&Chunk{}).
		Where("embedding_blob IS NOT NULL AND vec_indexed = ?", false).
		Count(&unindexedCount).Error; err != nil {
		return &DatabaseError{Op: "check_unindexed", Err: err}
	}

	if unindexedCount > 0 {
		return fmt.Errorf("cannot cleanup: %d chunks still not vec_indexed", unindexedCount)
	}

	// Clear embedding_blob to save space
	if err := db.Model(&Chunk{}).
		Where("vec_indexed = ?", true).
		Update("embedding_blob", nil).Error; err != nil {
		return &DatabaseError{Op: "clear_blob", Err: err}
	}

	logger.Info("Cleared legacy embedding_blob data")

	return nil
}

// insertVecChunk inserts a chunk embedding into the vec_chunks virtual table
func insertVecChunk(tx *gorm.DB, chunkID uint, embedding []float32) error {
	// Encode embedding as binary blob for sqlite-vec
	blob := floatsToBytes(embedding)

	// sqlite-vec INSERT syntax: INSERT INTO vec_chunks(chunk_id, embedding) VALUES (?, ?)
	// The embedding parameter should be the raw binary blob
	sql := "INSERT OR REPLACE INTO vec_chunks(chunk_id, embedding) VALUES (?, ?)"
	if err := tx.Exec(sql, chunkID, blob).Error; err != nil {
		return fmt.Errorf("insert vec chunk: %w", err)
	}

	return nil
}

// deleteVecChunk removes a chunk from the vec_chunks virtual table
func deleteVecChunk(tx *gorm.DB, chunkID uint) error {
	sql := "DELETE FROM vec_chunks WHERE chunk_id = ?"
	if err := tx.Exec(sql, chunkID).Error; err != nil {
		return fmt.Errorf("delete vec chunk: %w", err)
	}
	return nil
}
