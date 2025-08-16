package company

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
	"github.com/pivoten/financialsx/desktop/internal/debug"
)

// Service handles company management operations
type Service struct {
	isWindows bool
	platform  string
}

// NewService creates a new company service
func NewService() *Service {
	return &Service{
		isWindows: runtime.GOOS == "windows",
		platform:  runtime.GOOS,
	}
}

// ReadDBFFileFromPath reads a DBF file directly from a path
func ReadDBFFileFromPath(filePath string) (map[string]interface{}, error) {
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   filePath,
		TrimSpaces: true,
		ReadOnly:   true,
	})
	if err != nil {
		return nil, err
	}
	defer table.Close()
	
	// Get column names
	columns := table.Fields()
	columnNames := make([]string, len(columns))
	for i, field := range columns {
		columnNames[i] = field.Name()
	}
	
	// Read all rows
	var rows [][]interface{}
	for !table.EOF() {
		record, err := table.Next()
		if err != nil {
			break
		}
		
		if record.Deleted {
			continue
		}
		
		row := make([]interface{}, len(columns))
		for i, field := range columns {
			row[i] = record.Field(field.Name())
		}
		rows = append(rows, row)
	}
	
	return map[string]interface{}{
		"columns": columnNames,
		"rows":    rows,
		"total":   len(rows),
	}, nil
}

// findCompmastDBF recursively searches for compmast.dbf file
func findCompmastDBF(startPath string, maxDepth int) string {
	if maxDepth <= 0 {
		return ""
	}

	var result string
	filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		// Skip hidden directories and common non-data directories
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") ||
			info.Name() == "node_modules" ||
			info.Name() == "build" ||
			info.Name() == "dist") {
			return filepath.SkipDir
		}

		// Check if this is compmast.dbf (case-insensitive)
		if !info.IsDir() && strings.EqualFold(info.Name(), "compmast.dbf") {
			result = path
			return filepath.SkipDir // Stop walking once found
		}

		// Limit depth to prevent excessive searching
		relPath, _ := filepath.Rel(startPath, path)
		depth := len(strings.Split(relPath, string(filepath.Separator)))
		if depth > maxDepth {
			return filepath.SkipDir
		}

		return nil
	})

	return result
}

