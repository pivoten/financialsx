/**
 * Vendor Management Component
 * Properly structured with explicit field definitions and intentional UI design
 */

import React, { useState, useEffect, useMemo } from 'react'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import { Alert, AlertDescription } from './ui/alert'
import { Badge } from './ui/badge'
import { Checkbox } from './ui/checkbox'
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select'
import {
  Search,
  RefreshCw,
  Plus,
  Edit,
  Save,
  X,
  CheckCircle,
  AlertCircle,
  Building2,
  Phone,
  Mail,
  MapPin,
  CreditCard,
  Eye,
  EyeOff,
  Filter,
  Download,
  Upload,
  ChevronUp,
  ChevronDown
} from 'lucide-react'

import * as WailsApp from '../../wailsjs/go/main/App'
import { Vendor, VendorFormData, VendorFilters, VendorSortConfig, VendorTableColumn } from '../types/vendor'
import { dbfArrayToVendors, vendorToDbf, getDisplayTaxId, formatPhone, validateVendor } from '../utils/vendorTransform'
import logger from '../services/logger'

interface Props {
  companyName: string
  currentUser?: any
}

// Define table columns explicitly
const TABLE_COLUMNS: VendorTableColumn[] = [
  {
    key: 'vendorId',
    label: 'Vendor ID',
    sortable: true,
    width: '120px'
  },
  {
    key: 'vendorName',
    label: 'Vendor Name',
    sortable: true,
    width: '250px'
  },
  {
    key: 'contactName',
    label: 'Contact',
    sortable: true,
    width: '180px'
  },
  {
    key: 'phone',
    label: 'Phone',
    sortable: false,
    width: '140px',
    format: (value) => formatPhone(value)
  },
  {
    key: 'email',
    label: 'Email',
    sortable: true,
    width: '200px'
  },
  {
    key: 'city',
    label: 'City',
    sortable: true,
    width: '120px'
  },
  {
    key: 'state',
    label: 'State',
    sortable: true,
    width: '80px',
    align: 'center'
  },
  {
    key: 'isActive',
    label: 'Status',
    sortable: true,
    width: '100px',
    align: 'center',
    format: (value) => (
      <Badge
        variant={value ? 'default' : 'secondary'}
        className={value ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}
      >
        {value ? 'Active' : 'Inactive'}
      </Badge>
    )
  }
]

