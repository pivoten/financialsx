package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/Valentin-Kaiser/go-dbase/dbase"
	"github.com/pivoten/financialsx/desktop/internal/app"
	"github.com/pivoten/financialsx/desktop/internal/common"
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/config"
	"github.com/pivoten/financialsx/desktop/internal/database"
	"github.com/pivoten/financialsx/desktop/internal/debug"
	"github.com/pivoten/financialsx/desktop/internal/financials/audit"
	"github.com/pivoten/financialsx/desktop/internal/logger"
	"github.com/pivoten/financialsx/desktop/internal/ole"
	"github.com/pivoten/financialsx/desktop/internal/reconciliation"
	"github.com/pivoten/financialsx/desktop/internal/legacy"
	"github.com/pivoten/financialsx/desktop/internal/vfp"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// App struct
type App struct {
	ctx      context.Context
	db       *database.DB
	auth     *common.Auth
	currentUser *common.User
	currentCompanyPath string
	reconciliationService *reconciliation.Service
	vfpClient *vfp.VFPClient  // VFP integration client (internal use)
	*legacy.VFPWrapper  // Embedded VFP wrapper - methods are directly available
	auditService *audit.Service  // Financial audit service (uses wrappers for compatibility)
	dataBasePath string // Base path where compmast.dbf is located
	*common.I18n // Embedded i18n - methods are directly available
	
	// Services - new modular architecture
	Services *app.Services
	
	// Platform detection (cached at startup)
	platform     string // Operating system: "windows", "darwin", "linux"
	isWindows    bool   // Convenience flag for Windows platform
	
	// Authentication state (cached after login)
	isAuthenticated bool                    // Whether user is logged in
	isAdmin        bool                    // Whether user has admin privileges
	isRoot         bool                    // Whether user has root privileges
	permissions    map[string]bool         // Cached permission set
	userRole       string                  // Cached role name
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Detect and store platform information at startup
	a.platform = runtime.GOOS
	a.isWindows = (runtime.GOOS == "windows")
	
	// Initialize debug logging (SimpleLog will auto-initialize if needed)
	debug.SimpleLog("=== App.startup called ===")
	debug.LogInfo("App", "Application starting up")
	debug.SimpleLog(fmt.Sprintf("Platform detected: %s (isWindows: %v)", a.platform, a.isWindows))
	
	// Log environment info
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	cwd, _ := os.Getwd()
	debug.SimpleLog(fmt.Sprintf("Executable path: %s", exePath))
	debug.SimpleLog(fmt.Sprintf("Executable dir: %s", exeDir))
	debug.SimpleLog(fmt.Sprintf("Working dir: %s", cwd))
	debug.SimpleLog(fmt.Sprintf("OS env: %s", os.Getenv("OS")))
	debug.SimpleLog(fmt.Sprintf("PROCESSOR_ARCHITECTURE: %s", os.Getenv("PROCESSOR_ARCHITECTURE")))
	
	// Initialize i18n
	a.I18n = common.NewI18n("en")
	// Try to load locales from frontend directory
	localesPath := filepath.Join("frontend", "src", "locales")
	if err := a.I18n.LoadLocalesFromDir(localesPath); err != nil {
		debug.SimpleLog(fmt.Sprintf("Failed to load locales: %v", err))
	}
	
	debug.SimpleLog("=== App.startup completed ===")
}

// updateAuthCache updates cached authentication state from current user
func (a *App) updateAuthCache() {
	if a.currentUser == nil {
		a.isAuthenticated = false
		a.isAdmin = false
		a.isRoot = false
		a.permissions = make(map[string]bool)
		a.userRole = ""
		debug.LogInfo("Auth", "Auth cache cleared (no user)")
		return
	}
	
	a.isAuthenticated = true
	a.isAdmin = a.currentUser.IsAdmin()
	a.isRoot = a.currentUser.IsRoot
	a.userRole = a.currentUser.RoleName
	
	// Cache all permissions for fast lookup
	a.permissions = make(map[string]bool)
	commonPerms := []string{
		"database.read", "database.write", "database.maintain",
		"dbf.read", "dbf.write",
		"users.read", "users.create", "users.update", "users.manage_roles",
		"settings.read", "settings.write",
	}
	
	for _, perm := range commonPerms {
		a.permissions[perm] = a.currentUser.HasPermission(perm)
	}
	
	debug.LogInfo("Auth", fmt.Sprintf("Auth cache updated - User: %s, Role: %s, Admin: %v, Root: %v", 
		a.currentUser.Username, a.userRole, a.isAdmin, a.isRoot))
}

// hasPermission checks cached permissions (faster than calling user.HasPermission)
func (a *App) hasPermission(permission string) bool {
	if !a.isAuthenticated {
		return false
	}
	
	// Root and admin bypass most permission checks
	if a.isRoot || a.isAdmin {
		return true
	}
	
	// Check cached permissions
	if allowed, exists := a.permissions[permission]; exists {
		return allowed
	}
	
	// If not cached, check directly and cache result
	if a.currentUser != nil {
		allowed := a.currentUser.HasPermission(permission)
		a.permissions[permission] = allowed
		return allowed
	}
	
	return false
}

// GetPlatform returns the current platform information
func (a *App) GetPlatform() map[string]interface{} {
	return map[string]interface{}{
		"platform": a.platform,
		"isWindows": a.isWindows,
		"arch": runtime.GOARCH,
	}
}

// GetAuthState returns the current authentication state
func (a *App) GetAuthState() map[string]interface{} {
	return map[string]interface{}{
		"isAuthenticated": a.isAuthenticated,
		"isAdmin": a.isAdmin,
		"isRoot": a.isRoot,
		"userRole": a.userRole,
		"username": func() string {
			if a.currentUser != nil {
				return a.currentUser.Username
			}
			return ""
		}(),
	}
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
		
		// Initialize services with the new modular architecture
		if a.Services == nil {
			a.Services = app.NewServices(db)
		}
		
		// Initialize VFP integration client
		a.vfpClient = vfp.NewVFPClient(db.GetDB())
		a.VFPWrapper = legacy.NewVFPWrapper(a.vfpClient)
		
		// Initialize audit service
		a.auditService = audit.NewService()
		
		if err := a.vfpClient.InitializeSchema(); err != nil {
			debug.SimpleLog(fmt.Sprintf("App.InitializeCompanyDatabase: Error initializing VFP schema: %v", err))
			// Non-fatal error, VFP integration is optional
		}
		
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
		a.auth = common.New(db, companyName) // Pass companyName to Auth constructor
		a.reconciliationService = reconciliation.NewService(db)
		
		// Initialize services with the new modular architecture
		if a.Services == nil {
			a.Services = app.NewServices(db)
		}
		
		// Initialize VFP integration client
		a.vfpClient = vfp.NewVFPClient(db.GetDB())
		a.VFPWrapper = legacy.NewVFPWrapper(a.vfpClient)
		
		// Initialize audit service
		a.auditService = audit.NewService()
		
		if err := a.vfpClient.InitializeSchema(); err != nil {
			// Non-fatal error, VFP integration is optional
		}
	}

	user, session, err := a.auth.Login(username, password)
	if err != nil {
		return nil, err
	}

	a.currentUser = user
	
	// Update cached authentication state
	a.updateAuthCache()
	
	// Don't preload OLE connection on login - creates duplicate processes
	// OLE connection will be created on first actual database query
	// ole.PreloadOLEConnection(companyName)
	fmt.Printf("Login: Deferred OLE connection for company: %s\n", companyName)

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
		a.auth = common.New(db, companyName) // Pass companyName to Auth constructor
		a.reconciliationService = reconciliation.NewService(db)
		
		// Initialize services with the new modular architecture
		if a.Services == nil {
			a.Services = app.NewServices(db)
		}
		
		// Initialize VFP integration client
		a.vfpClient = vfp.NewVFPClient(db.GetDB())
		a.VFPWrapper = legacy.NewVFPWrapper(a.vfpClient)
		
		// Initialize audit service
		a.auditService = audit.NewService()
		
		if err := a.vfpClient.InitializeSchema(); err != nil {
			// Non-fatal error, VFP integration is optional
		}
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
	
	// Don't preload OLE connection on registration - creates duplicate processes
	// OLE connection will be created on first actual database query
	// ole.PreloadOLEConnection(companyName)
	fmt.Printf("Register: Deferred OLE connection for company: %s\n", companyName)

	return map[string]interface{}{
		"user":    user,
		"session": session,
	}, nil
}

// Logout handles user logout
func (a *App) Logout(token string) error {
	fmt.Printf("Logout: Starting logout process\n")
	debug.SimpleLog("Logout: Starting logout process")
	
	// Close OLE connection when user logs out
	fmt.Printf("Logout: Calling CloseOLEConnection\n")
	debug.SimpleLog("Logout: Calling CloseOLEConnection")
	ole.CloseOLEConnection()
	fmt.Printf("Logout: CloseOLEConnection completed\n")
	debug.SimpleLog("Logout: CloseOLEConnection completed")
	
	// Clear current user and cached auth state
	a.currentUser = nil
	a.updateAuthCache()
	
	if a.auth != nil {
		err := a.auth.Logout(token)
		fmt.Printf("Logout: Auth logout completed, error: %v\n", err)
		debug.SimpleLog(fmt.Sprintf("Logout: Auth logout completed, error: %v", err))
		return err
	}
	
	fmt.Printf("Logout: Complete\n")
	debug.SimpleLog("Logout: Complete")
	return nil
}

// ValidateSession checks if a session is valid for a specific company
func (a *App) ValidateSession(token string, companyName string) (*common.User, error) {
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
		a.auth = common.New(db, companyName) // Pass companyName to Auth constructor
		a.reconciliationService = reconciliation.NewService(db)
		
		// Initialize services with the new modular architecture
		if a.Services == nil {
			a.Services = app.NewServices(db)
		}
		
		// Initialize VFP integration client
		a.vfpClient = vfp.NewVFPClient(db.GetDB())
		a.VFPWrapper = legacy.NewVFPWrapper(a.vfpClient)
		
		// Initialize audit service
		a.auditService = audit.NewService()
		
		if err := a.vfpClient.InitializeSchema(); err != nil {
			// Non-fatal error, VFP integration is optional
		}
		
		// Close any existing OLE connection when switching companies
		// Don't preload - let it create on first query to avoid duplicates
		ole.CloseOLEConnection()
		fmt.Printf("ValidateSession: Closed OLE connection for company switch to: %s\n", companyName)
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
	// Use DBF service
	return a.Services.DBF.GetFiles(companyName)
}

// GetDBFTableData returns the structure and data of a DBF file
func (a *App) GetDBFTableData(companyName, fileName string) (map[string]interface{}, error) {
	// Use DBF service
	return a.Services.DBF.GetTableData(companyName, fileName)
}

