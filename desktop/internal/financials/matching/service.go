package matching

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/logger"
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
	ID                 int                    `json:"id"`
	CompanyName        string                 `json:"company_name"`
	AccountNumber      string                 `json:"account_number"`
	StatementID        int                    `json:"statement_id"`
	TransactionDate    string                 `json:"transaction_date"`
	CheckNumber        string                 `json:"check_number"`
	Description        string                 `json:"description"`
	Amount             float64                `json:"amount"`
	TransactionType    string                 `json:"transaction_type"`
	ImportBatchID      string                 `json:"import_batch_id"`
	ImportDate         string                 `json:"import_date"`
	ImportedBy         string                 `json:"imported_by"`
	MatchedCheckID     string                 `json:"matched_check_id"`
	MatchedDBFRowIndex int                    `json:"matched_dbf_row_index"`
	MatchConfidence    float64                `json:"match_confidence"`
	MatchType          string                 `json:"match_type"`
	IsMatched          bool                   `json:"is_matched"`
	CreatedAt          time.Time              `json:"created_at"`
}

// MatchResult represents the result of a matching operation
type MatchResult struct {
	BankTransaction BankTransaction        `json:"bankTransaction"`
	MatchedCheck    map[string]interface{} `json:"matchedCheck"`
	Confidence      float64                `json:"confidence"`
	MatchType       string                 `json:"matchType"`
	Confirmed       bool                   `json:"confirmed"`
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
func (s *Service) RunMatching(companyName, accountNumber string, options map[string]interface{}) (map[string]interface{}, error) {
	logger.WriteInfo("RunMatching", fmt.Sprintf("Called for company: %s, account: %s, options: %+v", companyName, accountNumber, options))
	
	// Extract options
	var statementDate *time.Time
	includeAllDates := true // Default to matching all dates
	
	if options != nil {
		// Check if we should limit to statement date
		if limitToStatement, ok := options["limitToStatementDate"].(bool); ok && limitToStatement {
			includeAllDates = false
			
			// Get the statement date
			if dateStr, ok := options["statementDate"].(string); ok && dateStr != "" {
				if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
					statementDate = &parsedDate
					logger.WriteInfo("RunMatching", fmt.Sprintf("Will limit matching to checks dated on or before: %s", statementDate.Format("2006-01-02")))
				}
			}
		}
	}
	
	// Get unmatched bank transactions
	transactions, err := s.GetBankTransactions(companyName, accountNumber, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get bank transactions: %w", err)
	}
	
	// Get existing checks for matching
	checksData, err := s.getOutstandingChecks(companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing checks: %w", err)
	}
	
	existingChecks, ok := checksData["checks"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid checks data format")
	}
	
	// Filter checks by date if requested
	var checksToMatch []map[string]interface{}
	if !includeAllDates && statementDate != nil {
		checksToMatch = make([]map[string]interface{}, 0)
		for _, check := range existingChecks {
			// Get check date
			if checkDateStr, ok := check["checkDate"].(string); ok {
				checkDate, err := time.Parse("2006-01-02", checkDateStr)
				if err == nil && !checkDate.After(*statementDate) {
					checksToMatch = append(checksToMatch, check)
				}
			}
		}
		logger.WriteInfo("RunMatching", fmt.Sprintf("Filtered checks from %d to %d (on or before %s)", 
			len(existingChecks), len(checksToMatch), statementDate.Format("2006-01-02")))
	} else {
		checksToMatch = existingChecks
		logger.WriteInfo("RunMatching", fmt.Sprintf("Using all %d checks for matching (no date filter)", len(checksToMatch)))
	}
	
	// Run matching algorithm
	logger.WriteInfo("RunMatching", fmt.Sprintf("Matching %d bank transactions with %d checks", len(transactions), len(checksToMatch)))
	matches := s.autoMatchBankTransactions(transactions, checksToMatch)
	logger.WriteInfo("RunMatching", fmt.Sprintf("Found %d matches", len(matches)))
	
	// Update the database with matches
	matchedCount := 0
	for _, match := range matches {
		if match.Confidence > 0.5 {
			updateQuery := `
				UPDATE bank_transactions 
				SET matched_check_id = ?, 
				    matched_dbf_row_index = ?,
				    match_confidence = ?,
				    match_type = ?,
				    is_matched = 1
				WHERE id = ?
			`
			
			checkID := ""
			rowIndex := 0
			if id, ok := match.MatchedCheck["id"]; ok {
				checkID = fmt.Sprintf("%v", id)
			}
			if idx, ok := match.MatchedCheck["_rowIndex"]; ok {
				if fidx, ok := idx.(float64); ok {
					rowIndex = int(fidx)
				}
			}
			
			_, err := s.db.Exec(updateQuery, checkID, rowIndex, match.Confidence, match.MatchType, match.BankTransaction.ID)
			if err == nil {
				matchedCount++
				logger.WriteInfo("RunMatching", fmt.Sprintf("Successfully matched bank txn %d to check %s", match.BankTransaction.ID, checkID))
			} else {
				logger.WriteError("RunMatching", fmt.Sprintf("Failed to update match for bank txn %d: %v", match.BankTransaction.ID, err))
			}
		}
	}
	
	return map[string]interface{}{
		"status": "success",
		"totalMatched": matchedCount,
		"totalProcessed": len(transactions),
		"matches": matches,
	}, nil
}

