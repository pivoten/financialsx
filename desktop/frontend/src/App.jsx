import { useState, useEffect } from 'react'
import { Login, Register, GetCompanies, ValidateSession } from '../wailsjs/go/main/App'
import { Button } from './components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './components/ui/card'
import { Input } from './components/ui/input'
import { Sidebar, SidebarHeader, SidebarContent, SidebarNav, SidebarNavItem, SidebarNavGroup } from './components/ui/sidebar'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './components/ui/tabs'
import { RevenueChart } from './components/charts/RevenueChart'
import { ProductionChart } from './components/charts/ProductionChart'
import { TransactionsTable } from './components/tables/TransactionsTable'
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
  Activity
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
  return <AdvancedDashboard currentUser={currentUser} onLogout={handleLogout} />
}

// Advanced Dashboard Component with Sidebar and Multiple Views
function AdvancedDashboard({ currentUser, onLogout }) {
  const [activeView, setActiveView] = useState('dashboard')

  const stats = [
    { title: 'Monthly Revenue', value: '$67,000', change: '+12.5%', icon: DollarSign, trend: 'up' },
    { title: 'Active Wells', value: '24', change: '+2', icon: Activity, trend: 'up' },
    { title: 'Oil Production', value: '1,240 bbl', change: '+8.2%', icon: TrendingUp, trend: 'up' },
    { title: 'Gas Production', value: '890 mcf', change: '-2.1%', icon: BarChart3, trend: 'down' },
  ]

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <Sidebar className="hidden lg:flex">
        <SidebarHeader>
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
          
          <SidebarNavGroup title="Account">
            <SidebarNavItem href="#">
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
        <header className="border-b bg-card shadow-sm px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-bold text-foreground">
                {activeView === 'dashboard' && 'Dashboard'}
                {activeView === 'transactions' && 'Transactions'}
                {activeView === 'analytics' && 'Analytics'}
                {activeView === 'db-maintenance' && 'Database Maintenance'}
                {activeView === 'state-reporting' && 'State Reporting'}
              </h2>
              <p className="text-sm text-muted-foreground">
                {activeView === 'dashboard' && 'Overview of your financial data'}
                {activeView === 'transactions' && 'View and manage all transactions'}
                {activeView === 'analytics' && 'Detailed charts and reports'}
                {activeView === 'db-maintenance' && 'Database management tools'}
                {activeView === 'state-reporting' && 'West Virginia reporting'}
              </p>
            </div>
            <div className="flex items-center space-x-4">
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
        <main className="flex-1 overflow-auto p-6">
          {activeView === 'dashboard' && (
            <div className="space-y-6">
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
                          {stat.change} from last month
                        </p>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>

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
        </main>
      </div>
    </div>
  )
}

export default App
