// Package operations handles business operations including vendors, purchases, inventory, and workflow
package operations

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/pivoten/financialsx/desktop/internal/company"
)

// Service handles all business operations
type Service struct {
	// Add dependencies here as needed
}

// NewService creates a new operations service
func NewService() *Service {
	return &Service{}
}

// Vendor represents a vendor/supplier
type Vendor struct {
	VendorID      string    `json:"vendorId"`
	VendorName    string    `json:"vendorName"`
	Address1      string    `json:"address1"`
	Address2      string    `json:"address2"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	Zip           string    `json:"zip"`
	Country       string    `json:"country"`
	TaxID         string    `json:"taxId"`
	Phone         string    `json:"phone"`
	Contact       string    `json:"contact"`
	Email         string    `json:"email"`
	Terms         string    `json:"terms"`
	IsActive      bool      `json:"isActive"`
	Is1099        bool      `json:"is1099"`
	CreditLimit   float64   `json:"creditLimit"`
	CurrentBalance float64  `json:"currentBalance"`
	LastPurchase  *time.Time `json:"lastPurchase,omitempty"`
	LastPayment   *time.Time `json:"lastPayment,omitempty"`
}

// Purchase represents a purchase/bill
type Purchase struct {
	PurchaseID    string    `json:"purchaseId"`
	VendorID      string    `json:"vendorId"`
	InvoiceNumber string    `json:"invoiceNumber"`
	InvoiceDate   time.Time `json:"invoiceDate"`
	DueDate       time.Time `json:"dueDate"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description"`
	Status        string    `json:"status"` // open, paid, partial, void
	PaidAmount    float64   `json:"paidAmount"`
	BatchNumber   string    `json:"batchNumber"`
}

// Well represents an oil/gas well
type Well struct {
	WellID       string  `json:"wellId"`
	WellName     string  `json:"wellName"`
	APINumber    string  `json:"apiNumber"`
	County       string  `json:"county"`
	State        string  `json:"state"`
	Field        string  `json:"field"`
	Operator     string  `json:"operator"`
	WorkingInterest float64 `json:"workingInterest"`
	NetRevenue   float64 `json:"netRevenue"`
	Status       string  `json:"status"` // active, inactive, plugged
	SpudDate     *time.Time `json:"spudDate,omitempty"`
	FirstProdDate *time.Time `json:"firstProdDate,omitempty"`
}

// Owner represents a working interest owner
type Owner struct {
	OwnerID        string  `json:"ownerId"`
	OwnerName      string  `json:"ownerName"`
	Address        string  `json:"address"`
	City           string  `json:"city"`
	State          string  `json:"state"`
	Zip            string  `json:"zip"`
	TaxID          string  `json:"taxId"`
	OwnershipType  string  `json:"ownershipType"` // WI, RI, ORRI
	InterestPercent float64 `json:"interestPercent"`
	IsActive       bool    `json:"isActive"`
}

// BatchResult contains the results of following a batch number through the system
type BatchResult struct {
	BatchNumber string                 `json:"batch_number"`
	CompanyName string                 `json:"company_name"`
	Checks      TableSearchResult      `json:"checks"`
	GLMaster    TableSearchResult      `json:"glmaster"`
	APPmtHdr    TableSearchResult      `json:"appmthdr"`
	APPmtDet    TableSearchResult      `json:"appmtdet"`
	APPurchH    TableSearchResult      `json:"appurchh"`
	APPurchD    TableSearchResult      `json:"appurchd"`
}

// TableSearchResult contains search results for a specific table
type TableSearchResult struct {
	TableName string                   `json:"table_name"`
	Records   []map[string]interface{} `json:"records"`
	Count     int                      `json:"count"`
	Columns   []string                 `json:"columns"`
	Error     string                   `json:"error,omitempty"`
}

