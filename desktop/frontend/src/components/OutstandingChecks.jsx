import React, { useState, useEffect, useMemo } from 'react'
import { GetOutstandingChecks, GetBankAccounts, UpdateDBFRecord, GetDBFTableData } from '../../wailsjs/go/main/App'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Badge } from './ui/badge'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './ui/table'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from './ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { 
  AlertTriangle,
  RefreshCw,
  Calendar,
  DollarSign,
  Search,
  Filter,
  ChevronLeft,
  ChevronRight,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Edit,
  Eye,
  X,
  Check
} from 'lucide-react'

const OutstandingChecks = ({ companyName, currentUser }) => {
  console.log('OutstandingChecks component rendered with:', { companyName, currentUser })
  
  // State Management
  const [outstandingChecks, setOutstandingChecks] = useState([])
  const [bankAccounts, setBankAccounts] = useState([])
  const [selectedAccount, setSelectedAccount] = useState('all')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [totalAmount, setTotalAmount] = useState(0)
  
  // Pagination
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(25)
  
  // Sorting
  const [sortField, setSortField] = useState('date')
  const [sortDirection, setSortDirection] = useState('desc')
  
  // Search & Filter
  const [searchTerm, setSearchTerm] = useState('')
  const [showStaleOnly, setShowStaleOnly] = useState(false)
  
  // Detail Modal
  const [selectedCheck, setSelectedCheck] = useState(null)
  const [editMode, setEditMode] = useState(false)
  const [editedCheck, setEditedCheck] = useState({})
  
  // Check if user can edit
  const canEdit = currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')

  // Calculate days outstanding
  const calculateDaysOutstanding = (checkDate) => {
    if (!checkDate) return 'N/A'
    
    try {
      const today = new Date()
      const checkDateTime = new Date(checkDate)
      if (isNaN(checkDateTime.getTime())) return 'N/A'
      
      const diffTime = today - checkDateTime
      const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
      return diffDays
    } catch (err) {
      return 'N/A'
    }
  }

  // Format currency
  const formatCurrency = (amount) => {
    if (typeof amount !== 'number') return '$0.00'
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  // Format date
  const formatDate = (dateStr) => {
    if (!dateStr) return 'N/A'
    try {
      const date = new Date(dateStr)
      if (isNaN(date.getTime())) return dateStr
      return date.toLocaleDateString()
    } catch (err) {
      return dateStr
    }
  }

  // Load bank accounts - robust version with fallback like BankingSection
  const loadBankAccounts = async () => {
    console.log('OutstandingChecks: loadBankAccounts called with companyName:', companyName)
    console.log('OutstandingChecks: currentUser object:', currentUser)
    
    if (!companyName) {
      console.log('OutstandingChecks: No company name provided, companyName is:', companyName)
      console.log('OutstandingChecks: currentUser.company_name is:', currentUser?.company_name)
      return
    }
    
    try {
      let bankAccountsData = []
      
      // Method 1: Try GetBankAccounts first
      if (typeof GetBankAccounts === 'function') {
        console.log('OutstandingChecks: GetBankAccounts function is available, calling it...')
        try {
          bankAccountsData = await GetBankAccounts(companyName)
          console.log('OutstandingChecks: GetBankAccounts response:', bankAccountsData)
          console.log('OutstandingChecks: GetBankAccounts response type:', typeof bankAccountsData)
          console.log('OutstandingChecks: GetBankAccounts response length:', bankAccountsData?.length)
          
          if (!bankAccountsData || !Array.isArray(bankAccountsData)) {
            console.log('OutstandingChecks: GetBankAccounts returned invalid data, using fallback')
            throw new Error('GetBankAccounts returned invalid data: ' + typeof bankAccountsData)
          }
        } catch (getBankAccountsErr) {
          console.error('OutstandingChecks: GetBankAccounts call failed:', getBankAccountsErr)
          throw getBankAccountsErr // Fall through to fallback
        }
      } else {
        console.log('OutstandingChecks: GetBankAccounts function not available')
        throw new Error('GetBankAccounts function not available')
      }
      
      // Success with primary method
      setBankAccounts(bankAccountsData)
      console.log('OutstandingChecks: Successfully set', bankAccountsData.length, 'bank accounts via GetBankAccounts')
      
    } catch (err) {
      console.log('OutstandingChecks: Primary method failed, trying fallback...')
      // Fallback: Try to read COA.dbf directly using GetDBFTableData
      try {
        console.log('OutstandingChecks: Using GetDBFTableData fallback...')
        const coaData = await GetDBFTableData(companyName, 'COA.dbf')
        console.log('OutstandingChecks: COA.dbf data loaded:', coaData)
        
        if (coaData && coaData.rows) {
          const bankAccounts = coaData.rows
            .filter((row, index) => {
              // Check if LBANKACCT is true (column 6 based on COA structure)
              const bankFlag = row[6]
              if (index < 5) { // Only log first 5 rows to avoid spam
                console.log('OutstandingChecks - Row', index, 'Account:', row[0], 'LBANKACCT flag:', bankFlag, 'type:', typeof bankFlag)
              }
              return bankFlag === true || bankFlag === 'T' || bankFlag === '.T.' || bankFlag === 'true'
            })
            .map(row => ({
              account_number: row[0] || '',     // Cacctno
              account_name: row[2] || '',       // Cacctdesc (Account description)
              account_type: row[1] || 'Checking', // Caccttype
              balance: 0,                       // Balance not in COA
              description: row[2] || '',        // Cacctdesc
              is_bank_account: true
            }))
          
          console.log('OutstandingChecks: Filtered bank accounts via fallback:', bankAccounts)
          setBankAccounts(bankAccounts)
          console.log('OutstandingChecks: Successfully set', bankAccounts.length, 'bank accounts via fallback')
        } else {
          console.error('OutstandingChecks: No data in COA.dbf')
          setBankAccounts([])
        }
      } catch (fallbackErr) {
        console.error('OutstandingChecks: Fallback method also failed:', fallbackErr)
        setBankAccounts([])
      }
    }
  }

  // Load outstanding checks
  const loadOutstandingChecks = async () => {
    if (!companyName) return
    
    setLoading(true)
    setError('')
    
    try {
      // Pass selected account filter to backend
      const accountFilter = selectedAccount === 'all' ? '' : selectedAccount
      console.log('Loading outstanding checks for account:', accountFilter || 'all')
      
      const result = await GetOutstandingChecks(companyName, accountFilter)
      
      if (result.status === 'error') {
        setError(result.error || 'Failed to load outstanding checks')
        setOutstandingChecks([])
      } else {
        const checks = result.checks || []
        setOutstandingChecks(checks)
        
        // Calculate total amount
        const total = checks.reduce((sum, check) => sum + (check.amount || 0), 0)
        setTotalAmount(total)
        
        console.log(`Loaded ${checks.length} outstanding checks, total: $${total.toFixed(2)}`)
      }
    } catch (err) {
      console.error('Failed to load outstanding checks:', err)
      setError(err.message || 'Failed to load outstanding checks')
      setOutstandingChecks([])
    } finally {
      setLoading(false)
    }
  }

  // Load data when component mounts or dependencies change
  useEffect(() => {
    loadBankAccounts()
  }, [companyName])

  useEffect(() => {
    loadOutstandingChecks()
  }, [companyName, selectedAccount])

  // Filter and sort checks
  const processedChecks = useMemo(() => {
    let filtered = [...outstandingChecks]
    
    // Apply search filter
    if (searchTerm) {
      const term = searchTerm.toLowerCase()
      filtered = filtered.filter(check => 
        (check.checkNumber && check.checkNumber.toLowerCase().includes(term)) ||
        (check.payee && check.payee.toLowerCase().includes(term)) ||
        (check.account && check.account.toLowerCase().includes(term)) ||
        (check.amount && check.amount.toString().includes(term))
      )
    }
    
    // Apply stale filter (>90 days)
    if (showStaleOnly) {
      filtered = filtered.filter(check => {
        const days = calculateDaysOutstanding(check.date)
        return days !== 'N/A' && days > 90
      })
    }
    
    // Sort
    filtered.sort((a, b) => {
      let aVal, bVal
      
      switch (sortField) {
        case 'checkNumber':
          aVal = parseInt(a.checkNumber) || 0
          bVal = parseInt(b.checkNumber) || 0
          break
        case 'date':
          aVal = new Date(a.date || 0)
          bVal = new Date(b.date || 0)
          break
        case 'amount':
          aVal = a.amount || 0
          bVal = b.amount || 0
          break
        case 'payee':
          aVal = (a.payee || '').toLowerCase()
          bVal = (b.payee || '').toLowerCase()
          break
        case 'daysOutstanding':
          aVal = calculateDaysOutstanding(a.date)
          bVal = calculateDaysOutstanding(b.date)
          if (aVal === 'N/A') aVal = -1
          if (bVal === 'N/A') bVal = -1
          break
        default:
          return 0
      }
      
      if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1
      if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1
      return 0
    })
    
    return filtered
  }, [outstandingChecks, searchTerm, showStaleOnly, sortField, sortDirection])

  // Pagination
  const paginatedChecks = useMemo(() => {
    const startIndex = (currentPage - 1) * pageSize
    return processedChecks.slice(startIndex, startIndex + pageSize)
  }, [processedChecks, currentPage, pageSize])

  const totalPages = Math.ceil(processedChecks.length / pageSize)

  // Handle sort
  const handleSort = (field) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortDirection('asc')
    }
  }

  // Get sort icon
  const getSortIcon = (field) => {
    if (sortField !== field) return <ArrowUpDown className="h-4 w-4" />
    return sortDirection === 'asc' ? 
      <ArrowUp className="h-4 w-4" /> : 
      <ArrowDown className="h-4 w-4" />
  }

  // Handle check selection
  const handleCheckSelect = (check) => {
    setSelectedCheck(check)
    setEditedCheck({ ...check })
    setEditMode(false)
  }

  // Handle save edit
  const handleSaveEdit = async () => {
    if (!selectedCheck || !editedCheck) return
    
    try {
      // Call backend to update the check record
      await UpdateDBFRecord(companyName, 'checks.dbf', selectedCheck._rowIndex, editedCheck)
      
      // Reload checks to reflect changes
      await loadOutstandingChecks()
      
      // Close modal
      setSelectedCheck(null)
      setEditMode(false)
    } catch (err) {
      console.error('Failed to save check:', err)
      alert('Failed to save changes: ' + err.message)
    }
  }

  // Get badge variant for days outstanding
  const getDaysOutstandingBadge = (days) => {
    if (days === 'N/A') return { variant: 'secondary', text: 'N/A' }
    if (days <= 30) return { variant: 'default', text: `${days} days` }
    if (days <= 60) return { variant: 'secondary', text: `${days} days` }
    if (days <= 90) return { variant: 'destructive', text: `${days} days` }
    return { variant: 'destructive', text: `${days} days (STALE)` }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Outstanding Checks</h2>
          <p className="text-muted-foreground">
            Checks that have not been cleared by the bank
          </p>
        </div>
        <Button onClick={loadOutstandingChecks} disabled={loading}>
          <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {/* Filters Row */}
      <div className="flex gap-4 items-end">
        {/* Account Filter */}
        <div className="flex-1 max-w-xs">
          <Label htmlFor="account-filter">Bank Account</Label>
          {/* Temporary debugging with basic HTML select */}
          <div>
            <select 
              id="account-filter" 
              value={selectedAccount} 
              onChange={(e) => {
                console.log('OutstandingChecks: Account selected via HTML select:', e.target.value)
                setSelectedAccount(e.target.value)
              }}
              className="w-full p-2 border rounded-md"
            >
              <option value="all">All Accounts (Debug: {bankAccounts.length} loaded)</option>
              {bankAccounts.map((account, idx) => {
                console.log(`OutstandingChecks: Rendering HTML option ${idx}:`, account)
                const accountNumber = account.account_number || account.accountNumber || ''
                const accountName = account.account_name || account.accountName || ''
                
                return (
                  <option key={`${idx}-${accountNumber}`} value={accountNumber}>
                    {accountNumber} - {accountName}
                  </option>
                )
              })}
            </select>
            <div className="text-xs text-gray-500 mt-1">
              Debug Info: {bankAccounts.length} accounts loaded, selected: '{selectedAccount}'
            </div>
          </div>
        </div>

        {/* Search */}
        <div className="flex-1 max-w-sm">
          <Label htmlFor="search">Search</Label>
          <div className="relative">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              id="search"
              placeholder="Search checks..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-8"
            />
          </div>
        </div>

        {/* Page Size */}
        <div className="w-32">
          <Label htmlFor="page-size">Page Size</Label>
          <Select value={pageSize.toString()} onValueChange={(v) => setPageSize(parseInt(v))}>
            <SelectTrigger id="page-size">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="10">10</SelectItem>
              <SelectItem value="25">25</SelectItem>
              <SelectItem value="50">50</SelectItem>
              <SelectItem value="100">100</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Stale Filter */}
        <Button
          variant={showStaleOnly ? "default" : "outline"}
          onClick={() => setShowStaleOnly(!showStaleOnly)}
        >
          <Filter className="mr-2 h-4 w-4" />
          Stale Only
        </Button>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Outstanding Checks
            </CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{processedChecks.length}</div>
            <p className="text-xs text-muted-foreground">
              {selectedAccount === 'all' ? 'All accounts' : selectedAccount}
            </p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Total Amount
            </CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatCurrency(processedChecks.reduce((sum, check) => sum + (check.amount || 0), 0))}
            </div>
            <p className="text-xs text-muted-foreground">
              Outstanding check amount
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Stale Checks
            </CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {processedChecks.filter(check => {
                const days = calculateDaysOutstanding(check.date)
                return days !== 'N/A' && days > 90
              }).length}
            </div>
            <p className="text-xs text-muted-foreground">
              Over 90 days old
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Error Display */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2 text-destructive">
              <AlertTriangle className="h-4 w-4" />
              <span>{error}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Outstanding Checks Table */}
      <Card>
        <CardHeader>
          <CardTitle>Outstanding Checks List</CardTitle>
          <CardDescription>
            {loading ? 'Loading checks...' : 
             processedChecks.length === 0 ? 'No outstanding checks found' :
             `Showing ${paginatedChecks.length} of ${processedChecks.length} checks (Page ${currentPage} of ${totalPages})`}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <RefreshCw className="h-6 w-6 animate-spin" />
              <span className="ml-2">Loading outstanding checks...</span>
            </div>
          ) : processedChecks.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              {error ? 'Unable to load checks' : 'No outstanding checks found'}
            </div>
          ) : (
            <>
              <div className="overflow-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleSort('checkNumber')}
                          className="h-8 px-2"
                        >
                          Check #
                          {getSortIcon('checkNumber')}
                        </Button>
                      </TableHead>
                      <TableHead>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleSort('date')}
                          className="h-8 px-2"
                        >
                          Date
                          {getSortIcon('date')}
                        </Button>
                      </TableHead>
                      <TableHead>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleSort('payee')}
                          className="h-8 px-2"
                        >
                          Payee
                          {getSortIcon('payee')}
                        </Button>
                      </TableHead>
                      <TableHead className="text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleSort('amount')}
                          className="h-8 px-2"
                        >
                          Amount
                          {getSortIcon('amount')}
                        </Button>
                      </TableHead>
                      <TableHead>Account</TableHead>
                      <TableHead className="text-center">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleSort('daysOutstanding')}
                          className="h-8 px-2"
                        >
                          Days Outstanding
                          {getSortIcon('daysOutstanding')}
                        </Button>
                      </TableHead>
                      <TableHead className="text-center">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedChecks.map((check, index) => {
                      const daysOut = calculateDaysOutstanding(check.date)
                      const daysBadge = getDaysOutstandingBadge(daysOut)
                      
                      return (
                        <TableRow 
                          key={index} 
                          className="cursor-pointer hover:bg-muted/50"
                          onClick={() => handleCheckSelect(check)}
                        >
                          <TableCell className="font-medium">
                            {check.checkNumber || 'N/A'}
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center">
                              <Calendar className="mr-2 h-4 w-4 text-muted-foreground" />
                              {formatDate(check.date)}
                            </div>
                          </TableCell>
                          <TableCell>
                            {check.payee || 'N/A'}
                          </TableCell>
                          <TableCell className="text-right font-medium">
                            {formatCurrency(check.amount)}
                          </TableCell>
                          <TableCell>
                            {check.account || 'N/A'}
                          </TableCell>
                          <TableCell className="text-center">
                            <Badge variant={daysBadge.variant}>
                              {daysBadge.text}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-center">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={(e) => {
                                e.stopPropagation()
                                handleCheckSelect(check)
                              }}
                            >
                              <Eye className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              </div>

              {/* Pagination Controls */}
              {totalPages > 1 && (
                <div className="flex items-center justify-between mt-4">
                  <div className="text-sm text-muted-foreground">
                    Showing {((currentPage - 1) * pageSize) + 1} to {Math.min(currentPage * pageSize, processedChecks.length)} of {processedChecks.length} checks
                  </div>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                      disabled={currentPage === 1}
                    >
                      <ChevronLeft className="h-4 w-4" />
                      Previous
                    </Button>
                    
                    {/* Page numbers */}
                    <div className="flex gap-1">
                      {[...Array(Math.min(5, totalPages))].map((_, i) => {
                        let pageNum
                        if (totalPages <= 5) {
                          pageNum = i + 1
                        } else if (currentPage <= 3) {
                          pageNum = i + 1
                        } else if (currentPage >= totalPages - 2) {
                          pageNum = totalPages - 4 + i
                        } else {
                          pageNum = currentPage - 2 + i
                        }
                        
                        return (
                          <Button
                            key={i}
                            variant={pageNum === currentPage ? "default" : "outline"}
                            size="sm"
                            onClick={() => setCurrentPage(pageNum)}
                            className="w-8"
                          >
                            {pageNum}
                          </Button>
                        )
                      })}
                    </div>
                    
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                      disabled={currentPage === totalPages}
                    >
                      Next
                      <ChevronRight className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Check Detail/Edit Modal */}
      {selectedCheck && (
        <Dialog open={!!selectedCheck} onOpenChange={() => setSelectedCheck(null)}>
          <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle className="flex items-center justify-between">
                <span>Check #{selectedCheck.checkNumber}</span>
                {canEdit && !editMode && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setEditMode(true)}
                  >
                    <Edit className="h-4 w-4 mr-2" />
                    Edit
                  </Button>
                )}
              </DialogTitle>
              <DialogDescription>
                {editMode ? 'Edit check details' : 'View check details'}
              </DialogDescription>
            </DialogHeader>

            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="check-number">Check Number</Label>
                  <Input
                    id="check-number"
                    value={editedCheck.checkNumber || ''}
                    onChange={(e) => setEditedCheck({...editedCheck, checkNumber: e.target.value})}
                    disabled={!editMode}
                  />
                </div>
                <div>
                  <Label htmlFor="check-date">Date</Label>
                  <Input
                    id="check-date"
                    type="date"
                    value={editedCheck.date || ''}
                    onChange={(e) => setEditedCheck({...editedCheck, date: e.target.value})}
                    disabled={!editMode}
                  />
                </div>
              </div>

              <div>
                <Label htmlFor="payee">Payee</Label>
                <Input
                  id="payee"
                  value={editedCheck.payee || ''}
                  onChange={(e) => setEditedCheck({...editedCheck, payee: e.target.value})}
                  disabled={!editMode}
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="amount">Amount</Label>
                  <Input
                    id="amount"
                    type="number"
                    step="0.01"
                    value={editedCheck.amount || ''}
                    onChange={(e) => setEditedCheck({...editedCheck, amount: parseFloat(e.target.value)})}
                    disabled={!editMode}
                  />
                </div>
                <div>
                  <Label htmlFor="account">Account</Label>
                  <Input
                    id="account"
                    value={editedCheck.account || ''}
                    onChange={(e) => setEditedCheck({...editedCheck, account: e.target.value})}
                    disabled={!editMode}
                  />
                </div>
              </div>

              {/* Additional Info */}
              <div className="border-t pt-4">
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div>
                    <span className="font-medium">Days Outstanding:</span>
                    <span className="ml-2">{calculateDaysOutstanding(selectedCheck.date)}</span>
                  </div>
                  <div>
                    <span className="font-medium">Status:</span>
                    <Badge className="ml-2" variant={calculateDaysOutstanding(selectedCheck.date) > 90 ? "destructive" : "default"}>
                      {calculateDaysOutstanding(selectedCheck.date) > 90 ? "Stale" : "Outstanding"}
                    </Badge>
                  </div>
                </div>
              </div>
            </div>

            <DialogFooter>
              {editMode ? (
                <>
                  <Button variant="outline" onClick={() => {
                    setEditedCheck({...selectedCheck})
                    setEditMode(false)
                  }}>
                    <X className="h-4 w-4 mr-2" />
                    Cancel
                  </Button>
                  <Button onClick={handleSaveEdit}>
                    <Check className="h-4 w-4 mr-2" />
                    Save Changes
                  </Button>
                </>
              ) : (
                <Button variant="outline" onClick={() => setSelectedCheck(null)}>
                  Close
                </Button>
              )}
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default OutstandingChecks