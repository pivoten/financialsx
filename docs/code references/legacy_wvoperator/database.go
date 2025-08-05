package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DatabaseManager handles SQLite database operations
type DatabaseManager struct {
	db *sql.DB
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(dbPath string) (*DatabaseManager, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseManager{db: db}, nil
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	return dm.db.Close()
}

// InitializeDatabase creates the database schema
func (dm *DatabaseManager) InitializeDatabase() error {
	// Read and execute the schema SQL
	schemaSQL := `
	PRAGMA foreign_keys = ON;

	-- 1) Operator (top-level company info)
	CREATE TABLE IF NOT EXISTS operator (
		operator_id               TEXT    PRIMARY KEY,
		sw_version_cid_comp       TEXT,
		producer_name             TEXT,
		producer_code             TEXT,
		address                   TEXT,
		city                      TEXT,
		state                     TEXT,
		zip_code                  TEXT,
		dba_attn_agent            TEXT,
		phone                     TEXT,
		email                     TEXT,
		created_at                DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at                DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- 2) Generic lookup table for enums
	CREATE TABLE IF NOT EXISTS lookup (
		domain      TEXT    NOT NULL,   -- 'gas', 'status', 'formation', etc.
		code        TEXT    NOT NULL,
		description TEXT    NOT NULL,
		PRIMARY KEY(domain, code)
	);

	-- 3) Well master table (Schedule 1 & 2)
	CREATE TABLE IF NOT EXISTS well (
		well_id                    TEXT    PRIMARY KEY,
		operator_id                TEXT    NOT NULL REFERENCES operator(operator_id) ON DELETE CASCADE,
		sw_cwell_id                TEXT    NOT NULL UNIQUE,

		-- Schedule 1 fields
		county_name                TEXT,
		county_number              TEXT,
		nra_number                 TEXT,
		api_number                 TEXT,
		well_name                  TEXT,
		land_acreage               REAL,
		lease_acreage              REAL,

		-- Schedule 2 fields
		status_domain              TEXT    NOT NULL DEFAULT 'status',
		status_code                TEXT    NOT NULL,
		formation_domain           TEXT    NOT NULL DEFAULT 'formation',
		formation_code             TEXT    NOT NULL,
		initial_production_date    DATE,

		created_at                 DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at                 DATETIME DEFAULT CURRENT_TIMESTAMP,

		FOREIGN KEY(status_domain, status_code)     REFERENCES lookup(domain, code),
		FOREIGN KEY(formation_domain, formation_code) REFERENCES lookup(domain, code)
	);

	-- 4) Well-gas bridge (many-to-many for gas types)
	CREATE TABLE IF NOT EXISTS well_gas (
		well_id      TEXT NOT NULL REFERENCES well(well_id) ON DELETE CASCADE,
		gas_domain   TEXT NOT NULL DEFAULT 'gas',
		gas_code     TEXT NOT NULL,
		created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(well_id, gas_code),
		FOREIGN KEY(gas_domain, gas_code) REFERENCES lookup(domain, code)
	);

	-- 5) Owner master table
	CREATE TABLE IF NOT EXISTS owner (
		owner_id                   TEXT    PRIMARY KEY,
		sw_cowner_id               TEXT    NOT NULL UNIQUE,
		last_name                  TEXT    NOT NULL,
		first_name                 TEXT,
		address                    TEXT,
		created_at                 DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at                 DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- 6) Accounting financial snapshot (with reporting_period_year)
	CREATE TABLE IF NOT EXISTS well_financial_accounting (
		id                                INTEGER PRIMARY KEY AUTOINCREMENT,
		well_id                           TEXT    NOT NULL REFERENCES well(well_id) ON DELETE CASCADE,
		reporting_period_year             INTEGER NOT NULL,

		production_total_bbl              REAL,
		production_total_mcf              REAL,
		production_total_ngl              REAL,

		revenue_gross_oil                 REAL,
		revenue_gross_gas                 REAL,
		revenue_gross_ngl                 REAL,

		revenue_working_interest_net_oil  REAL,
		revenue_working_interest_net_gas  REAL,
		revenue_working_interest_net_ngl  REAL,

		expenses_working_interest_gross_oil REAL,
		expenses_working_interest_gross_gas REAL,
		expenses_working_interest_gross_ngl REAL,

		revenue_royalty_interest_net_oil  REAL,
		revenue_royalty_interest_net_gas  REAL,
		revenue_royalty_interest_net_ngl  REAL,

		total_revenue_working_interest    REAL,
		total_revenue_royalty_interest    REAL,

		created_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,

		UNIQUE(well_id, reporting_period_year)
	);

	-- 7) Production financial snapshot (with reporting_period_year)
	CREATE TABLE IF NOT EXISTS well_financial_production (
		id                                INTEGER PRIMARY KEY AUTOINCREMENT,
		well_id                           TEXT    NOT NULL REFERENCES well(well_id) ON DELETE CASCADE,
		reporting_period_year             INTEGER NOT NULL,

		production_total_bbl              REAL,
		production_total_mcf              REAL,
		production_total_ngl              REAL,

		revenue_gross_oil                 REAL,
		revenue_gross_gas                 REAL,
		revenue_gross_ngl                 REAL,

		revenue_working_interest_net_oil  REAL,
		revenue_working_interest_net_gas  REAL,
		revenue_working_interest_net_ngl  REAL,

		expenses_working_interest_gross_oil REAL,
		expenses_working_interest_gross_gas REAL,
		expenses_working_interest_gross_ngl REAL,

		revenue_royalty_interest_net_oil  REAL,
		revenue_royalty_interest_net_gas  REAL,
		revenue_royalty_interest_net_ngl  REAL,

		total_revenue_working_interest    REAL,
		total_revenue_royalty_interest    REAL,

		created_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,

		UNIQUE(well_id, reporting_period_year)
	);

	-- 8) Accounting owner interest snapshot (linked to owner)
	CREATE TABLE IF NOT EXISTS well_owner_interest_accounting (
		id                                INTEGER PRIMARY KEY AUTOINCREMENT,
		well_id                           TEXT    NOT NULL REFERENCES well(well_id) ON DELETE CASCADE,
		owner_id                          TEXT    NOT NULL REFERENCES owner(owner_id) ON DELETE CASCADE,
		reporting_period_year             INTEGER NOT NULL,

		decimal_interest                  REAL    NOT NULL,
		income                            REAL    NOT NULL,

		created_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,

		UNIQUE(well_id, owner_id, reporting_period_year)
	);

	-- 9) Production owner interest snapshot (linked to owner)
	CREATE TABLE IF NOT EXISTS well_owner_interest_production (
		id                                INTEGER PRIMARY KEY AUTOINCREMENT,
		well_id                           TEXT    NOT NULL REFERENCES well(well_id) ON DELETE CASCADE,
		owner_id                          TEXT    NOT NULL REFERENCES owner(owner_id) ON DELETE CASCADE,
		reporting_period_year             INTEGER NOT NULL,

		decimal_interest                  REAL    NOT NULL,
		income                            REAL    NOT NULL,

		created_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at                        DATETIME DEFAULT CURRENT_TIMESTAMP,

		UNIQUE(well_id, owner_id, reporting_period_year)
	);

	-- Create indexes
	CREATE INDEX IF NOT EXISTS idx_well_sw_cwell_id    ON well(sw_cwell_id);
	CREATE INDEX IF NOT EXISTS idx_well_status_code     ON well(status_code);
	CREATE INDEX IF NOT EXISTS idx_well_formation_code  ON well(formation_code);
	CREATE INDEX IF NOT EXISTS idx_owner_sw_cowner_id ON owner(sw_cowner_id);
	CREATE INDEX IF NOT EXISTS idx_wfa_well_reporting_period_year ON well_financial_accounting(well_id, reporting_period_year);
	CREATE INDEX IF NOT EXISTS idx_wfp_well_reporting_period_year ON well_financial_production(well_id, reporting_period_year);
	CREATE INDEX IF NOT EXISTS idx_woi_a_well_reporting_period_year ON well_owner_interest_accounting(well_id, reporting_period_year);
	CREATE INDEX IF NOT EXISTS idx_woi_p_well_reporting_period_year ON well_owner_interest_production(well_id, reporting_period_year);
	`

	_, err := dm.db.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to create database schema: %w", err)
	}

	// Seed the lookup table
	if err := dm.seedLookupTable(); err != nil {
		return fmt.Errorf("failed to seed lookup table: %w", err)
	}

	return nil
}

