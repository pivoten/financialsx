import React, { useState, useEffect } from 'react'
import { GetOutstandingChecks } from '../../wailsjs/go/main/App'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Badge } from './ui/badge'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './ui/table'
import { 
  AlertTriangle,
  RefreshCw,
  Calendar,
  DollarSign
} from 'lucide-react'

const OutstandingChecks = ({ companyName, currentUser }) => {
  const [outstandingChecks, setOutstandingChecks] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [totalAmount, setTotalAmount] = useState(0)

  // Calculate days outstanding
  const calculateDaysOutstanding = (checkDate) => {
    if (!checkDate) return 'N/A'
    
    try {
      const today = new Date()
      const checkDateTime = new Date(checkDate)
      if (isNaN(checkDateTime.getTime())) return 'N/A'
      
      const diffTime = today - checkDateTime
      const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
      return diffDays
    } catch (err) {
      return 'N/A'
    }
  }

  // Format currency
  const formatCurrency = (amount) => {
    if (typeof amount !== 'number') return '$0.00'
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  // Format date
  const formatDate = (dateStr) => {
    if (!dateStr) return 'N/A'
    try {
      const date = new Date(dateStr)
      if (isNaN(date.getTime())) return dateStr
      return date.toLocaleDateString()
    } catch (err) {
      return dateStr
    }
  }

  // Load outstanding checks
  const loadOutstandingChecks = async () => {
    if (!companyName) return
    
    setLoading(true)
    setError('')
    
    try {
      console.log('Loading outstanding checks for company:', companyName)
      const result = await GetOutstandingChecks(companyName)
      
      if (result.status === 'error') {
        setError(result.error || 'Failed to load outstanding checks')
        setOutstandingChecks([])
      } else {
        const checks = result.checks || []
        setOutstandingChecks(checks)
        
        // Calculate total amount
        const total = checks.reduce((sum, check) => sum + (check.amount || 0), 0)
        setTotalAmount(total)
        
        console.log(`Loaded ${checks.length} outstanding checks, total: $${total.toFixed(2)}`)
      }
    } catch (err) {
      console.error('Failed to load outstanding checks:', err)
      setError(err.message || 'Failed to load outstanding checks')
      setOutstandingChecks([])
    } finally {
      setLoading(false)
    }
  }

  // Load checks when component mounts or company changes
  useEffect(() => {
    loadOutstandingChecks()
  }, [companyName])

  // Get badge variant for days outstanding
  const getDaysOutstandingBadge = (days) => {
    if (days === 'N/A') return { variant: 'secondary', text: 'N/A' }
    if (days <= 30) return { variant: 'default', text: `${days} days` }
    if (days <= 60) return { variant: 'secondary', text: `${days} days` }
    if (days <= 90) return { variant: 'destructive', text: `${days} days` }
    return { variant: 'destructive', text: `${days} days (STALE)` }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Outstanding Checks</h2>
          <p className="text-muted-foreground">
            Checks that have not been cleared by the bank
          </p>
        </div>
        <Button onClick={loadOutstandingChecks} disabled={loading}>
          <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Outstanding Checks
            </CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{outstandingChecks.length}</div>
            <p className="text-xs text-muted-foreground">
              Checks not yet cleared
            </p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Total Amount
            </CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatCurrency(totalAmount)}</div>
            <p className="text-xs text-muted-foreground">
              Outstanding check amount
            </p>
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

      {/* Outstanding Checks Table */}
      <Card>
        <CardHeader>
          <CardTitle>Outstanding Checks List</CardTitle>
          <CardDescription>
            {loading ? 'Loading checks...' : 
             outstandingChecks.length === 0 ? 'No outstanding checks found' :
             `Showing ${outstandingChecks.length} outstanding checks`}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <RefreshCw className="h-6 w-6 animate-spin" />
              <span className="ml-2">Loading outstanding checks...</span>
            </div>
          ) : outstandingChecks.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              {error ? 'Unable to load checks' : 'No outstanding checks found'}
            </div>
          ) : (
            <div className="overflow-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Check #</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Payee</TableHead>
                    <TableHead className="text-right">Amount</TableHead>
                    <TableHead>Account</TableHead>
                    <TableHead className="text-center">Days Outstanding</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {outstandingChecks.map((check, index) => {
                    const daysOut = calculateDaysOutstanding(check.date)
                    const daysBadge = getDaysOutstandingBadge(daysOut)
                    
                    return (
                      <TableRow key={index}>
                        <TableCell className="font-medium">
                          {check.checkNumber || 'N/A'}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center">
                            <Calendar className="mr-2 h-4 w-4 text-muted-foreground" />
                            {formatDate(check.date)}
                          </div>
                        </TableCell>
                        <TableCell>
                          {check.payee || 'N/A'}
                        </TableCell>
                        <TableCell className="text-right font-medium">
                          {formatCurrency(check.amount)}
                        </TableCell>
                        <TableCell>
                          {check.account || 'N/A'}
                        </TableCell>
                        <TableCell className="text-center">
                          <Badge variant={daysBadge.variant}>
                            {daysBadge.text}
                          </Badge>
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
    </div>
  )
}

export default OutstandingChecks