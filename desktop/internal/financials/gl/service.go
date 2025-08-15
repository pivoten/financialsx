package gl

import (
	"database/sql"
	"fmt"
	"time"
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

// AnalyzeGLBalancesByYear analyzes GL balances grouped by year
func (s *Service) AnalyzeGLBalancesByYear(companyName, accountNumber string) (*BalanceAnalysis, error) {
	// Implementation will be moved from main.go
	// This will analyze GLMASTER.dbf entries by year
	return nil, fmt.Errorf("not implemented")
}

// ValidateGLBalances validates GL balances for consistency
func (s *Service) ValidateGLBalances(companyName, accountNumber string) (map[string]interface{}, error) {
	// Implementation will be moved from main.go
	// This will check for balance inconsistencies
	return nil, fmt.Errorf("not implemented")
}

// CheckGLPeriodFields checks for period field consistency
func (s *Service) CheckGLPeriodFields(companyName string) (map[string]interface{}, error) {
	// Implementation will be moved from main.go
	// This will validate period fields in GLMASTER.dbf
	return nil, fmt.Errorf("not implemented")
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