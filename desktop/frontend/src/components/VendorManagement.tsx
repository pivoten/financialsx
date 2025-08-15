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
  AlertCircle,
  Building,
  Phone,
  MapPin,
  Mail,
  CreditCard,
  Calendar,
  Hash,
  FileText,
  DollarSign,
  Briefcase
} from 'lucide-react'
import * as WailsApp from '../../wailsjs/go/main/App'

interface VendorRecord {
  _rowIndex: number
  CVENDNO?: string
  CCOMPANY?: string
  CADDRESS1?: string
  CADDRESS2?: string
  CCITY?: string
  CSTATE?: string
  CZIP?: string
  CPHONE?: string
  CFAX?: string
  CCONTACT?: string
  CEMAIL?: string
  CTERMS?: string
  CTAXID?: string
  CACCTNO?: string
  NDISC?: number
  NDISCDAYS?: number
  NNETDAYS?: number
  LINACTIVE?: boolean
  L1099?: boolean
  [key: string]: any
}

interface Props {
  companyName: string
  currentUser?: any
}

export default function VendorManagement({ companyName, currentUser }: Props) {
  // Data state
  const [vendors, setVendors] = useState<VendorRecord[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  
  // Search and filter state
  const [searchTerm, setSearchTerm] = useState('')
  const [sortColumn, setSortColumn] = useState<string>('')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc')
  
  // Edit modal state
  const [editModalOpen, setEditModalOpen] = useState(false)
  const [selectedVendor, setSelectedVendor] = useState<VendorRecord | null>(null)
  const [editedVendor, setEditedVendor] = useState<VendorRecord | null>(null)
  const [saving, setSaving] = useState(false)
  const [saveSuccess, setSaveSuccess] = useState(false)
  
  // Permissions - TODO: Implement proper auth/permissions
  // For now, allow editing for testing
  const canEdit = true // currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')

  // Load vendors from DBF
  const loadVendors = async () => {
    setLoading(true)
    setError(null)
    
    console.log('VendorManagement: Loading vendors for company:', companyName)
    
    try {
      const result = await WailsApp.GetVendors(companyName)
      console.log('VendorManagement: GetVendors result:', result)
      
      if (!result || !result.rows) {
        console.log('VendorManagement: No vendor data found in result')
        setError('No vendor data found')
        setVendors([])
        return
      }
      
      console.log('VendorManagement: Found', result.rows.length, 'vendor records')
      console.log('VendorManagement: First vendor record:', result.rows[0])
      console.log('VendorManagement: Columns available:', result.columns)
      
      // Convert array data to object format using column names
      const vendorsWithIndex = result.rows.map((row: any, index: number) => {
        const vendor: any = { _rowIndex: index }
        
        // If row is an array, convert to object using column names
        if (Array.isArray(row) && result.columns) {
          result.columns.forEach((col: string, colIndex: number) => {
            vendor[col] = row[colIndex]
          })
        } else {
          // If row is already an object, just spread it
          Object.assign(vendor, row)
        }
        
        return vendor
      })
      
      console.log('VendorManagement: First processed vendor:', vendorsWithIndex[0])
      
      console.log('VendorManagement: Processed vendors:', vendorsWithIndex.length)
      setVendors(vendorsWithIndex)
    } catch (err: any) {
      console.error('VendorManagement: Error loading vendors:', err)
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

  // Handle sort
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

  // Get sort icon for column
  const getSortIcon = (column: string) => {
    if (sortColumn !== column) {
      return null // No icon if not sorted
    }
    return sortDirection === 'asc' 
      ? <ArrowUp className="ml-2 h-4 w-4" />
      : <ArrowDown className="ml-2 h-4 w-4" />
  }

  // Handle row click
  const handleRowClick = (vendor: VendorRecord) => {
    setSelectedVendor(vendor)
    setEditedVendor({ ...vendor })
    setEditModalOpen(true)
    setSaveSuccess(false)
  }

  // Handle field change in edit modal
  const handleFieldChange = (field: string, value: any) => {
    if (!editedVendor) return
    setEditedVendor({
      ...editedVendor,
      [field]: value
    })
  }

  // Save vendor changes
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
    } catch (err: any) {
      setError(err.message || 'Failed to save vendor')
    } finally {
      setSaving(false)
    }
  }

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

      {/* Error Alert */}
      {error && (
        <Alert className="border-red-200 bg-red-50">
          <AlertCircle className="h-4 w-4 text-red-600" />
          <AlertDescription className="text-red-800">
            {error}
          </AlertDescription>
        </Alert>
      )}

      {/* Data Table with Fixed Header */}
      <div className="border rounded-lg">
        <div className="max-h-[600px] overflow-auto">
          <Table>
            <TableHeader className="sticky top-0 bg-white z-10 border-b">
              <TableRow>
                <TableHead 
                  className="cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('CVENDORID')}
                >
                  <div className="flex items-center">
                    Vendor ID
                    {getSortIcon('CVENDORID')}
                  </div>
                </TableHead>
                <TableHead 
                  className="cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('CVENDNAME')}
                >
                  <div className="flex items-center">
                    Vendor Name
                    {getSortIcon('CVENDNAME')}
                  </div>
                </TableHead>
                <TableHead 
                  className="cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('CCONTACT')}
                >
                  <div className="flex items-center">
                    Contact
                    {getSortIcon('CCONTACT')}
                  </div>
                </TableHead>
                <TableHead 
                  className="cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('CPHONE')}
                >
                  <div className="flex items-center">
                    Phone
                    {getSortIcon('CPHONE')}
                  </div>
                </TableHead>
                <TableHead 
                  className="cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('CEMAIL')}
                >
                  <div className="flex items-center">
                    Email
                    {getSortIcon('CEMAIL')}
                  </div>
                </TableHead>
                <TableHead 
                  className="cursor-pointer hover:bg-gray-50"
                  onClick={() => handleSort('LINACTIVE')}
                >
                  <div className="flex items-center">
                    Status
                    {getSortIcon('LINACTIVE')}
                  </div>
                </TableHead>
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
                processedVendors.map((vendor) => (
                  <TableRow
                    key={vendor._rowIndex}
                    className="cursor-pointer hover:bg-gray-50"
                    onClick={() => handleRowClick(vendor)}
                  >
                    <TableCell className="font-medium">
                      {vendor.CVENDORID || '-'}
                    </TableCell>
                    <TableCell>{vendor.CVENDNAME || '-'}</TableCell>
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

      {/* Edit Modal */}
      <Dialog open={editModalOpen} onOpenChange={setEditModalOpen}>
        <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {canEdit ? 'Edit Vendor' : 'View Vendor'}
            </DialogTitle>
            <DialogDescription>
              {canEdit 
                ? 'Modify vendor information and save changes to the database'
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

              {/* Basic Information */}
              <div className="space-y-4">
                <h3 className="text-sm font-semibold text-gray-700 uppercase tracking-wider">
                  Basic Information
                </h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="vendno">Vendor Number</Label>
                    <div className="flex items-center mt-1">
                      <Hash className="h-4 w-4 text-gray-400 mr-2" />
                      <Input
                        id="vendno"
                        value={editedVendor.CVENDNO || ''}
                        onChange={(e) => handleFieldChange('CVENDNO', e.target.value)}
                        disabled={!canEdit}
                      />
                    </div>
                  </div>
                  <div>
                    <Label htmlFor="company">Company Name</Label>
                    <div className="flex items-center mt-1">
                      <Building className="h-4 w-4 text-gray-400 mr-2" />
                      <Input
                        id="company"
                        value={editedVendor.CCOMPANY || ''}
                        onChange={(e) => handleFieldChange('CCOMPANY', e.target.value)}
                        disabled={!canEdit}
                      />
                    </div>
                  </div>
                </div>
              </div>

              {/* Contact Information */}
              <div className="space-y-4">
                <h3 className="text-sm font-semibold text-gray-700 uppercase tracking-wider">
                  Contact Information
                </h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="contact">Contact Person</Label>
                    <Input
                      id="contact"
                      value={editedVendor.CCONTACT || ''}
                      onChange={(e) => handleFieldChange('CCONTACT', e.target.value)}
                      disabled={!canEdit}
                    />
                  </div>
                  <div>
                    <Label htmlFor="email">Email</Label>
                    <div className="flex items-center mt-1">
                      <Mail className="h-4 w-4 text-gray-400 mr-2" />
                      <Input
                        id="email"
                        type="email"
                        value={editedVendor.CEMAIL || ''}
                        onChange={(e) => handleFieldChange('CEMAIL', e.target.value)}
                        disabled={!canEdit}
                      />
                    </div>
                  </div>
                  <div>
                    <Label htmlFor="phone">Phone</Label>
                    <div className="flex items-center mt-1">
                      <Phone className="h-4 w-4 text-gray-400 mr-2" />
                      <Input
                        id="phone"
                        value={editedVendor.CPHONE || ''}
                        onChange={(e) => handleFieldChange('CPHONE', e.target.value)}
                        disabled={!canEdit}
                      />
                    </div>
                  </div>
                  <div>
                    <Label htmlFor="fax">Fax</Label>
                    <Input
                      id="fax"
                      value={editedVendor.CFAX || ''}
                      onChange={(e) => handleFieldChange('CFAX', e.target.value)}
                      disabled={!canEdit}
                    />
                  </div>
                </div>
              </div>

              {/* Address */}
              <div className="space-y-4">
                <h3 className="text-sm font-semibold text-gray-700 uppercase tracking-wider">
                  Address
                </h3>
                <div className="space-y-4">
                  <div>
                    <Label htmlFor="address1">Address Line 1</Label>
                    <div className="flex items-center mt-1">
                      <MapPin className="h-4 w-4 text-gray-400 mr-2" />
                      <Input
                        id="address1"
                        value={editedVendor.CADDRESS1 || ''}
                        onChange={(e) => handleFieldChange('CADDRESS1', e.target.value)}
                        disabled={!canEdit}
                      />
                    </div>
                  </div>
                  <div>
                    <Label htmlFor="address2">Address Line 2</Label>
                    <Input
                      id="address2"
                      value={editedVendor.CADDRESS2 || ''}
                      onChange={(e) => handleFieldChange('CADDRESS2', e.target.value)}
                      disabled={!canEdit}
                    />
                  </div>
                  <div className="grid grid-cols-3 gap-4">
                    <div>
                      <Label htmlFor="city">City</Label>
                      <Input
                        id="city"
                        value={editedVendor.CCITY || ''}
                        onChange={(e) => handleFieldChange('CCITY', e.target.value)}
                        disabled={!canEdit}
                      />
                    </div>
                    <div>
                      <Label htmlFor="state">State</Label>
                      <Input
                        id="state"
                        value={editedVendor.CSTATE || ''}
                        onChange={(e) => handleFieldChange('CSTATE', e.target.value)}
                        disabled={!canEdit}
                        maxLength={2}
                      />
                    </div>
                    <div>
                      <Label htmlFor="zip">ZIP Code</Label>
                      <Input
                        id="zip"
                        value={editedVendor.CZIP || ''}
                        onChange={(e) => handleFieldChange('CZIP', e.target.value)}
                        disabled={!canEdit}
                      />
                    </div>
                  </div>
                </div>
              </div>

              {/* Payment Terms */}
              <div className="space-y-4">
                <h3 className="text-sm font-semibold text-gray-700 uppercase tracking-wider">
                  Payment Terms
                </h3>
                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <Label htmlFor="terms">Terms</Label>
                    <div className="flex items-center mt-1">
                      <Calendar className="h-4 w-4 text-gray-400 mr-2" />
                      <Input
                        id="terms"
                        value={editedVendor.CTERMS || ''}
                        onChange={(e) => handleFieldChange('CTERMS', e.target.value)}
                        disabled={!canEdit}
                      />
                    </div>
                  </div>
                  <div>
                    <Label htmlFor="discount">Discount %</Label>
                    <Input
                      id="discount"
                      type="number"
                      value={editedVendor.NDISC || ''}
                      onChange={(e) => handleFieldChange('NDISC', parseFloat(e.target.value) || 0)}
                      disabled={!canEdit}
                    />
                  </div>
                  <div>
                    <Label htmlFor="netdays">Net Days</Label>
                    <Input
                      id="netdays"
                      type="number"
                      value={editedVendor.NNETDAYS || ''}
                      onChange={(e) => handleFieldChange('NNETDAYS', parseInt(e.target.value) || 0)}
                      disabled={!canEdit}
                    />
                  </div>
                </div>
              </div>

              {/* Tax Information */}
              <div className="space-y-4">
                <h3 className="text-sm font-semibold text-gray-700 uppercase tracking-wider">
                  Tax Information
                </h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="taxid">Tax ID</Label>
                    <div className="flex items-center mt-1">
                      <FileText className="h-4 w-4 text-gray-400 mr-2" />
                      <Input
                        id="taxid"
                        value={editedVendor.CTAXID || ''}
                        onChange={(e) => handleFieldChange('CTAXID', e.target.value)}
                        disabled={!canEdit}
                      />
                    </div>
                  </div>
                  <div>
                    <Label htmlFor="acctno">GL Account Number</Label>
                    <div className="flex items-center mt-1">
                      <CreditCard className="h-4 w-4 text-gray-400 mr-2" />
                      <Input
                        id="acctno"
                        value={editedVendor.CACCTNO || ''}
                        onChange={(e) => handleFieldChange('CACCTNO', e.target.value)}
                        disabled={!canEdit}
                      />
                    </div>
                  </div>
                </div>
                <div className="flex items-center space-x-4">
                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      id="inactive"
                      checked={editedVendor.LINACTIVE === false}
                      onChange={(e) => handleFieldChange('LINACTIVE', !e.target.checked)}
                      disabled={!canEdit}
                      className="mr-2"
                    />
                    <Label htmlFor="inactive">Active</Label>
                  </div>
                  <div className="flex items-center">
                    <input
                      type="checkbox"
                      id="1099"
                      checked={editedVendor.L1099 === true}
                      onChange={(e) => handleFieldChange('L1099', e.target.checked)}
                      disabled={!canEdit}
                      className="mr-2"
                    />
                    <Label htmlFor="1099">1099 Vendor</Label>
                  </div>
                </div>
              </div>

              {/* All Fields (Debug View) */}
              {process.env.NODE_ENV === 'development' && (
                <details className="mt-6">
                  <summary className="cursor-pointer text-sm text-gray-500">Debug: All Fields</summary>
                  <pre className="mt-2 text-xs bg-gray-100 p-2 rounded overflow-auto">
                    {JSON.stringify(editedVendor, null, 2)}
                  </pre>
                </details>
              )}
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
    </div>
  )
}