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
- **Go**: `AuditCheckBatches()` in `main.go:651-777`
- **React**: `loadBankAccounts()` in `BankingSection.jsx:105-210`
- **React**: `loadAccountBalances()` in `BankingSection.jsx:212-237` - Loads GL balances for all bank accounts
- **React**: `OutstandingChecks.jsx` - Enhanced outstanding checks with full data management
- **React**: `CheckAudit.jsx` - Complete audit component
- **DBF Reader**: `ReadDBFFile()` in `company.go:225-443`

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
main.go                           # Main application & API endpoints
internal/company/company.go       # DBF file operations
internal/auth/auth.go            # User authentication
internal/database/database.go    # SQLite operations

frontend/src/
├── App.jsx                      # Main application component
├── components/
│   ├── BankingSection.jsx      # Banking module (CRITICAL)
│   ├── BankReconciliation.jsx  # Bank reconciliation
│   ├── DBFExplorer.jsx         # DBF file viewer/editor
│   ├── OutstandingChecks.jsx   # Enhanced outstanding checks (NEW)
│   ├── OutstandingChecksSimple.jsx # DataTable usage example (NEW)
│   └── ui/
│       ├── data-table.jsx      # Reusable data table component (NEW)
│       └── [other ShadCN UI components]
└── wailsjs/go/main/App.js      # Generated Wails bindings
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

### Company Data (DBF Files)
- Located in `../datafiles/{company_name}/`
- COA.dbf: Chart of Accounts
- CHECKS.dbf: Check transactions
- INCOME.dbf: Revenue records
- EXPENSE.dbf: Expense records
- WELLS.dbf: Well information

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

**Last Updated**: August 5, 2025
**Key Fix Applied**: Changed data structure key from "data" to "rows" for DBF file reading
**Latest Enhancement**: Enhanced Outstanding Checks with enterprise-level data management and created reusable DataTable component
**Major Features Added**:
- **Outstanding Checks Enhancement**: Account filtering, row selection, pagination, sorting, search, edit modals
- **Reusable DataTable Component**: Standard pattern for all data lists with full CRUD capabilities
- **Standard Data List Pattern**: Established consistent UX across all data management features
**Status**: Outstanding Checks fully functional with enterprise features, DataTable component ready for system-wide use