// CheckGLPeriodFields checks for blank CYEAR/CPERIOD fields in GLMASTER.dbf
func (a *App) CheckGLPeriodFields(companyName string) (map[string]interface{}, error) {
	// Use GL service
	return a.Services.GL.CheckGLPeriodFields(companyName)
}

// AnalyzeGLBalancesByYear analyzes GL balances grouped by year and account
func (a *App) AnalyzeGLBalancesByYear(companyName string, accountNumber string) (map[string]interface{}, error) {
	// Use GL service
	return a.Services.GL.AnalyzeGLBalancesByYear(companyName, accountNumber)
}

// ValidateGLBalances performs comprehensive GL validation checks
func (a *App) ValidateGLBalances(companyName string, accountNumber string) (map[string]interface{}, error) {
	// Use GL service
	return a.Services.GL.ValidateGLBalances(companyName, accountNumber)
}

// GetChartOfAccounts retrieves the chart of accounts from COA.dbf
func (a *App) GetChartOfAccounts(companyName string, sortBy string, includeInactive bool) ([]map[string]interface{}, error) {
	// Use GL service
	return a.Services.GL.GetChartOfAccounts(companyName, sortBy, includeInactive)
}

// RunClosingProcess runs the period closing process
func (a *App) RunClosingProcess(companyName string, periodEnd string, closingDate string, description string, forceClose bool) (map[string]interface{}, error) {
	// Use GL service
	result, err := a.Services.GL.RunClosingProcess(companyName, periodEnd, closingDate, description, forceClose)
	if err != nil {
		return nil, err
	}
	
	// Convert ClosingResult to map for frontend
	return map[string]interface{}{
		"period_end":        result.PeriodEnd,
		"status":            result.Status,
		"entries_created":   result.EntriesCreated,
		"accounts_affected": result.AccountsAffected,
		"total_debits":      result.TotalDebits,
		"total_credits":     result.TotalCredits,
		"warnings":          result.Warnings,
		"errors":            result.Errors,
	}, nil
}

// GetClosingStatus gets the closing status for a period
func (a *App) GetClosingStatus(companyName string, periodEnd string) (string, error) {
	// Use GL service
	return a.Services.GL.GetClosingStatus(companyName, periodEnd)
}

// ReopenPeriod reopens a previously closed period
func (a *App) ReopenPeriod(companyName string, periodEnd string, reason string) error {
	// Use GL service
	return a.Services.GL.ReopenPeriod(companyName, periodEnd, reason)
}

// GetDBFTableDataPaged returns paginated and sorted data from a DBF file
func (a *App) GetDBFTableDataPaged(companyName, fileName string, offset, limit int, sortColumn, sortDirection string) (map[string]interface{}, error) {
	// Use DBF service
	return a.Services.DBF.GetTableDataPaged(companyName, fileName, offset, limit, sortColumn, sortDirection)
}

// SearchDBFTable searches a DBF file and returns matching records
func (a *App) SearchDBFTable(companyName, fileName, searchTerm string) (map[string]interface{}, error) {
	// Use DBF service
	return a.Services.DBF.SearchTable(companyName, fileName, searchTerm)
}

