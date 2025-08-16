package app

import (
	"github.com/pivoten/financialsx/desktop/internal/common"
	"github.com/pivoten/financialsx/desktop/internal/company"
	"github.com/pivoten/financialsx/desktop/internal/database"
	"github.com/pivoten/financialsx/desktop/internal/dbf"
	"github.com/pivoten/financialsx/desktop/internal/financials/audit"
	"github.com/pivoten/financialsx/desktop/internal/financials/banking"
	"github.com/pivoten/financialsx/desktop/internal/financials/gl"
	"github.com/pivoten/financialsx/desktop/internal/financials/matching"
	"github.com/pivoten/financialsx/desktop/internal/legacy"
	"github.com/pivoten/financialsx/desktop/internal/logger"
	"github.com/pivoten/financialsx/desktop/internal/ole"
	"github.com/pivoten/financialsx/desktop/internal/operations"
	"github.com/pivoten/financialsx/desktop/internal/reconciliation"
	"github.com/pivoten/financialsx/desktop/internal/reports"
	"github.com/pivoten/financialsx/desktop/internal/vfp"
)

// Services contains all application services
type Services struct {
	// Core services
	Auth         *common.Auth
	*common.I18n // Embedded for direct method access
	Company      *company.Service
	Logger       *logger.Service
	OLE          *ole.Service
	
	// Data access
	DBF *dbf.Service
	
	// Financial services - using pointers to avoid embedding conflicts
	Banking      *banking.Service
	Matching     *matching.Service
	GL           *gl.Service
	Audit        *audit.Service
	
	// Reconciliation
	Reconciliation *reconciliation.Service
	
	// Reports
	Reports *reports.Service
	
	// Operations
	Operations *operations.Service
	
	// Legacy integration
	*legacy.VFPWrapper // Embedded for direct method access
	VFPClient *vfp.VFPClient // Internal use
}

// NewServices creates and initializes all services
func NewServices(dbConn *database.DB) *Services {
	// Get the underlying SQL DB for services that need it
	sqlDB := dbConn.GetDB()
	
	// Initialize VFP client
	vfpClient := vfp.NewVFPClient(sqlDB)
	
	// Create banking service and set database helper
	bankingService := banking.NewService(sqlDB)
	bankingService.SetDatabaseHelper(dbConn)
	
	return &Services{
		// Core services
		Auth:    common.New(dbConn, ""), // Company name will be set later
		I18n:    common.NewI18n("en"),
		Company: company.NewService(),
		Logger:  logger.NewService(),
		OLE:     ole.NewService(),
		
		// Data access
		DBF: dbf.NewService(),
		
		// Financial services
		Banking:  bankingService,
		Matching: matching.NewService(sqlDB),
		GL:       gl.NewService(sqlDB),
		Audit:    audit.NewService(),
		
		// Reconciliation
		Reconciliation: reconciliation.NewService(dbConn),
		
		// Reports
		Reports: reports.NewService(),
		
		// Operations
		Operations: operations.NewService(),
		
		// Legacy integration
		VFPWrapper: legacy.NewVFPWrapper(vfpClient),
		VFPClient:  vfpClient,
	}
}

// InitializeForCompany initializes services for a specific company
func (s *Services) InitializeForCompany(companyPath string) error {
	// Any company-specific initialization
	return nil
}

// Cleanup performs cleanup when switching companies or shutting down
func (s *Services) Cleanup() error {
	// Close connections, save state, etc.
	return nil
}