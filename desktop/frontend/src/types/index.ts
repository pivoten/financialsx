// Comprehensive type definitions for the application

// User and Auth types
export interface User {
  id: number
  username: string
  email?: string
  role_name: string
  is_root: boolean
  company_name: string
  company_path?: string
  permissions?: string[]
  role_id?: number
  is_active?: boolean
  created_at?: string
  last_login?: string
  [key: string]: any // Allow additional dynamic properties
}

export interface Role {
  id: number
  role_name?: string
  name?: string
  display_name?: string
  description?: string
  permissions?: string[]
  [key: string]: any // Allow additional dynamic properties
}

// Company types
export interface Company {
  name: string
  display_name: string
  path: string
  address?: string
  city?: string
  state?: string
  zip?: string
  selected?: boolean
}

// DBF and table types
export interface DBFData {
  columns: string[]
  rows: any[][]
  metadata?: {
    total_rows: number
    file_size?: number
    last_modified?: string
  }
}

export interface DBFColumn {
  name: string
  type: string
  length: number
  decimal_places?: number
}

export interface DBFRecord {
  [key: string]: any
}

// Banking types
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
  last_updated?: string
  freshness?: 'fresh' | 'aging' | 'stale'
  // Additional properties used in BankingSection
  id?: number
  name?: string
  accountNumber?: string
  bank?: string
  type?: string
  status?: string
  uncleared_deposits?: number
  uncleared_checks?: number
  deposit_count?: number
  check_count?: number
  outstanding_total?: number
  is_stale?: boolean
}

export interface Balance {
  account_number: string
  account_name: string
  account_type: number
  gl_balance: number
  outstanding_checks_total: number
  outstanding_checks_count: number
  bank_balance: number
  is_active: boolean
  is_bank_account: boolean
  created_at: string
  updated_at: string
  gl_last_updated: string
  outstanding_checks_last_updated: string
  metadata?: Record<string, any>
  freshness?: 'fresh' | 'aging' | 'stale'
  // Additional properties found in the codebase
  outstanding_total?: number
  outstanding_count?: number
  is_stale?: boolean
  gl_freshness?: 'fresh' | 'aging' | 'stale'
  checks_freshness?: 'fresh' | 'aging' | 'stale'
  checks_last_updated?: string
  gl_age_hours?: number
  checks_age_hours?: number
  [key: string]: any // Allow additional dynamic properties
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
}

// Audit types
export interface AuditResult {
  missing_gl_entries: any[]
  mismatched_amounts: any[]
  matched_entries: number
  total_checks: number
  summary?: {
    total_missing_amount: number
    total_mismatch_amount: number
  }
}

export interface AuditResultExtended {
  missing_entries: AuditEntry[]
  mismatched_amounts: AuditEntry[]
  matched_entries?: number
  total_checks?: number
  summary?: {
    total_missing_amount: number
    total_mismatch_amount: number
  }
}

export interface AuditEntry {
  cidchec: string
  check_number: string
  check_date: string
  payee: string
  check_amount: number
  gl_amount?: number
  amount?: number
  difference?: number
  account_number?: string
  batch_number?: string
  [key: string]: any // Index signature for dynamic property access
}

export interface BankAuditResult {
  account_number: string
  account_name: string
  issue_type: string
  gl_balance: number
  outstanding_checks: number
  outstanding_count: number
  bank_balance: number
  difference: number
  percentage_diff: number
  details: string
}

// API Response types
export interface APIResponse<T = any> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

// Form event types
export type FormEvent = React.FormEvent<HTMLFormElement>
export type ChangeEvent<T = HTMLInputElement> = React.ChangeEvent<T>
export type MouseEvent<T = HTMLButtonElement> = React.MouseEvent<T>

// Table column types
export interface TableColumn<T = any> {
  accessor: string
  header: string
  sortable?: boolean
  type?: 'string' | 'number' | 'date' | 'boolean'
  render?: (value: any, row: T, index: number) => React.ReactNode
  cellClassName?: string
  headerClassName?: string
}

// Filter types
export interface FilterConfig {
  key: string
  label: string
  placeholder?: string
  defaultValue?: string
  options: FilterOption[]
  filterFn?: (row: any, value: string) => boolean
}

export interface FilterOption {
  value: string
  label: string
}

// UI Component types
export type BadgeVariant = 'default' | 'destructive' | 'outline' | 'secondary'

// Select component types (for UI libraries like RadixUI, etc.)
export interface SelectProps {
  value?: string
  onValueChange?: (value: string) => void
  defaultValue?: string
  disabled?: boolean
  children?: React.ReactNode
}

export interface SelectItemProps {
  value: string
  disabled?: boolean
  children?: React.ReactNode
}

// State report types
export interface StateReport {
  id: string
  name: string
  status: 'draft' | 'pending' | 'completed' | 'failed'
  created_at: string
  updated_at: string
  parameters?: Record<string, any>
}

// Database test types
export interface DatabaseTestResult {
  success: boolean
  message: string
  details?: any
  timestamp?: string
  data?: any[]
  database?: string
  rowCount?: number
  executionTime?: string
  raw?: any
}

// Component prop types
export interface ComponentProps {
  currentUser?: User | null
  companyName?: string
  className?: string
}

// Window extension for Wails
declare global {
  interface Window {
    go?: {
      main?: {
        App?: any
      }
    }
    runtime?: any
  }
}

// Supabase types
export interface SupabaseUser {
  id: string
  email?: string
  user_metadata?: Record<string, any>
}

export interface SupabaseSession {
  access_token: string
  refresh_token: string
  user: SupabaseUser
}

// State types
export interface AppState {
  user: User | null
  currentCompany: Company | null
  companies: Company[]
  loading: boolean
  error: string | null
}

// Export all types from bank-reconciliation
export * from './bank-reconciliation'