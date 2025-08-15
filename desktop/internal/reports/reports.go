// Package reports handles all report generation including PDFs, Excel, and data exports
package reports

import (
	"fmt"
	"time"
)

// Service handles all reporting operations
type Service struct {
	// Add dependencies here as needed
}

// NewService creates a new reports service
func NewService() *Service {
	return &Service{}
}

// ReportType represents different types of reports
type ReportType string

const (
	ReportTypeChartOfAccounts   ReportType = "chart_of_accounts"
	ReportTypeOwnerStatements    ReportType = "owner_statements"
	ReportTypeBankReconciliation ReportType = "bank_reconciliation"
	ReportTypeVendorList         ReportType = "vendor_list"
	ReportTypeCheckRegister      ReportType = "check_register"
	ReportTypeGeneralLedger      ReportType = "general_ledger"
	ReportTypeFinancialSummary   ReportType = "financial_summary"
)

// ReportRequest represents a request to generate a report
type ReportRequest struct {
	Type        ReportType             `json:"type"`
	CompanyName string                 `json:"companyName"`
	StartDate   time.Time              `json:"startDate"`
	EndDate     time.Time              `json:"endDate"`
	Format      string                 `json:"format"` // "pdf", "excel", "csv"
	Filters     map[string]interface{} `json:"filters"`
}

// ReportMetadata contains information about a generated report
type ReportMetadata struct {
	ID          string    `json:"id"`
	Type        ReportType `json:"type"`
	Title       string    `json:"title"`
	GeneratedAt time.Time `json:"generatedAt"`
	GeneratedBy string    `json:"generatedBy"`
	RowCount    int       `json:"rowCount"`
	FilePath    string    `json:"filePath,omitempty"`
}

// Example function structure - to be implemented by moving from main.go
func (s *Service) GenerateChartOfAccountsPDF(companyName string, filters map[string]interface{}) (*ReportMetadata, error) {
	// TODO: Move implementation from main.go
	return nil, fmt.Errorf("not implemented yet - move from main.go")
}

// Example function structure - to be implemented by moving from main.go  
func (s *Service) GenerateOwnerStatements(companyName string, request ReportRequest) (*ReportMetadata, error) {
	// TODO: Move implementation from main.go
	return nil, fmt.Errorf("not implemented yet - move from main.go")
}