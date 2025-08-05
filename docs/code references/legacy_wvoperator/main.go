/*
WV Producer Information Reader

This program reads producer information from the version table (equivalent to m.goApp variables)
and allows configuration of report parameters for West Virginia annual returns.

Installation Requirements:
go mod init wv-operator-return
go get github.com/Valentin-Kaiser/go-dbase

Required DBF files in sourcedata folder:
- VERSION.DBF
*/

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Valentin-Kaiser/go-dbase/dbase"
)

// ReportType represents the basis for the annual return
type ReportType int
type DateType int

const (
	ProcessData ReportType = iota + 1
	AnalyzeData
	DatabaseOperations
)

const (
	AccountingDate DateType = iota + 1
	ProductionDate
)

func (rt ReportType) String() string {
	switch rt {
	case ProcessData:
		return "Process Data"
	case AnalyzeData:
		return "Analyze Data"
	case DatabaseOperations:
		return "Database Operations"
	default:
		return "Unknown"
	}
}

func (dt DateType) String() string {
	switch dt {
	case AccountingDate:
		return "Accounting Date"
	case ProductionDate:
		return "Production Date"
	default:
		return "Unknown"
	}
}

// ProducerInfo represents company/producer information from version table (m.goapp equivalent)
// Contains only fields needed for the first section of the WV annual return PDF
type ProducerInfo struct {
	IDComp      string // cidcomp - Producer Code
	CompanyName string // cproducer (maps to m.goapp.ccompanyname) - Producer Name
	AgentName   string // cprocessor (maps to m.goapp.cAgentname) - Agent
	Address1    string // caddress1 - Address
	City        string // ccity - City
	State       string // cstate - State
	ZipCode     string // czipcode - Zip
	Contact     string // ccontact - Contact (if different from agent)
	PhoneNo     string // cphoneno - Phone
	Email       string // cemail - Email
}

// WVConfig represents the configuration for WV annual return generation
type WVConfig struct {
	ReportType              ReportType
	DateType                DateType
	ReportingYear           int // 2-digit year for PDF header (e.g., 26)
	ProductionYear          int // 4-digit year for data filtering (e.g., 2024)
	ConsolidateWIToOperator bool
	SourceDataPath          string
}

// WellInfo represents well information for Section 2
type WellInfo struct {
	WellID       string  // cwellid
	WellName     string  // cwellname
	County       string  // ccounty
	CountyCode   string  // derived from county lookup
	NRA          string  // cnra1 (NRA number)
	NRA2         string  // cnra2 (NRA number 2)
	NRA3         string  // cnra3 (NRA number 3)
	NRA4         string  // cnra4 (NRA number 4)
	NRA5         string  // cnra5 (NRA number 5)
	NRA6         string  // cnra6 (NRA number 6)
	API          string  // cpermit1 (API number - first permit)
	Permit2      string  // cpermit2 (second permit)
	Permit3      string  // cpermit3 (third permit)
	Permit4      string  // cpermit4 (fourth permit)
	Permit5      string  // cpermit5 (fifth permit)
	Permit6      string  // cpermit6 (sixth permit)
	LandAcreage  float64 // nacres (Land Book Acreage)
	LeaseAcreage float64 // nacres (same as Land Book Acreage based on FoxPro code)

	// Well Status Fields
	WellStatus string // cwellstat (A=Active, P=Plugged, S=Shut-in, etc.)
	Formation  string // cformation
	FieldID    string // cfieldid

	// Gas Type Flags
	HasEthane    bool // lnglethane
	HasPropane   bool // lnglpropane
	HasButane    bool // lnglbutane
	HasIsobutane bool // lnglisobutane
	HasPentane   bool // lnglpentane
	HasHouseGas  bool // lhousegas

	// Production Data (from wellhist aggregation)
	TotalOilBBL        float64 // SUM(ntotbbl)
	TotalGasMCF        float64 // SUM(ntotmcf)
	TotalNGLS          float64 // SUM(ntotprod) - other production
	ProductionTotalBBL float64 // SUM(ngrossbbl) - total production in barrels
	DaysOn             int     // SUM(ndayson)

	// Revenue Data (from wellhist aggregation)
	OilRevenue   float64 // SUM(ngrossoil)
	GasRevenue   float64 // SUM(ngrossgas)
	OtherRevenue float64 // SUM(nothinc)

	// Expense Data (from wellhist aggregation)
	TotalExpenses float64 // SUM(ntotale + nexpcl1 + nexpcl2 + nexpcl3 + nexpcl4 + nexpcl5)

	// Working Interest and Royalty Data
	WorkingInterest float64 // noilint + ngasint
	RoyaltyInterest float64 // calculated from working interest

	// Interest Groups (from WELLINV table)
	RoyaltyOilInterest      float64 // SUM(NREVOIL) for Royalty group
	RoyaltyGasInterest      float64 // SUM(NREVGAS) for Royalty group
	RoyaltyOtherInterest    float64 // SUM(NREVOTH) for Royalty group
	OverridingOilInterest   float64 // SUM(NREVOIL) for Overriding Royalty group
	OverridingGasInterest   float64 // SUM(NREVGAS) for Overriding Royalty group
	OverridingOtherInterest float64 // SUM(NREVOTH) for Overriding Royalty group
	WorkingOilInterest      float64 // SUM(NREVOIL) for Working Interest group
	WorkingGasInterest      float64 // SUM(NREVGAS) for Working Interest group
	WorkingOtherInterest    float64 // SUM(NREVOTH) for Working Interest group

	// Additional Well Information
	ProductionDate string // dProdDate - Date of first production
}

// Helper function to safely get string from DBF field
func getStringField(value interface{}) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", value))
}

// Helper function to safely get date from DBF field
func getDateField(value interface{}) string {
	if value == nil {
		return ""
	}
	// Convert to string and trim
	dateStr := strings.TrimSpace(fmt.Sprintf("%v", value))
	if dateStr == "" || dateStr == "NULL" {
		return ""
	}
	return dateStr
}

// ReadProducerInfo reads producer information from the version table
func ReadProducerInfo(sourceDataPath string) (*ProducerInfo, error) {
	dbfPath := filepath.Join(sourceDataPath, "VERSION.DBF")

	// Check if file exists
	if _, err := os.Stat(dbfPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("VERSION.DBF file not found in %s - this file is required for producer information", sourceDataPath)
	}

	fmt.Printf("Opening version DBF file: %s\n", dbfPath)

	config := &dbase.Config{Filename: dbfPath}

	table, err := dbase.OpenTable(config)
	if err != nil {
		return nil, fmt.Errorf("failed to open version.dbf: %w", err)
	}
	defer table.Close()

	fmt.Printf("Successfully opened VERSION.DBF file!\n")

	// Read all records (there should typically be only one)
	for !table.EOF() {
		row, err := table.Next()
		if err != nil {
			return nil, fmt.Errorf("error reading version record: %v", err)
		}

		// Skip deleted records
		if row.Deleted {
			continue
		}

		// Get field values using the correct API - only what we need for PDF
		cidcomp, _ := row.ValueByName("CIDCOMP")
		cproducer, _ := row.ValueByName("CPRODUCER")
		cprocessor, _ := row.ValueByName("CPROCESSOR")
		caddress1, _ := row.ValueByName("CADDRESS1")
		ccity, _ := row.ValueByName("CCITY")
		cstate, _ := row.ValueByName("CSTATE")
		czipcode, _ := row.ValueByName("CZIPCODE")
		ccontact, _ := row.ValueByName("CCONTACT")
		cphoneno, _ := row.ValueByName("CPHONENO")
		cemail, _ := row.ValueByName("CEMAIL")

		producer := &ProducerInfo{
			IDComp:      getStringField(cidcomp),
			CompanyName: getStringField(cproducer),
			AgentName:   getStringField(cprocessor),
			Address1:    getStringField(caddress1),
			City:        getStringField(ccity),
			State:       getStringField(cstate),
			ZipCode:     getStringField(czipcode),
			Contact:     getStringField(ccontact),
			PhoneNo:     getStringField(cphoneno),
			Email:       getStringField(cemail),
		}

		return producer, nil
	}

	return nil, fmt.Errorf("no producer information found in version table")
}

// PromptForConfig prompts the user for configuration parameters
func PromptForConfig() (*WVConfig, error) {
	reader := bufio.NewReader(os.Stdin)
	config := &WVConfig{}

	fmt.Println("=== WV Annual Return Configuration ===")
	fmt.Println()

	// Ask for main operation type
	for {
		fmt.Println("Select operation:")
		fmt.Println("1. Process Data")
		fmt.Println("2. Analyze Data")
		fmt.Println("3. Database Operations")
		fmt.Println("0. Exit")
		fmt.Print("Enter choice (0, 1, 2, or 3): ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nExiting due to EOF...")
				os.Exit(0)
			}
			return nil, fmt.Errorf("error reading input: %v", err)
		}

		choice = strings.TrimSpace(choice)
		switch choice {
		case "0":
			fmt.Println("Exiting...")
			os.Exit(0)
		case "1":
			config.ReportType = ProcessData
		case "2":
			config.ReportType = AnalyzeData
		case "3":
			config.ReportType = DatabaseOperations
		default:
			fmt.Println("Invalid choice. Please enter 0, 1, 2, or 3.")
			continue
		}
		break
	}

	// Ask for reporting year (last 2 digits only)
	for {
		fmt.Print("Reporting Year: ")
		yearStr, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nExiting due to EOF...")
				os.Exit(0)
			}
			return nil, fmt.Errorf("error reading input: %v", err)
		}

		yearStr = strings.TrimSpace(yearStr)
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			fmt.Println("Invalid year. Please enter a valid 2-digit year (YY).")
			continue
		}

		// Only accept 2-digit years (00-99)
		if year < 0 || year > 99 {
			fmt.Println("Invalid year. Please enter a valid 2-digit year (YY).")
			continue
		}

		config.ReportingYear = year
		break
	}

	// Ask for production year (4-digit for data filtering)
	for {
		fmt.Print("Production Year: ")
		yearStr, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nExiting due to EOF...")
				os.Exit(0)
			}
			return nil, fmt.Errorf("error reading input: %v", err)
		}

		yearStr = strings.TrimSpace(yearStr)
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			fmt.Println("Invalid year. Please enter a valid 4-digit year (YYYY).")
			continue
		}

		// Validate 4-digit year
		if year < 1900 || year > 2100 {
			fmt.Println("Invalid year. Please enter a valid 4-digit year (YYYY).")
			continue
		}

		config.ProductionYear = year
		break
	}

	// Ask for date type (Accounting or Production) - only for Process Data
	if config.ReportType == ProcessData {
		for {
			fmt.Println("Base data on:")
			fmt.Println("1. Accounting Date")
			fmt.Println("2. Production Date")
			fmt.Print("Enter choice (1 or 2): ")

			choice, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("\nExiting due to EOF...")
					os.Exit(0)
				}
				return nil, fmt.Errorf("error reading input: %v", err)
			}

			choice = strings.TrimSpace(choice)
			switch choice {
			case "1":
				config.DateType = AccountingDate
			case "2":
				config.DateType = ProductionDate
			default:
				fmt.Println("Invalid choice. Please enter 1 or 2.")
				continue
			}
			break
		}
	} else {
		// For Analyze Data, set a default date type (it won't be used since we show both)
		config.DateType = AccountingDate
	}

	// Ask for consolidate WI to Operator (only for Process Data)
	if config.ReportType == ProcessData {
		for {
			fmt.Print("Consolidate Working Interest to Operator? (y/n): ")
			consolidateStr, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("\nExiting due to EOF...")
					os.Exit(0)
				}
				return nil, fmt.Errorf("error reading input: %v", err)
			}

			consolidateStr = strings.ToLower(strings.TrimSpace(consolidateStr))
			switch consolidateStr {
			case "y", "yes", "true", "1":
				config.ConsolidateWIToOperator = true
			case "n", "no", "false", "0":
				config.ConsolidateWIToOperator = false
			default:
				fmt.Println("Invalid choice. Please enter y or n.")
				continue
			}
			break
		}
	}

	config.SourceDataPath = "./sourcedata"

	return config, nil
}

