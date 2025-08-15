# DBF CRUD Standard Implementation Guide

## Overview
This guide establishes the standard pattern for implementing Create, Read, Update, and Delete operations for DBF files in the FinancialsX system. We provide two approaches: **Dynamic (recommended)** and Traditional.

## Technology Stack
- **Frontend**: React with TypeScript
- **UI Components**: ShadCN UI
- **Backend**: Go with Wails framework
- **DBF Library**: go-dbase v1.12.10
- **Features**: Configurable columns, Searchable, Sortable, Row click for details, Persistent preferences

## Approach 1: Dynamic DBF Table (RECOMMENDED)

### Overview
The DynamicDBFTable component automatically generates the UI based on the actual DBF structure at runtime. This approach requires **ZERO code changes** when DBF structures change.

### Key Benefits
- **Zero Maintenance**: DBF structure changes are automatically handled
- **One Component for All Tables**: Reuse for any DBF file
- **Persistent Configuration**: User column preferences saved per table/company
- **Smart Type Detection**: Automatically uses appropriate input controls
- **No Hardcoded Fields**: Discovers columns at runtime

### Usage

```tsx
import DynamicDBFTable from './components/DynamicDBFTable'

// Basic usage - fully automatic
<DynamicDBFTable
  tableName="VENDOR.dbf"
  companyName={companyName}
/>

// With customization
<DynamicDBFTable
  tableName="VENDOR.dbf"
  companyName={companyName}
  title="Vendor Management"
  description="Manage vendor records"
  canEdit={true}
  primaryFields={['CVENDORID', 'CVENDNAME', 'CPHONE']}
  maxTableColumns={6}
/>
```

### Column Configuration Features

#### Persistent Storage
- Column visibility and order saved to localStorage
- Settings persist per company and table
- Key format: `dbf_columns_{company}_{table}`

#### Configuration Modal
Users can:
- **Show/Hide Columns**: Checkbox for each column
- **Reorder Columns**: Up/down arrows to change order
- **Quick Actions**: Show All, Hide All, Reset to Defaults
- **View Metadata**: See field names, types, and display names

#### Smart Defaults
The component automatically:
- Detects column types from prefixes (L=boolean, D=date, N=number)
- Formats column headers (CVENDNAME → "Vendor Name")
- Identifies important fields to show by default
- Limits table display to 6 columns (configurable)

### Implementation Example

```tsx
export default function VendorManagementDynamic({ companyName, currentUser }: Props) {
  // Define primary fields to show by default
  const primaryFields = [
    'CVENDORID',
    'CVENDNAME', 
    'CCONTACT',
    'CPHONE',
    'CEMAIL',
    'LINACTIVE'
  ]
  
  return (
    <DynamicDBFTable
      tableName="VENDOR.dbf"
      companyName={companyName}
      title="Vendor Management"
      description="Manage vendor information and records"
      canEdit={currentUser?.is_admin || currentUser?.is_root}
      primaryFields={primaryFields}
      maxTableColumns={6}
    />
  )
}
```

### Type Detection Rules

| Prefix | Type | UI Control | Display Format |
|--------|------|------------|----------------|
| L | boolean | Checkbox | Badge (Yes/No) |
| D | date | Date picker | MM/DD/YYYY |
| N | number | Number input | Numeric |
| N + AMOUNT/BALANCE | currency | Number input | $1,234.56 |
| C (default) | text | Text input | Text |
| Sensitive fields* | encrypted | Password input with toggle | •••••••• / actual value |

*Sensitive fields are auto-detected based on field names containing: PASSWORD, SSN, TAXID, SECRET, TOKEN, CREDIT, CARD, PIN, CVV

### Encrypted/Sensitive Field Handling

The DynamicDBFTable component automatically detects and handles sensitive fields:

#### Detection
Fields are marked as sensitive if their names contain patterns like:
- PASSWORD, PASSWD, PWD
- SSN, SOCIAL
- TAXID, TAX_ID, CTAXID
- SECRET, KEY, TOKEN
- CREDIT, CARD
- ACCOUNT, ACCT
- PIN, CVV

