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

## Performance Optimizations

### Startup Caching
The application caches frequently-used values at startup to improve performance:

#### Platform Detection
- **Cached at startup**: `platform` and `isWindows` fields in App struct
- **Benefits**: Avoids repeated `runtime.GOOS` calls throughout the application
- **Access**: Available via `GetPlatform()` API for frontend if needed

#### Authentication State Caching
- **Cached on login**: Authentication and permission states
- **Fields cached**:
  - `isAuthenticated` - Whether user is logged in
  - `isAdmin` - Admin privileges flag
  - `isRoot` - Root privileges flag
  - `userRole` - Role name string
  - `permissions` - Map of permission strings to boolean values
- **Benefits**: 
  - Permission checks are O(1) map lookups instead of repeated function calls
  - Reduces overhead on every API call that checks permissions
  - Especially beneficial for frequently-checked permissions
- **Updates**: Cache is refreshed on login/logout via `updateAuthCache()`
- **Access**: Available via `GetAuthState()` API for frontend

#### Helper Methods
- `hasPermission(permission string)`: Fast cached permission check
- `updateAuthCache()`: Refreshes all cached auth values from current user

### Recommended Patterns
When adding new features that require frequent checks:
1. Cache the value at appropriate lifecycle points (startup, login, etc.)
2. Create helper methods for cached access
3. Update cache when state changes
4. Expose via API if frontend needs access

## Architecture Notes

- **Wails**: Desktop framework providing Go backend with React frontend
- **DBF Integration**: Reads legacy Visual FoxPro DBF files for GL and check data
- **SQLite**: Local database for persisting derived data and balance caching
- **Authentication**: JWT-based flow implemented with role-based permissions
- **Balance Caching**: High-performance SQLite cache for bank account balances with outstanding checks calculation

## UI/UX Design System

### Standard Dashboard Layout (Established December 2024)
All new dashboard sections should follow the clean, modern design pattern established in the Banking Section. This provides consistency and professional appearance across the application.

#### Core Layout Structure
```tsx
<div className="bg-white rounded-lg shadow-sm">
  <Tabs>
    {/* Tab Navigation */}
    <div className="border-b border-gray-200">
      <TabsList className="flex h-12 items-center justify-start space-x-8 px-6 bg-transparent">
        <TabsTrigger className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all 
                               data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 
                               data-[state=inactive]:hover:text-gray-700 
                               data-[state=active]:after:absolute data-[state=active]:after:bottom-0 
                               data-[state=active]:after:left-0 data-[state=active]:after:right-0 
                               data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600">
          Tab Name
        </TabsTrigger>
      </TabsList>
    </div>
    
    {/* Content Area */}
    <TabsContent className="p-6">
      {/* Section Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-xl font-semibold text-gray-900">Section Title</h2>
          <p className="text-sm text-gray-500 mt-1">Description text</p>
        </div>
        <Button variant="outline" className="border-gray-200 hover:bg-gray-50">
          Action Button
        </Button>
      </div>
      
      {/* Content Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card className="border border-gray-200 hover:shadow-md transition-all bg-white">
          <CardHeader className="pb-4 border-b border-gray-100">
            {/* Card header content */}
          </CardHeader>
          <CardContent className="p-4">
            {/* Card body content */}
          </CardContent>
        </Card>
      </div>
    </TabsContent>
  </Tabs>
</div>
```

#### Design Principles
1. **Color Palette**
   - **Backgrounds**: White (`bg-white`) for all content areas
   - **Text**: Gray-900 for headings, gray-500 for descriptions/labels
   - **Borders**: Gray-200 for primary borders, gray-100 for subtle dividers
   - **Accents**: Blue-600 for active states and primary actions
   - **Status Colors**: Green-600 (positive), Red-600 (negative), Amber-600 (warning)

2. **Spacing Standards**
   - **Container padding**: `p-6` for main content areas
   - **Card padding**: `p-4` for card content
   - **Section spacing**: `mb-6` between major sections
   - **Element spacing**: `space-y-3` or `space-y-4` for vertical lists

3. **Interactive Elements**
   - **Buttons**: `variant="outline"` with `border-gray-200 hover:bg-gray-50`
   - **Cards**: `hover:shadow-md transition-all` for hover effects
   - **Tabs**: Blue underline for active state, gray text for inactive

4. **Typography**
   - **Page titles**: `text-xl font-semibold text-gray-900`
   - **Section headers**: `text-lg font-semibold text-gray-900`
   - **Card titles**: `text-base font-semibold text-gray-900`
   - **Descriptions**: `text-sm text-gray-500`
   - **Labels**: `text-sm text-gray-500`
   - **Values**: `text-sm text-gray-900` (normal), `font-medium` (emphasized)

5. **Layout Patterns**
   - Use responsive grid layouts: `grid gap-4 md:grid-cols-2 lg:grid-cols-3`
   - Maintain consistent card heights with `h-full flex flex-col`
   - Use flexbox for header sections with actions: `flex items-center justify-between`

