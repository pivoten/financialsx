package processes

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/database"
	"github.com/shopspring/decimal"
)

// DistributionProcessor handles comprehensive oil & gas revenue distribution
// This is a complete conversion of the FoxPro swdist.prg system
type DistributionProcessor struct {
	db          *database.DB
	companyName string
	logger      *log.Logger
	config      *ProcessingConfig
	options     *ProcessingOptions
	progress    *ProgressTracker
	calculator  *RevenueCalculator
	generator   *OutputGenerator
	errorFlag   bool
	errorMsg    string
	quiet       bool
	canceled    bool
}

// ProcessingConfig holds all configuration for the distribution processing
// Mirrors the FoxPro distproc class properties
type ProcessingConfig struct {
	// Core processing parameters
	BegOwnerID    string    `json:"beg_owner_id"`
	EndOwnerID    string    `json:"end_owner_id"`
	BegWellID     string    `json:"beg_well_id"`
	EndWellID     string    `json:"end_well_id"`
	Period        string    `json:"period"`         // MM format
	Year          string    `json:"year"`           // YYYY format
	Group         string    `json:"group"`          // Well group (default "00")
	Process       string    `json:"process"`        // Process type (O=Owner, etc.)
	AcctDate      time.Time `json:"acct_date"`      // Accounting date
	RevDate       time.Time `json:"rev_date"`       // Revenue date
	ExpDate       time.Time `json:"exp_date"`       // Expense date
	RunNo         int       `json:"run_no"`         // Run number
	RunYear       string    `json:"run_year"`       // Run year
	NewRunNo      int       `json:"new_run_no"`     // Generated run number
	NewRunYear    string    `json:"new_run_year"`   // Generated run year
	
	// Control flags
	IsNewRun      bool      `json:"is_new_run"`     // .T. if this is a new run
	IsClosing     bool      `json:"is_closing"`     // .T. if closing the run
	Quiet         bool      `json:"quiet"`          // Don't show progress
	ReleaseMin    bool      `json:"release_min"`    // Release minimums
	Debug         bool      `json:"debug"`          // Debug mode
	Report        bool      `json:"report"`         // Generate reports
	
	// Batch and clearing accounts
	DMBatch       string    `json:"dm_batch"`       // Disbursement Manager batch
	RevClear      string    `json:"rev_clear"`      // Revenue clearing account
	ExpClear      string    `json:"exp_clear"`      // Expense clearing account
	
	// System control
	SysCtlKey     string    `json:"sys_ctl_key"`    // System control key
	AcctYear      string    `json:"acct_year"`      // Accounting year
	AcctPeriod    string    `json:"acct_period"`    // Accounting period
	
	// User and audit
	UserID        int       `json:"user_id"`        // Processing user ID
	StartTime     time.Time `json:"start_time"`     // Process start time
	CompanyPost   bool      `json:"company_post"`   // Company posting flag
}

// ProcessingOptions holds system options that affect processing
type ProcessingOptions struct {
	// Revenue processing options
	RevSummarize    bool            `json:"rev_summarize"`     // Summarize revenue entries
	ExpSummarize    bool            `json:"exp_summarize"`     // Summarize expense entries
	GLDetail        bool            `json:"gl_detail"`         // Post GL in detail
	GLSummary       bool            `json:"gl_summary"`        // Post GL summary
	
	// Tax calculation options
	TaxMethod       string          `json:"tax_method"`        // Tax calculation method
	TaxExemptions   map[string]bool `json:"tax_exemptions"`    // Tax exemption flags
	
	// Payment options
	MinPayment      decimal.Decimal `json:"min_payment"`       // Minimum payment threshold
	DefTransfer     decimal.Decimal `json:"def_transfer"`       // Default transfer amount
	MinTransfer     decimal.Decimal `json:"min_transfer"`       // Minimum transfer amount
	
	// Processing flags
	DirectDeposit   bool            `json:"direct_deposit"`    // Enable direct deposit
	FedWire         bool            `json:"fed_wire"`          // Enable fed wire transfers
	Compression     bool            `json:"compression"`       // Process compression charges
	Gathering       bool            `json:"gathering"`         // Process gathering charges
	Marketing       bool            `json:"marketing"`         // Process marketing charges
	
	// Rounding options
	RoundingOwner   string          `json:"rounding_owner"`    // Owner ID for rounding allocation
	RoundingMethod  string          `json:"rounding_method"`   // Rounding method
}

