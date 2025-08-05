package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// FormField represents a form field in the JSON structure
type FormField struct {
	Pages     []int       `json:"pages"`
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Value     interface{} `json:"value"` // Can be string or bool
	Multiline bool        `json:"multiline"`
	Locked    bool        `json:"locked"`
}

// FormData represents the complete form data structure
type FormData struct {
	Header struct {
		Source   string `json:"source"`
		Version  string `json:"version"`
		Creation string `json:"creation"`
		Title    string `json:"title"`
		Author   string `json:"author"`
		Creator  string `json:"creator"`
		Producer string `json:"producer"`
	} `json:"header"`
	Forms []struct {
		Textfield []FormField `json:"textfield"`
		Checkbox  []FormField `json:"checkbox"`
	} `json:"forms"`
}

// FillWVForm fills the West Virginia Oil and Gas Producer/Operator Return form using pdfcpu
func FillWVForm(producer *ProducerInfo, config *WVConfig, well *WellInfo, templatePath, outputPath string) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create form data structure
	formData := FormData{}
	formData.Header.Source = "stc1235.page1.pdf"
	formData.Header.Version = "pdfcpu v0.11.0 dev"
	formData.Header.Creation = "2025-07-29 15:07:12 MDT"
	formData.Header.Title = "STC 1235 2021.07.07.indd"
	formData.Header.Author = "E022366"
	formData.Header.Creator = "PScript5.dll Version 5.2.2"
	formData.Header.Producer = "macOS Version 26.0 (Build 25A5316i) Quartz PDFContext, AppendMode 1.1"

	// Format acreage for display
	landAcreageStr := ""
	leaseAcreageStr := ""
	if well.LandAcreage > 0 {
		landAcreageStr = fmt.Sprintf("%.2f", well.LandAcreage)
		leaseAcreageStr = fmt.Sprintf("%.2f", well.LeaseAcreage)
	}

	// Calculate adjusted working interest and royalty values for each resource type
	wiOil, riOil := adjustWorkingInterestForRounding(well.OilRevenue, well.OilRevenue*(well.WorkingOilInterest/100), well.OilRevenue*((well.RoyaltyOilInterest+well.OverridingOilInterest)/100))
	wiGas, riGas := adjustWorkingInterestForRounding(well.GasRevenue, well.GasRevenue*(well.WorkingGasInterest/100), well.GasRevenue*((well.RoyaltyGasInterest+well.OverridingGasInterest)/100))
	wiNgls, riNgls := adjustWorkingInterestForRounding(well.OtherRevenue, well.OtherRevenue*(well.WorkingOtherInterest/100), well.OtherRevenue*((well.RoyaltyOtherInterest+well.OverridingOtherInterest)/100))

	// Create form fields with producer data and well-specific data
	formData.Forms = []struct {
		Textfield []FormField `json:"textfield"`
		Checkbox  []FormField `json:"checkbox"`
	}{
		{
			Textfield: []FormField{
				// Section 1 - Producer Information
				{Pages: []int{1}, ID: "89", Name: "producerName", Value: producer.CompanyName, Multiline: false, Locked: false},                        // Producer Name
				{Pages: []int{1}, ID: "345", Name: "producerCode", Value: producer.IDComp, Multiline: false, Locked: false},                            // Producer Code
				{Pages: []int{1}, ID: "388", Name: "address1", Value: producer.Address1, Multiline: false, Locked: false},                              // Address
				{Pages: []int{1}, ID: "399", Name: "city", Value: producer.City, Multiline: false, Locked: false},                                      // City
				{Pages: []int{1}, ID: "410", Name: "state", Value: producer.State, Multiline: false, Locked: false},                                    // State
				{Pages: []int{1}, ID: "421", Name: "zipcode", Value: producer.ZipCode, Multiline: false, Locked: false},                                // Zip Code
				{Pages: []int{1}, ID: "443", Name: "phone", Value: producer.PhoneNo, Multiline: false, Locked: false},                                  // Phone
				{Pages: []int{1}, ID: "454", Name: "email", Value: producer.Email, Multiline: false, Locked: false},                                    // Email
				{Pages: []int{1}, ID: "432", Name: "agent", Value: producer.AgentName, Multiline: false, Locked: false},                                // Agent
				{Pages: []int{1}, ID: "465", Name: "reportingYear", Value: fmt.Sprintf("%02d", config.ReportingYear), Multiline: false, Locked: false}, // Reporting Year (2-digit format)

				// Schedule 1 - Well Information
				{Pages: []int{1}, ID: "2246", Name: "countyName", Value: well.County, Multiline: false, Locked: false},       // County Name
				{Pages: []int{1}, ID: "2257", Name: "countyNumber", Value: well.CountyCode, Multiline: false, Locked: false}, // County Number
				{Pages: []int{1}, ID: "2268", Name: "nraNumber", Value: well.NRA, Multiline: false, Locked: false},           // NRA Number
				{Pages: []int{1}, ID: "2279", Name: "apiNumber", Value: well.API, Multiline: false, Locked: false},           // API Number
				{Pages: []int{1}, ID: "2290", Name: "wellName", Value: well.WellName, Multiline: false, Locked: false},       // Well Name
				{Pages: []int{1}, ID: "2301", Name: "landAcreage", Value: landAcreageStr, Multiline: false, Locked: false},   // Land Acreage
				{Pages: []int{1}, ID: "2312", Name: "leaseAcreage", Value: leaseAcreageStr, Multiline: false, Locked: false}, // Lease Acreage

				// Gas Type Checkboxes
				{Pages: []int{1}, ID: "2745", Name: "gas_ethane", Value: getBoolValue(well.HasEthane), Multiline: false, Locked: false},       // Ethane
				{Pages: []int{1}, ID: "2756", Name: "gas_propane", Value: getBoolValue(well.HasPropane), Multiline: false, Locked: false},     // Propane
				{Pages: []int{1}, ID: "2767", Name: "gas_butane", Value: getBoolValue(well.HasButane), Multiline: false, Locked: false},       // Butane
				{Pages: []int{1}, ID: "2778", Name: "gas_isobutane", Value: getBoolValue(well.HasIsobutane), Multiline: false, Locked: false}, // Isobutane
				{Pages: []int{1}, ID: "2789", Name: "gas_pentane", Value: getBoolValue(well.HasPentane), Multiline: false, Locked: false},     // Pentane

				// Production Information
				{Pages: []int{1}, ID: "2801", Name: "intialProduction", Value: formatProductionDate(well.ProductionDate), Multiline: false, Locked: false}, // Initial Production
				{Pages: []int{1}, ID: "2822", Name: "formations", Value: well.Formation, Multiline: false, Locked: false},                                  // Formations (raw DBF value)

				// Production Totals
				{Pages: []int{1}, ID: "3003", Name: "production_totalBBL", Value: fmt.Sprintf("%.0f", roundToWholeNumber(well.TotalOilBBL)), Multiline: false, Locked: false}, // Total Oil BBL
				{Pages: []int{1}, ID: "3025", Name: "production_totalGAS", Value: fmt.Sprintf("%.0f", roundToWholeNumber(well.TotalGasMCF)), Multiline: false, Locked: false}, // Total Gas MCF
				{Pages: []int{1}, ID: "3066", Name: "production_totalNGLS", Value: fmt.Sprintf("%.0f", roundToWholeNumber(well.TotalNGLS)), Multiline: false, Locked: false},  // Total NGLS (NTOTPROD)

				// Revenue Fields
				{Pages: []int{1}, ID: "3078", Name: "revenue_oil", Value: fmt.Sprintf("%.0f", roundToWholeNumber(well.OilRevenue)), Multiline: false, Locked: false},    // Oil Revenue
				{Pages: []int{1}, ID: "3089", Name: "revenue_gas", Value: fmt.Sprintf("%.0f", roundToWholeNumber(well.GasRevenue)), Multiline: false, Locked: false},    // Gas Revenue
				{Pages: []int{1}, ID: "3100", Name: "revenue_ngls", Value: fmt.Sprintf("%.0f", roundToWholeNumber(well.OtherRevenue)), Multiline: false, Locked: false}, // NGL Revenue (NOTHERINC)

				// Working Interest Revenue (calculated from WELLINV interest groups) - with rounding adjustment
				{Pages: []int{1}, ID: "3111", Name: "wi_oilRevenue", Value: fmt.Sprintf("%.0f", wiOil), Multiline: false, Locked: false},   // WI Oil Revenue = adjusted to match total
				{Pages: []int{1}, ID: "3122", Name: "wi_gasRevenue", Value: fmt.Sprintf("%.0f", wiGas), Multiline: false, Locked: false},   // WI Gas Revenue = adjusted to match total
				{Pages: []int{1}, ID: "3133", Name: "wi_nglsRevenue", Value: fmt.Sprintf("%.0f", wiNgls), Multiline: false, Locked: false}, // WI NGL Revenue = adjusted to match total

				// Expenses
				{Pages: []int{1}, ID: "3144", Name: "expenses_oil", Value: fmt.Sprintf("%.0f", roundToWholeNumber(calculateExpenseByType(well.TotalExpenses, well.OilRevenue, well.GasRevenue, well.OtherRevenue, "oil"))), Multiline: false, Locked: false},   // Oil Expenses (calculated)
				{Pages: []int{1}, ID: "3155", Name: "expenses_gas", Value: fmt.Sprintf("%.0f", roundToWholeNumber(calculateExpenseByType(well.TotalExpenses, well.OilRevenue, well.GasRevenue, well.OtherRevenue, "gas"))), Multiline: false, Locked: false},   // Gas Expenses (calculated)
				{Pages: []int{1}, ID: "3166", Name: "expenses_ngls", Value: fmt.Sprintf("%.0f", roundToWholeNumber(calculateExpenseByType(well.TotalExpenses, well.OilRevenue, well.GasRevenue, well.OtherRevenue, "ngls"))), Multiline: false, Locked: false}, // NGL Expenses (calculated)

				// Royalty Revenue (calculated from WELLINV interest groups - Royalty + Overriding Royalty) - with rounding adjustment
				{Pages: []int{1}, ID: "3177", Name: "royalty_oilRevenue", Value: fmt.Sprintf("%.0f", riOil), Multiline: false, Locked: false},   // Royalty Oil Revenue = adjusted to match total
				{Pages: []int{1}, ID: "3188", Name: "royalty_gasRevenue", Value: fmt.Sprintf("%.0f", riGas), Multiline: false, Locked: false},   // Royalty Gas Revenue = adjusted to match total
				{Pages: []int{1}, ID: "3199", Name: "royalty_nglsRevenue", Value: fmt.Sprintf("%.0f", riNgls), Multiline: false, Locked: false}, // Royalty NGL Revenue = adjusted to match total

				// Totals
				{Pages: []int{1}, ID: "3210", Name: "wi_total", Value: fmt.Sprintf("%.0f", wiOil+wiGas+wiNgls), Multiline: false, Locked: false},      // WI Total = sum of all 3 columns of line 7
				{Pages: []int{1}, ID: "3221", Name: "wi_doiTotal", Value: "", Multiline: false, Locked: false},                                        // WI DOI Total = blank for now
				{Pages: []int{1}, ID: "3232", Name: "royalty_total", Value: fmt.Sprintf("%.0f", riOil+riGas+riNgls), Multiline: false, Locked: false}, // Royalty Total = sum of all 3 columns of line 9
				{Pages: []int{1}, ID: "3243", Name: "royalty_doiTotal", Value: "", Multiline: false, Locked: false},                                   // Royalty DOI Total = blank for now
			},
			Checkbox: []FormField{
				// Well Status Checkboxes
				{Pages: []int{1}, ID: "2354", Name: "status_active", Value: getStatusBoolValue(well.WellStatus, "A"), Multiline: false, Locked: false},              // Active
				{Pages: []int{1}, ID: "2425", Name: "status_plugged", Value: getStatusBoolValue(well.WellStatus, "P"), Multiline: false, Locked: false},             // Plugged
				{Pages: []int{1}, ID: "2436", Name: "status_shutin", Value: getStatusBoolValue(well.WellStatus, "S"), Multiline: false, Locked: false},              // Shut-in
				{Pages: []int{1}, ID: "2447", Name: "status_enhanced", Value: getStatusBoolValue(well.WellStatus, "E"), Multiline: false, Locked: false},            // Enhanced
				{Pages: []int{1}, ID: "2458", Name: "status_horizonal_marcellus", Value: getStatusBoolValue(well.WellStatus, "H"), Multiline: false, Locked: false}, // Horizontal Marcellus (note: typo in PDF field name)
				{Pages: []int{1}, ID: "2469", Name: "status_horizonal_other", Value: getStatusBoolValue(well.WellStatus, "O"), Multiline: false, Locked: false},     // Horizontal Other (note: typo in PDF field name)
				{Pages: []int{1}, ID: "2480", Name: "status_vertical_macellus", Value: getStatusBoolValue(well.WellStatus, "V"), Multiline: false, Locked: false},   // Vertical Marcellus
				{Pages: []int{1}, ID: "2491", Name: "status_cbm", Value: getStatusBoolValue(well.WellStatus, "C"), Multiline: false, Locked: false},                 // CBM
				{Pages: []int{1}, ID: "2502", Name: "status_began_produciton", Value: getStatusBoolValue(well.WellStatus, "B"), Multiline: false, Locked: false},    // Began Production
				{Pages: []int{1}, ID: "2513", Name: "status_flatrate", Value: getStatusBoolValue(well.WellStatus, "F"), Multiline: false, Locked: false},            // Flat Rate
				{Pages: []int{1}, ID: "2524", Name: "status_home", Value: getStatusBoolValue(well.WellStatus, "H"), Multiline: false, Locked: false},                // Home

				// Gas Type Checkboxes
				{Pages: []int{1}, ID: "2745", Name: "gas_ethane", Value: well.HasEthane, Multiline: false, Locked: false},       // Ethane
				{Pages: []int{1}, ID: "2756", Name: "gas_propane", Value: well.HasPropane, Multiline: false, Locked: false},     // Propane
				{Pages: []int{1}, ID: "2767", Name: "gas_butane", Value: well.HasButane, Multiline: false, Locked: false},       // Butane
				{Pages: []int{1}, ID: "2778", Name: "gas_isobutane", Value: well.HasIsobutane, Multiline: false, Locked: false}, // Isobutane
				{Pages: []int{1}, ID: "2789", Name: "gas_pentane", Value: well.HasPentane, Multiline: false, Locked: false},     // Pentane
			},
		},
	}

	// Create temporary JSON file
	tempJSON := "temp_form_data.json"
	jsonData, err := json.MarshalIndent(formData, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	err = os.WriteFile(tempJSON, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file: %v", err)
	}
	defer os.Remove(tempJSON) // Clean up temp file

	// Use pdfcpu to fill the form
	cmd := exec.Command("pdfcpu", "form", "fill", templatePath, tempJSON, outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to fill form with pdfcpu: %v\nOutput: %s", err, string(output))
	}

	fmt.Printf("🎉 SUCCESS! PDF form filled with producer data!\n")
	fmt.Printf("📄 Output: %s\n", outputPath)
	fmt.Printf("✅ Form fields populated with real data\n")

	return nil
}

