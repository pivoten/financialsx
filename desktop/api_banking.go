package main

import (
	"fmt"
)

// ============================================================================
// BANKING & FINANCIAL API
// This file contains all banking and financial-related API methods
// ============================================================================

// GetBankAccounts retrieves all bank accounts for a company
func (a *App) GetBankAccounts(companyName string) ([]map[string]interface{}, error) {
	return a.Services.Banking.GetBankAccountsAsMap(companyName)
}

// GetAccountBalance retrieves the GL balance for a specific account
func (a *App) GetAccountBalance(companyName, accountNumber string) (float64, error) {
	return a.Services.Banking.GetAccountBalance(companyName, accountNumber)
}

// GetCachedBalances retrieves all cached bank balances for a company
func (a *App) GetCachedBalances(companyName string) (interface{}, error) {
	return a.Services.Banking.GetCachedBalances(companyName)
}

// RefreshAccountBalance refreshes the cached balance for a single account
func (a *App) RefreshAccountBalance(companyName, accountNumber string) (map[string]interface{}, error) {
	balance, err := a.Services.Banking.RefreshAccountBalance(companyName, accountNumber, "system")
	if err != nil {
		return nil, err
	}
	
	// Convert to map for frontend
	return map[string]interface{}{
		"account_number":     balance.AccountNumber,
		"account_name":       balance.AccountName,
		"gl_balance":         balance.GLBalance,
		"outstanding_total":  balance.OutstandingChecksTotal,
		"outstanding_count":  balance.OutstandingChecksCount,
		"bank_balance":       balance.BankBalance,
		"gl_last_updated":    balance.GLLastUpdated,
		"checks_last_updated": balance.ChecksLastUpdated,
		"is_stale":           balance.IsStale,
	}, nil
}

// RefreshAllBalances refreshes cached balances for all bank accounts
func (a *App) RefreshAllBalances(companyName string) (map[string]interface{}, error) {
	return a.Services.Banking.RefreshAllBalances(companyName)
}

// GetOutstandingChecks retrieves outstanding checks for a bank account
func (a *App) GetOutstandingChecks(companyName, accountNumber string) ([]map[string]interface{}, error) {
	checks, err := a.Services.Banking.GetOutstandingChecks(companyName, accountNumber)
	if err != nil {
		return nil, err
	}
	
	// Convert to map format for frontend
	result := make([]map[string]interface{}, len(checks))
	for i, check := range checks {
		result[i] = map[string]interface{}{
			"checkNumber": check.CheckNumber,
			"date":        check.CheckDate,
			"payee":       check.Payee,
			"amount":      check.Amount,
			"account":     check.AccountNumber,
			"entryType":   check.EntryType,
			"cidchec":     check.CIDCHEC,
			"id":          check.ID,
			"_rowIndex":   check.RowIndex,
			"_rawData":    check.RawData,
		}
	}
	
	return result, nil
}

// GetBalanceHistory retrieves the balance change history for an account
func (a *App) GetBalanceHistory(companyName, accountNumber string, limit int) ([]map[string]interface{}, error) {
	// TODO: Implement in banking service
	return nil, fmt.Errorf("not implemented")
}

// GetBankAccountsForAudit retrieves bank accounts with audit information
func (a *App) GetBankAccountsForAudit(companyName string) ([]map[string]interface{}, error) {
	return a.Services.Audit.GetBankAccountsForAudit(companyName)
}