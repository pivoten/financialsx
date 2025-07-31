package company

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
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

// GetDBFFiles returns a list of DBF files in the company directory
func GetDBFFiles(companyName string) ([]string, error) {
	datafilesPath, err := getDatafilesPath()
	if err != nil {
		return nil, err
	}
	
	companyPath := filepath.Join(datafilesPath, companyName)
	
	// Check if company directory exists
	if _, err := os.Stat(companyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("company directory does not exist: %s", companyName)
	}
	
	// Find all DBF files (case insensitive)
	var dbfFiles []string
	
	err = filepath.Walk(companyPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".dbf" {
			// Return just the filename, not the full path
			filename := info.Name()
			dbfFiles = append(dbfFiles, filename)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}
	
	return dbfFiles, nil
}

// ReadDBFFile reads a DBF file and returns its structure and data
// If searchTerm is provided, it searches across all records and returns only matching ones
func ReadDBFFile(companyName, fileName, searchTerm string) (map[string]interface{}, error) {
	// Use defer/recover to prevent crashes
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC RECOVERED in ReadDBFFile %s/%s: %v\n", companyName, fileName, r)
		}
	}()
	
	fmt.Printf("ReadDBFFile: %s/%s - reading actual DBF data\n", companyName, fileName)
	
	datafilesPath, err := getDatafilesPath()
	if err != nil {
		return nil, err
	}
	
	filePath := filepath.Join(datafilesPath, companyName, fileName)
	fmt.Printf("Full file path: %s\n", filePath)
	
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("DBF file does not exist: %s", fileName)
	}
	
	// Open the DBF file using the correct API
	fmt.Printf("Attempting to open DBF file...\n")
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   filePath,
		TrimSpaces: true,
	})
	if err != nil {
		fmt.Printf("ERROR opening DBF file: %v\n", err)
		return nil, fmt.Errorf("failed to open DBF file: %w", err)
	}
	defer table.Close()
	fmt.Printf("DBF file opened successfully\n")
	
	// Get column names
	var columns []string
	for _, column := range table.Columns() {
		columns = append(columns, column.Name())
	}
	fmt.Printf("Found %d columns: %v\n", len(columns), columns)
	
	// Get total record count from header
	totalRecords := table.Header().RecordsCount()
	fmt.Printf("Total records in file: %d\n", totalRecords)
	
	// Read all records and count stats
	var rows [][]interface{}
	var deletedCount uint32 = 0
	var activeCount uint32 = 0
	var searchMatches uint32 = 0
	isSearching := searchTerm != ""
	searchLower := strings.ToLower(searchTerm)
	
	fmt.Printf("Starting to read records... (searching: %v)\n", isSearching)
	
	for !table.EOF() {
		row, err := table.Next()
		if err != nil {
			fmt.Printf("Error reading row: %v\n", err)
			break // End of file or error
		}
		
		// Count deleted vs active rows
		if row.Deleted {
			deletedCount++
			if deletedCount <= 10 {
				fmt.Printf("Skipping deleted row at position %d\n", row.Position)
			}
			continue
		}
		
		activeCount++
		
		// Convert row to interface slice
		rowData := make([]interface{}, len(columns))
		matchFound := false
		
		for i, column := range table.Columns() {
			field := row.FieldByName(column.Name())
			if field != nil {
				rowData[i] = field.GetValue()
				
				// Check if this field matches the search term
				if isSearching && !matchFound && field.GetValue() != nil {
					fieldStr := strings.ToLower(fmt.Sprintf("%v", field.GetValue()))
					if strings.Contains(fieldStr, searchLower) {
						matchFound = true
						searchMatches++
					}
				}
			} else {
				rowData[i] = ""
			}
		}
		
		// Only add row if we're not searching or if it matches
		if !isSearching || matchFound {
			rows = append(rows, rowData)
			
			// Limit results
			if len(rows) >= 1000 {
				fmt.Printf("Reached 1000 row limit, stopping\n")
				break
			}
		}
		
		// Log progress every 1000 rows when searching (since we're checking all)
		if isSearching && activeCount%1000 == 0 {
			fmt.Printf("Searched %d active rows, found %d matches so far...\n", activeCount, searchMatches)
		} else if !isSearching && len(rows)%100 == 0 {
			fmt.Printf("Read %d rows so far...\n", len(rows))
		}
	}
	
	if isSearching {
		fmt.Printf("Search complete. Searched %d active rows, found %d matches\n", activeCount, searchMatches)
	} else {
		fmt.Printf("Finished reading. Active rows: %d, Deleted rows: %d\n", activeCount, deletedCount)
	}
	
	fmt.Printf("Successfully read DBF file %s: %d columns, %d rows returned\n", fileName, len(columns), len(rows))
	
	return map[string]interface{}{
		"columns": columns,
		"rows":    rows,
		"stats": map[string]interface{}{
			"totalRecords":   totalRecords,
			"activeRecords":  activeCount,
			"deletedRecords": deletedCount,
			"loadedRecords":  len(rows),
			"hasMoreRecords": activeCount > uint32(len(rows)),
			"searchTerm":     searchTerm,
			"searchMatches":  searchMatches,
		},
	}, nil
}

// UpdateDBFRecord updates a specific cell in a DBF file
// Note: This is a simplified implementation that shows the update concept
// For production use, you'd need more robust DBF editing capabilities
func UpdateDBFRecord(companyName, fileName string, rowIndex, colIndex int, value string) error {
	// For now, return an informative error since DBF editing is complex
	return fmt.Errorf("DBF editing is not yet fully implemented - this would update row %d, column %d with value '%s' in file %s for company %s", 
		rowIndex, colIndex, value, fileName, companyName)
}