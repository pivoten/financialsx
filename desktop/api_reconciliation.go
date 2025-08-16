package main

import (
	"fmt"
)

// ============================================================================
// RECONCILIATION API
// This file contains all bank reconciliation-related API methods
// ============================================================================

// SaveReconciliationDraft saves a draft reconciliation to the database
func (a *App) SaveReconciliationDraft(companyName string, draftData map[string]interface{}) error {
	if err := a.requirePermission("reconciliation.write"); err != nil {
		return err
	}

	if err := a.requireService("reconciliation"); err != nil {
		return err
	}

	username := "system"
	if a.currentUser != nil {
		username = a.currentUser.Username
	}
	_, err := a.reconciliationService.SaveDraftFromMap(companyName, draftData, username)
	return err
}

// GetReconciliationDraft retrieves a draft reconciliation from the database
func (a *App) GetReconciliationDraft(companyName, accountNumber string) (map[string]interface{}, error) {
	if err := a.requirePermission("reconciliation.read"); err != nil {
		return nil, err
	}

	if err := a.requireService("reconciliation"); err != nil {
		return nil, err
	}

	return a.reconciliationService.GetDraftAsMap(companyName, accountNumber)
}

// DeleteReconciliationDraft deletes a draft reconciliation from the database
func (a *App) DeleteReconciliationDraft(companyName, accountNumber string) error {
	if err := a.requirePermission("reconciliation.delete"); err != nil {
		return err
	}

	if err := a.requireService("reconciliation"); err != nil {
		return err
	}

	return a.reconciliationService.DeleteDraft(companyName, accountNumber)
}

// CommitReconciliation commits a draft reconciliation to permanent status
func (a *App) CommitReconciliation(companyName, accountNumber string) error {
	if err := a.requirePermission("reconciliation.commit"); err != nil {
		return err
	}

	if err := a.requireService("reconciliation"); err != nil {
		return err
	}

	// Get the draft first
	draft, err := a.reconciliationService.GetDraft(companyName, accountNumber)
	if err != nil {
		return err
	}
	
	username := "system"
	if a.currentUser != nil {
		username = a.currentUser.Username
	}
	
	return a.reconciliationService.CommitReconciliation(draft.ID, username)
}

// GetReconciliationHistory retrieves historical reconciliations for an account
func (a *App) GetReconciliationHistory(companyName, accountNumber string) ([]map[string]interface{}, error) {
	if err := a.requirePermission("reconciliation.read"); err != nil {
		return nil, err
	}

	if err := a.requireService("reconciliation"); err != nil {
		return nil, err
	}

	history, err := a.reconciliationService.GetHistory(companyName, accountNumber, 50)
	if err != nil {
		return nil, err
	}
	
	// Convert to map format
	result := make([]map[string]interface{}, len(history))
	for i, rec := range history {
		result[i] = map[string]interface{}{
			"id":                rec.ID,
			"account_number":    rec.AccountNumber,
			"reconcile_date":    rec.ReconcileDate,
			"statement_date":    rec.StatementDate,
			"beginning_balance": rec.BeginningBalance,
			"ending_balance":    rec.EndingBalance,
			"status":            rec.Status,
			"created_at":        rec.CreatedAt,
			"created_by":        rec.CreatedBy,
		}
	}
	
	return result, nil
}

// MigrateReconciliationData migrates legacy reconciliation data from DBF to SQLite
func (a *App) MigrateReconciliationData(companyName string) error {
	if err := a.requirePermission("database.maintenance"); err != nil {
		return err
	}

	if err := a.requireService("reconciliation"); err != nil {
		return err
	}

	// TODO: Implement migration from CHECKREC.DBF to SQLite
	return fmt.Errorf("migration not yet implemented")
}

// ImportBankStatement imports a bank statement for matching
func (a *App) ImportBankStatement(companyName, accountNumber string, statementData string) (map[string]interface{}, error) {
	if err := a.requirePermission("bank.import"); err != nil {
		return nil, err
	}

	return a.Services.Matching.ImportBankStatement(companyName, accountNumber, statementData)
}

// GetBankTransactions retrieves imported bank transactions
func (a *App) GetBankTransactions(companyName, accountNumber string) (map[string]interface{}, error) {
	// TODO: Add statement ID parameter when needed
	return a.Services.Matching.GetBankTransactions(companyName, accountNumber, "")
}

// DeleteBankStatement deletes an imported bank statement
func (a *App) DeleteBankStatement(companyName, accountNumber, statementID string) error {
	if err := a.requirePermission("bank.delete"); err != nil {
		return err
	}

	// The DeleteStatement method takes companyName and statementID
	return a.Services.Matching.DeleteStatement(companyName, statementID)
}

// RunMatching runs the automatic matching algorithm
func (a *App) RunMatching(companyName, accountNumber string, options map[string]interface{}) (map[string]interface{}, error) {
	if err := a.requirePermission("matching.run"); err != nil {
		return nil, err
	}

	return a.Services.Matching.RunMatching(companyName, accountNumber, options)
}

// ClearMatchesAndRerun clears existing matches and reruns matching
func (a *App) ClearMatchesAndRerun(companyName, accountNumber string, options map[string]interface{}) (map[string]interface{}, error) {
	if err := a.requirePermission("matching.run"); err != nil {
		return nil, err
	}

	// Clear existing matches
	// TODO: Implement ClearMatches in matching service
	// For now, just rerun matching which should overwrite

	// Rerun matching
	return a.Services.Matching.RunMatching(companyName, accountNumber, options)
}

// GetMatchedTransactions retrieves matched transactions for review
func (a *App) GetMatchedTransactions(companyName, accountNumber string) (map[string]interface{}, error) {
	return a.Services.Matching.GetMatchedTransactions(companyName, accountNumber)
}

// UpdateMatchConfidence updates the confidence level of a match
func (a *App) UpdateMatchConfidence(companyName string, transactionID int, confidence float64) error {
	if err := a.requirePermission("matching.update"); err != nil {
		return err
	}

	// TODO: Implement UpdateMatchConfidence in matching service
	return fmt.Errorf("update match confidence not yet implemented")
}

// UnmatchTransaction removes a match between a bank transaction and check
func (a *App) UnmatchTransaction(companyName string, transactionID int) error {
	if err := a.requirePermission("matching.update"); err != nil {
		return err
	}

	_, err := a.Services.Matching.UnmatchTransaction(transactionID)
	return err
}

// ManualMatch manually matches a bank transaction to a check
func (a *App) ManualMatch(companyName string, transactionID int, checkID string) error {
	if err := a.requirePermission("matching.update"); err != nil {
		return err
	}

	// TODO: Implement ManualMatch in matching service
	return fmt.Errorf("manual match not yet implemented")
}