// Package reports handles all report generation including PDFs, Excel, and data exports
package reports

import (
	"fmt"
	"strings"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/logger"
	"github.com/pivoten/financialsx/desktop/internal/pdf"
)

// Service handles all reporting operations
type Service struct {
	pdfBuilder *pdf.ReportBuilder
}

// NewService creates a new reports service
func NewService() *Service {
	return &Service{
		pdfBuilder: pdf.NewReportBuilder(pdf.DefaultConfig()),
	}
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

// GenerateChartOfAccountsPDF generates a PDF report of the Chart of Accounts
func (s *Service) GenerateChartOfAccountsPDF(companyName string, sortBy string, includeInactive bool) ([]byte, error) {
	logger.WriteInfo("GenerateChartOfAccountsPDF", fmt.Sprintf("Called for company: %s, sortBy: %s, includeInactive: %v", companyName, sortBy, includeInactive))
	
	// Get the chart of accounts data from COA.dbf
	coaData, err := s.getChartOfAccountsData(companyName, sortBy, includeInactive)
	if err != nil {
		return nil, fmt.Errorf("failed to get chart of accounts: %v", err)
	}
	
	// Get company info from version.dbf
	companyInfo := s.getCompanyInfo(companyName)
	
	// Build the PDF using our PDF package
	pdfBytes, err := s.pdfBuilder.BuildChartOfAccountsReport(coaData, companyInfo, includeInactive)
	if err != nil {
		return nil, fmt.Errorf("failed to build PDF: %v", err)
	}
	
	return pdfBytes, nil
}

// GenerateOwnerStatementPDF generates a PDF for owner distribution statements
func (s *Service) GenerateOwnerStatementPDF(companyName string, fileName string) ([]byte, error) {
	logger.WriteInfo("GenerateOwnerStatementPDF", fmt.Sprintf("Called for company: %s, file: %s", companyName, fileName))
	
	// Read the DBF file from ownerstatements subdirectory
	dbfData, err := company.ReadDBFFile(companyName, "ownerstatements/"+fileName, "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("error reading DBF file: %v", err)
	}
	
	// Convert the DBF data to statement format
	statementData := s.convertDBFToStatementData(dbfData)
	
	// Get company info
	companyInfo := s.getCompanyInfo(companyName)
	
	// Build the PDF using our PDF package
	pdfBytes, err := s.pdfBuilder.BuildOwnerStatementReport(statementData, companyInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to build PDF: %v", err)
	}
	
	return pdfBytes, nil
}

// Helper functions for data retrieval and transformation

func (s *Service) getChartOfAccountsData(companyName string, sortBy string, includeInactive bool) ([]map[string]interface{}, error) {
	// Read COA.dbf
	coaData, err := company.ReadDBFFile(companyName, "COA.dbf", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("error reading COA.dbf: %v", err)
	}
	
	// Get columns and rows
	columns, _ := coaData["columns"].([]string)
	rows, _ := coaData["rows"].([][]interface{})
	
	var accounts []map[string]interface{}
	for _, row := range rows {
		account := make(map[string]interface{})
		for i, col := range columns {
			if i < len(row) {
				account[col] = row[i]
			}
		}
		
		// Filter inactive accounts if needed
		if !includeInactive {
			if inactive, ok := account["LINACTIVE"].(bool); ok && inactive {
				continue
			}
		}
		
		// Map to standard fields
		mappedAccount := map[string]interface{}{
			"account_number": account["CACCTNO"],
			"account_name":   account["CACCTDESC"],
			"account_type":   account["NACCTTYPE"],
			"parent_account": account["CPARENT"],
			"is_bank":        account["LBANKACCT"],
			"is_unit":        account["LACCTUNIT"],
			"is_dept":        account["LACCTDEPT"],
		}
		
		accounts = append(accounts, mappedAccount)
	}
	
	// Sort if requested
	if sortBy == "type" {
		// Sort by account type - implement sorting logic
	}
	
	return accounts, nil
}

func (s *Service) getCompanyInfo(companyName string) *pdf.CompanyInfo {
	info := &pdf.CompanyInfo{
		Name: companyName, // Default to folder name
	}
	
	// Try to read version.dbf for company details
	versionData, err := company.ReadDBFFile(companyName, "VERSION.DBF", "", 0, 1, "", "")
	if err != nil {
		logger.WriteInfo("getCompanyInfo", fmt.Sprintf("Could not read VERSION.DBF: %v", err))
		return info
	}
	
	if rows, ok := versionData["rows"].([][]interface{}); ok && len(rows) > 0 {
		columns, _ := versionData["columns"].([]string)
		record := make(map[string]interface{})
		for j, value := range rows[0] {
			if j < len(columns) {
				record[columns[j]] = value
			}
		}
		
		// Extract company information
		if val, ok := record["CPRODUCER"]; ok && val != nil {
			name := strings.TrimSpace(fmt.Sprintf("%v", val))
			if name != "" {
				info.Name = name
			}
		}
		
		if val, ok := record["CADDRESS1"]; ok && val != nil {
			info.Address1 = strings.TrimSpace(fmt.Sprintf("%v", val))
		}
		
		if val, ok := record["CCITY"]; ok && val != nil {
			info.City = strings.TrimSpace(fmt.Sprintf("%v", val))
		}
		
		if val, ok := record["CSTATE"]; ok && val != nil {
			info.State = strings.TrimSpace(fmt.Sprintf("%v", val))
		}
		
		if val, ok := record["CZIPCODE"]; ok && val != nil {
			info.Zip = strings.TrimSpace(fmt.Sprintf("%v", val))
		}
	}
	
	return info
}

func (s *Service) convertDBFToStatementData(dbfData map[string]interface{}) map[string]interface{} {
	// Convert DBF data to statement format
	statementData := make(map[string]interface{})
	
	// Get columns and rows
	columns, _ := dbfData["columns"].([]string)
	rows, _ := dbfData["rows"].([][]interface{})
	
	logger.WriteInfo("convertDBFToStatementData", fmt.Sprintf("Processing %d rows with %d columns", len(rows), len(columns)))
	
	// Convert rows to map format for easier processing
	var statements []map[string]interface{}
	for _, row := range rows {
		statement := make(map[string]interface{})
		for i, col := range columns {
			if i < len(row) {
				statement[col] = row[i]
			}
		}
		statements = append(statements, statement)
	}
	
	statementData["statements"] = statements
	statementData["columns"] = columns
	statementData["recordCount"] = len(statements)
	
	return statementData
}