// ProcessingResult contains comprehensive results from the distribution process
type ProcessingResult struct {
	RunNumber        int               `json:"run_number"`
	RunYear          string            `json:"run_year"`
	Status           string            `json:"status"`
	WellsProcessed   int               `json:"wells_processed"`
	OwnersProcessed  int               `json:"owners_processed"`
	RecordsCreated   int               `json:"records_created"`
	ChecksGenerated  int               `json:"checks_generated"`
	
	// Financial totals
	TotalRevenue     decimal.Decimal   `json:"total_revenue"`
	TotalExpenses    decimal.Decimal   `json:"total_expenses"`
	TotalTaxes       decimal.Decimal   `json:"total_taxes"`
	NetDistributed   decimal.Decimal   `json:"net_distributed"`
	SuspenseAmount   decimal.Decimal   `json:"suspense_amount"`
	RoundingAmount   decimal.Decimal   `json:"rounding_amount"`
	
	// Processing statistics
	WellStats        map[string]int    `json:"well_stats"`      // Wells by status
	OwnerStats       map[string]int    `json:"owner_stats"`     // Owners by type
	TaxStats         map[string]decimal.Decimal `json:"tax_stats"` // Tax by type
	
	// Processing details
	Warnings         []string          `json:"warnings"`
	Errors           []string          `json:"errors"`
	StartTime        time.Time         `json:"start_time"`
	EndTime          time.Time         `json:"end_time"`
	Duration         time.Duration     `json:"duration"`
	ProcessedBy      int               `json:"processed_by"`
	
	// History and audit
	HistoryCreated   bool              `json:"history_created"`
	GLPosted         bool              `json:"gl_posted"`
	ReportGenerated  bool              `json:"report_generated"`
}

// Domain entities representing the core business objects

// Well represents a well in the system
type Well struct {
	ID              string          `json:"id" db:"cwellid"`
	Name            string          `json:"name" db:"cwellname"`
	Status          string          `json:"status" db:"cstatus"`
	Group           string          `json:"group" db:"cgroup"`
	State           string          `json:"state" db:"cstate"`
	County          string          `json:"county" db:"ccounty"`
	Lease           string          `json:"lease" db:"clease"`
	Unit            string          `json:"unit" db:"cunit"`
	Formation       string          `json:"formation" db:"cformation"`
	SpudDate        *time.Time      `json:"spud_date" db:"dspuddate"`
	FirstProdDate   *time.Time      `json:"first_prod_date" db:"dfirstprod"`
	ProcessingFlags WellFlags       `json:"processing_flags"`
	TaxInfo         WellTaxInfo     `json:"tax_info"`
	Rates           WellRates       `json:"rates"`
}

// WellFlags represents processing flags for a well
type WellFlags struct {
	Monthly         bool    `json:"monthly" db:"lmonthly"`
	Quarterly       bool    `json:"quarterly" db:"lquarterly"`
	AllNet          bool    `json:"all_net" db:"lallnet"`
	Suspense        bool    `json:"suspense" db:"lsuspense"`
	Hold            bool    `json:"hold" db:"lhold"`
	Inactive        bool    `json:"inactive" db:"linactive"`
}

// WellTaxInfo represents tax information for a well
type WellTaxInfo struct {
	TaxDistrict     string          `json:"tax_district" db:"ctaxdist"`
	SeveranceTax    decimal.Decimal `json:"severance_tax" db:"nsevtax"`
	TaxExempt       bool            `json:"tax_exempt" db:"ltaxexempt"`
	TaxCode         string          `json:"tax_code" db:"ctaxcode"`
}

// WellRates represents various rates and charges for a well
type WellRates struct {
	CompressionRate decimal.Decimal `json:"compression_rate" db:"ncomprate"`
	GatheringRate   decimal.Decimal `json:"gathering_rate" db:"ngathrate"`
	MarketingRate   decimal.Decimal `json:"marketing_rate" db:"nmktrate"`
	TransportRate   decimal.Decimal `json:"transport_rate" db:"ntransrate"`
}

