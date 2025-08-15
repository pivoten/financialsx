package main

import (
	"fmt"
)

// ============================================================================
// AUDIT WRAPPER FUNCTIONS
// These functions wrap the audit service methods for backward compatibility
// The actual implementations are in internal/financials/audit package
// ============================================================================

// AuditCheckBatches wraps the audit service CheckBatches method
func (a *App) AuditCheckBatches(companyName string) (map[string]interface{}, error) {
	if a.auditService == nil {
		return nil, fmt.Errorf("audit service not initialized")
	}
	
	audit, err := a.auditService.CheckBatches(companyName)
	if err != nil {
		return nil, err
	}
	
	// Convert to map for backward compatibility
	return map[string]interface{}{
		"status":            "completed",
		"totalChecks":       audit.TotalChecks,
		"checksWithBatch":   audit.ChecksWithBatch,
		"checksWithoutBatch": audit.ChecksWithoutBatch,
		"matchedEntries":    audit.MatchedEntries,
		"missing_entries":   audit.MissingEntries,
		"mismatched_amounts": audit.MismatchedAmounts,
		"check_columns":     audit.CheckColumns,
		"summary": map[string]interface{}{
			"total_checks":       audit.TotalChecks,
			"matched_entries":    audit.MatchedEntries,
			"missing_entries":    len(audit.MissingEntries),
			"mismatched_amounts": len(audit.MismatchedAmounts),
		},
	}, nil
}

// AuditDuplicateCIDCHEC wraps the audit service DuplicateCIDCHEC method
func (a *App) AuditDuplicateCIDCHEC(companyName string) (map[string]interface{}, error) {
	if a.auditService == nil {
		return nil, fmt.Errorf("audit service not initialized")
	}
	
	result, err := a.auditService.DuplicateCIDCHEC(companyName)
	if err != nil {
		return nil, err
	}
	
	// Convert AuditResult to map
	return map[string]interface{}{
		"status":      "completed",
		"success":     result.Success,
		"message":     result.Message,
		"error":       result.Error,
		"totalChecks": result.TotalChecks,
		"issues":      result.Issues,
		"summary":     result.Summary,
		"metadata":    result.Metadata,
	}, nil
}

// AuditVoidChecks wraps the audit service VoidChecks method
func (a *App) AuditVoidChecks(companyName string) (map[string]interface{}, error) {
	if a.auditService == nil {
		return nil, fmt.Errorf("audit service not initialized")
	}
	
	audit, err := a.auditService.VoidChecks(companyName)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"status":           "completed",
		"totalVoidChecks":  audit.TotalVoidChecks,
		"properlyVoided":   audit.ProperlyVoided,
		"issues":           audit.Issues,
		"voidWithNonZero":  audit.VoidWithNonZero,
		"nonVoidWithZero":  audit.NonVoidWithZero,
		"voidButCleared":   audit.VoidButCleared,
		"summary": map[string]interface{}{
			"total_void_checks":       audit.TotalVoidChecks,
			"properly_voided":         audit.ProperlyVoided,
			"void_with_nonzero":       len(audit.VoidWithNonZero),
			"nonvoid_with_zero":       len(audit.NonVoidWithZero),
			"void_but_cleared":        len(audit.VoidButCleared),
			"total_issues":            len(audit.Issues),
		},
	}, nil
}

// AuditCheckGLMatching wraps the audit service CheckGLMatching method
func (a *App) AuditCheckGLMatching(companyName string, accountNumber string, startDate string, endDate string) (map[string]interface{}, error) {
	if a.auditService == nil {
		return nil, fmt.Errorf("audit service not initialized")
	}
	
	audit, err := a.auditService.CheckGLMatching(companyName, accountNumber, startDate, endDate)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"status":             "completed",
		"period":             audit.Period,
		"totalChecks":        audit.TotalChecks,
		"matchedChecks":      audit.MatchedChecks,
		"unmatchedChecks":    audit.UnmatchedChecks,
		"totalGLEntries":     audit.TotalGLEntries,
		"matchedGLEntries":   audit.MatchedGLEntries,
		"unmatchedGLEntries": audit.UnmatchedGLEntries,
		"summary": map[string]interface{}{
			"period":               audit.Period,
			"total_checks":         audit.TotalChecks,
			"matched_checks":       audit.MatchedChecks,
			"unmatched_checks":     len(audit.UnmatchedChecks),
			"total_gl_entries":     audit.TotalGLEntries,
			"matched_gl_entries":   audit.MatchedGLEntries,
			"unmatched_gl_entries": len(audit.UnmatchedGLEntries),
		},
	}, nil
}

// AuditPayeeCIDVerification wraps the audit service PayeeCIDVerification method
func (a *App) AuditPayeeCIDVerification(companyName string) (map[string]interface{}, error) {
	if a.auditService == nil {
		return nil, fmt.Errorf("audit service not initialized")
	}
	
	audit, err := a.auditService.PayeeCIDVerification(companyName)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"status":                 "completed",
		"totalChecks":            audit.TotalChecks,
		"checksWithCID":          audit.ChecksWithCID,
		"checksWithoutCID":       audit.ChecksWithoutCID,
		"uniquePayees":           audit.UniquePayees,
		"payeesWithMultipleCIDs": audit.PayeesWithMultipleCIDs,
		"missingCIDChecks":       audit.MissingCIDChecks,
		"summary": map[string]interface{}{
			"total_checks":              audit.TotalChecks,
			"checks_with_cid":           audit.ChecksWithCID,
			"checks_without_cid":        audit.ChecksWithoutCID,
			"unique_payees":             audit.UniquePayees,
			"payees_with_multiple_cids": len(audit.PayeesWithMultipleCIDs),
			"missing_cid_checks":        len(audit.MissingCIDChecks),
		},
	}, nil
}

// AuditBankReconciliation wraps the audit service BankReconciliation method
func (a *App) AuditBankReconciliation(companyName string) (map[string]interface{}, error) {
	if a.auditService == nil {
		return nil, fmt.Errorf("audit service not initialized")
	}
	
	audit, err := a.auditService.BankReconciliation(companyName)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"status":               "completed",
		"totalAccounts":        audit.TotalAccounts,
		"reconciledAccounts":   audit.ReconciledAccounts,
		"unreconciledAccounts": audit.UnreconciledAccounts,
		"outOfBalanceAccounts": audit.OutOfBalanceAccounts,
		"staleReconciliations": audit.StaleReconciliations,
		"summary": map[string]interface{}{
			"total_accounts":         audit.TotalAccounts,
			"reconciled_accounts":    audit.ReconciledAccounts,
			"unreconciled_accounts":  len(audit.UnreconciledAccounts),
			"out_of_balance_accounts": len(audit.OutOfBalanceAccounts),
			"stale_reconciliations":  len(audit.StaleReconciliations),
		},
	}, nil
}

// AuditSingleBankAccount wraps the audit service SingleBankAccount method
func (a *App) AuditSingleBankAccount(companyName, accountNumber string) (map[string]interface{}, error) {
	if a.auditService == nil {
		return nil, fmt.Errorf("audit service not initialized")
	}
	
	result, err := a.auditService.SingleBankAccount(companyName, accountNumber)
	if err != nil {
		return nil, err
	}
	
	// Convert AuditResult to map
	return map[string]interface{}{
		"status":      "completed",
		"success":     result.Success,
		"message":     result.Message,
		"totalChecks": result.TotalChecks,
		"issues":      result.Issues,
		"summary":     result.Summary,
		"metadata":    result.Metadata,
	}, nil
}