package main

import (
	"context"
	"embed"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/auth"
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/config"
	"github.com/pivoten/financialsx/desktop/internal/database"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// App struct
type App struct {
	ctx      context.Context
	db       *database.DB
	auth     *auth.Auth
	currentUser *auth.User
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return "Hello " + name + ", It's show time!"
}

// GetCompanies returns list of detected companies
func (a *App) GetCompanies() ([]company.Company, error) {
	return company.DetectCompanies()
}

// Login handles user login
func (a *App) Login(username, password, companyName string) (map[string]interface{}, error) {
	// Initialize database for the company if not already done
	if a.db == nil || a.currentUser == nil || a.currentUser.CompanyName != companyName {
		if a.db != nil {
			a.db.Close()
		}
		
		db, err := database.New(companyName)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		
		// Initialize balance cache tables
		err = database.InitializeBalanceCache(db)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize balance cache: %w", err)
		}
		
		a.db = db
		a.auth = auth.New(db, companyName) // Pass companyName to Auth constructor
	}

	user, session, err := a.auth.Login(username, password)
	if err != nil {
		return nil, err
	}

	a.currentUser = user

	return map[string]interface{}{
		"user":    user,
		"session": session,
	}, nil
}

// Register handles user registration
func (a *App) Register(username, password, email, companyName string) (map[string]interface{}, error) {
	// Create company directory structure (including sql folder) if it doesn't exist
	if err := company.CreateCompanyDirectory(companyName); err != nil {
		return nil, fmt.Errorf("failed to create company directory: %w", err)
	}

	// Initialize database for the company
	if a.db == nil || (a.currentUser != nil && a.currentUser.CompanyName != companyName) {
		if a.db != nil {
			a.db.Close()
		}
		
		db, err := database.New(companyName)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		
		// Initialize balance cache tables
		err = database.InitializeBalanceCache(db)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize balance cache: %w", err)
		}
		
		a.db = db
		a.auth = auth.New(db, companyName) // Pass companyName to Auth constructor
	}

	user, err := a.auth.Register(username, password, email) // Remove companyName parameter
	if err != nil {
		return nil, err
	}

	// Auto-login after registration
	_, session, err := a.auth.Login(username, password)
	if err != nil {
		return nil, err
	}

	a.currentUser = user

	return map[string]interface{}{
		"user":    user,
		"session": session,
	}, nil
}

// Logout handles user logout
func (a *App) Logout(token string) error {
	if a.auth != nil {
		return a.auth.Logout(token)
	}
	return nil
}

// ValidateSession checks if a session is valid for a specific company
func (a *App) ValidateSession(token string, companyName string) (*auth.User, error) {
	// Initialize database connection for the specific company
	if a.db == nil || a.currentUser == nil || a.currentUser.CompanyName != companyName {
		if a.db != nil {
			a.db.Close()
		}
		
		db, err := database.New(companyName)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to company database: %w", err)
		}
		
		// Initialize balance cache tables
		err = database.InitializeBalanceCache(db)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize balance cache: %w", err)
		}
		
		a.db = db
		a.auth = auth.New(db, companyName) // Pass companyName to Auth constructor
	}
	
	user, err := a.auth.ValidateSession(token) // Remove companyName parameter
	if err != nil {
		return nil, err
	}
	
	a.currentUser = user
	return user, nil
}

// GetDBFFiles returns list of DBF files for a company
func (a *App) GetDBFFiles(companyName string) ([]string, error) {
	files, err := company.GetDBFFiles(companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get DBF files: %w", err)
	}
	return files, nil
}

// GetDBFTableData returns the structure and data of a DBF file
func (a *App) GetDBFTableData(companyName, fileName string) (map[string]interface{}, error) {
	fmt.Printf("GetDBFTableData called: company=%s, file=%s\n", companyName, fileName)
	
	// Call the real DBF reading function without search - NO RECORD LIMIT
	return company.ReadDBFFile(companyName, fileName, "", 0, 0, "", "")
}

// GetDBFTableDataPaged returns paginated and sorted data from a DBF file
func (a *App) GetDBFTableDataPaged(companyName, fileName string, offset, limit int, sortColumn, sortDirection string) (map[string]interface{}, error) {
	fmt.Printf("GetDBFTableDataPaged called: company=%s, file=%s, offset=%d, limit=%d, sort=%s %s\n", 
		companyName, fileName, offset, limit, sortColumn, sortDirection)
	
	return company.ReadDBFFile(companyName, fileName, "", offset, limit, sortColumn, sortDirection)
}

// SearchDBFTable searches a DBF file and returns matching records
func (a *App) SearchDBFTable(companyName, fileName, searchTerm string) (map[string]interface{}, error) {
	fmt.Printf("SearchDBFTable called: company=%s, file=%s, search=%s\n", companyName, fileName, searchTerm)
	
	// Call the DBF reading function with search term (no limit - get all matching records)
	return company.ReadDBFFile(companyName, fileName, searchTerm, 0, 0, "", "")
}

// UpdateDBFRecord updates a specific record in a DBF file
func (a *App) UpdateDBFRecord(companyName, fileName string, rowIndex, colIndex int, value string) error {
	err := company.UpdateDBFRecord(companyName, fileName, rowIndex, colIndex, value)
	if err != nil {
		return fmt.Errorf("failed to update DBF record: %w", err)
	}
	return nil
}

// GetDashboardData returns aggregated data for the dashboard
func (a *App) GetDashboardData(companyName string) (map[string]interface{}, error) {
	fmt.Printf("GetDashboardData called for company: %s\n", companyName)
	
	// Security check: ensure user can only access their own company data
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if a.currentUser.CompanyName != companyName {
		return nil, fmt.Errorf("access denied: user can only access data for company '%s'", a.currentUser.CompanyName)
	}
	
	return company.GetDashboardData(companyName)
}

// User Management Functions

