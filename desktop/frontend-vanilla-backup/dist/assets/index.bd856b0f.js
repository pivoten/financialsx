(function(){const e=document.createElement("link").relList;if(e&&e.supports&&e.supports("modulepreload"))return;for(const t of document.querySelectorAll('link[rel="modulepreload"]'))a(t);new MutationObserver(t=>{for(const n of t)if(n.type==="childList")for(const i of n.addedNodes)i.tagName==="LINK"&&i.rel==="modulepreload"&&a(i)}).observe(document,{childList:!0,subtree:!0});function s(t){const n={};return t.integrity&&(n.integrity=t.integrity),t.referrerpolicy&&(n.referrerPolicy=t.referrerpolicy),t.crossorigin==="use-credentials"?n.credentials="include":t.crossorigin==="anonymous"?n.credentials="omit":n.credentials="same-origin",n}function a(t){if(t.ep)return;t.ep=!0;const n=s(t);fetch(t.href,n)}})();const m="/assets/pivoten.logo.be32852b.png";function v(){return window.go.main.App.GetCompanies()}function b(o,e,s){return window.go.main.App.Login(o,e,s)}function y(o){return window.go.main.App.Logout(o)}function f(o,e,s,a){return window.go.main.App.Register(o,e,s,a)}function w(o,e){return window.go.main.App.ValidateSession(o,e)}class L{constructor(){this.currentUser=null,this.session=null,this.companies=[]}async loadCompanies(){try{return console.log("Calling GetCompanies..."),this.companies=await v(),console.log("GetCompanies result:",this.companies),Array.isArray(this.companies)?this.companies:[]}catch(e){return console.error("Failed to load companies:",e),[]}}async login(e,s,a){try{const t=await b(e,s,a);return this.currentUser=t.user,this.session=t.session,localStorage.setItem("session_token",this.session.token),localStorage.setItem("company_name",a),t}catch(t){throw t}}async register(e,s,a,t){try{console.log("Registering:",{username:e,email:a,companyName:t});const n=await f(e,s,a,t);return console.log("Register result:",n),this.currentUser=n.user,this.session=n.session,localStorage.setItem("session_token",this.session.token),localStorage.setItem("company_name",t),n}catch(n){throw console.error("Register error:",n),n}}async checkSession(){const e=localStorage.getItem("session_token"),s=localStorage.getItem("company_name");if(console.log("Session token:",e?"present":"not found"),console.log("Company name:",s),!e||!s)return!1;try{return console.log("Validating session..."),this.currentUser=await w(e,s),console.log("Session valid, user:",this.currentUser),!0}catch(a){return console.log("Session invalid:",a.message),localStorage.removeItem("session_token"),localStorage.removeItem("company_name"),!1}}async logout(){const e=localStorage.getItem("session_token");if(e)try{await y(e)}catch(s){console.error("Logout error:",s)}this.currentUser=null,this.session=null,localStorage.removeItem("session_token"),localStorage.removeItem("company_name")}}const r=new L;function S(o,e=[],s){return`
        <div class="auth-container">
            <h2>Login to FinancialsX</h2>
            <form id="login-form" class="auth-form">
                <div class="form-group">
                    <label for="company">Company</label>
                    <select id="company" name="company" required>
                        <option value="">Select a company...</option>
                        ${(e||[]).map(a=>`<option value="${a.name}">${a.name}</option>`).join("")}
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
    `}function x(o,e=[],s){return`
        <div class="auth-container">
            <h2>Register for FinancialsX</h2>
            <form id="register-form" class="auth-form">
                <div class="form-group">
                    <label for="company">Company</label>
                    <select id="company" name="company" required>
                        <option value="">Select a company...</option>
                        ${(e||[]).map(a=>`<option value="${a.name}">${a.name}</option>`).join("")}
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
    `}function C(o,e="home"){return`
        <div class="dashboard">
            <!-- Top Header -->
            <header class="dashboard-header">
                <div class="header-left">
                    <h1>FinancialsX</h1>
                </div>
                <div class="header-right">
                    <div class="user-info">
                        <span class="username">Welcome, ${o.username}</span>
                        <span class="company-name">${o.company_name}</span>
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
                            <li class="nav-item ${e==="home"?"active":""}">
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
                            <li class="nav-item ${e==="db-maintenance"?"active":""}">
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
                            <li class="nav-item ${e==="state-reporting"?"active":""} has-submenu">
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
                                <ul class="submenu ${e.startsWith("state-reporting")?"expanded":""}">
                                    <li class="nav-item ${e==="state-reporting-wv"?"active":""}">
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
                        ${k(e,o)}
                    </div>
                </main>
            </div>
        </div>
    `}function k(o,e){switch(o){case"home":return`
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
                                <span>${e.company_name}</span>
                            </div>
                            <div class="info-row">
                                <label>User:</label>
                                <span>${e.username}</span>
                            </div>
                            <div class="info-row">
                                <label>Last Login:</label>
                                <span>${e.last_login?new Date(e.last_login).toLocaleDateString():"First login"}</span>
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
            `;case"db-maintenance":return`
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
            `;case"state-reporting":return`
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
            `;case"state-reporting-wv":return`
                <div class="page-header">
                    <div class="breadcrumb">
                        <a href="#" onclick="navigateToPage('state-reporting')">State Reporting</a>
                        <span class="separator">\u203A</span>
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
            `;default:return'<div class="page-content">Page not found</div>'}}class E{constructor(){this.appElement=document.querySelector("#app"),this.currentPage="home",this.init().catch(console.error)}async init(){try{console.log("Initializing app..."),this.appElement.innerHTML='<div style="padding: 20px; text-align: center;">Loading...</div>',console.log("Checking session...");const e=await r.checkSession();console.log("Has session:",e),e?(console.log("Showing dashboard"),this.showDashboard()):(console.log("Showing login"),this.showLogin())}catch(e){console.error("Error during app initialization:",e),this.appElement.innerHTML='<div style="padding: 20px; color: red;">Error loading app: '+e.message+"</div>"}}async showLogin(){try{console.log("Loading companies...");const e=await r.loadCompanies();console.log("Companies loaded:",e),this.appElement.innerHTML=S(null,e);const s=document.createElement("div");s.className="logo-container",s.innerHTML=`<img src="${m}" alt="Pivoten Logo" class="auth-logo">`,this.appElement.insertBefore(s,this.appElement.firstChild);const a=document.getElementById("login-form"),t=document.getElementById("login-error");a.addEventListener("submit",async n=>{n.preventDefault(),t.textContent="";const i=new FormData(a),d=i.get("username"),l=i.get("password"),c=i.get("company");if(!c){t.textContent="Please select a company";return}try{await r.login(d,l,c),this.showDashboard()}catch(p){t.textContent=p.message||"Login failed"}}),document.getElementById("switch-to-register").addEventListener("click",n=>{n.preventDefault(),this.showRegister()})}catch(e){console.error("Error in showLogin:",e),this.appElement.innerHTML='<div style="padding: 20px; color: red;">Error loading login page: '+e.message+"</div>"}}async showRegister(){const e=await r.loadCompanies();this.appElement.innerHTML=x(null,e);const s=document.createElement("div");s.className="logo-container",s.innerHTML=`<img src="${m}" alt="Pivoten Logo" class="auth-logo">`,this.appElement.insertBefore(s,this.appElement.firstChild);const a=document.getElementById("register-form"),t=document.getElementById("register-error");a.addEventListener("submit",async n=>{n.preventDefault(),t.textContent="";const i=new FormData(a),d=i.get("username"),l=i.get("password"),c=i.get("confirm-password"),p=i.get("email"),u=i.get("company");if(!u){t.textContent="Please select a company";return}if(l!==c){t.textContent="Passwords do not match";return}try{await r.register(d,l,p,u),this.showDashboard()}catch(h){t.textContent=h.message||"Registration failed"}}),document.getElementById("switch-to-login").addEventListener("click",n=>{n.preventDefault(),this.showLogin()})}showDashboard(){if(!r.currentUser){this.showLogin();return}this.appElement.innerHTML=C(r.currentUser,this.currentPage),document.getElementById("logout-btn").addEventListener("click",async()=>{await r.logout(),this.showLogin()}),this.setupNavigation()}setupNavigation(){document.querySelectorAll(".nav-link[data-page]").forEach(e=>{e.addEventListener("click",s=>{s.preventDefault();const a=e.getAttribute("data-page");this.navigateToPage(a)})}),document.querySelectorAll(".has-submenu > .nav-link").forEach(e=>{e.addEventListener("click",s=>{const a=e.parentElement.querySelector(".submenu");a&&a.classList.toggle("expanded")})})}navigateToPage(e){this.currentPage=e;const s=document.getElementById("page-content");s&&(s.innerHTML=this.getPageContent(e,r.currentUser)),document.querySelectorAll(".nav-item").forEach(t=>{t.classList.remove("active")});const a=document.querySelector(`[data-page="${e}"]`);if(a){a.closest(".nav-item").classList.add("active");const t=a.closest(".submenu");t&&t.classList.add("expanded")}}getPageContent(e,s){this.showDashboard()}}const g=document.createElement("style");g.textContent=`
    .logo-container {
        text-align: center;
        margin: 40px 0 30px 0;
        padding-top: 20px;
    }
    
    .auth-logo {
        max-width: 200px;
        height: auto;
    }
    
    #app {
        min-height: 100vh;
        background-color: #f5f5f5;
    }
    
    body {
        margin: 0;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
    }
`;document.head.appendChild(g);const D=new E;window.navigateToPage=function(o){D.navigateToPage(o)};