// PrintProducerInfo displays producer information in a formatted way
func PrintProducerInfo(producer *ProducerInfo, config *WVConfig) {
	fmt.Printf("\n===============================================\n")
	fmt.Printf("WEST VIRGINIA ANNUAL RETURN - PRODUCER INFORMATION\n")
	fmt.Printf("===============================================\n\n")

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Operation: %s\n", config.ReportType)
	fmt.Printf("  Reporting Year: %d\n", config.ReportingYear)
	fmt.Printf("  Production Year: %d\n", config.ProductionYear)
	fmt.Printf("  Date Type: %s\n", config.DateType)
	fmt.Printf("  Consolidate WI to Operator: %t\n", config.ConsolidateWIToOperator)
	fmt.Printf("  Source Data Path: %s\n", config.SourceDataPath)

	fmt.Printf("\nProducer Information (for PDF first section):\n")
	fmt.Printf("  Producer Name: %s\n", producer.CompanyName)
	fmt.Printf("  Reporting Year: %d\n", config.ReportingYear)
	fmt.Printf("  Producer Code: %s\n", producer.IDComp)
	fmt.Printf("  Address: %s\n", producer.Address1)
	fmt.Printf("  City: %s\n", producer.City)
	fmt.Printf("  State: %s\n", producer.State)
	fmt.Printf("  Zip: %s\n", producer.ZipCode)

	if producer.AgentName != "" {
		fmt.Printf("  Agent: %s\n", producer.AgentName)
	}
	if producer.PhoneNo != "" {
		fmt.Printf("  Phone: %s\n", producer.PhoneNo)
	}
	if producer.Email != "" {
		fmt.Printf("  Email: %s\n", producer.Email)
	}
	if producer.Contact != "" && producer.Contact != producer.AgentName {
		fmt.Printf("  Contact: %s\n", producer.Contact)
	}

	fmt.Printf("\n===============================================\n")
}

// getCountyCode looks up the county code from the county name
func getCountyCode(countyName string) string {
	// Read the county code file
	data, err := os.ReadFile("seeds/wv_countycode.csv")
	if err != nil {
		// Only print warning once, not for every well
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "CountyNumber") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 2 {
			code := strings.TrimSpace(parts[0])
			county := strings.TrimSpace(parts[1])
			if strings.EqualFold(county, countyName) {
				return code
			}
		}
	}
	return ""
}

// ReadWellsData reads wells data for WV wells
func ReadWellsData(sourceDataPath string, year int, dateType DateType) ([]*WellInfo, error) {
	// Check if county code file exists and warn once if missing
	if _, err := os.Stat("seeds/wv_countycode.csv"); os.IsNotExist(err) {
		fmt.Printf("Warning: County code file not found (seeds/wv_countycode.csv). County codes will be empty.\n")
	}

	dbfPath := filepath.Join(sourceDataPath, "WELLS.DBF")

	config := &dbase.Config{Filename: dbfPath}
	table, err := dbase.OpenTable(config)
	if err != nil {
		return nil, fmt.Errorf("failed to open wells.dbf: %w", err)
	}
	defer table.Close()

	var wells []*WellInfo

	// Read all records
	for !table.EOF() {
		row, err := table.Next()
		if err != nil {
			return nil, fmt.Errorf("error reading wells record: %v", err)
		}

		// Skip deleted records
		if row.Deleted {
			continue
		}

		// Get field values
		cwellid, _ := row.ValueByName("CWELLID")
		cwellname, _ := row.ValueByName("CWELLNAME")
		ccounty, _ := row.ValueByName("CCOUNTY")
		cnra1, _ := row.ValueByName("CNRA1")
		cpermit1, _ := row.ValueByName("CPERMIT1")
		ncstate, _ := row.ValueByName("CSTATE")
		nacres, _ := row.ValueByName("NACRES")

		// Only process WV wells
		state := getStringField(ncstate)
		if state != "WV" {
			continue
		}

		// Get NRA fields
		cnra2, _ := row.ValueByName("CNRA2")
		cnra3, _ := row.ValueByName("CNRA3")
		cnra4, _ := row.ValueByName("CNRA4")
		cnra5, _ := row.ValueByName("CNRA5")
		cnra6, _ := row.ValueByName("CNRA6")

		// Get permit fields
		cpermit2, _ := row.ValueByName("CPERMIT2")
		cpermit3, _ := row.ValueByName("CPERMIT3")
		cpermit4, _ := row.ValueByName("CPERMIT4")
		cpermit5, _ := row.ValueByName("CPERMIT5")
		cpermit6, _ := row.ValueByName("CPERMIT6")

		// Get NRA values - use "MISSING" if empty or null
		nra := getStringField(cnra1)
		if strings.TrimSpace(nra) == "" || nra == "NULL" {
			nra = "MISSING"
		}

		nra2 := getStringField(cnra2)
		if strings.TrimSpace(nra2) == "" || nra2 == "NULL" {
			nra2 = "MISSING"
		}

		nra3 := getStringField(cnra3)
		if strings.TrimSpace(nra3) == "" || nra3 == "NULL" {
			nra3 = "MISSING"
		}

		nra4 := getStringField(cnra4)
		if strings.TrimSpace(nra4) == "" || nra4 == "NULL" {
			nra4 = "MISSING"
		}

		nra5 := getStringField(cnra5)
		if strings.TrimSpace(nra5) == "" || nra5 == "NULL" {
			nra5 = "MISSING"
		}

		nra6 := getStringField(cnra6)
		if strings.TrimSpace(nra6) == "" || nra6 == "NULL" {
			nra6 = "MISSING"
		}

		// Get permit values - use "MISSING" if empty or null
		permit2 := getStringField(cpermit2)
		if strings.TrimSpace(permit2) == "" || permit2 == "NULL" {
			permit2 = "MISSING"
		}

		permit3 := getStringField(cpermit3)
		if strings.TrimSpace(permit3) == "" || permit3 == "NULL" {
			permit3 = "MISSING"
		}

		permit4 := getStringField(cpermit4)
		if strings.TrimSpace(permit4) == "" || permit4 == "NULL" {
			permit4 = "MISSING"
		}

		permit5 := getStringField(cpermit5)
		if strings.TrimSpace(permit5) == "" || permit5 == "NULL" {
			permit5 = "MISSING"
		}

		permit6 := getStringField(cpermit6)
		if strings.TrimSpace(permit6) == "" || permit6 == "NULL" {
			permit6 = "MISSING"
		}

		// Get acreage as float64 - use "MISSING" if 0, null, or empty
		var acreage float64
		var acreageStr string
		if nacres != nil {
			if acreageStr, ok := nacres.(string); ok {
				acreage, _ = strconv.ParseFloat(strings.TrimSpace(acreageStr), 64)
			} else if acreageNum, ok := nacres.(float64); ok {
				acreage = acreageNum
			}
		}

		// Check if acreage should be "MISSING"
		if acreage == 0 || strings.TrimSpace(acreageStr) == "" || acreageStr == "NULL" {
			acreage = -1 // Use -1 to indicate "MISSING" in float64
		}

		// Read additional well fields
		cwellstat, _ := row.ValueByName("CWELLSTAT")
		cformation, _ := row.ValueByName("CFORMATION")
		cfieldid, _ := row.ValueByName("CFIELDID")

		// Read gas type flags
		lnglethane, _ := row.ValueByName("LNGLETHANE")
		lnglpropane, _ := row.ValueByName("LNGLPROPANE")
		lnglbutane, _ := row.ValueByName("LNGLBUTANE")
		lnglisobutane, _ := row.ValueByName("LNGLISOBUTANE")
		lnglpentane, _ := row.ValueByName("LNGLPENTANE")
		lhousegas, _ := row.ValueByName("LHOUSEGAS")

		// Read working interest percentages
		noilint, _ := row.ValueByName("NOILINT")
		ngasint, _ := row.ValueByName("NGASINT")

		// Read production date
		dproddate, _ := row.ValueByName("DPRODDATE")

		// Convert to appropriate types
		wellStatus := getStringField(cwellstat)
		formation := getStringField(cformation)
		fieldID := getStringField(cfieldid)

		// Convert production date
		productionDate := getDateField(dproddate)

		hasEthane := false
		if lnglethane != nil {
			if ethaneBool, ok := lnglethane.(bool); ok {
				hasEthane = ethaneBool
			}
		}

		hasPropane := false
		if lnglpropane != nil {
			if propaneBool, ok := lnglpropane.(bool); ok {
				hasPropane = propaneBool
			}
		}

		hasButane := false
		if lnglbutane != nil {
			if butaneBool, ok := lnglbutane.(bool); ok {
				hasButane = butaneBool
			}
		}

		hasIsobutane := false
		if lnglisobutane != nil {
			if isobutaneBool, ok := lnglisobutane.(bool); ok {
				hasIsobutane = isobutaneBool
			}
		}

		hasPentane := false
		if lnglpentane != nil {
			if pentaneBool, ok := lnglpentane.(bool); ok {
				hasPentane = pentaneBool
			}
		}

		hasHouseGas := false
		if lhousegas != nil {
			if houseGasBool, ok := lhousegas.(bool); ok {
				hasHouseGas = houseGasBool
			}
		}

		// Calculate working interest
		oilInt := 0.0
		if noilint != nil {
			if oilIntNum, ok := noilint.(float64); ok {
				oilInt = oilIntNum
			}
		}

		gasInt := 0.0
		if ngasint != nil {
			if gasIntNum, ok := ngasint.(float64); ok {
				gasInt = gasIntNum
			}
		}

		workingInterest := oilInt + gasInt
		royaltyInterest := 1.0 - workingInterest // Assuming royalty is the remainder

		well := &WellInfo{
			WellID:       getStringField(cwellid),
			WellName:     getStringField(cwellname),
			County:       getStringField(ccounty),
			NRA:          nra,
			NRA2:         nra2,
			NRA3:         nra3,
			NRA4:         nra4,
			NRA5:         nra5,
			NRA6:         nra6,
			API:          getStringField(cpermit1),
			Permit2:      permit2,
			Permit3:      permit3,
			Permit4:      permit4,
			Permit5:      permit5,
			Permit6:      permit6,
			LandAcreage:  acreage,
			LeaseAcreage: acreage, // Same as Land Book Acreage based on FoxPro code

			// Well Status Fields
			WellStatus: wellStatus,
			Formation:  formation,
			FieldID:    fieldID,

			// Gas Type Flags
			HasEthane:    hasEthane,
			HasPropane:   hasPropane,
			HasButane:    hasButane,
			HasIsobutane: hasIsobutane,
			HasPentane:   hasPentane,
			HasHouseGas:  hasHouseGas,

			// Production Data (placeholder - would need wellhist aggregation)
			TotalOilBBL: 0.0,
			TotalGasMCF: 0.0,
			TotalNGLS:   0.0,
			DaysOn:      0,

			// Revenue Data (placeholder - would need wellhist aggregation)
			OilRevenue:   0.0,
			GasRevenue:   0.0,
			OtherRevenue: 0.0,

			// Expense Data (placeholder - would need wellhist aggregation)
			TotalExpenses: 0.0,

			// Working Interest and Royalty Data
			WorkingInterest: workingInterest,
			RoyaltyInterest: royaltyInterest,

			// Interest Groups (from WELLINV table)
			RoyaltyOilInterest:      0.0,
			RoyaltyGasInterest:      0.0,
			RoyaltyOtherInterest:    0.0,
			OverridingOilInterest:   0.0,
			OverridingGasInterest:   0.0,
			OverridingOtherInterest: 0.0,
			WorkingOilInterest:      0.0,
			WorkingGasInterest:      0.0,
			WorkingOtherInterest:    0.0,

			// Additional Well Information
			ProductionDate: productionDate,
		}

		// Look up county code
		well.CountyCode = getCountyCode(well.County)

		wells = append(wells, well)
	}

	fmt.Printf("Found %d WV wells\n", len(wells))

	// Now read and aggregate wellhist data for each well
	fmt.Printf("Reading wellhist data for production and revenue aggregation...\n")
	err = aggregateWellhistData(wells, sourceDataPath, year, dateType)
	if err != nil {
		fmt.Printf("Warning: Could not read wellhist data: %v\n", err)
	}

	// Now read and aggregate wellinv data for interest calculations
	fmt.Printf("Reading wellinv data for interest calculations...\n")
	err = aggregateWellinvData(wells, sourceDataPath, year, dateType)
	if err != nil {
		fmt.Printf("Warning: Could not read wellinv data: %v\n", err)
	}

	return wells, nil
}

