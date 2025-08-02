import { useState, useEffect, useCallback, useRef, useMemo } from 'react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table'
import { Select } from './ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog'
import { Folder, FileText, Search, Edit, Save, X, Plus, Trash2, ChevronUp, ChevronDown, ChevronsUpDown, Settings, Eye, EyeOff, GripVertical, Download, Upload, Filter, FilterX, Database } from 'lucide-react'
import { GetDBFFiles, GetDBFTableData, GetDBFTableDataPaged, SearchDBFTable, UpdateDBFRecord } from '../../wailsjs/go/main/App'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import {
  useSortable,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'

// Global API call tracker to prevent rapid calls across all instances
let globalApiCallInProgress = false
let lastGlobalApiCall = 0

export function DBFExplorer({ currentUser }) {
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
  const [currentPage, setCurrentPage] = useState(0)
  const [pageSize] = useState(100) // Load 100 records at a time
  const [hasMoreData, setHasMoreData] = useState(false)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [allLoadedRows, setAllLoadedRows] = useState([])
  const [columnOrder, setColumnOrder] = useState([])
  const [hiddenColumns, setHiddenColumns] = useState(new Set())
  const [showColumnSettings, setShowColumnSettings] = useState(false)
  const [showColumnFilters, setShowColumnFilters] = useState(false)
  const [columnFilters, setColumnFilters] = useState([])
  const [showDataExport, setShowDataExport] = useState(false)
  const [isEditMode, setIsEditMode] = useState(false)
  const [selectedRecord, setSelectedRecord] = useState(null)
  const [showRecordModal, setShowRecordModal] = useState(false)
  const lastClickTime = useRef(0)
  const clickTimeout = useRef(null)
  const searchTimeout = useRef(null)

  // Load current company and DBF files on component mount
  useEffect(() => {
    loadCurrentCompanyFiles()
  }, [])
  
  // Save column preferences when they change
  useEffect(() => {
    if (selectedFile && columnOrder.length > 0) {
      const prefs = {
        columnOrder,
        hiddenColumns: Array.from(hiddenColumns)
      }
      localStorage.setItem(`dbf_columns_${selectedFile}`, JSON.stringify(prefs))
    }
  }, [columnOrder, hiddenColumns, selectedFile])

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

  const loadTableData = async (fileName, resetPagination = true) => {
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
    
    if (resetPagination) {
      setCurrentPage(0)
      setAllLoadedRows([])
    }
    
    try {
      console.log('Making API call to GetDBFTableDataPaged')
      console.log('Parameters: company =', currentCompany, ', fileName =', fileName)
      
      // If currentCompany is empty, try to get it from localStorage again
      let companyToUse = currentCompany || localStorage.getItem('company_name') || 'cantrellenergy'
      console.log('Company to use for API call:', companyToUse)
      
      if (!companyToUse) {
        console.error('No company available for API call')
        setTableData({ columns: [], rows: [] })
        return
      }
      
      const offset = resetPagination ? 0 : currentPage * pageSize
      const sortCol = sortColumn !== null ? tableData.columns?.[sortColumn] : ''
      
      console.log('Calling GetDBFTableDataPaged with:', {
        company: companyToUse,
        fileName,
        offset,
        limit: pageSize,
        sortColumn: sortCol,
        sortDirection
      })
      
      const result = await GetDBFTableDataPaged(
        companyToUse, 
        fileName, 
        offset, 
        pageSize, 
        sortCol, 
        sortDirection
      )
      console.log('API call completed successfully', result)
      
      // Ensure we have valid data structure even for empty files
      const safeResult = {
        columns: result?.columns || [],
        rows: result?.rows || [],
        stats: result?.stats || {}
      }
      
      if (resetPagination) {
        setTableData(safeResult)
        setAllLoadedRows(safeResult.rows || [])
        
        // Initialize column order and load saved preferences
        if (safeResult.columns.length > 0) {
          // Try to load saved preferences
          const savedPrefs = localStorage.getItem(`dbf_columns_${fileName}`)
          if (savedPrefs) {
            try {
              const prefs = JSON.parse(savedPrefs)
              // Validate that saved columns still exist
              const validOrder = prefs.columnOrder.filter(idx => idx < safeResult.columns.length)
              setColumnOrder(validOrder.length > 0 ? validOrder : safeResult.columns.map((_, index) => index))
              setHiddenColumns(new Set(prefs.hiddenColumns.filter(idx => idx < safeResult.columns.length)))
            } catch (e) {
              // If preferences are invalid, use defaults
              setColumnOrder(safeResult.columns.map((_, index) => index))
            }
          } else {
            // No saved preferences, use defaults
            setColumnOrder(safeResult.columns.map((_, index) => index))
          }
        }
      } else {
        // Append new rows to existing data for pagination
        const newRows = [...allLoadedRows, ...(safeResult.rows || [])]
        setAllLoadedRows(newRows)
        setTableData({
          ...safeResult,
          rows: newRows
        })
      }
      
      // Check if there's more data available
      const totalMatching = result?.stats?.totalMatching || 0
      const currentlyLoaded = resetPagination ? (safeResult.rows?.length || 0) : allLoadedRows.length + (safeResult.rows?.length || 0)
      setHasMoreData(currentlyLoaded < totalMatching)
      
      if (!resetPagination) {
        setCurrentPage(prev => prev + 1)
      }
      
    } catch (error) {
      console.error('Failed to load table data:', error)
      setTableData({ columns: [], rows: [] })
      setAllLoadedRows([])
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
    
    // Reset pagination state when selecting new file
    setCurrentPage(0)
    setAllLoadedRows([])
    setSortColumn(null)
    setSortDirection('asc')
    setColumnOrder([])
    setHiddenColumns(new Set())
    setColumnFilters([])
    
    // Call API directly without timeout to reduce complexity
    loadTableData(fileName, true)
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

  // Check if current user can edit
  const canEdit = () => {
    console.log('DBFExplorer canEdit check:', {
      currentUser,
      is_root: currentUser?.is_root,
      role_name: currentUser?.role_name,
      hasUser: !!currentUser
    })
    return currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')
  }

  // Handle row click to show record details
  const handleRowClick = (rowIndex) => {
    if (!isEditMode) {
      const record = tableData.rows[rowIndex]
      const recordWithColumns = tableData.columns.map((col, index) => ({
        column: col,
        value: record[index],
        index: index
      }))
      setSelectedRecord(recordWithColumns)
      setShowRecordModal(true)
    }
  }

  // Toggle edit mode
  const toggleEditMode = () => {
    setIsEditMode(!isEditMode)
    // Cancel any ongoing edits when switching modes
    if (!isEditMode) {
      setEditingCell(null)
      setEditValue('')
    }
  }

  // Format values for display (logical, dates, nulls)
  const formatLogicalValue = (value) => {
    // Handle null, undefined, empty values
    if (value === null || value === undefined || value === '') {
      return ''
    }
    
    // Handle boolean values
    if (typeof value === 'boolean') {
      return value ? 'True' : 'False'
    }
    
    // Handle string values
    if (typeof value === 'string') {
      const lowerValue = value.toLowerCase()
      
      // Handle logical string values
      if (lowerValue === 't' || lowerValue === '.t.' || lowerValue === 'true') {
        return 'True'
      }
      if (lowerValue === 'f' || lowerValue === '.f.' || lowerValue === 'false') {
        return 'False'
      }
      
      // Handle null-like string values
      if (lowerValue === 'null' || lowerValue === 'nil' || lowerValue.trim() === '') {
        return ''
      }
    }
    
    return value
  }

  // Server-side search function
  const performServerSearch = async (search) => {
    if (!selectedFile || !currentCompany) return
    
    console.log('Performing server search:', search)
    setIsServerSearching(true)
    
    try {
      const companyToUse = currentCompany || localStorage.getItem('company_name') || 'cantrellenergy'
      
      if (search.trim() === '') {
        // If search is empty, load normal paginated data
        setCurrentPage(0)
        setAllLoadedRows([])
        loadTableData(selectedFile, true)
      } else {
        // Perform server-side search - use SearchDBFTable for now (could be enhanced later)
        const result = await SearchDBFTable(companyToUse, selectedFile, search)
        const safeResult = {
          columns: result?.columns || [],
          rows: result?.rows || [],
          stats: result?.stats || {}
        }
        setTableData(safeResult)
        setAllLoadedRows(safeResult.rows || [])
        setHasMoreData(false) // Search results don't paginate for now
      }
    } catch (error) {
      console.error('Search failed:', error)
      setTableData({ columns: [], rows: [], stats: {} })
      setAllLoadedRows([])
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
    const newDirection = sortColumn === columnIndex ? (sortDirection === 'asc' ? 'desc' : 'asc') : 'asc'
    setSortColumn(columnIndex)
    setSortDirection(newDirection)
    
    // Reload data with new sort parameters
    if (selectedFile) {
      loadTableData(selectedFile, true)
    }
  }

  // Apply column filters to the data with AND/OR logic
  const applyColumnFilters = (rows) => {
    if (columnFilters.length === 0) return rows
    
    return rows.filter(row => {
      // Helper function to evaluate a single filter
      const evaluateFilter = (filter) => {
        if (!filter.column || !filter.value) return true
        
        const columnIndex = tableData.columns?.indexOf(filter.column)
        if (columnIndex === -1) return true
        
        const cellValue = row[columnIndex]
        if (cellValue == null) return false
        
        let cellStr = String(cellValue)
        let filterValue = String(filter.value)
        
        if (!filter.caseSensitive) {
          cellStr = cellStr.toLowerCase()
          filterValue = filterValue.toLowerCase()
        }
        
        switch (filter.operator) {
          case 'equals':
            return cellStr === filterValue
          case 'contains':
            return cellStr.includes(filterValue)
          case 'startsWith':
            return cellStr.startsWith(filterValue)
          case 'endsWith':
            return cellStr.endsWith(filterValue)
          case 'notEquals':
            return cellStr !== filterValue
          case 'notContains':
            return !cellStr.includes(filterValue)
          case 'greaterThan':
            const numA = parseFloat(cellStr)
            const numB = parseFloat(filterValue)
            return !isNaN(numA) && !isNaN(numB) && numA > numB
          case 'lessThan':
            const numC = parseFloat(cellStr)
            const numD = parseFloat(filterValue)
            return !isNaN(numC) && !isNaN(numD) && numC < numD
          default:
            return true
        }
      }
      
      // Process filters with AND/OR logic
      if (columnFilters.length === 1) {
        return evaluateFilter(columnFilters[0])
      }
      
      // Start with the first filter (no logical operator)
      let result = evaluateFilter(columnFilters[0])
      
      // Process remaining filters with their logical operators
      for (let i = 1; i < columnFilters.length; i++) {
        const filter = columnFilters[i]
        const filterResult = evaluateFilter(filter)
        
        if (filter.logicalOperator === 'OR') {
          result = result || filterResult
        } else { // Default to AND
          result = result && filterResult
        }
      }
      
      return result
    })
  }
  
  // Use server-side sorted data, then apply client-side column filters
  const filteredRows = applyColumnFilters(tableData.rows || [])
  
  // Apply column order and hidden columns to display
  const displayColumns = columnOrder.length > 0 
    ? columnOrder.filter(idx => !hiddenColumns.has(idx)).map(idx => ({ 
        name: tableData.columns[idx], 
        index: idx 
      }))
    : tableData.columns?.map((col, idx) => ({ name: col, index: idx })).filter(col => !hiddenColumns.has(col.index)) || []
  
  // Function to load more data when scrolling
  const loadMoreData = async () => {
    if (isLoadingMore || !hasMoreData || !selectedFile) return
    
    console.log('Loading more data...')
    setIsLoadingMore(true)
    
    try {
      await loadTableData(selectedFile, false) // Don't reset pagination
    } catch (error) {
      console.error('Failed to load more data:', error)
    } finally {
      setIsLoadingMore(false)
    }
  }
  
  // Scroll event handler for virtual loading
  const handleScroll = (e) => {
    const { scrollTop, scrollHeight, clientHeight } = e.target
    
    // Load more data when user scrolls near the bottom (within 200px)
    if (scrollHeight - scrollTop - clientHeight < 200 && hasMoreData && !isLoadingMore) {
      loadMoreData()
    }
  }
  
  // Export column settings to JSON file
  const exportColumnSettings = () => {
    if (!selectedFile || !tableData.columns) return
    
    const settings = {
      file: selectedFile,
      exportDate: new Date().toISOString(),
      columns: tableData.columns.map((col, idx) => ({
        name: col,
        originalIndex: idx,
        visible: !hiddenColumns.has(idx)
      })),
      columnOrder: columnOrder,
      hiddenColumns: Array.from(hiddenColumns),
      columnFilters: columnFilters,
      version: "1.0"
    }
    
    const blob = new Blob([JSON.stringify(settings, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${selectedFile}_columns_${new Date().toISOString().split('T')[0]}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }
  
  // Import column settings from JSON file
  const handleImportSettings = (e) => {
    const file = e.target.files?.[0]
    if (!file) return
    
    const reader = new FileReader()
    reader.onload = (event) => {
      try {
        const settings = JSON.parse(event.target.result)
        
        // Validate the imported settings
        if (!settings.columnOrder || !Array.isArray(settings.columnOrder)) {
          throw new Error('Invalid settings file: missing columnOrder')
        }
        
        // Check if columns match current table
        const currentColumns = tableData.columns || []
        const importedColumns = settings.columns || []
        
        // Create a mapping of column names to indices for current table
        const columnMap = new Map()
        currentColumns.forEach((col, idx) => {
          columnMap.set(col, idx)
        })
        
        // Map imported order to current column indices
        const newOrder = []
        const newHidden = new Set()
        
        // Try to match columns by name
        importedColumns.forEach((importedCol) => {
          const currentIdx = columnMap.get(importedCol.name)
          if (currentIdx !== undefined) {
            newOrder.push(currentIdx)
            if (!importedCol.visible) {
              newHidden.add(currentIdx)
            }
          }
        })
        
        // Add any new columns that weren't in the import
        currentColumns.forEach((col, idx) => {
          if (!newOrder.includes(idx)) {
            newOrder.push(idx)
          }
        })
        
        // Apply the imported settings
        setColumnOrder(newOrder)
        setHiddenColumns(newHidden)
        
        // Import column filters if they exist and are valid
        if (settings.columnFilters && Array.isArray(settings.columnFilters)) {
          // Validate filters against current columns
          const validFilters = settings.columnFilters.filter(filter => {
            return filter.column && currentColumns.includes(filter.column)
          })
          setColumnFilters(validFilters)
        }
        
        // Show success message (you might want to add a toast notification here)
        console.log('Column settings imported successfully')
        
      } catch (error) {
        console.error('Failed to import settings:', error)
        alert('Failed to import settings. Please check the file format.')
      }
    }
    
    reader.readAsText(file)
    
    // Clear the input so the same file can be imported again
    e.target.value = ''
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
                      <span className="text-red-600">Soft Deletes: {tableData.stats.deletedRecords.toLocaleString()}</span>
                      {tableData.stats.searchTerm && (
                        <>
                          <span className="text-purple-600">
                            Search matches: {tableData.stats.loadedRecords.toLocaleString()} 
                            {tableData.stats.hasMoreRecords && " (limited to 1000)"}
                          </span>
                          {isServerSearching && <span className="text-gray-500">Searching...</span>}
                        </>
                      )}
                      {columnFilters.length > 0 && (
                        <span className="text-blue-600">
                          Filtered: {filteredRows.length.toLocaleString()} of {(tableData.rows || []).length.toLocaleString()}
                          {columnFilters.length > 1 && (
                            <span className="text-xs ml-1">
                              ({columnFilters.some(f => f.logicalOperator === 'OR') ? 'AND/OR' : 'AND'})
                            </span>
                          )}
                        </span>
                      )}
                    </div>
                  )}
                </CardDescription>
              </div>
              <div className="flex items-center gap-2">
                {canEdit() && (
                  <Button
                    variant={isEditMode ? "default" : "outline"}
                    size="sm"
                    onClick={toggleEditMode}
                    title={isEditMode ? "Exit Edit Mode" : "Enter Edit Mode"}
                  >
                    <Edit className="w-4 h-4" />
                    {isEditMode ? "Exit Edit" : "Edit"}
                  </Button>
                )}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowColumnSettings(!showColumnSettings)}
                  title="Column Settings"
                >
                  <Settings className="w-4 h-4" />
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowColumnFilters(!showColumnFilters)}
                  title="Column Filters"
                  className={columnFilters.length > 0 ? "bg-blue-50 border-blue-300" : ""}
                >
                  <Filter className="w-4 h-4" />
                  {columnFilters.length > 0 && (
                    <span className="ml-1 text-xs bg-blue-600 text-white rounded-full px-1">
                      {columnFilters.length}
                    </span>
                  )}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowDataExport(!showDataExport)}
                  title="Export Data"
                >
                  <Database className="w-4 h-4" />
                </Button>
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
            {/* Column Settings Panel */}
            {showColumnSettings && (
              <div className="mb-4 p-4 border rounded-lg bg-muted/50">
                <div className="flex justify-between items-center mb-3">
                  <h4 className="font-semibold text-sm">Column Settings</h4>
                  <Button 
                    variant="ghost" 
                    size="sm" 
                    onClick={() => setShowColumnSettings(false)}
                  >
                    <X className="w-4 h-4" />
                  </Button>
                </div>
                <ColumnSettingsList 
                  columnOrder={columnOrder}
                  setColumnOrder={setColumnOrder}
                  tableColumns={tableData.columns}
                  hiddenColumns={hiddenColumns}
                  setHiddenColumns={setHiddenColumns}
                />
                <div className="flex gap-2 mt-3 pt-3 border-t">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setColumnOrder(tableData.columns.map((_, i) => i))
                      setHiddenColumns(new Set())
                    }}
                  >
                    Reset to Default
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setHiddenColumns(new Set())}
                  >
                    Show All
                  </Button>
                  <div className="ml-auto flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => exportColumnSettings()}
                      title="Export column settings"
                    >
                      <Download className="w-4 h-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => document.getElementById('import-settings')?.click()}
                      title="Import column settings"
                    >
                      <Upload className="w-4 h-4" />
                    </Button>
                    <input
                      id="import-settings"
                      type="file"
                      accept=".json"
                      className="hidden"
                      onChange={handleImportSettings}
                    />
                  </div>
                </div>
              </div>
            )}
            
            {/* Column Filters Panel */}
            {showColumnFilters && (
              <div className="mb-4 p-4 border rounded-lg bg-blue-50/50">
                <div className="flex justify-between items-center mb-3">
                  <h4 className="font-semibold text-sm flex items-center gap-2">
                    <Filter className="w-4 h-4" />
                    Column Filters
                  </h4>
                  <div className="flex gap-2">
                    {columnFilters.length > 0 && (
                      <Button 
                        variant="ghost" 
                        size="sm"
                        onClick={() => setColumnFilters([])}
                        title="Clear all filters"
                      >
                        <FilterX className="w-4 h-4" />
                      </Button>
                    )}
                    <Button 
                      variant="ghost" 
                      size="sm" 
                      onClick={() => setShowColumnFilters(false)}
                    >
                      <X className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
                
                <div className="space-y-3">
                  {columnFilters.map((filter, index) => (
                    <div key={index}>
                      {index > 0 && (
                        <div className="flex justify-center py-2">
                          <div className="flex items-center gap-2 px-3 py-1 bg-gray-100 rounded-full border">
                            <button
                              type="button"
                              onClick={() => {
                                const newFilters = [...columnFilters]
                                newFilters[index] = { ...filter, logicalOperator: 'AND' }
                                setColumnFilters(newFilters)
                              }}
                              className={`px-2 py-1 text-xs font-medium rounded ${
                                filter.logicalOperator === 'AND' || !filter.logicalOperator
                                  ? 'bg-blue-600 text-white' 
                                  : 'bg-transparent text-gray-600 hover:bg-gray-200'
                              }`}
                            >
                              AND
                            </button>
                            <button
                              type="button"
                              onClick={() => {
                                const newFilters = [...columnFilters]
                                newFilters[index] = { ...filter, logicalOperator: 'OR' }
                                setColumnFilters(newFilters)
                              }}
                              className={`px-2 py-1 text-xs font-medium rounded ${
                                filter.logicalOperator === 'OR'
                                  ? 'bg-green-600 text-white' 
                                  : 'bg-transparent text-gray-600 hover:bg-gray-200'
                              }`}
                            >
                              OR
                            </button>
                          </div>
                        </div>
                      )}
                      <ColumnFilter
                        filter={filter}
                        index={index}
                        columns={tableData.columns || []}
                        onUpdate={(updatedFilter) => {
                          const newFilters = [...columnFilters]
                          newFilters[index] = updatedFilter
                          setColumnFilters(newFilters)
                        }}
                        onRemove={() => {
                          const newFilters = columnFilters.filter((_, i) => i !== index)
                          setColumnFilters(newFilters)
                        }}
                      />
                    </div>
                  ))}
                  
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setColumnFilters([...columnFilters, {
                        column: tableData.columns?.[0] || '',
                        operator: 'contains',
                        value: '',
                        caseSensitive: false,
                        logicalOperator: columnFilters.length > 0 ? 'AND' : null
                      }])
                    }}
                    disabled={!tableData.columns || tableData.columns.length === 0}
                  >
                    <Plus className="w-4 h-4 mr-2" />
                    Add Filter
                  </Button>
                </div>
              </div>
            )}
            
            {/* Data Export Panel */}
            {showDataExport && (
              <div className="mb-4 p-4 border rounded-lg bg-green-50/50">
                <div className="flex justify-between items-center mb-3">
                  <h4 className="font-semibold text-sm flex items-center gap-2">
                    <Database className="w-4 h-4" />
                    Export Data
                  </h4>
                  <Button 
                    variant="ghost" 
                    size="sm" 
                    onClick={() => setShowDataExport(false)}
                  >
                    <X className="w-4 h-4" />
                  </Button>
                </div>
                
                <DataExportOptions
                  selectedFile={selectedFile}
                  tableData={tableData}
                  filteredRows={filteredRows}
                  displayColumns={displayColumns}
                  columnFilters={columnFilters}
                  allLoadedRows={allLoadedRows}
                />
              </div>
            )}
            
            {loading ? (
              <div className="text-center py-8">Loading table data...</div>
            ) : tableData.columns.length > 0 ? (
              <div className="border rounded-md">
                <div className="relative h-[600px] overflow-auto" onScroll={handleScroll}>
                  <Table>
                    <TableHeader className="sticky top-0 z-10 bg-background">
                      <TableRow>
                        {displayColumns.map((col) => (
                          <TableHead 
                            key={col.index} 
                            className="cursor-pointer hover:bg-muted/50 select-none"
                            onClick={() => handleSort(col.index)}
                          >
                            <div className="flex items-center gap-1">
                              {col.name}
                              <span className="ml-auto">
                                {sortColumn === col.index ? (
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
                          {displayColumns.map((col) => {
                            const cellValue = row[col.index]
                            return (
                              <TableCell key={col.index}>
                                {editingCell?.row === rowIndex && editingCell?.col === col.index ? (
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
                                    className={`p-1 rounded min-h-6 ${
                                      isEditMode 
                                        ? "cursor-pointer hover:bg-muted/50" 
                                        : "cursor-pointer hover:bg-blue-50"
                                    }`}
                                    onClick={() => {
                                      if (isEditMode) {
                                        handleCellEdit(rowIndex, col.index)
                                      } else {
                                        handleRowClick(rowIndex)
                                      }
                                    }}
                                  >
                                    {formatLogicalValue(cellValue)}
                                  </div>
                                )}
                              </TableCell>
                            )
                          })}
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                  
                  {/* Loading indicator for pagination */}
                  {isLoadingMore && (
                    <div className="text-center py-4 text-muted-foreground">
                      Loading more records...
                    </div>
                  )}
                  
                  {/* End of data indicator */}
                  {!hasMoreData && filteredRows.length > pageSize && (
                    <div className="text-center py-4 text-muted-foreground text-sm">
                      All {filteredRows.length.toLocaleString()} records loaded
                    </div>
                  )}
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
      
      {/* Record Detail Modal */}
      {showRecordModal && selectedRecord && (
        <Dialog open={showRecordModal} onOpenChange={setShowRecordModal}>
          <DialogContent className="max-w-4xl max-h-[80vh] overflow-auto">
            <DialogHeader>
              <DialogTitle>Record Details - {selectedFile}</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div className="text-sm text-muted-foreground">
                Complete record data for row {tableData.rows?.indexOf(selectedRecord.map(r => r.value)) + 1 || 'N/A'}
              </div>
              <div className="grid gap-3">
                {selectedRecord.map((field, index) => (
                  <div key={index} className="grid grid-cols-3 gap-4 p-3 border rounded">
                    <div className="font-medium text-sm text-muted-foreground">
                      {field.column}
                    </div>
                    <div className="col-span-2 break-all">
                      <div className="p-2 bg-muted/30 rounded text-sm">
                        {formatLogicalValue(field.value) || <span className="text-muted-foreground italic">Empty</span>}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
              <div className="flex justify-end pt-4 border-t">
                <Button onClick={() => setShowRecordModal(false)}>Close</Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

// Sortable Column Item Component
function SortableColumnItem({ id, column, index, isVisible, onToggleVisibility, visibleIndex }) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`flex items-center gap-2 p-2 rounded hover:bg-background ${
        isDragging ? 'shadow-lg bg-background' : ''
      }`}
    >
      <div
        {...attributes}
        {...listeners}
        className="cursor-grab hover:cursor-grabbing"
      >
        <GripVertical className="w-4 h-4 text-muted-foreground" />
      </div>
      <input
        type="checkbox"
        checked={isVisible}
        onChange={(e) => onToggleVisibility(index, e.target.checked)}
        className="rounded"
      />
      <span className="flex-1 text-sm">{column}</span>
      {isVisible && (
        <span className="text-xs text-muted-foreground">#{visibleIndex + 1}</span>
      )}
    </div>
  )
}

// Column Settings List Component
function ColumnSettingsList({ columnOrder, setColumnOrder, tableColumns, hiddenColumns, setHiddenColumns }) {
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  const handleDragEnd = (event) => {
    const { active, over } = event

    if (active.id !== over.id) {
      setColumnOrder((items) => {
        const oldIndex = items.indexOf(active.id)
        const newIndex = items.indexOf(over.id)
        return arrayMove(items, oldIndex, newIndex)
      })
    }
  }

  const handleToggleVisibility = (colIndex, checked) => {
    const newHidden = new Set(hiddenColumns)
    if (checked) {
      newHidden.delete(colIndex)
    } else {
      newHidden.add(colIndex)
    }
    setHiddenColumns(newHidden)
  }

  const visibleColumns = columnOrder.filter(idx => !hiddenColumns.has(idx))

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragEnd={handleDragEnd}
    >
      <SortableContext
        items={columnOrder}
        strategy={verticalListSortingStrategy}
      >
        <div className="space-y-2 max-h-60 overflow-auto">
          {columnOrder.map((colIndex) => {
            const column = tableColumns?.[colIndex]
            if (!column) return null
            
            const isVisible = !hiddenColumns.has(colIndex)
            const visibleIndex = visibleColumns.indexOf(colIndex)
            
            return (
              <SortableColumnItem
                key={colIndex}
                id={colIndex}
                column={column}
                index={colIndex}
                isVisible={isVisible}
                onToggleVisibility={handleToggleVisibility}
                visibleIndex={visibleIndex}
              />
            )
          })}
        </div>
      </SortableContext>
    </DndContext>
  )
}

// Column Filter Component
function ColumnFilter({ filter, index, columns, onUpdate, onRemove }) {
  const operators = [
    { value: 'contains', label: 'Contains' },
    { value: 'equals', label: 'Equals' },
    { value: 'startsWith', label: 'Starts with' },
    { value: 'endsWith', label: 'Ends with' },
    { value: 'notEquals', label: 'Not equals' },
    { value: 'notContains', label: 'Not contains' },
    { value: 'greaterThan', label: 'Greater than' },
    { value: 'lessThan', label: 'Less than' },
  ]

  return (
    <div className="flex items-center gap-2 p-3 bg-white rounded border">
      <select
        value={filter.column || ''}
        onChange={(e) => onUpdate({ ...filter, column: e.target.value })}
        className="flex-1 h-8 rounded border border-input bg-transparent px-2 text-sm"
      >
        <option value="">Select column...</option>
        {columns.map((col, idx) => (
          <option key={idx} value={col}>{col}</option>
        ))}
      </select>
      
      <select
        value={filter.operator || 'contains'}
        onChange={(e) => onUpdate({ ...filter, operator: e.target.value })}
        className="w-32 h-8 rounded border border-input bg-transparent px-2 text-sm"
      >
        {operators.map(op => (
          <option key={op.value} value={op.value}>{op.label}</option>
        ))}
      </select>
      
      <Input
        value={filter.value || ''}
        onChange={(e) => onUpdate({ ...filter, value: e.target.value })}
        placeholder="Filter value..."
        className="flex-1 h-8"
      />
      
      <div className="flex items-center gap-1">
        <label className="flex items-center gap-1 text-xs">
          <input
            type="checkbox"
            checked={filter.caseSensitive || false}
            onChange={(e) => onUpdate({ ...filter, caseSensitive: e.target.checked })}
            className="rounded"
          />
          Aa
        </label>
      </div>
      
      <Button
        variant="ghost"
        size="sm"
        onClick={onRemove}
        title="Remove filter"
        className="h-8 w-8 p-0"
      >
        <X className="w-4 h-4" />
      </Button>
    </div>
  )
}

// Data Export Options Component
function DataExportOptions({ selectedFile, tableData, filteredRows, displayColumns, columnFilters, allLoadedRows }) {
  const [exportType, setExportType] = useState('filtered') // 'all' or 'filtered'
  const [exportFormat, setExportFormat] = useState('csv') // 'csv' or 'json'
  const [includeHeaders, setIncludeHeaders] = useState(true)
  const [visibleColumnsOnly, setVisibleColumnsOnly] = useState(true)

  const exportData = () => {
    if (!selectedFile || !tableData.columns) return

    // Determine which data to export
    const dataToExport = exportType === 'filtered' ? filteredRows : (allLoadedRows.length > 0 ? allLoadedRows : tableData.rows || [])
    
    // Determine which columns to include
    const columnsToExport = visibleColumnsOnly ? displayColumns : tableData.columns.map((col, idx) => ({ name: col, index: idx }))
    
    if (dataToExport.length === 0) {
      alert('No data to export')
      return
    }

    // Prepare the data
    const exportRows = dataToExport.map(row => {
      const exportRow = {}
      columnsToExport.forEach(col => {
        const columnName = typeof col === 'object' ? col.name : col
        const columnIndex = typeof col === 'object' ? col.index : tableData.columns.indexOf(col)
        exportRow[columnName] = row[columnIndex] || ''
      })
      return exportRow
    })

    // Export based on format
    if (exportFormat === 'csv') {
      exportToCSV(exportRows, columnsToExport)
    } else {
      exportToJSON(exportRows, columnsToExport)
    }
  }

  const exportToCSV = (data, columns) => {
    const columnNames = columns.map(col => typeof col === 'object' ? col.name : col)
    
    let csvContent = ''
    
    // Add headers if requested
    if (includeHeaders) {
      csvContent += columnNames.map(name => `"${name}"`).join(',') + '\n'
    }
    
    // Add data rows
    data.forEach(row => {
      const csvRow = columnNames.map(colName => {
        const value = row[colName] || ''
        // Escape quotes and wrap in quotes if contains comma, quote, or newline
        const stringValue = String(value)
        if (stringValue.includes(',') || stringValue.includes('"') || stringValue.includes('\n')) {
          return `"${stringValue.replace(/"/g, '""')}"`
        }
        return stringValue
      }).join(',')
      csvContent += csvRow + '\n'
    })

    // Create and download file
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    
    const filterSuffix = exportType === 'filtered' && (columnFilters.length > 0) ? '_filtered' : ''
    const columnSuffix = visibleColumnsOnly ? '_visible_columns' : '_all_columns'
    a.download = `${selectedFile.replace('.DBF', '')}${filterSuffix}${columnSuffix}_${new Date().toISOString().split('T')[0]}.csv`
    
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const exportToJSON = (data, columns) => {
    const exportObject = {
      file: selectedFile,
      exportDate: new Date().toISOString(),
      exportType: exportType,
      includeHeaders: includeHeaders,
      visibleColumnsOnly: visibleColumnsOnly,
      appliedFilters: columnFilters.length > 0 ? columnFilters : null,
      totalRecords: data.length,
      columns: columns.map(col => typeof col === 'object' ? col.name : col),
      data: data
    }

    const blob = new Blob([JSON.stringify(exportObject, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    
    const filterSuffix = exportType === 'filtered' && (columnFilters.length > 0) ? '_filtered' : ''
    const columnSuffix = visibleColumnsOnly ? '_visible_columns' : '_all_columns'
    a.download = `${selectedFile.replace('.DBF', '')}${filterSuffix}${columnSuffix}_${new Date().toISOString().split('T')[0]}.json`
    
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const getDataPreview = () => {
    const dataToPreview = exportType === 'filtered' ? filteredRows : (allLoadedRows.length > 0 ? allLoadedRows : tableData.rows || [])
    const columnsToPreview = visibleColumnsOnly ? displayColumns : tableData.columns.map((col, idx) => ({ name: col, index: idx }))
    
    return {
      recordCount: dataToPreview.length,
      columnCount: columnsToPreview.length,
      hasFilters: columnFilters.length > 0,
      isFiltered: exportType === 'filtered' && columnFilters.length > 0
    }
  }

  const preview = getDataPreview()

  return (
    <div className="space-y-4">
      {/* Export Type Selection */}
      <div>
        <label className="block text-sm font-medium mb-2">Data to Export:</label>
        <div className="space-y-2">
          <label className="flex items-center gap-2">
            <input
              type="radio"
              value="filtered"
              checked={exportType === 'filtered'}
              onChange={(e) => setExportType(e.target.value)}
              className="rounded"
            />
            <span className="text-sm">
              Current View (as filtered) - {preview.recordCount.toLocaleString()} records
              {preview.hasFilters && <span className="text-blue-600 ml-1">({columnFilters.length} filters applied)</span>}
            </span>
          </label>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              value="all"
              checked={exportType === 'all'}
              onChange={(e) => setExportType(e.target.value)}
              className="rounded"
            />
            <span className="text-sm">All Loaded Data - {(allLoadedRows.length || tableData.rows?.length || 0).toLocaleString()} records</span>
          </label>
        </div>
      </div>

      {/* Column Selection */}
      <div>
        <label className="block text-sm font-medium mb-2">Columns to Include:</label>
        <div className="space-y-2">
          <label className="flex items-center gap-2">
            <input
              type="radio"
              value="visible"
              checked={visibleColumnsOnly}
              onChange={(e) => setVisibleColumnsOnly(e.target.checked)}
              className="rounded"
            />
            <span className="text-sm">
              Visible Columns Only - {displayColumns.length} columns
              {displayColumns.length !== tableData.columns?.length && (
                <span className="text-gray-500 ml-1">({(tableData.columns?.length || 0) - displayColumns.length} hidden)</span>
              )}
            </span>
          </label>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              value="all"
              checked={!visibleColumnsOnly}
              onChange={(e) => setVisibleColumnsOnly(!e.target.checked)}
              className="rounded"
            />
            <span className="text-sm">All Columns - {tableData.columns?.length || 0} columns</span>
          </label>
        </div>
      </div>

      {/* Format Selection */}
      <div>
        <label className="block text-sm font-medium mb-2">Export Format:</label>
        <div className="flex gap-4">
          <label className="flex items-center gap-2">
            <input
              type="radio"
              value="csv"
              checked={exportFormat === 'csv'}
              onChange={(e) => setExportFormat(e.target.value)}
              className="rounded"
            />
            <span className="text-sm">CSV (Excel compatible)</span>
          </label>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              value="json"
              checked={exportFormat === 'json'}
              onChange={(e) => setExportFormat(e.target.value)}
              className="rounded"
            />
            <span className="text-sm">JSON (with metadata)</span>
          </label>
        </div>
      </div>

      {/* Additional Options */}
      <div>
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={includeHeaders}
            onChange={(e) => setIncludeHeaders(e.target.checked)}
            className="rounded"
          />
          <span className="text-sm">Include column headers</span>
        </label>
      </div>

      {/* Export Preview */}
      <div className="p-3 bg-gray-50 rounded border text-sm">
        <div className="font-medium mb-1">Export Preview:</div>
        <div className="text-gray-600 space-y-1">
          <div>• {preview.recordCount.toLocaleString()} records</div>
          <div>• {preview.columnCount} columns</div>
          <div>• Format: {exportFormat.toUpperCase()}</div>
          {preview.isFiltered && <div className="text-blue-600">• Filtered data (filters applied)</div>}
        </div>
      </div>

      {/* Export Button */}
      <div className="flex justify-end">
        <Button
          onClick={exportData}
          disabled={preview.recordCount === 0}
          className="bg-green-600 hover:bg-green-700"
        >
          <Download className="w-4 h-4 mr-2" />
          Export {exportFormat.toUpperCase()}
        </Button>
      </div>
    </div>
  )
}