package audit

import (
	"fmt"
	"math"

	"github.com/pivoten/financialsx/desktop/internal/company"
)

// VoidChecks audits void check integrity
func (s *Service) VoidChecks(companyName string) (*VoidCheckAudit, error) {
	// Read checks.dbf
	checksData, err := company.ReadDBFFile(companyName, "CHECKS.DBF", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read CHECKS.DBF: %w", err)
	}

	checks, ok := checksData["rows"].([]map[string]interface{})
	if !ok || len(checks) == 0 {
		return &VoidCheckAudit{
			TotalVoidChecks: 0,
			ProperlyVoided:  0,
			Issues:          []map[string]interface{}{},
		}, nil
	}

	result := &VoidCheckAudit{
		Issues:          []map[string]interface{}{},
		VoidWithNonZero: []map[string]interface{}{},
		NonVoidWithZero: []map[string]interface{}{},
		VoidButCleared:  []map[string]interface{}{},
	}

	for _, check := range checks {
		isVoid := getBool(check, "LVOID")
		isCleared := getBool(check, "LCLEARED")
		amount := parseAmount(check["NAMOUNT"])
		checkNum := getString(check, "CCHECKNO")
		payee := getString(check, "CPAYEE")

		if isVoid {
			result.TotalVoidChecks++

			// Check 1: Void checks should have zero amount
			if math.Abs(amount) > 0.001 { // Using small epsilon for float comparison
				issue := map[string]interface{}{
					"type":        "void_with_amount",
					"checkNumber": checkNum,
					"amount":      amount,
					"payee":       payee,
					"issue":       "Void check has non-zero amount",
					"rowData":     check,
				}
				result.VoidWithNonZero = append(result.VoidWithNonZero, issue)
				result.Issues = append(result.Issues, issue)
			}

			// Check 2: Void checks should not be cleared
			if isCleared {
				issue := map[string]interface{}{
					"type":        "void_but_cleared",
					"checkNumber": checkNum,
					"payee":       payee,
					"issue":       "Check is marked as both void and cleared",
					"rowData":     check,
				}
				result.VoidButCleared = append(result.VoidButCleared, issue)
				result.Issues = append(result.Issues, issue)
			}

			// If no issues, it's properly voided
			if math.Abs(amount) <= 0.001 && !isCleared {
				result.ProperlyVoided++
			}
		} else {
			// Check 3: Non-void checks with zero amount (potential data issue)
			if math.Abs(amount) <= 0.001 && checkNum != "" {
				issue := map[string]interface{}{
					"type":        "zero_amount_not_void",
					"checkNumber": checkNum,
					"payee":       payee,
					"cleared":     isCleared,
					"issue":       "Check has zero amount but is not marked as void",
					"rowData":     check,
				}
				result.NonVoidWithZero = append(result.NonVoidWithZero, issue)
				result.Issues = append(result.Issues, issue)
			}
		}
	}

	return result, nil
}

// VoidChecksAuditResult converts the VoidCheckAudit to a standard AuditResult
func (s *Service) VoidChecksAuditResult(companyName string) (*AuditResult, error) {
	audit, err := s.VoidChecks(companyName)
	if err != nil {
		return nil, err
	}

	issues := []AuditIssue{}

	// Convert void with non-zero amount issues
	for _, issue := range audit.VoidWithNonZero {
		issues = append(issues, AuditIssue{
			Type:        "void_with_amount",
			Severity:    "error",
			Description: fmt.Sprintf("Void check %s has amount $%.2f", issue["checkNumber"], issue["amount"]),
			Details:     issue,
		})
	}

	// Convert non-void with zero amount issues
	for _, issue := range audit.NonVoidWithZero {
		issues = append(issues, AuditIssue{
			Type:        "zero_amount_not_void",
			Severity:    "warning",
			Description: fmt.Sprintf("Check %s has zero amount but not marked void", issue["checkNumber"]),
			Details:     issue,
		})
	}

	// Convert void but cleared issues
	for _, issue := range audit.VoidButCleared {
		issues = append(issues, AuditIssue{
			Type:        "void_but_cleared",
			Severity:    "error",
			Description: fmt.Sprintf("Check %s is both void and cleared", issue["checkNumber"]),
			Details:     issue,
		})
	}

	summary := map[string]interface{}{
		"totalVoidChecks":       audit.TotalVoidChecks,
		"properlyVoided":        audit.ProperlyVoided,
		"voidWithNonZeroAmount": len(audit.VoidWithNonZero),
		"nonVoidWithZeroAmount": len(audit.NonVoidWithZero),
		"voidButCleared":        len(audit.VoidButCleared),
		"totalIssues":           len(audit.Issues),
	}

	var message string
	if len(audit.Issues) == 0 {
		message = fmt.Sprintf("All %d void checks are properly configured", audit.TotalVoidChecks)
	} else {
		message = fmt.Sprintf("Found %d issues with void checks", len(audit.Issues))
	}

	return &AuditResult{
		Success:     true,
		Message:     message,
		TotalChecks: audit.TotalVoidChecks,
		Issues:      issues,
		Summary:     summary,
	}, nil
}