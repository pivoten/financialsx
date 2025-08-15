package audit

import (
	"fmt"
	"sort"

	"github.com/pivoten/financialsx/desktop/internal/company"
)

// DuplicateCIDCHEC audits for duplicate CIDCHEC values in checks
func (s *Service) DuplicateCIDCHEC(companyName string) (*AuditResult, error) {
	// Read checks.dbf
	checksData, err := company.ReadDBFFile(companyName, "CHECKS.DBF", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read CHECKS.DBF: %w", err)
	}

	checks, ok := checksData["rows"].([]map[string]interface{})
	if !ok || len(checks) == 0 {
		return &AuditResult{
			Success: true,
			Message: "No check records found",
		}, nil
	}

	// Check if CIDCHEC field exists
	hasCIDCHEC := false
	if len(checks) > 0 {
		_, hasCIDCHEC = checks[0]["CIDCHEC"]
	}

	if !hasCIDCHEC {
		return &AuditResult{
			Success: false,
			Error:   "CIDCHEC field not found in CHECKS.DBF",
		}, nil
	}

	// Group checks by CIDCHEC
	cidchecGroups := make(map[string][]map[string]interface{})
	checksWithoutCID := []map[string]interface{}{}
	totalChecks := 0

	for _, check := range checks {
		// Skip void checks in count
		if !getBool(check, "LVOID") {
			totalChecks++
		}

		cidchec := getString(check, "CIDCHEC")
		if cidchec == "" {
			checksWithoutCID = append(checksWithoutCID, check)
		} else {
			cidchecGroups[cidchec] = append(cidchecGroups[cidchec], check)
		}
	}

	// Find duplicates
	duplicates := []map[string]interface{}{}
	for cidchec, group := range cidchecGroups {
		if len(group) > 1 {
			// Create a summary of the duplicate group
			duplicate := map[string]interface{}{
				"cidchec": cidchec,
				"count":   len(group),
				"checks":  []map[string]interface{}{},
			}

			for _, check := range group {
				checkInfo := map[string]interface{}{
					"checkNumber": getString(check, "CCHECKNO"),
					"checkDate":   check["DCHECKDATE"],
					"payee":       getString(check, "CPAYEE"),
					"amount":      parseAmount(check["NAMOUNT"]),
					"accountNo":   getString(check, "CACCTNO"),
					"batch":       getString(check, "CBATCH"),
					"void":        getBool(check, "LVOID"),
					"cleared":     getBool(check, "LCLEARED"),
				}
				duplicate["checks"] = append(duplicate["checks"].([]map[string]interface{}), checkInfo)
			}

			duplicates = append(duplicates, duplicate)
		}
	}

	// Sort duplicates by count (most duplicates first)
	sort.Slice(duplicates, func(i, j int) bool {
		return duplicates[i]["count"].(int) > duplicates[j]["count"].(int)
	})

	// Build issues list
	issues := []AuditIssue{}
	
	// Add duplicate CIDCHEC issues
	for _, dup := range duplicates {
		issues = append(issues, AuditIssue{
			Type:     "duplicate_cidchec",
			Severity: "error",
			Description: fmt.Sprintf("CIDCHEC %s appears %d times", 
				dup["cidchec"], dup["count"]),
			Details: dup,
		})
	}

	// Add missing CIDCHEC issues
	if len(checksWithoutCID) > 0 {
		for _, check := range checksWithoutCID {
			if !getBool(check, "LVOID") { // Only report non-void checks
				issues = append(issues, AuditIssue{
					Type:     "missing_cidchec",
					Severity: "warning",
					Description: fmt.Sprintf("Check %s has no CIDCHEC value",
						getString(check, "CCHECKNO")),
					Details: map[string]interface{}{
						"checkNumber": getString(check, "CCHECKNO"),
						"checkDate":   check["DCHECKDATE"],
						"payee":       getString(check, "CPAYEE"),
						"amount":      parseAmount(check["NAMOUNT"]),
					},
					RowData: check,
				})
			}
		}
	}

	summary := map[string]interface{}{
		"totalChecks":          totalChecks,
		"uniqueCIDCHECs":       len(cidchecGroups),
		"duplicateCIDCHECs":    len(duplicates),
		"checksWithoutCIDCHEC": len(checksWithoutCID),
	}

	// Calculate total checks affected by duplicates
	affectedChecks := 0
	for _, dup := range duplicates {
		affectedChecks += dup["count"].(int)
	}
	summary["checksAffectedByDuplicates"] = affectedChecks

	result := &AuditResult{
		Success:     true,
		TotalChecks: totalChecks,
		Issues:      issues,
		Summary:     summary,
		Metadata: map[string]interface{}{
			"duplicateGroups": duplicates,
			"auditDate":       fmt.Sprintf("%v", checksData["timestamp"]),
		},
	}

	if len(duplicates) > 0 {
		result.Message = fmt.Sprintf("Found %d duplicate CIDCHEC values affecting %d checks", 
			len(duplicates), affectedChecks)
	} else if len(checksWithoutCID) > 0 {
		result.Message = fmt.Sprintf("No duplicates found, but %d checks have no CIDCHEC", 
			len(checksWithoutCID))
	} else {
		result.Message = "No issues found - all checks have unique CIDCHEC values"
	}

	return result, nil
}