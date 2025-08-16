package banking

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/database"
	"github.com/pivoten/financialsx/desktop/internal/debug"
)

// Service handles all banking-related operations
type Service struct {
	db    *sql.DB
	dbHelper *database.DB  // Add database helper for balance cache operations
}

// NewService creates a new banking service
func NewService(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
}

// SetDatabaseHelper sets the database helper for balance cache operations
func (s *Service) SetDatabaseHelper(dbHelper *database.DB) {
	s.dbHelper = dbHelper
}

// BankAccount represents a bank account
type BankAccount struct {
	AccountNumber   string    `json:"account_number"`
	AccountName     string    `json:"account_name"`
	AccountType     int       `json:"account_type"`
	Balance         float64   `json:"balance"`
	LastUpdated     time.Time `json:"last_updated"`
	IsActive        bool      `json:"is_active"`
	IsBankAccount   bool      `json:"is_bank_account"`
	CompanyName     string    `json:"company_name"`
}

// OutstandingCheck represents an outstanding check
type OutstandingCheck struct {
	CheckNumber     string    `json:"checkNumber"`
	CheckDate       string    `json:"date"`
	Payee           string    `json:"payee"`
	Amount          float64   `json:"amount"`
	AccountNumber   string    `json:"account"`
	EntryType       string    `json:"entryType"`  // D = Deposit, C = Check
	CIDCHEC         string    `json:"cidchec"`
	ID              string    `json:"id"`
	RowIndex        int       `json:"_rowIndex"`
	RawData         []interface{} `json:"_rawData,omitempty"`
}

// BalanceInfo represents cached balance information
type BalanceInfo struct {
	AccountNumber           string    `json:"account_number"`
	AccountName             string    `json:"account_name"`
	GLBalance              float64   `json:"gl_balance"`
	OutstandingChecksTotal float64   `json:"outstanding_checks_total"`
	OutstandingChecksCount int       `json:"outstanding_checks_count"`
	BankBalance            float64   `json:"bank_balance"`
	GLLastUpdated          time.Time `json:"gl_last_updated"`
	ChecksLastUpdated      time.Time `json:"checks_last_updated"`
	IsStale                bool      `json:"is_stale"`
}

// GetBankAccounts retrieves all bank accounts for a company
func (s *Service) GetBankAccounts(companyName string) ([]BankAccount, error) {
	debug.SimpleLog(fmt.Sprintf("Banking.GetBankAccounts: company=%s", companyName))
	
	// Read COA.dbf file (no limit - get all records for financial accuracy)
	coaData, err := company.ReadDBFFile(companyName, "COA.dbf", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read COA.dbf: %w", err)
	}
	
	if coaData == nil {
		return nil, fmt.Errorf("coaData is nil")
	}

	data, ok := coaData["rows"].([][]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format from COA.dbf")
	}
	
	if len(data) == 0 {
		return []BankAccount{}, nil // Return empty slice instead of error
	}
	
	var bankAccounts []BankAccount

	for _, row := range data {
		if len(row) < 7 {
			continue // Skip incomplete rows (need at least 7 columns for LBANKACCT)
		}

		// Check LBANKACCT flag in column 6 (Lbankacct)
		bankAccountFlag := false
		if len(row) > 6 {
			switch v := row[6].(type) {
			case bool:
				bankAccountFlag = v
			case string:
				bankAccountFlag = v == "T" || v == ".T." || v == "true"
			}
		}

		if bankAccountFlag {
			// Parse account type
			accountType := 0
			if row[1] != nil {
				switch v := row[1].(type) {
				case int:
					accountType = v
				case float64:
					accountType = int(v)
				case string:
					fmt.Sscanf(v, "%d", &accountType)
				}
			}
			
			account := BankAccount{
				AccountNumber:   fmt.Sprintf("%v", row[0]),   // Cacctno (Account number)
				AccountName:     fmt.Sprintf("%v", row[2]),   // Cacctdesc (Account description)
				AccountType:     accountType,                 // Account type
				Balance:         0.0,                         // Balance not in COA, will be calculated
				IsBankAccount:   true,
				IsActive:        true,
				CompanyName:     companyName,
				LastUpdated:     time.Now(),
			}
			bankAccounts = append(bankAccounts, account)
		}
	}
	
	debug.SimpleLog(fmt.Sprintf("Banking.GetBankAccounts: returning %d bank accounts", len(bankAccounts)))
	return bankAccounts, nil
}

