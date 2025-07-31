export function LoginForm(onSubmit, companies = [], onSwitchToRegister) {
    return `
        <div class="auth-container">
            <h2>Login to FinancialsX</h2>
            <form id="login-form" class="auth-form">
                <div class="form-group">
                    <label for="company">Company</label>
                    <select id="company" name="company" required>
                        <option value="">Select a company...</option>
                        ${(companies || []).map(c => `<option value="${c.name}">${c.name}</option>`).join('')}
                    </select>
                </div>
                <div class="form-group">
                    <label for="username">Username</label>
                    <input type="text" id="username" name="username" required autocomplete="username">
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required autocomplete="current-password">
                </div>
                <div class="form-error" id="login-error"></div>
                <button type="submit" class="btn btn-primary">Login</button>
                <p class="auth-switch">
                    Don't have an account? 
                    <a href="#" id="switch-to-register">Register</a>
                </p>
            </form>
        </div>
    `;
}

export function RegisterForm(onSubmit, companies = [], onSwitchToLogin) {
    return `
        <div class="auth-container">
            <h2>Register for FinancialsX</h2>
            <form id="register-form" class="auth-form">
                <div class="form-group">
                    <label for="company">Company</label>
                    <select id="company" name="company" required>
                        <option value="">Select a company...</option>
                        ${(companies || []).map(c => `<option value="${c.name}">${c.name}</option>`).join('')}
                    </select>
                </div>
                <div class="form-group">
                    <label for="username">Username</label>
                    <input type="text" id="username" name="username" required autocomplete="username">
                </div>
                <div class="form-group">
                    <label for="email">Email (optional)</label>
                    <input type="email" id="email" name="email" autocomplete="email">
                </div>
                <div class="form-group">
                    <label for="password">Password</label>
                    <input type="password" id="password" name="password" required autocomplete="new-password">
                </div>
                <div class="form-group">
                    <label for="confirm-password">Confirm Password</label>
                    <input type="password" id="confirm-password" name="confirm-password" required>
                </div>
                <div class="form-error" id="register-error"></div>
                <button type="submit" class="btn btn-primary">Register</button>
                <p class="auth-switch">
                    Already have an account? 
                    <a href="#" id="switch-to-login">Login</a>
                </p>
            </form>
        </div>
    `;
}

export function Dashboard(user, currentPage = 'home') {
    return `
        <div class="dashboard">
            <!-- Top Header -->
            <header class="dashboard-header">
                <div class="header-left">
                    <h1>FinancialsX</h1>
                </div>
                <div class="header-right">
                    <div class="user-info">
                        <span class="username">Welcome, ${user.username}</span>
                        <span class="company-name">${user.company_name}</span>
                        <button id="logout-btn" class="btn btn-logout">
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
                                <polyline points="16,17 21,12 16,7"></polyline>
                                <line x1="21" y1="12" x2="9" y2="12"></line>
                            </svg>
                            Logout
                        </button>
                    </div>
                </div>
            </header>

            <!-- Main Layout -->
            <div class="dashboard-main">
                <!-- Sidebar Navigation -->
                <nav class="sidebar">
                    <div class="nav-section">
                        <h3>Main</h3>
                        <ul class="nav-menu">
                            <li class="nav-item ${currentPage === 'home' ? 'active' : ''}">
                                <a href="#" data-page="home" class="nav-link">
                                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
                                        <polyline points="9,22 9,12 15,12 15,22"></polyline>
                                    </svg>
                                    Dashboard
                                </a>
                            </li>
                        </ul>
                    </div>
                    
                    <div class="nav-section">
                        <h3>Tools</h3>
                        <ul class="nav-menu">
                            <li class="nav-item ${currentPage === 'db-maintenance' ? 'active' : ''}">
                                <a href="#" data-page="db-maintenance" class="nav-link">
                                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                                        <circle cx="8.5" cy="8.5" r="1.5"></circle>
                                        <path d="M21 15l-3.086-3.086a2 2 0 0 0-2.828 0L6 21"></path>
                                    </svg>
                                    DB Maintenance
                                </a>
                            </li>
                        </ul>
                    </div>
                    
                    <div class="nav-section">
                        <h3>Reporting</h3>
                        <ul class="nav-menu">
                            <li class="nav-item ${currentPage === 'state-reporting' ? 'active' : ''} has-submenu">
                                <a href="#" data-page="state-reporting" class="nav-link">
                                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                                        <polyline points="14,2 14,8 20,8"></polyline>
                                        <line x1="16" y1="13" x2="8" y2="13"></line>
                                        <line x1="16" y1="17" x2="8" y2="17"></line>
                                        <polyline points="10,9 9,9 8,9"></polyline>
                                    </svg>
                                    State Reporting
                                    <svg class="submenu-arrow" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <polyline points="6,9 12,15 18,9"></polyline>
                                    </svg>
                                </a>
                                <ul class="submenu ${currentPage.startsWith('state-reporting') ? 'expanded' : ''}">
                                    <li class="nav-item ${currentPage === 'state-reporting-wv' ? 'active' : ''}">
                                        <a href="#" data-page="state-reporting-wv" class="nav-link">
                                            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                <circle cx="12" cy="12" r="3"></circle>
                                                <path d="M12 1v6M12 17v6M4.22 4.22l4.24 4.24M15.54 15.54l4.24 4.24M1 12h6M17 12h6M4.22 19.78l4.24-4.24M15.54 8.46l4.24-4.24"></path>
                                            </svg>
                                            West Virginia
                                        </a>
                                    </li>
                                </ul>
                            </li>
                        </ul>
                    </div>
                </nav>

                <!-- Content Area -->
                <main class="content-area">
                    <div id="page-content">
                        ${getPageContent(currentPage, user)}
                    </div>
                </main>
            </div>
        </div>
    `;
}

