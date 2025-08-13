package company

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
	"github.com/pivoten/financialsx/desktop/internal/debug"
)

// Cache the datafiles path to avoid repeated directory scanning
var (
	cachedDatafilesPath string
	datafilesPathMutex  sync.RWMutex
	isWindows          bool = runtime.GOOS == "windows"
	platform           string = runtime.GOOS
)

// normalizeCompanyPath converts a company path based on the current platform
// On Windows: Use the path as-is (absolute or relative)
// On Mac/Linux: Extract just the company folder name since folders are relative to compmast.dbf
func normalizeCompanyPath(companyPath string) string {
	// If empty, return as-is
	if companyPath == "" {
		return companyPath
	}
	
	// On Windows, use the path as provided
	if isWindows {
		return companyPath
	}
	
	// On Mac/Linux, extract just the company folder name
	// Handle Windows-style paths (from Windows-created DBF files)
	if strings.Contains(companyPath, "\\") {
		// Split by backslash and get the last non-empty component
		parts := strings.Split(companyPath, "\\")
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] != "" {
				return parts[i]
			}
		}
	}
	
	// Handle Unix-style paths
	if strings.Contains(companyPath, "/") {
		return filepath.Base(companyPath)
	}
	
	// If it's already just a folder name, return as-is
	return companyPath
}

// writeErrorLog writes error messages to a log file
func writeErrorLog(message string) {
	exePath, _ := os.Executable()
	logDir := filepath.Join(filepath.Dir(exePath), "logs")
	os.MkdirAll(logDir, 0755)
	
	dateStamp := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("financialsx_dbf_%s.log", dateStamp))
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	file.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, message))
}

