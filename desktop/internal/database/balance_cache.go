package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/currency"
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
	
	// New fields for detailed breakdown (calculated from metadata or separate columns)
	UnclearedDeposits   float64 `json:"uncleared_deposits"`
	UnclearedChecks     float64 `json:"uncleared_checks"`
	DepositCount        int     `json:"deposit_count"`
	CheckCount          int     `json:"check_count"`
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
		bank_balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		is_bank_account BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT DEFAULT '{}',
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
		ab.id,
		ab.company_name,
		ab.account_number,
		ab.account_name,
		ab.account_type,
		ab.gl_balance,
		ab.gl_last_updated,
		ab.gl_record_count,
		ab.outstanding_checks_total,
		ab.outstanding_checks_count,
		ab.outstanding_checks_last_updated,
		(ab.gl_balance + ab.outstanding_checks_total) as bank_balance,
		ab.is_active,
		ab.is_bank_account,
		ab.created_at,
		ab.updated_at,
		ab.metadata,
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
	
	if err != nil {
		return nil, err
	}
	
	// Parse metadata JSON to populate detailed breakdown fields
	if balance.Metadata != "" {
		var metadataMap map[string]interface{}
		if err := json.Unmarshal([]byte(balance.Metadata), &metadataMap); err == nil {
			if val, ok := metadataMap["uncleared_deposits"].(float64); ok {
				balance.UnclearedDeposits = val
			}
			if val, ok := metadataMap["uncleared_checks"].(float64); ok {
				balance.UnclearedChecks = val
			}
			if val, ok := metadataMap["deposit_count"].(float64); ok {
				balance.DepositCount = int(val)
			}
			if val, ok := metadataMap["check_count"].(float64); ok {
				balance.CheckCount = int(val)
			}
		}
	}
	
	return &balance, nil
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
		
		// Parse metadata JSON to populate detailed breakdown fields
		if balance.Metadata != "" {
			var metadataMap map[string]interface{}
			if err := json.Unmarshal([]byte(balance.Metadata), &metadataMap); err == nil {
				if val, ok := metadataMap["uncleared_deposits"].(float64); ok {
					balance.UnclearedDeposits = val
				}
				if val, ok := metadataMap["uncleared_checks"].(float64); ok {
					balance.UnclearedChecks = val
				}
				if val, ok := metadataMap["deposit_count"].(float64); ok {
					balance.DepositCount = int(val)
				}
				if val, ok := metadataMap["check_count"].(float64); ok {
					balance.CheckCount = int(val)
				}
			}
		}
		
		balances = append(balances, balance)
	}
	
	return balances, nil
}