// aggregateWellhistData reads wellhist.DBF and aggregates production/revenue data for each well
func aggregateWellhistData(wells []*WellInfo, sourceDataPath string, year int, dateType DateType) error {
	dbfPath := filepath.Join(sourceDataPath, "WELLHIST.DBF")

	config := &dbase.Config{Filename: dbfPath}
	table, err := dbase.OpenTable(config)
	if err != nil {
		return fmt.Errorf("failed to open wellhist.dbf: %w", err)
	}
	defer table.Close()

	// Create a map for quick well lookup
	wellMap := make(map[string]*WellInfo)
	for _, well := range wells {
		wellMap[well.WellID] = well
	}

	fmt.Printf("Reading wellhist records...\n")
	recordCount := 0
	matchedCount := 0

	// Read all records and aggregate by well
	for !table.EOF() {
		row, err := table.Next()
		if err != nil {
			return fmt.Errorf("error reading wellhist record: %v", err)
		}

		recordCount++

		// Skip deleted records
		if row.Deleted {
			continue
		}

		// Get well ID
		cwellid, _ := row.ValueByName("CWELLID")
		wellID := getStringField(cwellid)

		// Check if this well is in our list
		well, exists := wellMap[wellID]
		if !exists {
			continue // Skip wells not in our WV wells list
		}

		// Apply year filtering based on date type
		var recordYear int
		var shouldInclude bool

		if dateType == AccountingDate {
			// For accounting date, filter on HDATE - check if it falls in the year
			hdate, _ := row.ValueByName("HDATE")
			recordDate := getDateField(hdate)

			if recordDate != "" {
				// Parse the date and check if it's in the specified year
				// Handle different date formats: MM/DD/YYYY, YYYY-MM-DD, etc.
				if len(recordDate) >= 4 {
					// Try to extract year from the date string
					var dateYear string
					if strings.Contains(recordDate, "/") {
						// Format: MM/DD/YYYY
						parts := strings.Split(recordDate, "/")
						if len(parts) >= 3 {
							dateYear = parts[2]
						}
					} else if strings.Contains(recordDate, "-") {
						// Format: YYYY-MM-DD
						parts := strings.Split(recordDate, "-")
						if len(parts) >= 1 {
							dateYear = parts[0]
						}
					} else if len(recordDate) >= 4 {
						// Format: YYYYMMDD or similar
						dateYear = recordDate[:4]
					}

					if dateYear != "" {
						// Compare with the production year directly
						targetYear := fmt.Sprintf("%d", year)
						shouldInclude = dateYear == targetYear
					} else {
						shouldInclude = false
					}
				} else {
					shouldInclude = false
				}
			} else {
				shouldInclude = false
			}
		} else {
			// For production date, filter on hyear
			hyear, _ := row.ValueByName("HYEAR")
			recordYear = int(getFloatField(hyear))
			shouldInclude = recordYear == year
		}

		if !shouldInclude {
			continue // Skip records not matching the year filter
		}

		matchedCount++

		// Get production data
		ntotbbl, _ := row.ValueByName("NTOTBBL")
		ntotmcf, _ := row.ValueByName("NTOTMCF")
		ntotprod, _ := row.ValueByName("NTOTPROD")
		ngrossbbl, _ := row.ValueByName("NGROSSBBL")
		ndayson, _ := row.ValueByName("NDAYSON")

		// Get revenue data
		ngrossoil, _ := row.ValueByName("NGROSSOIL")
		ngrossgas, _ := row.ValueByName("NGROSSGAS")
		nothinc, _ := row.ValueByName("NOTHINC")

		// Get expense data
		ntotale, _ := row.ValueByName("NTOTALE")
		nexpcl1, _ := row.ValueByName("NEXPCL1")
		nexpcl2, _ := row.ValueByName("NEXPCL2")
		nexpcl3, _ := row.ValueByName("NEXPCL3")
		nexpcl4, _ := row.ValueByName("NEXPCL4")
		nexpcl5, _ := row.ValueByName("NEXPCL5")

		// Convert to float64 and add to well totals
		well.TotalOilBBL += getFloatField(ntotbbl)
		well.TotalGasMCF += getFloatField(ntotmcf)
		well.TotalNGLS += getFloatField(ntotprod)
		well.ProductionTotalBBL += getFloatField(ngrossbbl)
		well.DaysOn += int(getFloatField(ndayson))

		well.OilRevenue += getFloatField(ngrossoil)
		well.GasRevenue += getFloatField(ngrossgas)
		well.OtherRevenue += getFloatField(nothinc)

		well.TotalExpenses += getFloatField(ntotale) +
			getFloatField(nexpcl1) +
			getFloatField(nexpcl2) +
			getFloatField(nexpcl3) +
			getFloatField(nexpcl4) +
			getFloatField(nexpcl5)
	}

	fmt.Printf("Wellhist aggregation complete: %d total records, %d matched WV wells\n", recordCount, matchedCount)

	return nil
}

// aggregateWellinvData aggregates interest data from WELLINV table for each well
func aggregateWellinvData(wells []*WellInfo, sourceDataPath string, year int, dateType DateType) error {
	dbfPath := filepath.Join(sourceDataPath, "WELLINV.DBF")

	// Check if file exists
	if _, err := os.Stat(dbfPath); os.IsNotExist(err) {
		fmt.Printf("Warning: WELLINV.DBF file not found in %s - interest calculations will be zero\n", sourceDataPath)
		return nil
	}

	fmt.Printf("Opening WELLINV DBF file: %s\n", dbfPath)

	config := &dbase.Config{Filename: dbfPath}

	table, err := dbase.OpenTable(config)
	if err != nil {
		return fmt.Errorf("failed to open WELLINV.dbf: %w", err)
	}
	defer table.Close()

	fmt.Printf("Successfully opened WELLINV.DBF file!\n")

	// Create a map for quick well lookup
	wellMap := make(map[string]*WellInfo)
	for _, well := range wells {
		wellMap[well.WellID] = well
	}

	recordCount := 0
	matchedCount := 0

	// Read all records
	for !table.EOF() {
		row, err := table.Next()
		if err != nil {
			return fmt.Errorf("error reading WELLINV record: %v", err)
		}

		recordCount++

		// Skip deleted records
		if row.Deleted {
			continue
		}

		// Get field values
		cwellid, _ := row.ValueByName("CWELLID")
		ctypeinv, _ := row.ValueByName("CTYPEINV") // Interest type (L=Royalty, O=Overriding, W=Working)
		nrevoil, _ := row.ValueByName("NREVOIL")
		nrevgas, _ := row.ValueByName("NREVGAS")
		nrevoth, _ := row.ValueByName("NREVOTH")

		wellID := getStringField(cwellid)
		interestType := getStringField(ctypeinv)

		// Check if this well is in our list
		well, exists := wellMap[wellID]
		if !exists {
			continue
		}

		matchedCount++

		// Convert interest values to float64
		oilInterest := getFloatField(nrevoil)
		gasInterest := getFloatField(nrevgas)
		otherInterest := getFloatField(nrevoth)

		// Add to appropriate interest group based on interest type
		switch strings.ToUpper(interestType) {
		case "L": // Royalty
			well.RoyaltyOilInterest += oilInterest
			well.RoyaltyGasInterest += gasInterest
			well.RoyaltyOtherInterest += otherInterest
		case "O": // Overriding Royalty
			well.OverridingOilInterest += oilInterest
			well.OverridingGasInterest += gasInterest
			well.OverridingOtherInterest += otherInterest
		case "W": // Working Interest
			well.WorkingOilInterest += oilInterest
			well.WorkingGasInterest += gasInterest
			well.WorkingOtherInterest += otherInterest
		default:
			// Unknown interest type - skip
		}
	}

	fmt.Printf("WELLINV aggregation complete: %d total records, %d matched WV wells\n", recordCount, matchedCount)

	return nil
}

// getFloatField converts interface{} to float64, returning 0.0 if conversion fails
func getFloatField(value interface{}) float64 {
	if value == nil {
		return 0.0
	}

	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f
		}
		return 0.0
	default:
		return 0.0
	}
}

// roundToWholeNumber rounds a float64 to the nearest whole number
func roundToWholeNumber(value float64) float64 {
	return math.Round(value)
}

// PromptForDataAnalysisMenu prompts the user for data analysis options
func PromptForDataAnalysisMenu() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== Data Analysis Menu ===")
	fmt.Println("1. Analyze NRA Numbers (find wells with multiple NRA values)")
	fmt.Println("2. Analyze Well Status Distribution")
	fmt.Println("3. Analyze County Distribution")
	fmt.Println("4. Analyze Formation Distribution")
	fmt.Println("5. Analyze Gas Type Distribution")
	fmt.Println("6. Analyze Permit Numbers (find wells with multiple permits)")
	fmt.Println("7. Analyze Single Well Revenue (compare wellhist, disbhist, ownerhist)")
	fmt.Println("8. Display Single Well PDF Fields (all fields for one well)")
	fmt.Println("0. Back to Main Menu")
	fmt.Print("Enter choice (0-8): ")

	choice, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Println("\nExiting due to EOF...")
			os.Exit(0)
		}
		fmt.Printf("Error reading input: %v\n", err)
		return "0" // Return "0" to go back to main menu on error
	}

	choice = strings.TrimSpace(choice)
	return choice
}

// DisplaySingleWellPDFFields shows all PDF fields for a single well in text format
func DisplaySingleWellPDFFields(wells []*WellInfo, producer *ProducerInfo) {
	fmt.Printf("\n===============================================\n")
	fmt.Printf("SINGLE WELL PDF FIELDS ANALYSIS\n")
	fmt.Printf("===============================================\n\n")

	// Show list of wells to choose from
	fmt.Printf("Available wells (showing first 10):\n")
	for i, well := range wells {
		if i >= 10 {
			break
		}
		fmt.Printf("%d. %s (%s) - %s\n", i+1, well.WellName, well.WellID, well.County)
	}

	// Get user selection
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nEnter well number (1-10), well ID, or 0 to go back: ")
	choice, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Println("\nExiting due to EOF...")
			return
		}
		fmt.Printf("Error reading input: %v\n", err)
		return
	}

	choice = strings.TrimSpace(choice)
	if choice == "0" {
		return
	}

	// Try to parse as a number first (1-10 selection)
	wellIndex, err := strconv.Atoi(choice)
	if err == nil && wellIndex >= 1 && wellIndex <= 10 && wellIndex <= len(wells) {
		// Valid number selection
		well := wells[wellIndex-1]
		displayWellDetails(well, producer)
		return
	}

	// If not a valid number, try to find by well ID
	fmt.Printf("Searching for well ID: %s\n", choice)
	var selectedWell *WellInfo
	for _, well := range wells {
		if well.WellID == choice {
			selectedWell = well
			break
		}
	}

	if selectedWell == nil {
		fmt.Printf("Well ID '%s' not found. Please enter a valid well number (1-10) or well ID.\n", choice)
		return
	}

	displayWellDetails(selectedWell, producer)
}

