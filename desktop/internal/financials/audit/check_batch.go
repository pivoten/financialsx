package audit

import (
	"fmt"

	"github.com/pivoten/financialsx/desktop/internal/company"
)

// CheckBatches performs a comprehensive audit of check batches
func (s *Service) CheckBatches(companyName string) (*CheckBatchAudit, error) {
	// Read checks.dbf
	checksData, err := company.ReadDBFFile(companyName, "CHECKS.DBF", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read CHECKS.DBF: %w", err)
	}

	checks, ok := checksData["rows"].([]map[string]interface{})
	if !ok || len(checks) == 0 {
		return nil, fmt.Errorf("no check records found")
	}

	// Read GLMASTER.dbf  
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.DBF", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read GLMASTER.DBF: %w", err)
	}

	glEntries, ok := glData["rows"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no GL records found")
	}

	// Get column names from checks
	checkColumns := make([]string, 0)
	if len(checks) > 0 {
		for key := range checks[0] {
			checkColumns = append(checkColumns, key)
		}
	}

	// Build maps for faster lookups
	glByBatch := make(map[string][]map[string]interface{})
	glByCheckNum := make(map[string][]map[string]interface{})
	
	for _, gl := range glEntries {
		// Look for AP source entries that might be checks
		source := getString(gl, "CSOURCE")
		if source == "CD" || source == "AP" {
			batch := getString(gl, "CBATCH")
			if batch != "" {
				glByBatch[batch] = append(glByBatch[batch], gl)
			}
			
			// Also try to match by check number in description
			desc := getString(gl, "CDESCRIPT")
			if desc != "" {
				glByCheckNum[desc] = append(glByCheckNum[desc], gl)
			}
		}
	}

	result := &CheckBatchAudit{
		TotalChecks:       len(checks),
		MissingEntries:    []map[string]interface{}{},
		MismatchedAmounts: []map[string]interface{}{},
		CheckColumns:      checkColumns,
	}

	checksWithBatch := 0
	checksWithoutBatch := 0
	matchedEntries := 0

	// Check if CBATCH field exists
	hasBatchField := false
	if len(checks) > 0 {
		_, hasBatchField = checks[0]["CBATCH"]
	}

	for _, check := range checks {
		// Skip void checks
		if getBool(check, "LVOID") {
			continue
		}

		checkNum := getString(check, "CCHECKNO")
		checkAmount := parseAmount(check["NAMOUNT"])
		
		var batch string
		if hasBatchField {
			batch = getString(check, "CBATCH")
		}

		if batch != "" {
			checksWithBatch++
			
			// Look for matching GL entry by batch
			if glEntries, found := glByBatch[batch]; found {
				matchFound := false
				for _, gl := range glEntries {
					glAmount := parseAmount(gl["NAMOUNT"])
					if glAmount == checkAmount {
						matchedEntries++
						matchFound = true
						break
					}
				}
				
				if !matchFound {
					// Amount mismatch
					result.MismatchedAmounts = append(result.MismatchedAmounts, map[string]interface{}{
						"checkNumber": checkNum,
						"batch":       batch,
						"checkAmount": checkAmount,
						"issue":       "GL entry found but amount doesn't match",
						"rowData":     check,
					})
				}
			} else {
				// No GL entry found for this batch
				result.MissingEntries = append(result.MissingEntries, map[string]interface{}{
					"checkNumber": checkNum,
					"batch":       batch,
					"amount":      checkAmount,
					"issue":       "No GL entry found for batch",
					"rowData":     check,
				})
			}
		} else {
			checksWithoutBatch++
			
			// Try to match by check number
			if glEntries, found := glByCheckNum[checkNum]; found {
				for _, gl := range glEntries {
					glAmount := parseAmount(gl["NAMOUNT"])
					if glAmount == checkAmount {
						matchedEntries++
						break
					}
				}
			} else {
				// No match found at all
				result.MissingEntries = append(result.MissingEntries, map[string]interface{}{
					"checkNumber": checkNum,
					"amount":      checkAmount,
					"issue":       "No batch number and no GL match by check number",
					"rowData":     check,
				})
			}
		}
	}

	result.ChecksWithBatch = checksWithBatch
	result.ChecksWithoutBatch = checksWithoutBatch
	result.MatchedEntries = matchedEntries

	return result, nil
}