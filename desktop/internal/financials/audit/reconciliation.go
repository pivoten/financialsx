package audit

import (
	"fmt"
	"math"
	"time"

	"github.com/pivoten/financialsx/desktop/internal/company"
)

// BankReconciliation performs a comprehensive bank reconciliation audit
func (s *Service) BankReconciliation(companyName string) (*ReconciliationAudit, error) {
	// Read COA.dbf to get bank accounts
	coaData, err := company.ReadDBFFile(companyName, "COA.DBF", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read COA.DBF: %w", err)
	}

	accounts, ok := coaData["rows"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid COA data format")
	}

	// Filter for bank accounts
	bankAccounts := []map[string]interface{}{}
	for _, account := range accounts {
		if getBool(account, "LBANKACCT") {
			bankAccounts = append(bankAccounts, account)
		}
	}

	// Read CHECKREC.dbf for reconciliation records
	recData, err := company.ReadDBFFile(companyName, "CHECKREC.DBF", "", 0, 0, "", "")
	if err != nil {
		// File might not exist, which is okay
		recData = map[string]interface{}{
			"rows": []map[string]interface{}{},
		}
	}

	reconciliations, _ := recData["rows"].([]map[string]interface{})

	// Read CHECKS.dbf for outstanding checks
	checksData, err := company.ReadDBFFile(companyName, "CHECKS.DBF", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read CHECKS.DBF: %w", err)
	}

	checks, _ := checksData["rows"].([]map[string]interface{})

	// Read GLMASTER.dbf for GL balances
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.DBF", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read GLMASTER.DBF: %w", err)
	}

	glEntries, _ := glData["rows"].([]map[string]interface{})

	result := &ReconciliationAudit{
		TotalAccounts:        len(bankAccounts),
		UnreconciledAccounts: []map[string]interface{}{},
		OutOfBalanceAccounts: []map[string]interface{}{},
		StaleReconciliations: []map[string]interface{}{},
	}

	// Build reconciliation map by account
	lastRecByAccount := make(map[string]map[string]interface{})
	for _, rec := range reconciliations {
		acctNo := getString(rec, "CACCTNO")
		if existing, found := lastRecByAccount[acctNo]; found {
			// Keep the most recent reconciliation
			existingDate, _ := parseDate(existing["DRECDATE"])
			currentDate, _ := parseDate(rec["DRECDATE"])
			if currentDate.After(existingDate) {
				lastRecByAccount[acctNo] = rec
			}
		} else {
			lastRecByAccount[acctNo] = rec
		}
	}

	// Calculate GL balance for each account
	glBalanceByAccount := make(map[string]float64)
	for _, gl := range glEntries {
		acctNo := getString(gl, "CACCTNO")
		amount := parseAmount(gl["NAMOUNT"])
		glBalanceByAccount[acctNo] += amount
	}

	// Calculate outstanding checks for each account
	outstandingByAccount := make(map[string]float64)
	outstandingCountByAccount := make(map[string]int)
	for _, check := range checks {
		if !getBool(check, "LCLEARED") && !getBool(check, "LVOID") {
			acctNo := getString(check, "CACCTNO")
			amount := parseAmount(check["NAMOUNT"])
			outstandingByAccount[acctNo] += amount
			outstandingCountByAccount[acctNo]++
		}
	}

	// Analyze each bank account
	for _, account := range bankAccounts {
		acctNo := getString(account, "CACCTNO")
		acctName := getString(account, "CACCTDESC")
		
		// Get GL balance
		glBalance := glBalanceByAccount[acctNo]
		
		// Get outstanding checks
		outstandingAmount := outstandingByAccount[acctNo]
		outstandingCount := outstandingCountByAccount[acctNo]
		
		// Calculate bank balance (GL + Outstanding)
		bankBalance := glBalance + outstandingAmount
		
		// Check for reconciliation record
		if lastRec, hasRec := lastRecByAccount[acctNo]; hasRec {
			// Check if reconciliation is stale (>90 days)
			recDate, _ := parseDate(lastRec["DRECDATE"])
			daysSince := int(time.Since(recDate).Hours() / 24)
			
			if daysSince > 90 {
				result.StaleReconciliations = append(result.StaleReconciliations, map[string]interface{}{
					"accountNo":   acctNo,
					"accountName": acctName,
					"lastRecDate": recDate.Format("2006-01-02"),
					"daysSince":   daysSince,
					"glBalance":   glBalance,
					"bankBalance": bankBalance,
				})
			}
			
			// Check if reconciliation is out of balance
			stmtBalance := parseAmount(lastRec["NSTMTBAL"])
			recBalance := parseAmount(lastRec["NRECBAL"])
			difference := math.Abs(stmtBalance - recBalance)
			
			if difference > 0.01 { // Allow for small rounding differences
				result.OutOfBalanceAccounts = append(result.OutOfBalanceAccounts, map[string]interface{}{
					"accountNo":       acctNo,
					"accountName":     acctName,
					"statementBalance": stmtBalance,
					"reconciledBalance": recBalance,
					"difference":      difference,
					"lastRecDate":     recDate.Format("2006-01-02"),
				})
			}
			
			result.ReconciledAccounts++
		} else {
			// No reconciliation record found
			result.UnreconciledAccounts = append(result.UnreconciledAccounts, map[string]interface{}{
				"accountNo":         acctNo,
				"accountName":       acctName,
				"glBalance":         glBalance,
				"bankBalance":       bankBalance,
				"outstandingChecks": outstandingCount,
				"outstandingAmount": outstandingAmount,
			})
		}
	}

	return result, nil
}