// GetAllUsers returns all users (admin/root only)
func (a *App) GetAllUsers() ([]auth.User, error) {
	if a.auth == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("users.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	return a.auth.GetAllUsers()
}

// GetAllRoles returns all available roles
func (a *App) GetAllRoles() ([]auth.Role, error) {
	if a.auth == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("users.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	return a.auth.GetAllRoles()
}

// UpdateUserRole updates a user's role (admin/root only)
func (a *App) UpdateUserRole(userID, newRoleID int) error {
	if a.auth == nil {
		return fmt.Errorf("not authenticated")
	}
	
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("users.manage_roles") {
		return fmt.Errorf("insufficient permissions")
	}
	
	return a.auth.UpdateUserRole(userID, newRoleID)
}

// UpdateUserStatus activates or deactivates a user (admin/root only)
func (a *App) UpdateUserStatus(userID int, isActive bool) error {
	if a.auth == nil {
		return fmt.Errorf("not authenticated")
	}
	
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("users.update") {
		return fmt.Errorf("insufficient permissions")
	}
	
	return a.auth.UpdateUserStatus(userID, isActive)
}

// CreateUser creates a new user (admin/root only)
func (a *App) CreateUser(username, password, email string, roleID int) (*auth.User, error) {
	if a.auth == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("users.create") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Note: This would need a new method in auth.go for admin-created users
	// For now, we'll use the existing Register method with some modifications
	return a.auth.Register(username, password, email)
}

// Configuration Management Functions

// GetAPIKey retrieves an API key for a service
func (a *App) GetAPIKey(service string) (string, error) {
	// Check permissions - only admins can view API keys
	if a.currentUser == nil || (!a.currentUser.IsAdmin() && !a.currentUser.HasPermission("settings.read")) {
		return "", fmt.Errorf("insufficient permissions")
	}
	
	key := config.GetAPIKey(service)
	return key, nil
}

// SetAPIKey sets an API key for a service
func (a *App) SetAPIKey(service, key string) error {
	// Check permissions - only admins can set API keys
	if a.currentUser == nil || (!a.currentUser.IsAdmin() && !a.currentUser.HasPermission("settings.write")) {
		return fmt.Errorf("insufficient permissions")
	}
	
	return config.UpdateAPIKey(service, key)
}

// GetConfig retrieves the current configuration
func (a *App) GetConfig() (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("settings.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	cfg := config.GetConfig()
	
	// Return sanitized config (without exposing full API keys)
	result := map[string]interface{}{
		"settings": cfg.Settings,
		"api_keys_configured": map[string]bool{
			"openweather": cfg.APIKeys.OpenWeather != "",
		},
	}
	
	return result, nil
}

// TestAPIKey tests if an API key is valid
func (a *App) TestAPIKey(service, key string) (bool, error) {
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("settings.write") {
		return false, fmt.Errorf("insufficient permissions")
	}
	
	switch service {
	case "openweather":
		// Here you would implement actual API testing
		// For now, just check if key is not empty
		return key != "" && len(key) > 10, nil
	default:
		return false, fmt.Errorf("unknown service: %s", service)
	}
}

// Process Management Functions

// RunClosingProcess executes the month-end closing process
func (a *App) RunClosingProcess(periodEnd, closingDate, description string, forceClose bool) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("database.maintain") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	// For now, return a placeholder response until we implement the closing process
	return map[string]interface{}{
		"status":   "not_implemented",
		"message":  "Closing process not yet implemented",
		"duration": "0s",
	}, nil
}

// GetClosingStatus returns the status of a period
func (a *App) GetClosingStatus(periodEnd string) (string, error) {
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("database.read") {
		return "", fmt.Errorf("insufficient permissions")
	}

	// For now, return a placeholder status
	return "open", nil
}

// ReopenPeriod reopens a closed period
func (a *App) ReopenPeriod(periodEnd, reason string) error {
	// Check permissions - only root/admin can reopen
	if a.currentUser == nil || !a.currentUser.IsAdmin() {
		return fmt.Errorf("insufficient permissions")
	}

	// For now, return success placeholder
	return nil
}

// Net Distribution Functions

// RunNetDistribution executes the net distribution process
func (a *App) RunNetDistribution(periodStart, periodEnd string, processType string, recalculateAll bool) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("database.maintain") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	// TODO: Uncomment when ready to use the distribution processor
	/*
	logger := log.New(log.Writer(), fmt.Sprintf("[NETDIST-%s] ", a.currentUser.CompanyName), log.LstdFlags)
	netDistProcess := processes.NewDistributionProcessor(a.db, a.currentUser.CompanyName, logger)

	periodStartDate, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, fmt.Errorf("invalid period start date: %w", err)
	}

	periodEndDate, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid period end date: %w", err)
	}

	config := &processes.ProcessingConfig{
		Period:      fmt.Sprintf("%02d", periodStartDate.Month()),
		Year:        fmt.Sprintf("%04d", periodStartDate.Year()),
		AcctDate:    periodEndDate,
		RevDate:     periodEndDate,
		ExpDate:     periodEndDate,
		UserID:      a.currentUser.ID,
		IsNewRun:    true,
		IsClosing:   false,
	}
	*/

	// For now, return a placeholder response while we're developing
	// TODO: Uncomment below when ready to hook up the full distribution processor
	/*
	options := &processes.ProcessingOptions{
		RevSummarize: true,
		ExpSummarize: true,
		GLSummary:    true,
	}

	if err := netDistProcess.Initialize(config, options); err != nil {
		return nil, fmt.Errorf("failed to initialize distribution processor: %w", err)
	}

	result, err := netDistProcess.Main()
	if err != nil {
		return nil, fmt.Errorf("net distribution process failed: %w", err)
	}

	// Convert result to map for JSON serialization
	return map[string]interface{}{
		"run_number":       result.RunNumber,
		"run_year":         result.RunYear,
		"status":          result.Status,
		"wells_processed": result.WellsProcessed,
		"owners_processed": result.OwnersProcessed,
		"records_created":  result.RecordsCreated,
		"total_revenue":    result.TotalRevenue.String(),
		"total_expenses":   result.TotalExpenses.String(),
		"net_distributed":  result.NetDistributed.String(),
		"warnings":        result.Warnings,
		"errors":          result.Errors,
		"duration":        result.Duration.String(),
		"start_time":      result.StartTime,
		"end_time":        result.EndTime,
	}, nil
	*/

	// Placeholder response for development
	return map[string]interface{}{
		"status":   "development",
		"message":  "Distribution processor in development - not yet connected",
		"duration": "0s",
	}, nil
}

// GetNetDistributionStatus returns the status of net distribution for a period
func (a *App) GetNetDistributionStatus(periodStart, periodEnd string) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	// Placeholder response for development
	return map[string]interface{}{
		"record_count":      0,
		"total_net_amount":  0.0,
		"well_count":        0,
		"owner_count":       0,
		"has_distributions": false,
		"last_processed":    nil,
	}, nil
}

// ExportNetDistribution exports distribution results to DBF format
func (a *App) ExportNetDistribution(periodStart, periodEnd, outputPath string) error {
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("database.read") {
		return fmt.Errorf("insufficient permissions")
	}

	// Placeholder for development - export not yet implemented
	return fmt.Errorf("export functionality not yet implemented")
}

