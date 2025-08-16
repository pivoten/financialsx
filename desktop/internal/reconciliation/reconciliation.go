package reconciliation

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	
	"github.com/pivoten/financialsx/desktop/internal/common"
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/database"
)

// Reconciliation represents a bank reconciliation record
type Reconciliation struct {
	ID                 int               `json:"id"`
	CompanyName        string            `json:"company_name"`
	AccountNumber      string            `json:"account_number"`
	ReconcileDate      time.Time         `json:"reconcile_date"`
	StatementDate      time.Time         `json:"statement_date"`
	BeginningBalance   float64           `json:"beginning_balance"`
	EndingBalance      float64           `json:"ending_balance"`
	StatementBalance   float64           `json:"statement_balance"`
	StatementCredits   float64           `json:"statement_credits"`
	StatementDebits    float64           `json:"statement_debits"`
	ExtendedData       map[string]interface{} `json:"extended_data"`
	SelectedChecksJSON string            `json:"selected_checks_json"`
	SelectedChecks     []SelectedCheck   `json:"selected_checks"`
	Status             string            `json:"status"`
	CreatedBy          string            `json:"created_by"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	CommittedAt        *time.Time        `json:"committed_at"`
	DBFRowIndex        *int              `json:"dbf_row_index"`
	DBFLastSync        *time.Time        `json:"dbf_last_sync"`
}

// SelectedCheck represents a check selected for reconciliation
type SelectedCheck struct {
	CIDCHEC     string  `json:"cidchec"`
	CheckNumber string  `json:"checkNumber"`
	Amount      float64 `json:"amount"`
	Payee       string  `json:"payee"`
	CheckDate   string  `json:"checkDate"`
	RowIndex    int     `json:"rowIndex"`
}

// SaveDraftRequest represents the request to save a draft reconciliation
type SaveDraftRequest struct {
	CompanyName      string          `json:"company_name"`
	AccountNumber    string          `json:"account_number"`
	StatementDate    string          `json:"statement_date"`
	StatementBalance float64         `json:"statement_balance"`
	StatementCredits float64         `json:"statement_credits"`
	StatementDebits  float64         `json:"statement_debits"`
	BeginningBalance float64         `json:"beginning_balance"`
	SelectedChecks   []SelectedCheck `json:"selected_checks"`
	CreatedBy        string          `json:"created_by"`
}

// Service provides reconciliation operations
type Service struct {
	db *database.DB
}

// NewService creates a new reconciliation service
func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// SaveDraft saves or updates a draft reconciliation
func (s *Service) SaveDraft(req SaveDraftRequest) (*Reconciliation, error) {
	// Parse statement date
	statementDate, err := time.Parse("2006-01-02", req.StatementDate)
	if err != nil {
		return nil, fmt.Errorf("invalid statement date format: %w", err)
	}
	
	// Use statement date as reconcile date for drafts
	reconcileDate := statementDate
	
	// Convert selected checks to JSON
	selectedChecksJSON, err := json.Marshal(req.SelectedChecks)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal selected checks: %w", err)
	}
	
	// Calculate ending balance: beginning + credits - debits
	endingBalance := req.BeginningBalance + req.StatementCredits - req.StatementDebits
	
	// Check if draft already exists
	existing, err := s.GetDraft(req.CompanyName, req.AccountNumber)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check for existing draft: %w", err)
	}
	
	var reconciliation *Reconciliation
	
	if existing != nil {
		// Update existing draft
		_, err = s.db.Exec(`
			UPDATE reconciliations 
			SET statement_date = ?, statement_balance = ?, statement_credits = ?, 
				statement_debits = ?, beginning_balance = ?, ending_balance = ?,
				selected_checks_json = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`,
			statementDate, req.StatementBalance, req.StatementCredits,
			req.StatementDebits, req.BeginningBalance, endingBalance,
			string(selectedChecksJSON), existing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to update draft reconciliation: %w", err)
		}
		
		// Fetch updated record
		reconciliation, err = s.GetByID(existing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch updated reconciliation: %w", err)
		}
	} else {
		// Create new draft
		result, err := s.db.Exec(`
			INSERT INTO reconciliations (
				company_name, account_number, reconcile_date, statement_date,
				beginning_balance, ending_balance, statement_balance,
				statement_credits, statement_debits, selected_checks_json,
				status, created_by
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'draft', ?)`,
			req.CompanyName, req.AccountNumber, reconcileDate, statementDate,
			req.BeginningBalance, endingBalance, req.StatementBalance,
			req.StatementCredits, req.StatementDebits, string(selectedChecksJSON),
			req.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to create draft reconciliation: %w", err)
		}
		
		id, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get last insert ID: %w", err)
		}
		
		// Fetch created record
		reconciliation, err = s.GetByID(int(id))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch created reconciliation: %w", err)
		}
	}
	
	return reconciliation, nil
}

// GetDraft retrieves the current draft reconciliation for an account
func (s *Service) GetDraft(companyName, accountNumber string) (*Reconciliation, error) {
	query := `
		SELECT id, company_name, account_number, reconcile_date, statement_date,
			beginning_balance, ending_balance, statement_balance,
			statement_credits, statement_debits, extended_data,
			selected_checks_json, status, created_by, created_at,
			updated_at, committed_at, dbf_row_index, dbf_last_sync
		FROM reconciliations 
		WHERE company_name = ? AND account_number = ? AND status = 'draft'
		ORDER BY updated_at DESC
		LIMIT 1`
	
	row := s.db.QueryRow(query, companyName, accountNumber)
	return s.scanReconciliation(row)
}

// GetByID retrieves a reconciliation by ID
func (s *Service) GetByID(id int) (*Reconciliation, error) {
	query := `
		SELECT id, company_name, account_number, reconcile_date, statement_date,
			beginning_balance, ending_balance, statement_balance,
			statement_credits, statement_debits, extended_data,
			selected_checks_json, status, created_by, created_at,
			updated_at, committed_at, dbf_row_index, dbf_last_sync
		FROM reconciliations 
		WHERE id = ?`
	
	row := s.db.QueryRow(query, id)
	return s.scanReconciliation(row)
}

// GetHistory retrieves reconciliation history for an account
func (s *Service) GetHistory(companyName, accountNumber string, limit int) ([]*Reconciliation, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	
	query := `
		SELECT id, company_name, account_number, reconcile_date, statement_date,
			beginning_balance, ending_balance, statement_balance,
			statement_credits, statement_debits, extended_data,
			selected_checks_json, status, created_by, created_at,
			updated_at, committed_at, dbf_row_index, dbf_last_sync
		FROM reconciliations 
		WHERE company_name = ? AND account_number = ? AND status != 'draft'
		ORDER BY reconcile_date DESC, created_at DESC
		LIMIT ?`
	
	rows, err := s.db.Query(query, companyName, accountNumber, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query reconciliation history: %w", err)
	}
	defer rows.Close()
	
	var reconciliations []*Reconciliation
	for rows.Next() {
		rec, err := s.scanReconciliation(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reconciliation: %w", err)
		}
		reconciliations = append(reconciliations, rec)
	}
	
	return reconciliations, rows.Err()
}

// GetLastCommitted retrieves the last committed reconciliation for an account
func (s *Service) GetLastCommitted(companyName, accountNumber string) (*Reconciliation, error) {
	query := `
		SELECT id, company_name, account_number, reconcile_date, statement_date,
			beginning_balance, ending_balance, statement_balance,
			statement_credits, statement_debits, extended_data,
			selected_checks_json, status, created_by, created_at,
			updated_at, committed_at, dbf_row_index, dbf_last_sync
		FROM reconciliations 
		WHERE company_name = ? AND account_number = ? AND status = 'committed'
		ORDER BY reconcile_date DESC, committed_at DESC
		LIMIT 1`
	
	row := s.db.QueryRow(query, companyName, accountNumber)
	return s.scanReconciliation(row)
}

// CommitReconciliation commits a draft reconciliation
func (s *Service) CommitReconciliation(id int, committedBy string) error {
	now := time.Now()
	
	_, err := s.db.Exec(`
		UPDATE reconciliations 
		SET status = 'committed', committed_at = ?, updated_at = ?
		WHERE id = ? AND status = 'draft'`,
		now, now, id)
	if err != nil {
		return fmt.Errorf("failed to commit reconciliation: %w", err)
	}
	
	return nil
}

// DeleteDraft deletes a draft reconciliation
func (s *Service) DeleteDraft(companyName, accountNumber string) error {
	_, err := s.db.Exec(`
		DELETE FROM reconciliations 
		WHERE company_name = ? AND account_number = ? AND status = 'draft'`,
		companyName, accountNumber)
	if err != nil {
		return fmt.Errorf("failed to delete draft reconciliation: %w", err)
	}
	
	return nil
}

// MigrateFromDBF imports existing reconciliation data from CHECKREC.DBF
func (s *Service) MigrateFromDBF(companyName string) (*MigrationResult, error) {
	// This would integrate with the company.ReadDBFFile function
	// For now, we'll implement a placeholder that can be called from the API
	result := &MigrationResult{
		CompanyName:      companyName,
		RecordsProcessed: 0,
		RecordsImported:  0,
		Errors:          []string{},
	}
	
	// TODO: Implement actual DBF migration
	// 1. Read CHECKREC.dbf using company.ReadDBFFile()
	// 2. Parse each record and convert to Reconciliation struct
	// 3. Insert into SQLite with status = 'committed'
	// 4. Set dbf_row_index for bidirectional sync
	
	return result, nil
}

// MigrationResult represents the result of a DBF migration operation
type MigrationResult struct {
	CompanyName      string   `json:"company_name"`
	RecordsProcessed int      `json:"records_processed"`
	RecordsImported  int      `json:"records_imported"`
	Errors          []string `json:"errors"`
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     time.Time `json:"completed_at"`
}

// scanReconciliation scans a row into a Reconciliation struct
func (s *Service) scanReconciliation(scanner interface{}) (*Reconciliation, error) {
	var rec Reconciliation
	var extendedDataJSON string
	var selectedChecksJSON sql.NullString
	var committedAt sql.NullTime
	var dbfRowIndex sql.NullInt64
	var dbfLastSync sql.NullTime
	
	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&rec.ID, &rec.CompanyName, &rec.AccountNumber,
			&rec.ReconcileDate, &rec.StatementDate,
			&rec.BeginningBalance, &rec.EndingBalance, &rec.StatementBalance,
			&rec.StatementCredits, &rec.StatementDebits, &extendedDataJSON,
			&selectedChecksJSON, &rec.Status, &rec.CreatedBy,
			&rec.CreatedAt, &rec.UpdatedAt, &committedAt,
			&dbfRowIndex, &dbfLastSync,
		)
	case *sql.Rows:
		err = s.Scan(
			&rec.ID, &rec.CompanyName, &rec.AccountNumber,
			&rec.ReconcileDate, &rec.StatementDate,
			&rec.BeginningBalance, &rec.EndingBalance, &rec.StatementBalance,
			&rec.StatementCredits, &rec.StatementDebits, &extendedDataJSON,
			&selectedChecksJSON, &rec.Status, &rec.CreatedBy,
			&rec.CreatedAt, &rec.UpdatedAt, &committedAt,
			&dbfRowIndex, &dbfLastSync,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan reconciliation: %w", err)
	}
	
	// Handle nullable fields
	if committedAt.Valid {
		rec.CommittedAt = &committedAt.Time
	}
	if dbfRowIndex.Valid {
		index := int(dbfRowIndex.Int64)
		rec.DBFRowIndex = &index
	}
	if dbfLastSync.Valid {
		rec.DBFLastSync = &dbfLastSync.Time
	}
	
	// Parse JSON fields
	if extendedDataJSON == "" {
		extendedDataJSON = "{}"
	}
	if err := json.Unmarshal([]byte(extendedDataJSON), &rec.ExtendedData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal extended data: %w", err)
	}
	
	if selectedChecksJSON.Valid && selectedChecksJSON.String != "" {
		rec.SelectedChecksJSON = selectedChecksJSON.String
		if err := json.Unmarshal([]byte(selectedChecksJSON.String), &rec.SelectedChecks); err != nil {
			return nil, fmt.Errorf("failed to unmarshal selected checks: %w", err)
		}
	} else {
		rec.SelectedChecksJSON = "[]"
		rec.SelectedChecks = []SelectedCheck{}
	}
	
	return &rec, nil
}

// GetLastReconciliationFromDBF reads the last reconciliation from CHECKREC.dbf
func (s *Service) GetLastReconciliationFromDBF(companyName, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("GetLastReconciliationFromDBF called for company: %s, account: %s\n", companyName, accountNumber)

	// Read CHECKREC.dbf
	checkrecData, err := company.ReadDBFFile(companyName, "CHECKREC.dbf", "", 0, 0, "", "")
	if err != nil {
		fmt.Printf("GetLastReconciliationFromDBF: Failed to read CHECKREC.dbf: %v\n", err)
		return map[string]interface{}{
			"status":  "no_data",
			"message": "No reconciliation history found",
		}, nil
	}

	// Get column indices
	columns, ok := checkrecData["columns"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid CHECKREC.dbf structure")
	}

	// Find relevant columns
	var accountIdx, dateIdx, endBalIdx, begBalIdx, clearedCountIdx, clearedAmtIdx int = -1, -1, -1, -1, -1, -1
	for i, col := range columns {
		colUpper := strings.ToUpper(col)
		switch colUpper {
		case "CACCTNO":
			accountIdx = i
		case "DRECDATE":
			dateIdx = i
		case "NENDBAL":
			endBalIdx = i
		case "NBEGBAL":
			begBalIdx = i
		case "NCLEARED":
			clearedCountIdx = i
		case "NCLEAREDAMT":
			clearedAmtIdx = i
		}
	}

	if accountIdx == -1 || dateIdx == -1 || endBalIdx == -1 {
		return nil, fmt.Errorf("required columns not found in CHECKREC.dbf")
	}

	// Process rows to find records for this account
	rows, _ := checkrecData["rows"].([][]interface{})
	var accountRecords []map[string]interface{}

	for _, row := range rows {
		if len(row) <= accountIdx {
			continue
		}

		rowAccount := strings.TrimSpace(fmt.Sprintf("%v", row[accountIdx]))
		if rowAccount != accountNumber {
			continue
		}

		record := map[string]interface{}{
			"account_number": rowAccount,
		}

		// Get date
		if dateIdx != -1 && len(row) > dateIdx && row[dateIdx] != nil {
			if t, ok := row[dateIdx].(time.Time); ok {
				record["date"] = t
				record["date_string"] = t.Format("2006-01-02")
			} else {
				dateStr := fmt.Sprintf("%v", row[dateIdx])
				record["date_string"] = dateStr
				// Try to parse the date
				for _, format := range []string{"2006-01-02", "01/02/2006", "1/2/2006"} {
					if parsedDate, err := time.Parse(format, dateStr); err == nil {
						record["date"] = parsedDate
						record["date_string"] = parsedDate.Format("2006-01-02")
						break
					}
				}
			}
		}

		// Get balances
		if endBalIdx != -1 && len(row) > endBalIdx {
			record["ending_balance"] = common.ParseFloat(row[endBalIdx])
		}
		if begBalIdx != -1 && len(row) > begBalIdx {
			record["beginning_balance"] = common.ParseFloat(row[begBalIdx])
		}
		if clearedCountIdx != -1 && len(row) > clearedCountIdx {
			record["cleared_count"] = int(common.ParseFloat(row[clearedCountIdx]))
		}
		if clearedAmtIdx != -1 && len(row) > clearedAmtIdx {
			record["cleared_amount"] = common.ParseFloat(row[clearedAmtIdx])
		}

		// Only add if we have a valid date
		if _, hasDate := record["date"]; hasDate {
			accountRecords = append(accountRecords, record)
		}
	}

	if len(accountRecords) == 0 {
		return map[string]interface{}{
			"status":  "no_data",
			"message": "No reconciliation history found for this account",
		}, nil
	}

	// Sort by date (most recent first)
	for i := 0; i < len(accountRecords)-1; i++ {
		for j := i + 1; j < len(accountRecords); j++ {
			date1, _ := accountRecords[i]["date"].(time.Time)
			date2, _ := accountRecords[j]["date"].(time.Time)
			if date2.After(date1) {
				accountRecords[i], accountRecords[j] = accountRecords[j], accountRecords[i]
			}
		}
	}

	// Return the most recent reconciliation
	lastRec := accountRecords[0]
	lastRec["status"] = "success"
	lastRec["total_records"] = len(accountRecords)

	return lastRec, nil
}