// formatProductionDate formats the production date for PDF display
func formatProductionDate(dateStr string) string {
	if dateStr == "" || dateStr == "0001-01-01 00:00:00 +0000 UTC" {
		return "" // Return blank for empty or default dates
	}

	// Try to parse the date and format it as MM/DD/YYYY
	if t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", dateStr); err == nil {
		return t.Format("01/02/2006")
	}

	// If parsing fails, return the original string
	return dateStr
}

// GenerateFormFilledPDF - the main function for form filling approach (for single PDF without well data)
func GenerateFormFilledPDF(producer *ProducerInfo, config *WVConfig, templatePath, outputPath string) error {
	// Create a dummy well for backward compatibility
	dummyWell := &WellInfo{
		WellID:       "",
		WellName:     "",
		County:       "",
		CountyCode:   "",
		NRA:          "",
		API:          "",
		LandAcreage:  -1,
		LeaseAcreage: -1,
	}
	return FillWVForm(producer, config, dummyWell, templatePath, outputPath)
}

// GenerateWellPDFs creates individual PDFs for each well
func GenerateWellPDFs(producer *ProducerInfo, config *WVConfig, wells []*WellInfo, templatePath, outputDir string) error {
	startTime := time.Now()
	fmt.Printf("Generating %d individual well PDFs...\n", len(wells))
	fmt.Printf("Started at: %s\n", startTime.Format("2006-01-02 15:04:05"))

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create a progress counter
	processedCount := 0
	totalWells := len(wells)

	for _, well := range wells {
		// Create well-specific filename
		wellID := strings.ReplaceAll(well.WellID, " ", "_")
		operatorName := strings.ReplaceAll(producer.CompanyName, " ", "_")
		filename := fmt.Sprintf("%d_wv_annualReturn_%s_%s.pdf",
			config.ReportingYear, operatorName, wellID)

		outputPath := filepath.Join(outputDir, filename)

		// Generate PDF for this well
		err := FillWVForm(producer, config, well, templatePath, outputPath)
		if err != nil {
			fmt.Printf("❌ Error generating PDF for well %s: %v\n", well.WellID, err)
			continue
		}

		processedCount++

		// Show progress every 100 wells
		if processedCount%100 == 0 || processedCount == totalWells {
			fmt.Printf("Progress: %d/%d wells processed (%.1f%%)\n",
				processedCount, totalWells, float64(processedCount)/float64(totalWells)*100)
		}
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	fmt.Printf("✅ Successfully generated %d well PDFs in: %s\n", processedCount, outputDir)
	fmt.Printf("Finished at: %s\n", endTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total duration: %s\n", duration)
	fmt.Printf("Average time per PDF: %s\n", duration/time.Duration(processedCount))
	return nil
}

// Helper function to get checkbox value for well status
func getStatusValue(wellStatus, expectedStatus string) string {
	if wellStatus == expectedStatus {
		return "Yes" // pdfcpu requires "Yes" for checked checkboxes
	}
	return "No" // pdfcpu requires "No" for unchecked checkboxes
}

// Helper function to get checkbox value for boolean fields
func getBoolValue(value bool) string {
	if value {
		return "Yes" // pdfcpu requires "Yes" for checked checkboxes
	}
	return "No" // pdfcpu requires "No" for unchecked checkboxes
}

// Helper function to get checkbox value for well status (returns bool for checkbox fields)
func getStatusBoolValue(wellStatus, expectedStatus string) bool {
	return wellStatus == expectedStatus
}

// calculateExpenseByType calculates expense allocation based on revenue percentages
func calculateExpenseByType(totalExpenses, oilRevenue, gasRevenue, nglRevenue float64, expenseType string) float64 {
	totalRevenue := oilRevenue + gasRevenue + nglRevenue

	// If no revenue, put all expenses in oil
	if totalRevenue == 0 {
		if expenseType == "oil" {
			return totalExpenses
		}
		return 0.0
	}

	// Calculate percentage based on revenue
	var revenueForType float64
	switch expenseType {
	case "oil":
		revenueForType = oilRevenue
	case "gas":
		revenueForType = gasRevenue
	case "ngls":
		revenueForType = nglRevenue
	default:
		return 0.0
	}

	// Calculate expense as percentage of total revenue
	percentage := revenueForType / totalRevenue
	return totalExpenses * percentage
}

// adjustWorkingInterestForRounding ensures that working interest + royalty interest = total receipts after rounding
func adjustWorkingInterestForRounding(totalReceipts, workingInterest, royaltyInterest float64) (float64, float64) {
	// Round all values to whole numbers
	roundedTotal := roundToWholeNumber(totalReceipts)
	roundedWI := roundToWholeNumber(workingInterest)
	roundedRI := roundToWholeNumber(royaltyInterest)

	// Check if they add up correctly
	if roundedWI+roundedRI == roundedTotal {
		return roundedWI, roundedRI
	}

	// If they don't add up, adjust working interest to make the math work
	// Keep royalty interest as is and adjust working interest
	adjustedWI := roundedTotal - roundedRI

	// Ensure working interest is not negative
	if adjustedWI < 0 {
		adjustedWI = 0
	}

	return adjustedWI, roundedRI
}