#### Display Features
- **Table View**: Shows "••••••••" by default with Lock/Unlock icon
- **Toggle Visibility**: Click icon to reveal/hide actual value
- **Modal View**: Password input field with Eye/EyeOff toggle button
- **Independent States**: Each field instance has its own decrypt state

#### Implementation
```tsx
// The component tracks decrypted fields in state
const [decryptedFields, setDecryptedFields] = useState<Set<string>>(new Set())

// Each field gets a unique key for tracking
const fieldKey = `${recordIndex}-${fieldName}` // In table
const fieldKey = `modal-${fieldName}` // In modal

// Toggle decrypt/encrypt for a field
const toggleFieldDecryption = (fieldKey: string) => {
  const newDecrypted = new Set(decryptedFields)
  if (newDecrypted.has(fieldKey)) {
    newDecrypted.delete(fieldKey)
  } else {
    newDecrypted.add(fieldKey)
  }
  setDecryptedFields(newDecrypted)
}
```

## Approach 2: Traditional Implementation

## Standard Implementation Pattern

### 1. Backend (Go) Implementation

#### 1.1 Read Operation - Get All Records
```go
// GetVendors returns all vendor records from VENDOR.dbf
func (a *App) GetVendors(companyName string) (map[string]interface{}, error) {
    logger.WriteInfo("GetVendors", fmt.Sprintf("Called for company: %s", companyName))
    
    // Use the standard ReadDBFFile function - it handles all the complexity
    vendorData, err := company.ReadDBFFile(companyName, "VENDOR.dbf", "", 0, 0, "", "")
    if err != nil {
        logger.WriteError("GetVendors", fmt.Sprintf("Error reading VENDOR.dbf: %v", err))
        return nil, fmt.Errorf("error reading vendor data: %v", err)
    }
    
    return vendorData, nil
}
```

**Key Points:**
- Always use `company.ReadDBFFile()` - it's the standard way to read DBF files
- Pass `0, 0` for offset and limit to get all records
- The function returns a map with `"rows"` containing the data and `"columns"` containing field names

#### 1.2 Update Operation - Modify a Record
```go
func (a *App) UpdateVendor(companyName string, vendorIndex int, vendorData map[string]interface{}) error {
    logger.WriteInfo("UpdateVendor", fmt.Sprintf("Updating vendor at index %d", vendorIndex))
    
    // Get the datafiles path
    datafilesPath, err := company.GetDatafilesPath()
    if err != nil {
        return fmt.Errorf("failed to get datafiles path: %w", err)
    }
    
    // Normalize the company path
    normalizedCompanyName := company.NormalizeCompanyPath(companyName)
    
    // Construct the full path to the DBF file
    vendorPath := filepath.Join(datafilesPath, normalizedCompanyName, "VENDOR.dbf")

    // Open the table for writing
    table, err := dbase.OpenTable(&dbase.Config{
        Filename:   vendorPath,
        ReadOnly:   false,
        TrimSpaces: true,
    })
    if err != nil {
        return fmt.Errorf("failed to open VENDOR.dbf for writing: %w", err)
    }
    defer table.Close()

    // Navigate to the specific record
    currentIndex := 0
    var targetRow *dbase.Row
    
    for {
        row, err := table.Next()
        if err != nil {
            if err.Error() == "EOF" {
                return fmt.Errorf("vendor record not found at index %d", vendorIndex)
            }
            return fmt.Errorf("error reading vendor table: %w", err)
        }

        // Skip deleted records
        if row.Deleted {
            continue
        }

        if currentIndex == vendorIndex {
            targetRow = row
            break
        }
        currentIndex++
    }

    // Update the fields in the row
    for fieldName, value := range vendorData {
        field := targetRow.FieldByName(fieldName)
        if field == nil {
            continue // Skip fields that don't exist
        }

        err := field.SetValue(value)
        if err != nil {
            return fmt.Errorf("failed to set field %s: %w", fieldName, err)
        }
    }

    // Write the updated row back to the file
    err = table.WriteRow(targetRow)
    if err != nil {
        return fmt.Errorf("failed to write updated record: %w", err)
    }

    return nil
}
```

### 2. Frontend (React/TypeScript) Implementation