// GetAccountBalance retrieves the GL balance for a specific account
func (s *Service) GetAccountBalance(companyName, accountNumber string) (float64, error) {
	debug.SimpleLog(fmt.Sprintf("Banking.GetAccountBalance: company=%s, account=%s", companyName, accountNumber))
	
	// Read GLMASTER.dbf to get account balance
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.dbf", "", 0, 0, "", "")
	if err != nil {
		return 0, fmt.Errorf("failed to read GLMASTER.dbf: %w", err)
	}
	
	// Get column indices for GLMASTER.dbf
	glColumns, ok := glData["columns"].([]string)
	if !ok {
		return 0, fmt.Errorf("invalid GLMASTER.dbf structure")
	}
	
	// Find relevant GL columns (GLMASTER has separate debit/credit columns)
	var accountIdx, debitIdx, creditIdx int = -1, -1, -1
	for i, col := range glColumns {
		colUpper := strings.ToUpper(col)
		if colUpper == "CACCTNO" || colUpper == "ACCOUNT" || colUpper == "ACCTNO" {
			accountIdx = i
		} else if colUpper == "NDEBITS" || colUpper == "DEBIT" || colUpper == "NDEBIT" {
			debitIdx = i
		} else if colUpper == "NCREDITS" || colUpper == "CREDIT" || colUpper == "NCREDIT" {
			creditIdx = i
		}
	}
	
	if accountIdx == -1 || (debitIdx == -1 && creditIdx == -1) {
		return 0, fmt.Errorf("required columns not found in GLMASTER.dbf")
	}
	
	// Sum all entries for this account (debits and credits)
	var totalBalance float64 = 0
	glRows, _ := glData["rows"].([][]interface{})
	
	for _, row := range glRows {
		if len(row) <= accountIdx {
			continue
		}
		
		// Check if this row is for our account
		rowAccount := fmt.Sprintf("%v", row[accountIdx])
		if rowAccount == accountNumber {
			var debit, credit float64 = 0, 0
			
			// Get debit amount if column exists
			if debitIdx != -1 && len(row) > debitIdx {
				debit = parseFloat(row[debitIdx])
			}
			
			// Get credit amount if column exists
			if creditIdx != -1 && len(row) > creditIdx {
				credit = parseFloat(row[creditIdx])
			}
			
			// For bank accounts, debits increase balance, credits decrease
			totalBalance += (debit - credit)
		}
	}
	
	debug.SimpleLog(fmt.Sprintf("Banking.GetAccountBalance: account=%s, balance=%.2f", accountNumber, totalBalance))
	return totalBalance, nil
}

// parseFloat is a helper to parse various numeric types to float64
func parseFloat(val interface{}) float64 {
	if val == nil {
		return 0
	}
	
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		var f float64
		fmt.Sscanf(v, "%f", &f)
		return f
	default:
		return 0
	}
}

// GetBankAccountsAsMap retrieves all bank accounts as map format for backward compatibility
func (s *Service) GetBankAccountsAsMap(companyName string) ([]map[string]interface{}, error) {
	accounts, err := s.GetBankAccounts(companyName)
	if err != nil {
		return nil, err
	}
	
	var result []map[string]interface{}
	for _, account := range accounts {
		result = append(result, map[string]interface{}{
			"account_number": account.AccountNumber,
			"account_name":   account.AccountName,
			"account_type":   account.AccountType,
			"balance":        account.Balance,
			"last_updated":   account.LastUpdated,
			"is_active":      account.IsActive,
			"is_bank_account": account.IsBankAccount,
			"company_name":   account.CompanyName,
		})
	}
	
	return result, nil
}

