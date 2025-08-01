import { useState, useEffect } from 'react'
import { Login, Register, GetCompanies, ValidateSession, GetDashboardData } from '../wailsjs/go/main/App'
import { Button } from './components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './components/ui/card'
import { Input } from './components/ui/input'
import { Sidebar, SidebarHeader, SidebarContent, SidebarNav, SidebarNavItem, SidebarNavGroup } from './components/ui/sidebar'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './components/ui/tabs'
import { RevenueChart } from './components/charts/RevenueChart'
import { ProductionChart } from './components/charts/ProductionChart'
import { TransactionsTable } from './components/tables/TransactionsTable'
import { DBFExplorer } from './components/DBFExplorer'
import { UserManagement } from './components/UserManagement'
import { ThemeProvider } from './components/theme-provider'
import { ThemeSwitcher } from './components/theme-switcher'
import { 
  Home, 
  Database, 
  FileText, 
  BarChart3, 
  Settings, 
  LogOut,
  DollarSign,
  TrendingUp,
  Users,
  Activity,
  Table,
  Calculator,
  FileSearch,
  Archive,
  Download,
  Upload,
  Copy,
  Calendar,
  Wrench
} from 'lucide-react'
import './globals.css'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [currentUser, setCurrentUser] = useState(null)
  const [companies, setCompanies] = useState([])
  const [showRegister, setShowRegister] = useState(false)
  const [loading, setLoading] = useState(true)

  // Form states
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [email, setEmail] = useState('')
  const [selectedCompany, setSelectedCompany] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    checkSession()
    loadCompanies()
  }, [])

  const checkSession = async () => {
    const token = localStorage.getItem('session_token')
    const companyName = localStorage.getItem('company_name')
    
    if (token && companyName) {
      try {
        const user = await ValidateSession(token, companyName)
        setCurrentUser(user)
        setIsAuthenticated(true)
      } catch (error) {
        localStorage.removeItem('session_token')
        localStorage.removeItem('company_name')
      }
    }
    setLoading(false)
  }

  const loadCompanies = async () => {
    try {
      const companiesList = await GetCompanies()
      setCompanies(companiesList || [])
    } catch (error) {
      console.error('Failed to load companies:', error)
    }
  }

  const handleLogin = async (e) => {
    e.preventDefault()
    setError('')
    
    if (!username || !password || !selectedCompany) {
      setError('Please fill in all fields')
      return
    }

    try {
      const result = await Login(username, password, selectedCompany)
      localStorage.setItem('session_token', result.session.token)
      localStorage.setItem('company_name', result.user.company_name)
      setCurrentUser(result.user)
      setIsAuthenticated(true)
    } catch (error) {
      setError(error.message || 'Login failed')
    }
  }

  const handleRegister = async (e) => {
    e.preventDefault()
    setError('')
    
    if (!username || !password || !email || !selectedCompany) {
      setError('Please fill in all fields')
      return
    }

    try {
      const result = await Register(username, password, email, selectedCompany)
      localStorage.setItem('session_token', result.session.token)
      localStorage.setItem('company_name', result.user.company_name)
      setCurrentUser(result.user)
      setIsAuthenticated(true)
    } catch (error) {
      setError(error.message || 'Registration failed')
    }
  }

  const handleLogout = () => {
    localStorage.removeItem('session_token')
    localStorage.removeItem('company_name')
    setCurrentUser(null)
    setIsAuthenticated(false)
    setUsername('')
    setPassword('')
    setEmail('')
    setSelectedCompany('')
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-lg">Loading...</div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <CardTitle className="text-2xl font-bold text-blue-900">
              Pivoten FinancialsX
            </CardTitle>
            <CardDescription>
              {showRegister ? 'Create your account' : 'Sign in to your account'}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={showRegister ? handleRegister : handleLogin} className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Username</label>
                <Input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="Enter your username"
                />
              </div>
              
              <div className="space-y-2">
                <label className="text-sm font-medium">Password</label>
                <Input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Enter your password"
                />
              </div>

              {showRegister && (
                <div className="space-y-2">
                  <label className="text-sm font-medium">Email</label>
                  <Input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="Enter your email"
                  />
                </div>
              )}

              <div className="space-y-2">
                <label className="text-sm font-medium">Company</label>
                <select
                  value={selectedCompany}
                  onChange={(e) => setSelectedCompany(e.target.value)}
                  className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                >
                  <option value="">Select a company</option>
                  {companies.map((company) => (
                    <option key={company.name} value={company.name}>
                      {company.name}
                    </option>
                  ))}
                </select>
              </div>

              {error && (
                <div className="text-sm text-red-600 bg-red-50 p-3 rounded-md">
                  {error}
                </div>
              )}

              <Button type="submit" className="w-full">
                {showRegister ? 'Register' : 'Sign In'}
              </Button>

              <div className="text-center">
                <button
                  type="button"
                  onClick={() => setShowRegister(!showRegister)}
                  className="text-sm text-blue-600 hover:underline"
                >
                  {showRegister 
                    ? 'Already have an account? Sign in' 
                    : "Don't have an account? Register"}
                </button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    )
  }

  // Advanced Dashboard with Sidebar Navigation
  return (
    <ThemeProvider defaultTheme="light" defaultColorScheme="blue">
      <AdvancedDashboard currentUser={currentUser} onLogout={handleLogout} />
    </ThemeProvider>
  )
}

