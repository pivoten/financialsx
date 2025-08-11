
import React, { useState, useEffect, useMemo } from 'react'
import { GetOutstandingChecks, GetBankAccounts, UpdateDBFRecord, GetDBFTableData } from '../../wailsjs/go/main/App'
import { getCompanyDataPath } from '../utils/companyPath'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Badge } from './ui/badge'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './ui/table'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from './ui/dialog'
import { Select } from './ui/select'
import { AlertTriangle, RefreshCw, Calendar, DollarSign, Search, Filter, ChevronLeft, ChevronRight, ArrowUpDown, ArrowUp, ArrowDown, Edit, Eye, X, Check as CheckIcon } from 'lucide-react'
import type { User, Check, BankAccount, ChangeEvent, MouseEvent } from '../types'

interface OutstandingChecksProps {
  companyName: string
  currentUser: User
}

interface OutstandingCheck {
  checkNumber: string
  date: string
  payee: string
  amount: number
  account: string
  _rowIndex?: number
}

interface BadgeInfo {
  variant: 'default' | 'secondary' | 'destructive'
  text: string
}

const OutstandingChecks = ({ companyName, currentUser }: OutstandingChecksProps) => {
  const [outstandingChecks, setOutstandingChecks] = useState<OutstandingCheck[]>([])
  const [bankAccounts, setBankAccounts] = useState<BankAccount[]>([])
  const [selectedAccount, setSelectedAccount] = useState<string>('all')
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string>('')
  const [totalAmount, setTotalAmount] = useState<number>(0)
  const [currentPage, setCurrentPage] = useState<number>(1)
  const [pageSize, setPageSize] = useState<number>(25)
  const [sortField, setSortField] = useState<string>('date')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc')
  const [searchTerm, setSearchTerm] = useState<string>('')
  const [showStaleOnly, setShowStaleOnly] = useState<boolean>(false)
  const [selectedCheck, setSelectedCheck] = useState<OutstandingCheck | null>(null)
  const [editMode, setEditMode] = useState<boolean>(false)
  const [editedCheck, setEditedCheck] = useState<Partial<OutstandingCheck>>({})

  const canEdit = currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')

  const calculateDaysOutstanding = (checkDate: string): number | 'N/A' => {
    if (!checkDate) return 'N/A'
    try {
      const today = new Date().getTime()
      const checkDateTime = new Date(checkDate).getTime()
      if (isNaN(checkDateTime)) return 'N/A'
      const diffTime = today - checkDateTime
      const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
      return diffDays
    } catch { return 'N/A' }
  }

  const formatCurrency = (amount: number): string => {
    if (typeof amount !== 'number') return '$0.00'
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount)
  }

  const formatDate = (dateStr: string): string => {
    if (!dateStr) return 'N/A'
    try { const date = new Date(dateStr); if (isNaN(date.getTime())) return dateStr; return date.toLocaleDateString() } catch { return dateStr }
  }

  const loadBankAccounts = async () => {
    const companyName = localStorage.getItem('company_name')
    if (!companyName) return
    try {
      let bankAccountsData = []
      if (typeof GetBankAccounts === 'function') {
        try {
          bankAccountsData = await GetBankAccounts(companyName)
          if (!bankAccountsData || !Array.isArray(bankAccountsData)) throw new Error('GetBankAccounts returned invalid data: ' + typeof bankAccountsData)
        } catch (err) {
          throw err
        }
      } else {
        throw new Error('GetBankAccounts function not available')
      }
  setBankAccounts(bankAccountsData as BankAccount[])
    } catch (err) {
      try {
        const coaData = await GetDBFTableData(companyName, 'COA.dbf')
        if (coaData && coaData.rows) {
          const bankAccounts = (coaData.rows as any[])
            .filter((row: any[]) => {
              const bankFlag = row[6]
              return bankFlag === true || bankFlag === 'T' || bankFlag === '.T.' || bankFlag === 'true'
            })
            .map((row: any[]) => ({ account_number: String(row[0] || ''), account_name: String(row[2] || ''), account_type: Number(row[1] ?? 0), balance: 0, description: String(row[2] || ''), is_bank_account: true })) as BankAccount[]
          setBankAccounts(bankAccounts as BankAccount[])
        } else { setBankAccounts([]) }
      } catch (fallbackErr) {
        setBankAccounts([])
      }
    }
  }

  const loadOutstandingChecks = async () => {
    const companyName = localStorage.getItem('company_name')
    if (!companyName) { setError('No company selected. Please select a company first.'); return }
    setLoading(true); setError('')
    try {
      const accountFilter = selectedAccount === 'all' ? '' : selectedAccount
      const result = await GetOutstandingChecks(companyName, accountFilter)
      if (result.status === 'error') { setError(result.error || 'Failed to load outstanding checks'); setOutstandingChecks([]) }
      else {
        const checks = result.checks || []
        setOutstandingChecks(checks)
  const total = checks.reduce((sum: number, check: OutstandingCheck) => sum + (check.amount || 0), 0)
        setTotalAmount(total)
      }
    } catch (err) {
      setError(err.message || 'Failed to load outstanding checks'); setOutstandingChecks([])
    } finally { setLoading(false) }
  }

  useEffect(() => { const companyName = localStorage.getItem('company_name'); if (companyName) loadBankAccounts() }, [companyName])
  useEffect(() => { const companyName = localStorage.getItem('company_name'); if (companyName) loadOutstandingChecks() }, [companyName, selectedAccount])

  const processedChecks = useMemo(() => {
    let filtered = [...outstandingChecks]
    if (searchTerm) {
      const term = searchTerm.toLowerCase()
      filtered = filtered.filter(check => (check.checkNumber && check.checkNumber.toLowerCase().includes(term)) || (check.payee && check.payee.toLowerCase().includes(term)) || (check.account && check.account.toLowerCase().includes(term)) || (check.amount && check.amount.toString().includes(term)))
    }
    if (showStaleOnly) {
      filtered = filtered.filter(check => { const days = calculateDaysOutstanding(check.date); return days !== 'N/A' && days > 90 })
    }
    filtered.sort((a, b) => {
      let aVal, bVal
      switch (sortField) {
        case 'checkNumber': aVal = parseInt(a.checkNumber) || 0; bVal = parseInt(b.checkNumber) || 0; break
        case 'date': aVal = new Date(a.date || 0); bVal = new Date(b.date || 0); break
        case 'amount': aVal = a.amount || 0; bVal = b.amount || 0; break
        case 'payee': aVal = (a.payee || '').toLowerCase(); bVal = (b.payee || '').toLowerCase(); break
        case 'daysOutstanding': aVal = calculateDaysOutstanding(a.date); bVal = calculateDaysOutstanding(b.date); if (aVal === 'N/A') aVal = -1; if (bVal === 'N/A') bVal = -1; break
        default: return 0
      }
      if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1
      if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1
      return 0
    })
    return filtered
  }, [outstandingChecks, searchTerm, showStaleOnly, sortField, sortDirection])

  const paginatedChecks = useMemo(() => {
    const startIndex = (currentPage - 1) * pageSize
    return processedChecks.slice(startIndex, startIndex + pageSize)
  }, [processedChecks, currentPage, pageSize])

  const totalPages = Math.ceil(processedChecks.length / pageSize)

  const handleSort = (field: string) => {
    if (sortField === field) setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    else { setSortField(field); setSortDirection('asc') }
  }

  const getSortIcon = (field: string) => {
    if (sortField !== field) return <ArrowUpDown className="h-4 w-4" />
    return sortDirection === 'asc' ? <ArrowUp className="h-4 w-4" /> : <ArrowDown className="h-4 w-4" />
  }

  const handleCheckSelect = (check: OutstandingCheck) => { setSelectedCheck(check); setEditedCheck({ ...check }); setEditMode(false) }

  const handleSaveEdit = async () => {
    if (!selectedCheck || !editedCheck) return
    try {
      await UpdateDBFRecord(companyName, 'checks.dbf', selectedCheck._rowIndex || 0, 0, JSON.stringify(editedCheck))
      await loadOutstandingChecks()
      setSelectedCheck(null); setEditMode(false)
    } catch (err) {
      alert('Failed to save changes: ' + err.message)
    }
  }

  const getDaysOutstandingBadge = (days: number | string): BadgeInfo => {
    if (days === 'N/A') return { variant: 'secondary', text: 'N/A' }
    if (Number(days) <= 30) return { variant: 'default', text: `${days} days` }
    if (Number(days) <= 60) return { variant: 'secondary', text: `${days} days` }
    if (Number(days) <= 90) return { variant: 'destructive', text: `${days} days` }
    return { variant: 'destructive', text: `${days} days (STALE)` }
  }

  const isStale = (days: number | string): boolean => {
    return typeof days === 'number' && days > 90
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Outstanding Checks</h2>
          <p className="text-muted-foreground">Checks that have not been cleared by the bank</p>
        </div>
        <Button onClick={loadOutstandingChecks} disabled={loading}>
          <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      <div className="flex gap-4 items-end">
        <div className="flex-1 max-w-xs">
          <Label htmlFor="account-filter">Bank Account</Label>
          <div>
            <Select id="account-filter" value={selectedAccount} onChange={(e: ChangeEvent<HTMLSelectElement>) => setSelectedAccount(e.target.value)} className="h-10">
              <option value="all">All Accounts ({bankAccounts.length} loaded)</option>
              {bankAccounts.map((account, idx) => {
                const accountNumber = account.account_number || account.accountNumber || ''
                const accountName = account.account_name || ''
                return (<option key={`${idx}-${accountNumber}`} value={accountNumber}>{accountNumber} - {accountName}</option>)
              })}
            </Select>
          </div>
        </div>

        <div className="flex-1 max-w-sm">
          <Label htmlFor="search">Search</Label>
          <div className="relative">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input id="search" placeholder="Search checks..." value={searchTerm} onChange={(e: ChangeEvent<HTMLInputElement>) => setSearchTerm(e.target.value)} className="pl-8" />
          </div>
        </div>

        <div className="w-32">
          <Label htmlFor="page-size">Page Size</Label>
          <Select 
            id="page-size" 
            value={pageSize.toString()} 
            onChange={(e: ChangeEvent<HTMLSelectElement>) => setPageSize(parseInt(e.target.value))}
            className="h-10"
          >
            <option value="10">10</option>
            <option value="25">25</option>
            <option value="50">50</option>
            <option value="100">100</option>
          </Select>
        </div>

        <Button variant={showStaleOnly ? 'default' : 'outline'} onClick={() => setShowStaleOnly(!showStaleOnly)}>
          <Filter className="mr-2 h-4 w-4" />
          Stale Only
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Outstanding Checks</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{processedChecks.length}</div>
            <p className="text-xs text-muted-foreground">{selectedAccount === 'all' ? 'All accounts' : selectedAccount}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Amount</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(processedChecks.reduce((sum, check) => sum + (check.amount || 0), 0))}</div>
            <p className="text-xs text-muted-foreground">Outstanding check amount</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Stale Checks</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{processedChecks.filter(check => { const days = calculateDaysOutstanding(check.date); return days !== 'N/A' && days > 90 }).length}</div>
            <p className="text-xs text-muted-foreground">Over 90 days old</p>
          </CardContent>
        </Card>
      </div>

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

      <Card>
        <CardHeader>
          <CardTitle>Outstanding Checks List</CardTitle>
          <CardDescription>{loading ? 'Loading checks...' : processedChecks.length === 0 ? 'No outstanding checks found' : `Showing ${paginatedChecks.length} of ${processedChecks.length} checks (Page ${currentPage} of ${totalPages})`}</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8"><RefreshCw className="h-6 w-6 animate-spin" /><span className="ml-2">Loading outstanding checks...</span></div>
          ) : processedChecks.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">{error ? 'Unable to load checks' : 'No outstanding checks found'}</div>
          ) : (
            <>
              <div className="overflow-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>
                        <Button variant="ghost" size="sm" onClick={() => handleSort('checkNumber')} className="h-8 px-2">Check # {getSortIcon('checkNumber')}</Button>
                      </TableHead>
                      <TableHead>
                        <Button variant="ghost" size="sm" onClick={() => handleSort('date')} className="h-8 px-2">Date {getSortIcon('date')}</Button>
                      </TableHead>
                      <TableHead>
                        <Button variant="ghost" size="sm" onClick={() => handleSort('payee')} className="h-8 px-2">Payee {getSortIcon('payee')}</Button>
                      </TableHead>
                      <TableHead className="text-right">
                        <Button variant="ghost" size="sm" onClick={() => handleSort('amount')} className="h-8 px-2">Amount {getSortIcon('amount')}</Button>
                      </TableHead>
                      <TableHead>Account</TableHead>
                      <TableHead className="text-center">
                        <Button variant="ghost" size="sm" onClick={() => handleSort('daysOutstanding')} className="h-8 px-2">Days Outstanding {getSortIcon('daysOutstanding')}</Button>
                      </TableHead>
                      <TableHead className="text-center">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedChecks.map((check, index) => {
                      const daysOut = calculateDaysOutstanding(check.date)
                      const daysBadge = getDaysOutstandingBadge(daysOut)
                      return (
                        <TableRow key={index} className="cursor-pointer hover:bg-muted/50" onClick={() => { setSelectedCheck(check); setEditedCheck({ ...check }); setEditMode(false) }}>
                          <TableCell className="font-medium">{check.checkNumber || 'N/A'}</TableCell>
                          <TableCell><div className="flex items-center"><Calendar className="mr-2 h-4 w-4 text-muted-foreground" />{formatDate(check.date)}</div></TableCell>
                          <TableCell>{check.payee || 'N/A'}</TableCell>
                          <TableCell className="text-right font-medium">{formatCurrency(check.amount)}</TableCell>
                          <TableCell>{check.account || 'N/A'}</TableCell>
                          <TableCell className="text-center"><Badge variant={daysBadge.variant}>{daysBadge.text}</Badge></TableCell>
                          <TableCell className="text-center">
                            <Button variant="ghost" size="sm" onClick={(e: MouseEvent<HTMLButtonElement>) => { e.stopPropagation(); setSelectedCheck(check); setEditedCheck({ ...check }) }}>
                              <Eye className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              </div>

              {totalPages > 1 && (
                <div className="flex items-center justify-between mt-4">
                  <div className="text-sm text-muted-foreground">Showing {((currentPage - 1) * pageSize) + 1} to {Math.min(currentPage * pageSize, processedChecks.length)} of {processedChecks.length} checks</div>
                  <div className="flex gap-2">
                    <Button variant="outline" size="sm" onClick={() => setCurrentPage(Math.max(1, currentPage - 1))} disabled={currentPage === 1}>
                      <ChevronLeft className="h-4 w-4" />
                      Previous
                    </Button>
                    <div className="flex gap-1">
                      {[...Array(Math.min(5, totalPages))].map((_, i) => {
                        let pageNum
                        if (totalPages <= 5) pageNum = i + 1
                        else if (currentPage <= 3) pageNum = i + 1
                        else if (currentPage >= totalPages - 2) pageNum = totalPages - 4 + i
                        else pageNum = currentPage - 2 + i
                        return (
                          <Button key={i} variant={pageNum === currentPage ? 'default' : 'outline'} size="sm" onClick={() => setCurrentPage(pageNum)} className="w-8">{pageNum}</Button>
                        )
                      })}
                    </div>
                    <Button variant="outline" size="sm" onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))} disabled={currentPage === totalPages}>
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

      {selectedCheck && (
        <Dialog open={!!selectedCheck} onOpenChange={() => setSelectedCheck(null)}>
          <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle className="flex items-center justify-between">
                <span>Check #{selectedCheck.checkNumber}</span>
                {canEdit && !editMode && (
                  <Button variant="outline" size="sm" onClick={() => setEditMode(true)}>
                    <Edit className="h-4 w-4 mr-2" />
                    Edit
                  </Button>
                )}
              </DialogTitle>
              <DialogDescription>{editMode ? 'Edit check details' : 'View check details'}</DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="check-number">Check Number</Label>
                  <Input id="check-number" value={editedCheck.checkNumber || ''} onChange={(e: ChangeEvent<HTMLInputElement>) => setEditedCheck({ ...editedCheck, checkNumber: e.target.value })} disabled={!editMode} />
                </div>
                <div>
                  <Label htmlFor="check-date">Date</Label>
                  <Input id="check-date" type="date" value={editedCheck.date || ''} onChange={(e: ChangeEvent<HTMLInputElement>) => setEditedCheck({ ...editedCheck, date: e.target.value })} disabled={!editMode} />
                </div>
              </div>
              <div>
                <Label htmlFor="payee">Payee</Label>
                <Input id="payee" value={editedCheck.payee || ''} onChange={(e: ChangeEvent<HTMLInputElement>) => setEditedCheck({ ...editedCheck, payee: e.target.value })} disabled={!editMode} />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="amount">Amount</Label>
                  <Input id="amount" type="number" step="0.01" value={editedCheck.amount || ''} onChange={(e: ChangeEvent<HTMLInputElement>) => setEditedCheck({ ...editedCheck, amount: parseFloat(e.target.value) || 0 })} disabled={!editMode} />
                </div>
                <div>
                  <Label htmlFor="account">Account</Label>
                  <Input id="account" value={editedCheck.account || ''} onChange={(e: ChangeEvent<HTMLInputElement>) => setEditedCheck({ ...editedCheck, account: e.target.value })} disabled={!editMode} />
                </div>
              </div>
              <div className="border-t pt-4">
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div><span className="font-medium">Days Outstanding:</span><span className="ml-2">{calculateDaysOutstanding(selectedCheck.date)}</span></div>
                  <div><span className="font-medium">Status:</span><Badge className="ml-2" variant={isStale(calculateDaysOutstanding(selectedCheck.date)) ? 'destructive' : 'default'}>{isStale(calculateDaysOutstanding(selectedCheck.date)) ? 'Stale' : 'Outstanding'}</Badge></div>
                </div>
              </div>
            </div>
            <DialogFooter>
              {editMode ? (
                <>
                  <Button variant="outline" onClick={() => { setEditedCheck({ ...selectedCheck }); setEditMode(false) }}>
                    <X className="h-4 w-4 mr-2" />
                    Cancel
                  </Button>
                  <Button onClick={handleSaveEdit}>
                    <CheckIcon className="h-4 w-4 mr-2" />
                    Save Changes
                  </Button>
                </>
              ) : (
                <Button variant="outline" onClick={() => setSelectedCheck(null)}>Close</Button>
              )}
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default OutstandingChecks
