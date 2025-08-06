# FinancialsX Desktop - Claude Code Documentation

## Project Overview
FinancialsX Desktop is a Wails-based application for oil & gas financial management with comprehensive banking features. Built with Go backend and React frontend using ShadCN UI components.

## Key Architecture
- **Backend**: Go with Wails framework
- **Frontend**: React with Vite, TypeScript, ShadCN UI
- **Database**: DBF files for legacy data + SQLite for user management
- **DBF Library**: `github.com/Valentin-Kaiser/go-dbase/dbase`

## Critical File Structure Issues & Fixes

### DBF Data Structure (CRITICAL)
**Issue**: Bank account loading was failing with "invalid data format from COA.dbf"
**Root Cause**: Mismatch between expected data keys
- `ReadDBFFile()` in `internal/company/company.go` returns data under `"rows"` key
- `GetBankAccounts()` in `main.go` was looking for `"data"` key
- **FIXED**: Changed `coaData["data"]` to `coaData["rows"]` in both Go and JS

### COA.dbf Structure (Bank Accounts)
```
Column Index | Field Name | Description
0           | CACCTNO    | Account Number  
1           | NACCTTYPE  | Account Type
2           | CACCTDESC  | Account Description
3           | CPARENT    | Parent Account
4           | LACCTUNIT  | Account Unit Flag
5           | LACCTDEPT  | Department Flag  
6           | LBANKACCT  | Bank Account Flag (TRUE/FALSE) **KEY COLUMN**
```

### CHECKS.dbf Structure (Outstanding Checks)
```
Field Name    | Description
CCHECKNO     | Check Number (string)
DCHECKDATE   | Check Date (date)
CPAYEE       | Payee Name (string)
NAMOUNT      | Check Amount (numeric)
CACCTNO      | Bank Account Number (string)
LCLEARED     | Cleared Flag (logical) - FALSE for outstanding
LVOID        | Void Flag (logical) - FALSE for valid checks
CBATCH       | Batch Number (optional, for audit matching)
```

**Outstanding Checks Logic**:
- Include only checks where `LCLEARED = false` AND `LVOID = false`
- Filter by `CACCTNO` when account-specific view is requested
- Calculate days outstanding from `DCHECKDATE` to current date

## Banking Section Implementation

### Bank Account Loading Process
1. **Primary**: `GetBankAccounts(companyName)` - Go function reads COA.dbf
2. **Fallback**: `GetDBFTableData(companyName, 'COA.dbf')` - Direct DBF read
3. **Filter**: Only accounts where `LBANKACCT = true` (column 6)
4. **Transform**: Convert to display format with account_number, account_name, etc.
5. **Balance Loading**: `GetAccountBalance()` reads GLMASTER.dbf to sum GL entries for each bank account

### GL Balance Integration
**New Feature**: Bank account cards now display actual General Ledger balances instead of hardcoded $0.00

**Implementation Details**:
1. **Backend Function**: `GetAccountBalance(companyName, accountNumber)` in `main.go:587-649`
   - Reads GLMASTER.dbf and sums all GL entries for the specified account number
   - Handles various column name variations (CACCTNO, ACCOUNT, ACCTNO for account numbers)
   - Supports different amount column names (AMOUNT, NAMOUNT, BALANCE, NBALANCE)
   - Returns total balance as float64 with proper error handling

2. **Frontend Integration**: `loadAccountBalances()` in `BankingSection.jsx:212-237`
   - Called automatically after bank accounts are loaded from COA.dbf
   - Iterates through all discovered bank accounts
   - Fetches real-time GL balance for each account using `GetAccountBalance()`
   - Updates account objects with actual balances for display
   - Graceful error handling - falls back to $0.00 if balance fetch fails

3. **User Experience**:
   - Bank account cards show color-coded balances (green for positive, red for negative)
   - Real-time data from GLMASTER.dbf ensures accuracy
   - Seamless loading experience with balance updates after initial account discovery
   - Maintains existing UI/UX while enhancing data accuracy

