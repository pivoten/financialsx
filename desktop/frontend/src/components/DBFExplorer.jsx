import { useState, useEffect, useCallback, useRef, useMemo } from 'react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table'
import { Select } from './ui/select'
import { Folder, FileText, Search, Edit, Save, X, Plus, Trash2, ChevronUp, ChevronDown, ChevronsUpDown } from 'lucide-react'
import { GetDBFFiles, GetDBFTableData, SearchDBFTable, UpdateDBFRecord } from '../../wailsjs/go/main/App'

// Global API call tracker to prevent rapid calls across all instances
let globalApiCallInProgress = false
let lastGlobalApiCall = 0

export function DBFExplorer() {
  const [dbfFiles, setDbfFiles] = useState([])
  const [selectedFile, setSelectedFile] = useState('')
  const [tableData, setTableData] = useState({ columns: [], rows: [], stats: {} })
  const [loading, setLoading] = useState(false)
  const [editingCell, setEditingCell] = useState(null)
  const [editValue, setEditValue] = useState('')
  const [searchTerm, setSearchTerm] = useState('')
  const [currentCompany, setCurrentCompany] = useState('')
  const [sortColumn, setSortColumn] = useState(null)
  const [sortDirection, setSortDirection] = useState('asc') // 'asc' or 'desc'
  const [isServerSearching, setIsServerSearching] = useState(false)
  const lastClickTime = useRef(0)
  const clickTimeout = useRef(null)
  const searchTimeout = useRef(null)

  // Load current company and DBF files on component mount
  useEffect(() => {
    loadCurrentCompanyFiles()
  }, [])

  const loadCurrentCompanyFiles = async () => {
    setLoading(true)
    try {
      // Get current company from localStorage (set during login)
      const companyName = localStorage.getItem('company_name')
      console.log('Loading company files, company_name from localStorage:', companyName)
      if (!companyName) {
        console.error('No company found in session')
        return
      }
      
      setCurrentCompany(companyName)
      
      // Load DBF files for the current company
      const result = await GetDBFFiles(companyName)
      setDbfFiles(result || [])
    } catch (error) {
      console.error('Failed to load DBF files:', error)
      setDbfFiles([])
    } finally {
      setLoading(false)
    }
  }

  const loadTableData = async (fileName) => {
    const now = Date.now()
    console.log('loadTableData called for:', fileName, 'at', now)
    console.log('currentCompany value:', currentCompany)
    
    // Global API call protection
    if (globalApiCallInProgress) {
      console.log('BLOCKED: Global API call already in progress')
      setLoading(false)
      return
    }
    
    // Minimum time between API calls (1 second)
    if (now - lastGlobalApiCall < 1000) {
      console.log('BLOCKED: API call too soon after last call')
      setLoading(false)
      return
    }
    
    // Set global lock
    globalApiCallInProgress = true
    lastGlobalApiCall = now
    setLoading(true)
    
    try {
      console.log('Making API call to GetDBFTableData')
      console.log('Parameters: company =', currentCompany, ', fileName =', fileName)
      
      // If currentCompany is empty, try to get it from localStorage again
      let companyToUse = currentCompany || localStorage.getItem('company_name') || 'cantrellenergy'
      console.log('Company to use for API call:', companyToUse)
      
      if (!companyToUse) {
        console.error('No company available for API call')
        setTableData({ columns: [], rows: [] })
        return
      }
      
      const result = await GetDBFTableData(companyToUse, fileName)
      console.log('API call completed successfully', result)
      
      // Ensure we have valid data structure even for empty files
      const safeResult = {
        columns: result?.columns || [],
        rows: result?.rows || [],
        stats: result?.stats || {}
      }
      
      setTableData(safeResult)
      setSortColumn(null) // Reset sorting when loading new data
      setSortDirection('asc')
    } catch (error) {
      console.error('Failed to load table data:', error)
      setTableData({ columns: [], rows: [] })
    } finally {
      // Release global lock after a delay
      setTimeout(() => {
        globalApiCallInProgress = false
        setLoading(false)
      }, 500)
    }
  }

  const handleFileSelect = useCallback((fileName) => {
    const now = Date.now()
    console.log('handleFileSelect called with:', fileName, 'at', now)
    
    // Global protection first
    if (globalApiCallInProgress) {
      console.log('BLOCKED: Global API call in progress, ignoring click')
      return
    }
    
    // Prevent multiple calls while loading
    if (loading) {
      console.log('Already loading, ignoring click')
      return
    }
    
    // Prevent calling same file multiple times
    if (selectedFile === fileName) {
      console.log('Same file already selected, ignoring')
      return
    }
    
    // Debounce rapid clicks (prevent clicks within 1 second)
    if (now - lastClickTime.current < 1000) {
      console.log('Click too fast, ignoring (debounced)')
      return
    }
    
    lastClickTime.current = now
    
    // Clear any pending timeout
    if (clickTimeout.current) {
      clearTimeout(clickTimeout.current)
    }
    
    // Set selected file immediately
    setSelectedFile(fileName)
    
    // Call API directly without timeout to reduce complexity
    loadTableData(fileName)
  }, [])

  const handleCellEdit = (rowIndex, columnIndex) => {
    const currentValue = tableData.rows[rowIndex]?.[columnIndex] || ''
    setEditingCell({ row: rowIndex, col: columnIndex })
    setEditValue(currentValue)
  }

  const handleSaveEdit = async () => {
    if (!editingCell) return
    
    try {
      // Update local state immediately for better UX
      const newRows = [...tableData.rows]
      newRows[editingCell.row][editingCell.col] = editValue
      setTableData({ ...tableData, rows: newRows })
      
      // Save changes to the DBF file
      await UpdateDBFRecord(
        currentCompany, 
        selectedFile, 
        editingCell.row, 
        editingCell.col, 
        editValue
      )
      
      setEditingCell(null)
      setEditValue('')
    } catch (error) {
      console.error('Failed to save changes:', error)
      // Reload data on error to restore original state
      loadTableData(selectedFile)
    }
  }

  const handleCancelEdit = () => {
    setEditingCell(null)
    setEditValue('')
  }

  // Server-side search function
  const performServerSearch = async (search) => {
    if (!selectedFile || !currentCompany) return
    
    console.log('Performing server search:', search)
    setIsServerSearching(true)
    
    try {
      const companyToUse = currentCompany || localStorage.getItem('company_name') || 'cantrellenergy'
      
      if (search.trim() === '') {
        // If search is empty, load normal data
        const result = await GetDBFTableData(companyToUse, selectedFile)
        const safeResult = {
          columns: result?.columns || [],
          rows: result?.rows || [],
          stats: result?.stats || {}
        }
        setTableData(safeResult)
      } else {
        // Perform server-side search
        const result = await SearchDBFTable(companyToUse, selectedFile, search)
        const safeResult = {
          columns: result?.columns || [],
          rows: result?.rows || [],
          stats: result?.stats || {}
        }
        setTableData(safeResult)
      }
    } catch (error) {
      console.error('Search failed:', error)
      setTableData({ columns: [], rows: [], stats: {} })
    } finally {
      setIsServerSearching(false)
    }
  }

  // Debounced search handler
  const handleSearchChange = (value) => {
    setSearchTerm(value)
    
    // Clear existing timeout
    if (searchTimeout.current) {
      clearTimeout(searchTimeout.current)
    }
    
    // Set new timeout for debounced search
    searchTimeout.current = setTimeout(() => {
      performServerSearch(value)
    }, 500) // Wait 500ms after user stops typing
  }

  const handleSort = (columnIndex) => {
    if (sortColumn === columnIndex) {
      // Toggle direction if same column
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      // New column, start with ascending
      setSortColumn(columnIndex)
      setSortDirection('asc')
    }
  }

  // Sort rows (no client-side filtering since we use server-side search)
  let filteredRows = [...(tableData.rows || [])]

  // Apply sorting if a column is selected
  if (sortColumn !== null) {
    filteredRows = [...filteredRows].sort((a, b) => {
      const aVal = a[sortColumn]
      const bVal = b[sortColumn]
      
      // Handle null/empty values
      if (!aVal && !bVal) return 0
      if (!aVal) return sortDirection === 'asc' ? 1 : -1
      if (!bVal) return sortDirection === 'asc' ? -1 : 1
      
      // Try numeric comparison first
      const aNum = parseFloat(aVal)
      const bNum = parseFloat(bVal)
      if (!isNaN(aNum) && !isNaN(bNum)) {
        return sortDirection === 'asc' ? aNum - bNum : bNum - aNum
      }
      
      // Fall back to string comparison
      const aStr = aVal.toString()
      const bStr = bVal.toString()
      return sortDirection === 'asc' 
        ? aStr.localeCompare(bStr)
        : bStr.localeCompare(aStr)
    })
  }

  return (
    <div className="space-y-6">
      {/* DBF File Selection */}
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <FileText className="w-4 h-4" />
          <span className="font-medium">Select DBF File:</span>
        </div>
        <div className="flex-1 max-w-sm">
          {loading ? (
            <div className="text-sm text-muted-foreground">Loading files...</div>
          ) : (
            <Select
              value={selectedFile}
              onChange={(e) => {
                e.preventDefault()
                e.stopPropagation()
                handleFileSelect(e.target.value)
              }}
              disabled={loading}
              className="w-full"
            >
              <option value="">
                {dbfFiles.length === 0 ? 'No DBF files found' : 'Choose a DBF file...'}
              </option>
              {dbfFiles.map((file) => (
                <option key={file} value={file}>
                  {file}
                </option>
              ))}
            </Select>
          )}
        </div>
        {currentCompany && (
          <div className="text-sm text-muted-foreground">
            Company: {currentCompany}
          </div>
        )}
      </div>

      {/* Table Data Display */}
      {selectedFile && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <FileText className="w-4 h-4" />
                  {selectedFile}
                </CardTitle>
                <CardDescription>
                  {tableData.stats?.totalRecords && (
                    <div className="flex items-center gap-4 text-sm">
                      <span>Total: {tableData.stats.totalRecords.toLocaleString()} records</span>
                      <span className="text-green-600">Active: {tableData.stats.activeRecords.toLocaleString()}</span>
                      <span className="text-red-600">Deleted: {tableData.stats.deletedRecords.toLocaleString()}</span>
                      {tableData.stats.searchTerm ? (
                        <>
                          <span className="text-purple-600">
                            Search matches: {tableData.stats.loadedRecords.toLocaleString()} 
                            {tableData.stats.hasMoreRecords && " (limited to 1000)"}
                          </span>
                          {isServerSearching && <span className="text-gray-500">Searching...</span>}
                        </>
                      ) : (
                        <>
                          <span className="text-blue-600">Loaded: {tableData.stats.loadedRecords.toLocaleString()}</span>
                          {tableData.stats.hasMoreRecords && (
                            <span className="text-orange-600">(More available)</span>
                          )}
                        </>
                      )}
                    </div>
                  )}
                </CardDescription>
              </div>
              <div className="flex items-center gap-2">
                <div className="relative">
                  <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="Search all records..."
                    value={searchTerm}
                    onChange={(e) => handleSearchChange(e.target.value)}
                    className="pl-8 w-64"
                    disabled={isServerSearching}
                  />
                </div>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="text-center py-8">Loading table data...</div>
            ) : tableData.columns.length > 0 ? (
              <div className="border rounded-md">
                <div className="max-h-[600px] overflow-auto relative">
                  <Table>
                    <TableHeader className="sticky top-0 bg-background z-10 shadow-sm">
                      <TableRow>
                        {tableData.columns.map((column, index) => (
                          <TableHead 
                            key={index} 
                            className="font-semibold cursor-pointer hover:bg-muted/50 select-none"
                            onClick={() => handleSort(index)}
                          >
                            <div className="flex items-center gap-1">
                              {column}
                              <span className="ml-auto">
                                {sortColumn === index ? (
                                  sortDirection === 'asc' ? 
                                    <ChevronUp className="w-4 h-4" /> : 
                                    <ChevronDown className="w-4 h-4" />
                                ) : (
                                  <ChevronsUpDown className="w-4 h-4 opacity-30" />
                                )}
                              </span>
                            </div>
                          </TableHead>
                        ))}
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredRows.map((row, rowIndex) => (
                        <TableRow key={rowIndex}>
                          {row.map((cell, cellIndex) => (
                            <TableCell key={cellIndex}>
                              {editingCell?.row === rowIndex && editingCell?.col === cellIndex ? (
                                <div className="flex items-center gap-2">
                                  <Input
                                    value={editValue}
                                    onChange={(e) => setEditValue(e.target.value)}
                                    className="h-8"
                                    autoFocus
                                    onKeyDown={(e) => {
                                      if (e.key === 'Enter') handleSaveEdit()
                                      if (e.key === 'Escape') handleCancelEdit()
                                    }}
                                  />
                                  <Button size="sm" onClick={handleSaveEdit}>
                                    <Save className="w-3 h-3" />
                                  </Button>
                                  <Button size="sm" variant="outline" onClick={handleCancelEdit}>
                                    <X className="w-3 h-3" />
                                  </Button>
                                </div>
                              ) : (
                                <div
                                  className="cursor-pointer hover:bg-muted/50 p-1 rounded min-h-6"
                                  onClick={() => handleCellEdit(rowIndex, cellIndex)}
                                >
                                  {cell}
                                </div>
                              )}
                            </TableCell>
                          ))}
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
                
                {filteredRows.length === 0 && searchTerm && (
                  <div className="text-center py-8 text-muted-foreground">
                    No records match your search
                  </div>
                )}
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                {selectedFile ? 'No data found in this file' : 'Select a DBF file to view data'}
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}