// getDatafilesPath returns the path to the datafiles directory
// It looks for the directory in multiple locations to handle both dev and production scenarios
// Uses caching to avoid repeated scanning after first discovery
func getDatafilesPath() (string, error) {
	// Check cache first
	datafilesPathMutex.RLock()
	if cachedDatafilesPath != "" {
		defer datafilesPathMutex.RUnlock()
		return cachedDatafilesPath, nil
	}
	datafilesPathMutex.RUnlock()
	
	// Need to find the path - upgrade to write lock
	datafilesPathMutex.Lock()
	defer datafilesPathMutex.Unlock()
	
	// Double-check after acquiring write lock (another goroutine might have set it)
	if cachedDatafilesPath != "" {
		return cachedDatafilesPath, nil
	}
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
					cachedDatafilesPath = path
					return path, nil
				}
			}
		}
	}
	
	// If no populated directories found, try again but accept empty ones
	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			fmt.Printf("Using empty datafiles path: %s\n", path)
			cachedDatafilesPath = path
			return path, nil
		}
	}
	
	// If not found, create in current directory
	datafilesPath := "./datafiles"
	if err := os.MkdirAll(datafilesPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create datafiles directory: %w", err)
	}
	
	fmt.Printf("Created new datafiles path: %s\n", datafilesPath)
	cachedDatafilesPath = datafilesPath
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
	// Normalize the company path based on platform
	name = normalizeCompanyPath(name)
	
	// Only validate for empty name and invalid characters, not existence
	if name == "" {
		return fmt.Errorf("company name cannot be empty")
	}

	// Check for invalid characters in the normalized name
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
	debug.LogInfo("GetDBFFiles", fmt.Sprintf("Called with: %s", companyName))
	
	// Normalize the company path based on platform
	companyName = normalizeCompanyPath(companyName)
	debug.LogInfo("GetDBFFiles", fmt.Sprintf("Normalized to: %s", companyName))
	writeErrorLog(fmt.Sprintf("GetDBFFiles: Normalized company path: %s", companyName))
	
	var companyPath string
	
	// Check if companyName looks like a full path (absolute)
	if filepath.IsAbs(companyName) {
		// It's an absolute path, use it directly (Windows scenario)
		companyPath = companyName
		writeErrorLog(fmt.Sprintf("GetDBFFiles: Using absolute path: %s", companyPath))
		debug.LogInfo("GetDBFFiles", fmt.Sprintf("Using absolute path: %s", companyPath))
	} else if strings.Contains(companyName, string(os.PathSeparator)) || strings.Contains(companyName, "/") {
		// It's a relative path - resolve it
		if isWindows {
			// On Windows, make it relative to executable directory
			exePath, _ := os.Executable()
			exeDir := filepath.Dir(exePath)
			companyPath = filepath.Join(exeDir, companyName)
		} else {
			// On Mac/Linux, it should be relative to datafiles
			datafilesPath, err := getDatafilesPath()
			if err != nil {
				return nil, err
			}
			companyPath = filepath.Join(datafilesPath, filepath.Base(companyName))
		}
		writeErrorLog(fmt.Sprintf("GetDBFFiles: Using relative path: %s (resolved to: %s)", companyName, companyPath))
		debug.LogInfo("GetDBFFiles", fmt.Sprintf("Using relative path: %s -> %s", companyName, companyPath))
	} else {
		// Just a folder name - use the datafiles structure
		datafilesPath, err := getDatafilesPath()
		if err != nil {
			return nil, err
		}
		companyPath = filepath.Join(datafilesPath, companyName)
		writeErrorLog(fmt.Sprintf("GetDBFFiles: Using folder name: %s -> %s", companyName, companyPath))
		debug.LogInfo("GetDBFFiles", fmt.Sprintf("Using folder name: %s -> %s", companyName, companyPath))
	}
	
	// Check if company directory exists
	if _, err := os.Stat(companyPath); os.IsNotExist(err) {
		debug.LogError("GetDBFFiles", fmt.Errorf("company directory does not exist: %s", companyPath))
		return nil, fmt.Errorf("company directory does not exist: %s", companyPath)
	}
	
	// Get absolute path for debugging
	absPath, _ := filepath.Abs(companyPath)
	debug.LogInfo("GetDBFFiles", fmt.Sprintf("Company directory exists at: %s", absPath))
	writeErrorLog(fmt.Sprintf("GetDBFFiles: Scanning directory: %s", absPath))
	
	// Find all DBF files (case insensitive)
	var dbfFiles []string
	
	fileCount := 0
	err := filepath.Walk(companyPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			writeErrorLog(fmt.Sprintf("GetDBFFiles: Error walking path %s: %v", path, err))
			return err
		}
		
		fileCount++
		ext := strings.ToLower(filepath.Ext(path))
		if !info.IsDir() {
			writeErrorLog(fmt.Sprintf("GetDBFFiles: Found file: %s (ext: %s)", info.Name(), ext))
			if ext == ".dbf" {
				// Return just the filename, not the full path
				filename := info.Name()
				dbfFiles = append(dbfFiles, filename)
				debug.LogInfo("GetDBFFiles", fmt.Sprintf("Added DBF file: %s", filename))
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}
	
	writeErrorLog(fmt.Sprintf("GetDBFFiles: Scanned %d total items, found %d DBF files", fileCount, len(dbfFiles)))
	debug.LogInfo("GetDBFFiles", fmt.Sprintf("Found %d DBF files out of %d total items", len(dbfFiles), fileCount))
	return dbfFiles, nil
}