// FollowBatchNumber traces a batch number through all related tables
func (s *Service) FollowBatchNumber(companyName string, batchNumber string) (*BatchResult, error) {
	// Trim and validate batch number
	batchNumber = strings.TrimSpace(batchNumber)
	if batchNumber == "" {
		return nil, fmt.Errorf("batch number cannot be empty")
	}
	
	fmt.Printf("FollowBatchNumber: Searching for batch '%s' in company '%s'\n", batchNumber, companyName)
	
	result := &BatchResult{
		BatchNumber: batchNumber,
		CompanyName: companyName,
		Checks:      s.searchTableForBatch(companyName, "checks.dbf", batchNumber, "CBATCH"),
		GLMaster:    s.searchTableForBatch(companyName, "glmaster.dbf", batchNumber, "CBATCH"),
		APPmtHdr:    s.searchTableForBatch(companyName, "appmthdr.dbf", batchNumber, "CBATCH"),
		APPmtDet:    s.searchTableForBatch(companyName, "appmtdet.dbf", batchNumber, "CBATCH"),
	}
	
	// If we found records in APPMTDET, look for CBILLTOKEN to find purchase records
	if result.APPmtDet.Count > 0 && len(result.APPmtDet.Records) > 0 {
		// Extract CBILLTOKEN from the first APPMTDET record
		if firstRecord := result.APPmtDet.Records[0]; firstRecord != nil {
			for colName, value := range firstRecord {
				if strings.ToUpper(colName) == "CBILLTOKEN" && value != nil {
					billToken := fmt.Sprintf("%v", value)
					billToken = strings.TrimSpace(billToken)
					if billToken != "" && billToken != "0" {
						fmt.Printf("FollowBatchNumber: Found CBILLTOKEN '%s', searching purchase tables\n", billToken)
						// Search purchase tables using CBILLTOKEN as the batch
						result.APPurchH = s.searchTableForBatch(companyName, "appurchh.dbf", billToken, "CBATCH")
						result.APPurchD = s.searchTableForBatch(companyName, "appurchd.dbf", billToken, "CBATCH")
					}
					break
				}
			}
		}
	} else {
		// No payment details found, initialize empty purchase results
		result.APPurchH = TableSearchResult{
			TableName: "APPURCHH.DBF",
			Records:   []map[string]interface{}{},
			Count:     0,
			Columns:   []string{},
		}
		result.APPurchD = TableSearchResult{
			TableName: "APPURCHD.DBF",
			Records:   []map[string]interface{}{},
			Count:     0,
			Columns:   []string{},
		}
	}
	
	// Also search GL for the original purchase entry if we have CBILLTOKEN
	if result.APPurchH.Count > 0 {
		// Search GLMASTER for CSOURCE = 'AP' with the same batch to find purchase GL entry
		glPurchaseData, err := company.ReadDBFFile(companyName, "glmaster.dbf", "", 0, 0, "", "")
		if err == nil {
			var purchaseGLRecords []map[string]interface{}
			if rows, ok := glPurchaseData["rows"].([][]interface{}); ok {
				cols := glPurchaseData["columns"].([]string)
				for _, row := range rows {
					record := make(map[string]interface{})
					for i, col := range cols {
						if i < len(row) {
							record[col] = row[i]
						}
					}
					// Check if this is an AP source entry
					if source, ok := record["CSOURCE"]; ok && fmt.Sprintf("%v", source) == "AP" {
						// Check if batch matches the CBILLTOKEN
						if batch, ok := record["CBATCH"]; ok {
							batchStr := strings.TrimSpace(fmt.Sprintf("%v", batch))
							// Check against original batch or CBILLTOKEN
							if batchStr == batchNumber {
								purchaseGLRecords = append(purchaseGLRecords, record)
							}
						}
					}
				}
			}
			// Add purchase GL records to main GL results
			if len(purchaseGLRecords) > 0 {
				result.GLMaster.Records = append(result.GLMaster.Records, purchaseGLRecords...)
				result.GLMaster.Count = len(result.GLMaster.Records)
			}
		}
	}
	
	fmt.Printf("FollowBatchNumber: Search complete. Found %d total records\n", 
		result.Checks.Count + result.GLMaster.Count + result.APPmtHdr.Count + 
		result.APPmtDet.Count + result.APPurchH.Count + result.APPurchD.Count)
	
	return result, nil
}

// searchTableForBatch searches a specific table for records matching a batch number
func (s *Service) searchTableForBatch(companyName, tableName, batchNumber, searchField string) TableSearchResult {
	result := TableSearchResult{
		TableName: strings.ToUpper(tableName),
		Records:   []map[string]interface{}{},
		Count:     0,
		Columns:   []string{},
	}
	
	fmt.Printf("FollowBatchNumber: Searching %s for %s='%s'\n", tableName, searchField, batchNumber)
	
	// Read the entire table
	data, err := company.ReadDBFFile(companyName, tableName, "", 0, 0, "", "")
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read %s: %v", tableName, err)
		return result
	}
	
	// Get columns
	if cols, ok := data["columns"].([]string); ok {
		result.Columns = cols
	}
	
	// Filter rows by batch number
	if rows, ok := data["rows"].([][]interface{}); ok {
		searchFieldIndex := -1
		for i, col := range result.Columns {
			if strings.EqualFold(col, searchField) {
				searchFieldIndex = i
				break
			}
		}
		
		if searchFieldIndex >= 0 {
			for _, row := range rows {
				if searchFieldIndex < len(row) && row[searchFieldIndex] != nil {
					fieldValue := strings.TrimSpace(fmt.Sprintf("%v", row[searchFieldIndex]))
					if strings.EqualFold(fieldValue, batchNumber) {
						// Convert row to map
						record := make(map[string]interface{})
						for i, col := range result.Columns {
							if i < len(row) {
								record[col] = row[i]
							}
						}
						result.Records = append(result.Records, record)
					}
				}
			}
		}
	}
	
	result.Count = len(result.Records)
	fmt.Printf("FollowBatchNumber: Found %d records in %s\n", result.Count, tableName)
	
	return result
}

