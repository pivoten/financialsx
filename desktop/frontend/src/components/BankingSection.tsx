
import { useState, useEffect } from 'react'
import { GetBankAccounts, GetDBFTableData, GetCachedBalances, RefreshAccountBalance, RefreshAllBalances } from '../../wailsjs/go/main/App'
import { getCompanyDataPath, getCompanyName } from '../utils/companyPath'
import logger from '../services/logger'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { BankReconciliation } from './BankReconciliation'
import { CheckAudit } from './CheckAudit'
import OutstandingChecks from './OutstandingChecks'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Badge } from './ui/badge'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './ui/table'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './ui/tabs'
import { 
  DropdownMenu, 
  DropdownMenuContent, 
  DropdownMenuItem, 
  DropdownMenuTrigger,
  DropdownMenuSeparator 
} from './ui/dropdown-menu'
import { 
  CreditCard, 
  Building2, 
  ArrowUpRight, 
  ArrowDownLeft, 
  Plus,
  Search,
  Filter,
  MoreVertical,
  CheckCircle,
  AlertCircle,
  Clock,
  RefreshCw,
  Calculator,
  FileText,
  TrendingUp
} from 'lucide-react'
import type { User, BankAccount } from '../types'

interface BankingSectionProps {
  companyName: string
  currentUser: User
}

interface Transaction {
  id: number
  date: string
  description: string
  type: string
  amount: number
  account: string
  status: string
  reference: string
}

