package matching

import (
	"database/sql"
	"fmt"
	"time"
)

// Service handles transaction matching between bank statements and checks
type Service struct {
	db *sql.DB
}

// NewService creates a new matching service
func NewService(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
}

// BankTransaction represents a bank statement transaction
type BankTransaction struct {
	ID                int       `json:"id"`
	ImportBatchID     string    `json:"import_batch_id"`
	StatementID       int       `json:"statement_id"`
	TransactionDate   time.Time `json:"transaction_date"`
	Description       string    `json:"description"`
	Amount            float64   `json:"amount"`
	TransactionType   string    `json:"transaction_type"`
	CheckNumber       string    `json:"check_number"`
	AccountNumber     string    `json:"account_number"`
	MatchedCheckID    string    `json:"matched_check_id"`
	MatchConfidence   float64   `json:"match_confidence"`
	MatchType         string    `json:"match_type"`
	IsMatched         bool      `json:"is_matched"`
	CreatedAt         time.Time `json:"created_at"`
}

// MatchResult represents the result of a matching operation
type MatchResult struct {
	TransactionID   int     `json:"transaction_id"`
	CheckID         string  `json:"check_id"`
	CheckNumber     string  `json:"check_number"`
	CheckAmount     float64 `json:"check_amount"`
	MatchScore      float64 `json:"match_score"`
	MatchType       string  `json:"match_type"`
	MatchReason     string  `json:"match_reason"`
}

// MatchOptions contains options for matching operations
type MatchOptions struct {
	LimitToStatementDate bool      `json:"limit_to_statement_date"`
	StatementDate        time.Time `json:"statement_date"`
	IncludeVoidChecks    bool      `json:"include_void_checks"`
	MinMatchScore        float64   `json:"min_match_score"`
}

// ImportResult represents the result of a bank statement import
type ImportResult struct {
	BatchID          string        `json:"batch_id"`
	StatementID      int           `json:"statement_id"`
	TransactionCount int           `json:"transaction_count"`
	MatchedCount     int           `json:"matched_count"`
	UnmatchedCount   int           `json:"unmatched_count"`
	Transactions     []BankTransaction `json:"transactions"`
	MatchResults     []MatchResult     `json:"match_results"`
}

// RunMatching performs automatic matching between bank transactions and checks
func (s *Service) RunMatching(companyName, accountNumber string, options MatchOptions) (*ImportResult, error) {
	// Implementation will be moved from main.go
	// This will run the matching algorithm with date filtering options
	return nil, fmt.Errorf("not implemented")
}

// ClearMatchesAndRerun clears existing matches and reruns matching
func (s *Service) ClearMatchesAndRerun(companyName, accountNumber string, options MatchOptions) (*ImportResult, error) {
	// Implementation will be moved from main.go
	return nil, fmt.Errorf("not implemented")
}

// ImportBankStatement imports a bank statement from CSV
func (s *Service) ImportBankStatement(companyName, csvContent, accountNumber string) (*ImportResult, error) {
	// Implementation will be moved from main.go
	// This will parse CSV, store transactions, and run matching
	return nil, fmt.Errorf("not implemented")
}

// ManualMatchTransaction manually matches a transaction to a check
func (s *Service) ManualMatchTransaction(transactionID int, checkID string, checkRowIndex int) (*MatchResult, error) {
	// Implementation will be moved from main.go
	return nil, fmt.Errorf("not implemented")
}

// UnmatchTransaction removes a match between a transaction and check
func (s *Service) UnmatchTransaction(transactionID int) error {
	// Implementation will be moved from main.go
	return fmt.Errorf("not implemented")
}

// RetryMatching retries matching for a specific statement
func (s *Service) RetryMatching(companyName, accountNumber string, statementID int) (*ImportResult, error) {
	// Implementation will be moved from main.go
	return nil, fmt.Errorf("not implemented")
}

// GetMatchedTransactions retrieves all matched transactions
func (s *Service) GetMatchedTransactions(companyName, accountNumber string) ([]BankTransaction, error) {
	// Implementation will be moved from main.go
	return nil, fmt.Errorf("not implemented")
}

// GetBankTransactions retrieves bank transactions for a specific import batch
func (s *Service) GetBankTransactions(companyName, accountNumber, importBatchID string) ([]BankTransaction, error) {
	// Implementation will be moved from main.go
	return nil, fmt.Errorf("not implemented")
}

// GetRecentStatements retrieves recent bank statements
func (s *Service) GetRecentStatements(companyName, accountNumber string) ([]map[string]interface{}, error) {
	// Implementation will be moved from main.go
	return nil, fmt.Errorf("not implemented")
}

// DeleteStatement deletes a bank statement and its transactions
func (s *Service) DeleteStatement(companyName, importBatchID string) error {
	// Implementation will be moved from main.go
	return fmt.Errorf("not implemented")
}

// Private helper methods

// parseCSV parses CSV content into bank transactions
func (s *Service) parseCSV(csvContent string) ([]BankTransaction, error) {
	// Implementation will be moved from main.go
	return nil, fmt.Errorf("not implemented")
}

// autoMatch performs automatic matching algorithm
func (s *Service) autoMatch(transactions []BankTransaction, checks []map[string]interface{}, options MatchOptions) []MatchResult {
	// Implementation will be moved from main.go
	return nil
}

// calculateMatchScore calculates the match score between a transaction and check
func (s *Service) calculateMatchScore(txn BankTransaction, check map[string]interface{}) float64 {
	// Implementation will be moved from main.go
	return 0.0
}