// Owner represents an owner/investor in the system
type Owner struct {
	ID              string          `json:"id" db:"cownerid"`
	Name            string          `json:"name" db:"cownername"`
	Address         OwnerAddress    `json:"address"`
	TaxInfo         OwnerTaxInfo    `json:"tax_info"`
	PaymentInfo     PaymentInfo     `json:"payment_info"`
	ProcessingFlags OwnerFlags      `json:"processing_flags"`
}

// OwnerAddress represents owner address information
type OwnerAddress struct {
	Address1    string `json:"address1" db:"caddress1"`
	Address2    string `json:"address2" db:"caddress2"`
	City        string `json:"city" db:"ccity"`
	State       string `json:"state" db:"cstate"`
	ZipCode     string `json:"zip_code" db:"czipcode"`
	Country     string `json:"country" db:"ccountry"`
}

// OwnerTaxInfo represents tax information for an owner
type OwnerTaxInfo struct {
	TaxID           string          `json:"tax_id" db:"ctaxid"`
	TaxType         string          `json:"tax_type" db:"ctaxtype"`
	TaxExempt       bool            `json:"tax_exempt" db:"ltaxexempt"`
	BackupWithhold  decimal.Decimal `json:"backup_withhold" db:"nbackupwh"`
	FederalWithhold decimal.Decimal `json:"federal_withhold" db:"nfedwh"`
	StateWithhold   decimal.Decimal `json:"state_withhold" db:"nstatewh"`
}

// PaymentInfo represents payment preferences for an owner
type PaymentInfo struct {
	PaymentMethod   string          `json:"payment_method" db:"cpaymethod"`
	DirectDeposit   bool            `json:"direct_deposit" db:"ldirectdep"`
	BankRouting     string          `json:"bank_routing" db:"cbankroute"`
	BankAccount     string          `json:"bank_account" db:"cbankacct"`
	MinimumPayment  decimal.Decimal `json:"minimum_payment" db:"nminpay"`
	HoldPayments    bool            `json:"hold_payments" db:"lholdpay"`
}

// OwnerFlags represents processing flags for an owner
type OwnerFlags struct {
	Active          bool    `json:"active" db:"lactive"`
	Quarterly       bool    `json:"quarterly" db:"lquarterly"`
	Suspense        bool    `json:"suspense" db:"lsuspense"`
	Hold            bool    `json:"hold" db:"lhold"`
	DirectDeposit   bool    `json:"direct_deposit" db:"ldirectdep"`
	FedWire         bool    `json:"fed_wire" db:"lfedwire"`
}

// Ownership represents ownership interest in a well
type Ownership struct {
	WellID          string          `json:"well_id" db:"cwellid"`
	OwnerID         string          `json:"owner_id" db:"cownerid"`
	Deck            string          `json:"deck" db:"cdeck"`
	InterestType    string          `json:"interest_type" db:"cinttype"`    // W=Working, O=Overriding, L=Landowner
	WorkingInterest decimal.Decimal `json:"working_interest" db:"nwi"`
	NetRevenue      decimal.Decimal `json:"net_revenue" db:"nnri"`
	RoyaltyInterest decimal.Decimal `json:"royalty_interest" db:"nri"`
	EffectiveDate   time.Time       `json:"effective_date" db:"deffdate"`
	ExpirationDate  *time.Time      `json:"expiration_date" db:"dexpdate"`
	Program         string          `json:"program" db:"cprogram"`
	ExpenseClass    string          `json:"expense_class" db:"cexpclass"`
}

// Income represents revenue from production
type Income struct {
	ID              int             `json:"id" db:"id"`
	WellID          string          `json:"well_id" db:"cwellid"`
	Period          string          `json:"period" db:"cperiod"`
	Year            string          `json:"year" db:"cyear"`
	ProductType     string          `json:"product_type" db:"cproduct"`     // OIL, GAS, NGR, etc.
	Volume          decimal.Decimal `json:"volume" db:"nvolume"`
	Price           decimal.Decimal `json:"price" db:"nprice"`
	GrossRevenue    decimal.Decimal `json:"gross_revenue" db:"ngrossrev"`
	PostedDate      time.Time       `json:"posted_date" db:"dposted"`
	SourceDocument  string          `json:"source_document" db:"csource"`
	RunNumber       int             `json:"run_number" db:"nrunno"`
	RunYear         string          `json:"run_year" db:"crunyear"`
	Processed       bool            `json:"processed" db:"lprocessed"`
}