export function BankingSection({ companyName, currentUser }: BankingSectionProps) {
  // State for banking data
  const [accounts, setAccounts] = useState<BankAccount[]>([])
  const [loadingAccounts, setLoadingAccounts] = useState<boolean>(true)
  const [refreshingAccount, setRefreshingAccount] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<string>('accounts')
  const [selectedAccountForReconciliation, setSelectedAccountForReconciliation] = useState<string | null>(null)

  const [transactions, setTransactions] = useState([
    {
      id: 1,
      date: "2024-08-01",
      description: "Revenue Distribution - Well ABC-001",
      type: "Credit",
      amount: 15250.00,
      account: "Primary Operating Account",
      status: "Completed",
      reference: "RD240801001"
    },
    {
      id: 2,
      date: "2024-08-01",
      description: "Owner Payment - Smith Trust",
      type: "Debit",
      amount: -8750.50,
      account: "Primary Operating Account",
      status: "Completed",
      reference: "OP240801002"
    },
    {
      id: 3,
      date: "2024-07-31",
      description: "Lease Operating Expense",
      type: "Debit",
      amount: -2350.25,
      account: "Primary Operating Account",
      status: "Completed",
      reference: "LOE240731001"
    },
    {
      id: 4,
      date: "2024-07-31",
      description: "ACH Transfer to Reserve",
      type: "Debit",
      amount: -25000.00,
      account: "Primary Operating Account",
      status: "Pending",
      reference: "TRF240731001"
    }
  ])

  const [reconciliation, setReconciliation] = useState([
    {
      id: 1,
      month: "July 2024",
      account: "Primary Operating Account",
      statementBalance: 248100.50,
      bookBalance: 245750.25,
      difference: -2350.25,
      status: "Reconciled",
      reconciledBy: "Admin",
      reconciledDate: "2024-08-01"
    },
    {
      id: 2,
      month: "July 2024",
      account: "Reserve Account",
      statementBalance: 125000.00,
      bookBalance: 125000.00,
      difference: 0.00,
      status: "Reconciled",
      reconciledBy: "Admin",
      reconciledDate: "2024-08-01"
    }
  ])

  // Load bank accounts from COA.dbf
  useEffect(() => {
    if (companyName) {
      loadBankAccounts()
    }
  }, [companyName])

  const loadBankAccounts = async () => {
    try {
      setLoadingAccounts(true)
      
      logger.debug('BankingSection: Loading bank accounts', { company: companyName })
      
      let bankAccounts = []
      
      // Check if GetBankAccounts is available (Wails bindings generated)
      if (typeof GetBankAccounts === 'function') {
        logger.debug('BankingSection: GetBankAccounts function available, calling it')
        try {
          // Always use company name directly, not path
          const companyName = localStorage.getItem('company_name')
          if (!companyName) {
            throw new Error('No company selected')
          }
          bankAccounts = await GetBankAccounts(companyName)
          logger.debug('BankingSection: GetBankAccounts response', {
            responseType: typeof bankAccounts,
            length: bankAccounts?.length
          })
          
          if (!bankAccounts || !Array.isArray(bankAccounts)) {
            logger.warn('BankingSection: GetBankAccounts returned invalid data, using fallback')
            throw new Error('GetBankAccounts returned invalid data: ' + typeof bankAccounts)
          }
        } catch (getBankAccountsErr) {
          logger.error('BankingSection: GetBankAccounts call failed', {
            error: getBankAccountsErr.message,
            stack: getBankAccountsErr.stack
          })
          // Fall through to fallback method
          throw getBankAccountsErr
        }
      } else {
        logger.warn('BankingSection: GetBankAccounts function not available')
        throw new Error('GetBankAccounts function not available')
      }
      
      // Transform COA bank accounts to display format
      const transformedAccounts = bankAccounts.map((account, index) => ({
        id: index + 1,
        name: account.account_name,
        accountNumber: account.account_number,
        bank: account.description || "Bank Name", // Use description or fallback
        balance: account.balance || 0,
        type: account.account_type || "Checking",
        status: "Active"
      }))
      
      logger.debug('BankingSection: Transformed accounts', { count: transformedAccounts?.length })
      setAccounts(transformedAccounts as BankAccount[])
      
      // Load GL balances for each account
      await loadAccountBalances(transformedAccounts)
    } catch (err) {
      logger.debug('BankingSection: Primary method failed, trying fallback')
      // Fallback: Try to read COA.dbf directly using existing function
      try {
        logger.debug('BankingSection: Using GetDBFTableData fallback')
        // Always use company name directly
        const companyName = localStorage.getItem('company_name')
        const coaData = await GetDBFTableData(companyName, 'COA.dbf')
        logger.debug('BankingSection: COA.dbf data loaded', { recordCount: coaData?.length })
        
        if (coaData && coaData.rows) {
          const bankAccounts = coaData.rows
            .filter((row: any[], index: number) => {
              // Check if LBANKACCT is true (column 6 based on COA structure)
              const bankFlag = row[6]
              if (index < 5) { // Only log first 5 rows to avoid spam
                logger.debug('BankingSection - Processing row', {
                  index,
                  account: row[0],
                  bankFlag,
                  flagType: typeof bankFlag
                })
              }
              return bankFlag === true || bankFlag === 'T' || bankFlag === '.T.' || bankFlag === 'true'
            })
            .map((row: any[]) => ({
              account_number: row[0] || '',     // Cacctno
              account_name: row[2] || '',       // Cacctdesc (Account description)
              account_type: row[1] || 'Checking', // Caccttype
              balance: 0,                       // Balance not in COA
              description: row[2] || '',        // Cacctdesc
              is_bank_account: true
            }))
          
          logger.debug('BankingSection: Filtered bank accounts', { count: bankAccounts?.length })
          
          // Transform COA bank accounts to display format
          const transformedAccounts = bankAccounts.map((account: any, index: number) => ({
            id: index + 1,
            name: account.account_name,
            accountNumber: account.account_number,
            bank: account.description || "Bank Name", // Use description or fallback
            balance: account.balance || 0,
            type: account.account_type || "Checking",
            status: "Active"
          }))
          
          logger.debug('BankingSection: Transformed accounts (fallback)', { count: transformedAccounts?.length })
          setAccounts(transformedAccounts)
          
          // Load GL balances for each account
          await loadAccountBalances(transformedAccounts)
        } else {
          logger.warn('BankingSection: COA.dbf fallback - no data found', { coaDataStructure: typeof coaData })
          setAccounts([]) // Set empty array for "no accounts found" display
        }
      } catch (fallbackErr) {
        logger.error('BankingSection: Fallback method also failed', { error: fallbackErr.message })
        setAccounts([]) // Set empty array for "no accounts found" display
      }
    } finally {
      setLoadingAccounts(false)
    }
  }

  // Load cached balances for bank accounts (MUCH FASTER!)
  const loadAccountBalances = async (accountList: any[]) => {
    logger.debug('BankingSection: Loading cached balances', { accountCount: accountList?.length })
    
    try {
      // Get all cached balances at once
      // Always use company name directly
      const companyName = localStorage.getItem('company_name')
      const cachedBalances = await GetCachedBalances(companyName)
      logger.debug('BankingSection: Retrieved cached balances', { count: cachedBalances?.length })
      
      // Handle null or undefined response
      if (!cachedBalances || !Array.isArray(cachedBalances)) {
        logger.debug('BankingSection: No cached balances found or invalid response, will refresh all accounts')
        
        // Trigger refresh for all accounts
        for (const account of accountList) {
          RefreshAccountBalance(companyName, account.accountNumber).catch(err => {
            logger.error('Failed to refresh account balance', {
              accountNumber: account.accountNumber,
              error: err.message
            })
          })
        }
        
        // Set accounts with zero balances for now
        const fallbackAccounts = accountList.map((account: any) => ({
          ...account,
          balance: 0,
          bank_balance: 0,
          outstanding_total: 0,
          outstanding_count: 0,
          uncleared_deposits: 0,
          uncleared_checks: 0,
          deposit_count: 0,
          check_count: 0,
          gl_freshness: 'stale',
          checks_freshness: 'stale',
          is_stale: true,
          last_updated: 0
        }))
        setAccounts(fallbackAccounts)
        return
      }
      
      // Create a map for fast lookup
      const balanceMap = new Map()
      cachedBalances.forEach(balance => {
        balanceMap.set(balance.account_number, balance)
      })
      
      // Update accounts with cached balance data
      const updatedAccounts = accountList.map((account: any) => {
        const cachedBalance = balanceMap.get(account.accountNumber)
        
        if (cachedBalance) {
          return {
            ...account,
            balance: cachedBalance.gl_balance,
            bank_balance: cachedBalance.gl_balance + cachedBalance.outstanding_total, // Calculate on-the-fly
            outstanding_total: cachedBalance.outstanding_total,
            outstanding_count: cachedBalance.outstanding_count,
            // New detailed breakdown fields
            uncleared_deposits: cachedBalance.uncleared_deposits || 0,
            uncleared_checks: cachedBalance.uncleared_checks || 0,
            deposit_count: cachedBalance.deposit_count || 0,
            check_count: cachedBalance.check_count || 0,
            gl_freshness: cachedBalance.gl_freshness,
            checks_freshness: cachedBalance.checks_freshness,
            is_stale: cachedBalance.is_stale,
            last_updated: Math.max(
              new Date(cachedBalance.gl_last_updated).getTime(),
              new Date(cachedBalance.checks_last_updated).getTime()
            )
          }
        } else {
          // No cached balance found - trigger a refresh
          logger.debug('BankingSection: No cached balance - will refresh', {
            accountNumber: account.accountNumber
          })
          RefreshAccountBalance(companyName, account.accountNumber).catch(err => {
            logger.error('Failed to refresh account balance', { error: err.message })
          })
          
          return {
            ...account,
            balance: 0,
            bank_balance: 0, // 0 + 0 = 0
            outstanding_total: 0,
            outstanding_count: 0,
            // New detailed breakdown fields (empty)
            uncleared_deposits: 0,
            uncleared_checks: 0,
            deposit_count: 0,
            check_count: 0,
            gl_freshness: 'stale',
            checks_freshness: 'stale',
            is_stale: true,
            last_updated: 0
          }
        }
      })
      
      logger.debug('BankingSection: Updated accounts with cached balances', { count: updatedAccounts?.length })
      setAccounts(updatedAccounts)
      
    } catch (error) {
      logger.error('BankingSection: Failed to load cached balances', { error: error.message })
      // Fallback to setting accounts without balance data
      const fallbackAccounts = accountList.map((account: any) => ({
        ...account,
        balance: 0,
        bank_balance: 0, // 0 + 0 = 0
        outstanding_total: 0,
        outstanding_count: 0,
        gl_freshness: 'stale',
        checks_freshness: 'stale',
        is_stale: true,
        last_updated: 0
      }))
      setAccounts(fallbackAccounts)
    }
  }

  // Refresh balance for a specific account
  const refreshAccountBalance = async (accountNumber: string) => {
    setRefreshingAccount(accountNumber)
    
    try {
      const companyName = localStorage.getItem('company_name')
      
      // Debug logging
      logger.debug('refreshAccountBalance called', {
        accountNumber,
        companyName,
        company_path: localStorage.getItem('company_path'),
        company_name: localStorage.getItem('company_name')
      })
      
      if (!companyName) {
        throw new Error('No company selected. Please select a company.')
      }
      
      if (!accountNumber) {
        throw new Error('No account number provided')
      }
      
      await RefreshAccountBalance(companyName, accountNumber)
      // Reload all account balances to reflect the update
      const currentAccounts = [...accounts]
      await loadAccountBalances(currentAccounts)
    } catch (error) {
      logger.error('Failed to refresh account balance', { error: error.message })
      alert('Failed to refresh balance: ' + (error.message || error))
    } finally {
      setRefreshingAccount(null)
    }
  }

  // Refresh all account balances
  const refreshAllBalances = async () => {
    setRefreshingAccount('all')
    
    try {
      const companyName = localStorage.getItem('company_name')
      
      // Debug logging
      logger.debug('refreshAllBalances called', {
        companyName,
        company_path: localStorage.getItem('company_path'),
        company_name: localStorage.getItem('company_name')
      })
      
      if (!companyName) {
        throw new Error('No company selected. Please select a company.')
      }
      
      await RefreshAllBalances(companyName)
      // Reload all account balances to reflect the updates
      const currentAccounts = [...accounts]
      await loadAccountBalances(currentAccounts)
    } catch (error) {
      logger.error('Failed to refresh all balances', { error: error.message })
      alert('Failed to refresh all balances: ' + (error.message || error))
    } finally {
      setRefreshingAccount(null)
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'Active':
      case 'Completed':
      case 'Reconciled':
        return <Badge variant="default" className="bg-green-100 text-green-800"><CheckCircle className="w-3 h-3 mr-1" />{status}</Badge>
      case 'Pending':
        return <Badge variant="secondary" className="bg-yellow-100 text-yellow-800"><Clock className="w-3 h-3 mr-1" />{status}</Badge>
      case 'Failed':
      case 'Error':
        return <Badge variant="destructive"><AlertCircle className="w-3 h-3 mr-1" />{status}</Badge>
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  return (
    <div className="space-y-6">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList>
          <TabsTrigger value="accounts">Bank Accounts</TabsTrigger>
          <TabsTrigger value="outstanding">Outstanding Checks</TabsTrigger>
          <TabsTrigger value="cleared">Cleared Checks</TabsTrigger>
          <TabsTrigger value="reports">Reports</TabsTrigger>
          {currentUser && (currentUser.is_root || currentUser.role_name === 'Admin') && (
            <TabsTrigger value="audit">Audit</TabsTrigger>
          )}
          {activeTab === 'reconciliation' && (
            <TabsTrigger value="reconciliation">Reconcile</TabsTrigger>
          )}
        </TabsList>

        {/* Bank Accounts Tab */}
        <TabsContent value="accounts" className="space-y-4">
          {/* Header with Refresh All */}
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-lg font-semibold">Bank Accounts</h3>
              <p className="text-sm text-muted-foreground">
                Bank Balance = GL Balance + Uncleared Deposits - Uncleared Checks
              </p>
            </div>
            <Button
              onClick={refreshAllBalances}
              disabled={refreshingAccount === 'all' || loadingAccounts}
              variant="outline"
            >
              <RefreshCw className={`mr-2 h-4 w-4 ${refreshingAccount === 'all' ? 'animate-spin' : ''}`} />
              Refresh All
            </Button>
          </div>

          {/* Account Summary Cards */}
          {loadingAccounts ? (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {[1, 2, 3].map((i) => (
                <Card key={i} className="animate-pulse">
                  <CardHeader className="pb-2">
                    <div className="h-4 bg-gray-300 rounded w-3/4"></div>
                    <div className="h-3 bg-gray-200 rounded w-1/2 mt-2"></div>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      <div className="h-3 bg-gray-200 rounded"></div>
                      <div className="h-3 bg-gray-200 rounded w-2/3"></div>
                      <div className="h-6 bg-gray-300 rounded w-1/2 mt-4"></div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : accounts.length === 0 ? (
            <Card>
              <CardContent className="p-6">
                <div className="text-center text-muted-foreground">
                  <AlertCircle className="w-8 h-8 mx-auto mb-4" />
                  <p>No bank accounts found in Chart of Accounts</p>
                  <p className="text-sm mt-2">Make sure LBANKACCT is set to true for bank accounts in COA.dbf</p>
                </div>
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {accounts.map((account) => (
              <Card key={account.id} className="relative">
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-lg">{account.name}</CardTitle>
                    {getStatusBadge(account.status)}
                  </div>
                  <CardDescription className="flex items-center gap-2">
                    <Building2 className="w-4 h-4" />
                    {account.bank}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">Account Number</span>
                      <span className="text-sm font-mono">{account.accountNumber}</span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">Type</span>
                      <span className="text-sm">{account.type}</span>
                    </div>
                    <div className="pt-2 border-t space-y-2">
                      <div className="flex justify-between items-center">
                        <span className="text-sm font-medium">GL Balance</span>
                        <span className={`text-lg font-bold ${account.balance >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                          {formatCurrency(account.balance)}
                        </span>
                      </div>
                      
                      {/* Show detailed breakdown when available */}
                      {(account.uncleared_deposits > 0 || account.uncleared_checks > 0) && (
                        <div className="space-y-1 text-sm">
                          <div className="flex justify-between items-center">
                            <span className="text-muted-foreground">Uncleared Deposits</span>
                            <span className="text-green-600">
                              {formatCurrency(account.uncleared_deposits || 0)} ({account.deposit_count || 0})
                            </span>
                          </div>
                          <div className="flex justify-between items-center">
                            <span className="text-muted-foreground">Uncleared Checks</span>
                            <span className="text-red-600">
                              {formatCurrency(account.uncleared_checks || 0)} ({account.check_count || 0})
                            </span>
                          </div>
                        </div>
                      )}
                      
                      {/* Show breakdown structure even for old data - needs refresh to get details */}
                      {!(account.uncleared_deposits > 0 || account.uncleared_checks > 0) && account.outstanding_total !== 0 && (
                        <div className="space-y-1 text-sm">
                          <div className="flex justify-between items-center">
                            <span className="text-muted-foreground">Uncleared Deposits</span>
                            <span className="text-gray-400">
                              Refresh for details
                            </span>
                          </div>
                          <div className="flex justify-between items-center">
                            <span className="text-muted-foreground">Uncleared Checks</span>
                            <span className="text-gray-400">
                              Refresh for details
                            </span>
                          </div>
                        </div>
                      )}
                      
                      <div className="flex justify-between items-center border-t pt-2">
                        <span className="text-sm font-medium">Bank Balance</span>
                        <span className={`text-xl font-bold ${(account.bank_balance || account.balance) >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                          {formatCurrency(account.bank_balance || account.balance)}
                        </span>
                      </div>
                      
                      {account.is_stale && (
                        <div className="flex items-center gap-1 text-xs text-amber-600">
                          <AlertCircle className="h-3 w-3" />
                          <span>Data may be stale</span>
                        </div>
                      )}
                    </div>
                  </div>
                  <div className="flex gap-2 mt-4">
                    <Button size="sm" variant="outline" className="flex-1">
                      <ArrowUpRight className="w-4 h-4 mr-1" />
                      Transfer
                    </Button>
                    
                    {account.is_stale ? (
                      <Button 
                        size="sm" 
                        variant="outline"
                        onClick={() => refreshAccountBalance(account.accountNumber)}
                        disabled={refreshingAccount === account.accountNumber}
                        className="text-amber-600 border-amber-200"
                      >
                        <RefreshCw className={`w-4 h-4 ${refreshingAccount === account.accountNumber ? 'animate-spin' : ''}`} />
                      </Button>
                    ) : (
                      <Button 
                        size="sm" 
                        variant="outline"
                        onClick={() => refreshAccountBalance(account.accountNumber)}
                        disabled={refreshingAccount === account.accountNumber}
                      >
                        <RefreshCw className={`w-4 h-4 ${refreshingAccount === account.accountNumber ? 'animate-spin' : ''}`} />
                      </Button>
                    )}
                    
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button size="sm" variant="outline">
                          <MoreVertical className="w-4 h-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem 
                          onClick={() => {
                            setActiveTab('reconciliation')
                            setSelectedAccountForReconciliation(account.accountNumber)
                          }}
                        >
                          <Calculator className="w-4 h-4 mr-2" />
                          Reconcile
                        </DropdownMenuItem>
                        <DropdownMenuItem>
                          <FileText className="w-4 h-4 mr-2" />
                          Statement
                        </DropdownMenuItem>
                        <DropdownMenuItem>
                          <TrendingUp className="w-4 h-4 mr-2" />
                          Analytics
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem>
                          <MoreVertical className="w-4 h-4 mr-2" />
                          More Options
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </CardContent>
              </Card>
            ))}
            </div>
          )}

          {/* Quick Actions */}
          <Card>
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
              <CardDescription>Common banking operations</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-4">
                <Button variant="outline" className="justify-start">
                  <Plus className="w-4 h-4 mr-2" />
                  Add Account
                </Button>
                <Button variant="outline" className="justify-start">
                  <ArrowUpRight className="w-4 h-4 mr-2" />
                  Transfer Funds
                </Button>
                <Button variant="outline" className="justify-start">
                  <CreditCard className="w-4 h-4 mr-2" />
                  Reconcile Account
                </Button>
                <Button variant="outline" className="justify-start">
                  <Filter className="w-4 h-4 mr-2" />
                  Generate Report
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Transactions Tab */}
        <TabsContent value="transactions" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Recent Transactions</CardTitle>
                  <CardDescription>Banking transaction history</CardDescription>
                </div>
                <div className="flex gap-2">
                  <Button variant="outline" size="sm">
                    <Search className="w-4 h-4 mr-2" />
                    Search
                  </Button>
                  <Button variant="outline" size="sm">
                    <Filter className="w-4 h-4 mr-2" />
                    Filter
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Date</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Account</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead className="text-right">Amount</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Reference</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {transactions.map((transaction) => (
                    <TableRow key={transaction.id}>
                      <TableCell className="font-mono text-sm">
                        {transaction.date}
                      </TableCell>
                      <TableCell>{transaction.description}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {transaction.account}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          {transaction.type === 'Credit' ? (
                            <ArrowDownLeft className="w-4 h-4 text-green-600" />
                          ) : (
                            <ArrowUpRight className="w-4 h-4 text-red-600" />
                          )}
                          {transaction.type}
                        </div>
                      </TableCell>
                      <TableCell className={`text-right font-mono ${
                        transaction.amount >= 0 ? 'text-green-600' : 'text-red-600'
                      }`}>
                        {formatCurrency(Math.abs(transaction.amount))}
                      </TableCell>
                      <TableCell>
                        {getStatusBadge(transaction.status)}
                      </TableCell>
                      <TableCell className="font-mono text-sm">
                        {transaction.reference}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Reconciliation Tab */}
        <TabsContent value="reconciliation" className="space-y-4">
          <BankReconciliation 
            companyName={companyName} 
            currentUser={currentUser}
            preSelectedAccount={selectedAccountForReconciliation}
            onBack={() => {
              setActiveTab('accounts')
              setSelectedAccountForReconciliation(null)
            }}
          />
        </TabsContent>

        {/* Transfers Tab */}
        <TabsContent value="transfers" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Transfer Funds</CardTitle>
                <CardDescription>Transfer money between accounts</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid gap-2">
                  <Label htmlFor="from-account">From Account</Label>
                  <select id="from-account" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                    <option>Select account...</option>
                    {accounts.map((account) => (
                      <option key={account.id} value={account.id}>
                        {account.name} - {formatCurrency(account.balance)}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="to-account">To Account</Label>
                  <select id="to-account" className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                    <option>Select account...</option>
                    {accounts.map((account) => (
                      <option key={account.id} value={account.id}>
                        {account.name} - {formatCurrency(account.balance)}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="amount">Amount</Label>
                  <Input 
                    id="amount" 
                    type="number" 
                    placeholder="0.00" 
                    step="0.01"
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="description">Description</Label>
                  <Input 
                    id="description" 
                    placeholder="Transfer description"
                  />
                </div>
                <Button className="w-full">
                  <ArrowUpRight className="w-4 h-4 mr-2" />
                  Initiate Transfer
                </Button>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Transfer History</CardTitle>
                <CardDescription>Recent fund transfers</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex items-center justify-between p-3 border rounded">
                    <div>
                      <div className="font-medium">Operating → Reserve</div>
                      <div className="text-sm text-muted-foreground">Jul 31, 2024</div>
                    </div>
                    <div className="text-right">
                      <div className="font-medium">{formatCurrency(25000.00)}</div>
                      {getStatusBadge('Pending')}
                    </div>
                  </div>
                  <div className="flex items-center justify-between p-3 border rounded">
                    <div>
                      <div className="font-medium">Reserve → Operating</div>
                      <div className="text-sm text-muted-foreground">Jul 15, 2024</div>
                    </div>
                    <div className="text-right">
                      <div className="font-medium">{formatCurrency(10000.00)}</div>
                      {getStatusBadge('Completed')}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Reports Tab */}
        <TabsContent value="reports" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <Card>
              <CardHeader>
                <CardTitle>Account Statements</CardTitle>
                <CardDescription>Generate account statements</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <Button variant="outline" className="w-full justify-start">
                    <CreditCard className="w-4 h-4 mr-2" />
                    Monthly Statement
                  </Button>
                  <Button variant="outline" className="w-full justify-start">
                    <CreditCard className="w-4 h-4 mr-2" />
                    Custom Date Range
                  </Button>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Reconciliation Reports</CardTitle>
                <CardDescription>Bank reconciliation reports</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <Button variant="outline" className="w-full justify-start">
                    <CheckCircle className="w-4 h-4 mr-2" />
                    Reconciliation Summary
                  </Button>
                  <Button variant="outline" className="w-full justify-start">
                    <AlertCircle className="w-4 h-4 mr-2" />
                    Outstanding Items
                  </Button>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Cash Flow Reports</CardTitle>
                <CardDescription>Cash flow analysis and forecasting</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <Button variant="outline" className="w-full justify-start">
                    <ArrowUpRight className="w-4 h-4 mr-2" />
                    Cash Flow Statement
                  </Button>
                  <Button variant="outline" className="w-full justify-start">
                    <ArrowDownLeft className="w-4 h-4 mr-2" />
                    Cash Flow Forecast
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Outstanding Checks Tab */}
        <TabsContent value="outstanding" className="space-y-4">
          <OutstandingChecks companyName={companyName} currentUser={currentUser} />
        </TabsContent>

        {/* Cleared Checks Tab - Placeholder */}
        <TabsContent value="cleared" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Cleared Checks</CardTitle>
              <CardDescription>Checks that have been cleared by the bank</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">This feature will be implemented in a future update.</p>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Audit Tab */}
        {currentUser && (currentUser.is_root || currentUser.role_name === 'Admin') && (
          <TabsContent value="audit" className="space-y-4">
            <CheckAudit companyName={companyName} currentUser={currentUser} />
          </TabsContent>
        )}
      </Tabs>
    </div>
  )
}
