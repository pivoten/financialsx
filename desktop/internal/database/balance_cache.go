package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/company"
)

type CachedBalance struct {
	ID                    int       `json:"id" db:"id"`
	CompanyName           string    `json:"company_name" db:"company_name"`
	AccountNumber         string    `json:"account_number" db:"account_number"`
	AccountName           string    `json:"account_name" db:"account_name"`
	AccountType           int       `json:"account_type" db:"account_type"`
	GLBalance             float64   `json:"gl_balance" db:"gl_balance"`
	GLLastUpdated         time.Time `json:"gl_last_updated" db:"gl_last_updated"`
	GLRecordCount         int       `json:"gl_record_count" db:"gl_record_count"`
	OutstandingTotal      float64   `json:"outstanding_checks_total" db:"outstanding_checks_total"`
	OutstandingCount      int       `json:"outstanding_checks_count" db:"outstanding_checks_count"`
	OutstandingLastUpdated time.Time `json:"outstanding_checks_last_updated" db:"outstanding_checks_last_updated"`
	BankBalance           float64   `json:"bank_balance" db:"bank_balance"`
	IsActive              bool      `json:"is_active" db:"is_active"`
	IsBankAccount         bool      `json:"is_bank_account" db:"is_bank_account"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
	Metadata              string    `json:"metadata" db:"metadata"`
	
	// From the view
	GLAgeHours      float64 `json:"gl_age_hours" db:"gl_age_hours"`
	ChecksAgeHours  float64 `json:"checks_age_hours" db:"checks_age_hours"`
	GLFreshness     string  `json:"gl_freshness" db:"gl_freshness"`
	ChecksFreshness string  `json:"checks_freshness" db:"checks_freshness"`
}

type BalanceHistory struct {
	ID                    int       `json:"id" db:"id"`
	AccountBalanceID      int       `json:"account_balance_id" db:"account_balance_id"`
	CompanyName           string    `json:"company_name" db:"company_name"`
	AccountNumber         string    `json:"account_number" db:"account_number"`
	ChangeType            string    `json:"change_type" db:"change_type"`
	OldGLBalance          *float64  `json:"old_gl_balance" db:"old_gl_balance"`
	NewGLBalance          *float64  `json:"new_gl_balance" db:"new_gl_balance"`
	OldOutstandingTotal   *float64  `json:"old_outstanding_total" db:"old_outstanding_total"`
	NewOutstandingTotal   *float64  `json:"new_outstanding_total" db:"new_outstanding_total"`
	OldAvailableBalance   *float64  `json:"old_available_balance" db:"old_available_balance"`
	NewAvailableBalance   *float64  `json:"new_available_balance" db:"new_available_balance"`
	ChangeReason          string    `json:"change_reason" db:"change_reason"`
	ChangedBy             string    `json:"changed_by" db:"changed_by"`
	ChangeTimestamp       time.Time `json:"change_timestamp" db:"change_timestamp"`
	Metadata              string    `json:"metadata" db:"metadata"`
}

// InitializeBalanceCache creates the balance cache tables
func InitializeBalanceCache(db *DB) error {
	// Read and execute the schema SQL
	schemaSQL := `
	-- Bank Account Balance Caching System
	CREATE TABLE IF NOT EXISTS account_balances (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		company_name TEXT NOT NULL,
		account_number TEXT NOT NULL,
		account_name TEXT NOT NULL,
		account_type INTEGER NOT NULL,
		gl_balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
		gl_last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		gl_record_count INTEGER NOT NULL DEFAULT 0,
		outstanding_checks_total DECIMAL(15,2) NOT NULL DEFAULT 0.00,
		outstanding_checks_count INTEGER NOT NULL DEFAULT 0,
		outstanding_checks_last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		bank_balance DECIMAL(15,2) GENERATED ALWAYS AS (gl_balance + outstanding_checks_total) STORED,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		is_bank_account BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		metadata JSON DEFAULT '{}',
		UNIQUE(company_name, account_number)
	);
	
	CREATE TABLE IF NOT EXISTS balance_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		account_balance_id INTEGER NOT NULL,
		company_name TEXT NOT NULL,
		account_number TEXT NOT NULL,
		change_type TEXT NOT NULL CHECK (change_type IN ('gl_refresh', 'checks_refresh', 'manual_adjustment', 'reconciliation')),
		old_gl_balance DECIMAL(15,2),
		new_gl_balance DECIMAL(15,2),
		old_outstanding_total DECIMAL(15,2),
		new_outstanding_total DECIMAL(15,2),
		old_available_balance DECIMAL(15,2),
		new_available_balance DECIMAL(15,2),
		change_reason TEXT,
		changed_by TEXT,
		change_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		metadata JSON DEFAULT '{}',
		FOREIGN KEY (account_balance_id) REFERENCES account_balances(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_account_balances_company_account ON account_balances(company_name, account_number);
	CREATE INDEX IF NOT EXISTS idx_account_balances_company_active ON account_balances(company_name, is_active, is_bank_account);
	CREATE INDEX IF NOT EXISTS idx_balance_history_account_timestamp ON balance_history(account_balance_id, change_timestamp);
	
	CREATE TRIGGER IF NOT EXISTS update_account_balances_timestamp 
		AFTER UPDATE ON account_balances
		FOR EACH ROW
	BEGIN
		UPDATE account_balances SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;
	
	CREATE VIEW IF NOT EXISTS account_balance_summary AS
	SELECT 
		ab.*,
		ROUND((julianday('now') - julianday(ab.gl_last_updated)) * 24, 2) as gl_age_hours,
		ROUND((julianday('now') - julianday(ab.outstanding_checks_last_updated)) * 24, 2) as checks_age_hours,
		CASE 
			WHEN (julianday('now') - julianday(ab.gl_last_updated)) * 24 > 24 THEN 'stale'
			WHEN (julianday('now') - julianday(ab.gl_last_updated)) * 24 > 4 THEN 'aging'
			ELSE 'fresh'
		END as gl_freshness,
		CASE 
			WHEN (julianday('now') - julianday(ab.outstanding_checks_last_updated)) * 24 > 4 THEN 'stale'
			WHEN (julianday('now') - julianday(ab.outstanding_checks_last_updated)) * 24 > 1 THEN 'aging'
			ELSE 'fresh'
		END as checks_freshness
	FROM account_balances ab
	WHERE ab.is_active = TRUE;
	`
	
	_, err := db.Exec(schemaSQL)
	return err
}

// GetCachedBalance retrieves the cached balance for an account
func GetCachedBalance(db *DB, companyName, accountNumber string) (*CachedBalance, error) {
	query := `
		SELECT * FROM account_balance_summary 
		WHERE company_name = ? AND account_number = ? AND is_active = TRUE
	`
	
	var balance CachedBalance
	err := db.QueryRow(query, companyName, accountNumber).Scan(
		&balance.ID, &balance.CompanyName, &balance.AccountNumber, &balance.AccountName,
		&balance.AccountType, &balance.GLBalance, &balance.GLLastUpdated, &balance.GLRecordCount,
		&balance.OutstandingTotal, &balance.OutstandingCount, &balance.OutstandingLastUpdated,
		&balance.BankBalance, &balance.IsActive, &balance.IsBankAccount,
		&balance.CreatedAt, &balance.UpdatedAt, &balance.Metadata,
		&balance.GLAgeHours, &balance.ChecksAgeHours, &balance.GLFreshness, &balance.ChecksFreshness,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil // No cached balance found
	}
	
	return &balance, err
}

// GetAllCachedBalances retrieves all cached balances for a company
func GetAllCachedBalances(db *DB, companyName string) ([]CachedBalance, error) {
	query := `
		SELECT * FROM account_balance_summary 
		WHERE company_name = ? AND is_active = TRUE AND is_bank_account = TRUE
		ORDER BY account_number
	`
	
	rows, err := db.Query(query, companyName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var balances []CachedBalance
	for rows.Next() {
		var balance CachedBalance
		err := rows.Scan(
			&balance.ID, &balance.CompanyName, &balance.AccountNumber, &balance.AccountName,
			&balance.AccountType, &balance.GLBalance, &balance.GLLastUpdated, &balance.GLRecordCount,
			&balance.OutstandingTotal, &balance.OutstandingCount, &balance.OutstandingLastUpdated,
			&balance.BankBalance, &balance.IsActive, &balance.IsBankAccount,
			&balance.CreatedAt, &balance.UpdatedAt, &balance.Metadata,
			&balance.GLAgeHours, &balance.ChecksAgeHours, &balance.GLFreshness, &balance.ChecksFreshness,
		)
		if err != nil {
			return nil, err
		}
		balances = append(balances, balance)
	}
	
	return balances, nil
}

// RefreshGLBalance updates the GL balance by scanning GLMASTER.dbf
func RefreshGLBalance(db *DB, companyName, accountNumber, username string) error {
	// Get current cached balance
	currentBalance, err := GetCachedBalance(db, companyName, accountNumber)
	
	// Calculate new GL balance (existing logic from GetAccountBalance)
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.dbf", "", 0, 50000, "", "")
	if err != nil {
		return fmt.Errorf("failed to read GLMASTER.dbf: %w", err)
	}
	
	// Get column indices for GLMASTER.dbf
	glColumns, ok := glData["columns"].([]string)
	if !ok {
		return fmt.Errorf("invalid GLMASTER.dbf structure")
	}
	
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
		return fmt.Errorf("required GL columns not found")
	}
	
	// Process GL entries
	var totalBalance float64
	var recordCount int
	glRows, _ := glData["rows"].([][]interface{})
	
	for _, row := range glRows {
		if len(row) <= accountIdx {
			continue
		}
		
		rowAccount := strings.TrimSpace(fmt.Sprintf("%v", row[accountIdx]))
		if rowAccount == accountNumber {
			recordCount++
			
			// Add debits
			if debitIdx != -1 && len(row) > debitIdx && row[debitIdx] != nil {
				if debit := parseFloat(row[debitIdx]); debit != 0 {
					totalBalance += debit
				}
			}
			
			// Subtract credits
			if creditIdx != -1 && len(row) > creditIdx && row[creditIdx] != nil {
				if credit := parseFloat(row[creditIdx]); credit != 0 {
					totalBalance -= credit
				}
			}
		}
	}
	
	// Update the cached balance
	if currentBalance == nil {
		// Insert new record
		_, err = db.Exec(`
			INSERT INTO account_balances 
			(company_name, account_number, account_name, account_type, 
			 gl_balance, gl_record_count, gl_last_updated)
			VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		`, companyName, accountNumber, "", 1, totalBalance, recordCount)
	} else {
		// Update existing record
		_, err = db.Exec(`
			UPDATE account_balances 
			SET gl_balance = ?, gl_record_count = ?, gl_last_updated = CURRENT_TIMESTAMP
			WHERE company_name = ? AND account_number = ?
		`, totalBalance, recordCount, companyName, accountNumber)
		
		// Record the change in history
		if err == nil {
			_, err = db.Exec(`
				INSERT INTO balance_history 
				(account_balance_id, company_name, account_number, change_type,
				 old_gl_balance, new_gl_balance, old_available_balance, new_available_balance,
				 change_reason, changed_by)
				VALUES (?, ?, ?, 'gl_refresh', ?, ?, ?, ?, 'GL balance refresh', ?)
			`, currentBalance.ID, companyName, accountNumber, 
				currentBalance.GLBalance, totalBalance,
				currentBalance.BankBalance, totalBalance+currentBalance.OutstandingTotal,
				username)
		}
	}
	
	return err
}

// RefreshOutstandingChecks updates the outstanding checks total
func RefreshOutstandingChecks(db *DB, companyName, accountNumber, username string) error {
	// Calculate outstanding checks total by reading CHECKS.dbf directly
	checksData, err := company.ReadDBFFile(companyName, "checks.dbf", "", 0, 10000, "", "")
	if err != nil {
		return fmt.Errorf("failed to read checks.dbf: %w", err)
	}
	
	// Get column indices for checks.dbf
	checksColumns, ok := checksData["columns"].([]string)
	if !ok {
		return fmt.Errorf("invalid checks.dbf structure")
	}
	
	// Find relevant check columns
	var checkNumIdx, amountIdx, accountIdx, clearedIdx, voidIdx int = -1, -1, -1, -1, -1
	for i, col := range checksColumns {
		colUpper := strings.ToUpper(col)
		if colUpper == "CCHECKNO" {
			checkNumIdx = i
		} else if colUpper == "NAMOUNT" {
			amountIdx = i
		} else if colUpper == "CACCTNO" {
			accountIdx = i
		} else if colUpper == "LCLEARED" {
			clearedIdx = i
		} else if colUpper == "LVOID" {
			voidIdx = i
		}
	}
	
	if checkNumIdx == -1 || amountIdx == -1 {
		return fmt.Errorf("required columns not found in checks.dbf")
	}
	
	// Process check rows to find outstanding checks
	var totalOutstanding float64
	var checkCount int
	checksRows, _ := checksData["rows"].([][]interface{})
	
	fmt.Printf("RefreshOutstandingChecks: Processing %d rows for account %s\n", len(checksRows), accountNumber)
	
	// Debug: Show first few rows to see what data we're actually reading
	for i := 0; i < 5 && i < len(checksRows); i++ {
		row := checksRows[i]
		if len(row) > accountIdx && len(row) > checkNumIdx && len(row) > amountIdx {
			checkNum := fmt.Sprintf("%v", row[checkNumIdx])
			checkAccount := ""
			if accountIdx != -1 && len(row) > accountIdx && row[accountIdx] != nil {
				checkAccount = strings.TrimSpace(fmt.Sprintf("%v", row[accountIdx]))
			}
			amount := parseFloat(row[amountIdx])
			fmt.Printf("RefreshOutstandingChecks: Sample row %d - Check: %s, Account: %s, Amount: $%.2f\n", 
				i+1, checkNum, checkAccount, amount)
		}
	}
	
	for _, row := range checksRows {
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
		
		// Check if cleared (default to false if no cleared column)
		isCleared := false
		if clearedIdx != -1 && len(row) > clearedIdx {
			clearedValue := row[clearedIdx]
			if clearedValue != nil {
				switch v := clearedValue.(type) {
				case bool:
					isCleared = v
				case string:
					lowerVal := strings.ToLower(strings.TrimSpace(v))
					isCleared = (lowerVal == "t" || lowerVal == ".t." || lowerVal == "true" || lowerVal == "1")
				}
			}
		}
		
		// Check if voided (default to false if no void column)
		isVoided := false
		if voidIdx != -1 && len(row) > voidIdx {
			voidValue := row[voidIdx]
			if voidValue != nil {
				switch v := voidValue.(type) {
				case bool:
					isVoided = v
				case string:
					lowerVal := strings.ToLower(strings.TrimSpace(v))
					isVoided = (lowerVal == "t" || lowerVal == ".t." || lowerVal == "true" || lowerVal == "1")
				}
			}
		}
		
		// Only include if not cleared and not voided
		if !isCleared && !isVoided {
			amount := parseFloat(row[amountIdx])
			totalOutstanding += amount
			checkCount++
			
			// Debug logging for first few outstanding checks
			if checkCount <= 5 {
				checkNum := fmt.Sprintf("%v", row[checkNumIdx])
				fmt.Printf("RefreshOutstandingChecks: Outstanding check #%d: %s, Amount: $%.2f, Account: %s\n", 
					checkCount, checkNum, amount, checkAccount)
			}
		} else {
			// Debug logging for first few cleared/voided checks
			if len(fmt.Sprintf("%v", row[checkNumIdx])) > 0 && (checkCount + 1) <= 10 {
				checkNum := fmt.Sprintf("%v", row[checkNumIdx])
				amount := parseFloat(row[amountIdx])
				fmt.Printf("RefreshOutstandingChecks: Skipped check %s (Amount: $%.2f, Cleared: %t, Voided: %t, Account: %s)\n", 
					checkNum, amount, isCleared, isVoided, checkAccount)
			}
		}
	}
	
	fmt.Printf("RefreshOutstandingChecks: Final totals for account %s: %d checks, $%.2f total outstanding\n", 
		accountNumber, checkCount, totalOutstanding)
	
	// Get current cached balance
	currentBalance, err := GetCachedBalance(db, companyName, accountNumber)
	if err != nil {
		return err
	}
	
	if currentBalance == nil {
		// Insert new record with outstanding checks only
		_, err = db.Exec(`
			INSERT INTO account_balances 
			(company_name, account_number, account_name, account_type, 
			 outstanding_checks_total, outstanding_checks_count, outstanding_checks_last_updated)
			VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		`, companyName, accountNumber, "", 1, totalOutstanding, checkCount)
	} else {
		// Update existing record
		_, err = db.Exec(`
			UPDATE account_balances 
			SET outstanding_checks_total = ?, outstanding_checks_count = ?, 
			    outstanding_checks_last_updated = CURRENT_TIMESTAMP
			WHERE company_name = ? AND account_number = ?
		`, totalOutstanding, checkCount, companyName, accountNumber)
		
		// Record the change in history
		if err == nil {
			_, err = db.Exec(`
				INSERT INTO balance_history 
				(account_balance_id, company_name, account_number, change_type,
				 old_outstanding_total, new_outstanding_total, 
				 old_available_balance, new_available_balance,
				 change_reason, changed_by)
				VALUES (?, ?, ?, 'checks_refresh', ?, ?, ?, ?, 'Outstanding checks refresh', ?)
			`, currentBalance.ID, companyName, accountNumber,
				currentBalance.OutstandingTotal, totalOutstanding,
				currentBalance.BankBalance, currentBalance.GLBalance+totalOutstanding,
				username)
		}
	}
	
	return err
}

// Helper function to parse float values
func parseFloat(value interface{}) float64 {
	if value == nil {
		return 0
	}
	
	switch v := value.(type) {
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
	
	return 0
}