
import React, { useState, useEffect, useRef } from 'react'
import { Login, Register, GetCompanies, GetCompanyList, ValidateSession, GetDashboardData } from '../wailsjs/go/main/App'
import { User, Company, FormEvent, ChangeEvent, MouseEvent, SupabaseUser } from './types'
import { supabase, isSupabaseConfigured, signIn, signUp } from './lib/supabase'
import { Button } from './components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from './components/ui/card'
import { Input } from './components/ui/input'
import { Badge } from './components/ui/badge'
import { Table, TableHeader, TableRow, TableHead, TableBody, TableCell } from './components/ui/table'
import { Sidebar, SidebarHeader, SidebarContent, SidebarNav, SidebarNavItem, SidebarNavGroup } from './components/ui/sidebar'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './components/ui/tabs'
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from './components/ui/dropdown-menu'
import { RevenueChart } from './components/charts/RevenueChart'
import { ProductionChart } from './components/charts/ProductionChart'
import { TransactionsTable } from './components/tables/TransactionsTable'
import { DBFExplorer } from './components/DBFExplorer'
import { DatabaseTest } from './components/DatabaseTest'
import { UserManagement } from './components/UserManagement'
import { StateReportsSection } from './components/StateReportsSection'
import { BankingSection } from './components/BankingSection'
import CompanyInformation from './components/CompanyInformation'
import LoggingTest from './components/LoggingTest'
import { ThemeProvider } from './components/theme-provider'
import { ThemeSwitcher } from './components/theme-switcher'
import pivotenLogo from './assets/pivoten-logo.png'
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
	ChevronLeft,
	Building2,
	Shield,
} from 'lucide-react'
import './globals.css'
import logger from './services/logger'

