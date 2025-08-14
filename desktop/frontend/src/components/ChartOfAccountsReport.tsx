import React, { useState, useEffect } from 'react'
import { Button } from './ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from './ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Label } from './ui/label'
import { FileDown, Printer, ArrowLeft, ChevronRight, Filter } from 'lucide-react'
import { GetChartOfAccounts, GenerateChartOfAccountsPDF } from '../../wailsjs/go/main/App'
import { Skeleton } from './ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from './ui/table'

interface ChartOfAccountsReportProps {
  companyName: string
  onBack?: () => void
}

interface Account {
  row_index: number
  account_number: string
  account_name: string
  account_type: string
  account_type_num: number
  is_bank_account: boolean
  parent_account: string
  is_unit: boolean
  is_department: boolean
  is_active: boolean
}

const ChartOfAccountsReport: React.FC<ChartOfAccountsReportProps> = ({ companyName, onBack }) => {
  const [loading, setLoading] = useState(false)
  const [accounts, setAccounts] = useState<Account[]>([])
  const [sortBy, setSortBy] = useState<'number' | 'type'>('number')
  const [includeInactive, setIncludeInactive] = useState(false)
  const [generatedAt, setGeneratedAt] = useState('')
  const [total, setTotal] = useState(0)
  const [downloading, setDownloading] = useState(false)

  const loadChartOfAccounts = async () => {
    setLoading(true)
    try {
      const result = await GetChartOfAccounts(companyName, sortBy, includeInactive)
      if (result && result.accounts) {
        setAccounts(result.accounts as Account[])
        setTotal(result.total || 0)
        setGeneratedAt(result.generated_at || new Date().toLocaleString())
      }
    } catch (error) {
      console.error('Failed to load chart of accounts:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadChartOfAccounts()
  }, [companyName, sortBy, includeInactive])

  const handleDownloadPDF = async () => {
    setDownloading(true)
    try {
      const filepath = await GenerateChartOfAccountsPDF(companyName, sortBy, includeInactive)
      if (filepath) {
        // Show success message with the saved file location
        alert(`PDF report saved successfully!\n\nLocation: ${filepath}`)
      }
    } catch (error: any) {
      console.error('Failed to generate PDF:', error)
      // Check if user cancelled the save dialog
      if (error?.message?.includes('cancelled')) {
        // User cancelled, no need to show error
      } else {
        alert('Failed to generate PDF report: ' + error?.message)
      }
    } finally {
      setDownloading(false)
    }
  }

  const getAccountTypeColor = (type: string) => {
    switch (type) {
      case 'Asset':
        return 'text-green-600'
      case 'Liability':
        return 'text-red-600'
      case 'Equity':
        return 'text-blue-600'
      case 'Revenue':
        return 'text-emerald-600'
      case 'Expense':
        return 'text-orange-600'
      default:
        return 'text-gray-600'
    }
  }

  const getAccountTypeBadge = (type: string) => {
    const color = getAccountTypeColor(type)
    return (
      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${color} bg-opacity-10 ${color.replace('text-', 'bg-')}`}>
        {type}
      </span>
    )
  }

  // Group accounts by type when sorting by type
  const groupAccountsByType = () => {
    const grouped: { [key: string]: Account[] } = {}
    accounts.forEach(account => {
      if (!grouped[account.account_type]) {
        grouped[account.account_type] = []
      }
      grouped[account.account_type].push(account)
    })
    return grouped
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          {onBack && (
            <Button variant="ghost" size="sm" onClick={onBack}>
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back
            </Button>
          )}
          <div>
            <h2 className="text-2xl font-bold text-gray-900">Chart of Accounts</h2>
            <p className="text-sm text-gray-500 mt-1">
              Complete listing of all general ledger accounts for {companyName}
            </p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleDownloadPDF}
            disabled={downloading || loading}
          >
            <FileDown className="h-4 w-4 mr-2" />
            {downloading ? 'Generating...' : 'Download PDF'}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => window.print()}
            disabled={loading}
          >
            <Printer className="h-4 w-4 mr-2" />
            Print
          </Button>
        </div>
      </div>

      {/* Filters Card */}
      <Card>
        <CardHeader className="pb-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Filter className="h-5 w-5 text-gray-500" />
              <CardTitle className="text-base">Report Options</CardTitle>
            </div>
            <div className="text-sm text-gray-500">
              Generated: {generatedAt}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center space-x-4">
            <div className="flex-1">
              <Label htmlFor="sort-by">Sort By</Label>
              <Select value={sortBy} onValueChange={(value: 'number' | 'type') => setSortBy(value)}>
                <SelectTrigger id="sort-by" className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="number">Account Number</SelectItem>
                  <SelectItem value="type">Account Type</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex-1">
              <Label htmlFor="filter">Account Status</Label>
              <Select value={includeInactive ? 'all' : 'active'} onValueChange={(value) => setIncludeInactive(value === 'all')}>
                <SelectTrigger id="filter" className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">Active Only</SelectItem>
                  <SelectItem value="all">Include Inactive Accounts</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex items-center space-x-2 text-sm text-gray-500">
              <span>Total Accounts:</span>
              <span className="font-semibold text-gray-900">{total}</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Accounts Table */}
      <Card>
        <CardHeader>
          <CardTitle>Account Listing</CardTitle>
          <CardDescription>
            {sortBy === 'type' ? 'Grouped by account type' : 'Sorted by account number'}
          </CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          {loading ? (
            <div className="p-6 space-y-3">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : sortBy === 'type' ? (
            // Grouped view by account type
            <div className="divide-y divide-gray-200">
              {Object.entries(groupAccountsByType()).map(([type, typeAccounts]) => (
                <div key={type} className="p-6">
                  <div className="mb-4">
                    <h3 className={`text-lg font-semibold ${getAccountTypeColor(type)}`}>
                      {type} Accounts
                    </h3>
                    <p className="text-sm text-gray-500">{typeAccounts.length} accounts</p>
                  </div>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[150px]">Account Number</TableHead>
                        <TableHead>Account Name</TableHead>
                        <TableHead className="w-[100px]">Bank Account</TableHead>
                        <TableHead className="w-[150px]">Parent Account</TableHead>
                        <TableHead className="w-[80px]">Unit</TableHead>
                        <TableHead className="w-[80px]">Dept</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {typeAccounts.map((account) => (
                        <TableRow key={account.account_number}>
                          <TableCell className="font-mono text-sm">{account.account_number}</TableCell>
                          <TableCell>{account.account_name}</TableCell>
                          <TableCell>
                            {account.is_bank_account && (
                              <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                                Bank
                              </span>
                            )}
                          </TableCell>
                          <TableCell className="font-mono text-sm">{account.parent_account || '-'}</TableCell>
                          <TableCell>{account.is_unit ? 'Yes' : '-'}</TableCell>
                          <TableCell>{account.is_department ? 'Yes' : '-'}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              ))}
            </div>
          ) : (
            // Simple table view sorted by account number
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[150px]">Account Number</TableHead>
                  <TableHead>Account Name</TableHead>
                  <TableHead className="w-[120px]">Type</TableHead>
                  <TableHead className="w-[100px]">Bank Account</TableHead>
                  <TableHead className="w-[150px]">Parent Account</TableHead>
                  <TableHead className="w-[80px]">Unit</TableHead>
                  <TableHead className="w-[80px]">Dept</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {accounts.map((account) => (
                  <TableRow key={account.account_number}>
                    <TableCell className="font-mono text-sm">{account.account_number}</TableCell>
                    <TableCell>{account.account_name}</TableCell>
                    <TableCell>{getAccountTypeBadge(account.account_type)}</TableCell>
                    <TableCell>
                      {account.is_bank_account && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                          Bank
                        </span>
                      )}
                    </TableCell>
                    <TableCell className="font-mono text-sm">{account.parent_account || '-'}</TableCell>
                    <TableCell>{account.is_unit ? 'Yes' : '-'}</TableCell>
                    <TableCell>{account.is_department ? 'Yes' : '-'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Print-only styles */}
      <style>{`
        @media print {
          .no-print {
            display: none !important;
          }
          @page {
            margin: 0.5in;
          }
          body {
            print-color-adjust: exact;
            -webkit-print-color-adjust: exact;
          }
        }
      `}</style>
    </div>
  )
}

export default ChartOfAccountsReport