// ReadDBFFileDirectly reads a DBF file from a specific path without company context
func ReadDBFFileDirectly(filePath, searchTerm string, offset, limit int, sortColumn, sortDirection string) (map[string]interface{}, error) {
	// Use defer/recover to prevent crashes
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC RECOVERED in ReadDBFFileDirectly %s: %v\n", filePath, r)
		}
	}()
	
	fmt.Printf("ReadDBFFileDirectly: %s - reading actual DBF data\n", filePath)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("DBF file does not exist: %s", filePath)
	}
	
	// Open the DBF file using the same API as ReadDBFFile
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   filePath,
		TrimSpaces: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open DBF file: %w", err)
	}
	defer table.Close()
	
	// Get column names
	var columns []string
	for _, column := range table.Columns() {
		columns = append(columns, column.Name())
	}
	
	// Get total record count from header
	totalRecords := table.Header().RecordsCount()
	fmt.Printf("Total records in file: %d\n", totalRecords)
	debug.LogInfo("ReadDBFFileDirectly", fmt.Sprintf("DBF header shows %d total records", totalRecords))
	
	// Read all rows
	var rows []map[string]interface{}
	totalCount := 0
	skipped := 0
	searchLower := strings.ToLower(searchTerm)
	isSearching := searchTerm != ""
	
	for !table.EOF() {
		row, err := table.Next()
		if err != nil {
			break
		}
		
		if row.Deleted {
			continue
		}
		
		// Convert row to map
		rowData := make(map[string]interface{})
		matchFound := false
		
		for _, column := range table.Columns() {
			field := row.FieldByName(column.Name())
			if field != nil {
				value := field.GetValue()
				rowData[column.Name()] = value
				
				// Special logging for CTAXID field in VENDOR.DBF
				baseFileName := filepath.Base(filePath)
				if strings.ToUpper(baseFileName) == "VENDOR.DBF" && column.Name() == "CTAXID" && value != nil {
					valueStr := fmt.Sprintf("%v", value)
					if valueStr != "" && strings.TrimSpace(valueStr) != "" {
						// Log the raw bytes of the CTAXID field
						bytes := []byte(valueStr)
						hexStr := ""
						for _, b := range bytes {
							hexStr += fmt.Sprintf("%02x ", b)
						}
						writeErrorLog(fmt.Sprintf("VENDOR CTAXID found - Raw: '%s', Length: %d, Hex: %s", 
							valueStr, len(valueStr), hexStr))
						debug.LogInfo("ReadDBFFileDirectly", fmt.Sprintf("VENDOR CTAXID - Hex: %s", hexStr))
					}
				}
				
				// Check if this field matches the search term
				if isSearching && !matchFound && value != nil {
					fieldStr := strings.ToLower(fmt.Sprintf("%v", value))
					if strings.Contains(fieldStr, searchLower) {
						matchFound = true
					}
				}
			} else {
				rowData[column.Name()] = ""
			}
		}
		
		// Skip if searching and no match found
		if isSearching && !matchFound {
			continue
		}
		
		totalCount++
		
		// Apply pagination
		if offset > 0 && skipped < offset {
			skipped++
			continue
		}
		
		if limit > 0 && len(rows) >= limit {
			continue
		}
		
		rows = append(rows, rowData)
	}
	
	// Build column info for output
	columnInfo := []map[string]interface{}{}
	for _, colName := range columns {
		columnInfo = append(columnInfo, map[string]interface{}{
			"name": colName,
			"type": "string", // We don't have type info readily available
		})
	}
	
	return map[string]interface{}{
		"columns":     columnInfo,
		"rows":        rows,
		"totalCount":  totalCount,
		"offset":      offset,
		"limit":       limit,
		"searchTerm":  searchTerm,
		"fileName":    filepath.Base(filePath),
	}, nil
}