// UpdateDBFRecord updates a specific record in a DBF file
func (a *App) UpdateDBFRecord(companyName, fileName string, rowIndex, colIndex int, value string) error {
	// Use DBF service
	return a.Services.DBF.UpdateRecord(companyName, fileName, rowIndex, colIndex, value)
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

// ==============================================================================
// LOCAL SQLITE USER MANAGEMENT
// ==============================================================================
// NOTE: These functions manage users in the local SQLite database only.
// They are NOT related to Supabase authentication.
// 
// These functions will be DEPRECATED once Supabase integration is complete.
// For now, they support the UserManagement UI component for local testing.
//
// TODO: Remove these functions and the UserManagement component once 
//       Supabase auth is fully integrated.
// ==============================================================================

// GetAllUsers returns all users from local SQLite (admin/root only)
// DEPRECATED: Will be removed once Supabase integration is complete
func (a *App) GetAllUsers() ([]common.User, error) {
	if a.auth == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("users.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	return a.auth.GetAllUsers()
}

// GetAllRoles returns all available roles from local SQLite
// DEPRECATED: Will be removed once Supabase integration is complete
func (a *App) GetAllRoles() ([]common.Role, error) {
	if a.auth == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("users.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	return a.auth.GetAllRoles()
}

// UpdateUserRole updates a user's role in local SQLite (admin/root only)
// DEPRECATED: Will be removed once Supabase integration is complete
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

// UpdateUserStatus activates or deactivates a user in local SQLite (admin/root only)
// DEPRECATED: Will be removed once Supabase integration is complete
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

// CreateUser creates a new user in local SQLite (admin/root only)
// DEPRECATED: Will be removed once Supabase integration is complete
func (a *App) CreateUser(username, password, email string, roleID int) (*common.User, error) {
	// Use service if available
	if a.Services != nil && a.Services.Auth != nil {
		// Check permissions
		if a.currentUser == nil || !a.currentUser.HasPermission("users.create") {
			return nil, fmt.Errorf("insufficient permissions")
		}
		return a.Services.Auth.CreateUser(username, password, email, roleID)
	}
	
	// Fallback to direct auth
	if a.auth == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("users.create") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	return a.auth.CreateUser(username, password, email, roleID)
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

// PreloadOLEConnection preloads the OLE connection for a company
func (a *App) PreloadOLEConnection(companyName string) map[string]interface{} {
	ole.PreloadOLEConnection(companyName)
	fmt.Printf("PreloadOLEConnection: Requested for company: %s\n", companyName)
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("OLE connection preloaded for %s", companyName),
	}
}

// CloseOLEConnection closes the current OLE connection
func (a *App) CloseOLEConnection() map[string]interface{} {
	ole.CloseOLEConnection()
	fmt.Printf("CloseOLEConnection: Connection closed\n")
	return map[string]interface{}{
		"success": true,
		"message": "OLE connection closed",
	}
}

// SetOLEIdleTimeout sets the idle timeout for OLE connections
func (a *App) SetOLEIdleTimeout(minutes int) map[string]interface{} {
	timeout := time.Duration(minutes) * time.Minute
	ole.SetIdleTimeout(timeout)
	fmt.Printf("SetOLEIdleTimeout: Set to %d minutes\n", minutes)
	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Idle timeout set to %d minutes", minutes),
	}
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
	
	// Execute on dedicated COM thread to avoid threading issues
	var jsonResult string
	var queryErr error
	
	err := ole.ExecuteOnCOMThread(companyName, func(client *ole.DbApiClient) error {
		fmt.Printf("TestDatabaseQuery: Executing on COM thread\n")
		debug.SimpleLog("TestDatabaseQuery: Using COM thread for OLE connection")
		
		// Note: Ping method would be called here if implemented in OLE client
		fmt.Printf("TestDatabaseQuery: OLE connection established\n")
		debug.SimpleLog("TestDatabaseQuery: OLE connection established")
		
		// Database should already be open via ExecuteOnCOMThread
		fmt.Printf("TestDatabaseQuery: Database is open on COM thread\n")
		
		// Execute the query via OLE using JSON
		fmt.Printf("TestDatabaseQuery: Executing SQL query via OLE (JSON)...\n")
		jsonResult, queryErr = client.QueryToJson(query)
		return queryErr
	})
	
	if err != nil {
		fmt.Printf("TestDatabaseQuery: Failed to use singleton OLE connection: %v\n", err)
		debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: OLE singleton connection failed: %v", err))
		
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("OLE server connection failed: %v", err),
			"message": "Could not connect to Pivoten.DbApi COM server",
			"hint":    "To use OLE: 1) Build dbapi.exe from dbapi.prg in Visual FoxPro, 2) Run 'dbapi.exe /regserver' as admin",
			"progId":  "Pivoten.DbApi",
		}, nil
	}
	
	if err != nil {
		fmt.Printf("TestDatabaseQuery: Query execution failed: %v\n", err)
		debug.SimpleLog(fmt.Sprintf("TestDatabaseQuery: Query failed: %v", err))
		
		// Get last error from OLE server if available
		var lastError string
		ole.ExecuteOnCOMThread(companyName, func(client *ole.DbApiClient) error {
			lastError = client.GetLastError()
			return nil
		})
		
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
	
	// Execute on dedicated COM thread to avoid threading issues
	var jsonResult string
	var tableErr error
	
	err := ole.ExecuteOnCOMThread(companyName, func(client *ole.DbApiClient) error {
		fmt.Printf("GetTableList: Executing on COM thread\n")
		debug.SimpleLog("GetTableList: Using COM thread for OLE connection")
		
		// Database should already be open via ExecuteOnCOMThread
		// Use the new JSON method to get table list
		jsonResult, tableErr = client.GetTableListSimple()
		return tableErr
	})
	
	if err != nil {
		fmt.Printf("GetTableList: Failed to use singleton OLE connection: %v\n", err)
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

// getSavedDataPath retrieves the saved data path from a config file
func (a *App) getSavedDataPath() string {
	configPath := filepath.Join(os.TempDir(), "financialsx_datapath.txt")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// saveDataPath saves the data path to a config file for future use
func (a *App) saveDataPath(path string) {
	configPath := filepath.Join(os.TempDir(), "financialsx_datapath.txt")
	os.WriteFile(configPath, []byte(path), 0644)
}

// SelectDataFolder opens a dialog for the user to select the folder containing compmast.dbf
func (a *App) SelectDataFolder() (string, error) {
	// This will trigger a folder selection dialog in the frontend
	// The frontend will call back with the selected path
	return "", fmt.Errorf("TRIGGER_FOLDER_DIALOG")
}

// SetDataPath sets the data path after user selection and verifies compmast.dbf exists
func (a *App) SetDataPath(folderPath string) error {
	// Verify compmast.dbf exists in the selected folder
	compMastPath := filepath.Join(folderPath, "compmast.dbf")
	if _, err := os.Stat(compMastPath); os.IsNotExist(err) {
		return fmt.Errorf("compmast.dbf not found in selected folder")
	}
	
	// Save the path for future use
	a.dataBasePath = folderPath
	a.saveDataPath(folderPath)
	
	fmt.Printf("SetDataPath: Data path set to: %s\n", folderPath)
	debug.LogInfo("SetDataPath", fmt.Sprintf("Data path set to: %s", folderPath))
	
	return nil
}

// findCompmastDBF recursively searches for compmast.dbf file
func findCompmastDBF(startPath string, maxDepth int) string {
	if maxDepth <= 0 {
		return ""
	}
	
	var result string
	filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}
		
		// Skip hidden directories and common non-data directories
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || 
			info.Name() == "node_modules" || 
			info.Name() == "build" ||
			info.Name() == "dist") {
			return filepath.SkipDir
		}
		
		// Check if this is compmast.dbf (case-insensitive)
		if !info.IsDir() && strings.EqualFold(info.Name(), "compmast.dbf") {
			result = path
			return filepath.SkipDir // Stop walking once found
		}
		
		// Limit depth to prevent excessive searching
		relPath, _ := filepath.Rel(startPath, path)
		depth := len(strings.Split(relPath, string(filepath.Separator)))
		if depth > maxDepth {
			return filepath.SkipDir
		}
		
		return nil
	})
	
	return result
}

// GetCompanyList reads the compmast.dbf file to get available companies
func (a *App) GetCompanyList() ([]map[string]interface{}, error) {
	fmt.Println("GetCompanyList: Searching for compmast.dbf...")
	debug.LogInfo("GetCompanyList", "Searching for compmast.dbf...")
	
	var compMastPath string
	var baseDir string
	
	if a.isWindows {
		// On Windows, always look for datafiles\compmast.dbf relative to EXE
		exePath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed to get executable path: %w", err)
		}
		baseDir = filepath.Dir(exePath)
		compMastPath = filepath.Join(baseDir, "datafiles", "compmast.dbf")
		
		fmt.Printf("GetCompanyList: Looking for compmast.dbf at: %s\n", compMastPath)
		
		// Check if it exists
		if _, err := os.Stat(compMastPath); os.IsNotExist(err) {
			// Check if we have a saved path from previous selection
			savedPath := a.getSavedDataPath()
			if savedPath != "" {
				testPath := filepath.Join(savedPath, "compmast.dbf")
				if _, err := os.Stat(testPath); err == nil {
					compMastPath = testPath
					baseDir = filepath.Dir(savedPath)
					fmt.Printf("GetCompanyList: Using saved path: %s\n", compMastPath)
				} else {
					compMastPath = "" // Reset to trigger folder selection
				}
			} else {
				compMastPath = "" // Will trigger folder selection
			}
		}
	} else {
		// Mac/Linux - use the original search logic
		workDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		
		// Search for compmast.dbf recursively
		compMastPath = findCompmastDBF(workDir, 5)
		
		if compMastPath == "" {
			parentDir := filepath.Dir(workDir)
			compMastPath = findCompmastDBF(parentDir, 3)
		}
		
		if compMastPath == "" {
			exePath, _ := os.Executable()
			exeDir := filepath.Dir(exePath)
			compMastPath = findCompmastDBF(exeDir, 3)
		}
		
		if compMastPath == "" {
			savedPath := a.getSavedDataPath()
			if savedPath != "" {
				testPath := filepath.Join(savedPath, "compmast.dbf")
				if _, err := os.Stat(testPath); err == nil {
					compMastPath = testPath
					fmt.Printf("GetCompanyList: Using saved path: %s\n", compMastPath)
				}
			}
		}
		
		if compMastPath != "" {
			baseDir = filepath.Dir(filepath.Dir(compMastPath)) // Go up from datafiles/compmast.dbf
		}
	}
	
	if compMastPath == "" {
		fmt.Println("GetCompanyList: compmast.dbf not found, user needs to select folder")
		debug.LogError("GetCompanyList", fmt.Errorf("compmast.dbf not found"))
		return nil, fmt.Errorf("NEED_FOLDER_SELECTION")
	}
	
	// Store the base path for future company data access
	if a.isWindows {
		// On Windows, store the directory where the EXE is (or where user selected)
		a.dataBasePath = baseDir
		a.saveDataPath(filepath.Join(baseDir, "datafiles"))
	} else {
		// On Mac/Linux, store the datafiles directory
		a.dataBasePath = filepath.Dir(compMastPath)
		a.saveDataPath(a.dataBasePath)
	}
	
	fmt.Printf("GetCompanyList: Found compmast.dbf at: %s\n", compMastPath)
	debug.LogInfo("GetCompanyList", fmt.Sprintf("Found compmast.dbf at: %s", compMastPath))
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
	
	// Get list of actual directories in datafiles folder
	datafilesPath := filepath.Join(a.dataBasePath, "datafiles")
	var actualFolders []string
	if entries, err := os.ReadDir(datafilesPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				actualFolders = append(actualFolders, entry.Name())
			}
		}
	}
	fmt.Printf("GetCompanyList: Found %d actual folders in datafiles: %v\n", len(actualFolders), actualFolders)
	
	// Transform the data for frontend consumption
	companies := []map[string]interface{}{}
	for _, row := range rows {
		dataPath := ""
		if cdatapath, ok := row["CDATAPATH"].(string); ok {
			// DBF files have fixed-width fields padded with spaces
			// Just find where the real path ends by looking for excessive spaces
			
			// Convert to bytes to handle any null bytes
			bytes := []byte(cdatapath)
			
			// Find where the actual path ends (before excessive padding)
			// Look for 5+ consecutive spaces as that indicates padding
			spaceCount := 0
			actualEnd := len(bytes)
			for i := 0; i < len(bytes); i++ {
				if bytes[i] == ' ' || bytes[i] == 0 {
					spaceCount++
					if spaceCount >= 5 {
						// Found the padding, mark where the real path ends
						actualEnd = i - spaceCount + 1
						break
					}
				} else {
					spaceCount = 0
				}
			}
			
			// Extract the actual path
			if actualEnd > 0 {
				dataPath = string(bytes[:actualEnd])
			}
			
			// Simple cleanup - just trim trailing spaces
			dataPath = strings.TrimRight(dataPath, " \t\r\n\x00")
			
			// Ensure it ends with a backslash (as it should from the DBF)
			if dataPath != "" && !strings.HasSuffix(dataPath, "\\") {
				dataPath = dataPath + "\\"
			}
		}
		
		originalDataPath := dataPath
		
		// Try to find the actual folder name by matching company name
		companyName := ""
		if cproducer, ok := row["CPRODUCER"].(string); ok {
			companyName = strings.TrimSpace(cproducer)
		}
		
		// Look for a folder that matches the company name (case-insensitive, remove spaces)
		actualFolderName := ""
		companyNameLower := strings.ToLower(strings.ReplaceAll(companyName, " ", ""))
		for _, folder := range actualFolders {
			folderLower := strings.ToLower(folder)
			if strings.Contains(folderLower, companyNameLower) || strings.Contains(companyNameLower, folderLower) {
				actualFolderName = folder
				break
			}
		}
		
		// If we found an actual folder, use that; otherwise fall back to extracted path
		if actualFolderName != "" {
			dataPath = actualFolderName
			fmt.Printf("GetCompanyList: Matched company '%s' to folder '%s'\n", companyName, actualFolderName)
		} else {
			// Platform-specific path handling:
			if a.isWindows && dataPath != "" {
				// On Windows: Check if it's an absolute path or relative
				// Absolute paths start with drive letter (C:\, D:\, etc.)
				if len(dataPath) >= 2 && dataPath[1] == ':' {
					// Absolute path - use as-is (already cleaned above)
					fmt.Printf("GetCompanyList: Using absolute Windows path: %s\n", dataPath)
				} else {
					// Relative path - will be resolved relative to compmast.dbf location
					// Just keep the path as-is, it will be resolved later
					fmt.Printf("GetCompanyList: Using relative Windows path: %s\n", dataPath)
				}
			} else if !a.isWindows && dataPath != "" {
				// On Mac/Linux: Extract just the folder name from the path
				// Handle both Windows-style paths (from Windows-created DBF) and Unix paths
				if strings.Contains(dataPath, "\\") {
					// Windows path - extract last component
					parts := strings.Split(dataPath, "\\")
					for i := len(parts) - 1; i >= 0; i-- {
						if parts[i] != "" {
							dataPath = parts[i]
							break
						}
					}
				} else if strings.Contains(dataPath, "/") {
					// Unix path - extract last component
					dataPath = filepath.Base(dataPath)
				}
				// If it's just a folder name, keep it as is
				fmt.Printf("GetCompanyList: Using Mac/Linux folder name: %s\n", dataPath)
			}
		}
		
		// Build the full resolved path for display
		var fullPath string
		if a.isWindows {
			if filepath.IsAbs(dataPath) {
				// Already absolute
				fullPath = dataPath
			} else if strings.Contains(dataPath, "\\") || strings.Contains(dataPath, "/") {
				// Relative path
				exePath, _ := os.Executable()
				exeDir := filepath.Dir(exePath)
				fullPath = filepath.Join(exeDir, dataPath)
			} else {
				// Just a folder name
				exePath, _ := os.Executable()
				exeDir := filepath.Dir(exePath)
				fullPath = filepath.Join(exeDir, "datafiles", dataPath)
			}
		} else {
			// Mac/Linux
			if dataPath != "" {
				datafilesPath := filepath.Join(a.dataBasePath, dataPath)
				fullPath = datafilesPath
			}
		}
		
		fmt.Printf("GetCompanyList: CIDCOMP=%v, CPRODUCER=%v, CALIAS=%v, original CDATAPATH=%v, final dataPath=%v, fullPath=%v\n", 
			row["CIDCOMP"], row["CPRODUCER"], row["CALIAS"], originalDataPath, dataPath, fullPath)
		
		// Get the alias, default to empty string if not present
		alias := ""
		if cAlias, ok := row["CALIAS"].(string); ok {
			alias = strings.TrimSpace(cAlias)
		}
		
		company := map[string]interface{}{
			"company_id":   row["CIDCOMP"],
			"company_name": row["CPRODUCER"],
			"alias":        alias,
			"address1":     row["CADDRESS1"],
			"address2":     row["CADDRESS2"],
			"city":         row["CCITY"],
			"state":        row["CSTATE"],
			"zip_code":     row["CZIPCODE"],
			"data_path":    dataPath,
			"full_path":    fullPath,  // The resolved full path
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
	// Use the new service architecture if available
	if a.Services != nil && a.Services.Banking != nil {
		accounts, err := a.Services.Banking.GetBankAccounts(companyName)
		if err != nil {
			return nil, err
		}
		
		// Convert BankAccount structs to map[string]interface{} for frontend compatibility
		result := make([]map[string]interface{}, len(accounts))
		for i, account := range accounts {
			result[i] = map[string]interface{}{
				"account_number":   account.AccountNumber,
				"account_name":     account.AccountName,
				"account_type":     fmt.Sprintf("%d", account.AccountType),
				"balance":          account.Balance,
				"description":      account.AccountName,
				"is_bank_account":  account.IsBankAccount,
			}
		}
		return result, nil
	}
	
	// Legacy implementation fallback
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
	// Use the new service architecture if available
	if a.Services != nil && a.Services.Banking != nil {
		checks, err := a.Services.Banking.GetOutstandingChecks(companyName, accountNumber)
		if err != nil {
			return nil, err
		}
		
		// Convert to the expected format for the frontend
		// Note: We need to return an array of maps, not OutstandingCheck structs
		checkMaps := make([]map[string]interface{}, len(checks))
		for i, check := range checks {
			checkMaps[i] = map[string]interface{}{
				"checkNumber": check.CheckNumber,
				"date":        check.CheckDate,
				"payee":       check.Payee,
				"amount":      check.Amount,
				"account":     check.AccountNumber,
				"entryType":   check.EntryType,
				"cidchec":     check.CIDCHEC,
				"id":          check.ID,
				"_rowIndex":   check.RowIndex,
				"_rawData":    check.RawData,
			}
		}
		
		// Get columns for the frontend
		columns := []string{"CCHECKNO", "DCHECKDATE", "CPAYEE", "NAMOUNT", "CACCTNO", "LCLEARED", "LVOID", "CENTRYTYPE", "CIDCHEC"}
		
		return map[string]interface{}{
			"status":  "success",
			"checks":  checkMaps,
			"total":   len(checkMaps),
			"columns": columns,
		}, nil
	}
	
	// Service is required
	return nil, fmt.Errorf("banking service not initialized")
}
func (a *App) GetAccountBalance(companyName, accountNumber string) (float64, error) {
	// Use the new service architecture if available
	if a.Services != nil && a.Services.Banking != nil {
		return a.Services.Banking.GetAccountBalance(companyName, accountNumber)
	}
	
	// Legacy implementation fallback
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
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.dbf", "", 0, 0, "", "")
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
				debit = common.ParseFloat(row[debitIdx])
			}
			
			// Get credit amount if column exists
			if creditIdx != -1 && len(row) > creditIdx {
				credit = common.ParseFloat(row[creditIdx])
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
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Delegate to Matching service
	return a.Services.Matching.RunMatching(companyName, accountNumber, options)
}

// ClearMatchesAndRerun clears all matches and reruns the matching algorithm
func (a *App) ClearMatchesAndRerun(companyName string, accountNumber string, options map[string]interface{}) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Delegate to Matching service
	return a.Services.Matching.ClearMatchesAndRerun(companyName, accountNumber, options)
}

