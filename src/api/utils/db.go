package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/the-digital-watchdog-initiative/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance
var DB *gorm.DB

// InitDB initializes the database connection
func InitDB() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// MySQL connection string format: user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Set up GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second, // Log SQL queries that take longer than 1 second
			LogLevel:                  logger.Info, // Log all SQL queries in development
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Enable color
		},
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		Logger.Fatalf("Failed to connect to database: %v", err)
	}

	Logger.Info("Database connection established")

	// Auto-migrate the models
	migrateDB()
}

// migrateDB performs database auto-migrations
func migrateDB() error {
	// Update existing transactions to set null for user references that don't exist
	DB.Exec(`UPDATE transactions SET approved_by_id = NULL WHERE approved_by_id NOT IN (SELECT id FROM users)`)
	DB.Exec(`UPDATE transactions SET rejected_by_id = NULL WHERE rejected_by_id NOT IN (SELECT id FROM users)`)
	DB.Exec(`UPDATE transactions SET reviewed_by_id = NULL WHERE reviewed_by_id NOT IN (SELECT id FROM users)`)
	DB.Exec(`UPDATE transactions SET created_by_id = NULL WHERE created_by_id NOT IN (SELECT id FROM users)`)

	// Auto-migrate the schema
	err := DB.AutoMigrate(
		&models.User{},
		&models.UserSession{},
		&models.Entity{},
		&models.Fund{},
		&models.Transaction{},
		&models.BudgetLineItem{},
		&models.AuditLog{},
		&models.SmartContractRecord{},
	)

	if err != nil {
		return fmt.Errorf("Database migration failed: %v", err)
	}

	return nil
}

// Paginate returns a GORM scope for pagination
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}

		if pageSize <= 0 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