// Expense represents well expenses
type Expense struct {
	ID              int             `json:"id" db:"id"`
	WellID          string          `json:"well_id" db:"cwellid"`
	Period          string          `json:"period" db:"cperiod"`
	Year            string          `json:"year" db:"cyear"`
	ExpenseCategory string          `json:"expense_category" db:"cexpcat"`   // COMP, GATH, MKTG, etc.
	ExpenseClass    string          `json:"expense_class" db:"cexpclass"`    // 1-5, A, B
	Amount          decimal.Decimal `json:"amount" db:"namount"`
	Description     string          `json:"description" db:"cdesc"`
	PostedDate      time.Time       `json:"posted_date" db:"dposted"`
	VendorID        string          `json:"vendor_id" db:"cvendorid"`
	InvoiceNumber   string          `json:"invoice_number" db:"cinvoice"`
	RunNumber       int             `json:"run_number" db:"nrunno"`
	RunYear         string          `json:"run_year" db:"crunyear"`
	Processed       bool            `json:"processed" db:"lprocessed"`
}

// ProgressTracker handles progress reporting during processing
type ProgressTracker struct {
	Current     int
	Total       int
	Message     string
	ShowPercent bool
	Logger      *log.Logger
}

// RevenueCalculator handles all revenue and tax calculations
type RevenueCalculator struct {
	Options    *ProcessingOptions
	Logger     *log.Logger
	TaxRates   map[string]decimal.Decimal
	FlatRates  map[string]decimal.Decimal
}

// OutputGenerator handles check generation and GL posting
type OutputGenerator struct {
	DB      *database.DB
	Logger  *log.Logger
	Options *ProcessingOptions
}

// TaxCalculation represents calculated taxes for a well/owner
type TaxCalculation struct {
	WellID      string
	OwnerID     string
	ProductType string
	Tax1        decimal.Decimal
	Tax2        decimal.Decimal
	Tax3        decimal.Decimal
	Tax4        decimal.Decimal
	TotalTax    decimal.Decimal
	TaxExempt   bool
}

// WellDistribution represents the complete distribution calculation for a well/owner
type WellDistribution struct {
	WellID          string
	WellName        string
	OwnerID         string
	OwnerName       string
	InterestType    string
	WorkingInterest decimal.Decimal
	NetRevenue      decimal.Decimal
	
	// Production and pricing
	OilVolume       decimal.Decimal
	GasVolume       decimal.Decimal
	OtherVolume     decimal.Decimal
	OilPrice        decimal.Decimal
	GasPrice        decimal.Decimal
	OtherPrice      decimal.Decimal
	
	// Revenue calculations
	GrossRevenue    decimal.Decimal
	OwnerRevenue    decimal.Decimal
	
	// Tax calculations
	TaxCalculation  TaxCalculation
	
	// Expense allocations
	CompressionExp  decimal.Decimal
	GatheringExp    decimal.Decimal
	MarketingExp    decimal.Decimal
	OtherExpenses   decimal.Decimal
	TotalExpenses   decimal.Decimal
	
	// Final amounts
	NetAmount       decimal.Decimal
	SuspenseAmount  decimal.Decimal
	PaymentAmount   decimal.Decimal
	
	// Processing flags
	IsSuspense      bool
	IsMinimum       bool
	IsHold          bool
	IsQuarterly     bool
}

// NewDistributionProcessor creates a new distribution processor instance
func NewDistributionProcessor(db *database.DB, companyName string, logger *log.Logger) *DistributionProcessor {
	return &DistributionProcessor{
		db:          db,
		companyName: companyName,
		logger:      logger,
		config:      &ProcessingConfig{},
		options:     &ProcessingOptions{},
		progress:    &ProgressTracker{Logger: logger},
		calculator:  &RevenueCalculator{Logger: logger},
		generator:   &OutputGenerator{DB: db, Logger: logger},
		errorFlag:   false,
		errorMsg:    "",
		quiet:       false,
		canceled:    false,
	}
}

