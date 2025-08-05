import React, { useState, useMemo } from 'react'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './card'
import { Button } from './button'
import { Input } from './input'
import { Label } from './label'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './table'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './select'
import { 
  Search,
  Filter,
  ChevronLeft,
  ChevronRight,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Eye,
  RefreshCw
} from 'lucide-react'

const DataTable = ({
  data = [],
  columns = [],
  title,
  description,
  loading = false,
  error = null,
  onRowClick,
  onRefresh,
  searchPlaceholder = "Search...",
  pageSize: initialPageSize = 25,
  showSearch = true,
  showPagination = true,
  showPageSize = true,
  filters = [],
  actions = [],
  className = ""
}) => {
  // State
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(initialPageSize)
  const [sortField, setSortField] = useState('')
  const [sortDirection, setSortDirection] = useState('asc')
  const [searchTerm, setSearchTerm] = useState('')
  const [activeFilters, setActiveFilters] = useState({})

  // Process data with filters, search, and sort
  const processedData = useMemo(() => {
    let filtered = [...data]
    
    // Apply search
    if (searchTerm && showSearch) {
      const term = searchTerm.toLowerCase()
      filtered = filtered.filter(row => 
        columns.some(col => {
          const value = row[col.accessor]
          return value && value.toString().toLowerCase().includes(term)
        })
      )
    }
    
    // Apply filters
    Object.entries(activeFilters).forEach(([key, value]) => {
      if (value && value !== 'all') {
        const filter = filters.find(f => f.key === key)
        if (filter && filter.filterFn) {
          filtered = filtered.filter(row => filter.filterFn(row, value))
        } else {
          // Default filter: exact match
          filtered = filtered.filter(row => row[key] === value)
        }
      }
    })
    
    // Sort
    if (sortField) {
      filtered.sort((a, b) => {
        const column = columns.find(col => col.accessor === sortField)
        let aVal = a[sortField]
        let bVal = b[sortField]
        
        // Handle custom sort function
        if (column && column.sortFn) {
          return column.sortFn(a, b, sortDirection)
        }
        
        // Handle different data types
        if (column && column.type === 'number') {
          aVal = parseFloat(aVal) || 0
          bVal = parseFloat(bVal) || 0
        } else if (column && column.type === 'date') {
          aVal = new Date(aVal || 0)
          bVal = new Date(bVal || 0)
        } else {
          aVal = (aVal || '').toString().toLowerCase()
          bVal = (bVal || '').toString().toLowerCase()
        }
        
        if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1
        if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1
        return 0
      })
    }
    
    return filtered
  }, [data, searchTerm, activeFilters, sortField, sortDirection, columns, filters, showSearch])

  // Pagination
  const paginatedData = useMemo(() => {
    if (!showPagination) return processedData
    const startIndex = (currentPage - 1) * pageSize
    return processedData.slice(startIndex, startIndex + pageSize)
  }, [processedData, currentPage, pageSize, showPagination])

  const totalPages = Math.ceil(processedData.length / pageSize)

  // Handle sort
  const handleSort = (field) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortDirection('asc')
    }
  }

  // Get sort icon
  const getSortIcon = (field) => {
    if (sortField !== field) return <ArrowUpDown className="h-4 w-4" />
    return sortDirection === 'asc' ? 
      <ArrowUp className="h-4 w-4" /> : 
      <ArrowDown className="h-4 w-4" />
  }

  // Handle filter change
  const handleFilterChange = (key, value) => {
    setActiveFilters(prev => ({
      ...prev,
      [key]: value
    }))
    setCurrentPage(1) // Reset to first page when filtering
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Header with filters */}
      {(showSearch || filters.length > 0 || actions.length > 0) && (
        <div className="flex gap-4 items-end">
          {/* Search */}
          {showSearch && (
            <div className="flex-1 max-w-sm">
              <Label htmlFor="search">Search</Label>
              <div className="relative">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  id="search"
                  placeholder={searchPlaceholder}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>
          )}

          {/* Filters */}
          {filters.map((filter) => (
            <div key={filter.key} className="flex-1 max-w-xs">
              <Label htmlFor={`filter-${filter.key}`}>{filter.label}</Label>
              <Select 
                value={activeFilters[filter.key] || filter.defaultValue || 'all'} 
                onValueChange={(value) => handleFilterChange(filter.key, value)}
              >
                <SelectTrigger id={`filter-${filter.key}`}>
                  <SelectValue placeholder={filter.placeholder} />
                </SelectTrigger>
                <SelectContent>
                  {filter.options.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          ))}

          {/* Page Size */}
          {showPageSize && showPagination && (
            <div className="w-32">
              <Label htmlFor="page-size">Page Size</Label>
              <Select value={pageSize.toString()} onValueChange={(v) => setPageSize(parseInt(v))}>
                <SelectTrigger id="page-size">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="10">10</SelectItem>
                  <SelectItem value="25">25</SelectItem>
                  <SelectItem value="50">50</SelectItem>
                  <SelectItem value="100">100</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}

          {/* Custom Actions */}
          {actions.map((action, index) => (
            <Button
              key={index}
              variant={action.variant || "outline"}
              onClick={action.onClick}
              disabled={action.disabled}
            >
              {action.icon}
              {action.label}
            </Button>
          ))}

          {/* Refresh Button */}
          {onRefresh && (
            <Button onClick={onRefresh} disabled={loading}>
              <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          )}
        </div>
      )}

      {/* Error Display */}
      {error && (
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2 text-destructive">
              <span>{error}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Data Table */}
      <Card>
        {(title || description) && (
          <CardHeader>
            {title && <CardTitle>{title}</CardTitle>}
            <CardDescription>
              {loading ? 'Loading...' : 
               processedData.length === 0 ? 'No data found' :
               showPagination ? 
                 `Showing ${paginatedData.length} of ${processedData.length} items (Page ${currentPage} of ${totalPages})` :
                 `Showing ${processedData.length} items`
              }
            </CardDescription>
          </CardHeader>
        )}
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <RefreshCw className="h-6 w-6 animate-spin" />
              <span className="ml-2">Loading...</span>
            </div>
          ) : processedData.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              {error ? 'Unable to load data' : 'No data found'}
            </div>
          ) : (
            <>
              <div className="overflow-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      {columns.map((column) => (
                        <TableHead 
                          key={column.accessor}
                          className={column.headerClassName || ''}
                        >
                          {column.sortable !== false ? (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleSort(column.accessor)}
                              className="h-8 px-2"
                            >
                              {column.header}
                              {getSortIcon(column.accessor)}
                            </Button>
                          ) : (
                            column.header
                          )}
                        </TableHead>
                      ))}
                      {onRowClick && (
                        <TableHead className="text-center">Actions</TableHead>
                      )}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedData.map((row, index) => (
                      <TableRow 
                        key={row.id || index} 
                        className={onRowClick ? "cursor-pointer hover:bg-muted/50" : ""}
                        onClick={() => onRowClick && onRowClick(row, index)}
                      >
                        {columns.map((column) => (
                          <TableCell 
                            key={column.accessor}
                            className={column.cellClassName || ''}
                          >
                            {column.render ? 
                              column.render(row[column.accessor], row, index) : 
                              row[column.accessor]
                            }
                          </TableCell>
                        ))}
                        {onRowClick && (
                          <TableCell className="text-center">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={(e) => {
                                e.stopPropagation()
                                onRowClick(row, index)
                              }}
                            >
                              <Eye className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        )}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {/* Pagination Controls */}
              {showPagination && totalPages > 1 && (
                <div className="flex items-center justify-between mt-4">
                  <div className="text-sm text-muted-foreground">
                    Showing {((currentPage - 1) * pageSize) + 1} to {Math.min(currentPage * pageSize, processedData.length)} of {processedData.length} items
                  </div>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                      disabled={currentPage === 1}
                    >
                      <ChevronLeft className="h-4 w-4" />
                      Previous
                    </Button>
                    
                    {/* Page numbers */}
                    <div className="flex gap-1">
                      {[...Array(Math.min(5, totalPages))].map((_, i) => {
                        let pageNum
                        if (totalPages <= 5) {
                          pageNum = i + 1
                        } else if (currentPage <= 3) {
                          pageNum = i + 1
                        } else if (currentPage >= totalPages - 2) {
                          pageNum = totalPages - 4 + i
                        } else {
                          pageNum = currentPage - 2 + i
                        }
                        
                        return (
                          <Button
                            key={i}
                            variant={pageNum === currentPage ? "default" : "outline"}
                            size="sm"
                            onClick={() => setCurrentPage(pageNum)}
                            className="w-8"
                          >
                            {pageNum}
                          </Button>
                        )
                      })}
                    </div>
                    
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                      disabled={currentPage === totalPages}
                    >
                      Next
                      <ChevronRight className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

export default DataTable