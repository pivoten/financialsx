package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DatabaseViewer provides a simple CLI interface to view and edit database data
type DatabaseViewer struct {
	db *sql.DB
}

// NewDatabaseViewer creates a new database viewer
func NewDatabaseViewer(dbPath string) (*DatabaseViewer, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseViewer{db: db}, nil
}

// Close closes the database connection
func (dv *DatabaseViewer) Close() error {
	return dv.db.Close()
}

// ShowMenu displays the main menu
func (dv *DatabaseViewer) ShowMenu() {
	fmt.Println("\n=== WV Operator Database Viewer ===")
	fmt.Println("1. View Operator Information")
	fmt.Println("2. View Wells Summary")
	fmt.Println("3. View Well Details")
	fmt.Println("4. View Financial Data")
	fmt.Println("5. Edit Well Data")
	fmt.Println("6. Export Data")
	fmt.Println("0. Exit")
	fmt.Print("Enter choice: ")
}

// ViewOperatorInfo displays operator information
func (dv *DatabaseViewer) ViewOperatorInfo() error {
	query := `SELECT operator_id, producer_name, producer_code, address, city, state, zip_code, phone, email FROM operator LIMIT 1`

	var operatorID, producerName, producerCode, address, city, state, zipCode, phone, email sql.NullString

	err := dv.db.QueryRow(query).Scan(&operatorID, &producerName, &producerCode, &address, &city, &state, &zipCode, &phone, &email)
	if err != nil {
		return fmt.Errorf("failed to query operator: %w", err)
	}

	fmt.Println("\n=== Operator Information ===")
	fmt.Printf("Operator ID: %s\n", operatorID.String)
	fmt.Printf("Producer Name: %s\n", producerName.String)
	fmt.Printf("Producer Code: %s\n", producerCode.String)
	fmt.Printf("Address: %s\n", address.String)
	fmt.Printf("City: %s, State: %s, Zip: %s\n", city.String, state.String, zipCode.String)
	fmt.Printf("Phone: %s\n", phone.String)
	fmt.Printf("Email: %s\n", email.String)

	return nil
}

// ViewWellsSummary displays a summary of all wells
func (dv *DatabaseViewer) ViewWellsSummary() error {
	query := `
	SELECT 
		w.well_id, w.well_name, w.county_name, w.nra_number, w.api_number,
		w.status_code, w.formation_code,
		COALESCE(wfa.production_total_bbl, 0) as oil_production,
		COALESCE(wfa.production_total_mcf, 0) as gas_production,
		COALESCE(wfa.revenue_gross_oil, 0) as oil_revenue,
		COALESCE(wfa.revenue_gross_gas, 0) as gas_revenue
	FROM well w
	LEFT JOIN well_financial_accounting wfa ON w.well_id = wfa.well_id
	ORDER BY w.well_id
	LIMIT 20
	`

	rows, err := dv.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query wells: %w", err)
	}
	defer rows.Close()

	fmt.Println("\n=== Wells Summary (First 20) ===")
	fmt.Printf("%-15s %-20s %-15s %-12s %-15s %-8s %-12s %-12s %-12s %-12s %-12s\n",
		"Well ID", "Well Name", "County", "NRA", "API", "Status", "Formation", "Oil Prod", "Gas Prod", "Oil Rev", "Gas Rev")
	fmt.Println(strings.Repeat("-", 150))

	for rows.Next() {
		var wellID, wellName, county, nra, api, status, formation sql.NullString
		var oilProd, gasProd, oilRev, gasRev sql.NullFloat64

		err := rows.Scan(&wellID, &wellName, &county, &nra, &api, &status, &formation, &oilProd, &gasProd, &oilRev, &gasRev)
		if err != nil {
			continue
		}

		fmt.Printf("%-15s %-20s %-15s %-12s %-15s %-8s %-12s %-12.0f %-12.0f %-12.0f %-12.0f\n",
			wellID.String, wellName.String, county.String, nra.String, api.String, status.String, formation.String,
			oilProd.Float64, gasProd.Float64, oilRev.Float64, gasRev.Float64)
	}

	return nil
}