// RefreshAllBalances refreshes cached balances for all bank accounts
func (s *Service) RefreshAllBalances(companyName string) (map[string]interface{}, error) {
	fmt.Printf("RefreshAllBalances called for company: %s\n", companyName)

	// Get all bank accounts
	bankAccounts, err := s.GetBankAccountsAsMap(companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank accounts: %w", err)
	}

	var successCount, errorCount int
	var errors []string

	for _, account := range bankAccounts {
		if account == nil {
			errorCount++
			errors = append(errors, "Nil account in list")
			continue
		}

		accountNumberInterface, ok := account["account_number"]
		if !ok || accountNumberInterface == nil {
			errorCount++
			errors = append(errors, "Account missing account_number field")
			continue
		}

		accountNumber, ok := accountNumberInterface.(string)
		if !ok {
			errorCount++
			errors = append(errors, fmt.Sprintf("Invalid account_number type: %T", accountNumberInterface))
			continue
		}

		_, err := s.RefreshAccountBalance(companyName, accountNumber, "system")
		if err != nil {
			errorCount++
			errors = append(errors, fmt.Sprintf("Account %s: %v", accountNumber, err))
		} else {
			successCount++
		}
	}

	return map[string]interface{}{
		"status":         "completed",
		"total_accounts": len(bankAccounts),
		"success_count":  successCount,
		"error_count":    errorCount,
		"errors":         errors,
		"refresh_time":   time.Now(),
	}, nil
}

// GetOutstandingChecks retrieves outstanding checks for an account
func (s *Service) GetOutstandingChecks(companyName, accountNumber string) ([]OutstandingCheck, error) {
	debug.SimpleLog(fmt.Sprintf("Banking.GetOutstandingChecks: company=%s, account=%s", companyName, accountNumber))
	
	// Read checks.dbf - get all records
	checksData, err := company.ReadDBFFile(companyName, "checks.dbf", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read checks.dbf: %w", err)
	}
	
	// Get column indices for checks.dbf
	checksColumns, ok := checksData["columns"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid checks.dbf structure")
	}
	
	// Find relevant check columns
	var checkNumIdx, dateIdx, payeeIdx, amountIdx, accountIdx, clearedIdx, voidIdx, entryTypeIdx, cidchecIdx int = -1, -1, -1, -1, -1, -1, -1, -1, -1
	for i, col := range checksColumns {
		colUpper := strings.ToUpper(col)
		switch colUpper {
		case "CCHECKNO":
			checkNumIdx = i
		case "DCHECKDATE":
			dateIdx = i
		case "CPAYEE":
			payeeIdx = i
		case "NAMOUNT":
			amountIdx = i
		case "CACCTNO":
			accountIdx = i
		case "LCLEARED":
			clearedIdx = i
		case "LVOID":
			voidIdx = i
		case "CENTRYTYPE":
			entryTypeIdx = i
		case "CIDCHEC":
			cidchecIdx = i
		}
	}
	
	if checkNumIdx == -1 || amountIdx == -1 {
		return nil, fmt.Errorf("required columns not found in checks.dbf")
	}
	
	// Process check rows to find outstanding checks
	var outstandingChecks []OutstandingCheck
	checksRows, _ := checksData["rows"].([][]interface{})
	
	for rowIdx, row := range checksRows {
		if len(row) <= checkNumIdx || len(row) <= amountIdx {
			continue
		}
		
		// Get account for this check
		checkAccount := ""
		if accountIdx != -1 && len(row) > accountIdx && row[accountIdx] != nil {
			checkAccount = strings.TrimSpace(fmt.Sprintf("%v", row[accountIdx]))
		}
		
		// If account filter is provided, only include checks for that account
		if accountNumber != "" && checkAccount != accountNumber {
			continue
		}
		
		// Check if cleared
		isCleared := false
		if clearedIdx != -1 && len(row) > clearedIdx {
			isCleared = parseBool(row[clearedIdx])
		}
		
		// Check if voided
		isVoided := false
		if voidIdx != -1 && len(row) > voidIdx {
			isVoided = parseBool(row[voidIdx])
		}
		
		// Only include if not cleared and not voided
		if !isCleared && !isVoided {
			// Get entry type
			entryType := ""
			if entryTypeIdx != -1 && len(row) > entryTypeIdx {
				entryType = strings.TrimSpace(fmt.Sprintf("%v", row[entryTypeIdx]))
			}
			
			// Get CIDCHEC for unique identification
			cidchec := ""
			if cidchecIdx != -1 && len(row) > cidchecIdx && row[cidchecIdx] != nil {
				cidchec = fmt.Sprintf("%v", row[cidchecIdx])
			}
			
			// Get date
			dateStr := ""
			if dateIdx != -1 && len(row) > dateIdx && row[dateIdx] != nil {
				// Handle time.Time or string
				if t, ok := row[dateIdx].(time.Time); ok {
					dateStr = t.Format("2006-01-02")
				} else {
					dateStr = fmt.Sprintf("%v", row[dateIdx])
				}
			}
			
			// Get payee
			payee := ""
			if payeeIdx != -1 && len(row) > payeeIdx && row[payeeIdx] != nil {
				payee = fmt.Sprintf("%v", row[payeeIdx])
			}
			
			check := OutstandingCheck{
				CheckNumber:   fmt.Sprintf("%v", row[checkNumIdx]),
				CheckDate:     dateStr,
				Payee:         payee,
				Amount:        parseFloat(row[amountIdx]),
				AccountNumber: checkAccount,
				EntryType:     entryType,
				CIDCHEC:       cidchec,
				ID:            cidchec,
				RowIndex:      rowIdx,
				RawData:       row,
			}
			
			outstandingChecks = append(outstandingChecks, check)
		}
	}
	
	debug.SimpleLog(fmt.Sprintf("Banking.GetOutstandingChecks: Found %d outstanding checks", len(outstandingChecks)))
	return outstandingChecks, nil
}

