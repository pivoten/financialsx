package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/database"
	"github.com/pivoten/financialsx/desktop/internal/reconciliation"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ============================================================================
// COMPANY MANAGEMENT API
// This file contains all company-related API methods
// ============================================================================

// GetCompanyList returns the list of available companies
func (a *App) GetCompanyList() ([]map[string]interface{}, error) {
	return a.Services.Company.GetCompanyList()
}

// GetCompanies returns company list (legacy method for compatibility)
func (a *App) GetCompanies() ([]company.Company, error) {
	return company.DetectCompanies()
}

// GetCompanyInfo retrieves detailed information about a company
func (a *App) GetCompanyInfo(companyName string) (map[string]interface{}, error) {
	return a.Services.Company.GetCompanyInfo(companyName)
}

// SelectDataFolder allows user to manually select the data folder
func (a *App) SelectDataFolder() (string, error) {
	folderPath, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select folder containing compmast.dbf",
	})
	if err != nil {
		return "", err
	}
	if folderPath == "" {
		return "", fmt.Errorf("no folder selected")
	}

	// Verify compmast.dbf exists in selected folder
	compMastPath := filepath.Join(folderPath, "compmast.dbf")
	if !fileExists(compMastPath) {
		// Try uppercase
		compMastPath = filepath.Join(folderPath, "COMPMAST.DBF")
		if !fileExists(compMastPath) {
			return "", fmt.Errorf("compmast.dbf not found in selected folder")
		}
	}

	// Save the path for future use
	if a.Services != nil && a.Services.Company != nil {
		if err := a.Services.Company.SaveDataPath(folderPath); err != nil {
			// Non-critical error, log but continue
			fmt.Printf("Warning: Could not save data path: %v\n", err)
		}
	}

	// Update the base path
	a.dataBasePath = folderPath
	if a.Services != nil && a.Services.Company != nil {
		a.Services.Company.SetDataPath(folderPath)
	}

	return folderPath, nil
}

// SetDataPath sets the base data path for company files
func (a *App) SetDataPath(dataPath string) error {
	if err := a.requirePermission("system.config"); err != nil {
		return err
	}

	a.dataBasePath = dataPath
	if a.Services != nil && a.Services.Company != nil {
		a.Services.Company.SetDataPath(dataPath)
		return a.Services.Company.SaveDataPath(dataPath)
	}
	return nil
}

// InitializeCompanyDatabase initializes the SQLite database for a company
func (a *App) InitializeCompanyDatabase(companyPath string) error {
	// Initialize database if not already done
	if a.db == nil {
		fmt.Printf("Initializing database for company: %s\n", companyPath)

		db, err := database.New(companyPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %v", err)
		}

		a.db = db

		// Initialize reconciliation service with the database
		a.reconciliationService = reconciliation.NewService(db)

		// Initialize balance cache tables
		if err := database.InitializeBalanceCache(db); err != nil {
			fmt.Printf("Warning: Failed to initialize balance cache: %v\n", err)
			// Don't fail completely, cache is optional
		}

		// Initialize services that need the database
		if a.Services != nil && a.Services.Banking != nil {
			a.Services.Banking.SetDatabaseHelper(a.db)
		}
	}

	return nil
}

// fileExists is a helper function to check if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}