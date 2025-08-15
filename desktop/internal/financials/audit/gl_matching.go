package audit

import (
	"fmt"
	"strings"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/company"
)

// CheckGLMatching audits check and GL entry matching for a specific period
func (s *Service) CheckGLMatching(companyName, accountNumber, startDate, endDate string) (*GLMatchingAudit, error) {
	// Parse dates
	var startDt, endDt time.Time
	var err error
	
	if startDate != "" {
		startDt, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			startDt = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
		}
	} else {
		startDt = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	
	if endDate != "" {
		endDt, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			endDt = time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)
		}
	} else {
		endDt = time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC)
	}

	// Read checks.dbf
	checksData, err := company.ReadDBFFile(companyName, "CHECKS.DBF", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read CHECKS.DBF: %w", err)
	}

	checks, ok := checksData["rows"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid checks data format")
	}

	// Read GLMASTER.dbf
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.DBF", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read GLMASTER.DBF: %w", err)
	}

	glEntries, ok := glData["rows"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid GL data format")
	}

	result := &GLMatchingAudit{
		Period:             fmt.Sprintf("%s to %s", startDate, endDate),
		UnmatchedChecks:    []map[string]interface{}{},
		UnmatchedGLEntries: []map[string]interface{}{},
	}

	// Filter checks by date and account
	filteredChecks := []map[string]interface{}{}
	for _, check := range checks {
		// Skip void checks
		if getBool(check, "LVOID") {
			continue
		}

		// Filter by account if specified
		if accountNumber != "" {
			checkAcct := getString(check, "CACCTNO")
			if checkAcct != accountNumber {
				continue
			}
		}

		// Filter by date
		if checkDateVal := check["DCHECKDATE"]; checkDateVal != nil {
			checkDate, err := parseDate(checkDateVal)
			if err == nil && checkDate.After(startDt) && checkDate.Before(endDt) {
				filteredChecks = append(filteredChecks, check)
			}
		}
	}

	result.TotalChecks = len(filteredChecks)

	// Filter GL entries by date and account
	filteredGL := []map[string]interface{}{}
	for _, gl := range glEntries {
		// Filter by source (CD = Cash Disbursement)
		source := getString(gl, "CSOURCE")
		if source != "CD" && source != "AP" {
			continue
		}

		// Filter by account if specified
		if accountNumber != "" {
			glAcct := getString(gl, "CACCTNO")
			if glAcct != accountNumber {
				continue
			}
		}

		// Filter by date
		if glDateVal := gl["DTRANSDATE"]; glDateVal != nil {
			glDate, err := parseDate(glDateVal)
			if err == nil && glDate.After(startDt) && glDate.Before(endDt) {
				filteredGL = append(filteredGL, gl)
			}
		}
	}

	result.TotalGLEntries = len(filteredGL)

	// Create maps for matching
	checksByNumber := make(map[string]map[string]interface{})
	checksByAmount := make(map[float64][]map[string]interface{})
	
	for _, check := range filteredChecks {
		checkNum := getString(check, "CCHECKNO")
		amount := parseAmount(check["NAMOUNT"])
		
		checksByNumber[checkNum] = check
		checksByAmount[amount] = append(checksByAmount[amount], check)
	}

	// Match GL entries to checks
	matchedChecks := make(map[string]bool)
	matchedGL := make(map[int]bool)

	for idx, gl := range filteredGL {
		glAmount := parseAmount(gl["NAMOUNT"])
		glDesc := getString(gl, "CDESCRIPT")
		glBatch := getString(gl, "CBATCH")
		
		matched := false
		
		// Try to match by check number in description
		for checkNum := range checksByNumber {
			if checkNum != "" && strings.Contains(glDesc, checkNum) {
				check := checksByNumber[checkNum]
				checkAmount := parseAmount(check["NAMOUNT"])
				if checkAmount == glAmount {
					matchedChecks[checkNum] = true
					matchedGL[idx] = true
					matched = true
					break
				}
			}
		}
		
		// If not matched, try by amount
		if !matched {
			if checksWithAmount, found := checksByAmount[glAmount]; found {
				for _, check := range checksWithAmount {
					checkNum := getString(check, "CCHECKNO")
					if !matchedChecks[checkNum] {
						// Try additional matching criteria
						checkBatch := getString(check, "CBATCH")
						if glBatch != "" && checkBatch == glBatch {
							matchedChecks[checkNum] = true
							matchedGL[idx] = true
							matched = true
							break
						}
					}
				}
			}
		}
	}

	// Find unmatched checks
	for _, check := range filteredChecks {
		checkNum := getString(check, "CCHECKNO")
		if !matchedChecks[checkNum] {
			result.UnmatchedChecks = append(result.UnmatchedChecks, map[string]interface{}{
				"checkNumber": checkNum,
				"checkDate":   check["DCHECKDATE"],
				"payee":       getString(check, "CPAYEE"),
				"amount":      parseAmount(check["NAMOUNT"]),
				"accountNo":   getString(check, "CACCTNO"),
				"batch":       getString(check, "CBATCH"),
			})
		}
	}

	// Find unmatched GL entries
	for idx, gl := range filteredGL {
		if !matchedGL[idx] {
			result.UnmatchedGLEntries = append(result.UnmatchedGLEntries, map[string]interface{}{
				"transDate":   gl["DTRANSDATE"],
				"description": getString(gl, "CDESCRIPT"),
				"amount":      parseAmount(gl["NAMOUNT"]),
				"accountNo":   getString(gl, "CACCTNO"),
				"batch":       getString(gl, "CBATCH"),
				"source":      getString(gl, "CSOURCE"),
			})
		}
	}

	result.MatchedChecks = len(matchedChecks)
	result.MatchedGLEntries = len(matchedGL)

	return result, nil
}