package main

import (
	"fmt"

	"github.com/pivoten/financialsx/desktop/internal/common"
)

// ============================================================================
// USER MANAGEMENT API
// This file contains all user management-related API methods
// ============================================================================

// GetAllUsers retrieves all users from the database
func (a *App) GetAllUsers() ([]common.User, error) {
	if err := a.requirePermission("user.read"); err != nil {
		return nil, err
	}

	if a.auth == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}

	return a.auth.GetAllUsers()
}

// GetAllRoles retrieves all available roles
func (a *App) GetAllRoles() ([]common.Role, error) {
	if err := a.requirePermission("role.read"); err != nil {
		return nil, err
	}

	if a.auth == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}

	return a.auth.GetAllRoles()
}

// UpdateUserRole updates a user's role
func (a *App) UpdateUserRole(userID, newRoleID int) error {
	if err := a.requireAdminOrRoot(); err != nil {
		return err
	}

	if a.auth == nil {
		return fmt.Errorf("auth service not initialized")
	}

	return a.auth.UpdateUserRole(userID, newRoleID)
}

// UpdateUserStatus updates a user's active status
func (a *App) UpdateUserStatus(userID int, isActive bool) error {
	if err := a.requireAdminOrRoot(); err != nil {
		return err
	}

	if a.auth == nil {
		return fmt.Errorf("auth service not initialized")
	}

	return a.auth.UpdateUserStatus(userID, isActive)
}

// CreateUser creates a new user account
func (a *App) CreateUser(username, password, email string, roleID int) (*common.User, error) {
	if err := a.requireAdminOrRoot(); err != nil {
		return nil, err
	}

	if a.auth == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}

	// Use the auth service to create the user
	user, err := a.auth.CreateUser(username, password, email, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	fmt.Printf("User created successfully: %s (ID: %d)\n", user.Username, user.ID)
	return user, nil
}

// DeleteUser deletes a user account
func (a *App) DeleteUser(userID int) error {
	if err := a.requireAdminOrRoot(); err != nil {
		return err
	}

	if a.auth == nil {
		return fmt.Errorf("auth service not initialized")
	}

	// TODO: Implement in auth service
	return fmt.Errorf("user deletion not yet implemented")
}

// ResetUserPassword resets a user's password
func (a *App) ResetUserPassword(userID int, newPassword string) error {
	if err := a.requireAdminOrRoot(); err != nil {
		return err
	}

	if a.auth == nil {
		return fmt.Errorf("auth service not initialized")
	}

	// TODO: Implement in auth service
	return fmt.Errorf("password reset not yet implemented")
}

// GetUserPermissions gets permissions for a specific user
func (a *App) GetUserPermissions(userID int) ([]string, error) {
	if err := a.requirePermission("user.read"); err != nil {
		return nil, err
	}

	if a.auth == nil {
		return nil, fmt.Errorf("auth service not initialized")
	}

	// TODO: Implement in auth service
	return nil, fmt.Errorf("get user permissions not yet implemented")
}