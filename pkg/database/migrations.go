package database

import (
	"context"
	"errors"
	"fmt"
	"notebit/pkg/config"
	"notebit/pkg/logger"

	"gorm.io/gorm"
)

const currentSchemaVersion = 1

type schemaVersion struct {
	Version int `gorm:"primaryKey"`
}

// AutoMigrate runs auto-migration for all models
func (m *Manager) AutoMigrate() error {
	db := m.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := db.AutoMigrate(
		&File{},
		&Chunk{},
		&Tag{},
		&FileTag{},
		&schemaVersion{},
	); err != nil {
		return err
	}

	if err := m.migrateSchemaVersion(db); err != nil {
		return err
	}

	return m.EnsureIndexes()
}

// EnsureIndexes creates additional indexes for performance
// Note: Most indexes are defined via gorm tags in models
func (m *Manager) EnsureIndexes() error {
	db := m.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_chunks_file_id ON chunks(file_id)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_chunks_embedding_model ON chunks(embedding_model)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_chunks_vec_indexed ON chunks(vec_indexed)").Error; err != nil {
		return err
	}

	return nil
}

func (m *Manager) migrateSchemaVersion(db *gorm.DB) error {
	var latest schemaVersion
	err := db.Order("version DESC").First(&latest).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	current := latest.Version
	for v := current + 1; v <= currentSchemaVersion; v++ {
		switch v {
		case 1:
			if err := applySchemaMigrationV1(db); err != nil {
				return err
			}
		}

		if err := db.Create(&schemaVersion{Version: v}).Error; err != nil {
			return err
		}
	}

	return nil
}

func applySchemaMigrationV1(db *gorm.DB) error {
	dimension := config.Get().AI.VectorDimension
	if dimension <= 0 {
		dimension = 1536
	}

	vecDDL := fmt.Sprintf(
		"CREATE VIRTUAL TABLE IF NOT EXISTS vec_chunks USING vec0(chunk_id INTEGER PRIMARY KEY, embedding float[%d])",
		dimension,
	)

	if err := db.Exec(vecDDL).Error; err != nil {
		logger.WarnWithFields(context.TODO(), map[string]interface{}{
			"error":     err.Error(),
			"dimension": dimension,
		}, "sqlite-vec unavailable, skip vec_chunks creation and keep brute-force fallback")
		return nil
	}

	return nil
}