function App() {
	logger.debug('App component rendering...')
  
	const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false)
	const [currentUser, setCurrentUser] = useState<User | null>(null)
	const [companies, setCompanies] = useState<Company[]>([])
	const [showRegister, setShowRegister] = useState<boolean>(false)
	const [loading, setLoading] = useState<boolean>(true)
	const [selectedCompany, setSelectedCompany] = useState<string>('')
	const [companySelected, setCompanySelected] = useState<boolean>(false)

	// Form states
	const [username, setUsername] = useState<string>('')
	const [password, setPassword] = useState<string>('')
	const [email, setEmail] = useState<string>('')
	const [error, setError] = useState<string>('')
	const [isSubmitting, setIsSubmitting] = useState<boolean>(false)

	// Log Supabase configuration on mount
	useEffect(() => {
		logger.debug('App mounted, checking Supabase configuration...')
		logger.debug('Supabase configured?', { configured: isSupabaseConfigured() })
		logger.debug('Supabase client exists?', { exists: !!supabase })
		if (supabase) {
			logger.debug('Supabase URL from client: configured')
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
		const savedCompanyPath = localStorage.getItem('company_path')
    
		logger.debug('checkSession: Found stored session?', { hasToken: !!token, authType })
    
		if (token) {
			try {
				if (authType === 'supabase' && isSupabaseConfigured()) {
					// Validate Supabase session
					const { data: { user }, error } = await supabase.auth.getUser(token)
					if (error) throw error
          
					logger.debug('checkSession: Supabase session valid', { userId: user?.id, email: user?.email })
					setCurrentUser({
						id: parseInt(user.id) || 0,
						email: user.email || '',
						username: user.email || '',
						role_name: 'user',
						is_root: false,
						company_name: savedCompany || ''
					})
					setIsAuthenticated(true)
          
					// Check if user had a company selected previously
					if (savedCompany) {
						setSelectedCompany(savedCompany)
						setCompanySelected(true)
					}
				} else if (authType === 'local' && savedCompany) {
					// Validate local SQLite session (requires company)
				logger.debug('checkSession: Validating local session', { company: savedCompany })
					const user = await ValidateSession(token, savedCompany)
				logger.debug('checkSession: Session valid', { username: user?.username, company: user?.company_name })
					setCurrentUser(user as User)
					setIsAuthenticated(true)
					setSelectedCompany(savedCompany)
					setCompanySelected(true)
				}
			} catch (error: any) {
				logger.warn('checkSession: Session validation failed', { error: error.message })
				localStorage.removeItem('session_token')
				localStorage.removeItem('company_name')
				localStorage.removeItem('auth_type')
			}
		} else {
			logger.debug('checkSession: No stored session found')
		}
		setLoading(false)
	}

	const loadCompanies = async () => {
		try {
			// Check if Wails runtime is available
			if (!window.go || !window.go.main || !window.go.main.App) {
				logger.warn('Wails runtime not available')
				// For testing without Wails, use mock data
				setCompanies([
					{ name: 'testdata', display_name: 'Pivoten Operating LLC', path: 'datafiles/testdata' },
					{ name: '00000198', display_name: 'Limecreek Energy, LLC', path: 'datafiles/limecreekenergydata' }
				])
				return
			}
      
			// Try to load companies from compmast.dbf first
			try {
				const companiesList = await GetCompanyList()
			logger.debug('Loaded companies from compmast.dbf', { count: companiesList?.length })
        
				if (companiesList && companiesList.length > 0) {
					// Transform the data to match expected format
					const transformedCompanies = companiesList.map(comp => ({
						name: comp.company_id || comp.company_name,
						display_name: comp.company_name,
						path: comp.data_path,
						address: comp.address1,
						city: comp.city,
						state: comp.state,
						zip: comp.zip_code
					}))
					setCompanies(transformedCompanies)
					return
				}
			} catch (dbfError) {
			logger.warn('Could not load from compmast.dbf', { error: dbfError.message })
			}
      
			// Fallback to GetCompanies if compmast.dbf is not available
			const companiesList = await GetCompanies()
			const transformedFallback = (companiesList || []).map((comp: any) => ({
				name: comp.company_id || comp.name || comp.company_name,
				display_name: comp.company_name || comp.display_name || comp.name,
				path: comp.data_path || comp.path || '',
				address: comp.address1 || comp.address || '',
				city: comp.city || '',
				state: comp.state || '',
				zip: comp.zip_code || comp.zip || ''
			}))
			setCompanies(transformedFallback)
		} catch (error: any) {
			logger.error('Failed to load companies', { error: error.message })
			// Set some default companies for testing
			setCompanies([
				{ name: 'testdata', display_name: 'Pivoten Operating LLC', path: 'datafiles/testdata' },
				{ name: '00000198', display_name: 'Limecreek Energy, LLC', path: 'datafiles/limecreekenergydata' }
			])
		}
	}

	const handleLogin = async (e: FormEvent) => {
		e.preventDefault()
		setError('')
    
		if (!username || !password) {
			setError('Please enter your email and password')
			return
		}

		setIsSubmitting(true)
		try {
			logger.debug('Login attempt', { username, supabaseConfigured: isSupabaseConfigured() })
      
			if (isSupabaseConfigured()) {
				// Use Supabase authentication
			logger.debug('Attempting Supabase login', { email: username, passwordLength: password.length })
				const { data, error } = await signIn(username, password)
        
				logger.debug('Supabase response', { hasData: !!data, hasError: !!error })
        
				if (error) {
				logger.error('Supabase authentication error', { error: error.message })
					throw error
				}
        
				// Store session info
				localStorage.setItem('session_token', data.session.access_token)
				localStorage.setItem('auth_type', 'supabase')
        
				// Set user info
				setCurrentUser({
					id: parseInt(data.user.id) || 0,
					email: data.user.email || '',
					username: data.user.email || '',
					role_name: 'user',
					is_root: false,
					company_name: ''
				})
				setIsAuthenticated(true)
			} else {
				// For local auth, we need to select a company first
				setError('Please use Supabase authentication or select a company for local auth')
			}
		} catch (error: any) {
				logger.error('Login error details', { error: error.message })
			setError(error.message || 'Invalid login credentials')
		} finally {
			setIsSubmitting(false)
		}
	}

	const handleRegister = async (e: FormEvent) => {
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
					id: parseInt(loginData.user.id) || 0,
					email: loginData.user.email || '',
					username: loginData.user.email || '',
					role_name: 'user',
					is_root: false,
					company_name: ''
				})
				setIsAuthenticated(true)
			} else {
				setError('Please use Supabase authentication for registration')
			}
		} catch (error: any) {
			setError(error.message || 'Registration failed')
		} finally {
			setIsSubmitting(false)
		}
	}

	const handleCompanySelect = async (companyObj: Company | string) => {
		setError('')
		setIsSubmitting(true)
    
		try {
			// Store selected company info
			const companyName = typeof companyObj === 'string' ? companyObj : companyObj.name
			const companyPath = typeof companyObj === 'string' ? null : companyObj.path
      
			localStorage.setItem('company_name', companyName)
			if (companyPath) {
				localStorage.setItem('company_path', companyPath)
			}
			setSelectedCompany(companyName)
      
			// Initialize the SQLite database for this company (creates SQL folder if needed)
			try {
				logger.debug('Initializing company database', { company: companyPath || companyName })
				await window.go.main.App.InitializeCompanyDatabase(companyPath || companyName)
				logger.debug('Company database initialized successfully')
			} catch (dbError) {
				logger.error('Failed to initialize company database', { error: dbError.message })
				// Non-fatal error - continue with company selection
			}
      
			// For Supabase auth, we just need to set the company context
			// The actual data loading will happen in the dashboard
			setCompanySelected(true)
      
			// Update current user with company info
			setCurrentUser((prev: User | null) => ({
				...prev!,
				company_name: companyName
			}))
		} catch (error: any) {
				logger.error('Company selection error', { error: error.message })
			setError(error.message || 'Failed to select company')
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
								src={pivotenLogo} 
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
											onClick={() => handleCompanySelect(company)}
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
														<h3 className="font-semibold" style={{ color: '#2C4471' }}>
															{company.display_name || company.name}
														</h3>
														{company.address && (
															<p className="text-sm text-gray-600">
																{company.address}{company.city ? `, ${company.city}` : ''}{company.state ? `, ${company.state}` : ''} {company.zip || ''}
															</p>
														)}
														<p className="text-xs text-gray-500 mt-1">
															ID: {company.name} • {company.path || `datafiles/${company.name}`}
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
						<div className="mx-auto mb-4">
							<img 
								src={pivotenLogo} 
								alt="Pivoten Logo" 
								className="w-16 h-16 object-contain drop-shadow-md"
							/>
						</div>
						<CardDescription className="text-gray-600 text-base">
							{showRegister ? 'Create your account to get started' : 'Sign in to your account'}
						</CardDescription>
					</CardHeader>
					<CardContent>
						<form onSubmit={(e: React.FormEvent<HTMLFormElement>) => {
							logger.debug('Form submitted', { showRegister, hasUsername: !!username, hasPassword: !!password })
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
										onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
											const value = e.target.value
											logger.debug('Input changed', { field: showRegister ? 'email' : 'username', hasValue: !!value })
											showRegister ? setEmail(value) : setUsername(value)
										}}
										placeholder={showRegister ? "john@example.com" : "Enter your email or username"}
										className="pl-10 h-11 border-gray-200 transition-all"
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
											onChange={(e: React.ChangeEvent<HTMLInputElement>) => setUsername(e.target.value)}
											placeholder="Choose a username"
											className="pl-10 h-11 border-gray-200 transition-all"
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
										onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
											const value = e.target.value
											logger.debug('Password changed', { length: value.length })
											setPassword(value)
										}}
										placeholder="••••••••"
										className="pl-10 h-11 border-gray-200 transition-all"
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
								onClick={() => logger.debug('Sign in button clicked')}
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
interface AdvancedDashboardProps {
	currentUser: User;
	onLogout: () => void;
}

