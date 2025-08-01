package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/auth"
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/config"
	"github.com/pivoten/financialsx/desktop/internal/database"
	"github.com/pivoten/financialsx/desktop/internal/processes"
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
	
	// Call the real DBF reading function without search
	return company.ReadDBFFile(companyName, fileName, "", 0, 1000, "", "")
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
	
	// Call the DBF reading function with search term
	return company.ReadDBFFile(companyName, fileName, searchTerm, 0, 1000, "", "")
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

	logger := log.New(log.Writer(), fmt.Sprintf("[CLOSING-%s] ", a.currentUser.CompanyName), log.LstdFlags)
	closingProcess := processes.NewClosingProcess(a.db, a.currentUser.CompanyName, logger)

	periodEndDate, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid period end date: %w", err)
	}

	closingDateTime, err := time.Parse("2006-01-02", closingDate)
	if err != nil {
		return nil, fmt.Errorf("invalid closing date: %w", err)
	}

	config := &processes.ClosingConfig{
		PeriodEnd:   periodEndDate,
		ClosingDate: closingDateTime,
		Description: description,
		UserId:      a.currentUser.ID,
		ForceClose:  forceClose,
	}

	result, err := closingProcess.RunClosing(config)
	if err != nil {
		return nil, fmt.Errorf("closing process failed: %w", err)
	}

	// Convert result to map for JSON serialization
	return map[string]interface{}{
		"closing_id":        result.ClosingID,
		"status":           result.Status,
		"records_processed": result.RecordsProcessed,
		"tables_updated":   result.TablesUpdated,
		"warnings":         result.Warnings,
		"errors":           result.Errors,
		"duration":         result.Duration.String(),
		"start_time":       result.StartTime,
		"end_time":         result.EndTime,
	}, nil
}

// GetClosingStatus returns the status of a period
func (a *App) GetClosingStatus(periodEnd string) (string, error) {
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("database.read") {
		return "", fmt.Errorf("insufficient permissions")
	}

	periodEndDate, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return "", fmt.Errorf("invalid period end date: %w", err)
	}

	logger := log.New(log.Writer(), fmt.Sprintf("[CLOSING-%s] ", a.currentUser.CompanyName), log.LstdFlags)
	closingProcess := processes.NewClosingProcess(a.db, a.currentUser.CompanyName, logger)

	return closingProcess.GetClosingStatus(periodEndDate)
}

// ReopenPeriod reopens a closed period
func (a *App) ReopenPeriod(periodEnd, reason string) error {
	// Check permissions - only root/admin can reopen
	if a.currentUser == nil || !a.currentUser.IsAdmin() {
		return fmt.Errorf("insufficient permissions")
	}

	periodEndDate, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return fmt.Errorf("invalid period end date: %w", err)
	}

	logger := log.New(log.Writer(), fmt.Sprintf("[CLOSING-%s] ", a.currentUser.CompanyName), log.LstdFlags)
	closingProcess := processes.NewClosingProcess(a.db, a.currentUser.CompanyName, logger)

	return closingProcess.ReopenPeriod(periodEndDate, a.currentUser.ID, reason)
}

// Net Distribution Functions

// RunNetDistribution executes the net distribution process
func (a *App) RunNetDistribution(periodStart, periodEnd string, processType string, recalculateAll bool) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("database.maintain") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	logger := log.New(log.Writer(), fmt.Sprintf("[NETDIST-%s] ", a.currentUser.CompanyName), log.LstdFlags)
	netDistProcess := processes.NewNetDistributionProcess(a.db, a.currentUser.CompanyName, logger)

	periodStartDate, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, fmt.Errorf("invalid period start date: %w", err)
	}

	periodEndDate, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid period end date: %w", err)
	}

	config := &processes.NetDistributionConfig{
		PeriodStart:    periodStartDate,
		PeriodEnd:      periodEndDate,
		ProcessType:    processType,
		RecalculateAll: recalculateAll,
		UserId:         a.currentUser.ID,
	}

	result, err := netDistProcess.RunDistribution(config)
	if err != nil {
		return nil, fmt.Errorf("net distribution process failed: %w", err)
	}

	// Convert result to map for JSON serialization
	return map[string]interface{}{
		"process_id":       result.ProcessID,
		"status":          result.Status,
		"wells_processed": result.WellsProcessed,
		"owners_processed": result.OwnersProcessed,
		"records_created":  result.RecordsCreated,
		"total_revenue":    result.TotalRevenue,
		"total_expenses":   result.TotalExpenses,
		"net_distributed":  result.NetDistributed,
		"warnings":        result.Warnings,
		"errors":          result.Errors,
		"duration":        result.Duration.String(),
		"start_time":      result.StartTime,
		"end_time":        result.EndTime,
	}, nil
}

// GetNetDistributionStatus returns the status of net distribution for a period
func (a *App) GetNetDistributionStatus(periodStart, periodEnd string) (map[string]interface{}, error) {
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("database.read") {
		return nil, fmt.Errorf("insufficient permissions")
	}

	periodStartDate, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, fmt.Errorf("invalid period start date: %w", err)
	}

	periodEndDate, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid period end date: %w", err)
	}

	logger := log.New(log.Writer(), fmt.Sprintf("[NETDIST-%s] ", a.currentUser.CompanyName), log.LstdFlags)
	netDistProcess := processes.NewNetDistributionProcess(a.db, a.currentUser.CompanyName, logger)

	return netDistProcess.GetDistributionStatus(periodStartDate, periodEndDate)
}

// ExportNetDistribution exports distribution results to DBF format
func (a *App) ExportNetDistribution(periodStart, periodEnd, outputPath string) error {
	// Check permissions
	if a.currentUser == nil || !a.currentUser.HasPermission("database.read") {
		return fmt.Errorf("insufficient permissions")
	}

	periodStartDate, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return fmt.Errorf("invalid period start date: %w", err)
	}

	periodEndDate, err := time.Parse("2006-01-02", periodEnd)
	if err != nil {
		return fmt.Errorf("invalid period end date: %w", err)
	}

	logger := log.New(log.Writer(), fmt.Sprintf("[NETDIST-%s] ", a.currentUser.CompanyName), log.LstdFlags)
	netDistProcess := processes.NewNetDistributionProcess(a.db, a.currentUser.CompanyName, logger)

	return netDistProcess.ExportDistributionToDBF(periodStartDate, periodEndDate, outputPath)
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Pivoten FinancialsX Desktop",
		Width:  1200,
		Height: 900,
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
