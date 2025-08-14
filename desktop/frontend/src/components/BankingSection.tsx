
import { useState, useEffect } from 'react'
import { getCompanyDataPath, getCompanyName } from '../utils/companyPath'
import logger from '../services/logger'

// Check if Wails is available
const isWailsAvailable = typeof window !== 'undefined' && (window as any).go?.main?.App

// Wails functions - will be loaded dynamically
let WailsFunctions: any = null

// Load Wails functions dynamically
async function loadWailsFunctions() {
  if (isWailsAvailable && !WailsFunctions) {
    try {
      WailsFunctions = await import('../../wailsjs/go/main/App')
    } catch (error) {
      console.error('Failed to load Wails functions:', error)
    }
  }
  return WailsFunctions
}
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  rectSortingStrategy,
} from '@dnd-kit/sortable'
import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
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
import { Select } from './ui/select'
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
  TrendingUp,
  GripVertical,
  Download,
  Eye,
  Edit,
  Copy,
  Trash
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

// Sortable Account Card Component
const SortableAccountCard = ({ account, isRefreshing, onRefresh, onReconcile }: any) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: account.accountNumber })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  const getFreshnessIndicator = (freshness: string) => {
    if (freshness === 'fresh') {
      return <div className="w-2 h-2 bg-green-500 rounded-full" title="Data is fresh" />
    } else if (freshness === 'aging') {
      return <div className="w-2 h-2 bg-yellow-500 rounded-full" title="Data is aging" />
    } else {
      return <div className="w-2 h-2 bg-red-500 rounded-full" title="Data is stale - refresh recommended" />
    }
  }

  return (
    <div ref={setNodeRef} style={style} className="h-full">
      <Card className="relative border border-gray-200 hover:shadow-md transition-all h-full flex flex-col bg-white">
        <CardHeader className="pb-4 border-b border-gray-100">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3 flex-1">
              <div 
                className="cursor-move text-gray-400 hover:text-gray-600 transition-colors"
                {...attributes}
                {...listeners}
              >
                <GripVertical className="w-5 h-5" />
              </div>
              <div className="flex items-center gap-2">
                <CreditCard className="h-5 w-5 text-gray-600" />
                <div>
                  <h4 className="font-semibold text-gray-900">{account.name || 'Unnamed Account'}</h4>
                  <p className="text-xs text-gray-500">Account #{account.accountNumber}</p>
                </div>
              </div>
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8">
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => onRefresh(account.accountNumber)}>
                  <RefreshCw className="mr-2 h-4 w-4" />
                  Refresh
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onReconcile(account)}>
                  <Calculator className="mr-2 h-4 w-4" />
                  Reconcile
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem>
                  <FileText className="mr-2 h-4 w-4" />
                  Statement History
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <TrendingUp className="mr-2 h-4 w-4" />
                  Analytics
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </CardHeader>
        <CardContent className="p-4 flex-1 flex flex-col">
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <span className="text-sm text-gray-500">GL Balance</span>
              <span className={`text-sm font-medium ${account.balance >= 0 ? 'text-gray-900' : 'text-red-600'}`}>
                {formatCurrency(account.balance || 0)}
              </span>
            </div>
            
            <div className="flex justify-between items-center pb-3 border-b border-gray-100">
              <span className="text-sm text-gray-500 flex items-center gap-1">
                Bank Balance
                {getFreshnessIndicator(account.gl_freshness || 'stale')}
              </span>
              <span className={`text-lg font-semibold ${(account.bank_balance || account.balance) >= 0 ? 'text-gray-900' : 'text-red-600'}`}>
                {formatCurrency(account.bank_balance || account.balance || 0)}
              </span>
            </div>
          </div>

          <div className="flex-1 flex flex-col justify-end">
            {(account.uncleared_checks || account.uncleared_deposits || account.outstanding_total) ? (
              <div className="space-y-2 mt-3">
                <div className="flex justify-between items-center">
                  <span className="text-xs text-gray-500">
                    Uncleared Checks {(account.check_count || account.outstanding_count) ? `(${account.check_count || account.outstanding_count})` : ''}
                  </span>
                  <span className="text-xs text-red-600 font-medium">
                    {(account.uncleared_checks || account.outstanding_total) ? 
                      `-${formatCurrency(Math.abs(account.uncleared_checks || account.outstanding_total || 0))}` : 
                      formatCurrency(0)}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-xs text-gray-500">
                    Uncleared Deposits {account.deposit_count ? `(${account.deposit_count})` : ''}
                  </span>
                  <span className="text-xs text-green-600 font-medium">
                    {account.uncleared_deposits ? 
                      `+${formatCurrency(account.uncleared_deposits)}` : 
                      formatCurrency(0)}
                  </span>
                </div>
              </div>
            ) : null}

            {account.is_stale && (
              <div className="flex items-center gap-2 p-2 bg-amber-50 border border-amber-200 rounded text-xs mt-3">
                <AlertCircle className="h-3 w-3 text-amber-600" />
                <span className="text-amber-700">Data may be stale</span>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export function BankingSection({ companyName, currentUser }: BankingSectionProps) {
  // State for banking data
  const [accounts, setAccounts] = useState<BankAccount[]>([])
  const [orderedAccountIds, setOrderedAccountIds] = useState<string[]>([])
  const [loadingAccounts, setLoadingAccounts] = useState<boolean>(true)
  const [refreshingAccount, setRefreshingAccount] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<string>('accounts')
  const [selectedAccountForReconciliation, setSelectedAccountForReconciliation] = useState<string | null>(null)
  const [selectedAccountForRegister, setSelectedAccountForRegister] = useState<string>('')

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

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
      // Load saved order from localStorage
      const savedOrder = localStorage.getItem(`bankAccountOrder_${companyName}`)
      if (savedOrder) {
        try {
          setOrderedAccountIds(JSON.parse(savedOrder))
        } catch (e) {
          logger.error('Failed to parse saved account order', { error: e })
        }
      }
    }
  }, [companyName])

  const loadBankAccounts = async () => {
    try {
      setLoadingAccounts(true)
      
      logger.debug('BankingSection: Loading bank accounts', { company: companyName })
      
      let bankAccounts = []
      
      // Check if Wails is available and load functions
      const funcs = await loadWailsFunctions()
      if (isWailsAvailable && funcs && typeof funcs.GetBankAccounts === 'function') {
        logger.debug('BankingSection: GetBankAccounts function available, calling it')
        try {
          // Always use company name directly, not path
          const companyName = localStorage.getItem('company_name')
          if (!companyName) {
            throw new Error('No company selected')
          }
          bankAccounts = await funcs.GetBankAccounts(companyName)
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
        logger.warn('BankingSection: GetBankAccounts function not available or Wails not available')
        throw new Error('GetBankAccounts function not available or Wails not available')
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
    } catch (err: any) {
      logger.debug('BankingSection: Primary method failed, trying fallback', { error: err.message })
      // Fallback: Try to read COA.dbf directly using existing function
      try {
        logger.debug('BankingSection: Using GetDBFTableData fallback')
        
        // If Wails is not available, use mock data
        if (!isWailsAvailable) {
          logger.debug('BankingSection: Using mock data for browser mode')
          const mockAccounts = [
            {
              id: 1,
              name: 'Primary Operating Account',
              accountNumber: '1001',
              bank: 'First National Bank',
              balance: 125000.00,
              bank_balance: 128500.00,
              outstanding_total: 3500.00,
              outstanding_count: 5,
              uncleared_checks: 3500.00,
              check_count: 5,
              type: 'Checking',
              status: 'Active',
              gl_freshness: 'stale',
              is_stale: true
            },
            {
              id: 2,
              name: 'Reserve Account',
              accountNumber: '1002',
              bank: 'First National Bank',
              balance: 250000.00,
              bank_balance: 250000.00,
              outstanding_total: 0,
              outstanding_count: 0,
              uncleared_checks: 0,
              check_count: 0,
              type: 'Savings',
              status: 'Active',
              gl_freshness: 'fresh',
              is_stale: false
            }
          ]
          setAccounts(mockAccounts as BankAccount[])
          // Initialize order if not set
          if (orderedAccountIds.length === 0) {
            const ids = mockAccounts.map((acc: any) => acc.accountNumber)
            setOrderedAccountIds(ids)
          }
          return
        }
        
        // Always use company name directly
        const companyName = localStorage.getItem('company_name')
        const funcs = await loadWailsFunctions()
        if (!funcs) {
          throw new Error('Wails functions not available')
        }
        const coaData = await funcs.GetDBFTableData(companyName, 'COA.dbf')
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
      } catch (fallbackErr: any) {
        logger.error('BankingSection: Fallback method also failed', { error: fallbackErr.message })
        // In browser mode, use mock data
        if (!isWailsAvailable) {
          logger.debug('BankingSection: Using mock data after fallback failure')
          const mockAccounts = [
            {
              id: 1,
              name: 'Demo Operating Account',
              accountNumber: '1001',
              bank: 'Demo Bank',
              balance: 50000.00,
              bank_balance: 52500.00,
              outstanding_total: 2500.00,
              outstanding_count: 3,
              uncleared_checks: 2500.00,
              check_count: 3,
              type: 'Checking',
              status: 'Active',
              gl_freshness: 'stale',
              is_stale: true
            }
          ]
          setAccounts(mockAccounts as BankAccount[])
          if (orderedAccountIds.length === 0) {
            setOrderedAccountIds(['1001'])
          }
        } else {
          setAccounts([]) // Set empty array for "no accounts found" display
        }
      }
    } finally {
      setLoadingAccounts(false)
    }
  }

  // Load cached balances for bank accounts (MUCH FASTER!)
  const loadAccountBalances = async (accountList: any[]) => {
    logger.debug('BankingSection: Loading cached balances', { accountCount: accountList?.length })
    
    try {
      // Check if Wails is available
      if (!isWailsAvailable) {
        logger.debug('BankingSection: Wails not available, skipping cached balance loading')
        // Just return the accounts as-is in browser mode
        setAccounts(accountList)
        return
      }
      
      // Get all cached balances at once
      // Always use company name directly
      const companyName = localStorage.getItem('company_name')
      const funcs = await loadWailsFunctions()
      if (!funcs) {
        logger.debug('BankingSection: Wails functions not available, skipping cached balance loading')
        return
      }
      const cachedBalances = await funcs.GetCachedBalances(companyName)
      logger.debug('BankingSection: Retrieved cached balances', { count: cachedBalances?.length })
      
      // Handle null or undefined response
      if (!cachedBalances || !Array.isArray(cachedBalances)) {
        logger.debug('BankingSection: No cached balances found or invalid response, will refresh all accounts')
        
        // Trigger refresh for all accounts
        for (const account of accountList) {
          funcs?.RefreshAccountBalance(companyName, account.accountNumber).catch((err: any) => {
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
          funcs?.RefreshAccountBalance(companyName, account.accountNumber).catch((err: any) => {
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
      
      // Initialize order if not set
      if (orderedAccountIds.length === 0) {
        const ids = updatedAccounts.map((acc: any) => acc.accountNumber)
        setOrderedAccountIds(ids)
      }
      
    } catch (error: any) {
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

  // Handle drag end for reordering accounts
  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event

    if (over && active.id !== over.id) {
      const oldIndex = orderedAccountIds.indexOf(active.id as string)
      const newIndex = orderedAccountIds.indexOf(over.id as string)

      const newOrder = arrayMove(orderedAccountIds, oldIndex, newIndex)
      setOrderedAccountIds(newOrder)
      
      // Save to localStorage
      localStorage.setItem(`bankAccountOrder_${companyName}`, JSON.stringify(newOrder))
    }
  }

  // Handle reconcile button click
  const handleReconcile = (account: any) => {
    setSelectedAccountForReconciliation(account.accountNumber)
    setActiveTab('reconciliation')
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
        company_name: localStorage.getItem('company_name'),
        isWailsAvailable
      })
      
      if (!companyName) {
        throw new Error('No company selected. Please select a company.')
      }
      
      if (!accountNumber) {
        throw new Error('No account number provided')
      }
      
      // Check if Wails is available
      if (!isWailsAvailable) {
        logger.warn('Wails not available - refresh not supported in browser mode')
        // In browser mode, just simulate refresh by clearing the flag after a delay
        setTimeout(() => {
          setRefreshingAccount(null)
        }, 1000)
        return
      }
      
      const funcs = await loadWailsFunctions()
      if (!funcs) {
        logger.warn('Wails functions not available')
        return
      }
      await funcs.RefreshAccountBalance(companyName, accountNumber)
      // Reload all account balances to reflect the update
      const currentAccounts = [...accounts]
      await loadAccountBalances(currentAccounts)
    } catch (error: any) {
      logger.error('Failed to refresh account balance', { error: error.message })
      // Only show alert for non-browser mode errors
      if (isWailsAvailable) {
        alert('Failed to refresh balance: ' + (error.message || error))
      }
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
        company_name: localStorage.getItem('company_name'),
        isWailsAvailable
      })
      
      if (!companyName) {
        throw new Error('No company selected. Please select a company.')
      }
      
      // Check if Wails is available
      if (!isWailsAvailable) {
        logger.warn('Wails not available - refresh not supported in browser mode')
        // In browser mode, just simulate refresh by clearing the flag after a delay
        setTimeout(() => {
          setRefreshingAccount(null)
        }, 1500)
        return
      }
      
      const funcs = await loadWailsFunctions()
      if (!funcs) {
        logger.warn('Wails functions not available')
        return
      }
      await funcs.RefreshAllBalances(companyName)
      // Reload all account balances to reflect the updates
      const currentAccounts = [...accounts]
      await loadAccountBalances(currentAccounts)
    } catch (error: any) {
      logger.error('Failed to refresh all balances', { error: error.message })
      // Only show alert for non-browser mode errors
      if (isWailsAvailable) {
        alert('Failed to refresh all balances: ' + (error.message || error))
      }
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
    <div className="bg-white rounded-lg shadow-sm">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <div className="border-b border-gray-200">
          <TabsList className="flex h-12 items-center justify-start space-x-8 px-6 bg-transparent">
            <TabsTrigger 
              value="accounts" 
              className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 data-[state=inactive]:hover:text-gray-700 data-[state=active]:after:absolute data-[state=active]:after:bottom-0 data-[state=active]:after:left-0 data-[state=active]:after:right-0 data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600"
            >
              Bank Accounts
            </TabsTrigger>
            <TabsTrigger 
              value="registers"
              className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 data-[state=inactive]:hover:text-gray-700 data-[state=active]:after:absolute data-[state=active]:after:bottom-0 data-[state=active]:after:left-0 data-[state=active]:after:right-0 data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600"
            >
              Registers
            </TabsTrigger>
            <TabsTrigger 
              value="outstanding"
              className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 data-[state=inactive]:hover:text-gray-700 data-[state=active]:after:absolute data-[state=active]:after:bottom-0 data-[state=active]:after:left-0 data-[state=active]:after:right-0 data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600"
            >
              Outstanding Checks
            </TabsTrigger>
            <TabsTrigger 
              value="cleared"
              className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 data-[state=inactive]:hover:text-gray-700 data-[state=active]:after:absolute data-[state=active]:after:bottom-0 data-[state=active]:after:left-0 data-[state=active]:after:right-0 data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600"
            >
              Cleared Checks
            </TabsTrigger>
            <TabsTrigger 
              value="reports"
              className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 data-[state=inactive]:hover:text-gray-700 data-[state=active]:after:absolute data-[state=active]:after:bottom-0 data-[state=active]:after:left-0 data-[state=active]:after:right-0 data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600"
            >
              Reports
            </TabsTrigger>
            {currentUser && (currentUser.is_root || currentUser.role_name === 'Admin') && (
              <TabsTrigger 
                value="audit"
                className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 data-[state=inactive]:hover:text-gray-700 data-[state=active]:after:absolute data-[state=active]:after:bottom-0 data-[state=active]:after:left-0 data-[state=active]:after:right-0 data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600"
              >
                Audit
              </TabsTrigger>
            )}
            {activeTab === 'reconciliation' && (
              <TabsTrigger 
                value="reconciliation"
                className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 data-[state=inactive]:hover:text-gray-700 data-[state=active]:after:absolute data-[state=active]:after:bottom-0 data-[state=active]:after:left-0 data-[state=active]:after:right-0 data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600"
              >
                Reconcile
              </TabsTrigger>
            )}
          </TabsList>
        </div>

        {/* Bank Accounts Tab */}
        <TabsContent value="accounts" className="p-6">
          {/* Header with Refresh All */}
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Bank Accounts</h2>
              <p className="text-sm text-gray-500 mt-1">
                Bank Balance = GL Balance + Uncleared Deposits - Uncleared Checks
              </p>
            </div>
            <Button
              onClick={refreshAllBalances}
              disabled={refreshingAccount === 'all' || loadingAccounts}
              variant="outline"
              className="border-gray-200 hover:bg-gray-50"
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
            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragEnd={handleDragEnd}
            >
              <SortableContext
                items={orderedAccountIds}
                strategy={rectSortingStrategy}
              >
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                  {orderedAccountIds
                    .map(accountId => accounts.find(acc => acc.accountNumber === accountId))
                    .filter(Boolean)
                    .map((account) => (
                      <SortableAccountCard
                        key={account.accountNumber}
                        account={account}
                        isRefreshing={refreshingAccount === account.accountNumber}
                        onRefresh={() => refreshAccountBalance(account.accountNumber)}
                        onReconcile={() => {
                          setActiveTab('reconciliation')
                          setSelectedAccountForReconciliation(account.accountNumber)
                        }}
                      />
                    ))}
                </div>
              </SortableContext>
            </DndContext>
          )}

          {/* Quick Actions */}
          <div className="mt-8 pt-6 border-t border-gray-100">
            <div className="mb-4">
              <h3 className="text-base font-semibold text-gray-900">Quick Actions</h3>
              <p className="text-sm text-gray-500">Common banking operations</p>
            </div>
            <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-4">
              <Button variant="outline" className="justify-start border-gray-200 hover:bg-gray-50">
                <Plus className="w-4 h-4 mr-2" />
                Add Account
              </Button>
              <Button variant="outline" className="justify-start border-gray-200 hover:bg-gray-50">
                <ArrowUpRight className="w-4 h-4 mr-2" />
                Transfer Funds
              </Button>
              <Button variant="outline" className="justify-start border-gray-200 hover:bg-gray-50">
                <CreditCard className="w-4 h-4 mr-2" />
                Reconcile Account
              </Button>
              <Button variant="outline" className="justify-start border-gray-200 hover:bg-gray-50">
                <Filter className="w-4 h-4 mr-2" />
                Generate Report
              </Button>
            </div>
          </div>
        </TabsContent>

        {/* Registers Tab */}
        <TabsContent value="registers" className="p-6">
          {/* Header with Actions */}
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Bank Registers</h2>
              <p className="text-sm text-gray-500 mt-1">
                View checks and transactions for a specific bank account
              </p>
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                className="border-gray-200 hover:bg-gray-50"
                onClick={() => {/* TODO: Export functionality */}}
              >
                <Download className="mr-2 h-4 w-4" />
                Export
              </Button>
              <Button
                className="bg-blue-600 hover:bg-blue-700 text-white"
                onClick={() => {/* TODO: Add new transaction modal */}}
              >
                <Plus className="mr-2 h-4 w-4" />
                Add Transaction
              </Button>
            </div>
          </div>

          {/* Search and Filter Bar */}
          <div className="flex items-center gap-3 mb-6">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
              <Input
                type="text"
                placeholder="Search register entries..."
                className="pl-10 h-10"
              />
            </div>
            <Select
              className="w-[200px] h-10"
              value={selectedAccountForRegister}
              onChange={(e) => setSelectedAccountForRegister(e.target.value)}
            >
              <option value="">Select Account</option>
              {accounts.map((account) => (
                <option key={account.id} value={account.accountNumber}>
                  {account.name}
                </option>
              ))}
            </Select>
            <Select className="w-[140px] h-10" defaultValue="all">
              <option value="all">All Types</option>
              <option value="credit">Credits</option>
              <option value="debit">Debits</option>
            </Select>
            <Button variant="outline" className="border-gray-200 hover:bg-gray-50 h-10">
              <Filter className="mr-2 h-4 w-4" />
              More Filters
            </Button>
          </div>

          {/* Quick Stats - Only show when account is selected */}
          {selectedAccountForRegister && (
          <div className="grid gap-4 md:grid-cols-4 mb-6">
            <Card className="border border-gray-200 bg-white">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-xs text-gray-500">Total Inflow</p>
                    <p className="text-lg font-semibold text-green-600">+{formatCurrency(45250.00)}</p>
                  </div>
                  <ArrowDownLeft className="h-8 w-8 text-green-100" />
                </div>
              </CardContent>
            </Card>
            <Card className="border border-gray-200 bg-white">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-xs text-gray-500">Total Outflow</p>
                    <p className="text-lg font-semibold text-red-600">-{formatCurrency(36450.75)}</p>
                  </div>
                  <ArrowUpRight className="h-8 w-8 text-red-100" />
                </div>
              </CardContent>
            </Card>
            <Card className="border border-gray-200 bg-white">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-xs text-gray-500">Net Flow</p>
                    <p className="text-lg font-semibold text-gray-900">{formatCurrency(8799.25)}</p>
                  </div>
                  <TrendingUp className="h-8 w-8 text-gray-200" />
                </div>
              </CardContent>
            </Card>
            <Card className="border border-gray-200 bg-white">
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-xs text-gray-500">Pending</p>
                    <p className="text-lg font-semibold text-amber-600">12</p>
                  </div>
                  <Clock className="h-8 w-8 text-amber-100" />
                </div>
              </CardContent>
            </Card>
          </div>
          )}

          {/* Register Table */}
          {selectedAccountForRegister ? (
            <Card className="border border-gray-200 bg-white">
              <CardContent className="p-0">
                <Table>
                <TableHeader>
                  <TableRow className="border-b border-gray-200">
                    <TableHead className="text-gray-600 font-medium">Date</TableHead>
                    <TableHead className="text-gray-600 font-medium">Description</TableHead>
                    <TableHead className="text-gray-600 font-medium">Account</TableHead>
                    <TableHead className="text-gray-600 font-medium">Category</TableHead>
                    <TableHead className="text-gray-600 font-medium text-right">Amount</TableHead>
                    <TableHead className="text-gray-600 font-medium">Status</TableHead>
                    <TableHead className="text-gray-600 font-medium">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {transactions.map((transaction) => (
                    <TableRow key={transaction.id} className="border-b border-gray-100 hover:bg-gray-50">
                      <TableCell className="font-mono text-sm text-gray-600">
                        {transaction.date}
                      </TableCell>
                      <TableCell className="font-medium text-gray-900">
                        {transaction.description}
                        <span className="block text-xs text-gray-500">Ref: {transaction.reference}</span>
                      </TableCell>
                      <TableCell className="text-sm text-gray-600">
                        {transaction.account}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline" className="border-gray-200 text-xs">
                          {transaction.type}
                        </Badge>
                      </TableCell>
                      <TableCell className={`text-right font-mono font-medium ${
                        transaction.amount >= 0 ? 'text-green-600' : 'text-red-600'
                      }`}>
                        {transaction.amount >= 0 ? '+' : '-'}{formatCurrency(Math.abs(transaction.amount))}
                      </TableCell>
                      <TableCell>
                        {getStatusBadge(transaction.status)}
                      </TableCell>
                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-8 w-8">
                              <MoreVertical className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem>
                              <Eye className="mr-2 h-4 w-4" />
                              View Details
                            </DropdownMenuItem>
                            <DropdownMenuItem>
                              <Edit className="mr-2 h-4 w-4" />
                              Edit
                            </DropdownMenuItem>
                            <DropdownMenuItem>
                              <Copy className="mr-2 h-4 w-4" />
                              Duplicate
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem className="text-red-600">
                              <Trash className="mr-2 h-4 w-4" />
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              
              {/* Pagination */}
              <div className="flex items-center justify-between px-6 py-4 border-t border-gray-200">
                <p className="text-sm text-gray-600">
                  Showing 1 to 10 of 156 register entries
                </p>
                <div className="flex gap-2">
                  <Button variant="outline" size="sm" className="border-gray-200">
                    Previous
                  </Button>
                  <Button variant="outline" size="sm" className="border-gray-200 bg-blue-50 text-blue-600">
                    1
                  </Button>
                  <Button variant="outline" size="sm" className="border-gray-200">
                    2
                  </Button>
                  <Button variant="outline" size="sm" className="border-gray-200">
                    3
                  </Button>
                  <Button variant="outline" size="sm" className="border-gray-200">
                    Next
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
          ) : (
            <Card className="border border-gray-200 bg-white">
              <CardContent className="p-6">
                <div className="text-center text-gray-500">
                  <CreditCard className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                  <p className="text-lg font-medium">No Account Selected</p>
                  <p className="text-sm mt-2">Please select a bank account from the dropdown above to view its register</p>
                </div>
              </CardContent>
            </Card>
          )}
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
                  <Select id="from-account" className="h-10">
                    <option>Select account...</option>
                    {accounts.map((account) => (
                      <option key={account.id} value={account.id}>
                        {account.name} - {formatCurrency(account.balance)}
                      </option>
                    ))}
                  </Select>
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="to-account">To Account</Label>
                  <Select id="to-account" className="h-10">
                    <option>Select account...</option>
                    {accounts.map((account) => (
                      <option key={account.id} value={account.id}>
                        {account.name} - {formatCurrency(account.balance)}
                      </option>
                    ))}
                  </Select>
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
