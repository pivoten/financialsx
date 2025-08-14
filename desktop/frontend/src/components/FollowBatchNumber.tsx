import React, { useState } from 'react'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import { Search, FileText, X } from 'lucide-react'
import { FollowBatchNumber as FollowBatchNumberAPI, UpdateBatchFields } from '../../wailsjs/go/main/App'
import BatchFlowChart from './BatchFlowChart'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from './ui/dialog'
import { Label } from './ui/label'
import { Checkbox } from './ui/checkbox'
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from './ui/select'

interface BatchData {
  batch_number: string
  company_name: string
  total_records_found: number
  checks: TableData
  glmaster: TableData
  appurchd: TableData
  appmthdr: TableData
  appmtdet: TableData
}

interface TableData {
  table_name: string
  records: any[]
  count: number
  columns: string[]
  error?: string
}

// Helper function to determine field type
const getFieldType = (fieldName: string): 'date' | 'numeric' | 'text' => {
  if (fieldName.startsWith('D')) return 'date'
  if (fieldName.startsWith('N') || fieldName.includes('AMOUNT') || 
      fieldName === 'NDEBITS' || fieldName === 'NCREDITS') return 'numeric'
  return 'text'
}

// Predefined field mappings for common update scenarios
const FIELD_MAPPING_TEMPLATES = {
  'Transaction Date': {
    description: 'Update all date fields related to this transaction',
    fieldType: 'date',
    sourceField: 'DCHECKDATE', // Primary source field
    mappings: {
      'CHECKS.DBF': 'DCHECKDATE',
      'GLMASTER.DBF': 'DDATE',
      'APPURCHD.DBF': 'DINVDATE',
      'APPMTHDR.DBF': 'DINVDATE',
      'APPMTDET.DBF': 'DINVDATE'
    }
  },
  'Post Date': {
    description: 'Update posting date across tables',
    fieldType: 'date',
    sourceField: 'DPOSTDATE',
    mappings: {
      'CHECKS.DBF': 'DPOSTDATE',
      'GLMASTER.DBF': 'DPOSTDATE',
      'APPURCHD.DBF': 'DPOSTDATE',
      'APPMTHDR.DBF': 'DPOSTDATE',
      'APPMTDET.DBF': 'DPOSTDATE'
    }
  },
  'Account Number': {
    description: 'Update account number across all tables',
    fieldType: 'text',
    sourceField: 'CACCTNO',
    mappings: {
      'CHECKS.DBF': 'CACCTNO',
      'GLMASTER.DBF': 'CACCTNO',
      'APPURCHD.DBF': 'CACCTNO',
      'APPMTHDR.DBF': 'CACCTNO',
      'APPMTDET.DBF': 'CACCTNO'
    }
  },
  'Transaction Amount': {
    description: 'Update transaction amounts (Note: GL may split to debits/credits)',
    fieldType: 'numeric',
    sourceField: 'NAMOUNT',
    mappings: {
      'CHECKS.DBF': 'NAMOUNT',
      'GLMASTER.DBF': 'NDEBITS',  // Special case: might need NCREDITS
      'APPURCHD.DBF': 'NAMOUNT',
      'APPMTHDR.DBF': 'NAMOUNT',
      'APPMTDET.DBF': 'NAMOUNT'
    }
  },
  'Accounting Period': {
    description: 'Update accounting period',
    fieldType: 'text',
    sourceField: 'CPERIOD',
    mappings: {
      'CHECKS.DBF': 'CPERIOD',
      'GLMASTER.DBF': 'CPERIOD',
      'APPURCHD.DBF': 'CPERIOD',
      'APPMTHDR.DBF': 'CPERIOD',
      'APPMTDET.DBF': 'CPERIOD'
    }
  },
  'Accounting Year': {
    description: 'Update accounting year',
    fieldType: 'text',
    sourceField: 'CYEAR',
    mappings: {
      'CHECKS.DBF': 'CYEAR',
      'GLMASTER.DBF': 'CYEAR',
      'APPURCHD.DBF': 'CYEAR',
      'APPMTHDR.DBF': 'CYEAR',
      'APPMTDET.DBF': 'CYEAR'
    }
  },
  'Vendor/Payee ID': {
    description: 'Update vendor or payee identification',
    fieldType: 'text',
    sourceField: 'CID',
    mappings: {
      'CHECKS.DBF': 'CID',
      'GLMASTER.DBF': 'CID',
      'APPURCHD.DBF': 'CVENDORID',
      'APPMTHDR.DBF': 'CVENDORID',
      'APPMTDET.DBF': 'CVENDORID'
    }
  },
  'Custom Mapping': {
    description: 'Define your own field mappings',
    fieldType: 'custom',
    sourceField: '',
    mappings: {}
  }
}

