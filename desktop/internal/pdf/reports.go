package pdf

import (
	"fmt"
	"strings"
)

// ReportBuilder provides specialized methods for different report types
type ReportBuilder struct {
	gen *Generator
}

// NewReportBuilder creates a new report builder
func NewReportBuilder(config *Config) *ReportBuilder {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &ReportBuilder{
		gen: NewGenerator(config),
	}
}

// GetGenerator returns the underlying generator for custom operations
func (rb *ReportBuilder) GetGenerator() *Generator {
	return rb.gen
}

// BuildChartOfAccountsReport creates a Chart of Accounts PDF
func (rb *ReportBuilder) BuildChartOfAccountsReport(accounts []map[string]interface{}, companyInfo *CompanyInfo, includeInactive bool) ([]byte, error) {
	gen := rb.gen
	
	// Set company info
	gen.SetCompanyInfo(companyInfo)
	gen.SetReportTitle("Chart of Accounts")
	
	// Add first page
	gen.AddPage()
	
	// Add report header
	gen.AddTitle("Chart of Accounts", 16)
	if includeInactive {
		gen.AddSubtitle("Including Inactive Accounts")
	} else {
		gen.AddSubtitle("Active Accounts Only")
	}
	
	gen.AddSeparator()
	
	// Define table columns
	headers := []string{"Account Number", "Account Name", "Type", "Status"}
	widths := []float64{40, 80, 30, 30}
	
	// Prepare data rows
	var data [][]string
	for _, account := range accounts {
		accountNum := fmt.Sprintf("%v", account["account_number"])
		accountName := fmt.Sprintf("%v", account["account_name"])
		
		// Get account type
		accountType := "Unknown"
		if typeVal, ok := account["account_type"].(float64); ok {
			accountType = formatAccountType(int(typeVal))
		}
		
		// Get status
		status := "Active"
		if active, ok := account["is_active"].(bool); ok && !active {
			status = "Inactive"
		}
		
		row := []string{
			accountNum,
			TruncateText(accountName, 40),
			accountType,
			status,
		}
		data = append(data, row)
	}
	
	// Add table
	gen.AddTable(headers, data, widths)
	
	// Add summary
	gen.GetPDF().Ln(10)
	gen.GetPDF().SetFont("Arial", "B", 10)
	gen.GetPDF().Cell(0, 6, fmt.Sprintf("Total Accounts: %d", len(accounts)))
	
	// Generate PDF
	return gen.Output()
}

// BuildOwnerStatementReport creates an Owner Statement PDF
func (rb *ReportBuilder) BuildOwnerStatementReport(statementData map[string]interface{}, companyInfo *CompanyInfo) ([]byte, error) {
	gen := rb.gen
	
	// Set company info
	gen.SetCompanyInfo(companyInfo)
	gen.SetReportTitle("Owner Distribution Statement")
	
	// Add first page
	gen.AddPage()
	
	// Get owner info
	ownerName := "Unknown Owner"
	if name, ok := statementData["owner_name"].(string); ok {
		ownerName = name
	}
	
	// Add report header
	gen.AddTitle("Owner Distribution Statement", 16)
	gen.AddSubtitle(fmt.Sprintf("Owner: %s", ownerName))
	
	// Add period info if available
	if period, ok := statementData["period"].(string); ok {
		gen.GetPDF().Cell(0, 6, fmt.Sprintf("Period: %s", period))
		gen.GetPDF().Ln(8)
	}
	
	gen.AddSeparator()
	
	// Add revenue section
	if revenues, ok := statementData["revenues"].([]map[string]interface{}); ok {
		rb.addRevenueSection(revenues)
	}
	
	// Add expense section
	if expenses, ok := statementData["expenses"].([]map[string]interface{}); ok {
		rb.addExpenseSection(expenses)
	}
	
	// Add net distribution
	if netAmount, ok := statementData["net_distribution"].(float64); ok {
		gen.GetPDF().Ln(5)
		gen.GetPDF().SetFont("Arial", "B", 12)
		gen.GetPDF().Cell(100, 8, "Net Distribution:")
		gen.GetPDF().Cell(0, 8, FormatCurrency(netAmount))
		gen.GetPDF().Ln(10)
	}
	
	// Generate PDF
	return gen.Output()
}