function AdvancedDashboard({ currentUser, onLogout }: AdvancedDashboardProps) {
	const [activeView, setActiveView] = useState<string>('dashboard')
	const [activeSubView, setActiveSubView] = useState<string>('')
	const [dashboardData, setDashboardData] = useState<any>(null)
	const [loadingDashboard, setLoadingDashboard] = useState<boolean>(true)
	const [isSidebarCollapsed, setIsSidebarCollapsed] = useState<boolean>(true) // Start collapsed
	const [isSidebarHovered, setIsSidebarHovered] = useState(false) // Track hover state
	const lastLoadedCompany = useRef(null)

	// Load dashboard data when company changes (not when user object changes)
	useEffect(() => {
		logger.debug('Dashboard useEffect triggered', {
			currentUserExists: !!currentUser,
			companyName: currentUser?.company_name,
			lastLoadedCompany: lastLoadedCompany.current
		})
    
		// Get the company identifier (path or name)
		const companyIdentifier = currentUser?.company_path || currentUser?.company_name
    
		// Only load if company actually changed
		if (companyIdentifier && companyIdentifier !== lastLoadedCompany.current) {
			logger.debug('Company changed, loading dashboard data', { company: companyIdentifier })
			lastLoadedCompany.current = companyIdentifier
			loadDashboardData()
		} else {
			logger.debug('No company change or already loaded, skipping')
		}
	}, [currentUser?.company_name, currentUser?.company_path])

	const loadDashboardData = async () => {
		logger.debug('loadDashboardData called', {
			currentUserExists: !!currentUser,
			companyName: currentUser?.company_name,
			companyPath: currentUser?.company_path
		})
    
		// Try to get company identifier from multiple sources
		const companyIdentifier = currentUser?.company_path || 
															currentUser?.company_name || 
															localStorage.getItem('company_path') || 
															localStorage.getItem('company_name')
    
		logger.debug('Company identifier resolved', { companyIdentifier })
    
		if (!companyIdentifier) {
			logger.warn('No company identifier available, cannot load dashboard')
			return
		}
    
		// Skip if we already have data for this company and it's not stale
		if (dashboardData && dashboardData.company === companyIdentifier) {
			logger.debug('Dashboard data already loaded', { company: companyIdentifier })
			return
		}
    
		logger.debug('Loading dashboard data', { company: companyIdentifier })
		setLoadingDashboard(true)
		try {
			// Check if Wails runtime is available
			if (!window.go || !window.go.main || !window.go.main.App) {
				logger.warn('Wails runtime not available - Run "wails dev" to access backend functions')
				// Set minimal dashboard data for browser development
				setDashboardData({
					company: companyIdentifier,
					totalRevenue: 0,
					totalExpenses: 0,
					netIncome: 0,
					wellCount: 0,
					recentTransactions: [],
					message: 'Backend not available - Run "wails dev" for full functionality'
				})
				return
			}
      
			logger.debug('Calling GetDashboardData', { company: companyIdentifier })
			const data = await GetDashboardData(companyIdentifier)
			logger.debug('Dashboard data loaded successfully', { hasData: !!data })
			setDashboardData(data)
		} catch (error: any) {
			logger.error('Failed to load dashboard data', { error: error.message })
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

		interface StatItem {
			title: string;
			value: string;
			change: string;
			icon: any;
			trend: 'up' | 'down';
		}
		const stats: StatItem[] = []
    
		// Show well types from WELLS.DBF
		if (dashboardData.wellTypes && dashboardData.wellTypes.length > 0) {
			dashboardData.wellTypes.forEach((wellType: any) => {
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
			{/* Auto-expanding Sidebar */}
			<div 
				className={`hidden lg:flex flex-col border-r transition-all duration-300 ease-in-out bg-card shadow-sm ${
					isSidebarHovered && isSidebarCollapsed ? 'w-64 shadow-lg' : isSidebarCollapsed ? 'w-16' : 'w-64'
				}`}
				onMouseEnter={() => setIsSidebarHovered(true)}
				onMouseLeave={() => setIsSidebarHovered(false)}
			>
				<div className="border-b px-4 py-4 flex items-center h-16">
					<div className={`flex items-center transition-all duration-300 ${
						(!isSidebarCollapsed || isSidebarHovered) ? 'w-full justify-between' : 'justify-center w-full'
					}`}>
						{(!isSidebarCollapsed || isSidebarHovered) && (
							<span className="font-bold text-lg tracking-tight">FinancialsX</span>
						)}
						<button
							onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
							className="p-2 hover:bg-accent rounded-md transition-colors"
						>
							{(!isSidebarCollapsed || isSidebarHovered) ? (
								<ChevronLeft className="h-4 w-4" />
							) : (
								<Menu className="h-4 w-4" />
							)}
						</button>
					</div>
				</div>
				<div className="flex-1 flex flex-col overflow-y-auto">
					<SidebarNav>
						<SidebarNavItem 
							href="#" 
							active={activeView === 'dashboard'}
							onClick={() => { setActiveView('dashboard'); setActiveSubView(''); }}
							className="flex items-center gap-3"
							title={isSidebarCollapsed && !isSidebarHovered ? "Dashboard" : ""}
						>
							<Home className="w-4 h-4 flex-shrink-0" />
							{(!isSidebarCollapsed || isSidebarHovered) && (
								<span className="ml-3 text-sm font-medium">Dashboard</span>
							)}
						</SidebarNavItem>
						<SidebarNavItem 
							href="#" 
							active={activeView === 'operations'}
							onClick={() => { setActiveView('operations'); setActiveSubView(''); }}
							className="flex items-center gap-3"
							title={isSidebarCollapsed && !isSidebarHovered ? "Operations" : ""}
						>
							<Activity className="w-4 h-4 flex-shrink-0" />
							{(!isSidebarCollapsed || isSidebarHovered) && (
								<span className="ml-3 text-sm font-medium">Operations</span>
							)}
						</SidebarNavItem>
						<SidebarNavItem 
							href="#" 
							active={activeView === 'financials'}
							onClick={() => { setActiveView('financials'); setActiveSubView(''); }}
							className="flex items-center gap-3"
							title={isSidebarCollapsed && !isSidebarHovered ? "Financials" : ""}
						>
							<DollarSign className="w-4 h-4 flex-shrink-0" />
							{(!isSidebarCollapsed || isSidebarHovered) && (
								<span className="ml-3 text-sm font-medium">Financials</span>
							)}
						</SidebarNavItem>
						<SidebarNavItem 
							href="#" 
							active={activeView === 'data'}
							onClick={() => { setActiveView('data'); setActiveSubView(''); }}
							className="flex items-center gap-3"
							title={isSidebarCollapsed && !isSidebarHovered ? "Data Management" : ""}
						>
							<Database className="w-4 h-4 flex-shrink-0" />
							{(!isSidebarCollapsed || isSidebarHovered) && (
								<span className="ml-3 text-sm font-medium">Data</span>
							)}
						</SidebarNavItem>
						<SidebarNavItem 
							href="#" 
							active={activeView === 'reporting'}
							onClick={() => { setActiveView('reporting'); setActiveSubView(''); }}
							className="flex items-center gap-3"
							title={isSidebarCollapsed && !isSidebarHovered ? "Reporting" : ""}
						>
							<FileText className="w-4 h-4 flex-shrink-0" />
							{(!isSidebarCollapsed || isSidebarHovered) && (
								<span className="ml-3 text-sm font-medium">Reporting</span>
							)}
						</SidebarNavItem>
						<SidebarNavItem 
							href="#" 
							active={activeView === 'utilities'}
							onClick={() => { setActiveView('utilities'); setActiveSubView(''); }}
							className="flex items-center gap-3"
							title={isSidebarCollapsed && !isSidebarHovered ? "Utilities" : ""}
						>
							<Wrench className="w-4 h-4 flex-shrink-0" />
							{(!isSidebarCollapsed || isSidebarHovered) && (
								<span className="ml-3 text-sm font-medium">Utilities</span>
							)}
						</SidebarNavItem>
						<SidebarNavItem 
							href="#" 
							active={activeView === 'settings'}
							onClick={() => { setActiveView('settings'); setActiveSubView(''); }}
							className="flex items-center gap-3"
							title={isSidebarCollapsed && !isSidebarHovered ? "Settings" : ""}
						>
							<Settings className="w-4 h-4 flex-shrink-0" />
							{(!isSidebarCollapsed || isSidebarHovered) && (
								<span className="ml-3 text-sm font-medium">Settings</span>
							)}
						</SidebarNavItem>
						<SidebarNavItem 
							href="#" 
							active={activeView === 'testing'}
							onClick={() => { setActiveView('testing'); setActiveSubView(''); }}
							className="flex items-center gap-3"
							title={isSidebarCollapsed && !isSidebarHovered ? "Testing" : ""}
						>
							<Wrench className="w-4 h-4 flex-shrink-0" />
							{(!isSidebarCollapsed || isSidebarHovered) && (
								<span className="ml-3 text-sm font-medium">Testing</span>
							)}
						</SidebarNavItem>
					</SidebarNav>
					<div className="mt-auto p-4 border-t">
						<Button 
							variant="ghost" 
							className={`w-full ${isSidebarCollapsed && !isSidebarHovered ? 'justify-center px-2' : 'justify-start'}`}
							onClick={onLogout}
							title={isSidebarCollapsed && !isSidebarHovered ? "Logout" : ""}
						>
							<LogOut className="w-4 h-4 flex-shrink-0" />
							{(!isSidebarCollapsed || isSidebarHovered) && (
								<span className="ml-3 text-sm font-medium">Logout</span>
							)}
						</Button>
					</div>
				</div>
			</div>

			{/* Main Content */}
			<div className="flex-1 flex flex-col overflow-hidden">
				{/* Header with Navigation */}
				<header className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-6 py-4">
					<div className="flex items-center justify-between">
						<div className="flex items-center gap-6">
							{/* Section Navigation Dropdown */}
							{activeView !== 'dashboard' && (
								<DropdownMenu>
									<DropdownMenuTrigger asChild>
										<Button variant="outline" className="gap-2">
											{activeView === 'operations' && <><Activity className="w-4 h-4" /> Operations</>}
											{activeView === 'financials' && <><DollarSign className="w-4 h-4" /> Financials</>}
											{activeView === 'data' && <><Database className="w-4 h-4" /> Data Management</>}
											{activeView === 'reporting' && <><FileText className="w-4 h-4" /> Reporting</>}
											{activeView === 'utilities' && <><Wrench className="w-4 h-4" /> Utilities</>}
											{activeView === 'settings' && <><Settings className="w-4 h-4" /> Settings</>}
											<ChevronLeft className="w-4 h-4 rotate-270" />
										</Button>
									</DropdownMenuTrigger>
									<DropdownMenuContent align="start" className="w-56">
										{activeView === 'operations' && (
											<>
												<DropdownMenuLabel>Operations Menu</DropdownMenuLabel>
												<DropdownMenuSeparator />
												<DropdownMenuItem onClick={() => setActiveSubView('wells')}>
													<Activity className="mr-2 h-4 w-4" />
													<span>Wells Management</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('production')}>
													<TrendingUp className="mr-2 h-4 w-4" />
													<span>Production Tracking</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('field-ops')}>
													<Wrench className="mr-2 h-4 w-4" />
													<span>Field Operations</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('maintenance')}>
													<Settings className="mr-2 h-4 w-4" />
													<span>Maintenance</span>
												</DropdownMenuItem>
											</>
										)}
										{activeView === 'financials' && (
											<>
												<DropdownMenuLabel>Financial Menu</DropdownMenuLabel>
												<DropdownMenuSeparator />
												<DropdownMenuItem onClick={() => setActiveSubView('transactions')}>
													<DollarSign className="mr-2 h-4 w-4" />
													<span>Transactions</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('revenue')}>
													<TrendingUp className="mr-2 h-4 w-4" />
													<span>Revenue Analysis</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('expenses')}>
													<Calculator className="mr-2 h-4 w-4" />
													<span>Expense Management</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('analytics')}>
													<BarChart3 className="mr-2 h-4 w-4" />
													<span>Financial Analytics</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('banking')}>
													<Home className="mr-2 h-4 w-4" />
													<span>Banking</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('accounting')}>
													<FileText className="mr-2 h-4 w-4" />
													<span>Accounting Tools</span>
												</DropdownMenuItem>
												<DropdownMenuSeparator />
												<DropdownMenuItem onClick={() => setActiveSubView('financial-settings')}>
													<Settings className="mr-2 h-4 w-4" />
													<span>Settings</span>
												</DropdownMenuItem>
											</>
										)}
										{activeView === 'data' && (
											<>
												<DropdownMenuLabel>Data Menu</DropdownMenuLabel>
												<DropdownMenuSeparator />
												<DropdownMenuItem onClick={() => setActiveSubView('dbf-explorer')}>
													<Database className="mr-2 h-4 w-4" />
													<span>DBF Explorer</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('import')}>
													<Upload className="mr-2 h-4 w-4" />
													<span>Import Data</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('export')}>
													<Download className="mr-2 h-4 w-4" />
													<span>Export Data</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('backup')}>
													<Archive className="mr-2 h-4 w-4" />
													<span>Backup & Restore</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('maintenance')}>
													<Wrench className="mr-2 h-4 w-4" />
													<span>Database Maintenance</span>
												</DropdownMenuItem>
											</>
										)}
										{activeView === 'reporting' && (
											<>
												<DropdownMenuLabel>Reports Menu</DropdownMenuLabel>
												<DropdownMenuSeparator />
												<DropdownMenuItem onClick={() => setActiveSubView('state')}>
													<FileText className="mr-2 h-4 w-4" />
													<span>State Reports</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('financial')}>
													<DollarSign className="mr-2 h-4 w-4" />
													<span>Financial Reports</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('production')}>
													<TrendingUp className="mr-2 h-4 w-4" />
													<span>Production Reports</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('custom')}>
													<FileSearch className="mr-2 h-4 w-4" />
													<span>Custom Reports</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('audit')}>
													<Copy className="mr-2 h-4 w-4" />
													<span>Audit Trail</span>
												</DropdownMenuItem>
											</>
										)}
										{activeView === 'utilities' && (
											<>
												<DropdownMenuLabel>Utilities Menu</DropdownMenuLabel>
												<DropdownMenuSeparator />
												<DropdownMenuItem onClick={() => setActiveSubView('calculator')}>
													<Calculator className="mr-2 h-4 w-4" />
													<span>Calculators</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('converter')}>
													<Activity className="mr-2 h-4 w-4" />
													<span>Unit Converter</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('scheduler')}>
													<Calendar className="mr-2 h-4 w-4" />
													<span>Task Scheduler</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('tools')}>
													<Wrench className="mr-2 h-4 w-4" />
													<span>Data Tools</span>
												</DropdownMenuItem>
											</>
										)}
										{activeView === 'settings' && (
											<>
												<DropdownMenuLabel>Settings Menu</DropdownMenuLabel>
												<DropdownMenuSeparator />
												<DropdownMenuItem onClick={() => setActiveSubView('users')}>
													<Users className="mr-2 h-4 w-4" />
													<span>User Management</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('appearance')}>
													<Settings className="mr-2 h-4 w-4" />
													<span>Appearance</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('system')}>
													<Database className="mr-2 h-4 w-4" />
													<span>System Configuration</span>
												</DropdownMenuItem>
												<DropdownMenuItem onClick={() => setActiveSubView('security')}>
													<FileText className="mr-2 h-4 w-4" />
													<span>Security Settings</span>
												</DropdownMenuItem>
											</>
										)}
									</DropdownMenuContent>
								</DropdownMenu>
							)}
              
							{/* Page Title */}
							<div>
								<h2 className="text-2xl font-semibold tracking-tight">
									{activeView === 'dashboard' && 'Dashboard'}
									{activeView === 'operations' && (activeSubView || 'Operations Dashboard')}
									{activeView === 'financials' && (activeSubView || 'Financial Dashboard')}
									{activeView === 'data' && (activeSubView || 'Data Management Dashboard')}
									{activeView === 'reporting' && (activeSubView || 'Reports Dashboard')}
									{activeView === 'utilities' && (activeSubView || 'Utilities Dashboard')}
									{activeView === 'settings' && (activeSubView || 'Settings Dashboard')}
									{activeView === 'testing' && 'Testing Tools'}
								</h2>
								<p className="text-sm text-muted-foreground">
									{activeView === 'dashboard' && 'Overview of your financial data'}
									{activeView === 'operations' && 'Manage wells, production, and field operations'}
									{activeView === 'financials' && 'Financial transactions, analytics, and accounting'}
									{activeView === 'data' && 'Database maintenance and data management'}
									{activeView === 'reporting' && 'Reports, compliance, and documentation'}
									{activeView === 'utilities' && 'Tools, calculators, and system utilities'}
									{activeView === 'settings' && 'System configuration and user management'}
									{activeView === 'testing' && 'Developer tools for testing and debugging'}
								</p>
							</div>
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
							{/* Operations Dashboard */}
							{!activeSubView && (
								<div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4 mb-8">
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('wells')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Wells Management</p>
													<h3 className="text-2xl font-bold">Active Wells</h3>
													<p className="text-sm text-muted-foreground mt-2">Manage well information</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Activity className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('production')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Production</p>
													<h3 className="text-2xl font-bold">Tracking</h3>
													<p className="text-sm text-muted-foreground mt-2">Monitor oil & gas production</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<TrendingUp className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('field-ops')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Field Ops</p>
													<h3 className="text-2xl font-bold">Operations</h3>
													<p className="text-sm text-muted-foreground mt-2">Field activities & schedules</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Wrench className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('maintenance')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Maintenance</p>
													<h3 className="text-2xl font-bold">Equipment</h3>
													<p className="text-sm text-muted-foreground mt-2">Track maintenance schedules</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Settings className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
								</div>
							)}
              
							{/* Sub-views */}
							{activeSubView && (
								<Tabs value={activeSubView} onValueChange={setActiveSubView} className="w-full">
									<TabsList className="hidden">
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
							)}
						</div>
					)}

					{activeView === 'financials' && (
						<div className="space-y-6">
							{/* Financial Dashboard */}
							{!activeSubView && (
								<div>
									<div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 mb-8">
										<Card 
											className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
											onClick={() => setActiveSubView('transactions')}
										>
											<CardContent className="p-8">
												<div className="flex items-start justify-between mb-4">
													<div className="space-y-1">
														<p className="text-sm font-medium text-muted-foreground">Transactions</p>
														<h3 className="text-2xl font-bold">Recent Activity</h3>
														<p className="text-sm text-muted-foreground mt-2">View and manage transactions</p>
													</div>
													<div className="p-3 bg-primary/10 rounded-lg">
														<DollarSign className="w-5 h-5 text-primary" />
													</div>
												</div>
											</CardContent>
										</Card>
										<Card 
											className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
											onClick={() => setActiveSubView('banking')}
										>
											<CardContent className="p-8">
												<div className="flex items-start justify-between mb-4">
													<div className="space-y-1">
														<p className="text-sm font-medium text-muted-foreground">Banking</p>
														<h3 className="text-2xl font-bold">Accounts</h3>
														<p className="text-sm text-muted-foreground mt-2">Bank accounts and reconciliation</p>
													</div>
													<div className="p-3 bg-primary/10 rounded-lg">
														<Home className="w-5 h-5 text-primary" />
													</div>
												</div>
											</CardContent>
										</Card>
										<Card 
											className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
											onClick={() => setActiveSubView('analytics')}
										>
											<CardContent className="p-8">
												<div className="flex items-start justify-between mb-4">
													<div className="space-y-1">
														<p className="text-sm font-medium text-muted-foreground">Analytics</p>
														<h3 className="text-2xl font-bold">Insights</h3>
														<p className="text-sm text-muted-foreground mt-2">Financial trends and analysis</p>
													</div>
													<div className="p-3 bg-primary/10 rounded-lg">
														<BarChart3 className="w-5 h-5 text-primary" />
													</div>
												</div>
											</CardContent>
										</Card>
									</div>
                  
									{/* Quick Stats */}
									<div className="grid gap-4 md:grid-cols-2">
										<Card>
											<CardHeader>
												<CardTitle>Revenue Overview</CardTitle>
											</CardHeader>
											<CardContent>
												<RevenueChart />
											</CardContent>
										</Card>
										<Card>
											<CardHeader>
												<CardTitle>Recent Transactions</CardTitle>
											</CardHeader>
											<CardContent>
												<div className="text-center py-8 text-muted-foreground">
													Click Transactions above to view details
												</div>
											</CardContent>
										</Card>
									</div>
								</div>
							)}
              
							{/* Sub-views */}
							{activeSubView && (
								<Tabs value={activeSubView} onValueChange={setActiveSubView} className="w-full">
									<TabsList className="hidden">
										<TabsTrigger value="transactions">Transactions</TabsTrigger>
										<TabsTrigger value="revenue">Revenue</TabsTrigger>
										<TabsTrigger value="expenses">Expenses</TabsTrigger>
										<TabsTrigger value="analytics">Analytics</TabsTrigger>
										<TabsTrigger value="banking">Banking</TabsTrigger>
										<TabsTrigger value="accounting">Accounting</TabsTrigger>
										<TabsTrigger value="financial-settings">Settings</TabsTrigger>
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
									<BankingSection companyName={currentUser?.company_path || currentUser?.company_name} currentUser={currentUser} />
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
								<TabsContent value="financial-settings" className="space-y-4">
									<div className="grid gap-6 md:grid-cols-2">
										<CompanyInformation currentUser={currentUser} />
										<Card>
											<CardHeader>
												<CardTitle>Database Testing</CardTitle>
												<CardDescription>Test OLE/COM Database Connection</CardDescription>
											</CardHeader>
											<CardContent>
												<Button 
													onClick={() => {
														setActiveView('data')
														setActiveSubView('database-test')
													}}
													className="w-full"
												>
													<Database className="mr-2 h-4 w-4" />
													Open Database Test Tool
												</Button>
												<p className="text-sm text-muted-foreground mt-2">
													Test your Visual FoxPro OLE automation and run database queries.
												</p>
											</CardContent>
										</Card>
									</div>
								</TabsContent>
							</Tabs>
							)}
						</div>
					)}

					{activeView === 'data' && (
						<div className="space-y-6">
							{/* Data Management Dashboard */}
							{!activeSubView && (
								<div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 mb-8">
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('dbf-explorer')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Database</p>
													<h3 className="text-2xl font-bold">DBF Explorer</h3>
													<p className="text-sm text-muted-foreground mt-2">Browse and edit DBF files</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Database className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('import')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Data</p>
													<h3 className="text-2xl font-bold">Import</h3>
													<p className="text-sm text-muted-foreground mt-2">Import external data</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Upload className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('export')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Data</p>
													<h3 className="text-2xl font-bold">Export</h3>
													<p className="text-sm text-muted-foreground mt-2">Export data to files</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Download className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('backup')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">System</p>
													<h3 className="text-2xl font-bold">Backup</h3>
													<p className="text-sm text-muted-foreground mt-2">Backup and restore data</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Archive className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('maintenance')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Database</p>
													<h3 className="text-2xl font-bold">Maintenance</h3>
													<p className="text-sm text-muted-foreground mt-2">Optimize database</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Wrench className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
								</div>
							)}
              
							{/* Sub-views */}
							{activeSubView === 'dbf-explorer' && (
								<div className="space-y-4">
									<DBFExplorer currentUser={currentUser} />
								</div>
							)}
							{activeSubView === 'database-test' && (
								<div className="space-y-4">
									<DatabaseTest currentUser={currentUser} />
								</div>
							)}
							{activeSubView === 'import' && (
								<div className="space-y-4">
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
								</div>
							)}
							{activeSubView === 'export' && (
								<div className="space-y-4">
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
								</div>
							)}
							{activeSubView === 'backup' && (
								<div className="space-y-4">
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
								</div>
							)}
							{activeSubView === 'maintenance' && (
								<div className="space-y-4">
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
								</div>
							)}
						</div>
					)}

					{activeView === 'reporting' && (
						<div className="space-y-6">
							{/* Reports Dashboard */}
							{!activeSubView && (
								<div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 mb-8">
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('state')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Compliance</p>
													<h3 className="text-2xl font-bold">State Reports</h3>
													<p className="text-sm text-muted-foreground mt-2">Generate state reports</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<FileText className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('financial')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Accounting</p>
													<h3 className="text-2xl font-bold">Financial Reports</h3>
													<p className="text-sm text-muted-foreground mt-2">Financial statements</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<DollarSign className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('production')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Operations</p>
													<h3 className="text-2xl font-bold">Production Reports</h3>
													<p className="text-sm text-muted-foreground mt-2">Well production data</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<TrendingUp className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('custom')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Analytics</p>
													<h3 className="text-2xl font-bold">Custom Reports</h3>
													<p className="text-sm text-muted-foreground mt-2">Build custom reports</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<BarChart3 className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('audit')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Security</p>
													<h3 className="text-2xl font-bold">Audit Trail</h3>
													<p className="text-sm text-muted-foreground mt-2">Activity and change logs</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Shield className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
								</div>
							)}
              
							{/* Sub-views */}
							{activeSubView === 'state' && (
								<div className="space-y-4">
									<StateReportsSection currentUser={currentUser} />
								</div>
							)}
							{activeSubView === 'financial' && (
								<div className="space-y-4">
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
								</div>
							)}
							{activeSubView === 'production' && (
								<div className="space-y-4">
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
								</div>
							)}
							{activeSubView === 'custom' && (
								<div className="space-y-4">
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
								</div>
							)}
							{activeSubView === 'audit' && (
								<div className="space-y-4">
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
								</div>
							)}
						</div>
					)}

					{activeView === 'settings' && (
						<div className="space-y-6">
							{/* Settings Dashboard */}
							{!activeSubView && (
								<div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4 mb-8">
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('users')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Users</p>
													<h3 className="text-2xl font-bold">Management</h3>
													<p className="text-sm text-muted-foreground mt-2">Manage user accounts</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Users className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('appearance')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Appearance</p>
													<h3 className="text-2xl font-bold">Theme</h3>
													<p className="text-sm text-muted-foreground mt-2">Customize look and feel</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Settings className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('system')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">System</p>
													<h3 className="text-2xl font-bold">Configuration</h3>
													<p className="text-sm text-muted-foreground mt-2">System-wide settings</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Database className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('security')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Security</p>
													<h3 className="text-2xl font-bold">Policies</h3>
													<p className="text-sm text-muted-foreground mt-2">Security settings</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<FileText className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
								</div>
							)}
              
							{/* Sub-views */}
							{activeSubView === 'users' && (
								<div className="space-y-4">
									<UserManagement currentUser={currentUser} />
								</div>
							)}
							{activeSubView === 'appearance' && (
								<div className="space-y-4">
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
								</div>
							)}
							{activeSubView === 'system' && (
								<div className="space-y-4">
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
								</div>
							)}
							{activeSubView === 'security' && (
								<div className="space-y-4">
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
								</div>
							)}
						</div>
					)}

					{activeView === 'testing' && (
						<div className="space-y-6">
							<LoggingTest />
						</div>
					)}

					{activeView === 'utilities' && (
						<div className="space-y-6">
							{/* Utilities Dashboard */}
							{!activeSubView && (
								<div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4 mb-8">
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('calculator')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Calculators</p>
													<h3 className="text-2xl font-bold">Financial</h3>
													<p className="text-sm text-muted-foreground mt-2">Various calculation tools</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Calculator className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('converter')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Converter</p>
													<h3 className="text-2xl font-bold">Units</h3>
													<p className="text-sm text-muted-foreground mt-2">Convert between units</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Activity className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('scheduler')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Scheduler</p>
													<h3 className="text-2xl font-bold">Tasks</h3>
													<p className="text-sm text-muted-foreground mt-2">Automated scheduling</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Calendar className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
									<Card 
										className="cursor-pointer hover:shadow-lg transition-all hover:scale-[1.02]"
										onClick={() => setActiveSubView('tools')}
									>
										<CardContent className="p-8">
											<div className="flex items-start justify-between mb-4">
												<div className="space-y-1">
													<p className="text-sm font-medium text-muted-foreground">Data Tools</p>
													<h3 className="text-2xl font-bold">Utilities</h3>
													<p className="text-sm text-muted-foreground mt-2">Data validation tools</p>
												</div>
												<div className="p-3 bg-primary/10 rounded-lg">
													<Wrench className="w-5 h-5 text-primary" />
												</div>
											</div>
										</CardContent>
									</Card>
								</div>
							)}
              
							{/* Sub-views */}
							{activeSubView === 'calculator' && (
								<div className="space-y-4">
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
								</div>
							)}
							{activeSubView === 'converter' && (
								<div className="space-y-4">
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
								</div>
							)}
							{activeSubView === 'scheduler' && (
								<div className="space-y-4">
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
								</div>
							)}
							{activeSubView === 'tools' && (
								<div className="space-y-4">
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
								</div>
							)}
						</div>
					)}
				</main>
			</div>
		</div>
	)
}

export default App