// UpdateBatchFields updates fields across multiple tables for a given batch
func (s *Service) UpdateBatchFields(companyName, batchNumber string, fieldMappings map[string]string, 
	newValue string, tablesToUpdate map[string]bool) (map[string]interface{}, error) {
	
	fmt.Printf("UpdateBatchFields: Updating fields for batch '%s' with new value '%s'\n", 
		batchNumber, newValue)
	
	result := map[string]interface{}{
		"batch_number":   batchNumber,
		"field_mappings": fieldMappings,
		"new_value":      newValue,
		"updates":        map[string]interface{}{},
		"errors":         []string{},
		"total_updated":  0,
	}
	
	// Implementation would go here - this is a complex operation that updates DBF files
	// For now, return not implemented
	return result, fmt.Errorf("UpdateBatchFields not fully implemented - DBF write operations needed")
}

// RunNetDistribution runs the net revenue distribution process
func (s *Service) RunNetDistribution(periodStart, periodEnd string, processType string, recalculateAll bool, userID int, companyName string) (map[string]interface{}, error) {
	// TODO: Uncomment when ready to use the distribution processor
	/*
		logger := log.New(log.Writer(), fmt.Sprintf("[NETDIST-%s] ", companyName), log.LstdFlags)
		netDistProcess := processes.NewDistributionProcessor(a.db, companyName, logger)
		periodStartDate, err := time.Parse("2006-01-02", periodStart)
		if err != nil {
			return nil, fmt.Errorf("invalid period start date: %w", err)
		}
		periodEndDate, err := time.Parse("2006-01-02", periodEnd)
		if err != nil {
			return nil, fmt.Errorf("invalid period end date: %w", err)
		}
		config := &processes.ProcessingConfig{
			Period:      fmt.Sprintf("%02d", periodStartDate.Month()),
			Year:        fmt.Sprintf("%04d", periodStartDate.Year()),
			AcctDate:    periodEndDate,
			RevDate:     periodEndDate,
			ExpDate:     periodEndDate,
			UserID:      userID,
			IsNewRun:    true,
			IsClosing:   false,
		}
	*/
	// For now, return a placeholder response while we're developing
	// TODO: Uncomment below when ready to hook up the full distribution processor
	/*
		options := &processes.ProcessingOptions{
			RevSummarize: true,
			ExpSummarize: true,
			GLSummary:    true,
		}
		if err := netDistProcess.Initialize(config, options); err != nil {
			return nil, fmt.Errorf("failed to initialize distribution processor: %w", err)
		}
		result, err := netDistProcess.Main()
		if err != nil {
			return nil, fmt.Errorf("net distribution process failed: %w", err)
		}
		// Convert result to map for JSON serialization
		return map[string]interface{}{
			"run_number":       result.RunNumber,
			"run_year":         result.RunYear,
			"status":          result.Status,
			"wells_processed": result.WellsProcessed,
			"owners_processed": result.OwnersProcessed,
			"records_created":  result.RecordsCreated,
			"total_revenue":    result.TotalRevenue.String(),
			"total_expenses":   result.TotalExpenses.String(),
			"net_distributed":  result.NetDistributed.String(),
			"warnings":        result.Warnings,
			"errors":          result.Errors,
			"duration":        result.Duration.String(),
			"start_time":      result.StartTime,
			"end_time":        result.EndTime,
		}, nil
	*/
	// Placeholder response for development
	return map[string]interface{}{
		"status":   "development",
		"message":  "Distribution processor in development - not yet connected",
		"duration": "0s",
	}, nil
}

// Example function structures - to be implemented by moving from main.go
func (s *Service) GetVendors(companyName string) ([]Vendor, error) {
	// TODO: Move implementation from main.go
	return nil, fmt.Errorf("not implemented yet - move from main.go")
}

func (s *Service) UpdateVendor(companyName string, vendorID string, vendor Vendor) error {
	// TODO: Move implementation from main.go
	return fmt.Errorf("not implemented yet - move from main.go")
}

func (s *Service) GetWells(companyName string) ([]Well, error) {
	// TODO: Move implementation from main.go
	return nil, fmt.Errorf("not implemented yet - move from main.go")
}

func (s *Service) GetOwners(companyName string) ([]Owner, error) {
	// TODO: Move implementation from main.go
	return nil, fmt.Errorf("not implemented yet - move from main.go")
}