package main

// ============================================================================
// GENERAL LEDGER API
// This file contains all GL and accounting-related API methods
// ============================================================================

// CheckGLPeriodFields checks for blank period fields in GL
func (a *App) CheckGLPeriodFields(companyName string) (map[string]interface{}, error) {
	return a.Services.GL.CheckGLPeriodFields(companyName)
}

// AnalyzeGLBalancesByYear analyzes GL balances by year
func (a *App) AnalyzeGLBalancesByYear(companyName string, accountNumber string) (map[string]interface{}, error) {
	return a.Services.GL.AnalyzeGLBalancesByYear(companyName, accountNumber)
}

// ValidateGLBalances validates GL balances
func (a *App) ValidateGLBalances(companyName string, accountNumber string) (map[string]interface{}, error) {
	return a.Services.GL.ValidateGLBalances(companyName, accountNumber)
}

// GetChartOfAccounts retrieves the chart of accounts
func (a *App) GetChartOfAccounts(companyName string, sortBy string, includeInactive bool) ([]map[string]interface{}, error) {
	return a.Services.GL.GetChartOfAccounts(companyName, sortBy, includeInactive)
}

// RunClosingProcess runs the period closing process
func (a *App) RunClosingProcess(companyName string, periodEnd string, closingDate string, description string, forceClose bool) (map[string]interface{}, error) {
	if err := a.requirePermission("gl.close"); err != nil {
		return nil, err
	}

	result, err := a.Services.GL.RunClosingProcess(companyName, periodEnd, closingDate, description, forceClose)
	if err != nil {
		return nil, err
	}
	
	// Convert result to map
	return map[string]interface{}{
		"status":           result.Status,
		"period_end":       result.PeriodEnd,
		"entries_created":  result.EntriesCreated,
		"accounts_affected": result.AccountsAffected,
		"total_debits":     result.TotalDebits,
		"total_credits":    result.TotalCredits,
		"warnings":         result.Warnings,
		"errors":           result.Errors,
	}, nil
}

// GetClosingStatus gets the status of a period closing
func (a *App) GetClosingStatus(companyName string, periodEnd string) (string, error) {
	return a.Services.GL.GetClosingStatus(companyName, periodEnd)
}

// ReopenPeriod reopens a closed period
func (a *App) ReopenPeriod(companyName string, periodEnd string, reason string) error {
	if err := a.requirePermission("gl.reopen"); err != nil {
		return err
	}

	return a.Services.GL.ReopenPeriod(companyName, periodEnd, reason)
}