// displayWellDetails shows all PDF fields for a specific well
func displayWellDetails(well *WellInfo, producer *ProducerInfo) {
	// Get the source data path from the current working directory
	sourceDataPath := "sourcedata"

	// Try to get data from database first, fall back to DBF files
	var accountingData, productionData *WellInfo
	var accountingRecords, productionRecords []map[string]interface{}
	var err error

	// Check if database exists and has data
	dbPath := filepath.Join("sourcedata", "sql", "wv_operator.db")
	if _, err := os.Stat(dbPath); err == nil {
		dbManager, err := NewDatabaseManager(dbPath)
		if err == nil {
			defer dbManager.Close()

			hasData, err := dbManager.CheckDatabaseExists()
			if err == nil && hasData {
				// Try to get well from database
				accountingData, err = dbManager.GetWellFromDatabase(well.WellID, 2024, true)
				if err == nil {
					fmt.Println("✅ Using database data for accounting dates")
				}

				productionData, err = dbManager.GetWellFromDatabase(well.WellID, 2024, false)
				if err == nil {
					fmt.Println("✅ Using database data for production dates")
				}
			}
		}
	}

	// Fall back to DBF files if database data not available
	if accountingData == nil {
		accountingData, accountingRecords, err = readSingleWellWellhistDataWithRecords(well.WellID, sourceDataPath, 2024, AccountingDate)
		if err != nil {
			fmt.Printf("Warning: Could not read accounting date data: %v\n", err)
			accountingData = well // Use the original well data as fallback
			accountingRecords = []map[string]interface{}{}
		}
	}

	if productionData == nil {
		productionData, productionRecords, err = readSingleWellWellhistDataWithRecords(well.WellID, sourceDataPath, 2024, ProductionDate)
		if err != nil {
			fmt.Printf("Warning: Could not read production date data: %v\n", err)
			productionData = well // Use the original well data as fallback
			productionRecords = []map[string]interface{}{}
		}
	}
	fmt.Printf("\n===============================================\n")
	fmt.Printf("PDF FIELDS FOR WELL: %s (%s)\n", well.WellName, well.WellID)
	fmt.Printf("===============================================\n\n")

	// Section 1 - Producer Information
	fmt.Printf("SECTION 1 - PRODUCER INFORMATION\n")
	fmt.Printf("================================\n")
	fmt.Printf("Producer Name:     %s\n", producer.CompanyName)
	fmt.Printf("Producer Code:     %s\n", producer.IDComp)
	fmt.Printf("Address:           %s\n", producer.Address1)
	fmt.Printf("City:              %s\n", producer.City)
	fmt.Printf("State:             %s\n", producer.State)
	fmt.Printf("Zip Code:          %s\n", producer.ZipCode)
	fmt.Printf("Phone:             %s\n", producer.PhoneNo)
	fmt.Printf("Email:             %s\n", producer.Email)
	fmt.Printf("Agent:             %s\n", producer.AgentName)
	fmt.Printf("\n")

	// Schedule 1 - Well Information
	fmt.Printf("SCHEDULE 1 - WELL INFORMATION\n")
	fmt.Printf("=============================\n")
	fmt.Printf("County Name:       %s\n", well.County)
	fmt.Printf("County Number:     %s\n", well.CountyCode)
	fmt.Printf("NRA Number:        %s\n", well.NRA)
	// Format acreage for display
	landAcreageStr := "MISSING"
	leaseAcreageStr := "MISSING"
	if well.LandAcreage > 0 {
		landAcreageStr = fmt.Sprintf("%.2f", well.LandAcreage)
		leaseAcreageStr = fmt.Sprintf("%.2f", well.LeaseAcreage)
	}

	fmt.Printf("API Number:        %s\n", well.API)
	fmt.Printf("Well Name:         %s\n", well.WellName)
	fmt.Printf("Land Acreage:      %s\n", landAcreageStr)
	fmt.Printf("Lease Acreage:     %s\n", leaseAcreageStr)
	fmt.Printf("\n")

	// Well Status Checkboxes
	fmt.Printf("WELL STATUS CHECKBOXES\n")
	fmt.Printf("======================\n")
	fmt.Printf("Status Active (A):     %s\n", getStatusValue(well.WellStatus, "A"))
	fmt.Printf("Status Plugged (P):    %s\n", getStatusValue(well.WellStatus, "P"))
	fmt.Printf("Status Shut-in (S):    %s\n", getStatusValue(well.WellStatus, "S"))
	fmt.Printf("Raw Status Value:      '%s'\n", well.WellStatus)
	fmt.Printf("\n")

	// Gas Type Checkboxes
	fmt.Printf("GAS TYPE CHECKBOXES\n")
	fmt.Printf("===================\n")
	fmt.Printf("Ethane:           %s\n", getBoolValue(well.HasEthane))
	fmt.Printf("Propane:          %s\n", getBoolValue(well.HasPropane))
	fmt.Printf("Butane:           %s\n", getBoolValue(well.HasButane))
	fmt.Printf("Isobutane:        %s\n", getBoolValue(well.HasIsobutane))
	fmt.Printf("Pentane:          %s\n", getBoolValue(well.HasPentane))
	fmt.Printf("\n")

	// Production Information
	fmt.Printf("PRODUCTION INFORMATION\n")
	fmt.Printf("======================\n")
	fmt.Printf("Initial Production Date: %s\n", well.ProductionDate)
	fmt.Printf("Formations:              %s\n", well.Formation)
	fmt.Printf("\n")

	// Production Totals - Compare Accounting vs Production Date
	fmt.Printf("PRODUCTION TOTALS COMPARISON\n")
	fmt.Printf("============================\n")
	fmt.Printf("                    Accounting Date    Production Date\n")
	fmt.Printf("                    ----------------    ---------------\n")
	fmt.Printf("Total Oil BBL:     %12.0f         %12.0f\n", roundToWholeNumber(accountingData.TotalOilBBL), roundToWholeNumber(productionData.TotalOilBBL))
	fmt.Printf("Total Gas MCF:     %12.0f         %12.0f\n", roundToWholeNumber(accountingData.TotalGasMCF), roundToWholeNumber(productionData.TotalGasMCF))
	fmt.Printf("Total NGLS:        %12.0f         %12.0f\n", roundToWholeNumber(accountingData.TotalNGLS), roundToWholeNumber(productionData.TotalNGLS))
	fmt.Printf("Production Total:  %12.0f         %12.0f\n", roundToWholeNumber(accountingData.ProductionTotalBBL), roundToWholeNumber(productionData.ProductionTotalBBL))
	fmt.Printf("Days On:           %12d         %12d\n", accountingData.DaysOn, productionData.DaysOn)
	fmt.Printf("\n")

	// Revenue Fields - Compare Accounting vs Production Date
	fmt.Printf("REVENUE FIELDS COMPARISON\n")
	fmt.Printf("=========================\n")
	fmt.Printf("                    Accounting Date    Production Date\n")
	fmt.Printf("                    ----------------    ---------------\n")
	fmt.Printf("Oil Revenue:       $%11.2f         $%11.2f\n", accountingData.OilRevenue, productionData.OilRevenue)
	fmt.Printf("Gas Revenue:       $%11.2f         $%11.2f\n", accountingData.GasRevenue, productionData.GasRevenue)
	fmt.Printf("NGL Revenue:       $%11.2f         $%11.2f\n", accountingData.OtherRevenue, productionData.OtherRevenue)
	fmt.Printf("\n")

	// Interest Groups
	fmt.Printf("INTEREST GROUPS (from WELLINV)\n")
	fmt.Printf("==============================\n")
	fmt.Printf("Royalty Oil Interest:      %.2f%%\n", well.RoyaltyOilInterest)
	fmt.Printf("Royalty Gas Interest:      %.2f%%\n", well.RoyaltyGasInterest)
	fmt.Printf("Royalty Other Interest:    %.2f%%\n", well.RoyaltyOtherInterest)
	fmt.Printf("Overriding Oil Interest:   %.2f%%\n", well.OverridingOilInterest)
	fmt.Printf("Overriding Gas Interest:   %.2f%%\n", well.OverridingGasInterest)
	fmt.Printf("Overriding Other Interest: %.2f%%\n", well.OverridingOtherInterest)
	fmt.Printf("Working Oil Interest:      %.2f%%\n", well.WorkingOilInterest)
	fmt.Printf("Working Gas Interest:      %.2f%%\n", well.WorkingGasInterest)
	fmt.Printf("Working Other Interest:    %.2f%%\n", well.WorkingOtherInterest)
	fmt.Printf("\n")

	// Working Interest Revenue (calculated) - Compare Accounting vs Production Date
	fmt.Printf("WORKING INTEREST REVENUE COMPARISON\n")
	fmt.Printf("===================================\n")

	// Accounting Date calculations
	accountingWiOilRevenue := accountingData.OilRevenue * (well.WorkingOilInterest / 100)
	accountingWiGasRevenue := accountingData.GasRevenue * (well.WorkingGasInterest / 100)
	accountingWiNglsRevenue := accountingData.OtherRevenue * (well.WorkingOtherInterest / 100)

	// Production Date calculations
	productionWiOilRevenue := productionData.OilRevenue * (well.WorkingOilInterest / 100)
	productionWiGasRevenue := productionData.GasRevenue * (well.WorkingGasInterest / 100)
	productionWiNglsRevenue := productionData.OtherRevenue * (well.WorkingOtherInterest / 100)

	fmt.Printf("                    Accounting Date    Production Date\n")
	fmt.Printf("                    ----------------    ---------------\n")
	fmt.Printf("WI Oil Revenue:    $%11.2f         $%11.2f\n", accountingWiOilRevenue, productionWiOilRevenue)
	fmt.Printf("WI Gas Revenue:    $%11.2f         $%11.2f\n", accountingWiGasRevenue, productionWiGasRevenue)
	fmt.Printf("WI NGL Revenue:    $%11.2f         $%11.2f\n", accountingWiNglsRevenue, productionWiNglsRevenue)
	fmt.Printf("\n")

	// Royalty Revenue (calculated) - Compare Accounting vs Production Date
	fmt.Printf("ROYALTY REVENUE COMPARISON\n")
	fmt.Printf("==========================\n")

	// Accounting Date calculations
	accountingRoyaltyOilRevenue := accountingData.OilRevenue * ((well.RoyaltyOilInterest + well.OverridingOilInterest) / 100)
	accountingRoyaltyGasRevenue := accountingData.GasRevenue * ((well.RoyaltyGasInterest + well.OverridingGasInterest) / 100)
	accountingRoyaltyNglsRevenue := accountingData.OtherRevenue * ((well.RoyaltyOtherInterest + well.OverridingOtherInterest) / 100)

	// Production Date calculations
	productionRoyaltyOilRevenue := productionData.OilRevenue * ((well.RoyaltyOilInterest + well.OverridingOilInterest) / 100)
	productionRoyaltyGasRevenue := productionData.GasRevenue * ((well.RoyaltyGasInterest + well.OverridingGasInterest) / 100)
	productionRoyaltyNglsRevenue := productionData.OtherRevenue * ((well.RoyaltyOtherInterest + well.OverridingOtherInterest) / 100)

	fmt.Printf("                    Accounting Date    Production Date\n")
	fmt.Printf("                    ----------------    ---------------\n")
	fmt.Printf("Royalty Oil Rev:   $%11.2f         $%11.2f\n", accountingRoyaltyOilRevenue, productionRoyaltyOilRevenue)
	fmt.Printf("Royalty Gas Rev:   $%11.2f         $%11.2f\n", accountingRoyaltyGasRevenue, productionRoyaltyGasRevenue)
	fmt.Printf("Royalty NGL Rev:   $%11.2f         $%11.2f\n", accountingRoyaltyNglsRevenue, productionRoyaltyNglsRevenue)
	fmt.Printf("\n")

	// Expenses - Compare Accounting vs Production Date
	fmt.Printf("EXPENSES COMPARISON\n")
	fmt.Printf("===================\n")
	fmt.Printf("                    Accounting Date    Production Date\n")
	fmt.Printf("                    ----------------    ---------------\n")
	fmt.Printf("Total Expenses:    $%11.2f         $%11.2f\n", accountingData.TotalExpenses, productionData.TotalExpenses)
	fmt.Printf("\n")

	// Calculate expenses based on revenue percentages for both date types
	fmt.Printf("EXPENSE ALLOCATION (Accounting Date):\n")
	accountingTotalRevenue := accountingData.OilRevenue + accountingData.GasRevenue + accountingData.OtherRevenue
	var accountingOilExpenses, accountingGasExpenses, accountingNglExpenses float64

	if accountingTotalRevenue == 0 {
		accountingOilExpenses = accountingData.TotalExpenses
		accountingGasExpenses = 0.0
		accountingNglExpenses = 0.0
		fmt.Printf("  Oil Expenses:      $%.2f (100%% - no revenue)\n", accountingOilExpenses)
		fmt.Printf("  Gas Expenses:      $%.2f (0%%)\n", accountingGasExpenses)
		fmt.Printf("  NGL Expenses:      $%.2f (0%%)\n", accountingNglExpenses)
	} else {
		accountingOilExpenses = accountingData.TotalExpenses * (accountingData.OilRevenue / accountingTotalRevenue)
		accountingGasExpenses = accountingData.TotalExpenses * (accountingData.GasRevenue / accountingTotalRevenue)
		accountingNglExpenses = accountingData.TotalExpenses * (accountingData.OtherRevenue / accountingTotalRevenue)
		fmt.Printf("  Oil Expenses:      $%.2f (%.1f%% of revenue)\n", accountingOilExpenses, (accountingData.OilRevenue/accountingTotalRevenue)*100)
		fmt.Printf("  Gas Expenses:      $%.2f (%.1f%% of revenue)\n", accountingGasExpenses, (accountingData.GasRevenue/accountingTotalRevenue)*100)
		fmt.Printf("  NGL Expenses:      $%.2f (%.1f%% of revenue)\n", accountingNglExpenses, (accountingData.OtherRevenue/accountingTotalRevenue)*100)
	}

	fmt.Printf("\nEXPENSE ALLOCATION (Production Date):\n")
	productionTotalRevenue := productionData.OilRevenue + productionData.GasRevenue + productionData.OtherRevenue
	var productionOilExpenses, productionGasExpenses, productionNglExpenses float64

	if productionTotalRevenue == 0 {
		productionOilExpenses = productionData.TotalExpenses
		productionGasExpenses = 0.0
		productionNglExpenses = 0.0
		fmt.Printf("  Oil Expenses:      $%.2f (100%% - no revenue)\n", productionOilExpenses)
		fmt.Printf("  Gas Expenses:      $%.2f (0%%)\n", productionGasExpenses)
		fmt.Printf("  NGL Expenses:      $%.2f (0%%)\n", productionNglExpenses)
	} else {
		productionOilExpenses = productionData.TotalExpenses * (productionData.OilRevenue / productionTotalRevenue)
		productionGasExpenses = productionData.TotalExpenses * (productionData.GasRevenue / productionTotalRevenue)
		productionNglExpenses = productionData.TotalExpenses * (productionData.OtherRevenue / productionTotalRevenue)
		fmt.Printf("  Oil Expenses:      $%.2f (%.1f%% of revenue)\n", productionOilExpenses, (productionData.OilRevenue/productionTotalRevenue)*100)
		fmt.Printf("  Gas Expenses:      $%.2f (%.1f%% of revenue)\n", productionGasExpenses, (productionData.GasRevenue/productionTotalRevenue)*100)
		fmt.Printf("  NGL Expenses:      $%.2f (%.1f%% of revenue)\n", productionNglExpenses, (productionData.OtherRevenue/productionTotalRevenue)*100)
	}
	fmt.Printf("\n")

	// Totals - Compare Accounting vs Production Date
	fmt.Printf("TOTALS COMPARISON\n")
	fmt.Printf("=================\n")
	accountingWiTotal := accountingWiOilRevenue + accountingWiGasRevenue + accountingWiNglsRevenue
	accountingRoyaltyTotal := accountingRoyaltyOilRevenue + accountingRoyaltyGasRevenue + accountingRoyaltyNglsRevenue

	productionWiTotal := productionWiOilRevenue + productionWiGasRevenue + productionWiNglsRevenue
	productionRoyaltyTotal := productionRoyaltyOilRevenue + productionRoyaltyGasRevenue + productionRoyaltyNglsRevenue

	fmt.Printf("                    Accounting Date    Production Date\n")
	fmt.Printf("                    ----------------    ---------------\n")
	fmt.Printf("WI Total:          $%11.2f         $%11.2f\n", accountingWiTotal, productionWiTotal)
	fmt.Printf("WI DOI Total:      (blank for now)    (blank for now)\n")
	fmt.Printf("Royalty Total:     $%11.2f         $%11.2f\n", accountingRoyaltyTotal, productionRoyaltyTotal)
	fmt.Printf("Royalty DOI Total: (blank for now)    (blank for now)\n")
	fmt.Printf("\n")

	// Interest Totals
	fmt.Printf("INTEREST TOTALS (should equal 100%%)\n")
	fmt.Printf("===================================\n")
	totalOilInterest := well.RoyaltyOilInterest + well.OverridingOilInterest + well.WorkingOilInterest
	totalGasInterest := well.RoyaltyGasInterest + well.OverridingGasInterest + well.WorkingGasInterest
	totalOtherInterest := well.RoyaltyOtherInterest + well.OverridingOtherInterest + well.WorkingOtherInterest
	fmt.Printf("Total Oil Interest:   %.2f%%\n", totalOilInterest)
	fmt.Printf("Total Gas Interest:   %.2f%%\n", totalGasInterest)
	fmt.Printf("Total Other Interest: %.2f%%\n", totalOtherInterest)
	fmt.Printf("\n")

	// Debug: Show the actual records used
	fmt.Printf("DEBUG: RECORDS USED\n")
	fmt.Printf("===================\n")
	fmt.Printf("ACCOUNTING DATE RECORDS (%d records):\n", len(accountingRecords))
	for i, record := range accountingRecords {
		fmt.Printf("  Record %d: WellID=%s, RunYear=%s, ProdYear=%s, HDate=%s, Oil=%.2f, Gas=%.2f, NGL=%.2f, Total=%.2f, Days=%d, OilRev=%.2f, GasRev=%.2f, OtherRev=%.2f, Expenses=%.2f\n",
			i+1,
			record["cwellid"],
			record["crunyear"],
			record["hyear"],
			record["hdate"],
			record["ntotbbl"],
			record["ntotmcf"],
			record["ntotprod"],
			record["ngrossbbl"],
			record["ndayson"],
			record["ngrossoil"],
			record["ngrossgas"],
			record["nothinc"],
			record["ntotale"])
	}
	fmt.Printf("\nPRODUCTION DATE RECORDS (%d records):\n", len(productionRecords))
	for i, record := range productionRecords {
		fmt.Printf("  Record %d: WellID=%s, RunYear=%s, ProdYear=%s, HDate=%s, Oil=%.2f, Gas=%.2f, NGL=%.2f, Total=%.2f, Days=%d, OilRev=%.2f, GasRev=%.2f, OtherRev=%.2f, Expenses=%.2f\n",
			i+1,
			record["cwellid"],
			record["crunyear"],
			record["hyear"],
			record["hdate"],
			record["ntotbbl"],
			record["ntotmcf"],
			record["ntotprod"],
			record["ngrossbbl"],
			record["ndayson"],
			record["ngrossoil"],
			record["ngrossgas"],
			record["nothinc"],
			record["ntotale"])
	}
	fmt.Printf("\n")

	fmt.Printf("===============================================\n")
	fmt.Printf("END OF SINGLE WELL ANALYSIS\n")
	fmt.Printf("===============================================\n\n")

	// Ask if user wants to generate PDF for this well
	fmt.Print("Generate PDF for this well? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Println("\nExiting due to EOF...")
			return
		}
		fmt.Printf("Error reading input: %v\n", err)
		return
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "y" || response == "yes" {
		// Generate PDF for this specific well
		templatePath := "template/stc1235.page1.pdf"
		outputDir := "output/single_well_analysis"

		// Create output directory
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("❌ Error creating output directory: %v\n", err)
			return
		}

		// Create well-specific filename
		wellID := strings.ReplaceAll(well.WellID, " ", "_")
		operatorName := strings.ReplaceAll(producer.CompanyName, " ", "_")
		filename := fmt.Sprintf("ANALYSIS_%s_%s.pdf", operatorName, wellID)
		outputPath := filepath.Join(outputDir, filename)

		// Create a dummy config for PDF generation
		dummyConfig := &WVConfig{
			ReportingYear:  26,   // Default reporting year
			ProductionYear: 2024, // Default production year
		}

		// Use accounting data for PDF generation (standard for forms)
		// Copy the well info but use accounting data for production/revenue fields
		pdfWell := *well // Copy the original well
		pdfWell.TotalOilBBL = accountingData.TotalOilBBL
		pdfWell.TotalGasMCF = accountingData.TotalGasMCF
		pdfWell.TotalNGLS = accountingData.TotalNGLS
		pdfWell.ProductionTotalBBL = accountingData.ProductionTotalBBL
		pdfWell.DaysOn = accountingData.DaysOn
		pdfWell.OilRevenue = accountingData.OilRevenue
		pdfWell.GasRevenue = accountingData.GasRevenue
		pdfWell.OtherRevenue = accountingData.OtherRevenue
		pdfWell.TotalExpenses = accountingData.TotalExpenses

		fmt.Printf("Generating PDF: %s\n", outputPath)
		err := FillWVForm(producer, dummyConfig, &pdfWell, templatePath, outputPath)
		if err != nil {
			fmt.Printf("❌ Error generating PDF: %v\n", err)
		} else {
			fmt.Printf("✅ PDF generated successfully: %s\n", outputPath)
			fmt.Printf("📄 You can now compare the text analysis above with the PDF output\n")
		}
	}
}