// ReadDBFFile reads a DBF file and returns its structure and data with pagination and sorting
// If searchTerm is provided, it searches across all records and returns only matching ones
//
// ⚠️ CRITICAL WARNING: NEVER use arbitrary limits for financial calculations!
// For GL balances, outstanding checks, or any financial reporting, ALWAYS use:
//   offset=0, limit=0 (which means read ALL records)
//
// Only use non-zero limits for:
//   - UI pagination/display
//   - User-requested limited views  
//   - Quick data sampling/preview
//
// Example of CORRECT usage for financial calculations:
//   ReadDBFFile(companyName, "GLMASTER.dbf", "", 0, 0, "", "") // Reads ALL records
//
// A hardcoded limit of 50,000 caused a $400,000 discrepancy in GL calculations!
func ReadDBFFile(companyName, fileName, searchTerm string, offset, limit int, sortColumn, sortDirection string) (map[string]interface{}, error) {
	// Use defer/recover to prevent crashes
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC RECOVERED in ReadDBFFile %s/%s: %v\n", companyName, fileName, r)
		}
	}()
	
	fmt.Printf("ReadDBFFile: %s/%s - reading actual DBF data\n", companyName, fileName)
	debug.LogInfo("ReadDBFFile", fmt.Sprintf("Called with company=%s, file=%s", companyName, fileName))
	
	var filePath string
	
	// Log the incoming parameters
	writeErrorLog(fmt.Sprintf("ReadDBFFile: START - company='%s', file='%s'", companyName, fileName))
	
	// Normalize the company path based on platform
	companyName = normalizeCompanyPath(companyName)
	debug.LogInfo("ReadDBFFile", fmt.Sprintf("Normalized to: %s", companyName))
	writeErrorLog(fmt.Sprintf("ReadDBFFile: Normalized company path: %s", companyName))
	
	// Check if companyName looks like a full path (absolute)
	if filepath.IsAbs(companyName) {
		// It's an absolute path, use it directly (Windows scenario)
		filePath = filepath.Join(companyName, fileName)
		writeErrorLog(fmt.Sprintf("ReadDBFFile: Using absolute path, result: %s", filePath))
		debug.LogInfo("ReadDBFFile", fmt.Sprintf("Using absolute path: %s", filePath))
	} else if strings.Contains(companyName, string(os.PathSeparator)) || strings.Contains(companyName, "/") {
		// It's a relative path - resolve it
		if isWindows {
			// On Windows, make it relative to executable directory
			exePath, _ := os.Executable()
			exeDir := filepath.Dir(exePath)
			filePath = filepath.Join(exeDir, companyName, fileName)
		} else {
			// On Mac/Linux, it should be relative to datafiles
			datafilesPath, err := getDatafilesPath()
			if err != nil {
				writeErrorLog(fmt.Sprintf("ReadDBFFile: Failed to get datafiles path: %v", err))
				return nil, err
			}
			filePath = filepath.Join(datafilesPath, filepath.Base(companyName), fileName)
		}
		writeErrorLog(fmt.Sprintf("ReadDBFFile: Using relative path, result: %s", filePath))
		debug.LogInfo("ReadDBFFile", fmt.Sprintf("Using relative path: %s", filePath))
	} else {
		// Just a folder name - use the datafiles structure
		datafilesPath, err := getDatafilesPath()
		if err != nil {
			writeErrorLog(fmt.Sprintf("ReadDBFFile: Failed to get datafiles path: %v", err))
			return nil, err
		}
		filePath = filepath.Join(datafilesPath, companyName, fileName)
		writeErrorLog(fmt.Sprintf("ReadDBFFile: Using folder name, datafilesPath='%s', result='%s'", datafilesPath, filePath))
		debug.LogInfo("ReadDBFFile", fmt.Sprintf("Using folder name: %s", filePath))
	}
	fmt.Printf("Full file path: %s\n", filePath)
	
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		writeErrorLog(fmt.Sprintf("ReadDBFFile: DBF file does not exist at path %s: %v", filePath, err))
		debug.LogError("ReadDBFFile", fmt.Errorf("DBF file does not exist at path %s: %v", filePath, err))
		return nil, fmt.Errorf("DBF file does not exist: %s", fileName)
	}
	writeErrorLog(fmt.Sprintf("ReadDBFFile: File exists at: %s", filePath))
	debug.LogInfo("ReadDBFFile", fmt.Sprintf("File exists at: %s", filePath))
	
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
	debug.LogInfo("ReadDBFFile", fmt.Sprintf("DBF header shows %d total records in %s", totalRecords, fileName))
	
	// Read and potentially sort all records first (for server-side sorting)
	var allRows [][]interface{}
	var deletedCount uint32 = 0
	var activeCount uint32 = 0
	var searchMatches uint32 = 0
	isSearching := searchTerm != ""
	searchLower := strings.ToLower(searchTerm)
	needsSorting := sortColumn != ""
	
	fmt.Printf("Starting to read records... (searching: %v, sorting: %v)\n", isSearching, needsSorting)
	
	// First pass: read all records (needed for sorting)
	for !table.EOF() {
		row, err := table.Next()
		if err != nil {
			fmt.Printf("Error reading row: %v\n", err)
			break
		}
		
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
					}
				}
			} else {
				rowData[i] = ""
			}
		}
		
		// Only add row if we're not searching or if it matches
		if !isSearching || matchFound {
			allRows = append(allRows, rowData)
			if matchFound {
				searchMatches++
			}
		}
		
		// Log progress
		if activeCount%1000 == 0 {
			fmt.Printf("Processed %d active rows...\n", activeCount)
		}
	}
	
	fmt.Printf("Read %d total matching rows\n", len(allRows))
	
	// Server-side sorting
	if needsSorting && len(allRows) > 0 {
		sortColumnIndex := -1
		for i, col := range columns {
			if col == sortColumn {
				sortColumnIndex = i
				break
			}
		}
		
		if sortColumnIndex >= 0 {
			fmt.Printf("Sorting by column %s (%d) %s\n", sortColumn, sortColumnIndex, sortDirection)
			sort.Slice(allRows, func(i, j int) bool {
				aVal := allRows[i][sortColumnIndex]
				bVal := allRows[j][sortColumnIndex]
				
				// Handle null/empty values
				if aVal == nil && bVal == nil {
					return false
				}
				if aVal == nil {
					return sortDirection != "desc"
				}
				if bVal == nil {
					return sortDirection == "desc"
				}
				
				// Try to parse as time first (for date columns)
				if aTime, aErr := parseDateTime(aVal); aErr == nil {
					if bTime, bErr := parseDateTime(bVal); bErr == nil {
						if sortDirection == "desc" {
							return aTime.After(bTime)
						}
						return aTime.Before(bTime)
					}
				}
				
				// Try numeric comparison
				if aNum, aErr := parseNumber(aVal); aErr == nil {
					if bNum, bErr := parseNumber(bVal); bErr == nil {
						if sortDirection == "desc" {
							return aNum > bNum
						}
						return aNum < bNum
					}
				}
				
				// Fall back to string comparison
				aStr := fmt.Sprintf("%v", aVal)
				bStr := fmt.Sprintf("%v", bVal)
				if sortDirection == "desc" {
					return aStr > bStr
				}
				return aStr < bStr
			})
		}
	}
	
	// Apply pagination (limit = 0 means no limit)
	totalRows := len(allRows)
	startIdx := offset
	endIdx := totalRows // Default to end of data
	
	fmt.Printf("Pagination: offset=%d, limit=%d, totalRows=%d\n", offset, limit, totalRows)
	
	if limit > 0 { // Only apply limit if it's specified
		endIdx = offset + limit
		fmt.Printf("Pagination: Applied limit, endIdx=%d\n", endIdx)
	} else {
		fmt.Printf("Pagination: No limit specified, endIdx=%d (all rows)\n", endIdx)
	}
	
	if startIdx >= totalRows {
		startIdx = totalRows
	}
	if endIdx > totalRows {
		endIdx = totalRows
	}
	
	var rows [][]interface{}
	if startIdx < endIdx {
		rows = allRows[startIdx:endIdx]
	}
	
	fmt.Printf("Returning page %d-%d of %d total rows\n", startIdx, endIdx, totalRows)
	
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
			"totalMatching":  totalRows,
			"hasMoreRecords": totalRows > len(rows),
			"searchTerm":     searchTerm,
			"searchMatches":  searchMatches,
			"offset":         offset,
			"limit":          limit,
			"sortColumn":     sortColumn,
			"sortDirection":  sortDirection,
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

