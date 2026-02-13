package database

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"notebit/pkg/logger"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Manager handles database operations
type Manager struct {
	db       *gorm.DB
	dbPath   string
	basePath string
	repo     *Repository
	mu       sync.RWMutex
	initOnce sync.Once
	initErr  error
}

var (
	instance *Manager
	once     sync.Once
)

// GetInstance returns the singleton database manager
func GetInstance() *Manager {
	once.Do(func() {
		instance = &Manager{}
	})
	return instance
}

// Init initializes the database connection
func (m *Manager) Init(basePath string) error {
	timer := logger.StartTimer()
	logger.InfoWithFields(context.TODO(), map[string]interface{}{"base_path": basePath}, "Initializing database")

	var err error
	m.initOnce.Do(func() {
		m.mu.Lock()
		m.basePath = basePath
		m.mu.Unlock()

		// Create data directory if not exists
		dataDir := filepath.Join(basePath, "data")
		if err = os.MkdirAll(dataDir, 0755); err != nil {
			logger.ErrorWithFields(context.TODO(), map[string]interface{}{
				"data_dir": dataDir,
				"error":    err.Error(),
			}, "Failed to create data directory")
			m.initErr = &DatabaseError{Op: "create_data_dir", Err: err}
			return
		}

		// Set database path
		dbPath := filepath.Join(dataDir, "notebit.sqlite")
		m.dbPath = dbPath

		// Open SQLite connection using pure Go driver (no CGO)
		m.db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		})
		if err != nil {
			logger.ErrorWithFields(context.TODO(), map[string]interface{}{
				"db_path": dbPath,
				"error":   err.Error(),
			}, "Failed to open database")
			m.initErr = &DatabaseError{Op: "open_database", Err: err}
			return
		}

		// Run migrations
		if err = m.AutoMigrate(); err != nil {
			logger.ErrorWithFields(context.TODO(), map[string]interface{}{
				"error": err.Error(),
			}, "Failed to run database migrations")
			m.initErr = &DatabaseError{Op: "migrate", Err: err}
			return
		}

		logger.InfoWithDuration(context.TODO(), timer(), "Database initialized successfully: %s", dbPath)
	})

	return m.initErr
}

// Close closes the database connection
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db != nil {
		sqlDB, err := m.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB returns the GORM DB instance (internal use)
func (m *Manager) GetDB() *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db
}

// GetDBPath returns the database file path
func (m *Manager) GetDBPath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dbPath
}

// GetBasePath returns the base notes directory path
func (m *Manager) GetBasePath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.basePath
}

// IsInitialized returns true if the database has been initialized
func (m *Manager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db != nil
}

// Reset resets the singleton (for testing purposes)
func Reset() {
	once = sync.Once{}
	instance = nil
}