// seedLookupTable populates the lookup table with initial data
func (dm *DatabaseManager) seedLookupTable() error {
	// Check if lookup table already has data
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM lookup").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check lookup table: %w", err)
	}

	if count > 0 {
		// Table already seeded
		return nil
	}

	// Insert seed data
	seedSQL := `
	PRAGMA foreign_keys = OFF;
	BEGIN TRANSACTION;

	-- 1) Gas types
	INSERT INTO lookup(domain, code, description) VALUES
		('gas','ethane',   'Ethane'),
		('gas','propane',  'Propane'),
		('gas','butane',   'Butane'),
		('gas','isobutane','Isobutane'),
		('gas','pentane',  'Pentane');

	-- 2) Well statuses
	INSERT INTO lookup(domain, code, description) VALUES
		('status','A','Active'),
		('status','B','Producing after being shut-in in 2023 (must include date well started producing again)'),
		('status','C','Coal Bed Methane'),
		('status','E','Enhanced'),
		('status','F','Flat Rate (Royalties paid when there was no working interest)'),
		('status','H','Home-Use (ALL Gas goes to a home(s), no oil, gas, or NGLs sold; must include recipient(s))'),
		('status','L','Horizontal Other Than Marcellus'),
		('status','M','Marcellus Vertical'),
		('status','P','Plugged'),
		('status','S','Shut-In'),
		('status','Z','Marcellus Horizontal');

	-- 3) Formations (abbreviated for brevity - you can add the full list from seed.sql)
	INSERT INTO lookup(domain, code, description) VALUES
		('formation','01','Oriskany'),
		('formation','02','Huron, Rhinestreet'),
		('formation','03','Devonian Shale'),
		('formation','04','Huron'),
		('formation','05','Huron, Shales above Huron'),
		('formation','06','Huron, Berea'),
		('formation','07','Berea, Devonian Shale'),
		('formation','08','Berea'),
		('formation','09','Any formation not on list'),
		('formation','10','Unknown'),
		('formation','110','Marcellus');

	COMMIT;
	PRAGMA foreign_keys = ON;
	`

	_, err = dm.db.Exec(seedSQL)
	if err != nil {
		return fmt.Errorf("failed to seed lookup table: %w", err)
	}

	return nil
}

