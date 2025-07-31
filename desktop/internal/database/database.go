package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// getDatafilesPath returns the path to the datafiles directory
func getDatafilesPath() (string, error) {
	// Possible locations for datafiles directory
	possiblePaths := []string{
		"./datafiles",        // Current directory (production)
		"../datafiles",       // One level up (dev from desktop folder)
		"../../datafiles",    // Two levels up (if nested deeper)
	}
	
	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path, nil
		}
	}
	
	// If not found, create in current directory
	datafilesPath := "./datafiles"
	if err := os.MkdirAll(datafilesPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create datafiles directory: %w", err)
	}
	
	return datafilesPath, nil
}

type DB struct {
	conn *sql.DB
}

func New(companyName string) (*DB, error) {
	datafilesPath, err := getDatafilesPath()
	if err != nil {
		return nil, err
	}
	
	dbPath := filepath.Join(datafilesPath, companyName, "sql", "financialsx.db")
	
	// Create directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	
	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		email TEXT,
		company_name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_login DATETIME
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	`

	_, err := db.conn.Exec(schema)
	return err
}

func (db *DB) GetConn() *sql.DB {
	return db.conn
}