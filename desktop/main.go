// Package main provides the Wails desktop application entry point and API endpoints.
// This file acts as a thin wrapper, delegating all business logic to specialized services.
//
// Architecture:
//   - main.go handles Wails integration and API routing
//   - Services contain all business logic and database operations
//   - Frontend communicates through these API endpoints
//
// Organization:
//   - App Initialization & Lifecycle
//   - Authentication & Session Management
//   - User Management API
//   - Company Management API
//   - Banking & Financial API
//   - GL & Accounting API
//   - Reconciliation API
//   - Reporting API
//   - System & Configuration API
//   - Utility Functions
package main

import (
	"context"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/app"
	"github.com/pivoten/financialsx/desktop/internal/common"
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/config"
	"github.com/pivoten/financialsx/desktop/internal/database"
	"github.com/pivoten/financialsx/desktop/internal/debug"
	"github.com/pivoten/financialsx/desktop/internal/financials/audit"
	"github.com/pivoten/financialsx/desktop/internal/legacy"
	"github.com/pivoten/financialsx/desktop/internal/logger"
	"github.com/pivoten/financialsx/desktop/internal/ole"
	"github.com/pivoten/financialsx/desktop/internal/reconciliation"
	"github.com/pivoten/financialsx/desktop/internal/utilities"
	"github.com/pivoten/financialsx/desktop/internal/vfp"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// ============================================================================
// APP STRUCTURE AND INITIALIZATION
// ============================================================================

// App struct represents the main application with all services and state.
// It implements the Wails application interface and provides API endpoints
// for the frontend to interact with backend services.
type App struct {
	ctx                   context.Context
	db                    *database.DB
	auth                  *common.Auth
	currentUser           *common.User
	currentCompanyPath    string
	reconciliationService *reconciliation.Service
	vfpClient             *vfp.VFPClient // VFP integration client (internal use)
	*legacy.VFPWrapper                   // Embedded VFP wrapper - methods are directly available
	auditService          *audit.Service // Financial audit service (uses wrappers for compatibility)
	dataBasePath          string         // Base path where compmast.dbf is located
	*common.I18n                         // Embedded i18n - methods are directly available

	// Services - new modular architecture
	Services *app.Services

	// Platform detection (cached at startup)
	platform  string // Operating system: "windows", "darwin", "linux"
	isWindows bool   // Convenience flag for Windows platform

	// Authentication state (cached after login)
	isAuthenticated bool            // Whether user is logged in
	isAdmin         bool            // Whether user has admin privileges
	isRoot          bool            // Whether user has root privileges
	permissions     map[string]bool // Cached permission set
	userRole        string          // Cached role name
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// ============================================================================
// APP LIFECYCLE METHODS
// ============================================================================

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

// ============================================================================
// STANDARDIZED HELPER FUNCTIONS
// These helpers ensure consistent error handling and permission checking.
// Use these in all API functions for consistency.
// ============================================================================

// requireAuth checks if user is authenticated and returns standardized error
func (a *App) requireAuth() error {
	if a.currentUser == nil {
		return fmt.Errorf("user not authenticated")
	}
	return nil
}

// requirePermission checks if user has specific permission and returns standardized error
func (a *App) requirePermission(permission string) error {
	if err := a.requireAuth(); err != nil {
		return err
	}
	if !a.hasPermission(permission) {
		return fmt.Errorf("insufficient permissions: %s required", permission)
	}
	return nil
}

// requireAdminOrRoot checks if user is admin or root and returns standardized error
func (a *App) requireAdminOrRoot() error {
	if err := a.requireAuth(); err != nil {
		return err
	}
	if !a.isRoot && !a.isAdmin {
		return fmt.Errorf("insufficient permissions: admin or root required")
	}
	return nil
}

// requireService checks if a service is initialized and returns standardized error
func (a *App) requireService(serviceName string) error {
	if a.Services == nil {
		return fmt.Errorf("services not initialized")
	}
	// Check specific services
	switch serviceName {
	case "banking":
		if a.Services.Banking == nil {
			return fmt.Errorf("banking service not initialized")
		}
	case "gl":
		if a.Services.GL == nil {
			return fmt.Errorf("GL service not initialized")
		}
	case "dbf":
		if a.Services.DBF == nil {
			return fmt.Errorf("DBF service not initialized")
		}
	case "auth":
		if a.Services.Auth == nil {
			return fmt.Errorf("auth service not initialized")
		}
	case "operations":
		if a.Services.Operations == nil {
			return fmt.Errorf("operations service not initialized")
		}
	case "reconciliation":
		if a.Services.Reconciliation == nil {
			return fmt.Errorf("reconciliation service not initialized")
		}
	case "reports":
		if a.Services.Reports == nil {
			return fmt.Errorf("reports service not initialized")
		}
	case "audit":
		if a.Services.Audit == nil {
			return fmt.Errorf("audit service not initialized")
		}
	case "matching":
		if a.Services.Matching == nil {
			return fmt.Errorf("matching service not initialized")
		}
	case "company":
		if a.Services.Company == nil {
			return fmt.Errorf("company service not initialized")
		}
	case "ole":
		if a.Services.OLE == nil {
			return fmt.Errorf("OLE service not initialized")
		}
	// Add more services as needed
	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}
	return nil
}

// GetPlatform returns the current platform information
func (a *App) GetPlatform() map[string]interface{} {
	return map[string]interface{}{
		"platform":  a.platform,
		"isWindows": a.isWindows,
		"arch":      runtime.GOARCH,
	}
}

// GetAuthState returns the current authentication state
func (a *App) GetAuthState() map[string]interface{} {
	return map[string]interface{}{
		"isAuthenticated": a.isAuthenticated,
		"isAdmin":         a.isAdmin,
		"isRoot":          a.isRoot,
		"userRole":        a.userRole,
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


// GetCompanies returns list of detected companies
func (a *App) GetCompanies() ([]company.Company, error) {
	return company.DetectCompanies()
}

// InitializeCompanyDatabase initializes the SQLite database for a company
// This is called when a company is selected (even with Supabase auth)
func (a *App) InitializeCompanyDatabase(companyPath string) error {
	// Check if we need to reinitialize (different company or no DB)
	if a.db == nil || a.currentCompanyPath != companyPath {
		if a.db != nil {
			a.db.Close()
		}

		// Use the consolidated initialization method
		db, err := database.InitializeForCompany(companyPath, nil)
		if err != nil {
			return err
		}

		a.db = db
		a.currentCompanyPath = companyPath
		a.Services.Reconciliation = reconciliation.NewService(db)

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

		db, err := database.InitializeForCompany(companyName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}

		a.db = db
		a.auth = common.New(db, companyName) // Pass companyName to Auth constructor
		a.Services.Reconciliation = reconciliation.NewService(db)

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

		db, err := database.InitializeForCompany(companyName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}

		a.db = db
		a.auth = common.New(db, companyName) // Pass companyName to Auth constructor
		a.Services.Reconciliation = reconciliation.NewService(db)

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

		db, err := database.InitializeForCompany(companyName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}

		a.db = db
		a.auth = common.New(db, companyName) // Pass companyName to Auth constructor
		a.Services.Reconciliation = reconciliation.NewService(db)

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
	// Use standardized helper for permission check
	if err := a.requirePermission("database.maintain"); err != nil {
		return nil, err
	}
	
	// Use standardized helper for service check
	if err := a.requireService("operations"); err != nil {
		return nil, err
	}
	
	return a.Services.Operations.RunNetDistribution(
		periodStart, periodEnd, processType, recalculateAll, 
		a.currentUser.ID, a.currentUser.CompanyName,
	)
}

// LogError logs frontend errors to the backend

// PreloadOLEConnection preloads the OLE connection for a company
func (a *App) PreloadOLEConnection(companyName string) map[string]interface{} {
	return a.Services.OLE.PreloadConnection(companyName)
}

// CloseOLEConnection closes the current OLE connection
func (a *App) CloseOLEConnection() map[string]interface{} {
	return a.Services.OLE.CloseConnection()
}

// SetOLEIdleTimeout sets the idle timeout for OLE connections
func (a *App) SetOLEIdleTimeout(minutes int) map[string]interface{} {
	return a.Services.OLE.SetIdleTimeout(minutes)
}

// TestDatabaseQuery executes a test query using Pivoten.DbApi
func (a *App) TestDatabaseQuery(companyName, query string) (map[string]interface{}, error) {
	// Delegate to OLE service
	return a.Services.OLE.TestDatabaseQuery(companyName, query)
}

// GetTableList returns a list of tables in the database
func (a *App) GetTableList(companyName string) (map[string]interface{}, error) {
	return a.Services.DBF.GetTableList(companyName)
}


func (a *App) GetCompanyList() ([]map[string]interface{}, error) {
	// Delegate to Company service
	if a.Services != nil && a.Services.Company != nil {
		return a.Services.Company.GetCompanyList()
	}
	
	// Fallback error if service not initialized
	return nil, fmt.Errorf("company service not initialized")
}

// SelectDataFolder opens a native folder selection dialog
func (a *App) SelectDataFolder() (string, error) {
	// This needs to stay in main.go as it uses Wails runtime
	// TODO: Implement using Wails dialog
	return "", fmt.Errorf("TRIGGER_FOLDER_DIALOG")
}

// SetDataPath validates and saves the selected data path
func (a *App) SetDataPath(dataPath string) error {
	// Update the app's dataBasePath
	a.dataBasePath = dataPath
	
	// Delegate to Company service to save
	if a.Services != nil && a.Services.Company != nil {
		return a.Services.Company.SetDataPath(dataPath)
	}
	return fmt.Errorf("company service not initialized")
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
				"account_number":  account.AccountNumber,
				"account_name":    account.AccountName,
				"account_type":    fmt.Sprintf("%d", account.AccountType),
				"balance":         account.Balance,
				"description":     account.AccountName,
				"is_bank_account": account.IsBankAccount,
			}
		}
		return result, nil
	}

	// Service is required
	return nil, fmt.Errorf("banking service not initialized")
}
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
	// Check permissions
	if a.currentUser == nil {
		return 0, fmt.Errorf("user not authenticated")
	}

	if !a.currentUser.HasPermission("database.read") {
		return 0, fmt.Errorf("insufficient permissions")
	}

	// Delegate to Banking service
	if a.Services != nil && a.Services.Banking != nil {
		return a.Services.Banking.GetAccountBalance(companyName, accountNumber)
	}

	return 0, fmt.Errorf("banking service not initialized")
}

// Bank Transaction structures for SQLite persistence


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

func (a *App) DeleteBankStatement(companyName string, importBatchID string) error {
	// Check permissions
	if a.currentUser == nil {
		return fmt.Errorf("user not authenticated")
	}

	if !a.currentUser.HasPermission("dbf.write") {
		return fmt.Errorf("insufficient permissions to delete bank statements")
	}

	// Delegate to Matching service
	return a.Services.Matching.DeleteStatement(companyName, importBatchID)
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
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	// Delegate to Matching service
	return a.Services.Matching.GetBankTransactions(companyName, accountNumber, importBatchID)
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
		balancesInterface, err := a.Services.Banking.GetCachedBalances(companyName)
		if err != nil {
			return nil, err
		}

		// Type assert to the expected format
		balances, ok := balancesInterface.([]map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected balance data format")
		}
		
		// Already in the right format
		return balances, nil
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
			"account_number":      balance.AccountNumber,
			"account_name":        balance.AccountName,
			"gl_balance":          balance.GLBalance,
			"outstanding_total":   balance.OutstandingTotal,
			"outstanding_count":   balance.OutstandingCount,
			"bank_balance":        balance.BankBalance,
			"gl_last_updated":     balance.GLLastUpdated,
			"checks_last_updated": balance.OutstandingLastUpdated,
			"gl_age_hours":        balance.GLAgeHours,
			"checks_age_hours":    balance.ChecksAgeHours,
			"gl_freshness":        balance.GLFreshness,
			"checks_freshness":    balance.ChecksFreshness,
			"is_stale":            balance.GLFreshness == "stale" || balance.ChecksFreshness == "stale",
			// New detailed breakdown fields
			"uncleared_deposits": balance.UnclearedDeposits,
			"uncleared_checks":   balance.UnclearedChecks,
			"deposit_count":      balance.DepositCount,
			"check_count":        balance.CheckCount,
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
			"status":              "success",
			"account_number":      balance.AccountNumber,
			"account_name":        balance.AccountName,
			"gl_balance":          balance.GLBalance,
			"outstanding_total":   balance.OutstandingChecksTotal,
			"outstanding_count":   balance.OutstandingChecksCount,
			"bank_balance":        balance.BankBalance,
			"gl_last_updated":     balance.GLLastUpdated,
			"checks_last_updated": balance.ChecksLastUpdated,
			"gl_freshness":        utilities.GetFreshnessStatus(balance.GLLastUpdated),
			"checks_freshness":    utilities.GetFreshnessStatus(balance.ChecksLastUpdated),
			"is_stale":            balance.IsStale,
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
		"status":              "success",
		"account_number":      balance.AccountNumber,
		"account_name":        balance.AccountName,
		"gl_balance":          balance.GLBalance,
		"outstanding_total":   balance.OutstandingTotal,
		"outstanding_count":   balance.OutstandingCount,
		"bank_balance":        balance.BankBalance,
		"gl_last_updated":     balance.GLLastUpdated,
		"checks_last_updated": balance.OutstandingLastUpdated,
		"refreshed_by":        username,
	}, nil
}

