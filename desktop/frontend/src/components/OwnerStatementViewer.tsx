import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Alert, AlertDescription } from './ui/alert'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { 
  ArrowLeft, 
  FileText, 
  User, 
  DollarSign, 
  AlertCircle, 
  RefreshCcw,
  TrendingUp,
  TrendingDown,
  Calculator
} from 'lucide-react'
import { GetOwnersList, GetOwnerStatementData } from '../../wailsjs/go/main/App'

interface OwnerStatementViewerProps {
  companyName: string
  fileName: string
  onBack: () => void
}

const OwnerStatementViewer: React.FC<OwnerStatementViewerProps> = ({ companyName, fileName, onBack }) => {
  const [loading, setLoading] = useState(true)
  const [owners, setOwners] = useState<any[]>([])
  const [selectedOwner, setSelectedOwner] = useState<string>('')
  const [statementData, setStatementData] = useState<any>(null)
  const [loadingStatement, setLoadingStatement] = useState(false)
  const [error, setError] = useState<string>('')

  useEffect(() => {
    loadOwners()
  }, [companyName, fileName])

  const loadOwners = async () => {
    setLoading(true)
    setError('')
    
    try {
      const ownersList = await GetOwnersList(companyName, fileName)
      setOwners(ownersList || [])
      
      // Auto-select first owner if there's only one
      if (ownersList && ownersList.length === 1) {
        setSelectedOwner(ownersList[0].key)
        loadStatementData(ownersList[0].key)
      }
    } catch (err) {
      console.error('Error loading owners:', err)
      setError('Error loading owners list')
    } finally {
      setLoading(false)
    }
  }

  const loadStatementData = async (ownerKey: string) => {
    setLoadingStatement(true)
    setError('')
    
    try {
      const data = await GetOwnerStatementData(companyName, fileName, ownerKey)
      setStatementData(data)
    } catch (err) {
      console.error('Error loading statement data:', err)
      setError('Error loading statement data')
    } finally {
      setLoadingStatement(false)
    }
  }

  const handleOwnerChange = (value: string) => {
    setSelectedOwner(value)
    setStatementData(null)
    if (value) {
      loadStatementData(value)
    }
  }

  const formatCurrency = (value: any) => {
    if (value === null || value === undefined || value === '') return '$0.00'
    const num = parseFloat(value)
    if (isNaN(num)) return '$0.00'
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(num)
  }

  const formatDate = (value: any) => {
    if (!value) return ''
    try {
      const date = new Date(value)
      if (isNaN(date.getTime())) return value
      return date.toLocaleDateString('en-US')
    } catch {
      return value
    }
  }

  const renderFieldValue = (key: string, value: any) => {
    const keyUpper = key.toUpperCase()
    
    // Format currency fields
    if (keyUpper.includes('AMOUNT') || keyUpper.includes('GROSS') || 
        keyUpper.includes('NET') || keyUpper.includes('TAX') || 
        keyUpper.includes('DEDUCT') || keyUpper.includes('REVENUE')) {
      return formatCurrency(value)
    }
    
    // Format date fields
    if (keyUpper.includes('DATE') || keyUpper.includes('DEXP')) {
      return formatDate(value)
    }
    
    // Format boolean fields
    if (typeof value === 'boolean') {
      return value ? 'Yes' : 'No'
    }
    
    // Default formatting
    if (value === null || value === undefined || value === '') return '-'
    return String(value).trim()
  }

  const getImportantFields = (row: any) => {
    // Extract the most important fields to show in the summary
    const important: { [key: string]: any } = {}
    
    for (const [key, value] of Object.entries(row)) {
      const keyUpper = key.toUpperCase()
      
      // Skip empty values
      if (value === null || value === undefined || value === '' || value === 0) continue
      
      // Include important fields
      if (keyUpper.includes('WELL') || keyUpper.includes('LEASE') ||
          keyUpper.includes('GROSS') || keyUpper.includes('NET') ||
          keyUpper.includes('TAX') || keyUpper.includes('DEDUCT') ||
          keyUpper.includes('DATE') || keyUpper.includes('CHECK') ||
          keyUpper.includes('DESC') || keyUpper.includes('PROD')) {
        important[key] = value
      }
    }
    
    return important
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={onBack}
          className="hover:bg-gray-100"
        >
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back
        </Button>
        <div className="flex-1">
          <h2 className="text-2xl font-bold text-gray-900">Owner Statement Viewer</h2>
          <p className="text-sm text-gray-500 mt-1">
            View detailed statement information for individual owners
          </p>
        </div>
      </div>

      {/* Owner Selection */}
      <Card>
        <CardHeader>
          <CardTitle>Select Owner</CardTitle>
          <CardDescription>
            Choose an owner to view their distribution statement details
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-4">
              <RefreshCcw className="h-6 w-6 animate-spin text-gray-400" />
              <span className="ml-2 text-gray-600">Loading owners...</span>
            </div>
          ) : owners.length === 0 ? (
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                No owners found in the statement file
              </AlertDescription>
            </Alert>
          ) : (
            <div className="flex items-center gap-4">
              <User className="h-5 w-5 text-gray-400" />
              <Select value={selectedOwner} onValueChange={handleOwnerChange}>
                <SelectTrigger className="flex-1">
                  <SelectValue placeholder="Select an owner..." />
                </SelectTrigger>
                <SelectContent>
                  {owners.map((owner) => (
                    <SelectItem key={owner.key} value={owner.key}>
                      {owner.name || owner.id || owner.key}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Badge variant="secondary">
                {owners.length} {owners.length === 1 ? 'Owner' : 'Owners'}
              </Badge>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Statement Data */}
      {selectedOwner && (
        <>
          {loadingStatement ? (
            <Card>
              <CardContent className="pt-6">
                <div className="flex items-center justify-center">
                  <RefreshCcw className="h-6 w-6 animate-spin text-gray-400" />
                  <span className="ml-2 text-gray-600">Loading statement data...</span>
                </div>
              </CardContent>
            </Card>
          ) : statementData ? (
            <>
              {/* Summary Card */}
              <Card>
                <CardHeader>
                  <CardTitle>Statement Summary</CardTitle>
                  <CardDescription>
                    Overview for {selectedOwner}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                    <div className="bg-gray-50 p-4 rounded-lg">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-500">Total Records</span>
                        <FileText className="h-4 w-4 text-gray-400" />
                      </div>
                      <p className="text-2xl font-bold text-gray-900 mt-1">
                        {statementData.rowCount || 0}
                      </p>
                    </div>
                    
                    <div className="bg-green-50 p-4 rounded-lg">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-500">Gross Revenue</span>
                        <TrendingUp className="h-4 w-4 text-green-600" />
                      </div>
                      <p className="text-2xl font-bold text-green-600 mt-1">
                        {formatCurrency(statementData.totals?.gross || 0)}
                      </p>
                    </div>
                    
                    <div className="bg-blue-50 p-4 rounded-lg">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-500">Net Amount</span>
                        <DollarSign className="h-4 w-4 text-blue-600" />
                      </div>
                      <p className="text-2xl font-bold text-blue-600 mt-1">
                        {formatCurrency(statementData.totals?.net || 0)}
                      </p>
                    </div>
                    
                    <div className="bg-amber-50 p-4 rounded-lg">
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-500">Deductions</span>
                        <TrendingDown className="h-4 w-4 text-amber-600" />
                      </div>
                      <p className="text-2xl font-bold text-amber-600 mt-1">
                        {formatCurrency(statementData.totals?.tax || 0)}
                      </p>
                    </div>
                  </div>
                  
                  {statementData.wellCount > 0 && (
                    <div className="mt-4 pt-4 border-t">
                      <div className="flex items-center gap-2">
                        <Calculator className="h-4 w-4 text-gray-400" />
                        <span className="text-sm text-gray-600">
                          Data from {statementData.wellCount} {statementData.wellCount === 1 ? 'well' : 'wells'}
                        </span>
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>

              {/* Detailed Records */}
              {statementData.rows && statementData.rows.length > 0 && (
                <Card>
                  <CardHeader>
                    <CardTitle>Statement Details</CardTitle>
                    <CardDescription>
                      Individual line items for this owner
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      {statementData.rows.map((row: any, index: number) => {
                        const importantFields = getImportantFields(row)
                        const fieldCount = Object.keys(importantFields).length
                        
                        if (fieldCount === 0) return null
                        
                        return (
                          <div
                            key={index}
                            className="border rounded-lg p-4 hover:bg-gray-50 transition-colors"
                          >
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                              {Object.entries(importantFields).map(([key, value]) => (
                                <div key={key} className="flex flex-col">
                                  <span className="text-xs text-gray-500 uppercase tracking-wider">
                                    {key.replace(/_/g, ' ')}
                                  </span>
                                  <span className="text-sm font-medium text-gray-900">
                                    {renderFieldValue(key, value)}
                                  </span>
                                </div>
                              ))}
                            </div>
                          </div>
                        )
                      })}
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Raw Data (Collapsible) */}
              <details className="group">
                <summary className="cursor-pointer list-none">
                  <Card className="hover:shadow-md transition-shadow">
                    <CardHeader>
                      <div className="flex items-center justify-between">
                        <div>
                          <CardTitle>Raw Data</CardTitle>
                          <CardDescription>
                            Click to view all fields and records
                          </CardDescription>
                        </div>
                        <div className="text-gray-400 group-open:rotate-90 transition-transform">
                          ▶
                        </div>
                      </div>
                    </CardHeader>
                  </Card>
                </summary>
                
                <Card className="mt-2">
                  <CardContent className="pt-6">
                    <div className="overflow-x-auto">
                      <table className="min-w-full divide-y divide-gray-200">
                        <thead className="bg-gray-50">
                          <tr>
                            {statementData.columns?.slice(0, 10).map((col: string) => (
                              <th
                                key={col}
                                className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                              >
                                {col}
                              </th>
                            ))}
                          </tr>
                        </thead>
                        <tbody className="bg-white divide-y divide-gray-200">
                          {statementData.rows?.slice(0, 10).map((row: any, index: number) => (
                            <tr key={index} className="hover:bg-gray-50">
                              {statementData.columns?.slice(0, 10).map((col: string) => (
                                <td
                                  key={col}
                                  className="px-3 py-2 whitespace-nowrap text-sm text-gray-900"
                                >
                                  {renderFieldValue(col, row[col])}
                                </td>
                              ))}
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                    {statementData.rows?.length > 10 && (
                      <p className="text-sm text-gray-500 mt-4">
                        Showing first 10 of {statementData.rows.length} records
                      </p>
                    )}
                  </CardContent>
                </Card>
              </details>
            </>
          ) : error ? (
            <Alert className="border-red-200 bg-red-50">
              <AlertCircle className="h-4 w-4 text-red-600" />
              <AlertDescription className="text-red-800">
                {error}
              </AlertDescription>
            </Alert>
          ) : null}
        </>
      )}
    </div>
  )
}

export default OwnerStatementViewer