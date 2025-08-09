import { useState, useEffect, useRef, ChangeEvent, FormEvent } from 'react'
import { getCompanyDataPath, getCompanyName } from '../utils/companyPath'
import logger from '../services/logger'
import { 
  GetDBFTableData, 
  UpdateDBFRecord, 
  GetBankAccounts, 
  GetLastReconciliation, 
  GetOutstandingChecks,
  SaveReconciliationDraft,
  GetReconciliationDraft,
  DeleteReconciliationDraft,
  CommitReconciliation,
  ImportBankStatement,
  GetBankTransactions,
  GetRecentBankStatements,
  DeleteBankStatement,
  GetMatchedTransactions,
  UnmatchTransaction,
  RunMatching,
  ClearMatchesAndRerun,
  ManualMatchTransaction
} from '../../wailsjs/go/main/App'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Badge } from './ui/badge'
import { Checkbox } from './ui/checkbox'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './ui/table'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './ui/tabs'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger, DialogDescription } from './ui/dialog'
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
  Calculator,
  ArrowDownLeft,
  ArrowUpRight,
  ChevronUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Plus,
  X,
  FileSpreadsheet,
  Loader2,
  Check as CheckIcon,
  Eye,
  History,
  Trash2,
  ArrowLeft,
  Building2
} from 'lucide-react'
import type {
  BankReconciliationProps,
  Check,
  BankAccount,
  ReconciliationDraft,
  LastReconciliation,
  BankTransaction,
  BankStatement,
  MatchedTransaction,
  CSVParseResult,
  ReconciliationTotals,
  SelectedCheck,
  MatchingOptions
} from '../types/bank-reconciliation'