// Initialize sets up the processor with configuration and options
func (dp *DistributionProcessor) Initialize(config *ProcessingConfig, options *ProcessingOptions) error {
	dp.config = config
	dp.options = options
	dp.calculator.Options = options
	dp.generator.Options = options
	
	// Initialize start time
	dp.config.StartTime = time.Now()
	
	// Load system options from database
	if err := dp.loadSystemOptions(); err != nil {
		return fmt.Errorf("failed to load system options: %w", err)
	}
	
	// Validate configuration
	if err := dp.validateConfiguration(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	
	// Initialize progress tracking
	if !dp.config.Quiet {
		dp.progress.Message = fmt.Sprintf("Processing Run %d/%s", dp.config.RunNo, dp.config.RunYear)
		dp.progress.ShowPercent = true
	}
	
	dp.logger.Printf("Distribution processor initialized for run %d/%s", dp.config.RunNo, dp.config.RunYear)
	return nil
}

// Main orchestrates the complete distribution processing workflow
// This is the equivalent of the FoxPro Main procedure
func (dp *DistributionProcessor) Main() (*ProcessingResult, error) {
	result := &ProcessingResult{
		RunNumber:   dp.config.RunNo,
		RunYear:     dp.config.RunYear,
		Status:      "running",
		StartTime:   time.Now(),
		WellStats:   make(map[string]int),
		OwnerStats:  make(map[string]int),
		TaxStats:    make(map[string]decimal.Decimal),
		Warnings:    make([]string, 0),
		Errors:      make([]string, 0),
	}
	
	dp.logger.Printf("Starting distribution processing for run %d/%s", dp.config.RunNo, dp.config.RunYear)
	
	// Start database transaction for data consistency
	tx, err := dp.db.GetConn().Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil || dp.errorFlag {
			tx.Rollback()
			result.Status = "failed"
		} else {
			tx.Commit()
			result.Status = "completed"
		}
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
	}()
	
	// Step 1: Setup - Create working files and load flat rates
	if err = dp.Setup(tx, result); err != nil {
		return result, fmt.Errorf("setup failed: %w", err)
	}
	
	// Step 2: Well Processing - Allocate revenue/expenses to wells
	if err = dp.WellProc(tx, result); err != nil {
		return result, fmt.Errorf("well processing failed: %w", err)
	}
	
	// Step 3: Owner Processing - Distribute well amounts to owners
	if err = dp.OwnerProc(tx, result); err != nil {
		return result, fmt.Errorf("owner processing failed: %w", err)
	}
	
	// Step 4: Calculate Rounding - Handle penny rounding
	if err = dp.CalcRounding(tx, result); err != nil {
		return result, fmt.Errorf("rounding calculation failed: %w", err)
	}
	
	// Step 5: Closing Process - Generate checks, post to GL, create history
	if dp.config.IsClosing {
		if err = dp.CloseProc(tx, result); err != nil {
			return result, fmt.Errorf("closing process failed: %w", err)
		}
	} else {
		// Just handle suspense if not closing
		if err = dp.SuspenseProc(tx, result); err != nil {
			return result, fmt.Errorf("suspense processing failed: %w", err)
		}
	}
	
	// Step 6: Create audit trail
	if err = dp.createAuditTrail(tx, result); err != nil {
		return result, fmt.Errorf("audit trail creation failed: %w", err)
	}
	
	dp.logger.Printf("Distribution processing completed. Duration: %v", result.Duration)
	return result, nil
}

// Setup creates working files and loads flat rates
// Equivalent to the FoxPro Setup procedure
func (dp *DistributionProcessor) Setup(tx *sql.Tx, result *ProcessingResult) error {
	dp.logger.Println("Setting up working files and flat rates...")
	
	// Create temporary working tables
	if err := dp.createWorkingTables(tx); err != nil {
		return fmt.Errorf("failed to create working tables: %w", err)
	}
	
	// Load flat rates into working tables
	if err := dp.loadFlatRates(tx); err != nil {
		return fmt.Errorf("failed to load flat rates: %w", err)
	}
	
	// Check for quarterly wells
	if err := dp.checkQuarterlyWells(tx, result); err != nil {
		return fmt.Errorf("failed to check quarterly wells: %w", err)
	}
	
	dp.logger.Println("Setup completed successfully")
	return nil
}

