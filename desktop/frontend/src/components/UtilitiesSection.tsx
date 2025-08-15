import React, { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Progress } from './ui/progress'
import { Alert, AlertDescription, AlertTitle } from './ui/alert'
import { 
  Database, 
  FileUp, 
  FileDown, 
  FolderOpen, 
  Archive, 
  Wrench,
  Upload,
  Download,
  HardDrive,
  RefreshCw,
  AlertCircle,
  CheckCircle,
  Info,
  FileText,
  Search,
  Settings
} from 'lucide-react'
import { GetDBFFiles } from '../../wailsjs/go/main/App'
// TODO: Import these when backend functions are implemented
// import { ImportData, ExportData, CreateBackup, RestoreBackup, OptimizeDatabase, GetDatabaseStats } from '../../wailsjs/go/main/App'
import { DBFExplorer } from './DBFExplorer'

interface UtilitiesSectionProps {
  currentUser: any
  currentCompany: string
}

const UtilitiesSection: React.FC<UtilitiesSectionProps> = ({ currentUser, currentCompany }) => {
  const [loading, setLoading] = useState(false)
  const [selectedFile, setSelectedFile] = useState('')
  const [importProgress, setImportProgress] = useState(0)
  const [exportProgress, setExportProgress] = useState(0)
  const [dbStats, setDbStats] = useState<any>(null)
  const [notification, setNotification] = useState<{ type: 'success' | 'error' | 'info', message: string } | null>(null)
  const [showDBFExplorer, setShowDBFExplorer] = useState(false)

  const canEdit = currentUser && (currentUser.is_root || currentUser.role_name === 'Admin')

  const handleImport = async () => {
    if (!canEdit) {
      setNotification({ type: 'error', message: 'You do not have permission to import data' })
      return
    }
    
    setLoading(true)
    setImportProgress(0)
    
    // Simulate import progress
    const interval = setInterval(() => {
      setImportProgress(prev => {
        if (prev >= 90) {
          clearInterval(interval)
          return prev
        }
        return prev + 10
      })
    }, 500)
    
    try {
      // TODO: Implement actual import logic
      setNotification({ type: 'info', message: 'Import functionality coming soon' })
      setImportProgress(100)
    } catch (error) {
      console.error('Import error:', error)
      setNotification({ type: 'error', message: 'Import failed' })
    } finally {
      clearInterval(interval)
      setLoading(false)
      setTimeout(() => setImportProgress(0), 2000)
    }
  }

  const handleExport = async () => {
    setLoading(true)
    setExportProgress(0)
    
    // Simulate export progress
    const interval = setInterval(() => {
      setExportProgress(prev => {
        if (prev >= 90) {
          clearInterval(interval)
          return prev
        }
        return prev + 10
      })
    }, 500)
    
    try {
      // TODO: Implement actual export logic
      setNotification({ type: 'info', message: 'Export functionality coming soon' })
      setExportProgress(100)
    } catch (error) {
      console.error('Export error:', error)
      setNotification({ type: 'error', message: 'Export failed' })
    } finally {
      clearInterval(interval)
      setLoading(false)
      setTimeout(() => setExportProgress(0), 2000)
    }
  }

  const handleBackup = async () => {
    if (!canEdit) {
      setNotification({ type: 'error', message: 'You do not have permission to create backups' })
      return
    }
    
    setLoading(true)
    try {
      // TODO: Implement actual backup logic
      setNotification({ type: 'success', message: 'Backup created successfully' })
    } catch (error) {
      console.error('Backup error:', error)
      setNotification({ type: 'error', message: 'Backup failed' })
    } finally {
      setLoading(false)
    }
  }

  const handleOptimizeDatabase = async () => {
    if (!canEdit) {
      setNotification({ type: 'error', message: 'You do not have permission to optimize the database' })
      return
    }
    
    setLoading(true)
    try {
      // TODO: Implement when backend function is available
      // await OptimizeDatabase(currentCompany, true)
      setNotification({ type: 'info', message: 'Database optimization coming soon' })
      // Refresh stats after optimization
      // const stats = await GetDatabaseStats(currentCompany)
      // setDbStats(stats)
    } catch (error) {
      console.error('Optimization error:', error)
      setNotification({ type: 'error', message: 'Database optimization failed' })
    } finally {
      setLoading(false)
    }
  }

  const loadDatabaseStats = async () => {
    try {
      // TODO: Implement when backend function is available
      // const stats = await GetDatabaseStats(currentCompany)
      // setDbStats(stats)
    } catch (error) {
      console.error('Error loading database stats:', error)
    }
  }

  React.useEffect(() => {
    loadDatabaseStats()
  }, [currentCompany])

  return (
    <div className="bg-white rounded-lg shadow-sm">
      <Tabs defaultValue="browse" className="w-full">
        {/* Tab Navigation */}
        <div className="border-b border-gray-200">
          <TabsList className="flex h-12 items-center justify-start space-x-8 px-6 bg-transparent">
            <TabsTrigger 
              value="browse" 
              className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all 
                         data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 
                         data-[state=inactive]:hover:text-gray-700 
                         data-[state=active]:after:absolute data-[state=active]:after:bottom-0 
                         data-[state=active]:after:left-0 data-[state=active]:after:right-0 
                         data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600"
            >
              <FolderOpen className="w-4 h-4 mr-2" />
              DBF Explorer
            </TabsTrigger>
            <TabsTrigger 
              value="import" 
              className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all 
                         data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 
                         data-[state=inactive]:hover:text-gray-700 
                         data-[state=active]:after:absolute data-[state=active]:after:bottom-0 
                         data-[state=active]:after:left-0 data-[state=active]:after:right-0 
                         data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600"
            >
              <Upload className="w-4 h-4 mr-2" />
              Import
            </TabsTrigger>
            <TabsTrigger 
              value="export" 
              className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all 
                         data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 
                         data-[state=inactive]:hover:text-gray-700 
                         data-[state=active]:after:absolute data-[state=active]:after:bottom-0 
                         data-[state=active]:after:left-0 data-[state=active]:after:right-0 
                         data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600"
            >
              <Download className="w-4 h-4 mr-2" />
              Export
            </TabsTrigger>
            <TabsTrigger 
              value="backup" 
              className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all 
                         data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 
                         data-[state=inactive]:hover:text-gray-700 
                         data-[state=active]:after:absolute data-[state=active]:after:bottom-0 
                         data-[state=active]:after:left-0 data-[state=active]:after:right-0 
                         data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600"
            >
              <Archive className="w-4 h-4 mr-2" />
              Backup & Restore
            </TabsTrigger>
            <TabsTrigger 
              value="maintenance" 
              className="relative h-12 px-1 pb-3 pt-3 text-sm font-medium transition-all 
                         data-[state=active]:text-gray-900 data-[state=inactive]:text-gray-500 
                         data-[state=inactive]:hover:text-gray-700 
                         data-[state=active]:after:absolute data-[state=active]:after:bottom-0 
                         data-[state=active]:after:left-0 data-[state=active]:after:right-0 
                         data-[state=active]:after:h-0.5 data-[state=active]:after:bg-blue-600"
            >
              <Wrench className="w-4 h-4 mr-2" />
              Maintenance
            </TabsTrigger>
          </TabsList>
        </div>

        {/* Notification Alert */}
        {notification && (
          <div className="p-6 pb-0">
            <Alert className={`border ${
              notification.type === 'success' ? 'border-green-200 bg-green-50' : 
              notification.type === 'error' ? 'border-red-200 bg-red-50' : 
              'border-blue-200 bg-blue-50'
            }`}>
              {notification.type === 'success' ? <CheckCircle className="h-4 w-4 text-green-600" /> : 
               notification.type === 'error' ? <AlertCircle className="h-4 w-4 text-red-600" /> :
               <Info className="h-4 w-4 text-blue-600" />}
              <AlertDescription className={`ml-2 ${
                notification.type === 'success' ? 'text-green-800' : 
                notification.type === 'error' ? 'text-red-800' : 
                'text-blue-800'
              }`}>
                {notification.message}
              </AlertDescription>
            </Alert>
          </div>
        )}

        {/* DBF Explorer Tab */}
        <TabsContent value="browse" className="p-6">
          {!showDBFExplorer ? (
            <>
              <div className="flex items-center justify-between mb-6">
                <div>
                  <h2 className="text-xl font-semibold text-gray-900">DBF Explorer</h2>
                  <p className="text-sm text-gray-500 mt-1">View and edit DBF files</p>
                </div>
                <Button 
                  onClick={() => setShowDBFExplorer(true)}
                  variant="outline" 
                  className="border-gray-200 hover:bg-gray-50"
                >
                  <FolderOpen className="w-4 h-4 mr-2" />
                  Open Explorer
                </Button>
              </div>

              <Card className="border border-gray-200 hover:shadow-md transition-all bg-white">
                <CardHeader className="pb-4 border-b border-gray-100">
                  <CardTitle className="text-base font-semibold text-gray-900">
                    <Database className="inline-block w-5 h-5 mr-2 text-gray-500" />
                    Browse DBF Files
                  </CardTitle>
                  <CardDescription className="text-sm text-gray-500">
                    View and edit legacy FoxPro database files
                  </CardDescription>
                </CardHeader>
                <CardContent className="p-4">
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-500">Company Data</span>
                      <span className="text-sm font-medium text-gray-900">{currentCompany || 'No company selected'}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-500">Access Level</span>
                      <span className="text-sm font-medium text-gray-900">
                        {canEdit ? 'Read/Write' : 'Read Only'}
                      </span>
                    </div>
                    <Button 
                      className="w-full mt-4" 
                      onClick={() => setShowDBFExplorer(true)}
                    >
                      Open DBF Explorer
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </>
          ) : (
            <DBFExplorer currentUser={currentUser} />
          )}
        </TabsContent>

        {/* Import Tab */}
        <TabsContent value="import" className="p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Data Import</h2>
              <p className="text-sm text-gray-500 mt-1">Import data from external sources</p>
            </div>
          </div>

          <Card className="border border-gray-200 hover:shadow-md transition-all bg-white">
            <CardHeader className="pb-4 border-b border-gray-100">
              <CardTitle className="text-base font-semibold text-gray-900">
                <FileUp className="inline-block w-5 h-5 mr-2 text-gray-500" />
                Import Data
              </CardTitle>
              <CardDescription className="text-sm text-gray-500">
                Import data from CSV, Excel, or other formats
              </CardDescription>
            </CardHeader>
            <CardContent className="p-4">
              <div className="space-y-4">
                <div>
                  <Label htmlFor="import-type">Import Type</Label>
                  <Select defaultValue="csv">
                    <SelectTrigger id="import-type" className="mt-1">
                      <SelectValue placeholder="Select import type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="csv">CSV File</SelectItem>
                      <SelectItem value="excel">Excel File</SelectItem>
                      <SelectItem value="json">JSON File</SelectItem>
                      <SelectItem value="bank">Bank Statement</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div>
                  <Label htmlFor="file-upload">Select File</Label>
                  <Input 
                    id="file-upload" 
                    type="file" 
                    className="mt-1"
                    accept=".csv,.xlsx,.xls,.json"
                    onChange={(e) => setSelectedFile(e.target.files?.[0]?.name || '')}
                  />
                </div>

                {importProgress > 0 && (
                  <div className="space-y-2">
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-500">Import Progress</span>
                      <span className="font-medium text-gray-900">{importProgress}%</span>
                    </div>
                    <Progress value={importProgress} className="h-2" />
                  </div>
                )}

                <Button 
                  className="w-full" 
                  onClick={handleImport}
                  disabled={loading || !selectedFile || !canEdit}
                >
                  <Upload className="w-4 h-4 mr-2" />
                  {loading ? 'Importing...' : 'Start Import'}
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Export Tab */}
        <TabsContent value="export" className="p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Data Export</h2>
              <p className="text-sm text-gray-500 mt-1">Export data to various formats</p>
            </div>
          </div>

          <Card className="border border-gray-200 hover:shadow-md transition-all bg-white">
            <CardHeader className="pb-4 border-b border-gray-100">
              <CardTitle className="text-base font-semibold text-gray-900">
                <FileDown className="inline-block w-5 h-5 mr-2 text-gray-500" />
                Export Data
              </CardTitle>
              <CardDescription className="text-sm text-gray-500">
                Export data to CSV, Excel, or PDF formats
              </CardDescription>
            </CardHeader>
            <CardContent className="p-4">
              <div className="space-y-4">
                <div>
                  <Label htmlFor="export-type">Export Type</Label>
                  <Select defaultValue="csv">
                    <SelectTrigger id="export-type" className="mt-1">
                      <SelectValue placeholder="Select export type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="csv">CSV Format</SelectItem>
                      <SelectItem value="excel">Excel Format</SelectItem>
                      <SelectItem value="pdf">PDF Report</SelectItem>
                      <SelectItem value="json">JSON Format</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div>
                  <Label htmlFor="export-table">Data Source</Label>
                  <Select defaultValue="all">
                    <SelectTrigger id="export-table" className="mt-1">
                      <SelectValue placeholder="Select data source" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All Data</SelectItem>
                      <SelectItem value="coa">Chart of Accounts</SelectItem>
                      <SelectItem value="checks">Checks</SelectItem>
                      <SelectItem value="gl">General Ledger</SelectItem>
                      <SelectItem value="vendor">Vendors</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {exportProgress > 0 && (
                  <div className="space-y-2">
                    <div className="flex justify-between text-sm">
                      <span className="text-gray-500">Export Progress</span>
                      <span className="font-medium text-gray-900">{exportProgress}%</span>
                    </div>
                    <Progress value={exportProgress} className="h-2" />
                  </div>
                )}

                <Button 
                  className="w-full" 
                  onClick={handleExport}
                  disabled={loading}
                >
                  <Download className="w-4 h-4 mr-2" />
                  {loading ? 'Exporting...' : 'Start Export'}
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Backup & Restore Tab */}
        <TabsContent value="backup" className="p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Backup & Restore</h2>
              <p className="text-sm text-gray-500 mt-1">Backup and restore database</p>
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <Card className="border border-gray-200 hover:shadow-md transition-all bg-white">
              <CardHeader className="pb-4 border-b border-gray-100">
                <CardTitle className="text-base font-semibold text-gray-900">
                  <Archive className="inline-block w-5 h-5 mr-2 text-gray-500" />
                  Create Backup
                </CardTitle>
                <CardDescription className="text-sm text-gray-500">
                  Create a backup of your data
                </CardDescription>
              </CardHeader>
              <CardContent className="p-4">
                <div className="space-y-4">
                  <div>
                    <Label htmlFor="backup-desc">Backup Description</Label>
                    <Input 
                      id="backup-desc" 
                      placeholder="Enter backup description"
                      className="mt-1"
                    />
                  </div>
                  <div className="flex items-center space-x-2">
                    <input type="checkbox" id="include-dbf" className="rounded" />
                    <Label htmlFor="include-dbf" className="text-sm font-normal">
                      Include DBF files
                    </Label>
                  </div>
                  <Button 
                    className="w-full" 
                    onClick={handleBackup}
                    disabled={loading || !canEdit}
                  >
                    <Archive className="w-4 h-4 mr-2" />
                    Create Backup
                  </Button>
                </div>
              </CardContent>
            </Card>

            <Card className="border border-gray-200 hover:shadow-md transition-all bg-white">
              <CardHeader className="pb-4 border-b border-gray-100">
                <CardTitle className="text-base font-semibold text-gray-900">
                  <RefreshCw className="inline-block w-5 h-5 mr-2 text-gray-500" />
                  Restore Backup
                </CardTitle>
                <CardDescription className="text-sm text-gray-500">
                  Restore data from a backup
                </CardDescription>
              </CardHeader>
              <CardContent className="p-4">
                <div className="space-y-4">
                  <div>
                    <Label htmlFor="backup-select">Select Backup</Label>
                    <Select>
                      <SelectTrigger id="backup-select" className="mt-1">
                        <SelectValue placeholder="Select a backup" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="backup1">Backup - 2025-01-15</SelectItem>
                        <SelectItem value="backup2">Backup - 2025-01-10</SelectItem>
                        <SelectItem value="backup3">Backup - 2025-01-05</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <Alert className="border-amber-200 bg-amber-50">
                    <AlertCircle className="h-4 w-4 text-amber-600" />
                    <AlertDescription className="ml-2 text-amber-800">
                      Restoring will replace current data
                    </AlertDescription>
                  </Alert>
                  <Button 
                    className="w-full" 
                    variant="destructive"
                    disabled={!canEdit}
                  >
                    <RefreshCw className="w-4 h-4 mr-2" />
                    Restore Backup
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Database Maintenance Tab */}
        <TabsContent value="maintenance" className="p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Database Maintenance</h2>
              <p className="text-sm text-gray-500 mt-1">Database optimization and repair</p>
            </div>
            <Button 
              onClick={loadDatabaseStats}
              variant="outline" 
              className="border-gray-200 hover:bg-gray-50"
            >
              <RefreshCw className="w-4 h-4 mr-2" />
              Refresh Stats
            </Button>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <Card className="border border-gray-200 hover:shadow-md transition-all bg-white">
              <CardHeader className="pb-4 border-b border-gray-100">
                <CardTitle className="text-base font-semibold text-gray-900">
                  <HardDrive className="inline-block w-5 h-5 mr-2 text-gray-500" />
                  Database Statistics
                </CardTitle>
              </CardHeader>
              <CardContent className="p-4">
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-500">Database Size</span>
                    <span className="text-sm font-medium text-gray-900">
                      {dbStats?.size || 'N/A'}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-500">Total Records</span>
                    <span className="text-sm font-medium text-gray-900">
                      {dbStats?.totalRecords || 'N/A'}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-500">Last Optimized</span>
                    <span className="text-sm font-medium text-gray-900">
                      {dbStats?.lastOptimized || 'Never'}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-500">Index Count</span>
                    <span className="text-sm font-medium text-gray-900">
                      {dbStats?.indexCount || 'N/A'}
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border border-gray-200 hover:shadow-md transition-all bg-white">
              <CardHeader className="pb-4 border-b border-gray-100">
                <CardTitle className="text-base font-semibold text-gray-900">
                  <Wrench className="inline-block w-5 h-5 mr-2 text-gray-500" />
                  Maintenance Tools
                </CardTitle>
              </CardHeader>
              <CardContent className="p-4">
                <div className="space-y-3">
                  <Button 
                    className="w-full" 
                    variant="outline"
                    onClick={handleOptimizeDatabase}
                    disabled={loading || !canEdit}
                  >
                    <Settings className="w-4 h-4 mr-2" />
                    Optimize Database
                  </Button>
                  <Button 
                    className="w-full" 
                    variant="outline"
                    disabled={!canEdit}
                  >
                    <Search className="w-4 h-4 mr-2" />
                    Check Integrity
                  </Button>
                  <Button 
                    className="w-full" 
                    variant="outline"
                    disabled={!canEdit}
                  >
                    <FileText className="w-4 h-4 mr-2" />
                    Rebuild Indexes
                  </Button>
                  <Button 
                    className="w-full" 
                    variant="outline"
                    disabled={!canEdit}
                  >
                    <RefreshCw className="w-4 h-4 mr-2" />
                    Clean Temp Files
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}

export default UtilitiesSection