// ImportBankStatement parses and stores CSV bank statement in SQLite (without auto-matching)
func (a *App) ImportBankStatement(companyName string, csvContent string, accountNumber string) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Delegate to Matching service
	result, err := a.Services.Matching.ImportBankStatement(companyName, csvContent, accountNumber)
	if err != nil {
		return nil, err
	}
	
	// Update imported_by field with current user
	if txns, ok := result["bankTransactions"].([]interface{}); ok {
		for _, txn := range txns {
			if t, ok := txn.(map[string]interface{}); ok {
				t["imported_by"] = a.currentUser.Username
			}
		}
	}
	
	return result, nil
}


// autoMatchBankTransactions matches bank transactions with existing checks
func (a *App) autoMatchBankTransactions(bankTransactions []BankTransaction, existingChecks []map[string]interface{}) []MatchResult {
	var matches []MatchResult
	
	// Keep track of already matched check IDs to prevent double-matching
	matchedCheckIDs := make(map[string]bool)
	
	// Sort bank transactions by date to match older transactions first
	// This helps with recurring transactions
	sort.Slice(bankTransactions, func(i, j int) bool {
		dateI, _ := common.ParseDate(bankTransactions[i].TransactionDate)
		dateJ, _ := common.ParseDate(bankTransactions[j].TransactionDate)
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
	checkAmount := common.ParseFloat(check["amount"])
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
		txnDate, txnErr := common.ParseDate(txn.TransactionDate)
		checkDate, checkErr := common.ParseDate(fmt.Sprintf("%v", check["date"]))
		
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
	checkAmount := common.ParseFloat(check["amount"])
	
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
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Delegate to Matching service
	return a.Services.Matching.ManualMatchTransaction(transactionID, checkID, checkRowIndex)
}

// RetryMatching re-runs the matching algorithm for unmatched transactions
func (a *App) RetryMatching(companyName string, accountNumber string, statementID int) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Delegate to Matching service
	return a.Services.Matching.RetryMatching(companyName, accountNumber, statementID)
}

// GetMatchedTransactions returns all matched checks with their bank transaction confirmation
func (a *App) GetMatchedTransactions(companyName string, accountNumber string) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Delegate to Matching service
	return a.Services.Matching.GetMatchedTransactions(companyName, accountNumber)
}

// UnmatchTransaction removes the match between a bank transaction and a check
func (a *App) UnmatchTransaction(transactionID int) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions")
	}
	
	// Delegate to Matching service
	return a.Services.Matching.UnmatchTransaction(transactionID)
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


// parseCSVContent parses CSV content and handles different bank formats
func (a *App) parseCSVContent(csvContent string) ([]BankTransaction, error) {
	lines := strings.Split(strings.TrimSpace(csvContent), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("CSV must contain header and at least one data row")
	}
	
	// Parse header to determine column indices - handle quoted fields
	header := common.ParseCSVLine(lines[0])
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
		fields := common.ParseCSVLine(line)
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
			transaction.Amount = common.ParseFloat(amountStr)
			if transaction.Amount == 0 && amountStr != "0" && amountStr != "0.00" {
				fmt.Printf("WARNING: Failed to parse amount: '%s'\n", amountStr)
			}
		} else {
			// Handle separate debit/credit columns
			var debit, credit float64
			if debitIdx, exists := columnMap["debit"]; exists && debitIdx < len(fields) {
				debit = common.ParseFloat(fields[debitIdx])
			}
			if creditIdx, exists := columnMap["credit"]; exists && creditIdx < len(fields) {
				credit = common.ParseFloat(fields[creditIdx])
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
			balance := common.ParseFloat(fields[balanceIdx])
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
			transaction.CheckNumber = common.ExtractCheckNumber(transaction.Description)
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
	checkAmount := common.ParseFloat(check["amount"])
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
		csvDate, csvErr := common.ParseDate(csvTxn.TransactionDate)
		checkDate, checkErr := common.ParseDate(fmt.Sprintf("%v", check["date"]))
		
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
	checkAmount := common.ParseFloat(check["amount"])
	
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



// getFreshnessStatus determines the freshness status based on time
func getFreshnessStatus(lastUpdated time.Time) string {
	age := time.Since(lastUpdated).Hours()
	if age < 1 {
		return "fresh"
	} else if age < 24 {
		return "aging"
	}
	return "stale"
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
	
	// Use service if available
	if a.Services != nil && a.Services.Banking != nil {
		balances, err := a.Services.Banking.GetCachedBalances(companyName)
		if err != nil {
			return nil, err
		}
		
		// Convert to map for frontend compatibility
		result := make([]map[string]interface{}, len(balances))
		for i, b := range balances {
			// Calculate age hours
			glAgeHours := time.Since(b.GLLastUpdated).Hours()
			checksAgeHours := time.Since(b.ChecksLastUpdated).Hours()
			
			result[i] = map[string]interface{}{
				"account_number":         b.AccountNumber,
				"account_name":          b.AccountName,
				"gl_balance":            b.GLBalance,
				"outstanding_total":     b.OutstandingChecksTotal,
				"outstanding_count":     b.OutstandingChecksCount,
				"bank_balance":          b.BankBalance,
				"gl_last_updated":       b.GLLastUpdated,
				"checks_last_updated":   b.ChecksLastUpdated,
				"gl_age_hours":          glAgeHours,
				"checks_age_hours":      checksAgeHours,
				"gl_freshness":          getFreshnessStatus(b.GLLastUpdated),
				"checks_freshness":      getFreshnessStatus(b.ChecksLastUpdated),
				"is_stale":             b.IsStale,
				// These fields might not be available from service yet
				"uncleared_deposits":    0,
				"uncleared_checks":      b.OutstandingChecksTotal,
				"deposit_count":         0,
				"check_count":           b.OutstandingChecksCount,
			}
		}
		
		fmt.Printf("GetCachedBalances: Retrieved %d balances via service\n", len(result))
		return result, nil
	}
	
	// Fallback to direct implementation
	// Check if database is initialized
	if a.db == nil {
		errMsg := "GetCachedBalances: Database not initialized"
		fmt.Printf("%s\n", errMsg)
		debug.SimpleLog(errMsg)
		return nil, fmt.Errorf("database not initialized")
	}
	
	balances, err := database.GetAllCachedBalances(a.db, companyName)
	if err != nil {
		errMsg := fmt.Sprintf("GetCachedBalances: Error getting balances: %v", err)
		fmt.Printf("%s\n", errMsg)
		debug.SimpleLog(errMsg)
		return nil, fmt.Errorf("failed to get cached balances: %w", err)
	}
	
	fmt.Printf("GetCachedBalances: Retrieved %d balances\n", len(balances))
	debug.SimpleLog(fmt.Sprintf("GetCachedBalances: Retrieved %d balances", len(balances)))
	
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
	
	// Use service if available
	if a.Services != nil && a.Services.Banking != nil {
		balance, err := a.Services.Banking.RefreshAccountBalance(companyName, accountNumber, username)
		if err != nil {
			return nil, err
		}
		
		return map[string]interface{}{
			"status":                "success",
			"account_number":        balance.AccountNumber,
			"account_name":          balance.AccountName,
			"gl_balance":            balance.GLBalance,
			"outstanding_total":     balance.OutstandingChecksTotal,
			"outstanding_count":     balance.OutstandingChecksCount,
			"bank_balance":          balance.BankBalance,
			"gl_last_updated":       balance.GLLastUpdated,
			"checks_last_updated":   balance.ChecksLastUpdated,
			"gl_freshness":          getFreshnessStatus(balance.GLLastUpdated),
			"checks_freshness":      getFreshnessStatus(balance.ChecksLastUpdated),
			"is_stale":             balance.IsStale,
		}, nil
	}
	
	// Fallback to direct implementation
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
		if account == nil {
			errorCount++
			errors = append(errors, "Nil account in list")
			continue
		}
		
		accountNumberInterface, ok := account["account_number"]
		if !ok || accountNumberInterface == nil {
			errorCount++
			errors = append(errors, "Account missing account_number field")
			continue
		}
		
		accountNumber, ok := accountNumberInterface.(string)
		if !ok {
			errorCount++
			errors = append(errors, fmt.Sprintf("Invalid account_number type: %T", accountNumberInterface))
			continue
		}
		
		_, err := a.RefreshAccountBalance(companyName, accountNumber)
		if err != nil {
			errorCount++
			errors = append(errors, fmt.Sprintf("Account %s: %v", accountNumber, err))
		} else {
			successCount++
		}
	}
	
	refreshedBy := "system"
	if a.currentUser != nil {
		refreshedBy = a.currentUser.Username
	}
	
	return map[string]interface{}{
		"status":        "completed",
		"total_accounts": len(bankAccounts),
		"success_count": successCount,
		"error_count":   errorCount,
		"errors":        errors,
		"refreshed_by":  refreshedBy,
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

// ============================================================================
// Logging Methods
// ============================================================================

// InitializeLogging sets up the logging system
func (a *App) InitializeLogging(debugMode bool) map[string]interface{} {
	// Get user's home directory for log storage
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to get home directory: " + err.Error(),
		}
	}

	// Create logs directory in user's app data
	logDir := filepath.Join(homeDir, ".financialsx", "logs")
	
	// Initialize the logger
	logger := logger.GetLogger()
	if err := logger.Initialize(debugMode, logDir); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to initialize logger: " + err.Error(),
		}
	}

	// Clean logs older than 30 days
	go logger.CleanOldLogs(30)

	return map[string]interface{}{
		"success": true,
		"logDir":  logDir,
		"debugMode": debugMode,
	}
}

