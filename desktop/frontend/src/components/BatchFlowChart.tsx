import React, { useState } from 'react'
import { Card } from './ui/card'
import { ArrowDown, Check, FileText, CreditCard, Receipt, ChevronRight } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from './ui/dialog'

interface TableData {
  tableName: string
  records: any[]
  count: number
  columns?: string[]
}

interface BatchFlowChartProps {
  batchNumber: string
  searchResults: {
    checks?: TableData
    glmaster?: TableData
    glmaster_purchase?: TableData
    appmthdr?: TableData
    appmtdet?: TableData
    appurchh?: TableData
    appurchd?: TableData
  }
}

const BatchFlowChart: React.FC<BatchFlowChartProps> = ({ batchNumber, searchResults }) => {
  const [selectedTable, setSelectedTable] = useState<TableData | null>(null)
  const [showRecordsDialog, setShowRecordsDialog] = useState(false)

  // Helper function to format values for display
  const formatValue = (value: any, fieldName?: string) => {
    if (value === null || value === undefined || value === '') return '-'
    
    // Check if it's a date field
    if (fieldName && (fieldName.toLowerCase().includes('date') || 
        fieldName.startsWith('D') || fieldName.includes('DCHANGED') || 
        fieldName.includes('DLASTMODIF') || fieldName.includes('DADDED'))) {
      // Handle various date formats
      if (typeof value === 'string') {
        // Remove any time component for cleaner display
        const dateStr = value.split('T')[0]
        if (dateStr && dateStr.includes('-')) {
          const parts = dateStr.split('-')
          if (parts.length === 3) {
            return `${parts[1]}/${parts[2]}/${parts[0]}` // Convert to MM/DD/YYYY
          }
        } else if (dateStr && dateStr.includes('/')) {
          return dateStr // Already in correct format
        }
      }
    }
    
    // Format numeric fields (amounts, debits, credits)
    if (fieldName && (fieldName.startsWith('N') || fieldName.includes('AMOUNT') || 
        fieldName === 'NDEBITS' || fieldName === 'NCREDITS')) {
      const num = typeof value === 'string' ? parseFloat(value) : value
      if (!isNaN(num)) {
        return new Intl.NumberFormat('en-US', {
          minimumFractionDigits: 2,
          maximumFractionDigits: 2
        }).format(num)
      }
    }
    
    // Format year fields
    if (fieldName === 'CYEAR' || fieldName === 'CPERIOD') {
      return String(value).trim()
    }
    
    if (typeof value === 'boolean') return value ? 'Yes' : 'No'
    
    // Trim long strings if needed
    const strValue = String(value).trim()
    if (strValue.length > 50) {
      return strValue.substring(0, 47) + '...'
    }
    return strValue
  }

  // Helper function to render a table node in the flow chart
  const renderTableNode = (
    icon: React.ReactNode,
    title: string,
    subtitle: string,
    data: TableData | undefined,
    highlight?: string,
    bgColor = "bg-white"
  ) => {
    const hasData = data && data.count > 0
    
    return (
      <div 
        className={`${bgColor} rounded-lg border-2 ${hasData ? 'border-blue-500 shadow-lg cursor-pointer hover:shadow-xl transition-shadow' : 'border-gray-300'} p-4 w-[240px] h-[140px] flex flex-col`}
        onClick={() => {
          if (hasData && data) {
            setSelectedTable(data)
            setShowRecordsDialog(true)
          }
        }}
      >
        <div className="flex items-center space-x-2 mb-2">
          <div className={`p-2 rounded-full ${hasData ? 'bg-blue-100 text-blue-600' : 'bg-gray-100 text-gray-400'}`}>
            {icon}
          </div>
          <div className="flex-1">
            <h3 className="font-semibold text-sm">{title}</h3>
            <p className="text-xs text-gray-500">{subtitle}</p>
          </div>
        </div>
        <div className="flex-1 flex flex-col justify-center">
          {hasData ? (
            <>
              <div className="text-sm font-medium text-green-600">
                {data.count} record{data.count !== 1 ? 's' : ''} found
              </div>
              {highlight && (
                <div className="text-xs text-gray-600 mt-1">
                  {highlight}
                </div>
              )}
              <div className="text-xs text-blue-500 mt-1">
                Click to view records
              </div>
            </>
          ) : (
            <div className="text-sm text-gray-400">No records</div>
          )}
        </div>
      </div>
    )
  }

  // Helper to render connecting arrow
  const renderArrow = (label?: string, isActive = false) => (
    <div className="flex flex-col items-center my-2">
      <ArrowDown className={`h-8 w-8 ${isActive ? 'text-blue-500' : 'text-gray-300'}`} />
      {label && (
        <span className={`text-xs mt-1 ${isActive ? 'text-blue-600 font-medium' : 'text-gray-400'}`}>
          {label}
        </span>
      )}
    </div>
  )

  // Helper to render horizontal connection
  const renderHorizontalConnection = (label: string, isActive = false) => (
    <div className="flex items-center mx-4">
      <div className={`flex-1 h-0.5 ${isActive ? 'bg-blue-500' : 'bg-gray-300'}`} />
      <span className={`mx-2 text-xs ${isActive ? 'text-blue-600 font-medium' : 'text-gray-400'}`}>
        {label}
      </span>
      <ChevronRight className={`h-4 w-4 ${isActive ? 'text-blue-500' : 'text-gray-300'}`} />
    </div>
  )

  // Determine connections based on data
  const hasChecks = searchResults.checks && searchResults.checks.count > 0
  const hasGL = searchResults.glmaster && searchResults.glmaster.count > 0
  const hasPurchaseGL = searchResults.glmaster_purchase && searchResults.glmaster_purchase.count > 0
  const hasAPMTDET = searchResults.appmtdet && searchResults.appmtdet.count > 0
  const hasAPMTHDR = searchResults.appmthdr && searchResults.appmthdr.count > 0
  const hasAPPURCHH = searchResults.appurchh && searchResults.appurchh.count > 0
  const hasAPPURCHD = searchResults.appurchd && searchResults.appurchd.count > 0

  return (
    <Card className="p-6 bg-gray-50">
      <div className="mb-4">
        <h3 className="text-lg font-semibold text-gray-900">Batch Flow Visualization</h3>
        <p className="text-sm text-gray-600 mt-1">
          Tracing batch <span className="font-mono font-semibold">{batchNumber}</span> through the system
        </p>
      </div>

      <div className="overflow-x-auto">
        <div className="min-w-[800px] p-4">
          {/* Level 1: Check Entry Point */}
          <div className="flex justify-center">
            {renderTableNode(
              <Check className="h-4 w-4" />,
              "CHECKS.DBF",
              "Check Register",
              searchResults.checks,
              `CBATCH: ${batchNumber}`,
              "bg-blue-50"
            )}
          </div>

          {renderArrow("Check posted to GL", hasChecks && hasGL)}

          {/* Level 2: GL Master - Check Payment */}
          <div className="flex justify-center">
            {renderTableNode(
              <FileText className="h-4 w-4" />,
              "GLMASTER.DBF",
              "GL - Check Payment",
              searchResults.glmaster,
              "Payment Entry"
            )}
          </div>

          {renderArrow("Find AP transactions", hasGL && (hasAPMTDET || hasAPMTHDR))}

          {/* Level 3: Payment Header and Details - Side by Side */}
          <div className="flex justify-center space-x-8">
            <div className="flex flex-col items-center">
              {renderTableNode(
                <Receipt className="h-4 w-4" />,
                "APPMTHDR.DBF",
                "Payment Header",
                searchResults.appmthdr,
                `CBATCH: ${batchNumber}`
              )}
            </div>
            <div className="flex flex-col items-center">
              {renderTableNode(
                <CreditCard className="h-4 w-4" />,
                "APPMTDET.DBF",
                "Payment Details",
                searchResults.appmtdet,
                `CBATCH: ${batchNumber}`
              )}
            </div>
          </div>

          {renderArrow("Link to original purchases", (hasAPMTHDR || hasAPMTDET) && (hasAPPURCHH || hasAPPURCHD))}

          {/* Level 4: GL Master - Original Purchase Entry */}
          <div className="flex justify-center">
            {renderTableNode(
              <FileText className="h-4 w-4" />,
              "GLMASTER.DBF",
              "GL - Purchase Entry",
              searchResults.glmaster_purchase || (searchResults.appmtdet && searchResults.appmtdet.count > 0 ? {
                tableName: "GLMASTER.DBF",
                records: [],
                count: 0,
                columns: []
              } : undefined),
              "CSOURCE = 'AP'"
            )}
          </div>

          {renderArrow("Original purchase documents", (hasPurchaseGL || hasAPMTDET) && (hasAPPURCHH || hasAPPURCHD))}

          {/* Level 5: Purchase Documents - Side by Side */}
          <div className="flex justify-center space-x-8">
            <div className="flex flex-col items-center">
              {renderTableNode(
                <FileText className="h-4 w-4" />,
                "APPURCHH.DBF",
                "Purchase Header",
                searchResults.appurchh,
                `Original Invoice`
              )}
            </div>
            <div className="flex flex-col items-center">
              {renderTableNode(
                <FileText className="h-4 w-4" />,
                "APPURCHD.DBF",
                "Purchase Details",
                searchResults.appurchd,
                `Line Items`
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Legend */}
      <div className="mt-6 pt-4 border-t border-gray-200">
        <h4 className="text-sm font-medium text-gray-700 mb-2">Legend:</h4>
        <div className="flex flex-wrap gap-4 text-xs">
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 bg-blue-50 border-2 border-blue-500 rounded"></div>
            <span>Entry Point</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 bg-white border-2 border-blue-500 rounded"></div>
            <span>Records Found</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-4 h-4 bg-white border-2 border-gray-300 rounded"></div>
            <span>No Records</span>
          </div>
          <div className="flex items-center space-x-2">
            <ArrowDown className="h-4 w-4 text-blue-500" />
            <span>Active Path</span>
          </div>
          <div className="flex items-center space-x-2">
            <ArrowDown className="h-4 w-4 text-gray-300" />
            <span>Inactive Path</span>
          </div>
        </div>
      </div>

      {/* Explanation */}
      <div className="mt-4 p-4 bg-amber-50 rounded-lg border border-amber-200">
        <h4 className="text-sm font-medium text-amber-900 mb-1">Complete Transaction Flow:</h4>
        <ol className="text-xs text-amber-800 space-y-1 list-decimal list-inside">
          <li><strong>Check Entry:</strong> Start with check in CHECKS.DBF using CBATCH</li>
          <li><strong>Payment GL:</strong> Find GL transaction for the check payment (same CBATCH)</li>
          <li><strong>Payment Records:</strong> Find APPMTHDR and APPMTDET using CBATCH</li>
          <li><strong>Extract Purchase Batch:</strong> Get CBILLTOKEN from APPMTDET - this is the original purchase batch</li>
          <li><strong>Purchase GL:</strong> Find GL with CBATCH = CBILLTOKEN (CSOURCE = 'AP')</li>
          <li><strong>Original Purchases:</strong> Find APPURCHH and APPURCHD with CBATCH = CBILLTOKEN</li>
        </ol>
        <p className="text-xs text-amber-700 mt-2 font-medium">
          Key: APPMTDET.CBILLTOKEN contains the original purchase batch number
        </p>
      </div>

      {/* Records Detail Dialog */}
      <Dialog open={showRecordsDialog} onOpenChange={setShowRecordsDialog}>
        <DialogContent className="max-w-4xl max-h-[90vh] flex flex-col p-0">
          <DialogHeader className="flex-shrink-0 px-6 pt-6 pb-4 border-b">
            <DialogTitle className="flex items-center justify-between">
              <span>{selectedTable?.tableName || 'Table Records'}</span>
              {selectedTable && selectedTable.count > 2 && (
                <span className="text-xs font-normal text-gray-500 bg-gray-100 px-2 py-1 rounded">
                  Scroll to view all {selectedTable.count} records
                </span>
              )}
            </DialogTitle>
            <DialogDescription>
              {selectedTable?.count} record{selectedTable?.count !== 1 ? 's' : ''} found with batch number: {batchNumber}
            </DialogDescription>
          </DialogHeader>
          
          <div className="flex-1 overflow-y-auto min-h-0 max-h-[calc(90vh-120px)] px-6 py-4 custom-scrollbar">
            {selectedTable && selectedTable.records && selectedTable.records.length > 0 ? (
              <div className="space-y-4">
                {selectedTable.records.map((record, idx) => (
                  <div key={idx} className="border rounded-lg p-4 hover:bg-gray-50">
                    <div className="flex items-center justify-between mb-3">
                      <span className="font-medium text-sm text-gray-700">
                        Record #{idx + 1}
                      </span>
                      {record.CBATCH && (
                        <span className="text-xs bg-blue-100 text-blue-700 px-2 py-1 rounded">
                          Batch: {record.CBATCH}
                        </span>
                      )}
                    </div>
                    <div className="space-y-2">
                      {Object.entries(record).map(([key, value]) => (
                        <div key={key} className="text-sm grid grid-cols-[140px_1fr] gap-2">
                          <span className="font-medium text-gray-600 text-right">{key}:</span>
                          <span className="text-gray-900">
                            {formatValue(value, key)}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-gray-500">
                No records to display
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* Custom scrollbar styles */}
      <style jsx>{`
        .custom-scrollbar::-webkit-scrollbar {
          width: 8px;
        }
        
        .custom-scrollbar::-webkit-scrollbar-track {
          background: #f1f1f1;
          border-radius: 4px;
        }
        
        .custom-scrollbar::-webkit-scrollbar-thumb {
          background: #888;
          border-radius: 4px;
        }
        
        .custom-scrollbar::-webkit-scrollbar-thumb:hover {
          background: #555;
        }
        
        /* For Firefox */
        .custom-scrollbar {
          scrollbar-width: thin;
          scrollbar-color: #888 #f1f1f1;
        }
      `}</style>
    </Card>
  )
}

export default BatchFlowChart