// ViewWellDetails displays detailed information for a specific well
func (dv *DatabaseViewer) ViewWellDetails() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Well ID: ")
	wellID, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read well ID: %w", err)
	}
	wellID = strings.TrimSpace(wellID)

	// Query well information
	query := `
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
	LEFT JOIN well_financial_accounting wfa ON w.well_id = wfa.well_id
	WHERE w.well_id = ?
	`

	var wellIDStr, wellName, county, countyCode, nra, api sql.NullString
	var landAcreage, leaseAcreage sql.NullFloat64
	var wellStatus, formation, productionDate sql.NullString
	var totalOilBBL, totalGasMCF, totalNGLS sql.NullFloat64
	var oilRevenue, gasRevenue, otherRevenue sql.NullFloat64
	var workingOilInterest, workingGasInterest, workingOtherInterest sql.NullFloat64
	var totalExpenses1, totalExpenses2, totalExpenses3 sql.NullFloat64
	var royaltyOilInterest, royaltyGasInterest, royaltyOtherInterest sql.NullFloat64
	var workingInterest, royaltyInterest sql.NullFloat64

	err = dv.db.QueryRow(query, wellID).Scan(
		&wellIDStr, &wellName, &county, &countyCode, &nra, &api,
		&landAcreage, &leaseAcreage, &wellStatus, &formation, &productionDate,
		&totalOilBBL, &totalGasMCF, &totalNGLS,
		&oilRevenue, &gasRevenue, &otherRevenue,
		&workingOilInterest, &workingGasInterest, &workingOtherInterest,
		&totalExpenses1, &totalExpenses2, &totalExpenses3,
		&royaltyOilInterest, &royaltyGasInterest, &royaltyOtherInterest,
		&workingInterest, &royaltyInterest,
	)

	if err != nil {
		return fmt.Errorf("failed to query well details: %w", err)
	}

	fmt.Printf("\n=== Well Details: %s ===\n", wellIDStr.String)
	fmt.Printf("Well Name: %s\n", wellName.String)
	fmt.Printf("County: %s (%s)\n", county.String, countyCode.String)
	fmt.Printf("NRA: %s\n", nra.String)
	fmt.Printf("API: %s\n", api.String)
	fmt.Printf("Land Acreage: %.2f\n", landAcreage.Float64)
	fmt.Printf("Lease Acreage: %.2f\n", leaseAcreage.Float64)
	fmt.Printf("Status: %s\n", wellStatus.String)
	fmt.Printf("Formation: %s\n", formation.String)
	fmt.Printf("Initial Production Date: %s\n", productionDate.String)

	fmt.Printf("\n=== Production Data ===\n")
	fmt.Printf("Oil Production (BBL): %.0f\n", totalOilBBL.Float64)
	fmt.Printf("Gas Production (MCF): %.0f\n", totalGasMCF.Float64)
	fmt.Printf("NGL Production: %.0f\n", totalNGLS.Float64)

	fmt.Printf("\n=== Revenue Data ===\n")
	fmt.Printf("Oil Revenue: $%.0f\n", oilRevenue.Float64)
	fmt.Printf("Gas Revenue: $%.0f\n", gasRevenue.Float64)
	fmt.Printf("Other Revenue: $%.0f\n", otherRevenue.Float64)

	fmt.Printf("\n=== Working Interest ===\n")
	fmt.Printf("Oil WI Revenue: $%.0f\n", workingOilInterest.Float64)
	fmt.Printf("Gas WI Revenue: $%.0f\n", workingGasInterest.Float64)
	fmt.Printf("NGL WI Revenue: $%.0f\n", workingOtherInterest.Float64)
	fmt.Printf("Total WI Revenue: $%.0f\n", workingInterest.Float64)

	fmt.Printf("\n=== Royalty Interest ===\n")
	fmt.Printf("Oil RI Revenue: $%.0f\n", royaltyOilInterest.Float64)
	fmt.Printf("Gas RI Revenue: $%.0f\n", royaltyGasInterest.Float64)
	fmt.Printf("NGL RI Revenue: $%.0f\n", royaltyOtherInterest.Float64)
	fmt.Printf("Total RI Revenue: $%.0f\n", royaltyInterest.Float64)

	return nil
}

// EditWellData allows editing well data
func (dv *DatabaseViewer) EditWellData() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Well ID to edit: ")
	wellID, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read well ID: %w", err)
	}
	wellID = strings.TrimSpace(wellID)

	// First, show current data
	fmt.Printf("\nCurrent data for well %s:\n", wellID)
	if err := dv.ViewWellDetails(); err != nil {
		return err
	}

	fmt.Println("\n=== Edit Options ===")
	fmt.Println("1. Edit Production Data")
	fmt.Println("2. Edit Revenue Data")
	fmt.Println("3. Edit Working Interest")
	fmt.Println("4. Edit Royalty Interest")
	fmt.Println("0. Cancel")
	fmt.Print("Enter choice: ")

	choice, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read choice: %w", err)
	}
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return dv.editProductionData(wellID)
	case "2":
		return dv.editRevenueData(wellID)
	case "3":
		return dv.editWorkingInterest(wellID)
	case "4":
		return dv.editRoyaltyInterest(wellID)
	case "0":
		fmt.Println("Edit cancelled.")
		return nil
	default:
		fmt.Println("Invalid choice.")
		return nil
	}
}

