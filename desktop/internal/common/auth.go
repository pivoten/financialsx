package common

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/database"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserExists         = errors.New("username already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidResetToken  = errors.New("invalid or expired reset token")
	ErrPasswordTooWeak    = errors.New("password does not meet requirements")
)

type Auth struct {
	db          *database.DB
	companyName string
}

type User struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email,omitempty"`
	CompanyName string    `json:"company_name"` // Derived from which database they're in
	RoleID      int       `json:"role_id"`
	RoleName    string    `json:"role_name"`
	IsActive    bool      `json:"is_active"`
	IsRoot      bool      `json:"is_root"`
	CreatedAt   time.Time `json:"created_at"`
	LastLogin   *time.Time `json:"last_login,omitempty"`
	Permissions []string  `json:"permissions,omitempty"`
}

type Session struct {
	Token       string    `json:"token"`
	UserID      int       `json:"user_id"`
	CompanyName string    `json:"company_name"` // Derived from which database the session is in
	ExpiresAt   time.Time `json:"expires_at"`
}

func New(db *database.DB, companyName string) *Auth {
	return &Auth{
		db:          db,
		companyName: companyName,
	}
}

func (a *Auth) Register(username, password, email string) (*User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Check if this is the first user (will be root)
	var userCount int
	err = a.db.GetConn().QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return nil, fmt.Errorf("failed to check user count: %w", err)
	}

	isRoot := userCount == 0
	roleID := 3 // Default to Read-Only
	if isRoot {
		roleID = 1 // Root role
	}

	// Insert user with role
	query := `
		INSERT INTO users (username, password_hash, email, role_id, is_root, is_active)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	result, err := a.db.GetConn().Exec(query, username, string(hashedPassword), email, roleID, isRoot, true)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: users.username" {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	// Get role name
	var roleName string
	err = a.db.GetConn().QueryRow("SELECT name FROM roles WHERE id = ?", roleID).Scan(&roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get role name: %w", err)
	}

	return &User{
		ID:          int(id),
		Username:    username,
		Email:       email,
		CompanyName: a.companyName, // Derived from which database this Auth instance is connected to
		RoleID:      roleID,
		RoleName:    roleName,
		IsActive:    true,
		IsRoot:      isRoot,
		CreatedAt:   time.Now(),
	}, nil
}

func (a *Auth) Login(username, password string) (*User, *Session, error) {
	// Get user (no company_name in database)
	var user User
	var passwordHash string
	
	query := `
		SELECT u.id, u.username, u.password_hash, u.email, u.role_id, u.is_active, u.is_root, u.created_at, u.last_login, r.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.username = ? AND u.is_active = TRUE
	`
	
	err := a.db.GetConn().QueryRow(query, username).Scan(
		&user.ID, &user.Username, &passwordHash, &user.Email, &user.RoleID, &user.IsActive, &user.IsRoot,
		&user.CreatedAt, &user.LastLogin, &user.RoleName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Set company name from Auth instance (derived from database location)
	user.CompanyName = a.companyName

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Update last login
	now := time.Now()
	_, err = a.db.GetConn().Exec("UPDATE users SET last_login = ? WHERE id = ?", now, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update last login: %w", err)
	}
	user.LastLogin = &now

	// Load user permissions
	permissions, err := a.getUserPermissions(user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load user permissions: %w", err)
	}
	user.Permissions = permissions

	// Create session
	session, err := a.createSession(user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &user, session, nil
}

func (a *Auth) ValidateSession(token string) (*User, error) {
	// Get session and user from this database (company isolation is implicit)
	var user User
	var expiresAt time.Time
	
	query := `
		SELECT u.id, u.username, u.email, u.role_id, u.is_active, u.is_root, u.created_at, u.last_login, r.name, s.expires_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		JOIN roles r ON u.role_id = r.id
		WHERE s.token = ? AND u.is_active = TRUE
	`
	
	err := a.db.GetConn().QueryRow(query, token).Scan(
		&user.ID, &user.Username, &user.Email, &user.RoleID, &user.IsActive, &user.IsRoot,
		&user.CreatedAt, &user.LastLogin, &user.RoleName, &expiresAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Set company name from Auth instance (derived from database location)
	user.CompanyName = a.companyName

	// Check if expired
	if time.Now().After(expiresAt) {
		// Clean up expired session
		a.db.GetConn().Exec("DELETE FROM sessions WHERE token = ?", token)
		return nil, ErrInvalidToken
	}

	// Load user permissions
	permissions, err := a.getUserPermissions(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user permissions: %w", err)
	}
	user.Permissions = permissions

	return &user, nil
}

func (a *Auth) Logout(token string) error {
	_, err := a.db.GetConn().Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

func (a *Auth) createSession(userID int) (*Session, error) {
	// Generate token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	
	hash := sha256.Sum256(tokenBytes)
	token := hex.EncodeToString(hash[:])
	
	// Set expiration (24 hours)
	expiresAt := time.Now().Add(24 * time.Hour)
	
	// Insert session (no company_name field)
	_, err := a.db.GetConn().Exec(
		"INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &Session{
		Token:       token,
		UserID:      userID,
		CompanyName: a.companyName, // Derived from Auth instance
		ExpiresAt:   expiresAt,
	}, nil
}

// getUserPermissions retrieves all permissions for a user based on their role
func (a *Auth) getUserPermissions(userID int) ([]string, error) {
	query := `
		SELECT DISTINCT p.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE u.id = ? AND u.is_active = TRUE
		ORDER BY p.name
	`

	rows, err := a.db.GetConn().Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (u *User) HasPermission(permission string) bool {
	for _, p := range u.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a user has any of the specified permissions
func (u *User) HasAnyPermission(permissions ...string) bool {
	for _, permission := range permissions {
		if u.HasPermission(permission) {
			return true
		}
	}
	return false
}

// IsAdmin checks if user has admin or root privileges
func (u *User) IsAdmin() bool {
	return u.IsRoot || u.RoleName == "admin"
}

// GetAllUsers retrieves all users with their roles (admin/root only)
func (a *Auth) GetAllUsers() ([]User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.role_id, u.is_active, u.is_root, u.created_at, u.last_login, r.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		ORDER BY u.created_at DESC
	`

	rows, err := a.db.GetConn().Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.RoleID, &user.IsActive, &user.IsRoot,
			&user.CreatedAt, &user.LastLogin, &user.RoleName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		user.CompanyName = a.companyName
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// UpdateUserRole updates a user's role (admin/root only)
func (a *Auth) UpdateUserRole(userID, newRoleID int) error {
	_, err := a.db.GetConn().Exec(
		"UPDATE users SET role_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		userID, newRoleID,
	)
	return err
}

