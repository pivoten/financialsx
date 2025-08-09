
import { useState, useEffect, useCallback, useRef } from 'react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table'
import { Select } from './ui/select'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './ui/dialog'
import { FileText, Search, Edit, Save, X, Plus, ChevronUp, ChevronDown, ChevronsUpDown, Settings, Database, GripVertical, Download, Upload, Filter, FilterX } from 'lucide-react'
import { GetDBFFiles, GetDBFTableDataPaged, SearchDBFTable, UpdateDBFRecord } from '../../wailsjs/go/main/App'
import logger from '../services/logger'
import { DndContext, closestCenter, KeyboardSensor, PointerSensor, useSensor, useSensors } from '@dnd-kit/core'
import { arrayMove, SortableContext, sortableKeyboardCoordinates, verticalListSortingStrategy, useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import type { User } from '../types'
import type { ChangeEvent } from 'react'

// Global API call tracker to prevent rapid calls across all instances
let globalApiCallInProgress = false
let lastGlobalApiCall = 0

interface DBFExplorerProps {
  currentUser: User
}

interface DBFTableData {
  columns: string[]
  rows: any[][]
  stats?: {
    totalRecords?: number
    deletedRecords?: number
    loadedRecords?: number
    totalMatching?: number
    hasMoreRecords?: boolean
    searchTerm?: string
  }
}

interface EditingCell {
  row: number
  col: number
}

interface ColumnFilter {
  column: string
  operator: string
  value: string
  caseSensitive: boolean
  logicalOperator?: 'AND' | 'OR' | null
}

interface DisplayColumn {
  name: string
  index: number
}

interface SelectedRecord {
  column: string
  value: any
  index: number
}

export function DBFExplorer({ currentUser }: DBFExplorerProps) {
  const [dbfFiles, setDbfFiles] = useState<string[]>([])
  const [selectedFile, setSelectedFile] = useState<string>('')
  const [tableData, setTableData] = useState<DBFTableData>({ columns: [], rows: [], stats: {} })
  const [loading, setLoading] = useState<boolean>(false)
  const [editingCell, setEditingCell] = useState<EditingCell | null>(null)
  const [editValue, setEditValue] = useState<string>('')
  const [searchTerm, setSearchTerm] = useState<string>('')
  const [currentCompany, setCurrentCompany] = useState<string>('')
  const [sortColumn, setSortColumn] = useState<number | null>(null)
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc')
  const [isServerSearching, setIsServerSearching] = useState<boolean>(false)
  const [currentPage, setCurrentPage] = useState<number>(0)
  const [pageSize] = useState<number>(100) // Load 100 records at a time
  const [hasMoreData, setHasMoreData] = useState<boolean>(false)
  const [isLoadingMore, setIsLoadingMore] = useState<boolean>(false)
  const [allLoadedRows, setAllLoadedRows] = useState<any[][]>([])
  const [columnOrder, setColumnOrder] = useState<number[]>([])
  const [hiddenColumns, setHiddenColumns] = useState<Set<number>>(new Set())
  const [showColumnSettings, setShowColumnSettings] = useState<boolean>(false)
  const [showColumnFilters, setShowColumnFilters] = useState<boolean>(false)
  const [columnFilters, setColumnFilters] = useState<ColumnFilter[]>([])
  const [showDataExport, setShowDataExport] = useState<boolean>(false)
  const [isEditMode, setIsEditMode] = useState<boolean>(false)
  const [selectedRecord, setSelectedRecord] = useState<SelectedRecord[] | null>(null)
  const [showRecordModal, setShowRecordModal] = useState<boolean>(false)
  const lastClickTime = useRef<number>(0)
  const clickTimeout = useRef<NodeJS.Timeout | null>(null)
  const searchTimeout = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => { loadCurrentCompanyFiles() }, [])

  useEffect(() => {
    if (selectedFile && columnOrder.length > 0) {
      const prefs = { columnOrder, hiddenColumns: Array.from(hiddenColumns) }
      localStorage.setItem(`dbf_columns_${selectedFile}`, JSON.stringify(prefs))
    }
  }, [columnOrder, hiddenColumns, selectedFile])

  const loadCurrentCompanyFiles = async () => {
    setLoading(true)
    try {
      const companyName = localStorage.getItem('company_name')
      const companyPath = localStorage.getItem('company_path')
      if (!companyName && !companyPath) return
      const dataPath = companyPath || companyName
      setCurrentCompany(dataPath)
      const result = await GetDBFFiles(dataPath)
      setDbfFiles(result || [])
    } catch (error) {
      logger.error('Failed to load DBF files', { error: error.message })
      setDbfFiles([])
    } finally {
      setLoading(false)
    }
  }

  const loadTableData = async (fileName: string, resetPagination = true) => {
    const now = Date.now()
    if (globalApiCallInProgress) { setLoading(false); return }
    if (now - lastGlobalApiCall < 1000) { setLoading(false); return }
    globalApiCallInProgress = true
    lastGlobalApiCall = now
    setLoading(true)
    if (resetPagination) { setCurrentPage(0); setAllLoadedRows([]) }

    try {
      const companyPath = localStorage.getItem('company_path')
      const companyName = localStorage.getItem('company_name')
      let companyToUse = currentCompany || companyPath || companyName || 'cantrellenergy'
      if (!companyToUse) { setTableData({ columns: [], rows: [] }); return }
      const offset = resetPagination ? 0 : currentPage * pageSize
      const sortCol = sortColumn !== null ? tableData.columns?.[sortColumn] : ''
      const result = await GetDBFTableDataPaged(companyToUse, fileName, offset, pageSize, sortCol, sortDirection)
      const safeResult = { columns: result?.columns || [], rows: result?.rows || [], stats: result?.stats || {} }

      if (resetPagination) {
        setTableData(safeResult)
        setAllLoadedRows(safeResult.rows || [])
        if (safeResult.columns.length > 0) {
          const savedPrefs = localStorage.getItem(`dbf_columns_${fileName}`)
          if (savedPrefs) {
            try {
              const prefs = JSON.parse(savedPrefs)
              const validOrder = prefs.columnOrder.filter((idx: number) => idx < safeResult.columns.length)
              setColumnOrder(validOrder.length > 0 ? validOrder : safeResult.columns.map((_: any, index: number) => index))
              setHiddenColumns(new Set((prefs.hiddenColumns || []).filter((idx: number) => idx < safeResult.columns.length)))
            } catch {
              setColumnOrder(safeResult.columns.map((_: any, index: number) => index))
            }
          } else {
            setColumnOrder(safeResult.columns.map((_: any, index: number) => index))
          }
        }
      } else {
        const newRows = [...allLoadedRows, ...(safeResult.rows || [])]
        setAllLoadedRows(newRows)
        setTableData({ ...safeResult, rows: newRows })
      }

      const totalMatching = result?.stats?.totalMatching || 0
      const currentlyLoaded = resetPagination ? (safeResult.rows?.length || 0) : allLoadedRows.length + (safeResult.rows?.length || 0)
      setHasMoreData(currentlyLoaded < totalMatching)
      if (!resetPagination) setCurrentPage(prev => prev + 1)
    } catch (error) {
      logger.error('Failed to load table data', { error: error.message })
      setTableData({ columns: [], rows: [] })
      setAllLoadedRows([])
    } finally {
      setTimeout(() => { globalApiCallInProgress = false; setLoading(false) }, 500)
    }
  }

  const handleFileSelect = useCallback((fileName: string) => {
    const now = Date.now()
    if (globalApiCallInProgress || loading) return
    if (selectedFile === fileName) return
    if (now - lastClickTime.current < 1000) return
    lastClickTime.current = now
    if (clickTimeout.current) clearTimeout(clickTimeout.current)
    setSelectedFile(fileName)
    setCurrentPage(0); setAllLoadedRows([]); setSortColumn(null); setSortDirection('asc'); setColumnOrder([]); setHiddenColumns(new Set()); setColumnFilters([])
    loadTableData(fileName, true)
  }, [loading, selectedFile])

  const handleCellEdit = (rowIndex: number, columnIndex: number) => {
    const currentValue = tableData.rows[rowIndex]?.[columnIndex] || ''
    setEditingCell({ row: rowIndex, col: columnIndex })
    setEditValue(currentValue)
  }

  const handleSaveEdit = async () => {
    if (!editingCell) return
    try {
      const newRows = [...tableData.rows]
      newRows[editingCell.row][editingCell.col] = editValue
      setTableData({ ...tableData, rows: newRows })
      await UpdateDBFRecord(currentCompany, selectedFile, editingCell.row, editingCell.col, editValue)
      setEditingCell(null); setEditValue('')
    } catch (error) {
      logger.error('Failed to save changes', { error: error.message })
      loadTableData(selectedFile)
    }
  }

  const handleCancelEdit = () => { setEditingCell(null); setEditValue('') }

  const canEdit = () => currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')

  const handleRowClick = (rowIndex: number) => {
    if (!isEditMode) {
      const record = tableData.rows[rowIndex]
      const recordWithColumns = tableData.columns.map((col, index) => ({ column: col, value: record[index], index }))
      setSelectedRecord(recordWithColumns)
      setShowRecordModal(true)
    }
  }

  const toggleEditMode = () => {
    setIsEditMode(!isEditMode)
    if (!isEditMode) { setEditingCell(null); setEditValue('') }
  }

  const formatLogicalValue = (value: any): string => {
    if (value === null || value === undefined || value === '') return ''
    if (typeof value === 'boolean') return value ? 'True' : 'False'
    if (typeof value === 'string') {
      const lowerValue = value.toLowerCase()
      if (lowerValue === 't' || lowerValue === '.t.' || lowerValue === 'true') return 'True'
      if (lowerValue === 'f' || lowerValue === '.f.' || lowerValue === 'false') return 'False'
      if (lowerValue === 'null' || lowerValue === 'nil' || lowerValue.trim() === '') return ''
    }
    return value
  }

  const performServerSearch = async (search: string) => {
    if (!selectedFile || !currentCompany) return
    setIsServerSearching(true)
    try {
      const companyToUse = currentCompany || localStorage.getItem('company_name') || 'cantrellenergy'
      if (search.trim() === '') {
        setCurrentPage(0); setAllLoadedRows([])
        loadTableData(selectedFile, true)
      } else {
        const result = await SearchDBFTable(companyToUse, selectedFile, search)
        const safeResult = { columns: result?.columns || [], rows: result?.rows || [], stats: result?.stats || {} }
        setTableData(safeResult)
        setAllLoadedRows(safeResult.rows || [])
        setHasMoreData(false)
      }
    } catch (error) {
      logger.error('Search failed', { error: error.message })
      setTableData({ columns: [], rows: [], stats: {} })
      setAllLoadedRows([])
    } finally {
      setIsServerSearching(false)
    }
  }

  const handleSearchChange = (value: string) => {
    setSearchTerm(value)
    if (searchTimeout.current) clearTimeout(searchTimeout.current)
    searchTimeout.current = setTimeout(() => { performServerSearch(value) }, 500)
  }

  const handleSort = (columnIndex: number) => {
    const newDirection = sortColumn === columnIndex ? (sortDirection === 'asc' ? 'desc' : 'asc') : 'asc'
    setSortColumn(columnIndex)
    setSortDirection(newDirection)
    if (selectedFile) loadTableData(selectedFile, true)
  }

  const applyColumnFilters = (rows: any[][]) => {
    if (columnFilters.length === 0) return rows
    return rows.filter(row => {
      const evaluateFilter = (filter: ColumnFilter) => {
        if (!filter.column || !filter.value) return true
        const columnIndex = tableData.columns?.indexOf(filter.column)
        if (columnIndex === -1) return true
        const cellValue = row[columnIndex]
        if (cellValue == null) return false
        let cellStr = String(cellValue)
        let filterValue = String(filter.value)
        if (!filter.caseSensitive) { cellStr = cellStr.toLowerCase(); filterValue = filterValue.toLowerCase() }
        switch (filter.operator) {
          case 'equals': return cellStr === filterValue
          case 'contains': return cellStr.includes(filterValue)
          case 'startsWith': return cellStr.startsWith(filterValue)
          case 'endsWith': return cellStr.endsWith(filterValue)
          case 'notEquals': return cellStr !== filterValue
          case 'notContains': return !cellStr.includes(filterValue)
          case 'greaterThan': { const numA = parseFloat(cellStr); const numB = parseFloat(filterValue); return !isNaN(numA) && !isNaN(numB) && numA > numB }
          case 'lessThan': { const numC = parseFloat(cellStr); const numD = parseFloat(filterValue); return !isNaN(numC) && !isNaN(numD) && numC < numD }
          default: return true
        }
      }
      if (columnFilters.length === 1) return evaluateFilter(columnFilters[0])
      let result = evaluateFilter(columnFilters[0])
      for (let i = 1; i < columnFilters.length; i++) {
        const filter = columnFilters[i]
        const filterResult = evaluateFilter(filter)
        if (filter.logicalOperator === 'OR') result = result || filterResult
        else result = result && filterResult
      }
      return result
    })
  }

  const filteredRows = applyColumnFilters(tableData.rows || [])

  const displayColumns = columnOrder.length > 0
    ? columnOrder.filter(idx => !hiddenColumns.has(idx)).map(idx => ({ name: tableData.columns[idx], index: idx }))
    : (tableData.columns?.map((col, idx) => ({ name: col, index: idx })).filter(col => !hiddenColumns.has(col.index)) || [])

  const loadMoreData = async () => {
    if (isLoadingMore || !hasMoreData || !selectedFile) return
    setIsLoadingMore(true)
    try { await loadTableData(selectedFile, false) } catch (e) { logger.error('Failed to load more data', { error: e.message }) } finally { setIsLoadingMore(false) }
  }

  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const { scrollTop, scrollHeight, clientHeight } = e.target as HTMLDivElement
    if (scrollHeight - scrollTop - clientHeight < 200 && hasMoreData && !isLoadingMore) { loadMoreData() }
  }

  const exportColumnSettings = () => {
    if (!selectedFile || !tableData.columns) return
    const settings = {
      file: selectedFile,
      exportDate: new Date().toISOString(),
      columns: tableData.columns.map((col, idx) => ({ name: col, originalIndex: idx, visible: !hiddenColumns.has(idx) })),
      columnOrder,
      hiddenColumns: Array.from(hiddenColumns),
      columnFilters,
      version: '1.0'
    }
    const blob = new Blob([JSON.stringify(settings, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${selectedFile}_columns_${new Date().toISOString().split('T')[0]}.json`
    document.body.appendChild(a); a.click(); document.body.removeChild(a); URL.revokeObjectURL(url)
  }

  const handleImportSettings = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = (event: ProgressEvent<FileReader>) => {
      try {
        const settings = JSON.parse(event.target?.result as string)
        if (!settings.columnOrder || !Array.isArray(settings.columnOrder)) throw new Error('Invalid settings file: missing columnOrder')
        const currentColumns = tableData.columns || []
        const importedColumns = settings.columns || []
        const columnMap = new Map()
        currentColumns.forEach((col, idx) => columnMap.set(col, idx))
        const newOrder: number[] = []
        const newHidden = new Set<number>()
        importedColumns.forEach((importedCol: any) => {
          const currentIdx = columnMap.get(importedCol.name)
          if (currentIdx !== undefined) {
            newOrder.push(currentIdx)
            if (!importedCol.visible) newHidden.add(currentIdx)
          }
        })
        currentColumns.forEach((col, idx) => { if (!newOrder.includes(idx)) newOrder.push(idx) })
        setColumnOrder(newOrder)
        setHiddenColumns(newHidden)
        if (settings.columnFilters && Array.isArray(settings.columnFilters)) {
          const validFilters = settings.columnFilters.filter((filter: any) => filter.column && currentColumns.includes(filter.column))
          setColumnFilters(validFilters)
        }
      } catch (error) {
        logger.error('Failed to import settings', { error: error.message })
        alert('Failed to import settings. Please check the file format.')
      }
    }
    reader.readAsText(file)
    e.target.value = ''
  }

  return (
    <div className="space-y-6">
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
              onChange={(e: ChangeEvent<HTMLSelectElement>) => { e.preventDefault(); e.stopPropagation(); handleFileSelect(e.target.value) }}
              disabled={loading}
              className="w-full"
            >
              <option value="">{dbfFiles.length === 0 ? 'No DBF files found' : 'Choose a DBF file...'}</option>
              {dbfFiles.map((file) => (
                <option key={file} value={file}>{file}</option>
              ))}
            </Select>
          )}
        </div>
        {currentCompany && (<div className="text-sm text-muted-foreground">Company: {currentCompany}</div>)}
      </div>

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
                            Search matches: {tableData.stats.loadedRecords.toLocaleString()} {tableData.stats.hasMoreRecords && ' (limited to 1000)'}
                          </span>
                          {isServerSearching && <span className="text-gray-500">Searching...</span>}
                        </>
                      )}
                      {columnFilters.length > 0 && (
                        <span className="text-blue-600">
                          Filtered: {filteredRows.length.toLocaleString()} of {(tableData.rows || []).length.toLocaleString()}
                          {columnFilters.length > 1 && (
                            <span className="text-xs ml-1">({columnFilters.some(f => f.logicalOperator === 'OR') ? 'AND/OR' : 'AND'})</span>
                          )}
                        </span>
                      )}
                    </div>
                  )}
                </CardDescription>
              </div>
              <div className="flex items-center gap-2">
                {canEdit() && (
                  <Button variant={isEditMode ? 'default' : 'outline'} size="sm" onClick={toggleEditMode} title={isEditMode ? 'Exit Edit Mode' : 'Enter Edit Mode'}>
                    <Edit className="w-4 h-4" />
                    {isEditMode ? 'Exit Edit' : 'Edit'}
                  </Button>
                )}
                <Button variant="outline" size="sm" onClick={() => setShowColumnSettings(!showColumnSettings)} title="Column Settings">
                  <Settings className="w-4 h-4" />
                </Button>
                <Button variant="outline" size="sm" onClick={() => setShowColumnFilters(!showColumnFilters)} title="Column Filters" className={columnFilters.length > 0 ? 'bg-blue-50 border-blue-300' : ''}>
                  <Filter className="w-4 h-4" />
                  {columnFilters.length > 0 && (<span className="ml-1 text-xs bg-blue-600 text-white rounded-full px-1">{columnFilters.length}</span>)}
                </Button>
                <Button variant="outline" size="sm" onClick={() => setShowDataExport(!showDataExport)} title="Export Data">
                  <Database className="w-4 h-4" />
                </Button>
                <div className="relative">
                  <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input placeholder="Search all records..." value={searchTerm} onChange={(e: ChangeEvent<HTMLInputElement>) => handleSearchChange(e.target.value)} className="pl-8 w-64" disabled={isServerSearching} />
                </div>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {showColumnSettings && (
              <div className="mb-4 p-4 border rounded-lg bg-muted/50">
                <div className="flex justify-between items-center mb-3">
                  <h4 className="font-semibold text-sm">Column Settings</h4>
                  <Button variant="ghost" size="sm" onClick={() => setShowColumnSettings(false)}>
                    <X className="w-4 h-4" />
                  </Button>
                </div>
                <ColumnSettingsList columnOrder={columnOrder} setColumnOrder={setColumnOrder} tableColumns={tableData.columns} hiddenColumns={hiddenColumns} setHiddenColumns={setHiddenColumns} />
                <div className="flex gap-2 mt-3 pt-3 border-t">
                  <Button variant="outline" size="sm" onClick={() => { setColumnOrder(tableData.columns.map((_, i) => i)); setHiddenColumns(new Set()) }}>Reset to Default</Button>
                  <Button variant="outline" size="sm" onClick={() => setHiddenColumns(new Set())}>Show All</Button>
                  <div className="ml-auto flex gap-2">
                    <Button variant="outline" size="sm" onClick={() => exportColumnSettings()} title="Export column settings"><Download className="w-4 h-4" /></Button>
                    <Button variant="outline" size="sm" onClick={() => document.getElementById('import-settings')?.click()} title="Import column settings"><Upload className="w-4 h-4" /></Button>
                    <input id="import-settings" type="file" accept=".json" className="hidden" onChange={handleImportSettings} />
                  </div>
                </div>
              </div>
            )}

            {showColumnFilters && (
              <div className="mb-4 p-4 border rounded-lg bg-blue-50/50">
                <div className="flex justify-between items-center mb-3">
                  <h4 className="font-semibold text-sm flex items-center gap-2">
                    <Filter className="w-4 h-4" />
                    Column Filters
                  </h4>
                  <div className="flex gap-2">
                    {columnFilters.length > 0 && (
                      <Button variant="ghost" size="sm" onClick={() => setColumnFilters([])} title="Clear all filters">
                        <FilterX className="w-4 h-4" />
                      </Button>
                    )}
                    <Button variant="ghost" size="sm" onClick={() => setShowColumnFilters(false)}>
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
                            <button type="button" onClick={() => { const newFilters = [...columnFilters]; newFilters[index] = { ...filter, logicalOperator: 'AND' }; setColumnFilters(newFilters) }} className={`px-2 py-1 text-xs font-medium rounded ${filter.logicalOperator === 'AND' || !filter.logicalOperator ? 'bg-blue-600 text-white' : 'bg-transparent text-gray-600 hover:bg-gray-200'}`}>AND</button>
                            <button type="button" onClick={() => { const newFilters = [...columnFilters]; newFilters[index] = { ...filter, logicalOperator: 'OR' }; setColumnFilters(newFilters) }} className={`px-2 py-1 text-xs font-medium rounded ${filter.logicalOperator === 'OR' ? 'bg-green-600 text-white' : 'bg-transparent text-gray-600 hover:bg-gray-200'}`}>OR</button>
                          </div>
                        </div>
                      )}
                      <ColumnFilter
                        filter={filter}
                        index={index}
                        columns={tableData.columns || []}
                        onUpdate={(updatedFilter: ColumnFilter) => { const newFilters = [...columnFilters]; newFilters[index] = updatedFilter; setColumnFilters(newFilters) }}
                        onRemove={() => { const newFilters = columnFilters.filter((_, i) => i !== index); setColumnFilters(newFilters) }}
                      />
                    </div>
                  ))}

                  <Button variant="outline" size="sm" onClick={() => { setColumnFilters([...columnFilters, { column: tableData.columns?.[0] || '', operator: 'contains', value: '', caseSensitive: false, logicalOperator: columnFilters.length > 0 ? 'AND' : null }]) }} disabled={!tableData.columns || tableData.columns.length === 0}>
                    <Plus className="w-4 h-4 mr-2" />
                    Add Filter
                  </Button>
                </div>
              </div>
            )}

            {showDataExport && (
              <div className="mb-4 p-4 border rounded-lg bg-green-50/50">
                <div className="flex justify-between items-center mb-3">
                  <h4 className="font-semibold text-sm flex items-center gap-2">
                    <Database className="w-4 h-4" />
                    Export Data
                  </h4>
                  <Button variant="ghost" size="sm" onClick={() => setShowDataExport(false)}>
                    <X className="w-4 h-4" />
                  </Button>
                </div>

                <DataExportOptions selectedFile={selectedFile} tableData={tableData} filteredRows={filteredRows} displayColumns={displayColumns} columnFilters={columnFilters} allLoadedRows={allLoadedRows} />
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
                          <TableHead key={col.index} className="cursor-pointer hover:bg-muted/50 select-none" onClick={() => handleSort(col.index)}>
                            <div className="flex items-center gap-1">
                              {col.name}
                              <span className="ml-auto">
                                {sortColumn === col.index ? (sortDirection === 'asc' ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />) : (<ChevronsUpDown className="w-4 h-4 opacity-30" />)}
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
                                    <Input value={editValue} onChange={(e: ChangeEvent<HTMLInputElement>) => setEditValue(e.target.value)} className="h-8" autoFocus onKeyDown={(e) => { if (e.key === 'Enter') handleSaveEdit(); if (e.key === 'Escape') handleCancelEdit() }} />
                                    <Button size="sm" onClick={handleSaveEdit}><Save className="w-3 h-3" /></Button>
                                    <Button size="sm" variant="outline" onClick={handleCancelEdit}><X className="w-3 h-3" /></Button>
                                  </div>
                                ) : (
                                  <div className={`p-1 rounded min-h-6 ${isEditMode ? 'cursor-pointer hover:bg-muted/50' : 'cursor-pointer hover:bg-blue-50'}`} onClick={() => { if (isEditMode) { handleCellEdit(rowIndex, col.index) } else { handleRowClick(rowIndex) } }}>
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

                  {isLoadingMore && (<div className="text-center py-4 text-muted-foreground">Loading more records...</div>)}
                  {!hasMoreData && filteredRows.length > pageSize && (<div className="text-center py-4 text-muted-foreground text-sm">All {filteredRows.length.toLocaleString()} records loaded</div>)}
                </div>
                {filteredRows.length === 0 && searchTerm && (<div className="text-center py-8 text-muted-foreground">No records match your search</div>)}
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                {selectedFile ? 'No data found in this file' : 'Select a DBF file to view data'}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {showRecordModal && selectedRecord && (
        <Dialog open={showRecordModal} onOpenChange={setShowRecordModal}>
          <DialogContent className="max-w-4xl max-h-[80vh] overflow-auto">
            <DialogHeader>
              <DialogTitle>Record Details - {selectedFile}</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div className="text-sm text-muted-foreground">Complete record data for row {tableData.rows?.indexOf(selectedRecord.map(r => r.value)) + 1 || 'N/A'}</div>
              <div className="grid gap-3">
                {selectedRecord.map((field, index) => (
                  <div key={index} className="grid grid-cols-3 gap-4 p-3 border rounded">
                    <div className="font-medium text-sm text-muted-foreground">{field.column}</div>
                    <div className="col-span-2 break-all">
                      <div className="p-2 bg-muted/30 rounded text-sm">{formatLogicalValue(field.value) || <span className="text-muted-foreground italic">Empty</span>}</div>
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

interface SortableColumnItemProps {
  id: number
  column: string
  index: number
  isVisible: boolean
  onToggleVisibility: (index: number, checked: boolean) => void
  visibleIndex: number
}

function SortableColumnItem({ id, column, index, isVisible, onToggleVisibility, visibleIndex }: SortableColumnItemProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id })
  const style = { transform: CSS.Transform.toString(transform), transition, opacity: isDragging ? 0.5 : 1 }
  return (
    <div ref={setNodeRef} style={style} className={`flex items-center gap-2 p-2 rounded hover:bg-background ${isDragging ? 'shadow-lg bg-background' : ''}`}>
      <div {...attributes} {...listeners} className="cursor-grab hover:cursor-grabbing">
        <GripVertical className="w-4 h-4 text-muted-foreground" />
      </div>
      <input type="checkbox" checked={isVisible} onChange={(e: ChangeEvent<HTMLInputElement>) => onToggleVisibility(index, e.target.checked)} className="rounded" />
      <span className="flex-1 text-sm">{column}</span>
      {isVisible && (<span className="text-xs text-muted-foreground">#{visibleIndex + 1}</span>)}
    </div>
  )
}

interface ColumnSettingsListProps {
  columnOrder: number[]
  setColumnOrder: (order: number[]) => void
  tableColumns: string[]
  hiddenColumns: Set<number>
  setHiddenColumns: (hidden: Set<number>) => void
}

function ColumnSettingsList({ columnOrder, setColumnOrder, tableColumns, hiddenColumns, setHiddenColumns }: ColumnSettingsListProps) {
  const sensors = useSensors(useSensor(PointerSensor), useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }))

  const handleDragEnd = (event: any) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    const oldIndex = columnOrder.indexOf(active.id)
    const newIndex = columnOrder.indexOf(over.id)
    setColumnOrder(arrayMove(columnOrder, oldIndex, newIndex))
  }

  const handleToggleVisibility = (colIndex: number, checked: boolean) => {
    const newHidden = new Set(hiddenColumns)
    if (checked) newHidden.delete(colIndex); else newHidden.add(colIndex)
    setHiddenColumns(newHidden)
  }

  const visibleColumns = columnOrder.filter(idx => !hiddenColumns.has(idx))

  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
      <SortableContext items={columnOrder} strategy={verticalListSortingStrategy}>
        <div className="space-y-2 max-h-60 overflow-auto">
          {columnOrder.map((colIndex) => {
            const column = tableColumns?.[colIndex]
            if (!column) return null
            const isVisible = !hiddenColumns.has(colIndex)
            const visibleIndex = visibleColumns.indexOf(colIndex)
            return (
              <SortableColumnItem key={colIndex} id={colIndex} column={column} index={colIndex} isVisible={isVisible} onToggleVisibility={handleToggleVisibility} visibleIndex={visibleIndex} />
            )
          })}
        </div>
      </SortableContext>
    </DndContext>
  )
}

interface ColumnFilterProps {
  filter: ColumnFilter
  index: number
  columns: string[]
  onUpdate: (filter: ColumnFilter) => void
  onRemove: () => void
}

function ColumnFilter({ filter, index, columns, onUpdate, onRemove }: ColumnFilterProps) {
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
      <select value={filter.column || ''} onChange={(e: ChangeEvent<HTMLSelectElement>) => onUpdate({ ...filter, column: e.target.value })} className="flex-1 h-8 rounded border border-input bg-transparent px-2 text-sm">
        <option value="">Select column...</option>
        {columns.map((col, idx) => (<option key={idx} value={col}>{col}</option>))}
      </select>
      <select value={filter.operator || 'contains'} onChange={(e: ChangeEvent<HTMLSelectElement>) => onUpdate({ ...filter, operator: e.target.value })} className="w-32 h-8 rounded border border-input bg-transparent px-2 text-sm">
        {operators.map(op => (<option key={op.value} value={op.value}>{op.label}</option>))}
      </select>
      <Input value={filter.value || ''} onChange={(e: ChangeEvent<HTMLInputElement>) => onUpdate({ ...filter, value: e.target.value })} placeholder="Filter value..." className="flex-1 h-8" />
      <div className="flex items-center gap-1">
        <label className="flex items-center gap-1 text-xs">
          <input type="checkbox" checked={filter.caseSensitive || false} onChange={(e: ChangeEvent<HTMLInputElement>) => onUpdate({ ...filter, caseSensitive: e.target.checked })} className="rounded" />
          Aa
        </label>
      </div>
      <Button variant="ghost" size="sm" onClick={onRemove} title="Remove filter" className="h-8 w-8 p-0">
        <X className="w-4 h-4" />
      </Button>
    </div>
  )
}

interface DataExportOptionsProps {
  selectedFile: string
  tableData: DBFTableData
  filteredRows: any[][]
  displayColumns: DisplayColumn[]
  columnFilters: ColumnFilter[]
  allLoadedRows: any[][]
}

function DataExportOptions({ selectedFile, tableData, filteredRows, displayColumns, columnFilters, allLoadedRows }: DataExportOptionsProps) {
  const [exportType, setExportType] = useState<'filtered' | 'all'>('filtered')
  const [exportFormat, setExportFormat] = useState<'csv' | 'json'>('csv')
  const [includeHeaders, setIncludeHeaders] = useState<boolean>(true)
  const [visibleColumnsOnly, setVisibleColumnsOnly] = useState<boolean>(true)

  const exportData = () => {
    if (!selectedFile || !tableData.columns) return
    const dataToExport = exportType === 'filtered' ? filteredRows : (allLoadedRows.length > 0 ? allLoadedRows : tableData.rows || [])
    const columnsToExport = visibleColumnsOnly ? displayColumns : tableData.columns.map((col, idx) => ({ name: col, index: idx }))
    if (dataToExport.length === 0) { alert('No data to export'); return }
    const exportRows = dataToExport.map(row => {
      const exportRow: Record<string, any> = {}
      columnsToExport.forEach(col => {
        const columnName = typeof col === 'object' ? col.name : col
        const columnIndex = typeof col === 'object' ? col.index : (tableData.columns as string[]).indexOf(col as string)
        exportRow[columnName] = row[columnIndex] || ''
      })
      return exportRow
    })
    if (exportFormat === 'csv') exportToCSV(exportRows, columnsToExport)
    else exportToJSON(exportRows, columnsToExport)
  }

  const exportToCSV = (data: any[], columns: (DisplayColumn | string)[]) => {
    const columnNames = columns.map(col => typeof col === 'object' ? col.name : col)
    let csvContent = ''
    if (includeHeaders) csvContent += columnNames.map(name => `"${name}"`).join(',') + '\n'
    data.forEach(row => {
      const csvRow = columnNames.map(colName => {
        const value = row[colName] || ''
        const stringValue = String(value)
        if (stringValue.includes(',') || stringValue.includes('"') || stringValue.includes('\n')) return `"${stringValue.replace(/"/g, '""')}"`
        return stringValue
      }).join(',')
      csvContent += csvRow + '\n'
    })
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    const filterSuffix = exportType === 'filtered' && (columnFilters.length > 0) ? '_filtered' : ''
    const columnSuffix = visibleColumnsOnly ? '_visible_columns' : '_all_columns'
    a.download = `${selectedFile.replace('.DBF', '')}${filterSuffix}${columnSuffix}_${new Date().toISOString().split('T')[0]}.csv`
    document.body.appendChild(a); a.click(); document.body.removeChild(a); URL.revokeObjectURL(url)
  }

  const exportToJSON = (data: any[], columns: (DisplayColumn | string)[]) => {
    const exportObject = {
      file: selectedFile,
      exportDate: new Date().toISOString(),
      exportType,
      includeHeaders,
      visibleColumnsOnly,
      appliedFilters: columnFilters.length > 0 ? columnFilters : null,
      totalRecords: data.length,
      columns: columns.map(col => typeof col === 'object' ? col.name : col),
      data,
    }
    const blob = new Blob([JSON.stringify(exportObject, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    const filterSuffix = exportType === 'filtered' && (columnFilters.length > 0) ? '_filtered' : ''
    const columnSuffix = visibleColumnsOnly ? '_visible_columns' : '_all_columns'
    a.download = `${selectedFile.replace('.DBF', '')}${filterSuffix}${columnSuffix}_${new Date().toISOString().split('T')[0]}.json`
    document.body.appendChild(a); a.click(); document.body.removeChild(a); URL.revokeObjectURL(url)
  }

  const getDataPreview = () => {
    const dataToPreview = exportType === 'filtered' ? filteredRows : (allLoadedRows.length > 0 ? allLoadedRows : tableData.rows || [])
    const columnsToPreview = visibleColumnsOnly ? displayColumns : tableData.columns.map((col, idx) => ({ name: col, index: idx }))
    return { recordCount: dataToPreview.length, columnCount: columnsToPreview.length, hasFilters: columnFilters.length > 0, isFiltered: exportType === 'filtered' && columnFilters.length > 0 }
  }

  const preview = getDataPreview()

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium mb-2">Data to Export:</label>
        <div className="space-y-2">
          <label className="flex items-center gap-2">
            <input type="radio" value="filtered" checked={exportType === 'filtered'} onChange={(e: ChangeEvent<HTMLInputElement>) => setExportType(e.target.value as 'filtered')} className="rounded" />
            <span className="text-sm">Current View (as filtered) - {preview.recordCount.toLocaleString()} records {preview.hasFilters && <span className="text-blue-600 ml-1">({columnFilters.length} filters applied)</span>}</span>
          </label>
          <label className="flex items-center gap-2">
            <input type="radio" value="all" checked={exportType === 'all'} onChange={(e: ChangeEvent<HTMLInputElement>) => setExportType(e.target.value as 'all')} className="rounded" />
            <span className="text-sm">All Loaded Data - {(allLoadedRows.length || tableData.rows?.length || 0).toLocaleString()} records</span>
          </label>
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium mb-2">Columns to Include:</label>
        <div className="space-y-2">
          <label className="flex items-center gap-2">
            <input type="radio" value="visible" checked={visibleColumnsOnly} onChange={(e: ChangeEvent<HTMLInputElement>) => setVisibleColumnsOnly(e.target.checked)} className="rounded" />
            <span className="text-sm">Visible Columns Only - {displayColumns.length} columns {displayColumns.length !== tableData.columns?.length && (<span className="text-gray-500 ml-1">({(tableData.columns?.length || 0) - displayColumns.length} hidden)</span>)}</span>
          </label>
          <label className="flex items-center gap-2">
            <input type="radio" value="all" checked={!visibleColumnsOnly} onChange={(e: ChangeEvent<HTMLInputElement>) => setVisibleColumnsOnly(!e.target.checked)} className="rounded" />
            <span className="text-sm">All Columns - {tableData.columns?.length || 0} columns</span>
          </label>
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium mb-2">Export Format:</label>
        <div className="flex gap-4">
          <label className="flex items-center gap-2">
            <input type="radio" value="csv" checked={exportFormat === 'csv'} onChange={(e: ChangeEvent<HTMLInputElement>) => setExportFormat(e.target.value as 'csv')} className="rounded" />
            <span className="text-sm">CSV (Excel compatible)</span>
          </label>
          <label className="flex items-center gap-2">
            <input type="radio" value="json" checked={exportFormat === 'json'} onChange={(e: ChangeEvent<HTMLInputElement>) => setExportFormat(e.target.value as 'json')} className="rounded" />
            <span className="text-sm">JSON (with metadata)</span>
          </label>
        </div>
      </div>

      <div>
        <label className="flex items-center gap-2">
          <input type="checkbox" checked={includeHeaders} onChange={(e: ChangeEvent<HTMLInputElement>) => setIncludeHeaders(e.target.checked)} className="rounded" />
          <span className="text-sm">Include column headers</span>
        </label>
      </div>

      <div className="p-3 bg-gray-50 rounded border text-sm">
        <div className="font-medium mb-1">Export Preview:</div>
        <div className="text-gray-600 space-y-1">
          <div>• {preview.recordCount.toLocaleString()} records</div>
          <div>• {preview.columnCount} columns</div>
          <div>• Format: {exportFormat.toUpperCase()}</div>
          {preview.isFiltered && <div className="text-blue-600">• Filtered data (filters applied)</div>}
        </div>
      </div>

      <div className="flex justify-end">
        <Button onClick={exportData} disabled={preview.recordCount === 0} className="bg-green-600 hover:bg-green-700">
          <Download className="w-4 h-4 mr-2" />
          Export {exportFormat.toUpperCase()}
        </Button>
      </div>
    </div>
  )
}
