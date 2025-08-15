// Package legacy handles all Visual FoxPro and DBF-related operations
package legacy

import (
	"fmt"
	"path/filepath"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
	"github.com/pivoten/financialsx/desktop/internal/common"
)

// Service handles all legacy system operations
type Service struct {
	// Add dependencies here as needed
}

// NewService creates a new legacy service
func NewService() *Service {
	return &Service{}
}

// DBFTable represents metadata about a DBF table
type DBFTable struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	RecordCount int      `json:"recordCount"`
	Columns     []Column `json:"columns"`
	LastModified string  `json:"lastModified"`
}

// Column represents a DBF column definition
type Column struct {
	Name      string `json:"name"`
	Type      string `json:"type"`      // C=Character, N=Numeric, D=Date, L=Logical, M=Memo
	Length    int    `json:"length"`
	Decimals  int    `json:"decimals"`
	Nullable  bool   `json:"nullable"`
}

// VFPForm represents a Visual FoxPro form that can be launched
type VFPForm struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	FormFile    string `json:"formFile"` // .SCX file
	Icon        string `json:"icon"`
	RequiresCompany bool `json:"requiresCompany"`
}

// DBFReadOptions contains options for reading DBF files
type DBFReadOptions struct {
	Offset      int      `json:"offset"`
	Limit       int      `json:"limit"`
	SearchField string   `json:"searchField"`
	SearchValue string   `json:"searchValue"`
	SortField   string   `json:"sortField"`
	SortOrder   string   `json:"sortOrder"` // asc or desc
	Columns     []string `json:"columns"`   // specific columns to return
}

// DBFWriteOptions contains options for writing to DBF files
type DBFWriteOptions struct {
	CreateBackup bool `json:"createBackup"`
	ValidateData bool `json:"validateData"`
}

// ReadDBF reads a DBF file with the given options
func (s *Service) ReadDBF(companyPath, fileName string, options DBFReadOptions) (map[string]interface{}, error) {
	// TODO: Move DBF reading logic from company.go
	return nil, fmt.Errorf("not implemented yet - move from company.go")
}

// WriteDBF writes data to a DBF file
func (s *Service) WriteDBF(companyPath, fileName string, rowIndex int, data map[string]interface{}, options DBFWriteOptions) error {
	// TODO: Move DBF writing logic from main.go
	return fmt.Errorf("not implemented yet - move from main.go")
}

// GetDBFStructure returns the structure of a DBF file
func (s *Service) GetDBFStructure(companyPath, fileName string) (*DBFTable, error) {
	filePath := filepath.Join(companyPath, fileName)
	
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   filePath,
		ReadOnly:   true,
		TrimSpaces: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open DBF file: %w", err)
	}
	defer table.Close()

	// Get column information
	columns := make([]Column, 0)
	for _, col := range table.Columns() {
		columns = append(columns, Column{
			Name:     col.Name(),
			Type:     string(col.Type()),
			Length:   col.Length(),
			Decimals: col.Decimals(),
		})
	}

	return &DBFTable{
		Name:        fileName,
		Path:        filePath,
		RecordCount: table.RecordsCount(),
		Columns:     columns,
	}, nil
}

// LaunchVFPForm launches a Visual FoxPro form
func (s *Service) LaunchVFPForm(formID, companyName string) error {
	// TODO: Move VFP integration logic from vfp package
	return fmt.Errorf("not implemented yet - move from vfp package")
}

// GetAvailableForms returns all available VFP forms
func (s *Service) GetAvailableForms() ([]VFPForm, error) {
	// TODO: Return list of available VFP forms
	return nil, fmt.Errorf("not implemented yet")
}