// BuildBankReconciliationReport creates a Bank Reconciliation PDF
func (rb *ReportBuilder) BuildBankReconciliationReport(reconciliation map[string]interface{}, companyInfo *CompanyInfo) ([]byte, error) {
	gen := rb.gen
	
	// Set company info
	gen.SetCompanyInfo(companyInfo)
	gen.SetReportTitle("Bank Reconciliation Report")
	
	// Add first page
	gen.AddPage()
	
	// Add report header
	gen.AddTitle("Bank Reconciliation Report", 16)
	
	// Add account and date info
	if accountNum, ok := reconciliation["account_number"].(string); ok {
		gen.AddSubtitle(fmt.Sprintf("Account: %s", accountNum))
	}
	if date, ok := reconciliation["statement_date"].(string); ok {
		gen.GetPDF().Cell(0, 6, fmt.Sprintf("Statement Date: %s", FormatDate(date)))
		gen.GetPDF().Ln(8)
	}
	
	gen.AddSeparator()
	
	// Add reconciliation summary
	pdf := gen.GetPDF()
	pdf.SetFont("Arial", "", 10)
	
	// Beginning balance
	if balance, ok := reconciliation["beginning_balance"].(float64); ok {
		pdf.Cell(100, 6, "Beginning Balance:")
		pdf.Cell(0, 6, FormatCurrency(balance))
		pdf.Ln(6)
	}
	
	// Add deposits
	if deposits, ok := reconciliation["deposits"].(float64); ok {
		pdf.Cell(100, 6, "Plus: Deposits:")
		pdf.Cell(0, 6, FormatCurrency(deposits))
		pdf.Ln(6)
	}
	
	// Less checks
	if checks, ok := reconciliation["checks"].(float64); ok {
		pdf.Cell(100, 6, "Less: Checks:")
		pdf.Cell(0, 6, FormatCurrency(checks))
		pdf.Ln(6)
	}
	
	// Ending balance
	if balance, ok := reconciliation["ending_balance"].(float64); ok {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(100, 6, "Ending Balance:")
		pdf.Cell(0, 6, FormatCurrency(balance))
		pdf.Ln(10)
	}
	
	// Add outstanding checks if available
	if outstandingChecks, ok := reconciliation["outstanding_checks"].([]map[string]interface{}); ok && len(outstandingChecks) > 0 {
		gen.AddSeparator()
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "Outstanding Checks")
		pdf.Ln(8)
		
		headers := []string{"Check #", "Date", "Payee", "Amount"}
		widths := []float64{25, 25, 80, 35}
		
		var data [][]string
		for _, check := range outstandingChecks {
			checkNum := fmt.Sprintf("%v", check["check_number"])
			date := FormatDate(fmt.Sprintf("%v", check["date"]))
			payee := fmt.Sprintf("%v", check["payee"])
			amount := FormatCurrency(check["amount"].(float64))
			
			row := []string{checkNum, date, payee, amount}
			data = append(data, row)
		}
		
		gen.AddTable(headers, data, widths)
	}
	
	// Generate PDF
	return gen.Output()
}

// Helper methods

func (rb *ReportBuilder) addRevenueSection(revenues []map[string]interface{}) {
	gen := rb.gen
	pdf := gen.GetPDF()
	
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 7, "REVENUES")
	pdf.Ln(7)
	
	pdf.SetFont("Arial", "", 9)
	total := 0.0
	
	for _, item := range revenues {
		description := fmt.Sprintf("%v", item["description"])
		amount := 0.0
		if val, ok := item["amount"].(float64); ok {
			amount = val
			total += amount
		}
		
		pdf.Cell(120, 5, description)
		pdf.Cell(0, 5, FormatCurrency(amount))
		pdf.Ln(5)
	}
	
	// Total line
	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(120, 6, "Total Revenues:")
	pdf.Cell(0, 6, FormatCurrency(total))
	pdf.Ln(8)
}

func (rb *ReportBuilder) addExpenseSection(expenses []map[string]interface{}) {
	gen := rb.gen
	pdf := gen.GetPDF()
	
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 7, "EXPENSES")
	pdf.Ln(7)
	
	pdf.SetFont("Arial", "", 9)
	total := 0.0
	
	for _, item := range expenses {
		description := fmt.Sprintf("%v", item["description"])
		amount := 0.0
		if val, ok := item["amount"].(float64); ok {
			amount = val
			total += amount
		}
		
		pdf.Cell(120, 5, description)
		pdf.Cell(0, 5, FormatCurrency(amount))
		pdf.Ln(5)
	}
	
	// Total line
	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(120, 6, "Total Expenses:")
	pdf.Cell(0, 6, FormatCurrency(total))
	pdf.Ln(8)
}

func formatAccountType(typeCode int) string {
	// Map account type codes to descriptions
	typeMap := map[int]string{
		1: "Asset",
		2: "Liability",
		3: "Equity",
		4: "Revenue",
		5: "Expense",
		6: "Other",
	}
	
	if desc, ok := typeMap[typeCode]; ok {
		return desc
	}
	return fmt.Sprintf("Type %d", typeCode)
}

// CreateFinancialReport creates a generic financial report
func (rb *ReportBuilder) CreateFinancialReport(title string, sections []ReportSection, companyInfo *CompanyInfo) ([]byte, error) {
	gen := rb.gen
	
	// Set company info
	gen.SetCompanyInfo(companyInfo)
	gen.SetReportTitle(title)
	
	// Add first page
	gen.AddPage()
	
	// Add title
	gen.AddTitle(title, 16)
	gen.AddSeparator()
	
	// Add sections
	for _, section := range sections {
		rb.addReportSection(section)
	}
	
	// Generate PDF
	return gen.Output()
}

// ReportSection represents a section in a report
type ReportSection struct {
	Title    string
	Subtitle string
	Type     string // "table", "list", "summary"
	Headers  []string
	Data     [][]string
	Widths   []float64
}

func (rb *ReportBuilder) addReportSection(section ReportSection) {
	gen := rb.gen
	pdf := gen.GetPDF()
	
	// Section title
	if section.Title != "" {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 7, strings.ToUpper(section.Title))
		pdf.Ln(7)
	}
	
	// Section subtitle
	if section.Subtitle != "" {
		pdf.SetFont("Arial", "", 9)
		pdf.Cell(0, 5, section.Subtitle)
		pdf.Ln(5)
	}
	
	// Section content based on type
	switch section.Type {
	case "table":
		if len(section.Headers) > 0 && len(section.Data) > 0 {
			gen.AddTable(section.Headers, section.Data, section.Widths)
		}
	case "list":
		pdf.SetFont("Arial", "", 9)
		for _, row := range section.Data {
			if len(row) > 0 {
				pdf.Cell(0, 5, row[0])
				pdf.Ln(5)
			}
		}
	case "summary":
		pdf.SetFont("Arial", "", 10)
		for _, row := range section.Data {
			if len(row) >= 2 {
				pdf.Cell(100, 6, row[0])
				pdf.Cell(0, 6, row[1])
				pdf.Ln(6)
			}
		}
	}
	
	pdf.Ln(5)
}