// AnalyzeNRANumbers analyzes wells and prints those with multiple NRA numbers
func AnalyzeNRANumbers(wells []*WellInfo) {
	fmt.Printf("\n===============================================\n")
	fmt.Printf("NRA NUMBER ANALYSIS\n")
	fmt.Printf("===============================================\n\n")

	wellsWithMultipleNRA := 0
	wellsWithValidNRA := 0

	for _, well := range wells {
		// Count non-MISSING NRA values
		nraCount := 0
		nraValues := []string{}

		if well.NRA != "MISSING" {
			nraCount++
			nraValues = append(nraValues, well.NRA)
		}
		if well.NRA2 != "MISSING" {
			nraCount++
			nraValues = append(nraValues, well.NRA2)
		}
		if well.NRA3 != "MISSING" {
			nraCount++
			nraValues = append(nraValues, well.NRA3)
		}
		if well.NRA4 != "MISSING" {
			nraCount++
			nraValues = append(nraValues, well.NRA4)
		}
		if well.NRA5 != "MISSING" {
			nraCount++
			nraValues = append(nraValues, well.NRA5)
		}
		if well.NRA6 != "MISSING" {
			nraCount++
			nraValues = append(nraValues, well.NRA6)
		}

		if nraCount > 0 {
			wellsWithValidNRA++
		}

		if nraCount > 1 {
			wellsWithMultipleNRA++
			fmt.Printf("Well: %s (%s)\n", well.WellName, well.WellID)
			fmt.Printf("  County: %s (%s)\n", well.County, well.CountyCode)
			fmt.Printf("  NRA Count: %d\n", nraCount)
			fmt.Printf("  NRA Values: %s\n", strings.Join(nraValues, ", "))
			fmt.Printf("  API: %s\n", well.API)
			fmt.Printf("  Formation: %s\n", well.Formation)
			fmt.Printf("  Field ID: %s\n", well.FieldID)
			fmt.Printf("  Well Status: %s\n", well.WellStatus)
			fmt.Printf("  ---\n")
		}
	}

	fmt.Printf("\nSUMMARY:\n")
	fmt.Printf("  Total wells analyzed: %d\n", len(wells))
	fmt.Printf("  Wells with at least one NRA: %d (%.1f%%)\n", wellsWithValidNRA, float64(wellsWithValidNRA)/float64(len(wells))*100)
	fmt.Printf("  Wells with multiple NRA numbers: %d (%.1f%%)\n", wellsWithMultipleNRA, float64(wellsWithMultipleNRA)/float64(len(wells))*100)
	fmt.Printf("===============================================\n\n")
}