// editProductionData allows editing production data
func (dv *DatabaseViewer) editProductionData(wellID string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter new Oil Production (BBL): ")
	oilStr, _ := reader.ReadString('\n')
	oilStr = strings.TrimSpace(oilStr)
	oilProd, err := strconv.ParseFloat(oilStr, 64)
	if err != nil {
		return fmt.Errorf("invalid oil production value: %w", err)
	}

	fmt.Print("Enter new Gas Production (MCF): ")
	gasStr, _ := reader.ReadString('\n')
	gasStr = strings.TrimSpace(gasStr)
	gasProd, err := strconv.ParseFloat(gasStr, 64)
	if err != nil {
		return fmt.Errorf("invalid gas production value: %w", err)
	}

	fmt.Print("Enter new NGL Production: ")
	nglStr, _ := reader.ReadString('\n')
	nglStr = strings.TrimSpace(nglStr)
	nglProd, err := strconv.ParseFloat(nglStr, 64)
	if err != nil {
		return fmt.Errorf("invalid NGL production value: %w", err)
	}

	// Update the database
	query := `
	UPDATE well_financial_accounting 
	SET production_total_bbl = ?, production_total_mcf = ?, production_total_ngl = ?, updated_at = CURRENT_TIMESTAMP
	WHERE well_id = ?
	`

	_, err = dv.db.Exec(query, oilProd, gasProd, nglProd, wellID)
	if err != nil {
		return fmt.Errorf("failed to update production data: %w", err)
	}

	fmt.Println("✅ Production data updated successfully!")
	return nil
}

// editRevenueData allows editing revenue data
func (dv *DatabaseViewer) editRevenueData(wellID string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter new Oil Revenue: ")
	oilStr, _ := reader.ReadString('\n')
	oilStr = strings.TrimSpace(oilStr)
	oilRev, err := strconv.ParseFloat(oilStr, 64)
	if err != nil {
		return fmt.Errorf("invalid oil revenue value: %w", err)
	}

	fmt.Print("Enter new Gas Revenue: ")
	gasStr, _ := reader.ReadString('\n')
	gasStr = strings.TrimSpace(gasStr)
	gasRev, err := strconv.ParseFloat(gasStr, 64)
	if err != nil {
		return fmt.Errorf("invalid gas revenue value: %w", err)
	}

	fmt.Print("Enter new Other Revenue: ")
	otherStr, _ := reader.ReadString('\n')
	otherStr = strings.TrimSpace(otherStr)
	otherRev, err := strconv.ParseFloat(otherStr, 64)
	if err != nil {
		return fmt.Errorf("invalid other revenue value: %w", err)
	}

	// Update the database
	query := `
	UPDATE well_financial_accounting 
	SET revenue_gross_oil = ?, revenue_gross_gas = ?, revenue_gross_ngl = ?, updated_at = CURRENT_TIMESTAMP
	WHERE well_id = ?
	`

	_, err = dv.db.Exec(query, oilRev, gasRev, otherRev, wellID)
	if err != nil {
		return fmt.Errorf("failed to update revenue data: %w", err)
	}

	fmt.Println("✅ Revenue data updated successfully!")
	return nil
}

// editWorkingInterest allows editing working interest data
func (dv *DatabaseViewer) editWorkingInterest(wellID string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter new Oil Working Interest Revenue: ")
	oilStr, _ := reader.ReadString('\n')
	oilStr = strings.TrimSpace(oilStr)
	oilWI, err := strconv.ParseFloat(oilStr, 64)
	if err != nil {
		return fmt.Errorf("invalid oil working interest value: %w", err)
	}

	fmt.Print("Enter new Gas Working Interest Revenue: ")
	gasStr, _ := reader.ReadString('\n')
	gasStr = strings.TrimSpace(gasStr)
	gasWI, err := strconv.ParseFloat(gasStr, 64)
	if err != nil {
		return fmt.Errorf("invalid gas working interest value: %w", err)
	}

	fmt.Print("Enter new NGL Working Interest Revenue: ")
	nglStr, _ := reader.ReadString('\n')
	nglStr = strings.TrimSpace(nglStr)
	nglWI, err := strconv.ParseFloat(nglStr, 64)
	if err != nil {
		return fmt.Errorf("invalid NGL working interest value: %w", err)
	}

	// Update the database
	query := `
	UPDATE well_financial_accounting 
	SET revenue_working_interest_net_oil = ?, revenue_working_interest_net_gas = ?, revenue_working_interest_net_ngl = ?,
		total_revenue_working_interest = ?, updated_at = CURRENT_TIMESTAMP
	WHERE well_id = ?
	`

	totalWI := oilWI + gasWI + nglWI
	_, err = dv.db.Exec(query, oilWI, gasWI, nglWI, totalWI, wellID)
	if err != nil {
		return fmt.Errorf("failed to update working interest data: %w", err)
	}

	fmt.Println("✅ Working interest data updated successfully!")
	return nil
}