// GetBankAccounts returns bank accounts from COA.dbf where LBANKACCT = true
func (a *App) GetBankAccounts(companyName string) ([]map[string]interface{}, error) {
	fmt.Printf("GetBankAccounts called for company: %s\n", companyName)
	
	// Check permissions
	if a.currentUser == nil {
		fmt.Printf("GetBankAccounts: currentUser is nil\n")
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		fmt.Printf("GetBankAccounts: user %s lacks database.read permission\n", a.currentUser.Username)
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	fmt.Printf("GetBankAccounts: permissions OK, reading COA.dbf\n")

	// Read COA.dbf file (no limit - get all records for financial accuracy)
	fmt.Printf("GetBankAccounts: About to read COA.dbf for company: %s\n", companyName)
	coaData, err := company.ReadDBFFile(companyName, "COA.dbf", "", 0, 0, "", "")
	if err != nil {
		fmt.Printf("GetBankAccounts: failed to read COA.dbf: %v\n", err)
		return []map[string]interface{}{}, fmt.Errorf("failed to read COA.dbf: %w", err)
	}
	
	if coaData == nil {
		fmt.Printf("GetBankAccounts: coaData is nil\n")
		return []map[string]interface{}{}, fmt.Errorf("coaData is nil")
	}

	data, ok := coaData["rows"].([][]interface{})
	if !ok {
		fmt.Printf("GetBankAccounts: coaData structure: %+v\n", coaData)
		return nil, fmt.Errorf("invalid data format from COA.dbf")
	}
	
	if len(data) == 0 {
		fmt.Printf("GetBankAccounts: COA.dbf contains no data rows\n")
		return []map[string]interface{}{}, nil // Return empty slice instead of error
	}
	
	fmt.Printf("GetBankAccounts: COA.dbf loaded successfully, %d rows found\n", len(data))

	var bankAccounts []map[string]interface{}

	fmt.Printf("GetBankAccounts: Starting to process %d rows\n", len(data))
	
	for i, row := range data {
		if len(row) < 7 {
			continue // Skip incomplete rows (need at least 7 columns for LBANKACCT)
		}

		// Check LBANKACCT flag in column 6 (Lbankacct)
		bankAccountFlag := false
		if len(row) > 6 {
			switch v := row[6].(type) {
			case bool:
				bankAccountFlag = v
				if v {
					fmt.Printf("GetBankAccounts: Found bank account flag (bool) for %v\n", row[0])
				}
			case string:
				bankAccountFlag = v == "T" || v == ".T." || v == "true"
				if bankAccountFlag {
					fmt.Printf("GetBankAccounts: Found bank account flag (string) for %v\n", row[0])
				}
			default:
				// For debugging, but don't spam logs for every row
			}
		}
		
		if i < 5 || bankAccountFlag {
			fmt.Printf("GetBankAccounts: Row %d, Account %v, BankFlag: %v, Processing...\n", i, row[0], row[6])
		}

		if bankAccountFlag {
			fmt.Printf("GetBankAccounts: Creating account record for %v\n", row[0])
			account := map[string]interface{}{
				"account_number": fmt.Sprintf("%v", row[0]),   // Cacctno (Account number)
				"account_name":   fmt.Sprintf("%v", row[2]),   // Cacctdesc (Account description)
				"account_type":   fmt.Sprintf("%v", row[1]),   // Caccttype (Account type)
				"balance":        0.0,                         // Balance not in COA, will be calculated
				"description":    fmt.Sprintf("%v", row[2]),   // Cacctdesc (Account description)
				"is_bank_account": true,
			}
			bankAccounts = append(bankAccounts, account)
			fmt.Printf("GetBankAccounts: Successfully created account %s - %s\n", account["account_number"], account["account_name"])
		}
	}
	
	fmt.Printf("GetBankAccounts: returning %d bank accounts\n", len(bankAccounts))
	
	// Debug: Print each account being returned
	for i, account := range bankAccounts {
		fmt.Printf("GetBankAccounts: Account %d: %+v\n", i, account)
	}
	
	fmt.Printf("GetBankAccounts: About to return success\n")
	return bankAccounts, nil
}

// GetOutstandingChecks retrieves all checks that have not been cleared (LCLEARED = false)
// Optionally filter by account number if provided
func (a *App) GetOutstandingChecks(companyName string, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("GetOutstandingChecks called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Read checks.dbf - increase limit to get all records instead of just 10,000
	checksData, err := company.ReadDBFFile(companyName, "checks.dbf", "", 0, 50000, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read checks.dbf: %w", err)
	}
	
	// Get column indices for checks.dbf
	checksColumns, ok := checksData["columns"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid checks.dbf structure")
	}
	
	// Debug: Print all available columns
	fmt.Printf("GetOutstandingChecks: Available columns in checks.dbf: %v\n", checksColumns)
	
	// Find relevant check columns based on actual CHECKS.dbf structure
	var checkNumIdx, dateIdx, payeeIdx, amountIdx, accountIdx, clearedIdx, voidIdx int = -1, -1, -1, -1, -1, -1, -1
	for i, col := range checksColumns {
		colUpper := strings.ToUpper(col)
		if colUpper == "CCHECKNO" {
			checkNumIdx = i
			fmt.Printf("GetOutstandingChecks: Found check number column at index %d: %s\n", i, col)
		} else if colUpper == "DCHECKDATE" {
			dateIdx = i
			fmt.Printf("GetOutstandingChecks: Found date column at index %d: %s\n", i, col)
		} else if colUpper == "CPAYEE" {
			payeeIdx = i
			fmt.Printf("GetOutstandingChecks: Found payee column at index %d: %s\n", i, col)
		} else if colUpper == "NAMOUNT" {
			amountIdx = i
			fmt.Printf("GetOutstandingChecks: Found amount column at index %d: %s\n", i, col)
		} else if colUpper == "CACCTNO" {
			accountIdx = i
			fmt.Printf("GetOutstandingChecks: Found account column at index %d: %s\n", i, col)
		} else if colUpper == "LCLEARED" {
			clearedIdx = i
			fmt.Printf("GetOutstandingChecks: Found cleared column at index %d: %s\n", i, col)
		} else if colUpper == "LVOID" {
			voidIdx = i
			fmt.Printf("GetOutstandingChecks: Found void column at index %d: %s\n", i, col)
		}
	}
	
	if checkNumIdx == -1 || amountIdx == -1 {
		return map[string]interface{}{
			"status": "error",
			"error": "Required columns not found",
			"columns": checksColumns,
		}, nil
	}
	
	// Process check rows to find outstanding checks
	var outstandingChecks []map[string]interface{}
	checksRows, _ := checksData["rows"].([][]interface{})
	
	fmt.Printf("GetOutstandingChecks: Processing %d rows, filtering by account: '%s'\n", len(checksRows), accountNumber)
	
	var totalProcessed, accountMatches, clearedCount, voidCount int
	
	for _, row := range checksRows {
		if len(row) <= checkNumIdx || len(row) <= amountIdx {
			continue
		}
		
		totalProcessed++
		
		// Get account for this check first for debugging
		checkAccount := ""
		if accountIdx != -1 && len(row) > accountIdx && row[accountIdx] != nil {
			checkAccount = strings.TrimSpace(fmt.Sprintf("%v", row[accountIdx]))
		}
		
		// Track account matches for debugging
		isAccountMatch := (accountNumber == "" || checkAccount == accountNumber)
		if isAccountMatch {
			accountMatches++
		}
		
		// Check if cleared (default to false if no cleared column)
		isCleared := false
		if clearedIdx != -1 && len(row) > clearedIdx {
			clearedValue := row[clearedIdx]
			if clearedValue != nil {
				// Handle different boolean representations
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
				default:
					// For any other type, treat as false
					isCleared = false
				}
			}
		}
		
		// Check if voided (default to false if no void column)
		isVoided := false
		if voidIdx != -1 && len(row) > voidIdx {
			voidValue := row[voidIdx]
			if voidValue != nil {
				// Handle different boolean representations
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
				default:
					// For any other type, treat as false
					isVoided = false
				}
			}
		}
		
		// Debug logging for account 100000 specifically
		if isAccountMatch && accountNumber == "100000" && accountMatches <= 20 {
			checkNum := fmt.Sprintf("%v", row[checkNumIdx])
			amount := parseFloat(row[amountIdx])
			
			// Show raw values from DBF
			clearedRaw := "nil"
			voidRaw := "nil"
			if clearedIdx != -1 && len(row) > clearedIdx {
				clearedRaw = fmt.Sprintf("%v (type: %T)", row[clearedIdx], row[clearedIdx])
			}
			if voidIdx != -1 && len(row) > voidIdx {
				voidRaw = fmt.Sprintf("%v (type: %T)", row[voidIdx], row[voidIdx])
			}
			
			fmt.Printf("GetOutstandingChecks: Check %s, Account %s, Amount $%.2f\n", checkNum, checkAccount, amount)
			fmt.Printf("  Raw LCLEARED: %s -> Parsed: %t\n", clearedRaw, isCleared)
			fmt.Printf("  Raw LVOID: %s -> Parsed: %t\n", voidRaw, isVoided)
		}
		
		if isCleared {
			clearedCount++
		}
		if isVoided {
			voidCount++
		}
		
		// Only include if not cleared and not voided
		if !isCleared && !isVoided {			
			// If account filter is provided, only include checks for that account
			if accountNumber != "" && checkAccount != accountNumber {
				continue
			}
			
			check := map[string]interface{}{
				"checkNumber": fmt.Sprintf("%v", row[checkNumIdx]),
				"amount": parseFloat(row[amountIdx]),
				"account": checkAccount,
			}
			
			// Add optional fields if available
			if dateIdx != -1 && len(row) > dateIdx && row[dateIdx] != nil {
				// Properly handle DBF date fields - they can be time.Time or strings
				if t, ok := row[dateIdx].(time.Time); ok {
					// DBF library returns time.Time directly
					check["date"] = t.Format("2006-01-02")
				} else {
					// Fallback to string representation
					check["date"] = fmt.Sprintf("%v", row[dateIdx])
				}
			}
			if payeeIdx != -1 && len(row) > payeeIdx && row[payeeIdx] != nil {
				check["payee"] = fmt.Sprintf("%v", row[payeeIdx])
			}
			
			// Add raw row data for editing
			check["_rowIndex"] = len(outstandingChecks)
			check["_rawData"] = row
			
			outstandingChecks = append(outstandingChecks, check)
		}
	}
	
	fmt.Printf("GetOutstandingChecks: Found %d outstanding checks\n", len(outstandingChecks))
	fmt.Printf("GetOutstandingChecks: Summary - Processed: %d, Account matches: %d, Cleared: %d, Voided: %d, Outstanding: %d\n", 
		totalProcessed, accountMatches, clearedCount, voidCount, len(outstandingChecks))
	
	return map[string]interface{}{
		"status": "success",
		"checks": outstandingChecks,
		"total": len(outstandingChecks),
		"columns": checksColumns,
	}, nil
}

