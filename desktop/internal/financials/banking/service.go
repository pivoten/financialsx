package banking

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/debug"
)

// Service handles all banking-related operations
type Service struct {
	db *sql.DB
}

// NewService creates a new banking service
func NewService(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
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
	CheckNumber     string    `json:"check_number"`
	CheckDate       time.Time `json:"check_date"`
	Payee           string    `json:"payee"`
	Amount          float64   `json:"amount"`
	AccountNumber   string    `json:"account_number"`
	DaysOutstanding int       `json:"days_outstanding"`
	IsStale         bool      `json:"is_stale"`
	RowIndex        int       `json:"row_index"`
	CIDCHEC         string    `json:"cidchec"`
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

// GetOutstandingChecks retrieves outstanding checks for an account
func (s *Service) GetOutstandingChecks(companyName, accountNumber string) ([]OutstandingCheck, error) {
	// Implementation will be moved from main.go
	// This will read from CHECKS.dbf where LCLEARED = false and LVOID = false
	return nil, fmt.Errorf("not implemented")
}

// GetCachedBalances retrieves all cached balances for a company
func (s *Service) GetCachedBalances(companyName string) ([]BalanceInfo, error) {
	// Implementation will be moved from main.go
	// This will read from account_balances table
	return nil, fmt.Errorf("not implemented")
}

// RefreshAccountBalance refreshes the cached balance for a single account
func (s *Service) RefreshAccountBalance(companyName, accountNumber string, username string) (*BalanceInfo, error) {
	// Implementation will be moved from main.go
	// This will call RefreshGLBalance and RefreshOutstandingChecks
	return nil, fmt.Errorf("not implemented")
}

// RefreshAllBalances refreshes all cached balances for a company
func (s *Service) RefreshAllBalances(companyName string, username string) ([]BalanceInfo, error) {
	// Implementation will be moved from main.go
	// This will refresh all accounts in parallel
	return nil, fmt.Errorf("not implemented")
}

// GetBalanceHistory retrieves the balance change history for an account
func (s *Service) GetBalanceHistory(companyName, accountNumber string, limit int) ([]map[string]interface{}, error) {
	// Implementation will be moved from main.go
	// This will read from balance_history table
	return nil, fmt.Errorf("not implemented")
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