#### 2.1 Component Structure
```tsx
import React, { useState, useEffect, useMemo } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Alert, AlertDescription } from './ui/alert'
import { Badge } from './ui/badge'
import { 
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from './ui/dialog'
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from './ui/table'
import { 
  Search, 
  RefreshCw, 
  ArrowUpDown, 
  ArrowUp, 
  ArrowDown,
  Edit,
  Save,
  X,
  CheckCircle,
  AlertCircle
} from 'lucide-react'
import * as WailsApp from '../../wailsjs/go/main/App'

interface DBFRecord {
  _rowIndex: number
  [key: string]: any
}

interface Props {
  companyName: string
  currentUser?: any
}
```

#### 2.2 State Management
```tsx
export default function VendorManagement({ companyName, currentUser }: Props) {
  // Data state
  const [vendors, setVendors] = useState<DBFRecord[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  
  // Search and filter state
  const [searchTerm, setSearchTerm] = useState('')
  const [sortColumn, setSortColumn] = useState<string>('')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc')
  
  // Edit modal state
  const [editModalOpen, setEditModalOpen] = useState(false)
  const [selectedVendor, setSelectedVendor] = useState<DBFRecord | null>(null)
  const [editedVendor, setEditedVendor] = useState<DBFRecord | null>(null)
  const [saving, setSaving] = useState(false)
  const [saveSuccess, setSaveSuccess] = useState(false)
  
  // Permissions
  const canEdit = currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')
```

#### 2.3 Data Loading
```tsx
  const loadVendors = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const result = await WailsApp.GetVendors(companyName)
      
      if (!result || !result.rows) {
        setError('No vendor data found')
        setVendors([])
        return
      }
      
      // Add row index to each vendor for tracking
      const vendorsWithIndex = result.rows.map((vendor: any, index: number) => ({
        ...vendor,
        _rowIndex: index
      }))
      
      setVendors(vendorsWithIndex)
    } catch (err) {
      setError(err.message || 'Failed to load vendors')
      setVendors([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (companyName) {
      loadVendors()
    }
  }, [companyName])
```

#### 2.4 Search and Sort Implementation
```tsx
  // Process vendors for display (search and sort)
  const processedVendors = useMemo(() => {
    let filtered = [...vendors]
    
    // Apply search filter
    if (searchTerm) {
      const searchLower = searchTerm.toLowerCase()
      filtered = filtered.filter(vendor => 
        Object.values(vendor).some(value => 
          value && value.toString().toLowerCase().includes(searchLower)
        )
      )
    }
    
    // Apply sorting
    if (sortColumn) {
      filtered.sort((a, b) => {
        const aVal = a[sortColumn] || ''
        const bVal = b[sortColumn] || ''
        
        // Handle numeric sorting
        const aNum = parseFloat(aVal)
        const bNum = parseFloat(bVal)
        if (!isNaN(aNum) && !isNaN(bNum)) {
          return sortDirection === 'asc' ? aNum - bNum : bNum - aNum
        }
        
        // String sorting
        const comparison = aVal.toString().localeCompare(bVal.toString())
        return sortDirection === 'asc' ? comparison : -comparison
      })
    }
    
    return filtered
  }, [vendors, searchTerm, sortColumn, sortDirection])
```

#### 2.5 Sort Handler
```tsx
  const handleSort = (column: string) => {
    if (sortColumn === column) {
      // Toggle direction if same column
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      // New column, default to ascending
      setSortColumn(column)
      setSortDirection('asc')
    }
  }

  const getSortIcon = (column: string) => {
    if (sortColumn !== column) {
      return <ArrowUpDown className="ml-2 h-4 w-4 text-gray-400" />
    }
    return sortDirection === 'asc' 
      ? <ArrowUp className="ml-2 h-4 w-4" />
      : <ArrowDown className="ml-2 h-4 w-4" />
  }
```