// ClearMatchesAndRerun clears existing matches and reruns matching
func (s *Service) ClearMatchesAndRerun(companyName, accountNumber string, options map[string]interface{}) (map[string]interface{}, error) {
	logger.WriteInfo("ClearMatchesAndRerun", fmt.Sprintf("Called for company: %s, account: %s", companyName, accountNumber))
	
	// Clear all existing matches for this account
	clearQuery := `
		UPDATE bank_transactions 
		SET matched_check_id = NULL,
		    matched_dbf_row_index = 0,
		    match_confidence = 0,
		    match_type = '',
		    is_matched = 0,
		    manually_matched = 0
		WHERE company_name = ? AND account_number = ?
	`
	
	result, err := s.db.Exec(clearQuery, companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to clear matches: %w", err)
	}
	
	clearedRows, _ := result.RowsAffected()
	logger.WriteInfo("ClearMatchesAndRerun", fmt.Sprintf("Cleared %d existing matches", clearedRows))
	
	// Now run matching again
	return s.RunMatching(companyName, accountNumber, options)
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
	query := `
		SELECT id, company_name, account_number, statement_id, transaction_date,
		       check_number, description, amount, transaction_type,
		       import_batch_id, import_date, imported_by,
		       matched_check_id, matched_dbf_row_index, match_confidence,
		       match_type, is_matched
		FROM bank_transactions
		WHERE company_name = ?
		  AND account_number = ?
	`
	
	args := []interface{}{companyName, accountNumber}
	
	if importBatchID != "" {
		query += " AND import_batch_id = ?"
		args = append(args, importBatchID)
	}
	
	query += " ORDER BY transaction_date DESC"
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var transactions []BankTransaction
	for rows.Next() {
		var t BankTransaction
		err := rows.Scan(
			&t.ID, &t.CompanyName, &t.AccountNumber, &t.StatementID,
			&t.TransactionDate, &t.CheckNumber, &t.Description, &t.Amount,
			&t.TransactionType, &t.ImportBatchID, &t.ImportDate, &t.ImportedBy,
			&t.MatchedCheckID, &t.MatchedDBFRowIndex, &t.MatchConfidence,
			&t.MatchType, &t.IsMatched,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	
	return transactions, nil
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

// getOutstandingChecks retrieves checks from DBF file
func (s *Service) getOutstandingChecks(companyName, accountNumber string) (map[string]interface{}, error) {
	// Read CHECKS.dbf file
	checksData, err := company.ReadDBFFile(companyName, "CHECKS.dbf", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("error reading CHECKS.dbf: %v", err)
	}
	
	// Process and filter checks
	// This is a simplified version - full implementation would filter by account
	return map[string]interface{}{
		"checks": s.processChecksData(checksData, accountNumber),
	}, nil
}

// processChecksData converts DBF data to check format
func (s *Service) processChecksData(dbfData map[string]interface{}, accountNumber string) []map[string]interface{} {
	var checks []map[string]interface{}
	
	columns, _ := dbfData["columns"].([]string)
	rows, _ := dbfData["rows"].([][]interface{})
	
	for i, row := range rows {
		checkMap := make(map[string]interface{})
		for j, col := range columns {
			if j < len(row) {
				checkMap[col] = row[j]
			}
		}
		
		// Add row index for tracking
		checkMap["_rowIndex"] = float64(i)
		
		// Map standard fields
		checkMap["id"] = fmt.Sprintf("%v", checkMap["CIDCHEC"])
		checkMap["checkNumber"] = fmt.Sprintf("%v", checkMap["CCHECKNO"])
		checkMap["date"] = fmt.Sprintf("%v", checkMap["DCHECKDATE"])
		checkMap["payee"] = fmt.Sprintf("%v", checkMap["CPAYEE"])
		checkMap["amount"] = checkMap["NAMOUNT"]
		
		// Filter by account if specified
		if accountNumber != "" {
			if acct, ok := checkMap["CACCTNO"]; ok {
				if fmt.Sprintf("%v", acct) != accountNumber {
					continue
				}
			}
		}
		
		// Only include uncleared, non-void checks
		if cleared, ok := checkMap["LCLEARED"].(bool); ok && cleared {
			continue
		}
		if void, ok := checkMap["LVOID"].(bool); ok && void {
			continue
		}
		
		checks = append(checks, checkMap)
	}
	
	return checks
}

// autoMatchBankTransactions performs automatic matching
func (s *Service) autoMatchBankTransactions(bankTransactions []BankTransaction, existingChecks []map[string]interface{}) []MatchResult {
	var matches []MatchResult
	
	// Keep track of already matched check IDs to prevent double-matching
	matchedCheckIDs := make(map[string]bool)
	
	// Sort bank transactions by date to match older transactions first
	// Note: Using simple date comparison for now
	for i := range bankTransactions {
		txn := &bankTransactions[i]
		
		// Skip only deposits, not checks
		if txn.TransactionType == "Deposit" {
			continue
		}
		
		// Filter out already matched checks
		var availableChecks []map[string]interface{}
		for _, check := range existingChecks {
			if checkID, ok := check["id"].(string); ok {
				if !matchedCheckIDs[checkID] {
					availableChecks = append(availableChecks, check)
				}
			}
		}
		
		bestMatch := s.findBestCheckMatch(txn, availableChecks)
		if bestMatch != nil && bestMatch.Confidence > 0.5 {
			// Mark this check as matched
			if checkID, ok := bestMatch.MatchedCheck["id"]; ok {
				matchedCheckIDs[fmt.Sprintf("%v", checkID)] = true
			}
			
			// Update the transaction with match info
			if checkID, ok := bestMatch.MatchedCheck["id"]; ok {
				txn.MatchedCheckID = fmt.Sprintf("%v", checkID)
			}
			if rowIndex, ok := bestMatch.MatchedCheck["_rowIndex"]; ok {
				if idx, ok := rowIndex.(float64); ok {
					txn.MatchedDBFRowIndex = int(idx)
				}
			}
			txn.MatchConfidence = bestMatch.Confidence
			txn.MatchType = bestMatch.MatchType
			txn.IsMatched = true
			
			matches = append(matches, *bestMatch)
		}
	}
	
	return matches
}

// findBestCheckMatch finds the best matching check for a bank transaction
func (s *Service) findBestCheckMatch(txn *BankTransaction, existingChecks []map[string]interface{}) *MatchResult {
	var bestMatch *MatchResult
	highestScore := 0.0
	
	for _, check := range existingChecks {
		score := s.calculateMatchScore(*txn, check)
		if score > highestScore && score > 0.5 { // Minimum confidence threshold
			matchType := s.determineMatchType(score, *txn, check)
			bestMatch = &MatchResult{
				BankTransaction: *txn,
				MatchedCheck:    check,
				Confidence:      score,
				MatchType:       matchType,
				Confirmed:       false,
			}
			bestMatch.BankTransaction.MatchedCheckID = fmt.Sprintf("%v", check["id"])
			bestMatch.BankTransaction.MatchConfidence = score
			bestMatch.BankTransaction.MatchType = matchType
			highestScore = score
		}
	}
	
	return bestMatch
}

// calculateMatchScore calculates the match score between a transaction and check
func (s *Service) calculateMatchScore(txn BankTransaction, check map[string]interface{}) float64 {
	score := 0.0
	
	// Amount matching (35% weight)
	checkAmount := s.parseFloat(check["amount"])
	txnAmount := math.Abs(txn.Amount)
	
	if checkAmount > 0 && math.Abs(txnAmount-checkAmount) < 0.01 {
		score += 0.35 // Exact amount match
	} else if checkAmount > 0 && math.Abs(txnAmount-checkAmount) < 1.0 {
		score += 0.2 // Close amount match
	} else {
		return 0.0 // No amount match
	}
	
	// Check number matching (25% weight)
	if txn.CheckNumber != "" {
		checkNumber := fmt.Sprintf("%v", check["checkNumber"])
		if txn.CheckNumber == checkNumber {
			score += 0.25
		} else if strings.Contains(txn.CheckNumber, checkNumber) || strings.Contains(checkNumber, txn.CheckNumber) {
			score += 0.1
		}
	}
	
	// Date proximity matching (40% weight)
	if txn.TransactionDate != "" {
		txnDate, txnErr := time.Parse("2006-01-02", txn.TransactionDate)
		checkDateStr := fmt.Sprintf("%v", check["date"])
		checkDate, checkErr := time.Parse("2006-01-02", checkDateStr)
		
		if txnErr == nil && checkErr == nil {
			daysDiff := math.Abs(txnDate.Sub(checkDate).Hours() / 24)
			
			if daysDiff == 0 {
				score += 0.4 // Same day
			} else if daysDiff <= 1 {
				score += 0.35 // Next day
			} else if daysDiff <= 3 {
				score += 0.25 // Within 3 days
			} else if daysDiff <= 7 {
				score += 0.15 // Within a week
			} else if daysDiff <= 14 {
				score += 0.05 // Within 2 weeks
			}
		}
	}
	
	// Description/Payee matching (bonus points)
	if description, ok := check["payee"].(string); ok && txn.Description != "" {
		descUpper := strings.ToUpper(description)
		txnDescUpper := strings.ToUpper(txn.Description)
		
		if strings.Contains(txnDescUpper, descUpper) || strings.Contains(descUpper, txnDescUpper) {
			score += 0.1 // Bonus for description match
		}
	}
	
	return score
}

// determineMatchType determines the type of match based on score and criteria
func (s *Service) determineMatchType(score float64, txn BankTransaction, check map[string]interface{}) string {
	if score >= 0.95 {
		return "exact"
	} else if score >= 0.8 {
		return "high_confidence"
	} else if score >= 0.6 {
		return "medium_confidence"
	} else {
		return "low_confidence"
	}
}

// parseFloat safely parses a float from an interface{}
func (s *Service) parseFloat(val interface{}) float64 {
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
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0.0
}