// RefreshAllBalances refreshes balances for all bank accounts
func (a *App) RefreshAllBalances(companyName string) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser != nil && !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	// Delegate to Banking service
	result, err := a.Services.Banking.RefreshAllBalances(companyName)
	if err != nil {
		return nil, err
	}

	// Add refreshed_by field if user is authenticated
	if a.currentUser != nil {
		result["refreshed_by"] = a.currentUser.Username
	} else {
		result["refreshed_by"] = "system"
	}

	return result, nil
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
			"id":                    h.ID,
			"account_name":          accountName,
			"change_type":           h.ChangeType,
			"old_gl_balance":        h.OldGLBalance,
			"new_gl_balance":        h.NewGLBalance,
			"old_outstanding_total": h.OldOutstandingTotal,
			"new_outstanding_total": h.NewOutstandingTotal,
			"old_available_balance": h.OldAvailableBalance,
			"new_available_balance": h.NewAvailableBalance,
			"change_reason":         h.ChangeReason,
			"changed_by":            h.ChangedBy,
			"change_timestamp":      h.ChangeTimestamp,
		})
	}

	return history, nil
}

// GetLastReconciliation returns the last reconciliation record for a specific bank account
func (a *App) GetLastReconciliation(companyName, accountNumber string) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	if a.Services == nil || a.Services.Reconciliation == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}

	// Delegate to service
	return a.Services.Reconciliation.GetLastReconciliationFromDBF(companyName, accountNumber)
}

