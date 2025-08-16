package vendor

import (
	"fmt"
	"path/filepath"
	
	"github.com/Valentin-Kaiser/go-dbase/dbase"
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/logger"
)

// Service handles vendor-related operations
type Service struct {
	// Add any dependencies here if needed
}

// NewService creates a new vendor service
func NewService() *Service {
	return &Service{}
}

// GetVendors retrieves all vendors from VENDOR.dbf
func (s *Service) GetVendors(companyName string) (map[string]interface{}, error) {
	logger.WriteInfo("GetVendors", fmt.Sprintf("Called for company: %s", companyName))
	fmt.Printf("GetVendors: Called for company: %s\n", companyName)

	// Read the VENDOR.dbf file
	vendorData, err := company.ReadDBFFile(companyName, "VENDOR.dbf", "", 0, 0, "", "")
	if err != nil {
		logger.WriteError("GetVendors", fmt.Sprintf("Error reading VENDOR.dbf: %v", err))
		fmt.Printf("GetVendors: Error reading VENDOR.dbf: %v\n", err)
		return nil, fmt.Errorf("error reading vendor data: %v", err)
	}

	// Log the data structure
	if rows, ok := vendorData["rows"].([][]interface{}); ok {
		fmt.Printf("GetVendors: Found %d vendor records\n", len(rows))
		if len(rows) > 0 {
			fmt.Printf("GetVendors: First vendor record has %d fields\n", len(rows[0]))
			// Log column names
			if columns, ok := vendorData["columns"].([]string); ok {
				fmt.Printf("GetVendors: Columns: %v\n", columns)
			}
		}
	} else {
		fmt.Printf("GetVendors: No rows found in vendor data\n")
	}

	return vendorData, nil
}

// UpdateVendor updates a vendor record in VENDOR.dbf
func (s *Service) UpdateVendor(companyName string, vendorIndex int, vendorData map[string]interface{}) error {
	logger.WriteInfo("UpdateVendor", fmt.Sprintf("Updating vendor at index %d for company %s", vendorIndex, companyName))

	var vendorPath string

	// Check if companyName is already an absolute path (Windows)
	if filepath.IsAbs(companyName) {
		// It's already an absolute path, just append VENDOR.dbf
		vendorPath = filepath.Join(companyName, "VENDOR.dbf")
		logger.WriteInfo("UpdateVendor", fmt.Sprintf("Using absolute path: %s", vendorPath))
	} else {
		// It's a relative path, construct the full path
		datafilesPath, err := company.GetDatafilesPath()
		if err != nil {
			return fmt.Errorf("failed to get datafiles path: %w", err)
		}

		// Normalize the company path
		normalizedCompanyName := company.NormalizeCompanyPath(companyName)

		// Construct the full path to VENDOR.dbf
		vendorPath = filepath.Join(datafilesPath, normalizedCompanyName, "VENDOR.dbf")
		logger.WriteInfo("UpdateVendor", fmt.Sprintf("Using constructed path: %s", vendorPath))
	}

	// Open the table for writing
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   vendorPath,
		ReadOnly:   false,
		TrimSpaces: true,
	})
	if err != nil {
		return fmt.Errorf("failed to open VENDOR.dbf for writing: %w", err)
	}
	defer table.Close()

	// Navigate to the specific record
	currentIndex := 0
	var targetRow *dbase.Row

	for {
		row, err := table.Next()
		if err != nil {
			if err.Error() == "EOF" {
				return fmt.Errorf("vendor record not found at index %d", vendorIndex)
			}
			return fmt.Errorf("error reading vendor table: %w", err)
		}

		// Skip deleted records
		if row.Deleted {
			continue
		}

		if currentIndex == vendorIndex {
			targetRow = row
			break
		}
		currentIndex++
	}

	if targetRow == nil {
		return fmt.Errorf("vendor record not found at index %d", vendorIndex)
	}

	// Update the fields in the row
	for fieldName, value := range vendorData {
		field := targetRow.FieldByName(fieldName)
		if field == nil {
			// Skip fields that don't exist in the DBF
			continue
		}

		// Set the field value using the actual API
		err := field.SetValue(value)
		if err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldName, err)
		}
	}

	// Write the updated row back to the file
	err = table.WriteRow(targetRow)
	if err != nil {
		return fmt.Errorf("failed to write updated vendor record: %w", err)
	}

	logger.WriteInfo("UpdateVendor", fmt.Sprintf("Successfully updated vendor at index %d", vendorIndex))
	return nil
}