// GetAccountBalance retrieves the current GL balance for a specific account
func (a *App) GetAccountBalance(companyName, accountNumber string) (float64, error) {
	fmt.Printf("GetAccountBalance called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions
	if a.currentUser == nil {
		return 0, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return 0, fmt.Errorf("insufficient permissions")
	}
	
	// Read GLMASTER.dbf to get account balance
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.dbf", "", 0, 50000, "", "")
	if err != nil {
		fmt.Printf("GetAccountBalance: failed to read GLMASTER.dbf: %v\n", err)
		return 0, fmt.Errorf("failed to read GLMASTER.dbf: %w", err)
	}
	
	// Get column indices for GLMASTER.dbf
	glColumns, ok := glData["columns"].([]string)
	if !ok {
		return 0, fmt.Errorf("invalid GLMASTER.dbf structure")
	}
	
	// Debug: Print all available columns
	fmt.Printf("GetAccountBalance: Available columns in GLMASTER.dbf: %v\n", glColumns)
	
	// Find relevant GL columns (GLMASTER has separate debit/credit columns)
	var accountIdx, debitIdx, creditIdx int = -1, -1, -1
	for i, col := range glColumns {
		colUpper := strings.ToUpper(col)
		if colUpper == "CACCTNO" || colUpper == "ACCOUNT" || colUpper == "ACCTNO" {
			accountIdx = i
			fmt.Printf("GetAccountBalance: Found account column at index %d: %s\n", i, col)
		} else if colUpper == "NDEBITS" || colUpper == "DEBIT" || colUpper == "NDEBIT" {
			debitIdx = i
			fmt.Printf("GetAccountBalance: Found debit column at index %d: %s\n", i, col)
		} else if colUpper == "NCREDITS" || colUpper == "CREDIT" || colUpper == "NCREDIT" {
			creditIdx = i
			fmt.Printf("GetAccountBalance: Found credit column at index %d: %s\n", i, col)
		}
	}
	
	if accountIdx == -1 || (debitIdx == -1 && creditIdx == -1) {
		fmt.Printf("GetAccountBalance: could not find required columns. accountIdx=%d, debitIdx=%d, creditIdx=%d\n", accountIdx, debitIdx, creditIdx)
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
			
			// For bank accounts, debits increase balance, credits decrease balance
			entryAmount := debit - credit
			totalBalance += entryAmount
			
			if debit != 0 || credit != 0 {
				fmt.Printf("GetAccountBalance: Found entry for account %s: debit=%f, credit=%f, net=%f, running total=%f\n", 
					accountNumber, debit, credit, entryAmount, totalBalance)
			}
		}
	}
	
	fmt.Printf("GetAccountBalance: Final balance for account %s: %f\n", accountNumber, totalBalance)
	return totalBalance, nil
}

// GetCachedBalances retrieves cached balances for all bank accounts
func (a *App) GetCachedBalances(companyName string) ([]map[string]interface{}, error) {
	fmt.Printf("GetCachedBalances called for company: %s\n", companyName)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	balances, err := database.GetAllCachedBalances(a.db, companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached balances: %w", err)
	}
	
	// Convert to interface for JSON response
	var result []map[string]interface{}
	for _, balance := range balances {
		result = append(result, map[string]interface{}{
			"account_number":         balance.AccountNumber,
			"account_name":          balance.AccountName,
			"gl_balance":            balance.GLBalance,
			"outstanding_total":     balance.OutstandingTotal,
			"outstanding_count":     balance.OutstandingCount,
			"bank_balance":          balance.BankBalance,
			"gl_last_updated":       balance.GLLastUpdated,
			"checks_last_updated":   balance.OutstandingLastUpdated,
			"gl_age_hours":          balance.GLAgeHours,
			"checks_age_hours":      balance.ChecksAgeHours,
			"gl_freshness":          balance.GLFreshness,
			"checks_freshness":      balance.ChecksFreshness,
			"is_stale":             balance.GLFreshness == "stale" || balance.ChecksFreshness == "stale",
		})
	}
	
	return result, nil
}

