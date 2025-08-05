import React, { useState, useEffect } from 'react'
import { 
  GetCachedBalances, 
  RefreshAccountBalance, 
  RefreshAllBalances,
  GetBalanceHistory 
} from '../../wailsjs/go/main/App'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Badge } from './ui/badge'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './ui/table'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from './ui/dialog'
import { 
  RefreshCw,
  Clock,
  AlertTriangle,
  CheckCircle,
  History,
  TrendingUp,
  TrendingDown,
  DollarSign,
  Banknote
} from 'lucide-react'

const BankingCachedBalances = ({ companyName, currentUser }) => {
  const [balances, setBalances] = useState([])
  const [loading, setLoading] = useState(false)
  const [refreshing, setRefreshing] = useState({})
  const [error, setError] = useState('')
  const [selectedAccount, setSelectedAccount] = useState(null)
  const [balanceHistory, setBalanceHistory] = useState([])
  const [showHistory, setShowHistory] = useState(false)

  // Load cached balances
  const loadCachedBalances = async () => {
    if (!companyName) return
    
    setLoading(true)
    setError('')
    
    try {
      const result = await GetCachedBalances(companyName)
      setBalances(result || [])
    } catch (err) {
      console.error('Failed to load cached balances:', err)
      setError(err.message || 'Failed to load cached balances')
      setBalances([])
    } finally {
      setLoading(false)
    }
  }

  // Refresh single account balance
  const refreshAccountBalance = async (accountNumber) => {
    setRefreshing(prev => ({ ...prev, [accountNumber]: true }))
    
    try {
      await RefreshAccountBalance(companyName, accountNumber)
      await loadCachedBalances() // Reload all balances
    } catch (err) {
      console.error('Failed to refresh account balance:', err)
      alert('Failed to refresh balance: ' + err.message)
    } finally {
      setRefreshing(prev => ({ ...prev, [accountNumber]: false }))
    }
  }

  // Refresh all balances
  const refreshAllBalances = async () => {
    setRefreshing(prev => ({ ...prev, all: true }))
    
    try {
      const result = await RefreshAllBalances(companyName)
      console.log('Refresh all result:', result)
      await loadCachedBalances() // Reload all balances
    } catch (err) {
      console.error('Failed to refresh all balances:', err)
      alert('Failed to refresh all balances: ' + err.message)
    } finally {
      setRefreshing(prev => ({ ...prev, all: false }))
    }
  }

  // Load balance history for an account
  const loadBalanceHistory = async (accountNumber) => {
    try {
      const history = await GetBalanceHistory(companyName, accountNumber, 20)
      setBalanceHistory(history || [])
      setSelectedAccount(accountNumber)
      setShowHistory(true)
    } catch (err) {
      console.error('Failed to load balance history:', err)
      alert('Failed to load balance history: ' + err.message)
    }
  }

  // Load balances when component mounts
  useEffect(() => {
    loadCachedBalances()
  }, [companyName])

  // Format currency
  const formatCurrency = (amount) => {
    if (typeof amount !== 'number') return '$0.00'
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  // Format timestamp
  const formatTimestamp = (timestamp) => {
    if (!timestamp) return 'Never'
    try {
      const date = new Date(timestamp)
      if (isNaN(date.getTime())) return 'Invalid Date'
      return date.toLocaleString()
    } catch (err) {
      return 'Invalid Date'
    }
  }

  // Get freshness badge
  const getFreshnessBadge = (freshness) => {
    switch (freshness) {
      case 'fresh':
        return { variant: 'default', icon: CheckCircle, text: 'Fresh' }
      case 'aging':
        return { variant: 'secondary', icon: Clock, text: 'Aging' }
      case 'stale':
        return { variant: 'destructive', icon: AlertTriangle, text: 'Stale' }
      default:
        return { variant: 'secondary', icon: Clock, text: 'Unknown' }
    }
  }

  // Calculate summary statistics
  const summary = balances.reduce((acc, balance) => {
    acc.totalGLBalance += balance.gl_balance || 0
    acc.totalOutstanding += balance.outstanding_total || 0
    acc.totalAvailable += balance.bank_balance || 0
    acc.staleCount += balance.is_stale ? 1 : 0
    return acc
  }, { totalGLBalance: 0, totalOutstanding: 0, totalAvailable: 0, staleCount: 0 })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Bank Account Balances</h2>
          <p className="text-muted-foreground">
            Cached GL balances with outstanding checks calculation
          </p>
        </div>
        <Button 
          onClick={refreshAllBalances} 
          disabled={loading || refreshing.all}
        >
          <RefreshCw className={`mr-2 h-4 w-4 ${refreshing.all ? 'animate-spin' : ''}`} />
          Refresh All
        </Button>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">GL Balance</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(summary.totalGLBalance)}</div>
            <p className="text-xs text-muted-foreground">Total from General Ledger</p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Outstanding Checks</CardTitle>
            <Banknote className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">
              {formatCurrency(summary.totalOutstanding)}
            </div>
            <p className="text-xs text-muted-foreground">Uncleared checks</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Bank Balance</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className={`text-2xl font-bold ${summary.totalAvailable >= 0 ? 'text-green-600' : 'text-red-600'}`}>
              {formatCurrency(summary.totalAvailable)}
            </div>
            <p className="text-xs text-muted-foreground">GL balance + outstanding</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Stale Balances</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{summary.staleCount}</div>
            <p className="text-xs text-muted-foreground">Need refresh</p>
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

      {/* Balances Table */}
      <Card>
        <CardHeader>
          <CardTitle>Account Balances</CardTitle>
          <CardDescription>
            {loading ? 'Loading balances...' : 
             balances.length === 0 ? 'No cached balances found' :
             `Showing ${balances.length} bank accounts`}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <RefreshCw className="h-6 w-6 animate-spin" />
              <span className="ml-2">Loading cached balances...</span>
            </div>
          ) : balances.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              {error ? 'Unable to load balances' : 'No cached balances found'}
            </div>
          ) : (
            <div className="overflow-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Account</TableHead>
                    <TableHead className="text-right">GL Balance</TableHead>
                    <TableHead className="text-right">Outstanding</TableHead>
                    <TableHead className="text-right">Bank Balance</TableHead>
                    <TableHead className="text-center">GL Status</TableHead>
                    <TableHead className="text-center">Checks Status</TableHead>
                    <TableHead className="text-center">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {balances.map((balance, index) => {
                    const glBadge = getFreshnessBadge(balance.gl_freshness)
                    const checksBadge = getFreshnessBadge(balance.checks_freshness)
                    const GLIcon = glBadge.icon
                    const ChecksIcon = checksBadge.icon
                    
                    return (
                      <TableRow key={index}>
                        <TableCell className="font-medium">
                          <div>
                            <div>{balance.account_number}</div>
                            <div className="text-sm text-muted-foreground">
                              {balance.account_name || 'Unnamed Account'}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="font-mono">
                            {formatCurrency(balance.gl_balance)}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            {formatTimestamp(balance.gl_last_updated)}
                          </div>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="font-mono text-red-600">
                            {formatCurrency(balance.outstanding_total)}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            {balance.outstanding_count} checks
                          </div>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className={`font-mono font-bold ${balance.bank_balance >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                            {formatCurrency(balance.bank_balance)}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            {formatTimestamp(balance.checks_last_updated)}
                          </div>
                        </TableCell>
                        <TableCell className="text-center">
                          <Badge variant={glBadge.variant} className="flex items-center gap-1 w-fit mx-auto">
                            <GLIcon className="h-3 w-3" />
                            {glBadge.text}
                          </Badge>
                          <div className="text-xs text-muted-foreground mt-1">
                            {balance.gl_age_hours ? `${balance.gl_age_hours.toFixed(1)}h` : 'N/A'}
                          </div>
                        </TableCell>
                        <TableCell className="text-center">
                          <Badge variant={checksBadge.variant} className="flex items-center gap-1 w-fit mx-auto">
                            <ChecksIcon className="h-3 w-3" />
                            {checksBadge.text}
                          </Badge>
                          <div className="text-xs text-muted-foreground mt-1">
                            {balance.checks_age_hours ? `${balance.checks_age_hours.toFixed(1)}h` : 'N/A'}
                          </div>
                        </TableCell>
                        <TableCell className="text-center">
                          <div className="flex gap-1 justify-center">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => refreshAccountBalance(balance.account_number)}
                              disabled={refreshing[balance.account_number]}
                            >
                              <RefreshCw className={`h-3 w-3 ${refreshing[balance.account_number] ? 'animate-spin' : ''}`} />
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => loadBalanceHistory(balance.account_number)}
                            >
                              <History className="h-3 w-3" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Balance History Dialog */}
      {showHistory && (
        <Dialog open={showHistory} onOpenChange={setShowHistory}>
          <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Balance History - Account {selectedAccount}</DialogTitle>
              <DialogDescription>
                Recent balance changes and refresh history
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4">
              {balanceHistory.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  No balance history found for this account
                </div>
              ) : (
                <div className="overflow-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Date/Time</TableHead>
                        <TableHead>Change Type</TableHead>
                        <TableHead className="text-right">GL Balance</TableHead>
                        <TableHead className="text-right">Outstanding</TableHead>
                        <TableHead className="text-right">Bank Balance</TableHead>
                        <TableHead>Changed By</TableHead>
                        <TableHead>Reason</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {balanceHistory.map((entry, index) => (
                        <TableRow key={index}>
                          <TableCell className="font-mono text-sm">
                            {formatTimestamp(entry.change_timestamp)}
                          </TableCell>
                          <TableCell>
                            <Badge 
                              variant={entry.change_type === 'gl_refresh' ? 'default' : 'secondary'}
                            >
                              {entry.change_type.replace('_', ' ')}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-right">
                            <div className="space-y-1">
                              {entry.old_gl_balance !== null && (
                                <div className="text-xs text-muted-foreground line-through">
                                  {formatCurrency(entry.old_gl_balance)}
                                </div>
                              )}
                              {entry.new_gl_balance !== null && (
                                <div className="font-mono">
                                  {formatCurrency(entry.new_gl_balance)}
                                </div>
                              )}
                            </div>
                          </TableCell>
                          <TableCell className="text-right">
                            <div className="space-y-1">
                              {entry.old_outstanding_total !== null && (
                                <div className="text-xs text-muted-foreground line-through">
                                  {formatCurrency(entry.old_outstanding_total)}
                                </div>
                              )}
                              {entry.new_outstanding_total !== null && (
                                <div className="font-mono">
                                  {formatCurrency(entry.new_outstanding_total)}
                                </div>
                              )}
                            </div>
                          </TableCell>
                          <TableCell className="text-right">
                            <div className="space-y-1">
                              {entry.old_bank_balance !== null && (
                                <div className="text-xs text-muted-foreground line-through">
                                  {formatCurrency(entry.old_bank_balance)}
                                </div>
                              )}
                              {entry.new_bank_balance !== null && (
                                <div className={`font-mono ${entry.new_bank_balance >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                                  {formatCurrency(entry.new_bank_balance)}
                                </div>
                              )}
                            </div>
                          </TableCell>
                          <TableCell>{entry.changed_by || 'System'}</TableCell>
                          <TableCell className="text-sm">{entry.change_reason}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setShowHistory(false)}>
                Close
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default BankingCachedBalances