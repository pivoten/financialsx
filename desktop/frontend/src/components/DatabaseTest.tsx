
import { useState, useEffect } from 'react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './ui/table'
import { Database, Play, AlertCircle, CheckCircle, Loader2, RefreshCw } from 'lucide-react'
import { TestDatabaseQuery, GetTableList } from '../../wailsjs/go/main/App'
import { User, DatabaseTestResult } from '../types'

export function DatabaseTest({ currentUser }: { currentUser: User | null }) {
  const [loading, setLoading] = useState(false)
  const [testResult, setTestResult] = useState<DatabaseTestResult | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [selectedTable, setSelectedTable] = useState('')
  const [customQuery, setCustomQuery] = useState('')
  const [tables, setTables] = useState<string[]>([])
  const [loadingTables, setLoadingTables] = useState(false)
  const [sortColumn, setSortColumn] = useState<string | null>(null)
  const [sortDirection, setSortDirection] = useState('asc')
  
  useEffect(() => {
    const timer = setTimeout(() => {
      loadTables()
    }, 100)
    return () => clearTimeout(timer)
  }, [currentUser])

  const testQueries = {
    'COA': 'SELECT TOP 10 * FROM COA ORDER BY 1',
    'CHECKS': 'SELECT TOP 10 * FROM CHECKS ORDER BY 1',
    'GLMASTER': 'SELECT TOP 10 * FROM GLMASTER ORDER BY 1',
    'VENDORS': 'SELECT TOP 10 * FROM VENDORS ORDER BY 1',
    'WELLS': 'SELECT TOP 10 * FROM WELLS ORDER BY 1',
    'CUSTOM': ''
  }
  
  const getQueryForTable = (tableName: string) => {
    if (tableName in testQueries) return testQueries[tableName as keyof typeof testQueries]
    return `SELECT TOP 10 * FROM ${tableName} ORDER BY 1`
  }

  const loadTables = async () => {
    setLoadingTables(true)
    try {
      const companyPath = localStorage.getItem('company_path')
      const companyName = localStorage.getItem('company_name')
      const companyIdentifier = companyPath || currentUser?.company_name || companyName
      if (!companyIdentifier) {
        setLoadingTables(false)
        return
      }
      const result = await GetTableList(companyIdentifier)
      if (result.success && result.tables) {
        setTables(result.tables)
      }
    } catch (err) {
    } finally {
      setLoadingTables(false)
    }
  }

  const handleSort = (column: string) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortColumn(column)
      setSortDirection('asc')
    }
  }

  const getSortedData = () => {
    if (!testResult?.data || !sortColumn) return testResult?.data || []
    return [...testResult.data].sort((a, b) => {
      const aVal = a[sortColumn] || ''
      const bVal = b[sortColumn] || ''
      const aNum = parseFloat(aVal)
      const bNum = parseFloat(bVal)
      if (!isNaN(aNum) && !isNaN(bNum)) {
        return sortDirection === 'asc' ? aNum - bNum : bNum - aNum
      } else {
        const comparison = String(aVal).localeCompare(String(bVal))
        return sortDirection === 'asc' ? comparison : -comparison
      }
    })
  }

  const runTest = async () => {
    setLoading(true)
    setError(null)
    setTestResult(null)
    try {
      let query = ''
      if (selectedTable === 'CUSTOM') {
        query = customQuery
      } else if (selectedTable) {
        query = getQueryForTable(selectedTable)
      } else {
        throw new Error('Please select a table or enter a custom query')
      }
      if (!query) throw new Error('Query cannot be empty')
      const companyPath = localStorage.getItem('company_path')
      const companyName = localStorage.getItem('company_name')
      const companyIdentifier = companyPath || currentUser?.company_name || companyName
      const result = await TestDatabaseQuery(companyIdentifier, query)
      if (result.success) {
        setTestResult(result as DatabaseTestResult)
      } else {
        setError(result.error || 'Query failed')
      }
    } catch (err) {
      setError(err.message || 'Failed to execute query')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Database className="w-5 h-5" />
            Database Connection Test
          </CardTitle>
          <CardDescription>
            Test your Pivoten.DbApi connection and run queries against your FoxPro database
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label>Select Table or Query Type</Label>
            <div className="flex gap-2">
              <select 
                value={selectedTable} 
                onChange={(e: React.ChangeEvent<HTMLSelectElement>) => setSelectedTable(e.target.value)}
                className="flex-1 p-2 border rounded-md bg-white"
              >
                <option value="">Choose a table to query</option>
                {loadingTables && (
                  <option disabled>Loading tables...</option>
                )}
                {!loadingTables && tables.length > 0 ? (
                  tables.map(table => (
                    <option key={table} value={table}>{table}</option>
                  ))
                ) : !loadingTables ? (
                  <>
                    <option value="COA">Chart of Accounts (COA)</option>
                    <option value="CHECKS">Checks Register</option>
                    <option value="GLMASTER">General Ledger</option>
                    <option value="VENDORS">Vendors</option>
                    <option value="WELLS">Wells</option>
                  </>
                ) : null}
                <option value="CUSTOM">--- Custom SQL Query ---</option>
              </select>
              <Button
                variant="outline"
                size="sm"
                onClick={loadTables}
                disabled={loadingTables}
                title="Refresh table list from database"
              >
                {loadingTables ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <RefreshCw className="w-4 h-4" />
                )}
              </Button>
            </div>
            <p className="text-xs text-gray-500">Selected: {selectedTable || 'none'} | Tables loaded: {tables.length}</p>
          </div>

          {selectedTable === 'CUSTOM' && (
            <div className="space-y-2">
              <Label>Custom SQL Query</Label>
              <textarea
                className="w-full min-h-[100px] p-3 border rounded-md font-mono text-sm"
                placeholder="SELECT * FROM tablename WHERE condition"
                value={customQuery}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setCustomQuery(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">
                Enter a FoxPro SQL query. Use .T. for true and .F. for false in WHERE clauses.
              </p>
            </div>
          )}

          {selectedTable && selectedTable !== 'CUSTOM' && (
            <div className="p-3 bg-muted rounded-md">
              <p className="text-sm font-mono">{getQueryForTable(selectedTable)}</p>
            </div>
          )}

          <div className="flex gap-3">
            <Button 
              onClick={runTest} 
              disabled={loading || !selectedTable}
              className="flex items-center gap-2"
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Running Query...
                </>
              ) : (
                <>
                  <Play className="w-4 h-4" />
                  Run Test Query
                </>
              )}
            </Button>

            <Button 
              variant="outline"
              onClick={loadTables}
              disabled={loadingTables}
            >
              {loadingTables ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin mr-2" />
                  Loading...
                </>
              ) : (
                'List All Tables'
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {(testResult || error) && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              {error ? (
                <>
                  <AlertCircle className="w-5 h-5 text-destructive" />
                  Query Failed
                </>
              ) : (
                <>
                  <CheckCircle className="w-5 h-5 text-green-600" />
                  Query Successful
                </>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {error ? (
              <div className="p-4 bg-destructive/10 text-destructive rounded-md">
                <p className="font-semibold">Error:</p>
                <p className="text-sm mt-1">{error}</p>
              </div>
            ) : testResult && (
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-muted-foreground">Server:</span>
                    <span className="ml-2 font-medium">Pivoten.DbApi</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Database:</span>
                    <span className="ml-2 font-medium">{testResult.database || currentUser.company_name}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Rows Returned:</span>
                    <span className="ml-2 font-medium">{testResult.rowCount || 0}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Execution Time:</span>
                    <span className="ml-2 font-medium">{testResult.executionTime || 'N/A'}</span>
                  </div>
                </div>

                {testResult.data && testResult.data.length > 0 && (
                  <div className="border rounded-md overflow-hidden">
                    <div className="max-h-[500px] overflow-auto">
                      <table className="w-full">
                        <thead className="sticky top-0 z-10 bg-white border-b">
                          <tr>
                            {Object.keys(testResult.data[0]).map((col) => (
                              <th 
                                key={col} 
                                className="text-left p-2 font-mono text-xs bg-gray-50 hover:bg-gray-100 cursor-pointer transition-colors"
                                onClick={() => handleSort(col)}
                              >
                                <div className="flex items-center gap-1">
                                  {col}
                                  <span className={`transition-colors ${sortColumn === col ? 'text-blue-600' : 'text-gray-400'}`}>
                                    {sortColumn === col 
                                      ? (sortDirection === 'asc' ? '\u2191' : '\u2193')
                                      : '\u2195'
                                    }
                                  </span>
                                </div>
                              </th>
                            ))}
                          </tr>
                        </thead>
                        <tbody>
                          {getSortedData().map((row: any, idx: number) => (
                            <tr key={idx} className="border-b hover:bg-gray-50">
                              {Object.values(row).map((val, cidx) => (
                                <td key={cidx} className="p-2 font-mono text-xs">
                                  {val !== null && val !== undefined ? String(val) : ''}
                                </td>
                              ))}
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                )}

                {testResult.raw && (
                  <details className="mt-4">
                    <summary className="cursor-pointer text-sm text-muted-foreground">
                      View Raw Response
                    </summary>
                    <pre className="mt-2 p-3 bg-muted rounded-md text-xs overflow-auto">
                      {JSON.stringify(testResult.raw, null, 2)}
                    </pre>
                  </details>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {tables.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Available Tables in Database</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 gap-2">
              {tables.map((table) => (
                <div key={table} className="p-2 bg-muted rounded text-sm font-mono">
                  {table}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
