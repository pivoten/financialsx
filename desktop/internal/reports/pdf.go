package reports

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/logger"
)

// PDFGenerator handles PDF report generation
type PDFGenerator struct {
	CompanyName string
}

// NewPDFGenerator creates a new PDF generator instance
func NewPDFGenerator(companyName string) *PDFGenerator {
	return &PDFGenerator{
		CompanyName: companyName,
	}
}

// GetCompanyInfo retrieves company information from VERSION.DBF
func (g *PDFGenerator) GetCompanyInfo() (displayName, address, cityStateZip string) {
	displayName = g.CompanyName // Default to folder name
	address = ""
	cityStateZip = ""
	
	// Try to read version.dbf for company details
	versionData, err := company.ReadDBFFile(g.CompanyName, "VERSION.DBF", "", 0, 1, "", "")
	if err != nil {
		logger.WriteInfo("PDFGenerator.GetCompanyInfo", fmt.Sprintf("Could not read VERSION.DBF: %v", err))
		return
	}
	
	if versionData == nil {
		return
	}
	
	logger.WriteInfo("PDFGenerator.GetCompanyInfo", "VERSION.DBF read successfully")
	
	rows, ok := versionData["rows"].([][]interface{})
	if !ok || len(rows) == 0 {
		return
	}
	
	// Get column names
	columns, _ := versionData["columns"].([]string)
	logger.WriteInfo("PDFGenerator.GetCompanyInfo", fmt.Sprintf("VERSION.DBF columns: %v", columns))
	
	// Convert first row to map
	if len(rows[0]) == 0 {
		return
	}
	
	record := make(map[string]interface{})
	for j, value := range rows[0] {
		if j < len(columns) {
			record[columns[j]] = value
		}
	}
	
	// Extract company information - use CPRODUCER for company name
	if val, ok := record["CPRODUCER"]; ok && val != nil {
		name := strings.TrimSpace(fmt.Sprintf("%v", val))
		if name != "" {
			displayName = name
			logger.WriteInfo("PDFGenerator.GetCompanyInfo", fmt.Sprintf("Found company name from CPRODUCER: %s", name))
		}
	}
	
	// Try CADDRESS1 and CADDRESS2 for address
	if val, ok := record["CADDRESS1"]; ok && val != nil {
		addr := strings.TrimSpace(fmt.Sprintf("%v", val))
		if addr != "" {
			address = addr
		}
	}
	if address == "" {
		if val, ok := record["CADDRESS2"]; ok && val != nil {
			addr := strings.TrimSpace(fmt.Sprintf("%v", val))
			if addr != "" {
				address = addr
			}
		}
	}
	
	// Build city, state, zip line
	city := ""
	if val, ok := record["CCITY"]; ok && val != nil {
		city = strings.TrimSpace(fmt.Sprintf("%v", val))
	}
	
	state := ""
	if val, ok := record["CSTATE"]; ok && val != nil {
		state = strings.TrimSpace(fmt.Sprintf("%v", val))
	}
	
	zip := ""
	if val, ok := record["CZIPCODE"]; ok && val != nil {
		zip = strings.TrimSpace(fmt.Sprintf("%v", val))
	}
	
	// Format city, state zip
	if city != "" && state != "" && zip != "" {
		cityStateZip = city + ", " + state + " " + zip
	} else if city != "" && state != "" {
		cityStateZip = city + ", " + state
	} else if city != "" {
		cityStateZip = city
	} else if state != "" && zip != "" {
		cityStateZip = state + " " + zip
	} else if state != "" {
		cityStateZip = state
	} else if zip != "" {
		cityStateZip = zip
	}
	
	return
}