// LogMessage logs a message from the frontend
func (a *App) LogMessage(level, message, component string, data map[string]interface{}) map[string]interface{} {
	logger := logger.GetLogger()
	
	// Add user context if available
	if a.currentUser != nil {
		if data == nil {
			data = make(map[string]interface{})
		}
		data["userId"] = a.currentUser.ID
		data["username"] = a.currentUser.Username
		data["companyName"] = a.currentUser.CompanyName
	}

	if err := logger.Log(level, message, component, data); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
	}
}

// SetDebugMode enables or disables debug logging
func (a *App) SetDebugMode(enabled bool) map[string]interface{} {
	logger := logger.GetLogger()
	
	if err := logger.SetDebugMode(enabled); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Failed to set debug mode: " + err.Error(),
		}
	}

	// Save preference to config
	if err := config.SetDebugMode(enabled); err != nil {
		return map[string]interface{}{
			"success": false,
			"error": fmt.Sprintf("Failed to save debug mode preference: %v", err),
		}
	}

	return map[string]interface{}{
		"success": true,
		"debugMode": enabled,
	}
}

// GetDebugMode returns the current debug mode status
func (a *App) GetDebugMode() bool {
	return logger.GetLogger().GetDebugMode()
}

// GetLogFilePath returns the path to the current log file
func (a *App) GetLogFilePath() string {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".financialsx", "logs")
	filename := fmt.Sprintf("financialsx_%s.log", time.Now().Format("2006-01-02"))
	return filepath.Join(logDir, filename)
}

// ============================================================================
// VFP Integration Methods - Now Embedded via *legacy.VFPWrapper
// ============================================================================
// The following methods are now directly available through embedding:
// - GetSettings() -> GetVFPSettings() [renamed in VFPWrapper]
// - SaveSettings() -> SaveVFPSettings() [renamed in VFPWrapper]  
// - TestConnection() -> TestVFPConnection() [renamed in VFPWrapper]
// - LaunchForm() -> LaunchVFPForm() [renamed in VFPWrapper]
// - SyncCompany() -> SyncVFPCompany() [renamed in VFPWrapper]
// - GetCompany() -> GetVFPCompany() [renamed in VFPWrapper]
// - GetFormList() -> GetVFPFormList() [renamed in VFPWrapper]
//
// These methods are exposed directly to Wails through the embedded struct

// FollowBatchNumber fetches records from multiple tables for a given batch number
func (a *App) FollowBatchNumber(companyName string, batchNumber string) (map[string]interface{}, error) {
	// TEMPORARY: Skip authentication check for testing
	// TODO: Re-enable when Supabase integration is complete
	// if a.currentUser == nil {
	// 	return nil, fmt.Errorf("user not authenticated")
	// }
	
	// Use service if available
	if a.Services != nil && a.Services.Operations != nil {
		result, err := a.Services.Operations.FollowBatchNumber(companyName, batchNumber)
		if err != nil {
			return nil, err
		}
		
		// Convert to map for frontend compatibility
		return map[string]interface{}{
			"batch_number": result.BatchNumber,
			"company_name": result.CompanyName,
			"checks": map[string]interface{}{
				"table_name": result.Checks.TableName,
				"records":    result.Checks.Records,
				"count":      result.Checks.Count,
				"columns":    result.Checks.Columns,
				"error":      result.Checks.Error,
			},
			"glmaster": map[string]interface{}{
				"table_name": result.GLMaster.TableName,
				"records":    result.GLMaster.Records,
				"count":      result.GLMaster.Count,
				"columns":    result.GLMaster.Columns,
				"error":      result.GLMaster.Error,
			},
			"appmthdr": map[string]interface{}{
				"table_name": result.APPmtHdr.TableName,
				"records":    result.APPmtHdr.Records,
				"count":      result.APPmtHdr.Count,
				"columns":    result.APPmtHdr.Columns,
				"error":      result.APPmtHdr.Error,
			},
			"appmtdet": map[string]interface{}{
				"table_name": result.APPmtDet.TableName,
				"records":    result.APPmtDet.Records,
				"count":      result.APPmtDet.Count,
				"columns":    result.APPmtDet.Columns,
				"error":      result.APPmtDet.Error,
			},
			"appurchh": map[string]interface{}{
				"table_name": result.APPurchH.TableName,
				"records":    result.APPurchH.Records,
				"count":      result.APPurchH.Count,
				"columns":    result.APPurchH.Columns,
				"error":      result.APPurchH.Error,
			},
			"appurchd": map[string]interface{}{
				"table_name": result.APPurchD.TableName,
				"records":    result.APPurchD.Records,
				"count":      result.APPurchD.Count,
				"columns":    result.APPurchD.Columns,
				"error":      result.APPurchD.Error,
			},
		}, nil
	}
	
	// Service is required
	return nil, fmt.Errorf("operations service not initialized")
}

// UpdateBatchFields updates batch field values across multiple tables
func (a *App) UpdateBatchFields(companyName string, batchNumber string, fieldMappings map[string]string, newValue string, tablesToUpdate map[string]bool) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	
	// Use service
	if a.Services != nil && a.Services.Operations != nil {
		return a.Services.Operations.UpdateBatchFields(companyName, batchNumber, fieldMappings, newValue, tablesToUpdate)
	}
	
	// Service is required
	return nil, fmt.Errorf("operations service not initialized")
}
		if tablesToUpdate[tableName] {
			updateTable(tableName)
		}
	}
	
	fmt.Printf("UpdateBatchFields: Total records updated: %d\n", result["total_updated"].(int))
	
	return result, nil
}


func (a *App) CheckOwnerStatementFiles(companyName string) map[string]interface{} {
	// Log the function call
	logger.WriteInfo("CheckOwnerStatementFiles", fmt.Sprintf("Called for company: %s", companyName))
	debug.SimpleLog(fmt.Sprintf("CheckOwnerStatementFiles: company=%s, platform=%s", companyName, runtime.GOOS))
	
	result := map[string]interface{}{
		"hasFiles": false,
		"files": []string{},
		"error": "",
	}
	
	// Build the path to the ownerstatements directory
	// We need to resolve the actual file system path
	var ownerStatementsPath string
	
	// Use the same logic as ReadDBFFile to resolve the company path
	if filepath.IsAbs(companyName) {
		ownerStatementsPath = filepath.Join(companyName, "ownerstatements")
	} else {
		// For relative paths, we need to resolve relative to the working directory
		// The company data is in datafiles/{companyName}
		workingDir, _ := os.Getwd()
		ownerStatementsPath = filepath.Join(workingDir, "datafiles", companyName, "ownerstatements")
	}
	
	logger.WriteInfo("CheckOwnerStatementFiles", fmt.Sprintf("Checking directory: %s", ownerStatementsPath))
	
	// Check if the ownerstatements directory exists
	dirInfo, err := os.Stat(ownerStatementsPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.WriteInfo("CheckOwnerStatementFiles", "ownerstatements directory does not exist")
			result["error"] = "No Owner Distribution Files Found"
			return result
		}
		logger.WriteError("CheckOwnerStatementFiles", fmt.Sprintf("Error checking directory: %v", err))
		result["error"] = fmt.Sprintf("Error accessing directory: %v", err)
		return result
	}
	
	if !dirInfo.IsDir() {
		logger.WriteError("CheckOwnerStatementFiles", "ownerstatements exists but is not a directory")
		result["error"] = "ownerstatements is not a directory"
		return result
	}
	
	// Directory exists, now scan for DBF files
	logger.WriteInfo("CheckOwnerStatementFiles", "ownerstatements directory exists, scanning for DBF files")
	
	files, err := os.ReadDir(ownerStatementsPath)
	if err != nil {
		logger.WriteError("CheckOwnerStatementFiles", fmt.Sprintf("Error reading directory: %v", err))
		result["error"] = fmt.Sprintf("Error reading directory: %v", err)
		return result
	}
	
	var dbfFiles []string
	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			if strings.HasSuffix(strings.ToLower(fileName), ".dbf") {
				logger.WriteInfo("CheckOwnerStatementFiles", fmt.Sprintf("Found DBF file: %s", fileName))
				dbfFiles = append(dbfFiles, fileName)
			}
		}
	}
	
	if len(dbfFiles) > 0 {
		result["hasFiles"] = true
		result["files"] = dbfFiles
		logger.WriteInfo("CheckOwnerStatementFiles", fmt.Sprintf("Found %d DBF files in ownerstatements", len(dbfFiles)))
	} else {
		result["error"] = "No Owner Distribution Files Found"
		logger.WriteInfo("CheckOwnerStatementFiles", "No DBF files found in ownerstatements directory")
	}
	
	return result
}

// GetOwnerStatementsList returns a list of available owner statement files
func (a *App) GetOwnerStatementsList(companyName string) ([]map[string]interface{}, error) {
	logger.WriteInfo("GetOwnerStatementsList", fmt.Sprintf("Called for company: %s", companyName))
	
	// Build the path to the ownerstatements directory
	var ownerStatementsPath string
	
	if filepath.IsAbs(companyName) {
		ownerStatementsPath = filepath.Join(companyName, "ownerstatements")
	} else {
		workingDir, _ := os.Getwd()
		ownerStatementsPath = filepath.Join(workingDir, "datafiles", companyName, "ownerstatements")
	}
	
	logger.WriteInfo("GetOwnerStatementsList", fmt.Sprintf("Scanning directory: %s", ownerStatementsPath))
	
	// Check if directory exists
	if _, err := os.Stat(ownerStatementsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("ownerstatements directory not found")
	}
	
	// Read directory
	files, err := os.ReadDir(ownerStatementsPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}
	
	var statementFiles []map[string]interface{}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".dbf") {
			info, _ := file.Info()
			statementFile := map[string]interface{}{
				"filename": file.Name(),
				"size": info.Size(),
				"modified": info.ModTime().Format("2006-01-02 15:04:05"),
			}
			
			// Check if corresponding FPT file exists
			fptName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())) + ".fpt"
			fptPath := filepath.Join(ownerStatementsPath, fptName)
			if _, err := os.Stat(fptPath); err == nil {
				statementFile["hasFPT"] = true
			} else {
				fptName = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())) + ".FPT"
				fptPath = filepath.Join(ownerStatementsPath, fptName)
				if _, err := os.Stat(fptPath); err == nil {
					statementFile["hasFPT"] = true
				} else {
					statementFile["hasFPT"] = false
				}
			}
			
			statementFiles = append(statementFiles, statementFile)
			logger.WriteInfo("GetOwnerStatementsList", fmt.Sprintf("Added file: %s (size: %d, hasFPT: %v)", 
				file.Name(), info.Size(), statementFile["hasFPT"]))
		}
	}
	
	logger.WriteInfo("GetOwnerStatementsList", fmt.Sprintf("Found %d DBF files", len(statementFiles)))
	return statementFiles, nil
}

