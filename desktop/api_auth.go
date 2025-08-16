package main

import (
	"fmt"

	"github.com/pivoten/financialsx/desktop/internal/common"
)

// ============================================================================
// AUTHENTICATION API
// This file contains all authentication-related API methods for the App struct
// ============================================================================

// Login authenticates a user with username and password
func (a *App) Login(username, password, companyName string) (map[string]interface{}, error) {
	fmt.Printf("Login attempt for user: %s, company: %s\n", username, companyName)

	// Initialize database for the specific company if needed
	if err := a.InitializeCompanyDatabase(companyName); err != nil {
		fmt.Printf("Failed to initialize company database: %v\n", err)
		return nil, fmt.Errorf("failed to initialize company database: %v", err)
	}

	// Initialize auth with the company-specific database
	a.auth = common.New(a.db, companyName)

	// Attempt login with company context
	user, token, err := a.auth.Login(username, password)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return nil, err
	}

	// Store the current user and company for the session
	a.currentUser = user
	a.currentCompanyPath = companyName

	// Update cached auth state for performance
	a.updateAuthCache()

	fmt.Printf("Login successful for user: %s (ID: %d, Root: %v)\n", user.Username, user.ID, user.IsRoot)

	// Return the user data and token
	return map[string]interface{}{
		"success": true,
		"token":   token,
		"user": map[string]interface{}{
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"role_id":      user.RoleID,
			"role_name":    user.RoleName,
			"is_active":    user.IsActive,
			"is_root":      user.IsRoot,
			"company_name": companyName,
			"permissions":  user.Permissions,
			"last_login":   user.LastLogin,
		},
	}, nil
}

// Register creates a new user account
func (a *App) Register(username, password, email, companyName string) (map[string]interface{}, error) {
	fmt.Printf("Register attempt for user: %s, company: %s\n", username, companyName)

	// Initialize database for the specific company if needed
	if err := a.InitializeCompanyDatabase(companyName); err != nil {
		return nil, fmt.Errorf("failed to initialize company database: %v", err)
	}

	// Initialize auth with the company-specific database
	a.auth = common.New(a.db, companyName)

	// Use a default role ID (e.g., 2 for "User" role)
	// This should be configurable or determined by business logic
	defaultRoleID := 2

	// Create the user with company context
	user, err := a.auth.CreateUser(username, password, email, defaultRoleID)
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		return nil, err
	}

	fmt.Printf("Registration successful for user: %s (ID: %d)\n", user.Username, user.ID)

	// Return the created user data
	return map[string]interface{}{
		"success": true,
		"user": map[string]interface{}{
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"role_id":      user.RoleID,
			"role_name":    user.RoleName,
			"is_active":    user.IsActive,
			"is_root":      user.IsRoot,
			"company_name": companyName,
			"created_at":   user.CreatedAt,
		},
	}, nil
}

// Logout logs out the current user
func (a *App) Logout(token string) error {
	fmt.Printf("Logout request for user: %v\n", a.currentUser)

	if a.auth != nil {
		// Logout with the provided token
		if err := a.auth.Logout(token); err != nil {
			// Log the error but don't fail the logout
			fmt.Printf("Warning: auth.Logout error: %v\n", err)
		}
	}

	// Clear the current user and auth state
	a.currentUser = nil
	a.currentCompanyPath = ""

	// Clear cached auth state
	a.updateAuthCache()

	fmt.Println("User logged out successfully")
	return nil
}

// ValidateSession validates a session token
func (a *App) ValidateSession(token string, companyName string) (*common.User, error) {
	fmt.Printf("ValidateSession called for company: %s\n", companyName)

	// Initialize database for the specific company if needed
	if err := a.InitializeCompanyDatabase(companyName); err != nil {
		return nil, fmt.Errorf("failed to initialize company database: %v", err)
	}

	// Initialize auth with the company-specific database
	a.auth = common.New(a.db, companyName)

	// Validate the session token
	user, err := a.auth.ValidateSession(token)
	if err != nil {
		fmt.Printf("Session validation failed: %v\n", err)
		return nil, err
	}

	// Update current user and auth state
	a.currentUser = user
	a.currentCompanyPath = companyName
	a.updateAuthCache()

	fmt.Printf("Session validated for user: %s (ID: %d)\n", user.Username, user.ID)
	return user, nil
}

// GetCurrentUser returns the currently logged-in user
func (a *App) GetCurrentUser() map[string]interface{} {
	if a.currentUser == nil {
		return nil
	}

	return map[string]interface{}{
		"id":           a.currentUser.ID,
		"username":     a.currentUser.Username,
		"email":        a.currentUser.Email,
		"role_id":      a.currentUser.RoleID,
		"role_name":    a.currentUser.RoleName,
		"is_active":    a.currentUser.IsActive,
		"is_root":      a.currentUser.IsRoot,
		"company_name": a.currentCompanyPath,
		"permissions":  a.currentUser.Permissions,
		"last_login":   a.currentUser.LastLogin,
	}
}