**Technical Requirements**:
- GLMASTER.dbf must exist in company data directory
- Account numbers in COA.dbf must match account numbers in GLMASTER.dbf
- User must have `database.read` permission to access GL data

### Outstanding Checks Feature (Enhanced)
**New Enhanced Feature**: Complete outstanding checks management with enterprise-level data handling

**Backend Implementation**:
1. **Function**: `GetOutstandingChecks(companyName, accountNumber)` in `main.go:587-722`
   - Enhanced to support account filtering (only show checks for specific bank account)
   - Filters checks where `LCLEARED = false` and `LVOID = false`
   - Returns raw row data for editing capabilities
   - Handles various DBF column name variations

**Frontend Implementation**: `OutstandingChecks.jsx`
1. **Account Filtering**: Dropdown to filter by specific bank account or show all
2. **Row Selection**: Click any row to view/edit individual check details
3. **Pagination**: Configurable page sizes (10, 25, 50, 100) with smart pagination controls
4. **Sorting**: All columns sortable (Check #, Date, Payee, Amount, Days Outstanding)
5. **Search**: Global search across all check fields (number, payee, account, amount)
6. **Filtering**: "Stale Only" filter for checks >90 days outstanding
7. **Edit Modal**: Full CRUD operations for Admin/Root users with permission checks
8. **Summary Cards**: Real-time totals for outstanding count, amount, and stale checks

**Key Features**:
- **Account-Specific View**: Only shows checks for selected bank account (COA.dbf where `LBANKACCT = true`)
- **Days Outstanding Calculation**: Color-coded badges (green ≤30 days, yellow ≤60, red ≤90, red+STALE >90)
- **Permission-Based Editing**: Edit button only appears for Admin/Root users
- **Responsive Design**: Full mobile and desktop support with ShadCN UI components

## Bank Reconciliation System (SQLite-Enhanced)

### Modern Reconciliation Architecture
The bank reconciliation system has been upgraded from localStorage to enterprise-grade SQLite persistence with JSON field flexibility.

### SQLite Reconciliation Schema
```sql
-- Mirror of CHECKREC.DBF with JSON extensions
CREATE TABLE reconciliations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  company_name TEXT NOT NULL,
  account_number TEXT NOT NULL,
  reconcile_date DATE NOT NULL,
  statement_date DATE NOT NULL,
  beginning_balance DECIMAL(15,2) NOT NULL DEFAULT 0,
  ending_balance DECIMAL(15,2) NOT NULL DEFAULT 0,
  statement_balance DECIMAL(15,2) NOT NULL DEFAULT 0,
  statement_credits DECIMAL(15,2) DEFAULT 0,
  statement_debits DECIMAL(15,2) DEFAULT 0,
  
  -- JSON field for extended data and future fields
  extended_data TEXT DEFAULT '{}',
  
  -- Selected checks as JSON array with CIDCHEC details
  selected_checks_json TEXT DEFAULT '[]',
  
  -- Status and metadata
  status TEXT DEFAULT 'draft', -- draft, committed, archived
  created_by TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  committed_at TIMESTAMP NULL,
  
  -- DBF sync metadata for bidirectional sync
  dbf_row_index INTEGER NULL,
  dbf_last_sync TIMESTAMP NULL,
  
  UNIQUE(company_name, account_number, reconcile_date, status)
);
```

### Reconciliation API Endpoints
New SQLite-backed API functions in `main.go`:

1. **`SaveReconciliationDraft(companyName, draftData)`**
   - Saves/updates draft reconciliations with CIDCHEC tracking
   - Auto-converts form data to structured JSON
   - Handles selected checks with full metadata

2. **`GetReconciliationDraft(companyName, accountNumber)`**
   - Retrieves current draft for an account
   - Returns structured data for form population
   - Supports CIDCHEC-based check matching

3. **`DeleteReconciliationDraft(companyName, accountNumber)`**
   - Clears draft reconciliation data
   - Admin/Root permission required

4. **`CommitReconciliation(companyName, accountNumber)`**
   - Commits draft to permanent status
   - Changes status from 'draft' to 'committed'
   - TODO: Will update DBF files in future version

5. **`GetReconciliationHistory(companyName, accountNumber)`**
   - Retrieves historical reconciliations for an account
   - Returns up to 50 most recent records
   - Useful for reconciliation audit trails

6. **`MigrateReconciliationData(companyName)`**
   - Framework for importing existing CHECKREC.DBF data
   - Database maintenance permission required
   - TODO: Full implementation pending

### Frontend Integration
Updated `BankReconciliation.jsx` for SQLite integration:

**Key Changes**:
- Replaced localStorage with SQLite API calls
- Maintained CIDCHEC-based check tracking for reliability
- Enhanced error handling and user feedback
- Auto-save functionality with database persistence

**CIDCHEC Tracking**:
```javascript
// Selected checks stored with full metadata
const selectedChecksDetails = Array.from(draftSelectedChecks).map(checkId => {
  const check = checks.find(c => c.id === checkId)
  return {
    cidchec: check.cidchec,        // Unique identifier
    checkNumber: check.checkNumber,
    amount: check.amount,
    payee: check.payee,
    checkDate: check.checkDate,
    rowIndex: check.rowIndex       // For DBF updates
  }
})
```

### Reconciliation Service Layer
New service in `internal/reconciliation/reconciliation.go`:

**Core Features**:
- CRUD operations for reconciliation management
- JSON-based check storage with CIDCHEC tracking
- Draft workflow with status management
- Migration framework for DBF integration
- Proper error handling and validation

**Data Structures**:
```go
type Reconciliation struct {
  ID                 int                 `json:"id"`
  CompanyName        string              `json:"company_name"`
  AccountNumber      string              `json:"account_number"`
  SelectedChecks     []SelectedCheck     `json:"selected_checks"`
  Status             string              `json:"status"`
  ExtendedData       map[string]interface{} `json:"extended_data"`
  // ... other fields
}

type SelectedCheck struct {
  CIDCHEC     string  `json:"cidchec"`      // Unique check ID
  CheckNumber string  `json:"checkNumber"`
  Amount      float64 `json:"amount"`
  RowIndex    int     `json:"rowIndex"`     // DBF row reference
  // ... other fields
}
```

### Benefits of SQLite Architecture

1. **Reliability**: Database ACID properties vs localStorage volatility
2. **Performance**: Indexed queries vs linear localStorage scans  
3. **Multi-user Support**: Concurrent access and locking
4. **Extensibility**: JSON fields allow future enhancements without schema changes
5. **Audit Trail**: Full tracking of who made changes and when
6. **Data Integrity**: Foreign keys and constraints
7. **Backup/Recovery**: Standard database backup procedures

### Migration Strategy

**Phase 1** ✅ **COMPLETED**: SQLite schema and API implementation
- Created reconciliations table with JSON fields
- Implemented CRUD API endpoints
- Updated frontend to use SQLite instead of localStorage

**Phase 2** (Future): DBF Bidirectional Sync
- Import existing CHECKREC.DBF records to SQLite
- Implement DBF update functionality when reconciliations are committed
- Add sync metadata for change tracking

**Phase 3** (Future): Advanced Features
- Reconciliation templates and automation
- Bank statement import integration
- Advanced reporting and analytics
- Conflict resolution for multi-user scenarios

### Transaction Matching System (Enhanced)

#### Intelligent Matching Algorithm
- **Date Proximity Weighting**: 40% of match score based on date closeness
- **Recurring Transaction Support**: Handles same-amount transactions by closest date matching
- **Double-Match Prevention**: Tracks already matched check IDs to prevent duplicates
- **Confidence Scoring**: Multi-factor scoring including amount, date, check number, and payee

#### Matching Options Dialog
When running matching or refresh operations, users can choose:
1. **Match All Available Checks**: Include all checks regardless of date
2. **Match Only to Statement Date**: Limit matching to checks dated on or before statement date

Backend implementation in `RunMatching()` accepts options parameter:
```go
options := map[string]interface{}{
    "limitToStatementDate": true/false,
    "statementDate": "2025-01-31"
}
```

#### Outstanding Checks Date Filter
- **Toggle Button**: Filter outstanding checks list by statement date
- **Visual Indicator**: Button shows `≤ [date]` when active, "All Dates" when inactive
- **Default Behavior**: Filter enabled by default to focus on relevant period
- **Dynamic Count**: Check count updates based on active filters

### Check Batch Audit Feature (Admin/Root Only)
1. **Function**: `AuditCheckBatches()` in `main.go:651-777` (updated line numbers)
2. **UI Component**: `CheckAudit.jsx` - Full audit interface with results display
3. **Purpose**: Compare checks.dbf entries with GLMASTER.dbf
4. **Checks for**:
   - Missing GL entries (checks without corresponding GL records)
   - Mismatched amounts (checks where amounts differ from GL)
5. **Fallback**: If CBATCH field doesn't exist, uses check number for matching
6. **Export**: Results can be exported to CSV format

### Key Functions
- **Go**: `GetBankAccounts()` in `main.go:484-585`
- **Go**: `GetAccountBalance()` in `main.go:587-649` - Fetches GL balance for specific account
- **Go**: `GetOutstandingChecks(companyName, accountNumber)` in `main.go:587-722` - Enhanced with account filtering
- **Go**: `RunMatching(companyName, accountNumber, options)` in `main.go:991-1112` - Intelligent transaction matching with date filtering
- **Go**: `ClearMatchesAndRerun(companyName, accountNumber, options)` in `main.go:1115-1149` - Clear and re-run matching with options
- **Go**: `autoMatchBankTransactions()` in `main.go:1394-1500` - Core matching algorithm with date proximity scoring
- **Go**: `AuditCheckBatches()` in `main.go:651-777`
- **Go**: `SaveReconciliationDraft()` in `main.go:2023-2110` - SQLite-based draft persistence
- **Go**: `GetReconciliationDraft()` in `main.go:2113-2144` - Retrieve draft from SQLite
- **Go**: `CommitReconciliation()` in `main.go:2175-2210` - Commit draft to permanent status
- **React**: `loadBankAccounts()` in `BankingSection.jsx:105-210`
- **React**: `loadAccountBalances()` in `BankingSection.jsx:212-237` - Loads GL balances for all bank accounts
- **React**: `handleRunMatching()` in `BankReconciliation.jsx:1271-1316` - Handles matching with date options dialog
- **React**: `handleRefreshMatching()` in `BankReconciliation.jsx:1320-1366` - Clear and re-match with options
- **React**: `getAvailableChecks()` in `BankReconciliation.jsx:961-1000` - Filters checks by statement date
- **React**: `OutstandingChecks.jsx` - Enhanced outstanding checks with full data management
- **React**: `CheckAudit.jsx` - Complete audit component
- **React**: `BankReconciliation.jsx` - SQLite-based reconciliation with CIDCHEC tracking, date filtering, and matching options
- **DBF Reader**: `ReadDBFFile()` in `company.go:225-443`
- **Reconciliation Service**: `internal/reconciliation/reconciliation.go` - Complete CRUD service layer

## User Management & Permissions

### Role System
- **Root**: Full system access (is_root = true)
- **Admin**: Administrative privileges (role_name = 'Admin')  
- **Read-Only**: Limited access (role_name = 'Read-Only')

### Permission Checks
```go
// Check if user can edit DBF files
func canEdit() bool {
    return currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')
}
```

## Reusable DataTable Component (NEW)

**Location**: `frontend/src/components/ui/data-table.jsx`

### Overview
Enterprise-level reusable data table component that establishes the standard pattern for all data lists in the system. Used by Outstanding Checks and designed for future list implementations.

### Core Features
1. **Configurable Columns**: Custom rendering, sorting, and cell styling
2. **Built-in Pagination**: Configurable page sizes with smart controls
3. **Global Search**: Search across all columns with highlighting
4. **Custom Filters**: Dropdown filters with custom filter functions
5. **Sorting**: Click headers to sort, visual indicators for sort direction
6. **Row Actions**: Click handlers for row selection and detail views
7. **Loading States**: Built-in loading and error state handling
8. **Responsive Design**: Full mobile/desktop support with ShadCN UI

### Usage Pattern
```javascript
<DataTable
  data={items}
  columns={columnConfig}
  title="Data List Title"
  loading={loading}
  error={error}
  onRowClick={handleRowClick}
  onRefresh={handleRefresh}
  searchPlaceholder="Search items..."
  filters={filterConfig}
  pageSize={25}
/>
```

### Column Configuration
```javascript
const columns = [
  {
    accessor: 'fieldName',
    header: 'Display Name',
    sortable: true,
    type: 'number|date|string',
    render: (value, row, index) => <CustomComponent />,
    cellClassName: 'text-right',
    headerClassName: 'text-center'
  }
]
```

### Filter Configuration
```javascript
const filters = [
  {
    key: 'filterKey',
    label: 'Filter Label',
    placeholder: 'Select option',
    defaultValue: 'all',
    options: [
      { value: 'all', label: 'All Items' },
      { value: 'active', label: 'Active Only' }
    ],
    filterFn: (row, value) => value === 'all' || row.status === value
  }
]
```

### Standard Data List Pattern
This component establishes the pattern for ALL data lists in the system:

1. **Account Filtering**: For bank-related data, filter by bank account
2. **Row Selection**: Click rows to view/edit details in modal
3. **Pagination**: Always include pagination for large datasets
4. **Sorting**: Make all relevant columns sortable
5. **Search**: Global search across all text fields
6. **Permission-Based Actions**: Show edit buttons only for Admin/Root
7. **Responsive Design**: Works on all screen sizes
8. **Loading States**: Show loading spinners and error messages

### Implementation Examples
- **OutstandingChecks.jsx**: Full implementation with all features
- **OutstandingChecksSimple.jsx**: Simplified example using DataTable component

## DBF Explorer Features

### Read-Only Mode
- Default: Read-only view for all users
- Edit button only shows for Admin/Root users
- Uses `canEdit()` permission check

### Data Type Formatting
```javascript
const formatLogicalValue = (value) => {
  if (value === null || value === undefined || value === '') return ''
  if (typeof value === 'boolean') return value ? 'True' : 'False'
  if (typeof value === 'string') {
    const lowerVal = value.toLowerCase().trim()
    if (lowerVal === 't' || lowerVal === '.t.' || lowerVal === 'true') return 'True'
    if (lowerVal === 'f' || lowerVal === '.f.' || lowerVal === 'false') return 'False'
  }
  return value
}
```

### Record Detail Modal
- Click any row to view complete record
- Shows all fields in formatted layout
- Handles null/empty values gracefully

## Application Settings

### Window Configuration
```go
// main.go:608-620
Width:  1400,  // Increased from 1200
Height: 1000,  // Increased from 900
```

### Development Server
- Frontend dev server: `npm run dev`
- Wails dev server: `wails dev` 
- Browser access: http://localhost:34115
- Chrome DevTools: F12 for console debugging

## Build & Development Commands

### Essential Commands
```bash
# Development
wails dev                    # Start development server
npm run dev                  # Frontend only development

# Building  
wails build                  # Production build
npm run build               # Frontend build only

# Dependencies
npm install                 # Install frontend dependencies
go mod tidy                 # Update Go dependencies
```

### Wails Binding Generation
When Go functions are added/changed:
```bash
wails generate              # Regenerate JavaScript bindings
```

## Common Issues & Solutions

### 1. "GetBankAccounts is not a function"
**Cause**: Wails bindings not regenerated after Go changes
**Solution**: Run `wails generate` or use fallback method

### 2. "No bank accounts found"  
**Check**: COA.dbf column 6 (LBANKACCT) has TRUE values
**Debug**: Enable console logging in browser DevTools

### 3. "invalid data format from COA.dbf"
**Cause**: Wrong data structure key (should be "rows" not "data")
**Solution**: Ensure `coaData["rows"]` is used

### 4. Edit button not showing
**Cause**: currentUser prop not passed to components
**Solution**: Verify `<DBFExplorer currentUser={currentUser} />`

## File Organization

### Key Files
```
main.go                                    # Main application & API endpoints
internal/company/company.go                # DBF file operations
internal/auth/auth.go                     # User authentication
internal/database/database.go             # SQLite operations & reconciliation schema
internal/reconciliation/reconciliation.go # SQLite reconciliation service (NEW)

frontend/src/
├── App.jsx                      # Main application component
├── components/
│   ├── BankingSection.jsx      # Banking module (CRITICAL)
│   ├── BankReconciliation.jsx  # SQLite-based bank reconciliation (UPDATED)
│   ├── DBFExplorer.jsx         # DBF file viewer/editor
│   ├── OutstandingChecks.jsx   # Enhanced outstanding checks
│   ├── OutstandingChecksSimple.jsx # DataTable usage example
│   └── ui/
│       ├── data-table.jsx      # Reusable data table component
│       └── [other ShadCN UI components]
└── wailsjs/go/main/App.js      # Generated Wails bindings (AUTO-UPDATED)
```

## Database Schema

### User Management (SQLite)
```sql
users table:
- id (INTEGER PRIMARY KEY)
- username (TEXT UNIQUE)
- password_hash (TEXT)  
- email (TEXT)
- role_id (INTEGER)
- is_active (BOOLEAN)
- is_root (BOOLEAN)
- company_name (TEXT)
```

### Bank Reconciliation (SQLite)
```sql
reconciliations table:
- id (INTEGER PRIMARY KEY)
- company_name (TEXT NOT NULL)
- account_number (TEXT NOT NULL)
- reconcile_date (DATE NOT NULL)
- statement_date (DATE NOT NULL)
- beginning_balance (DECIMAL(15,2))
- ending_balance (DECIMAL(15,2))
- statement_balance (DECIMAL(15,2))
- statement_credits (DECIMAL(15,2))
- statement_debits (DECIMAL(15,2))
- extended_data (TEXT) -- JSON for future fields
- selected_checks_json (TEXT) -- JSON array with CIDCHEC IDs
- status (TEXT) -- draft, committed, archived
- created_by (TEXT)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
- committed_at (TIMESTAMP)
- dbf_row_index (INTEGER) -- For DBF sync
- dbf_last_sync (TIMESTAMP) -- For DBF sync
```

### Company Data (DBF Files)
- Located in `../datafiles/{company_name}/`
- COA.dbf: Chart of Accounts
- CHECKS.dbf: Check transactions (with CIDCHEC unique IDs)
- CHECKREC.dbf: Bank reconciliation history (mirrors SQLite reconciliations table)
- INCOME.dbf: Revenue records
- EXPENSE.dbf: Expense records
- WELLS.dbf: Well information
- GLMASTER.dbf: General Ledger entries

## Testing & Debugging

### Browser Console Debugging
1. Open Chrome DevTools (F12)
2. Check Console tab for errors
3. Look for specific error patterns:
   - "GetBankAccounts returned undefined"
   - "No data found in COA.dbf" 
   - "invalid data format"

### Go Backend Debugging
- Check terminal output for printf statements
- Look for "GetBankAccounts:" prefixed log messages
- Verify file paths and permissions

## Future Development Notes

### Planned Features
1. Bank reconciliation with checks.dbf integration
2. Automated bank statement import
3. Enhanced reporting capabilities
4. Real-time data validation

### Architecture Considerations
- DBF files are legacy format - consider migration path
- Wails binding regeneration needed for Go API changes
- ShadCN UI provides consistent component library
- Role-based access control implemented throughout

## Emergency Recovery

### If Application Won't Start
1. Check Go compilation: `go build main.go`
2. Verify Node dependencies: `npm install`
3. Check datafiles directory exists: `../datafiles/`
4. Verify SQLite database permissions

### If Bank Accounts Don't Load
1. Verify COA.dbf exists in company directory
2. Check LBANKACCT column has TRUE values
3. Enable debug logging in browser console
4. Use fallback GetDBFTableData method

### If Reconciliation Drafts Don't Save
1. Check SQLite database connection
2. Verify reconciliation service initialization in Go
3. Check user permissions (`dbf.write` required for saving)
4. Monitor browser console for API errors

## Future Development: Audit Results Persistence

### Problem
Currently, audit results are only stored in React state. When users navigate away from the audit tab, the results are lost and the audit must be run again. This is inefficient for large datasets (e.g., 1460 mismatched entries).

### Proposed Solution
1. **Create SQLite Tables**:
   ```sql
   -- Audit runs table
   CREATE TABLE audit_runs (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     company_name TEXT NOT NULL,
     audit_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     audited_by TEXT NOT NULL,
     total_checks INTEGER,
     matched_entries INTEGER,
     missing_entries INTEGER,
     mismatched_amounts INTEGER,
     status TEXT DEFAULT 'completed'
   );

   -- Audit details table for issues found
   CREATE TABLE audit_details (
     id INTEGER PRIMARY KEY AUTOINCREMENT,
     audit_run_id INTEGER NOT NULL,
     issue_type TEXT NOT NULL, -- 'missing' or 'mismatched'
     check_id TEXT,
     check_amount DECIMAL(10,2),
     gl_amount DECIMAL(10,2),
     difference DECIMAL(10,2),
     row_data TEXT, -- JSON blob of full row data
     FOREIGN KEY (audit_run_id) REFERENCES audit_runs(id)
   );
   ```

2. **Backend Changes**:
   - Modify `AuditCheckBatches()` to save results to SQLite
   - Add `GetLastAuditResults(companyName)` to retrieve saved audit
   - Add `ClearAuditResults(companyName)` for rerun functionality

3. **Frontend Changes**:
   - On component mount, check for existing audit results
   - Change "Run Audit" to "Rerun Audit" when results exist
   - Add confirmation dialog for rerun (warns about clearing old data)
   - Display audit timestamp on results page

4. **Additional Features**:
   - Audit history view (list of past audits with dates)
   - Compare audits feature (show what changed between runs)
   - Export specific audit results by ID
   - Auto-save audit progress for very large datasets

### Implementation Notes
- Use transactions for atomic saves
- Consider pagination at the database level for better performance
- Add indexes on check_id and audit_run_id for fast queries
- Store check_columns metadata with audit run for consistency
- Consider archiving old audits after X days/months

---

**Last Updated**: August 6, 2025
**Key Fix Applied**: Changed data structure key from "data" to "rows" for DBF file reading
**Latest Enhancement**: Implemented SQLite-based Bank Reconciliation System with JSON field extensibility
**Major Features Added**:
- **SQLite Reconciliation System**: Enterprise-grade database persistence replacing localStorage
- **CIDCHEC-Based Tracking**: Reliable check identification across sessions using unique IDs
- **JSON Field Architecture**: Extensible schema for future enhancements without migrations
- **Reconciliation Service Layer**: Complete CRUD operations with proper error handling
- **6 New API Endpoints**: Full draft workflow with save/load/commit/history capabilities
- **Migration Framework**: Structure for importing existing CHECKREC.DBF data
- **Auto-Save Functionality**: Database persistence with user feedback
**Status**: Bank Reconciliation system upgraded to SQLite with modern architecture, ready for production use