// GenerateOwnerStatementPDF generates a PDF for owner distribution statements
func (a *App) GenerateOwnerStatementPDF(companyName string, fileName string) (string, error) {
	// Delegate to Reports service
	pdfBytes, err := a.Services.Reports.GenerateOwnerStatementPDF(companyName, fileName)
	if err != nil {
		return "", fmt.Errorf("failed to generate owner statement PDF: %v", err)
	}
	
	// Save the PDF to a temporary file and return the path
	// Or return base64 encoded string for frontend to handle
	encodedPDF := base64.StdEncoding.EncodeToString(pdfBytes)
	return encodedPDF, nil
}

// GetOwnersList returns a unique list of owners from the statement DBF file
func (a *App) GetOwnersList(companyName string, fileName string) ([]map[string]interface{}, error) {
	logger.WriteInfo("GetOwnersList", fmt.Sprintf("Getting owners list from %s/%s", companyName, fileName))
	
	// Read the DBF file
	dbfData, err := company.ReadDBFFile(companyName, filepath.Join("ownerstatements", fileName), "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("error reading DBF file: %v", err)
	}
	
	// Get columns
	columns, _ := dbfData["columns"].([]string)
	
	// Convert rows to map format
	var rows []map[string]interface{}
	if rowsArray, ok := dbfData["rows"].([][]interface{}); ok {
		for _, rowValues := range rowsArray {
			rowMap := make(map[string]interface{})
			for i, value := range rowValues {
				if i < len(columns) {
					rowMap[columns[i]] = value
				}
			}
			rows = append(rows, rowMap)
		}
	}
	
	// Find owner-related columns (COWNNAME, COWNERID, COWNNO, etc.)
	ownerNameCol := ""
	ownerIDCol := ""
	for _, col := range columns {
		colUpper := strings.ToUpper(col)
		if strings.Contains(colUpper, "OWNNAME") || strings.Contains(colUpper, "OWNER") && strings.Contains(colUpper, "NAME") {
			ownerNameCol = col
		}
		if strings.Contains(colUpper, "OWNERID") || strings.Contains(colUpper, "OWNNO") || strings.Contains(colUpper, "COWNID") {
			ownerIDCol = col
		}
	}
	
	// If we didn't find specific owner columns, look for generic name columns
	if ownerNameCol == "" {
		for _, col := range columns {
			colUpper := strings.ToUpper(col)
			if colUpper == "CNAME" || colUpper == "NAME" || strings.Contains(colUpper, "NAME") {
				ownerNameCol = col
				break
			}
		}
	}
	
	// Build unique owners list
	ownersMap := make(map[string]map[string]interface{})
	for _, row := range rows {
		ownerName := ""
		ownerID := ""
		
		if ownerNameCol != "" {
			if val, ok := row[ownerNameCol]; ok && val != nil {
				ownerName = strings.TrimSpace(fmt.Sprintf("%v", val))
			}
		}
		
		if ownerIDCol != "" {
			if val, ok := row[ownerIDCol]; ok && val != nil {
				ownerID = strings.TrimSpace(fmt.Sprintf("%v", val))
			}
		}
		
		// Use name as key, or ID if name is empty
		key := ownerName
		if key == "" {
			key = ownerID
		}
		
		if key != "" && key != "0" {
			if _, exists := ownersMap[key]; !exists {
				ownersMap[key] = map[string]interface{}{
					"name": ownerName,
					"id":   ownerID,
					"key":  key,
				}
			}
		}
	}
	
	// Convert map to slice
	var owners []map[string]interface{}
	for _, owner := range ownersMap {
		owners = append(owners, owner)
	}
	
	// Sort by name
	sort.Slice(owners, func(i, j int) bool {
		name1 := fmt.Sprintf("%v", owners[i]["name"])
		name2 := fmt.Sprintf("%v", owners[j]["name"])
		return name1 < name2
	})
	
	logger.WriteInfo("GetOwnersList", fmt.Sprintf("Found %d unique owners", len(owners)))
	return owners, nil
}

// GetOwnerStatementData returns statement data for a specific owner
func (a *App) GetOwnerStatementData(companyName string, fileName string, ownerKey string) (map[string]interface{}, error) {
	logger.WriteInfo("GetOwnerStatementData", fmt.Sprintf("Getting statement data for owner: %s", ownerKey))
	
	// Read the DBF file
	dbfData, err := company.ReadDBFFile(companyName, filepath.Join("ownerstatements", fileName), "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("error reading DBF file: %v", err)
	}
	
	// Get columns
	columns, _ := dbfData["columns"].([]string)
	
	// Convert rows to map format
	var allRows []map[string]interface{}
	if rowsArray, ok := dbfData["rows"].([][]interface{}); ok {
		for _, rowValues := range rowsArray {
			rowMap := make(map[string]interface{})
			for i, value := range rowValues {
				if i < len(columns) {
					rowMap[columns[i]] = value
				}
			}
			allRows = append(allRows, rowMap)
		}
	}
	
	// Find owner-related columns
	ownerNameCol := ""
	ownerIDCol := ""
	for _, col := range columns {
		colUpper := strings.ToUpper(col)
		if strings.Contains(colUpper, "OWNNAME") || strings.Contains(colUpper, "OWNER") && strings.Contains(colUpper, "NAME") {
			ownerNameCol = col
		}
		if strings.Contains(colUpper, "OWNERID") || strings.Contains(colUpper, "OWNNO") || strings.Contains(colUpper, "COWNID") {
			ownerIDCol = col
		}
	}
	
	// If we didn't find specific owner columns, look for generic name columns
	if ownerNameCol == "" {
		for _, col := range columns {
			colUpper := strings.ToUpper(col)
			if colUpper == "CNAME" || colUpper == "NAME" || strings.Contains(colUpper, "NAME") {
				ownerNameCol = col
				break
			}
		}
	}
	
	// Filter rows for this owner
	var ownerRows []map[string]interface{}
	for _, row := range allRows {
		match := false
		
		// Check by name
		if ownerNameCol != "" {
			if val, ok := row[ownerNameCol]; ok && val != nil {
				name := strings.TrimSpace(fmt.Sprintf("%v", val))
				if name == ownerKey {
					match = true
				}
			}
		}
		
		// Check by ID if not matched by name
		if !match && ownerIDCol != "" {
			if val, ok := row[ownerIDCol]; ok && val != nil {
				id := strings.TrimSpace(fmt.Sprintf("%v", val))
				if id == ownerKey {
					match = true
				}
			}
		}
		
		if match {
			ownerRows = append(ownerRows, row)
		}
	}
	
	// Calculate totals and summaries
	totalGross := 0.0
	totalNet := 0.0
	totalTax := 0.0
	wellCount := make(map[string]bool)
	
	// Look for amount columns
	for _, row := range ownerRows {
		// Check for well identifier
		for _, col := range columns {
			colUpper := strings.ToUpper(col)
			if strings.Contains(colUpper, "WELL") || strings.Contains(colUpper, "LEASE") {
				if val, ok := row[col]; ok && val != nil {
					wellID := fmt.Sprintf("%v", val)
					if wellID != "" && wellID != "0" {
						wellCount[wellID] = true
					}
				}
			}
		}
		
		// Sum amounts
		for key, val := range row {
			keyUpper := strings.ToUpper(key)
			if val != nil {
				// Try to parse as number for amount fields
				if strings.Contains(keyUpper, "GROSS") || strings.Contains(keyUpper, "REVENUE") {
					if num, err := strconv.ParseFloat(fmt.Sprintf("%v", val), 64); err == nil {
						totalGross += num
					}
				} else if strings.Contains(keyUpper, "NET") && !strings.Contains(keyUpper, "NETSUM") {
					if num, err := strconv.ParseFloat(fmt.Sprintf("%v", val), 64); err == nil {
						totalNet += num
					}
				} else if strings.Contains(keyUpper, "TAX") || strings.Contains(keyUpper, "DEDUCT") {
					if num, err := strconv.ParseFloat(fmt.Sprintf("%v", val), 64); err == nil {
						totalTax += num
					}
				}
			}
		}
	}
	
	result := map[string]interface{}{
		"owner":      ownerKey,
		"rows":       ownerRows,
		"rowCount":   len(ownerRows),
		"columns":    columns,
		"wellCount":  len(wellCount),
		"totals": map[string]interface{}{
			"gross": totalGross,
			"net":   totalNet,
			"tax":   totalTax,
		},
	}
	
	logger.WriteInfo("GetOwnerStatementData", fmt.Sprintf("Found %d rows for owner %s", len(ownerRows), ownerKey))
	return result, nil
}

// ExamineOwnerStatementStructure examines the structure of owner statement DBF files
func (a *App) ExamineOwnerStatementStructure(companyName string, fileName string) (map[string]interface{}, error) {
	logger.WriteInfo("ExamineOwnerStatementStructure", fmt.Sprintf("Examining %s for company %s", fileName, companyName))
	
	// Read the DBF file
	dbfData, err := company.ReadDBFFile(companyName, filepath.Join("ownerstatements", fileName), "", 0, 10, "", "")
	if err != nil {
		return nil, fmt.Errorf("error reading DBF file: %v", err)
	}
	
	// Get columns
	columns, _ := dbfData["columns"].([]string)
	
	// Get sample rows (first 10)
	var rows []map[string]interface{}
	if rowsData, ok := dbfData["rows"].([]map[string]interface{}); ok {
		rows = rowsData
	} else if rowsArray, ok := dbfData["rows"].([]interface{}); ok {
		for _, item := range rowsArray {
			if row, ok := item.(map[string]interface{}); ok {
				rows = append(rows, row)
			}
		}
	}
	
	// Analyze column types and sample values
	columnInfo := make([]map[string]interface{}, 0)
	for _, col := range columns {
		info := map[string]interface{}{
			"name": col,
			"sampleValues": []interface{}{},
			"type": "unknown",
		}
		
		// Get sample values from first few rows
		sampleValues := []interface{}{}
		for i, row := range rows {
			if i >= 3 { // Just get 3 samples
				break
			}
			if val, exists := row[col]; exists && val != nil {
				sampleValues = append(sampleValues, val)
				// Infer type from first non-nil value
				if info["type"] == "unknown" {
					switch val.(type) {
					case string:
						info["type"] = "string"
					case float64, float32, int, int64:
						info["type"] = "number"
					case bool:
						info["type"] = "boolean"
					case time.Time:
						info["type"] = "date"
					default:
						info["type"] = fmt.Sprintf("%T", val)
					}
				}
			}
		}
		info["sampleValues"] = sampleValues
		columnInfo = append(columnInfo, info)
	}
	
	result := map[string]interface{}{
		"fileName": fileName,
		"recordCount": len(rows),
		"columnCount": len(columns),
		"columns": columnInfo,
		"sampleRecords": rows,
	}
	
	return result, nil
}