export function BankReconciliation({ companyName, currentUser, preSelectedAccount, onBack }: BankReconciliationProps) {
  // State management
  const [checks, setChecks] = useState<Check[]>([])
  const [bankAccounts, setBankAccounts] = useState<BankAccount[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingAccounts, setLoadingAccounts] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedAccount, setSelectedAccount] = useState(preSelectedAccount || '')
  const [reconciliationPeriod, setReconciliationPeriod] = useState('')
  const [statementBalance, setStatementBalance] = useState('')
  const [statementDate, setStatementDate] = useState('')
  const [beginningBalance, setBeginningBalance] = useState('')
  const [statementCredits, setStatementCredits] = useState('')
  const [statementDebits, setStatementDebits] = useState('')
  const [selectedChecks, setSelectedChecks] = useState<Set<string>>(new Set())
  const [reconciliationInProgress, setReconciliationInProgress] = useState(false)
  const [lastReconciliation, setLastReconciliation] = useState<LastReconciliation | null>(null)
  const [loadingLastRec, setLoadingLastRec] = useState(false)

  // Draft reconciliation state
  const [draftMode, setDraftMode] = useState(true) // Always start in draft mode
  const [draftReconciliation, setDraftReconciliation] = useState<ReconciliationDraft | null>(null)
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false)
  const [draftSelectedChecks, setDraftSelectedChecks] = useState<Set<string>>(new Set()) // Separate from actual selectedChecks

  // Bank Transactions state (imported via CSV or manual entry)
  const [bankTransactions, setBankTransactions] = useState<BankTransaction[]>([])
  const [loadingBankTransactions, setLoadingBankTransactions] = useState(false)
  const [csvImportOpen, setCsvImportOpen] = useState(false)
  const [csvUploading, setCsvUploading] = useState(false)
  const [csvParseResult, setCsvParseResult] = useState<CSVParseResult | null>(null)
  const [csvMatches, setCsvMatches] = useState<MatchedTransaction[]>([])
  const [csvError, setCsvError] = useState<string | null>(null)
  const [showSideBySide, setShowSideBySide] = useState(false)
  const [showImportHistory, setShowImportHistory] = useState(false)
  const [importHistory, setImportHistory] = useState<BankStatement[]>([])
  const [loadingHistory, setLoadingHistory] = useState(false)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [importToDelete, setImportToDelete] = useState<string | null>(null)

  // Manual matching state
  const [selectedBankTxn, setSelectedBankTxn] = useState<BankTransaction | null>(null)
  const [selectedCheckForMatch, setSelectedCheckForMatch] = useState<Check | null>(null)
  const [selectedChecksForMatch, setSelectedChecksForMatch] = useState<Set<string>>(new Set()) // Multiple check selection
  const [isManualMatching, setIsManualMatching] = useState(false)

  // Filter state
  const [showCleared, setShowCleared] = useState(false)
  const [showUncleared, setShowUncleared] = useState(true)
  const [showTransactionType, setShowTransactionType] = useState<'all' | 'debits' | 'credits'>('all') // 'all', 'debits', 'credits'
  const [limitToStatementDate, setLimitToStatementDate] = useState(false) // Default to showing all checks
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [amountFrom, setAmountFrom] = useState('')
  const [amountTo, setAmountTo] = useState('')

  // Sort state for checks
  const [sortField, setSortField] = useState('checkDate')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc')

  // Sort state for bank transactions
  const [bankSortField, setBankSortField] = useState('transaction_date')
  const [bankSortDirection, setBankSortDirection] = useState<'asc' | 'desc'>('asc')

  // Sort state for checks in manual matching
  const [checkSortField, setCheckSortField] = useState('checkNumber')
  const [checkSortDirection, setCheckSortDirection] = useState<'asc' | 'desc'>('asc')

  // Matched transactions state
  const [matchedTransactions, setMatchedTransactions] = useState<MatchedTransaction[]>([])
  const [loadingMatched, setLoadingMatched] = useState(false)
  const [matchedSearchTerm, setMatchedSearchTerm] = useState('')
  const [selectedMatchedTxns, setSelectedMatchedTxns] = useState<Set<string>>(new Set())
  const [isBulkUnmatching, setIsBulkUnmatching] = useState(false)

  // Matching process state
  const [isRunningMatch, setIsRunningMatch] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [matchResult, setMatchResult] = useState<any>(null)
  const [showMatchingOptions, setShowMatchingOptions] = useState(false)
  const [matchingDateOption, setMatchingDateOption] = useState<'all' | 'statement'>('all') // 'all' or 'statement'

  // Refs for auto-save debouncing
  const saveTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const isCalculatingRef = useRef(false)

  // Auto-save effect with 10-second debounce
  useEffect(() => {
    if (hasUnsavedChanges && draftMode && selectedAccount) {
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current)
      }
      
      saveTimeoutRef.current = setTimeout(() => {
        saveDraftReconciliation()
      }, 10000) // 10 seconds debounce
    }

    return () => {
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current)
      }
    }
  }, [hasUnsavedChanges, draftMode, selectedAccount, beginningBalance, statementBalance, 
      statementCredits, statementDebits, statementDate, draftSelectedChecks])

  // Clear the draft when switching accounts
  useEffect(() => {
    if (selectedAccount && draftMode) {
      loadDraftReconciliation()
    }
  }, [selectedAccount])

  // Load bank accounts on mount
  useEffect(() => {
    loadBankAccounts()
  }, [companyName])

  // Load checks and reconciliation data when account is selected
  useEffect(() => {
    if (selectedAccount) {
      loadChecksData()
      loadLastReconciliation()
      loadBankTransactions()
      loadMatchedTransactions()
      if (draftMode) {
        loadDraftReconciliation()
      }
    }
  }, [selectedAccount])

  // Pre-populate next statement date when component loads
  useEffect(() => {
    if (!statementDate && lastReconciliation?.statement_date) {
      // Calculate next statement date (end of following month)
      const lastDate = new Date(lastReconciliation.statement_date)
      const nextMonth = new Date(lastDate.getFullYear(), lastDate.getMonth() + 2, 0) // Last day of next month
      const formattedDate = nextMonth.toISOString().split('T')[0]
      setStatementDate(formattedDate)
      setHasUnsavedChanges(true)
    } else if (!statementDate) {
      // No previous reconciliation, use end of current month
      const now = new Date()
      const endOfMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0)
      const formattedDate = endOfMonth.toISOString().split('T')[0]
      setStatementDate(formattedDate)
      setHasUnsavedChanges(true)
    }
  }, [lastReconciliation])

  // Update beginning balance from last reconciliation
  useEffect(() => {
    if (lastReconciliation?.ending_balance !== undefined && !beginningBalance) {
      setBeginningBalance(lastReconciliation.ending_balance.toString())
      setHasUnsavedChanges(true)
    }
  }, [lastReconciliation])

  const loadBankAccounts = async () => {
    setLoadingAccounts(true)
    try {
      logger.debug('Loading bank accounts', { company: companyName })
      
      // Try getting bank accounts first
      const accounts = await GetBankAccounts(companyName)
      logger.debug('Bank accounts loaded', { count: accounts?.length })
      
      if (accounts && accounts.length > 0) {
        setBankAccounts(accounts as BankAccount[])
        
        // If pre-selected account exists, use it
        if (preSelectedAccount) {
          setSelectedAccount(preSelectedAccount)
        }
      } else {
        // Fallback to reading COA.dbf directly
        logger.debug('No bank accounts from GetBankAccounts, trying COA.dbf directly')
        const coaData = await GetDBFTableData(companyName, 'COA')
        logger.debug('COA.dbf data loaded', { recordCount: coaData?.length })
        
        if (coaData && coaData.rows) {
          const bankAccounts = coaData.rows.filter((row: any[]) => {
            const isBankAccount = row[6] // LBANKACCT column
            return isBankAccount === true || isBankAccount === 'T' || isBankAccount === '.T.'
          }).map((row: any[]) => ({
            account_number: row[0], // CACCTNO
            account_name: row[2] || row[0], // CACCTDESC or fallback to CACCTNO
            account_type: row[1], // NACCTTYPE
            balance: 0
          }))
          
          logger.debug('Filtered bank accounts', { count: bankAccounts?.length })
          setBankAccounts(bankAccounts)
          
          // If pre-selected account exists, use it
          if (preSelectedAccount) {
            setSelectedAccount(preSelectedAccount)
          }
        } else {
          logger.error('COA.dbf data structure invalid', { coaData })
          setError('Could not load bank accounts')
        }
      }
    } catch (err) {
      logger.error('Error loading bank accounts', { error: err.message })
      setError('Failed to load bank accounts: ' + (err as Error).message)
    } finally {
      setLoadingAccounts(false)
    }
  }

  const loadChecksData = async () => {
    setLoading(true)
    setError(null)
    try {
      logger.debug('Loading checks for account', { account: selectedAccount })
      
      // Load outstanding checks for this account
      const result = await GetOutstandingChecks(companyName, selectedAccount)
      logger.debug('Outstanding checks loaded', { count: result?.checks?.length })
      
      if (result && result.rows) {
        const checksData = result.rows.map((row: any[], index: number) => {
          const checkNumber = row[result.columns.indexOf('CCHECKNO')] || ''
          const checkDate = row[result.columns.indexOf('DCHECKDATE')] || ''
          const payee = row[result.columns.indexOf('CPAYEE')] || ''
          const amount = parseFloat(row[result.columns.indexOf('NAMOUNT')] || 0)
          const cleared = row[result.columns.indexOf('LCLEARED')] || false
          const voidFlag = row[result.columns.indexOf('LVOID')] || false
          const accountNumber = row[result.columns.indexOf('CACCTNO')] || ''
          const cidchec = row[result.columns.indexOf('CIDCHEC')] || `${checkNumber}-${index}`
          const entryType = row[result.columns.indexOf('CENTRYTYPE')] || 'W'
          
          // Calculate days outstanding
          let daysOutstanding = 0
          if (checkDate) {
            const checkDateObj = new Date(checkDate)
            const today = new Date()
            daysOutstanding = Math.floor((today.getTime() - checkDateObj.getTime()) / (1000 * 60 * 60 * 24))
          }
          
          return {
            id: cidchec,
            cidchec,
            checkNumber,
            checkDate,
            payee,
            amount,
            cleared: cleared === true || cleared === 'T' || cleared === '.T.',
            void: voidFlag === true || voidFlag === 'T' || voidFlag === '.T.',
            accountNumber,
            daysOutstanding,
            rowIndex: index,
            type: entryType === 'D' ? 'deposit' : 'check'
          } as Check
        })
        
        logger.debug('Processed checks data', { count: checksData?.length })
        setChecks(checksData)
      } else {
        logger.debug('No checks data found')
        setChecks([])
      }
    } catch (err) {
      logger.error('Error loading checks', { error: err.message })
      setError('Failed to load checks: ' + (err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  const loadLastReconciliation = async () => {
    setLoadingLastRec(true)
    try {
      const result = await GetLastReconciliation(companyName, selectedAccount)
      logger.debug('Last reconciliation loaded', { hasResult: !!result })
      
      if (result && result.id) {
        setLastReconciliation(result as LastReconciliation)
      }
    } catch (err) {
      logger.error('Error loading last reconciliation', { error: err.message })
      // Not critical - might be first reconciliation
    } finally {
      setLoadingLastRec(false)
    }
  }

  const saveDraftReconciliation = async () => {
    if (!selectedAccount) return
    
    try {
      // Prepare selected checks with full details
      const selectedChecksDetails: SelectedCheck[] = Array.from(draftSelectedChecks).map(checkId => {
        const check = checks.find(c => c.id === checkId)
        if (!check) return null
        return {
          cidchec: check.cidchec || check.id,
          checkNumber: check.checkNumber,
          amount: check.amount,
          payee: check.payee,
          checkDate: check.checkDate,
          rowIndex: check.rowIndex
        }
      }).filter(Boolean) as SelectedCheck[]
      
      const draftData: ReconciliationDraft = {
        company_name: companyName,
        account_number: selectedAccount,
        statement_date: statementDate,
        beginning_balance: parseFloat(beginningBalance || '0'),
        statement_balance: parseFloat(statementBalance || '0'),
        statement_credits: parseFloat(statementCredits || '0'),
        statement_debits: parseFloat(statementDebits || '0'),
        selected_checks: selectedChecksDetails,
        status: 'draft'
      }
      
      logger.debug('Saving draft reconciliation', { draftData })
      const result = await SaveReconciliationDraft(companyName, draftData as any)
      logger.debug('Draft saved successfully', { result })
      
      setHasUnsavedChanges(false)
      setDraftReconciliation(result as ReconciliationDraft)
    } catch (err) {
      logger.error('Error saving draft', { error: err.message })
    }
  }

  const loadDraftReconciliation = async () => {
    if (!selectedAccount) return
    
    try {
      logger.debug('Loading draft for account', { account: selectedAccount })
      const result = await GetReconciliationDraft(companyName, selectedAccount)
      logger.debug('Draft loaded successfully', { hasDraft: !!result })
      
      if (result && result.id) {
        const draft = result as ReconciliationDraft
        setDraftReconciliation(draft)
        
        // Populate form fields from draft
        setStatementDate(draft.statement_date || '')
        setBeginningBalance(draft.beginning_balance?.toString() || '')
        setStatementBalance(draft.statement_balance?.toString() || '')
        setStatementCredits(draft.statement_credits?.toString() || '')
        setStatementDebits(draft.statement_debits?.toString() || '')
        
        // Restore selected checks by CIDCHEC
        const selectedCheckIds = new Set<string>()
        if (draft.selected_checks && Array.isArray(draft.selected_checks)) {
          draft.selected_checks.forEach((savedCheck: SelectedCheck) => {
            // Find check by CIDCHEC first, then fall back to check number
            const check = checks.find(c => 
              c.cidchec === savedCheck.cidchec || 
              c.checkNumber === savedCheck.checkNumber
            )
            if (check) {
              selectedCheckIds.add(check.id)
            }
          })
        }
        setDraftSelectedChecks(selectedCheckIds)
        setSelectedChecks(selectedCheckIds)
        
        setHasUnsavedChanges(false)
      }
    } catch (err) {
      logger.error('Error loading draft', { error: err.message })
      // Not critical - might not have a draft
    }
  }

  const clearDraftReconciliation = async () => {
    if (!selectedAccount) return
    
    try {
      await DeleteReconciliationDraft(companyName, selectedAccount)
      setDraftReconciliation(null)
      setDraftSelectedChecks(new Set())
      setSelectedChecks(new Set())
      setBeginningBalance('')
      setStatementBalance('')
      setStatementCredits('')
      setStatementDebits('')
      setStatementDate('')
      setHasUnsavedChanges(false)
    } catch (err) {
      logger.error('Error clearing draft', { error: err.message })
    }
  }

  const calculateEndingBalance = () => {
    // Prevent recursive calls during calculation
    if (isCalculatingRef.current) return
    isCalculatingRef.current = true
    
    const beginning = parseFloat(beginningBalance || '0')
    const credits = parseFloat(statementCredits || '0')
    const debits = parseFloat(statementDebits || '0')
    
    const ending = beginning + credits - debits
    
    // Only update if the value actually changed
    const newValue = ending.toFixed(2)
    if (statementBalance !== newValue) {
      setStatementBalance(newValue)
      setHasUnsavedChanges(true)
    }
    
    isCalculatingRef.current = false
  }

  const sortChecks = (a: Check, b: Check): number => {
    let aValue: any = a[sortField as keyof Check]
    let bValue: any = b[sortField as keyof Check]
    
    // Handle date sorting
    if (sortField === 'checkDate') {
      aValue = new Date(aValue as string).getTime()
      bValue = new Date(bValue as string).getTime()
    }
    
    // Handle numeric sorting
    if (sortField === 'amount' || sortField === 'daysOutstanding') {
      aValue = parseFloat(aValue as string) || 0
      bValue = parseFloat(bValue as string) || 0
    }
    
    // Handle string sorting
    if (typeof aValue === 'string') {
      aValue = aValue.toLowerCase()
      bValue = (bValue as string).toLowerCase()
    }
    
    if (sortDirection === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  }

  const handleSort = (field: string) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortDirection('asc')
    }
  }

  const handleBankSort = (field: string) => {
    if (bankSortField === field) {
      setBankSortDirection(bankSortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setBankSortField(field)
      setBankSortDirection('asc')
    }
  }

  const handleCheckSort = (field: string) => {
    if (checkSortField === field) {
      setCheckSortDirection(checkSortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setCheckSortField(field)
      setCheckSortDirection('asc')
    }
  }

  const renderSortableHeader = (field: string, label: string, className = '') => (
    <TableHead 
      className={`cursor-pointer hover:bg-gray-50 ${className}`}
      onClick={() => handleSort(field)}
    >
      <div className="flex items-center gap-1">
        {label}
        {sortField === field && (
          sortDirection === 'asc' ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />
        )}
      </div>
    </TableHead>
  )

  const filteredChecks = checks
    .filter(check => {
      // Apply cleared/uncleared filter
      if (!showCleared && check.cleared) return false
      if (!showUncleared && !check.cleared) return false
      
      // Apply transaction type filter
      if (showTransactionType === 'debits' && check.type === 'deposit') return false
      if (showTransactionType === 'credits' && check.type !== 'deposit') return false
      
      // Apply date range filter
      if (dateFrom && new Date(check.checkDate) < new Date(dateFrom)) return false
      if (dateTo && new Date(check.checkDate) > new Date(dateTo)) return false
      
      // Apply amount range filter
      if (amountFrom && check.amount < parseFloat(amountFrom)) return false
      if (amountTo && check.amount > parseFloat(amountTo)) return false
      
      return true
    })
    .sort(sortChecks)

  const calculateTotals = (): ReconciliationTotals => {
    const selectedChecksList = Array.from(draftSelectedChecks).map(id => 
      checks.find(c => c.id === id)
    ).filter(Boolean) as Check[]
    
    const credits = selectedChecksList
      .filter(c => c.type === 'deposit')
      .reduce((sum, c) => sum + c.amount, 0)
    
    const debits = selectedChecksList
      .filter(c => c.type !== 'deposit')
      .reduce((sum, c) => sum + c.amount, 0)
    
    const beginning = parseFloat(beginningBalance || '0')
    const calculatedBalance = beginning + credits - debits
    const statementBal = parseFloat(statementBalance || '0')
    const difference = statementBal - calculatedBalance
    
    return {
      statementCredits: credits,
      statementDebits: debits,
      calculatedBalance,
      balanceDifference: difference,
      selectedCheckCount: selectedChecksList.filter(c => c.type !== 'deposit').length,
      selectedDepositCount: selectedChecksList.filter(c => c.type === 'deposit').length,
      // Additional properties expected by the UI
      beginningBalance: beginning,
      selectedCredits: credits,
      selectedDebits: debits,
      statementBalance: statementBal,
      isInBalance: Math.abs(difference) < 0.01,
      reconciliationDifference: difference,
      selectedCount: selectedChecksList.length
    }
  }

  const totals = calculateTotals()

  const toggleCheckSelection = (checkId: string) => {
    const newSelected = new Set(draftSelectedChecks)
    if (newSelected.has(checkId)) {
      newSelected.delete(checkId)
    } else {
      newSelected.add(checkId)
    }
    setDraftSelectedChecks(newSelected)
    setSelectedChecks(newSelected)
    setHasUnsavedChanges(true)
  }

  const handleUnmatchTransaction = async (bankTxnId: string) => {
    try {
      const result = await UnmatchTransaction(parseInt(bankTxnId))
      logger.debug('Unmatch result', { result })
      
      // Reload both bank transactions and matched transactions
      await loadBankTransactions()
      await loadMatchedTransactions()
    } catch (err) {
      logger.error('Error unmatching transaction', { error: err.message })
    }
  }

  const handleManualMatch = async () => {
    if (!selectedBankTxn || !selectedCheckForMatch) return
    
    try {
      const matchData = {
        bank_transaction_id: parseInt(String(selectedBankTxn.transaction_id || '0')),
        check_id: selectedCheckForMatch.cidchec || selectedCheckForMatch.id,
        confidence_score: 100 // Manual match has 100% confidence
      }
      
      const accountNum: number = Number(selectedAccount) || 0
      const result = await (ManualMatchTransaction as any)(companyName, accountNum, matchData)
      logger.debug('Manual match result', { result })
      
      // Reset selection
      setSelectedBankTxn(null)
      setSelectedCheckForMatch(null)
      setIsManualMatching(false)
      
      // Reload transactions
      await loadBankTransactions()
      await loadMatchedTransactions()
    } catch (err) {
      logger.error('Error matching transaction', { error: err.message })
    }
  }

  const getUnmatchedBankTransactions = () => {
    // Filter out already matched transactions
    return bankTransactions.filter(txn => !txn.matched_check_id)
  }

  const getAvailableChecks = () => {
    // Get all checks that haven't been matched yet
    const matchedCheckIds = new Set(
      matchedTransactions
        .filter(m => m.matched_check)
        .map(m => m.matched_check!.cidchec || m.matched_check!.id)
    )
    
    let availableChecks = checks.filter(check => 
      !matchedCheckIds.has(check.cidchec || check.id) && !check.cleared
    )
    
    // Apply statement date filter if enabled
    if (limitToStatementDate && statementDate) {
      const stmtDate = new Date(statementDate)
      availableChecks = availableChecks.filter(check => 
        new Date(check.checkDate) <= stmtDate
      )
    }
    
    // Sort checks for manual matching
    return availableChecks.sort((a, b) => {
      let aValue: any = a[checkSortField as keyof Check]
      let bValue: any = b[checkSortField as keyof Check]
      
      // Handle date sorting
      if (checkSortField === 'checkDate') {
        aValue = new Date(aValue as string).getTime()
        bValue = new Date(bValue as string).getTime()
      }
      
      // Handle numeric sorting
      if (checkSortField === 'amount' || checkSortField === 'daysOutstanding') {
        aValue = parseFloat(aValue as string) || 0
        bValue = parseFloat(bValue as string) || 0
      }
      
      // Handle string sorting
      if (typeof aValue === 'string') {
        aValue = aValue.toLowerCase()
        bValue = (bValue as string).toLowerCase()
      }
      
      if (checkSortDirection === 'asc') {
        return aValue > bValue ? 1 : -1
      } else {
        return aValue < bValue ? 1 : -1
      }
    })
  }

  const commitReconciliation = async () => {
    if (!selectedAccount) return
    
    setReconciliationInProgress(true)
    try {
      // Save the current draft first
      await saveDraftReconciliation()
      
      // Commit the reconciliation
      const result = await CommitReconciliation(companyName, selectedAccount)
      logger.debug('Reconciliation committed successfully', { result })
      
      // Clear the draft
      await clearDraftReconciliation()
      
      // Reload last reconciliation
      await loadLastReconciliation()
      
      // Reset form
      setDraftMode(true)
      setReconciliationInProgress(false)
      
      alert('Reconciliation committed successfully!')
    } catch (err) {
      logger.error('Error committing reconciliation', { error: err.message })
      alert('Failed to commit reconciliation: ' + (err as Error).message)
      setReconciliationInProgress(false)
    }
  }

  const bulkSelectChecks = () => {
    const newSelected = new Set(draftSelectedChecks)
    filteredChecks.forEach(check => {
      newSelected.add(check.id)
    })
    setDraftSelectedChecks(newSelected)
    setSelectedChecks(newSelected)
    setHasUnsavedChanges(true)
  }

  const loadBankTransactions = async () => {
    if (!selectedAccount) return
    
    setLoadingBankTransactions(true)
    try {
      const result = await GetBankTransactions(companyName, selectedAccount, '')
      logger.debug('Bank transactions loaded', { count: result?.length })
      
      if (result && result.transactions) {
        setBankTransactions(result.transactions)
      }
    } catch (err) {
      logger.error('Error loading bank transactions', { error: err.message })
    } finally {
      setLoadingBankTransactions(false)
    }
  }

  const loadMatchedTransactions = async () => {
    if (!selectedAccount) return
    
    setLoadingMatched(true)
    try {
      const result = await GetMatchedTransactions(companyName, selectedAccount)
      logger.debug('Matched transactions loaded', { count: result?.length })
      
      if (result && result.matches) {
        setMatchedTransactions(result.matches)
      }
    } catch (err) {
      logger.error('Error loading matched transactions', { error: err.message })
    } finally {
      setLoadingMatched(false)
    }
  }

  const handleUnmatch = async (checkData: any) => {
    try {
      const bankTxnId = checkData.bank_transaction?.transaction_id || checkData.transaction_id
      if (bankTxnId) {
        await handleUnmatchTransaction(bankTxnId)
      }
    } catch (err) {
      logger.error('Error unmatching', { error: err.message })
    }
  }

  const handleBulkUnmatch = async () => {
    setIsBulkUnmatching(true)
    try {
      for (const txnId of selectedMatchedTxns) {
        await UnmatchTransaction(parseInt(txnId))
      }
      
      // Clear selection and reload
      setSelectedMatchedTxns(new Set())
      await loadBankTransactions()
      await loadMatchedTransactions()
    } catch (err) {
      logger.error('Error bulk unmatching', { error: err.message })
    } finally {
      setIsBulkUnmatching(false)
    }
  }

  const toggleMatchedSelection = (transactionId: string) => {
    const newSelected = new Set(selectedMatchedTxns)
    if (newSelected.has(transactionId)) {
      newSelected.delete(transactionId)
    } else {
      newSelected.add(transactionId)
    }
    setSelectedMatchedTxns(newSelected)
  }

  const toggleSelectAllMatched = (checked: boolean) => {
    if (checked) {
      const allIds = new Set(
        matchedTransactions
          .filter(m => m.bank_transaction?.transaction_id)
          .map(m => m.bank_transaction!.transaction_id!)
      )
      setSelectedMatchedTxns(allIds)
    } else {
      setSelectedMatchedTxns(new Set())
    }
  }

  const loadImportHistory = async () => {
    setLoadingHistory(true)
    try {
      const result = await GetRecentBankStatements(companyName, selectedAccount)
      if (result && Array.isArray(result)) {
        setImportHistory(result as BankStatement[])
      }
    } catch (err) {
      logger.error('Error loading import history', { error: err.message })
    } finally {
      setLoadingHistory(false)
    }
  }

  const handleDeleteImport = (importBatchId: string) => {
    setImportToDelete(importBatchId)
    setDeleteConfirmOpen(true)
  }

  const confirmDelete = async () => {
    if (!importToDelete) return
    
    try {
      await DeleteBankStatement(companyName, importToDelete)
      
      // Reload data
      await loadImportHistory()
      await loadBankTransactions()
      await loadMatchedTransactions()
      
      setDeleteConfirmOpen(false)
      setImportToDelete(null)
    } catch (err) {
      logger.error('Error deleting import', { error: err.message })
      alert('Failed to delete import: ' + (err as Error).message)
    }
  }

  const handleCSVUpload = async (file: File) => {
    setCsvUploading(true)
    setCsvError(null)
    
    try {
      // Read file content
      const content = await file.text()
      
      // Import the bank statement
      const result = await ImportBankStatement(content, file.name, selectedAccount)
      logger.debug('CSV import completed', { result })
      
      if (result && result.success) {
        setCsvParseResult(result as CSVParseResult)
        setCsvMatches(result.matches || [])
        
        // Reload bank transactions and matched transactions
        await loadBankTransactions()
        await loadMatchedTransactions()
      } else {
        setCsvError(result?.error || 'Failed to import CSV')
      }
    } catch (err) {
      logger.error('Error uploading CSV', { error: err.message })
      setCsvError((err as Error).message)
    } finally {
      setCsvUploading(false)
    }
  }

  const handleRunMatching = async (skipDialog = false) => {
    if (!skipDialog) {
      setShowMatchingOptions(true)
      return
    }
    
    setIsRunningMatch(true)
    try {
      const options: MatchingOptions = {
        limitToStatementDate: matchingDateOption === 'statement',
        statementDate: matchingDateOption === 'statement' ? statementDate : undefined
      }
      
      const result = await RunMatching(companyName, selectedAccount, options as any)
      logger.debug('Matching completed', { result })
      
      setMatchResult(result)
      
      // Reload matched transactions
      await loadMatchedTransactions()
      await loadBankTransactions()
    } catch (err) {
      logger.error('Error running matching', { error: err.message })
      alert('Failed to run matching: ' + (err as Error).message)
    } finally {
      setIsRunningMatch(false)
      setShowMatchingOptions(false)
    }
  }

  const handleRefreshMatching = async (skipDialog = false) => {
    if (!skipDialog) {
      setShowMatchingOptions(true)
      return
    }
    
    setIsRefreshing(true)
    try {
      const options: MatchingOptions = {
        limitToStatementDate: matchingDateOption === 'statement',
        statementDate: matchingDateOption === 'statement' ? statementDate : undefined
      }
      
      const result = await ClearMatchesAndRerun(companyName, selectedAccount, options as any)
      logger.debug('Refresh completed', { result })
      
      setMatchResult(result)
      
      // Reload matched transactions
      await loadMatchedTransactions()
      await loadBankTransactions()
    } catch (err) {
      logger.error('Error refreshing matches', { error: err.message })
      alert('Failed to refresh matches: ' + (err as Error).message)
    } finally {
      setIsRefreshing(false)
      setShowMatchingOptions(false)
    }
  }

  const handleConfirmCSVMatches = async () => {
    // Close the CSV import dialog
    setCsvImportOpen(false)
    setCsvParseResult(null)
    
    // Reload data to show the imported and matched transactions
    await loadBankTransactions()
    await loadMatchedTransactions()
  }

  const toggleCSVMatch = (matchIndex: number) => {
    // Toggle individual match selection in CSV preview
    // This would update the matches before confirmation
  }

  const formatCurrency = (amount: number | string) => {
    const num = typeof amount === 'string' ? parseFloat(amount) : amount
    if (isNaN(num)) return '$0.00'
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(num)
  }

  const formatDate = (dateStr: string) => {
    if (!dateStr) return ''
    try {
      const date = new Date(dateStr)
      // Check if date is valid
      if (isNaN(date.getTime())) {
        // Try to parse different formats
        const parts = dateStr.split(/[-/]/)
        if (parts.length === 3) {
          // Assume MM/DD/YYYY or MM-DD-YYYY
          const month = parseInt(parts[0]) - 1
          const day = parseInt(parts[1])
          const year = parseInt(parts[2])
          const parsedDate = new Date(year, month, day)
          if (!isNaN(parsedDate.getTime())) {
            return parsedDate.toLocaleDateString('en-US', {
              year: 'numeric',
              month: '2-digit',
              day: '2-digit'
            })
          }
        }
        return dateStr // Return original if can't parse
      }
      return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
      })
    } catch (err) {
      return dateStr
    }
  }

  const formatNumberForInput = (value: string) => {
    // Remove non-numeric characters except decimal point
    const cleaned = value.replace(/[^0-9.-]/g, '')
    // Ensure only one decimal point
    const parts = cleaned.split('.')
    if (parts.length > 2) {
      return parts[0] + '.' + parts.slice(1).join('')
    }
    return cleaned
  }

  const parseInputNumber = (value: string) => {
    const num = parseFloat(value.replace(/[^0-9.-]/g, ''))
    return isNaN(num) ? 0 : num
  }

  const formatTransactionType = (entryType: string) => {
    switch(entryType) {
      case 'D':
        return 'Deposit'
      case 'W':
        return 'Check'
      case 'C':
        return 'Check'
      default:
        return entryType
    }
  }

  const getTransactionTypeBadge = (entryType: string) => {
    const isDeposit = entryType === 'D' || entryType === 'deposit'
    return (
      <Badge 
        variant={isDeposit ? "default" : "secondary"}
        className={isDeposit ? "bg-green-100 text-green-800" : "bg-blue-100 text-blue-800"}
      >
        {isDeposit ? (
          <>
            <ArrowDownLeft className="w-3 h-3 mr-1" />
            Deposit
          </>
        ) : (
          <>
            <ArrowUpRight className="w-3 h-3 mr-1" />
            Check
          </>
        )}
      </Badge>
    )
  }

  const getStatusBadge = (cleared: boolean) => {
    return cleared ? (
      <Badge variant="default" className="bg-green-100 text-green-800">
        <CheckCircle className="w-3 h-3 mr-1" />
        Cleared
      </Badge>
    ) : (
      <Badge variant="secondary" className="bg-yellow-100 text-yellow-800">
        <Clock className="w-3 h-3 mr-1" />
        Outstanding
      </Badge>
    )
  }

  // Render the component
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
    <>
    <div className="space-y-6">
      {/* Account Header when coming from a specific account */}
      {preSelectedAccount && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <Building2 className="h-5 w-5" />
                  Bank Reconciliation
                </CardTitle>
                <CardDescription>
                  Account: {preSelectedAccount} - {bankAccounts.find(a => a.account_number === preSelectedAccount)?.account_name || 'Loading...'}
                </CardDescription>
              </div>
              {onBack && (
                <Button variant="outline" onClick={onBack}>
                  <ArrowLeft className="h-4 w-4 mr-2" />
                  Back to Accounts
                </Button>
              )}
            </div>
          </CardHeader>
        </Card>
      )}
      
      <Tabs 
        defaultValue="reconcile" 
        className="w-full"
      >
        <TabsList>
          <TabsTrigger value="reconcile">Reconcile</TabsTrigger>
          <TabsTrigger value="outstanding">Outstanding Checks</TabsTrigger>
          <TabsTrigger value="cleared">Cleared Checks</TabsTrigger>
          <TabsTrigger value="reports">Reports</TabsTrigger>
        </TabsList>
        
        {/* Reconcile Tab */}
        <TabsContent value="reconcile" className="space-y-4">
          {/* Account Selection - only show if no preSelectedAccount */}
          {!preSelectedAccount && (
            <Card>
              <CardHeader>
                <CardTitle>Select Bank Account</CardTitle>
                <CardDescription>Choose a bank account to start the reconciliation process</CardDescription>
              </CardHeader>
              <CardContent>
              <div className="space-y-2">
                <Label htmlFor="main-account-select">Bank Account</Label>
                <select 
                  id="main-account-select"
                  value={selectedAccount}
                  onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setSelectedAccount(e.target.value)}
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
            </CardContent>
          </Card>
          )}

          {/* Last Reconciliation Section */}
          {selectedAccount && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Clock className="w-5 h-5" />
                  Last Reconciliation
                </CardTitle>
                <CardDescription>
                  Previous reconciliation data from CHECKREC.DBF for account {selectedAccount}
                </CardDescription>
              </CardHeader>
              <CardContent>
                {loadingLastRec ? (
                  <div className="flex items-center gap-2 text-muted-foreground">
                    <RefreshCw className="w-4 h-4 animate-spin" />
                    Loading last reconciliation...
                  </div>
                ) : lastReconciliation ? (
                  <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                    <div className="space-y-2">
                      <p className="text-sm font-medium text-muted-foreground">Reconciliation Date</p>
                      <p className="text-lg font-semibold">{lastReconciliation.date_string || lastReconciliation.reconcile_date || 'N/A'}</p>
                    </div>
                    <div className="space-y-2">
                      <p className="text-sm font-medium text-muted-foreground">Statement Balance</p>
                      <p className="text-lg font-semibold text-green-600">
                        {formatCurrency(lastReconciliation.ending_balance || 0)}
                      </p>
                    </div>
                    <div className="space-y-2">
                      <p className="text-sm font-medium text-muted-foreground">Beginning Balance</p>
                      <p className="text-lg font-semibold">
                        {formatCurrency(lastReconciliation.beginning_balance || 0)}
                      </p>
                    </div>
                    <div className="space-y-2">
                      <p className="text-sm font-medium text-muted-foreground">Items Cleared</p>
                      <p className="text-lg font-semibold">
                        {lastReconciliation.cleared_count || 0} items
                        {lastReconciliation.cleared_amount && (
                          <span className="text-sm text-muted-foreground block">
                            {formatCurrency(lastReconciliation.cleared_amount)}
                          </span>
                        )}
                      </p>
                    </div>
                  </div>
                ) : (
                  <div className="text-center py-4">
                    <AlertCircle className="w-8 h-8 mx-auto text-muted-foreground mb-2" />
                    <p className="text-muted-foreground">No previous reconciliation found for this account</p>
                    <p className="text-sm text-muted-foreground mt-1">
                      This will be the first reconciliation for account {selectedAccount}
                    </p>
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* Reconciliation Setup */}
          {selectedAccount && (
            <Card>
            <CardHeader>
              <CardTitle>Bank Reconciliation Setup</CardTitle>
              <CardDescription>Configure reconciliation parameters for {selectedAccount}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-3 lg:grid-cols-6">
                <div className="space-y-2">
                  <Label htmlFor="statement-date">Statement Date</Label>
                  <Input 
                    id="statement-date"
                    type="date"
                    value={statementDate}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      setStatementDate(e.target.value)
                      setHasUnsavedChanges(true)
                    }}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="beginning-balance">Beginning Balance</Label>
                  <div className="relative">
                    <Input 
                      id="beginning-balance"
                      type="text"
                      value={formatCurrency(parseFloat(beginningBalance) || 0)}
                      className="bg-gray-50 font-mono text-right pr-4"
                      readOnly
                    />
                  </div>
                  <p className="text-xs text-muted-foreground">From last reconciliation</p>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="statement-credits">Statement Credits</Label>
                  <Input 
                    id="statement-credits"
                    type="text"
                    placeholder="0.00"
                    className="font-mono text-right"
                    value={formatNumberForInput(statementCredits)}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      const value = e.target.value
                      setStatementCredits(value)
                      if (!hasUnsavedChanges) setHasUnsavedChanges(true)
                    }}
                    onBlur={calculateEndingBalance}
                  />
                  <p className="text-xs text-muted-foreground">Deposits per bank statement</p>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="statement-debits">Statement Debits</Label>
                  <Input 
                    id="statement-debits"
                    type="text"
                    placeholder="0.00"
                    className="font-mono text-right"
                    value={formatNumberForInput(statementDebits)}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                      const value = e.target.value
                      setStatementDebits(value)
                      if (!hasUnsavedChanges) setHasUnsavedChanges(true)
                    }}
                    onBlur={calculateEndingBalance}
                  />
                  <p className="text-xs text-muted-foreground">Withdrawals per bank statement</p>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="statement-balance">Ending Statement Balance</Label>
                  <div className="relative">
                    <Input 
                      id="statement-balance"
                      type="text"
                      value={formatCurrency(parseFloat(statementBalance) || 0)}
                      className="bg-gray-50 font-mono text-right pr-4"
                      readOnly
                    />
                  </div>
                  <p className="text-xs text-muted-foreground">Auto-calculated</p>
                </div>
                <div className="space-y-2">
                  <Label>Actions</Label>
                  <div className="flex gap-2 flex-wrap">
                    <Button onClick={loadChecksData} size="sm" variant="outline">
                      <RefreshCw className="w-4 h-4" />
                    </Button>
                    {hasUnsavedChanges && (
                      <Badge variant="outline" className="text-amber-600">
                        Unsaved Changes
                      </Badge>
                    )}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
          )}

          {/* Reconciliation Summary */}
          {selectedAccount && (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-6">
            <Card>
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Beginning Balance</p>
                    <p className="text-xl font-bold">{formatCurrency(totals.beginningBalance)}</p>
                    <p className="text-xs text-muted-foreground">From last reconciliation</p>
                  </div>
                  <FileText className="w-6 h-6 text-muted-foreground" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Selected Credits</p>
                    <p className="text-xl font-bold text-green-600">{formatCurrency(totals.selectedCredits)}</p>
                    <p className="text-xs text-muted-foreground">Matched deposit transactions</p>
                  </div>
                  <ArrowDownLeft className="w-6 h-6 text-green-600" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Selected Debits</p>
                    <p className="text-xl font-bold text-red-600">{formatCurrency(totals.selectedDebits)}</p>
                    <p className="text-xs text-muted-foreground">Matched check transactions</p>
                  </div>
                  <ArrowUpRight className="w-6 h-6 text-red-600" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Calculated Balance</p>
                    <p className="text-xl font-bold text-blue-600">{formatCurrency(totals.calculatedBalance)}</p>
                    <p className="text-xs text-muted-foreground">Begin + Credits - Debits</p>
                  </div>
                  <Calculator className="w-6 h-6 text-blue-600" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Statement Balance</p>
                    <p className="text-xl font-bold text-blue-600">{formatCurrency(totals.statementBalance)}</p>
                    <p className="text-xs text-muted-foreground">Actual ending balance</p>
                  </div>
                  <FileText className="w-6 h-6 text-blue-600" />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Balance Difference</p>
                    <p className={`text-xl font-bold ${totals.isInBalance ? 'text-green-600' : 'text-red-600'}`}>
                      {formatCurrency(Math.abs(totals.reconciliationDifference))}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {totals.isInBalance ? 'Reconciled' : `Selected ${totals.selectedCount} checks`}
                    </p>
                  </div>
                  {totals.isInBalance ? 
                    <CheckCircle className="w-6 h-6 text-green-600" /> : 
                    <AlertCircle className="w-6 h-6 text-red-600" />
                  }
                </div>
              </CardContent>
            </Card>
          </div>
          )}

          {/* CSV Statement Import */}
          {selectedAccount && (
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="flex items-center gap-2">
                    <FileSpreadsheet className="w-5 h-5" />
                    Import Bank Statement
                  </CardTitle>
                  <CardDescription>Upload CSV file to auto-match and select transactions</CardDescription>
                </div>
                <div className="flex gap-2">
                  <Button onClick={() => setShowImportHistory(true)} variant="outline" size="sm">
                    <History className="w-4 h-4 mr-2" />
                    Manage
                  </Button>
                  <Button onClick={() => setCsvImportOpen(true)} variant="outline">
                    <Upload className="w-4 h-4 mr-2" />
                    Import CSV
                  </Button>
                  <Button 
                    onClick={() => handleRunMatching()} 
                    variant="outline"
                    disabled={isRunningMatch || !selectedAccount}
                  >
                    {isRunningMatch ? (
                      <>
                        <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                        Matching...
                      </>
                    ) : (
                      <>
                        <CheckIcon className="w-4 h-4 mr-2" />
                        Run Matching
                      </>
                    )}
                  </Button>
                  <Button 
                    onClick={() => handleRefreshMatching()} 
                    variant="outline"
                    disabled={isRefreshing || !selectedAccount}
                  >
                    {isRefreshing ? (
                      <>
                        <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                        Refreshing...
                      </>
                    ) : (
                      <>
                        <RefreshCw className="w-4 h-4 mr-2" />
                        Clear & Re-match
                      </>
                    )}
                  </Button>
                </div>
              </div>
            </CardHeader>
            
            {/* Match Result Status */}
            {matchResult && (
              <CardContent className="pt-0">
                <div className={`p-3 rounded-lg border ${
                  matchResult.totalMatched > 0 
                    ? 'bg-green-50 border-green-200' 
                    : 'bg-blue-50 border-blue-200'
                }`}>
                  <div className="flex items-center gap-2">
                    {matchResult.totalMatched > 0 ? (
                      <>
                        <CheckCircle className="w-4 h-4 text-green-600" />
                        <span className="text-sm text-green-800">
                          Successfully matched {matchResult.totalMatched} out of {matchResult.totalProcessed} transactions
                        </span>
                      </>
                    ) : (
                      <>
                        <AlertCircle className="w-4 h-4 text-blue-600" />
                        <span className="text-sm text-blue-800">
                          No matching transactions found. Import bank transactions and try again.
                        </span>
                      </>
                    )}
                  </div>
                </div>
              </CardContent>
            )}
          </Card>
          )}

          {/* CSV Import Dialog */}
          <Dialog open={csvImportOpen} onOpenChange={setCsvImportOpen}>
            <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
              <DialogHeader>
                <DialogTitle>Import Bank Statement CSV</DialogTitle>
                <DialogDescription>
                  Upload your bank statement CSV file to import transactions for reconciliation.
                </DialogDescription>
              </DialogHeader>
              
              {!csvParseResult ? (
                <div className="space-y-4">
                  {csvError && (
                    <div className="p-4 border border-red-200 bg-red-50 rounded-lg">
                      <div className="flex items-center gap-2 text-red-800">
                        <AlertCircle className="w-4 h-4" />
                        <span className="font-medium">Import Error</span>
                      </div>
                      <p className="text-sm text-red-700 mt-1">{csvError}</p>
                    </div>
                  )}
                  
                  <div 
                    className="border-dashed border-2 rounded-lg p-8 text-center"
                    onDrop={(e) => {
                      e.preventDefault()
                      e.stopPropagation()
                      const files = e.dataTransfer.files
                      if (files && files[0]) {
                        handleCSVUpload(files[0])
                      }
                    }}
                    onDragOver={(e) => {
                      e.preventDefault()
                      e.stopPropagation()
                    }}
                  >
                    <FileSpreadsheet className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
                    <input
                      type="file"
                      accept=".csv"
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleCSVUpload(e.target.files[0])}
                      className="hidden"
                      id="csv-upload-input"
                      disabled={csvUploading}
                    />
                    <Button 
                      variant="outline" 
                      disabled={csvUploading}
                      onClick={() => document.getElementById('csv-upload-input').click()}
                    >
                      {csvUploading ? (
                        <>
                          <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                          Processing & Importing CSV...
                        </>
                      ) : (
                        <>
                          <Upload className="w-4 h-4 mr-2" />
                          Import CSV Bank Statement
                        </>
                      )}
                    </Button>
                    <p className="text-sm text-muted-foreground mt-2">
                      Drop CSV file here or click to browse
                    </p>
                    <p className="text-xs text-muted-foreground mt-1">
                      CSV file will be imported to SQLite database and automatically matched with outstanding checks
                    </p>
                  </div>
                </div>
              ) : (
                <div className="space-y-4">
                  {/* Import Success Message */}
                  <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-5 h-5 text-green-600" />
                      <div>
                        <p className="font-medium text-green-800">
                          Successfully imported {csvParseResult.transactions?.length || 0} transactions
                        </p>
                        <p className="text-sm text-green-700 mt-1">
                          Close this dialog and click "Run Matching" to match transactions with checks
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Close and Run Matching Button */}
                  <div className="flex justify-end gap-2">
                    <Button 
                      onClick={() => {
                        setCsvImportOpen(false)
                        setCsvParseResult(null)
                      }} 
                      variant="outline"
                    >
                      Close
                    </Button>
                    <Button 
                      onClick={() => {
                        setCsvImportOpen(false)
                        setCsvParseResult(null)
                        handleRunMatching() // This will show the dialog
                      }}
                    >
                      <CheckIcon className="w-4 h-4 mr-2" />
                      Close & Run Matching
                    </Button>
                  </div>

                  {/* Action Buttons */}
                  <div className="flex justify-between items-center pt-4">
                    <div className="text-sm text-muted-foreground">
                      {csvMatches.filter(m => m.confirmed).length} of {csvMatches.length} matches selected
                    </div>
                    <div className="flex gap-2">
                      <Button variant="outline" onClick={() => {
                        setCsvImportOpen(false)
                        setCsvParseResult(null)
                        setCsvMatches([])
                        setCsvError(null)
                      }}>
                        Cancel
                      </Button>
                      <Button 
                        onClick={handleConfirmCSVMatches}
                        disabled={csvMatches.filter(m => m.confirmed).length === 0}
                        className="bg-blue-600 hover:bg-blue-700"
                      >
                        <CheckIcon className="w-4 h-4 mr-2" />
                        Apply {csvMatches.filter(m => m.confirmed).length} Matches
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </DialogContent>
          </Dialog>

          {/* Draft Status and Commit Actions */}
          {selectedAccount && draftSelectedChecks.size > 0 && totals.isInBalance && (
            <Card className="border-green-200">
              <CardHeader>
                <CardTitle className="text-green-800">Reconciliation Ready to Commit</CardTitle>
                <CardDescription>
                  Perfect! You have {draftSelectedChecks.size} checks selected and the reconciliation is balanced. 
                  {hasUnsavedChanges ? ' Changes are being auto-saved.' : ' All changes saved.'}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex gap-3 items-center flex-wrap">
                  <Button 
                    onClick={commitReconciliation}
                    disabled={reconciliationInProgress}
                    className="bg-blue-600 hover:bg-blue-700"
                  >
                    {reconciliationInProgress ? (
                      <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                    ) : (
                      <CheckCircle className="w-4 h-4 mr-2" />
                    )}
                    Complete Reconciliation
                  </Button>
                  
                  <Button 
                    onClick={saveDraftReconciliation}
                    variant="outline"
                    size="sm"
                  >
                    <Save className="w-4 h-4 mr-2" />
                    Save Draft Now
                  </Button>
                  
                  <Button 
                    onClick={clearDraftReconciliation}
                    variant="outline"
                    size="sm"
                    className="text-red-600 hover:text-red-700"
                  >
                    <X className="w-4 h-4 mr-2" />
                    Clear Draft
                  </Button>
                  
                  <div className="text-sm text-muted-foreground">
                    Selected: {draftSelectedChecks.size} checks • 
                    Total: {formatCurrency(
                      Array.from(draftSelectedChecks)
                        .reduce((sum, checkId) => {
                          const check = checks.find(c => c.id === checkId)
                          return sum + (check?.amount || 0)
                        }, 0)
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Side-by-Side Reconciliation View */}
          {selectedAccount && showSideBySide && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>Side-by-Side Reconciliation</span>
                <div className="flex gap-2">
                  <Button 
                    variant="outline" 
                    size="sm"
                    onClick={() => setShowSideBySide(false)}
                  >
                    <Eye className="w-4 h-4 mr-2" />
                    Manual Mode
                  </Button>
                  <Button 
                    variant="outline" 
                    size="sm"
                    onClick={loadBankTransactions}
                    disabled={loadingBankTransactions}
                  >
                    {loadingBankTransactions ? (
                      <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    ) : (
                      <RefreshCw className="w-4 h-4 mr-2" />
                    )}
                    Refresh
                  </Button>
                </div>
              </CardTitle>
              <CardDescription>
                Match bank transactions with outstanding checks. Matched transactions are automatically removed from this view.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-6">
                {/* Bank Transactions (Left Side) */}
                <div className="space-y-4">
                  <h3 className="font-semibold text-lg">Bank Statement Transactions ({getUnmatchedBankTransactions().length})</h3>
                  <div className="max-h-96 overflow-y-auto border rounded-lg">
                    <Table>
                      <TableHeader className="sticky top-0 bg-white">
                        <TableRow>
                          <TableHead className="w-12">Select</TableHead>
                          <TableHead 
                            className="cursor-pointer hover:bg-gray-50"
                            onClick={() => handleBankSort('transaction_date')}
                          >
                            <div className="flex items-center gap-1">
                              Date
                              {bankSortField === 'transaction_date' && (
                                bankSortDirection === 'asc' ? 
                                  <ChevronUp className="w-4 h-4" /> : 
                                  <ChevronDown className="w-4 h-4" />
                              )}
                            </div>
                          </TableHead>
                          <TableHead 
                            className="cursor-pointer hover:bg-gray-50"
                            onClick={() => handleBankSort('description')}
                          >
                            <div className="flex items-center gap-1">
                              Description
                              {bankSortField === 'description' && (
                                bankSortDirection === 'asc' ? 
                                  <ChevronUp className="w-4 h-4" /> : 
                                  <ChevronDown className="w-4 h-4" />
                              )}
                            </div>
                          </TableHead>
                          <TableHead 
                            className="cursor-pointer hover:bg-gray-50"
                            onClick={() => handleBankSort('check_number')}
                          >
                            <div className="flex items-center gap-1">
                              Check #
                              {bankSortField === 'check_number' && (
                                bankSortDirection === 'asc' ? 
                                  <ChevronUp className="w-4 h-4" /> : 
                                  <ChevronDown className="w-4 h-4" />
                              )}
                            </div>
                          </TableHead>
                          <TableHead 
                            className="cursor-pointer hover:bg-gray-50 text-right"
                            onClick={() => handleBankSort('amount')}
                          >
                            <div className="flex items-center justify-end gap-1">
                              Amount
                              {bankSortField === 'amount' && (
                                bankSortDirection === 'asc' ? 
                                  <ChevronUp className="w-4 h-4" /> : 
                                  <ChevronDown className="w-4 h-4" />
                              )}
                            </div>
                          </TableHead>
                          <TableHead>Actions</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {getUnmatchedBankTransactions().map((transaction) => (
                          <TableRow 
                            key={transaction.id}
                            className={selectedBankTxn?.id === transaction.id ? 'bg-blue-50' : ''}
                          >
                            <TableCell>
                              <input
                                type="radio"
                                name="bankTxnSelect"
                                checked={selectedBankTxn?.id === transaction.id}
                                onChange={() => setSelectedBankTxn(transaction)}
                                className="cursor-pointer"
                              />
                            </TableCell>
                            <TableCell>{formatDate(transaction.transaction_date)}</TableCell>
                            <TableCell className="max-w-32 truncate">{transaction.description}</TableCell>
                            <TableCell className="font-mono">{transaction.check_number || '-'}</TableCell>
                            <TableCell className="text-right font-mono">
                              <span className={transaction.amount < 0 ? 'text-red-600' : 'text-green-600'}>
                                {formatCurrency(Math.abs(transaction.amount))}
                              </span>
                            </TableCell>
                            <TableCell>
                              {transaction.matched_check_id ? (
                                <Badge variant="default" className="bg-green-100 text-green-800">
                                  <CheckIcon className="w-3 h-3 mr-1" />
                                  Matched
                                </Badge>
                              ) : (
                                <Badge variant="outline" className="bg-blue-100 text-blue-800">
                                  Available
                                </Badge>
                              )}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                </div>

                {/* Outstanding Checks (Right Side) */}
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold text-lg">
                      Outstanding Checks ({getAvailableChecks().length})
                      {draftSelectedChecks.size > 0 && (
                        <span className="text-sm font-normal text-muted-foreground ml-2">
                          ({draftSelectedChecks.size} already selected)
                        </span>
                      )}
                    </h3>
                    <div className="flex items-center gap-2">
                      <Button
                        variant={limitToStatementDate ? "default" : "outline"}
                        size="sm"
                        onClick={() => {
                          const newValue = !limitToStatementDate
                          logger.debug('Date filter toggled', { 
                            from: limitToStatementDate, 
                            to: newValue, 
                            statementDate,
                            checksBeforeToggle: getAvailableChecks().length 
                          })
                          setLimitToStatementDate(newValue)
                        }}
                        disabled={!statementDate}
                        title={!statementDate ? "Set a statement date first" : "Toggle date filter"}
                      >
                        <Calendar className="w-4 h-4 mr-1" />
                        {limitToStatementDate && statementDate ? `≤ ${formatDate(statementDate)}` : 'All Dates'}
                      </Button>
                    </div>
                  </div>
                  {limitToStatementDate && statementDate && (
                    <div className="text-sm text-muted-foreground">
                      Showing checks dated on or before {formatDate(statementDate)}
                    </div>
                  )}
                  {checks.length > 0 && (
                    <div className="space-y-1">
                      <div className="text-xs text-muted-foreground">
                        Total loaded: {checks.length} checks | 
                        Available: {getAvailableChecks().length} | 
                        Matched: {matchedTransactions.length} | 
                        Draft selected: {draftSelectedChecks.size}
                      </div>
                      {draftSelectedChecks.size > 0 && (
                        <div className="text-xs bg-blue-50 p-2 rounded border border-blue-200">
                          <div className="font-semibold text-blue-900 mb-1">Draft Selected Checks ({draftSelectedChecks.size}):</div>
                          <div className="text-blue-700 space-y-0.5">
                            {Array.from(draftSelectedChecks).slice(0, 5).map(checkId => {
                              const check = checks.find(c => c.id === checkId)
                              if (!check) return null
                              return (
                                <div key={checkId} className="flex items-center gap-2">
                                  <span>• Check #{check.checkNumber}</span>
                                  <span>{formatDate(check.checkDate)}</span>
                                  <span>{check.payee}</span>
                                  <span className="font-mono">{formatCurrency(check.amount)}</span>
                                </div>
                              )
                            })}
                            {draftSelectedChecks.size > 5 && (
                              <div className="italic">...and {draftSelectedChecks.size - 5} more</div>
                            )}
                          </div>
                          <Button 
                            size="sm" 
                            variant="outline" 
                            className="mt-2"
                            onClick={() => {
                              setDraftSelectedChecks(new Set())
                              setSelectedChecks(new Set())
                            }}
                          >
                            Clear All Selections
                          </Button>
                        </div>
                      )}
                    </div>
                  )}
                  <div className="max-h-96 overflow-y-auto border rounded-lg">
                    <Table>
                      <TableHeader className="sticky top-0 bg-white">
                        <TableRow>
                          <TableHead className="w-12">Select</TableHead>
                          <TableHead 
                            className="cursor-pointer hover:bg-gray-50"
                            onClick={() => handleCheckSort('checkNumber')}
                          >
                            <div className="flex items-center gap-1">
                              Check #
                              {checkSortField === 'checkNumber' && (
                                checkSortDirection === 'asc' ? 
                                  <ChevronUp className="w-4 h-4" /> : 
                                  <ChevronDown className="w-4 h-4" />
                              )}
                            </div>
                          </TableHead>
                          <TableHead 
                            className="cursor-pointer hover:bg-gray-50"
                            onClick={() => handleCheckSort('checkDate')}
                          >
                            <div className="flex items-center gap-1">
                              Date
                              {checkSortField === 'checkDate' && (
                                checkSortDirection === 'asc' ? 
                                  <ChevronUp className="w-4 h-4" /> : 
                                  <ChevronDown className="w-4 h-4" />
                              )}
                            </div>
                          </TableHead>
                          <TableHead 
                            className="cursor-pointer hover:bg-gray-50"
                            onClick={() => handleCheckSort('payee')}
                          >
                            <div className="flex items-center gap-1">
                              Payee
                              {checkSortField === 'payee' && (
                                checkSortDirection === 'asc' ? 
                                  <ChevronUp className="w-4 h-4" /> : 
                                  <ChevronDown className="w-4 h-4" />
                              )}
                            </div>
                          </TableHead>
                          <TableHead 
                            className="cursor-pointer hover:bg-gray-50 text-right"
                            onClick={() => handleCheckSort('amount')}
                          >
                            <div className="flex items-center justify-end gap-1">
                              Amount
                              {checkSortField === 'amount' && (
                                checkSortDirection === 'asc' ? 
                                  <ChevronUp className="w-4 h-4" /> : 
                                  <ChevronDown className="w-4 h-4" />
                              )}
                            </div>
                          </TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {getAvailableChecks().map((check) => (
                          <TableRow 
                            key={check.id}
                            className={selectedCheckForMatch?.id === check.id ? 'bg-blue-50' : ''}
                          >
                            <TableCell>
                              <Checkbox
                                checked={selectedChecksForMatch.has(check.id)}
                                onCheckedChange={(checked) => {
                                  const newSelected = new Set(selectedChecksForMatch)
                                  if (checked) {
                                    newSelected.add(check.id)
                                    // Also store the check object for amount calculation
                                    if (!selectedCheckForMatch) setSelectedCheckForMatch(check)
                                  } else {
                                    newSelected.delete(check.id)
                                    if (selectedCheckForMatch?.id === check.id) setSelectedCheckForMatch(null)
                                  }
                                  setSelectedChecksForMatch(newSelected)
                                }}
                              />
                            </TableCell>
                            <TableCell className="font-mono">{check.checkNumber}</TableCell>
                            <TableCell>{formatDate(check.checkDate)}</TableCell>
                            <TableCell>{check.payee}</TableCell>
                            <TableCell className="text-right font-mono">{formatCurrency(check.amount)}</TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                </div>
              </div>
              
              {/* Match Button and Status */}
              <div className="mt-6 space-y-2">
                <div className="flex justify-center">
                  <Button 
                    onClick={handleManualMatch}
                    disabled={!selectedBankTxn || selectedChecksForMatch.size === 0 || isManualMatching}
                    className="px-6"
                  >
                    {isManualMatching ? (
                      <>
                        <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                        Matching...
                      </>
                    ) : (
                      <>
                        <CheckIcon className="w-4 h-4 mr-2" />
                        Match Selected Items
                      </>
                    )}
                  </Button>
                </div>
                
                {/* Match Status Display */}
                {selectedBankTxn && selectedChecksForMatch.size > 0 && (
                  <div className="text-center text-sm text-muted-foreground">
                    {(() => {
                      const selectedChecksList = Array.from(selectedChecksForMatch).map(id => 
                        checks.find(c => c.id === id)
                      ).filter(Boolean)
                      const totalCheckAmount = selectedChecksList.reduce((sum, check) => sum + check.amount, 0)
                      const bankAmount = Math.abs(selectedBankTxn.amount)
                      const difference = Math.abs(bankAmount - totalCheckAmount)
                      
                      return (
                        <div className="space-y-1">
                          <div>
                            Bank transaction: {formatCurrency(bankAmount)} on {formatDate(selectedBankTxn.transaction_date)}
                          </div>
                          <div>
                            Selected {selectedChecksList.length} check(s): {formatCurrency(totalCheckAmount)}
                          </div>
                          {difference > 0.01 && (
                            <div className="text-amber-600">
                              Difference: {formatCurrency(difference)}
                            </div>
                          )}
                        </div>
                      )
                    })()}
                  </div>
                )}
              </div>
              
            </CardContent>
          </Card>
        )}
          
        {/* Matched Transactions Box Below */}
        {showSideBySide && (
          <Card className="mt-4">
            <CardHeader>
              <CardTitle className="text-lg">
                Matched Transactions ({matchedTransactions.length})
              </CardTitle>
              <CardDescription>
                Successfully matched bank transactions with checks
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="max-h-64 overflow-y-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Date</TableHead>
                      <TableHead>Bank Description</TableHead>
                      <TableHead>Check #</TableHead>
                      <TableHead>Payee</TableHead>
                      <TableHead className="text-right">Amount</TableHead>
                      <TableHead>Confidence</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {matchedTransactions.slice(0, 10).map((match) => (
                      <TableRow key={match.id}>
                        <TableCell>{formatDate(match.check_date)}</TableCell>
                        <TableCell className="max-w-xs truncate">
                          {match.bank_description || 'No description'}
                        </TableCell>
                        <TableCell className="font-mono">{match.check_number}</TableCell>
                        <TableCell>{match.payee}</TableCell>
                        <TableCell className="text-right font-mono">
                          {formatCurrency(match.amount)}
                        </TableCell>
                        <TableCell>
                          <Badge variant={match.match_confidence > 0.8 ? "default" : "secondary"}>
                            {Math.round(match.match_confidence * 100)}%
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => handleUnmatchTransaction(match.bank_txn_id)}
                          >
                            Unmatch
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
                {matchedTransactions.length > 10 && (
                  <div className="text-center mt-2 text-sm text-muted-foreground">
                    Showing 10 of {matchedTransactions.length} matched transactions
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        )}

          {/* Toggle Button for Manual Mode */}
          {selectedAccount && !showSideBySide && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>Manual Reconciliation Mode</span>
                <Button 
                  variant="outline" 
                  onClick={() => setShowSideBySide(true)}
                  disabled={loadingBankTransactions}
                >
                  <FileSpreadsheet className="w-4 h-4 mr-2" />
                  Show Side-by-Side View
                </Button>
              </CardTitle>
              <CardDescription>
                Manual mode allows you to select checks without imported bank transactions.
              </CardDescription>
            </CardHeader>
          </Card>
          )}

          {/* Filters and Checks Table */}
          {selectedAccount && !showSideBySide && (
          <>
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
                    onCheckedChange={(checked) => setShowCleared(checked === true)}
                  />
                  <Label htmlFor="show-cleared">Show Cleared</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <Checkbox 
                    id="show-uncleared"
                    checked={showUncleared}
                    onCheckedChange={(checked) => setShowUncleared(checked === true)}
                  />
                  <Label htmlFor="show-uncleared">Show Outstanding</Label>
                </div>
                <div className="space-y-1">
                  <Label className="text-xs">Transaction Type</Label>
                  <select
                    value={showTransactionType}
                    onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setShowTransactionType(e.target.value as 'all' | 'debits' | 'credits')}
                    className="flex h-8 w-full rounded-md border border-input bg-background px-2 py-1 text-xs"
                  >
                    <option value="all">All Types</option>
                    <option value="debits">Checks Only</option>
                    <option value="credits">Deposits Only</option>
                  </select>
                </div>
                <div className="flex items-center space-x-2">
                  <Checkbox 
                    id="limit-statement-date"
                    checked={limitToStatementDate}
                    onCheckedChange={(checked) => setLimitToStatementDate(checked === true)}
                  />
                  <Label htmlFor="limit-statement-date" className="text-xs">Through Statement Date</Label>
                </div>
                <div className="space-y-1">
                  <Label htmlFor="date-from" className="text-xs">Date From</Label>
                  <Input 
                    id="date-from"
                    type="date"
                    value={dateFrom}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setDateFrom(e.target.value)}
                    className="h-8 text-xs"
                  />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="date-to" className="text-xs">Date To</Label>
                  <Input 
                    id="date-to"
                    type="date"
                    value={dateTo}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setDateTo(e.target.value)}
                    className="h-8 text-xs"
                  />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="amount-from" className="text-xs">Amount From</Label>
                  <Input 
                    id="amount-from"
                    type="number"
                    step="0.01"
                    value={amountFrom}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setAmountFrom(e.target.value)}
                    className="h-8 text-xs"
                  />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="amount-to" className="text-xs">Amount To</Label>
                  <Input 
                    id="amount-to"
                    type="number"
                    step="0.01"
                    value={amountTo}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setAmountTo(e.target.value)}
                    className="h-8 text-xs"
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Transactions Table */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Transactions ({filteredChecks.length})</CardTitle>
                  <CardDescription>Click checkbox to select, click status to toggle cleared</CardDescription>
                </div>
                <div className="flex gap-2">
                  <Badge variant="outline">
                    {selectedChecks.size} selected
                  </Badge>
                </div>
              </div>
            </CardHeader>
            <CardContent className="p-0">
              <div className="relative">
                <div className="overflow-hidden border-b bg-white">
                  <Table>
                    <TableHeader>
                      <TableRow>
                    <TableHead className="w-12">
                      <Checkbox 
                        checked={selectedChecks.size === filteredChecks.length && filteredChecks.length > 0}
                        onCheckedChange={(checked) => {
                          if (checked) {
                            const allCheckIds = new Set(filteredChecks.map(c => c.id))
                            setSelectedChecks(allCheckIds)
                            setDraftSelectedChecks(new Set([...draftSelectedChecks, ...allCheckIds]))
                          } else {
                            // Remove all filtered checks from both selections
                            const filteredCheckIds = new Set(filteredChecks.map(c => c.id))
                            const newDraftSelected = new Set([...draftSelectedChecks].filter(id => !filteredCheckIds.has(id)))
                            setSelectedChecks(new Set())
                            setDraftSelectedChecks(newDraftSelected)
                          }
                          setHasUnsavedChanges(true)
                        }}
                      />
                    </TableHead>
                    {renderSortableHeader('checkNumber', 'Check #')}
                    {renderSortableHeader('checkDate', 'Date')}
                    {renderSortableHeader('payee', 'Payee')}
                    <TableHead>Type</TableHead>
                    {renderSortableHeader('memo', 'Memo')}
                    {renderSortableHeader('amount', 'Amount', 'text-right')}
                    <TableHead>Account</TableHead>
                    {renderSortableHeader('cleared', 'Status')}
                    <TableHead>Reconciled</TableHead>
                  </TableRow>
                </TableHeader>
                  </Table>
                </div>
                <div className="max-h-96 overflow-y-auto">
                  <Table>
                    <TableBody>
                  {filteredChecks.map((check) => {
                    const isDraftSelected = draftSelectedChecks.has(check.id)
                    const isUISelected = selectedChecks.has(check.id)
                    
                    return (
                    <TableRow 
                      key={check.id} 
                      className={`${check.cleared ? 'bg-green-50' : ''} ${isDraftSelected ? 'bg-blue-50 border-l-4 border-blue-500' : ''}`}
                    >
                      <TableCell>
                        <Checkbox 
                          checked={isUISelected}
                          onCheckedChange={(checked) => {
                            // Update both UI selection and draft reconciliation
                            const newSelected = new Set(selectedChecks)
                            const newDraftSelected = new Set(draftSelectedChecks)
                            
                            if (checked) {
                              newSelected.add(check.id)
                              newDraftSelected.add(check.id)
                            } else {
                              newSelected.delete(check.id)
                              newDraftSelected.delete(check.id)
                            }
                            
                            setSelectedChecks(newSelected)
                            setDraftSelectedChecks(newDraftSelected)
                            setHasUnsavedChanges(true)
                            logger.debug('Checkbox clicked - Draft selected checks', { count: newDraftSelected.size })
                          }}
                        />
                      </TableCell>
                      <TableCell className="font-mono">{check.checkNumber}</TableCell>
                      <TableCell>{formatDate(check.checkDate)}</TableCell>
                      <TableCell>{check.payee}</TableCell>
                      <TableCell>
                        {getTransactionTypeBadge(check.entryType)}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">{check.memo}</TableCell>
                      <TableCell className="text-right font-mono">{formatCurrency(check.amount)}</TableCell>
                      <TableCell className="font-mono text-sm">{check.accountNumber}</TableCell>
                      <TableCell>
                        <button onClick={() => toggleCheckSelection(check.id)}>
                          {isDraftSelected ? (
                            <Badge variant="default" className="bg-blue-100 text-blue-800">
                              <CheckCircle className="w-3 h-3 mr-1" />
                              Selected for Reconciliation
                            </Badge>
                          ) : (
                            getStatusBadge(check.cleared)
                          )}
                        </button>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {check.reconcileDate ? formatDate(check.reconcileDate) : '-'}
                      </TableCell>
                    </TableRow>
                    )
                  })}
                    </TableBody>
                  </Table>
                </div>
              </div>
            </CardContent>
          </Card>
          </>
          )}
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
                    <TableHead>Type</TableHead>
                    <TableHead className="text-right">Amount</TableHead>
                    <TableHead>Account</TableHead>
                    <TableHead>Days Outstanding</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {checks.filter(check => !check.cleared).map((check) => {
                    const daysOutstanding = check.checkDate 
                      ? Math.floor((new Date().getTime() - new Date(check.checkDate).getTime()) / (1000 * 60 * 60 * 24))
                      : 0
                    
                    return (
                      <TableRow key={check.id}>
                        <TableCell className="font-mono">{check.checkNumber}</TableCell>
                        <TableCell>{formatDate(check.checkDate)}</TableCell>
                        <TableCell>{check.payee}</TableCell>
                        <TableCell>
                          {getTransactionTypeBadge(check.entryType)}
                        </TableCell>
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
                    <TableHead>Type</TableHead>
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
                      <TableCell>
                        {getTransactionTypeBadge(check.entryType)}
                      </TableCell>
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

    {/* Import History Dialog */}
      <Dialog open={showImportHistory} onOpenChange={setShowImportHistory}>
        <DialogContent className="max-w-5xl max-h-[80vh] overflow-hidden flex flex-col">
          <DialogHeader className="flex-shrink-0">
            <DialogTitle>Import History</DialogTitle>
            <DialogDescription>
              View and manage your previously imported bank statements.
            </DialogDescription>
          </DialogHeader>
          
          <div className="flex-1 overflow-auto">
            {loadingHistory ? (
              <div className="flex items-center justify-center p-8">
                <Loader2 className="w-6 h-6 animate-spin" />
              </div>
            ) : importHistory.length === 0 ? (
              <div className="text-center p-8 text-muted-foreground">
                No imports found for this account
              </div>
            ) : (
              <div className="pr-2">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[120px]">Import Date</TableHead>
                      <TableHead className="w-[120px]">Statement</TableHead>
                      <TableHead className="w-[100px] text-center">Transactions</TableHead>
                      <TableHead className="w-[100px] text-center">Matched</TableHead>
                      <TableHead className="w-[120px]">Imported By</TableHead>
                      <TableHead className="w-[100px]">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {importHistory.map((imp) => {
                      logger.debug('Import record', { importId: imp?.import_batch_id });
                      return (
                        <TableRow key={imp.import_batch_id}>
                          <TableCell className="font-mono text-sm">
                            {new Date(imp.import_date).toLocaleDateString()}
                          </TableCell>
                          <TableCell className="font-mono text-sm">
                            {imp.statement_date ? new Date(imp.statement_date).toLocaleDateString() : '-'}
                          </TableCell>
                          <TableCell className="text-center">{imp.transaction_count}</TableCell>
                          <TableCell className="text-center">{imp.matched_count}</TableCell>
                          <TableCell>{imp.imported_by}</TableCell>
                          <TableCell>
                            <span 
                              onClick={() => {
                                logger.debug('Delete clicked for import', { importId: imp.import_batch_id });
                                handleDeleteImport(imp.import_batch_id);
                              }}
                              style={{
                                display: 'inline-block',
                                padding: '6px 16px',
                                backgroundColor: '#dc2626',
                                color: '#ffffff',
                                borderRadius: '6px',
                                cursor: 'pointer',
                                fontSize: '14px',
                                fontWeight: '500',
                                textAlign: 'center',
                                border: '1px solid #b91c1c',
                                boxShadow: '0 1px 2px rgba(0,0,0,0.1)'
                              }}
                              onMouseOver={(e) => {
                                (e.target as HTMLElement).style.backgroundColor = '#b91c1c';
                              }}
                              onMouseOut={(e) => {
                                (e.target as HTMLElement).style.backgroundColor = '#dc2626';
                              }}
                            >
                              🗑️ Delete
                            </span>
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Confirm Delete</DialogTitle>
            <DialogDescription>
              This action cannot be undone. All transactions from this import will be permanently deleted.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <p className="text-sm text-muted-foreground">
              Are you sure you want to delete this import and all its transactions?
            </p>
          </div>
          <div className="flex gap-2 justify-end">
            <Button
              variant="outline"
              onClick={() => {
                setDeleteConfirmOpen(false)
                setImportToDelete(null)
              }}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDelete}
            >
              Delete Import
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Matching Options Dialog */}
      <Dialog open={showMatchingOptions} onOpenChange={setShowMatchingOptions}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Transaction Matching Options</DialogTitle>
            <DialogDescription>
              Choose how to match bank transactions with checks
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div className="space-y-3">
              <label 
                className="flex items-start space-x-3 p-4 border rounded-lg cursor-pointer hover:bg-gray-50"
                onClick={() => setMatchingDateOption('all')}
              >
                <input
                  type="radio"
                  name="matchOption"
                  value="all"
                  checked={matchingDateOption === 'all'}
                  onChange={() => setMatchingDateOption('all')}
                  className="mt-1"
                />
                <div className="flex-1">
                  <div className="font-medium">
                    Match all available checks
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">
                    Include all checks regardless of date, including future-dated checks
                  </p>
                </div>
              </label>
              
              <label 
                className="flex items-start space-x-3 p-4 border rounded-lg cursor-pointer hover:bg-gray-50"
                onClick={() => setMatchingDateOption('statement')}
              >
                <input
                  type="radio"
                  name="matchOption"
                  value="statement"
                  checked={matchingDateOption === 'statement'}
                  onChange={() => setMatchingDateOption('statement')}
                  className="mt-1"
                />
                <div className="flex-1">
                  <div className="font-medium">
                    Match only up to statement date
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">
                    Only match checks dated on or before {statementDate || 'the statement date'}
                  </p>
                </div>
              </label>
            </div>
            
            {matchingDateOption === 'statement' && !statementDate && (
              <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-md">
                <p className="text-sm text-yellow-800">
                  Please set a statement date in the form above before proceeding
                </p>
              </div>
            )}
          </div>
          
          <div className="flex justify-end gap-2 mt-4">
            <Button variant="outline" onClick={() => setShowMatchingOptions(false)}>
              Cancel
            </Button>
            <Button 
              onClick={() => {
                if (isRefreshing) {
                  handleRefreshMatching(true)
                } else {
                  handleRunMatching(true)
                }
              }}
              disabled={matchingDateOption === 'statement' && !statementDate}
            >
              <CheckIcon className="w-4 h-4 mr-2" />
              Run Matching
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}