// GenerateChartOfAccountsPDF generates a PDF report of the Chart of Accounts
func (g *PDFGenerator) GenerateChartOfAccountsPDF(accounts []map[string]interface{}, sortBy string, includeInactive bool) (string, error) {
	logger.WriteInfo("GenerateChartOfAccountsPDF", fmt.Sprintf("Generating PDF for %d accounts", len(accounts)))
	
	// Get company info
	displayCompanyName, companyAddress, companyCityStateZip := g.GetCompanyInfo()
	
	// Create a new PDF document with landscape orientation for better table fit
	pdf := gofpdf.New("L", "mm", "Letter", "")
	pdf.SetAutoPageBreak(true, 20)
	
	// Add footer function BEFORE adding pages
	sortText := "Account Number"
	if sortBy == "type" {
		sortText = "Account Type"
	}
	
	filterText := "Active Only"
	if includeInactive {
		filterText = "All Accounts"
	}
	
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(128, 128, 128)
		// Footer with page number on left, filter/sort info on right
		pdf.CellFormat(0, 5, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 5, fmt.Sprintf("Sorted by: %s | Filter: %s", sortText, filterText), "", 0, "R", false, 0, "")
	})
	
	// Add the first page
	pdf.AddPage()
	
	// Set up colors
	pdf.SetFillColor(240, 240, 240) // Light gray for header
	pdf.SetTextColor(0, 0, 0)       // Black text
	pdf.SetDrawColor(200, 200, 200) // Gray borders
	
	// Title and Company Info
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Chart of Accounts", "", 1, "C", false, 0, "")
	
	// Company Name
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 7, displayCompanyName, "", 1, "C", false, 0, "")
	
	// Address if available
	pdf.SetFont("Arial", "", 10)
	if companyAddress != "" {
		pdf.CellFormat(0, 5, companyAddress, "", 1, "C", false, 0, "")
	}
	if companyCityStateZip != "" {
		pdf.CellFormat(0, 5, companyCityStateZip, "", 1, "C", false, 0, "")
	}
	
	// Generated date
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated: %s", time.Now().Format("January 2, 2006 at 3:04 PM")), "", 1, "C", false, 0, "")
	
	// Space before table
	pdf.Ln(5)
	
	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(240, 240, 240)
	
	// Calculate column widths for landscape letter (279mm - 20mm margins = 259mm usable)
	colWidths := []float64{30, 20, 140, 30, 39} // Total: 259mm
	headers := []string{"Account #", "Type", "Description", "Parent", "Status"}
	
	// Draw header row with borders
	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
	
	// Table body
	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)
	
	// Track alternating row colors
	rowColor := false
	
	for _, account := range accounts {
		// Check if we need a new page (leaving room for at least one row)
		if pdf.GetY() > 180 { // Letter height is 215.9mm, leaving margin
			pdf.AddPage()
			
			// Repeat header on new page
			pdf.SetFont("Arial", "B", 10)
			pdf.SetFillColor(240, 240, 240)
			for i, header := range headers {
				pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", true, 0, "")
			}
			pdf.Ln(-1)
			pdf.SetFont("Arial", "", 9)
			pdf.SetFillColor(255, 255, 255)
			rowColor = false
		}
		
		// Alternate row colors for better readability
		if rowColor {
			pdf.SetFillColor(250, 250, 250)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		rowColor = !rowColor
		
		// Extract account details with safe type assertions
		accountNumber := ""
		if val, ok := account["account_number"].(string); ok {
			accountNumber = val
		}
		
		accountType := ""
		if val, ok := account["account_type"].(string); ok {
			accountType = val
		}
		
		description := ""
		if val, ok := account["description"].(string); ok {
			description = val
		}
		
		parent := ""
		if val, ok := account["parent"].(string); ok {
			parent = val
		}
		
		status := "Active"
		if val, ok := account["is_active"].(bool); ok && !val {
			status = "Inactive"
			pdf.SetTextColor(150, 150, 150) // Gray text for inactive
		} else {
			pdf.SetTextColor(0, 0, 0)
		}
		
		// Draw row with borders
		pdf.CellFormat(colWidths[0], 7, accountNumber, "1", 0, "L", true, 0, "")
		pdf.CellFormat(colWidths[1], 7, accountType, "1", 0, "C", true, 0, "")
		
		// Truncate description if too long
		if len(description) > 60 {
			description = description[:57] + "..."
		}
		pdf.CellFormat(colWidths[2], 7, description, "1", 0, "L", true, 0, "")
		
		pdf.CellFormat(colWidths[3], 7, parent, "1", 0, "C", true, 0, "")
		pdf.CellFormat(colWidths[4], 7, status, "1", 0, "C", true, 0, "")
		pdf.Ln(-1)
	}
	
	// Summary
	pdf.Ln(5)
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(100, 100, 100)
	
	activeCount := 0
	inactiveCount := 0
	for _, account := range accounts {
		if isActive, ok := account["is_active"].(bool); ok && !isActive {
			inactiveCount++
		} else {
			activeCount++
		}
	}
	
	summaryText := fmt.Sprintf("Total Accounts: %d", len(accounts))
	if includeInactive {
		summaryText += fmt.Sprintf(" (Active: %d, Inactive: %d)", activeCount, inactiveCount)
	}
	pdf.CellFormat(0, 5, summaryText, "", 1, "L", false, 0, "")
	
	// Create a temporary file for the PDF
	tempDir := os.TempDir()
	
	// Sanitize company name for filename
	safeCompanyName := strings.ReplaceAll(displayCompanyName, "/", "-")
	safeCompanyName = strings.ReplaceAll(safeCompanyName, "\\", "-")
	safeCompanyName = strings.ReplaceAll(safeCompanyName, ":", "-")
	safeCompanyName = strings.ReplaceAll(safeCompanyName, "*", "-")
	safeCompanyName = strings.ReplaceAll(safeCompanyName, "?", "-")
	safeCompanyName = strings.ReplaceAll(safeCompanyName, "\"", "-")
	safeCompanyName = strings.ReplaceAll(safeCompanyName, "<", "-")
	safeCompanyName = strings.ReplaceAll(safeCompanyName, ">", "-")
	safeCompanyName = strings.ReplaceAll(safeCompanyName, "|", "-")
	
	fileName := fmt.Sprintf("%s - %s - Chart of Accounts.pdf",
		time.Now().Format("2006-01-02"),
		safeCompanyName)
	
	filePath := filepath.Join(tempDir, fileName)
	
	// Save the PDF
	err := pdf.OutputFileAndClose(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %v", err)
	}
	
	logger.WriteInfo("GenerateChartOfAccountsPDF", fmt.Sprintf("PDF generated successfully at: %s", filePath))
	return filePath, nil
}

