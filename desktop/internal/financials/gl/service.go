package gl

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
	
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/currency"
)

// Service handles general ledger operations
type Service struct {
	db *sql.DB
}

// NewService creates a new GL service
func NewService(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
}

// GLEntry represents a general ledger entry
type GLEntry struct {
	AccountNumber   string    `json:"account_number"`
	TransactionDate time.Time `json:"transaction_date"`
	Description     string    `json:"description"`
	Amount          float64   `json:"amount"`
	DebitAmount     float64   `json:"debit_amount"`
	CreditAmount    float64   `json:"credit_amount"`
	Source          string    `json:"source"`
	Reference       string    `json:"reference"`
	Period          string    `json:"period"`
	BatchNumber     string    `json:"batch_number"`
	RowIndex        int       `json:"row_index"`
}

// BalanceAnalysis represents GL balance analysis results
type BalanceAnalysis struct {
	AccountNumber   string                 `json:"account_number"`
	AccountName     string                 `json:"account_name"`
	TotalDebits     float64                `json:"total_debits"`
	TotalCredits    float64                `json:"total_credits"`
	NetBalance      float64                `json:"net_balance"`
	PeriodBalances  map[string]float64     `json:"period_balances"`
	YearlyBalances  map[int]float64        `json:"yearly_balances"`
	TransactionCount int                   `json:"transaction_count"`
	FirstEntry      time.Time              `json:"first_entry"`
	LastEntry       time.Time              `json:"last_entry"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// PeriodInfo represents GL period information
type PeriodInfo struct {
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	PeriodCode      string    `json:"period_code"`
	IsClosed        bool      `json:"is_closed"`
	ClosedDate      time.Time `json:"closed_date"`
	TransactionCount int      `json:"transaction_count"`
}

// ClosingResult represents the result of a period closing
type ClosingResult struct {
	PeriodEnd       string    `json:"period_end"`
	Status          string    `json:"status"`
	EntriesCreated  int       `json:"entries_created"`
	AccountsAffected int      `json:"accounts_affected"`
	TotalDebits     float64   `json:"total_debits"`
	TotalCredits    float64   `json:"total_credits"`
	Warnings        []string  `json:"warnings"`
	Errors          []string  `json:"errors"`
}

// AnalyzeGLBalancesByYear analyzes GL balances grouped by year and account
func (s *Service) AnalyzeGLBalancesByYear(companyName, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("AnalyzeGLBalancesByYear: Analyzing GLMASTER.dbf for account %s\n", accountNumber)
	
	// Read GLMASTER.dbf
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.dbf", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read GLMASTER.dbf: %w", err)
	}
	
	// Get column indices
	glColumns, ok := glData["columns"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid GLMASTER.dbf structure")
	}
	
	var yearIdx, periodIdx, accountIdx, debitIdx, creditIdx int = -1, -1, -1, -1, -1
	for i, col := range glColumns {
		colUpper := strings.ToUpper(col)
		switch colUpper {
		case "CYEAR":
			yearIdx = i
		case "CPERIOD":
			periodIdx = i
		case "CACCTNO":
			accountIdx = i
		case "NDEBITS", "NDEBIT":
			debitIdx = i
		case "NCREDITS", "NCREDIT":
			creditIdx = i
		}
	}
	
	if accountIdx == -1 {
		return nil, fmt.Errorf("account column not found")
	}
	
	// Structure to hold year-based totals
	type YearTotals struct {
		Debits  currency.Currency
		Credits currency.Currency
		Count   int
		Periods map[string]int
	}
	
	// Maps to store results
	yearlyTotals := make(map[string]*YearTotals)
	blankYearTotals := &YearTotals{Debits: currency.Zero(), Credits: currency.Zero(), Periods: make(map[string]int)}
	allAccountsTotals := make(map[string]*YearTotals) // For comparison
	
	// Process all rows
	glRows, _ := glData["rows"].([][]interface{})
	
	for _, row := range glRows {
		if len(row) <= accountIdx {
			continue
		}
		
		rowAccount := strings.TrimSpace(fmt.Sprintf("%v", row[accountIdx]))
		
		// Get year and period
		yearVal := ""
		periodVal := ""
		if yearIdx >= 0 && len(row) > yearIdx {
			yearVal = strings.TrimSpace(fmt.Sprintf("%v", row[yearIdx]))
		}
		if periodIdx >= 0 && len(row) > periodIdx {
			periodVal = strings.TrimSpace(fmt.Sprintf("%v", row[periodIdx]))
		}
		
		// Get amounts using decimal arithmetic
		debitVal := currency.Zero()
		if debitIdx >= 0 && len(row) > debitIdx && row[debitIdx] != nil {
			debitVal = currency.ParseFromDBF(row[debitIdx])
		}
		creditVal := currency.Zero()
		if creditIdx >= 0 && len(row) > creditIdx && row[creditIdx] != nil {
			creditVal = currency.ParseFromDBF(row[creditIdx])
		}
		
		// Process for all accounts (for comparison)
		if yearVal != "" && yearVal != "<nil>" {
			if allAccountsTotals[yearVal] == nil {
				allAccountsTotals[yearVal] = &YearTotals{Debits: currency.Zero(), Credits: currency.Zero(), Periods: make(map[string]int)}
			}
			allAccountsTotals[yearVal].Debits = allAccountsTotals[yearVal].Debits.Add(debitVal)
			allAccountsTotals[yearVal].Credits = allAccountsTotals[yearVal].Credits.Add(creditVal)
			allAccountsTotals[yearVal].Count++
		}
		
		// Process for specific account if provided
		if accountNumber == "" || rowAccount == accountNumber {
			if yearVal == "" || yearVal == "<nil>" {
				// Blank year entries
				blankYearTotals.Debits = blankYearTotals.Debits.Add(debitVal)
				blankYearTotals.Credits = blankYearTotals.Credits.Add(creditVal)
				blankYearTotals.Count++
				if periodVal != "" && periodVal != "<nil>" {
					blankYearTotals.Periods[periodVal]++
				}
			} else {
				// Normal year entries
				if yearlyTotals[yearVal] == nil {
					yearlyTotals[yearVal] = &YearTotals{Debits: currency.Zero(), Credits: currency.Zero(), Periods: make(map[string]int)}
				}
				yearlyTotals[yearVal].Debits = yearlyTotals[yearVal].Debits.Add(debitVal)
				yearlyTotals[yearVal].Credits = yearlyTotals[yearVal].Credits.Add(creditVal)
				yearlyTotals[yearVal].Count++
				if periodVal != "" && periodVal != "<nil>" {
					yearlyTotals[yearVal].Periods[periodVal]++
				}
			}
		}
	}
	
	// Convert to output format
	yearlyResults := make([]map[string]interface{}, 0)
	totalDebits := currency.Zero()
	totalCredits := currency.Zero()
	var totalRecords int
	
	// Sort years
	years := make([]string, 0, len(yearlyTotals))
	for year := range yearlyTotals {
		years = append(years, year)
	}
	sort.Strings(years)
	
	for _, year := range years {
		totals := yearlyTotals[year]
		balance := totals.Debits.Sub(totals.Credits)
		
		yearlyResults = append(yearlyResults, map[string]interface{}{
			"year":         year,
			"debits":       totals.Debits.ToFloat64(),
			"credits":      totals.Credits.ToFloat64(),
			"balance":      balance.ToFloat64(),
			"record_count": totals.Count,
			"periods":      len(totals.Periods),
		})
		
		totalDebits = totalDebits.Add(totals.Debits)
		totalCredits = totalCredits.Add(totals.Credits)
		totalRecords += totals.Count
	}
	
	// Add blank year totals if any
	var blankYearData map[string]interface{}
	if blankYearTotals.Count > 0 {
		blankBalance := blankYearTotals.Debits.Sub(blankYearTotals.Credits)
		blankYearData = map[string]interface{}{
			"debits":       blankYearTotals.Debits.ToFloat64(),
			"credits":      blankYearTotals.Credits.ToFloat64(),
			"balance":      blankBalance.ToFloat64(),
			"record_count": blankYearTotals.Count,
			"periods":      blankYearTotals.Periods,
		}
	}
	
	// Calculate overall balance
	overallBalance := totalDebits.Sub(totalCredits)
	
	return map[string]interface{}{
		"account_number":     accountNumber,
		"yearly_balances":    yearlyResults,
		"blank_year_totals":  blankYearData,
		"total_debits":       totalDebits.ToFloat64(),
		"total_credits":      totalCredits.ToFloat64(),
		"overall_balance":    overallBalance.ToFloat64(),
		"total_records":      totalRecords,
		"years_found":        len(yearlyTotals),
		"all_accounts_totals": allAccountsTotals, // For comparison
	}, nil
}

// ValidateGLBalances performs comprehensive GL validation checks
func (s *Service) ValidateGLBalances(companyName, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("ValidateGLBalances: Starting validation for account %s in company %s\n", accountNumber, companyName)
	
	// Read GLMASTER.dbf
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.dbf", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read GLMASTER.dbf: %w", err)
	}
	
	// Get column indices
	glColumns, ok := glData["columns"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid GLMASTER.dbf structure")
	}
	
	var accountIdx, debitIdx, creditIdx, yearIdx, periodIdx int = -1, -1, -1, -1, -1
	for i, col := range glColumns {
		colUpper := strings.ToUpper(col)
		switch colUpper {
		case "CACCTNO", "ACCOUNT", "ACCTNO":
			accountIdx = i
		case "NDEBITS", "DEBIT", "NDEBIT":
			debitIdx = i
		case "NCREDITS", "CREDIT", "NCREDIT":
			creditIdx = i
		case "CYEAR":
			yearIdx = i
		case "CPERIOD":
			periodIdx = i
		}
	}
	
	result := make(map[string]interface{})
	
	// Validation check 1: Debits = Credits for entire GL (double-entry bookkeeping)
	totalDebits := currency.Zero()
	totalCredits := currency.Zero()
	var debitCreditByYear = make(map[string]map[string]currency.Currency)
	var duplicateTransactions []map[string]interface{}
	var zeroAmountTransactions int
	var suspiciousAmounts []map[string]interface{}
	var outOfBalanceAccounts = make(map[string]map[string]currency.Currency)
	
	glRows, _ := glData["rows"].([][]interface{})
	
	// Track transactions for duplicate detection
	transactionMap := make(map[string][]int) // key: account+debit+credit+year+period, value: row indices
	
	for idx, row := range glRows {
		if len(row) <= accountIdx {
			continue
		}
		
		rowAccount := strings.TrimSpace(fmt.Sprintf("%v", row[accountIdx]))
		
		// Get amounts using decimal arithmetic
		debit := currency.Zero()
		if debitIdx != -1 && len(row) > debitIdx && row[debitIdx] != nil {
			debit = currency.ParseFromDBF(row[debitIdx])
		}
		
		credit := currency.Zero()
		if creditIdx != -1 && len(row) > creditIdx && row[creditIdx] != nil {
			credit = currency.ParseFromDBF(row[creditIdx])
		}
		
		// Get year
		year := ""
		if yearIdx != -1 && len(row) > yearIdx && row[yearIdx] != nil {
			year = strings.TrimSpace(fmt.Sprintf("%v", row[yearIdx]))
		}
		if year == "" {
			year = "BLANK"
		}
		
		// Get period
		period := ""
		if periodIdx != -1 && len(row) > periodIdx && row[periodIdx] != nil {
			period = strings.TrimSpace(fmt.Sprintf("%v", row[periodIdx]))
		}
		
		// Accumulate totals
		totalDebits = totalDebits.Add(debit)
		totalCredits = totalCredits.Add(credit)
		
		// Track by year for balance validation
		if _, exists := debitCreditByYear[year]; !exists {
			debitCreditByYear[year] = map[string]currency.Currency{"debits": currency.Zero(), "credits": currency.Zero()}
		}
		debitCreditByYear[year]["debits"] = debitCreditByYear[year]["debits"].Add(debit)
		debitCreditByYear[year]["credits"] = debitCreditByYear[year]["credits"].Add(credit)
		
		// Track by account for out-of-balance detection
		if accountNumber == "" || rowAccount == accountNumber {
			if _, exists := outOfBalanceAccounts[rowAccount]; !exists {
				outOfBalanceAccounts[rowAccount] = map[string]currency.Currency{"debits": currency.Zero(), "credits": currency.Zero()}
			}
			outOfBalanceAccounts[rowAccount]["debits"] = outOfBalanceAccounts[rowAccount]["debits"].Add(debit)
			outOfBalanceAccounts[rowAccount]["credits"] = outOfBalanceAccounts[rowAccount]["credits"].Add(credit)
		}
		
		// Check for zero amount transactions
		if debit.IsZero() && credit.IsZero() {
			zeroAmountTransactions++
		}
		
		// Check for suspicious amounts (very large transactions)
		oneMillion := currency.NewFromFloat(1000000)
		if debit.GreaterThan(oneMillion) || credit.GreaterThan(oneMillion) {
			suspiciousAmounts = append(suspiciousAmounts, map[string]interface{}{
				"row_index": idx + 1,
				"account":   rowAccount,
				"debit":     debit.ToFloat64(),
				"credit":    credit.ToFloat64(),
				"year":      year,
				"period":    period,
			})
		}
		
		// Check for duplicate transactions
		transKey := fmt.Sprintf("%s|%s|%s|%s|%s", rowAccount, debit.ToString(), credit.ToString(), year, period)
		if existingRows, exists := transactionMap[transKey]; exists {
			// Found potential duplicate
			if len(duplicateTransactions) < 10 { // Limit to first 10 duplicates
				duplicateTransactions = append(duplicateTransactions, map[string]interface{}{
					"row_indices":  append(existingRows, idx+1),
					"account":      rowAccount,
					"debit":        debit.ToFloat64(),
					"credit":       credit.ToFloat64(),
					"year":         year,
					"period":       period,
					"occurrence":   len(existingRows) + 1,
				})
			}
			transactionMap[transKey] = append(existingRows, idx+1)
		} else {
			transactionMap[transKey] = []int{idx + 1}
		}
	}
	
	// Calculate out-of-balance difference
	overallDifference := totalDebits.Sub(totalCredits).Abs()
	isBalanced := overallDifference.LessThan(currency.NewFromFloat(0.01)) // Allow for rounding errors
	
	// Build year-by-year balance check
	yearBalanceChecks := []map[string]interface{}{}
	for year, amounts := range debitCreditByYear {
		difference := amounts["debits"].Sub(amounts["credits"]).Abs()
		yearBalanceChecks = append(yearBalanceChecks, map[string]interface{}{
			"year":       year,
			"debits":     amounts["debits"].ToFloat64(),
			"credits":    amounts["credits"].ToFloat64(),
			"difference": difference.ToFloat64(),
			"balanced":   difference.LessThan(currency.NewFromFloat(0.01)),
		})
	}
	
	// Sort year balance checks
	sort.Slice(yearBalanceChecks, func(i, j int) bool {
		yearI := yearBalanceChecks[i]["year"].(string)
		yearJ := yearBalanceChecks[j]["year"].(string)
		return yearI > yearJ
	})
	
	// Find accounts with significant imbalances
	imbalancedAccounts := []map[string]interface{}{}
	for account, amounts := range outOfBalanceAccounts {
		difference := amounts["debits"].Sub(amounts["credits"]).Abs()
		if difference.GreaterThan(currency.NewFromFloat(0.01)) && account != "" { // Significant imbalance
			imbalancedAccounts = append(imbalancedAccounts, map[string]interface{}{
				"account":    account,
				"debits":     amounts["debits"].ToFloat64(),
				"credits":    amounts["credits"].ToFloat64(),
				"difference": difference.ToFloat64(),
			})
		}
	}
	
	// Sort imbalanced accounts by difference (largest first)
	sort.Slice(imbalancedAccounts, func(i, j int) bool {
		diffI := imbalancedAccounts[i]["difference"].(float64)
		diffJ := imbalancedAccounts[j]["difference"].(float64)
		return diffI > diffJ
	})
	
	// Limit to top 20 imbalanced accounts
	if len(imbalancedAccounts) > 20 {
		imbalancedAccounts = imbalancedAccounts[:20]
	}
	
	result["total_debits"] = totalDebits.ToFloat64()
	result["total_credits"] = totalCredits.ToFloat64()
	result["overall_difference"] = overallDifference.ToFloat64()
	result["is_balanced"] = isBalanced
	result["year_balance_checks"] = yearBalanceChecks
	result["duplicate_transactions"] = duplicateTransactions
	result["duplicate_count"] = len(duplicateTransactions)
	result["zero_amount_transactions"] = zeroAmountTransactions
	result["suspicious_amounts"] = suspiciousAmounts
	result["suspicious_count"] = len(suspiciousAmounts)
	result["imbalanced_accounts"] = imbalancedAccounts
	result["imbalanced_count"] = len(imbalancedAccounts)
	result["total_rows_checked"] = len(glRows)
	
	return result, nil
}

// CheckGLPeriodFields checks for blank CYEAR/CPERIOD fields in GLMASTER.dbf
func (s *Service) CheckGLPeriodFields(companyName string) (map[string]interface{}, error) {
	fmt.Printf("CheckGLPeriodFields: Checking GLMASTER.dbf for blank period fields\n")
	
	// Read GLMASTER.dbf
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.dbf", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read GLMASTER.dbf: %w", err)
	}
	
	// Get column indices
	glColumns, ok := glData["columns"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid GLMASTER.dbf structure")
	}
	
	var yearIdx, periodIdx, accountIdx, debitIdx, creditIdx int = -1, -1, -1, -1, -1
	for i, col := range glColumns {
		colUpper := strings.ToUpper(col)
		switch colUpper {
		case "CYEAR":
			yearIdx = i
		case "CPERIOD":
			periodIdx = i
		case "CACCTNO":
			accountIdx = i
		case "NDEBITS", "NDEBIT":
			debitIdx = i
		case "NCREDITS", "NCREDIT":
			creditIdx = i
		}
	}
	
	fmt.Printf("Column indices - CYEAR: %d, CPERIOD: %d, CACCTNO: %d\n", yearIdx, periodIdx, accountIdx)
	
	// Analyze the data
	glRows, _ := glData["rows"].([][]interface{})
	totalRows := len(glRows)
	blankYearCount := 0
	blankPeriodCount := 0
	blankBothCount := 0
	var sampleBlankRows []map[string]interface{}
	yearValues := make(map[string]int)
	periodValues := make(map[string]int)
	
	for i, row := range glRows {
		if len(row) <= accountIdx {
			continue
		}
		
		yearVal := ""
		periodVal := ""
		
		if yearIdx >= 0 && len(row) > yearIdx {
			yearVal = strings.TrimSpace(fmt.Sprintf("%v", row[yearIdx]))
		}
		if periodIdx >= 0 && len(row) > periodIdx {
			periodVal = strings.TrimSpace(fmt.Sprintf("%v", row[periodIdx]))
		}
		
		// Track unique values
		if yearVal != "" {
			yearValues[yearVal]++
		}
		if periodVal != "" {
			periodValues[periodVal]++
		}
		
		// Check for blanks
		yearBlank := yearVal == "" || yearVal == "<nil>"
		periodBlank := periodVal == "" || periodVal == "<nil>"
		
		if yearBlank {
			blankYearCount++
		}
		if periodBlank {
			blankPeriodCount++
		}
		if yearBlank && periodBlank {
			blankBothCount++
			
			// Capture sample blank rows
			if len(sampleBlankRows) < 5 {
				sampleRow := make(map[string]interface{})
				if accountIdx >= 0 && len(row) > accountIdx {
					sampleRow["account"] = row[accountIdx]
				}
				if debitIdx >= 0 && len(row) > debitIdx {
					sampleRow["debit"] = row[debitIdx]
				}
				if creditIdx >= 0 && len(row) > creditIdx {
					sampleRow["credit"] = row[creditIdx]
				}
				sampleRow["row_index"] = i
				sampleBlankRows = append(sampleBlankRows, sampleRow)
			}
		}
	}
	
	return map[string]interface{}{
		"total_rows":        totalRows,
		"blank_year_count":  blankYearCount,
		"blank_period_count": blankPeriodCount,
		"blank_both_count":  blankBothCount,
		"sample_blank_rows": sampleBlankRows,
		"unique_years":      yearValues,
		"unique_periods":    periodValues,
	}, nil
}

// RunClosingProcess runs the period closing process
func (s *Service) RunClosingProcess(companyName, periodEnd, closingDate, description string, forceClose bool) (*ClosingResult, error) {
	// Implementation will be moved from main.go
	// This will create closing entries and update period status
	return nil, fmt.Errorf("not implemented")
}

// GetClosingStatus gets the closing status for a period
func (s *Service) GetClosingStatus(companyName, periodEnd string) (string, error) {
	// Implementation will be moved from main.go
	return "", fmt.Errorf("not implemented")
}

// ReopenPeriod reopens a closed period
func (s *Service) ReopenPeriod(companyName, periodEnd, reason string) error {
	// Implementation will be moved from main.go
	// This will reverse closing entries and update status
	return fmt.Errorf("not implemented")
}

// GetGLEntries retrieves GL entries with optional filters
func (s *Service) GetGLEntries(companyName, accountNumber string, startDate, endDate time.Time) ([]GLEntry, error) {
	// Implementation to read from GLMASTER.dbf
	return nil, fmt.Errorf("not implemented")
}

// GetChartOfAccounts retrieves the chart of accounts
func (s *Service) GetChartOfAccounts(companyName, sortBy string, includeInactive bool) ([]map[string]interface{}, error) {
	// Implementation will be moved from main.go
	// This will read from COA.dbf
	return nil, fmt.Errorf("not implemented")
}

// GetAccountInfo retrieves detailed information for a specific account
func (s *Service) GetAccountInfo(companyName, accountNumber string) (map[string]interface{}, error) {
	// Implementation to get account details from COA.dbf
	return nil, fmt.Errorf("not implemented")
}

// ValidateAccountNumber validates if an account number exists
func (s *Service) ValidateAccountNumber(companyName, accountNumber string) (bool, error) {
	// Implementation to check if account exists in COA.dbf
	return false, fmt.Errorf("not implemented")
}

// Private helper methods

// calculatePeriodBalances calculates balances by period
func (s *Service) calculatePeriodBalances(entries []GLEntry) map[string]float64 {
	// Implementation will be moved from main.go
	return nil
}

// validateClosingEntries validates entries before closing
func (s *Service) validateClosingEntries(entries []GLEntry) []string {
	// Implementation to validate closing entries
	return nil
}