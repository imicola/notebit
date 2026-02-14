package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"notebit/pkg/logger"

	sqlite3 "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	defaultSQLiteDriver = "sqlite3"
	vecSQLiteDriver     = "sqlite3_vec"
)

// Manager handles database operations
type Manager struct {
	db       *gorm.DB
	dbPath   string
	basePath string
	repo     *Repository
	mu       sync.RWMutex
	initErr  error
}

var (
	instance *Manager
	once     sync.Once

	registerVecDriverOnce sync.Once
	vecDriverAvailable    bool
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

	m.mu.Lock()
	sameBase := m.basePath == basePath && basePath != ""
	if sameBase && m.db != nil && m.initErr == nil {
		m.mu.Unlock()
		return nil
	}
	if m.db != nil {
		if sqlDB, err := m.db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	m.db = nil
	m.dbPath = ""
	m.basePath = basePath
	m.repo = nil
	m.initErr = nil
	m.mu.Unlock()

	dataDir := filepath.Join(basePath, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.ErrorWithFields(context.TODO(), map[string]interface{}{
			"data_dir": dataDir,
			"error":    err.Error(),
		}, "Failed to create data directory")
		m.mu.Lock()
		m.initErr = &DatabaseError{Op: "create_data_dir", Err: err}
		m.mu.Unlock()
		return m.initErr
	}

	dbPath := filepath.Join(dataDir, "notebit.sqlite")
	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_foreign_keys=1", dbPath)
	driverName := defaultSQLiteDriver
	if registerSQLiteVecDriver() {
		driverName = vecSQLiteDriver
	}

	dialector := sqlite.New(sqlite.Config{
		DriverName: driverName,
		DSN:        dsn,
	})

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil && driverName == vecSQLiteDriver {
		logger.WarnWithFields(context.TODO(), map[string]interface{}{
			"error": err.Error(),
		}, "sqlite-vec driver open failed, fallback to default sqlite3 driver")

		dialector = sqlite.New(sqlite.Config{
			DriverName: defaultSQLiteDriver,
			DSN:        dsn,
		})
		db, err = gorm.Open(dialector, &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		})
	}
	if err != nil {
		logger.ErrorWithFields(context.TODO(), map[string]interface{}{
			"db_path": dbPath,
			"error":   err.Error(),
		}, "Failed to open database")
		m.mu.Lock()
		m.initErr = &DatabaseError{Op: "open_database", Err: err}
		m.mu.Unlock()
		return m.initErr
	}

	if err := applyPragmas(db); err != nil {
		logger.WarnWithFields(context.TODO(), map[string]interface{}{
			"error": err.Error(),
		}, "Failed to apply one or more SQLite PRAGMA settings")
	}

	m.mu.Lock()
	m.db = db
	m.dbPath = dbPath
	m.basePath = basePath
	m.repo = nil
	m.initErr = nil
	m.mu.Unlock()

	if err := m.AutoMigrate(); err != nil {
		logger.ErrorWithFields(context.TODO(), map[string]interface{}{
			"error": err.Error(),
		}, "Failed to run database migrations")
		if sqlDB, closeErr := db.DB(); closeErr == nil {
			_ = sqlDB.Close()
		}
		m.mu.Lock()
		m.db = nil
		m.initErr = &DatabaseError{Op: "migrate", Err: err}
		m.mu.Unlock()
		return m.initErr
	}

	go func() {
		if err := m.MigrateToVec(context.Background()); err != nil {
			logger.WarnWithFields(context.Background(), map[string]interface{}{
				"error": err.Error(),
			}, "Vec migration skipped or failed")
		}
	}()

	logger.InfoWithDuration(context.TODO(), timer(), "Database initialized successfully: %s", dbPath)
	return nil
}

func registerSQLiteVecDriver() bool {
	registerVecDriverOnce.Do(func() {
		for _, name := range sql.Drivers() {
			if name == vecSQLiteDriver {
				vecDriverAvailable = true
				return
			}
		}

		defer func() {
			if recover() != nil {
				vecDriverAvailable = false
			}
		}()

		sql.Register(vecSQLiteDriver, &sqlite3.SQLiteDriver{
			Extensions: []string{"vec0"},
		})

		for _, name := range sql.Drivers() {
			if name == vecSQLiteDriver {
				vecDriverAvailable = true
				return
			}
		}
	})

	return vecDriverAvailable
}

func applyPragmas(db *gorm.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA cache_size=-64000",
		"PRAGMA mmap_size=268435456",
		"PRAGMA foreign_keys=ON",
	}

	for _, pragma := range pragmas {
		if err := db.Exec(pragma).Error; err != nil {
			return err
		}
	}

	return nil
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
