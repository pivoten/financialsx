package company

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getDatafilesPath returns the path to the datafiles directory
// It looks for the directory in multiple locations to handle both dev and production scenarios
func getDatafilesPath() (string, error) {
	// Possible locations for datafiles directory
	possiblePaths := []string{
		"../datafiles",       // One level up (dev from desktop folder) - CHECK THIS FIRST
		"./datafiles",        // Current directory (production)
		"../../datafiles",    // Two levels up (if nested deeper)
	}
	
	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			// Check if this directory has any company folders (not empty)
			entries, err := os.ReadDir(path)
			if err == nil {
				// Count non-hidden directories (companies)
				companyCount := 0
				for _, entry := range entries {
					if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") && strings.ToLower(entry.Name()) != "datafiles" {
						companyCount++
					}
				}
				// Prefer directories that have companies in them
				if companyCount > 0 {
					fmt.Printf("Found datafiles path with %d companies: %s\n", companyCount, path)
					return path, nil
				}
			}
		}
	}
	
	// If no populated directories found, try again but accept empty ones
	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			fmt.Printf("Using empty datafiles path: %s\n", path)
			return path, nil
		}
	}
	
	// If not found, create in current directory
	datafilesPath := "./datafiles"
	if err := os.MkdirAll(datafilesPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create datafiles directory: %w", err)
	}
	
	fmt.Printf("Created new datafiles path: %s\n", datafilesPath)
	return datafilesPath, nil
}

type Company struct {
	Name     string `json:"name"`
	DataPath string `json:"data_path"`
	HasDBF   bool   `json:"has_dbf"`
	HasSQL   bool   `json:"has_sql"`
}

// DetectCompanies scans the datafiles directory for company folders
func DetectCompanies() ([]Company, error) {
	datafilesPath, err := getDatafilesPath()
	if err != nil {
		fmt.Printf("Error getting datafiles path: %v\n", err)
		return []Company{}, err
	}
	
	fmt.Printf("Using datafiles path: %s\n", datafilesPath)

	entries, err := os.ReadDir(datafilesPath)
	if err != nil {
		fmt.Printf("Error reading datafiles directory: %v\n", err)
		return []Company{}, fmt.Errorf("failed to read datafiles directory: %w", err)
	}
	
	fmt.Printf("Found %d entries in datafiles directory\n", len(entries))

	var companies []Company
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		companyName := entry.Name()
		
		// Skip if the folder is named "datafiles"
		if strings.ToLower(companyName) == "datafiles" {
			continue
		}
		
		companyPath := filepath.Join(datafilesPath, companyName)
		
		company := Company{
			Name:     companyName,
			DataPath: companyPath,
		}

		// Check for SQL directory
		sqlPath := filepath.Join(companyPath, "sql")
		if info, err := os.Stat(sqlPath); err == nil && info.IsDir() {
			company.HasSQL = true
		}

		// Check for DBF files
		dbfFiles, _ := filepath.Glob(filepath.Join(companyPath, "*.dbf"))
		if len(dbfFiles) > 0 {
			company.HasDBF = true
		}

		companies = append(companies, company)
	}

	fmt.Printf("Returning %d companies: %+v\n", len(companies), companies)
	return companies, nil
}

// ValidateCompanyName checks if a company name is valid for directory creation
func ValidateCompanyName(name string) error {
	if name == "" {
		return fmt.Errorf("company name cannot be empty")
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("company name contains invalid character: %s", char)
		}
	}

	// Check if it already exists
	datafilesPath, err := getDatafilesPath()
	if err != nil {
		return err
	}
	
	companyPath := filepath.Join(datafilesPath, name)
	if _, err := os.Stat(companyPath); err == nil {
		return fmt.Errorf("company '%s' already exists", name)
	}

	return nil
}

// CreateCompanyDirectory creates a new company directory structure
func CreateCompanyDirectory(name string) error {
	// Only validate for empty name and invalid characters, not existence
	if name == "" {
		return fmt.Errorf("company name cannot be empty")
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("company name contains invalid character: %s", char)
		}
	}

	datafilesPath, err := getDatafilesPath()
	if err != nil {
		return err
	}

	companyPath := filepath.Join(datafilesPath, name)
	sqlPath := filepath.Join(companyPath, "sql")

	// Create directories (os.MkdirAll doesn't fail if directory already exists)
	if err := os.MkdirAll(sqlPath, 0755); err != nil {
		return fmt.Errorf("failed to create company directories: %w", err)
	}

	return nil
}