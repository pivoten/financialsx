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
	"github.com/pivoten/financialsx/desktop/internal/config"
	"github.com/pivoten/financialsx/desktop/internal/database"
	"github.com/pivoten/financialsx/desktop/internal/debug"
	"github.com/pivoten/financialsx/desktop/internal/financials/audit"
	"github.com/pivoten/financialsx/desktop/internal/legacy"
	"github.com/pivoten/financialsx/desktop/internal/logger"
	"github.com/pivoten/financialsx/desktop/internal/ole"
	"github.com/pivoten/financialsx/desktop/internal/reconciliation"
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

// Configuration Management Functions

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

// Bank Transaction structures for SQLite persistence


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

// TestOLEConnection tests if we can connect to FoxPro OLE server
func (a *App) TestOLEConnection() (map[string]interface{}, error) {
	return a.Services.OLE.TestConnection()
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