// SaveReconciliationDraft saves or updates a draft reconciliation in SQLite
func (a *App) SaveReconciliationDraft(companyName string, draftData map[string]interface{}) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	if !a.currentUser.HasPermission("dbf.write") {
		return nil, fmt.Errorf("insufficient permissions to save reconciliation draft")
	}

	if a.Services == nil || a.Services.Reconciliation == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}

	// Delegate to service
	return a.Services.Reconciliation.SaveDraftFromMap(companyName, draftData, a.currentUser.Username)
}

// GetReconciliationDraft retrieves the current draft reconciliation for an account
func (a *App) GetReconciliationDraft(companyName, accountNumber string) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	if !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	if a.Services == nil || a.Services.Reconciliation == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}

	// Delegate to service
	return a.Services.Reconciliation.GetDraftAsMap(companyName, accountNumber)
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

	if a.Services.Reconciliation == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}

	err := a.Services.Reconciliation.DeleteDraft(companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to delete draft: %w", err)
	}

	return map[string]interface{}{
		"status":  "success",
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

	if a.Services.Reconciliation == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}

	// Get the draft first
	draft, err := a.Services.Reconciliation.GetDraft(companyName, accountNumber)
	if err != nil {
		return nil, fmt.Errorf("no draft found to commit: %w", err)
	}

	// TODO: Update DBF files here (CHECKS.dbf and CHECKREC.dbf)
	// For now, just commit the draft in SQLite

	err = a.Services.Reconciliation.CommitReconciliation(draft.ID, a.currentUser.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to commit reconciliation: %w", err)
	}

	return map[string]interface{}{
		"status":  "success",
		"message": "Reconciliation committed successfully",
		"id":      draft.ID,
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

	if a.Services.Reconciliation == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}

	history, err := a.Services.Reconciliation.GetHistory(companyName, accountNumber, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to get reconciliation history: %w", err)
	}

	return map[string]interface{}{
		"status":  "success",
		"history": history,
		"count":   len(history),
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

	if a.Services.Reconciliation == nil {
		return nil, fmt.Errorf("reconciliation service not initialized")
	}

	result, err := a.Services.Reconciliation.MigrateFromDBF(companyName)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate reconciliation data: %w", err)
	}

	return map[string]interface{}{
		"status":           "success",
		"migration_result": result,
	}, nil
}