// RefreshAccountBalance refreshes both GL and outstanding checks for an account
func (a *App) RefreshAccountBalance(companyName, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("RefreshAccountBalance called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	username := a.currentUser.Username
	
	// Refresh GL balance
	fmt.Printf("RefreshAccountBalance: Starting GL balance refresh for account %s\n", accountNumber)
	err := database.RefreshGLBalance(a.db, companyName, accountNumber, username)
	if err != nil {
		fmt.Printf("RefreshAccountBalance: GL balance refresh failed: %v\n", err)
		return nil, fmt.Errorf("failed to refresh GL balance: %w", err)
	}
	fmt.Printf("RefreshAccountBalance: GL balance refresh completed for account %s\n", accountNumber)
	
	// Refresh outstanding checks
	fmt.Printf("RefreshAccountBalance: Starting outstanding checks refresh for account %s\n", accountNumber)
	err = database.RefreshOutstandingChecks(a.db, companyName, accountNumber, username)
	if err != nil {
		fmt.Printf("RefreshAccountBalance: Outstanding checks refresh failed: %v\n", err)
		return nil, fmt.Errorf("failed to refresh outstanding checks: %w", err)
	}
	fmt.Printf("RefreshAccountBalance: Outstanding checks refresh completed for account %s\n", accountNumber)
	
	// Get the updated cached balance
	balance, err := database.GetCachedBalance(a.db, companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated balance: %w", err)
	}
	
	if balance == nil {
		return nil, fmt.Errorf("balance not found after refresh")
	}
	
	return map[string]interface{}{
		"status":                "success",
		"account_number":        balance.AccountNumber,
		"account_name":          balance.AccountName,
		"gl_balance":            balance.GLBalance,
		"outstanding_total":     balance.OutstandingTotal,
		"outstanding_count":     balance.OutstandingCount,
		"bank_balance":          balance.BankBalance,
		"gl_last_updated":       balance.GLLastUpdated,
		"checks_last_updated":   balance.OutstandingLastUpdated,
		"refreshed_by":          username,
	}, nil
}

// RefreshAllBalances refreshes balances for all bank accounts
func (a *App) RefreshAllBalances(companyName string) (map[string]interface{}, error) {
	fmt.Printf("RefreshAllBalances called for company: %s\n", companyName)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Get all bank accounts
	bankAccounts, err := a.GetBankAccounts(companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank accounts: %w", err)
	}
	
	var successCount, errorCount int
	var errors []string
	
	for _, account := range bankAccounts {
		accountNumber := account["account_number"].(string)
		_, err := a.RefreshAccountBalance(companyName, accountNumber)
		if err != nil {
			errorCount++
			errors = append(errors, fmt.Sprintf("Account %s: %v", accountNumber, err))
		} else {
			successCount++
		}
	}
	
	return map[string]interface{}{
		"status":        "completed",
		"total_accounts": len(bankAccounts),
		"success_count": successCount,
		"error_count":   errorCount,
		"errors":        errors,
		"refreshed_by":  a.currentUser.Username,
		"refresh_time":  time.Now(),
	}, nil
}

// GetBalanceHistory gets the balance change history for an account
func (a *App) GetBalanceHistory(companyName, accountNumber string, limit int) ([]map[string]interface{}, error) {
	fmt.Printf("GetBalanceHistory called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	if limit <= 0 {
		limit = 50
	}
	
	query := `
		SELECT bh.*, ab.account_name
		FROM balance_history bh
		JOIN account_balances ab ON bh.account_balance_id = ab.id
		WHERE bh.company_name = ? AND bh.account_number = ?
		ORDER BY bh.change_timestamp DESC
		LIMIT ?
	`
	
	rows, err := a.db.Query(query, companyName, accountNumber, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query balance history: %w", err)
	}
	defer rows.Close()
	
	var history []map[string]interface{}
	for rows.Next() {
		var h database.BalanceHistory
		var accountName string
		
		err := rows.Scan(
			&h.ID, &h.AccountBalanceID, &h.CompanyName, &h.AccountNumber,
			&h.ChangeType, &h.OldGLBalance, &h.NewGLBalance,
			&h.OldOutstandingTotal, &h.NewOutstandingTotal,
			&h.OldAvailableBalance, &h.NewAvailableBalance,
			&h.ChangeReason, &h.ChangedBy, &h.ChangeTimestamp, &h.Metadata,
			&accountName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan balance history: %w", err)
		}
		
		history = append(history, map[string]interface{}{
			"id":                     h.ID,
			"account_name":           accountName,
			"change_type":            h.ChangeType,
			"old_gl_balance":         h.OldGLBalance,
			"new_gl_balance":         h.NewGLBalance,
			"old_outstanding_total":  h.OldOutstandingTotal,
			"new_outstanding_total":  h.NewOutstandingTotal,
			"old_available_balance":  h.OldAvailableBalance,
			"new_available_balance":  h.NewAvailableBalance,
			"change_reason":          h.ChangeReason,
			"changed_by":             h.ChangedBy,
			"change_timestamp":       h.ChangeTimestamp,
		})
	}
	
	return history, nil
}

// AuditCheckBatches performs an audit comparing checks.dbf entries with GLMASTER.dbf
func (a *App) AuditCheckBatches(companyName string) (map[string]interface{}, error) {
	fmt.Printf("AuditCheckBatches called for company: %s\n", companyName)
	
	// Check permissions - only admin/root can audit
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.IsRoot && a.currentUser.RoleName != "Admin" {
		return nil, fmt.Errorf("insufficient permissions - admin or root required")
	}
	
	// Read checks.dbf (no limit - get all check records for complete audit)
	checksData, err := company.ReadDBFFile(companyName, "checks.dbf", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read checks.dbf: %w", err)
	}
	
	// Read GLMASTER.dbf
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.dbf", "", 0, 50000, "", "")
	if err != nil {
		// If GLMASTER.dbf doesn't exist, return informative error
		return map[string]interface{}{
			"status": "error",
			"error": "GLMASTER.dbf not found",
			"message": "The GLMASTER.dbf file is required for audit but was not found in the company directory",
		}, nil
	}
	
	// Get column indices for checks.dbf
	checksColumns, ok := checksData["columns"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid checks.dbf structure")
	}
	
	// Find CBATCH column index (if it exists)
	cbatchIdx := -1
	amountIdx := -1
	checkNumIdx := -1
	for i, col := range checksColumns {
		colUpper := strings.ToUpper(col)
		if colUpper == "CBATCH" {
			cbatchIdx = i
		} else if colUpper == "AMOUNT" || colUpper == "NAMOUNT" {
			amountIdx = i
		} else if colUpper == "CHECKNUM" || colUpper == "CCHECKNUM" {
			checkNumIdx = i
		}
	}
	
	// If no CBATCH column, try to use check number or other identifier
	if cbatchIdx == -1 {
		fmt.Printf("Warning: CBATCH column not found in checks.dbf, using check number for audit\n")
	}
	
	// Get column indices for GLMASTER.dbf
	glColumns, ok := glData["columns"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid GLMASTER.dbf structure")
	}
	
	// Find relevant GL columns
	glBatchIdx := -1
	glAmountIdx := -1
	glRefIdx := -1
	for i, col := range glColumns {
		colUpper := strings.ToUpper(col)
		if colUpper == "CBATCH" {
			glBatchIdx = i
		} else if colUpper == "AMOUNT" || colUpper == "NAMOUNT" {
			glAmountIdx = i
		} else if colUpper == "CREF" || colUpper == "REFERENCE" {
			glRefIdx = i
		}
	}
	
	// Build GL lookup map
	glMap := make(map[string][]map[string]interface{})
	glRows, _ := glData["rows"].([][]interface{})
	
	for _, row := range glRows {
		var key string
		if glBatchIdx >= 0 && glBatchIdx < len(row) {
			key = fmt.Sprintf("%v", row[glBatchIdx])
		} else if glRefIdx >= 0 && glRefIdx < len(row) {
			key = fmt.Sprintf("%v", row[glRefIdx])
		}
		
		if key != "" {
			entry := map[string]interface{}{
				"row": row,
				"columns": glColumns,
			}
			if glAmountIdx >= 0 && glAmountIdx < len(row) {
				entry["amount"] = parseFloat(row[glAmountIdx])
			}
			glMap[key] = append(glMap[key], entry)
		}
	}
	
	// Audit results
	var missingEntries []map[string]interface{}
	var mismatchedAmounts []map[string]interface{}
	totalChecks := 0
	totalMatched := 0
	
	// Check each entry in checks.dbf
	checksRows, _ := checksData["rows"].([][]interface{})
	
	for i, row := range checksRows {
		if len(row) == 0 {
			continue
		}
		
		totalChecks++
		
		// Get check identifier (CBATCH or check number)
		var checkID string
		if cbatchIdx >= 0 && cbatchIdx < len(row) {
			checkID = fmt.Sprintf("%v", row[cbatchIdx])
		} else if checkNumIdx >= 0 && checkNumIdx < len(row) {
			checkID = fmt.Sprintf("%v", row[checkNumIdx])
		} else {
			checkID = fmt.Sprintf("Row_%d", i)
		}
		
		// Get check amount
		var checkAmount float64
		if amountIdx >= 0 && amountIdx < len(row) {
			checkAmount = parseFloat(row[amountIdx])
		}
		
		// Look for matching GL entries
		glEntries, found := glMap[checkID]
		
		if !found || len(glEntries) == 0 {
			// No GL entry found
			missingEntries = append(missingEntries, map[string]interface{}{
				"check_id": checkID,
				"amount": checkAmount,
				"row_index": i,
				"check_data": row,
				"check_columns": checksColumns,
			})
		} else {
			// Check if amounts match
			matched := false
			for _, glEntry := range glEntries {
				glAmount, _ := glEntry["amount"].(float64)
				if math.Abs(checkAmount - glAmount) < 0.01 { // Allow for small rounding differences
					matched = true
					totalMatched++
					break
				}
			}
			
			if !matched {
				mismatchedAmounts = append(mismatchedAmounts, map[string]interface{}{
					"check_id": checkID,
					"check_amount": checkAmount,
					"gl_entries": glEntries,
					"row_index": i,
					"check_data": row,
					"check_columns": checksColumns,
				})
			}
		}
	}
	
	// Prepare audit report
	auditReport := map[string]interface{}{
		"status": "completed",
		"summary": map[string]interface{}{
			"total_checks": totalChecks,
			"matched_entries": totalMatched,
			"missing_entries": len(missingEntries),
			"mismatched_amounts": len(mismatchedAmounts),
		},
		"missing_entries": missingEntries,
		"mismatched_amounts": mismatchedAmounts,
		"check_columns": checksColumns,
		"gl_columns": glColumns,
		"audit_date": time.Now().Format("2006-01-02 15:04:05"),
		"audited_by": a.currentUser.Username,
	}
	
	fmt.Printf("Audit completed: %d checks, %d matched, %d missing, %d mismatched\n", 
		totalChecks, totalMatched, len(missingEntries), len(mismatchedAmounts))
	
	return auditReport, nil
}

// Helper function to safely parse float values from DBF
func parseFloat(value interface{}) float64 {
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
		if f, err := fmt.Sscanf(v, "%f"); err == nil {
			return float64(f)
		}
	}
	return 0.0
}