// AnalyzeWellStatusDistribution analyzes the distribution of well statuses
func AnalyzeWellStatusDistribution(wells []*WellInfo) {
	fmt.Printf("\n===============================================\n")
	fmt.Printf("WELL STATUS DISTRIBUTION ANALYSIS\n")
	fmt.Printf("===============================================\n\n")

	statusCounts := make(map[string]int)
	totalWells := len(wells)

	for _, well := range wells {
		status := well.WellStatus
		if status == "" {
			status = "UNKNOWN"
		}
		statusCounts[status]++
	}

	fmt.Printf("Well Status Distribution:\n")
	fmt.Printf("-------------------------\n")
	for status, count := range statusCounts {
		percentage := float64(count) / float64(totalWells) * 100
		fmt.Printf("  %s: %d wells (%.1f%%)\n", status, count, percentage)
	}
	fmt.Printf("\nTotal wells analyzed: %d\n", totalWells)
	fmt.Printf("===============================================\n\n")
}

// AnalyzeCountyDistribution analyzes the distribution of wells by county
func AnalyzeCountyDistribution(wells []*WellInfo) {
	fmt.Printf("\n===============================================\n")
	fmt.Printf("COUNTY DISTRIBUTION ANALYSIS\n")
	fmt.Printf("===============================================\n\n")

	countyCounts := make(map[string]int)
	totalWells := len(wells)

	for _, well := range wells {
		county := well.County
		if county == "" {
			county = "UNKNOWN"
		}
		countyCounts[county]++
	}

	fmt.Printf("County Distribution (Top 20):\n")
	fmt.Printf("------------------------------\n")

	// Sort counties by count (descending)
	type countyCount struct {
		county string
		count  int
	}
	var sortedCounties []countyCount
	for county, count := range countyCounts {
		sortedCounties = append(sortedCounties, countyCount{county, count})
	}

	// Sort by count descending
	for i := 0; i < len(sortedCounties)-1; i++ {
		for j := i + 1; j < len(sortedCounties); j++ {
			if sortedCounties[i].count < sortedCounties[j].count {
				sortedCounties[i], sortedCounties[j] = sortedCounties[j], sortedCounties[i]
			}
		}
	}

	// Display top 20
	for i, cc := range sortedCounties {
		if i >= 20 {
			break
		}
		percentage := float64(cc.count) / float64(totalWells) * 100
		fmt.Printf("  %d. %s: %d wells (%.1f%%)\n", i+1, cc.county, cc.count, percentage)
	}

	fmt.Printf("\nTotal wells analyzed: %d\n", totalWells)
	fmt.Printf("Total counties: %d\n", len(countyCounts))
	fmt.Printf("===============================================\n\n")
}

// AnalyzeFormationDistribution analyzes the distribution of wells by formation
func AnalyzeFormationDistribution(wells []*WellInfo) {
	fmt.Printf("\n===============================================\n")
	fmt.Printf("FORMATION DISTRIBUTION ANALYSIS\n")
	fmt.Printf("===============================================\n\n")

	formationCounts := make(map[string]int)
	totalWells := len(wells)

	for _, well := range wells {
		formation := well.Formation
		if formation == "" {
			formation = "UNKNOWN"
		}
		formationCounts[formation]++
	}

	fmt.Printf("Formation Distribution (Top 20):\n")
	fmt.Printf("--------------------------------\n")

	// Sort formations by count (descending)
	type formationCount struct {
		formation string
		count     int
	}
	var sortedFormations []formationCount
	for formation, count := range formationCounts {
		sortedFormations = append(sortedFormations, formationCount{formation, count})
	}

	// Sort by count descending
	for i := 0; i < len(sortedFormations)-1; i++ {
		for j := i + 1; j < len(sortedFormations); j++ {
			if sortedFormations[i].count < sortedFormations[j].count {
				sortedFormations[i], sortedFormations[j] = sortedFormations[j], sortedFormations[i]
			}
		}
	}

	// Display top 20
	for i, fc := range sortedFormations {
		if i >= 20 {
			break
		}
		percentage := float64(fc.count) / float64(totalWells) * 100
		fmt.Printf("  %d. %s: %d wells (%.1f%%)\n", i+1, fc.formation, fc.count, percentage)
	}

	fmt.Printf("\nTotal wells analyzed: %d\n", totalWells)
	fmt.Printf("Total formations: %d\n", len(formationCounts))
	fmt.Printf("===============================================\n\n")
}

// AnalyzeGasTypeDistribution analyzes the distribution of gas types
func AnalyzeGasTypeDistribution(wells []*WellInfo) {
	fmt.Printf("\n===============================================\n")
	fmt.Printf("GAS TYPE DISTRIBUTION ANALYSIS\n")
	fmt.Printf("===============================================\n\n")

	totalWells := len(wells)
	ethaneCount := 0
	propaneCount := 0
	butaneCount := 0
	isobutaneCount := 0
	pentaneCount := 0
	houseGasCount := 0

	for _, well := range wells {
		if well.HasEthane {
			ethaneCount++
		}
		if well.HasPropane {
			propaneCount++
		}
		if well.HasButane {
			butaneCount++
		}
		if well.HasIsobutane {
			isobutaneCount++
		}
		if well.HasPentane {
			pentaneCount++
		}
		if well.HasHouseGas {
			houseGasCount++
		}
	}

	fmt.Printf("Gas Type Distribution:\n")
	fmt.Printf("----------------------\n")
	fmt.Printf("  Ethane: %d wells (%.1f%%)\n", ethaneCount, float64(ethaneCount)/float64(totalWells)*100)
	fmt.Printf("  Propane: %d wells (%.1f%%)\n", propaneCount, float64(propaneCount)/float64(totalWells)*100)
	fmt.Printf("  Butane: %d wells (%.1f%%)\n", butaneCount, float64(butaneCount)/float64(totalWells)*100)
	fmt.Printf("  Isobutane: %d wells (%.1f%%)\n", isobutaneCount, float64(isobutaneCount)/float64(totalWells)*100)
	fmt.Printf("  Pentane: %d wells (%.1f%%)\n", pentaneCount, float64(pentaneCount)/float64(totalWells)*100)
	fmt.Printf("  House Gas: %d wells (%.1f%%)\n", houseGasCount, float64(houseGasCount)/float64(totalWells)*100)

	fmt.Printf("\nTotal wells analyzed: %d\n", totalWells)
	fmt.Printf("===============================================\n\n")
}

// AnalyzePermitNumbers analyzes wells and prints those with multiple permit numbers
func AnalyzePermitNumbers(wells []*WellInfo) {
	fmt.Printf("\n===============================================\n")
	fmt.Printf("PERMIT NUMBER ANALYSIS\n")
	fmt.Printf("===============================================\n\n")

	wellsWithMultiplePermits := 0
	wellsWithValidPermits := 0

	for _, well := range wells {
		// Count non-MISSING permit values
		permitCount := 0
		permitValues := []string{}

		if well.API != "MISSING" && well.API != "NULL" {
			permitCount++
			permitValues = append(permitValues, well.API)
		}
		if well.Permit2 != "MISSING" && well.Permit2 != "NULL" {
			permitCount++
			permitValues = append(permitValues, well.Permit2)
		}
		if well.Permit3 != "MISSING" && well.Permit3 != "NULL" {
			permitCount++
			permitValues = append(permitValues, well.Permit3)
		}
		if well.Permit4 != "MISSING" && well.Permit4 != "NULL" {
			permitCount++
			permitValues = append(permitValues, well.Permit4)
		}
		if well.Permit5 != "MISSING" && well.Permit5 != "NULL" {
			permitCount++
			permitValues = append(permitValues, well.Permit5)
		}
		if well.Permit6 != "MISSING" && well.Permit6 != "NULL" {
			permitCount++
			permitValues = append(permitValues, well.Permit6)
		}

		if permitCount > 0 {
			wellsWithValidPermits++
		}

		if permitCount > 1 {
			wellsWithMultiplePermits++
			fmt.Printf("Well: %s (%s)\n", well.WellName, well.WellID)
			fmt.Printf("  County: %s (%s)\n", well.County, well.CountyCode)
			fmt.Printf("  Permit Count: %d\n", permitCount)
			fmt.Printf("  Permit Values: %s\n", strings.Join(permitValues, ", "))
			fmt.Printf("  NRA: %s\n", well.NRA)
			fmt.Printf("  Formation: %s\n", well.Formation)
			fmt.Printf("  Field ID: %s\n", well.FieldID)
			fmt.Printf("  Well Status: %s\n", well.WellStatus)
			fmt.Printf("  ---\n")
		}
	}

	fmt.Printf("\nSUMMARY:\n")
	fmt.Printf("  Total wells analyzed: %d\n", len(wells))
	fmt.Printf("  Wells with at least one permit: %d (%.1f%%)\n", wellsWithValidPermits, float64(wellsWithValidPermits)/float64(len(wells))*100)
	fmt.Printf("  Wells with multiple permit numbers: %d (%.1f%%)\n", wellsWithMultiplePermits, float64(wellsWithMultiplePermits)/float64(len(wells))*100)
	fmt.Printf("===============================================\n\n")
}

