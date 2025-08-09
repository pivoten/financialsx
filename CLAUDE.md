# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Pivoten Financials X - Enhanced Legacy is a modern desktop companion app to the legacy Visual FoxPro Accounting Manager. It uses Wails with a Go backend + React frontend, mines DBF data and persists derived data in SQLite, and supports new reporting features like State Reporting.

## Commands

### Development
- **Run in dev mode**: `cd desktop && wails dev`
- **Build app**: `cd desktop && wails build`
- **Run tests**: `go test ./...`
- **Format code**: `go fmt ./...`
- **Lint**: `golangci-lint run`

### Frontend (from desktop/frontend directory)
- **Install dependencies**: `npm install`
- **Build frontend**: `npm run build`

## Project Structure

```
financialsx/
├── desktop/                  # Wails desktop application
│   ├── main.go              # Entry point with Wails setup
│   ├── go.mod               # Module: github.com/pivoten/financialsx/desktop
│   ├── wails.json           # Wails configuration
│   ├── build/               # Build configuration and assets
│   └── frontend/            # React + Vite frontend
│       ├── src/
│       │   ├── main.js      # Frontend entry point
│       │   ├── app.css      # Application styles
│       │   └── assets/      # Images and fonts
│       └── wailsjs/         # Generated Wails bindings
│           ├── go/          # Go struct bindings
│           └── runtime/     # Wails runtime API
└── go.mod                   # Root module: github.com/pivoten/financialsx
```

## Architecture Notes

- **Wails**: Desktop framework providing Go backend with React frontend
- **DBF Integration**: Reads legacy Visual FoxPro DBF files for GL and check data
- **SQLite**: Local database for persisting derived data and balance caching
- **Authentication**: JWT-based flow implemented with role-based permissions
- **Balance Caching**: High-performance SQLite cache for bank account balances with outstanding checks calculation

## Bank Balance Caching System

### Overview
Implemented a comprehensive SQLite-based caching system to solve the 5-minute loading time issue when calculating bank account balances. The system correctly calculates **Bank Balance = GL Balance + Outstanding Checks** to reflect actual spendable funds.

### Financial Logic
- **GL Balance**: Amount recorded in General Ledger when check is written (immediately reduced)
- **Outstanding Checks**: Checks written but not yet cleared by the bank
- **Bank Balance**: Actual spendable funds available at the bank

**Example Scenario**:
- Starting bank balance: $1,200
- Write check for $200 → GL Balance becomes $1,000 (immediate)
- Check hasn't cleared yet → Outstanding Checks: $200
- **Bank Balance: $1,000 + $200 = $1,200** (still spendable until check clears)

### Database Schema

#### account_balances Table
```sql
CREATE TABLE account_balances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_name TEXT NOT NULL,
    account_number TEXT NOT NULL,
    account_name TEXT NOT NULL,
    account_type INTEGER NOT NULL,
    
    -- GL Balance (from GLMASTER.dbf scan)
    gl_balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    gl_last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    gl_record_count INTEGER NOT NULL DEFAULT 0,
    
    -- Outstanding Checks (from CHECKS.dbf scan)
    outstanding_checks_total DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    outstanding_checks_count INTEGER NOT NULL DEFAULT 0,
    outstanding_checks_last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Calculated Bank Balance (GL + Outstanding Checks)
    bank_balance DECIMAL(15,2) GENERATED ALWAYS AS (gl_balance + outstanding_checks_total) STORED,
    
    -- Metadata
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_bank_account BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata JSON DEFAULT '{}',
    
    UNIQUE(company_name, account_number)
);
```