// GetDashboardData returns lightweight dashboard data with well types
func GetDashboardData(companyName string) (map[string]interface{}, error) {
	fmt.Printf("Getting dashboard data for company: %s\n", companyName)
	writeErrorLog(fmt.Sprintf("GetDashboardData: Called with company: %s", companyName))
	debug.LogInfo("GetDashboardData", fmt.Sprintf("Called with company: %s", companyName))
	
	dashboard := map[string]interface{}{
		"company": companyName,
		"widgets": map[string]interface{}{},
		"status": "ready",
		"message": "Dashboard loaded successfully",
	}
	
	// Get well types for top cards (lightweight operation)
	writeErrorLog("GetDashboardData: Getting well types...")
	debug.LogInfo("GetDashboardData", "Getting well types...")
	if wellTypes, err := getWellTypes(companyName); err == nil {
		dashboard["wellTypes"] = wellTypes
		writeErrorLog(fmt.Sprintf("GetDashboardData: Got %d well types", len(wellTypes)))
		debug.LogInfo("GetDashboardData", fmt.Sprintf("Got %d well types", len(wellTypes)))
	} else {
		writeErrorLog(fmt.Sprintf("GetDashboardData: Failed to get well types: %v", err))
		debug.LogError("GetDashboardData", fmt.Errorf("failed to get well types: %v", err))
	}
	
	writeErrorLog(fmt.Sprintf("GetDashboardData: Returning dashboard: %+v", dashboard))
	return dashboard, nil
}

