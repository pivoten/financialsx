
import React, { useState, useEffect } from 'react'
import { GetOutstandingChecks, GetBankAccounts } from '../../wailsjs/go/main/App'
import { Badge } from './ui/badge'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from './ui/table'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from './ui/dialog'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Card, CardHeader, CardTitle, CardContent } from './ui/card'
import { AlertTriangle, DollarSign, Calendar, Filter, Edit, X, Check } from 'lucide-react'
import logger from '../services/logger'

interface OutstandingChecksSimpleProps {
  companyName: string
  currentUser: {
    id: number
    username: string
    email: string
    role_name: string
    is_root: boolean
    company_name: string
  }
}

interface CheckRecord {
  checkNumber: string
  date: string
  payee: string
  amount: number
  account: string
  daysOutstanding?: number
}

const OutstandingChecksSimple = ({ companyName, currentUser }: OutstandingChecksSimpleProps) => {
  const [outstandingChecks, setOutstandingChecks] = useState<CheckRecord[]>([])
  const [bankAccounts, setBankAccounts] = useState<any[]>([])
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string>('')
  const [selectedCheck, setSelectedCheck] = useState<CheckRecord | null>(null)
  const [editMode, setEditMode] = useState<boolean>(false)
  const [editedCheck, setEditedCheck] = useState<Partial<CheckRecord>>({})

  const canEdit = currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')

  useEffect(() => {
    loadBankAccounts()
    loadOutstandingChecks('')
  }, [companyName])

  const loadBankAccounts = async () => {
    if (!companyName) return
    try {
      const accounts = await GetBankAccounts(companyName)
      if (accounts && Array.isArray(accounts)) setBankAccounts(accounts)
    } catch (err) {}
  }

  const loadOutstandingChecks = async (accountFilter = '') => {
    if (!companyName) return
    setLoading(true)
    setError('')
    try {
      const result = await GetOutstandingChecks(companyName, accountFilter)
      if (result.status === 'error') {
        setError(result.error || 'Failed to load outstanding checks')
        setOutstandingChecks([])
      } else {
        setOutstandingChecks(result.checks || [])
      }
    } catch (err) {
      setError(err.message || 'Failed to load outstanding checks')
      setOutstandingChecks([])
    } finally {
      setLoading(false)
    }
  }

  const calculateDaysOutstanding = (checkDate: string) => {
    if (!checkDate) return 'N/A'
    try {
      const today = new Date()
      const checkDateTime = new Date(checkDate)
      if (isNaN(checkDateTime.getTime())) return 'N/A'
      const diffTime = today.getTime() - checkDateTime.getTime()
      const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
      return diffDays
    } catch (err) {
      return 'N/A'
    }
  }

  const formatCurrency = (amount: number) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(typeof amount === 'number' ? amount : 0)
  const formatDate = (dateStr: string) => {
    if (!dateStr) return 'N/A'
    try {
      const date = new Date(dateStr)
      if (isNaN(date.getTime())) return dateStr
      return date.toLocaleDateString()
    } catch (err) {
      return dateStr
    }
  }

  const getDaysOutstandingBadge = (days: number | string): { variant: 'default' | 'secondary' | 'destructive' | 'outline', text: string } => {
    if (days === 'N/A' || typeof days === 'string') return { variant: 'secondary', text: 'N/A' }
    const numDays = typeof days === 'number' ? days : 0
    if (numDays <= 30) return { variant: 'default', text: `${numDays} days` }
    if (numDays <= 60) return { variant: 'secondary', text: `${numDays} days` }
    if (numDays <= 90) return { variant: 'destructive', text: `${numDays} days` }
    return { variant: 'destructive', text: `${numDays} days (STALE)` }
  }

  const columns = [
    { accessor: 'checkNumber', header: 'Check #', sortable: true, type: 'number' },
    { accessor: 'date', header: 'Date', sortable: true, type: 'date', render: (value: string) => (
      <div className="flex items-center">
        <Calendar className="mr-2 h-4 w-4 text-muted-foreground" />
        {formatDate(value)}
      </div>
    )},
    { accessor: 'payee', header: 'Payee', sortable: true },
    { accessor: 'amount', header: 'Amount', headerClassName: 'text-right', cellClassName: 'text-right font-medium', sortable: true, type: 'number', render: (value: number) => formatCurrency(value) },
    { accessor: 'account', header: 'Account', sortable: true },
    { accessor: 'daysOutstanding', header: 'Days Outstanding', headerClassName: 'text-center', cellClassName: 'text-center', sortable: false, render: (_: any, row: any) => {
      const days = calculateDaysOutstanding(row.date)
      const badge = getDaysOutstandingBadge(days)
      return <Badge variant={badge.variant}>{badge.text}</Badge>
    }}
  ]

  const filters = [
    { key: 'account', label: 'Bank Account', placeholder: 'Select account', defaultValue: 'all', options: [
      { value: 'all', label: 'All Accounts' },
      ...bankAccounts.map(account => ({ value: account.account_number, label: `${account.account_number} - ${account.account_name}` }))
    ], filterFn: (row: any, value: string) => value === 'all' || row.account === value }
  ]

  const actions = [
    { label: 'Stale Only', icon: <Filter className="mr-2 h-4 w-4" />, variant: 'outline', onClick: () => { logger.debug('Toggle stale filter') } }
  ]

  const handleRowClick = (check: CheckRecord) => { setSelectedCheck(check); setEditedCheck({ ...check }); setEditMode(false) }
  const handleRefresh = () => { loadOutstandingChecks('') }
  const handleSaveEdit = async () => { logger.debug('Save edit', { checkNumber: editedCheck?.checkNumber }); setSelectedCheck(null); setEditMode(false) }

  const totalAmount = outstandingChecks.reduce((sum, check) => sum + (check.amount || 0), 0)
  const staleCount = outstandingChecks.filter(check => { const days = calculateDaysOutstanding(check.date); return days !== 'N/A' && days > 90 }).length

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Outstanding Checks</h2>
        <p className="text-muted-foreground">Checks that have not been cleared by the bank</p>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Outstanding Checks</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{outstandingChecks.length}</div>
            <p className="text-xs text-muted-foreground">Total checks</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Amount</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(totalAmount)}</div>
            <p className="text-xs text-muted-foreground">Outstanding amount</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Stale Checks</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{staleCount}</div>
            <p className="text-xs text-muted-foreground">Over 90 days old</p>
          </CardContent>
        </Card>
      </div>

      <div className="data-table-placeholder">
        {loading ? (
          <div>Loading checks...</div>
        ) : error ? (
          <div>Error: {error}</div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                {columns.map((col, index) => (
                  <TableHead key={index}>{col.header}</TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {outstandingChecks.map((check, index) => (
                <TableRow key={index} onClick={() => handleRowClick(check)} className="cursor-pointer hover:bg-muted/50">
                  {columns.map((col, colIndex) => (
                    <TableCell key={colIndex}>
                      {(col as any).render ? (col as any).render((check as any)[col.accessor], check as any) : (check as any)[col.accessor]}
                    </TableCell>
                  ))}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </div>

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
                  <Input id="check-number" value={editedCheck.checkNumber || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEditedCheck({...editedCheck, checkNumber: e.target.value})} disabled={!editMode} />
                </div>
                <div>
                  <Label htmlFor="check-date">Date</Label>
                  <Input id="check-date" type="date" value={editedCheck.date || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEditedCheck({...editedCheck, date: e.target.value})} disabled={!editMode} />
                </div>
              </div>

              <div>
                <Label htmlFor="payee">Payee</Label>
                <Input id="payee" value={editedCheck.payee || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEditedCheck({...editedCheck, payee: e.target.value})} disabled={!editMode} />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="amount">Amount</Label>
                  <Input id="amount" type="number" step="0.01" value={editedCheck.amount || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEditedCheck({...editedCheck, amount: parseFloat(e.target.value)})} disabled={!editMode} />
                </div>
                <div>
                  <Label htmlFor="account">Account</Label>
                  <Input id="account" value={editedCheck.account || ''} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEditedCheck({...editedCheck, account: e.target.value})} disabled={!editMode} />
                </div>
              </div>

              <div className="border-t pt-4">
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div>
                    <span className="font-medium">Days Outstanding:</span>
                    <span className="ml-2">{calculateDaysOutstanding(selectedCheck.date)}</span>
                  </div>
                  <div>
                    <span className="font-medium">Status:</span>
                    {(() => {
                      const days = calculateDaysOutstanding(selectedCheck.date)
                      const isStale = typeof days === 'number' && days > 90
                      return (
                        <Badge className="ml-2" variant={isStale ? 'destructive' : 'default'}>
                          {isStale ? 'Stale' : 'Outstanding'}
                        </Badge>
                      )
                    })()}
                  </div>
                </div>
              </div>
            </div>

            <DialogFooter>
              {editMode ? (
                <>
                  <Button variant="outline" onClick={() => { setEditedCheck({...selectedCheck}); setEditMode(false) }}>
                    <X className="h-4 w-4 mr-2" />
                    Cancel
                  </Button>
                  <Button onClick={handleSaveEdit}>
                    <Check className="h-4 w-4 mr-2" />
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

export default OutstandingChecksSimple