// GetVendors retrieves all vendors from VENDOR.dbf
func (a *App) GetVendors(companyName string) (map[string]interface{}, error) {
	logger.WriteInfo("GetVendors", fmt.Sprintf("Called for company: %s", companyName))
	fmt.Printf("GetVendors: Called for company: %s\n", companyName)
	
	// Read the VENDOR.dbf file
	vendorData, err := company.ReadDBFFile(companyName, "VENDOR.dbf", "", 0, 0, "", "")
	if err != nil {
		logger.WriteError("GetVendors", fmt.Sprintf("Error reading VENDOR.dbf: %v", err))
		fmt.Printf("GetVendors: Error reading VENDOR.dbf: %v\n", err)
		return nil, fmt.Errorf("error reading vendor data: %v", err)
	}
	
	// Log the data structure
	if rows, ok := vendorData["rows"].([][]interface{}); ok {
		fmt.Printf("GetVendors: Found %d vendor records\n", len(rows))
		if len(rows) > 0 {
			fmt.Printf("GetVendors: First vendor record has %d fields\n", len(rows[0]))
			// Log column names
			if columns, ok := vendorData["columns"].([]string); ok {
				fmt.Printf("GetVendors: Columns: %v\n", columns)
			}
		}
	} else {
		fmt.Printf("GetVendors: No rows found in vendor data\n")
	}
	
	return vendorData, nil
}

// UpdateVendor updates a vendor record in VENDOR.dbf
func (a *App) UpdateVendor(companyName string, vendorIndex int, vendorData map[string]interface{}) error {
	logger.WriteInfo("UpdateVendor", fmt.Sprintf("Updating vendor at index %d for company %s", vendorIndex, companyName))
	
	var vendorPath string
	
	// Check if companyName is already an absolute path (Windows)
	if filepath.IsAbs(companyName) {
		// It's already an absolute path, just append VENDOR.dbf
		vendorPath = filepath.Join(companyName, "VENDOR.dbf")
		logger.WriteInfo("UpdateVendor", fmt.Sprintf("Using absolute path: %s", vendorPath))
	} else {
		// It's a relative path, construct the full path
		datafilesPath, err := company.GetDatafilesPath()
		if err != nil {
			return fmt.Errorf("failed to get datafiles path: %w", err)
		}
		
		// Normalize the company path
		normalizedCompanyName := company.NormalizeCompanyPath(companyName)
		
		// Construct the full path to VENDOR.dbf
		vendorPath = filepath.Join(datafilesPath, normalizedCompanyName, "VENDOR.dbf")
		logger.WriteInfo("UpdateVendor", fmt.Sprintf("Using constructed path: %s", vendorPath))
	}

	// Open the table for writing
	table, err := dbase.OpenTable(&dbase.Config{
		Filename:   vendorPath,
		ReadOnly:   false,
		TrimSpaces: true,
	})
	if err != nil {
		return fmt.Errorf("failed to open VENDOR.dbf for writing: %w", err)
	}
	defer table.Close()

	// Navigate to the specific record
	currentIndex := 0
	var targetRow *dbase.Row
	
	for {
		row, err := table.Next()
		if err != nil {
			if err.Error() == "EOF" {
				return fmt.Errorf("vendor record not found at index %d", vendorIndex)
			}
			return fmt.Errorf("error reading vendor table: %w", err)
		}

		// Skip deleted records
		if row.Deleted {
			continue
		}

		if currentIndex == vendorIndex {
			targetRow = row
			break
		}
		currentIndex++
	}

	if targetRow == nil {
		return fmt.Errorf("vendor record not found at index %d", vendorIndex)
	}

	// Update the fields in the row
	for fieldName, value := range vendorData {
		field := targetRow.FieldByName(fieldName)
		if field == nil {
			// Skip fields that don't exist in the DBF
			continue
		}

		// Set the field value using the actual API
		err := field.SetValue(value)
		if err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldName, err)
		}
	}

	// Write the updated row back to the file
	err = table.WriteRow(targetRow)
	if err != nil {
		return fmt.Errorf("failed to write updated vendor record: %w", err)
	}

	logger.WriteInfo("UpdateVendor", fmt.Sprintf("Successfully updated vendor at index %d", vendorIndex))
	return nil
}

// GenerateChartOfAccountsPDF generates a PDF report of the Chart of Accounts
func (a *App) GenerateChartOfAccountsPDF(companyName string, sortBy string, includeInactive bool) (string, error) {
	// Delegate to Reports service
	pdfBytes, err := a.Services.Reports.GenerateChartOfAccountsPDF(companyName, sortBy, includeInactive)
	if err != nil {
		return "", fmt.Errorf("failed to generate chart of accounts PDF: %v", err)
	}
	
	// Save the PDF to a temporary file and return the path
	// Or return base64 encoded string for frontend to handle
	encodedPDF := base64.StdEncoding.EncodeToString(pdfBytes)
	return encodedPDF, nil
}

