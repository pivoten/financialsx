import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Button } from './ui/button'
import { AlertCircle, CheckCircle, Clock, FileSearch, Loader2, TrendingUp, TrendingDown, Calendar, ShieldCheck, AlertTriangle, Calculator, Building2, FileText, DollarSign, Download, Search, XCircle, Menu, ChevronDown, Activity, Users, Package, CreditCard, TrendingUp as TrendUp, BarChart3 } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle } from './ui/alert'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs'
import { ScrollArea } from './ui/scroll-area'
import { Badge } from './ui/badge'
import FollowBatchNumber from './FollowBatchNumber'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
} from './ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from './ui/dialog'

interface AuditToolsProps {
  currentUser: any
  companyName: string
}

// Define audit categories and their items
const auditCategories = [
  {
    id: 'gl-audits',
    name: 'General Ledger Audits',
    icon: Calculator,
    items: [
      { id: 'gl-validation', name: 'GL Balance Validation', description: 'Verify debits equal credits' },
      { id: 'gl-activity', name: 'GL Activity Summary', description: 'Check period and year fields' },
      { id: 'year-analysis', name: 'Year-by-Year Analysis', description: 'Analyze balances by year' },
      { id: 'duplicate-detection', name: 'Duplicate Detection', description: 'Find duplicate GL entries' },
      { id: 'account-analysis', name: 'Account Analysis', description: 'Deep dive into specific accounts' }
    ]
  },
  {
    id: 'check-audits',
    name: 'Check Audits',
    icon: CreditCard,
    items: [
      { id: 'check-batch', name: 'Check Batch Audit', description: 'Compare checks to GL entries' },
      { id: 'check-gl-matching', name: 'Check-GL Matching Audit', description: 'Find checks without GL entries and vice versa' },
      { id: 'duplicate-cidchec', name: 'Duplicate CIDCHEC Audit', description: 'Find checks with duplicate CIDCHEC values' },
      { id: 'void-verification', name: 'Void Verification Audit', description: 'Verify voided checks have proper settings' },
      { id: 'payee-cid-verification', name: 'Payee-CID Verification', description: 'Verify check payees match investor/vendor CID records' },
      { id: 'outstanding-checks', name: 'Outstanding Checks Audit', description: 'Verify uncleared checks' },
      { id: 'void-checks', name: 'Void Checks Review', description: 'Review voided check entries' },
      { id: 'check-sequence', name: 'Check Sequence Audit', description: 'Find missing check numbers' },
      { id: 'follow-batch-number', name: 'Follow Batch Number', description: 'Search batch number across CHECKS, GLMASTER, APPURCHD, and APPMTHDR' }
    ]
  },
  {
    id: 'vendor-audits',
    name: 'Vendor Audits',
    icon: Users,
    items: [
      { id: 'vendor-activity', name: 'Vendor Activity Summary', description: 'Review vendor transactions' },
      { id: 'vendor-balances', name: 'Vendor Balance Verification', description: 'Verify AP balances' },
      { id: 'duplicate-vendors', name: 'Duplicate Vendor Detection', description: 'Find duplicate vendor records' },
      { id: 'inactive-vendors', name: 'Inactive Vendor Report', description: 'Identify inactive vendors' }
    ]
  },
  {
    id: 'customer-audits',
    name: 'Customer Audits',
    icon: Building2,
    items: [
      { id: 'customer-activity', name: 'Customer Activity Summary', description: 'Review customer transactions' },
      { id: 'ar-aging', name: 'AR Aging Verification', description: 'Verify accounts receivable aging' },
      { id: 'duplicate-customers', name: 'Duplicate Customer Detection', description: 'Find duplicate customer records' },
      { id: 'credit-limit', name: 'Credit Limit Review', description: 'Review customer credit limits' }
    ]
  },
  {
    id: 'inventory-audits',
    name: 'Inventory Audits',
    icon: Package,
    items: [
      { id: 'inventory-valuation', name: 'Inventory Valuation', description: 'Verify inventory values' },
      { id: 'negative-inventory', name: 'Negative Inventory Check', description: 'Find negative quantities' },
      { id: 'slow-moving', name: 'Slow Moving Items', description: 'Identify slow-moving inventory' },
      { id: 'cost-variance', name: 'Cost Variance Analysis', description: 'Analyze cost variations' }
    ]
  },
  {
    id: 'system-audits',
    name: 'System Audits',
    icon: Activity,
    items: [
      { id: 'data-integrity', name: 'Data Integrity Check', description: 'Verify database integrity' },
      { id: 'user-activity', name: 'User Activity Audit', description: 'Review user actions' },
      { id: 'permission-audit', name: 'Permission Audit', description: 'Review user permissions' },
      { id: 'backup-verification', name: 'Backup Verification', description: 'Verify backup integrity' }
    ]
  }
]

