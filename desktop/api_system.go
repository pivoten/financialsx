package main

import (
	"fmt"
	"runtime"
)

// ============================================================================
// SYSTEM & UTILITY API
// This file contains all system configuration and utility API methods
// ============================================================================

// GetPlatform returns platform information
func (a *App) GetPlatform() map[string]interface{} {
	return map[string]interface{}{
		"platform":  a.platform,
		"isWindows": a.isWindows,
		"arch":      runtime.GOARCH,
		"version":   runtime.Version(),
	}
}

// GetAuthState returns the current authentication state
func (a *App) GetAuthState() map[string]interface{} {
	return map[string]interface{}{
		"isAuthenticated": a.isAuthenticated,
		"isAdmin":         a.isAdmin,
		"isRoot":          a.isRoot,
		"userRole":        a.userRole,
		"permissions":     a.permissions,
	}
}

// Greet returns a greeting message (example method)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetDashboardData retrieves dashboard data for a company
func (a *App) GetDashboardData(companyIdentifier string) (map[string]interface{}, error) {
	// Determine if it's a company name or path
	companyName := companyIdentifier

	// Initialize the database for this company if needed
	if err := a.InitializeCompanyDatabase(companyName); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	// Get bank accounts
	bankAccounts, err := a.GetBankAccounts(companyName)
	if err != nil {
		// Don't fail the entire dashboard if bank accounts fail
		bankAccounts = []map[string]interface{}{}
	}

	// Build dashboard data
	dashboardData := map[string]interface{}{
		"company":       companyName,
		"bankAccounts":  bankAccounts,
		"quickStats":    getQuickStats(),
		"recentChecks":  []map[string]interface{}{}, // TODO: Implement
		"alerts":        []map[string]interface{}{}, // TODO: Implement
	}

	return dashboardData, nil
}

// GetAPIKey retrieves an API key for a service
func (a *App) GetAPIKey(service string) (string, error) {
	if err := a.requirePermission("config.read"); err != nil {
		return "", err
	}

	// TODO: Implement secure API key storage
	return "", fmt.Errorf("API key storage not yet implemented")
}

// SetAPIKey stores an API key for a service
func (a *App) SetAPIKey(service, key string) error {
	if err := a.requirePermission("config.write"); err != nil {
		return err
	}

	// TODO: Implement secure API key storage
	return fmt.Errorf("API key storage not yet implemented")
}

// TestAPIKey tests if an API key is valid
func (a *App) TestAPIKey(service, key string) (bool, error) {
	// TODO: Implement API key testing
	return false, fmt.Errorf("API key testing not yet implemented")
}

// GetConfig retrieves application configuration
func (a *App) GetConfig() (map[string]interface{}, error) {
	config := map[string]interface{}{
		"platform":       a.platform,
		"isWindows":      a.isWindows,
		"currentCompany": a.currentCompanyPath,
		"dataBasePath":   a.dataBasePath,
		"features": map[string]bool{
			"banking":        true,
			"reconciliation": true,
			"reporting":      true,
			"vfp":            a.VFPWrapper != nil,
		},
	}

	return config, nil
}

// getQuickStats returns quick statistics for the dashboard
func getQuickStats() map[string]interface{} {
	return map[string]interface{}{
		"totalAccounts":       0,
		"outstandingChecks":   0,
		"lastReconciliation":  nil,
		"pendingTransactions": 0,
	}
}