// parseBool parses various boolean representations
func parseBool(val interface{}) bool {
	if val == nil {
		return false
	}
	
	switch v := val.(type) {
	case bool:
		return v
	case string:
		lowerVal := strings.ToLower(strings.TrimSpace(v))
		// Empty string should be FALSE
		if lowerVal == "" {
			return false
		}
		return lowerVal == "t" || lowerVal == ".t." || lowerVal == "true" || lowerVal == "1"
	default:
		return false
	}
}

// GetCachedBalancesTyped retrieves all cached balances for a company  
func (s *Service) GetCachedBalancesTyped(companyName string) ([]BalanceInfo, error) {
	debug.SimpleLog(fmt.Sprintf("Banking.GetCachedBalancesTyped: company=%s", companyName))
	
	if s.dbHelper == nil {
		return nil, fmt.Errorf("database helper not initialized")
	}
	
	balances, err := database.GetAllCachedBalances(s.dbHelper, companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached balances: %w", err)
	}
	
	// Convert database.CachedBalance to BalanceInfo
	result := make([]BalanceInfo, len(balances))
	for i, b := range balances {
		result[i] = BalanceInfo{
			AccountNumber:           b.AccountNumber,
			AccountName:             b.AccountName,
			GLBalance:              b.GLBalance,
			OutstandingChecksTotal: b.OutstandingTotal,
			OutstandingChecksCount: b.OutstandingCount,
			BankBalance:            b.BankBalance,
			GLLastUpdated:          b.GLLastUpdated,
			ChecksLastUpdated:      b.OutstandingLastUpdated,
			IsStale:                b.GLFreshness == "stale" || b.ChecksFreshness == "stale",
		}
	}
	
	debug.SimpleLog(fmt.Sprintf("Banking.GetCachedBalancesTyped: Retrieved %d balances", len(result)))
	return result, nil
}

