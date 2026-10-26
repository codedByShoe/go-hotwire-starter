// Package db opens the app's SQLite database.
package db

//go:generate go tool gorm gen -i ../models -o ./g
//go:generate go tool gorm gen -i ./queries -o ./g

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codedbyshoe/go-hotwire-starter/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open connects to the SQLite database at path, creating the file and its
// folder if they don't exist yet.
func Open(path string) (*gorm.DB, error) {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create database folder: %w", err)
	}

	// Settings for every connection:
	//   _journal_mode=WAL   readers don't block the writer (and vice versa)
	//   _busy_timeout=5000  wait up to 5s for a lock instead of failing right away
	//   _foreign_keys=on    enforce foreign key constraints
	//   _synchronous=NORMAL safe with WAL and much faster than FULL
	//   _txlock=immediate   transactions take the write lock when they start,
	//                       so two transactions can't deadlock upgrading to write
	dsn := "file:" + path +
		"?_journal_mode=WAL" +
		"&_busy_timeout=5000" +
		"&_foreign_keys=on" +
		"&_synchronous=NORMAL" +
		"&_txlock=immediate"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	return db, nil
}

// Migrate creates or updates tables to match the models. AutoMigrate only
// adds tables, columns and indexes. It never drops or renames anything.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.Note{})
}

// Close closes the database's connections.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