export default function VendorManagementProper({ companyName, currentUser }: Props) {
  // State management
  const [vendors, setVendors] = useState<Vendor[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  // Filters and search
  const [filters, setFilters] = useState<VendorFilters>({
    search: '',
    showInactive: false,
    vendorType: undefined,
    has1099: undefined
  })

  // Sorting
  const [sortConfig, setSortConfig] = useState<VendorSortConfig>({
    key: 'vendorName',
    direction: 'asc'
  })

  // Edit modal
  const [editModalOpen, setEditModalOpen] = useState(false)
  const [selectedVendor, setSelectedVendor] = useState<Vendor | null>(null)
  const [editedVendor, setEditedVendor] = useState<VendorFormData | null>(null)
  const [saving, setSaving] = useState(false)
  const [validationErrors, setValidationErrors] = useState<string[]>([])

  // Tax ID visibility
  const [showTaxId, setShowTaxId] = useState<Set<string>>(new Set())

  // Permissions
  const canEdit = true // TODO: Implement proper permissions
  const canCreate = true
  const canDelete = false

  // Load vendors from DBF
  const loadVendors = async () => {
    setLoading(true)
    setError(null)

    try {
      logger.info('Loading vendors', { company: companyName })
      const result = await WailsApp.GetVendors(companyName)
      
      if (!result) {
        throw new Error('No data returned from server')
      }

      const vendorList = dbfArrayToVendors(result)
      logger.info('Vendors loaded', { count: vendorList.length })
      setVendors(vendorList)
    } catch (err: any) {
      logger.error('Failed to load vendors', { error: err.message })
      setError(err.message || 'Failed to load vendors')
    } finally {
      setLoading(false)
    }
  }

  // Initial load
  useEffect(() => {
    if (companyName) {
      loadVendors()
    }
  }, [companyName])

  // Filter and sort vendors
  const processedVendors = useMemo(() => {
    let filtered = [...vendors]

    // Apply search filter
    if (filters.search) {
      const searchLower = filters.search.toLowerCase()
      filtered = filtered.filter(vendor =>
        vendor.vendorId.toLowerCase().includes(searchLower) ||
        vendor.vendorName.toLowerCase().includes(searchLower) ||
        vendor.contactName.toLowerCase().includes(searchLower) ||
        vendor.email.toLowerCase().includes(searchLower) ||
        vendor.phone.includes(filters.search)
      )
    }

    // Apply status filter
    if (!filters.showInactive) {
      filtered = filtered.filter(vendor => vendor.isActive)
    }

    // Apply vendor type filter
    if (filters.vendorType) {
      filtered = filtered.filter(vendor => vendor.vendorType === filters.vendorType)
    }

    // Apply 1099 filter
    if (filters.has1099 !== undefined) {
      filtered = filtered.filter(vendor => vendor.is1099 === filters.has1099)
    }

    // Apply sorting
    if (sortConfig.key) {
      filtered.sort((a, b) => {
        const aVal = a[sortConfig.key!] ?? ''
        const bVal = b[sortConfig.key!] ?? ''

        // Handle different types
        if (typeof aVal === 'boolean' && typeof bVal === 'boolean') {
          return sortConfig.direction === 'asc' 
            ? (aVal === bVal ? 0 : aVal ? 1 : -1)
            : (aVal === bVal ? 0 : aVal ? -1 : 1)
        }

        if (typeof aVal === 'number' && typeof bVal === 'number') {
          return sortConfig.direction === 'asc' ? aVal - bVal : bVal - aVal
        }

        // String comparison
        const comparison = String(aVal).localeCompare(String(bVal))
        return sortConfig.direction === 'asc' ? comparison : -comparison
      })
    }

    return filtered
  }, [vendors, filters, sortConfig])

  // Handle sort
  const handleSort = (key: keyof Vendor) => {
    setSortConfig(prev => ({
      key,
      direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc'
    }))
  }

  // Handle row click
  const handleRowClick = (vendor: Vendor) => {
    setSelectedVendor(vendor)
    setEditedVendor({
      vendorId: vendor.vendorId,
      vendorName: vendor.vendorName,
      contactName: vendor.contactName,
      phone: vendor.phone,
      phoneExt: vendor.phoneExt,
      fax: vendor.fax,
      email: vendor.email,
      address1: vendor.address1,
      address2: vendor.address2,
      city: vendor.city,
      state: vendor.state,
      zipCode: vendor.zipCode,
      country: vendor.country,
      taxId: vendor.taxId,
      accountNumber: vendor.accountNumber,
      terms: vendor.terms,
      discountPercent: vendor.discountPercent,
      discountDays: vendor.discountDays,
      netDays: vendor.netDays,
      creditLimit: vendor.creditLimit,
      isActive: vendor.isActive,
      is1099: vendor.is1099,
      notes: vendor.notes,
      vendorType: vendor.vendorType,
      website: vendor.website
    })
    setValidationErrors([])
    setEditModalOpen(true)
  }

  // Handle field change
  const handleFieldChange = (field: keyof VendorFormData, value: any) => {
    if (!editedVendor) return
    setEditedVendor({
      ...editedVendor,
      [field]: value
    })
    // Clear validation errors when user makes changes
    if (validationErrors.length > 0) {
      setValidationErrors([])
    }
  }

  // Handle save
  const handleSave = async () => {
    if (!editedVendor || !selectedVendor) return

    // Validate
    const errors = validateVendor(editedVendor)
    if (errors.length > 0) {
      setValidationErrors(errors)
      return
    }

    setSaving(true)
    setValidationErrors([])

    try {
      // Convert to DBF format
      const dbfData = vendorToDbf(editedVendor)
      
      // Call update API
      await WailsApp.UpdateVendor(companyName, selectedVendor._rowIndex!, dbfData)
      
      // Update local state
      const updatedVendor = { ...selectedVendor, ...editedVendor }
      setVendors(prev => prev.map(v => 
        v.vendorId === selectedVendor.vendorId ? updatedVendor : v
      ))
      
      setSuccessMessage('Vendor updated successfully')
      setTimeout(() => {
        setEditModalOpen(false)
        setSuccessMessage(null)
      }, 1500)
    } catch (err: any) {
      logger.error('Failed to save vendor', { error: err.message })
      setValidationErrors([err.message || 'Failed to save vendor'])
    } finally {
      setSaving(false)
    }
  }

  // Toggle tax ID visibility
  const toggleTaxIdVisibility = (vendorId: string) => {
    setShowTaxId(prev => {
      const next = new Set(prev)
      if (next.has(vendorId)) {
        next.delete(vendorId)
      } else {
        next.add(vendorId)
      }
      return next
    })
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Vendor Management</h2>
          <p className="text-muted-foreground">
            Manage vendor information and payment terms
          </p>
        </div>
        <div className="flex gap-2">
          {canCreate && (
            <Button onClick={() => console.log('Create new vendor')}>
              <Plus className="mr-2 h-4 w-4" />
              New Vendor
            </Button>
          )}
          <Button variant="outline" onClick={loadVendors} disabled={loading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex gap-4 items-end">
            <div className="flex-1 max-w-sm">
              <Label htmlFor="search">Search</Label>
              <div className="relative">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  id="search"
                  placeholder="Search by ID, name, contact, email..."
                  value={filters.search}
                  onChange={(e) => setFilters(prev => ({ ...prev, search: e.target.value }))}
                  className="pl-8"
                />
              </div>
            </div>
            
            <div className="flex items-center space-x-2">
              <Checkbox
                id="show-inactive"
                checked={filters.showInactive}
                onCheckedChange={(checked) => 
                  setFilters(prev => ({ ...prev, showInactive: checked as boolean }))
                }
              />
              <Label htmlFor="show-inactive" className="cursor-pointer">
                Show inactive
              </Label>
            </div>

            <Button
              variant="outline"
              size="sm"
              onClick={() => setFilters({ search: '', showInactive: false })}
            >
              Clear Filters
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Success/Error Messages */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {successMessage && (
        <Alert className="border-green-200 bg-green-50">
          <CheckCircle className="h-4 w-4 text-green-600" />
          <AlertDescription className="text-green-800">{successMessage}</AlertDescription>
        </Alert>
      )}

      {/* Data Table with Sticky Header */}
      <div className="border rounded-lg bg-white">
        <div className="relative overflow-auto max-h-[600px]">
          <Table>
            <TableHeader className="sticky top-0 z-10 bg-white border-b">
              <TableRow>
                {TABLE_COLUMNS.map(column => (
                  <TableHead
                    key={column.key}
                    className={`
                      ${column.sortable ? 'cursor-pointer hover:bg-gray-50' : ''}
                      ${column.align === 'center' ? 'text-center' : ''}
                      ${column.align === 'right' ? 'text-right' : ''}
                    `}
                    style={{ width: column.width }}
                    onClick={() => column.sortable && handleSort(column.key)}
                  >
                    <div className="flex items-center gap-1">
                      {column.label}
                      {column.sortable && sortConfig.key === column.key && (
                        sortConfig.direction === 'asc' 
                          ? <ChevronUp className="h-4 w-4" />
                          : <ChevronDown className="h-4 w-4" />
                      )}
                    </div>
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={TABLE_COLUMNS.length} className="text-center py-8">
                    <div className="flex items-center justify-center gap-2">
                      <RefreshCw className="h-5 w-5 animate-spin" />
                      Loading vendors...
                    </div>
                  </TableCell>
                </TableRow>
              ) : processedVendors.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={TABLE_COLUMNS.length} className="text-center py-8 text-gray-500">
                    {filters.search || filters.showInactive 
                      ? 'No vendors match your filters'
                      : 'No vendors found'}
                  </TableCell>
                </TableRow>
              ) : (
                processedVendors.map((vendor) => (
                  <TableRow
                    key={vendor.vendorId}
                    className="cursor-pointer hover:bg-gray-50"
                    onClick={() => handleRowClick(vendor)}
                  >
                    {TABLE_COLUMNS.map(column => (
                      <TableCell
                        key={column.key}
                        className={`
                          ${column.align === 'center' ? 'text-center' : ''}
                          ${column.align === 'right' ? 'text-right' : ''}
                        `}
                      >
                        {column.format 
                          ? column.format(vendor[column.key], vendor)
                          : vendor[column.key] || '-'}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      {/* Status Bar */}
      <div className="flex items-center justify-between text-sm text-gray-600">
        <div>
          Showing {processedVendors.length} of {vendors.length} vendors
        </div>
        <div className="flex gap-4">
          <span>Active: {vendors.filter(v => v.isActive).length}</span>
          <span>Inactive: {vendors.filter(v => !v.isActive).length}</span>
          <span>1099: {vendors.filter(v => v.is1099).length}</span>
        </div>
      </div>

      {/* Edit Modal - We'll add this in the next step */}
    </div>
  )
}