// editRoyaltyInterest allows editing royalty interest data
func (dv *DatabaseViewer) editRoyaltyInterest(wellID string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter new Oil Royalty Interest Revenue: ")
	oilStr, _ := reader.ReadString('\n')
	oilStr = strings.TrimSpace(oilStr)
	oilRI, err := strconv.ParseFloat(oilStr, 64)
	if err != nil {
		return fmt.Errorf("invalid oil royalty interest value: %w", err)
	}

	fmt.Print("Enter new Gas Royalty Interest Revenue: ")
	gasStr, _ := reader.ReadString('\n')
	gasStr = strings.TrimSpace(gasStr)
	gasRI, err := strconv.ParseFloat(gasStr, 64)
	if err != nil {
		return fmt.Errorf("invalid gas royalty interest value: %w", err)
	}

	fmt.Print("Enter new NGL Royalty Interest Revenue: ")
	nglStr, _ := reader.ReadString('\n')
	nglStr = strings.TrimSpace(nglStr)
	nglRI, err := strconv.ParseFloat(nglStr, 64)
	if err != nil {
		return fmt.Errorf("invalid NGL royalty interest value: %w", err)
	}

	// Update the database
	query := `
	UPDATE well_financial_accounting 
	SET revenue_royalty_interest_net_oil = ?, revenue_royalty_interest_net_gas = ?, revenue_royalty_interest_net_ngl = ?,
		total_revenue_royalty_interest = ?, updated_at = CURRENT_TIMESTAMP
	WHERE well_id = ?
	`

	totalRI := oilRI + gasRI + nglRI
	_, err = dv.db.Exec(query, oilRI, gasRI, nglRI, totalRI, wellID)
	if err != nil {
		return fmt.Errorf("failed to update royalty interest data: %w", err)
	}

	fmt.Println("✅ Royalty interest data updated successfully!")
	return nil
}

// ExportData exports data to CSV format
func (dv *DatabaseViewer) ExportData() error {
	fmt.Println("Export functionality not yet implemented.")
	fmt.Println("You can use SQLite command line tools to export data:")
	fmt.Println("sqlite3 ../sourcedata/sql/wv_operator.db '.mode csv' '.headers on' 'SELECT * FROM well_financial_accounting' > export.csv")
	return nil
}

// Run starts the database viewer
func (dv *DatabaseViewer) Run() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		dv.ShowMenu()
		choice, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read choice: %w", err)
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "0":
			fmt.Println("Exiting database viewer...")
			return nil
		case "1":
			if err := dv.ViewOperatorInfo(); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		case "2":
			if err := dv.ViewWellsSummary(); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		case "3":
			if err := dv.ViewWellDetails(); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		case "4":
			if err := dv.ViewWellDetails(); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		case "5":
			if err := dv.EditWellData(); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		case "6":
			if err := dv.ExportData(); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		default:
			fmt.Println("Invalid choice. Please try again.")
		}

		fmt.Print("\nPress Enter to continue...")
		reader.ReadString('\n')
	}
}

func main() {
	dbPath := filepath.Join("..", "sourcedata", "sql", "wv_operator.db")

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("❌ Database not found: %s\n", dbPath)
		fmt.Println("Please run the main application with 'Database Operations' first to create the database.")
		os.Exit(1)
	}

	viewer, err := NewDatabaseViewer(dbPath)
	if err != nil {
		fmt.Printf("❌ Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer viewer.Close()

	fmt.Printf("✅ Connected to database: %s\n", dbPath)

	if err := viewer.Run(); err != nil {
		fmt.Printf("❌ Error running viewer: %v\n", err)
		os.Exit(1)
	}
}
