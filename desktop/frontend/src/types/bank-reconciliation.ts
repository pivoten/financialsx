// Type definitions for Bank Reconciliation module

export interface User {
  id: number
  username: string
  email?: string
  role_name: string
  is_root: boolean
  company_name: string
  permissions?: string[]
}

export interface BankAccount {
  account_number: string
  account_name: string
  account_description?: string
  account_type?: number
  balance?: number
  gl_balance?: number
  outstanding_checks_total?: number
  outstanding_checks_count?: number
  bank_balance?: number
  is_bank_account?: boolean
  // Additional properties used in code
  accountNumber?: string
  accountName?: string
}

export interface Check {
  id: string
  cidchec?: string
  checkNumber: string
  checkDate: string
  payee: string
  amount: number
  cleared: boolean
  void?: boolean
  accountNumber: string
  daysOutstanding?: number
  batchNumber?: string
  rowIndex?: number
  type?: 'check' | 'deposit'
  memo?: string
  entryType?: string
  reconcileDate?: string
}

export interface SelectedCheck {
  cidchec: string
  checkNumber: string
  amount: number
  payee: string
  checkDate: string
  rowIndex?: number
}

export interface ReconciliationDraft {
  id?: number
  company_name: string
  account_number: string
  reconcile_date?: string
  statement_date: string
  beginning_balance: number
  ending_balance?: number
  statement_balance: number
  statement_credits: number
  statement_debits: number
  selected_checks: SelectedCheck[]
  selected_checks_json?: string
  status: 'draft' | 'committed' | 'archived'
  created_by?: string
  created_at?: string
  updated_at?: string
  committed_at?: string
  extended_data?: Record<string, any>
}

export interface LastReconciliation {
  reconcile_date: string
  statement_date: string
  ending_balance: number
  statement_balance: number
  created_by: string
  created_at: string
  committed_at: string
  status: string
  // Additional properties used in code
  date_string?: string
  beginning_balance?: number
  cleared_count?: number
  cleared_amount?: number
}

export interface BankTransaction {
  id?: string
  transaction_id?: string
  account_number: string
  transaction_date: string
  description: string
  amount: number
  type: 'debit' | 'credit'
  check_number?: string
  reference?: string
  matched_check_id?: string
  confidence_score?: number
  import_batch_id?: string
  source?: string
}

export interface BankStatement {
  id: string
  statement_id: string
  account_number: string
  import_date: string
  transaction_count: number
  total_debits: number
  total_credits: number
  start_date: string
  end_date: string
  source: string
  imported_by: string
  // Additional properties used in code
  matched_count?: number
  import_batch_id?: string
  statement_date?: string
}

export interface MatchedTransaction {
  bank_transaction: BankTransaction
  matched_check?: Check
  confidence_score: number
  match_reason?: string
  // Additional properties used in code
  confirmed?: boolean
  id?: string
  check_date?: string
  bank_description?: string
  check_number?: string
  payee?: string
  amount?: number
  match_confidence?: number
  bank_txn_id?: string
}

export interface CSVParseResult {
  success: boolean
  transactions?: BankTransaction[]
  statement?: BankStatement
  matches?: MatchedTransaction[]
  error?: string
  columnMapping?: Record<string, string>
}

export interface ReconciliationTotals {
  statementCredits: number
  statementDebits: number
  calculatedBalance: number
  balanceDifference: number
  selectedCheckCount: number
  selectedDepositCount: number
  // Additional properties used in code
  beginningBalance?: number
  selectedCredits?: number
  selectedDebits?: number
  statementBalance?: number
  isInBalance?: boolean
  reconciliationDifference?: number
  selectedCount?: number
}

export interface BankReconciliationProps {
  companyName: string
  currentUser: User
  preSelectedAccount?: string
  onBack?: () => void
}

export interface MatchingOptions {
  limitToStatementDate: boolean
  statementDate?: string
}