// AnalyzeSingleWellRevenue analyzes revenue for a single well across wellhist, disbhist, and ownerhist tables
// Shows both accounting and production year data for comparison
func AnalyzeSingleWellRevenue(wells []*WellInfo, sourceDataPath string, year int) {
	fmt.Printf("\n===============================================\n")
	fmt.Printf("SINGLE WELL REVENUE ANALYSIS\n")
	fmt.Printf("===============================================\n\n")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Well ID to analyze: ")
	wellID, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Println("\nExiting due to EOF...")
			os.Exit(0)
		}
		fmt.Printf("Error reading input: %v\n", err)
		return
	}
	wellID = strings.TrimSpace(wellID)

	fmt.Printf("\nAnalyzing raw data for Well ID: %s\n", wellID)

	// Read wellhist data for both accounting and production years
	fmt.Printf("\n--- WELLHIST DATA COMPARISON ---\n")

	// Get the year from the first well (they should all be the same)
	// year := 2024 // Default, but we'll get it from the wells data
	// if len(wells) > 0 {
	// 	// We need to get the year from the config, but we don't have it here
	// 	// For now, let's assume 2024 and read both years
	// 	year = 2024
	// }

	// Read accounting year data
	accountingData, _, err := readSingleWellWellhistDataWithRecords(wellID, sourceDataPath, year, AccountingDate)
	if err != nil {
		fmt.Printf("Error reading accounting data: %v\n", err)
		return
	}

	// Read production year data
	productionData, _, err := readSingleWellWellhistDataWithRecords(wellID, sourceDataPath, year, ProductionDate)
	if err != nil {
		fmt.Printf("Error reading production data: %v\n", err)
		return
	}

	fmt.Printf("ACCOUNTING YEAR (%d) DATA:\n", year)
	fmt.Printf("  Total Oil (BBL): %.2f\n", accountingData.TotalOilBBL)
	fmt.Printf("  Total Gas (MCF): %.2f\n", accountingData.TotalGasMCF)
	fmt.Printf("  Total NGLS: %.2f\n", accountingData.TotalNGLS)
	fmt.Printf("  Production Total (BBL): %.2f\n", accountingData.ProductionTotalBBL)
	fmt.Printf("  Days On: %d\n", accountingData.DaysOn)

	fmt.Printf("\nPRODUCTION YEAR (%d) DATA:\n", year)
	fmt.Printf("  Total Oil (BBL): %.2f\n", productionData.TotalOilBBL)
	fmt.Printf("  Total Gas (MCF): %.2f\n", productionData.TotalGasMCF)
	fmt.Printf("  Total NGLS: %.2f\n", productionData.TotalNGLS)
	fmt.Printf("  Production Total (BBL): %.2f\n", productionData.ProductionTotalBBL)
	fmt.Printf("  Days On: %d\n", productionData.DaysOn)

	// TODO: Add disbhist and ownerhist analysis
	fmt.Printf("\n--- DISBHIST DATA ---\n")
	fmt.Println("(To be implemented - need to read disbhist.DBF)")

	fmt.Printf("\n--- OWNERHIST DATA ---\n")
	fmt.Println("(To be implemented - need to read ownerhist.DBF)")

	fmt.Printf("\n===============================================\n\n")
	fmt.Print("Press Enter to continue with more analysis, or 'q' to quit: ")
	_, err = reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Println("\nExiting due to EOF...")
			os.Exit(0)
		}
		return
	}
}

func readSingleWellWellhistDataWithRecords(wellID, sourceDataPath string, year int, dateType DateType) (*WellInfo, []map[string]interface{}, error) {
	dbfPath := filepath.Join(sourceDataPath, "WELLHIST.DBF")
	config := &dbase.Config{Filename: dbfPath}
	table, err := dbase.OpenTable(config)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening wellhist.dbf: %v", err)
	}
	defer table.Close()

	wellData := &WellInfo{WellID: wellID}
	var records []map[string]interface{}

	for !table.EOF() {
		row, err := table.Next()
		if err != nil {
			continue
		}

		// Skip deleted records
		if row.Deleted {
			continue
		}

		// Get well ID
		cwellid, _ := row.ValueByName("CWELLID")
		recordWellID := getStringField(cwellid)

		// Only process records for the specified well
		if recordWellID != wellID {
			continue
		}

		// Apply year filtering based on date type
		var shouldInclude bool
		var crunyear, hyear, hdate interface{}
		if dateType == AccountingDate {
			// For accounting date, filter on HDATE - check if it falls in the year
			hdate, _ = row.ValueByName("HDATE")
			recordDate := getDateField(hdate)
			if recordDate != "" {
				// Parse the date and check if it's in the specified year
				// Handle different date formats: MM/DD/YYYY, YYYY-MM-DD, etc.
				if len(recordDate) >= 4 {
					// Try to extract year from the date string
					var dateYear string
					if strings.Contains(recordDate, "/") {
						// Format: MM/DD/YYYY
						parts := strings.Split(recordDate, "/")
						if len(parts) >= 3 {
							dateYear = parts[2]
						}
					} else if strings.Contains(recordDate, "-") {
						// Format: YYYY-MM-DD
						parts := strings.Split(recordDate, "-")
						if len(parts) >= 1 {
							dateYear = parts[0]
						}
					} else if len(recordDate) >= 4 {
						// Format: YYYYMMDD or similar
						dateYear = recordDate[:4]
					}

					if dateYear != "" {
						// Compare with the production year directly
						shouldInclude = dateYear == fmt.Sprintf("%d", year)
					} else {
						shouldInclude = false
					}
				} else {
					shouldInclude = false
				}
			} else {
				shouldInclude = false
			}
		} else {
			// For production date, filter on HYEAR (Production Year)
			hyear, _ = row.ValueByName("HYEAR")
			recordProdYear := int(getFloatField(hyear))
			shouldInclude = recordProdYear == year
		}
		if !shouldInclude {
			continue
		}

		// Get production data
		ntotbbl, _ := row.ValueByName("NTOTBBL")
		ntotmcf, _ := row.ValueByName("NTOTMCF")
		ntotprod, _ := row.ValueByName("NTOTPROD")
		ngrossbbl, _ := row.ValueByName("NGROSSBBL")
		ndayson, _ := row.ValueByName("NDAYSON")

		// Get revenue data
		ngrossoil, _ := row.ValueByName("NGROSSOIL")
		ngrossgas, _ := row.ValueByName("NGROSSGAS")
		nothinc, _ := row.ValueByName("NOTHINC")

		// Get expense data
		ntotale, _ := row.ValueByName("NTOTALE")
		nexpcl1, _ := row.ValueByName("NEXPCL1")
		nexpcl2, _ := row.ValueByName("NEXPCL2")
		nexpcl3, _ := row.ValueByName("NEXPCL3")
		nexpcl4, _ := row.ValueByName("NEXPCL4")
		nexpcl5, _ := row.ValueByName("NEXPCL5")

		// Store the record for debugging
		record := make(map[string]interface{})
		record["cwellid"] = getStringField(cwellid)
		record["crunyear"] = getStringField(crunyear)
		record["hyear"] = getStringField(hyear)
		record["hdate"] = getDateField(hdate)
		record["ntotbbl"] = getFloatField(ntotbbl)
		record["ntotmcf"] = getFloatField(ntotmcf)
		record["ntotprod"] = getFloatField(ntotprod)
		record["ngrossbbl"] = getFloatField(ngrossbbl)
		record["ndayson"] = getFloatField(ndayson)
		record["ngrossoil"] = getFloatField(ngrossoil)
		record["ngrossgas"] = getFloatField(ngrossgas)
		record["nothinc"] = getFloatField(nothinc)
		record["ntotale"] = getFloatField(ntotale)
		record["nexpcl1"] = getFloatField(nexpcl1)
		record["nexpcl2"] = getFloatField(nexpcl2)
		record["nexpcl3"] = getFloatField(nexpcl3)
		record["nexpcl4"] = getFloatField(nexpcl4)
		record["nexpcl5"] = getFloatField(nexpcl5)
		records = append(records, record)

		wellData.TotalOilBBL += getFloatField(ntotbbl)
		wellData.TotalGasMCF += getFloatField(ntotmcf)
		wellData.TotalNGLS += getFloatField(ntotprod)
		wellData.ProductionTotalBBL += getFloatField(ngrossbbl)
		wellData.DaysOn += int(getFloatField(ndayson))

		// Add revenue data
		wellData.OilRevenue += getFloatField(ngrossoil)
		wellData.GasRevenue += getFloatField(ngrossgas)
		wellData.OtherRevenue += getFloatField(nothinc)

		// Add expense data
		wellData.TotalExpenses += getFloatField(ntotale) +
			getFloatField(nexpcl1) +
			getFloatField(nexpcl2) +
			getFloatField(nexpcl3) +
			getFloatField(nexpcl4) +
			getFloatField(nexpcl5)
	}

	return wellData, records, nil
}

// handleDataAnalysisMode handles the data analysis menu and options
func handleDataAnalysisMode(wells []*WellInfo, config *WVConfig, producer *ProducerInfo) {
	for {
		choice := PromptForDataAnalysisMenu()
		switch choice {
		case "0":
			return
		case "1":
			AnalyzeNRANumbers(wells)
		case "2":
			AnalyzeWellStatusDistribution(wells)
		case "3":
			AnalyzeCountyDistribution(wells)
		case "4":
			AnalyzeFormationDistribution(wells)
		case "5":
			AnalyzeGasTypeDistribution(wells)
		case "6":
			AnalyzePermitNumbers(wells)
		case "7":
			AnalyzeSingleWellRevenue(wells, config.SourceDataPath, config.ProductionYear)
		case "8":
			DisplaySingleWellPDFFields(wells, producer)
		default:
			fmt.Println("Invalid choice. Please enter 0-8.")
			continue
		}
	}
}

// ProducerLookup represents a producer entry from the CSV lookup file
type ProducerLookup struct {
	ProdCode     string
	ProducerName string
}

// getProducerLookup reads the producer lookup CSV file
func getProducerLookup() ([]ProducerLookup, error) {
	lookupPath := "seeds/wv_producers.csv"

	file, err := os.Open(lookupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open producer lookup file %s: %v", lookupPath, err)
	}
	defer file.Close()

	var producers []ProducerLookup
	scanner := bufio.NewScanner(file)

	// Skip header line
	if scanner.Scan() {
		// Skip header
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")
		if len(fields) >= 2 {
			producers = append(producers, ProducerLookup{
				ProdCode:     strings.TrimSpace(fields[0]),
				ProducerName: strings.TrimSpace(fields[1]),
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading producer lookup file: %v", err)
	}

	return producers, nil
}

// findProducerMatch attempts to match producer info from VERSION table with lookup data
func findProducerMatch(producer *ProducerInfo, lookups []ProducerLookup) (string, bool) {
	// Try to match by producer code first
	for _, lookup := range lookups {
		if strings.EqualFold(producer.IDComp, lookup.ProdCode) {
			return lookup.ProdCode, true
		}
	}

	// Try to match by company name
	for _, lookup := range lookups {
		if strings.EqualFold(strings.TrimSpace(producer.CompanyName), strings.TrimSpace(lookup.ProducerName)) {
			return lookup.ProdCode, true
		}
	}

	// Try partial name matching
	for _, lookup := range lookups {
		if strings.Contains(strings.ToUpper(producer.CompanyName), strings.ToUpper(lookup.ProducerName)) ||
			strings.Contains(strings.ToUpper(lookup.ProducerName), strings.ToUpper(producer.CompanyName)) {
			return lookup.ProdCode, true
		}
	}

	return "", false
}

// promptForProducerSelection allows user to select a producer from the lookup list
func promptForProducerSelection(lookups []ProducerLookup) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== Producer Selection ===")
	fmt.Printf("Found %d producers in lookup file.\n", len(lookups))
	fmt.Println("Please select the correct producer:")

	// Show first 20 producers as examples
	for i := 0; i < 20 && i < len(lookups); i++ {
		fmt.Printf("%d. %s (%s)\n", i+1, lookups[i].ProducerName, lookups[i].ProdCode)
	}

	if len(lookups) > 20 {
		fmt.Printf("... and %d more producers\n", len(lookups)-20)
	}

	fmt.Println("\nOptions:")
	fmt.Println("0. Enter custom producer code")
	fmt.Println("-1. Search by name")

	for {
		fmt.Print("Enter choice (0, -1, or producer number): ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nExiting due to EOF...")
				os.Exit(0)
			}
			return "", fmt.Errorf("error reading input: %v", err)
		}

		choice = strings.TrimSpace(choice)

		if choice == "0" {
			// Custom producer code
			fmt.Print("Enter custom producer code: ")
			customCode, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("\nExiting due to EOF...")
					os.Exit(0)
				}
				return "", fmt.Errorf("error reading input: %v", err)
			}
			return strings.TrimSpace(customCode), nil
		}

		if choice == "-1" {
			// Search by name
			fmt.Print("Enter producer name to search for: ")
			searchTerm, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("\nExiting due to EOF...")
					os.Exit(0)
				}
				return "", fmt.Errorf("error reading input: %v", err)
			}
			searchTerm = strings.TrimSpace(strings.ToUpper(searchTerm))

			var matches []ProducerLookup
			for _, lookup := range lookups {
				if strings.Contains(strings.ToUpper(lookup.ProducerName), searchTerm) {
					matches = append(matches, lookup)
				}
			}

			if len(matches) == 0 {
				fmt.Println("No matches found. Please try again.")
				continue
			}

			fmt.Printf("\nFound %d matches:\n", len(matches))
			for i, match := range matches {
				fmt.Printf("%d. %s (%s)\n", i+1, match.ProducerName, match.ProdCode)
			}

			fmt.Print("Select match (or 0 to search again): ")
			matchChoice, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("\nExiting due to EOF...")
					os.Exit(0)
				}
				return "", fmt.Errorf("error reading input: %v", err)
			}

			matchChoice = strings.TrimSpace(matchChoice)
			if matchChoice == "0" {
				continue
			}

			choiceNum, err := strconv.Atoi(matchChoice)
			if err != nil || choiceNum < 1 || choiceNum > len(matches) {
				fmt.Println("Invalid choice. Please try again.")
				continue
			}

			return matches[choiceNum-1].ProdCode, nil
		}

		// Try to parse as number for direct selection
		choiceNum, err := strconv.Atoi(choice)
		if err != nil || choiceNum < 1 || choiceNum > len(lookups) {
			fmt.Println("Invalid choice. Please try again.")
			continue
		}

		return lookups[choiceNum-1].ProdCode, nil
	}
}