#### 2.6 Table Implementation with Fixed Header
```tsx
  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Vendor Management</h2>
          <p className="text-muted-foreground">Manage vendor information and records</p>
        </div>
        <Button onClick={loadVendors} disabled={loading}>
          <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {/* Search Bar */}
      <div className="flex gap-4">
        <div className="flex-1 max-w-sm">
          <div className="relative">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input 
              placeholder="Search vendors..." 
              value={searchTerm} 
              onChange={(e) => setSearchTerm(e.target.value)} 
              className="pl-8" 
            />
          </div>
        </div>
      </div>

      {/* Data Table with Fixed Header */}
      <div className="border rounded-lg">
        <div className="max-h-[600px] overflow-auto">
          <Table>
            <TableHeader className="sticky top-0 bg-white z-10 border-b">
              <TableRow>
                <TableHead 
                  className="cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('CVENDNO')}
                >
                  <div className="flex items-center">
                    Vendor #
                    {sortColumn === 'CVENDNO' && getSortIcon('CVENDNO')}
                  </div>
                </TableHead>
                <TableHead 
                  className="cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('CCOMPANY')}
                >
                  <div className="flex items-center">
                    Company Name
                    {sortColumn === 'CCOMPANY' && getSortIcon('CCOMPANY')}
                  </div>
                </TableHead>
                <TableHead 
                  className="cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('CCONTACT')}
                >
                  <div className="flex items-center">
                    Contact
                    {sortColumn === 'CCONTACT' && getSortIcon('CCONTACT')}
                  </div>
                </TableHead>
                <TableHead>Phone</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-8">
                    Loading vendors...
                  </TableCell>
                </TableRow>
              ) : processedVendors.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-8 text-gray-500">
                    No vendors found
                  </TableCell>
                </TableRow>
              ) : (
                processedVendors.map((vendor, index) => (
                  <TableRow
                    key={vendor._rowIndex}
                    className="cursor-pointer hover:bg-gray-50"
                    onClick={() => handleRowClick(vendor)}
                  >
                    <TableCell className="font-medium">
                      {vendor.CVENDNO || '-'}
                    </TableCell>
                    <TableCell>{vendor.CCOMPANY || '-'}</TableCell>
                    <TableCell>{vendor.CCONTACT || '-'}</TableCell>
                    <TableCell>{vendor.CPHONE || '-'}</TableCell>
                    <TableCell>{vendor.CEMAIL || '-'}</TableCell>
                    <TableCell>
                      <Badge
                        variant={vendor.LINACTIVE === false ? 'default' : 'secondary'}
                        className={vendor.LINACTIVE === false 
                          ? 'bg-green-100 text-green-800' 
                          : 'bg-gray-100 text-gray-800'}
                      >
                        {vendor.LINACTIVE === false ? 'Active' : 'Inactive'}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  )
```

#### 2.7 Row Click Handler & Edit Modal
```tsx
  const handleRowClick = (vendor: DBFRecord) => {
    setSelectedVendor(vendor)
    setEditedVendor({ ...vendor })
    setEditModalOpen(true)
    setSaveSuccess(false)
  }

  const handleFieldChange = (field: string, value: any) => {
    if (!editedVendor) return
    setEditedVendor({
      ...editedVendor,
      [field]: value
    })
  }

  const handleSave = async () => {
    if (!editedVendor || !selectedVendor) return
    
    setSaving(true)
    setSaveSuccess(false)
    
    try {
      // Get only the changed fields
      const changes: any = {}
      Object.keys(editedVendor).forEach(key => {
        if (key !== '_rowIndex' && editedVendor[key] !== selectedVendor[key]) {
          changes[key] = editedVendor[key]
        }
      })
      
      if (Object.keys(changes).length === 0) {
        setSaveSuccess(true)
        setSaving(false)
        return
      }
      
      await WailsApp.UpdateVendor(companyName, selectedVendor._rowIndex, changes)
      
      // Update local state
      const updatedVendors = vendors.map(v => 
        v._rowIndex === selectedVendor._rowIndex ? editedVendor : v
      )
      setVendors(updatedVendors)
      setSelectedVendor(editedVendor)
      setSaveSuccess(true)
      
      // Auto-close after success
      setTimeout(() => {
        setEditModalOpen(false)
      }, 1500)
    } catch (err) {
      setError(err.message || 'Failed to save vendor')
    } finally {
      setSaving(false)
    }
  }
```

