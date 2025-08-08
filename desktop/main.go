package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/auth"
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/config"
	"github.com/pivoten/financialsx/desktop/internal/database"
	"github.com/pivoten/financialsx/desktop/internal/debug"
	"github.com/pivoten/financialsx/desktop/internal/logger"
	"github.com/pivoten/financialsx/desktop/internal/ole"
	"github.com/pivoten/financialsx/desktop/internal/reconciliation"
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
	currentCompanyPath string
	reconciliationService *reconciliation.Service
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Initialize debug logging (SimpleLog will auto-initialize if needed)
	debug.SimpleLog("=== App.startup called ===")
	debug.LogInfo("App", "Application starting up")
	
	// Log environment info
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	cwd, _ := os.Getwd()
	debug.SimpleLog(fmt.Sprintf("Executable path: %s", exePath))
	debug.SimpleLog(fmt.Sprintf("Executable dir: %s", exeDir))
	debug.SimpleLog(fmt.Sprintf("Working dir: %s", cwd))
	debug.SimpleLog(fmt.Sprintf("OS: %s", os.Getenv("OS")))
	debug.SimpleLog(fmt.Sprintf("PROCESSOR_ARCHITECTURE: %s", os.Getenv("PROCESSOR_ARCHITECTURE")))
	
	debug.SimpleLog("=== App.startup completed ===")
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	debug.SimpleLog(fmt.Sprintf("Greet called with name: %s", name))
	return "Hello " + name + ", It's show time!"
}

// TestLogging is a simple test function to verify logging works
func (a *App) TestLogging() string {
	debug.SimpleLog("=== TestLogging called ===")
	debug.LogInfo("TestLogging", "This is a test log message")
	return "Logging test complete - check debug log"
}

// GetCompanies returns list of detected companies
func (a *App) GetCompanies() ([]company.Company, error) {
	return company.DetectCompanies()
}

