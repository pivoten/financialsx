package app

import (
	"database/sql"
	
	"github.com/pivoten/financialsx/desktop/internal/common"
	"github.com/pivoten/financialsx/desktop/internal/financials/audit"
	"github.com/pivoten/financialsx/desktop/internal/financials/banking"
	"github.com/pivoten/financialsx/desktop/internal/financials/gl"
	"github.com/pivoten/financialsx/desktop/internal/financials/matching"
	"github.com/pivoten/financialsx/desktop/internal/legacy"
	"github.com/pivoten/financialsx/desktop/internal/operations"
	"github.com/pivoten/financialsx/desktop/internal/reconciliation"
	"github.com/pivoten/financialsx/desktop/internal/reports"
	"github.com/pivoten/financialsx/desktop/internal/utilities"
	"github.com/pivoten/financialsx/desktop/internal/vfp"
)

// Services contains all application services
type Services struct {
	// Core services
	Auth         *common.Auth
	*common.I18n // Embedded for direct method access
	
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
	
	// Utilities
	Utilities *utilities.Service
	
	// Legacy integration
	*legacy.VFPWrapper // Embedded for direct method access
	VFPClient *vfp.VFPClient // Internal use
}

// NewServices creates and initializes all services
func NewServices(db *sql.DB) *Services {
	// Initialize VFP client
	vfpClient := vfp.NewVFPClient(db)
	
	return &Services{
		// Core services
		Auth: common.NewAuth(db),
		I18n: common.NewI18n("en"),
		
		// Financial services
		Banking:  banking.NewService(db),
		Matching: matching.NewService(db),
		GL:       gl.NewService(db),
		Audit:    audit.NewService(),
		
		// Reconciliation
		Reconciliation: reconciliation.NewService(db),
		
		// Reports
		Reports: reports.NewService(db),
		
		// Operations
		Operations: operations.NewService(db),
		
		// Utilities
		Utilities: utilities.NewService(db),
		
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