#### balance_history Table
```sql
CREATE TABLE balance_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_balance_id INTEGER NOT NULL,
    company_name TEXT NOT NULL,
    account_number TEXT NOT NULL,
    change_type TEXT NOT NULL CHECK (change_type IN ('gl_refresh', 'checks_refresh', 'manual_adjustment', 'reconciliation')),
    
    -- Before/After values
    old_gl_balance DECIMAL(15,2),
    new_gl_balance DECIMAL(15,2),
    old_outstanding_total DECIMAL(15,2),
    new_outstanding_total DECIMAL(15,2),
    old_bank_balance DECIMAL(15,2),
    new_bank_balance DECIMAL(15,2),
    
    -- Change details
    change_reason TEXT,
    changed_by TEXT,
    change_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata JSON DEFAULT '{}',
    
    FOREIGN KEY (account_balance_id) REFERENCES account_balances(id) ON DELETE CASCADE
);
```

### Backend Implementation

#### Key Functions (`internal/database/balance_cache.go`)
- `InitializeBalanceCache(db *DB)` - Creates tables and views
- `GetCachedBalance(db *DB, companyName, accountNumber)` - Retrieves single cached balance
- `GetAllCachedBalances(db *DB, companyName)` - Retrieves all cached balances for company
- `RefreshGLBalance(db *DB, companyName, accountNumber, username)` - Updates GL balance from GLMASTER.dbf
- `RefreshOutstandingChecks(db *DB, companyName, accountNumber, username)` - Updates outstanding checks from CHECKS.dbf

#### API Endpoints (`main.go`)
- `GetCachedBalances(companyName)` - Fast retrieval of all cached balances
- `RefreshAccountBalance(companyName, accountNumber)` - Refresh single account
- `RefreshAllBalances(companyName)` - Refresh all accounts for company
- `GetBalanceHistory(companyName, accountNumber, limit)` - Audit trail of balance changes

### Frontend Integration

#### Banking Section (`BankingSection.jsx`)
- Displays three-tier balance information:
  1. **GL Balance** - From General Ledger
  2. **Uncleared Checks** - Outstanding check total (shown in red)
  3. **Bank Balance** - Calculated spendable amount (GL + Outstanding)
- Individual refresh buttons for each account
- "Refresh All" button for bulk updates
- Visual indicators for stale data
- Real-time loading states

#### Features
- **Fast Loading**: Cached balances load instantly vs. 5-minute GL scan
- **Freshness Tracking**: Visual indicators for data age (fresh/aging/stale)
- **Manual Refresh**: Individual and bulk refresh capabilities
- **Audit Trail**: Complete history of balance changes
- **Error Handling**: Graceful fallbacks and user feedback

### Performance Benefits
- **Before**: 5+ minutes to load bank account balances (scanning entire GL)
- **After**: Instant loading of cached balances (<1 second)
- **Refresh**: On-demand updates only when needed
- **Scalability**: Handles large datasets efficiently with SQLite indexing

## Development Workflow

1. The desktop app runs from the `desktop/` directory
2. Frontend changes auto-reload in dev mode
3. Backend changes require restart of `wails dev`
4. Generated bindings in `wailsjs/` should not be edited manually

## Completed Features

✅ **Bank Balance Caching System** - High-performance SQLite cache for instant balance loading
✅ **DBF Integration** - Complete integration with GLMASTER.dbf and CHECKS.dbf reading
✅ **Authentication System** - JWT-based flow with role-based permissions (Admin/Root/Read-Only)
✅ **Banking Module** - Full bank account management with real-time GL balances
✅ **Outstanding Checks** - Enhanced management with account filtering, pagination, and editing
✅ **Balance Audit Trail** - Complete history tracking of all balance changes
✅ **User Management** - SQLite-based user system with company-specific access
✅ **Bank Reconciliation System** - Complete SQLite-based reconciliation with intelligent workflow

## CI/CD
* This project uses github actions to produce builds to help distribute the application. Make sure `.github/workflows/build.yml` is kept up to date with the proper versions and build processes.
* These two versions in the build step should really be looked at and made sure they are right `go-version`, `node-version` and the wails go package version

## Bank Reconciliation System