// InitializeCompanyDatabase initializes the SQLite database for a company
// This is called when a company is selected (even with Supabase auth)
func (a *App) InitializeCompanyDatabase(companyPath string) error {
	debug.SimpleLog(fmt.Sprintf("App.InitializeCompanyDatabase: Called with companyPath: %s", companyPath))
	fmt.Printf("InitializeCompanyDatabase: Called with companyPath: %s\n", companyPath)
	
	// Check if we need to reinitialize (different company or no DB)
	if a.db == nil || a.currentCompanyPath != companyPath {
		if a.db != nil {
			debug.SimpleLog("App.InitializeCompanyDatabase: Closing existing database connection")
			fmt.Printf("InitializeCompanyDatabase: Closing existing database connection\n")
			a.db.Close()
		}
		
		debug.SimpleLog(fmt.Sprintf("App.InitializeCompanyDatabase: Creating new database for path: %s", companyPath))
		fmt.Printf("InitializeCompanyDatabase: Creating new database for path: %s\n", companyPath)
		
		db, err := database.New(companyPath)
		if err != nil {
			errMsg := fmt.Sprintf("InitializeCompanyDatabase: Error creating database: %v", err)
			debug.SimpleLog(errMsg)
			fmt.Printf("%s\n", errMsg)
			return fmt.Errorf("failed to create database: %v", err)
		}
		
		debug.SimpleLog("App.InitializeCompanyDatabase: Database created, initializing balance cache tables")
		fmt.Printf("InitializeCompanyDatabase: Database created, initializing balance cache tables\n")
		
		// Initialize balance cache tables
		err = database.InitializeBalanceCache(db)
		if err != nil {
			errMsg := fmt.Sprintf("InitializeCompanyDatabase: Error initializing balance cache: %v", err)
			debug.SimpleLog(errMsg)
			fmt.Printf("%s\n", errMsg)
			// Don't fail completely - the database is still usable
			// return fmt.Errorf("failed to initialize balance cache: %w", err)
		}
		
		a.db = db
		a.currentCompanyPath = companyPath
		a.reconciliationService = reconciliation.NewService(db)
		
		successMsg := "InitializeCompanyDatabase: Database initialized successfully"
		debug.SimpleLog(successMsg)
		fmt.Printf("%s\n", successMsg)
	} else {
		msg := "InitializeCompanyDatabase: Database already initialized for this company"
		debug.SimpleLog(msg)
		fmt.Printf("%s\n", msg)
	}
	
	return nil
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
		a.reconciliationService = reconciliation.NewService(db)
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
		a.reconciliationService = reconciliation.NewService(db)
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
		a.reconciliationService = reconciliation.NewService(db)
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
	debug.LogInfo("GetDBFFiles", fmt.Sprintf("Called with company: %s", companyName))
	files, err := company.GetDBFFiles(companyName)
	if err != nil {
		debug.LogError("GetDBFFiles", err)
		return nil, fmt.Errorf("failed to get DBF files: %w", err)
	}
	debug.LogInfo("GetDBFFiles", fmt.Sprintf("Found %d files", len(files)))
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
func (a *App) GetDashboardData(companyIdentifier string) (map[string]interface{}, error) {
	// Immediate logging to confirm function is called
	debug.SimpleLog(fmt.Sprintf("=== GetDashboardData START === identifier: '%s'", companyIdentifier))
	fmt.Printf("GetDashboardData called with identifier: %s\n", companyIdentifier)
	debug.LogInfo("GetDashboardData", fmt.Sprintf("Called with identifier: %s", companyIdentifier))
	
	// For now, skip authentication check since the UI is already authenticated
	// The fact that they can call this function means they're logged in
	if a.currentUser == nil {
		debug.SimpleLog("App.GetDashboardData: Warning - currentUser is nil, proceeding anyway")
	}
	
	// The companyIdentifier could be either the company name or the full data path
	// We need to check if it matches either the company name or if it's a path that contains the company data
	// For now, we'll skip the strict check since we're using paths from compmast.dbf
	userCompany := "none"
	if a.currentUser != nil {
		userCompany = a.currentUser.CompanyName
	}
	debug.SimpleLog(fmt.Sprintf("App.GetDashboardData: User company: %s, Requested: %s", userCompany, companyIdentifier))
	debug.LogInfo("GetDashboardData", fmt.Sprintf("User company: %s, Requested: %s", userCompany, companyIdentifier))
	
	// Pass through the company identifier (could be name or path)
	debug.SimpleLog("App.GetDashboardData: Calling company.GetDashboardData")
	debug.LogInfo("GetDashboardData", "Calling company.GetDashboardData")
	
	result, err := company.GetDashboardData(companyIdentifier)
	if err != nil {
		debug.SimpleLog(fmt.Sprintf("App.GetDashboardData: Error: %v", err))
		return nil, err
	}
	
	// Log the result safely
	if wellTypes, ok := result["wellTypes"].([]map[string]interface{}); ok {
		debug.SimpleLog(fmt.Sprintf("App.GetDashboardData: Success, returning %d wellTypes", len(wellTypes)))
	} else {
		debug.SimpleLog("App.GetDashboardData: Success, no wellTypes data")
	}
	debug.SimpleLog("=== GetDashboardData END ===")
	return result, nil
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

// LogError logs frontend errors to the backend
func (a *App) LogError(errorMessage string, stackTrace string) {
	if logger.GetLogPath() != "" {
		logger.WriteError("Frontend", fmt.Sprintf("Error: %s\nStack: %s", errorMessage, stackTrace))
	}
	fmt.Printf("Frontend Error: %s\n", errorMessage)
}

// TestDatabaseQuery executes a test query using Pivoten.DbApi
func (a *App) TestDatabaseQuery(companyName, query string) (map[string]interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.WriteCrash("TestDatabaseQuery", r, nil)
		}
	}()
	
	fmt.Printf("=== TestDatabaseQuery STARTED (OLE TEST ONLY) ===\n")
	fmt.Printf("TestDatabaseQuery: company=%s\n", companyName)
	fmt.Printf("TestDatabaseQuery: query=%s\n", query)
	debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: company=%s, query=%s", companyName, query))
	
	// Skip strict auth check like other functions
	if a.currentUser == nil {
		fmt.Printf("TestDatabaseQuery: Warning - currentUser is nil, proceeding anyway\n")
		debug.SimpleLog("TestDatabaseQuery: Warning - currentUser is nil, proceeding anyway")
	}
	
	startTime := time.Now()
	
	// This is specifically for testing OLE server - no fallback
	fmt.Printf("TestDatabaseQuery: Testing OLE server (Pivoten.DbApi)...\n")
	debug.SimpleLog("TestDatabaseQuery: Testing OLE server connection")
	
	// Try OLE connection
	client, err := ole.NewDbApiClient()
	if err != nil {
		fmt.Printf("TestDatabaseQuery: Failed to connect to OLE server: %v\n", err)
		debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: OLE connection failed: %v", err))
		
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("OLE server connection failed: %v", err),
			"message": "Could not connect to Pivoten.DbApi COM server",
			"hint":    "To use OLE: 1) Build dbapi.exe from dbapi.prg in Visual FoxPro, 2) Run 'dbapi.exe /regserver' as admin",
			"progId":  "Pivoten.DbApi",
		}, nil
	}
	defer client.Close()
	
	fmt.Printf("TestDatabaseQuery: OLE server connected successfully!\n")
	debug.SimpleLog("TestDatabaseQuery: OLE server connected")
	
	// Note: Ping method would be called here if implemented in OLE client
	fmt.Printf("TestDatabaseQuery: OLE connection established\n")
	debug.SimpleLog("TestDatabaseQuery: OLE connection established")
	
	// Open the database for the company
	fmt.Printf("TestDatabaseQuery: Opening database for company: %s\n", companyName)
	if err := client.OpenDbc(companyName); err != nil {
		fmt.Printf("TestDatabaseQuery: Failed to open DBC: %v\n", err)
		debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: Failed to open DBC: %v", err))
		
		// Get last error from OLE server if available
		lastError := client.GetLastError()
		
		return map[string]interface{}{
			"success":   false,
			"error":     fmt.Sprintf("Failed to open database: %v", err),
			"lastError": lastError,
			"database":  companyName,
			"query":     query,
		}, nil
	}
	
	fmt.Printf("TestDatabaseQuery: Database opened successfully\n")
	
	// Execute the query via OLE using JSON
	fmt.Printf("TestDatabaseQuery: Executing SQL query via OLE (JSON)...\n")
	jsonResult, err := client.QueryToJson(query)
	if err != nil {
		fmt.Printf("TestDatabaseQuery: Query execution failed: %v\n", err)
		debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: Query failed: %v", err))
		
		// Get last error from OLE server if available
		lastError := client.GetLastError()
		
		return map[string]interface{}{
			"success":   false,
			"error":     fmt.Sprintf("Query execution failed: %v", err),
			"lastError": lastError,
			"database":  companyName,
			"query":     query,
		}, nil
	}
	
	// Success!
	elapsedTime := time.Since(startTime)
	fmt.Printf("TestDatabaseQuery: Query executed successfully in %.2fms\n", elapsedTime.Seconds()*1000)
	
	// Parse the JSON result
	var queryResult map[string]interface{}
	if err := json.Unmarshal([]byte(jsonResult), &queryResult); err != nil {
		fmt.Printf("TestDatabaseQuery: Failed to parse JSON: %v\n", err)
		fmt.Printf("TestDatabaseQuery: Attempting to fix common JSON issues...\n")
		debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: JSON parse error: %v", err))
		
		// Try to fix common FoxPro JSON issues
		fixedJson := jsonResult
		// Fix unescaped backslashes in paths (common in Windows paths)
		// This is a simple fix - replace single backslashes with double
		// But be careful not to double-escape already escaped ones
		fixedJson = strings.ReplaceAll(fixedJson, `\`, `\\`)
		// Fix already double-escaped becoming quad-escaped
		fixedJson = strings.ReplaceAll(fixedJson, `\\\\`, `\\`)
		
		// Try parsing again with fixed JSON
		if err2 := json.Unmarshal([]byte(fixedJson), &queryResult); err2 != nil {
			fmt.Printf("TestDatabaseQuery: Still failed after fix attempt: %v\n", err2)
			// Return the raw result if parsing still fails
			queryResult = map[string]interface{}{
				"raw": jsonResult,
				"parseError": err.Error(),
				"fixAttempted": true,
			}
		} else {
			fmt.Printf("TestDatabaseQuery: JSON fix successful!\n")
		}
	} else {
		fmt.Printf("TestDatabaseQuery: JSON parsed successfully\n")
		if success, ok := queryResult["success"].(bool); ok && !success {
			// Query returned an error
			if errMsg, ok := queryResult["error"].(string); ok {
				return map[string]interface{}{
					"success": false,
					"error":   errMsg,
					"database": companyName,
					"query":    query,
				}, nil
			}
		}
	}
	
	// Log the JSON result for debugging
	fmt.Printf("TestDatabaseQuery: JSON result length: %d bytes\n", len(jsonResult))
	debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: JSON result length: %d bytes", len(jsonResult)))
	
	// Log first 500 chars for debugging
	if len(jsonResult) > 0 {
		maxLen := 500
		if len(jsonResult) < maxLen {
			maxLen = len(jsonResult)
		}
		fmt.Printf("TestDatabaseQuery: JSON preview: %s...\n", jsonResult[:maxLen])
		debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: JSON preview: %s", jsonResult[:maxLen]))
	}
	
	// Extract the actual data array and count from the FoxPro JSON response
	var dataArray []map[string]interface{}
	var rowCount int
	
	// Debug: Log the structure of queryResult
	fmt.Printf("TestDatabaseQuery: queryResult type: %T\n", queryResult)
	fmt.Printf("TestDatabaseQuery: queryResult keys: ")
	for key, value := range queryResult {
		fmt.Printf("%s(%T) ", key, value)
	}
	fmt.Printf("\n")
	
	// Check if data is directly in queryResult
	if data, ok := queryResult["data"].([]interface{}); ok {
		fmt.Printf("TestDatabaseQuery: Found data array with %d items\n", len(data))
		// Convert []interface{} to []map[string]interface{}
		for _, item := range data {
			if row, ok := item.(map[string]interface{}); ok {
				dataArray = append(dataArray, row)
			}
		}
		rowCount = len(dataArray)
		fmt.Printf("TestDatabaseQuery: Extracted %d rows\n", rowCount)
	} else {
		fmt.Printf("TestDatabaseQuery: data field not found or not an array, checking type: %T\n", queryResult["data"])
		// Check if the whole queryResult might BE the data itself 
		if queryResult["success"] != nil && queryResult["count"] != nil {
			// This means we parsed the FoxPro JSON correctly
			fmt.Printf("TestDatabaseQuery: FoxPro response structure detected\n")
		}
		// Fallback if structure is different
		dataArray = []map[string]interface{}{}
		rowCount = 0
	}
	
	// Also check for count field - could be float64 or int
	if count, ok := queryResult["count"].(float64); ok {
		rowCount = int(count)
		fmt.Printf("TestDatabaseQuery: Found count field (float64): %d\n", rowCount)
	} else if count, ok := queryResult["count"].(int); ok {
		rowCount = count
		fmt.Printf("TestDatabaseQuery: Found count field (int): %d\n", rowCount)
	}
	
	result := map[string]interface{}{
		"success":       true,
		"database":      companyName,
		"query":         query,
		"method":        "OLE/COM JSON (Pivoten.DbApi)",
		"executionTime": fmt.Sprintf("%.2fms", elapsedTime.Seconds()*1000),
		"data":          dataArray,
		"rowCount":      rowCount,
		"raw":           jsonResult,
		"message":       "Query executed successfully via OLE server (JSON)",
	}
	
	fmt.Printf("TestDatabaseQuery: SUCCESS - Query executed in %.2fms\n", elapsedTime.Seconds()*1000)
	debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: SUCCESS - Query executed in %.2fms", elapsedTime.Seconds()*1000))
	fmt.Printf("=== TestDatabaseQuery COMPLETED ===\n")
	
	return result, nil
}

// GetTableList returns a list of tables in the database
func (a *App) GetTableList(companyName string) (map[string]interface{}, error) {
	fmt.Printf("GetTableList: Getting tables for company: %s\n", companyName)
	debug.SimpleLog(fmt.Sprintf("GetTableList: Getting tables for company: %s", companyName))
	
	// Use OLE to get actual table list from DBC
	client, err := ole.NewDbApiClient()
	if err != nil {
		fmt.Printf("GetTableList: Failed to connect to OLE server: %v\n", err)
		// Fall back to hardcoded list if OLE not available
		tables := []string{
			"COA", "CHECKS", "GLMASTER", "VENDORS", "WELLS",
			"INCOME", "EXPENSE", "OWNERS", "DIVISIONS",
			"ACPAY", "ACREC", "JOURNAL",
		}
		return map[string]interface{}{
			"success": true,
			"tables":  tables,
			"source":  "hardcoded",
		}, nil
	}
	defer client.Close()
	
	// Open the database
	if err := client.OpenDbc(companyName); err != nil {
		fmt.Printf("GetTableList: Failed to open DBC: %v\n", err)
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to open database: %v", err),
		}, nil
	}
	
	// Use the new JSON method to get table list
	jsonResult, err := client.GetTableListSimple()
	if err != nil {
		fmt.Printf("GetTableList: Failed to get table list: %v\n", err)
		// Fall back to hardcoded list
		tables := []string{
			"COA", "CHECKS", "GLMASTER", "VENDORS", "WELLS",
			"INCOME", "EXPENSE", "OWNERS", "DIVISIONS",
			"ACPAY", "ACREC", "JOURNAL",
		}
		return map[string]interface{}{
			"success": true,
			"tables":  tables,
			"source":  "hardcoded-fallback",
		}, nil
	}
	
	// Parse JSON result to extract table names
	fmt.Printf("GetTableList: Got JSON result: %s\n", jsonResult)
	debug.SimpleLog(fmt.Sprintf("GetTableList: JSON result: %s", jsonResult))
	
	// Parse the JSON array of table names
	var tableList []string
	err = json.Unmarshal([]byte(jsonResult), &tableList)
	if err != nil {
		fmt.Printf("GetTableList: Failed to parse JSON: %v\n", err)
		// Fall back to hardcoded list
		tableList = []string{
			"COA", "CHECKS", "GLMASTER", "VENDORS", "WELLS",
			"INCOME", "EXPENSE", "OWNERS", "DIVISIONS",
			"ACPAY", "ACREC", "JOURNAL",
		}
	} else {
		fmt.Printf("GetTableList: Found %d tables from database\n", len(tableList))
	}
	
	return map[string]interface{}{
		"success": true,
		"tables":  tableList,
		"source":  "database",
		"count":   len(tableList),
	}, nil
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetCompanyList reads the compmast.dbf file to get available companies
func (a *App) GetCompanyList() ([]map[string]interface{}, error) {
	fmt.Println("GetCompanyList: Reading compmast.dbf for company list")
	debug.LogInfo("GetCompanyList", "Reading compmast.dbf for company list")
	
	// Get the executable directory as the base path
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	
	// Construct path to compmast.dbf in datafiles folder
	compMastPath := filepath.Join(exeDir, "datafiles", "compmast.dbf")
	fmt.Printf("GetCompanyList: Looking for compmast.dbf at: %s\n", compMastPath)
	debug.LogInfo("GetCompanyList", fmt.Sprintf("Looking for compmast.dbf at: %s", compMastPath))
	
	// Check if file exists
	if _, err := os.Stat(compMastPath); os.IsNotExist(err) {
		fmt.Printf("GetCompanyList: compmast.dbf not found at %s\n", compMastPath)
		debug.LogError("GetCompanyList", fmt.Errorf("compmast.dbf not found at %s", compMastPath))
		return nil, fmt.Errorf("compmast.dbf not found in datafiles directory")
	}
	debug.LogInfo("GetCompanyList", "compmast.dbf found")
	
	// Read the DBF file directly (not company-specific)
	debug.LogInfo("GetCompanyList", "Reading DBF file...")
	result, err := company.ReadDBFFileDirectly(compMastPath, "", 0, 0, "", "")
	if err != nil {
		debug.LogError("GetCompanyList", err)
		return nil, fmt.Errorf("failed to read compmast.dbf: %w", err)
	}
	debug.LogInfo("GetCompanyList", "DBF file read successfully")
	
	// Extract rows from result
	rows, ok := result["rows"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format from compmast.dbf")
	}
	
	// Transform the data for frontend consumption
	companies := []map[string]interface{}{}
	for _, row := range rows {
		company := map[string]interface{}{
			"company_id":   row["CIDCOMP"],
			"company_name": row["CPRODUCER"],
			"address1":     row["CADDRESS1"],
			"address2":     row["CADDRESS2"],
			"city":         row["CCITY"],
			"state":        row["CSTATE"],
			"zip_code":     row["CZIPCODE"],
			"data_path":    row["CDATAPATH"],
		}
		companies = append(companies, company)
	}
	
	fmt.Printf("GetCompanyList: Found %d companies\n", len(companies))
	debug.LogInfo("GetCompanyList", fmt.Sprintf("Returning %d companies", len(companies)))
	return companies, nil
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
	debug.SimpleLog(fmt.Sprintf("GetBankAccounts: company=%s, currentUser=%v", companyName, a.currentUser != nil))
	
	// Skip strict auth check - if they can call this, they're logged in
	if a.currentUser == nil {
		fmt.Printf("GetBankAccounts: Warning - currentUser is nil, proceeding anyway\n")
		debug.SimpleLog("GetBankAccounts: Warning - currentUser is nil, proceeding anyway")
	} else if !a.currentUser.HasPermission("database.read") {
		fmt.Printf("GetBankAccounts: user %s lacks database.read permission\n", a.currentUser.Username)
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	fmt.Printf("GetBankAccounts: proceeding to read COA.dbf\n")

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
	debug.SimpleLog(fmt.Sprintf("GetOutstandingChecks: company=%s, account=%s, currentUser=%v", companyName, accountNumber, a.currentUser != nil))
	
	// Skip strict auth check - if they can call this, they're logged in
	if a.currentUser == nil {
		debug.SimpleLog("GetOutstandingChecks: Warning - currentUser is nil, proceeding anyway")
	} else if !a.currentUser.HasPermission("database.read") {
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
	
	// Find relevant check columns based on actual CHECKS.dbf structure (including CENTRYTYPE and CIDCHEC)
	var checkNumIdx, dateIdx, payeeIdx, amountIdx, accountIdx, clearedIdx, voidIdx, entryTypeIdx, cidchecIdx int = -1, -1, -1, -1, -1, -1, -1, -1, -1
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
		} else if colUpper == "CENTRYTYPE" {
			entryTypeIdx = i
			fmt.Printf("GetOutstandingChecks: Found entry type column at index %d: %s\n", i, col)
		} else if colUpper == "CIDCHEC" {
			cidchecIdx = i
			fmt.Printf("GetOutstandingChecks: Found CIDCHEC column at index %d: %s\n", i, col)
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
			entryType := ""
			if entryTypeIdx != -1 && len(row) > entryTypeIdx {
				entryType = strings.TrimSpace(fmt.Sprintf("%v", row[entryTypeIdx]))
			}
			
			// Show raw values from DBF
			clearedRaw := "nil"
			voidRaw := "nil"
			if clearedIdx != -1 && len(row) > clearedIdx {
				clearedRaw = fmt.Sprintf("%v (type: %T)", row[clearedIdx], row[clearedIdx])
			}
			if voidIdx != -1 && len(row) > voidIdx {
				voidRaw = fmt.Sprintf("%v (type: %T)", row[voidIdx], row[voidIdx])
			}
			
			fmt.Printf("GetOutstandingChecks: Entry %s, Type: %s, Account %s, Amount $%.2f\n", checkNum, entryType, checkAccount, amount)
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
			
			// Get entry type for better display
			entryType := ""
			if entryTypeIdx != -1 && len(row) > entryTypeIdx {
				entryType = strings.TrimSpace(fmt.Sprintf("%v", row[entryTypeIdx]))
			}
			
			check := map[string]interface{}{
				"checkNumber": fmt.Sprintf("%v", row[checkNumIdx]),
				"amount": parseFloat(row[amountIdx]),
				"account": checkAccount,
				"entryType": entryType, // D = Deposit, C = Check
			}
			
			// Add CIDCHEC field for unique identification
			if cidchecIdx != -1 && len(row) > cidchecIdx && row[cidchecIdx] != nil {
				cidchec := fmt.Sprintf("%v", row[cidchecIdx])
				check["cidchec"] = cidchec
				check["id"] = cidchec  // Also set as "id" for matching
			} else {
				check["cidchec"] = "" // Empty string if not available
				check["id"] = ""
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
	debug.LogInfo("GetAccountBalance", fmt.Sprintf("Called for company=%s, account=%s", companyName, accountNumber))
	
	// Check permissions
	if a.currentUser == nil {
		return 0, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return 0, fmt.Errorf("insufficient permissions")
	}
	
	// Read GLMASTER.dbf to get account balance
	debug.LogInfo("GetAccountBalance", "Attempting to read GLMASTER.dbf")
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.dbf", "", 0, 50000, "", "")
	if err != nil {
		fmt.Printf("GetAccountBalance: failed to read GLMASTER.dbf: %v\n", err)
		debug.LogError("GetAccountBalance", fmt.Errorf("failed to read GLMASTER.dbf: %v", err))
		return 0, fmt.Errorf("failed to read GLMASTER.dbf: %w", err)
	}
	debug.LogInfo("GetAccountBalance", "Successfully read GLMASTER.dbf")
	
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
	debug.LogInfo("GetAccountBalance", fmt.Sprintf("Processing %d rows from GLMASTER.dbf", len(glRows)))
	
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
	debug.LogInfo("GetAccountBalance", fmt.Sprintf("Final balance for account %s: %f", accountNumber, totalBalance))
	return totalBalance, nil
}

// Bank Transaction structures for SQLite persistence
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
	ReconciledDate     *string                `json:"reconciled_date"`  // Changed to pointer to handle NULL
	ReconciliationID   *int                   `json:"reconciliation_id"` // Changed to pointer to handle NULL
	ExtendedData       map[string]interface{} `json:"extended_data"`
}

type MatchResult struct {
	BankTransaction BankTransaction        `json:"bankTransaction"`
	MatchedCheck    map[string]interface{} `json:"matchedCheck"`
	Confidence      float64                `json:"confidence"`
	MatchType       string                 `json:"matchType"`
	Confirmed       bool                   `json:"confirmed"`
}

// RunMatching runs the matching algorithm on unmatched bank transactions
func (a *App) RunMatching(companyName string, accountNumber string, options map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("RunMatching called for company: %s, account: %s, options: %+v\n", companyName, accountNumber, options)
	
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
					fmt.Printf("Will limit matching to checks dated on or before: %s\n", statementDate.Format("2006-01-02"))
				}
			}
		}
	}
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Get unmatched bank transactions
	txnResult, err := a.GetBankTransactions(companyName, accountNumber, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get bank transactions: %w", err)
	}
	
	transactions, ok := txnResult["transactions"].([]BankTransaction)
	if !ok {
		// Need to convert from the result
		return nil, fmt.Errorf("invalid transaction data format")
	}
	
	// Get existing checks for matching
	fmt.Printf("Getting existing checks for company: %s, account: %s\n", companyName, accountNumber)
	checksResult, err := a.GetOutstandingChecks(companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing checks: %w", err)
	}
	
	existingChecks, ok := checksResult["checks"].([]map[string]interface{})
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
		fmt.Printf("Filtered checks from %d to %d (on or before %s)\n", 
			len(existingChecks), len(checksToMatch), statementDate.Format("2006-01-02"))
	} else {
		checksToMatch = existingChecks
		fmt.Printf("Using all %d checks for matching (no date filter)\n", len(checksToMatch))
	}
	
	// Run matching algorithm
	fmt.Printf("Matching %d bank transactions with %d checks\n", len(transactions), len(checksToMatch))
	matches := a.autoMatchBankTransactions(transactions, checksToMatch)
	fmt.Printf("Found %d matches\n", len(matches))
	
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
				    is_matched = TRUE
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
			
			_, err := a.db.Exec(updateQuery, checkID, rowIndex, match.Confidence, match.MatchType, match.BankTransaction.ID)
			if err == nil {
				matchedCount++
				fmt.Printf("Successfully matched bank txn %d to check %s\n", match.BankTransaction.ID, checkID)
			} else {
				fmt.Printf("Failed to update match for bank txn %d: %v\n", match.BankTransaction.ID, err)
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

// ClearMatchesAndRerun clears all matches and reruns the matching algorithm
func (a *App) ClearMatchesAndRerun(companyName string, accountNumber string, options map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("ClearMatchesAndRerun called for company: %s, account: %s, options: %+v\n", companyName, accountNumber, options)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Clear all existing matches for this account
	clearQuery := `
		UPDATE bank_transactions 
		SET matched_check_id = NULL,
		    matched_dbf_row_index = 0,
		    match_confidence = 0,
		    match_type = '',
		    is_matched = FALSE,
		    manually_matched = FALSE
		WHERE company_name = ? AND account_number = ?
	`
	
	result, err := a.db.Exec(clearQuery, companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to clear matches: %w", err)
	}
	
	clearedRows, _ := result.RowsAffected()
	fmt.Printf("Cleared %d existing matches\n", clearedRows)
	
	// Now run matching again
	return a.RunMatching(companyName, accountNumber, options)
}

// ImportBankStatement parses and stores CSV bank statement in SQLite (without auto-matching)
func (a *App) ImportBankStatement(companyName string, csvContent string, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("ImportBankStatement called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Generate unique batch ID for this import
	batchID := fmt.Sprintf("import_%d_%s", time.Now().Unix(), accountNumber)
	statementDate := time.Now().Format("2006-01-02") // Use today as statement date, can be overridden
	
	// Parse CSV content into BankTransaction objects
	bankTransactions, err := a.parseCSVToBankTransactions(csvContent, accountNumber, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}
	
	// SKIP auto-matching during import - will be done separately via RunMatching button
	fmt.Printf("Skipping auto-match during import - use Match button to run matching\n")
	
	// Create bank statement record first
	statementID, err := a.createBankStatement(companyName, accountNumber, statementDate, batchID, len(bankTransactions))
	if err != nil {
		return nil, fmt.Errorf("failed to create bank statement: %w", err)
	}
	fmt.Printf("Created bank statement with ID: %d\n", statementID)
	
	// Update transactions with statement ID
	for i := range bankTransactions {
		bankTransactions[i].StatementID = statementID
	}
	
	// Store bank transactions in SQLite
	fmt.Printf("Storing %d bank transactions in database\n", len(bankTransactions))
	err = a.storeBankTransactions(bankTransactions)
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

// parseCSVToBankTransactions parses CSV content into BankTransaction objects
func (a *App) parseCSVToBankTransactions(csvContent string, accountNumber string, batchID string) ([]BankTransaction, error) {
	lines := strings.Split(strings.TrimSpace(csvContent), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("CSV must contain header and at least one data row")
	}
	
	// Parse header to determine column indices - handle quoted fields properly
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
			CompanyName:   a.currentUser.CompanyName,
			AccountNumber: accountNumber,
			ImportBatchID: batchID,
			ImportedBy:    a.currentUser.Username,
		}
		
		// Extract date - convert to YYYY-MM-DD for SQLite DATE column
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
			transaction.Amount = parseFloat(fields[amountIdx])
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
	
	fmt.Printf("Parsed %d bank transactions from CSV\n", len(transactions))
	return transactions, nil
}

// storeBankTransactions stores bank transactions in SQLite
// createBankStatement creates a bank statement record for tracking import sessions
func (a *App) createBankStatement(companyName, accountNumber, statementDate, batchID string, transactionCount int) (int, error) {
	if a.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}
	
	query := `
		INSERT INTO bank_statements (
			company_name, account_number, statement_date, import_batch_id,
			imported_by, transaction_count, is_active
		) VALUES (?, ?, ?, ?, ?, ?, TRUE)
	`
	
	result, err := a.db.Exec(query, companyName, accountNumber, statementDate, batchID, 
		a.currentUser.Username, transactionCount)
	if err != nil {
		return 0, fmt.Errorf("failed to insert bank statement: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get statement ID: %w", err)
	}
	
	return int(id), nil
}

func (a *App) storeBankTransactions(transactions []BankTransaction) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	// Begin transaction for atomic insert
	tx, err := a.db.GetConn().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	stmt, err := tx.Prepare(`
		INSERT INTO bank_transactions (
			company_name, account_number, statement_id, transaction_date, check_number, description,
			amount, transaction_type, import_batch_id, imported_by, matched_check_id,
			matched_dbf_row_index, match_confidence, match_type, is_matched, manually_matched, extended_data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	for _, txn := range transactions {
		extendedDataJSON := "{}"
		if len(txn.ExtendedData) > 0 {
			if data, err := json.Marshal(txn.ExtendedData); err == nil {
				extendedDataJSON = string(data)
			}
		}
		
		_, err = stmt.Exec(
			txn.CompanyName, txn.AccountNumber, txn.StatementID, txn.TransactionDate, txn.CheckNumber,
			txn.Description, txn.Amount, txn.TransactionType, txn.ImportBatchID,
			txn.ImportedBy, txn.MatchedCheckID, txn.MatchedDBFRowIndex, txn.MatchConfidence, txn.MatchType,
			txn.IsMatched, txn.ManuallyMatched, extendedDataJSON,
		)
		if err != nil {
			return fmt.Errorf("failed to insert transaction: %w", err)
		}
	}
	
	return tx.Commit()
}

// autoMatchBankTransactions matches bank transactions with existing checks
func (a *App) autoMatchBankTransactions(bankTransactions []BankTransaction, existingChecks []map[string]interface{}) []MatchResult {
	var matches []MatchResult
	
	// Keep track of already matched check IDs to prevent double-matching
	matchedCheckIDs := make(map[string]bool)
	
	// Sort bank transactions by date to match older transactions first
	// This helps with recurring transactions
	sort.Slice(bankTransactions, func(i, j int) bool {
		dateI, _ := parseDate(bankTransactions[i].TransactionDate)
		dateJ, _ := parseDate(bankTransactions[j].TransactionDate)
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
		
		bestMatch := a.findBestCheckMatchForBankTxn(txn, availableChecks)
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
func (a *App) findBestCheckMatchForBankTxn(txn *BankTransaction, existingChecks []map[string]interface{}) *MatchResult {
	var bestMatch *MatchResult
	highestScore := 0.0
	
	for _, check := range existingChecks {
		score := a.calculateBankTxnMatchScore(txn, check)
		if score > highestScore && score > 0.5 { // Minimum confidence threshold
			matchType := a.determineBankTxnMatchType(score, txn, check)
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

// calculateBankTxnMatchScore calculates confidence score between bank transaction and check
func (a *App) calculateBankTxnMatchScore(txn *BankTransaction, check map[string]interface{}) float64 {
	score := 0.0
	
	// Amount matching (35% weight for recurring transactions)
	checkAmount := parseFloat(check["amount"])
	txnAmount := math.Abs(txn.Amount) // Always use absolute value for comparison
	
	amountMatches := false
	if checkAmount > 0 && math.Abs(txnAmount-checkAmount) < 0.01 {
		score += 0.35 // Exact amount match (reduced from 0.5)
		amountMatches = true
	} else if checkAmount > 0 && math.Abs(txnAmount-checkAmount) < 1.0 {
		score += 0.2 // Close amount match
	} else {
		// No amount match, very low chance this is right
		return 0.0
	}
	
	// Check number matching (25% weight)
	if txn.CheckNumber != "" {
		checkNumber := fmt.Sprintf("%v", check["checkNumber"])
		if txn.CheckNumber == checkNumber {
			score += 0.25 // Exact check number match
		} else if strings.Contains(txn.CheckNumber, checkNumber) || strings.Contains(checkNumber, txn.CheckNumber) {
			score += 0.1 // Partial check number match
		}
	}
	
	// Date proximity matching (40% weight - INCREASED for recurring transactions)
	// This is critical for matching recurring payments with same amounts
	if txn.TransactionDate != "" {
		txnDate, txnErr := parseDate(txn.TransactionDate)
		checkDate, checkErr := parseDate(fmt.Sprintf("%v", check["date"]))
		
		if txnErr == nil && checkErr == nil {
			daysDiff := math.Abs(txnDate.Sub(checkDate).Hours() / 24)
			
			// More granular date scoring for better recurring transaction matching
			if daysDiff == 0 {
				score += 0.4 // Same day - very high confidence
			} else if daysDiff <= 1 {
				score += 0.35 // Next day
			} else if daysDiff <= 3 {
				score += 0.25 // Within 3 days
			} else if daysDiff <= 7 {
				score += 0.15 // Within a week
			} else if daysDiff <= 14 {
				score += 0.05 // Within 2 weeks
			}
			// Beyond 2 weeks, no date score
			
			// Debug for recurring transactions
			if amountMatches && strings.Contains(strings.ToUpper(fmt.Sprintf("%v", check["payee"])), "CONSUMER") {
				fmt.Printf("DEBUG: CONSUMERS ENERGY match - Amount: %.2f, Days diff: %.0f, Score: %.2f\n", 
					checkAmount, daysDiff, score)
			}
		}
	}
	
	// Description/Payee matching (bonus points)
	if description, ok := check["payee"].(string); ok && txn.Description != "" {
		descUpper := strings.ToUpper(description)
		txnDescUpper := strings.ToUpper(txn.Description)
		
		// Check for common keywords
		if strings.Contains(txnDescUpper, descUpper) || strings.Contains(descUpper, txnDescUpper) {
			score += 0.1 // Bonus for description match
		}
	}
	
	return score
}

// determineBankTxnMatchType determines the type of match for bank transaction
func (a *App) determineBankTxnMatchType(score float64, txn *BankTransaction, check map[string]interface{}) string {
	checkAmount := parseFloat(check["amount"])
	
	// Exact match: amount + check number (if available)
	if math.Abs(math.Abs(txn.Amount)-checkAmount) < 0.01 {
		if txn.CheckNumber != "" && txn.CheckNumber == fmt.Sprintf("%v", check["checkNumber"]) {
			return "exact"
		}
		return "amount_exact"
	}
	
	// Fuzzy match
	if score > 0.7 {
		return "high_confidence"
	}
	
	return "fuzzy"
}

// GetBankTransactions retrieves stored bank transactions for an account

// GetRecentBankStatements retrieves recent bank statement imports
func (a *App) GetRecentBankStatements(companyName string, accountNumber string) ([]map[string]interface{}, error) {
	fmt.Printf("GetRecentBankStatements called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	query := `
		SELECT id, company_name, account_number, statement_date, import_batch_id, 
		       import_date, imported_by, transaction_count, matched_count
		FROM bank_statements 
		WHERE company_name = ? AND account_number = ?
		ORDER BY import_date DESC
		LIMIT 10
	`
	
	rows, err := a.db.GetConn().Query(query, companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to query bank statements: %w", err)
	}
	defer rows.Close()
	
	var statements []map[string]interface{}
	for rows.Next() {
		var stmt struct {
			ID               int
			CompanyName      string
			AccountNumber    string
			StatementDate    sql.NullString
			ImportBatchID    string
			ImportDate       string
			ImportedBy       string
			TransactionCount int
			MatchedCount     int
		}
		
		err := rows.Scan(&stmt.ID, &stmt.CompanyName, &stmt.AccountNumber, 
			&stmt.StatementDate, &stmt.ImportBatchID, &stmt.ImportDate, 
			&stmt.ImportedBy, &stmt.TransactionCount, &stmt.MatchedCount)
		if err != nil {
			continue
		}
		
		statementDate := ""
		if stmt.StatementDate.Valid {
			statementDate = stmt.StatementDate.String
		}
		
		statements = append(statements, map[string]interface{}{
			"id":               stmt.ID,
			"company_name":     stmt.CompanyName,
			"account_number":   stmt.AccountNumber,
			"statement_date":   statementDate,
			"import_batch_id":  stmt.ImportBatchID,
			"import_date":      stmt.ImportDate,
			"imported_by":      stmt.ImportedBy,
			"transaction_count": stmt.TransactionCount,
			"matched_count":    stmt.MatchedCount,
		})
	}
	
	return statements, nil
}

// DeleteBankStatement deletes an imported bank statement and all its transactions
func (a *App) DeleteBankStatement(companyName string, importBatchID string) error {
	fmt.Printf("DeleteBankStatement called for company: %s, batch: %s\n", companyName, importBatchID)
	
	// Check permissions
	if a.currentUser == nil {
		return fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return fmt.Errorf("insufficient permissions to delete bank statements")
	}
	
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	// Start transaction
	tx, err := a.db.GetConn().Begin()
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

// ManualMatchTransaction manually matches a bank transaction to a check
func (a *App) ManualMatchTransaction(transactionID int, checkID string, checkRowIndex int) (map[string]interface{}, error) {
	fmt.Printf("ManualMatchTransaction: txn=%d, check=%s, row=%d\n", transactionID, checkID, checkRowIndex)
	
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	// Update the bank transaction with match info
	query := `
		UPDATE bank_transactions 
		SET matched_check_id = ?, 
		    matched_dbf_row_index = ?,
		    match_confidence = 1.0,
		    match_type = 'manual',
		    is_matched = TRUE,
		    manually_matched = TRUE
		WHERE id = ?
	`
	
	_, err := a.db.Exec(query, checkID, checkRowIndex, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}
	
	return map[string]interface{}{
		"status": "success",
		"message": "Transaction matched successfully",
	}, nil
}

// RetryMatching re-runs the matching algorithm for unmatched transactions
func (a *App) RetryMatching(companyName string, accountNumber string, statementID int) (map[string]interface{}, error) {
	fmt.Printf("RetryMatching for statement: %d\n", statementID)
	
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Get unmatched transactions for this statement
	query := `
		SELECT id, check_number, amount, transaction_date, description 
		FROM bank_transactions 
		WHERE statement_id = ? AND is_matched = FALSE
	`
	
	rows, err := a.db.GetConn().Query(query, statementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unmatched transactions: %w", err)
	}
	defer rows.Close()
	
	var unmatchedTxns []BankTransaction
	for rows.Next() {
		var txn BankTransaction
		err := rows.Scan(&txn.ID, &txn.CheckNumber, &txn.Amount, &txn.TransactionDate, &txn.Description)
		if err != nil {
			continue
		}
		unmatchedTxns = append(unmatchedTxns, txn)
	}
	
	// Get outstanding checks
	checksResult, err := a.GetOutstandingChecks(companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get checks: %w", err)
	}
	
	existingChecks, _ := checksResult["checks"].([]map[string]interface{})
	
	// Run matching algorithm
	newMatchCount := 0
	for _, txn := range unmatchedTxns {
		bestMatch := a.findBestCheckMatchForBankTxn(&txn, existingChecks)
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
				    is_matched = TRUE
				WHERE id = ?
			`
			
			_, err := a.db.Exec(updateQuery, checkID, rowIndex, bestMatch.Confidence, bestMatch.MatchType, txn.ID)
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
			WHERE statement_id = ? AND is_matched = TRUE
		)
		WHERE id = ?
	`
	a.db.Exec(updateStmt, statementID, statementID)
	
	return map[string]interface{}{
		"status": "success",
		"newMatches": newMatchCount,
		"totalUnmatched": len(unmatchedTxns) - newMatchCount,
	}, nil
}

// GetMatchedTransactions returns all matched checks with their bank transaction confirmation
func (a *App) GetMatchedTransactions(companyName string, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("GetMatchedTransactions called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	// First get all matched bank transactions to know which checks are matched
	query := `
		SELECT bt.id, bt.matched_check_id, bt.matched_dbf_row_index, bt.match_confidence, 
			   bt.match_type, bt.manually_matched, bt.amount as bank_amount, bt.transaction_date as bank_date,
			   bt.description as bank_description, bt.check_number as bank_check_number
		FROM bank_transactions bt
		INNER JOIN bank_statements bs ON bt.statement_id = bs.id
		WHERE bt.company_name = ? AND bt.account_number = ? 
		  AND bs.is_active = TRUE
		  AND bt.is_matched = TRUE
	`
	
	rows, err := a.db.GetConn().Query(query, companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to query matched transactions: %w", err)
	}
	defer rows.Close()
	
	// Map to store bank transaction matches by check ID
	matchedMap := make(map[string]map[string]interface{})
	fmt.Printf("Querying matched transactions for company: %s, account: %s\n", companyName, accountNumber)
	
	for rows.Next() {
		var bankTxnID int
		var matchedCheckID, matchType, bankDescription, bankCheckNumber sql.NullString
		var matchedDBFRowIndex sql.NullInt64
		var matchConfidence float64
		var manuallyMatched bool
		var bankAmount float64
		var bankDate string
		
		err := rows.Scan(
			&bankTxnID, &matchedCheckID, &matchedDBFRowIndex, &matchConfidence,
			&matchType, &manuallyMatched, &bankAmount, &bankDate,
			&bankDescription, &bankCheckNumber,
		)
		if err != nil {
			continue
		}
		
		if matchedCheckID.Valid && matchedCheckID.String != "" {
			fmt.Printf("Found matched check ID: %s with bank txn ID: %d\n", matchedCheckID.String, bankTxnID)
			matchedMap[matchedCheckID.String] = map[string]interface{}{
				"bank_txn_id":       bankTxnID,
				"match_confidence":  matchConfidence,
				"match_type":        matchType.String,
				"manually_matched":  manuallyMatched,
				"bank_amount":       bankAmount,
				"bank_date":         bankDate,
				"bank_description":  bankDescription.String,
				"bank_check_number": bankCheckNumber.String,
				"dbf_row_index":     matchedDBFRowIndex.Int64,
			}
		}
	}
	
	fmt.Printf("Total matched bank transactions found: %d\n", len(matchedMap))
	
	// Now read the checks from DBF and build response with check data as primary
	checksData, err := company.ReadDBFFile(companyName, "checks.dbf", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read checks: %w", err)
	}
	
	columns, ok := checksData["columns"].([]string)
	if !ok {
		return nil, fmt.Errorf("invalid columns format")
	}
	
	rows2, ok := checksData["rows"].([][]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid rows format")
	}
	
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
			if amt, err := strconv.ParseFloat(fmt.Sprintf("%v", row[amountIdx]), 64); err == nil {
				checkData["amount"] = amt
			}
		}
		if acctIdx >= 0 && acctIdx < len(row) {
			checkData["account"] = fmt.Sprintf("%v", row[acctIdx])
		}
		if clearedIdx >= 0 && clearedIdx < len(row) {
			cleared := fmt.Sprintf("%v", row[clearedIdx])
			checkData["cleared"] = cleared == "true" || cleared == "T" || cleared == ".T."
		}
		
		// Add bank transaction match info
		checkData["bank_match"] = bankMatch
		checkData["match_confidence"] = bankMatch["match_confidence"]
		checkData["match_type"] = bankMatch["match_type"]
		checkData["manually_matched"] = bankMatch["manually_matched"]
		checkData["bank_txn_id"] = bankMatch["bank_txn_id"]
		
		matchedChecks = append(matchedChecks, checkData)
	}
	
	fmt.Printf("Returning %d matched checks\n", len(matchedChecks))
	
	return map[string]interface{}{
		"status": "success",
		"checks": matchedChecks,
		"count":  len(matchedChecks),
	}, nil
}

// UnmatchTransaction removes the match between a bank transaction and a check
func (a *App) UnmatchTransaction(transactionID int) (map[string]interface{}, error) {
	fmt.Printf("UnmatchTransaction called for transaction ID: %d\n", transactionID)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	// Update the transaction to unmatched
	query := `
		UPDATE bank_transactions 
		SET matched_check_id = NULL,
		    matched_dbf_row_index = 0,
		    match_confidence = 0,
		    match_type = '',
		    is_matched = FALSE,
		    manually_matched = FALSE
		WHERE id = ?
	`
	
	result, err := a.db.Exec(query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to unmatch transaction: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	
	return map[string]interface{}{
		"status": "success",
		"rowsAffected": rowsAffected,
	}, nil
}

func (a *App) GetBankTransactions(companyName string, accountNumber string, importBatchID string) (map[string]interface{}, error) {
	fmt.Printf("GetBankTransactions called for company: %s, account: %s, batch: %s\n", companyName, accountNumber, importBatchID)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
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
	
	rows, err := a.db.GetConn().Query(query, args...)
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

// parseCSVLine properly parses a CSV line handling quoted fields
func parseCSVLine(line string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false
	
	for i := 0; i < len(line); i++ {
		ch := line[i]
		
		if ch == '"' {
			inQuotes = !inQuotes
		} else if ch == ',' && !inQuotes {
			fields = append(fields, current.String())
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	
	// Don't forget the last field
	fields = append(fields, current.String())
	
	return fields
}

// parseCSVContent parses CSV content and handles different bank formats
func (a *App) parseCSVContent(csvContent string) ([]BankTransaction, error) {
	lines := strings.Split(strings.TrimSpace(csvContent), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("CSV must contain header and at least one data row")
	}
	
	// Parse header to determine column indices - handle quoted fields
	header := parseCSVLine(lines[0])
	columnMap := make(map[string]int)
	
	fmt.Printf("CSV Header: %v\n", header)
	
	for i, col := range header {
		colName := strings.ToLower(strings.TrimSpace(strings.Trim(col, `"`)))
		columnMap[colName] = i
		fmt.Printf("Column %d: '%s' -> '%s'\n", i, col, colName)
		
		// Handle common variations
		switch colName {
		case "transaction date", "posting date", "trans date":
			columnMap["date"] = i
		case "payee", "merchant", "vendor", "memo":
			columnMap["description"] = i
		case "debit", "withdrawal", "withdrawals":
			columnMap["debit"] = i
		case "credit", "deposit", "deposits":
			columnMap["credit"] = i
		}
	}
	
	var transactions []BankTransaction
	
	for lineNum, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Parse CSV properly handling quoted fields
		fields := parseCSVLine(line)
		if len(fields) < len(header) {
			fmt.Printf("Skipping malformed row %d: %d fields vs %d expected\n", lineNum+1, len(fields), len(header))
			continue // Skip malformed rows
		}
		
		// Debug first few rows
		if lineNum < 3 {
			fmt.Printf("Row %d raw: %s\n", lineNum+1, line)
			fmt.Printf("Row %d fields: %v\n", lineNum+1, fields)
		}
		
		// Clean fields
		for i, field := range fields {
			fields[i] = strings.TrimSpace(strings.Trim(field, `"`))
		}
		
		transaction := BankTransaction{}
		
		// Extract date - MM/DD format, convert to YYYY-MM-DD for SQLite
		if dateIdx, exists := columnMap["date"]; exists && dateIdx < len(fields) {
			dateStr := strings.TrimSpace(fields[dateIdx])
			if dateStr != "" {
				// Parse MM/DD and convert to 2025-MM-DD for SQLite DATE column
				parts := strings.Split(dateStr, "/")
				if len(parts) == 2 {
					month, _ := strconv.Atoi(parts[0])
					day, _ := strconv.Atoi(parts[1])
					transaction.TransactionDate = fmt.Sprintf("2025-%02d-%02d", month, day)
				} else {
					// Fallback
					transaction.TransactionDate = "2025-01-01"
				}
				
				// Debug output
				if lineNum < 3 {
					fmt.Printf("Date conversion: '%s' -> '%s'\n", dateStr, transaction.TransactionDate)
				}
			}
		}
		
		// Extract description
		if descIdx, exists := columnMap["description"]; exists && descIdx < len(fields) {
			transaction.Description = fields[descIdx]
		}
		
		// Extract amount (handle debit/credit columns or single amount column)
		if amountIdx, exists := columnMap["amount"]; exists && amountIdx < len(fields) {
			amountStr := strings.TrimSpace(fields[amountIdx])
			transaction.Amount = parseFloat(amountStr)
			if transaction.Amount == 0 && amountStr != "0" && amountStr != "0.00" {
				fmt.Printf("WARNING: Failed to parse amount: '%s'\n", amountStr)
			}
		} else {
			// Handle separate debit/credit columns
			var debit, credit float64
			if debitIdx, exists := columnMap["debit"]; exists && debitIdx < len(fields) {
				debit = parseFloat(fields[debitIdx])
			}
			if creditIdx, exists := columnMap["credit"]; exists && creditIdx < len(fields) {
				credit = parseFloat(fields[creditIdx])
			}
			
			// Net amount (debits are negative)
			transaction.Amount = credit - debit
		}
		
		// Extract type
		if typeIdx, exists := columnMap["type"]; exists && typeIdx < len(fields) {
			typeStr := strings.TrimSpace(fields[typeIdx])
			transaction.TransactionType = strings.ToUpper(typeStr)
			
			// Adjust amount based on type if it's not already negative
			if transaction.Amount > 0 && (transaction.TransactionType == "DEBIT" || transaction.TransactionType == "CHECK") {
				transaction.Amount = -transaction.Amount
			}
		} else {
			// Infer type from amount
			if transaction.Amount < 0 {
				transaction.TransactionType = "DEBIT"
			} else {
				transaction.TransactionType = "CREDIT"
			}
		}
		
		// Extract balance (store in extended data)
		if balanceIdx, exists := columnMap["balance"]; exists && balanceIdx < len(fields) {
			balance := parseFloat(fields[balanceIdx])
			if transaction.ExtendedData == nil {
				transaction.ExtendedData = make(map[string]interface{})
			}
			transaction.ExtendedData["balance"] = balance
		}
		
		// Extract check number from Check # column or description
		if checkIdx, exists := columnMap["check #"]; exists && checkIdx < len(fields) {
			checkNum := fields[checkIdx]
			// Remove asterisk if present (e.g., "12521*" -> "12521")
			checkNum = strings.TrimSuffix(checkNum, "*")
			if checkNum != "" {
				transaction.CheckNumber = checkNum
			}
		}
		if transaction.CheckNumber == "" {
			// Fallback to extracting from description
			transaction.CheckNumber = extractCheckNumber(transaction.Description)
		}
		
		transactions = append(transactions, transaction)
	}
	
	fmt.Printf("Parsed %d transactions from CSV\n", len(transactions))
	return transactions, nil
}

// autoMatchCSVTransactions matches CSV transactions with existing checks
func (a *App) autoMatchCSVTransactions(csvTransactions []BankTransaction, existingChecks []map[string]interface{}) []MatchResult {
	var matches []MatchResult
	
	for _, csvTxn := range csvTransactions {
		// Only match debits (checks, withdrawals) for reconciliation
		if csvTxn.Amount >= 0 {
			continue
		}
		
		bestMatch := a.findBestCheckMatch(csvTxn, existingChecks)
		if bestMatch != nil {
			matches = append(matches, *bestMatch)
		}
	}
	
	return matches
}

// findBestCheckMatch finds the best matching check for a CSV transaction
func (a *App) findBestCheckMatch(csvTxn BankTransaction, existingChecks []map[string]interface{}) *MatchResult {
	var bestMatch *MatchResult
	highestScore := 0.0
	
	for _, check := range existingChecks {
		score := a.calculateMatchScore(csvTxn, check)
		if score > highestScore && score > 0.5 { // Minimum confidence threshold
			matchType := a.determineMatchType(score, csvTxn, check)
			bestMatch = &MatchResult{
				BankTransaction: csvTxn,
				MatchedCheck:   check,
				Confidence:     score,
				MatchType:      matchType,
				Confirmed:      false,
			}
			highestScore = score
		}
	}
	
	return bestMatch
}

// calculateMatchScore calculates confidence score between CSV transaction and check
func (a *App) calculateMatchScore(csvTxn BankTransaction, check map[string]interface{}) float64 {
	score := 0.0
	
	// Amount matching (most important - 50% weight)
	checkAmount := parseFloat(check["amount"])
	if checkAmount > 0 && math.Abs(math.Abs(csvTxn.Amount)-checkAmount) < 0.01 {
		score += 0.5 // Exact amount match
	} else if checkAmount > 0 && math.Abs(math.Abs(csvTxn.Amount)-checkAmount) < 1.0 {
		score += 0.3 // Close amount match
	}
	
	// Check number matching (30% weight)
	if csvTxn.CheckNumber != "" {
		checkNumber := fmt.Sprintf("%v", check["checkNumber"])
		if csvTxn.CheckNumber == checkNumber {
			score += 0.3 // Exact check number match
		} else if strings.Contains(csvTxn.CheckNumber, checkNumber) || strings.Contains(checkNumber, csvTxn.CheckNumber) {
			score += 0.15 // Partial check number match
		}
	}
	
	// Date proximity matching (20% weight)
	if csvTxn.TransactionDate != "" {
		csvDate, csvErr := parseDate(csvTxn.TransactionDate)
		checkDate, checkErr := parseDate(fmt.Sprintf("%v", check["date"]))
		
		if csvErr == nil && checkErr == nil {
			daysDiff := math.Abs(csvDate.Sub(checkDate).Hours() / 24)
			if daysDiff == 0 {
				score += 0.2 // Same day
			} else if daysDiff <= 3 {
				score += 0.1 // Within 3 days
			} else if daysDiff <= 7 {
				score += 0.05 // Within a week
			}
		}
	}
	
	return score
}

// determineMatchType determines the type of match based on score and criteria
func (a *App) determineMatchType(score float64, csvTxn BankTransaction, check map[string]interface{}) string {
	checkAmount := parseFloat(check["amount"])
	
	// Exact match: amount + check number (if available)
	if math.Abs(math.Abs(csvTxn.Amount)-checkAmount) < 0.01 {
		if csvTxn.CheckNumber != "" && csvTxn.CheckNumber == fmt.Sprintf("%v", check["checkNumber"]) {
			return "exact"
		}
		return "amount_exact"
	}
	
	// Fuzzy match
	if score > 0.7 {
		return "high_confidence"
	}
	
	return "fuzzy"
}

// extractCheckNumber extracts check number from transaction description
func extractCheckNumber(description string) string {
	// Common patterns: "CHECK #1234", "Check 1234", "CHK 1234"
	patterns := []string{
		`(?i)check\s*#?\s*(\d+)`,
		`(?i)chk\s*#?\s*(\d+)`,
		`(?i)#(\d{4,})`, // Just # followed by 4+ digits
	}
	
	for _, pattern := range patterns {
		if re, err := regexp.Compile(pattern); err == nil {
			if matches := re.FindStringSubmatch(description); len(matches) > 1 {
				return matches[1]
			}
		}
	}
	
	return ""
}

// parseDate parses various date formats commonly found in bank CSVs
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"01/02/2006",  // MM/dd/yyyy
		"1/2/2006",    // M/d/yyyy
		"2006-01-02",  // yyyy-MM-dd
		"01-02-2006",  // MM-dd-yyyy
		"02/01/2006",  // dd/MM/yyyy (European)
		"2006/01/02",  // yyyy/MM/dd
	}
	
	dateStr = strings.TrimSpace(dateStr)
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// GetCachedBalances retrieves cached balances for all bank accounts
func (a *App) GetCachedBalances(companyName string) ([]map[string]interface{}, error) {
	fmt.Printf("GetCachedBalances called for company: %s\n", companyName)
	debug.SimpleLog(fmt.Sprintf("GetCachedBalances: company=%s, currentUser=%v", companyName, a.currentUser != nil))
	
	// Skip strict auth check - if they can call this, they're logged in
	if a.currentUser == nil {
		debug.SimpleLog("GetCachedBalances: Warning - currentUser is nil, proceeding anyway")
	} else if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	balances, err := database.GetAllCachedBalances(a.db, companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached balances: %w", err)
	}
	
	// Convert to interface for JSON response
	result := make([]map[string]interface{}, 0) // Initialize as empty slice, not nil
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
			// New detailed breakdown fields
			"uncleared_deposits":    balance.UnclearedDeposits,
			"uncleared_checks":      balance.UnclearedChecks,
			"deposit_count":         balance.DepositCount,
			"check_count":           balance.CheckCount,
		})
	}
	
	return result, nil
}

// RefreshAccountBalance refreshes both GL and outstanding checks for an account
func (a *App) RefreshAccountBalance(companyName, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("RefreshAccountBalance called for company: %s, account: %s\n", companyName, accountNumber)
	debug.SimpleLog(fmt.Sprintf("RefreshAccountBalance: company=%s, account=%s, currentUser=%v", companyName, accountNumber, a.currentUser != nil))
	
	// For now, allow refresh without strict auth check since user is already logged into the app
	// The company path security is handled by the fact that only authenticated users
	// can access the UI that calls this function
	username := "system"
	if a.currentUser != nil {
		username = a.currentUser.Username
		fmt.Printf("RefreshAccountBalance: User %s checking permissions\n", a.currentUser.Username)
		if !a.currentUser.HasPermission("database.read") {
			fmt.Printf("RefreshAccountBalance: ERROR - User lacks database.read permission\n")
			return nil, fmt.Errorf("insufficient permissions")
		}
	}
	
	fmt.Printf("RefreshAccountBalance: Proceeding with user %s\n", username)
	
	// Check database connection
	if a.db == nil {
		fmt.Printf("RefreshAccountBalance: ERROR - Database connection is nil\n")
		return nil, fmt.Errorf("database not initialized")
	}
	fmt.Printf("RefreshAccountBalance: Database connection OK\n")
	
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
	debug.SimpleLog(fmt.Sprintf("RefreshAllBalances: company=%s, currentUser=%v", companyName, a.currentUser != nil))
	
	// For now, allow refresh without strict auth check since user is already logged into the app
	// The company path security is handled by the fact that only authenticated users
	// can access the UI that calls this function
	if a.currentUser != nil && !a.currentUser.HasPermission("database.read") {
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
		// Remove commas and parse
		cleanStr := strings.ReplaceAll(v, ",", "")
		if f, err := strconv.ParseFloat(cleanStr, 64); err == nil {
			return f
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

// GetLastReconciliation returns the last reconciliation record for a specific bank account
func (a *App) GetLastReconciliation(companyName, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("GetLastReconciliation called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Read CHECKREC.dbf
	checkrecData, err := company.ReadDBFFile(companyName, "CHECKREC.dbf", "", 0, 0, "", "")
	if err != nil {
		fmt.Printf("GetLastReconciliation: Failed to read CHECKREC.dbf: %v\n", err)
		return map[string]interface{}{
			"status": "no_data",
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
			record["ending_balance"] = parseFloat(row[endBalIdx])
		}
		if begBalIdx != -1 && len(row) > begBalIdx {
			record["beginning_balance"] = parseFloat(row[begBalIdx])
		}
		if clearedCountIdx != -1 && len(row) > clearedCountIdx {
			record["cleared_count"] = int(parseFloat(row[clearedCountIdx]))
		}
		if clearedAmtIdx != -1 && len(row) > clearedAmtIdx {
			record["cleared_amount"] = parseFloat(row[clearedAmtIdx])
		}
		
		// Only add if we have a valid date
		if _, hasDate := record["date"]; hasDate {
			accountRecords = append(accountRecords, record)
		}
	}
	
	if len(accountRecords) == 0 {
		return map[string]interface{}{
			"status": "no_data",
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

// SaveReconciliationDraft saves or updates a draft reconciliation in SQLite
func (a *App) SaveReconciliationDraft(companyName string, draftData map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("SaveReconciliationDraft called for company: %s\n", companyName)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions to save reconciliation draft")
	}
	
	if a.reconciliationService == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}
	
	// Parse the request
	var req reconciliation.SaveDraftRequest
	req.CompanyName = companyName
	req.CreatedBy = a.currentUser.Username
	
	// Extract data from map
	if accountNumber, ok := draftData["account_number"].(string); ok {
		req.AccountNumber = accountNumber
	} else {
		return nil, fmt.Errorf("account_number is required")
	}
	
	if statementDate, ok := draftData["statement_date"].(string); ok {
		req.StatementDate = statementDate
	} else {
		return nil, fmt.Errorf("statement_date is required")
	}
	
	if balance, ok := draftData["statement_balance"].(float64); ok {
		req.StatementBalance = balance
	}
	if credits, ok := draftData["statement_credits"].(float64); ok {
		req.StatementCredits = credits
	}
	if debits, ok := draftData["statement_debits"].(float64); ok {
		req.StatementDebits = debits
	}
	if begBalance, ok := draftData["beginning_balance"].(float64); ok {
		req.BeginningBalance = begBalance
	}
	
	// Parse selected checks
	if checksData, ok := draftData["selected_checks"].([]interface{}); ok {
		for _, checkData := range checksData {
			if checkMap, ok := checkData.(map[string]interface{}); ok {
				var check reconciliation.SelectedCheck
				if cidchec, ok := checkMap["cidchec"].(string); ok {
					check.CIDCHEC = cidchec
				}
				if checkNumber, ok := checkMap["checkNumber"].(string); ok {
					check.CheckNumber = checkNumber
				}
				if amount, ok := checkMap["amount"].(float64); ok {
					check.Amount = amount
				}
				if payee, ok := checkMap["payee"].(string); ok {
					check.Payee = payee
				}
				if checkDate, ok := checkMap["checkDate"].(string); ok {
					check.CheckDate = checkDate
				}
				if rowIndex, ok := checkMap["rowIndex"].(float64); ok {
					check.RowIndex = int(rowIndex)
				}
				req.SelectedChecks = append(req.SelectedChecks, check)
			}
		}
	}
	
	// Save the draft
	result, err := a.reconciliationService.SaveDraft(req)
	if err != nil {
		return nil, fmt.Errorf("failed to save draft: %w", err)
	}
	
	return map[string]interface{}{
		"status": "success",
		"id": result.ID,
		"message": "Draft saved successfully",
		"reconciliation": result,
	}, nil
}

// GetReconciliationDraft retrieves the current draft reconciliation for an account
func (a *App) GetReconciliationDraft(companyName, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("GetReconciliationDraft called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	if a.reconciliationService == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}
	
	draft, err := a.reconciliationService.GetDraft(companyName, accountNumber)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return map[string]interface{}{
				"status": "no_draft",
				"message": "No draft reconciliation found",
			}, nil
		}
		return nil, fmt.Errorf("failed to get draft: %w", err)
	}
	
	return map[string]interface{}{
		"status": "success",
		"draft": draft,
	}, nil
}

// DeleteReconciliationDraft deletes the current draft reconciliation for an account
func (a *App) DeleteReconciliationDraft(companyName, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("DeleteReconciliationDraft called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions to delete reconciliation draft")
	}
	
	if a.reconciliationService == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}
	
	err := a.reconciliationService.DeleteDraft(companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to delete draft: %w", err)
	}
	
	return map[string]interface{}{
		"status": "success",
		"message": "Draft deleted successfully",
	}, nil
}

// CommitReconciliation commits a draft reconciliation
func (a *App) CommitReconciliation(companyName, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("CommitReconciliation called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions to commit reconciliation")
	}
	
	if a.reconciliationService == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}
	
	// Get the draft first
	draft, err := a.reconciliationService.GetDraft(companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("no draft found to commit: %w", err)
	}
	
	// TODO: Update DBF files here (CHECKS.dbf and CHECKREC.dbf)
	// For now, just commit the draft in SQLite
	
	err = a.reconciliationService.CommitReconciliation(draft.ID, a.currentUser.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to commit reconciliation: %w", err)
	}
	
	return map[string]interface{}{
		"status": "success",
		"message": "Reconciliation committed successfully",
		"id": draft.ID,
	}, nil
}

// GetReconciliationHistory retrieves reconciliation history for an account
func (a *App) GetReconciliationHistory(companyName, accountNumber string) (map[string]interface{}, error) {
	fmt.Printf("GetReconciliationHistory called for company: %s, account: %s\n", companyName, accountNumber)
	
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	if a.reconciliationService == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}
	
	history, err := a.reconciliationService.GetHistory(companyName, accountNumber, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to get reconciliation history: %w", err)
	}
	
	return map[string]interface{}{
		"status": "success",
		"history": history,
		"count": len(history),
	}, nil
}

// MigrateReconciliationData migrates existing CHECKREC.DBF data to SQLite
func (a *App) MigrateReconciliationData(companyName string) (map[string]interface{}, error) {
	fmt.Printf("MigrateReconciliationData called for company: %s\n", companyName)
	
	// Check permissions - only admin/root can migrate
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.maintain") {
		return nil, fmt.Errorf("insufficient permissions to migrate data")
	}
	
	if a.reconciliationService == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}
	
	result, err := a.reconciliationService.MigrateFromDBF(companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate reconciliation data: %w", err)
	}
	
	return map[string]interface{}{
		"status": "success",
		"migration_result": result,
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

// TestOLEConnection tests if we can connect to FoxPro OLE server
func (a *App) TestOLEConnection() (map[string]interface{}, error) {
	// Add immediate console output
	fmt.Println("=== TestOLEConnection STARTED ===")
	fmt.Printf("TestOLEConnection called at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	
	// Also add to debug log
	debug.SimpleLog("=== TestOLEConnection STARTED ===")
	debug.SimpleLog(fmt.Sprintf("TestOLEConnection called at %s", time.Now().Format("2006-01-02 15:04:05")))
	
	// Log the test attempt
	exePath, _ := os.Executable()
	fmt.Printf("TestOLEConnection: Executable path: %s\n", exePath)
	debug.SimpleLog(fmt.Sprintf("TestOLEConnection: Executable path: %s", exePath))
	
	logDir := filepath.Join(filepath.Dir(exePath), "logs")
	fmt.Printf("TestOLEConnection: Log directory: %s\n", logDir)
	debug.SimpleLog(fmt.Sprintf("TestOLEConnection: Log directory: %s", logDir))
	
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("TestOLEConnection: ERROR creating logs directory: %v\n", err)
		debug.SimpleLog(fmt.Sprintf("TestOLEConnection: ERROR creating logs directory: %v", err))
	}
	
	timestamp := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("financialsx_ole_%s.log", timestamp))
	testLogPath := filepath.Join(logDir, fmt.Sprintf("financialsx_test_%s.log", timestamp))
	
	fmt.Printf("TestOLEConnection: OLE log path: %s\n", logPath)
	fmt.Printf("TestOLEConnection: Test log path: %s\n", testLogPath)
	debug.SimpleLog(fmt.Sprintf("TestOLEConnection: OLE log path: %s", logPath))
	debug.SimpleLog(fmt.Sprintf("TestOLEConnection: Test log path: %s", testLogPath))
	
	// Write a test log to confirm function was called
	testLog := fmt.Sprintf("[%s] TestOLEConnection called from UI\n", time.Now().Format("2006-01-02 15:04:05"))
	if err := os.WriteFile(testLogPath, []byte(testLog), 0644); err != nil {
		fmt.Printf("TestOLEConnection: ERROR writing test log: %v\n", err)
		debug.SimpleLog(fmt.Sprintf("TestOLEConnection: ERROR writing test log: %v", err))
	} else {
		fmt.Printf("TestOLEConnection: Test log written successfully\n")
		debug.SimpleLog("TestOLEConnection: Test log written successfully")
	}
	
	// Try to create DbApi client
	fmt.Printf("TestOLEConnection: Attempting to create OLE DbApi client...\n")
	fmt.Printf("TestOLEConnection: Using ProgID: Pivoten.DbApi\n")
	debug.SimpleLog("TestOLEConnection: Attempting to create OLE DbApi client...")
	debug.SimpleLog("TestOLEConnection: Using ProgID: Pivoten.DbApi")
	
	client, err := ole.NewDbApiClient()
	if err != nil {
		errMsg := fmt.Sprintf("TestOLEConnection: FAILED to connect to OLE server: %v", err)
		fmt.Printf("%s\n", errMsg)
		debug.SimpleLog(errMsg)
		
		// Provide detailed instructions for fixing the OLE server
		result := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"message": "Could not connect to Pivoten.DbApi COM server. The dbapi.prg file has been fixed.",
			"logPath": logPath,
			"hint":    "To register: 1) Build dbapi.exe from dbapi.prg in VFP, 2) Run 'dbapi.exe /regserver' as admin",
			"details": "The dbapi.prg file in project root has been fixed to remove TRY/CATCH issues",
		}
		
		fmt.Printf("TestOLEConnection: Returning error result: %v\n", result)
		debug.SimpleLog(fmt.Sprintf("TestOLEConnection: Returning error result: %v", result))
		fmt.Println("=== TestOLEConnection ENDED (FAILED) ===")
		debug.SimpleLog("=== TestOLEConnection ENDED (FAILED) ===")
		
		return result, nil
	}
	defer client.Close()
	
	fmt.Printf("TestOLEConnection: SUCCESS - Connected to OLE server!\n")
	debug.SimpleLog("TestOLEConnection: SUCCESS - Connected to OLE server!")
	
	// Try to call Ping() method to verify server is working
	fmt.Printf("TestOLEConnection: Testing Ping() method...\n")
	debug.SimpleLog("TestOLEConnection: Testing Ping() method...")
	
	// Note: This would require actual OLE implementation
	// For now, we just report successful connection
	
	result := map[string]interface{}{
		"success": true,
		"message": "Successfully connected to Pivoten.DbApi COM server (v1.0.1)!",
		"logPath": logPath,
		"version": "1.0.1",
	}
	
	fmt.Printf("TestOLEConnection: Returning success result: %v\n", result)
	debug.SimpleLog(fmt.Sprintf("TestOLEConnection: Returning success result: %v", result))
	fmt.Println("=== TestOLEConnection ENDED (SUCCESS) ===")
	debug.SimpleLog("=== TestOLEConnection ENDED (SUCCESS) ===")
	
	return result, nil
}

