import { useState } from 'react'
import { AuditCheckBatches, AuditBankReconciliation } from '../../wailsjs/go/main/App'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Badge } from './ui/badge'
import { Input } from './ui/input'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './ui/table'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './ui/tabs'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from './ui/dialog'
import { 
  AlertTriangle,
  CheckCircle,
  XCircle,
  FileSearch,
  Download,
  Loader2,
  AlertCircle,
  Search,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Eye,
  EyeOff,
  Settings
} from 'lucide-react'

export function CheckAudit({ companyName, currentUser }) {
  const [auditResults, setAuditResults] = useState(null)
  const [bankRecAuditResults, setBankRecAuditResults] = useState(null)
  const [loading, setLoading] = useState(false)
  const [bankRecLoading, setBankRecLoading] = useState(false)
  const [error, setError] = useState(null)
  const [bankRecError, setBankRecError] = useState(null)
  const [showDetails, setShowDetails] = useState(false)
  const [selectedEntry, setSelectedEntry] = useState(null)
  const [missingPage, setMissingPage] = useState(1)
  const [mismatchedPage, setMismatchedPage] = useState(1)
  const [itemsPerPage, setItemsPerPage] = useState(10)
  
  // Search and filtering
  const [missingSearch, setMissingSearch] = useState('')
  const [mismatchedSearch, setMismatchedSearch] = useState('')
  
  // Sorting
  const [missingSortColumn, setMissingSortColumn] = useState('')
  const [missingSortDirection, setMissingSortDirection] = useState('asc')
  const [mismatchedSortColumn, setMismatchedSortColumn] = useState('')
  const [mismatchedSortDirection, setMismatchedSortDirection] = useState('asc')
  
  // Column visibility
  const [missingVisibleColumns, setMissingVisibleColumns] = useState({
    check_id: true,
    amount: true,
    row_index: true
  })
  const [mismatchedVisibleColumns, setMismatchedVisibleColumns] = useState({
    check_id: true,
    check_amount: true,
    gl_amount: true,
    difference: true
  })

  // Check if user has permissions
  const canAudit = currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')

  const runAudit = async () => {
    try {
      setLoading(true)
      setError(null)
      setAuditResults(null)
      setMissingPage(1)
      setMismatchedPage(1)
      
      const results = await AuditCheckBatches(companyName)
      
      if (results.status === 'error') {
        setError(results.message || results.error)
      } else {
        setAuditResults(results)
      }
    } catch (err) {
      console.error('Audit failed:', err)
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
      
      if (results.status === 'error') {
        setBankRecError(results.message || results.error)
      } else {
        setBankRecAuditResults(results)
      }
    } catch (err) {
      console.error('Bank reconciliation audit failed:', err)
      setBankRecError(err.message || 'Failed to run bank reconciliation audit')
    } finally {
      setBankRecLoading(false)
    }
  }

  const formatCurrency = (amount) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount || 0)
  }

  const formatFieldValue = (value, fieldName) => {
    // Get uppercase field name for comparison
    const upperFieldName = fieldName ? fieldName.toUpperCase() : ''
    
    // Handle null/undefined/empty
    if (value === null || value === undefined || value === '') {
      // Special handling for specific logical fields when empty
      if (upperFieldName === 'LPRINTED') {
        return 'Not Printed'
      }
      if (upperFieldName === 'LVOID') {
        return 'Not Voided'
      }
      if (upperFieldName === 'LCLEARED') {
        return 'Not Cleared'
      }
      if (upperFieldName === 'LDELETED') {
        return 'Not Deleted'
      }
      if (upperFieldName === 'LMANUAL') {
        return 'Not Manual'
      }
      if (upperFieldName === 'LDEPOSITED') {
        return 'Not Deposited'
      }
      // For other fields that start with 'L' (likely logical fields)
      if (upperFieldName.startsWith('L') && upperFieldName.length > 1) {
        return 'False'
      }
      return 'null'
    }
    
    // Handle boolean values
    if (typeof value === 'boolean') {
      return value ? 'True' : 'False'
    }
    
    // Handle logical field values (T/F, .T./.F.)
    if (typeof value === 'string') {
      const upperVal = value.toUpperCase().trim()
      if (upperVal === 'T' || upperVal === '.T.' || upperVal === 'TRUE') {
        return 'True'
      }
      if (upperVal === 'F' || upperVal === '.F.' || upperVal === 'FALSE') {
        return 'False'
      }
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
      ...auditResults.check_columns.map((col, idx) => [`Column ${idx}`, col]),
      [],
      ['Missing Entries'],
      ['Check ID', 'Amount', 'Row Index', ...(auditResults.check_columns || [])],
      ...auditResults.missing_entries.map(entry => [
        entry.check_id,
        entry.amount,
        entry.row_index,
        ...(entry.check_data || [])
      ]),
      [],
      ['Mismatched Amounts'],
      ['Check ID', 'Check Amount', 'GL Amount', 'Row Index'],
      ...auditResults.mismatched_amounts.map(entry => [
        entry.check_id,
        entry.check_amount,
        entry.gl_entries[0]?.amount || 'N/A',
        entry.row_index
      ])
    ]
    
    const csv = csvContent.map(row => row.map(cell => 
      typeof cell === 'string' && cell.includes(',') ? `"${cell}"` : cell
    ).join(',')).join('\n')
    
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit_report_${new Date().toISOString().split('T')[0]}.csv`
    a.click()
  }

  const showEntryDetails = (entry) => {
    setSelectedEntry(entry)
    setShowDetails(true)
  }

  const handleSort = (column, type) => {
    if (type === 'missing') {
      if (missingSortColumn === column) {
        setMissingSortDirection(missingSortDirection === 'asc' ? 'desc' : 'asc')
      } else {
        setMissingSortColumn(column)
        setMissingSortDirection('asc')
      }
      setMissingPage(1)
    } else {
      if (mismatchedSortColumn === column) {
        setMismatchedSortDirection(mismatchedSortDirection === 'asc' ? 'desc' : 'asc')
      } else {
        setMismatchedSortColumn(column)
        setMismatchedSortDirection('asc')
      }
      setMismatchedPage(1)
    }
  }

  const getSortIcon = (column, type) => {
    const sortColumn = type === 'missing' ? missingSortColumn : mismatchedSortColumn
    const sortDirection = type === 'missing' ? missingSortDirection : mismatchedSortDirection
    
    if (sortColumn !== column) return <ArrowUpDown className="w-4 h-4" />
    return sortDirection === 'asc' ? <ArrowUp className="w-4 h-4" /> : <ArrowDown className="w-4 h-4" />
  }

  const filterAndSortData = (data, searchTerm, sortColumn, sortDirection) => {
    let filtered = data
    
    if (searchTerm) {
      filtered = data.filter(entry => 
        String(entry.check_id || '').toLowerCase().includes(searchTerm.toLowerCase()) ||
        String(entry.amount || entry.check_amount || '').includes(searchTerm) ||
        String(entry.row_index || '').includes(searchTerm)
      )
    }
    
    if (sortColumn) {
      filtered = [...filtered].sort((a, b) => {
        let aVal = a[sortColumn]
        let bVal = b[sortColumn]
        
        // Handle numeric columns
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

  const toggleColumnVisibility = (column, type) => {
    if (type === 'missing') {
      setMissingVisibleColumns(prev => ({
        ...prev,
        [column]: !prev[column]
      }))
    } else {
      setMismatchedVisibleColumns(prev => ({
        ...prev,
        [column]: !prev[column]
      }))
    }
  }

  const handleItemsPerPageChange = (newValue) => {
    setItemsPerPage(newValue)
    setMissingPage(1)
    setMismatchedPage(1)
  }

  const renderPagination = (currentPage, setPage, totalItems, label) => {
    const totalPages = Math.ceil(totalItems / itemsPerPage)
    
    if (totalPages <= 1) return null
    
    return (
      <div className="flex flex-col gap-4 mt-4">
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {((currentPage - 1) * itemsPerPage) + 1} to {Math.min(currentPage * itemsPerPage, totalItems)} of {totalItems} {label}
          </p>
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">Items per page:</span>
            <select 
              value={itemsPerPage} 
              onChange={(e) => handleItemsPerPageChange(Number(e.target.value))}
              className="h-8 rounded-md border border-input bg-background px-2 text-sm"
            >
              <option value={10}>10</option>
              <option value={25}>25</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
            </select>
          </div>
        </div>
        <div className="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage(currentPage - 1)}
            disabled={currentPage === 1}
          >
            Previous
          </Button>
          <div className="flex items-center gap-1">
            {/* Show first page */}
            <Button
              variant={currentPage === 1 ? "default" : "outline"}
              size="sm"
              onClick={() => setPage(1)}
            >
              1
            </Button>
            
            {/* Show dots if needed */}
            {currentPage > 3 && <span className="px-2">...</span>}
            
            {/* Show current page and neighbors */}
            {Array.from({ length: totalPages }, (_, i) => i + 1)
              .filter(page => page !== 1 && page !== totalPages && Math.abs(page - currentPage) <= 1)
              .map(page => (
                <Button
                  key={page}
                  variant={currentPage === page ? "default" : "outline"}
                  size="sm"
                  onClick={() => setPage(page)}
                >
                  {page}
                </Button>
              ))}
            
            {/* Show dots if needed */}
            {currentPage < totalPages - 2 && <span className="px-2">...</span>}
            
            {/* Show last page */}
            {totalPages > 1 && (
              <Button
                variant={currentPage === totalPages ? "default" : "outline"}
                size="sm"
                onClick={() => setPage(totalPages)}
              >
                {totalPages}
              </Button>
            )}
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage(currentPage + 1)}
            disabled={currentPage === totalPages}
          >
            Next
          </Button>
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
            <p className="text-muted-foreground">
              You need Admin or Root privileges to access the audit feature
            </p>
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
              <CardDescription>
                Compare checks.dbf entries with GLMASTER.dbf to find discrepancies
              </CardDescription>
            </div>
            <div className="flex gap-2">
              {auditResults && (
                <Button 
                  variant="outline" 
                  size="sm"
                  onClick={exportResults}
                >
                  <Download className="w-4 h-4 mr-2" />
                  Export CSV
                </Button>
              )}
              <Button 
                onClick={runAudit}
                disabled={loading}
              >
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
              {/* Summary Cards */}
              <div className="grid gap-4 md:grid-cols-4">
                <Card>
                  <CardContent className="p-4">
                    <div className="text-2xl font-bold">{auditResults.summary.total_checks}</div>
                    <p className="text-sm text-muted-foreground">Total Checks</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardContent className="p-4">
                    <div className="text-2xl font-bold text-green-600">
                      {auditResults.summary.matched_entries}
                    </div>
                    <p className="text-sm text-muted-foreground">Matched</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardContent className="p-4">
                    <div className="text-2xl font-bold text-red-600">
                      {auditResults.summary.missing_entries}
                    </div>
                    <p className="text-sm text-muted-foreground">Missing</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardContent className="p-4">
                    <div className="text-2xl font-bold text-yellow-600">
                      {auditResults.summary.mismatched_amounts}
                    </div>
                    <p className="text-sm text-muted-foreground">Mismatched</p>
                  </CardContent>
                </Card>
              </div>

              {/* Tabbed Results */}
              <Tabs defaultValue="summary" className="w-full">
                <TabsList className="grid w-full grid-cols-4">
                  <TabsTrigger value="summary">Summary</TabsTrigger>
                  <TabsTrigger value="missing" className="flex items-center gap-2">
                    Missing Entries
                    {auditResults.summary.missing_entries > 0 && (
                      <Badge variant="destructive" className="ml-1">
                        {auditResults.summary.missing_entries}
                      </Badge>
                    )}
                  </TabsTrigger>
                  <TabsTrigger value="mismatched" className="flex items-center gap-2">
                    Mismatched Amounts
                    {auditResults.summary.mismatched_amounts > 0 && (
                      <Badge variant="secondary" className="ml-1 bg-yellow-100 text-yellow-800">
                        {auditResults.summary.mismatched_amounts}
                      </Badge>
                    )}
                  </TabsTrigger>
                  <TabsTrigger value="matched" className="flex items-center gap-2">
                    Matched Entries
                    {auditResults.summary.matched_entries > 0 && (
                      <Badge variant="default" className="ml-1 bg-green-100 text-green-800">
                        {auditResults.summary.matched_entries}
                      </Badge>
                    )}
                  </TabsTrigger>
                </TabsList>

                {/* Summary Tab */}
                <TabsContent value="summary" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Audit Summary</CardTitle>
                      <CardDescription>Overview of audit results</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        <div className="text-center py-8">
                          {auditResults.summary.missing_entries === 0 && 
                           auditResults.summary.mismatched_amounts === 0 ? (
                            <>
                              <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
                              <h3 className="text-lg font-semibold mb-2">All Checks Verified</h3>
                              <p className="text-muted-foreground">
                                All {auditResults.summary.total_checks} check entries have matching GL entries with correct amounts
                              </p>
                            </>
                          ) : (
                            <>
                              <AlertTriangle className="w-16 h-16 text-yellow-500 mx-auto mb-4" />
                              <h3 className="text-lg font-semibold mb-2">Issues Found</h3>
                              <p className="text-muted-foreground mb-4">
                                Found {auditResults.summary.missing_entries + auditResults.summary.mismatched_amounts} issues
                                that require attention
                              </p>
                              <div className="grid gap-2 md:grid-cols-2 max-w-md mx-auto">
                                {auditResults.summary.missing_entries > 0 && (
                                  <div className="flex items-center gap-2 text-red-600">
                                    <XCircle className="w-4 h-4" />
                                    {auditResults.summary.missing_entries} missing GL entries
                                  </div>
                                )}
                                {auditResults.summary.mismatched_amounts > 0 && (
                                  <div className="flex items-center gap-2 text-yellow-600">
                                    <AlertTriangle className="w-4 h-4" />
                                    {auditResults.summary.mismatched_amounts} amount mismatches
                                  </div>
                                )}
                              </div>
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

                {/* Missing Entries Tab */}
                <TabsContent value="missing" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <div className="flex items-center justify-between">
                        <div>
                          <CardTitle className="flex items-center gap-2">
                            <XCircle className="w-5 h-5 text-red-500" />
                            Missing GL Entries ({auditResults.missing_entries.length})
                          </CardTitle>
                          <CardDescription>Check entries without corresponding GL records</CardDescription>
                        </div>
                        <div className="flex items-center gap-2">
                          <div className="flex items-center gap-2">
                            <Search className="w-4 h-4" />
                            <Input
                              placeholder="Search missing entries..."
                              value={missingSearch}
                              onChange={(e) => setMissingSearch(e.target.value)}
                              className="w-64"
                            />
                          </div>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              const newState = !Object.values(missingVisibleColumns).every(v => v)
                              setMissingVisibleColumns({
                                check_id: newState,
                                amount: newState,
                                row_index: newState
                              })
                            }}
                          >
                            <Settings className="w-4 h-4 mr-2" />
                            Columns
                          </Button>
                        </div>
                      </div>
                    </CardHeader>
                    <CardContent>
                      {(() => {
                        const filteredData = filterAndSortData(
                          auditResults.missing_entries,
                          missingSearch,
                          missingSortColumn,
                          missingSortDirection
                        )
                        const paginatedData = filteredData.slice(
                          (missingPage - 1) * itemsPerPage,
                          missingPage * itemsPerPage
                        )
                        
                        return (
                          <>
                            <div className="rounded-md border">
                              <Table>
                                <TableHeader className="sticky top-0 bg-background">
                                  <TableRow>
                                    {missingVisibleColumns.check_id && (
                                      <TableHead 
                                        className="cursor-pointer hover:bg-muted/50"
                                        onClick={() => handleSort('check_id', 'missing')}
                                      >
                                        <div className="flex items-center gap-2">
                                          Check ID/Batch
                                          {getSortIcon('check_id', 'missing')}
                                        </div>
                                      </TableHead>
                                    )}
                                    {missingVisibleColumns.amount && (
                                      <TableHead 
                                        className="text-right cursor-pointer hover:bg-muted/50"
                                        onClick={() => handleSort('amount', 'missing')}
                                      >
                                        <div className="flex items-center justify-end gap-2">
                                          Amount
                                          {getSortIcon('amount', 'missing')}
                                        </div>
                                      </TableHead>
                                    )}
                                    {missingVisibleColumns.row_index && (
                                      <TableHead 
                                        className="cursor-pointer hover:bg-muted/50"
                                        onClick={() => handleSort('row_index', 'missing')}
                                      >
                                        <div className="flex items-center gap-2">
                                          Row
                                          {getSortIcon('row_index', 'missing')}
                                        </div>
                                      </TableHead>
                                    )}
                                    <TableHead>Actions</TableHead>
                                  </TableRow>
                                </TableHeader>
                                <TableBody>
                                  {paginatedData.map((entry, idx) => (
                                    <TableRow key={idx}>
                                      {missingVisibleColumns.check_id && (
                                        <TableCell className="font-mono">{entry.check_id}</TableCell>
                                      )}
                                      {missingVisibleColumns.amount && (
                                        <TableCell className="text-right font-mono">
                                          {formatCurrency(entry.amount)}
                                        </TableCell>
                                      )}
                                      {missingVisibleColumns.row_index && (
                                        <TableCell>{entry.row_index}</TableCell>
                                      )}
                                      <TableCell>
                                        <Button 
                                          variant="ghost" 
                                          size="sm"
                                          onClick={() => showEntryDetails(entry)}
                                        >
                                          View Details
                                        </Button>
                                      </TableCell>
                                    </TableRow>
                                  ))}
                                </TableBody>
                              </Table>
                            </div>
                            {renderPagination(missingPage, setMissingPage, filteredData.length, "missing entries")}
                          </>
                        )
                      })()}
                    </CardContent>
                  </Card>
                </TabsContent>

                {/* Mismatched Amounts Tab */}
                <TabsContent value="mismatched" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2">
                        <AlertTriangle className="w-5 h-5 text-yellow-500" />
                        Mismatched Amounts ({auditResults.mismatched_amounts.length})
                      </CardTitle>
                      <CardDescription>Entries with amount differences between checks and GL</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <p className="text-sm text-muted-foreground">Mismatched amounts feature coming soon...</p>
                    </CardContent>
                  </Card>
                </TabsContent>

                {/* Matched Entries Tab */}
                <TabsContent value="matched" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2">
                        <CheckCircle className="w-5 h-5 text-green-500" />
                        Matched Entries ({auditResults.summary.matched_entries})
                      </CardTitle>
                      <CardDescription>Successfully verified check entries</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8">
                        <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
                        <h3 className="text-lg font-semibold mb-2">All Verified</h3>
                        <p className="text-muted-foreground">
                          {auditResults.summary.matched_entries} check entries were successfully verified
                        </p>
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
              <p className="text-muted-foreground mb-4">
                Click "Run Audit" to compare check entries with the General Ledger
              </p>
              <p className="text-sm text-muted-foreground">
                This will analyze the CBATCH field in checks.dbf against GLMASTER.dbf
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Bank Reconciliation Audit Section */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Bank Reconciliation Audit</CardTitle>
              <CardDescription>
                Verify that Reconciliation Balance - Outstanding Checks = GL Balance
              </CardDescription>
            </div>
            <div className="flex gap-2">
              {bankRecAuditResults && (
                <Button 
                  variant="outline" 
                  size="sm"
                  onClick={() => {
                    // Export bank reconciliation results to CSV
                    const csvData = bankRecAuditResults.discrepancies.map(item => ({
                      account_number: item.account_number,
                      account_name: item.account_name,
                      issue_type: item.issue_type,
                      reconciliation_balance: item.reconciliation_balance || 'N/A',
                      reconciliation_date: item.reconciliation_date || 'N/A',
                      gl_balance: item.gl_balance || 'N/A',
                      outstanding_checks: item.outstanding_checks || 'N/A',
                      expected_gl_balance: item.expected_gl_balance || 'N/A',
                      difference: item.difference || 'N/A',
                      description: item.description
                    }))
                    
                    const csv = [
                      Object.keys(csvData[0] || {}).join(','),
                      ...csvData.map(row => Object.values(row).join(','))
                    ].join('\n')
                    
                    const blob = new Blob([csv], { type: 'text/csv' })
                    const url = window.URL.createObjectURL(blob)
                    const a = document.createElement('a')
                    a.href = url
                    a.download = `bank-reconciliation-audit-${companyName}-${new Date().toISOString().split('T')[0]}.csv`
                    a.click()
                    window.URL.revokeObjectURL(url)
                  }}
                >
                  <Download className="w-4 h-4 mr-2" />
                  Export CSV
                </Button>
              )}
              <Button 
                onClick={runBankReconciliationAudit}
                disabled={bankRecLoading}
              >
                {bankRecLoading ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Running Audit...
                  </>
                ) : (
                  <>
                    <FileSearch className="w-4 h-4 mr-2" />
                    Run Bank Reconciliation Audit
                  </>
                )}
              </Button>
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

          {bankRecAuditResults && (
            <div className="space-y-6">
              {/* Summary Cards */}
              <div className="grid gap-4 md:grid-cols-4">
                <Card>
                  <CardContent className="p-4">
                    <div className="text-2xl font-bold">{bankRecAuditResults.accounts_audited}</div>
                    <p className="text-sm text-muted-foreground">Bank Accounts Audited</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardContent className="p-4">
                    <div className="text-2xl font-bold text-green-600">
                      {bankRecAuditResults.accounts_audited - bankRecAuditResults.total_discrepancies}
                    </div>
                    <p className="text-sm text-muted-foreground">Balanced</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardContent className="p-4">
                    <div className="text-2xl font-bold text-red-600">
                      {bankRecAuditResults.total_discrepancies}
                    </div>
                    <p className="text-sm text-muted-foreground">Discrepancies</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardContent className="p-4">
                    <div className="text-2xl font-bold text-blue-600">
                      {bankRecAuditResults.discrepancies.filter(d => d.issue_type === 'balance_mismatch').length}
                    </div>
                    <p className="text-sm text-muted-foreground">Balance Mismatches</p>
                  </CardContent>
                </Card>
              </div>

              {/* Results Table */}
              {bankRecAuditResults.discrepancies.length > 0 ? (
                <Card>
                  <CardHeader>
                    <CardTitle>Discrepancies Found</CardTitle>
                    <CardDescription>Accounts with reconciliation balance issues</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="rounded-md border">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Account</TableHead>
                            <TableHead>Issue Type</TableHead>
                            <TableHead className="text-right">Reconciliation Balance</TableHead>
                            <TableHead className="text-center">Reconciliation Date</TableHead>
                            <TableHead className="text-right">GL Balance</TableHead>
                            <TableHead className="text-right">Outstanding Checks</TableHead>
                            <TableHead className="text-right">Expected GL Balance</TableHead>
                            <TableHead className="text-right">Difference</TableHead>
                            <TableHead>Description</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {bankRecAuditResults.discrepancies.map((discrepancy, index) => (
                            <TableRow key={index}>
                              <TableCell className="font-medium">
                                <div>
                                  <div>{discrepancy.account_number}</div>
                                  <div className="text-sm text-muted-foreground">
                                    {discrepancy.account_name}
                                  </div>
                                </div>
                              </TableCell>
                              <TableCell>
                                <Badge variant={
                                  discrepancy.issue_type === 'balance_mismatch' ? 'destructive' :
                                  discrepancy.issue_type === 'no_reconciliation_data' ? 'secondary' :
                                  'outline'
                                }>
                                  {discrepancy.issue_type.replace('_', ' ')}
                                </Badge>
                              </TableCell>
                              <TableCell className="text-right font-mono">
                                {discrepancy.reconciliation_balance != null ? 
                                  formatCurrency(discrepancy.reconciliation_balance) : 'N/A'}
                              </TableCell>
                              <TableCell className="text-center font-mono text-sm">
                                {discrepancy.reconciliation_date || 'N/A'}
                              </TableCell>
                              <TableCell className="text-right font-mono">
                                {discrepancy.gl_balance != null ? 
                                  formatCurrency(discrepancy.gl_balance) : 'N/A'}
                              </TableCell>
                              <TableCell className="text-right font-mono">
                                {discrepancy.outstanding_checks != null ? 
                                  formatCurrency(discrepancy.outstanding_checks) : 'N/A'}
                              </TableCell>
                              <TableCell className="text-right font-mono">
                                {discrepancy.expected_gl_balance != null ? 
                                  formatCurrency(discrepancy.expected_gl_balance) : 'N/A'}
                              </TableCell>
                              <TableCell className={`text-right font-mono ${
                                discrepancy.difference > 0 ? 'text-green-600' : 
                                discrepancy.difference < 0 ? 'text-red-600' : ''
                              }`}>
                                {discrepancy.difference != null ? 
                                  formatCurrency(discrepancy.difference) : 'N/A'}
                              </TableCell>
                              <TableCell className="text-sm">
                                {discrepancy.description}
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                  </CardContent>
                </Card>
              ) : (
                <Card>
                  <CardContent className="p-6">
                    <div className="text-center text-green-600">
                      <CheckCircle className="w-12 h-12 mx-auto mb-4" />
                      <h3 className="text-lg font-semibold mb-2">All Bank Accounts Balanced!</h3>
                      <p className="text-muted-foreground">
                        All reconciliation balances match the calculated bank balances (GL + Outstanding Checks)
                      </p>
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Audit Summary */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <CheckCircle className="w-5 h-5 text-green-600" />
                    Audit Summary
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2 text-sm">
                    <p>Audit completed on: {new Date(bankRecAuditResults.audit_timestamp).toLocaleString()}</p>
                    <p>Audited by: {bankRecAuditResults.audited_by}</p>
                    <p className="text-muted-foreground mt-4">
                      This audit verifies that the reconciliation balance from CHECKREC.dbf 
                      minus outstanding checks equals the GL balance for each account.
                    </p>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          {!bankRecAuditResults && !bankRecError && !bankRecLoading && (
            <div className="text-center py-8">
              <FileSearch className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
              <h3 className="text-lg font-semibold mb-2">Bank Reconciliation Audit</h3>
              <p className="text-muted-foreground mb-4">
                Click "Run Bank Reconciliation Audit" to verify reconciliation balance minus outstanding checks equals GL balance
              </p>
              <p className="text-sm text-muted-foreground">
                This will analyze CHECKREC.dbf against cached balance data
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Details Dialog */}
      <Dialog open={showDetails} onOpenChange={setShowDetails}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Entry Details</DialogTitle>
            <DialogDescription>
              Full details for {selectedEntry?.check_id}
            </DialogDescription>
          </DialogHeader>
          {selectedEntry && (
            <div className="space-y-4">
              <div>
                <h4 className="font-medium mb-2">Check Information</h4>
                <div className="bg-gray-50 p-4 rounded-lg space-y-2">
                  <div className="flex justify-between">
                    <span className="text-sm text-muted-foreground">Check ID:</span>
                    <span className="font-mono">{selectedEntry.check_id}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-muted-foreground">Amount:</span>
                    <span className="font-mono">{formatCurrency(selectedEntry.amount || selectedEntry.check_amount)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-sm text-muted-foreground">Row Index:</span>
                    <span className="font-mono">{selectedEntry.row_index}</span>
                  </div>
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
                      <div className="flex justify-between mb-2">
                        <span className="text-sm text-muted-foreground">GL Amount:</span>
                        <span className="font-mono">{formatCurrency(glEntry.amount)}</span>
                      </div>
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