// Helper function to get basic file statistics
func getFileStatistics(companyName string) (map[string]interface{}, error) {
	dbfFiles, err := GetDBFFiles(companyName)
	if err != nil {
		return nil, err
	}
	
	stats := map[string]interface{}{
		"totalFiles": len(dbfFiles),
		"files":      dbfFiles,
	}
	
	return stats, nil
}

// Helper function to analyze financial data
func getFinancialSummary(companyName string) (map[string]interface{}, error) {
	financials := map[string]interface{}{
		"totalIncome":  0.0,
		"totalExpense": 0.0,
		"netIncome":    0.0,
		"hasData":      false,
	}
	
	// Try to count INCOME.DBF records
	if incomeCount, err := getRecordCount(companyName, "INCOME.DBF"); err == nil && incomeCount > 0 {
		financials["incomeRecords"] = incomeCount
		financials["hasIncomeData"] = true
		financials["hasData"] = true
	}
	
	// Try to count EXPENSE.DBF records
	if expenseCount, err := getRecordCount(companyName, "EXPENSE.DBF"); err == nil && expenseCount > 0 {
		financials["expenseRecords"] = expenseCount
		financials["hasExpenseData"] = true
		financials["hasData"] = true
	}
	
	return financials, nil
}

// Helper function to get well information
func getWellSummary(companyName string) (map[string]interface{}, error) {
	wells := map[string]interface{}{
		"totalWells": 0,
		"hasData":    false,
	}
	
	// Use the new count function instead of reading all data
	if count, err := getRecordCount(companyName, "WELLS.DBF"); err == nil {
		wells["totalWells"] = count
		wells["hasData"] = count > 0
	}
	
	return wells, nil
}

// Helper function to get check activity
func getCheckActivity(companyName string) (map[string]interface{}, error) {
	checks := map[string]interface{}{
		"totalChecks": 0,
		"hasData":     false,
	}
	
	// Use the new count function instead of reading all data
	if count, err := getRecordCount(companyName, "CHECKS.DBF"); err == nil {
		checks["totalChecks"] = count
		checks["hasData"] = count > 0
	}
	
	return checks, nil
}

// getRecordCount efficiently counts active records in a DBF file without loading all data
func getRecordCount(companyName, fileName string) (uint32, error) {
	datafilesPath, err := getDatafilesPath()
	if err != nil {
		return 0, err
	}
	
	filePath := filepath.Join(datafilesPath, companyName, fileName)
	
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		return 0, fmt.Errorf("DBF file does not exist: %s", fileName)
	}
	
	// Open the DBF file
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   filePath,
		TrimSpaces: true,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to open DBF file: %w", err)
	}
	defer table.Close()
	
	// Count active records without loading data
	var activeCount uint32 = 0
	
	for !table.EOF() {
		row, err := table.Next()
		if err != nil {
			break // End of file or error
		}
		
		// Only count non-deleted rows
		if !row.Deleted {
			activeCount++
		}
	}
	
	fmt.Printf("Counted %d active records in %s\n", activeCount, fileName)
	return activeCount, nil
}

// Helper function to parse various date/time formats
func parseDateTime(value interface{}) (time.Time, error) {
	if value == nil {
		return time.Time{}, fmt.Errorf("nil value")
	}
	
	str := fmt.Sprintf("%v", value)
	
	// Try common date formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01/02/2006",
		"01/02/06",
		"2006/01/02",
		"06/01/02",
		"20060102",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %v", str)
}

// Helper function to parse numeric values
func parseNumber(value interface{}) (float64, error) {
	if value == nil {
		return 0, fmt.Errorf("nil value")
	}
	
	str := strings.TrimSpace(fmt.Sprintf("%v", value))
	
	// Remove common currency symbols and commas
	str = strings.ReplaceAll(str, "$", "")
	str = strings.ReplaceAll(str, ",", "")
	
	return strconv.ParseFloat(str, 64)
}