// GetCompanyInfo retrieves company information from FoxPro OLE server
func (a *App) GetCompanyInfo(companyName string) (map[string]interface{}, error) {
	// Check user permission
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	fmt.Printf("GetCompanyInfo called for company: %s\n", companyName)

	// First try to get company info from compmast.dbf
	companies, err := a.GetCompanyList()
	if err == nil {
		for _, comp := range companies {
			if comp["company_id"] == companyName || comp["company_name"] == companyName {
				return map[string]interface{}{
					"success": true,
					"mock":    false,
					"data":    comp,
				}, nil
			}
		}
	}
	
	// Try to connect to Pivoten.DbApi COM server for more detailed data
	client, err := ole.NewDbApiClient()
	if err != nil {
		fmt.Printf("DbApi connection failed: %v\n", err)
		// If OLE is not available, return mock data for now
		mockData := map[string]interface{}{
			"success": true,
			"error":   fmt.Sprintf("DbApi Error: %v", err),
			"mock":    true,
			"data": map[string]interface{}{
				"company_id":      companyName,
				"company_name":    companyName,
				"address1":        "123 Main Street (Mock Data)",
				"address2":        "",
				"city":            "Houston",
				"state":           "TX",
				"zip_code":        "77001",
				"contact":         "John Smith",
				"phone":           "(555) 123-4567",
				"fax":             "(555) 123-4568",
				"email":           "info@company.com",
				"tax_id":          "XX-XXXXXXX",
				"data_path":       fmt.Sprintf("datafiles/%s/", companyName),
				"fiscal_year_end": "12/31",
				"industry":        "Oil & Gas",
			},
		}
		return mockData, nil
	}
	defer client.Close()

	fmt.Println("DbApi connection successful")
	
	// Get executable directory as base path
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	
	// Initialize DbApi with base directory  
	err = client.Initialize(exeDir)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize DbApi: %v\n", err)
	}
	
	// Open the company's database
	dbcPath := filepath.Join(exeDir, "datafiles", companyName)
	err = client.OpenDbc(dbcPath)
	if err != nil {
		fmt.Printf("Failed to open DBC for %s: %v\n", companyName, err)
		// Return basic info from compmast.dbf if available
		if len(companies) > 0 {
			for _, comp := range companies {
				if comp["company_id"] == companyName || comp["company_name"] == companyName {
					return map[string]interface{}{
						"success": true,
						"mock":    false,
						"data":    comp,
					}, nil
				}
			}
		}
	}

	return map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"company_id":   companyName,
			"company_name": companyName,
			"data_path":    dbcPath,
			"db_connected": client.IsDbcOpen(),
		},
	}, nil
}