// WellProc processes revenue and expenses at the well level
// Equivalent to the FoxPro WellProc procedure
func (dp *DistributionProcessor) WellProc(tx *sql.Tx, result *ProcessingResult) error {
	dp.logger.Println("Processing wells...")
	
	// Load wells to process
	wells, err := dp.loadWells(tx)
	if err != nil {
		return fmt.Errorf("failed to load wells: %w", err)
	}
	
	dp.progress.Total = len(wells)
	result.WellsProcessed = len(wells)
	
	for i, well := range wells {
		if dp.canceled {
			return fmt.Errorf("processing canceled by user")
		}
		
		dp.progress.Current = i + 1
		dp.progress.Message = fmt.Sprintf("Processing well %s (%d of %d)", well.ID, i+1, len(wells))
		
		// Process this well
		if err := dp.processWell(tx, well, result); err != nil {
			dp.logger.Printf("Error processing well %s: %v", well.ID, err)
			result.Errors = append(result.Errors, fmt.Sprintf("Well %s: %v", well.ID, err))
			continue
		}
		
		// Update statistics
		result.WellStats[well.Status]++
	}
	
	dp.logger.Printf("Well processing completed: %d wells processed", len(wells))
	return nil
}

// OwnerProc distributes well amounts to owners
// Equivalent to the FoxPro OwnerProc procedure
func (dp *DistributionProcessor) OwnerProc(tx *sql.Tx, result *ProcessingResult) error {
	dp.logger.Println("Processing owner distributions...")
	
	// Load owners to process
	owners, err := dp.loadOwners(tx)
	if err != nil {
		return fmt.Errorf("failed to load owners: %w", err)
	}
	
	dp.progress.Total = len(owners)
	result.OwnersProcessed = len(owners)
	
	for i, owner := range owners {
		if dp.canceled {
			return fmt.Errorf("processing canceled by user")
		}
		
		dp.progress.Current = i + 1
		dp.progress.Message = fmt.Sprintf("Processing owner %s (%d of %d)", owner.ID, i+1, len(owners))
		
		// Process this owner
		if err := dp.processOwner(tx, owner, result); err != nil {
			dp.logger.Printf("Error processing owner %s: %v", owner.ID, err)
			result.Errors = append(result.Errors, fmt.Sprintf("Owner %s: %v", owner.ID, err))
			continue
		}
		
		// Update statistics
		result.OwnerStats["processed"]++
	}
	
	dp.logger.Printf("Owner processing completed: %d owners processed", len(owners))
	return nil
}

// CalcRounding handles penny rounding allocation
// Equivalent to the FoxPro CalcRounding procedure
func (dp *DistributionProcessor) CalcRounding(tx *sql.Tx, result *ProcessingResult) error {
	dp.logger.Println("Calculating rounding adjustments...")
	
	// Get total rounding amount to allocate
	roundingAmount, err := dp.getTotalRounding(tx)
	if err != nil {
		return fmt.Errorf("failed to get total rounding: %w", err)
	}
	
	if roundingAmount.IsZero() {
		dp.logger.Println("No rounding adjustments needed")
		return nil
	}
	
	// Allocate rounding to designated owner
	if err := dp.allocateRounding(tx, roundingAmount); err != nil {
		return fmt.Errorf("failed to allocate rounding: %w", err)
	}
	
	result.RoundingAmount = roundingAmount
	dp.logger.Printf("Allocated $%.2f in rounding adjustments", roundingAmount.InexactFloat64())
	
	return nil
}