// GetCompanyList returns a list of all companies from compmast.dbf
func (s *Service) GetCompanyList() ([]map[string]interface{}, error) {
	fmt.Println("GetCompanyList: Searching for compmast.dbf...")
	debug.LogInfo("GetCompanyList", "Searching for compmast.dbf...")
	
	var compMastPath string
	var baseDir string
	
	if s.isWindows {
		// On Windows, always look for datafiles\compmast.dbf relative to EXE
		exePath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed to get executable path: %w", err)
		}
		
		exeDir := filepath.Dir(exePath)
		compMastPath = filepath.Join(exeDir, "..", "datafiles", "compmast.dbf")
		
		// Check if the file exists
		if _, err := os.Stat(compMastPath); os.IsNotExist(err) {
			// Try one more level up
			compMastPath = filepath.Join(exeDir, "..", "..", "datafiles", "compmast.dbf")
			if _, err := os.Stat(compMastPath); os.IsNotExist(err) {
				return nil, fmt.Errorf("compmast.dbf not found at expected location: %s", compMastPath)
			}
		}
		
		baseDir = filepath.Dir(compMastPath)
		fmt.Printf("GetCompanyList: Found compmast.dbf at: %s\n", compMastPath)
		fmt.Printf("GetCompanyList: Base directory: %s\n", baseDir)
	} else {
		// On Mac/Linux, use the existing search logic
		workDir, _ := os.Getwd()
		fmt.Printf("GetCompanyList: Starting search from working directory: %s\n", workDir)
		
		// First check if we have a saved data path
		savedPath := s.getSavedDataPath()
		if savedPath != "" {
			testPath := filepath.Join(savedPath, "compmast.dbf")
			if _, err := os.Stat(testPath); err == nil {
				compMastPath = testPath
				fmt.Printf("GetCompanyList: Using saved data path: %s\n", compMastPath)
			}
		}
		
		// If not found in saved path, search for it
		if compMastPath == "" {
			// Try from current working directory first
			compMastPath = findCompmastDBF(workDir, 5)
			
			if compMastPath == "" {
				// Try from parent directory
				parentDir := filepath.Dir(workDir)
				compMastPath = findCompmastDBF(parentDir, 3)
			}
			
			if compMastPath == "" {
				// Try from executable directory
				if exePath, err := os.Executable(); err == nil {
					exeDir := filepath.Dir(exePath)
					compMastPath = findCompmastDBF(exeDir, 3)
				}
			}
		}
		
		if compMastPath == "" {
			fmt.Println("GetCompanyList: compmast.dbf not found")
			debug.LogError("GetCompanyList", fmt.Errorf("compmast.dbf not found"))
			return nil, fmt.Errorf("compmast.dbf not found")
		}
		
		baseDir = filepath.Dir(compMastPath)
		fmt.Printf("GetCompanyList: Found compmast.dbf at: %s\n", compMastPath)
		fmt.Printf("GetCompanyList: Base directory: %s\n", baseDir)
	}
	
	// Read the DBF file using the existing company package function
	// We'll read directly from the file since we're in the company package
	dbfData, err := ReadDBFFileFromPath(compMastPath)
	if err != nil {
		fmt.Printf("GetCompanyList: Failed to open compmast.dbf: %v\n", err)
		debug.LogError("GetCompanyList", fmt.Errorf("failed to open compmast.dbf: %v", err))
		return nil, fmt.Errorf("failed to open compmast.dbf: %w", err)
	}
	
	// Get columns
	columns, ok := dbfData["columns"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid compmast.dbf structure")
	}
	
	fmt.Printf("GetCompanyList: Field names: %v\n", columns)
	
	// Find the indices of the fields we need
	var companyNameIdx, dataPathIdx int = -1, -1
	for i, name := range columns {
		upperName := strings.ToUpper(name)
		if upperName == "CCOMPNAME" {
			companyNameIdx = i
		} else if upperName == "CDATAPATH" {
			dataPathIdx = i
		}
	}
	
	if companyNameIdx == -1 || dataPathIdx == -1 {
		return nil, fmt.Errorf("required fields not found in compmast.dbf")
	}
	
	// Get rows
	rows, ok := dbfData["rows"].([][]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid rows in compmast.dbf")
	}
	
	// Read all records
	var companies []map[string]interface{}
	fmt.Printf("GetCompanyList: Processing %d records\n", len(rows))
	
	for _, row := range rows {
		if len(row) <= companyNameIdx || len(row) <= dataPathIdx {
			continue
		}
		
		companyName := strings.TrimSpace(fmt.Sprintf("%v", row[companyNameIdx]))
		dataPath := strings.TrimSpace(fmt.Sprintf("%v", row[dataPathIdx]))
		
		// Skip empty companies
		if companyName == "" {
			continue
		}
		
		// Normalize the data path based on platform
		normalizedPath := NormalizeCompanyPath(dataPath)
		
		// Resolve the full path based on platform
		var fullPath string
		if s.isWindows {
			// On Windows, use the path as-is if it's absolute, otherwise make it relative to base
			if filepath.IsAbs(normalizedPath) {
				fullPath = normalizedPath
			} else {
				fullPath = filepath.Join(baseDir, normalizedPath)
			}
		} else {
			// On Mac/Linux, paths are always relative to compmast.dbf location
			fullPath = filepath.Join(baseDir, normalizedPath)
		}
		
		// Clean the path
		fullPath = filepath.Clean(fullPath)
		
		// Check if the directory exists
		exists := false
		if stat, err := os.Stat(fullPath); err == nil && stat.IsDir() {
			exists = true
		}
		
		fmt.Printf("GetCompanyList: Company: %s, Path: %s, Exists: %v\n", companyName, fullPath, exists)
		
		companies = append(companies, map[string]interface{}{
			"name":       companyName,
			"path":       fullPath,
			"exists":     exists,
			"original":   dataPath,
			"normalized": normalizedPath,
		})
	}
	
	fmt.Printf("GetCompanyList: Found %d companies\n", len(companies))
	debug.LogInfo("GetCompanyList", fmt.Sprintf("Found %d companies", len(companies)))
	
	return companies, nil
}

// getSavedDataPath retrieves the saved data path from a config file
func (s *Service) getSavedDataPath() string {
	configPath := filepath.Join(os.TempDir(), "financialsx_datapath.txt")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// SaveDataPath saves the selected data folder path
func (s *Service) SaveDataPath(dataPath string) error {
	configPath := filepath.Join(os.TempDir(), "financialsx_datapath.txt")
	return os.WriteFile(configPath, []byte(dataPath), 0644)
}

// SelectDataFolder opens a folder selection dialog
func (s *Service) SelectDataFolder() (string, error) {
	// This would need to be implemented with platform-specific folder selection
	// For now, return an error indicating it needs to be implemented in main.go
	return "", fmt.Errorf("folder selection must be handled by main application")
}

// SetDataPath validates and saves a data path
func (s *Service) SetDataPath(dataPath string) error {
	// Validate that compmast.dbf exists in this location
	compMastPath := filepath.Join(dataPath, "compmast.dbf")
	if _, err := os.Stat(compMastPath); os.IsNotExist(err) {
		return fmt.Errorf("compmast.dbf not found in selected folder")
	}
	
	// Save the path
	return s.SaveDataPath(dataPath)
}