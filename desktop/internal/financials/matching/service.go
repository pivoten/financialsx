package matching

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
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
	ManuallyMatched    bool                   `json:"manually_matched"`
	IsReconciled       bool                   `json:"is_reconciled"`
	ReconciledDate     *string                `json:"reconciled_date"`   // Pointer to handle NULL
	ReconciliationID   *int                   `json:"reconciliation_id"` // Pointer to handle NULL
	ExtendedData       map[string]interface{} `json:"extended_data"`
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
	
	// Extract transactions array from the map
	var bankTxns []BankTransaction
	if txnList, ok := transactions["transactions"].([]interface{}); ok {
		for _, txn := range txnList {
			if txnMap, ok := txn.(map[string]interface{}); ok {
				bt := BankTransaction{}
				if id, ok := txnMap["id"].(float64); ok {
					bt.ID = int(id)
				}
				if date, ok := txnMap["transaction_date"].(string); ok {
					bt.TransactionDate = date
				}
				if desc, ok := txnMap["description"].(string); ok {
					bt.Description = desc
				}
				if amt, ok := txnMap["amount"].(float64); ok {
					bt.Amount = amt
				}
				if txnType, ok := txnMap["transaction_type"].(string); ok {
					bt.TransactionType = txnType
				}
				if account, ok := txnMap["account_number"].(string); ok {
					bt.AccountNumber = account
				}
				bankTxns = append(bankTxns, bt)
			}
		}
	}
	
	// Run matching algorithm
	logger.WriteInfo("RunMatching", fmt.Sprintf("Matching %d bank transactions with %d checks", len(bankTxns), len(checksToMatch)))
	matches := s.autoMatchBankTransactions(bankTxns, checksToMatch)
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
func (s *Service) ImportBankStatement(companyName, csvContent, accountNumber string) (map[string]interface{}, error) {
	logger.WriteInfo("ImportBankStatement", fmt.Sprintf("Called for company: %s, account: %s", companyName, accountNumber))
	
	// Generate unique batch ID for this import
	batchID := fmt.Sprintf("import_%d_%s", time.Now().Unix(), accountNumber)
	statementDate := time.Now().Format("2006-01-02")
	
	// Parse CSV content into BankTransaction objects
	bankTransactions, err := s.parseCSVToBankTransactions(csvContent, companyName, accountNumber, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}
	
	// Create bank statement record first
	statementID, err := s.createBankStatement(companyName, accountNumber, statementDate, batchID, len(bankTransactions))
	if err != nil {
		return nil, fmt.Errorf("failed to create bank statement: %w", err)
	}
	logger.WriteInfo("ImportBankStatement", fmt.Sprintf("Created bank statement with ID: %d", statementID))
	
	// Update transactions with statement ID
	for i := range bankTransactions {
		bankTransactions[i].StatementID = statementID
	}
	
	// Store bank transactions in SQLite
	logger.WriteInfo("ImportBankStatement", fmt.Sprintf("Storing %d bank transactions in database", len(bankTransactions)))
	err = s.storeBankTransactions(bankTransactions)
	if err != nil {
		return nil, fmt.Errorf("failed to store bank transactions: %w", err)
	}
	
	return map[string]interface{}{
		"status":            "success",
		"importBatchId":     batchID,
		"statementID":       statementID,
		"bankTransactions":  bankTransactions,
		"totalTransactions": len(bankTransactions),
		"message":           "Transactions imported successfully. Click 'Run Matching' to match with checks.",
	}, nil
}

// ManualMatchTransaction manually matches a transaction to a check
func (s *Service) ManualMatchTransaction(transactionID int, checkID string, checkRowIndex int) (map[string]interface{}, error) {
	logger.WriteInfo("ManualMatchTransaction", fmt.Sprintf("txn=%d, check=%s, row=%d", transactionID, checkID, checkRowIndex))
	
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	// Update the bank transaction with match info
	query := `
		UPDATE bank_transactions 
		SET matched_check_id = ?, 
		    matched_dbf_row_index = ?,
		    match_confidence = 1.0,
		    match_type = 'manual',
		    is_matched = 1,
		    manually_matched = 1
		WHERE id = ?
	`
	
	_, err := s.db.Exec(query, checkID, checkRowIndex, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}
	
	return map[string]interface{}{
		"status": "success",
		"message": "Transaction matched successfully",
	}, nil
}