#### Components to Follow This Pattern
- ✅ Banking Section (reference implementation)
- 🔄 Reports Section (to be updated)
- 🔄 Analytics Dashboard (to be updated)
- 🔄 User Management (to be updated)
- 🔄 Settings Pages (to be updated)
- 🔄 Any new dashboard sections

#### Drag and Drop Support
For sortable lists (like bank account cards), use `@dnd-kit/sortable`:
- Drag handle: `GripVertical` icon from lucide-react
- Visual feedback: `opacity: 0.5` when dragging
- Persistence: Save order to localStorage with company-specific keys

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
✅ **Visual FoxPro Integration** - TCP socket communication with legacy VFP application using Winsock2 API
✅ **SherWare Legacy Dashboard** - 260+ VFP forms organized with drag-drop, search, and categories

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

## Company Data Location

### Automatic Discovery
The application uses recursive search to find `compmast.dbf`:
1. Searches from current working directory (max depth: 5)
2. If not found, searches from parent directory (max depth: 3)
3. If not found, searches from executable directory (max depth: 3)
4. If not found, checks previously saved data path

### Manual Folder Selection
If `compmast.dbf` cannot be found automatically:
1. User sees "Cannot find company data files" message
2. "Select Data Folder" button appears
3. User selects folder containing `compmast.dbf`
4. Path is validated and saved for future sessions
5. Company folders are resolved relative to this location

### Data Path Persistence
- Selected path saved to: `{OS_TEMP_DIR}/financialsx_datapath.txt`
- Automatically reused in future sessions
- Company data paths resolved relative to `compmast.dbf` location

### Platform-Specific Path Resolution (IMPORTANT)

The application handles company paths differently based on the operating system:

#### macOS/Linux
- **Company folders are ALWAYS relative to the compmast.dbf location**
- When reading CDATAPATH from compmast.dbf, only the folder name is extracted
- Example: If CDATAPATH is `c:\program files\data\company1\`, only `company1` is used
- The actual path becomes: `{compmast_directory}/company1/`
- This ensures portability when DBF files are created on Windows but used on Mac/Linux

#### Windows
- **Uses the full path from CDATAPATH field**
- Supports both absolute paths (e.g., `C:\DataFiles\Company1\`) and relative paths
- Relative paths are resolved from the executable directory
- Maintains compatibility with legacy Visual FoxPro systems

#### Implementation Details
- **Platform detection at startup**: The application detects the OS platform once during startup and stores it in the App struct
- **Cached platform variables**: The `internal/company` package caches `isWindows` and `platform` variables to avoid repeated `runtime.GOOS` checks
- **Path normalization**: The `normalizeCompanyPath()` function in `internal/company/company.go` handles platform-specific path conversion
- **Consistent application**: Applied across all DBF operations:
  - `GetDBFFiles()` - Lists DBF files in company directory
  - `ReadDBFFile()` - Reads DBF file contents
  - `CreateCompanyDirectory()` - Creates SQL folder structure
  - `GetCompanyList()` - Returns normalized paths for frontend
- **SQL folder location**: Ensures SQL folder and SQLite database are created in the correct location based on platform
- **Frontend access**: Platform info available via `GetPlatform()` API for frontend if needed

## Bill Entry System (NEW - August 2025)

### Overview
Modern implementation of the FoxPro AP Bill Entry screen with enhanced validation and user experience.

### Features
- **Two Implementations**:
  - Basic version (`BillEntry.tsx`) - Traditional React
  - Enhanced version (`BillEntryEnhanced.tsx`) - React Hook Form + Zod + React Query
- **Comprehensive Form Management**: 
  - Vendor selection with lookup
  - Invoice details with automatic terms calculation
  - Dynamic line item management
  - Real-time validation and error feedback
- **Modern Architecture**:
  - Type-safe with TypeScript and Zod
  - Optimistic updates with React Query
  - Performance optimized with React Hook Form

### Access
**Navigation**: Financials → Accounts Payable (formerly Transactions)

### Backend Integration (Pending)
- Connect to APPURCHH.dbf (bill headers) and APPURCHD.dbf (line items)
- Vendor lookup from VENDOR.dbf
- Account lookup from COA.dbf
- Well lookup from WELLS.dbf

## User Profile System (NEW - August 2025)

### Overview
Comprehensive user profile management interface ready for Supabase integration.

### Features
- **Personal Information**: Editable profile fields
- **Security Settings**: Password change, 2FA setup
- **Notification Preferences**: Email notification controls
- **Display Preferences**: Theme, date format, regional settings
- **Avatar System**: Profile picture with initials fallback

### Access
- **Primary**: Click on your email in the sidebar
- **Secondary**: Settings → Profile card
- **Direct**: Settings menu → My Profile

## Visual FoxPro Integration System

### Overview
Bidirectional communication between FinancialsX and legacy Visual FoxPro application using TCP sockets with Winsock2 API (no dependencies required).

### Features
- **TCP Socket Communication**: NDJSON protocol on port 23456
- **Company Synchronization**: Ensures both apps have same company open
- **Form Launching**: Launch any VFP form from FinancialsX
- **No Dependencies**: Uses native Windows Sockets 2 API (no OCX/registration)
- **Settings Management**: Configure host, port, timeout from UI

### SherWare Legacy Dashboard
- **260+ Forms**: Complete VFP form library organized by category
- **Drag & Drop**: Reorder forms with persistent localStorage
- **Cross-Category Search**: Universal search across all forms
- **Categories**: GL, AP, AR, Cash Management, Oil & Gas, etc.
- **Quick Access**: Frequently used forms in dedicated section

### Technical Stack
- **Backend**: Go VFP client in `internal/vfp/vfp_integration.go`
- **Frontend**: React components with @dnd-kit for drag-drop
- **FoxPro**: Winsock2Listener class (no ActiveX required)
- **Protocol**: NDJSON over TCP with company context

### API Endpoints
- `GetVFPSettings()` - Retrieve connection settings
- `SaveVFPSettings()` - Update connection configuration
- `TestVFPConnection()` - Verify connectivity
- `LaunchVFPForm()` - Launch form with company sync
- `SyncVFPCompany()` - Synchronize company between apps
- `GetVFPCompany()` - Get current VFP company

## Chart of Accounts Report System

### Overview
Professional PDF generation for Chart of Accounts with filtering, sorting, and company branding.

### Features
- **PDF Generation**: High-quality PDF reports using gofpdf library
- **Active/Inactive Filter**: Filter accounts by LINACTIVE field status
- **Sorting Options**: Sort by account number or account type
- **Company Branding**: Pulls company info from VERSION.DBF
- **Professional Layout**: Landscape orientation with headers, footers, and table formatting
- **Native Save Dialog**: OS-integrated file save with sanitized filenames

### Technical Implementation
- **Backend**: `GenerateChartOfAccountsPDF()` in main.go
- **Frontend**: `ChartOfAccountsReport.tsx` component
- **Data Source**: COA.dbf for accounts, VERSION.DBF for company info
- **PDF Library**: github.com/jung-kurt/gofpdf
- **Filename Format**: YYYY-MM-DD - {Company Name} - Chart of Accounts.pdf

## Batch Flow Visualization System

### Overview
Interactive flow chart visualization for tracing batch numbers through the complete accounting cycle, from check payment to original purchase entry.

### Features
- **Visual Flow Chart**: Displays the complete transaction flow with connected nodes
- **Clickable Cards**: Each table node is clickable to view full record details
- **Consistent Card Sizing**: All cards maintain uniform 240px × 140px dimensions
- **Modal Record Viewer**: Scrollable modal displays all records with proper formatting
- **Color Coding**: Blue borders for tables with data, gray for empty tables
- **Bidirectional Tracing**: Shows both payment and purchase GL entries

### Flow Structure
1. **CHECKS.DBF** - Check entry point (CBATCH)
2. **GLMASTER.DBF** - Check payment GL entry
3. **APPMTHDR.DBF** | **APPMTDET.DBF** - Payment header and details (side by side)
4. **GLMASTER.DBF** - Original purchase GL entry (CSOURCE = 'AP')
5. **APPURCHH.DBF** | **APPURCHD.DBF** - Purchase header and details (side by side)

### Key Concepts
- **CBATCH**: Batch number that links related transactions
- **CBILLTOKEN**: Critical field linking payment side to entry side
- **Dual GL Entries**: Shows both payment posting and original purchase posting

### Technical Implementation
- **Components**: 
  - `FollowBatchNumber.tsx` - Main search interface with clickable result cards
  - `BatchFlowChart.tsx` - Interactive flow chart visualization
- **Backend**: `FollowBatchNumber()` API searches across multiple DBF files
- **Modal Features**: 
  - Max height 90vh with scrollable content area
  - Grid layout for field display
  - Formatted dates and currency values
- **Visual Design**: Consistent card sizing, hover effects, and professional styling

### User Experience
- **Search History**: Recent batch searches with dropdown selection
- **Update Batch Details**: Bulk field updates across related tables
- **Visual Indicators**: "Click to view records" prompt on cards with data
- **Responsive Layout**: Side-by-side display for header/detail relationships

## Next Steps

- **Bill Entry Backend**: Create Go API endpoints for bill CRUD operations
- **DBF Integration**: Connect bills to APPURCHH.dbf and APPURCHD.dbf
- **Vendor Management**: Implement vendor lookup and quick-add
- **Bank Statement Import** - Integrate with bank statement imports for automated reconciliation
- **State Reporting** - Implement state-specific financial reporting behind feature flags
- **Data Export** - Enhanced CSV/PDF export capabilities for all financial data
- **Advanced Analytics** - Cash flow forecasting and trend analysis
- **Mobile Responsive** - Optimize UI for tablet/mobile access