// Advanced Dashboard Component with Sidebar and Multiple Views
function AdvancedDashboard({ currentUser, onLogout }) {
  const [activeView, setActiveView] = useState('dashboard')
  const [dashboardData, setDashboardData] = useState(null)
  const [loadingDashboard, setLoadingDashboard] = useState(true)

  // Load dashboard data on component mount
  useEffect(() => {
    loadDashboardData()
  }, [currentUser])

  const loadDashboardData = async () => {
    if (!currentUser?.company_name) return
    
    setLoadingDashboard(true)
    try {
      console.log('Loading dashboard data for:', currentUser.company_name)
      const data = await GetDashboardData(currentUser.company_name)
      console.log('Dashboard data loaded:', data)
      setDashboardData(data)
    } catch (error) {
      console.error('Failed to load dashboard data:', error)
    } finally {
      setLoadingDashboard(false)
    }
  }

  // Generate stats from real data
  const generateStats = () => {
    if (!dashboardData) {
      return [
        { title: 'Loading...', value: '-', change: '-', icon: DollarSign, trend: 'up' },
        { title: 'Loading...', value: '-', change: '-', icon: Activity, trend: 'up' },
        { title: 'Loading...', value: '-', change: '-', icon: TrendingUp, trend: 'up' },
        { title: 'Loading...', value: '-', change: '-', icon: BarChart3, trend: 'up' },
      ]
    }

    const stats = []
    
    // DBF Files stat
    if (dashboardData.fileStats) {
      stats.push({
        title: 'DBF Files',
        value: dashboardData.fileStats.totalFiles.toString(),
        change: 'files available',
        icon: FileText,
        trend: 'up'
      })
    }

    // Wells stat
    if (dashboardData.wells && dashboardData.wells.hasData) {
      stats.push({
        title: 'Active Wells',
        value: dashboardData.wells.totalWells.toString(),
        change: 'wells in database',
        icon: Activity,
        trend: 'up'
      })
    }

    // Checks stat
    if (dashboardData.checks && dashboardData.checks.hasData) {
      stats.push({
        title: 'Check Records',
        value: dashboardData.checks.totalChecks.toLocaleString(),
        change: 'payment records',
        icon: DollarSign,
        trend: 'up'
      })
    }

    // Financial records stat
    if (dashboardData.financials && dashboardData.financials.hasData) {
      const totalRecords = (dashboardData.financials.incomeRecords || 0) + (dashboardData.financials.expenseRecords || 0)
      stats.push({
        title: 'Financial Records',
        value: totalRecords.toLocaleString(),
        change: 'income + expense records',
        icon: TrendingUp,
        trend: 'up'
      })
    }

    // Fill remaining slots if needed
    while (stats.length < 4) {
      stats.push({
        title: 'Data Available',
        value: 'Ready',
        change: 'system operational',
        icon: BarChart3,
        trend: 'up'
      })
    }

    return stats.slice(0, 4)
  }

  const stats = generateStats()

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <Sidebar className="hidden lg:flex border-r">
        <SidebarHeader className="border-b px-6 py-4">
          <h1 className="text-xl font-bold text-primary">FinancialsX</h1>
        </SidebarHeader>
        <SidebarContent>
          <SidebarNavGroup title="Main">
            <SidebarNavItem 
              href="#" 
              active={activeView === 'dashboard'}
              onClick={() => setActiveView('dashboard')}
            >
              <Home className="w-4 h-4" />
              Dashboard
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'transactions'}
              onClick={() => setActiveView('transactions')}
            >
              <DollarSign className="w-4 h-4" />
              Transactions
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'analytics'}
              onClick={() => setActiveView('analytics')}
            >
              <BarChart3 className="w-4 h-4" />
              Analytics
            </SidebarNavItem>
          </SidebarNavGroup>
          
          <SidebarNavGroup title="Management">
            <SidebarNavItem 
              href="#" 
              active={activeView === 'dbf-explorer'}
              onClick={() => setActiveView('dbf-explorer')}
            >
              <Table className="w-4 h-4" />
              DBF Explorer
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'db-maintenance'}
              onClick={() => setActiveView('db-maintenance')}
            >
              <Database className="w-4 h-4" />
              DB Maintenance
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'state-reporting'}
              onClick={() => setActiveView('state-reporting')}
            >
              <FileText className="w-4 h-4" />
              State Reporting
            </SidebarNavItem>
          </SidebarNavGroup>
          
          <SidebarNavGroup title="Utilities">
            <SidebarNavItem 
              href="#" 
              active={activeView === 'calculator'}
              onClick={() => setActiveView('calculator')}
            >
              <Calculator className="w-4 h-4" />
              Calculator
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'data-import'}
              onClick={() => setActiveView('data-import')}
            >
              <Upload className="w-4 h-4" />
              Data Import
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'data-export'}
              onClick={() => setActiveView('data-export')}
            >
              <Download className="w-4 h-4" />
              Data Export
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'backup'}
              onClick={() => setActiveView('backup')}
            >
              <Archive className="w-4 h-4" />
              Backup & Restore
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'audit-log'}
              onClick={() => setActiveView('audit-log')}
            >
              <FileSearch className="w-4 h-4" />
              Audit Log
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'scheduler'}
              onClick={() => setActiveView('scheduler')}
            >
              <Calendar className="w-4 h-4" />
              Task Scheduler
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'data-tools'}
              onClick={() => setActiveView('data-tools')}
            >
              <Wrench className="w-4 h-4" />
              Data Tools
            </SidebarNavItem>
          </SidebarNavGroup>
          
          <SidebarNavGroup title="System">
            <SidebarNavItem 
              href="#" 
              active={activeView === 'settings'}
              onClick={() => setActiveView('settings')}
            >
              <Settings className="w-4 h-4" />
              Settings
            </SidebarNavItem>
            <SidebarNavItem href="#" onClick={onLogout}>
              <LogOut className="w-4 h-4" />
              Logout
            </SidebarNavItem>
          </SidebarNavGroup>
        </SidebarContent>
      </Sidebar>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-semibold tracking-tight">
                {activeView === 'dashboard' && 'Dashboard'}
                {activeView === 'transactions' && 'Transactions'}
                {activeView === 'analytics' && 'Analytics'}
                {activeView === 'dbf-explorer' && 'DBF Explorer'}
                {activeView === 'db-maintenance' && 'Database Maintenance'}
                {activeView === 'state-reporting' && 'State Reporting'}
                {activeView === 'calculator' && 'Calculator'}
                {activeView === 'data-import' && 'Data Import'}
                {activeView === 'data-export' && 'Data Export'}
                {activeView === 'backup' && 'Backup & Restore'}
                {activeView === 'audit-log' && 'Audit Log'}
                {activeView === 'scheduler' && 'Task Scheduler'}
                {activeView === 'data-tools' && 'Data Tools'}
                {activeView === 'settings' && 'Settings'}
              </h2>
              <p className="text-sm text-muted-foreground">
                {activeView === 'dashboard' && 'Overview of your financial data'}
                {activeView === 'transactions' && 'View and manage all transactions'}
                {activeView === 'analytics' && 'Detailed charts and reports'}
                {activeView === 'dbf-explorer' && 'Browse and edit DBF tables like a spreadsheet'}
                {activeView === 'db-maintenance' && 'Database management tools'}
                {activeView === 'state-reporting' && 'West Virginia reporting'}
                {activeView === 'calculator' && 'Financial calculations and tools'}
                {activeView === 'data-import' && 'Import data from external sources'}
                {activeView === 'data-export' && 'Export data to various formats'}
                {activeView === 'backup' && 'Backup and restore system data'}
                {activeView === 'audit-log' && 'System activity and change tracking'}
                {activeView === 'scheduler' && 'Schedule automated tasks and reports'}
                {activeView === 'data-tools' && 'Data validation and transformation utilities'}
                {activeView === 'settings' && 'System configuration and user management'}
              </p>
            </div>
            <div className="flex items-center space-x-4">
              <ThemeSwitcher />
              <span className="text-sm text-muted-foreground">
                Welcome, {currentUser?.username}
              </span>
              <span className="px-2 py-1 bg-primary/10 text-primary rounded-full text-xs font-medium">
                {currentUser?.company_name}
              </span>
            </div>
          </div>
        </header>

        {/* Content Area */}
        <main className="flex-1 overflow-auto p-6 bg-muted/30">
          {activeView === 'dashboard' && (
            <div className="space-y-6">
              {loadingDashboard && (
                <div className="text-center py-8">
                  <p className="text-muted-foreground">Loading dashboard data...</p>
                </div>
              )}
              
              {/* Stats Grid */}
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                {stats.map((stat, index) => (
                  <Card key={index}>
                    <CardContent className="p-6">
                      <div className="flex items-center justify-between space-y-0 pb-2">
                        <div className="text-sm font-medium text-muted-foreground">
                          {stat.title}
                        </div>
                        <stat.icon className="w-4 h-4 text-muted-foreground" />
                      </div>
                      <div className="space-y-1">
                        <div className="text-2xl font-bold">{stat.value}</div>
                        <p className={`text-xs ${
                          stat.trend === 'up' ? 'text-green-600' : 'text-red-600'
                        }`}>
                          {stat.change}
                        </p>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>

              {/* Real Data Summary */}
              {dashboardData && !loadingDashboard && (
                <Card>
                  <CardHeader>
                    <CardTitle>Data Summary for {currentUser?.company_name}</CardTitle>
                    <CardDescription>Overview of available data in your DBF files</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid gap-4 md:grid-cols-2">
                      <div>
                        <h4 className="font-semibold mb-2">Available Files</h4>
                        <div className="space-y-1 text-sm">
                          {dashboardData.fileStats?.files?.slice(0, 8).map(file => (
                            <div key={file} className="flex items-center gap-2">
                              <FileText className="w-3 h-3 text-muted-foreground" />
                              {file}
                            </div>
                          ))}
                          {dashboardData.fileStats?.files?.length > 8 && (
                            <div className="text-muted-foreground text-xs">
                              ...and {dashboardData.fileStats.files.length - 8} more files
                            </div>
                          )}
                        </div>
                      </div>
                      <div>
                        <h4 className="font-semibold mb-2">Data Status</h4>
                        <div className="space-y-2 text-sm">
                          {dashboardData.wells?.hasData && (
                            <div className="flex items-center gap-2 text-green-600">
                              <Activity className="w-3 h-3" />
                              Wells data available ({dashboardData.wells.totalWells} records)
                            </div>
                          )}
                          {dashboardData.checks?.hasData && (
                            <div className="flex items-center gap-2 text-green-600">
                              <DollarSign className="w-3 h-3" />
                              Check data available ({dashboardData.checks.totalChecks.toLocaleString()} records)
                            </div>
                          )}
                          {dashboardData.financials?.hasIncomeData && (
                            <div className="flex items-center gap-2 text-green-600">
                              <TrendingUp className="w-3 h-3" />
                              Income data available ({dashboardData.financials.incomeRecords} records)
                            </div>
                          )}
                          {dashboardData.financials?.hasExpenseData && (
                            <div className="flex items-center gap-2 text-green-600">
                              <BarChart3 className="w-3 h-3" />
                              Expense data available ({dashboardData.financials.expenseRecords} records)
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Charts */}
              <div className="grid gap-4 md:grid-cols-2">
                <Card>
                  <CardHeader>
                    <CardTitle>Revenue vs Expenses</CardTitle>
                    <CardDescription>Monthly comparison over the last 6 months</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <RevenueChart />
                  </CardContent>
                </Card>
                
                <Card>
                  <CardHeader>
                    <CardTitle>Well Production</CardTitle>
                    <CardDescription>Oil and gas production by well</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <ProductionChart />
                  </CardContent>
                </Card>
              </div>
            </div>
          )}

          {activeView === 'transactions' && (
            <div className="space-y-6">
              <TransactionsTable />
            </div>
          )}

          {activeView === 'analytics' && (
            <div className="space-y-6">
              <Tabs defaultValue="revenue" className="w-full">
                <TabsList>
                  <TabsTrigger value="revenue">Revenue Analysis</TabsTrigger>
                  <TabsTrigger value="production">Production Analysis</TabsTrigger>
                  <TabsTrigger value="expenses">Expense Breakdown</TabsTrigger>
                </TabsList>
                <TabsContent value="revenue" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Revenue Trends</CardTitle>
                      <CardDescription>Detailed revenue analysis</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <RevenueChart />
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="production" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Production Performance</CardTitle>
                      <CardDescription>Well-by-well production metrics</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <ProductionChart />
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="expenses" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Expense Categories</CardTitle>
                      <CardDescription>Breakdown of operational expenses</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        Expense breakdown chart would go here
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>
            </div>
          )}

          {activeView === 'dbf-explorer' && (
            <div className="space-y-6">
              <DBFExplorer />
            </div>
          )}

          {activeView === 'db-maintenance' && (
            <div className="space-y-6">
              <div className="grid gap-4 md:grid-cols-2">
                <Card>
                  <CardHeader>
                    <CardTitle>Database Status</CardTitle>
                    <CardDescription>Current database health and statistics</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div className="flex justify-between">
                        <span>Database Size:</span>
                        <span className="font-medium">245 MB</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Total Records:</span>
                        <span className="font-medium">15,847</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Last Backup:</span>
                        <span className="font-medium">2024-07-30</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardHeader>
                    <CardTitle>Maintenance Tools</CardTitle>
                    <CardDescription>Database maintenance and optimization</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <Button className="w-full">Run Database Backup</Button>
                    <Button variant="outline" className="w-full">Optimize Tables</Button>
                    <Button variant="outline" className="w-full">Check Data Integrity</Button>
                  </CardContent>
                </Card>
              </div>
            </div>
          )}

          {activeView === 'state-reporting' && (
            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>West Virginia State Reporting</CardTitle>
                  <CardDescription>Generate and submit required state reports</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid gap-4 md:grid-cols-2">
                    <Button className="h-auto p-4 flex flex-col items-start">
                      <div className="font-semibold">Monthly Production Report</div>
                      <div className="text-sm text-muted-foreground">Due: August 15, 2024</div>
                    </Button>
                    <Button variant="outline" className="h-auto p-4 flex flex-col items-start">
                      <div className="font-semibold">Annual Tax Filing</div>
                      <div className="text-sm text-muted-foreground">Due: March 31, 2025</div>
                    </Button>
                    <Button variant="outline" className="h-auto p-4 flex flex-col items-start">
                      <div className="font-semibold">Environmental Compliance</div>
                      <div className="text-sm text-muted-foreground">Due: September 1, 2024</div>
                    </Button>
                    <Button variant="outline" className="h-auto p-4 flex flex-col items-start">
                      <div className="font-semibold">Royalty Statement</div>
                      <div className="text-sm text-muted-foreground">Due: August 10, 2024</div>
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          {activeView === 'settings' && (
            <div className="space-y-6">
              <Tabs defaultValue="users" className="w-full">
                <TabsList>
                  <TabsTrigger value="users">User Management</TabsTrigger>
                  <TabsTrigger value="system">System Settings</TabsTrigger>
                  <TabsTrigger value="security">Security</TabsTrigger>
                </TabsList>
                <TabsContent value="users" className="space-y-4">
                  <UserManagement currentUser={currentUser} />
                </TabsContent>
                <TabsContent value="system" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>System Configuration</CardTitle>
                      <CardDescription>Configure system-wide settings</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        System settings configuration would go here
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="security" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Security Settings</CardTitle>
                      <CardDescription>Manage security policies and configurations</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        Security settings would go here
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>
            </div>
          )}

          {/* Utilities Views */}
          {activeView === 'calculator' && (
            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Financial Calculator</CardTitle>
                  <CardDescription>Perform various financial calculations</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid gap-4 md:grid-cols-2">
                    <Button variant="outline" className="h-auto p-4">
                      <div className="text-left">
                        <div className="font-semibold">Interest Calculator</div>
                        <div className="text-sm text-muted-foreground">Calculate simple and compound interest</div>
                      </div>
                    </Button>
                    <Button variant="outline" className="h-auto p-4">
                      <div className="text-left">
                        <div className="font-semibold">Royalty Calculator</div>
                        <div className="text-sm text-muted-foreground">Calculate oil & gas royalties</div>
                      </div>
                    </Button>
                    <Button variant="outline" className="h-auto p-4">
                      <div className="text-left">
                        <div className="font-semibold">Tax Calculator</div>
                        <div className="text-sm text-muted-foreground">Estimate taxes and deductions</div>
                      </div>
                    </Button>
                    <Button variant="outline" className="h-auto p-4">
                      <div className="text-left">
                        <div className="font-semibold">Unit Converter</div>
                        <div className="text-sm text-muted-foreground">Convert between measurement units</div>
                      </div>
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          {activeView === 'data-import' && (
            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Data Import Wizard</CardTitle>
                  <CardDescription>Import data from various sources into your system</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="border-2 border-dashed border-muted-foreground/25 rounded-lg p-8 text-center">
                      <Upload className="mx-auto h-12 w-12 text-muted-foreground" />
                      <p className="mt-2 text-sm text-muted-foreground">
                        Drag and drop files here, or click to browse
                      </p>
                      <p className="text-xs text-muted-foreground mt-1">
                        Supported formats: CSV, Excel, JSON, DBF
                      </p>
                      <Button className="mt-4">Select Files</Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          {activeView === 'data-export' && (
            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Data Export Manager</CardTitle>
                  <CardDescription>Export your data in various formats</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid gap-4">
                    <div className="flex items-center justify-between p-4 border rounded-lg">
                      <div>
                        <h4 className="font-semibold">Export All Financial Data</h4>
                        <p className="text-sm text-muted-foreground">Complete export of all financial records</p>
                      </div>
                      <Button>Export</Button>
                    </div>
                    <div className="flex items-center justify-between p-4 border rounded-lg">
                      <div>
                        <h4 className="font-semibold">Export Well Production Data</h4>
                        <p className="text-sm text-muted-foreground">Oil and gas production records</p>
                      </div>
                      <Button>Export</Button>
                    </div>
                    <div className="flex items-center justify-between p-4 border rounded-lg">
                      <div>
                        <h4 className="font-semibold">Export Tax Reports</h4>
                        <p className="text-sm text-muted-foreground">Tax-ready financial summaries</p>
                      </div>
                      <Button>Export</Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          {activeView === 'backup' && (
            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Backup & Restore</CardTitle>
                  <CardDescription>Manage system backups and restore points</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="p-4 border rounded-lg bg-muted/50">
                      <div className="flex items-center justify-between">
                        <div>
                          <h4 className="font-semibold">Last Backup</h4>
                          <p className="text-sm text-muted-foreground">July 30, 2024 at 2:30 PM</p>
                        </div>
                        <Badge variant="outline" className="text-green-600">Successful</Badge>
                      </div>
                    </div>
                    <div className="grid gap-3">
                      <Button className="w-full">
                        <Archive className="w-4 h-4 mr-2" />
                        Create New Backup
                      </Button>
                      <Button variant="outline" className="w-full">
                        Schedule Automatic Backups
                      </Button>
                      <Button variant="outline" className="w-full">
                        Restore from Backup
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          {activeView === 'audit-log' && (
            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>System Audit Log</CardTitle>
                  <CardDescription>Track all system activities and changes</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex gap-4">
                      <Input placeholder="Search audit logs..." className="flex-1" />
                      <Button variant="outline">Filter</Button>
                      <Button variant="outline">Export</Button>
                    </div>
                    <div className="rounded-md border">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Timestamp</TableHead>
                            <TableHead>User</TableHead>
                            <TableHead>Action</TableHead>
                            <TableHead>Resource</TableHead>
                            <TableHead>Details</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          <TableRow>
                            <TableCell colSpan={5} className="text-center text-muted-foreground py-8">
                              Audit log entries would appear here
                            </TableCell>
                          </TableRow>
                        </TableBody>
                      </Table>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          {activeView === 'scheduler' && (
            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Task Scheduler</CardTitle>
                  <CardDescription>Schedule automated tasks and reports</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex justify-end">
                      <Button>
                        <Calendar className="w-4 h-4 mr-2" />
                        New Scheduled Task
                      </Button>
                    </div>
                    <div className="space-y-3">
                      <div className="p-4 border rounded-lg">
                        <div className="flex items-center justify-between">
                          <div>
                            <h4 className="font-semibold">Daily Database Backup</h4>
                            <p className="text-sm text-muted-foreground">Runs every day at 2:00 AM</p>
                          </div>
                          <div className="flex gap-2">
                            <Badge variant="outline" className="text-green-600">Active</Badge>
                            <Button size="sm" variant="outline">Edit</Button>
                          </div>
                        </div>
                      </div>
                      <div className="p-4 border rounded-lg">
                        <div className="flex items-center justify-between">
                          <div>
                            <h4 className="font-semibold">Monthly Production Report</h4>
                            <p className="text-sm text-muted-foreground">Runs on the 1st of each month</p>
                          </div>
                          <div className="flex gap-2">
                            <Badge variant="outline" className="text-green-600">Active</Badge>
                            <Button size="sm" variant="outline">Edit</Button>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}

          {activeView === 'data-tools' && (
            <div className="space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Data Tools</CardTitle>
                  <CardDescription>Utilities for data validation and transformation</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid gap-4 md:grid-cols-2">
                    <Button variant="outline" className="h-auto p-4">
                      <div className="text-left">
                        <div className="font-semibold">Data Validator</div>
                        <div className="text-sm text-muted-foreground">Check data integrity and consistency</div>
                      </div>
                    </Button>
                    <Button variant="outline" className="h-auto p-4">
                      <div className="text-left">
                        <div className="font-semibold">Duplicate Finder</div>
                        <div className="text-sm text-muted-foreground">Find and merge duplicate records</div>
                      </div>
                    </Button>
                    <Button variant="outline" className="h-auto p-4">
                      <div className="text-left">
                        <div className="font-semibold">Data Cleaner</div>
                        <div className="text-sm text-muted-foreground">Clean and standardize data formats</div>
                      </div>
                    </Button>
                    <Button variant="outline" className="h-auto p-4">
                      <div className="text-left">
                        <div className="font-semibold">Batch Operations</div>
                        <div className="text-sm text-muted-foreground">Perform bulk updates on records</div>
                      </div>
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}
        </main>
      </div>
    </div>
  )
}

export default App
