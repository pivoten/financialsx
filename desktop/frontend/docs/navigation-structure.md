# Navigation Structure Documentation

## Original Navigation Design (Main Branch)

### Primary Sidebar Navigation
The main navigation has the following top-level sections:

1. **Dashboard** (Home icon)
2. **Operations** (Activity icon)
3. **Financials** (DollarSign icon)
4. **Data** (Database icon)
5. **Reporting** (FileText icon)
6. **Utilities** (Wrench icon)
7. **Settings** (Settings icon)

### Collapsible Sidebar Features
- Sidebar can be collapsed/expanded with toggle button
- Shows icons only when collapsed
- Tooltips appear on hover when collapsed
- Smooth hover expansion when collapsed

## Section Details

### 1. Dashboard
- **Icon**: Home
- **Description**: "Overview of your financial data"
- **Content**: 
  - Stats Grid (4 cards with metrics)
  - Recent activity
  - Quick actions
  - Company overview

### 2. Operations
- **Icon**: Activity
- **Description**: "Manage wells, production, and field operations"
- **Subsections**:
  - Wells Management (Activity icon)
  - Production Tracking (TrendingUp icon)
  - Field Operations (Wrench icon)
  - Maintenance (Settings icon)

### 3. Financials
- **Icon**: DollarSign
- **Description**: "Financial transactions, analytics, and accounting"
- **Subsections**:
  - Transactions (DollarSign icon)
  - Revenue Analysis (TrendingUp icon)
  - Expense Management (Calculator icon)
  - Financial Analytics (BarChart3 icon)
  - Banking (Home icon)
  - Accounting Tools (FileText icon)
  - Settings (Settings icon)

### 4. Data Management
- **Icon**: Database
- **Description**: "Database maintenance and data management"
- **Subsections**:
  - DBF Explorer (Database icon)
  - Import Data (Upload icon)
  - Export Data (Download icon)
  - Backup & Restore (Archive icon)
  - Database Maintenance (Wrench icon)

### 5. Reporting
- **Icon**: FileText
- **Description**: "Reports, compliance, and documentation"
- **Subsections**:
  - State Reports (FileText icon)
  - Financial Reports (DollarSign icon)
  - Production Reports (TrendingUp icon)
  - Custom Reports (FileSearch icon)
  - Audit Trail (Copy icon)

### 6. Utilities
- **Icon**: Wrench
- **Description**: "Tools, calculators, and system utilities"
- **Subsections**:
  - Calculators (Calculator icon)
  - Unit Converter (Activity icon)
  - Task Scheduler (Calendar icon)
  - Data Tools (Wrench icon)

### 7. Settings
- **Icon**: Settings
- **Description**: "System configuration and user management"
- **Subsections**:
  - User Management (Users icon)
  - Appearance (Settings icon)
  - System Configuration (Database icon)
  - Security Settings (FileText icon)

## Navigation Implementation Details

### Header Navigation
- Shows dropdown menu when in any section other than Dashboard
- Dropdown button shows current section icon and name
- Dropdown contains all subsections for current section
- Page title shows current section/subsection
- User info and company name shown in top-right

### Dashboard Cards
Each section (except Dashboard) shows a grid of cards when no subsection is selected:
- Cards are clickable and navigate to subsections
- Cards show icon, title, and description
- Hover effect with shadow and slight scale
- Cards organized in responsive grid (2-4 columns)

### State Management
- `activeView`: Current main section (dashboard, operations, financials, etc.)
- `activeSubView`: Current subsection within a section
- Navigation resets `activeSubView` when changing main sections

## Key UI Components Used
- `SidebarNavItem`: Custom component for sidebar items
- `DropdownMenu`: For section navigation in header
- `Card`: For dashboard cards and content sections
- `Button`: For actions and navigation
- Icons from `lucide-react`

## Color Scheme
- Primary color: Orange (#F5981E) - from Pivoten theme
- Background: Gray-50 for main content area
- Sidebar: White background with gray borders
- Active items: Orange background with orange text
- Hover effects: Gray background for inactive items