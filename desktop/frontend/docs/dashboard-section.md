# Dashboard Section Documentation

## Overview
The Dashboard is the main landing page after login, providing an overview of financial data and quick access to key metrics.

## Stats Grid
The dashboard displays 4 key metric cards in a responsive grid:

### Card Structure
```javascript
const stats = [
  {
    title: "Total Revenue",
    value: "$1,245,890",
    change: "+12.5%",
    icon: DollarSign,
    trend: "up"
  },
  {
    title: "Active Wells",
    value: "42",
    change: "+2",
    icon: Activity,
    trend: "up"
  },
  {
    title: "Outstanding Checks",
    value: "156",
    change: "-8",
    icon: FileText,
    trend: "down"
  },
  {
    title: "Bank Balance",
    value: "$523,450",
    change: "+5.2%",
    icon: TrendingUp,
    trend: "up"
  }
]
```

### Visual Design
- Each stat card shows:
  - Title (small, muted text)
  - Main value (large, bold)
  - Change indicator with percentage/amount
  - Icon in top-right corner
  - Color-coded trend (green for up, red for down)

## Quick Actions Section
Grid of action buttons for common tasks:
- New Transaction
- Generate Report
- Bank Reconciliation
- Import Data
- Export Data
- User Management

## Recent Activity
Table or list showing recent transactions and activities:
- Date/Time
- Description
- User
- Amount (if applicable)
- Status badge

## Charts Section
Two main charts displayed side-by-side:

### Revenue Chart
- Line/Area chart showing revenue over time
- Monthly view by default
- Toggle for weekly/monthly/yearly

### Well Production Chart
- Bar chart showing production by well
- Top 10 wells displayed
- Color-coded by status

## Company Overview Card
Shows current company information:
- Company name
- Current period
- Last sync time
- Database status
- Active users count

## Loading States
- Skeleton loaders for cards while data loads
- Animated placeholders for charts
- Progressive loading (show what's ready first)

## Data Sources
Dashboard data should be fetched from:
- `GetDashboardData(companyName)` - Main dashboard metrics
- `GetRecentActivity(companyName)` - Recent transactions
- `GetCompanyInfo(companyName)` - Company overview

## Responsive Design
- Mobile: Single column, stacked cards
- Tablet: 2-column grid
- Desktop: 4-column grid for stats, 2-column for charts

## Color Coding
- Positive changes: Green (#10b981)
- Negative changes: Red (#ef4444)
- Neutral/Info: Blue (#3b82f6)
- Warning: Yellow (#f59e0b)

## Implementation Notes
- Dashboard should auto-refresh every 5 minutes
- Use React Query or SWR for data fetching and caching
- Implement error boundaries for failed data loads
- Show relative timestamps (e.g., "5 minutes ago")
- All monetary values should use proper formatting with commas