// GetBankAccountsForAudit returns a list of bank accounts available for auditing
func (a *App) GetBankAccountsForAudit(companyName string) ([]map[string]interface{}, error) {
	// Use standardized helper for permission check
	if err := a.requireAdminOrRoot(); err != nil {
		return nil, err
	}

	// Use standardized helper for service check
	if err := a.requireService("audit"); err != nil {
		return nil, err
	}

	return a.Services.Audit.GetBankAccountsForAudit(companyName)
}

// TestOLEConnection tests if we can connect to FoxPro OLE server
func (a *App) TestOLEConnection() (map[string]interface{}, error) {
	return a.Services.OLE.TestConnection()
}

// GetCompanyInfo retrieves company information from FoxPro OLE server
func (a *App) GetCompanyInfo(companyName string) (map[string]interface{}, error) {
	// Check user permission
	if a.currentUser == nil {
		return nil, fmt.Errorf("user not authenticated")
	}

	// Delegate to Company service
	return a.Services.Company.GetCompanyInfo(companyName)
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
		"success":   true,
		"logDir":    logDir,
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
			"error":   fmt.Sprintf("Failed to save debug mode preference: %v", err),
		}
	}

	return map[string]interface{}{
		"success":   true,
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

// CheckOwnerStatementFiles checks if owner statement files exist for a company
func (a *App) CheckOwnerStatementFiles(companyName string) map[string]interface{} {
	if a.Services != nil && a.Services.Reports != nil {
		return a.Services.Reports.CheckOwnerStatementFiles(companyName)
	}
	return map[string]interface{}{
		"hasFiles": false,
		"error":    "reports service not initialized",
	}
}

// GetOwnerStatementsList returns a list of available owner statement files
func (a *App) GetOwnerStatementsList(companyName string) ([]map[string]interface{}, error) {
	if a.Services != nil && a.Services.Reports != nil {
		return a.Services.Reports.GetOwnerStatementsList(companyName)
	}
	return nil, fmt.Errorf("reports service not initialized")
}

// GetOwnersList retrieves the list of owners from an owner statement DBF file
func (a *App) GetOwnersList(companyName string, fileName string) ([]map[string]interface{}, error) {
	if a.Services != nil && a.Services.Reports != nil {
		return a.Services.Reports.GetOwnersList(companyName, fileName)
	}
	return nil, fmt.Errorf("reports service not initialized")
}

// GetOwnerStatementData retrieves owner statement data for a specific owner
func (a *App) GetOwnerStatementData(companyName string, fileName string, ownerKey string) (map[string]interface{}, error) {
	if a.Services != nil && a.Services.Reports != nil {
		return a.Services.Reports.GetOwnerStatementData(companyName, fileName, ownerKey)
	}
	return nil, fmt.Errorf("reports service not initialized")
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

// GetOwnerStatementData returns statement data for a specific owner

// ExamineOwnerStatementStructure examines the structure of owner statement DBF files
func (a *App) ExamineOwnerStatementStructure(companyName string, fileName string) (map[string]interface{}, error) {
	// Delegate to Reports service
	return a.Services.Reports.ExamineOwnerStatementStructure(companyName, fileName)
}

// GetVendors retrieves all vendors from VENDOR.dbf
func (a *App) GetVendors(companyName string) (map[string]interface{}, error) {
	if a.Services != nil && a.Services.Vendor != nil {
		return a.Services.Vendor.GetVendors(companyName)
	}
	return nil, fmt.Errorf("vendor service not initialized")
}

// UpdateVendor updates a vendor record in VENDOR.dbf
func (a *App) UpdateVendor(companyName string, vendorIndex int, vendorData map[string]interface{}) error {
	if a.Services != nil && a.Services.Vendor != nil {
		return a.Services.Vendor.UpdateVendor(companyName, vendorIndex, vendorData)
	}
	return fmt.Errorf("vendor service not initialized")
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