export default function AuditTools({ currentUser, companyName }: AuditToolsProps) {
  const [loading, setLoading] = useState(false)
  const [selectedAudit, setSelectedAudit] = useState<{ category: string; item: string } | null>(null)
  const [results, setResults] = useState<any>(null)
  const [error, setError] = useState<string>('')
  const [accountFilter, setAccountFilter] = useState<string>('')
  const [yearFilter, setYearFilter] = useState<string>('')
  const [vendorFilter, setVendorFilter] = useState<string>('')
  const [customerFilter, setCustomerFilter] = useState<string>('')
  const [selectedRowData, setSelectedRowData] = useState<any>(null)
  const [voidIssueFilter, setVoidIssueFilter] = useState<string>('all')
  const [startDate, setStartDate] = useState<string>('')
  const [endDate, setEndDate] = useState<string>('')
  const [showFollowBatch, setShowFollowBatch] = useState<boolean>(false)

  // Get current audit details
  const getCurrentAudit = () => {
    if (!selectedAudit) return null
    const category = auditCategories.find(c => c.id === selectedAudit.category)
    if (!category) return null
    const item = category.items.find(i => i.id === selectedAudit.item)
    return { category, item }
  }

  const runAudit = async () => {
    if (!selectedAudit) return

    setLoading(true)
    setError('')
    setResults(null)

    try {
      // Run the appropriate audit based on selection
      switch (selectedAudit.item) {
        case 'gl-validation':
          // @ts-ignore
          const validationResults = await window.go.main.App.ValidateGLBalances(companyName, accountFilter || '')
          setResults(validationResults)
          break
        
        case 'gl-activity':
          // @ts-ignore
          const activityResults = await window.go.main.App.CheckGLPeriodFields(companyName)
          setResults(activityResults)
          break
        
        case 'year-analysis':
          // @ts-ignore
          const yearResults = await window.go.main.App.AnalyzeGLBalancesByYear(companyName, accountFilter || '102000')
          setResults(yearResults)
          break
        
        case 'check-batch':
          // @ts-ignore
          const checkResults = await window.go.main.App.AuditCheckBatches(companyName)
          setResults(checkResults)
          break
        
        case 'duplicate-cidchec':
          console.log('Running duplicate CIDCHEC audit for company:', companyName)
          // @ts-ignore
          const duplicateResults = await window.go.main.App.AuditDuplicateCIDCHEC(companyName)
          console.log('Duplicate CIDCHEC audit results:', duplicateResults)
          setResults(duplicateResults)
          break
        
        case 'void-verification':
          console.log('Running void verification audit for company:', companyName)
          // @ts-ignore
          const voidResults = await window.go.main.App.AuditVoidChecks(companyName)
          console.log('Void verification audit results:', voidResults)
          setResults(voidResults)
          break
        
        case 'check-gl-matching':
          console.log('Running check-GL matching audit for company:', companyName)
          // @ts-ignore
          const glMatchResults = await window.go.main.App.AuditCheckGLMatching(
            companyName, 
            accountFilter || '', 
            startDate || '', 
            endDate || ''
          )
          console.log('Check-GL matching audit results:', glMatchResults)
          setResults(glMatchResults)
          break
        
        case 'payee-cid-verification':
          console.log('Running payee-CID verification audit for company:', companyName)
          // @ts-ignore
          const payeeResults = await window.go.main.App.AuditPayeeCIDVerification(companyName)
          console.log('Payee-CID verification audit results:', payeeResults)
          setResults(payeeResults)
          break
        
        case 'follow-batch-number':
          // This opens the Follow Batch Number component
          setShowFollowBatch(true)
          setResults(null)
          break
        
        default:
          setError('This audit type is not yet implemented')
          break
      }
    } catch (err: any) {
      setError(err.message || 'Failed to run audit')
    } finally {
      setLoading(false)
    }
  }

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(value)
  }

  const formatPercentage = (value: string) => {
    const num = parseFloat(value)
    if (num > 50) return <span className="text-red-600 font-semibold">{value}</span>
    if (num > 10) return <span className="text-amber-600 font-semibold">{value}</span>
    return <span className="text-green-600">{value}</span>
  }

  const renderResults = () => {
    if (!results || !selectedAudit) return null

    // Render different result formats based on audit type
    switch (selectedAudit.item) {
      case 'gl-validation':
        return renderGLValidationResults()
      case 'gl-activity':
        return renderGLActivityResults()
      case 'year-analysis':
        return renderYearAnalysisResults()
      case 'check-batch':
        return renderCheckBatchResults()
      case 'duplicate-cidchec':
        return renderDuplicateCIDCHECResults()
      case 'void-verification':
        return renderVoidVerificationResults()
      case 'check-gl-matching':
        return renderCheckGLMatchingResults()
      case 'payee-cid-verification':
        return renderPayeeCIDVerificationResults()
      default:
        return <pre className="text-xs">{JSON.stringify(results, null, 2)}</pre>
    }
  }

  const renderGLValidationResults = () => {
    if (!results) return null
    
    return (
      <div className="space-y-4">
        {/* Overall Balance Check */}
        <Alert className={results.is_balanced ? "" : "border-red-500 bg-red-50"}>
          <div className="flex items-center gap-2">
            {results.is_balanced ? (
              <CheckCircle className="h-4 w-4 text-green-600" />
            ) : (
              <AlertTriangle className="h-4 w-4 text-red-600" />
            )}
            <AlertTitle>
              {results.is_balanced ? "GL is Balanced" : "GL Out of Balance!"}
            </AlertTitle>
          </div>
          <AlertDescription>
            <div className="mt-2 space-y-1">
              <p>Total Debits: {formatCurrency(results.total_debits)}</p>
              <p>Total Credits: {formatCurrency(results.total_credits)}</p>
              <p className={results.is_balanced ? "text-green-600" : "text-red-600 font-bold"}>
                Difference: {formatCurrency(results.overall_difference)}
              </p>
              <p className="text-sm text-gray-500 mt-2">Total Records: {results.total_rows_checked?.toLocaleString()}</p>
            </div>
          </AlertDescription>
        </Alert>

        {/* Issues Found */}
        <div className="grid gap-4 md:grid-cols-3">
          {results.duplicate_count > 0 && (
            <Card className="border-amber-500">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-amber-600">Duplicate Transactions</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{results.duplicate_count}</div>
                <p className="text-xs text-muted-foreground">Potential duplicates found</p>
              </CardContent>
            </Card>
          )}

          {results.zero_amount_transactions > 0 && (
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Zero Amount Entries</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{results.zero_amount_transactions}</div>
                <p className="text-xs text-muted-foreground">Transactions with no amounts</p>
              </CardContent>
            </Card>
          )}

          {results.suspicious_count > 0 && (
            <Card className="border-red-500">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-red-600">Suspicious Amounts</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{results.suspicious_count}</div>
                <p className="text-xs text-muted-foreground">Unusually large transactions</p>
              </CardContent>
            </Card>
          )}
        </div>

        {/* Imbalanced Accounts */}
        {results.imbalanced_accounts?.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="text-sm font-medium">Top Imbalanced Accounts</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {results.imbalanced_accounts.slice(0, 5).map((account: any, idx: number) => (
                  <div key={idx} className="flex justify-between text-sm">
                    <span>{account.account}</span>
                    <span className="text-red-600 font-medium">
                      {formatCurrency(account.difference)}
                    </span>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    )
  }

  const renderGLActivityResults = () => {
    if (!results) return null

    return (
      <div className="space-y-4">
        <div className="grid gap-4 md:grid-cols-3">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Total GL Records</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{results.total_rows?.toLocaleString()}</div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Blank Year Fields</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {results.blank_year_count?.toLocaleString()}
              </div>
              <p className="text-xs text-muted-foreground">
                {formatPercentage(results.blank_year_pct || '0%')}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Blank Period Fields</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {results.blank_period_count?.toLocaleString()}
              </div>
              <p className="text-xs text-muted-foreground">
                {formatPercentage(results.blank_period_pct || '0%')}
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Years and Periods */}
        <div className="grid gap-4 md:grid-cols-2">
          {results.unique_years && Object.keys(results.unique_years).length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-sm font-medium">Years Found</CardTitle>
              </CardHeader>
              <CardContent>
                <ScrollArea className="h-40">
                  <div className="space-y-1">
                    {Object.entries(results.unique_years)
                      .sort(([a], [b]) => b.localeCompare(a))
                      .map(([year, count]) => (
                        <div key={year} className="flex justify-between text-sm">
                          <span className="font-mono">{year}</span>
                          <span className="text-muted-foreground">{count} records</span>
                        </div>
                      ))}
                  </div>
                </ScrollArea>
              </CardContent>
            </Card>
          )}

          {results.unique_periods && Object.keys(results.unique_periods).length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-sm font-medium">Periods Found</CardTitle>
              </CardHeader>
              <CardContent>
                <ScrollArea className="h-40">
                  <div className="space-y-1">
                    {Object.entries(results.unique_periods)
                      .sort(([a], [b]) => a.localeCompare(b))
                      .map(([period, count]) => (
                        <div key={period} className="flex justify-between text-sm">
                          <span className="font-mono">{period || 'BLANK'}</span>
                          <span className="text-muted-foreground">{count} records</span>
                        </div>
                      ))}
                  </div>
                </ScrollArea>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    )
  }

  const renderYearAnalysisResults = () => {
    if (!results) return null

    return (
      <div className="space-y-4">
        <Alert>
          <AlertTitle>Analysis for Account {accountFilter || '102000'}</AlertTitle>
          <AlertDescription>
            Found data for {results.years_found} years
          </AlertDescription>
        </Alert>

        {results.yearly_totals && (
          <Card>
            <CardHeader>
              <CardTitle className="text-sm font-medium">Yearly Balance Progression</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {Object.entries(results.yearly_totals)
                  .sort(([a], [b]) => b.localeCompare(a))
                  .map(([year, data]: [string, any]) => (
                    <div key={year} className="flex justify-between items-center p-2 hover:bg-gray-50 rounded">
                      <span className="font-mono text-sm">{year}</span>
                      <div className="text-right">
                        <div className="text-sm font-medium">{formatCurrency(data.balance)}</div>
                        <div className="text-xs text-gray-500">
                          D: {formatCurrency(data.debits)} | C: {formatCurrency(data.credits)}
                        </div>
                      </div>
                    </div>
                  ))}
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    )
  }

  const renderCheckBatchResults = () => {
    if (!results) return null

    return (
      <div className="space-y-4">
        <Alert className={results.missing_entries === 0 ? "" : "border-amber-500 bg-amber-50"}>
          <AlertTitle>Check Batch Audit Results</AlertTitle>
          <AlertDescription>
            <div className="mt-2 space-y-1">
              <p>Total Checks: {results.total_checks}</p>
              <p>Missing GL Entries: {results.missing_entries}</p>
              <p>Mismatched Amounts: {results.mismatched_amounts}</p>
            </div>
          </AlertDescription>
        </Alert>

        {results.issues && results.issues.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="text-sm font-medium">Issues Found</CardTitle>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-60">
                <div className="space-y-2">
                  {results.issues.slice(0, 50).map((issue: any, idx: number) => (
                    <div key={idx} className="text-sm p-2 bg-amber-50 rounded">
                      <div className="font-medium">Check #{issue.check_number}</div>
                      <div className="text-xs text-gray-600">
                        {issue.issue_type}: {issue.details}
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        )}
      </div>
    )
  }

  const renderDuplicateCIDCHECResults = () => {
    if (!results) return null

    const severity = results.severity || 'low'
    const severityColors = {
      high: 'border-red-500 bg-red-50',
      medium: 'border-amber-500 bg-amber-50',
      low: 'border-green-500 bg-green-50'
    }

    const severityIcons = {
      high: <AlertTriangle className="h-4 w-4 text-red-600" />,
      medium: <AlertCircle className="h-4 w-4 text-amber-600" />,
      low: <CheckCircle className="h-4 w-4 text-green-600" />
    }

    return (
      <div className="space-y-4">
        {/* Summary Alert */}
        <Alert className={severityColors[severity as keyof typeof severityColors]}>
          <div className="flex items-center gap-2">
            {severityIcons[severity as keyof typeof severityIcons]}
            <AlertTitle>CIDCHEC Duplicate Detection Results</AlertTitle>
          </div>
          <AlertDescription>
            <div className="mt-2">
              <p className="font-medium">{results.message}</p>
              <div className="mt-2 space-y-1 text-sm">
                <p>Total Checks Analyzed: {results.summary?.total_checks?.toLocaleString()}</p>
                <p>Unique CIDCHEC Values: {results.summary?.unique_cidchec_values?.toLocaleString()}</p>
                <p>Empty/Null CIDCHECs: {results.summary?.empty_or_null_cidchec?.toLocaleString()}</p>
              </div>
            </div>
          </AlertDescription>
        </Alert>

        {/* Summary Cards */}
        {results.summary?.duplicate_groups_found > 0 && (
          <div className="grid gap-4 md:grid-cols-3">
            <Card className="border-red-500">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-red-600">Duplicate Groups</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{results.summary.duplicate_groups_found}</div>
                <p className="text-xs text-muted-foreground">Unique CIDCHEC values with duplicates</p>
              </CardContent>
            </Card>

            <Card className="border-red-500">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-red-600">Affected Checks</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{results.summary.total_duplicate_checks}</div>
                <p className="text-xs text-muted-foreground">Total checks with duplicate CIDCHECs</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Empty CIDCHECs</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{results.summary.empty_or_null_cidchec}</div>
                <p className="text-xs text-muted-foreground">Checks without CIDCHEC values</p>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Duplicate Details */}
        {results.duplicates && results.duplicates.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="text-sm font-medium">Duplicate CIDCHEC Details</CardTitle>
              <CardDescription>
                Showing checks that share the same CIDCHEC value (should be unique)
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-96">
                <div className="space-y-4">
                  {results.duplicates.map((group: any, groupIdx: number) => (
                    <div key={groupIdx} className="border border-red-200 rounded-lg p-4 bg-red-50">
                      <div className="flex items-center justify-between mb-2">
                        <div>
                          <span className="font-semibold text-sm">CIDCHEC: {group.cidchec}</span>
                          <Badge variant="destructive" className="ml-2">
                            {group.occurrence_count} duplicates
                          </Badge>
                        </div>
                        <span className="text-sm font-medium">
                          Total: {formatCurrency(group.total_amount)}
                        </span>
                      </div>
                      <div className="space-y-2 mt-3">
                        {group.checks.map((check: any, checkIdx: number) => (
                          <div key={checkIdx} className="flex items-center justify-between text-sm p-2 bg-white rounded border border-red-100">
                            <div className="flex-1 grid grid-cols-5 gap-2">
                              <div>
                                <span className="text-gray-500">Check #:</span>
                                <span className="ml-1 font-medium">{check.check_number || 'N/A'}</span>
                              </div>
                              <div>
                                <span className="text-gray-500">Date:</span>
                                <span className="ml-1">{check.date || 'N/A'}</span>
                              </div>
                              <div>
                                <span className="text-gray-500">Amount:</span>
                                <span className="ml-1 font-medium">{formatCurrency(check.amount || 0)}</span>
                              </div>
                              <div>
                                <span className="text-gray-500">Payee:</span>
                                <span className="ml-1">{check.payee || 'N/A'}</span>
                              </div>
                              <div>
                                <span className="text-gray-500">Row:</span>
                                <span className="ml-1 text-xs">{check.row_index}</span>
                              </div>
                            </div>
                            <div className="flex gap-2 ml-2">
                              {check.cleared && (
                                <Badge variant="secondary" className="text-xs">Cleared</Badge>
                              )}
                              {check.voided && (
                                <Badge variant="destructive" className="text-xs">Voided</Badge>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        )}

        {/* No Issues Found */}
        {(!results.duplicates || results.duplicates.length === 0) && results.severity === 'low' && (
          <Card className="border-green-500">
            <CardHeader>
              <CardTitle className="text-sm font-medium text-green-600 flex items-center gap-2">
                <CheckCircle className="h-4 w-4" />
                No Duplicates Found
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-gray-600">
                All CIDCHEC values in the checks database are unique. Data integrity is maintained.
              </p>
            </CardContent>
          </Card>
        )}
      </div>
    )
  }

  const renderVoidVerificationResults = () => {
    if (!results) return null

    const severity = results.severity || 'low'
    const severityColors = {
      high: 'border-red-500 bg-red-50',
      medium: 'border-amber-500 bg-amber-50',
      low: 'border-green-500 bg-green-50',
      info: 'border-blue-500 bg-blue-50'
    }

    const severityIcons = {
      high: <AlertTriangle className="h-4 w-4 text-red-600" />,
      medium: <AlertCircle className="h-4 w-4 text-amber-600" />,
      low: <CheckCircle className="h-4 w-4 text-green-600" />,
      info: <AlertCircle className="h-4 w-4 text-blue-600" />
    }

    // Calculate counts for each filter type (independent of current filter)
    const calculateFilterCounts = (issues: any[]) => {
      if (!issues) return {
        all: 0,
        amountMismatchNonZero: 0,
        amountMismatchZero: 0,
        notCleared: 0,
        noRecordDate: 0,
        missingVoidAmount: 0
      }

      return {
        all: issues.length,
        amountMismatchNonZero: issues.filter((i: any) => 
          i.issues?.some((x: string) => x.includes('Amount mismatch')) && i.void_amount !== 0
        ).length,
        amountMismatchZero: issues.filter((i: any) => 
          i.issues?.some((x: string) => x.includes('Amount mismatch')) && i.void_amount === 0
        ).length,
        notCleared: issues.filter((i: any) => 
          i.issues?.some((x: string) => x.includes('Not marked as cleared'))
        ).length,
        noRecordDate: issues.filter((i: any) => 
          i.issues?.some((x: string) => x.includes('Record date is null'))
        ).length,
        missingVoidAmount: issues.filter((i: any) => 
          i.issues?.some((x: string) => x.includes('NVOIDAMT column not found'))
        ).length
      }
    }

    const filterCounts = calculateFilterCounts(results.issues || [])

    // Filter issues based on selected filter
    const filterIssues = (issues: any[]) => {
      if (!issues || voidIssueFilter === 'all') return issues
      
      return issues.filter((issue: any) => {
        const hasAmountMismatch = issue.issues?.some((i: string) => i.includes('Amount mismatch'))
        const voidAmount = issue.void_amount || 0
        const amount = issue.amount || 0
        
        switch (voidIssueFilter) {
          case 'amount-mismatch-nonzero':
            // Amount mismatch where void amount is not 0.00
            return hasAmountMismatch && voidAmount !== 0
          case 'amount-mismatch-zero':
            // Amount mismatch where void amount is 0.00
            return hasAmountMismatch && voidAmount === 0
          case 'not-cleared':
            // Not marked as cleared
            return issue.issues?.some((i: string) => i.includes('Not marked as cleared'))
          case 'no-record-date':
            // Missing record date
            return issue.issues?.some((i: string) => i.includes('Record date is null'))
          case 'missing-void-amount':
            // NVOIDAMT column not found or void amount is missing
            return issue.issues?.some((i: string) => i.includes('NVOIDAMT column not found'))
          default:
            return true
        }
      })
    }

    const filteredIssues = filterIssues(results.issues || [])

    return (
      <div className="space-y-4">
        {/* Summary Alert */}
        <Alert className={severityColors[severity as keyof typeof severityColors]}>
          <div className="flex items-center gap-2">
            {severityIcons[severity as keyof typeof severityIcons]}
            <AlertTitle>Void Verification Audit Results</AlertTitle>
          </div>
          <AlertDescription>
            <div className="mt-2">
              <p className="font-medium">{results.message}</p>
              <div className="mt-2 space-y-1 text-sm">
                <p>Total Voided Checks: {results.summary?.total_voided_checks?.toLocaleString()}</p>
                <p>Issues Found: {results.summary?.total_issues_found?.toLocaleString()}</p>
                {results.summary?.issue_percentage > 0 && (
                  <p>Issue Rate: {results.summary.issue_percentage.toFixed(1)}%</p>
                )}
              </div>
            </div>
          </AlertDescription>
        </Alert>

        {/* Summary Cards */}
        {results.summary?.total_voided_checks > 0 && (
          <div className="grid gap-4 md:grid-cols-3">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Voided Checks</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{results.summary.total_voided_checks}</div>
                <p className="text-xs text-muted-foreground">Total voided checks found</p>
              </CardContent>
            </Card>

            <Card className={results.summary.total_issues_found > 0 ? "border-red-500" : ""}>
              <CardHeader className="pb-2">
                <CardTitle className={`text-sm font-medium ${results.summary.total_issues_found > 0 ? "text-red-600" : ""}`}>
                  Issues Found
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{results.summary.total_issues_found}</div>
                <p className="text-xs text-muted-foreground">Checks with improper settings</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Compliance Rate</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {(100 - (results.summary.issue_percentage || 0)).toFixed(1)}%
                </div>
                <p className="text-xs text-muted-foreground">Properly configured voids</p>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Issues List */}
        {results.issues && results.issues.length > 0 && (
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-sm font-medium">Voided Checks with Issues</CardTitle>
                  <CardDescription>
                    Click on any row to view complete check details
                  </CardDescription>
                </div>
                <select 
                  value={voidIssueFilter} 
                  onChange={(e) => setVoidIssueFilter(e.target.value)}
                  className="w-64 flex h-10 items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  <option value="all">
                    All Issues ({filterCounts.all})
                  </option>
                  <option value="amount-mismatch-nonzero">
                    Amount Mismatch (void amt ≠ 0.00) ({filterCounts.amountMismatchNonZero})
                  </option>
                  <option value="amount-mismatch-zero">
                    Amount Mismatch (void amt = 0.00) ({filterCounts.amountMismatchZero})
                  </option>
                  <option value="not-cleared">
                    Not Marked as Cleared ({filterCounts.notCleared})
                  </option>
                  <option value="no-record-date">
                    Missing Record Date ({filterCounts.noRecordDate})
                  </option>
                  <option value="missing-void-amount">
                    Missing Void Amount Column ({filterCounts.missingVoidAmount})
                  </option>
                </select>
              </div>
            </CardHeader>
            <CardContent>
              {filteredIssues.length > 0 ? (
                <>
                  <div className="text-sm text-gray-500 mb-2">
                    Showing {filteredIssues.length} of {results.issues.length} issues
                  </div>
                  <ScrollArea className="h-96">
                    <div className="space-y-2">
                      {filteredIssues.map((issue: any, idx: number) => (
                    <div 
                      key={idx} 
                      className="border border-red-200 rounded-lg p-3 bg-white hover:bg-red-50 cursor-pointer transition-colors"
                      onClick={() => setSelectedRowData(issue.row_data)}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center gap-4 mb-2">
                            <span className="font-semibold text-sm">
                              Check #{issue.check_number || 'N/A'}
                            </span>
                            {issue.check_date && (
                              <span className="text-sm text-gray-500">{issue.check_date}</span>
                            )}
                            <Badge variant="destructive" className="text-xs">
                              {issue.issue_count} issue{issue.issue_count > 1 ? 's' : ''}
                            </Badge>
                          </div>
                          
                          <div className="grid grid-cols-2 gap-2 text-sm mb-2">
                            <div>
                              <span className="text-gray-500">Payee:</span>
                              <span className="ml-2">{issue.payee || 'N/A'}</span>
                            </div>
                            <div>
                              <span className="text-gray-500">Account:</span>
                              <span className="ml-2">{issue.account || 'N/A'}</span>
                            </div>
                            <div>
                              <span className="text-gray-500">Amount:</span>
                              <span className="ml-2 font-medium">{formatCurrency(issue.amount || 0)}</span>
                            </div>
                            <div>
                              <span className="text-gray-500">Void Amount:</span>
                              <span className="ml-2 font-medium">{formatCurrency(issue.void_amount || 0)}</span>
                            </div>
                          </div>

                          <div className="space-y-1">
                            {issue.issues.map((issueDetail: string, detailIdx: number) => (
                              <div key={detailIdx} className="flex items-start gap-2">
                                <XCircle className="h-3 w-3 text-red-500 mt-0.5 flex-shrink-0" />
                                <span className="text-xs text-red-700">{issueDetail}</span>
                              </div>
                            ))}
                          </div>

                          <div className="flex gap-4 mt-2 text-xs text-gray-500">
                            <span>Row: {issue.row_index}</span>
                            {issue.cidchec && <span>CIDCHEC: {issue.cidchec}</span>}
                            {issue.is_cleared !== undefined && (
                              <Badge variant={issue.is_cleared ? "secondary" : "outline"} className="text-xs">
                                {issue.is_cleared ? 'Cleared' : 'Not Cleared'}
                              </Badge>
                            )}
                            {issue.record_date && (
                              <span>Rec Date: {issue.record_date}</span>
                            )}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                    </div>
                  </ScrollArea>
                </>
              ) : (
                <div className="text-center py-8 text-gray-500">
                  <AlertCircle className="h-12 w-12 mx-auto mb-2 text-gray-400" />
                  <p className="text-sm">No issues match the selected filter</p>
                  <p className="text-xs mt-1">Try selecting a different filter option</p>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* No Issues Found */}
        {results.summary?.total_voided_checks > 0 && (!results.issues || results.issues.length === 0) && (
          <Card className="border-green-500">
            <CardHeader>
              <CardTitle className="text-sm font-medium text-green-600 flex items-center gap-2">
                <CheckCircle className="h-4 w-4" />
                All Voided Checks Properly Configured
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-gray-600">
                All {results.summary.total_voided_checks} voided checks have the correct settings:
                amount matches void amount, marked as cleared, and have record dates.
              </p>
            </CardContent>
          </Card>
        )}

        {/* Modal for Row Data */}
        <Dialog open={!!selectedRowData} onOpenChange={() => setSelectedRowData(null)}>
          <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Complete Check Record</DialogTitle>
              <DialogDescription>
                All fields from the checks.dbf record
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-2 mt-4">
              {selectedRowData && Object.entries(selectedRowData).map(([key, value]) => (
                <div key={key} className="grid grid-cols-3 gap-2 py-1 border-b">
                  <span className="font-medium text-sm">{key}:</span>
                  <span className="col-span-2 text-sm">
                    {value === null || value === undefined ? 
                      '<null>' : 
                      typeof value === 'boolean' ? 
                        (value ? 'True' : 'False') : 
                        String(value)}
                  </span>
                </div>
              ))}
            </div>
          </DialogContent>
        </Dialog>
      </div>
    )
  }

  const renderCheckGLMatchingResults = () => {
    if (!results) return null

    const severity = results.severity || 'low'
    const severityColors = {
      high: 'border-red-500 bg-red-50',
      medium: 'border-amber-500 bg-amber-50',
      low: 'border-green-500 bg-green-50'
    }

    const severityIcons = {
      high: <AlertTriangle className="h-4 w-4 text-red-600" />,
      medium: <AlertCircle className="h-4 w-4 text-amber-600" />,
      low: <CheckCircle className="h-4 w-4 text-green-600" />
    }

    return (
      <div className="space-y-4">
        {/* Summary Alert */}
        <Alert className={severityColors[severity as keyof typeof severityColors]}>
          <div className="flex items-center gap-2">
            {severityIcons[severity as keyof typeof severityIcons]}
            <AlertTitle>Check-GL Matching Audit Results</AlertTitle>
          </div>
          <AlertDescription>
            <div className="mt-2">
              <p className="font-medium">{results.message}</p>
              <div className="mt-2 space-y-1 text-sm">
                {results.account && <p>Account: {results.account || 'All Accounts'}</p>}
                {results.date_range && (
                  <p>Date Range: {results.date_range.start} to {results.date_range.end}</p>
                )}
              </div>
            </div>
          </AlertDescription>
        </Alert>

        {/* Summary Cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Total Checks</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{results.summary?.total_checks || 0}</div>
              <p className="text-xs text-muted-foreground">In date range</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Total GL Entries</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{results.summary?.total_gl_entries || 0}</div>
              <p className="text-xs text-muted-foreground">In date range</p>
            </CardContent>
          </Card>

          <Card className={results.summary?.unmatched_checks > 0 ? "border-red-500" : ""}>
            <CardHeader className="pb-2">
              <CardTitle className={`text-sm font-medium ${results.summary?.unmatched_checks > 0 ? "text-red-600" : ""}`}>
                Unmatched Checks
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{results.summary?.unmatched_checks || 0}</div>
              <p className="text-xs text-muted-foreground">No GL entry found</p>
            </CardContent>
          </Card>

          <Card className={results.summary?.unmatched_gl > 0 ? "border-amber-500" : ""}>
            <CardHeader className="pb-2">
              <CardTitle className={`text-sm font-medium ${results.summary?.unmatched_gl > 0 ? "text-amber-600" : ""}`}>
                Unmatched GL
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{results.summary?.unmatched_gl || 0}</div>
              <p className="text-xs text-muted-foreground">No check found</p>
            </CardContent>
          </Card>
        </div>

        {/* Unmatched Checks */}
        {results.unmatched_checks && results.unmatched_checks.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="text-sm font-medium">Checks Without GL Entries</CardTitle>
              <CardDescription>
                These checks exist in CHECKS.DBF but have no matching GL entry
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-64">
                <div className="space-y-2">
                  {results.unmatched_checks.map((check: any, idx: number) => (
                    <div 
                      key={idx} 
                      className="border border-red-200 rounded-lg p-3 bg-white hover:bg-red-50 cursor-pointer transition-colors"
                      onClick={() => setSelectedRowData(check.row_data)}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex-1">
                          <div className="flex items-center gap-4 mb-1">
                            <span className="font-semibold text-sm">
                              Check #{check.check_num || 'N/A'}
                            </span>
                            <Badge variant={check.entry_type === 'C' ? 'default' : 'secondary'}>
                              {check.entry_type === 'C' ? 'Credit' : 'Debit'}
                            </Badge>
                            <span className="text-sm text-gray-500">{check.date}</span>
                          </div>
                          <div className="grid grid-cols-3 gap-2 text-sm">
                            <div>
                              <span className="text-gray-500">Payee:</span>
                              <span className="ml-2">{check.payee || 'N/A'}</span>
                            </div>
                            <div>
                              <span className="text-gray-500">Amount:</span>
                              <span className="ml-2 font-medium">{formatCurrency(check.amount || 0)}</span>
                            </div>
                            <div>
                              <span className="text-gray-500">CID:</span>
                              <span className="ml-2 font-mono text-xs">{check.cid || 'N/A'}</span>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        )}

        {/* Unmatched GL Entries */}
        {results.unmatched_gl && results.unmatched_gl.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="text-sm font-medium">GL Entries Without Checks</CardTitle>
              <CardDescription>
                These GL entries exist in GLMASTER.DBF but have no matching check
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-64">
                <div className="space-y-2">
                  {results.unmatched_gl.map((gl: any, idx: number) => (
                    <div 
                      key={idx} 
                      className="border border-amber-200 rounded-lg p-3 bg-white hover:bg-amber-50 cursor-pointer transition-colors"
                      onClick={() => setSelectedRowData(gl.row_data)}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex-1">
                          <div className="flex items-center gap-4 mb-1">
                            <span className="font-semibold text-sm">
                              GL Entry
                            </span>
                            <span className="text-sm text-gray-500">{gl.date}</span>
                          </div>
                          <div className="grid grid-cols-3 gap-2 text-sm">
                            <div>
                              <span className="text-gray-500">Description:</span>
                              <span className="ml-2">{gl.description || 'N/A'}</span>
                            </div>
                            <div>
                              <span className="text-gray-500">Credits:</span>
                              <span className="ml-2 font-medium text-green-600">
                                {gl.credits > 0 ? formatCurrency(gl.credits) : '-'}
                              </span>
                            </div>
                            <div>
                              <span className="text-gray-500">Debits:</span>
                              <span className="ml-2 font-medium text-red-600">
                                {gl.debits > 0 ? formatCurrency(gl.debits) : '-'}
                              </span>
                            </div>
                          </div>
                          <div className="mt-1 text-xs text-gray-500">
                            CID: {gl.cid || 'N/A'} | Account: {gl.account}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        )}

        {/* Perfect Match */}
        {results.summary?.unmatched_checks === 0 && results.summary?.unmatched_gl === 0 && (
          <Card className="border-green-500">
            <CardHeader>
              <CardTitle className="text-sm font-medium text-green-600 flex items-center gap-2">
                <CheckCircle className="h-4 w-4" />
                Perfect Match
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-gray-600">
                All {results.summary?.total_checks} checks have matching GL entries and 
                all {results.summary?.total_gl_entries} GL entries have matching checks.
              </p>
            </CardContent>
          </Card>
        )}

        {/* Modal for Row Data */}
        <Dialog open={!!selectedRowData} onOpenChange={() => setSelectedRowData(null)}>
          <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Complete Record Details</DialogTitle>
              <DialogDescription>
                All fields from the database record
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-2 mt-4">
              {selectedRowData && Object.entries(selectedRowData).map(([key, value]) => (
                <div key={key} className="grid grid-cols-3 gap-2 py-1 border-b">
                  <span className="font-medium text-sm">{key}:</span>
                  <span className="col-span-2 text-sm">
                    {value === null || value === undefined ? 
                      '<null>' : 
                      typeof value === 'boolean' ? 
                        (value ? 'True' : 'False') : 
                        String(value)}
                  </span>
                </div>
              ))}
            </div>
          </DialogContent>
        </Dialog>
      </div>
    )
  }

  const renderPayeeCIDVerificationResults = () => {
    if (!results) return null

    const severity = results.severity || 'success'
    const severityColors = {
      error: 'border-red-500 bg-red-50',
      warning: 'border-amber-500 bg-amber-50',
      success: 'border-green-500 bg-green-50'
    }

    const severityIcons = {
      error: <AlertTriangle className="h-4 w-4 text-red-600" />,
      warning: <AlertCircle className="h-4 w-4 text-amber-600" />,
      success: <CheckCircle className="h-4 w-4 text-green-600" />
    }

    return (
      <div className="space-y-4">
        {/* Summary Alert */}
        <Alert className={severityColors[severity as keyof typeof severityColors]}>
          <div className="flex items-center gap-2">
            {severityIcons[severity as keyof typeof severityIcons]}
            <AlertTitle>Payee-CID Verification Results</AlertTitle>
          </div>
          <AlertDescription>
            <div className="mt-2">
              <p className="font-medium">{results.message}</p>
              <div className="mt-2 space-y-1 text-sm">
                <p>Checks Processed: {results.checks_processed}</p>
                <p>Total Investors: {results.total_investors}</p>
                <p>Total Vendors: {results.total_vendors}</p>
              </div>
            </div>
          </AlertDescription>
        </Alert>

        {/* Summary Cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Checks Processed</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{results.checks_processed || 0}</div>
              <p className="text-xs text-muted-foreground">Total checks analyzed</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Investors</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{results.total_investors || 0}</div>
              <p className="text-xs text-muted-foreground">In INVESTOR.dbf</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Vendors</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{results.total_vendors || 0}</div>
              <p className="text-xs text-muted-foreground">In VENDOR.dbf</p>
            </CardContent>
          </Card>

          <Card className={results.mismatches_found > 0 ? "border-red-500" : ""}>
            <CardHeader className="pb-2">
              <CardTitle className={`text-sm font-medium ${results.mismatches_found > 0 ? "text-red-600" : ""}`}>
                Mismatches Found
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{results.mismatches_found || 0}</div>
              <p className="text-xs text-muted-foreground">Payee/CID issues</p>
            </CardContent>
          </Card>
        </div>

        {/* Mismatches List */}
        {results.mismatches && results.mismatches.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="text-sm font-medium">Payee/CID Mismatches</CardTitle>
              <CardDescription>
                Checks where the payee doesn't match the CID record in investor/vendor tables
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ScrollArea className="h-96">
                <div className="space-y-2">
                  {results.mismatches.map((mismatch: any, idx: number) => (
                    <div 
                      key={idx} 
                      className="border border-red-200 rounded-lg p-3 bg-white hover:bg-red-50 cursor-pointer transition-colors"
                      onClick={() => setSelectedRowData(mismatch.full_row)}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex-1">
                          <div className="flex items-center gap-4 mb-1">
                            <span className="font-semibold text-sm">
                              Check #{mismatch.check_number || 'N/A'}
                            </span>
                            <span className="text-sm text-gray-500">{mismatch.check_date}</span>
                            <span className="font-medium text-sm">
                              {formatCurrency(mismatch.amount)}
                            </span>
                          </div>
                          <div className="grid grid-cols-2 gap-2 text-sm mb-2">
                            <div>
                              <span className="text-gray-500">Payee:</span>
                              <span className="ml-2 font-medium">{mismatch.payee}</span>
                            </div>
                            <div>
                              <span className="text-gray-500">CID:</span>
                              <span className="ml-2 font-mono text-xs">{mismatch.cid}</span>
                            </div>
                          </div>
                          <div className="flex items-center gap-2 mb-2">
                            {mismatch.found_in_vendor && (
                              <Badge variant="outline" className="text-purple-600">
                                Found in Vendor Table
                              </Badge>
                            )}
                            {mismatch.found_in_investor && (
                              <Badge variant="outline" className="text-blue-600">
                                Found in Investor Table
                              </Badge>
                            )}
                            {!mismatch.found_in_vendor && !mismatch.found_in_investor && (
                              <Badge variant="outline" className="text-red-600">
                                Payee Not Found
                              </Badge>
                            )}
                            {mismatch.matched_table && (
                              <Badge variant="outline" className="text-green-600">
                                CID Match in {mismatch.matched_table}
                              </Badge>
                            )}
                          </div>
                          {mismatch.possible_cids && mismatch.possible_cids.length > 0 && (
                            <div className="text-sm">
                              <span className="text-gray-500">Expected CID{mismatch.possible_cids.length > 1 ? 's' : ''}:</span>
                              <span className="ml-2 font-medium text-green-600">
                                {mismatch.possible_cids.join(', ')}
                              </span>
                            </div>
                          )}
                          <div className="mt-2 space-y-1">
                            {mismatch.issues && mismatch.issues.map((issue: string, issueIdx: number) => (
                              <div key={issueIdx} className="flex items-start gap-2">
                                <AlertCircle className="h-3 w-3 text-red-500 mt-0.5" />
                                <span className="text-xs text-red-600">{issue}</span>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        )}

        {/* Perfect Match */}
        {results.mismatches_found === 0 && (
          <Card className="border-green-500">
            <CardHeader>
              <CardTitle className="text-sm font-medium text-green-600 flex items-center gap-2">
                <CheckCircle className="h-4 w-4" />
                All Payees Match
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-gray-600">
                All {results.checks_processed} checks have payees that correctly match their CID records
                in the investor and vendor tables.
              </p>
            </CardContent>
          </Card>
        )}

        {/* Modal for Row Data */}
        <Dialog open={!!selectedRowData} onOpenChange={() => setSelectedRowData(null)}>
          <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Complete Check Record</DialogTitle>
              <DialogDescription>
                All fields from the CHECKS.dbf record
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-2 mt-4">
              {selectedRowData && Object.entries(selectedRowData).map(([key, value]) => (
                <div key={key} className="grid grid-cols-3 gap-2 py-1 border-b">
                  <span className="font-medium text-sm">{key}:</span>
                  <span className="col-span-2 text-sm">
                    {value === null || value === undefined ? 
                      '<null>' : 
                      typeof value === 'boolean' ? 
                        (value ? 'True' : 'False') : 
                        String(value)}
                  </span>
                </div>
              ))}
            </div>
          </DialogContent>
        </Dialog>
      </div>
    )
  }

  const currentAudit = getCurrentAudit()

  // If Follow Batch Number is selected, show that component instead
  if (showFollowBatch) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold tracking-tight">Audit Tools</h2>
            <p className="text-muted-foreground">
              Follow Batch Number - Search across multiple tables
            </p>
          </div>
          <Button 
            variant="outline" 
            onClick={() => {
              setShowFollowBatch(false)
              setSelectedAudit(null)
            }}
          >
            Back to Audit Selection
          </Button>
        </div>
        <FollowBatchNumber />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Audit Tools</h2>
          <p className="text-muted-foreground">
            Comprehensive financial data validation and analysis for {companyName}
          </p>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="grid gap-6">
        {/* Audit Selection Card */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle>Select Audit Type</CardTitle>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" className="w-64">
                    {currentAudit ? currentAudit.item.name : 'Choose an audit...'}
                    <ChevronDown className="ml-2 h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-80">
                  {auditCategories.map(category => {
                    const Icon = category.icon
                    return (
                      <DropdownMenuSub key={category.id}>
                        <DropdownMenuSubTrigger>
                          <Icon className="mr-2 h-4 w-4" />
                          {category.name}
                        </DropdownMenuSubTrigger>
                        <DropdownMenuSubContent className="w-72">
                          {category.items.map(item => (
                            <DropdownMenuItem
                              key={item.id}
                              onClick={() => {
                                if (item.id === 'follow-batch-number') {
                                  // Go straight to Follow Batch Number interface
                                  setShowFollowBatch(true)
                                  setSelectedAudit(null)
                                } else {
                                  setSelectedAudit({ category: category.id, item: item.id })
                                }
                              }}
                              className="flex flex-col items-start py-2"
                            >
                              <div className="font-medium">{item.name}</div>
                              <div className="text-xs text-gray-500">{item.description}</div>
                            </DropdownMenuItem>
                          ))}
                        </DropdownMenuSubContent>
                      </DropdownMenuSub>
                    )
                  })}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
            {currentAudit && (
              <CardDescription className="mt-2">
                {currentAudit.category.name} → {currentAudit.item.description}
              </CardDescription>
            )}
          </CardHeader>
          
          {selectedAudit && (
            <CardContent className="space-y-4">
              {/* Filters based on audit type */}
              <div className="grid gap-4 md:grid-cols-3">
                {(selectedAudit.item === 'gl-validation' || selectedAudit.item === 'year-analysis' || selectedAudit.item === 'account-analysis' || selectedAudit.item === 'check-gl-matching') && (
                  <div>
                    <Label htmlFor="account-filter">Account Number (Optional)</Label>
                    <Input
                      id="account-filter"
                      value={accountFilter}
                      onChange={(e) => setAccountFilter(e.target.value)}
                      placeholder="e.g., 102000"
                      className="mt-1"
                    />
                  </div>
                )}
                
                {(selectedAudit.category === 'vendor-audits') && (
                  <div>
                    <Label htmlFor="vendor-filter">Vendor Code (Optional)</Label>
                    <Input
                      id="vendor-filter"
                      value={vendorFilter}
                      onChange={(e) => setVendorFilter(e.target.value)}
                      placeholder="e.g., V001"
                      className="mt-1"
                    />
                  </div>
                )}
                
                {(selectedAudit.category === 'customer-audits') && (
                  <div>
                    <Label htmlFor="customer-filter">Customer Code (Optional)</Label>
                    <Input
                      id="customer-filter"
                      value={customerFilter}
                      onChange={(e) => setCustomerFilter(e.target.value)}
                      placeholder="e.g., C001"
                      className="mt-1"
                    />
                  </div>
                )}
                
                {(selectedAudit.item === 'duplicate-detection' || selectedAudit.item === 'year-analysis') && (
                  <div>
                    <Label htmlFor="year-filter">Year (Optional)</Label>
                    <Input
                      id="year-filter"
                      value={yearFilter}
                      onChange={(e) => setYearFilter(e.target.value)}
                      placeholder="e.g., 2024"
                      className="mt-1"
                    />
                  </div>
                )}
                
                {selectedAudit.item === 'check-gl-matching' && (
                  <>
                    <div>
                      <Label htmlFor="start-date">Start Date (Optional)</Label>
                      <Input
                        id="start-date"
                        type="date"
                        value={startDate}
                        onChange={(e) => setStartDate(e.target.value)}
                        className="mt-1"
                      />
                    </div>
                    <div>
                      <Label htmlFor="end-date">End Date (Optional)</Label>
                      <Input
                        id="end-date"
                        type="date"
                        value={endDate}
                        onChange={(e) => setEndDate(e.target.value)}
                        className="mt-1"
                      />
                    </div>
                  </>
                )}
              </div>

              {/* Run Button */}
              <div className="flex items-center gap-4">
                <Button 
                  onClick={runAudit}
                  disabled={loading || !selectedAudit}
                  className="w-full sm:w-auto"
                >
                  {loading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Running Audit...
                    </>
                  ) : (
                    <>
                      <FileSearch className="mr-2 h-4 w-4" />
                      Run Audit
                    </>
                  )}
                </Button>
                
                {results && (
                  <Button variant="outline" onClick={() => {
                    setResults(null)
                    setVoidIssueFilter('all')
                  }}>
                    Clear Results
                  </Button>
                )}
              </div>
            </CardContent>
          )}
        </Card>

        {/* Error Display */}
        {error && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* Results Display */}
        {results && !error && (
          <Card>
            <CardHeader>
              <CardTitle>Audit Results</CardTitle>
              {currentAudit && (
                <CardDescription>
                  {currentAudit.item.name} - Completed at {new Date().toLocaleTimeString()}
                </CardDescription>
              )}
            </CardHeader>
            <CardContent>
              {renderResults()}
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}