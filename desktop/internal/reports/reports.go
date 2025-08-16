// Package reports handles all report generation including PDFs, Excel, and data exports
package reports

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/debug"
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

// CheckOwnerStatementFiles checks if owner statement files exist for a company
func (s *Service) CheckOwnerStatementFiles(companyName string) map[string]interface{} {
	// Log the function call
	logger.WriteInfo("CheckOwnerStatementFiles", fmt.Sprintf("Called for company: %s", companyName))
	debug.SimpleLog(fmt.Sprintf("CheckOwnerStatementFiles: company=%s, platform=%s", companyName, runtime.GOOS))

	result := map[string]interface{}{
		"hasFiles": false,
		"files":    []string{},
		"error":    "",
	}

	// Build the path to the ownerstatements directory
	var ownerStatementsPath string

	// Use the same logic as ReadDBFFile to resolve the company path
	if filepath.IsAbs(companyName) {
		ownerStatementsPath = filepath.Join(companyName, "ownerstatements")
	} else {
		// For relative paths, we need to resolve relative to the working directory
		workingDir, _ := os.Getwd()
		ownerStatementsPath = filepath.Join(workingDir, "datafiles", companyName, "ownerstatements")
	}

	logger.WriteInfo("CheckOwnerStatementFiles", fmt.Sprintf("Checking directory: %s", ownerStatementsPath))

	// Check if the ownerstatements directory exists
	dirInfo, err := os.Stat(ownerStatementsPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.WriteInfo("CheckOwnerStatementFiles", "ownerstatements directory does not exist")
			result["error"] = "No Owner Distribution Files Found"
			return result
		}
		logger.WriteError("CheckOwnerStatementFiles", fmt.Sprintf("Error checking directory: %v", err))
		result["error"] = fmt.Sprintf("Error accessing directory: %v", err)
		return result
	}

	if !dirInfo.IsDir() {
		logger.WriteError("CheckOwnerStatementFiles", "ownerstatements exists but is not a directory")
		result["error"] = "ownerstatements is not a directory"
		return result
	}

	// Directory exists, now scan for DBF files
	logger.WriteInfo("CheckOwnerStatementFiles", "ownerstatements directory exists, scanning for DBF files")

	files, err := os.ReadDir(ownerStatementsPath)
	if err != nil {
		logger.WriteError("CheckOwnerStatementFiles", fmt.Sprintf("Error reading directory: %v", err))
		result["error"] = fmt.Sprintf("Error reading directory: %v", err)
		return result
	}

	var dbfFiles []string
	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			if strings.HasSuffix(strings.ToLower(fileName), ".dbf") {
				logger.WriteInfo("CheckOwnerStatementFiles", fmt.Sprintf("Found DBF file: %s", fileName))
				dbfFiles = append(dbfFiles, fileName)
			}
		}
	}

	if len(dbfFiles) > 0 {
		result["hasFiles"] = true
		result["files"] = dbfFiles
		logger.WriteInfo("CheckOwnerStatementFiles", fmt.Sprintf("Found %d DBF files in ownerstatements", len(dbfFiles)))
	} else {
		result["error"] = "No Owner Distribution Files Found"
		logger.WriteInfo("CheckOwnerStatementFiles", "No DBF files found in ownerstatements directory")
	}

	return result
}

// GetOwnerStatementsList returns a list of available owner statement files
func (s *Service) GetOwnerStatementsList(companyName string) ([]map[string]interface{}, error) {
	// Check if files exist first
	checkResult := s.CheckOwnerStatementFiles(companyName)
	if !checkResult["hasFiles"].(bool) {
		return nil, fmt.Errorf("no owner statement files found")
	}

	files := checkResult["files"].([]string)
	statements := make([]map[string]interface{}, 0)

	for _, fileName := range files {
		// For each file, get basic info
		statement := map[string]interface{}{
			"fileName": fileName,
			"name":     strings.TrimSuffix(fileName, filepath.Ext(fileName)),
		}
		statements = append(statements, statement)
	}

	return statements, nil
}

// GetOwnersList retrieves the list of owners from an owner statement DBF file
func (s *Service) GetOwnersList(companyName string, fileName string) ([]map[string]interface{}, error) {
	// Build the full path to the DBF file
	var dbfPath string
	if filepath.IsAbs(companyName) {
		dbfPath = filepath.Join(companyName, "ownerstatements", fileName)
	} else {
		workingDir, _ := os.Getwd()
		dbfPath = filepath.Join(workingDir, "datafiles", companyName, "ownerstatements", fileName)
	}

	// Open the DBF file
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   dbfPath,
		ReadOnly:   true,
		TrimSpaces: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open DBF file: %w", err)
	}
	defer table.Close()

	// Get field names
	columns := table.Columns()
	var ownerFieldIndex int = -1
	var ownerKeyIndex int = -1
	var ownerFieldName string
	var keyFieldName string
	
	for i, column := range columns {
		fieldName := strings.ToUpper(column.Name())
		if strings.Contains(fieldName, "OWNER") && !strings.Contains(fieldName, "KEY") {
			ownerFieldIndex = i
			ownerFieldName = column.Name()
		}
		if strings.Contains(fieldName, "KEY") || strings.Contains(fieldName, "ID") {
			ownerKeyIndex = i
			keyFieldName = column.Name()
		}
	}

	if ownerFieldIndex == -1 {
		return nil, fmt.Errorf("could not find owner name field in DBF")
	}

	// Read unique owners
	ownersMap := make(map[string]map[string]interface{})
	
	for !table.EOF() {
		record, err := table.Next()
		if err != nil {
			break
		}

		if record.Deleted {
			continue
		}

		ownerName := strings.TrimSpace(fmt.Sprintf("%v", record.Field(ownerFieldName)))
		if ownerName == "" {
			continue
		}

		// Use owner key if available, otherwise use name as key
		key := ownerName
		if ownerKeyIndex >= 0 && keyFieldName != "" {
			keyValue := strings.TrimSpace(fmt.Sprintf("%v", record.Field(keyFieldName)))
			if keyValue != "" {
				key = keyValue
			}
		}

		if _, exists := ownersMap[key]; !exists {
			ownersMap[key] = map[string]interface{}{
				"key":  key,
				"name": ownerName,
			}
		}
	}

	// Convert map to slice
	owners := make([]map[string]interface{}, 0, len(ownersMap))
	for _, owner := range ownersMap {
		owners = append(owners, owner)
	}

	return owners, nil
}

