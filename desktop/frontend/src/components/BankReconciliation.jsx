import { useState, useEffect } from 'react'
import { GetDBFTableData, UpdateDBFRecord, GetBankAccounts } from '../../wailsjs/go/main/App'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Badge } from './ui/badge'
import { Checkbox } from './ui/checkbox'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './ui/table'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './ui/tabs'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from './ui/dialog'
import { 
  CheckCircle, 
  AlertCircle, 
  Clock,
  DollarSign,
  Calendar,
  FileText,
  Download,
  Upload,
  Save,
  RefreshCw,
  Filter,
  Search,
  Calculator
} from 'lucide-react'

export function BankReconciliation({ companyName }) {
  // State management
  const [checks, setChecks] = useState([])
  const [bankAccounts, setBankAccounts] = useState([])
  const [loading, setLoading] = useState(true)
  const [loadingAccounts, setLoadingAccounts] = useState(true)
  const [error, setError] = useState(null)
  const [selectedAccount, setSelectedAccount] = useState('')
  const [reconciliationPeriod, setReconciliationPeriod] = useState('')
  const [statementBalance, setStatementBalance] = useState('')
  const [statementDate, setStatementDate] = useState('')
  const [selectedChecks, setSelectedChecks] = useState(new Set())
  const [reconciliationInProgress, setReconciliationInProgress] = useState(false)

  // Filters
  const [showCleared, setShowCleared] = useState(true)
  const [showUncleared, setShowUncleared] = useState(true)
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [amountFrom, setAmountFrom] = useState('')
  const [amountTo, setAmountTo] = useState('')

  // Load data when component mounts
  useEffect(() => {
    if (companyName) {
      loadBankAccounts()
      loadChecksData()
    }
  }, [companyName])

  // Load bank accounts from COA.dbf
  const loadBankAccounts = async () => {
    try {
      setLoadingAccounts(true)
      setError(null)
      
      console.log('Loading bank accounts for company:', companyName)
      
      // Check if GetBankAccounts is available (Wails bindings generated)
      if (typeof GetBankAccounts === 'function') {
        console.log('GetBankAccounts function is available, calling it...')
        try {
          const accounts = await GetBankAccounts(companyName)
          console.log('GetBankAccounts response:', accounts)
          console.log('GetBankAccounts response type:', typeof accounts)
          console.log('GetBankAccounts response length:', accounts?.length)
          
          if (accounts && Array.isArray(accounts)) {
            setBankAccounts(accounts)
            
            // Auto-select first account if available
            if (accounts.length > 0 && !selectedAccount) {
              setSelectedAccount(accounts[0].account_number)
              console.log('Auto-selected account:', accounts[0].account_number)
            }
          } else {
            console.log('GetBankAccounts returned invalid data, using fallback')
            throw new Error('GetBankAccounts returned invalid data: ' + typeof accounts)
          }
        } catch (getBankAccountsErr) {
          console.error('GetBankAccounts call failed:', getBankAccountsErr)
          console.error('Error details:', getBankAccountsErr.message, getBankAccountsErr.stack)
          // Fall through to fallback method
          throw getBankAccountsErr
        }
      } else {
        console.log('GetBankAccounts function not available')
        throw new Error('GetBankAccounts function not available')
      }
    } catch (err) {
      console.log('Primary method failed, trying fallback...')
      // Fallback: Try to read COA.dbf directly using existing function
      try {
        const coaData = await GetDBFTableData(companyName, 'COA.dbf')
        console.log('COA.dbf data loaded:', coaData)
        
        if (coaData && coaData.data) {
          const bankAccounts = coaData.data
            .filter(row => {
              // Check if LBANKACCT is true (column 6 based on COA structure)
              const bankFlag = row[6]
              console.log('Checking bank flag for account:', row[0], 'flag:', bankFlag, 'type:', typeof bankFlag)
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
          
          console.log('Filtered bank accounts:', bankAccounts)
          setBankAccounts(bankAccounts)
          
          // Auto-select first account if available
          if (bankAccounts.length > 0 && !selectedAccount) {
            setSelectedAccount(bankAccounts[0].account_number)
            console.log('Auto-selected account (fallback):', bankAccounts[0].account_number)
          }
        } else {
          console.log('COA.dbf fallback - no data found. coaData structure:', coaData)
          throw new Error('COA.dbf file exists but contains no data rows')
        }
      } catch (fallbackErr) {
        console.error('Fallback method also failed:', fallbackErr)
        setError(`Failed to load bank accounts: ${err.message}. Fallback also failed: ${fallbackErr.message}`)
      }
    } finally {
      setLoadingAccounts(false)
    }
  }

  const loadChecksData = async () => {
    try {
      setLoading(true)
      setError(null)
      
      const response = await GetDBFTableData(companyName, 'checks.dbf')
      
      if (response && response.data) {
        // Process the DBF data into our check format
        const processedChecks = response.data.map((row, index) => {
          // Assuming DBF structure based on typical checks.dbf format
          // You may need to adjust column indices based on actual DBF structure
          return {
            id: index,
            checkNumber: row[0] || '', // Assuming first column is check number
            payee: row[1] || '', // Payee name
            amount: parseFloat(row[2]) || 0, // Check amount
            checkDate: row[3] || '', // Check date
            accountNumber: row[4] || '', // CACCTNO field
            cleared: row[5] === true || row[5] === 'T' || row[5] === '.T.', // LCLEARED field
            reconcileDate: row[6] || '', // DRECDATE field
            memo: row[7] || '', // Memo/description
            voidFlag: row[8] || false, // Void flag if exists
            originalRow: row // Keep original for updates
          }
        })
        
        setChecks(processedChecks)
      }
    } catch (err) {
      console.error('Error loading checks data:', err)
      setError('Failed to load checks data: ' + err.message)
    } finally {
      setLoading(false)
    }
  }

  // Bank accounts are now loaded from COA.dbf

  // Filter checks based on current filters
  const filteredChecks = checks.filter(check => {
    // Account filter
    if (selectedAccount && check.accountNumber !== selectedAccount) return false
    
    // Cleared status filter
    if (!showCleared && check.cleared) return false
    if (!showUncleared && !check.cleared) return false
    
    // Date range filter
    if (dateFrom && check.checkDate < dateFrom) return false
    if (dateTo && check.checkDate > dateTo) return false
    
    // Amount range filter
    if (amountFrom && check.amount < parseFloat(amountFrom)) return false
    if (amountTo && check.amount > parseFloat(amountTo)) return false
    
    return true
  })

  // Calculate reconciliation totals
  const calculateTotals = () => {
    const accountChecks = selectedAccount 
      ? checks.filter(check => check.accountNumber === selectedAccount)
      : checks

    const clearedChecks = accountChecks.filter(check => check.cleared)
    const unclearedChecks = accountChecks.filter(check => !check.cleared)
    
    const clearedTotal = clearedChecks.reduce((sum, check) => sum + check.amount, 0)
    const unclearedTotal = unclearedChecks.reduce((sum, check) => sum + check.amount, 0)
    const bookBalance = clearedTotal + unclearedTotal
    
    const statementBal = parseFloat(statementBalance) || 0
    const difference = statementBal - clearedTotal

    return {
      clearedCount: clearedChecks.length,
      unclearedCount: unclearedChecks.length,
      clearedTotal,
      unclearedTotal,
      bookBalance,
      statementBalance: statementBal,
      difference
    }
  }

  const totals = calculateTotals()

  // Toggle check cleared status
  const toggleCheckCleared = async (checkId) => {
    try {
      const check = checks.find(c => c.id === checkId)
      if (!check) return

      const newClearedStatus = !check.cleared
      const currentDate = new Date().toISOString().split('T')[0]

      // Update the DBF record
      // You'll need to determine the correct column indices for LCLEARED and DRECDATE
      const clearedColumnIndex = 5 // Adjust based on actual DBF structure
      const dateColumnIndex = 6 // Adjust based on actual DBF structure

      // Update cleared status
      await UpdateDBFRecord(
        companyName, 
        'checks.dbf', 
        checkId, 
        clearedColumnIndex, 
        newClearedStatus ? 'T' : 'F'
      )

      // Update reconcile date if being cleared
      if (newClearedStatus) {
        await UpdateDBFRecord(
          companyName, 
          'checks.dbf', 
          checkId, 
          dateColumnIndex, 
          currentDate
        )
      }

      // Update local state
      setChecks(prevChecks => 
        prevChecks.map(c => 
          c.id === checkId 
            ? { 
                ...c, 
                cleared: newClearedStatus, 
                reconcileDate: newClearedStatus ? currentDate : ''
              }
            : c
        )
      )

    } catch (err) {
      console.error('Error updating check status:', err)
      setError('Failed to update check status: ' + err.message)
    }
  }

  // Bulk clear selected checks
  const bulkClearChecks = async () => {
    setReconciliationInProgress(true)
    try {
      const currentDate = new Date().toISOString().split('T')[0]
      
      for (const checkId of selectedChecks) {
        await toggleCheckCleared(checkId)
      }
      
      setSelectedChecks(new Set())
    } catch (err) {
      console.error('Error in bulk clear:', err)
      setError('Failed to clear checks: ' + err.message)
    } finally {
      setReconciliationInProgress(false)
    }
  }

  // Format currency
  const formatCurrency = (amount) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount || 0)
  }

  // Format date
  const formatDate = (dateStr) => {
    if (!dateStr) return ''
    try {
      const date = new Date(dateStr)
      return date.toLocaleDateString()
    } catch {
      return dateStr
    }
  }

  // Get status badge
  const getStatusBadge = (cleared) => {
    return cleared 
      ? <Badge variant="default" className="bg-green-100 text-green-800"><CheckCircle className="w-3 h-3 mr-1" />Cleared</Badge>
      : <Badge variant="secondary" className="bg-yellow-100 text-yellow-800"><Clock className="w-3 h-3 mr-1" />Outstanding</Badge>
  }

  if (loading) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="text-center">
            <RefreshCw className="w-8 h-8 animate-spin mx-auto mb-4 text-muted-foreground" />
            <p className="text-muted-foreground">Loading checks data...</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="text-center">
            <AlertCircle className="w-8 h-8 mx-auto mb-4 text-red-500" />
            <p className="text-red-500 mb-4">{error}</p>
            <Button onClick={loadChecksData}>
              <RefreshCw className="w-4 h-4 mr-2" />
              Retry
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      <Tabs defaultValue="reconcile" className="w-full">
        <TabsList>
          <TabsTrigger value="reconcile">Reconcile</TabsTrigger>
          <TabsTrigger value="outstanding">Outstanding Checks</TabsTrigger>
          <TabsTrigger value="cleared">Cleared Checks</TabsTrigger>
          <TabsTrigger value="reports">Reports</TabsTrigger>
        </TabsList>

        {/* Main Reconciliation Tab */}
        <TabsContent value="reconcile" className="space-y-4">
          {/* Reconciliation Setup */}
          <Card>
            <CardHeader>
              <CardTitle>Bank Reconciliation Setup</CardTitle>
              <CardDescription>Configure reconciliation parameters</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-4">
                <div className="space-y-2">
                  <Label htmlFor="account">Bank Account</Label>
                  <select 
                    id="account" 
                    value={selectedAccount}
                    onChange={(e) => setSelectedAccount(e.target.value)}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    disabled={loadingAccounts}
                  >
                    <option value="">Select Bank Account</option>
                    {bankAccounts.map(account => (
                      <option key={account.account_number} value={account.account_number}>
                        {account.account_number} - {account.account_name}
                      </option>
                    ))}
                  </select>
                  {loadingAccounts && (
                    <p className="text-sm text-muted-foreground mt-1">Loading accounts...</p>
                  )}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="statement-date">Statement Date</Label>
                  <Input 
                    id="statement-date"
                    type="date"
                    value={statementDate}
                    onChange={(e) => setStatementDate(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="statement-balance">Statement Balance</Label>
                  <Input 
                    id="statement-balance"
                    type="number"
                    step="0.01"
                    placeholder="0.00"
                    value={statementBalance}
                    onChange={(e) => setStatementBalance(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Actions</Label>
                  <div className="flex gap-2">
                    <Button onClick={loadChecksData} size="sm">
                      <RefreshCw className="w-4 h-4" />
                    </Button>
                    <Button 
                      onClick={bulkClearChecks} 
                      size="sm" 
                      disabled={selectedChecks.size === 0 || reconciliationInProgress}
                    >
                      {reconciliationInProgress ? (
                        <RefreshCw className="w-4 h-4 animate-spin" />
                      ) : (
                        <CheckCircle className="w-4 h-4" />
                      )}
                    </Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Reconciliation Summary */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Statement Balance</p>
                    <p className="text-2xl font-bold">{formatCurrency(totals.statementBalance)}</p>
                  </div>
                  <FileText className="w-8 h-8 text-muted-foreground" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Cleared Checks</p>
                    <p className="text-2xl font-bold text-green-600">{formatCurrency(totals.clearedTotal)}</p>
                    <p className="text-xs text-muted-foreground">{totals.clearedCount} checks</p>
                  </div>
                  <CheckCircle className="w-8 h-8 text-green-600" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Outstanding Checks</p>
                    <p className="text-2xl font-bold text-yellow-600">{formatCurrency(totals.unclearedTotal)}</p>
                    <p className="text-xs text-muted-foreground">{totals.unclearedCount} checks</p>
                  </div>
                  <Clock className="w-8 h-8 text-yellow-600" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Difference</p>
                    <p className={`text-2xl font-bold ${totals.difference === 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {formatCurrency(totals.difference)}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {totals.difference === 0 ? 'Reconciled' : 'Needs Review'}
                    </p>
                  </div>
                  <Calculator className="w-8 h-8 text-muted-foreground" />
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Filters */}
          <Card>
            <CardHeader>
              <CardTitle>Filters</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-6">
                <div className="flex items-center space-x-2">
                  <Checkbox 
                    id="show-cleared"
                    checked={showCleared}
                    onCheckedChange={setShowCleared}
                  />
                  <Label htmlFor="show-cleared">Show Cleared</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <Checkbox 
                    id="show-uncleared"
                    checked={showUncleared}
                    onCheckedChange={setShowUncleared}
                  />
                  <Label htmlFor="show-uncleared">Show Outstanding</Label>
                </div>
                <div className="space-y-1">
                  <Label htmlFor="date-from" className="text-xs">Date From</Label>
                  <Input 
                    id="date-from"
                    type="date"
                    value={dateFrom}
                    onChange={(e) => setDateFrom(e.target.value)}
                    size="sm"
                  />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="date-to" className="text-xs">Date To</Label>
                  <Input 
                    id="date-to"
                    type="date"
                    value={dateTo}
                    onChange={(e) => setDateTo(e.target.value)}
                    size="sm"
                  />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="amount-from" className="text-xs">Amount From</Label>
                  <Input 
                    id="amount-from"
                    type="number"
                    step="0.01"
                    value={amountFrom}
                    onChange={(e) => setAmountFrom(e.target.value)}
                    size="sm"
                  />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="amount-to" className="text-xs">Amount To</Label>
                  <Input 
                    id="amount-to"
                    type="number"
                    step="0.01"
                    value={amountTo}
                    onChange={(e) => setAmountTo(e.target.value)}
                    size="sm"
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Checks Table */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Checks ({filteredChecks.length})</CardTitle>
                  <CardDescription>Click checkbox to select, click status to toggle cleared</CardDescription>
                </div>
                <div className="flex gap-2">
                  <Badge variant="outline">
                    {selectedChecks.size} selected
                  </Badge>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-12">
                      <Checkbox 
                        checked={selectedChecks.size === filteredChecks.length && filteredChecks.length > 0}
                        onCheckedChange={(checked) => {
                          if (checked) {
                            setSelectedChecks(new Set(filteredChecks.map(c => c.id)))
                          } else {
                            setSelectedChecks(new Set())
                          }
                        }}
                      />
                    </TableHead>
                    <TableHead>Check #</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Payee</TableHead>
                    <TableHead>Memo</TableHead>
                    <TableHead className="text-right">Amount</TableHead>
                    <TableHead>Account</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Reconciled</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredChecks.map((check) => (
                    <TableRow key={check.id} className={check.cleared ? 'bg-green-50' : ''}>
                      <TableCell>
                        <Checkbox 
                          checked={selectedChecks.has(check.id)}
                          onCheckedChange={(checked) => {
                            const newSelected = new Set(selectedChecks)
                            if (checked) {
                              newSelected.add(check.id)
                            } else {
                              newSelected.delete(check.id)
                            }
                            setSelectedChecks(newSelected)
                          }}
                        />
                      </TableCell>
                      <TableCell className="font-mono">{check.checkNumber}</TableCell>
                      <TableCell>{formatDate(check.checkDate)}</TableCell>
                      <TableCell>{check.payee}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{check.memo}</TableCell>
                      <TableCell className="text-right font-mono">{formatCurrency(check.amount)}</TableCell>
                      <TableCell className="font-mono text-sm">{check.accountNumber}</TableCell>
                      <TableCell>
                        <button onClick={() => toggleCheckCleared(check.id)}>
                          {getStatusBadge(check.cleared)}
                        </button>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {check.reconcileDate ? formatDate(check.reconcileDate) : '-'}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Outstanding Checks Tab */}
        <TabsContent value="outstanding" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Outstanding Checks</CardTitle>
              <CardDescription>Checks that have not been cleared by the bank</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Check #</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Payee</TableHead>
                    <TableHead className="text-right">Amount</TableHead>
                    <TableHead>Account</TableHead>
                    <TableHead>Days Outstanding</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {checks.filter(check => !check.cleared).map((check) => {
                    const daysOutstanding = check.checkDate 
                      ? Math.floor((new Date() - new Date(check.checkDate)) / (1000 * 60 * 60 * 24))
                      : 0
                    
                    return (
                      <TableRow key={check.id}>
                        <TableCell className="font-mono">{check.checkNumber}</TableCell>
                        <TableCell>{formatDate(check.checkDate)}</TableCell>
                        <TableCell>{check.payee}</TableCell>
                        <TableCell className="text-right font-mono">{formatCurrency(check.amount)}</TableCell>
                        <TableCell className="font-mono text-sm">{check.accountNumber}</TableCell>
                        <TableCell>
                          <span className={`${daysOutstanding > 90 ? 'text-red-600' : daysOutstanding > 30 ? 'text-yellow-600' : 'text-green-600'}`}>
                            {daysOutstanding} days
                          </span>
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Cleared Checks Tab */}
        <TabsContent value="cleared" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Cleared Checks</CardTitle>
              <CardDescription>Checks that have been cleared by the bank</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Check #</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Payee</TableHead>
                    <TableHead className="text-right">Amount</TableHead>
                    <TableHead>Account</TableHead>
                    <TableHead>Reconciled Date</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {checks.filter(check => check.cleared).map((check) => (
                    <TableRow key={check.id}>
                      <TableCell className="font-mono">{check.checkNumber}</TableCell>
                      <TableCell>{formatDate(check.checkDate)}</TableCell>
                      <TableCell>{check.payee}</TableCell>
                      <TableCell className="text-right font-mono">{formatCurrency(check.amount)}</TableCell>
                      <TableCell className="font-mono text-sm">{check.accountNumber}</TableCell>
                      <TableCell>{formatDate(check.reconcileDate)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Reports Tab */}
        <TabsContent value="reports" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <Card>
              <CardHeader>
                <CardTitle>Reconciliation Report</CardTitle>
                <CardDescription>Summary of bank reconciliation</CardDescription>
              </CardHeader>
              <CardContent>
                <Button className="w-full">
                  <Download className="w-4 h-4 mr-2" />
                  Generate Report
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Outstanding Checks Report</CardTitle>
                <CardDescription>List of uncleared checks</CardDescription>
              </CardHeader>
              <CardContent>
                <Button className="w-full">
                  <Download className="w-4 h-4 mr-2" />
                  Generate Report
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Check Register</CardTitle>
                <CardDescription>Complete check register</CardDescription>
              </CardHeader>
              <CardContent>
                <Button className="w-full">
                  <Download className="w-4 h-4 mr-2" />
                  Generate Report
                </Button>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}