// UpdateUserStatus activates or deactivates a user (admin/root only)
func (a *Auth) UpdateUserStatus(userID int, isActive bool) error {
	_, err := a.db.GetConn().Exec(
		"UPDATE users SET is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		isActive, userID,
	)
	return err
}

// GetAllRoles retrieves all available roles
func (a *Auth) GetAllRoles() ([]Role, error) {
	query := `
		SELECT id, name, display_name, description, is_system_role
		FROM roles
		ORDER BY id
	`

	rows, err := a.db.GetConn().Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		err := rows.Scan(&role.ID, &role.Name, &role.DisplayName, &role.Description, &role.IsSystemRole)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// Role represents a user role
type Role struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Description  string `json:"description"`
	IsSystemRole bool   `json:"is_system_role"`
}

// PasswordResetToken represents a password reset request
type PasswordResetToken struct {
	Token     string    `json:"token"`
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

// ============================================================================
// PASSWORD MANAGEMENT FUNCTIONS
// ============================================================================

// ChangePassword allows a user to change their own password
func (a *Auth) ChangePassword(userID int, oldPassword, newPassword string) error {
	// Validate new password strength
	if err := validatePasswordStrength(newPassword); err != nil {
		return err
	}

	// Get current password hash
	var currentHash string
	err := a.db.GetConn().QueryRow(
		"SELECT password_hash FROM users WHERE id = ? AND is_active = 1",
		userID,
	).Scan(&currentHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	_, err = a.db.GetConn().Exec(
		"UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(newHash), userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// RequestPasswordReset generates a password reset token for a user
func (a *Auth) RequestPasswordReset(email string) (*PasswordResetToken, error) {
	// Find user by email
	var userID int
	var username string
	err := a.db.GetConn().QueryRow(
		"SELECT id, username FROM users WHERE email = ? AND is_active = 1",
		email,
	).Scan(&userID, &username)
	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal if email exists or not for security
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Generate reset token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Hash token for storage
	hasher := sha256.New()
	hasher.Write([]byte(token))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))

	// Store reset token (expires in 1 hour)
	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = a.db.GetConn().Exec(`
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, userID, tokenHash, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to store reset token: %w", err)
	}

	return &PasswordResetToken{
		Token:     token,
		UserID:    userID,
		Email:     email,
		ExpiresAt: expiresAt,
		Used:      false,
	}, nil
}

// ResetPassword resets a user's password using a valid reset token
func (a *Auth) ResetPassword(token, newPassword string) error {
	// Validate new password strength
	if err := validatePasswordStrength(newPassword); err != nil {
		return err
	}

	// Hash the provided token to match stored hash
	hasher := sha256.New()
	hasher.Write([]byte(token))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))

	// Find valid token
	var userID int
	var expiresAt time.Time
	var used bool
	err := a.db.GetConn().QueryRow(`
		SELECT user_id, expires_at, used 
		FROM password_reset_tokens 
		WHERE token_hash = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, tokenHash).Scan(&userID, &expiresAt, &used)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrInvalidResetToken
		}
		return fmt.Errorf("failed to find reset token: %w", err)
	}

	// Check if token is expired or used
	if time.Now().After(expiresAt) || used {
		return ErrInvalidResetToken
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Start transaction
	tx, err := a.db.GetConn().Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Update password
	_, err = tx.Exec(
		"UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(newHash), userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	_, err = tx.Exec(
		"UPDATE password_reset_tokens SET used = 1 WHERE token_hash = ?",
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// AdminResetPassword allows an admin to reset any user's password
func (a *Auth) AdminResetPassword(userID int, newPassword string) error {
	// Validate new password strength
	if err := validatePasswordStrength(newPassword); err != nil {
		return err
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	result, err := a.db.GetConn().Exec(
		"UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(newHash), userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// validatePasswordStrength checks if a password meets minimum requirements
func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", ErrPasswordTooWeak)
	}

	// Check for at least one uppercase letter
	hasUpper := false
	hasLower := false
	hasNumber := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return fmt.Errorf("%w: password must contain uppercase, lowercase, and numbers", ErrPasswordTooWeak)
	}

	return nil
}