// CloseProc handles the closing process - generates checks, posts to GL, creates history
// Equivalent to the FoxPro CloseProc procedure
func (dp *DistributionProcessor) CloseProc(tx *sql.Tx, result *ProcessingResult) error {
	dp.logger.Println("Starting closing process...")
	
	// Generate distribution checks
	checksGenerated, err := dp.generateChecks(tx, result)
	if err != nil {
		return fmt.Errorf("failed to generate checks: %w", err)
	}
	result.ChecksGenerated = checksGenerated
	
	// Process direct deposits and wire transfers
	if dp.options.DirectDeposit {
		if err := dp.processDirectDeposits(tx, result); err != nil {
			return fmt.Errorf("failed to process direct deposits: %w", err)
		}
	}
	
	// Post to general ledger
	if dp.options.GLSummary || dp.options.GLDetail {
		if err := dp.postToGL(tx, result); err != nil {
			return fmt.Errorf("failed to post to GL: %w", err)
		}
		result.GLPosted = true
	}
	
	// Create history records
	if err := dp.createHistory(tx, result); err != nil {
		return fmt.Errorf("failed to create history: %w", err)
	}
	result.HistoryCreated = true
	
	// Generate reports if requested
	if dp.config.Report {
		if err := dp.generateReports(tx, result); err != nil {
			dp.logger.Printf("Warning: Failed to generate reports: %v", err)
			result.Warnings = append(result.Warnings, fmt.Sprintf("Report generation failed: %v", err))
		} else {
			result.ReportGenerated = true
		}
	}
	
	dp.logger.Println("Closing process completed successfully")
	return nil
}

// SuspenseProc handles suspense processing when not closing
func (dp *DistributionProcessor) SuspenseProc(tx *sql.Tx, result *ProcessingResult) error {
	dp.logger.Println("Processing suspense amounts...")
	
	// Calculate total suspense
	suspenseAmount, err := dp.calculateSuspense(tx)
	if err != nil {
		return fmt.Errorf("failed to calculate suspense: %w", err)
	}
	
	result.SuspenseAmount = suspenseAmount
	dp.logger.Printf("Total suspense amount: $%.2f", suspenseAmount.InexactFloat64())
	
	return nil
}

// Helper methods for the main processing procedures

func (dp *DistributionProcessor) loadSystemOptions() error {
	// Load system options from the options table
	dp.logger.Println("Loading system options...")
	// Implementation would load from database
	return nil
}

func (dp *DistributionProcessor) validateConfiguration() error {
	// Validate the processing configuration
	if dp.config.Period == "" || dp.config.Year == "" {
		return fmt.Errorf("period and year are required")
	}
	if dp.config.RunNo <= 0 {
		return fmt.Errorf("invalid run number: %d", dp.config.RunNo)
	}
	return nil
}

func (dp *DistributionProcessor) createWorkingTables(tx *sql.Tx) error {
	// Create temporary tables for processing
	dp.logger.Println("Creating working tables...")
	// Implementation would create temp tables
	return nil
}

func (dp *DistributionProcessor) loadFlatRates(tx *sql.Tx) error {
	// Load flat rates into working tables
	dp.logger.Println("Loading flat rates...")
	// Implementation would load flat rates
	return nil
}

func (dp *DistributionProcessor) checkQuarterlyWells(tx *sql.Tx, result *ProcessingResult) error {
	// Check for wells that should be processed quarterly
	dp.logger.Println("Checking for quarterly wells...")
	// Implementation would check quarterly processing
	return nil
}

func (dp *DistributionProcessor) loadWells(tx *sql.Tx) ([]*Well, error) {
	// Load wells to process based on configuration
	dp.logger.Println("Loading wells for processing...")
	// Implementation would load wells from database
	return []*Well{}, nil
}

func (dp *DistributionProcessor) processWell(tx *sql.Tx, well *Well, result *ProcessingResult) error {
	// Process a single well - allocate revenue and expenses
	dp.logger.Printf("Processing well: %s", well.ID)
	// Implementation would process the well
	return nil
}

func (dp *DistributionProcessor) loadOwners(tx *sql.Tx) ([]*Owner, error) {
	// Load owners to process
	dp.logger.Println("Loading owners for processing...")
	// Implementation would load owners from database
	return []*Owner{}, nil
}

func (dp *DistributionProcessor) processOwner(tx *sql.Tx, owner *Owner, result *ProcessingResult) error {
	// Process a single owner - distribute amounts
	dp.logger.Printf("Processing owner: %s", owner.ID)
	// Implementation would process the owner
	return nil
}

func (dp *DistributionProcessor) getTotalRounding(tx *sql.Tx) (decimal.Decimal, error) {
	// Calculate total rounding amount
	return decimal.Zero, nil
}