// RefreshGLBalance updates the GL balance by scanning GLMASTER.dbf
func RefreshGLBalance(db *DB, companyName, accountNumber, username string) error {
	fmt.Printf("RefreshGLBalance: Starting for account %s in company %s\n", accountNumber, companyName)
	
	// Get current cached balance
	currentBalance, err := GetCachedBalance(db, companyName, accountNumber)
	if currentBalance != nil {
		fmt.Printf("RefreshGLBalance: Found existing balance - GL: %.2f, Outstanding: %.2f\n", 
			currentBalance.GLBalance, currentBalance.OutstandingTotal)
	} else {
		fmt.Printf("RefreshGLBalance: No existing balance found, will create new\n")
	}
	
	// First, get the account type from COA.dbf
	fmt.Printf("RefreshGLBalance: Reading COA.dbf to get account type...\n")
	// Read ALL COA records to ensure we find the account
	coaData, err := company.ReadDBFFile(companyName, "COA.dbf", "", 0, 0, "", "")
	if err != nil {
		fmt.Printf("RefreshGLBalance: ERROR reading COA.dbf: %v\n", err)
		return fmt.Errorf("failed to read COA.dbf: %w", err)
	}
	
	// Get column indices for COA.dbf
	coaColumns, ok := coaData["columns"].([]string)
	if !ok {
		return fmt.Errorf("invalid COA.dbf structure")
	}
	
	var coaAccountIdx, accountTypeIdx int = -1, -1
	for i, col := range coaColumns {
		colUpper := strings.ToUpper(col)
		if colUpper == "CACCTNO" {
			coaAccountIdx = i
		} else if colUpper == "NACCTTYPE" {
			accountTypeIdx = i
		}
	}
	
	if coaAccountIdx == -1 || accountTypeIdx == -1 {
		return fmt.Errorf("required COA columns not found")
	}
	
	// Find the account type
	var accountType int = 1 // Default to asset if not found
	coaRows, _ := coaData["rows"].([][]interface{})
	for _, row := range coaRows {
		if len(row) > coaAccountIdx && len(row) > accountTypeIdx {
			rowAccount := strings.TrimSpace(fmt.Sprintf("%v", row[coaAccountIdx]))
			if rowAccount == accountNumber {
				if typeVal := parseFloat(row[accountTypeIdx]); typeVal != 0 {
					accountType = int(typeVal)
					fmt.Printf("RefreshGLBalance: Account %s has type %d\n", accountNumber, accountType)
					break
				}
			}
		}
	}
	
	// Calculate new GL balance
	fmt.Printf("RefreshGLBalance: Reading GLMASTER.dbf...\n")
	// IMPORTANT: Read ALL records (0, 0) to ensure we don't miss any transactions
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.dbf", "", 0, 0, "", "")
	if err != nil {
		fmt.Printf("RefreshGLBalance: ERROR reading GLMASTER.dbf: %v\n", err)
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
	
	// Process GL entries with account type-specific logic using decimal arithmetic
	totalDebits := currency.Zero()
	totalCredits := currency.Zero()
	var recordCount int
	glRows, _ := glData["rows"].([][]interface{})
	
	for _, row := range glRows {
		if len(row) <= accountIdx {
			continue
		}
		
		rowAccount := strings.TrimSpace(fmt.Sprintf("%v", row[accountIdx]))
		if rowAccount == accountNumber {
			recordCount++
			
			// Sum debits using decimal arithmetic
			if debitIdx != -1 && len(row) > debitIdx && row[debitIdx] != nil {
				debit := currency.ParseFromDBF(row[debitIdx])
				if !debit.IsZero() {
					totalDebits = totalDebits.Add(debit)
				}
			}
			
			// Sum credits using decimal arithmetic
			if creditIdx != -1 && len(row) > creditIdx && row[creditIdx] != nil {
				credit := currency.ParseFromDBF(row[creditIdx])
				if !credit.IsZero() {
					totalCredits = totalCredits.Add(credit)
				}
			}
		}
	}
	
	// Apply correct formula based on account type using decimal arithmetic
	// Account Types (standard):
	// 1 = Assets (Debit normal balance)
	// 2 = Liabilities (Credit normal balance)
	// 3 = Equity (Credit normal balance)
	// 4 = Revenue/Income (Credit normal balance)
	// 5 = Expenses (Debit normal balance)
	var totalBalance currency.Currency
	switch accountType {
	case 1, 5: // Assets and Expenses (Debit normal balance)
		totalBalance = totalDebits.Sub(totalCredits)
		fmt.Printf("RefreshGLBalance: Asset/Expense account - Debits: %s, Credits: %s, Balance: %s\n", 
			totalDebits.ToString(), totalCredits.ToString(), totalBalance.ToString())
	case 2, 3, 4: // Liabilities, Equity, Revenue (Credit normal balance)
		totalBalance = totalCredits.Sub(totalDebits)
		fmt.Printf("RefreshGLBalance: Liability/Equity/Revenue account - Credits: %s, Debits: %s, Balance: %s\n", 
			totalCredits.ToString(), totalDebits.ToString(), totalBalance.ToString())
	default:
		// Default to asset behavior if unknown type
		totalBalance = totalDebits.Sub(totalCredits)
		fmt.Printf("RefreshGLBalance: Unknown account type %d, using asset formula - Balance: %s\n", 
			accountType, totalBalance.ToString())
	}
	
	fmt.Printf("RefreshGLBalance: Processed %d records, total balance: %s\n", recordCount, totalBalance.ToString())
	
	// Update the cached balance - use UPSERT to handle race conditions
	// Note: bank_balance is GENERATED and will auto-calculate as gl_balance + outstanding_checks_total
	// Convert Currency to float64 for database storage (SQLite stores as REAL)
	// Store as string to preserve decimal precision
	_, err = db.Exec(`
		INSERT INTO account_balances 
		(company_name, account_number, account_name, account_type, 
		 gl_balance, gl_record_count, gl_last_updated)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(company_name, account_number) 
		DO UPDATE SET 
			account_type = excluded.account_type,
			gl_balance = excluded.gl_balance, 
			gl_record_count = excluded.gl_record_count, 
			gl_last_updated = CURRENT_TIMESTAMP
	`, companyName, accountNumber, "", accountType, totalBalance.ToString(), recordCount)
	
	if err != nil {
		fmt.Printf("RefreshGLBalance: UPSERT error: %v, trying fallback UPDATE\n", err)
		// Fallback to simple UPDATE if INSERT OR REPLACE is not working
		_, err = db.Exec(`
			UPDATE account_balances 
			SET account_type = ?, gl_balance = ?, gl_record_count = ?, gl_last_updated = CURRENT_TIMESTAMP
			WHERE company_name = ? AND account_number = ?
		`, accountType, totalBalance.ToString(), recordCount, companyName, accountNumber)
		
		// Record the change in history
		if err == nil {
			// Calculate new bank balance using Currency type to avoid precision loss
			newBankBalance := totalBalance.Add(currency.NewFromFloat(currentBalance.OutstandingTotal))
			_, err = db.Exec(`
				INSERT INTO balance_history 
				(account_balance_id, company_name, account_number, change_type,
				 old_gl_balance, new_gl_balance, old_available_balance, new_available_balance,
				 change_reason, changed_by)
				VALUES (?, ?, ?, 'gl_refresh', ?, ?, ?, ?, 'GL balance refresh', ?)
			`, currentBalance.ID, companyName, accountNumber, 
				currentBalance.GLBalance, totalBalance.ToString(),
				currentBalance.BankBalance, newBankBalance.ToString(),
				username)
		}
	}
	
	return err
}

// RefreshOutstandingChecks updates the outstanding checks/deposits total for proper bank reconciliation
// Bank Reconciliation Formula: GL Balance + Uncleared Deposits - Uncleared Checks = Bank Balance
func RefreshOutstandingChecks(db *DB, companyName, accountNumber, username string) error {
	// Read CHECKS.dbf which contains both checks (CENTRYTYPE=C) and deposits (CENTRYTYPE=D)
	// IMPORTANT: Read ALL records (0, 0) to ensure we don't miss any checks/deposits
	checksData, err := company.ReadDBFFile(companyName, "checks.dbf", "", 0, 0, "", "")
	if err != nil {
		return fmt.Errorf("failed to read checks.dbf: %w", err)
	}
	
	// Get column indices for checks.dbf
	checksColumns, ok := checksData["columns"].([]string)
	if !ok {
		return fmt.Errorf("invalid checks.dbf structure")
	}
	
	// Find relevant columns including CENTRYTYPE
	var checkNumIdx, amountIdx, accountIdx, clearedIdx, voidIdx, entryTypeIdx int = -1, -1, -1, -1, -1, -1
	for i, col := range checksColumns {
		colUpper := strings.ToUpper(col)
		switch colUpper {
		case "CCHECKNO":
			checkNumIdx = i
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
		}
	}
	
	if checkNumIdx == -1 || amountIdx == -1 || entryTypeIdx == -1 {
		return fmt.Errorf("required columns not found in checks.dbf (need CCHECKNO, NAMOUNT, CENTRYTYPE)")
	}
	
	// Process rows to calculate reconciliation amount using decimal arithmetic
	unclearedDeposits := currency.Zero()
	unclearedChecks := currency.Zero()
	var depositCount, checkCount int
	checksRows, _ := checksData["rows"].([][]interface{})
	
	fmt.Printf("RefreshOutstandingChecks: Processing %d rows for account %s (with CENTRYTYPE logic)\n", len(checksRows), accountNumber)
	
	// Debug: Show first few rows to see what data we're actually reading
	for i := 0; i < 5 && i < len(checksRows); i++ {
		row := checksRows[i]
		if len(row) > accountIdx && len(row) > checkNumIdx && len(row) > amountIdx && len(row) > entryTypeIdx {
			checkNum := fmt.Sprintf("%v", row[checkNumIdx])
			checkAccount := ""
			if accountIdx != -1 && len(row) > accountIdx && row[accountIdx] != nil {
				checkAccount = strings.TrimSpace(fmt.Sprintf("%v", row[accountIdx]))
			}
			amount := parseFloat(row[amountIdx])
			entryType := strings.TrimSpace(fmt.Sprintf("%v", row[entryTypeIdx]))
			fmt.Printf("RefreshOutstandingChecks: Sample row %d - Entry: %s, Type: %s, Account: %s, Amount: $%.2f\n", 
				i+1, checkNum, entryType, checkAccount, amount)
		}
	}
	
	for _, row := range checksRows {
		if len(row) <= checkNumIdx || len(row) <= amountIdx || len(row) <= entryTypeIdx {
			continue
		}
		
		// Get account for this entry
		checkAccount := ""
		if accountIdx != -1 && len(row) > accountIdx && row[accountIdx] != nil {
			checkAccount = strings.TrimSpace(fmt.Sprintf("%v", row[accountIdx]))
		}
		
		// If account filter is provided, only include entries for that account
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
					// CRITICAL FIX: Empty string should be FALSE (not cleared)
					if lowerVal == "" {
						isCleared = false
					} else {
						isCleared = (lowerVal == "t" || lowerVal == ".t." || lowerVal == "true" || lowerVal == "1")
					}
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
					// CRITICAL FIX: Empty string should be FALSE (not voided)
					if lowerVal == "" {
						isVoided = false
					} else {
						isVoided = (lowerVal == "t" || lowerVal == ".t." || lowerVal == "true" || lowerVal == "1")
					}
				}
			}
		}
		
		// Only include if not cleared and not voided
		if !isCleared && !isVoided {
			amount := currency.ParseFromDBF(row[amountIdx])
			entryType := strings.TrimSpace(strings.ToUpper(fmt.Sprintf("%v", row[entryTypeIdx])))
			
			if entryType == "D" {
				// Deposit - adds to bank balance
				unclearedDeposits = unclearedDeposits.Add(amount)
				depositCount++
				
				// Debug logging for first few uncleared deposits
				if depositCount <= 3 {
					checkNum := fmt.Sprintf("%v", row[checkNumIdx])
					fmt.Printf("RefreshOutstandingChecks: Uncleared Deposit #%d: %s, Amount: %s, Account: %s\n", 
						depositCount, checkNum, amount.String(), checkAccount)
				}
			} else if entryType == "C" {
				// Check - subtracts from bank balance
				unclearedChecks = unclearedChecks.Add(amount)
				checkCount++
				
				// Debug logging for first few uncleared checks
				if checkCount <= 3 {
					checkNum := fmt.Sprintf("%v", row[checkNumIdx])
					fmt.Printf("RefreshOutstandingChecks: Uncleared Check #%d: %s, Amount: %s, Account: %s\n", 
						checkCount, checkNum, amount.String(), checkAccount)
				}
			}
		}
	}
	
	// Calculate the net reconciliation adjustment: Deposits - Checks
	// This represents the adjustment needed to get from GL Balance to Bank Balance
	totalReconciliationAdjustment := unclearedDeposits.Sub(unclearedChecks)
	totalItemCount := depositCount + checkCount
	
	fmt.Printf("RefreshOutstandingChecks: Account %s reconciliation summary:\n", accountNumber)
	fmt.Printf("  - Uncleared Deposits: %d items, %s total\n", depositCount, unclearedDeposits.String())
	fmt.Printf("  - Uncleared Checks: %d items, %s total\n", checkCount, unclearedChecks.String())
	fmt.Printf("  - Net Reconciliation Adjustment: %s (GL Balance + this = Bank Balance)\n", totalReconciliationAdjustment.String())
	
	// Get current cached balance
	currentBalance, err := GetCachedBalance(db, companyName, accountNumber)
	if err != nil {
		return err
	}
	
	// Create detailed metadata for frontend display
	metadata := fmt.Sprintf(`{"uncleared_deposits": %s, "uncleared_checks": %s, "deposit_count": %d, "check_count": %d}`,
		unclearedDeposits.ToString(), unclearedChecks.ToString(), depositCount, checkCount)
	
	// Use UPSERT to handle race conditions
	// Store as string to preserve decimal precision
	_, err = db.Exec(`
		INSERT INTO account_balances 
		(company_name, account_number, account_name, account_type, 
		 outstanding_checks_total, outstanding_checks_count, outstanding_checks_last_updated, metadata)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
		ON CONFLICT(company_name, account_number) 
		DO UPDATE SET 
			outstanding_checks_total = excluded.outstanding_checks_total, 
			outstanding_checks_count = excluded.outstanding_checks_count, 
			outstanding_checks_last_updated = CURRENT_TIMESTAMP,
			metadata = excluded.metadata
	`, companyName, accountNumber, "", 1, totalReconciliationAdjustment.ToString(), totalItemCount, metadata)
	
	if err != nil {
		// Fallback to simple UPDATE if UPSERT fails
		_, err = db.Exec(`
			UPDATE account_balances 
			SET outstanding_checks_total = ?, outstanding_checks_count = ?, 
			    outstanding_checks_last_updated = CURRENT_TIMESTAMP, metadata = ?
			WHERE company_name = ? AND account_number = ?
		`, totalReconciliationAdjustment.ToString(), totalItemCount, metadata, companyName, accountNumber)
		
		// Record the change in history
		if err == nil {
			// Calculate new bank balance using Currency type to avoid precision loss
			newBankBalance := currency.NewFromFloat(currentBalance.GLBalance).Add(totalReconciliationAdjustment)
			_, err = db.Exec(`
				INSERT INTO balance_history 
				(account_balance_id, company_name, account_number, change_type,
				 old_outstanding_total, new_outstanding_total, 
				 old_available_balance, new_available_balance,
				 change_reason, changed_by)
				VALUES (?, ?, ?, 'checks_refresh', ?, ?, ?, ?, 'Bank reconciliation refresh (deposits-checks)', ?)
			`, currentBalance.ID, companyName, accountNumber,
				currentBalance.OutstandingTotal, totalReconciliationAdjustment.ToString(),
				currentBalance.BankBalance, newBankBalance.ToString(),
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