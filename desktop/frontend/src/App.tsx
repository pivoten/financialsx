import React, { useState, useEffect, useRef } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from './lib/queryClient'
// Import Wails functions
import * as WailsApp from '../wailsjs/go/main/App';

// Check if Wails runtime is available
const checkWailsAvailable = () => {
  return typeof window !== 'undefined' && 
         (window as any).go?.main?.App;
};

// Wait for Wails to be ready
const waitForWails = async (maxRetries = 50) => {
  for (let i = 0; i < maxRetries; i++) {
    if (checkWailsAvailable()) {
      return true;
    }
    await new Promise(resolve => setTimeout(resolve, 100));
  }
  return false;
};

// Wrapper functions that wait for Wails to be ready
const GetCompanyList = async () => {
  // First check if the function is directly available (in dev mode)
  if (typeof WailsApp.GetCompanyList === 'function') {
    try {
      return await WailsApp.GetCompanyList();
    } catch (error) {
      console.error('Direct Wails call failed:', error);
    }
  }
  
  // Wait for Wails runtime
  const ready = await waitForWails();
  if (!ready) {
    console.error('Wails runtime not available after waiting');
    // Return empty array instead of throwing
    return [];
  }
  return WailsApp.GetCompanyList();
};

const SetDataPath = async (path: string) => {
  const ready = await waitForWails();
  if (!ready) {
    throw new Error('Wails runtime not available');
  }
  return WailsApp.SetDataPath(path);
};

const InitializeCompanyDatabase = async (company: string) => {
  const ready = await waitForWails();
  if (!ready) {
    throw new Error('Wails runtime not available');
  }
  return WailsApp.InitializeCompanyDatabase(company);
};

const GetDashboardData = async (company: string) => {
  const ready = await waitForWails();
  if (!ready) {
    return null;
  }
  return WailsApp.GetDashboardData(company);
};
import { signIn, signUp, signOut, getCurrentUser, onAuthStateChange, supabase } from './lib/supabase'
import LoginForm from './components/LoginForm'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from './components/ui/card'
import { Button } from './components/ui/button'
import { 
  DropdownMenu, 
  DropdownMenuContent, 
  DropdownMenuItem, 
  DropdownMenuTrigger,
  DropdownMenuSeparator,
  DropdownMenuLabel
} from './components/ui/dropdown-menu'
import { BankingSection } from './components/BankingSection'
import { DBFExplorer } from './components/DBFExplorer'
import OutstandingChecks from './components/OutstandingChecks'
import UtilitiesSection from './components/UtilitiesSection'
import { BankReconciliation } from './components/BankReconciliation'
import { CheckAudit } from './components/CheckAudit'
import { UserManagement } from './components/UserManagement'
import BillEntry from './components/BillEntry'
import BillEntryEnhanced from './components/BillEntryEnhanced'
import UserProfile from './components/UserProfile'
import WellManagement from './components/WellManagement'
import LegacyIntegration from './components/LegacyIntegration'
import SherWareLegacy from './components/SherWareLegacy'
import VendorManagement from './components/VendorManagement'
import VendorManagementDynamic from './components/VendorManagementDynamic'
import AuditTools from './components/AuditTools'
import FollowBatchNumber from './components/FollowBatchNumber'
import FinancialReports from './components/FinancialReports'
import logger from './services/logger'
import pivotenLogo from './assets/pivoten-logo.png'
import { DashboardCard } from './components/DashboardCard'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from './components/ui/breadcrumb'
import { 
  Building2, 
  FolderOpen, 
  Home,
  Database,
  FileText,
  BarChart3,
  Settings,
  LogOut,
  DollarSign,
  TrendingUp,
  Users,
  User,
  Activity,
  Calculator,
  FileSearch,
  Menu,
  ChevronLeft,
  Shield,
  Wrench,
  Upload,
  Download,
  Archive,
  Calendar,
  Copy,
  ExternalLink
} from 'lucide-react'

interface User {
  id: number
  username: string
  email: string
  name?: string
  picture_url?: string
  role_name: string
  is_root: boolean
  company_name: string
}

