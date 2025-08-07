import { useState, useEffect, useRef } from 'react'
import { Login, Register, GetCompanies, ValidateSession, GetDashboardData } from '../wailsjs/go/main/App'
import { supabase, isSupabaseConfigured, signIn, signUp } from './lib/supabase'
import { Button } from './components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './components/ui/card'
import { Input } from './components/ui/input'
import { Badge } from './components/ui/badge'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './components/ui/table'
import { Sidebar, SidebarHeader, SidebarContent, SidebarNav, SidebarNavItem, SidebarNavGroup } from './components/ui/sidebar'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './components/ui/tabs'
import { RevenueChart } from './components/charts/RevenueChart'
import { ProductionChart } from './components/charts/ProductionChart'
import { TransactionsTable } from './components/tables/TransactionsTable'
import { DBFExplorer } from './components/DBFExplorer'
import { UserManagement } from './components/UserManagement'
import { StateReportsSection } from './components/StateReportsSection'
import { BankingSection } from './components/BankingSection'
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
  Calculator,
  FileSearch,
  Archive,
  Download,
  Upload,
  Copy,
  Calendar,
  Wrench,
  Menu,
  ChevronLeft
} from 'lucide-react'
import './globals.css'

function App() {
  console.log('App component rendering...')
  
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [currentUser, setCurrentUser] = useState(null)
  const [companies, setCompanies] = useState([])
  const [showRegister, setShowRegister] = useState(false)
  const [loading, setLoading] = useState(true)
  const [selectedCompany, setSelectedCompany] = useState('')
  const [companySelected, setCompanySelected] = useState(false)

  // Form states
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Log Supabase configuration on mount
  useEffect(() => {
    console.log('App mounted, checking Supabase configuration...')
    console.log('Supabase configured?', isSupabaseConfigured())
    console.log('Supabase client exists?', !!supabase)
    if (supabase) {
      console.log('Supabase URL from client:', supabase.supabaseUrl)
    }
  }, [])

  useEffect(() => {
    checkSession()
  }, [])

  // Load companies after authentication
  useEffect(() => {
    if (isAuthenticated && !companySelected && !loading) {
      loadCompanies()
    }
  }, [isAuthenticated, companySelected, loading])

  const checkSession = async () => {
    const token = localStorage.getItem('session_token')
    const authType = localStorage.getItem('auth_type')
    const savedCompany = localStorage.getItem('company_name')
    
    console.log('checkSession: Found stored session?', { hasToken: !!token, authType })
    
    if (token) {
      try {
        if (authType === 'supabase' && isSupabaseConfigured()) {
          // Validate Supabase session
          const { data: { user }, error } = await supabase.auth.getUser(token)
          if (error) throw error
          
          console.log('checkSession: Supabase session valid, user:', user)
          setCurrentUser({
            ...user,
            email: user.email,
            username: user.email
          })
          setIsAuthenticated(true)
          
          // Check if user had a company selected previously
          if (savedCompany) {
            setSelectedCompany(savedCompany)
            setCompanySelected(true)
          }
        } else if (authType === 'local' && savedCompany) {
          // Validate local SQLite session (requires company)
          console.log('checkSession: Validating local session for company:', savedCompany)
          const user = await ValidateSession(token, savedCompany)
          console.log('checkSession: Session valid, user:', user)
          setCurrentUser(user)
          setIsAuthenticated(true)
          setSelectedCompany(savedCompany)
          setCompanySelected(true)
        }
      } catch (error) {
        console.log('checkSession: Session validation failed:', error)
        localStorage.removeItem('session_token')
        localStorage.removeItem('company_name')
        localStorage.removeItem('auth_type')
      }
    } else {
      console.log('checkSession: No stored session found')
    }
    setLoading(false)
  }

  const loadCompanies = async () => {
    try {
      // Check if Wails runtime is available
      if (!window.go || !window.go.main || !window.go.main.App) {
        console.log('Wails runtime not available')
        // For testing without Wails, use mock data
        setCompanies([
          { name: 'apollo', path: 'datafiles/apollo' },
          { name: 'cantrellenergy', path: 'datafiles/cantrellenergy' },
          { name: 'pivotenoperating', path: 'datafiles/pivotenoperating' }
        ])
        return
      }
      
      const companiesList = await GetCompanies()
      setCompanies(companiesList || [])
    } catch (error) {
      console.error('Failed to load companies:', error)
      // Set some default companies for testing
      setCompanies([
        { name: 'apollo', path: 'datafiles/apollo' },
        { name: 'cantrellenergy', path: 'datafiles/cantrellenergy' },
        { name: 'pivotenoperating', path: 'datafiles/pivotenoperating' }
      ])
    }
  }

  const handleLogin = async (e) => {
    e.preventDefault()
    setError('')
    
    if (!username || !password) {
      setError('Please enter your email and password')
      return
    }

    setIsSubmitting(true)
    try {
      console.log('Login attempt:', { username, supabaseConfigured: isSupabaseConfigured() })
      
      if (isSupabaseConfigured()) {
        // Use Supabase authentication
        console.log('Attempting Supabase login with:', { email: username, passwordLength: password.length })
        const { data, error } = await signIn(username, password)
        
        console.log('Supabase response:', { data, error })
        
        if (error) {
          console.error('Supabase error:', error)
          throw error
        }
        
        // Store session info
        localStorage.setItem('session_token', data.session.access_token)
        localStorage.setItem('auth_type', 'supabase')
        
        // Set user info
        setCurrentUser({
          ...data.user,
          email: data.user.email,
          username: data.user.email
        })
        setIsAuthenticated(true)
      } else {
        // For local auth, we need to select a company first
        setError('Please use Supabase authentication or select a company for local auth')
      }
    } catch (error) {
      console.error('Login error details:', error)
      setError(error.message || 'Invalid login credentials')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleRegister = async (e) => {
    e.preventDefault()
    setError('')
    
    if (!username || !password || !email) {
      setError('Please fill in all fields')
      return
    }

    setIsSubmitting(true)
    try {
      if (isSupabaseConfigured()) {
        // Use Supabase registration
        const { data, error } = await signUp(email, password, {
          username: username
        })
        if (error) throw error
        
        // Auto-login after registration
        const { data: loginData, error: loginError } = await signIn(email, password)
        if (loginError) throw loginError
        
        localStorage.setItem('session_token', loginData.session.access_token)
        localStorage.setItem('auth_type', 'supabase')
        
        setCurrentUser({
          ...loginData.user,
          email: loginData.user.email,
          username: loginData.user.email
        })
        setIsAuthenticated(true)
      } else {
        setError('Please use Supabase authentication for registration')
      }
    } catch (error) {
      setError(error.message || 'Registration failed')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleCompanySelect = async (company) => {
    setError('')
    setIsSubmitting(true)
    
    try {
      // Store selected company
      localStorage.setItem('company_name', company)
      setSelectedCompany(company)
      
      // For Supabase auth, we just need to set the company context
      // The actual data loading will happen in the dashboard
      setCompanySelected(true)
      
      // Update current user with company info
      setCurrentUser(prev => ({
        ...prev,
        company_name: company
      }))
    } catch (error) {
      setError('Failed to select company')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleLogout = () => {
    localStorage.removeItem('session_token')
    localStorage.removeItem('company_name')
    localStorage.removeItem('auth_type')
    setCurrentUser(null)
    setIsAuthenticated(false)
    setCompanySelected(false)
    setUsername('')
    setPassword('')
    setEmail('')
    setSelectedCompany('')
    setError('')
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-lg">Loading...</div>
      </div>
    )
  }

  // Show company selector if authenticated but no company selected
  if (isAuthenticated && !companySelected) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4" style={{ background: 'linear-gradient(135deg, #F4EBE8 0%, #E7ECEF 50%, #F4EBE8 100%)' }}>
        {/* Subtle animated background using brand colors */}
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute -top-40 -right-40 w-80 h-80 rounded-full blur-3xl animate-blob" style={{ backgroundColor: 'rgba(245, 152, 30, 0.1)' }}></div>
          <div className="absolute -bottom-40 -left-40 w-80 h-80 rounded-full blur-3xl animate-blob animation-delay-2000" style={{ backgroundColor: 'rgba(44, 68, 113, 0.1)' }}></div>
          <div className="absolute top-40 left-40 w-80 h-80 rounded-full blur-3xl animate-blob animation-delay-4000" style={{ backgroundColor: 'rgba(110, 142, 201, 0.1)' }}></div>
        </div>

        <Card className="w-full max-w-2xl relative backdrop-blur-sm bg-white/90 shadow-2xl border border-white/50">
          <CardHeader className="text-center space-y-1 pb-6">
            <div className="mx-auto mb-4">
              <img 
                src="/src/assets/pivoten-logo.png" 
                alt="Pivoten Logo" 
                className="w-24 h-24 object-contain drop-shadow-md"
              />
            </div>
            <CardTitle className="text-2xl font-bold" style={{ color: '#2C4471' }}>
              Select Company
            </CardTitle>
            <CardDescription className="text-gray-600">
              Welcome back, {currentUser?.email}! Please select a company to continue.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {companies.length === 0 ? (
                <div className="text-center py-8">
                  <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-600 mx-auto mb-4"></div>
                  <p className="text-gray-600">Loading companies...</p>
                </div>
              ) : (
                <div className="grid gap-3">
                  {companies.map((company) => (
                    <button
                      key={company.name}
                      onClick={() => handleCompanySelect(company.name)}
                      disabled={isSubmitting}
                      className="w-full p-4 bg-white border-2 border-gray-200 rounded-lg hover:shadow-md transition-all text-left group"
                      onMouseEnter={(e) => {
                        e.currentTarget.style.borderColor = '#F5981E'
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.borderColor = ''
                      }}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                          <div className="w-12 h-12 rounded-lg flex items-center justify-center transition-colors"
                            style={{ backgroundColor: 'rgba(110, 142, 201, 0.1)' }}>
                            <Home className="w-6 h-6" style={{ color: '#6E8EC9' }} />
                          </div>
                          <div>
                            <h3 className="font-semibold" style={{ color: '#2C4471' }}>{company.name}</h3>
                            <p className="text-sm text-gray-500">
                              {company.path || 'Company database'}
                            </p>
                          </div>
                        </div>
                        <ChevronLeft className="w-5 h-5 text-gray-400 rotate-180 transition-colors group-hover:text-[#F5981E]" />
                      </div>
                    </button>
                  ))}
                </div>
              )}

              {error && (
                <div className="flex items-start gap-2 text-sm text-red-600 bg-red-50 p-3 rounded-lg border border-red-200">
                  <div className="mt-0.5">⚠️</div>
                  <div>{error}</div>
                </div>
              )}

              <div className="pt-4 border-t">
                <button
                  onClick={handleLogout}
                  className="w-full text-sm text-gray-600 hover:text-gray-900 transition-colors"
                >
                  Sign out and use a different account
                </button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4" style={{ background: 'linear-gradient(135deg, #F4EBE8 0%, #E7ECEF 50%, #F4EBE8 100%)' }}>
        {/* Subtle animated background using brand colors */}
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute -top-40 -right-40 w-80 h-80 rounded-full blur-3xl animate-blob" style={{ backgroundColor: 'rgba(245, 152, 30, 0.1)' }}></div>
          <div className="absolute -bottom-40 -left-40 w-80 h-80 rounded-full blur-3xl animate-blob animation-delay-2000" style={{ backgroundColor: 'rgba(44, 68, 113, 0.1)' }}></div>
          <div className="absolute top-40 left-40 w-80 h-80 rounded-full blur-3xl animate-blob animation-delay-4000" style={{ backgroundColor: 'rgba(110, 142, 201, 0.1)' }}></div>
        </div>

        <Card className="w-full max-w-md relative backdrop-blur-sm bg-white/90 shadow-2xl border border-white/50">
          <CardHeader className="text-center space-y-1 pb-8">
            <div className="mx-auto mb-6">
              <img 
                src="/src/assets/pivoten-logo.png" 
                alt="Pivoten Logo" 
                className="w-28 h-28 object-contain drop-shadow-md"
              />
            </div>
            <CardDescription className="text-gray-600 text-base">
              {showRegister ? 'Create your account to get started' : 'Sign in to your account'}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={(e) => {
              console.log('Form submitted!', { showRegister, username, password })
              return showRegister ? handleRegister(e) : handleLogin(e)
            }} className="space-y-5">
              {/* Username/Email field with icon */}
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-700 flex items-center gap-2">
                  <Users className="w-4 h-4" style={{ color: '#F5981E' }} />
                  {showRegister ? 'Email' : 'Email or Username'}
                </label>
                <div className="relative">
                  <Input
                    type={showRegister ? "email" : "text"}
                    value={showRegister ? email : username}
                    onChange={(e) => {
                      const value = e.target.value
                      console.log('Input changed:', { field: showRegister ? 'email' : 'username', value })
                      showRegister ? setEmail(value) : setUsername(value)
                    }}
                    placeholder={showRegister ? "john@example.com" : "Enter your email or username"}
                    className="pl-10 h-11 border-gray-200 transition-all"
                    style={{ '--tw-ring-color': 'rgba(245, 152, 30, 0.2)' }}
                    onFocus={(e) => {
                      e.target.style.borderColor = '#F5981E'
                      e.target.style.boxShadow = '0 0 0 3px rgba(245, 152, 30, 0.1)'
                    }}
                    onBlur={(e) => {
                      e.target.style.borderColor = ''
                      e.target.style.boxShadow = ''
                    }}
                  />
                  <div className="absolute left-3 top-3.5" style={{ color: '#6E8EC9' }}>
                    <Activity className="w-4 h-4" />
                  </div>
                </div>
              </div>
              
              {/* Username field for registration */}
              {showRegister && (
                <div className="space-y-2">
                  <label className="text-sm font-medium text-gray-700 flex items-center gap-2">
                    <Users className="w-4 h-4" />
                    Username
                  </label>
                  <div className="relative">
                    <Input
                      type="text"
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      placeholder="Choose a username"
                      className="pl-10 h-11 border-gray-200 transition-all"
                    style={{ '--tw-ring-color': 'rgba(245, 152, 30, 0.2)' }}
                    onFocus={(e) => {
                      e.target.style.borderColor = '#F5981E'
                      e.target.style.boxShadow = '0 0 0 3px rgba(245, 152, 30, 0.1)'
                    }}
                    onBlur={(e) => {
                      e.target.style.borderColor = ''
                      e.target.style.boxShadow = ''
                    }}
                    />
                    <div className="absolute left-3 top-3.5" style={{ color: '#6E8EC9' }}>
                      <Users className="w-4 h-4" />
                    </div>
                  </div>
                </div>
              )}
              
              {/* Password field with icon */}
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-700 flex items-center gap-2">
                  <FileText className="w-4 h-4" style={{ color: '#F5981E' }} />
                  Password
                </label>
                <div className="relative">
                  <Input
                    type="password"
                    value={password}
                    onChange={(e) => {
                      const value = e.target.value
                      console.log('Password changed:', { length: value.length })
                      setPassword(value)
                    }}
                    placeholder="••••••••"
                    className="pl-10 h-11 border-gray-200 transition-all"
                    style={{ '--tw-ring-color': 'rgba(245, 152, 30, 0.2)' }}
                    onFocus={(e) => {
                      e.target.style.borderColor = '#F5981E'
                      e.target.style.boxShadow = '0 0 0 3px rgba(245, 152, 30, 0.1)'
                    }}
                    onBlur={(e) => {
                      e.target.style.borderColor = ''
                      e.target.style.boxShadow = ''
                    }}
                  />
                  <div className="absolute left-3 top-3.5" style={{ color: '#6E8EC9' }}>
                    <FileText className="w-4 h-4" />
                  </div>
                </div>
                {showRegister && password && (
                  <div className="flex gap-1 mt-2">
                    <div className={`h-1 flex-1 rounded-full transition-colors`}
                      style={{ backgroundColor: password.length >= 8 ? '#1E9B5F' : '#E7ECEF' }}></div>
                    <div className={`h-1 flex-1 rounded-full transition-colors`}
                      style={{ backgroundColor: password.length >= 12 ? '#1E9B5F' : '#E7ECEF' }}></div>
                    <div className={`h-1 flex-1 rounded-full transition-colors`}
                      style={{ backgroundColor: password.length >= 16 ? '#1E9B5F' : '#E7ECEF' }}></div>
                  </div>
                )}
              </div>


              {/* Error message with better styling */}
              {error && (
                <div className="flex items-start gap-2 text-sm text-red-600 bg-red-50 p-3 rounded-lg border border-red-200 animate-pulse">
                  <div className="mt-0.5">⚠️</div>
                  <div>{error}</div>
                </div>
              )}

              {/* Submit button */}
              <Button 
                type="submit" 
                className="w-full h-11 text-white font-medium shadow-lg hover:shadow-xl transition-all transform hover:scale-[1.02]"
                style={{ 
                  background: 'linear-gradient(135deg, #F5981E 0%, #F5981E 100%)',
                  transition: 'all 0.2s'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'linear-gradient(135deg, #2C4471 0%, #2C4471 100%)'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'linear-gradient(135deg, #F5981E 0%, #F5981E 100%)'
                }}
                disabled={isSubmitting}
                onClick={() => console.log('Button clicked!')}
              >
                {isSubmitting ? (
                  <div className="flex items-center gap-2">
                    <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                    <span className="animate-pulse">{showRegister ? 'Creating account...' : 'Signing in...'}</span>
                  </div>
                ) : (
                  showRegister ? 'Create Account' : 'Sign In'
                )}
              </Button>

              {/* Auth type indicator */}
              <div className="text-center pt-2">
                <p className="text-xs text-gray-500">
                  {isSupabaseConfigured() ? '🔒 Secure cloud authentication' : '💾 Local authentication'}
                </p>
              </div>

              {/* Toggle between login and register */}
              <div className="text-center pt-2">
                <button
                  type="button"
                  onClick={() => {
                    setShowRegister(!showRegister)
                    setError('')
                  }}
                  className="text-sm font-medium hover:underline transition-colors"
                  style={{ color: '#F5981E' }}
                  onMouseEnter={(e) => e.currentTarget.style.color = '#2C4471'}
                  onMouseLeave={(e) => e.currentTarget.style.color = '#F5981E'}
                >
                  {showRegister 
                    ? 'Already have an account? Sign in' 
                    : "Don't have an account? Create one"}
                </button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    )
  }

  // Show dashboard only when authenticated AND company is selected
  if (isAuthenticated && companySelected) {
    return (
      <ThemeProvider defaultTheme="light" defaultColorScheme="pivoten">
        <AdvancedDashboard currentUser={currentUser} onLogout={handleLogout} />
      </ThemeProvider>
    )
  }

  // Fallback (should not reach here)
  return null
}

// Advanced Dashboard Component with Sidebar and Multiple Views
function AdvancedDashboard({ currentUser, onLogout }) {
  const [activeView, setActiveView] = useState('dashboard')
  const [dashboardData, setDashboardData] = useState(null)
  const [loadingDashboard, setLoadingDashboard] = useState(true)
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)
  const lastLoadedCompany = useRef(null)

  // Load dashboard data when company changes (not when user object changes)
  useEffect(() => {
    console.log('🔄 UPDATED Dashboard useEffect triggered')
    console.log('  currentUser exists:', !!currentUser)
    console.log('  company_name:', currentUser?.company_name)
    console.log('  lastLoadedCompany:', lastLoadedCompany.current)
    
    // Only load if company actually changed
    if (currentUser?.company_name && currentUser.company_name !== lastLoadedCompany.current) {
      console.log('  Company changed, loading dashboard data')
      lastLoadedCompany.current = currentUser.company_name
      loadDashboardData()
    } else {
      console.log('  No company change or already loaded, skipping')
    }
  }, [currentUser?.company_name])

  const loadDashboardData = async () => {
    if (!currentUser?.company_name) return
    
    // Skip if we already have data for this company and it's not stale
    if (dashboardData && dashboardData.company === currentUser.company_name) {
      console.log('Dashboard data already loaded for company:', currentUser.company_name)
      return
    }
    
    console.log('Loading dashboard data for company:', currentUser.company_name)
    setLoadingDashboard(true)
    try {
      // Check if Wails runtime is available
      if (!window.go || !window.go.main || !window.go.main.App) {
        console.log('⚠️ Wails runtime not available - Run "wails dev" to access backend functions')
        // Set minimal dashboard data for browser development
        setDashboardData({
          company: currentUser.company_name,
          totalRevenue: 0,
          totalExpenses: 0,
          netIncome: 0,
          wellCount: 0,
          recentTransactions: [],
          message: 'Backend not available - Run "wails dev" for full functionality'
        })
        return
      }
      
      const data = await GetDashboardData(currentUser.company_name)
      console.log('Dashboard data loaded:', data)
      setDashboardData(data)
    } catch (error) {
      console.error('Failed to load dashboard data:', error)
    } finally {
      setLoadingDashboard(false)
    }
  }

  // Generate well type stats from WELLS.DBF
  const generateStats = () => {
    if (!dashboardData) {
      return [
        { title: 'Loading...', value: '-', change: '-', icon: Activity, trend: 'up' },
        { title: 'Loading...', value: '-', change: '-', icon: Activity, trend: 'up' },
        { title: 'Loading...', value: '-', change: '-', icon: Activity, trend: 'up' },
        { title: 'Loading...', value: '-', change: '-', icon: Activity, trend: 'up' },
      ]
    }

    const stats = []
    
    // Show well types from WELLS.DBF
    if (dashboardData.wellTypes && dashboardData.wellTypes.length > 0) {
      dashboardData.wellTypes.forEach(wellType => {
        stats.push({
          title: `${wellType.type} Wells`,
          value: wellType.count.toString(),
          change: `${wellType.type.toLowerCase()} wells active`,
          icon: Activity,
          trend: 'up'
        })
      })
    }

    // Only show actual well status data - no fallback cards
    return stats
  }

  const stats = generateStats()

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <div 
        className="hidden lg:flex flex-col border-r transition-all duration-300 bg-card"
        style={{ width: isSidebarCollapsed ? '4rem' : '16rem' }}
      >
        <div className="border-b px-4 py-4 flex items-center justify-between h-16">
          {!isSidebarCollapsed && (
            <h1 className="text-xl font-bold text-primary">FinancialsX</h1>
          )}
          <button
            onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
            className={`p-2 hover:bg-accent rounded-md transition-colors ${isSidebarCollapsed ? 'mx-auto' : 'ml-auto'}`}
          >
            {isSidebarCollapsed ? <Menu className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
          </button>
        </div>
        <div className="flex-1 flex flex-col overflow-y-auto">
          <SidebarNav>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'dashboard'}
              onClick={() => setActiveView('dashboard')}
              className="flex items-center gap-3"
              title={isSidebarCollapsed ? "Dashboard" : ""}
            >
              <Home className="w-4 h-4 flex-shrink-0" />
              {!isSidebarCollapsed && <span>Dashboard</span>}
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'operations'}
              onClick={() => setActiveView('operations')}
              className="flex items-center gap-3"
              title={isSidebarCollapsed ? "Operations" : ""}
            >
              <Activity className="w-4 h-4 flex-shrink-0" />
              {!isSidebarCollapsed && <span>Operations</span>}
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'financials'}
              onClick={() => setActiveView('financials')}
              className="flex items-center gap-3"
              title={isSidebarCollapsed ? "Financials" : ""}
            >
              <DollarSign className="w-4 h-4 flex-shrink-0" />
              {!isSidebarCollapsed && <span>Financials</span>}
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'data'}
              onClick={() => setActiveView('data')}
              className="flex items-center gap-3"
              title={isSidebarCollapsed ? "Data Management" : ""}
            >
              <Database className="w-4 h-4 flex-shrink-0" />
              {!isSidebarCollapsed && <span>Data Management</span>}
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'reporting'}
              onClick={() => setActiveView('reporting')}
              className="flex items-center gap-3"
              title={isSidebarCollapsed ? "Reporting" : ""}
            >
              <FileText className="w-4 h-4 flex-shrink-0" />
              {!isSidebarCollapsed && <span>Reporting</span>}
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'utilities'}
              onClick={() => setActiveView('utilities')}
              className="flex items-center gap-3"
              title={isSidebarCollapsed ? "Utilities" : ""}
            >
              <Wrench className="w-4 h-4 flex-shrink-0" />
              {!isSidebarCollapsed && <span>Utilities</span>}
            </SidebarNavItem>
            <SidebarNavItem 
              href="#" 
              active={activeView === 'settings'}
              onClick={() => setActiveView('settings')}
              className="flex items-center gap-3"
              title={isSidebarCollapsed ? "Settings" : ""}
            >
              <Settings className="w-4 h-4 flex-shrink-0" />
              {!isSidebarCollapsed && <span>Settings</span>}
            </SidebarNavItem>
          </SidebarNav>
          <div className="mt-auto p-4 border-t">
            <Button 
              variant="ghost" 
              className={`w-full ${isSidebarCollapsed ? 'justify-center px-2' : 'justify-start'}`}
              onClick={onLogout}
              title={isSidebarCollapsed ? "Logout" : ""}
            >
              <LogOut className="w-4 h-4 flex-shrink-0" />
              {!isSidebarCollapsed && <span className="ml-2">Logout</span>}
            </Button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-semibold tracking-tight">
                {activeView === 'dashboard' && 'Dashboard'}
                {activeView === 'operations' && 'Operations'}
                {activeView === 'financials' && 'Financials'}
                {activeView === 'data' && 'Data Management'}
                {activeView === 'reporting' && 'Reporting'}
                {activeView === 'utilities' && 'Utilities'}
                {activeView === 'settings' && 'Settings'}
              </h2>
              <p className="text-sm text-muted-foreground">
                {activeView === 'dashboard' && 'Overview of your financial data'}
                {activeView === 'operations' && 'Manage wells, production, and field operations'}
                {activeView === 'financials' && 'Financial transactions, analytics, and accounting'}
                {activeView === 'data' && 'Database maintenance and data management'}
                {activeView === 'reporting' && 'Reports, compliance, and documentation'}
                {activeView === 'utilities' && 'Tools, calculators, and system utilities'}
                {activeView === 'settings' && 'System configuration and user management'}
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

              {/* Data Summary section removed per user request */}

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

          {activeView === 'operations' && (
            <div className="space-y-6">
              <Tabs defaultValue="wells" className="w-full">
                <TabsList>
                  <TabsTrigger value="wells">Wells</TabsTrigger>
                  <TabsTrigger value="production">Production</TabsTrigger>
                  <TabsTrigger value="field-ops">Field Operations</TabsTrigger>
                  <TabsTrigger value="maintenance">Maintenance</TabsTrigger>
                </TabsList>
                <TabsContent value="wells" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Well Management</CardTitle>
                      <CardDescription>Manage well information and status</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        Well management interface would go here
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="production" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Production Tracking</CardTitle>
                      <CardDescription>Monitor oil and gas production</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <ProductionChart />
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="field-ops" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Field Operations</CardTitle>
                      <CardDescription>Manage field activities and schedules</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        Field operations management would go here
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="maintenance" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Equipment Maintenance</CardTitle>
                      <CardDescription>Track equipment maintenance schedules</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        Maintenance tracking would go here
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>
            </div>
          )}

          {activeView === 'financials' && (
            <div className="space-y-6">
              <Tabs defaultValue="transactions" className="w-full">
                <TabsList>
                  <TabsTrigger value="transactions">Transactions</TabsTrigger>
                  <TabsTrigger value="revenue">Revenue</TabsTrigger>
                  <TabsTrigger value="expenses">Expenses</TabsTrigger>
                  <TabsTrigger value="analytics">Analytics</TabsTrigger>
                  <TabsTrigger value="banking">Banking</TabsTrigger>
                  <TabsTrigger value="accounting">Accounting</TabsTrigger>
                </TabsList>
                <TabsContent value="transactions" className="space-y-4">
                  <TransactionsTable />
                </TabsContent>
                <TabsContent value="revenue" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Revenue Analysis</CardTitle>
                      <CardDescription>Track and analyze revenue streams</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <RevenueChart />
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="expenses" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Expense Management</CardTitle>
                      <CardDescription>Monitor and categorize expenses</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        Expense tracking interface would go here
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="analytics" className="space-y-4">
                  <div className="grid gap-4 md:grid-cols-2">
                    <Card>
                      <CardHeader>
                        <CardTitle>Financial Trends</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <RevenueChart />
                      </CardContent>
                    </Card>
                    <Card>
                      <CardHeader>
                        <CardTitle>Production vs Revenue</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <ProductionChart />
                      </CardContent>
                    </Card>
                  </div>
                </TabsContent>
                <TabsContent value="banking" className="space-y-4">
                  <BankingSection companyName={currentUser?.company_name} currentUser={currentUser} />
                </TabsContent>
                <TabsContent value="accounting" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Accounting Tools</CardTitle>
                      <CardDescription>General ledger and financial statements</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        Accounting interface would go here
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>
            </div>
          )}

          {activeView === 'data' && (
            <div className="space-y-6">
              <Tabs defaultValue="dbf-explorer" className="w-full">
                <TabsList>
                  <TabsTrigger value="dbf-explorer">DBF Explorer</TabsTrigger>
                  <TabsTrigger value="import">Import</TabsTrigger>
                  <TabsTrigger value="export">Export</TabsTrigger>
                  <TabsTrigger value="backup">Backup</TabsTrigger>
                  <TabsTrigger value="maintenance">Maintenance</TabsTrigger>
                </TabsList>
                <TabsContent value="dbf-explorer" className="space-y-4">
                  <DBFExplorer currentUser={currentUser} />
                </TabsContent>
                <TabsContent value="import" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Data Import</CardTitle>
                      <CardDescription>Import data from external sources</CardDescription>
                    </CardHeader>
                    <CardContent>
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
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="export" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Data Export</CardTitle>
                      <CardDescription>Export data in various formats</CardDescription>
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
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="backup" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Backup & Restore</CardTitle>
                      <CardDescription>Manage database backups</CardDescription>
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
                          <Button variant="outline" className="w-full">Restore from Backup</Button>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="maintenance" className="space-y-4">
                  <div className="grid gap-4 md:grid-cols-2">
                    <Card>
                      <CardHeader>
                        <CardTitle>Database Status</CardTitle>
                        <CardDescription>Current database health</CardDescription>
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
                        </div>
                      </CardContent>
                    </Card>
                    <Card>
                      <CardHeader>
                        <CardTitle>Maintenance Tools</CardTitle>
                        <CardDescription>Database optimization</CardDescription>
                      </CardHeader>
                      <CardContent className="space-y-3">
                        <Button variant="outline" className="w-full">Optimize Tables</Button>
                        <Button variant="outline" className="w-full">Check Data Integrity</Button>
                      </CardContent>
                    </Card>
                  </div>
                </TabsContent>
              </Tabs>
            </div>
          )}

          {activeView === 'reporting' && (
            <div className="space-y-6">
              <Tabs defaultValue="state" className="w-full">
                <TabsList>
                  <TabsTrigger value="state">State Reports</TabsTrigger>
                  <TabsTrigger value="financial">Financial Reports</TabsTrigger>
                  <TabsTrigger value="production">Production Reports</TabsTrigger>
                  <TabsTrigger value="custom">Custom Reports</TabsTrigger>
                  <TabsTrigger value="audit">Audit Trail</TabsTrigger>
                </TabsList>
                <TabsContent value="state" className="space-y-4">
                  <StateReportsSection currentUser={currentUser} />
                </TabsContent>
                <TabsContent value="financial" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Financial Reports</CardTitle>
                      <CardDescription>Generate financial statements and summaries</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        Financial report generator would go here
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="production" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Production Reports</CardTitle>
                      <CardDescription>Well production and field reports</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        Production report generator would go here
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="custom" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Custom Report Builder</CardTitle>
                      <CardDescription>Create custom reports with your data</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        Custom report builder would go here
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="audit" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Audit Trail</CardTitle>
                      <CardDescription>System activity and change logs</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="rounded-md border">
                        <Table>
                          <TableHeader>
                            <TableRow>
                              <TableHead>Timestamp</TableHead>
                              <TableHead>User</TableHead>
                              <TableHead>Action</TableHead>
                              <TableHead>Details</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            <TableRow>
                              <TableCell colSpan={4} className="text-center text-muted-foreground py-8">
                                Audit log entries would appear here
                              </TableCell>
                            </TableRow>
                          </TableBody>
                        </Table>
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>
            </div>
          )}

          {activeView === 'settings' && (
            <div className="space-y-6">
              <Tabs defaultValue="users" className="w-full">
                <TabsList>
                  <TabsTrigger value="users">User Management</TabsTrigger>
                  <TabsTrigger value="appearance">Appearance</TabsTrigger>
                  <TabsTrigger value="system">System</TabsTrigger>
                  <TabsTrigger value="security">Security</TabsTrigger>
                </TabsList>
                <TabsContent value="users" className="space-y-4">
                  <UserManagement currentUser={currentUser} />
                </TabsContent>
                <TabsContent value="appearance" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Appearance Settings</CardTitle>
                      <CardDescription>Customize the look and feel of your application</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-6">
                      <div className="space-y-2">
                        <h4 className="text-sm font-medium">Theme</h4>
                        <p className="text-sm text-muted-foreground">Select your preferred theme</p>
                        <div className="pt-2">
                          <ThemeSwitcher />
                        </div>
                      </div>
                      <div className="border-t pt-6">
                        <h4 className="text-sm font-medium mb-4">Additional Display Options</h4>
                        <div className="space-y-4">
                          <div className="flex items-center justify-between">
                            <div>
                              <p className="text-sm font-medium">Compact Mode</p>
                              <p className="text-sm text-muted-foreground">Reduce spacing for more content</p>
                            </div>
                            <Button variant="outline" size="sm" disabled>Coming Soon</Button>
                          </div>
                          <div className="flex items-center justify-between">
                            <div>
                              <p className="text-sm font-medium">Font Size</p>
                              <p className="text-sm text-muted-foreground">Adjust text size for better readability</p>
                            </div>
                            <Button variant="outline" size="sm" disabled>Coming Soon</Button>
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="system" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>System Configuration</CardTitle>
                      <CardDescription>Configure system-wide settings and API keys</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        API Key management will be available once the application is built with Wails.
                        <br />
                        <small className="text-xs">The configuration system has been implemented and is ready to use.</small>
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

          {activeView === 'utilities' && (
            <div className="space-y-6">
              <Tabs defaultValue="calculator" className="w-full">
                <TabsList>
                  <TabsTrigger value="calculator">Calculator</TabsTrigger>
                  <TabsTrigger value="converter">Converter</TabsTrigger>
                  <TabsTrigger value="scheduler">Scheduler</TabsTrigger>
                  <TabsTrigger value="tools">Data Tools</TabsTrigger>
                </TabsList>
                <TabsContent value="calculator" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Financial Calculators</CardTitle>
                      <CardDescription>Various calculation tools</CardDescription>
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
                            <div className="font-semibold">Production Calculator</div>
                            <div className="text-sm text-muted-foreground">Calculate production metrics</div>
                          </div>
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="converter" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Unit Converter</CardTitle>
                      <CardDescription>Convert between different measurement units</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-center py-8 text-muted-foreground">
                        Unit conversion tools would go here
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="scheduler" className="space-y-4">
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
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
                <TabsContent value="tools" className="space-y-4">
                  <Card>
                    <CardHeader>
                      <CardTitle>Data Tools</CardTitle>
                      <CardDescription>Data validation and transformation utilities</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="grid gap-4 md:grid-cols-2">
                        <Button variant="outline" className="h-auto p-4">
                          <div className="text-left">
                            <div className="font-semibold">Data Validator</div>
                            <div className="text-sm text-muted-foreground">Check data integrity</div>
                          </div>
                        </Button>
                        <Button variant="outline" className="h-auto p-4">
                          <div className="text-left">
                            <div className="font-semibold">Duplicate Finder</div>
                            <div className="text-sm text-muted-foreground">Find duplicate records</div>
                          </div>
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>
              </Tabs>
            </div>
          )}
        </main>
      </div>
    </div>
  )
}

export default App
