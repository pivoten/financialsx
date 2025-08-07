import { useState, useEffect } from 'react'
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
  Check,
  Eye,
  History,
  Trash2,
  ArrowLeft,
  Building2
} from 'lucide-react'

export function BankReconciliation({ companyName, currentUser, preSelectedAccount, onBack }) {
  // State management
  const [checks, setChecks] = useState([])
  const [bankAccounts, setBankAccounts] = useState([])
  const [loading, setLoading] = useState(true)
  const [loadingAccounts, setLoadingAccounts] = useState(true)
  const [error, setError] = useState(null)
  const [selectedAccount, setSelectedAccount] = useState(preSelectedAccount || '')
  const [reconciliationPeriod, setReconciliationPeriod] = useState('')
  const [statementBalance, setStatementBalance] = useState('')
  const [statementDate, setStatementDate] = useState('')
  const [beginningBalance, setBeginningBalance] = useState('')
  const [statementCredits, setStatementCredits] = useState('')
  const [statementDebits, setStatementDebits] = useState('')
  const [selectedChecks, setSelectedChecks] = useState(new Set())
  const [reconciliationInProgress, setReconciliationInProgress] = useState(false)
  const [lastReconciliation, setLastReconciliation] = useState(null)
  const [loadingLastRec, setLoadingLastRec] = useState(false)

  // Draft reconciliation state
  const [draftMode, setDraftMode] = useState(true) // Always start in draft mode
  const [draftReconciliation, setDraftReconciliation] = useState(null)
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false)
  const [draftSelectedChecks, setDraftSelectedChecks] = useState(new Set()) // Separate from actual selectedChecks

  // Bank Transactions state (imported via CSV or manual entry)
  const [bankTransactions, setBankTransactions] = useState([])
  const [loadingBankTransactions, setLoadingBankTransactions] = useState(false)
  const [csvImportOpen, setCsvImportOpen] = useState(false)
  const [csvUploading, setCsvUploading] = useState(false)
  const [csvParseResult, setCsvParseResult] = useState(null)
  const [csvMatches, setCsvMatches] = useState([])
  const [csvError, setCsvError] = useState(null)
  const [showSideBySide, setShowSideBySide] = useState(false)
  const [showImportHistory, setShowImportHistory] = useState(false)
  const [importHistory, setImportHistory] = useState([])
  const [loadingHistory, setLoadingHistory] = useState(false)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [importToDelete, setImportToDelete] = useState(null)
  
  // Manual matching state
  const [selectedBankTxn, setSelectedBankTxn] = useState(null)
  const [selectedCheckForMatch, setSelectedCheckForMatch] = useState(null)
  const [selectedChecksForMatch, setSelectedChecksForMatch] = useState(new Set()) // Multiple check selection
  const [isManualMatching, setIsManualMatching] = useState(false)

  // Filters
  const [showCleared, setShowCleared] = useState(false)
  const [showUncleared, setShowUncleared] = useState(true)
  const [showTransactionType, setShowTransactionType] = useState('all') // 'all', 'debits', 'credits'
  const [limitToStatementDate, setLimitToStatementDate] = useState(false) // Default to showing all checks
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [amountFrom, setAmountFrom] = useState('')
  const [amountTo, setAmountTo] = useState('')

  // Sorting
  const [sortField, setSortField] = useState('checkDate')
  const [sortDirection, setSortDirection] = useState('asc')
  
  // Bank transactions sorting
  const [bankSortField, setBankSortField] = useState('transaction_date')
  const [bankSortDirection, setBankSortDirection] = useState('asc')
  
  // Outstanding checks sorting (for side-by-side view)
  const [checkSortField, setCheckSortField] = useState('checkNumber')
  const [checkSortDirection, setCheckSortDirection] = useState('asc')
  
  // Matched transactions state
  const [matchedTransactions, setMatchedTransactions] = useState([])
  const [loadingMatched, setLoadingMatched] = useState(false)
  const [matchedSearchTerm, setMatchedSearchTerm] = useState('')
  const [selectedMatchedTxns, setSelectedMatchedTxns] = useState(new Set())
  const [isBulkUnmatching, setIsBulkUnmatching] = useState(false)
  
  // Matching operations state
  const [isRunningMatch, setIsRunningMatch] = useState(false)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [matchResult, setMatchResult] = useState(null)
  const [showMatchingOptions, setShowMatchingOptions] = useState(false)
  const [matchingDateOption, setMatchingDateOption] = useState('all') // 'all' or 'statement'

  // Load data when component mounts
  useEffect(() => {
    if (companyName) {
      loadBankAccounts()
      // Don't load checks data here - wait for account selection
      // loadChecksData()
      // loadBankTransactions()
      // loadMatchedTransactions()
    }
  }, [companyName])
  
  // Set selectedAccount when preSelectedAccount is provided
  useEffect(() => {
    if (preSelectedAccount && preSelectedAccount !== selectedAccount) {
      setSelectedAccount(preSelectedAccount)
    }
  }, [preSelectedAccount])
  
  // Debug log when filter changes
  useEffect(() => {
    if (checks.length > 0) {
      const available = getAvailableChecks()
      console.log('Filter state changed:', {
        limitToStatementDate,
        statementDate,
        totalChecks: checks.length,
        availableChecks: available.length,
        selectedAccount
      })
    }
  }, [limitToStatementDate, statementDate, checks.length])

  // Load last reconciliation and checks data when selected account changes
  useEffect(() => {
    if (companyName && selectedAccount) {
      // Load data sequentially to avoid overwriting saved draft values
      const loadAccountData = async () => {
        await loadLastReconciliation()
        await loadChecksData() 
        await loadBankTransactions()
        await loadMatchedTransactions()
        // Load draft AFTER everything else is loaded to avoid overwriting
        await loadDraftReconciliation()
      }
      loadAccountData()
    }
  }, [companyName, selectedAccount])

  // Auto-save functionality when form values or selected checks change
  useEffect(() => {
    console.log('🔍 Auto-save useEffect triggered:', {
      hasUnsavedChanges,
      selectedAccount,
      companyName,
      draftSelectedChecksSize: draftSelectedChecks.size,
      condition: hasUnsavedChanges && selectedAccount && companyName
    })
    
    if (hasUnsavedChanges && selectedAccount && companyName) {
      console.log('🔄 Auto-save timer started - will save in 10 seconds. Selected checks:', draftSelectedChecks.size)
      // Debounce auto-save to avoid too many saves
      const timeoutId = setTimeout(() => {
        console.log('💾 Auto-save triggered with', draftSelectedChecks.size, 'selected checks')
        saveDraftReconciliation()
      }, 10000) // Auto-save after 10 seconds of inactivity

      return () => clearTimeout(timeoutId)
    }
  }, [hasUnsavedChanges, selectedAccount, companyName, Array.from(draftSelectedChecks).join(',')]) // Use actual selected check IDs to trigger auto-save

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
    if (!companyName || !selectedAccount) {
      console.log('BankReconciliation: Skipping check load - no account selected')
      return
    }
    
    try {
      setLoading(true)
      setError(null)
      
      // Use the dedicated GetOutstandingChecks function with account filter
      const accountFilter = selectedAccount
      console.log('BankReconciliation: Loading checks for account:', accountFilter)
      
      const result = await GetOutstandingChecks(companyName, accountFilter)
      
      console.log('BankReconciliation: GetOutstandingChecks response:', result)
      
      if (result.status === 'error') {
        setError(result.error || 'Failed to load checks data')
        setChecks([])
      } else {
        const checksData = result.checks || []
        console.log('BankReconciliation: Loaded checks:', checksData.length, 'checks')
        
        // The GetOutstandingChecks function returns properly structured data
        // Use CIDCHEC as the unique identifier for reliable tracking
        const processedChecks = checksData.map((check, index) => {
          // Create a unique ID: prefer CIDCHEC, fallback to composite key
          const uniqueId = check.cidchec && check.cidchec.trim() !== '' 
            ? check.cidchec 
            : `${check.account || 'unknown'}-${check.checkNumber || 'unknown'}-${check.amount || 0}-${index}`
          
          return {
            id: uniqueId,
            cidchec: check.cidchec || '', // Store CIDCHEC separately for explicit access
            checkNumber: check.checkNumber || '',
            payee: check.payee || '',
            amount: check.amount || 0,
            checkDate: check.date || '',
            accountNumber: check.account || check.accountNumber || selectedAccount, // Handle both field names
            cleared: check.cleared || false,
            reconcileDate: check.reconcileDate || '',
            memo: check.memo || '',
            voidFlag: check.voidFlag || false,
            daysOutstanding: check.daysOutstanding || 0,
            entryType: check.entryType || '', // D = Deposit, C = Check
            rowIndex: check._rowIndex || index, // Keep row index for DBF updates
            originalCheck: check // Keep original data for updates
          }
        })
        
        console.log('BankReconciliation: Sample check CIDCHEC IDs:', processedChecks.slice(0, 3).map(c => ({id: c.id, cidchec: c.cidchec})))
        
        setChecks(processedChecks)
        
        // Debug first few checks to see their properties
        console.log('Sample checks:', processedChecks.slice(0, 3).map(c => ({
          id: c.id,
          checkNumber: c.checkNumber,
          cleared: c.cleared,
          voidFlag: c.voidFlag,
          accountNumber: c.accountNumber
        })))
      }
    } catch (err) {
      console.error('Error loading checks data:', err)
      setError('Failed to load checks data: ' + err.message)
      setChecks([])
    } finally {
      setLoading(false)
    }
  }

  // Load last reconciliation data from CHECKREC.dbf
  const loadLastReconciliation = async () => {
    try {
      setLoadingLastRec(true)
      console.log('Loading last reconciliation for account:', selectedAccount)
      
      const response = await GetLastReconciliation(companyName, selectedAccount)
      console.log('Last reconciliation response:', response)
      
      if (response && response.status === 'success') {
        setLastReconciliation(response)
        
        // Pre-populate the reconciliation form with the last data
        // The ending balance from last reconciliation becomes the beginning balance for new reconciliation (only if not already set)
        if (response.ending_balance && !beginningBalance) {
          setBeginningBalance(response.ending_balance.toString())
        }
        
        // Pre-populate statement date with end of following month from last reconciliation (only if not already set)
        const lastDate = response.statement_date || response.date || response.reconcile_date
        if (lastDate && !statementDate) {
          // Parse the date string safely (assumes YYYY-MM-DD format from backend)
          const dateStr = lastDate.split('T')[0] // Remove time if present
          const [year, month, day] = dateStr.split('-').map(Number)
          const lastStatementDate = new Date(year, month - 1, day) // month is 0-based in JS
          
          console.log('Last statement date from response:', lastDate)
          console.log('Parsed date components:', { year, month: month-1, day })
          console.log('Last statement date object:', lastStatementDate)
          
          // Calculate the end of the following month
          const nextMonth = lastStatementDate.getMonth() + 1 // Get next month (0-based)
          const nextYear = nextMonth > 11 ? lastStatementDate.getFullYear() + 1 : lastStatementDate.getFullYear()
          const adjustedMonth = nextMonth > 11 ? 0 : nextMonth
          
          // Get last day of the next month
          const endOfFollowingMonth = new Date(nextYear, adjustedMonth + 1, 0)
          
          console.log('Next month calculations:', { nextMonth, nextYear, adjustedMonth })
          console.log('End of following month:', endOfFollowingMonth)
          
          // Format as YYYY-MM-DD for the date input
          const formattedDate = endOfFollowingMonth.getFullYear() + '-' + 
                               String(endOfFollowingMonth.getMonth() + 1).padStart(2, '0') + '-' +
                               String(endOfFollowingMonth.getDate()).padStart(2, '0')
          
          console.log('Formatted date for input:', formattedDate)
          setStatementDate(formattedDate)
        }
        
        // Clear other statement fields for the new reconciliation only if they're not already set (but don't trigger unsaved changes)
        if (!statementBalance) setStatementBalance('')
        if (!statementCredits) setStatementCredits('')
        if (!statementDebits) setStatementDebits('')
        
        // Don't mark as unsaved changes since this is auto-population
        // setHasUnsavedChanges will be triggered by user input
      } else if (response && response.status === 'no_data') {
        setLastReconciliation(null)
        console.log('No reconciliation history found for this account')
        
        // If no last reconciliation, set statement date to end of current month
        const now = new Date()
        const endOfCurrentMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0)
        const formattedDate = endOfCurrentMonth.toISOString().split('T')[0]
        setStatementDate(formattedDate)
      }
    } catch (err) {
      console.error('Error loading last reconciliation:', err)
      setLastReconciliation(null)
    } finally {
      setLoadingLastRec(false)
    }
  }

  // Draft Reconciliation Management using SQLite
  const saveDraftReconciliation = async () => {
    if (!selectedAccount || !companyName) return
    // Allow saving draft even without statement date for partial saves
    if (!statementDate && !statementCredits && !statementDebits && draftSelectedChecks.size === 0) {
      console.log('Cannot save empty draft reconciliation')
      return
    }
    
    // Create detailed check selections with CIDCHEC IDs or composite IDs as fallback
    const selectedChecksDetails = Array.from(draftSelectedChecks).map(checkId => {
      const check = checks.find(c => c.id === checkId)
      if (check) {
        console.log('💾 Saving check with ID:', {
          checkId: check.id,
          checkNumber: check.checkNumber,
          cidchec: check.cidchec,
          cidchecType: typeof check.cidchec,
          usingCompositeId: !check.cidchec || check.cidchec.trim() === ''
        })
      }
      return check ? {
        cidchec: check.cidchec || check.id, // Use composite ID if CIDCHEC is empty
        checkNumber: check.checkNumber,
        amount: check.amount,
        payee: check.payee,
        checkDate: check.checkDate,
        rowIndex: check.rowIndex // Needed for DBF updates
      } : null
    }).filter(Boolean)
    
    const draftData = {
      account_number: selectedAccount,
      statement_date: statementDate ? statementDate.split('T')[0] : '', // Remove timestamp if present
      statement_balance: parseFloat(statementBalance) || 0,
      statement_credits: parseFloat(statementCredits) || 0,
      statement_debits: parseFloat(statementDebits) || 0,
      beginning_balance: parseFloat(beginningBalance) || 0,
      selected_checks: selectedChecksDetails
    }
    
    try {
      console.log('💾 Saving draft reconciliation to SQLite:', {
        account: selectedAccount,
        checksSelected: selectedChecksDetails.length,
        cidchecIds: selectedChecksDetails.map(c => c.cidchec),
        draftSelectedChecksSize: draftSelectedChecks.size,
        selectedChecksDetails: selectedChecksDetails
      })
      
      const result = await SaveReconciliationDraft(companyName, draftData)
      
      if (result.status === 'success') {
        setHasUnsavedChanges(false)
        console.log('✅ Draft reconciliation saved successfully with', selectedChecksDetails.length, 'selected checks')
      } else {
        console.error('Failed to save draft:', result)
        setError('Failed to save draft reconciliation')
      }
    } catch (err) {
      console.error('Failed to save draft reconciliation:', err)
      const errorMessage = err?.message || err?.toString() || 'Unknown error occurred'
      setError('Failed to save draft reconciliation: ' + errorMessage)
    }
  }

  const loadDraftReconciliation = async () => {
    if (!selectedAccount || !companyName) return
    
    try {
      console.log('Loading draft reconciliation from SQLite:', {
        company: companyName,
        account: selectedAccount
      })
      
      const result = await GetReconciliationDraft(companyName, selectedAccount)
      
      if (result.status === 'success' && result.draft) {
        const draftData = result.draft
        console.log('📋 Loading draft reconciliation from SQLite:', {
          account: draftData.account_number,
          selectedChecks: draftData.selected_checks?.length || 0,
          cidchecs: draftData.selected_checks?.map(c => c.cidchec) || [],
          rawSelectedChecks: draftData.selected_checks,
          firstCheck: draftData.selected_checks?.[0]
        })
        
        // Restore form data
        if (draftData.statement_date) setStatementDate(draftData.statement_date)
        if (draftData.statement_balance) setStatementBalance(draftData.statement_balance.toString())
        if (draftData.statement_credits) setStatementCredits(draftData.statement_credits.toString())
        if (draftData.statement_debits) setStatementDebits(draftData.statement_debits.toString())
        if (draftData.beginning_balance) setBeginningBalance(draftData.beginning_balance.toString())
        
        // Restore selected checks by matching CIDCHEC or composite IDs
        const selectedCIDCHECs = new Set()
        if (draftData.selected_checks && Array.isArray(draftData.selected_checks)) {
          draftData.selected_checks.forEach((check, index) => {
            console.log(`🔍 Processing saved check ${index}:`, {
              cidchec: check.cidchec,
              checkNumber: check.checkNumber || check.check_number,
              amount: check.amount,
              hasValidCidchec: !!(check.cidchec && check.cidchec.trim() !== '')
            })
            if (check.cidchec) {
              selectedCIDCHECs.add(check.cidchec)
            }
          })
        }
        console.log('🎯 Extracted saved IDs from draft:', Array.from(selectedCIDCHECs))
        
        // Wait for checks to be loaded, then match by CIDCHEC or composite ID
        const matchSelectedChecks = () => {
          const matchedCheckIds = new Set()
          checks.forEach(check => {
            // Match by CIDCHEC (if both have valid CIDCHEC) or by composite ID
            const shouldMatch = (check.cidchec && check.cidchec.trim() !== '' && selectedCIDCHECs.has(check.cidchec)) ||
                                selectedCIDCHECs.has(check.id)
            
            if (shouldMatch) {
              matchedCheckIds.add(check.id)
            }
          })
          
          console.log('🔗 Matched saved IDs to current checks:', {
            savedIDs: Array.from(selectedCIDCHECs),
            matchedIds: Array.from(matchedCheckIds),
            totalChecksLoaded: checks.length
          })
          
          if (matchedCheckIds.size > 0) {
            setDraftSelectedChecks(matchedCheckIds)
            setSelectedChecks(matchedCheckIds) // Also update checkbox state
            console.log('✅ Restored', matchedCheckIds.size, 'selected checks from draft')
          } else {
            console.log('⚠️ No ID matches found - selected checks not restored')
          }
        }
        
        if (checks.length > 0) {
          // Checks already loaded, match immediately
          matchSelectedChecks()
        } else {
          // Checks not loaded yet, set up to match when they are
          // Store the CIDCHEC IDs temporarily for matching after checks load
          setDraftReconciliation({...draftData, selectedCIDCHECs})
        }
        
        setHasUnsavedChanges(false)
        return true // Draft found and loaded
      } else if (result.status === 'no_draft') {
        console.log('No draft reconciliation found in SQLite')
        return false
      } else {
        console.error('Error loading draft:', result)
        return false
      }
      
    } catch (err) {
      console.error('Failed to load draft reconciliation:', err)
      return false
    }
  }

  const clearDraftReconciliation = async () => {
    if (!selectedAccount || !companyName) return
    
    try {
      console.log('Clearing draft reconciliation from SQLite')
      const result = await DeleteReconciliationDraft(companyName, selectedAccount)
      
      if (result.status === 'success') {
        setDraftReconciliation(null)
        setDraftSelectedChecks(new Set())
        setHasUnsavedChanges(false)
        console.log('Draft reconciliation cleared from SQLite')
      } else {
        console.error('Failed to clear draft:', result)
      }
    } catch (err) {
      console.error('Failed to clear draft reconciliation:', err)
      // Clear local state even if SQLite deletion fails
      setDraftReconciliation(null)
      setDraftSelectedChecks(new Set())
      setHasUnsavedChanges(false)
    }
  }

  // Calculate ending balance (called when user leaves input fields)
  const calculateEndingBalance = () => {
    const beginningBal = parseFloat(beginningBalance) || 0
    const credits = parseFloat(statementCredits) || 0
    const debits = parseFloat(statementDebits) || 0
    
    // Bank reconciliation formula: Beginning + Credits - Debits = Ending Balance
    const calculatedEnding = beginningBal + credits - debits
    const newBalance = calculatedEnding.toFixed(2)
    
    // Only update if the value actually changed to avoid infinite loops
    if (newBalance !== statementBalance) {
      setStatementBalance(newBalance)
    }
  }

  // Secondary auto-save for form fields only (disabled to avoid conflicts with main auto-save)
  // useEffect(() => {
  //   if (hasUnsavedChanges && selectedAccount) {
  //     const timeoutId = setTimeout(() => {
  //       saveDraftReconciliation()
  //     }, 2000) // Auto-save after 2 seconds of inactivity
  //     
  //     return () => clearTimeout(timeoutId)
  //   }
  // }, [hasUnsavedChanges, selectedAccount, statementDate, statementBalance, statementCredits, statementDebits, draftSelectedChecks])

  // Load draft when account changes
  useEffect(() => {
    if (companyName && selectedAccount) {
      loadDraftReconciliation()
    }
  }, [companyName, selectedAccount])

  // Handle ID matching when checks are loaded (supports both CIDCHEC and composite IDs)
  useEffect(() => {
    if (checks.length > 0 && draftReconciliation?.selectedCIDCHECs) {
      console.log('🔄 Matching saved IDs after checks loaded...', {
        checksLoaded: checks.length,
        savedIDs: draftReconciliation.selectedCIDCHECs?.length || 0
      })
      const selectedCIDCHECs = new Set(draftReconciliation.selectedCIDCHECs)
      const matchedCheckIds = new Set()
      
      checks.forEach(check => {
        // Match by CIDCHEC (if both have valid CIDCHEC) or by composite ID
        const shouldMatch = (check.cidchec && check.cidchec.trim() !== '' && selectedCIDCHECs.has(check.cidchec)) ||
                            selectedCIDCHECs.has(check.id)
        
        console.log('🔍 Checking ID match:', {
          checkId: check.id,
          checkNumber: check.checkNumber,
          currentCidchec: check.cidchec,
          cidchecType: typeof check.cidchec,
          shouldMatch: shouldMatch,
          matchType: shouldMatch ? (selectedCIDCHECs.has(check.cidchec) ? 'CIDCHEC' : 'composite') : 'none'
        })
        
        if (shouldMatch) {
          matchedCheckIds.add(check.id)
        }
      })
      
      console.log('Post-load ID matching:', {
        savedIDs: Array.from(selectedCIDCHECs),
        matchedIds: Array.from(matchedCheckIds),
        matchedCount: matchedCheckIds.size
      })
      
      setDraftSelectedChecks(matchedCheckIds)
      setSelectedChecks(matchedCheckIds) // Also update checkbox state
      
      // Clear the temporary ID data
      setDraftReconciliation(prev => prev ? {...prev, selectedCIDCHECs: undefined} : null)
    }
  }, [checks, draftReconciliation])

  // Bank accounts are now loaded from COA.dbf

  // Sort function for checks
  const sortChecks = (a, b) => {
    let aValue, bValue
    
    switch (sortField) {
      case 'checkNumber':
        aValue = a.checkNumber || ''
        bValue = b.checkNumber || ''
        break
      case 'checkDate':
        aValue = new Date(a.checkDate || '1900-01-01')
        bValue = new Date(b.checkDate || '1900-01-01')
        break
      case 'payee':
        aValue = a.payee || ''
        bValue = b.payee || ''
        break
      case 'amount':
        aValue = a.amount || 0
        bValue = b.amount || 0
        break
      case 'memo':
        aValue = a.memo || ''
        bValue = b.memo || ''
        break
      case 'cleared':
        aValue = a.cleared ? 1 : 0
        bValue = b.cleared ? 1 : 0
        break
      default:
        aValue = a[sortField] || ''
        bValue = b[sortField] || ''
    }
    
    if (sortDirection === 'asc') {
      return aValue < bValue ? -1 : aValue > bValue ? 1 : 0
    } else {
      return aValue > bValue ? -1 : aValue < bValue ? 1 : 0
    }
  }

  // Handle column sorting
  const handleSort = (field) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortDirection('asc')
    }
  }
  
  // Handle bank transaction sorting
  const handleBankSort = (field) => {
    if (bankSortField === field) {
      setBankSortDirection(bankSortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setBankSortField(field)
      setBankSortDirection('asc')
    }
  }
  
  // Handle check sorting (for side-by-side view)
  const handleCheckSort = (field) => {
    if (checkSortField === field) {
      setCheckSortDirection(checkSortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setCheckSortField(field)
      setCheckSortDirection('asc')
    }
  }

  // Render sortable column header
  const renderSortableHeader = (field, label, className = '') => (
    <TableHead 
      className={`cursor-pointer hover:bg-gray-50 ${className}`}
      onClick={() => handleSort(field)}
    >
      <div className="flex items-center gap-1">
        <span>{label}</span>
        {sortField === field && (
          sortDirection === 'asc' ? 
            <ChevronUp className="w-4 h-4" /> : 
            <ChevronDown className="w-4 h-4" />
        )}
      </div>
    </TableHead>
  )

  // Filter and sort checks based on current filters
  const filteredChecks = checks
    .filter(check => {
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
      
      // Transaction type filter
      if (showTransactionType !== 'all') {
        const entryType = check.entryType?.toUpperCase()
        if (showTransactionType === 'debits' && entryType !== 'C') return false
        if (showTransactionType === 'credits' && entryType !== 'D') return false
      }
      
      // Statement date filter - only show transactions through statement date
      if (limitToStatementDate && statementDate && check.checkDate > statementDate) return false
      
      return true
    })
    .sort(sortChecks)

  // Calculate reconciliation totals
  const calculateTotals = () => {
    const accountChecks = selectedAccount 
      ? checks.filter(check => check.accountNumber === selectedAccount)
      : checks

    const clearedChecks = accountChecks.filter(check => check.cleared)
    const unclearedChecks = accountChecks.filter(check => !check.cleared)
    
    // Get checks selected for reconciliation (draft)
    const selectedForReconciliation = Array.from(draftSelectedChecks).map(id => 
      checks.find(check => check.id === id)
    ).filter(Boolean)
    
    const clearedTotal = clearedChecks.reduce((sum, check) => sum + check.amount, 0)
    const unclearedTotal = unclearedChecks.reduce((sum, check) => sum + check.amount, 0)
    const selectedTotal = selectedForReconciliation.reduce((sum, check) => sum + check.amount, 0)
    
    // Calculate credits and debits from matched checks
    // Now using check data instead of bank transaction data
    const matchedChecks = matchedTransactions.filter(check => 
      selectedAccount ? check.account === selectedAccount : true
    )
    
    // For now, all checks are debits (withdrawals)
    // In the future, we might have deposits in the checks table too
    const selectedCredits = 0  // No deposits in checks table currently
    
    // All matched checks are debits
    const selectedDebits = matchedChecks
      .reduce((sum, check) => sum + Math.abs(check.amount || 0), 0)
    
    const beginningBal = parseFloat(beginningBalance) || 0
    const statementBal = parseFloat(statementBalance) || 0
    const stmtCredits = parseFloat(statementCredits) || 0
    const stmtDebits = parseFloat(statementDebits) || 0
    
    // Bank reconciliation calculations
    // 1. Calculated balance from selected items: Beginning + Selected Credits - Selected Debits
    const calculatedBalance = beginningBal + selectedCredits - selectedDebits
    
    // 2. Book balance after reconciling selected checks: Beginning - Selected Outstanding Checks
    const bookBalanceAfterReconciliation = beginningBal - selectedTotal
    
    // 3. Difference between statement balance and calculated balance from selected items
    const reconciliationDifference = statementBal - calculatedBalance
    
    // 4. Are we in balance? (difference should be zero or very close)
    const isInBalance = Math.abs(reconciliationDifference) < 0.01

    return {
      clearedCount: clearedChecks.length,
      unclearedCount: unclearedChecks.length,
      selectedCount: selectedForReconciliation.length,
      clearedTotal,
      unclearedTotal,
      selectedTotal,
      beginningBalance: beginningBal,
      statementBalance: statementBal,
      statementCredits: stmtCredits, // Input field values
      statementDebits: stmtDebits,   // Input field values
      selectedCredits,               // Selected deposit amounts
      selectedDebits,                // Selected check amounts
      calculatedBalance,
      bookBalanceAfterReconciliation,
      reconciliationDifference,
      isInBalance
    }
  }

  const totals = calculateTotals()

  // Toggle check selection for draft reconciliation (NO DBF updates)
  const toggleCheckSelection = (checkId) => {
    const newDraftSelected = new Set(draftSelectedChecks)
    if (newDraftSelected.has(checkId)) {
      newDraftSelected.delete(checkId)
    } else {
      newDraftSelected.add(checkId)
    }
    setDraftSelectedChecks(newDraftSelected)
    setHasUnsavedChanges(true)
    console.log('Check selection updated:', Array.from(newDraftSelected))
  }

  // Handler to unmatch a transaction
  const handleUnmatchTransaction = async (bankTxnId) => {
    try {
      const result = await UnmatchTransaction(bankTxnId)
      if (result.status === 'success') {
        // Reload matched transactions to reflect the change
        await loadMatchedTransactions()
        await loadBankTransactions()
        await loadChecksData()
      } else {
        setError('Failed to unmatch transaction')
      }
    } catch (err) {
      console.error('Error unmatching transaction:', err)
      setError('Failed to unmatch transaction: ' + (err.message || err.toString()))
    }
  }

  // Manual match function for bank transactions
  const handleManualMatch = async () => {
    if (!selectedBankTxn || selectedChecksForMatch.size === 0) {
      setError('Please select both a bank transaction and at least one check to match')
      return
    }
    
    setIsManualMatching(true)
    try {
      // Get the selected checks
      const selectedChecksList = Array.from(selectedChecksForMatch).map(id => 
        checks.find(c => c.id === id)
      ).filter(Boolean)
      
      if (selectedChecksList.length === 0) {
        setError('No valid checks selected')
        return
      }
      
      // TODO: Backend should support matching multiple checks to one bank transaction
      // For now, we'll match the first check
      const checkToMatch = selectedChecksList[0]
      
      console.log('Manual matching bank txn:', selectedBankTxn.id, 'to check:', checkToMatch.id)
      
      // Call the backend to create the match
      const result = await ManualMatchTransaction(
        selectedBankTxn.id, 
        checkToMatch.id, 
        checkToMatch._rowIndex || 0
      )
      
      if (result.status === 'success') {
        console.log('✅ Manual match successful')
        
        // Clear selections
        setSelectedBankTxn(null)
        setSelectedCheckForMatch(null)
        setSelectedChecksForMatch(new Set())
        
        // Reload data to reflect the match
        await loadBankTransactions()
        await loadMatchedTransactions()
        await loadChecksData()
      } else {
        setError('Failed to create manual match')
      }
    } catch (err) {
      console.error('Error creating manual match:', err)
      setError('Failed to create manual match: ' + (err.message || err.toString()))
    } finally {
      setIsManualMatching(false)
    }
    // For now, reload to refresh the view
    loadBankTransactions()
  }

  // Get unmatched bank transactions (not already matched to checks)
  const getUnmatchedBankTransactions = () => {
    const unmatched = bankTransactions.filter(transaction => !transaction.is_matched)
    
    // Apply sorting
    return unmatched.sort((a, b) => {
      let aVal, bVal
      
      switch(bankSortField) {
        case 'transaction_date':
          aVal = new Date(a.transaction_date)
          bVal = new Date(b.transaction_date)
          break
        case 'description':
          aVal = a.description || ''
          bVal = b.description || ''
          break
        case 'check_number':
          aVal = a.check_number || ''
          bVal = b.check_number || ''
          break
        case 'amount':
          aVal = a.amount
          bVal = b.amount
          break
        default:
          return 0
      }
      
      if (bankSortDirection === 'asc') {
        return aVal > bVal ? 1 : aVal < bVal ? -1 : 0
      } else {
        return aVal < bVal ? 1 : aVal > bVal ? -1 : 0
      }
    })
  }

  // Get available checks for matching (not already selected for reconciliation)
  const getAvailableChecks = () => {
    // Build comprehensive sets of matched check identifiers
    const matchedCheckIds = new Set()
    const matchedCidchecs = new Set()
    const matchedCheckNumbers = new Set()
    
    matchedTransactions.forEach(txn => {
      // Add all possible identifiers from matched transactions
      if (txn.id) matchedCheckIds.add(txn.id)
      if (txn.cidchec) matchedCidchecs.add(txn.cidchec)
      if (txn.check_number) matchedCheckNumbers.add(txn.check_number)
      if (txn.matched_check_id) matchedCheckIds.add(txn.matched_check_id)
      if (txn.checkNumber) matchedCheckNumbers.add(txn.checkNumber)
    })
    
    // Debug: log filtering details
    const clearedCount = checks.filter(c => c.cleared).length
    const draftSelectedCount = checks.filter(c => draftSelectedChecks.has(c.id)).length
    const voidCount = checks.filter(c => c.voidFlag).length
    const matchedCount = checks.filter(c => {
      return matchedCheckIds.has(c.id) || 
             matchedCidchecs.has(c.cidchec) ||
             matchedCheckNumbers.has(c.checkNumber)
    }).length
    
    console.log('getAvailableChecks filtering:', {
      totalChecks: checks.length,
      clearedChecks: clearedCount,
      voidedChecks: voidCount,
      matchedChecks: matchedCount,
      matchedCheckIdentifiers: {
        ids: matchedCheckIds.size,
        cidchecs: matchedCidchecs.size,
        checkNumbers: matchedCheckNumbers.size
      },
      draftSelected: draftSelectedCount,
      draftSelectedIds: Array.from(draftSelectedChecks).slice(0, 5) // Show first 5 IDs
    })
    
    // Start with checks for the selected account only
    let available = checks.filter(check => {
      // Check if this check is matched using any identifier
      const isMatched = matchedCheckIds.has(check.id) || 
                       matchedCidchecs.has(check.cidchec) ||
                       matchedCheckNumbers.has(check.checkNumber)
      
      // Filter out already selected, cleared, and matched checks
      // Note: voided checks are already filtered by the backend
      return !draftSelectedChecks.has(check.id) && !check.cleared && !isMatched
    })
    
    // Apply statement date filter if enabled
    if (limitToStatementDate && statementDate) {
      available = available.filter(check => {
        if (!check.checkDate) return true // Include checks without dates
        
        // Parse dates for proper comparison
        const checkDateObj = new Date(check.checkDate)
        const statementDateObj = new Date(statementDate)
        
        // Check if dates are valid
        if (isNaN(checkDateObj.getTime())) return true // Include if date is invalid
        if (isNaN(statementDateObj.getTime())) return true // Include if statement date is invalid
        
        // Compare dates
        return checkDateObj <= statementDateObj
      })
    }
    
    // Apply sorting
    return available.sort((a, b) => {
      let aVal, bVal
      
      switch(checkSortField) {
        case 'checkNumber':
          aVal = a.checkNumber || ''
          bVal = b.checkNumber || ''
          break
        case 'checkDate':
          aVal = new Date(a.checkDate || 0)
          bVal = new Date(b.checkDate || 0)
          break
        case 'payee':
          aVal = a.payee || ''
          bVal = b.payee || ''
          break
        case 'amount':
          aVal = parseFloat(a.amount) || 0
          bVal = parseFloat(b.amount) || 0
          break
        default:
          return 0
      }
      
      if (checkSortDirection === 'asc') {
        return aVal > bVal ? 1 : aVal < bVal ? -1 : 0
      } else {
        return aVal < bVal ? 1 : aVal > bVal ? -1 : 0
      }
    })
  }

  // Commit reconciliation (final step - commits SQLite draft and updates DBF files)
  const commitReconciliation = async () => {
    if (draftSelectedChecks.size === 0) {
      alert('Please select checks to reconcile before committing.')
      return
    }

    if (!confirm('This will permanently update check records and create a reconciliation entry. Are you sure?')) {
      return
    }

    setReconciliationInProgress(true)
    try {
      // First, save any unsaved changes to draft
      await saveDraftReconciliation()

      // Commit the draft in SQLite (this will change status from 'draft' to 'committed')
      console.log('Committing reconciliation via SQLite API')
      const commitResult = await CommitReconciliation(companyName, selectedAccount)
      
      if (commitResult.status !== 'success') {
        throw new Error(commitResult.message || 'Failed to commit reconciliation')
      }

      console.log('Reconciliation committed successfully in SQLite:', commitResult)

      // TODO: Update DBF files (CHECKS.dbf and CHECKREC.dbf)
      // This will be implemented later when we add DBF sync functionality
      // For now, the reconciliation is stored in SQLite as committed

      // Clear local draft state and reload data
      setDraftSelectedChecks(new Set())
      setDraftReconciliation(null)
      setHasUnsavedChanges(false)
      
      // Clear form
      setStatementDate('')
      setStatementBalance('')
      setStatementCredits('')
      setStatementDebits('')
      
      // Reload checks data
      await loadChecksData()

      alert('Reconciliation committed successfully!')

    } catch (err) {
      console.error('Error committing reconciliation:', err)
      setError('Failed to commit reconciliation: ' + err.message)
    } finally {
      setReconciliationInProgress(false)
    }
  }

  // Bulk select checks for draft reconciliation
  const bulkSelectChecks = () => {
    if (selectedChecks.size === 0) return
    
    const newDraftSelected = new Set([...draftSelectedChecks, ...selectedChecks])
    setDraftSelectedChecks(newDraftSelected)
    setSelectedChecks(new Set()) // Clear UI selection
    setHasUnsavedChanges(true)
  }

  // Load bank transactions from SQLite
  const loadBankTransactions = async () => {
    if (!companyName || !selectedAccount) return
    
    try {
      setLoadingBankTransactions(true)
      console.log('Loading bank transactions for account:', selectedAccount)
      
      const result = await GetBankTransactions(companyName, selectedAccount, '')
      console.log('Bank transactions result:', result)
      
      if (result.status === 'success') {
        setBankTransactions(result.transactions || [])
        setShowSideBySide(result.transactions?.length > 0)
      } else {
        setBankTransactions([])
        setShowSideBySide(false)
      }
    } catch (err) {
      console.error('Error loading bank transactions:', err)
      setBankTransactions([])
      setShowSideBySide(false)
    } finally {
      setLoadingBankTransactions(false)
    }
  }

  // Load matched transactions (returns check data with bank match info)
  const loadMatchedTransactions = async () => {
    if (!companyName || !selectedAccount) return
    
    try {
      setLoadingMatched(true)
      const result = await GetMatchedTransactions(companyName, selectedAccount)
      if (result.status === 'success') {
        const matched = result.checks || []
        setMatchedTransactions(matched)
        
        // Debug: Log matched check IDs
        console.log('Loaded matched transactions:', {
          count: matched.length,
          sampleIds: matched.slice(0, 5).map(m => ({
            id: m.id,
            cidchec: m.cidchec,
            matched_check_id: m.matched_check_id,
            check_number: m.check_number
          }))
        })
      } else {
        setMatchedTransactions([])
      }
    } catch (err) {
      console.error('Error loading matched transactions:', err)
      setMatchedTransactions([])
    } finally {
      setLoadingMatched(false)
    }
  }
  
  // Unmatch a transaction (uses bank_txn_id from the check data)
  const handleUnmatch = async (checkData) => {
    try {
      // Use the bank_txn_id from the matched check data
      const bankTxnId = checkData.bank_txn_id
      if (!bankTxnId) {
        console.error('No bank transaction ID found for this check')
        return
      }
      
      const result = await UnmatchTransaction(bankTxnId)
      if (result.status === 'success') {
        // Reload both matched and unmatched transactions
        await loadMatchedTransactions()
        await loadBankTransactions()
      }
    } catch (err) {
      console.error('Error unmatching transaction:', err)
    }
  }
  
  // Bulk unmatch selected transactions
  const handleBulkUnmatch = async () => {
    if (selectedMatchedTxns.size === 0) return
    
    setIsBulkUnmatching(true)
    try {
      // Get bank_txn_id for each selected check
      const promises = Array.from(selectedMatchedTxns).map(checkId => {
        const check = matchedTransactions.find(c => c.id === checkId)
        if (check && check.bank_txn_id) {
          return UnmatchTransaction(check.bank_txn_id)
        }
        return Promise.resolve()
      })
      
      await Promise.all(promises)
      
      // Clear selections and reload
      setSelectedMatchedTxns(new Set())
      await loadMatchedTransactions()
      await loadBankTransactions()
      
      console.log(`✅ Successfully unmatched ${selectedMatchedTxns.size} transactions`)
    } catch (err) {
      console.error('Error bulk unmatching transactions:', err)
      setError('Failed to unmatch some transactions')
    } finally {
      setIsBulkUnmatching(false)
    }
  }
  
  // Toggle selection for a matched transaction
  const toggleMatchedSelection = (transactionId) => {
    const newSelection = new Set(selectedMatchedTxns)
    if (newSelection.has(transactionId)) {
      newSelection.delete(transactionId)
    } else {
      newSelection.add(transactionId)
    }
    setSelectedMatchedTxns(newSelection)
  }
  
  // Select/deselect all visible matched transactions
  const toggleSelectAllMatched = (checked) => {
    if (checked) {
      const visibleIds = matchedTransactions
        .filter(check => 
          !matchedSearchTerm || 
          check.check_number?.includes(matchedSearchTerm) ||
          check.payee?.toLowerCase().includes(matchedSearchTerm.toLowerCase()) ||
          check.id?.includes(matchedSearchTerm)
        )
        .map(check => check.id)
      setSelectedMatchedTxns(new Set(visibleIds))
    } else {
      setSelectedMatchedTxns(new Set())
    }
  }
  
  // Load import history
  const loadImportHistory = async () => {
    if (!companyName || !selectedAccount) return
    
    try {
      setLoadingHistory(true)
      const history = await GetRecentBankStatements(companyName, selectedAccount)
      setImportHistory(history || [])
    } catch (err) {
      console.error('Error loading import history:', err)
      setImportHistory([])
    } finally {
      setLoadingHistory(false)
    }
  }

  // Delete an import - opens confirmation dialog
  const handleDeleteImport = (importBatchId) => {
    console.log('Opening delete confirmation for:', importBatchId)
    setImportToDelete(importBatchId)
    setDeleteConfirmOpen(true)
  }

  // Confirm and execute deletion
  const confirmDelete = async () => {
    if (!importToDelete) return
    
    try {
      console.log('Calling DeleteBankStatement with:', companyName, importToDelete)
      await DeleteBankStatement(companyName, importToDelete)
      console.log('Delete successful')
      setDeleteConfirmOpen(false)
      setImportToDelete(null)
      await loadImportHistory()
      await loadBankTransactions()
    } catch (err) {
      console.error('Error deleting import:', err)
      setDeleteConfirmOpen(false)
      setImportToDelete(null)
    }
  }

  // Show import history when dialog opens
  useEffect(() => {
    if (showImportHistory) {
      loadImportHistory()
    }
  }, [showImportHistory, selectedAccount])

  // CSV Import functions
  const handleCSVUpload = async (file) => {
    if (!file || !selectedAccount) return
    
    setCsvUploading(true)
    setCsvError(null)
    
    try {
      const csvContent = await file.text()
      console.log('📂 Processing CSV file:', file.name, 'Size:', csvContent.length, 'chars')
      
      // Use new ImportBankStatement function that persists to SQLite
      const result = await ImportBankStatement(companyName, csvContent, selectedAccount)
      console.log('🎯 Import Bank Statement Result:', result)
      
      if (result.status === 'success') {
        setCsvParseResult(result)
        setCsvMatches(result.matches || [])
        console.log(`✅ Successfully imported ${result.totalTransactions} transactions. Click 'Run Matching' to match with checks.`)
        
        // Reload bank transactions to show imported data
        await loadBankTransactions()
      } else {
        throw new Error(result.error || 'Failed to import bank statement')
      }
    } catch (err) {
      console.error('❌ CSV Upload error:', err)
      setCsvError('Failed to process CSV: ' + (err.message || err.toString()))
    } finally {
      setCsvUploading(false)
    }
  }
  
  // Handler for running matching algorithm
  const handleRunMatching = async (skipDialog = false) => {
    if (!selectedAccount) {
      setError('Please select a bank account first')
      return
    }
    
    // Show options dialog if not skipping
    if (!skipDialog) {
      setShowMatchingOptions(true)
      return
    }
    
    setIsRunningMatch(true)
    setError(null)
    setShowMatchingOptions(false)
    
    try {
      console.log('🔄 Running matching algorithm for account:', selectedAccount)
      
      // Prepare options based on user selection
      const options = {
        limitToStatementDate: matchingDateOption === 'statement',
        statementDate: matchingDateOption === 'statement' ? statementDate : null
      }
      
      const result = await RunMatching(companyName, selectedAccount, options)
      console.log('✅ Matching complete:', result)
      
      setMatchResult(result)
      
      // Show success message
      if (result.totalMatched > 0) {
        console.log(`✅ Successfully matched ${result.totalMatched} out of ${result.totalProcessed} transactions`)
      } else {
        console.log('ℹ️ No transactions matched')
      }
      
      // Reload data to show matched transactions
      await loadBankTransactions()
      await loadMatchedTransactions()
    } catch (err) {
      console.error('❌ Error running matching:', err)
      setError('Failed to run matching: ' + (err.message || err.toString()))
    } finally {
      setIsRunningMatch(false)
    }
  }
  
  // Handler for clearing matches and re-running
  const handleRefreshMatching = async (skipDialog = false) => {
    if (!selectedAccount) {
      setError('Please select a bank account first')
      return
    }
    
    // Show options dialog if not skipping
    if (!skipDialog) {
      setShowMatchingOptions(true)
      return
    }
    
    setIsRefreshing(true)
    setError(null)
    setShowMatchingOptions(false)
    
    try {
      console.log('🔄 Clearing and re-running matches for account:', selectedAccount)
      
      // Prepare options based on user selection
      const options = {
        limitToStatementDate: matchingDateOption === 'statement',
        statementDate: matchingDateOption === 'statement' ? statementDate : null
      }
      
      const result = await ClearMatchesAndRerun(companyName, selectedAccount, options)
      console.log('✅ Refresh complete:', result)
      
      setMatchResult(result)
      
      // Show success message
      if (result.totalMatched > 0) {
        console.log(`✅ Successfully re-matched ${result.totalMatched} out of ${result.totalProcessed} transactions`)
      } else {
        console.log('ℹ️ No transactions matched after refresh')
      }
      
      // Reload data to show updated matches
      await loadBankTransactions()
      await loadMatchedTransactions()
      
      // Clear any draft reconciliation data
      setDraftSelectedChecks(new Set())
      setHasUnsavedChanges(false)
    } catch (err) {
      console.error('❌ Error refreshing matches:', err)
      setError('Failed to refresh matches: ' + (err.message || err.toString()))
    } finally {
      setIsRefreshing(false)
    }
  }

  const handleConfirmCSVMatches = async () => {
    const confirmedMatches = csvMatches.filter(match => match.confirmed && match.matchedCheck)
    const confirmedCheckIds = new Set(confirmedMatches.map(match => match.matchedCheck.id))
    
    console.log('🔒 Confirming', confirmedCheckIds.size, 'CSV matches')
    
    // Add confirmed checks to draft selection
    const newDraftSelected = new Set([...draftSelectedChecks, ...confirmedCheckIds])
    setDraftSelectedChecks(newDraftSelected)
    setSelectedChecks(confirmedCheckIds) // Also update checkbox state
    setHasUnsavedChanges(true)
    
    // Close CSV import dialog
    setCsvImportOpen(false)
    setCsvParseResult(null)
    setCsvMatches([])
    
    // Reload bank transactions to reflect updated match status
    await loadBankTransactions()
    
    console.log('✅ Applied', confirmedCheckIds.size, 'matches to reconciliation')
  }

  const toggleCSVMatch = (matchIndex) => {
    const updatedMatches = [...csvMatches]
    updatedMatches[matchIndex].confirmed = !updatedMatches[matchIndex].confirmed
    setCsvMatches(updatedMatches)
  }

  // Format currency
  const formatCurrency = (amount) => {
    const numValue = parseFloat(amount) || 0
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(numValue)
  }

  // Format date
  const formatDate = (dateStr) => {
    if (!dateStr) return ''
    
    // Convert to string if it's not already
    const dateString = String(dateStr)
    
    // Handle format like "11/14 00:00:00 +0000U/2024" 
    // This appears to be MM/DD HH:MM:SS +ZZZZUYYYY
    if (dateString.includes(' ') && dateString.includes('U/')) {
      // Extract month/day from start
      const parts = dateString.split(' ')
      const monthDay = parts[0] // "11/14"
      
      // Extract year from the end after "U/"
      const yearMatch = dateString.match(/U\/(\d{4})/)
      const year = yearMatch ? yearMatch[1] : new Date().getFullYear()
      
      // Return MM/DD/YYYY format
      return `${monthDay}/${year}`
    }
    
    // Handle ISO date format (2025-01-02T00:00:00Z)
    if (dateString.includes('T')) {
      const datePart = dateString.split('T')[0]
      const [year, month, day] = datePart.split('-')
      return `${month}/${day}/${year}`
    }
    
    // Handle MM/DD/YYYY format
    if (dateString.match(/^\d{1,2}\/\d{1,2}\/\d{4}$/)) {
      return dateString
    }
    
    // Handle YYYY-MM-DD format
    if (dateString.match(/^\d{4}-\d{2}-\d{2}$/)) {
      const [year, month, day] = dateString.split('-')
      return `${month}/${day}/${year}`
    }
    
    // Return original if no match
    return dateString
  }

  // Format number for input display (preserves partial decimal entries)
  const formatNumberForInput = (value) => {
    if (!value || value === '0') return ''
    // If the value ends with a decimal point or has trailing zeros after decimal, preserve it
    if (typeof value === 'string' && (value.endsWith('.') || value.match(/\.\d*$/))) {
      return value
    }
    return parseFloat(value).toString()
  }

  // Parse input value to number
  const parseInputNumber = (value) => {
    if (!value || value === '') return 0
    return parseFloat(value) || 0
  }

  // Format transaction type for display
  const formatTransactionType = (entryType) => {
    switch (entryType?.toUpperCase()) {
      case 'D':
        return 'Deposit'
      case 'C':
        return 'Check'
      default:
        return entryType || 'Unknown'
    }
  }

  // Get transaction type badge component
  const getTransactionTypeBadge = (entryType) => {
    switch (entryType?.toUpperCase()) {
      case 'D':
        return (
          <Badge variant="default" className="bg-green-100 text-green-800">
            Deposit
          </Badge>
        )
      case 'C':
        return (
          <Badge variant="secondary" className="bg-blue-100 text-blue-800">
            Check
          </Badge>
        )
      default:
        return (
          <Badge variant="outline">
            {entryType || 'Unknown'}
          </Badge>
        )
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
                  Account: {preSelectedAccount} - {bankAccounts.find(a => a.accountNumber === preSelectedAccount)?.accountName || 'Loading...'}
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
                      <p className="text-lg font-semibold">{lastReconciliation.date_string || 'N/A'}</p>
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
                    onChange={(e) => {
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
                    onChange={(e) => {
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
                    onChange={(e) => {
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
                    onClick={handleRunMatching} 
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
                        <Check className="w-4 h-4 mr-2" />
                        Run Matching
                      </>
                    )}
                  </Button>
                  <Button 
                    onClick={handleRefreshMatching} 
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
                      onChange={(e) => handleCSVUpload(e.target.files[0])}
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
                          Successfully imported {csvParseResult.totalTransactions} transactions
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
                      <Check className="w-4 h-4 mr-2" />
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
                        <Check className="w-4 h-4 mr-2" />
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
                                  <Check className="w-3 h-3 mr-1" />
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
                          console.log('Date filter toggled:', { 
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
                              // Clear from formData as well
                              setFormData(prev => ({
                                ...prev,
                                selectedChecks: []
                              }))
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
                        <Check className="w-4 h-4 mr-2" />
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
                  <Label className="text-xs">Transaction Type</Label>
                  <select
                    value={showTransactionType}
                    onChange={(e) => setShowTransactionType(e.target.value)}
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
                    onCheckedChange={setLimitToStatementDate}
                  />
                  <Label htmlFor="limit-statement-date" className="text-xs">Through Statement Date</Label>
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
                            console.log('✅ CHECKBOX CLICKED - Draft selected checks now:', newDraftSelected.size)
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
                      ? Math.floor((new Date() - new Date(check.checkDate)) / (1000 * 60 * 60 * 24))
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
                      console.log('Import record:', imp);
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
                                console.log('Delete clicked for:', imp.import_batch_id);
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
                                e.target.style.backgroundColor = '#b91c1c';
                              }}
                              onMouseOut={(e) => {
                                e.target.style.backgroundColor = '#dc2626';
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
              <Check className="w-4 h-4 mr-2" />
              Run Matching
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}