# Package Structure Guide

This document describes the organization of the internal packages and what belongs in each.

## Package Overview

### 📦 `internal/common`
**Shared types, constants, and utilities used across all packages**
- Common data structures (Company, Result, etc.)
- Shared constants (date formats, page sizes)
- Common error types
- Shared interfaces

### 📊 `internal/reports`
**All report generation and data export functionality**
- PDF generation (Chart of Accounts, Owner Statements)
- Excel exports
- CSV exports
- Report templates
- Report scheduling
- Examples:
  - `GenerateChartOfAccountsPDF()`
  - `GenerateOwnerStatements()`
  - `ExportCheckRegister()`

### 💰 `internal/financials`
**Core financial and accounting operations**
- General Ledger (GL) operations
- Bank account management
- Bank reconciliation
- Check processing
- Account balances
- Financial calculations
- Examples:
  - `GetBankAccounts()`
  - `GetOutstandingChecks()`
  - `ReconcileBank()`
  - `GetGLBalance()`
  - `AuditCheckBatches()`

### 🏢 `internal/operations`
**Business operations and management**
- Vendor management
- Purchase/AP management
- Well management (oil & gas specific)
- Owner management
- Inventory (if applicable)
- Workflow management
- Examples:
  - `GetVendors()`
  - `UpdateVendor()`
  - `CreatePurchase()`
  - `GetWells()`
  - `GetOwners()`

### 🗄️ `internal/legacy`
**Visual FoxPro and DBF file operations**
- DBF file reading/writing
- VFP form launching
- Legacy data migration
- DBF structure analysis
- Legacy system integration
- Examples:
  - `ReadDBF()`
  - `WriteDBF()`
  - `LaunchVFPForm()`
  - `GetDBFStructure()`

### 🛠️ `internal/utilities`
**Helper functions and utilities**
- String manipulation
- Date/time formatting
- Currency formatting
- File path operations
- Validation helpers
- Encryption/decryption
- ID generation
- Examples:
  - `FormatCurrency()`
  - `ParseDate()`
  - `SanitizeFilename()`
  - `IsValidEmail()`
  - `Encrypt()`/`Decrypt()`

## Migration Strategy

### Phase 1: Setup (Complete) ✅
- Created package structure
- Added basic service definitions
- Created placeholder functions

### Phase 2: Gradual Migration (Current)
1. Start with the easiest/most isolated functions
2. Move related functions together
3. Update imports in main.go
4. Test each migration
5. Commit after each successful migration

### Phase 3: Refactoring
- Improve function signatures
- Add proper error handling
- Add unit tests
- Remove duplication

## How to Move a Function

1. **Identify the target package** based on the function's purpose
2. **Copy the function** to the appropriate package file
3. **Update the function signature** if needed (remove App receiver, etc.)
4. **Add necessary imports** to the package file
5. **Update main.go** to call the package function:
   ```go
   // Before (in main.go):
   func (a *App) GetVendors(company string) ([]map[string]interface{}, error) {
       // 100 lines of code
   }
   
   // After (in main.go):
   func (a *App) GetVendors(company string) ([]operations.Vendor, error) {
       return a.operationsService.GetVendors(company)
   }
   ```
6. **Test** the function still works
7. **Commit** the change

## Package Dependencies

```
common (no dependencies on other internal packages)
   ↑
   ├── utilities (uses common)
   ├── legacy (uses common)
   ├── financials (uses common, utilities, legacy)
   ├── operations (uses common, utilities, legacy)
   └── reports (uses common, utilities, financials, operations)
```

## Existing Packages (Don't Duplicate)

These packages already exist and should continue to be used:
- `internal/auth` - Authentication and authorization
- `internal/company` - Company management and data paths
- `internal/config` - Configuration management
- `internal/currency` - Currency conversion and formatting
- `internal/database` - Database operations
- `internal/logger` - Logging
- `internal/reconciliation` - Bank reconciliation (can be merged with financials later)
- `internal/vfp` - VFP integration (can be merged with legacy later)

## Best Practices

1. **Keep packages focused** - Each package should have a clear, single purpose
2. **Minimize dependencies** - Avoid circular dependencies
3. **Export only what's needed** - Keep internal implementation details private
4. **Document exports** - Add comments to all exported functions and types
5. **Test each package** - Create `*_test.go` files for unit tests
6. **Use interfaces** - Define interfaces for dependencies to make testing easier