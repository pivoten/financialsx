
import { useState, useEffect } from 'react'
import { AuditCheckBatches, AuditBankReconciliation, AuditSingleBankAccount, GetBankAccountsForAudit } from '../../wailsjs/go/main/App'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Badge } from './ui/badge'
import { Input } from './ui/input'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './ui/table'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './ui/tabs'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from './ui/dialog'
import { AlertTriangle, CheckCircle, XCircle, FileSearch, Download, Loader2, AlertCircle, Search, ArrowUpDown, ArrowUp, ArrowDown, Eye, EyeOff, Settings, RefreshCw } from 'lucide-react'
import type { User, AuditResult, ChangeEvent } from '../types'

interface CheckAuditProps {
  companyName: string
  currentUser: User
}

interface BankAccount {
  account_number: string
  account_name: string
  gl_balance: number
  outstanding_checks: number
  outstanding_count: number
}

interface AuditEntry {
  check_id: string
  amount: number
  check_amount?: number  // Add this property for compatibility
  row_index: number
  issue_type: string
  check_data?: any[]
  check_columns?: string[]
  gl_entries?: GLEntry[]
  [key: string]: any // Add index signature for dynamic access
}

interface GLEntry {
  amount: number
  row?: any[]
  columns?: string[]
}

interface AuditResultExtended {
  status?: string
  message?: string
  error?: string
  audit_date?: string
  audited_by?: string
  summary?: {
    total_checks: number
    matched_entries: number
    missing_entries: number
    mismatched_amounts: number
  }
  missing_entries: AuditEntry[]
  mismatched_amounts: AuditEntry[]
  check_columns?: string[]
}

interface BankAuditResult {
  issue_type: string
  gl_balance: number
  outstanding_checks: number
  outstanding_count: number
  reconciliation_balance: number | null
  reconciliation_date: string | null
  difference: number | null
  expected_gl_balance?: number
  audit_timestamp: string
  audited_by: string
}

interface ColumnVisibility {
  [key: string]: boolean
}