interface Company {
  name: string
  display_name: string
  alias?: string
  path: string
  full_path?: string
  address?: string
  city?: string
  state?: string
  zip?: string
}

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [currentUser, setCurrentUser] = useState<User | null>(null)
  const [companies, setCompanies] = useState<Company[]>([])
  const [selectedCompany, setSelectedCompany] = useState<string>('')
  const [selectedCompanyDisplay, setSelectedCompanyDisplay] = useState<string>('')
  const [companySelected, setCompanySelected] = useState(false)
  const [needsFolderSelection, setNeedsFolderSelection] = useState(false)
  const [isRegistering, setIsRegistering] = useState(false)
  
  // Form states
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Check Supabase session on mount
  useEffect(() => {
    checkSession()
    
    // Subscribe to auth changes
    const { data: authListener } = onAuthStateChange((event, session) => {
      logger.debug('Auth state changed', { event, hasSession: !!session })
      if (session?.user) {
        setCurrentUser({
          id: 0,
          username: session.user.email || '',
          email: session.user.email || '',
          role_name: 'user',
          is_root: false,
          company_name: ''
        })
        setIsAuthenticated(true)
      } else {
        setIsAuthenticated(false)
        setCurrentUser(null)
      }
    })
    
    return () => {
      authListener?.subscription?.unsubscribe()
    }
  }, [])
  
  // Load companies after authentication
  useEffect(() => {
    if (isAuthenticated && !companySelected) {
      loadCompanies()
    }
  }, [isAuthenticated, companySelected])
  
  const checkSession = async () => {
    const { user } = await getCurrentUser()
    if (user) {
      logger.debug('Existing Supabase session found', { userId: user.id })
      setCurrentUser({
        id: 0,
        username: user.email || '',
        email: user.email || '',
        role_name: 'user',
        is_root: false,
        company_name: ''
      })
      setIsAuthenticated(true)
    }
  }

  const loadCompanies = async () => {
    try {
      logger.debug('Loading companies from compmast.dbf')
      const companiesList = await GetCompanyList()
      
      if (companiesList && companiesList.length > 0) {
        // Transform the data to match expected format
        const transformedCompanies = companiesList.map((comp: any) => ({
          // Use data_path (folder name) as the identifier, not company_id which contains numeric values
          name: comp.data_path || comp.company_name || comp.name,
          display_name: comp.company_name || comp.display_name || comp.name,
          alias: comp.alias || '',  // Add the alias field
          // Use data_path as it contains the actual folder name
          path: comp.data_path || comp.company_name || comp.name || '',
          full_path: comp.full_path || '',  // Add the full resolved path
          address: comp.address1 || comp.address || '',
          city: comp.city || '',
          state: comp.state || '',
          zip: comp.zip_code || comp.zip || ''
        }))
        setCompanies(transformedCompanies)
        logger.debug('Loaded companies', { count: transformedCompanies.length })
      } else {
        logger.warn('No companies returned from backend')
        setCompanies([])
      }
    } catch (error: any) {
      logger.error('Failed to load companies', { 
        error: error.message,
        stack: error.stack,
        details: error
      })
      
      // Check if we need folder selection
      if (error.message && error.message.includes('NEED_FOLDER_SELECTION')) {
        setError('Please select the folder containing your company data files (compmast.dbf)')
        // Show folder selection button
        setNeedsFolderSelection(true)
      } else {
        setError('Failed to load companies. Please try again.')
      }
    }
  }

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setError('')
    setIsSubmitting(true)
    
    try {
      logger.debug('Attempting Supabase login', { email: email || username })
      
      // Use Supabase authentication
      const { data, error } = await signIn(email || username, password)
      
      if (error) {
        throw error
      }
      
      if (data?.user) {
        logger.debug('Supabase login successful', { userId: data.user.id })
        setCurrentUser({
          id: 0,
          username: data.user.email || username,
          email: data.user.email || '',
          role_name: 'user',
          is_root: false,
          company_name: ''
        })
        setIsAuthenticated(true) // Show company selection
      }
    } catch (err: any) {
      logger.error('Login failed', { error: err.message })
      setError(err.message || 'Failed to login')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleCompanySelect = async (company: Company) => {
    setError('')
    setIsSubmitting(true)
    
    try {
      logger.debug('Company selected, initializing database', { 
        company: company.name,
        path: company.path
      })
      
      // Initialize the company database for this user
      // This creates/opens the SQLite database for the selected company
      // Use company.path for the actual data directory path
      await InitializeCompanyDatabase(company.path || company.name)
      
      // Use company.path for the selected company (this is the actual data directory)
      setSelectedCompany(company.path || company.name)
      // Store the display name separately (use alias or display_name if available)
      setSelectedCompanyDisplay(company.alias || company.display_name || company.name)
      setCompanySelected(true)
      
      // Store for session persistence
      localStorage.setItem('company_name', company.name)
      localStorage.setItem('company_display', company.alias || company.display_name || company.name)
      // Don't store company_path anymore to avoid confusion
      localStorage.removeItem('company_path')
      
      logger.debug('Company database initialized successfully', { company: company.name })
    } catch (err: any) {
      logger.error('Failed to initialize company database', { error: err.message })
      setError(err.message || 'Failed to initialize company database. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleRequestLogin = () => {
    // Stubbed out for now - will implement actual request process later
    logger.debug('Request Login clicked')
    alert('Request Login functionality coming soon!\n\nPlease contact your administrator to request access.')
  }
  
  const handleResetPassword = async () => {
    logger.debug('Reset Password clicked')
    
    // Prompt for email
    const resetEmail = prompt('Enter your email address to receive a password reset link:')
    
    if (resetEmail) {
      try {
        setIsSubmitting(true)
        
        // Call Supabase password reset (this sends an email with reset link)
        const { error } = await supabase?.auth.resetPasswordForEmail(resetEmail, {
          redirectTo: `${window.location.origin}/reset-password`,
        })
        
        if (error) {
          throw error
        }
        
        alert('Password reset email sent! Please check your inbox.')
        logger.debug('Password reset email sent', { email: resetEmail })
      } catch (err: any) {
        logger.error('Password reset failed', { error: err.message })
        alert(`Failed to send reset email: ${err.message}`)
      } finally {
        setIsSubmitting(false)
      }
    }
  }

  const handleLogout = async () => {
    try {
      await signOut()
      logger.debug('Supabase logout successful')
    } catch (err: any) {
      logger.error('Logout error', { error: err.message })
    }
    
    setIsAuthenticated(false)
    setCompanySelected(false)
    setCurrentUser(null)
    setSelectedCompany('')
    setSelectedCompanyDisplay('')
    setUsername('')
    setPassword('')
    setEmail('')
    setError('')
    setIsRegistering(false)
    setNeedsFolderSelection(false)
    localStorage.removeItem('company_name')
    localStorage.removeItem('company_path')
  }

  const handleSelectDataFolder = async () => {
    try {
      // For now, prompt user to enter the path manually
      // In a production app, you'd want to use a proper file dialog
      const selectedPath = prompt('Enter the full path to the folder containing compmast.dbf:')
      
      if (selectedPath) {
        logger.debug('Folder path entered', { path: selectedPath })
        
        // Set the data path in the backend
        await SetDataPath(selectedPath)
        
        // Reset the error and folder selection flag
        setError('')
        setNeedsFolderSelection(false)
        
        // Try loading companies again
        await loadCompanies()
      }
    } catch (err: any) {
      setError(err.message || 'Failed to set data folder')
    }
  }

  // Show login form if not authenticated
  if (!isAuthenticated) {
    return (
      <QueryClientProvider client={queryClient}>
        <LoginForm
          onLogin={handleLogin}
          onRequestLogin={handleRequestLogin}
          onResetPassword={handleResetPassword}
        username={username}
        setUsername={setUsername}
        email={email}
        setEmail={setEmail}
        password={password}
        setPassword={setPassword}
        error={error}
        isSubmitting={isSubmitting}
      />
      </QueryClientProvider>
    )
  }

  // Show company selection if authenticated but no company selected
  if (!companySelected) {
    return (
      <QueryClientProvider client={queryClient}>
        <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4">
        <Card className="w-full max-w-2xl shadow-xl">
          <CardHeader>
            <CardTitle className="text-2xl">
              {isRegistering ? 'Register New Account' : 'Select Company'}
            </CardTitle>
            <p className="text-gray-600">
              {isRegistering 
                ? 'Choose a company to create your account in' 
                : 'Choose a company to continue'}
            </p>
          </CardHeader>
          <CardContent>
            {error && (
              <div className="mb-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
                {error}
              </div>
            )}
            
            {needsFolderSelection ? (
              <div className="text-center py-8">
                <FolderOpen className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                <p className="text-gray-600 mb-4">
                  Cannot find company data files (compmast.dbf)
                </p>
                <Button
                  onClick={handleSelectDataFolder}
                  className="bg-orange-500 hover:bg-orange-600"
                >
                  <FolderOpen className="h-4 w-4 mr-2" />
                  Select Data Folder
                </Button>
              </div>
            ) : companies.length === 0 ? (
              <div className="text-center py-8">
                <p className="text-gray-500">Loading companies...</p>
              </div>
            ) : (
              <div className="grid gap-3">
                {companies.map((company) => (
                  <button
                    key={company.name}
                    onClick={() => handleCompanySelect(company)}
                    disabled={isSubmitting}
                    className="w-full p-4 bg-white border-2 border-gray-200 rounded-lg hover:shadow-md hover:border-orange-500 transition-all text-left group"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <div className="flex items-baseline gap-3">
                          <h3 className="font-semibold text-lg">{company.display_name}</h3>
                          {company.alias && (
                            <span className="text-sm text-gray-600 font-medium">({company.alias})</span>
                          )}
                        </div>
                        {company.city && company.state && (
                          <p className="text-sm text-gray-500">
                            {company.city}, {company.state}
                          </p>
                        )}
                        {company.full_path && (
                          <p className="text-xs text-gray-400 mt-1 font-mono">
                            {company.full_path}
                          </p>
                        )}
                      </div>
                      <Building2 className="h-5 w-5 text-gray-400 group-hover:text-orange-500" />
                    </div>
                  </button>
                ))}
              </div>
            )}
            
            <div className="mt-6 text-center">
              <Button
                variant="ghost"
                onClick={handleLogout}
                className="text-gray-600"
              >
                Back to Login
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
      </QueryClientProvider>
    )
  }

  // Show main application dashboard
  return (
    <QueryClientProvider client={queryClient}>
      <AdvancedDashboard 
        currentUser={currentUser} 
        onLogout={handleLogout} 
        selectedCompany={selectedCompany}
        selectedCompanyDisplay={selectedCompanyDisplay} 
      />
    </QueryClientProvider>
  )
}

// Advanced Dashboard Component
interface AdvancedDashboardProps {
  currentUser: User | null
  onLogout: () => void
  selectedCompany: string
  selectedCompanyDisplay: string
}

function AdvancedDashboard({ currentUser, onLogout, selectedCompany, selectedCompanyDisplay }: AdvancedDashboardProps) {
  const [activeView, setActiveView] = useState('dashboard')
  const [activeSubView, setActiveSubView] = useState('')
  const [activeSubSubView, setActiveSubSubView] = useState('')
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)

  // Main navigation items
  const menuItems = [
    { id: 'dashboard', label: 'Dashboard', icon: Home },
    { id: 'sherware', label: 'Legacy', icon: ExternalLink },
    { id: 'operations', label: 'Operations', icon: Activity },
    { id: 'financials', label: 'Financials', icon: DollarSign },
    { id: 'reporting', label: 'Reporting', icon: FileText },
    { id: 'utilities', label: 'Utilities', icon: Wrench },
    { id: 'settings', label: 'Settings', icon: Settings },
  ]

  const getPageTitle = () => {
    if (activeView === 'dashboard') return 'Dashboard'
    if (activeView === 'sherware') return 'Legacy'
    
    // Sub-view titles
    if (activeSubView) {
      const subViewTitles: Record<string, string> = {
        // Financials
        'banking': 'Banking',
        'transactions': 'Accounts Payable',
        'revenue': 'Revenue Analysis',
        'expenses': 'Expense Management',
        'analytics': 'Financial Analytics',
        'accounting': 'Accounting Tools',
        // Data
        'dbf-explorer': 'DBF Explorer',
        'import': 'Import Data',
        'export': 'Export Data',
        'backup': 'Backup & Restore',
        'maintenance': 'Database Maintenance',
        // Settings
        'users': 'User Management',
        'profile': 'My Profile',
        'appearance': 'Appearance',
        'system': 'System Configuration',
        'security': 'Security Settings',
        // Operations
        'wells': 'Well Management',
        'production': 'Production Data',
        'field-ops': 'Field Operations',
        // Reporting
        'state': 'State Reports',
        'financial': 'Financial Reports',
        'custom': 'Custom Reports',
        'audit': 'Audit Trail',
        // Utilities
        'calculators': 'Calculators',
        'converters': 'Converters',
        'tools': 'System Tools'
      }
      return subViewTitles[activeSubView] || activeSubView
    }
    
    const viewTitles: Record<string, string> = {
      operations: 'Operations Dashboard',
      financials: 'Financial Dashboard',
      reporting: 'Reports Dashboard',
      utilities: 'Utilities Dashboard',
      settings: 'Settings Dashboard',
      sherware: 'Legacy'
    }
    
    return viewTitles[activeView] || 'Dashboard'
  }

  const getPageDescription = () => {
    const descriptions: Record<string, string> = {
      dashboard: 'Overview of your financial data',
      operations: 'Manage wells, production, and field operations',
      financials: 'Financial transactions, analytics, and accounting',
      data: 'Database maintenance and data management',
      reporting: 'Reports, compliance, and documentation',
      utilities: 'Tools, calculators, and system utilities',
      settings: 'System configuration and user management',
      sherware: 'Launch Visual FoxPro forms directly from FinancialsX'
    }
    
    return descriptions[activeView] || ''
  }

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sidebar */}
      <div 
        className={`${isSidebarCollapsed ? 'w-16' : 'w-64'} bg-white border-r border-gray-200 transition-all duration-300 flex flex-col shadow-sm relative`}
      >
        <div className="p-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            {!isSidebarCollapsed ? (
              <>
                <div className="flex items-center space-x-2">
                  <img src={pivotenLogo} alt="Pivoten" className="h-8 w-8 object-contain" />
                  <span className="font-semibold text-lg">FinancialsX</span>
                </div>
                <button
                  onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
                  className="p-1.5 hover:bg-gray-100 rounded-md transition-colors"
                  aria-label="Collapse sidebar"
                >
                  <ChevronLeft className="h-5 w-5 text-gray-600" />
                </button>
              </>
            ) : (
              <button
                onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
                className="mx-auto p-2 hover:bg-gray-100 rounded-md transition-colors"
                aria-label="Expand sidebar"
              >
                <Menu className="h-5 w-5 text-gray-600" />
              </button>
            )}
          </div>
        </div>

        <nav className="flex-1 p-3">
          {menuItems.map((item) => {
            const Icon = item.icon
            return (
              <button
                key={item.id}
                onClick={() => {
                  setActiveView(item.id)
                  setActiveSubView('')
                  setActiveSubSubView('')
                }}
                className={`w-full flex items-center space-x-3 px-3 py-2 rounded-md mb-1 transition-all ${
                  activeView === item.id
                    ? 'bg-orange-50 text-orange-600 font-medium'
                    : 'hover:bg-gray-50 text-gray-600 hover:text-gray-900'
                }`}
                title={isSidebarCollapsed ? item.label : ''}
              >
                <Icon className="h-5 w-5 flex-shrink-0" />
                {!isSidebarCollapsed && <span>{item.label}</span>}
              </button>
            )
          })}
        </nav>

        <div className="p-3 border-t border-gray-200">
          {!isSidebarCollapsed && (
            <div 
              className="mb-3 cursor-pointer hover:bg-gray-50 rounded p-2 transition-colors"
              onClick={() => {
                setActiveView('settings');
                setActiveSubView('profile');
              }}
              title="View Profile"
            >
              <p className="text-sm text-gray-600 break-all">{currentUser?.email}</p>
              <p className="text-xs text-gray-500 break-all">Company: {selectedCompany}</p>
            </div>
          )}
          <Button
            onClick={onLogout}
            variant="outline"
            size="sm"
            className={isSidebarCollapsed ? "p-2" : "w-full"}
            title={isSidebarCollapsed ? "Logout" : ""}
          >
            <LogOut className="h-4 w-4" />
            {!isSidebarCollapsed && <span className="ml-2">Logout</span>}
          </Button>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="border-b bg-white px-6 py-4">
          {/* Breadcrumbs */}
          <div className="mb-3">
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbLink 
                    onClick={() => {
                      setActiveView('dashboard')
                      setActiveSubView('')
                    }}
                  >
                    Dashboard
                  </BreadcrumbLink>
                </BreadcrumbItem>
                {activeView !== 'dashboard' && (
                  <>
                    <BreadcrumbSeparator />
                    <BreadcrumbItem>
                      {activeSubView ? (
                        <BreadcrumbLink
                          onClick={() => {
                            setActiveSubView('')
                            setActiveSubSubView('')
                          }}
                        >
                          {menuItems.find(item => item.id === activeView)?.label}
                        </BreadcrumbLink>
                      ) : (
                        <BreadcrumbPage>
                          {menuItems.find(item => item.id === activeView)?.label}
                        </BreadcrumbPage>
                      )}
                    </BreadcrumbItem>
                  </>
                )}
                {activeSubView && (
                  <>
                    <BreadcrumbSeparator />
                    <BreadcrumbItem>
                      <BreadcrumbPage>
                        {getPageTitle()}
                      </BreadcrumbPage>
                    </BreadcrumbItem>
                  </>
                )}
              </BreadcrumbList>
            </Breadcrumb>
          </div>
          
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-6">
              
              {/* Page Title */}
              <div>
                <h2 className="text-xl font-semibold tracking-tight">{getPageTitle()}</h2>
                <p className="text-sm text-gray-500">{getPageDescription()}</p>
              </div>
              
              {/* Browser Mode Indicator - removed since we're always using real data now */}
            </div>
            
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-600">Welcome, {currentUser?.username}</span>
              <span className="px-2 py-1 bg-orange-100 text-orange-700 rounded-full text-xs font-medium">
                {selectedCompanyDisplay || selectedCompany}
              </span>
            </div>
          </div>
        </header>

        {/* Main Content */}
        <main className="flex-1 overflow-auto p-6 bg-gray-50">
          {/* Dashboard View */}
          {activeView === 'dashboard' && (
            <div className="space-y-6">
              {/* Stats Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-gray-500">Total Revenue</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">$1,245,890</div>
                    <p className="text-xs text-green-600">+12.5% from last month</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-gray-500">Active Wells</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">42</div>
                    <p className="text-xs text-green-600">+2 this month</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-gray-500">Outstanding Checks</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">156</div>
                    <p className="text-xs text-red-600">-8 from last week</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-gray-500">Bank Balance</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">$523,450</div>
                    <p className="text-xs text-green-600">+5.2% this week</p>
                  </CardContent>
                </Card>
              </div>

              {/* Quick Actions */}
              <Card>
                <CardHeader>
                  <CardTitle>Quick Actions</CardTitle>
                  <CardDescription>Common tasks and operations</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
                    <Button variant="outline" className="h-auto flex-col py-4">
                      <DollarSign className="h-5 w-5 mb-2" />
                      <span className="text-xs">New Transaction</span>
                    </Button>
                    <Button variant="outline" className="h-auto flex-col py-4">
                      <FileText className="h-5 w-5 mb-2" />
                      <span className="text-xs">Generate Report</span>
                    </Button>
                    <Button variant="outline" className="h-auto flex-col py-4">
                      <Calculator className="h-5 w-5 mb-2" />
                      <span className="text-xs">Reconciliation</span>
                    </Button>
                    <Button variant="outline" className="h-auto flex-col py-4">
                      <Upload className="h-5 w-5 mb-2" />
                      <span className="text-xs">Import Data</span>
                    </Button>
                    <Button variant="outline" className="h-auto flex-col py-4">
                      <Download className="h-5 w-5 mb-2" />
                      <span className="text-xs">Export Data</span>
                    </Button>
                    <Button variant="outline" className="h-auto flex-col py-4">
                      <Users className="h-5 w-5 mb-2" />
                      <span className="text-xs">Users</span>
                    </Button>
                  </div>
                </CardContent>
              </Card>

              {/* Recent Activity */}
              <Card>
                <CardHeader>
                  <CardTitle>Recent Activity</CardTitle>
                  <CardDescription>Latest transactions and events</CardDescription>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-gray-500">No recent activity to display</p>
                </CardContent>
              </Card>
            </div>
          )}

          {/* Financials Section */}
          {activeView === 'financials' && (
            <div className="space-y-6">
              {!activeSubView && (
                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                  <DashboardCard
                    title="Accounts"
                    subtitle="Banking"
                    description="Bank accounts and reconciliation"
                    icon={Home}
                    onClick={() => setActiveSubView('banking')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Activity"
                    subtitle="Transactions"
                    description="View and manage transactions"
                    icon={DollarSign}
                    onClick={() => setActiveSubView('transactions')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Analysis"
                    subtitle="Revenue"
                    description="Revenue streams and trends"
                    icon={TrendingUp}
                    onClick={() => setActiveSubView('revenue')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Audit Tools"
                    subtitle="Data Validation"
                    description="Check data integrity and validate GL entries"
                    icon={FileSearch}
                    onClick={() => setActiveSubView('audit')}
                    accentColor="amber"
                  />
                  <DashboardCard
                    title="Maintain"
                    subtitle="Account Management"
                    description="Maintain accounts, vendors, and financial records"
                    icon={Wrench}
                    onClick={() => setActiveSubView('maintain')}
                    accentColor="gray"
                  />
                </div>
              )}
              
              {activeSubView === 'banking' && <BankingSection currentUser={currentUser} companyName={selectedCompany} />}
              {activeSubView === 'audit' && <AuditTools currentUser={currentUser} companyName={selectedCompany} />}
              {activeSubView === 'transactions' && <BillEntryEnhanced currentUser={currentUser} companyName={selectedCompany} />}
              {activeSubView === 'maintain' && !activeSubSubView && (
                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                  <DashboardCard
                    title="Vendors"
                    subtitle="Vendor Management"
                    description="Manage vendor information and records"
                    icon={Users}
                    onClick={() => setActiveSubSubView('vendors')}
                    accentColor="blue"
                  />
                </div>
              )}
              {activeSubView === 'maintain' && activeSubSubView === 'vendors' && (
                <VendorManagementDynamic currentUser={currentUser} companyName={selectedCompany} />
              )}
              {activeSubView === 'revenue' && (
                <Card>
                  <CardHeader>
                    <CardTitle>Revenue Analysis</CardTitle>
                    <CardDescription>Analyze revenue trends and performance</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <p className="text-gray-500">Revenue analysis coming soon</p>
                  </CardContent>
                </Card>
              )}
              {activeSubView === 'expenses' && (
                <Card>
                  <CardHeader>
                    <CardTitle>Expense Management</CardTitle>
                    <CardDescription>Track and manage business expenses</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <p className="text-gray-500">Expense management coming soon</p>
                  </CardContent>
                </Card>
              )}
              {activeSubView === 'analytics' && (
                <Card>
                  <CardHeader>
                    <CardTitle>Financial Analytics</CardTitle>
                    <CardDescription>Advanced financial analysis and reporting</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <p className="text-gray-500">Financial analytics coming soon</p>
                  </CardContent>
                </Card>
              )}
              {activeSubView === 'accounting' && <BillEntryEnhanced currentUser={currentUser} companyName={selectedCompany} />}
            </div>
          )}


          {/* SherWare Legacy Section */}
          {activeView === 'sherware' && (
            <SherWareLegacy />
          )}

          {/* Settings Section */}
          {activeView === 'settings' && (
            <div className="space-y-6">
              {!activeSubView && (
                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                  <DashboardCard
                    title="Profile"
                    subtitle="My Account"
                    description="View and edit your profile information"
                    icon={User}
                    onClick={() => setActiveSubView('profile')}
                    accentColor="blue"
                  />
                  <DashboardCard
                    title="Management"
                    subtitle="Users"
                    description="Manage user accounts and roles"
                    icon={Users}
                    onClick={() => setActiveSubView('users')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Appearance"
                    subtitle="Interface"
                    description="Customize interface and themes"
                    icon={Settings}
                    onClick={() => setActiveSubView('appearance')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Configuration"
                    subtitle="System"
                    description="System settings and preferences"
                    icon={Database}
                    onClick={() => setActiveSubView('system')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Security"
                    subtitle="Access Control"
                    description="Security and access settings"
                    icon={Shield}
                    onClick={() => setActiveSubView('security')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Legacy Integration"
                    subtitle="Visual FoxPro"
                    description="Configure VFP application integration"
                    icon={ExternalLink}
                    onClick={() => setActiveSubView('legacy')}
                    accentColor="blue"
                  />
                </div>
              )}
              
              {activeSubView === 'users' && <UserManagement currentUser={currentUser} />}
              {activeSubView === 'profile' && <UserProfile currentUser={currentUser} companyName={selectedCompany} />}
              {activeSubView === 'legacy' && <LegacyIntegration onBack={() => setActiveSubView(null)} />}
            </div>
          )}

          {/* Operations Section */}
          {activeView === 'operations' && (
            <div className="space-y-6">
              {!activeSubView && (
                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                  <DashboardCard
                    title="Wells"
                    subtitle="Well Management"
                    description="Manage wells and production data"
                    icon={Activity}
                    onClick={() => setActiveSubView('wells')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Production"
                    subtitle="Production Data"
                    description="Track production volumes and metrics"
                    icon={TrendingUp}
                    onClick={() => setActiveSubView('production')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Field Ops"
                    subtitle="Field Operations"
                    description="Field operations and maintenance"
                    icon={Wrench}
                    onClick={() => setActiveSubView('field-ops')}
                    accentColor="gray"
                  />
                </div>
              )}
              
              {activeSubView === 'wells' && <WellManagement currentUser={currentUser} companyName={selectedCompany} />}
            </div>
          )}

          {/* Reporting Section */}
          {activeView === 'reporting' && (
            <div className="space-y-6">
              {!activeSubView && (
                <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                  <DashboardCard
                    title="State Reports"
                    subtitle="Compliance"
                    description="State regulatory reports"
                    icon={FileText}
                    onClick={() => setActiveSubView('state')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Financial"
                    subtitle="Reports"
                    description="Financial statements and reports"
                    icon={DollarSign}
                    onClick={() => setActiveSubView('financial')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Production"
                    subtitle="Reports"
                    description="Production and revenue reports"
                    icon={TrendingUp}
                    onClick={() => setActiveSubView('production')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Custom"
                    subtitle="Reports"
                    description="Create custom reports"
                    icon={FileSearch}
                    onClick={() => setActiveSubView('custom')}
                    accentColor="gray"
                  />
                  <DashboardCard
                    title="Audit Trail"
                    subtitle="Tracking"
                    description="System audit and activity logs"
                    icon={Copy}
                    onClick={() => setActiveSubView('audit')}
                    accentColor="gray"
                  />
                </div>
              )}
              {activeSubView === 'financial' && (
                <FinancialReports companyName={selectedCompany} currentUser={currentUser} />
              )}
            </div>
          )}

          {/* Utilities Section */}
          {activeView === 'utilities' && (
            <UtilitiesSection 
              currentUser={currentUser} 
              currentCompany={selectedCompany} 
            />
          )}
        </main>
      </div>
    </div>
  )
}

export default App