import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Button } from './ui/button'
import { AlertCircle, CheckCircle, Clock, FileSearch, Loader2, TrendingUp, TrendingDown, Calendar, ShieldCheck, AlertTriangle } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle } from './ui/alert'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Input } from './ui/input'
import { Label } from './ui/label'

interface AuditToolsProps {
  currentUser: any
  companyName: string
}

export default function AuditTools({ currentUser, companyName }: AuditToolsProps) {
  const [loading, setLoading] = useState(false)
  const [loadingYearly, setLoadingYearly] = useState(false)
  const [loadingValidation, setLoadingValidation] = useState(false)
  const [periodCheckResults, setPeriodCheckResults] = useState<any>(null)
  const [yearlyBalanceResults, setYearlyBalanceResults] = useState<any>(null)
  const [validationResults, setValidationResults] = useState<any>(null)
  const [selectedAccount, setSelectedAccount] = useState<string>('102000')
  const [glAccounts, setGLAccounts] = useState<any[]>([])
  const [error, setError] = useState<string>('')
  const [yearlyError, setYearlyError] = useState<string>('')
  const [validationError, setValidationError] = useState<string>('')

  const checkGLPeriodFields = async () => {
    setLoading(true)
    setError('')
    setPeriodCheckResults(null)

    try {
      // @ts-ignore
      const results = await window.go.main.App.CheckGLPeriodFields(companyName)
      setPeriodCheckResults(results)
    } catch (err: any) {
      setError(err.message || 'Failed to check GL period fields')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadGLAccounts()
  }, [companyName])

  const loadGLAccounts = async () => {
    try {
      // @ts-ignore
      const coaData = await window.go.main.App.GetDBFTableData(companyName, 'COA.dbf')
      if (coaData && coaData.rows) {
        const accounts = coaData.rows.map((row: any[]) => ({
          account_number: row[0],
          account_type: row[1],
          description: row[2]
        }))
        setGLAccounts(accounts)
      }
    } catch (err) {
      console.error('Failed to load GL accounts', err)
    }
  }

  const analyzeGLBalancesByYear = async () => {
    setLoadingYearly(true)
    setYearlyError('')
    setYearlyBalanceResults(null)

    try {
      // @ts-ignore
      const results = await window.go.main.App.AnalyzeGLBalancesByYear(companyName, selectedAccount)
      setYearlyBalanceResults(results)
    } catch (err: any) {
      setYearlyError(err.message || 'Failed to analyze GL balances by year')
    } finally {
      setLoadingYearly(false)
    }
  }

  const validateGLBalances = async () => {
    setLoadingValidation(true)
    setValidationError('')
    setValidationResults(null)

    try {
      // @ts-ignore
      const results = await window.go.main.App.ValidateGLBalances(companyName, '')
      setValidationResults(results)
    } catch (err: any) {
      setValidationError(err.message || 'Failed to validate GL balances')
    } finally {
      setLoadingValidation(false)
    }
  }

  const formatPercentage = (value: string) => {
    const num = parseFloat(value)
    if (num > 50) return <span className="text-red-600 font-semibold">{value}</span>
    if (num > 10) return <span className="text-amber-600 font-semibold">{value}</span>
    return <span className="text-green-600">{value}</span>
  }

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(value)
  }

  const getAccountTypeName = (typeNum: number) => {
    switch (typeNum) {
      case 1: return 'Asset'
      case 2: return 'Liability'
      case 3: return 'Equity'
      case 4: return 'Revenue'
      case 5: return 'Expense'
      default: return 'Unknown'
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Audit Tools</h2>
          <p className="text-muted-foreground">
            Data validation and integrity checks for {companyName}
          </p>
        </div>
      </div>

      {/* GL Activity Summary Audit */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Calendar className="h-5 w-5" />
            GL Activity Summary Audit
          </CardTitle>
          <CardDescription>
            Comprehensive audit of GL activity by year to identify trends, anomalies, and data quality issues
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <Button 
              onClick={checkGLPeriodFields}
              disabled={loading}
              className="w-full sm:w-auto"
            >
              {loading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Checking GL Records...
                </>
              ) : (
                <>
                  <FileSearch className="mr-2 h-4 w-4" />
                  Check GL Period Fields
                </>
              )}
            </Button>

            {error && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {periodCheckResults && (
              <div className="space-y-4">
                {/* Summary Stats */}
                <div className="grid gap-4 md:grid-cols-3">
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Total GL Records</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{periodCheckResults.total_rows?.toLocaleString()}</div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Blank Year Fields</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {periodCheckResults.blank_year_count?.toLocaleString()}
                      </div>
                      <p className="text-xs text-muted-foreground">
                        {formatPercentage(periodCheckResults.blank_year_pct || '0%')}
                      </p>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Blank Period Fields</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {periodCheckResults.blank_period_count?.toLocaleString()}
                      </div>
                      <p className="text-xs text-muted-foreground">
                        {formatPercentage(periodCheckResults.blank_period_pct || '0%')}
                      </p>
                    </CardContent>
                  </Card>
                </div>

                {/* Unique Values */}
                <div className="grid gap-4 md:grid-cols-2">
                  {periodCheckResults.unique_years && Object.keys(periodCheckResults.unique_years).length > 0 && (
                    <Card>
                      <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium">Years Found</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div className="space-y-1 max-h-40 overflow-y-auto">
                          {Object.entries(periodCheckResults.unique_years)
                            .sort(([a], [b]) => b.localeCompare(a))
                            .map(([year, count]) => (
                              <div key={year} className="flex justify-between text-sm">
                                <span className="font-mono">{year}</span>
                                <span className="text-muted-foreground">{count} records</span>
                              </div>
                            ))}
                        </div>
                      </CardContent>
                    </Card>
                  )}

                  {periodCheckResults.unique_periods && Object.keys(periodCheckResults.unique_periods).length > 0 && (
                    <Card>
                      <CardHeader className="pb-2">
                        <CardTitle className="text-sm font-medium">Periods Found</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div className="space-y-1 max-h-40 overflow-y-auto">
                          {Object.entries(periodCheckResults.unique_periods)
                            .sort(([a], [b]) => a.localeCompare(b))
                            .map(([period, count]) => (
                              <div key={period} className="flex justify-between text-sm">
                                <span className="font-mono">{period}</span>
                                <span className="text-muted-foreground">{count} records</span>
                              </div>
                            ))}
                        </div>
                      </CardContent>
                    </Card>
                  )}
                </div>

                {/* Sample Blank Rows */}
                {periodCheckResults.sample_blank_rows && periodCheckResults.sample_blank_rows.length > 0 && (
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Sample Records with Blank Period Fields</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-2">
                        {periodCheckResults.sample_blank_rows.map((row: any, idx: number) => (
                          <div key={idx} className="p-2 bg-muted rounded-md text-sm font-mono">
                            <div>Row #{row.row_index}: Account {row.account || 'N/A'}</div>
                            <div className="text-xs text-muted-foreground">
                              Debit: ${parseFloat(row.debit || 0).toFixed(2)} | 
                              Credit: ${parseFloat(row.credit || 0).toFixed(2)}
                            </div>
                          </div>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                )}

                {/* Analysis Summary */}
                <Alert>
                  <AlertCircle className="h-4 w-4" />
                  <AlertTitle>Analysis Summary</AlertTitle>
                  <AlertDescription>
                    {periodCheckResults.blank_both_count > 0 ? (
                      <div className="space-y-2">
                        <p>
                          Found <strong>{periodCheckResults.blank_both_count}</strong> records with both CYEAR and CPERIOD blank.
                        </p>
                        <p>
                          These records may be causing discrepancies in balance calculations. FoxPro may handle these differently
                          by excluding them from period-based calculations or treating them as opening balances.
                        </p>
                      </div>
                    ) : (
                      <p>All GL records have valid year and period values.</p>
                    )}
                  </AlertDescription>
                </Alert>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* GL Balance Analysis by Year */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Calendar className="h-5 w-5" />
            GL Balance Analysis by Year
          </CardTitle>
          <CardDescription>
            Analyze GL balances by year for any account to identify trends and anomalies
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="account-select">Select GL Account</Label>
                <div className="flex gap-2">
                  <Input
                    id="account-input"
                    type="text"
                    placeholder="Enter account number"
                    value={selectedAccount}
                    onChange={(e) => setSelectedAccount(e.target.value)}
                    className="flex-1"
                  />
                  <Select value={selectedAccount} onValueChange={setSelectedAccount}>
                    <SelectTrigger className="w-[300px]">
                      <SelectValue placeholder="Or select from list" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">All Accounts</SelectItem>
                      {glAccounts.map((account) => (
                        <SelectItem key={account.account_number} value={account.account_number}>
                          {account.account_number} - {account.description}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <p className="text-xs text-muted-foreground">
                  Enter account number (e.g., 102000) or select from the dropdown
                </p>
              </div>

              <div className="flex items-end">
                <Button 
                  onClick={analyzeGLBalancesByYear}
                  disabled={loadingYearly}
                  className="w-full sm:w-auto"
                >
                  {loadingYearly ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Analyzing...
                    </>
                  ) : (
                    <>
                      <TrendingUp className="mr-2 h-4 w-4" />
                      Analyze Yearly Balances
                    </>
                  )}
                </Button>
              </div>
            </div>

            {yearlyError && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>{yearlyError}</AlertDescription>
              </Alert>
            )}

            {yearlyBalanceResults && (
              <div className="space-y-4">
                {/* Account Info */}
                <Alert>
                  <AlertTitle>Account: {yearlyBalanceResults.account_number || 'All Accounts'}</AlertTitle>
                  <AlertDescription>
                    {yearlyBalanceResults.account_number && glAccounts.find(a => a.account_number === yearlyBalanceResults.account_number) && (
                      <div>
                        <p>{glAccounts.find(a => a.account_number === yearlyBalanceResults.account_number)?.description}</p>
                        <p className="font-semibold">
                          Account Type: {getAccountTypeName(glAccounts.find(a => a.account_number === yearlyBalanceResults.account_number)?.account_type)}
                        </p>
                      </div>
                    )}
                  </AlertDescription>
                </Alert>

                {/* Summary Totals */}
                <div className="grid gap-4 md:grid-cols-4">
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Total Debits</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-xl font-bold">{formatCurrency(yearlyBalanceResults.total_debits || 0)}</div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Total Credits</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-xl font-bold">{formatCurrency(yearlyBalanceResults.total_credits || 0)}</div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Overall Balance</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className={`text-xl font-bold ${yearlyBalanceResults.overall_balance >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                        {formatCurrency(yearlyBalanceResults.overall_balance || 0)}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Total Records</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-xl font-bold">{yearlyBalanceResults.total_records?.toLocaleString() || 0}</div>
                      <p className="text-xs text-muted-foreground">{yearlyBalanceResults.years_found} years</p>
                    </CardContent>
                  </Card>
                </div>

                {/* Yearly Breakdown */}
                {yearlyBalanceResults.yearly_balances && yearlyBalanceResults.yearly_balances.length > 0 && (
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm font-medium">Balance by Year</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-2">
                        <div className="grid grid-cols-6 gap-2 text-xs font-medium text-muted-foreground border-b pb-2">
                          <div>Year</div>
                          <div className="text-right">Debits</div>
                          <div className="text-right">Credits</div>
                          <div className="text-right">Balance</div>
                          <div className="text-right">Records</div>
                          <div className="text-right">Periods</div>
                        </div>
                        {yearlyBalanceResults.yearly_balances.map((year: any) => (
                          <div key={year.year} className="grid grid-cols-6 gap-2 text-sm py-1 hover:bg-muted/50 rounded">
                            <div className="font-mono">{year.year}</div>
                            <div className="text-right">{formatCurrency(year.debits)}</div>
                            <div className="text-right">{formatCurrency(year.credits)}</div>
                            <div className={`text-right font-semibold ${year.balance >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                              {formatCurrency(year.balance)}
                            </div>
                            <div className="text-right">{year.record_count}</div>
                            <div className="text-right">{year.periods}</div>
                          </div>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                )}

                {/* Blank Year Records */}
                {yearlyBalanceResults.blank_year_totals && (
                  <Alert variant="destructive">
                    <AlertCircle className="h-4 w-4" />
                    <AlertTitle>Records with Blank Year</AlertTitle>
                    <AlertDescription>
                      <div className="space-y-1 mt-2">
                        <p>Found {yearlyBalanceResults.blank_year_totals.record_count} records with blank year field</p>
                        <p>Total Debits: {formatCurrency(yearlyBalanceResults.blank_year_totals.debits)}</p>
                        <p>Total Credits: {formatCurrency(yearlyBalanceResults.blank_year_totals.credits)}</p>
                        <p>Net Balance: {formatCurrency(yearlyBalanceResults.blank_year_totals.balance)}</p>
                        <p className="font-semibold text-red-600">
                          These records may be causing balance discrepancies!
                        </p>
                      </div>
                    </AlertDescription>
                  </Alert>
                )}
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* GL Validation Checks */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ShieldCheck className="h-5 w-5" />
            GL Validation Checks
          </CardTitle>
          <CardDescription>
            Comprehensive validation to ensure GL integrity and identify data quality issues
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <Button 
              onClick={validateGLBalances}
              disabled={loadingValidation}
              className="w-full sm:w-auto"
            >
              {loadingValidation ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Running Validation...
                </>
              ) : (
                <>
                  <ShieldCheck className="mr-2 h-4 w-4" />
                  Run GL Validation
                </>
              )}
            </Button>

            {validationError && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>{validationError}</AlertDescription>
              </Alert>
            )}

            {validationResults && (
              <div className="space-y-4">
                {/* Overall Balance Check */}
                <Alert className={validationResults.is_balanced ? "" : "border-red-500 bg-red-50"}>
                  <div className="flex items-center gap-2">
                    {validationResults.is_balanced ? (
                      <CheckCircle className="h-4 w-4 text-green-600" />
                    ) : (
                      <AlertTriangle className="h-4 w-4 text-red-600" />
                    )}
                    <AlertTitle>
                      {validationResults.is_balanced ? "GL is Balanced" : "GL Out of Balance!"}
                    </AlertTitle>
                  </div>
                  <AlertDescription>
                    <div className="mt-2 space-y-1">
                      <p>Total Debits: {formatCurrency(validationResults.total_debits)}</p>
                      <p>Total Credits: {formatCurrency(validationResults.total_credits)}</p>
                      <p className={validationResults.is_balanced ? "text-green-600" : "text-red-600 font-bold"}>
                        Difference: {formatCurrency(validationResults.overall_difference)}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        Checked {validationResults.total_rows_checked?.toLocaleString()} GL entries
                      </p>
                    </div>
                  </AlertDescription>
                </Alert>

                {/* Year-by-Year Balance Checks */}
                {validationResults.year_balance_checks && validationResults.year_balance_checks.length > 0 && (
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm font-medium">Balance Check by Year</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-2">
                        <div className="grid grid-cols-5 gap-2 text-xs font-medium text-muted-foreground border-b pb-2">
                          <div>Year</div>
                          <div className="text-right">Debits</div>
                          <div className="text-right">Credits</div>
                          <div className="text-right">Difference</div>
                          <div className="text-center">Status</div>
                        </div>
                        {validationResults.year_balance_checks.map((year: any) => (
                          <div key={year.year} className="grid grid-cols-5 gap-2 text-sm py-1">
                            <div className="font-mono">{year.year}</div>
                            <div className="text-right">{formatCurrency(year.debits)}</div>
                            <div className="text-right">{formatCurrency(year.credits)}</div>
                            <div className="text-right">{formatCurrency(year.difference)}</div>
                            <div className="text-center">
                              {year.balanced ? (
                                <span className="text-green-600">✓</span>
                              ) : (
                                <span className="text-red-600">✗</span>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                )}

                {/* Data Quality Issues */}
                <div className="grid gap-4 md:grid-cols-3">
                  {/* Duplicate Transactions */}
                  <Card className={validationResults.duplicate_count > 0 ? "border-amber-500" : ""}>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Duplicate Transactions</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {validationResults.duplicate_count || 0}
                      </div>
                      {validationResults.duplicate_count > 0 && (
                        <p className="text-xs text-amber-600 mt-1">
                          Potential duplicate entries detected
                        </p>
                      )}
                    </CardContent>
                  </Card>

                  {/* Zero Amount Transactions */}
                  <Card className={validationResults.zero_amount_transactions > 0 ? "border-amber-500" : ""}>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Zero Amount Entries</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {validationResults.zero_amount_transactions || 0}
                      </div>
                      {validationResults.zero_amount_transactions > 0 && (
                        <p className="text-xs text-amber-600 mt-1">
                          Entries with no debit or credit
                        </p>
                      )}
                    </CardContent>
                  </Card>

                  {/* Suspicious Amounts */}
                  <Card className={validationResults.suspicious_count > 0 ? "border-red-500" : ""}>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Suspicious Amounts</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {validationResults.suspicious_count || 0}
                      </div>
                      {validationResults.suspicious_count > 0 && (
                        <p className="text-xs text-red-600 mt-1">
                          Transactions over $1M detected
                        </p>
                      )}
                    </CardContent>
                  </Card>
                </div>

                {/* Imbalanced Accounts */}
                {validationResults.imbalanced_accounts && validationResults.imbalanced_accounts.length > 0 && (
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm font-medium">
                        Top Imbalanced Accounts ({validationResults.imbalanced_count} total)
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-2 max-h-60 overflow-y-auto">
                        <div className="grid grid-cols-4 gap-2 text-xs font-medium text-muted-foreground border-b pb-2">
                          <div>Account</div>
                          <div className="text-right">Debits</div>
                          <div className="text-right">Credits</div>
                          <div className="text-right">Difference</div>
                        </div>
                        {validationResults.imbalanced_accounts.slice(0, 10).map((account: any, idx: number) => (
                          <div key={idx} className="grid grid-cols-4 gap-2 text-sm py-1">
                            <div className="font-mono">{account.account}</div>
                            <div className="text-right">{formatCurrency(account.debits)}</div>
                            <div className="text-right">{formatCurrency(account.credits)}</div>
                            <div className="text-right text-red-600 font-semibold">
                              {formatCurrency(account.difference)}
                            </div>
                          </div>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                )}

                {/* Sample Duplicate Transactions */}
                {validationResults.duplicate_transactions && validationResults.duplicate_transactions.length > 0 && (
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-sm font-medium">Sample Duplicate Transactions</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-2">
                        {validationResults.duplicate_transactions.slice(0, 5).map((dup: any, idx: number) => (
                          <div key={idx} className="p-2 bg-amber-50 rounded-md text-sm">
                            <div className="font-mono">Account: {dup.account}</div>
                            <div className="text-xs text-muted-foreground">
                              Debit: {formatCurrency(dup.debit)} | Credit: {formatCurrency(dup.credit)}
                            </div>
                            <div className="text-xs text-amber-600">
                              Found in rows: {dup.row_indices.join(', ')}
                            </div>
                          </div>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                )}
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}