// getWellTypes returns a lightweight summary of well types from WELLS.DBF
func getWellTypes(companyName string) ([]map[string]interface{}, error) {
	fmt.Printf("getWellTypes: Reading WELLS.dbf for company: %s\n", companyName)
	writeErrorLog(fmt.Sprintf("getWellTypes: Reading WELLS.dbf for company: %s", companyName))
	debug.LogInfo("getWellTypes", fmt.Sprintf("Reading WELLS.dbf for company: %s", companyName))
	
	// Read WELLS.DBF file
	wellsData, err := ReadDBFFile(companyName, "WELLS.dbf", "", 0, 0, "", "")
	if err != nil {
		fmt.Printf("getWellTypes: Failed to read WELLS.dbf: %v\n", err)
		writeErrorLog(fmt.Sprintf("getWellTypes: Failed to read WELLS.dbf: %v", err))
		debug.LogError("getWellTypes", fmt.Errorf("failed to read WELLS.dbf: %v", err))
		return nil, fmt.Errorf("failed to read WELLS.dbf: %w", err)
	}
	
	columns, ok := wellsData["columns"].([]string)
	if !ok {
		fmt.Printf("getWellTypes: Invalid WELLS.dbf structure - no columns\n")
		return nil, fmt.Errorf("invalid WELLS.dbf structure - no columns")
	}
	
	rows, ok := wellsData["rows"].([][]interface{})
	if !ok {
		fmt.Printf("getWellTypes: Invalid WELLS.dbf structure - no rows\n")
		return nil, fmt.Errorf("invalid WELLS.dbf structure - no rows")
	}
	
	fmt.Printf("getWellTypes: Found %d columns: %v\n", len(columns), columns)
	fmt.Printf("getWellTypes: Found %d rows\n", len(rows))
	
	// Find relevant columns (adjust based on actual WELLS.DBF structure)
	var wellNameIdx, wellTypeIdx, statusIdx int = -1, -1, -1
	for i, col := range columns {
		colUpper := strings.ToUpper(col)
		switch colUpper {
		case "WELLNAME", "WELL_NAME", "NAME", "CWELLNAME":
			wellNameIdx = i
			fmt.Printf("getWellTypes: Found well name column at index %d: %s\n", i, col)
		case "WELLTYPE", "WELL_TYPE", "TYPE", "CWELLTYPE", "CGROUP", "CFORMATION":
			wellTypeIdx = i
			fmt.Printf("getWellTypes: Found well type column at index %d: %s\n", i, col)
		case "STATUS", "ACTIVE", "LSTATUS", "CSTATUS", "CWELLSTAT":
			statusIdx = i
			fmt.Printf("getWellTypes: Found status column at index %d: %s\n", i, col)
		}
	}
	
	fmt.Printf("getWellTypes: Column indices - Name: %d, Type: %d, Status: %d\n", wellNameIdx, wellTypeIdx, statusIdx)
	
	// Count wells by status
	statusCounts := make(map[string]int)
	totalWells := 0
	
	for _, row := range rows {
		if len(row) == 0 {
			continue
		}
		
		totalWells++
		
		// Get well status (A=Active, P=Plugged, S=Shutin, I=Inactive)
		wellStatus := "Unknown"
		if statusIdx != -1 && len(row) > statusIdx && row[statusIdx] != nil {
			wellStatus = strings.TrimSpace(strings.ToUpper(fmt.Sprintf("%v", row[statusIdx])))
		}
		
		// Count by status
		switch wellStatus {
		case "A":
			statusCounts["Active"]++
		case "P":
			statusCounts["Plugged"]++
		case "S":
			statusCounts["Shut-in"]++
		case "I":
			statusCounts["Inactive"]++
		default:
			statusCounts["Unknown"]++
		}
		
		// For well types, we'll show status-based groupings instead of formation/group
		// since status is more actionable for operations
	}
	
	// Convert status counts to array for frontend (showing well status instead of type)
	var wellTypes []map[string]interface{}
	for status, count := range statusCounts {
		if count > 0 { // Only show statuses that have wells
			wellTypes = append(wellTypes, map[string]interface{}{
				"type":  status,
				"count": count,
			})
		}
	}
	
	// Sort by count (descending)
	sort.Slice(wellTypes, func(i, j int) bool {
		return wellTypes[i]["count"].(int) > wellTypes[j]["count"].(int)
	})
	
	fmt.Printf("getWellTypes: Total wells: %d, Status breakdown: %v\n", totalWells, statusCounts)
	fmt.Printf("getWellTypes: Returning %d status categories: %v\n", len(wellTypes), wellTypes)
	return wellTypes, nil
}