// GetOwnerStatementData retrieves owner statement data for a specific owner
func (s *Service) GetOwnerStatementData(companyName string, fileName string, ownerKey string) (map[string]interface{}, error) {
	// Build the full path to the DBF file
	var dbfPath string
	if filepath.IsAbs(companyName) {
		dbfPath = filepath.Join(companyName, "ownerstatements", fileName)
	} else {
		workingDir, _ := os.Getwd()
		dbfPath = filepath.Join(workingDir, "datafiles", companyName, "ownerstatements", fileName)
	}

	// Open the DBF file
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   dbfPath,
		ReadOnly:   true,
		TrimSpaces: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open DBF file: %w", err)
	}
	defer table.Close()

	// Get field names
	columns := table.Columns()
	fieldNames := make([]string, len(columns))
	for i, column := range columns {
		fieldNames[i] = column.Name()
	}

	// Find owner field
	var ownerFieldName string
	for _, name := range fieldNames {
		upperName := strings.ToUpper(name)
		if strings.Contains(upperName, "OWNER") || strings.Contains(upperName, "KEY") {
			ownerFieldName = name
			break
		}
	}

	// Collect records for this owner
	var records []map[string]interface{}
	for !table.EOF() {
		record, err := table.Next()
		if err != nil {
			break
		}

		if record.Deleted {
			continue
		}

		// Check if this record belongs to the requested owner
		ownerValue := strings.TrimSpace(fmt.Sprintf("%v", record.Field(ownerFieldName)))
		if ownerValue != ownerKey {
			continue
		}

		// Build record map
		recordMap := make(map[string]interface{})
		for _, fieldName := range fieldNames {
			recordMap[fieldName] = record.Field(fieldName)
		}
		records = append(records, recordMap)
	}

	return map[string]interface{}{
		"owner":   ownerKey,
		"records": records,
		"count":   len(records),
		"columns": fieldNames,
	}, nil
}

// ExamineOwnerStatementStructure examines the structure of owner statement DBF files
func (s *Service) ExamineOwnerStatementStructure(companyName string, fileName string) (map[string]interface{}, error) {
	logger.WriteInfo("ExamineOwnerStatementStructure", fmt.Sprintf("Examining %s for company %s", fileName, companyName))

	// Read the DBF file
	dbfData, err := company.ReadDBFFile(companyName, filepath.Join("ownerstatements", fileName), "", 0, 10, "", "")
	if err != nil {
		return nil, fmt.Errorf("error reading DBF file: %v", err)
	}

	// Get columns
	columns, _ := dbfData["columns"].([]string)

	// Get sample rows (first 10)
	var rows []map[string]interface{}
	if rowsData, ok := dbfData["rows"].([]map[string]interface{}); ok {
		rows = rowsData
	} else if rowsArray, ok := dbfData["rows"].([]interface{}); ok {
		for _, item := range rowsArray {
			if row, ok := item.(map[string]interface{}); ok {
				rows = append(rows, row)
			}
		}
	}

	// Analyze column types and sample values
	columnInfo := make([]map[string]interface{}, 0)
	for _, col := range columns {
		info := map[string]interface{}{
			"name":         col,
			"sampleValues": []interface{}{},
			"type":         "unknown",
		}

		// Get sample values from first few rows
		sampleValues := []interface{}{}
		for i, row := range rows {
			if i >= 3 { // Just get 3 samples
				break
			}
			if val, exists := row[col]; exists && val != nil {
				sampleValues = append(sampleValues, val)
				// Infer type from first non-nil value
				if info["type"] == "unknown" {
					switch val.(type) {
					case string:
						info["type"] = "string"
					case float64, float32, int, int64:
						info["type"] = "number"
					case bool:
						info["type"] = "boolean"
					case time.Time:
						info["type"] = "date"
					default:
						info["type"] = fmt.Sprintf("%T", val)
					}
				}
			}
		}
		info["sampleValues"] = sampleValues
		columnInfo = append(columnInfo, info)
	}

	result := map[string]interface{}{
		"fileName":      fileName,
		"recordCount":   len(rows),
		"columnCount":   len(columns),
		"columns":       columnInfo,
		"sampleRecords": rows,
	}

	return result, nil
}