// GenerateOwnerStatementPDF generates a PDF for owner distribution statements
func (g *PDFGenerator) GenerateOwnerStatementPDF(fileName string) (string, error) {
	logger.WriteInfo("GenerateOwnerStatementPDF", fmt.Sprintf("Generating PDF for file: %s", fileName))
	
	// Read the DBF file from ownerstatements subdirectory
	dbfData, err := company.ReadDBFFile(g.CompanyName, filepath.Join("ownerstatements", fileName), "", 0, 0, "", "")
	if err != nil {
		return "", fmt.Errorf("error reading DBF file: %v", err)
	}
	
	// Get columns to understand the structure
	columns, _ := dbfData["columns"].([]string)
	logger.WriteInfo("GenerateOwnerStatementPDF", fmt.Sprintf("DBF Columns: %v", columns))
	
	// Get the rows - they come as [][]interface{} from ReadDBFFile
	var rows []map[string]interface{}
	if rowsData, ok := dbfData["rows"].([]map[string]interface{}); ok {
		rows = rowsData
	} else if rowsArray, ok := dbfData["rows"].([][]interface{}); ok {
		// Convert [][]interface{} to []map[string]interface{}
		for _, rowValues := range rowsArray {
			rowMap := make(map[string]interface{})
			for i, value := range rowValues {
				if i < len(columns) {
					rowMap[columns[i]] = value
				}
			}
			rows = append(rows, rowMap)
		}
		logger.WriteInfo("GenerateOwnerStatementPDF", fmt.Sprintf("Converted %d rows from array format to map format", len(rows)))
	}
	
	logger.WriteInfo("GenerateOwnerStatementPDF", fmt.Sprintf("Found %d records in %s", len(rows), fileName))
	
	// Log first record to understand the data structure
	if len(rows) > 0 {
		logger.WriteInfo("GenerateOwnerStatementPDF", "First record sample:")
		for key, value := range rows[0] {
			valueStr := fmt.Sprintf("%v", value)
			if len(valueStr) > 100 {
				valueStr = valueStr[:100] + "..."
			}
			logger.WriteInfo("GenerateOwnerStatementPDF", fmt.Sprintf("  %s: %s", key, valueStr))
		}
	}
	
	// For now, return analysis info
	// TODO: Implement actual PDF generation based on the DBF structure
	result := fmt.Sprintf("DBF Analysis Complete:\n")
	result += fmt.Sprintf("- File: %s\n", fileName)
	result += fmt.Sprintf("- Records: %d\n", len(rows))
	result += fmt.Sprintf("- Columns: %d\n", len(columns))
	result += fmt.Sprintf("\nColumn Names:\n")
	for _, col := range columns {
		result += fmt.Sprintf("  - %s\n", col)
	}
	
	return result, nil
}