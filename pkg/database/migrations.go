package database

// AutoMigrate runs auto-migration for all models
func (m *Manager) AutoMigrate() error {
	db := m.GetDB()
	return db.AutoMigrate(
		&File{},
		&Chunk{},
		&Tag{},
		&FileTag{},
	)
}

// EnsureIndexes creates additional indexes for performance
// Note: Most indexes are defined via gorm tags in models
func (m *Manager) EnsureIndexes() error {
	// Custom indexes can be added here using raw SQL if needed
	// For example, full-text search indexes:
	// db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(content, content=chunks, content_rowid=rowid)")
	return nil
}
