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
	-- Roles table with predefined system roles
	CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		display_name TEXT NOT NULL,
		description TEXT,
		is_system_role BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Permissions table
	CREATE TABLE IF NOT EXISTS permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		display_name TEXT NOT NULL,
		description TEXT,
		resource TEXT NOT NULL,  -- e.g., 'users', 'dbf_files', 'settings'
		action TEXT NOT NULL,    -- e.g., 'create', 'read', 'update', 'delete'
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Role permissions mapping
	CREATE TABLE IF NOT EXISTS role_permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		role_id INTEGER NOT NULL,
		permission_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (role_id) REFERENCES roles(id),
		FOREIGN KEY (permission_id) REFERENCES permissions(id),
		UNIQUE(role_id, permission_id)
	);

	-- Updated users table with role
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		email TEXT,
		role_id INTEGER NOT NULL DEFAULT 3, -- Default to Read-Only
		is_active BOOLEAN DEFAULT TRUE,
		is_root BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_login DATETIME,
		created_by INTEGER,
		FOREIGN KEY (role_id) REFERENCES roles(id),
		FOREIGN KEY (created_by) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
	CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
	CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);
	`

	if _, err := db.conn.Exec(schema); err != nil {
		return err
	}

	// Insert default roles and permissions
	return db.insertDefaultRolesAndPermissions()
}

func (db *DB) insertDefaultRolesAndPermissions() error {
	// Insert default roles
	roles := []struct {
		name        string
		displayName string
		description string
		isSystem    bool
	}{
		{"root", "Root", "System administrator with full access", true},
		{"admin", "Administrator", "Full access to manage users and data", true},
		{"readonly", "Read-Only", "Can view data but cannot modify anything", true},
	}

	for _, role := range roles {
		_, err := db.conn.Exec(`
			INSERT OR IGNORE INTO roles (name, display_name, description, is_system_role) 
			VALUES (?, ?, ?, ?)
		`, role.name, role.displayName, role.description, role.isSystem)
		if err != nil {
			return fmt.Errorf("failed to insert role %s: %w", role.name, err)
		}
	}

	// Insert default permissions
	permissions := []struct {
		name        string
		displayName string
		description string
		resource    string
		action      string
	}{
		// User management permissions
		{"users.create", "Create Users", "Create new user accounts", "users", "create"},
		{"users.read", "View Users", "View user accounts and profiles", "users", "read"},
		{"users.update", "Update Users", "Modify user accounts and profiles", "users", "update"},
		{"users.delete", "Delete Users", "Delete user accounts", "users", "delete"},
		{"users.manage_roles", "Manage User Roles", "Assign and modify user roles", "users", "manage_roles"},
		
		// DBF file permissions
		{"dbf.read", "View DBF Files", "View and browse DBF files", "dbf_files", "read"},
		{"dbf.write", "Edit DBF Files", "Edit and modify DBF file data", "dbf_files", "write"},
		{"dbf.export", "Export DBF Data", "Export DBF data to various formats", "dbf_files", "export"},
		{"dbf.import", "Import DBF Data", "Import data into DBF files", "dbf_files", "import"},
		
		// System settings permissions
		{"settings.read", "View Settings", "View system and application settings", "settings", "read"},
		{"settings.write", "Modify Settings", "Modify system and application settings", "settings", "write"},
		
		// Database maintenance permissions
		{"database.read", "View Database", "View database status and information", "database", "read"},
		{"database.maintain", "Database Maintenance", "Perform database maintenance tasks", "database", "maintain"},
		
		// Reporting permissions
		{"reports.read", "View Reports", "View and generate reports", "reports", "read"},
		{"reports.create", "Create Reports", "Create and customize reports", "reports", "create"},
	}

	for _, perm := range permissions {
		_, err := db.conn.Exec(`
			INSERT OR IGNORE INTO permissions (name, display_name, description, resource, action) 
			VALUES (?, ?, ?, ?, ?)
		`, perm.name, perm.displayName, perm.description, perm.resource, perm.action)
		if err != nil {
			return fmt.Errorf("failed to insert permission %s: %w", perm.name, err)
		}
	}

	// Assign permissions to roles
	return db.assignRolePermissions()
}

func (db *DB) assignRolePermissions() error {
	// Root role gets all permissions
	_, err := db.conn.Exec(`
		INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id 
		FROM roles r, permissions p 
		WHERE r.name = 'root'
	`)
	if err != nil {
		return fmt.Errorf("failed to assign permissions to root role: %w", err)
	}

	// Admin role gets most permissions except some root-only ones
	adminPermissions := []string{
		"users.create", "users.read", "users.update", "users.manage_roles",
		"dbf.read", "dbf.write", "dbf.export", "dbf.import",
		"settings.read", "settings.write",
		"database.read", "database.maintain",
		"reports.read", "reports.create",
	}
	
	for _, permName := range adminPermissions {
		_, err := db.conn.Exec(`
			INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
			SELECT r.id, p.id 
			FROM roles r, permissions p 
			WHERE r.name = 'admin' AND p.name = ?
		`, permName)
		if err != nil {
			return fmt.Errorf("failed to assign permission %s to admin role: %w", permName, err)
		}
	}

	// Read-only role gets only read permissions
	readOnlyPermissions := []string{
		"users.read", "dbf.read", "dbf.export", "settings.read", "database.read", "reports.read",
	}
	
	for _, permName := range readOnlyPermissions {
		_, err := db.conn.Exec(`
			INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
			SELECT r.id, p.id 
			FROM roles r, permissions p 
			WHERE r.name = 'readonly' AND p.name = ?
		`, permName)
		if err != nil {
			return fmt.Errorf("failed to assign permission %s to readonly role: %w", permName, err)
		}
	}

	return nil
}

func (db *DB) GetConn() *sql.DB {
	return db.conn
}