// SingleBankAccount performs an audit on a single bank account
func (s *Service) SingleBankAccount(companyName, accountNumber string) (*AuditResult, error) {
	// Read CHECKS.dbf
	checksData, err := company.ReadDBFFile(companyName, "CHECKS.DBF", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read CHECKS.DBF: %w", err)
	}

	checks, _ := checksData["rows"].([]map[string]interface{})

	// Read GLMASTER.dbf
	glData, err := company.ReadDBFFile(companyName, "GLMASTER.DBF", "", 0, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to read GLMASTER.DBF: %w", err)
	}

	glEntries, _ := glData["rows"].([]map[string]interface{})

	// Filter checks for this account
	accountChecks := []map[string]interface{}{}
	outstandingChecks := []map[string]interface{}{}
	clearedChecks := []map[string]interface{}{}
	voidChecks := []map[string]interface{}{}
	
	var totalCheckAmount, outstandingAmount, clearedAmount float64
	
	for _, check := range checks {
		if getString(check, "CACCTNO") == accountNumber {
			accountChecks = append(accountChecks, check)
			amount := parseAmount(check["NAMOUNT"])
			
			if getBool(check, "LVOID") {
				voidChecks = append(voidChecks, check)
			} else if getBool(check, "LCLEARED") {
				clearedChecks = append(clearedChecks, check)
				clearedAmount += amount
			} else {
				outstandingChecks = append(outstandingChecks, check)
				outstandingAmount += amount
			}
			
			if !getBool(check, "LVOID") {
				totalCheckAmount += amount
			}
		}
	}

	// Calculate GL balance for this account
	var glBalance float64
	glTransactions := 0
	for _, gl := range glEntries {
		if getString(gl, "CACCTNO") == accountNumber {
			glBalance += parseAmount(gl["NAMOUNT"])
			glTransactions++
		}
	}

	// Calculate bank balance
	bankBalance := glBalance + outstandingAmount

	// Build issues list
	issues := []AuditIssue{}
	
	// Check for stale outstanding checks (>90 days)
	staleChecks := 0
	for _, check := range outstandingChecks {
		if dateVal := check["DCHECKDATE"]; dateVal != nil {
			checkDate, err := parseDate(dateVal)
			if err == nil {
				daysSince := int(time.Since(checkDate).Hours() / 24)
				if daysSince > 90 {
					staleChecks++
					issues = append(issues, AuditIssue{
						Type:     "stale_check",
						Severity: "warning",
						Description: fmt.Sprintf("Check %s outstanding for %d days",
							getString(check, "CCHECKNO"), daysSince),
						Details: map[string]interface{}{
							"checkNumber": getString(check, "CCHECKNO"),
							"checkDate":   checkDate.Format("2006-01-02"),
							"payee":       getString(check, "CPAYEE"),
							"amount":      parseAmount(check["NAMOUNT"]),
							"daysSince":   daysSince,
						},
					})
				}
			}
		}
	}

	// Check for duplicate check numbers
	checkNumbers := make(map[string]int)
	for _, check := range accountChecks {
		if !getBool(check, "LVOID") {
			checkNum := getString(check, "CCHECKNO")
			if checkNum != "" {
				checkNumbers[checkNum]++
			}
		}
	}
	
	for checkNum, count := range checkNumbers {
		if count > 1 {
			issues = append(issues, AuditIssue{
				Type:     "duplicate_check_number",
				Severity: "error",
				Description: fmt.Sprintf("Check number %s appears %d times", checkNum, count),
				Details: map[string]interface{}{
					"checkNumber": checkNum,
					"count":       count,
				},
			})
		}
	}

	summary := map[string]interface{}{
		"accountNumber":      accountNumber,
		"totalChecks":        len(accountChecks),
		"outstandingChecks":  len(outstandingChecks),
		"clearedChecks":      len(clearedChecks),
		"voidChecks":         len(voidChecks),
		"staleChecks":        staleChecks,
		"totalCheckAmount":   totalCheckAmount,
		"outstandingAmount":  outstandingAmount,
		"clearedAmount":      clearedAmount,
		"glBalance":          glBalance,
		"bankBalance":        bankBalance,
		"glTransactions":     glTransactions,
	}

	message := fmt.Sprintf("Account %s: %d total checks, %d outstanding ($%.2f), GL Balance: $%.2f, Bank Balance: $%.2f",
		accountNumber, len(accountChecks), len(outstandingChecks), outstandingAmount, glBalance, bankBalance)

	return &AuditResult{
		Success:     true,
		Message:     message,
		TotalChecks: len(accountChecks),
		Issues:      issues,
		Summary:     summary,
	}, nil
}