// AuditBankReconciliation performs a bank reconciliation audit comparing:
// Bank Reconciliation Balance vs (GL Balance + Outstanding Checks)
func (a *App) AuditBankReconciliation(companyName string) (map[string]interface{}, error) {
	fmt.Printf("AuditBankReconciliation called for company: %s\n", companyName)
	
	// Check permissions - only admin/root can audit
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.IsRoot && a.currentUser.RoleName != "Admin" {
		return nil, fmt.Errorf("insufficient permissions - admin or root required")
	}

	// Force refresh of cached balances for each bank account before audit (as requested by user)
	fmt.Printf("AuditBankReconciliation: Getting bank accounts first to refresh individually...\n")
	tempBankAccounts, err := a.GetBankAccounts(companyName)
	if err == nil && len(tempBankAccounts) > 0 {
		fmt.Printf("AuditBankReconciliation: Refreshing cached balances for %d bank accounts...\n", len(tempBankAccounts))
		for _, account := range tempBankAccounts {
			if accountNum, ok := account["account_number"].(string); ok && accountNum != "" {
				fmt.Printf("AuditBankReconciliation: Refreshing balance for account %s\n", accountNum)
				_, err := a.RefreshAccountBalance(companyName, accountNum)
				if err != nil {
					fmt.Printf("AuditBankReconciliation: Warning - failed to refresh balance for account %s: %v\n", accountNum, err)
				}
			}
		}
	} else {
		fmt.Printf("AuditBankReconciliation: Warning - could not get bank accounts for refresh: %v\n", err)
	}

	// Get bank accounts from COA.dbf
	bankAccounts, err := a.GetBankAccounts(companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank accounts: %w", err)
	}
	
	if len(bankAccounts) == 0 {
		return map[string]interface{}{
			"status": "warning",
			"message": "No bank accounts found in Chart of Accounts",
			"discrepancies": []interface{}{},
			"accounts_audited": 0,
			"total_discrepancies": 0,
		}, nil
	}

	// Read CHECKREC.dbf for reconciliation data (no limit - get all reconciliation records)
	fmt.Printf("AuditBankReconciliation: Reading CHECKREC.dbf for company: %s\n", companyName)
	checkrecData, err := company.ReadDBFFile(companyName, "CHECKREC.dbf", "", 0, 0, "", "")
	if err != nil {
		fmt.Printf("AuditBankReconciliation: Error reading CHECKREC.dbf: %v\n", err)
		return map[string]interface{}{
			"status": "error",
			"error": "CHECKREC.dbf not found",
			"message": "The CHECKREC.dbf file is required for bank reconciliation audit but was not found in the company directory",
		}, nil
	}

	// Get column indices for CHECKREC.dbf
	checkrecColumns, ok := checkrecData["columns"].([]string)
	if !ok {
		fmt.Printf("AuditBankReconciliation: Invalid CHECKREC.dbf structure - columns not found\n")
		return nil, fmt.Errorf("invalid CHECKREC.dbf structure")
	}
	
	fmt.Printf("AuditBankReconciliation: CHECKREC.dbf columns: %v\n", checkrecColumns)
	
	// Get rows for debugging
	checkrecRows, rowsOk := checkrecData["rows"].([][]interface{})
	if !rowsOk {
		fmt.Printf("AuditBankReconciliation: Invalid CHECKREC.dbf structure - rows not found\n")
		return nil, fmt.Errorf("invalid CHECKREC.dbf structure - no rows")
	}
	
	fmt.Printf("AuditBankReconciliation: CHECKREC.dbf has %d rows\n", len(checkrecRows))

	// Find relevant columns in CHECKREC.dbf
	var accountIdx, endingBalanceIdx, dateIdx int = -1, -1, -1
	for i, col := range checkrecColumns {
		colUpper := strings.ToUpper(col)
		if colUpper == "CACCTNO" {
			accountIdx = i
		} else if colUpper == "NENDBAL" {
			endingBalanceIdx = i
		} else if colUpper == "DRECDATE" {
			dateIdx = i
		}
	}

	if accountIdx == -1 || endingBalanceIdx == -1 {
		return nil, fmt.Errorf("required columns not found in CHECKREC.dbf (need account and ending balance)")
	}

	// Build reconciliation data per account
	type ReconciliationRecord struct {
		Date    time.Time
		Balance float64
	}
	
	accountRecords := make(map[string][]ReconciliationRecord)

	// Collect all reconciliation records per account
	for i, row := range checkrecRows {
		if len(row) <= accountIdx || len(row) <= endingBalanceIdx {
			if i < 5 { // Debug first few rows
				fmt.Printf("AuditBankReconciliation: Row %d skipped - insufficient columns (need idx %d and %d, got %d)\n", i, accountIdx, endingBalanceIdx, len(row))
			}
			continue
		}

		accountNum := strings.TrimSpace(fmt.Sprintf("%v", row[accountIdx]))
		endingBalance := parseFloat(row[endingBalanceIdx])
		
		if i < 5 { // Debug first few rows
			fmt.Printf("AuditBankReconciliation: Row %d - Account: %s, Balance: %f\n", i, accountNum, endingBalance)
		}
		
		// Parse date (required for proper logic)
		var recDate time.Time
		if dateIdx != -1 && len(row) > dateIdx && row[dateIdx] != nil {
			// Check if it's already a time.Time object from DBF
			if t, ok := row[dateIdx].(time.Time); ok {
				recDate = t
				if i < 5 { // Debug first few rows
					fmt.Printf("AuditBankReconciliation: Row %d - Got time.Time directly: %s\n", i, recDate.Format("2006-01-02"))
				}
			} else if dateStr := fmt.Sprintf("%v", row[dateIdx]); dateStr != "" {
				if i < 5 { // Debug first few rows
					fmt.Printf("AuditBankReconciliation: Row %d - Date string: %s\n", i, dateStr)
				}
				// Handle various string formats
				for _, format := range []string{
					"2006-01-02 15:04:05 -0700 MST", // Go time.Time string format
					"2006-01-02T15:04:05Z", 
					"2006-01-02T15:04:05", 
					"2006-01-02", 
					"01/02/2006", 
					"1/2/2006"} {
					if parsedDate, err := time.Parse(format, dateStr); err == nil {
						recDate = parsedDate
						if i < 5 { // Debug first few rows
							fmt.Printf("AuditBankReconciliation: Row %d - Parsed date from string: %s\n", i, recDate.Format("2006-01-02"))
						}
						break
					}
				}
			}
		} else {
			if i < 5 { // Debug first few rows
				fmt.Printf("AuditBankReconciliation: Row %d - No date column or data\n", i)
			}
		}
		
		// Only include records with valid dates
		if !recDate.IsZero() {
			accountRecords[accountNum] = append(accountRecords[accountNum], ReconciliationRecord{
				Date:    recDate,
				Balance: endingBalance,
			})
			if i < 5 { // Debug first few rows
				fmt.Printf("AuditBankReconciliation: Row %d - Added record for account %s\n", i, accountNum)
			}
		} else {
			if i < 5 { // Debug first few rows
				fmt.Printf("AuditBankReconciliation: Row %d - Skipped due to invalid date for account %s\n", i, accountNum)
			}
		}
	}
	
	fmt.Printf("AuditBankReconciliation: Processed records for %d accounts\n", len(accountRecords))

	// Process each account to find the correct reconciliation balance and date
	lastRecBalances := make(map[string]float64)
	lastRecDates := make(map[string]time.Time)
	
	for accountNum, records := range accountRecords {
		if len(records) == 0 {
			continue
		}
		
		// Sort by date (latest first)
		for i := 0; i < len(records)-1; i++ {
			for j := i + 1; j < len(records); j++ {
				if records[j].Date.After(records[i].Date) {
					records[i], records[j] = records[j], records[i]
				}
			}
		}
		
		// Logic: If latest reconciliation has NENDBAL = 0 (interim), use previous reconciliation's balance
		// but use the latest reconciliation's date
		latestRecord := records[0]
		lastRecDates[accountNum] = latestRecord.Date
		
		if latestRecord.Balance == 0 && len(records) > 1 {
			// This is an interim reconciliation - use previous reconciliation's balance
			lastRecBalances[accountNum] = records[1].Balance
		} else {
			// Normal case - use the latest reconciliation's balance
			lastRecBalances[accountNum] = latestRecord.Balance
		}
	}

	// Get current cached balances for comparison
	cachedBalances, err := database.GetAllCachedBalances(a.db, companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached balances: %w", err)
	}

	// Create map for fast lookup
	balanceMap := make(map[string]*database.CachedBalance)
	for i, balance := range cachedBalances {
		balanceMap[balance.AccountNumber] = &cachedBalances[i]
	}

	// Perform the audit
	var discrepancies []map[string]interface{}
	accountsAudited := 0
	totalDiscrepancies := 0

	for _, account := range bankAccounts {
		accountNum := account["account_number"].(string)
		accountsAudited++

		// Get reconciliation balance and date
		reconciliationBalance, hasRecData := lastRecBalances[accountNum]
		reconciliationDate, hasRecDate := lastRecDates[accountNum]
		
		// Get cached balance data
		cachedBalance, hasCachedData := balanceMap[accountNum]
		
		var glBalance, outstandingChecks float64
		var glFreshness, checksFreshness string
		
		if hasCachedData {
			glBalance = cachedBalance.GLBalance
			outstandingChecks = cachedBalance.OutstandingTotal
			glFreshness = cachedBalance.GLFreshness
			checksFreshness = cachedBalance.ChecksFreshness
		} else {
			// If no cached data, skip this account or mark as no data
			discrepancies = append(discrepancies, map[string]interface{}{
				"account_number": accountNum,
				"account_name": account["account_name"],
				"issue_type": "no_cached_data",
				"description": "No cached balance data available for this account (run refresh)",
				"reconciliation_balance": reconciliationBalance,
				"reconciliation_date": func() interface{} {
					if hasRecDate && !reconciliationDate.IsZero() {
						return reconciliationDate.Format("2006-01-02")
					}
					return nil
				}(),
				"has_reconciliation_data": hasRecData,
			})
			totalDiscrepancies++
			continue
		}

		// Check if we have reconciliation data
		if !hasRecData {
			discrepancies = append(discrepancies, map[string]interface{}{
				"account_number": accountNum,
				"account_name": account["account_name"],
				"issue_type": "no_reconciliation_data",
				"description": "No reconciliation data found in CHECKREC.dbf",
				"gl_balance": glBalance,
				"outstanding_checks": outstandingChecks,
				"reconciliation_date": nil,
				"gl_freshness": glFreshness,
				"checks_freshness": checksFreshness,
			})
			totalDiscrepancies++
			continue
		}

		// Calculate the expected GL balance: Reconciliation Balance - Outstanding Checks
		expectedGLBalance := reconciliationBalance - outstandingChecks
		difference := glBalance - expectedGLBalance
		tolerance := 0.01 // 1 cent tolerance for rounding

		if difference > tolerance || difference < -tolerance {
			discrepancies = append(discrepancies, map[string]interface{}{
				"account_number": accountNum,
				"account_name": account["account_name"],
				"issue_type": "balance_mismatch",
				"description": fmt.Sprintf("GL balance does not match (Reconciliation Balance - Outstanding Checks)"),
				"reconciliation_balance": reconciliationBalance,
				"reconciliation_date": func() interface{} {
					if hasRecDate && !reconciliationDate.IsZero() {
						return reconciliationDate.Format("2006-01-02")
					}
					return nil
				}(),
				"gl_balance": glBalance,
				"outstanding_checks": outstandingChecks,
				"expected_gl_balance": expectedGLBalance,
				"difference": difference,
				"gl_freshness": glFreshness,
				"checks_freshness": checksFreshness,
			})
			totalDiscrepancies++
		}
	}

	return map[string]interface{}{
		"status": "success",
		"message": fmt.Sprintf("Bank reconciliation audit completed for %d accounts", accountsAudited),
		"discrepancies": discrepancies,
		"accounts_audited": accountsAudited,
		"total_discrepancies": totalDiscrepancies,
		"audit_timestamp": time.Now().Format(time.RFC3339),
		"audited_by": a.currentUser.Username,
	}, nil
}