#### 2.8 Edit Modal Implementation
```tsx
  {/* Edit Modal */}
  <Dialog open={editModalOpen} onOpenChange={setEditModalOpen}>
    <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>
          {canEdit ? 'Edit Vendor' : 'View Vendor'}
        </DialogTitle>
        <DialogDescription>
          {canEdit 
            ? 'Modify vendor information and save changes'
            : 'View vendor information (read-only)'}
        </DialogDescription>
      </DialogHeader>
      
      {editedVendor && (
        <div className="space-y-6 py-4">
          {/* Success Message */}
          {saveSuccess && (
            <Alert className="border-green-200 bg-green-50">
              <CheckCircle className="h-4 w-4 text-green-600" />
              <AlertDescription className="text-green-800">
                Vendor updated successfully!
              </AlertDescription>
            </Alert>
          )}

          {/* Form Fields - Group by sections */}
          <div className="space-y-4">
            <h3 className="text-sm font-semibold text-gray-700 uppercase tracking-wider">
              Basic Information
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label htmlFor="vendno">Vendor Number</Label>
                <Input
                  id="vendno"
                  value={editedVendor.CVENDNO || ''}
                  onChange={(e) => handleFieldChange('CVENDNO', e.target.value)}
                  disabled={!canEdit}
                />
              </div>
              <div>
                <Label htmlFor="company">Company Name</Label>
                <Input
                  id="company"
                  value={editedVendor.CCOMPANY || ''}
                  onChange={(e) => handleFieldChange('CCOMPANY', e.target.value)}
                  disabled={!canEdit}
                />
              </div>
            </div>
          </div>

          {/* Add more sections for Contact, Address, etc. */}
        </div>
      )}
      
      <DialogFooter>
        <Button variant="outline" onClick={() => setEditModalOpen(false)}>
          <X className="mr-2 h-4 w-4" />
          Close
        </Button>
        {canEdit && (
          <Button onClick={handleSave} disabled={saving}>
            {saving ? (
              <>
                <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                Save Changes
              </>
            )}
          </Button>
        )}
      </DialogFooter>
    </DialogContent>
  </Dialog>
```

## Critical: Sticky Table Headers with ShadCN UI

### The Problem
ShadCN UI's `Table` component wraps the HTML table in a div with `overflow-auto`, which breaks sticky header positioning. This is because `position: sticky` requires specific conditions:
1. The sticky element must be a direct child of the scrolling container
2. There can only be ONE scrolling context between the sticky element and its scroll container

### What DOESN'T Work
```jsx
// ❌ WRONG - ShadCN Table component breaks sticky headers
<div className="overflow-auto h-[600px]">
  <Table>  {/* This adds ANOTHER div with overflow-auto! */}
    <TableHeader className="sticky top-0">  {/* Won't stick properly */}
      <TableRow>
        <TableHead>...</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>...</TableBody>
  </Table>
</div>
```

The ShadCN `Table` component source shows why:
```jsx
// From ShadCN's table.tsx
const Table = React.forwardRef(({ className, ...props }, ref) => (
  <div className="relative w-full overflow-auto">  {/* This breaks sticky! */}
    <table ref={ref} className={cn("w-full caption-bottom text-sm", className)} {...props} />
  </div>
))
```

### What DOES Work
```jsx
// ✅ CORRECT - Use native HTML with ShadCN styling classes
<div className="rounded-md border">
  <div className="relative w-full overflow-auto" style={{ height: '600px' }}>
    <table className="w-full caption-bottom text-sm">
      <thead className="sticky top-0 z-10 bg-white [&_tr]:border-b">
        <tr className="border-b transition-colors hover:bg-muted/50">
          <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">
            Header Content
          </th>
        </tr>
      </thead>
      <tbody className="[&_tr:last-child]:border-0">
        <tr className="border-b transition-colors hover:bg-muted/50">
          <td className="p-4 align-middle">Cell Content</td>
        </tr>
      </tbody>
    </table>
  </div>
</div>
```

### Key Points for Sticky Headers
1. **Use native HTML elements**: `<table>`, `<thead>`, `<tbody>`, `<tr>`, `<th>`, `<td>`
2. **Apply ShadCN classes manually**: Copy the classes from ShadCN components
3. **Control the scroll container**: You need direct control over the overflow element
4. **Required classes on thead**: `sticky top-0 z-10 bg-white`
5. **Background color is critical**: Without `bg-white`, content shows through when scrolling

### ShadCN Class Reference
When using native HTML instead of ShadCN components, use these classes:

