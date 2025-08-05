import React, { useState, useEffect } from 'react'
import { GetOutstandingChecks, GetBankAccounts } from '../../wailsjs/go/main/App'
import { Badge } from './ui/badge'
import DataTable from './ui/data-table'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from './ui/dialog'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Card, CardHeader, CardTitle, CardContent } from './ui/card'
import { 
  AlertTriangle,
  DollarSign,
  Calendar,
  Filter,
  Edit,
  X,
  Check
} from 'lucide-react'

const OutstandingChecksSimple = ({ companyName, currentUser }) => {
  // State
  const [outstandingChecks, setOutstandingChecks] = useState([])
  const [bankAccounts, setBankAccounts] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [selectedCheck, setSelectedCheck] = useState(null)
  const [editMode, setEditMode] = useState(false)
  const [editedCheck, setEditedCheck] = useState({})

  // Permission check
  const canEdit = currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')

  // Load data
  useEffect(() => {
    loadBankAccounts()
    loadOutstandingChecks('')
  }, [companyName])

  const loadBankAccounts = async () => {
    if (!companyName) return
    try {
      const accounts = await GetBankAccounts(companyName)
      if (accounts && Array.isArray(accounts)) {
        setBankAccounts(accounts)
      }
    } catch (err) {
      console.error('Failed to load bank accounts:', err)
    }
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
      console.error('Failed to load outstanding checks:', err)
      setError(err.message || 'Failed to load outstanding checks')
      setOutstandingChecks([])
    } finally {
      setLoading(false)
    }
  }

  // Helper functions
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

  const formatCurrency = (amount) => {
    if (typeof amount !== 'number') return '$0.00'
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

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

  const getDaysOutstandingBadge = (days) => {
    if (days === 'N/A') return { variant: 'secondary', text: 'N/A' }
    if (days <= 30) return { variant: 'default', text: `${days} days` }
    if (days <= 60) return { variant: 'secondary', text: `${days} days` }
    if (days <= 90) return { variant: 'destructive', text: `${days} days` }
    return { variant: 'destructive', text: `${days} days (STALE)` }
  }

  // DataTable configuration
  const columns = [
    {
      accessor: 'checkNumber',
      header: 'Check #',
      sortable: true,
      type: 'number'
    },
    {
      accessor: 'date',
      header: 'Date',
      sortable: true,
      type: 'date',
      render: (value) => (
        <div className="flex items-center">
          <Calendar className="mr-2 h-4 w-4 text-muted-foreground" />
          {formatDate(value)}
        </div>
      )
    },
    {
      accessor: 'payee',
      header: 'Payee',
      sortable: true
    },
    {
      accessor: 'amount',
      header: 'Amount',
      headerClassName: 'text-right',
      cellClassName: 'text-right font-medium',
      sortable: true,
      type: 'number',
      render: (value) => formatCurrency(value)
    },
    {
      accessor: 'account',
      header: 'Account',
      sortable: true
    },
    {
      accessor: 'daysOutstanding',
      header: 'Days Outstanding',
      headerClassName: 'text-center',
      cellClassName: 'text-center',
      sortable: false,
      render: (_, row) => {
        const days = calculateDaysOutstanding(row.date)
        const badge = getDaysOutstandingBadge(days)
        return (
          <Badge variant={badge.variant}>
            {badge.text}
          </Badge>
        )
      }
    }
  ]

  const filters = [
    {
      key: 'account',
      label: 'Bank Account',
      placeholder: 'Select account',
      defaultValue: 'all',
      options: [
        { value: 'all', label: 'All Accounts' },
        ...bankAccounts.map(account => ({
          value: account.account_number,
          label: `${account.account_number} - ${account.account_name}`
        }))
      ],
      filterFn: (row, value) => value === 'all' || row.account === value
    }
  ]

  const actions = [
    {
      label: 'Stale Only',
      icon: <Filter className="mr-2 h-4 w-4" />,
      variant: 'outline',
      onClick: () => {
        // Toggle stale filter - this would need to be implemented in the DataTable
        console.log('Toggle stale filter')
      }
    }
  ]

  // Event handlers
  const handleRowClick = (check) => {
    setSelectedCheck(check)
    setEditedCheck({ ...check })
    setEditMode(false)
  }

  const handleRefresh = () => {
    // Get current account filter and refresh
    loadOutstandingChecks('')
  }

  const handleSaveEdit = async () => {
    // Implementation would go here
    console.log('Save edit:', editedCheck)
    setSelectedCheck(null)
    setEditMode(false)
  }

  // Calculate summary stats
  const totalAmount = outstandingChecks.reduce((sum, check) => sum + (check.amount || 0), 0)
  const staleCount = outstandingChecks.filter(check => {
    const days = calculateDaysOutstanding(check.date)
    return days !== 'N/A' && days > 90
  }).length

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Outstanding Checks</h2>
        <p className="text-muted-foreground">
          Checks that have not been cleared by the bank
        </p>
      </div>

      {/* Summary Cards */}
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

      {/* Data Table */}
      <DataTable
        data={outstandingChecks}
        columns={columns}
        title="Outstanding Checks List"
        loading={loading}
        error={error}
        onRowClick={handleRowClick}
        onRefresh={handleRefresh}
        searchPlaceholder="Search checks..."
        filters={filters}
        actions={actions}
        pageSize={25}
      />

      {/* Detail/Edit Modal */}
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

export default OutstandingChecksSimple