// AuditSingleBankAccount performs bank reconciliation audit for a single account
func (a *App) AuditSingleBankAccount(companyName, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("AuditSingleBankAccount called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions - only admin/root can audit
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.IsRoot && a.currentUser.RoleName != "Admin" {
		return nil, fmt.Errorf("insufficient permissions - admin or root required")
	}

	// Refresh balance for this specific account
	fmt.Printf("AuditSingleBankAccount: Refreshing balance for account %s\n", accountNumber)
	_, err := a.RefreshAccountBalance(companyName, accountNumber)
	if err != nil {
		fmt.Printf("AuditSingleBankAccount: Warning - failed to refresh balance for account %s: %v\n", accountNumber, err)
	}

	// Get the specific bank account info
	bankAccounts, err := a.GetBankAccounts(companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank accounts: %w", err)
	}
	
	var targetAccount map[string]interface{}
	for _, account := range bankAccounts {
		if accNum, ok := account["account_number"].(string); ok && accNum == accountNumber {
			targetAccount = account
			break
		}
	}
	
	if targetAccount == nil {
		return map[string]interface{}{
			"status": "error",
			"error": "account_not_found",
			"message": fmt.Sprintf("Bank account %s not found in Chart of Accounts", accountNumber),
		}, nil
	}

	// Get cached balance data
	cachedBalances, err := a.GetCachedBalances(companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached balances: %w", err)
	}
	
	var glBalance, outstandingChecks float64
	var outstandingCount int
	for _, balance := range cachedBalances {
		if balance["account_number"].(string) == accountNumber {
			glBalance = parseFloat(balance["gl_balance"])
			outstandingChecks = parseFloat(balance["outstanding_total"])
			outstandingCount = int(parseFloat(balance["outstanding_count"]))
			break
		}
	}

	// Read CHECKREC.dbf for reconciliation data (no limit - get all reconciliation records)
	checkrecData, err := company.ReadDBFFile(companyName, "CHECKREC.dbf", "", 0, 0, "", "")
	if err != nil {
		return map[string]interface{}{
			"status": "error",
			"error": "CHECKREC.dbf not found",
			"message": "The CHECKREC.dbf file is required for bank reconciliation audit",
		}, nil
	}

	// Get column indices for CHECKREC.dbf
	checkrecColumns, ok := checkrecData["columns"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid CHECKREC.dbf structure")
	}

	// Find relevant columns
	var accountIdx, endingBalanceIdx, dateIdx int = -1, -1, -1
	for i, col := range checkrecColumns {
		colUpper := strings.ToUpper(col)
		if colUpper == "CACCTNO" {
			accountIdx = i
		} else if colUpper == "NENDBAL" {
			endingBalanceIdx = i
		} else if colUpper == "DRECDATE" {
			dateIdx = i
		}
	}

	if accountIdx == -1 || endingBalanceIdx == -1 {
		return nil, fmt.Errorf("required columns not found in CHECKREC.dbf")
	}

	// Find reconciliation records for this account
	checkrecRows, _ := checkrecData["rows"].([][]interface{})
	
	type ReconciliationRecord struct {
		Date    time.Time
		Balance float64
	}
	
	var accountRecords []ReconciliationRecord
	
	for _, row := range checkrecRows {
		if len(row) <= accountIdx || len(row) <= endingBalanceIdx {
			continue
		}

		rowAccountNum := strings.TrimSpace(fmt.Sprintf("%v", row[accountIdx]))
		if rowAccountNum != accountNumber {
			continue // Skip records for other accounts
		}

		endingBalance := parseFloat(row[endingBalanceIdx])
		
		var recDate time.Time
		if dateIdx != -1 && len(row) > dateIdx && row[dateIdx] != nil {
			if t, ok := row[dateIdx].(time.Time); ok {
				recDate = t
			}
		}
		
		if !recDate.IsZero() {
			accountRecords = append(accountRecords, ReconciliationRecord{
				Date:    recDate,
				Balance: endingBalance,
			})
		}
	}

	// Process reconciliation records
	if len(accountRecords) == 0 {
		return map[string]interface{}{
			"status": "success",
			"account_number": accountNumber,
			"account_name": targetAccount["account_name"],
			"issue_type": "no_reconciliation_data",
			"gl_balance": glBalance,
			"outstanding_checks": outstandingChecks,
			"outstanding_count": outstandingCount,
			"reconciliation_balance": nil,
			"reconciliation_date": nil,
			"expected_gl_balance": nil,
			"difference": nil,
			"audit_timestamp": time.Now().Format(time.RFC3339),
			"audited_by": a.currentUser.Username,
		}, nil
	}

	// Sort by date (latest first)
	for i := 0; i < len(accountRecords)-1; i++ {
		for j := i + 1; j < len(accountRecords); j++ {
			if accountRecords[j].Date.After(accountRecords[i].Date) {
				accountRecords[i], accountRecords[j] = accountRecords[j], accountRecords[i]
			}
		}
	}

	// Apply reconciliation logic
	latestRecord := accountRecords[0]
	reconciliationBalance := latestRecord.Balance
	reconciliationDate := latestRecord.Date

	// Handle interim reconciliation (NENDBAL = 0)
	if reconciliationBalance == 0 && len(accountRecords) > 1 {
		for _, record := range accountRecords[1:] {
			if record.Balance > 0 {
				reconciliationBalance = record.Balance
				break
			}
		}
	}

	// Calculate expected GL balance: Reconciliation Balance - Outstanding Checks = GL Balance
	expectedGLBalance := reconciliationBalance - outstandingChecks
	difference := glBalance - expectedGLBalance

	// Determine issue type
	var issueType string
	if math.Abs(difference) < 0.01 { // Allow for small rounding differences
		issueType = "balanced"
	} else {
		issueType = "discrepancy_found"
	}

	return map[string]interface{}{
		"status": "success",
		"account_number": accountNumber,
		"account_name": targetAccount["account_name"],
		"issue_type": issueType,
		"gl_balance": glBalance,
		"outstanding_checks": outstandingChecks,
		"outstanding_count": outstandingCount,
		"reconciliation_balance": reconciliationBalance,
		"reconciliation_date": reconciliationDate.Format("2006-01-02"),
		"expected_gl_balance": expectedGLBalance,
		"difference": difference,
		"audit_timestamp": time.Now().Format(time.RFC3339),
		"audited_by": a.currentUser.Username,
	}, nil
}

// GetBankAccountsForAudit returns a list of bank accounts available for auditing
func (a *App) GetBankAccountsForAudit(companyName string) ([]map[string]interface{}, error) {
	fmt.Printf("GetBankAccountsForAudit called for company: %s\n", companyName)
	
	// Check permissions - only admin/root can audit
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.IsRoot && a.currentUser.RoleName != "Admin" {
		return nil, fmt.Errorf("insufficient permissions - admin or root required")
	}

	// Get bank accounts
	bankAccounts, err := a.GetBankAccounts(companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank accounts: %w", err)
	}

	// Get cached balances for additional info
	cachedBalances, err := a.GetCachedBalances(companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached balances: %w", err)
	}

	// Create balance lookup map
	balanceMap := make(map[string]map[string]interface{})
	for _, balance := range cachedBalances {
		accountNum := balance["account_number"].(string)
		balanceMap[accountNum] = balance
	}

	// Enhance bank accounts with balance info
	var auditAccounts []map[string]interface{}
	for _, account := range bankAccounts {
		accountNum := account["account_number"].(string)
		auditAccount := map[string]interface{}{
			"account_number": accountNum,
			"account_name":   account["account_name"],
			"account_type":   account["account_type"],
			"gl_balance":     0.0,
			"outstanding_checks": 0.0,
			"outstanding_count": 0,
			"last_audited":   nil,
		}

		if balance, exists := balanceMap[accountNum]; exists {
			auditAccount["gl_balance"] = balance["gl_balance"]
			auditAccount["outstanding_checks"] = balance["outstanding_total"]
			auditAccount["outstanding_count"] = balance["outstanding_count"]
		}

		auditAccounts = append(auditAccounts, auditAccount)
	}

	return auditAccounts, nil
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Pivoten FinancialsX Desktop",
		Width:  1400,
		Height: 1000,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