const FollowBatchNumber: React.FC = () => {
  const [batchNumber, setBatchNumber] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [batchData, setBatchData] = useState<BatchData | null>(null)
  const [selectedRecord, setSelectedRecord] = useState<any>(null)
  const [showRecordDialog, setShowRecordDialog] = useState(false)
  const [searchHistory, setSearchHistory] = useState<string[]>([])
  const [showHistory, setShowHistory] = useState(false)
  const [showUpdateDialog, setShowUpdateDialog] = useState(false)
  const [selectedField, setSelectedField] = useState('')
  const [newValue, setNewValue] = useState('')
  const [tablesToUpdate, setTablesToUpdate] = useState<Record<string, boolean>>({
    'CHECKS.DBF': true,
    'GLMASTER.DBF': true,
    'APPURCHD.DBF': true,
    'APPMTHDR.DBF': true,
    'APPMTDET.DBF': true
  })
  const [availableFields, setAvailableFields] = useState<Record<string, string[]>>({})
  const [fieldMappings, setFieldMappings] = useState<Record<string, string>>({})
  const [updateMode, setUpdateMode] = useState<'simple' | 'mapped'>('mapped')
  const [selectedTemplate, setSelectedTemplate] = useState<string>('')

  // Load search history on component mount
  React.useEffect(() => {
    const history = localStorage.getItem('batchSearchHistory')
    if (history) {
      try {
        const parsed = JSON.parse(history)
        setSearchHistory(parsed.slice(0, 10)) // Keep only last 10 searches
      } catch (e) {
        console.error('Failed to parse search history:', e)
      }
    }
  }, [])

  // Click outside handler to close history dropdown
  React.useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as HTMLElement
      if (!target.closest('[data-search-container]')) {
        setShowHistory(false)
      }
    }
    
    if (showHistory) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [showHistory])

  // Save to search history
  const addToHistory = (searchTerm: string) => {
    const trimmed = searchTerm.trim()
    if (!trimmed) return
    
    // Remove duplicates and add to front
    const newHistory = [trimmed, ...searchHistory.filter(h => h !== trimmed)].slice(0, 10)
    setSearchHistory(newHistory)
    localStorage.setItem('batchSearchHistory', JSON.stringify(newHistory))
  }

  const handleSearch = async () => {
    if (!batchNumber.trim()) {
      setError('Please enter a batch number')
      return
    }

    setLoading(true)
    setError('')
    setBatchData(null)
    setShowHistory(false) // Hide history dropdown when searching

    try {
      const companyName = localStorage.getItem('company_name') || ''
      const companyPath = localStorage.getItem('company_path') || ''
      const companyToUse = companyPath || companyName

      if (!companyToUse) {
        throw new Error('No company selected')
      }

      const result = await FollowBatchNumberAPI(companyToUse, batchNumber.trim())
      setBatchData(result as BatchData)
      
      // Add to history after successful search
      addToHistory(batchNumber.trim())
      
      if (result.total_records_found === 0) {
        setError(`No records found for batch number: ${batchNumber}`)
      }
    } catch (err: any) {
      console.error('Error searching batch:', err)
      setError(err.message || 'Failed to search batch number')
    } finally {
      setLoading(false)
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch()
    }
  }

  const renderTable = (tableData: TableData, tableName: string) => {
    if (tableData.error) {
      return (
        <div className="text-red-500 text-sm p-4 bg-red-50 rounded">
          Error: {tableData.error}
        </div>
      )
    }

    if (tableData.count === 0) {
      return (
        <div className="text-gray-500 text-sm p-4 bg-gray-50 rounded">
          No records found in {tableName}
        </div>
      )
    }

    // Get key fields to display based on table
    const getKeyFields = (tableName: string) => {
      switch(tableName.toLowerCase()) {
        case 'checks.dbf':
          return ['CCHECKNO', 'DCHECKDATE', 'CPAYEE', 'NAMOUNT', 'CBATCH', 'CACCTNO']
        case 'glmaster.dbf':
          return ['CACCTNO', 'DDATE', 'CDESC', 'NDEBITS', 'NCREDITS', 'CBATCH', 'CSOURCE']
        case 'appurchd.dbf':
          return ['CINVOICE', 'CACCTNO', 'CDESCRIPT', 'NAMOUNT', 'CBATCH', 'CVENDORID']
        case 'appmthdr.dbf':
          return ['CINVOICE', 'CVENDORID', 'DINVDATE', 'NAMOUNT', 'CBATCH', 'CPAYTO']
        case 'appmtdet.dbf':
          return ['CINVOICE', 'CACCTNO', 'CDESCRIPT', 'NAMOUNT', 'CBILLTOKEN', 'CBATCH', 'CVENDORID']
        default:
          return tableData.columns.slice(0, 6) // First 6 columns
      }
    }

    const keyFields = getKeyFields(tableData.table_name)
    const displayColumns = keyFields.filter(field => tableData.columns.includes(field))

    return (
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">
                #
              </th>
              {displayColumns.map(col => {
                const isNumeric = col.startsWith('N') || col.includes('AMOUNT') || 
                                  col === 'NDEBITS' || col === 'NCREDITS'
                return (
                  <th key={col} className={`px-3 py-2 text-xs font-medium text-gray-500 uppercase ${isNumeric ? 'text-right' : 'text-left'}`}>
                    {col}
                  </th>
                )
              })}
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {tableData.records.map((record, idx) => (
              <tr 
                key={idx} 
                className="hover:bg-gray-50 cursor-pointer"
                onClick={() => {
                  setSelectedRecord({ ...record, _tableName: tableData.table_name })
                  setShowRecordDialog(true)
                }}
              >
                <td className="px-3 py-2 text-sm text-gray-500">
                  {idx + 1}
                </td>
                {displayColumns.map(col => {
                  const isNumeric = col.startsWith('N') || col.includes('AMOUNT') || 
                                    col === 'NDEBITS' || col === 'NCREDITS'
                  return (
                    <td key={col} className={`px-3 py-2 text-sm text-gray-900 ${isNumeric ? 'text-right' : ''}`}>
                      {formatValue(record[col], col)}
                    </td>
                  )
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    )
  }

  const formatValue = (value: any, fieldName?: string) => {
    if (value === null || value === undefined || value === '') return ''
    
    // Check if it's a date field based on field name or value format
    if (fieldName && (fieldName.toLowerCase().includes('date') || fieldName.startsWith('D'))) {
      // If it's a date string, format it
      if (typeof value === 'string' && value.includes('T')) {
        const date = new Date(value)
        if (!isNaN(date.getTime())) {
          return date.toLocaleDateString('en-US', { 
            year: 'numeric', 
            month: '2-digit', 
            day: '2-digit' 
          })
        }
      }
    }
    
    // Format numeric fields (amounts, debits, credits)
    if (fieldName && (fieldName.startsWith('N') || fieldName.includes('AMOUNT') || 
        fieldName === 'NDEBITS' || fieldName === 'NCREDITS')) {
      const num = typeof value === 'string' ? parseFloat(value) : value
      if (!isNaN(num)) {
        // Format as currency with proper alignment
        return new Intl.NumberFormat('en-US', {
          minimumFractionDigits: 2,
          maximumFractionDigits: 2
        }).format(num)
      }
    }
    
    if (typeof value === 'number') {
      return value.toFixed(2)
    }
    if (typeof value === 'boolean') return value ? 'Yes' : 'No'
    return String(value).trim()
  }

  return (
    <div className="p-6">
      {/* If no results yet, show the search interface */}
      {!batchData ? (
        <div className="space-y-6">
          {/* Header */}
          <div>
            <h2 className="text-2xl font-bold text-gray-900">Follow Batch Number</h2>
            <p className="text-sm text-gray-500 mt-1">
              Search for a batch number across CHECKS, GLMASTER, APPURCHD, APPMTHDR, and APPMTDET tables
            </p>
          </div>

          {/* Search Input */}
          <div className="flex justify-center">
            <Card className="shadow-sm max-w-2xl w-full">
              <CardContent className="flex flex-col items-center justify-center min-h-[120px] px-8">
                <div className="flex gap-4 w-full">
            <div className="flex-1 relative" data-search-container>
              <Input
                type="text"
                placeholder="Enter batch number (e.g., 001234)"
                value={batchNumber}
                onChange={(e) => setBatchNumber(e.target.value)}
                onKeyPress={handleKeyPress}
                onFocus={() => searchHistory.length > 0 && setShowHistory(true)}
                className="w-full"
                disabled={loading}
              />
              {/* Search History Dropdown */}
              {showHistory && searchHistory.length > 0 && (
                <div className="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-md shadow-lg max-h-60 overflow-auto">
                  <div className="py-1">
                    <div className="px-3 py-2 text-xs font-semibold text-gray-500 border-b">
                      Recent Searches
                    </div>
                    {searchHistory.map((item, idx) => (
                      <button
                        key={idx}
                        className="w-full px-3 py-2 text-left hover:bg-gray-100 focus:bg-gray-100 focus:outline-none text-sm"
                        onClick={() => {
                          setBatchNumber(item)
                          setShowHistory(false)
                          // Auto-search when selecting from history
                          setTimeout(() => {
                            const searchButton = document.querySelector('[data-search-button]') as HTMLButtonElement
                            searchButton?.click()
                          }, 0)
                        }}
                      >
                        {item}
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>
            <Button
              onClick={handleSearch}
              disabled={loading || !batchNumber.trim()}
              className="bg-blue-600 hover:bg-blue-700"
              data-search-button
            >
              <Search className="h-4 w-4 mr-2" />
              {loading ? 'Searching...' : 'Search'}
            </Button>
            {batchData && (
              <Button
                onClick={() => {
                  setBatchData(null)
                  setBatchNumber('')
                  setError('')
                  setShowHistory(false)
                }}
                variant="outline"
              >
                <X className="h-4 w-4 mr-2" />
                Clear
              </Button>
            )}
                </div>
                {error && (
                  <div className="mt-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded w-full">
                    {error}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      ) : (
        /* When we have results, show full layout */
        <div className="space-y-6">
          {/* Header */}
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-bold text-gray-900">Follow Batch Number</h2>
              <p className="text-sm text-gray-500 mt-1">
                Search for a batch number across CHECKS, GLMASTER, APPURCHD, and APPMTHDR tables
              </p>
            </div>
          </div>

          {/* Search Input */}
          <div className="flex justify-center">
            <Card className="shadow-sm max-w-2xl w-full">
              <CardContent className="flex flex-col items-center justify-center min-h-[100px] px-6">
                <div className="flex gap-4 w-full">
                <div className="flex-1 relative" data-search-container>
                  <Input
                    type="text"
                    placeholder="Enter batch number (e.g., 001234)"
                    value={batchNumber}
                    onChange={(e) => setBatchNumber(e.target.value)}
                    onKeyPress={handleKeyPress}
                    onFocus={() => searchHistory.length > 0 && setShowHistory(true)}
                    className="w-full"
                    disabled={loading}
                  />
                  {/* Search History Dropdown */}
                  {showHistory && searchHistory.length > 0 && (
                    <div className="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-md shadow-lg max-h-60 overflow-auto">
                      <div className="py-1">
                        <div className="px-3 py-2 text-xs font-semibold text-gray-500 border-b">
                          Recent Searches
                        </div>
                        {searchHistory.map((item, idx) => (
                          <button
                            key={idx}
                            className="w-full px-3 py-2 text-left hover:bg-gray-100 focus:bg-gray-100 focus:outline-none text-sm"
                            onClick={() => {
                              setBatchNumber(item)
                              setShowHistory(false)
                              // Auto-search when selecting from history
                              setTimeout(() => {
                                const searchButton = document.querySelector('[data-search-button]') as HTMLButtonElement
                                searchButton?.click()
                              }, 0)
                            }}
                          >
                            {item}
                          </button>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
                <Button
                  onClick={handleSearch}
                  disabled={loading || !batchNumber.trim()}
                  className="bg-blue-600 hover:bg-blue-700"
                  data-search-button
                >
                  <Search className="h-4 w-4 mr-2" />
                  {loading ? 'Searching...' : 'Search'}
                </Button>
                {batchData && (
                  <Button
                    onClick={() => {
                      setBatchData(null)
                      setBatchNumber('')
                      setError('')
                      setShowHistory(false)
                    }}
                    variant="outline"
                  >
                    <X className="h-4 w-4 mr-2" />
                    Clear
                  </Button>
                )}
                </div>
                {error && (
                  <div className="mt-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded w-full">
                    {error}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Results */}
          {batchData && batchData.total_records_found > 0 && (
            <div className="space-y-4">
          {/* Summary */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center">
                <FileText className="h-5 w-5 text-blue-600 mr-2" />
                <span className="font-medium text-blue-900">
                  Batch: {batchData.batch_number}
                </span>
              </div>
              <div className="flex items-center gap-4">
                <span className="text-sm text-blue-700">
                  Total Records Found: {batchData.total_records_found}
                </span>
                <Button
                  onClick={() => {
                    // Collect all unique fields from all tables that have records
                    const fields: Record<string, string[]> = {}
                    
                    // Helper function to add fields from a table
                    const addTableFields = (tableName: string, tableData: TableData) => {
                      if (tableData.count > 0 && tableData.columns) {
                        fields[tableName] = tableData.columns
                      }
                    }
                    
                    // Add fields from each table
                    if (batchData.checks) addTableFields('CHECKS.DBF', batchData.checks)
                    if (batchData.glmaster) addTableFields('GLMASTER.DBF', batchData.glmaster)
                    if (batchData.appurchd) addTableFields('APPURCHD.DBF', batchData.appurchd)
                    if (batchData.appmthdr) addTableFields('APPMTHDR.DBF', batchData.appmthdr)
                    if (batchData.appmtdet) addTableFields('APPMTDET.DBF', batchData.appmtdet)
                    
                    setAvailableFields(fields)
                    setShowUpdateDialog(true)
                  }}
                  variant="outline"
                  size="sm"
                  className="bg-white"
                >
                  Update Batch Details
                </Button>
              </div>
            </div>
          </div>

          {/* Flow Chart Visualization */}
          <BatchFlowChart 
            batchNumber={batchData.batch_number}
            searchResults={{
              checks: batchData.checks,
              glmaster: batchData.glmaster,
              appmthdr: batchData.appmthdr,
              appmtdet: batchData.appmtdet,
              appurchh: batchData.appurchd, // Using appurchd data for appurchh in flow
              appurchd: batchData.appurchd
            }}
          />

          {/* Tables Grid - Clickable Cards */}
          <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
            {/* CHECKS.DBF */}
            <Card className="h-fit">
              <CardHeader className="pb-3">
                <CardTitle className="text-base flex items-center justify-between">
                  <span>CHECKS.DBF</span>
                  <span className="text-sm font-normal text-gray-500">
                    {batchData.checks.count} record{batchData.checks.count !== 1 ? 's' : ''}
                  </span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                {renderTable(batchData.checks, 'CHECKS.DBF')}
              </CardContent>
            </Card>

            {/* GLMASTER.DBF */}
            <Card className="h-fit">
              <CardHeader className="pb-3">
                <CardTitle className="text-base flex items-center justify-between">
                  <span>GLMASTER.DBF</span>
                  <span className="text-sm font-normal text-gray-500">
                    {batchData.glmaster.count} record{batchData.glmaster.count !== 1 ? 's' : ''}
                  </span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                {renderTable(batchData.glmaster, 'GLMASTER.DBF')}
              </CardContent>
            </Card>

            {/* APPURCHD.DBF */}
            <Card className="h-fit">
              <CardHeader className="pb-3">
                <CardTitle className="text-base flex items-center justify-between">
                  <span>APPURCHD.DBF</span>
                  <span className="text-sm font-normal text-gray-500">
                    {batchData.appurchd.count} record{batchData.appurchd.count !== 1 ? 's' : ''}
                  </span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                {renderTable(batchData.appurchd, 'APPURCHD.DBF')}
              </CardContent>
            </Card>

            {/* APPMTHDR.DBF */}
            <Card className="h-fit">
              <CardHeader className="pb-3">
                <CardTitle className="text-base flex items-center justify-between">
                  <span>APPMTHDR.DBF</span>
                  <span className="text-sm font-normal text-gray-500">
                    {batchData.appmthdr.count} record{batchData.appmthdr.count !== 1 ? 's' : ''}
                  </span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                {renderTable(batchData.appmthdr, 'APPMTHDR.DBF')}
              </CardContent>
            </Card>

            {/* APPMTDET.DBF */}
            <Card className="h-fit">
              <CardHeader className="pb-3">
                <CardTitle className="text-base flex items-center justify-between">
                  <span>APPMTDET.DBF</span>
                  <span className="text-sm font-normal text-gray-500">
                    {batchData.appmtdet.count} record{batchData.appmtdet.count !== 1 ? 's' : ''}
                  </span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                {renderTable(batchData.appmtdet, 'APPMTDET.DBF')}
              </CardContent>
            </Card>
          </div>
            </div>
          )}
        </div>
      )}
      
      {/* Record Detail Dialog */}
      <Dialog open={showRecordDialog} onOpenChange={setShowRecordDialog}>
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Record Details</DialogTitle>
            <DialogDescription>
              {selectedRecord?._tableName || 'Complete record from database'}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2 mt-4">
            {selectedRecord && Object.entries(selectedRecord)
              .filter(([key]) => key !== '_tableName')
              .map(([key, value]) => (
                <div key={key} className="grid grid-cols-3 gap-2 py-1 border-b">
                  <span className="font-medium text-sm">{key}:</span>
                  <span className="col-span-2 text-sm">
                    {formatValue(value, key)}
                  </span>
                </div>
              ))}
          </div>
        </DialogContent>
      </Dialog>

      {/* Update Batch Details Dialog */}
      <Dialog open={showUpdateDialog} onOpenChange={setShowUpdateDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Update Batch Details</DialogTitle>
            <DialogDescription>
              Update field values across all tables for batch: {batchData?.batch_number}
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-6 py-4">
            {/* Update Template Selection */}
            <div className="space-y-2">
              <Label>Select Update Template</Label>
              <Select 
                value={selectedTemplate} 
                onValueChange={(template) => {
                  setSelectedTemplate(template)
                  const templateConfig = FIELD_MAPPING_TEMPLATES[template]
                  if (templateConfig) {
                    // Set the field mappings from the template
                    setFieldMappings(templateConfig.mappings)
                    
                    // Auto-select tables that have records and valid mappings
                    const newTableSelections: Record<string, boolean> = {}
                    Object.keys(tablesToUpdate).forEach(table => {
                      const tableKey = table.toLowerCase().replace('.dbf', '')
                      const hasRecords = batchData && batchData[tableKey]?.count > 0
                      const hasMapping = templateConfig.mappings[table]
                      const mappedField = templateConfig.mappings[table]
                      const fieldExists = mappedField && availableFields[table]?.includes(mappedField)
                      
                      newTableSelections[table] = hasRecords && hasMapping && fieldExists
                    })
                    setTablesToUpdate(newTableSelections)
                  }
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Choose an update template" />
                </SelectTrigger>
                <SelectContent>
                  {Object.entries(FIELD_MAPPING_TEMPLATES).map(([name, config]) => (
                    <SelectItem key={name} value={name}>
                      <div className="flex items-center gap-2">
                        <span>
                          {config.fieldType === 'date' && '📅'}
                          {config.fieldType === 'numeric' && '🔢'}
                          {config.fieldType === 'text' && '📝'}
                          {config.fieldType === 'custom' && '⚙️'}
                        </span>
                        <span>{name}</span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {selectedTemplate && FIELD_MAPPING_TEMPLATES[selectedTemplate] && (
                <p className="text-xs text-gray-500">
                  {FIELD_MAPPING_TEMPLATES[selectedTemplate].description}
                </p>
              )}
            </div>

            {/* New Value Input */}
            <div className="space-y-2">
              <Label>New Value</Label>
              {(() => {
                const template = selectedTemplate && FIELD_MAPPING_TEMPLATES[selectedTemplate]
                const fieldType = template?.fieldType || 'text'
                
                return (
                  <>
                    <Input
                      type={fieldType === 'date' ? 'date' : fieldType === 'numeric' ? 'number' : 'text'}
                      value={newValue}
                      onChange={(e) => setNewValue(e.target.value)}
                      placeholder={`Enter new ${fieldType === 'date' ? 'date' : fieldType === 'numeric' ? 'amount' : 'value'}`}
                      disabled={!selectedTemplate || selectedTemplate === 'Custom Mapping'}
                    />
                    {fieldType === 'date' && (
                      <p className="text-xs text-gray-500">Date format: YYYY-MM-DD</p>
                    )}
                    {fieldType === 'numeric' && (
                      <p className="text-xs text-gray-500">Enter numeric value (decimals allowed)</p>
                    )}
                  </>
                )
              })()}
            </div>

            {/* Table Selection with Field Mappings */}
            <div className="space-y-2">
              <Label>Tables and Field Mappings</Label>
              <div className="space-y-3 border rounded-lg p-4">
                {Object.keys(tablesToUpdate).map((table) => {
                  const tableKey = table.toLowerCase().replace('.dbf', '')
                  const hasRecords = batchData && batchData[tableKey]?.count > 0
                  const mappedField = fieldMappings[table]
                  const fieldExists = mappedField && availableFields[table]?.includes(mappedField)
                  const template = selectedTemplate && FIELD_MAPPING_TEMPLATES[selectedTemplate]
                  
                  // Type validation
                  const sourceFieldType = template?.fieldType
                  const targetFieldType = mappedField ? getFieldType(mappedField) : null
                  const typeMatch = !mappedField || !sourceFieldType || sourceFieldType === 'custom' || 
                                    sourceFieldType === targetFieldType
                  
                  return (
                    <div key={table} className="flex items-start space-x-2">
                      <Checkbox
                        id={table}
                        checked={tablesToUpdate[table]}
                        onCheckedChange={(checked) => 
                          setTablesToUpdate(prev => ({ ...prev, [table]: checked as boolean }))
                        }
                        disabled={!hasRecords || !fieldExists || !typeMatch}
                        className="mt-0.5"
                      />
                      <div className="flex-1">
                        <div className="flex items-center justify-between">
                          <Label htmlFor={table} className="cursor-pointer">
                            {table}
                            {hasRecords && (
                              <span className="text-sm text-gray-500 ml-2">
                                ({batchData[tableKey].count} record{batchData[tableKey].count !== 1 ? 's' : ''})
                              </span>
                            )}
                            {!hasRecords && (
                              <span className="text-sm text-gray-400 ml-2">(No records)</span>
                            )}
                          </Label>
                          {mappedField && (
                            <span className="text-xs font-mono bg-gray-100 px-2 py-1 rounded">
                              → {mappedField}
                            </span>
                          )}
                        </div>
                        {selectedTemplate && (
                          <div className="text-xs mt-1 space-y-1">
                            {!mappedField && (
                              <span className="text-gray-400">⚠️ No field mapping defined</span>
                            )}
                            {mappedField && !fieldExists && (
                              <span className="text-red-500">✗ Field "{mappedField}" not found in table</span>
                            )}
                            {mappedField && fieldExists && !typeMatch && (
                              <span className="text-amber-500">
                                ⚠️ Type mismatch: {sourceFieldType} → {targetFieldType}
                              </span>
                            )}
                            {mappedField && fieldExists && typeMatch && (
                              <span className="text-green-600">
                                ✓ Will update "{mappedField}" ({targetFieldType} field)
                              </span>
                            )}
                          </div>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
              {selectedTemplate && (
                <p className="text-xs text-gray-500">
                  Only tables with valid field mappings and matching field types can be updated.
                </p>
              )}
            </div>

            {/* Update Summary */}
            {selectedTemplate && newValue && (
              <div className="bg-blue-50 border border-blue-200 p-3 rounded">
                <p className="text-sm text-blue-800">
                  <strong>Update Summary:</strong>
                </p>
                <p className="text-sm text-blue-700 mt-1">
                  Template: <strong>{selectedTemplate}</strong>
                </p>
                <p className="text-sm text-blue-700">
                  New Value: <strong>{newValue}</strong>
                </p>
                <div className="mt-2">
                  <p className="text-sm text-blue-800 font-medium">Fields to be updated:</p>
                  <ul className="text-sm text-blue-700 mt-1 space-y-1">
                    {Object.entries(tablesToUpdate)
                      .filter(([table, isSelected]) => {
                        if (!isSelected) return false
                        const mappedField = fieldMappings[table]
                        const fieldExists = mappedField && availableFields[table]?.includes(mappedField)
                        const tableKey = table.toLowerCase().replace('.dbf', '')
                        const hasRecords = batchData && batchData[tableKey]?.count > 0
                        const template = FIELD_MAPPING_TEMPLATES[selectedTemplate]
                        const sourceFieldType = template?.fieldType
                        const targetFieldType = mappedField ? getFieldType(mappedField) : null
                        const typeMatch = sourceFieldType === 'custom' || sourceFieldType === targetFieldType
                        
                        return isSelected && fieldExists && hasRecords && typeMatch
                      })
                      .map(([table]) => {
                        const tableKey = table.toLowerCase().replace('.dbf', '')
                        const recordCount = batchData[tableKey]?.count || 0
                        const mappedField = fieldMappings[table]
                        return (
                          <li key={table}>
                            <strong>{table}</strong>: Field "{mappedField}" in {recordCount} record{recordCount !== 1 ? 's' : ''}
                          </li>
                        )
                      })}
                  </ul>
                  {Object.entries(tablesToUpdate).filter(([table, isSelected]) => {
                    if (!isSelected) return false
                    const mappedField = fieldMappings[table]
                    const fieldExists = mappedField && availableFields[table]?.includes(mappedField)
                    const tableKey = table.toLowerCase().replace('.dbf', '')
                    const hasRecords = batchData && batchData[tableKey]?.count > 0
                    return isSelected && fieldExists && hasRecords
                  }).length === 0 && (
                    <p className="text-sm text-yellow-600 mt-2">
                      No valid tables selected for update. Check field mappings and types.
                    </p>
                  )}
                </div>
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowUpdateDialog(false)}>
              Cancel
            </Button>
            <Button 
              onClick={async () => {
                if (!batchData || !selectedTemplate || !newValue) return
                
                try {
                  const companyName = localStorage.getItem('company_name') || ''
                  const companyPath = localStorage.getItem('company_path') || ''
                  const companyToUse = companyPath || companyName
                  
                  // Call the backend to update the batch fields with mappings
                  const result = await UpdateBatchFields(
                    companyToUse,
                    batchData.batch_number,
                    fieldMappings,  // Pass the field mappings
                    newValue,
                    tablesToUpdate
                  )
                  
                  if (result.success) {
                    alert(`Successfully updated ${result.total_updated} records across ${Object.keys(tablesToUpdate).filter(t => tablesToUpdate[t]).length} tables`)
                    
                    // Re-run the search to refresh the data
                    setShowUpdateDialog(false)
                    setSelectedTemplate('')
                    setFieldMappings({})
                    setNewValue('')
                    handleSearch() // Refresh the search results
                  } else {
                    const errors = result.errors || []
                    if (errors.length > 0) {
                      alert(`Update completed with errors:\n${errors.join('\n')}`)
                    } else {
                      alert('Update failed. Please check the logs for details.')
                    }
                  }
                } catch (error: any) {
                  console.error('Error updating batch fields:', error)
                  alert(`Failed to update batch fields: ${error.message || 'Unknown error'}`)
                }
              }}
              disabled={!selectedTemplate || !newValue || Object.keys(fieldMappings).length === 0 || !Object.values(tablesToUpdate).some(v => v)}
              className="bg-blue-600 hover:bg-blue-700"
            >
              Update Batch
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

export default FollowBatchNumber