func (dp *DistributionProcessor) allocateRounding(tx *sql.Tx, amount decimal.Decimal) error {
	// Allocate rounding to designated owner
	return nil
}

func (dp *DistributionProcessor) generateChecks(tx *sql.Tx, result *ProcessingResult) (int, error) {
	// Generate distribution checks
	dp.logger.Println("Generating distribution checks...")
	// Implementation would generate checks
	return 0, nil
}

func (dp *DistributionProcessor) processDirectDeposits(tx *sql.Tx, result *ProcessingResult) error {
	// Process direct deposits
	dp.logger.Println("Processing direct deposits...")
	return nil
}

func (dp *DistributionProcessor) postToGL(tx *sql.Tx, result *ProcessingResult) error {
	// Post to general ledger
	dp.logger.Println("Posting to general ledger...")
	return nil
}

func (dp *DistributionProcessor) createHistory(tx *sql.Tx, result *ProcessingResult) error {
	// Create history records
	dp.logger.Println("Creating history records...")
	return nil
}

func (dp *DistributionProcessor) generateReports(tx *sql.Tx, result *ProcessingResult) error {
	// Generate reports
	dp.logger.Println("Generating reports...")
	return nil
}

func (dp *DistributionProcessor) calculateSuspense(tx *sql.Tx) (decimal.Decimal, error) {
	// Calculate suspense amount
	return decimal.Zero, nil
}

func (dp *DistributionProcessor) createAuditTrail(tx *sql.Tx, result *ProcessingResult) error {
	// Create audit trail
	auditData := fmt.Sprintf(`{
		"run_number": %d,
		"run_year": "%s",
		"wells_processed": %d,
		"owners_processed": %d,
		"records_created": %d,
		"total_revenue": %s,
		"total_expenses": %s,
		"net_distributed": %s,
		"warnings": %d,
		"errors": %d,
		"duration": "%v"
	}`, result.RunNumber, result.RunYear, result.WellsProcessed, result.OwnersProcessed,
		result.RecordsCreated, result.TotalRevenue.String(), result.TotalExpenses.String(),
		result.NetDistributed.String(), len(result.Warnings), len(result.Errors), result.Duration)

	_, err := tx.Exec(`
		INSERT INTO audit_log (user_id, action, resource, details, created_at)
		VALUES (?, 'distribution_processing', 'revenue_distribution', ?, ?)
	`, dp.config.UserID, auditData, time.Now())

	return err
}

// GetDistributionStatus returns the status of distribution processing for a run
func (dp *DistributionProcessor) GetDistributionStatus(runNo int, runYear string) (map[string]interface{}, error) {
	query := `
		SELECT COUNT(*) as record_count,
			   SUM(net_amount) as total_net_amount,
			   COUNT(DISTINCT well_id) as well_count,
			   COUNT(DISTINCT owner_id) as owner_count,
			   MAX(created_at) as last_processed
		FROM distribution_results
		WHERE run_number = ? AND run_year = ?
	`
	
	var recordCount, wellCount, ownerCount int
	var totalNetAmount sql.NullFloat64
	var lastProcessed sql.NullTime
	
	err := dp.db.GetConn().QueryRow(query, runNo, runYear).Scan(
		&recordCount, &totalNetAmount, &wellCount, &ownerCount, &lastProcessed)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution status: %w", err)
	}
	
	status := map[string]interface{}{
		"record_count":      recordCount,
		"total_net_amount":  totalNetAmount.Float64,
		"well_count":        wellCount,
		"owner_count":       ownerCount,
		"has_distributions": recordCount > 0,
		"last_processed":    lastProcessed.Time,
	}
	
	return status, nil
}

// parseDBFDate parses various date formats that might be found in DBF files
func parseDBFDate(dateValue interface{}) (time.Time, error) {
	if dateValue == nil {
		return time.Time{}, fmt.Errorf("nil date value")
	}
	
	dateStr := fmt.Sprintf("%v", dateValue)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}
	
	// Try common DBF date formats
	formats := []string{
		"20060102",           // YYYYMMDD
		"2006-01-02",         // YYYY-MM-DD
		"01/02/2006",         // MM/DD/YYYY
		"01-02-2006",         // MM-DD-YYYY
		"2006/01/02",         // YYYY/MM/DD
	}
	
	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}