// UnmatchTransaction removes a match between a transaction and check
func (s *Service) UnmatchTransaction(transactionID int) (map[string]interface{}, error) {
	logger.WriteInfo("UnmatchTransaction", fmt.Sprintf("Called for transaction ID: %d", transactionID))
	
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	// Update the transaction to unmatched
	query := `
		UPDATE bank_transactions 
		SET matched_check_id = NULL,
		    matched_dbf_row_index = 0,
		    match_confidence = 0,
		    match_type = '',
		    is_matched = 0,
		    manually_matched = 0
		WHERE id = ?
	`
	
	result, err := s.db.Exec(query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to unmatch transaction: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	
	return map[string]interface{}{
		"status": "success",
		"rowsAffected": rowsAffected,
	}, nil
}

// RetryMatching retries matching for a specific statement
func (s *Service) RetryMatching(companyName, accountNumber string, statementID int) (map[string]interface{}, error) {
	logger.WriteInfo("RetryMatching", fmt.Sprintf("For statement: %d", statementID))
	
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	// Get unmatched transactions for this statement
	query := `
		SELECT id, check_number, amount, transaction_date, description, company_name, account_number
		FROM bank_transactions 
		WHERE statement_id = ? AND is_matched = 0
	`
	
	rows, err := s.db.Query(query, statementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unmatched transactions: %w", err)
	}
	defer rows.Close()
	
	var unmatchedTxns []BankTransaction
	for rows.Next() {
		var txn BankTransaction
		err := rows.Scan(&txn.ID, &txn.CheckNumber, &txn.Amount, &txn.TransactionDate, 
			&txn.Description, &txn.CompanyName, &txn.AccountNumber)
		if err != nil {
			continue
		}
		unmatchedTxns = append(unmatchedTxns, txn)
	}
	
	// Get outstanding checks
	checksData, err := s.getOutstandingChecks(companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get checks: %w", err)
	}
	
	existingChecks, _ := checksData["checks"].([]map[string]interface{})
	
	// Run matching algorithm
	newMatchCount := 0
	for _, txn := range unmatchedTxns {
		bestMatch := s.findBestCheckMatch(&txn, existingChecks)
		if bestMatch != nil && bestMatch.Confidence > 0.5 {
			// Update the transaction
			checkID := ""
			rowIndex := 0
			if id, ok := bestMatch.MatchedCheck["id"]; ok {
				checkID = fmt.Sprintf("%v", id)
			}
			if idx, ok := bestMatch.MatchedCheck["_rowIndex"]; ok {
				if fidx, ok := idx.(float64); ok {
					rowIndex = int(fidx)
				}
			}
			
			updateQuery := `
				UPDATE bank_transactions 
				SET matched_check_id = ?, 
				    matched_dbf_row_index = ?,
				    match_confidence = ?,
				    match_type = ?,
				    is_matched = 1
				WHERE id = ?
			`
			
			_, err := s.db.Exec(updateQuery, checkID, rowIndex, bestMatch.Confidence, bestMatch.MatchType, txn.ID)
			if err == nil {
				newMatchCount++
			}
		}
	}
	
	// Update statement matched count
	updateStmt := `
		UPDATE bank_statements 
		SET matched_count = (
			SELECT COUNT(*) FROM bank_transactions 
			WHERE statement_id = ? AND is_matched = 1
		)
		WHERE id = ?
	`
	s.db.Exec(updateStmt, statementID, statementID)
	
	return map[string]interface{}{
		"status": "success",
		"newMatches": newMatchCount,
		"totalUnmatched": len(unmatchedTxns) - newMatchCount,
	}, nil
}

// GetMatchedTransactions retrieves all matched transactions
func (s *Service) GetMatchedTransactions(companyName, accountNumber string) (map[string]interface{}, error) {
	logger.WriteInfo("GetMatchedTransactions", fmt.Sprintf("Called for company: %s, account: %s", companyName, accountNumber))
	
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	// First get all matched bank transactions
	query := `
		SELECT bt.id, bt.matched_check_id, bt.matched_dbf_row_index, bt.match_confidence, 
			   bt.match_type, bt.manually_matched, bt.amount as bank_amount, bt.transaction_date as bank_date,
			   bt.description as bank_description, bt.check_number as bank_check_number
		FROM bank_transactions bt
		INNER JOIN bank_statements bs ON bt.statement_id = bs.id
		WHERE bt.company_name = ? AND bt.account_number = ? 
		  AND bs.is_active = 1
		  AND bt.is_matched = 1
	`
	
	rows, err := s.db.Query(query, companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to query matched transactions: %w", err)
	}
	defer rows.Close()
	
	// Map to store bank transaction matches by check ID
	matchedMap := make(map[string]map[string]interface{})
	
	for rows.Next() {
		var bankTxnID int
		var matchedCheckID, matchType, bankDescription, bankCheckNumber, bankDate sql.NullString
		var matchedDBFRowIndex sql.NullInt64
		var matchConfidence, bankAmount float64
		var manuallyMatched bool
		
		err := rows.Scan(
			&bankTxnID, &matchedCheckID, &matchedDBFRowIndex, &matchConfidence,
			&matchType, &manuallyMatched, &bankAmount, &bankDate,
			&bankDescription, &bankCheckNumber,
		)
		if err != nil {
			continue
		}
		
		if matchedCheckID.Valid && matchedCheckID.String != "" {
			matchedMap[matchedCheckID.String] = map[string]interface{}{
				"bank_txn_id":       bankTxnID,
				"match_confidence":  matchConfidence,
				"match_type":        matchType.String,
				"manually_matched":  manuallyMatched,
				"bank_amount":       bankAmount,
				"bank_date":         bankDate.String,
				"bank_description":  bankDescription.String,
				"bank_check_number": bankCheckNumber.String,
				"dbf_row_index":     matchedDBFRowIndex.Int64,
			}
		}
	}
	
	logger.WriteInfo("GetMatchedTransactions", fmt.Sprintf("Total matched bank transactions found: %d", len(matchedMap)))
	
	// Now read the checks from DBF and build response
	checksData, err := company.ReadDBFFile(companyName, "CHECKS.dbf", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read checks: %w", err)
	}
	
	columns, _ := checksData["columns"].([]string)
	rows2, _ := checksData["rows"].([][]interface{})
	
	// Find column indices
	var cidchecIdx, checkNoIdx, dateIdx, payeeIdx, amountIdx, acctIdx, clearedIdx int = -1, -1, -1, -1, -1, -1, -1
	for i, col := range columns {
		upperCol := strings.ToUpper(col)
		switch upperCol {
		case "CIDCHEC":
			cidchecIdx = i
		case "CCHECKNO":
			checkNoIdx = i
		case "DCHECKDATE":
			dateIdx = i
		case "CPAYEE":
			payeeIdx = i
		case "NAMOUNT":
			amountIdx = i
		case "CACCTNO":
			acctIdx = i
		case "LCLEARED":
			clearedIdx = i
		}
	}
	
	var matchedChecks []map[string]interface{}
	
	// Build response with check data as primary
	for rowIdx, row := range rows2 {
		if len(row) <= cidchecIdx || cidchecIdx < 0 {
			continue
		}
		
		checkID := fmt.Sprintf("%v", row[cidchecIdx])
		
		// Skip if this check isn't matched
		bankMatch, isMatched := matchedMap[checkID]
		if !isMatched {
			continue
		}
		
		// Skip if not for this account
		if acctIdx >= 0 && accountNumber != "" {
			checkAcct := fmt.Sprintf("%v", row[acctIdx])
			if checkAcct != accountNumber {
				continue
			}
		}
		
		// Build check data
		checkData := map[string]interface{}{
			"id":           checkID,
			"row_index":    rowIdx,
			"check_number": "",
			"check_date":   "",
			"payee":        "",
			"amount":       0.0,
			"account":      "",
			"cleared":      false,
		}
		
		if checkNoIdx >= 0 && checkNoIdx < len(row) {
			checkData["check_number"] = fmt.Sprintf("%v", row[checkNoIdx])
		}
		if dateIdx >= 0 && dateIdx < len(row) {
			checkData["check_date"] = fmt.Sprintf("%v", row[dateIdx])
		}
		if payeeIdx >= 0 && payeeIdx < len(row) {
			checkData["payee"] = fmt.Sprintf("%v", row[payeeIdx])
		}
		if amountIdx >= 0 && amountIdx < len(row) {
			checkData["amount"] = s.parseFloat(row[amountIdx])
		}
		if acctIdx >= 0 && acctIdx < len(row) {
			checkData["account"] = fmt.Sprintf("%v", row[acctIdx])
		}
		if clearedIdx >= 0 && clearedIdx < len(row) {
			if cleared, ok := row[clearedIdx].(bool); ok {
				checkData["cleared"] = cleared
			}
		}
		
		// Add bank match info
		for k, v := range bankMatch {
			checkData[k] = v
		}
		
		matchedChecks = append(matchedChecks, checkData)
	}
	
	logger.WriteInfo("GetMatchedTransactions", fmt.Sprintf("Returning %d matched checks", len(matchedChecks)))
	
	return map[string]interface{}{
		"status": "success",
		"checks": matchedChecks,
		"count":  len(matchedChecks),
	}, nil
}

// GetBankTransactions retrieves bank transactions for a specific account and optionally import batch
func (s *Service) GetBankTransactions(companyName string, accountNumber string, importBatchID string) (map[string]interface{}, error) {
	fmt.Printf("GetBankTransactions called for company: %s, account: %s, batch: %s\n", companyName, accountNumber, importBatchID)

	var query string
	var args []interface{}

	if importBatchID != "" {
		query = `
			SELECT bt.id, bt.company_name, bt.account_number, bt.statement_id, bt.transaction_date, bt.check_number,
				   bt.description, bt.amount, bt.transaction_type, bt.import_batch_id, bt.import_date,
				   bt.imported_by, bt.matched_check_id, bt.matched_dbf_row_index, bt.match_confidence, bt.match_type,
				   bt.is_matched, bt.manually_matched, bt.is_reconciled, bt.reconciled_date,
				   bt.reconciliation_id, bt.extended_data
			FROM bank_transactions bt
			WHERE bt.company_name = ? AND bt.account_number = ? AND bt.import_batch_id = ?
			ORDER BY bt.transaction_date, bt.id
		`
		args = []interface{}{companyName, accountNumber, importBatchID}
	} else {
		// Only show unmatched transactions from active statements
		query = `
			SELECT bt.id, bt.company_name, bt.account_number, bt.statement_id, bt.transaction_date, bt.check_number,
				   bt.description, bt.amount, bt.transaction_type, bt.import_batch_id, bt.import_date,
				   bt.imported_by, bt.matched_check_id, bt.matched_dbf_row_index, bt.match_confidence, bt.match_type,
				   bt.is_matched, bt.manually_matched, bt.is_reconciled, bt.reconciled_date,
				   bt.reconciliation_id, bt.extended_data
			FROM bank_transactions bt
			INNER JOIN bank_statements bs ON bt.statement_id = bs.id
			WHERE bt.company_name = ? AND bt.account_number = ? 
			  AND bs.is_active = TRUE
			  AND bt.is_matched = FALSE
			ORDER BY bt.import_date DESC, bt.transaction_date, bt.id
		`
		args = []interface{}{companyName, accountNumber}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query bank transactions: %w", err)
	}
	defer rows.Close()

	var transactions []BankTransaction
	for rows.Next() {
		var txn BankTransaction
		var extendedDataStr string
		var reconciledDate sql.NullString
		var reconciliationID sql.NullInt64
		var matchedCheckID sql.NullString
		var matchedDBFRowIndex sql.NullInt64

		err := rows.Scan(
			&txn.ID, &txn.CompanyName, &txn.AccountNumber, &txn.StatementID, &txn.TransactionDate,
			&txn.CheckNumber, &txn.Description, &txn.Amount, &txn.TransactionType,
			&txn.ImportBatchID, &txn.ImportDate, &txn.ImportedBy, &matchedCheckID,
			&matchedDBFRowIndex, &txn.MatchConfidence, &txn.MatchType, &txn.IsMatched, &txn.ManuallyMatched,
			&txn.IsReconciled, &reconciledDate, &reconciliationID, &extendedDataStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		// Handle nullable matched_check_id
		if matchedCheckID.Valid {
			txn.MatchedCheckID = matchedCheckID.String
		}
		if matchedDBFRowIndex.Valid {
			txn.MatchedDBFRowIndex = int(matchedDBFRowIndex.Int64)
		}

		// Debug first transaction
		if len(transactions) == 0 {
			fmt.Printf("DEBUG: First transaction date from DB: '%s'\n", txn.TransactionDate)
		}

		// Handle nullable fields
		if reconciledDate.Valid {
			txn.ReconciledDate = &reconciledDate.String
		}
		if reconciliationID.Valid {
			recID := int(reconciliationID.Int64)
			txn.ReconciliationID = &recID
		}

		// Parse extended data JSON
		if extendedDataStr != "" {
			json.Unmarshal([]byte(extendedDataStr), &txn.ExtendedData)
		}

		transactions = append(transactions, txn)
	}

	return map[string]interface{}{
		"status":       "success",
		"transactions": transactions,
		"count":        len(transactions),
	}, nil
}

// GetRecentStatements retrieves recent bank statements
func (s *Service) GetRecentStatements(companyName, accountNumber string) ([]map[string]interface{}, error) {
	// Implementation will be moved from main.go
	return nil, fmt.Errorf("not implemented")
}

// DeleteStatement deletes a bank statement and its transactions
func (s *Service) DeleteStatement(companyName, importBatchID string) error {
	fmt.Printf("DeleteStatement called for company: %s, batch: %s\n", companyName, importBatchID)

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete bank transactions first (due to foreign key)
	_, err = tx.Exec(`
		DELETE FROM bank_transactions 
		WHERE company_name = ? AND import_batch_id = ?
	`, companyName, importBatchID)
	if err != nil {
		return fmt.Errorf("failed to delete bank transactions: %w", err)
	}

	// Delete bank statement
	_, err = tx.Exec(`
		DELETE FROM bank_statements 
		WHERE company_name = ? AND import_batch_id = ?
	`, companyName, importBatchID)
	if err != nil {
		return fmt.Errorf("failed to delete bank statement: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("Successfully deleted bank statement batch: %s\n", importBatchID)
	return nil
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

// parseCSVToBankTransactions parses CSV content into BankTransaction objects
func (s *Service) parseCSVToBankTransactions(csvContent, companyName, accountNumber, batchID string) ([]BankTransaction, error) {
	lines := strings.Split(strings.TrimSpace(csvContent), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("CSV must contain header and at least one data row")
	}
	
	// Parse header to determine column indices
	header := parseCSVLine(lines[0])
	columnMap := make(map[string]int)
	
	for i, col := range header {
		colName := strings.ToLower(strings.TrimSpace(strings.Trim(col, `"`)))
		columnMap[colName] = i
		
		// Handle common variations
		switch colName {
		case "transaction date", "posting date", "trans date":
			columnMap["date"] = i
		case "payee", "merchant", "vendor", "memo":
			columnMap["description"] = i
		case "check #", "check number", "chk #":
			columnMap["check_number"] = i
		}
	}
	
	var transactions []BankTransaction
	
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		fields := parseCSVLine(line)
		if len(fields) < len(header) {
			continue // Skip malformed rows
		}
		
		// Clean fields
		for i, field := range fields {
			fields[i] = strings.TrimSpace(strings.Trim(field, `"`))
		}
		
		transaction := BankTransaction{
			CompanyName:   companyName,
			AccountNumber: accountNumber,
			ImportBatchID: batchID,
			ImportedBy:    "system", // Will be set by main.go wrapper
			ImportDate:    time.Now().Format("2006-01-02 15:04:05"),
		}
		
		// Extract date
		if dateIdx, exists := columnMap["date"]; exists && dateIdx < len(fields) {
			dateStr := strings.TrimSpace(fields[dateIdx])
			// Parse MM/DD and convert to 2025-MM-DD
			parts := strings.Split(dateStr, "/")
			if len(parts) == 2 {
				month, _ := strconv.Atoi(parts[0])
				day, _ := strconv.Atoi(parts[1])
				transaction.TransactionDate = fmt.Sprintf("2025-%02d-%02d", month, day)
			} else if len(parts) == 3 {
				// MM/DD/YYYY - convert to YYYY-MM-DD
				month, _ := strconv.Atoi(parts[0])
				day, _ := strconv.Atoi(parts[1])
				year, _ := strconv.Atoi(parts[2])
				transaction.TransactionDate = fmt.Sprintf("%04d-%02d-%02d", year, month, day)
			} else {
				transaction.TransactionDate = "2025-01-01"
			}
		}
		
		// Extract check number
		if checkIdx, exists := columnMap["check_number"]; exists && checkIdx < len(fields) {
			checkNum := strings.TrimSpace(fields[checkIdx])
			// Remove asterisks from check numbers
			checkNum = strings.ReplaceAll(checkNum, "*", "")
			transaction.CheckNumber = checkNum
		} else if checkIdx, exists := columnMap["check #"]; exists && checkIdx < len(fields) {
			checkNum := strings.TrimSpace(fields[checkIdx])
			checkNum = strings.ReplaceAll(checkNum, "*", "")
			transaction.CheckNumber = checkNum
		}
		
		// Extract description
		if descIdx, exists := columnMap["description"]; exists && descIdx < len(fields) {
			transaction.Description = fields[descIdx]
		}
		
		// Extract amount
		if amountIdx, exists := columnMap["amount"]; exists && amountIdx < len(fields) {
			amountStr := strings.ReplaceAll(fields[amountIdx], "$", "")
			amountStr = strings.ReplaceAll(amountStr, ",", "")
			if val, err := strconv.ParseFloat(amountStr, 64); err == nil {
				transaction.Amount = val
			}
		}
		
		// Extract type
		if typeIdx, exists := columnMap["type"]; exists && typeIdx < len(fields) {
			transaction.TransactionType = fields[typeIdx]
		} else {
			// Infer type from amount or other context
			if transaction.Amount < 0 {
				transaction.TransactionType = "Debit"
			} else if transaction.CheckNumber != "" {
				transaction.TransactionType = "Check"
			} else {
				transaction.TransactionType = "Deposit"
			}
		}
		
		transactions = append(transactions, transaction)
	}
	
	logger.WriteInfo("parseCSVToBankTransactions", fmt.Sprintf("Parsed %d bank transactions from CSV", len(transactions)))
	return transactions, nil
}

// parseCSVLine handles quoted CSV fields properly
func parseCSVLine(line string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false
	
	for i := 0; i < len(line); i++ {
		ch := line[i]
		
		if ch == '"' {
			inQuotes = !inQuotes
		} else if ch == ',' && !inQuotes {
			result = append(result, current.String())
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	
	// Add the last field
	result = append(result, current.String())
	
	return result
}

// createBankStatement creates a bank statement record for tracking import sessions
func (s *Service) createBankStatement(companyName, accountNumber, statementDate, batchID string, transactionCount int) (int, error) {
	if s.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}
	
	query := `
		INSERT INTO bank_statements (
			company_name, account_number, statement_date, import_batch_id,
			imported_by, transaction_count, is_active
		) VALUES (?, ?, ?, ?, ?, ?, 1)
	`
	
	result, err := s.db.Exec(query, companyName, accountNumber, statementDate, batchID, 
		"system", transactionCount)
	if err != nil {
		return 0, fmt.Errorf("failed to insert bank statement: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get statement ID: %w", err)
	}
	
	return int(id), nil
}

// storeBankTransactions stores bank transactions in SQLite
func (s *Service) storeBankTransactions(transactions []BankTransaction) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	// Begin transaction for atomic insert
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	stmt, err := tx.Prepare(`
		INSERT INTO bank_transactions (
			company_name, account_number, statement_id, transaction_date, check_number, description,
			amount, transaction_type, import_batch_id, imported_by, matched_check_id,
			matched_dbf_row_index, match_confidence, match_type, is_matched, manually_matched
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	for _, txn := range transactions {
		_, err = stmt.Exec(
			txn.CompanyName, txn.AccountNumber, txn.StatementID, txn.TransactionDate, txn.CheckNumber,
			txn.Description, txn.Amount, txn.TransactionType, txn.ImportBatchID,
			txn.ImportedBy, txn.MatchedCheckID, txn.MatchedDBFRowIndex, txn.MatchConfidence, txn.MatchType,
			txn.IsMatched, false, // manually_matched defaults to false
		)
		if err != nil {
			return fmt.Errorf("failed to insert transaction: %w", err)
		}
	}
	
	return tx.Commit()
}

// AutoMatchBankTransactions performs automatic matching of bank transactions with existing checks
func (s *Service) AutoMatchBankTransactions(bankTransactions []BankTransaction, existingChecks []map[string]interface{}) []MatchResult {
	var matches []MatchResult

	// Keep track of already matched check IDs to prevent double-matching
	matchedCheckIDs := make(map[string]bool)

	// Sort bank transactions by date to match older transactions first
	// This helps with recurring transactions
	sort.Slice(bankTransactions, func(i, j int) bool {
		dateI, _ := s.parseDate(bankTransactions[i].TransactionDate)
		dateJ, _ := s.parseDate(bankTransactions[j].TransactionDate)
		return dateI.Before(dateJ)
	})

	for i := range bankTransactions {
		txn := &bankTransactions[i]

		// Skip only deposits, not checks (checks may have positive amounts in bank statements)
		if txn.TransactionType == "Deposit" {
			continue
		}

		// Filter out already matched checks
		availableChecks := []map[string]interface{}{}
		for _, check := range existingChecks {
			if checkID, ok := check["id"].(string); ok {
				if !matchedCheckIDs[checkID] {
					availableChecks = append(availableChecks, check)
				}
			}
		}

		bestMatch := s.findBestCheckMatchForBankTxn(txn, availableChecks)
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

// findBestCheckMatchForBankTxn finds the best matching check for a bank transaction
func (s *Service) findBestCheckMatchForBankTxn(txn *BankTransaction, existingChecks []map[string]interface{}) *MatchResult {
	var bestMatch *MatchResult
	var highestScore float64 = 0

	for _, check := range existingChecks {
		score := s.calculateBankTxnMatchScore(txn, check)
		if score > highestScore {
			highestScore = score
			bestMatch = &MatchResult{
				BankTransaction: *txn,
				MatchedCheck:    check,
				Confidence:      score,
				MatchType:       s.determineBankTxnMatchType(score, txn, check),
			}
		}
	}

	return bestMatch
}

// calculateBankTxnMatchScore calculates the match score between a bank transaction and a check
func (s *Service) calculateBankTxnMatchScore(txn *BankTransaction, check map[string]interface{}) float64 {
	var score float64 = 0
	var totalWeight float64 = 0

	// Amount matching (weight: 40%)
	amountWeight := 0.4
	totalWeight += amountWeight
	
	checkAmount := s.parseAmount(check["amount"])
	txnAmount := math.Abs(txn.Amount)
	
	if math.Abs(checkAmount-txnAmount) < 0.01 {
		score += amountWeight
	} else if math.Abs(checkAmount-txnAmount) < 1.00 {
		score += amountWeight * 0.5
	}

	// Date proximity matching (weight: 40%)
	dateWeight := 0.4
	totalWeight += dateWeight
	
	if checkDateStr, ok := check["date"].(string); ok {
		checkDate, err1 := s.parseDate(checkDateStr)
		txnDate, err2 := s.parseDate(txn.TransactionDate)
		
		if err1 == nil && err2 == nil {
			daysDiff := math.Abs(checkDate.Sub(txnDate).Hours() / 24)
			
			if daysDiff == 0 {
				score += dateWeight
			} else if daysDiff <= 3 {
				score += dateWeight * 0.8
			} else if daysDiff <= 7 {
				score += dateWeight * 0.6
			} else if daysDiff <= 14 {
				score += dateWeight * 0.4
			} else if daysDiff <= 30 {
				score += dateWeight * 0.2
			}
		}
	}

	// Check number matching (weight: 20% if available)
	if txn.CheckNumber != "" {
		checkNumWeight := 0.2
		totalWeight += checkNumWeight
		
		if checkNum, ok := check["checkNumber"].(string); ok && checkNum == txn.CheckNumber {
			score += checkNumWeight
		}
	}

	return score / totalWeight
}

// determineBankTxnMatchType determines the type of match based on score and attributes
func (s *Service) determineBankTxnMatchType(score float64, txn *BankTransaction, check map[string]interface{}) string {
	if score >= 0.95 {
		return "exact"
	}
	
	if score >= 0.8 {
		// Check if it's exact amount but different date
		checkAmount := s.parseAmount(check["amount"])
		txnAmount := math.Abs(txn.Amount)
		
		if math.Abs(checkAmount-txnAmount) < 0.01 {
			return "exact_amount"
		}
		return "high_confidence"
	}
	
	if score >= 0.6 {
		return "probable"
	}
	
	if score >= 0.5 {
		return "possible"
	}
	
	return "fuzzy"
}

// Helper method to parse date from various formats
func (s *Service) parseDate(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case string:
		// Try various date formats
		formats := []string{
			"2006-01-02",
			"01/02/2006",
			"01/02/06",
			"2006-01-02 15:04:05",
		}
		
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("unable to parse date: %s", v)
	case time.Time:
		return v, nil
	default:
		return time.Time{}, fmt.Errorf("unsupported date type: %T", value)
	}
}

// Helper method to parse amount from various formats
func (s *Service) parseAmount(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		// Remove currency symbols and parse
		cleaned := strings.ReplaceAll(v, "$", "")
		cleaned = strings.ReplaceAll(cleaned, ",", "")
		if val, err := strconv.ParseFloat(cleaned, 64); err == nil {
			return val
		}
	}
	return 0
}