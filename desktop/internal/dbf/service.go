package dbf

import (
	"fmt"

	"github.com/pivoten/financialsx/desktop/internal/company"
)

// Service handles all DBF file operations
type Service struct {
	// Could add configuration here if needed
}

// NewService creates a new DBF service instance
func NewService() *Service {
	return &Service{}
}

// GetFiles lists all DBF files in a company's data directory
func (s *Service) GetFiles(companyName string) ([]string, error) {
	// Use the existing GetDBFFiles function from company package
	return company.GetDBFFiles(companyName)
}

// GetTableData reads all data from a DBF file
func (s *Service) GetTableData(companyName, fileName string) (map[string]interface{}, error) {
	// Use the existing company.ReadDBFFile function for consistency
	return company.ReadDBFFile(companyName, fileName, "", 0, 0, "", "")
}

// GetTableDataPaged reads paginated data from a DBF file with optional sorting
func (s *Service) GetTableDataPaged(companyName, fileName string, offset, limit int, sortColumn, sortDirection string) (map[string]interface{}, error) {
	// Use the existing company.ReadDBFFile function with pagination
	return company.ReadDBFFile(companyName, fileName, "", offset, limit, sortColumn, sortDirection)
}

// SearchTable searches for records in a DBF file matching the search term
func (s *Service) SearchTable(companyName, fileName, searchTerm string) (map[string]interface{}, error) {
	// Use the existing company.ReadDBFFile function with search
	return company.ReadDBFFile(companyName, fileName, searchTerm, 0, 0, "", "")
}

// UpdateRecord updates a specific cell in a DBF file
func (s *Service) UpdateRecord(companyName, fileName string, rowIndex, colIndex int, value string) error {
	// Use the existing UpdateDBFRecord function from company package
	return company.UpdateDBFRecord(companyName, fileName, rowIndex, colIndex, value)
}

// GetTableInfo returns metadata about a DBF table
func (s *Service) GetTableInfo(companyName, fileName string) (map[string]interface{}, error) {
	// Get basic file data first
	data, err := company.ReadDBFFile(companyName, fileName, "", 0, 1, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read DBF file: %w", err)
	}

	// Build file info from available data
	info := map[string]interface{}{
		"fileName": fileName,
		"company":  companyName,
	}

	// Add columns information if available
	if columns, ok := data["columns"].([]map[string]interface{}); ok {
		info["fields"] = columns
		info["fieldCount"] = len(columns)
	}

	// Add record count if available
	if total, ok := data["total"].(int); ok {
		info["recordCount"] = total
	}

	return info, nil
}

// ValidateDBFFile checks if a DBF file exists and is readable
func (s *Service) ValidateDBFFile(companyName, fileName string) error {
	// Try to read the file with minimal data to validate it exists and is readable
	_, err := company.ReadDBFFile(companyName, fileName, "", 0, 1, "", "")
	if err != nil {
		return fmt.Errorf("DBF file validation failed for %s: %w", fileName, err)
	}
	return nil
}

// FileExists checks if a DBF file exists in the company directory
func (s *Service) FileExists(companyName, fileName string) bool {
	err := s.ValidateDBFFile(companyName, fileName)
	return err == nil
}