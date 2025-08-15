// +build ignore

// This file contains the deprecated audit functions that have been moved to internal/financials/audit
// These are kept here for reference but are not compiled
// To compile this file temporarily, remove the "+build ignore" directive above

package main

// NOTE: All these functions have been moved to internal/financials/audit package
// They are preserved here for reference only and will be removed in a future version

/*
The following audit functions have been moved:
- AuditCheckBatches -> audit.Service.CheckBatches()
- AuditDuplicateCIDCHEC -> audit.Service.DuplicateCIDCHEC()
- AuditVoidChecks -> audit.Service.VoidChecks()
- AuditCheckGLMatching -> audit.Service.CheckGLMatching()
- AuditPayeeCIDVerification -> audit.Service.PayeeCIDVerification()
- AuditBankReconciliation -> audit.Service.BankReconciliation()
- AuditSingleBankAccount -> audit.Service.SingleBankAccount()

Total lines removed from main.go: ~1,878 lines
*/