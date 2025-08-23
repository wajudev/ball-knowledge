package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"game-knowledge/backend/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDatabase initializes the database connection
func ConnectDatabase() error {
	// Configure GORM logger
	logLevel := logger.Silent
	if os.Getenv("GIN_MODE") == "debug" {
		logLevel = logger.Warn // Changed from Info to Warn to reduce noise
	}

	// Database configuration
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Database path
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "ball_knowledge.db"
	}

	// SQLite connection string with optimizations
	dsn := fmt.Sprintf("%s?cache=shared&mode=rwc&_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=on", dbPath)

	database, err := gorm.Open(sqlite.Open(dsn), config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Configure connection pool
	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// SQLite optimized settings
	sqlDB.SetMaxIdleConns(1)   // SQLite doesn't benefit from multiple connections
	sqlDB.SetMaxOpenConns(1)   // SQLite is file-based, single connection is optimal
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto-migrate database schema
	if err := autoMigrate(database); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	DB = database
	log.Println("✅ Database connected and migrated successfully")
	return nil
}

// autoMigrate runs database migrations
func autoMigrate(db *gorm.DB) error {
	// Drop and recreate tables if there are schema issues (development only)
	if os.Getenv("GIN_MODE") == "debug" {
		// Only do this in debug mode and if the database is empty
		var tableCount int64
		db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tableCount)

		if tableCount == 0 {
			log.Println("🔄 Creating fresh database schema...")
		}
	}

	return db.AutoMigrate(
		&models.User{},
		&models.Match{},
		&models.Prediction{},
	)
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// HealthCheck performs a database health check
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	return nil
}