export function CheckAudit({ companyName, currentUser }: CheckAuditProps) {
  const [auditResults, setAuditResults] = useState<AuditResultExtended | null>(null)
  const [bankRecAuditResults, setBankRecAuditResults] = useState<any>(null)
  const [bankRecLoading, setBankRecLoading] = useState<boolean>(false)
  const [bankRecError, setBankRecError] = useState<string | null>(null)
  const [availableAccounts, setAvailableAccounts] = useState<BankAccount[]>([])
  const [auditResults_v2, setAuditResults_v2] = useState<Map<string, BankAuditResult>>(new Map())
  const [selectedAccount, setSelectedAccount] = useState<string>('')
  const [auditingAccount, setAuditingAccount] = useState<string | null>(null)
  const [loadingAccounts, setLoadingAccounts] = useState<boolean>(false)
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [showDetails, setShowDetails] = useState<boolean>(false)
  const [selectedEntry, setSelectedEntry] = useState<AuditEntry | null>(null)
  const [missingPage, setMissingPage] = useState<number>(1)
  const [mismatchedPage, setMismatchedPage] = useState<number>(1)
  const [itemsPerPage, setItemsPerPage] = useState<number>(10)
  const [missingSearch, setMissingSearch] = useState<string>('')
  const [mismatchedSearch, setMismatchedSearch] = useState<string>('')
  const [missingSortColumn, setMissingSortColumn] = useState<string>('')
  const [missingSortDirection, setMissingSortDirection] = useState<'asc' | 'desc'>('asc')
  const [mismatchedSortColumn, setMismatchedSortColumn] = useState<string>('')
  const [mismatchedSortDirection, setMismatchedSortDirection] = useState<'asc' | 'desc'>('asc')
  const [missingVisibleColumns, setMissingVisibleColumns] = useState<ColumnVisibility>({ check_id: true, amount: true, row_index: true })
  const [mismatchedVisibleColumns, setMismatchedVisibleColumns] = useState<ColumnVisibility>({ check_id: true, check_amount: true, gl_amount: true, difference: true })

  const canAudit = currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')

  const runAudit = async () => {
    try {
      setLoading(true)
      setError(null)
      setAuditResults(null)
      setMissingPage(1)
      setMismatchedPage(1)
      const results = await AuditCheckBatches(companyName)
      if (results.status === 'error') setError(results.message || results.error)
      else setAuditResults(results as AuditResultExtended)
    } catch (err) {
      setError(err.message || 'Failed to run audit')
    } finally {
      setLoading(false)
    }
  }

  const runBankReconciliationAudit = async () => {
    try {
      setBankRecLoading(true)
      setBankRecError(null)
      setBankRecAuditResults(null)
      const results = await AuditBankReconciliation(companyName)
      if (results.status === 'error') setBankRecError(results.message || results.error)
      else setBankRecAuditResults(results)
    } catch (err) {
      setBankRecError(err.message || 'Failed to run bank reconciliation audit')
    } finally {
      setBankRecLoading(false)
    }
  }

  const loadBankAccountsForAudit = async () => {
    try {
      setLoadingAccounts(true)
      setBankRecError(null)
      const accounts = await GetBankAccountsForAudit(companyName)
      setAvailableAccounts(accounts as BankAccount[])
      if (accounts.length > 0 && !selectedAccount) setSelectedAccount(accounts[0].account_number)
    } catch (err) {
      setBankRecError('Failed to load bank accounts: ' + err.message)
    } finally {
      setLoadingAccounts(false)
    }
  }

  const auditSingleAccount = async (accountNumber: string) => {
    try {
      setAuditingAccount(accountNumber)
      setBankRecError(null)
      const result = await AuditSingleBankAccount(companyName, accountNumber)
      if (result.status === 'error') {
        setBankRecError(result.message || result.error)
      } else {
        const newResults = new Map(auditResults_v2)
        newResults.set(accountNumber, result as BankAuditResult)
        setAuditResults_v2(newResults)
      }
    } catch (err) {
      setBankRecError(`Failed to audit account ${accountNumber}: ${err.message}`)
    } finally {
      setAuditingAccount(null)
    }
  }

  const refreshAccountAudit = async (accountNumber: string) => {
    const newResults = new Map(auditResults_v2)
    newResults.delete(accountNumber)
    setAuditResults_v2(newResults)
    await auditSingleAccount(accountNumber)
  }

  useEffect(() => {
    if (companyName && canAudit) loadBankAccountsForAudit()
  }, [companyName, canAudit])

  const formatCurrency = (amount: number) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount || 0)

  const formatFieldValue = (value: any, fieldName?: string): string => {
    const upperFieldName = fieldName ? fieldName.toUpperCase() : ''
    if (value === null || value === undefined || value === '') {
      if (upperFieldName === 'LPRINTED') return 'Not Printed'
      if (upperFieldName === 'LVOID') return 'Not Voided'
      if (upperFieldName === 'LCLEARED') return 'Not Cleared'
      if (upperFieldName === 'LDELETED') return 'Not Deleted'
      if (upperFieldName === 'LMANUAL') return 'Not Manual'
      if (upperFieldName === 'LDEPOSITED') return 'Not Deposited'
      if (upperFieldName.startsWith('L') && upperFieldName.length > 1) return 'False'
      return 'null'
    }
    if (typeof value === 'boolean') return value ? 'True' : 'False'
    if (typeof value === 'string') {
      const upperVal = value.toUpperCase().trim()
      if (upperVal === 'T' || upperVal === '.T.' || upperVal === 'TRUE') return 'True'
      if (upperVal === 'F' || upperVal === '.F.' || upperVal === 'FALSE') return 'False'
    }
    return value
  }

  const exportResults = () => {
    if (!auditResults) return
    const csvContent = [
      ['Audit Report - ' + auditResults.audit_date],
      ['Audited by: ' + auditResults.audited_by],
      [],
      ['Summary'],
      ['Total Checks', auditResults.summary.total_checks],
      ['Matched Entries', auditResults.summary.matched_entries],
      ['Missing Entries', auditResults.summary.missing_entries],
      ['Mismatched Amounts', auditResults.summary.mismatched_amounts],
      [],
      ['checks.dbf Column Structure'],
      ...(auditResults.check_columns || []).map((col, idx) => [`Column ${idx}`, col]),
      [],
      ['Missing Entries'],
      ['Check ID', 'Amount', 'Row Index', ...(auditResults.check_columns || [])],
      ...(auditResults.missing_entries || []).map(entry => [entry.check_id, entry.amount, entry.row_index, ...(entry.check_data || [])]),
      [],
      ['Mismatched Amounts'],
      ['Check ID', 'Check Amount', 'GL Amount', 'Row Index'],
      ...(auditResults.mismatched_amounts || []).map(entry => [entry.check_id, entry.check_amount || (entry as any).amount, entry.gl_entries?.[0]?.amount || 'N/A', entry.row_index])
    ]
    const csv = csvContent.map(row => row.map(cell => typeof cell === 'string' && cell.includes(',') ? `"${cell}"` : cell).join(',')).join('\n')
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit_report_${new Date().toISOString().split('T')[0]}.csv`
    a.click()
  }

  const showEntryDetails = (entry: AuditEntry) => { setSelectedEntry(entry); setShowDetails(true) }

  const handleSort = (column: string, type: 'missing' | 'mismatched') => {
    if (type === 'missing') {
      if (missingSortColumn === column) setMissingSortDirection(missingSortDirection === 'asc' ? 'desc' : 'asc')
      else { setMissingSortColumn(column); setMissingSortDirection('asc') }
      setMissingPage(1)
    } else {
      if (mismatchedSortColumn === column) setMismatchedSortDirection(mismatchedSortDirection === 'asc' ? 'desc' : 'asc')
      else { setMismatchedSortColumn(column); setMismatchedSortDirection('asc') }
      setMismatchedPage(1)
    }
  }

  const getSortIcon = (column: string, type: 'missing' | 'mismatched') => {
    const sortColumn = type === 'missing' ? missingSortColumn : mismatchedSortColumn
    const sortDirection = type === 'missing' ? missingSortDirection : mismatchedSortDirection
    if (sortColumn !== column) return <ArrowUpDown className="w-4 h-4" />
    return sortDirection === 'asc' ? <ArrowUp className="w-4 h-4" /> : <ArrowDown className="w-4 h-4" />
  }

  const filterAndSortData = (data: AuditEntry[], searchTerm: string, sortColumn: string, sortDirection: 'asc' | 'desc') => {
    let filtered = data
    if (searchTerm) {
      filtered = data.filter(entry => String(entry.check_id || '').toLowerCase().includes(searchTerm.toLowerCase()) || String(entry.amount || (entry as any).check_amount || '').includes(searchTerm) || String(entry.row_index || '').includes(searchTerm))
    }
    if (sortColumn) {
      filtered = [...filtered].sort((a, b) => {
        let aVal = (a as any)[sortColumn]
        let bVal = (b as any)[sortColumn]
        if (sortColumn.includes('amount') || sortColumn === 'row_index') {
          aVal = parseFloat(aVal) || 0
          bVal = parseFloat(bVal) || 0
        }
        if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1
        if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1
        return 0
      })
    }
    return filtered
  }

  const toggleColumnVisibility = (column: string, type: 'missing' | 'mismatched') => {
    if (type === 'missing') setMissingVisibleColumns(prev => ({ ...prev, [column]: !prev[column] }))
    else setMismatchedVisibleColumns(prev => ({ ...prev, [column]: !prev[column] }))
  }

  const handleItemsPerPageChange = (newValue: number) => { setItemsPerPage(newValue); setMissingPage(1); setMismatchedPage(1) }

  const renderPagination = (currentPage: number, setPage: (page: number) => void, totalItems: number, label: string) => {
    const totalPages = Math.ceil(totalItems / itemsPerPage)
    if (totalPages <= 1) return null
    return (
      <div className="flex flex-col gap-4 mt-4">
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">Showing {((currentPage - 1) * itemsPerPage) + 1} to {Math.min(currentPage * itemsPerPage, totalItems)} of {totalItems} {label}</p>
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">Items per page:</span>
            <select value={itemsPerPage} onChange={(e: ChangeEvent<HTMLSelectElement>) => handleItemsPerPageChange(Number(e.target.value))} className="h-8 rounded-md border border-input bg-background px-2 text-sm">
              <option value={10}>10</option>
              <option value={25}>25</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
            </select>
          </div>
        </div>
        <div className="flex items-center justify-center gap-2">
          <Button variant="outline" size="sm" onClick={() => setPage(currentPage - 1)} disabled={currentPage === 1}>Previous</Button>
          <div className="flex items-center gap-1">
            <Button variant={currentPage === 1 ? 'default' : 'outline'} size="sm" onClick={() => setPage(1)}>1</Button>
            {currentPage > 3 && <span className="px-2">...</span>}
            {Array.from({ length: totalPages }, (_, i) => i + 1)
              .filter(page => page !== 1 && page !== totalPages && Math.abs(page - currentPage) <= 1)
              .map(page => (
                <Button key={page} variant={currentPage === page ? 'default' : 'outline'} size="sm" onClick={() => setPage(page)}>{page}</Button>
              ))}
            {currentPage < totalPages - 2 && <span className="px-2">...</span>}
            {totalPages > 1 && (
              <Button variant={currentPage === totalPages ? 'default' : 'outline'} size="sm" onClick={() => setPage(totalPages)}>{totalPages}</Button>
            )}
          </div>
          <Button variant="outline" size="sm" onClick={() => setPage(currentPage + 1)} disabled={currentPage === totalPages}>Next</Button>
        </div>
      </div>
    )
  }

  if (!canAudit) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Check Batch Audit</CardTitle>
          <CardDescription>Verify check entries against General Ledger</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8">
            <AlertCircle className="w-12 h-12 text-yellow-500 mx-auto mb-4" />
            <p className="text-muted-foreground">You need Admin or Root privileges to access the audit feature</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Check Batch Audit</CardTitle>
              <CardDescription>Compare checks.dbf entries with GLMASTER.dbf to find discrepancies</CardDescription>
            </div>
            <div className="flex gap-2">
              {auditResults && (
                <Button variant="outline" size="sm" onClick={exportResults}>
                  <Download className="w-4 h-4 mr-2" />
                  Export CSV
                </Button>
              )}
              <Button onClick={runAudit} disabled={loading}>
                {loading ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Running Audit...
                  </>
                ) : (
                  <>
                    <FileSearch className="w-4 h-4 mr-2" />
                    Run Audit
                  </>
                )}
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-800 rounded-lg p-4 mb-4">
              <div className="flex items-center gap-2">
                <AlertTriangle className="w-5 h-5" />
                <p className="font-medium">Audit Error</p>
              </div>
              <p className="mt-1 text-sm">{error}</p>
            </div>
          )}

          {auditResults && (
            <div className="space-y-6">
              <div className="grid gap-4 md:grid-cols-4">
                <Card><CardContent className="p-4"><div className="text-2xl font-bold">{auditResults.summary.total_checks}</div><p className="text-sm text-muted-foreground">Total Checks</p></CardContent></Card>
                <Card><CardContent className="p-4"><div className="text-2xl font-bold text-green-600">{auditResults.summary.matched_entries}</div><p className="text-sm text-muted-foreground">Matched</p></CardContent></Card>
                <Card><CardContent className="p-4"><div className="text-2xl font-bold text-red-600">{auditResults.summary.missing_entries}</div><p className="text-sm text-muted-foreground">Missing</p></CardContent></Card>
                <Card><CardContent className="p-4"><div className="text-2xl font-bold text-yellow-600">{auditResults.summary.mismatched_amounts}</div><p className="text-sm text-muted-foreground">Mismatched</p></CardContent></Card>
              </div>

              <Tabs defaultValue="summary" className="w-full">
                <TabsList className="grid w-full grid-cols-4">
                  <TabsTrigger value="summary">Summary</TabsTrigger>
                  <TabsTrigger value="missing" className="flex items-center gap-2">Missing Entries {auditResults.summary.missing_entries > 0 && (<Badge variant="destructive" className="ml-1">{auditResults.summary.missing_entries}</Badge>)}</TabsTrigger>
                  <TabsTrigger value="mismatched" className="flex items-center gap-2">Mismatched Amounts {auditResults.summary.mismatched_amounts > 0 && (<Badge variant="secondary" className="ml-1 bg-yellow-100 text-yellow-800">{auditResults.summary.mismatched_amounts}</Badge>)}</TabsTrigger>
                  <TabsTrigger value="matched" className="flex items-center gap-2">Matched Entries {auditResults.summary.matched_entries > 0 && (<Badge variant="default" className="ml-1 bg-green-100 text-green-800">{auditResults.summary.matched_entries}</Badge>)}</TabsTrigger>
                </TabsList>

                <TabsContent value="summary" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Audit Summary</CardTitle>
                      <CardDescription>Overview of audit results</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        <div className="text-center py-8">
                          {auditResults.summary.missing_entries === 0 && auditResults.summary.mismatched_amounts === 0 ? (
                            <>
                              <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
                              <h3 className="text-lg font-semibold mb-2">All Checks Verified</h3>
                              <p className="text-muted-foreground">All {auditResults.summary.total_checks} check entries have matching GL entries with correct amounts</p>
                            </>
                          ) : (
                            <>
                              <AlertTriangle className="w-16 h-16 text-yellow-500 mx-auto mb-4" />
                              <h3 className="text-lg font-semibold mb-2">Issues Found</h3>
                              <p className="text-muted-foreground mb-4">Found {auditResults.summary.missing_entries + auditResults.summary.mismatched_amounts} issues that require attention</p>
                            </>
                          )}
                        </div>
                        <div className="text-sm text-muted-foreground border-t pt-4">
                          <p>Audit performed on: {auditResults.audit_date}</p>
                          <p>Audited by: {auditResults.audited_by}</p>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>

                <TabsContent value="missing" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <div className="flex items-center justify-between">
                        <div>
                          <CardTitle className="flex items-center gap-2"><XCircle className="w-5 h-5 text-red-500" />Missing GL Entries ({auditResults.missing_entries.length})</CardTitle>
                          <CardDescription>Check entries without corresponding GL records</CardDescription>
                        </div>
                        <div className="flex items-center gap-2">
                          <div className="flex items-center gap-2">
                            <Search className="w-4 h-4" />
                            <Input placeholder="Search missing entries..." value={missingSearch} onChange={(e: ChangeEvent<HTMLInputElement>) => setMissingSearch(e.target.value)} className="w-64" />
                          </div>
                          <Button variant="outline" size="sm" onClick={() => {
                            const newState = !Object.values(missingVisibleColumns).every(v => v)
                            setMissingVisibleColumns({ check_id: newState, amount: newState, row_index: newState })
                          }}>
                            <Settings className="w-4 h-4 mr-2" />
                            Columns
                          </Button>
                        </div>
                      </div>
                    </CardHeader>
                    <CardContent>
                      {(() => {
                        const filteredData = filterAndSortData(auditResults.missing_entries, missingSearch, missingSortColumn, missingSortDirection)
                        const paginatedData = filteredData.slice((missingPage - 1) * itemsPerPage, missingPage * itemsPerPage)
                        return (
                          <>
                            <div className="rounded-md border">
                              <Table>
                                <TableHeader className="sticky top-0 bg-background">
                                  <TableRow>
                                    {missingVisibleColumns.check_id && (
                                      <TableHead className="cursor-pointer hover:bg-muted/50" onClick={() => handleSort('check_id', 'missing')}>
                                        <div className="flex items-center gap-2">Check ID/Batch {getSortIcon('check_id', 'missing')}</div>
                                      </TableHead>
                                    )}
                                    {missingVisibleColumns.amount && (
                                      <TableHead className="text-right cursor-pointer hover:bg-muted/50" onClick={() => handleSort('amount', 'missing')}>
                                        <div className="flex items-center justify-end gap-2">Amount {getSortIcon('amount', 'missing')}</div>
                                      </TableHead>
                                    )}
                                    {missingVisibleColumns.row_index && (
                                      <TableHead className="cursor-pointer hover:bg-muted/50" onClick={() => handleSort('row_index', 'missing')}>
                                        <div className="flex items-center gap-2">Row {getSortIcon('row_index', 'missing')}</div>
                                      </TableHead>
                                    )}
                                    <TableHead>Actions</TableHead>
                                  </TableRow>
                                </TableHeader>
                                <TableBody>
                                  {paginatedData.map((entry, idx) => (
                                    <TableRow key={idx}>
                                      {missingVisibleColumns.check_id && (<TableCell className="font-mono">{entry.check_id}</TableCell>)}
                                      {missingVisibleColumns.amount && (<TableCell className="text-right font-mono">{formatCurrency(entry.amount)}</TableCell>)}
                                      {missingVisibleColumns.row_index && (<TableCell>{entry.row_index}</TableCell>)}
                                      <TableCell>
                                        <Button variant="ghost" size="sm" onClick={() => showEntryDetails(entry)}>View Details</Button>
                                      </TableCell>
                                    </TableRow>
                                  ))}
                                </TableBody>
                              </Table>
                            </div>
                            {renderPagination(missingPage, setMissingPage, filteredData.length, 'missing entries')}
                          </>
                        )
                      })()}
                    </CardContent>
                  </Card>
                </TabsContent>

                <TabsContent value="mismatched" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2"><AlertTriangle className="w-5 h-5 text-yellow-500" />Mismatched Amounts ({auditResults.mismatched_amounts.length})</CardTitle>
                      <CardDescription>Entries with amount differences between checks and GL</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <p className="text-sm text-muted-foreground">Mismatched amounts feature coming soon...</p>
                    </CardContent>
                  </Card>
                </TabsContent>

                <TabsContent value="matched" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2"><CheckCircle className="w-5 h-5 text-green-500" />Matched Entries ({auditResults.summary.matched_entries})</CardTitle>
                      <CardDescription>Successfully verified check entries</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8">
                        <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
                        <h3 className="text-lg font-semibold mb-2">All Verified</h3>
                        <p className="text-muted-foreground">{auditResults.summary.matched_entries} check entries were successfully verified</p>
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>
            </div>
          )}

          {!auditResults && !error && !loading && (
            <div className="text-center py-12">
              <FileSearch className="w-16 h-16 text-gray-400 mx-auto mb-4" />
              <p className="text-muted-foreground mb-4">Click "Run Audit" to compare check entries with the General Ledger</p>
              <p className="text-sm text-muted-foreground">This will analyze the CBATCH field in checks.dbf against GLMASTER.dbf</p>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Bank Reconciliation Audit</CardTitle>
              <CardDescription>Select and audit individual bank accounts. Results are saved until refreshed.</CardDescription>
            </div>
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <label className="text-sm font-medium">Account:</label>
                <select 
                  value={selectedAccount} 
                  onChange={(e: ChangeEvent<HTMLSelectElement>) => setSelectedAccount(e.target.value)}
                  className="w-64 p-2 border rounded-md bg-background"
                >
                  <option value="">Select account to audit...</option>
                  {availableAccounts?.map((account) => (
                    <option key={account.account_number} value={account.account_number}>
                      {account.account_number} - {account.account_name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="flex gap-2">
                <Button onClick={() => selectedAccount && auditSingleAccount(selectedAccount)} disabled={!selectedAccount || auditingAccount === selectedAccount || loadingAccounts} size="sm">
                  {auditingAccount === selectedAccount ? (
                    <>
                      <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                      Auditing...
                    </>
                  ) : (
                    <>
                      <FileSearch className="w-4 h-4 mr-2" />
                      Audit Account
                    </>
                  )}
                </Button>
                {selectedAccount && auditResults_v2.has(selectedAccount) && (
                  <Button onClick={() => refreshAccountAudit(selectedAccount)} disabled={auditingAccount === selectedAccount} variant="outline" size="sm">
                    <RefreshCw className="w-4 h-4 mr-2" />
                    Refresh
                  </Button>
                )}
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {bankRecError && (
            <div className="bg-red-50 border border-red-200 text-red-800 rounded-lg p-4 mb-4">
              <div className="flex items-center gap-2">
                <AlertTriangle className="w-4 h-4" />
                <span className="font-medium">Audit Error</span>
              </div>
              <p className="mt-1 text-sm">{bankRecError}</p>
            </div>
          )}

          {loadingAccounts && (
            <div className="text-center py-8">
              <Loader2 className="w-8 h-8 mx-auto mb-4 animate-spin text-muted-foreground" />
              <p className="text-muted-foreground">Loading bank accounts...</p>
            </div>
          )}

          {!loadingAccounts && availableAccounts.length > 0 && (
            <div className="space-y-6">
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Account</TableHead>
                      <TableHead className="text-right">GL Balance</TableHead>
                      <TableHead className="text-right">Outstanding Checks</TableHead>
                      <TableHead className="text-center">Last Audited</TableHead>
                      <TableHead className="text-center">Status</TableHead>
                      <TableHead className="text-center">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {availableAccounts.map((account) => {
                      const result = auditResults_v2.get(account.account_number)
                      const isAuditing = auditingAccount === account.account_number
                      let statusBadge
                      if (isAuditing) {
                        statusBadge = <Badge variant="secondary" className="bg-blue-100 text-blue-800"><Loader2 className="w-3 h-3 mr-1 animate-spin" />Auditing</Badge>
                      } else if (result) {
                        if (result.issue_type === 'balanced') {
                          statusBadge = <Badge variant="default" className="bg-green-100 text-green-800"><CheckCircle className="w-3 h-3 mr-1" />Balanced</Badge>
                        } else if (result.issue_type === 'discrepancy_found') {
                          statusBadge = <Badge variant="destructive"><AlertTriangle className="w-3 h-3 mr-1" />Discrepancy</Badge>
                        } else {
                          statusBadge = <Badge variant="outline"><AlertCircle className="w-3 h-3 mr-1" />No Reconciliation</Badge>
                        }
                      } else {
                        statusBadge = <Badge variant="outline">Not Audited</Badge>
                      }

                      return (
                        <TableRow key={account.account_number} className={selectedAccount === account.account_number ? 'bg-blue-50' : ''}>
                          <TableCell className="font-medium">
                            <div>
                              <div>{account.account_number}</div>
                              <div className="text-sm text-muted-foreground">{account.account_name}</div>
                            </div>
                          </TableCell>
                          <TableCell className="text-right font-mono">{formatCurrency(account.gl_balance)}</TableCell>
                          <TableCell className="text-right font-mono">
                            {account.outstanding_checks > 0 ? (
                              <span className="text-amber-600">{formatCurrency(account.outstanding_checks)} ({account.outstanding_count})</span>
                            ) : (
                              <span className="text-muted-foreground">None</span>
                            )}
                          </TableCell>
                          <TableCell className="text-center">
                            {result ? (
                              <span className="text-sm text-muted-foreground">{new Date(result.audit_timestamp).toLocaleString()}</span>
                            ) : (
                              <span className="text-sm text-muted-foreground">Never</span>
                            )}
                          </TableCell>
                          <TableCell className="text-center">{statusBadge}</TableCell>
                          <TableCell className="text-center">
                            <div className="flex gap-1 justify-center">
                              <Button size="sm" variant="outline" onClick={() => { setSelectedAccount(account.account_number); auditSingleAccount(account.account_number) }} disabled={isAuditing}>
                                <FileSearch className="w-3 h-3" />
                              </Button>
                              {result && (
                                <Button size="sm" variant="outline" onClick={() => refreshAccountAudit(account.account_number)} disabled={isAuditing}>
                                  <RefreshCw className="w-3 h-3" />
                                </Button>
                              )}
                            </div>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              </div>

              {selectedAccount && auditResults_v2.has(selectedAccount) && (
                <Card>
                  <CardHeader>
                    <CardTitle>Audit Results: {selectedAccount}</CardTitle>
                    <CardDescription>Detailed reconciliation information for the selected account</CardDescription>
                  </CardHeader>
                  <CardContent>
                    {(() => {
                      const result = auditResults_v2.get(selectedAccount)
                      return (
                        <div className="space-y-4">
                          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                            <div className="bg-gray-50 p-4 rounded-lg"><div className="text-sm text-muted-foreground">GL Balance</div><div className="text-xl font-bold">{formatCurrency(result.gl_balance)}</div></div>
                            <div className="bg-gray-50 p-4 rounded-lg"><div className="text-sm text-muted-foreground">Outstanding Checks</div><div className="text-xl font-bold text-amber-600">{formatCurrency(result.outstanding_checks)}{result.outstanding_count > 0 && (<span className="text-sm ml-2">({result.outstanding_count})</span>)}</div></div>
                            <div className="bg-gray-50 p-4 rounded-lg"><div className="text-sm text-muted-foreground">Reconciliation Balance</div><div className="text-xl font-bold">{result.reconciliation_balance ? formatCurrency(result.reconciliation_balance) : 'N/A'}</div>{result.reconciliation_date && (<div className="text-xs text-muted-foreground mt-1">Date: {result.reconciliation_date}</div>)}</div>
                            <div className="bg-gray-50 p-4 rounded-lg"><div className="text-sm text-muted-foreground">Difference</div><div className={`text-xl font-bold ${result.difference === null ? 'text-muted-foreground' : Math.abs(result.difference) < 0.01 ? 'text-green-600' : 'text-red-600'}`}>{result.difference !== null ? formatCurrency(result.difference) : 'N/A'}</div></div>
                          </div>
                          <div className="bg-blue-50 p-4 rounded-lg">
                            <div className="text-sm font-medium text-blue-800 mb-2">Reconciliation Formula</div>
                            <div className="text-sm text-blue-700">Bank Statement Balance - Outstanding Checks = GL Balance</div>
                            {result.reconciliation_balance !== null && (
                              <div className="text-sm text-blue-600 mt-2 font-mono">
                                {formatCurrency(result.reconciliation_balance)} - {formatCurrency(result.outstanding_checks)} = {formatCurrency(result.expected_gl_balance || 0)}
                                {result.difference !== null && (
                                  <span className="ml-2">(Actual GL: {formatCurrency(result.gl_balance)}, Difference: {formatCurrency(result.difference)})</span>
                                )}
                              </div>
                            )}
                          </div>
                          <div className="text-sm text-muted-foreground">Audited on {new Date(result.audit_timestamp).toLocaleString()} by {result.audited_by}</div>
                        </div>
                      )
                    })()}
                  </CardContent>
                </Card>
              )}
            </div>
          )}

          {!loadingAccounts && availableAccounts.length === 0 && !bankRecError && (
            <div className="text-center py-8">
              <AlertCircle className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
              <h3 className="text-lg font-medium mb-2">No Bank Accounts Found</h3>
              <p className="text-muted-foreground">No bank accounts found in Chart of Accounts. Make sure LBANKACCT is set to true for bank accounts in COA.dbf.</p>
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={showDetails} onOpenChange={setShowDetails}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Entry Details</DialogTitle>
            <DialogDescription>Full details for {selectedEntry?.check_id}</DialogDescription>
          </DialogHeader>
          {selectedEntry && (
            <div className="space-y-4">
              <div>
                <h4 className="font-medium mb-2">Summary</h4>
                <div className="grid grid-cols-2 gap-4">
                  <div className="flex justify-between"><span className="text-sm text-muted-foreground">Check ID:</span><span className="font-mono">{selectedEntry.check_id}</span></div>
                  <div className="flex justify-between"><span className="text-sm text-muted-foreground">Amount:</span><span className="font-mono">{formatCurrency(selectedEntry.amount)}</span></div>
                  <div className="flex justify-between"><span className="text-sm text-muted-foreground">Issue Type:</span><span className="font-mono">{selectedEntry.issue_type}</span></div>
                  <div className="flex justify-between"><span className="text-sm text-muted-foreground">Row Index:</span><span className="font-mono">{selectedEntry.row_index}</span></div>
                </div>
              </div>
              {selectedEntry.check_data && (
                <div>
                  <h4 className="font-medium mb-2">Check Data</h4>
                  <div className="bg-gray-50 p-4 rounded-lg">
                    <div className="grid grid-cols-2 gap-2">
                      {selectedEntry.check_data.map((value, idx) => {
                        const fieldName = selectedEntry.check_columns?.[idx] || `Column ${idx}`
                        const formattedValue = formatFieldValue(value, fieldName)
                        return (
                          <div key={idx} className="text-sm">
                            <span className="text-muted-foreground">{fieldName}:</span>{' '}
                            <span className="font-mono">{formattedValue}</span>
                          </div>
                        )
                      })}
                    </div>
                  </div>
                </div>
              )}
              {selectedEntry.gl_entries && selectedEntry.gl_entries.length > 0 && (
                <div>
                  <h4 className="font-medium mb-2">GL Entries</h4>
                  {selectedEntry.gl_entries.map((glEntry, idx) => (
                    <div key={idx} className="bg-gray-50 p-4 rounded-lg mb-2">
                      <div className="flex justify-between mb-2"><span className="text-sm text-muted-foreground">GL Amount:</span><span className="font-mono">{formatCurrency(glEntry.amount)}</span></div>
                      {glEntry.row && (
                        <div className="grid grid-cols-2 gap-2">
                          {glEntry.row.map((value, colIdx) => {
                            const fieldName = glEntry.columns?.[colIdx] || `Column ${colIdx}`
                            const formattedValue = formatFieldValue(value, fieldName)
                            return (
                              <div key={colIdx} className="text-sm">
                                <span className="text-muted-foreground">{fieldName}:</span>{' '}
                                <span className="font-mono text-xs">{formattedValue}</span>
                              </div>
                            )
                          })}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  )
}