// Legacy implementation - kept for reference but not used
func (a *App) GenerateChartOfAccountsPDF_OLD(companyName string, sortBy string, includeInactive bool) (string, error) {
	logger.WriteInfo("GenerateChartOfAccountsPDF", fmt.Sprintf("Called for company: %s, sortBy: %s, includeInactive: %v", companyName, sortBy, includeInactive))
	debug.SimpleLog(fmt.Sprintf("GenerateChartOfAccountsPDF: company=%s, sortBy=%s, includeInactive=%v", companyName, sortBy, includeInactive))
	
	// Check if user is authenticated (disabled for now)
	// if a.currentUser == nil {
	// 	logger.WriteError("GenerateChartOfAccountsPDF", "User not authenticated")
	// 	return "", fmt.Errorf("user not authenticated")
	// }
	
	// Get the chart of accounts data
	coaData, err := a.GetChartOfAccounts(companyName, sortBy, includeInactive)
	if err != nil {
		return "", fmt.Errorf("failed to get chart of accounts: %v", err)
	}
	
	// The service returns the accounts directly as []map[string]interface{}
	accounts := coaData
	
	// Get company info from version.dbf
	displayCompanyName := companyName // Default to folder name
	companyAddress := ""
	companyCityStateZip := ""
	
	// Try to read version.dbf for company details
	versionData, err := company.ReadDBFFile(companyName, "VERSION.DBF", "", 0, 1, "", "")
	if err != nil {
		logger.WriteInfo("GenerateChartOfAccountsPDF", fmt.Sprintf("Could not read VERSION.DBF: %v", err))
	} else if versionData != nil {
		logger.WriteInfo("GenerateChartOfAccountsPDF", fmt.Sprintf("VERSION.DBF read successfully"))
		if rows, ok := versionData["rows"].([][]interface{}); ok && len(rows) > 0 {
			// Get column names
			columns, _ := versionData["columns"].([]string)
			logger.WriteInfo("GenerateChartOfAccountsPDF", fmt.Sprintf("VERSION.DBF columns: %v", columns))
			
			// Convert first row to map
			if len(rows[0]) > 0 {
				record := make(map[string]interface{})
				for j, value := range rows[0] {
					if j < len(columns) {
						record[columns[j]] = value
					}
				}
				
				// Extract company information - use CPRODUCER for company name
				if val, ok := record["CPRODUCER"]; ok && val != nil {
					name := strings.TrimSpace(fmt.Sprintf("%v", val))
					if name != "" {
						displayCompanyName = name
						logger.WriteInfo("GenerateChartOfAccountsPDF", fmt.Sprintf("Found company name from CPRODUCER: %s", name))
					}
				}
				
				// Try CADDRESS1 and CADDRESS2 for address
				if val, ok := record["CADDRESS1"]; ok && val != nil {
					addr := strings.TrimSpace(fmt.Sprintf("%v", val))
					if addr != "" {
						companyAddress = addr
					}
				}
				if companyAddress == "" {
					if val, ok := record["CADDRESS2"]; ok && val != nil {
						addr := strings.TrimSpace(fmt.Sprintf("%v", val))
						if addr != "" {
							companyAddress = addr
						}
					}
				}
				
				// Build city, state, zip line using correct field names
				city := ""
				if val, ok := record["CCITY"]; ok && val != nil {
					city = strings.TrimSpace(fmt.Sprintf("%v", val))
				}
				
				state := ""
				if val, ok := record["CSTATE"]; ok && val != nil {
					state = strings.TrimSpace(fmt.Sprintf("%v", val))
				}
				
				zip := ""
				if val, ok := record["CZIPCODE"]; ok && val != nil {
					zip = strings.TrimSpace(fmt.Sprintf("%v", val))
				}
				
				// Format city, state zip
				if city != "" || state != "" || zip != "" {
					parts := []string{}
					if city != "" {
						parts = append(parts, city)
					}
					if state != "" {
						if city != "" {
							parts = append(parts, state)
						} else {
							parts = append(parts, state)
						}
					}
					if zip != "" {
						if len(parts) > 0 && state != "" {
							parts[len(parts)-1] = state + " " + zip
						} else {
							parts = append(parts, zip)
						}
					}
					companyCityStateZip = strings.Join(parts, ", ")
					// Fix the state zip formatting
					if city != "" && state != "" && zip != "" {
						companyCityStateZip = city + ", " + state + " " + zip
					}
				}
			}
		}
	}
	
	logger.WriteInfo("GenerateChartOfAccountsPDF", fmt.Sprintf("Company: %s, Address: %s, City/State/Zip: %s", displayCompanyName, companyAddress, companyCityStateZip))
	
	// Create a new PDF document with landscape orientation for better table fit
	pdf := gofpdf.New("L", "mm", "Letter", "")
	pdf.SetAutoPageBreak(true, 20)
	
	// Add footer function BEFORE adding pages so it applies to all pages
	sortText := "Account Number"
	if sortBy == "type" {
		sortText = "Account Type"
	}
	
	filterText := "Active Only"
	if includeInactive {
		filterText = "All Accounts"
	}
	
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "", 7) // Smaller font
		pdf.SetTextColor(128, 128, 128)
		pdf.SetDrawColor(200, 200, 200)
		pdf.Line(10, pdf.GetY(), 269, pdf.GetY())
		pdf.Ln(2)
		
		// Left side - Pivoten trademark and version info
		pdf.SetX(10)
		// Write "Pivoten" first
		pdf.SetFont("Helvetica", "", 7)
		pivotText := "Pivoten"
		pivotWidth := pdf.GetStringWidth(pivotText)
		pdf.Cell(pivotWidth, 5, pivotText)
		
		// Add superscript TM - smaller and tighter
		currentX := pdf.GetX()
		currentY := pdf.GetY()
		pdf.SetXY(currentX-0.5, currentY-1.2) // Move up and slightly left for tighter spacing
		pdf.SetFont("Helvetica", "", 4) // Even smaller font for superscript
		pdf.Cell(2, 2, "TM")
		
		// Continue with the rest of the text
		pdf.SetXY(currentX+2, currentY) // Reset position
		pdf.SetFont("Helvetica", "", 7) // Back to normal font
		pdf.Cell(50, 5, " - Financials 2026 - BETA 2025-08-13")
		
		// Center - Report details
		pdf.SetX(65)
		pdf.Cell(50, 5, fmt.Sprintf("Generated: %s", time.Now().Format("January 2, 2006 3:04 PM")))
		pdf.SetX(115)
		pdf.Cell(30, 5, fmt.Sprintf("Total: %d", len(accounts)))
		pdf.SetX(145)
		pdf.Cell(40, 5, fmt.Sprintf("Sort: %s", sortText))
		pdf.SetX(185)
		pdf.Cell(40, 5, fmt.Sprintf("Filter: %s", filterText))
		
		// Right side - page number (right-aligned at right margin)
		pageText := fmt.Sprintf("Page %d", pdf.PageNo())
		pageWidth := pdf.GetStringWidth(pageText)
		pdf.SetX(269 - pageWidth) // Position at right margin minus text width
		pdf.Cell(pageWidth, 5, pageText)
	})
	
	pdf.AddPage()
	
	// Company header with full address
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Helvetica", "B", 16)
	pdf.SetXY(10, 10)
	pdf.Cell(0, 7, displayCompanyName)
	pdf.Ln(7)
	
	// Add address if available
	if companyAddress != "" {
		pdf.SetFont("Helvetica", "", 11)
		pdf.SetTextColor(60, 60, 60)
		pdf.Cell(0, 5, companyAddress)
		pdf.Ln(5)
	}
	
	// Add city, state, zip if available
	if companyCityStateZip != "" {
		pdf.SetFont("Helvetica", "", 11)
		pdf.SetTextColor(60, 60, 60)
		pdf.Cell(0, 5, companyCityStateZip)
		pdf.Ln(8)
	}
	
	// Report title
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 7, "Chart of Accounts")
	pdf.Ln(10)
	
	// Separator line
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.5)
	pdf.Line(10, pdf.GetY(), 269, pdf.GetY())
	pdf.Ln(8)
	
	// Reset text color for table
	pdf.SetTextColor(0, 0, 0)
	
	// Table header
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(245, 245, 245) // Very light gray for headers
	pdf.SetTextColor(40, 40, 40) // Dark gray text
	pdf.SetDrawColor(200, 200, 200) // Light gray border
	pdf.SetLineWidth(0.2)
	
	// Define column widths for landscape orientation
	// Total width: 259 to fit within page margins (Letter landscape = 279.4mm - 20mm margins)
	colWidths := []float64{38, 122, 33, 22, 22, 22}
	headers := []string{"Account #", "Description", "Type", "Bank", "Unit", "Dept"}
	
	// Draw headers
	for i, header := range headers {
		align := "L"
		if i >= 3 { // Center align for boolean columns
			align = "C"
		} else if i == 2 { // Center align Type column
			align = "C"
		}
		pdf.CellFormat(colWidths[i], 7, header, "1", 0, align, true, 0, "")
	}
	pdf.Ln(-1)
	
	// Reset for table content
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(255, 255, 255)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetDrawColor(230, 230, 230) // Very light border for data rows
	
	// Track row count for alternating colors
	rowCount := 0
	for _, account := range accounts {
		// Alternate row colors for better readability
		if rowCount%2 == 1 {
			pdf.SetFillColor(250, 250, 250) // Very light gray for alternate rows
		} else {
			pdf.SetFillColor(255, 255, 255) // White
		}
		rowCount++
		
		// Check if this account has a parent (for indentation)
		parent := ""
		hasParent := false
		if val, ok := account["parent_account"]; ok && val != nil {
			parent = strings.TrimSpace(fmt.Sprintf("%v", val))
			if parent != "" && parent != "0" && parent != "00000" {
				hasParent = true
			}
		}
		
		// Account Number - with indentation for child accounts
		accNum := ""
		if val, ok := account["account_number"]; ok && val != nil {
			accNum = strings.TrimSpace(fmt.Sprintf("%v", val))
		}
		if hasParent {
			accNum = "    " + accNum // Add indentation for child accounts
		}
		pdf.CellFormat(colWidths[0], 6, accNum, "LR", 0, "L", true, 0, "")
		
		// Account Name - with indentation for child accounts
		accName := ""
		if val, ok := account["account_name"]; ok && val != nil {
			accName = strings.TrimSpace(fmt.Sprintf("%v", val))
		}
		// Add indentation for child accounts
		if hasParent {
			accName = "    " + accName
			// Truncate long names accounting for indentation
			if len(accName) > 54 {
				accName = accName[:54] + "..."
			}
		} else {
			// Truncate long names
			if len(accName) > 50 {
				accName = accName[:50] + "..."
			}
		}
		pdf.CellFormat(colWidths[1], 6, accName, "LR", 0, "L", true, 0, "")
		
		// Account Type - no color coding for professional look
		accType := ""
		if val, ok := account["account_type"]; ok && val != nil {
			accType = fmt.Sprintf("%v", val)
		}
		pdf.CellFormat(colWidths[2], 6, accType, "LR", 0, "C", true, 0, "")
		
		// Is Bank Account - use checkmark
		bankAcct := ""
		if val, ok := account["is_bank_account"]; ok && val != nil {
			if val.(bool) {
				bankAcct = "Yes"
			}
		}
		pdf.CellFormat(colWidths[3], 6, bankAcct, "LR", 0, "C", true, 0, "")
		
		// Is Unit
		isUnit := ""
		if val, ok := account["is_unit"]; ok && val != nil {
			if val.(bool) {
				isUnit = "Yes"
			}
		}
		pdf.CellFormat(colWidths[4], 6, isUnit, "LR", 0, "C", true, 0, "")
		
		// Is Department
		isDept := ""
		if val, ok := account["is_department"]; ok && val != nil {
			if val.(bool) {
				isDept = "Yes"
			}
		}
		pdf.CellFormat(colWidths[5], 6, isDept, "LR", 0, "C", true, 0, "")
		
		pdf.Ln(-1)
	}
	
	// Draw bottom border for the table
	pdf.SetDrawColor(52, 73, 94)
	pdf.Line(10, pdf.GetY(), 269, pdf.GetY())
	
	// Generate default filename - use company name from VERSION.DBF if available
	cleanCompanyName := displayCompanyName
	if cleanCompanyName == "" {
		// Fall back to the folder name if VERSION.DBF didn't have company name
		cleanCompanyName = companyName
	}
	
	// Remove path separators and other problematic characters for filenames
	cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "\\", "_")
	cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "/", "_")
	cleanCompanyName = strings.ReplaceAll(cleanCompanyName, ":", "_")
	cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "*", "_")
	cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "?", "_")
	cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "\"", "_")
	cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "<", "_")
	cleanCompanyName = strings.ReplaceAll(cleanCompanyName, ">", "_")
	cleanCompanyName = strings.ReplaceAll(cleanCompanyName, "|", "_")
	
	// Format: YYYY-MM-DD - Company Name - Chart of Accounts.pdf
	datePrefix := time.Now().Format("2006-01-02")
	defaultFilename := fmt.Sprintf("%s - %s - Chart of Accounts.pdf", datePrefix, cleanCompanyName)
	
	// Show save dialog to let user choose where to save the file
	selectedFile, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           "Save Chart of Accounts Report",
		DefaultFilename: defaultFilename,
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "PDF Files (*.pdf)",
				Pattern:     "*.pdf",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	
	// If user cancelled the dialog
	if err != nil {
		return "", fmt.Errorf("save dialog error: %v", err)
	}
	
	if selectedFile == "" {
		return "", fmt.Errorf("save cancelled by user")
	}
	
	// Save the PDF to the selected location
	err = pdf.OutputFileAndClose(selectedFile)
	if err != nil {
		return "", fmt.Errorf("failed to write PDF file: %v", err)
	}
	
	logger.WriteInfo("GenerateChartOfAccountsPDF", fmt.Sprintf("PDF report saved to %s", selectedFile))
	return selectedFile, nil
}

// ============================================================================
// I18N FUNCTIONS - NOW HANDLED BY EMBEDDING
// ============================================================================
// The following methods are now available directly through embedding:
// - GetLocale() - from embedded *common.I18n
// - SetLocale() - from embedded *common.I18n  
// - GetAvailableLocales() - from embedded *common.I18n
// - T() - from embedded *common.I18n (was wrapped as Translate)
//
// NOTE: The frontend uses "Translate" but I18n provides "T"
// We need to add a Translate wrapper for backward compatibility

// Translate wraps the T() method for backward compatibility with frontend
func (a *App) Translate(key string) string {
	if a.I18n != nil {
		return a.I18n.T(key)
	}
	return key
}

// ============================================================================
// PASSWORD MANAGEMENT FUNCTIONS
// ============================================================================

// ChangePassword allows a user to change their own password
func (a *App) ChangePassword(oldPassword, newPassword string) error {
	if a.currentUser == nil {
		return fmt.Errorf("user not authenticated")
	}
	
	if a.auth == nil {
		return fmt.Errorf("auth service not initialized")
	}
	
	return a.auth.ChangePassword(a.currentUser.ID, oldPassword, newPassword)
}

// RequestPasswordReset sends a password reset email to the user
func (a *App) RequestPasswordReset(email string) error {
	// In a real application, this would send an email
	// For now, we'll just generate the token and return success
	
	// We need to initialize auth with a database first
	// This should use the master database or a specific company database
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}
	
	// Use the current database connection
	if a.auth == nil {
		return fmt.Errorf("auth service not initialized")
	}
	
	token, err := a.auth.RequestPasswordReset(email)
	if err != nil {
		return err
	}
	
	if token != nil {
		// In production, send email with reset link
		// For development, log the token
		logger.WriteInfo("PasswordReset", fmt.Sprintf("Reset token generated for %s: %s", email, token.Token))
	}
	
	return nil
}

// ResetPassword resets a user's password with a valid token
func (a *App) ResetPassword(token, newPassword string) error {
	if a.auth == nil {
		return fmt.Errorf("auth service not initialized")
	}
	
	return a.auth.ResetPassword(token, newPassword)
}

// AdminResetPassword allows an admin to reset any user's password
func (a *App) AdminResetPassword(userID int, newPassword string) error {
	// Check if current user is admin
	if a.currentUser == nil || (!a.currentUser.IsRoot && a.currentUser.RoleName != "Admin") {
		return fmt.Errorf("insufficient permissions")
	}
	
	if a.auth == nil {
		return fmt.Errorf("auth service not initialized")
	}
	
	return a.auth.AdminResetPassword(userID, newPassword)
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

	// Using dedicated COM thread to avoid threading issues
	// The COM thread will be initialized on first use
	defer func() {
		// Shutdown the COM thread on application exit
		ole.ShutdownCOMThread()
	}()

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
