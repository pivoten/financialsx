package auth

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
	CreatedAt   time.Time `json:"created_at"`
	LastLogin   *time.Time `json:"last_login,omitempty"`
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

	// Insert user (no company_name field)
	query := `
		INSERT INTO users (username, password_hash, email)
		VALUES (?, ?, ?)
	`
	
	result, err := a.db.GetConn().Exec(query, username, string(hashedPassword), email)
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

	return &User{
		ID:          int(id),
		Username:    username,
		Email:       email,
		CompanyName: a.companyName, // Derived from which database this Auth instance is connected to
		CreatedAt:   time.Now(),
	}, nil
}

func (a *Auth) Login(username, password string) (*User, *Session, error) {
	// Get user (no company_name in database)
	var user User
	var passwordHash string
	
	query := `
		SELECT id, username, password_hash, email, created_at, last_login
		FROM users WHERE username = ?
	`
	
	err := a.db.GetConn().QueryRow(query, username).Scan(
		&user.ID, &user.Username, &passwordHash, &user.Email, 
		&user.CreatedAt, &user.LastLogin,
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
		SELECT u.id, u.username, u.email, u.created_at, u.last_login, s.expires_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token = ?
	`
	
	err := a.db.GetConn().QueryRow(query, token).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.CreatedAt, &user.LastLogin, &expiresAt,
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