// InsertOperator inserts operator data
func (dm *DatabaseManager) InsertOperator(operatorID string, producer *ProducerInfo) error {
	query := `
	INSERT OR REPLACE INTO operator (
		operator_id, sw_version_cid_comp, producer_name, producer_code,
		address, city, state, zip_code, dba_attn_agent, phone, email,
		updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := dm.db.Exec(query,
		operatorID,
		producer.IDComp,
		producer.CompanyName,
		producer.IDComp,
		producer.Address1,
		producer.City,
		producer.State,
		producer.ZipCode,
		producer.AgentName,
		producer.PhoneNo,
		producer.Email,
		time.Now(),
	)

	return err
}

// InsertWell inserts well data
func (dm *DatabaseManager) InsertWell(well *WellInfo, operatorID string) error {
	query := `
	INSERT OR REPLACE INTO well (
		well_id, operator_id, sw_cwell_id,
		county_name, county_number, nra_number, api_number, well_name,
		land_acreage, lease_acreage,
		status_code, formation_code, initial_production_date,
		updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := dm.db.Exec(query,
		well.WellID,
		operatorID,
		well.WellID,
		well.County,
		well.CountyCode,
		well.NRA,
		well.API,
		well.WellName,
		well.LandAcreage,
		well.LeaseAcreage,
		well.WellStatus,
		well.Formation,
		well.ProductionDate,
		time.Now(),
	)

	return err
}

// ClearFinancialDataForYear clears all financial data for a specific production year
func (dm *DatabaseManager) ClearFinancialDataForYear(productionYear int) error {
	// Clear accounting data
	accountingQuery := `DELETE FROM well_financial_accounting WHERE reporting_period_year = ?`
	_, err := dm.db.Exec(accountingQuery, productionYear)
	if err != nil {
		return fmt.Errorf("failed to clear accounting data: %w", err)
	}

	// Clear production data
	productionQuery := `DELETE FROM well_financial_production WHERE reporting_period_year = ?`
	_, err = dm.db.Exec(productionQuery, productionYear)
	if err != nil {
		return fmt.Errorf("failed to clear production data: %w", err)
	}

	fmt.Printf("✅ Cleared financial data for production year %d\n", productionYear)
	return nil
}

// InsertWellFinancialData inserts financial data for both accounting and production dates
func (dm *DatabaseManager) InsertWellFinancialData(wellID string, productionYear int, accountingWell, productionWell *WellInfo) error {
	// Insert accounting data
	if accountingWell != nil {
		if err := dm.insertFinancialRecord("well_financial_accounting", wellID, productionYear, accountingWell); err != nil {
			return fmt.Errorf("failed to insert accounting data: %w", err)
		}
	}

	// Insert production data
	if productionWell != nil {
		if err := dm.insertFinancialRecord("well_financial_production", wellID, productionYear, productionWell); err != nil {
			return fmt.Errorf("failed to insert production data: %w", err)
		}
	}

	return nil
}

// insertFinancialRecord inserts a single financial record
func (dm *DatabaseManager) insertFinancialRecord(tableName, wellID string, productionYear int, well *WellInfo) error {
	query := fmt.Sprintf(`
	INSERT INTO %s (
		well_id, reporting_period_year,
		production_total_bbl, production_total_mcf, production_total_ngl,
		revenue_gross_oil, revenue_gross_gas, revenue_gross_ngl,
		revenue_working_interest_net_oil, revenue_working_interest_net_gas, revenue_working_interest_net_ngl,
		expenses_working_interest_gross_oil, expenses_working_interest_gross_gas, expenses_working_interest_gross_ngl,
		revenue_royalty_interest_net_oil, revenue_royalty_interest_net_gas, revenue_royalty_interest_net_ngl,
		total_revenue_working_interest, total_revenue_royalty_interest,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tableName)

	_, err := dm.db.Exec(query,
		wellID,
		productionYear,
		well.TotalOilBBL,
		well.TotalGasMCF,
		well.TotalNGLS,
		well.OilRevenue,
		well.GasRevenue,
		well.OtherRevenue,
		well.OilRevenue*(well.WorkingOilInterest/100),
		well.GasRevenue*(well.WorkingGasInterest/100),
		well.OtherRevenue*(well.WorkingOtherInterest/100),
		well.TotalExpenses*(well.WorkingOilInterest/100),
		well.TotalExpenses*(well.WorkingGasInterest/100),
		well.TotalExpenses*(well.WorkingOtherInterest/100),
		well.OilRevenue*(well.RoyaltyOilInterest/100),
		well.GasRevenue*(well.RoyaltyGasInterest/100),
		well.OtherRevenue*(well.RoyaltyOtherInterest/100),
		well.OilRevenue*(well.WorkingOilInterest/100)+well.GasRevenue*(well.WorkingGasInterest/100)+well.OtherRevenue*(well.WorkingOtherInterest/100),
		well.OilRevenue*(well.RoyaltyOilInterest/100)+well.GasRevenue*(well.RoyaltyGasInterest/100)+well.OtherRevenue*(well.RoyaltyOtherInterest/100),
		time.Now(),
		time.Now(),
	)

	return err
}

// GetWellFinancialData retrieves financial data for a well
func (dm *DatabaseManager) GetWellFinancialData(wellID string, reportingYear int, isAccounting bool) (*WellInfo, error) {
	tableName := "well_financial_production"
	if isAccounting {
		tableName = "well_financial_accounting"
	}

	query := fmt.Sprintf(`
	SELECT 
		production_total_bbl, production_total_mcf, production_total_ngl,
		revenue_gross_oil, revenue_gross_gas, revenue_gross_ngl,
		revenue_working_interest_net_oil, revenue_working_interest_net_gas, revenue_working_interest_net_ngl,
		expenses_working_interest_gross_oil, expenses_working_interest_gross_gas, expenses_working_interest_gross_ngl,
		revenue_royalty_interest_net_oil, revenue_royalty_interest_net_gas, revenue_royalty_interest_net_ngl,
		total_revenue_working_interest, total_revenue_royalty_interest
	FROM %s 
	WHERE well_id = ? AND reporting_period_year = ?
	`, tableName)

	well := &WellInfo{WellID: wellID}
	err := dm.db.QueryRow(query, wellID, reportingYear).Scan(
		&well.TotalOilBBL, &well.TotalGasMCF, &well.TotalNGLS,
		&well.OilRevenue, &well.GasRevenue, &well.OtherRevenue,
		&well.WorkingOilInterest, &well.WorkingGasInterest, &well.WorkingOtherInterest,
		&well.TotalExpenses, &well.TotalExpenses, &well.TotalExpenses,
		&well.RoyaltyOilInterest, &well.RoyaltyGasInterest, &well.RoyaltyOtherInterest,
		&well.WorkingInterest, &well.RoyaltyInterest,
	)

	if err != nil {
		return nil, err
	}

	return well, nil
}

// GetAllWellsForYear retrieves all wells for a specific reporting year
func (dm *DatabaseManager) GetAllWellsForYear(reportingYear int) ([]*WellInfo, error) {
	query := `
	SELECT 
		w.well_id, w.well_name, w.county_name, w.county_number, w.nra_number, w.api_number,
		w.land_acreage, w.lease_acreage, w.status_code, w.formation_code, w.initial_production_date,
		wfa.production_total_bbl, wfa.production_total_mcf, wfa.production_total_ngl,
		wfa.revenue_gross_oil, wfa.revenue_gross_gas, wfa.revenue_gross_ngl,
		wfa.total_revenue_working_interest, wfa.total_revenue_royalty_interest
	FROM well w
	LEFT JOIN well_financial_accounting wfa ON w.well_id = wfa.well_id AND wfa.reporting_period_year = ?
	ORDER BY w.well_id
	`

	rows, err := dm.db.Query(query, reportingYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wells []*WellInfo
	for rows.Next() {
		well := &WellInfo{}
		err := rows.Scan(
			&well.WellID, &well.WellName, &well.County, &well.CountyCode, &well.NRA, &well.API,
			&well.LandAcreage, &well.LeaseAcreage, &well.WellStatus, &well.Formation, &well.ProductionDate,
			&well.TotalOilBBL, &well.TotalGasMCF, &well.TotalNGLS,
			&well.OilRevenue, &well.GasRevenue, &well.OtherRevenue,
			&well.WorkingInterest, &well.RoyaltyInterest,
		)
		if err != nil {
			return nil, err
		}
		wells = append(wells, well)
	}

	return wells, nil
}

// GetWellFromDatabase retrieves a single well with financial data from the database
func (dm *DatabaseManager) GetWellFromDatabase(wellID string, productionYear int, useAccountingDate bool) (*WellInfo, error) {
	tableName := "well_financial_production"
	if useAccountingDate {
		tableName = "well_financial_accounting"
	}

	query := fmt.Sprintf(`
	SELECT 
		w.well_id, w.well_name, w.county_name, w.county_number, w.nra_number, w.api_number,
		w.land_acreage, w.lease_acreage, w.status_code, w.formation_code, w.initial_production_date,
		wfa.production_total_bbl, wfa.production_total_mcf, wfa.production_total_ngl,
		wfa.revenue_gross_oil, wfa.revenue_gross_gas, wfa.revenue_gross_ngl,
		wfa.revenue_working_interest_net_oil, wfa.revenue_working_interest_net_gas, wfa.revenue_working_interest_net_ngl,
		wfa.expenses_working_interest_gross_oil, wfa.expenses_working_interest_gross_gas, wfa.expenses_working_interest_gross_ngl,
		wfa.revenue_royalty_interest_net_oil, wfa.revenue_royalty_interest_net_gas, wfa.revenue_royalty_interest_net_ngl,
		wfa.total_revenue_working_interest, wfa.total_revenue_royalty_interest
	FROM well w
	LEFT JOIN %s wfa ON w.well_id = wfa.well_id AND wfa.reporting_period_year = ?
	WHERE w.well_id = ?
	`, tableName)

	well := &WellInfo{}
	err := dm.db.QueryRow(query, productionYear, wellID).Scan(
		&well.WellID, &well.WellName, &well.County, &well.CountyCode, &well.NRA, &well.API,
		&well.LandAcreage, &well.LeaseAcreage, &well.WellStatus, &well.Formation, &well.ProductionDate,
		&well.TotalOilBBL, &well.TotalGasMCF, &well.TotalNGLS,
		&well.OilRevenue, &well.GasRevenue, &well.OtherRevenue,
		&well.WorkingOilInterest, &well.WorkingGasInterest, &well.WorkingOtherInterest,
		&well.TotalExpenses, &well.TotalExpenses, &well.TotalExpenses,
		&well.RoyaltyOilInterest, &well.RoyaltyGasInterest, &well.RoyaltyOtherInterest,
		&well.WorkingInterest, &well.RoyaltyInterest,
	)

	if err != nil {
		return nil, err
	}

	return well, nil
}

// GetProducerFromDatabase retrieves producer information from the database
func (dm *DatabaseManager) GetProducerFromDatabase() (*ProducerInfo, error) {
	query := `
	SELECT operator_id, producer_name, producer_code, address, city, state, zip_code, 
	       dba_attn_agent, phone, email
	FROM operator 
	LIMIT 1
	`

	producer := &ProducerInfo{}
	err := dm.db.QueryRow(query).Scan(
		&producer.IDComp, &producer.CompanyName, &producer.IDComp, &producer.Address1,
		&producer.City, &producer.State, &producer.ZipCode, &producer.AgentName,
		&producer.PhoneNo, &producer.Email,
	)

	if err != nil {
		return nil, err
	}

	return producer, nil
}

// CheckDatabaseExists checks if the database exists and has data
func (dm *DatabaseManager) CheckDatabaseExists() (bool, error) {
	// Check if operator table has data
	var count int
	err := dm.db.QueryRow("SELECT COUNT(*) FROM operator").Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