| ShadCN Component | HTML Element | Classes to Apply |
|------------------|--------------|------------------|
| `<Table>` | `<table>` | `w-full caption-bottom text-sm` |
| `<TableHeader>` | `<thead>` | `[&_tr]:border-b` + `sticky top-0 z-10 bg-white` for sticky |
| `<TableBody>` | `<tbody>` | `[&_tr:last-child]:border-0` |
| `<TableRow>` | `<tr>` | `border-b transition-colors hover:bg-muted/50` |
| `<TableHead>` | `<th>` | `h-12 px-4 text-left align-middle font-medium text-muted-foreground` |
| `<TableCell>` | `<td>` | `p-4 align-middle` |

### When to Use Each Approach

**Use ShadCN Table component when:**
- You don't need sticky headers
- The table won't be scrollable
- You want the simplest implementation

**Use native HTML with ShadCN classes when:**
- You need sticky headers
- You need fine control over scrolling behavior
- You need custom scroll containers
- You're implementing virtual scrolling

### Additional Resources
- [Medium: ShadCN UI Sticky Table Header Implementation](https://medium.com/@shuhan.chan08/shadcn-ui-sticky-table-header-implementation-74b313d5c02e)
- [MDN: Position Sticky](https://developer.mozilla.org/en-US/docs/Web/CSS/position#sticky)
- [Quasar Tables with Sticky Headers](https://quasar.dev/vue-components/table/#sticky-header-column)

## Key Design Patterns

### 1. Data Structure
- All DBF data comes through `company.ReadDBFFile()` which returns:
  ```typescript
  {
    rows: any[],      // Array of record objects
    columns: string[] // Array of column names
  }
  ```
- Always add `_rowIndex` to track the original position for updates

### 2. Permissions
- Check user permissions: Admin and Root can edit, others are read-only
- Use the pattern:
  ```typescript
  const canEdit = currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')
  ```

### 3. Table Features
- **Fixed Header**: Use `sticky top-0` with `max-h-[600px] overflow-auto` on container
- **Sortable Columns**: Click header to sort, show arrow only on sorted column
- **Clickable Rows**: Every row opens detail modal
- **Search**: Global search across all fields
- **Loading States**: Show spinner in refresh button and loading message in table

### 4. Edit Modal
- Shows all fields in organized sections
- Read-only mode for non-admin users
- Only sends changed fields to backend
- Success message with auto-close
- Proper error handling

### 5. Styling Standards
- Use ShadCN UI components consistently
- Gray color scheme: `text-gray-500`, `bg-gray-50`, `border-gray-200`
- Status badges: Green for active, gray for inactive
- Icons from lucide-react with consistent sizing (`h-4 w-4`)

## Performance Considerations

1. **Pagination**: For large datasets (>1000 records), implement pagination
2. **Memoization**: Use `useMemo` for filtered/sorted data
3. **Debouncing**: Consider debouncing search input for large datasets
4. **Lazy Loading**: Load detailed data only when modal opens

## Error Handling

1. Always wrap async operations in try-catch
2. Show user-friendly error messages
3. Log detailed errors to console for debugging
4. Maintain loading states during operations

## Testing Checklist

- [ ] Table loads and displays data correctly
- [ ] Search filters records in real-time
- [ ] Sort works for all columns (ascending/descending)
- [ ] Click any row to open detail modal
- [ ] Edit and save works for admin users
- [ ] Read-only mode works for non-admin users
- [ ] Error messages display appropriately
- [ ] Loading states show during operations
- [ ] Table header stays fixed when scrolling
- [ ] Sort arrows only show on active column

## Example Implementations

- **Outstanding Checks**: `components/OutstandingChecks.tsx`
- **Vendor Management**: `components/VendorManagement.tsx`
- **DBF Explorer**: `components/DBFExplorer.tsx`

## Future Enhancements

1. **Drag & Drop Columns**: Implement column reordering
2. **Column Visibility**: Allow hiding/showing columns
3. **Export**: Add CSV/Excel export functionality
4. **Bulk Operations**: Select multiple rows for bulk updates
5. **Advanced Filters**: Add column-specific filters
6. **Keyboard Navigation**: Arrow keys to navigate, Enter to open modal

---

**Last Updated**: January 2025
**Author**: FinancialsX Development Team