// RefreshAccountBalance refreshes the cached balance for a single account
func (s *Service) RefreshAccountBalance(companyName, accountNumber string, username string) (*BalanceInfo, error) {
	debug.SimpleLog(fmt.Sprintf("Banking.RefreshAccountBalance: company=%s, account=%s, user=%s", 
		companyName, accountNumber, username))
	
	if s.dbHelper == nil {
		return nil, fmt.Errorf("database helper not initialized")
	}
	
	// Refresh GL balance
	fmt.Printf("RefreshAccountBalance: Starting GL balance refresh for account %s\n", accountNumber)
	err := database.RefreshGLBalance(s.dbHelper, companyName, accountNumber, username)
	if err != nil {
		fmt.Printf("RefreshAccountBalance: GL balance refresh failed: %v\n", err)
		// Don't return error, continue with outstanding checks refresh
		debug.SimpleLog(fmt.Sprintf("Warning: GL balance refresh failed: %v", err))
	} else {
		fmt.Printf("RefreshAccountBalance: GL balance refresh successful\n")
	}
	
	// Refresh outstanding checks/deposits
	fmt.Printf("RefreshAccountBalance: Starting outstanding checks refresh for account %s\n", accountNumber)
	err = database.RefreshOutstandingChecks(s.dbHelper, companyName, accountNumber, username)
	if err != nil {
		fmt.Printf("RefreshAccountBalance: Outstanding checks refresh failed: %v\n", err)
		// Don't return error, try to get current balance
		debug.SimpleLog(fmt.Sprintf("Warning: Outstanding checks refresh failed: %v", err))
	} else {
		fmt.Printf("RefreshAccountBalance: Outstanding checks refresh successful\n")
	}
	
	// Get the updated cached balance
	balance, err := database.GetCachedBalance(s.dbHelper, companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated balance: %w", err)
	}
	
	result := &BalanceInfo{
		AccountNumber:           balance.AccountNumber,
		AccountName:             balance.AccountName,
		GLBalance:              balance.GLBalance,
		OutstandingChecksTotal: balance.OutstandingTotal,
		OutstandingChecksCount: balance.OutstandingCount,
		BankBalance:            balance.BankBalance,
		GLLastUpdated:          balance.GLLastUpdated,
		ChecksLastUpdated:      balance.OutstandingLastUpdated,
		IsStale:                balance.GLFreshness == "stale" || balance.ChecksFreshness == "stale",
	}
	
	fmt.Printf("RefreshAccountBalance: Complete - GL: %.2f, Outstanding: %.2f, Bank: %.2f\n",
		result.GLBalance, result.OutstandingChecksTotal, result.BankBalance)
	
	return result, nil
}


// GetBalanceHistory retrieves the balance change history for an account
func (s *Service) GetBalanceHistory(companyName, accountNumber string, limit int) ([]map[string]interface{}, error) {
	// Implementation will be moved from main.go
	// This will read from balance_history table
	return nil, fmt.Errorf("not implemented")
}

// GetCachedBalances returns cached balances as interface{} for compatibility
func (s *Service) GetCachedBalances(companyName string) (interface{}, error) {
	balances, err := s.GetCachedBalancesTyped(companyName)
	if err != nil {
		return nil, err
	}
	
	// Convert to map format for compatibility
	result := make([]map[string]interface{}, len(balances))
	for i, b := range balances {
		result[i] = map[string]interface{}{
			"account_number":     b.AccountNumber,
			"account_name":       b.AccountName,
			"gl_balance":         b.GLBalance,
			"outstanding_total":  b.OutstandingChecksTotal,
			"outstanding_count":  b.OutstandingChecksCount,
			"bank_balance":       b.BankBalance,
			"gl_last_updated":    b.GLLastUpdated,
			"checks_last_updated": b.ChecksLastUpdated,
			"is_stale":           b.IsStale,
		}
	}
	
	return result, nil
}

// Helper methods that will be needed

// refreshGLBalance updates the GL balance from GLMASTER.dbf
func (s *Service) refreshGLBalance(companyName, accountNumber string, username string) error {
	// Implementation from database.RefreshGLBalance
	return fmt.Errorf("not implemented")
}

// refreshOutstandingChecks updates outstanding checks from CHECKS.dbf
func (s *Service) refreshOutstandingChecks(companyName, accountNumber string, username string) error {
	// Implementation from database.RefreshOutstandingChecks
	return fmt.Errorf("not implemented")
}