function getPageContent(page, user) {
    switch(page) {
        case 'home':
            return `
                <div class="page-header">
                    <h2>Dashboard Overview</h2>
                    <p>Welcome to your FinancialsX dashboard</p>
                </div>
                <div class="dashboard-cards">
                    <div class="card">
                        <div class="card-header">
                            <h3>Company Information</h3>
                        </div>
                        <div class="card-content">
                            <div class="info-row">
                                <label>Company:</label>
                                <span>${user.company_name}</span>
                            </div>
                            <div class="info-row">
                                <label>User:</label>
                                <span>${user.username}</span>
                            </div>
                            <div class="info-row">
                                <label>Last Login:</label>
                                <span>${user.last_login ? new Date(user.last_login).toLocaleDateString() : 'First login'}</span>
                            </div>
                        </div>
                    </div>
                    
                    <div class="card">
                        <div class="card-header">
                            <h3>System Status</h3>
                        </div>
                        <div class="card-content">
                            <div class="status-item">
                                <span class="status-indicator success"></span>
                                <span>Database Connected</span>
                            </div>
                            <div class="status-item">
                                <span class="status-indicator success"></span>
                                <span>System Online</span>
                            </div>
                        </div>
                    </div>
                </div>
            `;
        case 'db-maintenance':
            return `
                <div class="page-header">
                    <h2>Database Maintenance</h2>
                    <p>Manage and maintain your database connections and data</p>
                </div>
                <div class="maintenance-section">
                    <div class="card">
                        <div class="card-header">
                            <h3>Database Tools</h3>
                        </div>
                        <div class="card-content">
                            <p>Database maintenance tools will be available here.</p>
                            <div class="button-group">
                                <button class="btn btn-primary" disabled>Check Connection</button>
                                <button class="btn btn-secondary" disabled>Optimize Database</button>
                                <button class="btn btn-secondary" disabled>Backup Data</button>
                            </div>
                        </div>
                    </div>
                    
                    <div class="card">
                        <div class="card-header">
                            <h3>DBF File Management</h3>
                        </div>
                        <div class="card-content">
                            <p>Tools for managing Visual FoxPro DBF files.</p>
                            <div class="button-group">
                                <button class="btn btn-primary" disabled>Import DBF Files</button>
                                <button class="btn btn-secondary" disabled>Validate Data</button>
                            </div>
                        </div>
                    </div>
                </div>
            `;
        case 'state-reporting':
            return `
                <div class="page-header">
                    <h2>State Reporting</h2>
                    <p>Generate reports for state regulatory compliance</p>
                </div>
                <div class="reporting-section">
                    <div class="card">
                        <div class="card-header">
                            <h3>Available States</h3>
                        </div>
                        <div class="card-content">
                            <p>Select a state to access specific reporting tools and requirements.</p>
                            <div class="state-grid">
                                <div class="state-card">
                                    <h4>West Virginia</h4>
                                    <p>Generate WV state compliance reports</p>
                                    <button class="btn btn-primary" onclick="navigateToPage('state-reporting-wv')">
                                        Access WV Reports
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            `;
        case 'state-reporting-wv':
            return `
                <div class="page-header">
                    <div class="breadcrumb">
                        <a href="#" onclick="navigateToPage('state-reporting')">State Reporting</a>
                        <span class="separator">›</span>
                        <span>West Virginia</span>
                    </div>
                    <h2>West Virginia State Reporting</h2>
                    <p>Generate compliance reports for West Virginia regulatory requirements</p>
                </div>
                <div class="wv-reporting-section">
                    <div class="card">
                        <div class="card-header">
                            <h3>WV Tax Reports</h3>
                        </div>
                        <div class="card-content">
                            <p>Generate West Virginia tax compliance reports from your accounting data.</p>
                            <div class="button-group">
                                <button class="btn btn-primary" disabled>WV Sales Tax Report</button>
                                <button class="btn btn-primary" disabled>WV Use Tax Report</button>
                                <button class="btn btn-secondary" disabled>WV Quarterly Filing</button>
                            </div>
                        </div>
                    </div>
                    
                    <div class="card">
                        <div class="card-header">
                            <h3>WV Business Reports</h3>
                        </div>
                        <div class="card-content">
                            <p>Business registration and compliance reports for West Virginia.</p>
                            <div class="button-group">
                                <button class="btn btn-primary" disabled>Annual Report</button>
                                <button class="btn btn-secondary" disabled>Business License Status</button>
                            </div>
                        </div>
                    </div>
                    
                    <div class="card">
                        <div class="card-header">
                            <h3>Report History</h3>
                        </div>
                        <div class="card-content">
                            <p>View and download previously generated WV reports.</p>
                            <div class="report-history">
                                <div class="report-item">
                                    <span class="report-name">No reports generated yet</span>
                                    <span class="report-date">-</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            `;
        default:
            return '<div class="page-content">Page not found</div>';
    }
}