// resolveProducerCode determines the correct producer code using lookup and user input
func resolveProducerCode(producer *ProducerInfo) (string, error) {
	// Read producer lookup data
	lookups, err := getProducerLookup()
	if err != nil {
		fmt.Printf("Warning: Could not read producer lookup file: %v\n", err)
		fmt.Printf("Using producer code from VERSION table: %s\n", producer.IDComp)
		return producer.IDComp, nil
	}

	// Try to find a match
	if matchedCode, found := findProducerMatch(producer, lookups); found {
		fmt.Printf("✅ Found matching producer: %s (%s)\n", producer.CompanyName, matchedCode)
		return matchedCode, nil
	}

	// No match found, ask user to select
	fmt.Printf("❌ No match found for producer: %s (Code: %s)\n", producer.CompanyName, producer.IDComp)

	selectedCode, err := promptForProducerSelection(lookups)
	if err != nil {
		return "", err
	}

	fmt.Printf("✅ Selected producer code: %s\n", selectedCode)
	return selectedCode, nil
}

// handleDatabaseOperations handles database operations
func handleDatabaseOperations(config *WVConfig, producer *ProducerInfo) {
	fmt.Println("\n=== Database Operations ===")

	// Check if SQL folder and database exist
	sqlDir := filepath.Join("sourcedata", "sql")
	dbPath := filepath.Join(sqlDir, "wv_operator.db")

	// Check if SQL directory exists
	if _, err := os.Stat(sqlDir); os.IsNotExist(err) {
		fmt.Printf("❌ SQL directory not found: %s\n", sqlDir)
		fmt.Print("Would you like to create the SQL directory and database? (y/n): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Error reading input: %v\n", err)
			return
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ Database creation cancelled.")
			return
		}

		// Create SQL directory
		if err := os.MkdirAll(sqlDir, 0755); err != nil {
			fmt.Printf("❌ Error creating SQL directory: %v\n", err)
			return
		}
		fmt.Printf("✅ Created SQL directory: %s\n", sqlDir)
	}

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("❌ Database not found: %s\n", dbPath)
		fmt.Print("Would you like to create a new database? (y/n): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Error reading input: %v\n", err)
			return
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("❌ Database creation cancelled.")
			return
		}

		fmt.Println("Creating new database...")
	}

	// Initialize database
	dbManager, err := NewDatabaseManager(dbPath)
	if err != nil {
		fmt.Printf("❌ Error initializing database: %v\n", err)
		return
	}
	defer dbManager.Close()

	// Initialize database schema
	fmt.Println("Initializing database schema...")
	if err := dbManager.InitializeDatabase(); err != nil {
		fmt.Printf("❌ Error initializing database schema: %v\n", err)
		return
	}
	fmt.Println("✅ Database schema initialized successfully")

	// Clear existing data for this production year
	fmt.Printf("Clearing existing data for production year %d...\n", config.ProductionYear)
	if err := dbManager.ClearFinancialDataForYear(config.ProductionYear); err != nil {
		fmt.Printf("❌ Error clearing existing data: %v\n", err)
		return
	}

	// Read wells data for both accounting and production dates
	fmt.Println("Reading wells data for accounting dates...")
	accountingWells, err := ReadWellsData(config.SourceDataPath, config.ProductionYear, AccountingDate)
	if err != nil {
		fmt.Printf("❌ Error reading accounting wells data: %v\n", err)
		return
	}
	fmt.Printf("✅ Read %d wells for accounting dates\n", len(accountingWells))

	fmt.Println("Reading wells data for production dates...")
	productionWells, err := ReadWellsData(config.SourceDataPath, config.ProductionYear, ProductionDate)
	if err != nil {
		fmt.Printf("❌ Error reading production wells data: %v\n", err)
		return
	}
	fmt.Printf("✅ Read %d wells for production dates\n", len(productionWells))

	// Insert operator data
	fmt.Println("Inserting operator data...")
	if err := dbManager.InsertOperator(producer.IDComp, producer); err != nil {
		fmt.Printf("❌ Error inserting operator data: %v\n", err)
		return
	}
	fmt.Println("✅ Operator data inserted")

	// Create a map of wells by ID for easy lookup
	accountingWellsMap := make(map[string]*WellInfo)
	for _, well := range accountingWells {
		accountingWellsMap[well.WellID] = well
	}

	productionWellsMap := make(map[string]*WellInfo)
	for _, well := range productionWells {
		productionWellsMap[well.WellID] = well
	}

	// Get all unique well IDs
	allWellIDs := make(map[string]bool)
	for _, well := range accountingWells {
		allWellIDs[well.WellID] = true
	}
	for _, well := range productionWells {
		allWellIDs[well.WellID] = true
	}

	// Insert well data and financial data
	fmt.Println("Inserting well data and financial data...")
	wellCount := 0
	for wellID := range allWellIDs {
		// Get well data (use accounting well for basic info, fallback to production)
		var wellData *WellInfo
		if accountingWell, exists := accountingWellsMap[wellID]; exists {
			wellData = accountingWell
		} else if productionWell, exists := productionWellsMap[wellID]; exists {
			wellData = productionWell
		} else {
			continue
		}

		// Insert well data
		if err := dbManager.InsertWell(wellData, producer.IDComp); err != nil {
			fmt.Printf("❌ Error inserting well %s: %v\n", wellID, err)
			continue
		}

		// Get accounting and production data for this well
		accountingWell := accountingWellsMap[wellID]
		productionWell := productionWellsMap[wellID]

		// Insert financial data for both accounting and production
		if err := dbManager.InsertWellFinancialData(wellID, config.ProductionYear, accountingWell, productionWell); err != nil {
			fmt.Printf("❌ Error inserting financial data for well %s: %v\n", wellID, err)
			continue
		}

		wellCount++
	}

	fmt.Printf("✅ Inserted %d wells with financial data\n", wellCount)

	fmt.Printf("\n✅ Database operations completed successfully!\n")
	fmt.Printf("Database location: %s\n", dbPath)
	fmt.Printf("Total wells processed: %d\n", wellCount)
	fmt.Printf("Production year: %d\n", config.ProductionYear)
	fmt.Printf("Data types: Accounting dates (%d wells), Production dates (%d wells)\n", len(accountingWells), len(productionWells))
}

func main() {
	fmt.Printf("===============================================\n")
	fmt.Printf("WV ANNUAL RETURN - PRODUCER INFORMATION\n")
	fmt.Printf("===============================================\n\n")

	// Get configuration from user
	config, err := PromptForConfig()
	if err != nil {
		log.Fatalf("Failed to get configuration: %v", err)
	}

	// Check if source data directory exists
	if _, err := os.Stat(config.SourceDataPath); os.IsNotExist(err) {
		fmt.Printf("ERROR: Source data directory does not exist: %s\n", config.SourceDataPath)
		fmt.Printf("Please create the directory and add your DBF files.\n")
		os.Exit(1)
	}

	// List files in source data directory
	fmt.Printf("\nChecking source data directory...\n")
	files, err := os.ReadDir(config.SourceDataPath)
	if err != nil {
		log.Fatalf("Failed to read source data directory: %v", err)
	}

	fmt.Printf("Files found in %s:\n", config.SourceDataPath)
	for _, file := range files {
		fmt.Printf("  %s\n", file.Name())
	}

	// Read producer information
	producer, err := ReadProducerInfo(config.SourceDataPath)
	if err != nil {
		log.Fatalf("Failed to read producer information: %v", err)
	}

	// Resolve producer code using lookup
	fmt.Printf("\n=== Producer Code Resolution ===\n")
	resolvedCode, err := resolveProducerCode(producer)
	if err != nil {
		log.Fatalf("Failed to resolve producer code: %v", err)
	}

	// Update producer with resolved code
	producer.IDComp = resolvedCode

	// Display the information
	PrintProducerInfo(producer, config)

	// Check if database exists and has data
	dbPath := filepath.Join("sourcedata", "sql", "wv_operator.db")
	var wells []*WellInfo
	var dbManager *DatabaseManager

	if _, err := os.Stat(dbPath); err == nil {
		// Database exists, try to use it
		dbManager, err = NewDatabaseManager(dbPath)
		if err == nil {
			defer dbManager.Close()

			hasData, err := dbManager.CheckDatabaseExists()
			if err == nil && hasData {
				fmt.Println("✅ Found existing database with data. Using database for PDF generation.")

				// Read wells from database
				wells, err = dbManager.GetAllWellsForYear(config.ProductionYear)
				if err != nil {
					fmt.Printf("❌ Error reading wells from database: %v\n", err)
					fmt.Println("Falling back to DBF files...")
				} else {
					fmt.Printf("✅ Loaded %d wells from database\n", len(wells))
				}
			}
		}
	}

	// If no database data, fall back to DBF files
	if len(wells) == 0 {
		fmt.Println("Reading wells data from DBF files...")
		wells, err = ReadWellsData(config.SourceDataPath, config.ProductionYear, config.DateType)
		if err != nil {
			fmt.Printf("❌ Error reading wells data: %v\n", err)
			return
		}
		fmt.Printf("✅ Loaded %d wells from DBF files\n", len(wells))
	}

	// Handle data analysis mode
	if config.ReportType == AnalyzeData {
		handleDataAnalysisMode(wells, config, producer)
		return
	}

	// Handle database operations mode
	if config.ReportType == DatabaseOperations {
		handleDatabaseOperations(config, producer)
		return
	}

	// Generate individual PDFs for each well
	fmt.Printf("\nGenerating individual well PDFs...\n")
	templatePath := filepath.Join("template", "stc1235.page1.pdf")
	outputDir := "output/well_pdfs" // Directory for individual well PDFs

	// Generate individual PDFs for each well
	err = GenerateWellPDFs(producer, config, wells, templatePath, outputDir)
	if err != nil {
		fmt.Printf("❌ Error generating well PDFs: %v\n", err)
		return
	}

	// Print sample wells data
	fmt.Printf("\nSample wells data (first 5 wells):\n")
	for i, well := range wells {
		if i >= 5 {
			break
		}
		// Format acreage display
		acreageDisplay := "MISSING"
		if well.LandAcreage > 0 {
			acreageDisplay = fmt.Sprintf("%.2f", well.LandAcreage)
		}

		fmt.Printf("  Well %d: %s (%s) - County: %s (%s) - NRA: %s - API: %s - Acres: %s\n",
			i+1, well.WellName, well.WellID, well.County, well.CountyCode, well.NRA, well.API, acreageDisplay)
	}

	fmt.Printf("COMPLETED SUCCESSFULLY\n")
	fmt.Printf("===============================================\n")
}