### Overview
Advanced bank reconciliation interface with real-time calculations, intelligent auto-save, and intuitive user experience. Replaces manual reconciliation process with modern SQLite-based workflow.

### Key Features
- **Smart Date Pre-population**: Automatically calculates next statement date (end of following month)
- **Auto-calculation**: Ending balance computed from Beginning + Credits - Debits (triggered onBlur)
- **Real-time Reconciliation**: Live balance tracking as checks are selected
- **Transaction Type Display**: Visual badges showing Deposits (green) vs Checks (blue) across all tables
- **Intelligent Auto-save**: 10-second debounced save with draft persistence
- **Conditional Commit**: Green commit button only appears when reconciliation is perfectly balanced
- **Progressive Disclosure**: Clean UI that reveals options as needed

### User Experience Improvements
- **Intuitive Selection**: Single-click checkboxes immediately add items to reconciliation
- **Live Calculations**: 
  - Statement Credits: Shows selected deposit amounts (starts at $0.00)
  - Statement Debits: Shows selected check amounts (starts at $0.00)
  - Calculated Balance: Beginning Balance +/- selected net amounts
  - Balance Difference: Real-time gap between statement and calculated balance
- **Visual Feedback**: Row highlighting, status badges, and color-coded amounts
- **Performance Optimized**: Removed calculation lag during typing, fast checkbox responses

### Technical Implementation
- **SQLite Draft System**: Auto-saves reconciliation progress with 10-second debounce
- **CIDCHEC Integration**: Uses unique check IDs for reliable cross-session tracking
- **Sequential Data Loading**: Prevents saved draft values from being overwritten during load
- **Optimized Rendering**: Reduced React re-renders for smooth typing experience
- **Type-aware Calculations**: Separates deposits from checks for accurate reconciliation math

### Components Enhanced
- `BankReconciliation.jsx`: Complete reconciliation interface with all improvements
- `getTransactionTypeBadge()`: Unified badge component for transaction type display
- `calculateTotals()`: Enhanced calculation engine with real-time balance tracking
- Auto-save system with intelligent debouncing and draft persistence

## UI/UX Enhancements

### Collapsible Sidebar Navigation
The main application sidebar can be toggled between collapsed and expanded states to maximize screen space for data-intensive views like bank reconciliation.

**Implementation Details**:
- **Location**: `App.jsx` - `AdvancedDashboard` component
- **Toggle Button**: Located in top-right of sidebar header (Menu icon when collapsed, ChevronLeft when expanded)
- **Collapsed Width**: 4rem (64px) - shows only icons with tooltips
- **Expanded Width**: 16rem (256px) - shows icons and labels
- **State Management**: `isSidebarCollapsed` state in `AdvancedDashboard` component
- **Responsive Design**: Hidden on mobile (lg:flex), always visible on desktop
- **Smooth Transition**: 300ms CSS transition for width changes

**Benefits**:
- Maximizes horizontal space for wide tables and data grids
- Particularly useful for bank reconciliation with side-by-side check matching
- Maintains navigation accessibility while prioritizing content space
- Smooth animation provides visual feedback during state changes

### Direct Account Reconciliation Access
Bank reconciliation can be accessed directly from bank account cards via the three-dot menu, eliminating the need to navigate through multiple tabs.

**Implementation**:
- **Location**: `BankingSection.jsx` - Bank account cards
- **Access Point**: Three-dot dropdown menu on each bank account card
- **Direct Navigation**: Clicking "Reconcile" immediately opens reconciliation for that specific account
- **No Account Selection**: Since accessed from a specific card, account is pre-selected

## Next Steps

- **Bank Statement Import** - Integrate with bank statement imports for automated reconciliation
- **State Reporting** - Implement state-specific financial reporting behind feature flags
- **Data Export** - Enhanced CSV/PDF export capabilities for all financial data
- **Advanced Analytics** - Cash flow forecasting and trend analysis
- **Mobile Responsive** - Optimize UI for tablet/mobile access