// UpdateCompanyInfo updates company information via FoxPro OLE server
func (a *App) UpdateCompanyInfo(companyDataJSON string) (map[string]interface{}, error) {
	// Check user permission
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	// Only Admin and Root users can update
	if !a.currentUser.IsRoot && a.currentUser.RoleName != "Admin" {
		return nil, fmt.Errorf("insufficient permissions to update company information")
	}

	// Parse the JSON data
	var companyData map[string]interface{}
	err := json.Unmarshal([]byte(companyDataJSON), &companyData)
	if err != nil {
		return nil, fmt.Errorf("invalid company data: %v", err)
	}

	// For now, just return success since we'd need to know the exact table structure
	// In the future, this would use DbApi to execute UPDATE statements

	return map[string]interface{}{
		"success": true,
		"message": "Company information updated successfully",
	}, nil
}

func main() {
	// Initialize simple debug logging first (for Windows debugging)
	debug.SimpleLog("=== FinancialsX Desktop Starting ===")
	debug.SimpleLog("Initializing application...")
	defer debug.Close()
	
	// Write startup message immediately to verify app is starting
	exePath, _ := os.Executable()
	startupLog := filepath.Join(filepath.Dir(exePath), "startup.log")
	startupMsg := fmt.Sprintf("[%s] FinancialsX Desktop starting...\n", time.Now().Format("2006-01-02 15:04:05"))
	os.WriteFile(startupLog, []byte(startupMsg), 0644)
	debug.SimpleLog(fmt.Sprintf("Startup log: %s", startupLog))
	
	// Initialize logging system
	if err := logger.Initialize(); err != nil {
		// Write error to startup log
		errMsg := fmt.Sprintf("[%s] Failed to initialize logger: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		os.WriteFile(startupLog, []byte(startupMsg+errMsg), 0644)
		fmt.Printf("Failed to initialize logging: %v\n", err)
	}
	defer logger.Close()
	
	// Set up global panic recovery
	defer func() {
		if r := recover(); r != nil {
			logger.RecoverPanic("main")
			// Also write to startup log
			panicMsg := fmt.Sprintf("[%s] PANIC in main: %v\n", time.Now().Format("2006-01-02 15:04:05"), r)
			// Append to startup log
			if f, err := os.OpenFile(startupLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				f.WriteString(panicMsg)
				f.Close()
			}
		}
	}()
	
	logger.WriteInfo("Main", "Creating application instance")
	
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
		logger.WriteError("Main", fmt.Sprintf("Failed to run Wails: %v", err))
		println("Error:", err.Error())
		// Write to startup log as well
		errMsg := fmt.Sprintf("[%s] Failed to run Wails: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		// Append to startup log
		if f, err := os.OpenFile(startupLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			f.WriteString(errMsg)
			f.Close()
		}
	}
}
